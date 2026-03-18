#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-install}"
TARGET="${2:-CLIAPI}"
APP_DIR="/opt/cliapi"
APP_NAME="CLIAPI"
PUBLIC_BASE_URL="${CLIPROXYAPI_PUBLIC_BASE_URL:-https://raw.githubusercontent.com/xzy593232500/cliapi/main}"
TARBALL_URL="${CLIPROXYAPI_TARBALL_URL:-https://github.com/xzy593232500/cliapi/releases/download/v0.1.0/cliapi-latest.tar.gz}"
ENV_URL="${CLIPROXYAPI_ENV_URL:-https://raw.githubusercontent.com/xzy593232500/cliapi/main/.env.example}"

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

download_release() {
  rm -rf "$APP_DIR"
  mkdir -p "$APP_DIR"
  tmpdir=$(mktemp -d)
  curl -fsSL "$TARBALL_URL" -o "$tmpdir/release.tar.gz"
  tar -xzf "$tmpdir/release.tar.gz" -C "$APP_DIR"
  rm -rf "$tmpdir"
}

prepare_env() {
  if [ ! -f "$APP_DIR/.env" ]; then
    curl -fsSL "$ENV_URL" -o "$APP_DIR/.env"
    sed -i "s#^POSTGRES_PASSWORD=.*#POSTGRES_PASSWORD=$(randhex)#" "$APP_DIR/.env"
    sed -i "s#^JWT_SECRET=.*#JWT_SECRET=$(randhex)#" "$APP_DIR/.env"
  fi
}

render_notice() {
  cat <<MSG

${APP_NAME} 已部署到 ${APP_DIR}

接下来请编辑：
  ${APP_DIR}/.env

至少填这几个：
  OPENAI_API_KEY=
  APP_BASE_URL=http://你的域名或IP:${PORT:-8787}

然后执行：
  cd ${APP_DIR} && docker compose up -d --build

健康检查：
  http://服务器IP:${PORT:-8787}/health

初始化管理员：
  cd ${APP_DIR}
  bash scripts/init-super-admin.sh admin@example.com 'StrongPass123!' 100000
MSG
}

main_install() {
  need_cmd curl
  need_cmd tar
  need_cmd openssl
  install_docker_if_needed
  install_compose_if_needed
  download_release
  prepare_env
  cd "$APP_DIR"
  docker compose up -d --build
  render_notice
}

main_update() {
  need_cmd curl
  need_cmd tar
  download_release
  prepare_env
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
    echo "用法: bash <(curl -sL ${PUBLIC_BASE_URL}/install.sh) [install|update|uninstall] CLIAPI"
    exit 1
    ;;
esac
