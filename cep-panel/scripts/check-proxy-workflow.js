#!/usr/bin/env node

var assert = require("assert");
var fs = require("fs");
var path = require("path");
var vm = require("vm");

var hostPath = path.resolve(__dirname, "..", "src", "host", "premiere.jsx");
var hostSource = fs.readFileSync(hostPath, "utf8");
var existingFiles = {};
var existingFolders = {};

function MockFile(value) {
    var canonicalPath = path.resolve(String(value));
    this.fsName = canonicalPath;
    this.name = path.basename(canonicalPath);
    this.exists = existingFiles[canonicalPath] === true;
    this.parent = {
        fsName: path.dirname(canonicalPath),
        exists: existingFolders[path.dirname(canonicalPath)] === true
    };
}

function markFile(filePath) {
    existingFiles[path.resolve(filePath)] = true;
}

function markFolder(folderPath) {
    existingFolders[path.resolve(folderPath)] = true;
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

function makeItem(options) {
    options = options || {};
    var state = {
        hasProxy: options.hasProxy === true,
        proxyPath: options.proxyPath || ""
    };
    return {
        name: "Camera A",
        state: state,
        canProxy: options.canProxy || function () { return true; },
        hasProxy: options.hasProxyFn || function () { return state.hasProxy; },
        getProxyPath: options.getProxyPath || function () { return state.proxyPath || 0; },
        attachProxy: options.attachProxy || function (proxyPath, isHiRes) {
            assert.strictEqual(isHiRes, 0, "attachProxy must identify media as proxy, not high resolution");
            state.hasProxy = true;
            state.proxyPath = proxyPath;
            return 0;
        },
        detachProxy: options.detachProxy || function () {
            state.hasProxy = false;
            state.proxyPath = "";
            return 0;
        }
    };
}

function setProjectItem(item, encoder) {
    var children = [item];
    children.numItems = children.length;
    context.app = { project: { rootItem: { children: children } }, encoder: encoder };
}

function call(name) {
    var args = Array.prototype.slice.call(arguments, 1);
    return JSON.parse(context[name].apply(null, args));
}

function expectFailure(result, messagePattern) {
    assert.strictEqual(result.success, false, JSON.stringify(result));
    assert.match(result.error, messagePattern);
}

markFile("/tmp/proxy-preset.epr");
markFile("/tmp/proxy.mov");
markFile("/tmp/different.mov");
markFile("/tmp/existing-output.mov");
markFolder("/tmp");

var item = makeItem();
setProjectItem(item);
expectFailure(call("hasProxy", -1), /non-negative integer/);
expectFailure(call("hasProxy", 0.5), /non-negative integer/);
expectFailure(call("hasProxy", 1), /out of range/);
item.canProxy = function () { return false; };
expectFailure(call("hasProxy", 0), /cannot use proxy media/);

item = makeItem();
setProjectItem(item);
expectFailure(call("createProxy", 0, "relative.mov", "/tmp/proxy-preset.epr"), /outputPath must be an absolute file path/);
expectFailure(call("createProxy", 0, "/missing/output.mov", "/tmp/proxy-preset.epr"), /parent directory does not exist/);
expectFailure(call("createProxy", 0, "/tmp/no-extension", "/tmp/proxy-preset.epr"), /must include a file extension/);
expectFailure(call("createProxy", 0, "/tmp/existing-output.mov", "/tmp/proxy-preset.epr"), /refusing to overwrite/);
expectFailure(call("createProxy", 0, "/tmp/generated-proxy.mov", "relative.epr"), /absolute file path/);
expectFailure(call("createProxy", 0, "/tmp/generated-proxy.mov", "/tmp/missing.epr"), /does not exist/);
expectFailure(call("createProxy", 0, "/tmp/generated-proxy.mov", "/tmp/proxy.mov"), /must be a \.epr file/);

var encodeCalls = 0;
var batchCalls = 0;
item = makeItem();
setProjectItem(item, {
    launchEncoder: function () { return 0; },
    encodeProjectItem: function (projectItem, outputPath, presetPath, workArea, removeUponCompletion) {
        encodeCalls++;
        assert.strictEqual(projectItem, item);
        assert.strictEqual(outputPath, "/tmp/generated-proxy.mov");
        assert.strictEqual(presetPath, "/tmp/proxy-preset.epr");
        assert.strictEqual(workArea, 0);
        assert.strictEqual(removeUponCompletion, 0);
        return "proxy-job-42";
    },
    startBatch: function () {
        batchCalls++;
        return 0;
    }
});
var pendingCreate = JSON.parse(context.mcpDispatch("createProxy", JSON.stringify({
    project_item_index: 0,
    output_path: "/tmp/generated-proxy.mov",
    preset_path: "/tmp/proxy-preset.epr"
})));
assert.strictEqual(pendingCreate.success, true, JSON.stringify(pendingCreate));
assert.strictEqual(encodeCalls, 1);
assert.strictEqual(batchCalls, 1);
assert.strictEqual(pendingCreate.data.requestAccepted, true);
assert.strictEqual(pendingCreate.data.jobId, "proxy-job-42");
assert.strictEqual(pendingCreate.data.outputPath, "/tmp/generated-proxy.mov");
assert.strictEqual(pendingCreate.data.state, "pending_encode");
assert.strictEqual(pendingCreate.data.encodePending, true);
assert.strictEqual(pendingCreate.data.attachRequired, true);
assert.strictEqual(pendingCreate.data.attached, false);
assert.strictEqual(pendingCreate.data.verified, false);
assert.strictEqual(Object.prototype.hasOwnProperty.call(pendingCreate.data, "proxyCreated"), false);

item = makeItem();
setProjectItem(item, {
    launchEncoder: function () { return 0; },
    encodeProjectItem: function () { return 0; },
    startBatch: function () { throw new Error("should not start a failed job"); }
});
expectFailure(call("createProxy", 0, "/tmp/generated-proxy.mov", "/tmp/proxy-preset.epr"), /did not queue/);

item = makeItem();
setProjectItem(item, {
    launchEncoder: function () { return 0; },
    encodeProjectItem: function () { return "proxy-job-43"; },
    startBatch: function () { return 1; }
});
expectFailure(call("createProxy", 0, "/tmp/generated-proxy.mov", "/tmp/proxy-preset.epr"), /job proxy-job-43 was queued.*status 1/);

item = makeItem({ hasProxyFn: undefined });
item.hasProxy = undefined;
setProjectItem(item);
expectFailure(call("hasProxy", 0), /hasProxy is not supported/);

item = makeItem({
    attachProxy: function () { return 1; }
});
setProjectItem(item);
expectFailure(call("attachProxy", 0, "/tmp/proxy.mov"), /status 1/);

item = makeItem({
    attachProxy: function () {
        this.state.hasProxy = true;
        this.state.proxyPath = "/tmp/different.mov";
        return 0;
    }
});
setProjectItem(item);
expectFailure(call("attachProxy", 0, "/tmp/proxy.mov"), /readback mismatch/);

item = makeItem();
setProjectItem(item);
var attached = call("attachProxy", 0, "/tmp/proxies/../proxy.mov");
assert.strictEqual(attached.success, true, JSON.stringify(attached));
assert.strictEqual(attached.data.proxyPath, "/tmp/proxy.mov");
assert.strictEqual(attached.data.hasProxy, true);
assert.strictEqual(attached.data.verified, true);
var attachedState = call("getProxyPath", 0);
assert.strictEqual(attachedState.success, true, JSON.stringify(attachedState));
assert.strictEqual(attachedState.data.proxyPath, "/tmp/proxy.mov");

var attachCalls = 0;
var detachCalls = 0;
item = makeItem({
    hasProxy: true,
    proxyPath: "/tmp/proxy.mov",
    attachProxy: function () { attachCalls++; return 0; },
    detachProxy: function () {
        detachCalls++;
        this.state.hasProxy = false;
        this.state.proxyPath = "";
        return 0;
    }
});
setProjectItem(item);
var detached = call("detachProxy", 0);
assert.strictEqual(detached.success, true, JSON.stringify(detached));
assert.strictEqual(detached.data.verified, true);
assert.strictEqual(detached.data.previousProxyPath, "/tmp/proxy.mov");
assert.strictEqual(attachCalls, 0, "detach must not attach an empty path");
assert.strictEqual(detachCalls, 1);

item = makeItem({
    hasProxy: true,
    proxyPath: "/tmp/proxy.mov",
    detachProxy: function () { return 1; }
});
setProjectItem(item);
expectFailure(call("detachProxy", 0), /status 1/);

var proxyPlaybackState = 0;
context.app = {
    project: {},
    getEnableProxies: function () { return proxyPlaybackState; },
    setEnableProxies: function (enabled) { proxyPlaybackState = enabled; return 1; }
};
var enabled = call("toggleProxies", true);
assert.strictEqual(enabled.success, true, JSON.stringify(enabled));
assert.strictEqual(enabled.data.proxiesEnabled, true);
assert.strictEqual(enabled.data.verified, true);

context.app = {
    project: {},
    getEnableProxies: function () { return 0; },
    setEnableProxies: function () { return 1; }
};
expectFailure(call("toggleProxies", true), /readback mismatch/);

context.app = { project: {}, setEnableProxies: function () { return 1; } };
expectFailure(call("toggleProxies", true), /Verified proxy playback control is not supported/);

context.app = {
    project: {},
    getEnableProxies: function () { return 0; },
    setEnableProxies: function () { throw new Error("should not be called"); }
};
expectFailure(call("toggleProxies", "true"), /must be a boolean/);

console.log("Verified proxy workflow smoke tests passed.");
