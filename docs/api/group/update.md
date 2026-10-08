# /group/update

更新用户组，管理员接口。

## 请求

```
POST /group/create

{
	"id": 1,
	"name": "",
	"describe": "test groDDDup",
	"image_path": ""
}
```


参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 组数据库唯一标记 | int | | 是 |
| name |    组名   |   string | 非空时更改 | |
| describe | 描述 | string | 非空时更改 | |
| image_path | 背景图、头像 | string | 非空时更改 | 否 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "7ab531c074974e359e1618ef8e7ac467",
  "data": {
    "id": 1,
    "name": "",
    "describe": "test groDDDup",
    "create_time": 0,
    "update_time": 1562492748,
    "image_path": ""
  }
}
```

用户组不存在：

```
{
  "flag": false,
  "cid": "babbb07edc654c5bbe80a6921ba4a32b",
  "error": {
    "id": 100041,
    "msg": "group not found"
  }
}
```

用户组头像或背景图不存在：

```
{
  "flag": false,
  "cid": "028867b1d02f41edba3036a0ef305bf1",
  "error": {
    "id": 100030,
    "msg": "file can not be found:image url not exist"
  }
}
```

组描述太长：

```
{
  "flag": false,
  "cid": "0202566c86c44ebdaece184f8c1a33a0",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'UpdateGroupRequest.Describe' Error:Field validation for 'Describe' failed on the 'lt' tag"
  }
}
```

组名被占用：

```
{
  "flag": false,
  "cid": "b63a8e39f4ce47dcbd468d2ace0f50c4",
  "error": {
    "id": 100040,
    "msg": "group name already be used"
  }
}
```