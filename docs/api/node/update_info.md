# /node/update/info

更新节点名称或描述。

## 请求

```
POST /node/update/info

{
	"id":1,
	"name":"1234567",
	"describe":"ssssssssssss"
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 节点ID | int | | 是 |
| name | 节点名称 | string | | 否 |
| describe | 节点描述 | string | | 是 |


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