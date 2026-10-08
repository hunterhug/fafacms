# /node/update/seo

更新节点SEO。

## 请求

```
POST /node/update/seo

{
	"id":1,
	"seo":"qqq"
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 节点ID | int | | 是 |
| seo | SEO唯一标记 | string | 只能由unicode字符组成| 是 |

更新节点SEO，会级联更新该节点下所有内容的 `node_seo` 字段。

## 响应

正常：

```
{
  "flag": true,
  "cid": "42220e965ac441b68348cfbcd3005e9b"
}
```

节点不存在:

```
{
  "flag": false,
  "cid": "3c7f9f8a20d148fcb04fb0d8c31ad8fd",
  "error": {
    "id": 101001,
    "msg": "content node not found"
  }
}
```

SEO被占用：

```
{
  "flag": false,
  "cid": "71d453c8987443b7a26e05f0d68fdbb2",
  "error": {
    "id": 101000,
    "msg": "content node seo already be used"
  }
}
```