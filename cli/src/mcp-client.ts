/**
 * MCP client that spawns and communicates with the Go MCP server over stdio.
 *
 * Uses the official @modelcontextprotocol/sdk to handle the JSON-RPC protocol,
 * and converts MCP tool definitions into the Anthropic API tool format so they
 * can be sent directly to Claude.
 */

import * as path from "node:path";
import { existsSync, readFileSync } from "node:fs";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import {
  getDefaultEnvironment,
  StdioClientTransport,
} from "@modelcontextprotocol/sdk/client/stdio.js";
import type Anthropic from "@anthropic-ai/sdk";
import type OpenAI from "openai";
import { parseDotEnv } from "./environment.js";

// ── Types ─────────────────────────────────────────────────────────────

export interface MCPTool {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
}

export interface ToolCallResult {
  content: string;
  isError: boolean;
}

interface MCPToolPage {
  tools: Array<{
    name: string;
    description?: string;
    inputSchema: Record<string, unknown>;
  }>;
  nextCursor?: string;
}

export async function collectToolPages(
  fetchPage: (cursor?: string) => Promise<MCPToolPage>,
): Promise<MCPTool[]> {
  const tools: MCPTool[] = [];
  const seenCursors = new Set<string>();
  let cursor: string | undefined;

  do {
    const result = await fetchPage(cursor);
    tools.push(...result.tools.map((tool) => ({
      name: tool.name,
      description: tool.description ?? "",
      inputSchema: tool.inputSchema,
    })));

    cursor = result.nextCursor;
    if (cursor) {
      if (seenCursors.has(cursor)) {
        throw new Error(`MCP server repeated tools/list cursor ${cursor}`);
      }
      seenCursors.add(cursor);
    }
  } while (cursor);

  return tools;
}

const SERVER_ENV_PREFIXES = [
  "MCP_",
  "RUST_",
  "PYTHON_",
  "TS_",
  "BRIDGE_",
  "PREMIERE_",
  "INTEL_",
  "MEDIA_",
];

export function buildServerEnvironment(
  repositoryRoot: string,
  sourceEnvironment: NodeJS.ProcessEnv = process.env,
): Record<string, string> {
  const environment = getDefaultEnvironment();
  const dotenvPath = path.join(repositoryRoot, ".env");
  if (existsSync(dotenvPath)) {
    for (const [key, value] of parseDotEnv(readFileSync(dotenvPath, "utf-8"))) {
      if (isServerEnvironmentKey(key)) environment[key] = value;
    }
  }
  for (const [key, value] of Object.entries(sourceEnvironment)) {
    if (value !== undefined && isServerEnvironmentKey(key)) {
      environment[key] = value;
    }
  }
  environment["MCP_TOOL_PROFILE"] ??= "standard";
  // This client always speaks MCP over the child's stdio pipes. Repository or
  // shell SSE settings apply to standalone servers, never this subprocess.
  environment["MCP_TRANSPORT"] = "stdio";
  return environment;
}

function isServerEnvironmentKey(key: string): boolean {
  return SERVER_ENV_PREFIXES.some((prefix) => key.startsWith(prefix));
}

// ── MCPClient ─────────────────────────────────────────────────────────

export class MCPClient {
  private client: Client;
  private transport: StdioClientTransport | null = null;
  private tools: MCPTool[] = [];

  constructor() {
    this.client = new Client(
      { name: "premierpro-cli", version: "0.1.0" },
      { capabilities: {} },
    );
  }

  /**
   * Spawn the MCP server binary and establish a connection.
   */
  async connect(): Promise<void> {
    const repositoryRoot = path.resolve(
      import.meta.dirname,
      "..",
      "..",
    );
    const serverPath = path.join(
      repositoryRoot,
      "go-orchestrator",
      "bin",
      "premierpro-mcp",
    );

    this.transport = new StdioClientTransport({
      command: serverPath,
      args: ["--transport", "stdio", "--log-level", "error"],
      env: buildServerEnvironment(repositoryRoot),
      stderr: "ignore",
    });

    await this.client.connect(this.transport);
  }

  /**
   * Fetch all tools from the MCP server and cache them.
   */
  async listTools(): Promise<MCPTool[]> {
    this.tools = await collectToolPages(async (cursor) => {
      const result = await this.client.listTools(cursor ? { cursor } : undefined);
      return {
        tools: result.tools.map((tool) => ({
          name: tool.name,
          description: tool.description,
          inputSchema: tool.inputSchema as Record<string, unknown>,
        })),
        nextCursor: result.nextCursor,
      };
    });

    return this.tools;
  }

  /**
   * Call a tool on the MCP server and return the text result.
   */
  async callTool(
    name: string,
    args: Record<string, unknown>,
  ): Promise<ToolCallResult> {
    const result = await this.client.callTool({ name, arguments: args });

    // MCP tool results contain an array of content blocks.
    // We concatenate all text blocks into a single string.
    const textParts: string[] = [];
    let isError = result.isError === true;

    if (Array.isArray(result.content)) {
      for (const block of result.content) {
        if (
          typeof block === "object" &&
          block !== null &&
          "type" in block &&
          block.type === "text" &&
          "text" in block
        ) {
          textParts.push(block.text as string);
        }
      }
    }

    return {
      content: textParts.join("\n") || "(no output)",
      isError,
    };
  }

  /**
   * Convert MCP tool definitions to the Anthropic API tool format.
   */
  getAnthropicTools(): Anthropic.Messages.Tool[] {
    return this.tools.map((tool) => ({
      name: tool.name,
      description: tool.description,
      input_schema: tool.inputSchema as Anthropic.Messages.Tool["input_schema"],
    }));
  }

  /**
   * Convert MCP tool definitions to the OpenAI function-calling format.
   */
  getOpenAITools(): OpenAI.Chat.Completions.ChatCompletionTool[] {
    return this.tools.map((tool) => ({
      type: "function" as const,
      function: {
        name: tool.name,
        description: tool.description,
        parameters: tool.inputSchema,
      },
    }));
  }

  /**
   * Return the cached tool count.
   */
  getToolCount(): number {
    return this.tools.length;
  }

  /**
   * Cleanly disconnect from the MCP server.
   */
  async disconnect(): Promise<void> {
    try {
      await this.client.close();
    } catch {
      // Ignore errors during shutdown
    }
    try {
      await this.transport?.close();
    } catch {
      // Ignore errors during shutdown
    }
  }
}
