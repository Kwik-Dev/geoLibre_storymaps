# Handout — Smithers workflow operations (and the smithers.db protection feature)

This is the **operations handout** for running the `storymap-pieces` Smithers
workflow in this repo. It exists because on **2026-08-24** the workflow's
`smithers.db` was deleted *while a run was active*, which killed the run and
wiped its durable orchestration state. This doc records (1) how to operate the
workflow, (2) what happened, and (3) a **feature request to protect
`smithers.db` from deletion while a workflow is running**.

---

## 1. The workflow (`storymap-pieces`)

- **File:** `.smithers/workflows/storymap-pieces.tsx`
- **What it does:** builds the user-created storymaps feature one card at a
  time. Each card is a `ReviewLoop` (implement → AI-judge → one auto-fix →
  re-judge); each milestone ends with an `EscalationChain` + a human
  `Approval` gate. Card content lives in `delegation/cards/*.md`; the workflow
  only orchestrates.
- **Agent pool:** `agents.implement` / `agents.review` / `agents.research` all
  resolve to `pi_ollama_cloud` (`SmithersPiAgent` → `ollama-cloud` /
  `deepseek-v4-flash:cloud`). See `.smithers/agents.ts`.

### Run / monitor / resume

```bash
# start (detached)
smithers workflow run storymap-pieces --detach

# watch
smithers ps
smithers monitor <run-id>          # live web UI
smithers logs <run-id> -f          # event stream

# clear an approval gate, then resume
smithers approve <run-id> --node <gate> --by <name>
smithers up .smithers/workflows/storymap-pieces.tsx --resume --run-id <run-id> --detach
```

### Gates (advisory — `onDeny="continue"`)

`env-gate` → `M1-gate` … `M7-gate`. Each pauses the run until approved/denied.
The run does **not** auto-advance past a gate; a human (or the orchestrating
agent) must `approve`/`deny` and then `--resume`.

### Known gotchas

- **Structured output:** the `pi` agent reports `tokenUsage` as all zeros to
  smithers (no per-node token accounting). Use ollama-cloud request counts
  (`curl -s -H "Authorization: Bearer $OLLAMA_API_KEY" https://ollama.com/api/usage`)
  as a proxy.
- **Node ids with dots** (`card-P1.1-produce`) fail `smithers output` /
  `smithers node` validation (`InvalidNodeId` — the regex rejects `.`). Query
  via `smithers inspect --format json` instead.
- **Watcher subagents hang** on `smithers logs … | tail` (tail-on-a-stream) and
  on `for` loops over `smithers node`. Poll with a single `smithers ps` only.

---

## 2. The incident (2026-08-24)

**What happened:** while `run-1787548805388` was on `card-P7.3-produce`, the
run vanished from `smithers ps`. Investigation found:

- `smithers.db` (project root) was **deleted** — not in Trash, no backup.
- The gateway was **down** (`smithers gateway status` → `running: false`).
- `.smithers/node_modules` was **partially removed** (`zod` and `incur` packages
  missing), which matches the transient `Cannot find package 'zod'` error seen
  right before the run died.

**Impact:** the run's durable orchestration state (which nodes finished, gate
decisions) was lost. The **code was safe** — every card's work was already
committed to its `storymap-<card-id>` GitButler branch, so only the
*orchestration* state was lost, not the implementation.

**Root cause (probable):** something rewrote/cleaned `.smithers/` (all files
show the same `19:35` mtime) and removed `smithers.db` + parts of
`node_modules` while the gateway was live. The gateway then crashed on the
missing `zod` import, and the db was gone.

---

## 3. Feature request — protect `smithers.db` from deletion while a workflow is running

> **Problem:** `smithers.db` is the single source of truth for run state. It is
> a plain file in the project root, gitignored, with no lock, no backup, and no
> deletion guard. Deleting it (or letting a cleanup/`bun install`/`smithers
> init` touch it) while a run is active silently kills the run and loses all
> checkpoint state.

**Proposed feature (for the Smithers tool itself):**

1. **Active-run deletion guard.** While any run is `running`/`waiting-*`, the
   gateway holds an exclusive lock (or a sentinel) on `smithers.db` and
   **refuses** to delete/overwrite it. A `smithers` command that would remove
   or reset the db must fail with a clear error ("N active runs — cannot delete
   smithers.db") unless forced with an explicit `--force` + confirmation.

2. **Periodic snapshot / backup.** The gateway writes a rolling backup
   (`smithers.db.bak` or a timestamped copy) on a short interval (e.g. every
   5 min) and on every node completion. On startup, if `smithers.db` is missing
   but a `.bak` exists, offer to restore it.

3. **Recovery from git.** Because each card commits to a `storymap-<id>`
   branch, the gateway (or a `smithers recover` command) can reconstruct
   "which cards are done" from the branch list + `delegation/STATUS.md`, so a
   lost db degrades to "re-verify done cards" rather than "start over".

4. **Crash-safe teardown.** The gateway should not delete `smithers.db` on
   shutdown, and should detect a missing `zod`/`incur` (broken `node_modules`)
   *before* touching the db, surfacing "reinstall `.smithers` deps" instead of
   silently dying.

**Acceptance criteria:**

- Deleting `smithers.db` while a run is active is **impossible** without an
  explicit force flag, and the active run keeps running.
- A killed gateway leaves `smithers.db` (or a restorable `.bak`) intact.
- `smithers recover` lists which cards are already committed and offers to
  resume from the last completed node.

---

## 4. Recovery playbook (if it happens again)

1. **Do not panic — the code is safe.** Every card's work is in a
   `storymap-<card-id>` GitButler branch. `but status` lists them.
2. **Check what's actually lost.** Only the orchestration state (node
   completion, gate decisions) is gone; the implementation is intact.
3. **Reinstall `.smithers` deps** if `node_modules` is broken:
   `cd .smithers && bun install`.
4. **Re-run the workflow** (`smithers workflow run storymap-pieces --detach`).
   Done cards re-verify quickly (idempotent); the run advances to the first
   unfinished card.
5. **Or finish the remaining cards directly** (bounded work) and record
   evidence in `delegation/STATUS.md`.
