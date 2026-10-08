# /captcha/unblock

限流解封。被临时拉黑（限流触发）的 IP，通过验证码后解除拉黑。

## 请求

不需要前缀。

```
POST /captcha/unblock

{
	"captcha_id": "yI1HRmuLsLaA7vJ68vhX",
	"captcha_code": "1234"
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| captcha_id | 验证码 id（由 /captcha 获取） | string | 是 |
| captcha_code | 验证码答案 | string | 是 |

## 响应

正常：

```
{
  "flag": true
}
```

验证码缺失：

```
{
  "flag": false,
  "error": { "id": 100031, "msg": "need captcha" }
}
```

验证码错误：

```
{
  "flag": false,
  "error": { "id": 100032, "msg": "captcha wrong" }
}
```
