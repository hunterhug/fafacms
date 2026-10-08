# /content/update/top

更新内容置顶。

## 请求

```
POST /node/update/top

{
	"id": 1,
	"top":1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| top |内容是否置顶 | int | 1表示置顶，默认0 | 是 |


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

参数不对：

```
{
  "flag": false,
  "cid": "4c2c2f01b4de41f7b91c0777c3e1e41d",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateTopOfContentRequest.Top' Error:Field validation for 'Top' failed on the 'oneof' tag"
  }
}
```