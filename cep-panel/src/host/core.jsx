// Core ExtendScript functions for PremierPro MCP Bridge
// This minimal file loads fast and provides essential functions.

function _ok(data) { return JSON.stringify({ success: true, data: data }); }
function _err(message) { return JSON.stringify({ success: false, error: String(message) }); }

// JSON polyfill for older ExtendScript
if (typeof JSON === "undefined") {
    JSON = {
        stringify: function(obj) {
            if (obj === null) return "null";
            if (typeof obj === "string") return '"' + obj.replace(/"/g, '\\"') + '"';
            if (typeof obj === "number" || typeof obj === "boolean") return String(obj);
            if (obj instanceof Array) {
                var a = [];
                for (var i = 0; i < obj.length; i++) a.push(JSON.stringify(obj[i]));
                return "[" + a.join(",") + "]";
            }
            if (typeof obj === "object") {
                var p = [];
                for (var k in obj) if (obj.hasOwnProperty(k)) p.push('"' + k + '":' + JSON.stringify(obj[k]));
                return "{" + p.join(",") + "}";
            }
            return '""';
        },
        parse: function(s) { return eval("(" + s + ")"); }
    };
}

function ping() {
    try {
        var ver = "unknown";
        try { ver = app.version; } catch(e) {}
        var projOpen = false;
        try { projOpen = app.project && app.project.name ? true : false; } catch(e) {}
        return _ok({
            premiere_running: true,
            premiere_version: ver,
            project_open: projOpen,
            project_name: projOpen ? app.project.name : ""
        });
    } catch(e) { return _err(e.message); }
}

function getProjectInfo() {
    try {
        if (!app.project) return _err("No project open");
        var seqs = [];
        for (var i = 0; i < app.project.sequences.numItems; i++) {
            var s = app.project.sequences[i];
            seqs.push({ index: i, name: s.name, id: s.sequenceID });
        }
        return _ok({
            name: app.project.name,
            path: app.project.path,
            sequences: seqs,
            sequence_count: app.project.sequences.numItems
        });
    } catch(e) { return _err(e.message); }
}

function getProjectState() {
    return getProjectInfo();
}

function newProject(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        app.newProject(args.path || "");
        return _ok({ message: "Project created", path: args.path });
    } catch(e) { return _err(e.message); }
}

function openProject(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        app.openDocument(args.path);
        return _ok({ message: "Project opened", path: args.path });
    } catch(e) { return _err(e.message); }
}

function saveProject() {
    try {
        app.project.save();
        return _ok({ message: "Project saved" });
    } catch(e) { return _err(e.message); }
}

function createSequence(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var name = args.name || "New Sequence";
        app.project.createNewSequence(name, name);
        var seq = app.project.activeSequence;
        return _ok({
            name: seq.name,
            id: seq.sequenceID,
            width: seq.frameSizeHorizontal,
            height: seq.frameSizeVertical
        });
    } catch(e) { return _err(e.message); }
}

function getActiveSequence() {
    try {
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        return _ok({
            name: seq.name,
            id: seq.sequenceID,
            width: seq.frameSizeHorizontal,
            height: seq.frameSizeVertical,
            duration: seq.end
        });
    } catch(e) { return _err(e.message); }
}

function importFiles(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var paths = args.paths || args.filePaths || [args.path || args.filePath];
        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " files", count: paths.length });
    } catch(e) { return _err(e.message); }
}

function getSequenceList() {
    try {
        var seqs = [];
        for (var i = 0; i < app.project.sequences.numItems; i++) {
            var s = app.project.sequences[i];
            seqs.push({ index: i, name: s.name, id: s.sequenceID });
        }
        return _ok({ sequences: seqs, count: seqs.length });
    } catch(e) { return _err(e.message); }
}

// Generic eval for any function not defined in core
function evalDynamic(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        return eval(args.script);
    } catch(e) { return _err(e.message); }
}
