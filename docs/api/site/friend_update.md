# /friend/update

更新友情链接。管理员接口。

## 请求

需要前缀 `/v1`。

```
POST /v1/friend/update

{
	"id": 1,
	"name": "新名",
	"url": "...",
	"sort_num": 2,
	"hide": 0,
	"open_new": 1
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| id | 友情链接 ID | int | 是 |
| name / url / sort_num / hide / open_new | 其余字段不传则保持原值 | - | 否 |

## 响应

正常：

```
{
  "flag": true
}
```
