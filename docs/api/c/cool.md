# /content/cool

点赞内容。只有登录的用户可以点赞。只能对已经发布的正常显示的内容进行点赞。

## 请求

```
POST /content/cool

{
	"id": 10
}
```

## 响应

点赞正常：

```
{
  "flag": true,
  "cid": "2e5b753d299c4d54b6b3b44382c6e36b",
  "data": "+"
}
```

取消点赞正常：

```
{
  "flag": true,
  "cid": "791152136ecd4e79888060431f0c6a81",
  "data": "-"
}
```

第一次点赞时，返回`+`，再一次返回`-`表示取消点赞。

内容不存在：

```
{
  "flag": false,
  "cid": "9a523e098ee749bb8ec3135e3c8ac56c",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

内容被违禁：

```
{
  "flag": false,
  "cid": "7d3b11515f654c4a81a85f1dd0ad28c5",
  "error": {
    "id": 110002,
    "msg": "content ban permit"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "7cd02a7d5db047d8870742231cb208f1",
  "error": {
    "id": 100010,
    "msg": "paras input not right:content_id empty"
  }
}
```