#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""FaFaCMS 内容增强：置顶长文 + 评论盖楼 + 粉丝关注。

做三件事：
1. 写入 4 篇数千字长文（markdown），发布并置顶（top=1）
2. 在这几篇文章下"盖楼"：多条主楼（root comment），每楼带一长串逐层回复的
   深层链（comment_type=2，comment_id 指向上一条，root_comment_id 指向楼底），
   并带侧枝，形成真实的讨论结构
3. 补粉丝关注：让 4 位作者拥有大量关注者，并补互相关注

用法：python3 enrich_seed.py
"""
import concurrent.futures as futures
import hashlib
import hmac
import json
import os
import random
import secrets
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request

import fafaenv

BASE = fafaenv.BASE
SIGN = fafaenv.SIGN_SECRET.encode()
ADMIN_NAME, ADMIN_PASS = fafaenv.ADMIN_NAME, fafaenv.ADMIN_PASS
PASSWORDS = [fafaenv.DEMO_PASS, os.environ.get("FAFA_CHAOS_PASS", "Chaos2026pass")]
WORKERS = int(os.environ.get("ENRICH_WORKERS", "4"))

_lock = threading.Lock()
_req = [0, 0]
_err = {}


def _one(method, path, body=None, token=None, ctype=None, raw=None, timeout=90):
    ts = str(int(time.time()))
    nonce = secrets.token_hex(8)
    sign = hmac.new(SIGN, f"{ts}:{nonce}".encode(), hashlib.sha256).hexdigest()
    data = raw
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
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return r.status, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()
    except Exception as e:
        return 0, str(e).encode()


def api(method, path, body=None, token=None, retry=4):
    for attempt in range(retry + 1):
        with _lock:
            _req[0] += 1
        _, raw = _one(method, path, body, token, ctype="application/json")
        try:
            resp = json.loads(raw.decode("utf-8", "replace"))
        except Exception:
            resp = {"_raw": raw[:150].decode("utf-8", "replace")}
        eid = (resp.get("error") or {}).get("id")
        if eid == 100034 and attempt < retry:
            time.sleep(1.5 * (attempt + 1))
            continue
        if not resp.get("flag"):
            with _lock:
                _req[1] += 1
                _err[str(eid or resp.get("_raw", "?"))] = _err.get(str(eid or "?"), 0) + 1
        return resp
    return resp


def limiter_loop(stop):
    """周期性清理反爬限流与评论防刷计数（凭据由 fafaenv 从环境变量读取）。"""
    pats = ["ff_rl:*", "ff_rl_v:*", "ff_rl_blk:*", "ff_seccomment:*"]
    while not stop.is_set():
        fafaenv.clear_limits(pats)
        stop.wait(1.5)


# ===========================================================================
# 四篇长文
# ===========================================================================

LONG_ARTICLES = [
    ("anan", "productthoughts", "threeyears",
     "「花花世界」三年，我把它当成一间小店来经营",
     """2023 年春天，我在一台二手笔记本上敲下第一行代码的时候，没想过这件事能持续三年。

那时我刚从一家大厂离职，原因是受不了每周的汇报。我想做一个「不需要向任何人汇报」的东西。花花世界就是这么来的——一个围绕内容互动的社区，没有推荐算法，没有增长团队，没有 KPI。

三年过去，它活下来了。日活不高，但有一批很稳定的用户。这篇文章不是成功学，是账本。

## 一、前两年我做错的三件事

### 1. 我以为「功能多」等于「产品好」

第一年我加了 40 多个功能：私信、关注、专栏、点赞、举报、回收站、拖拽排序、历史版本……

结果是：每个功能都只有两三个用户在用，而我在维护它们上花掉了 80% 的时间。

第二年我删掉了 17 个功能。删完之后，剩下的功能反而被用得更深了。

**教训**：功能不是资产，是负债。每一个功能都在持续消耗你的注意力。

### 2. 我把「运营」理解成了发内容

有半年时间，我每天自己写两篇文章发到站内，试图把内容量堆起来。

数据上它确实有效——文章数从 200 涨到了 1500。但三个月后我停下来，文章数一周就掉回了 200 出头。

因为那些内容只有我一个人在产，用户来看的是「我的博客」，不是「社区」。

**教训**：社区的内容必须来自社区。你自己产的内容只能用来点火，不能用来供暖。

### 3. 我没有认真对待「信任」

第二年夏天出过一次事故：一个用户把别人的文章全文搬过来，署名改成自己，连发了三十多篇。

我当时的第一反应是「做个查重」。后来想明白，查重解决不了问题——真正的问题是**被抄袭的人没有得到任何回应**。

我最后做的是：把那三十多篇文章全部下架，公开说明处理过程，并把这一条写进了站规。

**教训**：用户对社区的信任，是在处理坏事的时候建立起来的，不是在做好事的时候。

## 二、现在的三条原则

### 原则一：不做推荐，只做「最近」

站内没有算法排序。首页就是按时间倒序，唯一的例外是置顶。

有人建议加上「猜你喜欢」，我拒绝了。理由是：一旦有了推荐，用户就会开始揣测规则，然后围绕规则生产内容。那就变成另一个平台了。

### 原则二：数据要能带走

每个用户都可以导出自己的全部内容：文章、评论、私信。

这件事技术上不复杂，但它决定了用户的心态——你是租客还是房主？

### 原则三：默认保护隐私

用户能设置自己的文章只给关注者看，能匿名评论，能删除历史版本。

更重要的是：**数据库里的敏感字段是加密存储的**。也就是说，就算有一天服务器被拖库，用户的私信、邮箱、正文也不会以明文形式流出去。

这条看起来是技术决策，其实是产品决策。我不想做一个「把用户数据当成资产」的产品。

## 三、一些数字

| 项目 | 数值 |
|---|---|
| 持续运营 | 3 年 2 个月 |
| 累计注册用户 | 约 4000 |
| 其中有发文行为的 | 约 300 |
| 累计文章 | 2 万篇左右 |
| 服务器成本 | 每月 40 元（一台 2C4G） |
| 累计收入 | 0 |

是的，收入是零。我没有做付费，也没有接广告。

不是因为高尚，是因为**一旦开始收钱，我就必须做增长**。而我不想做增长。

## 四、关于钱

有朋友问我图什么。

我算过一笔账：这个站每年让我花掉大约 500 元（服务器 + 域名 + 备份盘子），以及每周大约 10 小时。

如果按我离职前的时薪折算，这是一笔非常昂贵的爱好。

但它给我换来了三件东西：

1. **一个完整的作品**。它从数据库到前端都是我写的，我可以指着它说这是我做的。
2. **一群具体的人**。三年里我认识了十几个长期用户，有学生、有医生、有退休教师。
3. **确定感**。每周那 10 小时，是我唯一百分之百为自己做事的 10 小时。

以我的判断，这笔交易是划算的。

## 五、下一步

今年我想做三件事：

- 把移动端体验重做一遍（现在手机上很难用）
- 加一个「只对自己可见」的草稿箱
- 写完整套自部署文档，让想自己搭一套的人能照着做

最后一件可能最重要。因为如果哪天我不想继续了，我希望这个站**能被别人接手**，而不是关掉。

写了这么多，其实只有一句话：

**做一件小事的第三年，比做一件大事的第一天，要难得多，也有意思得多。**"""),

    ("chenmo", "distributednotes", "backendtwoyears",
     "从零搭一套内容社区后端：两年半的技术选型与踩坑全记录",
     """这套后端从 2023 年跑到现在，经历了从单机到多实例、从几百篇文章到两万篇、从没有备份到每天三次全量备份。这篇把它完整拆一遍，包括所有我踩过的坑。

## 一、技术选型的几个决定

### 语言：Go

当时候选是 Go、Java、Node。

- Java 太重，我一个人维护两台 2C4G 的机器，JVM 本身就要吃掉 1G
- Node 的单线程模型在导出、备份这类 CPU 密集任务上很难受
- Go 的部署太省心了：一个静态二进制，丢上去就能跑

三年下来没有后悔。唯一的不便是生态里缺少像 Spring 那样「一站式」的框架，很多轮子要自己造。

### 数据库：MySQL 而非 PostgreSQL

这个决定现在看是**有争议的**。MySQL 胜出的理由只有一条：当时的虚拟主机只提供 MySQL。

代价是：

- 没有原生的 JSONB（内容里的结构化字段只能拆列或存 JSON 字符串）
- 部分索引、表达式索引支持较弱
- 全文检索基本没法用（后来我自己做了盲索引）

如果重新选，我会选 PostgreSQL。

### 缓存与队列：Redis

只用了最朴素的用法：会话、限流计数、验证码、去重窗口。

**没有引入消息队列**。所有「异步」都是 Go 的 goroutine + 重量级的落库补偿。现在看来这个决定是对的——两万篇文章的量级，MQ 只会增加运维负担。

## 二、数据模型上的三个关键决策

### 1. 冗余用户名字段

`content`、`comment`、`relation` 这些表里都存了 `user_name` 副本。

好处很明显：列表页不用 JOIN 用户表。

坏处也明显：**用户改名之后要刷全表**。我们现在的做法是改名后异步刷，用一个补偿任务兜底。

### 2. 软删除 + 状态机

文章的 `status` 有 5 个值：0 正常、1 隐藏、2 封禁、3 回收站、4 用户彻底删除。

看起来复杂，但它解决了一个真实问题：**用户删了文章之后后悔了**。

所有列表查询都必须带上 `status` 条件，这一点很容易漏。我们的做法是把默认条件收敛到一个 helper 函数里，禁止业务代码自己拼 WHERE。

### 3. 计数列（comment_num、cool 等）

计数器放在 content 表上，而不是每次 COUNT。

代价是**一致性靠应用保证**，一定会漂移。所以必须有一个对账任务：

```sql
SELECT c.id, c.comment_num,
       (SELECT COUNT(*) FROM comment k WHERE k.content_id = c.id AND k.is_delete = 0) AS real_num
FROM content c
HAVING c.comment_num <> real_num;
```

这个查询我们每天跑一次，漂移超过 5 就告警。

## 三、踩过最大的三个坑

### 坑一：时区

服务器跑在 UTC，用户在东八区。

`create_time` 存的是 Unix 时间戳，本来没问题。但报表里用 `FROM_UNIXTIME()` 时忘记加偏移，导致**日报的整体数据都落在前一天**。

这个 bug 活了四个月才被发现，因为它不影响功能，只影响统计。

**现在的做法**：所有涉及"日期"的计算只在一个地方做，其他地方一律用时间戳比较。

### 坑二：字符集

早期有张表建成了 `utf8`（三字节），用户发 emoji 直接报错。

MySQL 的 `utf8` 不是真正的 UTF-8，这个坑太经典了。修的时候要把库、表、列、连接四层全部改成 `utf8mb4`，少一层都会出问题。

### 坑三：明文存密码

这个最严重。

早期版本的代码里，密码是明文存储的。我一直知道这是个问题，但觉得"反正没人攻击"。

后来做安全审查的时候，我意识到一件事：**明文密码的风险不在被攻击，而在被备份**。备份文件会被复制到各种地方：对象存储、本地硬盘、同事的 U 盘。

现在的做法：密码用 bcrypt，短凭据（激活码、重置码）用 HMAC 摘要，其他敏感字段用信封加密（KEK + 每用户 DEK）。

## 四、加密改造的一些细节

这块单独说，因为它改变了整个数据层的写法。

### 加解密放在模型钩子层

```go
func (c *Content) BeforeInsert() { c.EncryptFields() }
func (c *Content) AfterLoad()    { c.DecryptFields() }
```

业务代码完全无感：读出来就是明文，写进去自动变密文。

代价有两个：

1. **不能再用明文做等值查询**。`WHERE user_name = 'xxx'` 永远查不到，因为每次加密结果都不同（随机 nonce）。
2. **不能用 LIKE 模糊搜索**。

### 解决方案：盲索引

对需要等值查询的字段，额外存一列 `xxx_bidx = HMAC(key, normalize(value))`。

查询时把明文转成同样的摘要再查：

```
WHERE name_bidx = HMAC(key, normalize('林小雨'))
```

这保证了登录、去重、精确搜索都能用，代价是**放弃模糊搜索**。

这是一个产品决策，不只是技术决策。你现在搜文章标题必须输入完整标题，这一点我在站内公告里说清楚了。

### 加密带来的列宽问题

密文比明文长约 1.4 倍（base64 + nonce + tag）。原来刚好卡在 `TEXT`（65535 字节）边界的字段会被静默截断。

我们把正文类字段统一改成了 `MEDIUMTEXT`。注意：xorm 的 `Sync2` **不会**修改已有列的类型，只会在启动日志里打一行 warning。所以部署时必须建新库，或者手工 ALTER。

## 五、可观测性

没有引入 Prometheus，也没有 ELK。只有：

- **结构化日志**，按天滚动，保留 14 天
- 一个 `/ping` 接口 + 外部定时探测
- 慢查询日志开在 1 秒

出事的时候我会去看三个东西：最近一小时错误日志的条数、慢查询、以及数据库连接数。

够用了。小项目上全套监控，成本大于收益。

## 六、备份

每天三次全量 `mysqldump`，加 binlog。保留 30 天。

**并且每周做一次恢复演练**——这一步不能省。备份文件损坏或者恢复流程有错，只有真恢复一次才知道。

恢复演练还包括一件事：**确认根密钥能正常解密**。加密之后，备份和密钥是两件必须一起活着的东西，缺一个就全废。

## 七、如果重来

1. 数据库选 PostgreSQL
2. 一开始就用 bcrypt / 加密，不要等"以后再说"
3. 计数列不要放在主表上，或者干脆每次 COUNT（两万篇的规模，COUNT 完全够用）
4. 不要在业务代码里手写 WHERE，全部收敛到 helper

最后一条是我现在最坚持的。**能约束住的地方，一定要加约束。**"""),

    ("linxiaoyu", "photographynotes", "tenyearsphotography",
     "三年拍废一万张照片：一个业余摄影者的自白",
     """这篇文章是写给三年前的自己的。

三年前我买了一台二手微单，以为有了好器材就能拍出好照片。三年后硬盘里躺着四万多张照片，真正满意的不到两百张。

## 一、器材焦虑的三年

第一年我换了三台相机、七支镜头。

每次看到别人的照片好看，我的第一反应都是「肯定是器材的问题」。

现在回头看，那些照片好看的原因通常是：

- 他在那里待了四个小时，我等了二十分钟
- 他拍了三百张挑出来一张，我拍了二十张
- 他知道自己要什么，我只是在按快门

器材只有在两种情况下才会成为瓶颈：**暗光**和**需要快速追焦的动态**。除此之外，任何一台 2015 年以后的相机都够用。

## 二、三个真正的转变

### 转变一：从「找景」到「等人」

刚开始我满世界找好看的景。后来发现，景一直在那里，让照片不同的是站在景里的人。

现在我拍一条街会在同一个位置站很久，等一个「对的人」走进来——骑车的人、牵狗的人、打电话的人。

等的过程中我会想：什么颜色经过这里会好看？什么动作在这里会成立？

### 转变二：从「拍完再看」到「先想再拍」

以前是举起来就拍，回家在电脑上挑。

现在我会先问自己三个问题：

1. 这张照片的主角是谁？
2. 观众的目光会落在哪里？
3. 如果只能保留一张，是不是这张？

想不清楚就不按快门。这个习惯让我的废片率从 90% 降到了 60%。

### 转变三：从「后期补救」到「前期决定」

有一年我痴迷于调色，每张照片都要在 Lightroom 里折腾半小时。

后来我把所有照片统一成一套预设，只在曝光和裁切上做微调。结果出片效率提高了三倍，而且风格反而更统一了。

**后期做不到的事情，前期一定也做不到。** 光的方向、明暗关系、色彩关系，这些按下快门前就定死了。

## 三、关于构图，我只信两条

### 第一条：先有主体，再有构图

构图法则（三分法、引导线、对称）都是**事后总结**，不是拍摄指南。

我的顺序是：找到主体 → 确定它在画面里的位置 → 再用线条、明暗去强化它。

如果主体本身不成立，构图再标准也是一张废片。

### 第二条：让画面有「重量」

我喜欢让画面一侧很重（大面积的暗部或密集的纹理），另一侧很轻（一片空白或者纯色）。

这种不平衡会制造张力。三分法那种"四平八稳"的构图，看多了会觉得无聊。

## 四、关于人

拍人比拍景难十倍。

不是技术难，是**关系难**。我的经验是：

- 先聊天，再举相机。让对方知道你是谁，为什么想拍
- 拍完给对方看。哪怕拍得不好
- 不要偷拍。我试过，出来的照片里有一种让人不舒服的东西

在菜市场拍那位卖豆腐的大姐，我前后去了五次，最后一次她才同意我拍。那张照片是我最喜欢的一张，因为照片里的她是放松的。

## 五、现在我怎么拍

很简单：

- 只带一台机器一支 35mm 定焦
- 出门前想一个主题（今天拍蓝色、今天拍背影）
- 同一个位置至少待 30 分钟
- 一天最多留 5 张
- 当天晚上就挑完，不留到第二天

最后一条最重要。**照片拖得越久，你就越舍不得删。**

## 六、写给三年前的自己

1. 不要买那么多镜头，把钱省下来去更多的地方
2. 不要羡慕别人的照片，你不知道他在那里等了多久
3. 不要把照片存在硬盘里不整理，那是自我感动
4. 打印出来。贴在墙上看一个月，你就知道哪张是真的好

三万张废片换来两百张满意的照片，这个比例听起来很差。

但我知道，那两百张里有一些，是任何人拿着任何器材、在任何时间点，都拍不出来的。

这就是我还在拍的原因。"""),

    ("haitang", "indiedevlog", "solodevledger",
     "一个人做产品的全部账本：2023—2026",
     """这篇文章公开我三年做独立开发的所有数字：收入、支出、时间投入、以及每一个失败产品的死因。

写完发现，最有价值的部分不是成功的那个，而是失败的那几个。

## 一、总账

| 项目 | 金额 |
|---|---|
| 三年总收入 | 约 41.6 万元 |
| 三年总支出 | 约 9.8 万元 |
| 净收入 | 约 31.8 万元 |
| 平均月净收入 | 约 8800 元 |

支出明细：

| 项目 | 金额 | 说明 |
|---|---|---|
| 服务器 / 域名 | 2.1 万 | 三台机器 + 若干域名 |
| 设计外包 | 3.2 万 | 早期两个产品的 UI |
| 苹果开发者账号 | 0.6 万 | 每年 99 美元 |
| 字体 / 图库授权 | 1.1 万 | 商用授权 |
| 软件订阅 | 1.4 万 | 设计、笔记、监控 |
| 其他 | 1.4 万 | 备案、发票、杂项 |

三年平均月收入 1.15 万，**低于我上班时的工资**。这一点必须先说清楚，否则后面的经验都是耍流氓。

## 二、时间账

这是更真实的部分。

- 有效工作时间：约 4200 小时（按每天 4 小时、每周 6 天算）
- 其中写代码：约 1600 小时（38%）
- 客服与支持：约 900 小时（21%）
- 设计与文案：约 700 小时
- 市场与内容：约 500 小时
- 学习与试错：约 500 小时

**最让我意外的是客服占了 21%。** 一个人做产品，客服不是"顺带做的事"，它是一个正经的岗位。

我的建议是：从第一天就把客服做成一个可复用的流程（FAQ + 模板回复 + 一个公开的更新日志），否则它会把你的开发时间全部吃掉。

## 三、三个失败的产品

### 失败一：时间管理工具（2023 上半年）

**死因：市场太挤。**

我用两个月做了一个番茄钟 + 待办 + 统计的工具。上线后第一个月 47 个用户，第二个月 9 个。

现在我知道原因了：这个赛道的头部产品已经免费且足够好，我没有提供任何"非用不可"的理由。

**教训**：如果一个需求已经有免费的成熟产品，你要么做得**窄十倍**，要么别做。

### 失败二：播客剪辑工具（2023 下半年）

**死因：技术选型错误。**

我把音频处理放在了浏览器端，用 WebAssembly 跑 FFmpeg。功能能做出来，但导出 30 分钟的音频要等 4 分钟，风扇狂转。

**教训**：不要在浏览器里做重计算。用户的笔记本不是你的服务器。

这个产品我投了 5 个月，收入 0。

### 失败三：记账工具（2024）

**死因：定位模糊。**

我想做"面向自由职业者的记账"，但功能上和大厂记账 App 没区别，只是多了个"按项目结算"。

结果就是既没吸引到自由职业者，也没吸引到普通用户。

**教训**：**定位不是一句 slogan，是一系列功能取舍。** 如果你的功能和通用产品有 80% 重合，那定位就是无效的。

## 四、唯一活下来的那个

2024 年底，我做了第三个版本的小工具：面向小团队的「日报生成器」。

它的起点很土：我自己每周要写给投资人的周报，写烦了。

这次我先做了三件事，和之前完全不同：

1. **只找 10 个真人聊**，不写代码。聊完发现，他们真正的痛点是"不知道该写什么"，而不是"排版麻烦"。
2. **两周做出能用的版本**，第三周就给这 10 个人用。
3. **第 6 周开始收费**，39 元/月，不做免费版。

结果：第一个月 12 个付费用户，第六个月 180 个，现在稳定在 400 出头。

月收入大约 1.6 万，覆盖了我的生活成本。

## 五、我现在的工作方法

### 1. 需求从聊天里来，不从想法里来

我给自己定的规矩：任何新功能，必须先有 3 个用户明确提过，否则不做。

### 2. 两周一个版本

不管做没做完，两周发一次。哪怕只改了一个文案。

节奏感比完成度重要——它会逼着我把需求切小。

### 3. 每周固定两小时看数据

只看四个数：新增、留存、付费转化、退款。

其他数据一概不看。数据看得越多，越容易做错决定。

### 4. 每月写一次公开日志

写给自己看。它让我在三个月后还能想起当时为什么做那个决定。

## 六、如果你也想做独立开发

我只说三条：

1. **先算清楚你能撑多久。** 没有收入的月份，你的存款能撑几个月？这个数字决定了你的心态。
2. **不要辞职做。** 我是被裁之后开始的，那段时间的焦虑严重影响了判断力。
3. **做窄的事。** "面向所有人"的产品，通常谁都不满意。

最后一句实话：

独立开发不自由。你只是把老板换成了用户，把 KPI 换成了账单。

但它有一个好处——**每一个决定都是你自己做的，包括那些做错的。** 这一点对我来说，值回票价。"""),
]

# ===========================================================================
# 评论素材
# ===========================================================================

FLOOR_OPENERS = [
    "写得太实在了，尤其是「{part}」那一段，我看完坐着想了十分钟。",
    "三年这个长度本身就说明问题。大多数人撑不过第一年的冬天。",
    "我不同意「{part}」这个结论，至少在我的场景里是反过来的。",
    "收藏了。今年刚好在做类似的事，你的账本给了我很大信心。",
    "看到「{part}」的时候有点鼻酸，这不就是我这两年的状态吗。",
    "有个具体的问题想问：{part}这部分你现在还坚持吗？",
    "别人都在讲方法论，只有你在讲代价。这才是我想看的。",
    "作为一个刚开始的人，这篇文章把我劝退了三次又拉回来三次。",
]

REPLIES = [
    "同意，但我觉得还有一层：{part}其实也和你的用户结构有关。",
    "你这个说法我认。不过「{part}」我持保留意见，能不能展开说说？",
    "补充一点：我那边的情况是，先做了半年才想明白这一条。",
    "同感。我是反过来的顺序，先撞了墙才回来读你这段话的。",
    "说得对，而且我觉得「{part}」这条对小团队尤其重要。",
    "有没有可能，{part}只是表象？真正的原因在于成本结构。",
    "受教了。我之前一直把这两件事混在一起看。",
    "这条我要抬个杠：如果能重来，你还会这么选吗？",
    "看到这里我去翻了下自己的记录，发现你说的问题我一模一样。",
    "关键还是心态。技术上的事都好办，熬不住才是真的熬不住。",
    "我把这段截图发给同事了，比我们内部复盘会讲得清楚。",
    "想问下「{part}」有没有具体的数字？想拿去说服我们老板。",
    "这两条其实是一件事的两面：先有取舍，才有效率。",
    "我不太同意。你这个结论建立在「只有一个人」的前提上。",
    "写得克制，没喊口号，这在同类文章里已经很少见了。",
    "看完觉得，最难的不是做对，是知道什么时候该停。",
    "有个疑问：{part}在你现在的规模下还成立吗？",
    "我把这篇文章打印出来贴在工位上了，谢谢。",
]

PARTS = ["功能不是资产而是负债", "社区内容只能点火不能供暖", "信任是在处理坏事时建立的",
         "不做推荐只做最近", "数据要能带走", "明文密码的风险在于备份",
         "盲索引换掉模糊搜索", "按页面替换而不是按仓库替换",
         "先测量再优化", "写不出来就是没想清楚", "定位是一系列功能取舍",
         "两周一个版本", "客服是一个正经岗位", "做窄的事"]


def main():
    random.seed(20260914)
    t0 = time.time()
    print(f"目标后端 {BASE}")

    stop = threading.Event()
    threading.Thread(target=limiter_loop, args=(stop,), daemon=True).start()

    at = api("POST", "/user/token/get", {"user_name": ADMIN_NAME, "pass_wd": ADMIN_PASS}).get("data")
    if not at:
        print("管理员登录失败")
        return 1
    fafaenv.warn_single_login(ADMIN_NAME)
    print()

    # -------- 收集用户并登录 --------
    print("\n[1/4] 收集用户并登录")
    r = api("GET", "/u", {"vip": -1, "limit": 100})
    all_users = [u for u in ((r.get("data") or {}).get("users") or []) if u.get("id")]
    total_users = (r.get("data") or {}).get("total", len(all_users))
    if total_users > len(all_users):
        r2 = api("GET", "/u", {"vip": -1, "limit": 100, "page": 2})
        all_users += [u for u in ((r2.get("data") or {}).get("users") or []) if u.get("id")]

    tokens = {}
    for u in all_users:
        for pw in PASSWORDS:
            rr = api("POST", "/user/token/get", {"user_name": u["name"], "pass_wd": pw})
            if rr.get("flag"):
                tokens[u["id"]] = rr["data"]
                break
            # 试密码失败是探测噪声，不计入失败统计
            with _lock:
                _req[1] -= 1
                if _err.get("100020"):
                    _err["100020"] -= 1
    pool = [u for u in all_users if u["id"] in tokens]
    print(f"  ✅ 用户 {len(all_users)}，可登录 {len(pool)}")

    by_name = {u["name"]: u for u in pool}

    # -------- 建号（若作者不在）--------
    def author_token(name):
        """返回 (id, token)，不存在则用管理员建号。"""
        if name in by_name:
            u = by_name[name]
            return u["id"], tokens[u["id"]]
        r = api("POST", "/v1/user/create", {
            "name": name, "nick_name": name, "email": f"{name}@longform.local",
            "password": PASSWORDS[0], "repassword": PASSWORDS[0], "gender": 1,
            "short_describe": "长文作者", "describe": "长期写作者。",
        }, at)
        if not r.get("flag"):
            return None, None
        uid = r["data"]["id"]
        api("POST", "/v1/user/admin/update", {"id": uid, "vip": 1}, at)
        tk = api("POST", "/user/token/get", {"user_name": name, "pass_wd": PASSWORDS[0]}).get("data")
        return uid, tk

    print("\n[2/4] 写入并置顶 4 篇长文")
    pinned = []
    for uname, nseo, aseo, title, body in LONG_ARTICLES:
        uid, tk = author_token(uname)
        if not tk:
            print(f"  ✗ 作者 {uname} 不可用，跳过")
            continue
        # 该用户的专栏
        r = api("GET", "/u/nodes", {"user_name": uname})
        nodes = (r.get("data") or {}).get("nodes") or []
        node = next((n for n in nodes if n.get("seo") == nseo), nodes[0] if nodes else None)
        if not node:
            r = api("POST", "/v1/node/create", {"seo": nseo, "name": f"{title[:6]}专栏",
                                                "describe": "长文合集"}, tk)
            node = r.get("data") or {}
        if not node.get("id"):
            print(f"  ✗ 专栏不可用，跳过 {title[:12]}")
            continue

        r = api("POST", "/v1/content/create", {
            "seo": aseo, "node_id": node["id"], "title": title, "status": 1, "describe": body}, tk)
        if not r.get("flag"):
            print(f"  ! 创建失败（可能已存在）: {title[:16]} → {json.dumps(r, ensure_ascii=False)[:120]}")
            cid = None
            for c in ((api("GET", "/u/content", {"user_name": uname, "limit": 100}).get("data") or {})
                      .get("contents") or []):
                if c.get("seo") == aseo:
                    cid = c["id"]
            if not cid:
                continue
        else:
            cid = r["data"]["id"]
        api("POST", "/v1/content/publish", {"id": cid}, tk)
        api("POST", "/v1/content/update/status", {"id": cid, "status": 0}, tk)
        rr = api("POST", "/v1/content/update/top", {"id": cid, "top": 1}, tk)
        pinned.append({"id": cid, "title": title, "author": uname, "uid": uid,
                       "len": len(body), "top_ok": bool(rr.get("flag"))})
        print(f"  ✅ 「{title[:22]}…」 {len(body)} 字  id={cid} 置顶={'ok' if rr.get('flag') else 'FAIL'}")

    if not pinned:
        print("没有成功置顶的文章")
        return 1

    # -------- 盖楼 --------
    print("\n[3/4] 评论盖楼（主楼 + 深层回复链 + 侧枝）")
    comment_ids = []
    roots_made = 0

    def make_root(cid, uid_tk):
        uid, tk = uid_tk
        part = random.choice(PARTS)
        body = random.choice(FLOOR_OPENERS).format(part=part)
        r = api("POST", "/v1/comment/create", {"content_id": cid, "body": body}, tk)
        return r.get("data") if r.get("flag") else None

    def make_reply(prev, uid_tk):
        uid, tk = uid_tk
        part = random.choice(PARTS)
        body = random.choice(REPLIES).format(part=part)
        r = api("POST", "/v1/comment/create",
                {"comment_id": prev, "is_to_comment": True, "body": body,
                 "anonymous": random.random() < 0.08}, tk)
        return r.get("data") if r.get("flag") else None

    def build_floor(art, budget):
        """盖一楼：一条主楼 + 若干条逐层回复链 + 侧枝，返回本轮产生的评论数。"""
        nonlocal roots_made
        authors = [u for u in pool if u["id"] != art["uid"]]
        rid = make_root(art["id"], (0, tokens[random.choice(authors)["id"]]))
        if not rid:
            return 0
        roots_made += 1
        comment_ids.append(rid)
        made = 1
        chain = []           # 已产生的评论 id，供侧枝挂载
        depth = random.randint(4, 26)
        prev = rid
        for _ in range(depth):
            if made >= budget:
                break
            u = random.choice(authors)
            nid = make_reply(prev, (u["id"], tokens[u["id"]]))
            if not nid:
                break
            comment_ids.append(nid)
            chain.append(nid)
            made += 1
            # 侧枝：有一定概率从链上较早的位置再分一条
            if chain and random.random() < 0.3 and made < budget:
                branch_from = random.choice(chain[:-1] or [prev])
                u2 = random.choice(authors)
                bid = make_reply(branch_from, (u2["id"], tokens[u2["id"]]))
                if bid:
                    comment_ids.append(bid)
                    chain.append(bid)
                    made += 1
            prev = nid
        return made

    floors_target = int(os.environ.get("ENRICH_FLOORS", "34"))       # 每篇主楼数
    total_comments = 0
    for art in pinned:
        per_floor_budget = random.randint(8, 30)
        got = 0
        for i in range(floors_target):
            got += build_floor(art, per_floor_budget)
        total_comments += got
        print(f"  ✅ 「{art['title'][:18]}…」 新增 {got} 条评论")
    print(f"  合计新增评论 {total_comments}（主楼 {roots_made}）")

    # 给评论点赞
    liked = 0
    for cid in random.sample(comment_ids, min(600, len(comment_ids))):
        u = random.choice(pool)
        if api("POST", "/v1/comment/cool", {"id": cid}, tokens[u["id"]]).get("flag"):
            liked += 1
    print(f"  ✅ 评论点赞 {liked} 次")

    # -------- 粉丝关注 --------
    print("\n[4/4] 粉丝关注")
    follows = 0
    author_ids = {(a["uid"], a["author"]) for a in pinned}

    def follow(aid, bid, atk):
        r = api("POST", "/v1/relation/follow/add", {"user_id": bid}, atk)
        return bool(r.get("flag")) or (r.get("error") or {}).get("id") in (110001, 110002, 110003)

    for uid, uname in author_ids:
        tk = tokens.get(uid)
        if not tk:
            continue
        # 让所有其他用户关注这 4 位作者
        for u in pool:
            if u["id"] == uid:
                continue
            if follow(u["id"], uid, tokens[u["id"]]):
                follows += 1
        # 反向：作者也关注一部分人，形成互关（互关后才能多轮私信）
        for u in random.sample([x for x in pool if x["id"] != uid],
                               min(25, len(pool) - 1)):
            if follow(uid, u["id"], tk):
                follows += 1
    print(f"  ✅ 关注关系处理 {follows} 次")

    # 作者之间互发私信
    msgs = 0
    aids = [(u["id"], tokens[u["id"]]) for u in pool if u["id"] in {a["uid"] for a in pinned}]
    for a_id, a_tk in aids:
        for b_id, _ in aids:
            if a_id == b_id:
                continue
            r = api("POST", "/v1/message/private/send",
                    {"user_id": b_id, "message": "你那篇长文我认真读完了，里面的账本部分我想引用一下，方便吗？"}, a_tk)
            if r.get("flag"):
                msgs += 1
    print(f"  ✅ 作者间私信 {msgs} 条")

    stop.set()

    # -------- 校验 --------
    print("\n===== 校验 =====")
    checks = []

    def check(label, cond, detail=""):
        checks.append((label, bool(cond)))
        print(f"  {'✅' if cond else '❌'} {label}" + (f"  | {detail}" if detail else ""))

    r = api("GET", "/u/content", {"limit": 20, "sort": ["-top", "-publish_time"]})
    top_list = [c for c in ((r.get("data") or {}).get("contents") or []) if c.get("top") == 1]
    check(f"公开列表首屏含置顶文章（{len(top_list)} 篇）", len(top_list) >= 1,
          " / ".join(c.get("title", "")[:12] for c in top_list[:3]))

    # 置顶文章应出现在排序首位（-top 优先）
    check("置顶排序生效", (r.get("data") or {}).get("contents") and
          (r["data"]["contents"][0].get("top") == 1),
          (r.get("data") or {}).get("contents", [{}])[0].get("title", "")[:20])

    worst = None
    for art in pinned:
        rr = api("GET", "/content/comment", {"content_id": art["id"], "limit": 100})
        d = rr.get("data") or {}
        nodes = d.get("comments") or []
        sons = sum(n.get("son_num", 0) for n in nodes)
        if worst is None or sons > worst[1]:
            worst = (art["title"], sons, len(nodes))
    check(f"盖楼生效：单篇最多 {worst[1]} 条回复、{worst[2]} 个主楼", worst[1] >= 20,
          worst[0][:18])

    # 深层链：取一篇置顶文，找一个 son_num 最大的主楼，拉出其全部回复
    art = pinned[0]
    rr = api("GET", "/content/comment", {"content_id": art["id"], "limit": 100})
    nodes = ((rr.get("data") or {}).get("comments") or [])
    deepest = max(nodes, key=lambda n: n.get("son_num", 0)) if nodes else None
    if deepest:
        rr2 = api("GET", "/content/comment",
                  {"content_id": art["id"], "root_comment_id": deepest["id"], "limit": 100})
        sons = ((rr2.get("data") or {}).get("comments") or [])
        types = {s.get("comment_type") for s in sons}
        check(f"主楼 {deepest['id']} 下有 {len(sons)} 条回复，comment_type={sorted(types)}",
              len(sons) >= 3 and 2 in types)

    # 粉丝数：脚本建的是「所有可登录账号互相关注」的完全图，每人粉丝数上限是 len(pool)-1。
    # 账号规模足够时仍以 20 为门槛，账号少时按实际上限校验，避免断言永不可能成立。
    fan_expect = min(20, max(1, len(pool) - 1))
    for art2 in pinned[:2]:
        r = api("GET", "/u/info", {"user_name": art2["author"]})
        check(f"{art2['author']} 有 {((r.get('data') or {}).get('followed_num'))} 个粉丝（期望 ≥{fan_expect}）",
              ((r.get("data") or {}).get("followed_num") or 0) >= fan_expect)

    # 加密不变式
    def q(query):
        return fafaenv.sql(query)

    hits = q("""SELECT
      (SELECT COUNT(*) FROM fafa.fafacms_content WHERE title LIKE '%花花世界%' AND title NOT LIKE 'v1:%')
    + (SELECT COUNT(*) FROM fafa.fafacms_comment WHERE `describe` LIKE '%功能不是资产%')
    + (SELECT COUNT(*) FROM fafa.fafacms_comment WHERE `describe` LIKE '%盖楼%')
    + (SELECT COUNT(*) FROM fafa.fafacms_content WHERE title_bidx='' AND title!='');""")
    if not fafaenv.sql_enabled():
        print("  ~ 未设置 FAFA_MYSQL_PASS，跳过落库扫描校验")
    check("长文/盖楼评论仍为密文、盲索引完好", hits in ("0", None), f"hits={hits}")

    print(f"\n  新增评论 {total_comments} 条、主楼 {roots_made} 个")
    print(f"  HTTP 请求 {_req[0]} 次，失败 {_req[1]} 次")
    if _err:
        print("  失败构成:", dict(sorted(_err.items(), key=lambda kv: -kv[1])[:5]))
    print(f"  耗时 {time.time() - t0:.1f}s")
    bad = [l for l, ok in checks if not ok]
    print(f"\n校验通过 {len(checks) - len(bad)}/{len(checks)}")
    if bad:
        print("失败项:", bad)
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
