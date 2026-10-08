# 种子 / 压测脚本

给一套**已经跑起来的** FaFaCMS 灌数据，用于演示、压测和回归。

所有数据都通过 HTTP 接口创建，**不直接写 SQL**。加密改造后这一点很重要：
直接 INSERT 明文会同时破坏两个不变量——落库是明文、`*_bidx` 盲索引为空
（后者会让登录和精确搜索全部失效）。走接口则由模型钩子自动加密并补盲索引。

## 前置

| 变量 | 默认值 | 说明 |
|---|---|---|
| `FAFA_BASE` | `http://127.0.0.1:8080` | 后端地址 |
| `FAFA_SIGN_SECRET` | `fafacms-sign-v1-2026` | 请求签名密钥，需与后端一致 |
| `FAFA_ADMIN` / `FAFA_ADMIN_PASS` | `admin` / `admin` | 首次建库自动创建的超级管理员 |
| `FAFA_DEMO_PASS` | `Fafa2026demo` | `seed_demo.py` 建的演示账号密码 |
| `FAFA_CHAOS_PASS` | `Chaos2026pass` | `chaos_seed.py` 建的压测账号密码 |
| `FAFA_MYSQL_CONTAINER` | `fafamysql` | MySQL 容器名 |
| `FAFA_MYSQL_USER` / `FAFA_MYSQL_PASS` / `FAFA_MYSQL_DB` | `root` / *（无默认）* / `fafa` | 用于"落库是密文"类校验 |
| `FAFA_REDIS_CONTAINER` / `FAFA_REDIS_PASS` | `fafaredis` / *（无默认）* | 用于清理风控计数 |
| `FAFA_APP_CONTAINER` / `FAFA_STORAGE_PATH` | `fafacms` / `/root/fafacms/storage` | 用于扫盘校验 |

> **仓库里不存任何真实凭据。** MySQL / Redis 密码必须由环境变量提供；
> 没提供时依赖它的校验会**跳过并打印提示**，脚本其余部分照常运行。

```bash
export FAFA_MYSQL_PASS='你的 MySQL 密码'
export FAFA_REDIS_PASS='你的 Redis 密码'
```

## 三个脚本

### `seed_demo.py` —— 演示数据（幂等，可重复执行）

8 个正常用户（含头像、简介、微信/GitHub/QQ）、9 个专栏、15 篇有正文的文章、
评论、关注、私信、点赞、友情链接、全站公告。

重跑只会跳过已存在的数据，不会重复创建。

```bash
python3 seed_demo.py
```

### `chaos_seed.py` —— 混沌压测（并发 + 随机，非幂等）

并发随机生成大量用户 / 专栏 / 文章 / 文件，再做随机的评论、关注、私信、点赞，
最后校验列表、精确搜索、文件下载 sha256、落库密文与盲索引。

```bash
CHAOS_USERS=40 CHAOS_NODES=2 CHAOS_ARTICLES=6 CHAOS_COMMENTS=800 \
CHAOS_FOLLOWS=400 CHAOS_MESSAGES=300 CHAOS_LIKES=600 CHAOS_FILES=120 \
CHAOS_WORKERS=6 python3 chaos_seed.py
```

| 变量 | 默认 |
|---|---|
| `CHAOS_USERS` / `CHAOS_NODES` / `CHAOS_ARTICLES` | 40 / 2 / 6 |
| `CHAOS_COMMENTS` / `CHAOS_FOLLOWS` / `CHAOS_MESSAGES` / `CHAOS_LIKES` / `CHAOS_FILES` | 800 / 400 / 300 / 600 / 120 |
| `CHAOS_WORKERS` | 6 |

### `enrich_seed.py` —— 置顶长文 + 评论盖楼 + 粉丝

写入 4 篇数千字长文并置顶（`top=1`），在它们下面"盖楼"：多个主楼，
每楼一串逐层回复的深链（`comment_type=2`，`comment_id` 指向上一条，
`root_comment_id` 指向楼底）加侧枝；再补齐粉丝关注与作者间私信。

```bash
ENRICH_FLOORS=34 python3 enrich_seed.py     # 每篇主楼数，默认 34
```

## 关于限流

线上有三层风控，压测时会被触发，属预期而非故障：

- **反爬限流**：每 IP 每分钟 240 次请求，超限会临时拉黑
- **评论防刷**：同一用户 60 秒内超过 5 条评论要求验证码
- **注册防刷**：同一 IP 每小时超过 3 次注册要求验证码（脚本用管理员建号接口，不受影响）

脚本里有一个监控线程周期性清理这些计数（`ff_rl:*` / `ff_seccomment:*`），
目的是让**业务链路**而不是限流器成为瓶颈。未配置 `FAFA_REDIS_PASS` 时会退化为
"命中限流就退避重试"，会明显变慢但仍能跑完。

## ⚠️ 单点登录与管理员账号（容易踩）

后端 `RUN_OPTS` 默认带 `-single_login=true`：**同一账号同一时间只能有一处在线**，
再次登录会把之前的会话顶掉。

这三个脚本都会以管理员账号登录（默认 `admin`），所以在**你正用 admin 浏览站点时跑脚本，
你浏览器里的登录会被顶下去**，表现为被弹回登录页、提示「请先登录」。

两种应对方式：

1. **给灌数据单独准备一个账号**（推荐）：把该账号名写进 `FAFA_ADMIN` 再跑脚本，互不影响。
2. **关掉单点登录**：编辑 `install/docker-compose.yaml` 的 `RUN_OPTS`，去掉 `-single_login=true`，
   然后 `./deploy.sh rebuild`。

> 另注：只有**名字恰好是 `admin`** 的账号才拥有超级管理员权限（见 `core/controllers/auth.go`），
> 换用其他账号时需要先给它的用户组分配相应管理资源。

## 注意

- 这些脚本会**真实写数据**，请只对测试 / 演示实例运行
- `chaos_seed.py` 与 `enrich_seed.py` 的评论是**追加**的，重跑会继续增加
- 想清库重来：删库重建后启动后端（`-init_db=true`）会自动建表并创建 `admin/admin`
