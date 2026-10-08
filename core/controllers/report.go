package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
	"math"
)

// ================= 内容举报列表（管理员） =================

type ListContentBadAdminRequest struct {
	Status int      `json:"status" validate:"oneof=-1 0 1 2 3"` // 被举报文章当前状态：-1 全部
	Sort   []string `json:"sort"`
	PageHelp
}

type ContentBadItem struct {
	Id              int64  `json:"id"`
	UserId          int64  `json:"user_id"` // 举报人
	UserName        string `json:"user_name"`
	UserNick        string `json:"user_nick"`
	ContentId       int64  `json:"content_id"` // 被举报文章
	ContentTitle    string `json:"content_title"`
	ContentUserId   int64  `json:"content_user_id"` // 被举报文章作者
	ContentUserName string `json:"content_user_name"`
	ContentStatus   int    `json:"content_status"` // 0 正常 1 隐藏 2 违禁 3 回收站
	Reason          string `json:"reason"`
	CreateTime      int64  `json:"create_time"`
}

type ListContentBadAdminResponse struct {
	Bad []ContentBadItem `json:"bad"`
	PageHelp
}

func ListContentBadAdmin(c *gin.Context) {
	resp := new(Resp)
	respResult := new(ListContentBadAdminResponse)
	req := new(ListContentBadAdminRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("ListContentBadAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()
	session.Table(new(model.ContentBad)).Where("1=1")

	// 状态过滤：先取满足状态的文章 id 集合
	if req.Status != -1 {
		var cs []model.Content
		if err := model.FaFaRdb.Client.Cols("id").Where("status=?", req.Status).Find(&cs); err != nil {
			log.Errorf("ListContentBadAdmin err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if len(cs) == 0 {
			respResult.Bad = []ContentBadItem{}
			respResult.PageHelp = req.PageHelp
			resp.Data = respResult
			resp.Flag = true
			return
		}
		ids := make([]int64, 0, len(cs))
		for _, v := range cs {
			ids = append(ids, v.Id)
		}
		session.In("content_id", ids)
	}

	cs := make([]model.ContentBad, 0)
	p := &req.PageHelp
	p.build(session, req.Sort, []string{"=id"})

	total, err := session.FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListContentBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	contentIds := make([]int64, 0, len(cs))
	userIds := make([]int64, 0, len(cs))
	for _, v := range cs {
		contentIds = append(contentIds, v.ContentId)
		userIds = append(userIds, v.UserId)
	}

	contents, err := model.GetContentHelper(contentIds, true, 0)
	if err != nil {
		log.Errorf("ListContentBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	users, err := model.GetUser(userIds)
	if err != nil {
		log.Errorf("ListContentBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	items := make([]ContentBadItem, 0, len(cs))
	for _, v := range cs {
		item := ContentBadItem{
			Id:         v.Id,
			UserId:     v.UserId,
			ContentId:  v.ContentId,
			Reason:     v.Reason,
			CreateTime: v.CreateTime,
		}
		if u, ok := users[v.UserId]; ok {
			item.UserName = u.Name
			item.UserNick = u.NickName
		}
		if c, ok := contents[v.ContentId]; ok {
			item.ContentTitle = c.Title
			item.ContentUserId = c.UserId
			item.ContentUserName = c.UserName
			item.ContentStatus = c.Status
		}
		items = append(items, item)
	}

	respResult.Bad = items
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

// ================= 评论举报列表（管理员） =================

type ListCommentBadAdminRequest struct {
	Status int      `json:"status" validate:"oneof=-1 0 1"` // 被举报评论当前状态：-1 全部
	Sort   []string `json:"sort"`
	PageHelp
}

type CommentBadItem struct {
	Id              int64  `json:"id"`
	UserId          int64  `json:"user_id"` // 举报人
	UserName        string `json:"user_name"`
	UserNick        string `json:"user_nick"`
	CommentId       int64  `json:"comment_id"` // 被举报评论
	CommentDescribe string `json:"comment_describe"`
	CommentUserId   int64  `json:"comment_user_id"` // 被举报评论作者
	CommentUserName string `json:"comment_user_name"`
	CommentStatus   int    `json:"comment_status"` // 0 正常 1 违禁
	CommentIsDelete int    `json:"comment_is_delete"`
	ContentId       int64  `json:"content_id"` // 所属文章
	ContentTitle    string `json:"content_title"`
	Reason          string `json:"reason"`
	CreateTime      int64  `json:"create_time"`
}

type ListCommentBadAdminResponse struct {
	Bad []CommentBadItem `json:"bad"`
	PageHelp
}

func ListCommentBadAdmin(c *gin.Context) {
	resp := new(Resp)
	respResult := new(ListCommentBadAdminResponse)
	req := new(ListCommentBadAdminRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("ListCommentBadAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()
	session.Table(new(model.CommentBad)).Where("1=1")

	if req.Status != -1 {
		var cms []model.Comment
		if err := model.FaFaRdb.Client.Cols("id").Where("status=?", req.Status).Find(&cms); err != nil {
			log.Errorf("ListCommentBadAdmin err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if len(cms) == 0 {
			respResult.Bad = []CommentBadItem{}
			respResult.PageHelp = req.PageHelp
			resp.Data = respResult
			resp.Flag = true
			return
		}
		ids := make([]int64, 0, len(cms))
		for _, v := range cms {
			ids = append(ids, v.Id)
		}
		session.In("comment_id", ids)
	}

	cs := make([]model.CommentBad, 0)
	p := &req.PageHelp
	p.build(session, req.Sort, []string{"=id"})

	total, err := session.FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListCommentBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	commentIds := make([]int64, 0, len(cs))
	contentIds := make([]int64, 0, len(cs))
	reporterIds := make([]int64, 0, len(cs))
	for _, v := range cs {
		commentIds = append(commentIds, v.CommentId)
		contentIds = append(contentIds, v.ContentId)
		reporterIds = append(reporterIds, v.UserId)
	}

	comments, users, err := model.GetCommentAndCommentUser(commentIds, true, reporterIds, 0)
	if err != nil {
		log.Errorf("ListCommentBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	contents, err := model.GetContentHelper(contentIds, true, 0)
	if err != nil {
		log.Errorf("ListCommentBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	items := make([]CommentBadItem, 0, len(cs))
	for _, v := range cs {
		item := CommentBadItem{
			Id:         v.Id,
			UserId:     v.UserId,
			CommentId:  v.CommentId,
			ContentId:  v.ContentId,
			Reason:     v.Reason,
			CreateTime: v.CreateTime,
		}
		if u, ok := users[v.UserId]; ok {
			item.UserName = u.Name
			item.UserNick = u.NickName
		}
		if cm, ok := comments[v.CommentId]; ok {
			item.CommentDescribe = cm.Describe
			item.CommentUserId = cm.UserId
			if cm.IsBan {
				item.CommentStatus = 1
			}
			if cm.CommentDelete {
				item.CommentIsDelete = 1
			}
			if cu, ok := users[cm.UserId]; ok {
				item.CommentUserName = cu.Name
			}
		}
		if c, ok := contents[v.ContentId]; ok {
			item.ContentTitle = c.Title
		}
		items = append(items, item)
	}

	respResult.Bad = items
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}
