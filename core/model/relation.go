package model

import (
	"errors"
	"time"
)

// a follow b
type Relation struct {
	Id          int64  `json:"id" xorm:"bigint pk autoincr"`
	CreateTime  int64  `json:"create_time"`
	UserAId     int64  `json:"user_a_id" xorm:"index(gr)"`
	UserBId     int64  `json:"user_b_id" xorm:"index(gr)"`
	UserAName   string `json:"user_a_name" xorm:"varchar(255)"` // 加密存储
	UserBName   string `json:"user_b_name" xorm:"varchar(255)"` // 加密存储
	IsBoth      bool   `json:"is_both" xorm:"-"`
	IsFollowing bool   `json:"is_following" xorm:"-"`
}

var RelationSortName = []string{"=id", "=user_a_id", "=user_b_id", "-create_time"}

func (r *Relation) Add() (err error) {
	num, err := r.Count()
	if err != nil {
		return err
	}

	if num > 0 {
		return
	}

	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return
	}

	r.CreateTime = time.Now().Unix()
	_, err = se.InsertOne(r)
	if err != nil {
		se.Rollback()
		return
	}

	num, err = se.Where("id=?", r.UserAId).Incr("following_num").Update(new(User))
	if err != nil {
		se.Rollback()
		return
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", r.UserBId).Incr("followed_num").Update(new(User))
	if err != nil {
		se.Rollback()
		return
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return
	}

	go FollowYou(r.UserAId, r.UserBId)
	return
}

func (r *Relation) Count() (num int64, err error) {
	if r.UserAId == 0 || r.UserBId == 0 {
		err = errors.New("user id empty")
		return
	}

	num, err = FaFaRdb.Client.Where("user_a_id=?", r.UserAId).And("user_b_id=?", r.UserBId).Count(new(Relation))
	return
}

func (r *Relation) Minute() (err error) {
	num, err := r.Count()
	if err != nil {
		return err
	}

	if num == 0 {
		return
	}

	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return
	}

	num, err = se.Where("user_a_id=?", r.UserAId).And("user_b_id=?", r.UserBId).Delete(new(Relation))
	if err != nil {
		se.Rollback()
		return
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", r.UserAId).And("following_num>=?", 1).Decr("following_num").Update(new(User))
	if err != nil {
		se.Rollback()
		return
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", r.UserBId).And("followed_num>=?", 1).Decr("followed_num").Update(new(User))
	if err != nil {
		se.Rollback()
		return
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return
	}
	return
}
