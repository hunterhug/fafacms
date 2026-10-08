# /captcha

生成图形验证码（数字，PNG base64）。用于登录/注册防爆破：正常用户无感，连续失败或同 IP 频繁注册时才要求验证码。

## 请求

```
POST /captcha

{}
```

## 响应

```
{
  "flag": true,
  "cid": "xxxx",
  "data": {
    "captcha_id": "yI1HRmuLsLaA7vJ68vhX",
    "image": "data:image/png;base64,iVBORw0KGgo..."
  }
}
```

| 字段 | 含义 |
|------|------|
| captcha_id | 验证码 id，登录/注册时随 `captcha_code` 一起提交 |
| image | base64 图片（可直接用于 `<img src>`） |

## 校验时机（无感触发）

- **登录**：同一 `IP+用户名` 连续失败 5 次后，登录接口返回错误码 `100031`（need captcha），前端此时渲染验证码；达到 10 次后返回 `100033`（login try too many，临时锁定）。
- **注册**：同一 IP 1 小时内注册超过 3 次后，注册接口返回 `100031`，前端此时渲染验证码。

验证码错误返回 `100032`（captcha wrong）。验证码有效期约 10 分钟（服务端内存）。
