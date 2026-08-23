# STATUS — live board (user-created storymaps)

Update **after every card**. A card goes `running` (claimed) → `done` (VERIFY
green, output pasted) ; or `blocked`/`needs-human` with a reason. One card may
be `running` per worker. Do **not** mark `done` without the real VERIFY output.

Legend: `todo` · `running` · `done` · `blocked` · `needs-human` · `opt`(optional)

## Board

| ID   | Status | Owner        | Verify command (see card)                          | Date / note |
|------|--------|--------------|----------------------------------------------------|-------------|
| P1.1 | todo   | —            | `CGO_ENABLED=0 go build ./...` + config test        |            |
| P1.2 | todo   | —            | `go test ./internal/db -run TestMigrate -v`         |            |
| P1.3 | todo   | —            | build + `curl /api/health` = 200                    |            |
| P1.4 | todo·opt| —           | `go test ./internal/server -run TestSeed -v`       | optional   |
| P2.1 | todo   | —            | `go test ./internal/auth -run TestGitHubOAuth -v`  |            |
| P2.2 | todo   | —            | `go test ./internal/auth -run TestAdminLogin -v`   |            |
| P2.3 | todo   | —            | `go test ./internal/auth -run TestMiddleware -v`    |            |
| P2.4 | todo   | —            | `go test ./internal/auth -run TestWhoami -v`        |            |
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
