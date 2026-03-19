#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-install}"
TARGET="${2:-CLIAPI}"
APP_DIR="/opt/cliapi"
APP_NAME="CLIAPI"
REPO_URL="${CLIAPI_REPO_URL:-https://github.com/xzy593232500/cliapi.git}"
BRANCH="${CLIAPI_BRANCH:-main}"

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
    if grep -q '^# MANAGEMENT_PASSWORD=' "$APP_DIR/.env"; then
      sed -i "s#^# MANAGEMENT_PASSWORD=.*##g" "$APP_DIR/.env" >/dev/null 2>&1 || true
    fi
    echo "MANAGEMENT_PASSWORD=$(randhex)" >> "$APP_DIR/.env"
  fi

  if ! grep -q '^MANAGEMENT_PASSWORD=' "$APP_DIR/.env"; then
    echo "MANAGEMENT_PASSWORD=$(randhex)" >> "$APP_DIR/.env"
  fi

  if [ ! -f "$APP_DIR/config.yaml" ]; then
    cp "$APP_DIR/config.example.yaml" "$APP_DIR/config.yaml"
  fi
}

render_notice() {
  MGMT_PASSWORD=$(grep '^MANAGEMENT_PASSWORD=' "$APP_DIR/.env" | tail -n1 | cut -d= -f2- || true)
  cat <<MSG

${APP_NAME} 已部署到 ${APP_DIR}

下一步请重点检查这两个文件：
  ${APP_DIR}/.env
  ${APP_DIR}/config.yaml

至少建议补这些配置：
  MANAGEMENT_PASSWORD=${MGMT_PASSWORD}
  CPA_QUOTA_POSTGRES_DSN=postgresql://user:pass@host:5432/cliapi

如果你希望远程访问管理后台，请在 config.yaml 里确认：
  remote-management:
    allow-remote: true

启动 / 更新服务：
  cd ${APP_DIR} && docker compose up -d --build

查看日志：
  cd ${APP_DIR} && docker compose logs -f

基础地址：
  http://服务器IP:8317/

原后台：
  http://服务器IP:8317/management.html

额度后台：
  http://服务器IP:8317/management-quota.html

用户余额接口：
  GET http://服务器IP:8317/v1/quota/balance

用户兑换接口：
  POST http://服务器IP:8317/v1/quota/redeem
MSG
}

main_install() {
  need_cmd curl
  need_cmd git
  need_cmd openssl
  install_docker_if_needed
  install_compose_if_needed
  clone_or_update_repo
  prepare_layout
  cd "$APP_DIR"
  docker compose up -d --build
  render_notice
}

main_update() {
  need_cmd git
  clone_or_update_repo
  prepare_layout
  cd "$APP_DIR"
  docker compose up -d --build
  echo "更新完成"
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
