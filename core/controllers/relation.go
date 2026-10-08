package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	log "github.com/hunterhug/golog"
	"math"
)

type AddRelationRequest struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}

func AddRelation(c *gin.Context) {
	resp := new(Resp)
	req := new(AddRelationRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.UserId == 0 && req.UserName == "" {
		log.Errorf("RelationAdd err: %s", "user info empty")
		resp.Error = Error(ParasError, "user info empty")
		return
	}

	who := new(model.User)
	who.Id = req.UserId
	who.Name = req.UserName
	ok, err := who.GetActivateRaw()
	if err != nil {
		log.Errorf("RelationAdd err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("RelationAdd err: %s", "user not fund")
		resp.Error = Error(UserNotFound, "")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("RelationAdd err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	r := new(model.Relation)
	r.UserAId = uu.Id
	r.UserBId = who.Id
	r.UserAName = uu.Name
	r.UserBName = who.Name
	err = r.Add()
	if err != nil {
		log.Errorf("RelationAdd err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type MinuteRelationRequest struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}

func MinuteRelation(c *gin.Context) {
	resp := new(Resp)
	req := new(MinuteRelationRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.UserId == 0 && req.UserName == "" {
		log.Errorf("RelationMinute err: %s", "user info empty")
		resp.Error = Error(ParasError, "user info empty")
		return
	}

	who := new(model.User)
	who.Id = req.UserId
	who.Name = req.UserName
	ok, err := who.GetActivateRaw()
	if err != nil {
		log.Errorf("RelationMinute err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("RelationMinute err: %s", "user not fund")
		resp.Error = Error(UserNotFound, "")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("RelationMinute err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	r := new(model.Relation)
	r.UserAId = uu.Id
	r.UserBId = who.Id
	err = r.Minute()
	if err != nil {
		log.Errorf("RelationMinute err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type ListRelationRequest struct {
	UserAId         int64    `json:"user_a_id"`
	UserBId         int64    `json:"user_b_id"`
	UserAName       string   `json:"user_a_name"`
	UserBName       string   `json:"user_b_name"`
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	Sort            []string `json:"sort"`
	PageHelp
}

type ListRelationResponse struct {
	Relations []model.Relation           `json:"relations"`
	Users     map[int64]model.UserHelper `json:"users"`
	PageHelp
}

// who follow you
func ListFollowedRelationOfMe(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListFollowedRelationOfMe err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListRelation(c, 0, uid, uid)
}

// you follow who
func ListFollowingRelationOfMe(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListFollowingRelationOfMe err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListRelation(c, uid, 0, uid)
}

// who follow B
func ListFollowedRelation(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListFollowedRelation err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListRelation(c, 0, -1, uid)
}

// A follow who
func ListFollowingRelation(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListFollowingRelation err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListRelation(c, -1, 0, uid)
}

// who follow who
func ListAllRelation(c *gin.Context) {
	ListRelation(c, 0, 0, 0)
}

func ListRelation(c *gin.Context, userAId int64, userBId int64, me int64) {
	resp := new(Resp)
	req := new(ListRelationRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Relation)).Where("1=1")

	if userAId == -1 {
		// A follow who, so A should not empty
		if req.UserAId == 0 && req.UserAName == "" {
			log.Errorf("ListRelation err: %s", "user A info empty")
			resp.Error = Error(ParasError, "user A info empty")
			return
		}
	}

	// who follow B, so B should not empty
	if userBId == -1 {
		if req.UserBId == 0 && req.UserBName == "" {
			log.Errorf("ListRelation err: %s", "user B info empty")
			resp.Error = Error(ParasError, "user B info empty")
			return
		}
	}

	// you follow who
	if userAId > 0 {
		req.UserAId = userAId
		req.UserAName = ""
	}

	// who follow you
	if userBId > 0 {
		req.UserBId = userBId
		req.UserBName = ""
	}

	if req.UserAId != 0 {
		session.And("user_a_id=?", req.UserAId)
	}

	if req.UserBId != 0 {
		session.And("user_b_id=?", req.UserBId)
	}

	if req.UserAName != "" {
		uid, uerr := userFilterId(req.UserAName)
		if uerr != nil {
			log.Errorf("ListRelationAdmin err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("user_a_id=?", uid)
	}

	if req.UserBName != "" {
		uid, uerr := userFilterId(req.UserBName)
		if uerr != nil {
			log.Errorf("ListRelationAdmin err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("user_b_id=?", uid)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	cs := make([]model.Relation, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.RelationSortName)

	// do query
	total, err := session.FindAndCount(&cs)
	if err != nil {
		log.Errorf("ListRelation err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// search the user info
	userIds := make(map[int64]struct{})

	// search other's relation
	if userAId == -1 || userBId == -1 {
		// A follow who
		if userAId == -1 {
			for k, v := range cs {
				// you follow who
				temp := new(model.Relation)
				temp.UserAId = me
				temp.UserBId = v.UserBId

				// those people should take info
				userIds[v.UserBId] = struct{}{}
				userIds[v.UserAId] = struct{}{}
				num, err := temp.Count()

				if err != nil {
					log.Errorf("ListRelation err:%s", err.Error())
					resp.Error = Error(DBError, err.Error())
					return
				}

				if num == 0 {
					continue
				}

				cs[k].IsFollowing = true

				// who follow you
				temp1 := new(model.Relation)
				temp1.UserAId = v.UserBId
				temp1.UserBId = me
				num, err = temp.Count()
				if err != nil {
					log.Errorf("ListRelation err:%s", err.Error())
					resp.Error = Error(DBError, err.Error())
					return
				}

				if num == 0 {
					continue
				}

				cs[k].IsBoth = true
			}
		}

		// who follow B
		if userBId == -1 {
			for k, v := range cs {
				// you follow who
				temp := new(model.Relation)
				temp.UserAId = me
				temp.UserBId = v.UserAId

				// those people should take info
				userIds[v.UserBId] = struct{}{}
				userIds[v.UserAId] = struct{}{}
				num, err := temp.Count()
				if err != nil {
					log.Errorf("ListRelation err:%s", err.Error())
					resp.Error = Error(DBError, err.Error())
					return
				}

				if num == 0 {
					continue
				}

				cs[k].IsFollowing = true

				// who follow you
				temp1 := new(model.Relation)
				temp1.UserAId = v.UserAId
				temp1.UserBId = me
				num, err = temp.Count()
				if err != nil {
					log.Errorf("ListRelation err:%s", err.Error())
					resp.Error = Error(DBError, err.Error())
					return
				}

				if num == 0 {
					continue
				}

				cs[k].IsBoth = true
			}
		}
	} else {
		// admin or me search relation
		for k, v := range cs {
			temp := new(model.Relation)
			temp.UserAId = v.UserBId
			temp.UserBId = v.UserAId

			// those people should take info
			userIds[v.UserBId] = struct{}{}
			userIds[v.UserAId] = struct{}{}
			num, err := temp.Count()
			if err != nil {
				log.Errorf("ListRelation err:%s", err.Error())
				resp.Error = Error(DBError, err.Error())
				return
			}

			if num > 0 {
				cs[k].IsBoth = true
			}
		}
	}

	users, err := model.GetUser(util.MapToArray(userIds))
	if err != nil {
		log.Errorf("ListRelation err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	respResult := new(ListRelationResponse)
	respResult.Relations = cs
	respResult.Users = users
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}
