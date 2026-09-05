# Cards — M2 Auth (GitHub OAuth + admin-only local login + sessions)

Read `delegation/HANDOUT.md §0/§4` first. Card fields as in M1. A card is
`closed` only when its `VERIFY` passes. No PR, no out-of-scope edits.

---

## P2.1 — GitHub OAuth2 login + upsert by `github_id` + 1×-use state
- **DEPS:** P1.3 · **core** · est ~30m
- **READ:** feature_request §7 (auth), §10 (CSRF); HANDOUT §4
- **SCOPE:**
   - `server/internal/auth/github.go` + `oauth.go`.
   - `GET /api/auth/github`: generate random `state` (1×-use, short TTL; store
      in a small TTL map or a `oauth_states` table), redirect to
      `https://github.com/login/oauth/authorize?client_id=<id>&redirect_uri=<cb>&scope=read:user&state=<state>`.
   - `GET /api/auth/github/callback`: **verify + consume** `state` (replay →
      reject 400), exchange `code` at
      `https://github.com/login/oauth/access_token` (client secret), fetch
      `https://api.github.com/user` (+ optional `/user/emails`), **upsert**
      `users` by `github_id` (set `github_login`; role `user`; leave
      `admin_email` null), issue a session (cookie + JWT — see P2.3), redirect
      to `FRONTEND_ORIGIN` (e.g. `http://localhost:5173`) carrying the
      user id in the hash, or set the cookie and redirect to `#/`.
   - Env: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_OAUTH_BASE`
      (default `https://github.com/login/oauth`), `GITHUB_API_BASE`
      (default `https://api.github.com`), `FRONTEND_ORIGIN`.
   - `server/.../auth/oauth_test.go` using `net/http/httptest` to fake the
      authorize/token/user servers by pointing `*_BASE` env at the test server.
- **VERIFY:**
    ```
    cd server
    GITHUB_OAUTH_BASE=http://127.0.0.1:18080 GITHUB_API_BASE=http://127.0.0.1:18081 \
      CGO_ENABLED=0 go test ./internal/auth -run TestGitHubOAuth -v
    ```
   expects: authorize → callback upserts **one** user row by `github_id` and
   issues a session; a **replayed** `state` → 400; a **forged** `state` → 400.
- **HANDOFF:** `auth.GitHubConfig`, `auth.IssueSession(user)` (cookie+JWT);
   user upsert keyed on `github_id`.
- **GOTCHAS:** CSRF = single-use short-lived `state` bound to the redirect;
   consume before use; client secret never in logs; minimal scope `read:user`.
   Do **not** implement a public sign-up route.

---

## P2.2 — Admin local login (bcrypt, env-seeded, admin-only)
- **DEPS:** P2.1 · **core** · est ~20m
- **READ:** feature_request §7 (admin), §10; HANDOUT §4
- **SCOPE:**
   - `server/internal/auth/admin.go` + startup hook in
      `server/cmd/server/main.go`.
   - On startup, if `ADMIN_EMAIL` & `ADMIN_PASSWORD` are set: `bcrypt`-hash the
      password and **idempotently** upsert a `users` row
      (`role='admin'`, matched by `admin_email`, set `password_hash`). If either
      is unset, skip (a pure GitHub-auth server is fine).
   - `POST /api/auth/admin/login` — `{"email","password"}` → verify bcrypt →
      issue admin session (token+cookie). Wrong creds → **401**.
   - `POST /api/auth/admin/refresh` — rotate the session.
   - Assert **no** public registration route exists: `/api/auth/register` and
      `/api/users` → **404**.
- **VERIFY:**
    ```
    ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
      CGO_ENABLED=0 go test ./internal/auth -run TestAdminLogin -v
    ```
   expects: the seeded admin logs in → token; wrong password → 401; with no
   admin env the route returns 503/404 without crashing; `/api/auth/register` →
   404.
- **HANDOFF:** `auth.LoginAdmin(email, password) (*User, error)`; admin user
   ensured at boot.
- **GOTCHAS:** admin row idempotent on restart; bcrypt cost ≥ 10; never log the
   password/hash; **no** public self-registration under any conditions.

---

## P2.3 — JWT + httpOnly refresh cookie + auth middleware on `/api`
- **DEPS:** P2.2 · **core** · est ~20m
- **READ:** feature_request §7 (JWT), §10 (authz); HANDOUT §4
- **SCOPE:**
   - `server/internal/auth/jwt.go` — issue/verify with
      `golang-jwt/jwt/v5` (HS256, `JWT_SECRET`), `exp` ≈ 15m for the access
      token; claims `{sub: user_id, role, iat, exp}`.
   - `server/internal/auth/middleware.go` — `RequireAuth(next)`:
        - read `Authorization: Bearer <access>` first; else a valid refresh
          cookie; verify; on fail → **401** `{"error":"unauthorized"}`; on
          success attach the user to the request context (`UserFrom(ctx)`).
   - Mount `RequireAuth` on **all** `/api/*` **except** this allowlist of
      public routes:
        - `GET /api/health`
        - `GET /api/auth/github`, `GET /api/auth/github/callback`
        - `POST /api/auth/admin/login`, `POST /api/auth/admin/refresh`
        - `GET /api/stories` (public listing, anon sees public only)
        - `GET /api/stories/:id/export` **when the target story is public**
   - Refresh token in an **httpOnly, SameSite=Strict, Secure (in prod)**
     cookie.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/auth -run TestMiddleware -v
    ```
   expects: `/api/stories` **without** a token → **401**; with a valid token →
   200/expected; expired/invalid token → 401; the public allowlist paths return
   **without** a token; a private write route without a token → 401.
- **HANDOFF:** `auth.RequireAuth(http.Handler) http.Handler`,
   `auth.UserFrom(ctx) *User`.
- **GOTCHAS:** allowlist the exact public paths; don't require a token on the
   public list; cookie `Secure` only in prod; never leak user PII in 401 bodies.

---

## P2.4 — `whoami`
- **DEPS:** P2.3 · **core** · est ~5m
- **READ:** HANDOUT §4
- **SCOPE:** `GET /api/auth/whoami` under `RequireAuth` →
   `{"id","github_login","admin_email","role","admin": role=='admin'}`.
- **VERIFY:**
    ```
    CGO_ENABLED=0 go test ./internal/auth -run TestWhoami -v
    ```
   expects: with a token → the user object incl. role; without a token → 401;
   the response never contains `password_hash`.
- **HANDOFF:** none (endpoint, used by the frontend to know the current user).
- **GOTCHAS:** redact secrets/hashes; admin flag derived from `role`.
