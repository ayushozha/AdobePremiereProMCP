#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const { constants } = require('node:os');
const { ensureBinary } = require('../lib/install.cjs');

async function main() {
  const binary = await ensureBinary();
  // No shell: preserve every argument literally. Inherited stdio keeps stdout
  // exclusively available for MCP JSON-RPC messages, with no buffering layer.
  const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit', shell: false });
  const signals = process.platform === 'win32' ? ['SIGINT', 'SIGTERM'] : ['SIGINT', 'SIGTERM', 'SIGHUP'];
  const handlers = new Map(signals.map((signal) => [signal, () => child.kill(signal)]));
  for (const [signal, handler] of handlers) process.on(signal, handler);
  function cleanup() {
    for (const [signal, handler] of handlers) process.removeListener(signal, handler);
  }
  child.once('error', (error) => {
    cleanup();
    process.stderr.write(`premierpro-mcp: cannot start the native server: ${error.message}\n`);
    process.exitCode = 1;
  });
  child.once('exit', (code, signal) => {
    cleanup();
    if (signal && process.platform !== 'win32') process.kill(process.pid, signal);
    else process.exitCode = code ?? (128 + (constants.signals[signal] || 1));
  });
}

main().catch((error) => {
  process.stderr.write(`premierpro-mcp: ${error.message}\n`);
  process.exitCode = 1;
});
