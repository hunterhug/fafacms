package session

import (
	"errors"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	"github.com/hunterhug/gosession"
	"github.com/hunterhug/gosession/kv"
)

var (
	// redis key
	redisToken = "ff_tokens"
	redisUser  = "ff_users"
)

var Mgr gosession.TokenManage

func InitSession(redisConf config.MyRedisConf, singleLogin bool) error {
	pool, err := kv.NewRedis(&kv.MyRedisConf{
		RedisHost:        redisConf.RedisHost,
		RedisMaxIdle:     redisConf.RedisMaxIdle,
		RedisMaxActive:   redisConf.RedisMaxActive,
		RedisIdleTimeout: redisConf.RedisIdleTimeout,
		RedisDB:          redisConf.RedisDB,
		RedisPass:        redisConf.RedisPass,
		IsCluster:        false,
		MasterName:       "",
	})
	if err != nil {
		return err
	}

	Mgr, _ = gosession.NewRedisSessionWithPool(pool)
	Mgr.ConfigTokenKeyPrefix(redisToken)
	Mgr.ConfigUserKeyPrefix(redisUser)
	Mgr.ConfigGetUserInfoFunc(func(id string) (*gosession.User, error) {
		user := new(model.User)
		idInt, err := util.SI(id)
		if err != nil {
			return nil, err
		}

		if idInt == 0 {
			return nil, errors.New("user id wrong")
		}

		user.Id = int64(idInt)
		exist, err := user.GetRaw()
		if err != nil {
			return nil, err
		}

		if !exist {
			return nil, errors.New("user not exist")
		}

		user.Password = ""
		user.ActivateCode = ""
		user.ActivateCodeExpired = 0
		user.ResetCode = ""
		user.ResetCodeExpired = 0

		u := new(gosession.User)
		u.Id = id
		u.Detail = user
		return u, nil
	})

	if singleLogin {
		Mgr.SetSingleMode()
	}
	return nil
}
