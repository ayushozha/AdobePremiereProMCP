/**
 * MCP client that spawns and communicates with the Go MCP server over stdio.
 *
 * Uses the official @modelcontextprotocol/sdk to handle the JSON-RPC protocol,
 * and converts MCP tool definitions into the Anthropic API tool format so they
 * can be sent directly to Claude.
 */

import * as path from "node:path";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";
import type Anthropic from "@anthropic-ai/sdk";
import type OpenAI from "openai";

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
    const serverPath = path.resolve(
      import.meta.dirname,
      "..",
      "..",
      "go-orchestrator",
      "bin",
      "premierpro-mcp",
    );

    this.transport = new StdioClientTransport({
      command: serverPath,
      args: ["--log-level", "error"],
      stderr: "ignore",
    });

    await this.client.connect(this.transport);
  }

  /**
   * Fetch all tools from the MCP server and cache them.
   */
  async listTools(): Promise<MCPTool[]> {
    const result = await this.client.listTools();

    this.tools = result.tools.map((tool) => ({
      name: tool.name,
      description: tool.description ?? "",
      inputSchema: tool.inputSchema as Record<string, unknown>,
    }));

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
