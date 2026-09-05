#!/usr/bin/env bash
# deploy.sh — build locally and deploy the GeoLibre Storymaps app to a VPS
# (DigitalOcean Droplet or Sakura VPS) behind Caddy + systemd.
#
# Reads .env.prod (the production config) for the domain, base path, ports, and
# secrets. CLI args override the values derived from .env.prod.
#
# Usage:
#   ./deploy/deploy.sh user@host [domain] [base_path]
#
#   user@host  SSH target, e.g. root@203.0.113.10 (or your sudo user)
#   domain     override the domain (default: DOMAIN or FRONTEND_ORIGIN in .env.prod)
#   base_path  override the subpath (default: BASE_PATH / VITE_BASE_PATH in .env.prod)
#
# Prereqs on the VPS: Caddy installed and running, and a `storymaps` user.
#   sudo useradd -r -s /usr/sbin/nologin storymaps
#   sudo apt install caddy   (or your distro's package)
#
# The script uploads .env.prod to /opt/storymaps/.env (chmod 600, owned by the
# storymaps service user), installs the systemd unit + Caddyfile, and restarts
# the server so the new binary + env take effect.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT/.env.prod"
REMOTE_ROOT=/opt/storymaps

HOST="${1:?usage: ./deploy/deploy.sh user@host [domain] [base_path]}"

# ── Load production config from .env.prod ────────────────────────────────
if [ -f "$ENV_FILE" ]; then
    echo "==> Loading production config from .env.prod ..."
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
else
    echo "WARNING: $ENV_FILE not found — using CLI args / defaults only."
fi

# ── Derive the domain ─────────────────────────────────────────────────────
# Precedence: CLI arg > $DOMAIN > $FRONTEND_ORIGIN (scheme/path stripped) > default.
derive_domain() {
    local origin="${FRONTEND_ORIGIN:-}"
    if [ -n "$origin" ]; then
        printf '%s' "$origin" | sed -E 's#^https?://##; s#/.*$##'
    fi
}
DOMAIN="${2:-${DOMAIN:-$(derive_domain)}}"
DOMAIN="${DOMAIN:-yourdomain.com}"

# ── Derive the base path ──────────────────────────────────────────────────
# Precedence: CLI arg > $BASE_PATH > $VITE_BASE_PATH > /maps.
BASE_PATH="${3:-${BASE_PATH:-${VITE_BASE_PATH:-/maps}}}"
# Normalize: leading slash, no trailing slash; "/" → "" (root deployment).
BASE_PATH="$(printf '%s' "$BASE_PATH" | sed -E 's#^/*#/#; s#/+$##')"
[ "$BASE_PATH" = "/" ] && BASE_PATH=""

# ── Derive ports ──────────────────────────────────────────────────────────
APP_PORT="${APP_PORT:-8081}"   # the Go server (Caddy proxies to this)
WP_PORT="${WP_PORT:-8080}"     # WordPress/Apache at the root

# Guard: the Go server and WordPress must not share a port. A collision here
# (e.g. APP_PORT=8080 in .env.prod) makes Caddy proxy /maps/* to Apache and
# the Go server crash-loop on a bind failure.
if [ "$APP_PORT" = "$WP_PORT" ]; then
    echo "ERROR: APP_PORT ($APP_PORT) and WP_PORT ($WP_PORT) must differ." >&2
    echo "       The Go server and WordPress/Apache cannot bind the same port." >&2
    echo "       Set APP_PORT=8081 in .env.prod (WordPress uses 8080)." >&2
    exit 1
fi

echo "==> Deploying to $HOST"
echo "    domain=$DOMAIN  base_path=${BASE_PATH:-/}  app_port=$APP_PORT  wp_port=$WP_PORT"

# ── Build frontend ────────────────────────────────────────────────────────
echo "==> Building frontend (dist/) with base path '${BASE_PATH:-/}' ..."
VITE_BASE_PATH="$BASE_PATH" npm run build

# ── Build Go server binary ────────────────────────────────────────────────
echo "==> Building Go server binary (static, portable) ..."
# CGO_ENABLED=0 + GOOS/GOARCH produce a fully static binary that runs on any
# Linux (incl. older glibc like Ubuntu 18.04). modernc.org/sqlite is pure Go,
# so no CGO is needed.
(cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o storymap-server ./cmd/server)

# ── Generate the Caddyfile from the template ──────────────────────────────
echo "==> Generating Caddyfile (domain=$DOMAIN, base_path=${BASE_PATH:-/}) ..."
CADDY_TMP="$(mktemp)"
sed -e "s#__DOMAIN__#$DOMAIN#g" \
    -e "s#__BASE_PATH__#${BASE_PATH:-/}#g" \
    -e "s#__APP_PORT__#$APP_PORT#g" \
    -e "s#__WP_PORT__#$WP_PORT#g" \
    "$ROOT/deploy/Caddyfile" > "$CADDY_TMP"

# ── Upload to /tmp, then sudo-install into place with correct ownership ───
echo "==> Uploading artifacts ..."
scp dist/index.html "$HOST:/tmp/storymaps-index.html"
scp server/storymap-server "$HOST:/tmp/storymap-server"
scp deploy/storymap-server.service "$HOST:/tmp/storymap-server.service"
scp "$CADDY_TMP" "$HOST:/tmp/Caddyfile"
rm -f "$CADDY_TMP"
if [ -f "$ENV_FILE" ]; then
    scp "$ENV_FILE" "$HOST:/tmp/storymaps.env"
fi

echo "==> Installing on the VPS (systemd + Caddy + env) ..."
ssh "$HOST" "
    set -e
    sudo mkdir -p $REMOTE_ROOT/server $REMOTE_ROOT/dist $REMOTE_ROOT/data
    sudo cp /tmp/storymaps-index.html $REMOTE_ROOT/dist/index.html
    # Replace the binary atomically: cp over a *running* executable fails with
    # "Text file busy" (ETXTBSY), so copy to a temp name in the same dir and
    # mv (rename) over it — the running process keeps its old inode, and the
    # new binary is in place for the restart below.
    sudo cp /tmp/storymap-server $REMOTE_ROOT/server/storymap-server.new
    sudo chmod +x $REMOTE_ROOT/server/storymap-server.new
    sudo mv -f $REMOTE_ROOT/server/storymap-server.new $REMOTE_ROOT/server/storymap-server
    sudo chown -R storymaps:storymaps $REMOTE_ROOT
    if [ -f /tmp/storymaps.env ]; then
        sudo cp /tmp/storymaps.env $REMOTE_ROOT/.env
        sudo chown storymaps:storymaps $REMOTE_ROOT/.env
        sudo chmod 600 $REMOTE_ROOT/.env
    fi
    sudo cp /tmp/storymap-server.service /etc/systemd/system/
    sudo cp /tmp/Caddyfile /etc/caddy/Caddyfile
    sudo systemctl daemon-reload
    sudo systemctl enable storymap-server
    sudo systemctl reload caddy
    sudo systemctl restart storymap-server
    rm -f /tmp/storymaps-index.html /tmp/storymap-server /tmp/storymap-server.service /tmp/Caddyfile /tmp/storymaps.env
"

echo
echo "Deploy complete."
echo "  App:      https://$DOMAIN${BASE_PATH:+$BASE_PATH}/"
echo "  Health:   https://$DOMAIN${BASE_PATH:+$BASE_PATH}/api/health"
echo "  GitHub OAuth callback: https://$DOMAIN${BASE_PATH:+$BASE_PATH}/api/auth/github/callback"
echo
echo "Verify:"
echo "  ssh $HOST 'sudo systemctl status storymap-server'"
echo "  curl https://$DOMAIN${BASE_PATH:+$BASE_PATH}/api/health"
