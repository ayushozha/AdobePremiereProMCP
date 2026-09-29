import assert from "node:assert/strict";
import childProcess from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import test from "node:test";
import { StandaloneBridge } from "./standalone-bridge.js";

function bridge(): StandaloneBridge {
  return new StandaloneBridge({
    grpcPort: 50054,
    grpcHost: "127.0.0.1",
    premierePath: "/Applications/Adobe Premiere Pro 2026.app",
    bridgeMode: "standalone",
    logLevel: "error",
    cepWsPort: 9801,
  });
}

test("standalone health never sends AppleScript to a stopped application", async (t) => {
  const commands: string[] = [];
  t.mock.method(childProcess, "execFileSync", (command: string) => {
    commands.push(command);
    throw new Error("No matching process");
  });
  syncBuiltinESMExports();
  t.after(() => { t.mock.restoreAll(); syncBuiltinESMExports(); });
  const result = await bridge().ping();
  assert.equal(result.premiereRunning, false);
  assert.deepEqual(commands, ["pgrep"]);
});

test("standalone health queries the host when its exact application process exists", async (t) => {
  const commands: {command: string; args: string[]}[] = [];
  t.mock.method(childProcess, "execFileSync", (command: string, args: string[]) => {
    commands.push({ command, args });
    if (command === "pgrep") return "12345\n";
    return JSON.stringify({ premiereRunning: true, premiereVersion: "26.0", projectOpen: false, bridgeMode: "standalone" });
  });
  syncBuiltinESMExports();
  t.after(() => { t.mock.restoreAll(); syncBuiltinESMExports(); });
  assert.equal((await bridge().ping()).premiereRunning, true);
  assert.equal(commands.length, 2);
  assert.deepEqual(commands[0], { command: "pgrep", args: ["-f", "/Adobe Premiere Pro 2026\\.app/Contents/MacOS/"] });
  assert.equal(commands[1]?.command, "osascript");
});
