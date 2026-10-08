# /content/update/info

更新内容信息。此处事实更新的是预发布字段，更新完后需要使用发布API进行正式发布。

只有VIP用户才可以更新内容。

## 请求

```
POST /content/update/info

{
	"id": 1,
	"title": "1234",
	"describe": "123"
	"save": false
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 内容ID | int | | 是 |
| title |内容标题 | string |  | 是 |
| describe | 内容描述，也就是内容的真实body | string | 不能为空  | 是 |
| save | 保存上次预发布记录进历史表 | bool |  | 否 |


更新内容时，如果服务开启了历史记录设置，且`save`为`true`，会将预发布的内容存储在历史表中，历史表状态码为`0`，表示上一次的草稿，再更新预发布内容。

## 响应

正常：

```
{
  "flag": true,
  "cid": "65cfe4a4fd194d0f9f6242c8f255088c"
}
```

内容不存在:

```
{
  "flag": false,
  "cid": "e7dae50a4706424ab09360113a7ceab4",
  "error": {
    "id": 110000,
    "msg": "content not found"
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