# /user/list

用户列表，管理员接口。

## 请求

```
POST /user/list

{
	"id": 0,
	"name": "",
	"create_time_begin": 0,
	"create_time_end": 0,
	"update_time_begin": 0,
	"update_time_end": 0,
	"limit": 1,
	"page": 1,
	"sort": [
            "=id",
            "=name",
            "-vip",
            "-activate_time",
            "=followed_num",
            "=following_num",
            "=content_num",
            "=content_cool_num",
            "=create_time",
            "=update_time",
            "=gender"
	],
	"email": "",
	"wechat": "",
	"weibo": "",
	"github": "",
	"qq": "",
	"gender": -1,
	"status": -1,
	"vip": -1
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的用户 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的用户  |  int | | |
| update_time_begin | 在此时间后修改过的用户 | int | 秒，留空不筛选| |
| update_time_end |    在此时间前修改过的用户   |   int | | |
| sort |    组排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该用户唯一id  |   int | 可查询单一用户信息 | |
| name |    用户登录名 |   string | 可查询单一组信息 | |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |
| wechat | 微博URL | string |  | 否 |
| weibo | 微信URL | string |  | 否 |
| github | github URL | string |  | 否 |
| qq | QQ号 | string |  | 否 |
| email | 邮箱 | string | 唯一，可用来登录 | 否 |
| gender | 性别 | int | -1找出全部，1表示男，2女，0未知 | 是 |
| status | 用户状态 | int | -1找出全部，0表示未激活，1已激活，2被记入黑名单 | 是 |
| vip | VIP用户 | int | -1找出全部，0表示查找非vip用户，1查找vip用户。 | 是 |

## 响应

```
{
  "flag": true,
  "cid": "a7646e438c2f45edac459f41fca66b52",
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
        "activate_time": 1568445819,
        "status": 1,
        "vip": 1
      }
    ],
    "limit": 1,
    "page": 1,
    "total_pages": 6
  }
}
```
