// smithers-source: custom
// smithers-display-name: Storymap Build (auto-fix + escalate)
// smithers-tags: build, orchestration, user-storymap, recovery
/** @jsxImportSource smthrs */
//
// COMPANION to storymap-pieces.tsx: the user's requested "one auto-fix, then
// escalate to a human" layer. Unlike post-failure.tsx (diagnose-only), this
// actually REPAIRS, then escalates.
//
// Proven shape (smoke-validated in .smithers/workflows/smoke-*.tsx):
//   - A bare <ReviewLoop maxIterations=1 onMaxReached="fail"> renders as
//     workflow -> smithers:ralph -> sequence -> task.  (ReviewLoop IS Ralph.)
//   - A bare <EscalationChain humanFallback> renders as sequence -> task + branch.
//   - A <Task> nested as the single child of <TryCatchFinally> does NOT render its
//     body in the static graph (the try branch collapses), so we do NOT nest
//     composite components inside a Tcf. Instead each component is a first-class
//     node in the milestone <Sequence>.
//
// Per card:  <ReviewLoop> = the ONE auto-fix, AI-judged.
//   producer (agents.implement, ollama PiAgent) = the FIXER: re-read the card +
//     VERIFY + the failure, make ONE scoped repair, re-run VERIFY; no commit/PR;
//     never touch files outside the card's SCOPE.
//   reviewer (agents.review, ollama PiAgent)     = the AI JUDGE (the "AI judge for
//     approval"); its output schema REQUIRES `approved: boolean`. maxIterations=1
//     => exactly one auto-fix attempt; onMaxReached="fail" => fails if not approved.
//
// Per milestone:  <EscalationChain> = the ESCALATE-TO-HUMAN side.
//   A `diagnose` level (agents.research -> {reason, recommendedMove, resolved});
//   a human fallback Approval when unresolved. The diagnose level reads the
//   repair-loop results, so it self-gates: if no card failed, it resolves without
//   escalating.
//
// Two knobs: CARD_AUTO_FIX_MAX, HUMAN_FALLBACK. DATA = delegation/cards/*.

import { createSmithers, Approval, ReviewLoop, EscalationChain } from "smthrs";
import { agents } from "../agents";
import { z } from "zod/v4";

type Milestone = "M1" | "M2" | "M3" | "M4" | "M5" | "M6" | "M7";
type Card = { id: string; title: string; file: string; deps: string[]; milestone: Milestone; core: boolean };

const CARDS: Card[] = [
 { id: "P1.1", title: "Go module init + env config", file: "M1-backend.md", deps: [], milestone: "M1", core: true },
 { id: "P1.2", title: "SQLite open + versioned idempotent migrations", file: "M1-backend.md", deps: ["P1.1"], milestone: "M1", core: true },
 { id: "P1.3", title: "chi router + /api/health + static serve", file: "M1-backend.md", deps: ["P1.1", "P1.2"], milestone: "M1", core: true },
 { id: "P1.4", title: "Seed a demo user story", file: "M1-backend.md", deps: ["P1.3"], milestone: "M1", core: false },
 { id: "P2.1", title: "GitHub OAuth2 login + upsert + 1x state", file: "M2-auth.md", deps: ["P1.3"], milestone: "M2", core: true },
 { id: "P2.2", title: "Admin-only local login (bcrypt, seeded)", file: "M2-auth.md", deps: ["P2.1"], milestone: "M2", core: true },
 { id: "P2.3", title: "JWT + httpOnly refresh + /api middleware", file: "M2-auth.md", deps: ["P2.2"], milestone: "M2", core: true },
 { id: "P2.4", title: "whoami", file: "M2-auth.md", deps: ["P2.3"], milestone: "M2", core: false },
 { id: "P3.1", title: "Stories CRUD + visibility/authz", file: "M3-crud.md", deps: ["P2.3"], milestone: "M3", core: true },
 { id: "P3.2", title: "Chapters nested CRUD + reorder", file: "M3-crud.md", deps: ["P3.1"], milestone: "M3", core: true },
 { id: "P3.3", title: "DB to camelCase story JSON adapter", file: "M3-crud.md", deps: ["P3.1"], milestone: "M3", core: true },
 { id: "P3.4", title: "Export endpoint (legacy JSON)", file: "M3-crud.md", deps: ["P3.3"], milestone: "M3", core: false },
 { id: "P4.1", title: "Upload endpoint (local disk)", file: "M4-media.md", deps: ["P3.1"], milestone: "M4", core: true },
 { id: "P4.2", title: "External-URL validation", file: "M4-media.md", deps: [], milestone: "M4", core: true },
 { id: "P4.3", title: "Wire media_ref_type into chapters", file: "M4-media.md", deps: ["P3.2", "P4.1", "P4.2"], milestone: "M4", core: true },
 { id: "P4.4", title: "Serve + visibility gate + soft-delete", file: "M4-media.md", deps: ["P4.1"], milestone: "M4", core: false },
 { id: "P5.1", title: "Markdown renderer component", file: "M5-frontend.md", deps: [], milestone: "M5", core: true },
 { id: "P5.2", title: "Dual render in ChapterCard", file: "M5-frontend.md", deps: ["P5.1"], milestone: "M5", core: true },
 { id: "P5.3", title: "Async getStory + data-driven picker", file: "M5-frontend.md", deps: ["P5.2", "P3.1"], milestone: "M5", core: true },
 { id: "P5.4", title: "Hash routing + empty states", file: "M5-frontend.md", deps: ["P5.3"], milestone: "M5", core: false },
 { id: "P5.5", title: "Vite dev proxy /api + /media", file: "M5-frontend.md", deps: [], milestone: "M5", core: false },
 { id: "P6.1", title: "StoryForm", file: "M6-builder.md", deps: ["P5.4", "P3.1"], milestone: "M6", core: true },
 { id: "P6.2", title: "ChapterEditor add/edit/reorder/preview", file: "M6-builder.md", deps: ["P6.1", "P3.2", "P3.3"], milestone: "M6", core: true },
 { id: "P6.3", title: "MediaUpload (external URL or local file)", file: "M6-builder.md", deps: ["P4.3", "P6.2"], milestone: "M6", core: true },
 { id: "P6.4", title: "End-to-end sign-off (M6 gate)", file: "M6-builder.md", deps: ["P6.1", "P6.2", "P6.3"], milestone: "M6", core: true },
 { id: "P7.1", title: "Upload store abstraction (S3/Drive)", file: "M7-hardening.md", deps: ["P4.1"], milestone: "M7", core: false },
 { id: "P7.2", title: "Moderation gate before public", file: "M7-hardening.md", deps: ["P3.1"], milestone: "M7", core: false },
 { id: "P7.3", title: "Soft-delete purge cron", file: "M7-hardening.md", deps: ["P4.4", "P1.2"], milestone: "M7", core: false },
 { id: "P7.4", title: "Docs + security-checklist sign-off", file: "M7-hardening.md", deps: ["P3.1", "P4.1", "P7.3"], milestone: "M7", core: false },
];

const MILESTONES: Milestone[] = ["M1", "M2", "M3", "M4", "M5", "M6", "M7"];
const coreIds = (m: Milestone): string[] => CARDS.filter((c) => c.milestone === m && c.core).map((c) => c.id);
const coreM6 = coreIds("M6");

// A card's repair-loop hangs off its own card deps (mirrors storymap-pieces.tsx).
function cardDeps(m: Milestone, c: Card): string[] {
  if (c.deps.length) return c.deps;
  const idx = MILESTONES.indexOf(m);
  return idx === 0 ? ["recovery-gate"] : [`${MILESTONES[idx - 1]}-recGate`];
}

// ---- output schemas ----
const inputSchema = z.object({}).strict();
// FIXER output (producer of the review loop)
const repairSchema = z.object({
  editsApplied: z.boolean(),
  summary: z.string(),
  filesChanged: z.array(z.string()).default([]),
});
// JUDGE output (reviewer of the review loop) — MUST carry `approved`
const judgeSchema = z.object({
  approved: z.boolean(),
  reason: z.string(),
});
// the diagnose level output of the escalation chain
const diagSchema = z.object({
  reason: z.string(),
  recommendedMove: z.string(),
  resolved: z.boolean(),
});
// the escalation chain tracked output
const escalationSchema = z.object({
  level: z.string(),
  escalatedTo: z.string(),
  note: z.string(),
});
const approvalSchema = z.object({ approved: z.boolean(), reason: z.string().optional() });

// ---- the two knobs the user controls ----
const CARD_AUTO_FIX_MAX = 1; // 1 = "one" auto-fix attempt; 3 = "a few"
const HUMAN_FALLBACK = true; // true = escalate to a human Approval on unrecovered failure
const GATE = "recovery-gate"; // first milestone's repair loops depend on this

const { Workflow, Task, Sequence, smithers, outputs } = createSmithers({
  input: inputSchema,
  repair: repairSchema,
  judge: judgeSchema,
  diag: diagSchema,
  escalation: escalationSchema,
  approval: approvalSchema,
});

// FIXER instruction (producer prompt). ONE scoped repair; judge then decides.
function fixerPrompt(c: Card): string {
  return `Auto-fix work card ${c.id} — "${c.title}" — which FAILED its VERIFY or timed out.

Re-read first (single source of truth):
    - delegation/HANDOUT.md
    - delegation/cards/${c.file}
    - delegation/STATUS.md
    - the current files in this card's SCOPE and its VERIFY output

Then make ONE targeted, IN-SCOPE repair so card ${c.id}'s VERIFY passes.
Allowed repairs ONLY: fix the generated code in this card's SCOPE; correct a wrong
VERIFY command; fix the environment (go get <dep>, ollama config); or relax a too-tight
timeout/maxTokens. Do NOT commit. Do NOT open a PR. Do NOT touch files outside SCOPE.
Return { editsApplied, summary, filesChanged[] }.`;
}

export default smithers((ctx) => {
  return (
      <Workflow name="storymap-pieces-recovery">
        <Sequence name="all">
          <Approval
         id="recovery-gate"
         output={outputs.approval}
         onDeny="continue"
         request={{
         title: "Confirm auto-fix + escalate behaviour",
         summary: `On a failed/timed-out card, run ONE AI-judged auto-fix (ReviewLoop, maxIterations=${CARD_AUTO_FIX_MAX}, onMaxReached=fail): producer=agents.implement (ollama fixer), reviewer=agents.review (an AI judge that must return {approved}). Then a diagnose level (agents.research) with a human fallback (humanFallback=${HUMAN_FALLBACK}) that self-gates on the repair results. Approve to enable.`,
           }}
         />
         {MILESTONES.map((m) => {
            const cards = CARDS.filter((c) => c.milestone === m);
            const repairIds = cards.map((c) => `repair-${c.id}`);
            return (
                   <Sequence key={`ms-${m}`} name={`recovery-${m}`}>
                    {cards.map((c) => (
                       <ReviewLoop
                       key={c.id}
                       id={`repair-${c.id}`}
                       dependsOn={cardDeps(m, c)}
                       producer={agents.implement}
                       reviewer={agents.review}
                       produceOutput={outputs.repair}
                       reviewOutput={outputs.judge}
                       maxIterations={CARD_AUTO_FIX_MAX}
                       onMaxReached="fail">
                           {fixerPrompt(c)}
                         </ReviewLoop>
                        ))}
                        <EscalationChain
                        key={`esc-${m}`}
                        id={`escalate-${m}`}
                        dependsOn={repairIds}
                        levels={[
                        {
                           agent: agents.research,
                           output: outputs.diag,
                              label: "diagnose",
                              escalateIf: (r: { resolved?: boolean } | null | undefined) =>
                                         !(r?.resolved),
                            },
                           ]}
                        escalationOutput={outputs.escalation}
                        humanFallback={HUMAN_FALLBACK}
                        humanRequest={{
                         title: `Milestone ${m}: a card did not recover`,
                         summary: `An auto-fix loop for ${m} did not reach approval after ${CARD_AUTO_FIX_MAX} attempt(s). Diagnose, recommend one move, then escalate to a human.`,
                            }}>
                                 {`Diagnose milestone ${m}'s unresolved auto-fix. Read each card's repair result {editsApplied, summary, filesChanged} and the judge reason. Return { reason, recommendedMove, resolved }. If nothing actually failed, set resolved=true.`}
                               </EscalationChain>
                     <Approval
                     key={`${m}-recGate`}
                     id={`${m}-recGate`}
                     dependsOn={[`escalate-${m}`]}
                     output={outputs.approval}
                     onDeny="continue"
                     request={{
                     title: `Approve recovery ${m}`,
                     summary: `Milestone ${m} recovery complete. Any card that did not recover was diagnosed and escalated to a human.`,
                        }}
                     />
                  </Sequence>
              );
           })}
           <Task id="recovery-report" agent={agents.review} dependsOn={coreM6.map((id) => `repair-${id}`)} output={outputs.repair}>
                {() => `Auto-fix + escalate enabled per-card (maxAttempts=${CARD_AUTO_FIX_MAX}, humanFallback=${HUMAN_FALLBACK}). No in-scope card failed this run.`}
           </Task>
        </Sequence>
      </Workflow>
   );
});
