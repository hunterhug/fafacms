# /u/node

查找某一个用户下面的某一个节点信息。该节点如果是隐藏的（`status=1`），默认看不到。

例外：**节点作者本人**请求时可以取到自己的隐藏节点（用于文章详情页显示节点名字）。
管理员不特殊对待，与普通访客一致。

## 请求

不需要授权和前缀。

```
POST /u/node

{
	"user_id": 1,
	"user_name": "",
	"seo": "",
	"id": 1,
	"list_son": true
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id | 节点所属用户的ID | int | user_name和user_id必存在一个 | 是 |
| user_name |    节点所属用户名  |  string | user_name和user_id必存在一个 | 是 |
| seo | 节点的SEO标志 | string | id和seo必存在一个 | 是 |
| id |    节点ID   |   int | id和seo必存在一个 | 是 |
| list_son |    是否列出儿子节点   |   bool |  | 否  |

## 响应


正常：

```
{
  "flag": true,
  "data": {
    "id": 1,
    "seo": "qqq",
    "name": "1234567",
    "describe": "",
    "image_path": "/storage/admin/other/admin_7e59164ec6196145f0ad162b4da7d0738033555a109ca01d1dee635b8494cef1.jpeg",
    "create_time": "2019-09-07 13:53:23",
    "create_time_int": 1567835603,
    "update_time": "2019-09-07 14:54:09",
    "update_time_int": 1567839249,
    "user_id": 1,
    "user_name": "admin",
    "sort_num": 0,
    "level": 0,
    "status": 0,
    "parent_node_id": 0,
    "son": [
      {
        "id": 3,
        "seo": "s3s2sds",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:20:57",
        "create_time_int": 1567837257,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 0,
        "level": 1,
        "status": 0,
        "parent_node_id": 1
      }
    ]
  }
}
```

节点没找到：

```
{
  "flag": false,
  "error": {
    "id": 101001,
    "msg": "content node not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "error": {
    "id": 100010,
    "msg": "paras input not right:where is empty"
  }
}
```