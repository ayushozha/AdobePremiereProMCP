"use strict";
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const zlib = require("node:zlib");
const { test } = require("node:test");
const source = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");

// Independent bit-at-a-time CRC for building real PNG fixtures; the host uses
// a lookup table. Image bytes are deflated and decoded in the success test.
function chunk(type, data = Buffer.alloc(0)) {
    const name = Buffer.from(type, "ascii"), checked = Buffer.concat([name, data]);
    let crc = 0xffffffff;
    for (const byte of checked) {
        crc ^= byte;
        for (let bit = 0; bit < 8; bit++) crc = crc & 1 ? (crc >>> 1) ^ 0xedb88320 : crc >>> 1;
    }
    const result = Buffer.alloc(data.length + 12);
    result.writeUInt32BE(data.length, 0); name.copy(result, 4); data.copy(result, 8);
    result.writeUInt32BE((crc ^ 0xffffffff) >>> 0, result.length - 4);
    return result;
}

function pngFixture(options = {}) {
    const width = options.wrongSize ? 640 : 320, height = 180;
    const signature = Buffer.from("89504e470d0a1a0a", "hex");
    const header = Buffer.alloc(13);
    header.writeUInt32BE(width, 0); header.writeUInt32BE(height, 4);
    header[8] = 8; header[9] = options.missingPalette ? 3 : 6;
    if (options.invalidHeader) header[12] = 2;
    const raw = Buffer.alloc((width * 4 + 1) * height);
    for (let row = 0; row < height; row++) {
        for (let column = 0; column < width; column++) {
            const offset = row * (width * 4 + 1) + 1 + column * 4;
            raw[offset] = 64; raw[offset + 1] = 128; raw[offset + 2] = 192; raw[offset + 3] = 255;
        }
    }
    const ihdr = chunk("IHDR", header), idat = chunk("IDAT", zlib.deflateSync(raw)), iend = chunk("IEND");
    let parts = [signature, ihdr];
    if (options.duplicateHeader) parts.push(ihdr);
    if (options.unknownCritical) parts.push(chunk("ABCD"));
    if (!options.missingData) parts.push(idat);
    if (options.nonconsecutiveData) parts.push(chunk("tEXt", Buffer.from("key\0value")), chunk("IDAT"));
    if (!options.missingEnd) parts.push(options.invalidEnd ? chunk("IEND", Buffer.from([0])) : iend);
    let png = Buffer.concat(parts);
    if (options.headerOnly) png = png.subarray(0, 24);
    if (options.truncatedChunk) png = png.subarray(0, 50);
    if (options.corruptChecksum) png[32] ^= 1;
    if (options.trailingBytes) png = Buffer.concat([png, Buffer.from("extra")]);
    return { png, raw };
}

function fixture(options = {}) {
    const calls = [], files = {}, removed = [];
    const { png, raw } = pngFixture(options);
    function File(filename) {
        let isOpen = false;
        this.fsName = filename;
        if (options.collision) files[filename] = "existing user file";
        Object.defineProperty(this, "exists", { get() { return Object.prototype.hasOwnProperty.call(files, filename); } });
        Object.defineProperty(this, "length", { get() { return files[filename]?.length || 0; } });
        this.open = () => { isOpen = !options.openFailure; return isOpen; };
        this.close = () => { isOpen = false; };
        this.read = () => { if (options.readFailure) throw new Error("read failed"); return files[filename]; };
        this.remove = () => { assert.equal(isOpen, false, "owned output must be closed before removal"); removed.push(filename); delete files[filename]; };
    }
    const position = { ticks: "254016000000", seconds: 1, getFormatted() { throw new Error("Display formatting must not replace QE CTI timecode"); } };
    const timecode = Object.prototype.hasOwnProperty.call(options, "timecode") ? options.timecode : "00:00:01:00";
    const seq = { getPlayerPosition: () => position, videoDisplayFormat: options.framesDisplay ? 109 : 100, getSettings: () => ({ videoFrameRate: { ticks: "10584000000" }, videoFrameWidth: 320, videoFrameHeight: 180 }),
        exportFramePNG() { throw new Error("Regular DOM/ticks export must not be used"); } };
    const qeSeq = { CTI: options.noCTI ? null : { timecode }, exportFramePNG(requestedTimecode, stem) {
        calls.push({ timecode: requestedTimecode, stem });
        assert.equal(requestedTimecode, timecode); assert.ok(!stem.endsWith(".png"));
        if (options.exportThrows) { files[stem + ".png"] = "partial"; throw new Error("export failed"); }
        if (!options.noOutput) files[stem + ".png"] = options.invalidPNG ? "not an image" : png.toString("latin1");
    } };
    const context = vm.createContext({ File, Folder: { temp: { fsName: "/temp" } }, app: { project: { activeSequence: seq }, enableQE() {} }, qe: { project: { getActiveSequence: () => options.noQE ? null : qeSeq } } });
    vm.runInContext(source, context);
    return { calls, files, removed, raw, run: () => JSON.parse(context.captureFrameAsBase64("{}")) };
}

test("captures the formatted playhead through QE with a suffix-free output stem", () => {
    const f = fixture(), r = f.run();
    assert.equal(r.success, true); assert.equal(r.data.width, 320); assert.equal(r.data.height, 180); assert.equal(r.data.timecode, 1);
    const png = Buffer.from(r.data.image_base64, "base64"), imageData = [];
    assert.equal(png.subarray(0, 8).toString("hex"), "89504e470d0a1a0a");
    for (let offset = 8; offset < png.length;) {
        const length = png.readUInt32BE(offset), type = png.toString("ascii", offset + 4, offset + 8);
        if (type === "IDAT") imageData.push(png.subarray(offset + 8, offset + 8 + length));
        offset += length + 12;
    }
    assert.deepEqual(zlib.inflateSync(Buffer.concat(imageData)), f.raw, "returned PNG contains a real decodable image");
    assert.equal(f.calls.length, 1); assert.equal(f.removed.length, 1); assert.deepEqual(f.files, {});
});

for (const option of ["noOutput", "invalidPNG", "wrongSize", "exportThrows", "openFailure", "readFailure", "noQE", "noCTI", "headerOnly", "truncatedChunk", "corruptChecksum", "missingData", "missingEnd", "duplicateHeader", "invalidHeader", "invalidEnd", "trailingBytes", "unknownCritical", "nonconsecutiveData", "missingPalette"]) {
    test(option + " cannot return a frame and cleans its owned temporary output", () => {
        const f = fixture({ [option]: true }), r = f.run();
        assert.equal(r.success, false); assert.equal(r.data, undefined); assert.deepEqual(f.files, {});
    });
}

for (const timecode of ["00:00:01;00", "00;00;01;00"]) {
    test("drop-frame QE timecode " + timecode + " is passed without display conversion", () => {
        const f = fixture({ timecode }), r = f.run();
        assert.equal(r.success, true); assert.equal(f.calls[0].timecode, timecode);
    });
}

test("a sequence showing frame counts still captures using QE CTI timecode", () => {
    const f = fixture({ framesDisplay: true }), r = f.run();
    assert.equal(r.success, true); assert.equal(f.calls[0].timecode, "00:00:01:00");
});

for (const timecode of ["24", "000+00", "", undefined]) {
    test("invalid QE timecode " + String(timecode) + " fails before export", () => {
        const f = fixture({ timecode }), r = f.run();
        assert.equal(r.success, false); assert.equal(f.calls.length, 0); assert.deepEqual(f.files, {});
    });
}

test("a colliding existing file is neither exported over nor removed", () => {
    const f = fixture({ collision: true }), r = f.run();
    assert.equal(r.success, false); assert.equal(f.calls.length, 0); assert.equal(f.removed.length, 0);
    assert.equal(Object.values(f.files)[0], "existing user file");
});
