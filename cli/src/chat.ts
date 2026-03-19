/**
 * AI chat loop — sends user messages to Claude with MCP tools attached,
 * executes tool calls against the MCP server, and feeds results back
 * until Claude produces a final text response.
 */

import Anthropic from "@anthropic-ai/sdk";
import type { MCPClient } from "./mcp-client.js";
import { printToolCall, printToolResult, printAssistant, Spinner } from "./ui.js";

// ── Constants ─────────────────────────────────────────────────────────

const MODEL = "claude-sonnet-4-20250514";

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

// ── Types ─────────────────────────────────────────────────────────────

type MessageParam = Anthropic.Messages.MessageParam;
type ContentBlockParam = Anthropic.Messages.ContentBlockParam;
type ToolResultBlockParam = Anthropic.Messages.ToolResultBlockParam;

// ── ChatLoop ──────────────────────────────────────────────────────────

export class ChatLoop {
  private anthropic: Anthropic;
  private mcpClient: MCPClient;
  private history: MessageParam[] = [];

  constructor(mcpClient: MCPClient) {
    this.anthropic = new Anthropic();
    this.mcpClient = mcpClient;
  }

  /**
   * Process a single user turn: send to Claude, execute any tool calls,
   * and return the final assistant text.
   */
  async processUserMessage(userText: string): Promise<string> {
    // Add user message to history
    this.history.push({ role: "user", content: userText });

    // Run the agent loop: Claude may respond with tool calls that we need
    // to execute, then feed the results back, possibly multiple times.
    let assistantText = "";

    // eslint-disable-next-line no-constant-condition
    while (true) {
      const spinner = new Spinner("Thinking");
      spinner.start();

      let response: Anthropic.Messages.Message;
      try {
        response = await this.anthropic.messages.create({
          model: MODEL,
          max_tokens: 4096,
          system: SYSTEM_PROMPT,
          tools: this.mcpClient.getAnthropicTools(),
          messages: this.history,
        });
      } finally {
        spinner.stop();
      }

      // Collect text blocks for the final response and tool_use blocks to execute
      const textParts: string[] = [];
      const toolUseBlocks: Anthropic.Messages.ToolUseBlock[] = [];

      for (const block of response.content) {
        if (block.type === "text") {
          textParts.push(block.text);
        } else if (block.type === "tool_use") {
          toolUseBlocks.push(block);
        }
      }

      // Save assistant message to history (include all content blocks)
      this.history.push({
        role: "assistant",
        content: response.content as ContentBlockParam[],
      });

      // If there are no tool calls, we are done
      if (toolUseBlocks.length === 0) {
        assistantText = textParts.join("\n");
        break;
      }

      // Execute each tool call and collect results
      const toolResults: ToolResultBlockParam[] = [];

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
          const errMsg =
            err instanceof Error ? err.message : String(err);
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

      // Feed tool results back as a user message so Claude can continue
      this.history.push({ role: "user", content: toolResults });

      // If stop reason is "end_turn", Claude is done even though there
      // were tool calls (it gave text + tools). In practice this rarely
      // happens; the SDK returns "tool_use" as stop_reason when tools
      // are present. But handle it for completeness.
      if (response.stop_reason === "end_turn") {
        assistantText = textParts.join("\n");
        break;
      }

      // Otherwise loop back: Claude will see the tool results and either
      // make more tool calls or produce a final text response.
    }

    return assistantText;
  }
}
