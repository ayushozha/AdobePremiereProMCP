/**
 * MCP stdio root: owns ts-bridge (child) and go-orchestrator server.exe (child).
 * stdin/stdout stay on server.exe via inherit → Cursor JSON-RPC unchanged.
 *
 * On Windows, Cursor often stops MCP by terminating the root process; children
 * can survive (orphans). We use taskkill /T on shutdown paths and spawnSync on
 * process 'exit' so the bridge is reliably torn down.
 */

"use strict";

const { spawn, spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const scriptsDir = __dirname;
const repoRoot = path.resolve(scriptsDir, "..");
const bridgeCwd = path.join(repoRoot, "ts-bridge");
const bridgeEntry = path.join(bridgeCwd, "dist", "index.js");
const serverExe = path.join(repoRoot, "go-orchestrator", "bin", "server.exe");
const isWin = process.platform === "win32";
const taskkillExe = isWin
  ? path.join(process.env.SystemRoot || "C:\\Windows", "System32", "taskkill.exe")
  : null;

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

/** Windows: kill pid and all descendants. No-op elsewhere. */
function killProcessTreeWin(pid, sync) {
  if (!isWin || !pid || !taskkillExe || !fs.existsSync(taskkillExe)) return;
  const args = ["/PID", String(pid), "/T", "/F"];
  try {
    if (sync) {
      spawnSync(taskkillExe, args, { stdio: "ignore", windowsHide: true });
    } else {
      const t = spawn(taskkillExe, args, {
        stdio: "ignore",
        windowsHide: true,
        detached: true,
      });
      t.unref();
    }
  } catch {
    /* ignore */
  }
}

function killBridge(bridge) {
  if (!bridge || !bridge.pid) return;
  if (isWin) {
    killProcessTreeWin(bridge.pid, false);
    return;
  }
  try {
    if (!bridge.killed) bridge.kill();
  } catch {
    /* ignore */
  }
}

function killServer(server) {
  if (!server || !server.pid) return;
  if (isWin) {
    killProcessTreeWin(server.pid, false);
    return;
  }
  try {
    if (!server.killed) server.kill();
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

  function cleanupAll() {
    killServer(server);
    killBridge(bridge);
  }

  function shutdownFromParentSignal() {
    killServer(server);
  }

  if (isWin) {
    process.on("exit", () => {
      killProcessTreeWin(bridge.pid, true);
      killProcessTreeWin(server.pid, true);
    });
  }

  if (!process.stdin.isTTY) {
    process.stdin.resume();
    process.stdin.on("end", () => {
      cleanupAll();
      setTimeout(() => process.exit(0), 0);
    });
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
      shutdownFromParentSignal();
    }
  });

  process.on("SIGINT", shutdownFromParentSignal);
  process.on("SIGTERM", shutdownFromParentSignal);
})().catch((e) => {
  err(String(e && e.stack ? e.stack : e));
  process.exit(1);
});
