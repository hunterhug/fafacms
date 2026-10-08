# /node/delete

删除节点

## 请求

```
POST /node/delete

{
	"id":1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 节点ID | int | | 是 |

节点有儿子或者有内容存在不能删除，删除节点后，节点会重新自动排序。

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

节点下有儿子：

```
{
  "flag": false,
  "cid": "9cdc4e67414842f28273d6d4fcfdc184",
  "error": {
    "id": 101004,
    "msg": "content node has children"
  }
}
```

节点下存在内容：

```
{
  "flag": false,
  "cid": "9cdc4e67414842f28273d6d4fcfdc184",
  "error": {
    "id": 101005,
    "msg": "content node has content can not delete"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "63302b846d9540709dbf0abb77fdfcfe",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'DeleteNodeRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```