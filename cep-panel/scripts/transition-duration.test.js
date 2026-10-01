"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const source = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");
const ticksPerSecond = 254016000000;

function collection(items, countKey) {
    Object.defineProperty(items, countKey, { get() { return items.length; } });
    return items;
}

function fixture(fps = 24, options = {}) {
    const calls = [];
    const transitions = collection(options.existing ? [
        { nodeId: "existing", displayName: "Dip to Black", start: { seconds: 1 }, end: { seconds: 2 } },
    ] : [], "numItems");
    const clips = Array.from({ length: options.clipCount || 1 }, (_, index) => ({ nodeId: "clip-" + index }));
    const track = { transitions, clips: collection(clips, "numItems") };
    const sequence = {
        videoTracks: collection([track], "numTracks"),
        audioTracks: collection([track], "numTracks"),
        timebase: options.unknownRate ? "" : String(ticksPerSecond / fps),
    };
    if (options.settingsRate) {
        sequence.getSettings = () => ({ videoFrameRate: { ticks: String(ticksPerSecond / options.settingsRate) } });
    }
    const nativeFPS = options.settingsRate || fps;
    const qeClip = {
        addTransition(transition, applyToEnd, qeDuration, clipIndex) {
            calls.push({ transition, applyToEnd, qeDuration, clipIndex });
            // Model the observed QE contract independently of the seconds API:
            // the suffix is an integer frame field, so "0.5" is FIVE frames.
            assert.match(qeDuration, /^\d+\.\d+$/);
            const [seconds, frameField] = qeDuration.split(".").map(Number);
            const nativeFrameCount = seconds * Math.round(nativeFPS) + frameField;
            if (options.noop || (options.noopIndices && options.noopIndices.includes(clipIndex))) return true;
            const appliedFrames = options.appliedFrames === undefined ? nativeFrameCount : options.appliedFrames;
            const duration = appliedFrames / nativeFPS;
            const start = options.start === undefined ? 4 : options.start;
            transitions.push({ nodeId: "new-" + transitions.length, displayName: transition.name,
                start: { seconds: start }, end: { seconds: start + duration }, duration: { seconds: duration } });
            return true;
        },
    };
    const qeTrack = { getItemAt(index) {
        assert.ok(Number.isInteger(index) && index >= 0 && index < clips.length);
        return { addTransition(...args) { return qeClip.addTransition(...args, index); } };
    } };
    const context = vm.createContext({
        app: { project: { activeSequence: sequence }, enableQE() { calls.push("enableQE"); } },
        qe: { project: {
            getActiveSequence() { return { getVideoTrackAt() { return qeTrack; }, getAudioTrackAt() { return qeTrack; } }; },
            getVideoTransitionByName(name) { return { name }; },
            getAudioTransitionByName(name) { return { name }; },
        } },
    });
    vm.runInContext(source, context);
    return {
        calls, transitions, sequence,
        apply(type = "video", duration = 0.5, applyToEnd = true) {
            return JSON.parse(type === "audio" ? context.addAudioTransition(0, 0, "Constant Power", duration) :
                context.addVideoTransition(0, 0, "Cross Dissolve", duration, applyToEnd));
        },
        applyDefaults(type) {
            return JSON.parse(type === "audio" ? context.addAudioTransition(0, 0) : context.addVideoTransition(0, 0));
        },
        dispatch(command, duration = 0.5, extra = {}) {
            return JSON.parse(context.mcpDispatch(command, JSON.stringify({
                track_index: 0, clip_index: 0, transition_name: "Cross Dissolve", duration, ...extra,
            })));
        },
        readback(type = "video") { return JSON.parse(context.getTransitions(type, 0)); },
    };
}

const cases = [
    { fps: 24, seconds: 0.5, qeDuration: "0.12", frames: 12 },
    { fps: 25, seconds: 0.5, qeDuration: "0.13", frames: 13 },
    { fps: 30, seconds: 0.5, qeDuration: "0.15", frames: 15 },
    { fps: 24000 / 1001, seconds: 0.5, qeDuration: "0.12", frames: 12 },
    { fps: 30000 / 1001, seconds: 0.5, qeDuration: "0.15", frames: 15 },
    { fps: 60000 / 1001, seconds: 1.25, qeDuration: "1.15", frames: 75 },
    { fps: 24, seconds: 1.5, qeDuration: "1.12", frames: 36 },
    { fps: 24, seconds: 0.999, qeDuration: "1.00", frames: 24 },
    { fps: 24, seconds: 1 / 24, qeDuration: "0.01", frames: 1 },
];

for (const type of ["video", "audio"]) {
    for (const example of cases) {
        test(`${type}: ${example.seconds}s at ${example.fps}fps encodes ${example.frames} frames and reads native duration`, () => {
            const host = fixture(example.fps);
            const result = host.apply(type, example.seconds);
            assert.equal(result.success, true, result.error);
            const nativeCall = host.calls.find(call => typeof call === "object");
            assert.equal(nativeCall.qeDuration, example.qeDuration);
            assert.equal(result.data.qeDuration, example.qeDuration);
            assert.equal(result.data.requestedFrames, example.frames);
            assert.equal(result.data.requestedDuration, example.seconds);
            assert.ok(Math.abs(result.data.frameRate - example.fps) < 1e-9);
            assert.ok(Math.abs(result.data.quantizedDuration - example.frames / example.fps) < 1e-9);
            assert.ok(Math.abs(result.data.duration - example.frames / example.fps) < 1e-9);
            assert.equal(result.data.verified, true);
            const readback = host.readback(type);
            assert.equal(readback.success, true);
            assert.ok(Math.abs(readback.data.transitions[0].duration - result.data.duration) < 1e-9);
        });
    }
}

for (const type of ["video", "audio"]) {
    test(type + ": omitted duration defaults to one second encoded as seconds.frames", () => {
        const result = fixture().applyDefaults(type);
        assert.equal(result.success, true);
        assert.equal(result.data.requestedDuration, 1);
        assert.equal(result.data.requestedFrames, 24);
        assert.equal(result.data.qeDuration, "1.00");
        assert.equal(result.data.duration, 1);
    });
}

test("settings frame-rate readback takes precedence over a stale sequence timebase", () => {
    const host = fixture(24, { settingsRate: 30 });
    const result = host.apply();
    assert.equal(result.success, true);
    assert.equal(result.data.qeDuration, "0.15");
    assert.equal(result.data.frameRate, 30);
    assert.equal(result.data.requestedFrames, 15);
    assert.equal(result.data.duration, 0.5);
});

test("a positive subframe duration is rounded to one frame", () => {
    const result = fixture().apply("video", 0.001);
    assert.equal(result.success, true);
    assert.equal(result.data.requestedFrames, 1);
    assert.equal(result.data.qeDuration, "0.01");
    assert.ok(Math.abs(result.data.duration - 1 / 24) < 1e-9);
    assert.equal(result.data.durationAdjusted, true);
});

for (const duration of [0, -0.5, "bad", "0.5seconds", Infinity, NaN]) {
    test(`invalid transition duration ${String(duration)} fails before native mutation`, () => {
        const host = fixture();
        const result = host.apply("video", duration);
        assert.equal(result.success, false);
        assert.match(result.error, /duration must be a positive number/);
        assert.deepEqual(host.calls, []);
        assert.equal(host.transitions.numItems, 0);
    });
}

test("unknown sequence frame rate fails before native mutation", () => {
    const host = fixture(24, { unknownRate: true });
    const result = host.apply();
    assert.equal(result.success, false);
    assert.match(result.error, /Cannot read the sequence frame rate/);
    assert.deepEqual(host.calls, []);
});

test("the applied duration comes from native readback even when source handles shorten it", () => {
    const host = fixture(24, { appliedFrames: 8, existing: true, start: 9 });
    const result = host.apply();
    assert.equal(result.success, true);
    assert.equal(result.data.transitionIndex, 1);
    assert.equal(result.data.requestedFrames, 12);
    assert.equal(result.data.quantizedDuration, 0.5);
    assert.ok(Math.abs(result.data.duration - 8 / 24) < 1e-9);
    assert.equal(result.data.durationAdjusted, true);
    assert.equal(host.transitions[0].nodeId, "existing");
});

test("a successful QE return cannot replace missing transition readback", () => {
    const host = fixture(24, { noop: true });
    const result = host.apply();
    assert.equal(result.success, false);
    assert.match(result.error, /did not add a verifiable transition/);
    assert.equal(host.transitions.numItems, 0);
});

test("a zero-length native transition cannot be reported as verified", () => {
    const result = fixture(24, { appliedFrames: 0 }).apply();
    assert.equal(result.success, false);
    assert.match(result.error, /duration could not be verified/);
});

test("transition placement still passes the caller's existing flag to QE", () => {
    const host = fixture();
    assert.equal(host.apply("video", 0.5, false).success, true);
    assert.equal(host.calls.find(call => typeof call === "object").applyToEnd, false);
});

for (const command of ["addTransition", "addAudioCrossfade", "applyTransitionToAllCuts", "batchApplyTransitions"]) {
    for (const example of [cases[0], cases[3], cases[5]]) {
        test(`${command} dispatcher uses exact frame encoding at ${example.fps}fps`, () => {
            const bulk = command === "applyTransitionToAllCuts" || command === "batchApplyTransitions";
            const host = fixture(example.fps, { clipCount: bulk ? 3 : 1 });
            const result = host.dispatch(command, example.seconds, { type: "constant_gain" });
            assert.equal(result.success, true, result.error);
            assert.equal(result.data.verified, true);
            const nativeCalls = host.calls.filter(call => typeof call === "object");
            assert.equal(nativeCalls.length, bulk ? 2 : 1);
            for (const call of nativeCalls) {
                assert.equal(call.qeDuration, example.qeDuration);
                assert.equal(typeof call.transition, "object");
                assert.equal(call.applyToEnd, !bulk);
            }
            const actual = bulk ? result.data.transitions : [result.data];
            for (const transition of actual) {
                assert.equal(transition.requestedFrames, example.frames);
                assert.equal(transition.requestedDuration, example.seconds);
                assert.ok(Math.abs(transition.duration - example.frames / example.fps) < 1e-9);
            }
            if (bulk) {
                assert.equal(result.data.cutsProcessed, 2);
                assert.equal(result.data.transitionsApplied, 2);
            }
            if (command === "batchApplyTransitions") {
                assert.equal(result.data.applied, 2);
                assert.equal(result.data.totalCuts, 2);
            }
            if (command === "addAudioCrossfade") {
                assert.equal(result.data.type, "Constant Gain");
                assert.equal(result.data.transitionName, "Constant Gain");
            }
        });
    }

    test(command + " fails closed without a sequence frame rate", () => {
        const host = fixture(24, { unknownRate: true, clipCount: 3 });
        const result = host.dispatch(command);
        assert.equal(result.success, false);
        assert.match(result.error, /Cannot read the sequence frame rate/);
        assert.deepEqual(host.calls, []);
    });

    test(command + " rejects a successful QE no-op without native readback", () => {
        const host = fixture(24, { noop: true, clipCount: 3 });
        const result = host.dispatch(command);
        assert.equal(result.success, false);
        assert.match(result.error, /did not add a verifiable transition/);
        assert.equal(host.transitions.numItems, 0);
    });
}

for (const command of ["applyTransitionToAllCuts", "batchApplyTransitions"]) {
    test(command + " reports verified partial application when a later cut fails", () => {
        const host = fixture(24, { clipCount: 3, noopIndices: [1] });
        const result = host.dispatch(command);
        assert.equal(result.success, false);
        assert.match(result.error, /Applied 1 of 2 transitions with readback/);
        assert.match(result.error, /Clip 1.*did not add a verifiable transition/);
        assert.equal(host.transitions.numItems, 1);
    });
}
