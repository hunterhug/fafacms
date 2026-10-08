package util

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
)

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Base64E(urlstring string) string {
	str := []byte(urlstring)
	data := base64.StdEncoding.EncodeToString(str)
	return data
}

func Base64D(urlxxstring string) string {
	data, err := base64.StdEncoding.DecodeString(urlxxstring)
	if err != nil {
		return ""
	}
	s := fmt.Sprintf("%q", data)
	s = strings.Replace(s, "\"", "", -1)
	return s
}

func UrlE(s string) string {
	return url.QueryEscape(s)
}

func UrlD(s string) string {
	s, e := url.QueryUnescape(s)
	if e != nil {
		return e.Error()
	} else {
		return s
	}
}

func Md5FS(src io.Reader) string {
	h := md5.New()
	if err := CopyFF(src, h); err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return fmt.Sprintf("%x", h.Sum([]byte("hunterhug")))
}
