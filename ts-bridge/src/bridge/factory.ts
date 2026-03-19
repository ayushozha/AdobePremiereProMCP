/**
 * Bridge factory — returns the correct PremiereBridge implementation
 * based on configuration, with optional auto-detection.
 *
 * Detection order:
 *   1. If config.bridgeMode is set explicitly, use that mode.
 *   2. Otherwise, try to connect to the CEP panel WebSocket first.
 *   3. If CEP is unreachable, fall back to standalone (osascript).
 */

import { createLogger, format, transports, type Logger } from "winston";
import WebSocket from "ws";

import type { BridgeConfig } from "../config.js";
import type { PremiereBridge } from "./interface.js";
import { CepBridge } from "./cep-bridge.js";
import { StandaloneBridge } from "./standalone-bridge.js";

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const CEP_PROBE_TIMEOUT_MS = 3_000;

// ---------------------------------------------------------------------------
// Logger
// ---------------------------------------------------------------------------

function makeLogger(config: BridgeConfig): Logger {
  return createLogger({
    level: config.logLevel,
    format: format.combine(
      format.timestamp(),
      format.printf(({ timestamp, level, message }) =>
        `${timestamp as string} [factory] ${level}: ${message as string}`,
      ),
    ),
    transports: [new transports.Console()],
  });
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Create and return a {@link PremiereBridge} for the given configuration.
 *
 * - `"cep"` — communicates with Premiere Pro through a CEP panel WebSocket.
 * - `"standalone"` — drives Premiere Pro via the ExtendScript CLI toolkit.
 */
export function createBridge(config: BridgeConfig): PremiereBridge {
  switch (config.bridgeMode) {
    case "cep":
      return new CepBridge(config);
    case "standalone":
      return new StandaloneBridge(config);
    default: {
      const _exhaustive: never = config.bridgeMode;
      throw new Error(`Unknown bridge mode: ${String(_exhaustive)}`);
    }
  }
}

/**
 * Auto-detect the best available bridge mode and return a connected bridge.
 *
 * Steps:
 *   1. Probe the CEP panel's WebSocket port.
 *   2. If it responds, use the CEP bridge.
 *   3. Otherwise fall back to the standalone bridge.
 *
 * The returned bridge is already connected (its `connect()` has been called).
 */
export async function createBridgeAutoDetect(
  config: BridgeConfig,
): Promise<PremiereBridge> {
  const log = makeLogger(config);

  log.info("Auto-detecting bridge mode...");

  // --- Try CEP first ---
  const cepAvailable = await probeCepPanel(config, log);

  if (cepAvailable) {
    log.info("CEP panel WebSocket is reachable. Using CEP bridge.");
    const bridge = new CepBridge(config);
    await bridge.connect();
    return bridge;
  }

  // --- Fall back to standalone ---
  log.info("CEP panel not reachable. Falling back to standalone bridge.");
  const bridge = new StandaloneBridge(config);
  await bridge.connect();
  return bridge;
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

/**
 * Attempt a quick WebSocket connection to the CEP panel's port to see
 * whether it is alive.
 *
 * Returns `true` if the socket opens within {@link CEP_PROBE_TIMEOUT_MS},
 * `false` otherwise. The socket is closed immediately after probing.
 */
function probeCepPanel(config: BridgeConfig, log: Logger): Promise<boolean> {
  const port = config.cepWsPort || 9801;
  const url = `ws://localhost:${port}`;

  return new Promise<boolean>((resolve) => {
    log.debug(`Probing CEP panel at ${url}...`);

    const ws = new WebSocket(url);
    let settled = false;

    const timer = setTimeout(() => {
      if (!settled) {
        settled = true;
        ws.terminate();
        log.debug("CEP probe timed out.");
        resolve(false);
      }
    }, CEP_PROBE_TIMEOUT_MS);

    ws.on("open", () => {
      if (!settled) {
        settled = true;
        clearTimeout(timer);
        ws.close(1000);
        log.debug("CEP probe succeeded.");
        resolve(true);
      }
    });

    ws.on("error", (err: Error) => {
      if (!settled) {
        settled = true;
        clearTimeout(timer);
        log.debug(`CEP probe failed: ${err.message}`);
        resolve(false);
      }
    });

    ws.on("close", () => {
      if (!settled) {
        settled = true;
        clearTimeout(timer);
        resolve(false);
      }
    });
  });
}
