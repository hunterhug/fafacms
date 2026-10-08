package model

import (
	log "github.com/hunterhug/golog"
)

// 消息表需要加密的字段。
//
// 注意 ContentTitle / CommentDescribe 是**冗余副本**（通知里带的文章标题、评论摘要），
// 若只加密正文而漏掉它们，文章标题仍会以明文躺在库里。
const (
	msgFieldSendMessage     = "send_message"
	msgFieldContentTitle    = "content_title"
	msgFieldCommentDescribe = "comment_describe"

	globalMsgFieldSendMessage = "send_message"
)

// dataOwnerId 返回该消息行的数据归属用户，用于 AAD 绑定（同一行读写在 AAD 上必须一致）。
//
// 三类消息的归属字段不同：
//   - 通知类（评论/点赞/关注等）：UserId 为触发者
//   - 私信：UserId 为空，SendUserId 为发送者
//   - 全局公告投递：UserId/SendUserId 均为空，ReceiveUserId 为接收者
func (m *Message) dataOwnerId() int64 {
	if m.UserId != 0 {
		return m.UserId
	}
	if m.SendUserId != 0 {
		return m.SendUserId
	}
	return m.ReceiveUserId
}

// EncryptFields 加密私信正文与冗余副本（幂等：已是密文则跳过）。
func (m *Message) EncryptFields() error {
	owner := m.dataOwnerId()
	var err error

	if m.SendMessage, err = EncryptOptional("message", msgFieldSendMessage, owner, m.SendMessage); err != nil {
		return err
	}
	if m.ContentTitle, err = EncryptOptional("message", msgFieldContentTitle, owner, m.ContentTitle); err != nil {
		return err
	}
	if m.CommentDescribe, err = EncryptOptional("message", msgFieldCommentDescribe, owner, m.CommentDescribe); err != nil {
		return err
	}
	return nil
}

// DecryptFields 解密私信正文与冗余副本。
func (m *Message) DecryptFields() error {
	owner := m.dataOwnerId()
	var err error

	if m.SendMessage, err = DecryptOptional("message", msgFieldSendMessage, owner, m.SendMessage); err != nil {
		return err
	}
	if m.ContentTitle, err = DecryptOptional("message", msgFieldContentTitle, owner, m.ContentTitle); err != nil {
		return err
	}
	if m.CommentDescribe, err = DecryptOptional("message", msgFieldCommentDescribe, owner, m.CommentDescribe); err != nil {
		return err
	}
	return nil
}

// BeforeInsert xorm 钩子：入库前自动加密，避免遗漏某个写入路径。
func (m *Message) BeforeInsert() {
	if err := m.EncryptFields(); err != nil {
		log.Errorf("Message.BeforeInsert encrypt err: %s", err.Error())
	}
}

// AfterLoad xorm 钩子：从库中读出后自动解密，调用方无感。
func (m *Message) AfterLoad() {
	if err := m.DecryptFields(); err != nil {
		log.Errorf("Message.AfterLoad decrypt err: %s", err.Error())
	}
}

// AfterInsert / AfterUpdate 钩子：写入完成后把内存中的结构体恢复为明文，
// 避免调用方在写库后继续使用该结构体时拿到密文。
func (m *Message) AfterInsert() {
	if err := m.DecryptFields(); err != nil {
		log.Errorf("Message.AfterInsert decrypt err: %s", err.Error())
	}
}

func (m *Message) AfterUpdate() {
	if err := m.DecryptFields(); err != nil {
		log.Errorf("Message.AfterUpdate decrypt err: %s", err.Error())
	}
}

// EncryptFields 加密群发公告正文。
func (g *GlobalMessage) EncryptFields() error {
	var err error
	// 群发公告不归属任何用户，AAD 的归属 ID 固定为 0
	if g.SendMessage, err = EncryptOptional("global_message", globalMsgFieldSendMessage, 0, g.SendMessage); err != nil {
		return err
	}
	return nil
}

// DecryptFields 解密群发公告正文。
func (g *GlobalMessage) DecryptFields() error {
	var err error
	if g.SendMessage, err = DecryptOptional("global_message", globalMsgFieldSendMessage, 0, g.SendMessage); err != nil {
		return err
	}
	return nil
}

// BeforeInsert xorm 钩子：入库前自动加密。
func (g *GlobalMessage) BeforeInsert() {
	if err := g.EncryptFields(); err != nil {
		log.Errorf("GlobalMessage.BeforeInsert encrypt err: %s", err.Error())
	}
}

// AfterLoad xorm 钩子：读出后自动解密。
func (g *GlobalMessage) AfterLoad() {
	if err := g.DecryptFields(); err != nil {
		log.Errorf("GlobalMessage.AfterLoad decrypt err: %s", err.Error())
	}
}

// AfterInsert xorm 钩子：写入完成后恢复内存中的明文。
// 这样调用方在插入后拿到的 gm.SendMessage 仍是明文（扇出时可直接使用）。
func (g *GlobalMessage) AfterInsert() {
	if err := g.DecryptFields(); err != nil {
		log.Errorf("GlobalMessage.AfterInsert decrypt err: %s", err.Error())
	}
}

func (g *GlobalMessage) AfterUpdate() {
	if err := g.DecryptFields(); err != nil {
		log.Errorf("GlobalMessage.AfterUpdate decrypt err: %s", err.Error())
	}
}
