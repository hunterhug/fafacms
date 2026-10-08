# /group/resource/list

列出用户组下的资源，管理员接口。

## 请求

```
POST /group/resource/list

{
	"group_id": 1
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| group_id | 用户组ID | int |  | 是 |

## 响应

```
{
  "flag": true,
  "cid": "0cf7a59135cb4556a9d5d6e60f9ac1a2",
  "data": {
    "resources": [
      19,
      18,
      4,
      5
    ]
  }
}
```

该接口没有分页功能。