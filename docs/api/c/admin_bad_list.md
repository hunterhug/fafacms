# /content/admin/bad/list

管理员列出内容举报列表（`ContentBad` 表），联查被举报文章与举报人信息。

## 请求

```
POST /content/admin/bad/list

{
	"status": -1,
	"sort": ["-id"],
	"limit": 20,
	"page": 1
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| status | 被举报文章当前状态 | int | -1 全部 / 0 正常 / 1 隐藏 / 2 违禁 / 3 回收站 | 否 |
| sort | 排序 | []string | 默认按 id 倒序（最新举报在前） | 否 |
| limit | 每页数量 | int | 默认 20，最大 100 | 否 |
| page | 第几页 | int | 默认 1 | 否 |

## 响应

```
{
  "flag": true,
  "cid": "xxxx",
  "data": {
    "bad": [
      {
        "id": 1,
        "user_id": 3,
        "user_name": "test2",
        "user_nick": "鸡鸡2",
        "content_id": 10,
        "content_title": "文章标题",
        "content_user_id": 2,
        "content_user_name": "test1",
        "content_status": 0,
        "reason": "垃圾广告",
        "create_time": 1570345455
      }
    ],
    "limit": 20,
    "total": 1,
    "page": 1,
    "total_pages": 1
  }
}
```

字段说明：

| 字段   |      含义   | 类型  |
|----------|--------|------|
| id | 举报记录唯一 id | int |
| user_id / user_name / user_nick | 举报人的 id / 登录名 / 昵称 | int/string |
| content_id / content_title | 被举报文章的 id / 标题 | int/string |
| content_user_id / content_user_name | 被举报文章作者的 id / 登录名 | int/string |
| content_status | 被举报文章当前状态：0 正常 / 1 隐藏 / 2 违禁 / 3 回收站 / 4 用户彻底删除 | int |
| reason | 举报理由 | string |
| create_time | 举报时间 | int |
