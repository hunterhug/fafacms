package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	log "github.com/hunterhug/golog"
	"strings"
	"time"
)

type LoginRequest struct {
	UserName    string `json:"user_name"`
	PassWd      string `json:"pass_wd"`
	CaptchaId   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

func Login(c *gin.Context) {
	resp := new(Resp)
	req := new(LoginRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	// check session
	//userInfo, _ := GetUserSession(c)
	//if userInfo != nil {
	//	//c.Set("skipLog", true)
	//	c.Set("uid", userInfo.Id)
	//	resp.Flag = true
	//	return
	//}

	// paras not empty
	if req.UserName == "" || req.PassWd == "" {
		log.Errorf("login err:%s", "paras wrong")
		resp.Error = Error(ParasError, "field username or pass_wd")
		return
	}

	ip := c.ClientIP()

	// 防爆破：连续失败要求验证码 / 临时锁定
	needCaptcha, blocked, remain := checkLoginBlock(ip, req.UserName)
	if blocked {
		log.Errorf("login err:%s", "try too many")
		resp.Error = Error(LoginTryTooMany, fmt.Sprintf("retry after %d seconds", remain))
		return
	}
	if needCaptcha {
		if req.CaptchaId == "" || req.CaptchaCode == "" {
			resp.Error = Error(CaptchaNeed, "")
			return
		}
		if !verifyCaptcha(req.CaptchaId, req.CaptchaCode) {
			resp.Error = Error(CaptchaWrong, "")
			return
		}
	}

	// common people login (只按用户名/邮箱查询，密码单独校验)
	uu := new(model.User)
	if strings.Contains(req.UserName, "@") {
		uu.Email = req.UserName
	} else {
		uu.Name = req.UserName
	}
	ok, err := uu.GetRaw()
	if err != nil {
		log.Errorf("login err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		recordLoginFail(ip, req.UserName)
		log.Errorf("login err:%s", "user or password wrong")
		resp.Error = Error(LoginWrong, "user or password wrong")
		return
	}

	pwOk, needUpgrade := model.CheckPassword(uu.Password, req.PassWd)
	if !pwOk {
		recordLoginFail(ip, req.UserName)
		log.Errorf("login err:%s", "user or password wrong")
		resp.Error = Error(LoginWrong, "user or password wrong")
		return
	}

	// 旧明文密码平滑升级为 bcrypt
	if needUpgrade {
		if hash, e := model.HashPassword(req.PassWd); e == nil {
			uu.Password = hash
			_ = uu.UpdatePassword()
		}
	}
	clearLoginFail(ip, req.UserName)

	// 已开启 2FA：先返回待确认令牌，需二次校验动态码
	if uu.TwoFaSecret != "" {
		pending := newTwoFaPending(uu.Id)
		resp.Error = Error(TwoFaNeed, "")
		resp.Data = map[string]string{"pending": pending}
		return
	}

	c.Set("uid", uu.Id)

	u := new(model.User)
	u.Id = uu.Id
	u.LoginIp = c.ClientIP()
	u.LoginTime = time.Now().Unix()

	// Update the login ip into db
	err = u.UpdateLoginInfo()
	if err != nil {
		log.Errorf("login err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// Activate or black user can login, but those auth api can not use
	token, err := SetUserSession(uu.Id)
	if err != nil {
		log.Errorf("login err:%s", err.Error())
		resp.Error = Error(SetUserSessionError, err.Error())
		return
	}

	resp.Data = token
	resp.Flag = true
}

// LoginTwoFa 两步验证：校验 TOTP 动态码后发放真实 token。
type TwoFaLoginRequest struct {
	Pending string `json:"pending"`
	Code    string `json:"code"`
}

func LoginTwoFa(c *gin.Context) {
	resp := new(Resp)
	req := new(TwoFaLoginRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Pending == "" || req.Code == "" {
		resp.Error = Error(ParasError, "field pending or code")
		return
	}

	userId, ok := consumeTwoFaPending(req.Pending)
	if !ok {
		resp.Error = Error(TwoFaPendingExpired, "")
		return
	}

	uu := new(model.User)
	uu.Id = userId
	exist, err := uu.GetRaw()
	if err != nil {
		log.Errorf("LoginTwoFa err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		resp.Error = Error(UserNotFound, "")
		return
	}

	// 2FA 秘钥在库中为密文，校验前先解密
	twoFaSecret, tfErr := uu.TwoFaSecretPlain()
	if tfErr != nil {
		log.Errorf("TwoFaCheck err:%s", tfErr.Error())
		resp.Error = Error(SystemProblem, tfErr.Error())
		return
	}
	if !util.ValidateTOTP(twoFaSecret, req.Code) {
		resp.Error = Error(TwoFaWrong, "")
		return
	}

	u := new(model.User)
	u.Id = uu.Id
	u.LoginIp = c.ClientIP()
	u.LoginTime = time.Now().Unix()
	_ = u.UpdateLoginInfo()

	token, err := SetUserSession(uu.Id)
	if err != nil {
		log.Errorf("LoginTwoFa err:%s", err.Error())
		resp.Error = Error(SetUserSessionError, err.Error())
		return
	}

	resp.Data = token
	resp.Flag = true
}

func Logout(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSON(c, 200, resp)
	}()
	user, err := GetUserSession(c)

	if err != nil {
		log.Errorf("logout err:%s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	if user != nil {
		err = DeleteUserSession(c)
		if err != nil {
			log.Errorf("logout err:%s", err.Error())
			resp.Error = Error(DeleteUserSessionError, err.Error())
			return
		}
	}
	resp.Flag = true
}

func Refresh(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSON(c, 200, resp)
	}()

	user, err := GetUserSession(c)
	if err != nil {
		log.Errorf("refresh err:%s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	if user != nil {
		err = RefreshUserSession(c)
		if err != nil {
			log.Errorf("refresh err:%s", err.Error())
			resp.Error = Error(RefreshUserCacheError, err.Error())
			return
		}
	}
	resp.Flag = true
}
