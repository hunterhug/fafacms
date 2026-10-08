# /file/list

列出自己上传的文件。

## 请求

```
POST /file/list

{
  "create_time_begin": 0,
  "create_time_end": 1562487749,
  "update_time_begin": 0,
  "update_time_end": 0,
  "size_begin": 0,
  "size_end": 0,
  "sort": ["=id", "-create_time", "-update_time", "=user_id", "=type", "=tag", "=store_type", "=status", "=size"],
  "hash_code": "",
  "url": "",
  "store_type": -1,
  "status": 0,
  "type": "",
  "tag": "",
  "id": 0,
  "is_picture": -1,
  "limit": 2,
  "page": 1
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的文件 | int | 秒，留空不筛选|  |
| create_time_end |    在此时间前创建的文件   |  int | | |
| update_time_begin | 在此时间后修改过的文件 | int | 秒，留空不筛选| |
| update_time_end |    在此时间前修改过的文件   |  int | | |
| size_begin | 容量大于此字节的文件 | int | 留空不筛选 | |
| size_end |    容量小于此字节的文件   |   int | | |
| sort |    文件列表排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| hash_code |    文件创建后的去冗余唯一标志  |   string | 可查询单一文件| |
| url |    文件URl  |   string | 可查询单一文件 | |
| store_type |    存储类型   |   int | 0表示查本地存储，1表示查对象存储,-1表示两者都查，目前不能两者混合使用，默认：0，请置-1 | 是 |
| status |    文件状态   |   int | 0表示正常文件，1表示文件被隐藏,-1表示两者都查，默认0 | 是 |
| type |    文件类型   |   string | 可选：file、image、other等，默认全部类型 | |
| tag |    文件标签，可根据标签进行筛选  |   string | | |
| id |    数据库该文件记录id  |   int | 可查询单一文件 | |
| is_picture |    查询图片文件 |   int | 0表示非图片文件，1表示图片，-1表示全部文件，默认：0，请置-1 | 是 |
| limit |   列表每页数量|   int | 最大页数：100，默认：20 | |
| page |    列表第几页 |   int | 默认：1 | 否|

查单个文件：

```
{
	"url": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
	"store_type": -1,
	"is_picture": -1
}
```

## 响应

```
{
  "flag": true,
  "cid": "fc020c6e7f5544fda63b90ce5bbf414b",
  "data": {
    "files": [
      {
        "id": 3,
        "type": "other",
        "tag": "other",
        "user_id": 1,
        "user_name": "admin",
        "file_name": "admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
        "really_file_name": "Overlooking_by_Lance_Asper.jpg",
        "hash_code": "admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673",
        "url": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
        "url_hash_code": "4345f0a754be0a74824aa4d76352011739cb5661b8baf0bef5a26b1fe2a0d3a9",
        "describe": "test",
        "create_time": 1562486429,
        "status": 0,
        "store_type": 0,
        "is_picture": 1,
        "size": 1459819
      }
    ],
    "limit": 1,
    "page": 1,
    "total_pages": 3
  }
}
```

`total_pages`为列表总页数。