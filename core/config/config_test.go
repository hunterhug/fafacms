package config

import (
	"fmt"
	"github.com/hunterhug/fafacms/core/util/rdb"
	"testing"
)

// gen a config json
func TestJsonOutConfig(t *testing.T) {
	c := Config{}

	c.DefaultConfig.StoragePath = "/root/data"
	c.DefaultConfig.WebPort = "8080"
	c.DefaultConfig.LogDebug = true

	c.DbConfig.Host = "127.0.0.1"
	c.DbConfig.User = "root"
	c.DbConfig.Port = "3306"
	c.DbConfig.Prefix = "fafa_"
	c.DbConfig.Name = "blog"
	c.DbConfig.Pass = "123456789"
	c.DbConfig.Debug = true
	c.DbConfig.DebugToFile = true
	c.DbConfig.DebugToFileName = "/root/data/fafacms.log"
	c.DbConfig.MaxIdleConns = 20
	c.DbConfig.MaxOpenConns = 20
	c.DbConfig.DriverName = rdb.MYSQL

	c.SessionConfig.RedisHost = "127.0.0.1:6379"
	c.SessionConfig.RedisIdleTimeout = 120
	c.SessionConfig.RedisMaxActive = 0
	c.SessionConfig.RedisMaxIdle = 64
	c.SessionConfig.RedisDB = 0
	j, err := JsonOutConfig(c)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(j)
}
