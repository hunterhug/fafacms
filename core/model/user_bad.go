package model

import (
	"errors"
	"time"
)

// UserBad 举报用户
type UserBad struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	UserId     int64  `json:"user_id" xorm:"bigint index(gr)"`     // 举报人
	BadUserId  int64  `json:"bad_user_id" xorm:"bigint index(gr)"` // 被举报用户
	Reason     string `json:"reason" xorm:"varchar(255)"`          // 加密存储
	CreateTime int64  `json:"create_time"`
}

// Exist 24 小时内是否举报过（同一用户对同一被举报用户 24 小时只能举报一次）
func (c *UserBad) Exist() (ok bool, err error) {
	if c.UserId == 0 || c.BadUserId == 0 {
		return false, errors.New("where is empty")
	}
	num, err := FaFaRdb.Client.Where("user_id=?", c.UserId).And("bad_user_id=?", c.BadUserId).And("create_time>=?", time.Now().Unix()-24*3600).Count(new(UserBad))
	if err != nil {
		return false, err
	}
	if num >= 1 {
		return true, nil
	}
	return
}

func (c *UserBad) Create() (err error) {
	if c.UserId == 0 || c.BadUserId == 0 {
		return errors.New("where is empty")
	}
	c.CreateTime = time.Now().Unix()
	_, err = FaFaRdb.Client.InsertOne(c)
	return
}
