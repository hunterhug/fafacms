package model

import (
	log "github.com/hunterhug/golog"
)

// file 表纳入加解密的字段。
//
// 未纳入的字段及原因：
//   - FileName / Url：随机存储名与对外访问路径，本身就出现在 URL 中（公开可见），
//     加密它们没有安全收益，且会导致下载时无法定位文件；
//   - Bidx / UrlHashCode：盲索引与 URL 摘要，是确定性摘要，非明文；
//   - Size / Type / Tag / Status / StoreType / 各时间：需要排序、过滤等。
const (
	fileFieldReallyName = "really_file_name"
	fileFieldDescribe   = "describe"
	fileFieldUserName   = "user_name"
)

// fileFieldOwner 返回该文件行的归属用户，用于 AAD 绑定。
// 上传时 UserId 已在 INSERT 之前设置，因此读写两侧一致。
func (f *File) fileFieldOwner() int64 {
	return f.UserId
}

// EncryptFields 加密文件元数据中的隐私字段（幂等）。
func (f *File) EncryptFields() error {
	owner := f.fileFieldOwner()
	var err error

	if f.ReallyFileName, err = EncryptOptional("file", fileFieldReallyName, owner, f.ReallyFileName); err != nil {
		return err
	}
	if f.Describe, err = EncryptOptional("file", fileFieldDescribe, owner, f.Describe); err != nil {
		return err
	}
	if f.UserName, err = EncryptOptional("file", fileFieldUserName, owner, f.UserName); err != nil {
		return err
	}
	return nil
}

// DecryptFields 解密文件元数据中的隐私字段。
func (f *File) DecryptFields() error {
	owner := f.fileFieldOwner()
	var err error

	if f.ReallyFileName, err = DecryptOptional("file", fileFieldReallyName, owner, f.ReallyFileName); err != nil {
		return err
	}
	if f.Describe, err = DecryptOptional("file", fileFieldDescribe, owner, f.Describe); err != nil {
		return err
	}
	if f.UserName, err = DecryptOptional("file", fileFieldUserName, owner, f.UserName); err != nil {
		return err
	}
	return nil
}

// BeforeInsert xorm 钩子：入库前加密。
func (f *File) BeforeInsert() {
	if err := f.EncryptFields(); err != nil {
		log.Errorf("File.BeforeInsert encrypt err: %s", err.Error())
	}
}

// BeforeUpdate xorm 钩子：更新前加密。
func (f *File) BeforeUpdate() {
	if err := f.EncryptFields(); err != nil {
		log.Errorf("File.BeforeUpdate encrypt err: %s", err.Error())
	}
}

// AfterLoad xorm 钩子：读取后解密。
func (f *File) AfterLoad() {
	if err := f.DecryptFields(); err != nil {
		log.Errorf("File.AfterLoad decrypt err: %s", err.Error())
	}
}

// AfterInsert / AfterUpdate 钩子：写入后把内存恢复为明文，
// 避免调用方写库后继续使用该结构体时拿到密文。
func (f *File) AfterInsert() {
	if err := f.DecryptFields(); err != nil {
		log.Errorf("File.AfterInsert decrypt err: %s", err.Error())
	}
}

func (f *File) AfterUpdate() {
	if err := f.DecryptFields(); err != nil {
		log.Errorf("File.AfterUpdate decrypt err: %s", err.Error())
	}
}
