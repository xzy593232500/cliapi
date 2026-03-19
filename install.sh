#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-install}"
TARGET="${2:-CLIAPI}"
APP_DIR="/opt/cliapi"
APP_NAME="CLIAPI"
REPO_URL="${CLIAPI_REPO_URL:-https://github.com/xzy593232500/cliapi.git}"
BRANCH="${CLIAPI_BRANCH:-main}"
DEFAULT_PORT="${CLIAPI_PORT:-8317}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "缺少命令: $1"; exit 1; }
}

randhex() {
  openssl rand -hex 24
}

install_docker_if_needed() {
  if command -v docker >/dev/null 2>&1; then
    echo "[ok] docker 已存在"
    return
  fi
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
}

install_compose_if_needed() {
  if docker compose version >/dev/null 2>&1; then
    echo "[ok] docker compose 已存在"
    return
  fi
  mkdir -p /usr/local/lib/docker/cli-plugins
  ARCH=$(uname -m)
  case "$ARCH" in
    x86_64|amd64) BIN_ARCH="x86_64" ;;
    aarch64|arm64) BIN_ARCH="aarch64" ;;
    *) echo "不支持的架构: $ARCH"; exit 1 ;;
  esac
  curl -L "https://github.com/docker/compose/releases/download/v2.39.1/docker-compose-linux-${BIN_ARCH}" -o /usr/local/lib/docker/cli-plugins/docker-compose
  chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
}

install_go_if_needed() {
  if command -v go >/dev/null 2>&1; then
    echo "[ok] go 已存在"
    return
  fi
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y golang-go
    return
  fi
  if command -v apk >/dev/null 2>&1; then
    apk add --no-cache go
    return
  fi
  echo "缺少 go，且当前系统不支持自动安装，请手动安装 Go 后重试"
  exit 1
}

clone_or_update_repo() {
  if [ -d "$APP_DIR/.git" ]; then
    echo "[info] 检测到已有仓库，拉取最新代码"
    git -C "$APP_DIR" fetch origin
    git -C "$APP_DIR" checkout "$BRANCH"
    git -C "$APP_DIR" pull --ff-only origin "$BRANCH"
    return
  fi
  rm -rf "$APP_DIR"
  git clone --depth=1 --branch "$BRANCH" "$REPO_URL" "$APP_DIR"
}

prepare_layout() {
  mkdir -p "$APP_DIR/auths" "$APP_DIR/logs"

  if [ ! -f "$APP_DIR/.env" ]; then
    cp "$APP_DIR/.env.example" "$APP_DIR/.env"
  fi

  if ! grep -q '^MANAGEMENT_PASSWORD=' "$APP_DIR/.env"; then
    echo "MANAGEMENT_PASSWORD=$(randhex)" >> "$APP_DIR/.env"
  fi

  if ! grep -q '^CLI_PROXY_IMAGE=' "$APP_DIR/.env"; then
    echo "CLI_PROXY_IMAGE=cliapi:local" >> "$APP_DIR/.env"
  else
    sed -i 's#^CLI_PROXY_IMAGE=.*#CLI_PROXY_IMAGE=cliapi:local#' "$APP_DIR/.env"
  fi

  if [ ! -f "$APP_DIR/config.yaml" ]; then
    cp "$APP_DIR/config.example.yaml" "$APP_DIR/config.yaml"
  fi

  sed -i 's/allow-remote: false/allow-remote: true/' "$APP_DIR/config.yaml" || true
  sed -i "s/^port: .*/port: ${DEFAULT_PORT}/" "$APP_DIR/config.yaml" || true
}

build_binary() {
  cd "$APP_DIR"
  echo "[info] 正在本机构建静态二进制..."
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o CLIProxyAPI ./cmd/server
  file CLIProxyAPI || true
}

start_service() {
  cd "$APP_DIR"
  docker compose down --remove-orphans || true
  docker compose build --no-cache
  docker compose up -d --pull never
}

render_notice() {
  local mgmt_password
  mgmt_password=$(grep '^MANAGEMENT_PASSWORD=' "$APP_DIR/.env" | tail -n1 | cut -d= -f2- || true)
  cat <<MSG

${APP_NAME} 已部署到 ${APP_DIR}

管理密钥:
  ${mgmt_password}

建议检查文件:
  ${APP_DIR}/.env
  ${APP_DIR}/config.yaml

建议补充:
  CPA_QUOTA_POSTGRES_DSN=postgresql://user:pass@host:5432/cliapi

当前访问地址:
  http://服务器IP:${DEFAULT_PORT}/
  http://服务器IP:${DEFAULT_PORT}/management.html

额度管理:
  已整合进 management.html 页面内

本机验证:
  curl http://127.0.0.1:${DEFAULT_PORT}/
  curl http://127.0.0.1:${DEFAULT_PORT}/management.html

查看日志:
  cd ${APP_DIR} && docker compose logs -f --tail=100
MSG
}

main_install() {
  need_cmd curl
  need_cmd git
  need_cmd openssl
  install_docker_if_needed
  install_compose_if_needed
  install_go_if_needed
  clone_or_update_repo
  prepare_layout
  build_binary
  start_service
  render_notice
}

main_update() {
  need_cmd git
  install_go_if_needed
  clone_or_update_repo
  prepare_layout
  build_binary
  start_service
  echo "更新完成"
  render_notice
}

main_uninstall() {
  if [ -d "$APP_DIR" ]; then
    cd "$APP_DIR"
    docker compose down -v || true
    rm -rf "$APP_DIR"
  fi
  echo "已卸载"
}

case "$ACTION" in
  install|app)
    main_install
    ;;
  update)
    main_update
    ;;
  uninstall)
    main_uninstall
    ;;
  *)
    echo "用法: bash <(curl -sL https://raw.githubusercontent.com/xzy593232500/cliapi/main/install.sh) [install|app|update|uninstall] CLIAPI"
    exit 1
    ;;
esac
