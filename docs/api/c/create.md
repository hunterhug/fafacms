# /content/create

创建内容。每个内容都归属一个节点。

只有VIP用户才可以创建内容。

## 请求

```
POST  /content/create

{
	"seo": "dddd",
	"title": "xxxxxxx",
	"describe": "ddddd",
	"top": 0,
	"node_id": 1,
	"image_path": "",
	"password": "",
	"close_comment": 0
	
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| seo | SEO唯一标记 | string | 只能由unicode字符组成，不能为空 | 是 |
| title |    内容标题   |   string | | 是 |
| describe | 内容描述，也就是内容的真实body | string | 不能为空  | 是 |
| node_id | 内容所属节点 | int |  | 是 |
| top | 内容是否置顶 | int | 1表示置顶，默认0 | 否 |
| password | 内容所属节点 | string | 非空时前端访问需要密码 | 否 |
| image_path | 背景图 | string | 非空时更改，必须通过上传接口上传的 | 否 |
| close_comment | 评论设置 | int | 0表示开启评论，1表示关闭评论，默认开启评论 | 否 |


创建内容时，内容标题和描述会先存储在预发布字段，需要使用`publish API`才真正存储在指定字段。越晚创建的内容`sort_num`越大，系统智能创建排序。

## 响应

正常：

```
{
  "flag": true,
  "cid": "a4ff74a6009341049d6fb05850d4fff3",
  "data": {
    "id": 9,
    "seo": "dddd",
    "title": "",
    "pre_title": "xxxxxxx",
    "user_id": 24,
    "user_name": "admin",
    "node_id": 1,
    "node_seo": "1234567",
    "status": 0,
    "top": 0,
    "describe": "",
    "pre_describe": "ddddd",
    "pre_flush": 0,
    "close_comment": 0,
    "version": 0,
    "create_time": 1568194462,
    "image_path": "",
    "views": 0,
    "sort_num": 7
  }
}
```

参数说明：

| 字段   |      含义   |类型  |   详情 |
|----------|--------|------|------|
| id | 内容ID，数据库唯一标记 | int | 作为唯一标记定位该节点|
| user_id | 内容所属用户ID | int | |
| user_name | 内容所属用户名 | string | |
| seo | SEO唯一标记 | string | 同一个用户，不同内容有不同的SEO标记|
| title |    内容标题   |   string | 第一次创建时为空，只有发布后的标题才会显示在这里|
| describe | 内容描述 | string | 第一次创建时为空，只有发布后的描述才会显示在这里  |
| pre_title |    预发布的内容标题   |   string | 使用publish api会刷进title字段|
| pre_describe | 预发布的内容描述 | string |  使用publish api会刷进describe字段 |
| node_id | 内容所属节点ID | int |  |
| node_seo | 内容所属节点SEO | int | 冗余字段，当节点SEO更新时会级联更新 |
| top | 内容是否置顶 | int | 1表示置顶 |
| image_path | 背景图 | string |  |
| create_time | 内容创建时间 | int | 时间戳 |
| status | 内容状态 | int | 0表示正常，1表示隐藏内容，2表示内容违禁被拉黑，3表示已送到垃圾箱，只有0的内容才会显示在前端 |
| sort_num | 内容排序 | int | 数字越大排越后，创建时会自动生成，新建的内容都在该层最后 |
| version | 发布版本 | int | 0表示没有发布过，大于1表示已经发布过一次，只有发布过的内容才会显示在前端 |
| pre_flush | 是否预发布内容已清空 | int | 1表示已经发布了，没有预发布内容，0表示存在预发布的内容，可以提醒UI层 |
| views | 内容浏览量 | int | 不记名递增 |
| password | 内容所属节点 | string | 非空时前端访问需要密码 |
| close_comment | 评论设置 | int |  0表示开启评论，1表示关闭评论 |

SEO被占用：

```
{
  "flag": false,
  "cid": "4b02c35c134740558037ffbe88a6bf1f",
  "error": {
    "id": 110003,
    "msg": "content seo already be used"
  }
}
```

内容节点不存在：

```
{
  "flag": false,
  "cid": "15d190322f7c4067878925b85503ffcb",
  "error": {
    "id": 101001,
    "msg": "content node not found"
  }
}
```

背景图不存在：

```
{
  "flag": false,
  "cid": "d0b4c1971f614694a0cce659b15aaf91",
  "error": {
    "id": 100030,
    "msg": "file can not be found"
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
  "cid": "cbe3ad59491949b2b076333a2ed34ade",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'CreateContentRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"
  }
}
```