package model

import (
	"errors"
	"time"
)

const (
	CommentTypeOfContent     = 0
	CommentTypeOfRootComment = 1
	CommentTypeOfComment     = 2
	CommentAnonymous         = 1

	AnonymousUser = "匿名" // just ignore this, not use at all
)

type ContentHelper struct {
	Id          int64  `json:"id"`
	Title       string `json:"title"`
	IsHide      bool   `json:"is_hide"`
	IsInRubbish bool   `json:"is_in_rubbish"`
	IsBan       bool   `json:"is_ban"`
	BanTime     int64  `json:"ban_time"`
	UserId      int64  `json:"user_id"`
	UserName    string `json:"user_name"`
	Seo         string `json:"seo"`
	NodeSeo     string `json:"node_seo"`
	Status      int    `json:"status"`
	CommentNum  int64  `json:"comment_num"`
	IsYourself  bool   `json:"is_yourself"`
}

// get contents from id, if all false, which is deleted or hide will not include in map
func GetContentHelper(ids []int64, all bool, yourUserId int64) (back map[int64]ContentHelper, err error) {
	back = make(map[int64]ContentHelper)
	if len(ids) == 0 {
		return
	}
	cs := make([]Content, 0)
	err = FaFaRdb.Client.Cols("id", "ban_time", "title", "status", "seo", "node_seo", "user_name", "user_id", "comment_num").In("id", ids).Find(&cs)
	if err != nil {
		return
	}

	for _, v := range cs {
		temp := ContentHelper{
			Id:         v.Id,
			Title:      v.Title,
			Seo:        v.Seo,
			UserName:   v.UserName,
			UserId:     v.UserId,
			Status:     v.Status,
			BanTime:    v.BanTime,
			CommentNum: v.CommentNum,
			NodeSeo:    v.NodeSeo,
		}

		if v.Status == 1 {
			temp.IsHide = true
		}
		if v.Status == 2 {
			temp.IsBan = true
		}
		if v.Status == 3 {
			temp.IsInRubbish = true
		}

		if v.Status == 1 || v.Status == 3 {
			// if is yourself will still back the data
			if !all && yourUserId != temp.UserId {
				continue
			}

			if yourUserId == temp.UserId {
				temp.IsYourself = true
			}
		}
		back[v.Id] = temp
	}

	return
}

type UserHelper struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	NickName      string `json:"nick_name"`
	HeadPhoto     string `json:"head_photo"`
	IsVip         bool   `json:"is_vip"`
	ShortDescribe string `json:"short_describe"`
	Gender        int    `json:"gender"`
	IsBlack       bool   `json:"is_black"`
}

type CommentHelper struct {
	Id            int64  `json:"id"`
	Describe      string `json:"describe"`
	ContentId     int64  `json:"content_id"`
	CreateTime    int64  `json:"create_time"`
	CommentDelete bool   `json:"is_delete"`
	IsBan         bool   `json:"is_ban"`
	BanTime       int64  `json:"ban_time"`
	DeleteTime    int64  `json:"delete_time"`
	IsAnonymous   bool   `json:"is_anonymous"`
	UserId        int64  `json:"user_id"`
	Cool          int64  `json:"cool"`
	Bad           int64  `json:"bad"`
	IsYourself    bool   `json:"is_yourself"`
	IsCool        bool   `json:"is_cool"`
}

// get comments from id, if all false, some will not show, and user info related will also show
func GetCommentAndCommentUser(ids []int64, all bool, extraUserId []int64, yourUserId int64) (comments map[int64]CommentHelper, users map[int64]UserHelper, err error) {
	comments = make(map[int64]CommentHelper)
	users = make(map[int64]UserHelper)
	if len(ids) == 0 && len(extraUserId) == 0 {
		return
	}
	cms := make([]Comment, 0)
	err = FaFaRdb.Client.Cols("id", "content_id", "user_id", "create_time", "describe", "is_delete", "delete_time", "bad", "cool", "status", "comment_anonymous", "ban_time").In("id", ids).Find(&cms)
	if err != nil {
		return
	}

	for _, v := range cms {
		temp := CommentHelper{
			Id:            v.Id,
			CreateTime:    v.CreateTime,
			Describe:      v.Describe,
			CommentDelete: v.IsDelete == 1,
			IsBan:         v.Status == 1,
			IsAnonymous:   v.CommentAnonymous == CommentAnonymous,
			UserId:        v.UserId,
			Cool:          v.Cool,
			Bad:           v.Bad,
			BanTime:       v.BanTime,
			DeleteTime:    v.DeleteTime,
			ContentId:     v.ContentId,
		}
		if !all {
			// delete will not show others, but should keep some important field
			if temp.CommentDelete {
				temp = CommentHelper{
					Id:            v.Id,
					CreateTime:    v.CreateTime,
					CommentDelete: v.IsDelete == 1,
					IsAnonymous:   v.CommentAnonymous == CommentAnonymous,
					UserId:        v.UserId,
					DeleteTime:    v.DeleteTime,
					ContentId:     v.ContentId,
				}
				// user info hide
				if yourUserId != temp.UserId && temp.IsAnonymous {
					temp.UserId = 0
				}

				if yourUserId == temp.UserId {
					temp.IsYourself = true
				}
			} else {

				// userId is you, do nothing
				if yourUserId != temp.UserId {
					// user info hide
					if temp.IsAnonymous {
						temp.UserId = 0
					}

					// describe hide
					if temp.IsBan {
						temp.Describe = ""
					}
				} else {
					temp.IsYourself = true
				}
			}
		}

		comments[v.Id] = temp
	}

	// 当前用户点赞过的评论（用于前端点赞态回显 is_cool）
	if yourUserId != 0 && len(comments) > 0 {
		coolIds := make([]int64, 0, len(comments))
		for k := range comments {
			coolIds = append(coolIds, k)
		}
		ccs := make([]CommentCool, 0)
		if e := FaFaRdb.Client.Cols("comment_id").Where("user_id=?", yourUserId).In("comment_id", coolIds).Find(&ccs); e == nil {
			for _, cc := range ccs {
				if c, ok := comments[cc.CommentId]; ok {
					c.IsCool = true
					comments[cc.CommentId] = c
				}
			}
		}
	}

	// user id trip repeat
	userHelper := make(map[int64]struct{})
	for _, v := range comments {
		if v.UserId != 0 {
			userHelper[v.UserId] = struct{}{}
		}
	}

	if extraUserId != nil {
		for _, v := range extraUserId {
			userHelper[v] = struct{}{}
		}
	}

	userIds := make([]int64, 0)
	for k := range userHelper {
		userIds = append(userIds, k)
	}

	users, err = GetUser(userIds)
	return
}

func GetComment(ids []int64, isAdmin bool) (comments map[int64]Comment, err error) {
	comments = make(map[int64]Comment)
	cs := make([]Comment, 0)
	if len(ids) == 0 {
		return
	}
	sess := FaFaRdb.Client.In("id", ids)

	if !isAdmin {
		sess.Where("is_delete=?", 0)
	}

	err = sess.Find(&cs)
	if err == nil {
		for _, v := range cs {
			comments[v.Id] = v
		}
	}
	return
}

func GetUser(userIds []int64) (users map[int64]UserHelper, err error) {
	users = make(map[int64]UserHelper)
	uu := make([]User, 0)
	err = FaFaRdb.Client.Cols("id", "vip", "name", "nick_name", "head_photo", "status", "short_describe", "gender").Where("status!=?", 0).In("id", userIds).Find(&uu)
	if err != nil {
		return
	}

	for _, v := range uu {
		temp := UserHelper{
			Id:            v.Id,
			Name:          v.Name,
			NickName:      v.NickName,
			HeadPhoto:     v.HeadPhoto,
			ShortDescribe: v.ShortDescribe,
			Gender:        v.Gender,
		}

		temp.IsBlack = v.Status == 2
		temp.IsVip = v.Vip == 1
		users[v.Id] = temp
	}

	return
}

type Comment struct {
	Id                  int64     `json:"id" xorm:"bigint pk autoincr"`
	UserId              int64     `json:"-" xorm:"bigint index"`
	UserName            string    `json:"-" xorm:"varchar(255)"` // 加密存储
	ContentId           int64     `json:"content_id" xorm:"bigint index"`
	ContentTitle        string    `json:"content_title" xorm:"TEXT"` // 加密存储
	ContentUserId       int64     `json:"-" xorm:"bigint index"`
	ContentUserName     string    `json:"-" xorm:"varchar(255)"` // 加密存储
	CommentId           int64     `json:"comment_id" xorm:"bigint index"`
	CommentUserId       int64     `json:"-" xorm:"bigint index"`
	CommentUserName     string    `json:"-" xorm:"varchar(255)"` // 加密存储
	RootCommentId       int64     `json:"root_comment_id" xorm:"bigint index"`
	RootCommentUserId   int64     `json:"-" xorm:"bigint index"`
	RootCommentUserName string    `json:"-" xorm:"varchar(255)"` // 加密存储
	Describe            string    `json:"-" xorm:"MEDIUMTEXT"`   // 加密存储
	CreateTime          int64     `json:"-"`
	Status              int       `json:"-" xorm:"notnull default(0) comment('0 normal, 1 ban') TINYINT(1) index"`
	BanTime             int64     `json:"-"`
	Cool                int64     `json:"-" xorm:"notnull default(0)"`
	Bad                 int64     `json:"-" xorm:"notnull default(0)"`
	CommentType         int       `json:"comment_type"` // 0 comment to content, 1 comment to comment, 2 comment to comment more
	CommentAnonymous    int       `json:"-"`
	IsDelete            int       `json:"-"`
	DeleteTime          int64     `json:"-"`
	Son                 []Comment `json:"son,omitempty" xorm:"-"`
	SonNum              int64     `json:"son_num,omitempty" xorm:"-"`
}

var CommentSortName = []string{"=id", "-create_time", "=content_id", "=user_id", "=cool", "=bad"}
var CommentHomeSortName = []string{"-create_time", "-cool"}

type CommentExtra struct {
	Users    map[int64]UserHelper    `json:"users"`
	Comments map[int64]CommentHelper `json:"comments"`
	Contents map[int64]ContentHelper `json:"contents"`
}

type CommentCool struct {
	Id         int64 `json:"id" xorm:"bigint pk autoincr"`
	UserId     int64 `json:"user_id" xorm:"bigint index(gr)"`
	CommentId  int64 `json:"comment_id,omitempty" xorm:"bigint index(gr)"`
	ContentId  int64 `json:"content_id" xorm:"bigint index"`
	CreateTime int64 `json:"create_time"`
}

type CommentBad struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	UserId     int64  `json:"user_id" xorm:"bigint index(gr)"`
	CommentId  int64  `json:"comment_id,omitempty" xorm:"bigint index(gr)"`
	ContentId  int64  `json:"content_id" xorm:"bigint index"`
	Reason     string `json:"reason" xorm:"varchar(255)"` // 加密存储
	CreateTime int64  `json:"create_time"`
}

func (c *Comment) InsertOne() error {
	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err := se.Begin()
	if err != nil {
		return err
	}

	c.CreateTime = time.Now().Unix()
	num, err := se.InsertOne(c)
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", c.ContentId).Incr("comment_num").Update(new(Content))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return err
	}

	return nil
}

func (c *Comment) Get() (bool, error) {
	if c.Id == 0 {
		return false, errors.New("where is empty")
	}
	return FaFaRdb.Client.Get(c)
}

func (c *Comment) UpdateStatus() (int64, error) {
	if c.Id == 0 {
		return 0, errors.New("where is empty")
	}

	if c.Status == 1 {
		c.BanTime = time.Now().Unix()
		num, err := FaFaRdb.Client.Cols("status", "ban_time").Where("id=?", c.Id).And("status!=?", 1).Update(c)
		if err != nil {
			return 0, err
		}

		if num > 0 {
			go BanComment(c.UserId, c.ContentId, c.ContentTitle, c.Id, c.Describe)
		}
		// 违禁后直接返回：不删除举报记录、不重置举报数、不触发解除违禁通知
		return 0, nil
	}

	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err := se.Begin()
	if err != nil {
		return 0, err
	}

	_, err = se.Where("comment_id=?", c.Id).Delete(new(CommentBad))
	if err != nil {
		se.Rollback()
		return 0, err
	}

	c.BanTime = 0
	c.Bad = 0
	num, err := se.Cols("status", "ban_time", "bad").Where("id=?", c.Id).And("status=?", 1).Update(c)
	if err != nil {
		se.Rollback()
		return 0, err
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return 0, err
	}

	if num > 0 {
		go RecoverComment(c.UserId, c.ContentId, c.ContentTitle, c.Id, c.Describe)
	}
	return 0, nil
}

func (c *Comment) Delete() (err error) {
	if c.Id == 0 {
		return errors.New("where is empty")
	}

	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return err
	}

	c.IsDelete = 1
	c.DeleteTime = time.Now().Unix()
	num, err := se.Where("id=?", c.Id).Cols("is_delete", "delete_time").Update(c)
	if err != nil {
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", c.ContentId).Decr("comment_num").Update(new(Content))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return err
	}

	return
}

func (c *CommentCool) Exist() (ok bool, err error) {
	if c.UserId == 0 || c.CommentId == 0 {
		return false, errors.New("where is empty")
	}
	num, err := FaFaRdb.Client.Where("user_id=?", c.UserId).And("comment_id=?", c.CommentId).Count(new(CommentCool))
	if err != nil {
		return false, err
	}
	if num >= 1 {
		return true, nil
	}

	return
}

func (c *CommentCool) Create() (err error) {
	if c.UserId == 0 || c.CommentId == 0 || c.ContentId == 0 {
		return errors.New("where is empty")
	}

	c.CreateTime = time.Now().Unix()

	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return err
	}

	num, err := se.InsertOne(c)
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", c.CommentId).Incr("cool").Update(new(Comment))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return err
	}
	return
}

func (c *CommentCool) Delete() (err error) {
	if c.UserId == 0 || c.CommentId == 0 {
		return errors.New("where is empty")
	}
	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return err
	}

	num, err := se.Where("user_id=?", c.UserId).And("comment_id=?", c.CommentId).Delete(new(CommentCool))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", c.CommentId).And("cool>=?", 1).Decr("cool").Update(new(Comment))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return err
	}
	return
}

func (c *CommentBad) Exist() (ok bool, err error) {
	if c.UserId == 0 || c.CommentId == 0 {
		return false, errors.New("where is empty")
	}
	// 24 小时内是否举报过（同一用户对同一评论 24 小时只能举报一次）
	num, err := FaFaRdb.Client.Where("user_id=?", c.UserId).And("comment_id=?", c.CommentId).And("create_time>=?", time.Now().Unix()-24*3600).Count(new(CommentBad))
	if err != nil {
		return false, err
	}
	if num >= 1 {
		return true, nil
	}

	return
}

func (c *CommentBad) Create() (err error) {
	if c.UserId == 0 || c.CommentId == 0 || c.ContentId == 0 {
		return errors.New("where is empty")
	}

	c.CreateTime = time.Now().Unix()
	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return err
	}

	num, err := se.InsertOne(c)
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", c.CommentId).Incr("bad").Update(new(Comment))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return err
	}
	return
}

func (c *CommentBad) Delete() (err error) {
	if c.UserId == 0 || c.CommentId == 0 {
		return errors.New("where is empty")
	}
	se := FaFaRdb.Client.NewSession()
	defer se.Close()
	err = se.Begin()
	if err != nil {
		return err
	}

	num, err := se.Where("user_id=?", c.UserId).And("comment_id=?", c.CommentId).Delete(new(CommentBad))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	num, err = se.Where("id=?", c.CommentId).And("bad>=?", 1).Decr("bad").Update(new(Comment))
	if err != nil {
		se.Rollback()
		return err
	}

	if num == 0 {
		se.Rollback()
		return errors.New("some err")
	}

	err = se.Commit()
	if err != nil {
		se.Rollback()
		return err
	}
	return
}

func (c *Comment) Ban(num int64) (err error) {
	if c.Id == 0 {
		return errors.New("where is empty")
	}

	c.BanTime = time.Now().Unix()
	c.Status = 1
	num, err = FaFaRdb.Client.Where("id=?", c.Id).And("status!=?", 1).And("bad>?", num).Cols("ban_time", "status").Update(c)

	if num > 0 {
		go BanComment(c.UserId, c.ContentId, c.ContentTitle, c.CommentId, c.Describe)
	}
	return
}

func (c *Comment) UpdateToShowName() (int64, error) {
	if c.Id == 0 {
		return 0, errors.New("where is empty")
	}

	c.CommentAnonymous = 0
	num, err := FaFaRdb.Client.Cols("comment_anonymous").Where("id=?", c.Id).And("comment_anonymous!=?", 0).Update(c)

	return num, err
}
