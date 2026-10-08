# /content/sort

内容排序，可实现拖曳排序。只有同一个节点的内容才可以拖曳。原理同节点排序。

## 请求

```
POST /content/sort


{
	"xid": 1,
	"yid": 2
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| xid | X内容ID | int | X内容不能等于Y内容| 是 |
| yid | Y内容ID | int | | 否 |

把X内容拖到Y内容下面，X会成为Y的弟弟，当YID为空时，直接把X内容拉到最顶层。排序数字`sort_num`越大，排越后，本质是自动维护一个递增的排序队列。


## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

X内容不存在:

```
{
  "flag": false,
  "cid": "fa465db45b6f4f2187defd8b7857e5d5",
  "error": {
    "id": 110000,
    "msg": "content not found:x content not found"
  }
}
```


Y内容不存在：

```
{
  "flag": false,
  "cid": "fa465db45b6f4f2187defd8b7857e5d5",
  "error": {
    "id": 110000,
    "msg": "content not found:y content not found"
  }
}
```

内容X和Y不在同一层：

```
{
  "flag": false,
  "cid": "ea01f4f3edc648c885e320031343c0b6",
  "error": {
    "id": 110005,
    "msg": "contents are in different node"
  }
}
```

内容X和Y不能一样：

```
{
  "flag": false,
  "cid": "196a6bfd4a424ce4bd6dd714041846d5",
  "error": {
    "id": 100010,
    "msg": "paras input not right:xid=yid not right"
  }
}
```