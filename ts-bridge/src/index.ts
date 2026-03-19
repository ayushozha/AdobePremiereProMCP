/**
 * ts-bridge entry point.
 *
 * Starts the gRPC server that the Go orchestrator calls to drive
 * Adobe Premiere Pro operations through the TypeScript bridge layer.
 */

import { createLogger, format, transports } from "winston";
import { loadConfig } from "./config.js";
import { createBridge } from "./bridge/factory.js";
import { createGrpcServer } from "./grpc/server.js";

async function main(): Promise<void> {
  // ── Configuration ──────────────────────────────────────────────────────
  const config = loadConfig();

  // ── Logger ─────────────────────────────────────────────────────────────
  const logger = createLogger({
    level: config.logLevel,
    format: format.combine(
      format.timestamp(),
      format.errors({ stack: true }),
      format.printf(({ timestamp, level, message, ...meta }) => {
        const extra = Object.keys(meta).length
          ? ` ${JSON.stringify(meta)}`
          : "";
        return `${timestamp as string} [${level.toUpperCase()}] ${message as string}${extra}`;
      }),
    ),
    transports: [new transports.Console()],
  });

  logger.info("Premiere Pro TypeScript Bridge starting", {
    mode: config.bridgeMode,
    grpcPort: config.grpcPort,
    logLevel: config.logLevel,
  });

  // ── Bridge ─────────────────────────────────────────────────────────────
  const bridge = createBridge(config);
  await bridge.connect();
  logger.info(`Bridge connected in "${config.bridgeMode}" mode`);

  // ── gRPC server ────────────────────────────────────────────────────────
  const grpcServer = await createGrpcServer(config, bridge, logger);
  await grpcServer.start();

  // ── Graceful shutdown ──────────────────────────────────────────────────
  let shuttingDown = false;

  async function shutdown(signal: string): Promise<void> {
    if (shuttingDown) return;
    shuttingDown = true;
    logger.info(`Received ${signal}, shutting down...`);

    try {
      await grpcServer.stop();
      await bridge.disconnect();
      logger.info("Shutdown complete");
    } catch (err) {
      logger.error("Error during shutdown", { error: err });
      process.exitCode = 1;
    }
  }

  process.on("SIGINT", () => void shutdown("SIGINT"));
  process.on("SIGTERM", () => void shutdown("SIGTERM"));
}

main().catch((err: unknown) => {
  // eslint-disable-next-line no-console
  console.error("Fatal error starting ts-bridge:", err);
  process.exit(1);
});
