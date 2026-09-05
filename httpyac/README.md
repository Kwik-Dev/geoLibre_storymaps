# httpYac tests — GeoLibre Storymaps backend

[httpYac](https://httpyac.github.io/) request files that exercise every endpoint
of the Go backend (`server/`). Each `.http` file is self-contained and defines
its own `@host` variable.

## Prerequisites

1. **Start the backend** (from the project root):

   ```bash
   cd server
   JWT_SECRET=dev-secret \
   ADMIN_EMAIL=admin@example.com \
   ADMIN_PASSWORD=change-me \
   go run ./cmd/server
   ```

   The server listens on `http://localhost:8080` by default (`APP_PORT`).
   `ADMIN_EMAIL` / `ADMIN_PASSWORD` seed the admin-only local login user
   (idempotent). `JWT_SECRET` is required.

2. **Install httpYac** (CLI or VS Code extension):

   ```bash
   npm install -g httpyac
   ```

## Run order

The requests depend on each other via response capture, so run them in this
order (or run the whole folder as a group):

1. `health.http` — no auth.
2. `auth.http` — **admin login** (captures the JWT as `login`), then whoami/refresh.
3. `stories.http` — create/list/get/update/export/delete (uses the `login` token).
4. `chapters.http` — chapter CRUD + reorder (uses the `login` token + a story id).
5. `media.http` — upload/serve/delete media (uses the `login` token).

## Run a single file

```bash
httpyac send health.http
httpyac send auth.http
httpyac send stories.http
```

## Run everything

```bash
httpyac send --all httpyac/
```

## Notes

- The JWT is captured from the login response as
  `{{login.response.body.$.token}}` and sent as `Authorization: Bearer …`.
  Run `auth.http` (the `login` request) before the files that need auth.
- `@host`, `@admin_email`, `@admin_password` are defined at the top of each
  file — edit them to match your running server / seeded admin.
- The media upload request (`media.http`) needs a real file at the referenced
  path; drop a small image there first (or point it at any local file).
