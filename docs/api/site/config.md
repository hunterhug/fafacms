# /site/config

获取站点配置 + 可见友情链接。无需授权。

字段含义：`site_title` 站点标题；`site_subtitle` 站点副标题（拼在标题后面，用于浏览器标题栏等）；
`site_intro` 社区介绍（首页「社区公告」与发现页那句「是一个围绕内容互动的社区…」，不含站名）；
`footer_intro` 页脚介绍。三个新字段为空时服务端回落默认值。

## 请求

不需要前缀。

```
POST /site/config

{}
```

## 响应

返回站点标题、站点副标题、社区介绍、底部介绍、未隐藏的友情链接（按 sort_num 升序）：

```
{
  "flag": true,
  "data": {
    "site_title": "花花世界",
    "site_subtitle": "发现更大的世界",
    "site_intro": "是一个围绕内容互动的社区：写文章、交朋友、发现世界",
    "footer_intro": "花花世界 · 一个围绕内容互动的社区",
    "friend_links": [
      { "id": 1, "name": "GitHub", "url": "https://github.com/hunterhug/fafacms", "sort_num": 0, "hide": 0, "open_new": 1, "create_time": 1788100000 }
    ]
  }
}
```
