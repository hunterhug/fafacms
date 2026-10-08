package controllers

import (
	"strings"
	"testing"
)

func hasDanglingTag(s string) bool {
	return strings.LastIndex(s, "<") > strings.LastIndex(s, ">")
}

func TestBuildExcerpt(t *testing.T) {
	// 短正文不截断
	if got := BuildExcerpt("你好，世界", 200); got != "你好，世界" {
		t.Fatalf("short text should not be truncated, got %q", got)
	}

	// 按 rune 截断，中文不被切坏
	got := BuildExcerpt(strings.Repeat("中", 300), 200)
	if len([]rune(got)) != 200 {
		t.Fatalf("expect 200 runes, got %d", len([]rune(got)))
	}
	if strings.Contains(got, "\ufffd") {
		t.Fatalf("truncated text contains replacement char: %q", got)
	}

	// 截断点恰好落在 <img ... 标签中间时，半截标签与半截 URL 必须被丢掉
	dangling := strings.Repeat("字", 190) + "<p>" + `<img src="/storage/admin/image/87a001c7da0f459528d1770246c5f862.webp" alt="图片">`
	got = BuildExcerpt(dangling, 200)
	if hasDanglingTag(got) {
		t.Fatalf("dangling html tag should be dropped, got %q", got)
	}
	if strings.Contains(got, "<img") || strings.Contains(got, "/storage") {
		t.Fatalf("half image url should be dropped, got %q", got)
	}
	if !strings.Contains(got, "字") {
		t.Fatalf("normal text should be kept, got %q", got)
	}

	// 真实场景：正文前 200 字符里含完整图片标签，截断残留的半截标签同样被丢掉
	body := "sssK\n\n你好\n\n<p><img src=\"/storage/admin/image/b1b4c854bd4ea6bcc9a35d87e745ac6a.png\" alt=\"图片\" style=\"width: 261.00px; height: 261.00px\"></p>\n\n<p><img src=\"/storage/admin/image/87a001c7da0f459528d1770246c5f862.webp\" alt=\"图片\" style=\"width: 641.40px; height: 481.05px\"></p>"
	got = BuildExcerpt(body, 200)
	if hasDanglingTag(got) {
		t.Fatalf("dangling html tag should be dropped, got %q", got)
	}
	if strings.Contains(got, "/storage/admin/image/87a001c7") {
		t.Fatalf("half image url should be dropped, got %q", got)
	}
	if !strings.Contains(got, "你好") {
		t.Fatalf("normal text should be kept, got %q", got)
	}

	// 完整标签不在截断边界上时保持原样
	body2 := "abc <b>bold</b> def"
	if got = BuildExcerpt(body2, 200); got != body2 {
		t.Fatalf("complete html should be kept as is, got %q", got)
	}

	// maxRunes<=0 表示不截断
	long := strings.Repeat("a", 500)
	if got = BuildExcerpt(long, 0); got != long {
		t.Fatalf("maxRunes<=0 should keep all, got len %d", len(got))
	}
}
