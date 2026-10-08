# /u/info

查找某一个用户的信息。只能找出正常的已激活用户。未激活和黑名单的用户会找不到。

`content_num`（文章数）的口径与 `/u/content` 列表一致：

- 游客与其他用户：只统计**能看到的**文章（排除隐藏节点及其子节点下的文章）；
- 作者本人：统计自己的全部已发布文章（含隐藏节点下的文章）。

## 请求

不需要授权和前缀。

```
POST /u/info

{
	"user_id": 1,
	"user_name": ""
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id | 用户ID | int | user_name和user_id必存在一个 | 是 |
| user_name |    用户名  |  string | user_name和user_id必存在一个 | 是 |

## 响应

正常：

```
{
  "flag": true,
  "data": {
    "id": 1,
    "name": "admin",
    "nick_name": "admin",
    "email": "admin@admin",
    "wechat": "",
    "weibo": "",
    "github": "",
    "qq": "",
    "gender": 0,
    "describe": "",
    "head_photo": "",
    "create_time": "2019-08-11 13:45:45",
    "create_time_int": 1565502345,
    "login_time": "2019-09-14 15:39:23",
    "login_time_int": 1568446763,
    "is_vip": true,
    "is_in_black": false
  }
}
```

用户找不到：

```
{
  "flag": false,
  "error": {
    "id": 100003,
    "msg": "user not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "error": {
    "id": 100010,
    "msg": "paras input not right:where is empty"
  }
}
```