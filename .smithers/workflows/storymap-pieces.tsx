// smithers-source: custom
// smithers-display-name: Storymap Build (per-card pieces)
// smithers-tags: build, orchestration, user-storymap
/** @jsxImportSource smthrs */
//
// Per-card, durable, resumable workflow for the "user-created storymaps"
// feature, with the auto-fix + escalate layer merged in from recovery.tsx.
// Each card is a ReviewLoop (implement -> AI-judge -> ONE auto-fix -> judge),
// and each milestone ends with an EscalationChain (diagnose -> human fallback).
// Card content is the single source of truth in delegation/cards/*.md, so this
// file only ORCHESTRATES it.
//
//   RUN whole thing:  smithers up .smithers/workflows/storymap-pieces.tsx
//   VALIDATE only:    smithers graph --compact .smithers/workflows/storymap-pieces.tsx
//   ONE card only:    smithers oneshot "implement delegation card P1.1 in server/; run its VERIFY; update STATUS.md; commit via but; no push/PR"
//
// The agent pool (agents.implement / agents.review / agents.research) resolves
// to the local `pi` coding agent wired to the LOCAL Ollama model via
// .smithers/agents/*.ts + env. Without such an agent, `smithers graph` still
// validates but agent tasks won't execute until one is installed.
//
// ⚠️ KNOWN LIMITATION (read before running):
//   ReviewLoop / EscalationChain do not forward `dependsOn` (a latent gap also
//   present in recovery.tsx), so cross-card ordering is enforced by the
//   <Sequence> structure, not by per-node dependsOn.
//
//   The per-card 30-min timeout is NOT enforced here — ReviewLoop does not
//   expose `timeoutMs`. maxIterations bounds the number of attempts, but a
//   single hung generation (the local Ollama model can stall) will block the
//   run indefinitely. If a card hangs: `smithers cancel <run-id>` (or
//   `smithers retry-task`), then resume. To restore a hard timeout, swap the
//   card ReviewLoop for a raw <Loop> with `timeoutMs` on the producer Task.

import { createSmithers, Approval, ReviewLoop, EscalationChain } from "smthrs";
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
// JUDGE output (reviewer of the ReviewLoop) — MUST carry `approved`.
const judgeSchema = z.object({
  approved: z.boolean(),
  reason: z.string(),
});
// diagnose level output of the EscalationChain.
const diagSchema = z.object({
  reason: z.string(),
  recommendedMove: z.string(),
  resolved: z.boolean(),
});
// escalation tracking output.
const escalationSchema = z.object({
  level: z.string(),
  escalatedTo: z.string(),
  note: z.string(),
});
const shipSchema = z.object({
  coreCardIds: z.array(z.string()),
  allGreen: z.boolean(),
  summary: z.string(),
});
const approvalSchema = z.object({ approved: z.boolean() });

// ---- the two knobs the user controls ----
// ReviewLoop cycles: cycle 1 = implement, cycle 2 = ONE auto-fix. Set 3 for "a few".
const CARD_MAX_ITERATIONS = 2;
// true = escalate to a human Approval when a card did not recover.
const HUMAN_FALLBACK = true;

const { Workflow, Task, Sequence, smithers, outputs } = createSmithers({
  input: inputSchema,
  envCheck: envCheckSchema,
  card: cardSchema,
  judge: judgeSchema,
  diag: diagSchema,
  escalation: escalationSchema,
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
      ` mark ${c.id} "done" in delegation/STATUS.md (paste the VERIFY output + a timestamp).`,
     "Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then",
     `  but commit -b storymap-${c.id} -m "${c.id}: ${c.title}" <ids>`,
     "Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a",
     "shared live board; leave its changes uncommitted for the human to integrate.",
     `This creates a dedicated virtual branch storymap-${c.id} (one per card). COMMIT ONLY —`,
     "do NOT run `but push` and do NOT open a PR.",
     "Do NOT edit files outside this card's SCOPE.",
     "If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.",
     "",
     "Model note: you are running on the cloud model deepseek-v4-flash (ollama-cloud).",
     "Return your result as valid JSON matching the declared output schema — do not emit",
     "free-form prose or partial token fragments.",
    ].join("\n");
}

function diagnosePrompt(m: Milestone, cards: Card[]): string {
  const ids = cards.map((c) => c.id).join(", ");
  return `Diagnose milestone ${m}'s cards (${ids}). For each card, check whether its AI judge approved it — read the judge output (\`smthrs output <run> card-<id>-review\`) and delegation/STATUS.md. Return { reason, recommendedMove, resolved }. If every card is approved, set resolved=true. If any card is unapproved, set resolved=false and recommend the next move (which card, and how to fix it).`;
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
           "Agent = local `pi` -> CLOUD Ollama model deepseek-v4-flash (ollama-cloud, OLLAMA_API_KEY set). " +
            "The local qwen3.8:27b-mlx model was dropped for the implement role because it returned " +
            "garbage instead of structured JSON. " +
            "(agent wired in .smithers/agents/pi.ts -> pool in .smithers/agents/index.ts). " +
            "Approve to proceed; deny to continue anyway (advisory checkpoint).",
          }}
          />
          {MILESTONES.map((m) => {
          const cards = CARDS.filter((c) => c.milestone === m);
          return (
               <Sequence key={`ms-${m}`} name={`cards-${m}`}>
                {cards.map((c: Card) => (
                   <ReviewLoop
                    key={c.id}
                    id={`card-${c.id}`}
                    producer={agents.implement}
                    reviewer={agents.review}
                    produceOutput={outputs.card}
                    reviewOutput={outputs.judge}
                    maxIterations={CARD_MAX_ITERATIONS}
                    onMaxReached="return-last"
                   >
                      {cardPrompt(c)}
                   </ReviewLoop>
                ))}
                <EscalationChain
                 key={`esc-${m}`}
                 id={`escalate-${m}`}
                 levels={[
                {
                   agent: agents.research,
                   output: outputs.diag,
                   label: "diagnose",
                   escalateIf: (r: { resolved?: boolean } | null | undefined) => !(r?.resolved),
                },
                ]}
                 escalationOutput={outputs.escalation}
                 humanFallback={HUMAN_FALLBACK}
                 humanRequest={{
                title: `Milestone ${m}: a card did not recover`,
                summary: `An auto-fix loop for ${m} did not reach approval after ${CARD_MAX_ITERATIONS} cycle(s). Diagnose, recommend one move, then escalate to a human.`,
                }}
                >
                   {diagnosePrompt(m, cards)}
                </EscalationChain>
                <Approval
                 key={`${m}-gate`}
                 id={`${m}-gate`}
                 output={outputs.approval}
                 onDeny="continue"
                 request={{
                title: `Approve ${m}`,
                summary: `Approve ${m} to continue; deny to continue anyway (advisory). Any card with done=false needs a fix first.`,
                }}
                />
               </Sequence>
          );
         })}
          <Task id="ship-report" agent={agents.review} dependsOn={coreM6.map((id) => `card-${id}`)} output={outputs.ship}>
            {async () => {
            const ids = coreM6;
            let green = true;
            for (const id of ids) {
              const o = ctx.outputMaybe(outputs.judge, { nodeId: `card-${id}-review` });
              if (!o || o.approved !== true) green = false;
             }
            return {
              coreCardIds: ids,
              allGreen: green,
              summary: green
               ? "All core cards approved by the AI judge"
               : "Not all core cards approved — see delegation/STATUS.md",
             };
           }}
          </Task>
        </Sequence>
      </Workflow>
    );
});
