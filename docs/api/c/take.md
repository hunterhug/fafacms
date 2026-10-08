# /content/take

获取单一内容。只能查自己的内容。

## 请求

```
POST /content/take

{
	"id": 10
}
```

可使用`内容id`获取内容信息。

## 响应

正常：

```
{
  "flag": true,
  "cid": "e220edd5aa8d4e0f90944dc620bbcefb",
  "data": {
    "id": 10,
    "seo": "124",
    "title": "",
    "pre_title": "sss",
    "user_id": 24,
    "user_name": "admin",
    "node_id": 1,
    "node_seo": "ddssdd",
    "status": 0,
    "top": 0,
    "describe": "",
    "pre_describe": "ddddd",
    "pre_flush": 0,
    "close_comment": 0,
    "version": 0,
    "create_time": 1568195282,
    "image_path": "",
    "views": 0,
    "sort_num": 7
  }
}
```

内容不存在：

```
{
  "flag": false,
  "cid": "d723af3927f94b7486f94d862ec1839c",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

参数不对：

```
{
  "flag": false,
  "cid": "52d4b77ae16348eab83fcae81a22bab0",
  "error": {
    "id": 100010,
    "msg": "paras input not right:Key: 'TakeContentRequest.Id' Error:Field validation for 'Id' failed on the 'required' tag"
  }
}
```