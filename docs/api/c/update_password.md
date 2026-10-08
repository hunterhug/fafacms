# /content/update/password

更新内容密码。内容如果存在密码时，前端查看内容时需输入密码。

## 请求

```
POST /node/update/password

{
	"id": 1,
	"password": "ssss"
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| password |内容密码 | string | 留空表示不需要密码 | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "65cfe4a4fd194d0f9f6242c8f255088c"
}
```

内容不存在:

```
{
  "flag": false,
  "cid": "e7dae50a4706424ab09360113a7ceab4",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```