# /friend/create

创建友情链接。管理员接口。

## 请求

需要前缀 `/v1`。

```
POST /v1/friend/create

{
	"name": "链接名",
	"url": "https://example.com",
	"sort_num": 0,
	"hide": 0,
	"open_new": 1
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| name | 链接名 | string | 是 |
| url | 链接地址 | string | 是 |
| sort_num | 排序（默认最大 +1） | int | 否 |
| hide | 0 显示 / 1 隐藏（默认 0） | int | 否 |
| open_new | 0 同窗 / 1 新窗（默认 1） | int | 否 |

## 响应

正常：

```
{
  "flag": true
}
```
