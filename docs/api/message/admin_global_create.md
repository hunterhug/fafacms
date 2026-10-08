# /message/admin/global/create

创建管理员通知。管理员接口。

## 请求

点对点发送，实时发送：

```
POST /message/admin/global/create

{
	"all_people": false,
	"user_ids": [
		1,
		2,
		3
	],
	"message": "this is admin message"
}
```

广播式的通知，发送给所有人：

```
POST /message/admin/global/create

{
	"all_people": true,
	"message": "this is admin message",
	"right_now": true
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| all_people |    是否是广播式通知   |   bool | | 是 |
| user_ids | 点对点发送时的用户列表 | array | all_people为false时有效 | 否 |
| message | 通知的内容 | string |  | 是 |
| right_now | 广播马上生效 | bool | all_people为true时有效，表示广播马上生效，如果`right_now`为true时，用户会在拉取信息接口的下一次拉取信息得到这个通知。 | 否 |



## 响应

正常：

```
{
  "flag": true,
  "cid": "d1792391b6b94a25b0832f0458c516ed"
}
```


参数有误：

```
{
  "flag": false,
  "cid": "caf3c29118b34d13bf3ff8054d6c481d",
  "error": {
    "id": 100010,
    "msg": "paras input not right:user_ids not right"
  }
}
```