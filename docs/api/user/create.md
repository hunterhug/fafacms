# /user/create

创建用户，管理员接口。

用户默认是激活状态，但是不是VIP，需要调用另外的接口设置VIP，VIP用户可以操作内容相关的API。
## 请求

```
POST /use/create

{
	"name": "xxx",
	"nick_name": "xxx",
	"email": "user@example.com",
	"password": "123456789",
	"repassword": "123456789",
	"wechat": "",
	"weibo": "",
	"github": "",
	"qq": "",
	"gender": 1,
	"short_describe":"",
	"describe": "",
	"image_path": ""
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| name | 用户唯一标志 | string | 可用来登录，全局唯一，不能修改 | 是 |
| nick_name |    用户昵称   |   string | 昵称，全局唯一，可以修改| 是|
| email | 邮箱 | string | 唯一，可用来登录 | 是 |
| password | 用户密码 | string |  | 是 |
| repassword | 重复密码 | string | 必须与用户密码一样 | 是 |
| wechat | 微博URL | string |  | 否 |
| weibo | 微信URL | string |  | 否 |
| github | github URL | string |  | 否 |
| qq | QQ号 | string |  | 否 |
| gender | 性别 | int | 1表示男，2女，0未知 | 否 |
| short_describe | 一句话简介 | string | | 否 |
| describe | 描述 | string | | 否 |
| image_path | 头像地址 | string | 必须是通过上传接口上传的 | 否 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "037960a2cc594404ac6de5bfc7358dd8",
  "data": {
    "id": 19,
    "name": "xxssx",
    "nick_name": "xxx",
    "email": "user@example.com",
    "wechat": "",
    "weibo": "",
    "github": "",
    "qq": "",
    "password": "123456789",
    "gender": 1,
    "describe": "",
    "head_photo": "",
    "create_time": 1567048018,
    "activate_time": 1568445819,
    "status": 1,
    "vip": 0
  }
}
```

用户名被占用：

```
{
  "flag": false,
  "cid": "2df5cca8138b4ae78b5b1e0869cbfa7a",
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

邮箱被占用：

```
{
  "flag": false,
  "cid": "475d18031abe46a2921f223cb3406657",
  "error": {
    "id": 100023,
    "msg": "email already be used"
  }
}
```

用户头像不存在：

```
{
  "flag": false,
  "cid": "22cc8f1e0f764f3ba705b559f306b1f2",
  "error": {
    "id": 100030,
    "msg": "file can not be found"
  }
}
```

参数有误：

```
{
  "flag": false,
  "cid": "fab238b996964f1382a5d20d5bd1da89",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'RegisterUserRequest.RePassword' Error:Field validation for 'RePassword' failed on the 'eqfield' tag"
  }
}
```