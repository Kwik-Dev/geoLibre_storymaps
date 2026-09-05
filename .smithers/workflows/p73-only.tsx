// smithers-source: custom
// smithers-display-name: P7.3 only (soft-delete purge cron)
/** @jsxImportSource smthrs */
//
// Minimal one-card workflow: re-runs ONLY P7.3 (soft-delete purge cron) using
// the same agent pool (agents.implement / agents.review -> pi_ollama_cloud)
// that ran the full storymap-pieces workflow. Used to finish the one remaining
// M7 card after the full run's durable state was lost.

import { createSmithers, ReviewLoop } from "smthrs";
import { agents } from "../agents";
import { z } from "zod/v4";

const cardSchema = z.object({
  cardId: z.string(),
  done: z.boolean(),
  verifyPassed: z.boolean(),
  summary: z.string(),
  statusUpdate: z.string(),
});
const judgeSchema = z.object({
  approved: z.boolean(),
  reason: z.string(),
});

const { Workflow, smithers, outputs } = createSmithers({
  card: cardSchema,
  judge: judgeSchema,
});

const cardPrompt = [
  `Implement work card P7.3 — "Soft-delete purge cron" of the user-created storymaps feature.`,
  "",
  "Read first (single source of truth):",
  "    - delegation/HANDOUT.md (decisions, worker model, env deps)",
  "    - delegation/cards/M7-hardening.md (the P7.3 section)",
  "    - delegation/STATUS.md",
  "",
  "Implement ONLY card P7.3's SCOPE, then run its VERIFY until it passes, then",
  " mark P7.3 \"done\" in delegation/STATUS.md (paste the VERIFY output + a timestamp).",
  "Commit your work with GitButler (`but`): run `but diff` to get file/hunk IDs, then",
  "  but commit -b storymap-P7.3 -m \"P7.3: Soft-delete purge cron\" <ids>",
  "Commit ONLY this card's source files. Do NOT commit delegation/STATUS.md — it is a",
  "shared live board; leave its changes uncommitted for the human to integrate.",
  "COMMIT ONLY — do NOT run `but push` and do NOT open a PR.",
  "Do NOT edit files outside this card's SCOPE.",
  "If VERIFY keeps failing or you hit a blocker, return done=false, verifyPassed=false, with a one-line summary.",
  "",
  "Model note: you are running on the cloud model deepseek-v4-flash (ollama-cloud).",
  "Return your result as valid JSON matching the declared output schema — do not emit",
  "free-form prose or partial token fragments.",
].join("\n");

export default smithers((ctx) => (
  <Workflow name="p73-only">
    <ReviewLoop
      id="card-P7.3"
      producer={agents.implement}
      reviewer={agents.review}
      produceOutput={outputs.card}
      reviewOutput={outputs.judge}
      maxIterations={2}
      onMaxReached="return-last"
    >
      {cardPrompt}
    </ReviewLoop>
  </Workflow>
));
