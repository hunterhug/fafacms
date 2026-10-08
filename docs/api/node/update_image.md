# /node/update/image

更新节点背景图。

## 请求

```
POST /node/update/image

{
	"id":1,
	"image_path":"/storage/admin/other/admin_7e59164ec6196145f0ad162b4da7d0738033555a109ca01d1dee635b8494cef1.jpeg"
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 节点ID | int | | 是 |
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
  "cid": "fbd6fbdda07b4d15a7397a892ceabb78",
  "error": {
    "id": 100030,
    "msg": "file can not be found"
  }
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
  "cid": "ec644f3d452d4d6d83f1433db4fdc11d",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateImageOfNodeRequest.ImagePath' Error:Field validation for 'ImagePath' failed on the 'required' tag"
  }
}
```