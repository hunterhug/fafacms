# /u

获取已经激活的正常用户列表，可翻页。

## 请求

不需要授权和前缀。

```
POST /u

{
    "vip": -1,
    "nick_name": "",
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
	"limit": 2,
	"offset": 0
}
```

`vip`为-1时表示查找所有用户，0时表示查找普通用户，1表示查找VIP用户。

`nick_name` 非空时按昵称模糊搜索（`nick_name like '%x%'`）；为空时返回推荐列表（此时超管 admin 不出现，搜索时可搜到）。

## 响应


```
{
  "flag": true,
  "data": {
    "users": [
      {
        "id": 14,
        "name": "123456",
        "nick_name": "xxx",
        "email": "user@example.com",
        "wechat": "",
        "weibo": "",
        "github": "",
        "qq": "",
        "gender": 1,
        "describe": "",
        "head_photo": "",
        "create_time": "2019-09-14 15:51:32",
        "create_time_int": 1568447492,
        "activate_time": "2019-09-14 15:51:32",
        "activate_time_int": 1568447492,
        "login_time": "2019-09-14 15:51:46",
        "login_time_int": 1568447506,
        "is_vip": true,
        "is_in_black": false
      },
      {
        "id": 10,
        "name": "xxswwsx",
        "nick_name": "xxx",
        "email": "user@example.com",
        "wechat": "",
        "weibo": "",
        "github": "",
        "qq": "",
        "gender": 1,
        "describe": "",
        "head_photo": "",
        "create_time": "2019-09-14 15:23:39",
        "create_time_int": 1568445819,
        "activate_time": "2019-09-14 15:23:39",
        "activate_time_int": 1568445819,
        "is_vip": true,
        "is_in_black": false
      }
    ],
    "limit": 2,
    "page": 1,
    "total_pages": 2
  }
}
```