# /u/nodes

查找某一个用户下面的所有节点。只列出不隐藏的节点。

## 请求

不需要授权和前缀。

```
POST /u/nodes

{
	"user_id": 1,
	"user_name": "",
	"sort": [
		"=id",
		"+sort_num",
		"-create_time",
		"-update_time",
		"+status",
		"=seo",
		"=content_num"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id | 节点所属用户的ID | int | user_name和user_id必存在一个 | 是 |
| user_name |    节点所属用户名  |  string | user_name和user_id必存在一个 | 是 |
| sort |    排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |

## 响应

正常：

```
{
  "flag": true,
  "data": {
    "nodes": [
      {
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
      },
      {
        "id": 4,
        "seo": "s3sd2sds",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:21:18",
        "create_time_int": 1567837278,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 1,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 2,
        "seo": "s3s2ss",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:09:03",
        "create_time_int": 1567836543,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 1,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 5,
        "seo": "11你好1",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:23:18",
        "create_time_int": 1567837398,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 3,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 6,
        "seo": "22222",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:30:24",
        "create_time_int": 1567837824,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 4,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 7,
        "seo": "5",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:33:47",
        "create_time_int": 1567838027,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 5,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 8,
        "seo": "123456",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:37:16",
        "create_time_int": 1567838236,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 6,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 9,
        "seo": "1234567",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-07 14:39:34",
        "create_time_int": 1567838374,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 7,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      },
      {
        "id": 10,
        "seo": "123457",
        "name": "node",
        "describe": "ddddd",
        "image_path": "",
        "create_time": "2019-09-14 15:07:38",
        "create_time_int": 1568444858,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 8,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      }
    ]
  }
}
```

会返回所有节点。

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