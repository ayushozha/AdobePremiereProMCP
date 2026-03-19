/**
 * MCP client that spawns and communicates with the Go MCP server over stdio.
 *
 * Uses the official @modelcontextprotocol/sdk to handle the JSON-RPC protocol,
 * and converts MCP tool definitions into the Anthropic API tool format so they
 * can be sent directly to Claude.
 */
import type Anthropic from "@anthropic-ai/sdk";
export interface MCPTool {
    name: string;
    description: string;
    inputSchema: Record<string, unknown>;
}
export interface ToolCallResult {
    content: string;
    isError: boolean;
}
export declare class MCPClient {
    private client;
    private transport;
    private tools;
    constructor();
    /**
     * Spawn the MCP server binary and establish a connection.
     */
    connect(): Promise<void>;
    /**
     * Fetch all tools from the MCP server and cache them.
     */
    listTools(): Promise<MCPTool[]>;
    /**
     * Call a tool on the MCP server and return the text result.
     */
    callTool(name: string, args: Record<string, unknown>): Promise<ToolCallResult>;
    /**
     * Convert MCP tool definitions to the Anthropic API tool format.
     */
    getAnthropicTools(): Anthropic.Messages.Tool[];
    /**
     * Return the cached tool count.
     */
    getToolCount(): number;
    /**
     * Cleanly disconnect from the MCP server.
     */
    disconnect(): Promise<void>;
}
//# sourceMappingURL=mcp-client.d.ts.map