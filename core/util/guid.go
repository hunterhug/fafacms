package util

/*
//const char* build_time(void)
//{
//static const char* psz_build_time = ""__DATE__ "-" __TIME__ "";
//return psz_build_time;
//}
*/
//import "C"

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	uuid "github.com/gofrs/uuid"
	"strings"
	"time"
)

//func BuildTime() string {
//	// use gcc image to get build time will cause 102M image sizes, not worth, but i love it
//	s := C.GoString(C.build_time())
//	t, _ := time.ParseInLocation("Jan 2 2006-15:04:05", s, time.Local)
//	return t.UTC().Format("20060102_15:04:05_UTC")
//}

// GetGUID
func GetGUID() (valueGUID string) {
	objID, _ := uuid.NewV4()
	objIdStr := objID.String()
	objIdStr = strings.Replace(objIdStr, "-", "", -1)
	valueGUID = objIdStr
	return valueGUID
}

// sha256 256 bit
func Sha256(raw []byte) (string, error) {
	h := sha256.New()
	num, err := h.Write(raw)
	if err != nil {
		return "", err
	}
	if num == 0 {
		return "", errors.New("num 0")
	}
	data := h.Sum([]byte(""))
	return fmt.Sprintf("%x", data), nil
}

func Md5(raw []byte) (string, error) {
	h := md5.New()
	num, err := h.Write(raw)
	if err != nil {
		return "", err
	}
	if num == 0 {
		return "", errors.New("num 0")
	}
	data := h.Sum([]byte("hunterhug"))
	return fmt.Sprintf("%x", data), nil
}

// GenCode6 生成 6 位纯数字短码（含前导零，如 "001234"），
// 用于邮箱激活码 / 忘记密码重置码。安全靠「错 5 次作废 + 5 分钟过期」兜底。
func GenCode6() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand 几乎不会失败；兜底用纳秒时间，避免返回空码
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", binary.BigEndian.Uint32(b[:])%1000000)
}
