# /message/admin/global/list

列出管理员通知。管理员接口。

## 请求

```
POST /message/admin/global/list

{
	"id": 0,
	"status": -1,
	"create_time_begin": 0,
	"create_time_end": 0,
	"limit": 10,
	"page": 1,
	"sort": [
		"=id",
		"-create_time",
		"status",
		"=total",
		"=success"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的管理员通知 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的管理员通知  |  int | | |
| sort |    管理员通知排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该管理员通知唯一id  |   int | 可查询单一管理员通知| |
| status |    管理员通知状态 |   int | 可作筛选条件，-1表示查找全部状态，0表示查找生效的通知，1表示查找暂停的通知，2表示查找失效的通知 | 是 |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

调用创建接口时，如果是广播式通知，并且 `right_now` 为true，那么 `status` 等于0，表示生效的通知，否则等于1，表示暂停的通知。

当用户登录后拉取站内信接口时，会查找一个星期内 `status` 等于0（生效的通知），然后进行数据库比对是否曾经收到过，如果没有则插入站内信表，第二次拉取站内信时会发现管理员通知已到。

## 响应

正常：

```
{
  "flag": true,
  "cid": "36d97785fbaf472796f8f2d876962955",
  "data": {
    "message": [
      {
        "id": 1,
        "create_time": 1570334454,
        "update_time": 0,
        "send_message": "this is admin message",
        "status": 1,
        "total": 4,
        "success": 0
      }
    ],
    "limit": 10,
    "total": 1,
    "page": 1,
    "total_pages": 1
  }
}
```

`total` 表示创建管理员通知时，系统总的激活用户数。`success` 表示成功收到管理员通知的用户数。`status` 为通知的状态，0表示生效的通知，1表示暂停的通知，2表示失效的通知。