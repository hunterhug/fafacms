# /comment/admin/bad/list

管理员列出评论举报列表（`CommentBad` 表），联查被举报评论、所属文章与举报人信息。

## 请求

```
POST /comment/admin/bad/list

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
| status | 被举报评论当前状态 | int | -1 全部 / 0 正常 / 1 违禁 | 否 |
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
        "comment_id": 6,
        "comment_describe": "评论内容",
        "comment_user_id": 2,
        "comment_user_name": "test1",
        "comment_status": 0,
        "comment_is_delete": 0,
        "content_id": 10,
        "content_title": "文章标题",
        "reason": "色情低俗",
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
| comment_id / comment_describe | 被举报评论的 id / 正文 | int/string |
| comment_user_id / comment_user_name | 被举报评论作者的 id / 登录名 | int/string |
| comment_status | 被举报评论状态：0 正常 / 1 违禁 | int |
| comment_is_delete | 被举报评论是否已删除：0 否 / 1 是 | int |
| content_id / content_title | 所属文章的 id / 标题 | int/string |
| reason | 举报理由 | string |
| create_time | 举报时间 | int |
