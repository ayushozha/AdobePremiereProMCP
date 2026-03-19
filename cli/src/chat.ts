/**
 * AI chat loop — supports both Anthropic (Claude) and OpenAI (GPT/Codex).
 * Sends user messages with MCP tools attached, executes tool calls against
 * the MCP server, and feeds results back until the model produces a final response.
 */

import Anthropic from "@anthropic-ai/sdk";
import OpenAI from "openai";
import type { MCPClient } from "./mcp-client.js";
import type { AuthResult } from "./auth.js";
import { printToolCall, printToolResult, Spinner } from "./ui.js";

// ── Constants ─────────────────────────────────────────────────────────

const SYSTEM_PROMPT = `You are a Premiere Pro video editing assistant running inside an interactive CLI.
You have tools to control Adobe Premiere Pro — launching it, inspecting projects,
editing timelines, importing media, adding transitions and text, exporting, and more.

When the user asks you to do something, use the appropriate tool(s). You may call
multiple tools in sequence if a task requires it.

When reporting results:
- Be concise and conversational.
- Summarize what happened rather than dumping raw JSON.
- If a tool returns an error, explain what went wrong in plain language.
- If no tool is needed (e.g. the user is just chatting), respond normally.`;

// ── ChatLoop ──────────────────────────────────────────────────────────

export class ChatLoop {
  private mcpClient: MCPClient;
  private auth: AuthResult;

  // Anthropic
  private anthropic?: Anthropic;
  private anthropicHistory: Anthropic.Messages.MessageParam[] = [];

  // OpenAI
  private openai?: OpenAI;
  private openaiHistory: OpenAI.Chat.Completions.ChatCompletionMessageParam[] = [];

  constructor(mcpClient: MCPClient, auth: AuthResult) {
    this.mcpClient = mcpClient;
    this.auth = auth;

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
        response = await this.anthropic!.messages.create({
          model: this.auth.model,
          max_tokens: 4096,
          system: SYSTEM_PROMPT,
          tools: this.mcpClient.getAnthropicTools(),
          messages: this.anthropicHistory,
        });
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
          callSpinner.stop();
          const errMsg = err instanceof Error ? err.message : String(err);
          printToolResult(errMsg, true);
          toolResults.push({
            type: "tool_result",
            tool_use_id: toolUse.id,
            content: errMsg,
            is_error: true,
          });
          continue;
        }

        callSpinner.stop();
        printToolResult(result.content, result.isError);

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
        response = await this.openai!.chat.completions.create({
          model: this.auth.model,
          max_tokens: 4096,
          messages: [
            { role: "system", content: SYSTEM_PROMPT },
            ...this.openaiHistory,
          ],
          tools,
        });
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
          callSpinner.stop();
          const errMsg = err instanceof Error ? err.message : String(err);
          printToolResult(errMsg, true);
          this.openaiHistory.push({
            role: "tool",
            tool_call_id: toolCall.id,
            content: errMsg,
          });
          continue;
        }

        callSpinner.stop();
        printToolResult(result.content, result.isError);

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
