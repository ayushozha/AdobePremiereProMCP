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
// ── MCPClient ─────────────────────────────────────────────────────────
export class MCPClient {
    client;
    transport = null;
    tools = [];
    constructor() {
        this.client = new Client({ name: "premierpro-cli", version: "0.1.0" }, { capabilities: {} });
    }
    /**
     * Spawn the MCP server binary and establish a connection.
     */
    async connect() {
        const serverPath = path.resolve(import.meta.dirname, "..", "..", "go-orchestrator", "bin", "premierpro-mcp");
        this.transport = new StdioClientTransport({
            command: serverPath,
            args: [],
            env: {
                ...process.env,
                // Suppress noisy server logs in the CLI's stderr
                PREMIERPRO_LOG_LEVEL: "error",
            },
        });
        await this.client.connect(this.transport);
    }
    /**
     * Fetch all tools from the MCP server and cache them.
     */
    async listTools() {
        const result = await this.client.listTools();
        this.tools = result.tools.map((tool) => ({
            name: tool.name,
            description: tool.description ?? "",
            inputSchema: tool.inputSchema,
        }));
        return this.tools;
    }
    /**
     * Call a tool on the MCP server and return the text result.
     */
    async callTool(name, args) {
        const result = await this.client.callTool({ name, arguments: args });
        // MCP tool results contain an array of content blocks.
        // We concatenate all text blocks into a single string.
        const textParts = [];
        let isError = result.isError === true;
        if (Array.isArray(result.content)) {
            for (const block of result.content) {
                if (typeof block === "object" &&
                    block !== null &&
                    "type" in block &&
                    block.type === "text" &&
                    "text" in block) {
                    textParts.push(block.text);
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
    getAnthropicTools() {
        return this.tools.map((tool) => ({
            name: tool.name,
            description: tool.description,
            input_schema: tool.inputSchema,
        }));
    }
    /**
     * Return the cached tool count.
     */
    getToolCount() {
        return this.tools.length;
    }
    /**
     * Cleanly disconnect from the MCP server.
     */
    async disconnect() {
        try {
            await this.client.close();
        }
        catch {
            // Ignore errors during shutdown
        }
        try {
            await this.transport?.close();
        }
        catch {
            // Ignore errors during shutdown
        }
    }
}
//# sourceMappingURL=mcp-client.js.map