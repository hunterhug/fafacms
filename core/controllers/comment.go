package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	log "github.com/hunterhug/golog"
	"math"
	"strings"
)

// comment escape
var htmlEscaper = strings.NewReplacer(
	`<`, "&lt;",
	`>`, "&gt;",
)

type RealNameCommentRequest struct {
	CommentId int64 `json:"id"`
}

func RealNameComment(c *gin.Context) {
	resp := new(Resp)
	req := new(RealNameCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CommentId == 0 {
		log.Errorf("RealNameComment err: %s", "comment_id empty")
		resp.Error = Error(ParasError, "comment_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("RealNameComment err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	targetComment := new(model.Comment)
	targetComment.Id = req.CommentId
	ok, err := targetComment.Get()
	if err != nil {
		log.Errorf("RealNameComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok || targetComment.IsDelete == 1 {
		log.Errorf("RealNameComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	if targetComment.CommentAnonymous != 1 {
		resp.Flag = true
		return
	}

	if targetComment.UserId != uu.Id {
		log.Errorf("RealNameComment err: %s", "comment not your's")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	_, err = targetComment.UpdateToShowName()
	if err != nil {
		log.Errorf("RealNameComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}

type CreateCommentRequest struct {
	ContentId   int64  `json:"content_id"`
	CommentId   int64  `json:"comment_id"`
	IsToComment bool   `json:"is_to_comment"`
	Body        string `json:"body"`
	Anonymous   bool   `json:"anonymous"`
	CaptchaId   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

func CreateComment(c *gin.Context) {
	resp := new(Resp)
	req := new(CreateCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	req.Body = strings.TrimSpace(req.Body)
	if len(req.Body) == 0 {
		log.Errorf("CreateComment err: %s", "body empty")
		resp.Error = Error(ParasError, "body empty")
		return
	}

	req.Body = htmlEscaper.Replace(req.Body)
	if !req.IsToComment && req.ContentId == 0 {
		log.Errorf("CreateComment err: %s", "content_id empty")
		resp.Error = Error(ParasError, "content_id empty")
		return
	}

	if req.IsToComment && req.CommentId == 0 {
		log.Errorf("CreateComment err: %s", "comment_id empty")
		resp.Error = Error(ParasError, "comment_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("CreateComment err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	// 防刷评论：近期评论过频繁则要求验证码（无感触发）
	if checkCommentBlock(uu.Id) {
		if req.CaptchaId == "" || req.CaptchaCode == "" {
			resp.Error = Error(CaptchaNeed, "")
			return
		}
		if !verifyCaptcha(req.CaptchaId, req.CaptchaCode) {
			resp.Error = Error(CaptchaWrong, "")
			return
		}
	}

	// comment to content
	if !req.IsToComment {
		content := new(model.Content)
		content.Id = req.ContentId
		ok, err := content.GetByRaw()
		if err != nil {
			log.Errorf("CreateComment err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("CreateComment err: %s", "content not found")
			resp.Error = Error(ContentNotFound, "")
			return
		}

		if content.Status == 0 && content.Version > 0 {
			if content.CloseComment == 1 {
				log.Errorf("CreateComment err: %s", "content can not comment")
				resp.Error = Error(CommentClose, "")
				return
			}
			cm := new(model.Comment)
			cm.ContentId = content.Id
			cm.ContentTitle = content.Title
			cm.ContentUserId = content.UserId
			cm.ContentUserName = content.UserName
			cm.UserId = uu.Id
			cm.UserName = uu.Name
			cm.Describe = req.Body
			cm.CommentType = model.CommentTypeOfContent
			if req.Anonymous {
				cm.CommentAnonymous = model.CommentAnonymous
			}
			err = cm.InsertOne()
			if err != nil {
				log.Errorf("CreateComment err: %s", err.Error())
				resp.Error = Error(DBError, err.Error())
				return
			}

			recordComment(uu.Id)
			go model.CommentForContent(uu.Id, content.UserId, content.Id, content.Title, cm.Id, cm.Describe, req.Anonymous)
			resp.Data = cm.Id
		} else {
			log.Errorf("CreateComment err: %s", "content status not 0 or not publish")
			if content.Status == 2 {
				resp.Error = Error(ContentBanPermit, "")
			} else {
				resp.Error = Error(ContentNotFound, "")
			}
			return
		}

		resp.Flag = true
		return
	}

	targetComment := new(model.Comment)
	targetComment.Id = req.CommentId
	ok, err := targetComment.Get()
	if err != nil {
		log.Errorf("CreateComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok || targetComment.IsDelete == 1 {
		log.Errorf("CreateComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	if targetComment.Status == 1 {
		log.Errorf("CreateComment err: %s", "comment ban")
		resp.Error = Error(CommentBanPermit, "")
		return
	}

	content := new(model.Content)
	content.Id = targetComment.ContentId
	ok, err = content.GetByRaw()
	if err != nil {
		log.Errorf("CreateComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("CreateComment err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.Status != 0 || content.Version == 0 {
		log.Errorf("CreateComment err: %s", "content status not 0 or not publish")
		if content.Status == 2 {
			resp.Error = Error(ContentBanPermit, "")
		} else {
			resp.Error = Error(ContentNotFound, "")
		}
		return
	}

	if content.CloseComment == 1 {
		log.Errorf("CreateComment err: %s", "content can not comment")
		resp.Error = Error(CommentClose, "")
		return
	}

	newComment := new(model.Comment)
	newComment.ContentId = content.Id
	newComment.ContentUserId = content.UserId
	newComment.ContentUserName = content.UserName
	newComment.ContentTitle = content.Title
	newComment.UserId = uu.Id
	newComment.UserName = uu.Name
	newComment.Describe = req.Body
	if targetComment.CommentType == model.CommentTypeOfContent {
		newComment.RootCommentId = targetComment.Id
		newComment.RootCommentUserId = targetComment.UserId
		newComment.RootCommentUserName = targetComment.UserName
		newComment.CommentType = model.CommentTypeOfRootComment
	} else {
		newComment.CommentId = targetComment.Id
		newComment.CommentUserId = targetComment.UserId
		newComment.CommentUserName = targetComment.UserName
		newComment.RootCommentId = targetComment.RootCommentId
		newComment.RootCommentUserId = targetComment.RootCommentUserId
		newComment.RootCommentUserName = targetComment.RootCommentUserName
		newComment.CommentType = model.CommentTypeOfComment
	}

	if req.Anonymous {
		newComment.CommentAnonymous = model.CommentAnonymous
	}

	err = newComment.InsertOne()
	if err != nil {
		log.Errorf("CreateComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	recordComment(uu.Id)
	// 通知被回复的评论作者：XXX 回复了你的评论（reply_comment_id 记录被回复的那条评论）
	go model.CommentForComment(uu.Id, targetComment.UserId, content.Id, content.Title, targetComment.Id, newComment.Id, newComment.Describe, req.Anonymous)
	// 通知文章作者：XXX 评论了你的内容（若与父评论作者是同一人则不再重复通知）
	if content.UserId != targetComment.UserId {
		go model.CommentForContent(uu.Id, content.UserId, content.Id, content.Title, newComment.Id, newComment.Describe, req.Anonymous)
	}
	resp.Data = newComment.Id
	resp.Flag = true
}

type DeleteCommentRequest struct {
	CommentId int64 `json:"id"`
}

func DeleteComment(c *gin.Context) {
	resp := new(Resp)
	req := new(DeleteCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CommentId == 0 {
		log.Errorf("DeleteComment err: %s", "comment_id empty")
		resp.Error = Error(ParasError, "comment_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("DeleteComment err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	comment := new(model.Comment)
	comment.Id = req.CommentId
	comment.UserId = uu.Id
	ok, err := comment.Get()
	if err != nil {
		log.Errorf("DeleteComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok || comment.IsDelete == 1 {
		log.Errorf("DeleteComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	err = comment.Delete()
	if err != nil {
		log.Errorf("DeleteComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type TakeCommentRequest struct {
	CommentId int64 `json:"id"`
}

func TakeComment(c *gin.Context) {
	resp := new(Resp)
	req := new(TakeCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CommentId == 0 {
		log.Errorf("TakeComment err: %s", "comment_id empty")
		resp.Error = Error(ParasError, "comment_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("TakeComment err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	comment := new(model.Comment)
	comment.Id = req.CommentId
	ok, err := comment.Get()
	if err != nil {
		log.Errorf("TakeComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// if not found
	if !ok || comment.IsDelete == 1 {
		log.Errorf("TakeComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	commentIds := make([]int64, 0)
	commentIds = append(commentIds, comment.Id)
	if comment.CommentType >= model.CommentTypeOfRootComment {
		commentIds = append(commentIds, comment.RootCommentId)
	}
	if comment.CommentType >= model.CommentTypeOfComment {
		commentIds = append(commentIds, comment.CommentId)
	}

	backContents, err := model.GetContentHelper([]int64{comment.ContentId}, false, uu.Id)
	if err != nil {
		log.Errorf("TakeComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	backComments, backUsers, err := model.GetCommentAndCommentUser(commentIds, false, nil, uu.Id)
	if err != nil {
		log.Errorf("TakeComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Data = map[string]interface{}{
		"comment": comment,
		"extra": model.CommentExtra{
			Users:    backUsers,
			Comments: backComments,
			Contents: backContents,
		},
	}
	resp.Flag = true

}

type ListCommentRequest struct {
	Id                  int64    `json:"id"`
	UserId              int64    `json:"user_id"`
	UserName            string   `json:"user_name"`
	ContentId           int64    `json:"content_id"`
	ContentUserId       int64    `json:"content_user_id"`
	ContentUserName     string   `json:"content_user_name"`
	CommentId           int64    `json:"comment_id"`
	CommentUserId       int64    `json:"comment_user_id"`
	CommentUserName     string   `json:"comment_user_name"`
	RootCommentId       int64    `json:"root_comment_id"`
	RootCommentUserId   int64    `json:"root_comment_user_id"`
	RootCommentUserName string   `json:"root_comment_user_name"`
	CommentType         int      `json:"comment_type" validate:"oneof=-1 0 1 2"`
	Status              int      `json:"status" validate:"oneof=-1 0 1"`
	IsDelete            int      `json:"is_delete" validate:"oneof=-1 0 1"`
	IsAnonymous         int      `json:"is_anonymous" validate:"oneof=-1 0 1"`
	CreateTimeBegin     int64    `json:"create_time_begin"`
	CreateTimeEnd       int64    `json:"create_time_end"`
	Sort                []string `json:"sort"`
	PageHelp
}

type ListCommentResponse struct {
	Comments     []model.Comment    `json:"comments"`
	CommentExtra model.CommentExtra `json:"extra"`
	PageHelp
}

func ListComment(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ListCommentResponse)
	req := new(ListCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		log.Errorf("ListComment err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Comment)).Where("1=1")

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if req.UserId != 0 {
		session.And("user_id=?", req.UserId)
	}

	if req.UserName != "" {
		uid, uerr := userFilterId(req.UserName)
		if uerr != nil {
			log.Errorf("ListCommentHelper err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("user_id=?", uid)
	}

	if req.ContentId != 0 {
		session.And("content_id=?", req.ContentId)
	}

	if req.ContentUserId != 0 {
		session.And("content_user_id=?", req.ContentUserId)
	}

	if req.ContentUserName != "" {
		uid, uerr := userFilterId(req.ContentUserName)
		if uerr != nil {
			log.Errorf("ListCommentHelper err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("content_user_id=?", uid)
	}

	if req.CommentId != 0 {
		session.And("comment_id=?", req.CommentId)
	}

	if req.CommentUserId != 0 {
		session.And("comment_user_id=?", req.CommentUserId)
	}

	if req.CommentUserName != "" {
		uid, uerr := userFilterId(req.CommentUserName)
		if uerr != nil {
			log.Errorf("ListCommentHelper err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("comment_user_id=?", uid)
	}

	if req.RootCommentId != -1 {
		session.And("root_comment_id=?", req.RootCommentId)
	}

	if req.RootCommentUserId != 0 {
		session.And("root_comment_user_id=?", req.RootCommentUserId)
	}

	if req.RootCommentUserName != "" {
		uid, uerr := userFilterId(req.RootCommentUserName)
		if uerr != nil {
			log.Errorf("ListCommentHelper err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("root_comment_user_id=?", uid)
	}

	if req.CommentType != -1 {
		session.And("comment_type=?", req.CommentType)
	}

	if req.Status != -1 {
		session.And("status=?", req.Status)
	}

	if req.IsDelete != -1 {
		session.And("is_delete=?", req.IsDelete)
	}

	if req.IsAnonymous != -1 {
		session.And("comment_anonymous=?", req.IsAnonymous)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	// if count>0 start list
	cs := make([]model.Comment, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.CommentSortName)

	// do query
	total, err := session.FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListComment err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	commentIds := make(map[int64]struct{})
	contentIds := make(map[int64]struct{})
	for _, c := range cs {
		commentIds[c.Id] = struct{}{}
		if c.CommentType >= model.CommentTypeOfRootComment && c.RootCommentId != 0 {
			commentIds[c.RootCommentId] = struct{}{}
		}
		if c.CommentType >= model.CommentTypeOfComment && c.CommentId != 0 {
			commentIds[c.CommentId] = struct{}{}
		}

		contentIds[c.ContentId] = struct{}{}
	}
	backContents, err := model.GetContentHelper(util.MapToArray(contentIds), true, 0)
	if err != nil {
		log.Errorf("ListComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	backComments, backUsers, err := model.GetCommentAndCommentUser(util.MapToArray(commentIds), true, nil, 0)
	if err != nil {
		log.Errorf("ListComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	respResult.Comments = cs
	respResult.CommentExtra = model.CommentExtra{
		Users:    backUsers,
		Comments: backComments,
		Contents: backContents,
	}
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type ListHomeCommentRequest struct {
	ContentId     int64    `json:"content_id"`
	RootCommentId int64    `json:"root_comment_id"`
	Sort          []string `json:"sort"`
	PageHelp
}

func ListHomeComment(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ListCommentResponse)
	req := new(ListHomeCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.ContentId == 0 {
		log.Errorf("ListHomeComment err: %s", "content_id empty")
		resp.Error = Error(ParasError, "content_id empty")
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Comment)).Where("1=1")

	uu, err := GetUserSession(c)
	var yourUserId int64 = 0
	if err == nil {
		yourUserId = uu.Id
	}

	if req.ContentId != 0 {
		session.And("content_id=?", req.ContentId)
	}

	if req.RootCommentId != -1 {
		session.And("root_comment_id=?", req.RootCommentId)
	}

	backContents, err := model.GetContentHelper([]int64{req.ContentId}, false, yourUserId)
	if err != nil {
		log.Errorf("ListHomeComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if len(backContents) == 0 {
		log.Errorf("ListHomeComment err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	session.And("is_delete=?", 0)

	cs := make([]model.Comment, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.CommentHomeSortName)

	// do query
	total, err := session.FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListHomeComment err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	commentIds := make(map[int64]struct{})
	for _, c := range cs {
		commentIds[c.Id] = struct{}{}
		if c.CommentType >= model.CommentTypeOfRootComment && c.RootCommentId != 0 {
			commentIds[c.RootCommentId] = struct{}{}
		}
		if c.CommentType >= model.CommentTypeOfComment && c.CommentId != 0 {
			commentIds[c.CommentId] = struct{}{}
		}
	}

	// in order to show:
	// aaaaaa
	// 	aaaa1
	//  aaaa2
	// bbbbbb
	// cccccc
	if req.RootCommentId == 0 {
		for k, c := range cs {
			innerSession := model.FaFaRdb.Client.NewSession()
			innerSession.Where("root_comment_id=?", c.Id).And("is_delete=?", 0)
			Build(innerSession, req.Sort, model.CommentHomeSortName)

			innerCs := make([]model.Comment, 0)
			num, err := innerSession.Limit(3).FindAndCount(&innerCs)
			if err != nil {
				log.Errorf("ListHomeComment err: %s", err.Error())
				resp.Error = Error(DBError, err.Error())
				innerSession.Close()
				return
			}

			if num > 0 {
				cs[k].SonNum = num
				cs[k].Son = innerCs
				for _, vv := range innerCs {
					commentIds[vv.Id] = struct{}{}
					if vv.CommentType >= model.CommentTypeOfComment && vv.CommentId != 0 {
						commentIds[vv.CommentId] = struct{}{}
					}
				}
			}

			innerSession.Close()
		}
	}
	backComments, backUsers, err := model.GetCommentAndCommentUser(util.MapToArray(commentIds), false, nil, yourUserId)
	if err != nil {
		log.Errorf("ListHomeComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	respResult.Comments = cs
	respResult.CommentExtra = model.CommentExtra{
		Users:    backUsers,
		Comments: backComments,
		Contents: backContents,
	}
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type CoolCommentRequest struct {
	CommentId int64 `json:"id"`
}

func CoolComment(c *gin.Context) {
	resp := new(Resp)
	req := new(CoolCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CommentId == 0 {
		log.Errorf("CoolComment err: %s", "comment_id empty")
		resp.Error = Error(ParasError, "comment_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("CoolComment err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	comment := new(model.Comment)
	comment.Id = req.CommentId
	ok, err := comment.Get()
	if err != nil {
		log.Errorf("CoolComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok || comment.IsDelete == 1 {
		log.Errorf("CoolComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	if comment.Status == 1 {
		log.Errorf("CoolComment err: %s", "comment ban")
		resp.Error = Error(CommentBanPermit, "")
		return
	}

	cool := new(model.CommentCool)
	cool.CommentId = req.CommentId
	cool.UserId = uu.Id
	ok, err = cool.Exist()
	if err != nil {
		log.Errorf("CoolComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	cool.ContentId = comment.ContentId

	if ok {
		err = cool.Delete()
		if err != nil {
			log.Errorf("CoolContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	} else {
		err = cool.Create()
		if err != nil {
			log.Errorf("CoolContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		} else {
			go model.GoodComment(uu.Id, comment.UserId, comment.ContentId, comment.ContentTitle, comment.Id, comment.Describe)
		}
	}

	resp.Flag = true
	if ok {
		resp.Data = "-"
	} else {
		resp.Data = "+"
	}
	return
}

type BadCommentRequest struct {
	CommentId int64  `json:"id"`
	Reason    string `json:"reason"`
}

func BadComment(c *gin.Context) {
	resp := new(Resp)
	req := new(BadCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CommentId == 0 {
		log.Errorf("BadComment err: %s", "comment_id empty")
		resp.Error = Error(ParasError, "comment_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("BadComment err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	comment := new(model.Comment)
	comment.Id = req.CommentId
	ok, err := comment.Get()
	if err != nil {
		log.Errorf("BadComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok || comment.IsDelete == 1 {
		log.Errorf("BadComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	if comment.Status == 1 {
		log.Errorf("BadComment err: %s", "comment ban")
		resp.Error = Error(CommentBanPermit, "")
		return
	}

	bad := new(model.CommentBad)
	bad.CommentId = req.CommentId
	bad.UserId = uu.Id
	bad.Reason = req.Reason
	ok, err = bad.Exist()
	if err != nil {
		log.Errorf("BadComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	bad.ContentId = comment.ContentId
	// 24 小时内已举报过：拒绝重复举报（叠加举报记录，不取消）
	if ok {
		log.Errorf("BadComment err: %s", "already bad in 24h")
		resp.Error = Error(AlreadyBad, "")
		return
	}

	err = bad.Create()
	if err != nil {
		log.Errorf("BadComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	cc := new(model.Comment)
	cc.Id = req.CommentId
	cc.UserId = comment.UserId
	cc.Describe = comment.Describe
	cc.ContentId = comment.ContentId
	cc.ContentTitle = comment.ContentTitle

	resp.Flag = true
	if AutoBan {
		err = cc.Ban(BadTime)
		if err != nil {
			log.Errorf("BadComment ban err: %s", err.Error())
		}
	}
	resp.Data = "+"
	return
}

type UpdateCommentRequest struct {
	CommentId int64 `json:"id"`
	Status    int   `json:"status"`
}

func UpdateComment(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateCommentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.CommentId == 0 {
		log.Errorf("UpdateComment err: %s", "id empty")
		resp.Error = Error(ParasError, "id empty")
		return
	}

	if req.Status != 0 && req.Status != 1 {
		log.Errorf("UpdateComment err: %s", "status should be 0 or 1")
		resp.Error = Error(ParasError, "status should be 0 or 1")
		return
	}
	comment := new(model.Comment)
	comment.Id = req.CommentId
	ok, err := comment.Get()
	if err != nil {
		log.Errorf("UpdateComment err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok || comment.IsDelete == 1 {
		log.Errorf("UpdateComment err: %s", "comment not found")
		resp.Error = Error(CommentNotFound, "")
		return
	}

	if comment.Status == req.Status {
	} else {
		comment.Status = req.Status
		_, err := comment.UpdateStatus()
		if err != nil {
			log.Errorf("UpdateComment err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
		}

	}
	resp.Flag = true
}
