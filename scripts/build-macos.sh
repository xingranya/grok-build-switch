#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="Grok Build Switch"
BINARY_NAME="grok_switch"
BUNDLE_ID="com.grokbuildswitch.app"
VERSION="${VERSION:-0.4.0}"
BUILD_NUMBER="${BUILD_NUMBER:-1}"
GO_BIN="${GO_BIN:-go}"
SKIP_DMG="${SKIP_DMG:-0}"
DIST_DIR="${DIST_DIR:-$ROOT_DIR/dist}"
APP_BUNDLE="$DIST_DIR/$APP_NAME.app"
CONTENTS_DIR="$APP_BUNDLE/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"
DMG_PATH="$DIST_DIR/Grok-Build-Switch-macos-arm64.dmg"

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
[[ "$VERSION" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]] || fail "VERSION 必须是数字版本，例如 0.4.0"
[[ "$BUILD_NUMBER" =~ ^[0-9]+$ ]] || fail "BUILD_NUMBER 必须是正整数"
[[ "$SKIP_DMG" == "0" || "$SKIP_DMG" == "1" ]] || fail "SKIP_DMG 只能是 0 或 1"
[[ -n "$DIST_DIR" && "$DIST_DIR" != "/" ]] || fail "DIST_DIR 不能是根目录"

require_command "$GO_BIN"
require_command /usr/bin/codesign
require_command /usr/bin/ditto
require_command /usr/bin/lipo
require_command /usr/bin/plutil
require_command /usr/libexec/PlistBuddy
if [[ "$SKIP_DMG" == "0" ]]; then
  require_command /usr/bin/hdiutil
fi

mkdir -p "$DIST_DIR"
rm -rf "$APP_BUNDLE"
rm -f "$DMG_PATH"
mkdir -p "$MACOS_DIR" "$RESOURCES_DIR"

echo "运行 Go 测试..."
(
  cd "$ROOT_DIR"
  "$GO_BIN" test ./...
)

echo "构建 macOS arm64 应用..."
CGO_CFLAGS_VALUE="${CGO_CFLAGS:-}"
CGO_LDFLAGS_VALUE="${CGO_LDFLAGS:-}"
if [[ "$(uname -m)" != "arm64" ]]; then
  require_command /usr/bin/clang
  CGO_CFLAGS_VALUE="$CGO_CFLAGS_VALUE -arch arm64"
  CGO_LDFLAGS_VALUE="$CGO_LDFLAGS_VALUE -arch arm64"
fi
(
  cd "$ROOT_DIR"
  CC=/usr/bin/clang \
    CGO_ENABLED=1 \
    CGO_CFLAGS="$CGO_CFLAGS_VALUE" \
    CGO_LDFLAGS="$CGO_LDFLAGS_VALUE" \
    GOOS=darwin \
    GOARCH=arm64 \
    "$GO_BIN" build \
    -trimpath \
    -ldflags "-s -w" \
    -o "$MACOS_DIR/$BINARY_NAME" .
)

/usr/bin/install -m 0644 "$ROOT_DIR/packaging/macos/Info.plist" "$CONTENTS_DIR/Info.plist"
/usr/bin/install -m 0644 "$ROOT_DIR/packaging/macos/PkgInfo" "$CONTENTS_DIR/PkgInfo"
/usr/bin/install -m 0644 "$ROOT_DIR/assets/AppIcon.icns" "$RESOURCES_DIR/AppIcon.icns"
/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier $BUNDLE_ID" "$CONTENTS_DIR/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString $VERSION" "$CONTENTS_DIR/Info.plist"
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion $BUILD_NUMBER" "$CONTENTS_DIR/Info.plist"
/usr/bin/plutil -lint "$CONTENTS_DIR/Info.plist" >/dev/null

echo "执行 ad-hoc 签名..."
/usr/bin/codesign --force --sign - --timestamp=none "$APP_BUNDLE"

if [[ "$SKIP_DMG" == "1" ]]; then
  "$ROOT_DIR/scripts/verify-macos-artifacts.sh" "$APP_BUNDLE"
  echo "构建完成："
  echo "  $APP_BUNDLE"
  echo "已按 SKIP_DMG=1 跳过 DMG，仅用于受限环境验证。"
  exit 0
fi

DMG_STAGE="$(mktemp -d "${TMPDIR:-/tmp}/grok-build-switch-dmg.XXXXXX")"
cleanup() {
  rm -rf "$DMG_STAGE"
}
trap cleanup EXIT

/usr/bin/ditto "$APP_BUNDLE" "$DMG_STAGE/$APP_NAME.app"
/bin/ln -s /Applications "$DMG_STAGE/Applications"

echo "生成 DMG..."
/usr/bin/hdiutil create \
  -volname "$APP_NAME" \
  -srcfolder "$DMG_STAGE" \
  -ov \
  -format UDZO \
  "$DMG_PATH"

"$ROOT_DIR/scripts/verify-macos-artifacts.sh" "$APP_BUNDLE" "$DMG_PATH"

echo "构建完成："
echo "  $APP_BUNDLE"
echo "  $DMG_PATH"
