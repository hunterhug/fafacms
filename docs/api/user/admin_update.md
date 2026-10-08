# /user/admin/update

修改用户昵称，密码，拉黑用户，管理员接口


## 请求

```
POST /user/admin/update

{
	"id": 1,
	"nick_name": "ssss",
	"password": "sssss",
	"status": 0，
	"vip": 0
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 用户唯一标志 | int | 需要修改的用户ID | 是 |
| nick_name |    用户昵称   |   string | 全局唯一，可修改，管理员可无限制修改用户昵称 | 否|
| password | 用户密码 | string | 留空不修改 | 否 |
| status | 用户状态 | int | 1表示激活，2表示拉黑，留空不作用 | 否 |
| vip | VIP状态 | int | 1表示增加vip，2表示去除vip，留空不作用 | 否 |

非空字段将会被更新。只有VIP的用户才可以创建节点，创建、编辑和删除内容。所有用户都可以评论。

## 响应

正常：

```
{
  "flag": true,
  "cid": "681ed82fdaec4f11a21291359d5d2278",
  "data": {
    "id": 1,
    "name": "",
    "nick_name": "ssss",
    "email": "",
    "wechat": "",
    "weibo": "",
    "github": "",
    "qq": "",
    "password": "sssss",
    "gender": 0,
    "describe": "",
    "head_photo": "",
    "create_time": 0,
    "update_time": 1567059334,
    "status": 0
  }
}
```

返回可忽视。

用户不存在：

```
{
  "flag": false,
  "cid": "3d920f0ed0b84739a9fbb13ab3070398",
  "error": {
    "id": 100003,
    "msg": "user not found"
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