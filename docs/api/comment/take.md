# /comment/take

获取评论。任何人的评论都可以获取。

## 请求

```
POST  /comment/take

{
	"id": 3
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
  "cid": "77c144c72d984839addb61cc3dc7bf51",
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
          "ban_time": 0,          
          "is_anonymous": true,
          "user_id": 26,
          "cool": 0,
          "bad": 0
        },
        "2": {
          "id": 2,
          "describe": "我不愿意你说：大哥，我第一个跟随你",
          "create_time": 1569049669,
          "is_delete": false,
          "is_ban": false,
          "ban_time": 0,
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
    }
  }
}
```

关联数据`contents`和`comment`找不到这种情况不可能发生，获取关联数据`comment`的具体内容时请先判断`is_delete`字段。

参数说明：

| 字段   |      含义   |类型  |   详情 |
|----------|--------|------|------|
| comment | 这条评论的主体| struct |  |
| comment.id | 评论的ID| int | 具体评论信息查找 extra.comments  |
| comment.content_id | 评论所在的内容ID| int | 相关信息查找 extra.contents |
| comment.content_title | 评论时内容的标题，内容标题可能变动 | string | 相关信息查找 extra.contents  |
| comment.comment_type | 评论的类型| int | 0表示对内容评论，1表示对第一级评论的评论，2表示其他评论的评论 |
| comment.root_comment_id | 被评论的评论ID(超过一级时为底部评论)| int | comment.comment_type为1和2有效，具体评论信息查找 extra.comments  |
| comment.comment_id | 被评论的评论ID| int | comment.comment_type为2有效，具体评论信息查找 extra.comments  |
| extra | 评论的补充信息| struct |  |
| extra.contents | 评论所在的内容详情 | struct | 如果查找不到，表示内容不存在或对调用用户不可见（被隐藏 status=1、违禁 status=2、移到垃圾 status=3、用户彻底删除 status=4） |
| extra.comments | 具体的评论详情 | struct | 如果查找不到，表示评论不存在或对调用用户不可见（评论为逻辑删除 is_delete，不物理删除） |
| extra.comments.is_anonymous | 具体的评论是否匿名 | bool | 如果是匿名，那么不会查找出其用户信息，但是，如果评论属于该API的调用用户，那么会查找，供业务端实现：匿名（你自己） |
| extra.comments.is_delete | 具体的评论是否被删除 | bool | 如果被删除，那么不会查找出该评论具体描述describe，供业务端实现： 评论已删除|
| extra.comments.is_ban | 具体的评论是否违禁 | bool | 如果被违禁，那么不会查找出该评论具体描述describe，供业务端实现： 评论已违禁，但如果评论属于该API的调用用户，那么查找，业务端实现： 辣鸡你王八蛋（评论已违禁）|
| extra.users | 用户信息| struct | 如果查找不到，表示用户不存在 |


评论效果如下：

在我的角度（小花花是我）：

````
标题：今天放学后我做了一件好事

小花花：我也不愿意 // @小花花：我不愿意你说：大哥，我第一个跟随

@匿名(小花花)：
    大哥，我第一个跟随你
````

在其他人的角度：

```
标题：今天放学后我做了一件好事

小花花：我也不愿意 // @小花花：我不愿意你说：大哥，我第一个跟随

@匿名：
    大哥，我第一个跟随你
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
