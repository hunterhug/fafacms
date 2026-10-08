# /user/admin/2fa/reset

管理员重置用户 2FA（用户丢失 2FA 码时联系管理员处理）。

## 请求

需要前缀 `/v1`，且为管理员接口。

```
POST /v1/user/admin/2fa/reset

{
	"id": 10
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| id | 用户 ID | int | 是 |

## 响应

正常：

```
{
  "flag": true
}
```

参数缺失：

```
{
  "flag": false,
  "error": { "id": 100010, "msg": "paras input not right:id empty" }
}
```
