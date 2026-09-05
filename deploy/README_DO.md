# Running WordPress + Caddy + GeoLibre Storymaps on a DigitalOcean Droplet

This guide sets up `storyboard.ink` so WordPress owns the root and GeoLibre
Storymaps runs under `/maps`, all behind a single Caddy reverse proxy that
terminates HTTPS.

```
storyboard.ink/        → WordPress (Apache on 127.0.0.1:8080)
storyboard.ink/maps/   → GeoLibre Storymaps (Go server on 127.0.0.1:8081)
```

Caddy listens on 80/443 and routes by path. The Go server strips the `/maps`
prefix itself (`BASE_PATH=/maps`), so Caddy forwards the **full** path.

## Server component diagram

```
                        Internet
                           │
                           ▼
                 ┌─────────────────────┐
                 │   Caddy (80/443)    │   TLS termination + path routing
                 │  storyboard.ink     │
                 └──────┬──────────┬───┘
                        │          │
              /maps/*   │          │   /* (everything else)
                        ▼          ▼
        ┌─────────────────────┐  ┌──────────────────────────┐
        │  GeoLibre Go server │  │  Apache (127.0.0.1:8080) │
        │  (127.0.0.1:8081)    │  │  WordPress               │
        │  BASE_PATH=/maps     │  │  /var/www/html           │
        └───────┬─────────────┘  └───────────┬──────────────┘
                │                             │
        ┌───────┴────────┐            ┌───────┴────────┐
        │  dist/index.html│            │  MariaDB       │
        │  (frontend)     │            │  (wordpress DB)│
        └───────┬─────────┘            └────────────────┘
                │
        ┌───────┴────────┐
        │  data/          │   SQLite (stories) + uploaded media
        └────────────────┘
```

- **Caddy** owns ports 80/443, terminates HTTPS, and routes by path.
- **`/maps/*`** → GeoLibre Go server on `127.0.0.1:8081`. The Go server serves
  both the API (`/api`) and the built frontend (`dist/index.html`), and stores
  stories + media in `data/`.
- **`/*`** → Apache on `127.0.0.1:8080` running WordPress, backed by MariaDB.

---

## 1. Install the stack

```bash
# Caddy (reverse proxy + HTTPS)
sudo apt install -y caddy

# WordPress stack (Apache + PHP + MariaDB)
sudo apt install -y apache2 php php-mysql libapache2-mod-php mariadb-server

# GeoLibre runtime deps (Go binary is static; nothing extra needed)
```

## 2. Move Apache to port 8080

Caddy owns 80/443, so Apache must listen on 8080 (loopback only).

```bash
sudo sed -i 's/^Listen 80$/Listen 127.0.0.1:8080/' /etc/apache2/ports.conf
sudo sed -i 's/^<VirtualHost \*:80>$/<VirtualHost 127.0.0.1:8080>/' /etc/apache2/sites-available/000-default.conf
sudo systemctl restart apache2
```

## 3. Install WordPress

```bash
# Download WordPress into the web root
sudo mkdir -p /var/www/html
cd /tmp
curl -O https://wordpress.org/latest.tar.gz
sudo tar -xzf latest.tar.gz -C /var/www/html --strip-components=1
sudo chown -R www-data:www-data /var/www/html

# Create the database
sudo mysql -e "CREATE DATABASE wordpress;"
sudo mysql -e "CREATE USER 'wpuser'@'localhost' IDENTIFIED BY 'CHANGE_ME';"
sudo mysql -e "GRANT ALL PRIVILEGES ON wordpress.* TO 'wpuser'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"
```

Then finish the WordPress install in the browser at `https://storyboard.ink/`
(once Caddy is up). If you are migrating an existing WordPress, restore your
`wp-content/` and database dump instead (see the migration notes at the bottom).

## 4. Deploy GeoLibre Storymaps

First, create the production env file on your machine (in the project root).
Copy `deploy/.env.example` → `.env.prod` and fill in real values:

```
JWT_SECRET=<openssl rand -hex 32>
ADMIN_EMAIL=you@example.com
ADMIN_PASSWORD=<a strong password>
GITHUB_CLIENT_ID=<from GitHub OAuth app>
GITHUB_CLIENT_SECRET=<from GitHub OAuth app>
FRONTEND_ORIGIN=https://storyboard.ink
BASE_PATH=/maps
DATA_DIR=/opt/storymaps/data
```

Then deploy from the project root:

```bash
./deploy/deploy.sh user@host
```

The domain and base path are read from `.env.prod` (domain from
`FRONTEND_ORIGIN`, base path from `BASE_PATH`); pass them explicitly to
override:

```bash
./deploy/deploy.sh user@host storyboard.ink /maps
```

This builds the frontend with `VITE_BASE_PATH=/maps`, builds the static Go
binary, generates the Caddyfile from `deploy/Caddyfile` (a template), uploads
everything (including `.env.prod` → `/opt/storymaps/.env`), installs the
systemd unit + Caddy config, reloads Caddy, and restarts the server.

## 5. Caddy config

`deploy/Caddyfile` is a template with `__DOMAIN__` / `__BASE_PATH__` /
`__APP_PORT__` / `__WP_PORT__` placeholders. `deploy.sh` substitutes the values
from `.env.prod` and installs the result to `/etc/caddy/Caddyfile`, then
reloads Caddy. It routes `/maps/*` to the Go server and everything else to
Apache:

```caddyfile
storyboard.ink {
    handle /maps/* {
        reverse_proxy 127.0.0.1:8081
    }
    handle {
        reverse_proxy 127.0.0.1:8080
    }
}
```

Caddy auto-provisions the Let's Encrypt cert for `storyboard.ink`. Point the
domain's A record at the droplet IP first.

## 6. GitHub OAuth app (SSO)

GeoLibre uses GitHub OAuth2 for sign-in. You create an OAuth App in GitHub,
then put its credentials in `.env.prod` (which `deploy.sh` uploads to
`/opt/storymaps/.env`).

### 6a. Create the OAuth App

1. Go to **GitHub → Settings → Developer settings → OAuth Apps**.
   (Direct link: `https://github.com/settings/developers` → **OAuth Apps**.)
2. Click **New OAuth App**.
3. Fill in:
   - **Application name** — e.g. `GeoLibre Storymaps`
   - **Homepage URL** — `https://storyboard.ink/maps/`
   - **Authorization callback URL** — `https://storyboard.ink/maps/api/auth/github/callback`
4. Click **Register application**.
5. On the app's page, copy the **Client ID** and click **Generate a new client
   secret** to get the **Client secret**.

> The callback URL **must** include `/maps` and be HTTPS. The server appends
> `BASE_PATH` to `FRONTEND_ORIGIN` when building the redirect, so the two must
> agree.

### 6b. Put the credentials in `.env.prod`

Add the credentials to `.env.prod` on your machine (the file you created in
step 4), then re-run the deploy:

```
GITHUB_CLIENT_ID=<the Client ID from step 5>
GITHUB_CLIENT_SECRET=<the Client secret from step 5>
FRONTEND_ORIGIN=https://storyboard.ink
BASE_PATH=/maps
```

```bash
./deploy/deploy.sh user@host
```

The script uploads `.env.prod` to `/opt/storymaps/.env` and restarts the
server automatically.

### 6c. Test the flow

1. Open `https://storyboard.ink/maps/`.
2. Click **Sign in with GitHub**.
3. Authorize the app. You should be redirected back to
   `https://storyboard.ink/maps/#/user/<id>` and see your GitHub username in
   the header.

If login fails, see the troubleshooting section below.

---

## Verification

```bash
# WordPress at the root
curl -I https://storyboard.ink/

# GeoLibre API under /maps
curl https://storyboard.ink/maps/api/health
# → {"db":"ok","status":"ok"}

# GeoLibre app
curl -I https://storyboard.ink/maps/
```

## Troubleshooting

- **HTTPS cert fails** → A record must point at the droplet; open ports 80/443.
- **WordPress 404s** → confirm Apache is on `127.0.0.1:8080` and Caddy proxies
  the root there.
- **GeoLibre 404s** → confirm `BASE_PATH=/maps` in `/opt/storymaps/.env` and
  that Caddy forwards the full `/maps` path (do not strip it).
- **GitHub login fails** → check, in order:
  1. The callback URL in the GitHub OAuth App is exactly
     `https://storyboard.ink/maps/api/auth/github/callback`.
  2. `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` in `/opt/storymaps/.env` match
     the app, and the server was restarted after editing.
  3. `BASE_PATH=/maps` and `FRONTEND_ORIGIN=https://storyboard.ink` are set.
  4. The server logs: `sudo journalctl -u storymap-server -e` — look for
     `token exchange failed` (bad secret) or `invalid or expired state`
     (callback URL mismatch).
- **Server not starting** → `sudo journalctl -u storymap-server -e`.

---

## Migrating an existing WordPress (18.04 → 22.04)

You can't restore an 18.04 snapshot directly onto a 22.04 droplet (the snapshot
retains the old OS image). Instead:

1. **Snapshot** the 18.04 droplet (Control Panel → Snapshots → Take a snapshot).
2. **Create a new droplet** from the Ubuntu 22.04-LTS image.
3. **Restore WordPress data**:
   - **Database** – `mysqldump` (or a plugin like UpdraftPlus) from the old
     droplet, import into the new MariaDB.
   - **Files** – `rsync`/`scp` the `wp-content` directory (or the whole
     filesystem) to the new droplet.
   - **Config** – update `wp-config.php` for the new DB credentials/paths.
4. **Test** the site thoroughly on 22.04 before decommissioning the old droplet.

For details, see DigitalOcean's [Create and Restore Droplets from Snapshots](https://docs.digitalocean.com/products/snapshots/how-to/create-and-restore-droplets/).
