# Deploying GeoLibre Storymaps to a VPS (DigitalOcean Droplet or Sakura VPS)

This deploys the app as a single Go binary behind **Caddy** (free HTTPS) and
**systemd** (auto-restart). The Go server serves both the API and the built
frontend (`dist/index.html`).

## Coexisting with WordPress

If `storyboard.ink` already runs WordPress (Apache) at the root, GeoLibre
Storymaps is served under a subpath (`/maps`) so both apps share the domain:

```
storyboard.ink/        → WordPress (Apache on 127.0.0.1:8080)
storyboard.ink/maps/   → GeoLibre Storymaps (Go server on 127.0.0.1:8081)
```

Caddy terminates HTTPS and routes by path. The Go server strips the `/maps`
prefix itself (`BASE_PATH=/maps`), so Caddy forwards the **full** path.

## Layout on the VPS

```
/opt/storymaps/
├── server/storymap-server   # the Go binary
├── dist/index.html          # the built frontend
├── data/                    # SQLite + uploaded media (persistent)
└── .env                     # env vars (uploaded from .env.prod)
```

## 1. One-time VPS setup

```bash
# create a service user
sudo useradd -r -s /usr/sbin/nologin storymaps

# install Caddy (Debian/Ubuntu)
sudo apt install -y caddy
# or: https://caddyserver.com/docs/install
```

Point your domain's **A record** at the VPS IP before Caddy tries to get a cert.

> **Ubuntu 18.04 note:** 18.04 is end-of-life (no security updates) and its glibc
> (2.27) is too old for the **latest Caddy** (needs 2.28+). The Go binary is built
> statically so it runs fine, but Caddy may not. **Recommended: use Ubuntu 20.04
> or 22.04.** If you must stay on 18.04, install an older Caddy (2.6.x) or use
> Nginx + certbot instead.

## 2. Create the production env file (`.env.prod`)

On your machine, in the project root, create `.env.prod` (copy from
`deploy/.env.example`) and fill in real values:

```
JWT_SECRET=<openssl rand -hex 32>
ADMIN_EMAIL=you@example.com
ADMIN_PASSWORD=<a strong password>
GITHUB_CLIENT_ID=<from GitHub OAuth app>
GITHUB_CLIENT_SECRET=<from GitHub OAuth app>
FRONTEND_ORIGIN=https://yourdomain.com
BASE_PATH=/maps
```

`deploy.sh` reads this file to derive the domain (from `FRONTEND_ORIGIN`), the
base path, and the ports, and uploads it to `/opt/storymaps/.env` on the VPS
(chmod 600, owned by the `storymaps` service user). `BASE_PATH` must match the
subpath in the Caddyfile and `VITE_BASE_PATH` used at build time. When
WordPress owns the root, use `/maps`.

> `DATA_DIR` must be an absolute path (`/opt/storymaps/data`), not the dev
> default `./data` — the server runs with `WorkingDirectory=/opt/storymaps/server`.

## 3. Deploy

From the project root (on your machine):

```bash
./deploy/deploy.sh user@host
```

The domain and base path are read from `.env.prod`; pass them explicitly to
override:

```bash
./deploy/deploy.sh user@host yourdomain.com /maps
```

This builds the frontend + Go binary, generates the Caddyfile from
`deploy/Caddyfile` (a template), uploads everything (including `.env.prod` →
`/opt/storymaps/.env`), installs the systemd unit + Caddy config, reloads
Caddy, and restarts the server.

## 4. GitHub OAuth app

In your GitHub OAuth app settings, set the **Authorization callback URL** to:

```
https://yourdomain.com/maps/api/auth/github/callback
```

The base path (`/maps`) is required because the server appends `BASE_PATH` to
`FRONTEND_ORIGIN` when building the redirect. It must be HTTPS.

## Updating

Re-run `./deploy/deploy.sh user@host` — it rebuilds, re-uploads, and restarts
the server automatically. No manual `systemctl restart` needed.

## Troubleshooting

- **HTTPS cert fails** → make sure the domain's A record points at the VPS and
  port 80/443 are open in the firewall.
- **Server not starting** → check `sudo journalctl -u storymap-server -e`.
- **GitHub login fails** → confirm the callback URL is exactly
  `https://yourdomain.com/maps/api/auth/github/callback` and that `BASE_PATH`
  in `/opt/storymaps/.env` matches the subpath in the Caddyfile.
