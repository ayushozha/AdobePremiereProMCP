/**
 * Terminal UI helpers -- colored output, spinners, tool formatting.
 * Uses raw ANSI escape codes so we have zero runtime dependencies.
 */
import * as readline from "node:readline";
// ── ANSI color codes ──────────────────────────────────────────────────
const ESC = "\x1b[";
export const color = {
    reset: `${ESC}0m`,
    bold: `${ESC}1m`,
    dim: `${ESC}2m`,
    italic: `${ESC}3m`,
    red: `${ESC}31m`,
    green: `${ESC}32m`,
    yellow: `${ESC}33m`,
    blue: `${ESC}34m`,
    magenta: `${ESC}35m`,
    cyan: `${ESC}36m`,
    white: `${ESC}37m`,
    gray: `${ESC}90m`,
};
// ── Formatted printers ────────────────────────────────────────────────
export function banner() {
    console.log();
    console.log(`${color.bold}${color.magenta}  PremierPro AI Editor${color.reset}`);
    console.log(`${color.dim}  Type naturally to control Premiere Pro. Type "exit" or Ctrl+C to quit.${color.reset}`);
    console.log();
}
export function printToolCall(name, args) {
    const argsStr = Object.keys(args).length > 0
        ? ` ${color.gray}${JSON.stringify(args)}${color.reset}`
        : "";
    console.log(`${color.cyan}${color.bold}  > Calling${color.reset} ${color.yellow}${name}${color.reset}${argsStr}`);
}
export function printToolResult(result, isError) {
    const prefix = isError
        ? `${color.red}  ! Error:${color.reset}`
        : `${color.green}  < Result:${color.reset}`;
    // Truncate very long results for display
    const maxLen = 1500;
    const display = result.length > maxLen ? result.slice(0, maxLen) + "..." : result;
    // Indent multiline results
    const indented = display
        .split("\n")
        .map((line, i) => (i === 0 ? line : `           ${line}`))
        .join("\n");
    console.log(`${prefix} ${color.dim}${indented}${color.reset}`);
}
export function printAssistant(text) {
    console.log();
    console.log(`${color.green}${color.bold}assistant>${color.reset} ${text}`);
    console.log();
}
export function printError(message) {
    console.error(`${color.red}${color.bold}error:${color.reset} ${message}`);
}
export function printInfo(message) {
    console.log(`${color.dim}${message}${color.reset}`);
}
// ── Spinner ───────────────────────────────────────────────────────────
const SPINNER_FRAMES = [".", "..", "...", "....", "...."];
export class Spinner {
    interval = null;
    frameIdx = 0;
    message;
    constructor(message) {
        this.message = message;
    }
    start() {
        this.frameIdx = 0;
        process.stdout.write(`${color.dim}  ${this.message}${SPINNER_FRAMES[0]}${color.reset}`);
        this.interval = setInterval(() => {
            this.frameIdx = (this.frameIdx + 1) % SPINNER_FRAMES.length;
            // Clear line and rewrite
            process.stdout.write(`\r${color.dim}  ${this.message}${SPINNER_FRAMES[this.frameIdx]}${color.reset}`);
        }, 300);
    }
    stop() {
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
        // Clear the spinner line
        process.stdout.write("\r\x1b[K");
    }
}
// ── Readline interface ────────────────────────────────────────────────
export function createReadlineInterface() {
    return readline.createInterface({
        input: process.stdin,
        output: process.stdout,
        terminal: true,
    });
}
export function prompt(rl) {
    return new Promise((resolve) => {
        rl.question(`${color.blue}${color.bold}you>${color.reset} `, (answer) => {
            resolve(answer);
        });
        // Handle close (Ctrl+D)
        rl.once("close", () => resolve(null));
    });
}
//# sourceMappingURL=ui.js.map