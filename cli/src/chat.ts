/**
 * AI chat loop -- supports both Anthropic (Claude) and OpenAI (GPT/Codex).
 * Sends user messages with MCP tools attached, executes tool calls against
 * the MCP server, and feeds results back until the model produces a final response.
 */

import Anthropic from "@anthropic-ai/sdk";
import OpenAI from "openai";
import type { MCPClient } from "./mcp-client.js";
import type { AuthResult } from "./auth.js";
import { printToolCall, printToolResult, printError, Spinner } from "./ui.js";

// ── Constants ─────────────────────────────────────────────────────────

const TOOL_RESULT_DISPLAY_MAX = 500;
const MAX_RETRIES = 1;
const RETRY_DELAY_MS = 2000;

function buildSystemPrompt(toolCount: number): string {
  return `You are a Premiere Pro video editing assistant running inside an interactive CLI.
You have access to ${toolCount.toLocaleString()} tools to control Adobe Premiere Pro — covering
project management, timeline editing, clip operations, effects, transitions, color grading,
audio mixing, keyframes, markers, export, and more.

Key capabilities:
- Open, create, save, and manage projects
- Create sequences and manage timelines
- Import media, place clips, trim, split, move, and delete clips
- Add transitions (cross dissolve, dip to black, etc.) and video effects
- Adjust Lumetri color: brightness, contrast, saturation, temperature, exposure
- Set keyframes for any effect parameter
- Control audio levels and add audio transitions
- Export sequences in multiple formats (H.264, ProRes, AAF, OMF, FCPXML)
- Full automated editing pipeline (AutoEdit) from script + assets

When the user asks you to do something, use the appropriate tool(s). You may call
multiple tools in sequence if a task requires it.

When reporting results:
- Be concise and conversational.
- Summarize what happened rather than dumping raw JSON.
- If a tool returns an error, explain what went wrong in plain language and suggest next steps.
- If no tool is needed (e.g. the user is just chatting), respond normally.`;
}

// ── Retry helper ──────────────────────────────────────────────────────

async function withRetry<T>(fn: () => Promise<T>, retries: number = MAX_RETRIES): Promise<T> {
  let lastError: unknown;
  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;
      if (attempt < retries) {
        const isRetryable = isRetryableError(err);
        if (!isRetryable) throw err;
        const delay = RETRY_DELAY_MS * (attempt + 1);
        printError(`API request failed, retrying in ${delay / 1000}s...`);
        await sleep(delay);
      }
    }
  }
  throw lastError;
}

function isRetryableError(err: unknown): boolean {
  if (err instanceof Error) {
    const msg = err.message.toLowerCase();
    // Retry on rate limits, server errors, timeouts, and network issues
    if (msg.includes("rate limit") || msg.includes("429")) return true;
    if (msg.includes("500") || msg.includes("502") || msg.includes("503") || msg.includes("504")) return true;
    if (msg.includes("timeout") || msg.includes("econnreset") || msg.includes("econnrefused")) return true;
    if (msg.includes("overloaded")) return true;
  }
  return false;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ── ChatLoop ──────────────────────────────────────────────────────────

export class ChatLoop {
  private mcpClient: MCPClient;
  private auth: AuthResult;
  private systemPrompt: string;

  // Anthropic
  private anthropic?: Anthropic;
  private anthropicHistory: Anthropic.Messages.MessageParam[] = [];

  // OpenAI
  private openai?: OpenAI;
  private openaiHistory: OpenAI.Chat.Completions.ChatCompletionMessageParam[] = [];

  constructor(mcpClient: MCPClient, auth: AuthResult) {
    this.mcpClient = mcpClient;
    this.auth = auth;
    this.systemPrompt = buildSystemPrompt(mcpClient.getToolCount());

    if (auth.provider === "anthropic") {
      this.anthropic = new Anthropic({ apiKey: auth.apiKey });
    } else {
      this.openai = new OpenAI({ apiKey: auth.apiKey });
    }
  }

  async processUserMessage(userText: string): Promise<string> {
    if (this.auth.provider === "anthropic") {
      return this.processAnthropic(userText);
    }
    return this.processOpenAI(userText);
  }

  // ── Anthropic (Claude) ──────────────────────────────────────────────

  private async processAnthropic(userText: string): Promise<string> {
    this.anthropicHistory.push({ role: "user", content: userText });

    let assistantText = "";

    while (true) {
      const spinner = new Spinner("Thinking");
      spinner.start();

      let response: Anthropic.Messages.Message;
      try {
        response = await withRetry(() =>
          this.anthropic!.messages.create({
            model: this.auth.model,
            max_tokens: 4096,
            system: this.systemPrompt,
            tools: this.mcpClient.getAnthropicTools(),
            messages: this.anthropicHistory,
          }),
        );
      } finally {
        spinner.stop();
      }

      const textParts: string[] = [];
      const toolUseBlocks: Anthropic.Messages.ToolUseBlock[] = [];

      for (const block of response.content) {
        if (block.type === "text") textParts.push(block.text);
        else if (block.type === "tool_use") toolUseBlocks.push(block);
      }

      this.anthropicHistory.push({
        role: "assistant",
        content: response.content as Anthropic.Messages.ContentBlockParam[],
      });

      if (toolUseBlocks.length === 0) {
        assistantText = textParts.join("\n");
        break;
      }

      const toolResults: Anthropic.Messages.ToolResultBlockParam[] = [];

      for (const toolUse of toolUseBlocks) {
        const args = (toolUse.input as Record<string, unknown>) ?? {};
        printToolCall(toolUse.name, args);

        const callSpinner = new Spinner(`Running ${toolUse.name}`);
        callSpinner.start();

        let result;
        try {
          result = await this.mcpClient.callTool(toolUse.name, args);
        } catch (err) {
          const elapsedMs = callSpinner.stop();
          const errMsg = err instanceof Error ? err.message : String(err);
          printToolResult(errMsg, true, elapsedMs);
          toolResults.push({
            type: "tool_result",
            tool_use_id: toolUse.id,
            content: errMsg,
            is_error: true,
          });
          continue;
        }

        const elapsedMs = callSpinner.stop();

        // Truncate long results for display, but send full content to model
        const displayContent =
          result.content.length > TOOL_RESULT_DISPLAY_MAX
            ? result.content.slice(0, TOOL_RESULT_DISPLAY_MAX) + "..."
            : result.content;
        printToolResult(displayContent, result.isError, elapsedMs);

        toolResults.push({
          type: "tool_result",
          tool_use_id: toolUse.id,
          content: result.content,
          is_error: result.isError,
        });
      }

      this.anthropicHistory.push({ role: "user", content: toolResults });

      if (response.stop_reason === "end_turn") {
        assistantText = textParts.join("\n");
        break;
      }
    }

    return assistantText;
  }

  // ── OpenAI (GPT / Codex) ────────────────────────────────────────────

  private async processOpenAI(userText: string): Promise<string> {
    this.openaiHistory.push({ role: "user", content: userText });

    let assistantText = "";

    // Convert MCP tools to OpenAI function format
    const tools = this.mcpClient.getOpenAITools();

    while (true) {
      const spinner = new Spinner("Thinking");
      spinner.start();

      let response: OpenAI.Chat.Completions.ChatCompletion;
      try {
        response = await withRetry(() =>
          this.openai!.chat.completions.create({
            model: this.auth.model,
            max_tokens: 4096,
            messages: [
              { role: "system", content: this.systemPrompt },
              ...this.openaiHistory,
            ],
            tools,
          }),
        );
      } finally {
        spinner.stop();
      }

      const choice = response.choices[0];
      if (!choice) break;

      const message = choice.message;

      // Save assistant message to history
      this.openaiHistory.push(message);

      // If no tool calls, we're done
      if (!message.tool_calls || message.tool_calls.length === 0) {
        assistantText = message.content ?? "";
        break;
      }

      // Execute tool calls
      for (const toolCall of message.tool_calls) {
        if (toolCall.type !== "function") continue;
        const name = toolCall.function.name;
        let args: Record<string, unknown> = {};
        try {
          args = JSON.parse(toolCall.function.arguments || "{}");
        } catch {
          // empty args
        }

        printToolCall(name, args);

        const callSpinner = new Spinner(`Running ${name}`);
        callSpinner.start();

        let result;
        try {
          result = await this.mcpClient.callTool(name, args);
        } catch (err) {
          const elapsedMs = callSpinner.stop();
          const errMsg = err instanceof Error ? err.message : String(err);
          printToolResult(errMsg, true, elapsedMs);
          this.openaiHistory.push({
            role: "tool",
            tool_call_id: toolCall.id,
            content: errMsg,
          });
          continue;
        }

        const elapsedMs = callSpinner.stop();

        // Truncate long results for display, but send full content to model
        const displayContent =
          result.content.length > TOOL_RESULT_DISPLAY_MAX
            ? result.content.slice(0, TOOL_RESULT_DISPLAY_MAX) + "..."
            : result.content;
        printToolResult(displayContent, result.isError, elapsedMs);

        this.openaiHistory.push({
          role: "tool",
          tool_call_id: toolCall.id,
          content: result.content,
        });
      }

      if (choice.finish_reason === "stop") {
        assistantText = message.content ?? "";
        break;
      }
    }

    return assistantText;
  }
}
