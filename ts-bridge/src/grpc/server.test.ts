import * as grpc from "@grpc/grpc-js";
import assert from "node:assert/strict";
import test from "node:test";

import { isAuthorizedMetadata } from "./server.js";

const TOKEN = "0123456789abcdef0123456789abcdef";

test("requires the shared bearer token on Premiere mutation RPCs", () => {
  const valid = new grpc.Metadata();
  valid.set("authorization", `Bearer ${TOKEN}`);
  assert.equal(isAuthorizedMetadata(valid, TOKEN), true);

  const missing = new grpc.Metadata();
  assert.equal(isAuthorizedMetadata(missing, TOKEN), false);

  const wrong = new grpc.Metadata();
  wrong.set("authorization", `Bearer ${"f".repeat(32)}`);
  assert.equal(isAuthorizedMetadata(wrong, TOKEN), false);
});
