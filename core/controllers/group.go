package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
	"math"
	"time"
)

type CreateGroupRequest struct {
	Name      string `json:"name" validate:"required"`
	Describe  string `json:"describe"`
	ImagePath string `json:"image_path"`
}

func CreateGroup(c *gin.Context) {
	resp := new(Resp)
	req := new(CreateGroupRequest)
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
		log.Errorf("CreateGroup err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// if exist group
	g := new(model.Group)
	g.Name = req.Name
	ok, err := g.Exist()
	if err != nil {
		log.Errorf("CreateGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if ok {
		log.Errorf("CreateGroup err: group name exist")
		resp.Error = Error(GroupNameAlreadyBeUsed, "")
		return
	}

	// if image not empty
	if req.ImagePath != "" {
		// picture table exist
		g.ImagePath = req.ImagePath
		p := new(model.File)
		p.Url = g.ImagePath
		ok, err = p.Exist()
		if err != nil {
			log.Errorf("CreateGroup err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("CreateGroup err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}

	}

	// insert now
	g.Describe = req.Describe
	g.CreateTime = time.Now().Unix()
	_, err = model.FaFaRdb.InsertOne(g)
	if err != nil {
		log.Errorf("CreateGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
	resp.Data = g
}

type UpdateGroupRequest struct {
	Id        int64  `json:"id" validate:"required"`
	Name      string `json:"name"`
	Describe  string `json:"describe"`
	ImagePath string `json:"image_path"`
}

func UpdateGroup(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateGroupRequest)
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
		log.Errorf("UpdateGroup err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// if group exist
	gg := new(model.Group)
	gg.Id = req.Id
	ok, err := gg.GetById()
	if err != nil {
		log.Errorf("UpdateGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("UpdateGroup err: group not exist")
		resp.Error = Error(GroupNotFound, "")
		return
	}

	g := new(model.Group)
	g.Id = req.Id

	// if image not empty
	if req.ImagePath != "" {
		g.ImagePath = req.ImagePath
		p := new(model.File)
		p.Url = g.ImagePath
		ok, err := p.Exist()
		if err != nil {
			log.Errorf("UpdateGroup err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("UpdateGroup err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "image url not exist")
			return
		}
	}

	// if group name change repeat
	if req.Name != "" && req.Name != gg.Name {
		g.Name = req.Name
		temp := new(model.Group)
		temp.Name = req.Name
		// exist the same name
		ok, err := temp.Exist()
		if err != nil {
			log.Errorf("UpdateGroup err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if ok {
			log.Errorf("UpdateGroup err: group name repeat")
			resp.Error = Error(GroupNameAlreadyBeUsed, "")
			return
		}
	}

	if req.Describe != "" {
		g.Describe = req.Describe
	}

	err = g.Update()
	if err != nil {
		log.Errorf("UpdateGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
	resp.Data = g
}

type DeleteGroupRequest struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func DeleteGroup(c *gin.Context) {
	resp := new(Resp)
	req := new(DeleteGroupRequest)
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
		log.Errorf("DeleteGroup err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// take group info
	temp := new(model.Group)
	temp.Id = req.Id
	temp.Name = req.Name
	ok, err := temp.Take()
	if err != nil {
		log.Errorf("DeleteGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !ok {
		log.Errorf("DeleteGroup err:%s", "group not found")
		resp.Error = Error(GroupNotFound, "")
		return
	}

	// resource exist under group
	gr := new(model.GroupResource)
	gr.GroupId = temp.Id
	ok, err = gr.Exist()
	if err != nil {
		log.Errorf("DeleteGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if ok {
		// found can not delete
		log.Errorf("DeleteGroup err:%s", "exist resource")
		resp.Error = Error(GroupHasResourceHookIn, "")
		return
	}

	// user exist under group
	u := new(model.User)
	u.GroupId = temp.Id
	ok, err = u.Exist()
	if err != nil {
		log.Errorf("DeleteGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if ok {
		// found can not delete
		log.Errorf("DeleteGroup err:%s", "exist user")
		resp.Error = Error(GroupHasUserHookIn, "exist user")
		return
	}

	// delete group
	g := new(model.Group)
	g.Id = temp.Id
	err = g.Delete()
	if err != nil {
		log.Errorf("DeleteGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type TakeGroupRequest struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func TakeGroup(c *gin.Context) {
	resp := new(Resp)
	req := new(TakeGroupRequest)
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
		log.Errorf("TakeGroup err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// take group info
	g := new(model.Group)
	g.Id = req.Id
	g.Name = req.Name
	ok, err := g.Take()
	if err != nil {
		log.Errorf("TakeGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !ok {
		log.Errorf("TakeGroup err:%s", "group not found")
		resp.Error = Error(GroupNotFound, "")
		return
	}

	resp.Flag = true
	resp.Data = g
}

type ListGroupRequest struct {
	Id              int64    `json:"id"`
	Name            string   `json:"name"`
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	UpdateTimeBegin int64    `json:"update_time_begin"`
	UpdateTimeEnd   int64    `json:"update_time_end"`
	Sort            []string `json:"sort"`
	PageHelp
}

type ListGroupResponse struct {
	Groups []model.Group `json:"groups"`
	PageHelp
}

func ListGroup(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ListGroupResponse)
	req := new(ListGroupRequest)
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
		log.Errorf("ListGroup err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Group)).Where("1=1")

	// query prepare
	if req.Id != 0 {
		session.And("id=?", req.Id)
	}
	if req.Name != "" {
		// 组名已加密，无法等值比较：改为按盲索引精确匹配（模糊搜索不可用）
		bidx, berr := model.BlindIndexOf(req.Name)
		if berr != nil {
			log.Errorf("ListGroup err: %s", berr.Error())
			resp.Error = Error(DBError, berr.Error())
			return
		}
		session.And("name_bidx=?", bidx)
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

	groups := make([]model.Group, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.GroupSortName)

	// do query
	total, err := session.FindAndCount(&groups)
	if err != nil {
		log.Errorf("ListGroup err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	respResult.Groups = groups
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type ListGroupResourceRequest struct {
	GroupId int64 `json:"group_id" validate:"required"`
}

type ListGroupResourceResponse struct {
	Resources []int64 `json:"resources"`
}

func ListGroupResource(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ListGroupResourceResponse)
	req := new(ListGroupResourceRequest)
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
		log.Errorf("ListGroupResource err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	grs := make([]model.GroupResource, 0)

	// group list where prepare
	err = session.Table(new(model.GroupResource)).Where("group_id=?", req.GroupId).Find(&grs)
	if err != nil {
		log.Errorf("ListUser err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	rs := make([]int64, 0)
	for _, v := range grs {
		rs = append(rs, v.ResourceId)
	}

	respResult.Resources = rs
	resp.Data = respResult
	resp.Flag = true
}
