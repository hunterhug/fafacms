package model

import (
	"errors"
	"time"
)

// SiteConfig single row site config (title + subtitle + community intro + footer intro)
type SiteConfig struct {
	Id           int64  `json:"id" xorm:"bigint pk autoincr"`
	SiteTitle    string `json:"site_title" xorm:"varchar(200) notnull"`
	SiteSubtitle string `json:"site_subtitle" xorm:"varchar(200)"`
	SiteIntro    string `json:"site_intro" xorm:"TEXT"`
	FooterIntro  string `json:"footer_intro" xorm:"TEXT"`
	UpdateTime   int64  `json:"update_time"`
}

// FriendLink friend link shown in the footer
type FriendLink struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	Name       string `json:"name" xorm:"varchar(100) notnull"`
	Url        string `json:"url" xorm:"varchar(500) notnull"`
	SortNum    int    `json:"sort_num" xorm:"int notnull default(0)"`
	Hide       int    `json:"hide" xorm:"int notnull default(0) comment('0 show 1 hide')"`
	OpenNew    int    `json:"open_new" xorm:"int notnull default(1) comment('0 same 1 new window')"`
	CreateTime int64  `json:"create_time"`
}

var FriendLinkSortName = []string{"+sort_num", "=id"}

// DefaultSiteTitle default title when config not set
const DefaultSiteTitle = "花花世界"

// DefaultSiteSubtitle default subtitle（放在站点标题后面，用于浏览器标题栏等）
const DefaultSiteSubtitle = "发现更大的世界"

// DefaultSiteIntro default community intro（首页「社区公告」与发现页的那句话，不含站名）
const DefaultSiteIntro = "是一个围绕内容互动的社区：写文章、交朋友、发现世界"

// DefaultFooterIntro default footer intro when config not set
const DefaultFooterIntro = "花花世界 · 一个围绕内容互动的社区"

// GetSiteConfig return the single row site config, create default if missing
func GetSiteConfig() (*SiteConfig, error) {
	c := new(SiteConfig)
	has, err := FaFaRdb.Client.Where("1=1").Get(c)
	if err != nil {
		return nil, err
	}
	if !has {
		c = &SiteConfig{
			SiteTitle:    DefaultSiteTitle,
			SiteSubtitle: DefaultSiteSubtitle,
			SiteIntro:    DefaultSiteIntro,
			FooterIntro:  DefaultFooterIntro,
			UpdateTime:   time.Now().Unix(),
		}
		_, err = FaFaRdb.Client.InsertOne(c)
		if err != nil {
			return nil, err
		}
	}
	if c.SiteTitle == "" {
		c.SiteTitle = DefaultSiteTitle
	}
	if c.SiteSubtitle == "" {
		c.SiteSubtitle = DefaultSiteSubtitle
	}
	if c.SiteIntro == "" {
		c.SiteIntro = DefaultSiteIntro
	}
	if c.FooterIntro == "" {
		c.FooterIntro = DefaultFooterIntro
	}
	return c, nil
}

// UpdateSiteConfig update title and footer intro
// UpdateSiteConfig 更新站点配置：标题必填，副标题/社区介绍/页脚介绍留空则回落默认值。
func UpdateSiteConfig(siteTitle, siteSubtitle, siteIntro, footerIntro string) error {
	c, err := GetSiteConfig()
	if err != nil {
		return err
	}
	c.SiteTitle = siteTitle
	if siteSubtitle == "" {
		siteSubtitle = DefaultSiteSubtitle
	}
	c.SiteSubtitle = siteSubtitle
	if siteIntro == "" {
		siteIntro = DefaultSiteIntro
	}
	c.SiteIntro = siteIntro
	if footerIntro == "" {
		footerIntro = DefaultFooterIntro
	}
	c.FooterIntro = footerIntro
	c.UpdateTime = time.Now().Unix()
	_, err = FaFaRdb.Client.Where("id=?", c.Id).
		Cols("site_title", "site_subtitle", "site_intro", "footer_intro", "update_time").Update(c)
	return err
}

func (f *FriendLink) InsertOne() error {
	f.CreateTime = time.Now().Unix()
	_, err := FaFaRdb.Client.InsertOne(f)
	return err
}

func (f *FriendLink) Get() (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}
	return FaFaRdb.Client.Get(f)
}

func (f *FriendLink) Update() error {
	if f.Id == 0 {
		return errors.New("where is empty")
	}
	_, err := FaFaRdb.Client.Where("id=?", f.Id).Cols("name", "url", "sort_num", "hide", "open_new").Update(f)
	return err
}

// UpdateSortNum 仅更新排序号（不得把 name/url 等字段误清空）
func (f *FriendLink) UpdateSortNum() error {
	if f.Id == 0 {
		return errors.New("where is empty")
	}
	_, err := FaFaRdb.Client.Where("id=?", f.Id).Cols("sort_num").Update(f)
	return err
}

// UpdateHide 仅更新隐藏状态
func (f *FriendLink) UpdateHide() error {
	if f.Id == 0 {
		return errors.New("where is empty")
	}
	_, err := FaFaRdb.Client.Where("id=?", f.Id).Cols("hide").Update(f)
	return err
}

func (f *FriendLink) Delete() error {
	if f.Id == 0 {
		return errors.New("where is empty")
	}
	_, err := FaFaRdb.Client.Where("id=?", f.Id).Delete(new(FriendLink))
	return err
}

// ListVisibleFriendLink list friend links not hidden, ordered by sort_num
func ListVisibleFriendLink() ([]FriendLink, error) {
	fs := make([]FriendLink, 0)
	err := FaFaRdb.Client.Where("hide=?", 0).Asc("sort_num").Asc("id").Find(&fs)
	return fs, err
}
