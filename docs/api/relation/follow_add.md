# /relation/follow/add

添加关注。只有登录用户可用。

## 请求

```
POST /relation/follow/add

{
	"user_id": 0,
	"user_name": "test2"
}
```

表示我要关注用户：`user_name=test2`。

## 响应

关注成功：

```
{
  "flag": true,
  "cid": "0eee9432306c4434aaf3bcac596ea5c0"
}
```

关注的用户不存在：

```
{
  "flag": false,
  "cid": "b594ad1e8cd64463ae9c65b4ca9e0a9e",
  "error": {
    "id": 100003,
    "msg": "user not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "d20a05405ca347afac0138add43691ab",
  "error": {
    "id": 100010,
    "msg": "paras input not right:user info empty"
  }
}
```