"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const source = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");
const ticksPerSecond = 254016000000;
const request = { recipeId: "soft-dissolve", transitionName: "Cross Dissolve", durationFrames: 12, trackIndex: 0, clipIndex: 0 };

function collection(items, countKey) {
    Object.defineProperty(items, countKey, { get() { return items.length; } });
    return items;
}

// Premiere itself is external to Node. This fixture models its mutation and
// readback boundaries; all recipe validation and verification runs real JSX.
function fixture(options = {}) {
    const fps = options.fps || 24;
    const calls = [];
    const projectItems = [0, 1].map(index => ({
        nodeId: "media-" + index, name: "Media " + index,
        getMediaDuration() { return options.unknownMediaDuration ? undefined : { seconds: options.mediaDuration === undefined ? 20 : options.mediaDuration }; },
        getInPoint() { return { seconds: 0 }; },
        getOutPoint() { return { seconds: options.sourceMarkEnd === undefined ? 20 : options.sourceMarkEnd }; },
        setInPoint() { assert.fail("recipe preflight must not change source marks"); },
        setOutPoint() { assert.fail("recipe preflight must not change source marks"); },
        isOffline() { return !!options.offline; },
        isSequence() { return !!options.nestedSequence; },
    }));
    function makeSequence(id) {
        const clips = collection([0, 1].map(index => ({
            nodeId: id + "-clip-" + index, name: "Clip " + index, projectItem: projectItems[index],
            start: { seconds: index === 0 ? 2 : 6 + (options.gap || 0) },
            end: { seconds: index === 0 ? 6 : 10 + (options.gap || 0) },
            inPoint: { seconds: index === 0 ? (options.outgoingOut === undefined ? 2 : options.outgoingOut - 4) : (options.incomingIn === undefined ? 3 : options.incomingIn) },
            outPoint: { seconds: index === 0 ? (options.outgoingOut === undefined ? 6 : options.outgoingOut) : (options.incomingIn === undefined ? 7 : options.incomingIn + 4) },
            getSpeed() { if (options.speedThrows) throw new Error("speed unreadable"); return options.speed === undefined ? 1 : options.speed; },
            isSpeedReversed() { return !!options.reversed; },
            components: collection([], "numItems"),
        })), "numItems");
        if (options.unknownSpeed) delete clips[0].getSpeed;
        const transitions = collection(options.existingCut || options.existingElsewhere ? [{ nodeId: id + "-existing", name: "Dip to Black",
            start: { seconds: options.existingElsewhere ? 2 : 5.75 }, end: { seconds: options.existingElsewhere ? 2.5 : 6.25 } }] : [], "numItems");
        const videoTrack = { clips, transitions, isLocked() { return !!options.locked; } };
        if (options.noTransitionReadback) delete videoTrack.transitions;
        return { sequenceID: id, name: id === "source" ? "Source edit" : "Source edit Copy",
            timebase: options.unknownRate ? "" : String(ticksPerSecond / fps),
            getSettings() { return options.unknownRate ? {} : { videoFrameRate: { ticks: String(ticksPerSecond / fps) } }; },
            videoTracks: collection([videoTrack], "numTracks"), audioTracks: collection([], "numTracks") };
    }
    const original = makeSequence("source");
    const sequences = collection([original], "numSequences");
    const project = { sequences, activeSequence: original,
        openSequence(id) {
            calls.push({ operation: "activate", id });
            if (!options.ignoreActivation) project.activeSequence = sequences.find(sequence => sequence.sequenceID === id);
            return !options.activationRejected;
        } };
    original.clone = function () {
        calls.push({ operation: "clone" });
        if (options.cloneNoop) return true;
        const copy = makeSequence(options.sameCopyID ? "source" : "copy");
        if (options.incompleteCopy) copy.videoTracks[0].clips.pop();
        if (options.copyTimingChanged) copy.videoTracks[0].clips[1].start.seconds = 7;
        if (options.aliasCopy) copy.videoTracks = original.videoTracks;
        sequences.unshift(copy); // New sequences are not necessarily appended.
        return true;
    };
    function qeSequence(sequence) {
        const track = sequence.videoTracks[0];
        const items = [{ type: "Empty", start: { secs: 0 }, end: { secs: 2 } }];
        track.clips.forEach((clip, domIndex) => {
            const item = { type: "Clip", name: clip.name,
                start: { secs: clip.start.seconds }, end: { secs: clip.end.seconds },
                addTransition(native, addToStart, duration, offset, alignment, singleSided) {
                    calls.push({ operation: "add", sequenceID: sequence.sequenceID, domIndex, addToStart, duration, offset, alignment, singleSided });
                    if (options.nativeThrows) throw new Error("native add failed");
                    if (options.nativeRejected) return false;
                    if (options.nativeNoop) return true;
                    // QE SS.FF stores integer frames in the fractional field.
                    const [seconds, frames] = duration.split(".").map(Number);
                    const appliedFrames = options.actualFrames === undefined ? seconds * Math.round(fps) + frames : options.actualFrames;
                    const cut = (addToStart ? clip.start.seconds : clip.end.seconds) + (options.cutOffset || 0);
                    const durationSeconds = appliedFrames / fps;
                    const transition = { nodeId: sequence.sequenceID + "-transition-" + track.transitions.length,
                        name: options.actualName || native.name, displayName: options.actualName || native.name,
                        start: { seconds: cut - durationSeconds / 2 }, end: { seconds: cut + durationSeconds / 2 },
                        duration: { seconds: durationSeconds } };
                    track.transitions.push(transition);
                    if (options.extraTransition) track.transitions.push({ ...transition, nodeId: "extra" });
                    if (options.replaceExisting) {
                        track.transitions.shift();
                        track.transitions.push({ ...transition, nodeId: "extra" });
                    }
                    if (options.changeExistingTransition) {
                        track.transitions[0].name = "Cross Dissolve";
                        track.transitions[0].end.seconds = 3;
                    }
                    if (options.changeCopyClipTiming) track.clips[1].end.seconds += 1;
                    if (options.changeSource) original.videoTracks[0].clips[0].end.seconds = 5;
                    return true;
                } };
            if (options.qeTimingMismatch) item.start.secs += 1;
            if (options.noQEAdd) delete item.addTransition;
            items.push(item);
            if (domIndex === 0) items.push({ type: "Transition", start: { secs: 2 }, end: { secs: 3 } });
            if (options.ambiguousQE) items.push({ ...item });
        });
        return { guid: options.qeWrongSequence ? "other" : sequence.sequenceID,
            getVideoTrackAt() { return { numItems: items.length, getItemAt(index) { return items[index]; } }; } };
    }
    const context = vm.createContext({ app: { project, enableQE() {} }, qe: { project: {
        getActiveSequence() { return qeSequence(project.activeSequence); },
        getVideoTransitionList() { return options.catalogUnavailable ? undefined : ["Cross Dissolve", "Dip to Black"]; },
        getVideoTransitionByName(name) { return options.transitionUnavailable ? null : { name: options.resolvedName || name }; },
    } } });
    vm.runInContext(source, context);
    return { calls, original, project, sequences, apply(args = {}) {
        return JSON.parse(context.mcpDispatch("applyTransitionRecipe", JSON.stringify({ ...request, ...args })));
    } };
}

test("omitted options produce a dry run with proven handles and no timeline mutation", () => {
    const host = fixture();
    const result = host.apply();
    assert.equal(result.success, true, result.error);
    assert.equal(result.data.dryRun, true);
    assert.equal(result.data.duplicateSequence, true);
    assert.equal(result.data.applied, false);
    assert.equal(result.data.status, "ready");
    assert.equal(result.data.sourceSequence.sequenceID, "source");
    assert.equal(result.data.appliedSequence, null);
    assert.equal(result.data.requested.cutSeconds, 6);
    assert.equal(result.data.requested.durationFrames, 12);
    assert.equal(result.data.diagnostics.frameRate, 24);
    assert.equal(result.data.diagnostics.qeClipIndex, 1);
    assert.equal(host.sequences.length, 1);
    assert.equal(host.original.videoTracks[0].transitions.length, 0);
    assert.deepEqual(host.calls, []);
});

for (const args of [
    { durationFrames: 0 }, { durationFrames: -1 }, { durationFrames: 1.5 }, { durationFrames: "12frames" },
    { durationFrames: null }, { trackIndex: -1 }, { trackIndex: 0.5 }, { clipIndex: -1 }, { clipIndex: null },
    { clipIndex: 1 }, { trackIndex: 1 }, { dryRun: "false" }, { duplicateSequence: "false" },
    { recipeId: "" }, { transitionName: "" },
]) {
    test("invalid recipe input fails without clone or transition: " + JSON.stringify(args), () => {
        const host = fixture();
        const result = host.apply(args);
        assert.equal(result.success, false);
        assert.match(result.error, /must|required|out of range|adjacent/i);
        assert.deepEqual(host.calls, []);
        assert.equal(host.original.videoTracks[0].transitions.length, 0);
    });
}

for (const [options, reason] of [
    [{ transitionUnavailable: true }, /transition.*unavailable|not found/i],
    [{ resolvedName: "Dip to Black" }, /exact|name/i],
    [{ catalogUnavailable: true }, /catalog/i],
    [{ unknownRate: true }, /frame rate/i],
    [{ gap: 0.2 }, /adjacent|cut/i],
    [{ existingCut: true }, /existing transition/i],
    [{ noTransitionReadback: true }, /transition.*readback/i],
    [{ incomingIn: 0.1 }, /handle/i],
    [{ outgoingOut: 19.8 }, /handle|source bounds/i],
    [{ unknownMediaDuration: true, sourceMarkEnd: 6 }, /handle/i],
    [{ speed: 2 }, /speed|playback/i],
    [{ speed: true }, /speed|playback/i],
    [{ unknownSpeed: true }, /speed|playback/i],
    [{ speedThrows: true }, /speed/i],
    [{ reversed: true }, /reversed|speed/i],
    [{ nestedSequence: true }, /nested|source/i],
    [{ offline: true }, /offline|source/i],
    [{ qeTimingMismatch: true }, /QE.*mapping|QE.*clip/i],
    [{ ambiguousQE: true }, /ambiguous|unique/i],
    [{ noQEAdd: true }, /QE.*transition/i],
    [{ qeWrongSequence: true }, /QE.*identity|QE.*sequence/i],
    [{ locked: true }, /locked/i],
]) {
    test("unsafe preflight fails before clone or mutation: " + Object.keys(options).join(","), () => {
        const host = fixture(options);
        const result = host.apply({ dryRun: false });
        assert.equal(result.success, false);
        assert.match(result.error, reason);
        assert.deepEqual(host.calls, []);
    });
}

test("typed source marks can conservatively prove handles when media duration is unavailable", () => {
    const host = fixture({ unknownMediaDuration: true, sourceMarkEnd: 8 });
    assert.equal(host.apply().success, true);
    assert.deepEqual(host.calls, []);
});

test("native apply targets the outgoing QE clip on a verified copy and preserves source", () => {
    const host = fixture();
    const result = host.apply({ dryRun: false });
    assert.equal(result.success, true, result.error);
    assert.equal(result.data.status, "applied");
    assert.equal(result.data.applied, true);
    assert.equal(result.data.verified, true);
    assert.equal(result.data.sourcePreserved, true);
    assert.equal(result.data.sourceSequence.sequenceID, "source");
    assert.equal(result.data.appliedSequence.sequenceID, "copy");
    assert.equal(result.data.actual.transitionName, "Cross Dissolve");
    assert.equal(result.data.actual.durationFrames, 12);
    assert.equal(result.data.actual.cutSeconds, 6);
    assert.equal(host.project.activeSequence.sequenceID, "copy");
    assert.equal(host.original.videoTracks[0].transitions.length, 0);
    assert.equal(host.sequences[0].videoTracks[0].transitions.length, 1);
    assert.deepEqual(host.calls.map(call => call.operation), ["clone", "activate", "add"]);
    assert.deepEqual(host.calls[2], { operation: "add", sequenceID: "copy", domIndex: 0,
        addToStart: false, duration: "0.12", offset: undefined, alignment: undefined, singleSided: undefined });
});

test("an explicit duplicateSequence=false applies to the current sequence", () => {
    const host = fixture();
    const result = host.apply({ dryRun: false, duplicateSequence: false });
    assert.equal(result.success, true, result.error);
    assert.equal(result.data.appliedSequence.sequenceID, "source");
    assert.equal(result.data.sourcePreserved, false);
    assert.deepEqual(host.calls.map(call => call.operation), ["add"]);
    assert.equal(host.original.videoTracks[0].transitions.length, 1);
});

for (const [options, reason] of [
    [{ cloneNoop: true }, /clone/i], [{ sameCopyID: true }, /identity|distinct/i],
    [{ incompleteCopy: true }, /copy|clone/i], [{ copyTimingChanged: true }, /copy|clone/i],
    [{ ignoreActivation: true }, /activate|activation/i], [{ activationRejected: true }, /open|activate|activation/i],
]) {
    test("unverified copy never receives a transition: " + Object.keys(options).join(","), () => {
        const host = fixture(options);
        const result = host.apply({ dryRun: false });
        assert.equal(result.success, false);
        assert.match(result.error, reason);
        assert.ok(!host.calls.some(call => call.operation === "add"));
        assert.equal(host.original.videoTracks[0].transitions.length, 0);
    });
}

for (const [options, reason] of [
    [{ nativeThrows: true }, /native add failed/], [{ nativeRejected: true }, /rejected/i],
    [{ nativeNoop: true }, /count|readback|verifiable/i],
    [{ actualName: "Dip to Black" }, /name/i], [{ actualFrames: 8 }, /duration/i],
    [{ cutOffset: 1 }, /cut|placement/i], [{ extraTransition: true }, /count/i],
    [{ changeSource: true }, /source.*changed|source.*preserv/i],
    [{ aliasCopy: true }, /source.*changed|source.*preserv|copy|clone/i],
]) {
    test("failed native apply reports copy identity and never claims success: " + Object.keys(options).join(","), () => {
        const host = fixture(options);
        const result = host.apply({ dryRun: false });
        assert.equal(result.success, false);
        assert.match(result.error, reason);
        assert.equal(result.data.sourceSequence.sequenceID, "source");
        assert.equal(result.data.appliedSequence.sequenceID, "copy");
        assert.equal(result.data.verified, false);
        if (!options.changeSource) assert.equal(host.original.videoTracks[0].transitions.length, 0);
    });
}

test("one-frame readback tolerance accepts rounding but rejects larger shortening", () => {
    const result = fixture({ actualFrames: 11 }).apply({ dryRun: false });
    assert.equal(result.success, true, result.error);
    assert.equal(result.data.actual.durationFrames, 11);
});

test("fractional sequence rates retain integer frame requests in QE duration", () => {
    const host = fixture({ fps: 30000 / 1001 });
    const result = host.apply({ dryRun: false, durationFrames: 45 });
    assert.equal(result.success, true, result.error);
    assert.equal(host.calls.find(call => call.operation === "add").duration, "1.15");
    assert.equal(result.data.actual.durationFrames, 45);
    assert.ok(Math.abs(result.data.actual.durationSeconds - 1.5015) < 1e-9);
});

test("transition readback must preserve every pre-existing transition", () => {
    const host = fixture({ existingElsewhere: true, replaceExisting: true });
    const result = host.apply({ dryRun: false });
    assert.equal(result.success, false);
    assert.match(result.error, /existing|preserv/i);
    assert.equal(host.original.videoTracks[0].transitions[0].name, "Dip to Black");
});

test("a preserved native transition ID cannot hide changed existing name and timing", () => {
    const host = fixture({ existingElsewhere: true, changeExistingTransition: true });
    const result = host.apply({ dryRun: false });
    assert.equal(result.success, false);
    assert.match(result.error, /timeline|existing|preserv/i);
    assert.equal(result.data.applied, true);
    assert.equal(result.data.verified, false);
    assert.equal(host.original.videoTracks[0].transitions[0].name, "Dip to Black");
    assert.equal(host.original.videoTracks[0].transitions[0].end.seconds, 2.5);
});

test("adding the requested transition cannot silently alter copied clip timing", () => {
    const host = fixture({ changeCopyClipTiming: true });
    const result = host.apply({ dryRun: false });
    assert.equal(result.success, false);
    assert.match(result.error, /timeline|preserv/i);
    assert.equal(result.data.applied, true);
    assert.equal(result.data.verified, false);
    assert.equal(host.original.videoTracks[0].clips[1].end.seconds, 10);
});

test("a count mismatch reports that native mutation was observed on the copy", () => {
    const host = fixture({ extraTransition: true });
    const result = host.apply({ dryRun: false });
    assert.equal(result.success, false);
    assert.equal(result.data.applied, true);
    assert.equal(result.data.mutationAttempted, true);
    assert.equal(result.data.diagnostics.transitionCountBefore, 0);
    assert.equal(result.data.diagnostics.transitionCountAfter, 2);
    assert.equal(result.data.appliedSequence.sequenceID, "copy");
});

for (const [options, applied] of [
    [{ nativeNoop: true }, false],
    [{ actualName: "Dip to Black" }, true],
]) {
    test("copy recovery diagnostics survive transport that retains only error text: applied=" + applied, () => {
        const host = fixture(options);
        const result = host.apply({ dryRun: false });
        assert.equal(result.success, false);
        assert.match(result.error, /sourceSequenceID=source/);
        assert.match(result.error, /appliedSequenceID=copy/);
        assert.match(result.error, /mutationAttempted=true/);
        assert.ok(result.error.includes("applied=" + applied));
        assert.match(result.error, /sourcePreserved=true/);
        assert.match(result.error, /verified=false/);
        assert.equal(result.data.sourceSequence.sequenceID, "source");
        assert.equal(result.data.appliedSequence.sequenceID, "copy");
        assert.equal(result.data.verified, false);
    });
}

test("a missing transition collection count fails before approving a dry run", () => {
    const host = fixture();
    host.original.videoTracks[0].transitions = [];
    const result = host.apply();
    assert.equal(result.success, false);
    assert.match(result.error, /transition.*readback|transition.*collection/i);
    assert.deepEqual(host.calls, []);
});

for (const [label, corrupt] of [
    ["audio track count", host => { host.original.audioTracks = []; }],
    ["target clip count", host => { host.original.videoTracks[0].clips = [...host.original.videoTracks[0].clips]; }],
    ["other track transition count", host => {
        host.original.videoTracks.push({ clips: collection([], "numItems"), transitions: [] });
    }],
    ["other track clip count", host => {
        host.original.audioTracks.push({ clips: [], transitions: collection([], "numItems") });
    }],
]) {
    test("unreadable " + label + " cannot silently disappear from source verification", () => {
        const host = fixture();
        corrupt(host);
        const result = host.apply({ dryRun: false });
        assert.equal(result.success, false);
        assert.deepEqual(host.calls, []);
        assert.match(result.error, /collection|count|readback/i);
    });
}
