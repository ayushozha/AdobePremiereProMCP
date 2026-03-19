/**
 * Terminal UI helpers -- colored output, spinners, tool formatting.
 * Uses raw ANSI escape codes so we have zero runtime dependencies.
 */
import * as readline from "node:readline";
export declare const color: {
    readonly reset: "\u001B[0m";
    readonly bold: "\u001B[1m";
    readonly dim: "\u001B[2m";
    readonly italic: "\u001B[3m";
    readonly red: "\u001B[31m";
    readonly green: "\u001B[32m";
    readonly yellow: "\u001B[33m";
    readonly blue: "\u001B[34m";
    readonly magenta: "\u001B[35m";
    readonly cyan: "\u001B[36m";
    readonly white: "\u001B[37m";
    readonly gray: "\u001B[90m";
};
export declare function banner(): void;
export declare function printToolCall(name: string, args: Record<string, unknown>): void;
export declare function printToolResult(result: string, isError: boolean): void;
export declare function printAssistant(text: string): void;
export declare function printError(message: string): void;
export declare function printInfo(message: string): void;
export declare class Spinner {
    private interval;
    private frameIdx;
    private message;
    constructor(message: string);
    start(): void;
    stop(): void;
}
export declare function createReadlineInterface(): readline.Interface;
export declare function prompt(rl: readline.Interface): Promise<string | null>;
//# sourceMappingURL=ui.d.ts.map