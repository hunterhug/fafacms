# /user/activate/code

用户重发激活验证码

## 请求

不需要前缀。

```
POST /user/activate/code

{
	"email": "your_email@example.com"
}
```

## 响应

正常：

```
{
  "flag": true,
  "cid": "b233c451c7cf4462a09da368f7c35a09"
}
```

激活验证码将会再次发送到邮箱中。

激活邮箱未找到：

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

激活码未过期：

```
{
  "flag": false,
  "cid": "8e6b8270dd1446bea322f8ef485f222b",
  "error": {
    "id": 100026,
    "msg": "activate code not expired"
  }
}
```

激活码过期后，才能重新请求激活码，激活码有效期五分钟。

参数不对：

```
{
  "flag": false,
  "cid": "5a42530d1bc74f2799230f68a45bbbf6",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ResendActivateCodeToUserRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
  }
}
```