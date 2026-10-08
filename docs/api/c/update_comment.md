# /content/update/comment

更新内容评论设置。

## 请求

```
POST /node/update/comment

{
	"id": 1,
	"close_comment": 1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| close_comment |内容评论设置 | int |  0表示开启评论，1表示关闭评论 | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "65cfe4a4fd194d0f9f6242c8f255088c"
}
```

内容不存在:

```
{
  "flag": false,
  "cid": "e7dae50a4706424ab09360113a7ceab4",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```