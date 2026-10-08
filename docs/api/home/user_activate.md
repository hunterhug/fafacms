# /user/activate

用户激活

## 请求

不需要前缀。

```
POST /user/activate

{
	"code": "260479",
	"email": "your_email@example.com"
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| code | 激活码 | string |  6 位数字，邮箱中查看，有效期五分钟 | 是 |
| email |    邮箱 | string | 全局唯一，用来登录 | 是 |

## 响应


正常：

```
{
  "flag": true,
  "cid": "08bc37539ebb4433b5067e9aeb6b52a3",
  "data": "8_8c600476737a4b5a9adc61ff58a0aa97"
}
```

激活成功后返回了 `token` 为 `8_8c600476737a4b5a9adc61ff58a0aa97`。

激活码错误（剩余次数随错误返回，错满 5 次激活码作废，需重新发送）：

```
{
  "flag": false,
  "cid": "a5ed93969eb7420485bbbb092e7af532",
  "error": {
    "id": 100024,
    "msg": "activate code wrong:4 attempts left"
  }
}
```

激活码已过期（或已因错满 5 次作废）：

```
{
  "flag": false,
  "cid": "8f2d8a01a8c346989c42d5b2f2633ceb",
  "error": {
    "id": 100025,
    "msg": "activate code expired"
  }
}
```

激活码过期或作废时，可请求重发激活码API申请新的激活码。

参数不对：

```
{
  "flag": false,
  "cid": "5a1819a474f64c96aeffbbea4a256cc8",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ActivateUserRequest.Code' Error:Field validation for 'Code' failed on the 'required' tag"
  }
}
```