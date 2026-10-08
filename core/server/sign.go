package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/util/rds"
)

// 轻量请求签名：防重放。
// 客户端对每个请求带 X-Ts(秒) / X-Nonce(随机) / X-Sign(HMAC(secret, ts:nonce))。
// 服务端校验时间窗 + nonce 一次性 + 签名。
// 说明：纯 SPA 中 secret 存在于前端 JS，只能抬高逆向门槛，配合限流/验证码使用。

var (
	signSecret = getSignSecret()
	// 默认强制校验签名；设置 FAFA_SIGN_STRICT=0 关闭强制（兼容脚本/调试）
	signStrict = os.Getenv("FAFA_SIGN_STRICT") != "0"
	signWindow = int64(300) // 5 分钟时间窗
)

func signNonceKey(nonce string) string { return "ff_nonce:" + nonce }

func getSignSecret() string {
	if s := os.Getenv("FAFA_SIGN_SECRET"); s != "" {
		return s
	}
	return "fafacms-sign-v1-2026"
}

func signToken(ts, nonce string) string {
	mac := hmac.New(sha256.New, []byte(signSecret))
	mac.Write([]byte(ts + ":" + nonce))
	return hex.EncodeToString(mac.Sum(nil))
}

// RequestSign 请求签名校验中间件（默认强制）。
func RequestSign() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// 健康检查与静态资源（图片等 <img> 请求不带签名头）豁免
		if path == "/ping" || strings.HasPrefix(path, "/storage") {
			c.Next()
			return
		}

		ts := c.GetHeader("X-Ts")
		nonce := c.GetHeader("X-Nonce")
		sign := c.GetHeader("X-Sign")

		if ts == "" && nonce == "" && sign == "" {
			if signStrict {
				abortSign(c)
				return
			}
			c.Next()
			return
		}

		tsInt, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			abortSign(c)
			return
		}
		now := time.Now().Unix()
		if now-tsInt > signWindow || tsInt-now > signWindow {
			abortSign(c)
			return
		}

		// nonce 一次性（Redis SET NX EX，多副本共享）；Redis 异常时失败放行
		if set, err := rds.SetNxEx(signNonceKey(nonce), "1", int(signWindow)); err != nil {
			// ignore: fail-open
		} else if !set {
			abortSign(c)
			return
		}

		expected := signToken(ts, nonce)
		if !hmac.Equal([]byte(expected), []byte(sign)) {
			abortSign(c)
			return
		}
		c.Next()
	}
}

func abortSign(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"flag":  false,
		"error": gin.H{"id": 100035, "msg": "request sign wrong"},
	})
}
