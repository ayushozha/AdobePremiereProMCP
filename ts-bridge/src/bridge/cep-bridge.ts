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

import { resolveSharedToken } from "../auth/token.js";
import {
  resolveExportPresetPath,
  type BridgeConfig,
  type ExportPresetPaths,
} from "../config.js";
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
const DEFAULT_MAX_RECONNECT_ATTEMPTS = 100;
const DEFAULT_RECONNECT_BASE_MS = 5_000;
const DEFAULT_RECONNECT_MAX_MS = 30_000;
const HEARTBEAT_INTERVAL_MS = 15_000;
const PONG_TIMEOUT_MS = 5_000;

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

export class CepBridge implements PremiereBridge {
  private readonly log: Logger;
  private readonly wsEndpoint: string;
  private readonly wsHeaders: Readonly<Record<string, string>>;
  private readonly commandTimeoutMs: number;
  private readonly maxReconnectAttempts: number;
  private readonly reconnectBaseMs: number;
  private readonly reconnectMaxMs: number;
  private readonly exportPresetPaths: ExportPresetPaths;

  private ws: WebSocket | null = null;
  private pendingRequests = new Map<string, PendingRequest>();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private pongTimer: ReturnType<typeof setTimeout> | null = null;
  private awaitingPong = false;
  private intentionalClose = false;
  private _isConnected = false;

  constructor(config: BridgeConfig) {
    const port = config.cepWsPort || DEFAULT_WS_PORT;
    const token = resolveSharedToken(config.cepToken);
    this.wsEndpoint = `ws://127.0.0.1:${port}`;
    this.wsHeaders = { Authorization: `Bearer ${token}` };
    this.commandTimeoutMs = DEFAULT_COMMAND_TIMEOUT_MS;
    this.maxReconnectAttempts = DEFAULT_MAX_RECONNECT_ATTEMPTS;
    this.reconnectBaseMs = DEFAULT_RECONNECT_BASE_MS;
    this.reconnectMaxMs = DEFAULT_RECONNECT_MAX_MS;
    this.exportPresetPaths = config.exportPresetPaths ?? {};

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

    return new Promise<void>((resolve) => {
      this.log.info(`Connecting to CEP panel at ${this.wsEndpoint}...`);

      const ws = new WebSocket(this.wsEndpoint, { headers: this.wsHeaders });
      let settled = false;
      // Track whether the socket ever successfully opened so we only
      // auto-reconnect on connections that were previously established.
      let wasOpen = false;

      const connectionTimeout = setTimeout(() => {
        if (!settled) {
          settled = true;
          ws.terminate();
          this.log.warn(
            `Connection to CEP panel at ${this.wsEndpoint} timed out. ` +
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

        this.log.info(`Connected to CEP panel at ${this.wsEndpoint}`);
        resolve();
      });

      ws.on("pong", () => {
        this.awaitingPong = false;
        if (this.pongTimer) {
          clearTimeout(this.pongTimer);
          this.pongTimer = null;
        }
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
            this.scheduleReconnect();
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
    this.cancelReconnect();
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
    const raw = await this.invokeHost<Record<string, unknown>>(
      "getProjectState",
    );
    const rawSequences = Array.isArray(raw["sequences"])
      ? raw["sequences"] as Array<Record<string, unknown>>
      : [];

    return {
      projectName: String(raw["projectName"] ?? raw["name"] ?? ""),
      projectPath: String(raw["projectPath"] ?? raw["path"] ?? ""),
      sequences: rawSequences.map((sequence, index) => ({
        id: String(
          sequence["id"] ?? sequence["sequenceId"] ??
          sequence["sequenceID"] ?? index,
        ),
        name: String(sequence["name"] ?? `Sequence ${index + 1}`),
        resolution: {
          width: Number(
            (sequence["resolution"] as Record<string, unknown> | undefined)?.["width"] ??
            sequence["frameSizeHorizontal"] ?? sequence["width"] ?? 0,
          ),
          height: Number(
            (sequence["resolution"] as Record<string, unknown> | undefined)?.["height"] ??
            sequence["frameSizeVertical"] ?? sequence["height"] ?? 0,
          ),
        },
        frameRate: Number(sequence["frameRate"] ?? sequence["fps"] ?? 0),
        durationSeconds: Number(
          sequence["durationSeconds"] ?? sequence["outPoint"] ?? 0,
        ),
        videoTrackCount: Number(sequence["videoTrackCount"] ?? 0),
        audioTrackCount: Number(sequence["audioTrackCount"] ?? 0),
      })),
      binCount: Number(raw["binCount"] ?? 0),
      // The CEP host cannot reliably expose Premiere's dirty-document state.
      // Treat an absent value as unknown/unsaved instead of reporting a false
      // guarantee that the project is safely on disk.
      isSaved: Boolean(raw["isSaved"] ?? false),
    };
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
    const raw = await this.invokeHost<Record<string, unknown>>(
      "createSequence",
      {
        name: params.name,
        width: params.resolution.width,
        height: params.resolution.height,
        fps: params.frameRate,
        videoTracks: params.videoTracks,
        audioTracks: params.audioTracks,
      },
    );
    return {
      sequenceId: String(
        raw["sequenceId"] ?? raw["sequenceID"] ?? raw["id"] ?? "",
      ),
      name: String(raw["name"] ?? params.name),
    };
  }

  async getTimelineState(sequenceId: string): Promise<TimelineState> {
    return this.invokeHost<TimelineState>("mcpGetTimelineState", {
      sequenceId,
    });
  }

  // -----------------------------------------------------------------------
  // Clip operations
  // -----------------------------------------------------------------------

  async importMedia(params: {
    filePath: string;
    targetBin: string;
  }): Promise<{ projectItemId: string; name: string }> {
    const raw = await this.invokeHost<Record<string, unknown>>(
      "importMedia",
      { filePath: params.filePath, binPath: params.targetBin },
    );
    return {
      projectItemId: String(
        raw["projectItemId"] ?? raw["nodeId"] ?? raw["mediaPath"] ??
        params.filePath,
      ),
      name: String(raw["name"] ?? params.filePath.split(/[\\/]/).pop() ?? ""),
    };
  }

  async placeClip(params: {
    sourcePath: string;
    track: TrackTarget;
    position: Timecode;
    sourceRange?: TimeRange;
    speed: number;
  }): Promise<{ clipId: string }> {
    return this.invokeHost<{ clipId: string }>("mcpPlaceClip", params);
  }

  async removeClip(params: {
    clipId: string;
    sequenceId: string;
  }): Promise<void> {
    await this.invokeHost("mcpRemoveClip", params);
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
    return this.invokeHost<{ transitionId: string }>(
      "mcpAddTransition",
      params,
    );
  }

  async addText(params: {
    sequenceId: string;
    text: string;
    style: TextStyle;
    track: TrackTarget;
    position: Timecode;
    durationSeconds: number;
  }): Promise<{ clipId: string }> {
    return this.invokeHost<{ clipId: string }>("mcpAddText", params);
  }

  async applyEffect(params: {
    clipId: string;
    sequenceId: string;
    effect: EffectParams;
  }): Promise<void> {
    await this.invokeHost("mcpApplyEffect", params);
  }

  // -----------------------------------------------------------------------
  // Audio
  // -----------------------------------------------------------------------

  async setAudioLevel(params: {
    clipId: string;
    sequenceId: string;
    levelDb: number;
  }): Promise<void> {
    await this.invokeHost("mcpSetAudioLevel", params);
  }

  // -----------------------------------------------------------------------
  // Export
  // -----------------------------------------------------------------------

  async exportSequence(params: {
    sequenceId: string;
    outputPath: string;
    preset: ExportPreset;
  }): Promise<ExportResult> {
    const presetPath = resolveExportPresetPath(
      this.exportPresetPaths,
      params.preset,
    );
    const raw = await this.invokeHost<Record<string, unknown>>(
      "exportSequence",
      {
        sequenceId: params.sequenceId,
        outputPath: params.outputPath,
        presetPath,
      },
    );
    const rawStatus = String(raw["status"] ?? "pending");
    const status: ExportResult["status"] = rawStatus === "completed" || rawStatus === "export_complete"
      ? "completed"
      : rawStatus === "failed"
        ? "failed"
        : rawStatus === "running"
          ? "running"
          : "pending";
    return {
      jobId: String(raw["jobId"] ?? raw["jobID"] ?? "queued"),
      status,
      outputPath: String(raw["outputPath"] ?? params.outputPath),
    };
  }

  // -----------------------------------------------------------------------
  // Batch
  // -----------------------------------------------------------------------

  async executeEDL(params: {
    edl: EditDecisionList;
    autoImport: boolean;
    autoCreateSequence: boolean;
  }): Promise<EDLExecutionResult> {
    return this.invokeHost<EDLExecutionResult>("mcpExecuteEDL", params);
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
      const raw = await this.invokeHost<Record<string, unknown>>("ping");
      return {
        premiereRunning: Boolean(
          raw["premiereRunning"] ?? raw["premiere_running"] ??
          raw["status"] === "ok",
        ),
        premiereVersion: String(
          raw["premiereVersion"] ?? raw["premiere_version"] ??
          raw["version"] ?? "unknown",
        ),
        projectOpen: Boolean(raw["projectOpen"] ?? raw["project_open"] ?? false),
        bridgeMode: "cep",
      };
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

  /** Invoke a named premiere.jsx function through the generic dispatcher. */
  private async invokeHost<T = unknown>(
    functionName: string,
    args: object = {},
  ): Promise<T> {
    const response = await this.evalCommand(functionName, JSON.stringify(args));
    if (response.isError) {
      throw new CepCommandError(functionName, response.errorMessage);
    }
    if (response.resultJson === "") return undefined as T;
    try {
      return JSON.parse(response.resultJson) as T;
    } catch {
      return response.resultJson as T;
    }
  }

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
  // Private: reconnection with exponential backoff
  // -----------------------------------------------------------------------

  /**
   * Compute the next reconnection delay using exponential backoff.
   * Base delay doubles each attempt: 5s, 10s, 20s, capped at 30s.
   */
  private getReconnectDelay(): number {
    const exponential = this.reconnectBaseMs * Math.pow(2, this.reconnectAttempts - 1);
    return Math.min(exponential, this.reconnectMaxMs);
  }

  /**
   * Schedule a reconnection attempt. Called when the WebSocket closes
   * unexpectedly (not a clean/intentional shutdown).
   */
  private scheduleReconnect(): void {
    if (this.intentionalClose) return;

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.log.error(
        `Max reconnection attempts (${this.maxReconnectAttempts}) reached. ` +
          "Giving up. Restart the bridge manually or call connect() again.",
      );
      this.rejectAllPending("Max reconnection attempts exceeded");
      return;
    }

    this.reconnectAttempts++;
    const delay = this.getReconnectDelay();
    this.log.info(
      `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`,
    );

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.intentionalClose) return;

      this.log.info(
        `Attempting reconnection (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`,
      );

      this.connect()
        .then(() => {
          if (this._isConnected) {
            this.log.info(
              `Reconnected to CEP panel at ${this.wsEndpoint} after ${this.reconnectAttempts} attempt(s).`,
            );
            // Reset attempt counter on successful connection -- already
            // done inside connect()'s "open" handler.
          } else {
            // connect() resolved but we are not connected (timeout/error path).
            // Schedule another attempt.
            this.scheduleReconnect();
          }
        })
        .catch((err: unknown) => {
          const detail = err instanceof Error ? err.message : String(err);
          this.log.warn(`Reconnection attempt failed: ${detail}`);
          this.scheduleReconnect();
        });
    }, delay);
  }

  /**
   * Cancel any pending reconnection timer.
   */
  private cancelReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  // -----------------------------------------------------------------------
  // Private: heartbeat with dead-connection detection
  // -----------------------------------------------------------------------

  /**
   * Start a periodic ping. Every HEARTBEAT_INTERVAL_MS we send a WebSocket
   * ping frame. If no pong is received within PONG_TIMEOUT_MS the connection
   * is considered dead and we trigger a reconnect.
   */
  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.awaitingPong = false;

    this.heartbeatTimer = setInterval(() => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;

      if (this.awaitingPong) {
        // Previous ping never got a pong -- connection is dead.
        this.log.warn("Pong timeout: connection appears dead. Triggering reconnect.");
        this.awaitingPong = false;
        if (this.pongTimer) {
          clearTimeout(this.pongTimer);
          this.pongTimer = null;
        }
        // Force-close the socket so the "close" handler fires and reconnect kicks in.
        this.ws.terminate();
        return;
      }

      this.awaitingPong = true;
      this.ws.ping();

      // Set a deadline for the pong response.
      this.pongTimer = setTimeout(() => {
        if (this.awaitingPong && this.ws) {
          this.log.warn(
            `No pong received within ${PONG_TIMEOUT_MS}ms. Connection is dead.`,
          );
          this.awaitingPong = false;
          this.pongTimer = null;
          this.ws.terminate();
        }
      }, PONG_TIMEOUT_MS);
    }, HEARTBEAT_INTERVAL_MS);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
    if (this.pongTimer) {
      clearTimeout(this.pongTimer);
      this.pongTimer = null;
    }
    this.awaitingPong = false;
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
