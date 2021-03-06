package util

import (
	"encoding/hex"
	"fmt"
	"github.com/hunterhug/go_image"
	"testing"
)

func TestListFile(t *testing.T) {
	err := go_image.ScaleF2F("./timg.jpeg", "./timg_x.jpeg", 500)
	fmt.Printf("%#v", err)
}

func TestMd5(t *testing.T) {
	raw, err := ReadfromFile("./timg_x.jpeg")
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
