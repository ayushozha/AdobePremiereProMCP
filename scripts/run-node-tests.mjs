#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { readdirSync } from "node:fs";
import { resolve } from "node:path";

const roots = process.argv.slice(2);
if (roots.length === 0) {
  throw new Error("usage: run-node-tests.mjs <compiled-test-directory> [...]");
}

function collectTests(directory) {
  const tests = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = resolve(directory, entry.name);
    if (entry.isDirectory()) {
      tests.push(...collectTests(path));
    } else if (entry.isFile() && /\.test\.(?:c|m)?js$/.test(entry.name)) {
      tests.push(path);
    }
  }
  return tests;
}

const tests = roots.flatMap((root) => collectTests(resolve(root))).sort();
if (tests.length === 0) {
  throw new Error(`no compiled tests found below: ${roots.join(", ")}`);
}

const result = spawnSync(process.execPath, ["--test", ...tests], {
  stdio: "inherit",
});
if (result.error) {
  throw result.error;
}
process.exit(result.status ?? 1);
