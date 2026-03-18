#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
OUT_DIR="$ROOT_DIR/release"
mkdir -p "$OUT_DIR"

VERSION="${1:-latest}"
TAR_NAME="cliapi-${VERSION}.tar.gz"
LATEST_NAME="cliapi-latest.tar.gz"

rm -f "$OUT_DIR/$TAR_NAME" "$OUT_DIR/$LATEST_NAME"

tar \
  --exclude='./release' \
  --exclude='./.git' \
  --exclude='./.env' \
  --exclude='./node_modules' \
  -czf "$OUT_DIR/$TAR_NAME" \
  -C "$ROOT_DIR" \
  .

if [ "$TAR_NAME" != "$LATEST_NAME" ]; then
  cp "$OUT_DIR/$TAR_NAME" "$OUT_DIR/$LATEST_NAME"
fi

echo "已生成:"
echo "  $OUT_DIR/$TAR_NAME"
echo "  $OUT_DIR/$LATEST_NAME"
