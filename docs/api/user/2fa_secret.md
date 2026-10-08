# /user/2fa/secret

生成并暂存 2FA 秘钥，返回 `secret` + `otpauth://` URI（仅未开启时）。前端用 URI 生成二维码。

## 请求

需要前缀 `/v1`。

```
POST /v1/user/2fa/secret

{}
```

## 响应

正常：

```
{
  "flag": true,
  "data": {
    "secret": "BASE32SECRET",
    "uri": "otpauth://totp/FaFaCMS:xxx?secret=xxx&issuer=FaFaCMS"
  }
}
```

已开启 2FA 时：

```
{
  "flag": false,
  "error": { "id": 100036, "msg": "two factor auth code needed:already enabled" }
}
```
