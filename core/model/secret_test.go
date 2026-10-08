package model

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hunterhug/fafacms/core/util/secure"
)

// initTestKeyring 为测试初始化加密密钥（幂等）。
func initTestKeyring(t *testing.T) {
	t.Helper()
	if _, err := secure.Get(); err == nil {
		return
	}
	if err := secure.InitWithKey(bytes.Repeat([]byte{0xAB}, secure.KeySize)); err != nil {
		t.Fatalf("初始化测试密钥失败: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 短凭据摘要（激活码 / 重置码）
// ---------------------------------------------------------------------------

func TestDigestSecretCode(t *testing.T) {
	initTestKeyring(t)

	d, err := DigestSecretCode("483920")
	if err != nil {
		t.Fatalf("DigestSecretCode: %v", err)
	}
	if len(d) != 64 {
		t.Errorf("摘要应为 64 位 hex，实际 %d: %s", len(d), d)
	}
	if strings.Contains(d, "483920") {
		t.Error("摘要不应包含明文")
	}

	// 确定性
	d2, _ := DigestSecretCode("483920")
	if d != d2 {
		t.Error("同一凭据的摘要应确定一致")
	}

	// 不同凭据 → 不同摘要
	d3, _ := DigestSecretCode("483921")
	if d == d3 {
		t.Error("不同凭据应得到不同摘要")
	}
}

func TestSecretCodeMatch(t *testing.T) {
	initTestKeyring(t)

	digest, _ := DigestSecretCode("123456")

	if !SecretCodeMatch(digest, "123456") {
		t.Error("正确凭据应匹配")
	}
	if SecretCodeMatch(digest, "123457") {
		t.Error("错误凭据不应匹配")
	}
	if SecretCodeMatch("", "123456") {
		t.Error("空摘要不应匹配")
	}
	if SecretCodeMatch(digest, "") {
		t.Error("空凭据不应匹配")
	}
}

// ---------------------------------------------------------------------------
// user 字段加解密（含 AAD 绑定）
// ---------------------------------------------------------------------------

func TestEncryptDecryptUserField(t *testing.T) {
	initTestKeyring(t)

	plain := "JBSWY3DPEHPK3PXP" // 典型 TOTP base32 密钥
	enc, err := EncryptUserField("two_fa_secret", 42, plain)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if !secure.IsCipher(enc) {
		t.Errorf("密文应带版本前缀: %q", enc)
	}
	if strings.Contains(enc, plain) {
		t.Error("密文不应包含明文")
	}

	got, err := DecryptUserField("two_fa_secret", 42, enc)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if got != plain {
		t.Errorf("往返不一致: got=%q want=%q", got, plain)
	}

	// AAD 绑定：换用户或换字段名都无法解密
	if _, err := DecryptUserField("two_fa_secret", 43, enc); err == nil {
		t.Error("换用户 ID 解密应当失败")
	}
	if _, err := DecryptUserField("email", 42, enc); err == nil {
		t.Error("换字段名解密应当失败")
	}
}

func TestEncryptUserField_EmptyValueRoundTrip(t *testing.T) {
	initTestKeyring(t)

	// 空字符串也可加密并还原（"关闭 2FA"由调用方直接存空值，不走加密）
	enc, err := EncryptUserField("two_fa_secret", 1, "")
	if err != nil {
		t.Fatalf("空值加密不应报错: %v", err)
	}
	got, err := DecryptUserField("two_fa_secret", 1, enc)
	if err != nil {
		t.Fatalf("空值解密失败: %v", err)
	}
	if got != "" {
		t.Errorf("应还原为空字符串，实际 %q", got)
	}
}
