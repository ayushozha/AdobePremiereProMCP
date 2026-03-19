/**
 * Standalone Premiere Pro bridge — fallback mode.
 *
 * Controls Adobe Premiere Pro from OUTSIDE the application by sending
 * ExtendScript via macOS `osascript` (AppleScript's DoScript command).
 * This is used when the CEP panel is not installed or not running.
 *
 * Transport:
 *   osascript -e 'tell application "Adobe Premiere Pro 2025" to DoScript "..."'
 *
 * Each method:
 *   1. Builds an ExtendScript string using the templates module.
 *   2. Wraps it in an AppleScript DoScript call.
 *   3. Executes it via child_process.execSync.
 *   4. Parses the JSON result.
 *   5. Throws a typed error on failure.
 */

import { execSync, type ExecSyncOptionsWithStringEncoding } from "node:child_process";
import { createLogger, format, transports, type Logger } from "winston";

import type { BridgeConfig } from "../config.js";
import type {
  PremiereBridge,
  ProjectState,
  TimelineState,
  ExportResult,
  EDLExecutionResult,
  PingResult,
  EvalCommandResult,
  Resolution,
  TrackTarget,
  Timecode,
  TimeRange,
  TextStyle,
  EffectParams,
  EditDecisionList,
  ExportPreset,
} from "./interface.js";
import * as ES from "./extendscript-templates.js";

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

export class BridgeError extends Error {
  constructor(
    message: string,
    public readonly command: string,
    public readonly cause?: unknown,
  ) {
    super(message);
    this.name = "BridgeError";
  }
}

export class PremiereNotRunningError extends BridgeError {
  constructor(cause?: unknown) {
    super(
      "Adobe Premiere Pro is not running or did not respond.",
      "ping",
      cause,
    );
    this.name = "PremiereNotRunningError";
  }
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const DEFAULT_TIMEOUT_MS = 30_000;
const MAX_OUTPUT_BUFFER = 10 * 1024 * 1024; // 10 MB

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

export class StandaloneBridge implements PremiereBridge {
  private readonly log: Logger;
  private readonly premierePath: string;
  private readonly timeoutMs: number;
  private connected = false;

  constructor(config: BridgeConfig) {
    this.premierePath = config.premierePath;
    this.timeoutMs = DEFAULT_TIMEOUT_MS;

    this.log = createLogger({
      level: config.logLevel,
      format: format.combine(
        format.timestamp(),
        format.printf(({ timestamp, level, message }) =>
          `${timestamp as string} [standalone] ${level}: ${message as string}`,
        ),
      ),
      transports: [new transports.Console()],
    });
  }

  // -----------------------------------------------------------------------
  // Lifecycle
  // -----------------------------------------------------------------------

  async connect(): Promise<void> {
    this.log.info("Standalone bridge: verifying Premiere Pro connectivity...");
    try {
      const result = await this.ping();
      if (!result.premiereRunning) {
        this.connected = false;
        this.log.warn(
          "Premiere Pro is not running. Bridge will operate in disconnected mode.",
        );
        return;
      }
      this.connected = true;
      this.log.info(
        `Connected to Premiere Pro ${result.premiereVersion} (standalone mode)`,
      );
    } catch (err) {
      this.connected = false;
      const detail = err instanceof Error ? err.message : String(err);
      this.log.warn(
        `Could not reach Premiere Pro: ${detail}. Bridge will operate in disconnected mode.`,
      );
    }
  }

  async disconnect(): Promise<void> {
    this.connected = false;
    this.log.info("Standalone bridge disconnected.");
  }

  // -----------------------------------------------------------------------
  // Project
  // -----------------------------------------------------------------------

  async getProjectState(): Promise<ProjectState> {
    const script = ES.getProjectState();
    return this.execute<ProjectState>(script, "getProjectState");
  }

  // -----------------------------------------------------------------------
  // Sequence
  // -----------------------------------------------------------------------

  async createSequence(params: {
    name: string;
    resolution: Resolution;
    frameRate: number;
    videoTracks: number;
    audioTracks: number;
  }): Promise<{ sequenceId: string; name: string }> {
    const script = ES.createSequence(params);
    return this.execute<{ sequenceId: string; name: string }>(
      script,
      "createSequence",
    );
  }

  async getTimelineState(sequenceId: string): Promise<TimelineState> {
    const script = ES.getTimelineState(sequenceId);
    return this.execute<TimelineState>(script, "getTimelineState");
  }

  // -----------------------------------------------------------------------
  // Clip operations
  // -----------------------------------------------------------------------

  async importMedia(params: {
    filePath: string;
    targetBin: string;
  }): Promise<{ projectItemId: string; name: string }> {
    const script = ES.importMedia(params.filePath, params.targetBin);
    return this.execute<{ projectItemId: string; name: string }>(
      script,
      "importMedia",
    );
  }

  async placeClip(params: {
    sourcePath: string;
    track: TrackTarget;
    position: Timecode;
    sourceRange?: TimeRange;
    speed: number;
  }): Promise<{ clipId: string }> {
    const script = ES.placeClip(params);
    return this.execute<{ clipId: string }>(script, "placeClip");
  }

  async removeClip(params: {
    clipId: string;
    sequenceId: string;
  }): Promise<void> {
    const script = ES.removeClip(params.clipId, params.sequenceId);
    await this.execute(script, "removeClip");
  }

  // -----------------------------------------------------------------------
  // Effects & transitions
  // -----------------------------------------------------------------------

  async addTransition(params: {
    sequenceId: string;
    track: TrackTarget;
    position: Timecode;
    transitionType: string;
    durationSeconds: number;
  }): Promise<{ transitionId: string }> {
    const script = ES.addTransition(params);
    return this.execute<{ transitionId: string }>(script, "addTransition");
  }

  async addText(params: {
    sequenceId: string;
    text: string;
    style: TextStyle;
    track: TrackTarget;
    position: Timecode;
    durationSeconds: number;
  }): Promise<{ clipId: string }> {
    const script = ES.addText(params);
    return this.execute<{ clipId: string }>(script, "addText");
  }

  async applyEffect(params: {
    clipId: string;
    sequenceId: string;
    effect: EffectParams;
  }): Promise<void> {
    const script = ES.applyEffect(
      params.clipId,
      params.sequenceId,
      params.effect,
    );
    await this.execute(script, "applyEffect");
  }

  // -----------------------------------------------------------------------
  // Audio
  // -----------------------------------------------------------------------

  async setAudioLevel(params: {
    clipId: string;
    sequenceId: string;
    levelDb: number;
  }): Promise<void> {
    const script = ES.setAudioLevel(
      params.clipId,
      params.sequenceId,
      params.levelDb,
    );
    await this.execute(script, "setAudioLevel");
  }

  // -----------------------------------------------------------------------
  // Export
  // -----------------------------------------------------------------------

  async exportSequence(params: {
    sequenceId: string;
    outputPath: string;
    preset: ExportPreset;
  }): Promise<ExportResult> {
    const script = ES.exportSequence(params);
    return this.execute<ExportResult>(script, "exportSequence");
  }

  // -----------------------------------------------------------------------
  // Batch
  // -----------------------------------------------------------------------

  async executeEDL(params: {
    edl: EditDecisionList;
    autoImport: boolean;
    autoCreateSequence: boolean;
  }): Promise<EDLExecutionResult> {
    const script = ES.executeEDL(params.edl);
    return this.execute<EDLExecutionResult>(script, "executeEDL");
  }

  // -----------------------------------------------------------------------
  // Generic Command
  // -----------------------------------------------------------------------

  async evalCommand(functionName: string, argsJson: string): Promise<EvalCommandResult> {
    try {
      // Build ExtendScript: functionName('argsJson') or functionName()
      let script: string;
      if (argsJson && argsJson !== "{}" && argsJson !== "[]") {
        // Escape the JSON string for embedding inside a single-quoted ExtendScript string.
        const escapedArgs = argsJson
          .replace(/\\/g, "\\\\")
          .replace(/'/g, "\\'")
          .replace(/"/g, '\\"')
          .replace(/\n/g, "\\n")
          .replace(/\r/g, "\\r");
        script = `(function(){ try { var r = ${functionName}('${escapedArgs}'); return JSON.stringify({result: r}); } catch(e) { return JSON.stringify({error: e.message || String(e)}); } })()`;
      } else {
        script = `(function(){ try { var r = ${functionName}(); return JSON.stringify({result: r}); } catch(e) { return JSON.stringify({error: e.message || String(e)}); } })()`;
      }

      const raw = this.execute<Record<string, unknown>>(script, `evalCommand:${functionName}`);

      if (typeof raw === "object" && raw !== null && typeof raw["error"] === "string") {
        return {
          resultJson: "",
          isError: true,
          errorMessage: raw["error"] as string,
        };
      }

      const result = typeof raw === "object" && raw !== null && "result" in raw
        ? raw["result"]
        : raw;

      return {
        resultJson: typeof result === "string" ? result : JSON.stringify(result),
        isError: false,
        errorMessage: "",
      };
    } catch (err) {
      return {
        resultJson: "",
        isError: true,
        errorMessage: err instanceof Error ? err.message : String(err),
      };
    }
  }

  // -----------------------------------------------------------------------
  // Health
  // -----------------------------------------------------------------------

  async ping(): Promise<PingResult> {
    const script = ES.ping();
    try {
      return this.execute<PingResult>(script, "ping");
    } catch {
      return {
        premiereRunning: false,
        premiereVersion: "unknown",
        projectOpen: false,
        bridgeMode: "standalone",
      };
    }
  }

  // -----------------------------------------------------------------------
  // Private: execution engine
  // -----------------------------------------------------------------------

  /**
   * Send an ExtendScript string to Premiere Pro via `osascript` and parse
   * the JSON response.
   *
   * The flow:
   *   Node -> osascript -> AppleScript -> Premiere Pro DoScript -> ExtendScript
   *   ExtendScript returns a JSON string <- AppleScript <- osascript <- Node
   */
  private execute<T = unknown>(script: string, command: string): T {
    this.log.debug(`Executing command: ${command}`);

    const appName = this.resolveAppName();
    const escapedScript = this.escapeForAppleScript(script);

    // Build the AppleScript:
    //   tell application "Adobe Premiere Pro 2025" to DoScript "..."
    const appleScript =
      `tell application "${appName}" to DoScript "${escapedScript}"`;

    // Wrap in a shell-safe osascript invocation
    const shellCommand =
      `osascript -e ${this.shellQuote(appleScript)}`;

    const execOpts: ExecSyncOptionsWithStringEncoding = {
      encoding: "utf-8",
      timeout: this.timeoutMs,
      maxBuffer: MAX_OUTPUT_BUFFER,
      windowsHide: true,
    };

    let rawOutput: string;
    try {
      rawOutput = execSync(shellCommand, execOpts).trim();
    } catch (err: unknown) {
      const detail = err instanceof Error ? err.message : String(err);
      this.log.error(`osascript failed for ${command}: ${detail}`);
      throw new BridgeError(
        `Failed to execute "${command}" via osascript: ${detail}`,
        command,
        err,
      );
    }

    if (this.log.isDebugEnabled?.() ?? false) {
      this.log.debug(
        `Raw output [${command}]: ${rawOutput.slice(0, 500)}${rawOutput.length > 500 ? "..." : ""}`,
      );
    }

    // Parse the JSON string returned by the ExtendScript IIFE.
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(rawOutput) as Record<string, unknown>;
    } catch {
      throw new BridgeError(
        `"${command}" returned non-JSON output: ${rawOutput.slice(0, 200)}`,
        command,
      );
    }

    // ExtendScript error convention: { error: "message" }
    if (typeof parsed["error"] === "string") {
      throw new BridgeError(
        `Premiere Pro error in "${command}": ${parsed["error"]}`,
        command,
      );
    }

    return parsed as T;
  }

  // -----------------------------------------------------------------------
  // Helpers
  // -----------------------------------------------------------------------

  /**
   * Derive the macOS application name from the configured path.
   *
   * Example:
   *   "/Applications/Adobe Premiere Pro 2025/Adobe Premiere Pro 2025.app"
   *    -> "Adobe Premiere Pro 2025"
   */
  private resolveAppName(): string {
    const match = this.premierePath.match(/([^/]+)\.app$/);
    if (match?.[1]) {
      return match[1];
    }
    return "Adobe Premiere Pro 2025";
  }

  /**
   * Escape a string so it can be embedded inside double-quoted AppleScript
   * string literals.
   */
  private escapeForAppleScript(s: string): string {
    return s
      .replace(/\\/g, "\\\\")
      .replace(/"/g, '\\"')
      .replace(/\n/g, "\\n")
      .replace(/\r/g, "\\r")
      .replace(/\t/g, "\\t");
  }

  /**
   * Safely quote a string for use as a single shell argument.
   * Uses the $'...' quoting form which handles internal single quotes.
   */
  private shellQuote(s: string): string {
    // Simple approach: wrap in single quotes and escape internal single quotes
    // using the '\'' idiom (end quote, escaped quote, start quote).
    return "'" + s.replace(/'/g, "'\\''") + "'";
  }
}
