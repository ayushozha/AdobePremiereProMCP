#!/usr/bin/env node

"use strict";

const path = require("node:path");
const os = require("node:os");
const readline = require("node:readline");
const { spawn } = require("node:child_process");

const root = path.resolve(__dirname, "..");
const binary = process.env.MCP_SERVER_BIN || path.join(root, "go-orchestrator", "bin", "premierpro-mcp");
const child = spawn(binary, ["--transport=stdio", "--log-level=error"], {
  cwd: root,
  env: {
    ...process.env,
    MCP_TOOL_PROFILE: process.env.MCP_TOOL_PROFILE || "standard",
    BRIDGE_CEP_TOKEN: process.env.BRIDGE_CEP_TOKEN || "mcp-smoke-token-000000000000000000000000000000000000000000000000",
  },
  stdio: ["pipe", "pipe", "pipe"],
});

let nextID = 1;
let stderr = "";
const pending = new Map();

child.stderr.setEncoding("utf8");
child.stderr.on("data", (chunk) => { stderr += chunk; });

function send(payload) {
  child.stdin.write(`${JSON.stringify(payload)}\n`);
}

function request(method, params = {}) {
  const id = nextID++;
  send({ jsonrpc: "2.0", id, method, params });
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      pending.delete(id);
      reject(new Error(`timed out waiting for ${method}`));
    }, 10000);
    pending.set(id, { method, resolve, reject, timer });
  });
}

readline.createInterface({ input: child.stdout }).on("line", (line) => {
  if (!line.trim()) return;
  let message;
  try { message = JSON.parse(line); }
  catch (error) {
    for (const item of pending.values()) item.reject(new Error(`invalid MCP JSON: ${line}`));
    pending.clear();
    return;
  }
  if (message.id === undefined || !pending.has(message.id)) return;
  const item = pending.get(message.id);
  pending.delete(message.id);
  clearTimeout(item.timer);
  if (message.error) item.reject(new Error(`${item.method}: ${JSON.stringify(message.error)}`));
  else item.resolve(message.result);
});

child.on("error", (error) => {
  for (const item of pending.values()) item.reject(error);
  pending.clear();
});

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

async function listAllTools() {
  const tools = [];
  const seen = new Set();
  let cursor;
  do {
    const result = await request("tools/list", cursor ? { cursor } : {});
    tools.push(...result.tools);
    cursor = result.nextCursor || "";
    assert(!cursor || !seen.has(cursor), `repeated tools/list cursor: ${cursor}`);
    if (cursor) seen.add(cursor);
  } while (cursor);
  return tools;
}

async function main() {
  await request("initialize", {
    protocolVersion: "2024-11-05",
    capabilities: {},
    clientInfo: { name: "premierpro-mcp-smoke", version: "1.0.0" },
  });
  send({ jsonrpc: "2.0", method: "notifications/initialized", params: {} });

  const tools = await listAllTools();
  const toolNames = new Set(tools.map((tool) => tool.name));
  const resources = await request("resources/list");
  const prompts = await request("prompts/list");
  // Every profile includes core, and set_active_sequence has a required
  // argument. This keeps the validation probe valid for specialized profiles.
  const invalidCall = await request("tools/call", { name: "premiere_set_active_sequence", arguments: {} });
  const profile = process.env.MCP_TOOL_PROFILE || "standard";
  const verifyBackends = process.env.MCP_SMOKE_BACKENDS === "1";

  if (profile === "standard") {
    assert(tools.length >= 50 && tools.length <= 128, `standard profile has unexpected size: ${tools.length}`);
    assert(toolNames.has("premiere_add_subtitles_from_srt"), "caption import tool missing");
    assert(toolNames.has("premiere_apply_video_effect"), "verified effect tool missing");
    assert(toolNames.has("premiere_add_video_transition"), "verified transition tool missing");
    assert(!toolNames.has("premiere_execute_system_command"), "unsafe system-command tool leaked into standard profile");
    assert(!toolNames.has("premiere_split_long_captions"), "unverified caption splitter leaked into standard profile");
    assert(prompts.prompts.length === 4, `standard prompt count = ${prompts.prompts.length}, want 4`);
  } else if (profile === "all,unsafe") {
    assert(tools.length === 1064, `full profile tool count = ${tools.length}, want 1064`);
    assert(toolNames.has("premiere_execute_system_command"), "explicit unsafe profile omitted system-command tool");
    assert(prompts.prompts.length === 5, `full prompt count = ${prompts.prompts.length}, want 5`);
  }
  assert(resources.resources.length === 5, `resource count = ${resources.resources.length}, want 5`);
  assert(invalidCall.isError === true, "missing required tool arguments were accepted");

  if (verifyBackends) {
    const scan = await request("tools/call", {
      name: "premiere_scan_assets",
      arguments: {
        directory: os.tmpdir(),
        recursive: false,
        extensions: [".premiere-mcp-smoke-no-match"],
      },
    });
    assert(scan.isError !== true, `Rust media-engine smoke call failed: ${JSON.stringify(scan.content)}`);

    const parsed = await request("tools/call", {
      name: "premiere_parse_script",
      arguments: {
        text: "INTRO\nWelcome to the PremierPro MCP backend smoke test.",
        format: "youtube",
      },
    });
    assert(parsed.isError !== true, `Python intelligence smoke call failed: ${JSON.stringify(parsed.content)}`);

    const bridge = await request("tools/call", {
      name: "premiere_is_running",
      arguments: {},
    });
    assert(bridge.isError !== true, `TypeScript bridge smoke call failed: ${JSON.stringify(bridge.content)}`);
  }

  console.log(JSON.stringify({
    profile,
    tools: tools.length,
    resources: resources.resources.length,
    prompts: prompts.prompts.length,
    paginationVerified: true,
    requiredArgumentsVerified: true,
    backendServicesVerified: verifyBackends,
  }));
}

main()
  .then(() => { child.stdin.end(); })
  .catch((error) => {
    child.kill("SIGTERM");
    console.error(error.message);
    if (stderr.trim()) console.error(stderr.trim());
    process.exitCode = 1;
  });
