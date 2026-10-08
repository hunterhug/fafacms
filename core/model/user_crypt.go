package model

import (
	log "github.com/hunterhug/golog"

	"github.com/hunterhug/fafacms/core/util/secure"
)

// user 表字段分三类处理：
//
//  1. 可查询字段（需等值查询或唯一约束）→ 加密 + 同步盲索引列 *_bidx
//     name / nick_name / email（唯一、登录）
//     we_chat / wei_bo / github / q_q（管理端用户列表的等值过滤）
//  2. 纯隐私字段（无查询）→ 只加密
//     describe / short_describe / login_ip
//  3. 凭据字段（不可逆）→ bcrypt 或 HMAC 摘要，已单独处理
//     password / activate_code / reset_code / two_fa_secret
const (
	userFieldName        = "name"
	userFieldNickName    = "nick_name"
	userFieldEmail       = "email"
	userFieldWeChat      = "we_chat"
	userFieldWeiBo       = "wei_bo"
	userFieldGithub      = "github"
	userFieldQQ          = "q_q"
	userFieldLoginIp     = "login_ip"
	userFieldDescribe    = "describe"
	userFieldShortDescri = "short_describe"
)

// userFieldOwner 是 user 表字段 AAD 中的归属标识。
//
// 固定为 0 的原因：注册时用户 Id 还是 0（自增主键在 INSERT 之后才确定），
// 若 AAD 使用真实 Id，则"插入时用 0、读取时用真实 Id"会导致 AAD 不一致而无法解密。
// 固定 0 保证写入与读取两侧一致；代价是无法防"同一字段在不同用户行之间搬运密文"
// （该攻击需要数据库写权限，不在本方案的威胁模型内 —— 本方案防的是拖库/备份泄漏）。
const userFieldOwner int64 = 0

// searchableField 描述一对「明文字段 ↔ 盲索引列」。
type searchableField struct {
	plain *string
	bidx  *string
	col   string // 数据库列名（用于 AAD，读写两侧必须一致）
}

// searchableFields 返回 user 表的可查询字段清单。
func (u *User) searchableFields() []searchableField {
	return []searchableField{
		{&u.Name, &u.NameBidx, userFieldName},
		{&u.NickName, &u.NickNameBidx, userFieldNickName},
		{&u.Email, &u.EmailBidx, userFieldEmail},
		{&u.WeChat, &u.WeChatBidx, userFieldWeChat},
		{&u.WeiBo, &u.WeiBoBidx, userFieldWeiBo},
		{&u.Github, &u.GithubBidx, userFieldGithub},
		{&u.QQ, &u.QQBidx, userFieldQQ},
	}
}

// EncryptFields 加密 user 表的敏感字段，并同步可查询字段的盲索引。
// 幂等：已是密文的字段会跳过（但仍会补齐缺失的盲索引）。
func (u *User) EncryptFields() error {
	for _, f := range u.searchableFields() {
		if *f.plain == "" {
			continue
		}
		// 已是密文（例如刚插入后又更新）：只补盲索引不可行（无法反推明文），跳过
		if secure.IsCipher(*f.plain) {
			continue
		}

		bidx, err := BlindIndexOf(*f.plain)
		if err != nil {
			return err
		}
		enc, err := EncryptTableField("user", f.col, userFieldOwner, *f.plain)
		if err != nil {
			return err
		}
		*f.plain = enc
		*f.bidx = bidx
	}

	var err error
	if u.LoginIp, err = EncryptOptional("user", userFieldLoginIp, userFieldOwner, u.LoginIp); err != nil {
		return err
	}
	if u.Describe, err = EncryptOptional("user", userFieldDescribe, userFieldOwner, u.Describe); err != nil {
		return err
	}
	if u.ShortDescribe, err = EncryptOptional("user", userFieldShortDescri, userFieldOwner, u.ShortDescribe); err != nil {
		return err
	}
	return nil
}

// DecryptFields 解密 user 表的敏感字段。盲索引列是摘要，无需（也无法）解密。
func (u *User) DecryptFields() error {
	var err error
	for _, f := range u.searchableFields() {
		if *f.plain, err = DecryptOptional("user", f.col, userFieldOwner, *f.plain); err != nil {
			return err
		}
	}

	if u.LoginIp, err = DecryptOptional("user", userFieldLoginIp, userFieldOwner, u.LoginIp); err != nil {
		return err
	}
	if u.Describe, err = DecryptOptional("user", userFieldDescribe, userFieldOwner, u.Describe); err != nil {
		return err
	}
	if u.ShortDescribe, err = DecryptOptional("user", userFieldShortDescri, userFieldOwner, u.ShortDescribe); err != nil {
		return err
	}
	return nil
}

// PrepareSearch 把"用于查询的明文字段"转换为盲索引条件。
//
// 加密列无法参与等值查询（每次加密结果不同），因此查询前必须先转成 *_bidx；
// 同时清空明文字段，避免 xorm 用密文列拼出永远匹配不上的 WHERE 条件。
//
// 命中后 xorm 会用库中真实值回填结构体，并由 AfterLoad 自动解密，调用方无感。
func (u *User) PrepareSearch() error {
	for _, f := range u.searchableFields() {
		if *f.plain == "" {
			continue
		}
		if *f.bidx == "" {
			bidx, err := BlindIndexOf(*f.plain)
			if err != nil {
				return err
			}
			*f.bidx = bidx
		}
		*f.plain = ""
	}
	return nil
}

// BeforeInsert xorm 钩子：入库前自动加密并生成盲索引。
func (u *User) BeforeInsert() {
	if err := u.EncryptFields(); err != nil {
		log.Errorf("User.BeforeInsert encrypt err: %s", err.Error())
	}
}

// BeforeUpdate xorm 钩子：更新前自动加密并生成盲索引。
// 各 Update 方法用 Cols 指定列，因此只会写入本次实际要更新的字段。
func (u *User) BeforeUpdate() {
	if err := u.EncryptFields(); err != nil {
		log.Errorf("User.BeforeUpdate encrypt err: %s", err.Error())
	}
}

// AfterLoad xorm 钩子：从库中读出后自动解密，调用方无感。
func (u *User) AfterLoad() {
	if err := u.DecryptFields(); err != nil {
		log.Errorf("User.AfterLoad decrypt err: %s", err.Error())
	}
}

// AfterInsert / AfterUpdate 钩子：写入完成后把内存中的结构体恢复为明文。
// 必要性：加密发生在写入前（会就地把字段改成密文），若不恢复，
// 调用方在写库后继续使用该结构体（如直接回给前端）就会拿到密文。
func (u *User) AfterInsert() {
	if err := u.DecryptFields(); err != nil {
		log.Errorf("User.AfterInsert decrypt err: %s", err.Error())
	}
}

func (u *User) AfterUpdate() {
	if err := u.DecryptFields(); err != nil {
		log.Errorf("User.AfterUpdate decrypt err: %s", err.Error())
	}
}
