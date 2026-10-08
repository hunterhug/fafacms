# /comment/create

创建评论。可以对内容评论，也可以对`评论`评论。当内容被设置为不可评论时，无法评论。

如对内容评论时：

```
~~~~标题：告白气球...

小花：周杰伦发歌啦
```

对第一层评论的评论：

```
~~~~标题：告白气球...

鸡鸡：是啊，周杰伦的歌好好听哦

@小花：周杰伦发歌啦
```

对超过一层评论的评论：

```
~~~~标题：告白气球...

铁柱的弟弟：+2 // @铁柱：+1

@小花：周杰伦发歌啦

----

铁柱：+1 // @鸡鸡：是啊，周杰伦的歌好好听哦

@小花：周杰伦发歌啦
```

## 请求

```
POST  /comment/create

{
	"content_id": 0,
	"comment_id": 0,
	"is_to_comment": true,
	"body": "<dddd><ddd>",
	"anonymous":false
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| content_id | 内容ID | int | 当is_to_comment为false有效，评论针对的是内容 | 否 |
| comment_id |    评论ID   |   int | 当is_to_comment为true有效，评论针对的是评论 | 否 |
| is_to_comment | 评论的对象是不是评论 | bool | 当true时表示要对`评论`进行评论，会自动查找目标评论的关联数据，true时content_id无效，false时comment_id无效 | 是 |
| body | 内容所属节点 | int | 不能为空 | 是 |
| anonymous | 是否匿名评论 | bool | 一旦匿名后无法取消匿名 | 否 |

评论时，会对内容的`comment_num`字段进行`+1`操作。

## 响应

正常：

```
{
  "flag": true,
  "cid": "19546a49592140b9b2e61fef6edbc6bd",
  "data": 1
}
```

`data`字段的值表示新创建的评论`ID`

内容违禁无法评论：

```
{
  "flag": false,
  "cid": "744b2c2fe2704c4c939759295bd2f479",
  "error": {
    "id": 110002,
    "msg": "content ban permit"
  }
}
```


内容不存在无法评论：

```
{
  "flag": false,
  "cid": "5c7896172a2f41afa8ea032ecd090147",
  "error": {
    "id": 110000,
    "msg": "content not found"
  }
}
```

评论不存在，无法评论`评论`：

```
{
  "flag": false,
  "cid": "bacfe721917948a0b948a25381130660",
  "error": {
    "id": 110008,
    "msg": "comment not found"
  }
}
```

评论违禁无法评论`评论`：

```
{
  "flag": false,
  "cid": "744b2c2fe2704c4c939759295bd2f479",
  "error": {
    "id": 110009,
    "msg": "comment ban permit"
  }
}
```

内容设置为不允许评论：

```
{
  "flag": false,
  "cid": "744b2c2fe2704c4c939759295bd2f479",
  "error": {
    "id": 110010,
    "msg": "content close comment"
  }
}
```

参数有误：

评论内容为空：

```
{
  "flag": false,
  "cid": "242ca89bd7e3446b9e3f3c3e89cbb23f",
  "error": {
    "id": 100010,
    "msg": "paras input not right:body empty"
  }
}
```

针对内容评论，内容ID为空：

```
{
  "flag": false,
  "cid": "6599434a6aa84dd894cabb7c42b36e97",
  "error": {
    "id": 100010,
    "msg": "paras input not right:content_id empty"
  }
}
```

针对`评论`评论，评论ID为空：

```
{
  "flag": false,
  "cid": "dc147b6d2f3f450ca023666099c6e1cf",
  "error": {
    "id": 100010,
    "msg": "paras input not right:comment_id empty"
  }
}
```