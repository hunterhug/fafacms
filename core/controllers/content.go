package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	log "github.com/hunterhug/golog"
	"math"
)

type CreateContentRequest struct {
	Seo          string `json:"seo" validate:"omitempty,alphanumunicode"` // unique mark in node's content
	Title        string `json:"title" validate:"required"`                // content's title
	Status       int    `json:"status" validate:"oneof=0 1"`              // 1 stand for content hide in front end, 0 show.
	Top          int    `json:"top" validate:"oneof=0 1"`                 // 1 stand for let content on the top
	Describe     string `json:"describe" validate:"omitempty"`            // content's body
	ImagePath    string `json:"image_path" validate:"omitempty"`          // picture
	NodeId       int64  `json:"node_id"`                                  // node
	Password     string `json:"password"`                                 // if not empty will need a password in front end
	CloseComment int    `json:"close_comment" validate:"oneof=0 1"`       // 0 stand for open comment, 1 close comment
}

func CreateContent(c *gin.Context) {
	resp := new(Resp)
	req := new(CreateContentRequest)
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
		log.Errorf("CreateContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if len(req.Describe) == 0 {
		log.Errorf("CreateContent err: %s", "describe empty")
		resp.Error = Error(ParasError, "describe empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("CreateContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	if uu.Vip == 0 {
		log.Errorf("CreateContent err: %s", "not vip")
		resp.Error = Error(VipError, "")
		return
	}

	content := new(model.Content)
	content.UserId = uu.Id

	if req.NodeId == 0 {
		log.Errorf("CreateContent err: %s", "node_id can not empty")
		resp.Error = Error(ParasError, "node_id can not empty")
		return
	}

	content.NodeId = req.NodeId
	contentNode := new(model.ContentNode)
	contentNode.Id = req.NodeId
	contentNode.UserId = uu.Id
	exist, err := contentNode.Get()
	if err != nil {
		log.Errorf("CreateContent err: %s", err.Error())
		resp.Error = Error(DBError, "")
		return
	}

	if !exist {
		log.Errorf("CreateContent err: %s", "node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	content.NodeSeo = contentNode.Seo

	// 文章 SEO：留空则自动生成（前端会预填随机值，这里兜底其它客户端），
	// 唯一性范围是「所属节点内唯一」，所以必须在确定 NodeId 之后检查。
	if req.Seo != "" {
		content.Seo = req.Seo
	} else {
		content.Seo = util.GenSeo(8)
	}

	seoOk := false
	for i := 0; i < 5; i++ {
		exist, err = content.CheckSeoValid()
		if err != nil {
			log.Errorf("CreateContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if !exist {
			seoOk = true
			break
		}
		if req.Seo != "" {
			// 用户显式填写的 SEO 撞名：直接报占用
			log.Errorf("CreateContent err: %s", "seo repeat")
			resp.Error = Error(ContentSeoAlreadyBeUsed, "")
			return
		}
		// 自动生成的撞名：换一个再试
		content.Seo = util.GenSeo(8)
	}

	if !seoOk {
		log.Errorf("CreateContent err: %s", "gen seo repeat")
		resp.Error = Error(DBError, "gen seo repeat")
		return
	}

	if req.ImagePath != "" {
		content.ImagePath = req.ImagePath
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			log.Errorf("CreateContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("CreateContent err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}
	}

	content.Status = req.Status
	content.PreDescribe = req.Describe
	content.PreTitle = req.Title
	// 访问密码以 bcrypt 哈希存储（不可逆），避免数据库泄漏导致访问密码明文外泄
	if req.Password != "" {
		hash, herr := model.HashPassword(req.Password)
		if herr != nil {
			log.Errorf("CreateContent err:%s", herr.Error())
			resp.Error = Error(DBError, herr.Error())
			return
		}
		content.Password = hash
	}
	content.CloseComment = req.CloseComment
	content.Top = req.Top
	content.UserName = uu.Name
	content.SortNum, _ = content.CountNumUnderNode()
	_, err = content.Insert()
	if err != nil {
		log.Errorf("CreateContent err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Data = content
	resp.Flag = true
}

// update SEO
type UpdateSeoOfContentRequest struct {
	Id  int64  `json:"id" validate:"required"`
	Seo string `json:"seo" validate:"required,alphanumunicode"`
}

func UpdateSeoOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateSeoOfContentRequest)
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
		log.Errorf("UpdateSeoOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateSeoOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateSeoOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateSeoOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	// 文章 SEO 唯一范围是所属节点内唯一，查重必须带上节点
	content.NodeId = contentBefore.NodeId
	if req.Seo != contentBefore.Seo {
		content.Seo = req.Seo
		exist, err := content.CheckSeoValid()
		if err != nil {
			log.Errorf("UpdateSeoOfContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if exist {
			log.Errorf("UpdateSeoOfContent err: %s", "seo repeat")
			resp.Error = Error(ContentSeoAlreadyBeUsed, "")
			return
		}

		_, err = content.UpdateSeo()
		if err != nil {
			log.Errorf("UpdateSeoOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// update the picture
type UpdateImageOfContentRequest struct {
	Id        int64  `json:"id" validate:"required"`
	ImagePath string `json:"image_path" validate:"required"`
}

func UpdateImageOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateImageOfContentRequest)
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
		log.Errorf("UpdateImageOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateImageOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateImageOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateImageOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.ImagePath != contentBefore.ImagePath {
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			log.Errorf("UpdateImageOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("UpdateImageOfContent err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}

		content.ImagePath = req.ImagePath
		_, err = content.UpdateImage()
		if err != nil {
			log.Errorf("UpdateImageOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// admin user update the status of content, 0 normal, 1 hide，2 ban, 3 rubbish
type UpdateStatusOfContentAdminRequest struct {
	Id     int64 `json:"id" validate:"required"`
	Status int   `json:"status" validate:"oneof=0 1 2 3"`
}

func UpdateStatusOfContentAdmin(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateStatusOfContentAdminRequest)
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
		log.Errorf("UpdateStatusOfContentAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	exist, err := contentBefore.GetByRaw()
	if err != nil {
		log.Errorf("UpdateStatusOfContentAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateStatusOfContentAdmin err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	if req.Status != contentBefore.Status {
		content.Status = req.Status
		content.UserId = contentBefore.UserId
		content.Title = contentBefore.Title
		_, err = content.UpdateStatus(false, contentBefore.Status == 2)
		if err != nil {
			log.Errorf("UpdateStatusOfContentAdmin err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		go SendToLoop(contentBefore.UserId, 0, 1)
		go SendToLoop(contentBefore.UserId, contentBefore.NodeId, 2)
		go SendToLoop(contentBefore.UserId, 0, 3)

	}
	resp.Flag = true
}

// user update the status of content, 0 normal, 1 hide
type UpdateStatusOfContentRequest struct {
	Id     int64 `json:"id" validate:"required"`
	Status int   `json:"status" validate:"oneof=0 1"`
}

func UpdateStatusOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateStatusOfContentRequest)
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
		log.Errorf("UpdateStatusOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateStatusOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateStatusOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateStatusOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if contentBefore.Status == 2 {
		log.Errorf("UpdateStatusOfContent err: %s", "content ban")
		resp.Error = Error(ContentBanPermit, "")
		return
	}

	if contentBefore.Status == 3 {
		log.Errorf("UpdateStatusOfContent err: %s", "content rubbish")
		resp.Error = Error(ContentInRubbish, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.Status != contentBefore.Status {
		content.Status = req.Status
		_, err = content.UpdateStatus(true, false)
		if err != nil {
			log.Errorf("UpdateStatusOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		go SendToLoop(contentBefore.UserId, 0, 1)
		go SendToLoop(contentBefore.UserId, contentBefore.NodeId, 2)
		go SendToLoop(contentBefore.UserId, 0, 3)
	}
	resp.Flag = true
}

// update the node of content
type UpdateNodesOfContentRequest struct {
	Id     int64 `json:"id" validate:"required"`
	NodeId int64 `json:"node_id" validate:"required"`
}

func UpdateNodeOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateNodesOfContentRequest)
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
		log.Errorf("UpdateNodeOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateNodeOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateNodeOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateNodeOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if req.NodeId != contentBefore.NodeId {
		contentNode := new(model.ContentNode)
		contentNode.Id = req.NodeId
		contentNode.UserId = uu.Id
		exist, err := contentNode.Get()
		if err != nil {
			log.Errorf("UpdateNodeOfContent err: %s", err.Error())
			resp.Error = Error(DBError, "")
			return
		}
		if !exist {
			log.Errorf("UpdateNodeOfContent err: %s", "node not found")
			resp.Error = Error(ContentNodeNotFound, "")
			return
		}

		content := new(model.Content)
		content.Id = req.Id
		content.UserId = uu.Id
		content.NodeId = req.NodeId
		content.NodeSeo = contentNode.Seo
		content.SortNum = contentBefore.SortNum

		// 文章 SEO 在所属节点内唯一：移动到目标节点后可能与已有文章撞名。
		// 移动是常规操作，不该被 SEO 阻塞，所以撞名时自动加 4 位随机后缀。
		content.Seo = contentBefore.Seo
		check := new(model.Content)
		check.NodeId = req.NodeId
		check.Seo = contentBefore.Seo
		dup, err := check.CheckSeoValid()
		if err != nil {
			log.Errorf("UpdateNodeOfContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if dup {
			content.Seo = contentBefore.Seo + "-" + util.GenSeoSuffix(4)
		}

		err = content.UpdateNode(contentBefore.NodeId)
		if err != nil {
			log.Errorf("UpdateNodeOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		go SendToLoop(uu.Id, contentBefore.NodeId, 2)
		go SendToLoop(uu.Id, req.NodeId, 2)
	}
	resp.Flag = true
}

// update the top of content
type UpdateTopOfContentRequest struct {
	Id  int64 `json:"id" validate:"required"`
	Top int   `json:"top" validate:"oneof=0 1"`
}

func UpdateTopOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateTopOfContentRequest)
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
		log.Errorf("UpdateTopOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateTopOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateTopOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateTopOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.Top != contentBefore.Top {
		content.Top = req.Top
		_, err = content.UpdateTop()
		if err != nil {
			log.Errorf("UpdateTopOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// update the comment of content
type UpdateTopOfCommentRequest struct {
	Id           int64 `json:"id" validate:"required"`
	CloseComment int   `json:"close_comment" validate:"oneof=0 1"`
}

func UpdateCommentOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateTopOfCommentRequest)
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
		log.Errorf("UpdateCommentOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateCommentOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateCommentOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateCommentOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	if req.CloseComment != contentBefore.CloseComment {
		content.CloseComment = req.CloseComment
		_, err = content.UpdateComment()
		if err != nil {
			log.Errorf("UpdateCommentOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// update the password of content, if password empty will not need password in front end
type UpdatePasswordOfContentRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Password string `json:"password"`
}

func UpdatePasswordOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdatePasswordOfContentRequest)
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
		log.Errorf("UpdatePasswordOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdatePasswordOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdatePasswordOfContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdatePasswordOfContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id

	// 访问密码以 bcrypt 哈希存储（不可逆），此处判断"是否需要变更"：
	//   - 传空：表示清除访问密码
	//   - 传非空：与现有哈希比对，不同（或历史明文需升级）才重新哈希
	passwordChanged := false
	if req.Password == "" {
		passwordChanged = contentBefore.Password != ""
	} else {
		ok, needUpgrade := model.CheckPassword(contentBefore.Password, req.Password)
		passwordChanged = !ok || needUpgrade
	}

	if passwordChanged {
		if req.Password == "" {
			content.Password = ""
		} else {
			hash, herr := model.HashPassword(req.Password)
			if herr != nil {
				log.Errorf("UpdatePasswordOfContent err:%s", herr.Error())
				resp.Error = Error(DBError, herr.Error())
				return
			}
			content.Password = hash
		}

		_, err = content.UpdatePassword()
		if err != nil {
			log.Errorf("UpdatePasswordOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// update the body and title of content
type UpdateInfoOfContentRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Title    string `json:"title" validate:"required"`
	Describe string `json:"describe" validate:"omitempty"`
	Save     bool   `json:"save"`
}

func UpdateInfoOfContent(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateInfoOfContentRequest)
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
		log.Errorf("UpdateInfoOfContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if len(req.Describe) == 0 {
		log.Errorf("UpdateInfoOfContent err: %s", "describe empty")
		resp.Error = Error(ParasError, "describe empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateInfoOfContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	if uu.Vip == 0 {
		log.Errorf("UpdateInfoOfContent err: %s", "not vip")
		resp.Error = Error(VipError, "")
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("UpdateInfoOfContent err: %s", err.Error())
		resp.Error = Error(DBError, "")
		return
	}

	if !exist {
		log.Errorf("UpdateInfoOfContent err: %s", "content not found")
		resp.Error = Error(DbNotFound, "content not found")
		return
	}

	if contentBefore.PreDescribe != req.Describe || contentBefore.PreTitle != req.Title {
		content := new(model.Content)
		content.Id = req.Id
		content.UserId = uu.Id
		content.NodeId = contentBefore.NodeId
		content.PreDescribe = contentBefore.PreDescribe
		content.PreTitle = contentBefore.PreTitle
		content.PreFlush = contentBefore.PreFlush
		content.Describe = req.Describe
		content.Title = req.Title
		err = content.UpdateDescribeAndHistory(req.Save)
		if err != nil {
			log.Errorf("UpdateInfoOfContent err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

// put Y on top of X
// can drag sort
type SortContentRequest struct {
	XID int64 `json:"xid" validate:"required"`
	YID int64 `json:"yid"`
}

// sort the content in a skr way
// sort_num more small, the more forward the content is
func SortContent(c *gin.Context) {
	resp := new(Resp)
	req := new(SortContentRequest)
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
		log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if req.XID == req.YID {
		log.Errorf("SortContent err: %s", "xid=yid not right")
		resp.Error = Error(ParasError, "xid=yid not right")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	x := new(model.Content)
	x.Id = req.XID
	x.UserId = uu.Id
	exist, err := x.Get()
	if err != nil {
		log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("SortContent err: %s", "x content not found")
		resp.Error = Error(ContentNotFound, "x content not found")
		return
	}

	// 收集同节点且同置顶分组的文章（按 sort_num, id 排序），重建连续 sort_num，避免重复/空洞导致的排序错乱
	// 置顶(top=1)与非置顶(top=0)是两个独立排序域：置顶永远在前，拖拽只在同组内重排
	var all []model.Content
	err = model.FaFaRdb.Client.Where("user_id=?", uu.Id).And("node_id=?", x.NodeId).And("top=?", x.Top).Asc("sort_num").Asc("id").Find(&all)
	if err != nil {
		log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 目标顺序：先移除 X，再插入到 Y 之后（YID=0 表示插到最前）
	order := make([]int64, 0, len(all))
	for _, cc := range all {
		if cc.Id != x.Id {
			order = append(order, cc.Id)
		}
	}
	insertAt := 0
	if req.YID != 0 {
		found := false
		for i, id := range order {
			if id == req.YID {
				insertAt = i + 1
				found = true
				break
			}
		}
		if !found {
			log.Errorf("SortContent err: %s", "y content not in same node")
			resp.Error = Error(ContentNotFound, "y content not found")
			return
		}
	}

	newOrder := make([]int64, 0, len(order)+1)
	newOrder = append(newOrder, order[:insertAt]...)
	newOrder = append(newOrder, x.Id)
	newOrder = append(newOrder, order[insertAt:]...)

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()
	err = session.Begin()
	if err != nil {
		log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	for i, id := range newOrder {
		_, err = session.Exec("update fafacms_content set sort_num=? where id=? and user_id=?", i, id, uu.Id)
		if err != nil {
			session.Rollback()
			log.Errorf("SortContent err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	err = session.Commit()
	if err != nil {
		session.Rollback()
		log.Errorf("SortContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
	return
}

type PublishContentRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func PublishContent(c *gin.Context) {
	resp := new(Resp)
	req := new(PublishContentRequest)
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
		log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	exist, err := content.Get()
	if err != nil {
		log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("PublishContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.PreFlush == 1 {
		resp.Flag = true
		return
	}

	err = content.PublishDescribe()
	if err != nil {
		log.Errorf("PublishContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if content.Version == 1 {
		go model.PublishContent(uu.Id, 0, content.Id, content.Title, false)
		go SendToLoop(uu.Id, 0, 1)
		go SendToLoop(uu.Id, content.NodeId, 2)
	} else {
		go model.PublishContent(uu.Id, 0, content.Id, content.Title, true)
	}
	resp.Flag = true
}

type RestoreContentRequest struct {
	HistoryId int64 `json:"history_id" validate:"required"`
	Save      bool  `json:"save"`
}

func RestoreContent(c *gin.Context) {
	resp := new(Resp)
	req := new(RestoreContentRequest)
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
		log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentH := new(model.ContentHistory)
	contentH.Id = req.HistoryId
	contentH.UserId = uu.Id
	exist, err := contentH.GetRaw()
	if err != nil {
		log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("RestoreContent err: %s", "content history not found")
		resp.Error = Error(ContentHistoryNotFound, "")
		return
	}

	content := new(model.Content)
	content.Id = contentH.ContentId
	content.UserId = uu.Id
	exist, err = content.Get()
	if err != nil {
		log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("RestoreContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	content.Title = contentH.Title
	content.Describe = contentH.Describe
	err = content.ResetDescribe(req.Save)
	if err != nil {
		log.Errorf("RestoreContent err: %s", err.Error())
		resp.Error = Error(DBError, "")
		return
	}
	resp.Flag = true
}

type ListContentRequest struct {
	Id                    int64    `json:"id"`
	Seo                   string   `json:"seo" validate:"omitempty,alphanumunicode"`
	NodeId                int64    `json:"node_id"`
	NodeSeo               string   `json:"node_seo"`
	Top                   int      `json:"top" validate:"oneof=-1 0 1"`
	Status                int      `json:"status" validate:"oneof=-1 0 1 2 3 4"`
	CloseComment          int      `json:"close_comment" validate:"oneof=-1 0 1"`
	PasswordType          int      `json:"password_type" validate:"oneof=-1 0 1"`
	PublishType           int      `json:"publish_type" validate:"oneof=-1 0 1 2 3"`
	UserId                int64    `json:"user_id"`
	UserName              string   `json:"user_name"`
	Title                 string   `json:"title"`
	CreateTimeBegin       int64    `json:"create_time_begin"`
	CreateTimeEnd         int64    `json:"create_time_end"`
	UpdateTimeBegin       int64    `json:"update_time_begin"`
	UpdateTimeEnd         int64    `json:"update_time_end"`
	FirstPublishTimeBegin int64    `json:"first_publish_time_begin"`
	FirstPublishTimeEnd   int64    `json:"first_publish_time_end"`
	PublishTimeBegin      int64    `json:"publish_time_begin"`
	PublishTimeEnd        int64    `json:"publish_time_end"`
	Sort                  []string `json:"sort"`
	PageHelp
}

type ListContentResponse struct {
	Contents []model.Content `json:"contents"`
	PageHelp
}

func ListContent(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListContentHelper(c, uid)
}

func ListContentAdmin(c *gin.Context) {
	ListContentHelper(c, 0)
}

func ListContentHelper(c *gin.Context, userId int64) {
	resp := new(Resp)

	respResult := new(ListContentResponse)
	req := new(ListContentRequest)
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
		log.Errorf("ListContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Content)).Where("1=1")

	// query prepare
	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if userId != 0 {
		session.And("user_id=?", userId)

		// 不设置条件，只列出非垃圾（排除回收站 3 与用户彻底删除 4）
		if req.Status == -1 {
			session.And("status < ?", 3)
		} else {
			session.And("status=?", req.Status)
		}

	} else {
		if req.Status != -1 {
			session.And("status=?", req.Status)
		}
		if req.UserName != "" {
			uid, uerr := userFilterId(req.UserName)
			if uerr != nil {
				log.Errorf("ListContentHelper err: %s", uerr.Error())
				resp.Error = Error(DBError, uerr.Error())
				return
			}
			session.And("user_id=?", uid)
		}
		if req.UserId != 0 {
			session.And("user_id=?", req.UserId)
		}
	}

	if req.Top != -1 {
		session.And("top=?", req.Top)
	}

	if req.Title != "" {
		// 标题列已加密，无法模糊匹配：改为按标题盲索引做精确匹配（正式标题 + 草稿标题）
		bidx, berr := model.BlindIndexOf(req.Title)
		if berr != nil {
			log.Errorf("ListContentHelper err: %s", berr.Error())
			resp.Error = Error(DBError, berr.Error())
			return
		}
		session.And("(title_bidx=? OR pre_title_bidx=?)", bidx, bidx)
	}

	if req.PasswordType != -1 {
		if req.PasswordType == 0 {
			session.And("password=?", "")
		} else {
			session.And("password!=?", "")
		}
	}

	if req.PublishType != -1 {
		if req.PublishType == 0 {
			session.And("version=?", 0)
		} else if req.PublishType == 1 {
			session.And("version>?", 0)
		} else if req.PublishType == 2 {
			session.And("version>?", 0)
			session.And("pre_flush=?", 1)
		} else {
			session.And("version>?", 0)
			session.And("pre_flush=?", 0)
		}
	}
	if req.Seo != "" {
		session.And("seo=?", req.Seo)
	}

	if req.CloseComment != -1 {
		session.And("close_comment=?", req.CloseComment)
	}

	if req.NodeId != 0 {
		session.And("node_id=?", req.NodeId)
	}

	if req.NodeSeo != "" {
		session.And("node_seo=?", req.NodeSeo)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.PublishTimeBegin > 0 {
		session.And("publish_time>=?", req.PublishTimeBegin)
	}

	if req.PublishTimeEnd > 0 {
		session.And("publish_time<?", req.PublishTimeEnd)
	}

	if req.FirstPublishTimeBegin > 0 {
		session.And("first_publish_time>=?", req.FirstPublishTimeBegin)
	}

	if req.FirstPublishTimeEnd > 0 {
		session.And("first_publish_time<?", req.FirstPublishTimeEnd)
	}

	// if count>0 start list
	cs := make([]model.Content, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.ContentSortName)

	// do query
	total, err := session.Omit("describe", "pre_describe").FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListContent err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	respResult.Contents = cs
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type ListContentHistoryRequest struct {
	Id              int64    `json:"content_id"`
	UserId          int64    `json:"user_id"`
	Types           int      `json:"types" validate:"oneof=-1 0 1 2"`
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	Sort            []string `json:"sort"`
	PageHelp
}

type ListContentHistoryResponse struct {
	Contents []model.ContentHistory `json:"contents"`
	PageHelp
}

func ListContentHistory(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListContentHistory err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListContentHistoryHelper(c, uid)
}

func ListContentHistoryAdmin(c *gin.Context) {
	ListContentHistoryHelper(c, 0)
}

func ListContentHistoryHelper(c *gin.Context, userId int64) {
	resp := new(Resp)

	respResult := new(ListContentHistoryResponse)
	req := new(ListContentHistoryRequest)
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
		log.Errorf("ListContentHistory err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.ContentHistory)).Where("1=1")

	if req.Id != 0 {
		session.And("content_id=?", req.Id)
	}

	if userId != 0 {
		session.And("user_id=?", userId)
	} else {
		if req.UserId != 0 {
			session.And("user_id=?", req.UserId)
		}
	}

	if req.Types != -1 {
		session.And("types=?", req.Types)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	// if count>0 start list
	cs := make([]model.ContentHistory, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.ContentHistorySortName)

	// do query
	total, err := session.Omit("describe").FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListContentHistory err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	respResult.Contents = cs
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type TakeContentRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func TakeContentHelper(c *gin.Context, userId int64) {
	resp := new(Resp)
	req := new(TakeContentRequest)
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
		log.Errorf("TakeContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = userId
	exist, err := content.Get()
	if err != nil {
		log.Errorf("TakeContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("TakeContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	resp.Data = content
	resp.Flag = true
}

func TakeContent(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("TakeContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	TakeContentHelper(c, uid)
}

func TakeContentAdmin(c *gin.Context) {
	TakeContentHelper(c, 0)
}

type TakeContentHistoryRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func TakeContentHistoryHelper(c *gin.Context, userId int64) {
	resp := new(Resp)
	req := new(TakeContentHistoryRequest)
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
		log.Errorf("TakeContentHistory err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	content := new(model.ContentHistory)
	content.Id = req.Id
	content.UserId = userId
	exist, err := content.GetRaw()
	if err != nil {
		log.Errorf("TakeContentHistory err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("TakeContentHistory err: %s", "content history not found")
		resp.Error = Error(ContentHistoryNotFound, "")
		return
	}

	resp.Data = content
	resp.Flag = true
}

func TakeContentHistory(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("TakeContentHistory err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	TakeContentHistoryHelper(c, uid)
}

func TakeContentHistoryAdmin(c *gin.Context) {
	TakeContentHistoryHelper(c, 0)
}

type SentContentToRubbishRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func SentContentToRubbish(c *gin.Context) {
	resp := new(Resp)
	req := new(SentContentToRubbishRequest)
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
		log.Errorf("SentContentToRubbish err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("SentContentToRubbish err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	if uu.Vip == 0 {
		log.Errorf("SentContentToRubbish err: %s", "not vip")
		resp.Error = Error(VipError, "")
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("SentContentToRubbish err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("SentContentToRubbish err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if contentBefore.Status == 3 {
		resp.Flag = true
		return
	}

	//if contentBefore.Status == 2 {
	//	log.Errorf("SentContentToRubbish err: %s", "can not sent to rubbish")
	//	resp.Error = Error(ContentBanPermit, "can not sent to rubbish")
	//	return
	//}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	content.Status = 3
	_, err = content.UpdateStatus(true, false)
	if err != nil {
		log.Errorf("SentContentToRubbish err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	go SendToLoop(uu.Id, 0, 1)
	go SendToLoop(uu.Id, contentBefore.NodeId, 2)
	go SendToLoop(uu.Id, 0, 3)
	resp.Flag = true
}

type ReCycleOfContentInRubbishRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func ReCycleOfContentInRubbish(c *gin.Context) {
	resp := new(Resp)
	req := new(ReCycleOfContentInRubbishRequest)
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
		log.Errorf("ReCycleOfContentInRubbish err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ReCycleOfContentInRubbish err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("ReCycleOfContentInRubbish err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("ReCycleOfContentInRubbish err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if contentBefore.Status == 3 {
		content := new(model.Content)
		content.Id = req.Id
		content.UserId = uu.Id

		if contentBefore.BanTime == 0 {
			content.Status = 1
		} else {
			content.Status = 2
		}
		_, err = content.UpdateStatus(true, false)
		if err != nil {
			log.Errorf("ReCycleOfContentInRubbish err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		go SendToLoop(uu.Id, 0, 1)
		go SendToLoop(uu.Id, contentBefore.NodeId, 2)
		go SendToLoop(uu.Id, 0, 3)
	}

	resp.Flag = true
}

type ReallyDeleteContentRequest struct {
	Id int64 `json:"id" validate:"required"`
}

// ReallyDeleteContent 用户从回收站「彻底删除」：仍是软删除（status=4），
// 仅从用户世界消失，管理后台仍可见。
func ReallyDeleteContent(c *gin.Context) {
	resp := new(Resp)
	req := new(ReallyDeleteContentRequest)
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
		log.Errorf("ReallyDeleteContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ReallyDeleteContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBefore := new(model.Content)
	contentBefore.Id = req.Id
	contentBefore.UserId = uu.Id
	exist, err := contentBefore.Get()
	if err != nil {
		log.Errorf("ReallyDeleteContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("ReallyDeleteContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	// 仅回收站(status=3)中的内容可彻底删除
	if contentBefore.Status != 3 {
		log.Errorf("ReallyDeleteContent err: %s", "content can not delete")
		resp.Error = Error(ContentCanNotDelete, "")
		return
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = uu.Id
	content.Status = 4 // 软删除：用户世界消失，管理员仍可见
	_, err = content.UpdateStatus(true, false)
	if err != nil {
		log.Errorf("ReallyDeleteContent err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	go SendToLoop(uu.Id, 0, 1)
	go SendToLoop(uu.Id, contentBefore.NodeId, 2)
	go SendToLoop(uu.Id, 0, 3)
	resp.Flag = true
}

type ReallyDeleteContentHistoryRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func ReallyDeleteHistoryContent(c *gin.Context) {
	resp := new(Resp)
	req := new(ReallyDeleteContentHistoryRequest)
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
		log.Errorf("ReallyDeleteHistoryContent err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ReallyDeleteHistoryContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	contentBeforeH := new(model.ContentHistory)
	contentBeforeH.Id = req.Id
	contentBeforeH.UserId = uu.Id
	exist, err := contentBeforeH.GetRaw()
	if err != nil {
		log.Errorf("ReallyDeleteHistoryContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("ReallyDeleteHistoryContent err: %s", "content not found")
		resp.Error = Error(ContentHistoryNotFound, "")
		return
	}

	contentH := new(model.ContentHistory)
	contentH.Id = req.Id
	contentH.UserId = uu.Id
	_, err = contentH.Delete()
	if err != nil {
		log.Errorf("ReallyDeleteHistoryContent err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}
