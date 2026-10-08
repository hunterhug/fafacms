# /content/update/image

更新内容背景图。

## 请求

```
POST /content/update/image

{
	"id":1,
	"image_path":"/storage/admin/other/admin_7e59164ec6196145f0ad162b4da7d0738033555a109ca01d1dee635b8494cef1.jpeg"
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| image_path | 节点背景图 | string | 必须是通过上传接口上传的| 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

背景图不存在：

```
{
  "flag": false,
  "cid": "63e5ca0689bf417299e1512713d68b81",
  "error": {
    "id": 100030,
    "msg": "file can not be found"
  }
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
  "cid": "b4c12c7bc4ff4624a8d7e8b0194c5943",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateImageOfContentRequest.ImagePath' Error:Field validation for 'ImagePath' failed on the 'required' tag"
  }
}
```