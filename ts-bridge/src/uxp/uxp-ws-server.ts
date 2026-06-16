/**
 * WebSocket server for the PremierPro MCP UXP panel.
 *
 * The UXP panel connects as a client. Transcript operations (Text panel data)
 * are routed here when available. ExtendScript caption-track fallback covers
 * Premiere 24.x when UXP is unavailable.
 */

import { randomUUID } from "node:crypto";
import WebSocket, { WebSocketServer } from "ws";
import type { Logger } from "winston";

import type { BridgeConfig } from "../config.js";

export interface UxpCommandResult {
  success: boolean;
  result?: unknown;
  error?: string;
}

interface PendingRequest {
  resolve: (value: UxpCommandResult) => void;
  reject: (reason: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

export class UxpWsServer {
  private wss: WebSocketServer | null = null;
  private client: WebSocket | null = null;
  private readonly pending = new Map<string, PendingRequest>();
  private readonly commandTimeoutMs: number;

  constructor(
    private readonly config: BridgeConfig,
    private readonly logger: Logger,
  ) {
    this.commandTimeoutMs = config.uxpCommandTimeoutMs;
  }

  get isConnected(): boolean {
    return this.client !== null && this.client.readyState === WebSocket.OPEN;
  }

  async start(): Promise<void> {
    if (this.wss) return;

    const port = this.config.uxpWsPort;
    this.wss = new WebSocketServer({ host: "127.0.0.1", port });

    await new Promise<void>((resolve, reject) => {
      this.wss!.once("listening", () => resolve());
      this.wss!.once("error", (err) => reject(err));
    });

    this.wss.on("connection", (socket, req) => {
      const remote = req.socket.remoteAddress ?? "unknown";
      this.logger.info("UXP panel connected", { remote });

      if (this.client && this.client.readyState === WebSocket.OPEN) {
        this.logger.warn("Replacing existing UXP panel connection");
        try {
          this.client.close(1000, "replaced");
        } catch {
          /* ignore */
        }
      }

      this.client = socket;

      socket.on("message", (data) => {
        this.handleMessage(data.toString());
      });

      socket.on("close", () => {
        if (this.client === socket) {
          this.client = null;
          this.logger.warn("UXP panel disconnected");
        }
      });

      socket.on("error", (err) => {
        this.logger.error("UXP panel socket error", { error: err.message });
      });
    });

    this.logger.info("UXP WebSocket server listening", { port });
  }

  async stop(): Promise<void> {
    for (const [, pending] of this.pending) {
      clearTimeout(pending.timer);
      pending.reject(new Error("UXP WebSocket server shutting down"));
    }
    this.pending.clear();

    if (this.client) {
      try {
        this.client.close(1000);
      } catch {
        /* ignore */
      }
      this.client = null;
    }

    if (this.wss) {
      await new Promise<void>((resolve) => {
        this.wss!.close(() => resolve());
      });
      this.wss = null;
    }
  }

  async sendCommand(
    action: string,
    params: Record<string, unknown> = {},
  ): Promise<UxpCommandResult> {
    if (!this.isConnected || !this.client) {
      return {
        success: false,
        error:
          "UXP bridge panel is not connected. Open Window > UXP Plugins > PremierPro MCP UXP Bridge in Premiere Pro (requires Premiere 25.0+).",
      };
    }

    const requestId = randomUUID();
    const payload = JSON.stringify({ requestId, action, params });

    return new Promise<UxpCommandResult>((resolve, reject) => {
      const timer = setTimeout(() => {
        this.pending.delete(requestId);
        reject(
          new Error(
            `UXP command "${action}" timed out after ${this.commandTimeoutMs}ms`,
          ),
        );
      }, this.commandTimeoutMs);

      this.pending.set(requestId, { resolve, reject, timer });

      try {
        this.client!.send(payload);
      } catch (err) {
        clearTimeout(timer);
        this.pending.delete(requestId);
        reject(err instanceof Error ? err : new Error(String(err)));
      }
    });
  }

  private handleMessage(raw: string): void {
    let message: Record<string, unknown>;
    try {
      message = JSON.parse(raw) as Record<string, unknown>;
    } catch {
      this.logger.warn("UXP panel sent non-JSON message");
      return;
    }

    if (message["type"] === "hello") {
      this.logger.info("UXP panel hello", {
        role: message["role"],
        version: message["version"],
      });
      return;
    }

    const requestId = message["requestId"];
    if (typeof requestId !== "string") return;

    const pending = this.pending.get(requestId);
    if (!pending) return;

    clearTimeout(pending.timer);
    this.pending.delete(requestId);

    pending.resolve({
      success: message["success"] === true,
      result: message["result"],
      error:
        typeof message["error"] === "string" ? message["error"] : undefined,
    });
  }
}

let singleton: UxpWsServer | null = null;

export function setUxpWsServer(server: UxpWsServer): void {
  singleton = server;
}

export function getUxpWsServer(): UxpWsServer | null {
  return singleton;
}
