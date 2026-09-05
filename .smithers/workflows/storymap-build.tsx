// smithers-source: hand-authored
// smithers-display-name: Storymap Feature Build
// smithers-description:
//   Implement the "user-created storymaps" feature milestone by milestone
//   (M1 backend skeleton → M7), each with an implement agent, an independent
//   review agent, and a human Approval gate. Pauses on any decision that needs
//   the user (per "notify me when implementing or reviewing").
// smithers-tags: feature, go, sql, auth, frontend

/** @jsxImportSource smthrs */
import { $ } from "bun";
import { createSmithers, Approval } from "smthrs";
import { z } from "zod/v4";
import { agents } from "../agents";

// ── Inputs ────────────────────────────────────────────────────────────────
const inputSchema = z.object({
   spec: z
      .string()
      .default("feature_request_user_created_storymap.md")
      .describe("Path to the locked design doc that is the source of truth."),
   lastMilestone: z
      .number()
      .int()
      .min(1)
      .max(7)
      .default(7)
      .describe("Stop after this milestone (inclusive). Default: build M1..M7."),
});

// ── Output schemas ────────────────────────────────────────────────────────
const envCheckSchema = z.object({
   haveAccount: z
      .boolean()
      .describe("A model account/agent is resolvable, so agent Tasks can run."),
   detail: z.string().describe("What was found or what is missing."),
   nextAction: z
      .string()
      .describe("The single most useful command/step if haveAccount is false."),
});

const decisionSchema = z.object({ approved: z.boolean() });

const implSchema = z.object({
   milestone: z.string().describe("Which milestone this task implemented."),
   summary: z.string().describe("What changed, grounded in the spec."),
   filesChanged: z.array(z.string()).default([]),
   buildPassed: z
      .boolean()
      .describe("True if the relevant build/test command ran green.")
      .default(false),
   blockers: z
      .array(z.string())
      .default([])
      .describe("Open issues that block full completion of this milestone."),
});

const reviewSchema = z.object({
   verdict: z.enum(["pass", "fail"]).describe("pass = ship it; fail = must fix."),
   blocking: z.boolean().describe("True if the review found must-fix issues."),
   findings: z.array(z.string()).default([]),
   specConformance: z
      .string()
      .describe("Does it match the design doc + the locked decisions?"),
});

const shipReportSchema = z.object({
   done: z.boolean(),
   milestonesReached: z.array(z.string()).default([]),
   gatesGranted: z.array(z.string()).default([]),
   openDecisions: z.array(z.string()).default([]),
   summary: z.string(),
   nextSteps: z.array(z.string()).default([]),
});

// ── Smithers API ───────────────────────────────────────────────────────────
const { Workflow, Task, Sequence, smithers, outputs } = createSmithers({
   input: inputSchema,
   envCheck: envCheckSchema,
   decision: decisionSchema,
   impl: implSchema,
   review: reviewSchema,
   shipReport: shipReportSchema,
});

// ── Spec text injected into every agent prompt ─────────────────────────────
function specRef(ctx) {
   const spec = ctx.input?.spec ?? "feature_request_user_created_storymap.md";
   return `The authoritative spec is \`${spec}\` (read it fully). Also relevant: \`storymap_architecure.md\`. Locked decisions the spec already encodes: (1) GitHub OAuth2 sign-in for users, local email/password is ADMIN-only; (2) per-story private/public visibility chosen by the owner; (3) per-chapter media is either an external URL the user supplies (S3/object-store/Google Drive/any https URL) OR a file uploaded to the Go server's local disk; (4) existing embedded storymaps keep HTML descriptions and keep working with no server; (5) backend-stored storymaps use Markdown descriptions (rendered + sanitized). Backend = Go + SQLite + a simple REST API. Do NOT contradict these decisions; if one is wrong, stop and raise it via the ask-human protocol instead of guessing.`;
}

const cliRunner = process.env.SMITHERS_BUNX ?? "bunx";

// A reusable "gate" Approval after a phase. Durable: the run pauses until a
// human resolves it — this is the "notify me when reviewing" behavior.
function phaseGate(ctx, phaseId, label) {
   const review = ctx.outputMaybe("review", { nodeId: `${phaseId}_review` });
   const impl = ctx.outputMaybe("impl", { nodeId: `${phaseId}_impl` });
   if (!impl || !review) return null;
   return (
      <Approval
         id={`${phaseId}_gate`}
         output={outputs.decision}
         onDeny="continue"
         request={{
            title: `${label}: approve to continue to the next milestone`,
            summary:
               `Milestone ${phaseId}\n\n` +
               `Implementer summary: ${impl.summary}\n` +
               `Files changed: ${(impl.filesChanged || []).join(", ") || "(none reported)"}\n` +
               `Build/test green: ${impl.buildPassed ? "yes" : "no"}\n` +
               `Open blockers: ${(impl.blockers || []).join("; ") || "none"}\n\n` +
               `Independent review verdict: ${review.verdict}${review.blocking ? " (BLOCKING)" : ""}\n` +
               `Spec conformance: ${review.specConformance}\n` +
               `Findings: ${(review.findings || []).join(" | ") || "none"}\n\n` +
               `Approve to proceed; deny with a --note to send the milestone back for fixes.`,
         }}
      />
   );
}

// An agent task: implement → (next phase reads the impl+review outputs).
function implTask(phaseId, label, mandate) {
   return (
      <Task
         id={`${phaseId}_impl`}
         output={outputs.impl}
         agent={agents.implement}
         deps={{ impl: outputs.impl, review: outputs.review }}
         timeoutMs={45 * 60_000}
      >
         {(deps) =>
            `Implement milestone ${label} of the user-created storymaps feature.

${specRef()}

Specific mandate for this milestone:
${mandate}

Work in the repo at the project root. Prefer small, verifiable changes. After coding, run the appropriate build/test command (Go: \`go build ./...\` and \`go vet ./...\`; frontend: \`npm run build\`) and report buildPassed honestly. List every file you changed in filesChanged. Put anything you could not finish in blockers. When you hit a genuine ambiguity or an irreversible/destructive action, do NOT guess — call \`smithers ask-human\` and wait for the decision. Do not open a pull request.`
         }
      </Task>
   );
}

function reviewTask(phaseId) {
   return (
      <Task
         id={`${phaseId}_review`}
         output={outputs.review}
         agent={agents.review}
         deps={{ impl: outputs.impl }}
         timeoutMs={25 * 60_000}
      >
         {(deps) =>
            `Independently review milestone ${phaseId} of the storymaps feature. The implementer reported:
${JSON.stringify(deps.impl, null, 2)}

${specRef()}

Check: does it match the spec and the locked decisions? Does the code build and (if present) test green? Are there security holes (XSS in the Markdown renderer, unvalidated uploads, missing auth on private-story/media routes, CSRF on the GitHub OAuth \state\, secrets in the client)? Is backward compatibility preserved (embedded HTML stories still render; no-server mode still works)? Do NOT edit files — review only. Set verdict "fail" and blocking=true if any must-fix issue remains, listing each in findings.`
         }
      </Task>
   );
}

export default smithers((ctx) => {
   const spec = ctx.input?.spec ?? "feature_request_user_created_storymap.md";
   const last = ctx.input?.lastMilestone ?? 7;

   const env = ctx.outputMaybe("envCheck", { nodeId: "env-check" });
   const designDecision = ctx.outputMaybe("decision", { nodeId: "design_gate" });
   const designGranted = designDecision?.approved === true;

   const milestones = [
      { id: "m1", label: "M1 Backend skeleton", mandate: "Create the Go module `server/` (pure-Go SQLite via modernc.org/sqlite, chi router): config from env, DB open + versioned idempotent migrations (WAL, busy_timeout) + optional seed of embedded *-storymap.json, and GET /api/health. Serve ../dist static. Run `go build ./...` green. (network: `go get` chi + modernc.org/sqlite)." },
      { id: "m2", label: "M2 GitHub SSO + admin", mandate: "Implement GitHub OAuth2 (GET /api/auth/github + /callback with CSRF \state verification, upsert users by github_id) and ADMIN-only local login (bcrypt, seeded from ADMIN_EMAIL/ADMIN_PASSWORD). JWT sessions, auth middleware requiring a session on all other /api routes. No public sign-up." },
      { id: "m3", label: "M3 Stories + chapters CRUD", mandate: "Stories + chapters endpoints (list/get/create/update/delete/reorder), DB→camelCase JSON adapter, private/public + owner/admin authz. Include the /api/stories/:id/export endpoint that dumps the legacy story JSON shape." },
      { id: "m4", label: "M4 Media", mandate: "POST /api/media/upload (magic-byte MIME validation, size cap, random basename, no traversal) to local disk; per-chapter media is EITHER an external https URL the user supplies (validated: https-only, length-bounded, optional host allow-list) OR a local asset. GET /media/:aid serves stored files, gated by story visibility for private stories. DELETE is soft." },
      { id: "m5", label: "M5 Frontend cut-over", mandate: "Async getStory, #/stories/<id> hash routing, data-driven story picker merging embedded + API stories, and the Markdown renderer (react-markdown + remark-gfm + rehype-sanitize) with an HTML fallback for embedded stories. Preserve no-server/static mode. `npm run build` green." },
      { id: "m6", label: "M6 Builder UI", mandate: "StoryForm (title/subtitle/byline/theme/visibility), ChapterEditor (add/edit/reorder, location pick, media type + external-URL-or-upload, live Markdown preview), MediaUpload. Create → set visibility → open via #/stories/<id>. `npm run build` green." },
      { id: "m7", label: "M7 Hardening", mandate: "Optional server-mediated object-store/Drive upload, moderation gate before a story becomes public, and a soft-delete purge cron. Document the new workflows in README.md and verify the security checklist in the spec." },
   ].filter((m) => m.id.replace("m", "") <= last);

   return (
      <Workflow name="storymap-build">
         <Sequence>
            {/* 1 — Environment gate: surface a missing model account instead of failing deep. */}
            <Task id="env-check" output={outputs.envCheck}>
               {async () => {
                const res = await $`${cliRunner} smthrs agents list`.nothrow().quiet();
                const out = res.stdout?.toString() ?? "";
                const haveAccount =
                  !/no accounts registered/i.test(out) && /accounts:\s*(?!\[\s*\])/.test(out);
                const detail = haveAccount
                  ? "A model account is registered; agent Tasks can run."
                  : "No model account is registered. Agent Tasks (implement/review) will fail until one is added.";
                return {
                  haveAccount,
                  detail,
                  nextAction: haveAccount
                     ? ""
                     : "Set OPENROUTER_API_KEY (recommended: one key, many models) or run `${cliRunner} smthrs agents add` to register a provider account, then resume this run.",
               };
              }}
            </Task>

            {/* 2 — Pause until the user has satisfied the agent requirement. */}
            {env && !env.haveAccount ? (
               <Approval
                  id="env_gate"
                  output={outputs.decision}
                  onDeny="continue"
                  request={{
                     title: "No model account for Smithers agents — needs your action",
                     summary:
                       `${env.detail}\n\n${env.nextAction}\n\n` +
                       `The workflow will not execute implement/review agents until this is resolved. ` +
                       `Add the account/key, then run:  ${cliRunner} smthrs approve --node env_gate --by you`,
                  }}
               />
            ) : null}

            {/* 3 — Confirm the locked design is the spec before writing code. */}
            <Approval
               id="design_gate"
               output={outputs.decision}
               onDeny="fail"
               request={{
                  title: `Confirm spec \`${spec}\` and the locked decisions`,
                  summary:
                     `I will implement milestones ${milestones.map((m) => m.id).join(" → ")} from \`${spec}\`.\n\n` +
                     `Locked decisions: GitHub OAuth for users (local pw = admin only); per-story private/public; media = external URL OR local upload; embedded HTML stories keep working server-less; backend stories use Markdown.\n\n` +
                     `Approve to start coding. Deny to re-open the design.`,
               }}
            />

            {[...Array(last).keys()].map((i) => {
               const m = milestones[i];
               if (!m) return null;
               return (
                  <Sequence key={m.id}>
                     {implTask(m.id, m.label, m.mandate)}
                     {reviewTask(m.id)}
                     {phaseGate(ctx, m.id, m.label)}
                  </Sequence>
               );
            })}

            {/* Final durable summary row for `smithers output`. */}
            <Task id="ship-report" output={outputs.shipReport}>
               {async () => {
                const reached = milestones
                   .map((m) => m.id)
                   .filter((id) => !!ctx.outputMaybe("review", { nodeId: `${id}_review` }));
                const gates = reached.filter((id) => ctx.outputMaybe("decision", { nodeId: `${id}_gate` })?.approved);
                return {
                  done: reached.length === milestones.length,
                  milestonesReached: reached,
                  gatesGranted: gates,
                  openDecisions: milestones
                     .map((m) => m.id)
                     .filter((id) => {
                        const rv = ctx.outputMaybe("review", { nodeId: `${id}_review` });
                        return rv && rv.blocking;
                     }),
                  summary: `${reached.length}/${milestones.length} milestones reviewed; ${gates.length} phase gates granted.`,
                  nextSteps:
                     reached.length < milestones.length
                        ? [`Resume: ${cliRunner} smthrs up .smithers/workflows/storymap-build.tsx --resume true`]
                        : ["Hand off to the user for manual QA; no PR opened."],
               };
              }}
            </Task>
         </Sequence>
      </Workflow>
   );
});
