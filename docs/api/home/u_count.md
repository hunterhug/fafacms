# /u/count

统计某一用户的文章内容数，按照第一次发布时间(first_publish_time)的天数进行统计。只有已经发布和正常显示的内容才会被统计。根据服务端的时区设置进行统计。

`count`（按天统计的文章数）的口径与 `/u/content` 列表一致：

- 游客与其他用户：排除隐藏节点及其子节点下的文章；
- 作者本人：包含自己的全部已发布文章。

## 请求

不需要授权和前缀。

```
POST /u/count

{
	"user_id": 1,
	"user_name": ""
}
```

参数说明：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| user_id | 用户ID | int | user_name和user_id必存在一个 | 是 |
| user_name |    用户名  |  string | user_name和user_id必存在一个 | 是 |

## 响应

正常：

```
{
  "flag": true,
  "data": {
    "info": [
      {
        "count": 1,
        "days": "20190527",
        "first_publish_time_begin": 1558944000,
        "first_publish_time_end": 1559030400
      },
      {
        "count": 1,
        "days": "20190914",
        "first_publish_time_begin": 1568448000,
        "first_publish_time_end": 1568534400
      }
    ],
    "user_id": 1,
    "user_name": "admin"
  }
}
```

`days`时为本地格式化时间字符串，`first_publish_time_begin`和`first_publish_time_end`是返回给用户的`UTC时间戳`，统计在`first_publish_time_begin<= first_publish_time <first_publish_time_end`时间内的文章内容。在前端用户看来，`first_publish_time` 就是内容的创建时间。

用户找不到：

```
{
  "flag": false,
  "error": {
    "id": 100003,
    "msg": "user not found"
  }
}
```

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