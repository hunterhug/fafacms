# /content/comment

列出内容下评论。如果游客访问该API，那么状态为隐藏和在垃圾箱的内容将会报内容不存在错误，而违禁的评论将会被和谐掉。如果和授权API一样在HTTP头部携带token，那么这些自己的内容和评论都正常显示。

## 请求

```
POST /content/comment

{
	"content_id": 0,
	"root_comment_id": -1,
	"limit": 10,
	"page": 1,
	"sort": [
		"-create_time",
		"-cool"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| content_id |    评论所在的内容ID |   int | | 是 |
| root_comment_id |   被评论的评论ID（底部评论） |   int | -1表示不筛选，0表示只查找第一级评论(这些一级评论会默认携带下属的3条评论，以及下属的数量)，其他表示查找该评论下面的评论 | 是 |
| sort |    组排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |



## 响应

正常，如果`root_comment_id!=0`：


```
{
  "flag": true,
  "cid": "4759c21956704fa3b93400946f82ffc4",
  "data": {
    "comments": [
      {
        "id": 5,
        "content_id": 15,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 0,
        "root_comment_id": 0,
        "comment_type": 0
      },
      {
        "id": 4,
        "content_id": 15,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 0,
        "root_comment_id": 0,
        "comment_type": 0
      },
      {
        "id": 3,
        "content_id": 15,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 2,
        "root_comment_id": 1,
        "comment_type": 2
      },
      {
        "id": 1,
        "content_id": 15,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 0,
        "root_comment_id": 0,
        "comment_type": 0
      }
    ],
    "extra": {
      "users": {
        "26": {
          "id": 26,
          "name": "fafa",
          "nick_name": "小花花",
          "head_photo": "",
          "is_vip": true
        }
      },
      "comments": {
        "1": {
          "id": 1,
          "describe": "大哥，我第一个跟随你",
          "content_id": 15,
          "create_time": 1569049596,
          "is_delete": false,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 0,
          "is_anonymous": true,
          "user_id": 0,
          "cool": 1,
          "bad": 0
        },
        "2": {
          "id": 2,
          "describe": "",
          "content_id": 0,
          "create_time": 0,
          "is_delete": true,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 0,
          "is_anonymous": false,
          "user_id": 0,
          "cool": 0,
          "bad": 0
        },
        "3": {
          "id": 3,
          "describe": "",
          "content_id": 15,
          "create_time": 1569049681,
          "is_delete": false,
          "is_ban": true,
          "ban_time": 1569308786,
          "delete_time": 0,
          "is_anonymous": false,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        },
        "4": {
          "id": 4,
          "describe": "大哥，我第一个跟随你",
          "content_id": 15,
          "create_time": 1569295618,
          "is_delete": false,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 0,
          "is_anonymous": true,
          "user_id": 0,
          "cool": 0,
          "bad": 0
        },
        "5": {
          "id": 5,
          "describe": "大哥fsfsfsdf",
          "content_id": 15,
          "create_time": 1569295636,
          "is_delete": false,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 0,
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
          "ban_time": 0,
          "user_id": 26,
          "user_name": "fafa",
          "seo": "today",
          "status": 0,
          "comment_num": 4
        }
      }
    },
    "limit": 10,
    "page": 1,
    "total_pages": 1
  }
}
```

关联数据`contents`和`comment`找不到这种情况不可能发生，获取关联数据`comment`的具体内容时请先判断`is_delete`字段。

如果`root_comment_id=0`会多出`son`和`son_num`字段，其他照旧。

内容不存在：


```
{
  "flag": false,
  "cid": "a20f3762b4e74a1a83708f3043740daa",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

参数不对：


```
{
  "flag": false,
  "cid": "a20f3762b4e74a1a83708f3043740daa",
  "error": {
    "id": 100010,
    "msg": "paras input not right:content_id empty"
  }
}
```