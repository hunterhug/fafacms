# /node/take

获取节点信息。只能查自己的节点。

## 请求

```
POST /node/take

{
	"id": 1,
	"list_son":true,
	"seo":""
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 节点ID | int |seo和id不能同时为空 | 是 |
| list_son | 是否列出子节点 | bool | 如果有的话将列出子节点 | 否 |
| seo | 节点SEO | string |seo和id不能同时为空  | 是 |


## 响应

正常：

```
{
  "flag": true,
  "cid": "c60021f866dd4d98b47849a968eafca9",
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
  }
}
```

参数说明：

| 字段   |      含义   |类型  |   详情 |
|----------|--------|------|------|
| id | 节点数据库唯一标记 | int | 作为唯一标记定位该节点|
| user_id | 节点所属用户ID | int | |
| user_name | 节点所属用户名 | string | |
| seo | SEO唯一标记 | string | 同一个用户，不同节点有不同的SEO标记|
| name |    节点名称   |   string | |
| describe | 描述 | string |  |
| parent_node_id | 节点父亲 | int | 0表示一级节点,否则为二级节点 |
| image_path | 背景图 | string |  |
| create_time_int | 节点创建时间 | int | UTC时间戳 |
| update_time_int | 节点更新时间 | int | UTC时间戳 |
| create_time | 节点创建格式化时间| string | 东八区北京时间 |
| update_time | 节点更新格式化时间 | string | 东八区北京时间 |
| level | 节点级别 | int | 0表示顶层，1表示二级节点 |
| status | 节点状态 | int | 0表示节点正常，1表示隐藏节点 |
| sort_num | 节点排序 | int | 数字越大排越后，创建时会自动生成 |
| son | 孩子节点| []object | list_son为true时会返回孩子节点 |

节点不存在:

```
{
  "flag": false,
  "cid": "7033e3e4e77f4328a1810c6dad11aaa8",
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
  "cid": "6a6473c4b02e45ee9bb5529dcef30e98",
  "error": {
    "id": 100010,
    "msg": "paras input not right:id or seo empty"
  }
}
```