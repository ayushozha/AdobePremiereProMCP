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
import { banner, printAssistant, printError, printInfo, createReadlineInterface, prompt, color, } from "./ui.js";
// ── Main ──────────────────────────────────────────────────────────────
async function main() {
    banner();
    // 1. Check for API key
    if (!process.env.ANTHROPIC_API_KEY) {
        printError("ANTHROPIC_API_KEY environment variable is not set.");
        console.log();
        console.log("  Set it with:");
        console.log(`    ${color.cyan}export ANTHROPIC_API_KEY="sk-ant-..."${color.reset}`);
        console.log();
        process.exit(1);
    }
    // 2. Spawn and connect to the MCP server
    const mcpClient = new MCPClient();
    printInfo("  Connecting to PremierPro MCP server...");
    try {
        await mcpClient.connect();
    }
    catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        printError(`Failed to start MCP server: ${msg}`);
        console.log();
        console.log("  Make sure the server binary exists at:");
        console.log(`    ${color.cyan}go-orchestrator/bin/premierpro-mcp${color.reset}`);
        console.log();
        console.log("  Build it with:");
        console.log(`    ${color.cyan}cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server${color.reset}`);
        console.log();
        process.exit(1);
    }
    // 3. Fetch available tools
    let toolCount;
    try {
        const tools = await mcpClient.listTools();
        toolCount = tools.length;
    }
    catch (err) {
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
    const shutdown = async () => {
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
        if (trimmed.toLowerCase() === "exit" ||
            trimmed.toLowerCase() === "quit" ||
            trimmed.toLowerCase() === "q") {
            await shutdown();
            break;
        }
        // Process the message through Claude
        try {
            const response = await chatLoop.processUserMessage(trimmed);
            if (response) {
                printAssistant(response);
            }
        }
        catch (err) {
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
//# sourceMappingURL=index.js.map