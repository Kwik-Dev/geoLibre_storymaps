## runId

graph

## frameNo

0

## xml.kind

element

## xml.tag

smithers:workflow

## xml.props

| Key  | Value           |
|------|-----------------|
| name | storymap-pieces |

## xml.children

| kind    | tag               | props           | children                                                                                                                                                        |
|---------|-------------------|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| element | smithers:sequence | [object Object] | [object Object],[object Object],[object Object],[object Object],[object Object],[object Object],[object Object],[object Object],[object Object],[object Object] |

## tasks

| nodeId                     | ordinal | iteration | outputTable     | outputTableName | outputRef       | kind    | needsApproval | waitAsync | approvalMode | skipIf | retries | retryPolicy     | timeoutMs | heartbeatTimeoutMs | continueOnFail | hijack | onHijackExit | dependsOn                               | approvalOnDeny | label                                | meta            | ralphId   | agent           | prompt                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
|----------------------------|---------|-----------|-----------------|-----------------|-----------------|---------|---------------|-----------|--------------|--------|---------|-----------------|-----------|--------------------|----------------|--------|--------------|-----------------------------------------|----------------|--------------------------------------|-----------------|-----------|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| env-check                  | 0       | 0         | [object Object] | envCheck        | [object Object] | compute | false         | false     | gate         | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         |                |                                      |                 |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| env-gate                   | 1       | 0         | [object Object] | approval        | [object Object] | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     | env-check                               | continue       | Environment readiness                | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P1.1-produce          | 2       | 0         | [object Object] | card            | [object Object] | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P1.1 | [object Object] | Implement work card P1.1 — "Go module init + env config" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M1-backend.md
    - delegation/STATUS.md

Implement ONLY card P1.1's SCOPE, then run its VERIFY until it passes, then mark P1.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P1.1 -m "P1.1: Go module init + env config" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P1.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                       |
| card-P1.2-produce          | 3       | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P1.2 |                 | Implement work card P1.2 — "SQLite open + versioned, idempotent migrations" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M1-backend.md
    - delegation/STATUS.md

Implement ONLY card P1.2's SCOPE, then run its VERIFY until it passes, then mark P1.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P1.2 -m "P1.2: SQLite open + versioned, idempotent migrations" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P1.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer. |
| card-P1.3-produce          | 4       | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P1.3 |                 | Implement work card P1.3 — "chi router + /api/health + static serve" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M1-backend.md
    - delegation/STATUS.md

Implement ONLY card P1.3's SCOPE, then run its VERIFY until it passes, then mark P1.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P1.3 -m "P1.3: chi router + /api/health + static serve" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P1.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.               |
| card-P1.4-produce          | 5       | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P1.4 |                 | Implement work card P1.4 — "Seed a demo user story" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M1-backend.md
    - delegation/STATUS.md

Implement ONLY card P1.4's SCOPE, then run its VERIFY until it passes, then mark P1.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P1.4 -m "P1.4: Seed a demo user story" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P1.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                                 |
| escalate-M1-level-0        | 6       | 0         | [object Object] | diag            | [object Object] | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M1's cards (P1.1, P1.2, P1.3, P1.4). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| escalate-M1-human-fallback | 7       | 0         | [object Object] | escalation      | [object Object] | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M1: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M1-gate                    | 8       | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M1                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P2.1-produce          | 9       | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P2.1 |                 | Implement work card P2.1 — "GitHub OAuth2 login + upsert + 1x state" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M2-auth.md
    - delegation/STATUS.md

Implement ONLY card P2.1's SCOPE, then run its VERIFY until it passes, then mark P2.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P2.1 -m "P2.1: GitHub OAuth2 login + upsert + 1x state" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P2.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                  |
| card-P2.2-produce          | 10      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P2.2 |                 | Implement work card P2.2 — "Admin-only local login (bcrypt, seeded)" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M2-auth.md
    - delegation/STATUS.md

Implement ONLY card P2.2's SCOPE, then run its VERIFY until it passes, then mark P2.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P2.2 -m "P2.2: Admin-only local login (bcrypt, seeded)" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P2.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                  |
| card-P2.3-produce          | 11      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P2.3 |                 | Implement work card P2.3 — "JWT + httpOnly refresh + /api middleware" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M2-auth.md
    - delegation/STATUS.md

Implement ONLY card P2.3's SCOPE, then run its VERIFY until it passes, then mark P2.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P2.3 -m "P2.3: JWT + httpOnly refresh + /api middleware" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P2.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                |
| card-P2.4-produce          | 12      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P2.4 |                 | Implement work card P2.4 — "whoami" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M2-auth.md
    - delegation/STATUS.md

Implement ONLY card P2.4's SCOPE, then run its VERIFY until it passes, then mark P2.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P2.4 -m "P2.4: whoami" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P2.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                                                                    |
| escalate-M2-level-0        | 13      | 0         |                 | diag            |                 | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M2's cards (P2.1, P2.2, P2.3, P2.4). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| escalate-M2-human-fallback | 14      | 0         |                 | escalation      |                 | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M2: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M2-gate                    | 15      | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M2                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P3.1-produce          | 16      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P3.1 |                 | Implement work card P3.1 — "Stories CRUD + visibility/authz" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M3-crud.md
    - delegation/STATUS.md

Implement ONLY card P3.1's SCOPE, then run its VERIFY until it passes, then mark P3.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P3.1 -m "P3.1: Stories CRUD + visibility/authz" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P3.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                  |
| card-P3.2-produce          | 17      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P3.2 |                 | Implement work card P3.2 — "Chapters nested CRUD + reorder" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M3-crud.md
    - delegation/STATUS.md

Implement ONLY card P3.2's SCOPE, then run its VERIFY until it passes, then mark P3.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P3.2 -m "P3.2: Chapters nested CRUD + reorder" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P3.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                    |
| card-P3.3-produce          | 18      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P3.3 |                 | Implement work card P3.3 — "DB to camelCase story JSON adapter" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M3-crud.md
    - delegation/STATUS.md

Implement ONLY card P3.3's SCOPE, then run its VERIFY until it passes, then mark P3.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P3.3 -m "P3.3: DB to camelCase story JSON adapter" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P3.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                            |
| card-P3.4-produce          | 19      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P3.4 |                 | Implement work card P3.4 — "Export endpoint (legacy JSON)" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M3-crud.md
    - delegation/STATUS.md

Implement ONLY card P3.4's SCOPE, then run its VERIFY until it passes, then mark P3.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P3.4 -m "P3.4: Export endpoint (legacy JSON)" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P3.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                      |
| escalate-M3-level-0        | 20      | 0         |                 | diag            |                 | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M3's cards (P3.1, P3.2, P3.3, P3.4). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| escalate-M3-human-fallback | 21      | 0         |                 | escalation      |                 | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M3: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M3-gate                    | 22      | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M3                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P4.1-produce          | 23      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P4.1 |                 | Implement work card P4.1 — "Upload endpoint (local disk)" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M4-media.md
    - delegation/STATUS.md

Implement ONLY card P4.1's SCOPE, then run its VERIFY until it passes, then mark P4.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P4.1 -m "P4.1: Upload endpoint (local disk)" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P4.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                       |
| card-P4.2-produce          | 24      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P4.2 |                 | Implement work card P4.2 — "External-URL validation" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M4-media.md
    - delegation/STATUS.md

Implement ONLY card P4.2's SCOPE, then run its VERIFY until it passes, then mark P4.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P4.2 -m "P4.2: External-URL validation" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P4.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                                 |
| card-P4.3-produce          | 25      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P4.3 |                 | Implement work card P4.3 — "Wire media_ref_type into chapters" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M4-media.md
    - delegation/STATUS.md

Implement ONLY card P4.3's SCOPE, then run its VERIFY until it passes, then mark P4.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P4.3 -m "P4.3: Wire media_ref_type into chapters" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P4.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                             |
| card-P4.4-produce          | 26      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P4.4 |                 | Implement work card P4.4 — "Serve + visibility gate + soft-delete" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M4-media.md
    - delegation/STATUS.md

Implement ONLY card P4.4's SCOPE, then run its VERIFY until it passes, then mark P4.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P4.4 -m "P4.4: Serve + visibility gate + soft-delete" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P4.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                     |
| escalate-M4-level-0        | 27      | 0         |                 | diag            |                 | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M4's cards (P4.1, P4.2, P4.3, P4.4). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| escalate-M4-human-fallback | 28      | 0         |                 | escalation      |                 | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M4: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M4-gate                    | 29      | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M4                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P5.1-produce          | 30      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P5.1 |                 | Implement work card P5.1 — "Markdown renderer component" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M5-frontend.md
    - delegation/STATUS.md

Implement ONLY card P5.1's SCOPE, then run its VERIFY until it passes, then mark P5.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P5.1 -m "P5.1: Markdown renderer component" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P5.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                      |
| card-P5.2-produce          | 31      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P5.2 |                 | Implement work card P5.2 — "Dual render in ChapterCard" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M5-frontend.md
    - delegation/STATUS.md

Implement ONLY card P5.2's SCOPE, then run its VERIFY until it passes, then mark P5.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P5.2 -m "P5.2: Dual render in ChapterCard" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P5.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                        |
| card-P5.3-produce          | 32      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P5.3 |                 | Implement work card P5.3 — "Async getStory + data-driven picker" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M5-frontend.md
    - delegation/STATUS.md

Implement ONLY card P5.3's SCOPE, then run its VERIFY until it passes, then mark P5.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P5.3 -m "P5.3: Async getStory + data-driven picker" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P5.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                      |
| card-P5.4-produce          | 33      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P5.4 |                 | Implement work card P5.4 — "Hash routing + empty states" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M5-frontend.md
    - delegation/STATUS.md

Implement ONLY card P5.4's SCOPE, then run its VERIFY until it passes, then mark P5.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P5.4 -m "P5.4: Hash routing + empty states" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P5.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                      |
| card-P5.5-produce          | 34      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P5.5 |                 | Implement work card P5.5 — "Vite dev proxy /api + /media" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M5-frontend.md
    - delegation/STATUS.md

Implement ONLY card P5.5's SCOPE, then run its VERIFY until it passes, then mark P5.5 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P5.5 -m "P5.5: Vite dev proxy /api + /media" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P5.5 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                    |
| escalate-M5-level-0        | 35      | 0         |                 | diag            |                 | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M5's cards (P5.1, P5.2, P5.3, P5.4, P5.5). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| escalate-M5-human-fallback | 36      | 0         |                 | escalation      |                 | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M5: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M5-gate                    | 37      | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M5                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P6.1-produce          | 38      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P6.1 |                 | Implement work card P6.1 — "StoryForm" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M6-builder.md
    - delegation/STATUS.md

Implement ONLY card P6.1's SCOPE, then run its VERIFY until it passes, then mark P6.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P6.1 -m "P6.1: StoryForm" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P6.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                                                           |
| card-P6.2-produce          | 39      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P6.2 |                 | Implement work card P6.2 — "ChapterEditor add/edit/reorder/preview" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M6-builder.md
    - delegation/STATUS.md

Implement ONLY card P6.2's SCOPE, then run its VERIFY until it passes, then mark P6.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P6.2 -m "P6.2: ChapterEditor add/edit/reorder/preview" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P6.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                 |
| card-P6.3-produce          | 40      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P6.3 |                 | Implement work card P6.3 — "MediaUpload (external URL or local file)" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M6-builder.md
    - delegation/STATUS.md

Implement ONLY card P6.3's SCOPE, then run its VERIFY until it passes, then mark P6.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P6.3 -m "P6.3: MediaUpload (external URL or local file)" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P6.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.             |
| card-P6.4-produce          | 41      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P6.4 |                 | Implement work card P6.4 — "End-to-end sign-off (M6 gate)" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M6-builder.md
    - delegation/STATUS.md

Implement ONLY card P6.4's SCOPE, then run its VERIFY until it passes, then mark P6.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P6.4 -m "P6.4: End-to-end sign-off (M6 gate)" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P6.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                   |
| escalate-M6-level-0        | 42      | 0         |                 | diag            |                 | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M6's cards (P6.1, P6.2, P6.3, P6.4). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| escalate-M6-human-fallback | 43      | 0         |                 | escalation      |                 | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M6: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M6-gate                    | 44      | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M6                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| card-P7.1-produce          | 45      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P7.1 |                 | Implement work card P7.1 — "Upload store abstraction (S3/Drive)" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M7-hardening.md
    - delegation/STATUS.md

Implement ONLY card P7.1's SCOPE, then run its VERIFY until it passes, then mark P7.1 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P7.1 -m "P7.1: Upload store abstraction (S3/Drive)" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P7.1 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                     |
| card-P7.2-produce          | 46      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P7.2 |                 | Implement work card P7.2 — "Moderation gate before public" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M7-hardening.md
    - delegation/STATUS.md

Implement ONLY card P7.2's SCOPE, then run its VERIFY until it passes, then mark P7.2 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P7.2 -m "P7.2: Moderation gate before public" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P7.2 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                 |
| card-P7.3-produce          | 47      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P7.3 |                 | Implement work card P7.3 — "Soft-delete purge cron" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M7-hardening.md
    - delegation/STATUS.md

Implement ONLY card P7.3's SCOPE, then run its VERIFY until it passes, then mark P7.3 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P7.3 -m "P7.3: Soft-delete purge cron" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P7.3 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                                               |
| card-P7.4-produce          | 48      | 0         |                 | card            |                 | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     |                                         |                |                                      |                 | card-P7.4 |                 | Implement work card P7.4 — "Docs + security-checklist sign-off" of the user-created storymaps feature.

Read first (single source of truth):
    - delegation/HANDOUT.md (decisions, worker model, env deps)
    - delegation/cards/M7-hardening.md
    - delegation/STATUS.md

Implement ONLY card P7.4's SCOPE, then run its VERIFY until it passes, then mark P7.4 "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).
Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then
  but commit -b storymap-P7.4 -m "P7.4: Docs + security-checklist sign-off" <ids>
Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a
shared live board; leave its changes uncommitted for the human to integrate.
This creates a dedicated virtual branch storymap-P7.4 (one per card). COMMIT ONLY —
do NOT run `but push` and do NOT open a PR.
Do NOT edit files outside this card's SCOPE.
If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.

Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or
pass options:{think:false}; a tiny cap returns an empty answer.                       |
| escalate-M7-level-0        | 49      | 0         |                 | diag            |                 | agent   | false         | false     | gate         | false  | 1       | [object Object] |           | 600000             | true           | false  | complete     |                                         |                | diagnose                             |                 |           |                 | Diagnose milestone M7's cards (P7.1, P7.2, P7.3, P7.4). For each card, check whether its AI judge approved it — read the judge output (`smthrs output <run> card-<id>-review`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| escalate-M7-human-fallback | 50      | 0         |                 | escalation      |                 | compute | true          | false     | decision     | false  | 0       |                 |           |                    | true           | false  | complete     |                                         | continue       | Milestone M7: a card did not recover | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| M7-gate                    | 51      | 0         |                 | approval        |                 | compute | true          | false     | decision     | false  |         | [object Object] |           |                    | false          | false  | complete     |                                         | continue       | Approve M7                           | [object Object] |           |                 |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| ship-report                | 52      | 0         | [object Object] | ship            | [object Object] | agent   | false         | false     | gate         | false  |         | [object Object] |           | 600000             | false          | false  | complete     | card-P6.1,card-P6.2,card-P6.3,card-P6.4 |                |                                      |                 |           |                 | [object Promise]                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |

## cta.description

Suggest to the user:
1. Build the user a custom workflow UI: author .smithers/ui/storymap-pieces.tsx by composing the smthrs/gateway-ui widgets (WorkflowUiShell, RunTree, RunEventLog, ApprovalPanel, ...) and smthrs/ui primitives over the gateway-react hooks, add `<UI entry="../ui/storymap-pieces.tsx" />` to the workflow, then open it with `smithers ui graph`. `smithers ui --app` opens the full Smithers control-plane UI.
2. Visualize the workflow and how it works: `smithers tree graph` prints the live run tree, `smithers up --interactive` opens the full-screen TUI.
3. Ask the user clarifying questions about what they want next, then route it by size: a clear single-agent task runs as `smithers oneshot "<goal>"` (background, durable, no workflow file); genuinely multi-stage / approval-gated / reusable work goes through `smithers workflow run create-workflow --prompt "<what the workflow should do>"` (or `smithers make-workflow`).

## cta.commands

| command                                                                  | description                                                             |
|--------------------------------------------------------------------------|-------------------------------------------------------------------------|
| smithers ui graph                                                        | Open the workflow UI (after authoring .smithers/ui/storymap-pieces.tsx) |
| smithers ui --app                                                        | Open the full Smithers control-plane UI                                 |
| smithers tree graph                                                      | Visualize the run tree                                                  |
| smithers oneshot "<goal>"                                                | Run a clear single-agent task in the background, no workflow file       |
| smithers workflow run create-workflow --prompt "<describe the workflow>" | Have smithers build a new workflow from a description                   |
