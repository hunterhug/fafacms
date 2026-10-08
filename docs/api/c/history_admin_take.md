# /content/history/admin/take

获取单一历史内容。管理员接口，可以查看所有人的历史内容。

## 请求

```
POST /content/history/take

{
	"id": 10
}
```

可使用`历史内容id`获取信息。

## 响应

正常：

```
{
  "flag": true,
  "cid": "8425be66d0fc4d9c8b617c70c17b3a95",
  "data": {
    "id": 9,
    "content_id": 1,
    "title": "1sss234",
    "user_id": 24,
    "node_id": 2,
    "describe": "12ddddd3",
    "types": 1,
    "create_time": 1568282900
  }
}
```

历史内容不存在：

```
{
  "flag": false,
  "cid": "bd8a95e81c2a46c4b75cc52eacdaf055",
  "error": {
    "id": 110006,
    "msg": "content history can not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "45ef13b7552545f6bf5d5da047368768",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'TakeContentHistoryRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```