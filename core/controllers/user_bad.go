package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
	"math"
)

// ================= 举报用户 =================

type BadUserRequest struct {
	Id     int64  `json:"id" validate:"required"` // 被举报用户 id
	Reason string `json:"reason"`
}

func BadUser(c *gin.Context) {
	resp := new(Resp)
	req := new(BadUserRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("BadUser err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("BadUser err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	if req.Id == uu.Id {
		log.Errorf("BadUser err: %s", "can not bad self")
		resp.Error = Error(ParasError, "can not report yourself")
		return
	}

	target := new(model.User)
	target.Id = req.Id
	ok, err := target.GetRaw()
	if err != nil {
		log.Errorf("BadUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("BadUser err: %s", "user not found")
		resp.Error = Error(UserNotFound, "")
		return
	}

	// 不能举报超管（admin）
	if target.Name == "admin" {
		log.Errorf("BadUser err: %s", "can not bad admin")
		resp.Error = Error(ParasError, "can not report admin")
		return
	}

	bad := new(model.UserBad)
	bad.UserId = uu.Id
	bad.BadUserId = req.Id
	bad.Reason = req.Reason
	ok, err = bad.Exist()
	if err != nil {
		log.Errorf("BadUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 24 小时内已举报过：拒绝重复举报
	if ok {
		log.Errorf("BadUser err: %s", "already bad in 24h")
		resp.Error = Error(AlreadyBad, "")
		return
	}

	err = bad.Create()
	if err != nil {
		log.Errorf("BadUser err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Data = "+"
	resp.Flag = true
}

// ================= 举报用户列表（管理员） =================

type ListUserBadAdminRequest struct {
	Status int      `json:"status" validate:"oneof=-1 0 1 2"` // 被举报用户当前状态：-1 全部
	Sort   []string `json:"sort"`
	PageHelp
}

type UserBadItem struct {
	Id            int64  `json:"id"`
	UserId        int64  `json:"user_id"` // 举报人
	UserName      string `json:"user_name"`
	UserNick      string `json:"user_nick"`
	BadUserId     int64  `json:"bad_user_id"` // 被举报用户
	BadUserName   string `json:"bad_user_name"`
	BadUserNick   string `json:"bad_user_nick"`
	BadUserStatus int    `json:"bad_user_status"` // 0 未激活 1 正常 2 拉黑
	Reason        string `json:"reason"`
	CreateTime    int64  `json:"create_time"`
}

type ListUserBadAdminResponse struct {
	Bad []UserBadItem `json:"bad"`
	PageHelp
}

func ListUserBadAdmin(c *gin.Context) {
	resp := new(Resp)
	respResult := new(ListUserBadAdminResponse)
	req := new(ListUserBadAdminRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	if err := validate.Struct(req); err != nil {
		log.Errorf("ListUserBadAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()
	session.Table(new(model.UserBad)).Where("1=1")

	// 状态过滤：按被举报用户状态过滤
	if req.Status != -1 {
		var us []model.User
		if err := model.FaFaRdb.Client.Cols("id").Where("status=?", req.Status).Find(&us); err != nil {
			log.Errorf("ListUserBadAdmin err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if len(us) == 0 {
			respResult.Bad = []UserBadItem{}
			respResult.PageHelp = req.PageHelp
			resp.Data = respResult
			resp.Flag = true
			return
		}
		ids := make([]int64, 0, len(us))
		for _, v := range us {
			ids = append(ids, v.Id)
		}
		session.In("bad_user_id", ids)
	}

	cs := make([]model.UserBad, 0)
	p := &req.PageHelp
	p.build(session, req.Sort, []string{"=id"})

	total, err := session.FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListUserBadAdmin err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 收集举报人和被举报用户 id，一次性查用户信息
	userIds := make([]int64, 0, len(cs)*2)
	for _, v := range cs {
		userIds = append(userIds, v.UserId, v.BadUserId)
	}
	users := make(map[int64]model.User)
	if len(userIds) > 0 {
		us := make([]model.User, 0)
		if err := model.FaFaRdb.Client.In("id", userIds).Find(&us); err != nil {
			log.Errorf("ListUserBadAdmin err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		for _, u := range us {
			users[u.Id] = u
		}
	}

	items := make([]UserBadItem, 0, len(cs))
	for _, v := range cs {
		item := UserBadItem{
			Id:         v.Id,
			UserId:     v.UserId,
			BadUserId:  v.BadUserId,
			Reason:     v.Reason,
			CreateTime: v.CreateTime,
		}
		if u, ok := users[v.UserId]; ok {
			item.UserName = u.Name
			item.UserNick = u.NickName
		}
		if u, ok := users[v.BadUserId]; ok {
			item.BadUserName = u.Name
			item.BadUserNick = u.NickName
			item.BadUserStatus = u.Status
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
