# Distributed Session Implement By Golang

[![GitHub forks](https://img.shields.io/github/forks/hunterhug/gosession.svg?style=social&label=Forks)](https://github.com/hunterhug/gosession/network)
[![GitHub stars](https://img.shields.io/github/stars/hunterhug/gosession.svg?style=social&label=Stars)](https://github.com/hunterhug/gosession/stargazers)
[![GitHub last commit](https://img.shields.io/github/last-commit/hunterhug/gosession.svg)](https://github.com/hunterhug/gosession)
[![Go Report Card](https://goreportcard.com/badge/github.com/hunterhug/gosession)](https://goreportcard.com/report/github.com/hunterhug/gosession)
[![GitHub issues](https://img.shields.io/github/issues/hunterhug/gosession.svg)](https://github.com/hunterhug/gosession/issues)

[中文说明](/README_ZH.md)

Support multi web service share session token, which can keep union state in many diff service.

Now session token can store in:

1. Single mode Redis.
2. Sentinel mode Redis.

## Usage

simple get it by:

```
go get -v github.com/hunterhug/gosession
```

core api:

```go
// token manage
// token will be put in cache database such redis and user info relate with that token will cache too
type TokenManage interface {
	SetToken(id string, tokenValidTimes int64) (token string, err error)                               // Set token, expire after some second
	RefreshToken(token string, tokenValidTimes int64) error                                            // Refresh token，token expire will be again after some second
	DeleteToken(token string) error                                                                    // Delete token when you do action such logout
	CheckToken(token string) (user *User, exist bool, err error)                                       // Check the token,  but not refresh user info cache
	CheckTokenOrUpdateUser(token string, userInfoValidTimes int64) (user *User, exist bool, err error) // Check the token, when cache database exist return user info directly, others hit the persistent database and save newest user in cache database then return. such redis check, not check load from mysql.
	ListUserToken(id string) ([]string, error)                                                         // List all token of one user
	DeleteUserToken(id string) error                                                                   // Delete all token of this user
	RefreshUser(id []string, userInfoValidTimes int64) error                                           // Refresh cache of user info batch
	DeleteUser(id string) error                                                                        // Delete user info in cache
	AddUser(id string, userInfoValidTimes int64) (user *User, exist bool, err error)                   // Add the user info to cache，expire after some second
	ConfigTokenKeyPrefix(tokenKey string) TokenManage                                                  // Config chain, just cache key prefix
	ConfigUserKeyPrefix(userKey string) TokenManage                                                    // Config chain, just cache key prefix
	ConfigExpireTime(second int64) TokenManage                                                         // Config chain, token expire after second
	ConfigGetUserInfoFunc(fn GetUserInfoFunc) TokenManage                                              // Config chain, when cache not found user info, will load from this func
	SetSingleMode() TokenManage                                                                        // Can set single mode, before one new token gen, will destroy other token
}

// core user info, it's Id will be the primary key store in cache database such redis
type User struct {
	Id                  string      `json:"id"`     // unique mark
	TokenRemainLiveTime int64       `json:"-"`      // token remain live time in cache
	Detail              interface{} `json:"detail"` // can diy your real user info by config ConfigGetUserInfoFunc()
}
```

one simple example:

```go
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

another example:

```go
package main

import (
	"fmt"
	"github.com/hunterhug/gosession"
	"time"
)

func main() {
	// 1. config redis
	redisHost := "127.0.0.1:6379"
	redisDb := 0
	redisPass := "hunterhug" // may redis has password
	redisConfig := gosession.NewRedisSessionSingleModeConfig(redisHost, redisDb, redisPass)
	// or
	//gosession.NewRedisSessionSentinelModeConfig(":26379,:26380,:26381",0,"mymaster")

	// 2. connect redis session
	tokenManage, err := gosession.NewRedisSession(redisConfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 3. config token manage
	tokenManage.ConfigDefaultExpireTime(600)
	tokenManage.ConfigUserKeyPrefix("go-user")
	tokenManage.ConfigTokenKeyPrefix("go-token")
	fn := func(id string) (user *gosession.User, err error) {
		return &gosession.User{
			Id:     id,
			Detail: map[string]string{"detail": id},
		}, nil
	} // get user func diy, you can set it nil
	tokenManage.ConfigGetUserInfoFunc(fn)
	//tokenManage.SetSingleMode()

	// 4. set token
	id := "000001"
	var tokenExpireTimeAlone int64 = 2

	token, err := tokenManage.SetToken(id, tokenExpireTimeAlone)
	if err != nil {
		fmt.Println("set token err:", err.Error())
		return
	}

	fmt.Println("token:", token)

	// can set a lot token
	tokenManage.SetToken(id, 100)
	tokenManage.SetToken(id, 100)
	tokenManage.SetToken(id, 100)

	// 5. list all token
	tokenList, err := tokenManage.ListUserToken(id)
	if err != nil {
		fmt.Println("list token err:", err.Error())
		return
	}
	fmt.Println("list token:", tokenList)

	// 6. check token
	var userExpireTimeAlone int64 = 10 // if ConfigGetUserInfoFunc!=nil, will load user info from func if not exist in redis cache
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

	// 7. sleep to see token is exist?
	time.Sleep(10 * time.Second)
	u, exist, err = tokenManage.CheckTokenOrUpdateUser(token, userExpireTimeAlone)
	if err != nil {
		fmt.Println("sleep check token err:", err.Error())
		return
	}

	fmt.Printf("sleep check token:%#v, %#v,%#v\n", token, u, exist)

	// you can delete all token of one user
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

# TODO

1. Support JWT token, that will enable token verify in client side to find the token expire early, also we can filter in server side.
2. Support Other Db such MySQL or Mongo.
3. Support Feature of multiple clients single sign on.

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