# /content/restore

从历史内容中恢复预发布内容。

## 请求

```
POST /content/restore

{
	"history_id": 8,
	"save": true
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| history_id | 历史内容ID | int | | 是 |
| save | 是否将之前的预发布字段（草稿）刷新进历史 | bool | | 否 |

从该历史记录中获取之前的内容，刷到预发布内容中。

如果 `save` 为 `true` 时，会先将之前的预发布内容保存进历史记录，然后再做后继操作。保存的历史内容，类型为2，表示恢复时的草稿。

## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

历史内容不存在:

```
{
  "flag": false,
  "cid": "8197cd0a55d14b7d89a8ea226362d2fb",
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
  "cid": "08f87d8d1f49458d950bc33cd2e7def1",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'RestoreContentRequest.HistoryId' Error:Field validation for 'HistoryId' failed on the 'required' tag"
  }
}
```