// smithers-source: custom
// smithers-display-name: Storymap Build (per-card pieces)
// smithers-tags: build, orchestration, user-storymap
/** @jsxImportSource smthrs */
//
// Per-card, durable, resumbable workflow for the "user-created storymaps"
// feature. One short, independently-verifiable Task node per work card, each
// with its own VERIFY self-check, plus a per-milestone human Approval gate
// ("notify me when a decision is needed"). Card content is the single source
// of truth in delegation/cards/*.md, so this file only ORCHESTRATES it.
//
//   RUN whole thing:  smithers up .smithers/workflows/storymap-pieces.tsx
//   VALIDATE only:    smithers graph --compact .smithers/workflows/storymap-pieces.tsx
//   ONE card only:    smithers oneshot "implement delegation card P1.1 in server/; run its VERIFY; update STATUS.md; no PR"
//
// The agent pool (agents.implement / agents.review) resolves to a coding-agent
// CLI (codex/opencode/claude/cursor) wired to the LOCAL Ollama model via
// .smithers/agents/*.ts + env. Without such a CLI, `smithers graph` still
// validates but agent tasks won't execute until one is installed.

import { createSmithers, Approval } from "smthrs";
import { agents } from "../agents";
import { z } from "zod/v4";
import { $ } from "bun";

type Milestone = "M1" | "M2" | "M3" | "M4" | "M5" | "M6" | "M7";
type Card = { id: string; title: string; file: string; deps: string[]; milestone: Milestone; core: boolean };

const CARDS: Card[] = [
    // M1 backend skeleton
    { id: "P1.1", title: "Go module init + env config", file: "M1-backend.md", deps: [], milestone: "M1", core: true },
    { id: "P1.2", title: "SQLite open + versioned, idempotent migrations", file: "M1-backend.md", deps: ["P1.1"], milestone: "M1", core: true },
    { id: "P1.3", title: "chi router + /api/health + static serve", file: "M1-backend.md", deps: ["P1.1", "P1.2"], milestone: "M1", core: true },
    { id: "P1.4", title: "Seed a demo user story", file: "M1-backend.md", deps: ["P1.3"], milestone: "M1", core: false },
    // M2 auth
    { id: "P2.1", title: "GitHub OAuth2 login + upsert + 1x state", file: "M2-auth.md", deps: ["P1.3"], milestone: "M2", core: true },
    { id: "P2.2", title: "Admin-only local login (bcrypt, seeded)", file: "M2-auth.md", deps: ["P2.1"], milestone: "M2", core: true },
    { id: "P2.3", title: "JWT + httpOnly refresh + /api middleware", file: "M2-auth.md", deps: ["P2.2"], milestone: "M2", core: true },
    { id: "P2.4", title: "whoami", file: "M2-auth.md", deps: ["P2.3"], milestone: "M2", core: true },
    // M3 stories + chapters CRUD
    { id: "P3.1", title: "Stories CRUD + visibility/authz", file: "M3-crud.md", deps: ["P2.3"], milestone: "M3", core: true },
    { id: "P3.2", title: "Chapters nested CRUD + reorder", file: "M3-crud.md", deps: ["P3.1"], milestone: "M3", core: true },
    { id: "P3.3", title: "DB to camelCase story JSON adapter", file: "M3-crud.md", deps: ["P3.1"], milestone: "M3", core: true },
    { id: "P3.4", title: "Export endpoint (legacy JSON)", file: "M3-crud.md", deps: ["P3.3"], milestone: "M3", core: true },
    // M4 media
    { id: "P4.1", title: "Upload endpoint (local disk)", file: "M4-media.md", deps: ["P3.1"], milestone: "M4", core: true },
     { id: "P4.2", title: "External-URL validation", file: "M4-media.md", deps: [], milestone: "M4", core: true },
     { id: "P4.3", title: "Wire media_ref_type into chapters", file: "M4-media.md", deps: ["P3.2", "P4.1", "P4.2"], milestone: "M4", core: true },
     { id: "P4.4", title: "Serve + visibility gate + soft-delete", file: "M4-media.md", deps: ["P4.1"], milestone: "M4", core: true },
     // M5 frontend cut over
    { id: "P5.1", title: "Markdown renderer component", file: "M5-frontend.md", deps: [], milestone: "M5", core: true },
     { id: "P5.2", title: "Dual render in ChapterCard", file: "M5-frontend.md", deps: ["P5.1"], milestone: "M5", core: true },
     { id: "P5.3", title: "Async getStory + data-driven picker", file: "M5-frontend.md", deps: ["P5.2", "P3.1"], milestone: "M5", core: true },
     { id: "P5.4", title: "Hash routing + empty states", file: "M5-frontend.md", deps: ["P5.3"], milestone: "M5", core: true },
     { id: "P5.5", title: "Vite dev proxy /api + /media", file: "M5-frontend.md", deps: [], milestone: "M5", core: true },
     // M6 builder UI
     { id: "P6.1", title: "StoryForm", file: "M6-builder.md", deps: ["P5.4", "P3.1"], milestone: "M6", core: true },
     { id: "P6.2", title: "ChapterEditor add/edit/reorder/preview", file: "M6-builder.md", deps: ["P6.1", "P3.2", "P3.3"], milestone: "M6", core: true },
     { id: "P6.3", title: "MediaUpload (external URL or local file)", file: "M6-builder.md", deps: ["P4.3", "P6.2"], milestone: "M6", core: true },
     { id: "P6.4", title: "End-to-end sign-off (M6 gate)", file: "M6-builder.md", deps: ["P6.1", "P6.2", "P6.3"], milestone: "M6", core: true },
     // M7 hardening (optional)
     { id: "P7.1", title: "Upload store abstraction (S3/Drive)", file: "M7-hardening.md", deps: ["P4.1"], milestone: "M7", core: false },
     { id: "P7.2", title: "Moderation gate before public", file: "M7-hardening.md", deps: ["P3.1"], milestone: "M7", core: false },
     { id: "P7.3", title: "Soft-delete purge cron", file: "M7-hardening.md", deps: ["P4.4", "P1.2"], milestone: "M7", core: false },
     { id: "P7.4", title: "Docs + security-checklist sign-off", file: "M7-hardening.md", deps: ["P3.1", "P4.1", "P7.3"], milestone: "M7", core: false },
];

const MILESTONES: Milestone[] = ["M1", "M2", "M3", "M4", "M5", "M6", "M7"];
const coreIds = (m: Milestone): string[] => CARDS.filter((c) => c.milestone === m && c.core).map((c) => c.id);
const coreM6 = coreIds("M6");

// A card's task depends on its own card deps; milestones without card-level deps
// hang off the previous milestone's gate, and the first milestone hangs off env-gate.
function cardDeps(m: Milestone, c: Card): string[] {
  if (c.deps.length) return c.deps;
  const idx = MILESTONES.indexOf(m);
  return idx === 0 ? ["env-gate"] : [`${MILESTONES[idx - 1]}-gate`];
}

// ---- output schemas ----
const inputSchema = z.object({}).strict();
const envCheckSchema = z.looseObject({
  go: z.string().default(""),
  node: z.string().default(""),
  ollama: z.boolean().default(false),
  ollamaModel: z.string().default("qwen3.8:27b-mlx"),
  agentCliInstalled: z.boolean().default(false),
  serverDir: z.boolean().default(false),
});
const cardSchema = z.object({
  cardId: z.string(),
  done: z.boolean(),
  verifyPassed: z.boolean(),
  summary: z.string(),
  statusUpdate: z.string(),
});
const shipSchema = z.object({
  coreCardIds: z.array(z.string()),
  allGreen: z.boolean(),
  summary: z.string(),
});
const approvalSchema = z.object({ approved: z.boolean() });

// Per-agent-task timeout: a long local-model generation cannot hang the run.
// On timeout the TASK fails (verify never runs -> auto autopsy); the RUN resumes
// from the last completed card (not a restart). Tune per machine.
const CARD_TIMEOUT_MS = 30 * 60_000; // 30 minutes per card generation
const { Workflow, Task, Sequence, smithers, outputs } = createSmithers({
  input: inputSchema,
  envCheck: envCheckSchema,
  card: cardSchema,
  ship: shipSchema,
  approval: approvalSchema,
});

function cardPrompt(c: Card): string {
  return [
     `Implement work card ${c.id} — "${c.title}" of the user-created storymaps feature.`,
     "",
     "Read first (single source of truth):",
     "    - delegation/HANDOUT.md (decisions, worker model, env deps)",
     `    - delegation/cards/${c.file}`,
     "    - delegation/STATUS.md",
     "",
     `Implement ONLY card ${c.id}'s SCOPE, then run its VERIFY until it passes, then` +
      ` mark ${c.id} "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp), and stop.`,
     "Do NOT open a PR. Do NOT edit files outside this card's SCOPE.",
     "If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.",
     "",
     "Model note: qwen3.8:27b-mlx is a *thinking* model — use a generous maxTokens or",
     "pass options:{think:false}; a tiny cap returns an empty answer.",
    ].join("\n");
}

export default smithers((ctx) => {
  return (
      <Workflow name="storymap-pieces">
        <Sequence name="all">
          <Task id="env-check" output={outputs.envCheck}>
            {async () => {
              const go = (await $`go version`.nothrow().quiet()).stdout?.toString().trim() || "missing";
              const node = (await $`node --version`.nothrow().quiet()).stdout?.toString().trim() || "missing";
              const ollamaHit =
                 (await $`curl -s http://127.0.0.1:11434/v1/models`.nothrow().quiet())
                  .stdout?.toString()
                  .includes("qwen3.8");
              let agentCliInstalled = false;
              for (const bin of ["pi", "codex", "opencode", "claude", "claude-code", "cursor"]) {
                const r = await $`which ${bin}`.nothrow().quiet();
                if ((r.stdout?.toString() ?? "").trim().length > 0) {
                  agentCliInstalled = true;
                  break;
                 }
               }
              const serverDir =
                 (await $`test -d server && echo yes`.nothrow().quiet()).stdout?.toString().trim() === "yes";
              return {
                go,
                node,
                ollama: ollamaHit,
                ollamaModel: "qwen3.8:27b-mlx",
                agentCliInstalled,
                serverDir,
               };
             }}
          </Task>
          <Approval
           id="env-gate"
           dependsOn={["env-check"]}
           output={outputs.approval}
           onDeny="continue"
           request={{
          title: "Environment readiness",
          summary:
           "Agent = local `pi` -> Ollama model qwen3.8:27b-mlx (installed + running at http://127.0.0.1:11434). " +
            "No coding-agent install is required. " +
            "(agent wired in .smithers/agents/pi.ts -> pool in .smithers/agents/index.ts). " +
            "Approve to proceed, deny to re-check.",
          }}
          />
          {MILESTONES.map((m) => {
          const cards = CARDS.filter((c) => c.milestone === m);
          return (
               <Sequence key={`ms-${m}`} name={`cards-${m}`}>
                {cards.map((c: Card) => (
                   <Task
                    key={c.id}
                    id={c.id}
                    agent={agents.implement}
                    dependsOn={cardDeps(m, c)}
                    output={outputs.card}
                    timeoutMs={CARD_TIMEOUT_MS}
                   >
                      {cardPrompt(c)}
                   </Task>
                ))}
                <Approval
                 key={`${m}-gate`}
                 id={`${m}-gate`}
                 dependsOn={cards.map((c: Card) => c.id)}
                 output={outputs.approval}
                 onDeny="continue"
                 request={{
                title: `Approve ${m}`,
                summary: `Approve ${m} to continue; any card with done=false needs a fix first.`,
                }}
                />
               </Sequence>
          );
         })}
          <Task id="ship-report" agent={agents.review} dependsOn={coreM6} output={outputs.ship}>
            {async () => {
            const ids = coreM6;
            let green = true;
            for (const id of ids) {
              const o = ctx.outputMaybe(outputs.card, { nodeId: id });
              if (!o || o.done === false || o.verifyPassed === false) green = false;
             }
            return {
              coreCardIds: ids,
              allGreen: green,
              summary: green
               ? "All core cards closed & VERIFY-green"
               : "Not all core cards green — see delegation/STATUS.md",
             };
           }}
          </Task>
        </Sequence>
      </Workflow>
    );
});
