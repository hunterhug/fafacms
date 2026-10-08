# /relation/following/list

列出关系。管理员接口，列出所有人的关系。

## 请求

```
POST /relation/following/list

{
	"user_a_id": 0,
	"user_a_name": "",
	"user_b_id": 0,
	"user_b_name": "",
	"create_time_begin": 0,
	"create_time_end": 0,
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


可以查找谁关注了谁。如果 `A` 和 `B` 用户信息都为空，表示不筛选。`A` 非空，表示查找他关注了谁，`B` 非空，表示查找谁关注了他。

## 响应

正常：

```
{
  "flag": true,
  "cid": "497e64783dd740e4b0ec7212b37c01ed",
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
      },
      {
        "id": 1,
        "create_time": 1569565571,
        "user_a_id": 29,
        "user_b_id": 30,
        "user_a_name": "test1",
        "user_b_name": "test2",
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

`relations`为关系，表示 `A` 关注了 `B`。

此时`is_following`没有意义。`is_both`表示是否互相关注。

`users`为用户补充信息。

找不到任何关系：

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