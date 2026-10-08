package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
)

var (
	BadTime int64 = 0
	AutoBan       = false
)

type CoolContentRequest struct {
	ContentId int64 `json:"id"`
}

func CoolContent(c *gin.Context) {
	resp := new(Resp)
	req := new(CoolContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.ContentId == 0 {
		log.Errorf("CoolContent err: %s", "content_id empty")
		resp.Error = Error(ParasError, "content_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("CoolContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	content := new(model.Content)
	content.Id = req.ContentId
	ok, err := content.GetByRaw()
	if err != nil {
		log.Errorf("CoolContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("CoolContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.Status != 0 {
		log.Errorf("CoolContent err: %s", "content status not 0 or not publish")
		if content.Status == 2 {
			resp.Error = Error(ContentBanPermit, "")
		} else {
			resp.Error = Error(ContentNotFound, "")
		}
		return
	}

	cool := new(model.ContentCool)
	cool.ContentId = req.ContentId
	cool.UserId = uu.Id
	ok, err = cool.Exist()
	if err != nil {
		log.Errorf("CoolContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

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
			go model.GoodContent(uu.Id, content.UserId, content.Id, content.Title)
		}
	}

	resp.Flag = true
	if ok {
		resp.Data = "-"
	} else {
		resp.Data = "+"
	}

	go SendToLoop(content.UserId, 0, 3)
	return
}

type BadContentRequest struct {
	ContentId int64  `json:"id"`
	Reason    string `json:"reason"`
}

func BadContent(c *gin.Context) {
	resp := new(Resp)
	req := new(BadContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.ContentId == 0 {
		log.Errorf("BadContent err: %s", "content_id empty")
		resp.Error = Error(ParasError, "content_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	content := new(model.Content)
	content.Id = req.ContentId
	ok, err := content.GetByRaw()
	if err != nil {
		log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		log.Errorf("BadContent err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.Status != 0 || content.Version == 0 {
		log.Errorf("BadContent err: %s", "content status not 0 or not publish")
		if content.Status == 2 {
			resp.Error = Error(ContentBanPermit, "")
		} else {
			resp.Error = Error(ContentNotFound, "")
		}
		return
	}

	bad := new(model.ContentBad)
	bad.ContentId = req.ContentId
	bad.UserId = uu.Id
	bad.Reason = req.Reason
	ok, err = bad.Exist()
	if err != nil {
		log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// 24 小时内已举报过：拒绝重复举报（叠加举报记录，不取消）
	if ok {
		log.Errorf("BadContent err: %s", "already bad in 24h")
		resp.Error = Error(AlreadyBad, "")
		return
	}

	err = bad.Create()
	if err != nil {
		log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	cc := new(model.Content)
	cc.Id = req.ContentId
	cc.UserId = content.UserId
	cc.Title = content.Title
	resp.Flag = true
	if AutoBan {
		err = cc.Ban(BadTime)
		if err != nil {
			log.Errorf("BadContent ban err: %s", err.Error())
		}
	}
	resp.Data = "+"
	return
}
