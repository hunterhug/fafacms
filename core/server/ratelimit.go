package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/util/rds"
	log "github.com/hunterhug/golog"
)

// 反爬限流：Redis 固定窗口计数 + 连续超限递增拉黑（多副本共享）。

const (
	rateWindowSec = 60   // 统计窗口（秒）
	rateLimit     = 240  // 每 IP 每分钟请求上限
	rateBlockBase = 300  // 首次拉黑 5 分钟，之后递增
	rateBlockMax  = 3600 // 最长拉黑 1 小时
)

// rateSkipPaths 跳过限流的路径（健康检查、验证码、解封等，保证拉黑时也能取验证码并解封）
var rateSkipPaths = map[string]struct{}{
	"/ping":            {},
	"/captcha":         {},
	"/captcha/unblock": {},
}

func rateBlockKey(ip string) string     { return "ff_rl_blk:" + ip }
func rateWindowKey(ip string) string    { return "ff_rl:" + ip }
func rateViolationKey(ip string) string { return "ff_rl_v:" + ip }

// ClearBlock 解封某 IP（验证码校验通过后调用）：清除拉黑、计数窗口与违规计数。
func ClearBlock(ip string) {
	_ = rds.Del(rateBlockKey(ip), rateWindowKey(ip), rateViolationKey(ip))
}

// RateLimit 反爬限流中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, skip := rateSkipPaths[c.Request.URL.Path]; skip {
			c.Next()
			return
		}

		ip := c.ClientIP()
		now := time.Now().Unix()

		var blocked bool
		var retryAfter int64

		// 1. 检查拉黑状态
		if v, err := rds.Get(rateBlockKey(ip)); err == nil && v != "" {
			if until, e := strconv.ParseInt(v, 10, 64); e == nil && until > now {
				blocked = true
				retryAfter = until - now
			}
		}

		// 2. 未拉黑：固定窗口计数，超限则拉黑
		if !blocked {
			n, err := rds.IncrWindow(rateWindowKey(ip), rateWindowSec)
			if err != nil {
				// Redis 异常：失败放行，避免 Redis 抖动导致全站不可用
				log.Errorf("ratelimit incr err: %s", err.Error())
				c.Next()
				return
			}
			if n > rateLimit {
				v, _ := rds.IncrWindow(rateViolationKey(ip), 3600)
				retryAfter = int64(rateBlockBase) * v
				if retryAfter > rateBlockMax {
					retryAfter = rateBlockMax
				}
				_ = rds.SetEx(rateBlockKey(ip), strconv.FormatInt(now+retryAfter, 10), int(retryAfter))
				_ = rds.Del(rateWindowKey(ip))
				blocked = true
			}
		}

		if blocked {
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"flag":  false,
				"error": gin.H{"id": 100034, "msg": "request too frequent, please retry later"},
			})
			return
		}
		c.Next()
	}
}
