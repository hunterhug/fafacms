package model

import (
	"errors"
	"github.com/hunterhug/fafacms/core/util"
	"time"
)

type File struct {
	Id     int64  `json:"id" xorm:"bigint pk autoincr"`
	Type   string `json:"type" xorm:"index"`
	Tag    string `json:"tag" xorm:"index"`
	UserId int64  `json:"user_id" xorm:"bigint index"`
	// UserName 是用户名的冗余副本，加密存储（不再参与查询，故去掉索引）
	UserName string `json:"user_name" xorm:"varchar(255)"`
	// FileName 是落盘的随机文件名，它同时出现在对外 URL 中（公开可见），
	// 因此保持明文：加密它而 URL 仍可见并无安全收益，反而妨碍路径定位。
	FileName string `json:"file_name"`
	// ReallyFileName 是用户上传时的原始文件名，可能含隐私，加密存储
	ReallyFileName string `json:"really_file_name" xorm:"TEXT"`
	// Bidx 是去重盲索引（原 hash_code）：HMAC(K_index, userName + ":" + sha256(明文))。
	// 语义与原来完全一致（同用户同内容只存一份），但不再泄漏明文内容指纹。
	Bidx string `json:"bidx,omitempty" xorm:"char(64) unique"`
	// Url 为对外访问路径（随机文件名），保持明文以便下载时定位
	Url         string `json:"url" xorm:"varchar(700)"`
	UrlHashCode string `json:"url_hash_code" xorm:"varchar(100) unique"`
	// Describe 加密存储
	Describe   string `json:"describe" xorm:"TEXT"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time,omitempty"`
	Status     int    `json:"status" xorm:"notnull default(0) comment('0 normal，1 hide but can use') TINYINT(1)"`
	StoreType  int    `json:"store_type" xorm:"notnull default(0) comment('0 local，1 oss') TINYINT(1)"`
	IsPicture  int    `json:"is_picture"`
	// Size 为明文原始大小（Range 计算需要），保持明文
	Size int64 `json:"size"`
	// SizeX 为缩略图的明文大小（缩略图与原图块几何不同，Range 计算需要）
	SizeX int64 `json:"size_x" xorm:"notnull default(0)"`

	// EncVersion 文件内容加密版本：0 = 未加密（含全部 OSS 文件）；1 = v1 分块加密
	EncVersion int    `json:"enc_version" xorm:"int notnull default(0)"`
	FekWrapped string `json:"-" xorm:"varchar(255)"` // 每文件 FEK 的包裹密文
	EncNonce   string `json:"-" xorm:"varchar(32)"`  // 分块加密的 base nonce（8 字节 hex）
}

var FileSortName = []string{"=id", "-create_time", "-update_time", "=user_id", "=type", "=tag", "=store_type", "=status", "=size"}

func (f *File) Exist() (bool, error) {
	if f.Id == 0 && f.Url == "" {
		return false, errors.New("where is empty")
	}
	s := FaFaRdb.Client.Table(f)
	s.Where("1=1")

	if f.Id != 0 {
		s.And("id=?", f.Id)
	}
	if f.Url != "" {
		h, err := util.Sha256([]byte(f.Url))
		if err != nil {
			return false, err
		}
		s.And("url_hash_code=?", h)
	}

	c, err := s.Where("is_picture=?", 1).Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (f *File) Get() (bool, error) {
	if f.Id == 0 && f.Url == "" && f.UrlHashCode == "" && f.Bidx == "" {
		return false, errors.New("where is empty")
	}

	if f.Url != "" {
		h, err := util.Sha256([]byte(f.Url))
		if err != nil {
			return false, err
		}
		f.UrlHashCode = h
		f.Url = ""
	}

	return FaFaRdb.Client.Get(f)
}

// UpdateHide 显式设置隐藏状态（hide=true 隐藏 / hide=false 取消隐藏），只动 status 列
func (f *File) UpdateHide(hide bool) (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := FaFaRdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id)
	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	f.Status = 1
	if !hide {
		f.Status = 0
	}
	s.Cols("status")
	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")

	_, err := s.Update(f)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateInfo 仅更新描述/标签（不触碰隐藏状态）
func (f *File) UpdateInfo() (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := FaFaRdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id)
	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	if f.Describe != "" {
		s.Cols("describe")
	}
	if f.Tag != "" {
		s.Cols("tag")
	}
	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")

	_, err := s.Update(f)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (f *File) UpdateStatus() (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := FaFaRdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id).Cols("status")

	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")

	_, err := s.Update(f)
	if err != nil {
		return false, err
	}

	return true, nil
}
