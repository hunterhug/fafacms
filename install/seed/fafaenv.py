#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""种子脚本共用的本地环境访问层。

**仓库里不写任何真实凭据**：MySQL / Redis 密码、容器名、存储路径全部从环境变量读取。
未提供必需变量时，依赖它的校验会**安全降级**（跳过并打印提示），而不是报错中断。

环境变量一览：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `FAFA_BASE` | `http://127.0.0.1:8080` | 后端地址；也可指向前端反代 `http://127.0.0.1:3000`，此时接口前缀要用 `/api/`、`/app/` |
| `FAFA_SIGN_SECRET` | `fafacms-sign-v1-2026` | 请求签名密钥，与后端一致（这是文档里公开的默认值，不是真实密钥） |
| `FAFA_ADMIN` / `FAFA_ADMIN_PASS` | `admin` / `admin` | 首次建库自动创建的超级管理员 |
| `FAFA_DEMO_PASS` | `Fafa2026demo` | 演示账号统一密码 |
| `FAFA_MYSQL_CONTAINER` | `fafamysql` | MySQL 容器名 |
| `FAFA_MYSQL_USER` | `root` | MySQL 用户 |
| `FAFA_MYSQL_PASS` | *（必填才启用 SQL 校验）* | MySQL 密码 |
| `FAFA_MYSQL_DB` | `fafa` | 库名 |
| `FAFA_REDIS_CONTAINER` | `fafaredis` | Redis 容器名 |
| `FAFA_REDIS_PASS` | *（必填才启用风控清理）* | Redis 密码 |
| `FAFA_APP_CONTAINER` | `fafacms` | 后端容器名（用于扫盘校验） |
| `FAFA_STORAGE_PATH` | `/root/fafacms/storage` | 容器内存储目录 |
"""
import os
import subprocess

BASE = os.environ.get("FAFA_BASE", "http://127.0.0.1:8080")
SIGN_SECRET = os.environ.get("FAFA_SIGN_SECRET", "fafacms-sign-v1-2026")
ADMIN_NAME = os.environ.get("FAFA_ADMIN", "admin")
ADMIN_PASS = os.environ.get("FAFA_ADMIN_PASS", "admin")
DEMO_PASS = os.environ.get("FAFA_DEMO_PASS", "Fafa2026demo")

MYSQL_CONTAINER = os.environ.get("FAFA_MYSQL_CONTAINER", "fafamysql")
MYSQL_USER = os.environ.get("FAFA_MYSQL_USER", "root")
MYSQL_PASS = os.environ.get("FAFA_MYSQL_PASS", "")
MYSQL_DB = os.environ.get("FAFA_MYSQL_DB", "fafa")

REDIS_CONTAINER = os.environ.get("FAFA_REDIS_CONTAINER", "fafaredis")
REDIS_PASS = os.environ.get("FAFA_REDIS_PASS", "")

APP_CONTAINER = os.environ.get("FAFA_APP_CONTAINER", "fafacms")
STORAGE_PATH = os.environ.get("FAFA_STORAGE_PATH", "/root/fafacms/storage")


SINGLE_LOGIN_WARNING = """⚠️  本脚本会以管理员账号 {name} 登录。若后端 RUN_OPTS 里带了 -single_login=true\
（同一账号同一时间只能一处在线），你在浏览器里的该账号会被顶掉，需要重新登录。\
灌数据时建议指定一个专用账号：FAFA_ADMIN=你的专用账号 python3 本脚本"""


def warn_single_login(name, quiet=False):
    """打印单点登录提醒（脚本反复登录同一账号时会把浏览器会话顶掉）。"""
    if not quiet:
        print(SINGLE_LOGIN_WARNING.format(name=name))


def sql_enabled():
    return bool(MYSQL_PASS)


def redis_enabled():
    return bool(REDIS_PASS)


def sql(query, timeout=120):
    """执行一条 SQL，返回 stdout（-N -B 制表符分隔）。未配置密码时返回 None。"""
    if not MYSQL_PASS:
        return None
    argv = ["docker", "exec", MYSQL_CONTAINER, "mysql",
            "-u" + MYSQL_USER, "-p" + MYSQL_PASS, "-N", "-B", "-e", query]
    try:
        return subprocess.run(argv, capture_output=True, text=True, timeout=timeout).stdout.strip()
    except Exception as e:
        print(f"  ~ SQL 执行失败（{e}），跳过该项校验")
        return None


def redis_cli(*args, timeout=20):
    """执行 redis-cli 子命令，返回 stdout。未配置密码时返回 ""。"""
    if not REDIS_PASS:
        return ""
    argv = ["docker", "exec", REDIS_CONTAINER, "redis-cli",
            "-a", REDIS_PASS, "--no-auth-warning", *args]
    try:
        return subprocess.run(argv, capture_output=True, text=True, timeout=timeout).stdout
    except Exception:
        return ""


def clear_limits(patterns):
    """清理风控计数（反爬限流窗口 / 评论防刷），让压测瓶颈落在业务链路而不是限流器上。

    未配置 `FAFA_REDIS_PASS` 时静默跳过——脚本仍可运行，只是可能被限流重试拖慢。
    """
    if not REDIS_PASS:
        return False
    keys = []
    for p in patterns:
        keys += redis_cli("--scan", "--pattern", p).split()
    if keys:
        redis_cli("DEL", *keys)
    return True


def storage_plaintext_scan(keywords):
    """在容器内对存储目录做明文扫描，返回命中文件数；失败返回 None。"""
    pattern = "\\|".join(keywords)
    cmd = f"grep -rl '{pattern}' {STORAGE_PATH} 2>/dev/null | wc -l"
    try:
        out = subprocess.run(["docker", "exec", APP_CONTAINER, "sh", "-c", cmd],
                             capture_output=True, text=True, timeout=180).stdout.strip()
        return int(out or "0")
    except Exception:
        return None
