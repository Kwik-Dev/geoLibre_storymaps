// smithers-source: generated
import { type AgentLike } from "smthrs";
import { OpenAIAgent as SmithersOpenAIAgent } from "smthrs";
import { PiAgent as SmithersPiAgent } from "smthrs";
// import { ClaudeCodeAgent as SmithersClaudeCodeAgent } from "smthrs";
// import { CodexAgent as SmithersCodexAgent } from "smthrs";
// import { CursorAgent as SmithersCursorAgent } from "smthrs";
// import { OpenCodeAgent as SmithersOpenCodeAgent } from "smthrs";
// import { AntigravityAgent as SmithersAntigravityAgent } from "smthrs";
// import { OmpAgent as SmithersOmpAgent } from "smthrs";
// import { KimiAgent as SmithersKimiAgent } from "smthrs";
// import { AmpAgent as SmithersAmpAgent } from "smthrs";
// import { VibeAgent as SmithersVibeAgent } from "smthrs";
// import { HermesCliAgent as SmithersHermesCliAgent } from "smthrs";
// import { OpenClawAgent as SmithersOpenClawAgent } from "smthrs";
// import { PoolAgent as SmithersPoolAgent } from "smthrs";

// export { ClaudeCodeAgent } from "./agents/claude-code";
// export { CodexAgent } from "./agents/codex";
// export { CursorAgent } from "./agents/cursor";
// export { OpenCodeAgent } from "./agents/opencode";
// export { AntigravityAgent } from "./agents/antigravity";

class SmithersOpenRouterAgent extends SmithersOpenAIAgent {
  generate(args = {}) {
    if (!process.env.OPENROUTER_API_KEY) {
      throw new Error("Smithers generated an OpenRouter default agent, but OPENROUTER_API_KEY is not set. Set OPENROUTER_API_KEY, or run `smithers agent add` to configure another agent, then rerun this workflow.");
    }
    return super.generate(args);
  }
}

function createOpenRouterAgent() {
  return new SmithersOpenRouterAgent({
    model: "openai/gpt-5.4-mini",
    baseURL: "https://openrouter.ai/api/v1",
    apiKey: process.env.OPENROUTER_API_KEY,
  });
}

export const providers = {
//   claude: new SmithersClaudeCodeAgent({ model: "claude-fable-5" }),
//   codex: new SmithersCodexAgent({ model: "gpt-5.6-luna", config: { model_reasoning_effort: "medium" }, skipGitRepoCheck: true }),
//   cursor: new SmithersCursorAgent({ cwd: process.cwd() }),
  openrouter: createOpenRouterAgent(),
//   opencode: new SmithersOpenCodeAgent({ model: "anthropic/claude-fable-5" }),
//   antigravity: new SmithersAntigravityAgent(),
  pi: new SmithersPiAgent({ provider: "ollama-cloud", model: "deepseek-v4-flash:preview", mode: "text", config: { maxTokens: 256000 }, skipGitRepoCheck: true, noExtensions: true }),
//   omp: new SmithersOmpAgent({ model: "gpt-5.6-luna" }),
//   kimi: new SmithersKimiAgent({ model: "kimi-k2.7-code" }),
//   amp: new SmithersAmpAgent(),
//   vibe: new SmithersVibeAgent({ agent: "auto-approve" }),
//   hermes: new SmithersHermesCliAgent(),
//   openclaw: new SmithersOpenClawAgent(),
//   pool: new SmithersPoolAgent(),
//   codexSol: new SmithersCodexAgent({ model: "gpt-5.6-sol", config: { model_reasoning_effort: "xhigh" }, skipGitRepoCheck: true }),
//   codexTerra: new SmithersCodexAgent({ model: "gpt-5.6-terra", config: { model_reasoning_effort: "medium" }, skipGitRepoCheck: true }),
//   codexLuna: new SmithersCodexAgent({ model: "gpt-5.6-luna", config: { model_reasoning_effort: "medium" }, skipGitRepoCheck: true }),
//   claudeOpus: new SmithersClaudeCodeAgent({ model: "claude-opus-5" }),
//   claudeSonnet: new SmithersClaudeCodeAgent({ model: "claude-sonnet-5" }),
} as const;

export const agents = {
  // cheapFast: Smithers would normally suggest Codex Luna here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // cheapFast: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // cheapFast: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // cheapFast: Smithers would normally suggest Vibe here, but Vibe is not available: missing `vibe` on PATH; missing credentials (~/.vibe/.env or ~/.vibe/config.toml or $MISTRAL_API_KEY).
  // cheapFast: Smithers would normally suggest Antigravity here, but Antigravity is not available: missing `agy` on PATH; no registered Antigravity account and not authenticated (~/.gemini/antigravity-cli/settings.json or ~/.gemini/antigravity-cli).
  // cheapFast: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  cheapFast: [
    providers.pi,
    // providers.codexLuna,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.vibe,
    // providers.antigravity,
    // providers.openclaw,
    // providers.cursor,
  ],
  // research: Smithers would normally suggest Codex Luna here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // research: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // research: Smithers would normally suggest Antigravity here, but Antigravity is not available: missing `agy` on PATH; no registered Antigravity account and not authenticated (~/.gemini/antigravity-cli/settings.json or ~/.gemini/antigravity-cli).
  // research: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // research: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // research: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  // research: Smithers would normally suggest Cursor here, but Cursor is not available: missing `cursor-agent` on PATH; missing credentials (~/.cursor/auth.json or $CURSOR_API_KEY).
  research: [
    providers.pi,
    // providers.codexLuna,
    // providers.kimi,
    // providers.antigravity,
    // providers.opencode,
    // providers.claudeSonnet,
    // providers.openclaw,
    // providers.cursor,
  ],
  // implement: Smithers would normally suggest Claude Opus here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // implement: Smithers would normally suggest Codex Terra here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // implement: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // implement: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // implement: Smithers would normally suggest Antigravity here, but Antigravity is not available: missing `agy` on PATH; no registered Antigravity account and not authenticated (~/.gemini/antigravity-cli/settings.json or ~/.gemini/antigravity-cli).
  // implement: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // implement: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // implement: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  implement: [
    providers.pi,
    // providers.claudeOpus,
    // providers.codexTerra,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.antigravity,
    // providers.claude,
    // providers.opencode,
    // providers.openclaw,
    // providers.cursor,
  ],
  // midTier: Smithers would normally suggest Codex Terra here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // midTier: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // midTier: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // midTier: Smithers would normally suggest Antigravity here, but Antigravity is not available: missing `agy` on PATH; no registered Antigravity account and not authenticated (~/.gemini/antigravity-cli/settings.json or ~/.gemini/antigravity-cli).
  // midTier: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // midTier: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // midTier: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  // midTier: Smithers would normally suggest Cursor here, but Cursor is not available: missing `cursor-agent` on PATH; missing credentials (~/.cursor/auth.json or $CURSOR_API_KEY).
  midTier: [
    providers.pi,
    // providers.codexTerra,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.antigravity,
    // providers.opencode,
    // providers.claude,
    // providers.openclaw,
    // providers.cursor,
  ],
  // smartTool: Smithers would normally suggest Codex Terra here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // smartTool: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // smartTool: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // smartTool: Smithers would normally suggest Antigravity here, but Antigravity is not available: missing `agy` on PATH; no registered Antigravity account and not authenticated (~/.gemini/antigravity-cli/settings.json or ~/.gemini/antigravity-cli).
  // smartTool: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // smartTool: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // smartTool: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  // smartTool: Smithers would normally suggest Cursor here, but Cursor is not available: missing `cursor-agent` on PATH; missing credentials (~/.cursor/auth.json or $CURSOR_API_KEY).
  smartTool: [
    providers.pi,
    // providers.codexTerra,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.antigravity,
    // providers.opencode,
    // providers.claude,
    // providers.openclaw,
    // providers.cursor,
  ],
  // validate: Smithers would normally suggest Codex Terra here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // validate: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // validate: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // validate: Smithers would normally suggest Antigravity here, but Antigravity is not available: missing `agy` on PATH; no registered Antigravity account and not authenticated (~/.gemini/antigravity-cli/settings.json or ~/.gemini/antigravity-cli).
  // validate: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // validate: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // validate: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  // validate: Smithers would normally suggest Cursor here, but Cursor is not available: missing `cursor-agent` on PATH; missing credentials (~/.cursor/auth.json or $CURSOR_API_KEY).
  validate: [
    providers.pi,
    // providers.codexTerra,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.antigravity,
    // providers.opencode,
    // providers.claude,
    // providers.openclaw,
    // providers.cursor,
  ],
  // smart: Smithers would normally suggest Claude Opus here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // smart: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // smart: Smithers would normally suggest Codex Sol here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // smart: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // smart: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  smart: [
    providers.pi,
    // providers.claudeOpus,
    // providers.claude,
    // providers.codexSol,
    // providers.opencode,
    // providers.openclaw,
    // providers.antigravity,
    // providers.amp,
    // providers.kimi,
    // providers.cursor,
  ],
  // review: Smithers would normally suggest Codex Sol here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // review: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // review: Smithers would normally suggest Claude Opus here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // review: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // review: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // review: Smithers would normally suggest Amp here, but Amp is not available: missing `amp` on PATH; missing credentials (~/.amp).
  // review: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // review: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  review: [
    providers.pi,
    // providers.codexSol,
    // providers.claude,
    // providers.claudeOpus,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.amp,
    // providers.opencode,
    // providers.openclaw,
    // providers.cursor,
  ],
  // planning: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // planning: Smithers would normally suggest Claude Opus here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // planning: Smithers would normally suggest Codex Sol here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // planning: Smithers would normally suggest Claude Sonnet here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // planning: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // planning: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // planning: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  // planning: Smithers would normally suggest Cursor here, but Cursor is not available: missing `cursor-agent` on PATH; missing credentials (~/.cursor/auth.json or $CURSOR_API_KEY).
  planning: [
    providers.pi,
    // providers.claude,
    // providers.claudeOpus,
    // providers.codexSol,
    // providers.claudeSonnet,
    // providers.kimi,
    // providers.opencode,
    // providers.openclaw,
    // providers.cursor,
  ],
  // orchestrator: Smithers would normally suggest Claude Opus here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // orchestrator: Smithers would normally suggest Claude Code here, but Claude Code is not available: missing `claude` on PATH; no registered Claude Code account and not authenticated (~/.claude/.credentials.json or ~/.claude.json).
  // orchestrator: Smithers would normally suggest Kimi here, but Kimi is not available: missing `kimi` on PATH; no registered Kimi account and not authenticated (~/.kimi).
  // orchestrator: Smithers would normally suggest Codex Sol here, but Codex is not available: missing `codex` on PATH; no registered Codex account and not authenticated (~/.codex/auth.json or $OPENAI_API_KEY).
  // orchestrator: Smithers would normally suggest OpenCode here, but OpenCode is not available: missing `opencode` on PATH; missing credentials (~/.local/share/opencode/auth.json or ~/.config/opencode or ~/.local/share/opencode or $OPENCODE_API_KEY or $ANTHROPIC_API_KEY or $OPENAI_API_KEY or $GEMINI_API_KEY or $GOOGLE_API_KEY).
  // orchestrator: Smithers would normally suggest OpenClaw here, but OpenClaw is not available: missing `openclaw` on PATH; missing credentials (~/.openclaw/openclaw.json or ~/.openclaw).
  // orchestrator: Smithers would normally suggest Cursor here, but Cursor is not available: missing `cursor-agent` on PATH; missing credentials (~/.cursor/auth.json or $CURSOR_API_KEY).
  orchestrator: [
    providers.pi,
    // providers.claudeOpus,
    // providers.claude,
    // providers.kimi,
    // providers.codexSol,
    // providers.opencode,
    // providers.openclaw,
    // providers.cursor,
  ],
} as const satisfies Record<string, AgentLike[]>;
