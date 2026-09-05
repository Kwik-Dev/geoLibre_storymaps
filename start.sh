#!/usr/bin/env bash
# start.sh — start the GeoLibre Storymaps backend and frontend.
# Never kills or reuses ports held by other projects. If :5173 is taken by some
# other app, the frontend starts on the next free port (5174, 5175, ...).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOGS="$ROOT/logs"
RUN_DIR="$ROOT/.run"
mkdir -p "$LOGS" "$RUN_DIR"

# Load a single shared .env (frontend + backend) if present. `set -a` exports
# every var so the Go backend (which reads os.Getenv) and the nohup'd child
# process inherit them. Vite reads .env itself for the frontend. Copy
# .env.example → .env and fill in your values.
if [ -f "$ROOT/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

# Backend env (overridable, e.g. JWT_SECRET=... ./start.sh)
export JWT_SECRET="${JWT_SECRET:-dev-secret}"
export ADMIN_EMAIL="${ADMIN_EMAIL:-admin@example.com}"
export ADMIN_PASSWORD="${ADMIN_PASSWORD:-change-me}"
export APP_PORT="${APP_PORT:-8080}"

# Print the working directory of the process listening on a port, or empty if
# nothing is listening there.
port_cwd() {
  local pid
  pid="$(lsof -ti ":$1" 2>/dev/null | head -1)"
  [ -z "$pid" ] && return 0
  lsof -p "$pid" 2>/dev/null | awk '$4=="cwd"{print $NF; exit}'
}

# Return 0 if the port is free, 1 if occupied.
port_in_use() {
  lsof -ti ":$1" >/dev/null 2>&1
}

# Pick the first free frontend port starting from FRONTEND_PORT or 5173.
find_frontend_port() {
  local port="${FRONTEND_PORT:-5173}"
  while port_in_use "$port"; do
    local cwd
    cwd="$(port_cwd "$port")"
    if [ "$cwd" = "$ROOT" ]; then
      echo "$port"
      return 0
    fi
    port=$((port + 1))
  done
  echo "$port"
}

# --- Backend (Go server) ---
BE_CWD="$(port_cwd "$APP_PORT")"
BACKEND_URL="http://localhost:${APP_PORT}"
if [ -n "$BE_CWD" ]; then
  if [ "$BE_CWD" = "$ROOT/server" ] || [ "$BE_CWD" = "$ROOT" ]; then
    echo "Backend already running on :$APP_PORT (this project, skip)."
  else
    echo "WARNING: :$APP_PORT is used by a DIFFERENT process (cwd: $BE_CWD)."
    echo "  Not starting this project's backend on :$APP_PORT."
    echo "  Stop that process first, or set APP_PORT to a different port."
  fi
else
  echo "Building backend binary ..."
  (cd "$ROOT/server" && go build -o "$ROOT/logs/storymap-server" ./cmd/server)
  echo "Starting backend (Go server) on :$APP_PORT ..."
  (
    cd "$ROOT/server"
    nohup env JWT_SECRET="$JWT_SECRET" ADMIN_EMAIL="$ADMIN_EMAIL" ADMIN_PASSWORD="$ADMIN_PASSWORD" APP_PORT="$APP_PORT" \
      "$ROOT/logs/storymap-server" > "$LOGS/backend.log" 2>&1 < /dev/null &
    echo $! > "$RUN_DIR/backend.pid"
    disown
  )
  echo "$APP_PORT" > "$RUN_DIR/backend.port"
  echo "$BACKEND_URL" > "$RUN_DIR/backend.url"
  echo "  log: $LOGS/backend.log"
fi

# --- Frontend (Vite dev server) ---
FRONTEND_PORT="$(find_frontend_port)"
FRONTEND_URL="http://localhost:${FRONTEND_PORT}"
FE_CWD="$(port_cwd "$FRONTEND_PORT")"
if [ -n "$FE_CWD" ] && [ "$FE_CWD" = "$ROOT" ]; then
  echo "Frontend already running on :$FRONTEND_PORT (this project, skip)."
elif [ -n "$FE_CWD" ] && [ "$FE_CWD" != "$ROOT" ]; then
  echo "WARNING: :$FRONTEND_PORT is used by a DIFFERENT process (cwd: $FE_CWD)."
  echo "  This project's frontend will start on the next free port instead."
else
  if [ "$FRONTEND_PORT" != "${FRONTEND_PORT:-5173}" ]; then
    :
  fi
fi

if ! port_in_use "$FRONTEND_PORT"; then
  echo "Starting frontend (Vite) on :$FRONTEND_PORT ..."
  (
    cd "$ROOT"
    nohup npm run dev -- --host --port "$FRONTEND_PORT" > "$LOGS/frontend.log" 2>&1 < /dev/null &
    echo $! > "$RUN_DIR/frontend.pid"
    disown
  )
  echo "$FRONTEND_PORT" > "$RUN_DIR/frontend.port"
  echo "$FRONTEND_URL" > "$RUN_DIR/frontend.url"
  echo "  log: $LOGS/frontend.log"
fi

echo
echo "Backend:  $BACKEND_URL  (health: /api/health)"
echo "Frontend: $FRONTEND_URL"
if [ "$FRONTEND_PORT" != "5173" ]; then
  echo "Note: :5173 was occupied by another project, so GeoLibre is using :$FRONTEND_PORT instead."
fi
