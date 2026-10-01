"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const sources = {};
for (const filename of ["core.jsx", "premiere.jsx"]) {
    sources[filename] = fs.readFileSync(path.join(__dirname, "../src/host", filename), "utf8");
}

function host(filename, setup = "JSON = undefined;") {
    const context = vm.createContext({});
    vm.runInContext(setup, context);
    vm.runInContext(sources[filename], context);
    return context;
}

function controlPayload() {
    const data = { hasOwnProperty: "a parameter may shadow this method", unicode: "café 🎬" };
    for (let code = 0; code < 32; code++) {
        const control = String.fromCharCode(code);
        data['key' + control + '"\\'] = 'value' + control + '"\\';
    }
    data.separators = "first\u2028second\u2029third";
    return data;
}

for (const filename of Object.keys(sources)) {
    test(filename + " round-trips all 32 control characters in values and keys", () => {
        const context = host(filename);
        const payload = controlPayload();
        const encoded = context._ok(payload);
        assert.doesNotMatch(encoded, /[\x00-\x1f\u2028\u2029]/);
        assert.deepEqual(JSON.parse(encoded), { success: true, data: payload });
        context.encoded = encoded;
        context.roundTrip = vm.runInContext("JSON.parse(encoded)", context);
        assert.deepEqual(JSON.parse(context._ok(context.roundTrip.data)), { success: true, data: payload });
    });

    test(filename + " encodes native Lumetri control-character values as valid JSON", () => {
        const context = host(filename);
        const payload = { displayName: "Hue vs Sat", value: String.fromCharCode(3) };
        const encoded = context._ok(payload);
        assert.match(encoded, /"value":"\\u0003"/);
        assert.deepEqual(JSON.parse(encoded).data, payload);
    });

    test(filename + " preserves control characters in error messages", () => {
        const context = host(filename);
        const message = 'Host failed\b\f\n\r\t\u0000\u0003"\\';
        assert.deepEqual(JSON.parse(context._err(message)), { success: false, error: message });
    });

    test(filename + " replaces an already-loaded broken legacy serializer", () => {
        const context = host(filename, [
            "JSON = { stringify: function () { return '{\"value\":\"' + String.fromCharCode(3) + '\"}'; },",
            "parse: function (text) { return eval('(' + text + ')'); } };",
            "var legacyStringify = JSON.stringify;",
        ].join("\n"));
        assert.equal(vm.runInContext("JSON.stringify === legacyStringify", context), false);
        assert.deepEqual(JSON.parse(context._ok(controlPayload())).data, controlPayload());
    });

    test(filename + " keeps a working native JSON implementation", () => {
        const context = host(filename, "var nativeStringify = JSON.stringify; var nativeParse = JSON.parse;");
        assert.equal(vm.runInContext("JSON.stringify === nativeStringify && JSON.parse === nativeParse", context), true);
        assert.deepEqual(JSON.parse(context._ok(controlPayload())).data, controlPayload());
    });

    test(filename + " omits undefined object fields and emits null array entries", () => {
        const context = host(filename);
        const encoded = vm.runInContext("JSON.stringify({empty:undefined,callable:function(){},values:[undefined,function(){},,NaN,Infinity,1],present:false})", context);
        assert.deepEqual(JSON.parse(encoded), { values: [null, null, null, null, null, 1], present: false });
        assert.equal(vm.runInContext("JSON.stringify(undefined)", context), undefined);
        assert.equal(vm.runInContext("JSON.stringify(function(){})", context), undefined);
    });
}

test("core then full host loading retains valid escaping", () => {
    const context = host("core.jsx");
    vm.runInContext(sources["premiere.jsx"], context);
    assert.deepEqual(JSON.parse(context._ok(controlPayload())).data, controlPayload());
});

test("full-host pretty JSON also quotes object keys and values correctly", () => {
    const context = host("premiere.jsx");
    context.payload = controlPayload();
    const encoded = vm.runInContext("JSON.stringify(payload, null, 2)", context);
    assert.ok(encoded.includes("\n  "));
    assert.deepEqual(JSON.parse(encoded), controlPayload());
});

test("the read-only quote helper supports authenticated missing-symbol reload calls", () => {
    const context = host("premiere.jsx");
    const value = 'quote"\\\u0003\n';
    const encoded = context.mcpDispatch("_mcpQuoteJSONString", JSON.stringify({ value }));
    assert.deepEqual(JSON.parse(encoded), value);
});
