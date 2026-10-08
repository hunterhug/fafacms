# 授权规范

## Token 授权

除前端接口（`/u`、`/content` 等 HomeRouter）与静态资源外，其他接口都需要授权。

- 登录（`/user/token/get`）后返回 `token`（形如 `1_xxxx`）。
- 后续请求在 HTTP 头携带：`Auth: <token>`。
- token 默认 7 天有效（Redis 存储，`SessionExpireTime`）。

## 请求签名（防重放，默认强制）

所有 API 请求（静态资源 `/storage*` 与 `/ping` 除外）默认必须携带签名头：

| 头 | 含义 |
|----|------|
| `X-Ts` | 请求时间戳（Unix 秒） |
| `X-Nonce` | 随机字符串（一次性，防重放） |
| `X-Sign` | `HMAC-SHA256(secret, "<X-Ts>:<X-Nonce>")` 的十六进制小写 |

- 服务端校验：时间戳在 ±5 分钟内、nonce 未被使用过、签名一致。
- 默认 `secret = fafacms-sign-v1-2026`，可用环境变量 `FAFA_SIGN_SECRET` 覆盖（前后端保持一致）。
- 默认**强制**校验（未带签名头返回 `403` + 错误码 `100035`）；设置环境变量 `FAFA_SIGN_STRICT=0` 可关闭强制（放行无签名请求，用于脚本/调试）。
- 说明：纯前端 SPA 中 secret 会打包进 JS，签名只能抬高逆向门槛，需配合限流/验证码/HTTPS 使用。

前端签名示例（crypto-js，纯 JS，HTTP 环境也可用）：

```js
const sign = CryptoJS.HmacSHA256(`${ts}:${nonce}`, 'fafacms-sign-v1-2026').toString(CryptoJS.enc.Hex)
```

## 验证码（登录/注册/评论防刷）

登录/注册/评论在触发风控时要求验证码（无感触发）：

- 登录：同一 `IP+用户名` 连续失败 5 次后返回 `100031`（need captcha），此时请求 `/captcha` 获取验证码，随登录请求提交 `captcha_id + captcha_code`；失败 10 次返回 `100033`（临时锁定）。
- 注册：同一 IP 1 小时内注册超过 3 次后返回 `100031`。
- 评论：同一用户 60 秒内评论超过 5 次后返回 `100031`（防刷评论）。
- 验证码接口：`/captcha`，返回 `captcha_id + image(base64 png)`；验证码错误返回 `100032`。

## 限流与临时拉黑

- 全局限流：每 IP 每分钟 240 次请求，超限返回 `429` + `Retry-After` 头。
- 连续超限：该 IP 被临时拉黑（5 分钟起，递增至 1 小时）。
- 豁免：`/ping`、`/captcha`、`/captcha/unblock`（保证拉黑时也能取验证码并解封）。
- **解封**：被拉黑后前端弹出「人机验证」，请求 `/captcha` 取验证码，再请求 `/captcha/unblock`（`captcha_id + captcha_code`）校验通过即解除该 IP 拉黑。

## 两步验证（2FA / TOTP）

标准 TOTP（RFC 6238，HMAC-SHA1，6 位动态码，30 秒，±30 秒容差），兼容微软 Authenticator / Google Authenticator。

- **开启**：`/v1/user/2fa/secret`（未开启时生成并暂存秘钥，返回 `secret + uri`，前端用 `otpauth://` URI 生成二维码）；`/v1/user/2fa/enable`（`code`，用验证器算出的动态码确认绑定）。
- **登录**：密码通过且已开启 2FA 时，`/user/token/get` 返回 `100036` + `data.pending`（一次性待确认令牌，5 分钟）；再请求 `/user/token/2fa`（`pending + code`）校验通过后发放真实 token。
- **关闭**：`/v1/user/2fa/disable`（`password` 确认）。
- **管理员重置**：`/v1/user/admin/2fa/reset`（`id`）——用户丢失 2FA 码时联系管理员处理。
- 错误码：`100036` 需两步验证、`100037` 动态码错误、`100038` 两步验证会话过期。

## 忘记密码 / 改密防爆破

- 忘记密码 `POST /user/password/forget`：同一 IP 1 小时内超过 3 次要求验证码 `100031`（防邮箱轰炸）；验证码未过期时重复请求返回 `100028`。
- 改密 `POST /user/password/change`：验证码连续错误 5 次作废验证码（需重新申请），并返回 `100031`。

## 客户端 IP 获取

后端通过 `c.ClientIP()` 获取客户端 IP：nginx 反代时设置 `X-Real-IP` / `X-Forwarded-For` 头，gin 在信任代理时读取这些头，否则回退到 `RemoteAddr`（直连 IP）。
