# GitHub SSO Setup Plan — GeoLibre Storymaps

Goal: enable "Sign in with GitHub" so users can create their own storymaps in the
web UI. This is a **two-part** setup: (A) wire the OAuth routes into the server
(code change, currently missing), and (B) configure a GitHub OAuth app + env vars.

---

## Part A — Code wiring (currently missing)

The GitHub OAuth handlers already exist in `server/internal/auth/github.go` and
`oauth.go`, but they are **not mounted** in the running server. `main.go` only
creates the admin + whoami handlers, and `server.New` only mounts admin
login/refresh and whoami. `GET /api/auth/github` currently returns 404.

### A1. Build the handler

In `server/cmd/server/main.go`, after the whoami handler is created:

```go
githubHandler := auth.NewGitHubHandler(auth.GitHubConfigFromEnv(), database)
```

### A2. Mount the routes

In `server/internal/server/server.go`, inside the `/api` router's
`r.Route("/auth", ...)` block, add:

```go
ar.Get("/github", githubHandler.Authorize)
ar.Get("/github/callback", githubHandler.Callback)
```

`server.New` will need the `githubHandler` passed in (add a parameter, or
construct it inside `New` from `cfg` + `db`).

### A3. Verify

- `GET /api/auth/github` → 302 redirect to GitHub's authorize URL.
- `GET /api/auth/github/callback?code=...&state=...` → exchanges code, upserts
  the user, sets the httpOnly `refresh` cookie, redirects to `FRONTEND_ORIGIN/#/`.

---

## Part B — Create a GitHub OAuth app

1. GitHub → **Settings → Developer settings → OAuth Apps → New OAuth App**.
2. **Homepage URL:** your frontend origin, e.g. `http://localhost:5174` (dev)
   or `https://yourdomain.com` (prod).
3. **Authorization callback URL:** must equal
   `FRONTEND_ORIGIN + "/api/auth/github/callback"`, e.g.
   `http://localhost:5174/api/auth/github/callback`.
4. Copy the **Client ID** and generate a **Client Secret**.

---

## Part C — Env vars

See `.env.example` for the full list. The critical ones for SSO:

| Var | Example (dev) | Notes |
|---|---|---|
| `GITHUB_CLIENT_ID` | `Iv1.xxxx` | from the OAuth app |
| `GITHUB_CLIENT_SECRET` | `xxxx` | from the OAuth app |
| `FRONTEND_ORIGIN` | `http://localhost:5174` | **must match your real frontend origin** — default is `:5173`, but this app runs on `:5174` |

---

## Part D — Loading the `.env`

- **Frontend (Vite):** loads `.env` automatically; only `VITE_*` vars are exposed
  to the browser. The only frontend var is `VITE_API`.
- **Backend (Go):** does **not** read `.env` automatically — it reads
  `os.Getenv`. To use a single `.env` for both, source it before starting the
  server, e.g.:

  ```bash
  set -a; source .env; set +a
  cd server && go run ./cmd/server
  ```

  (Or add `set -a; source .env; set +a` to `start.sh`.)

---

## Part E — After login

The callback redirects to `FRONTEND_ORIGIN/#/` (the list page). The user then
clicks **"Create a story"** to reach `#/create` and the builder form. The
httpOnly `refresh` cookie authenticates the `POST /api/stories` call.

---

## Open items / gotchas

- **`FRONTEND_ORIGIN` default is `:5173`** — this project runs on `:5174`
  (5173 is occupied by another project). Set it explicitly.
- **No `.env` loader for the backend** — must be sourced (Part D).
- **No separate dev/prod config file** — everything is env vars.
- The admin local login (`ADMIN_EMAIL`/`ADMIN_PASSWORD`) is a separate, already
  working auth path; GitHub SSO adds the user-facing login.
