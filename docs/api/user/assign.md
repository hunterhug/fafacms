# /user/assign

批量给用户绑定组，解除组，管理员接口。

## 请求

```
POST /user/assign

{
	"group_id": 1,
	"group_release": 0,
	"users": [
		19,
		18
	]
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| group_id | 组ID | int | 将用户绑定到该组，每个用户只能属于一个组 | 否 |
| group_release |    是否解除组   |   int | 1表示解除组，此时group_id无效| 是|
| users | 用户ID数组 | []int | 操作的用户们 | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "61a45c17e90a4e88ba58b3a5af583e4a",
  "data": 0
}
```

`data` 表示数据库修改成功的数量，业务端可忽视。


组不存在：

```
{
  "flag": false,
  "cid": "11c50f6276b344a1827bd2dae8206850",
  "error": {
    "id": 100041,
    "msg": "group not found"
  }
}
```

用户组为空：

```
{
  "flag": false,
  "cid": "f69cf7a6613d40b3a6235fc94e519516",
  "error": {
    "id": 100010,
    "msg": "paras input not right:users empty"
  }
}
```