# /user/info

获取自己的个人信息。

## 请求

```
GET /user/info
```

## 响应

正常：

```
{
  "flag": true,
  "cid": "c20dee87f9394c1398af321c2b3b2ae2",
  "data": {
    "id": 24,
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
    "create_time": "2019-09-11 17:10:24",
    "create_time_int": 1568193024,
    "update_time": "2019-09-21 15:03:43",
    "update_time_int": 1569049423,
    "login_time": "2019-09-24 11:28:16",
    "login_time_int": 1569295696,
    "login_ip": "127.0.0.1",
    "is_in_black": false,
    "is_vip": true,
    "followed_num": 0,
    "following_num": 0
  }
}
```

`HTTP` 头部需要携带 `token`，返回的结构和前端接口 `/u` 一样，只不过多了登录IP和登录时间。