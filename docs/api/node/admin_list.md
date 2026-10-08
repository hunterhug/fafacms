# /node/admin/list

获取节点列表，管理员接口。可以列出某用户下的所有节点。

## 请求

```
POST /node/admin/list

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

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id | 用户ID | int | user_id和user_name两者必须存在一个 | 是 |
| user_name | 用户唯一名字标识 | string | user_id和user_name两者必须存在一个 | 是 |
| sort | 排序 | []string | 默认按照以上字段排序，+表示降序，-升序，=啥也不做 | 否 |

没有分页功能，每个人的节点数量是有限的。

## 响应

正常：

```
{
  "flag": true,
  "cid": "1732177b50664a2bac8dd75991f37f8e",
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
            "update_time": "",
            "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
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
        "update_time": "",
        "update_time_int": 0,
        "user_id": 1,
        "user_name": "admin",
        "sort_num": 7,
        "level": 0,
        "status": 0,
        "parent_node_id": 0
      }
    ]
  }
}
```

参数不正确：

```
{
  "flag": false,
  "cid": "c3bcaa541fd44e62987c396e0abe7f32",
  "error": {
    "id": 100010,
    "msg": "paras input not right:where is empty"
  }
}
```