# /content/list

列出内容。只能列出自己的内容。

## 请求

```
POST /content/list

{
	"id": 0,
	"seo": "",
	"node_id": 0,
	"node_seo": "",
	"top": -1,
	"status": -1,
	"close_comment": -1,
	"publish_type": -1,
	"create_time_begin": 0,
	"create_time_end": 0,
	"update_time_begin": 0,
	"update_time_end": 0,
	"publish_time_begin": 0,
	"publish_time_end": 0,
	"first_publish_time_begin": 0,
	"first_publish_time_end": 0,
	"limit": 1,
	"page": 1,
	"sort": [
		"=id",
		"-user_id",
		"-top",
		"+sort_num",
		"-first_publish_time",
		"-publish_time",
		"-create_time",
		"-update_time",
		"-views",
		"=comment_num",
		"=bad",
		"=cool",
		"=version",
		"+status",
		"=seo"
	]
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| create_time_begin | 在此时间后创建的内容 | int | 秒，留空不筛选 | |
| create_time_end |    在此时间前创建的内容  |  int | | |
| update_time_begin | 在此时间后修改过的内容 | int | 秒，留空不筛选| |
| update_time_end |    在此时间前修改过的内容   |   int | | |
| first_publish_time_begin | 在此时间后第一次发布的内容 | int | 秒，留空不筛选| |
| first_publish_time_end |    在此时间前第一次发布的内容   |   int | | |
| publish_time_begin | 在此时间后最后一次发布的内容 | int | 秒，留空不筛选| |
| publish_time_end |    在此时间前最后一次发布的内容   |   int | | |
| sort |    内容排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | |
| id |    数据库该内容唯一id  |   int | 可查询单一内容信息 | |
| seo |    内容SEO |   string | 可查询单一内容信息 | |
| node_id |    内容所属节点ID |   int | 可作筛选条件 | |
| node_seo |    内容所属节点SEO |   string | 可作筛选条件 | |
| top |    置顶 |   int | 可作筛选条件，-1表示全部，1表示查找置顶的，0表示不置顶 | 是 |
| status |    状态 |   int | 可作筛选条件，-1表示查找全部状态(垃圾箱3与用户彻底删除4不会被包含进来)，0表示正常，1表示查找隐藏内容，2表示查找拉黑的内容，3表示查找垃圾箱内容，4表示查找用户彻底删除的内容(仅管理员可见) | 是 |
| close_comment |  评论状态 |   int | 可作筛选条件，-1表示查找所有，0查找评论关闭，1评论开启 | 是 |
| publish_type |  发布状态 |   int | 可作筛选条件，-1表示查找所有，0查找从未发布的，1查找发布过的，2查找发布过且没有缓冲草稿的，3表示查找发布过但已编辑待再次发布的 | 是 |
| password_type |  密码状态 |   int | 可作筛选条件，-1表示查找所有，0查找没有密码的，1查找有密码的 | 是 |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

## 响应

```
{
  "flag": true,
  "cid": "e9e25873149f4533830a844d832aa3ef",
  "data": {
    "contents": [
      {
        "id": 1,
        "seo": "qqq",
        "title": "1234",
        "pre_title": "1234",
        "user_id": 24,
        "user_name": "admin",
        "node_id": 2,
        "node_seo": "123457",
        "status": 1,
        "top": 1,
        "describe": "",
        "pre_describe": "",
        "pre_flush": 1,
        "close_comment": 0,
        "version": 1,
        "create_time": 1568193256,
        "update_time": 1568277661,
        "first_publish_time": 1568277661,
        "publish_time": 1568277661,
        "image_path": "",
        "views": 0,
        "password": "4545",
        "sort_num": 0
      }
    ],
    "limit": 5,
    "page": 1,
    "total_pages": 1
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
| update_time | 内容修改时间 | int | 时间戳 |
| first_publish_time | 内容第一次发布时间 | int | 时间戳 |
| publish_time | 内容最后一次发布时间 | int | 时间戳 |
| status | 内容状态 | int | 0表示正常，1表示隐藏内容，2表示内容违禁被拉黑，3表示已送到垃圾箱，4表示用户彻底删除(仅管理员可见)，只有0的内容才会显示在前端 |
| sort_num | 内容排序 | int | 数字越大排越后，创建时会自动生成，新建的内容都在该层最后 |
| version | 发布版本 | int | 0表示没有发布过，大于1表示已经发布过一次，只有发布过的内容才会显示在前端 |
| pre_flush | 是否预发布内容已清空 | int | 1表示已经发布了，没有预发布内容，0表示存在预发布的内容，可以提醒UI层 |
| views | 内容浏览量 | int | 不记名递增 |
| password | 内容所属节点 | string | 非空时前端访问需要密码 |
| close_comment | 评论设置 | int |  0表示开启评论，1表示关闭评论 |
| bad | 举报数 | int |  |
| cool | 点赞数 | int |  |
| comment_num | 评论数 | int | 内容下的评论数 |