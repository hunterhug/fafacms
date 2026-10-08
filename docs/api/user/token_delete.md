# /user/token/delete

删除授权令牌。

## 请求

不需要前缀。

```
POST /user/token/delete
```

请在 `HTTP请求头部` 增加：
`Auth=1_122934fe1a0d403ab5728776d4a36f0b`

上述为临时令牌。

## 响应

正常：

```
{
  "flag": true
}
```

临时令牌不存在：

```
{
  "flag": false,
  "error": {
    "id": 100000,
    "msg": "get user session err:token not found"
  }
}
```
