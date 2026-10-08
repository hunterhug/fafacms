package oss

import (
	"fmt"
	"testing"
)

func TestSaveFile(t *testing.T) {
	// 本测试需要真实 OSS 凭据，不随仓库提交；如需本地验证请自行填 Key 后临时打开。
	t.Skip("skip: requires real OSS credentials, do not commit keys")
	k := Key{
		Endpoint:        "oss-cn-qingdao.aliyuncs.com",
		AccessKeyId:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		BucketName:      "your-bucket",
	}
	err := SaveFile(k, "jj/xx/afsaf.jpg", []byte("ddddd"))
	fmt.Printf("%#v", err)
}
