# /content

查看文章内容。

## 请求

不需要授权和前缀。

```
POST /content

{
	"id": 2,
	"user_name": "",
	"node_seo": "",
	"seo": "",
	"password": "",
	"more": true
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | 按 id 查询时使用（前端只用于把 id 翻译成 SEO 地址，地址栏不会出现 id） | 否 |
| user_name | 作者用户名 | string | 按 seo 查询时必须给 user_name 或 user_id | 否 |
| node_seo | 所属节点的 SEO | string | **按 seo 查询时必填**：文章 SEO 只在所属节点内唯一，`用户名 + 节点 SEO + 文章 SEO` 三者一起才能唯一定位一篇文章 | 否 |
| seo |    内容SEO  |  string |  按 seo 查询时使用 | 否 |
| password |    内容密码 |   int | 如果内容需要密码，需要将密码带上，否则将无法查看内容 | 否 |
| more |    是否显示更多  |  bool |  会显示同一个节点之前或之后的文章，实现next和pre的功能,sort_num大于该内容的为next，否则pre | 否 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "24a22acb5f6d433294a3151bf5a04939",
  "data": {
    "id": 5,
    "seo": "125",
    "title": "sss",
    "user_id": 1,
    "user_name": "admin",
    "node_id": 1,
    "node_seo": "qqq",
    "top": 0,
    "first_publish_time": "2019-09-14 19:35:07",
    "publish_time": "2019-09-14 19:35:39",
    "first_publish_time_int": 1568460907,
    "publish_time_int": 1568460939,
    "image_path": "",
    "views": 4,
    "is_lock": false,
    "describe": "ddddd",
    "next": {
      "id": 6,
      "seo": "1256",
      "title": "sss",
      "user_id": 1,
      "user_name": "admin",
      "node_id": 1,
      "node_seo": "qqq",
      "top": 0,
      "first_publish_time": "2019-09-14 19:35:10",
      "publish_time": "2019-09-14 19:35:41",
      "first_publish_time_int": 1568460910,
      "publish_time_int": 1568460941,
      "image_path": "",
      "views": 9,
      "is_lock": false,
      "describe": "",
      "sort_num": 5
    },
    "pre": {
      "id": 4,
      "seo": "123",
      "title": "",
      "user_id": 1,
      "user_name": "admin",
      "node_id": 1,
      "node_seo": "qqq",
      "top": 0,
      "first_publish_time": "2019-09-14 19:35:02",
      "publish_time": "1970-01-01 08:00:00",
      "first_publish_time_int": 1568460902,
      "publish_time_int": 0,
      "image_path": "",
      "views": 0,
      "is_lock": false,
      "describe": "",
      "sort_num": 3
    },
    "sort_num": 4
  }
}
```


`is_lock` 表示是否有密码。

`first_publish_time`是内容第一次发布的时间，`publish_time`是内容最新一次发布的时间。


内容不存在：

```
{
  "flag": false,
  "cid": "d80269f03d67484f92ddf1a5568c6ff4",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

内容违禁：

```
{
  "flag": false,
  "cid": "88dedb8aa6784c68b36e8949c7dcb83d",
  "error": {
    "id": 110002,
    "msg": "content ban permit"
  }
}
```

内容密码不对：

```
{
  "flag": false,
  "cid": "e58934cf92a94a39a5eb0dda021758dc",
  "error": {
    "id": 110001,
    "msg": "content password wrong"
  }
}
```