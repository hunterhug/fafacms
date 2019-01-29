package controllers

import (
	"convoice/core/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/flog"
	"github.com/hunterhug/fafacms/core/model"
	"strconv"
	"strings"
	"sync"
)

var (
	AuthResource = sync.Map{}
)

// url===>resource_id
func InitAuthResource() {
	r := make([]model.Resource, 0)
	err := config.FafaRdb.Client.Find(r)
	if err != nil {
		flog.Log.Errorf("InitAuthResource err:%ss", err.Error())
	}
	for _, v := range r {
		flog.Log.Debugf("InitAuthResource load %s=%s", v.Url, v.Id)
		AuthResource.Store(v.Url, v.Id)
	}
}

// filter
var AuthFilter = func(c *gin.Context) {
	resp := new(Resp)
	defer func() {
		if resp.Error == nil {
			return
		}
		c.AbortWithStatusJSON(403, resp)
	}()

	u, _ := GetUserSession(c)
	if u == nil {
		// check cookie
		success, userInfo, _ := CheckCookie(c)
		if success {
			err := SetUserSession(c, userInfo)
			if err != nil {
				flog.Log.Errorf("filter err:%s", err.Error())
				resp.Error = &ErrorResp{
					ErrorID:  AuthPermit,
					ErrorMsg: ErrorMap[AuthPermit],
				}
				return
			}
			u = userInfo
		} else {
			flog.Log.Errorf("filter err:%s", "no cookie")
			resp.Error = &ErrorResp{
				ErrorID:  AuthPermit,
				ErrorMsg: ErrorMap[AuthPermit],
			}
			return
		}
	}

	c.Set("uid", u.Id)

	v, ok := AuthResource.Load(c.Request.URL.Path)
	if !ok {
		return
	}

	resourceId, ok := v.(int)
	if !ok {
		return
	}

	nowUser := new(model.User)
	err := nowUser.Get(u.Id)
	if err != nil {
		flog.Log.Errorf("filter err:%s", err.Error())
		resp.Error = &ErrorResp{
			ErrorID:  AuthPermit,
			ErrorMsg: ErrorMap[AuthPermit],
		}
		return
	}
	group := new(model.Group)
	err = group.Get(nowUser.GroupId)
	if err != nil {
		flog.Log.Errorf("filter err:%s", err.Error())
		resp.Error = &ErrorResp{
			ErrorID:  AuthPermit,
			ErrorMsg: ErrorMap[AuthPermit],
		}
		return
	}

	gr := new(model.GroupResource)
	gr.GroupId = group.Id
	gr.ResourceId = resourceId
	exist, err := config.FafaRdb.Client.Exist(gr)
	if err != nil {
		flog.Log.Errorf("filter err:%s", err.Error())
		resp.Error = &ErrorResp{
			ErrorID:  AuthPermit,
			ErrorMsg: ErrorMap[AuthPermit],
		}
		return
	}

	if !exist {
		flog.Log.Errorf("filter err:%s", "resource not allow")
		resp.Error = &ErrorResp{
			ErrorID:  AuthPermit,
			ErrorMsg: ErrorMap[AuthPermit],
		}
		return
	}
}

func CheckCookie(c *gin.Context) (success bool, user *model.User, err error) {
	cookieString, err := c.Cookie("auth")
	if err != nil {
		return false, nil, err
	}
	arr := strings.Split(cookieString, "|")
	if len(arr) < 2 {
		err = errors.New("paras less than 2")
		return
	}

	var userId int64
	str, password := arr[0], arr[1]
	userId, err = strconv.ParseInt(str, 10, 0)
	if err != nil {
		return
	}

	user = &model.User{}
	err = user.Get(int(userId))
	if err != nil {
		return
	}

	if password == utils.Md5(c.ClientIP()+"|"+user.Password) {
		success = true
		return
	} else {
		c.SetCookie("auth", "", -1, "/", "", false, true)
		return
	}
}

func GetUserSession(c *gin.Context) (*model.User, error) {
	u := new(model.User)
	s := config.FafaSessionMgr.Load(c.Request)
	err := s.GetObject("user", u)
	if err != nil {
		return nil, err
	}

	if u.Id == 0 {
		return nil, errors.New("no session")
	}
	return u, err
}

func SetUserSession(c *gin.Context, user *model.User) error {
	s := config.FafaSessionMgr.Load(c.Request)
	err := s.PutObject(c.Writer, "user", user)
	return err
}

func DeleteUserSession(c *gin.Context) error {
	s := config.FafaSessionMgr.Load(c.Request)
	err := s.Destroy(c.Writer)
	return err
}
