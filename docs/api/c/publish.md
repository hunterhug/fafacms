# /content/publish

发布内容。

## 请求

```
POST /content/publish

{
	"id": 10
}
```

可使用`内容id`发布内容。发布内容时，如果服务开启了历史记录设置，会将预发布的内容存储在历史表中，历史表状态码为`1`，表示已发布历史。

发布内容时，会将预发布的缓冲区刷进正式字段，并且置`pre_flush`为1，表示已经刷新，而字段`version`将为加1，表示发布的版本号，第一次发布后是1，第二次发布是2，以此类推。

`pre_flush`已经为1的内容不做任何操作，因为没有任何东西可以发布。

## 响应

正常：

```
{
  "flag": true,
  "cid": "d69021939a0f46b080bf63634bc00417"
}
```

内容不存在：

```
{
  "flag": false,
  "cid": "d723af3927f94b7486f94d862ec1839c",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "52d4b77ae16348eab83fcae81a22bab0",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'TakeContentRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```