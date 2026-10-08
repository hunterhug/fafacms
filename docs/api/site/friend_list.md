# /friend/list

列出全部友情链接（含隐藏项）。管理员接口。

## 请求

需要前缀 `/v1`。

```
POST /v1/friend/list

{
	"limit": 100,
	"page": 1,
	"sort": ["+sort_num", "=id"]
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| limit / page | 分页 | int | 否 |
| sort | 排序（默认按 sort_num 升序、id 升序） | []string | 否 |

## 响应

```
{
  "flag": true,
  "data": {
    "friend_links": [
      { "id": 1, "name": "GitHub", "url": "https://github.com/hunterhug/fafacms", "sort_num": 0, "hide": 0, "open_new": 1, "create_time": 1788100000 }
    ],
    "total": 1,
    "total_pages": 1
  }
}
```
