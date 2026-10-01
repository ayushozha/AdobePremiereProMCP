"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const { test } = require("node:test");
const hostSource = fs.readFileSync(path.join(__dirname, "../src/host/premiere.jsx"), "utf8");
const coreSource = fs.readFileSync(path.join(__dirname, "../src/host/core.jsx"), "utf8");
const requestedPath = "/tmp/new project.prproj";

function fixture(operation, options = {}) {
    const calls = [];
    const files = new Map(operation === "openProject" || options.fileExists ? [[requestedPath, 100]] : []);
    const previous = { documentID: "previous", path: "/tmp/previous.prproj", name: "Previous",
        closeDocument() { assert.fail("must not close a previously open project"); } };
    const target = { documentID: "new-document", path: options.actualPath || requestedPath, name: "Requested",
        closeDocument() { assert.fail("must not close the requested project"); } };
    const projects = [previous];
    if (options.alreadyOpen) projects.push(target);
    if (options.duplicatePath) projects.push({ ...target, documentID: "other-document" }, target);
    if (options.duplicateID) projects.push({ ...previous });
    if (options.unknownID) projects.push({ path: "/tmp/unknown.prproj" });
    let active = options.activeExisting ? target : previous;
    Object.defineProperty(projects, "numProjects", { get() { return projects.length; } });
    function mutate() {
        calls.push(operation);
        if (options.action) options.action({ projects, previous, target, files, setActive(value) { active = value; } });
        else if (!options.noDocument) {
            projects.push(target);
            if (!options.noFile) files.set(target.path, options.emptyFile ? 0 : 321);
            if (!options.inactive) active = target;
        }
        return Object.hasOwn(options, "status") ? options.status : true;
    }
    const app = { projects: options.noCollection ? null : projects, newProject: mutate, openDocument: mutate };
    Object.defineProperty(app, "project", { get() { return active; }, set() { assert.fail("must not guess an activation setter"); } });
    const context = vm.createContext({ app, Folder: { fs: options.windows ? "Windows" : "Macintosh" },
        File: function (filename) {
            this.fsName = path.posix.normalize(String(filename));
            const key = options.windows ? [...files.keys()].find(value => value.toLowerCase() === this.fsName.toLowerCase()) : this.fsName;
            this.exists = files.has(key);
            this.length = files.get(key) || 0;
        } });
    vm.runInContext(hostSource, context);
    return { calls, projects, previous, target, files, context,
        dispatch(value) { if (arguments.length === 0) value = requestedPath; return JSON.parse(context.mcpDispatch(operation, JSON.stringify({ path: value }))); } };
}

for (const operation of ["newProject", "openProject"]) {
    test(operation + " succeeds only with strict true and exact active document/file readback", () => {
        const { dispatch, calls, projects, previous } = fixture(operation);
        const result = dispatch();
        assert.equal(result.success, true);
        assert.deepEqual(calls, [operation]);
        assert.equal(result.data.path, requestedPath);
        assert.equal(result.data.projectName, "Requested");
        assert.equal(result.data.documentID, "new-document");
        assert.equal(result.data.active, true);
        assert.equal(result.data.created, operation === "newProject");
        assert.equal(result.data.nativeStatus, true);
        assert.equal(result.data.fileSize, 321);
        assert.equal(result.data.verified, true);
        assert.equal(projects[0], previous);
    });

    test(operation + " preserves confirmed creation/open evidence but fails while another project is active", () => {
        const { dispatch, projects, previous } = fixture(operation, { inactive: true });
        const result = dispatch();
        assert.equal(result.success, false);
        assert.match(result.error, /Focus its Project panel/);
        assert.match(result.error, /Do not repeat newProject/);
        assert.ok(result.error.includes(requestedPath));
        assert.equal(result.data.opened, true);
        assert.equal(result.data.created, operation === "newProject");
        assert.equal(result.data.active, false);
        assert.equal(result.data.documentID, "new-document");
        assert.equal(result.data.activeDocumentID, previous.documentID);
        assert.equal(projects.length, 2);
    });

    for (const status of [false, 0, 1, undefined, null, "true", "0"]) {
        test(operation + " rejects unsupported status " + String(status) + " even after apparent document creation", () => {
            const { dispatch } = fixture(operation, { status });
            const result = dispatch();
            assert.equal(result.success, false);
            assert.match(result.error, /failed with status/);
            assert.equal(result.data.created, false);
            assert.equal(result.data.opened, true);
        });
    }

    test(operation + " rejects true without the requested document", () => {
        const { dispatch } = fixture(operation, { noDocument: true });
        assert.match(dispatch().error, /requested project path is absent/);
    });

    test(operation + " rejects a true result that opens a different project path", () => {
        const { dispatch } = fixture(operation, { actualPath: "/tmp/different.prproj" });
        assert.match(dispatch().error, /requested project path is absent/);
    });

    test(operation + " fails before mutation if the collection is unavailable", () => {
        const { dispatch, calls } = fixture(operation, { noCollection: true });
        assert.match(dispatch().error, /collection is unavailable/);
        assert.deepEqual(calls, []);
    });

    for (const option of ["duplicateID", "unknownID", "duplicatePath"]) {
        test(operation + " rejects " + option + " before mutation", () => {
            const { dispatch, calls } = fixture(operation, { [option]: true });
            assert.equal(dispatch().success, false);
            assert.deepEqual(calls, []);
        });
    }

    test(operation + " detects a previously open project disappearing, without rollback or close calls", () => {
        const { dispatch, calls } = fixture(operation, { action({ projects, target, files, setActive }) {
            projects.splice(0, projects.length, target);
            files.set(target.path, 321);
            setActive(target);
        } });
        assert.match(dispatch().error, /previously open project changed or disappeared/);
        assert.deepEqual(calls, [operation]);
    });

    test(operation + " detects multiple new documents instead of guessing a target", () => {
        const { dispatch } = fixture(operation, { action({ projects, target, files, setActive }) {
            projects.push(target, { documentID: "extra", path: "/tmp/extra.prproj" });
            files.set(target.path, 321);
            setActive(target);
        } });
        assert.match(dispatch().error, /exactly one new document/);
    });

    test(operation + " does not reuse a previous document identity at a changed path", () => {
        const { dispatch } = fixture(operation, { action({ projects, previous, files, setActive }) {
            previous.path = requestedPath;
            files.set(requestedPath, 321);
            setActive(previous);
        } });
        assert.match(dispatch().error, /previously open project changed or disappeared/);
    });

    for (const windows of [false, true]) {
        test(operation + " path case follows " + (windows ? "Windows" : "case-sensitive") + " comparison", () => {
            const { dispatch } = fixture(operation, { windows, actualPath: "/tmp/New Project.prproj" });
            const result = dispatch();
            assert.equal(result.success, windows);
            if (!windows) assert.match(result.error, /requested project path is absent/);
        });
    }

    test(operation + " validates empty and non-string paths without mutation", () => {
        const { dispatch, calls } = fixture(operation);
        for (const value of ["", "  ", undefined, 123, null]) assert.equal(dispatch(value).success, false);
        assert.deepEqual(calls, []);
    });
}

for (const noFile of [true, false]) {
    test("newProject rejects " + (noFile ? "missing" : "empty") + " saved files despite true status and matching identity", () => {
        const { dispatch } = fixture("newProject", { noFile, emptyFile: !noFile });
        assert.match(dispatch().error, /file is missing or empty after/);
    });
}

test("newProject refuses an existing output file before native mutation", () => {
    const { dispatch, calls } = fixture("newProject", { fileExists: true });
    assert.match(dispatch().error, /Project already exists; use openProject/);
    assert.deepEqual(calls, []);
});

test("newProject refuses an already-open document and returns its identity without recreating it", () => {
    const { dispatch, calls } = fixture("newProject", { alreadyOpen: true });
    const result = dispatch();
    assert.equal(result.success, false);
    assert.equal(result.data.documentID, "new-document");
    assert.equal(result.data.alreadyOpen, true);
    assert.equal(result.data.created, false);
    assert.deepEqual(calls, []);
});

for (const activeExisting of [false, true]) {
    test("openProject verifies an already-open document without reopening it; active=" + activeExisting, () => {
        const { dispatch, calls } = fixture("openProject", { alreadyOpen: true, activeExisting, status: false });
        const result = dispatch();
        assert.equal(result.success, activeExisting);
        assert.equal(result.data.alreadyOpen, true);
        assert.equal(result.data.nativeStatus, null);
        assert.equal(result.data.documentID, "new-document");
        assert.equal(result.data.active, activeExisting);
        assert.deepEqual(calls, []);
    });
}

test("openProject rejects a missing file before calling native openDocument", () => {
    const { dispatch, calls, files } = fixture("openProject");
    files.clear();
    assert.match(dispatch().error, /file is missing or empty/);
    assert.deepEqual(calls, []);
});

test("path canonicalization accepts an equivalent dot-segment path", () => {
    const { dispatch } = fixture("newProject");
    const result = dispatch("/tmp/sibling/../new project.prproj");
    assert.equal(result.success, true);
    assert.equal(result.data.path, requestedPath);
});

test("core-only lifecycle calls fail safely and require the verified full dispatcher", () => {
    const context = vm.createContext({ app: { newProject() { assert.fail("unverified native creation"); }, openDocument() { assert.fail("unverified native opening"); } } });
    vm.runInContext(coreSource, context);
    for (const operation of ["newProject", "openProject"]) {
        const result = JSON.parse(context[operation](JSON.stringify({ path: requestedPath })));
        assert.equal(result.success, false);
        assert.match(result.error, /requires the full host dispatcher/);
    }
});

for (const mismatch of ["documentID", "path"]) {
    test("active project must match both target document identity and path: wrong " + mismatch, () => {
        const { dispatch } = fixture("newProject", { action({ projects, target, files, setActive }) {
            projects.push(target);
            files.set(target.path, 321);
            setActive({ ...target, [mismatch]: mismatch === "documentID" ? "other-document" : "/tmp/other.prproj" });
        } });
        const result = dispatch();
        assert.equal(result.success, false);
        assert.equal(result.data.created, true);
        assert.equal(result.data.active, false);
        assert.match(result.error, /Focus its Project panel/);
    });
}
