"use strict";

var assert = require("node:assert/strict");
var fs = require("node:fs");
var path = require("node:path");
var test = require("node:test");
var vm = require("node:vm");
var acorn = require("acorn");

var hostDirectory = path.join(__dirname, "..", "src", "host");
var hostSource = fs.readFileSync(path.join(hostDirectory, "premiere.jsx"), "utf8");

["core.jsx", "premiere.jsx"].forEach(function (filename) {
    test(filename + " uses ExtendScript-compatible syntax and quoted reserved properties", function () {
        acorn.parse(fs.readFileSync(path.join(hostDirectory, filename), "utf8"), {
            ecmaVersion: 3,
            allowReserved: "never",
            locations: true,
        });
    });
});

test("the compatibility gate rejects unquoted reserved property names", function () {
    assert.throws(function () {
        acorn.parse("var value = { package: {} };", { ecmaVersion: 3, allowReserved: "never" });
    }, SyntaxError);
});

test("health checks work when the host has no Date.toISOString", function () {
    var context = vm.createContext({ app: { version: "26.3.2", project: null } });
    vm.runInContext("Date.prototype.toISOString = undefined;", context);
    vm.runInContext(hostSource, context);
    var result = JSON.parse(context.mcpDispatch("ping", "{}"));
    assert.equal(result.success, true);
    assert.equal(result.data.status, "ok");
    assert.equal(result.data.version, "26.3.2");
    assert.equal(result.data.projectOpen, false);
    assert.match(result.data.timestamp, /^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{3}Z$/);
    assert.equal(vm.runInContext("new Date(0).toISOString()", context), "1970-01-01T00:00:00.000Z");
    assert.equal(vm.runInContext("new Date(-1).toISOString()", context), "1969-12-31T23:59:59.999Z");
    assert.equal(vm.runInContext("new Date(253402300800000).toISOString()", context), "+010000-01-01T00:00:00.000Z");
    assert.throws(function () { vm.runInContext("new Date(NaN).toISOString()", context); }, /Invalid time value/);
});

test("the host keeps a working native Date.toISOString", function () {
    var context = vm.createContext({});
    vm.runInContext("var originalISOString = Date.prototype.toISOString;", context);
    vm.runInContext(hostSource, context);
    assert.equal(vm.runInContext("Date.prototype.toISOString === originalISOString", context), true);
});

test("a real host timeline uses tick duration and second-valued in/out points", function () {
    var context = vm.createContext({});
    vm.runInContext("JSON = undefined; Date.prototype.toISOString = undefined;", context);
    vm.runInContext(fs.readFileSync(path.join(hostDirectory, "core.jsx"), "utf8"), context);
    vm.runInContext(hostSource, context);
    vm.runInContext([
        "var tracks = []; tracks.numTracks = 0;",
        "var fixtureSequence = {sequenceID:'fixture',name:'Fixture',end:'254016000000',zeroPoint:'0',",
        "  getInPoint:function(){return '0.25';},getOutPoint:function(){return '0.75';},",
        "  videoTracks:tracks,audioTracks:tracks,timebase:'10584000000'};",
        "var sequences=[fixtureSequence];sequences.numSequences=1;",
        "app={project:{name:'Fixture',sequences:sequences,activeSequence:fixtureSequence}};",
    ].join("\n"), context);
    var timeline = JSON.parse(context.mcpDispatch("mcpGetTimelineState", "{}"));
    assert.equal(timeline.success, true);
    assert.equal(timeline.data.totalDurationSeconds, 1);
    assert.deepEqual(timeline.data.videoTracks, []);
    var project = JSON.parse(context.mcpDispatch("getProjectState", "{}"));
    assert.equal(project.success, true);
    assert.equal(project.data.sequences[0].inPoint, 0.25);
    assert.equal(project.data.sequences[0].outPoint, 0.75);
    assert.equal(project.data.sequences[0].durationSeconds, 1);
    var sessionStats = JSON.parse(context.mcpDispatch("getEditingSessionStats", "{}"));
    assert.equal(sessionStats.success, true);
    assert.equal(sessionStats.data.activeSequence.duration, 1);
    assert.equal(context._timeToSeconds({seconds:1.25}), 1.25);
    assert.equal(context._timeToSeconds({ticks:"254016000000"}), 1);
    assert.equal(context._ticksToSeconds("-254016000000"), -1);
    assert.throws(function () { context._timeToSeconds({}); }, /Invalid Premiere time value/);
});

["core.jsx", "premiere.jsx"].forEach(function (filename) {
    test(filename + " serializes nonfinite numbers as valid JSON nulls", function () {
        var context = vm.createContext({});
        vm.runInContext("JSON = undefined;", context);
        vm.runInContext(fs.readFileSync(path.join(hostDirectory, filename), "utf8"), context);
        var serialized = vm.runInContext("JSON.stringify({nan:NaN,pos:Infinity,neg:-Infinity,value:1.25})", context);
        assert.deepEqual(JSON.parse(serialized), {nan:null,pos:null,neg:null,value:1.25});
    });
});
