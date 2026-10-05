"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");
function fixture(options = {}) {
    const item = { getMediaPath: () => "/fixtures/source.mp4", isOffline: () => !!options.offline,
        isSequence: () => !!options.nested, duration: { seconds: 20 } };
    const clip = { projectItem: item, disabled: !!options.disabled,
        start: { seconds: 10 }, end: { seconds: 14 }, inPoint: { seconds: 3 }, outPoint: { seconds: 7 },
        getSpeed: () => options.speed === undefined ? 1 : options.speed, isSpeedReversed: () => options.reverseValue === undefined ? !!options.reversed : options.reverseValue };
    const clips = [clip]; clips.numItems = 1;
    const transitions = options.transition ? [{ start: { seconds: 11 }, end: { seconds: 12 } }] : [];
    transitions.numItems = transitions.length;
    const track = { clips, transitions, isMuted: () => !!options.muted };
    const tracks = options.overlap ? [track, track] : [track]; tracks.numTracks = tracks.length;
    const sequence = { sequenceID: "fixture", videoTracks: tracks, getPlayerPosition: () => ({ seconds: 11.5 }) };
    if (options.unreadable) delete clips.numItems;
    if (options.missingBounds) delete clip.inPoint;
    if (options.missingReverse) delete clip.isSpeedReversed;
    const context = vm.createContext({ app: { project: { activeSequence: sequence } } });
    vm.runInContext(fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8"), context);
    return { run: seconds => typeof context.getFrameSourceAtTime === "function"
        ? JSON.parse(context.getFrameSourceAtTime(seconds)) : { success: false, error: "Host source-frame command unavailable" } };
}
test("source frame maps a trimmed clip using source in plus timeline offset", () => {
    assert.deepEqual(fixture().run(), { success: true, data: {
        mediaPath: "/fixtures/source.mp4", seconds: 4.5, timelineSeconds: 11.5
    } });
});
test("source frame supports explicit time without moving the playhead", () => {
    assert.equal(fixture().run(10).data.seconds, 3);
});
for (const [name, options, seconds, error] of [
    ["gap", {}, 9, /No eligible/], ["out point", {}, 14, /No eligible/],
    ["overlap", { overlap: true }, 11.5, /exactly one/],
    ["transition", { transition: true }, 11.5, /transition/],
    ["retimed", { speed: 2 }, 11.5, /speed/], ["reversed", { reversed: true }, 11.5, /revers/],
    ["unknown reverse state", { missingReverse: true }, 11.5, /reverse/],
    ["boolean speed", { speed: true }, 11.5, /speed/],
    ["offline", { offline: true }, 11.5, /offline/], ["nested", { nested: true }, 11.5, /nested/],
    ["muted", { muted: true }, 11.5, /No eligible/], ["disabled", { disabled: true }, 11.5, /No eligible/],
    ["unreadable collection", { unreadable: true }, 11.5, /count/],
    ["missing source trim", { missingBounds: true }, 11.5, /bounds/],
    ["invalid time", {}, -1, /seconds/], ["string time", {}, "11.5", /seconds/],
]) test("source frame refuses " + name, () => {
    const result = fixture(options).run(seconds);
    assert.equal(result.success, false);
    assert.match(result.error, error);
});

test("source frame accepts native numeric forward state", () => { assert.equal(fixture({reverseValue:0}).run().success,true); });
for (const value of [1,2,"0",null]) test("source frame refuses reverse value " + JSON.stringify(value), () => { assert.equal(fixture({reverseValue:value}).run().success,false); });
