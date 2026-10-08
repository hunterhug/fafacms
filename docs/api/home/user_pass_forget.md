# /user/password/forget

忘记密码

## 请求

不需要前缀。

```
POST /user/password/forget

{
	"email": "your_email@example.com"
}
```

## 响应

正常：

```
{
  "flag": true,
  "cid": "ad2244297e604903a61964ced346b4bb"
}
```

修改密码验证码（**6 位数字**）将会发送到邮箱中。

> 验证码有效期五分钟。验证码未过期时重复请求**不再发信**，返回 100028（提示已发送，请查收邮箱）。
> 防邮箱轰炸：同一 IP 每小时前 3 次直接发送，之后必须通过图形验证码。

用户邮箱未找到：

```
{
  "flag": false,
  "cid": "ec36c1116b6d4bc684c9833dd94924e9",
  "error": {
    "id": 100027,
    "msg": "email not found"
  }
}
```

验证码已发送（未过期，不重复发信）：

```
{
  "flag": false,
  "cid": "faad0d5126794bcabd6433cdb4d2c0f8",
  "error": {
    "id": 100028,
    "msg": "reset code already sent, please retry later"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "c61529f7d0cf4fc3a8c6ad2665ae2e17",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ForgetPasswordRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
  }
}
```