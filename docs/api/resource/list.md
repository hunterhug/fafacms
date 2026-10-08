# /resource/list

列出资源，管理员接口。这些资源可以被分配给组，拥有这些资源的组成员可以访问这些接口。

## 请求

```
POST /resource/list

{
	"id": 0,
	"name": "",
	"url": "",
	"limit": 5,
	"page": 1,
	"sort": [
		"=id",
		"+create_time",
		"-name"
	]
}
```

参数说明：


| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| sort |    组排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该资源唯一id  |   int | 可查询单一资源信息 | |
| name |    资源名字 |   string | 可查询单一资源信息 | |
| url |    资源链接 |   string | 可查询单一资源信息 | |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

## 响应

```
{
  "flag": true,
  "cid": "30710f6d25d849f28dda05dab3900082",
  "data": {
    "resources": [
      {
        "id": 4,
        "name": "User Update Admin",
        "url": "/v1/user/admin/update",
        "url_hash": "b5430876552376dc3c6eec90ec6df77046345da69647cb88ec1353c308d3818c",
        "describe": "User Update Admin",
        "admin": true,
        "create_time": 1565680140
      },
      {
        "id": 11,
        "name": "User List All",
        "url": "/v1/user/list",
        "url_hash": "c68704e19eb1c4c1c548441122edc2efe76e2e3f9dae69b1ef021681944d62d5",
        "describe": "User List All",
        "admin": true,
        "create_time": 1565680140
      },
      {
        "id": 5,
        "name": "User Create",
        "url": "/v1/user/create",
        "url_hash": "33d16be5100dcde2686f1c661dc2ed1bad02013d1c7719df9f76366837410ba3",
        "describe": "User Create",
        "admin": true,
        "create_time": 1565680140
      },
      {
        "id": 2,
        "name": "User Assign Group",
        "url": "/v1/user/assign",
        "url_hash": "f0b89af629c6e48bb88dcb81f95776fb7034331f77104cdf68c54ccc945dcf14",
        "describe": "User Assign Group",
        "admin": true,
        "create_time": 1565680140
      },
      {
        "id": 17,
        "name": "Update Group",
        "url": "/v1/group/update",
        "url_hash": "2ea81e793ace90e96b32178881796c5e31f0deb99a373112e4cd7e5bb65d7286",
        "describe": "Update Group",
        "admin": true,
        "create_time": 1565680140
      }
    ],
    "limit": 5,
    "page": 1,
    "total_pages": 5
  }
}
```