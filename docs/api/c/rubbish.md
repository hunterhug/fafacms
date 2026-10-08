# /content/rubbish

将内容送到垃圾箱（软删除 status=3）。可调用回收接口从垃圾箱中恢复。在具体业务中，你可以把这个当成删除，类似知乎，然后可以撤销删除。彻底删除走 `/content/delete`（也是软删除 status=4，仅用户世界消失、管理员可见）。

只有VIP用户才可以将内容送到垃圾箱。

## 请求

```
POST /content/rubbish

{
	"id": 1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |


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

不是VIP：

```
{
  "flag": false,
  "cid": "443634cfa5664fb48c365a670b9a65cd",
  "error": {
    "id": 99996,
    "msg": "you are not vip"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "f8c4a38c9c4048d28e4ceb5174987e64",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'SentContentToRubbishRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```