import { PiAgent as SmithersPiAgent } from "smthrs";

// Back the agent pool with the LOCAL pi coding agent (shells out to the real
// `pi` binary on PATH, headless -p mode) running the local Ollama model
// qwen3.8:27b-mlx — no external API.
//
//  - `pi` is installed at ~/.local/share/fnm/.../bin/pi and supports headless
//    (`pi -p`).
//  - Ollama runs at http://127.0.0.1:11434 and lists `ollama/qwen3.8:27b-mlx`.
//
// Per the pi integration (https://smithers.sh/integrations/pi-integration):
//     new PiAgent({ provider, model, mode })
// `provider: "ollama"` routes to pi's native Ollama provider → local model.
//
// qwen3.8 is a *thinking* model; a tiny token budget gets consumed by hidden
// reasoning and returns empty text — give it a generous budget.
export const PiAgent = new SmithersPiAgent({
  provider: "ollama",
  model: "qwen3.8:27b-mlx",
  mode: "text",
  config: {
   maxTokens: 16000,
   },
  skipGitRepoCheck: true,
});

export const pi = PiAgent;
