/**
 * Authentication module — resolves API credentials for Claude (Anthropic)
 * or OpenAI/Codex from environment variables, CLI tools, and config files.
 */

import { execSync } from "node:child_process";
import { readFileSync, existsSync } from "node:fs";
import { homedir } from "node:os";
import * as path from "node:path";

// ── Types ─────────────────────────────────────────────────────────────

export type Provider = "anthropic" | "openai";

export interface AuthResult {
  provider: Provider;
  apiKey: string;
  model: string;
}

// ── Default models per provider ───────────────────────────────────────

const DEFAULT_MODELS: Record<Provider, string> = {
  anthropic: "claude-sonnet-4-20250514",
  openai: "gpt-4o",
};

// ── Main resolve function ─────────────────────────────────────────────

/**
 * Resolve authentication by checking all sources in priority order.
 * Returns the provider, API key, and default model — or null if nothing found.
 */
export async function resolveAuth(): Promise<AuthResult | null> {
  // Priority 1: Explicit env vars
  if (process.env.ANTHROPIC_API_KEY) {
    return {
      provider: "anthropic",
      apiKey: process.env.ANTHROPIC_API_KEY,
      model: process.env.MODEL || DEFAULT_MODELS.anthropic,
    };
  }

  if (process.env.OPENAI_API_KEY) {
    return {
      provider: "openai",
      apiKey: process.env.OPENAI_API_KEY,
      model: process.env.MODEL || DEFAULT_MODELS.openai,
    };
  }

  // Priority 2: Claude Code auth (claude login)
  const claudeKey = getClaudeAuthKey();
  if (claudeKey) {
    return {
      provider: "anthropic",
      apiKey: claudeKey,
      model: process.env.MODEL || DEFAULT_MODELS.anthropic,
    };
  }

  // Priority 3: Codex CLI auth
  const codexKey = getCodexAuthKey();
  if (codexKey) {
    return {
      provider: "openai",
      apiKey: codexKey,
      model: process.env.MODEL || DEFAULT_MODELS.openai,
    };
  }

  // Priority 4: Config files
  const configAuth = getAuthFromConfigFiles();
  if (configAuth) {
    return configAuth;
  }

  return null;
}

// ── Claude Code auth ──────────────────────────────────────────────────

function getClaudeAuthKey(): string | null {
  // Try `claude auth print-api-key`
  try {
    const key = execSync("claude auth print-api-key 2>/dev/null", {
      encoding: "utf-8",
      timeout: 5000,
    }).trim();
    if (key && key.startsWith("sk-ant-")) return key;
  } catch {
    // not available
  }

  // Check Claude config files
  const home = homedir();
  const candidates = [
    path.join(home, ".claude", "credentials.json"),
    path.join(home, ".claude", "config.json"),
    path.join(home, ".claude.json"),
  ];

  for (const filePath of candidates) {
    const key = extractKeyFromJson(filePath, "sk-ant-");
    if (key) return key;
  }

  return null;
}

// ── Codex / OpenAI auth ───────────────────────────────────────────────

function getCodexAuthKey(): string | null {
  // Try `codex auth print-api-key` (if codex CLI exists)
  try {
    const key = execSync("codex auth print-api-key 2>/dev/null", {
      encoding: "utf-8",
      timeout: 5000,
    }).trim();
    if (key && key.startsWith("sk-")) return key;
  } catch {
    // not available
  }

  // Check OpenAI/Codex config files
  const home = homedir();
  const candidates = [
    path.join(home, ".codex", "credentials.json"),
    path.join(home, ".codex", "config.json"),
    path.join(home, ".config", "openai", "credentials.json"),
    path.join(home, ".openai", "credentials.json"),
  ];

  for (const filePath of candidates) {
    const key = extractKeyFromJson(filePath, "sk-");
    if (key) return key;
  }

  return null;
}

// ── Config file parsing ───────────────────────────────────────────────

function getAuthFromConfigFiles(): AuthResult | null {
  const home = homedir();

  // Check for a premierpro-specific config
  const ppConfig = path.join(home, ".premierpro-mcp", "config.json");
  if (existsSync(ppConfig)) {
    try {
      const data = JSON.parse(readFileSync(ppConfig, "utf-8"));
      if (data.anthropic_api_key || data.ANTHROPIC_API_KEY) {
        return {
          provider: "anthropic",
          apiKey: data.anthropic_api_key || data.ANTHROPIC_API_KEY,
          model: data.model || DEFAULT_MODELS.anthropic,
        };
      }
      if (data.openai_api_key || data.OPENAI_API_KEY) {
        return {
          provider: "openai",
          apiKey: data.openai_api_key || data.OPENAI_API_KEY,
          model: data.model || DEFAULT_MODELS.openai,
        };
      }
    } catch {
      // skip
    }
  }

  return null;
}

function extractKeyFromJson(filePath: string, prefix: string): string | null {
  if (!existsSync(filePath)) return null;
  try {
    const data = JSON.parse(readFileSync(filePath, "utf-8"));
    const candidates = [
      data.apiKey,
      data.api_key,
      data.anthropicApiKey,
      data.ANTHROPIC_API_KEY,
      data.openaiApiKey,
      data.OPENAI_API_KEY,
      data.credentials?.apiKey,
      data.credentials?.api_key,
    ];
    for (const key of candidates) {
      if (key && typeof key === "string" && key.startsWith(prefix)) {
        return key;
      }
    }
  } catch {
    // skip
  }
  return null;
}

// ── Auth help message ─────────────────────────────────────────────────

export function printAuthHelp(color: { cyan: string; yellow: string; reset: string }): void {
  console.log();
  console.log("  Authenticate using any of these methods:");
  console.log();
  console.log(`  ${color.yellow}Anthropic (Claude):${color.reset}`);
  console.log(`    ${color.cyan}claude login${color.reset}                        — browser OAuth`);
  console.log(`    ${color.cyan}claude login --method api-key${color.reset}       — paste API key`);
  console.log(`    ${color.cyan}export ANTHROPIC_API_KEY="sk-ant-..."${color.reset}`);
  console.log();
  console.log(`  ${color.yellow}OpenAI / Codex:${color.reset}`);
  console.log(`    ${color.cyan}codex login${color.reset}                         — Codex CLI auth`);
  console.log(`    ${color.cyan}export OPENAI_API_KEY="sk-..."${color.reset}`);
  console.log();
}
