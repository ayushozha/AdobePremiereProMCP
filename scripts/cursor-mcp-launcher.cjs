/**
 * MCP stdio root: owns ts-bridge (child) and go-orchestrator server.exe (child).
 * stdin/stdout stay on server.exe via inherit → Cursor JSON-RPC unchanged.
 * On server exit OR launcher signal, bridge is terminated so Premiere/CEP sees disconnect.
 */

"use strict";

const { spawn } = require("child_process");
const fs = require("fs");
const path = require("path");

const scriptsDir = __dirname;
const repoRoot = path.resolve(scriptsDir, "..");
const bridgeCwd = path.join(repoRoot, "ts-bridge");
const bridgeEntry = path.join(bridgeCwd, "dist", "index.js");
const serverExe = path.join(repoRoot, "go-orchestrator", "bin", "server.exe");

function err(msg) {
  console.error("[cursor-mcp-launcher]", msg);
}

if (!fs.existsSync(bridgeEntry)) {
  err(`Missing ts-bridge build: ${bridgeEntry}`);
  process.exit(1);
}
if (!fs.existsSync(serverExe)) {
  err(`Missing go orchestrator binary: ${serverExe}`);
  process.exit(1);
}

const launcherArgs = process.argv.slice(2);

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function killBridge(bridge) {
  try {
    if (bridge && bridge.pid && !bridge.killed) {
      bridge.kill();
    }
  } catch {
    /* ignore */
  }
}

(async function main() {
  const bridge = spawn(process.execPath, [bridgeEntry], {
    cwd: bridgeCwd,
    stdio: "ignore",
    windowsHide: true,
    detached: false,
  });

  bridge.on("error", (e) => {
    err(`Failed to start ts-bridge: ${e.message}`);
    process.exit(1);
  });

  await sleep(2000);

  const server = spawn(serverExe, launcherArgs, {
    cwd: path.dirname(serverExe),
    stdio: "inherit",
    windowsHide: false,
    detached: false,
    env: process.env,
  });

  function shutdownFromParent() {
    try {
      if (server.pid && !server.killed) {
        server.kill();
      }
    } catch {
      /* ignore */
    }
  }

  server.on("error", (e) => {
    err(`Failed to start server.exe: ${e.message}`);
    killBridge(bridge);
    process.exit(1);
  });

  server.on("exit", (code) => {
    killBridge(bridge);
    process.exit(code === null ? 1 : code);
  });

  bridge.on("exit", (code, signal) => {
    if (signal || (code !== 0 && code !== null)) {
      err(`ts-bridge exited unexpectedly (code=${code}, signal=${signal})`);
      shutdownFromParent();
    }
  });

  process.on("SIGINT", shutdownFromParent);
  process.on("SIGTERM", shutdownFromParent);
  process.on("exit", () => killBridge(bridge));
})().catch((e) => {
  err(String(e && e.stack ? e.stack : e));
  process.exit(1);
});
