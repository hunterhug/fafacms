# /group/take

获取组信息，管理员接口。

## 请求

```
POST /group/take

{
	"id": 0,
	"name":"groupname"
}
```

可使用`组名`或者`组id`获取组信息。

## 响应

正常：

```
{
  "flag": true,
  "cid": "0fb98f1f7f064205a1456af4a83fe4f9",
  "data": {
    "id": 1,
    "name": "groupname",
    "describe": "test groDDDup",
    "create_time": 1562492308,
    "update_time": 1562492748,
    "image_path": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg"
  }
}
```

组不存在：

```
{
  "flag": false,
  "cid": "083f2bfa10a442d3960da03599aba182",
  "error": {
    "id": 100041,
    "msg": "group not found"
  }
}
```