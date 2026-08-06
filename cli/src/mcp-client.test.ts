import assert from "node:assert/strict";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { buildServerEnvironment, collectToolPages } from "./mcp-client.js";

test("collects every cursor-paginated MCP tool page", async () => {
  const cursors: Array<string | undefined> = [];
  const tools = await collectToolPages(async (cursor) => {
    cursors.push(cursor);
    if (cursor === undefined) {
      return {
        tools: [{ name: "one", description: "first", inputSchema: {} }],
        nextCursor: "page-2",
      };
    }
    if (cursor === "page-2") {
      return {
        tools: [{ name: "two", inputSchema: { type: "object" } }],
        nextCursor: "page-3",
      };
    }
    return {
      tools: [{ name: "three", description: "last", inputSchema: {} }],
    };
  });

  assert.deepEqual(cursors, [undefined, "page-2", "page-3"]);
  assert.deepEqual(tools.map((tool) => tool.name), ["one", "two", "three"]);
  assert.equal(tools[1]?.description, "");
});

test("rejects a repeated pagination cursor", async () => {
  await assert.rejects(
    collectToolPages(async () => ({ tools: [], nextCursor: "same" })),
    /repeated tools\/list cursor/,
  );
});

test("loads repository MCP settings while preserving explicit environment overrides", () => {
  const repositoryRoot = mkdtempSync(join(tmpdir(), "premierpro-cli-env-"));
  try {
    writeFileSync(
      join(repositoryRoot, ".env"),
      "MCP_TOOL_PROFILE=captions\nMCP_TRANSPORT=sse\nTS_BRIDGE_ADDR=localhost:59999\nUNRELATED_SECRET=do-not-forward\n",
    );
    const environment = buildServerEnvironment(repositoryRoot, {
      MCP_TOOL_PROFILE: "effects",
    });

    assert.equal(environment["MCP_TOOL_PROFILE"], "effects");
    assert.equal(environment["TS_BRIDGE_ADDR"], "localhost:59999");
    assert.equal(environment["MCP_TRANSPORT"], "stdio");
    assert.equal(environment["UNRELATED_SECRET"], undefined);
  } finally {
    rmSync(repositoryRoot, { recursive: true, force: true });
  }
});
