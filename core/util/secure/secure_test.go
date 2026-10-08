package secure

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

// newTestKeyring 构造测试用 Keyring。
func newTestKeyring(t *testing.T) *Keyring {
	t.Helper()
	root := bytes.Repeat([]byte{0x42}, KeySize)
	k, err := NewKeyring(root)
	if err != nil {
		t.Fatalf("NewKeyring: %v", err)
	}
	return k
}

// ---------------------------------------------------------------------------
// HKDF：RFC 5869 官方测试向量
// ---------------------------------------------------------------------------

func TestHKDF_RFC5869_Case1(t *testing.T) {
	ikm := bytes.Repeat([]byte{0x0b}, 22)
	salt, _ := hex.DecodeString("000102030405060708090a0b0c")
	info, _ := hex.DecodeString("f0f1f2f3f4f5f6f7f8f9")

	got := HKDF(ikm, salt, info, 42)
	want, _ := hex.DecodeString("3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865")

	if !bytes.Equal(got, want) {
		t.Errorf("HKDF 结果与 RFC 5869 测试向量不符\n got=%x\nwant=%x", got, want)
	}
}

func TestHKDF_RFC5869_Case3_EmptySaltAndInfo(t *testing.T) {
	// Case 3：salt 与 info 均为空（与本项目的用法一致：salt=nil, info=label）
	ikm := bytes.Repeat([]byte{0x0b}, 22)

	got := HKDF(ikm, nil, nil, 42)
	want, _ := hex.DecodeString("8da4e775a563c18f715f802a063c5a31b8a11f5c5ee1879ec3454e5f3c738d2d9d201395faa4b61a96c8")

	if !bytes.Equal(got, want) {
		t.Errorf("HKDF(空 salt/info) 与 RFC 5869 Case 3 不符\n got=%x\nwant=%x", got, want)
	}
}

// ---------------------------------------------------------------------------
// 密钥解析与派生
// ---------------------------------------------------------------------------

func TestParseKey(t *testing.T) {
	raw := bytes.Repeat([]byte{0x01}, KeySize)

	for name, encoded := range map[string]string{
		"base64":     base64.StdEncoding.EncodeToString(raw),
		"hex":        hex.EncodeToString(raw),
		"with space": "  " + base64.StdEncoding.EncodeToString(raw) + "  ",
	} {
		got, err := ParseKey(encoded)
		if err != nil {
			t.Errorf("%s: 解析失败: %v", name, err)
			continue
		}
		if !bytes.Equal(got, raw) {
			t.Errorf("%s: 解析结果不符", name)
		}
	}

	for name, bad := range map[string]string{
		"空":             "",
		"长度不足":          base64.StdEncoding.EncodeToString([]byte("short")),
		"非法字符":          "!!!not-a-key!!!",
		"合法 base64 但太短": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16)),
	} {
		if _, err := ParseKey(bad); err == nil {
			t.Errorf("%s: 期望报错但成功了", name)
		}
	}
}

func TestNewKeyring_RejectsWrongSize(t *testing.T) {
	if _, err := NewKeyring(bytes.Repeat([]byte{1}, 16)); err == nil {
		t.Error("16 字节根密钥应被拒绝")
	}
	if _, err := NewKeyring(bytes.Repeat([]byte{1}, KeySize)); err != nil {
		t.Errorf("32 字节根密钥应被接受: %v", err)
	}
}

func TestKeyring_SubkeysAreDistinctAndDeterministic(t *testing.T) {
	k1 := newTestKeyring(t)
	k2 := newTestKeyring(t)

	if bytes.Equal(k1.kWrap, k1.kIndex) || bytes.Equal(k1.kWrap, k1.kField) || bytes.Equal(k1.kIndex, k1.kField) {
		t.Error("三个子密钥必须互不相同")
	}
	if !bytes.Equal(k1.kWrap, k2.kWrap) {
		t.Error("相同根密钥必须派生出相同子密钥（确定性）")
	}
}

// ---------------------------------------------------------------------------
// 加解密
// ---------------------------------------------------------------------------

func TestSealOpen_RoundTrip(t *testing.T) {
	k := newTestKeyring(t)

	for _, plain := range []string{"", "a", "小明", "user@example.com", strings.Repeat("长文本", 1000)} {
		ct, err := Seal(k.kField, "aad", plain)
		if err != nil {
			t.Fatalf("Seal(%q): %v", plain, err)
		}
		if !strings.HasPrefix(ct, PrefixField) {
			t.Errorf("密文缺少前缀: %q", ct)
		}
		got, err := Open(k.kField, "aad", ct)
		if err != nil {
			t.Fatalf("Open(%q): %v", plain, err)
		}
		if got != plain {
			t.Errorf("往返不一致: got=%q want=%q", got, plain)
		}
	}
}

func TestSeal_RandomNonce(t *testing.T) {
	k := newTestKeyring(t)
	plain := "same-plaintext"

	a, _ := Seal(k.kField, "", plain)
	b, _ := Seal(k.kField, "", plain)
	if a == b {
		t.Error("同一明文两次加密必须得到不同密文（随机 nonce）")
	}

	// 两次都能正确解密
	for _, ct := range []string{a, b} {
		if got, err := Open(k.kField, "", ct); err != nil || got != plain {
			t.Errorf("解密失败: got=%q err=%v", got, err)
		}
	}
}

func TestOpen_WrongKeyFails(t *testing.T) {
	k := newTestKeyring(t)
	ct, _ := Seal(k.kField, "", "secret")

	other, _ := NewKeyring(bytes.Repeat([]byte{0x99}, KeySize))
	if _, err := Open(other.kField, "", ct); err == nil {
		t.Error("用错误密钥解密应当失败")
	}
}

func TestOpen_TamperDetection(t *testing.T) {
	k := newTestKeyring(t)
	ct, _ := Seal(k.kField, "", "secret-value")

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ct, PrefixField))
	if err != nil {
		t.Fatal(err)
	}
	// 翻转密文正文中的一个字节（跳过 12 字节 nonce）
	raw[len(raw)-1] ^= 0x01
	tampered := PrefixField + base64.StdEncoding.EncodeToString(raw)

	if _, err := Open(k.kField, "", tampered); err == nil {
		t.Error("密文被篡改后解密应当失败（GCM 认证）")
	}
}

func TestOpen_AADMismatchFails(t *testing.T) {
	k := newTestKeyring(t)
	ct, _ := Seal(k.kField, "user:email:1", "a@b.com")

	for name, aad := range map[string]string{
		"空 AAD":    "",
		"字段名不同":    "user:nick_name:1",
		"用户 ID 不同": "user:email:2",
		"表名不同":     "content:email:1",
	} {
		if _, err := Open(k.kField, aad, ct); err == nil {
			t.Errorf("%s：AAD 不符时解密应当失败", name)
		}
	}
	// 正确的 AAD 可以解开
	if got, err := Open(k.kField, "user:email:1", ct); err != nil || got != "a@b.com" {
		t.Errorf("正确 AAD 解密失败: got=%q err=%v", got, err)
	}
}

func TestOpen_LegacyPlaintextPassThrough(t *testing.T) {
	k := newTestKeyring(t)
	// 无前缀的值视为历史明文，原样返回（便于灰度上线）
	got, err := Open(k.kField, "aad", "plaintext-legacy")
	if err != nil {
		t.Fatalf("历史明文应兼容读取: %v", err)
	}
	if got != "plaintext-legacy" {
		t.Errorf("got=%q", got)
	}
}

// ---------------------------------------------------------------------------
// 字段级加解密 + AAD 绑定
// ---------------------------------------------------------------------------

func TestEncryptDecryptField(t *testing.T) {
	k := newTestKeyring(t)

	ct, err := k.EncryptField("user", "email", 7, "u7@example.com")
	if err != nil {
		t.Fatal(err)
	}
	got, err := k.DecryptField("user", "email", 7, ct)
	if err != nil {
		t.Fatal(err)
	}
	if got != "u7@example.com" {
		t.Errorf("got=%q", got)
	}

	// 跨字段：把 email 的密文当 nick_name 解
	if _, err := k.DecryptField("user", "nick_name", 7, ct); err == nil {
		t.Error("跨字段解密应当失败")
	}
	// 跨用户：把用户 7 的密文当用户 8 的解
	if _, err := k.DecryptField("user", "email", 8, ct); err == nil {
		t.Error("跨用户解密应当失败")
	}
	// 跨表
	if _, err := k.DecryptField("content", "email", 7, ct); err == nil {
		t.Error("跨表解密应当失败")
	}
}

func TestFieldAAD(t *testing.T) {
	if got, want := FieldAAD("user", "email", 5), "user:email:5"; got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}

// ---------------------------------------------------------------------------
// 盲索引
// ---------------------------------------------------------------------------

func TestBlindIndex_DeterministicAndNormalized(t *testing.T) {
	k := newTestKeyring(t)

	a := k.BlindIndex("User@Example.com")
	b := k.BlindIndex("  user@example.com  ")
	c := k.BlindIndex("USER@EXAMPLE.COM")

	if a != b || b != c {
		t.Errorf("规范化后应得到相同盲索引:\n a=%s\n b=%s\n c=%s", a, b, c)
	}
	if len(a) != 64 {
		t.Errorf("盲索引应为 64 位 hex，实际 %d", len(a))
	}
	if strings.Contains(a, "@") {
		t.Error("盲索引不应包含明文")
	}
	if k.BlindIndex("other@example.com") == a {
		t.Error("不同明文应得到不同盲索引")
	}
}

func TestBlindIndex_DependsOnKey(t *testing.T) {
	k1 := newTestKeyring(t)
	k2, _ := NewKeyring(bytes.Repeat([]byte{0x77}, KeySize))

	if k1.BlindIndex("x@y.com") == k2.BlindIndex("x@y.com") {
		t.Error("不同密钥下同一明文的盲索引应当不同")
	}
}

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"  AbC  ": "abc",
		"":        "",
		"ABC":     "abc",
		"  a b  ": "a b",
	} {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q)=%q want=%q", in, got, want)
		}
	}
}

func TestCredentialDigest(t *testing.T) {
	k := newTestKeyring(t)

	code := "483920"
	d1 := k.CredentialDigest(code)
	d2 := k.CredentialDigest("  " + code + "  ")

	if d1 != d2 {
		t.Error("凭据摘要应忽略首尾空白（确定性）")
	}
	if d1 == k.CredentialDigest("000000") {
		t.Error("不同凭据应得到不同摘要")
	}
	// 与同字符串的盲索引不同（域隔离）
	if d1 == k.BlindIndex(code) {
		t.Error("凭据摘要应与盲索引用不同域，避免混用")
	}
	if len(d1) != 64 {
		t.Errorf("摘要应为 64 位 hex，实际 %d", len(d1))
	}
}

// ---------------------------------------------------------------------------
// 密钥包裹（DEK / FEK）
// ---------------------------------------------------------------------------

func TestWrapUnwrapKey(t *testing.T) {
	k := newTestKeyring(t)

	dek, err := NewKey()
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := k.WrapKey(dek)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(wrapped, PrefixWrapped) {
		t.Errorf("包裹结果缺少前缀 %s: %q", PrefixWrapped, wrapped)
	}
	got, err := k.UnwrapKey(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, dek) {
		t.Error("解包结果与原密钥不一致")
	}

	// 相同 DEK 两次包裹，密文不同但都能解出
	w2, _ := k.WrapKey(dek)
	if wrapped == w2 {
		t.Error("两次包裹结果不应相同（随机 nonce）")
	}
	if got2, err := k.UnwrapKey(w2); err != nil || !bytes.Equal(got2, dek) {
		t.Errorf("第二次解包失败: err=%v", err)
	}
}

func TestUnwrapKey_RejectsBadInput(t *testing.T) {
	k := newTestKeyring(t)

	if _, err := k.UnwrapKey("no-prefix-at-all"); err == nil {
		t.Error("缺少前缀应报错")
	}
	if _, err := k.UnwrapKey(PrefixWrapped + "!!!bad-base64!!!"); err == nil {
		t.Error("非法 base64 应报错")
	}

	other, _ := NewKeyring(bytes.Repeat([]byte{0x55}, KeySize))
	wrapped, _ := k.WrapKey(bytes.Repeat([]byte{0x11}, KeySize))
	if _, err := other.UnwrapKey(wrapped); err == nil {
		t.Error("用错误根密钥解包应当失败")
	}
}

func TestNewKey_UniqueAndSized(t *testing.T) {
	a, err := NewKey()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := NewKey()

	if len(a) != KeySize {
		t.Errorf("密钥长度应为 %d，实际 %d", KeySize, len(a))
	}
	if bytes.Equal(a, b) {
		t.Error("两次生成的密钥不应相同")
	}
}

// ---------------------------------------------------------------------------
// 全局 Keyring
// ---------------------------------------------------------------------------

func TestGlobalKeyring_InitFromEnv(t *testing.T) {
	// 未初始化时应报错
	SetKeyring(nil)
	if _, err := Get(); err == nil {
		t.Error("未初始化时 Get 应报错")
	}

	keyB64, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvKEKRoot, keyB64)
	usedDefault, err := InitFromEnv()
	if err != nil {
		t.Fatalf("InitFromEnv: %v", err)
	}
	if usedDefault {
		t.Error("显式设置环境变量时不应使用默认密钥")
	}
	if _, err := Get(); err != nil {
		t.Errorf("初始化后 Get 应成功: %v", err)
	}

	// 环境变量缺失 → 回退内置默认密钥，并标记 usedDefault
	t.Setenv(EnvKEKRoot, "")
	SetKeyring(nil)
	usedDefault, err = InitFromEnv()
	if err != nil {
		t.Fatalf("缺失环境变量时应回退默认密钥，实际报错: %v", err)
	}
	if !usedDefault {
		t.Error("缺失环境变量时应返回 usedDefault=true")
	}
	if _, err := Get(); err != nil {
		t.Errorf("回退默认密钥后 Get 应成功: %v", err)
	}
	fp, ferr := Fingerprint()
	if ferr != nil || len(fp) != 16 {
		t.Errorf("Fingerprint 应为 16 位 hex，实际 %q err=%v", fp, ferr)
	}

	// 环境变量非法
	t.Setenv(EnvKEKRoot, "not-a-valid-key")
	if _, err := InitFromEnv(); err == nil {
		t.Error("环境变量非法时应报错")
	}

	SetKeyring(nil)
}

// 默认密钥必须是合法的 32 字节密钥，且指纹稳定（换值会导致既有密文不可解）。
func TestDefaultKEKRoot_Valid(t *testing.T) {
	raw, err := ParseKey(DefaultKEKRoot)
	if err != nil {
		t.Fatalf("默认密钥非法: %v", err)
	}
	if len(raw) != KeySize {
		t.Fatalf("默认密钥长度应为 %d，实际 %d", KeySize, len(raw))
	}
	if _, err := NewKeyring(raw); err != nil {
		t.Fatalf("默认密钥无法派生: %v", err)
	}
}

func TestGenerateKey_RoundTrip(t *testing.T) {
	s, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := ParseKey(s)
	if err != nil {
		t.Fatalf("生成的密钥应能被解析: %v", err)
	}
	if len(raw) != KeySize {
		t.Errorf("长度应为 %d", KeySize)
	}
}
