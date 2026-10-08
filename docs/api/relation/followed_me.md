# /relation/followed/me

列出关注你的人。只有登录用户可用。

## 请求

```
POST /relation/followed/me

{
	"user_a_id": 0,
	"user_a_name": "",
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

此时 `B` 是我，如果 `A` 的用户信息不填，那么列出关注我的所有人。如果 `A` 信息填了，表示查找是否有 `A->B` 的关系。

## 响应

正常：

```
{
  "flag": true,
  "cid": "b37023c3286a43e2ada946b95c4595df",
  "data": {
    "relations": [
      {
        "id": 5,
        "create_time": 1569571951,
        "user_a_id": 31,
        "user_b_id": 29,
        "user_a_name": "test3",
        "user_b_name": "test1",
        "is_both": false,
        "is_following": false
      },
      {
        "id": 4,
        "create_time": 1569571947,
        "user_a_id": 30,
        "user_b_id": 29,
        "user_a_name": "test2",
        "user_b_name": "test1",
        "is_both": true,
        "is_following": false
      },
      {
        "id": 2,
        "create_time": 1569567088,
        "user_a_id": 29,
        "user_b_id": 29,
        "user_a_name": "test1",
        "user_b_name": "test1",
        "is_both": true,
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

`relations`为关系，表示 `A` 关注了 `B`，此时 `B` 都是我。上面表示`鸡鸡1`，`鸡鸡2`和`鸡鸡3`都关注了我。`is_both`为`true`表示互相关注，如果自己关注自己也是可以的。`is_following`在此无意义。

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