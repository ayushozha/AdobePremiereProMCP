/**
 * AI chat loop — sends user messages to Claude with MCP tools attached,
 * executes tool calls against the MCP server, and feeds results back
 * until Claude produces a final text response.
 */
import type { MCPClient } from "./mcp-client.js";
export declare class ChatLoop {
    private anthropic;
    private mcpClient;
    private history;
    constructor(mcpClient: MCPClient);
    /**
     * Process a single user turn: send to Claude, execute any tool calls,
     * and return the final assistant text.
     */
    processUserMessage(userText: string): Promise<string>;
}
//# sourceMappingURL=chat.d.ts.map