"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

function sequence(id, name) {
    return { sequenceID: id, name, frameSizeHorizontal: 1920, frameSizeVertical: 1080,
        timebase: "10584000000", videoTracks: { numTracks: 3 }, audioTracks: { numTracks: 2 } };
}
function run(host, project) {
    const context = vm.createContext({ app: { project } });
    vm.runInContext(fs.readFileSync(path.join(__dirname, "../src/host/" + host + ".jsx"), "utf8"), context);
    return JSON.parse(context.getSequenceList());
}
for (const host of ["core", "premiere"]) {
    test(host + ": native numSequences and canonical metadata survive sequence listing", () => {
        const a = sequence("first", "Cut A"), b = sequence("second", "Cut B");
        const sequences = [a, b]; sequences.numSequences = 2;
        const result = run(host, { sequences, activeSequence: b });
        assert.equal(result.success, true);
        assert.deepEqual(result.data, { count: 2, active_sequence_id: "second", sequences: [
            { index: 0, name: "Cut A", sequence_id: "first", frame_size_horizontal: 1920,
                frame_size_vertical: 1080, timebase: "10584000000", video_track_count: 3, audio_track_count: 2, is_active: false },
            { index: 1, name: "Cut B", sequence_id: "second", frame_size_horizontal: 1920,
                frame_size_vertical: 1080, timebase: "10584000000", video_track_count: 3, audio_track_count: 2, is_active: true }
        ] });
    });
    test(host + ": genuinely empty project reports zero sequences", () => {
        const sequences = []; sequences.numSequences = 0;
        assert.deepEqual(run(host, { sequences, activeSequence: null }), {
            success: true, data: { count: 0, active_sequence_id: "", sequences: [] }
        });
    });
    test(host + ": unreadable collection fails instead of claiming empty project", () => {
        assert.equal(run(host, { sequences: {}, activeSequence: null }).success, false);
    });
    test(host + ": unreadable track counts fail instead of claiming zero tracks", () => {
        const a = sequence("first", "Cut A"); a.videoTracks = {};
        const sequences = [a]; sequences.numSequences = 1;
        assert.equal(run(host, { sequences, activeSequence: a }).success, false);
    });
    test(host + ": missing sequence identity fails", () => {
        const a = sequence("", "Cut A"); const sequences = [a]; sequences.numSequences = 1;
        assert.equal(run(host, { sequences, activeSequence: a }).success, false);
    });
    test(host + ": no project remains an explicit error", () => {
        assert.equal(run(host, null).success, false);
    });
    for (const timebase of [undefined, "", "invalid", "0", "-1"]) test(host + ": unreadable timebase " + timebase + " fails", () => {
        const a = sequence("first", "Cut A"); a.timebase = timebase;
        const sequences = [a]; sequences.numSequences = 1;
        assert.equal(run(host, {sequences, activeSequence:a}).success, false);
    });

}
