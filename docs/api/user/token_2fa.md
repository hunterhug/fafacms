# /user/token/2fa

两步验证登录。当用户已开启 2FA 且密码校验通过时，`/user/token/get` 返回 `100036` + `data.pending`（一次性待确认令牌，5 分钟）；再请求本接口校验动态码，通过后发放真实 token。

## 请求

不需要前缀。

```
POST /user/token/2fa

{
	"pending": "xxx",
	"code": "123456"
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| pending | 待确认令牌（/user/token/get 返回） | string | 是 |
| code | 验证器动态码 | string | 是 |

## 响应

正常：

```
{
  "flag": true,
  "data": "1_xxx"
}
```

动态码错误：

```
{
  "flag": false,
  "error": { "id": 100037, "msg": "two factor auth code wrong" }
}
```

会话过期：

```
{
  "flag": false,
  "error": { "id": 100038, "msg": "two factor auth session expired" }
}
```
