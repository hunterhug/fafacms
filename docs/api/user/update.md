# /user/update

更新信息，用户更新自己的信息。

## 请求

```
POST /user/update

{
	"nick_name": "xxx",
	"wechat": "",
	"weibo": "",
	"github": "",
	"qq": "",
	"gender": 0,
	"short_describe":"",	
	"describe": "",
	"image_path": ""
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| nick_name |    用户昵称   |   string | 全局唯一，可以修改。用户每个月只能修改两次昵称。 | 否|
| wechat | 微博URL | string |  | 否 |
| weibo | 微信URL | string |  | 否 |
| github | github URL | string |  | 否 |
| qq | QQ号 | string |  | 否 |
| gender | 性别 | int | 1表示男，2女，0表示不修改性别 | 否 |
| short_describe | 一句话简介 | string | | 否 |
| describe | 描述 | string | | 否 |
| image_path | 头像地址 | string | 必须是通过上传接口上传的 | 否 |

非空字段将会被更新。

## 响应

正常：

```
{
  "flag": true,
  "cid": "6fb2f646f36844eba265a3e1aa88cd68",
  "data": {
    "id": 1,
    "name": "",
    "nick_name": "xxx",
    "email": "",
    "wechat": "",
    "weibo": "",
    "github": "",
    "qq": "",
    "gender": 0,
    "describe": "",
    "head_photo": "",
    "create_time": 0,
    "update_time": 1567050899,
    "status": 0
  }
}
```

参数原封不动返回，可忽视返回的数据。

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

昵称不能修改，因为一个月只能修改两次：

```
{
  "flag": false,
  "cid": "8b8950f1889549c8bbc9c8de759d751a",
  "error": {
    "id": 100018,
    "msg": "user nickname can not change for time not reach:remain 2 days"
  }
}
```

上面表示还剩2天才能修改用户昵称。

参数有误：

```
{
  "flag": false,
  "cid": "0b5d984cd633465abb22dfacf351a846",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateUserRequest.WeiBo' Error:Field validation for 'WeiBo' failed on the 'url' tag"
  }
}
```