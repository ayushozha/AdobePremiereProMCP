"use strict";
const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const os = require("node:os");
const path = require("node:path");
const { exportSourceTranscript } = require("../transcript-export.js");
const raw = ' {"language":"en-us","segments":[{"start":1,"duration":2,"language":"en-us","speaker":"speaker-1","words":[{"text":"Hello.","start":1,"duration":2,"confidence":1,"eos":true,"tags":[],"type":"word"}]}],"speakers":[{"id":"speaker-1","name":"Person"}]}\n';
function boundary(options = {}) {
  const clip = { name: "Interview.mov", isSequence: async () => Boolean(options.sequence) };
  const premiere = {
    Project: { getActiveProject: async () => options.noProject ? null : {} },
    ProjectUtils: { getSelection: async () => ({ getItems: async () => options.items || [clip] }) },
    ClipProjectItem: { cast: () => options.bin ? null : clip },
    Transcript: { exportToJSON: async () => {
      if (options.exportError) throw new Error("Host export failed");
      return options.raw === undefined ? raw : options.raw;
    } },
  };
  if (options.hasTranscript !== undefined) premiere.Transcript.hasTranscript = () => options.hasTranscript;
  const storage = { formats: { utf8: "utf8" }, localFileSystem: {
    getFileForSaving: async (name, opts) => {
      assert.equal(name, "source-transcript.json");
      assert.deepEqual(opts, { types: ["json"] });
      if (options.pickerError) throw new Error("Picker denied");
      if (options.cancel) return null;
      return { nativePath: options.filePath, write: async (data, format) => {
        assert.deepEqual(format, { format: "utf8" });
        if (options.writeError) throw new Error("Disk full");
        await fs.writeFile(options.filePath, data, "utf8");
      }, read: async () => options.badRead ? "truncated" : fs.readFile(options.filePath, "utf8") };
    },
  } };
  return { premiere, storage };
}
test("exports exact raw source JSON without requiring newer hasTranscript API", async () => {
  assert.equal(typeof exportSourceTranscript, "function", "source transcript export is missing");
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), "source-transcript-"));
  try {
    const filePath = path.join(dir, "transcript.json");
    const b = boundary({ filePath });
    const result = await exportSourceTranscript(b.premiere, b.storage);
    assert.equal(await fs.readFile(filePath, "utf8"), raw);
    assert.deepEqual(result, { status: "exported", source: "source_clip_transcript", clipName: "Interview.mov", outputPath: filePath, segmentCount: 1, verified: true });
  } finally { await fs.rm(dir, { recursive: true, force: true }); }
});
for (const [name, options, expected] of [
  ["no project", { noProject: true }, /No project/],
  ["no selection", { items: [] }, /exactly one/],
  ["multiple selection", { items: [{}, {}] }, /exactly one/],
  ["bin selection", { bin: true }, /source media clip/],
  ["sequence selection", { sequence: true }, /source media clip/],
  ["no transcript", { hasTranscript: false }, /existing transcript/],
  ["empty response", { raw: "" }, /empty/],
  ["malformed JSON", { raw: "{" }, /JSON/],
  ["wrong shape", { raw: "[]" }, /segments/],
  ["empty segments", { raw: '{"segments":[],"speakers":[]}' }, /segments/],
  ["host error", { exportError: true }, /Host export failed/],
  ["picker error", { pickerError: true }, /Picker denied/],
  ["write error", { writeError: true }, /Disk full/],
  ["readback mismatch", { badRead: true }, /readback/],
]) test("rejects " + name + " without a successful export", async () => {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), "source-transcript-"));
  try {
    const b = boundary({ ...options, filePath: path.join(dir, "out.json") });
    await assert.rejects(exportSourceTranscript(b.premiere, b.storage), expected);
  } finally { await fs.rm(dir, { recursive: true, force: true }); }
});
test("picker cancellation is explicit", async () => {
  const b = boundary({ cancel: true });
  assert.deepEqual(await exportSourceTranscript(b.premiere, b.storage), { status: "cancelled" });
});
test("missing documented API fails before attempting export", async () => {
  const b = boundary();
  delete b.premiere.Transcript.exportToJSON;
  await assert.rejects(exportSourceTranscript(b.premiere, b.storage), /exportToJSON/);
});
