# /content/recycle

内容垃圾回收，从垃圾箱中回收内容，内容会变成正常。

## 请求

```
POST /content/recycle

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

参数不对：

```
{
  "flag": false,
  "cid": "7b4822ef83ad4f00bcc0e01a7002c1ea",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ReCycleOfContentInRubbishRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```