# 分布式 Session Golang库

[![GitHub forks](https://img.shields.io/github/forks/hunterhug/gosession.svg?style=social&label=Forks)](https://github.com/hunterhug/gosession/network)
[![GitHub stars](https://img.shields.io/github/stars/hunterhug/gosession.svg?style=social&label=Stars)](https://github.com/hunterhug/gosession/stargazers)
[![GitHub last commit](https://img.shields.io/github/last-commit/hunterhug/gosession.svg)](https://github.com/hunterhug/gosession)
[![Go Report Card](https://goreportcard.com/badge/github.com/hunterhug/gosession)](https://goreportcard.com/report/github.com/hunterhug/gosession)
[![GitHub issues](https://img.shields.io/github/issues/hunterhug/gosession.svg)](https://github.com/hunterhug/gosession/issues)

[English README](/README.md)

开源目的：工作中很多大小项目重复的需要登录注册功能，这是一个要复杂可复杂，要简单可简单的模块，借鉴了非常多的开源项目，于是把一些复用较高的，逻辑性不强，不涉及机密的代码抽离出轮子，回馈社区，感谢大家。

支持多个 `Web` 服务共享 `Session` 令牌 `token`，这样可以实现多个服务间共享状态。

现在 Session 令牌可以存储在：

1. 单机模式的 Redis。
2. 哨兵模式的 Redis。什么是哨兵，我们知道 Redis 有主从复制的功能，主服务器提供服务，从服务器作为数据同步来进行备份。当主服务器挂掉时，哨兵可以将从服务器提升到主角色。

## 如何使用

很简单，执行：

```
go get -v github.com/hunterhug/gosession
```

核心 API:

```go
// 分布式Session管理
// Token
type TokenManage interface {
	SetToken(id string, tokenValidTimes int64) (token string, err error)                               // 设置令牌，传入用户ID和令牌过期时间，单位秒，会生成一个Token返回，登录时可以调用
	RefreshToken(token string, tokenValidTimes int64) error                                            // 刷新令牌过期时间，令牌会继续存活
	DeleteToken(token string) error                                                                    // 删除令牌，退出登录时可以调用
	CheckToken(token string) (user *User, exist bool, err error)                                       // 检查令牌是否存在（检查会话是否存在）
	CheckTokenOrUpdateUser(token string, userInfoValidTimes int64) (user *User, exist bool, err error) // 检查令牌是否存在（检查会话是否存在），并缓存用户信息，如果有的话，默认不更新用户信息，可设置ConfigGetUserInfoFunc
	ListUserToken(id string) ([]string, error)                                                         // 列出用户的所有令牌
	DeleteUserToken(id string) error                                                                   // 删除用户的所有令牌
	RefreshUser(id []string, userInfoValidTimes int64) error                                           // 批量刷新用户信息，如果有的话，默认不缓存用户信息，可设置ConfigGetUserInfoFunc
	DeleteUser(id string) error                                                                        // 删除用户信息，默认不缓存用户信息，可设置ConfigGetUserInfoFunc
	AddUser(id string, userInfoValidTimes int64) (user *User, exist bool, err error)                   // 新增缓存用户信息，默认不缓存用户信息，可设置ConfigGetUserInfoFunc
	ConfigTokenKeyPrefix(tokenKey string) TokenManage                                                  // 设置令牌前缀
	ConfigUserKeyPrefix(userKey string) TokenManage                                                    // 设置用户信息前缀，默认不缓存用户信息，可设置ConfigGetUserInfoFunc
	ConfigDefaultExpireTime(second int64) TokenManage                                                  // 设置令牌默认过期时间
	ConfigGetUserInfoFunc(fn GetUserInfoFunc) TokenManage                                              // 设置获取用户信息的函数
	SetSingleMode() TokenManage                                                                        // 是否独占单点登录，新生成一个令牌，会挤掉其他令牌
}

// 用户信息，存token在缓存里，比如redis
// 如果有设置ConfigGetUserInfoFunc(fn GetUserInfoFunc)，那么同时也会缓存该用户信息，你可以在函数 type GetUserInfoFunc func(id string) (*User, error) 里将业务用户信息存入 Detail 并返回。
type User struct {
	Id                  string      `json:"id"`     // 用户标志，唯一
	TokenRemainLiveTime int64       `json:"-"`      // token还有多少秒就过期了
	Detail              interface{} `json:"detail"` // 可以存放用户业务信息
}
```

简单的例子：

```
package main

import (
	"fmt"
	"github.com/hunterhug/gosession"
)

func main() {
	tokenManage, err := gosession.NewRedisSessionSimple("127.0.0.1:6379", 0, "hunterhug")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	userId := "000001"
	token, err := tokenManage.SetToken(userId, 20)
	if err != nil {
		fmt.Println("set token err:", err.Error())
		return
	}
	fmt.Println("set token:", userId, token)

	user, exist, err := tokenManage.CheckToken(token)
	if err != nil {
		fmt.Println("check token err:", err.Error())
		return
	}

	if exist {
		fmt.Printf("check token exist: %#v\n", user)
	} else {
		fmt.Println("check token not exist")
	}

	err = tokenManage.DeleteToken(token)
	if err != nil {
		fmt.Println("delete token err:", err.Error())
		return
	} else {
		fmt.Println("delete token:", token)
	}

	user, exist, err = tokenManage.CheckToken(token)
	if err != nil {
		fmt.Println("after delete check token err:", err.Error())
		return
	}

	if exist {
		fmt.Printf("after delete check delete token exist: %#v\n", user)
	} else {
		fmt.Println("after delete check delete token not exist")
	}

	tokenManage.SetToken(userId, 20)
	tokenManage.SetToken(userId, 20)
	tokenManage.SetToken(userId, 20)
	tokenManage.SetToken(userId, 20)

	tokenList, err := tokenManage.ListUserToken(userId)
	if err != nil {
		fmt.Println("list token err:", err.Error())
		return
	}

	for _, v := range tokenList {
		fmt.Println("list token:", v)
	}

	err = tokenManage.DeleteUserToken(userId)
	if err != nil {
		fmt.Println("delete user all token err:", err.Error())
		return
	} else {
		fmt.Println("delete user all token")
	}

	tokenList, err = tokenManage.ListUserToken(userId)
	if err != nil {
		fmt.Println("after delete user all list token err:", err.Error())
		return
	}

	if len(tokenList) == 0 {
		fmt.Println("user token empty")
	}
}
```

另外一个带有解释的例子：

```go
package main

import (
	"fmt"
	"github.com/hunterhug/gosession"
	"time"
)

func main() {
	// 1. 配置Redis，目前支持单机和哨兵
	redisHost := "127.0.0.1:6379"
	redisDb := 0
	redisPass := "hunterhug" // Redis一般是没有密码的，可以留空
	redisConfig := gosession.NewRedisSessionSingleModeConfig(redisHost, redisDb, redisPass)
	// or
	//gosession.NewRedisSessionSentinelModeConfig()

	// 2. 连接Session管理器
	tokenManage, err := gosession.NewRedisSession(redisConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 3. 配置Session管理器，比如Token 600秒过期，以及token和用户信息key的前缀
	tokenManage.ConfigDefaultExpireTime(600)
	tokenManage.ConfigUserKeyPrefix("go-user")
	tokenManage.ConfigTokenKeyPrefix("go-token")
	
	fn := func(id string) (user *gosession.User, err error) {
		return &gosession.User{
			Id:     id,
			Detail: map[string]string{"detail": id},
		}, nil
	} // 可以设置获取用户信息的函数，如果用户没有缓存，会从该函数加载后存进redis，允许nil
	tokenManage.ConfigGetUserInfoFunc(fn)
	
	//tokenManage.SetSingleMode() // 你可以设置单点的token

	// 4. 为某用户设置Token
	id := "000001"
	var tokenExpireTimeAlone int64 = 2 // token过期时间设置2秒

	token, err := tokenManage.SetToken(id, tokenExpireTimeAlone)
	if err != nil {
		fmt.Println("set token err:", err.Error())
		return
	}

	fmt.Println("token:", token)

	// 可以设置多个令牌
	tokenManage.SetToken(id, 100)
	tokenManage.SetToken(id, 100)
	tokenManage.SetToken(id, 100)

	// 5. 列出用户所有的令牌
	tokenList, err := tokenManage.ListUserToken(id)
	if err != nil {
		fmt.Println("list token err:", err.Error())
		return
	}
	fmt.Println("list token:", tokenList)

	// 6. 检查token是否存在，存在会返回用户信息
	var userExpireTimeAlone int64 = 10 // 如果用户不存在并且ConfigGetUserInfoFunc!=nil，将会加载用户信息，重新放入redis
	u, exist, err := tokenManage.CheckTokenOrUpdateUser(token, userExpireTimeAlone)
	if err != nil {
		fmt.Println("check token err:", err.Error())
		return
	}

	fmt.Printf("check token:%#v, %#v,%#v\n", token, u, exist)

	err = tokenManage.RefreshToken(token, 5)
	if err != nil {
		fmt.Println("refresh token err:", err.Error())
		return
	}

	u, exist, err = tokenManage.CheckTokenOrUpdateUser(token, userExpireTimeAlone)
	if err != nil {
		fmt.Println("after refresh check token err:", err.Error())
		return
	}

	fmt.Printf("after refresh token:%#v, %#v,%#v\n", token, u, exist)

	// 7. 睡眠一下，看token是不是失效了
	time.Sleep(10 * time.Second)
	u, exist, err = tokenManage.CheckTokenOrUpdateUser(token, userExpireTimeAlone)
	if err != nil {
		fmt.Println("sleep check token err:", err.Error())
		return
	}

	fmt.Printf("sleep check token:%#v, %#v,%#v\n", token, u, exist)

	// 可以删除用户的所有令牌
	tokenList, err = tokenManage.ListUserToken(id)
	if err != nil {
		fmt.Println("sleep list token err:", err.Error())
		return
	}
	fmt.Println("sleep token:", tokenList)

	err = tokenManage.DeleteUserToken(id)
	if err != nil {
		fmt.Println("delete user token err:", err.Error())
		return
	}

	tokenList, err = tokenManage.ListUserToken(id)
	if err != nil {
		fmt.Println("after delete user token list err:", err.Error())
		return
	}
	fmt.Println("after delete user token list:", tokenList)
}
```

# 待做事项

1. 支持 JWT（JSON Web Token），特点是可以将部分客户端需要知道的信息保存在令牌里面，客户端可以无状态就发现令牌过期而不需要调用服务端。原理见：[博客](https://www.lenggirl.com/micro/auth-jwt.html) 。
2. 支持存储在 MySQL 或者 Mongo ，好处是排序，数据转移较容易，可以做更多业务操作。
3. 支持多客户端的一些资源隔离，主要是业务上的，比如 Android，IOS，Web端的多点和单点登录，以及审计的记录。

# License

```
Copyright [2019-2021] [github.com/hunterhug]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```