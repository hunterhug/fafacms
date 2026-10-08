package model

import (
	log "github.com/hunterhug/golog"
)

// comment 表纳入加解密的字段。
//
// 说明：ContentTitle / ContentUserName / CommentUserName / RootCommentUserName
// 都是**冗余副本**（评论里带的文章标题与各级被回复者昵称），
// 只加密正文而漏掉它们，等于正文隐私仍然从副本泄漏。
const (
	commentFieldUserName            = "user_name"
	commentFieldContentTitle        = "content_title"
	commentFieldContentUserName     = "content_user_name"
	commentFieldCommentUserName     = "comment_user_name"
	commentFieldRootCommentUserName = "root_comment_user_name"
	commentFieldDescribe            = "describe"

	commentBadFieldReason = "reason"
)

// EncryptFields 加密评论正文与各级名称副本。
func (c *Comment) EncryptFields() error {
	var err error
	if c.UserName, err = EncryptOptional("comment", commentFieldUserName, c.UserId, c.UserName); err != nil {
		return err
	}
	if c.ContentTitle, err = EncryptOptional("comment", commentFieldContentTitle, c.UserId, c.ContentTitle); err != nil {
		return err
	}
	if c.ContentUserName, err = EncryptOptional("comment", commentFieldContentUserName, c.UserId, c.ContentUserName); err != nil {
		return err
	}
	if c.CommentUserName, err = EncryptOptional("comment", commentFieldCommentUserName, c.UserId, c.CommentUserName); err != nil {
		return err
	}
	if c.RootCommentUserName, err = EncryptOptional("comment", commentFieldRootCommentUserName, c.UserId, c.RootCommentUserName); err != nil {
		return err
	}
	if c.Describe, err = EncryptOptional("comment", commentFieldDescribe, c.UserId, c.Describe); err != nil {
		return err
	}
	return nil
}

// DecryptFields 解密评论正文与各级名称副本。
func (c *Comment) DecryptFields() error {
	var err error
	if c.UserName, err = DecryptOptional("comment", commentFieldUserName, c.UserId, c.UserName); err != nil {
		return err
	}
	if c.ContentTitle, err = DecryptOptional("comment", commentFieldContentTitle, c.UserId, c.ContentTitle); err != nil {
		return err
	}
	if c.ContentUserName, err = DecryptOptional("comment", commentFieldContentUserName, c.UserId, c.ContentUserName); err != nil {
		return err
	}
	if c.CommentUserName, err = DecryptOptional("comment", commentFieldCommentUserName, c.UserId, c.CommentUserName); err != nil {
		return err
	}
	if c.RootCommentUserName, err = DecryptOptional("comment", commentFieldRootCommentUserName, c.UserId, c.RootCommentUserName); err != nil {
		return err
	}
	if c.Describe, err = DecryptOptional("comment", commentFieldDescribe, c.UserId, c.Describe); err != nil {
		return err
	}
	return nil
}

func (c *Comment) BeforeInsert() {
	if err := c.EncryptFields(); err != nil {
		log.Errorf("Comment.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (c *Comment) BeforeUpdate() {
	if err := c.EncryptFields(); err != nil {
		log.Errorf("Comment.BeforeUpdate encrypt err: %s", err.Error())
	}
}

func (c *Comment) AfterLoad() {
	if err := c.DecryptFields(); err != nil {
		log.Errorf("Comment.AfterLoad decrypt err: %s", err.Error())
	}
}

func (c *Comment) AfterInsert() {
	if err := c.DecryptFields(); err != nil {
		log.Errorf("Comment.AfterInsert decrypt err: %s", err.Error())
	}
}

func (c *Comment) AfterUpdate() {
	if err := c.DecryptFields(); err != nil {
		log.Errorf("Comment.AfterUpdate decrypt err: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// CommentBad：举报理由
// ---------------------------------------------------------------------------

func (b *CommentBad) EncryptFields() error {
	var err error
	if b.Reason, err = EncryptOptional("comment_bad", commentBadFieldReason, b.UserId, b.Reason); err != nil {
		return err
	}
	return nil
}

func (b *CommentBad) DecryptFields() error {
	var err error
	if b.Reason, err = DecryptOptional("comment_bad", commentBadFieldReason, b.UserId, b.Reason); err != nil {
		return err
	}
	return nil
}

func (b *CommentBad) BeforeInsert() {
	if err := b.EncryptFields(); err != nil {
		log.Errorf("CommentBad.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (b *CommentBad) AfterLoad() {
	if err := b.DecryptFields(); err != nil {
		log.Errorf("CommentBad.AfterLoad decrypt err: %s", err.Error())
	}
}

func (b *CommentBad) AfterInsert() {
	if err := b.DecryptFields(); err != nil {
		log.Errorf("CommentBad.AfterInsert decrypt err: %s", err.Error())
	}
}
