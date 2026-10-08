# /relation/followed/list

列出关注B用户的人，并且查找你和这些人们的关系。

## 请求

```
POST /relation/followed/list

{
	"user_b_id": 0,
	"user_b_name": "test2",
	"limit": 5,
	"page": 1,
	"sort": [
		"=id",
		"=user_a_id",
		"=user_b_id",
		"-create_time"
	]
}
```

此时 `B` 用户的信息必填，表示查找关注 `B` 用户的人们。

## 响应

正常：

```
{
  "flag": true,
  "cid": "56f6d064c4ac4fe983391e2dcb7c26ba",
  "data": {
    "relations": [
      {
        "id": 6,
        "create_time": 1569572331,
        "user_a_id": 31,
        "user_b_id": 30,
        "user_a_name": "test3",
        "user_b_name": "test2",
        "is_both": false,
        "is_following": false
      },
      {
        "id": 1,
        "create_time": 1569565571,
        "user_a_id": 29,
        "user_b_id": 30,
        "user_a_name": "test1",
        "user_b_name": "test2",
        "is_both": true,
        "is_following": true
      }
    ],
    "users": {
      "29": {
        "id": 29,
        "name": "test1",
        "nick_name": "鸡鸡1",
        "head_photo": "",
        "is_vip": false,
        "short_describe": "",
        "gender": 1
      },
      "30": {
        "id": 30,
        "name": "test2",
        "nick_name": "鸡鸡2",
        "head_photo": "",
        "is_vip": false,
        "short_describe": "",
        "gender": 1
      },
      "31": {
        "id": 31,
        "name": "test3",
        "nick_name": "鸡鸡3",
        "head_photo": "",
        "is_vip": false,
        "short_describe": "",
        "gender": 1
      }
    },
    "limit": 5,
    "page": 1,
    "total_pages": 1
  }
}
```

参数解释：

`relations`为关系，表示 `A` 关注了 `B`，此时 `B` 都是用户B。上面表示`鸡鸡1`和`鸡鸡3`都关注了用户B。

此时`is_following`的意义是，我是否关注了`关注用户B的用户A`，`is_both`表示`关注用户B的用户A`是否和我互相关注。

`users`为用户补充信息。

参数不对：

```
{
  "flag": false,
  "cid": "f40a792f6bf24345a2e9b0212d39bb65",
  "error": {
    "id": 100010,
    "msg": "paras input not right:user B info empty"
  }
}
```

没有任何关注的人：

```
{
  "flag": true,
  "cid": "f5de27bf579243bfbc3c5eafa2d40b16",
  "data": {
    "relations": [],
    "users": {},
    "limit": 5,
    "page": 1,
    "total_pages": 0
  }
}
```