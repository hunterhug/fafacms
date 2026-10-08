package model

import (
	log "github.com/hunterhug/golog"

	"github.com/hunterhug/fafacms/core/util/secure"
)

// content 表纳入加解密的字段（含历史版本与举报理由）。
//
// 未纳入：seo / node_seo（URL 路由）、status / top / version / 各计数与时间（排序统计）。
// user_name 是用户名的冗余副本，一并加密；相关过滤已改为按 user_id。
const (
	contentFieldTitle       = "title"
	contentFieldPreTitle    = "pre_title"
	contentFieldDescribe    = "describe"
	contentFieldPreDescribe = "pre_describe"
	contentFieldUserName    = "user_name"

	contentHistoryFieldTitle    = "title"
	contentHistoryFieldDescribe = "describe"

	contentBadFieldReason = "reason"
)

// searchableContentFields 返回需要"加密 + 盲索引"的字段（供精确搜索）。
func (c *Content) searchableContentFields() []searchableField {
	return []searchableField{
		{&c.Title, &c.TitleBidx, contentFieldTitle},
		{&c.PreTitle, &c.PreTitleBidx, contentFieldPreTitle},
	}
}

// EncryptFields 加密文章正文与冗余副本，并同步标题盲索引。
func (c *Content) EncryptFields() error {
	for _, f := range c.searchableContentFields() {
		if *f.plain == "" || secure.IsCipher(*f.plain) {
			continue
		}
		bidx, err := BlindIndexOf(*f.plain)
		if err != nil {
			return err
		}
		enc, err := EncryptTableField("content", f.col, c.UserId, *f.plain)
		if err != nil {
			return err
		}
		*f.plain = enc
		*f.bidx = bidx
	}

	var err error
	if c.Describe, err = EncryptOptional("content", contentFieldDescribe, c.UserId, c.Describe); err != nil {
		return err
	}
	if c.PreDescribe, err = EncryptOptional("content", contentFieldPreDescribe, c.UserId, c.PreDescribe); err != nil {
		return err
	}
	if c.UserName, err = EncryptOptional("content", contentFieldUserName, c.UserId, c.UserName); err != nil {
		return err
	}
	return nil
}

// DecryptFields 解密文章字段。
func (c *Content) DecryptFields() error {
	var err error
	for _, f := range c.searchableContentFields() {
		if *f.plain, err = DecryptOptional("content", f.col, c.UserId, *f.plain); err != nil {
			return err
		}
	}
	if c.Describe, err = DecryptOptional("content", contentFieldDescribe, c.UserId, c.Describe); err != nil {
		return err
	}
	if c.PreDescribe, err = DecryptOptional("content", contentFieldPreDescribe, c.UserId, c.PreDescribe); err != nil {
		return err
	}
	if c.UserName, err = DecryptOptional("content", contentFieldUserName, c.UserId, c.UserName); err != nil {
		return err
	}
	return nil
}

func (c *Content) BeforeInsert() {
	if err := c.EncryptFields(); err != nil {
		log.Errorf("Content.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (c *Content) BeforeUpdate() {
	if err := c.EncryptFields(); err != nil {
		log.Errorf("Content.BeforeUpdate encrypt err: %s", err.Error())
	}
}

func (c *Content) AfterLoad() {
	if err := c.DecryptFields(); err != nil {
		log.Errorf("Content.AfterLoad decrypt err: %s", err.Error())
	}
}

// AfterInsert / AfterUpdate：写入后把内存恢复为明文，避免调用方拿到密文。
func (c *Content) AfterInsert() {
	if err := c.DecryptFields(); err != nil {
		log.Errorf("Content.AfterInsert decrypt err: %s", err.Error())
	}
}

func (c *Content) AfterUpdate() {
	if err := c.DecryptFields(); err != nil {
		log.Errorf("Content.AfterUpdate decrypt err: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// ContentHistory：历史版本同样含正文，最易漏
// ---------------------------------------------------------------------------

func (h *ContentHistory) EncryptFields() error {
	var err error
	if h.Title, err = EncryptOptional("content_history", contentHistoryFieldTitle, h.UserId, h.Title); err != nil {
		return err
	}
	if h.Describe, err = EncryptOptional("content_history", contentHistoryFieldDescribe, h.UserId, h.Describe); err != nil {
		return err
	}
	return nil
}

func (h *ContentHistory) DecryptFields() error {
	var err error
	if h.Title, err = DecryptOptional("content_history", contentHistoryFieldTitle, h.UserId, h.Title); err != nil {
		return err
	}
	if h.Describe, err = DecryptOptional("content_history", contentHistoryFieldDescribe, h.UserId, h.Describe); err != nil {
		return err
	}
	return nil
}

func (h *ContentHistory) BeforeInsert() {
	if err := h.EncryptFields(); err != nil {
		log.Errorf("ContentHistory.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (h *ContentHistory) AfterLoad() {
	if err := h.DecryptFields(); err != nil {
		log.Errorf("ContentHistory.AfterLoad decrypt err: %s", err.Error())
	}
}

func (h *ContentHistory) AfterInsert() {
	if err := h.DecryptFields(); err != nil {
		log.Errorf("ContentHistory.AfterInsert decrypt err: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// ContentBad：举报理由
// ---------------------------------------------------------------------------

func (b *ContentBad) EncryptFields() error {
	var err error
	if b.Reason, err = EncryptOptional("content_bad", contentBadFieldReason, b.UserId, b.Reason); err != nil {
		return err
	}
	return nil
}

func (b *ContentBad) DecryptFields() error {
	var err error
	if b.Reason, err = DecryptOptional("content_bad", contentBadFieldReason, b.UserId, b.Reason); err != nil {
		return err
	}
	return nil
}

func (b *ContentBad) BeforeInsert() {
	if err := b.EncryptFields(); err != nil {
		log.Errorf("ContentBad.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (b *ContentBad) AfterLoad() {
	if err := b.DecryptFields(); err != nil {
		log.Errorf("ContentBad.AfterLoad decrypt err: %s", err.Error())
	}
}

func (b *ContentBad) AfterInsert() {
	if err := b.DecryptFields(); err != nil {
		log.Errorf("ContentBad.AfterInsert decrypt err: %s", err.Error())
	}
}
