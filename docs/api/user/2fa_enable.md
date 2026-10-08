# /user/2fa/enable

绑定（开启）2FA。先调用 `/user/2fa/secret` 生成并暂存秘钥，再用验证器算出的动态码确认绑定。

## 请求

需要前缀 `/v1`。

```
POST /v1/user/2fa/enable

{
	"code": "123456"
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| code | 验证器动态码 | string | 是 |

## 响应

正常：

```
{
  "flag": true
}
```

动态码错误：

```
{
  "flag": false,
  "error": { "id": 100037, "msg": "two factor auth code wrong" }
}
```

秘钥暂存已过期：

```
{
  "flag": false,
  "error": { "id": 100038, "msg": "two factor auth session expired" }
}
```
