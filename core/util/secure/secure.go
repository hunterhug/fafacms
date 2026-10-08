// Package secure 提供字段级加密、盲索引与密钥包裹能力。
//
// # 密钥体系（信封加密）
//
//	KEK_root（环境变量 FAFACMS_KEK_ROOT，32 字节）
//	  ├─ HKDF("hunterhug:kek:wrap")  → K_wrap   包裹每用户 DEK / 每文件 FEK
//	  ├─ HKDF("hunterhug:kek:index") → K_index  盲索引 HMAC
//	  └─ HKDF("hunterhug:kek:field") → K_field  字段加密
//
// KEK_root 只来自环境变量，绝不入库、不进仓库；子密钥仅驻留内存。
//
// # 密文格式
//
//	字段密文：  v1:base64( nonce[12] ‖ ciphertext ‖ tag[16] )
//	包裹的密钥：kekv1:base64( nonce[12] ‖ ciphertext ‖ tag[16] )
//
// 无前缀的值视为历史明文（兼容读取），便于灰度上线。
//
// # 设计要点
//
//   - AES-256-GCM：对称加密 + 完整性认证（AEAD），篡改任意字节即解密失败
//   - 每次加密使用随机 nonce，同一明文两次加密结果不同
//   - 字段身份通过 AAD（附加认证数据）绑定，密文跨字段/跨用户搬移会解密失败
//   - 盲索引 = HMAC(K_index, 规范化明文)，确定性，可用于等值查询与唯一约束
package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	// EnvKEKRoot 根密钥所在的环境变量名
	EnvKEKRoot = "FAFACMS_KEK_ROOT"

	// 密文前缀（用于版本分派，便于将来轮换算法）
	PrefixField   = "v1:"
	PrefixWrapped = "kekv1:"

	// 用途分离 label
	labelWrap  = "hunterhug:kek:wrap"
	labelIndex = "hunterhug:kek:index"
	labelField = "hunterhug:kek:field"

	// KeySize 对称密钥长度（AES-256）
	KeySize = 32
)

// DefaultKEKRoot 是**未设置环境变量 FAFACMS_KEK_ROOT 时使用的内置默认根密钥**。
//
// ⚠️ 安全警告（务必阅读）：
// 本值随源码公开，不构成任何密钥保护。任何"忘记配置 FAFACMS_KEK_ROOT"的实例，
// 其数据库/备份一旦泄漏，任何人都可以用本值离线解密全部密文——等同于没有加密。
// 它存在的唯一意义是：让本地开发、临时演示、CI 免配置即可跑起来。
//
// 生产环境必须通过环境变量 FAFACMS_KEK_ROOT（或用 `./deploy.sh`，它会自动生成并
// 持久化到 $DATA_DIR/secret/kek.key）提供一把真实随机密钥。
const DefaultKEKRoot = "ZmFmYWNtcy1kZWZhdWx0LWtlay1pbnNlY3VyZS0zMmI="

// ---------------------------------------------------------------------------
// HKDF（RFC 5869）
// ---------------------------------------------------------------------------

// hkdfExtract RFC 5869 §2.2。salt 为空时使用全 0。
func hkdfExtract(secret, salt []byte) []byte {
	if len(salt) == 0 {
		salt = make([]byte, sha256.Size)
	}
	m := hmac.New(sha256.New, salt)
	m.Write(secret)
	return m.Sum(nil)
}

// hkdfExpand RFC 5869 §2.3
func hkdfExpand(prk []byte, info []byte, length int) []byte {
	var (
		okm []byte
		t   []byte
	)
	for i := byte(1); len(okm) < length; i++ {
		m := hmac.New(sha256.New, prk)
		m.Write(t)
		m.Write(info)
		m.Write([]byte{i})
		t = m.Sum(nil)
		okm = append(okm, t...)
	}
	return okm[:length]
}

// HKDF 完整实现，便于用 RFC 5869 官方测试向量验证。
func HKDF(secret, salt, info []byte, length int) []byte {
	return hkdfExpand(hkdfExtract(secret, salt), info, length)
}

// ---------------------------------------------------------------------------
// Keyring
// ---------------------------------------------------------------------------

// Keyring 持有根密钥与派生出的子密钥。
type Keyring struct {
	kWrap  []byte // 包裹 DEK / FEK
	kIndex []byte // 盲索引、凭据摘要
	kField []byte // 字段加密
}

// NewKeyring 由根密钥派生子密钥。
func NewKeyring(root []byte) (*Keyring, error) {
	if len(root) != KeySize {
		return nil, fmt.Errorf("根密钥长度必须为 %d 字节，实际 %d", KeySize, len(root))
	}
	return &Keyring{
		kWrap:  HKDF(root, nil, []byte(labelWrap), KeySize),
		kIndex: HKDF(root, nil, []byte(labelIndex), KeySize),
		kField: HKDF(root, nil, []byte(labelField), KeySize),
	}, nil
}

// ParseKey 解析密钥文本：支持 base64 或 hex，长度必须为 32 字节。
func ParseKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("密钥为空")
	}
	// 依次尝试 base64 标准编码、base64 无填充、hex
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == KeySize {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil && len(b) == KeySize {
		return b, nil
	}
	if b, err := hex.DecodeString(s); err == nil && len(b) == KeySize {
		return b, nil
	}
	return nil, fmt.Errorf("密钥格式无效（需 32 字节的 base64 或 hex）")
}

// GenerateKey 生成一把新的随机密钥（base64），用于初始化环境变量。
func GenerateKey() (string, error) {
	b := make([]byte, KeySize)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// NewKey 生成 32 字节随机密钥（用于 DEK / FEK）。
func NewKey() ([]byte, error) {
	b := make([]byte, KeySize)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// ---------------------------------------------------------------------------
// 全局 Keyring
// ---------------------------------------------------------------------------

var (
	mu      sync.RWMutex
	current *Keyring
)

// SetKeyring 设置全局 Keyring。
func SetKeyring(k *Keyring) {
	mu.Lock()
	current = k
	mu.Unlock()
}

// Get 返回全局 Keyring；未初始化时返回错误。
func Get() (*Keyring, error) {
	mu.RLock()
	k := current
	mu.RUnlock()
	if k == nil {
		return nil, fmt.Errorf("secure: 密钥未初始化，请设置环境变量 %s", EnvKEKRoot)
	}
	return k, nil
}

// InitFromEnv 从环境变量读取根密钥并初始化全局 Keyring。
//
// 未设置 FAFACMS_KEK_ROOT 时回退到内置默认密钥 DefaultKEKRoot，
// 并返回 usedDefault=true 供调用方打出醒目警告；环境变量存在但非法则报错。
func InitFromEnv() (usedDefault bool, err error) {
	raw := strings.TrimSpace(os.Getenv(EnvKEKRoot))
	if raw == "" {
		raw = DefaultKEKRoot
		usedDefault = true
	}
	root, err := ParseKey(raw)
	if err != nil {
		if usedDefault {
			return false, fmt.Errorf("内置默认密钥无效（这不应发生）：%w", err)
		}
		return false, fmt.Errorf("环境变量 %s 无效：%w", EnvKEKRoot, err)
	}
	k, err := NewKeyring(root)
	if err != nil {
		return usedDefault, err
	}
	SetKeyring(k)
	return usedDefault, nil
}

// Fingerprint 返回当前根密钥的短指纹（派生密钥的 SHA-256 前 8 字节 hex）。
//
// 用途：运维可在日志里比对"线上用的是哪把密钥"，而指纹不可反推密钥本身。
func Fingerprint() (string, error) {
	k, err := Get()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(k.kField)
	return hex.EncodeToString(sum[:8]), nil
}

// InitWithKey 用给定根密钥初始化全局 Keyring（测试用）。
func InitWithKey(root []byte) error {
	k, err := NewKeyring(root)
	if err != nil {
		return err
	}
	SetKeyring(k)
	return nil
}

// ---------------------------------------------------------------------------
// 规范化与盲索引
// ---------------------------------------------------------------------------

// Normalize 规范化用于盲索引的值。
//
// 依据：现有列排序规则为 utf8mb4_general_ci（大小写不敏感），
// 当前 SQL 的等值查询与唯一约束已把 "Abc" 与 "abc" 视为同一值，
// 因此这里统一 TrimSpace + ToLower 以保持一致。
func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// hmacHex 计算 HMAC-SHA256 的十六进制摘要。
func hmacHex(key []byte, msg string) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))
}

// BlindIndex 计算盲索引：HMAC(K_index, 规范化明文)。
// 确定性：同一明文永远得到同一结果，可用于等值查询与唯一约束。
func (k *Keyring) BlindIndex(plain string) string {
	return hmacHex(k.kIndex, Normalize(plain))
}

// NormalizedBlindIndex 对已规范化的值直接算盲索引（调用方自行规范化时使用）。
func (k *Keyring) NormalizedBlindIndex(normalized string) string {
	return hmacHex(k.kIndex, normalized)
}

// CredentialDigest 计算短凭据（激活码/重置码等）的摘要。
// 与盲索引共用 K_index，但通过独立的域前缀做用途隔离；不可逆，只用于校验。
func (k *Keyring) CredentialDigest(plain string) string {
	return hmacHex(k.kIndex, "hunterhug:cred:"+strings.TrimSpace(plain))
}

// ---------------------------------------------------------------------------
// 对称加解密
// ---------------------------------------------------------------------------

// Seal 用 key 加密 plain，aad 为附加认证数据（可为空）。
func Seal(key []byte, aad, plain string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	var aadBytes []byte
	if aad != "" {
		aadBytes = []byte(aad)
	}
	ct := gcm.Seal(nil, nonce, []byte(plain), aadBytes)
	return PrefixField + base64.StdEncoding.EncodeToString(append(nonce, ct...)), nil
}

// Open 用 key 解密 stored，aad 必须与加密时一致。
// 无前缀的值视为历史明文原样返回（兼容灰度上线）。
func Open(key []byte, aad, stored string) (string, error) {
	if !strings.HasPrefix(stored, PrefixField) {
		return stored, nil // 历史明文
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, PrefixField))
	if err != nil {
		return "", fmt.Errorf("secure: 密文编码无效: %w", err)
	}
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	n := gcm.NonceSize()
	if len(raw) < n {
		return "", fmt.Errorf("secure: 密文长度不足")
	}
	var aadBytes []byte
	if aad != "" {
		aadBytes = []byte(aad)
	}
	pt, err := gcm.Open(nil, raw[:n], raw[n:], aadBytes)
	if err != nil {
		return "", fmt.Errorf("secure: 解密失败（密钥不符或数据被篡改）: %w", err)
	}
	return string(pt), nil
}

// IsCipher 判断一个值是否为本包产生的密文（带版本前缀）。
func IsCipher(s string) bool {
	return strings.HasPrefix(s, PrefixField)
}

// IsWrapped 判断一个值是否为本包产生的包裹密钥密文。
func IsWrapped(s string) bool {
	return strings.HasPrefix(s, PrefixWrapped)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("secure: 密钥长度必须为 %d 字节，实际 %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// ---------------------------------------------------------------------------
// 字段级加解密（K_field + AAD 绑定字段身份）
// ---------------------------------------------------------------------------

// FieldAAD 构造字段的附加认证数据：表名:字段名:用户ID。
// 作用：把密文绑定到它的字段与归属，跨字段或跨用户搬移密文将解密失败。
func FieldAAD(table, field string, uid int64) string {
	return fmt.Sprintf("%s:%s:%d", table, field, uid)
}

// EncryptField 用 K_field 加密字段值。
func (k *Keyring) EncryptField(table, field string, uid int64, plain string) (string, error) {
	return Seal(k.kField, FieldAAD(table, field, uid), plain)
}

// DecryptField 用 K_field 解密字段值。
func (k *Keyring) DecryptField(table, field string, uid int64, stored string) (string, error) {
	return Open(k.kField, FieldAAD(table, field, uid), stored)
}

// ---------------------------------------------------------------------------
// 密钥包裹（DEK / FEK）
// ---------------------------------------------------------------------------

// WrapKey 用 K_wrap 包裹一把密钥（DEK/FEK），返回可直接入库的字符串。
func (k *Keyring) WrapKey(key []byte) (string, error) {
	s, err := Seal(k.kWrap, "", base64.StdEncoding.EncodeToString(key))
	if err != nil {
		return "", err
	}
	return PrefixWrapped + strings.TrimPrefix(s, PrefixField), nil
}

// UnwrapKey 解包由 WrapKey 产生的密文。
func (k *Keyring) UnwrapKey(wrapped string) ([]byte, error) {
	if !strings.HasPrefix(wrapped, PrefixWrapped) {
		return nil, fmt.Errorf("secure: 包裹密钥格式无效（缺少 %s 前缀）", PrefixWrapped)
	}
	s := PrefixField + strings.TrimPrefix(wrapped, PrefixWrapped)
	b64, err := Open(k.kWrap, "", s)
	if err != nil {
		return nil, err
	}
	key, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("secure: 解包内容无效: %w", err)
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("secure: 解包密钥长度异常: %d", len(key))
	}
	return key, nil
}
