package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
)

// 2FA（TOTP）账号安全接口：生成秘钥/绑定/关闭/管理员重置。

const twoFaIssuer = "FaFaCMS"

// TwoFaSecret 生成并暂存 2FA 秘钥，返回 secret + otpauth URI（仅未开启时）。
func TwoFaSecret(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSONL(c, 200, nil, resp)
	}()

	uu, err := GetUserSession(c)
	if err != nil {
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	if uu.TwoFaSecret != "" {
		resp.Error = Error(TwoFaNeed, "already enabled")
		return
	}

	secret, err := util.GenerateTOTPSecret()
	if err != nil {
		resp.Error = Error(SystemProblem, err.Error())
		return
	}
	account := uu.Email
	if account == "" {
		account = uu.Name
	}
	uri := util.TOTPURI(twoFaIssuer, account, secret)
	setTwoFaSecretPending(uu.Id, secret)

	resp.Data = map[string]string{"secret": secret, "uri": uri}
	resp.Flag = true
}

// EnableTwoFa 校验动态码后绑定 2FA。
type EnableTwoFaRequest struct {
	Code string `json:"code"`
}

func EnableTwoFa(c *gin.Context) {
	resp := new(Resp)
	req := new(EnableTwoFaRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	secret, ok := getTwoFaSecretPending(uu.Id)
	if !ok {
		resp.Error = Error(TwoFaPendingExpired, "")
		return
	}

	if !util.ValidateTOTP(secret, req.Code) {
		resp.Error = Error(TwoFaWrong, "")
		return
	}

	u := new(model.User)
	u.Id = uu.Id
	u.TwoFaSecret = secret
	if err := u.UpdateTwoFa(); err != nil {
		resp.Error = Error(DBError, err.Error())
		return
	}
	clearTwoFaSecretPending(uu.Id)
	resp.Flag = true
}

// DisableTwoFa 关闭 2FA（需输入密码确认；丢失 2FA 码时请联系管理员）。
type DisableTwoFaRequest struct {
	Password string `json:"password"`
}

func DisableTwoFa(c *gin.Context) {
	resp := new(Resp)
	req := new(DisableTwoFaRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	pwOk, _ := model.CheckPassword(uu.Password, req.Password)
	if !pwOk {
		resp.Error = Error(LoginWrong, "password wrong")
		return
	}

	u := new(model.User)
	u.Id = uu.Id
	u.TwoFaSecret = ""
	if err := u.UpdateTwoFa(); err != nil {
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}

// ResetTwoFaAdmin 管理员重置用户 2FA（用户丢失 2FA 码时联系管理员处理）。
type ResetTwoFaRequest struct {
	Id int64 `json:"id"`
}

func ResetTwoFaAdmin(c *gin.Context) {
	resp := new(Resp)
	req := new(ResetTwoFaRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Id == 0 {
		resp.Error = Error(ParasError, "id empty")
		return
	}

	u := new(model.User)
	u.Id = req.Id
	u.TwoFaSecret = ""
	if err := u.UpdateTwoFa(); err != nil {
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}
