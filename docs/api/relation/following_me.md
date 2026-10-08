# /relation/following/me

列出你关注的人。只有登录用户可用。

## 请求

```
POST /relation/following/me

{
	"user_b_id": 0,
	"user_b_name": "",
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

此时 `A` 是我，如果 `B` 的用户信息不填，那么列出我关注的所有人。如果 `B` 信息填了，表示查找是否有 `A->B` 的关系。

## 响应

正常：

```
{
  "flag": true,
  "cid": "50e17b3b5b8049d6aa9204117798e2a0",
  "data": {
    "relations": [
      {
        "id": 2,
        "create_time": 1569567088,
        "user_a_id": 29,
        "user_b_id": 29,
        "user_a_name": "test1",
        "user_b_name": "test1",
        "is_both": true,
        "is_following": false
      },
      {
        "id": 1,
        "create_time": 1569565571,
        "user_a_id": 29,
        "user_b_id": 30,
        "user_a_name": "test1",
        "user_b_name": "test2",
        "is_both": false,
        "is_following": false
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

`relations`为关系，表示 `A` 关注了 `B`，此时 `A` 都是我。上面表示我关注了`鸡鸡1`和`鸡鸡2`。`is_both`为`true`表示互相关注，如果自己关注自己也是可以的。`is_following`在此无意义。

`users`为用户补充信息。

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