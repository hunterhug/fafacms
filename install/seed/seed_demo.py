#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""FaFaCMS 演示数据填充脚本（走真实 API，密文由模型钩子自动生成）。

背景：加密改造后不能再用明文 SQL 灌数据（那样 *_bidx 为空、且数据是明文），
所以本脚本一律通过线上接口创建，保证"落库即密文、盲索引同步"。

用法：
    python3 seed_demo.py                 # 使用默认地址 http://127.0.0.1:8080
    FAFA_BASE=http://127.0.0.1:8080 python3 seed_demo.py

依赖：仅标准库（urllib / hmac / hashlib / zlib / struct）。
"""
import hashlib
import hmac
import json
import os
import secrets
import sys
import time
import urllib.error
import urllib.request
import zlib
import struct

import fafaenv

BASE = fafaenv.BASE
SIGN_SECRET = fafaenv.SIGN_SECRET.encode()
ADMIN_NAME = fafaenv.ADMIN_NAME
ADMIN_PASS = fafaenv.ADMIN_PASS
USER_PASS = fafaenv.DEMO_PASS

# ---------------------------------------------------------------------------
# HTTP（带请求签名）
# ---------------------------------------------------------------------------


def _request(method, path, body=None, token=None, ctype=None, raw_body=None):
    ts = str(int(time.time()))
    nonce = secrets.token_hex(8)
    sign = hmac.new(SIGN_SECRET, f"{ts}:{nonce}".encode(), hashlib.sha256).hexdigest()
    data = raw_body
    if data is None and body is not None:
        data = json.dumps(body, ensure_ascii=False).encode()
    req = urllib.request.Request(BASE + path, data=data, method=method)
    req.add_header("X-Ts", ts)
    req.add_header("X-Nonce", nonce)
    req.add_header("X-Sign", sign)
    if ctype:
        req.add_header("Content-Type", ctype)
    if token:
        req.add_header("Auth", token)
    try:
        with urllib.request.urlopen(req, timeout=120) as r:
            return r.status, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()


THROTTLE_SEC = float(os.environ.get("FAFA_SEED_THROTTLE", "0.4"))
_last_call = [0.0]


def api(method, path, body=None, token=None, retry=3):
    """带节流的 API 调用；命中反爬限流（100034）时退避重试。"""
    for attempt in range(retry + 1):
        gap = THROTTLE_SEC - (time.time() - _last_call[0])
        if gap > 0:
            time.sleep(gap)
        _last_call[0] = time.time()
        _, raw = _request(method, path, body, token, ctype="application/json")
        try:
            resp = json.loads(raw.decode("utf-8", "replace"))
        except Exception:
            return {"_raw": raw[:200].decode("utf-8", "replace")}
        if resp.get("error", {}).get("id") == 100034 and attempt < retry:
            wait = 8 * (attempt + 1)
            print(f"  ~ 触发限流，{wait}s 后重试 {path}")
            time.sleep(wait)
            continue
        return resp
    return resp


def must(label, resp):
    """断言接口成功，失败直接终止，避免灌出半截数据。"""
    if not resp.get("flag"):
        print(f"  ✗ {label} 失败: {json.dumps(resp, ensure_ascii=False)[:400]}")
        sys.exit(1)
    return resp


def find_user(name):
    """已存在则返回 (id, head_photo)，否则 (None, "")。"""
    d = api("GET", "/u/info", {"user_name": name}).get("data") or {}
    if not d.get("id"):
        return None, ""
    return d["id"], d.get("head_photo", "")


def find_node(user_name, seo):
    d = api("GET", "/u/nodes", {"user_name": user_name, "limit": 100}).get("data") or {}
    for n in d.get("nodes") or []:
        if n.get("seo") == seo:
            return n.get("id")
    return None


def find_content(user_name, seo):
    d = api("GET", "/u/content", {"user_name": user_name, "limit": 200}).get("data") or {}
    for c in d.get("contents") or []:
        if c.get("seo") == seo:
            return c.get("id")
    return None


def redis_del_pattern(pattern):
    """清理风控计数（评论限流）；由 fafaenv 统一处理凭据，未配置 Redis 密码时静默跳过。"""
    return fafaenv.clear_limits([pattern])


def upload(token, filename, content, ftype="image", describe=""):
    boundary = "----fafa" + secrets.token_hex(8)
    parts = []
    for k, v in (("type", ftype), ("tag", "demo"), ("describe", describe)):
        parts.append(f"--{boundary}\r\nContent-Disposition: form-data; name=\"{k}\"\r\n\r\n{v}\r\n".encode())
    parts.append(
        f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"{filename}\"\r\n"
        f"Content-Type: application/octet-stream\r\n\r\n".encode() + content + b"\r\n")
    parts.append(f"--{boundary}--\r\n".encode())
    _, raw = _request("POST", "/v1/file/upload", token=token,
                      ctype=f"multipart/form-data; boundary={boundary}", raw_body=b"".join(parts))
    try:
        return json.loads(raw.decode("utf-8", "replace"))
    except Exception:
        return {"_raw": raw[:200].decode("utf-8", "replace")}


# ---------------------------------------------------------------------------
# 生成一张有辨识度的头像/封面 PNG（纯色底 + 对角条纹 + 中心方块）
# ---------------------------------------------------------------------------


def make_png(size, base_rgb, accent_rgb, seed=0):
    rows = []
    for y in range(size):
        row = bytearray(b"\x00")
        for x in range(size):
            stripe = ((x + y + seed * 7) // (size // 6 or 1)) % 2 == 0
            inner = size // 4 <= x < size * 3 // 4 and size // 4 <= y < size * 3 // 4
            c = accent_rgb if inner else (base_rgb if stripe else accent_rgb)
            row += bytes(c)
        rows.append(bytes(row))
    raw = b"".join(rows)

    def chunk(tag, data):
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)

    return (b"\x89PNG\r\n\x1a\n"
            + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 2, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(raw, 6))
            + chunk(b"IEND", b""))


PALETTE = [
    ((0x2F, 0x4F, 0x6F), (0x8F, 0xC1, 0xE8)),
    ((0x6B, 0x3F, 0x2E), (0xE8, 0xB4, 0x8F)),
    ((0x2E, 0x5B, 0x45), (0x9C, 0xD8, 0xB4)),
    ((0x5B, 0x2E, 0x5B), (0xD8, 0x9C, 0xD8)),
    ((0x3F, 0x3F, 0x3F), (0xC9, 0xC9, 0xC9)),
    ((0x6B, 0x5B, 0x1E), (0xE8, 0xDC, 0x8F)),
    ((0x1E, 0x4B, 0x5B), (0x8F, 0xD8, 0xE8)),
    ((0x5B, 0x1E, 0x2E), (0xE8, 0x8F, 0x9C)),
]

# ---------------------------------------------------------------------------
# 演示数据定义
# ---------------------------------------------------------------------------

USERS = [
    {
        "name": "linxiaoyu", "nick": "林小雨", "email": "linxiaoyu@fafacms.local",
        "gender": 2, "wechat": "linxiaoyupic", "github": "https://github.com/linxiaoyu",
        "short": "前端工程师 / 街头摄影爱好者",
        "describe": "写了六年 Web，最近三年主要在做设计系统与前端性能。业余时间背着相机在老城区闲逛，"
                    "喜欢记录光影和普通人的日常。相信「好的界面应该让人感觉不到界面的存在」。",
    },
    {
        "name": "chenmo", "nick": "陈默", "email": "chenmo@fafacms.local",
        "gender": 1, "github": "https://github.com/chenmo", "qq": "100200300",
        "short": "后端开发，分布式系统与存储",
        "describe": "做基础设施的，日常和一致性、复制、故障恢复打交道。喜欢把复杂问题拆成能验证的小假设，"
                    "也喜欢把踩过的坑写成文档——因为半年后的自己一定会感谢现在的自己。",
    },
    {
        "name": "anan", "nick": "安安", "email": "anan@fafacms.local",
        "gender": 2, "wechat": "ananpm", "weibo": "https://weibo.com/anan",
        "short": "产品经理，喜欢写字",
        "describe": "在 B 端和 C 端之间反复横跳的产品。坚信需求文档写得越短，团队跑得越快。"
                    "下班后会写点不成气候的散文，主题大多是雨、旧书店和没赶上末班车的人。",
    },
    {
        "name": "zhouyu", "nick": "周渔", "email": "zhouyu@fafacms.local",
        "gender": 1, "weibo": "https://weibo.com/zhouyu", "qq": "400500600",
        "short": "插画师 / 视觉设计",
        "describe": "画了十年画，从纸笔到数位板再到 iPad。擅长低饱和的叙事插画，最近在尝试把中国画里的"
                    "留白语言用到界面插图里。接稿不多，喜欢慢慢画。",
    },
    {
        "name": "haitang", "nick": "海棠", "email": "haitang@fafacms.local",
        "gender": 2, "github": "https://github.com/haitang", "wechat": "haitangdev",
        "short": "独立开发者，做小工具",
        "describe": "一个人从需求、设计、开发到客服全包的那种。做过记账、番茄钟、播客剪辑三个小产品，"
                    "都还活着。信奉「小而美」和「发布比完美重要」。",
    },
    {
        "name": "laowang", "nick": "老王", "email": "laowang@fafacms.local",
        "gender": 1, "github": "https://github.com/laowang-ops", "qq": "700800900",
        "short": "运维工程师，Linux 与容器",
        "describe": "管过三千台机器的运维老兵。喜欢用最土的办法解决问题，也喜欢把事故复盘写成"
                    "别人能照着做的检查清单。最近在研究内核参数和 IO 调度。",
    },
    {
        "name": "suyi", "nick": "苏一", "email": "suyi@fafacms.local",
        "gender": 1, "github": "https://github.com/suyi-algo",
        "short": "计算机专业学生，算法竞赛退役选手",
        "describe": "打过三年 ACM，最好成绩是区域赛银。现在在补基础：操作系统、编译原理、概率论。"
                    "习惯把每道题的思考过程写下来，因为「写不出来就是没想清楚」。",
    },
    {
        "name": "nanke", "nick": "南柯", "email": "nanke@fafacms.local",
        "gender": 2, "weibo": "https://weibo.com/nanke", "wechat": "nanketrip",
        "short": "旅行博主，在路上",
        "describe": "去过 27 个省，最喜欢西北的风和西南的雨。写游记不写攻略，只写当时的心情和遇到的人。"
                    "相机里存着两万张照片，真正满意的不到一百张。",
    },
]

# 专栏： (归属用户名, seo, 名称, 描述)
NODES = [
    ("linxiaoyu", "photographynotes", "光影笔记", "街头摄影的观察、器材心得，以及照片背后的故事。"),
    ("linxiaoyu", "frontendnotes", "前端手记", "工程实践、性能优化与设计系统的落地经验。"),
    ("chenmo", "distributednotes", "分布式笔记", "一致性、复制与故障恢复，把踩过的坑写清楚。"),
    ("anan", "productthoughts", "产品随想", "做产品的一些判断，以及判断错之后的反思。"),
    ("zhouyu", "illustration", "插画与设计", "插画创作过程、配色思路和一些不成熟的观点。"),
    ("haitang", "indiedevlog", "独立开发日志", "一个人做产品的完整过程，包含失败的那部分。"),
    ("laowang", "opshandbook", "运维手记", "Linux、容器与线上事故复盘，尽量写成可执行的清单。"),
    ("suyi", "algonotes", "算法题解", "把每道题的思考过程写完整，而不是只贴代码。"),
    ("nanke", "ontheroad", "在路上", "游记与人，不写攻略。"),
]

# 文章： (归属用户名, 专栏 seo, 文章 seo, 标题, 正文)
ARTICLES = [
    ("linxiaoyu", "photographynotes", "morningmarketlight", "清晨菜市场的光，是我见过最诚实的光",
     """早上六点半的菜市场，光线是斜着切进来的。

摊位上方的白炽灯还没关，冷白的光和刚亮起来的天光混在一起，落在塑料布和青菜叶子上，颜色脏得很，但又意外地真实。我蹲在卖豆腐的摊子边拍了二十分钟，老板终于忍不住问我在拍什么。

我说在拍光。

他说这有什么好拍的。然后他自己也抬头看了一眼头顶的灯泡。

## 关于器材

那天用的是 35mm 定焦，光圈开到 f/2.8。菜市场进深不大，35 段刚好能容下一个摊位加一个人；再长一点，就只剩下商品，没有人了。

- 曝光补偿 −0.7，避免白炽灯区域过曝
- 白平衡锁在 4000K 左右，保留一点冷调
- 全程手动对焦，自动对焦在暗部会来回拉风箱

## 后来

这组照片我挑了三张，都是在同一个摊位、前后不超过五分钟拍的。摊主从抗拒到习惯，最后一张他甚至笑了一下。

比技术更难的是让别人接受被拍。我通常的做法是先买点东西，聊两句，再问能不能拍。十次里有七八次会被答应。"""),

    ("linxiaoyu", "frontendnotes", "designsystemmigration", "把 12 万行代码迁到设计系统，我们踩过的五个坑",
     """接手这个项目时，产品线上有 4 个前端仓库，按钮组件有 11 种实现，颜色写死了 63 个十六进制值。

迁移花了五个月。这里记录最疼的五个坑。

## 一、先对齐"什么是组件"，再动手

我们一开始就急着抽 `Button`、`Input`，结果两周后返工——因为设计侧说的「主按钮」和开发侧理解的「主按钮」不是一回事：一个是语义层级，一个是视觉权重。

**教训**：先花一周把设计 token 和组件语义表对齐，比后面返工三次便宜得多。

## 二、不要试图一次性替换

我们试过按仓库替换，第一个仓库改到一半就顶不住了：新老组件在同一页面上并存，样式互相污染。

后来改成**按页面替换**，一个页面彻底切干净再进下一个。虽然慢，但每天都在往前走。

## 三、把"历史包袱"变成数据

我们写了个脚本扫描所有仓库，统计每个组件的使用次数：

```bash
rg -o "components/[A-Za-z]+" --no-filename src | sort | uniq -c | sort -rn
```

结果很反直觉：11 个 Button 实现里，有 6 个只被用过 1~3 次。删掉它们没有任何风险，我们却为它们纠结了两周。

## 四、样式覆盖是最大的敌人

新组件上线后，老仓库的全局样式会把它们改得面目全非。我们的解法是给设计系统加一层 `@layer`，并禁止在业务侧使用 `!important`——用 lint 规则强制。

## 五、别忘了写迁移文档

最没技术含量、但收益最高的一件事：给每个组件写一页「老写法 → 新写法」对照表。它让迁移不再是只有两个人会做的事。

## 结果

- 组件数：从 147 个降到 68 个
- 打包体积：主包减小 31%
- 新页面开发时间：从平均 3 天降到 1.5 天

最重要的收获是：**迁移的目标不是代码变漂亮，而是让下一次改需求变便宜**。"""),

    ("chenmo", "distributednotes", "raftjointconsensus", "Raft 的联合共识：为什么不能一次加两个节点",
     """集群从 3 节点扩到 5 节点，如果直接改配置，会出现两个不相交的多数派同时选出 leader 的窗口。

原因很简单：新旧配置的多数派计算方式不同。

- 旧配置 3 节点：多数派 = 2
- 新配置 5 节点：多数派 = 3
- 如果 leader 先改自己的配置再广播，可能 2 个旧节点 + 1 个新节点各自认为自己是多数派

联合共识（joint consensus）的做法是引入一个中间状态 `C_old,new`，任何决策**必须同时获得新旧两个配置的多数派同意**。

## 具体流程

1. leader 收到成员变更请求，写入 `C_old,new` 日志并复制
2. `C_old,new` 提交后，leader 写入 `C_new` 日志
3. `C_new` 提交后，变更完成，可以安全下线旧节点

关键点是第 2 步：**leader 在 `C_old,new` 提交之后、`C_new` 提交之前，不属于任何一个配置的多数派时依然要正常服务**——它只需要能在新配置里被选出来即可。

## 常见的错误实现

| 错误 | 后果 |
|---|---|
| 一次变更多个节点 | 多数派可能不相交，脑裂 |
| 不做联合共识，直接切换配置 | 同上 |
| 只在 leader 本地改配置不落日志 | leader 挂掉后配置回滚，状态不一致 |

## 我们的实践

生产环境一次只加/减一个节点，并在变更期间监控 `raft_config_change_duration`。如果超过 30 秒，告警——通常意味着有个节点在反复重试。"""),

    ("chenmo", "distributednotes", "idempotentconsumer", "消息队列的幂等消费，别只靠唯一索引",
     """「用唯一索引兜住重复」是很常见的答案，但它有三个前提，缺一个就漏。

## 前提一：业务写入和幂等记录在同一个事务

如果先写业务表、再写去重表，中间进程被杀，重复消息会再次通过检查。

正确做法是两者在同一个本地事务里，或者干脆把去重键做成业务表的唯一索引。

## 前提二：重复消息的"业务效果"是可合并的

唯一索引能挡住「插入两条订单」，挡不住「库存扣了两次」——因为后者的效果是累加。

对于累加类操作，要么改成幂等的赋值语义（`SET stock = 100 - n` 而不是 `stock = stock - 1`），要么引入版本号做 CAS。

## 前提三：消费位点和业务提交的原子性

这个最容易被忽略。Kafka 的 offset 提交如果晚于业务提交，进程崩溃后消息会重放；如果早于业务提交，消息会丢。

我们的做法是**业务数据里记录 offset**，重启时从业务数据里恢复位点：

```sql
-- 处理消息时一并写入
INSERT INTO orders (id, biz_id, kafka_offset, ...) VALUES (...);
-- 启动时
SELECT MAX(kafka_offset) FROM orders WHERE partition = ?;
```

## 小结

唯一索引是必要不充分的。判断一套消费逻辑是否真的幂等，可以问自己：**这条消息处理两次，数据库的最终状态和只处理一次一样吗？**"""),

    ("anan", "productthoughts", "whyusersleave", "用户不是流失的，是被我们推走的",
     """复盘一个日活掉了 40% 的产品，我们发现「流失」这个词用错了。

用户不是慢慢淡忘的，而是在某个具体的时刻被推走的。把最近三个月的用户访谈和埋点对齐之后，我们找到了四个推手。

## 一、第一次使用时的等待

新用户注册后平均要等 **11 秒**才能看到第一个有效内容。这 11 秒里我们做的是同步拉取用户的历史数据。

**改法**：先渲染空状态 + 骨架屏，数据回来后局部刷新。改完次留提升了 6 个百分点。

## 二、一次失败后的不解释

支付失败页只写「操作失败，请重试」。而实际上 62% 的失败是银行卡限额，用户重试一百次也没用。

**改法**：把失败原因翻译成人话，并给出替代路径。这一条单独带来了 1.8% 的支付成功率提升。

## 三、被强制的"升级"

我们在没有告知的情况下把免费额度从 100 次降到 30 次。当月退了 2100 个订阅。

**改法**：额度调整必须提前 30 天站内信 + 邮件告知，且老用户保留原额度。

## 四、找不到"上一次做的事"

用户的核心场景是"继续上次的项目"，但我们把入口藏在二级菜单里。

**改法**：首页第一屏直接放「最近打开」。

## 一点方法论

与其看流失曲线，不如看**流失前最后一次会话的完整回放**。曲线告诉你"少了多少人"，回放告诉你"为什么少"。"""),

    ("anan", "productthoughts", "shortprd", "需求文档写到三页就够了",
     """我见过 60 页的 PRD，也见过 3 页的 PRD。后者交付质量明显更好。

不是因为写得少，而是因为**写不下废话**。

## 三页 PRD 的结构

**第一页：问题与不做的事**

- 谁在什么场景下遇到什么问题（一句话）
- 现在的替代方案是什么，为什么不够好
- 这次**不做**什么（这一条比做更重要）

**第二页：方案与边界**

- 用户能看到的完整流程（画图，不写字）
- 异常路径有哪几条，分别怎么处理
- 涉及的数据、权限、合规边界

**第三页：验收与度量**

- 上线后看哪两个指标
- 什么情况下回滚

## 为什么有效

页面多了，作者会不自觉地用文字填补思考的空洞。限制在三页，会强迫你在动笔前把问题想清楚。

另外，短文档让**评审**变得可能。没人会认真读完 60 页，但每个人都会读完 3 页。

## 反例

不是所有需求都适合三页。涉及多方对账、资金清算、监管报备的需求，附录该长还是要长。但**正文**依然应该在三页内——附录是给实现者查的，正文是给决策者看的。"""),

    ("zhouyu", "illustration", "colorrestraint", "低饱和配色，不是把颜色调灰那么简单",
     """很多人以为低饱和就是把饱和度滑块往左拉。这样出来的画面只会脏。

真正起作用的是**明度关系**和**色相距离**。

## 明度先于色相

先只用黑白灰画出明暗层次。如果黑白稿里主体和背景分不开，上色也救不回来。

我通常会准备三级明度：主体最亮或最暗（占画面 20%），中间调承担大部分面积（60%），剩下一层做氛围（20%）。

## 色相不要超过三个

低饱和画面里，色相越多越糊。我的习惯是：

- 主色相（占 60%）
- 辅色相（占 30%，与主色相相距 30°~60°）
- 点缀色（占 10%，可以跳到互补色）

## 用"脏色"制造空气感

纯粹的灰色会显得死。在暗部里加一点冷色（比如群青 + 熟褐），亮部里加一点暖色（镉黄 + 一点点朱红），画面立刻有了空气。

## 一个练习

找一张自己喜欢的低饱和插画，用取色器取 10 个点，记下 HSV 值。你会发现：

- 饱和度很少低于 8%，也很少高于 45%
- 明度的跨度往往比想象中大得多

**结论**：低饱和的秘诀不在饱和度，而在明度层次的丰富度。"""),

    ("haitang", "indiedevlog", "firstproductpostmortem", "我的第一个产品，死在「再改一版」上",
     """第一个产品做了 14 个月，上线 3 个月后停止维护。死因不是技术，是我一直在改。

## 时间线

- 第 1~3 月：很兴奋，核心功能两周就做完了
- 第 4~8 月：觉得 UI 不够好，重做了两次
- 第 9~12 月：觉得架构不行，重写了一遍
- 第 13 月：上线
- 第 14 月：发现没人用

## 真正的问题

我在第 3 个月就有了可以给用户用的版本，但我没给。因为"还不够好"。

而"够不够好"这件事，**只有用户能回答**，我自己回答不了。我用一年时间回答了一个我无法回答的问题。

## 现在的做法

- 从想法到第一个可用的版本，不超过 2 周
- 第 3 周必须给至少 5 个真人用
- 之后每次改动都必须基于一条具体的用户反馈，否则不做
- 每 90 天做一次"停还是继续"的判断

## 一句话

**发布不是终点，是获取信息的开始。** 拖着不发布，等于一直在没有信息的情况下做决策。"""),

    ("haitang", "indiedevlog", "pricingsmalltools", "小工具的定价：我从 9 元改到 39 元之后",
     """小工具该定多少钱？我试过 9 元、19 元、39 元、免费+捐赠，最后停在 39 元一年。

## 数据

| 定价 | 月销量 | 月收入 | 退款率 |
|---|---|---|---|
| 9 元买断 | 210 | 1890 | 1.2% |
| 19 元买断 | 95 | 1805 | 2.1% |
| 39 元/年 | 62 | 2418 | 0.8% |
| 免费 + 捐赠 | — | 约 150 | — |

## 几个反直觉的结论

**一、9 元和 19 元的收入差不多，但 19 元的用户更认真。**

低价的用户里有相当一部分是"顺手买了试试"，这部分人不会提反馈，也不会传播。

**二、订阅比买断更适合小工具。**

不是因为收入高，而是因为它**强制我持续维护**。买断制下，第二年我就没有动力修 bug 了。

**三、退款率随价格下降。**

39 元档的退款率只有 9 元档的 2/3。愿意为工具付费 39 元的人，通常是真的有需求。

## 现在的定价原则

- 只做一个付费档，不做三档（省掉 90% 的定价纠结）
- 价格锚在"用户因此省下的时间值多少钱"，不是"我投入了多少时间"
- 允许无条件退款，不做挽留流程

## 最后

定价是可以改的。**先定一个，观察三个月，再改。** 比纠结一个月不动要好得多。"""),

    ("laowang", "opshandbook", "diskfullchecklist", "磁盘满导致的事故，80% 是可以提前发现的",
     """线上磁盘写满，MySQL 拒绝写入，业务雪崩。这次事故的直接原因是 binlog 暴涨，但根因是我们从来没有认真看过磁盘的增长趋势。

## 事故链条

1. 02:14 某台从库磁盘使用率到 95%
2. 02:31 达到 100%，MySQL 报 `Disk full`
3. 02:33 主库复制延迟飙升，应用连接池打满
4. 02:35 接到告警

**问题在于：为什么 95% 的时候没有告警？** 因为我们只配了 98% 的静态阈值，而报警到 100% 之间只有 17 分钟。

## 改进后的检查清单

**必配告警**

- 使用率 > 80%
- **增长率** > 5%/小时（这条最关键，它能提前几小时发现异常）
- inode 使用率 > 80%（很多人只看容量不看 inode）

**必查项**

```bash
df -h              # 容量
df -i              # inode
du -sh /var/lib/mysql/* | sort -rh | head
ls -lh /var/lib/mysql/binlog/ | tail
```

**必做配置**

- binlog 过期时间：`binlog_expire_logs_seconds = 604800`（别用默认的永久保留）
- 慢日志、错误日志做 logrotate
- 容器日志限制：`--log-opt max-size=100m --log-opt max-file=3`

## 应急

留一条**不依赖数据库**的应急通道：

1. 清 binlog：`PURGE BINARY LOGS BEFORE DATE_SUB(NOW(), INTERVAL 3 DAY);`
2. 清临时文件
3. 实在不行，先扩容再排查

千万不要在业务高峰期**一边清文件一边不改配置**——第二天还会再来一次。"""),

    ("laowang", "opshandbook", "containermemorylimit", "容器里看内存，别信 free",
     """容器里跑 `free -h`，看到的是**宿主机**的内存，不是这个容器的限额。用它来判断要不要扩容，会出大问题。

## 为什么

`free` 读的是 `/proc/meminfo`，而 lxcfs 默认没有挂载时，这个文件就是宿主机的。

## 正确的做法

**cgroup v1**：

```bash
cat /sys/fs/cgroup/memory/memory.limit_in_bytes
cat /sys/fs/cgroup/memory/memory.usage_in_bytes
cat /sys/fs/cgroup/memory/memory.stat | grep -E 'pgmajfault|total_rss'
```

**cgroup v2**：

```bash
cat /sys/fs/cgroup/memory.max
cat /sys/fs/cgroup/memory.current
```

**更省事的办法**：直接挂载 lxcfs，让容器内的 `free` 和 `/proc/meminfo` 显示为容器视角。

## 一个真实案例

某服务容器限额 2Gi，`free` 显示宿主机 128G 空闲，于是 JVM 默认把最大堆设成了宿主机内存的 1/4 = 32G。结果容器一启动就被 OOM Kill，日志里没有任何 Java 异常。

**修法**：显式设置 `-XX:MaxRAMPercentage`，或者用 `-XX:MaxRAM=1500m`。

## 检查清单

- [ ] 容器内存限额已设置（不设置等于无限制，会被邻居影响）
- [ ] 应用感知到了限额（JVM / Go GOMEMLIMIT / Node `--max-old-space-size`）
- [ ] 有 `memory.usage_in_bytes` 的监控
- [ ] OOM Kill 有告警（`dmesg | grep -i "killed process"`）

**记住**：容器不是虚拟机，它看到的"系统信息"可能是假的。"""),

    ("suyi", "algonotes", "monotonicstack", "单调栈：从「下一个更大元素」到「接雨水」",
     """单调栈的核心只有一句话：**维护一个"还没找到答案"的候选集合，并保持它的单调性。**

## 模板

以"下一个更大元素"为例：

```python
def next_greater(nums):
    res = [-1] * len(nums)
    stack = []          # 存下标，对应的值单调递减
    for i, x in enumerate(nums):
        while stack and nums[stack[-1]] < x:
            res[stack.pop()] = x
        stack.append(i)
    return res
```

每个元素最多进栈一次、出栈一次，所以是 O(n)。

## 为什么正确

弹出 `j` 的那一刻，`x` 就是 `j` 右边第一个比它大的元素——因为栈里所有比 `x` 小的都已经被更早的元素弹走了，而栈的单调性保证了 `j` 下面压着的都比 `j` 大。

## 迁移到"接雨水"

接雨水的关键观察是：**每个位置的水量 = min(左边最高, 右边最高) - 自身高度**。

用单调递减栈，当遇到比栈顶高的柱子时，栈顶就是"洼地的底"，它左边是栈里下一个元素，右边是当前元素：

```python
def trap(height):
    ans, stack = 0, []
    for i, h in enumerate(height):
        while stack and height[stack[-1]] < h:
            bottom = stack.pop()
            if not stack:
                break
            left = stack[-1]
            width = i - left - 1
            ans += width * (min(height[left], h) - height[bottom])
        stack.append(i)
    return ans
```

## 什么时候想到单调栈

题目里出现这些信号时，优先考虑：

- "左边/右边第一个比它大/小的元素"
- 需要计算以某个元素为边界的区间贡献
- 直方图、柱状图相关的面积问题

## 易错点

1. **存值还是存下标**：需要算宽度时存下标
2. **严格 vs 非严格**：重复元素时用 `<` 还是 `<=` 会改变结果，取决于题目对"更大"的定义
3. **弹出后栈空**：说明左边没有边界，直接 break"""),

    ("suyi", "algonotes", "dpspaceoptimization", "动态规划的空间优化：滚动数组之外还有什么",
     """背包问题把二维数组压成一维，几乎所有人都会。但压缩之后**遍历方向**就成了陷阱。

## 01 背包为什么倒序

```python
for i in range(n):
    for j in range(W, w[i] - 1, -1):     # 倒序
        dp[j] = max(dp[j], dp[j - w[i]] + v[i])
```

倒序保证 `dp[j - w[i]]` 还是**上一轮**的值（即"不选第 i 件"的状态）。

正序的话，`dp[j - w[i]]` 已经被本轮更新过，等于第 i 件被选了多次——那正好是完全背包。

**所以：01 背包倒序，完全背包正序。** 这不是技巧，是语义。

## 更一般的压缩：只依赖前若干个状态

比如"打家劫舍"只依赖 `i-1` 和 `i-2`，那只需要两个变量：

```python
prev2 = prev1 = 0
for x in nums:
    prev2, prev1 = prev1, max(prev1, prev2 + x)
```

## 状态压缩的前提

压缩的本质是**丢弃不再被需要的信息**。做之前先问：

- 下一轮状态的计算依赖哪几层？
- 被丢弃的层，后面还会不会用到？

如果答案不清楚，**不要压缩**。多一个维度换来的是"能写对"，这在比赛里比省几十 MB 重要得多。

## 什么时候不该省

- 需要输出方案（要回溯）
- 需要多个维度的历史（如区间 DP）

写不出来的时候，先把完整状态写出来跑对，再考虑压。"""),

    ("nanke", "ontheroad", "qilianwind", "祁连山下的风，吹了三天",
     """从张掖往西，过了民乐就是祁连山脚下。

第一天住在山丹，县城小得半小时能走完。傍晚在汽车站旁边吃牛肉小饭，老板是个五十来岁的女人，问我一个人来干什么。我说看山。她笑了笑，说山一直在那儿。

## 第二天

早上六点出发去军马场。车在草原上开了两个小时，除了偶尔的羊群，什么都没有。

海拔三千米，风大得人站不稳。草是黄的，天是蓝得发黑的那种蓝，云跑得很快。我坐在一个土坡上待了很久，直到手指冻得按不下快门。

同行的一个牧民说，这片草原十年前的草比现在高，能没过膝盖。

## 第三天

翻过扁都口，进入青海境内。海拔一路升到 3685 米，路边开始有雪。

在垭口遇到三个骑行者，从西宁骑往张掖，已经骑了四天。他们分给我半瓶热水，我们在风里站着聊了十分钟，然后各自上路。

## 关于"看山"

有人问我，跑这么远就看几座山，值不值。

我想不清楚这个问题。但如果一定要回答：**值不值不是山决定的，是那三天里不用想别的。** 风一直在吹，除了把外套裹紧，你什么都不用做。"""),

    ("nanke", "ontheroad", "oldbookstore", "县城里的旧书店，和看店的老人",
     """在川东一个县城，我误打误撞进了一家旧书店。

门脸很小，招牌上的字掉了漆。推门进去是一股纸张受潮又晒干的味道。三面墙都是书架，从地板顶到天花板，中间只留一条能侧身走过的通道。

## 老人

看店的是个七十多岁的老先生，戴着眼镜在看报纸。我问他书怎么卖，他说自己找，找到了论斤称。

我愣了一下。他解释：这些书都是收废品收来的，论本算不划算。

## 找到的三本

- 一本 1982 年的《中国自然地理》，扉页上写着"赠给李建国同志"
- 一本 1979 年的《新华字典》，边角磨圆了，里面有铅笔画的下划线
- 一本没写年份的诗集，最后一页有人用钢笔抄了一首诗，字很好看

三本一共九块钱。

## 聊天

我问他这些书卖不掉怎么办。他说卖不掉就放着，"放着也是放着"。

他说他原来在县中学教语文，退休后没别的事做，就守着这个店。来的大多是老人，有时候一整天没人进来。

"现在的年轻人不看书了。"他说这话的时候没有抱怨的语气，像是在说天气。

## 离开

我在店里待了两个多小时。走的时候他站起来送我，说"常来"。

那个县城我大概不会再去了。但每次在书架上看到那本《中国自然地理》，都会想起下午三点的光和那股纸的味道。"""),
]

# 评论： (文章 seo 前缀, 评论者用户名, 内容)
COMMENTS = [
    ("morningmarketlight", "nanke", "菜市场那组照片我特别喜欢，尤其是摊主从抗拒到笑的那张。想问下你一般怎么开口？"),
    ("morningmarketlight", "zhouyu", "光确实是斜着切进来的那种感觉，第二张的暗部处理得很舒服，没有硬拉亮。"),
    ("morningmarketlight", "anan", "「比技术更难的是让别人接受被拍」——这句话可以直接放进我的产品文档里。"),
    ("designsystemmigration", "haitang", "按页面替换这个建议太实用了。我上次按组件替换，做到一半样式全乱了。"),
    ("designsystemmigration", "chenmo", "扫描统计使用次数这一步很关键，我们做后端接口清理时也是同样的思路，先看调用量再决定删不删。"),
    ("designsystemmigration", "suyi", "请问 @layer 那部分有推荐的 lint 配置吗？我们项目也想禁掉 !important。"),
    ("raftjointconsensus", "laowang", "运维视角补充一句：变更期间一定要限制并发，我们上次两个人同时扩缩容，直接脑裂了。"),
    ("raftjointconsensus", "suyi", "联合共识这段终于看懂了，之前一直不明白为什么要引入中间状态。"),
    ("idempotentconsumer", "chenmo", "自己回复一下：把 offset 存业务表这个做法，业务表会多一列，但换来重启时的确定性，很值。"),
    ("idempotentconsumer", "haitang", "「这条消息处理两次，最终状态和只处理一次一样吗」——这个自检问题收下了。"),
    ("whyusersleave", "linxiaoyu", "第一次使用的 11 秒这个数据太扎心了。我们最近也发现首屏同步请求是次留杀手。"),
    ("whyusersleave", "haitang", "被强制的升级这条我有切身体会，作为用户被降额度的时候真的会直接走。"),
    ("colorrestraint", "linxiaoyu", "明度先于色相，这个顺序我原来一直是反的，难怪画面显脏。"),
    ("colorrestraint", "anan", "打算照着你说的练习做一遍，取色器取十个点，看看自己的直觉差多少。"),
    ("firstproductpostmortem", "anan", "「我用一年时间回答了一个我无法回答的问题」，这句写得真好。"),
    ("firstproductpostmortem", "chenmo", "发布不是终点，是获取信息的开始。这句话应该贴在每个独立开发者的显示器上。"),
    ("pricingsmalltools", "linxiaoyu", "只做一个付费档这点我认同，三档定价我自己纠结了两个月最后也没选出来。"),
    ("diskfullchecklist", "chenmo", "增长率告警这条太重要了，静态阈值只能告诉你「已经晚了」。"),
    ("diskfullchecklist", "laowang", "补充一点：如果是云盘，还要看一下云监控的 IOPS 限额，磁盘没满但 IOPS 打满一样会挂。"),
    ("containermemorylimit", "suyi", "JVM 按宿主机内存算堆大小这个坑我们组也踩过，第一次排查花了整整一天。"),
    ("monotonicstack", "chenmo", "存下标还是存值这一条总结得好，我以前每次都要重新想一遍。"),
    ("dpspaceoptimization", "chenmo", "「01 背包倒序、完全背包正序，这不是技巧，是语义」，这句我记下了。"),
    ("qilianwind", "anan", "看到「山一直在那儿」的时候愣了一会儿。写得真好。"),
    ("oldbookstore", "zhouyu", "「放着也是放着」——这个老人和这家店，我想画下来。"),
    ("oldbookstore", "linxiaoyu", "想看你拍的旧书店的照片，有吗？"),
]

# 关注关系： (关注者, 被关注者)
FOLLOWS = [
    ("linxiaoyu", "nanke"), ("linxiaoyu", "zhouyu"), ("linxiaoyu", "anan"),
    ("chenmo", "laowang"), ("chenmo", "suyi"), ("chenmo", "linxiaoyu"),
    ("anan", "linxiaoyu"), ("anan", "haitang"), ("anan", "nanke"),
    ("zhouyu", "nanke"), ("zhouyu", "linxiaoyu"), ("zhouyu", "anan"),
    ("haitang", "anan"), ("haitang", "chenmo"), ("haitang", "laowang"),
    ("laowang", "chenmo"), ("laowang", "haitang"),
    ("suyi", "chenmo"), ("suyi", "laowang"),
    ("nanke", "linxiaoyu"), ("nanke", "zhouyu"),
]

# 私信： (发送者, 接收者, 内容)
MESSAGES = [
    ("anan", "linxiaoyu", "小雨你好，看到你写的设计系统迁移那篇，能不能约个时间聊聊？我们团队正准备做类似的事。"),
    ("linxiaoyu", "anan", "没问题，这周下午都可以。有具体想先聊的部分吗？我建议先从「什么是组件」这个对齐会说起。"),
    ("haitang", "chenmo", "陈默，你之前提的幂等消费那套方案，有没有开源出来的示例代码？想在我们的小工具里用。"),
    ("chenmo", "haitang", "有的，我整理一份脱敏版发你。核心其实就是把 offset 写进业务表，其他都是细节。"),
    ("nanke", "zhouyu", "周渔，旧书店那家店的具体位置我记在备忘录里了，回头发你，你要是去记得先打电话。"),
    ("zhouyu", "nanke", "太好了，谢谢！我打算秋天去，顺便拍点素材回来画。"),
    ("suyi", "chenmo", "师兄，单调栈那篇里的「弹出后栈空」我一开始没看懂，现在明白了，谢谢。"),
]

# 友情链接
FRIEND_LINKS = [
    ("Go 语言中文网", "https://studygolang.com"),
    ("Vue.js 官方文档", "https://cn.vuejs.org"),
    ("MySQL 官方文档", "https://dev.mysql.com/doc/"),
]

# 公告
GLOBAL_MESSAGE = "花花世界已启用数据库与本地文件加密存储。欢迎在这里记录与分享——你的文字只有你自己和站方能看到。"


def main():
    print(f"目标后端: {BASE}")
    admin = must("管理员登录", api("POST", "/user/token/get",
                                    {"user_name": ADMIN_NAME, "pass_wd": ADMIN_PASS}))
    at = admin["data"]
    print("  ✓ 管理员登录成功")
    fafaenv.warn_single_login(ADMIN_NAME)

    # ------------------------------------------------------------------
    print("\n[1/8] 头像 / 账号（已存在的直接复用，脚本可重复执行）")
    avatar_urls = {}
    user_ids = {}
    for i, u in enumerate(USERS):
        uid, photo = find_user(u["name"])
        if uid:
            user_ids[u["name"]] = uid
            avatar_urls[u["name"]] = photo
            print(f"  = {u['nick']} 已存在 id={uid}，复用头像")
            continue
        png = make_png(240, *PALETTE[i % len(PALETTE)], seed=i)
        r = upload(at, f"avatar_{u['name']}.png", png, "image", f"{u['nick']}的头像")
        must(f"上传头像 {u['name']}", r)
        avatar_urls[u["name"]] = r["data"]["url"]

        r = api("POST", "/v1/user/create", {
            "name": u["name"], "nick_name": u["nick"], "email": u["email"],
            "password": USER_PASS, "repassword": USER_PASS, "gender": u["gender"],
            "wechat": u.get("wechat", ""), "weibo": u.get("weibo", ""),
            "github": u.get("github", ""), "qq": u.get("qq", ""),
            "short_describe": u["short"], "describe": u["describe"],
            "image_path": avatar_urls[u["name"]],
        }, at)
        must(f"创建用户 {u['nick']}", r)
        uid = r["data"]["id"] if isinstance(r.get("data"), dict) else None
        user_ids[u["name"]] = uid
        print(f"  + {u['nick']} ({u['name']}) id={uid}")
        must(f"设置 VIP {u['nick']}", api("POST", "/v1/user/admin/update", {"id": uid, "vip": 1}, at))

    # ------------------------------------------------------------------
    print("\n[2/8] 登录演示账号")
    tokens = {}
    for u in USERS:
        r = must(f"登录 {u['nick']}", api("POST", "/user/token/get",
                                          {"user_name": u["name"], "pass_wd": USER_PASS}))
        tokens[u["name"]] = r["data"]
    print(f"  ✓ {len(tokens)} 个演示账号登录成功（统一密码 {USER_PASS}）")

    # ------------------------------------------------------------------
    print("\n[3/8] 创建专栏")
    node_ids = {}
    for uname, seo, name, desc in NODES:
        nid = find_node(uname, seo)
        if nid:
            node_ids[(uname, seo)] = nid
            print(f"  = 专栏「{name}」已存在 id={nid}")
            continue
        r = must(f"创建专栏 {name}", api("POST", "/v1/node/create",
                                         {"seo": seo, "name": name, "describe": desc}, tokens[uname]))
        node_ids[(uname, seo)] = r["data"]["id"]
        print(f"  + 专栏「{name}」 → {uname}")

    # ------------------------------------------------------------------
    print("\n[4/8] 上传文章封面并创建文章")
    art_ids = {}
    for i, (uname, nseo, aseo, title, body) in enumerate(ARTICLES):
        cid0 = find_content(uname, aseo)
        if cid0:
            art_ids[aseo] = cid0
            print(f"  = 文章「{title[:20]}…」已存在 id={cid0}")
            continue
        # 封面用另一套配色，和头像区分开
        png = make_png(640, *PALETTE[(i + 3) % len(PALETTE)], seed=i + 5)
        up = must(f"上传封面 {title[:12]}", upload(tokens[uname], f"cover_{aseo}.png", png, "image", title))
        cover = up["data"]["url"]

        r = must(f"创建文章 {title[:12]}", api("POST", "/v1/content/create", {
            "seo": aseo, "node_id": node_ids[(uname, nseo)],
            "title": title, "status": 1, "describe": body,
        }, tokens[uname]))
        cid = r["data"]["id"]
        must(f"发布文章 {title[:12]}", api("POST", "/v1/content/publish", {"id": cid}, tokens[uname]))
        must(f"公开文章 {title[:12]}", api("POST", "/v1/content/update/status", {"id": cid, "status": 0}, tokens[uname]))
        must(f"设置封面 {title[:12]}", api("POST", "/v1/content/update/image", {"id": cid, "image_path": cover}, tokens[uname]))
        art_ids[aseo] = cid
        print(f"  ✓ 文章「{title[:20]}…」 id={cid} → {uname}")

    # ------------------------------------------------------------------
    print("\n[5/8] 发表评论")
    ok_cnt, skip_c = 0, 0
    seen_commenters = {}

    def commenters_of(cid):
        if cid not in seen_commenters:
            d = api("GET", "/content/comment", {"content_id": cid, "limit": 200}).get("data") or {}
            seen_commenters[cid] = {
                c.get("user_id") for c in ((d.get("extra") or {}).get("comments") or {}).values()
            }
        return seen_commenters[cid]

    for aseo, uname, body in COMMENTS:
        cid = art_ids.get(aseo)
        if not cid:
            continue
        if user_ids.get(uname) in commenters_of(cid):
            skip_c += 1
            continue
        r = api("POST", "/v1/comment/create", {"content_id": cid, "body": body}, tokens[uname])
        if not r.get("flag"):
            # 同一用户 60 秒内评论超过 5 条会触发验证码风控：清计数后重试一次
            redis_del_pattern("ff_seccomment:*")
            time.sleep(0.5)
            r = api("POST", "/v1/comment/create", {"content_id": cid, "body": body}, tokens[uname])
        if r.get("flag"):
            ok_cnt += 1
            seen_commenters[cid].add(user_ids.get(uname))
        else:
            print(f"  ! 评论失败({uname} → {aseo}): {json.dumps(r, ensure_ascii=False)[:160]}")
    if skip_c:
        print(f"  = 跳过 {skip_c} 条已存在的评论")
    print(f"  ✓ 本次新增 {ok_cnt} 条评论")

    # ------------------------------------------------------------------
    print("\n[6/8] 建立关注关系")
    ok_cnt = 0
    for a, b in FOLLOWS:
        r = api("POST", "/v1/relation/follow/add", {"user_name": b}, tokens[a])
        if r.get("flag"):
            ok_cnt += 1
    print(f"  ✓ 成功建立 {ok_cnt}/{len(FOLLOWS)} 条关注关系")

    # ------------------------------------------------------------------
    print("\n[7/8] 发送私信 + 点赞")
    uid_of = {}
    for u in USERS:
        info = api("GET", "/u/info", {"user_name": u["name"]}).get("data", {})
        uid_of[u["name"]] = info.get("id")

    ok_cnt, skip_cnt = 0, 0
    for a, b, msg in MESSAGES:
        if not uid_of.get(b):
            continue
        # 已发过同样内容就跳过，保证脚本可重复执行
        seen = json.dumps(api("GET", "/v1/message/list", {"message_type": -1, "limit": 200},
                              tokens[a]).get("data", {}), ensure_ascii=False)
        if msg[:20] in seen:
            skip_cnt += 1
            continue
        r = api("POST", "/v1/message/private/send", {"user_id": uid_of[b], "message": msg}, tokens[a])
        if r.get("flag"):
            ok_cnt += 1
    if skip_cnt:
        print(f"  = 跳过 {skip_cnt} 条已存在的私信")
    print(f"  ✓ 成功发送 {ok_cnt}/{len(MESSAGES)} 条私信")

    cool_cnt = 0
    for i, (uname, nseo, aseo, title, body) in enumerate(ARTICLES):
        cid = art_ids.get(aseo)
        # 每篇文章被 2~4 个不同用户点赞
        for j in range(2 + (i % 3)):
            liker = USERS[(i + j + 1) % len(USERS)]["name"]
            if liker == uname:
                continue
            if api("POST", "/v1/content/cool", {"id": cid}, tokens[liker]).get("flag"):
                cool_cnt += 1
    print(f"  ✓ 成功点赞 {cool_cnt} 次")

    # ------------------------------------------------------------------
    print("\n[8/8] 友情链接 + 全站公告")
    existing_links = json.dumps(api("GET", "/v1/friend/list", {"limit": 100}, at).get("data", {}),
                                ensure_ascii=False)
    for name, url in FRIEND_LINKS:
        if name in existing_links:
            print(f"  = 友链已存在 {name}")
            continue
        r = api("POST", "/v1/friend/create", {"name": name, "url": url, "sort_num": 0, "hide": 0, "open_new": 1}, at)
        print(f"  {'✓' if r.get('flag') else '!'} 友链 {name}")

    existing_global = json.dumps(api("GET", "/v1/message/admin/global/list", {"limit": 100}, at).get("data", {}),
                                 ensure_ascii=False)
    if GLOBAL_MESSAGE[:20] in existing_global:
        print("  = 全站公告已存在")
    else:
        r = api("POST", "/v1/message/admin/global/create",
                {"all_people": True, "right_now": True, "message": GLOBAL_MESSAGE}, at)
        print(f"  {'✓' if r.get('flag') else '!'} 全站公告")

    # ------------------------------------------------------------------
    print("\n===== 校验 =====")
    for path, body in (("/u/content", {"user_name": "linxiaoyu", "limit": 50}),
                       ("/u/nodes", {"user_name": "linxiaoyu", "limit": 50}),
                       ("/u", {"vip": -1, "limit": 50}),
                       ("/content", {"user_name": "nanke", "node_seo": "ontheroad",
                                     "seo": "qilianwind"})):
        r = api("GET", path, body)
        d = r.get("data", {}) or {}
        total = d.get("total", "?")
        print(f"  {path:14s} flag={r.get('flag')} total={total}")

    print(f"\n演示账号统一密码: {USER_PASS}   管理员: {ADMIN_NAME}/{ADMIN_PASS}")


if __name__ == "__main__":
    main()
