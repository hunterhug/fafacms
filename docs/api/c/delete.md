# /content/delete

用户从回收站「彻底删除」内容：**仍是软删除**，把状态由「回收站(status=3)」改为「用户彻底删除(status=4)」，不物理删除、不级联删除评论/点赞/通知等他人生成记录。

- 仅能从回收站(status=3)中的内容彻底删除；其它状态调用返回 110007。
- 彻底删除后，内容在**用户世界消失**（用户「全部/回收站」列表都不再出现），**管理员后台仍可见**（可按 status=4 过滤）。
- 内容历史不会物理删除。

## 请求

```
POST /content/delete

{
	"id": 1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

内容不存在:

```
{
  "flag": false,
  "cid": "d65f8089f6674713981eb461a91156d2",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

不能删除内容：

```
{
  "flag": false,
  "cid": "4a6a3f70b7fc42659d7006cd83a1e56c",
  "error": {
    "id": 110007,
    "msg": "content can not delete for content not in rubbish"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "a591ca7626c84897b3e2ebf8f8c91d0f",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'ReallyDeleteContentRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```