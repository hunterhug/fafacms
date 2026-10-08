# /comment/real/name

取消自己匿名的评论。

## 请求

```
POST  /comment/real/name

{
	"id": 2
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | 评论id | 是 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "2e5b753d299c4d54b6b3b44382c6e36b"
}
```


评论不存在：

```
{
  "flag": false,
  "cid": "33d3fc07efc64da182524e9ad400a8c8",
  "error": {
    "id": 110008,
    "msg": "comment not found"
  }
}
```

参数有误：

```
{
  "flag": false,
  "cid": "576444a50afd48de898c938dbc122640",
  "error": {
    "id": 100010,
    "msg": "paras input not right:comment_id empty"
  }
}
```