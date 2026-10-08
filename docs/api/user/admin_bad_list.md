# /user/admin/bad/list

管理员列出用户举报记录（管理后台「举报管理 → 用户举报」）。

## 请求

```
POST /user/admin/bad/list

{
	"status": -1,
	"limit": 10,
	"page": 1,
	"sort": ["-id"]
}
```

| 字段 | 含义 | 类型 | 必选 |
|------|------|------|------|
| status | 被举报用户当前状态：-1 全部 / 0 未激活 / 1 正常 / 2 拉黑 | int | 否 |
| sort | 排序（默认按 id 降序，最新在前） | []string | 否 |
| limit / page | 分页 | int | 否 |

## 响应

```
{
  "flag": true,
  "data": {
    "bad": [
      {
        "id": 1,
        "user_id": 6,          // 举报人
        "user_name": "vip_05",
        "user_nick": "VIP05",
        "bad_user_id": 7,      // 被举报用户
        "bad_user_name": "vip_06",
        "bad_user_nick": "VIP06",
        "bad_user_status": 1,  // 0 未激活 1 正常 2 拉黑
        "reason": "垃圾广告",
        "create_time": 1788100000
      }
    ],
    "total": 1,
    "total_pages": 1
  }
}
```

管理员可对「被举报用户」执行拉黑 / 解除拉黑（调用 `/user/admin/update` 设置 status=2 / status=1）。
