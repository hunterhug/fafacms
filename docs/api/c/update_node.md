# /content/update/node

更新内容节点。内容更改节点后，会变成该节点下面排最后的内容。

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
| status | 内容状态 | int | 0表示正常，1表示隐藏，2表示内容违禁拉黑，3表示送到垃圾箱 | 是 |


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
  "cid": "e7dae50a4706424ab09360113a7ceab4",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

内容节点不存在：

```
{
  "flag": false,
  "cid": "8c4b0c3dbf5f4366bcd9b2937808f7ad",
  "error": {
    "id": 101001,
    "msg": "content node not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "6a50f1f123c9429dbb5272d1ce60b433",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateNodesOfContentRequest.NodeId' Error:Field validation for 'NodeId' failed on the 'required' tag"
  }
}
```