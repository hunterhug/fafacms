# /message/admin/list

列出站内信。可以列出所有的信息，管理员接口。

## 请求

```
POST /message/list

{
	"receive_user_id": 0,
	"send_status": -1,
	"message_id": 0,
	"message_type": -1,
	"receive_status": -1,
	"chanel_user_id": 0,
	"create_time_begin": 0,
	"create_time_end": 0,
	"global_message_id": 0,
	"limit": 10,
	"page": 1,
	"sort": [
		"=id",
		"-create_time",
		"=receive_status",
		"=send_status",
		"=message_type",
		"=send_user_id",
		"=receive_user_id"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的信息 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的信息  |  int | | |
| sort |    排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该信息唯一id  |   int | 可查询单一信息| |
| message_type |    信息类型 |   int | 可作筛选条件，-1表示查找全部类型，具体类型如下 | 是 |
| receive_status |    信息的接收状态 |   int | 可作筛选条件，-1表示查找全部状态，0表示未读信息，1已读 | 是 |
| chanel_user_id |    私信频道，用户ID |   int | 非空时为查找私聊，可以列出与该用户的来往信息，不仅仅是接收的信息，而是发送和接收的信息，比较特殊 | 否 |
| global_message_id |    管理员通知ID |   int | 可作筛选条件，表示查找通过该管理员通知发送的信息 | 否 |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

与普通的列出接口比较，多了两个参数：


| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| receive_user_id | 站内信接收的用户ID | int | 非空时查找发送给该用户的所有信息 | 否 |
| send_status |    信息的发送状态  |  int | 私聊状态的信息有效，-1表示查找所有，0表示查找正常的信息，1表示已经被发送方删除的信息 | 是 |

信息类型如下：

```
	// who comment your content or comment message
	MessageTypeCommentForContent = 0 // 内容被评论
	MessageTypeCommentForComment = 1 // 评论被评论

	// who good your content or comment message
	MessageTypeGoodContent = 2 // 内容被点赞
	MessageTypeGoodComment = 3 // 评论被点赞

	// comment or content be ban by system message
	MessageTypeContentBan = 4 // 内容被违禁
	MessageTypeCommentBan = 5 // 评论被违禁

	// comment or content be ban by recover message
	MessageTypeContentRecover = 6 // 内容被解除违禁
	MessageTypeCommentRecover = 7 // 评论被解除违禁

	// who follow you message
	MessageTypeFollow = 8 // 有人关注你

	// who you follow publish content
	MessageTypeContentPublish = 9 // 你关注的人发布了内容

	// who send a message to you
	MessageTypePrivate = 10 // 私信

	// global send a message to you
	MessageTypeGlobal = 11 // 管理员通知
```

## 响应

正常：

```
{
  "flag": true,
  "cid": "b6ec84f3f4c940e9a48efeeb9dd0a0b4",
  "data": {
    "messages": [
      {
        "id": 45,
        "send_message": "this is admin message",
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 0,
        "create_time": 1570345455,
        "read_time": 0,
        "user_id": 0,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 11,
        "comment_is_your_self": 0,
        "global_message_id": 1
      },
      {
        "id": 43,
        "private_chanel": "2_1",
        "send_user_id": 1,
        "send_message": "eeeeeeeeeeeeeeee",
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 0,
        "create_time": 1570343461,
        "read_time": 0,
        "user_id": 0,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 10,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 42,
        "private_chanel": "2_1",
        "send_user_id": 1,
        "send_message": "eeeeeeeeeeeeeeee",
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 0,
        "create_time": 1570343460,
        "read_time": 0,
        "user_id": 0,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 10,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 34,
        "send_message": "this is admin message",
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 0,
        "create_time": 1570334819,
        "read_time": 0,
        "user_id": 0,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 11,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 31,
        "send_message": "this is admin message",
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 0,
        "create_time": 1570334393,
        "read_time": 0,
        "user_id": 0,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 11,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 28,
        "send_message": "this is admin message",
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 0,
        "create_time": 1570334389,
        "read_time": 0,
        "user_id": 0,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 11,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 26,
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 1,
        "create_time": 1569668483,
        "read_time": 1569669391,
        "user_id": 2,
        "content_id": 0,
        "content_title": "",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 8,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 25,
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 1,
        "create_time": 1569658534,
        "read_time": 0,
        "user_id": 3,
        "content_id": 1,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 0,
        "comment_describe": "",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 2,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 23,
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 1,
        "create_time": 1569658499,
        "read_time": 1569669391,
        "user_id": 3,
        "content_id": 1,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 3,
        "comment_describe": "wo  ye ai ni aaaaaaaaa",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 3,
        "comment_is_your_self": 0,
        "global_message_id": 0
      },
      {
        "id": 18,
        "send_status": 0,
        "receive_user_id": 2,
        "receive_status": 1,
        "create_time": 1569658438,
        "read_time": 1569669391,
        "user_id": 3,
        "content_id": 1,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 6,
        "comment_describe": "大哥，cssssss222hong aaaaaaaaa",
        "comment_anonymous": 0,
        "publish_again": 0,
        "message_type": 1,
        "comment_is_your_self": 0,
        "global_message_id": 0
      }
    ],
    "comments": {
      "3": {
        "id": 3,
        "content_id": 1,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 0,
        "root_comment_id": 2,
        "comment_type": 1
      },
      "6": {
        "id": 6,
        "content_id": 1,
        "content_title": "今天放学后我做了一件好事",
        "comment_id": 4,
        "root_comment_id": 2,
        "comment_type": 2
      }
    },
    "contents": {
      "1": {
        "id": 1,
        "title": "今天放学后我做了一件好事",
        "is_hide": false,
        "is_in_rubbish": false,
        "is_ban": false,
        "ban_time": 0,
        "user_id": 2,
        "user_name": "test1",
        "seo": "today",
        "status": 0,
        "comment_num": 6,
        "is_yourself": false
      }
    },
    "extra_users": {
      "2": {
        "id": 2,
        "name": "test1",
        "nick_name": "鸡鸡1",
        "head_photo": "",
        "is_vip": true,
        "short_describe": "",
        "gender": 1,
        "is_black": false
      },
      "3": {
        "id": 3,
        "name": "test2",
        "nick_name": "鸡鸡2",
        "head_photo": "",
        "is_vip": false,
        "short_describe": "",
        "gender": 1,
        "is_black": false
      }
    },
    "extra_comments": {
      "3": {
        "id": 3,
        "describe": "wo  ye ai ni aaaaaaaaa",
        "content_id": 1,
        "create_time": 1569658392,
        "is_delete": false,
        "is_ban": false,
        "ban_time": 0,
        "delete_time": 0,
        "is_anonymous": false,
        "user_id": 2,
        "cool": 1,
        "bad": 0,
        "is_yourself": true
      },
      "6": {
        "id": 6,
        "describe": "大哥，cssssss222hong aaaaaaaaa",
        "content_id": 1,
        "create_time": 1569658438,
        "is_delete": false,
        "is_ban": false,
        "ban_time": 0,
        "delete_time": 0,
        "is_anonymous": true,
        "user_id": 0,
        "cool": 0,
        "bad": 0,
        "is_yourself": false
      }
    },
    "un_read": {
      "10": 2,
      "11": 3
    },
    "limit": 10,
    "total": 19,
    "page": 1,
    "total_pages": 2
  }
}
```

`un_read`字段表示不同类型的未读消息数量。

`messages`是主体：

参数说明：

| 字段   |      含义   |类型  |   详情 |
|----------|--------|------|------|
| id | 消息ID，数据库唯一标记 | int | |
| receive_user_id | 消息接收的人 | int | 在这里为自己 |
| receive_status | 消息接收状态 | int | 0表示未读，1表示已读 |
| create_time |    消息创建的时间   |   int | |
| read_time |  消息已读的时间 | int |   |
| user_id | 哪个用户触发了这个信息 | int | 一般为站内信通知，比如某某评论了你的内容或评论，点赞了你的内容或评论，某某关注了你，某某发布了内容 |
| content_id |  内容ID | int |  与内容有关的通知信息会携带该字段 |
| content_title | 触发通知时，内容的瞬时标题 | string |标题不是最新的，通过后面的`contents`可以获取最新 |
| comment_id |  评论ID | int |  与评论有关的通知信息会携带该字段，有人评论了您 |
| comment_describe | 评论的内容 | string | 具体的评论情况查看后面的`comments` |
| comment_is_your_self |  评论是否自己的 | int |  该评论是自己发出的，自己评论了自己 |
| comment_anonymous |  评论是否匿名 | int |  如果匿名了且该评论不是你的，那么该评论者不会显示 |
| global_message_id |  管理员通知ID | int |  如果信息类型为管理员通知，该字段非空 |
| publish_again |  内容发布是否是再次发布的 | int |  如果信息类型为关注人发布了内容，该字段可以表明，该内容是否为再次发布的 |
| private_chanel |  私聊频道名 | string |  如果信息类型为私信，该字段为与其他人的频道名 |
| send_user_id |  私聊发送信息的用户 | int |  发送信息的人 |
| send_status |  私聊发送信息的状态 | int |  0表示正常，1表示被发件方删除了，不影响收件方 |
| send_message |  私聊的内容 | string |  私聊接收或者发送的内容 |
| message_type | 信息的类型 | int | 具体查看之前的介绍 |

`contents`和`comments`是信息所关联的内容和评论数据，参考评论接口说明。

`extra_users`和`extra_comments`是补充的信息，方便显示发件人的信息，还有具体的评论内容。

管理员接口会列出所有数据，不再关心评论是否匿名，内容是否被隐藏或违禁等。