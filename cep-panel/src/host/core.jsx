// Core ExtendScript functions for PremierPro MCP Bridge
// This minimal file loads fast and provides essential functions.
// All functions are hardened with input validation, null safety, and contextual errors.

function _ok(data) { return JSON.stringify({ success: true, data: data }); }
function _err(message) { return JSON.stringify({ success: false, error: String(message) }); }

// JSON polyfill for older ExtendScript
if (typeof JSON === "undefined") {
    JSON = {
        stringify: function(obj) {
            if (obj === null) return "null";
            if (typeof obj === "undefined") return undefined;
            if (typeof obj === "string") return '"' + obj.replace(/\\/g, "\\\\").replace(/"/g, '\\"').replace(/\n/g, "\\n").replace(/\r/g, "\\r").replace(/\t/g, "\\t") + '"';
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

// -- Shared Helpers ----------------------------------------------------

/**
 * Safely parse JSON arguments and validate required fields.
 * Returns the parsed object on success, or an object with an .error string on failure.
 */
function _parseArgs(argsJson, requiredFields) {
    if (argsJson === undefined || argsJson === null || argsJson === "") {
        return { error: "No arguments provided" };
    }
    var args;
    try { args = JSON.parse(argsJson); }
    catch (e) { return { error: "Invalid JSON arguments: " + e.message }; }
    if (requiredFields) {
        for (var i = 0; i < requiredFields.length; i++) {
            if (args[requiredFields[i]] === undefined || args[requiredFields[i]] === null) {
                return { error: "Missing required parameter: " + requiredFields[i] };
            }
        }
    }
    return args;
}

/**
 * Return the active sequence or null, guarding against app.project being null.
 */
function _getActiveSequence() {
    if (!app.project) return null;
    return app.project.activeSequence || null;
}

/**
 * Require app.project to be open. Returns null if OK, or an error-result string.
 */
function _requireProject() {
    if (!app.project) return _err("No project is open. Open or create a project first.");
    return null;
}

/**
 * Require an active sequence. Returns the sequence on success, or null.
 * If null, caller should return _err("No active sequence...").
 */
function _requireSequence() {
    if (!app.project) return null;
    return app.project.activeSequence || null;
}

// -- Project -----------------------------------------------------------

function ping() {
    try {
        var ver = "unknown";
        try { ver = app.version; } catch (e1) {}
        var projOpen = false;
        var projName = "";
        try {
            projOpen = !!(app.project && app.project.name);
            if (projOpen) projName = app.project.name;
        } catch (e2) {}
        return _ok({ premiere_running: true, premiere_version: ver, project_open: projOpen, project_name: projName });
    } catch (e) { return _err("Ping failed: " + e.message); }
}

function getProjectInfo() {
    try {
        if (!app.project) return _err("No project is open. Open or create a project first.");
        var seqs = [];
        var numSeqs = 0;
        try { numSeqs = app.project.sequences.numItems; } catch (e1) {}
        for (var i = 0; i < numSeqs; i++) {
            var s = app.project.sequences[i];
            if (s) {
                seqs.push({ index: i, name: s.name, id: s.sequenceID });
            }
        }
        return _ok({
            name: app.project.name,
            path: app.project.path,
            sequences: seqs,
            sequence_count: numSeqs
        });
    } catch (e) { return _err("Failed to get project info: " + e.message); }
}

function getProjectState() { return getProjectInfo(); }

function newProject(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["path"]);
        if (args.error) return _err(args.error);
        app.newProject(args.path);
        return _ok({ message: "Project created", path: args.path });
    } catch (e) { return _err("Failed to create project at '" + (args && args.path ? args.path : "unknown") + "': " + e.message); }
}

function openProject(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["path"]);
        if (args.error) return _err(args.error);
        var f = new File(args.path);
        if (!f.exists) return _err("Project file not found: " + args.path);
        app.openDocument(args.path);
        var projName = "";
        try { projName = app.project.name; } catch (e1) {}
        return _ok({ message: "Project opened: " + projName, path: args.path, name: projName });
    } catch (e) { return _err("Failed to open project '" + (args && args.path ? args.path : "unknown") + "': " + e.message); }
}

function saveProject() {
    try {
        if (!app.project) return _err("No project is open to save.");
        app.project.save();
        return _ok({ message: "Project saved", name: app.project.name });
    } catch (e) { return _err("Failed to save project: " + e.message); }
}

function saveProjectAs(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["path"]);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open to save.");
        app.project.saveAs(args.path);
        return _ok({ message: "Project saved as", path: args.path });
    } catch (e) { return _err("Failed to save project as '" + (args && args.path ? args.path : "unknown") + "': " + e.message); }
}

function closeProject(argsJson) {
    try {
        if (!app.project) return _err("No project is open to close.");
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        app.project.closeDocument(args.save_first !== false, true);
        return _ok({ message: "Project closed" });
    } catch (e) { return _err("Failed to close project: " + e.message); }
}

// -- Sequences ---------------------------------------------------------

function createSequence(argsJson) {
    try {
        if (!app.project) return _err("No project is open. Open or create a project first.");
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        var name = args.name || "New Sequence";
        app.project.createNewSequence(name, name);
        var seq = _getActiveSequence();
        if (!seq) return _err("Sequence '" + name + "' was created but could not be activated.");
        return _ok({
            name: seq.name,
            id: seq.sequenceID,
            width: seq.frameSizeHorizontal,
            height: seq.frameSizeVertical
        });
    } catch (e) { return _err("Failed to create sequence '" + (name || "New Sequence") + "': " + e.message); }
}

function getActiveSequence() {
    try {
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence. Create or open a sequence first.");
        return _ok({
            name: seq.name,
            id: seq.sequenceID,
            width: seq.frameSizeHorizontal,
            height: seq.frameSizeVertical,
            end: seq.end
        });
    } catch (e) { return _err("Failed to get active sequence: " + e.message); }
}

function getSequenceList() {
    try {
        if (!app.project) return _err("No project is open.");
        var seqs = [];
        var numSeqs = 0;
        try { numSeqs = app.project.sequences.numSequences || app.project.sequences.numItems || 0; } catch (e1) {}
        var activeID = app.project.activeSequence ? (app.project.activeSequence.sequenceID || "") : "";
        for (var i = 0; i < numSeqs; i++) {
            var s = app.project.sequences[i];
            if (s) {
                seqs.push({
                    index: i,
                    name: s.name || "",
                    sequence_id: s.sequenceID || "",
                    frame_size_horizontal: s.frameSizeHorizontal || 0,
                    frame_size_vertical: s.frameSizeVertical || 0,
                    timebase: s.timebase || "",
                    video_track_count: (s.videoTracks && s.videoTracks.numTracks !== undefined) ? s.videoTracks.numTracks : 0,
                    audio_track_count: (s.audioTracks && s.audioTracks.numTracks !== undefined) ? s.audioTracks.numTracks : 0,
                    is_active: (s.sequenceID === activeID)
                });
            }
        }
        return _ok({ sequences: seqs, count: seqs.length, active_sequence_id: activeID });
    } catch (e) { return _err("Failed to list sequences: " + e.message); }
}

// -- Media Import ------------------------------------------------------

function importFiles(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open. Open or create a project first.");

        var paths = args.paths || args.filePaths || (args.path ? [args.path] : (args.filePath ? [args.filePath] : null));
        if (!paths || paths.length === 0) return _err("Missing required parameter: paths (array of file paths)");

        // Validate that files exist
        var missing = [];
        for (var i = 0; i < paths.length; i++) {
            var f = new File(paths[i]);
            if (!f.exists) missing.push(paths[i]);
        }
        if (missing.length > 0) return _err("Files not found: " + missing.join(", "));

        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " file(s)", count: paths.length });
    } catch (e) { return _err("Failed to import files: " + e.message); }
}

function importFolder(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open. Open or create a project first.");

        var folderPath = args.path || args.folderPath;
        if (!folderPath) return _err("Missing required parameter: path");

        var folder = new Folder(folderPath);
        if (!folder.exists) return _err("Folder not found: " + folderPath);

        var files = folder.getFiles();
        var paths = [];
        for (var i = 0; i < files.length; i++) {
            if (files[i] instanceof File) paths.push(files[i].fsName);
        }
        if (paths.length === 0) return _err("No files found in folder: " + folderPath);

        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " file(s) from folder", count: paths.length, folder: folderPath });
    } catch (e) { return _err("Failed to import folder '" + (folderPath || "unknown") + "': " + e.message); }
}

// -- Clips -------------------------------------------------------------

function getTimelineState(argsJson) {
    try {
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence. Create or open a sequence first.");

        var tracks = [];
        var numVideoTracks = 0;
        var numAudioTracks = 0;
        try { numVideoTracks = seq.videoTracks.numTracks; } catch (e1) {}
        try { numAudioTracks = seq.audioTracks.numTracks; } catch (e2) {}

        for (var t = 0; t < numVideoTracks; t++) {
            var vTrack = seq.videoTracks[t];
            if (!vTrack) continue;
            var vClips = [];
            var numVClips = 0;
            try { numVClips = vTrack.clips.numItems; } catch (e3) {}
            for (var c = 0; c < numVClips; c++) {
                var vc = vTrack.clips[c];
                if (!vc) continue;
                vClips.push({
                    index: c,
                    name: vc.name,
                    start: vc.start.seconds,
                    end: vc.end.seconds,
                    duration: vc.duration.seconds
                });
            }
            var vTrackInfo = {
                index: t,
                name: vTrack.name || ("Video " + (t + 1)),
                type: "video",
                clipCount: numVClips,
                clips: vClips
            };
            try { vTrackInfo.isMuted = vTrack.isMuted(); } catch (e4) {}
            try { vTrackInfo.isLocked = vTrack.isLocked(); } catch (e5) {}
            tracks.push(vTrackInfo);
        }

        for (var at = 0; at < numAudioTracks; at++) {
            var aTrack = seq.audioTracks[at];
            if (!aTrack) continue;
            var aClips = [];
            var numAClips = 0;
            try { numAClips = aTrack.clips.numItems; } catch (e6) {}
            for (var ac = 0; ac < numAClips; ac++) {
                var acl = aTrack.clips[ac];
                if (!acl) continue;
                aClips.push({
                    index: ac,
                    name: acl.name,
                    start: acl.start.seconds,
                    end: acl.end.seconds,
                    duration: acl.duration.seconds
                });
            }
            var aTrackInfo = {
                index: at,
                name: aTrack.name || ("Audio " + (at + 1)),
                type: "audio",
                clipCount: numAClips,
                clips: aClips
            };
            try { aTrackInfo.isMuted = aTrack.isMuted(); } catch (e7) {}
            try { aTrackInfo.isLocked = aTrack.isLocked(); } catch (e8) {}
            tracks.push(aTrackInfo);
        }

        return _ok({
            sequence: seq.name,
            sequenceId: seq.sequenceID,
            videoTrackCount: numVideoTracks,
            audioTrackCount: numAudioTracks,
            tracks: tracks
        });
    } catch (e) { return _err("Failed to get timeline state: " + e.message); }
}

function insertClip(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence. Create or open a sequence first.");

        var idx = args.projectItemIndex || 0;
        var numItems = 0;
        try { numItems = app.project.rootItem.children.numItems; } catch (e1) {}
        if (numItems === 0) return _err("Project has no items. Import media first.");
        if (idx < 0 || idx >= numItems) {
            return _err("Project item index " + idx + " out of range (0-" + (numItems - 1) + ")");
        }
        var item = app.project.rootItem.children[idx];
        if (!item) return _err("Project item at index " + idx + " not found");

        seq.insertClip(item, args.time || 0, args.videoTrackIndex || 0, args.audioTrackIndex || 0);
        return _ok({ message: "Clip '" + item.name + "' inserted at " + (args.time || 0) + "s" });
    } catch (e) { return _err("Failed to insert clip: " + e.message); }
}

function placeClip(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence. Create or open a sequence first.");

        var idx = args.projectItemIndex || 0;
        var numItems = 0;
        try { numItems = app.project.rootItem.children.numItems; } catch (e1) {}
        if (numItems === 0) return _err("Project has no items. Import media first.");
        if (idx < 0 || idx >= numItems) {
            return _err("Project item index " + idx + " out of range (0-" + (numItems - 1) + ")");
        }
        var item = app.project.rootItem.children[idx];
        if (!item) return _err("Project item at index " + idx + " not found");

        var ti = args.trackIndex || 0;
        var numVTracks = 0;
        try { numVTracks = seq.videoTracks.numTracks; } catch (e2) {}
        if (ti < 0 || ti >= numVTracks) {
            return _err("Video track index " + ti + " out of range (0-" + (numVTracks - 1) + ")");
        }
        var track = seq.videoTracks[ti];
        if (!track) return _err("Video track " + ti + " not found");

        track.overwriteClip(item, args.startTime || 0);
        return _ok({ message: "Clip '" + item.name + "' placed on video track " + ti + " at " + (args.startTime || 0) + "s" });
    } catch (e) { return _err("Failed to place clip: " + e.message); }
}

// -- Export -------------------------------------------------------------

function exportSequence(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["outputPath"]);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence to export.");

        seq.exportAsMediaDirect(args.outputPath, args.presetPath || "", 0);
        return _ok({ message: "Export started", output: args.outputPath, sequence: seq.name });
    } catch (e) { return _err("Failed to export sequence: " + e.message); }
}

// -- Bins --------------------------------------------------------------

function createBin(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");

        var binName = args.name || "New Bin";
        app.project.rootItem.createBin(binName);
        return _ok({ message: "Bin created", name: binName });
    } catch (e) { return _err("Failed to create bin '" + (binName || "New Bin") + "': " + e.message); }
}

function getProjectItems(argsJson) {
    try {
        if (!app.project) return _err("No project is open.");
        var root = app.project.rootItem;
        if (!root) return _err("Project root item is not accessible.");

        var items = [];
        var numItems = 0;
        try { numItems = root.children.numItems; } catch (e1) {}
        for (var i = 0; i < numItems; i++) {
            var item = root.children[i];
            if (!item) continue;
            items.push({
                index: i,
                name: item.name,
                type: item.type,
                path: item.treePath
            });
        }
        return _ok({ items: items, count: items.length });
    } catch (e) { return _err("Failed to get project items: " + e.message); }
}

// -- Markers -----------------------------------------------------------

function addSequenceMarker(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var timeVal = args.time || 0;
        var marker = seq.markers.createMarker(timeVal);
        if (!marker) return _err("Failed to create marker at " + timeVal + "s");

        if (args.name) marker.name = args.name;
        if (args.comment) marker.comments = args.comment;
        return _ok({ message: "Marker added at " + timeVal + "s", name: args.name || "", comment: args.comment || "" });
    } catch (e) { return _err("Failed to add sequence marker: " + e.message); }
}

// -- Playback ----------------------------------------------------------

function getPlayheadPosition() {
    try {
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var pos = seq.getPlayerPosition();
        if (!pos) return _err("Could not read playhead position.");
        return _ok({ seconds: pos.seconds, ticks: pos.ticks });
    } catch (e) { return _err("Failed to get playhead position: " + e.message); }
}

function setPlayheadPosition(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["seconds"]);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var ticks = String(Math.round(args.seconds * 254016000000));
        seq.setPlayerPosition(ticks);
        return _ok({ message: "Playhead moved to " + args.seconds + "s" });
    } catch (e) { return _err("Failed to set playhead to " + (args && args.seconds !== undefined ? args.seconds + "s" : "unknown position") + ": " + e.message); }
}

// -- Audio -------------------------------------------------------------

function setAudioLevel(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var ti = args.trackIndex || 0;
        var numATracks = 0;
        try { numATracks = seq.audioTracks.numTracks; } catch (e1) {}
        if (ti < 0 || ti >= numATracks) {
            return _err("Audio track " + ti + " out of range (0-" + (numATracks - 1) + ")");
        }
        var track = seq.audioTracks[ti];
        if (!track) return _err("Audio track " + ti + " not accessible.");

        var ci = args.clipIndex || 0;
        var numClips = 0;
        try { numClips = track.clips.numItems; } catch (e2) {}
        if (ci < 0 || ci >= numClips) {
            return _err("Clip " + ci + " out of range on audio track " + ti + " (0-" + (numClips - 1) + ")");
        }
        var clip = track.clips[ci];
        if (!clip) return _err("Clip " + ci + " on audio track " + ti + " not accessible.");

        if (!clip.components || clip.components.numItems === 0) {
            return _err("Clip '" + clip.name + "' has no audio components.");
        }
        var vol = clip.components[0].properties[0];
        if (!vol) return _err("Volume property not found on clip '" + clip.name + "'.");

        var level = args.levelDb || 0;
        vol.setValue(level, true);
        return _ok({ message: "Audio level on '" + clip.name + "' set to " + level + " dB" });
    } catch (e) { return _err("Failed to set audio level: " + e.message); }
}

// -- Effects -----------------------------------------------------------

function getClipEffects(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var trackType = args.trackType || "video";
        var ti = args.trackIndex || 0;
        var ci = args.clipIndex || 0;

        var track;
        if (trackType === "audio") {
            var numATracks = 0;
            try { numATracks = seq.audioTracks.numTracks; } catch (e1) {}
            if (ti < 0 || ti >= numATracks) return _err("Audio track " + ti + " out of range (0-" + (numATracks - 1) + ")");
            track = seq.audioTracks[ti];
        } else {
            var numVTracks = 0;
            try { numVTracks = seq.videoTracks.numTracks; } catch (e2) {}
            if (ti < 0 || ti >= numVTracks) return _err("Video track " + ti + " out of range (0-" + (numVTracks - 1) + ")");
            track = seq.videoTracks[ti];
        }
        if (!track) return _err(trackType + " track " + ti + " not accessible.");

        var numClips = 0;
        try { numClips = track.clips.numItems; } catch (e3) {}
        if (ci < 0 || ci >= numClips) return _err("Clip " + ci + " out of range on " + trackType + " track " + ti + " (0-" + (numClips - 1) + ")");
        var clip = track.clips[ci];
        if (!clip) return _err("Clip " + ci + " on " + trackType + " track " + ti + " not accessible.");

        var effects = [];
        var numComps = 0;
        try { numComps = clip.components.numItems; } catch (e4) {}
        for (var i = 0; i < numComps; i++) {
            var comp = clip.components[i];
            if (!comp) continue;
            effects.push({ index: i, name: comp.displayName, matchName: comp.matchName });
        }
        return _ok({ clip: clip.name, effects: effects, count: effects.length });
    } catch (e) { return _err("Failed to get clip effects: " + e.message); }
}

// -- Transitions (QE DOM) ----------------------------------------------

function addVideoTransition(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence. Ensure a sequence is open.");

        var ti = args.trackIndex || 0;
        var qeTrack;
        try { qeTrack = qeSeq.getVideoTrackAt(ti); } catch (e2) {}
        if (!qeTrack) return _err("Video track " + ti + " not found in QE DOM.");

        var ci = args.clipIndex || 0;
        var qeClip;
        try { qeClip = qeTrack.getItemAt(ci); } catch (e3) {}
        if (!qeClip) return _err("Clip " + ci + " not found on video track " + ti + ".");

        var transName = args.transitionName || "Cross Dissolve";
        var transition;
        try { transition = qe.project.getVideoTransitionByName(transName); } catch (e4) {}
        if (!transition) return _err("Transition '" + transName + "' not found. Use premiere_get_available_transitions to see available options.");

        qeClip.addTransition(transition, args.applyToEnd !== false, args.duration || 1);
        return _ok({ message: "Transition '" + transName + "' added to clip " + ci + " on track " + ti });
    } catch (e) { return _err("Failed to add transition: " + e.message); }
}

// -- Color -------------------------------------------------------------

function setLumetriProperty(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["property", "value"]);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var ti = args.trackIndex || 0;
        var numVTracks = 0;
        try { numVTracks = seq.videoTracks.numTracks; } catch (e1) {}
        if (ti < 0 || ti >= numVTracks) return _err("Video track " + ti + " out of range (0-" + (numVTracks - 1) + ")");
        var track = seq.videoTracks[ti];
        if (!track) return _err("Video track " + ti + " not accessible.");

        var ci = args.clipIndex || 0;
        var numClips = 0;
        try { numClips = track.clips.numItems; } catch (e2) {}
        if (ci < 0 || ci >= numClips) return _err("Clip " + ci + " out of range on video track " + ti + " (0-" + (numClips - 1) + ")");
        var clip = track.clips[ci];
        if (!clip) return _err("Clip " + ci + " on video track " + ti + " not accessible.");

        var lumetri = null;
        var numComps = 0;
        try { numComps = clip.components.numItems; } catch (e3) {}
        for (var i = 0; i < numComps; i++) {
            var comp = clip.components[i];
            if (comp && comp.matchName === "AE.ADBE Lumetri") { lumetri = comp; break; }
        }
        if (!lumetri) return _err("No Lumetri Color effect on this clip. Use premiere_apply_video_effect to add 'Lumetri Color' first, then set properties.");

        var prop = null;
        var numProps = 0;
        try { numProps = lumetri.properties.numItems; } catch (e4) {}
        for (var j = 0; j < numProps; j++) {
            var p = lumetri.properties[j];
            if (p && p.displayName === args.property) { prop = p; break; }
        }
        if (!prop) {
            // List available properties for a helpful error
            var available = [];
            for (var k = 0; k < numProps; k++) {
                var pk = lumetri.properties[k];
                if (pk) available.push(pk.displayName);
            }
            return _err("Property '" + args.property + "' not found on Lumetri Color. Available: " + available.join(", "));
        }

        prop.setValue(args.value, true);
        return _ok({ message: args.property + " set to " + args.value, clip: clip.name });
    } catch (e) { return _err("Failed to set Lumetri property '" + (args && args.property ? args.property : "unknown") + "': " + e.message); }
}

// -- Motion ------------------------------------------------------------

function setPosition(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var ti = args.trackIndex || 0;
        var numVTracks = 0;
        try { numVTracks = seq.videoTracks.numTracks; } catch (e1) {}
        if (ti < 0 || ti >= numVTracks) return _err("Video track " + ti + " out of range (0-" + (numVTracks - 1) + ")");

        var ci = args.clipIndex || 0;
        var track = seq.videoTracks[ti];
        if (!track) return _err("Video track " + ti + " not accessible.");
        var numClips = 0;
        try { numClips = track.clips.numItems; } catch (e2) {}
        if (ci < 0 || ci >= numClips) return _err("Clip " + ci + " out of range on video track " + ti + " (0-" + (numClips - 1) + ")");
        var clip = track.clips[ci];
        if (!clip) return _err("Clip " + ci + " on video track " + ti + " not accessible.");

        if (!clip.components || clip.components.numItems === 0) return _err("Clip '" + clip.name + "' has no components.");
        var motion = clip.components[0];
        if (!motion) return _err("Motion component not found on clip '" + clip.name + "'.");
        if (!motion.properties || motion.properties.numItems < 2) return _err("Position properties not found on clip '" + clip.name + "'.");

        var xVal = (args.x !== undefined) ? args.x : 0;
        var yVal = (args.y !== undefined) ? args.y : 0;
        motion.properties[0].setValue(xVal, true);
        motion.properties[1].setValue(yVal, true);
        return _ok({ message: "Position set to (" + xVal + ", " + yVal + ")", clip: clip.name });
    } catch (e) { return _err("Failed to set position: " + e.message); }
}

function setScale(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var ti = args.trackIndex || 0;
        var numVTracks = 0;
        try { numVTracks = seq.videoTracks.numTracks; } catch (e1) {}
        if (ti < 0 || ti >= numVTracks) return _err("Video track " + ti + " out of range (0-" + (numVTracks - 1) + ")");

        var ci = args.clipIndex || 0;
        var track = seq.videoTracks[ti];
        if (!track) return _err("Video track " + ti + " not accessible.");
        var numClips = 0;
        try { numClips = track.clips.numItems; } catch (e2) {}
        if (ci < 0 || ci >= numClips) return _err("Clip " + ci + " out of range on video track " + ti + " (0-" + (numClips - 1) + ")");
        var clip = track.clips[ci];
        if (!clip) return _err("Clip " + ci + " on video track " + ti + " not accessible.");

        if (!clip.components || clip.components.numItems === 0) return _err("Clip '" + clip.name + "' has no components.");
        var motion = clip.components[0];
        if (!motion || !motion.properties || motion.properties.numItems < 2) return _err("Scale property not found on clip '" + clip.name + "'.");

        var scaleVal = (args.scale !== undefined) ? args.scale : 100;
        motion.properties[1].setValue(scaleVal, true);
        return _ok({ message: "Scale set to " + scaleVal, clip: clip.name });
    } catch (e) { return _err("Failed to set scale: " + e.message); }
}

function setOpacity(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var ti = args.trackIndex || 0;
        var numVTracks = 0;
        try { numVTracks = seq.videoTracks.numTracks; } catch (e1) {}
        if (ti < 0 || ti >= numVTracks) return _err("Video track " + ti + " out of range (0-" + (numVTracks - 1) + ")");

        var ci = args.clipIndex || 0;
        var track = seq.videoTracks[ti];
        if (!track) return _err("Video track " + ti + " not accessible.");
        var numClips = 0;
        try { numClips = track.clips.numItems; } catch (e2) {}
        if (ci < 0 || ci >= numClips) return _err("Clip " + ci + " out of range on video track " + ti + " (0-" + (numClips - 1) + ")");
        var clip = track.clips[ci];
        if (!clip) return _err("Clip " + ci + " on video track " + ti + " not accessible.");

        if (!clip.components || clip.components.numItems < 2) return _err("Opacity component not found on clip '" + clip.name + "'.");
        var opacityComp = clip.components[1];
        if (!opacityComp || !opacityComp.properties || opacityComp.properties.numItems === 0) {
            return _err("Opacity property not found on clip '" + clip.name + "'.");
        }

        var opacityVal = (args.opacity !== undefined) ? args.opacity : 100;
        opacityComp.properties[0].setValue(opacityVal, true);
        return _ok({ message: "Opacity set to " + opacityVal, clip: clip.name });
    } catch (e) { return _err("Failed to set opacity: " + e.message); }
}

// -- System ------------------------------------------------------------

function getSystemInfo() {
    try {
        return _ok({
            premiere_version: app.version,
            premiere_build: app.build,
            os: $.os,
            engine: $.engineName || "ExtendScript",
            locale: $.locale
        });
    } catch (e) { return _err("Failed to get system info: " + e.message); }
}

// -- Tool Categories (for discoverability) -----------------------------

function getToolCategories() {
    return _ok({
        categories: [
            { tag: "effects", name: "Effects", description: "Apply and manage video/audio effects", tool_count: 66 },
            { tag: "color", name: "Color", description: "Color grading, Lumetri Color, LUTs, color matching", tool_count: 30 },
            { tag: "video_editing", name: "Video Editing", description: "Timeline editing, clip operations, trimming", tool_count: 91 },
            { tag: "audio", name: "Audio", description: "Audio mixing, levels, effects, Essential Sound", tool_count: 62 },
            { tag: "social_media", name: "Social Media Videos", description: "Export for YouTube, Instagram, TikTok, Twitter", tool_count: 15 },
            { tag: "graphics", name: "Social Media Graphics", description: "Titles, lower thirds, MOGRTs, shapes", tool_count: 30 },
            { tag: "get_started", name: "Get Started", description: "Open/create projects, import media, basic setup", tool_count: 23 },
            { tag: "templates", name: "Templates", description: "Sequence, effect, export presets and templates", tool_count: 30 },
            { tag: "collaboration", name: "Collaboration", description: "Review comments, delivery checklists, versioning", tool_count: 30 },
            { tag: "text", name: "Text & Captions", description: "Subtitles, captions, SRT, text overlays", tool_count: 15 },
            { tag: "export", name: "Export & Encoding", description: "Export, render, format conversion, Media Encoder", tool_count: 44 },
            { tag: "ai", name: "AI Workflows", description: "Smart cut, auto-edit, script parsing, rough cut", tool_count: 25 },
            { tag: "motion", name: "Motion & Transform", description: "Position, scale, rotation, crop, PIP, stabilization", tool_count: 30 },
            { tag: "markers", name: "Markers & Metadata", description: "Add/edit markers, clip metadata, XMP data", tool_count: 30 },
            { tag: "multicam", name: "Multicam & Proxy", description: "Multicam editing, proxy workflow", tool_count: 10 },
            { tag: "playback", name: "Playback & Navigation", description: "Play, pause, seek, zoom, timeline navigation", tool_count: 30 },
            { tag: "diagnostics", name: "Diagnostics", description: "Performance monitoring, health checks, debugging", tool_count: 30 },
            { tag: "vr", name: "VR & Immersive", description: "360 video, HDR, stereoscopic 3D", tool_count: 30 },
            { tag: "integration", name: "App Integration", description: "After Effects, Photoshop, Audition, Media Encoder", tool_count: 28 },
            { tag: "batch", name: "Batch Operations", description: "Batch import, export, effects, color, cleanup", tool_count: 30 },
            { tag: "media_browser", name: "Media Browser & Stock", description: "Browse filesystem, find media, search Adobe Stock", tool_count: 15 }
        ]
    });
}

// -- Media Browser -----------------------------------------------------

function browsePath(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["path"]);
        if (args.error) return _err(args.error);

        var folder = new Folder(args.path);
        if (!folder.exists) return _err("Path not found: " + args.path);

        var items = [];
        var files = folder.getFiles();
        for (var i = 0; i < files.length; i++) {
            var f = files[i];
            if (!f) continue;
            items.push({
                name: f.name,
                path: f.fsName,
                isFolder: f instanceof Folder,
                size: f instanceof File ? f.length : 0,
                modified: f.modified ? f.modified.toString() : ""
            });
        }
        return _ok({ path: args.path, items: items, count: items.length });
    } catch (e) { return _err("Failed to browse path '" + (args && args.path ? args.path : "unknown") + "': " + e.message); }
}

function browseMediaFiles(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["path"]);
        if (args.error) return _err(args.error);

        var folder = new Folder(args.path);
        if (!folder.exists) return _err("Path not found: " + args.path);

        var mediaExts = ["mp4","mov","avi","mkv","mxf","m4v","wmv","mpg","mpeg","m2t","mts","wav","mp3","aac","aif","aiff","flac","ogg","png","jpg","jpeg","tif","tiff","psd","ai","bmp","gif","webp","prproj","mogrt"];
        var items = [];
        var allFiles = folder.getFiles();
        for (var i = 0; i < allFiles.length; i++) {
            var f = allFiles[i];
            if (!f) continue;
            if (f instanceof Folder) {
                items.push({ name: f.name, path: f.fsName, isFolder: true, type: "folder" });
            } else {
                var ext = f.name.split(".").pop().toLowerCase();
                var isMedia = false;
                for (var j = 0; j < mediaExts.length; j++) {
                    if (ext === mediaExts[j]) { isMedia = true; break; }
                }
                if (isMedia) {
                    items.push({ name: f.name, path: f.fsName, isFolder: false, type: ext, size: f.length });
                }
            }
        }
        return _ok({ path: args.path, items: items, count: items.length });
    } catch (e) { return _err("Failed to browse media files at '" + (args && args.path ? args.path : "unknown") + "': " + e.message); }
}

function getFavoriteLocations() {
    try {
        var locations = [];
        try { locations.push({ name: "Home", path: Folder.myDocuments.parent.fsName }); } catch (e1) {}
        try { locations.push({ name: "Documents", path: Folder.myDocuments.fsName }); } catch (e2) {}
        try { locations.push({ name: "Desktop", path: Folder.desktop.fsName }); } catch (e3) {}
        try { locations.push({ name: "Movies", path: Folder.myDocuments.parent.fsName + "/Movies" }); } catch (e4) {}
        try { locations.push({ name: "Downloads", path: Folder.myDocuments.parent.fsName + "/Downloads" }); } catch (e5) {}
        try { locations.push({ name: "Premiere Projects", path: Folder.myDocuments.fsName + "/Adobe/Premiere Pro" }); } catch (e6) {}
        return _ok({ locations: locations });
    } catch (e) { return _err("Failed to get favorite locations: " + e.message); }
}

function importFromMediaBrowser(argsJson) {
    try {
        var args = _parseArgs(argsJson);
        if (args.error) return _err(args.error);
        if (!app.project) return _err("No project is open. Open or create a project first.");

        var paths = args.paths || (args.path ? [args.path] : null);
        if (!paths || paths.length === 0) return _err("Missing required parameter: paths (array of file paths)");

        // Validate files exist
        var missing = [];
        for (var i = 0; i < paths.length; i++) {
            var f = new File(paths[i]);
            if (!f.exists) missing.push(paths[i]);
        }
        if (missing.length > 0) return _err("Files not found: " + missing.join(", "));

        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " file(s)", count: paths.length });
    } catch (e) { return _err("Failed to import from media browser: " + e.message); }
}

function getRecentLocations() {
    try {
        var home = "";
        try { home = Folder.myDocuments.parent.fsName; } catch (e1) {}
        var ppFolderPath = "";
        try { ppFolderPath = Folder.myDocuments.fsName + "/Adobe/Premiere Pro"; } catch (e2) {}
        var ppFolder = new Folder(ppFolderPath);
        var versions = [];
        if (ppFolder.exists) {
            var subFolders = ppFolder.getFiles();
            for (var i = 0; i < subFolders.length; i++) {
                if (subFolders[i] instanceof Folder) {
                    versions.push({ name: subFolders[i].name, path: subFolders[i].fsName });
                }
            }
        }
        return _ok({ premiere_projects_folder: ppFolderPath, versions: versions });
    } catch (e) { return _err("Failed to get recent locations: " + e.message); }
}

// -- Timeline Panel Menu Items -----------------------------------------

function setAudioWaveformLabelColor(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        qeSeq.setAudioWaveformsUseLabelColor(args.enabled !== false);
        return _ok({ message: "Audio waveform label color " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch (e) { return _err("Failed to set audio waveform label color: " + e.message); }
}

function setLogarithmicWaveformScaling(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        qeSeq.setLogarithmicWaveformScaling(args.enabled !== false);
        return _ok({ message: "Logarithmic waveform scaling " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch (e) { return _err("Failed to set logarithmic waveform scaling: " + e.message); }
}

function setTimeRulerNumbers(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        try {
            qeSeq.setTimeRulerNumbersEnabled(args.enabled !== false);
        } catch (e2) {
            return _err("Time ruler numbers toggle not available via scripting: " + e2.message);
        }
        return _ok({ message: "Time ruler numbers " + (args.enabled !== false ? "shown" : "hidden") });
    } catch (e) { return _err("Failed to set time ruler numbers: " + e.message); }
}

function setMultiCameraAudioFollowsVideo(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        qeSeq.setMultiCameraAudioFollowsVideo(args.enabled !== false);
        return _ok({ message: "Multi-camera audio follows video " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch (e) { return _err("Failed to set multi-camera audio follows video: " + e.message); }
}

function setMultiCameraSelectionTopPanel(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        qeSeq.setMultiCameraSelectionTopPanel(args.enabled !== false);
        return _ok({ message: "Multi-camera selection top panel " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch (e) { return _err("Failed to set multi-camera selection top panel: " + e.message); }
}

function setMultiCameraFollowsNestSetting(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        qeSeq.setMultiCameraFollowsNestSetting(args.enabled !== false);
        return _ok({ message: "Multi-camera follows nest setting " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch (e) { return _err("Failed to set multi-camera follows nest setting: " + e.message); }
}

function setRectifiedAudioWaveforms(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for active sequence.");

        qeSeq.setRectifiedWaveforms(args.enabled !== false);
        return _ok({ message: "Rectified audio waveforms " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch (e) { return _err("Failed to set rectified audio waveforms: " + e.message); }
}

// -- Frame Capture -----------------------------------------------------

function captureFrameAsBase64(argsJson) {
    try {
        var args = {};
        if (argsJson && argsJson !== "") {
            var parsed = _parseArgs(argsJson);
            if (!parsed.error) args = parsed;
        }
        if (!app.project) return _err("No project is open.");
        var seq = _getActiveSequence();
        if (!seq) return _err("No active sequence.");

        var tempDir = Folder.temp.fsName;
        var tempFile = tempDir + "/mcp_frame_" + (new Date()).getTime() + ".png";

        app.enableQE();
        var qeSeq;
        try { qeSeq = qe.project.getActiveSequence(); } catch (e1) {}
        if (!qeSeq) return _err("QE DOM not available for frame capture.");

        var pos = seq.getPlayerPosition();
        if (!pos) return _err("Could not read playhead position for frame capture.");

        qeSeq.exportFramePNG(pos.ticks, tempFile);

        var file = new File(tempFile);
        if (!file.exists) return _err("Failed to capture frame: exported PNG file not created.");

        file.open("r");
        file.encoding = "BINARY";
        var binary = file.read();
        file.close();

        var base64 = _binaryToBase64(binary);

        // Clean up temp file
        try { file.remove(); } catch (e2) {}

        return _ok({
            image_base64: base64,
            format: "png",
            width: seq.frameSizeHorizontal,
            height: seq.frameSizeVertical,
            timecode: pos.seconds
        });
    } catch (e) { return _err("Failed to capture frame: " + e.message); }
}

// Base64 encoder for binary data
function _binaryToBase64(data) {
    var chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    var result = "";
    var i = 0;
    while (i < data.length) {
        var a = data.charCodeAt(i++) || 0;
        var b = data.charCodeAt(i++) || 0;
        var c = data.charCodeAt(i++) || 0;
        result += chars.charAt(a >> 2);
        result += chars.charAt(((a & 3) << 4) | (b >> 4));
        result += chars.charAt(((b & 15) << 2) | (c >> 6));
        result += chars.charAt(c & 63);
    }
    // Add padding
    var pad = data.length % 3;
    if (pad === 1) result = result.slice(0, -2) + "==";
    else if (pad === 2) result = result.slice(0, -1) + "=";
    return result;
}

// -- Secure Script Execution -------------------------------------------

function executeSecureScript(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["script"]);
        if (args.error) return _err(args.error);

        var script = args.script;
        var validate = args.validate !== false;

        if (validate) {
            var blocked = ["System.callSystem", "$.system", "app.quit", "File.remove",
                           "Folder.remove", "$.sleep(9", "while(true)", "for(;;)"];
            for (var i = 0; i < blocked.length; i++) {
                if (script.indexOf(blocked[i]) >= 0) {
                    return _err("Blocked: script contains forbidden operation: " + blocked[i]);
                }
            }
        }

        var result = eval(script);
        return _ok({ result: String(result) });
    } catch (e) { return _err("Failed to execute script: " + e.message); }
}

function executeQEScript(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["script"]);
        if (args.error) return _err(args.error);

        app.enableQE();
        var result = eval(args.script);
        return _ok({ result: String(result) });
    } catch (e) { return _err("Failed to execute QE script: " + e.message); }
}

// -- Panel Docking via macOS Accessibility (AppleScript) ---------------

function simulateMenuClick(argsJson) {
    try {
        var args = _parseArgs(argsJson, ["menu_path"]);
        if (args.error) return _err(args.error);

        var menuPath = args.menu_path;
        var parts = menuPath.split("/");
        if (parts.length < 2 || parts.length > 3) {
            return _err("menu_path must have 2 or 3 segments separated by '/' (e.g., 'Window/Extensions/PremierPro MCP Bridge')");
        }

        var script = 'tell application "System Events"\n';
        script += '  tell process "Adobe Premiere Pro 2026"\n';
        script += '    set frontmost to true\n';

        if (parts.length === 2) {
            script += '    click menu item "' + parts[1] + '" of menu 1 of menu bar item "' + parts[0] + '" of menu bar 1\n';
        } else if (parts.length === 3) {
            script += '    click menu item "' + parts[2] + '" of menu 1 of menu item "' + parts[1] + '" of menu 1 of menu bar item "' + parts[0] + '" of menu bar 1\n';
        }

        script += '  end tell\n';
        script += 'end tell';

        app.doScript(script, ScriptLanguage.APPLESCRIPT);
        return _ok({ message: "Menu clicked: " + menuPath });
    } catch (e) { return _err("Failed to simulate menu click '" + (args && args.menu_path ? args.menu_path : "unknown") + "': " + e.message); }
}
