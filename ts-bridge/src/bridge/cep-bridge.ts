/**
 * CEP bridge — communicates with Adobe Premiere Pro through the CEP panel
 * that exposes a WebSocket server inside the host application.
 *
 * Protocol:
 *   Client sends:  { action: string, params: Record<string, unknown>, requestId: string }
 *   Server replies: { requestId: string, result?: unknown, error?: string }
 *
 * The CEP panel runs inside Premiere Pro's embedded Chromium (CEF) runtime
 * and has direct access to the ExtendScript DOM. This bridge merely relays
 * commands over WebSocket and correlates request/response pairs via
 * unique request IDs.
 */

import { randomUUID } from "node:crypto";
import WebSocket from "ws";
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

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

export class CepConnectionError extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message);
    this.name = "CepConnectionError";
  }
}

export class CepTimeoutError extends Error {
  constructor(action: string, timeoutMs: number) {
    super(`CEP command "${action}" timed out after ${timeoutMs}ms`);
    this.name = "CepTimeoutError";
  }
}

export class CepCommandError extends Error {
  constructor(action: string, detail: string) {
    super(`CEP command "${action}" failed: ${detail}`);
    this.name = "CepCommandError";
  }
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** Outgoing message to the CEP panel. */
interface CepRequest {
  action: string;
  params: Record<string, unknown>;
  requestId: string;
}

/** Incoming message from the CEP panel. */
interface CepResponse {
  requestId: string;
  result?: unknown;
  error?: string;
}

/** Pending request awaiting its response. */
interface PendingRequest {
  resolve: (value: unknown) => void;
  reject: (reason: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const DEFAULT_WS_PORT = 9801;
const DEFAULT_COMMAND_TIMEOUT_MS = 30_000;
const DEFAULT_MAX_RECONNECT_ATTEMPTS = 5;
const DEFAULT_RECONNECT_INTERVAL_MS = 2_000;
const HEARTBEAT_INTERVAL_MS = 10_000;

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

export class CepBridge implements PremiereBridge {
  private readonly log: Logger;
  private readonly wsUrl: string;
  private readonly commandTimeoutMs: number;
  private readonly maxReconnectAttempts: number;
  private readonly reconnectIntervalMs: number;

  private ws: WebSocket | null = null;
  private pendingRequests = new Map<string, PendingRequest>();
  private reconnectAttempts = 0;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private intentionalClose = false;
  private _isConnected = false;

  constructor(config: BridgeConfig) {
    const port = config.cepWsPort || DEFAULT_WS_PORT;
    this.wsUrl = `ws://localhost:${port}`;
    this.commandTimeoutMs = DEFAULT_COMMAND_TIMEOUT_MS;
    this.maxReconnectAttempts = DEFAULT_MAX_RECONNECT_ATTEMPTS;
    this.reconnectIntervalMs = DEFAULT_RECONNECT_INTERVAL_MS;

    this.log = createLogger({
      level: config.logLevel,
      format: format.combine(
        format.timestamp(),
        format.printf(({ timestamp, level, message }) =>
          `${timestamp as string} [cep] ${level}: ${message as string}`,
        ),
      ),
      transports: [new transports.Console()],
    });
  }

  // -----------------------------------------------------------------------
  // Lifecycle
  // -----------------------------------------------------------------------

  async connect(): Promise<void> {
    if (this._isConnected && this.ws?.readyState === WebSocket.OPEN) {
      this.log.debug("Already connected to CEP panel.");
      return;
    }

    this.intentionalClose = false;
    this.reconnectAttempts = 0;

    return new Promise<void>((resolve) => {
      this.log.info(`Connecting to CEP panel at ${this.wsUrl}...`);

      const ws = new WebSocket(this.wsUrl);
      let settled = false;
      // Track whether the socket ever successfully opened so we only
      // auto-reconnect on connections that were previously established.
      let wasOpen = false;

      const connectionTimeout = setTimeout(() => {
        if (!settled) {
          settled = true;
          ws.terminate();
          this.log.warn(
            `Connection to CEP panel at ${this.wsUrl} timed out. ` +
              "Bridge will operate in disconnected mode.",
          );
          resolve();
        }
      }, this.commandTimeoutMs);

      ws.on("open", () => {
        if (settled) return;
        settled = true;
        wasOpen = true;
        clearTimeout(connectionTimeout);

        this.ws = ws;
        this._isConnected = true;
        this.reconnectAttempts = 0;
        this.startHeartbeat();

        this.log.info(`Connected to CEP panel at ${this.wsUrl}`);
        resolve();
      });

      ws.on("message", (data: WebSocket.Data) => {
        this.handleMessage(data);
      });

      ws.on("close", (code: number, reason: Buffer) => {
        this._isConnected = false;
        this.stopHeartbeat();

        if (wasOpen) {
          // Only log and reconnect if the socket was previously established
          this.log.warn(
            `WebSocket closed: code=${code} reason=${reason.toString("utf-8")}`,
          );
          this.rejectAllPending("WebSocket connection closed");

          if (!this.intentionalClose) {
            this.attemptReconnect();
          }
        }
        // If the socket was never opened (initial connection failure),
        // the error handler already resolved the promise.
      });

      ws.on("error", (err: Error) => {
        if (!settled) {
          settled = true;
          clearTimeout(connectionTimeout);
          this.log.warn(
            `Could not connect to CEP panel: ${err.message}. ` +
              "Bridge will operate in disconnected mode.",
          );
          resolve();
        } else {
          this.log.error(`WebSocket error: ${err.message}`);
        }
      });
    });
  }

  async disconnect(): Promise<void> {
    this.intentionalClose = true;
    this.stopHeartbeat();
    this.rejectAllPending("Bridge disconnecting");

    if (this.ws) {
      this.ws.close(1000, "Client disconnect");
      this.ws = null;
    }

    this._isConnected = false;
    this.log.info("CEP bridge disconnected.");
  }

  isConnected(): boolean {
    return this._isConnected && this.ws?.readyState === WebSocket.OPEN;
  }

  // -----------------------------------------------------------------------
  // Project
  // -----------------------------------------------------------------------

  async getProjectState(): Promise<ProjectState> {
    return this.send<ProjectState>("getProjectState", {});
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
    return this.send<{ sequenceId: string; name: string }>(
      "createSequence",
      params,
    );
  }

  async getTimelineState(sequenceId: string): Promise<TimelineState> {
    return this.send<TimelineState>("getTimelineState", { sequenceId });
  }

  // -----------------------------------------------------------------------
  // Clip operations
  // -----------------------------------------------------------------------

  async importMedia(params: {
    filePath: string;
    targetBin: string;
  }): Promise<{ projectItemId: string; name: string }> {
    return this.send<{ projectItemId: string; name: string }>(
      "importMedia",
      params,
    );
  }

  async placeClip(params: {
    sourcePath: string;
    track: TrackTarget;
    position: Timecode;
    sourceRange?: TimeRange;
    speed: number;
  }): Promise<{ clipId: string }> {
    return this.send<{ clipId: string }>("placeClip", params);
  }

  async removeClip(params: {
    clipId: string;
    sequenceId: string;
  }): Promise<void> {
    await this.send("removeClip", params);
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
    return this.send<{ transitionId: string }>("addTransition", params);
  }

  async addText(params: {
    sequenceId: string;
    text: string;
    style: TextStyle;
    track: TrackTarget;
    position: Timecode;
    durationSeconds: number;
  }): Promise<{ clipId: string }> {
    return this.send<{ clipId: string }>("addText", params);
  }

  async applyEffect(params: {
    clipId: string;
    sequenceId: string;
    effect: EffectParams;
  }): Promise<void> {
    await this.send("applyEffect", params);
  }

  // -----------------------------------------------------------------------
  // Audio
  // -----------------------------------------------------------------------

  async setAudioLevel(params: {
    clipId: string;
    sequenceId: string;
    levelDb: number;
  }): Promise<void> {
    await this.send("setAudioLevel", params);
  }

  // -----------------------------------------------------------------------
  // Export
  // -----------------------------------------------------------------------

  async exportSequence(params: {
    sequenceId: string;
    outputPath: string;
    preset: ExportPreset;
  }): Promise<ExportResult> {
    return this.send<ExportResult>("exportSequence", params);
  }

  // -----------------------------------------------------------------------
  // Batch
  // -----------------------------------------------------------------------

  async executeEDL(params: {
    edl: EditDecisionList;
    autoImport: boolean;
    autoCreateSequence: boolean;
  }): Promise<EDLExecutionResult> {
    return this.send<EDLExecutionResult>("executeEDL", params);
  }

  // -----------------------------------------------------------------------
  // Generic Command
  // -----------------------------------------------------------------------

  async evalCommand(functionName: string, argsJson: string): Promise<EvalCommandResult> {
    try {
      const result = await this.send<unknown>("evalCommand", {
        function_name: functionName,
        args_json: argsJson,
      });
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
    try {
      return await this.send<PingResult>("ping", {});
    } catch {
      return {
        premiereRunning: false,
        premiereVersion: "unknown",
        projectOpen: false,
        bridgeMode: "cep",
      };
    }
  }

  // -----------------------------------------------------------------------
  // Private: WebSocket command transport
  // -----------------------------------------------------------------------

  /**
   * Send a command to the CEP panel and wait for its response.
   *
   * Each command gets a unique requestId. The promise resolves when the
   * CEP panel sends back a message with a matching requestId.
   */
  private send<T = unknown>(
    action: string,
    params: Record<string, unknown>,
  ): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(
          new CepConnectionError(
            `Cannot send "${action}": WebSocket is not open.`,
          ),
        );
        return;
      }

      const requestId = randomUUID();

      // Set up a timeout for this individual request.
      const timer = setTimeout(() => {
        this.pendingRequests.delete(requestId);
        reject(new CepTimeoutError(action, this.commandTimeoutMs));
      }, this.commandTimeoutMs);

      this.pendingRequests.set(requestId, {
        resolve: resolve as (value: unknown) => void,
        reject,
        timer,
      });

      const message: CepRequest = { action, params, requestId };

      this.log.debug(`Sending: ${action} (${requestId})`);
      this.ws.send(JSON.stringify(message), (err) => {
        if (err) {
          clearTimeout(timer);
          this.pendingRequests.delete(requestId);
          reject(
            new CepCommandError(action, `Failed to send: ${err.message}`),
          );
        }
      });
    });
  }

  /**
   * Handle an incoming WebSocket message from the CEP panel.
   */
  private handleMessage(data: WebSocket.Data): void {
    let response: CepResponse;
    try {
      const text = typeof data === "string" ? data : data.toString("utf-8");
      response = JSON.parse(text) as CepResponse;
    } catch {
      this.log.warn("Received non-JSON message from CEP panel; ignoring.");
      return;
    }

    const { requestId } = response;
    if (!requestId) {
      this.log.debug("Received message without requestId; ignoring.");
      return;
    }

    const pending = this.pendingRequests.get(requestId);
    if (!pending) {
      this.log.debug(`No pending request for id=${requestId}; ignoring.`);
      return;
    }

    clearTimeout(pending.timer);
    this.pendingRequests.delete(requestId);

    if (response.error) {
      pending.reject(
        new CepCommandError(requestId, response.error),
      );
    } else {
      pending.resolve(response.result);
    }
  }

  // -----------------------------------------------------------------------
  // Private: reconnection
  // -----------------------------------------------------------------------

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.log.error(
        `Max reconnection attempts (${this.maxReconnectAttempts}) reached. Giving up.`,
      );
      this.rejectAllPending("Max reconnection attempts exceeded");
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectIntervalMs * this.reconnectAttempts;
    this.log.info(
      `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`,
    );

    setTimeout(() => {
      if (this.intentionalClose) return;
      this.connect().catch((err: unknown) => {
        const detail = err instanceof Error ? err.message : String(err);
        this.log.warn(`Reconnection attempt failed: ${detail}`);
      });
    }, delay);
  }

  // -----------------------------------------------------------------------
  // Private: heartbeat
  // -----------------------------------------------------------------------

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.ping();
      }
    }, HEARTBEAT_INTERVAL_MS);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  // -----------------------------------------------------------------------
  // Private: cleanup
  // -----------------------------------------------------------------------

  /**
   * Reject every pending request with the given reason. Used on disconnect
   * and connection loss.
   */
  private rejectAllPending(reason: string): void {
    for (const [id, pending] of this.pendingRequests) {
      clearTimeout(pending.timer);
      pending.reject(new CepConnectionError(reason));
      this.pendingRequests.delete(id);
    }
  }
}
