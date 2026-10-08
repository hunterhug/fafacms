package util

import (
	"regexp"
	"strings"
	"testing"
)

// seoValid 与服务端 alphanumunicode 校验口径一致：只允许字母、数字、中文。
var seoValid = regexp.MustCompile("^[A-Za-z0-9\\x{4e00}-\\x{9fa5}]+$")

func TestGenSeo(t *testing.T) {
	// 默认长度与指定长度
	if got := GenSeo(0); len(got) != 8 {
		t.Fatalf("GenSeo(0) should fall back to 8 chars, got %q", got)
	}
	for _, n := range []int{4, 8, 12} {
		got := GenSeo(n)
		if len(got) != n {
			t.Fatalf("GenSeo(%d) length = %d, got %q", n, len(got), got)
		}
		if !seoValid.MatchString(got) {
			t.Fatalf("GenSeo(%d) = %q is not alphanumunicode", n, got)
		}
		// 首字符必须是字母
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyz", rune(got[0])) {
			t.Fatalf("GenSeo(%d) = %q should start with a letter", n, got)
		}
		// 不含易混字符
		if strings.ContainsAny(got, "ol i01") {
			t.Fatalf("GenSeo(%d) = %q contains confusing characters", n, got)
		}
	}
}

func TestGenSeoUnique(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		s := GenSeo(8)
		if _, dup := seen[s]; dup {
			t.Fatalf("GenSeo(8) duplicated value %q at %d", s, i)
		}
		seen[s] = struct{}{}
	}
}

func TestGenSeoSuffix(t *testing.T) {
	if got := GenSeoSuffix(0); len(got) != 4 {
		t.Fatalf("GenSeoSuffix(0) should fall back to 4 chars, got %q", got)
	}
	for _, n := range []int{3, 4, 6} {
		got := GenSeoSuffix(n)
		if len(got) != n {
			t.Fatalf("GenSeoSuffix(%d) length = %d, got %q", n, len(got), got)
		}
		if !seoValid.MatchString(got) {
			t.Fatalf("GenSeoSuffix(%d) = %q is not alphanumunicode", n, got)
		}
	}
}
