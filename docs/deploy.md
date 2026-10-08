# 部署指南

本文介绍如何把 FaFa CMS 后端（`fafacms`）与前端（`fafafront`）部署到一台云服务器上，并对外提供 HTTPS 访问。前后端为两个独立仓库，各自一键部署，通过共享 Docker 网络互通。

## 一、前置准备

### 服务器

- 建议 **2 核 4G**、系统盘 40G 以上、类 Unix 系统（推荐 **Ubuntu 22.04**）。
- 需已安装 **Docker** 与 **docker compose**。
- 若面向中国大陆访问且不想备案，机房选**中国香港**（香港服务器不要求 ICP 备案）。

### 域名

- 准备一个已实名认证的域名（建议 `.com`）。
- 在域名解析处加两条 **A 记录**：`@` 与 `www`，均指向服务器公网 IP。

## 二、安装 Docker

```bash
sudo apt update && sudo apt install -y curl git
curl -fsSL https://get.docker.com | sudo sh
sudo systemctl enable --now docker
docker --version && docker compose version
```

## 三、部署后端（fafacms）

```bash
git clone https://github.com/hunterhug/fafacms.git
cd fafacms/install
./deploy.sh
```

脚本会一键拉起以下服务（首次需拉取镜像并编译，可能较慢）：

| 服务 | 地址 |
| --- | --- |
| 后端 API | http://127.0.0.1:8080 |
| 接口文档 | http://127.0.0.1:9889 |
| phpMyAdmin | http://127.0.0.1:8000 |

数据持久化目录：Linux 默认 `/opt/fafacms`（可用环境变量 `FAFACMS_DATA_DIR` 覆盖）。

验证：

```bash
./deploy.sh ps     # 容器应均为 Up
./deploy.sh logs   # 查看后端日志
```

> 后端更详细的部署与配置说明见仓库内 `install/README.md`。

## 四、生产配置（务必修改）

编辑运行期配置：

```bash
vi /opt/fafacms/backend/config.yaml
```

| 配置项 | 说明 |
| --- | --- |
| `MailConfig.*` | 填真实发信邮箱与 App 专用密码（激活/改密邮件用） |
| `DbConfig.Pass` | 改掉默认值，并同步 `install/docker-compose.yaml` 的 `MYSQL_ROOT_PASSWORD` |
| `SessionConfig.RedisPass` | 改掉默认值，并同步 `install/redis.conf` 的 `requirepass` |

改完生效：

```bash
./deploy.sh rebuild
```

> 默认超级管理员 `admin/admin` 为示例账号，首次登录后请立即修改密码。

## 五、数据加密与密钥（必读）

后端会把**敏感内容加密后**再落库落盘：用户资料（用户名、邮箱、微信/微博/GitHub/QQ、简介、登录 IP）、
私信正文、文章标题与正文、评论内容、专栏名与描述、举报理由、激活码/重置码摘要，以及**本地上传的文件**。

### 5.1 先认识几个缩写（英文全称 + 一句话解释）

| 缩写 | 英文全称 | 中文说法 | 一句话解释 |
| --- | --- | --- | --- |
| **KEK** | **K**ey **E**ncryption **K**ey | 密钥加密密钥 | "钥匙柜的钥匙"。它**不直接**加密文章，而是用来**加密别的密钥**。 |
| **DEK** | **D**ata **E**ncryption **K**ey | 数据加密密钥 | 真正加密**用户数据**的钥匙，**每个用户一把**。 |
| **FEK** | **F**ile **E**ncryption **K**ey | 文件加密密钥 | 真正加密**每个上传文件**的钥匙，**一个文件一把**。 |
| **HKDF** | **H**MAC-based **E**xtract-and-E**x**pand **K**ey **D**erivation **F**unction | 基于 HMAC 的密钥派生函数 | 一台"配钥匙的机器"：给一份原料，稳定地配出多把用途不同的子钥匙。 |
| **AAD** | **A**dditional **A**uthenticated **D**ata | 附加认证数据 | 加密时一并"上封条"的标签，密文被挪到别处就解不开。 |

> 记不住英文没关系：**KEK 是总钥匙，DEK/FEK 是开具体数据的钥匙，HKDF 是配钥匙的机器。**

### 5.2 为什么要两层密钥

把它想成**钥匙柜**：DEK/FEK 是真正开锁的钥匙，KEK 是锁住钥匙柜的钥匙。

每次解密，程序先用 KEK 从钥匙柜取出对应的 DEK，再用 DEK 解数据。
好处是将来**更换总密钥**时，只需用新 KEK 把钥匙重新"包"一遍，**不必把整库数据重新加密一遍**。

这种"用总钥匙保管数据钥匙"的做法叫作**信封加密**（Envelope Encryption）。

### 5.3 HKDF 做什么

不应该用同一把钥匙开所有的锁。HKDF 负责从你提供的**根密钥**派生出三把用途隔离的子密钥：

```
FAFACMS_KEK_ROOT（根密钥，32 字节）
   ├─ 派生出「包裹」子密钥 → 加密每用户 DEK / 每文件 FEK
   ├─ 派生出「索引」子密钥 → 计算盲索引、激活码/重置码摘要
   └─ 派生出「字段」子密钥 → 加密具体字段内容
```

同样的原料 + 同样的用途标签，**永远得到同一把**子密钥（所以重启后还能解密）；
用途标签不同则钥匙完全不同，**一把泄漏不影响另外两把**。

### 5.4 密钥从哪来、怎么备份

`./deploy.sh` 会**自动生成**一把随机密钥并落盘到 `$DATA_DIR/secret/kek.key`（权限 600），
你**不需要手动做任何事**；也可以在部署前自行设置环境变量 `FAFACMS_KEK_ROOT` 来指定。

```bash
# 查看当前密钥指纹是否与线上一致（启动日志里也有）
docker logs fafacms 2>&1 | grep "encryption key ready"
# Data encryption key ready (fingerprint=6b2a96212e81978a, builtin_default=false)
```

| 事项 | 说明 |
| --- | --- |
| 密钥文件 | `/opt/fafacms/secret/kek.key`（Linux）/ `~/.fafacms/secret/kek.key`（macOS） |
| **必须备份** | ⚠️ **密钥丢失 = 数据库里的密文和磁盘上的加密文件永久无法恢复**。请**与数据库备份分开存放**（放一起等于没备份） |
| 不要提交 | 该文件在数据目录内，本来就在仓库之外；切勿复制进仓库 |
| 怎么确认 | 启动日志 `builtin_default=false` 表示用的是你自己生成的密钥 |

### 5.5 忘了配密钥会怎样

不设置 `FAFACMS_KEK_ROOT` 时后端**不会拒绝启动**，而是回退到**代码里内置的默认密钥**，
并打印 `SECURITY WARNING: ... using the BUILT-IN DEFAULT key`。

⚠️ 内置默认密钥**随源码公开**：任何忘了配置密钥的实例，其数据库或备份一旦泄漏，
任何人都能用这个默认值离线解密全部内容——**等于没有加密**。

用 `./deploy.sh` 部署时脚本会自动生成真密钥，正常不会命中该情况；判断依据就是日志里的
`builtin_default=false`。

### 5.6 搜索为什么只能精确匹配

密文每次加密结果都不同，所以 `WHERE user_name = '张三'` 匹配不到。系统对需要等值查询的字段
（用户名、昵称、邮箱、微信/微博/GitHub/QQ、文章标题、用户组名）额外存了一列**盲索引**——
明文的**不可逆摘要**，查询时把输入做同样的摘要再比对，实现**精确匹配**。

代价是**模糊搜索不可用**：搜文章标题必须输入完整标题。这是加密换来的成本，**不是故障**。

> 更详细的原理、密钥轮换步骤与全部环境变量见仓库内 `install/README.md`。

## 六、部署前端（fafafront）

```bash
git clone https://github.com/hunterhug/fafafront.git
cd fafafront
./deploy.sh
```

脚本会构建前端（Vue 3 + Vite → nginx 静态托管）并启动：

| 服务 | 地址 |
| --- | --- |
| 前端网站 | http://127.0.0.1:3000 |
| 产品文档 | http://127.0.0.1:8889 |

前端 nginx 已内置反向代理，会把以下路径转发到后端（默认走共享网络 `fafacms:8080`）：

| 路径 | 转发目标 |
| --- | --- |
| `/app/*` | 后端公开接口 |
| `/api/*` | 后端登录后接口 `/v1/*` |
| `/storage/*`、`/storage_x/*` | 后端上传的图片与缩略图 |

- 同机部署：无需任何配置，前后端通过共享网络 `fafa-net` 互通。
- 跨机部署（前端与后端不在同一台机器）：`BACKEND_UPSTREAM=后端IP:8080 ./deploy.sh rebuild`。

## 七、反向代理 + 自动 HTTPS（Caddy）

前端 nginx 已转发所有接口，因此公网入口只需放一个 Caddy，把流量全部交给前端即可，同时自动申请并续期 HTTPS 证书。

在 `/opt/caddy/` 目录下建两个文件：

**`Caddyfile`**

```caddyfile
blog.example.com {
    reverse_proxy fafafront-web:80
}
```

**`docker-compose.yml`**

```yaml
services:
  caddy:
    image: caddy:2-alpine
    container_name: caddy
    restart: always
    ports:
      - "80:80"
      - "443:443"
      - "443:443/udp"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - ./data:/data        # 证书持久化，勿删
      - ./config:/config
    networks:
      - fafa-net

networks:
  fafa-net:
    external: true
```

启动：

```bash
cd /opt/caddy
docker compose up -d
```

Caddy 会向 Let's Encrypt 自动申请证书并到期自动续期。前提：域名已解析到服务器、80/443 端口已开放。

## 八、安全组 / 防火墙

公网只开放以下端口，其余一律关闭：

| 端口 | 用途 | 是否开放 |
| --- | --- | --- |
| 22 | SSH | ✅ |
| 80 | HTTP | ✅ |
| 443 | HTTPS | ✅ |
| 8080 / 9889 / 8000 | 后端 / 接口文档 / phpMyAdmin | ❌ |
| 13307 / 16379 | MySQL / Redis | ❌ |

## 九、验收与备份

验收：

- `https://你的域名` 可打开、可登录、可发布内容。
- 接口经前端反代可用。
- 后台上传一张图片、发一条私信能正常显示。
- 启动日志里出现 `Data encryption key ready (fingerprint=..., builtin_default=false)` —— 说明用的是你自己的密钥。

备份（数据都在 `/opt/fafacms`）：

```bash
/opt/fafacms/mysql/data        # 数据库（密文）
/opt/fafacms/backend/storage   # 上传文件与图片（密文）
/opt/fafacms/secret/kek.key    # 加密根密钥（KEK）—— 缺了它上面两份都是乱码
```

⚠️ **三份都要备份，而且密钥必须与数据分开存放**：

- 数据库和文件在磁盘上是**密文**，只有配合这把密钥才能还原成明文；
- 密钥丢了，**数据永远无法恢复**（不是"打不开"，是数学上解不出来）；
- 密钥和数据放在同一个备份盘里，等于没备份——一次拖库或一次误删就同时失去两者。

建议：数据目录定时备份到对象存储，密钥单独保存在密码管理器或离线介质中。

> 密钥的完整说明（KEK / DEK / FEK / HKDF 分别是什么、轮换步骤、环境变量清单）见仓库内 `install/README.md`。

