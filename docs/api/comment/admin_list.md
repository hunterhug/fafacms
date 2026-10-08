# /comment/admin/list

列出评论。管理员接口。无论是匿名还是违禁，还是被删除的评论都会被列出，方便管理员查看。游客的评论列出API请查看前端API页面。

## 请求

```
POST /comment/admin/list

{
	"id": 0,
	"user_id": 0,
	"user_name": "",
	"content_id": 0,
	"content_user_id": 0,
	"content_user_name": "",
	"comment_id": 0,
	"comment_user_id": 0,
	"comment_user_name": "",
	"root_comment_id": 0,
	"root_comment_user_id": 0,
	"root_comment_user_name": "",
	"comment_type": -1,
	"status": -1,
	"is_delete": -1,
	"comment_anonymous": -1,
	"create_time_begin": 0,
	"create_time_end": 0,
	"limit": 10,
	"page": 1,
	"sort": [
		"=id",
		"-create_time",
		"=content_id",
		"=user_id",
		"=cool",
		"=bad"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的评论 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的评论  |  int | | |
| sort |    评论排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该评论唯一id  |   int | 可查询单一评论信息 | |
| status |    评论状态 |   int | 可作筛选条件，-1表示查找全部状态，0查找正常，1查找违禁状态的 | 是 |
| is_delete |    是否查找删除的评论  |   int | 可作筛选条件，-1表示查找全部，0表示查找正常评论，1查找已经被删除的 | 是 |
| comment_anonymous |   是否查找匿名的评论 |   int | 可作筛选条件，-1表示查找全部，0表示查找正常评论，1查找匿名的评论 | 是 |
| comment_type |    评论类型 |   int | 可作筛选条件，-1表示查找全部类型，0表示对内容的评论，1表示第一级别的对评论的评论，2表示超过一级的对评论的评论 | 是 |
| user_id |    评论人ID |   int | |  |
| user_name |    评论人名字 |   string |  | |
| content_id |    评论所在的内容ID |   int | |  |
| content_user_id |    评论所在的内容，这些内容属于某用户 |   int | |  |
| content_user_name |    评论所在的内容，这些内容属于某用户 |   string | |  |
| comment_id |   被评论的评论ID（超过一级的） |   int | comment_type=2时有效 |  |
| comment_user_id |    被评论的评论，这些评论属于某用户 |   int | comment_type=2时有效 |  |
| comment_user_name |    被评论的评论，这些评论属某用户 |   string | comment_type=2时有效  |  |
| root_comment_id |   根评论ID |   int | 可作筛选条件，-1表示查找全部层级，0表示只查顶层评论（对内容的评论），>0表示指定根评论 |  |
| root_comment_user_id |    被评论的评论，这些评论属于某用户 |   int | comment_type=1,2时有效 |  |
| root_comment_user_name |    被评论的评论，这些评论属某用户 |   string | ccomment_type=1,2时有效  |  |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |


当`comment_type`为0，表示只是对内容评论了，这时`comment_id`和`root_comment_id`都为空。

当`comment_type`为1，表示只是对`comment_type=0`的评论进行评论，这时`comment_id`为0，而`root_comment_id`非空。

当`comment_type`为2，表示只是对`comment_type=1`的评论进行评论，这时`comment_id`和`root_comment_id`都非空。

`root_comment_id`表示最根源的评论，所有评论的评论，只要在这个评论`A`的基础上，那么`root_comment_id`都是`A`.

## 响应

```
{
  "flag": true,
  "cid": "361723b06c86496e9c356fdd89af01ab",
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
        "id": 3,
        "content_id": 15,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 2,
        "root_comment_id": 1,
        "comment_type": 2
      },
      {
        "id": 2,
        "content_id": 15,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 0,
        "root_comment_id": 1,
        "comment_type": 1
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
          "create_time": 1569049596,
          "is_delete": false,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 0,
          "is_anonymous": true,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        },
        "2": {
          "id": 2,
          "describe": "我不愿意你说：大哥，我第一个跟随你",
          "create_time": 1569049669,
          "is_delete": true,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 1569053889,
          "is_anonymous": false,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        },
        "3": {
          "id": 3,
          "describe": "我也不愿意",
          "create_time": 1569049681,
          "is_delete": false,
          "is_ban": false,
          "ban_time": 0,
          "delete_time": 0,
          "is_anonymous": false,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        },
        "5": {
          "id": 5,
          "describe": "大哥fsfsfsdf",
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
          "status": 0
        }
      }
    },
    "limit": 10,
    "page": 1,
    "total_pages": 1
  }
}
```

参数说明：

| 字段   |      含义   |
|----------|--------|
| comments | 列表数据 |
| extra | 额外关联数据 |

关联数据`contents`和`comment`找不到这种情况不可能发生，获取关联数据`comment`的具体内容时请先判断`is_delete`字段。