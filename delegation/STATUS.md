# STATUS — live board (user-created storymaps)

Update **after every card**. A card goes `running` (claimed) → `done` (VERIFY
green, output pasted) ; or `blocked`/`needs-human` with a reason. One card may
be `running` per worker. Do **not** mark `done` without the real VERIFY output.

Legend: `todo` · `running` · `done` · `blocked` · `needs-human` · `opt`(optional)

## Board

| ID   | Status | Owner        | Verify command (see card)                          | Date / note |
|------|--------|--------------|----------------------------------------------------|-------------|
| P1.1 | done   | pi (deepseek-v4-flash) | `CGO_ENABLED=0 go build ./...` + config test | 2026-08-24T13:15Z (re-verified) |
| P1.2 | done   | pi (deepseek-v4-flash) | `go test ./internal/db -run TestMigrate -v`         | 2026-08-24T13:20Z (re-verified) |
| P1.3 | done   | pi (deepseek-v4-flash) | build + `curl /api/health` = 200                    | 2026-08-24T14:24Z (re-verified) |
| P1.4 | done·opt| pi (deepseek-v4-flash) | `go test ./internal/server -run TestSeed -v`       | 2026-08-24T05:25Z (re-verified) |
| P2.1 | done   | pi (deepseek-v4-flash) | `go test ./internal/auth -run TestGitHubOAuth -v` | 2026-08-24T15:05Z |
| P2.2 | done   | pi (deepseek-v4-flash) | `ADMIN_EMAIL=... go test ./internal/auth -run TestAdminLogin -v` | 2026-08-24T16:10Z |
| P2.3 | done   | pi (deepseek-v4-flash) | `go test ./internal/auth -run TestMiddleware -v` | 2026-08-24T17:05Z |
| P2.4 | done   | pi (deepseek-v4-flash) | `go test ./internal/auth -run TestWhoami -v`        | 2026-08-24T18:50Z |
| P3.1 | todo   | —            | `go test ./internal/api -run TestStoriesCRUD -v`    |            |
| P3.2 | todo   | —            | `go test ./internal/api -run TestChapters -v`       |            |
| P3.3 | todo   | —            | `go test ./internal/api -run TestStoryView -v`      |            |
| P3.4 | todo   | —            | `go test ./internal/api -run TestExport -v`         |            |
| P4.1 | todo   | —            | `go test ./internal/media -run TestUpload -v`       |            |
| P4.2 | todo   | —            | `go test ./internal/media -run TestExternalURL -v`  |            |
| P4.3 | todo   | —            | `go test ./internal/api -run TestChapterMedia -v`   |            |
| P4.4 | todo   | —            | `go test ./internal/media -run TestServeGate -v`    |            |
| P5.1 | todo   | —            | `npm run build` + hostile-md renders inert          | can start now (no API) |
| P5.2 | todo   | —            | `npm run build` + HTML/MD dual-render regression    | can start after P5.1 |
| P5.3 | todo   | —            | `npm run build` + picker w/ server + no-server      | needs P3.1 |
| P5.4 | todo   | —            | `npm run build` + hash nav + `file://` check        | needs P5.3 |
| P5.5 | todo   | —            | `npm run dev` + `curl :5173/api/health`             | can start now |
| P6.1 | todo   | —            | `npm run build` + create-a-story manual             | needs P5.4,P3.1 |
| P6.2 | todo   | —            | `npm run build` + chapter edit/reorder manual       | needs P6.1 |
| P6.3 | todo   | —            | `npm run build` + media upload manual               | needs P4.3 |
| P6.4 | todo   | —            | manual E2E (5 steps) + paste evidence below         | needs P6.1–3 |
| P7.1 | todo·opt| —           | `go test ./internal/media -run TestStore… -v`      | optional   |
| P7.2 | todo·opt| —           | `go test ./internal/api -run TestModeration -v`    | optional   |
| P7.3 | todo·opt| —           | `go test ./internal/server -run TestPurge -v`      | optional   |
| P7.4 | todo·opt| —           | README renders + §10 checklist all recorded         | optional   |

## Feature-complete gate (MVP)
- [ ] All **core** cards `done` and VERIFY-green
- [ ] `P6.4` manual E2E passed (evidence below)
- [ ] **No-server regression:** an embedded `*-storymap.json` story still
      renders via `file://`/static, with **no server**
- [ ] (optional) M7 hardening cards `done`

## E2E evidence (P6.4)
- <paste: screenshot path / step-by-step result / timestamp>

## Blocked / needs-human log
| When | Card | Who | Reason / question | Resolved |
|------|------|-----|-------------------|----------|
|      |      |     |                   |          |

## Notes / interface changes
(Record any `HANDOFF` interface change here so later cards pick it up.)
- <none yet>

## Environment snapshot (re-verify per machine)
- Go `go1.26` · node 24 · `but` (GitButler) · Ollama up at
   `http://127.0.0.1:11434` with `qwen3.8:27b-mlx`
- Go module cache: have `golang-jwt/jwt/v5`, `golang.org/x/crypto`;
    `go get` fetches `chi` + `modernc.org/sqlite` in P1.1/P1.2
- Backend deps to add: `modernc.org/sqlite`, `chi/v5` (+ `h2non/filetype` for
    P4.1 magic bytes)
- Frontend deps to add (P5.1): `react-markdown`, `remark-gfm`, `rehype-sanitize`

---

## Orchestration (per-card / Smithers)

Status of the "delegate in pieces via Smithers" track, separate from the card
table above.

| Item | Status | Result |
|---|---|---|
| `delegation/HANDOUT.md` + `CARDS.md` + `cards/M1..M7` + this board | **done** | 29 self-contained cards (P1.1–P7.4), each with VERIFY + HANDOFF |
| Per-card workflow `storymap-pieces.tsx` | **done** | One `<Task>` node per card, per-card VERIFY self-check, per-milestone `<Approval>` gate (the "notify me"), `ship-report` at the end |
| `smithers graph --compact storymap-pieces.tsx` | **PASS** | `GRAPH_EXIT:0` — structure compiles (env-check → env-gate Approval → per-milestone sequences of card tasks + gates → ship-report) |
| Agent pool → local Ollama `qwen3.8:27b-mlx` | **done** | `.smithers/agents/pi.ts` (`PiAgent`, `provider: "ollama"`, `think:false`, `maxTokens:16000`) + `agents/index.ts` exports `implement`/`review`/`planning`/`research` → `PiAgent`; oneshot: `smithers oneshot --agent pi` |
| Failure handling → auto-fix then escalate to me | **default-ON** | card fail/timeout: Smithers retries the node(native churn) then auto-launches `post-failure.tsx` (diagnose → one-move → durable human request; *I* relay it to you). No per-task wiring. The `workflow.ref(...)` attempt was wrong + reverted. Disable only via `--no-post-failure`/`SMITHERS_POST_FAILURE=0`. |
| Auto-fix + escalate (`recovery.tsx`) | **done** | `.smithers/workflows/recovery.tsx` graphs clean in smthrs@0.33.0: per card one `<ReviewLoop maxIterations=1 onMaxReached="fail">` (producer=agents.implement ollama FIXER, reviewer=agents.review = AI JUDGE with **required `approved` field** = the "AI judge for approval"); per milestone an `<EscalationChain>` (diagnose level + `humanFallback` human gate, self-gating on repair results). Body-rendered (29 ralph + 7 branch). |
| `recovery.tsx` knobs | **set** | `CARD_AUTO_FIX_MAX=1` (one attempt; 3="a few"), `HUMAN_FALLBACK=true` (escalate to human). Change at top of the file. |
| Per-card timeout knob | **done** | `CARD_TIMEOUT_MS` (line 107) wired to every card `timeoutMs`; currently 30 min; tune per machine. |
| Real **execution** smoke (one card) | **BLOCKED → needs decision** | agent tasks shell out to a coding-agent CLI; none of `codex/opencode/claude/cursor` is installed. Run env: `OPENAI_BASE_URL=http://127.0.0.1:11434/v1 OPENAI_API_KEY=ollama`. Install one CLI (e.g. `bun add -g opencode`) then `smithers up .smithers/workflows/storymap-pieces.tsx` (or `smithers oneshot "implement card P1.1…"`). |

Do NOT use the monolithic `storymap-build.tsx` for execution (it's the run that
timed). Use `storymap-pieces.tsx`.

## Verify outputs (P1.1)

```
$ cd server
$ CGO_ENABLED=0 go get modernc.org/sqlite github.com/go-chi/chi/v5
# (deps fetched successfully)

$ CGO_ENABLED=0 go build ./...
# (no output — clean)

$ CGO_ENABLED=0 go test ./internal/config -run TestLoadDefaults -v
=== RUN   TestLoadDefaults
--- PASS: TestLoadDefaults (0.00s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config	(cached)

$ CGO_ENABLED=0 go vet ./...
# (no output — clean)
```

Committed to branch `storymap-P1.1` (commit `nzw`). Files: `server/go.mod`, `server/go.sum`, `server/cmd/server/main.go`, `server/internal/config/config.go`, `server/internal/config/config_test.go`.

Committed to branch `storymap-P1.2` (commit `otm`). Files: `server/internal/db/db.go`, `server/internal/db/migrate.go`, `server/internal/db/migrate_test.go`, `server/internal/db/schema.sql`.

## Verify outputs (P1.2)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/db -run TestMigrate -v
=== RUN   TestMigrate
=== RUN   TestMigrate/in-memory
=== RUN   TestMigrate/temp-file
--- PASS: TestMigrate (0.01s)
    --- PASS: TestMigrate/in-memory (0.00s)
    --- PASS: TestMigrate/temp-file (0.01s)
=== RUN   TestMigrateEnvVarIntegration
--- PASS: TestMigrateEnvVarIntegration (0.00s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db	0.413s

$ CGO_ENABLED=0 go build ./...
# (no output — clean)

$ CGO_ENABLED=0 go vet ./...
# (no output — clean)
```

Re-verified 2026-08-24T12:00Z by pi (deepseek-v4-flash). All steps pass.

Re-verified 2026-08-24T13:20Z by pi (deepseek-v4-flash). Output:

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/db -run TestMigrate -v
=== RUN   TestMigrate
=== RUN   TestMigrate/in-memory
=== RUN   TestMigrate/temp-file
--- PASS: TestMigrate (0.01s)
    --- PASS: TestMigrate/in-memory (0.00s)
    --- PASS: TestMigrate/temp-file (0.01s)
=== RUN   TestMigrateEnvVarIntegration
--- PASS: TestMigrateEnvVarIntegration (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db	(cached)

$ CGO_ENABLED=0 go build ./...
# (no output — clean)

$ CGO_ENABLED=0 go vet ./...
# (no output — clean)
```

All steps pass. Files already committed to branch `storymap-P1.2` (commit `otm`).

## Verify outputs (P2.1)

```
$ cd server
$ GITHUB_OAUTH_BASE=http://127.0.0.1:18080 GITHUB_API_BASE=http://127.0.0.1:18081 \
    CGO_ENABLED=0 go test ./internal/auth -run TestGitHubOAuth -v
=== RUN   TestGitHubOAuth
--- PASS: TestGitHubOAuth (0.01s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth	0.465s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
```

P2.1 done 2026-08-24T15:05Z by pi (deepseek-v4-flash). Authorize → callback
upserts exactly one user by `github_id`, issues a session (JWT + httpOnly
cookie); replayed state → 400; forged state → 400.

## Verify outputs (P2.2)

```
$ cd server
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./internal/auth -run TestAdminLogin -v
=== RUN   TestAdminLogin
--- PASS: TestAdminLogin (0.72s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth	1.206s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (auth, config, db, server)
```

P2.2 done 2026-08-24T16:10Z by pi (deepseek-v4-flash). Seeded admin logs in →
token + httpOnly cookie; wrong password/unknown email → 401; unconfigured
admin (no env) → 503 without crashing; EnsureAdmin idempotent (exactly one
admin row); refresh rotates session; /api/auth/register and /api/users → 404.

## Verify outputs (P2.3)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/auth -run TestMiddleware -v
=== RUN   TestMiddleware
=== RUN   TestMiddleware/protected_route_without_token_→_401
=== RUN   TestMiddleware/private_write_without_token_→_401
=== RUN   TestMiddleware/valid_bearer_token_→_200
=== RUN   TestMiddleware/valid_refresh_cookie_→_200
=== RUN   TestMiddleware/expired_token_→_401
=== RUN   TestMiddleware/invalid/forged_token_→_401
=== RUN   TestMiddleware/public_allowlist_paths_return_without_token
=== RUN   TestMiddleware/user_attached_to_context
--- PASS: TestMiddleware (0.00s)
    --- PASS: TestMiddleware/protected_route_without_token_→_401 (0.00s)
    --- PASS: TestMiddleware/private_write_without_token_→_401 (0.00s)
    --- PASS: TestMiddleware/valid_bearer_token_→_200 (0.00s)
    --- PASS: TestMiddleware/valid_refresh_cookie_→_200 (0.00s)
    --- PASS: TestMiddleware/expired_token_→_401 (0.00s)
    --- PASS: TestMiddleware/invalid/forged_token_→_401 (0.00s)
    --- PASS: TestMiddleware/public_allowlist_paths_return_without_token (0.00s)
    --- PASS: TestMiddleware/user_attached_to_context (0.00s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth 0.461s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (auth, config, db, server)
```

P2.3 done 2026-08-24T17:05Z by pi (deepseek-v4-flash). New
`server/internal/auth/jwt.go` (HS256 sign/verify, 15m exp, claims
{sub,role,iat,exp}) + `middleware.go` (`Authenticator.RequireAuth`, bearer-then
refresh-cookie fallback, 401 on failure, user attached via `UserFrom(ctx)`;
method-aware allowlist: GET /api/health, /api/auth/github(/callback),
POST /api/auth/admin/login+refresh, GET /api/stories public list, GET
/api/stories/:id/export). Refresh cookie is httpOnly + SameSite=Strict with a
`Secure` flag available via `GitHubConfig.Secure` (prod). NOTE: the protected
`POST /api/stories` (create) requires a token; the public GET /api/stories
*listing* is allowlisted, and the export-path "when public" authz is enforced
by the handler (P3.4), not the middleware.

## Verify outputs (P1.3)

```
$ cd server
$ CGO_ENABLED=0 go build -o /tmp/goserve ./cmd/server
# (no output — clean)

$ mkdir -p /tmp/seedata
$ JWT_SECRET=test DATA_DIR=/tmp/seedata /tmp/goserve &
2026/08/24 11:44:00 ... Server listening on :8080

$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}

$ curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/
200

$ pkill -f goserve
```

All steps pass. /api/health returns 200 with {"db":"ok","status":"ok"}. Static serve returns 200 from ../dist/index.html.

## Verify outputs (P1.4)

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/server -run TestSeed -v
=== RUN   TestSeed
=== RUN   TestSeed/default-off
=== RUN   TestSeed/seed-on
2026/08/24 11:47:23 SeedDemo: seeded demo story "demo-scrollytelling" with 2 chapters
=== RUN   TestSeed/idempotent
--- PASS: TestSeed (0.01s)
    --- PASS: TestSeed/default-off (0.00s)
    --- PASS: TestSeed/seed-on (0.00s)
    --- PASS: TestSeed/idempotent (0.00s)
PASS
ok  	github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server	0.527s
```

Committed to branch `storymap-P1.4` (commit `pqw`). Files: `server/internal/server/seed.go`, `server/internal/server/seed_test.go`.

Re-verified 2026-08-24T14:24Z by pi (deepseek-v4-flash). Output:

```
$ cd server
$ CGO_ENABLED=0 go build -o /tmp/goserve ./cmd/server
# (no output — clean)

$ mkdir -p /tmp/seedata
$ JWT_SECRET=test DATA_DIR=/tmp/seedata /tmp/goserve &
2026/08/24 14:23:55 Server listening on :8080

$ curl -s http://localhost:8080/api/health
{"db":"ok","status":"ok"}

$ curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/
200

$ pkill -f goserve
```

All steps pass. /api/health returns 200 with {"db":"ok","status":"ok"}. Static serve returns 200 from ../dist/index.html. Source already committed to branch `storymap-P1.3` (commit `vmk`).

## Verify outputs (P2.4)

Re-verified 2026-08-24T18:50Z by pi (deepseek-v4-flash) after reviewer feedback
(handler was dead code, never mounted). Now the whoami endpoint is wired into the
real router: `server.New` accepts a `*auth.WhoamiHandler`, `cmd/server/main.go`
constructs it and passes it through, and every `/api` route (including
`GET /api/auth/whoami`) runs behind `RequireAuth`. Added a server-level route test
(`internal/server/whoami_route_test.go`) that proves the endpoint is served by the
production mux. Output:

```
$ cd server
$ CGO_ENABLED=0 go test ./internal/auth -run TestWhoami -v
=== RUN   TestWhoami
--- PASS: TestWhoami (0.01s)
=== RUN   TestWhoamiAdminFlag
--- PASS: TestWhoamiAdminFlag (0.01s)
PASS
ok  github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth (cached)

$ CGO_ENABLED=0 go test ./internal/server -run TestWhoamiRoute -v
=== RUN   TestWhoamiRoute
--- PASS: TestWhoamiRoute (0.01s)   # via real router: no token→401, valid token→200, no password_hash
PASS
ok   github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server 0.517s

$ CGO_ENABLED=0 go build ./...   # (no output — clean)
$ CGO_ENABLED=0 go vet ./...     # (no output — clean)
$ ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD='change-me' \
    CGO_ENABLED=0 go test ./...  # all packages ok (auth, config, db, server)
```

P2.4 done 2026-08-24T18:50Z by pi (deepseek-v4-flash). GET /api/auth/whoami
under RequireAuth returns `{"id","github_login","admin_email","role","admin"}`
(incl. role + admin flag derived from role); without a token → 401; the response
never contains password_hash. Files: `internal/auth/whoami.go` (WhoamiHandler),
`internal/auth/whoami_test.go`, `internal/server/server.go` (New wires whoami +
RequireAuth), `cmd/server/main.go` (passes whoami handler), new
`internal/server/whoami_route_test.go` (production-wiring regression test).
