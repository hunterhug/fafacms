# /file/update

用户更新自己创建的文件描述，标签等。只能修改自己上传的文件。

## 请求

```
POST /file/update

{
	"id": 3,
	"tag": "mother_tag",
	"describe": "change this file to mother tag",
	"hide":true
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| id | 文件数据库唯一标记 | int | | 是 |
| tag |    标签   |   string | 非空时更改 | |
| describe | 描述 | string | 非空时更改 | |
| hide | 隐藏文件 | bool | 使用true时可模拟删除文件 | 否 |

## 响应

```
{
  "flag": true,
  "cid": "4f58ad9d0ece4785826449fee9e807df",
  "data": true
}
```

即使文件 `id` 不存在，返回的结果也为 `true`，除非内部程序请求数据库，网络出错等。