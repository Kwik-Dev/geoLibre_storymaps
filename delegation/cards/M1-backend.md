# Cards — M1 Backend skeleton (Go module, deps, config, DB)

Each card is self-contained. Read `delegation/HANDOUT.md §0/§4` first. A card
is `closed` only when its `VERIFY` passes (paste the real output into
`delegation/STATUS.md`). Do **not** open a PR or touch files outside `SCOPE`.

Card fields: `DEPS` (cards that must be `closed` first), `READ` (what to open),
`SCOPE` (the only files to create/edit), `VERIFY` (the exact pass command),
`HANDOFF` (the interface the next card consumes — don't change it silently),
`GOTCHAS`.

---

## P1.1 — Go module init + env config
- **DEPS:** none · **core** · est ~10m
- **READ:** feature_request §7 (config), §13; HANDOUT §0/§1/§4
- **SCOPE:**
   - `server/go.mod` — module `github.com/Kwik-Dev/geoLibre_storymaps/server`,
      `go 1.26`; `go get modernc.org/sqlite github.com/go-chi/chi/v5`
      (also `aws-sdk-go-v2` later if P7.1 wants S3). Build/run with
      `CGO_ENABLED=0` everywhere (pure-Go, no cgo).
   - `server/internal/config/config.go` + `config_test.go` — `Config` struct +
      `Load() (*Config, error)` from env with defaults:
        - `DATA_DIR` default `./data`
        - `DB_PATH` default `$DATA_DIR/sqlite.db`
        - `JWT_SECRET` **required** (fatal/err if empty)
        - `ADMIN_EMAIL`, `ADMIN_PASSWORD` optional (if both set, seed admin at boot)
        - `MEDIA_DIR` default `$DATA_DIR/media`
        - `ALLOWED_MEDIA_HOSTS` optional comma list (empty ⇒ all https allowed)
        - `MEDIA_MAX_BYTES` default `25 * 1024 * 1024`
        - `APP_PORT` default `8080`
   - `server/cmd/server/main.go` — load config, log resolved paths, `exit 0`
      (routing comes in P1.3).
- **VERIFY:**
    ```
    cd server
    CGO_ENABLED=0 go get modernc.org/sqlite github.com/go-chi/chi/v5
    CGO_ENABLED=0 go build ./...
    CGO_ENABLED=0 go test ./internal/config -run TestLoadDefaults -v
    CGO_ENABLED=0 go vet ./...
    ```
   expects: deps fetched; build + vet clean; test prints resolved paths and
   asserts `JWT_SECRET` empty → error and `DATA_DIR` default → derived `DB_PATH`.
- **HANDOFF:** `config.Config` (exported) with `DBPath`, `MediaDir`,
   `MaxUploadBytes`, `AllowedMediaHosts`; `config.Load()`.
- **GOTCHAS:** keep `CGO_ENABLED=0`. The first `go get` needs network
   (`proxy.golang.org` — reachable). `ALLOWED_MEDIA_HOSTS` empty ⇒ permissive.

---

## P1.2 — SQLite open + versioned, idempotent migrations
- **DEPS:** P1.1 · **core** · est ~20m
- **READ:** feature_request §4 (tables); HANDOUT §4 schema
- **SCOPE:**
   - `server/internal/db/schema.sql` — v1 DDL, the 4 tables **exactly** from
      HANDOUT §4:
        - `users(id, github_login, github_id, admin_email, password_hash,
          role, created_at)` — `role` ∈ {user,admin}; `github_id` & `admin_email`
          nullable & **unique** (one is set); `password_hash` nullable.
        - `stories(id, slug, author_id→users.id, title, subtitle, byline,
          visibility, status, global_view, created_at, updated_at, deleted_at)`
          — `visibility` ∈ {private,public}; `status` ∈ {draft,pending,approved};
          `unique(slug)`.
        - `chapters(id, story_id→stories.id, position int, title,
          description_md, alignment, hidden, location, map_animation,
          rotate_animation, on_chapter_enter, on_chapter_exit, source,
          media_type, media_ref_type, media_external_url, media_asset_id,
          created_at, updated_at, deleted_at)` — FK on `story_id`.
        - `media_assets(id, kind, stored_path, filename, bytes, mime,
          created_at, deleted_at)` — `kind` ∈ {image,video,audio}.
      Add indexes on `slug`, `author_id`, `story_id`(chapters),
      `media_asset_id`(chapters), `deleted_at`. DDL must be idempotent
      (`IF NOT EXISTS`).
   - `server/internal/db/migrate.go` — a `schema_migrations(version int pk,
      applied_at)` table; apply embedded SQL in version order, skip if already
      applied (idempotent). `Migrate(*sql.DB) error`.
   - `server/internal/db/db.go` — `Open(cfg) (*sql.DB, error)` using the
      `modernc.org/sqlite` driver; after open run
      `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA
      foreign_keys=ON`.
   - `server/internal/db/migrate_test.go`.
- **VERIFY:**
    ```
    cd server
    CGO_ENABLED=0 go test ./internal/db -run TestMigrate -v
    CGO_ENABLED=0 go build ./...   # proves pure-Go, no cgo link
    ```
   expects: `sqlite_master` shows all 4 tables; running migrate twice is
   idempotent (no error, no duplicate version row); `foreign_keys=ON`.
- **HANDOFF:** `db.Open(cfg) (*sql.DB, error)`, `db.Migrate(*sql.DB) error`.
- **GOTCHAS:** modernc driver: pass pragmas via DSN
   (`file:<path>?_pragma=wire_format|foreign_keys(1)`-style) **or** set them
   after open; test both `:memory:` and a temp file. WAL on `:memory:` is a
   no-op — fine. Never enable cgo.

---

## P1.3 — chi router + `/api/health` + static serve `dist/`
- **DEPS:** P1.1, P1.2 · **core** · est ~20m
- **READ:** feature_request §7 (router/static), §9; HANDOUT §5
- **SCOPE:**
   - `server/internal/server/server.go` — `New(cfg, db)` builds a
      `github.com/go-chi/chi/v5` `*http.Server`/`Mux`:
        - `GET /api/health` → `200 {"status":"ok","db":"ok"}` after a
           `db.Ping()` (fail loudly if DB is down).
        - Static: serve `../dist` for non-`/api`, non-`/media` paths — real
           files first, SPA fallback to `index.html`. **Never** serve `/api`
           or `/media` statically.
        - Dev CORS (`CORS_ORIGINS`, default dev localhost) on `/api`.
   - `server/cmd/server/main.go` — wire `config → db.Open → db.Migrate →
      server.New → ListenAndServe(:8080)`, log start line.
- **VERIFY:**
    ```
    cd server
    CGO_ENABLED=0 go build -o /tmp/goserve ./cmd/server
    JWT_SECRET=test DATA_DIR=/tmp/seedata /tmp/goserve & sleep 1
    curl -s http://localhost:8080/api/health      #  200 {"status":"ok","db":"ok"}
    curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/   # 200 if dist/ exists; else documented 302/404
    pkill -f goserve
    ```
- **HANDOFF:** `server.New(cfg, db)` returns the server; `/api/health` live.
- **GOTCHAS:** health must fail if DB unreachable; static must exclude
   `/api`/`/media`; `dist/` may not exist yet (build frontend first) — document
   the empty case.

---

## P1.4 — Seed a demo user story (OPTIONAL)
- **DEPS:** P1.3 · **optional** · est ~20m
- **READ:** HANDOUT §0.4, §4
- **NOTE:** Do **NOT** seed the 4 embedded `*-storymap.json` stories into the
  DB. Embedded stories stay **static/HTML** and must work with **no server**
   (§0.4). The DB renderer is Markdown-only. Seed at most **one demo Markdown
   story** so a fresh DB isn't empty — or skip this card entirely.
- **SCOPE:** `server/internal/server/seed.go` + `server/internal/server/seed_test.go`
   — guarded by env `SEED_DEMO=1` (default off). Insert one `stories` row
   (author = seeded admin if present, else a system user) with 1–2
   `chapters` carrying `description_md`.
- **VERIFY:**
    ```
    cd server
    CGO_ENABLED=0 go test ./internal/server -run TestSeed -v
    ```
   expects: with a temp DB + `SEED_DEMO=1`, one demo story + chapters exist;
   with it off, nothing is inserted.
- **HANDOFF:** none required.
- **GOTCHAS:** keep it off by default and idempotent (match on slug).
