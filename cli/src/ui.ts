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
  underline: `${ESC}4m`,

  red: `${ESC}31m`,
  green: `${ESC}32m`,
  yellow: `${ESC}33m`,
  blue: `${ESC}34m`,
  magenta: `${ESC}35m`,
  cyan: `${ESC}36m`,
  white: `${ESC}37m`,
  gray: `${ESC}90m`,

  bgRed: `${ESC}41m`,
  bgGreen: `${ESC}42m`,
  bgYellow: `${ESC}43m`,
  bgBlue: `${ESC}44m`,
  bgMagenta: `${ESC}45m`,
} as const;

// ── Formatted printers ────────────────────────────────────────────────

export function banner(toolCount?: number): void {
  console.log();
  console.log(
    `${color.bold}${color.magenta}    ____                    _              ____            ${color.reset}`,
  );
  console.log(
    `${color.bold}${color.magenta}   / __ \\________  ____ ___(_)__  ______  / __ \\_________  ${color.reset}`,
  );
  console.log(
    `${color.bold}${color.magenta}  / /_/ / ___/ _ \\/ __ \`__ / / _ \\/ ___/ / /_/ / ___/ __ \\ ${color.reset}`,
  );
  console.log(
    `${color.bold}${color.magenta} / ____/ /  /  __/ / / / / /  __/ /    / ____/ /  / /_/ / ${color.reset}`,
  );
  console.log(
    `${color.bold}${color.magenta}/_/   /_/   \\___/_/ /_/ /_/\\___/_/    /_/   /_/   \\____/  ${color.reset}`,
  );
  console.log();
  console.log(
    `${color.bold}${color.cyan}  AI Video Editor${color.reset}  ${color.dim}— powered by MCP${color.reset}`,
  );
  if (toolCount !== undefined && toolCount > 0) {
    console.log(
      `${color.dim}  ${toolCount.toLocaleString()} tools available${color.reset}`,
    );
  }
  console.log(
    `${color.dim}  Type naturally to control Premiere Pro. Type "exit" or Ctrl+C to quit.${color.reset}`,
  );
  console.log();
}

export function printToolCall(name: string, args: Record<string, unknown>): void {
  const argsStr =
    Object.keys(args).length > 0
      ? ` ${color.dim}${JSON.stringify(args)}${color.reset}`
      : "";
  console.log(
    `${color.yellow}${color.bold}  >> ${name}${color.reset}${argsStr}`,
  );
}

export function printToolResult(result: string, isError: boolean, durationMs?: number): void {
  const timingStr =
    durationMs !== undefined
      ? ` ${color.dim}(${formatDuration(durationMs)})${color.reset}`
      : "";

  if (isError) {
    const prefix = `${color.red}${color.bold}  !! Error${color.reset}${timingStr}`;

    // Indent multiline errors
    const indented = result
      .split("\n")
      .map((line, i) => (i === 0 ? line : `           ${line}`))
      .join("\n");

    console.log(`${prefix} ${color.red}${indented}${color.reset}`);

    // Show suggestions for common errors
    const suggestion = getErrorSuggestion(result);
    if (suggestion) {
      console.log(`${color.yellow}     Tip: ${suggestion}${color.reset}`);
    }
  } else {
    const prefix = `${color.green}  << OK${color.reset}${timingStr}`;

    // Truncate very long results for display
    const maxLen = 500;
    const display = result.length > maxLen ? result.slice(0, maxLen) + "..." : result;

    // Indent multiline results
    const indented = display
      .split("\n")
      .map((line, i) => (i === 0 ? line : `           ${line}`))
      .join("\n");

    console.log(`${prefix} ${color.dim}${indented}${color.reset}`);
  }
}

export function printAssistant(text: string): void {
  console.log();
  console.log(`${color.green}${color.bold}assistant>${color.reset} ${text}`);
  console.log();
}

export function printError(message: string): void {
  console.error(`${color.red}${color.bold}error:${color.reset} ${message}`);
}

export function printInfo(message: string): void {
  console.log(`${color.dim}${message}${color.reset}`);
}

export function printSuccess(message: string): void {
  console.log(`${color.green}${message}${color.reset}`);
}

// ── Error suggestions ───────────────────────────────────────────────

function getErrorSuggestion(errorMessage: string): string | null {
  const lower = errorMessage.toLowerCase();

  if (lower.includes("no project") || lower.includes("project is not open") || lower.includes("project not found")) {
    return 'Open a project first. Try: "open my project at /path/to/project.prproj"';
  }
  if (lower.includes("no sequence") || lower.includes("no active sequence") || lower.includes("sequence not found")) {
    return 'Create or select a sequence first. Try: "create a new 1080p sequence called MyEdit"';
  }
  if (lower.includes("premiere") && (lower.includes("not running") || lower.includes("not responding"))) {
    return 'Premiere Pro may need to be restarted. Try: "launch premiere pro"';
  }
  if (lower.includes("file not found") || lower.includes("no such file")) {
    return "Double-check the file path. Use an absolute path (e.g. /Users/you/Videos/clip.mp4).";
  }
  if (lower.includes("timeout") || lower.includes("timed out")) {
    return "The operation took too long. Premiere Pro may be busy rendering. Try again in a moment.";
  }
  if (lower.includes("permission") || lower.includes("access denied")) {
    return "Check file/folder permissions. You may need to grant Premiere Pro access in System Settings.";
  }

  return null;
}

// ── Duration formatting ─────────────────────────────────────────────

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`;
  const minutes = Math.floor(ms / 60_000);
  const seconds = ((ms % 60_000) / 1000).toFixed(0);
  return `${minutes}m${seconds}s`;
}

// ── Spinner ───────────────────────────────────────────────────────────

const SPINNER_FRAMES = [
  "\u280B", // braille dots
  "\u2819",
  "\u2839",
  "\u2838",
  "\u283C",
  "\u2834",
  "\u2826",
  "\u2827",
  "\u2807",
  "\u280F",
];

export class Spinner {
  private interval: ReturnType<typeof setInterval> | null = null;
  private frameIdx = 0;
  private message: string;
  private startTime: number = 0;

  constructor(message: string) {
    this.message = message;
  }

  start(): void {
    this.frameIdx = 0;
    this.startTime = Date.now();
    process.stdout.write(
      `${color.cyan}  ${SPINNER_FRAMES[0]} ${this.message}${color.reset}`,
    );
    this.interval = setInterval(() => {
      this.frameIdx = (this.frameIdx + 1) % SPINNER_FRAMES.length;
      const elapsed = formatDuration(Date.now() - this.startTime);
      // Clear line and rewrite
      process.stdout.write(
        `\r${color.cyan}  ${SPINNER_FRAMES[this.frameIdx]} ${this.message} ${color.dim}${elapsed}${color.reset}\x1b[K`,
      );
    }, 80);
  }

  /** Stop the spinner and return elapsed time in ms. */
  stop(): number {
    const elapsed = Date.now() - this.startTime;
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
    // Clear the spinner line
    process.stdout.write("\r\x1b[K");
    return elapsed;
  }
}

// ── Readline interface ────────────────────────────────────────────────

export function createReadlineInterface(): readline.Interface {
  return readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: true,
  });
}

export function prompt(rl: readline.Interface): Promise<string | null> {
  return new Promise((resolve) => {
    rl.question(`${color.blue}${color.bold}you>${color.reset} `, (answer) => {
      resolve(answer);
    });
    // Handle close (Ctrl+D)
    rl.once("close", () => resolve(null));
  });
}
