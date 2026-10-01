"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const hostSource = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");
const presetPath = "C:\\Presets with spaces\\Editor's 1080p.sqpreset";

function fixture(options = {}) {
    const previous = { sequenceID: "existing", name: "Existing" };
    const sequences = [previous];
    sequences.numSequences = 1;
    const calls = [];
    const project = {
        sequences,
        activeSequence: previous,
        createNewSequenceFromClips() { assert.fail("clip-based creation cannot apply a sequence preset"); },
        deleteSequence(sequence) {
            const index = sequences.indexOf(sequence);
            if (index !== -1) sequences.splice(index, 1);
            sequences.numSequences = sequences.length;
            return true;
        },
    };
    const context = vm.createContext({
        app: {
            project,
            enableQE() {
                calls.push("enableQE");
                if (!options.noQE) {
                    context.qe = { project: {
                        newSequence(name, preset) {
                            calls.push({ name, preset });
                            if (options.create) return options.create(project, name);
                            const created = {
                                sequenceID: "new-sequence",
                                name,
                                frameSizeHorizontal: 1920,
                                frameSizeVertical: 1080,
                                timebase: "10584000000",
                                videoTracks: { numTracks: 3 },
                                audioTracks: { numTracks: 2 },
                            };
                            sequences.unshift(created);
                            sequences.numSequences = sequences.length;
                            // Deliberately keep the original active sequence:
                            // creation must be identified from the collection.
                        },
                    } };
                }
            },
        },
        File: function (name) { this.exists = name === presetPath; },
    });
    vm.runInContext(hostSource, context);
    function dispatch(args) {
        return JSON.parse(context.mcpDispatch("createSequenceFromPreset", JSON.stringify(args)));
    }
    return { context, project, calls, previous, dispatch };
}

for (const field of ["presetPath", "preset_path"]) {
    test("dispatcher maps " + field + " and reads the new sequence identity", () => {
        const { calls, previous, project, dispatch } = fixture();
        const name = "Launch (final) \"cut\"";
        const result = dispatch({ name, [field]: presetPath });
        assert.equal(result.success, true);
        assert.deepEqual(calls, ["enableQE", { name, preset: presetPath }]);
        assert.equal(project.activeSequence, previous);
        assert.deepEqual(result.data, {
            created: true, name, presetPath,
            sequenceID: "new-sequence", sequenceIndex: 0,
            settings: { width: 1920, height: 1080, fps: 24, videoTracks: 3, audioTracks: 2 },
            method: "qe", creationVerified: true,
        });
    });
}

test("missing fields and nonexistent presets fail before enabling QE", () => {
    const { calls, dispatch } = fixture();
    for (const args of [{ presetPath }, { name: "New" }, { name: "New", presetPath: "/missing.sqpreset" }]) {
        assert.equal(dispatch(args).success, false);
    }
    assert.deepEqual(calls, []);
});

test("QE unavailable fails without invoking an unrelated clip-creation fallback", () => {
    const { dispatch } = fixture({ noQE: true });
    const result = dispatch({ name: "New", presetPath });
    assert.equal(result.success, false);
    assert.match(result.error, /requires the QE DOM/);
});

test("a QE success return without a new sequence is rejected", () => {
    const { dispatch } = fixture({ create: () => true });
    const result = dispatch({ name: "New", presetPath });
    assert.equal(result.success, false);
    assert.match(result.error, /expected one new sequence/);
});

test("a new collection entry reusing an existing sequence ID is rejected", () => {
    const { dispatch } = fixture({ create: (project, name) => {
        project.sequences.push({ sequenceID: "existing", name });
        project.sequences.numSequences++;
    } });
    const result = dispatch({ name: "New", presetPath });
    assert.equal(result.success, false);
    assert.match(result.error, /distinct preset sequence ID/);
});

test("an unexpectedly named new sequence is cleaned up without deleting existing work", () => {
    const { dispatch, project, previous } = fixture({ create: (project) => {
        const created = { sequenceID: "wrong-name", name: "Wrong" };
        project.sequences.push(created);
        project.sequences.numSequences++;
        project.activeSequence = created;
    } });
    const result = dispatch({ name: "New", presetPath });
    assert.equal(result.success, false);
    assert.match(result.error, /unexpected name; it was deleted/);
    assert.deepEqual(project.sequences.slice(), [previous]);
    assert.equal(project.activeSequence, previous);
});

test("ambiguous multiple new sequences cannot produce a success result", () => {
    const { dispatch } = fixture({ create: (project, name) => {
        project.sequences.push({ sequenceID: "one", name }, { sequenceID: "two", name });
        project.sequences.numSequences += 2;
    } });
    const result = dispatch({ name: "New", presetPath });
    assert.equal(result.success, false);
    assert.match(result.error, /expected one new sequence/);
});
