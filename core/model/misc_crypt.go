package model

import (
	log "github.com/hunterhug/golog"

	"github.com/hunterhug/fafacms/core/util/secure"
)

// 其余内容类表的加解密：专栏(node)、关注关系(relation)、用户组(group)、举报(user_bad)。
//
// 这些表里的用户名称字段都是**冗余副本**，一并加密；
// 相关过滤已改为按对应的 *_id 列（见 controllers/name_resolve.go）。

const (
	nodeFieldName     = "name"
	nodeFieldDescribe = "describe"
	nodeFieldUserName = "user_name"

	relationFieldUserAName = "user_a_name"
	relationFieldUserBName = "user_b_name"

	groupFieldName     = "name"
	groupFieldDescribe = "describe"

	userBadFieldReason = "reason"
)

// groupFieldOwner 固定为 0：Group.Id 为自增主键，插入前尚为 0，
// 用真实 Id 会导致"插入与读取 AAD 不一致"而无法解密（同 user 表的原因）。
const groupFieldOwner int64 = 0

// ---------------------------------------------------------------------------
// ContentNode（专栏）
// ---------------------------------------------------------------------------

func (n *ContentNode) EncryptFields() error {
	var err error
	if n.Name, err = EncryptOptional("content_node", nodeFieldName, n.UserId, n.Name); err != nil {
		return err
	}
	if n.Describe, err = EncryptOptional("content_node", nodeFieldDescribe, n.UserId, n.Describe); err != nil {
		return err
	}
	if n.UserName, err = EncryptOptional("content_node", nodeFieldUserName, n.UserId, n.UserName); err != nil {
		return err
	}
	return nil
}

func (n *ContentNode) DecryptFields() error {
	var err error
	if n.Name, err = DecryptOptional("content_node", nodeFieldName, n.UserId, n.Name); err != nil {
		return err
	}
	if n.Describe, err = DecryptOptional("content_node", nodeFieldDescribe, n.UserId, n.Describe); err != nil {
		return err
	}
	if n.UserName, err = DecryptOptional("content_node", nodeFieldUserName, n.UserId, n.UserName); err != nil {
		return err
	}
	return nil
}

func (n *ContentNode) BeforeInsert() {
	if err := n.EncryptFields(); err != nil {
		log.Errorf("ContentNode.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (n *ContentNode) BeforeUpdate() {
	if err := n.EncryptFields(); err != nil {
		log.Errorf("ContentNode.BeforeUpdate encrypt err: %s", err.Error())
	}
}

func (n *ContentNode) AfterLoad() {
	if err := n.DecryptFields(); err != nil {
		log.Errorf("ContentNode.AfterLoad decrypt err: %s", err.Error())
	}
}

func (n *ContentNode) AfterInsert() {
	if err := n.DecryptFields(); err != nil {
		log.Errorf("ContentNode.AfterInsert decrypt err: %s", err.Error())
	}
}

func (n *ContentNode) AfterUpdate() {
	if err := n.DecryptFields(); err != nil {
		log.Errorf("ContentNode.AfterUpdate decrypt err: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Relation（关注关系）：两个名称副本，归属取 user_a_id（插入前已设置且稳定）
// ---------------------------------------------------------------------------

func (r *Relation) EncryptFields() error {
	var err error
	if r.UserAName, err = EncryptOptional("relation", relationFieldUserAName, r.UserAId, r.UserAName); err != nil {
		return err
	}
	if r.UserBName, err = EncryptOptional("relation", relationFieldUserBName, r.UserAId, r.UserBName); err != nil {
		return err
	}
	return nil
}

func (r *Relation) DecryptFields() error {
	var err error
	if r.UserAName, err = DecryptOptional("relation", relationFieldUserAName, r.UserAId, r.UserAName); err != nil {
		return err
	}
	if r.UserBName, err = DecryptOptional("relation", relationFieldUserBName, r.UserAId, r.UserBName); err != nil {
		return err
	}
	return nil
}

func (r *Relation) BeforeInsert() {
	if err := r.EncryptFields(); err != nil {
		log.Errorf("Relation.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (r *Relation) AfterLoad() {
	if err := r.DecryptFields(); err != nil {
		log.Errorf("Relation.AfterLoad decrypt err: %s", err.Error())
	}
}

func (r *Relation) AfterInsert() {
	if err := r.DecryptFields(); err != nil {
		log.Errorf("Relation.AfterInsert decrypt err: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Group（用户组）：Name 有唯一约束与等值过滤 → 加密 + 盲索引
// ---------------------------------------------------------------------------

func (g *Group) EncryptFields() error {
	// 名称：加密 + 同步盲索引（唯一性由 name_bidx 承担）
	if g.Name != "" && !secure.IsCipher(g.Name) {
		bidx, err := BlindIndexOf(g.Name)
		if err != nil {
			return err
		}
		enc, err := EncryptTableField("group", groupFieldName, groupFieldOwner, g.Name)
		if err != nil {
			return err
		}
		g.Name = enc
		g.NameBidx = bidx
	}

	var err error
	if g.Describe, err = EncryptOptional("group", groupFieldDescribe, groupFieldOwner, g.Describe); err != nil {
		return err
	}
	return nil
}

func (g *Group) DecryptFields() error {
	var err error
	if g.Name, err = DecryptOptional("group", groupFieldName, groupFieldOwner, g.Name); err != nil {
		return err
	}
	if g.Describe, err = DecryptOptional("group", groupFieldDescribe, groupFieldOwner, g.Describe); err != nil {
		return err
	}
	return nil
}

func (g *Group) BeforeInsert() {
	if err := g.EncryptFields(); err != nil {
		log.Errorf("Group.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (g *Group) BeforeUpdate() {
	if err := g.EncryptFields(); err != nil {
		log.Errorf("Group.BeforeUpdate encrypt err: %s", err.Error())
	}
}

func (g *Group) AfterLoad() {
	if err := g.DecryptFields(); err != nil {
		log.Errorf("Group.AfterLoad decrypt err: %s", err.Error())
	}
}

func (g *Group) AfterInsert() {
	if err := g.DecryptFields(); err != nil {
		log.Errorf("Group.AfterInsert decrypt err: %s", err.Error())
	}
}

func (g *Group) AfterUpdate() {
	if err := g.DecryptFields(); err != nil {
		log.Errorf("Group.AfterUpdate decrypt err: %s", err.Error())
	}
}

// PrepareSearch 把用于查询的组名明文转换为盲索引条件（并清空明文列）。
func (g *Group) PrepareSearch() error {
	if g.Name == "" {
		return nil
	}
	if g.NameBidx == "" {
		bidx, err := BlindIndexOf(g.Name)
		if err != nil {
			return err
		}
		g.NameBidx = bidx
	}
	g.Name = ""
	return nil
}

// ---------------------------------------------------------------------------
// UserBad（举报用户）：理由
// ---------------------------------------------------------------------------

func (b *UserBad) EncryptFields() error {
	var err error
	if b.Reason, err = EncryptOptional("user_bad", userBadFieldReason, b.UserId, b.Reason); err != nil {
		return err
	}
	return nil
}

func (b *UserBad) DecryptFields() error {
	var err error
	if b.Reason, err = DecryptOptional("user_bad", userBadFieldReason, b.UserId, b.Reason); err != nil {
		return err
	}
	return nil
}

func (b *UserBad) BeforeInsert() {
	if err := b.EncryptFields(); err != nil {
		log.Errorf("UserBad.BeforeInsert encrypt err: %s", err.Error())
	}
}

func (b *UserBad) AfterLoad() {
	if err := b.DecryptFields(); err != nil {
		log.Errorf("UserBad.AfterLoad decrypt err: %s", err.Error())
	}
}

func (b *UserBad) AfterInsert() {
	if err := b.DecryptFields(); err != nil {
		log.Errorf("UserBad.AfterInsert decrypt err: %s", err.Error())
	}
}
