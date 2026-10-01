"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const hostSource = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");

function fixture(options = {}) {
    const marks = {
        1: { inPoint: 1, outPoint: 7 },
        2: { inPoint: 0, outPoint: 8 },
    };
    const calls = [];
    const sourcePath = options.audioOnly ? "/tmp/tone.wav" : "/tmp/av.mov";
    const item = {
        type: 1, nodeId: "source", name: "Source",
        getMediaPath: () => sourcePath,
        hasVideo: () => !options.audioOnly,
        hasAudio: () => true,
        getInPoint(mediaType) {
            calls.push(["getIn", mediaType]);
            if (options.failTypedRead && mediaType === 2) throw new Error("typed audio marks unavailable");
            // Model the native silent-default behavior that fooled the old
            // getter(4) verification; only 1/2 select the requested stream.
            return { seconds: (marks[mediaType] || marks[1]).inPoint };
        },
        getOutPoint(mediaType) {
            calls.push(["getOut", mediaType]);
            return { seconds: (marks[mediaType] || marks[1]).outPoint };
        },
        setInPoint(value, mediaType) {
            calls.push(["setIn", mediaType, value]);
            assert.ok(mediaType === 1 || mediaType === 2, "writes must affect only the target stream");
            if (!options.ignoreMarks) marks[mediaType].inPoint = value;
            return 0;
        },
        setOutPoint(value, mediaType) {
            calls.push(["setOut", mediaType, value]);
            assert.ok(mediaType === 1 || mediaType === 2, "writes must affect only the target stream");
            if (!options.ignoreMarks) marks[mediaType].outPoint = value;
            return 0;
        },
    };
    function createTrack(mediaType) {
        const clips = [];
        clips.numItems = 0;
        return {
            clips,
            overwriteClip(source, time) {
                calls.push(["overwrite", mediaType]);
                if (options.failOverwrite) throw new Error("overwrite unavailable");
                if (options.redirectAudioToVideo && mediaType === 2) {
                    videoTrack.overwriteClip(source, time);
                    return;
                }
                const range = marks[mediaType];
                const clip = {
                    nodeId: "placed", projectItem: source,
                    start: { seconds: time.seconds },
                    end: { seconds: time.seconds + range.outPoint - range.inPoint },
                    inPoint: { seconds: options.wrongClipRange ? 0 : range.inPoint },
                    outPoint: { seconds: options.wrongClipRange ? 8 : range.outPoint },
                    getSpeed: () => 1,
                    remove() { clips.splice(clips.indexOf(clip), 1); clips.numItems = clips.length; },
                };
                clips.push(clip);
                clips.numItems = clips.length;
            },
        };
    }
    const videoTrack = createTrack(1);
    const audioTrack = createTrack(2);
    const videoTracks = [videoTrack]; videoTracks.numTracks = 1;
    const audioTracks = [audioTrack]; audioTracks.numTracks = 1;
    const sequence = { sequenceID: "sequence", videoTracks, audioTracks, timebase: "10584000000" };
    const sequences = [sequence]; sequences.numSequences = 1;
    const children = [item]; children.numItems = 1;
    const context = vm.createContext({
        app: { project: { activeSequence: sequence, sequences, rootItem: { children } } },
        ProjectItemType: { BIN: 2 },
        Time: function () { this.seconds = 0; },
    });
    vm.runInContext(hostSource, context);
    function place(trackType, range = { inPoint: 0, outPoint: 4 }) {
        return JSON.parse(context.mcpPlaceClip(JSON.stringify({ sourcePath, track: { type: trackType, trackIndex: 0 }, position: 0, sourceRange: range, speed: 1 })));
    }
    return { context, marks, calls, sourcePath, videoTrack, audioTrack, place };
}

test("WAV placement reads, sets, and restores audio marks with selector 2", () => {
    const { place, marks, calls, audioTrack } = fixture({ audioOnly: true });
    const result = place("audio");
    assert.equal(result.success, true, result.error);
    assert.match(result.data.clipId, /^audio:0:0:/);
    assert.equal(audioTrack.clips[0].inPoint.seconds, 0);
    assert.equal(audioTrack.clips[0].outPoint.seconds, 4);
    assert.equal(audioTrack.clips[0].end.seconds, 4);
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
    assert.ok(calls.filter((call) => call[0] !== "overwrite").every((call) => call[1] === 2));
});

test("audio placement from AV preserves distinct video marks and omitted audio boundary", () => {
    const { place, marks, audioTrack } = fixture();
    const result = place("audio", { outPoint: 4 });
    assert.equal(result.success, true, result.error);
    assert.equal(audioTrack.clips[0].inPoint.seconds, 0, "preserve original audio in point instead of video in point");
    assert.equal(audioTrack.clips[0].outPoint.seconds, 4);
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
});

test("video placement uses selector 1 while preserving audio marks", () => {
    const { place, marks, calls, videoTrack } = fixture();
    const result = place("video", { inPoint: 2, outPoint: 5 });
    assert.equal(result.success, true, result.error);
    assert.equal(videoTrack.clips[0].inPoint.seconds, 2);
    assert.equal(videoTrack.clips[0].outPoint.seconds, 5);
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
    assert.ok(calls.filter((call) => call[0] !== "overwrite").every((call) => call[1] === 1));
});

test("ignored audio source writes cannot proceed to a timeline overwrite", () => {
    const { place, marks, calls, audioTrack } = fixture({ ignoreMarks: true });
    const result = place("audio");
    assert.equal(result.success, false);
    assert.match(result.error, /did not retain/);
    assert.equal(audioTrack.clips.length, 0);
    assert.ok(!calls.some((call) => call[0] === "overwrite"));
    assert.deepEqual(marks[2], { inPoint: 0, outPoint: 8 });
});

test("overwrite failure restores audio marks and leaves video marks untouched", () => {
    const { place, marks } = fixture({ failOverwrite: true });
    const result = place("audio");
    assert.equal(result.success, false);
    assert.match(result.error, /overwrite unavailable/);
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
});

test("wrong inserted audio range is removed even when project source marks were applied", () => {
    const { place, marks, audioTrack } = fixture({ wrongClipRange: true });
    const result = place("audio");
    assert.equal(result.success, false);
    assert.match(result.error, /did not apply/);
    assert.equal(audioTrack.clips.length, 0);
    assert.deepEqual(marks[2], { inPoint: 0, outPoint: 8 });
});

test("EDL audio assembly uses and restores the audio stream range", () => {
    const { context, marks, sourcePath, audioTrack, calls } = fixture({ audioOnly: true });
    const result = JSON.parse(context.assembleFromEDL(JSON.stringify({ clips: [{ file: sourcePath, autoImport: false, trackType: "audio", trackIndex: 0, position: 0, inPoint: 0, outPoint: 4 }] })));
    assert.equal(result.success, true, result.error);
    assert.deepEqual(result.data.errors, []);
    assert.equal(result.data.placed, 1);
    assert.equal(audioTrack.clips[0].outPoint.seconds, 4);
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
    assert.ok(calls.filter((call) => call[0] !== "overwrite").every((call) => call[1] === 2));
});

test("typed audio source read failure stops placement without an untyped fallback", () => {
    const { place, calls, marks } = fixture({ failTypedRead: true });
    const result = place("audio");
    assert.equal(result.success, false);
    assert.match(result.error, /typed audio marks unavailable/);
    assert.deepEqual(calls, [["getIn", 2]]);
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
});

test("EDL placement on the wrong media track cannot report success", () => {
    const { context, sourcePath, marks, audioTrack, videoTrack } = fixture({ redirectAudioToVideo: true });
    const result = JSON.parse(context.mcpExecuteEDL(JSON.stringify({
        autoImport: false, autoCreateSequence: false,
        edl: { entries: [{ sourceAssetId: sourcePath, track: { type: "audio", trackIndex: 0 }, sourceRange: { inPoint: 0, outPoint: 4 }, timelineRange: { inPoint: 0, outPoint: 4 } }] },
    })));
    assert.equal(result.success, true, result.error);
    assert.equal(result.data.status, "failed");
    assert.equal(result.data.clipsPlaced, 0);
    assert.match(result.data.errors.join(" "), /clip could not be located/i);
    assert.equal(audioTrack.clips.length, 0);
    assert.equal(videoTrack.clips.length, 1, "native wrong-track mutation must not be counted as requested audio placement");
    assert.deepEqual(marks, { 1: { inPoint: 1, outPoint: 7 }, 2: { inPoint: 0, outPoint: 8 } });
});
