package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/flog"
	"github.com/hunterhug/fafacms/core/model"
)

var BadTime int64 = 0

type CoolContentRequest struct {
	ContentId int64 `json:"content_id"`
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
		flog.Log.Errorf("CoolContent err: %s", "content_id empty")
		resp.Error = Error(ParasError, "content_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("CoolContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	cool := new(model.ContentCool)
	cool.ContentId = req.ContentId
	cool.UserId = uu.Id
	ok, err := cool.Exist()
	if err != nil {
		flog.Log.Errorf("CoolContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
	}

	if ok {
		err = cool.Delete()
	} else {
		err = cool.Create()
	}

	if err != nil {
		flog.Log.Errorf("CoolContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
	}

	resp.Flag = true
	if ok {
		resp.Data = "-"
	} else {
		resp.Data = "+"
	}
	return
}

type BadContentRequest struct {
	ContentId int64 `json:"content_id"`
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
		flog.Log.Errorf("BadContent err: %s", "content_id empty")
		resp.Error = Error(ParasError, "content_id empty")
		return
	}

	uu, err := GetUserSession(c)
	if err != nil {
		flog.Log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	bad := new(model.ContentBad)
	bad.ContentId = req.ContentId
	bad.UserId = uu.Id
	ok, err := bad.Exist()
	if err != nil {
		flog.Log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
	}

	if ok {
		err = bad.Delete()
	} else {
		err = bad.Create()
	}

	if err != nil {
		flog.Log.Errorf("BadContent err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
	}

	cc := new(model.Content)
	cc.Id = req.ContentId

	resp.Flag = true
	if ok {
		resp.Data = "-"
	} else {
		err = cc.Ban(BadTime)
		if err != nil {
			flog.Log.Errorf("BadContent ban err: %s", err.Error())
		}

		resp.Data = "+"
	}
	return
}