package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/hunterhug/fafacms/core/util/secure"
)

// newTempUser 创建一个临时用户用于集成测试，返回清理函数。
func newTempUser(t *testing.T) (*User, func()) {
	t.Helper()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	u := &User{
		Name:          "__enc_test_" + suffix,
		NickName:      "__enc_test_nick_" + suffix,
		Email:         "__enc_test_" + suffix + "@example.com",
		Password:      "test123456",
		Status:        1,
		ShortDescribe: "encryption integration test",
	}
	if err := u.InsertOne(); err != nil {
		t.Fatalf("创建临时用户失败: %v", err)
	}
	cleanup := func() {
		if _, err := FaFaRdb.Client.Exec("DELETE FROM fafacms_user WHERE id=?", u.Id); err != nil {
			t.Logf("清理临时用户失败: %v", err)
		}
	}
	return u, cleanup
}

// reloadUser 从库中重新读取用户，用于检查落库形态。
func reloadUser(t *testing.T, id int64) *User {
	t.Helper()
	got := &User{Id: id}
	exist, err := FaFaRdb.Client.Get(got)
	if err != nil {
		t.Fatalf("读取用户失败: %v", err)
	}
	if !exist {
		t.Fatalf("用户不存在: id=%d", id)
	}
	return got
}

// ---------------------------------------------------------------------------
// 2FA 秘钥：加密落库
// ---------------------------------------------------------------------------

func TestUser_TwoFaSecret_EncryptedAtRest(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	u, cleanup := newTempUser(t)
	defer cleanup()

	secret := "JBSWY3DPEHPK3PXP"

	// 绑定 2FA
	upd := &User{Id: u.Id, TwoFaSecret: secret}
	if err := upd.UpdateTwoFa(); err != nil {
		t.Fatalf("UpdateTwoFa: %v", err)
	}

	// 直接读库：必须是密文，不能是明文
	raw := reloadUser(t, u.Id)
	if raw.TwoFaSecret == secret {
		t.Fatal("库中不应存 2FA 明文秘钥")
	}
	if !secure.IsCipher(raw.TwoFaSecret) {
		t.Fatalf("库中 2FA 秘钥应为密文，实际: %q", raw.TwoFaSecret)
	}
	if len(raw.TwoFaSecret) > 255 {
		t.Errorf("密文长度超出列宽: %d", len(raw.TwoFaSecret))
	}

	// 解密还原
	plain, err := raw.TwoFaSecretPlain()
	if err != nil {
		t.Fatalf("TwoFaSecretPlain: %v", err)
	}
	if plain != secret {
		t.Errorf("解密结果不符: got=%q want=%q", plain, secret)
	}

	// 重复保存同一明文（幂等）：已是密文则不应二次加密
	again := &User{Id: u.Id, TwoFaSecret: raw.TwoFaSecret}
	if err := again.UpdateTwoFa(); err != nil {
		t.Fatalf("二次 UpdateTwoFa: %v", err)
	}
	raw2 := reloadUser(t, u.Id)
	if raw2.TwoFaSecret != raw.TwoFaSecret {
		t.Error("已是密文的值不应被二次加密")
	}
	if p, err := raw2.TwoFaSecretPlain(); err != nil || p != secret {
		t.Errorf("二次保存后解密应仍正确: got=%q err=%v", p, err)
	}
}

func TestUser_TwoFaSecret_DisableStoresEmpty(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	u, cleanup := newTempUser(t)
	defer cleanup()

	if err := (&User{Id: u.Id, TwoFaSecret: "JBSWY3DPEHPK3PXP"}).UpdateTwoFa(); err != nil {
		t.Fatal(err)
	}

	// 关闭 2FA：直接存空，不加密
	if err := (&User{Id: u.Id, TwoFaSecret: ""}).UpdateTwoFa(); err != nil {
		t.Fatalf("关闭 2FA 失败: %v", err)
	}
	raw := reloadUser(t, u.Id)
	if raw.TwoFaSecret != "" {
		t.Errorf("关闭后应为空，实际 %q", raw.TwoFaSecret)
	}
	if p, err := raw.TwoFaSecretPlain(); err != nil || p != "" {
		t.Errorf("关闭后解密应为空: got=%q err=%v", p, err)
	}
}

// ---------------------------------------------------------------------------
// 激活码 / 重置码：库中只存摘要
// ---------------------------------------------------------------------------

func TestUser_ActivateCode_StoredAsDigest(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	u, cleanup := newTempUser(t)
	defer cleanup()

	// 重新生成激活码：返回明文用于发信，库中存摘要
	code, err := u.UpdateActivateCode()
	if err != nil {
		t.Fatalf("UpdateActivateCode: %v", err)
	}
	if len(code) == 0 {
		t.Fatal("应返回明文激活码用于发信")
	}

	raw := reloadUser(t, u.Id)
	if raw.ActivateCode == code {
		t.Fatal("库中不应存激活码明文")
	}
	if len(raw.ActivateCode) != 64 {
		t.Errorf("激活码摘要应为 64 位 hex，实际 %d: %q", len(raw.ActivateCode), raw.ActivateCode)
	}

	// 正确码可匹配，错误码不可
	if !SecretCodeMatch(raw.ActivateCode, code) {
		t.Error("正确激活码应匹配")
	}
	if SecretCodeMatch(raw.ActivateCode, "000000") && code != "000000" {
		t.Error("错误激活码不应匹配")
	}
}

func TestUser_ResetCode_StoredAsDigest(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	u, cleanup := newTempUser(t)
	defer cleanup()

	code, err := u.UpdateCode()
	if err != nil {
		t.Fatalf("UpdateCode: %v", err)
	}

	raw := reloadUser(t, u.Id)
	if raw.ResetCode == code {
		t.Fatal("库中不应存重置码明文")
	}
	if !SecretCodeMatch(raw.ResetCode, code) {
		t.Error("正确重置码应匹配")
	}

	// 改密码后应清空重置码（一次性）
	raw.Password = "newpass123456"
	if err := raw.UpdatePassword(); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}
	after := reloadUser(t, u.Id)
	if after.ResetCode != "" {
		t.Errorf("改密后重置码应清空，实际 %q", after.ResetCode)
	}
}

// ---------------------------------------------------------------------------
// 内容访问密码：bcrypt 哈希（不可逆）
// ---------------------------------------------------------------------------

func TestContentPassword_BcryptRoundTrip(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	// 走 model 的哈希工具，确认存储形态不可逆且可校验
	hash, err := HashPassword("mypass123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "mypass123" {
		t.Fatal("不应存明文")
	}
	if hash[0] != '$' {
		t.Errorf("bcrypt 哈希应以 $ 开头: %q", hash)
	}

	if ok, _ := CheckPassword(hash, "mypass123"); !ok {
		t.Error("正确密码应校验通过")
	}
	if ok, _ := CheckPassword(hash, "wrong"); ok {
		t.Error("错误密码不应校验通过")
	}

	// 同一密码两次哈希不同（加盐）
	h2, _ := HashPassword("mypass123")
	if hash == h2 {
		t.Error("bcrypt 应加盐，两次哈希不应相同")
	}
}
