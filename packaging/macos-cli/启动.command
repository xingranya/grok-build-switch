#!/bin/bash

set -u

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$BASE_DIR/grok_switch"
DATA_DIR="$HOME/.grok_switch"
SETTINGS_FILE="$DATA_DIR/settings.json"
PID_FILE="$DATA_DIR/headless.pid"
APP_LOG="$DATA_DIR/grok_switch.log"
LAUNCH_LOG="$DATA_DIR/headless-launch.log"
DEFAULT_PORT=17878

read_actual_port() {
  local port=""
  if [[ -f "$SETTINGS_FILE" ]]; then
    port="$(/usr/bin/plutil -extract actual_port raw -o - "$SETTINGS_FILE" 2>/dev/null)"
  fi
  if [[ ! "$port" =~ ^[0-9]+$ ]]; then
    port="$DEFAULT_PORT"
  fi
  printf '%s' "$port"
}

server_ready() {
  /usr/bin/curl --silent --fail --max-time 1 "http://127.0.0.1:$1/api/status" >/dev/null 2>&1
}

open_panel() {
  local url="http://127.0.0.1:$1"
  if ! /usr/bin/open "$url"; then
    echo "浏览器未能自动打开，请手动访问：$url"
  fi
  echo "Grok Build Switch 已运行：$url"
}

/bin/mkdir -p "$DATA_DIR"

port="$(read_actual_port)"
if server_ready "$port"; then
  open_panel "$port"
  exit 0
fi
if [[ "$port" != "$DEFAULT_PORT" ]] && server_ready "$DEFAULT_PORT"; then
  open_panel "$DEFAULT_PORT"
  exit 0
fi

if [[ ! -x "$BINARY" ]]; then
  echo "启动失败：找不到可执行文件 $BINARY"
  exit 1
fi

"$BINARY" --headless --silent >>"$LAUNCH_LOG" 2>&1 &
server_pid=$!
printf '%s\n' "$server_pid" >"$PID_FILE"

cleanup() {
  /bin/rm -f "$PID_FILE"
}

stop_server() {
  /bin/kill -TERM "$server_pid" >/dev/null 2>&1 || true
}

trap stop_server HUP INT TERM
trap cleanup EXIT

for _ in {1..200}; do
  if ! /bin/kill -0 "$server_pid" 2>/dev/null; then
    break
  fi
  port="$(read_actual_port)"
  if server_ready "$port"; then
    open_panel "$port"
    echo "服务正在运行，请保留此终端窗口；窗口可以最小化。"
    wait "$server_pid"
    exit $?
  fi
  /bin/sleep 0.1
done

stop_server
echo "启动失败。应用日志：$APP_LOG"
if [[ -f "$APP_LOG" ]]; then
  /usr/bin/tail -n 20 "$APP_LOG"
fi
if [[ -f "$LAUNCH_LOG" ]]; then
  echo "运行日志：$LAUNCH_LOG"
  /usr/bin/tail -n 40 "$LAUNCH_LOG"
fi
exit 1
