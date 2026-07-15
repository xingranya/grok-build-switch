#!/bin/bash

set -euo pipefail

APP_BUNDLE="${1:-}"
DMG_PATH="${2:-}"

fail() {
  echo "验证失败：$*" >&2
  exit 1
}

[[ -n "$APP_BUNDLE" && -d "$APP_BUNDLE" ]] || fail "缺少 .app 目录"

PLIST="$APP_BUNDLE/Contents/Info.plist"
BINARY="$APP_BUNDLE/Contents/MacOS/grok_switch"
ICON="$APP_BUNDLE/Contents/Resources/AppIcon.icns"

[[ -f "$PLIST" ]] || fail "缺少 Info.plist"
[[ -x "$BINARY" ]] || fail "缺少可执行二进制"
[[ -s "$ICON" ]] || fail "缺少应用图标"

/usr/bin/plutil -lint "$PLIST" >/dev/null
EXECUTABLE_NAME="$(/usr/bin/plutil -extract CFBundleExecutable raw -o - "$PLIST")"
BUNDLE_ID="$(/usr/bin/plutil -extract CFBundleIdentifier raw -o - "$PLIST")"
MENU_BAR_ONLY="$(/usr/bin/plutil -extract LSUIElement raw -o - "$PLIST")"
[[ "$EXECUTABLE_NAME" == "grok_switch" ]] || fail "CFBundleExecutable 不正确：$EXECUTABLE_NAME"
[[ "$BUNDLE_ID" == "com.grokbuildswitch.app" ]] || fail "CFBundleIdentifier 不正确：$BUNDLE_ID"
[[ "$MENU_BAR_ONLY" == "true" ]] || fail "LSUIElement 必须为 true"

ARCHS="$(/usr/bin/lipo -archs "$BINARY")"
[[ " $ARCHS " == *" arm64 "* ]] || fail "二进制不包含 arm64：$ARCHS"
/usr/bin/codesign --verify --strict --verbose=2 "$APP_BUNDLE"

if [[ -n "$DMG_PATH" ]]; then
  [[ -f "$DMG_PATH" ]] || fail "缺少 DMG 文件"
  /usr/bin/hdiutil imageinfo "$DMG_PATH" >/dev/null
  echo "验证通过：arm64、Info.plist、ad-hoc 签名和 DMG 均有效"
  exit 0
fi

echo "验证通过：arm64、Info.plist 和 ad-hoc 签名均有效；未要求检查 DMG"
