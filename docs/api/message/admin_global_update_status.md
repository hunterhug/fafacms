# /message/admin/global/update/status

修改管理员通知状态。管理员接口。

## 请求

```
POST /message/admin/global/update/status

{
	"id": 1,
	"status": 0
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id |    数据库该管理员通知唯一id  |   int | | 是 |
| status |    管理员通知状态 |   int | 0表示生效的通知，1表示暂停的通知，2表示失效的通知 | 是 |


用户登录后，拉取站内信的时候会将生效的管理员通知，插入站内信表，下一次拉取站内信的时候就可以看到。

## 响应

正常：

```
{
  "flag": true,
  "cid": "cf20e1be6de445248c1a92501955c91c"
}
```

不存在：

```
{
  "flag": false,
  "cid": "84f903cc1a3142aaa94ba7d73250cf4e",
  "error": {
    "id": 110011,
    "msg": "global message not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "de061195d6234867a09fe3452d34f2be",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateGlobalMessageRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```
