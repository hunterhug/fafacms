# /node/update/parent

更新节点的父亲。

## 请求

```
POST /node/update/parent

{
	"id": 1,
	"to_be_root": true,
	"parent_node_id": 0
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 节点ID | int | | 是 |
| to_be_root | 是否成为顶层节点 | bool | 当true时忽视parent_node_id，直接成为顶层| 否 |
| parent_node_id | 父亲节点ID | int | | 是 |

有儿子的节点不能成为其他人的儿子，因为只设计了两层节点。节点父亲改变后，会自动进行节点排序，该节点会在同一层中排序最大，即排在最后。

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


有儿子了不能成为别人的儿子：

```
{
  "flag": false,
  "cid": "63e15a9b1cdc4471a1eb07dc2e4a13d2",
  "error": {
    "id": 101004,
    "msg": "content node has children:has child"
  }
}
```

父亲节点不存在：

```
{
  "flag": false,
  "cid": "84beef38a27146e5a00d1a153f8e0c49",
  "error": {
    "id": 101002,
    "msg": "parent content node not found"
  }
}
```

参数不对：


父亲不能是自己：

```
{
  "flag": false,
  "cid": "ab63215293914a80b805f5681733ff05",
  "error": {
    "id": 100010,
    "msg": "paras input not right:self can not be parent"
  }
}
```