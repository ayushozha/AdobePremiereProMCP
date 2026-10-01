"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const hostSource = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");

function fixture(options = {}) {
    let amplitude = 1;
    const writes = [];
    const status = Object.hasOwn(options, "status") ? options.status : 0;
    const parameter = {
        displayName: "Level",
        setValue(value, updateUI) {
            writes.push({ value, updateUI });
            if (options.throwWrite) throw new Error("audio write unavailable");
            if (!options.ignoreWrite) amplitude = value;
            return status;
        },
        getValue() {
            if (options.throwRead && writes.length) throw new Error("audio readback unavailable");
            return writes.length && Object.hasOwn(options, "readback") ? options.readback : amplitude;
        },
        getKeys: () => [],
    };
    const properties = [parameter]; properties.numItems = 1;
    properties.getParamForDisplayName = (name) => name === "Level" ? parameter : null;
    const components = [{ displayName: "Volume", matchName: "audioGain", properties }]; components.numItems = 1;
    const clips = [{ nodeId: "audio-clip", components }]; clips.numItems = 1;
    const audioTracks = [{ clips }]; audioTracks.numTracks = 1;
    const videoTracks = []; videoTracks.numTracks = 0;
    const sequence = { sequenceID: "audio-sequence", audioTracks, videoTracks };
    const sequences = [sequence]; sequences.numSequences = 1;
    const context = vm.createContext({ app: { project: { activeSequence: sequence, sequences } } });
    vm.runInContext(hostSource, context);
    function call(name) {
        const raw = name === "mcpSetAudioLevel" ? context[name](JSON.stringify({ clipId: context._mcpCanonicalClipId("audio", 0, 0, clips[0]), levelDb: -6 })) : context[name](0, 0, -6);
        return JSON.parse(raw);
    }
    return { context, parameter, call, writes, amplitude: () => amplitude };
}

for (const name of ["setAudioLevel", "normalizeAudio", "mcpSetAudioLevel"]) {
    for (const status of [true, 0]) {
        test(`${name} accepts strict ${status} only with matching audio readback`, () => {
            const { call, writes, amplitude } = fixture({ status });
            const result = call(name);
            assert.equal(result.success, true, result.error);
            assert.equal(result.data.verified, true);
            assert.ok(Math.abs((result.data.actualLevelDb ?? result.data.levelDb) + 6) < 0.0001);
            assert.ok(Math.abs(amplitude() - Math.pow(10, -6 / 20)) < 0.000001);
            assert.deepEqual(writes, [{ value: Math.pow(10, -6 / 20), updateUI: true }]);
        });
        test(`${name} rejects ${status} when Premiere ignores the requested audio value`, () => {
            const { call } = fixture({ status, ignoreWrite: true });
            const result = call(name);
            assert.equal(result.success, false);
            assert.match(result.error, /read back audio amplitude 1 instead of/);
        });
    }
    for (const status of [false, 1, undefined, null, "0", "true"]) {
        test(`${name} rejects status ${typeof status} ${status} even with matching stored value`, () => {
            const { call, amplitude } = fixture({ status });
            const result = call(name);
            assert.equal(result.success, false);
            assert.match(result.error, /returned failure status/);
            // The host may mutate before reporting failure; a matching value
            // alone must never turn a rejected status into success.
            if (name !== "normalizeAudio") assert.ok(Math.abs(amplitude() - Math.pow(10, -6 / 20)) < 0.000001);
        });
    }
}

for (const readback of [undefined, null, false, NaN, Infinity, "0.501bad"]) {
    test(`mcpSetAudioLevel rejects invalid ${typeof readback} ${readback} readback after native true`, () => {
        const { call } = fixture({ status: true, readback });
        const result = call("mcpSetAudioLevel");
        assert.equal(result.success, false);
        assert.match(result.error, /Audio level readback/);
    });
}

for (const options of [{ throwWrite: true }, { throwRead: true }]) {
    test(`mcpSetAudioLevel reports ${options.throwWrite ? "write" : "readback"} exceptions`, () => {
        const { call } = fixture({ status: true, ...options });
        const result = call("mcpSetAudioLevel");
        assert.equal(result.success, false);
        assert.match(result.error, /unavailable/);
    });
}

test("a missing audio readback method fails before mutation", () => {
    const { call, parameter, writes } = fixture({ status: true });
    delete parameter.getValue;
    const result = call("mcpSetAudioLevel");
    assert.equal(result.success, false);
    assert.match(result.error, /not readable and writable/);
    assert.deepEqual(writes, []);
});

test("normalization restores the original level after a mismatching true write", () => {
    const { call, writes, amplitude } = fixture({ status: true, ignoreWrite: true });
    const result = call("normalizeAudio");
    assert.equal(result.success, false);
    assert.match(result.error, /original level and keyframes were restored/);
    assert.equal(writes.length, 2);
    assert.equal(writes[1].value, 1);
    assert.equal(amplitude(), 1);
});
