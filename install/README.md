# FaFaCMS 后端部署说明（一键部署）

本目录是后端 `fafacms` 的一键部署入口，用 Docker 拉起整套后端服务：MySQL、Redis、phpMyAdmin、后端 API、接口文档（docsify）。

环境要求：类 Unix（Linux / macOS），已安装 Docker（推荐含 `docker compose` 插件，旧版 `docker-compose` 也能兼容）。

## 一键部署

```bash
cd install
./deploy.sh
```

脚本会自动：识别系统 → 选择数据持久化目录 → 准备目录与权限 → 创建共享网络 → 构建并启动镜像 → 打印访问地址。

执行成功后控制台会输出（含局域网 IP，可直接转发给同事）：

```
✅ FaFaCMS 后端服务已就绪
   后端 API    http://127.0.0.1:8080
   接口文档    http://127.0.0.1:9889
   phpMyAdmin  http://127.0.0.1:8000   (root/123456789)
   局域网访问（可转发给同事）:
   后端 API    http://192.168.x.x:8080
   ...
```

## 常用子命令

```bash
./deploy.sh up        # 启动（默认，可重复执行）
./deploy.sh down      # 停止并移除容器
./deploy.sh ps        # 查看容器状态
./deploy.sh logs      # 查看后端日志
./deploy.sh rebuild   # 重新构建镜像并重建容器
```

## 数据持久化目录（不在仓库内）

| 系统 | 默认目录 | 说明 |
| --- | --- | --- |
| macOS | `~/.fafacms` | Docker Desktop 默认只共享家目录 |
| Linux | `/opt/fafacms` | 系统数据目录 |

可用环境变量 `FAFACMS_DATA_DIR` 覆盖，例如：`FAFACMS_DATA_DIR=/data/fafacms ./deploy.sh`。

目录结构：

```
<数据目录>/
├── mysql/            # MySQL 数据 + 配置(my.cnf)
├── redis/            # Redis 数据 + 配置(redis.conf)
└── backend/
    ├── config.yaml   # 后端配置（由 install/config.yaml 复制而来）
    ├── storage/      # 上传文件
    ├── storage_x/    # 缩略图
    └── log/          # 后端日志
```

> 脚本会对数据目录执行 `chmod -R 777`，允许容器内 MySQL/Redis 的非 root 用户写入，避免挂载目录权限问题。

## 端口与账号

| 服务 | 地址 | 账号/密码 |
| --- | --- | --- |
| 后端 API | `:8080` | 超级管理员 `admin/admin` |
| 接口文档 | `:9889` | 无 |
| phpMyAdmin | `:8000` | `root/123456789` |
| MySQL | `:13307`（容器内 3306） | `root/123456789` |
| Redis | `:16379`（容器内 6379） | 密码 `123456789` |

> 后端通过 compose 服务网络直连 `fafamysql:3306` / `fafaredis:6379`，因此同一份配置在 Mac / Linux 通用，无需区分系统。宿主端口 `13307/16379` 仅用于外部访问（phpMyAdmin、裸机调试）。

## 修改配置（运维必读）

后端**运行期配置**在数据目录下：Linux `/opt/fafacms/backend/config.yaml`，macOS `~/.fafacms/backend/config.yaml`（可用 `FAFACMS_DATA_DIR` 覆盖）。**真实的邮箱、数据库、Redis 凭据都要运维直接编辑这份运行期配置**，而不是仓库里的 `install/config.yaml`。

> 仓库内 `install/config.yaml` 只是**占位模板**（`your_email@example.com` 等）：仅当数据目录里还不存在 config.yaml（首次部署）时才会被复制过去；之后无论执行 `./deploy.sh` 还是 `rebuild`，都**不会覆盖**运行期配置（日志会提示「已存在 … 保留本地配置」）。
>
> 因此正确做法：首次部署后 → 运维编辑运行期 config.yaml 填入真实值；**请勿把真实凭据提交进仓库**（仓库只放占位符）。

需要重点填写 / 检查的项：

| 配置项 | 说明 |
| --- | --- |
| `MailConfig.Host / Account / Password` | 发信邮箱。例：iCloud SMTP `smtp.mail.me.com`、`你的账号@icloud.com`、密码填 **App 专用密码** |
| `DbConfig.Pass` | MySQL 密码（compose 模板默认 `123456789`，生产请改，需与 `install/docker-compose.yaml` 的 `MYSQL_ROOT_PASSWORD` 一致） |
| `SessionConfig.RedisPass` | Redis 密码（生产请改，需与 `install/redis.conf` 的 `requirepass` 一致） |
| `OssConfig.*` | 仅当 `StorageOss: true`（使用对象存储）时填写 |
| 超级管理员 `admin/admin` | 首次启动自动创建，请及时在后台修改 |

## 数据加密与密钥（`FAFACMS_KEK_ROOT`，请读完并务必备份）

后端会把**敏感内容加密后**再写进数据库和磁盘，包括：

- 用户资料：用户名、昵称、邮箱、微信 / 微博 / GitHub / QQ、简介、登录 IP
- 私信正文、文章标题与正文、评论内容、专栏名与描述、举报理由
- 2FA 密钥（key）、激活码 / 重置码（只存不可逆摘要）
- **本地存储的上传文件**（对象存储模式不加密）

### 一、先认识几个缩写（含英文全称与中文说法）

| 缩写 | 英文全称 | 中文说法 | 一句话解释 |
| --- | --- | --- | --- |
| **KEK** | **K**ey **E**ncryption **K**ey | 密钥加密密钥 | "钥匙柜的钥匙"。它**不直接**加密你的文章，而是用来**加密别的密钥**。你环境变量里设置的就是它。 |
| **DEK** | **D**ata **E**ncryption **K**ey | 数据加密密钥 | 真正用来加密**用户数据**的钥匙。**每个用户一把**，互不相通。 |
| **FEK** | **F**ile **E**ncryption **K**ey | 文件加密密钥 | 真正用来加密**每个上传文件**的钥匙。**一个文件一把**。 |
| **HKDF** | **H**MAC-based **E**xtract-and-E**x**pand **K**ey **D**erivation **F**unction | 基于 HMAC 的密钥派生函数 | 一台"配钥匙的机器"：给它一份原料，它能**稳定地**配出多把用途不同的子钥匙。 |
| **AAD** | **A**dditional **A**uthenticated **D**ata | 附加认证数据 | 加密时一并"上封条"的标签（例如"这是哪张表、哪一列、属于哪个用户"）。有人把密文挪到别的位置，解密就会失败。 |
| **Blind Index** | —（无通用缩写） | 盲索引 | 明文的**不可逆摘要**，用来做等值查询，见下文第四节。 |

> 记不住英文不要紧，只要记住**KEK 是总钥匙、DEK/FEK 是具体开数据的钥匙、HKDF 是配钥匙的机器**就够了。

### 二、为什么要 KEK 和 DEK 两层？

把它想成一个**钥匙柜**：

- **DEK / FEK** 是真正开锁的钥匙（开某个用户的数据、开某个文件）
- **KEK** 是锁住"钥匙柜"的那把钥匙

程序每次要解密数据，不是拿着总钥匙去开一万把锁，而是：**先用 KEK 从钥匙柜里取出对应的 DEK，再用 DEK 打开数据**。

这样做的好处是：将来要**更换总密钥**时，只需要用新 KEK 把钥匙柜里的钥匙重新"包"一遍，**不需要把整库数据重新加密一遍**（库里可能有几百万行）。

这种"用总钥匙保管数据钥匙"的做法，术语叫**信封加密**（Envelope Encryption）——把钥匙装进信封，再用总钥匙把信封封口。

### 三、HKDF 是什么，为什么需要它

不应该拿同一把钥匙去开所有的锁。HKDF 就是"用一份原料配出多把钥匙"的机器：

- 给**同样的原料 + 同样的用途标签** → 永远配出**同一把**子钥匙（确定性，所以重启后还能解密）
- 给**不同的用途标签** → 配出的钥匙**完全不同**（用途隔离，一把泄漏不影响其他）

本项目的做法：你提供的 `FAFACMS_KEK_ROOT` 是**原料**，HKDF 把它派生成三把用途不同的子密钥：

```
FAFACMS_KEK_ROOT（你提供的根密钥，32 字节）
   │
   ├─ HKDF（标签 hunterhug:kek:wrap）  → 包裹子密钥：用来加密"每用户 DEK / 每文件 FEK"
   ├─ HKDF（标签 hunterhug:kek:index） → 索引子密钥：用来算"盲索引"、激活码/重置码摘要
   └─ HKDF（标签 hunterhug:kek:field） → 字段子密钥：用来加密具体的字段内容
```

三把子密钥之间**无法互相推导**。也就是说，即使某一处出了泄漏，另外两把仍然安全。

### 四、盲索引是什么（以及为什么搜索变"笨"了）

密文有一个特点：**同样的明文，每次加密出来的结果都不一样**（里面带了随机数）。
所以 `WHERE user_name = '张三'` 这种查询，永远匹配不到。

解决办法是：对需要**等值查询**的字段（用户名、昵称、邮箱、微信/微博/GitHub/QQ、文章标题、用户组名），
额外存一列**盲索引**——把明文用"索引子密钥"做一次 HMAC，得到一串固定的摘要。

查询时，把用户输入也做同样的摘要，再拿摘要去比对，就能**精确匹配**。

⚠️ 代价是：**模糊搜索不可用**。搜文章标题必须输入**完整标题**，输入片段搜不到。
这是加密换来的成本，**不是故障**；站内公告里也向用户说明了这一点。

### 五、运维清单

| 事项 | 说明 |
| --- | --- |
| 生成密钥 | `./deploy.sh up` 会**自动生成并落盘**到 `$DATA_DIR/secret/kek.key`（权限 600）；也可手动 `openssl rand -base64 32`，或 `docker run --rm 镜像 -gen_kek` 打印一把新的 |
| 注入方式 | `deploy.sh` 优先级：环境变量 `FAFACMS_KEK_ROOT` → `$DATA_DIR/secret/kek.key` → 都没有就自动生成一把并落盘。compose 只做变量透传（**密钥绝不写入仓库文件**） |
| **必须备份** | ⚠️ **根密钥丢失 = 数据库里的密文和磁盘上的加密文件永久无法恢复**（不是"打不开"，是数学上解不出来）。请把 `kek.key` 放进你的密钥备份流程，并且**与数据库备份分开存放**（放一起等于没备份） |
| 不要提交 | 该文件在数据目录内（`/opt/fafacms` 或 `~/.fafacms`），本来就在仓库之外；切勿复制进仓库 |
| 怎么确认用的是哪把 | 后端启动日志会打印密钥**指纹**：`Data encryption key ready (fingerprint=6b2a96212e81978a, builtin_default=false)`。指纹不能反推出密钥，但可以用来**比对线上/本地是不是同一把** |
| 轮换密钥 | 换根密钥需要用新 KEK 重新包裹每用户 DEK（按用户量级操作） |

### 六、不传密钥时的行为（重要）

- 未设置 `FAFACMS_KEK_ROOT` 时，后端**不会拒绝启动**，而是回退到代码里内置的默认密钥，
  并在启动日志打印 `SECURITY WARNING: ... using the BUILT-IN DEFAULT key`。
- ⚠️ 内置默认密钥**随源码公开**。也就是说：只要某个实例忘了配置密钥，
  它的数据库或备份一旦泄漏，任何人拿这个默认值就能离线解密全部内容——**等于没有加密**。
- 因此：本地开发、临时演示可以不管它；**任何真实部署都必须使用自己的密钥**。
  用 `./deploy.sh` 部署时脚本会自动生成，正常情况下不会命中默认值。
- 判断方法：看启动日志里 `builtin_default=false` 即表示用的是你自己的密钥。


修改后应用（config.yaml 不会被覆盖，直接重建后端容器即可）：

```bash
./deploy.sh            # up 即可生效
```

也可以 `docker compose up -d --no-deps fafacms` 或 `docker restart fafacms`。

> 注意：
> - `my.cnf`、`redis.conf` 每次部署仍会用仓库版本强制覆盖；`config.yaml` 不会。
> - 想重置运行期配置：手动删除该文件后再执行一次部署（会重新生成模板，**本地修改会丢失**，谨慎操作）。

## 灌演示数据 / 压测（可选）

`install/seed/` 下有三个脚本，用来给已经跑起来的实例灌数据、压测、造长文与评论楼：

| 脚本 | 作用 |
| --- | --- |
| `seed_demo.py` | 演示数据：8 个用户（含头像与资料）、9 个专栏、15 篇有正文的文章、评论、关注、私信、点赞、友链、公告。**幂等**，可重复执行 |
| `chaos_seed.py` | 并发随机压测：大批用户/专栏/文章/文件 + 随机互动，跑完自动校验列表、精确搜索、下载校验和与落库密文 |
| `enrich_seed.py` | 4 篇数千字长文并置顶、评论盖楼（深层回复链）、粉丝关注 |

```bash
cd install/seed
export FAFA_MYSQL_PASS='你的 MySQL 密码'   # 用于"落库是密文"类校验，可不设
export FAFA_REDIS_PASS='你的 Redis 密码'   # 用于清理风控计数，可不设
python3 seed_demo.py
```

> - 所有数据都**通过 HTTP 接口**创建，不直接写 SQL：加密改造后直接 INSERT 会存成明文，
>   且盲索引列为空（会导致登录和精确搜索失效）。
> - **仓库里不存任何真实凭据**，密码一律从环境变量读取；不提供时相关校验会跳过。
> - 详细说明（全部环境变量、规模参数、会被触发的风控）见 `install/seed/README.md`。

## 后端启动参数（flag）

后端通过 `install/docker-compose.yaml` 里 `fafacms` 服务的 `RUN_OPTS` 环境变量传入命令行参数，当前默认：

```
-config=/root/fafacms/config.yaml -history_record=true -init_db=true -single_login=true
```

修改 `RUN_OPTS` 后执行 `./deploy.sh rebuild` 生效。完整参数列表（对应 `main.go` 的 `init()`，`./fafacms -h` 也可查看）：

| 参数 | 默认 | 说明 |
| --- | --- | --- |
| `-config` | `./config.yaml` | 配置文件路径（Docker 里固定 `/root/fafacms/config.yaml`） |
| `-init_db` | `true` | 首次启动自动建库建表；之后可设 `false` |
| `-history_record` | `true` | 是否记录内容历史版本 |
| `-email_debug` | `false` | 邮件调试（生产环境建议 `false`） |
| `-auth_skip_debug` | `false` | 跳过鉴权调试（生产环境务必 `false`） |
| `-single_login` | `false` | 单点登录：同一用户后登录挤掉先前登录（当前已开启 `true`） |
| `-session_expire_time` | `604800`（7 天） | 登录 token 有效期（秒） |
| `-can_scale` | `true` | 上传图片是否自动缩放 |
| `-scale_width` | `500` | 缩放宽度（像素） |
| `-time_zone` | `8` | 时区偏移（东八区北京时间） |
| `-auto_ban` | `false` | 举报超过阈值自动违禁 |
| `-ban_time` | `10` | 举报达到该次数自动违禁（需 `-auto_ban=true`） |

安全相关环境变量（也可在 compose 的 `environment` 里设置）：

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `FAFA_SIGN_SECRET` | `fafacms-sign-v1-2026` | 请求签名 HMAC-SHA256 密钥（前后端需一致，前端 `src/utils/sign.js` 硬编码同款默认值） |
| `FAFA_SIGN_STRICT` | 未设置即强制开启 | 设为 `0` 关闭强制校验：放行无签名请求（脚本/调试用；带了签名头仍会校验） |
| `FAFA_CORS_ORIGINS` | 空（允许任意 Origin） | CORS 白名单，逗号分隔多个 Origin；配置后仅白名单内 Origin 可跨域 |

## 前端

前端仓库 `fafafront` 请到其仓库独立部署（见其 `README.md`）。两者通过共享 Docker 网络 `fafa-net` 互通，前端 nginx 反代到 `fafacms:8080`。

## 裸机部署（不用 Docker，可选）

需要本机安装 Go 1.17+、MySQL、Redis：

```bash
go build -o fafacms main.go
./fafacms -config=./config.yaml
```

> 注意：`install/config.yaml` 里数据库/Redis 用的是 Docker 服务名（`fafamysql` / `fafaredis`），仅 Docker 网络内可解析。裸机运行时请改为 `Host: 127.0.0.1`、`RedisHost: 127.0.0.1:6379`。命令参数见 `./fafacms -h`。
