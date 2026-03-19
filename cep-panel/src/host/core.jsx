// Core ExtendScript functions for PremierPro MCP Bridge
// This minimal file loads fast and provides essential functions.

function _ok(data) { return JSON.stringify({ success: true, data: data }); }
function _err(message) { return JSON.stringify({ success: false, error: String(message) }); }

// JSON polyfill for older ExtendScript
if (typeof JSON === "undefined") {
    JSON = {
        stringify: function(obj) {
            if (obj === null) return "null";
            if (typeof obj === "string") return '"' + obj.replace(/\\/g, "\\\\").replace(/"/g, '\\"').replace(/\n/g, "\\n") + '"';
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

// ── Project ───────────────────────────────────────────────────────────

function ping() {
    try {
        var ver = "unknown";
        try { ver = app.version; } catch(e) {}
        var projOpen = false;
        var projName = "";
        try { projOpen = app.project && app.project.name ? true : false; projName = app.project.name; } catch(e) {}
        return _ok({ premiere_running: true, premiere_version: ver, project_open: projOpen, project_name: projName });
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
        return _ok({ name: app.project.name, path: app.project.path, sequences: seqs, sequence_count: app.project.sequences.numItems });
    } catch(e) { return _err(e.message); }
}

function getProjectState() { return getProjectInfo(); }

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
        return _ok({ message: "Project opened", path: args.path, name: app.project.name });
    } catch(e) { return _err(e.message); }
}

function saveProject() {
    try { app.project.save(); return _ok({ message: "Project saved" }); }
    catch(e) { return _err(e.message); }
}

function saveProjectAs(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        app.project.saveAs(args.path);
        return _ok({ message: "Project saved as", path: args.path });
    } catch(e) { return _err(e.message); }
}

function closeProject(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.project.closeDocument(args.save_first !== false, true);
        return _ok({ message: "Project closed" });
    } catch(e) { return _err(e.message); }
}

// ── Sequences ─────────────────────────────────────────────────────────

function createSequence(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var name = args.name || "New Sequence";
        app.project.createNewSequence(name, name);
        var seq = app.project.activeSequence;
        return _ok({ name: seq.name, id: seq.sequenceID, width: seq.frameSizeHorizontal, height: seq.frameSizeVertical });
    } catch(e) { return _err(e.message); }
}

function getActiveSequence() {
    try {
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        return _ok({ name: seq.name, id: seq.sequenceID, width: seq.frameSizeHorizontal, height: seq.frameSizeVertical, end: seq.end });
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

// ── Media Import ──────────────────────────────────────────────────────

function importFiles(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var paths = args.paths || args.filePaths || [args.path || args.filePath];
        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " files", count: paths.length });
    } catch(e) { return _err(e.message); }
}

function importFolder(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var folder = new Folder(args.path || args.folderPath);
        if (!folder.exists) return _err("Folder not found: " + args.path);
        var files = folder.getFiles();
        var paths = [];
        for (var i = 0; i < files.length; i++) {
            if (files[i] instanceof File) paths.push(files[i].fsName);
        }
        if (paths.length === 0) return _err("No files found in folder");
        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " files from folder", count: paths.length });
    } catch(e) { return _err(e.message); }
}

// ── Clips ─────────────────────────────────────────────────────────────

function getTimelineState(argsJson) {
    try {
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var tracks = [];
        for (var t = 0; t < seq.videoTracks.numTracks; t++) {
            var track = seq.videoTracks[t];
            var clips = [];
            for (var c = 0; c < track.clips.numItems; c++) {
                var clip = track.clips[c];
                clips.push({ index: c, name: clip.name, start: clip.start.seconds, end: clip.end.seconds, duration: clip.duration.seconds });
            }
            tracks.push({ index: t, name: track.name, type: "video", clips: clips });
        }
        for (var t = 0; t < seq.audioTracks.numTracks; t++) {
            var track = seq.audioTracks[t];
            var clips = [];
            for (var c = 0; c < track.clips.numItems; c++) {
                var clip = track.clips[c];
                clips.push({ index: c, name: clip.name, start: clip.start.seconds, end: clip.end.seconds, duration: clip.duration.seconds });
            }
            tracks.push({ index: t, name: track.name, type: "audio", clips: clips });
        }
        return _ok({ sequence: seq.name, tracks: tracks });
    } catch(e) { return _err(e.message); }
}

function insertClip(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var item = app.project.rootItem.children[args.projectItemIndex || 0];
        if (!item) return _err("Project item not found");
        seq.insertClip(item, args.time || 0, args.videoTrackIndex || 0, args.audioTrackIndex || 0);
        return _ok({ message: "Clip inserted" });
    } catch(e) { return _err(e.message); }
}

function placeClip(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var idx = args.projectItemIndex || 0;
        var item = app.project.rootItem.children[idx];
        if (!item) return _err("Project item " + idx + " not found");
        var track = seq.videoTracks[args.trackIndex || 0];
        if (!track) return _err("Video track not found");
        track.overwriteClip(item, args.startTime || 0);
        return _ok({ message: "Clip placed on track" });
    } catch(e) { return _err(e.message); }
}

// ── Export ─────────────────────────────────────────────────────────────

function exportSequence(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        seq.exportAsMediaDirect(args.outputPath, args.presetPath || "", 0);
        return _ok({ message: "Export started", output: args.outputPath });
    } catch(e) { return _err(e.message); }
}

// ── Bins ──────────────────────────────────────────────────────────────

function createBin(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var bin = app.project.rootItem.createBin(args.name || "New Bin");
        return _ok({ message: "Bin created", name: args.name });
    } catch(e) { return _err(e.message); }
}

function getProjectItems(argsJson) {
    try {
        var root = app.project.rootItem;
        var items = [];
        for (var i = 0; i < root.children.numItems; i++) {
            var item = root.children[i];
            items.push({ index: i, name: item.name, type: item.type, path: item.treePath });
        }
        return _ok({ items: items, count: items.length });
    } catch(e) { return _err(e.message); }
}

// ── Markers ───────────────────────────────────────────────────────────

function addSequenceMarker(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var marker = seq.markers.createMarker(args.time || 0);
        if (args.name) marker.name = args.name;
        if (args.comment) marker.comments = args.comment;
        return _ok({ message: "Marker added at " + (args.time || 0) + "s" });
    } catch(e) { return _err(e.message); }
}

// ── Playback ──────────────────────────────────────────────────────────

function getPlayheadPosition() {
    try {
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var pos = seq.getPlayerPosition();
        return _ok({ seconds: pos.seconds, ticks: pos.ticks });
    } catch(e) { return _err(e.message); }
}

function setPlayheadPosition(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        seq.setPlayerPosition(args.seconds + "254016000000");
        return _ok({ message: "Playhead moved to " + args.seconds + "s" });
    } catch(e) { return _err(e.message); }
}

// ── Audio ─────────────────────────────────────────────────────────────

function setAudioLevel(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        var clip = seq.audioTracks[args.trackIndex || 0].clips[args.clipIndex || 0];
        var vol = clip.components[0].properties[0];
        vol.setValue(args.levelDb || 0, true);
        return _ok({ message: "Audio level set to " + args.levelDb + " dB" });
    } catch(e) { return _err(e.message); }
}

// ── Effects ───────────────────────────────────────────────────────────

function getClipEffects(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        var track = args.trackType === "audio" ? seq.audioTracks[args.trackIndex || 0] : seq.videoTracks[args.trackIndex || 0];
        var clip = track.clips[args.clipIndex || 0];
        var effects = [];
        for (var i = 0; i < clip.components.numItems; i++) {
            var comp = clip.components[i];
            effects.push({ index: i, name: comp.displayName, matchName: comp.matchName });
        }
        return _ok({ effects: effects, count: effects.length });
    } catch(e) { return _err(e.message); }
}

// ── Transitions (QE DOM) ──────────────────────────────────────────────

function addVideoTransition(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        var qeTrack = qeSeq.getVideoTrackAt(args.trackIndex || 0);
        var qeClip = qeTrack.getItemAt(args.clipIndex || 0);
        qeClip.addTransition(qe.project.getVideoTransitionByName(args.transitionName || "Cross Dissolve"), args.applyToEnd !== false, args.duration || 1);
        return _ok({ message: "Transition added: " + (args.transitionName || "Cross Dissolve") });
    } catch(e) { return _err(e.message); }
}

// ── Color ─────────────────────────────────────────────────────────────

function setLumetriProperty(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        var clip = seq.videoTracks[args.trackIndex || 0].clips[args.clipIndex || 0];
        var lumetri = null;
        for (var i = 0; i < clip.components.numItems; i++) {
            if (clip.components[i].matchName === "AE.ADBE Lumetri") { lumetri = clip.components[i]; break; }
        }
        if (!lumetri) return _err("No Lumetri Color effect on clip. Apply it first.");
        var prop = null;
        for (var j = 0; j < lumetri.properties.numItems; j++) {
            if (lumetri.properties[j].displayName === args.property) { prop = lumetri.properties[j]; break; }
        }
        if (!prop) return _err("Property not found: " + args.property);
        prop.setValue(args.value, true);
        return _ok({ message: args.property + " set to " + args.value });
    } catch(e) { return _err(e.message); }
}

// ── Motion ────────────────────────────────────────────────────────────

function setPosition(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        var clip = seq.videoTracks[args.trackIndex || 0].clips[args.clipIndex || 0];
        var motion = clip.components[0]; // Motion component
        motion.properties[0].setValue(args.x || 0, true); // Position X
        motion.properties[1].setValue(args.y || 0, true); // Position Y
        return _ok({ message: "Position set" });
    } catch(e) { return _err(e.message); }
}

function setScale(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        var clip = seq.videoTracks[args.trackIndex || 0].clips[args.clipIndex || 0];
        clip.components[0].properties[1].setValue(args.scale || 100, true);
        return _ok({ message: "Scale set to " + args.scale });
    } catch(e) { return _err(e.message); }
}

function setOpacity(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var seq = app.project.activeSequence;
        var clip = seq.videoTracks[args.trackIndex || 0].clips[args.clipIndex || 0];
        clip.components[1].properties[0].setValue(args.opacity || 100, true);
        return _ok({ message: "Opacity set to " + args.opacity });
    } catch(e) { return _err(e.message); }
}

// ── System ────────────────────────────────────────────────────────────

function getSystemInfo() {
    try {
        return _ok({
            premiere_version: app.version,
            premiere_build: app.build,
            os: $.os,
            engine: $.engineName || "ExtendScript",
            locale: $.locale
        });
    } catch(e) { return _err(e.message); }
}

// ── Tool Categories (for discoverability) ─────────────────────────────

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

// ── Media Browser ─────────────────────────────────────────────────────

function browsePath(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var folder = new Folder(args.path);
        if (!folder.exists) return _err("Path not found: " + args.path);
        var items = [];
        var files = folder.getFiles();
        for (var i = 0; i < files.length; i++) {
            var f = files[i];
            items.push({
                name: f.name,
                path: f.fsName,
                isFolder: f instanceof Folder,
                size: f instanceof File ? f.length : 0,
                modified: f.modified ? f.modified.toString() : ""
            });
        }
        return _ok({ path: args.path, items: items, count: items.length });
    } catch(e) { return _err(e.message); }
}

function browseMediaFiles(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var folder = new Folder(args.path);
        if (!folder.exists) return _err("Path not found: " + args.path);
        var mediaExts = ["mp4","mov","avi","mkv","mxf","m4v","wmv","mpg","mpeg","m2t","mts","wav","mp3","aac","aif","aiff","flac","ogg","png","jpg","jpeg","tif","tiff","psd","ai","bmp","gif","webp","prproj","mogrt"];
        var items = [];
        var allFiles = folder.getFiles();
        for (var i = 0; i < allFiles.length; i++) {
            var f = allFiles[i];
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
    } catch(e) { return _err(e.message); }
}

function getFavoriteLocations() {
    try {
        var locations = [
            { name: "Home", path: Folder.myDocuments.parent.fsName },
            { name: "Documents", path: Folder.myDocuments.fsName },
            { name: "Desktop", path: Folder.desktop.fsName },
            { name: "Movies", path: Folder.myDocuments.parent.fsName + "/Movies" },
            { name: "Downloads", path: Folder.myDocuments.parent.fsName + "/Downloads" },
            { name: "Premiere Projects", path: Folder.myDocuments.fsName + "/Adobe/Premiere Pro" }
        ];
        return _ok({ locations: locations });
    } catch(e) { return _err(e.message); }
}

function importFromMediaBrowser(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var paths = args.paths || [args.path];
        app.project.importFiles(paths, true, app.project.getInsertionBin(), false);
        return _ok({ message: "Imported " + paths.length + " files", count: paths.length });
    } catch(e) { return _err(e.message); }
}

function getRecentLocations() {
    try {
        // Check default Premiere Pro project locations
        var home = Folder.myDocuments.parent.fsName;
        var ppFolder = new Folder(Folder.myDocuments.fsName + "/Adobe/Premiere Pro");
        var versions = [];
        if (ppFolder.exists) {
            var subFolders = ppFolder.getFiles();
            for (var i = 0; i < subFolders.length; i++) {
                if (subFolders[i] instanceof Folder) {
                    versions.push({ name: subFolders[i].name, path: subFolders[i].fsName });
                }
            }
        }
        return _ok({ premiere_projects_folder: ppFolder.fsName, versions: versions });
    } catch(e) { return _err(e.message); }
}

// ── Timeline Panel Menu Items ─────────────────────────────────────────

function setAudioWaveformLabelColor(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        // Toggle via QE DOM or menu command simulation
        app.enableQE();
        qe.project.getActiveSequence().setAudioWaveformsUseLabelColor(args.enabled !== false);
        return _ok({ message: "Audio waveform label color " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch(e) { return _err(e.message); }
}

function setLogarithmicWaveformScaling(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.enableQE();
        qe.project.getActiveSequence().setLogarithmicWaveformScaling(args.enabled !== false);
        return _ok({ message: "Logarithmic waveform scaling " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch(e) { return _err(e.message); }
}

function setTimeRulerNumbers(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.enableQE();
        // Try QE DOM or app properties
        try { qe.project.getActiveSequence().setTimeRulerNumbersEnabled(args.enabled !== false); }
        catch(e2) { return _err("Time ruler numbers toggle not available via scripting: " + e2.message); }
        return _ok({ message: "Time ruler numbers " + (args.enabled !== false ? "shown" : "hidden") });
    } catch(e) { return _err(e.message); }
}

function setMultiCameraAudioFollowsVideo(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        qeSeq.setMultiCameraAudioFollowsVideo(args.enabled !== false);
        return _ok({ message: "Multi-camera audio follows video " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch(e) { return _err(e.message); }
}

function setMultiCameraSelectionTopPanel(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        qeSeq.setMultiCameraSelectionTopPanel(args.enabled !== false);
        return _ok({ message: "Multi-camera selection top panel " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch(e) { return _err(e.message); }
}

function setMultiCameraFollowsNestSetting(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        qeSeq.setMultiCameraFollowsNestSetting(args.enabled !== false);
        return _ok({ message: "Multi-camera follows nest setting " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch(e) { return _err(e.message); }
}

function setRectifiedAudioWaveforms(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        app.enableQE();
        qe.project.getActiveSequence().setRectifiedWaveforms(args.enabled !== false);
        return _ok({ message: "Rectified audio waveforms " + (args.enabled !== false ? "enabled" : "disabled") });
    } catch(e) { return _err(e.message); }
}

// ── Panel Docking via macOS Accessibility (AppleScript) ───────────────
// These use System Events to simulate clicking Premiere Pro's UI menus
// since Adobe doesn't expose panel docking via ExtendScript.

// ── Frame Capture ─────────────────────────────────────────────────────

function captureFrameAsBase64(argsJson) {
    try {
        var args = argsJson ? JSON.parse(argsJson) : {};
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");

        // Export current frame to temp file
        var tempDir = Folder.temp.fsName;
        var tempFile = tempDir + "/mcp_frame_" + Date.now() + ".png";

        // Use QE DOM to export frame
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        qeSeq.exportFramePNG(seq.getPlayerPosition().ticks, tempFile);

        // Read file and convert to base64
        var file = new File(tempFile);
        if (!file.exists) return _err("Failed to capture frame");

        file.open("r");
        file.encoding = "BINARY";
        var binary = file.read();
        file.close();

        // Convert to base64 using ExtendScript binary encoder
        var base64 = _binaryToBase64(binary);

        // Clean up temp file
        file.remove();

        return _ok({
            image_base64: base64,
            format: "png",
            width: seq.frameSizeHorizontal,
            height: seq.frameSizeVertical,
            timecode: seq.getPlayerPosition().seconds
        });
    } catch(e) { return _err(e.message); }
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

// ── Secure Script Execution ───────────────────────────────────────────

function executeSecureScript(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var script = args.script || "";
        var validate = args.validate !== false;

        if (validate) {
            // Security validation - block dangerous operations
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
    } catch(e) { return _err(e.message); }
}

function executeQEScript(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        app.enableQE();
        var result = eval(args.script);
        return _ok({ result: String(result) });
    } catch(e) { return _err(e.message); }
}

// ── Panel Docking via macOS Accessibility (AppleScript) ───────────────
// These use System Events to simulate clicking Premiere Pro's UI menus
// since Adobe doesn't expose panel docking via ExtendScript.

function simulateMenuClick(argsJson) {
    try {
        var args = JSON.parse(argsJson);
        var menuPath = args.menu_path; // e.g., "Window/Extensions/PremierPro MCP Bridge"
        if (!menuPath) return _err("menu_path is required");

        // Build AppleScript to click the menu item
        var parts = menuPath.split("/");
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
    } catch(e) { return _err(e.message); }
}
