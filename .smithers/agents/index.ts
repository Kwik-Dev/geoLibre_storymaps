export { ClaudeCodeAgent } from "./claude-code";
export { CodexAgent } from "./codex";
export { CursorAgent } from "./cursor";
export { OpenCodeAgent } from "./opencode";
export { AntigravityAgent } from "./antigravity";

// Per-card build workflow pools. The LOCAL pi agent on qwen3.8:27b-mlx
// (see ./pi.ts) does NOT reliably emit the structured JSON the card output
// schemas require — it returns incoherent token fragments, so the card
// producer fails its schema validation in a retry loop. The CLOUD
// deepseek-v4-flash model (ollama-cloud) DOES produce valid structured
// output, so `implement`/`planning` are wired to it (same model as `review`).
// Run a workflow with:
//   smithers up .smithers/workflows/storymap-pieces.tsx
//   or one card at a time:
//   smithers oneshot "implement card P1.1 in server/; run VERIFY; update STATUS.md; no PR"
import { PiAgent, pi as PiAgentInstance, PiAgentCloud } from "./pi";
export { PiAgent, PiAgentCloud };
export const implement = PiAgentCloud;
export const review = PiAgentCloud;
export const planning = PiAgentCloud;
export const research = PiAgentCloud;

// Legacy per-runner instances remain importable for direct use (vendor-specific).
export { ClaudeCodeAgent as claudeCode } from "./claude-code";
export { CodexAgent as codex } from "./codex";
export { CursorAgent as cursor } from "./cursor";
export { OpenCodeAgent as opencode } from "./opencode";
export { AntigravityAgent as antigravity } from "./antigravity";
