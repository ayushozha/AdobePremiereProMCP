"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");
const source = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");

function fixture(options = {}) {
    const calls = [];
    const other = { name: "Same project name", documentID: "other-document", path: "/other.prproj" };
    const target = {
        name: "Same project name", documentID: "target-document", path: "/target.prproj",
        save() {
            calls.push("save");
            if (options.switchAfterSave) app.project = { ...other, path: target.path };
            return true;
        },
        closeDocument(saveFirst, promptIfDirty) {
            calls.push([saveFirst, promptIfDirty]);
            const status = Object.prototype.hasOwnProperty.call(options, "status") ? options.status : true;
            if (status === true || status === 0) {
                if (!options.retainInCollection) projects.splice(projects.findIndex(p => p.documentID === target.documentID || p.path === target.path), 1);
                if (!options.retainActive) app.project = options.keepOther || options.retainInCollection ? other : null;
                if (options.readbackThrows) Object.defineProperty(app, "projects", { get() { throw new Error("Host project collection expired"); } });
            }
            return status;
        },
    };
    if (options.noDocumentID) delete target.documentID;
    if (options.noIdentity) { delete target.documentID; target.path = ""; }
    const projects = [options.wrapper ? { ...target } : target];
    if (options.keepOther || options.retainInCollection) projects.push(other);
    if (options.unknownProject) projects.push({});
    Object.defineProperty(projects, "numProjects", { get() { return projects.length; } });
    const app = { project: target, projects: options.noCollection ? undefined : projects };
    const context = vm.createContext({ app, Folder: { fs: "Macintosh" }, File: function (filename) {
        this.fsName = filename; this.exists = true; this.length = 100;
    } });
    vm.runInContext(source, context);
    return { app, target, other, projects, calls,
        close(saveFirst = false) { return JSON.parse(context.mcpDispatch("closeProject", JSON.stringify({ saveFirst }))); },
        readback() { return JSON.parse(context.mcpDispatch("_mcpProjectCloseReadback", "{}")); } };
}

test("the reload diagnostic returns valid JSON without saving or closing", () => {
    const { readback, calls } = fixture({ keepOther: true });
    assert.deepEqual(readback(), { success: true, data: {
        documentID: "target-document", projectPath: "/target.prproj", openProjects: 2, matchingDocuments: 1,
    } });
    assert.deepEqual(calls, []);
});

for (const status of [true, 0]) {
    test("close status " + status + " succeeds only with absent-project readback", () => {
        const { close, calls, app } = fixture({ status });
        const result = close();
        assert.equal(result.success, true);
        assert.deepEqual(result.data, { closed: true, projectName: "Same project name", savedFirst: false, closeStatus: status, verified: true });
        assert.equal(app.project, null);
        assert.equal(app.projects.numProjects, 0);
        assert.deepEqual(calls, [[0, 0]]);
    });
}

for (const status of [false, undefined, null, 1, "true", "0"]) {
    test("close status " + typeof status + " " + String(status) + " cannot report success", () => {
        const { close, calls, app, target } = fixture({ status });
        const result = close();
        assert.equal(result.success, false);
        assert.match(result.error, /close failed with status/);
        assert.equal(app.project, target);
        assert.deepEqual(calls, [[0, 0]]);
    });
}

test("a changed active project cannot hide the original project remaining open", () => {
    const { close, app, other, projects } = fixture({ retainInCollection: true });
    const result = close();
    assert.equal(app.project, other);
    assert.equal(projects.numProjects, 2);
    assert.equal(result.success, false);
    assert.match(result.error, /requested project remains open/);
});

test("a retained active project contradicts an empty project collection", () => {
    const { close } = fixture({ retainActive: true });
    const result = close();
    assert.equal(result.success, false);
    assert.match(result.error, /requested project remains open/);
});

test("other open projects with the same display name do not prevent verified closure", () => {
    const { close, projects, app, other } = fixture({ keepOther: true, wrapper: true });
    const result = close();
    assert.equal(result.success, true);
    assert.deepEqual(projects.slice(), [other]);
    assert.equal(app.project, other);
});

test("a saved project path identifies hosts without document IDs", () => {
    const { close } = fixture({ noDocumentID: true, wrapper: true });
    assert.equal(close().success, true);
});

for (const option of ["noCollection", "noIdentity", "unknownProject"]) {
    test(option + " fails before a close whose readback cannot be verified", () => {
        const { close, calls } = fixture({ [option]: true });
        const result = close();
        assert.equal(result.success, false);
        assert.match(result.error, /closure cannot be verified/);
        assert.deepEqual(calls, []);
    });
}

test("a successful host status with failed readback is still an error", () => {
    const { close } = fixture({ readbackThrows: true });
    const result = close();
    assert.equal(result.success, false);
    assert.match(result.error, /Host project collection expired/);
});

test("saveFirst cannot close a different project after the save changes active identity", () => {
    const { close, calls } = fixture({ switchAfterSave: true });
    const result = close(true);
    assert.equal(result.success, false);
    assert.match(result.error, /active project changed.*no close was attempted/);
    assert.deepEqual(calls, ["save"]);
});
