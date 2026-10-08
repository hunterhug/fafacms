# /resource/assign

分配资源给组，解除组的资源绑定，管理员接口。

## 请求

```
POST /resource/assign

{
	"group_id": 1,
	"resource_release": 0,
	"resources": [
		4,
		5
	]
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| group_id | 组ID | int | 将资源绑定到该组 | 否 |
| resource_release |    是否解除资源与组的绑定   |   int | 1表示解除绑定| 是|
| resources | 资源ID数组 | []int | 操作的资源接口们 | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "d8786850ede44129825929b588a92d62"
}
```


组不存在：

```
{
  "flag": false,
  "cid": "149f0a6a9c2547138c7eb56965c1536e",
  "error": {
    "id": 100041,
    "msg": "group not found"
  }
}
```

资源数组为空：

```
{
  "flag": false,
  "cid": "95119cf1588f44d18e2e3568bfe71f28",
  "error": {
    "id": 100010,
    "msg": "paras input not right:resources empty"
  }
}
```

某些资源ID不存在：

```
{
  "flag": false,
  "cid": "921f44de453f43158556a3d727137323",
  "error": {
    "id": 100050,
    "msg": "resource count not right:resource wrong:4!=5"
  }
}
```