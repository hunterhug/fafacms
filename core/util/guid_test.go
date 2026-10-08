package util

import (
	"encoding/hex"
	"fmt"
	"github.com/hunterhug/go_image"
	"path/filepath"
	"testing"
)

func TestListFile(t *testing.T) {
	// 输出到临时目录，避免污染仓库（原实现写 ./timg_x.jpeg 会反复 dirty 工作区）
	out := filepath.Join(t.TempDir(), "timg_x.jpeg")
	err := go_image.ScaleF2F("./timg.jpeg", out, 500)
	fmt.Printf("%#v", err)
}

func TestMd5(t *testing.T) {
	raw, err := ReadfromFile("./timg.jpeg")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(Md5(raw))

	raw, err = ReadfromFile("./timg.jpeg")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	d, _ := Sha256(raw)
	fmt.Println(len(d))

	fmt.Println(hex.EncodeToString([]byte("123456789")))
}
