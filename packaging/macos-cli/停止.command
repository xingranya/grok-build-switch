#!/bin/bash

set -u

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
BINARY="$BASE_DIR/grok_switch"
DATA_DIR="$HOME/.grok_switch"
SETTINGS_FILE="$DATA_DIR/settings.json"
PID_FILE="$DATA_DIR/headless.pid"
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

stop_at_port() {
  /usr/bin/curl --silent --fail --max-time 2 \
    --request POST "http://127.0.0.1:$1/api/app/quit" >/dev/null 2>&1
}

port="$(read_actual_port)"
stopped="false"
if stop_at_port "$port" || { [[ "$port" != "$DEFAULT_PORT" ]] && stop_at_port "$DEFAULT_PORT"; }; then
  stopped="true"
fi

if [[ -f "$PID_FILE" ]]; then
  pid="$(/bin/cat "$PID_FILE" 2>/dev/null)"
  if [[ "$pid" =~ ^[0-9]+$ ]] && /bin/kill -0 "$pid" 2>/dev/null; then
    command_line="$(/bin/ps -p "$pid" -o command= 2>/dev/null)"
    if [[ "$command_line" == *"$BINARY"* ]]; then
      /bin/kill -TERM "$pid" >/dev/null 2>&1 || true
      stopped="true"
    fi
  fi
fi
/bin/rm -f "$PID_FILE"

if [[ "$stopped" == "true" ]]; then
  echo "Grok Build Switch 已停止。"
else
  echo "Grok Build Switch 当前没有运行。"
fi
