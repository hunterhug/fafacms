# /message/private/send

发送私信。

## 请求

```
POST /message/private/send

{
	"user_id": 2,
	"message": "eeeeeeeeeeeeeeee"
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id |    私信的目标用户  |   int | | 是 |
| message |    私信的内容 |   string | | 是 |


登录的用户，向某某发消息。

## 响应

正常：

```
{
  "flag": true,
  "cid": "cf20e1be6de445248c1a92501955c91c"
}
```

用户不存在：

```
{
  "flag": false,
  "cid": "f89cfd87f1244350b9cc8d3f10f784f3",
  "error": {
    "id": 100003,
    "msg": "user not found"
  }
}
```