# /group/user/list

列出组下的用户，管理员接口。

## 请求

```
POST /group/user/list

{
	"group_id": 1
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| group_id | 用户组ID | int |  | 是 |

## 响应

```
{
  "flag": true,
  "cid": "b2b3c776b2644ab68f6bcda429a91eac",
  "data": {
    "users": [
      {
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
        "update_time": 1567059689,
        "status": 1,
        "group_id": 1
      }
    ]
  }
}
```

该接口没有分页功能，因为管理员与普通用户相比，相当少。