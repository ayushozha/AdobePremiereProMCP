import assert from "node:assert/strict";
import { chmodSync, mkdtempSync, mkdirSync, rmSync, statSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { resolveSharedToken } from "./token.js";

test("does not chmod an existing custom token directory", () => {
  const root = mkdtempSync(join(tmpdir(), "premierpro-token-"));
  const customParent = join(root, "shared");
  mkdirSync(customParent, { mode: 0o755 });
  if (process.platform !== "win32") {
    chmodSync(customParent, 0o755);
  }
  const modeBefore = statSync(customParent).mode & 0o777;
  const previousPath = process.env["PREMIERE_MCP_TOKEN_FILE"];
  process.env["PREMIERE_MCP_TOKEN_FILE"] = join(customParent, "token");
  try {
    assert.equal(resolveSharedToken().length, 64);
    assert.equal(statSync(customParent).mode & 0o777, modeBefore);
  } finally {
    if (previousPath === undefined) delete process.env["PREMIERE_MCP_TOKEN_FILE"];
    else process.env["PREMIERE_MCP_TOKEN_FILE"] = previousPath;
    rmSync(root, { recursive: true, force: true });
  }
});
