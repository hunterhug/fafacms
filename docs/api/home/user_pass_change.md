# /user/password/change

修改密码

## 请求

不需要前缀。

```
POST /user/password/change

{
	"email": "your_email@example.com",
	"password": "12345678",
	"repassword": "12345678",
	"code": "520637"
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| code | 修改密码验证码 | string |  6 位数字，邮箱中查看，有效期五分钟 | 是 |
| email |    邮箱 | string | 全局唯一，用来登录 | 是 |
| password | 新的密码 | string |  | 是 |
| repassword |    重复密码   |   string | 必须与password相同|是 |

## 响应


正常：

```
{
  "flag": true,
  "cid": "6ce755621e8f43bd89205400eed5dcaf"
}
```

邮箱不存在：

```
{
  "flag": false,
  "cid": "7ba512fd98d24f97ae6fc711cfb95f17",
  "error": {
    "id": 100027,
    "msg": "email not found"
  }
}
```

验证码错误（剩余次数随错误返回，错满 5 次验证码作废，需重新申请）：

```
{
  "flag": false,
  "cid": "3738e58e35354452baab8e198712c6ab",
  "error": {
    "id": 100029,
    "msg": "reset code wrong:4 attempts left"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "db5e75ff5dd841d7ba247ac77a89b0f2",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ChangePasswordRequest.Code' Error:Field validation for 'Code' failed on the 'gt' tag"
  }
}
```
