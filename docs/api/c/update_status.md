# /content/update/status

更新内容状态。设置是否隐藏。

## 请求

```
POST /node/update/status

{
	"id":1,
	"status":0
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| status | 内容状态 | int | 0表示正常，1表示隐藏，不能是其他值| 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

内容不存在:

```
{
  "flag": false,
  "cid": "d65f8089f6674713981eb461a91156d2",
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
  "cid": "950338d3559a40ec81c471371b9973b3",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateStatusOfContentRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"
  }
}
```