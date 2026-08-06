import assert from "node:assert/strict";
import test from "node:test";

import { parseDotEnv } from "./environment.js";

test("parses quoted dotenv values and ignores comments", () => {
  assert.deepEqual(
    parseDotEnv("# comment\nOPENAI_API_KEY='sk-test'\nexport MODEL=example\ninvalid line\n"),
    [
      ["OPENAI_API_KEY", "sk-test"],
      ["MODEL", "example"],
    ],
  );
});
