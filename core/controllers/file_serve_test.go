package controllers

import (
	"testing"
)

func TestParseRangeHeader(t *testing.T) {
	const size = int64(1000)

	cases := []struct {
		name      string
		header    string
		wantStart int64
		wantEnd   int64
		wantOK    bool
	}{
		{"完整区间", "bytes=0-99", 0, 99, true},
		{"无结束", "bytes=100-", 100, 999, true},
		{"后缀区间", "bytes=-100", 900, 999, true},
		{"后缀超过总长", "bytes=-5000", 0, 999, true},
		{"结束超过总长自动收敛", "bytes=900-5000", 900, 999, true},
		{"单字节", "bytes=0-0", 0, 0, true},
		{"末尾单字节", "bytes=999-999", 999, 999, true},
		{"多区间只取第一段", "bytes=0-9,20-29", 0, 9, true},

		{"起始越界", "bytes=1000-1200", 0, 0, false},
		{"起点大于终点", "bytes=100-50", 0, 0, false},
		{"缺少前缀", "0-99", 0, 0, false},
		{"空后缀", "bytes=-", 0, 0, false},
		{"非法数字", "bytes=abc-def", 0, 0, false},
		{"缺横线", "bytes=100", 0, 0, false},
	}

	for _, c := range cases {
		start, end, ok := parseRangeHeader(c.header, size)
		if ok != c.wantOK {
			t.Errorf("%s: ok=%v want=%v", c.name, ok, c.wantOK)
			continue
		}
		if ok && (start != c.wantStart || end != c.wantEnd) {
			t.Errorf("%s: got [%d,%d] want [%d,%d]", c.name, start, end, c.wantStart, c.wantEnd)
		}
	}

	// 总长为 0 时任何区间都不合法
	if _, _, ok := parseRangeHeader("bytes=0-10", 0); ok {
		t.Error("总长为 0 时应返回不合法")
	}
}

func TestContentTypeByPath(t *testing.T) {
	cases := map[string]string{
		"storage/u/image/a.jpg":        "image/jpeg",
		"storage/u/image/a.jpeg":       "image/jpeg",
		"storage/u/image/a.PNG":        "image/png",
		"storage/u/image/a.gif":        "image/gif",
		"storage/u/file/a.pdf":         "application/pdf",
		"storage/u/media/a.mp4":        "video/mp4",
		"storage/u/media/a.mp3":        "audio/mpeg",
		"storage/u/other/a.unknownext": "application/octet-stream",
	}
	for path, want := range cases {
		if got := contentTypeByPath(path); got != want {
			t.Errorf("%s: got=%q want=%q", path, got, want)
		}
	}
}
