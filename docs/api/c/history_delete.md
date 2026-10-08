# /content/history/delete

删除内容历史。只能删除自己的历史内容。

## 请求

```
POST /content/history/delete

{
	"id": 1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 历史内容ID | int | | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

历史内容不存在:

```
{
  "flag": false,
  "cid": "bd8a95e81c2a46c4b75cc52eacdaf055",
  "error": {
    "id": 110006,
    "msg": "content history can not found"
  }
}
```

不能删除内容：

```
{
  "flag": false,
  "cid": "4a6a3f70b7fc42659d7006cd83a1e56c",
  "error": {
    "id": 110007,
    "msg": "content can not delete for content not in rubbish"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "a591ca7626c84897b3e2ebf8f8c91d0f",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ReallyDeleteContentRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```