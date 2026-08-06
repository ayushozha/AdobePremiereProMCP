#!/usr/bin/env node

"use strict";

var assert = require("assert");
var fs = require("fs");
var path = require("path");
var vm = require("vm");

var hostPath = path.resolve(__dirname, "..", "src", "host", "premiere.jsx");
var hostSource = fs.readFileSync(hostPath, "utf8");
var context = vm.createContext({
    console: console,
    JSON: JSON,
    Date: Date,
    Math: Math,
    File: function MockFile() {},
    $: { os: "Macintosh" }
});
vm.runInContext(hostSource, context, { filename: hostPath });

function call(name) {
    var args = Array.prototype.slice.call(arguments, 1);
    return JSON.parse(context[name].apply(null, args));
}

function expectFailure(result, pattern) {
    assert.strictEqual(result.success, false, JSON.stringify(result));
    assert.match(result.error, pattern);
}

function makeSequenceCollection(items) {
    items.numSequences = items.length;
    return items;
}

function makeProject(items, activeSequence) {
    return {
        activeSequence: activeSequence,
        sequences: items,
        openSequence: function (sequenceID) {
            for (var i = 0; i < this.sequences.numSequences; i++) {
                if (String(this.sequences[i].sequenceID || "") === String(sequenceID)) {
                    this.activeSequence = this.sequences[i];
                    return true;
                }
            }
            return false;
        }
    };
}

var sourceSequence;
var sequenceCollection;
var capturedAutoReframeArgs;
sourceSequence = {
    name: "Master",
    sequenceID: "source-1",
    videoTracks: { numTracks: 1 },
    audioTracks: { numTracks: 1 },
    autoReframeSequence: function (numerator, denominator, motionPreset, newName, useNestedSequences) {
        capturedAutoReframeArgs = Array.prototype.slice.call(arguments);
        var derivative = {
            name: newName,
            sequenceID: "derivative-1",
            videoTracks: { numTracks: 1 },
            audioTracks: { numTracks: 1 },
            getSettings: function () {
                return { videoFrameWidth: 1080, videoFrameHeight: 1920 };
            }
        };
        sequenceCollection.push(derivative);
        sequenceCollection.numSequences = sequenceCollection.length;
        return derivative;
    }
};
sequenceCollection = makeSequenceCollection([sourceSequence]);
context.app = {
    project: makeProject(sequenceCollection, sourceSequence)
};

var reframed = call("autoReframeSequence", JSON.stringify({
    sourceSequenceID: "source-1",
    numerator: 9,
    denominator: 16,
    motionPreset: "faster",
    newName: "Vertical Master",
    useNestedSequences: true
}));
assert.strictEqual(reframed.success, true, JSON.stringify(reframed));
assert.deepStrictEqual(capturedAutoReframeArgs, [9, 16, "faster", "Vertical Master", true]);
assert.strictEqual(reframed.data.sourceSequenceID, "source-1");
assert.strictEqual(reframed.data.sequenceID, "derivative-1");
assert.strictEqual(reframed.data.frameSizeHorizontal, 1080);
assert.strictEqual(reframed.data.frameSizeVertical, 1920);
assert.strictEqual(reframed.data.verified, true);
assert.strictEqual(reframed.data.activeSequenceID, "derivative-1");
assert.strictEqual(context.app.project.activeSequence.sequenceID, "derivative-1");

sequenceCollection = makeSequenceCollection([sourceSequence]);
context.app.project.sequences = sequenceCollection;
context.app.project.activeSequence = sourceSequence;
expectFailure(call("autoReframeSequence", JSON.stringify({ sourceSequenceID: "wrong-source", numerator: 9, denominator: 16 })), /Active sequence mismatch/);
expectFailure(call("autoReframeSequence", JSON.stringify({ sourceSequenceID: "source-1", numerator: 9, denominator: 16, motionPreset: "fast" })), /slower, default, or faster/);
expectFailure(call("autoReframeSequence", JSON.stringify({ sourceSequenceID: "source-1", numerator: 9.5, denominator: 16 })), /positive integers/);

sequenceCollection = makeSequenceCollection([sourceSequence]);
var activationLockedProject = { sequences: sequenceCollection, openSequence: function () { return true; } };
Object.defineProperty(activationLockedProject, "activeSequence", {
    get: function () { return sourceSequence; },
    set: function () {}
});
context.app = { project: activationLockedProject };
expectFailure(call("autoReframeSequence", JSON.stringify({
    sourceSequenceID: "source-1",
    numerator: 9,
    denominator: 16,
    newName: "Cannot Activate"
})), /could not verify derivative activation/);

var cloneSettings = { videoFrameWidth: 1920, videoFrameHeight: 1080 };
var verticalClone = {
    name: "Master",
    sequenceID: "vertical-1",
    videoTracks: { numTracks: 1 },
    audioTracks: { numTracks: 1 },
    getSettings: function () { return cloneSettings; },
    setSettings: function (settings) { cloneSettings = settings; return 0; }
};
sourceSequence.getSettings = function () { return { videoFrameWidth: 1920, videoFrameHeight: 1080 }; };
sourceSequence.clone = function () {
    sequenceCollection.push(verticalClone);
    sequenceCollection.numSequences = sequenceCollection.length;
};
sequenceCollection = makeSequenceCollection([sourceSequence]);
context.app = { project: makeProject(sequenceCollection, sourceSequence) };
var vertical = call("createVerticalVersion", 0, "Vertical Master");
assert.strictEqual(vertical.success, true, JSON.stringify(vertical));
assert.strictEqual(vertical.data.newSequenceID, "vertical-1");
assert.strictEqual(vertical.data.activeSequenceID, "vertical-1");
assert.strictEqual(context.app.project.activeSequence.sequenceID, "vertical-1");

function makeParameter(displayName, initialValue, status, readbackOverride) {
    var value = initialValue;
    return {
        displayName: displayName,
        setValue: function (nextValue) {
            value = nextValue;
            return status === undefined ? 0 : status;
        },
        getValue: function () {
            return readbackOverride === undefined ? value : readbackOverride;
        }
    };
}

function setMotionProject(positionParam, scaleParam) {
    var properties = [positionParam, scaleParam];
    properties.numItems = properties.length;
    var components = [{ properties: properties }];
    components.numItems = components.length;
    var clips = [{ components: components }];
    clips.numItems = clips.length;
    var tracks = [{ clips: clips }];
    tracks.numTracks = tracks.length;
    context.app = {
        project: {
            activeSequence: { videoTracks: tracks }
        }
    };
}

setMotionProject(
    makeParameter("Position", [0.5, 0.5]),
    makeParameter("Scale", 100)
);
var positioned = call("setPosition", 0, 0, 0.25, 0.75);
assert.strictEqual(positioned.success, true, JSON.stringify(positioned));
assert.strictEqual(positioned.data.verified, true);
assert.deepStrictEqual(positioned.data.x, 0.25);
assert.deepStrictEqual(positioned.data.y, 0.75);
var scaled = call("setScale", 0, 0, 125);
assert.strictEqual(scaled.success, true, JSON.stringify(scaled));
assert.strictEqual(scaled.data.scale, 125);
assert.strictEqual(scaled.data.verified, true);

expectFailure(call("setPosition", 0.5, 0, 0.5, 0.5), /non-negative integer/);
expectFailure(call("setPosition", 0, 0, 11, 0.5), /normalized coordinate/);
expectFailure(call("setScale", 0, 0, 0), /greater than 0/);

setMotionProject(
    makeParameter("Position", [0.5, 0.5], 1),
    makeParameter("Scale", 100)
);
expectFailure(call("setPosition", 0, 0, 0.25, 0.75), /failure status 1/);

setMotionProject(
    makeParameter("Position", [0.5, 0.5]),
    makeParameter("Scale", 100, 0, 99)
);
expectFailure(call("setScale", 0, 0, 125), /read back 99 instead of 125/);

console.log("Verified Auto Reframe and motion workflow smoke tests passed.");
