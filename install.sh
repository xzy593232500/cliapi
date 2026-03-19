#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-install}"
TARGET="${2:-CLIAPI}"
APP_DIR="/opt/cliapi"
APP_NAME="CLIAPI"
REPO_URL="${CLIAPI_REPO_URL:-https://github.com/xzy593232500/cliapi.git}"
BRANCH="${CLIAPI_BRANCH:-main}"
DEFAULT_PORT="${CLIAPI_PORT:-8317}"
SELECTED_PORT="${CLIAPI_PORT:-}"
SELECTED_MANAGEMENT_PASSWORD="${CLIAPI_MANAGEMENT_PASSWORD:-}"

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
    DEBIAN_FRONTEND=noninteractive apt-get install -y golang-go file
    return
  fi
  if command -v apk >/dev/null 2>&1; then
    apk add --no-cache go file
    return
  fi
  echo "缺少 go，且当前系统不支持自动安装，请手动安装 Go 后重试"
  exit 1
}

prompt_install_settings() {
  if [ -z "$SELECTED_PORT" ]; then
    read -r -p "输入应用对外服务端口，回车默认使用 ${DEFAULT_PORT} 端口: " input_port || true
    SELECTED_PORT="${input_port:-$DEFAULT_PORT}"
  fi
  if [ -z "$SELECTED_MANAGEMENT_PASSWORD" ]; then
    read -r -p "请输入管理密钥，回车则自动生成: " input_password || true
    SELECTED_MANAGEMENT_PASSWORD="${input_password:-$(randhex)}"
  fi
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

  if grep -q '^MANAGEMENT_PASSWORD=' "$APP_DIR/.env"; then
    sed -i "s#^MANAGEMENT_PASSWORD=.*#MANAGEMENT_PASSWORD=${SELECTED_MANAGEMENT_PASSWORD}#" "$APP_DIR/.env"
  else
    echo "MANAGEMENT_PASSWORD=${SELECTED_MANAGEMENT_PASSWORD}" >> "$APP_DIR/.env"
  fi

  if grep -q '^CLI_PROXY_IMAGE=' "$APP_DIR/.env"; then
    sed -i 's#^CLI_PROXY_IMAGE=.*#CLI_PROXY_IMAGE=cliapi:local#' "$APP_DIR/.env"
  else
    echo "CLI_PROXY_IMAGE=cliapi:local" >> "$APP_DIR/.env"
  fi

  if grep -q '^CLIAPI_PORT=' "$APP_DIR/.env"; then
    sed -i "s#^CLIAPI_PORT=.*#CLIAPI_PORT=${SELECTED_PORT}#" "$APP_DIR/.env"
  else
    echo "CLIAPI_PORT=${SELECTED_PORT}" >> "$APP_DIR/.env"
  fi

  if [ ! -f "$APP_DIR/config.yaml" ]; then
    cp "$APP_DIR/config.example.yaml" "$APP_DIR/config.yaml"
  fi

  sed -i 's/allow-remote: false/allow-remote: true/' "$APP_DIR/config.yaml" || true
  sed -i "s/^port: .*/port: ${SELECTED_PORT}/" "$APP_DIR/config.yaml" || true
}

build_binary() {
  cd "$APP_DIR"
  echo "[info] 正在本机构建静态二进制..."
  rm -f CLIProxyAPI
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
  cat <<MSG

${APP_NAME} 已部署到 ${APP_DIR}

管理密钥:
  ${SELECTED_MANAGEMENT_PASSWORD}

建议检查文件:
  ${APP_DIR}/.env
  ${APP_DIR}/config.yaml

建议补充:
  CPA_QUOTA_POSTGRES_DSN=postgresql://user:pass@host:5432/cliapi

当前访问地址:
  http://服务器IP:${SELECTED_PORT}/
  http://服务器IP:${SELECTED_PORT}/management.html

额度管理:
  已整合进 management.html 页面内

本机验证:
  curl http://127.0.0.1:${SELECTED_PORT}/
  curl http://127.0.0.1:${SELECTED_PORT}/management.html
  curl -H "Authorization: Bearer ${SELECTED_MANAGEMENT_PASSWORD}" http://127.0.0.1:${SELECTED_PORT}/v0/management/config

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
  prompt_install_settings
  clone_or_update_repo
  prepare_layout
  build_binary
  start_service
  render_notice
}

main_update() {
  need_cmd git
  install_go_if_needed
  prompt_install_settings
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
