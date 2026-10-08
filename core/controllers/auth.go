package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/session"
	log "github.com/hunterhug/golog"
)

// AuthHeader every API which need auth should take a HTTP header `Auth`
const AuthHeader = "Auth"

const innerContextAuthCacheKey = "everAuth"

var (
	// AuthDebug if you want skip auth you can set it true
	AuthDebug = false

	// AdminUrl those api will be checked resource
	AdminUrl map[string]int64

	// SessionExpireTime redis key expire time
	SessionExpireTime int64 = 24 * 3600 * 7
)

// AuthFilter api access auth filter
var AuthFilter = func(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		if resp.Error == nil {
			return
		}
		c.AbortWithStatusJSON(200, resp)
	}()

	// get session
	nowUser, err := GetUserSession(c)
	if err != nil {
		log.Errorf("filter err:%s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	// record log will need uid, monitor who op
	c.Set("uid", nowUser.Id)

	// skip next auth
	if AuthDebug {
		return
	}

	// root user can ignore next auth
	if nowUser.Id == -1 {
		return
	}

	// admin user is skr
	if nowUser.Name == "admin" {
		return
	}

	// not active will be refuse
	if nowUser.Status == 0 {
		log.Errorf("filter err: not active")
		resp.Error = Error(UserNotActivate, "not active")
		return
	}

	// black user will be refuse
	if nowUser.Status == 2 {
		log.Errorf("filter err: black lock, contact admin")
		resp.Error = Error(UserIsInBlack, "black lock, contact admin")
		return
	}

	// resource is existed
	url := c.Request.URL.Path

	// resource not found can skip auth
	resourceId, exist := AdminUrl[url]
	if !exist {
		return
	}

	// if group has this resource
	gr := new(model.GroupResource)
	gr.GroupId = nowUser.GroupId
	gr.ResourceId = resourceId
	exist, err = model.FaFaRdb.Client.Exist(gr)
	if err != nil {
		log.Errorf("filter err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// resource not found in group will be refuse
	if !exist {
		log.Errorf("filter err:%s", "resource not allow")
		resp.Error = Error(UserAuthPermit, "resource not allow")
		return
	}
}

// GetUserSession get the info of user，will save in redis Session
func GetUserSession(c *gin.Context) (*model.User, error) {
	// get the info from context if exist
	if v, exist := c.Get(innerContextAuthCacheKey); exist {
		return v.(*model.User), nil
	}

	// get token from HTTP header and check if it is existed
	token := c.GetHeader(AuthHeader)
	user, exist, err := session.Mgr.CheckTokenOrUpdateUser(token, SessionExpireTime)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errors.New("user not found")
	}

	if user.Detail == nil {
		return nil, errors.New("user not found in cache")
	}

	u, ok := user.Detail.(*model.User)
	if !ok {
		// user.Detail 从 Redis 读回时是 map[string]interface{}（gosession 使用 JSON 序列化缓存用户信息）
		raw, err := json.Marshal(user.Detail)
		if err != nil {
			return nil, err
		}
		u = new(model.User)
		if err := json.Unmarshal(raw, u); err != nil {
			return nil, err
		}
	}

	// set the info into context
	c.Set(innerContextAuthCacheKey, u)
	return u, nil
}

func SetUserSession(id int64) (string, error) {
	if id == 0 {
		return "", errors.New("user nil")
	}

	idS := fmt.Sprintf("%d", id)
	err := session.Mgr.RefreshUser([]string{idS}, SessionExpireTime)
	if err != nil {
		return "", err
	}
	return session.Mgr.SetToken(idS, SessionExpireTime)
}

func DeleteUserSession(c *gin.Context) error {
	token := c.GetHeader(AuthHeader)
	err := session.Mgr.DeleteToken(token)
	return err
}

func DeleteUserAllSession(id int64) error {
	err := session.Mgr.DeleteUserToken(fmt.Sprintf("%d", id))
	return err
}

func RefreshUserSession(c *gin.Context) error {
	token := c.GetHeader(AuthHeader)
	err := session.Mgr.RefreshToken(token, SessionExpireTime)
	return err
}
