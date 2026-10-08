#!/usr/bin/env bash
# FaFaCMS 后端一键部署脚本（Mac / Linux 通用）
# 用法: ./deploy.sh [up|down|ps|logs|rebuild]   （默认 up）
set -euo pipefail

# 兼容精简 PATH 环境（某些自动化 shell），确保 ipconfig/ifconfig/hostname 可用
export PATH="/usr/sbin:/sbin:$PATH"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

OS_TYPE="$(uname -s)"

# ---------- 数据持久化目录（不放仓库内，按系统选择） ----------
# macOS: Docker Desktop 默认只共享家目录，放 ~/.fafacms
# Linux: 放系统数据目录 /opt/fafacms
# 可用环境变量 FAFACMS_DATA_DIR 覆盖
if [ -n "${FAFACMS_DATA_DIR:-}" ]; then
  DATA_DIR="$FAFACMS_DATA_DIR"
elif [ "$OS_TYPE" = "Darwin" ]; then
  DATA_DIR="${HOME}/.fafacms"
else
  DATA_DIR="/opt/fafacms"
fi
export DATA_DIR

# ---------- docker compose 命令探测（v2 插件优先，v1 回退） ----------
if docker compose version >/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DC="docker-compose"
else
  echo "❌ 未找到 docker compose / docker-compose，请先安装 Docker Desktop / Docker Engine。" >&2
  exit 1
fi

ensure_network() {
  if ! docker network inspect fafa-net >/dev/null 2>&1; then
    docker network create fafa-net >/dev/null
    echo "🔗 已创建共享网络 fafa-net（供前端仓库跨项目访问后端）"
  fi
}

lan_ip() {
  if [ "$OS_TYPE" = "Darwin" ]; then
    local i ip
    for i in en0 en1; do
      ip="$(ipconfig getifaddr "$i" 2>/dev/null || true)"
      [ -n "$ip" ] && { echo "$ip"; return; }
    done
  else
    hostname -I 2>/dev/null | awk '{print $1}'
  fi
}

print_urls() {
  local ip; ip="$(lan_ip)"
  echo ""
  echo "======================================================"
  echo " ✅ FaFaCMS 后端服务已就绪"
  echo "------------------------------------------------------"
  echo "   后端 API    http://127.0.0.1:8080"
  echo "   接口文档    http://127.0.0.1:9889"
  echo "   phpMyAdmin  http://127.0.0.1:8000   (root/123456789)"
  if [ -n "$ip" ]; then
    echo "------------------------------------------------------"
    echo "   局域网访问（可转发给同事）:"
    echo "   后端 API    http://$ip:8080"
    echo "   接口文档    http://$ip:9889"
    echo "   phpMyAdmin  http://$ip:8000"
  fi
  echo "======================================================"
}

prepare() {
  mkdir -p "$DATA_DIR/mysql/data" "$DATA_DIR/mysql/conf" \
           "$DATA_DIR/redis/data" "$DATA_DIR/redis/conf" \
           "$DATA_DIR/backend/storage" "$DATA_DIR/backend/storage_x" "$DATA_DIR/backend/log"
  # 容器内 MySQL/Redis 以非 root 用户运行，放开写权限，避免挂载目录权限问题。
  # 注意：MySQL 容器生成的文件属于 root，非 root 运维执行 chmod 必然报"不允许的操作"；
  # 这类文件本来就已可写，不能因为 set -e 让整个部署中断。
  chmod -R 777 "$DATA_DIR" 2>/dev/null || true
  # config.yaml 仅在首次部署时写入；之后保留本地修改（含真实邮箱密码等凭据），避免 rebuild 覆盖
  if [ ! -f "$DATA_DIR/backend/config.yaml" ]; then
    cp -f config.yaml "$DATA_DIR/backend/config.yaml"
  else
    echo "📄 已存在 $DATA_DIR/backend/config.yaml，保留本地配置（如需重置请手动删除该文件）"
  fi
  cp -f my.cnf       "$DATA_DIR/mysql/conf/my.cnf"
  cp -f redis.conf   "$DATA_DIR/redis/conf/redis.conf"
  echo "📁 数据目录: $DATA_DIR"
}

# ---------- 数据加密根密钥（FAFACMS_KEK_ROOT） ----------
# 用于派生字段加密与盲索引子密钥。必须持久保存：
# 一旦丢失，数据库中的密文将永久无法解密；切勿提交进仓库。
# 优先级：环境变量 > $DATA_DIR/secret/kek.key > 自动生成并落盘（600 权限）
ensure_kek() {
  local kek_file="$DATA_DIR/secret/kek.key"

  if [ -n "${FAFACMS_KEK_ROOT:-}" ]; then
    echo "🔐 加密根密钥：使用环境变量 FAFACMS_KEK_ROOT"
  elif [ -f "$kek_file" ]; then
    FAFACMS_KEK_ROOT="$(cat "$kek_file")"
    echo "🔐 加密根密钥：使用已存在文件 $kek_file"
  else
    mkdir -p "$(dirname "$kek_file")"
    if command -v openssl >/dev/null 2>&1; then
      FAFACMS_KEK_ROOT="$(openssl rand -base64 32)"
    else
      FAFACMS_KEK_ROOT="$(head -c 32 /dev/urandom | base64 | tr -d '\n')"
    fi
    printf '%s' "$FAFACMS_KEK_ROOT" > "$kek_file"
    echo "🔐 加密根密钥：已生成新密钥 → $kek_file"
    echo "   ⚠️  请立即备份该文件！丢失后所有已加密数据将永久不可解密。"
  fi

  # 该目录不挂载进任何容器，收紧权限
  chmod 700 "$(dirname "$kek_file")" 2>/dev/null || true
  chmod 600 "$kek_file" 2>/dev/null || true
  export FAFACMS_KEK_ROOT
}

up() {
  prepare
  ensure_kek
  ensure_network
  echo "🚀 构建并启动服务（首次需拉取镜像，可能较慢）..."
  # docker.io 偶发 TLS 超时，重试
  local n=0
  until $DC up -d --build; do
    n=$((n+1))
    if [ $n -ge 4 ]; then
      echo "❌ 启动失败，请检查网络 / 镜像拉取。" >&2
      exit 1
    fi
    echo "⚠️  启动失败，6 秒后重试 ($n/3)..."
    sleep 6
  done
  print_urls
}

case "${1:-up}" in
  up)      up ;;
  down)    $DC down ;;
  ps)      $DC ps ;;
  logs)    $DC logs -f --tail 50 ;;
  rebuild) prepare; ensure_kek; ensure_network; $DC up -d --build --force-recreate; print_urls ;;
  *)       echo "用法: $0 [up|down|ps|logs|rebuild]" ;;
esac
