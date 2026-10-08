package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
)

type CreateNodeRequest struct {
	Seo          string `json:"seo" validate:"omitempty,alphanumunicode"`
	Name         string `json:"name" validate:"required"`
	Describe     string `json:"describe"`
	ImagePath    string `json:"image_path"`
	ParentNodeId int64  `json:"parent_node_id"`
}

func CreateNode(c *gin.Context) {
	resp := new(Resp)
	req := new(CreateNodeRequest)
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
		log.Errorf("CreateNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("CreateNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	if uu.Vip == 0 {
		log.Errorf("CreateNode err: %s", "not vip")
		resp.Error = Error(VipError, "")
		return
	}

	n := new(model.ContentNode)
	n.UserId = uu.Id

	// If seo not empty, check valid
	if req.Seo != "" {
		n.Seo = req.Seo
		exist, err := n.CheckSeoValid()
		if err != nil {
			log.Errorf("CreateNode err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if exist {
			log.Errorf("CreateNode err: %s", "node seo already be use")
			resp.Error = Error(ContentNodeSeoAlreadyBeUsed, "")
			return
		}
	} else {
		resp.Error = Error(ParasError, "seo can not empty")
		return
	}

	// if node has parent
	if req.ParentNodeId != 0 {
		n.ParentNodeId = req.ParentNodeId
		exist, err := n.CheckParentValid()
		if err != nil {
			log.Errorf("CreateNode err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if !exist {
			// parent not exist
			log.Errorf("CreateNode err: %s", "parent content node not found")
			resp.Error = Error(ContentParentNodeNotFound, "")
			return
		}

		n.Level = 1
	}

	// if image not empty
	if req.ImagePath != "" {
		n.ImagePath = req.ImagePath
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			log.Errorf("CreateNode err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("CreateNode err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "image url not exist")
			return
		}
	}
	n.Name = req.Name
	n.Describe = req.Describe
	n.ParentNodeId = req.ParentNodeId
	n.UserName = uu.Name
	n.SortNum, _ = n.CountNodeNum()
	err = n.InsertOne()
	if err != nil {
		log.Errorf("CreateNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
	resp.Data = n
}

type UpdateInfoOfNodeRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Name     string `json:"name"`
	Describe string `json:"describe"`
}

type UpdateImageOfNodeRequest struct {
	Id        int64  `json:"id" validate:"required"`
	ImagePath string `json:"image_path" validate:"required"`
}

type UpdateStatusOfNodeRequest struct {
	Id     int64 `json:"id" validate:"required"`
	Status int   `json:"status" validate:"oneof=0 1"`
}

type UpdateSeoOfNodeRequest struct {
	Id  int64  `json:"id" validate:"required"`
	Seo string `json:"seo" validate:"required,alphanumunicode"`
}

type UpdateParentOfNodeRequest struct {
	Id           int64 `json:"id" validate:"required"`
	ToBeRoot     bool  `json:"to_be_root"` // let the node to be root node, in the first level
	ParentNodeId int64 `json:"parent_node_id"`
}

func UpdateSeoOfNode(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateSeoOfNodeRequest)
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
		log.Errorf("UpdateSeoOfNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateSeoOfNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	n := new(model.ContentNode)
	n.Id = req.Id
	n.UserId = uu.Id

	// Get info of node
	exist, err := n.Get()
	if err != nil {
		log.Errorf("UpdateSeoOfNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("UpdateSeoOfNode err: %s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	after := new(model.ContentNode)
	after.UserId = n.UserId
	after.Id = n.Id

	seoChange := false

	// Seo change
	if req.Seo != n.Seo {
		after.Seo = req.Seo
		seoChange = true
		// check seo is valid
		exist, err := after.CheckSeoValid()
		if err != nil {
			log.Errorf("UpdateSeoOfNode err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if exist {
			// SEO been occupy will err
			log.Errorf("UpdateSeoOfNode err: %s", "seo been used")
			resp.Error = Error(ContentNodeSeoAlreadyBeUsed, "")
			return
		}
	}

	if seoChange {
		// update the seo
		err = after.UpdateSeo()
		if err != nil {
			log.Errorf("UpdateSeoOfNode err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	resp.Flag = true
}

func UpdateInfoOfNode(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateInfoOfNodeRequest)
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
		log.Errorf("UpdateInfoOfNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateInfoOfNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, "")
		return
	}
	n := new(model.ContentNode)
	n.Id = req.Id
	n.UserId = uu.Id

	exist, err := n.Get()
	if err != nil {
		log.Errorf("UpdateInfoOfNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("UpdateInfoOfNode err: %s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	after := new(model.ContentNode)
	after.UserId = n.UserId
	after.Id = n.Id

	// Only name change will update
	if req.Name != "" {
		if req.Name != n.Name {
			after.Name = req.Name
		}
	}

	after.Describe = req.Describe

	err = after.UpdateInfo()
	if err != nil {
		log.Errorf("UpdateNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}

func UpdateImageOfNode(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateImageOfNodeRequest)
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
		log.Errorf("UpdateInfoOfNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateInfoOfNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, "")
		return
	}
	n := new(model.ContentNode)
	n.Id = req.Id
	n.UserId = uu.Id

	exist, err := n.Get()
	if err != nil {
		log.Errorf("UpdateInfoOfNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("UpdateInfoOfNode err: %s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	after := new(model.ContentNode)
	after.UserId = n.UserId
	after.Id = n.Id

	if req.ImagePath != n.ImagePath {
		after.ImagePath = req.ImagePath
		p := new(model.File)
		p.Url = req.ImagePath
		ok, err := p.Exist()
		if err != nil {
			log.Errorf("UpdateInfoOfNode err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			log.Errorf("UpdateInfoOfNode err: image not exist")
			resp.Error = Error(FileCanNotBeFound, "")
			return
		}

		err = after.UpdateImage()
		if err != nil {
			log.Errorf("UpdateNode err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	resp.Flag = true
}

func UpdateStatusOfNode(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateStatusOfNodeRequest)
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
		log.Errorf("UpdateStatusOfNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateStatusOfNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	n := new(model.ContentNode)
	n.Id = req.Id
	n.UserId = uu.Id

	exist, err := n.Get()
	if err != nil {
		log.Errorf("UpdateStatusOfNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("UpdateStatusOfNode err: %s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	after := new(model.ContentNode)
	after.UserId = n.UserId
	after.Id = n.Id
	after.Status = req.Status

	err = after.UpdateStatus()
	if err != nil {
		log.Errorf("UpdateStatusOfNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	go SendToLoop(n.UserId, 0, 1)
	go SendToLoop(n.UserId, 0, 3)
	resp.Flag = true
}

// UpdateStatusOfNodeAdmin 管理员隐藏/显示任意用户的节点。
// 权限由路由上的 Admin 标记 + 用户组资源校验保证（AuthFilter），这里只做数据操作。
// 语义与本人接口一致：隐藏节点后，该节点及其子节点下的文章退出公开列表，
// 只有作者本人仍能看到；管理后台的列表接口不受影响。
func UpdateStatusOfNodeAdmin(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateStatusOfNodeRequest)
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
		log.Errorf("UpdateStatusOfNodeAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// 管理员不限用户：先按 id 取节点，再改状态
	exist, err := model.FaFaRdb.Client.ID(req.Id).Get(new(model.ContentNode))
	if err != nil {
		log.Errorf("UpdateStatusOfNodeAdmin err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("UpdateStatusOfNodeAdmin err:%s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	_, err = model.FaFaRdb.Client.Table(new(model.ContentNode)).
		Where("id=?", req.Id).
		Cols("status").
		Update(map[string]interface{}{"status": req.Status})
	if err != nil {
		log.Errorf("UpdateStatusOfNodeAdmin err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

func UpdateParentOfNode(c *gin.Context) {
	resp := new(Resp)
	req := new(UpdateParentOfNodeRequest)
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
		log.Errorf("UpdateParentOfNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if req.ParentNodeId == req.Id {
		log.Errorf("UpdateParentOfNode err: %s", "self can not be parent")
		resp.Error = Error(ParasError, "self can not be parent")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateParentOfNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	n := new(model.ContentNode)
	n.Id = req.Id
	n.UserId = uu.Id

	exist, err := n.Get()
	if err != nil {
		log.Errorf("UpdateParentOfNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UpdateParentOfNode err: %s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	// Who has children can not be child due to we only design 2 level
	childNum, err := n.CheckChildrenNum()
	if err != nil {
		log.Errorf("UpdateParentOfNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if childNum > 0 {
		log.Errorf("UpdateParentOfNode err: %s", "has child")
		resp.Error = Error(ContentNodeHasChildren, "has child")
		return
	}

	beforeParentNode := n.ParentNodeId

	after := new(model.ContentNode)
	after.UserId = n.UserId
	after.Id = n.Id
	after.SortNum = n.SortNum

	// Let the node to be the first level
	if req.ToBeRoot {
		// has been
		if n.ParentNodeId == 0 {
			resp.Flag = true
			return
		}
		// level first and parent zero
		after.Level = 0
		after.ParentNodeId = 0
	} else {
		// not change at all
		if n.ParentNodeId == req.ParentNodeId {
			resp.Flag = true
			return
		}

		after.ParentNodeId = req.ParentNodeId

		// parent is exit?
		exist, err := after.CheckParentValid()
		if err != nil {
			log.Errorf("UpdateParentOfNode err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if !exist {
			log.Errorf("UpdateParentOfNode err: %s", "parent content node not found")
			resp.Error = Error(ContentParentNodeNotFound, "")
			return
		}

		// set in to 1
		after.Level = 1
	}

	err = after.UpdateParent(beforeParentNode)
	if err != nil {
		log.Errorf("UpdateParentOfNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}

type DeleteNodeRequest struct {
	Id int64 `json:"id" validate:"required"`
}

// Delete node, those nodes after it will be auto sorted
func DeleteNode(c *gin.Context) {
	resp := new(Resp)
	req := new(DeleteNodeRequest)
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
		log.Errorf("DeleteNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("DeleteNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}
	n := new(model.ContentNode)
	n.Id = req.Id
	n.UserId = uu.Id

	exist, err := n.Get()
	if err != nil {
		log.Errorf("DeleteNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	if !exist {
		log.Errorf("DeleteNode err: %s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	// can not delete when has node children
	childNum, err := n.CheckChildrenNum()
	if err != nil {
		log.Errorf("DeleteNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if childNum >= 1 {
		log.Errorf("DeleteNode err:%s", "has node child")
		resp.Error = Error(ContentNodeHasChildren, "")
		return
	}

	content := new(model.Content)
	content.UserId = uu.Id
	content.NodeId = n.Id

	// can not delete when has content
	normalContentNum, err := content.CountNumUnderNode()
	if err != nil {
		log.Errorf("DeleteNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if normalContentNum >= 1 {
		log.Errorf("DeleteNode err:%s", "has content child")
		resp.Error = Error(ContentNodeHasContentCanNotDelete, "")
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	err = session.Begin()
	if err != nil {
		log.Errorf("DeleteNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// sort_num-1 in the same level, replace the delete's node position
	_, err = session.Exec("update fafacms_content_node SET sort_num=sort_num-1 where sort_num > ? and user_id = ? and parent_node_id = ?", n.SortNum, n.UserId, n.ParentNodeId)
	if err != nil {
		session.Rollback()
		log.Errorf("DeleteNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	_, err = session.Where("id=?", n.Id).Delete(new(model.ContentNode))
	if err != nil {
		session.Rollback()
		log.Errorf("DeleteNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		log.Errorf("DeleteNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
}

func TakeNode(c *gin.Context) {
	resp := new(Resp)
	req := new(NodeInfoRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("TakeNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.ContentNode)).Where("1=1").And("user_id=?", uu.Id)

	isOne := false
	if req.Id != 0 {
		isOne = true
		session.And("id=?", req.Id)
	}

	if req.Seo != "" {
		isOne = true
		session.And("seo=?", req.Seo)
	}

	if !isOne {
		log.Errorf("Node err:%s", "id or seo empty")
		resp.Error = Error(ParasError, "id or seo empty")
		return
	}

	v := new(model.ContentNode)
	exist, err := session.Get(v)
	if err != nil {
		log.Errorf("Node err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("Node err:%s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	f := Node{}
	f.Id = v.Id
	f.Seo = v.Seo
	f.Describe = v.Describe
	f.ImagePath = v.ImagePath
	f.Name = v.Name
	if v.UpdateTime > 0 {
		f.UpdateTime = GetSecond2DateTimes(v.UpdateTime)
		f.UpdateTimeInt = v.UpdateTime
	}
	f.CreateTime = GetSecond2DateTimes(v.CreateTime)
	f.CreateTimeInt = v.CreateTime
	f.SortNum = v.SortNum
	f.UserName = v.UserName
	f.UserId = v.UserId
	f.Level = v.Level
	f.ParentNodeId = v.ParentNodeId
	f.Status = v.Status

	// is the root level and want list son
	if f.Level == 0 && req.ListSon {
		ns := make([]model.ContentNode, 0)
		err = model.FaFaRdb.Client.Where("parent_node_id=?", f.Id).Find(&ns)
		if err != nil {
			log.Errorf("Node err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		for _, vv := range ns {
			ff := Node{}
			ff.Id = vv.Id
			ff.Seo = vv.Seo
			ff.Describe = vv.Describe
			ff.ImagePath = vv.ImagePath
			ff.Name = vv.Name
			if vv.UpdateTime > 0 {
				ff.UpdateTime = GetSecond2DateTimes(vv.UpdateTime)
				ff.UpdateTimeInt = vv.UpdateTime
			}
			ff.CreateTime = GetSecond2DateTimes(vv.CreateTime)
			ff.CreateTimeInt = vv.CreateTime
			ff.SortNum = vv.SortNum
			ff.UserName = vv.UserName
			ff.UserId = vv.UserId
			ff.Level = vv.Level
			ff.ParentNodeId = vv.ParentNodeId
			ff.Status = vv.Status
			f.Son = append(f.Son, ff)
		}
	}
	resp.Flag = true
	resp.Data = f

}

func ListNode(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListNodeHelper(c, uid)
}

func ListNodeAdmin(c *gin.Context) {
	ListNodeHelper(c, 0)
}

func ListNodeHelper(c *gin.Context, userId int64) {
	resp := new(Resp)

	respResult := new(NodesResponse)
	req := new(NodesInfoRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if userId != 0 {
		req.UserId = userId
		req.UserName = ""
	}
	// 注意：userId==0 是 admin 路由（ListNodeAdmin），允许查全部用户节点（可空条件）

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.ContentNode)).Where("1=1")

	if req.UserId != 0 {
		session.And("user_id=?", req.UserId)
	}

	if req.UserName != "" {
		uid, uerr := userFilterId(req.UserName)
		if uerr != nil {
			log.Errorf("ListNodeHelper err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("user_id=?", uid)
	}

	nodes := make([]model.ContentNode, 0)
	Build(session, req.Sort, model.ContentNodeSortName)
	err := session.Find(&nodes)
	if err != nil {
		log.Errorf("ListNode err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	father := make([]model.ContentNode, 0)
	son := make([]model.ContentNode, 0)
	for _, v := range nodes {
		if v.Level == 0 {
			father = append(father, v)
		} else {
			son = append(son, v)
		}
	}

	n := make([]Node, 0)
	for _, v := range father {
		f := Node{}
		f.Id = v.Id
		f.Seo = v.Seo
		f.Describe = v.Describe
		f.ImagePath = v.ImagePath
		f.Name = v.Name
		if v.UpdateTime > 0 {
			f.UpdateTime = GetSecond2DateTimes(v.UpdateTime)
			f.UpdateTimeInt = v.UpdateTime
		}
		f.CreateTime = GetSecond2DateTimes(v.CreateTime)
		f.CreateTimeInt = v.CreateTime
		f.SortNum = v.SortNum
		f.UserName = v.UserName
		f.UserId = v.UserId
		f.Level = v.Level
		f.ParentNodeId = v.ParentNodeId
		f.Status = v.Status
		// 实时统计节点下内容数（管理后台用，保证准确）
		f.ContentNum = CountContentNumOfNode(v.UserId, v.Id, false)
		for _, vv := range son {
			if vv.ParentNodeId == f.Id {
				s := Node{}
				s.Id = vv.Id
				s.Seo = vv.Seo
				s.Describe = vv.Describe
				s.ImagePath = vv.ImagePath
				s.Name = vv.Name
				if vv.UpdateTime > 0 {
					s.UpdateTimeInt = vv.UpdateTime
					s.UpdateTime = GetSecond2DateTimes(vv.UpdateTime)
				}
				s.CreateTime = GetSecond2DateTimes(vv.CreateTime)
				s.CreateTimeInt = vv.CreateTime
				s.SortNum = vv.SortNum
				s.UserId = vv.UserId
				s.UserName = vv.UserName
				s.Level = vv.Level
				s.ParentNodeId = vv.ParentNodeId
				s.Status = vv.Status
				s.ContentNum = CountContentNumOfNode(vv.UserId, vv.Id, false)
				f.Son = append(f.Son, s)
			}
		}

		n = append(n, f)
	}

	respResult.Nodes = n
	resp.Flag = true
	resp.Data = respResult
}

// 实时统计某用户某节点下的内容数量。
// onlyNormal=true：公开口径，只算正常已发布（status=0 + version>0）。
// onlyNormal=false：管理口径，算非回收站（status!=3，含草稿/隐藏/违禁/已发布），与内容管理「全部」列表一致。
func CountContentNumOfNode(userId, nodeId int64, onlyNormal bool) int64 {
	if userId == 0 || nodeId == 0 {
		return 0
	}
	sess := model.FaFaRdb.Client.Table(new(model.Content)).
		Where("user_id=?", userId).And("node_id=?", nodeId)
	if onlyNormal {
		sess.And("status=?", 0).And("version>?", 0)
	} else {
		sess.And("status!=?", 3)
	}
	num, err := sess.Count()
	if err != nil {
		return 0
	}
	return num
}

// put x behind y
// when y is zero, x will be the top one.
type SortNodeRequest struct {
	XID int64 `json:"xid" validate:"required"`
	YID int64 `json:"yid"`
}

// Sort the node
// sort_num more small more forward
func SortNode(c *gin.Context) {
	resp := new(Resp)
	req := new(SortNodeRequest)
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
		log.Errorf("SortNode err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if req.XID == req.YID {
		log.Errorf("SortNode err: %s", "xid=yid not right")
		resp.Error = Error(ParasError, "xid=yid not right")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("SortNode err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	x := new(model.ContentNode)
	x.Id = req.XID
	x.UserId = uu.Id
	exist, err := x.GetSortOneNode()
	if err != nil {
		log.Errorf("SortNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("SortNode err: %s", "x node not found")
		resp.Error = Error(ContentNodeNotFound, "x node not found")
		return
	}

	// 收集同级节点（同 parent_node_id），按 sort_num, id 排序，重建连续 sort_num
	var all []model.ContentNode
	err = model.FaFaRdb.Client.Where("user_id=?", uu.Id).And("parent_node_id=?", x.ParentNodeId).Asc("sort_num").Asc("id").Find(&all)
	if err != nil {
		log.Errorf("SortNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 目标顺序：先移除 X，再插入到 Y 之后（YID=0 表示插到最前）
	order := make([]int64, 0, len(all))
	for _, nn := range all {
		if nn.Id != x.Id {
			order = append(order, nn.Id)
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
			log.Errorf("SortNode err: %s", "y node not in same level")
			resp.Error = Error(ContentNodeNotFound, "y node not found")
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
		log.Errorf("SortNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	for i, id := range newOrder {
		_, err = session.Exec("update fafacms_content_node set sort_num=? where id=? and user_id=?", i, id, uu.Id)
		if err != nil {
			session.Rollback()
			log.Errorf("SortNode err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}
	err = session.Commit()
	if err != nil {
		session.Rollback()
		log.Errorf("SortNode err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}
	resp.Flag = true
	return
}
