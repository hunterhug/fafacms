package controllers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/hunterhug/fafacms/core/util"
	"github.com/hunterhug/fafacms/core/util/rds"
)

// 登录/注册/评论/忘记/改密防爆破 + 2FA 待确认：Redis 存储（多副本共享）。

type failRecord struct {
	Count      int   `json:"count"`
	FirstTime  int64 `json:"first_time"`
	BlockUntil int64 `json:"block_until"`
}

const (
	loginWindowSec        = 600 // 10 分钟统计窗口
	loginCaptchaThreshold = 5   // 失败达到该次数要求验证码
	loginBlockThreshold   = 10  // 失败达到该次数临时锁定
	loginBlockSec         = 900 // 锁定 15 分钟

	registerWindowSec        = 3600 // 1 小时统计窗口
	registerCaptchaThreshold = 3    // 每 IP 每小时注册超过该次数要求验证码

	commentWindowSec        = 60 // 60 秒统计窗口
	commentCaptchaThreshold = 5  // 每用户每 60 秒评论超过该次数要求验证码

	forgetWindowSec        = 3600 // 忘记密码：每 IP 每小时
	forgetCaptchaThreshold = 3    // 超过该次数要求验证码（防邮箱轰炸）

	changeWindowSec        = 600 // 改密验证码爆破：10 分钟窗口
	changeCaptchaThreshold = 5   // 错 5 次作废验证码，需重发

	activateWindowSec = 600 // 激活码爆破：10 分钟窗口（覆盖 5 分钟码期）
	activateMaxFail   = 5   // 错 5 次作废激活码，需重发

	twoFaPendingSec = 300 // 2FA 待确认 5 分钟
)

// ---- failRecord 通用读写 ----

func getFailRecord(key string) *failRecord {
	s, err := rds.Get(key)
	if err != nil || s == "" {
		return nil
	}
	var r failRecord
	if json.Unmarshal([]byte(s), &r) != nil {
		return nil
	}
	return &r
}

func setFailRecord(key string, r *failRecord, ttlSec int) {
	b, _ := json.Marshal(r)
	_ = rds.SetEx(key, string(b), ttlSec)
}

func delFailRecord(key string) {
	_ = rds.Del(key)
}

func loginFailKey(ip, name string) string { return "ff_seclogin:" + ip + ":" + name }

// checkLoginBlock 返回 (needCaptcha, blocked, remainSec)。
func checkLoginBlock(ip, name string) (needCaptcha, blocked bool, remain int64) {
	key := loginFailKey(ip, name)
	r := getFailRecord(key)
	if r == nil {
		return false, false, 0
	}
	now := time.Now().Unix()
	if r.BlockUntil > now {
		return false, true, r.BlockUntil - now
	}
	if now-r.FirstTime > loginWindowSec {
		delFailRecord(key)
		return false, false, 0
	}
	return r.Count >= loginCaptchaThreshold, false, 0
}

func recordLoginFail(ip, name string) {
	key := loginFailKey(ip, name)
	r := getFailRecord(key)
	if r == nil {
		r = &failRecord{}
	}
	now := time.Now().Unix()
	r.Count++
	if r.FirstTime == 0 {
		r.FirstTime = now
	}
	if r.Count >= loginBlockThreshold {
		r.BlockUntil = now + loginBlockSec
	}
	// TTL 取 block 窗口，保证锁定信息不提前过期
	setFailRecord(key, r, loginBlockSec)
}

func clearLoginFail(ip, name string) {
	delFailRecord(loginFailKey(ip, name))
}

// ---- 注册 ----

func checkRegisterBlock(ip string) bool {
	key := "ff_secreg:" + ip
	r := getFailRecord(key)
	if r == nil {
		return false
	}
	now := time.Now().Unix()
	if now-r.FirstTime > registerWindowSec {
		delFailRecord(key)
		return false
	}
	return r.Count >= registerCaptchaThreshold
}

func recordRegister(ip string) {
	key := "ff_secreg:" + ip
	r := getFailRecord(key)
	if r == nil {
		r = &failRecord{}
	}
	now := time.Now().Unix()
	r.Count++
	if r.FirstTime == 0 {
		r.FirstTime = now
	}
	setFailRecord(key, r, registerWindowSec)
}

// ---- 评论 ----

func checkCommentBlock(userId int64) bool {
	key := "ff_seccomment:" + strconv.FormatInt(userId, 10)
	r := getFailRecord(key)
	if r == nil {
		return false
	}
	now := time.Now().Unix()
	if now-r.FirstTime > commentWindowSec {
		delFailRecord(key)
		return false
	}
	return r.Count >= commentCaptchaThreshold
}

func recordComment(userId int64) {
	key := "ff_seccomment:" + strconv.FormatInt(userId, 10)
	r := getFailRecord(key)
	if r == nil {
		r = &failRecord{}
	}
	now := time.Now().Unix()
	r.Count++
	if r.FirstTime == 0 {
		r.FirstTime = now
	}
	setFailRecord(key, r, commentWindowSec)
}

// ---- 忘记密码（防邮箱轰炸） ----

func checkForgetBlock(ip string) bool {
	key := "ff_secforget:" + ip
	r := getFailRecord(key)
	if r == nil {
		return false
	}
	now := time.Now().Unix()
	if now-r.FirstTime > forgetWindowSec {
		delFailRecord(key)
		return false
	}
	return r.Count >= forgetCaptchaThreshold
}

func recordForget(ip string) {
	key := "ff_secforget:" + ip
	r := getFailRecord(key)
	if r == nil {
		r = &failRecord{}
	}
	now := time.Now().Unix()
	r.Count++
	if r.FirstTime == 0 {
		r.FirstTime = now
	}
	setFailRecord(key, r, forgetWindowSec)
}

// ---- 改密验证码爆破 ----

func checkChangeBlock(email string) bool {
	key := "ff_secchange:" + email
	r := getFailRecord(key)
	if r == nil {
		return false
	}
	now := time.Now().Unix()
	if now-r.FirstTime > changeWindowSec {
		delFailRecord(key)
		return false
	}
	return r.Count >= changeCaptchaThreshold
}

func recordChangeFail(email string) {
	key := "ff_secchange:" + email
	r := getFailRecord(key)
	if r == nil {
		r = &failRecord{}
	}
	now := time.Now().Unix()
	r.Count++
	if r.FirstTime == 0 {
		r.FirstTime = now
	}
	setFailRecord(key, r, changeWindowSec)
}

func clearChangeFail(email string) {
	delFailRecord("ff_secchange:" + email)
}

// changeRemain 返回改密验证码剩余尝试次数（0 表示已到上限、码已作废）。
func changeRemain(email string) int {
	key := "ff_secchange:" + email
	r := getFailRecord(key)
	if r == nil {
		return changeCaptchaThreshold
	}
	now := time.Now().Unix()
	if now-r.FirstTime > changeWindowSec {
		delFailRecord(key)
		return changeCaptchaThreshold
	}
	remain := changeCaptchaThreshold - r.Count
	if remain < 0 {
		remain = 0
	}
	return remain
}

// ---- 激活码爆破 ----

func recordActivateFail(email string) {
	key := "ff_secactivate:" + email
	r := getFailRecord(key)
	if r == nil {
		r = &failRecord{}
	}
	now := time.Now().Unix()
	r.Count++
	if r.FirstTime == 0 {
		r.FirstTime = now
	}
	setFailRecord(key, r, activateWindowSec)
}

func clearActivateFail(email string) {
	delFailRecord("ff_secactivate:" + email)
}

// activateRemain 返回激活码剩余尝试次数（0 表示已到上限、码应作废）。
func activateRemain(email string) int {
	key := "ff_secactivate:" + email
	r := getFailRecord(key)
	if r == nil {
		return activateMaxFail
	}
	now := time.Now().Unix()
	if now-r.FirstTime > activateWindowSec {
		delFailRecord(key)
		return activateMaxFail
	}
	remain := activateMaxFail - r.Count
	if remain < 0 {
		remain = 0
	}
	return remain
}

// ---- 2FA 登录待确认 ----

func newTwoFaPending(userId int64) string {
	token := util.GetGUID()
	_ = rds.SetEx("ff_2fa_pending:"+token, strconv.FormatInt(userId, 10), twoFaPendingSec)
	return token
}

func consumeTwoFaPending(token string) (int64, bool) {
	key := "ff_2fa_pending:" + token
	s, err := rds.Get(key)
	if err != nil || s == "" {
		return 0, false
	}
	_ = rds.Del(key)
	uid, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return uid, true
}

// ---- 2FA 绑定待确认秘钥 ----

func setTwoFaSecretPending(userId int64, secret string) {
	_ = rds.SetEx("ff_2fa_secret:"+strconv.FormatInt(userId, 10), secret, twoFaPendingSec)
}

func getTwoFaSecretPending(userId int64) (string, bool) {
	s, err := rds.Get("ff_2fa_secret:" + strconv.FormatInt(userId, 10))
	if err != nil || s == "" {
		return "", false
	}
	return s, true
}

func clearTwoFaSecretPending(userId int64) {
	_ = rds.Del("ff_2fa_secret:" + strconv.FormatInt(userId, 10))
}
