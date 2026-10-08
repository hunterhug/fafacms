package model

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// user 隐私字段：describe / short_describe / login_ip
// ---------------------------------------------------------------------------

func TestUser_PrivacyFieldsEncryptedAtRest(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	bio := "个人简介：喜欢摄影与写作 user-bio-secret"
	shortBio := "短签名-隐私检测"
	ip := "203.0.113.77"

	u := &User{
		Name:          "__uc_test_" + suffix,
		NickName:      "__uc_nick_" + suffix,
		Email:         "__uc_" + suffix + "@example.com",
		Password:      "test123456",
		Status:        1,
		Describe:      bio,
		ShortDescribe: shortBio,
		LoginIp:       ip,
	}
	// 注意：此处 Id 仍为 0，加密发生在 INSERT 之前（AAD 用固定归属 0）
	if err := u.InsertOne(); err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_user WHERE id=%d", u.Id))
	}()

	if u.Id == 0 {
		t.Fatal("插入后应拿到自增 Id")
	}

	// ① 直接看库：三个字段都必须是密文
	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT `describe`, short_describe, login_ip FROM fafacms_user WHERE id=%d", u.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("原始查询失败: %v", err)
	}
	raw := rows[0]
	for field, want := range map[string]string{
		"describe":       bio,
		"short_describe": shortBio,
		"login_ip":       ip,
	} {
		got := raw[field]
		if got == want {
			t.Errorf("%s 不应明文落库", field)
		}
		if !strings.HasPrefix(got, "v1:") {
			t.Errorf("%s 应为密文，实际: %q", field, got)
		}
	}
	if strings.Contains(raw["describe"], "user-bio-secret") {
		t.Error("密文中不应残留明文片段")
	}

	// ② 通过模型读回：必须能解密（验证插入时 AAD=0 与读取时一致）
	got := new(User)
	exist, err := FaFaRdb.Client.ID(u.Id).Get(got)
	if err != nil || !exist {
		t.Fatalf("读取用户失败: exist=%v err=%v", exist, err)
	}
	if got.Describe != bio {
		t.Errorf("describe 未正确解密: got=%q want=%q", got.Describe, bio)
	}
	if got.ShortDescribe != shortBio {
		t.Errorf("short_describe 未正确解密: got=%q want=%q", got.ShortDescribe, shortBio)
	}
	if got.LoginIp != ip {
		t.Errorf("login_ip 未正确解密: got=%q want=%q", got.LoginIp, ip)
	}
}

func TestUser_UpdateLoginInfoEncryptsIp(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	u, cleanup := newTempUser(t)
	defer cleanup()

	newIp := "198.51.100.42"
	upd := &User{Id: u.Id, LoginIp: newIp, LoginTime: time.Now().Unix()}
	if err := upd.UpdateLoginInfo(); err != nil {
		t.Fatalf("UpdateLoginInfo: %v", err)
	}

	// 落库为密文
	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT login_ip FROM fafacms_user WHERE id=%d", u.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("原始查询失败: %v", err)
	}
	if got := rows[0]["login_ip"]; got == newIp || !strings.HasPrefix(got, "v1:") {
		t.Errorf("更新后 login_ip 应加密落库，实际: %q", got)
	}

	// 读回可解密
	reloaded := reloadUser(t, u.Id)
	if reloaded.LoginIp != newIp {
		t.Errorf("login_ip 解密失败: got=%q want=%q", reloaded.LoginIp, newIp)
	}
}

func TestUser_EmptyPrivacyFieldsStayEmpty(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	u, cleanup := newTempUser(t)
	defer cleanup()

	// newTempUser 只设了 ShortDescribe，未设 Describe/LoginIp
	reloaded := reloadUser(t, u.Id)
	if reloaded.Describe != "" {
		t.Errorf("未填写的 describe 应为空: %q", reloaded.Describe)
	}
	if reloaded.LoginIp != "" {
		t.Errorf("未填写的 login_ip 应为空: %q", reloaded.LoginIp)
	}

	// 更新一次不改动这些字段：不应被写成密文垃圾
	upd := &User{Id: u.Id, UpdateTime: time.Now().Unix()}
	if _, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("update_time").Update(upd); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	after := reloadUser(t, u.Id)
	if after.Describe != "" || after.LoginIp != "" {
		t.Errorf("未更新字段应保持为空: describe=%q login_ip=%q", after.Describe, after.LoginIp)
	}
}

func TestUser_EncryptFieldsIdempotent(t *testing.T) {
	initTestKeyring(t)

	u := &User{Id: 123, Describe: "hello", ShortDescribe: "hi", LoginIp: "1.2.3.4"}
	if err := u.EncryptFields(); err != nil {
		t.Fatal(err)
	}
	first := u.Describe
	// 再加密一次不应二次加密
	if err := u.EncryptFields(); err != nil {
		t.Fatal(err)
	}
	if u.Describe != first {
		t.Error("EncryptFields 应幂等，不应二次加密")
	}
	// 解密后应还原
	if err := u.DecryptFields(); err != nil {
		t.Fatal(err)
	}
	if u.Describe != "hello" || u.ShortDescribe != "hi" || u.LoginIp != "1.2.3.4" {
		t.Errorf("解密失败: %q %q %q", u.Describe, u.ShortDescribe, u.LoginIp)
	}
}
