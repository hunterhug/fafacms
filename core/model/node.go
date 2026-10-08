package model

import (
	"errors"
	"time"
)

type ContentNode struct {
	Id           int64  `json:"id" xorm:"bigint pk autoincr"`
	UserId       int64  `json:"user_id" xorm:"bigint index"`
	UserName     string `json:"user_name" xorm:"varchar(255)"` // 加密存储
	Seo          string `json:"seo" xorm:"index"`
	Status       int    `json:"status" xorm:"notnull default(0) comment('0 normal,1 hide') TINYINT(1) index"`
	Name         string `json:"name" xorm:"varchar(255) notnull"` // 加密存储
	Describe     string `json:"describe" xorm:"MEDIUMTEXT"`       // 加密存储
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time,omitempty"`
	ImagePath    string `json:"image_path" xorm:"varchar(700)"`
	ParentNodeId int64  `json:"parent_node_id" xorm:"bigint index"`
	Level        int    `json:"level"`
	SortNum      int64  `json:"sort_num" xorm:"notnull default(0)"`
	ContentNum   int64  `json:"content_num" xorm:"notnull default(0)"` // normal publish content num
}

var ContentNodeSortName = []string{"=id", "+sort_num", "-create_time", "-update_time", "+status", "=seo", "=content_num"}

func (n *ContentNode) CountNodeNum() (int64, error) {
	if n.UserId == 0 {
		return 0, errors.New("where is empty")
	}

	num, err := FaFaRdb.Client.Table(n).Where("user_id=?", n.UserId).Where("parent_node_id=?", n.ParentNodeId).Count()
	return num, err
}

func (n *ContentNode) CheckSeoValid() (bool, error) {
	if n.UserId == 0 || n.Seo == "" {
		return false, errors.New("where is empty")
	}

	c, err := FaFaRdb.Client.Table(n).Where("user_id=?", n.UserId).And("seo=?", n.Seo).Count()

	if c >= 1 {
		return true, nil
	}
	return false, err
}

func (n *ContentNode) CheckParentValid() (bool, error) {
	if n.UserId == 0 || n.ParentNodeId == 0 {
		return false, errors.New("where is empty")
	}

	c, err := FaFaRdb.Client.Table(n).Where("user_id=?", n.UserId).And("id=?", n.ParentNodeId).And("level=?", 0).Count()

	if c >= 1 {
		return true, nil
	}
	return false, err
}

func (n *ContentNode) CheckChildrenNum() (int, error) {
	if n.UserId == 0 || n.Id == 0 {
		return 0, errors.New("where is empty")
	}
	num, err := FaFaRdb.Client.Table(n).Where("user_id=?", n.UserId).And("parent_node_id=?", n.Id).Count()
	return int(num), err
}

func (n *ContentNode) InsertOne() error {
	n.CreateTime = time.Now().Unix()
	_, err := FaFaRdb.Insert(n)
	return err
}

func (n *ContentNode) Get() (bool, error) {
	if n.Id == 0 && n.Seo == "" {
		return false, errors.New("where is empty")
	}
	return FaFaRdb.Client.Get(n)
}

func (n *ContentNode) GetSortOneNode() (bool, error) {
	if n.UserId == 0 {
		return false, errors.New("where is empty")
	}
	return FaFaRdb.Client.Get(n)
}

func (n *ContentNode) Exist() (bool, error) {
	if n.Id == 0 {
		return false, errors.New("where is empty")
	}
	num, err := FaFaRdb.Client.Table(n).Where("id=?", n.Id).And("user_id=?", n.UserId).Count()
	if err != nil {
		return false, err
	}

	return num >= 1, nil
}

func (n *ContentNode) UpdateSeo() error {
	if n.UserId == 0 || n.Id == 0 {
		return errors.New("where is empty")
	}

	session := FaFaRdb.Client.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}

	_, err = session.Exec("update fafacms_content SET node_seo=? where user_id=? and node_id=?", n.Seo, n.UserId, n.Id)
	if err != nil {
		session.Rollback()
		return err
	}

	_, err = session.Where("id=?", n.Id).And("user_id=?", n.UserId).Cols("seo").Update(n)
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}
	return nil
}

func (n *ContentNode) UpdateInfo() error {
	if n.UserId == 0 || n.Id == 0 {
		return errors.New("where is empty")
	}

	session := FaFaRdb.Client.NewSession()
	defer session.Close()
	n.UpdateTime = time.Now().Unix()
	_, err := session.Where("id=?", n.Id).And("user_id=?", n.UserId).MustCols("describe").Omit("id", "user_id").Update(n)
	return err
}

func (n *ContentNode) UpdateImage() error {
	if n.UserId == 0 || n.Id == 0 {
		return errors.New("where is empty")
	}

	session := FaFaRdb.Client.NewSession()
	defer session.Close()
	n.UpdateTime = time.Now().Unix()
	_, err := session.Where("id=?", n.Id).And("user_id=?", n.UserId).MustCols("image_path").Omit("id", "user_id").Update(n)
	return err
}

func (n *ContentNode) UpdateStatus() error {
	if n.UserId == 0 || n.Id == 0 {
		return errors.New("where is empty")
	}

	session := FaFaRdb.Client.NewSession()
	defer session.Close()
	_, err := session.Where("id=?", n.Id).And("user_id=?", n.UserId).Cols("status").Update(n)
	return err
}

func (n *ContentNode) UpdateParent(beforeParentNode int64) error {
	if n.UserId == 0 || n.Id == 0 {
		return errors.New("where is empty")
	}

	session := FaFaRdb.Client.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}

	_, err = session.Exec("update fafacms_content_node SET sort_num=sort_num-1 where sort_num > ? and user_id = ? and parent_node_id = ?", n.SortNum, n.UserId, beforeParentNode)
	if err != nil {
		session.Rollback()
		return err
	}

	c, err := session.Table(n).Where("user_id=?", n.UserId).And("parent_node_id", n.ParentNodeId).Count()
	if err != nil {
		session.Rollback()
		return err
	}

	n.SortNum = c

	_, err = session.Exec("update fafacms_content_node SET sort_num=?, level=?, parent_node_id=? where id = ? and user_id = ?", n.SortNum, n.Level, n.ParentNodeId, n.Id, n.UserId)
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}

	return nil
}
