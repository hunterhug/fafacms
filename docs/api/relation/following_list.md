# /relation/following/list

列出A用户关注的人，并且查找你和这些人们的关系。

## 请求

```
POST /relation/following/list

{
	"user_a_id": 0,
	"user_a_name": "test2",
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

此时 `A` 用户的信息必填，表示查找 `A` 用户关注的人们。

## 响应

正常：

```
{
  "flag": true,
  "cid": "bd548403fc1d436581e1a2d1ce882acd",
  "data": {
    "relations": [
      {
        "id": 4,
        "create_time": 1569571947,
        "user_a_id": 30,
        "user_b_id": 29,
        "user_a_name": "test2",
        "user_b_name": "test1",
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
      }
    },
    "limit": 5,
    "page": 1,
    "total_pages": 1
  }
}
```

参数解释：

`relations`为关系，表示 `A` 关注了 `B`，此时 `A` 都是用户A。上面表示用户A关注了`鸡鸡1`。

此时`is_following`的意义是，我是否关注了`用户A关注的用户B`，`is_both`表示`用户A关注的用户B`是否和我互相关注。

`users`为用户补充信息。

参数不对：

```
{
  "flag": false,
  "cid": "f40a792f6bf24345a2e9b0212d39bb65",
  "error": {
    "id": 100010,
    "msg": "paras input not right:user A info empty"
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