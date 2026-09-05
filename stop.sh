#!/usr/bin/env bash
# stop.sh — stop only THIS project's backend/frontend.
# Reads PIDs/ports from .run/ and verifies the cwd before killing, so it never
# kills guaardvark or any other project by accident.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="$ROOT/.run"

port_cwd() {
  local pid
  pid="$(lsof -ti ":$1" 2>/dev/null | head -1)"
  [ -z "$pid" ] && return 0
  lsof -p "$pid" 2>/dev/null | awk '$4=="cwd"{print $NF; exit}'
}

kill_pid_if_ours() {
  local pid="$1"
  local expected_cwd="$2"
  if [ -z "$pid" ] || ! ps -p "$pid" >/dev/null 2>&1; then
    return 0
  fi
  local cwd
  cwd="$(lsof -p "$pid" 2>/dev/null | awk '$4=="cwd"{print $NF; exit}')"
  if [ "$cwd" = "$expected_cwd" ] || [ "$cwd" = "$ROOT" ]; then
    kill "$pid" 2>/dev/null || true
    return 0
  fi
  echo "Refusing to kill pid $pid (cwd: $cwd) — not this project."
}

# --- Backend ---
if [ -f "$RUN_DIR/backend.pid" ]; then
  echo "Stopping backend ..."
  kill_pid_if_ours "$(cat "$RUN_DIR/backend.pid")" "$ROOT/server"
  rm -f "$RUN_DIR/backend.pid" "$RUN_DIR/backend.port" "$RUN_DIR/backend.url"
else
  echo "Backend pid file not found; nothing to stop."
fi

# --- Frontend ---
if [ -f "$RUN_DIR/frontend.pid" ]; then
  echo "Stopping frontend ..."
  kill_pid_if_ours "$(cat "$RUN_DIR/frontend.pid")" "$ROOT"
  rm -f "$RUN_DIR/frontend.pid" "$RUN_DIR/frontend.port" "$RUN_DIR/frontend.url"
else
  echo "Frontend pid file not found; nothing to stop."
fi

echo "Done."
