# /content/update/seo

更新内容SEO。

## 请求

```
POST /content/update/seo

{
	"id":1,
	"seo":"qqq"
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| seo | SEO唯一标记 | string | 只能由unicode字符组成| 是 |

更新节点SEO，会级联更新该节点下所有内容的 `node_seo` 字段。

## 响应

正常：

```
{
  "flag": true,
  "cid": "c9c23cb972624ec5898b0a7fcd5a875f"
}
```

内容不存在:

```
{
  "flag": false,
  "cid": "66d5e6600a9d445481eff8134ddc9bf0",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

SEO被占用：

```
{
  "flag": false,
  "cid": "4223a99e257c461ab94c55e2ae82a143",
  "error": {
    "id": 110003,
    "msg": "content seo already be used"
  }
}
```