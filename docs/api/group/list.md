# /group/list

组列表，管理员接口。

## 请求

```
POST /group/list

{
	"id": 0,
	"name": "",
	"create_time_begin": 0,
	"create_time_end": 0,
	"update_time_begin": 0,
	"update_time_end": 0,
	"limit": 5,
	"page": 1,
	"sort": [
		"=id",
		"=name",
		"-create_time",
		"=update_time"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的组 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的组  |  int | | |
| update_time_begin | 在此时间后修改过的组 | int | 秒，留空不筛选| |
| update_time_end |    在此时间前修改过的组   |   int | | |
| sort |    组排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该组唯一id  |   int | 可查询单一组信息 | |
| name |    组名字 |   string | 可查询单一组信息 | |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

## 响应

```
{
  "flag": true,
  "cid": "ba96190fde9340d5ab3f13fb80c32dca",
  "data": {
    "groups": [
      {
        "id": 2,
        "name": "groupsname1",
        "describe": "test grouffpcccccccccccc",
        "create_time": 1562493116,
        "image_path": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg"
      },
      {
        "id": 1,
        "name": "groupname",
        "describe": "test groDDDup",
        "create_time": 1562492308,
        "update_time": 1562492748,
        "image_path": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg"
      }
    ],
    "limit": 5,
    "page": 1,
    "total_pages": 1
  }
}
```