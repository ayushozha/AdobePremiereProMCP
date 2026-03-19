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

/** Returned when the user is logged in via OAuth but no API key is available. */
export interface OAuthDetected {
  kind: "oauth-no-key";
  email: string | null;
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
 * If an OAuth session is detected but no API key is extractable, returns an
 * OAuthDetected object so the caller can show a helpful message.
 */
export async function resolveAuth(): Promise<AuthResult | OAuthDetected | null> {
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
  const claudeAuth = getClaudeAuth();
  if (claudeAuth && "apiKey" in claudeAuth) {
    return {
      provider: "anthropic",
      apiKey: claudeAuth.apiKey,
      model: process.env.MODEL || DEFAULT_MODELS.anthropic,
    };
  }
  // If OAuth detected but no key, remember it but keep looking for other sources
  const oauthDetected = claudeAuth as OAuthDetected | null;

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

  // If we detected an OAuth session earlier but found no usable API key
  // anywhere, surface that so the caller can guide the user.
  if (oauthDetected) {
    return oauthDetected;
  }

  return null;
}

// ── Claude Code auth ──────────────────────────────────────────────────

interface ClaudeAuthStatus {
  loggedIn?: boolean;
  authMethod?: string;
  email?: string;
  subscriptionType?: string;
  apiProvider?: string;
}

/**
 * Try to obtain an API key from Claude Code.
 * Returns { apiKey } if a key was found, an OAuthDetected sentinel if the user
 * is logged in via OAuth (no key available), or null if Claude CLI isn't set up.
 */
function getClaudeAuth(): { apiKey: string } | OAuthDetected | null {
  // Step 1: Check `claude auth status` to understand the login state.
  let status: ClaudeAuthStatus | null = null;
  try {
    const raw = execSync("claude auth status 2>/dev/null", {
      encoding: "utf-8",
      timeout: 5000,
    }).trim();
    if (raw) {
      status = JSON.parse(raw) as ClaudeAuthStatus;
    }
  } catch {
    // CLI not available or errored
  }

  // Step 2: Try to read an API key from Claude config files.
  const home = homedir();
  const candidates = [
    path.join(home, ".claude", "credentials.json"),
    path.join(home, ".claude", "config.json"),
    path.join(home, ".claude.json"),
  ];

  for (const filePath of candidates) {
    const key = extractKeyFromJson(filePath, "sk-ant-");
    if (key) return { apiKey: key };
  }

  // Step 3: If the user is logged in via OAuth (claude.ai), there is no API key
  // we can extract — the Anthropic SDK needs an sk-ant-* key. Surface this so
  // the caller can show a helpful message.
  if (status?.loggedIn && status.authMethod === "claude.ai") {
    return { kind: "oauth-no-key", email: status.email ?? null };
  }

  // Step 4: If logged in via console (API key method), the key might be in the
  // keychain and inaccessible to us. Nothing more we can do.
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

// ── Auth help messages ────────────────────────────────────────────────

export function printAuthHelp(color: { cyan: string; yellow: string; reset: string }): void {
  console.log();
  console.log("  Authenticate using any of these methods:");
  console.log();
  console.log(`  ${color.yellow}Anthropic (Claude):${color.reset}`);
  console.log(`    ${color.cyan}export ANTHROPIC_API_KEY="sk-ant-..."${color.reset}`);
  console.log(`    ${color.cyan}claude auth login --console${color.reset}         — API key via Anthropic Console`);
  console.log();
  console.log(`  ${color.yellow}OpenAI / Codex:${color.reset}`);
  console.log(`    ${color.cyan}codex login${color.reset}                         — Codex CLI auth`);
  console.log(`    ${color.cyan}export OPENAI_API_KEY="sk-..."${color.reset}`);
  console.log();
}

export function printOAuthHelp(
  color: { cyan: string; yellow: string; red: string; reset: string },
  email: string | null,
): void {
  console.log();
  if (email) {
    console.log(`  You are logged into Claude as ${color.cyan}${email}${color.reset}, but this`);
  } else {
    console.log(`  You are logged into Claude via ${color.cyan}claude.ai${color.reset} OAuth, but this`);
  }
  console.log(`  session does not provide an API key that the SDK can use.`);
  console.log();
  console.log(`  ${color.yellow}To fix this, do one of the following:${color.reset}`);
  console.log();
  console.log(`    1. ${color.cyan}export ANTHROPIC_API_KEY="sk-ant-..."${color.reset}`);
  console.log(`       Get a key from ${color.cyan}https://console.anthropic.com/settings/keys${color.reset}`);
  console.log();
  console.log(`    2. ${color.cyan}claude auth login --console${color.reset}`);
  console.log(`       Re-authenticate using Anthropic Console (API billing).`);
  console.log();
}
