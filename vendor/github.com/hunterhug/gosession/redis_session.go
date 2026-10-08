package gosession

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/hunterhug/gosession/kv"
	"strings"
	"time"
)

var (
	// default prefix of token key
	tokenKeyDefault = "gosession-token"
	// default prefix of user key
	userKeyDefault = "gosession-user"
	// default key expire second
	expireTimeDefault int64 = 3600 * 24 * 7

	// TokenMapKeyExpireTime list all token, hash key expire
	TokenMapKeyExpireTime int64 = 3600 * 24 * 30
)

var (
	// GetUserInfoFuncDefault default get user info func you can choose
	GetUserInfoFuncDefault GetUserInfoFunc = func(id string) (*User, error) { return &User{Id: id}, nil }
)

// GetUserInfoFunc func get user info from where
type GetUserInfoFunc func(id string) (*User, error)

// RedisSession session by redis
type RedisSession struct {
	pool         *redis.Pool                    // redis pool can single mode or other mode
	getUserFunc  func(id string) (*User, error) // when not hit cache will get user from this func
	tokenKey     string                         // prefix of token，default 'got'
	userKey      string                         // prefix of user info cache ，default 'gou'
	expireTime   int64                          // token expire how much second，default  7 days
	isSingleMode bool                           // is single token, new token will destroy other token
}

// NewRedisSession new a redis session with redisConf config
func NewRedisSession(redisConf *kv.MyRedisConf) (TokenManage, error) {
	if redisConf == nil {
		return nil, errors.New("config is nil")
	}

	pool, err := kv.NewRedis(redisConf)
	if err != nil {
		return nil, err
	}

	return NewRedisSessionWithPool(pool)
}

// NewRedisSessionSimple new a redis session with simple config
func NewRedisSessionSimple(redisHost string, redisDB int, redisPass string) (TokenManage, error) {
	redisConf := NewRedisSessionSingleModeConfig(redisHost, redisDB, redisPass)
	return NewRedisSession(redisConf)
}

// NewRedisSessionWithPool new a redis session by redis pool
func NewRedisSessionWithPool(pool *redis.Pool) (TokenManage, error) {
	if pool == nil {
		return nil, errors.New("redis pool is nil")
	}
	return &RedisSession{pool: pool, tokenKey: tokenKeyDefault, userKey: userKeyDefault, expireTime: expireTimeDefault}, nil
}

// NewRedisSessionAll new a redis session, config all
// define prefix of token and user key
func NewRedisSessionAll(redisConf *kv.MyRedisConf, tokenKey, userKey string, expireTime int64, getUserInfoFunc GetUserInfoFunc) (TokenManage, error) {
	if redisConf == nil {
		return nil, errors.New("config is nil")
	}

	pool, err := kv.NewRedis(redisConf)
	if err != nil {
		return nil, err
	}

	tokenKey = strings.Replace(tokenKey, "_", "-", -1)
	userKey = strings.Replace(userKey, "_", "-", -1)

	if expireTime <= 0 {
		expireTime = expireTimeDefault
	}
	return &RedisSession{pool: pool, tokenKey: tokenKey, userKey: userKey, expireTime: expireTime, getUserFunc: getUserInfoFunc}, nil
}

// NewRedisSessionSingleModeConfig redis single mode config
func NewRedisSessionSingleModeConfig(redisHost string, redisDB int, redisPass string) *kv.MyRedisConf {
	return &kv.MyRedisConf{
		RedisHost:        redisHost,
		RedisPass:        redisPass,
		RedisDB:          redisDB,
		RedisIdleTimeout: 15,
		RedisMaxActive:   20,
		RedisMaxIdle:     30,
	}
}

// NewRedisSessionSentinelModeConfig redis sentinel mode config
// redisHost is sentinel address, not redis address
// redisPass is redis password
func NewRedisSessionSentinelModeConfig(redisHost string, redisDB int, redisPass string, masterName string) *kv.MyRedisConf {
	return &kv.MyRedisConf{
		RedisHost:        redisHost,
		RedisDB:          redisDB,
		RedisIdleTimeout: 15,
		RedisMaxActive:   20,
		RedisMaxIdle:     30,
		IsCluster:        true,
		MasterName:       masterName,
		RedisPass:        redisPass,
	}
}

// ConfigTokenKeyPrefix config by chain
func (s *RedisSession) ConfigTokenKeyPrefix(tokenKey string) TokenManage {
	tokenKey = strings.Replace(tokenKey, "_", "-", -1)
	s.tokenKey = tokenKey
	return s
}

// ConfigUserKeyPrefix config by chain
func (s *RedisSession) ConfigUserKeyPrefix(userKey string) TokenManage {
	userKey = strings.Replace(userKey, "_", "-", -1)
	s.userKey = userKey
	return s
}

// ConfigDefaultExpireTime config by chain
func (s *RedisSession) ConfigDefaultExpireTime(second int64) TokenManage {
	if second <= 0 {
		second = expireTimeDefault
	}
	s.expireTime = second
	return s
}

// ConfigGetUserInfoFunc config by chain
func (s *RedisSession) ConfigGetUserInfoFunc(fn GetUserInfoFunc) TokenManage {
	s.getUserFunc = fn
	return s
}

// SetSingleMode set single mode, new token will destroy other token
func (s *RedisSession) SetSingleMode() TokenManage {
	s.isSingleMode = true
	return s
}

// SetToken Set token, expire after some second
func (s *RedisSession) SetToken(useId string, tokenValidTimes int64) (token string, err error) {
	// user id can not nil
	if useId == "" {
		err = errors.New("user id nil")
		return
	}

	if tokenValidTimes <= 0 {
		tokenValidTimes = s.expireTime
	}

	// gen token by user id, everytime will gen new
	token = s.genToken(useId)

	// gen user key by user id
	userKey := s.hashUserKey(useId)

	// if single, destroy other token first
	if s.isSingleMode {
		err = s.DeleteUserToken(useId)
		if err != nil {
			return "", err
		}
	} else {
		// clear Token
		go func() {
			err := s.clearToken(useId)
			if err != nil {
			}
		}()
	}

	// relate token and user in redis
	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return "", err
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	err = conn.Send("MULTI")
	if err != nil {
		return "", err
	}

	err = conn.Send("SETEX", s.hashTokenKey(token), tokenValidTimes, []byte(userKey))
	if err != nil {
		return "", err
	}

	tokenMapKey := s.userTokenMapKey(useId)

	err = conn.Send("HSET", tokenMapKey, token, time.Now().Unix()+tokenValidTimes)
	if err != nil {
		return "", err
	}

	err = conn.Send("EXPIRE", tokenMapKey, TokenMapKeyExpireTime)
	if err != nil {
		return "", err
	}

	_, err = conn.Do("EXEC")
	return token, nil
}

// RefreshToken Refresh token，token expire will be again after some second
func (s *RedisSession) RefreshToken(token string, tokenValidTimes int64) (err error) {
	if token == "" {
		err = errors.New("token empty")
		return
	}

	temp := strings.Split(token, "_")
	if len(temp) < 2 || temp[0] == "" {
		err = errors.New("token wrong")
		return
	}

	if tokenValidTimes <= 0 {
		tokenValidTimes = s.expireTime
	}

	userId := temp[0]

	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	err = conn.Send("MULTI")
	if err != nil {
		return
	}

	userKey := s.hashUserKey(userId)
	err = conn.Send("SETEX", s.hashTokenKey(token), tokenValidTimes, []byte(userKey))
	if err != nil {
		return
	}

	tokenMapKey := s.userTokenMapKey(userId)
	err = conn.Send("HSET", tokenMapKey, token, time.Now().Unix()+tokenValidTimes)
	if err != nil {
		return
	}

	err = conn.Send("EXPIRE", tokenMapKey, TokenMapKeyExpireTime)
	if err != nil {
		return
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		return
	}

	return
}

// DeleteToken Delete token when you do action such logout
func (s *RedisSession) DeleteToken(token string) (err error) {
	if token == "" {
		err = errors.New("token empty")
		return
	}

	temp := strings.Split(token, "_")
	if len(temp) < 2 || temp[0] == "" {
		err = errors.New("token wrong")
		return
	}

	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	userId := temp[0]
	return s.deleteToken(conn, userId, token)
}

func (s *RedisSession) deleteToken(conn redis.Conn, userId string, token string) (err error) {
	if token == "" {
		err = errors.New("token empty")
		return
	}

	if userId == "" {
		err = errors.New("user id empty")
		return
	}

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}

	err = conn.Send("DEL", s.hashTokenKey(token))
	if err != nil {
		return err
	}

	err = conn.Send("HDEL", s.userTokenMapKey(userId), token)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	return err
}

// CheckTokenOrUpdateUser Check the token, when cache database exist return user info directly,
// others hit the persistent database and save newest user in cache database then return. such redis check, not check load from mysql.
// you can check user info by that token, if s.getUserFunc == nil do nothing
func (s *RedisSession) CheckTokenOrUpdateUser(token string, userInfoValidTimes int64) (user *User, exist bool, err error) {
	if token == "" {
		err = errors.New("token empty")
		return
	}

	temp := strings.Split(token, "_")
	if len(temp) < 2 || temp[0] == "" {
		err = errors.New("token wrong")
		return
	}

	userId := temp[0]

	// get user key
	value, ttl, exist, err := s.get(s.hashTokenKey(token))
	if err != nil {
		return nil, false, err
	}

	tokenMapKey := s.userTokenMapKey(userId)

	if !exist || ttl <= 1 {
		err = s.deleteMap(tokenMapKey, token)
		if err != nil {
			return nil, false, err
		}
		return nil, false, nil
	}

	// get user id from user key
	userKey := string(value)
	temp = strings.Split(userKey, "_")
	if len(temp) != 2 || temp[0] != s.userKey || temp[1] != userId {
		return nil, false, errors.New("user key invalid")
	}

	expireTime, exist, err := s.hGet(tokenMapKey, token)
	if err != nil {
		return nil, false, err
	}

	if !exist {
		return nil, false, nil
	}

	if s.getUserFunc == nil || userInfoValidTimes < 0 {
		user = new(User)
		user.Id = userId
		user.TokenRemainLiveTime = ttl
		user.Token = token
		user.TokenExpireTime = expireTime
		return user, true, nil
	}

	// get user info by user key
	value, _, exist, err = s.get(userKey)
	if err != nil {
		return nil, false, err
	}

	// when exit user info return directly
	user = new(User)
	if exist {
		err = json.Unmarshal(value, user)
		if err != nil {
			return nil, false, err
		}
		user.Id = userId
		user.TokenRemainLiveTime = ttl
		user.Token = token
		user.TokenExpireTime = expireTime
		return user, true, nil
	}

	// load user and add into cache
	user, exist, err = s.AddUser(userId, userInfoValidTimes)
	if err != nil {
		return nil, false, err
	}

	if !exist {
		return nil, false, nil
	}

	user.TokenRemainLiveTime = ttl
	user.Token = token
	user.TokenExpireTime = expireTime
	return user, true, nil
}

func (s *RedisSession) CheckToken(token string) (user *User, exist bool, err error) {
	return s.CheckTokenOrUpdateUser(token, -1)
}

// AddUser Add the user info to cache，expire after some second
func (s *RedisSession) AddUser(userId string, userInfoValidTimes int64) (user *User, exist bool, err error) {
	if s.getUserFunc == nil {
		return nil, false, errors.New("getUserFunc nil")
	}

	if userId == "" {
		err = errors.New("user id empty")
		return
	}

	// get user info from outer func
	user, err = s.getUserFunc(userId)
	if err != nil {
		return nil, false, err
	}

	if user == nil {
		user = new(User)
		user.Id = userId
	}

	user.Id = userId

	// gen user key
	userKey := s.hashUserKey(user.Id)

	// get user info raw
	raw, err := json.Marshal(user)
	if err != nil {
		return nil, false, err
	}

	// set into redis
	err = s.set(userKey, raw, userInfoValidTimes)
	if err != nil {
		return nil, false, err
	}

	return user, true, nil
}

// RefreshUser Refresh cache of user info batch
func (s *RedisSession) RefreshUser(ids []string, userInfoValidTimes int64) (err error) {
	// very rude
	for _, id := range ids {
		_, _, err = s.AddUser(id, userInfoValidTimes)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteUserToken Delete all token of this user
func (s *RedisSession) DeleteUserToken(userId string) (err error) {
	if userId == "" {
		err = errors.New("user id empty")
		return
	}

	tokenMapKey := s.userTokenMapKey(userId)
	result, exist, err := s.getUserTokenMapKeys(tokenMapKey)
	if err != nil {
		return err
	}

	if !exist || len(result) == 0 {
		return nil
	}

	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}

	for _, v := range result {
		err = conn.Send("HDEL", tokenMapKey, v)
		if err != nil {
			return err
		}

		err = conn.Send("DEL", s.hashTokenKey(v))
		if err != nil {
			return err
		}
	}

	_, err = conn.Do("EXEC")
	return
}

// ListUserToken List all token in one user
func (s *RedisSession) ListUserToken(userId string) ([]string, error) {
	if userId == "" {
		err := errors.New("user id empty")
		return nil, err
	}

	result, _, err := s.getUserTokenMapKeys(s.userTokenMapKey(userId))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *RedisSession) clearToken(userId string) error {
	_, err := s.ListUserToken(userId)
	return err
}

// DeleteUser Delete user info in cache
func (s *RedisSession) DeleteUser(userId string) (err error) {
	if userId == "" {
		err = errors.New("user id empty")
		return
	}
	return s.delete(s.hashUserKey(userId))
}

// help func to set redis key which use MULTI order
func (s *RedisSession) set(key string, value []byte, expireSecond int64) (err error) {
	// when expireSecond not large 0 will use default second
	if expireSecond <= 0 {
		expireSecond = s.expireTime
	}

	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	_, err = conn.Do("SETEX", key, expireSecond, value)
	if err != nil {
		return err
	}

	return
}

// help func to delete redis key
func (s *RedisSession) delete(key string) (err error) {
	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	_, err = conn.Do("DEL", key)
	return err
}

func (s *RedisSession) deleteMap(key, subKey string) (err error) {
	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	_, err = conn.Do("HDEL", key, subKey)
	return err
}

func (s *RedisSession) deleteMapWithConn(conn redis.Conn, key, subKey string) (err error) {
	_, err = conn.Do("HDEL", key, subKey)
	return err
}

func (s *RedisSession) getUserTokenMapKeys(mapKey string) (result []string, exist bool, err error) {
	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	keys, err := redis.StringMap(conn.Do("HGETALL", mapKey))
	if err == redis.ErrNil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	result = make([]string, 0, len(keys))
	for k, v := range keys {
		if SI(v) <= time.Now().Unix() {
			err = s.deleteMapWithConn(conn, mapKey, k)
			if err != nil {
				return nil, false, err
			}
			continue
		}

		result = append(result, k)
	}
	return result, true, nil
}

// help func to get redis key
func (s *RedisSession) get(key string) (value []byte, ttl int64, exist bool, err error) {
	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	value, err = redis.Bytes(conn.Do("GET", key))
	if err == redis.ErrNil {
		return nil, 0, false, nil
	} else if err != nil {
		return nil, 0, false, err
	}

	ttl, err = redis.Int64(conn.Do("TTL", key))
	if err == redis.ErrNil {
		return nil, 0, false, nil
	} else if err != nil {
		return nil, 0, false, err
	}

	return value, ttl, true, nil
}

// help func to hGet redis key
func (s *RedisSession) hGet(key, subKey string) (value int64, exist bool, err error) {
	conn := s.pool.Get()
	if conn.Err() != nil {
		err = conn.Err()
		return
	}

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	value, err = redis.Int64(conn.Do("HGET", key, subKey))
	if err == redis.ErrNil {
		return 0, false, nil
	} else if err != nil {
		return 0, false, err
	}

	return value, true, nil
}

// gen token, will random gen string
func (s *RedisSession) genToken(userId string) string {
	// has prefix user id
	return fmt.Sprintf("%s_%s", userId, GetGUID())
}

// gen hashTokenKey, as a key in redis, it's value will be hashUserKey
func (s *RedisSession) hashTokenKey(token string) string {
	return fmt.Sprintf("%s_%s", s.tokenKey, token)
}

// gen hashUserKey, as a key in redis, it's value will be user info
func (s *RedisSession) hashUserKey(userId string) string {
	return fmt.Sprintf("%s_%s", s.userKey, userId)
}

// hash map key which struct store all token
func (s *RedisSession) userTokenMapKey(id string) string {
	return fmt.Sprintf("%s_%s", s.tokenKey, id)
}
