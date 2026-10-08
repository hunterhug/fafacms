# /content/history/admin/list

列出内容历史，管理员接口。可查看所有人的内容历史。

## 请求

```
POST /content/history/admin/list

{
	"content_id": 0,
	"user_id":0,
	"create_time_begin": 0,
	"create_time_end": 0,
	"sort": [
		"=id",
		"-user_id",
		"-create_time",
		"-content_id"
	],
	"limit": 1,
	"page": 1
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| content_id |    历史所属的内容id  |   int | 可作筛选条件 | 否 |
| types | 历史的类型 | int | -1查找全部，0表示查找更新时保存的草稿，1表示发布时保存的内容，2表示恢复时保存的草稿 | 是 | 
| user_id |    历史内容所属用户id  |   int | 可作筛选条件 | 否 |
| create_time_begin | 在此时间后创建的历史内容 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的历史内容  |  int | | |
| sort |    内容排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

## 响应

```
{
  "flag": true,
  "cid": "6740c2a90c8942cd89aec3f2cdc2cf2a",
  "data": {
    "contents": [
      {
        "id": 9,
        "content_id": 1,
        "title": "1sss234",
        "user_id": 24,
        "node_id": 2,
        "describe": "",
        "types": 1,
        "create_time": 1568282900
      }
    ],
    "limit": 1,
    "page": 1,
    "total_pages": 9
  }
}
```

参数说明：

| 字段   |      含义   |类型  |   详情 |
|----------|--------|------|------|
| id | 内容历史ID，数据库唯一标记 | int | 作为唯一标记定位该节点|
| user_id | 内容历史所属用户ID | int | |
| node_id | 内容历史所属的内容的结点 | int | |
| title |    记录瞬间的内容标题   |   string ||
| describe |  记录瞬间的内容描述 | string |   |
| create_time | 内容历史创建时间 | int | 时间戳 |
| types | 历史的类型 | int | 0表示更新时保存的草稿，1表示发布时保存的内容，2表示恢复时保存的草稿 |


原理如下：

更新内容 ---》 将预发布内容刷新进历史表，类型0 ----》 更新预发布内容

发布内容 ---》 将预发布内容刷新进历史表，类型1 ----》 将预发布内容刷进正式内容

从历史内容恢复内容 ---》 将预发布内容刷新进历史表，类型2 ----》 将历史中的内容刷进预发布内容