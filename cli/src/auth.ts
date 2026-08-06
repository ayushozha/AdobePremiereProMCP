/**
 * Authentication module — resolves API credentials for Claude (Anthropic)
 * or OpenAI from explicit environment variables and config files.
 */

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
 * Application OAuth sessions are deliberately not inspected or reused.
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

  // Priority 2: Explicit PremierPro MCP config file
  const configAuth = getAuthFromConfigFiles();
  if (configAuth) {
    return configAuth;
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

// ── Auth help messages ────────────────────────────────────────────────

export function printAuthHelp(color: { cyan: string; yellow: string; reset: string }): void {
  console.log();
  console.log("  Authenticate using any of these methods:");
  console.log();
  console.log(`  ${color.yellow}Anthropic (Claude):${color.reset}`);
  console.log(`    ${color.cyan}export ANTHROPIC_API_KEY="sk-ant-..."${color.reset}`);
  console.log();
  console.log(`  ${color.yellow}OpenAI:${color.reset}`);
  console.log(`    ${color.cyan}export OPENAI_API_KEY="sk-..."${color.reset}`);
  console.log();
  console.log(`  Or store either key in ~/.premierpro-mcp/config.json.`);
  console.log(`  Claude/Codex subscription OAuth sessions are not API keys.`);
  console.log();
}
