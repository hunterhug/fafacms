package model

import (
	"errors"
	"fmt"
	"github.com/hunterhug/fafacms/core/util"
	"github.com/hunterhug/fafacms/core/util/secure"
	"time"
)

type User struct {
	Id       int64  `json:"id" xorm:"bigint pk autoincr"`
	Name     string `json:"name" xorm:"varchar(255) notnull"`      // 加密存储
	NickName string `json:"nick_name" xorm:"varchar(255) notnull"` // 加密存储
	Email    string `json:"email" xorm:"varchar(255) notnull"`     // 加密存储
	WeChat   string `json:"wechat" xorm:"varchar(255)"`            // 加密存储
	WeiBo    string `json:"weibo" xorm:"TEXT"`                     // 加密存储
	Github   string `json:"github" xorm:"TEXT"`                    // 加密存储
	QQ       string `json:"qq" xorm:"varchar(255)"`                // 加密存储

	// 盲索引列：等值查询与唯一约束用（HMAC 摘要，非明文、不可逆）
	NameBidx     string `json:"-" xorm:"char(64) unique"`
	NickNameBidx string `json:"-" xorm:"char(64) unique"`
	EmailBidx    string `json:"-" xorm:"char(64) unique"`
	WeChatBidx   string `json:"-" xorm:"char(64) index"`
	WeiBoBidx    string `json:"-" xorm:"char(64) index"`
	GithubBidx   string `json:"-" xorm:"char(64) index"`
	QQBidx       string `json:"-" xorm:"char(64) index"`

	NickNameUpdateTime  int64  `json:"nick_name_update_time"`
	Password            string `json:"password,omitempty" xorm:"varchar(100)"`
	Gender              int    `json:"gender" xorm:"notnull default(0) comment('0 unknow,1 boy,2 girl') TINYINT(1)"`
	ShortDescribe       string `json:"short_describe" xorm:"TEXT"`
	Describe            string `json:"describe" xorm:"MEDIUMTEXT"`
	HeadPhoto           string `json:"head_photo" xorm:"varchar(700)"`
	CreateTime          int64  `json:"create_time"`
	UpdateTime          int64  `json:"update_time,omitempty"`
	ActivateTime        int64  `json:"activate_time,omitempty"`              // activate time
	ActivateCode        string `json:"activate_code,omitempty" xorm:"index"` // activate code（存摘要）
	ActivateCodeExpired int64  `json:"activate_code_expired,omitempty"`      // activate code expired time
	Status              int    `json:"status" xorm:"notnull default(0) comment('0 un active, 1 normal, 2 black') TINYINT(1) index"`
	GroupId             int64  `json:"group_id,omitempty" xorm:"bigint index"`
	ResetCode           string `json:"reset_code,omitempty" xorm:"index"` // forget password code（存摘要）
	ResetCodeExpired    int64  `json:"reset_code_expired,omitempty"`      // forget password code expired
	LoginTime           int64  `json:"login_time,omitempty"`              // login time last time
	LoginIp             string `json:"login_ip,omitempty"`                // login ip last time（加密存储）
	Vip                 int    `json:"vip"`                               // only vip can op node and content
	TwoFaSecret         string `json:"-" xorm:"varchar(255)"`             // 2FA TOTP secret（加密存储，密文形态，绝不外露）
	TwoFa               bool   `json:"two_fa,omitempty" xorm:"-"`         // computed: whether 2FA enabled (not a db column)
	FollowedNum         int64  `json:"followed_num" xorm:"notnull default(0)"`
	FollowingNum        int64  `json:"following_num" xorm:"notnull default(0)"`
	ContentNum          int64  `json:"content_num" xorm:"notnull default(0)"`      // normal publish content num
	ContentCoolNum      int64  `json:"content_cool_num" xorm:"notnull default(0)"` // normal content cool num
}

var UserSortName = []string{"=id", "=name", "-vip", "-activate_time", "=followed_num", "=following_num", "=content_num", "=content_cool_num", "=create_time", "=update_time", "=gender"}

// Get 按结构体已设置的条件查询用户（命中后解密回填）。
// 注意：name / email 等加密列必须先转成盲索引条件，见 PrepareSearch。
func (u *User) Get() (err error) {
	var exist bool
	if err = u.PrepareSearch(); err != nil {
		return
	}
	exist, err = FaFaRdb.Client.Get(u)
	if err != nil {
		return
	}
	if !exist {
		return fmt.Errorf("user not found")
	}
	return
}

// GetRaw 同 Get，但返回是否存在而不报错。
func (u *User) GetRaw() (bool, error) {
	if err := u.PrepareSearch(); err != nil {
		return false, err
	}
	return FaFaRdb.Client.Get(u)
}

// GetActivateRaw 查询已激活用户（status != 0）。
func (u *User) GetActivateRaw() (bool, error) {
	if err := u.PrepareSearch(); err != nil {
		return false, err
	}
	return FaFaRdb.Client.Where("status!=?", 0).Get(u)
}

// Exist 判断用户是否存在。name 为加密列，因此按盲索引列查询。
func (u *User) Exist() (bool, error) {
	if u.Id == 0 && u.Name == "" && u.NameBidx == "" && u.GroupId == 0 {
		return false, errors.New("where is empty")
	}

	s := FaFaRdb.Client.Table(u)
	s.Where("1=1")

	if u.Id != 0 {
		s.And("id=?", u.Id)
	}

	if u.Name != "" {
		bidx, err := BlindIndexOf(u.Name)
		if err != nil {
			return false, err
		}
		s.And("name_bidx=?", bidx)
	} else if u.NameBidx != "" {
		s.And("name_bidx=?", u.NameBidx)
	}

	if u.GroupId != 0 {
		s.And("group_id=?", u.GroupId)
	}

	c, err := s.Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (u *User) IsNameRepeat() (bool, error) {
	if u.Name == "" {
		return false, errors.New("where is empty")
	}
	bidx, err := BlindIndexOf(u.Name)
	if err != nil {
		return false, err
	}
	c, err := FaFaRdb.Client.Table(u).Where("name_bidx=?", bidx).Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (u *User) IsNickNameRepeat() (bool, error) {
	if u.NickName == "" {
		return false, errors.New("where is empty")
	}
	bidx, err := BlindIndexOf(u.NickName)
	if err != nil {
		return false, err
	}
	c, err := FaFaRdb.Client.Table(u).Where("nick_name_bidx=?", bidx).Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (u *User) IsEmailRepeat() (bool, error) {
	if u.Email == "" {
		return false, errors.New("where is empty")
	}
	bidx, err := BlindIndexOf(u.Email)
	if err != nil {
		return false, err
	}
	c, err := FaFaRdb.Client.Table(u).Where("email_bidx=?", bidx).Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (u *User) InsertOne() error {
	u.CreateTime = time.Now().Unix()
	if u.Password != "" {
		hash, err := HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hash
	}
	_, err := FaFaRdb.Insert(u)
	return err
}

func (u *User) IsActivateCodeExist() (bool, error) {
	if u.ActivateCode == "" || u.Email == "" {
		return false, errors.New("where is empty")
	}
	if err := u.PrepareSearch(); err != nil {
		return false, err
	}
	c, err := FaFaRdb.Client.Get(u)
	return c, err
}

func (u *User) UpdateActivateStatus() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}
	u.ActivateTime = time.Now().Unix()
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("status", "activate_time").Update(u)
	return err
}

// UpdateActivateCode 重新生成激活码：库中只存摘要，返回明文供发信使用。
func (u *User) UpdateActivateCode() (string, error) {
	if u.Id == 0 {
		return "", errors.New("where is empty")
	}
	code := util.GenCode6()
	digest, err := DigestSecretCode(code)
	if err != nil {
		return "", err
	}
	u.UpdateTime = time.Now().Unix()
	u.ActivateCode = digest
	u.ActivateCodeExpired = time.Now().Add(5 * time.Minute).Unix()
	_, err = FaFaRdb.Client.Where("id=?", u.Id).Cols("activate_code", "activate_code_expired", "update_time").Update(u)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (u *User) GetUserByEmail() (bool, error) {
	if u.Email == "" {
		return false, errors.New("where is empty")
	}
	if err := u.PrepareSearch(); err != nil {
		return false, err
	}
	c, err := FaFaRdb.Client.Get(u)
	return c, err
}

// UpdateCode 重新生成忘记密码验证码：库中只存摘要，返回明文供发信使用。
func (u *User) UpdateCode() (string, error) {
	if u.Id == 0 {
		return "", errors.New("where is empty")
	}
	code := util.GenCode6()
	digest, err := DigestSecretCode(code)
	if err != nil {
		return "", err
	}
	u.UpdateTime = time.Now().Unix()
	u.ResetCode = digest
	u.ResetCodeExpired = time.Now().Unix() + 300
	_, err = FaFaRdb.Client.Where("id=?", u.Id).Cols("reset_code", "reset_code_expired", "update_time").Update(u)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (u *User) UpdatePassword() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}
	u.UpdateTime = time.Now().Unix()
	u.ResetCode = ""
	u.ResetCodeExpired = 0
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("reset_code", "reset_code_expired", "update_time", "password").Update(u)
	return err
}

// UpdateTwoFa 更新 2FA 秘钥（绑定/关闭/管理员重置）。
// 传入明文 secret 时自动加密入库；空字符串表示关闭（不加密）。
func (u *User) UpdateTwoFa() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}
	if u.TwoFaSecret != "" && !secure.IsCipher(u.TwoFaSecret) {
		enc, err := EncryptUserField("two_fa_secret", u.Id, u.TwoFaSecret)
		if err != nil {
			return err
		}
		u.TwoFaSecret = enc
	}
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("two_fa_secret").Update(u)
	return err
}

// TwoFaSecretPlain 返回解密后的 2FA 秘钥明文，用于 TOTP 校验。
// 未开启 2FA（空值）返回空字符串。
func (u *User) TwoFaSecretPlain() (string, error) {
	if u.TwoFaSecret == "" {
		return "", nil
	}
	return DecryptUserField("two_fa_secret", u.Id, u.TwoFaSecret)
}

// ClearResetCode 作废忘记密码验证码（爆破防护触发时调用）。
func (u *User) ClearResetCode() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}
	u.ResetCode = ""
	u.ResetCodeExpired = 0
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("reset_code", "reset_code_expired").Update(u)
	return err
}

// ClearActivateCode 作废激活码（激活错满次数时调用），并把过期时间置为过去，
// 使 ResendActivateCodeToUser 的"未过期不可重发"检查放行、可重新发码。
func (u *User) ClearActivateCode() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}
	u.ActivateCode = ""
	u.ActivateCodeExpired = time.Now().Unix() - 1
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("activate_code", "activate_code_expired").Update(u)
	return err
}

func (u *User) UpdateInfo() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}

	u.UpdateTime = time.Now().Unix()
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Omit("id").Update(u)
	return err
}

func (u *User) UpdateInfoMustVip() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}

	u.UpdateTime = time.Now().Unix()
	_, err := FaFaRdb.Client.Where("id=?", u.Id).Omit("id").MustCols("vip").Update(u)
	return err
}

func (u *User) UpdateLoginInfo() error {
	if u.Id == 0 {
		return errors.New("where is empty")
	}

	_, err := FaFaRdb.Client.Where("id=?", u.Id).Cols("login_time", "login_ip").Update(u)
	return err
}

func UserAllExist(userIds []int64) bool {
	num, _ := FaFaRdb.Client.Where("status!=?", 0).In("id", userIds).Count(new(User))
	return len(userIds) == int(num)
}

func UserCount() (int64, error) {
	return FaFaRdb.Client.Where("status!=?", 0).Count(new(User))
}
