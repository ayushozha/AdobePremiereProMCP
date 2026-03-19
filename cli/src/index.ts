#!/usr/bin/env node

/**
 * PremierPro AI Editor CLI
 *
 * An interactive, AI-powered command-line interface for controlling
 * Adobe Premiere Pro. Supports both Claude (Anthropic) and GPT/Codex (OpenAI).
 */

import { MCPClient } from "./mcp-client.js";
import { ChatLoop } from "./chat.js";
import { resolveAuth, printAuthHelp, printOAuthHelp } from "./auth.js";
import type { AuthResult } from "./auth.js";
import {
  banner,
  printAssistant,
  printError,
  printInfo,
  createReadlineInterface,
  prompt,
  color,
} from "./ui.js";

// ── Main ──────────────────────────────────────────────────────────────

async function main(): Promise<void> {
  banner();

  // 1. Resolve authentication (Anthropic or OpenAI)
  const authResult = await resolveAuth();

  if (!authResult) {
    printError("No authentication found.");
    printAuthHelp(color);
    process.exit(1);
  }

  // Handle OAuth detection (logged in via claude.ai but no API key)
  if ("kind" in authResult && authResult.kind === "oauth-no-key") {
    printError("Claude OAuth session detected, but no API key available.");
    printOAuthHelp(color, authResult.email);
    process.exit(1);
  }

  const auth = authResult as AuthResult;

  printInfo(
    `  Authenticated with ${auth.provider === "anthropic" ? "Claude" : "OpenAI"} (model: ${auth.model})`,
  );

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

  // 4. Auto-launch Premiere Pro if it is not already running
  try {
    const statusResult = await mcpClient.callTool("premiere_is_running", {});
    const isRunning =
      !statusResult.isError &&
      statusResult.content.toLowerCase().includes("true");

    if (!isRunning) {
      printInfo("  Premiere Pro is not running. Launching...");
      const launchResult = await mcpClient.callTool("premiere_open", {});
      if (launchResult.isError) {
        printError(`Failed to launch Premiere Pro: ${launchResult.content}`);
      } else {
        printInfo("  Premiere Pro launched.");
      }
    } else {
      printInfo("  Premiere Pro is running.");
    }
  } catch (err) {
    // Non-fatal: tools may not include premiere_is_running if the server
    // is a different version. Continue into the chat loop regardless.
    const msg = err instanceof Error ? err.message : String(err);
    printInfo(`  (Could not check Premiere Pro status: ${msg})`);
  }

  console.log();

  // 5. Set up the chat loop with the resolved auth
  const chatLoop = new ChatLoop(mcpClient, auth);
  const rl = createReadlineInterface();

  // 6. Handle graceful shutdown
  const shutdown = async (): Promise<void> => {
    console.log();
    printInfo("  Shutting down...");
    rl.close();
    await mcpClient.disconnect();
    process.exit(0);
  };

  process.on("SIGINT", () => void shutdown());
  process.on("SIGTERM", () => void shutdown());

  // 7. Interactive loop
  while (true) {
    const input = await prompt(rl);

    if (input === null) {
      await shutdown();
      break;
    }

    const trimmed = input.trim();
    if (trimmed === "") continue;

    if (["exit", "quit", "q"].includes(trimmed.toLowerCase())) {
      await shutdown();
      break;
    }

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
