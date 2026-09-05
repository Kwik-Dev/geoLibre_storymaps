#!/usr/bin/env bash
# deploy.sh — build locally and deploy the GeoLibre Storymaps app to a VPS
# (DigitalOcean Droplet or Sakura VPS) behind Caddy + systemd.
#
# Usage:
#   ./deploy/deploy.sh user@host [domain] [base_path]
#
#   user@host  SSH target, e.g. root@203.0.113.10 (or your sudo user)
#   domain     your public domain, e.g. storymaps.example.com
#              (defaults to yourdomain.com — you MUST pass it)
#   base_path  subpath the app is served under, e.g. /maps when WordPress owns
#              the root (defaults to /maps; pass "" for a root deployment)
#
# Prereqs on the VPS: Caddy installed and running, and a `storymaps` user.
#   sudo useradd -r -s /usr/sbin/nologin storymaps
#   sudo apt install caddy   (or your distro's package)
#
# After this script, create /opt/storymaps/.env (see deploy/.env.example),
# then: sudo systemctl restart storymap-server
set -euo pipefail

HOST="${1:?usage: ./deploy/deploy.sh user@host [domain]}"
DOMAIN="${2:-yourdomain.com}"
ROOT=/opt/storymaps
# Subpath the app is served under (e.g. /maps when WordPress owns the root).
# Empty = root. Must match BASE_PATH in /opt/storymaps/.env.
BASE_PATH="${3:-/maps}"

echo "==> Building frontend (dist/) with base path '$BASE_PATH' ..."
VITE_BASE_PATH="$BASE_PATH" npm run build

echo "==> Building Go server binary (static, portable) ..."
# CGO_ENABLED=0 + GOOS/GOARCH produce a fully static binary that runs on any
# Linux (incl. older glibc like Ubuntu 18.04). modernc.org/sqlite is pure Go,
# so no CGO is needed.
(cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o storymap-server ./cmd/server)

echo "==> Creating remote layout ..."
ssh "$HOST" "sudo mkdir -p $ROOT/server $ROOT/dist $ROOT/data && sudo chown -R \$(whoami) $ROOT"

echo "==> Uploading binary, frontend, service, Caddyfile ..."
scp dist/index.html "$HOST:$ROOT/dist/index.html"
scp server/storymap-server "$HOST:$ROOT/server/storymap-server"
scp deploy/storymap-server.service "$HOST:/tmp/storymap-server.service"
scp deploy/Caddyfile "$HOST:/tmp/Caddyfile"

echo "==> Installing systemd unit ..."
ssh "$HOST" "sudo cp /tmp/storymap-server.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable storymap-server"

echo "==> Installing Caddy config (domain: $DOMAIN) ..."
ssh "$HOST" "sed 's/yourdomain.com/$DOMAIN/g' /tmp/Caddyfile | sudo tee /etc/caddy/Caddyfile >/dev/null && sudo systemctl reload caddy"

echo
echo "Deploy files are in place."
echo "Next steps:"
echo "  1. Create the env file:  ssh $HOST 'sudo nano $ROOT/.env'"
echo "     (copy from deploy/.env.example; set JWT_SECRET, GITHUB_*, FRONTEND_ORIGIN=https://$DOMAIN)"
echo "  2. Start the server:      ssh $HOST 'sudo systemctl start storymap-server'"
echo "  3. Check it:              ssh $HOST 'sudo systemctl status storymap-server'"
echo "  4. GitHub OAuth callback URL must be: https://$DOMAIN/api/auth/github/callback"
