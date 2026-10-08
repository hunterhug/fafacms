# /comment/delete

删除自己的评论，是**逻辑删除**（`is_delete=1`），不物理删除。内容移入回收站(status=3)或彻底删除(status=4)也不会级联删除评论，评论记录与通知引用仍保留（前端对已删除评论显示占位符「评论被删除」）。

## 请求

```
POST  /comment/delete

{
	"id": 2
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | 评论id | 是 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "2e5b753d299c4d54b6b3b44382c6e36b"
}
```


评论不存在：

```
{
  "flag": false,
  "cid": "33d3fc07efc64da182524e9ad400a8c8",
  "error": {
    "id": 110008,
    "msg": "comment not found"
  }
}
```

参数有误：

```
{
  "flag": false,
  "cid": "576444a50afd48de898c938dbc122640",
  "error": {
    "id": 100010,
    "msg": "paras input not right:comment_id empty"
  }
}
```

评论删除后，如果是其他评论的关联数据，调用`take API`后，评论ID为2的详情字段：`is_delete=true`

```
{
  "flag": true,
  "cid": "8a988566ab20482984399929816ad97f",
  "data": {
    "comment": {
      "id": 3,
      "content_id": 15,
      "content_title": "今天放学后我做了一件好事",
      "comment_id": 2,
      "root_comment_id": 1,
      "comment_type": 2
    },
    "extra": {
      "users": {
        "26": {
          "id": 26,
          "name": "fafa",
          "nick_name": "小花花",
          "head_photo": "",
          "is_vip": false
        }
      },
      "comments": {
        "1": {
          "id": 1,
          "describe": "大哥，我第一个跟随你",
          "create_time": 1569049596,
          "is_delete": false,
          "is_ban": false,
          "is_anonymous": true,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        },
        "2": {
          "id": 2,
          "describe": "",
          "create_time": 0,
          "is_delete": true,
          "is_ban": false,
          "is_anonymous": false,
          "user_id": 0,
          "cool": 0,
          "bad": 0
        },
        "3": {
          "id": 3,
          "describe": "我也不愿意",
          "create_time": 1569049681,
          "is_delete": false,
          "is_ban": false,
          "is_anonymous": false,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        }
      },
      "contents": {
        "15": {
          "id": 15,
          "title": "今天放学后我做了一件好事",
          "is_hide": false,
          "is_in_rubbish": false,
          "is_ban": false,
          "user_id": 26,
          "user_name": "fafa",
          "seo": "today",
          "status": 0
        }
      }
    }
  }
}
```
