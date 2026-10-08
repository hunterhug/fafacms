# /user/2fa/disable

关闭 2FA。需输入密码确认（丢失 2FA 码时请联系管理员重置）。

## 请求

需要前缀 `/v1`。

```
POST /v1/user/2fa/disable

{
	"password": "12345678"
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| password | 登录密码 | string | 是 |

## 响应

正常：

```
{
  "flag": true
}
```

密码错误：

```
{
  "flag": false,
  "error": { "id": 100020, "msg": "username or password wrong:password wrong" }
}
```
