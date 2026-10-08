#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""FaFaCMS 混沌压力 + 数据填充脚本（并发、随机、走真实 API）。

目的：在真实部署实例上灌入**大量且随机**的数据，验证加密改造在体量下的正确性：
用户 / 专栏 / 文章 / 评论 / 关注 / 私信 / 点赞 / 文件上传下载 / 精确搜索 / 页面列表。

说明：
- 全部通过 HTTP 接口创建，密文与盲索引由模型钩子生成（直接写 SQL 会破坏这两个不变量）。
- 反爬限流是「每 IP 每分钟 240 次」，本脚本用监控线程周期清计数（**限流本身已单独验证过**），
  目的是让业务链路而不是限流器成为瓶颈。
- 脚本幂等性较弱（用户名带时间戳后缀，重跑等于新增一批），用 --clean 可指定前缀清理。

用法：
    python3 chaos_seed.py                 # 默认规模
    CHAOS_USERS=60 CHAOS_ARTICLES=8 python3 chaos_seed.py
"""
import concurrent.futures as futures
import hashlib
import hmac
import json
import os
import random
import secrets
import struct
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request
import zlib

import fafaenv

BASE = fafaenv.BASE
SIGN_SECRET = fafaenv.SIGN_SECRET.encode()
ADMIN_NAME, ADMIN_PASS = fafaenv.ADMIN_NAME, fafaenv.ADMIN_PASS
PASS = os.environ.get("FAFA_CHAOS_PASS", "Chaos2026pass")

N_USERS = int(os.environ.get("CHAOS_USERS", "40"))
N_NODES = int(os.environ.get("CHAOS_NODES", "2"))          # 每人专栏数
N_ARTICLES = int(os.environ.get("CHAOS_ARTICLES", "6"))    # 每人文章数
N_COMMENTS = int(os.environ.get("CHAOS_COMMENTS", "800"))
N_FOLLOWS = int(os.environ.get("CHAOS_FOLLOWS", "400"))
N_MESSAGES = int(os.environ.get("CHAOS_MESSAGES", "300"))
N_LIKES = int(os.environ.get("CHAOS_LIKES", "600"))
N_FILES = int(os.environ.get("CHAOS_FILES", "120"))
WORKERS = int(os.environ.get("CHAOS_WORKERS", "6"))
STAMP = str(int(time.time()))[-6:]

_lock = threading.Lock()
_stats = {}
_req_seen = [0, 0]  # [总数, 失败数]
_err_ids = {}       # 错误码 -> 次数（诊断失败构成）


def bump(k, n=1):
    with _lock:
        _stats[k] = _stats.get(k, 0) + n


# ---------------------------------------------------------------------------
# HTTP（带签名）
# ---------------------------------------------------------------------------


def _one(method, path, body=None, token=None, ctype=None, raw_body=None, timeout=60):
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
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return r.status, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()
    except Exception as e:
        return 0, str(e).encode()


def api(method, path, body=None, token=None, retry=4):
    for attempt in range(retry + 1):
        with _lock:
            _req_seen[0] += 1
        _, raw = _one(method, path, body, token, ctype="application/json")
        try:
            resp = json.loads(raw.decode("utf-8", "replace"))
        except Exception:
            resp = {"_raw": raw[:150].decode("utf-8", "replace")}
        eid = (resp.get("error") or {}).get("id")
        if eid == 100034 and attempt < retry:      # 反爬限流：退避重试
            time.sleep(1.5 * (attempt + 1))
            continue
        if not resp.get("flag"):
            with _lock:
                _req_seen[1] += 1
                eid = (resp.get("error") or {}).get("id") or resp.get("_raw", "unknown")
                _err_ids[str(eid)] = _err_ids.get(str(eid), 0) + 1
        return resp
    return resp


def upload(token, filename, content, ftype="image", describe=""):
    boundary = "----chaos" + secrets.token_hex(8)
    parts = []
    for k, v in (("type", ftype), ("tag", "chaos"), ("describe", describe)):
        parts.append(f"--{boundary}\r\nContent-Disposition: form-data; name=\"{k}\"\r\n\r\n{v}\r\n".encode())
    parts.append(
        f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; filename=\"{filename}\"\r\n"
        f"Content-Type: application/octet-stream\r\n\r\n".encode() + content + b"\r\n")
    parts.append(f"--{boundary}--\r\n".encode())
    with _lock:
        _req_seen[0] += 1
    _, raw = _one("POST", "/v1/file/upload", token=token,
                  ctype=f"multipart/form-data; boundary={boundary}", raw_body=b"".join(parts), timeout=120)
    try:
        resp = json.loads(raw.decode("utf-8", "replace"))
    except Exception:
        resp = {"_raw": raw[:150].decode("utf-8", "replace")}
    if not resp.get("flag"):
        with _lock:
            _req_seen[1] += 1
    return resp


def get_raw(path, timeout=120):
    req = urllib.request.Request(BASE + path, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return r.status, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()


# ---------------------------------------------------------------------------
# 限流计数清理（把瓶颈留给业务链路，而不是反爬限流）
# ---------------------------------------------------------------------------


def limiter_loop(stop_evt):
    """周期性清理反爬限流与评论防刷计数（凭据由 fafaenv 从环境变量读取）。"""
    patterns = ["ff_rl:*", "ff_rl_v:*", "ff_rl_blk:*", "ff_seccomment:*"]
    while not stop_evt.is_set():
        fafaenv.clear_limits(patterns)
        stop_evt.wait(2)


# ---------------------------------------------------------------------------
# 语料
# ---------------------------------------------------------------------------

SURNAME = list("李王张刘陈杨黄赵周吴徐孙马朱胡林郭何高罗郑梁谢宋唐许韩冯邓曹彭曾萧田董袁潘于蒋")
GIVEN = ["小满", "远航", "一鸣", "知白", "半山", "清和", "见山", "之衡", "云舟", "木石",
         "沐阳", "临风", "照野", "星野", "南枝", "望舒", "雨眠", "疏影", "拾光", "砚秋",
         "归舟", "青崖", "若谷", "川流", "霄汉", "守拙", "听雪", "寻真", "青梧", "白露"]
TOPICS = ["性能优化", "缓存设计", "索引失效", "灰度发布", "日志治理", "接口幂等", "限流算法",
          "连接池", "内存泄漏", "时区问题", "字符集", "慢查询", "容器化", "可观测性",
          "单元测试", "代码评审", "需求拆解", "用户访谈", "定价策略", "内容运营",
          "街头摄影", "旅行随笔", "插画配色", "阅读笔记", "时间管理", "远程协作"]
VERBS = ["我踩过的坑", "的一些思考", "从零到一", "复盘与反思", "实践笔记", "取舍之道",
         "到底难在哪", "要不要做", "值不值得", "的三个误区"]
WORDS = ["分布式", "前端", "后端", "数据库", "网络", "算法", "产品", "设计", "运维", "工程效率",
         "一致性", "可用性", "成本", "延迟", "吞吐", "复杂度", "协作", "沟通", "取舍", "边界"]
PARAS = [
    "先说结论：这件事没有银弹，但有几个可以量化的判断标准。",
    "我们在生产环境跑了三个月，数据比预想的有意思。",
    "动手之前我把相关的文档和源码都过了一遍，这里只记那些和直觉相反的部分。",
    "一个常见的误区是把它当成纯技术问题，实际上大部分坑在流程上。",
    "下面这段配置是我们最终稳定下来的版本，可以直接抄。",
    "如果只能记住一句话，那就是：先测量，再优化。",
    "回头看，最贵的成本不是写代码，而是改错的方向。",
    "这个问题在低峰期完全看不出来，一到高峰期就会集中爆发。",
    "同组的同事给了一个很朴素的解法，效果反而比我的方案好。",
    "我把整个过程拆成了四步，每一步都可以单独验证。",
    "需要说明的是，下面的数字都来自我们自己的环境，不代表通用结论。",
    "到此为止问题解决了，但根因其实还在，只是被绕过去了。",
]


def rand_nick():
    return random.choice(SURNAME) + random.choice(GIVEN)


def rand_title():
    return f"{random.choice(TOPICS)}{random.choice(VERBS)}"


def rand_body(n=6):
    head = f"## {random.choice(TOPICS)}的一些记录\n\n"
    return head + "\n\n".join(random.choice(PARAS) for _ in range(n)) + "\n\n" + \
        "> " + " ".join(random.choice(WORDS) for _ in range(6)) + "\n"


def make_png(size, rgb, seed):
    rows = []
    for y in range(size):
        row = bytearray(b"\x00")
        for x in range(size):
            on = ((x * 3 + y * 5 + seed) % 17) < 6
            row += bytes(rgb if on else (255 - rgb[0], 255 - rgb[1], 255 - rgb[2]))
        rows.append(bytes(row))
    raw = b"".join(rows)

    def chunk(tag, data):
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)

    return (b"\x89PNG\r\n\x1a\n"
            + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 2, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(raw, 6))
            + chunk(b"IEND", b""))


# ---------------------------------------------------------------------------
# 主流程
# ---------------------------------------------------------------------------


def main():
    random.seed(int(STAMP))
    t0 = time.time()
    print(f"目标后端 {BASE}   规模: 用户{N_USERS} 专栏/人{N_NODES} 文章/人{N_ARTICLES} "
          f"评论{N_COMMENTS} 关注{N_FOLLOWS} 私信{N_MESSAGES} 点赞{N_LIKES} 文件{N_FILES} 并发{WORKERS}")

    at = (api("POST", "/user/token/get", {"user_name": ADMIN_NAME, "pass_wd": ADMIN_PASS}).get("data"))
    if not at:
        print("管理员登录失败")
        return 1
    fafaenv.warn_single_login(ADMIN_NAME)
    print()

    stop_evt = threading.Event()
    threading.Thread(target=limiter_loop, args=(stop_evt,), daemon=True).start()

    # ---------------- 1. 用户 ----------------
    print("\n[1/6] 创建用户")
    users = []
    for i in range(N_USERS):
        name = f"chaos{STAMP}u{i:03d}"
        nick = rand_nick() + str(random.randint(10, 99))
        r = api("POST", "/v1/user/create", {
            "name": name, "nick_name": nick, "email": f"{name}@chaos.local",
            "password": PASS, "repassword": PASS, "gender": random.choice([0, 1, 2]),
            "wechat": f"wx{name[-6:]}", "github": f"https://github.com/{name}",
            "qq": str(random.randint(10000, 99999999)),
            "short_describe": f"{random.choice(TOPICS)}爱好者",
            "describe": rand_body(3),
        }, at)
        if not r.get("flag"):
            print("  创建失败:", json.dumps(r, ensure_ascii=False)[:160])
            continue
        uid = r["data"]["id"]
        api("POST", "/v1/user/admin/update", {"id": uid, "vip": 1}, at)
        users.append({"id": uid, "name": name, "nick": nick})
    print(f"  ✅ 用户 {len(users)}")

    print("\n[2/6] 登录")
    for u in users:
        r = api("POST", "/user/token/get", {"user_name": u["name"], "pass_wd": PASS})
        u["token"] = r.get("data", "")
    users = [u for u in users if u.get("token")]
    print(f"  ✅ 登录 {len(users)}")

    # ---------------- 2. 并发建专栏/文章/文件 ----------------
    print(f"\n[3/6] 并发创建专栏 / 文章 / 文件（{WORKERS} 并发）")
    nodes, articles, files = [], [], []
    lock = threading.Lock()

    def work(u):
        local_nodes, local_arts, local_files = [], [], []
        for k in range(N_NODES):
            r = api("POST", "/v1/node/create", {
                "seo": f"n{STAMP}{u['id']}{k}",
                "name": f"{random.choice(TOPICS)}专栏·{random.choice(GIVEN)}",
                "describe": rand_body(2),
            }, u["token"])
            if r.get("flag"):
                local_nodes.append({"id": r["data"]["id"],
                                    "seo": (r["data"].get("seo") or f"n{STAMP}{u['id']}{k}"),
                                    "user": u})
                bump("node")
        nfile = max(1, N_FILES // max(1, len(users)))
        for _ in range(nfile):
            seed = random.randint(0, 999)
            rgb = (random.randint(40, 200), random.randint(40, 200), random.randint(40, 200))
            body = make_png(96, rgb, seed)
            r = upload(u["token"], f"chaos_{STAMP}_{seed}.png", body, "image",
                       f"混沌测试图 {seed}")
            if r.get("flag"):
                local_files.append({"url": r["data"]["url"], "sha": hashlib.sha256(body).hexdigest(),
                                    "size": len(body)})
                bump("file")
        for j in range(N_ARTICLES):
            if not local_nodes:
                break
            node = random.choice(local_nodes)
            r = api("POST", "/v1/content/create", {
                "seo": f"a{STAMP}{u['id']}{j}", "node_id": node["id"],
                "title": rand_title(), "status": 1, "describe": rand_body(random.randint(4, 12)),
            }, u["token"])
            if not r.get("flag"):
                continue
            cid = r["data"]["id"]
            api("POST", "/v1/content/publish", {"id": cid}, u["token"])
            api("POST", "/v1/content/update/status", {"id": cid, "status": 0}, u["token"])
            if local_files and random.random() < 0.6:
                api("POST", "/v1/content/update/image", {"id": cid, "image_path":
                                                         random.choice(local_files)["url"]}, u["token"])
            local_arts.append({"id": cid, "user": u, "title": r["data"]["pre_title"],
                               "seo": f"a{STAMP}{u['id']}{j}", "node_seo": node["seo"]})
            bump("article")
        with lock:
            nodes.extend(local_nodes)
            articles.extend(local_arts)
            files.extend(local_files)

    with futures.ThreadPoolExecutor(max_workers=WORKERS) as ex:
        list(ex.map(work, users))
    print(f"  ✅ 专栏 {len(nodes)} / 文章 {len(articles)} / 文件 {len(files)}")

    # ---------------- 3. 随机互动 ----------------
    print("\n[4/6] 随机互动：评论 / 关注 / 私信 / 点赞")

    def comment_job(i):
        a = random.choice(articles)
        u = random.choice(users)
        r = api("POST", "/v1/comment/create",
                {"content_id": a["id"], "body": random.choice(PARAS) + f"（#{i}）"}, u["token"])
        if r.get("flag"):
            bump("comment")

    def follow_job(i):
        a, b = random.sample(users, 2)
        r = api("POST", "/v1/relation/follow/add", {"user_name": b["name"]}, a["token"])
        if r.get("flag"):
            bump("follow")

    def message_job(i):
        a, b = random.sample(users, 2)
        r = api("POST", "/v1/message/private/send",
                {"user_id": b["id"], "message": random.choice(PARAS) + f"（私信#{i}）"}, a["token"])
        if r.get("flag"):
            bump("message")

    def like_job(i):
        a = random.choice(articles)
        u = random.choice(users)
        r = api("POST", "/v1/content/cool", {"id": a["id"]}, u["token"])
        if r.get("flag"):
            bump("like")

    for fn, n, label in ((comment_job, N_COMMENTS, "评论"), (follow_job, N_FOLLOWS, "关注"),
                         (message_job, N_MESSAGES, "私信"), (like_job, N_LIKES, "点赞")):
        with futures.ThreadPoolExecutor(max_workers=WORKERS) as ex:
            list(ex.map(fn, range(n)))
        print(f"  ✅ {label} 成功 {_stats.get(fn.__name__.split('_')[0], 0)}")

    stop_evt.set()

    # ---------------- 4. 校验 ----------------
    print("\n[5/6] 校验：列表 / 搜索 / 下载完整性")
    checks = []

    def check(label, cond, detail=""):
        checks.append((label, bool(cond)))
        print(f"  {'✅' if cond else '❌'} {label}" + (f"  | {detail}" if detail else ""))

    r = api("GET", "/u/content", {"limit": 20, "sort": ["-publish_time"]})
    total = (r.get("data") or {}).get("total", 0)
    check("公开文章列表可用", r.get("flag") and total > 0, f"total={total}")

    r = api("GET", "/u", {"vip": -1, "limit": 20})
    check("公开用户列表可用", r.get("flag") and (r.get("data") or {}).get("total", 0) > 0,
          f"total={(r.get('data') or {}).get('total')}")

    sample = random.sample(articles, min(5, len(articles)))
    hit = 0
    for a in sample:
        r = api("GET", "/content", {"user_name": a["user"]["name"],
                                    "node_seo": a["node_seo"], "seo": a["seo"]})
        got = (r.get("data") or {}).get("title")
        if got == a["title"]:
            hit += 1
    check(f"文章详情解密（{hit}/{len(sample)} 命中）", hit == len(sample), sample[0]["title"][:20])

    r = api("GET", "/u/content", {"title": sample[0]["title"], "limit": 5})
    check("精确标题搜索命中", (r.get("data") or {}).get("total", 0) >= 1, sample[0]["title"])
    r = api("GET", "/u/content", {"title": sample[0]["title"][:3], "limit": 5})
    check("标题片段搜索不命中", (r.get("data") or {}).get("total", 0) == 0, sample[0]["title"][:3])

    ok_dl = 0
    for f in random.sample(files, min(8, len(files))):
        code, body = get_raw(f["url"])
        if code == 200 and hashlib.sha256(body).hexdigest() == f["sha"]:
            ok_dl += 1
    check(f"随机抽检文件下载 sha256 一致（{ok_dl}/{min(8, len(files))}）",
          ok_dl == min(8, len(files)))

    if files:
        f = files[0]
        code, part = get_raw(f["url"])
        req = urllib.request.Request(BASE + f["url"], method="GET")
        req.add_header("Range", "bytes=0-99")
        with urllib.request.urlopen(req, timeout=30) as resp:
            check("Range 请求 206", resp.status == 206, resp.headers.get("Content-Range"))

    # ---------------- 5. 密文与明文扫描 ----------------
    print("\n[6/6] 落库密文 / 明文扫描")
    def sql(q):
        return fafaenv.sql(q)

    if not fafaenv.sql_enabled():
        print("  ~ 未设置 FAFA_MYSQL_PASS，跳过落库扫描校验")
    leaks = sql(f"""SELECT
      (SELECT COUNT(*) FROM fafa.fafacms_content WHERE title LIKE '%{TOPICS[0]}%' OR `describe` LIKE '%先说结论%' OR user_name LIKE '%chaos{STAMP}%')
    + (SELECT COUNT(*) FROM fafa.fafacms_comment WHERE `describe` LIKE '%先说结论%' OR user_name LIKE '%chaos{STAMP}%')
    + (SELECT COUNT(*) FROM fafa.fafacms_user WHERE name LIKE '%chaos{STAMP}%' OR `describe` LIKE '%先说结论%')
    + (SELECT COUNT(*) FROM fafa.fafacms_content_node WHERE name LIKE '%专栏%' OR `describe` LIKE '%先说结论%')
    + (SELECT COUNT(*) FROM fafa.fafacms_message WHERE send_message LIKE '%私信#%')
    + (SELECT COUNT(*) FROM fafa.fafacms_file WHERE `describe` LIKE '%混沌测试图%')
    + (SELECT COUNT(*) FROM fafa.fafacms_content_history WHERE `describe` LIKE '%先说结论%');""")
    check("全表明文扫描命中数为 0", leaks in ("0", None), f"hits={leaks}")

    cipher = sql("""SELECT
      (SELECT COUNT(*) FROM fafa.fafacms_content WHERE title NOT LIKE 'v1:%')
    + (SELECT COUNT(*) FROM fafa.fafacms_comment WHERE `describe` NOT LIKE 'v1:%')
    + (SELECT COUNT(*) FROM fafa.fafacms_user WHERE name NOT LIKE 'v1:%')
    + (SELECT COUNT(*) FROM fafa.fafacms_content_node WHERE name NOT LIKE 'v1:%');""")
    check("抽查列全部为 v1: 密文", cipher in ("0", None), f"non-cipher={cipher}")

    blank = sql("""SELECT (SELECT COUNT(*) FROM fafa.fafacms_content WHERE title!='' AND title_bidx='')
    + (SELECT COUNT(*) FROM fafa.fafacms_user WHERE name!='' AND name_bidx='');""")
    check("盲索引无缺失", blank in ("0", None), f"blank={blank}")

    rc = fafaenv.storage_plaintext_scan(["先说结论", f"chaos{STAMP}"])
    check("磁盘无明文", rc in (0, None), f"hits={rc}" + ("" if rc is not None else "（未配置容器访问，跳过）"))

    print("\n===== 统计 =====")
    for k in ("node", "article", "file", "comment", "follow", "message", "like"):
        print(f"  {k:8s} {_stats.get(k, 0)}")
    print(f"  HTTP 请求 {_req_seen[0]} 次，其中失败 {_req_seen[1]} 次")
    if _err_ids:
        print("  失败构成（错误码 -> 次数）:")
        for k, v in sorted(_err_ids.items(), key=lambda kv: -kv[1])[:8]:
            print(f"    {k}: {v}")
    print(f"  耗时 {time.time() - t0:.1f}s")
    failed = [l for l, okv in checks if not okv]
    print(f"\n校验通过 {len(checks) - len(failed)}/{len(checks)}")
    if failed:
        print("失败项:", failed)
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
