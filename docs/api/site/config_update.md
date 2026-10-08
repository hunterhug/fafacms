# /site/config/update

更新站点配置（标题 / 副标题 / 社区介绍 / 底部介绍）。管理员接口。

## 请求

需要前缀 `/v1`。

```
POST /v1/site/config/update

{
	"site_title": "花花世界",
	"site_subtitle": "发现更大的世界",
	"site_intro": "是一个围绕内容互动的社区：写文章、交朋友、发现世界",
	"footer_intro": "花花世界 · 一个围绕内容互动的社区"
}
```

参数说明：

| 字段 | 含义 | 类型 | 必选 |
| --- | --- | --- | --- |
| site_title | 站点标题 | string | 是 |
| site_subtitle | 站点副标题（浏览器标题栏用「标题 - 副标题」） | string | 否，留空回落默认值 |
| site_intro | 社区介绍（首页「社区公告」与发现页，不含站名） | string | 否，留空回落默认值 |
| footer_intro | 页脚介绍 | string | 否，留空回落默认值 |

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| site_title | 站点标题 | string | 是 |
| footer_intro | 底部介绍 | string | 否 |

## 响应

正常：

```
{
  "flag": true
}
```
