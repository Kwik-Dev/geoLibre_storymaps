# CARDS — index & dependency map (user-created storymaps)

Read `delegation/HANDOUT.md` first. This is the **index**. The actual
self-contained cards live in `delegation/cards/`. **One card = one short
session** (a local Ollama run, a coding-agent session, or a human). Each card is
`closed` only when its `VERIFY` passes (paste the output into `STATUS.md`).

> The big `smithers up .smithers/workflows/storymap-build.tsx` run is the thing
> that **timed out**. Do **not** use it for this work — walk these cards one at a
> time. (Keep that workflow as the orchestration shell only.)

## How to delegate a single card

Point any OpenAI-compatible coding agent at the **local** Ollama model
(`http://127.0.0.1:11434/v1`, model `qwen3.8:27b-mlx`, key `ollama`) and give it:

    Read delegation/HANDOUT.md and delegation/cards/<file>.md card <Px.y>.
    Implement ONLY that card's SCOPE. Run its VERIFY until green.
    Update delegation/STATUS.md (mark running→closed, paste the verify output).
    Do not open a PR. Do not edit files outside the card's SCOPE.

Give the model a **generous `max_tokens` (≈ 8000–16000)** or pass
`options: { think: false }` — `qwen3.8` reasons first and will otherwise return an
empty answer (see HANDOUT §2 gotcha).

## Dependency map (DAG)

```
P1.1 → P1.2 → P1.3 ── P1.4(opt)
P1.3 → P2.1 → P2.2 → P2.3 → P2.4
P2.3 → P3.1 → P3.2 → P3.3 → P3.4
P3.1, P4.1, P4.2 → P4.3 ; P4.1 → P4.4
P5.1 (indep) → P5.2 ; P5.5 (indep) ; P3.1+P5.2 → P5.3 → P5.4
P5.4+P3.1 → P6.1 → P6.2; P4.3+P6.2 → P6.3; P6.1,P6.2,P6.3 → P6.4(gate)
M7: P4.1 → P7.1 ; P3.1 → P7.2 ; P4.4,P1.2 → P7.3 ; all → P7.4   (all OPTIONAL)
```

**Parallelizable without the backend:** `P5.1`, `P5.2`, `P5.5` (no API needed) —
start these immediately, alongside the backend spine.

**Suggested 3-worker split (partition files to avoid conflicts):**
- A — backend spine: `P1.1 P1.2 P1.3 P2.1 P2.2 P2.3 P2.4 P3.1 P3.2 P3.3 P3.4 P4.1 P4.2 P4.3 P4.4`
- B — frontend: `P5.1 P5.2 P5.5` now, then `P5.3 P5.4 P6.1 P6.2 P6.3 P6.4`
- C — optional: `P1.4 P7.1 P7.2 P7.3 P7.4` after the MVP

## Card list

| ID  | Title                                             | File                       | Deps            |      |
|-----|---------------------------------------------------|----------------------------|-----------------|------|
| P1.1| Go module init + env config                       | `cards/M1-backend.md`      | —               | core |
| P1.2| SQLite open + versioned migrations                | `cards/M1-backend.md`      | P1.1            | core |
| P1.3| chi router + `/api/health` + static serve         | `cards/M1-backend.md`      | P1.1,P1.2       | core |
| P1.4| Seed a demo user story                            | `cards/M1-backend.md`      | P1.3            | opt  |
| P2.1| GitHub OAuth2 login + upsert + 1× state           | `cards/M2-auth.md`         | P1.3            | core |
| P2.2| Admin-only local login (bcrypt, seeded)           | `cards/M2-auth.md`         | P2.1            | core |
| P2.3| JWT + httpOnly refresh + `/api` middleware         | `cards/M2-auth.md`         | P2.2            | core |
| P2.4| `whoami`                                          | `cards/M2-auth.md`         | P2.3            | core |
| P3.1| Stories CRUD + visibility/authz                   | `cards/M3-crud.md`         | P2.3            | core |
| P3.2| Chapters nested CRUD + reorder                    | `cards/M3-crud.md`         | P3.1            | core |
| P3.3| DB → camelCase story JSON adapter                 | `cards/M3-crud.md`         | P3.1            | core |
| P3.4| Export endpoint (legacy JSON)                      | `cards/M3-crud.md`         | P3.3            | core |
| P4.1| Upload endpoint (local disk)                       | `cards/M4-media.md`        | P3.1            | core |
| P4.2| External-URL validation                           | `cards/M4-media.md`        | —               | core |
| P4.3| Wire `media_ref_type` into chapters               | `cards/M4-media.md`        | P3.2,P4.1,P4.2  | core |
| P4.4| Serve + visibility gate + soft-delete             | `cards/M4-media.md`        | P4.1            | core |
| P5.1| Markdown renderer component                       | `cards/M5-frontend.md`     | deps install    | core |
| P5.2| Dual render in `ChapterCard`                      | `cards/M5-frontend.md`     | P5.1            | core |
| P5.3| Async `getStory` + data-driven picker             | `cards/M5-frontend.md`     | P5.2,P3.1       | core |
| P5.4| Hash routing + empty states                        | `cards/M5-frontend.md`     | P5.3            | core |
| P5.5| Vite dev proxy `/api` + `/media`                  | `cards/M5-frontend.md`     | —               | core |
| P6.1| StoryForm                                         | `cards/M6-builder.md`      | P5.4,P3.1       | core |
| P6.2| ChapterEditor (add/edit/reorder/preview)          | `cards/M6-builder.md`      | P6.1,P3.2,P3.3  | core |
| P6.3| MediaUpload (external URL or local file)          | `cards/M6-builder.md`      | P4.x,P6.2       | core |
| P6.4| End-to-end sign-off (M6 gate)                     | `cards/M6-builder.md`      | P6.1–P6.3       | core |
| P7.1| Upload store abstraction (S3/Drive)               | `cards/M7-hardening.md`    | P4.1            | opt  |
| P7.2| Moderation gate before `public`                  | `cards/M7-hardening.md`    | P3.1            | opt  |
| P7.3| Soft-delete purge cron                            | `cards/M7-hardening.md`    | P4.4,P1.2       | opt  |
| P7.4| Docs + security-checklist sign-off                | `cards/M7-hardening.md`    | all             | opt  |

## Definition of done (MVP)
All **core** cards `closed` and `VERIFY`-green, plus the manual E2E in `P6.4`,
plus the **no-server regression**: an embedded `*-storymap.json` story still
renders with **no server** (`file://`/static intact). M7 is out of MVP scope.
