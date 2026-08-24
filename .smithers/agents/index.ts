export { ClaudeCodeAgent } from "./claude-code";
export { CodexAgent } from "./codex";
export { CursorAgent } from "./cursor";
export { OpenCodeAgent } from "./opencode";
export { AntigravityAgent } from "./antigravity";

// Per-card build workflow pools, backed by the LOCAL pi coding agent on the
// Ollama model qwen3.8:27b-mlx (see ./pi.ts). pi shells out to the real `pi`
// binary in headless mode — no external API. Run a workflow with:
//   smithers up .smithers/workflows/storymap-pieces.tsx
//   or one card at a time:
//   smithers oneshot "implement card P1.1 in server/; run VERIFY; update STATUS.md; no PR"
import { PiAgent, pi as PiAgentInstance, PiAgentCloud } from "./pi";
export { PiAgent, PiAgentCloud };
export const implement = PiAgentInstance;
export const review = PiAgentCloud;
export const planning = PiAgentInstance;
export const research = PiAgentCloud;

// Legacy per-runner instances remain importable for direct use (vendor-specific).
export { ClaudeCodeAgent as claudeCode } from "./claude-code";
export { CodexAgent as codex } from "./codex";
export { CursorAgent as cursor } from "./cursor";
export { OpenCodeAgent as opencode } from "./opencode";
export { AntigravityAgent as antigravity } from "./antigravity";
