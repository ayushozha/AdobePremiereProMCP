/**
 * Configuration for the Premiere Pro TypeScript bridge.
 *
 * All values are loaded from environment variables with sensible defaults.
 */

/** Bridge communication mode: CEP panel inside Premiere or standalone Node process. */
export type BridgeMode = "cep" | "standalone";

/** Log verbosity level. */
export type LogLevel = "error" | "warn" | "info" | "debug";

export interface BridgeConfig {
  /** Port the gRPC server listens on. */
  grpcPort: number;

  /** Absolute path to the Adobe Premiere Pro executable. */
  premierePath: string;

  /** How this bridge communicates with Premiere Pro. */
  bridgeMode: BridgeMode;

  /** Winston log level. */
  logLevel: LogLevel;

  /** WebSocket port for CEP panel communication (only used in cep mode). */
  cepWsPort: number;

  /** gRPC server host/bind address. */
  grpcHost: string;
}

const VALID_BRIDGE_MODES: ReadonlySet<string> = new Set<BridgeMode>([
  "cep",
  "standalone",
]);

const VALID_LOG_LEVELS: ReadonlySet<string> = new Set<LogLevel>([
  "error",
  "warn",
  "info",
  "debug",
]);

/**
 * Load configuration from environment variables.
 *
 * | Variable              | Default                                              |
 * |-----------------------|------------------------------------------------------|
 * | BRIDGE_GRPC_PORT      | 50054                                                |
 * | BRIDGE_GRPC_HOST      | 0.0.0.0                                              |
 * | PREMIERE_PATH         | /Applications/Adobe Premiere Pro 2025/...             |
 * | BRIDGE_MODE           | cep                                                  |
 * | BRIDGE_LOG_LEVEL      | info                                                 |
 * | BRIDGE_CEP_WS_PORT    | 8089                                                 |
 */
export function loadConfig(): BridgeConfig {
  const rawMode = process.env["BRIDGE_MODE"] ?? "cep";
  if (!VALID_BRIDGE_MODES.has(rawMode)) {
    throw new Error(
      `Invalid BRIDGE_MODE "${rawMode}". Must be one of: ${[...VALID_BRIDGE_MODES].join(", ")}`,
    );
  }

  const rawLogLevel = process.env["BRIDGE_LOG_LEVEL"] ?? "info";
  if (!VALID_LOG_LEVELS.has(rawLogLevel)) {
    throw new Error(
      `Invalid BRIDGE_LOG_LEVEL "${rawLogLevel}". Must be one of: ${[...VALID_LOG_LEVELS].join(", ")}`,
    );
  }

  return {
    grpcPort: parsePort("BRIDGE_GRPC_PORT", 50054),
    grpcHost: process.env["BRIDGE_GRPC_HOST"] ?? "0.0.0.0",
    premierePath:
      process.env["PREMIERE_PATH"] ??
      "/Applications/Adobe Premiere Pro 2025/Adobe Premiere Pro 2025.app",
    bridgeMode: rawMode as BridgeMode,
    logLevel: rawLogLevel as LogLevel,
    cepWsPort: parsePort("BRIDGE_CEP_WS_PORT", 9801),
  };
}

function parsePort(envKey: string, fallback: number): number {
  const raw = process.env[envKey];
  if (raw === undefined) return fallback;

  const parsed = Number.parseInt(raw, 10);
  if (Number.isNaN(parsed) || parsed < 1 || parsed > 65535) {
    throw new Error(
      `Invalid ${envKey} "${raw}". Must be a port number between 1 and 65535.`,
    );
  }
  return parsed;
}
