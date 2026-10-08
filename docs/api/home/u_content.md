# /u/content

列出文章内容。只有已发布的正常内容才会被列出。

列表口径：

- 排除文章自身状态为 1 隐藏 / 2 封禁 / 3 回收站 / 4 已删除，以及未发布（`version=0`）的文章。
- **排除隐藏节点（`content_node.status=1`）及其子节点下的文章**：这些文章属于作者的私人内容，
  只有作者本人（请求头带 `Auth` 且就是该作者）仍能看到，返回项会带 `node_hidden: true`；
  其他人（含管理员）与游客都看不到。文章详情接口 `/content` 不受此限制，直链仍可打开。
- 因此首页、发现页、按标题搜索、他人个人主页看到的内容一致；作者本人看自己的主页时
  包含隐藏节点的文章。

## 请求

不需要授权和前缀。

```
POST /u/content

{
    "user_id": 0,
    "user_name": "",
    "node_id": 0,
    "node_seo": "",
    "first_publish_time_begin": 0,
    "first_publish_time_end": 0,
    "publish_time_begin": 0,
    "publish_time_end": 0,
    "limit": 1,
    "page": 1,
    "sort": [
		"=id",
		"-user_id",
		"-top",
		"+sort_num",
		"-first_publish_time",
		"-publish_time",
		"-views",
		"=comment_num",
		"=bad",
		"=cool",
		"=seo"
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id | 用户ID | int |  | 否 |
| user_name |    用户名  |  string |  | 否 |
| title |    标题关键字（模糊匹配已发布标题 title，不搜正文/草稿） |  string | 可作搜索条件 | 否 |
| node_id |    内容所属节点ID |   int | 可作筛选条件 | 否 |
| include_children | 是否下钻：node_id 为一级节点时，true 则同时列出其二级子节点的文章 | bool | 默认 false | 否 |
| node_seo |    内容所属节点SEO |   string | 可作筛选条件 | 否 |
| first_publish_time_begin | 在此时间后第一次发布的内容 | int | 秒，留空不筛选 | 否 |
| first_publish_time_end |    在此时间前第一次发布创建的内容  |  int | | 否 |
| publish_time_begin | 在此时间后最后一次发布过的内容 | int | 秒，留空不筛选| |
| publish_time_end |    在此时间前最后一次发布过的内容   |   int | | |
| sort |    内容排序   |   []string | 默认按照以上进行多列排序.=表示不排序，-表示降序，+表示升序 | 否 |
| limit |  列表每页数量 |   int | 最大页数：100，默认：20 | |
| page |  列表第几页 |   int | 默认：1 | 否 |

## 响应

正常：

```
{
  "flag": true,
  "cid": "4e20a95a9280450283e410bb240cece6",
  "data": {
    "contents": [
      {
        "id": 2,
        "seo": "1243",
        "title": "sss",
        "user_id": 1,
        "user_name": "admin",
        "user_nick_name": "管理员",
        "user_head_photo": "https://example.com/avatar.png",
        "node_id": 1,
        "node_seo": "qqq",
        "top": 0,
        "first_publish_time": "2019-05-27 00:41:28",
        "publish_time": "2019-09-14 15:08:07",
        "first_publish_time_int": 1558888888,
        "publish_time_int": 1568444887,
        "image_path": "",
        "views": 0,
        "is_lock": false,
        "describe": "ddddd"
      }
    ],
    "limit": 1,
    "page": 1,
    "total_pages": 2
  }
}
```


`is_lock` 表示是否有密码。

`first_publish_time`是内容第一次发布的时间，`publish_time`是内容最新一次发布的时间。

`describe` 是正文**摘要**（列表场景，已按字符安全截断，最长约 200 个字符），完整正文请用 `/content` 接口获取。