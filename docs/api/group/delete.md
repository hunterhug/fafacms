# /group/delete

删除组，管理员接口。

## 请求

```
POST /group/delete

{
	"id": 3,
	"name": ""
}
```

可使用`组名`或者`组id`删除组。

## 响应

正常：

```
{
  "flag": true,
  "cid": "a3f531b9a9a3448c82ebf98260aa09ff"
}
```

组不存在：

```
{
  "flag": false,
  "cid": "462cb5b7f1794acbba9ffaa79800239a",
  "error": {
    "id": 100041,
    "msg": "group not found"
  }
}
```

当组下有用户时也会报错。