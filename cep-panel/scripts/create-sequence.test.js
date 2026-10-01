"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const source = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");
const requested = { name: "Fixture \"cut\"", width: 320, height: 180, fps: 24, videoTracks: 1, audioTracks: 1 };

function collection(items, countKey) {
    Object.defineProperty(items, countKey, { get() { return items.length; } });
    return items;
}

function fixture(options = {}) {
    const seed = { type: 4, nodeId: "video-seed", name: "Seed.mp4",
        getMediaPath() { return "/fixture.mp4"; }, hasVideo() { return true; },
        isOffline() { return false; }, sourceIn: 1, sourceOut: 5 };
    const audio = { type: 4, nodeId: "audio-seed", getMediaPath() { return "/fixture.wav"; }, hasVideo() { return false; } };
    const offline = { ...seed, nodeId: "offline", isOffline() { return true; } };
    const bin = { type: 2, children: collection([offline, seed], "numItems") };
    const existingClip = { projectItem: seed, remove() { assert.fail("existing clips must remain untouched"); } };
    const previous = { sequenceID: "existing", name: "Existing", videoTracks: collection([{ clips: collection([existingClip], "numItems") }], "numTracks") };
    const sequences = collection([previous], "numSequences");
    const calls = [];
    let created;
    const project = {
        sequences, activeSequence: previous,
        rootItem: { children: collection(options.empty ? [] : [audio, bin], "numItems") },
        createNewSequence() { assert.fail("the modal New Sequence API must never run"); },
        createNewSequenceFromClips(name, seeds, destination) {
            calls.push("create");
            assert.equal(seeds.length, 1);
            assert.equal(seeds[0], seed);
            assert.equal(destination, project.rootItem);
            if (options.noNewSequence) return previous;
            let settings = { videoFrameWidth: 1280, videoFrameHeight: 720, videoFrameRate: { ticks: "8467200000" } };
            function track() {
                const clips = collection([], "numItems");
                const clip = { projectItem: options.foreignSeed ? audio : seed,
                    remove(ripple, align) {
                        calls.push("remove-seed");
                        assert.equal(ripple, false);
                        assert.equal(align, false);
                        if (!options.ignoreRemove) clips.splice(clips.indexOf(clip), 1);
                    } };
                clips.push(clip);
                return { clips };
            }
            created = { sequenceID: "created", name, timebase: "8467200000",
                videoTracks: collection([track()], "numTracks"), audioTracks: collection([track()], "numTracks"),
                getSettings() { return { ...settings, videoFrameRate: { ...settings.videoFrameRate } }; },
                setSettings(value) {
                    calls.push("settings");
                    if (!options.ignoreSettings) {
                        settings = value;
                        created.timebase = value.videoFrameRate.ticks;
                    }
                    if (options.nameAfterSettings !== undefined) created.name = options.nameAfterSettings;
                    return true;
                } };
            let sequenceName = options.initialName === undefined ? name : options.initialName;
            Object.defineProperty(created, "name", {
                get() { return sequenceName; },
                set(value) {
                    calls.push("rename");
                    if (options.throwRename) throw new Error("Name property rejected");
                    if (!options.ignoreRename) sequenceName = value;
                },
            });
            // Creation need not activate it: identity must come from the list.
            sequences.unshift(created);
            if (options.ambiguous) sequences.push({ sequenceID: "unexpected", name });
            return options.wrongReturn ? previous : created;
        },
        openSequence(id) {
            calls.push("activate");
            assert.equal(id, "created");
            project.activeSequence = options.ignoreActivation ? previous : created;
            return true;
        },
        deleteSequence(sequence) {
            calls.push("rollback");
            assert.equal(sequence, created);
            assert.notEqual(sequence, previous);
            if (options.ignoreDelete) return false;
            sequences.splice(sequences.indexOf(sequence), 1);
            return true;
        },
    };
    if (options.unsupported) delete project.createNewSequenceFromClips;
    const context = vm.createContext({
        app: { project, enableQE() { throw new Error("QE unavailable in this fixture"); } },
        ProjectItemType: { BIN: 2, CLIP: 1, FILE: 4 },
        Time: function () { this.ticks = "0"; this.seconds = 0; },
    });
    vm.runInContext(source, context);
    return { project, previous, seed, calls,
        dispatch(args = requested) { return JSON.parse(context.mcpDispatch("createSequence", JSON.stringify(args))); } };
}

test("creates an empty sequence without modal APIs and reads settings back", () => {
    const { dispatch, calls, project, seed, previous } = fixture();
    const result = dispatch();
    assert.equal(result.success, true);
    assert.deepEqual(result.data, { name: requested.name, sequenceID: "created", width: 320, height: 180,
        fps: 24, videoTrackCount: 1, audioTrackCount: 1, timebase: "10584000000", empty: true, verified: true });
    assert.deepEqual(calls, ["create", "rename", "remove-seed", "remove-seed", "activate", "settings"]);
    assert.equal(project.activeSequence.sequenceID, "created");
    assert.equal(project.activeSequence.videoTracks[0].clips.numItems, 0);
    assert.equal(project.activeSequence.audioTracks[0].clips.numItems, 0);
    assert.equal(previous.videoTracks[0].clips.numItems, 1);
    assert.deepEqual([seed.sourceIn, seed.sourceOut], [1, 5]);
});

for (const name of [
    'Sequence - "e2e_test_pattern.mp4"',
    'Ayush’s edit (第１版) — café 🎬 [cut]; "take 2".mp4',
]) {
    test("restores the exact requested name after native creation mangles " + JSON.stringify(name), () => {
        const { dispatch, calls, project, previous } = fixture({ initialName: "Native sanitized filename" });
        const result = dispatch({ ...requested, name });
        assert.equal(result.success, true);
        assert.equal(result.data.name, name);
        assert.equal(project.activeSequence.name, name);
        assert.equal(previous.name, "Existing");
        assert.equal(calls.filter(call => call === "rename").length, 1);
        assert.equal(calls.includes("rollback"), false);
    });
}

for (const option of ["ignoreRename", "throwRename"]) {
    test(option + " reports exact expected and actual names and rolls back only its new sequence", () => {
        const initialName = 'Native \"wrong\" 名.mp4';
        const { dispatch, project, previous, calls } = fixture({ initialName, [option]: true });
        const result = dispatch();
        assert.equal(result.success, false);
        assert.ok(result.error.includes("expected " + JSON.stringify(requested.name)));
        assert.ok(result.error.includes("got " + JSON.stringify(initialName)));
        assert.match(result.error, /invalid sequence was deleted/);
        assert.deepEqual(project.sequences.slice(), [previous]);
        assert.equal(project.activeSequence, previous);
        assert.equal(previous.name, "Existing");
        assert.equal(calls.filter(call => call === "rollback").length, 1);
        assert.equal(calls.includes("remove-seed"), false);
    });
}

test("final name readback detects a name changed after settings were applied", () => {
    const actualName = "Settings changed the name";
    const { dispatch, project, previous } = fixture({ nameAfterSettings: actualName });
    const result = dispatch();
    assert.equal(result.success, false);
    assert.ok(result.error.includes("expected " + JSON.stringify(requested.name)));
    assert.ok(result.error.includes("got " + JSON.stringify(actualName)));
    assert.deepEqual(project.sequences.slice(), [previous]);
    assert.equal(project.activeSequence, previous);
});

test("ignored sequence settings fail readback and roll back only the new sequence", () => {
    const { dispatch, project, previous, calls } = fixture({ ignoreSettings: true });
    const result = dispatch();
    assert.equal(result.success, false);
    assert.match(result.error, /did not apply.*invalid sequence was deleted/);
    assert.deepEqual(project.sequences.slice(), [previous]);
    assert.equal(project.activeSequence, previous);
    assert.equal(calls.filter(c => c === "rollback").length, 1);
});

for (const option of ["ignoreRemove", "foreignSeed", "ignoreActivation"]) {
    test(option + " cannot produce a verified empty sequence", () => {
        const { dispatch, project, previous } = fixture({ [option]: true });
        const result = dispatch();
        assert.equal(result.success, false);
        assert.match(result.error, /invalid sequence was deleted/);
        assert.deepEqual(project.sequences.slice(), [previous]);
    });
}

for (const option of ["noNewSequence", "ambiguous", "wrongReturn"]) {
    test(option + " leaves unowned sequences untouched and reports failure", () => {
        const { dispatch, calls, previous, project } = fixture({ [option]: true });
        const result = dispatch();
        assert.equal(result.success, false);
        assert.match(result.error, /no cleanup was attempted/);
        assert.equal(calls.includes("rollback"), false);
        assert.equal(calls.includes("rename"), false);
        assert.ok(project.sequences.includes(previous));
    });
}

for (const option of ["empty", "unsupported"]) {
    test(option + " returns actionable instructions without opening a dialog", () => {
        const { dispatch, calls } = fixture({ [option]: true });
        const result = dispatch();
        assert.equal(result.success, false);
        assert.match(result.error, /createSequenceFromPreset/);
        assert.deepEqual(calls, []);
    });
}

test("a failed rollback remains explicit", () => {
    const { dispatch, project } = fixture({ ignoreRemove: true, ignoreDelete: true });
    const result = dispatch();
    assert.equal(result.success, false);
    assert.match(result.error, /cleanup failed.*inspect the project/);
    assert.equal(project.sequences.numSequences, 2);
});

test("core-only legacy calls fail without using the modal creation API", () => {
    const context = vm.createContext({ app: { project: { createNewSequence() { assert.fail("modal API"); } } } });
    vm.runInContext(fs.readFileSync(path.join(__dirname, "../src/host/core.jsx"), "utf8"), context);
    const result = JSON.parse(context.createSequence(JSON.stringify(requested)));
    assert.equal(result.success, false);
    assert.match(result.error, /full host dispatcher/);
});

test("the legacy panel creation action uses the verified full dispatcher", () => {
    const panel = fs.readFileSync(path.join(__dirname, "../src/panel.js"), "utf8");
    const map = panel.match(/var ACTION_MAP = \{[\s\S]*?\n    \};/)[0];
    const context = vm.createContext({ premiereJsxPath: "/extension/src/host/premiere.jsx", hostLoader: {
        buildDispatchScript(host, name, args) { return { host, name, args: JSON.parse(args) }; },
    } });
    vm.runInContext(map, context);
    const call = context.ACTION_MAP.createSequence(requested);
    assert.equal(call.name, "createSequence");
    assert.equal(call.host, "/extension/src/host/premiere.jsx");
    assert.deepEqual(call.args, requested);
});
