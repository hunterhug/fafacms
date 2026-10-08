# /user/token/get

获取授权令牌，主要为登录接口服务。

此接口是你必须浏览的第一个`API`接口。不特殊说明，所有接口都以 `HTTP JSON` 请求服务端，且所有接口都需要加前缀 `/v1`。

## 请求

不需要前缀。

```
POST /user/token/get

{
	"user_name": "admin",
	"pass_wd": "987654321"
}
```

默认超级管理员账号密码为：`admin/admin`，`user_name`可以是用户唯一标志，也可以是用户注册时的邮箱。

## 响应

正常：

```
{
  "flag": true,
  "cid": "ef53847a4cf247dc8316562d453378dc",
  "data": "1_122934fe1a0d403ab5728776d4a36f0b"
}
```

其中 `data` 为登录授权令牌，请记住该令牌。授权令牌过期时间为7天，可以进行刷新。获取令牌时会获取到用户登录的IP和登录时间。

所有授权接口访问时，请在 `HTTP请求头部` 增加 `Auth=1_122934fe1a0d403ab5728776d4a36f0b`

用户名或密码出错：

```
{
  "flag": false,
  "cid": "b772498e6246486583eeb5e7ce95c592",
  "error": {
    "id": 100020,
    "msg": "username or password wrong:user or password wrong"
  }
}
```

当 `flag` 为 `false` 时业务出错，请查看错误码结构体:`error`。其中 `cid` 表示该次请求的唯一ID，用来调试时反查日志，可忽视。

## 通用错误

当用户未激活时，调用授权接口会出现:

```
{
  "flag": false,
  "error": {
    "id": 100004,
    "msg": "user not active:not active"
  }
}
```

用户被加入黑名单时会出现：

```
{
  "flag": false,
  "error": {
    "id": 100005,
    "msg": "user is in black"
  }
}
```

此两种情况，未激活应引导激活，黑名单时拒绝服务。未激活和黑名单用户，不能访问任何授权接口。

某些超级管理员接口，用户必须被赋予权限才可以访问，当用户属于的用户组不存在该权限，会拒绝服务：

```
{
  "flag": false,
  "error": {
    "id": 100006,
    "msg": "resource not allow"
  }
}
```

极少情况下，请求API时数据库会报错：

```
{
  "flag": false,
  "cid": "95aab4f2019a40ce938c85d91d7c5701",
  "error": {
    "id": 200000,
    "msg": "db operation err:Error 1146: Table 'fafa.[]' doesn't exist"
  }
}
```