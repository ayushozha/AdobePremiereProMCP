"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const { inspectProject } = require("../inspect.js");

// Adobe is the external boundary. These fixtures exercise our normalization
// and error accounting, and do not prove the plugin runs inside Premiere.
function api(sequences) {
  return { Project: { getActiveProject: async () => ({
    name: "Fixture.prproj",
    path: "/fixtures/Fixture.prproj",
    getSequences: async () => sequences,
  }) } };
}

test("an open project reports awaited video, audio and caption counts", async () => {
  const result = await inspectProject(api([{
    name: "Captioned cut",
    getVideoTrackCount: async () => 2,
    getAudioTrackCount: async () => 3,
    getCaptionTrackCount: async () => 1,
  }]), { version: "25.6.0" });
  assert.equal(result.projectOpen, true);
  assert.equal(result.premiereVersion, "25.6.0");
  assert.deepEqual(result.project, { name: "Fixture.prproj", path: "/fixtures/Fixture.prproj" });
  assert.deepEqual(result.sequences, [{ name: "Captioned cut", videoTracks: 2, audioTracks: 3, captionTracks: 1 }]);
  assert.deepEqual(result.warnings, []);
});

test("no project is reported explicitly instead of failing", async () => {
  const result = await inspectProject({ Project: { getActiveProject: async () => null } });
  assert.equal(result.projectOpen, false);
  assert.equal(result.project, null);
  assert.deepEqual(result.sequences, []);
});

test("unavailable caption APIs report unknown counts instead of zero captions", async () => {
  const result = await inspectProject(api([{
    name: "Partial host",
    getVideoTrackCount: async () => 1,
    getAudioTrackCount: async () => 0,
  }]));
  assert.equal(result.sequences[0].captionTracks, null);
  assert.equal(result.status, "partial");
  assert.match(result.warnings.join(" "), /getCaptionTrackCount/);
});

test("invalid and failed reads cannot produce false successful track counts", async () => {
  const result = await inspectProject(api([{
    name: "Changed sequence",
    getVideoTrackCount: async () => -1,
    getAudioTrackCount: async () => { throw new Error("Host handle expired"); },
    getCaptionTrackCount: async () => 0,
  }]));
  assert.equal(result.sequences[0].videoTracks, null);
  assert.equal(result.sequences[0].audioTracks, null);
  assert.equal(result.sequences[0].captionTracks, 0);
  assert.equal(result.warnings.length, 2);
  assert.match(result.warnings.join(" "), /Host handle expired/);
});

test("missing Project API and host failures reject rather than claim no project", async () => {
  await assert.rejects(inspectProject({}), /Project\.getActiveProject/);
  await assert.rejects(inspectProject({ Project: { getActiveProject: async () => {
    throw new Error("Host unavailable");
  } } }), /Host unavailable/);
});
