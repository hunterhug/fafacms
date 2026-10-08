# /node/create

创建节点。

只有VIP用户才可以创建节点。

## 请求

```
POST /node/create

{
	"seo":"s3s2ss",
	"name":"node",
	"describe":"ddddd",
	"parent_node_id":0,
	"image_path":""
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| seo | SEO唯一标记 | string | 只能由unicode字符组成，不能为空 | 是 |
| name |    节点名称   |   string | | 是 |
| describe | 描述 | string | 非空时更改 | |
| parent_node_id | 节点父亲 | int | 留空时为一级节点,否则为二级节点 | |
| image_path | 背景图 | string | 非空时更改 | 否 |

目前节点只有两级设计。

## 响应

正常：

```
{
  "flag": true,
  "cid": "89b4da8f1c504223983742dacbc2a422",
  "data": {
    "id": 2,
    "user_id": 1,
    "user_name": "admin",
    "seo": "s3s2ss",
    "status": 0,
    "name": "node",
    "describe": "ddddd",
    "create_time": 1567836543,
    "image_path": "",
    "parent_node_id": 0,
    "level": 0,
    "sort_num": 1
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
| create_time | 节点创建时间 | int | 时间戳 |
| level | 节点级别 | int | 0表示顶层，1表示二级节点 |
| status | 节点状态 | int | 0表示节点正常，1表示隐藏节点 |
| sort_num | 节点排序 | int | 数字越大排越后，创建时会自动生成，新建的节点都在该层最后 |


SEO被占用：

```
{
  "flag": false,
  "cid": "c32d870ad2fd4367b7f483d868d3ef68",
  "error": {
    "id": 101000,
    "msg": "content node seo already be used"
  }
}
```

父节点不存在：

```
{
  "flag": false,
  "cid": "b1fae8736fea4c05be94f351c691c012",
  "error": {
    "id": 101002,
    "msg": "parent content node not found"
  }
}
```

背景图不存在：

```
{
  "flag": false,
  "cid": "443634cfa5664fb48c365a670b9a65cd",
  "error": {
    "id": 100030,
    "msg": "file can not be found:image url not exist"
  }
}
```

不是VIP：

```
{
  "flag": false,
  "cid": "443634cfa5664fb48c365a670b9a65cd",
  "error": {
    "id": 99996,
    "msg": "you are not vip"
  }
}
```

参数有误：

```
{
  "flag": false,
  "cid": "3717979eadbd46f481ae1bcf2dce4bca",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'CreateNodeRequest.Seo' Error:Field validation for 'Seo' failed on the 'alphanumunicode' tag"
  }
}
```