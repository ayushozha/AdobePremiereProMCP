/**
 * Configuration for the Premiere Pro TypeScript bridge.
 *
 * All values are loaded from environment variables with sensible defaults.
 */

import type { ExportPreset } from "./bridge/interface.js";

/** Bridge communication mode: CEP panel inside Premiere or standalone Node process. */
export type BridgeMode = "cep" | "standalone";

/** Log verbosity level. */
export type LogLevel = "error" | "warn" | "info" | "debug";

/** User-configured Adobe Media Encoder .epr paths for typed preset names. */
export type ExportPresetPaths = Partial<Record<ExportPreset, string>>;

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

  /** Optional shared CEP authentication token; a per-user token file is used when omitted. */
  cepToken?: string;

  /** gRPC server host/bind address. */
  grpcHost: string;

  /** Absolute .epr paths used by the typed ExportSequence RPC. */
  exportPresetPaths?: ExportPresetPaths;
}

const EXPORT_PRESET_ENV_KEYS: Record<ExportPreset, string> = {
  h264_1080p: "PREMIERE_EXPORT_PRESET_H264_1080P",
  h264_4k: "PREMIERE_EXPORT_PRESET_H264_4K",
  prores_422: "PREMIERE_EXPORT_PRESET_PRORES_422",
  prores_4444: "PREMIERE_EXPORT_PRESET_PRORES_4444",
  dnx_hr: "PREMIERE_EXPORT_PRESET_DNX_HR",
  custom: "PREMIERE_EXPORT_PRESET_CUSTOM",
};

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
 * | BRIDGE_GRPC_HOST      | 127.0.0.1                                            |
 * | PREMIERE_PATH         | /Applications/Adobe Premiere Pro 2025/...             |
 * | BRIDGE_MODE           | cep                                                  |
 * | BRIDGE_LOG_LEVEL      | info                                                 |
 * | BRIDGE_CEP_WS_PORT    | 9801                                                 |
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
    // Premiere mutation RPCs are intentionally local-only by default. Remote
    // deployments must opt in and add transport authentication/TLS.
    grpcHost: process.env["BRIDGE_GRPC_HOST"] ?? "127.0.0.1",
    premierePath:
      process.env["PREMIERE_PATH"] ??
      "/Applications/Adobe Premiere Pro 2025/Adobe Premiere Pro 2025.app",
    bridgeMode: rawMode as BridgeMode,
    logLevel: rawLogLevel as LogLevel,
    cepWsPort: parsePort("BRIDGE_CEP_WS_PORT", 9801),
    cepToken: process.env["BRIDGE_CEP_TOKEN"] ?? process.env["MCP_CEP_TOKEN"],
    exportPresetPaths: loadExportPresetPaths(),
  };
}

function loadExportPresetPaths(): ExportPresetPaths {
  const paths: ExportPresetPaths = {};
  for (const [preset, envKey] of Object.entries(EXPORT_PRESET_ENV_KEYS) as Array<
    [ExportPreset, string]
  >) {
    const value = process.env[envKey]?.trim();
    if (value) paths[preset] = value;
  }
  return paths;
}

/** Resolve one typed preset or fail before sending a misleading export job. */
export function resolveExportPresetPath(
  paths: ExportPresetPaths | undefined,
  preset: ExportPreset,
): string {
  const value = paths?.[preset]?.trim();
  if (value) return value;
  throw new Error(
    `Export preset "${preset}" is not configured. Set ${EXPORT_PRESET_ENV_KEYS[preset]} to an absolute Adobe Media Encoder .epr path, or use an explicit preset_path export tool.`,
  );
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
