# /group/create

创建用户组，管理员接口。

## 请求

```
POST /group/create

{
	"name": "groupname",
	"describe": "test grouffpcccccccccccc",
	"image_path": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg"
}
```

用户组名全局唯一，`image_path` 背景图或者头像，可以留空，非空时必须存在。

## 响应

正常：

```
{
  "flag": true,
  "cid": "c8de1af25dbd49858355ee71e36a639f",
  "data": {
    "id": 1,
    "name": "groupname",
    "describe": "test grouffpcccccccccccc",
    "create_time": 1562492308,
    "image_path": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg"
  }
}
```

用户组名被占用：

```
{
  "flag": false,
  "cid": "9629a5b9df1d4a55bdc339bee08778ec",
  "error": {
    "id": 100040,
    "msg": "group name already be used"
  }
}
```

用户组头像或背景图不存在：

```
{
  "flag": false,
  "cid": "2f8d01b4e45b4cd6898171aa9f6c9b65",
  "error": {
    "id": 100030,
    "msg": "file can not be found"
  }
}
```

参数有误：

```
{
  "flag": false,
  "cid": "a3255228d8a54655afc65bdca5c217ad",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'CreateGroupRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"
  }
}
```