# /user/register

用户注册。注册后的用户，激活后可以进行评论，如果需要编辑，发布内容相关，需要联系管理员授予vip，以此来区分评论的用户和内容作者。

## 请求

不需要前缀。

```
POST /user/register

{
	"name": "hunterhug",
	"nick_name": "小可爱",
	"password": "12345678",
	"repassword": "12345678",
	"email": "your_email@example.com"
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| name | 用户名 | string |  全局唯一，用来登录，不可以修改 | 是 |
| nick_name |    昵称  |  string | 自定义昵称，全局唯一，可以修改| 是 |
| password | 密码 | string |  | 是 |
| repassword |    重复密码   |   string | 必须与password相同|是 |
| email |    邮箱 | string | 全局唯一，用来登录 | 是 |

## 响应

将会发送一个 **6 位数字激活码**（有效期五分钟，连续错 5 次作废）到用户的邮箱，使用激活API可激活用户。

正常：

```
{
  "flag": true,
  "cid": "36158a4fc44f4a8eb3f8df53f15ebf6f"
}
```

用户名被占用：

```
{
  "flag": false,
  "cid": "8b8950f1889549c8bbc9c8de759d751a",
  "error": {
    "id": 100022,
    "msg": "user name already be used"
  }
}
```

昵称已被使用：

```
{
  "flag": false,
  "cid": "8b8950f1889549c8bbc9c8de759d751a",
  "error": {
    "id": 100019,
    "msg": "user nickname already be used"
  }
}
```

邮箱已被使用：

```
{
  "flag": false,
  "cid": "93e8cf9cbbd1425a98569812d3599866",
  "error": {
    "id": 100023,
    "msg": "email already be used"
  }
}
```

参数有误：

```
{
  "flag": false,
  "cid": "4c1c858943224a85b1f65731ecb4fd71",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'RegisterUserRequest.RePassword' Error:Field validation for 'RePassword' failed on the 'eqfield' tag"
  }
}
```