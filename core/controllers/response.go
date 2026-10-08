package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	log "github.com/hunterhug/golog"
	"io/ioutil"
	"runtime"
	"strings"
	"time"
)

// ParseJSON Parse the json into request struct
func ParseJSON(c *gin.Context, req interface{}) *ErrorResp {
	pc, _, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	requestBody, _ := ioutil.ReadAll(c.Request.Body)

	ip := c.ClientIP()

	//log.Debugf("%s ParseJSON [%v,line:%v]:%s", ip, f.Name(), line, string(requestBody))
	if err := json.Unmarshal(requestBody, req); err != nil {
		log.Debugf("%s ParseJSONErr [%v,line:%v]:%s", ip, f.Name(), line, err.Error())

		// if parse wrong will not record log
		c.Set("skipLog", true)
		return Error(ParseJsonError, err.Error())
	}
	return nil
}

// JSONL Log the json output
func JSONL(c *gin.Context, code int, req interface{}, obj *Resp) {
	if c.GetBool("skipLog") {
		c.Render(code, render.JSON{Data: obj})
		return
	}

	// log will record
	record := new(model.Log)
	record.Ip = c.ClientIP()
	record.Url = c.Request.URL.Path
	record.LogTime = time.Now().Unix()
	record.Ua = c.Request.UserAgent()
	record.UserId = c.GetInt("uid")
	flag := obj.Flag
	if !flag && obj.Error != nil {
		errStr := obj.Error.Error()
		errStrSplit := strings.Split(errStr, "|")
		if len(errStrSplit) >= 2 {
			record.ErrorId = errStrSplit[0]
			record.ErrorMessage = strings.Join(errStrSplit[1:], "|")
		}
	}
	record.Flag = flag

	// 安全说明：不再把请求体/响应体写入日志。
	// 它们可能包含明文密码、token、正文等敏感内容（历史上会经 Debugf 打进日志文件）。
	// record.In / record.Out 字段保留以兼容表结构，但不再赋值。

	cid := util.GetGUID()
	record.Cid = cid

	// 只输出非敏感摘要：不含请求体、响应体、UA
	log.Debugf("FaFa Monitor: cid=%s ip=%s uid=%d flag=%v url=%s errId=%s errMsg=%s",
		record.Cid, record.Ip, record.UserId, record.Flag, record.Url, record.ErrorId, record.ErrorMessage)

	// log table not read fot it will slow the service
	//_, err := model.FaFaRdb.InsertOne(record)
	//if err != nil {
	//	log.Errorf("insert log record:%s", err.Error())
	//}

	obj.Cid = cid
	c.Render(code, render.JSON{Data: obj})
}

// JSON Just render the json
func JSON(c *gin.Context, code int, obj *Resp) {
	c.Render(code, render.JSON{Data: obj})
}
