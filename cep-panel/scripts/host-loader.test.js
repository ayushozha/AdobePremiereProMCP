"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");
const {
    resolveHostPaths,
    buildLoadScript,
    buildDispatchScript,
} = require("../src/host-loader.js");

// Run generated ExtendScript in one shared global context. The file fixtures
// model $.evalFile loading declarations into the host engine, rather than
// replacing the loader with a stub that always reports success.
function createHost(files) {
    const fixtures = new Map(Object.entries(files || {}));
    const loads = [];
    const context = vm.createContext({});

    function File(filePath) {
        if (!(this instanceof File)) return new File(filePath);
        this.fsName = String(filePath);
        this.fullName = this.fsName;
        this.exists = fixtures.has(this.fsName);
    }
    File.prototype.toString = function () { return this.fsName; };

    context.File = File;
    context.$ = {
        global: context,
        evalFile: function (file) {
            const filePath = String(file);
            loads.push(filePath);
            if (!fixtures.has(filePath)) {
                throw new Error("Fixture file does not exist: " + filePath);
            }
            return vm.runInContext(fixtures.get(filePath), context, {
                filename: filePath,
                timeout: 1000,
            });
        },
    };

    return {
        context,
        loads,
        run: function (script) {
            const result = vm.runInContext(script, context, { timeout: 1000 });
            assert.equal(typeof result, "string", "host result must be serialized JSON");
            return JSON.parse(result);
        },
    };
}

const dispatcherFixture = [
    "function echo(args) { return JSON.stringify({success:true,data:args}); }",
    "function mcpDispatch(name, argsJson) { return eval(name)(JSON.parse(argsJson)); }",
].join("\n");

test("host paths use the supplied extension root, independently of module location", function () {
    const root = path.resolve("/tmp/Premiere MCP Extension");
    assert.deepEqual(resolveHostPaths(root), {
        core: path.join(root, "src", "host", "core.jsx"),
        premiere: path.join(root, "src", "host", "premiere.jsx"),
    });
});

test("loading verifies the requested functions in the shared host global", function () {
    const hostPath = "/tmp/extension/src/host/premiere.jsx";
    const host = createHost({ [hostPath]: dispatcherFixture });
    assert.deepEqual(host.run(buildLoadScript(hostPath, ["mcpDispatch", "echo"])), {
        success: true,
    });
    assert.deepEqual(host.loads, [hostPath]);
    assert.equal(typeof host.context.mcpDispatch, "function");
    assert.equal(typeof host.context.echo, "function");
});

test("quoted paths, backslashes, spaces, and argument text survive generated scripts", function () {
    const hostPath = "C:\\Editors\\Editor's \"Premiere\" library\\src\\host\\premiere.jsx";
    const normalizedPath = "C:/Editors/Editor's \"Premiere\" library/src/host/premiere.jsx";
    const host = createHost({ [normalizedPath]: dispatcherFixture });
    assert.equal(host.run(buildLoadScript(hostPath, ["mcpDispatch", "echo"])).success, true);
    const args = { title: "Editor's \"cut\"\\draft\nsecond line", enabled: false };
    assert.deepEqual(host.run(buildDispatchScript(hostPath, "echo", JSON.stringify(args))), {
        success: true,
        data: args,
    });
    assert.deepEqual(host.loads, [normalizedPath]);
});

test("a missing host file produces an explicit failure instead of load success", function () {
    const hostPath = "/tmp/missing-extension/src/host/premiere.jsx";
    const host = createHost();
    const result = host.run(buildLoadScript(hostPath, ["mcpDispatch"]));
    assert.equal(result.success, false);
    assert.match(result.error, /missing|not found|does not exist/i);
    assert.ok(result.error.includes(hostPath), "diagnostic identifies the missing file");
    assert.equal(host.loads.length, 0, "missing files must be detected before evalFile");
});

test("an existing file without a required function is not treated as loaded", function () {
    const hostPath = "/tmp/extension/src/host/incomplete.jsx";
    const host = createHost({ [hostPath]: "var mcpDispatch = 42; function echo() {}" });
    const result = host.run(buildLoadScript(hostPath, ["mcpDispatch", "echo"]));
    assert.equal(result.success, false);
    assert.match(result.error, /mcpDispatch/);
    assert.deepEqual(host.loads, [hostPath]);
});

test("dispatch lazily loads once and reloads after the ExtendScript global is reset", function () {
    const hostPath = "/tmp/extension/src/host/premiere.jsx";
    const host = createHost({ [hostPath]: dispatcherFixture });
    const script = buildDispatchScript(hostPath, "echo", '{"take":3}');
    const expected = { success: true, data: { take: 3 } };

    assert.deepEqual(host.run(script), expected);
    assert.deepEqual(host.loads, [hostPath]);
    assert.deepEqual(host.run(script), expected);
    assert.deepEqual(host.loads, [hostPath], "available functions should not reload the file");

    // vm global declarations can remain internally bound after an external
    // delete. Clearing from inside the engine models unavailable functions.
    vm.runInContext("mcpDispatch = undefined; echo = undefined;", host.context);
    assert.deepEqual(host.run(script), expected);
    assert.deepEqual(host.loads, [hostPath, hostPath]);
});

test("dispatch reloads when the dispatcher survives but the requested function is missing", function () {
    const hostPath = "/tmp/extension/src/host/premiere.jsx";
    const host = createHost({ [hostPath]: dispatcherFixture });
    const script = buildDispatchScript(hostPath, "echo", "{}");
    assert.equal(host.run(script).success, true);
    vm.runInContext("echo = undefined;", host.context);
    assert.deepEqual(host.run(script), { success: true, data: {} });
    assert.deepEqual(host.loads, [hostPath, hostPath]);
});

test("dispatch preserves host error envelopes", function () {
    const hostPath = "/tmp/extension/src/host/premiere.jsx";
    const fixture = dispatcherFixture + "\nfunction denied() { return JSON.stringify({success:false,error:'No project is open'}); }";
    const host = createHost({ [hostPath]: fixture });
    assert.deepEqual(host.run(buildDispatchScript(hostPath, "denied", "{}")), {
        success: false,
        error: "No project is open",
    });
});

test("invalid function identifiers cannot inject host code", function () {
    for (const name of ["", "1ping", "ping()", "x.y", "ping);evil();//", "ping\n"] ) {
        assert.throws(function () {
            buildDispatchScript("/tmp/host.jsx", name, "{}");
        }, /function|identifier|command|invalid/i, "must reject " + JSON.stringify(name));
    }
});

test("fixture load failures retain their message and ExtendScript line number", function () {
    const hostPath = "/tmp/extension/src/host/broken.jsx";
    const fixture = "var failure = new Error('fixture load exploded'); failure.line = 23; throw failure;";
    const host = createHost({ [hostPath]: fixture });
    for (const script of [
        buildLoadScript(hostPath, ["mcpDispatch"]),
        buildDispatchScript(hostPath, "echo", "{}"),
    ]) {
        const result = host.run(script);
        assert.equal(result.success, false);
        assert.match(result.error, /fixture load exploded/);
        assert.match(result.error, /23/, "preserve the host line number");
    }
});
