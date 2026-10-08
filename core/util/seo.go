package util

import (
	"crypto/rand"
	"strings"
	"time"
)

// seoAlphabet 是随机 SEO 短码使用的字符集：小写字母 + 数字。
// 刻意去掉容易看错的 o / l / i 与 0 / 1，方便手抄和口头传达。
const seoAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// seoLetterNum 是字符集中字母部分的长度（首位只取字母，便于一眼看出是标识而不是数字）。
const seoLetterNum = 23

// GenSeo 生成 n 位随机 SEO 短码（首字符必为字母）。
// 用于文章与节点 SEO 的默认值：用户不填时由服务端生成，仍会走唯一性校验。
// 随机源是 crypto/rand；极端失败时退化为时间派生，保证不返回空串。
func GenSeo(n int) string {
	if n <= 0 {
		n = 8
	}
	return randomSeo(n, true)
}

// GenSeoSuffix 生成 n 位随机后缀，用于 SEO 撞名时自动加尾
// （例如把文章移动到已有同名 SEO 的节点）。允许数字开头。
func GenSeoSuffix(n int) string {
	if n <= 0 {
		n = 4
	}
	return randomSeo(n, false)
}

// randomSeo 从字符集取 n 个字符；letterFirst 为真时首位只取字母段。
func randomSeo(n int, letterFirst bool) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return fallbackSeo(n, letterFirst)
	}

	var b strings.Builder
	b.Grow(n)
	for i, v := range buf {
		limit := len(seoAlphabet)
		if letterFirst && i == 0 {
			limit = seoLetterNum
		}
		b.WriteByte(seoAlphabet[int(v)%limit])
	}
	return b.String()
}

// fallbackSeo crypto/rand 不可用时的兜底：用纳秒时间做线性同余，避免返回空串。
func fallbackSeo(n int, letterFirst bool) string {
	seed := uint64(time.Now().UnixNano())
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		seed = seed*6364136223846793005 + 1442695040888963407
		limit := len(seoAlphabet)
		if letterFirst && i == 0 {
			limit = seoLetterNum
		}
		b.WriteByte(seoAlphabet[int(seed>>33)%limit])
	}
	return b.String()
}
