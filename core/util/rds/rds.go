package rds

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// Redis 连接池与基础命令封装，供多副本共享的安全状态使用（限流/签名/风控/2FA/验证码）。

var pool *redis.Pool

// Init 初始化 Redis 连接池（复用 SessionConfig 的 Redis 配置）。
func Init(host string, db int, pass string, maxIdle, maxActive, idleTimeout int) error {
	pool = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			opts := []redis.DialOption{redis.DialDatabase(db)}
			if pass != "" {
				opts = append(opts, redis.DialPassword(pass))
			}
			return redis.Dial("tcp", host, opts...)
		},
	}
	c := pool.Get()
	defer c.Close()
	_, err := c.Do("PING")
	return err
}

func conn() redis.Conn {
	return pool.Get()
}

// SetEx 写 key，TTL 秒。
func SetEx(key, val string, ttlSec int) error {
	c := conn()
	defer c.Close()
	_, err := c.Do("SET", key, val, "EX", ttlSec)
	return err
}

// SetNxEx 仅当 key 不存在时写入并设 TTL，返回是否写入成功（用于 nonce 一次性）。
func SetNxEx(key, val string, ttlSec int) (bool, error) {
	c := conn()
	defer c.Close()
	res, err := redis.String(c.Do("SET", key, val, "EX", ttlSec, "NX"))
	if err == redis.ErrNil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return res == "OK", nil
}

// Get 读 key，key 不存在返回空串。
func Get(key string) (string, error) {
	c := conn()
	defer c.Close()
	s, err := redis.String(c.Do("GET", key))
	if err == redis.ErrNil {
		return "", nil
	}
	return s, err
}

// Del 删除若干 key。
func Del(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	c := conn()
	defer c.Close()
	args := make([]interface{}, len(keys))
	for i, k := range keys {
		args[i] = k
	}
	_, err := c.Do("DEL", args...)
	return err
}

// Incr 自增 key（不设 TTL，调用方自行 EXPIRE）。
func Incr(key string) (int64, error) {
	c := conn()
	defer c.Close()
	return redis.Int64(c.Do("INCR", key))
}

// incrWindowScript 固定窗口计数：首次 INCR 时设置 EXPIRE，之后复用窗口。
var incrWindowScript = redis.NewScript(1, `
local c = redis.call('INCR', KEYS[1])
if c == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return c
`)

// IncrWindow 固定窗口自增并返回计数（首次设置 TTL 秒）。
func IncrWindow(key string, ttlSec int) (int64, error) {
	c := conn()
	defer c.Close()
	return redis.Int64(incrWindowScript.Do(c, key, ttlSec))
}
