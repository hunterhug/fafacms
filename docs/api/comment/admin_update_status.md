# /comment/admin/update/status

修改评论状态，管理员接口。可以将评论违禁或解除违禁。

## 请求

```
POST  /comment/admin/update/status

{
	"id": 3,
	"status":1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 评论ID | int |  | 是 |
| status | 评论状态 | int | 0表示正常，1表示拉黑违禁 | 是 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "b3cf41004b8648f4990ad8c24fce70ad"
}
```

评论不存在或者评论已删除：

```
{
  "flag": false,
  "cid": "9d170024a4fb44dab0c1d43620b7183d",
  "error": {
    "id": 110008,
    "msg": "comment not found"
  }
}
```