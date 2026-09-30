"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");

const hostSource = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");
const originalPath = "/tmp/original.prproj";
const newPath = "/tmp/new cut.prproj";

function fixture(status, options = {}) {
    const files = new Map([[originalPath, 100]]);
    const calls = [];
    const project = {
        name: "Fixture",
        path: options.unsaved ? "" : originalPath,
        save() {
            calls.push("save");
            if (options.save) options.save(project, files);
            return status;
        },
        saveAs(requestedPath) {
            calls.push({ saveAs: requestedPath });
            if (options.saveAs) options.saveAs(project, files, requestedPath);
            else {
                project.path = requestedPath;
                files.set(requestedPath, 321);
            }
            return status;
        },
        closeDocument() { calls.push("closeDocument"); return true; },
    };
    // The ExtendScript Project API has no portable dirty-state readback.
    // Tests fail if save verification starts depending on a guessed property.
    Object.defineProperty(project, "isDirty", { get() { assert.fail("unsupported dirty-state access"); } });
    const context = vm.createContext({
        app: { project },
        Folder: { fs: options.windows ? "Windows" : "Macintosh" },
        File: function (filename) {
            this.fsName = filename;
            this.exists = files.has(filename);
            this.length = files.get(filename) || 0;
        },
    });
    vm.runInContext(hostSource, context);
    const dispatch = (name, args = {}) => JSON.parse(context.mcpDispatch(name, JSON.stringify(args)));
    return { context, project, files, calls, dispatch };
}

for (const status of [0, true]) {
    test("saveProject accepts " + status + " and returns file/path readback", () => {
        const { dispatch, calls } = fixture(status, { save: (_project, files) => files.set(originalPath, 321) });
        const result = dispatch("saveProject");
        assert.equal(result.success, true);
        assert.deepEqual(calls, ["save"]);
        assert.deepEqual(result.data, { saved: true, projectName: "Fixture", projectPath: originalPath, fileSize: 321, saveStatus: status });
    });

    test("saveProjectAs accepts " + status + " with the exact saved path", () => {
        const { dispatch, calls } = fixture(status);
        const result = dispatch("saveProjectAs", { path: newPath });
        assert.equal(result.success, true);
        assert.deepEqual(calls, [{ saveAs: newPath }]);
        assert.deepEqual(result.data, { saved: true, newPath, projectName: "Fixture", fileSize: 321, saveStatus: status });
    });
}

for (const status of [false, 1, undefined, null, "0", "true"]) {
    test("save APIs reject status " + String(status) + " despite existing output files", () => {
        const { dispatch } = fixture(status);
        const saved = dispatch("saveProject");
        const savedAs = dispatch("saveProjectAs", { path: newPath });
        assert.equal(saved.success, false);
        assert.equal(savedAs.success, false);
        assert.match(saved.error, /failed with status/);
        assert.match(savedAs.error, /failed with status/);
    });
}

test("saving an unsaved project fails before invoking a modal save", () => {
    const { dispatch, calls } = fixture(true, { unsaved: true });
    const result = dispatch("saveProject");
    assert.equal(result.success, false);
    assert.match(result.error, /use saveProjectAs/);
    assert.deepEqual(calls, []);
});

test("saveProject rejects a changed project path despite an accepted status", () => {
    const { dispatch } = fixture(true, { save: (project, files) => {
        project.path = newPath;
        files.set(newPath, 100);
    } });
    const result = dispatch("saveProject");
    assert.equal(result.success, false);
    assert.match(result.error, /path readback mismatch/);
});

for (const size of [undefined, 0]) {
    test("save APIs reject " + (size === undefined ? "missing" : "empty") + " files after success status", () => {
        const { dispatch } = fixture(0, {
            save: (_project, files) => { if (size === undefined) files.delete(originalPath); else files.set(originalPath, size); },
            saveAs: (project, files, requestedPath) => {
                project.path = requestedPath;
                if (size !== undefined) files.set(requestedPath, size);
            },
        });
        assert.match(dispatch("saveProject").error, /file is missing or empty/);
        assert.match(dispatch("saveProjectAs", { path: newPath }).error, /file is missing or empty/);
    });
}

test("saveProjectAs rejects a stale project path even when the requested file exists", () => {
    const { dispatch } = fixture(true, { saveAs: (_project, files, requestedPath) => files.set(requestedPath, 321) });
    const result = dispatch("saveProjectAs", { path: newPath });
    assert.equal(result.success, false);
    assert.match(result.error, /path readback mismatch/);
});

for (const windows of [false, true]) {
    test("saveProjectAs path case follows " + (windows ? "Windows" : "case-sensitive") + " comparison", () => {
        const { dispatch } = fixture(true, { windows, saveAs: (project, files) => {
            project.path = "/tmp/New Cut.prproj";
            files.set(project.path, 321);
        } });
        const result = dispatch("saveProjectAs", { path: newPath });
        assert.equal(result.success, windows);
        if (!windows) assert.match(result.error, /path readback mismatch/);
    });
}

test("saveProjectAs validates its path before mutating the project", () => {
    const { dispatch, calls } = fixture(true);
    for (const requestedPath of [undefined, "", 123]) {
        assert.equal(dispatch("saveProjectAs", { path: requestedPath }).success, false);
    }
    assert.deepEqual(calls, []);
});

for (const status of [0, true]) {
    test("closeProject saves and verifies before closing for status " + status, () => {
        const { dispatch, calls } = fixture(status);
        const result = dispatch("closeProject", { saveFirst: true });
        assert.equal(result.success, true);
        assert.equal(result.data.savedFirst, true);
        assert.deepEqual(calls, ["save", "closeDocument"]);
    });
}

for (const status of [false, 1, undefined]) {
    test("closeProject never closes after failed save status " + String(status), () => {
        const { dispatch, calls } = fixture(status);
        const result = dispatch("closeProject", { saveFirst: "true" });
        assert.equal(result.success, false);
        assert.match(result.error, /save before close failed/);
        assert.deepEqual(calls, ["save"]);
    });
}

test("closeProject never closes after a successful status with missing file readback", () => {
    const { dispatch, calls } = fixture(true, { save: (_project, files) => files.delete(originalPath) });
    const result = dispatch("closeProject", { saveFirst: true });
    assert.equal(result.success, false);
    assert.match(result.error, /file is missing or empty/);
    assert.deepEqual(calls, ["save"]);
});

test("closeProject without saveFirst retains the direct close behavior", () => {
    const { dispatch, calls } = fixture(false, { unsaved: true });
    const result = dispatch("closeProject", { saveFirst: false });
    assert.equal(result.success, true);
    assert.equal(result.data.savedFirst, false);
    assert.deepEqual(calls, ["closeDocument"]);
});
