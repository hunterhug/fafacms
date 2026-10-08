# /comment/bad

举报评论。只有登录的用户可以举报。只能对正常的评论进行举报。当服务端开启自动违禁功能时，举报超过某个数值，评论将会被自动违禁。

同一用户对同一评论 **24 小时内只能举报一次**（重复举报返回 `110012`，不会取消之前的举报；24 小时后可再次举报，举报记录叠加，管理后台可见）。

## 请求

```
POST /comment/bad

{
	"id": 10,
	"reason": "垃圾广告"
}
```

`reason` 为举报原因（可选，长度 ≤50，支持自定义）。

## 响应

举报正常：

```
{
  "flag": true,
  "cid": "2e5b753d299c4d54b6b3b44382c6e36b",
  "data": "+"
}
```

24 小时内重复举报：

```
{
  "flag": false,
  "cid": "791152136ecd4e79888060431f0c6a81",
  "error": {
    "id": 110012,
    "msg": "already reported in 24 hours"
  }
}
```

评论不存在：

```
{
  "flag": false,
  "cid": "9a523e098ee749bb8ec3135e3c8ac56c",
  "error": {
    "id": 110008,
    "msg": "comment not found"
  }
}
```

评论被违禁：

```
{
  "flag": false,
  "cid": "7d3b11515f654c4a81a85f1dd0ad28c5",
  "error": {
    "id": 110009,
    "msg": "comment ban permit"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "7cd02a7d5db047d8870742231cb208f1",
  "error": {
    "id": 100010,
    "msg": "paras input not right:comment_id empty"
  }
}
```