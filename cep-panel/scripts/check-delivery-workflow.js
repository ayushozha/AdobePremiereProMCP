#!/usr/bin/env node

"use strict";

var assert = require("assert");
var fs = require("fs");
var path = require("path");
var vm = require("vm");

var hostPath = path.resolve(__dirname, "..", "src", "host", "premiere.jsx");
var hostSource = fs.readFileSync(hostPath, "utf8");
var files = {};

function MockFile(value) {
    var canonicalPath = path.resolve(String(value));
    this.fsName = canonicalPath;
    Object.defineProperty(this, "exists", {
        get: function () { return files[canonicalPath] !== undefined; }
    });
    Object.defineProperty(this, "length", {
        get: function () { return files[canonicalPath] || 0; }
    });
}

function markFile(filePath, size) {
    files[path.resolve(filePath)] = size;
}

var context = vm.createContext({
    console: console,
    JSON: JSON,
    Date: Date,
    Math: Math,
    File: MockFile,
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

var captions = [{
    name: "Hello",
    start: { seconds: 0 },
    end: { seconds: 1 }
}];
captions.numItems = captions.length;
var captionTracks = [{ clips: captions }];
captionTracks.numTracks = captionTracks.length;
var sequence = {
    name: "Delivery Master",
    sequenceID: "delivery-sequence-1",
    captionTracks: captionTracks,
    exportCaptions: function (outputPath) {
        markFile(outputPath, 128);
        return true;
    }
};
var sequences = [sequence];
sequences.numSequences = sequences.length;

markFile("/tmp/delivery-preset.epr", 512);

var encodeCalls = 0;
var startCalls = 0;
context.app = {
    project: { activeSequence: sequence, sequences: sequences },
    encoder: {
        launchEncoder: function () { return 0; },
        encodeSequence: function (seq, outputPath, presetPath, workAreaType, removeOnDone) {
            encodeCalls++;
            assert.strictEqual(seq, sequence);
            assert.strictEqual(outputPath, "/tmp/delivery.mp4");
            assert.strictEqual(presetPath, "/tmp/delivery-preset.epr");
            assert.strictEqual(workAreaType, 0);
            assert.strictEqual(removeOnDone, 0);
            return "delivery-job-1";
        },
        startBatch: function () { startCalls++; return 0; }
    }
};

var queued = call("exportViaAME", -1, "/tmp/delivery.mp4", "/tmp/delivery-preset.epr", 0, false);
assert.strictEqual(queued.success, true, JSON.stringify(queued));
assert.strictEqual(queued.data.jobID, "delivery-job-1");
assert.strictEqual(queued.data.status, "queued_in_ame");
assert.strictEqual(queued.data.sequenceID, "delivery-sequence-1");
assert.strictEqual(encodeCalls, 1);
assert.strictEqual(startCalls, 1);

context.app.encoder = {
    launchEncoder: function () { return 1; },
    encodeSequence: function () { throw new Error("must not encode after launch failure"); },
    startBatch: function () { throw new Error("must not start after launch failure"); }
};
expectFailure(call("exportViaAME", -1, "/tmp/delivery.mp4", "/tmp/delivery-preset.epr", 0, false), /launchEncoder failed with status 1/);

context.app.encoder = {
    launchEncoder: function () { return 0; },
    encodeSequence: function () { return "delivery-job-2"; },
    startBatch: function () { return 1; }
};
expectFailure(call("exportViaAME", -1, "/tmp/delivery.mp4", "/tmp/delivery-preset.epr", 0, false), /job delivery-job-2 was queued.*status 1/);

expectFailure(call("exportCaptions", "/tmp/delivery.srt", "SRT", "wrong-sequence"), /Active sequence mismatch/);
expectFailure(call("exportCaptions", "/tmp/delivery.srt", "SRT", ""), /expectedSequenceId is required/);
var sidecar = call("exportCaptions", "/tmp/delivery.srt", "SRT", "delivery-sequence-1");
assert.strictEqual(sidecar.success, true, JSON.stringify(sidecar));
assert.strictEqual(sidecar.data.sequenceID, "delivery-sequence-1");
assert.strictEqual(sidecar.data.captionCount, 1);
assert.strictEqual(sidecar.data.verified, true);

console.log("Verified AME queue and caption-sidecar workflow smoke tests passed.");
