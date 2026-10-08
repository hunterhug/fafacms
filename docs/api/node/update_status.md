# /node/update/status

更新节点状态。设置是否隐藏。

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
| id | 节点ID | int | | 是 |
| status | 节点状态 | int | 0表示正常，1表示隐藏| 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

节点不存在:

```
{
  "flag": false,
  "cid": "3c7f9f8a20d148fcb04fb0d8c31ad8fd",
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
  "cid": "259bd3df0b93404db2be4d755c4841ed",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateStatusOfNodeRequest.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"
  }
}
```