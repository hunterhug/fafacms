package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
)

// ============ 公开：站点配置 + 可见友情链接 ============

type SiteConfigResponse struct {
	SiteTitle    string             `json:"site_title"`
	SiteSubtitle string             `json:"site_subtitle"`
	SiteIntro    string             `json:"site_intro"`
	FooterIntro  string             `json:"footer_intro"`
	FriendLinks  []model.FriendLink `json:"friend_links"`
}

func SiteConfig(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		JSONL(c, 200, nil, resp)
	}()

	cfg, err := model.GetSiteConfig()
	if err != nil {
		log.Errorf("SiteConfig err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	links, err := model.ListVisibleFriendLink()
	if err != nil {
		log.Errorf("SiteConfig err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	respResult := &SiteConfigResponse{
		SiteTitle:    cfg.SiteTitle,
		SiteSubtitle: cfg.SiteSubtitle,
		SiteIntro:    cfg.SiteIntro,
		FooterIntro:  cfg.FooterIntro,
		FriendLinks:  links,
	}
	resp.Data = respResult
	resp.Flag = true
}

// ============ 管理：站点配置 ============

type UpdateSiteConfigRequest struct {
	SiteTitle    string `json:"site_title" validate:"required"`
	SiteSubtitle string `json:"site_subtitle"`
	SiteIntro    string `json:"site_intro"`
	FooterIntro  string `json:"footer_intro"`
}

func UpdateSiteConfig(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateSiteConfigRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("UpdateSiteConfig err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if err := model.UpdateSiteConfig(req.SiteTitle, req.SiteSubtitle, req.SiteIntro, req.FooterIntro); err != nil {
		log.Errorf("UpdateSiteConfig err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

// ============ 管理：友情链接 ============

type FriendLinkRequest struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Url     string `json:"url"`
	SortNum *int   `json:"sort_num"`
	Hide    *int   `json:"hide"`
	OpenNew *int   `json:"open_new"`
}

type ListFriendLinkRequest struct {
	PageHelp
}

type ListFriendLinkResponse struct {
	FriendLinks []model.FriendLink `json:"friend_links"`
	PageHelp
}

func ListFriendLinkAdmin(c *gin.Context) {
	resp := new(Resp)
	respResult := new(ListFriendLinkResponse)
	req := new(ListFriendLinkRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.FriendLink)).Where("1=1")

	fs := make([]model.FriendLink, 0)
	p := &req.PageHelp
	p.build(session, []string{"+sort_num", "=id"}, model.FriendLinkSortName)

	total, err := session.FindAndCount(&fs)
	if err != nil {
		log.Errorf("ListFriendLinkAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	respResult.FriendLinks = fs
	p.Pages = totalPages(total, p.Limit)
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

func CreateFriendLink(c *gin.Context) {
	resp := new(Resp)
	req := new(FriendLinkRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Name == "" || req.Url == "" {
		log.Errorf("CreateFriendLink err: %s", "name or url empty")
		resp.Error = Error(ParasError, "name or url empty")
		return
	}

	f := new(model.FriendLink)
	f.Name = req.Name
	f.Url = req.Url
	// defaults: hide=0, open_new=1, sort_num = max+1
	f.Hide = 0
	f.OpenNew = 1
	if req.Hide != nil {
		f.Hide = *req.Hide
	}
	if req.OpenNew != nil {
		f.OpenNew = *req.OpenNew
	}
	if req.SortNum != nil {
		f.SortNum = *req.SortNum
	} else {
		f.SortNum = nextFriendLinkSortNum()
	}

	if err := f.InsertOne(); err != nil {
		log.Errorf("CreateFriendLink err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
	resp.Data = f
}

func UpdateFriendLink(c *gin.Context) {
	resp := new(Resp)
	req := new(FriendLinkRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("UpdateFriendLink err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if req.Id == 0 {
		log.Errorf("UpdateFriendLink err: id empty")
		resp.Error = Error(ParasError, "id empty")
		return
	}

	old := new(model.FriendLink)
	old.Id = req.Id
	exist, err := old.Get()
	if err != nil {
		log.Errorf("UpdateFriendLink err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("UpdateFriendLink err: friend link not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	f := new(model.FriendLink)
	f.Id = req.Id
	f.Name = old.Name
	f.Url = old.Url
	f.SortNum = old.SortNum
	f.Hide = old.Hide
	f.OpenNew = old.OpenNew

	if req.Name != "" {
		f.Name = req.Name
	}
	if req.Url != "" {
		f.Url = req.Url
	}
	if req.SortNum != nil {
		f.SortNum = *req.SortNum
	}
	if req.Hide != nil {
		f.Hide = *req.Hide
	}
	if req.OpenNew != nil {
		f.OpenNew = *req.OpenNew
	}

	if err := f.Update(); err != nil {
		log.Errorf("UpdateFriendLink err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
	resp.Data = f
}

type DeleteFriendLinkRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func DeleteFriendLink(c *gin.Context) {
	resp := new(Resp)
	req := new(DeleteFriendLinkRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("DeleteFriendLink err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	f := new(model.FriendLink)
	f.Id = req.Id
	if err := f.Delete(); err != nil {
		log.Errorf("DeleteFriendLink err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type SortFriendLinkRequest struct {
	Ids []int64 `json:"ids"`
}

func SortFriendLink(c *gin.Context) {
	resp := new(Resp)
	req := new(SortFriendLinkRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if len(req.Ids) == 0 {
		log.Errorf("SortFriendLink err: ids empty")
		resp.Error = Error(ParasError, "ids empty")
		return
	}

	for i, id := range req.Ids {
		f := new(model.FriendLink)
		f.Id = id
		f.SortNum = i
		if err := f.UpdateSortNum(); err != nil {
			log.Errorf("SortFriendLink err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	resp.Flag = true
}

// nextFriendLinkSortNum max sort_num + 1
func nextFriendLinkSortNum() int {
	f := new(model.FriendLink)
	has, err := model.FaFaRdb.Client.Table(new(model.FriendLink)).Desc("sort_num").Get(f)
	if err != nil || !has {
		return 0
	}
	return f.SortNum + 1
}

func totalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}
