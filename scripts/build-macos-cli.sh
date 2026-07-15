#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-0.4.4}"
MIN_MACOS_VERSION="${MIN_MACOS_VERSION:-12.0}"
GO_BIN="${GO_BIN:-go}"
DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
PACKAGE_NAME="Grok Build Switch CLI"
PACKAGE_DIR="$DIST_DIR/$PACKAGE_NAME"
ZIP_PATH="$DIST_DIR/Grok-Build-Switch-macos-arm64-cli-v$VERSION.zip"
BINARY="$PACKAGE_DIR/grok_switch"

fail() {
  echo "错误：$*" >&2
  exit 1
}

require_command() {
  local command_name="$1"
  if [[ "$command_name" == */* ]]; then
    [[ -x "$command_name" ]] || fail "找不到可执行命令：$command_name"
    return
  fi
  command -v "$command_name" >/dev/null 2>&1 || fail "缺少命令：$command_name"
}

[[ "$(uname -s)" == "Darwin" ]] || fail "macOS 构建脚本只能在 macOS 上运行"
[[ "$VERSION" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]] || fail "VERSION 必须是数字版本，例如 0.4.4"
[[ "$MIN_MACOS_VERSION" =~ ^[0-9]+\.[0-9]+$ ]] || fail "MIN_MACOS_VERSION 必须是主版本和次版本，例如 12.0"
[[ -n "$DIST_DIR" && "$DIST_DIR" != "/" ]] || fail "DIST_DIR 不能是根目录"

require_command "$GO_BIN"
require_command /usr/bin/clang
require_command /usr/bin/codesign
require_command /usr/bin/ditto
require_command /usr/bin/lipo
require_command /usr/bin/unzip
require_command /usr/bin/vtool

/bin/mkdir -p "$DIST_DIR"
/bin/rm -rf "$PACKAGE_DIR"
/bin/rm -f "$ZIP_PATH"
/bin/mkdir -p "$PACKAGE_DIR"

echo "运行 Go 测试..."
(
  cd "$ROOT_DIR"
  "$GO_BIN" test ./...
)

CGO_CFLAGS_VALUE="${CGO_CFLAGS:-} -mmacosx-version-min=$MIN_MACOS_VERSION"
CGO_LDFLAGS_VALUE="${CGO_LDFLAGS:-} -mmacosx-version-min=$MIN_MACOS_VERSION"
if [[ "$(uname -m)" != "arm64" ]]; then
  CGO_CFLAGS_VALUE="$CGO_CFLAGS_VALUE -arch arm64"
  CGO_LDFLAGS_VALUE="$CGO_LDFLAGS_VALUE -arch arm64"
fi

echo "构建纯后台 macOS arm64 可执行文件..."
(
  cd "$ROOT_DIR"
  MACOSX_DEPLOYMENT_TARGET="$MIN_MACOS_VERSION" \
    CC=/usr/bin/clang \
    CGO_ENABLED=1 \
    CGO_CFLAGS="$CGO_CFLAGS_VALUE" \
    CGO_LDFLAGS="$CGO_LDFLAGS_VALUE" \
    GOOS=darwin \
    GOARCH=arm64 \
    "$GO_BIN" build \
    -trimpath \
    -ldflags "-s -w" \
    -o "$BINARY" .
)

/usr/bin/install -m 0755 "$ROOT_DIR/packaging/macos-cli/启动.command" "$PACKAGE_DIR/启动.command"
/usr/bin/install -m 0755 "$ROOT_DIR/packaging/macos-cli/停止.command" "$PACKAGE_DIR/停止.command"
/usr/bin/install -m 0644 "$ROOT_DIR/packaging/macos-cli/使用说明.txt" "$PACKAGE_DIR/使用说明.txt"
/usr/bin/codesign --force --sign - --timestamp=none "$BINARY"
printf '%s\n' "$VERSION" >"$PACKAGE_DIR/版本.txt"

ARCHS="$(/usr/bin/lipo -archs "$BINARY")"
[[ " $ARCHS " == *" arm64 "* ]] || fail "二进制不包含 arm64：$ARCHS"
ACTUAL_MIN_VERSION="$(/usr/bin/vtool -show-build "$BINARY" | /usr/bin/awk '$1 == "minos" { print $2; exit }')"
[[ "$ACTUAL_MIN_VERSION" == "$MIN_MACOS_VERSION" ]] || fail "二进制最低系统版本不正确：$ACTUAL_MIN_VERSION"
/usr/bin/codesign --verify --strict --verbose=2 "$BINARY"

/usr/bin/ditto -c -k --sequesterRsrc --keepParent "$PACKAGE_DIR" "$ZIP_PATH"
/usr/bin/unzip -t "$ZIP_PATH" >/dev/null

echo "构建完成：$ZIP_PATH"
