package controllers

import (
	"bytes"
	"encoding/base64"
	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/server"
	"github.com/hunterhug/fafacms/core/util/rds"
)

// redisCaptchaStore 将验证码存到 Redis（多副本共享）。
type redisCaptchaStore struct{}

func (redisCaptchaStore) Set(id string, digits []byte) {
	_ = rds.SetEx("ff_captcha:"+id, string(digits), 600)
}

func (redisCaptchaStore) Get(id string, clear bool) []byte {
	s, err := rds.Get("ff_captcha:" + id)
	if clear {
		_ = rds.Del("ff_captcha:" + id)
	}
	if err != nil || s == "" {
		return nil
	}
	return []byte(s)
}

func init() {
	captcha.SetCustomStore(redisCaptchaStore{})
}

// Captcha 生成图形验证码，返回 captcha_id + base64 png 图片。
// 验证码存储于 Redis（10 分钟过期，多副本共享）。
func Captcha(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSONL(c, 200, nil, resp)
	}()

	id := captcha.New()
	var buf bytes.Buffer
	if err := captcha.WriteImage(&buf, id, 160, 60); err != nil {
		resp.Error = Error(SystemProblem, err.Error())
		return
	}

	resp.Data = map[string]interface{}{
		"captcha_id": id,
		"image":      "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}
	resp.Flag = true
}

// verifyCaptcha 校验验证码（数字验证码，大小写无关）。
func verifyCaptcha(id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	return captcha.VerifyString(id, code)
}

// Unblock 限流解封：验证码通过后清除该 IP 的拉黑状态。
type UnblockRequest struct {
	CaptchaId   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

func Unblock(c *gin.Context) {
	resp := new(Resp)
	req := new(UnblockRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CaptchaId == "" || req.CaptchaCode == "" {
		resp.Error = Error(CaptchaNeed, "")
		return
	}

	if !verifyCaptcha(req.CaptchaId, req.CaptchaCode) {
		resp.Error = Error(CaptchaWrong, "")
		return
	}

	server.ClearBlock(c.ClientIP())
	resp.Flag = true
}
