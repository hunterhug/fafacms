package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// TOTP 标准实现（RFC 6238，HMAC-SHA1，6 位，30 秒，±1 时间步容差）。
// 兼容微软 Authenticator / Google Authenticator 等主流验证器。

// GenerateTOTPSecret 生成 20 字节随机秘钥，返回 base32（无填充）字符串。
func GenerateTOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// TOTPURI 生成 otpauth:// 协议 URI，供验证器扫码或手输绑定。
func TOTPURI(issuer, account, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, account, secret, issuer)
}

// ValidateTOTP 校验 6 位动态码（±1 时间步容差，即 ±30 秒）。
func ValidateTOTP(secret, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(
		strings.ToUpper(strings.TrimSpace(secret)),
	)
	if err != nil {
		return false
	}
	step := time.Now().Unix() / 30
	for _, off := range []int64{-1, 0, 1} {
		if totpCodeAt(key, step+off) == code {
			return true
		}
	}
	return false
}

func totpCodeAt(key []byte, counter int64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(counter))
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := int(sum[len(sum)-1] & 0x0f)
	binCode := int64(binary.BigEndian.Uint32(sum[offset:offset+4])) & 0x7fffffff
	return fmt.Sprintf("%06d", binCode%1000000)
}
