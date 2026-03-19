#!/usr/bin/env node

/**
 * PremierPro AI Editor CLI
 *
 * An interactive, AI-powered command-line interface for controlling
 * Adobe Premiere Pro. The user types natural language; Claude decides
 * which MCP tools to call; results are shown in the terminal.
 */

import { MCPClient } from "./mcp-client.js";
import { ChatLoop } from "./chat.js";
import {
  banner,
  printAssistant,
  printError,
  printInfo,
  createReadlineInterface,
  prompt,
  color,
} from "./ui.js";

import { execSync } from "node:child_process";
import { readFileSync, existsSync } from "node:fs";
import { homedir } from "node:os";
import * as path from "node:path";

/**
 * Try to get an Anthropic API key from Claude Code's stored credentials.
 * Checks multiple locations where Claude Code may store auth data.
 */
async function getClaudeApiKey(): Promise<string | null> {
  // Method 1: Try `claude auth print-api-key` (if available)
  try {
    const key = execSync("claude auth print-api-key 2>/dev/null", {
      encoding: "utf-8",
      timeout: 5000,
    }).trim();
    if (key && key.startsWith("sk-ant-")) return key;
  } catch {
    // Command not available or failed
  }

  // Method 2: Check Claude Code's config files for stored API key
  const home = homedir();
  const candidates = [
    path.join(home, ".claude", "credentials.json"),
    path.join(home, ".claude", "config.json"),
    path.join(home, ".claude.json"),
  ];

  for (const filePath of candidates) {
    if (!existsSync(filePath)) continue;
    try {
      const data = JSON.parse(readFileSync(filePath, "utf-8"));
      // Look for API key in various possible fields
      const key =
        data.apiKey ||
        data.api_key ||
        data.anthropicApiKey ||
        data.ANTHROPIC_API_KEY ||
        data.credentials?.apiKey ||
        data.credentials?.api_key;
      if (key && typeof key === "string" && key.startsWith("sk-ant-")) {
        return key;
      }
    } catch {
      // Skip unparseable files
    }
  }

  return null;
}

// ── Main ──────────────────────────────────────────────────────────────

async function main(): Promise<void> {
  banner();

  // 1. Check for API key — support ANTHROPIC_API_KEY or claude login
  if (!process.env.ANTHROPIC_API_KEY) {
    // Try to get API key from Claude Code's auth
    const apiKey = await getClaudeApiKey();
    if (apiKey) {
      process.env.ANTHROPIC_API_KEY = apiKey;
    } else {
      printError("No authentication found.");
      console.log();
      console.log("  Authenticate using one of these methods:");
      console.log();
      console.log(
        `    ${color.cyan}claude login${color.reset}                    — browser OAuth (recommended)`,
      );
      console.log(
        `    ${color.cyan}claude login --method api-key${color.reset}   — paste an API key`,
      );
      console.log(
        `    ${color.cyan}export ANTHROPIC_API_KEY="sk-ant-..."${color.reset}  — set env var directly`,
      );
      console.log();
      process.exit(1);
    }
  }

  // 2. Spawn and connect to the MCP server
  const mcpClient = new MCPClient();

  printInfo("  Connecting to PremierPro MCP server...");

  try {
    await mcpClient.connect();
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    printError(`Failed to start MCP server: ${msg}`);
    console.log();
    console.log("  Make sure the server binary exists at:");
    console.log(
      `    ${color.cyan}go-orchestrator/bin/premierpro-mcp${color.reset}`,
    );
    console.log();
    console.log("  Build it with:");
    console.log(
      `    ${color.cyan}cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server${color.reset}`,
    );
    console.log();
    process.exit(1);
  }

  // 3. Fetch available tools
  let toolCount: number;
  try {
    const tools = await mcpClient.listTools();
    toolCount = tools.length;
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    printError(`Failed to list MCP tools: ${msg}`);
    await mcpClient.disconnect();
    process.exit(1);
  }

  printInfo(`  Connected. ${toolCount} tools available.`);
  console.log();

  // 4. Set up the chat loop
  const chatLoop = new ChatLoop(mcpClient);
  const rl = createReadlineInterface();

  // 5. Handle graceful shutdown
  const shutdown = async (): Promise<void> => {
    console.log();
    printInfo("  Shutting down...");
    rl.close();
    await mcpClient.disconnect();
    process.exit(0);
  };

  process.on("SIGINT", () => {
    void shutdown();
  });
  process.on("SIGTERM", () => {
    void shutdown();
  });

  // 6. Interactive loop
  // eslint-disable-next-line no-constant-condition
  while (true) {
    const input = await prompt(rl);

    // Ctrl+D or closed input
    if (input === null) {
      await shutdown();
      break;
    }

    const trimmed = input.trim();

    // Skip empty input
    if (trimmed === "") {
      continue;
    }

    // Exit commands
    if (
      trimmed.toLowerCase() === "exit" ||
      trimmed.toLowerCase() === "quit" ||
      trimmed.toLowerCase() === "q"
    ) {
      await shutdown();
      break;
    }

    // Process the message through Claude
    try {
      const response = await chatLoop.processUserMessage(trimmed);
      if (response) {
        printAssistant(response);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      printError(`Chat error: ${msg}`);
    }
  }
}

// ── Entry ─────────────────────────────────────────────────────────────

main().catch((err) => {
  printError(`Fatal: ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});
