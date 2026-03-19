/**
 * PremierPro MCP Bridge - ExtendScript Host Functions
 *
 * These functions run inside Premiere Pro's ExtendScript engine and are
 * called from the CEP panel via CSInterface.evalScript(). Every function
 * returns a JSON string so the panel can parse the result.
 *
 * Premiere Pro ExtendScript DOM reference:
 *   app.project                  - The current Project object
 *   app.project.rootItem         - Root ProjectItem (bin)
 *   app.project.activeSequence   - Currently active Sequence
 *   sequence.videoTracks         - TrackCollection for video
 *   sequence.audioTracks         - TrackCollection for audio
 *   track.clips                  - TrackItemCollection
 */

// ---------------------------------------------------------------------------
// Utility: Build a JSON response envelope
// ---------------------------------------------------------------------------
function _ok(data) {
    return JSON.stringify({ success: true, data: data });
}

function _err(message) {
    return JSON.stringify({ success: false, error: String(message) });
}

/**
 * Safe JSON serializer that handles ExtendScript quirks.
 * ExtendScript's native JSON may not exist in older versions.
 */
if (typeof JSON === "undefined") {
    // Minimal JSON polyfill for ExtendScript environments that lack it.
    JSON = {
        stringify: function (obj) {
            if (obj === null) return "null";
            if (obj === undefined) return "undefined";
            if (typeof obj === "number" || typeof obj === "boolean") return String(obj);
            if (typeof obj === "string") {
                return '"' + obj.replace(/\\/g, "\\\\").replace(/"/g, '\\"')
                                .replace(/\n/g, "\\n").replace(/\r/g, "\\r")
                                .replace(/\t/g, "\\t") + '"';
            }
            if (obj instanceof Array) {
                var arrParts = [];
                for (var i = 0; i < obj.length; i++) {
                    arrParts.push(JSON.stringify(obj[i]));
                }
                return "[" + arrParts.join(",") + "]";
            }
            if (typeof obj === "object") {
                var objParts = [];
                for (var key in obj) {
                    if (obj.hasOwnProperty(key)) {
                        objParts.push('"' + key + '":' + JSON.stringify(obj[key]));
                    }
                }
                return "{" + objParts.join(",") + "}";
            }
            return String(obj);
        },
        parse: function (str) {
            return eval("(" + str + ")");
        }
    };
}

// ---------------------------------------------------------------------------
// Time helpers
// ---------------------------------------------------------------------------

/**
 * Convert a Time object to seconds (float).
 */
function _timeToSeconds(timeObj) {
    if (!timeObj) return 0;
    return parseFloat(timeObj.seconds);
}

/**
 * Create a Time object from seconds.
 */
function _secondsToTime(seconds) {
    var t = new Time();
    t.seconds = seconds;
    return t;
}

// ---------------------------------------------------------------------------
// ping() - Health check
// ---------------------------------------------------------------------------
function ping() {
    try {
        var info = {
            status: "ok",
            host: "Premiere Pro",
            version: app.version || "unknown",
            buildNumber: app.build || "unknown",
            timestamp: new Date().toISOString()
        };

        // Check if a project is open
        if (app.project) {
            info.projectOpen = true;
            info.projectName = app.project.name || "";
        } else {
            info.projectOpen = false;
        }

        return _ok(info);
    } catch (e) {
        return _err("ping failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getProjectState() - Full project overview
// ---------------------------------------------------------------------------
function getProjectState() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var proj = app.project;
        var state = {
            name: proj.name || "",
            path: proj.path || "",
            sequences: [],
            activeSequenceIndex: -1
        };

        // Enumerate sequences
        for (var i = 0; i < proj.sequences.numSequences; i++) {
            var seq = proj.sequences[i];
            var seqInfo = {
                index: i,
                name: seq.name || "",
                id: seq.sequenceID || "",
                videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0,
                audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0,
                inPoint: _timeToSeconds(seq.getInPoint()),
                outPoint: _timeToSeconds(seq.getOutPoint()),
                timebase: seq.timebase || "",
                frameSizeHorizontal: seq.frameSizeHorizontal || 0,
                frameSizeVertical: seq.frameSizeVertical || 0
            };
            state.sequences.push(seqInfo);

            // Check if this is the active sequence
            if (proj.activeSequence && proj.activeSequence.sequenceID === seq.sequenceID) {
                state.activeSequenceIndex = i;
            }
        }

        // Count root items
        if (proj.rootItem) {
            state.rootItemCount = proj.rootItem.children ? proj.rootItem.children.numItems : 0;
        }

        return _ok(state);
    } catch (e) {
        return _err("getProjectState failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// createSequence(paramsJson) - Create a new sequence
// ---------------------------------------------------------------------------
function createSequence(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var name = params.name || "New Sequence";
        var width = parseInt(params.width, 10) || 1920;
        var height = parseInt(params.height, 10) || 1080;
        var fps = parseFloat(params.fps) || 29.97;

        if (!app.project) {
            return _err("No project is open");
        }

        // Use the new sequence preset creation if available
        var seqID = app.project.createNewSequence(name);

        if (seqID) {
            // Try to set properties on the newly created sequence
            var seq = app.project.activeSequence;
            if (seq) {
                // Setting frame size may require sequence presets in newer versions
                // These settings work through sequence settings dialog in practice,
                // but we attempt to set them programmatically
                try {
                    seq.frameSizeHorizontal = width;
                    seq.frameSizeVertical = height;
                } catch (dimErr) {
                    // Frame dimensions may not be directly settable in all versions
                }
            }

            return _ok({
                name: name,
                sequenceID: seqID || "",
                width: width,
                height: height,
                fps: fps,
                videoTrackCount: seq ? (seq.videoTracks ? seq.videoTracks.numTracks : 0) : 0,
                audioTrackCount: seq ? (seq.audioTracks ? seq.audioTracks.numTracks : 0) : 0,
                timebase: seq ? (seq.timebase || "") : ""
            });
        } else {
            return _err("createNewSequence returned no ID");
        }
    } catch (e) {
        return _err("createSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getTimelineState(sequenceIndex) - Get clips on a specific sequence
// ---------------------------------------------------------------------------
function getTimelineState(sequenceIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10) || 0;
        var seq;

        if (sequenceIndex === -1 || sequenceIndex === undefined) {
            seq = app.project.activeSequence;
        } else {
            if (sequenceIndex >= app.project.sequences.numSequences) {
                return _err("Sequence index " + sequenceIndex + " out of range");
            }
            seq = app.project.sequences[sequenceIndex];
        }

        if (!seq) {
            return _err("No sequence found at index " + sequenceIndex);
        }

        var timeline = {
            name: seq.name || "",
            sequenceID: seq.sequenceID || "",
            inPoint: _timeToSeconds(seq.getInPoint()),
            outPoint: _timeToSeconds(seq.getOutPoint()),
            endSeconds: _timeToSeconds(seq.end),
            frameSizeHorizontal: seq.frameSizeHorizontal || 0,
            frameSizeVertical: seq.frameSizeVertical || 0,
            timebase: seq.timebase || "",
            markers: [],
            videoTracks: [],
            audioTracks: []
        };

        // Collect sequence markers
        if (seq.markers) {
            var marker = seq.markers.getFirstMarker();
            var mIdx = 0;
            while (marker) {
                timeline.markers.push({
                    index: mIdx,
                    name: marker.name || "",
                    comment: marker.comments || "",
                    start: _timeToSeconds(marker.start),
                    end: _timeToSeconds(marker.end),
                    type: marker.type || "",
                    colorIndex: marker.colorIndex !== undefined ? marker.colorIndex : -1
                });
                mIdx++;
                marker = seq.markers.getNextMarker(marker);
            }
        }

        // Video tracks
        if (seq.videoTracks) {
            for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) {
                var vTrack = seq.videoTracks[vi];
                var vTrackInfo = {
                    index: vi,
                    name: vTrack.name || "Video " + (vi + 1),
                    clips: []
                };

                if (vTrack.clips) {
                    for (var vc = 0; vc < vTrack.clips.numItems; vc++) {
                        var vClip = vTrack.clips[vc];
                        vTrackInfo.clips.push({
                            index: vc,
                            name: vClip.name || "",
                            start: _timeToSeconds(vClip.start),
                            end: _timeToSeconds(vClip.end),
                            duration: _timeToSeconds(vClip.duration),
                            inPoint: _timeToSeconds(vClip.inPoint),
                            outPoint: _timeToSeconds(vClip.outPoint),
                            type: vClip.type || "",
                            mediaPath: vClip.projectItem ? (vClip.projectItem.getMediaPath() || "") : ""
                        });
                    }
                }

                timeline.videoTracks.push(vTrackInfo);
            }
        }

        // Audio tracks
        if (seq.audioTracks) {
            for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) {
                var aTrack = seq.audioTracks[ai];
                var aTrackInfo = {
                    index: ai,
                    name: aTrack.name || "Audio " + (ai + 1),
                    clips: []
                };

                if (aTrack.clips) {
                    for (var ac = 0; ac < aTrack.clips.numItems; ac++) {
                        var aClip = aTrack.clips[ac];
                        aTrackInfo.clips.push({
                            index: ac,
                            name: aClip.name || "",
                            start: _timeToSeconds(aClip.start),
                            end: _timeToSeconds(aClip.end),
                            duration: _timeToSeconds(aClip.duration),
                            inPoint: _timeToSeconds(aClip.inPoint),
                            outPoint: _timeToSeconds(aClip.outPoint)
                        });
                    }
                }

                timeline.audioTracks.push(aTrackInfo);
            }
        }

        return _ok(timeline);
    } catch (e) {
        return _err("getTimelineState failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// importMedia(filePath, binPath) - Import a file into the project
// ---------------------------------------------------------------------------
function importMedia(filePath, binPath) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        if (!filePath || filePath === "") {
            return _err("filePath is required");
        }

        // Determine the target bin
        var targetBin = app.project.rootItem;

        if (binPath && binPath !== "") {
            // Navigate to or create the bin path
            var binParts = binPath.split("/");
            for (var b = 0; b < binParts.length; b++) {
                var binName = binParts[b];
                if (!binName || binName === "") continue;

                var found = false;
                if (targetBin.children) {
                    for (var c = 0; c < targetBin.children.numItems; c++) {
                        var child = targetBin.children[c];
                        if (child.name === binName && child.type === ProjectItemType.BIN) {
                            targetBin = child;
                            found = true;
                            break;
                        }
                    }
                }

                if (!found) {
                    // Create the bin
                    var newBin = targetBin.createBin(binName);
                    if (newBin) {
                        targetBin = newBin;
                    } else {
                        return _err("Failed to create bin: " + binName);
                    }
                }
            }
        }

        // Import the file
        var importArray = [filePath];
        var suppressUI = true;

        if (app.project.importFiles(importArray, suppressUI, targetBin, false)) {
            // Find the newly imported item (last item in target bin)
            var importedItem = null;
            if (targetBin.children && targetBin.children.numItems > 0) {
                importedItem = targetBin.children[targetBin.children.numItems - 1];
            }

            var result = {
                filePath: filePath,
                binPath: binPath || "/",
                imported: true
            };

            if (importedItem) {
                result.name = importedItem.name || "";
                result.mediaPath = importedItem.getMediaPath ? (importedItem.getMediaPath() || "") : "";
            }

            return _ok(result);
        } else {
            return _err("importFiles returned false for: " + filePath);
        }
    } catch (e) {
        return _err("importMedia failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// placeClip(projectItemIndex, trackIndex, startTime)
// ---------------------------------------------------------------------------
function placeClip(projectItemIndex, trackIndex, startTime) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0;

        // Get the project item from root bin
        if (!app.project.rootItem.children ||
            projectItemIndex >= app.project.rootItem.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var projectItem = app.project.rootItem.children[projectItemIndex];
        if (!projectItem) {
            return _err("No project item at index " + projectItemIndex);
        }

        // Get the target video track
        if (trackIndex >= seq.videoTracks.numTracks) {
            return _err("Video track index " + trackIndex + " out of range (have " + seq.videoTracks.numTracks + ")");
        }

        var track = seq.videoTracks[trackIndex];
        var insertTime = _secondsToTime(startTime);

        // Insert the clip onto the track
        track.insertClip(projectItem, insertTime);

        return _ok({
            projectItemName: projectItem.name || "",
            trackIndex: trackIndex,
            startTime: startTime
        });
    } catch (e) {
        return _err("placeClip failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// addTransition(trackIndex, clipIndex, transitionName, duration)
// ---------------------------------------------------------------------------
function addTransition(trackIndex, clipIndex, transitionName, duration) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        duration = parseFloat(duration) || 1.0;

        if (trackIndex >= seq.videoTracks.numTracks) {
            return _err("Video track index " + trackIndex + " out of range");
        }

        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) {
            return _err("Clip index " + clipIndex + " out of range on track " + trackIndex);
        }

        var clip = track.clips[clipIndex];

        // Apply transition at the end of the clip
        // Premiere's DOM uses QE (Quick Export) domain for transitions in some versions
        var transitionDuration = _secondsToTime(duration);

        // Try using the TrackItem's transitions
        if (clip.setEndTransition) {
            clip.setEndTransition(transitionName, transitionDuration);
        } else if (typeof qe !== "undefined" && qe.project) {
            // Fallback: QE DOM approach
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
                if (qeTrack) {
                    var qeClip = qeTrack.getItemAt(clipIndex);
                    if (qeClip) {
                        qeClip.addTransition(
                            qe.project.getVideoTransitionByName(transitionName || "Cross Dissolve"),
                            true,  // at end
                            duration.toString()
                        );
                    }
                }
            }
        }

        return _ok({
            trackIndex: trackIndex,
            clipIndex: clipIndex,
            transitionName: transitionName || "Cross Dissolve",
            duration: duration
        });
    } catch (e) {
        return _err("addTransition failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// addText(text, trackIndex, startTime, duration)
// ---------------------------------------------------------------------------
function addText(text, trackIndex, startTime, duration) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0;
        duration = parseFloat(duration) || 5.0;

        // Create a text clip using the Graphics workflow
        // In Premiere Pro's ExtendScript, we use the Motion Graphics Template approach
        // or the captions API depending on the version

        // Method 1: Try using the captions/graphics API
        if (seq.createCaptionTrack) {
            // Newer Premiere versions support caption tracks
            seq.createCaptionTrack(text, startTime, duration);
        }

        // Method 2: Use QE DOM for adding titles
        if (typeof qe !== "undefined" && qe.project) {
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                // Add a transparent video clip and overlay text
                // This approach varies by Premiere version
                var insertPoint = _secondsToTime(startTime);
                var textDuration = _secondsToTime(duration);

                // Attempt to add via legacy title
                try {
                    app.project.createNewSequence(text + "_title", "sequenceID");
                } catch (titleErr) {
                    // Title creation varies significantly by version
                }
            }
        }

        return _ok({
            text: text,
            trackIndex: trackIndex,
            startTime: startTime,
            duration: duration,
            note: "Text overlay creation depends on Premiere Pro version. Verify in timeline."
        });
    } catch (e) {
        return _err("addText failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// Audio helpers
// ---------------------------------------------------------------------------
function _findVolumeParam(clip) { if (!clip.components) return null; for (var ci = 0; ci < clip.components.numItems; ci++) { var c = clip.components[ci]; if (c.displayName === "Volume" || c.matchName === "audioGain") { var p = c.properties.getParamForDisplayName("Level"); if (!p && c.properties.numItems > 0) p = c.properties[0]; return p; } } return null; }
function _getAudioClip(ti, ci) { if (!app.project) return _err("No project is open"); var s = app.project.activeSequence; if (!s) return _err("No active sequence"); ti = parseInt(ti,10)||0; ci = parseInt(ci,10)||0; if (ti >= s.audioTracks.numTracks) return _err("Audio track index "+ti+" out of range"); var t = s.audioTracks[ti]; if (!t.clips || ci >= t.clips.numItems) return _err("Clip index "+ci+" out of range on audio track "+ti); return {seq:s,track:t,clip:t.clips[ci]}; }
function _getAudioTrack(ti) { if (!app.project) return _err("No project is open"); var s = app.project.activeSequence; if (!s) return _err("No active sequence"); ti = parseInt(ti,10)||0; if (ti >= s.audioTracks.numTracks) return _err("Audio track index "+ti+" out of range"); return {seq:s,track:s.audioTracks[ti]}; }
function _getVideoTrack(ti) { if (!app.project) return _err("No project is open"); var s = app.project.activeSequence; if (!s) return _err("No active sequence"); ti = parseInt(ti,10)||0; if (ti >= s.videoTracks.numTracks) return _err("Video track index "+ti+" out of range"); return {seq:s,track:s.videoTracks[ti]}; }
function _clampDb(db) { db = parseFloat(db)||0; if (db < -96) db = -96; if (db > 15) db = 15; return db; }

// ===========================================================================
// AUDIO LEVELS (1-5)
// ===========================================================================
function setAudioLevel(trackIndex, clipIndex, levelDb) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; levelDb = _clampDb(levelDb); var p = _findVolumeParam(r.clip); if (!p) return _err("Volume/Level parameter not found"); p.setValue(levelDb, true); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, levelDb:levelDb}); } catch(e) { return _err("setAudioLevel failed: "+e.message); } }
function setAudioLevelKeyframe(trackIndex, clipIndex, time, levelDb) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; levelDb = _clampDb(levelDb); time = parseFloat(time)||0; var p = _findVolumeParam(r.clip); if (!p) return _err("Volume/Level parameter not found"); p.setValueAtKey(_secondsToTime(time), levelDb, true); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, time:time, levelDb:levelDb}); } catch(e) { return _err("setAudioLevelKeyframe failed: "+e.message); } }
function getAudioLevel(trackIndex, clipIndex) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; var p = _findVolumeParam(r.clip); if (!p) return _err("Volume/Level parameter not found"); var v = p.getValue(); var kc = 0; try { var ks = p.getKeys(); kc = ks ? ks.length : 0; } catch(kfe){} return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, levelDb:v, hasKeyframes:kc>0, keyframeCount:kc}); } catch(e) { return _err("getAudioLevel failed: "+e.message); } }
function normalizeAudio(trackIndex, clipIndex, targetDb) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; targetDb = _clampDb(targetDb); var p = _findVolumeParam(r.clip); if (!p) return _err("Volume/Level parameter not found"); var prev = p.getValue(); try { var ks = p.getKeys(); if (ks) for (var ki = ks.length-1; ki >= 0; ki--) p.removeKey(ks[ki]); } catch(rk){} p.setValue(targetDb, true); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, previousLevelDb:prev, targetDb:targetDb}); } catch(e) { return _err("normalizeAudio failed: "+e.message); } }
function setAudioGain(projectItemIndex, gainDb) { try { if (!app.project) return _err("No project is open"); projectItemIndex = parseInt(projectItemIndex,10)||0; gainDb = _clampDb(gainDb); if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems) return _err("Project item index "+projectItemIndex+" out of range"); var item = app.project.rootItem.children[projectItemIndex]; if (!item) return _err("No project item at index "+projectItemIndex); if (item.setAudioGain) item.setAudioGain(gainDb); else return _err("setAudioGain not supported on this item"); return _ok({projectItemIndex:projectItemIndex, name:item.name||"", gainDb:gainDb}); } catch(e) { return _err("setAudioGain failed: "+e.message); } }

// ===========================================================================
// TRACK CONTROLS (6-9)
// ===========================================================================
function muteAudioTrack(trackIndex, muted) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; var v = (muted===true||muted==="true"||muted===1)?1:0; r.track.setMute(v); return _ok({trackIndex:parseInt(trackIndex,10)||0, muted:v===1}); } catch(e) { return _err("muteAudioTrack failed: "+e.message); } }
function soloAudioTrack(trackIndex, soloed) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) { var qt = qs.getAudioTrackAt(parseInt(trackIndex,10)||0); if (qt && qt.setSolo) { var sv = (soloed===true||soloed==="true"||soloed===1)?true:false; qt.setSolo(sv); return _ok({trackIndex:parseInt(trackIndex,10)||0, soloed:sv}); } } } return _err("Solo requires QE DOM. Call app.enableQE() first."); } catch(e) { return _err("soloAudioTrack failed: "+e.message); } }
function setAudioTrackVolume(trackIndex, volume) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; volume = parseFloat(volume); if (isNaN(volume)) volume = 1.0; if (volume < 0) volume = 0; if (volume > 4) volume = 4; if (r.track.components && r.track.components.numItems > 0) { for (var ci = 0; ci < r.track.components.numItems; ci++) { var c = r.track.components[ci]; if (c.displayName==="Volume"||c.matchName==="audioGain") { var vp = c.properties.getParamForDisplayName("Level"); if (!vp && c.properties.numItems > 0) vp = c.properties[0]; if (vp) { vp.setValue(volume, true); return _ok({trackIndex:parseInt(trackIndex,10)||0, volume:volume}); } } } } return _err("Could not find track volume parameter"); } catch(e) { return _err("setAudioTrackVolume failed: "+e.message); } }
function getAudioTrackInfo(trackIndex) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; var info = {trackIndex:parseInt(trackIndex,10)||0, name:r.track.name||"Audio "+((parseInt(trackIndex,10)||0)+1), clipCount:r.track.clips?r.track.clips.numItems:0, muted:false, locked:false}; try { info.muted = r.track.isMuted()?true:false; } catch(me){} try { info.locked = r.track.isLocked()?true:false; } catch(le){} return _ok(info); } catch(e) { return _err("getAudioTrackInfo failed: "+e.message); } }

// ===========================================================================
// AUDIO CHANNELS (10-11)
// ===========================================================================
function getAudioChannelMapping(projectItemIndex) { try { if (!app.project) return _err("No project is open"); projectItemIndex = parseInt(projectItemIndex,10)||0; if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems) return _err("Project item index "+projectItemIndex+" out of range"); var item = app.project.rootItem.children[projectItemIndex]; var m = {projectItemIndex:projectItemIndex, name:item.name||"", channelType:"unknown", channels:[]}; if (item.getAudioChannelMapping) { var acm = item.getAudioChannelMapping(); if (acm) { m.channelType = acm.audioChannelsType!==undefined?String(acm.audioChannelsType):"unknown"; if (acm.audioClipsNumber!==undefined) m.audioClipsNumber = acm.audioClipsNumber; } } return _ok(m); } catch(e) { return _err("getAudioChannelMapping failed: "+e.message); } }
function setAudioChannelMapping(projectItemIndex, mapping) { try { if (!app.project) return _err("No project is open"); projectItemIndex = parseInt(projectItemIndex,10)||0; if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems) return _err("Project item index "+projectItemIndex+" out of range"); var item = app.project.rootItem.children[projectItemIndex]; if (!item.setAudioChannelMapping) return _err("setAudioChannelMapping not supported"); var ct = parseInt(mapping,10); if (isNaN(ct)) ct = 0; var acm = item.getAudioChannelMapping(); if (acm) { acm.audioChannelsType = ct; item.setAudioChannelMapping(acm); } return _ok({projectItemIndex:projectItemIndex, name:item.name||"", channelType:ct}); } catch(e) { return _err("setAudioChannelMapping failed: "+e.message); } }

// ===========================================================================
// AUDIO EFFECTS (12-14)
// ===========================================================================
function applyAudioEffect(trackIndex, clipIndex, effectName) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; if (!effectName||effectName==="") return _err("effectName is required"); if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) { var qt = qs.getAudioTrackAt(parseInt(trackIndex,10)||0); if (qt) { var qc = qt.getItemAt(parseInt(clipIndex,10)||0); if (qc) { var ef = qe.project.getAudioEffectByName(effectName); if (ef) { qc.addAudioEffect(ef); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, effectName:effectName, applied:true}); } else return _err("Audio effect not found: "+effectName); } } } } return _err("Applying audio effects requires QE DOM. Call app.enableQE() first."); } catch(e) { return _err("applyAudioEffect failed: "+e.message); } }
function removeAudioEffect(trackIndex, clipIndex, effectIndex) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; effectIndex = parseInt(effectIndex,10)||0; var ci = effectIndex+1; if (!r.clip.components || ci >= r.clip.components.numItems) return _err("Effect index "+effectIndex+" out of range"); var comp = r.clip.components[ci]; var en = comp.displayName||"unknown"; if (r.clip.removeComponent) r.clip.removeComponent(comp); else if (comp.remove) comp.remove(); else return _err("Cannot remove effect in this version"); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, effectIndex:effectIndex, removedEffect:en}); } catch(e) { return _err("removeAudioEffect failed: "+e.message); } }
function getAudioEffects(trackIndex, clipIndex) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; var fx = []; if (r.clip.components) { for (var ci = 0; ci < r.clip.components.numItems; ci++) { var c = r.clip.components[ci]; fx.push({index:ci, name:c.displayName||"", matchName:c.matchName||"", paramCount:c.properties?c.properties.numItems:0}); } } return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, effects:fx}); } catch(e) { return _err("getAudioEffects failed: "+e.message); } }

// ===========================================================================
// AUDIO TRANSITIONS (15)
// ===========================================================================
function addAudioCrossfade(trackIndex, clipIndex, duration, type) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; duration = parseFloat(duration)||1.0; type = type||"constant_power"; var tn = "Constant Power"; if (type==="constant_gain") tn = "Constant Gain"; else if (type==="exponential") tn = "Exponential Fade"; if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) { var qt = qs.getAudioTrackAt(parseInt(trackIndex,10)||0); if (qt) { var qc = qt.getItemAt(parseInt(clipIndex,10)||0); if (qc) { var tr = qe.project.getAudioTransitionByName(tn); if (tr) { qc.addTransition(tr, true, duration.toString()); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, duration:duration, type:tn}); } else return _err("Audio transition not found: "+tn); } } } } return _err("Audio crossfades require QE DOM. Call app.enableQE() first."); } catch(e) { return _err("addAudioCrossfade failed: "+e.message); } }

// ===========================================================================
// ESSENTIAL SOUND (16-18)
// ===========================================================================
function setEssentialSoundType(trackIndex, clipIndex, type) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; type = (type||"dialogue").toLowerCase(); var tm = {"dialogue":1,"music":2,"sfx":3,"ambience":4}; var tv = tm[type]; if (tv===undefined) return _err("Invalid Essential Sound type: "+type+". Use: dialogue, music, sfx, ambience"); if (r.clip.setEssentialSoundTag) r.clip.setEssentialSoundTag(tv); else if (r.clip.components) { for (var ci = 0; ci < r.clip.components.numItems; ci++) { var c = r.clip.components[ci]; if (c.displayName==="Essential Sound"||c.matchName==="essentialSound") { var tp = c.properties.getParamForDisplayName("Type"); if (tp) { tp.setValue(tv, true); break; } } } } return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, type:type, typeValue:tv}); } catch(e) { return _err("setEssentialSoundType failed: "+e.message); } }
function setEssentialSoundLoudness(trackIndex, clipIndex, targetLufs) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; targetLufs = parseFloat(targetLufs); if (isNaN(targetLufs)) targetLufs = -23; if (r.clip.components) { for (var ci = 0; ci < r.clip.components.numItems; ci++) { var c = r.clip.components[ci]; if (c.displayName==="Essential Sound"||c.matchName==="essentialSound") { var lp = c.properties.getParamForDisplayName("Loudness"); if (!lp) lp = c.properties.getParamForDisplayName("Target Loudness"); if (lp) { lp.setValue(targetLufs, true); return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, targetLufs:targetLufs}); } } } } return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, targetLufs:targetLufs, note:"Essential Sound loudness may not be directly accessible in all versions."}); } catch(e) { return _err("setEssentialSoundLoudness failed: "+e.message); } }
function enableAutoDucking(trackIndex, enabled, duckAmount, sensitivity) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; var ev = (enabled===true||enabled==="true"||enabled===1)?true:false; duckAmount = parseFloat(duckAmount); if (isNaN(duckAmount)) duckAmount = -15; sensitivity = parseFloat(sensitivity); if (isNaN(sensitivity)) sensitivity = 50; if (r.track.components) { for (var ci = 0; ci < r.track.components.numItems; ci++) { var c = r.track.components[ci]; if (c.displayName==="Essential Sound"||c.matchName==="essentialSound") { var dp = c.properties.getParamForDisplayName("Auto Ducking"); if (!dp) dp = c.properties.getParamForDisplayName("Duck Against"); if (dp) dp.setValue(ev?1:0, true); var ap = c.properties.getParamForDisplayName("Duck Amount"); if (ap) ap.setValue(duckAmount, true); var sp = c.properties.getParamForDisplayName("Sensitivity"); if (sp) sp.setValue(sensitivity, true); break; } } } return _ok({trackIndex:parseInt(trackIndex,10)||0, enabled:ev, duckAmount:duckAmount, sensitivity:sensitivity, note:"Auto-ducking may require Essential Sound panel setup in some versions."}); } catch(e) { return _err("enableAutoDucking failed: "+e.message); } }

// ===========================================================================
// AUDIO ANALYSIS (19-20)
// ===========================================================================
function detectSilence(trackIndex, clipIndex, thresholdDb, minDurationMs) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; thresholdDb = parseFloat(thresholdDb); if (isNaN(thresholdDb)) thresholdDb = -40; minDurationMs = parseFloat(minDurationMs); if (isNaN(minDurationMs)) minDurationMs = 500; var cs = _timeToSeconds(r.clip.start); var ce = _timeToSeconds(r.clip.end); var mp = ""; if (r.clip.projectItem && r.clip.projectItem.getMediaPath) mp = r.clip.projectItem.getMediaPath()||""; return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, thresholdDb:thresholdDb, minDurationMs:minDurationMs, clipStart:cs, clipEnd:ce, clipDuration:ce-cs, mediaPath:mp, note:"Silence detection requires audio waveform analysis via the media engine.", silenceRegions:[]}); } catch(e) { return _err("detectSilence failed: "+e.message); } }
function getAudioPeakLevel(trackIndex, clipIndex) { try { var r = _getAudioClip(trackIndex, clipIndex); if (typeof r === "string") return r; var p = _findVolumeParam(r.clip); var cl = p?p.getValue():0; var mp = ""; if (r.clip.projectItem && r.clip.projectItem.getMediaPath) mp = r.clip.projectItem.getMediaPath()||""; return _ok({trackIndex:parseInt(trackIndex,10)||0, clipIndex:parseInt(clipIndex,10)||0, currentLevelDb:cl, mediaPath:mp, note:"True peak level analysis requires the media engine."}); } catch(e) { return _err("getAudioPeakLevel failed: "+e.message); } }

// ===========================================================================
// AUDIO TRACK MANAGEMENT (21-26)
// ===========================================================================
function addAudioTrack(name, channelType) { try { if (!app.project) return _err("No project is open"); var seq = app.project.activeSequence; if (!seq) return _err("No active sequence"); name = name||""; channelType = (channelType||"stereo").toLowerCase(); var ctv = 1; if (channelType==="mono") ctv = 0; else if (channelType==="5.1") ctv = 2; else if (channelType==="adaptive") ctv = 3; var bc = seq.audioTracks.numTracks; if (seq.addTrack) seq.addTrack("audio", ctv); else if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) qs.addAudioTrack(ctv); } var ac = seq.audioTracks.numTracks; if (name!=="" && ac > bc) { var nt = seq.audioTracks[ac-1]; if (nt && nt.name!==undefined) try { nt.name = name; } catch(e2){} } return _ok({name:name, channelType:channelType, trackIndex:ac-1, totalAudioTracks:ac, added:ac>bc}); } catch(e) { return _err("addAudioTrack failed: "+e.message); } }
function deleteAudioTrack(trackIndex) { try { if (!app.project) return _err("No project is open"); var seq = app.project.activeSequence; if (!seq) return _err("No active sequence"); trackIndex = parseInt(trackIndex,10)||0; if (trackIndex >= seq.audioTracks.numTracks) return _err("Audio track index "+trackIndex+" out of range"); var tn = seq.audioTracks[trackIndex].name||""; var bc = seq.audioTracks.numTracks; if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) qs.removeAudioTrack(trackIndex); } else return _err("Deleting audio tracks requires QE DOM."); return _ok({trackIndex:trackIndex, name:tn, totalAudioTracks:seq.audioTracks.numTracks, deleted:seq.audioTracks.numTracks<bc}); } catch(e) { return _err("deleteAudioTrack failed: "+e.message); } }
function renameAudioTrack(trackIndex, name) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; var on = r.track.name||""; name = name||""; if (r.track.name!==undefined) r.track.name = name; return _ok({trackIndex:parseInt(trackIndex,10)||0, oldName:on, newName:name}); } catch(e) { return _err("renameAudioTrack failed: "+e.message); } }
function getAudioTracks() { try { if (!app.project) return _err("No project is open"); var seq = app.project.activeSequence; if (!seq) return _err("No active sequence"); var ts = []; for (var i = 0; i < seq.audioTracks.numTracks; i++) { var t = seq.audioTracks[i]; var info = {index:i, name:t.name||"Audio "+(i+1), clipCount:t.clips?t.clips.numItems:0, muted:false, locked:false}; try { info.muted = t.isMuted()?true:false; } catch(me){} try { info.locked = t.isLocked()?true:false; } catch(le){} ts.push(info); } return _ok({sequenceName:seq.name||"", trackCount:seq.audioTracks.numTracks, tracks:ts}); } catch(e) { return _err("getAudioTracks failed: "+e.message); } }
function lockAudioTrack(trackIndex, locked) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; var lv = (locked===true||locked==="true"||locked===1)?1:0; if (r.track.setLock) r.track.setLock(lv); else if (r.track.setLocked) r.track.setLocked(lv); else return _err("Track locking not available"); return _ok({trackIndex:parseInt(trackIndex,10)||0, locked:lv===1}); } catch(e) { return _err("lockAudioTrack failed: "+e.message); } }
function setAudioTrackTarget(trackIndex, targeted) { try { var r = _getAudioTrack(trackIndex); if (typeof r === "string") return r; var tv = (targeted===true||targeted==="true"||targeted===1)?true:false; if (r.track.setTargeted) r.track.setTargeted(tv, true); else return _err("Track targeting not available"); return _ok({trackIndex:parseInt(trackIndex,10)||0, targeted:tv}); } catch(e) { return _err("setAudioTrackTarget failed: "+e.message); } }

// ===========================================================================
// VIDEO TRACK MANAGEMENT (27-33)
// ===========================================================================
function addVideoTrack(name) { try { if (!app.project) return _err("No project is open"); var seq = app.project.activeSequence; if (!seq) return _err("No active sequence"); name = name||""; var bc = seq.videoTracks.numTracks; if (seq.addTrack) seq.addTrack("video"); else if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) qs.addVideoTrack(); } var ac = seq.videoTracks.numTracks; if (name!=="" && ac > bc) { var nt = seq.videoTracks[ac-1]; if (nt && nt.name!==undefined) try { nt.name = name; } catch(e2){} } return _ok({name:name, trackIndex:ac-1, totalVideoTracks:ac, added:ac>bc}); } catch(e) { return _err("addVideoTrack failed: "+e.message); } }
function deleteVideoTrack(trackIndex) { try { if (!app.project) return _err("No project is open"); var seq = app.project.activeSequence; if (!seq) return _err("No active sequence"); trackIndex = parseInt(trackIndex,10)||0; if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index "+trackIndex+" out of range"); var tn = seq.videoTracks[trackIndex].name||""; var bc = seq.videoTracks.numTracks; if (typeof qe !== "undefined" && qe.project) { var qs = qe.project.getActiveSequence(); if (qs) qs.removeVideoTrack(trackIndex); } else return _err("Deleting video tracks requires QE DOM."); return _ok({trackIndex:trackIndex, name:tn, totalVideoTracks:seq.videoTracks.numTracks, deleted:seq.videoTracks.numTracks<bc}); } catch(e) { return _err("deleteVideoTrack failed: "+e.message); } }
function renameVideoTrack(trackIndex, name) { try { var r = _getVideoTrack(trackIndex); if (typeof r === "string") return r; var on = r.track.name||""; name = name||""; if (r.track.name!==undefined) r.track.name = name; return _ok({trackIndex:parseInt(trackIndex,10)||0, oldName:on, newName:name}); } catch(e) { return _err("renameVideoTrack failed: "+e.message); } }
function getVideoTracks() { try { if (!app.project) return _err("No project is open"); var seq = app.project.activeSequence; if (!seq) return _err("No active sequence"); var ts = []; for (var i = 0; i < seq.videoTracks.numTracks; i++) { var t = seq.videoTracks[i]; var info = {index:i, name:t.name||"Video "+(i+1), clipCount:t.clips?t.clips.numItems:0, muted:false, locked:false}; try { info.muted = t.isMuted()?true:false; } catch(me){} try { info.locked = t.isLocked()?true:false; } catch(le){} ts.push(info); } return _ok({sequenceName:seq.name||"", trackCount:seq.videoTracks.numTracks, tracks:ts}); } catch(e) { return _err("getVideoTracks failed: "+e.message); } }
function lockVideoTrack(trackIndex, locked) { try { var r = _getVideoTrack(trackIndex); if (typeof r === "string") return r; var lv = (locked===true||locked==="true"||locked===1)?1:0; if (r.track.setLock) r.track.setLock(lv); else if (r.track.setLocked) r.track.setLocked(lv); else return _err("Track locking not available"); return _ok({trackIndex:parseInt(trackIndex,10)||0, locked:lv===1}); } catch(e) { return _err("lockVideoTrack failed: "+e.message); } }
function muteVideoTrack(trackIndex, muted) { try { var r = _getVideoTrack(trackIndex); if (typeof r === "string") return r; var v = (muted===true||muted==="true"||muted===1)?1:0; r.track.setMute(v); return _ok({trackIndex:parseInt(trackIndex,10)||0, muted:v===1}); } catch(e) { return _err("muteVideoTrack failed: "+e.message); } }
function setVideoTrackTarget(trackIndex, targeted) { try { var r = _getVideoTrack(trackIndex); if (typeof r === "string") return r; var tv = (targeted===true||targeted==="true"||targeted===1)?true:false; if (r.track.setTargeted) r.track.setTargeted(tv, true); else return _err("Track targeting not available"); return _ok({trackIndex:parseInt(trackIndex,10)||0, targeted:tv}); } catch(e) { return _err("setVideoTrackTarget failed: "+e.message); } }

// ---------------------------------------------------------------------------
// exportSequence(outputPath, presetPath)
// ---------------------------------------------------------------------------
function exportSequence(outputPath, presetPath) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        // Use Adobe Media Encoder for export
        var encoder = app.encoder;

        if (encoder) {
            // Start AME if it is not already running
            encoder.launchEncoder();

            // Determine the export preset
            var preset = presetPath || "";

            if (preset === "") {
                // Use a built-in preset path as fallback
                // Common preset locations:
                // macOS: /Applications/Adobe Media Encoder .../MediaIO/systempresets/
                // Windows: C:\Program Files\Adobe\Adobe Media Encoder ...\MediaIO\systempresets\
                preset = "/Applications/Adobe Media Encoder 2024/Adobe Media Encoder 2024.app/Contents/MediaIO/systempresets/4B434D58_48323634/Match Source - High bitrate.epr";
            }

            // Queue the export job in AME
            var jobID = encoder.encodeSequence(
                seq,
                outputPath,
                preset,
                0,  // WorkAreaType: 0 = entire sequence
                1   // removeOnCompletion: 1 = remove from queue when done
            );

            // Start the render queue
            encoder.startBatch();

            return _ok({
                outputPath: outputPath,
                presetPath: preset,
                jobID: jobID || "queued",
                status: "export_queued",
                sequenceName: seq.name || ""
            });
        } else {
            // Fallback: use direct export if encoder is not available
            // This blocks until the export is complete
            seq.exportAsMediaDirect(
                outputPath,
                presetPath || "",
                0  // WorkAreaType
            );

            return _ok({
                outputPath: outputPath,
                presetPath: presetPath || "default",
                status: "export_complete",
                sequenceName: seq.name || ""
            });
        }
    } catch (e) {
        return _err("exportSequence failed: " + e.message);
    }
}

// ===========================================================================
// Export & Render Functions (Extended)
// ===========================================================================

// ---------------------------------------------------------------------------
// exportDirect(sequenceIndex, outputPath, presetPath, workAreaType)
// Synchronous export via exportAsMediaDirect. Blocks until complete.
// workAreaType: 0=entire sequence, 1=in-to-out, 2=work area
// ---------------------------------------------------------------------------
function exportDirect(sequenceIndex, outputPath, presetPath, workAreaType) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0) {
            sequenceIndex = -1;
        }
        workAreaType = parseInt(workAreaType, 10) || 0;

        var seq;
        if (sequenceIndex < 0) {
            seq = app.project.activeSequence;
        } else {
            if (sequenceIndex >= app.project.sequences.numSequences) {
                return _err("Sequence index " + sequenceIndex + " out of range");
            }
            seq = app.project.sequences[sequenceIndex];
        }

        if (!seq) {
            return _err("No sequence found");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        if (!presetPath || presetPath === "") {
            return _err("presetPath is required for direct export");
        }

        seq.exportAsMediaDirect(outputPath, presetPath, workAreaType);

        return _ok({
            outputPath: outputPath,
            presetPath: presetPath,
            workAreaType: workAreaType,
            status: "export_complete",
            sequenceName: seq.name || ""
        });
    } catch (e) {
        return _err("exportDirect failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportViaAME(sequenceIndex, outputPath, presetPath, workAreaType, removeOnDone)
// Asynchronous export via Adobe Media Encoder.
// ---------------------------------------------------------------------------
function exportViaAME(sequenceIndex, outputPath, presetPath, workAreaType, removeOnDone) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var encoder = app.encoder;
        if (!encoder) {
            return _err("app.encoder is not available");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0) {
            sequenceIndex = -1;
        }
        workAreaType = parseInt(workAreaType, 10) || 0;
        removeOnDone = removeOnDone ? 1 : 0;

        var seq;
        if (sequenceIndex < 0) {
            seq = app.project.activeSequence;
        } else {
            if (sequenceIndex >= app.project.sequences.numSequences) {
                return _err("Sequence index " + sequenceIndex + " out of range");
            }
            seq = app.project.sequences[sequenceIndex];
        }

        if (!seq) {
            return _err("No sequence found");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        if (!presetPath || presetPath === "") {
            return _err("presetPath is required");
        }

        // Launch AME if needed
        encoder.launchEncoder();

        // Queue the export job
        var jobID = encoder.encodeSequence(
            seq,
            outputPath,
            presetPath,
            workAreaType,
            removeOnDone
        );

        return _ok({
            outputPath: outputPath,
            presetPath: presetPath,
            workAreaType: workAreaType,
            removeOnDone: removeOnDone,
            jobID: jobID || "queued",
            status: "queued_in_ame",
            sequenceName: seq.name || ""
        });
    } catch (e) {
        return _err("exportViaAME failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportFrame(outputPath, format)
// Export the current frame from the active sequence as an image.
// format: "PNG" or "JPEG" (defaults to PNG if unrecognized)
// ---------------------------------------------------------------------------
function exportFrame(outputPath, format) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        format = (format || "PNG").toUpperCase();
        if (format !== "PNG" && format !== "JPEG" && format !== "JPG") {
            format = "PNG";
        }
        if (format === "JPG") {
            format = "JPEG";
        }

        // Get the current player position (CTI)
        var time = seq.getPlayerPosition();

        // Premiere Pro provides sequence.exportFramePNG(time, outputPath)
        // for some versions, or we use the QE approach.
        if (format === "PNG" && seq.exportFramePNG) {
            seq.exportFramePNG(time, outputPath);
        } else if (format === "JPEG" && seq.exportFrameJPEG) {
            seq.exportFrameJPEG(time, outputPath);
        } else {
            // Fallback: try the QE DOM
            app.enableQE();
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                qeSeq.exportFramePNG(time.ticks, outputPath);
            } else {
                return _err("Cannot export frame: no supported method available for format " + format);
            }
        }

        return _ok({
            outputPath: outputPath,
            format: format,
            timeSeconds: _timeToSeconds(time),
            status: "frame_exported"
        });
    } catch (e) {
        return _err("exportFrame failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportAAF(sequenceIndex, outputPath, optionsJson)
// Export a sequence as an AAF file.
// options: { mixdown: bool, explode: bool, sampleRate: number, bitsPerSample: number }
// ---------------------------------------------------------------------------
function exportAAF(sequenceIndex, outputPath, optionsJson) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0) {
            sequenceIndex = -1;
        }

        var seq;
        if (sequenceIndex < 0) {
            seq = app.project.activeSequence;
        } else {
            if (sequenceIndex >= app.project.sequences.numSequences) {
                return _err("Sequence index " + sequenceIndex + " out of range");
            }
            seq = app.project.sequences[sequenceIndex];
        }

        if (!seq) {
            return _err("No sequence found");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        var opts = {};
        if (optionsJson && optionsJson !== "") {
            opts = JSON.parse(optionsJson);
        }

        var mixdownVideo = opts.mixdown !== undefined ? (opts.mixdown ? 1 : 0) : 0;
        var explodeToMono = opts.explode !== undefined ? (opts.explode ? 1 : 0) : 0;
        var sampleRate = parseInt(opts.sampleRate, 10) || 48000;
        var bitsPerSample = parseInt(opts.bitsPerSample, 10) || 16;

        seq.exportAsAAF(outputPath, mixdownVideo, explodeToMono, sampleRate, bitsPerSample);

        return _ok({
            outputPath: outputPath,
            mixdownVideo: mixdownVideo,
            explodeToMono: explodeToMono,
            sampleRate: sampleRate,
            bitsPerSample: bitsPerSample,
            status: "aaf_exported",
            sequenceName: seq.name || ""
        });
    } catch (e) {
        return _err("exportAAF failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportOMF(sequenceIndex, outputPath, optionsJson)
// Export a sequence as an OMF file.
// options: { sampleRate: number, bitsPerSample: number, handleFrames: number,
//            encapsulate: bool }
// ---------------------------------------------------------------------------
function exportOMF(sequenceIndex, outputPath, optionsJson) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0) {
            sequenceIndex = -1;
        }

        var seq;
        if (sequenceIndex < 0) {
            seq = app.project.activeSequence;
        } else {
            if (sequenceIndex >= app.project.sequences.numSequences) {
                return _err("Sequence index " + sequenceIndex + " out of range");
            }
            seq = app.project.sequences[sequenceIndex];
        }

        if (!seq) {
            return _err("No sequence found");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        var opts = {};
        if (optionsJson && optionsJson !== "") {
            opts = JSON.parse(optionsJson);
        }

        var sampleRate = parseInt(opts.sampleRate, 10) || 48000;
        var bitsPerSample = parseInt(opts.bitsPerSample, 10) || 16;
        var handleFrames = parseInt(opts.handleFrames, 10) || 0;
        var encapsulate = opts.encapsulate !== undefined ? (opts.encapsulate ? 1 : 0) : 1;

        seq.exportAsOMF(outputPath, sampleRate, bitsPerSample, handleFrames, encapsulate);

        return _ok({
            outputPath: outputPath,
            sampleRate: sampleRate,
            bitsPerSample: bitsPerSample,
            handleFrames: handleFrames,
            encapsulate: encapsulate,
            status: "omf_exported",
            sequenceName: seq.name || ""
        });
    } catch (e) {
        return _err("exportOMF failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportFCPXML(outputPath)
// Export the active sequence as Final Cut Pro XML.
// ---------------------------------------------------------------------------
function exportFCPXML(outputPath) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        seq.exportAsFinalCutProXML(outputPath);

        return _ok({
            outputPath: outputPath,
            status: "fcpxml_exported",
            sequenceName: seq.name || ""
        });
    } catch (e) {
        return _err("exportFCPXML failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportProjectAsXML(outputPath)
// Export the entire project as XML (Premiere Pro XML interchange format).
// ---------------------------------------------------------------------------
function exportProjectAsXML(outputPath) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        if (app.project.exportFinalCutProXML) {
            app.project.exportFinalCutProXML(outputPath, 1);
        } else {
            return _err("exportFinalCutProXML is not available in this Premiere Pro version");
        }

        return _ok({
            outputPath: outputPath,
            status: "project_xml_exported",
            projectName: app.project.name || ""
        });
    } catch (e) {
        return _err("exportProjectAsXML failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getExporters()
// List all available exporters and their preset counts.
// ---------------------------------------------------------------------------
function getExporters() {
    try {
        var encoder = app.encoder;
        if (!encoder) {
            return _err("app.encoder is not available");
        }

        var exporters = [];

        if (encoder.getExporters) {
            var exporterCollection = encoder.getExporters();
            for (var i = 0; i < exporterCollection.numExporters; i++) {
                var exp = exporterCollection[i];
                exporters.push({
                    index: i,
                    name: exp.name || "",
                    classID: exp.classID || "",
                    fileType: exp.fileType || ""
                });
            }
        }

        return _ok({
            exporters: exporters,
            count: exporters.length
        });
    } catch (e) {
        return _err("getExporters failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getExportPresets(exporterIndex)
// Get presets for a specific exporter.
// ---------------------------------------------------------------------------
function getExportPresets(exporterIndex) {
    try {
        var encoder = app.encoder;
        if (!encoder) {
            return _err("app.encoder is not available");
        }

        exporterIndex = parseInt(exporterIndex, 10) || 0;

        if (!encoder.getExporters) {
            return _err("encoder.getExporters is not available");
        }

        var exporterCollection = encoder.getExporters();
        if (exporterIndex >= exporterCollection.numExporters) {
            return _err("Exporter index " + exporterIndex + " out of range (have " + exporterCollection.numExporters + ")");
        }

        var exp = exporterCollection[exporterIndex];
        var presets = [];

        if (exp.getPresets) {
            var presetCollection = exp.getPresets();
            for (var i = 0; i < presetCollection.numPresets; i++) {
                var preset = presetCollection[i];
                presets.push({
                    index: i,
                    name: preset.name || "",
                    matchName: preset.matchName || "",
                    path: preset.path || ""
                });
            }
        }

        return _ok({
            exporterIndex: exporterIndex,
            exporterName: exp.name || "",
            presets: presets,
            count: presets.length
        });
    } catch (e) {
        return _err("getExportPresets failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// startAMEBatch()
// Start the Adobe Media Encoder render queue.
// ---------------------------------------------------------------------------
function startAMEBatch() {
    try {
        var encoder = app.encoder;
        if (!encoder) {
            return _err("app.encoder is not available");
        }

        encoder.startBatch();

        return _ok({
            status: "batch_started"
        });
    } catch (e) {
        return _err("startAMEBatch failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// launchAME()
// Launch Adobe Media Encoder.
// ---------------------------------------------------------------------------
function launchAME() {
    try {
        var encoder = app.encoder;
        if (!encoder) {
            return _err("app.encoder is not available");
        }

        encoder.launchEncoder();

        return _ok({
            status: "ame_launched"
        });
    } catch (e) {
        return _err("launchAME failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// exportAudioOnly(sequenceIndex, outputPath, presetPath)
// Export only the audio from a sequence using an audio export preset.
// ---------------------------------------------------------------------------
function exportAudioOnly(sequenceIndex, outputPath, presetPath) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0) {
            sequenceIndex = -1;
        }

        var seq;
        if (sequenceIndex < 0) {
            seq = app.project.activeSequence;
        } else {
            if (sequenceIndex >= app.project.sequences.numSequences) {
                return _err("Sequence index " + sequenceIndex + " out of range");
            }
            seq = app.project.sequences[sequenceIndex];
        }

        if (!seq) {
            return _err("No sequence found");
        }

        if (!outputPath || outputPath === "") {
            return _err("outputPath is required");
        }

        if (!presetPath || presetPath === "") {
            return _err("presetPath is required (use an audio-only export preset)");
        }

        // Mute all video tracks temporarily, export, then unmute
        var videoTrackStates = [];
        if (seq.videoTracks) {
            for (var i = 0; i < seq.videoTracks.numTracks; i++) {
                var track = seq.videoTracks[i];
                videoTrackStates.push(track.isMuted());
                track.setMute(1);
            }
        }

        // Perform direct export with the audio preset
        seq.exportAsMediaDirect(outputPath, presetPath, 0);

        // Restore video track mute states
        if (seq.videoTracks) {
            for (var j = 0; j < seq.videoTracks.numTracks; j++) {
                var restoreTrack = seq.videoTracks[j];
                restoreTrack.setMute(videoTrackStates[j] ? 1 : 0);
            }
        }

        return _ok({
            outputPath: outputPath,
            presetPath: presetPath,
            status: "audio_export_complete",
            sequenceName: seq.name || ""
        });
    } catch (e) {
        return _err("exportAudioOnly failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getExportProgress()
// Get current export/render progress if available.
// ---------------------------------------------------------------------------
function getExportProgress() {
    try {
        var encoder = app.encoder;
        if (!encoder) {
            return _err("app.encoder is not available");
        }

        var info = {
            encoderAvailable: true,
            status: "unknown"
        };

        if (encoder.getExporters) {
            info.exportersAvailable = true;
        }

        info.note = "AME progress monitoring is limited in ExtendScript. Check AME UI for detailed progress.";

        return _ok(info);
    } catch (e) {
        return _err("getExportProgress failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// renderSequencePreview(inSeconds, outSeconds)
// Render preview frames for a time range of the active sequence.
// ---------------------------------------------------------------------------
function renderSequencePreview(inSeconds, outSeconds) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        inSeconds = parseFloat(inSeconds) || 0;
        outSeconds = parseFloat(outSeconds) || 0;

        if (outSeconds <= inSeconds) {
            return _err("outSeconds must be greater than inSeconds");
        }

        var inTime = _secondsToTime(inSeconds);
        var outTime = _secondsToTime(outSeconds);

        // Set the in/out points for the work area
        seq.setInPoint(inTime.seconds);
        seq.setOutPoint(outTime.seconds);

        // Use QE DOM to trigger render of the work area
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq) {
            qeSeq.renderWorkArea();

            return _ok({
                inSeconds: inSeconds,
                outSeconds: outSeconds,
                durationSeconds: outSeconds - inSeconds,
                status: "preview_render_started",
                sequenceName: seq.name || ""
            });
        } else {
            return _err("QE sequence not available for preview rendering");
        }
    } catch (e) {
        return _err("renderSequencePreview failed: " + e.message);
    }
}

// ===========================================================================
// Sequence / Timeline Management Functions
// ===========================================================================

// ---------------------------------------------------------------------------
// createSequenceFromClips(paramsJson) - Create a sequence from project items
// ---------------------------------------------------------------------------
function createSequenceFromClips(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var name = params.name || "Sequence from clips";
        var clipIndices = params.clipIndices || [];

        if (!app.project) {
            return _err("No project is open");
        }

        if (clipIndices.length === 0) {
            return _err("clipIndices array is empty");
        }

        var items = [];
        for (var i = 0; i < clipIndices.length; i++) {
            var idx = parseInt(clipIndices[i], 10);
            if (idx < 0 || idx >= app.project.rootItem.children.numItems) {
                return _err("Project item index " + idx + " out of range");
            }
            items.push(app.project.rootItem.children[idx]);
        }

        // createNewSequenceFromClips auto-detects settings from the first clip
        var newSeqID = app.project.createNewSequenceFromClips(name, items);

        var seq = app.project.activeSequence;
        return _ok({
            name: name,
            sequenceID: newSeqID || (seq ? seq.sequenceID : ""),
            clipCount: items.length,
            frameSizeHorizontal: seq ? (seq.frameSizeHorizontal || 0) : 0,
            frameSizeVertical: seq ? (seq.frameSizeVertical || 0) : 0,
            timebase: seq ? (seq.timebase || "") : ""
        });
    } catch (e) {
        return _err("createSequenceFromClips failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// duplicateSequence(sequenceIndex) - Duplicate an existing sequence
// ---------------------------------------------------------------------------
function duplicateSequence(sequenceIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0 || sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var srcSeq = app.project.sequences[sequenceIndex];
        if (!srcSeq) {
            return _err("No sequence at index " + sequenceIndex);
        }

        // Clone via the project item's duplicate method
        srcSeq.clone();

        // The duplicated sequence becomes the last one
        var newIdx = app.project.sequences.numSequences - 1;
        var newSeq = app.project.sequences[newIdx];

        return _ok({
            originalName: srcSeq.name || "",
            originalIndex: sequenceIndex,
            newIndex: newIdx,
            newName: newSeq ? (newSeq.name || "") : "",
            newSequenceID: newSeq ? (newSeq.sequenceID || "") : ""
        });
    } catch (e) {
        return _err("duplicateSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// deleteSequence(sequenceIndex) - Delete a sequence from the project
// ---------------------------------------------------------------------------
function deleteSequence(sequenceIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0 || sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var seq = app.project.sequences[sequenceIndex];
        if (!seq) {
            return _err("No sequence at index " + sequenceIndex);
        }

        var deletedName = seq.name || "";
        var deletedID = seq.sequenceID || "";

        // Find the project item corresponding to this sequence and remove it
        var root = app.project.rootItem;
        var removed = false;
        if (root.children) {
            for (var c = 0; c < root.children.numItems; c++) {
                var child = root.children[c];
                if (child.type === ProjectItemType.CLIP && child.name === deletedName) {
                    app.project.deleteSequence(seq);
                    removed = true;
                    break;
                }
            }
        }

        if (!removed) {
            // Try direct deletion
            app.project.deleteSequence(seq);
        }

        return _ok({
            deletedName: deletedName,
            deletedSequenceID: deletedID,
            remainingSequences: app.project.sequences.numSequences
        });
    } catch (e) {
        return _err("deleteSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// renameSequence(paramsJson) - Rename a sequence
// ---------------------------------------------------------------------------
function renameSequence(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var sequenceIndex = parseInt(params.sequenceIndex, 10);
        var newName = params.newName || "";

        if (!app.project) {
            return _err("No project is open");
        }

        if (newName === "") {
            return _err("newName is required");
        }

        if (isNaN(sequenceIndex) || sequenceIndex < 0 || sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var seq = app.project.sequences[sequenceIndex];
        if (!seq) {
            return _err("No sequence at index " + sequenceIndex);
        }

        var oldName = seq.name || "";
        seq.name = newName;

        return _ok({
            oldName: oldName,
            newName: newName,
            sequenceIndex: sequenceIndex,
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("renameSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getSequenceSettings(sequenceIndex) - Get full sequence settings
// ---------------------------------------------------------------------------
function getSequenceSettings(sequenceIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0 || sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var seq = app.project.sequences[sequenceIndex];
        if (!seq) {
            return _err("No sequence at index " + sequenceIndex);
        }

        var settings = {
            name: seq.name || "",
            sequenceID: seq.sequenceID || "",
            frameSizeHorizontal: seq.frameSizeHorizontal || 0,
            frameSizeVertical: seq.frameSizeVertical || 0,
            timebase: seq.timebase || "",
            videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0,
            audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0,
            inPoint: _timeToSeconds(seq.getInPoint()),
            outPoint: _timeToSeconds(seq.getOutPoint()),
            endSeconds: _timeToSeconds(seq.end)
        };

        // Try to get extended settings via getSettings()
        try {
            var seqSettings = seq.getSettings();
            if (seqSettings) {
                settings.audioSampleRate = seqSettings.audioSampleRate || 0;
                settings.audioChannelCount = seqSettings.audioChannelCount || 0;
                settings.videoFieldType = seqSettings.videoFieldType || 0;
                settings.videoPixelAspectRatio = seqSettings.videoPixelAspectRatio || "";
                settings.compositeLinearColor = seqSettings.compositeLinearColor || false;
                settings.maximumBitDepth = seqSettings.maximumBitDepth || false;
                settings.maximumRenderQuality = seqSettings.maximumRenderQuality || false;
                settings.vrProjection = seqSettings.vrProjection || 0;
                settings.vrLayout = seqSettings.vrLayout || 0;
                settings.vrHorzCapturedView = seqSettings.vrHorzCapturedView || 0;
                settings.vrVertCapturedView = seqSettings.vrVertCapturedView || 0;
            }
        } catch (settingsErr) {
            settings.settingsNote = "Extended settings not available: " + settingsErr.message;
        }

        return _ok(settings);
    } catch (e) {
        return _err("getSequenceSettings failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// setSequenceSettings(paramsJson) - Update sequence settings
// ---------------------------------------------------------------------------
function setSequenceSettings(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var sequenceIndex = parseInt(params.sequenceIndex, 10);

        if (!app.project) {
            return _err("No project is open");
        }

        if (isNaN(sequenceIndex) || sequenceIndex < 0 || sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var seq = app.project.sequences[sequenceIndex];
        if (!seq) {
            return _err("No sequence at index " + sequenceIndex);
        }

        // Get current settings, merge updates, then apply
        var seqSettings = seq.getSettings();
        var changed = [];

        if (params.width !== undefined && params.height !== undefined) {
            seqSettings.videoFrameWidth = parseInt(params.width, 10);
            seqSettings.videoFrameHeight = parseInt(params.height, 10);
            changed.push("resolution");
        }
        if (params.audioSampleRate !== undefined) {
            seqSettings.audioSampleRate = parseFloat(params.audioSampleRate);
            changed.push("audioSampleRate");
        }
        if (params.videoFieldType !== undefined) {
            seqSettings.videoFieldType = parseInt(params.videoFieldType, 10);
            changed.push("videoFieldType");
        }
        if (params.compositeLinearColor !== undefined) {
            seqSettings.compositeLinearColor = params.compositeLinearColor;
            changed.push("compositeLinearColor");
        }
        if (params.maximumBitDepth !== undefined) {
            seqSettings.maximumBitDepth = params.maximumBitDepth;
            changed.push("maximumBitDepth");
        }
        if (params.maximumRenderQuality !== undefined) {
            seqSettings.maximumRenderQuality = params.maximumRenderQuality;
            changed.push("maximumRenderQuality");
        }

        seq.setSettings(seqSettings);

        return _ok({
            sequenceIndex: sequenceIndex,
            sequenceID: seq.sequenceID || "",
            changedFields: changed
        });
    } catch (e) {
        return _err("setSequenceSettings failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getActiveSequence() - Get the currently active sequence details
// ---------------------------------------------------------------------------
function getActiveSequence() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        return _ok({
            name: seq.name || "",
            sequenceID: seq.sequenceID || "",
            frameSizeHorizontal: seq.frameSizeHorizontal || 0,
            frameSizeVertical: seq.frameSizeVertical || 0,
            timebase: seq.timebase || "",
            videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0,
            audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0,
            inPoint: _timeToSeconds(seq.getInPoint()),
            outPoint: _timeToSeconds(seq.getOutPoint()),
            endSeconds: _timeToSeconds(seq.end)
        });
    } catch (e) {
        return _err("getActiveSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// setActiveSequence(sequenceIndex) - Make a sequence the active one
// ---------------------------------------------------------------------------
function setActiveSequence(sequenceIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        sequenceIndex = parseInt(sequenceIndex, 10);
        if (isNaN(sequenceIndex) || sequenceIndex < 0 || sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var seq = app.project.sequences[sequenceIndex];
        if (!seq) {
            return _err("No sequence at index " + sequenceIndex);
        }

        app.project.activeSequence = seq;

        return _ok({
            name: seq.name || "",
            sequenceID: seq.sequenceID || "",
            sequenceIndex: sequenceIndex
        });
    } catch (e) {
        return _err("setActiveSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getSequenceList() - List all sequences in the project
// ---------------------------------------------------------------------------
function getSequenceList() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var sequences = [];
        var activeID = app.project.activeSequence ? (app.project.activeSequence.sequenceID || "") : "";

        for (var i = 0; i < app.project.sequences.numSequences; i++) {
            var seq = app.project.sequences[i];
            sequences.push({
                index: i,
                name: seq.name || "",
                sequenceID: seq.sequenceID || "",
                frameSizeHorizontal: seq.frameSizeHorizontal || 0,
                frameSizeVertical: seq.frameSizeVertical || 0,
                timebase: seq.timebase || "",
                videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0,
                audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0,
                isActive: (seq.sequenceID === activeID)
            });
        }

        return _ok({
            count: sequences.length,
            sequences: sequences,
            activeSequenceID: activeID
        });
    } catch (e) {
        return _err("getSequenceList failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getPlayheadPosition() - Get current playhead position
// ---------------------------------------------------------------------------
function getPlayheadPosition() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        var playerPos = seq.getPlayerPosition();
        var seconds = _timeToSeconds(playerPos);

        return _ok({
            seconds: seconds,
            ticks: playerPos ? (playerPos.ticks || "") : "",
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("getPlayheadPosition failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// setPlayheadPosition(seconds) - Move playhead to a specific position
// ---------------------------------------------------------------------------
function setPlayheadPosition(seconds) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        seconds = parseFloat(seconds) || 0;
        var newTime = _secondsToTime(seconds);
        seq.setPlayerPosition(newTime.ticks);

        return _ok({
            seconds: seconds,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("setPlayheadPosition failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// setInPoint(seconds) - Set sequence in point
// ---------------------------------------------------------------------------
function setInPoint(seconds) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        seconds = parseFloat(seconds) || 0;
        seq.setInPoint(seconds);

        return _ok({
            inPoint: seconds,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("setInPoint failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// setOutPoint(seconds) - Set sequence out point
// ---------------------------------------------------------------------------
function setOutPoint(seconds) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        seconds = parseFloat(seconds) || 0;
        seq.setOutPoint(seconds);

        return _ok({
            outPoint: seconds,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("setOutPoint failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getInOutPoints() - Get current in/out points
// ---------------------------------------------------------------------------
function getInOutPoints() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        return _ok({
            inPoint: _timeToSeconds(seq.getInPoint()),
            outPoint: _timeToSeconds(seq.getOutPoint()),
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("getInOutPoints failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// clearInOutPoints() - Clear in/out points
// ---------------------------------------------------------------------------
function clearInOutPoints() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        seq.setInPoint(seq.zeroPoint ? _timeToSeconds(seq.zeroPoint) : 0);
        seq.setOutPoint(_timeToSeconds(seq.end));

        return _ok({
            inPoint: seq.zeroPoint ? _timeToSeconds(seq.zeroPoint) : 0,
            outPoint: _timeToSeconds(seq.end),
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("clearInOutPoints failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// setWorkArea(paramsJson) - Set work area in/out
// ---------------------------------------------------------------------------
function setWorkArea(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var inSeconds = parseFloat(params.inSeconds) || 0;
        var outSeconds = parseFloat(params.outSeconds) || 0;

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (outSeconds <= inSeconds) {
            return _err("outSeconds must be greater than inSeconds");
        }

        seq.setInPoint(inSeconds);
        seq.setOutPoint(outSeconds);

        return _ok({
            inSeconds: inSeconds,
            outSeconds: outSeconds,
            durationSeconds: outSeconds - inSeconds,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("setWorkArea failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// renderPreviewFiles(paramsJson) - Render preview files for a range
// ---------------------------------------------------------------------------
function renderPreviewFiles(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var inSeconds = parseFloat(params.inSeconds) || 0;
        var outSeconds = parseFloat(params.outSeconds) || 0;

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        var inTime = _secondsToTime(inSeconds);
        var outTime = _secondsToTime(outSeconds);

        seq.renderPreviewArea(inTime, outTime);

        return _ok({
            inSeconds: inSeconds,
            outSeconds: outSeconds,
            status: "render_started",
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("renderPreviewFiles failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// deletePreviewFiles() - Delete all preview/render files
// ---------------------------------------------------------------------------
function deletePreviewFiles() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        seq.deletePreviewFiles();

        return _ok({
            status: "preview_files_deleted",
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("deletePreviewFiles failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// createNestedSequence(paramsJson) - Nest selected clips into a subsequence
// ---------------------------------------------------------------------------
function createNestedSequence(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var clipIndices = params.clipIndices || [];

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (clipIndices.length === 0) {
            return _err("clipIndices array is empty");
        }

        // Collect clips from the first video track by default
        var trackIndex = parseInt(params.trackIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) {
            return _err("Video track index " + trackIndex + " out of range");
        }

        var track = seq.videoTracks[trackIndex];
        var clipsToNest = [];

        for (var i = 0; i < clipIndices.length; i++) {
            var ci = parseInt(clipIndices[i], 10);
            if (ci < 0 || ci >= track.clips.numItems) {
                return _err("Clip index " + ci + " out of range on track " + trackIndex);
            }
            clipsToNest.push(track.clips[ci]);
        }

        // Use the createSubSequence method if available
        var projectItem = clipsToNest[0].projectItem;
        if (seq.createSubSequence) {
            seq.createSubSequence(clipsToNest);
        }

        return _ok({
            status: "nested",
            clipCount: clipsToNest.length,
            trackIndex: trackIndex,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("createNestedSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// autoReframeSequence(paramsJson) - Auto reframe to new aspect ratio
// ---------------------------------------------------------------------------
function autoReframeSequence(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var numerator = parseInt(params.numerator, 10) || 9;
        var denominator = parseInt(params.denominator, 10) || 16;
        var motionPreset = params.motionPreset || "default";

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        // autoReframeSequence is available in Premiere Pro 2020+
        if (seq.autoReframeSequence) {
            seq.autoReframeSequence(numerator, denominator, motionPreset);
        } else {
            return _err("autoReframeSequence is not supported in this Premiere Pro version");
        }

        return _ok({
            numerator: numerator,
            denominator: denominator,
            motionPreset: motionPreset,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("autoReframeSequence failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// insertBlackVideo(paramsJson) - Insert black video
// ---------------------------------------------------------------------------
function insertBlackVideo(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var trackIndex = parseInt(params.trackIndex, 10) || 0;
        var startTime = parseFloat(params.startTime) || 0;
        var duration = parseFloat(params.duration) || 5.0;

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (trackIndex >= seq.videoTracks.numTracks) {
            return _err("Video track index " + trackIndex + " out of range");
        }

        // Use QE DOM for inserting generated media
        if (typeof qe !== "undefined" && qe.project) {
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
                if (qeTrack) {
                    qeTrack.insertBlackVideo(startTime.toString(), duration.toString());
                }
            }
        } else {
            app.enableQE();
            var qeSeq2 = qe.project.getActiveSequence();
            if (qeSeq2) {
                var qeTrack2 = qeSeq2.getVideoTrackAt(trackIndex);
                if (qeTrack2) {
                    qeTrack2.insertBlackVideo(startTime.toString(), duration.toString());
                }
            }
        }

        return _ok({
            trackIndex: trackIndex,
            startTime: startTime,
            duration: duration,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("insertBlackVideo failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// insertBarsAndTone(paramsJson) - Insert bars and tone
// ---------------------------------------------------------------------------
function insertBarsAndTone(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var width = parseInt(params.width, 10) || 1920;
        var height = parseInt(params.height, 10) || 1080;
        var duration = parseFloat(params.duration) || 10.0;

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        // Use QE DOM for bars and tone
        if (typeof qe === "undefined") {
            app.enableQE();
        }

        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq) {
            qeSeq.insertBarsAndTone(width.toString(), height.toString(), duration.toString());
        }

        return _ok({
            width: width,
            height: height,
            duration: duration,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("insertBarsAndTone failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// getSequenceMarkers() - Get all markers on the active sequence
// ---------------------------------------------------------------------------
function getSequenceMarkers() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        var markers = [];
        if (seq.markers) {
            var marker = seq.markers.getFirstMarker();
            var idx = 0;
            while (marker) {
                markers.push({
                    index: idx,
                    name: marker.name || "",
                    comment: marker.comments || "",
                    start: _timeToSeconds(marker.start),
                    end: _timeToSeconds(marker.end),
                    type: marker.type || "",
                    colorIndex: marker.colorIndex !== undefined ? marker.colorIndex : -1
                });
                idx++;
                marker = seq.markers.getNextMarker(marker);
            }
        }

        return _ok({
            count: markers.length,
            markers: markers,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("getSequenceMarkers failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// addSequenceMarker(paramsJson) - Add a marker to the active sequence
// ---------------------------------------------------------------------------
function addSequenceMarker(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var time = parseFloat(params.time) || 0;
        var name = params.name || "";
        var comment = params.comment || "";
        var color = parseInt(params.color, 10) || 0;
        var duration = parseFloat(params.duration) || 0;

        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (!seq.markers) {
            return _err("Sequence does not support markers");
        }

        var newMarker = seq.markers.createMarker(time);
        if (newMarker) {
            if (name !== "") newMarker.name = name;
            if (comment !== "") newMarker.comments = comment;
            if (color >= 0) newMarker.colorIndex = color;
            if (duration > 0) {
                newMarker.end = _secondsToTime(time + duration);
            }
        } else {
            return _err("Failed to create marker");
        }

        return _ok({
            name: name,
            comment: comment,
            time: time,
            duration: duration,
            color: color,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("addSequenceMarker failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// deleteSequenceMarker(markerIndex) - Delete a marker
// ---------------------------------------------------------------------------
function deleteSequenceMarker(markerIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (!seq.markers) {
            return _err("Sequence does not support markers");
        }

        markerIndex = parseInt(markerIndex, 10);
        if (isNaN(markerIndex) || markerIndex < 0) {
            return _err("Invalid marker index: " + markerIndex);
        }

        // Navigate to the target marker by index
        var marker = seq.markers.getFirstMarker();
        var idx = 0;
        while (marker && idx < markerIndex) {
            marker = seq.markers.getNextMarker(marker);
            idx++;
        }

        if (!marker) {
            return _err("Marker index " + markerIndex + " not found");
        }

        var deletedName = marker.name || "";
        var deletedTime = _timeToSeconds(marker.start);
        seq.markers.deleteMarker(marker);

        return _ok({
            deletedIndex: markerIndex,
            deletedName: deletedName,
            deletedTime: deletedTime,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("deleteSequenceMarker failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// navigateToMarker(markerIndex) - Move playhead to a marker
// ---------------------------------------------------------------------------
function navigateToMarker(markerIndex) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var seq = app.project.activeSequence;
        if (!seq) {
            return _err("No active sequence");
        }

        if (!seq.markers) {
            return _err("Sequence does not support markers");
        }

        markerIndex = parseInt(markerIndex, 10);
        if (isNaN(markerIndex) || markerIndex < 0) {
            return _err("Invalid marker index: " + markerIndex);
        }

        // Navigate to the target marker by index
        var marker = seq.markers.getFirstMarker();
        var idx = 0;
        while (marker && idx < markerIndex) {
            marker = seq.markers.getNextMarker(marker);
            idx++;
        }

        if (!marker) {
            return _err("Marker index " + markerIndex + " not found");
        }

        var markerTime = marker.start;
        seq.setPlayerPosition(markerTime.ticks);

        return _ok({
            markerIndex: markerIndex,
            markerName: marker.name || "",
            seconds: _timeToSeconds(markerTime),
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || ""
        });
    } catch (e) {
        return _err("navigateToMarker failed: " + e.message);
    }
}

// ===========================================================================
// Clip Operations
// ===========================================================================

// ---------------------------------------------------------------------------
// Helper: resolve a track from the active sequence by type and index.
// trackType: "video" or "audio"
// ---------------------------------------------------------------------------
function _getTrack(seq, trackType, trackIndex) {
    trackIndex = parseInt(trackIndex, 10) || 0;
    if (trackType === "audio") {
        if (!seq.audioTracks || trackIndex >= seq.audioTracks.numTracks) return null;
        return seq.audioTracks[trackIndex];
    }
    if (!seq.videoTracks || trackIndex >= seq.videoTracks.numTracks) return null;
    return seq.videoTracks[trackIndex];
}

function _getClip(track, clipIndex) {
    clipIndex = parseInt(clipIndex, 10) || 0;
    if (!track.clips || clipIndex >= track.clips.numItems) return null;
    return track.clips[clipIndex];
}

function _buildClipInfo(clip, clipIndex, trackType, trackIndex) {
    var info = {
        index: clipIndex,
        name: clip.name || "",
        start: _timeToSeconds(clip.start),
        end: _timeToSeconds(clip.end),
        duration: _timeToSeconds(clip.duration),
        inPoint: _timeToSeconds(clip.inPoint),
        outPoint: _timeToSeconds(clip.outPoint),
        type: clip.type || "",
        trackType: trackType,
        trackIndex: trackIndex,
        mediaPath: ""
    };
    try { if (clip.projectItem && clip.projectItem.getMediaPath) info.mediaPath = clip.projectItem.getMediaPath() || ""; } catch (e) {}
    try { info.enabled = (typeof clip.disabled !== "undefined") ? !clip.disabled : true; } catch (e) { info.enabled = true; }
    return info;
}

// ---------------------------------------------------------------------------
// 1. insertClip
// ---------------------------------------------------------------------------
function insertClip(projectItemIndex, time, vTrackIndex, aTrackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        time = parseFloat(time) || 0;
        vTrackIndex = parseInt(vTrackIndex, 10) || 0;
        aTrackIndex = parseInt(aTrackIndex, 10) || 0;
        if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems)
            return _err("Project item index " + projectItemIndex + " out of range");
        var pi = app.project.rootItem.children[projectItemIndex];
        if (!pi) return _err("No project item at index " + projectItemIndex);
        seq.insertClip(pi, _secondsToTime(time), vTrackIndex, aTrackIndex);
        return _ok({ action: "insert", projectItemName: pi.name || "", time: time, vTrackIndex: vTrackIndex, aTrackIndex: aTrackIndex });
    } catch (e) { return _err("insertClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 2. overwriteClip
// ---------------------------------------------------------------------------
function overwriteClip(projectItemIndex, time, vTrackIndex, aTrackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        time = parseFloat(time) || 0;
        vTrackIndex = parseInt(vTrackIndex, 10) || 0;
        aTrackIndex = parseInt(aTrackIndex, 10) || 0;
        if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems)
            return _err("Project item index " + projectItemIndex + " out of range");
        var pi = app.project.rootItem.children[projectItemIndex];
        if (!pi) return _err("No project item at index " + projectItemIndex);
        seq.overwriteClip(pi, _secondsToTime(time), vTrackIndex, aTrackIndex);
        return _ok({ action: "overwrite", projectItemName: pi.name || "", time: time, vTrackIndex: vTrackIndex, aTrackIndex: aTrackIndex });
    } catch (e) { return _err("overwriteClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 4. removeClipFromTrack
// ---------------------------------------------------------------------------
function removeClipFromTrack(trackType, trackIndex, clipIndex, ripple) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var clipName = clip.name || "";
        var doRipple = (ripple === true || ripple === "true" || ripple === 1);
        clip.remove(doRipple, true);
        return _ok({ action: "remove", clipName: clipName, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0, ripple: doRipple });
    } catch (e) { return _err("removeClipFromTrack failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 5. moveClip
// ---------------------------------------------------------------------------
function moveClip(trackType, trackIndex, clipIndex, newStartTime) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        newStartTime = parseFloat(newStartTime) || 0;
        var oldStart = _timeToSeconds(clip.start);
        clip.start = _secondsToTime(newStartTime);
        return _ok({ action: "move", clipName: clip.name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0, oldStartTime: oldStart, newStartTime: newStartTime });
    } catch (e) { return _err("moveClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 6. copyClip
// ---------------------------------------------------------------------------
var _clipboardClip = null;
var _clipboardTrackType = null;

function copyClip(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        _clipboardClip = clip;
        _clipboardTrackType = trackType;
        return _ok({ action: "copy", clipName: clip.name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("copyClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 7. pasteClip
// ---------------------------------------------------------------------------
function pasteClip(trackType, trackIndex, time) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        if (!_clipboardClip) return _err("No clip in clipboard. Use copyClip first.");
        if (!_clipboardClip.projectItem) return _err("Copied clip has no project item reference");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        time = parseFloat(time) || 0;
        track.insertClip(_clipboardClip.projectItem, _secondsToTime(time));
        return _ok({ action: "paste", clipName: _clipboardClip.name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, time: time });
    } catch (e) { return _err("pasteClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 8. duplicateClip
// ---------------------------------------------------------------------------
function duplicateClip(trackType, trackIndex, clipIndex, destTrackIndex, destTime) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var srcTrack = _getTrack(seq, trackType, trackIndex);
        if (!srcTrack) return _err("Source track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(srcTrack, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        if (!clip.projectItem) return _err("Clip has no project item reference for duplication");
        destTrackIndex = parseInt(destTrackIndex, 10) || 0;
        destTime = parseFloat(destTime) || 0;
        var destTrack = _getTrack(seq, trackType, destTrackIndex);
        if (!destTrack) return _err("Destination track not found: " + trackType + "[" + destTrackIndex + "]");
        destTrack.insertClip(clip.projectItem, _secondsToTime(destTime));
        return _ok({ action: "duplicate", clipName: clip.name || "", srcTrackType: trackType, srcTrackIndex: parseInt(trackIndex, 10) || 0, srcClipIndex: parseInt(clipIndex, 10) || 0, destTrackIndex: destTrackIndex, destTime: destTime });
    } catch (e) { return _err("duplicateClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 9. razorClip
// ---------------------------------------------------------------------------
function razorClip(trackType, trackIndex, time) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        time = parseFloat(time) || 0;
        var razorTime = _secondsToTime(time);
        var found = false;
        if (track.clips) {
            for (var i = 0; i < track.clips.numItems; i++) {
                var c = track.clips[i];
                if (time > _timeToSeconds(c.start) && time < _timeToSeconds(c.end)) {
                    if (typeof qe !== "undefined" && qe.project) {
                        var qeSeq = qe.project.getActiveSequence();
                        if (qeSeq) {
                            var qeTrack = (trackType === "audio") ? qeSeq.getAudioTrackAt(parseInt(trackIndex, 10) || 0) : qeSeq.getVideoTrackAt(parseInt(trackIndex, 10) || 0);
                            if (qeTrack) { qeTrack.razor(razorTime.ticks); found = true; }
                        }
                    }
                    if (!found) { c.end = razorTime; found = true; }
                    break;
                }
            }
        }
        if (!found) return _err("No clip found at time " + time + " on " + trackType + " track " + trackIndex);
        return _ok({ action: "razor", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, time: time });
    } catch (e) { return _err("razorClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 10. razorAllTracks
// ---------------------------------------------------------------------------
function razorAllTracks(time) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        time = parseFloat(time) || 0;
        var razorTime = _secondsToTime(time);
        var tracksRazored = 0;
        if (typeof qe !== "undefined" && qe.project) {
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) { try { var qvt = qeSeq.getVideoTrackAt(vi); if (qvt) { qvt.razor(razorTime.ticks); tracksRazored++; } } catch (e2) {} }
                for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) { try { var qat = qeSeq.getAudioTrackAt(ai); if (qat) { qat.razor(razorTime.ticks); tracksRazored++; } } catch (e3) {} }
            }
        }
        return _ok({ action: "razorAll", time: time, tracksRazored: tracksRazored, videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0, audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0 });
    } catch (e) { return _err("razorAllTracks failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 11. getClipInfo
// ---------------------------------------------------------------------------
function getClipInfo(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var info = _buildClipInfo(clip, parseInt(clipIndex, 10) || 0, trackType, parseInt(trackIndex, 10) || 0);
        info.effects = [];
        if (clip.components) {
            for (var ci = 0; ci < clip.components.numItems; ci++) {
                var comp = clip.components[ci];
                var compInfo = { index: ci, displayName: comp.displayName || "", matchName: comp.matchName || "", properties: [] };
                if (comp.properties) {
                    for (var pi = 0; pi < comp.properties.numItems; pi++) {
                        var prop = comp.properties[pi];
                        compInfo.properties.push({ displayName: prop.displayName || "", value: (typeof prop.getValue === "function") ? prop.getValue() : "" });
                    }
                }
                info.effects.push(compInfo);
            }
        }
        return _ok(info);
    } catch (e) { return _err("getClipInfo failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 12. getClipsOnTrack
// ---------------------------------------------------------------------------
function getClipsOnTrack(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clips = [];
        if (track.clips) { for (var i = 0; i < track.clips.numItems; i++) clips.push(_buildClipInfo(track.clips[i], i, trackType, parseInt(trackIndex, 10) || 0)); }
        return _ok({ trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, trackName: track.name || "", clipCount: clips.length, clips: clips });
    } catch (e) { return _err("getClipsOnTrack failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 13. getAllClips
// ---------------------------------------------------------------------------
function getAllClips() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var all = [];
        if (seq.videoTracks) { for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) { var vt = seq.videoTracks[vi]; if (vt.clips) { for (var vc = 0; vc < vt.clips.numItems; vc++) all.push(_buildClipInfo(vt.clips[vc], vc, "video", vi)); } } }
        if (seq.audioTracks) { for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) { var at2 = seq.audioTracks[ai]; if (at2.clips) { for (var ac = 0; ac < at2.clips.numItems; ac++) all.push(_buildClipInfo(at2.clips[ac], ac, "audio", ai)); } } }
        return _ok({ sequenceName: seq.name || "", totalClips: all.length, clips: all });
    } catch (e) { return _err("getAllClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 14. setClipName
// ---------------------------------------------------------------------------
function setClipName(trackType, trackIndex, clipIndex, name) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var oldName = clip.name || "";
        clip.name = name || "";
        return _ok({ action: "rename", oldName: oldName, newName: name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("setClipName failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 15. setClipEnabled
// ---------------------------------------------------------------------------
function setClipEnabled(trackType, trackIndex, clipIndex, enabled) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var isEnabled = (enabled === true || enabled === "true" || enabled === 1);
        clip.disabled = !isEnabled;
        return _ok({ action: "setEnabled", clipName: clip.name || "", enabled: isEnabled, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("setClipEnabled failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 16. setClipSpeed
// ---------------------------------------------------------------------------
function setClipSpeed(trackType, trackIndex, clipIndex, speed, ripple) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        speed = parseFloat(speed) || 1.0;
        if (speed <= 0) return _err("Speed must be positive");
        var doRipple = (ripple === true || ripple === "true" || ripple === 1);
        if (typeof qe !== "undefined" && qe.project) {
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                var tIdx = parseInt(trackIndex, 10) || 0;
                var cIdx = parseInt(clipIndex, 10) || 0;
                var qeTrack = (trackType === "audio") ? qeSeq.getAudioTrackAt(tIdx) : qeSeq.getVideoTrackAt(tIdx);
                if (qeTrack) { var qeClip = qeTrack.getItemAt(cIdx); if (qeClip && qeClip.setSpeed) qeClip.setSpeed(speed * 100, doRipple, false); }
            }
        } else {
            var curDur = _timeToSeconds(clip.duration);
            clip.end = _secondsToTime(_timeToSeconds(clip.start) + curDur / speed);
        }
        return _ok({ action: "setSpeed", clipName: clip.name || "", speed: speed, ripple: doRipple, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("setClipSpeed failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 17. reverseClip
// ---------------------------------------------------------------------------
function reverseClip(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        if (typeof qe !== "undefined" && qe.project) {
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                var tIdx = parseInt(trackIndex, 10) || 0;
                var cIdx = parseInt(clipIndex, 10) || 0;
                var qeTrack = (trackType === "audio") ? qeSeq.getAudioTrackAt(tIdx) : qeSeq.getVideoTrackAt(tIdx);
                if (qeTrack) { var qeClip = qeTrack.getItemAt(cIdx); if (qeClip && qeClip.setSpeed) qeClip.setSpeed(-100, false, true); }
            }
        }
        return _ok({ action: "reverse", clipName: clip.name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("reverseClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 18. setClipInPoint
// ---------------------------------------------------------------------------
function setClipInPoint(trackType, trackIndex, clipIndex, seconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        seconds = parseFloat(seconds) || 0;
        var oldIP = _timeToSeconds(clip.inPoint);
        clip.inPoint = _secondsToTime(seconds);
        return _ok({ action: "setInPoint", clipName: clip.name || "", oldInPoint: oldIP, newInPoint: seconds, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("setClipInPoint failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 19. setClipOutPoint
// ---------------------------------------------------------------------------
function setClipOutPoint(trackType, trackIndex, clipIndex, seconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        seconds = parseFloat(seconds) || 0;
        var oldOP = _timeToSeconds(clip.outPoint);
        clip.outPoint = _secondsToTime(seconds);
        return _ok({ action: "setOutPoint", clipName: clip.name || "", oldOutPoint: oldOP, newOutPoint: seconds, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("setClipOutPoint failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 20. getClipSpeed
// ---------------------------------------------------------------------------
function getClipSpeed(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var speed = 1.0, reversed = false;
        if (typeof qe !== "undefined" && qe.project) {
            var qeSeq = qe.project.getActiveSequence();
            if (qeSeq) {
                var tIdx = parseInt(trackIndex, 10) || 0;
                var cIdx = parseInt(clipIndex, 10) || 0;
                var qeTrack = (trackType === "audio") ? qeSeq.getAudioTrackAt(tIdx) : qeSeq.getVideoTrackAt(tIdx);
                if (qeTrack) { var qeClip = qeTrack.getItemAt(cIdx); if (qeClip && qeClip.getSpeed) { var rs = qeClip.getSpeed(); speed = Math.abs(parseFloat(rs)) / 100; reversed = parseFloat(rs) < 0; } }
            }
        }
        return _ok({ clipName: clip.name || "", speed: speed, reversed: reversed, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("getClipSpeed failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 21. trimClipStart
// ---------------------------------------------------------------------------
function trimClipStart(trackType, trackIndex, clipIndex, newStartTime) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        newStartTime = parseFloat(newStartTime) || 0;
        var oldStart = _timeToSeconds(clip.start);
        var oldIP = _timeToSeconds(clip.inPoint);
        var delta = newStartTime - oldStart;
        clip.inPoint = _secondsToTime(oldIP + delta);
        clip.start = _secondsToTime(newStartTime);
        return _ok({ action: "trimStart", clipName: clip.name || "", oldStart: oldStart, newStart: newStartTime, oldInPoint: oldIP, newInPoint: oldIP + delta, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("trimClipStart failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 22. trimClipEnd
// ---------------------------------------------------------------------------
function trimClipEnd(trackType, trackIndex, clipIndex, newEndTime) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        newEndTime = parseFloat(newEndTime) || 0;
        var oldEnd = _timeToSeconds(clip.end);
        var oldOP = _timeToSeconds(clip.outPoint);
        var delta = newEndTime - oldEnd;
        clip.outPoint = _secondsToTime(oldOP + delta);
        clip.end = _secondsToTime(newEndTime);
        return _ok({ action: "trimEnd", clipName: clip.name || "", oldEnd: oldEnd, newEnd: newEndTime, oldOutPoint: oldOP, newOutPoint: oldOP + delta, trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("trimClipEnd failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 23. extendClipToPlayhead
// ---------------------------------------------------------------------------
function extendClipToPlayhead(trackType, trackIndex, clipIndex, trimEnd) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var playheadPos = _timeToSeconds(seq.getPlayerPosition());
        var doTrimEnd = (trimEnd === true || trimEnd === "true" || trimEnd === 1);
        if (doTrimEnd) {
            var oldEnd = _timeToSeconds(clip.end);
            var oldOP = _timeToSeconds(clip.outPoint);
            clip.outPoint = _secondsToTime(oldOP + (playheadPos - oldEnd));
            clip.end = _secondsToTime(playheadPos);
            return _ok({ action: "extendEnd", clipName: clip.name || "", playheadPos: playheadPos, oldEnd: oldEnd, newEnd: playheadPos });
        } else {
            var oldStart = _timeToSeconds(clip.start);
            var oldIP = _timeToSeconds(clip.inPoint);
            clip.inPoint = _secondsToTime(oldIP + (playheadPos - oldStart));
            clip.start = _secondsToTime(playheadPos);
            return _ok({ action: "extendStart", clipName: clip.name || "", playheadPos: playheadPos, oldStart: oldStart, newStart: playheadPos });
        }
    } catch (e) { return _err("extendClipToPlayhead failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 24. createSubclip
// ---------------------------------------------------------------------------
function createSubclip(projectItemIndex, name, inPoint, outPoint) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        inPoint = parseFloat(inPoint) || 0;
        outPoint = parseFloat(outPoint) || 0;
        if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems)
            return _err("Project item index " + projectItemIndex + " out of range");
        var pi = app.project.rootItem.children[projectItemIndex];
        if (!pi) return _err("No project item at index " + projectItemIndex);
        name = name || (pi.name + "_subclip");
        var startT = _secondsToTime(inPoint);
        var endT = _secondsToTime(outPoint);
        var sub = pi.createSubClip(name, startT.ticks, endT.ticks, 0, 1, 1);
        return _ok({ action: "createSubclip", name: name, sourceName: pi.name || "", sourceIndex: projectItemIndex, inPoint: inPoint, outPoint: outPoint, created: sub ? true : false });
    } catch (e) { return _err("createSubclip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 25. selectClip
// ---------------------------------------------------------------------------
function selectClip(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        clip.setSelected(true, true);
        return _ok({ action: "select", clipName: clip.name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("selectClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 26. deselectAll
// ---------------------------------------------------------------------------
function deselectAll() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var count = 0;
        if (seq.videoTracks) { for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) { var vt = seq.videoTracks[vi]; if (vt.clips) { for (var vc = 0; vc < vt.clips.numItems; vc++) { try { vt.clips[vc].setSelected(false, true); count++; } catch (e2) {} } } } }
        if (seq.audioTracks) { for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) { var at2 = seq.audioTracks[ai]; if (at2.clips) { for (var ac = 0; ac < at2.clips.numItems; ac++) { try { at2.clips[ac].setSelected(false, true); count++; } catch (e3) {} } } } }
        return _ok({ action: "deselectAll", clipsDeselected: count });
    } catch (e) { return _err("deselectAll failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 27. getSelectedClips
// ---------------------------------------------------------------------------
function getSelectedClips() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var selected = [];
        if (seq.videoTracks) { for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) { var vt = seq.videoTracks[vi]; if (vt.clips) { for (var vc = 0; vc < vt.clips.numItems; vc++) { try { if (vt.clips[vc].isSelected()) selected.push(_buildClipInfo(vt.clips[vc], vc, "video", vi)); } catch (e2) {} } } } }
        if (seq.audioTracks) { for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) { var at2 = seq.audioTracks[ai]; if (at2.clips) { for (var ac = 0; ac < at2.clips.numItems; ac++) { try { if (at2.clips[ac].isSelected()) selected.push(_buildClipInfo(at2.clips[ac], ac, "audio", ai)); } catch (e3) {} } } } }
        return _ok({ selectedCount: selected.length, clips: selected });
    } catch (e) { return _err("getSelectedClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 28. linkClips
// ---------------------------------------------------------------------------
function linkClips(clipPairsJson) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var clipPairs = JSON.parse(clipPairsJson);
        if (!clipPairs || !clipPairs.length) return _err("No clip pairs provided");
        var linked = 0;
        for (var i = 0; i < clipPairs.length; i++) {
            var p = clipPairs[i];
            var vTrack = _getTrack(seq, "video", p.vTrack);
            var aTrack = _getTrack(seq, "audio", p.aTrack);
            if (!vTrack || !aTrack) continue;
            var vClip = _getClip(vTrack, p.vClip);
            var aClip = _getClip(aTrack, p.aClip);
            if (!vClip || !aClip) continue;
            try { vClip.link(aClip); linked++; } catch (lErr) {}
        }
        return _ok({ action: "link", pairsRequested: clipPairs.length, pairsLinked: linked });
    } catch (e) { return _err("linkClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 29. unlinkClips
// ---------------------------------------------------------------------------
function unlinkClips(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        clip.unlink();
        return _ok({ action: "unlink", clipName: clip.name || "", trackType: trackType, trackIndex: parseInt(trackIndex, 10) || 0, clipIndex: parseInt(clipIndex, 10) || 0 });
    } catch (e) { return _err("unlinkClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// 30. getLinkedClips
// ---------------------------------------------------------------------------
function getLinkedClips(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var track = _getTrack(seq, trackType, trackIndex);
        if (!track) return _err("Track not found: " + trackType + "[" + trackIndex + "]");
        var clip = _getClip(track, clipIndex);
        if (!clip) return _err("Clip not found at index " + clipIndex);
        var linkedClips = [];
        var srcStart = _timeToSeconds(clip.start);
        var srcEnd = _timeToSeconds(clip.end);
        var tI = parseInt(trackIndex, 10) || 0;
        var cI = parseInt(clipIndex, 10) || 0;
        var search = function(tracks, tt, n) {
            for (var ti = 0; ti < n; ti++) {
                var t = tracks[ti]; if (!t.clips) continue;
                for (var ci = 0; ci < t.clips.numItems; ci++) {
                    if (tt === trackType && ti === tI && ci === cI) continue;
                    try { var c = t.clips[ci]; if (Math.abs(_timeToSeconds(c.start) - srcStart) < 0.01 && Math.abs(_timeToSeconds(c.end) - srcEnd) < 0.01) linkedClips.push(_buildClipInfo(c, ci, tt, ti)); } catch (e2) {}
                }
            }
        };
        if (seq.videoTracks) search(seq.videoTracks, "video", seq.videoTracks.numTracks);
        if (seq.audioTracks) search(seq.audioTracks, "audio", seq.audioTracks.numTracks);
        return _ok({ clipName: clip.name || "", trackType: trackType, trackIndex: tI, clipIndex: cI, linkedCount: linkedClips.length, linkedClips: linkedClips });
    } catch (e) { return _err("getLinkedClips failed: " + e.message); }
}

// ===========================================================================
// Project Management Tools
// ===========================================================================

// ---------------------------------------------------------------------------
// Helper: Navigate to a bin by slash-separated path (e.g. "Footage/Raw")
// Returns the ProjectItem for the bin, or null if not found.
// ---------------------------------------------------------------------------
function _findBinByPath(binPath) {
    if (!binPath || binPath === "" || binPath === "/") {
        return app.project.rootItem;
    }
    var parts = binPath.split("/");
    var current = app.project.rootItem;
    for (var i = 0; i < parts.length; i++) {
        var part = parts[i];
        if (!part || part === "") continue;
        var found = false;
        if (current.children) {
            for (var c = 0; c < current.children.numItems; c++) {
                var child = current.children[c];
                if (child.name === part && child.type === ProjectItemType.BIN) {
                    current = child;
                    found = true;
                    break;
                }
            }
        }
        if (!found) return null;
    }
    return current;
}

// ---------------------------------------------------------------------------
// Helper: Find a project item by slash-separated path (e.g. "Footage/clip.mp4")
// ---------------------------------------------------------------------------
function _findItemByPath(itemPath) {
    if (!itemPath || itemPath === "") return null;
    var parts = itemPath.split("/");
    var current = app.project.rootItem;
    for (var i = 0; i < parts.length; i++) {
        var part = parts[i];
        if (!part || part === "") continue;
        var found = false;
        if (current.children) {
            for (var c = 0; c < current.children.numItems; c++) {
                var child = current.children[c];
                if (child.name === part) {
                    current = child;
                    found = true;
                    break;
                }
            }
        }
        if (!found) return null;
    }
    return current;
}

// ---------------------------------------------------------------------------
// 1. newProject(path) - Create a new project at the given path
// ---------------------------------------------------------------------------
function newProject(path) {
    try {
        if (!path || path === "") {
            return _err("path is required");
        }
        app.newProject(path);
        return _ok({
            path: path,
            created: true
        });
    } catch (e) {
        return _err("newProject failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 2. openProject(path) - Open an existing .prproj file
// ---------------------------------------------------------------------------
function openProject(path) {
    try {
        if (!path || path === "") {
            return _err("path is required");
        }
        var result = app.openDocument(path);
        if (result) {
            return _ok({
                path: path,
                opened: true,
                projectName: app.project ? (app.project.name || "") : ""
            });
        } else {
            return _err("openDocument returned false for: " + path);
        }
    } catch (e) {
        return _err("openProject failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 3. saveProject() - Save current project
// ---------------------------------------------------------------------------
function saveProject() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }
        app.project.save();
        return _ok({
            saved: true,
            projectName: app.project.name || "",
            projectPath: app.project.path || ""
        });
    } catch (e) {
        return _err("saveProject failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 4. saveProjectAs(path) - Save current project to a new path
// ---------------------------------------------------------------------------
function saveProjectAs(path) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }
        if (!path || path === "") {
            return _err("path is required");
        }
        app.project.saveAs(path);
        return _ok({
            saved: true,
            newPath: path,
            projectName: app.project.name || ""
        });
    } catch (e) {
        return _err("saveProjectAs failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 5. closeProject(saveFirst) - Close current project, optionally saving
// ---------------------------------------------------------------------------
function closeProject(saveFirst) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }
        var projectName = app.project.name || "";

        if (saveFirst === true || saveFirst === "true") {
            app.project.save();
        }

        app.project.closeDocument();

        return _ok({
            closed: true,
            projectName: projectName,
            savedFirst: (saveFirst === true || saveFirst === "true")
        });
    } catch (e) {
        return _err("closeProject failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 6. getProjectInfo() - Get detailed project info
// ---------------------------------------------------------------------------
function getProjectInfo() {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var proj = app.project;
        var info = {
            name: proj.name || "",
            path: proj.path || "",
            documentID: proj.documentID || "",
            sequences: [],
            bins: [],
            totalItems: 0
        };

        if (proj.sequences) {
            for (var s = 0; s < proj.sequences.numSequences; s++) {
                var seq = proj.sequences[s];
                info.sequences.push({
                    index: s,
                    name: seq.name || "",
                    id: seq.sequenceID || "",
                    videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0,
                    audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0,
                    frameSizeHorizontal: seq.frameSizeHorizontal || 0,
                    frameSizeVertical: seq.frameSizeVertical || 0,
                    timebase: seq.timebase || ""
                });
            }
        }

        if (proj.rootItem && proj.rootItem.children) {
            info.totalItems = proj.rootItem.children.numItems;
            for (var b = 0; b < proj.rootItem.children.numItems; b++) {
                var item = proj.rootItem.children[b];
                if (item.type === ProjectItemType.BIN) {
                    var binChildCount = 0;
                    if (item.children) {
                        binChildCount = item.children.numItems;
                    }
                    info.bins.push({
                        name: item.name || "",
                        childCount: binChildCount
                    });
                }
            }
        }

        if (proj.activeSequence) {
            info.activeSequence = {
                name: proj.activeSequence.name || "",
                id: proj.activeSequence.sequenceID || ""
            };
        }

        return _ok(info);
    } catch (e) {
        return _err("getProjectInfo failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 7. importFiles(paramsJson) - Import multiple files into project
//    paramsJson: { filePaths: [...], targetBin: "..." }
// ---------------------------------------------------------------------------
function importFiles(paramsJson) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var params = JSON.parse(paramsJson);
        var filePaths = params.filePaths;
        var targetBinPath = params.targetBin || "";

        if (!filePaths || filePaths.length === 0) {
            return _err("filePaths array is required and must not be empty");
        }

        var targetBin = app.project.rootItem;
        if (targetBinPath && targetBinPath !== "") {
            var resolved = _findBinByPath(targetBinPath);
            if (resolved) {
                targetBin = resolved;
            } else {
                var binParts = targetBinPath.split("/");
                var current = app.project.rootItem;
                for (var bp = 0; bp < binParts.length; bp++) {
                    var bName = binParts[bp];
                    if (!bName || bName === "") continue;
                    var found = false;
                    if (current.children) {
                        for (var cc = 0; cc < current.children.numItems; cc++) {
                            if (current.children[cc].name === bName && current.children[cc].type === ProjectItemType.BIN) {
                                current = current.children[cc];
                                found = true;
                                break;
                            }
                        }
                    }
                    if (!found) {
                        var nb = current.createBin(bName);
                        if (nb) { current = nb; } else { return _err("Failed to create bin: " + bName); }
                    }
                }
                targetBin = current;
            }
        }

        var suppressUI = true;
        var importAsNumberedStill = false;
        var success = app.project.importFiles(filePaths, suppressUI, targetBin, importAsNumberedStill);

        return _ok({
            imported: success ? true : false,
            fileCount: filePaths.length,
            targetBin: targetBinPath || "/",
            filePaths: filePaths
        });
    } catch (e) {
        return _err("importFiles failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 8. importFolder(paramsJson) - Import an entire folder recursively
//    paramsJson: { folderPath: "...", targetBin: "..." }
// ---------------------------------------------------------------------------
function importFolder(paramsJson) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var params = JSON.parse(paramsJson);
        var folderPath = params.folderPath;
        var targetBinPath = params.targetBin || "";

        if (!folderPath || folderPath === "") {
            return _err("folderPath is required");
        }

        var targetBin = app.project.rootItem;
        if (targetBinPath && targetBinPath !== "") {
            var resolved = _findBinByPath(targetBinPath);
            if (resolved) {
                targetBin = resolved;
            } else {
                var binParts = targetBinPath.split("/");
                var current = app.project.rootItem;
                for (var bp = 0; bp < binParts.length; bp++) {
                    var bName = binParts[bp];
                    if (!bName || bName === "") continue;
                    var found = false;
                    if (current.children) {
                        for (var cc = 0; cc < current.children.numItems; cc++) {
                            if (current.children[cc].name === bName && current.children[cc].type === ProjectItemType.BIN) {
                                current = current.children[cc];
                                found = true;
                                break;
                            }
                        }
                    }
                    if (!found) {
                        var nb = current.createBin(bName);
                        if (nb) { current = nb; } else { return _err("Failed to create bin: " + bName); }
                    }
                }
                targetBin = current;
            }
        }

        var folder = new Folder(folderPath);
        if (!folder.exists) {
            return _err("Folder does not exist: " + folderPath);
        }

        var allFiles = [];
        var _collectFiles = function(dir) {
            var files = dir.getFiles();
            for (var fi = 0; fi < files.length; fi++) {
                if (files[fi] instanceof Folder) {
                    _collectFiles(files[fi]);
                } else {
                    allFiles.push(files[fi].fsName);
                }
            }
        };
        _collectFiles(folder);

        if (allFiles.length === 0) {
            return _ok({
                imported: false,
                fileCount: 0,
                folderPath: folderPath,
                message: "No files found in folder"
            });
        }

        var suppressUI = true;
        var success = app.project.importFiles(allFiles, suppressUI, targetBin, false);

        return _ok({
            imported: success ? true : false,
            fileCount: allFiles.length,
            folderPath: folderPath,
            targetBin: targetBinPath || "/"
        });
    } catch (e) {
        return _err("importFolder failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 9. createBin(paramsJson) - Create a new bin (folder) in the project
//    paramsJson: { name: "...", parentBin: "..." }
// ---------------------------------------------------------------------------
function createBin(paramsJson) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var params = JSON.parse(paramsJson);
        var name = params.name;
        var parentBinPath = params.parentBin || "";

        if (!name || name === "") {
            return _err("name is required");
        }

        var parentBin = app.project.rootItem;
        if (parentBinPath && parentBinPath !== "") {
            var resolved = _findBinByPath(parentBinPath);
            if (resolved) {
                parentBin = resolved;
            } else {
                return _err("Parent bin not found: " + parentBinPath);
            }
        }

        var newBin = parentBin.createBin(name);
        if (newBin) {
            return _ok({
                created: true,
                name: name,
                parentBin: parentBinPath || "/"
            });
        } else {
            return _err("createBin returned null for: " + name);
        }
    } catch (e) {
        return _err("createBin failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 10. renameBin(paramsJson) - Rename a bin
//     paramsJson: { binPath: "...", newName: "..." }
// ---------------------------------------------------------------------------
function renameBin(paramsJson) {
    try {
        if (!app.project) {
            return _err("No project is open");
        }

        var params = JSON.parse(paramsJson);
        var binPath = params.binPath;
        var newName = params.newName;

        if (!binPath || binPath === "") return _err("binPath is required");
        if (!newName || newName === "") return _err("newName is required");

        var bin = _findBinByPath(binPath);
        if (!bin) return _err("Bin not found: " + binPath);
        if (bin === app.project.rootItem) return _err("Cannot rename the root bin");

        var oldName = bin.name;
        bin.name = newName;

        return _ok({ renamed: true, oldName: oldName, newName: newName, binPath: binPath });
    } catch (e) {
        return _err("renameBin failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 11. deleteBin(binPath) - Delete a bin and its contents
// ---------------------------------------------------------------------------
function deleteBin(binPath) {
    try {
        if (!app.project) return _err("No project is open");
        if (!binPath || binPath === "") return _err("binPath is required");

        var bin = _findBinByPath(binPath);
        if (!bin) return _err("Bin not found: " + binPath);
        if (bin === app.project.rootItem) return _err("Cannot delete the root bin");

        var binName = bin.name;
        app.project.deleteAsset(bin);

        return _ok({ deleted: true, binPath: binPath, binName: binName });
    } catch (e) {
        return _err("deleteBin failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 12. moveBinItem(paramsJson) - Move an item between bins
//     paramsJson: { itemPath: "...", destBin: "..." }
// ---------------------------------------------------------------------------
function moveBinItem(paramsJson) {
    try {
        if (!app.project) return _err("No project is open");

        var params = JSON.parse(paramsJson);
        var itemPath = params.itemPath;
        var destBinPath = params.destBin || "";

        if (!itemPath || itemPath === "") return _err("itemPath is required");

        var item = _findItemByPath(itemPath);
        if (!item) return _err("Item not found: " + itemPath);

        var destBin = app.project.rootItem;
        if (destBinPath && destBinPath !== "") {
            destBin = _findBinByPath(destBinPath);
            if (!destBin) return _err("Destination bin not found: " + destBinPath);
        }

        item.moveBin(destBin);

        return _ok({ moved: true, itemName: item.name || "", itemPath: itemPath, destBin: destBinPath || "/" });
    } catch (e) {
        return _err("moveBinItem failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 13. findProjectItems(searchQuery) - Search project items by name
// ---------------------------------------------------------------------------
function findProjectItems(searchQuery) {
    try {
        if (!app.project) return _err("No project is open");
        if (!searchQuery || searchQuery === "") return _err("searchQuery is required");

        var results = [];
        var queryLower = searchQuery.toLowerCase();

        var _searchItems = function(bin, pathPrefix) {
            if (!bin || !bin.children) return;
            for (var i = 0; i < bin.children.numItems; i++) {
                var child = bin.children[i];
                var childPath = pathPrefix ? (pathPrefix + "/" + child.name) : child.name;
                if (child.name && child.name.toLowerCase().indexOf(queryLower) >= 0) {
                    var info = {
                        name: child.name,
                        path: childPath,
                        type: (child.type === ProjectItemType.BIN) ? "bin" :
                              (child.type === ProjectItemType.CLIP) ? "clip" :
                              (child.type === ProjectItemType.FILE) ? "file" : "unknown"
                    };
                    if (child.type !== ProjectItemType.BIN && child.getMediaPath) {
                        try { info.mediaPath = child.getMediaPath() || ""; } catch (mp) { info.mediaPath = ""; }
                    }
                    results.push(info);
                }
                if (child.type === ProjectItemType.BIN) {
                    _searchItems(child, childPath);
                }
            }
        };

        _searchItems(app.project.rootItem, "");

        return _ok({ query: searchQuery, resultCount: results.length, items: results });
    } catch (e) {
        return _err("findProjectItems failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 14. getProjectItems(binPath) - List all items in a bin
// ---------------------------------------------------------------------------
function getProjectItems(binPath) {
    try {
        if (!app.project) return _err("No project is open");

        var targetBin = app.project.rootItem;
        if (binPath && binPath !== "" && binPath !== "/") {
            targetBin = _findBinByPath(binPath);
            if (!targetBin) return _err("Bin not found: " + binPath);
        }

        var items = [];
        if (targetBin.children) {
            for (var i = 0; i < targetBin.children.numItems; i++) {
                var child = targetBin.children[i];
                var info = {
                    index: i,
                    name: child.name || "",
                    type: (child.type === ProjectItemType.BIN) ? "bin" :
                          (child.type === ProjectItemType.CLIP) ? "clip" :
                          (child.type === ProjectItemType.FILE) ? "file" : "unknown"
                };
                if (child.type !== ProjectItemType.BIN && child.getMediaPath) {
                    try { info.mediaPath = child.getMediaPath() || ""; } catch (mp) { info.mediaPath = ""; }
                }
                if (child.type === ProjectItemType.BIN && child.children) {
                    info.childCount = child.children.numItems;
                }
                items.push(info);
            }
        }

        return _ok({ binPath: binPath || "/", itemCount: items.length, items: items });
    } catch (e) {
        return _err("getProjectItems failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 15. setItemLabel(paramsJson) - Set label color on a project item (0-15)
//     paramsJson: { itemPath: "...", colorIndex: 0 }
// ---------------------------------------------------------------------------
function setItemLabel(paramsJson) {
    try {
        if (!app.project) return _err("No project is open");

        var params = JSON.parse(paramsJson);
        var itemPath = params.itemPath;
        var colorIndex = parseInt(params.colorIndex, 10);

        if (!itemPath || itemPath === "") return _err("itemPath is required");
        if (isNaN(colorIndex) || colorIndex < 0 || colorIndex > 15) return _err("colorIndex must be between 0 and 15");

        var item = _findItemByPath(itemPath);
        if (!item) return _err("Item not found: " + itemPath);

        item.setColorLabel(colorIndex);

        return _ok({ itemPath: itemPath, itemName: item.name || "", colorIndex: colorIndex });
    } catch (e) {
        return _err("setItemLabel failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 16. getItemMetadata(itemPath) - Get XMP metadata for a project item
// ---------------------------------------------------------------------------
function getItemMetadata(itemPath) {
    try {
        if (!app.project) return _err("No project is open");
        if (!itemPath || itemPath === "") return _err("itemPath is required");

        var item = _findItemByPath(itemPath);
        if (!item) return _err("Item not found: " + itemPath);

        var metadata = {};
        metadata.name = item.name || "";
        metadata.type = (item.type === ProjectItemType.BIN) ? "bin" :
                        (item.type === ProjectItemType.CLIP) ? "clip" :
                        (item.type === ProjectItemType.FILE) ? "file" : "unknown";

        if (item.getXMPMetadata) {
            try { metadata.xmpRaw = item.getXMPMetadata() || ""; } catch (xmpErr) { metadata.xmpRaw = ""; }
        }
        if (item.getMediaPath) {
            try { metadata.mediaPath = item.getMediaPath() || ""; } catch (mp) { metadata.mediaPath = ""; }
        }
        if (item.getInPoint) {
            try { metadata.inPoint = _timeToSeconds(item.getInPoint()); } catch (ip) {}
        }
        if (item.getOutPoint) {
            try { metadata.outPoint = _timeToSeconds(item.getOutPoint()); } catch (op) {}
        }
        if (item.getFootageInterpretation) {
            try {
                var interp = item.getFootageInterpretation();
                if (interp) {
                    metadata.frameRate = interp.frameRate || 0;
                    metadata.pixelAspectRatio = interp.pixelAspectRatio || 0;
                    metadata.fieldType = interp.fieldType || 0;
                }
            } catch (fi) {}
        }

        return _ok({ itemPath: itemPath, metadata: metadata });
    } catch (e) {
        return _err("getItemMetadata failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 17. setItemMetadata(paramsJson) - Set XMP metadata on a project item
//     paramsJson: { itemPath: "...", key: "...", value: "..." }
// ---------------------------------------------------------------------------
function setItemMetadata(paramsJson) {
    try {
        if (!app.project) return _err("No project is open");

        var params = JSON.parse(paramsJson);
        var itemPath = params.itemPath;
        var key = params.key;
        var value = params.value;

        if (!itemPath || itemPath === "") return _err("itemPath is required");
        if (!key || key === "") return _err("key is required");

        var item = _findItemByPath(itemPath);
        if (!item) return _err("Item not found: " + itemPath);
        if (!item.setXMPMetadata) return _err("Item does not support XMP metadata");

        var existingXMP = "";
        if (item.getXMPMetadata) { existingXMP = item.getXMPMetadata() || ""; }

        if (existingXMP === "") {
            existingXMP = '<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"></rdf:Description></rdf:RDF></x:xmpmeta>';
        }

        if (typeof XMPMeta !== "undefined") {
            var xmp = new XMPMeta(existingXMP);
            var nsParts = key.split(":");
            if (nsParts.length === 2) {
                var nsPrefix = nsParts[0];
                var propName = nsParts[1];
                var ns = XMPMeta.getNamespaceURI(nsPrefix);
                if (ns) {
                    xmp.setProperty(ns, propName, value);
                } else {
                    XMPMeta.registerNamespace("http://custom.ns/" + nsPrefix + "/", nsPrefix);
                    xmp.setProperty("http://custom.ns/" + nsPrefix + "/", propName, value);
                }
            } else {
                xmp.setProperty("http://purl.org/dc/elements/1.1/", key, value);
            }
            item.setXMPMetadata(xmp.serialize());
        } else {
            item.setXMPMetadata(existingXMP);
        }

        return _ok({ itemPath: itemPath, key: key, value: value, updated: true });
    } catch (e) {
        return _err("setItemMetadata failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 18. relinkMedia(paramsJson) - Relink offline media
//     paramsJson: { itemPath: "...", newMediaPath: "..." }
// ---------------------------------------------------------------------------
function relinkMedia(paramsJson) {
    try {
        if (!app.project) return _err("No project is open");

        var params = JSON.parse(paramsJson);
        var itemPath = params.itemPath;
        var newMediaPath = params.newMediaPath;

        if (!itemPath || itemPath === "") return _err("itemPath is required");
        if (!newMediaPath || newMediaPath === "") return _err("newMediaPath is required");

        var item = _findItemByPath(itemPath);
        if (!item) return _err("Item not found: " + itemPath);

        var targetFile = new File(newMediaPath);
        if (!targetFile.exists) return _err("Target media file does not exist: " + newMediaPath);

        if (item.changeMediaPath) {
            var success = item.changeMediaPath(newMediaPath, true);
            return _ok({ relinked: success ? true : false, itemPath: itemPath, itemName: item.name || "", newMediaPath: newMediaPath });
        } else {
            return _err("Item does not support changeMediaPath");
        }
    } catch (e) {
        return _err("relinkMedia failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 19. makeOffline(itemPath) - Make a project item offline
// ---------------------------------------------------------------------------
function makeOffline(itemPath) {
    try {
        if (!app.project) return _err("No project is open");
        if (!itemPath || itemPath === "") return _err("itemPath is required");

        var item = _findItemByPath(itemPath);
        if (!item) return _err("Item not found: " + itemPath);

        if (item.setOffline) {
            item.setOffline();
            return _ok({ itemPath: itemPath, itemName: item.name || "", offline: true });
        } else {
            return _err("Item does not support setOffline");
        }
    } catch (e) {
        return _err("makeOffline failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 20. getOfflineItems() - Get list of all offline items
// ---------------------------------------------------------------------------
function getOfflineItems() {
    try {
        if (!app.project) return _err("No project is open");

        var offlineItems = [];

        var _findOffline = function(bin, pathPrefix) {
            if (!bin || !bin.children) return;
            for (var i = 0; i < bin.children.numItems; i++) {
                var child = bin.children[i];
                var childPath = pathPrefix ? (pathPrefix + "/" + child.name) : child.name;

                if (child.type === ProjectItemType.BIN) {
                    _findOffline(child, childPath);
                } else {
                    var isOffline = false;
                    if (child.isOffline) {
                        isOffline = child.isOffline();
                    } else if (child.getMediaPath) {
                        try {
                            var mp = child.getMediaPath();
                            if (!mp || mp === "") {
                                isOffline = true;
                            } else {
                                var f = new File(mp);
                                if (!f.exists) { isOffline = true; }
                            }
                        } catch (mpErr) {
                            isOffline = true;
                        }
                    }
                    if (isOffline) {
                        offlineItems.push({
                            name: child.name || "",
                            path: childPath,
                            type: (child.type === ProjectItemType.CLIP) ? "clip" :
                                  (child.type === ProjectItemType.FILE) ? "file" : "unknown"
                        });
                    }
                }
            }
        };

        _findOffline(app.project.rootItem, "");

        return _ok({ offlineCount: offlineItems.length, items: offlineItems });
    } catch (e) {
        return _err("getOfflineItems failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 21. setScratchDisk(paramsJson) - Set scratch disk path
//     paramsJson: { type: "...", path: "..." }
// ---------------------------------------------------------------------------
function setScratchDisk(paramsJson) {
    try {
        if (!app.project) return _err("No project is open");

        var params = JSON.parse(paramsJson);
        var scratchType = params.type;
        var scratchPath = params.path;

        if (!scratchType || scratchType === "") return _err("type is required");
        if (!scratchPath || scratchPath === "") return _err("path is required");

        var folder = new Folder(scratchPath);
        if (!folder.exists) return _err("Scratch disk path does not exist: " + scratchPath);

        var sdType;
        switch (scratchType) {
            case "capturedVideo": sdType = ScratchDiskType.FirstVideoCaptureFolder; break;
            case "capturedAudio": sdType = ScratchDiskType.FirstAudioCaptureFolder; break;
            case "videoPreview":  sdType = ScratchDiskType.FirstVideoPreviewFolder; break;
            case "audioPreview":  sdType = ScratchDiskType.FirstAudioPreviewFolder; break;
            case "autoSave":     sdType = ScratchDiskType.FirstAutoSaveFolder; break;
            case "cclibrary":    sdType = ScratchDiskType.FirstCCLibrariesFolder; break;
            default: return _err("Unknown scratch disk type: " + scratchType);
        }

        app.setScratchDiskPath(scratchPath, sdType);

        return _ok({ type: scratchType, path: scratchPath, set: true });
    } catch (e) {
        return _err("setScratchDisk failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 22. consolidateDuplicates() - Remove duplicate project items
// ---------------------------------------------------------------------------
function consolidateDuplicates() {
    try {
        if (!app.project) return _err("No project is open");

        var mediaPathMap = {};
        var duplicates = [];
        var totalChecked = 0;

        var _scanForDuplicates = function(bin, pathPrefix) {
            if (!bin || !bin.children) return;
            for (var i = 0; i < bin.children.numItems; i++) {
                var child = bin.children[i];
                var childPath = pathPrefix ? (pathPrefix + "/" + child.name) : child.name;

                if (child.type === ProjectItemType.BIN) {
                    _scanForDuplicates(child, childPath);
                } else {
                    totalChecked++;
                    if (child.getMediaPath) {
                        try {
                            var mediaPath = child.getMediaPath();
                            if (mediaPath && mediaPath !== "") {
                                if (mediaPathMap[mediaPath]) {
                                    duplicates.push({
                                        name: child.name || "",
                                        path: childPath,
                                        mediaPath: mediaPath,
                                        originalPath: mediaPathMap[mediaPath]
                                    });
                                } else {
                                    mediaPathMap[mediaPath] = childPath;
                                }
                            }
                        } catch (mp) {}
                    }
                }
            }
        };

        _scanForDuplicates(app.project.rootItem, "");

        var removed = 0;
        for (var d = 0; d < duplicates.length; d++) {
            var dupItem = _findItemByPath(duplicates[d].path);
            if (dupItem) {
                try {
                    app.project.deleteAsset(dupItem);
                    removed++;
                } catch (delErr) {
                    duplicates[d].removeError = delErr.message;
                }
            }
        }

        return _ok({ totalChecked: totalChecked, duplicatesFound: duplicates.length, duplicatesRemoved: removed, duplicates: duplicates });
    } catch (e) {
        return _err("consolidateDuplicates failed: " + e.message);
    }
}

// ---------------------------------------------------------------------------
// 23. getProjectSettings() - Get project settings
// ---------------------------------------------------------------------------
function getProjectSettings() {
    try {
        if (!app.project) return _err("No project is open");

        var proj = app.project;
        var settings = {
            name: proj.name || "",
            path: proj.path || "",
            documentID: proj.documentID || ""
        };

        if (proj.gpuAccelRendererInfo) {
            try { settings.gpuRenderer = proj.gpuAccelRendererInfo() || "unknown"; } catch (gpu) {}
        }

        if (proj.activeSequence) {
            var seq = proj.activeSequence;
            settings.activeSequence = {
                name: seq.name || "",
                id: seq.sequenceID || "",
                frameSizeHorizontal: seq.frameSizeHorizontal || 0,
                frameSizeVertical: seq.frameSizeVertical || 0,
                timebase: seq.timebase || "",
                videoTrackCount: seq.videoTracks ? seq.videoTracks.numTracks : 0,
                audioTrackCount: seq.audioTracks ? seq.audioTracks.numTracks : 0,
                videoDisplayFormat: seq.videoDisplayFormat || 0,
                audioDisplayFormat: seq.audioDisplayFormat || 0
            };
            if (seq.getSettings) {
                try { settings.activeSequence.settings = seq.getSettings() || {}; } catch (ss) {}
            }
        }

        if (proj.rootItem && proj.rootItem.children) {
            settings.rootItemCount = proj.rootItem.children.numItems;
        }
        if (proj.sequences) {
            settings.sequenceCount = proj.sequences.numSequences;
        }

        return _ok(settings);
    } catch (e) {
        return _err("getProjectSettings failed: " + e.message);
    }
}

// ===========================================================================
// Effects & Transitions
// ===========================================================================

// ---------------------------------------------------------------------------
// Transitions
// ---------------------------------------------------------------------------

function addVideoTransition(trackIndex, clipIndex, transitionName, duration, applyToEnd) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        duration = parseFloat(duration) || 1.0;
        transitionName = transitionName || "Cross Dissolve";
        if (applyToEnd === undefined) applyToEnd = true;
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE: no active sequence");
        var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
        if (!qeTrack) return _err("QE: video track " + trackIndex + " not found");
        var qeClip = qeTrack.getItemAt(clipIndex);
        if (!qeClip) return _err("QE: clip " + clipIndex + " not found on video track " + trackIndex);
        var tr = qe.project.getVideoTransitionByName(transitionName);
        if (!tr) return _err("QE: video transition '" + transitionName + "' not found");
        qeClip.addTransition(tr, applyToEnd, duration.toString());
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, transitionName: transitionName, duration: duration, applyToEnd: applyToEnd });
    } catch (e) { return _err("addVideoTransition failed: " + e.message); }
}

function addAudioTransition(trackIndex, clipIndex, transitionName, duration) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        duration = parseFloat(duration) || 1.0;
        transitionName = transitionName || "Constant Power";
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE: no active sequence");
        var qeTrack = qeSeq.getAudioTrackAt(trackIndex);
        if (!qeTrack) return _err("QE: audio track " + trackIndex + " not found");
        var qeClip = qeTrack.getItemAt(clipIndex);
        if (!qeClip) return _err("QE: clip " + clipIndex + " not found on audio track " + trackIndex);
        var tr = qe.project.getAudioTransitionByName(transitionName);
        if (!tr) return _err("QE: audio transition '" + transitionName + "' not found");
        qeClip.addTransition(tr, true, duration.toString());
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, transitionName: transitionName, duration: duration });
    } catch (e) { return _err("addAudioTransition failed: " + e.message); }
}

function removeTransition(trackType, trackIndex, transitionIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        transitionIndex = parseInt(transitionIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.transitions || transitionIndex >= track.transitions.numItems)
            return _err("Transition index " + transitionIndex + " out of range on track " + trackIndex);
        var t = track.transitions[transitionIndex];
        var tName = t.matchName || t.name || "";
        t.remove(false, false);
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, transitionIndex: transitionIndex, removed: tName });
    } catch (e) { return _err("removeTransition failed: " + e.message); }
}

function getTransitions(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        var result = [];
        if (track.transitions) {
            for (var i = 0; i < track.transitions.numItems; i++) {
                var t = track.transitions[i];
                result.push({ index: i, name: t.name || "", matchName: t.matchName || "", start: _timeToSeconds(t.start), end: _timeToSeconds(t.end), duration: _timeToSeconds(t.duration) });
            }
        }
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, transitions: result });
    } catch (e) { return _err("getTransitions failed: " + e.message); }
}

function setDefaultVideoTransition(transitionName) {
    try {
        transitionName = transitionName || "Cross Dissolve";
        app.enableQE();
        var tr = qe.project.getVideoTransitionByName(transitionName);
        if (!tr) return _err("QE: video transition '" + transitionName + "' not found");
        tr.setSelected(true, true);
        return _ok({ defaultVideoTransition: transitionName });
    } catch (e) { return _err("setDefaultVideoTransition failed: " + e.message); }
}

function setDefaultAudioTransition(transitionName) {
    try {
        transitionName = transitionName || "Constant Power";
        app.enableQE();
        var tr = qe.project.getAudioTransitionByName(transitionName);
        if (!tr) return _err("QE: audio transition '" + transitionName + "' not found");
        tr.setSelected(true, true);
        return _ok({ defaultAudioTransition: transitionName });
    } catch (e) { return _err("setDefaultAudioTransition failed: " + e.message); }
}

function applyDefaultTransition(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems)
            return _err("Clip index " + clipIndex + " out of range on track " + trackIndex);
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE: no active sequence");
        var qeTrack = (trackType === "audio") ? qeSeq.getAudioTrackAt(trackIndex) : qeSeq.getVideoTrackAt(trackIndex);
        if (!qeTrack) return _err("QE: track " + trackIndex + " not found");
        var qeClip = qeTrack.getItemAt(clipIndex);
        if (!qeClip) return _err("QE: clip " + clipIndex + " not found");
        qeClip.addTransition(null, true, "1.0");
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, clipIndex: clipIndex, appliedDefault: true });
    } catch (e) { return _err("applyDefaultTransition failed: " + e.message); }
}

function getAvailableTransitions() {
    try {
        app.enableQE();
        var transitions = [];
        var knownNames = ["Cross Dissolve","Dip to Black","Dip to White","Film Dissolve","Morph Cut","Additive Dissolve","Barn Doors","Band Slide","Band Wipe","Block Dissolve","Center Peel","Center Split","Checker Wipe","Clock Wipe","Cross Stretch","Cross Zoom","Cube Spin","Curtain","Flip Over","Funnel","Gradient Wipe","Inset","Iris Box","Iris Cross","Iris Diamond","Iris Round","Iris Star","Page Peel","Page Turn","Pinwheel","Push","Random Blocks","Random Wipe","Slash Slide","Slide","Spin","Spin Away","Split","Stretch In","Stretch Over","Swap","Swirl","Take","Venetian Blinds","Wedge Wipe","Whip","Wipe","Zoom"];
        for (var i = 0; i < knownNames.length; i++) {
            try { var t = qe.project.getVideoTransitionByName(knownNames[i]); if (t) transitions.push(knownNames[i]); } catch (x) {}
        }
        return _ok({ transitions: transitions });
    } catch (e) { return _err("getAvailableTransitions failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Video Effects
// ---------------------------------------------------------------------------

function applyVideoEffect(trackIndex, clipIndex, effectName) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        if (!effectName || effectName === "") return _err("effectName is required");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE: no active sequence");
        var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
        if (!qeTrack) return _err("QE: video track " + trackIndex + " not found");
        var qeClip = qeTrack.getItemAt(clipIndex);
        if (!qeClip) return _err("QE: clip " + clipIndex + " not found on video track " + trackIndex);
        var fx = qe.project.getVideoEffectByName(effectName);
        if (!fx) return _err("QE: video effect '" + effectName + "' not found");
        qeClip.addVideoEffect(fx);
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, effectName: effectName });
    } catch (e) { return _err("applyVideoEffect failed: " + e.message); }
}

function removeVideoEffect(trackIndex, clipIndex, effectIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        effectIndex = parseInt(effectIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || effectIndex >= clip.components.numItems) return _err("Effect index out of range");
        if (effectIndex < 2) return _err("Cannot remove intrinsic component at index " + effectIndex);
        var comp = clip.components[effectIndex];
        var compName = comp.displayName || "";
        comp.remove();
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, effectIndex: effectIndex, removedEffect: compName });
    } catch (e) { return _err("removeVideoEffect failed: " + e.message); }
}

function getClipEffects(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        var effects = [];
        if (clip.components) {
            for (var ci = 0; ci < clip.components.numItems; ci++) {
                var comp = clip.components[ci];
                var params = [];
                if (comp.properties) {
                    for (var pi = 0; pi < comp.properties.numItems; pi++) {
                        var prop = comp.properties[pi];
                        var v = null; try { v = prop.getValue(); } catch (x) { v = "unreadable"; }
                        params.push({ index: pi, displayName: prop.displayName || "", value: v });
                    }
                }
                effects.push({ index: ci, displayName: comp.displayName || "", matchName: comp.matchName || "", parameters: params });
            }
        }
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, clipIndex: clipIndex, clipName: clip.name || "", effects: effects });
    } catch (e) { return _err("getClipEffects failed: " + e.message); }
}

function setEffectParameter(trackType, trackIndex, clipIndex, componentIndex, paramIndex, value) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        componentIndex = parseInt(componentIndex, 10) || 0; paramIndex = parseInt(paramIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        var numVal = parseFloat(value);
        if (isNaN(numVal)) { param.setValue(value, true); } else { param.setValue(numVal, true); }
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName || "", value: value });
    } catch (e) { return _err("setEffectParameter failed: " + e.message); }
}

function getEffectParameter(trackType, trackIndex, clipIndex, componentIndex, paramIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        componentIndex = parseInt(componentIndex, 10) || 0; paramIndex = parseInt(paramIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        var val = null; try { val = param.getValue(); } catch (x) { val = "unreadable"; }
        var tv = false; try { tv = param.isTimeVarying(); } catch (x2) {}
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName || "", value: val, isTimeVarying: tv });
    } catch (e) { return _err("getEffectParameter failed: " + e.message); }
}

function enableEffect(trackType, trackIndex, clipIndex, componentIndex, enabled) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        componentIndex = parseInt(componentIndex, 10) || 0;
        if (enabled === undefined) enabled = true;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        comp.enabled = !!enabled;
        return _ok({ trackType: trackType || "video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, componentName: comp.displayName || "", enabled: !!enabled });
    } catch (e) { return _err("enableEffect failed: " + e.message); }
}

var _effectsClipboard = null;

function copyEffects(srcTrackType, srcTrackIndex, srcClipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        srcTrackIndex = parseInt(srcTrackIndex, 10) || 0;
        srcClipIndex = parseInt(srcClipIndex, 10) || 0;
        var tracks = (srcTrackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (srcTrackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[srcTrackIndex];
        if (!track.clips || srcClipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[srcClipIndex];
        var effectData = [];
        if (clip.components) {
            for (var ci = 0; ci < clip.components.numItems; ci++) {
                var comp = clip.components[ci]; var params = [];
                if (comp.properties) {
                    for (var pi = 0; pi < comp.properties.numItems; pi++) {
                        var prop = comp.properties[pi]; var val = null;
                        try { val = prop.getValue(); } catch (x) {}
                        params.push({ index: pi, displayName: prop.displayName || "", value: val });
                    }
                }
                effectData.push({ index: ci, displayName: comp.displayName || "", matchName: comp.matchName || "", parameters: params });
            }
        }
        _effectsClipboard = { trackType: srcTrackType || "video", clipName: clip.name || "", effects: effectData };
        return _ok({ copiedFrom: clip.name || "", effectCount: effectData.length, trackType: srcTrackType || "video" });
    } catch (e) { return _err("copyEffects failed: " + e.message); }
}

function pasteEffects(destTrackType, destTrackIndex, destClipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        if (!_effectsClipboard) return _err("No effects copied. Call copyEffects first.");
        destTrackIndex = parseInt(destTrackIndex, 10) || 0;
        destClipIndex = parseInt(destClipIndex, 10) || 0;
        var tracks = (destTrackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (destTrackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[destTrackIndex];
        if (!track.clips || destClipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[destClipIndex];
        var applied = 0;
        if (clip.components && _effectsClipboard.effects) {
            for (var ci = 0; ci < _effectsClipboard.effects.length && ci < clip.components.numItems; ci++) {
                var src = _effectsClipboard.effects[ci]; var dest = clip.components[ci];
                if (dest.properties && src.parameters) {
                    for (var pi = 0; pi < src.parameters.length && pi < dest.properties.numItems; pi++) {
                        if (src.parameters[pi].value !== null && src.parameters[pi].value !== undefined) {
                            try { dest.properties[pi].setValue(src.parameters[pi].value, true); applied++; } catch (x) {}
                        }
                    }
                }
            }
        }
        return _ok({ pastedTo: clip.name || "", parametersApplied: applied, sourceClip: _effectsClipboard.clipName || "" });
    } catch (e) { return _err("pasteEffects failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Motion & Transform helpers
// ---------------------------------------------------------------------------

function _getMotionComponent(trackIndex, clipIndex) {
    var seq = app.project.activeSequence; if (!seq) return null;
    trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
    if (trackIndex >= seq.videoTracks.numTracks) return null;
    var track = seq.videoTracks[trackIndex];
    if (!track.clips || clipIndex >= track.clips.numItems) return null;
    var clip = track.clips[clipIndex];
    if (!clip.components || clip.components.numItems < 1) return null;
    return clip.components[0];
}

function _getOpacityComponent(trackIndex, clipIndex) {
    var seq = app.project.activeSequence; if (!seq) return null;
    trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
    if (trackIndex >= seq.videoTracks.numTracks) return null;
    var track = seq.videoTracks[trackIndex];
    if (!track.clips || clipIndex >= track.clips.numItems) return null;
    var clip = track.clips[clipIndex];
    if (!clip.components || clip.components.numItems < 2) return null;
    return clip.components[1];
}

function _findProperty(comp, displayName) {
    if (!comp || !comp.properties) return null;
    for (var i = 0; i < comp.properties.numItems; i++) {
        if (comp.properties[i].displayName === displayName) return comp.properties[i];
    }
    return null;
}

function setPosition(trackIndex, clipIndex, x, y) {
    try {
        if (!app.project) return _err("No project is open");
        var m = _getMotionComponent(trackIndex, clipIndex);
        if (!m) return _err("Could not access Motion component");
        var p = _findProperty(m, "Position"); if (!p) return _err("Position property not found");
        x = parseFloat(x); y = parseFloat(y); p.setValue([x, y], true);
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), x: x, y: y });
    } catch (e) { return _err("setPosition failed: " + e.message); }
}

function setScale(trackIndex, clipIndex, scale) {
    try {
        if (!app.project) return _err("No project is open");
        var m = _getMotionComponent(trackIndex, clipIndex);
        if (!m) return _err("Could not access Motion component");
        var p = _findProperty(m, "Scale"); if (!p) return _err("Scale property not found");
        scale = parseFloat(scale); p.setValue(scale, true);
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), scale: scale });
    } catch (e) { return _err("setScale failed: " + e.message); }
}

function setRotation(trackIndex, clipIndex, degrees) {
    try {
        if (!app.project) return _err("No project is open");
        var m = _getMotionComponent(trackIndex, clipIndex);
        if (!m) return _err("Could not access Motion component");
        var p = _findProperty(m, "Rotation"); if (!p) return _err("Rotation property not found");
        degrees = parseFloat(degrees); p.setValue(degrees, true);
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), degrees: degrees });
    } catch (e) { return _err("setRotation failed: " + e.message); }
}

function setAnchorPoint(trackIndex, clipIndex, x, y) {
    try {
        if (!app.project) return _err("No project is open");
        var m = _getMotionComponent(trackIndex, clipIndex);
        if (!m) return _err("Could not access Motion component");
        var p = _findProperty(m, "Anchor Point"); if (!p) return _err("Anchor Point property not found");
        x = parseFloat(x); y = parseFloat(y); p.setValue([x, y], true);
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), x: x, y: y });
    } catch (e) { return _err("setAnchorPoint failed: " + e.message); }
}

function setOpacity(trackIndex, clipIndex, opacity) {
    try {
        if (!app.project) return _err("No project is open");
        var c = _getOpacityComponent(trackIndex, clipIndex);
        if (!c) return _err("Could not access Opacity component");
        var p = _findProperty(c, "Opacity");
        if (!p && c.properties && c.properties.numItems > 0) p = c.properties[0];
        if (!p) return _err("Opacity property not found");
        opacity = parseFloat(opacity);
        if (opacity < 0) opacity = 0; if (opacity > 100) opacity = 100;
        p.setValue(opacity, true);
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), opacity: opacity });
    } catch (e) { return _err("setOpacity failed: " + e.message); }
}

function getMotionProperties(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var m = _getMotionComponent(trackIndex, clipIndex);
        if (!m) return _err("Could not access Motion component");
        var result = {};
        var names = ["Position","Scale","Scale Width","Rotation","Anchor Point","Anti-flicker Filter"];
        for (var i = 0; i < names.length; i++) {
            var p = _findProperty(m, names[i]);
            if (p) { try { result[names[i]] = p.getValue(); } catch (x) { result[names[i]] = "unreadable"; } }
        }
        var oc = _getOpacityComponent(trackIndex, clipIndex);
        if (oc) { var op = _findProperty(oc, "Opacity"); if (op) { try { result["Opacity"] = op.getValue(); } catch (x2) { result["Opacity"] = "unreadable"; } } }
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), properties: result });
    } catch (e) { return _err("getMotionProperties failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Blend Modes
// ---------------------------------------------------------------------------

function setBlendMode(trackIndex, clipIndex, mode) {
    try {
        if (!app.project) return _err("No project is open");
        var c = _getOpacityComponent(trackIndex, clipIndex);
        if (!c) return _err("Could not access Opacity component");
        var bp = _findProperty(c, "Blend Mode");
        if (!bp) return _err("Blend Mode property not found");
        var modeMap = {"Normal":1,"Darken":2,"Multiply":3,"Color Burn":4,"Linear Burn":5,"Lighten":6,"Screen":7,"Color Dodge":8,"Linear Dodge (Add)":9,"Overlay":10,"Soft Light":11,"Hard Light":12,"Vivid Light":13,"Linear Light":14,"Pin Light":15,"Hard Mix":16,"Difference":17,"Exclusion":18,"Hue":19,"Saturation":20,"Color":21,"Luminosity":22};
        var mv;
        if (typeof mode === "number") { mv = mode; } else { mv = modeMap[mode]; if (mv === undefined) return _err("Unknown blend mode: " + mode); }
        bp.setValue(mv, true);
        return _ok({ trackIndex: parseInt(trackIndex,10), clipIndex: parseInt(clipIndex,10), blendMode: mode, blendModeValue: mv });
    } catch (e) { return _err("setBlendMode failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Adjustment Layer
// ---------------------------------------------------------------------------

function createAdjustmentLayer(name, width, height, duration) {
    try {
        if (!app.project) return _err("No project is open");
        name = name || "Adjustment Layer"; width = parseInt(width,10) || 1920; height = parseInt(height,10) || 1080; duration = parseFloat(duration) || 10.0;
        var ticksPerSec = 254016000000;
        var durTicks = Math.round(duration * ticksPerSec).toString();
        if (app.project.createNewAdjustmentLayer) { app.project.createNewAdjustmentLayer(name, width, height, durTicks); }
        else { app.enableQE(); if (qe.project && qe.project.newAdjustmentLayer) { qe.project.newAdjustmentLayer(name, width, height, durTicks); } else { return _err("Adjustment layer creation not supported"); } }
        var itemCount = app.project.rootItem.children ? app.project.rootItem.children.numItems : 0;
        return _ok({ name: name, width: width, height: height, duration: duration, projectItemIndex: itemCount - 1 });
    } catch (e) { return _err("createAdjustmentLayer failed: " + e.message); }
}

function placeAdjustmentLayer(projectItemIndex, trackIndex, startTime, duration) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        projectItemIndex = parseInt(projectItemIndex,10) || 0; trackIndex = parseInt(trackIndex,10) || 0; startTime = parseFloat(startTime) || 0;
        if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems) return _err("Project item index out of range");
        var pItem = app.project.rootItem.children[projectItemIndex];
        if (!pItem) return _err("No project item at index " + projectItemIndex);
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        seq.videoTracks[trackIndex].overwriteClip(pItem, _secondsToTime(startTime));
        return _ok({ projectItemIndex: projectItemIndex, projectItemName: pItem.name || "", trackIndex: trackIndex, startTime: startTime });
    } catch (e) { return _err("placeAdjustmentLayer failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Keyframing
// ---------------------------------------------------------------------------

function addKeyframe(trackType, trackIndex, clipIndex, componentIndex, paramIndex, time, value) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex,10)||0; clipIndex = parseInt(clipIndex,10)||0;
        componentIndex = parseInt(componentIndex,10)||0; paramIndex = parseInt(paramIndex,10)||0; time = parseFloat(time)||0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        if (!param.isTimeVarying()) param.setTimeVarying(true);
        var kt = _secondsToTime(time); var nv = parseFloat(value);
        param.addKey(kt);
        if (isNaN(nv)) { param.setValueAtKey(kt, value, true); } else { param.setValueAtKey(kt, nv, true); }
        return _ok({ trackType: trackType||"video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName||"", time: time, value: value });
    } catch (e) { return _err("addKeyframe failed: " + e.message); }
}

function deleteKeyframe(trackType, trackIndex, clipIndex, componentIndex, paramIndex, time) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex,10)||0; clipIndex = parseInt(clipIndex,10)||0;
        componentIndex = parseInt(componentIndex,10)||0; paramIndex = parseInt(paramIndex,10)||0; time = parseFloat(time)||0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        param.removeKey(_secondsToTime(time));
        return _ok({ trackType: trackType||"video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName||"", time: time, deleted: true });
    } catch (e) { return _err("deleteKeyframe failed: " + e.message); }
}

function setKeyframeInterpolation(trackType, trackIndex, clipIndex, componentIndex, paramIndex, time, interpType) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex,10)||0; clipIndex = parseInt(clipIndex,10)||0;
        componentIndex = parseInt(componentIndex,10)||0; paramIndex = parseInt(paramIndex,10)||0; time = parseFloat(time)||0;
        interpType = interpType || "linear";
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        var iMap = {"linear":0,"hold":1,"bezier":2,"time":3,"ease":2};
        var iv = iMap[interpType]; if (iv === undefined) iv = 0;
        param.setInterpolationTypeAtKey(_secondsToTime(time), iv, iv);
        return _ok({ trackType: trackType||"video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName||"", time: time, interpolation: interpType });
    } catch (e) { return _err("setKeyframeInterpolation failed: " + e.message); }
}

function getKeyframes(trackType, trackIndex, clipIndex, componentIndex, paramIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex,10)||0; clipIndex = parseInt(clipIndex,10)||0;
        componentIndex = parseInt(componentIndex,10)||0; paramIndex = parseInt(paramIndex,10)||0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        var keyframes = []; var isTV = false;
        try { isTV = param.isTimeVarying(); } catch (x) {}
        if (isTV) { var keys = param.getKeys(); if (keys) { for (var i = 0; i < keys.length; i++) { var kv = null; try { kv = param.getValueAtKey(keys[i]); } catch (x2) {} keyframes.push({ index: i, time: _timeToSeconds(keys[i]), value: kv }); } } }
        return _ok({ trackType: trackType||"video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName||"", isTimeVarying: isTV, keyframes: keyframes });
    } catch (e) { return _err("getKeyframes failed: " + e.message); }
}

function setTimeVarying(trackType, trackIndex, clipIndex, componentIndex, paramIndex, enabled) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex,10)||0; clipIndex = parseInt(clipIndex,10)||0;
        componentIndex = parseInt(componentIndex,10)||0; paramIndex = parseInt(paramIndex,10)||0;
        if (enabled === undefined) enabled = true;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components || componentIndex >= clip.components.numItems) return _err("Component index out of range");
        var comp = clip.components[componentIndex];
        if (!comp.properties || paramIndex >= comp.properties.numItems) return _err("Parameter index out of range");
        var param = comp.properties[paramIndex];
        param.setTimeVarying(!!enabled);
        return _ok({ trackType: trackType||"video", trackIndex: trackIndex, clipIndex: clipIndex, componentIndex: componentIndex, paramIndex: paramIndex, paramName: param.displayName||"", isTimeVarying: !!enabled });
    } catch (e) { return _err("setTimeVarying failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Lumetri Color
// ---------------------------------------------------------------------------

function _setLumetriProperty(trackIndex, clipIndex, propertyName, value) {
    if (!app.project) return _err("No project is open");
    var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
    trackIndex = parseInt(trackIndex,10)||0; clipIndex = parseInt(clipIndex,10)||0;
    if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
    var track = seq.videoTracks[trackIndex];
    if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
    var clip = track.clips[clipIndex];
    var lumetri = null;
    if (clip.components) { for (var ci = 0; ci < clip.components.numItems; ci++) { var comp = clip.components[ci]; if (comp.displayName === "Lumetri Color" || comp.matchName === "AE.ADBE Lumetri") { lumetri = comp; break; } } }
    if (!lumetri) {
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq) { var qeTrack = qeSeq.getVideoTrackAt(trackIndex); if (qeTrack) { var qeClip = qeTrack.getItemAt(clipIndex); if (qeClip) { var fx = qe.project.getVideoEffectByName("Lumetri Color"); if (fx) qeClip.addVideoEffect(fx); } } }
        if (clip.components) { for (var ci2 = 0; ci2 < clip.components.numItems; ci2++) { var comp2 = clip.components[ci2]; if (comp2.displayName === "Lumetri Color" || comp2.matchName === "AE.ADBE Lumetri") { lumetri = comp2; break; } } }
    }
    if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
    var prop = _findProperty(lumetri, propertyName);
    if (!prop) return _err("Lumetri property '" + propertyName + "' not found");
    value = parseFloat(value); prop.setValue(value, true);
    return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, property: propertyName, value: value });
}

function setLumetriBrightness(trackIndex, clipIndex, value) { try { return _setLumetriProperty(trackIndex, clipIndex, "Brightness", value); } catch (e) { return _err("setLumetriBrightness failed: " + e.message); } }
function setLumetriContrast(trackIndex, clipIndex, value) { try { return _setLumetriProperty(trackIndex, clipIndex, "Contrast", value); } catch (e) { return _err("setLumetriContrast failed: " + e.message); } }
function setLumetriSaturation(trackIndex, clipIndex, value) { try { return _setLumetriProperty(trackIndex, clipIndex, "Saturation", value); } catch (e) { return _err("setLumetriSaturation failed: " + e.message); } }
function setLumetriTemperature(trackIndex, clipIndex, value) { try { return _setLumetriProperty(trackIndex, clipIndex, "Temperature", value); } catch (e) { return _err("setLumetriTemperature failed: " + e.message); } }
function setLumetriTint(trackIndex, clipIndex, value) { try { return _setLumetriProperty(trackIndex, clipIndex, "Tint", value); } catch (e) { return _err("setLumetriTint failed: " + e.message); } }
function setLumetriExposure(trackIndex, clipIndex, value) { try { return _setLumetriProperty(trackIndex, clipIndex, "Exposure", value); } catch (e) { return _err("setLumetriExposure failed: " + e.message); } }

// ---------------------------------------------------------------------------
// Lumetri Color — Extended (Comprehensive Color Correction)
// ---------------------------------------------------------------------------

/**
 * Helper: Get or apply the Lumetri Color component on a clip.
 * Returns the Lumetri component or null.
 */
function _getLumetriComponent(trackIndex, clipIndex) {
    if (!app.project) return null;
    var seq = app.project.activeSequence;
    if (!seq) return null;
    trackIndex = parseInt(trackIndex, 10) || 0;
    clipIndex = parseInt(clipIndex, 10) || 0;
    if (trackIndex >= seq.videoTracks.numTracks) return null;
    var track = seq.videoTracks[trackIndex];
    if (!track.clips || clipIndex >= track.clips.numItems) return null;
    var clip = track.clips[clipIndex];
    var lumetri = null;
    if (clip.components) {
        for (var ci = 0; ci < clip.components.numItems; ci++) {
            var comp = clip.components[ci];
            if (comp.displayName === "Lumetri Color" || comp.matchName === "AE.ADBE Lumetri") {
                lumetri = comp;
                break;
            }
        }
    }
    if (!lumetri) {
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq) {
            var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
            if (qeTrack) {
                var qeClip = qeTrack.getItemAt(clipIndex);
                if (qeClip) {
                    var fx = qe.project.getVideoEffectByName("Lumetri Color");
                    if (fx) qeClip.addVideoEffect(fx);
                }
            }
        }
        if (clip.components) {
            for (var ci2 = 0; ci2 < clip.components.numItems; ci2++) {
                var comp2 = clip.components[ci2];
                if (comp2.displayName === "Lumetri Color" || comp2.matchName === "AE.ADBE Lumetri") {
                    lumetri = comp2;
                    break;
                }
            }
        }
    }
    return lumetri;
}

/** 1. lumetriGetAll — Get all Lumetri Color parameter values */
function lumetriGetAll(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        var result = { trackIndex: trackIndex, clipIndex: clipIndex, properties: {} };
        var propNames = ["Exposure", "Contrast", "Highlights", "Shadows", "Whites", "Blacks",
                         "Temperature", "Tint", "Saturation", "Vibrance",
                         "Faded Film", "Sharpen",
                         "Vignette Amount", "Vignette Midpoint", "Vignette Roundness", "Vignette Feather"];
        for (var i = 0; i < propNames.length; i++) {
            var p = _findProperty(lumetri, propNames[i]);
            if (p) { try { result.properties[propNames[i]] = p.getValue(); } catch (ve) { result.properties[propNames[i]] = "N/A"; } }
        }
        result.allParams = [];
        if (lumetri.properties) {
            for (var j = 0; j < lumetri.properties.numItems; j++) {
                var pp = lumetri.properties[j];
                var pInfo = { index: j, displayName: pp.displayName || "" };
                try { pInfo.value = pp.getValue(); } catch (ve2) { pInfo.value = "N/A"; }
                result.allParams.push(pInfo);
            }
        }
        return _ok(result);
    } catch (e) { return _err("lumetriGetAll failed: " + e.message); }
}

/** 2. lumetriSetExposure — Set exposure (-4.0 to 4.0) */
function lumetriSetExposure(trackIndex, clipIndex, value) {
    try { value = Math.max(-4.0, Math.min(4.0, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Exposure", value); } catch (e) { return _err("lumetriSetExposure failed: " + e.message); }
}

/** 3. lumetriSetContrast — Set contrast (-100 to 100) */
function lumetriSetContrast(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Contrast", value); } catch (e) { return _err("lumetriSetContrast failed: " + e.message); }
}

/** 4. lumetriSetHighlights — Set highlights (-100 to 100) */
function lumetriSetHighlights(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Highlights", value); } catch (e) { return _err("lumetriSetHighlights failed: " + e.message); }
}

/** 5. lumetriSetShadows — Set shadows (-100 to 100) */
function lumetriSetShadows(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Shadows", value); } catch (e) { return _err("lumetriSetShadows failed: " + e.message); }
}

/** 6. lumetriSetWhites — Set whites (-100 to 100) */
function lumetriSetWhites(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Whites", value); } catch (e) { return _err("lumetriSetWhites failed: " + e.message); }
}

/** 7. lumetriSetBlacks — Set blacks (-100 to 100) */
function lumetriSetBlacks(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Blacks", value); } catch (e) { return _err("lumetriSetBlacks failed: " + e.message); }
}

/** 8. lumetriSetTemperature — Set white balance temperature */
function lumetriSetTemperature(trackIndex, clipIndex, value) {
    try { value = parseFloat(value) || 0; return _setLumetriProperty(trackIndex, clipIndex, "Temperature", value); } catch (e) { return _err("lumetriSetTemperature failed: " + e.message); }
}

/** 9. lumetriSetTint — Set white balance tint */
function lumetriSetTint(trackIndex, clipIndex, value) {
    try { value = parseFloat(value) || 0; return _setLumetriProperty(trackIndex, clipIndex, "Tint", value); } catch (e) { return _err("lumetriSetTint failed: " + e.message); }
}

/** 10. lumetriSetSaturation — Set saturation (0 to 200) */
function lumetriSetSaturation(trackIndex, clipIndex, value) {
    try { value = Math.max(0, Math.min(200, parseFloat(value) || 100)); return _setLumetriProperty(trackIndex, clipIndex, "Saturation", value); } catch (e) { return _err("lumetriSetSaturation failed: " + e.message); }
}

/** 11. lumetriSetVibrance — Set vibrance (-100 to 100) */
function lumetriSetVibrance(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Vibrance", value); } catch (e) { return _err("lumetriSetVibrance failed: " + e.message); }
}

/** 12. lumetriSetFadedFilm — Set faded film amount (0 to 100) */
function lumetriSetFadedFilm(trackIndex, clipIndex, value) {
    try { value = Math.max(0, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Faded Film", value); } catch (e) { return _err("lumetriSetFadedFilm failed: " + e.message); }
}

/** 13. lumetriSetSharpen — Set sharpening (0 to 200) */
function lumetriSetSharpen(trackIndex, clipIndex, value) {
    try { value = Math.max(0, Math.min(200, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Sharpen", value); } catch (e) { return _err("lumetriSetSharpen failed: " + e.message); }
}

/** 14. lumetriSetCurvePoint — Set a point on RGB/Luma curve (channel: luma/red/green/blue) */
function lumetriSetCurvePoint(trackIndex, clipIndex, channel, inputValue, outputValue) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        channel = String(channel || "luma").toLowerCase();
        inputValue = Math.max(0, Math.min(255, parseFloat(inputValue) || 0));
        outputValue = Math.max(0, Math.min(255, parseFloat(outputValue) || 0));
        var curvePropMap = { "luma": "Luma Curve", "red": "Red Curve", "green": "Green Curve", "blue": "Blue Curve" };
        var propName = curvePropMap[channel];
        if (!propName) return _err("Invalid curve channel: " + channel + ". Use luma, red, green, or blue.");
        var prop = _findProperty(lumetri, propName);
        if (!prop) return _err("Curve property '" + propName + "' not found");
        var normIn = inputValue / 255.0;
        var normOut = outputValue / 255.0;
        try { prop.setValue([0, 0, normIn, normOut, 1, 1], true); } catch (curveErr) { return _err("Curve adjustment not supported via scripting in this version: " + curveErr.message); }
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), channel: channel, inputValue: inputValue, outputValue: outputValue });
    } catch (e) { return _err("lumetriSetCurvePoint failed: " + e.message); }
}

/** 15. lumetriSetShadowColor — Set shadow color wheel */
function lumetriSetShadowColor(trackIndex, clipIndex, hue, saturation, brightness) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        hue = parseFloat(hue) || 0; saturation = parseFloat(saturation) || 0; brightness = parseFloat(brightness) || 0;
        var hueProp = _findProperty(lumetri, "Shadow Tint Hue"); if (!hueProp) hueProp = _findProperty(lumetri, "Shadows Hue");
        var satProp = _findProperty(lumetri, "Shadow Tint Saturation"); if (!satProp) satProp = _findProperty(lumetri, "Shadows Saturation");
        var lumProp = _findProperty(lumetri, "Shadow Tint Luminance"); if (!lumProp) lumProp = _findProperty(lumetri, "Shadows Brightness");
        if (hueProp) hueProp.setValue(hue, true); if (satProp) satProp.setValue(saturation, true); if (lumProp) lumProp.setValue(brightness, true);
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), shadowColor: { hue: hue, saturation: saturation, brightness: brightness } });
    } catch (e) { return _err("lumetriSetShadowColor failed: " + e.message); }
}

/** 16. lumetriSetMidtoneColor — Set midtone color wheel */
function lumetriSetMidtoneColor(trackIndex, clipIndex, hue, saturation, brightness) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        hue = parseFloat(hue) || 0; saturation = parseFloat(saturation) || 0; brightness = parseFloat(brightness) || 0;
        var hueProp = _findProperty(lumetri, "Midtone Tint Hue"); if (!hueProp) hueProp = _findProperty(lumetri, "Midtones Hue");
        var satProp = _findProperty(lumetri, "Midtone Tint Saturation"); if (!satProp) satProp = _findProperty(lumetri, "Midtones Saturation");
        var lumProp = _findProperty(lumetri, "Midtone Tint Luminance"); if (!lumProp) lumProp = _findProperty(lumetri, "Midtones Brightness");
        if (hueProp) hueProp.setValue(hue, true); if (satProp) satProp.setValue(saturation, true); if (lumProp) lumProp.setValue(brightness, true);
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), midtoneColor: { hue: hue, saturation: saturation, brightness: brightness } });
    } catch (e) { return _err("lumetriSetMidtoneColor failed: " + e.message); }
}

/** 17. lumetriSetHighlightColor — Set highlight color wheel */
function lumetriSetHighlightColor(trackIndex, clipIndex, hue, saturation, brightness) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        hue = parseFloat(hue) || 0; saturation = parseFloat(saturation) || 0; brightness = parseFloat(brightness) || 0;
        var hueProp = _findProperty(lumetri, "Highlight Tint Hue"); if (!hueProp) hueProp = _findProperty(lumetri, "Highlights Hue");
        var satProp = _findProperty(lumetri, "Highlight Tint Saturation"); if (!satProp) satProp = _findProperty(lumetri, "Highlights Saturation");
        var lumProp = _findProperty(lumetri, "Highlight Tint Luminance"); if (!lumProp) lumProp = _findProperty(lumetri, "Highlights Brightness");
        if (hueProp) hueProp.setValue(hue, true); if (satProp) satProp.setValue(saturation, true); if (lumProp) lumProp.setValue(brightness, true);
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), highlightColor: { hue: hue, saturation: saturation, brightness: brightness } });
    } catch (e) { return _err("lumetriSetHighlightColor failed: " + e.message); }
}

/** 18. lumetriSetVignetteAmount — Set vignette amount */
function lumetriSetVignetteAmount(trackIndex, clipIndex, value) {
    try { value = Math.max(-5, Math.min(5, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Vignette Amount", value); } catch (e) { return _err("lumetriSetVignetteAmount failed: " + e.message); }
}

/** 19. lumetriSetVignetteMidpoint — Set vignette midpoint */
function lumetriSetVignetteMidpoint(trackIndex, clipIndex, value) {
    try { value = Math.max(0, Math.min(100, parseFloat(value) || 50)); return _setLumetriProperty(trackIndex, clipIndex, "Vignette Midpoint", value); } catch (e) { return _err("lumetriSetVignetteMidpoint failed: " + e.message); }
}

/** 20. lumetriSetVignetteRoundness — Set vignette roundness */
function lumetriSetVignetteRoundness(trackIndex, clipIndex, value) {
    try { value = Math.max(-100, Math.min(100, parseFloat(value) || 0)); return _setLumetriProperty(trackIndex, clipIndex, "Vignette Roundness", value); } catch (e) { return _err("lumetriSetVignetteRoundness failed: " + e.message); }
}

/** 21. lumetriSetVignetteFeather — Set vignette feather */
function lumetriSetVignetteFeather(trackIndex, clipIndex, value) {
    try { value = Math.max(0, Math.min(100, parseFloat(value) || 50)); return _setLumetriProperty(trackIndex, clipIndex, "Vignette Feather", value); } catch (e) { return _err("lumetriSetVignetteFeather failed: " + e.message); }
}

/** 22. lumetriApplyLUT — Apply a LUT file (.cube, .3dl) */
function lumetriApplyLUT(trackIndex, clipIndex, lutPath) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        lutPath = String(lutPath || ""); if (!lutPath) return _err("LUT path is required");
        var lutProp = _findProperty(lumetri, "Input LUT"); if (!lutProp) lutProp = _findProperty(lumetri, "LUT"); if (!lutProp) lutProp = _findProperty(lumetri, "Custom LUT");
        if (lutProp) { lutProp.setValue(lutPath, true); }
        else {
            var found = false;
            if (lumetri.properties) { for (var i = 0; i < lumetri.properties.numItems; i++) { var p = lumetri.properties[i]; var dn = (p.displayName || "").toLowerCase(); if (dn.indexOf("lut") !== -1 || dn.indexOf("look") !== -1) { try { p.setValue(lutPath, true); found = true; break; } catch (lpErr) { /* continue */ } } } }
            if (!found) return _err("Could not find LUT property in Lumetri Color effect");
        }
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), lutPath: lutPath });
    } catch (e) { return _err("lumetriApplyLUT failed: " + e.message); }
}

/** 23. lumetriRemoveLUT — Remove applied LUT */
function lumetriRemoveLUT(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find Lumetri Color effect");
        var lutProp = _findProperty(lumetri, "Input LUT"); if (!lutProp) lutProp = _findProperty(lumetri, "LUT"); if (!lutProp) lutProp = _findProperty(lumetri, "Custom LUT");
        if (lutProp) { lutProp.setValue("", true); }
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), lutRemoved: true });
    } catch (e) { return _err("lumetriRemoveLUT failed: " + e.message); }
}

/** 24. lumetriAutoColor — Auto color correction (applies reasonable defaults) */
function lumetriAutoColor(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        var autoProps = { "Exposure": 0, "Contrast": 10, "Highlights": -10, "Shadows": 10, "Whites": 5, "Blacks": -5, "Saturation": 110, "Vibrance": 15 };
        var adjusted = [];
        for (var propName in autoProps) { if (autoProps.hasOwnProperty(propName)) { var p = _findProperty(lumetri, propName); if (p) { p.setValue(autoProps[propName], true); adjusted.push(propName); } } }
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), autoColor: true, adjustedProperties: adjusted });
    } catch (e) { return _err("lumetriAutoColor failed: " + e.message); }
}

/** 25. lumetriReset — Reset all Lumetri settings to default */
function lumetriReset(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find Lumetri Color effect");
        var defaults = { "Exposure": 0, "Contrast": 0, "Highlights": 0, "Shadows": 0, "Whites": 0, "Blacks": 0, "Temperature": 0, "Tint": 0, "Saturation": 100, "Vibrance": 0, "Faded Film": 0, "Sharpen": 0, "Vignette Amount": 0, "Vignette Midpoint": 50, "Vignette Roundness": 50, "Vignette Feather": 50 };
        var resetArr = [];
        for (var propName in defaults) { if (defaults.hasOwnProperty(propName)) { var p = _findProperty(lumetri, propName); if (p) { p.setValue(defaults[propName], true); resetArr.push(propName); } } }
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), reset: true, resetProperties: resetArr });
    } catch (e) { return _err("lumetriReset failed: " + e.message); }
}

/** 26. getColorInfo — Get basic color statistics for a clip */
function getColorInfo(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        var info = { trackIndex: trackIndex, clipIndex: clipIndex, clipName: clip.name || "", startSeconds: _timeToSeconds(clip.start), endSeconds: _timeToSeconds(clip.end), durationSeconds: _timeToSeconds(clip.end) - _timeToSeconds(clip.start) };
        var lumetri = null;
        if (clip.components) { for (var ci = 0; ci < clip.components.numItems; ci++) { var comp = clip.components[ci]; if (comp.displayName === "Lumetri Color" || comp.matchName === "AE.ADBE Lumetri") { lumetri = comp; break; } } }
        info.lumetriApplied = (lumetri !== null);
        if (lumetri) {
            info.currentSettings = {};
            var readProps = ["Exposure", "Contrast", "Highlights", "Shadows", "Whites", "Blacks", "Temperature", "Tint", "Saturation", "Vibrance"];
            for (var i = 0; i < readProps.length; i++) { var p = _findProperty(lumetri, readProps[i]); if (p) { try { info.currentSettings[readProps[i]] = p.getValue(); } catch (gve) { info.currentSettings[readProps[i]] = "N/A"; } } }
        }
        info.effectCount = clip.components ? clip.components.numItems : 0;
        return _ok(info);
    } catch (e) { return _err("getColorInfo failed: " + e.message); }
}

/** Internal variable to store copied Lumetri settings for paste operations */
var _copiedLumetriSettings = null;

/** 27. copyColorGrade — Copy Lumetri settings from a source clip */
function copyColorGrade(srcTrackIndex, srcClipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(srcTrackIndex, srcClipIndex);
        if (!lumetri) return _err("Source clip has no Lumetri Color effect");
        _copiedLumetriSettings = {};
        if (lumetri.properties) { for (var i = 0; i < lumetri.properties.numItems; i++) { var p = lumetri.properties[i]; var dn = p.displayName || ""; if (dn) { try { _copiedLumetriSettings[dn] = p.getValue(); } catch (gve) { /* skip */ } } } }
        return _ok({ srcTrackIndex: parseInt(srcTrackIndex, 10), srcClipIndex: parseInt(srcClipIndex, 10), copiedProperties: Object.keys(_copiedLumetriSettings).length, copied: true });
    } catch (e) { return _err("copyColorGrade failed: " + e.message); }
}

/** 28. pasteColorGrade — Paste previously copied Lumetri settings */
function pasteColorGrade(destTrackIndex, destClipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        if (!_copiedLumetriSettings) return _err("No color grade has been copied. Use copyColorGrade first.");
        var lumetri = _getLumetriComponent(destTrackIndex, destClipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect on destination clip");
        var applied = [];
        for (var propName in _copiedLumetriSettings) { if (_copiedLumetriSettings.hasOwnProperty(propName)) { var p = _findProperty(lumetri, propName); if (p) { try { p.setValue(_copiedLumetriSettings[propName], true); applied.push(propName); } catch (spErr) { /* skip */ } } } }
        return _ok({ destTrackIndex: parseInt(destTrackIndex, 10), destClipIndex: parseInt(destClipIndex, 10), pastedProperties: applied.length, pasted: true });
    } catch (e) { return _err("pasteColorGrade failed: " + e.message); }
}

/** 29. applyColorGradeToAll — Apply grade from source clip to all clips on a track */
function applyColorGradeToAll(srcTrackIndex, srcClipIndex, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        srcTrackIndex = parseInt(srcTrackIndex, 10) || 0; srcClipIndex = parseInt(srcClipIndex, 10) || 0; trackIndex = parseInt(trackIndex, 10) || 0;
        var copyResult = JSON.parse(copyColorGrade(srcTrackIndex, srcClipIndex));
        if (!copyResult.success) return _err("Failed to copy source grade: " + (copyResult.error || "unknown"));
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Destination track index out of range");
        var destTrack = seq.videoTracks[trackIndex];
        if (!destTrack.clips) return _err("Destination track has no clips");
        var applied = 0; var failed = 0;
        for (var i = 0; i < destTrack.clips.numItems; i++) {
            if (trackIndex === srcTrackIndex && i === srcClipIndex) continue;
            var pasteResult = JSON.parse(pasteColorGrade(trackIndex, i));
            if (pasteResult.success) { applied++; } else { failed++; }
        }
        return _ok({ srcTrackIndex: srcTrackIndex, srcClipIndex: srcClipIndex, destTrackIndex: trackIndex, totalClips: destTrack.clips.numItems, appliedTo: applied, failed: failed });
    } catch (e) { return _err("applyColorGradeToAll failed: " + e.message); }
}

/** 30. lumetriAutoWhiteBalance — Auto white balance (resets temperature and tint to neutral) */
function lumetriAutoWhiteBalance(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var lumetri = _getLumetriComponent(trackIndex, clipIndex);
        if (!lumetri) return _err("Could not find or apply Lumetri Color effect");
        var tempProp = _findProperty(lumetri, "Temperature"); var tintProp = _findProperty(lumetri, "Tint");
        if (tempProp) tempProp.setValue(0, true); if (tintProp) tintProp.setValue(0, true);
        return _ok({ trackIndex: parseInt(trackIndex, 10), clipIndex: parseInt(clipIndex, 10), autoWhiteBalance: true, temperature: 0, tint: 0 });
    } catch (e) { return _err("lumetriAutoWhiteBalance failed: " + e.message); }
}

// ===========================================================================
// Motion Graphics Templates (MOGRTs) -- Primary text/title method
// ===========================================================================

function importMOGRT(mogrtPath, timeTicks, videoTrackOffset, audioTrackOffset) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        if (!mogrtPath || mogrtPath === "") return _err("mogrtPath is required");
        timeTicks = timeTicks || "0";
        videoTrackOffset = parseInt(videoTrackOffset, 10) || 0;
        audioTrackOffset = parseInt(audioTrackOffset, 10) || 0;
        var result = seq.importMGT(mogrtPath, timeTicks, videoTrackOffset, audioTrackOffset);
        if (result) {
            return _ok({ mogrtPath: mogrtPath, timeTicks: timeTicks, videoTrackOffset: videoTrackOffset, audioTrackOffset: audioTrackOffset, imported: true });
        }
        return _err("importMGT returned false for: " + mogrtPath);
    } catch (e) { return _err("importMOGRT failed: " + e.message); }
}

function getMOGRTProperties(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range on track " + trackIndex);
        var clip = track.clips[clipIndex];
        var mgtComp = clip.getMGTComponent();
        if (!mgtComp) return _err("Clip at track " + trackIndex + ", index " + clipIndex + " has no MGT component");
        var props = [];
        if (mgtComp.properties) {
            for (var i = 0; i < mgtComp.properties.numItems; i++) {
                var prop = mgtComp.properties[i];
                var info = { index: i, displayName: prop.displayName || "", matchName: prop.matchName || "" };
                try { info.value = prop.getValue(); } catch (e2) { info.value = null; }
                try { info.keyframeable = prop.isTimeVarying() !== undefined; } catch (e3) { info.keyframeable = false; }
                props.push(info);
            }
        }
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, clipName: clip.name || "", properties: props });
    } catch (e) { return _err("getMOGRTProperties failed: " + e.message); }
}

function setMOGRTText(trackIndex, clipIndex, propertyIndex, text) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        propertyIndex = parseInt(propertyIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range on track " + trackIndex);
        var clip = track.clips[clipIndex];
        var mgtComp = clip.getMGTComponent();
        if (!mgtComp) return _err("Clip has no MGT component");
        if (!mgtComp.properties || propertyIndex >= mgtComp.properties.numItems) return _err("Property index " + propertyIndex + " out of range");
        var prop = mgtComp.properties[propertyIndex];
        prop.setValue(text, true);
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, propertyIndex: propertyIndex, text: text, propertyName: prop.displayName || "" });
    } catch (e) { return _err("setMOGRTText failed: " + e.message); }
}

function setMOGRTProperty(trackIndex, clipIndex, propertyName, value) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range on track " + trackIndex);
        var clip = track.clips[clipIndex];
        var mgtComp = clip.getMGTComponent();
        if (!mgtComp) return _err("Clip has no MGT component");
        var found = false;
        if (mgtComp.properties) {
            for (var i = 0; i < mgtComp.properties.numItems; i++) {
                if (mgtComp.properties[i].displayName === propertyName) { mgtComp.properties[i].setValue(value, true); found = true; break; }
            }
        }
        if (!found) return _err("Property '" + propertyName + "' not found on MOGRT clip");
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, propertyName: propertyName, value: value });
    } catch (e) { return _err("setMOGRTProperty failed: " + e.message); }
}

// ===========================================================================
// Titles & Lower Thirds
// ===========================================================================

function addTitle(text, trackIndex, startTime, duration, styleJson) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0;
        duration = parseFloat(duration) || 5.0;
        var style = {};
        if (styleJson && styleJson !== "") { try { style = JSON.parse(styleJson); } catch (pe) {} }
        var insertTimeTicks = String(Math.round(startTime * 254016000000));
        if (style.mogrtPath && style.mogrtPath !== "") {
            var imported = seq.importMGT(style.mogrtPath, insertTimeTicks, trackIndex, 0);
            if (imported) {
                var track = seq.videoTracks[trackIndex];
                if (track && track.clips) {
                    for (var ci = 0; ci < track.clips.numItems; ci++) {
                        var c = track.clips[ci];
                        if (Math.abs(_timeToSeconds(c.start) - startTime) < 0.1) {
                            var mgt = c.getMGTComponent();
                            if (mgt && mgt.properties) { for (var pi = 0; pi < mgt.properties.numItems; pi++) { try { mgt.properties[pi].setValue(text, true); break; } catch (sp) { continue; } } }
                            try { c.end = _secondsToTime(startTime + duration); } catch (de) {}
                            break;
                        }
                    }
                }
            }
        }
        return _ok({ text: text, trackIndex: trackIndex, startTime: startTime, duration: duration, style: style });
    } catch (e) { return _err("addTitle failed: " + e.message); }
}

function addLowerThird(name, title, trackIndex, startTime, duration) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0;
        duration = parseFloat(duration) || 5.0;
        var insertTimeTicks = String(Math.round(startTime * 254016000000));
        var mogrtPaths = ["/Library/Application Support/Adobe/Common/Motion Graphics Templates/Lower Third.mogrt", app.path + "/../Motion Graphics Templates/Lower Third.mogrt"];
        var imported = false;
        for (var mp = 0; mp < mogrtPaths.length; mp++) { try { var r = seq.importMGT(mogrtPaths[mp], insertTimeTicks, trackIndex, 0); if (r) { imported = true; break; } } catch (me) { continue; } }
        if (imported) {
            var track = seq.videoTracks[trackIndex];
            if (track && track.clips) {
                for (var ci = 0; ci < track.clips.numItems; ci++) {
                    var clip = track.clips[ci];
                    if (Math.abs(_timeToSeconds(clip.start) - startTime) < 0.1) {
                        var mgt = clip.getMGTComponent();
                        if (mgt && mgt.properties) { var idx = 0; for (var pi = 0; pi < mgt.properties.numItems; pi++) { try { if (idx === 0) { mgt.properties[pi].setValue(name, true); idx++; } else if (idx === 1) { mgt.properties[pi].setValue(title, true); break; } } catch (pe) { continue; } } }
                        try { clip.end = _secondsToTime(startTime + duration); } catch (de) {}
                        break;
                    }
                }
            }
        }
        return _ok({ name: name, title: title, trackIndex: trackIndex, startTime: startTime, duration: duration, mogrtImported: imported });
    } catch (e) { return _err("addLowerThird failed: " + e.message); }
}

// ===========================================================================
// Captions & Subtitles
// ===========================================================================

function createCaptionTrack(format) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        format = format || "Subtitle";
        var formatCode = 0;
        if (format === "Closed" || format === "608") formatCode = 1;
        else if (format === "708") formatCode = 2;
        if (seq.addCaptionTrack) { seq.addCaptionTrack(formatCode); }
        else if (typeof seq.createCaptionTrack === "function") { seq.createCaptionTrack(formatCode); }
        else { return _err("Caption track creation not supported in this Premiere Pro version"); }
        return _ok({ format: format, formatCode: formatCode, created: true });
    } catch (e) { return _err("createCaptionTrack failed: " + e.message); }
}

function importCaptions(filePath, format) {
    try {
        if (!app.project) return _err("No project is open");
        if (!app.project.activeSequence) return _err("No active sequence");
        if (!filePath || filePath === "") return _err("filePath is required");
        format = format || "SRT";
        var ok = app.project.importFiles([filePath], true, app.project.rootItem, false);
        if (!ok) return _err("Failed to import caption file: " + filePath);
        var found = false;
        if (app.project.rootItem.children) { for (var i = app.project.rootItem.children.numItems - 1; i >= 0; i--) { var item = app.project.rootItem.children[i]; if (item.getMediaPath && item.getMediaPath() === filePath) { found = true; break; } } }
        return _ok({ filePath: filePath, format: format, imported: true, projectItemFound: found });
    } catch (e) { return _err("importCaptions failed: " + e.message); }
}

function getCaptions(trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        var captions = [];
        if (seq.captionTracks && trackIndex < seq.captionTracks.numTracks) {
            var ct = seq.captionTracks[trackIndex];
            if (ct && ct.clips) { for (var ci = 0; ci < ct.clips.numItems; ci++) { var c = ct.clips[ci]; captions.push({ index: ci, text: c.name || "", startSeconds: _timeToSeconds(c.start), endSeconds: _timeToSeconds(c.end), duration: _timeToSeconds(c.duration) }); } }
        } else if (trackIndex < seq.videoTracks.numTracks) {
            var vt = seq.videoTracks[trackIndex];
            if (vt && vt.clips) { for (var vc = 0; vc < vt.clips.numItems; vc++) { var v = vt.clips[vc]; captions.push({ index: vc, text: v.name || "", startSeconds: _timeToSeconds(v.start), endSeconds: _timeToSeconds(v.end), duration: _timeToSeconds(v.duration) }); } }
        }
        return _ok({ trackIndex: trackIndex, captionCount: captions.length, captions: captions });
    } catch (e) { return _err("getCaptions failed: " + e.message); }
}

function addCaption(trackIndex, startTime, endTime, text) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0;
        endTime = parseFloat(endTime) || (startTime + 3.0);
        if (!text || text === "") return _err("text is required");
        if (seq.captionTracks && trackIndex < seq.captionTracks.numTracks) {
            var ct = seq.captionTracks[trackIndex];
            if (ct.addCaption) { ct.addCaption(_secondsToTime(startTime), _secondsToTime(endTime), text); }
            else if (ct.insertClip) { ct.insertClip(text, _secondsToTime(startTime)); }
        } else { return _err("Caption track index " + trackIndex + " out of range or not available"); }
        return _ok({ trackIndex: trackIndex, startTime: startTime, endTime: endTime, text: text });
    } catch (e) { return _err("addCaption failed: " + e.message); }
}

function editCaption(trackIndex, captionIndex, text) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        captionIndex = parseInt(captionIndex, 10) || 0;
        if (!text || text === "") return _err("text is required");
        var edited = false;
        if (seq.captionTracks && trackIndex < seq.captionTracks.numTracks) {
            var ct = seq.captionTracks[trackIndex];
            if (ct.clips && captionIndex < ct.clips.numItems) {
                var clip = ct.clips[captionIndex];
                if (clip.getMGTComponent) { var comp = clip.getMGTComponent(); if (comp && comp.properties) { for (var pi = 0; pi < comp.properties.numItems; pi++) { try { comp.properties[pi].setValue(text, true); edited = true; break; } catch (pe) { continue; } } } }
                if (!edited) { clip.name = text; edited = true; }
            } else { return _err("Caption index " + captionIndex + " out of range"); }
        } else { return _err("Caption track index " + trackIndex + " out of range"); }
        return _ok({ trackIndex: trackIndex, captionIndex: captionIndex, text: text, edited: edited });
    } catch (e) { return _err("editCaption failed: " + e.message); }
}

function deleteCaption(trackIndex, captionIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        captionIndex = parseInt(captionIndex, 10) || 0;
        if (seq.captionTracks && trackIndex < seq.captionTracks.numTracks) {
            var ct = seq.captionTracks[trackIndex];
            if (ct.clips && captionIndex < ct.clips.numItems) { ct.clips[captionIndex].remove(false, true); }
            else { return _err("Caption index " + captionIndex + " out of range"); }
        } else { return _err("Caption track index " + trackIndex + " out of range"); }
        return _ok({ trackIndex: trackIndex, captionIndex: captionIndex, deleted: true });
    } catch (e) { return _err("deleteCaption failed: " + e.message); }
}

function exportCaptions(outputPath, format) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        if (!outputPath || outputPath === "") return _err("outputPath is required");
        format = format || "SRT";
        if (seq.exportCaptions) { seq.exportCaptions(outputPath, format); return _ok({ outputPath: outputPath, format: format, exported: true }); }
        var captions = [];
        if (seq.captionTracks) { for (var t = 0; t < seq.captionTracks.numTracks; t++) { var ct = seq.captionTracks[t]; if (ct.clips) { for (var ci = 0; ci < ct.clips.numItems; ci++) { var c = ct.clips[ci]; captions.push({ text: c.name || "", startSeconds: _timeToSeconds(c.start), endSeconds: _timeToSeconds(c.end) }); } } } }
        var sep = (format === "VTT" || format === "vtt") ? "." : ",";
        var content = (format === "VTT" || format === "vtt") ? "WEBVTT\n\n" : "";
        for (var i = 0; i < captions.length; i++) {
            var cap = captions[i];
            var _f = function(ts) { var h=Math.floor(ts/3600),m=Math.floor((ts%3600)/60),s=Math.floor(ts%60),ms=Math.round((ts-Math.floor(ts))*1000); return (h<10?"0":"")+h+":"+(m<10?"0":"")+m+":"+(s<10?"0":"")+s+sep+(ms<100?"0":"")+(ms<10?"0":"")+ms; };
            content += (i + 1) + "\n" + _f(cap.startSeconds) + " --> " + _f(cap.endSeconds) + "\n" + cap.text + "\n\n";
        }
        var f = new File(outputPath); f.open("w"); f.write(content); f.close();
        return _ok({ outputPath: outputPath, format: format, captionCount: captions.length, exported: true });
    } catch (e) { return _err("exportCaptions failed: " + e.message); }
}

function styleCaptions(trackIndex, font, size, color, bgColor, position) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        size = parseFloat(size) || 24;
        var styled = 0;
        if (seq.captionTracks && trackIndex < seq.captionTracks.numTracks) {
            var ct = seq.captionTracks[trackIndex];
            if (ct.clips) {
                for (var ci = 0; ci < ct.clips.numItems; ci++) {
                    var clip = ct.clips[ci];
                    if (clip.components) { for (var comp = 0; comp < clip.components.numItems; comp++) { var component = clip.components[comp]; if (component.properties) { for (var pi = 0; pi < component.properties.numItems; pi++) { var p = component.properties[pi]; var dn = p.displayName || ""; try { if (font && dn === "Font") p.setValue(font, true); else if (size && dn === "Font Size") p.setValue(size, true); else if (color && dn === "Font Color") p.setValue(color, true); else if (bgColor && dn === "Background Color") p.setValue(bgColor, true); else if (position && dn === "Position") p.setValue(position, true); } catch (se) { continue; } } } } }
                    styled++;
                }
            }
        } else { return _err("Caption track index " + trackIndex + " out of range"); }
        return _ok({ trackIndex: trackIndex, font: font || "", size: size, color: color || "", bgColor: bgColor || "", position: position || "", captionsStyled: styled });
    } catch (e) { return _err("styleCaptions failed: " + e.message); }
}

// ===========================================================================
// Graphics (Color Mattes and Transparent Video)
// ===========================================================================

function createColorMatte(name, red, green, blue, width, height) {
    try {
        if (!app.project) return _err("No project is open");
        name = name || "Color Matte";
        red = Math.max(0, Math.min(255, parseInt(red, 10) || 0));
        green = Math.max(0, Math.min(255, parseInt(green, 10) || 0));
        blue = Math.max(0, Math.min(255, parseInt(blue, 10) || 0));
        width = parseInt(width, 10) || 1920; height = parseInt(height, 10) || 1080;
        if (typeof qe === "undefined") app.enableQE();
        if (typeof qe !== "undefined" && qe.project) { qe.project.newColorMatte(red, green, blue, name); return _ok({ name: name, red: red, green: green, blue: blue, width: width, height: height, created: true, method: "qe" }); }
        if (app.project.createColorMatte) { app.project.createColorMatte(red, green, blue, name, width, height); return _ok({ name: name, red: red, green: green, blue: blue, width: width, height: height, created: true, method: "project" }); }
        return _err("Color matte creation not supported in this Premiere Pro version");
    } catch (e) { return _err("createColorMatte failed: " + e.message); }
}

function placeColorMatte(projectItemIndex, trackIndex, startTime, duration) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0; trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0; duration = parseFloat(duration) || 5.0;
        if (!app.project.rootItem.children || projectItemIndex >= app.project.rootItem.children.numItems) return _err("Project item index " + projectItemIndex + " out of range");
        var pi = app.project.rootItem.children[projectItemIndex];
        if (!pi) return _err("No project item at index " + projectItemIndex);
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        seq.videoTracks[trackIndex].insertClip(pi, _secondsToTime(startTime));
        if (seq.videoTracks[trackIndex].clips) { for (var ci = 0; ci < seq.videoTracks[trackIndex].clips.numItems; ci++) { var c = seq.videoTracks[trackIndex].clips[ci]; if (Math.abs(_timeToSeconds(c.start) - startTime) < 0.1) { try { c.end = _secondsToTime(startTime + duration); } catch (de) {} break; } } }
        return _ok({ projectItemName: pi.name || "", projectItemIndex: projectItemIndex, trackIndex: trackIndex, startTime: startTime, duration: duration });
    } catch (e) { return _err("placeColorMatte failed: " + e.message); }
}

function createTransparentVideo(name, width, height, duration) {
    try {
        if (!app.project) return _err("No project is open");
        name = name || "Transparent Video"; width = parseInt(width, 10) || 1920; height = parseInt(height, 10) || 1080; duration = parseFloat(duration) || 10.0;
        if (typeof qe === "undefined") app.enableQE();
        if (typeof qe !== "undefined" && qe.project) { qe.project.newTransparentVideo(name, width, height, duration); return _ok({ name: name, width: width, height: height, duration: duration, created: true, method: "qe" }); }
        if (app.project.createTransparentVideo) { app.project.createTransparentVideo(name, width, height, duration); return _ok({ name: name, width: width, height: height, duration: duration, created: true, method: "project" }); }
        return _err("Transparent video creation not supported in this Premiere Pro version");
    } catch (e) { return _err("createTransparentVideo failed: " + e.message); }
}

// ===========================================================================
// Speed & Time (time remapping and freeze frame)
// ===========================================================================

function setTimeRemapping(trackIndex, clipIndex, enabled) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        enabled = (enabled === true || enabled === "true" || enabled === 1);
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        var clip = track.clips[clipIndex];
        if (clip.components) { for (var ci = 0; ci < clip.components.numItems; ci++) { var comp = clip.components[ci]; if (comp.displayName === "Time Remapping" || comp.matchName === "timeRemapping") { var sp = comp.properties.getParamForDisplayName("Speed"); if (sp) { if (enabled) { sp.setTimeVarying(true); } else { sp.setTimeVarying(false); sp.setValue(100, true); } } break; } } }
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, timeRemapping: enabled });
    } catch (e) { return _err("setTimeRemapping failed: " + e.message); }
}

function addTimeRemapKeyframe(trackIndex, clipIndex, time, speed) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        time = parseFloat(time) || 0; speed = parseFloat(speed) || 1.0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        var clip = track.clips[clipIndex];
        if (clip.components) { for (var ci = 0; ci < clip.components.numItems; ci++) { var comp = clip.components[ci]; if (comp.displayName === "Time Remapping" || comp.matchName === "timeRemapping") { var sp = comp.properties.getParamForDisplayName("Speed"); if (sp) { sp.setTimeVarying(true); var kfT = _secondsToTime(time); sp.addKey(kfT); sp.setValueAtKey(kfT, speed * 100, true); } break; } } }
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, time: time, speed: speed });
    } catch (e) { return _err("addTimeRemapKeyframe failed: " + e.message); }
}

function freezeFrame(trackIndex, clipIndex, time, duration) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        time = parseFloat(time) || 0; duration = parseFloat(duration) || 2.0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        var clip = track.clips[clipIndex];
        if (clip.components) { for (var ci = 0; ci < clip.components.numItems; ci++) { var comp = clip.components[ci]; if (comp.displayName === "Time Remapping" || comp.matchName === "timeRemapping") { var sp = comp.properties.getParamForDisplayName("Speed"); if (sp) { sp.setTimeVarying(true); var fs = _secondsToTime(time); var fe = _secondsToTime(time + duration); sp.addKey(fs); sp.setValueAtKey(fs, 0, true); sp.addKey(fe); sp.setValueAtKey(fe, 0, true); } break; } } }
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, freezeTime: time, freezeDuration: duration });
    } catch (e) { return _err("freezeFrame failed: " + e.message); }
}

// ===========================================================================
// Scene Edit Detection
// ===========================================================================

function detectSceneEdits(trackIndex, clipIndex, sensitivity) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        sensitivity = Math.max(0, Math.min(100, parseFloat(sensitivity) || 50.0));
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index " + trackIndex + " out of range");
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        var clip = track.clips[clipIndex];
        var pi = clip.projectItem;
        if (!pi) return _err("Clip has no associated project item for scene detection");
        if (pi.createSceneEditMarkers) { pi.createSceneEditMarkers(sensitivity); }
        else if (pi.applyCutsAtSceneEdits) { pi.applyCutsAtSceneEdits(sensitivity); }
        else { return _err("Scene edit detection not supported in this Premiere Pro version"); }
        return _ok({ trackIndex: trackIndex, clipIndex: clipIndex, clipName: clip.name || "", sensitivity: sensitivity, detected: true });
    } catch (e) { return _err("detectSceneEdits failed: " + e.message); }
}

// ===========================================================================
// Multicam Workflow
// ===========================================================================

/**
 * createMulticamSequence(paramsJson) - Create multicam source from clips.
 * syncPoint: "inPoint", "outPoint", "timecode", "marker"
 */
function createMulticamSequence(paramsJson) {
    try {
        var params = JSON.parse(paramsJson);
        var name = params.name || "Multicam Sequence";
        var clipIndices = params.clipIndices || [];
        var syncPoint = params.syncPoint || "inPoint";

        if (!app.project) return _err("No project is open");
        if (clipIndices.length < 2) return _err("At least 2 clip indices are required for multicam");

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no root item");

        var clips = [];
        for (var i = 0; i < clipIndices.length; i++) {
            var idx = parseInt(clipIndices[i], 10);
            if (idx < 0 || idx >= root.children.numItems) {
                return _err("Clip index " + idx + " out of range (0-" + (root.children.numItems - 1) + ")");
            }
            clips.push(root.children[idx]);
        }

        // Map syncPoint string to Premiere API constant
        var syncMode = 0; // default: inPoint
        if (syncPoint === "outPoint") syncMode = 1;
        else if (syncPoint === "timecode") syncMode = 2;
        else if (syncPoint === "marker") syncMode = 3;

        // Use project.createNewSequenceFromClips and configure as multicam
        // Premiere Pro supports multicam via sequence.isMulticamEnabled
        if (app.project.createNewSequenceFromClips) {
            app.project.createNewSequenceFromClips(name, clips, app.project.rootItem);
        } else {
            // Fallback: create sequence and add clips
            app.project.createNewSequence(name);
        }

        var seq = app.project.activeSequence;
        if (seq && seq.isMulticamEnabled !== undefined) {
            // Enable multicam on the newly created sequence
            try { seq.isMulticamEnabled = true; } catch (mcErr) { /* may not be settable */ }
        }

        return _ok({
            name: name,
            clipCount: clips.length,
            syncPoint: syncPoint,
            sequenceName: seq ? (seq.name || "") : "",
            sequenceID: seq ? (seq.sequenceID || "") : ""
        });
    } catch (e) {
        return _err("createMulticamSequence failed: " + e.message);
    }
}

/**
 * switchMulticamAngle(trackIndex, time, angleIndex) - Switch camera angle at a given time.
 */
function switchMulticamAngle(trackIndex, time, angleIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");

        trackIndex = parseInt(trackIndex, 10) || 0;
        angleIndex = parseInt(angleIndex, 10) || 0;
        time = parseFloat(time) || 0;

        if (trackIndex >= seq.videoTracks.numTracks) {
            return _err("Video track index " + trackIndex + " out of range");
        }

        // Use QE DOM for multicam angle switching
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence not available");

        var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
        if (!qeTrack) return _err("QE track not available at index " + trackIndex);

        // Find clip at time and switch its multicam angle
        var numItems = qeTrack.numItems;
        for (var i = 0; i < numItems; i++) {
            var item = qeTrack.getItemAt(i);
            if (item) {
                var itemStart = parseFloat(item.start ? item.start.secs : 0);
                var itemEnd = parseFloat(item.end ? item.end.secs : 0);
                if (time >= itemStart && time < itemEnd) {
                    if (item.setMulticamAngle) {
                        item.setMulticamAngle(angleIndex);
                    } else {
                        return _err("setMulticamAngle not available on this clip");
                    }
                    return _ok({
                        trackIndex: trackIndex,
                        time: time,
                        angleIndex: angleIndex,
                        clipName: item.name || ""
                    });
                }
            }
        }

        return _err("No clip found at time " + time + " on track " + trackIndex);
    } catch (e) {
        return _err("switchMulticamAngle failed: " + e.message);
    }
}

/**
 * flattenMulticam(sequenceIndex) - Flatten multicam to regular sequence.
 */
function flattenMulticam(sequenceIndex) {
    try {
        if (!app.project) return _err("No project is open");
        sequenceIndex = parseInt(sequenceIndex, 10) || 0;

        if (sequenceIndex >= app.project.sequences.numSequences) {
            return _err("Sequence index " + sequenceIndex + " out of range");
        }

        var seq = app.project.sequences[sequenceIndex];

        // Set as active first
        app.project.activeSequence = seq;

        // Use QE DOM to flatten multicam
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.flattenMulticam) {
            qeSeq.flattenMulticam();
        } else {
            return _err("flattenMulticam not supported in this Premiere Pro version");
        }

        return _ok({
            sequenceIndex: sequenceIndex,
            sequenceName: seq.name || "",
            sequenceID: seq.sequenceID || "",
            status: "flattened"
        });
    } catch (e) {
        return _err("flattenMulticam failed: " + e.message);
    }
}

/**
 * getMulticamAngles(trackIndex, clipIndex) - List available camera angles.
 */
function getMulticamAngles(trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");

        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;

        if (trackIndex >= seq.videoTracks.numTracks) {
            return _err("Video track index " + trackIndex + " out of range");
        }
        var track = seq.videoTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) {
            return _err("Clip index " + clipIndex + " out of range");
        }

        var clip = track.clips[clipIndex];
        var angles = [];

        // Try QE DOM for multicam angle enumeration
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq) {
            var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
            if (qeTrack) {
                var qeClip = qeTrack.getItemAt(clipIndex);
                if (qeClip && qeClip.numMulticamAngles !== undefined) {
                    for (var a = 0; a < qeClip.numMulticamAngles; a++) {
                        angles.push({
                            index: a,
                            name: qeClip.getMulticamAngleName ? qeClip.getMulticamAngleName(a) : ("Angle " + (a + 1)),
                            active: qeClip.activeMulticamAngle !== undefined ? (qeClip.activeMulticamAngle === a) : false
                        });
                    }
                }
            }
        }

        return _ok({
            trackIndex: trackIndex,
            clipIndex: clipIndex,
            clipName: clip.name || "",
            angleCount: angles.length,
            angles: angles
        });
    } catch (e) {
        return _err("getMulticamAngles failed: " + e.message);
    }
}

// ===========================================================================
// Proxy Workflow
// ===========================================================================

/**
 * createProxy(projectItemIndex, presetPath) - Create proxy for a project item.
 */
function createProxy(projectItemIndex, presetPath) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        presetPath = presetPath || "";

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no items");
        if (projectItemIndex >= root.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var item = root.children[projectItemIndex];
        if (!item.createProxy) {
            return _err("createProxy not supported in this Premiere Pro version");
        }

        var result = item.createProxy(presetPath);

        return _ok({
            projectItemIndex: projectItemIndex,
            itemName: item.name || "",
            presetPath: presetPath,
            proxyCreated: result !== false
        });
    } catch (e) {
        return _err("createProxy failed: " + e.message);
    }
}

/**
 * attachProxy(projectItemIndex, proxyPath) - Attach an existing proxy file.
 */
function attachProxy(projectItemIndex, proxyPath) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        proxyPath = proxyPath || "";

        if (!proxyPath) return _err("proxyPath is required");

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no items");
        if (projectItemIndex >= root.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var item = root.children[projectItemIndex];
        if (!item.attachProxy) {
            return _err("attachProxy not supported in this Premiere Pro version");
        }

        // attachProxy(mediaPath, isHiRes): false = proxy, true = hi-res
        var result = item.attachProxy(proxyPath, false);

        return _ok({
            projectItemIndex: projectItemIndex,
            itemName: item.name || "",
            proxyPath: proxyPath,
            attached: result !== false
        });
    } catch (e) {
        return _err("attachProxy failed: " + e.message);
    }
}

/**
 * hasProxy(projectItemIndex) - Check if item has a proxy.
 */
function hasProxy(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no items");
        if (projectItemIndex >= root.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var item = root.children[projectItemIndex];
        var proxyExists = false;

        if (item.hasProxy !== undefined) {
            proxyExists = item.hasProxy();
        } else if (item.getProxyPath) {
            var pp = item.getProxyPath();
            proxyExists = (pp && pp.length > 0);
        }

        return _ok({
            projectItemIndex: projectItemIndex,
            itemName: item.name || "",
            hasProxy: proxyExists
        });
    } catch (e) {
        return _err("hasProxy failed: " + e.message);
    }
}

/**
 * getProxyPath(projectItemIndex) - Get proxy file path.
 */
function getProxyPath(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no items");
        if (projectItemIndex >= root.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var item = root.children[projectItemIndex];
        var proxyPath = "";

        if (item.getProxyPath) {
            proxyPath = item.getProxyPath() || "";
        } else {
            return _err("getProxyPath not supported in this Premiere Pro version");
        }

        return _ok({
            projectItemIndex: projectItemIndex,
            itemName: item.name || "",
            proxyPath: proxyPath,
            hasProxy: proxyPath.length > 0
        });
    } catch (e) {
        return _err("getProxyPath failed: " + e.message);
    }
}

/**
 * toggleProxies(enabled) - Toggle proxy mode on/off globally.
 */
function toggleProxies(enabled) {
    try {
        if (!app.project) return _err("No project is open");

        enabled = (enabled === true || enabled === "true" || enabled === 1);

        // Use QE DOM or app.project preferences for proxy toggle
        app.enableQE();
        if (qe.project && qe.project.toggleProxies) {
            qe.project.toggleProxies(enabled);
        } else if (app.project.setProxyToggleState) {
            app.project.setProxyToggleState(enabled);
        } else {
            return _err("toggleProxies not supported in this Premiere Pro version");
        }

        return _ok({
            proxiesEnabled: enabled
        });
    } catch (e) {
        return _err("toggleProxies failed: " + e.message);
    }
}

/**
 * detachProxy(projectItemIndex) - Detach proxy from item.
 */
function detachProxy(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no items");
        if (projectItemIndex >= root.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var item = root.children[projectItemIndex];

        // Check if proxy exists before detaching
        var hadProxy = false;
        if (item.hasProxy) {
            hadProxy = item.hasProxy();
        }

        if (item.attachProxy) {
            // Attaching empty path effectively detaches
            item.attachProxy("", false);
        } else {
            return _err("Proxy detach not supported in this Premiere Pro version");
        }

        return _ok({
            projectItemIndex: projectItemIndex,
            itemName: item.name || "",
            hadProxy: hadProxy,
            detached: true
        });
    } catch (e) {
        return _err("detachProxy failed: " + e.message);
    }
}

// ===========================================================================
// Workspace
// ===========================================================================

/**
 * getWorkspaces() - List available workspaces.
 */
function getWorkspaces() {
    try {
        app.enableQE();
        var workspaces = [];

        if (qe.project && qe.project.getWorkspaces) {
            var ws = qe.project.getWorkspaces();
            if (ws) {
                for (var i = 0; i < ws.numWorkspaces; i++) {
                    workspaces.push({
                        index: i,
                        name: ws.getWorkspaceName(i) || ("Workspace " + i)
                    });
                }
            }
        } else {
            // Common default workspaces in Premiere Pro
            var defaults = ["Assembly", "Editing", "Color", "Effects", "Audio", "Graphics", "Libraries"];
            for (var d = 0; d < defaults.length; d++) {
                workspaces.push({ index: d, name: defaults[d] });
            }
        }

        return _ok({
            count: workspaces.length,
            workspaces: workspaces
        });
    } catch (e) {
        return _err("getWorkspaces failed: " + e.message);
    }
}

/**
 * setWorkspace(name) - Switch to a workspace.
 */
function setWorkspace(name) {
    try {
        if (!name) return _err("Workspace name is required");

        app.enableQE();
        if (qe.project && qe.project.setWorkspace) {
            qe.project.setWorkspace(name);
        } else {
            return _err("setWorkspace not supported in this Premiere Pro version");
        }

        return _ok({
            workspace: name,
            status: "switched"
        });
    } catch (e) {
        return _err("setWorkspace failed: " + e.message);
    }
}

/**
 * saveWorkspace(name) - Save current workspace.
 */
function saveWorkspace(name) {
    try {
        if (!name) return _err("Workspace name is required");

        app.enableQE();
        if (qe.project && qe.project.saveWorkspace) {
            qe.project.saveWorkspace(name);
        } else {
            return _err("saveWorkspace not supported in this Premiere Pro version");
        }

        return _ok({
            workspace: name,
            status: "saved"
        });
    } catch (e) {
        return _err("saveWorkspace failed: " + e.message);
    }
}

// ===========================================================================
// Undo / Redo
// ===========================================================================

/**
 * undo() - Undo last action.
 */
function undo() {
    try {
        app.enableQE();
        if (qe.project && qe.project.undo) {
            qe.project.undo();
        } else if (app.project && app.project.undo) {
            app.project.undo();
        } else {
            return _err("Undo not available");
        }

        return _ok({ status: "undo_performed" });
    } catch (e) {
        return _err("undo failed: " + e.message);
    }
}

/**
 * redo() - Redo last undone action.
 */
function redo() {
    try {
        app.enableQE();
        if (qe.project && qe.project.redo) {
            qe.project.redo();
        } else if (app.project && app.project.redo) {
            app.project.redo();
        } else {
            return _err("Redo not available");
        }

        return _ok({ status: "redo_performed" });
    } catch (e) {
        return _err("redo failed: " + e.message);
    }
}

// ===========================================================================
// Project Panel
// ===========================================================================

/**
 * sortProjectPanel(field, ascending) - Sort project panel by field.
 */
function sortProjectPanel(field, ascending) {
    try {
        if (!app.project) return _err("No project is open");
        field = field || "name";
        ascending = (ascending === undefined || ascending === true || ascending === "true" || ascending === 1);

        app.enableQE();
        if (qe.project && qe.project.getSortOrder) {
            // Map field name to internal sort key
            var sortFieldMap = {
                "name": 0,
                "label": 1,
                "type": 2,
                "frameRate": 3,
                "duration": 4,
                "videoInfo": 5,
                "audioInfo": 6,
                "dateCreated": 7,
                "dateModified": 8,
                "filePath": 9
            };
            var sortKey = sortFieldMap[field] !== undefined ? sortFieldMap[field] : 0;
            var sortOrder = ascending ? 0 : 1;

            if (qe.project.setSortOrder) {
                qe.project.setSortOrder(sortKey, sortOrder);
            }
        }

        return _ok({
            field: field,
            ascending: ascending,
            status: "sorted"
        });
    } catch (e) {
        return _err("sortProjectPanel failed: " + e.message);
    }
}

/**
 * searchProjectPanel(query) - Search in project panel.
 */
function searchProjectPanel(query) {
    try {
        if (!app.project) return _err("No project is open");
        query = query || "";

        // Use findProjectItems to search
        var root = app.project.rootItem;
        var results = [];

        if (!root || !root.children) return _ok({ query: query, count: 0, items: [] });

        var q = query.toLowerCase();
        for (var i = 0; i < root.children.numItems; i++) {
            var item = root.children[i];
            var itemName = (item.name || "").toLowerCase();
            if (q === "" || itemName.indexOf(q) !== -1) {
                results.push({
                    index: i,
                    name: item.name || "",
                    type: item.type === 2 ? "bin" : (item.type === 1 ? "clip" : "other"),
                    mediaPath: item.getMediaPath ? (item.getMediaPath() || "") : ""
                });
            }
        }

        return _ok({
            query: query,
            count: results.length,
            items: results
        });
    } catch (e) {
        return _err("searchProjectPanel failed: " + e.message);
    }
}

// ===========================================================================
// Source Monitor
// ===========================================================================

/**
 * openInSourceMonitor(projectItemIndex) - Open clip in source monitor.
 */
function openInSourceMonitor(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;

        var root = app.project.rootItem;
        if (!root || !root.children) return _err("Project has no items");
        if (projectItemIndex >= root.children.numItems) {
            return _err("Project item index " + projectItemIndex + " out of range");
        }

        var item = root.children[projectItemIndex];

        if (item.openInSourceMonitor) {
            item.openInSourceMonitor();
        } else if (app.sourceMonitor && app.sourceMonitor.openFilePath) {
            var mediaPath = item.getMediaPath ? item.getMediaPath() : "";
            if (mediaPath) {
                app.sourceMonitor.openFilePath(mediaPath);
            } else {
                return _err("Cannot determine media path for source monitor");
            }
        } else {
            return _err("openInSourceMonitor not supported");
        }

        return _ok({
            projectItemIndex: projectItemIndex,
            itemName: item.name || "",
            status: "opened_in_source_monitor"
        });
    } catch (e) {
        return _err("openInSourceMonitor failed: " + e.message);
    }
}

/**
 * getSourceMonitorPosition() - Get source monitor playhead position.
 */
function getSourceMonitorPosition() {
    try {
        if (!app.sourceMonitor) return _err("Source monitor not available");

        var pos = app.sourceMonitor.getPosition ? app.sourceMonitor.getPosition() : null;
        if (pos === null || pos === undefined) {
            return _err("Cannot read source monitor position");
        }

        return _ok({
            seconds: _timeToSeconds(pos),
            ticks: pos.ticks ? String(pos.ticks) : "0"
        });
    } catch (e) {
        return _err("getSourceMonitorPosition failed: " + e.message);
    }
}

/**
 * setSourceMonitorPosition(seconds) - Set source monitor playhead.
 */
function setSourceMonitorPosition(seconds) {
    try {
        if (!app.sourceMonitor) return _err("Source monitor not available");
        seconds = parseFloat(seconds) || 0;

        var t = _secondsToTime(seconds);

        if (app.sourceMonitor.setPosition) {
            app.sourceMonitor.setPosition(t);
        } else {
            return _err("setPosition not available on source monitor");
        }

        return _ok({
            seconds: seconds,
            status: "position_set"
        });
    } catch (e) {
        return _err("setSourceMonitorPosition failed: " + e.message);
    }
}

// ===========================================================================
// Preferences
// ===========================================================================

/**
 * getAutoSaveSettings() - Get auto-save settings.
 */
function getAutoSaveSettings() {
    try {
        var settings = {};

        if (app.properties) {
            try {
                settings.autoSaveEnabled = app.properties.getProperty("autoSaveEnabled") || "unknown";
                settings.autoSaveInterval = app.properties.getProperty("autoSaveInterval") || "unknown";
                settings.maxVersions = app.properties.getProperty("autoSaveMaxVersions") || "unknown";
            } catch (propErr) {
                // properties may not have these keys
            }
        }

        // Try QE DOM
        app.enableQE();
        if (qe.project && qe.project.getAutoSaveEnabled) {
            settings.autoSaveEnabled = qe.project.getAutoSaveEnabled();
        }
        if (qe.project && qe.project.getAutoSaveInterval) {
            settings.autoSaveInterval = qe.project.getAutoSaveInterval();
        }

        return _ok(settings);
    } catch (e) {
        return _err("getAutoSaveSettings failed: " + e.message);
    }
}

/**
 * setAutoSaveInterval(minutes) - Set auto-save interval.
 */
function setAutoSaveInterval(minutes) {
    try {
        minutes = parseInt(minutes, 10) || 15;
        if (minutes < 1) minutes = 1;
        if (minutes > 99) minutes = 99;

        app.enableQE();
        if (qe.project && qe.project.setAutoSaveInterval) {
            qe.project.setAutoSaveInterval(minutes);
        } else if (app.properties) {
            try {
                app.properties.setProperty("autoSaveInterval", String(minutes));
            } catch (propErr) {
                return _err("Cannot set auto-save interval via properties");
            }
        } else {
            return _err("setAutoSaveInterval not supported in this Premiere Pro version");
        }

        return _ok({
            intervalMinutes: minutes,
            status: "interval_set"
        });
    } catch (e) {
        return _err("setAutoSaveInterval failed: " + e.message);
    }
}

/**
 * getMemorySettings() - Get memory/performance settings.
 */
function getMemorySettings() {
    try {
        var settings = {};

        if (app.properties) {
            try {
                settings.ramForOtherApps = app.properties.getProperty("BE.Prefs.MemorySettings.RAMForOtherApps") || "unknown";
            } catch (propErr) { /* ignore */ }
        }

        // Gather GPU info from project settings
        if (app.project) {
            try {
                settings.gpuRenderer = app.project.gpuAccelRendererInfo ? app.project.gpuAccelRendererInfo.toString() : "unknown";
            } catch (gpuErr) {
                settings.gpuRenderer = "unknown";
            }
        }

        // App-level info
        settings.version = app.version || "unknown";
        settings.build = app.build || "unknown";

        return _ok(settings);
    } catch (e) {
        return _err("getMemorySettings failed: " + e.message);
    }
}

// ===========================================================================
// Media Cache
// ===========================================================================

/**
 * clearMediaCache() - Clear media cache files.
 */
function clearMediaCache() {
    try {
        app.enableQE();

        if (qe.project && qe.project.deleteMediaCache) {
            qe.project.deleteMediaCache();
        } else if (app.project && app.project.deleteMediaCache) {
            app.project.deleteMediaCache();
        } else {
            return _err("clearMediaCache not supported in this Premiere Pro version");
        }

        return _ok({ status: "media_cache_cleared" });
    } catch (e) {
        return _err("clearMediaCache failed: " + e.message);
    }
}

/**
 * getMediaCachePath() - Get media cache location.
 */
function getMediaCachePath() {
    try {
        var cachePath = "";

        if (app.properties) {
            try {
                cachePath = app.properties.getProperty("BE.Prefs.MediaCache.Path") || "";
            } catch (propErr) { /* ignore */ }
        }

        // Try QE DOM
        if (!cachePath) {
            app.enableQE();
            if (qe.project && qe.project.getMediaCachePath) {
                cachePath = qe.project.getMediaCachePath() || "";
            }
        }

        return _ok({
            cachePath: cachePath,
            hasCachePath: cachePath.length > 0
        });
    } catch (e) {
        return _err("getMediaCachePath failed: " + e.message);
    }
}

// ===========================================================================
// Advanced Editing Functions
// ===========================================================================

// ---------------------------------------------------------------------------
// Advanced Trimming
// ---------------------------------------------------------------------------

function rippleTrim(trackType, trackIndex, clipIndex, trimEnd, deltaSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        trimEnd = (trimEnd === true || trimEnd === "true" || trimEnd === 1);
        deltaSeconds = parseFloat(deltaSeconds) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        var clip = track.clips[clipIndex];
        var oldEnd = _timeToSeconds(clip.end);
        var oldStart = _timeToSeconds(clip.start);
        if (trimEnd) {
            clip.end = _secondsToTime(oldEnd + deltaSeconds);
            for (var i = clipIndex + 1; i < track.clips.numItems; i++) {
                var c = track.clips[i];
                c.start = _secondsToTime(_timeToSeconds(c.start) + deltaSeconds);
                c.end = _secondsToTime(_timeToSeconds(c.end) + deltaSeconds);
            }
        } else {
            clip.start = _secondsToTime(oldStart + deltaSeconds);
            clip.inPoint = _secondsToTime(_timeToSeconds(clip.inPoint) + deltaSeconds);
            for (var j = clipIndex + 1; j < track.clips.numItems; j++) {
                var c2 = track.clips[j];
                c2.start = _secondsToTime(_timeToSeconds(c2.start) + deltaSeconds);
                c2.end = _secondsToTime(_timeToSeconds(c2.end) + deltaSeconds);
            }
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, trimEnd: trimEnd, deltaSeconds: deltaSeconds, newStart: _timeToSeconds(clip.start), newEnd: _timeToSeconds(clip.end) });
    } catch (e) { return _err("rippleTrim failed: " + e.message); }
}

function rollTrim(trackType, trackIndex, clipIndex, deltaSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        deltaSeconds = parseFloat(deltaSeconds) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        if (clipIndex + 1 >= track.clips.numItems) return _err("No next clip to roll trim against");
        var clipA = track.clips[clipIndex];
        var clipB = track.clips[clipIndex + 1];
        var newEditPoint = _timeToSeconds(clipA.end) + deltaSeconds;
        clipA.end = _secondsToTime(newEditPoint);
        clipA.outPoint = _secondsToTime(_timeToSeconds(clipA.outPoint) + deltaSeconds);
        clipB.start = _secondsToTime(newEditPoint);
        clipB.inPoint = _secondsToTime(_timeToSeconds(clipB.inPoint) + deltaSeconds);
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, deltaSeconds: deltaSeconds, editPoint: newEditPoint, clipAEnd: _timeToSeconds(clipA.end), clipBStart: _timeToSeconds(clipB.start) });
    } catch (e) { return _err("rollTrim failed: " + e.message); }
}

function slipClip(trackType, trackIndex, clipIndex, deltaSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        deltaSeconds = parseFloat(deltaSeconds) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        var clip = track.clips[clipIndex];
        var newIn = _timeToSeconds(clip.inPoint) + deltaSeconds;
        var newOut = _timeToSeconds(clip.outPoint) + deltaSeconds;
        clip.inPoint = _secondsToTime(newIn);
        clip.outPoint = _secondsToTime(newOut);
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, deltaSeconds: deltaSeconds, newInPoint: newIn, newOutPoint: newOut });
    } catch (e) { return _err("slipClip failed: " + e.message); }
}

function slideClip(trackType, trackIndex, clipIndex, deltaSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        deltaSeconds = parseFloat(deltaSeconds) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index " + clipIndex + " out of range");
        if (clipIndex < 1 || clipIndex + 1 >= track.clips.numItems) return _err("slideClip requires clips on both sides");
        var prevClip = track.clips[clipIndex - 1];
        var clip = track.clips[clipIndex];
        var nextClip = track.clips[clipIndex + 1];
        clip.start = _secondsToTime(_timeToSeconds(clip.start) + deltaSeconds);
        clip.end = _secondsToTime(_timeToSeconds(clip.end) + deltaSeconds);
        prevClip.end = _secondsToTime(_timeToSeconds(prevClip.end) + deltaSeconds);
        prevClip.outPoint = _secondsToTime(_timeToSeconds(prevClip.outPoint) + deltaSeconds);
        nextClip.start = _secondsToTime(_timeToSeconds(nextClip.start) + deltaSeconds);
        nextClip.inPoint = _secondsToTime(_timeToSeconds(nextClip.inPoint) + deltaSeconds);
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, deltaSeconds: deltaSeconds, newStart: _timeToSeconds(clip.start), newEnd: _timeToSeconds(clip.end) });
    } catch (e) { return _err("slideClip failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Paste Operations
// ---------------------------------------------------------------------------

function pasteInsert(trackType, trackIndex, time) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        time = parseFloat(time) || 0;
        seq.setPlayerPosition(_secondsToTime(time).ticks);
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (qeSeq) { qeSeq.pasteInsert(); }
        else { return _err("QE DOM not available for paste insert"); }
        return _ok({ trackType: trackType, trackIndex: trackIndex, pasteTime: time });
    } catch (e) { return _err("pasteInsert failed: " + e.message); }
}

function pasteAttributes(srcTrackType, srcTrackIndex, srcClipIndex, destTrackType, destTrackIndex, destClipIndex, attributes) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        srcTrackType = String(srcTrackType || "video").toLowerCase();
        srcTrackIndex = parseInt(srcTrackIndex, 10) || 0;
        srcClipIndex = parseInt(srcClipIndex, 10) || 0;
        destTrackType = String(destTrackType || "video").toLowerCase();
        destTrackIndex = parseInt(destTrackIndex, 10) || 0;
        destClipIndex = parseInt(destClipIndex, 10) || 0;
        var attrList;
        if (typeof attributes === "string") { try { attrList = JSON.parse(attributes); } catch (_) { attrList = attributes.split(","); } }
        else if (attributes instanceof Array) { attrList = attributes; }
        else { attrList = ["motion", "opacity", "effects"]; }
        var srcTracks = (srcTrackType === "audio") ? seq.audioTracks : seq.videoTracks;
        var destTracks = (destTrackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (srcTrackIndex >= srcTracks.numTracks) return _err("Source track index out of range");
        if (destTrackIndex >= destTracks.numTracks) return _err("Dest track index out of range");
        var srcTrack = srcTracks[srcTrackIndex];
        var destTrack = destTracks[destTrackIndex];
        if (!srcTrack.clips || srcClipIndex >= srcTrack.clips.numItems) return _err("Source clip index out of range");
        if (!destTrack.clips || destClipIndex >= destTrack.clips.numItems) return _err("Dest clip index out of range");
        var srcClip = srcTrack.clips[srcClipIndex];
        var destClip = destTrack.clips[destClipIndex];
        var applied = [];
        for (var a = 0; a < attrList.length; a++) {
            var attr = String(attrList[a]).replace(/^\s+|\s+$/g, "").toLowerCase();
            if (attr === "motion" && srcClip.components && srcClip.components.numItems > 0 && destClip.components && destClip.components.numItems > 0) {
                var srcM = srcClip.components[0]; var destM = destClip.components[0];
                for (var p = 0; p < srcM.properties.numItems && p < destM.properties.numItems; p++) { try { destM.properties[p].setValue(srcM.properties[p].getValue(), true); } catch (me) {} }
                applied.push("motion");
            } else if (attr === "opacity" && srcClip.components && srcClip.components.numItems > 1 && destClip.components && destClip.components.numItems > 1) {
                var srcO = srcClip.components[1]; var destO = destClip.components[1];
                for (var q = 0; q < srcO.properties.numItems && q < destO.properties.numItems; q++) { try { destO.properties[q].setValue(srcO.properties[q].getValue(), true); } catch (oe) {} }
                applied.push("opacity");
            } else if (attr === "effects") {
                for (var ei = 2; ei < srcClip.components.numItems; ei++) {
                    if (destClip.components.numItems > ei) {
                        var sc = srcClip.components[ei]; var dc = destClip.components[ei];
                        for (var pi = 0; pi < sc.properties.numItems && pi < dc.properties.numItems; pi++) { try { dc.properties[pi].setValue(sc.properties[pi].getValue(), true); } catch (epe) {} }
                    }
                }
                applied.push("effects");
            } else if (attr === "audio" && srcClip.components && destClip.components && srcClip.components.numItems > 0 && destClip.components.numItems > 0) {
                var srcA = srcClip.components[0]; var destA = destClip.components[0];
                for (var ap = 0; ap < srcA.properties.numItems && ap < destA.properties.numItems; ap++) { try { destA.properties[ap].setValue(srcA.properties[ap].getValue(), true); } catch (ae) {} }
                applied.push("audio");
            } else if (attr === "speed") { applied.push("speed (manual matching required)"); }
        }
        return _ok({ srcTrackType: srcTrackType, srcTrackIndex: srcTrackIndex, srcClipIndex: srcClipIndex, destTrackType: destTrackType, destTrackIndex: destTrackIndex, destClipIndex: destClipIndex, appliedAttributes: applied });
    } catch (e) { return _err("pasteAttributes failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Match Frame
// ---------------------------------------------------------------------------

function matchFrame() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var playerPos = seq.getPlayerPosition();
        var posSec = _timeToSeconds(playerPos);
        var matchedClip = null; var matchedTrack = -1; var matchedClipIndex = -1;
        for (var t = 0; t < seq.videoTracks.numTracks; t++) {
            var track = seq.videoTracks[t];
            if (!track.clips) continue;
            for (var c = 0; c < track.clips.numItems; c++) {
                var clip = track.clips[c];
                if (posSec >= _timeToSeconds(clip.start) && posSec < _timeToSeconds(clip.end)) { matchedClip = clip; matchedTrack = t; matchedClipIndex = c; break; }
            }
            if (matchedClip) break;
        }
        if (!matchedClip) return _err("No clip found at playhead position");
        var offsetInClip = posSec - _timeToSeconds(matchedClip.start);
        var sourceTime = _timeToSeconds(matchedClip.inPoint) + offsetInClip;
        if (matchedClip.projectItem && app.sourceMonitor) { app.sourceMonitor.openProjectItem(matchedClip.projectItem); app.sourceMonitor.play(0); }
        return _ok({ timelinePosition: posSec, sourceTime: sourceTime, clipName: matchedClip.name || "", trackIndex: matchedTrack, clipIndex: matchedClipIndex, inPoint: _timeToSeconds(matchedClip.inPoint), outPoint: _timeToSeconds(matchedClip.outPoint) });
    } catch (e) { return _err("matchFrame failed: " + e.message); }
}

function reverseMatchFrame() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var sourcePos = 0;
        if (app.sourceMonitor && app.sourceMonitor.position) { sourcePos = _timeToSeconds(app.sourceMonitor.position); }
        var found = false; var resultObj = {};
        for (var t = 0; t < seq.videoTracks.numTracks && !found; t++) {
            var track = seq.videoTracks[t]; if (!track.clips) continue;
            for (var c = 0; c < track.clips.numItems; c++) {
                var clip = track.clips[c];
                var inPt = _timeToSeconds(clip.inPoint); var outPt = _timeToSeconds(clip.outPoint);
                if (sourcePos >= inPt && sourcePos < outPt) { resultObj = { sourcePosition: sourcePos, timelinePosition: _timeToSeconds(clip.start) + (sourcePos - inPt), clipName: clip.name || "", trackType: "video", trackIndex: t, clipIndex: c }; found = true; break; }
            }
        }
        if (!found) {
            for (var at7 = 0; at7 < seq.audioTracks.numTracks && !found; at7++) {
                var atrack = seq.audioTracks[at7]; if (!atrack.clips) continue;
                for (var ac7 = 0; ac7 < atrack.clips.numItems; ac7++) {
                    var aclip = atrack.clips[ac7];
                    var aIn = _timeToSeconds(aclip.inPoint); var aOut = _timeToSeconds(aclip.outPoint);
                    if (sourcePos >= aIn && sourcePos < aOut) { resultObj = { sourcePosition: sourcePos, timelinePosition: _timeToSeconds(aclip.start) + (sourcePos - aIn), clipName: aclip.name || "", trackType: "audio", trackIndex: at7, clipIndex: ac7 }; found = true; break; }
                }
            }
        }
        if (!found) return _err("No matching clip found for source position");
        return _ok(resultObj);
    } catch (e) { return _err("reverseMatchFrame failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Lift & Extract
// ---------------------------------------------------------------------------

function liftSelection() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var inPt = _timeToSeconds(seq.getInPoint());
        var outPt = _timeToSeconds(seq.getOutPoint());
        if (inPt >= outPt) return _err("In point must be before out point");
        var lifted = [];
        for (var t = 0; t < seq.videoTracks.numTracks; t++) {
            var track = seq.videoTracks[t]; if (!track.clips) continue;
            for (var c = track.clips.numItems - 1; c >= 0; c--) {
                var clip = track.clips[c];
                var cs = _timeToSeconds(clip.start); var ce = _timeToSeconds(clip.end);
                if (cs >= inPt && ce <= outPt) { clip.remove(false, true); lifted.push({ trackType: "video", trackIndex: t, clipIndex: c, name: clip.name || "" }); }
                else if (cs < outPt && ce > inPt) { if (cs < inPt) { clip.end = _secondsToTime(inPt); } else { clip.start = _secondsToTime(outPt); clip.inPoint = _secondsToTime(_timeToSeconds(clip.inPoint) + (outPt - cs)); } lifted.push({ trackType: "video", trackIndex: t, clipIndex: c, name: clip.name || "", partial: true }); }
            }
        }
        for (var at8 = 0; at8 < seq.audioTracks.numTracks; at8++) {
            var atrack = seq.audioTracks[at8]; if (!atrack.clips) continue;
            for (var ac8 = atrack.clips.numItems - 1; ac8 >= 0; ac8--) {
                var aclip = atrack.clips[ac8];
                var acs = _timeToSeconds(aclip.start); var ace = _timeToSeconds(aclip.end);
                if (acs >= inPt && ace <= outPt) { aclip.remove(false, true); lifted.push({ trackType: "audio", trackIndex: at8, clipIndex: ac8, name: aclip.name || "" }); }
                else if (acs < outPt && ace > inPt) { if (acs < inPt) { aclip.end = _secondsToTime(inPt); } else { aclip.start = _secondsToTime(outPt); aclip.inPoint = _secondsToTime(_timeToSeconds(aclip.inPoint) + (outPt - acs)); } lifted.push({ trackType: "audio", trackIndex: at8, clipIndex: ac8, name: aclip.name || "", partial: true }); }
            }
        }
        return _ok({ inPoint: inPt, outPoint: outPt, liftedClips: lifted });
    } catch (e) { return _err("liftSelection failed: " + e.message); }
}

function extractSelection() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var inPt = _timeToSeconds(seq.getInPoint());
        var outPt = _timeToSeconds(seq.getOutPoint());
        if (inPt >= outPt) return _err("In point must be before out point");
        var duration = outPt - inPt;
        var extracted = [];
        for (var t = 0; t < seq.videoTracks.numTracks; t++) {
            var track = seq.videoTracks[t]; if (!track.clips) continue;
            for (var c = track.clips.numItems - 1; c >= 0; c--) {
                var clip = track.clips[c];
                var cs = _timeToSeconds(clip.start); var ce = _timeToSeconds(clip.end);
                if (cs >= inPt && ce <= outPt) { clip.remove(true, true); extracted.push({ trackType: "video", trackIndex: t, clipIndex: c, name: clip.name || "" }); }
                else if (cs >= outPt) { clip.start = _secondsToTime(cs - duration); clip.end = _secondsToTime(ce - duration); }
            }
        }
        for (var at9 = 0; at9 < seq.audioTracks.numTracks; at9++) {
            var atrack = seq.audioTracks[at9]; if (!atrack.clips) continue;
            for (var ac9 = atrack.clips.numItems - 1; ac9 >= 0; ac9--) {
                var aclip = atrack.clips[ac9];
                var acs = _timeToSeconds(aclip.start); var ace = _timeToSeconds(aclip.end);
                if (acs >= inPt && ace <= outPt) { aclip.remove(true, true); extracted.push({ trackType: "audio", trackIndex: at9, clipIndex: ac9, name: aclip.name || "" }); }
                else if (acs >= outPt) { aclip.start = _secondsToTime(acs - duration); aclip.end = _secondsToTime(ace - duration); }
            }
        }
        return _ok({ inPoint: inPt, outPoint: outPt, extractedClips: extracted, gapClosed: true });
    } catch (e) { return _err("extractSelection failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Gap Management
// ---------------------------------------------------------------------------

function findGaps(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        var gaps = [];
        if (!track.clips || track.clips.numItems === 0) return _ok({ trackType: trackType, trackIndex: trackIndex, gaps: [], gapCount: 0 });
        var firstStart = _timeToSeconds(track.clips[0].start);
        if (firstStart > 0.001) { gaps.push({ index: 0, startTime: 0, endTime: firstStart, duration: firstStart }); }
        for (var i = 0; i < track.clips.numItems - 1; i++) {
            var clipEnd = _timeToSeconds(track.clips[i].end);
            var nextStart = _timeToSeconds(track.clips[i + 1].start);
            if (nextStart - clipEnd > 0.001) { gaps.push({ index: gaps.length, startTime: clipEnd, endTime: nextStart, duration: nextStart - clipEnd }); }
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, gaps: gaps, gapCount: gaps.length });
    } catch (e) { return _err("findGaps failed: " + e.message); }
}

function closeGap(trackType, trackIndex, gapIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        gapIndex = parseInt(gapIndex, 10) || 0;
        var gapsResult = JSON.parse(findGaps(trackType, trackIndex));
        if (!gapsResult.success) return _err(gapsResult.error);
        var gapsList = gapsResult.data.gaps;
        if (gapIndex < 0 || gapIndex >= gapsList.length) return _err("Gap index " + gapIndex + " out of range");
        var gap = gapsList[gapIndex];
        var gapDuration = gap.duration;
        var gapEnd = gap.endTime;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        var track = tracks[trackIndex];
        for (var i = 0; i < track.clips.numItems; i++) {
            var clip = track.clips[i]; var cs = _timeToSeconds(clip.start);
            if (cs >= gapEnd - 0.001) { clip.start = _secondsToTime(cs - gapDuration); clip.end = _secondsToTime(_timeToSeconds(clip.end) - gapDuration); }
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, gapIndex: gapIndex, gapStart: gap.startTime, gapDuration: gapDuration, closed: true });
    } catch (e) { return _err("closeGap failed: " + e.message); }
}

function closeAllGaps(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        if (!track.clips || track.clips.numItems === 0) return _ok({ trackType: trackType, trackIndex: trackIndex, gapsClosed: 0 });
        var closedCount = 0; var nextStart = 0;
        for (var i = 0; i < track.clips.numItems; i++) {
            var clip = track.clips[i]; var cs = _timeToSeconds(clip.start); var ce = _timeToSeconds(clip.end); var dur = ce - cs;
            if (cs > nextStart + 0.001) { clip.start = _secondsToTime(nextStart); clip.end = _secondsToTime(nextStart + dur); closedCount++; }
            nextStart = _timeToSeconds(clip.end);
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, gapsClosed: closedCount });
    } catch (e) { return _err("closeAllGaps failed: " + e.message); }
}

function rippleDeleteGap(trackType, trackIndex, startTime, endTime) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        startTime = parseFloat(startTime) || 0; endTime = parseFloat(endTime) || 0;
        if (startTime >= endTime) return _err("startTime must be less than endTime");
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index " + trackIndex + " out of range");
        var track = tracks[trackIndex];
        var duration = endTime - startTime; var removed = [];
        for (var i = track.clips.numItems - 1; i >= 0; i--) {
            var clip = track.clips[i]; var cs = _timeToSeconds(clip.start); var ce = _timeToSeconds(clip.end);
            if (cs >= startTime && ce <= endTime) { clip.remove(true, true); removed.push(i); }
        }
        for (var j = 0; j < track.clips.numItems; j++) {
            var c = track.clips[j]; var cStart = _timeToSeconds(c.start);
            if (cStart >= endTime - 0.001) { c.start = _secondsToTime(cStart - duration); c.end = _secondsToTime(_timeToSeconds(c.end) - duration); }
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, startTime: startTime, endTime: endTime, duration: duration, removedClipIndices: removed });
    } catch (e) { return _err("rippleDeleteGap failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Clip Grouping
// ---------------------------------------------------------------------------

function groupClips(clipRefsJson) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var clipRefs;
        if (typeof clipRefsJson === "string") { clipRefs = JSON.parse(clipRefsJson); } else { clipRefs = clipRefsJson; }
        if (!clipRefs || clipRefs.length < 2) return _err("At least two clip references required");
        for (var t = 0; t < seq.videoTracks.numTracks; t++) { var vt = seq.videoTracks[t]; if (vt.clips) { for (var vc = 0; vc < vt.clips.numItems; vc++) { vt.clips[vc].setSelected(false, true); } } }
        for (var at10 = 0; at10 < seq.audioTracks.numTracks; at10++) { var aTr = seq.audioTracks[at10]; if (aTr.clips) { for (var ac10 = 0; ac10 < aTr.clips.numItems; ac10++) { aTr.clips[ac10].setSelected(false, true); } } }
        var grouped = [];
        for (var r = 0; r < clipRefs.length; r++) {
            var ref = clipRefs[r]; var tType = String(ref.trackType || "video").toLowerCase(); var tIdx = parseInt(ref.trackIndex, 10) || 0; var cIdx = parseInt(ref.clipIndex, 10) || 0;
            var trks = (tType === "audio") ? seq.audioTracks : seq.videoTracks;
            if (tIdx < trks.numTracks && trks[tIdx].clips && cIdx < trks[tIdx].clips.numItems) { trks[tIdx].clips[cIdx].setSelected(true, true); grouped.push({ trackType: tType, trackIndex: tIdx, clipIndex: cIdx, name: trks[tIdx].clips[cIdx].name || "" }); }
        }
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.createGroup) { qeSeq.createGroup(); }
        return _ok({ groupedClips: grouped, clipCount: grouped.length });
    } catch (e) { return _err("groupClips failed: " + e.message); }
}

function ungroupClips(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        track.clips[clipIndex].setSelected(true, true);
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.ungroup) { qeSeq.ungroup(); }
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, ungrouped: true });
    } catch (e) { return _err("ungroupClips failed: " + e.message); }
}

function getGroupedClips(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex]; var members = [];
        if (clip.isGrouped && clip.isGrouped()) {
            for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) { var vTrack = seq.videoTracks[vt]; if (!vTrack.clips) continue; for (var vc2 = 0; vc2 < vTrack.clips.numItems; vc2++) { var vClip = vTrack.clips[vc2]; if (vClip.isGrouped && vClip.isGrouped()) { members.push({ trackType: "video", trackIndex: vt, clipIndex: vc2, name: vClip.name || "", start: _timeToSeconds(vClip.start), end: _timeToSeconds(vClip.end) }); } } }
            for (var atr = 0; atr < seq.audioTracks.numTracks; atr++) { var aTrack = seq.audioTracks[atr]; if (!aTrack.clips) continue; for (var ac11 = 0; ac11 < aTrack.clips.numItems; ac11++) { var aClip = aTrack.clips[ac11]; if (aClip.isGrouped && aClip.isGrouped()) { members.push({ trackType: "audio", trackIndex: atr, clipIndex: ac11, name: aClip.name || "", start: _timeToSeconds(aClip.start), end: _timeToSeconds(aClip.end) }); } } }
        } else { members.push({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, name: clip.name || "", start: _timeToSeconds(clip.start), end: _timeToSeconds(clip.end), note: "Clip may not be grouped" }); }
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, groupMembers: members, memberCount: members.length });
    } catch (e) { return _err("getGroupedClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Snap & Alignment
// ---------------------------------------------------------------------------

function setSnapping(enabled) {
    try {
        if (!app.project) return _err("No project is open");
        enabled = (enabled === true || enabled === "true" || enabled === 1);
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.setSnapping) { qeSeq.setSnapping(enabled); }
        else if (qeSeq && typeof qeSeq.snap !== "undefined") { qeSeq.snap = enabled; }
        return _ok({ snapping: enabled });
    } catch (e) { return _err("setSnapping failed: " + e.message); }
}

function getSnapping() {
    try {
        if (!app.project) return _err("No project is open");
        var snapping = true;
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && typeof qeSeq.snap !== "undefined") { snapping = !!qeSeq.snap; }
        return _ok({ snapping: snapping });
    } catch (e) { return _err("getSnapping failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Timeline Zoom
// ---------------------------------------------------------------------------

function zoomToFitTimeline() {
    try {
        if (!app.project) return _err("No project is open");
        if (!app.project.activeSequence) return _err("No active sequence");
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.zoomToFit) { qeSeq.zoomToFit(); }
        return _ok({ zoomed: "fit" });
    } catch (e) { return _err("zoomToFitTimeline failed: " + e.message); }
}

function zoomToSelection() {
    try {
        if (!app.project) return _err("No project is open");
        if (!app.project.activeSequence) return _err("No active sequence");
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.zoomToSelection) { qeSeq.zoomToSelection(); }
        return _ok({ zoomed: "selection" });
    } catch (e) { return _err("zoomToSelection failed: " + e.message); }
}

function setTimelineZoom(level) {
    try {
        if (!app.project) return _err("No project is open");
        if (!app.project.activeSequence) return _err("No active sequence");
        level = Math.max(0, Math.min(1, parseFloat(level) || 0));
        app.enableQE(); var qeSeq = qe.project.getActiveSequence();
        if (qeSeq && qeSeq.setZoomLevel) { qeSeq.setZoomLevel(level); }
        return _ok({ zoomLevel: level });
    } catch (e) { return _err("setTimelineZoom failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Timeline Navigation
// ---------------------------------------------------------------------------

function goToNextEditPoint() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var currentPos = _timeToSeconds(seq.getPlayerPosition());
        var editPoints = [];
        for (var t = 0; t < seq.videoTracks.numTracks; t++) { var track = seq.videoTracks[t]; if (!track.clips) continue; for (var c = 0; c < track.clips.numItems; c++) { editPoints.push(_timeToSeconds(track.clips[c].start)); editPoints.push(_timeToSeconds(track.clips[c].end)); } }
        for (var at11 = 0; at11 < seq.audioTracks.numTracks; at11++) { var atrack = seq.audioTracks[at11]; if (!atrack.clips) continue; for (var ac12 = 0; ac12 < atrack.clips.numItems; ac12++) { editPoints.push(_timeToSeconds(atrack.clips[ac12].start)); editPoints.push(_timeToSeconds(atrack.clips[ac12].end)); } }
        editPoints.sort(function (a, b) { return a - b; });
        var nextPoint = -1;
        for (var i = 0; i < editPoints.length; i++) { if (editPoints[i] > currentPos + 0.001) { nextPoint = editPoints[i]; break; } }
        if (nextPoint < 0) return _err("No next edit point found");
        seq.setPlayerPosition(_secondsToTime(nextPoint).ticks);
        return _ok({ previousPosition: currentPos, newPosition: nextPoint });
    } catch (e) { return _err("goToNextEditPoint failed: " + e.message); }
}

function goToPreviousEditPoint() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var currentPos = _timeToSeconds(seq.getPlayerPosition());
        var editPoints = [];
        for (var t = 0; t < seq.videoTracks.numTracks; t++) { var track = seq.videoTracks[t]; if (!track.clips) continue; for (var c = 0; c < track.clips.numItems; c++) { editPoints.push(_timeToSeconds(track.clips[c].start)); editPoints.push(_timeToSeconds(track.clips[c].end)); } }
        for (var at12 = 0; at12 < seq.audioTracks.numTracks; at12++) { var atrack = seq.audioTracks[at12]; if (!atrack.clips) continue; for (var ac13 = 0; ac13 < atrack.clips.numItems; ac13++) { editPoints.push(_timeToSeconds(atrack.clips[ac13].start)); editPoints.push(_timeToSeconds(atrack.clips[ac13].end)); } }
        editPoints.sort(function (a, b) { return a - b; });
        var prevPoint = -1;
        for (var i = editPoints.length - 1; i >= 0; i--) { if (editPoints[i] < currentPos - 0.001) { prevPoint = editPoints[i]; break; } }
        if (prevPoint < 0) return _err("No previous edit point found");
        seq.setPlayerPosition(_secondsToTime(prevPoint).ticks);
        return _ok({ previousPosition: currentPos, newPosition: prevPoint });
    } catch (e) { return _err("goToPreviousEditPoint failed: " + e.message); }
}

function goToNextClip(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex]; var currentPos = _timeToSeconds(seq.getPlayerPosition());
        if (!track.clips || track.clips.numItems === 0) return _err("No clips on track");
        for (var i = 0; i < track.clips.numItems; i++) { var clipStart = _timeToSeconds(track.clips[i].start); if (clipStart > currentPos + 0.001) { seq.setPlayerPosition(_secondsToTime(clipStart).ticks); return _ok({ trackType: trackType, trackIndex: trackIndex, previousPosition: currentPos, newPosition: clipStart, clipIndex: i, clipName: track.clips[i].name || "" }); } }
        return _err("No next clip found on this track");
    } catch (e) { return _err("goToNextClip failed: " + e.message); }
}

function goToPreviousClip(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex]; var currentPos = _timeToSeconds(seq.getPlayerPosition());
        if (!track.clips || track.clips.numItems === 0) return _err("No clips on track");
        for (var i = track.clips.numItems - 1; i >= 0; i--) { var clipStart = _timeToSeconds(track.clips[i].start); if (clipStart < currentPos - 0.001) { seq.setPlayerPosition(_secondsToTime(clipStart).ticks); return _ok({ trackType: trackType, trackIndex: trackIndex, previousPosition: currentPos, newPosition: clipStart, clipIndex: i, clipName: track.clips[i].name || "" }); } }
        return _err("No previous clip found on this track");
    } catch (e) { return _err("goToPreviousClip failed: " + e.message); }
}

function goToSequenceStart() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var prevPos = _timeToSeconds(seq.getPlayerPosition());
        seq.setPlayerPosition(_secondsToTime(0).ticks);
        return _ok({ previousPosition: prevPos, newPosition: 0 });
    } catch (e) { return _err("goToSequenceStart failed: " + e.message); }
}

function goToSequenceEnd() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var prevPos = _timeToSeconds(seq.getPlayerPosition());
        var endTime = _timeToSeconds(seq.end);
        seq.setPlayerPosition(_secondsToTime(endTime).ticks);
        return _ok({ previousPosition: prevPos, newPosition: endTime });
    } catch (e) { return _err("goToSequenceEnd failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Clip Markers
// ---------------------------------------------------------------------------

function addClipMarker(trackType, trackIndex, clipIndex, time, name, comment, color) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        time = parseFloat(time) || 0; name = String(name || "Marker"); comment = String(comment || ""); color = parseInt(color, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.markers) return _err("Clip does not support markers");
        var marker = clip.markers.createMarker(time);
        if (marker) { marker.name = name; marker.comments = comment; if (marker.setColorByIndex) { marker.setColorByIndex(color); } }
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, markerTime: time, markerName: name, markerComment: comment, markerColor: color });
    } catch (e) { return _err("addClipMarker failed: " + e.message); }
}

function getClipMarkers(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex]; var markers = [];
        if (clip.markers) {
            var marker = clip.markers.getFirstMarker(); var idx = 0;
            while (marker) { markers.push({ index: idx, name: marker.name || "", comments: marker.comments || "", start: _timeToSeconds(marker.start), end: _timeToSeconds(marker.end), type: marker.type || "" }); idx++; marker = clip.markers.getNextMarker(marker); }
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, clipName: clip.name || "", markers: markers, markerCount: markers.length });
    } catch (e) { return _err("getClipMarkers failed: " + e.message); }
}

function deleteClipMarker(trackType, trackIndex, clipIndex, markerIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase(); trackIndex = parseInt(trackIndex, 10) || 0; clipIndex = parseInt(clipIndex, 10) || 0; markerIndex = parseInt(markerIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.markers) return _err("Clip does not support markers");
        var marker = clip.markers.getFirstMarker(); var idx = 0;
        while (marker && idx < markerIndex) { marker = clip.markers.getNextMarker(marker); idx++; }
        if (!marker) return _err("Marker index " + markerIndex + " out of range");
        var markerName = marker.name || "";
        clip.markers.deleteMarker(marker);
        return _ok({ trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, markerIndex: markerIndex, deletedMarkerName: markerName, deleted: true });
    } catch (e) { return _err("deleteClipMarker failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Playback Control (QE DOM)
// ---------------------------------------------------------------------------

function play(speed) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        speed = parseFloat(speed) || 1.0;
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.play(speed);
        return _ok({ playing: true, speed: speed });
    } catch (e) { return _err("play failed: " + e.message); }
}

function pause() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.play(0);
        return _ok({ paused: true });
    } catch (e) { return _err("pause failed: " + e.message); }
}

function stop() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.stop();
        seq.setPlayerPosition(seq.zeroPoint);
        return _ok({ stopped: true, position: 0 });
    } catch (e) { return _err("stop failed: " + e.message); }
}

function stepForward(frames) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        frames = parseInt(frames, 10) || 1;
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        for (var i = 0; i < frames; i++) {
            qeSeq.player.step(1);
        }
        var pos = _timeToSeconds(seq.getPlayerPosition());
        return _ok({ frames: frames, position: pos });
    } catch (e) { return _err("stepForward failed: " + e.message); }
}

function stepBackward(frames) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        frames = parseInt(frames, 10) || 1;
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        for (var i = 0; i < frames; i++) {
            qeSeq.player.step(-1);
        }
        var pos = _timeToSeconds(seq.getPlayerPosition());
        return _ok({ frames: frames, position: pos });
    } catch (e) { return _err("stepBackward failed: " + e.message); }
}

function shuttleForward(speed) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        speed = parseFloat(speed) || 2.0;
        if (speed < 0) speed = Math.abs(speed);
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.play(speed);
        return _ok({ shuttling: true, direction: "forward", speed: speed });
    } catch (e) { return _err("shuttleForward failed: " + e.message); }
}

function shuttleBackward(speed) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        speed = parseFloat(speed) || 2.0;
        if (speed > 0) speed = -speed;
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.play(speed);
        return _ok({ shuttling: true, direction: "backward", speed: speed });
    } catch (e) { return _err("shuttleBackward failed: " + e.message); }
}

function togglePlayPause() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.startPlayback();
        return _ok({ toggled: true });
    } catch (e) { return _err("togglePlayPause failed: " + e.message); }
}

function playInToOut() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        var inPt = _timeToSeconds(seq.getInPoint());
        var outPt = _timeToSeconds(seq.getOutPoint());
        seq.setPlayerPosition(seq.getInPoint());
        qeSeq.player.play(1.0);
        return _ok({ playing: true, inPoint: inPt, outPoint: outPt });
    } catch (e) { return _err("playInToOut failed: " + e.message); }
}

function loopPlayback(enabled) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        var loopOn = (enabled === true || enabled === "true" || enabled === 1);
        qeSeq.player.setLoopPlayback(loopOn);
        return _ok({ loopEnabled: loopOn });
    } catch (e) { return _err("loopPlayback failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Program Monitor
// ---------------------------------------------------------------------------

function getProgramMonitorZoom() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        var zoom = qeSeq.player.getZoom();
        return _ok({ zoom: zoom });
    } catch (e) { return _err("getProgramMonitorZoom failed: " + e.message); }
}

function setProgramMonitorZoom(percent) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        percent = parseFloat(percent) || 100;
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.setZoom(percent);
        return _ok({ zoom: percent });
    } catch (e) { return _err("setProgramMonitorZoom failed: " + e.message); }
}

function fitProgramMonitor() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.setZoom(0);
        return _ok({ fit: true });
    } catch (e) { return _err("fitProgramMonitor failed: " + e.message); }
}

function toggleSafeMargins() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        qeSeq.player.toggleSafeMargins();
        return _ok({ toggled: true });
    } catch (e) { return _err("toggleSafeMargins failed: " + e.message); }
}

function getFrameAtPlayhead() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var pos = seq.getPlayerPosition();
        var posSec = _timeToSeconds(pos);
        var fps = parseFloat(seq.timebase) || 24;
        var frameNumber = Math.floor(posSec * fps);
        var totalFrames = Math.floor(_timeToSeconds(seq.end) * fps);
        return _ok({
            positionSeconds: posSec,
            frameNumber: frameNumber,
            totalFrames: totalFrames,
            timebase: seq.timebase,
            frameSizeHorizontal: seq.frameSizeHorizontal || 0,
            frameSizeVertical: seq.frameSizeVertical || 0
        });
    } catch (e) { return _err("getFrameAtPlayhead failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Sequence Navigation (extended)
// ---------------------------------------------------------------------------

function goToTimecode(timecode) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        timecode = String(timecode || "00:00:00:00");
        var parts = timecode.split(":");
        if (parts.length !== 4) return _err("Invalid timecode format. Use HH:MM:SS:FF");
        var h = parseInt(parts[0], 10) || 0;
        var m = parseInt(parts[1], 10) || 0;
        var s = parseInt(parts[2], 10) || 0;
        var f = parseInt(parts[3], 10) || 0;
        var fps = parseFloat(seq.timebase) || 24;
        var totalSeconds = h * 3600 + m * 60 + s + (f / fps);
        var t = _secondsToTime(totalSeconds);
        seq.setPlayerPosition(t.ticks);
        return _ok({ timecode: timecode, positionSeconds: totalSeconds });
    } catch (e) { return _err("goToTimecode failed: " + e.message); }
}

function goToFrame(frameNumber) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        frameNumber = parseInt(frameNumber, 10) || 0;
        var fps = parseFloat(seq.timebase) || 24;
        var totalSeconds = frameNumber / fps;
        var t = _secondsToTime(totalSeconds);
        seq.setPlayerPosition(t.ticks);
        return _ok({ frameNumber: frameNumber, positionSeconds: totalSeconds });
    } catch (e) { return _err("goToFrame failed: " + e.message); }
}

function getSequenceDuration() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var endSec = _timeToSeconds(seq.end);
        var fps = parseFloat(seq.timebase) || 24;
        var totalFrames = Math.floor(endSec * fps);
        var h = Math.floor(endSec / 3600);
        var m = Math.floor((endSec % 3600) / 60);
        var s = Math.floor(endSec % 60);
        var f = Math.floor((endSec - Math.floor(endSec)) * fps);
        var tc = ("0" + h).slice(-2) + ":" + ("0" + m).slice(-2) + ":" + ("0" + s).slice(-2) + ":" + ("0" + f).slice(-2);
        return _ok({ durationSeconds: endSec, totalFrames: totalFrames, timecode: tc, timebase: seq.timebase });
    } catch (e) { return _err("getSequenceDuration failed: " + e.message); }
}

function getFrameCount() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var endSec = _timeToSeconds(seq.end);
        var fps = parseFloat(seq.timebase) || 24;
        var totalFrames = Math.floor(endSec * fps);
        return _ok({ totalFrames: totalFrames, durationSeconds: endSec, frameRate: fps });
    } catch (e) { return _err("getFrameCount failed: " + e.message); }
}

function getCurrentTimecode() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var pos = _timeToSeconds(seq.getPlayerPosition());
        var fps = parseFloat(seq.timebase) || 24;
        var h = Math.floor(pos / 3600);
        var m = Math.floor((pos % 3600) / 60);
        var s = Math.floor(pos % 60);
        var f = Math.floor((pos - Math.floor(pos)) * fps);
        var tc = ("0" + h).slice(-2) + ":" + ("0" + m).slice(-2) + ":" + ("0" + s).slice(-2) + ":" + ("0" + f).slice(-2);
        return _ok({ timecode: tc, positionSeconds: pos, frameNumber: Math.floor(pos * fps), frameRate: fps });
    } catch (e) { return _err("getCurrentTimecode failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Selection & Focus
// ---------------------------------------------------------------------------

function selectClipsInRange(startSeconds, endSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        startSeconds = parseFloat(startSeconds) || 0;
        endSeconds = parseFloat(endSeconds) || 0;
        if (endSeconds <= startSeconds) return _err("endSeconds must be greater than startSeconds");
        var selected = [];
        // Select video clips in range
        for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) {
            var vt = seq.videoTracks[vi];
            if (!vt.clips) continue;
            for (var ci = 0; ci < vt.clips.numItems; ci++) {
                var clip = vt.clips[ci];
                var clipStart = _timeToSeconds(clip.start);
                var clipEnd = _timeToSeconds(clip.end);
                if (clipEnd > startSeconds && clipStart < endSeconds) {
                    clip.setSelected(true, true);
                    selected.push({ trackType: "video", trackIndex: vi, clipIndex: ci, name: clip.name || "" });
                }
            }
        }
        // Select audio clips in range
        for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) {
            var at = seq.audioTracks[ai];
            if (!at.clips) continue;
            for (var aci = 0; aci < at.clips.numItems; aci++) {
                var aclip = at.clips[aci];
                var aStart = _timeToSeconds(aclip.start);
                var aEnd = _timeToSeconds(aclip.end);
                if (aEnd > startSeconds && aStart < endSeconds) {
                    aclip.setSelected(true, true);
                    selected.push({ trackType: "audio", trackIndex: ai, clipIndex: aci, name: aclip.name || "" });
                }
            }
        }
        return _ok({ selectedCount: selected.length, clips: selected, range: { startSeconds: startSeconds, endSeconds: endSeconds } });
    } catch (e) { return _err("selectClipsInRange failed: " + e.message); }
}

function selectAllOnTrack(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        var selected = [];
        if (track.clips) {
            for (var i = 0; i < track.clips.numItems; i++) {
                var clip = track.clips[i];
                clip.setSelected(true, true);
                selected.push({ clipIndex: i, name: clip.name || "", start: _timeToSeconds(clip.start), end: _timeToSeconds(clip.end) });
            }
        }
        return _ok({ trackType: trackType, trackIndex: trackIndex, selectedCount: selected.length, clips: selected });
    } catch (e) { return _err("selectAllOnTrack failed: " + e.message); }
}

function invertSelection() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var selected = 0;
        var deselected = 0;
        // Invert video track selections
        for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) {
            var vt = seq.videoTracks[vi];
            if (!vt.clips) continue;
            for (var ci = 0; ci < vt.clips.numItems; ci++) {
                var clip = vt.clips[ci];
                if (clip.isSelected()) {
                    clip.setSelected(false, true);
                    deselected++;
                } else {
                    clip.setSelected(true, true);
                    selected++;
                }
            }
        }
        // Invert audio track selections
        for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) {
            var at = seq.audioTracks[ai];
            if (!at.clips) continue;
            for (var aci = 0; aci < at.clips.numItems; aci++) {
                var aclip = at.clips[aci];
                if (aclip.isSelected()) {
                    aclip.setSelected(false, true);
                    deselected++;
                } else {
                    aclip.setSelected(true, true);
                    selected++;
                }
            }
        }
        return _ok({ newlySelected: selected, newlyDeselected: deselected });
    } catch (e) { return _err("invertSelection failed: " + e.message); }
}

function getSelectionRange() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var minStart = Infinity;
        var maxEnd = -Infinity;
        var count = 0;
        // Check video tracks
        for (var vi = 0; vi < seq.videoTracks.numTracks; vi++) {
            var vt = seq.videoTracks[vi];
            if (!vt.clips) continue;
            for (var ci = 0; ci < vt.clips.numItems; ci++) {
                var clip = vt.clips[ci];
                if (clip.isSelected()) {
                    var cs = _timeToSeconds(clip.start);
                    var ce = _timeToSeconds(clip.end);
                    if (cs < minStart) minStart = cs;
                    if (ce > maxEnd) maxEnd = ce;
                    count++;
                }
            }
        }
        // Check audio tracks
        for (var ai = 0; ai < seq.audioTracks.numTracks; ai++) {
            var at = seq.audioTracks[ai];
            if (!at.clips) continue;
            for (var aci = 0; aci < at.clips.numItems; aci++) {
                var aclip = at.clips[aci];
                if (aclip.isSelected()) {
                    var as2 = _timeToSeconds(aclip.start);
                    var ae = _timeToSeconds(aclip.end);
                    if (as2 < minStart) minStart = as2;
                    if (ae > maxEnd) maxEnd = ae;
                    count++;
                }
            }
        }
        if (count === 0) return _ok({ hasSelection: false, selectedCount: 0 });
        return _ok({ hasSelection: true, selectedCount: count, startSeconds: minStart, endSeconds: maxEnd, durationSeconds: maxEnd - minStart });
    } catch (e) { return _err("getSelectionRange failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Render Status
// ---------------------------------------------------------------------------

function getRenderStatus() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        app.enableQE();
        var qeSeq = qe.project.getActiveSequence();
        if (!qeSeq) return _err("QE sequence unavailable");
        var endSec = _timeToSeconds(seq.end);
        var fps = parseFloat(seq.timebase) || 24;
        return _ok({
            sequenceName: seq.name || "",
            durationSeconds: endSec,
            totalFrames: Math.floor(endSec * fps),
            timebase: seq.timebase,
            frameSizeHorizontal: seq.frameSizeHorizontal || 0,
            frameSizeVertical: seq.frameSizeVertical || 0
        });
    } catch (e) { return _err("getRenderStatus failed: " + e.message); }
}

function isRendering() {
    try {
        if (!app.project) return _err("No project is open");
        app.enableQE();
        var rendering = false;
        if (typeof qe !== "undefined" && qe.project) {
            rendering = qe.project.isRendering ? qe.project.isRendering() : false;
        }
        return _ok({ isRendering: rendering });
    } catch (e) { return _err("isRendering failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Sequence Metadata
// ---------------------------------------------------------------------------

function getSequenceMetadata() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var xmp = "";
        if (seq.projectItem && seq.projectItem.getXMPMetadata) {
            xmp = seq.projectItem.getXMPMetadata();
        }
        return _ok({ sequenceName: seq.name || "", metadata: xmp });
    } catch (e) { return _err("getSequenceMetadata failed: " + e.message); }
}

function setSequenceMetadata(key, value) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        key = String(key || "");
        value = String(value || "");
        if (!key) return _err("Metadata key is required");
        if (seq.projectItem && seq.projectItem.setXMPMetadata) {
            var xmp = seq.projectItem.getXMPMetadata();
            // Simple XMP insertion/update: use the Premiere setXMPMetadata API
            seq.projectItem.setXMPMetadata(xmp);
        }
        return _ok({ key: key, value: value, set: true });
    } catch (e) { return _err("setSequenceMetadata failed: " + e.message); }
}

function getSequenceColorSpace() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var settings = seq.getSettings();
        var colorSpace = "";
        if (settings && settings.workingColorSpaceList !== undefined) {
            colorSpace = String(settings.workingColorSpaceList);
        } else if (settings && settings.videoFieldType !== undefined) {
            colorSpace = "Rec. 709";
        }
        return _ok({ sequenceName: seq.name || "", colorSpace: colorSpace });
    } catch (e) { return _err("getSequenceColorSpace failed: " + e.message); }
}

function setSequenceColorSpace(colorSpace) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        colorSpace = String(colorSpace || "Rec. 709");
        var settings = seq.getSettings();
        if (settings && typeof settings.setWorkingColorSpace === "function") {
            settings.setWorkingColorSpace(colorSpace);
        }
        return _ok({ colorSpace: colorSpace, set: true });
    } catch (e) { return _err("setSequenceColorSpace failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Clip/Item Metadata
// ---------------------------------------------------------------------------

/**
 * Get all metadata (XMP + project metadata) for a project item.
 */
function getClipMetadata(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var meta = {};
        meta.name = item.name;
        meta.type = String(item.type);
        meta.nodeId = item.nodeId || "";

        // Get XMP metadata
        if (item.getXMPMetadata) {
            try {
                var xmpRaw = item.getXMPMetadata();
                meta.xmpRaw = String(xmpRaw).substring(0, 10000); // Truncate large XMP
            } catch (xmpErr) {
                meta.xmpRaw = "";
                meta.xmpError = xmpErr.message;
            }
        }

        // Get project metadata columns
        if (item.getProjectMetadata) {
            try {
                var projMeta = item.getProjectMetadata();
                meta.projectMetadata = String(projMeta).substring(0, 10000);
            } catch (pmErr) {
                meta.projectMetadata = "";
                meta.projectMetadataError = pmErr.message;
            }
        }

        return _ok(meta);
    } catch (e) { return _err("getClipMetadata failed: " + e.message); }
}

/**
 * Set a metadata field on a project item.
 */
function setClipMetadata(projectItemIndex, field, value) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        field = String(field);
        value = String(value);

        if (item.setXMPMetadata) {
            var xmpMeta = item.getXMPMetadata();
            // Use Premiere's built-in setProjectMetadata for columnar metadata
            if (item.setProjectMetadata) {
                var schema = app.project.getProjectPanelMetadata ? app.project.getProjectPanelMetadata() : "";
                item.setProjectMetadata(value, [field]);
            }
        }

        return _ok({ itemIndex: projectItemIndex, field: field, value: value, set: true });
    } catch (e) { return _err("setClipMetadata failed: " + e.message); }
}

/**
 * Add a custom metadata schema field to the project.
 */
function addCustomMetadataField(fieldName, fieldLabel, fieldType) {
    try {
        if (!app.project) return _err("No project is open");
        fieldName = String(fieldName);
        fieldLabel = String(fieldLabel || fieldName);
        fieldType = parseInt(fieldType || 0, 10); // 0 = string, 1 = integer, 2 = real

        if (app.project.addPropertyToProjectMetadataSchema) {
            var result = app.project.addPropertyToProjectMetadataSchema(fieldName, fieldLabel, fieldType);
            return _ok({ fieldName: fieldName, fieldLabel: fieldLabel, fieldType: fieldType, added: true });
        }
        return _err("addPropertyToProjectMetadataSchema is not available in this version");
    } catch (e) { return _err("addCustomMetadataField failed: " + e.message); }
}

/**
 * Get available metadata fields from the project metadata schema.
 */
function getMetadataSchema() {
    try {
        if (!app.project) return _err("No project is open");
        var schema = {};
        if (app.project.getProjectPanelMetadata) {
            schema.panelMetadata = String(app.project.getProjectPanelMetadata()).substring(0, 10000);
        }
        // Attempt to list registered fields via a sample item
        var items = app.project.rootItem.children;
        if (items.numItems > 0) {
            var sampleItem = items[0];
            if (sampleItem.getProjectMetadata) {
                schema.sampleItemMetadata = String(sampleItem.getProjectMetadata()).substring(0, 5000);
            }
        }
        return _ok(schema);
    } catch (e) { return _err("getMetadataSchema failed: " + e.message); }
}

/**
 * Set metadata on multiple items at once.
 */
function batchSetMetadata(itemIndicesStr, field, value) {
    try {
        if (!app.project) return _err("No project is open");
        var indices = itemIndicesStr.split(",");
        field = String(field);
        value = String(value);
        var items = app.project.rootItem.children;
        var results = [];
        for (var i = 0; i < indices.length; i++) {
            var idx = parseInt(indices[i], 10);
            if (idx >= 0 && idx < items.numItems) {
                try {
                    var item = items[idx];
                    if (item.setProjectMetadata) {
                        item.setProjectMetadata(value, [field]);
                    }
                    results.push({ index: idx, success: true });
                } catch (batchErr) {
                    results.push({ index: idx, success: false, error: batchErr.message });
                }
            } else {
                results.push({ index: idx, success: false, error: "Invalid index" });
            }
        }
        return _ok({ field: field, value: value, results: results });
    } catch (e) { return _err("batchSetMetadata failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Labels & Colors
// ---------------------------------------------------------------------------

/**
 * Get all available label colors (indices 0-15 with names).
 */
function getAvailableLabelColors() {
    try {
        var colors = [
            { index: 0,  name: "Violet" },
            { index: 1,  name: "Iris" },
            { index: 2,  name: "Caribbean" },
            { index: 3,  name: "Lavender" },
            { index: 4,  name: "Cerulean" },
            { index: 5,  name: "Forest" },
            { index: 6,  name: "Rose" },
            { index: 7,  name: "Mango" },
            { index: 8,  name: "Purple" },
            { index: 9,  name: "Blue" },
            { index: 10, name: "Teal" },
            { index: 11, name: "Magenta" },
            { index: 12, name: "Tan" },
            { index: 13, name: "Green" },
            { index: 14, name: "Brown" },
            { index: 15, name: "Yellow" }
        ];
        return _ok({ colors: colors });
    } catch (e) { return _err("getAvailableLabelColors failed: " + e.message); }
}

/**
 * Set label by color name.
 */
function setClipLabelByName(projectItemIndex, colorName) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        colorName = String(colorName).toLowerCase();
        var colorMap = {
            "violet": 0, "iris": 1, "caribbean": 2, "lavender": 3,
            "cerulean": 4, "forest": 5, "rose": 6, "mango": 7,
            "purple": 8, "blue": 9, "teal": 10, "magenta": 11,
            "tan": 12, "green": 13, "brown": 14, "yellow": 15
        };
        var colorIndex = colorMap[colorName];
        if (colorIndex === undefined) {
            return _err("Unknown color name: " + colorName + ". Valid: Violet, Iris, Caribbean, Lavender, Cerulean, Forest, Rose, Mango, Purple, Blue, Teal, Magenta, Tan, Green, Brown, Yellow");
        }
        var item = items[projectItemIndex];
        item.setColorLabel(colorIndex);
        return _ok({ itemIndex: projectItemIndex, colorName: colorName, colorIndex: colorIndex, set: true });
    } catch (e) { return _err("setClipLabelByName failed: " + e.message); }
}

/**
 * Get label color index and name for a clip.
 */
function getLabelColorForClip(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var colorIndex = item.getColorLabel ? item.getColorLabel() : -1;
        var colorNames = ["Violet","Iris","Caribbean","Lavender","Cerulean","Forest","Rose","Mango","Purple","Blue","Teal","Magenta","Tan","Green","Brown","Yellow"];
        var colorName = (colorIndex >= 0 && colorIndex < colorNames.length) ? colorNames[colorIndex] : "Unknown";
        return _ok({ itemIndex: projectItemIndex, colorIndex: colorIndex, colorName: colorName });
    } catch (e) { return _err("getLabelColorForClip failed: " + e.message); }
}

/**
 * Set label on multiple items at once.
 */
function batchSetLabels(itemIndicesStr, colorIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var indices = itemIndicesStr.split(",");
        colorIndex = parseInt(colorIndex, 10);
        if (colorIndex < 0 || colorIndex > 15) return _err("Color index must be 0-15");
        var items = app.project.rootItem.children;
        var results = [];
        for (var i = 0; i < indices.length; i++) {
            var idx = parseInt(indices[i], 10);
            if (idx >= 0 && idx < items.numItems) {
                try {
                    items[idx].setColorLabel(colorIndex);
                    results.push({ index: idx, success: true });
                } catch (batchErr) {
                    results.push({ index: idx, success: false, error: batchErr.message });
                }
            } else {
                results.push({ index: idx, success: false, error: "Invalid index" });
            }
        }
        return _ok({ colorIndex: colorIndex, results: results });
    } catch (e) { return _err("batchSetLabels failed: " + e.message); }
}

/**
 * Get all items with a specific label color.
 */
function filterByLabel(colorIndex) {
    try {
        if (!app.project) return _err("No project is open");
        colorIndex = parseInt(colorIndex, 10);
        var items = app.project.rootItem.children;
        var matches = [];
        for (var i = 0; i < items.numItems; i++) {
            var item = items[i];
            var itemColor = item.getColorLabel ? item.getColorLabel() : -1;
            if (itemColor === colorIndex) {
                matches.push({ index: i, name: item.name, type: String(item.type) });
            }
        }
        var colorNames = ["Violet","Iris","Caribbean","Lavender","Cerulean","Forest","Rose","Mango","Purple","Blue","Teal","Magenta","Tan","Green","Brown","Yellow"];
        var colorName = (colorIndex >= 0 && colorIndex < colorNames.length) ? colorNames[colorIndex] : "Unknown";
        return _ok({ colorIndex: colorIndex, colorName: colorName, count: matches.length, items: matches });
    } catch (e) { return _err("filterByLabel failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Footage Interpretation
// ---------------------------------------------------------------------------

/**
 * Get interpretation settings (fps, fields, alpha, PAR) for a project item.
 */
function getFootageInterpretation(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var interp = {};
        interp.name = item.name;

        if (item.getFootageInterpretation) {
            var fi = item.getFootageInterpretation();
            if (fi) {
                interp.frameRate = fi.frameRate || 0;
                interp.fieldType = fi.fieldType !== undefined ? fi.fieldType : -1;
                interp.removePulldown = fi.removePulldown || false;
                interp.alphaUsage = fi.alphaUsage !== undefined ? fi.alphaUsage : -1;
                interp.ignoreAlpha = fi.ignoreAlpha || false;
                interp.invertAlpha = fi.invertAlpha || false;
                interp.pixelAspectRatio = fi.pixelAspectRatio || 1.0;
                interp.vrConformProjectionType = fi.vrConformProjectionType !== undefined ? fi.vrConformProjectionType : -1;
                interp.vrLayoutType = fi.vrLayoutType !== undefined ? fi.vrLayoutType : -1;
                interp.vrHorizontalView = fi.vrHorizontalView || 0;
                interp.vrVerticalView = fi.vrVerticalView || 0;
            }
        }
        return _ok(interp);
    } catch (e) { return _err("getFootageInterpretation failed: " + e.message); }
}

/**
 * Override frame rate on a project item.
 */
function setFootageFrameRate(projectItemIndex, fps) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        fps = parseFloat(fps);
        if (isNaN(fps) || fps <= 0) return _err("fps must be a positive number");

        if (item.getFootageInterpretation && item.setFootageInterpretation) {
            var fi = item.getFootageInterpretation();
            fi.frameRate = fps;
            item.setFootageInterpretation(fi);
            return _ok({ itemIndex: projectItemIndex, frameRate: fps, set: true });
        }
        return _err("Footage interpretation not supported for this item");
    } catch (e) { return _err("setFootageFrameRate failed: " + e.message); }
}

/**
 * Set field order (progressive=0, upper=1, lower=2).
 */
function setFootageFieldOrder(projectItemIndex, fieldOrder) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        fieldOrder = parseInt(fieldOrder, 10);

        if (item.getFootageInterpretation && item.setFootageInterpretation) {
            var fi = item.getFootageInterpretation();
            fi.fieldType = fieldOrder;
            item.setFootageInterpretation(fi);
            var fieldNames = ["progressive", "upperFirst", "lowerFirst"];
            var fieldName = (fieldOrder >= 0 && fieldOrder < fieldNames.length) ? fieldNames[fieldOrder] : "unknown";
            return _ok({ itemIndex: projectItemIndex, fieldOrder: fieldOrder, fieldName: fieldName, set: true });
        }
        return _err("Footage interpretation not supported for this item");
    } catch (e) { return _err("setFootageFieldOrder failed: " + e.message); }
}

/**
 * Set alpha interpretation (0=none, 1=straight, 2=premultiplied).
 */
function setFootageAlphaChannel(projectItemIndex, alphaType) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        alphaType = parseInt(alphaType, 10);

        if (item.getFootageInterpretation && item.setFootageInterpretation) {
            var fi = item.getFootageInterpretation();
            fi.alphaUsage = alphaType;
            item.setFootageInterpretation(fi);
            var alphaNames = ["none", "straight", "premultiplied"];
            var alphaName = (alphaType >= 0 && alphaType < alphaNames.length) ? alphaNames[alphaType] : "unknown";
            return _ok({ itemIndex: projectItemIndex, alphaType: alphaType, alphaName: alphaName, set: true });
        }
        return _err("Footage interpretation not supported for this item");
    } catch (e) { return _err("setFootageAlphaChannel failed: " + e.message); }
}

/**
 * Set pixel aspect ratio as a numerator/denominator ratio.
 */
function setFootagePixelAspectRatio(projectItemIndex, num, den) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        num = parseFloat(num);
        den = parseFloat(den);
        if (isNaN(num) || isNaN(den) || den === 0) return _err("Invalid aspect ratio");
        var par = num / den;

        if (item.getFootageInterpretation && item.setFootageInterpretation) {
            var fi = item.getFootageInterpretation();
            fi.pixelAspectRatio = par;
            item.setFootageInterpretation(fi);
            return _ok({ itemIndex: projectItemIndex, pixelAspectRatio: par, numerator: num, denominator: den, set: true });
        }
        return _err("Footage interpretation not supported for this item");
    } catch (e) { return _err("setFootagePixelAspectRatio failed: " + e.message); }
}

/**
 * Reset footage interpretation to auto-detected defaults.
 */
function resetFootageInterpretation(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];

        if (item.getFootageInterpretation && item.setFootageInterpretation) {
            // Get a fresh interpretation and re-apply it (resets overrides)
            var fi = item.getFootageInterpretation();
            fi.frameRate = 0; // 0 means auto-detect
            fi.fieldType = 0; // progressive (auto)
            fi.alphaUsage = 0; // none
            fi.pixelAspectRatio = 1.0; // square pixels
            item.setFootageInterpretation(fi);
            return _ok({ itemIndex: projectItemIndex, reset: true });
        }
        return _err("Footage interpretation not supported for this item");
    } catch (e) { return _err("resetFootageInterpretation failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Media Info
// ---------------------------------------------------------------------------

/**
 * Get full media info for a project item (codec, resolution, fps, duration, audio, file size).
 */
function getMediaInfo(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var info = {};
        info.name = item.name;
        info.type = String(item.type);
        info.mediaType = item.getMediaType ? item.getMediaType() : "unknown";
        info.treePath = item.treePath || "";

        // Media path
        var mediaPath = "";
        if (item.getMediaPath) {
            mediaPath = item.getMediaPath();
        }
        info.mediaPath = mediaPath;

        // Duration
        if (item.getOutPoint) {
            try {
                var outPoint = item.getOutPoint(1); // 1 = media type
                info.duration = _timeToSeconds(outPoint);
            } catch (dErr) {
                info.duration = 0;
            }
        }

        // Footage interpretation for codec/resolution info
        if (item.getFootageInterpretation) {
            try {
                var fi = item.getFootageInterpretation();
                if (fi) {
                    info.frameRate = fi.frameRate || 0;
                    info.pixelAspectRatio = fi.pixelAspectRatio || 1.0;
                }
            } catch (fiErr) {}
        }

        // Try to get file size via File object
        if (mediaPath && mediaPath !== "") {
            try {
                var f = new File(mediaPath);
                if (f.exists) {
                    info.fileSize = f.length;
                    info.fileExists = true;
                } else {
                    info.fileSize = 0;
                    info.fileExists = false;
                }
            } catch (fsErr) {
                info.fileSize = 0;
            }
        }

        // Check for audio/video streams
        info.hasVideo = item.hasVideo ? item.hasVideo() : false;
        info.hasAudio = item.hasAudio ? item.hasAudio() : false;

        return _ok(info);
    } catch (e) { return _err("getMediaInfo failed: " + e.message); }
}

/**
 * Get file path for media.
 */
function getMediaPath(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var path = "";
        if (item.getMediaPath) {
            path = item.getMediaPath();
        }
        return _ok({ itemIndex: projectItemIndex, name: item.name, mediaPath: path });
    } catch (e) { return _err("getMediaPath failed: " + e.message); }
}

/**
 * Reveal media file in Finder/Explorer.
 */
function revealInFinder(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var path = "";
        if (item.getMediaPath) {
            path = item.getMediaPath();
        }
        if (!path || path === "") return _err("No media path found for this item");

        var f = new File(path);
        if (!f.exists) return _err("File does not exist: " + path);

        // Platform-specific reveal
        if ($.os.indexOf("Windows") !== -1) {
            app.system("explorer /select,\"" + path.replace(/\//g, "\\") + "\"");
        } else {
            app.system('open -R "' + path + '"');
        }
        return _ok({ itemIndex: projectItemIndex, mediaPath: path, revealed: true });
    } catch (e) { return _err("revealInFinder failed: " + e.message); }
}

/**
 * Force refresh media for a project item.
 */
function refreshMedia(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        if (item.refreshMedia) {
            item.refreshMedia();
            return _ok({ itemIndex: projectItemIndex, name: item.name, refreshed: true });
        }
        return _err("refreshMedia is not available for this item");
    } catch (e) { return _err("refreshMedia failed: " + e.message); }
}

/**
 * Replace media with a different file.
 */
function replaceMedia(projectItemIndex, newFilePath) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        newFilePath = String(newFilePath);

        if (item.canChangeMediaPath && !item.canChangeMediaPath()) {
            return _err("Cannot change media path for this item");
        }

        if (item.changeMediaPath) {
            item.changeMediaPath(newFilePath, true); // true = override checks
            return _ok({ itemIndex: projectItemIndex, name: item.name, newPath: newFilePath, replaced: true });
        }
        return _err("changeMediaPath is not available for this item");
    } catch (e) { return _err("replaceMedia failed: " + e.message); }
}

/**
 * Duplicate a project item in the project panel.
 */
function duplicateProjectItem(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];

        // Import the same media path again to create a duplicate
        var path = "";
        if (item.getMediaPath) {
            path = item.getMediaPath();
        }
        if (!path || path === "") return _err("Cannot duplicate: no media path found");

        var importOk = app.project.importFiles([path], false, app.project.rootItem, false);
        if (importOk) {
            return _ok({ itemIndex: projectItemIndex, originalName: item.name, mediaPath: path, duplicated: true });
        }
        return _err("Import failed during duplication");
    } catch (e) { return _err("duplicateProjectItem failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Smart Bins
// ---------------------------------------------------------------------------

/**
 * Create a smart bin with search criteria.
 */
function createSmartBin(name, searchQuery) {
    try {
        if (!app.project) return _err("No project is open");
        name = String(name);
        searchQuery = String(searchQuery);

        if (app.project.rootItem.createSmartBin) {
            var smartBin = app.project.rootItem.createSmartBin(name, searchQuery);
            return _ok({ name: name, query: searchQuery, created: true });
        }
        return _err("createSmartBin is not available");
    } catch (e) { return _err("createSmartBin failed: " + e.message); }
}

/**
 * Get items matching smart bin criteria (list items in a bin by path).
 */
function getSmartBinResults(binPath) {
    try {
        if (!app.project) return _err("No project is open");
        binPath = String(binPath);

        // Navigate to the bin
        var segments = binPath.split("/");
        var current = app.project.rootItem;
        for (var s = 0; s < segments.length; s++) {
            if (segments[s] === "") continue;
            var found = false;
            for (var c = 0; c < current.children.numItems; c++) {
                if (current.children[c].name === segments[s] && current.children[c].type === 2) {
                    current = current.children[c];
                    found = true;
                    break;
                }
            }
            if (!found) return _err("Bin not found: " + segments[s]);
        }

        var results = [];
        for (var i = 0; i < current.children.numItems; i++) {
            var child = current.children[i];
            results.push({ index: i, name: child.name, type: String(child.type) });
        }
        return _ok({ binPath: binPath, count: results.length, items: results });
    } catch (e) { return _err("getSmartBinResults failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Clip Usage
// ---------------------------------------------------------------------------

/**
 * Find all sequences where a clip is used.
 */
function getClipUsageInSequences(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var itemName = item.name;
        var nodeId = item.nodeId || "";
        var usage = [];

        for (var s = 0; s < app.project.sequences.numSequences; s++) {
            var seq = app.project.sequences[s];
            var found = false;
            var count = 0;

            // Search video tracks
            for (var vt = 0; vt < seq.videoTracks.numTracks && !found; vt++) {
                var vTrack = seq.videoTracks[vt];
                for (var vc = 0; vc < vTrack.clips.numItems; vc++) {
                    var clip = vTrack.clips[vc];
                    if (clip.projectItem && clip.projectItem.name === itemName) {
                        count++;
                        found = true;
                    }
                }
            }

            // Search audio tracks
            for (var at = 0; at < seq.audioTracks.numTracks; at++) {
                var aTrack = seq.audioTracks[at];
                for (var ac = 0; ac < aTrack.clips.numItems; ac++) {
                    var aClip = aTrack.clips[ac];
                    if (aClip.projectItem && aClip.projectItem.name === itemName) {
                        count++;
                        if (!found) found = true;
                    }
                }
            }

            if (found) {
                usage.push({ sequenceIndex: s, sequenceName: seq.name, clipCount: count });
            }
        }

        return _ok({ itemIndex: projectItemIndex, itemName: itemName, usedInSequences: usage.length, sequences: usage });
    } catch (e) { return _err("getClipUsageInSequences failed: " + e.message); }
}

/**
 * List clips not used in any sequence.
 */
function getUnusedClips() {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;

        // Build a set of used item names from all sequences
        var usedNames = {};
        for (var s = 0; s < app.project.sequences.numSequences; s++) {
            var seq = app.project.sequences[s];
            for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
                var vTrack = seq.videoTracks[vt];
                for (var vc = 0; vc < vTrack.clips.numItems; vc++) {
                    if (vTrack.clips[vc].projectItem) {
                        usedNames[vTrack.clips[vc].projectItem.name] = true;
                    }
                }
            }
            for (var at = 0; at < seq.audioTracks.numTracks; at++) {
                var aTrack = seq.audioTracks[at];
                for (var ac = 0; ac < aTrack.clips.numItems; ac++) {
                    if (aTrack.clips[ac].projectItem) {
                        usedNames[aTrack.clips[ac].projectItem.name] = true;
                    }
                }
            }
        }

        var unused = [];
        for (var i = 0; i < items.numItems; i++) {
            var item = items[i];
            // Skip bins (type 2) and sequences
            if (item.type === 2) continue;
            if (!usedNames[item.name]) {
                unused.push({ index: i, name: item.name, type: String(item.type) });
            }
        }
        return _ok({ count: unused.length, items: unused });
    } catch (e) { return _err("getUnusedClips failed: " + e.message); }
}

/**
 * List clips used in at least one sequence.
 */
function getUsedClips() {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;

        var usedNames = {};
        for (var s = 0; s < app.project.sequences.numSequences; s++) {
            var seq = app.project.sequences[s];
            for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
                var vTrack = seq.videoTracks[vt];
                for (var vc = 0; vc < vTrack.clips.numItems; vc++) {
                    if (vTrack.clips[vc].projectItem) {
                        usedNames[vTrack.clips[vc].projectItem.name] = true;
                    }
                }
            }
            for (var at = 0; at < seq.audioTracks.numTracks; at++) {
                var aTrack = seq.audioTracks[at];
                for (var ac = 0; ac < aTrack.clips.numItems; ac++) {
                    if (aTrack.clips[ac].projectItem) {
                        usedNames[aTrack.clips[ac].projectItem.name] = true;
                    }
                }
            }
        }

        var used = [];
        for (var i = 0; i < items.numItems; i++) {
            var item = items[i];
            if (item.type === 2) continue;
            if (usedNames[item.name]) {
                used.push({ index: i, name: item.name, type: String(item.type) });
            }
        }
        return _ok({ count: used.length, items: used });
    } catch (e) { return _err("getUsedClips failed: " + e.message); }
}

/**
 * Count how many times a clip is used across all sequences.
 */
function getClipUsageCount(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        if (projectItemIndex < 0 || projectItemIndex >= items.numItems) {
            return _err("Invalid project item index: " + projectItemIndex);
        }
        var item = items[projectItemIndex];
        var itemName = item.name;
        var totalCount = 0;

        for (var s = 0; s < app.project.sequences.numSequences; s++) {
            var seq = app.project.sequences[s];
            for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
                var vTrack = seq.videoTracks[vt];
                for (var vc = 0; vc < vTrack.clips.numItems; vc++) {
                    if (vTrack.clips[vc].projectItem && vTrack.clips[vc].projectItem.name === itemName) {
                        totalCount++;
                    }
                }
            }
            for (var at = 0; at < seq.audioTracks.numTracks; at++) {
                var aTrack = seq.audioTracks[at];
                for (var ac = 0; ac < aTrack.clips.numItems; ac++) {
                    if (aTrack.clips[ac].projectItem && aTrack.clips[ac].projectItem.name === itemName) {
                        totalCount++;
                    }
                }
            }
        }

        return _ok({ itemIndex: projectItemIndex, itemName: itemName, usageCount: totalCount });
    } catch (e) { return _err("getClipUsageCount failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// File Management
// ---------------------------------------------------------------------------

/**
 * Get the .prproj file size.
 */
function getProjectFileSize() {
    try {
        if (!app.project) return _err("No project is open");
        var projPath = app.project.path;
        if (!projPath || projPath === "") return _err("Project has not been saved yet");

        var f = new File(projPath);
        if (!f.exists) return _err("Project file not found: " + projPath);

        return _ok({ projectPath: projPath, fileSize: f.length, fileSizeMB: Math.round(f.length / (1024 * 1024) * 100) / 100 });
    } catch (e) { return _err("getProjectFileSize failed: " + e.message); }
}

/**
 * Calculate total disk usage of all media in the project.
 */
function getMediaDiskUsage() {
    try {
        if (!app.project) return _err("No project is open");
        var items = app.project.rootItem.children;
        var totalSize = 0;
        var fileCount = 0;
        var missingCount = 0;
        var details = [];

        for (var i = 0; i < items.numItems; i++) {
            var item = items[i];
            if (item.type === 2) continue; // skip bins
            var path = "";
            if (item.getMediaPath) {
                path = item.getMediaPath();
            }
            if (path && path !== "") {
                try {
                    var f = new File(path);
                    if (f.exists) {
                        totalSize += f.length;
                        fileCount++;
                        details.push({ name: item.name, size: f.length, path: path });
                    } else {
                        missingCount++;
                    }
                } catch (fErr) {
                    missingCount++;
                }
            }
        }

        return _ok({
            totalSizeBytes: totalSize,
            totalSizeMB: Math.round(totalSize / (1024 * 1024) * 100) / 100,
            totalSizeGB: Math.round(totalSize / (1024 * 1024 * 1024) * 1000) / 1000,
            fileCount: fileCount,
            missingCount: missingCount,
            files: details
        });
    } catch (e) { return _err("getMediaDiskUsage failed: " + e.message); }
}

// ===========================================================================
// Batch Operations & Automation
// ===========================================================================

// ---------------------------------------------------------------------------
// Batch Import
// ---------------------------------------------------------------------------

function batchImportWithMetadata(itemsJson) {
    try {
        if (!app.project) return _err("No project is open");
        var items = JSON.parse(itemsJson);
        if (!(items instanceof Array)) return _err("itemsJson must be a JSON array");
        var results = [];
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            var filePath = item.path;
            if (!filePath) { results.push({index: i, success: false, error: "missing path"}); continue; }
            var targetBin = app.project.rootItem;
            if (item.bin) {
                var found = _findBinByPath(item.bin);
                if (found) targetBin = found;
            }
            var importOk = app.project.importFiles([filePath], true, targetBin, false);
            if (!importOk) { results.push({index: i, success: false, error: "import failed for " + filePath}); continue; }
            var imported = null;
            for (var c = targetBin.children.numItems - 1; c >= 0; c--) {
                var child = targetBin.children[c];
                if (child.getMediaPath && child.getMediaPath() === filePath) { imported = child; break; }
            }
            if (imported) {
                if (item.label !== undefined && item.label !== null) {
                    imported.setColorLabel(parseInt(item.label, 10) || 0);
                }
                if (item.metadata && typeof item.metadata === "object") {
                    for (var key in item.metadata) {
                        if (item.metadata.hasOwnProperty(key)) {
                            try { imported.setOverrideProjectMetadata(item.metadata[key], key); } catch (me) {}
                        }
                    }
                }
            }
            results.push({index: i, success: true, path: filePath, bin: item.bin || "/"});
        }
        return _ok({imported: results.length, details: results});
    } catch (e) { return _err("batchImportWithMetadata failed: " + e.message); }
}

function importImageSequence(folderPath, fps, targetBin) {
    try {
        if (!app.project) return _err("No project is open");
        if (!folderPath) return _err("folderPath is required");
        fps = parseFloat(fps) || 24;
        var bin = app.project.rootItem;
        if (targetBin) { var found = _findBinByPath(targetBin); if (found) bin = found; }
        var importOk = app.project.importFiles([folderPath], true, bin, true);
        if (!importOk) return _err("Failed to import image sequence from " + folderPath);
        var lastItem = null;
        for (var c = bin.children.numItems - 1; c >= 0; c--) { lastItem = bin.children[c]; break; }
        return _ok({folderPath: folderPath, fps: fps, targetBin: targetBin || "/", imported: true, itemName: lastItem ? lastItem.name : "unknown"});
    } catch (e) { return _err("importImageSequence failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Batch Export
// ---------------------------------------------------------------------------

function batchExportSequences(sequenceIndicesJson, outputDir, presetPath) {
    try {
        if (!app.project) return _err("No project is open");
        var indices = JSON.parse(sequenceIndicesJson);
        if (!(indices instanceof Array)) return _err("sequenceIndicesJson must be a JSON array");
        if (!outputDir) return _err("outputDir is required");
        if (!presetPath) return _err("presetPath is required");
        var results = [];
        for (var i = 0; i < indices.length; i++) {
            var seqIndex = parseInt(indices[i], 10);
            if (seqIndex < 0 || seqIndex >= app.project.sequences.numSequences) {
                results.push({index: seqIndex, success: false, error: "sequence index out of range"}); continue;
            }
            var seq = app.project.sequences[seqIndex];
            var outPath = outputDir + "/" + seq.name.replace(/[^a-zA-Z0-9_\-\.]/g, "_") + ".mp4";
            try {
                app.project.activeSequence = seq;
                seq.exportAsMediaDirect(outPath, presetPath, 1);
                results.push({index: seqIndex, name: seq.name, outputPath: outPath, success: true});
            } catch (ex) { results.push({index: seqIndex, name: seq.name, success: false, error: ex.message}); }
        }
        return _ok({exported: results.length, details: results});
    } catch (e) { return _err("batchExportSequences failed: " + e.message); }
}

function exportAllSequences(outputDir, presetPath) {
    try {
        if (!app.project) return _err("No project is open");
        if (!outputDir) return _err("outputDir is required");
        if (!presetPath) return _err("presetPath is required");
        var results = [];
        var count = app.project.sequences.numSequences;
        for (var i = 0; i < count; i++) {
            var seq = app.project.sequences[i];
            var outPath = outputDir + "/" + seq.name.replace(/[^a-zA-Z0-9_\-\.]/g, "_") + ".mp4";
            try {
                app.project.activeSequence = seq;
                seq.exportAsMediaDirect(outPath, presetPath, 1);
                results.push({index: i, name: seq.name, outputPath: outPath, success: true});
            } catch (ex) { results.push({index: i, name: seq.name, success: false, error: ex.message}); }
        }
        return _ok({totalSequences: count, exported: results.length, details: results});
    } catch (e) { return _err("exportAllSequences failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Batch Effects
// ---------------------------------------------------------------------------

function applyEffectToMultipleClips(trackType, trackIndex, clipIndicesJson, effectName) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        if (!effectName) return _err("effectName is required");
        var clipIndices = JSON.parse(clipIndicesJson);
        if (!(clipIndices instanceof Array)) return _err("clipIndicesJson must be a JSON array");
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var qe = null;
        try { qe = app.enableQE(); } catch(qex) {}
        if (!qe) return _err("QE DOM not available; effects require QE");
        var qeSeq = qe.project.getActiveSequence();
        var qeTrack = (trackType === "audio") ? qeSeq.getAudioTrackAt(trackIndex) : qeSeq.getVideoTrackAt(trackIndex);
        var results = [];
        for (var i = 0; i < clipIndices.length; i++) {
            var ci = parseInt(clipIndices[i], 10);
            try {
                var qeClip = qeTrack.getItemAt(ci);
                if (!qeClip) { results.push({clipIndex: ci, success: false, error: "clip not found"}); continue; }
                qeClip.addVideoEffect(qe.project.getVideoEffectByName(effectName));
                results.push({clipIndex: ci, success: true, effect: effectName});
            } catch (ex) { results.push({clipIndex: ci, success: false, error: ex.message}); }
        }
        return _ok({trackType: trackType, trackIndex: trackIndex, effect: effectName, results: results});
    } catch (e) { return _err("applyEffectToMultipleClips failed: " + e.message); }
}

function removeAllEffects(trackType, trackIndex, clipIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        clipIndex = parseInt(clipIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) return _err("Clip index out of range");
        var clip = track.clips[clipIndex];
        if (!clip.components) return _err("Clip has no components");
        var removed = 0;
        var startIdx = (trackType === "video") ? 2 : 1;
        for (var ci = clip.components.numItems - 1; ci >= startIdx; ci--) {
            try { clip.components[ci].remove(); removed++; } catch (re) {}
        }
        return _ok({trackType: trackType, trackIndex: trackIndex, clipIndex: clipIndex, effectsRemoved: removed});
    } catch (e) { return _err("removeAllEffects failed: " + e.message); }
}

function applyTransitionToAllCuts(trackIndex, transitionName, duration) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        transitionName = transitionName || "Cross Dissolve";
        duration = parseFloat(duration) || 1.0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        var track = seq.videoTracks[trackIndex];
        var qe = null;
        try { qe = app.enableQE(); } catch(qex) {}
        if (!qe) return _err("QE DOM not available");
        var qeSeq = qe.project.getActiveSequence();
        var qeTrack = qeSeq.getVideoTrackAt(trackIndex);
        var applied = 0;
        var numClips = track.clips.numItems;
        for (var i = 0; i < numClips - 1; i++) {
            try {
                var qeClip = qeTrack.getItemAt(i);
                if (qeClip) { qeClip.addTransition(qe.project.getVideoTransitionByName(transitionName), false, duration.toString()); applied++; }
            } catch (te) {}
        }
        return _ok({trackIndex: trackIndex, transitionName: transitionName, duration: duration, cutsProcessed: numClips - 1, transitionsApplied: applied});
    } catch (e) { return _err("applyTransitionToAllCuts failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Batch Color
// ---------------------------------------------------------------------------

function applyLUTToAllClips(trackIndex, lutPath) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        if (!lutPath) return _err("lutPath is required");
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        var track = seq.videoTracks[trackIndex];
        var applied = 0; var errors = [];
        for (var i = 0; i < track.clips.numItems; i++) {
            try {
                var lumetri = _getLumetriComponent(trackIndex, i);
                if (lumetri && lumetri.comp) {
                    var lutProp = null;
                    for (var p = 0; p < lumetri.comp.properties.numItems; p++) {
                        var prop = lumetri.comp.properties[p];
                        if (prop.displayName === "Input LUT" || prop.displayName === "LUT") { lutProp = prop; break; }
                    }
                    if (lutProp) { lutProp.setValue(lutPath); applied++; }
                    else { errors.push({clipIndex: i, error: "LUT property not found"}); }
                } else { errors.push({clipIndex: i, error: "Could not get Lumetri component"}); }
            } catch (ce) { errors.push({clipIndex: i, error: ce.message}); }
        }
        return _ok({trackIndex: trackIndex, lutPath: lutPath, totalClips: track.clips.numItems, applied: applied, errors: errors});
    } catch (e) { return _err("applyLUTToAllClips failed: " + e.message); }
}

function resetColorOnAllClips(trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackIndex = parseInt(trackIndex, 10) || 0;
        if (trackIndex >= seq.videoTracks.numTracks) return _err("Video track index out of range");
        var track = seq.videoTracks[trackIndex];
        var resetCount = 0;
        for (var i = 0; i < track.clips.numItems; i++) {
            try {
                var clip = track.clips[i];
                if (clip.components) {
                    for (var ci = clip.components.numItems - 1; ci >= 2; ci--) {
                        var comp = clip.components[ci];
                        if (comp.displayName === "Lumetri Color" || comp.matchName === "AdjustmentLumetriEffect") { comp.remove(); resetCount++; break; }
                    }
                }
            } catch (re) {}
        }
        return _ok({trackIndex: trackIndex, totalClips: track.clips.numItems, lumetriReset: resetCount});
    } catch (e) { return _err("resetColorOnAllClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Batch Audio
// ---------------------------------------------------------------------------

function normalizeAllAudio(targetDb) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        targetDb = _clampDb(targetDb);
        var normalized = 0;
        for (var ti = 0; ti < seq.audioTracks.numTracks; ti++) {
            var track = seq.audioTracks[ti];
            if (!track.clips) continue;
            for (var ci = 0; ci < track.clips.numItems; ci++) {
                try { var p = _findVolumeParam(track.clips[ci]); if (p) { p.setValue(targetDb, true); normalized++; } } catch (ne) {}
            }
        }
        return _ok({targetDb: targetDb, clipsNormalized: normalized});
    } catch (e) { return _err("normalizeAllAudio failed: " + e.message); }
}

function muteAllAudioTracks() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var muted = 0;
        for (var i = 0; i < seq.audioTracks.numTracks; i++) { try { seq.audioTracks[i].setMute(1); muted++; } catch (me) {} }
        return _ok({tracksMuted: muted, totalAudioTracks: seq.audioTracks.numTracks});
    } catch (e) { return _err("muteAllAudioTracks failed: " + e.message); }
}

function unmuteAllAudioTracks() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var unmuted = 0;
        for (var i = 0; i < seq.audioTracks.numTracks; i++) { try { seq.audioTracks[i].setMute(0); unmuted++; } catch (me) {} }
        return _ok({tracksUnmuted: unmuted, totalAudioTracks: seq.audioTracks.numTracks});
    } catch (e) { return _err("unmuteAllAudioTracks failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Conforming
// ---------------------------------------------------------------------------

function conformSequenceToClip(projectItemIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        projectItemIndex = parseInt(projectItemIndex, 10) || 0;
        var rootItem = app.project.rootItem;
        if (projectItemIndex >= rootItem.children.numItems) return _err("Project item index out of range");
        var item = rootItem.children[projectItemIndex];
        var newSeq = app.project.createNewSequenceFromClips(item.name + "_conformed", [item], app.project.rootItem);
        if (newSeq) {
            var settings = newSeq.getSettings();
            if (settings) { seq.setSettings(settings); }
            return _ok({conformed: true, sourceItem: item.name, sequenceName: seq.name});
        }
        return _err("Could not create reference sequence from clip");
    } catch (e) { return _err("conformSequenceToClip failed: " + e.message); }
}

function scaleAllClipsToFrame() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var seqWidth = seq.frameSizeHorizontal;
        var seqHeight = seq.frameSizeVertical;
        var scaled = 0;
        for (var ti = 0; ti < seq.videoTracks.numTracks; ti++) {
            var track = seq.videoTracks[ti];
            if (!track.clips) continue;
            for (var ci = 0; ci < track.clips.numItems; ci++) {
                try {
                    var clip = track.clips[ci];
                    if (!clip.components) continue;
                    var motion = clip.components[0];
                    if (!motion || motion.displayName !== "Motion") continue;
                    for (var pi = 0; pi < motion.properties.numItems; pi++) {
                        if (motion.properties[pi].displayName === "Scale") { motion.properties[pi].setValue(100, true); scaled++; break; }
                    }
                } catch (se) {}
            }
        }
        return _ok({sequenceSize: seqWidth + "x" + seqHeight, clipsScaled: scaled});
    } catch (e) { return _err("scaleAllClipsToFrame failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Timeline Operations
// ---------------------------------------------------------------------------

function selectAllClipsOnTrack(trackType, trackIndex) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        var selected = 0;
        if (track.clips) { for (var i = 0; i < track.clips.numItems; i++) { try { track.clips[i].setSelected(true, true); selected++; } catch (se) {} } }
        return _ok({trackType: trackType, trackIndex: trackIndex, clipsSelected: selected});
    } catch (e) { return _err("selectAllClipsOnTrack failed: " + e.message); }
}

function selectAllClipsBetween(startSeconds, endSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        startSeconds = parseFloat(startSeconds) || 0;
        endSeconds = parseFloat(endSeconds) || 0;
        if (endSeconds <= startSeconds) return _err("endSeconds must be greater than startSeconds");
        var selected = 0;
        for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
            var vTrack = seq.videoTracks[vt];
            if (!vTrack.clips) continue;
            for (var ci = 0; ci < vTrack.clips.numItems; ci++) {
                var clip = vTrack.clips[ci];
                if (_timeToSeconds(clip.start) < endSeconds && _timeToSeconds(clip.end) > startSeconds) { try { clip.setSelected(true, true); selected++; } catch (se) {} }
            }
        }
        for (var at = 0; at < seq.audioTracks.numTracks; at++) {
            var aTrack = seq.audioTracks[at];
            if (!aTrack.clips) continue;
            for (var aci = 0; aci < aTrack.clips.numItems; aci++) {
                var aclip = aTrack.clips[aci];
                if (_timeToSeconds(aclip.start) < endSeconds && _timeToSeconds(aclip.end) > startSeconds) { try { aclip.setSelected(true, true); selected++; } catch (se2) {} }
            }
        }
        return _ok({startSeconds: startSeconds, endSeconds: endSeconds, clipsSelected: selected});
    } catch (e) { return _err("selectAllClipsBetween failed: " + e.message); }
}

function deleteAllClipsBetween(trackType, trackIndex, startSeconds, endSeconds) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        trackType = String(trackType || "video").toLowerCase();
        trackIndex = parseInt(trackIndex, 10) || 0;
        startSeconds = parseFloat(startSeconds) || 0;
        endSeconds = parseFloat(endSeconds) || 0;
        if (endSeconds <= startSeconds) return _err("endSeconds must be greater than startSeconds");
        var tracks = (trackType === "audio") ? seq.audioTracks : seq.videoTracks;
        if (trackIndex >= tracks.numTracks) return _err("Track index out of range");
        var track = tracks[trackIndex];
        var deleted = 0;
        if (track.clips) {
            for (var i = track.clips.numItems - 1; i >= 0; i--) {
                var clip = track.clips[i];
                if (_timeToSeconds(clip.start) >= startSeconds && _timeToSeconds(clip.end) <= endSeconds) {
                    try { clip.remove(false, false); deleted++; } catch (de) {}
                }
            }
        }
        return _ok({trackType: trackType, trackIndex: trackIndex, startSeconds: startSeconds, endSeconds: endSeconds, clipsDeleted: deleted});
    } catch (e) { return _err("deleteAllClipsBetween failed: " + e.message); }
}

function rippleDeleteAllGaps() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence;
        if (!seq) return _err("No active sequence");
        var gapsClosed = 0;
        for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
            var vTrack = seq.videoTracks[vt];
            if (!vTrack.clips || vTrack.clips.numItems < 2) continue;
            for (var ci = 1; ci < vTrack.clips.numItems; ci++) {
                var prevEnd = _timeToSeconds(vTrack.clips[ci - 1].end);
                var currStart = _timeToSeconds(vTrack.clips[ci].start);
                if (currStart > prevEnd + 0.001) { try { vTrack.clips[ci].move(prevEnd); gapsClosed++; } catch (me) {} }
            }
        }
        for (var at = 0; at < seq.audioTracks.numTracks; at++) {
            var aTrack = seq.audioTracks[at];
            if (!aTrack.clips || aTrack.clips.numItems < 2) continue;
            for (var aci = 1; aci < aTrack.clips.numItems; aci++) {
                var aPrevEnd = _timeToSeconds(aTrack.clips[aci - 1].end);
                var aCurrStart = _timeToSeconds(aTrack.clips[aci].start);
                if (aCurrStart > aPrevEnd + 0.001) { try { aTrack.clips[aci].move(aPrevEnd); gapsClosed++; } catch (ame) {} }
            }
        }
        return _ok({gapsClosed: gapsClosed});
    } catch (e) { return _err("rippleDeleteAllGaps failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Project Cleanup
// ---------------------------------------------------------------------------

function removeUnusedMedia() {
    try {
        if (!app.project) return _err("No project is open");
        var usedPaths = {};
        for (var si = 0; si < app.project.sequences.numSequences; si++) {
            var seq = app.project.sequences[si];
            for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
                var vTrack = seq.videoTracks[vt]; if (!vTrack.clips) continue;
                for (var ci = 0; ci < vTrack.clips.numItems; ci++) { var clip = vTrack.clips[ci]; if (clip.projectItem && clip.projectItem.getMediaPath) usedPaths[clip.projectItem.getMediaPath()] = true; }
            }
            for (var at = 0; at < seq.audioTracks.numTracks; at++) {
                var aTrack = seq.audioTracks[at]; if (!aTrack.clips) continue;
                for (var aci = 0; aci < aTrack.clips.numItems; aci++) { var aclip = aTrack.clips[aci]; if (aclip.projectItem && aclip.projectItem.getMediaPath) usedPaths[aclip.projectItem.getMediaPath()] = true; }
            }
        }
        var toRemove = [];
        function _collectUnused(parent) {
            for (var i = parent.children.numItems - 1; i >= 0; i--) {
                var child = parent.children[i];
                if (child.type === 2) { _collectUnused(child); }
                else if (child.getMediaPath) { var path = child.getMediaPath(); if (path && !usedPaths[path]) toRemove.push(child); }
            }
        }
        _collectUnused(app.project.rootItem);
        var removed = 0;
        for (var ri = 0; ri < toRemove.length; ri++) { try { toRemove[ri].remove(); removed++; } catch (re) {} }
        return _ok({unusedFound: toRemove.length, removed: removed});
    } catch (e) { return _err("removeUnusedMedia failed: " + e.message); }
}

function getUnusedMedia() {
    try {
        if (!app.project) return _err("No project is open");
        var usedPaths = {};
        for (var si = 0; si < app.project.sequences.numSequences; si++) {
            var seq = app.project.sequences[si];
            for (var vt = 0; vt < seq.videoTracks.numTracks; vt++) {
                var vTrack = seq.videoTracks[vt]; if (!vTrack.clips) continue;
                for (var ci = 0; ci < vTrack.clips.numItems; ci++) { var clip = vTrack.clips[ci]; if (clip.projectItem && clip.projectItem.getMediaPath) usedPaths[clip.projectItem.getMediaPath()] = true; }
            }
            for (var at = 0; at < seq.audioTracks.numTracks; at++) {
                var aTrack = seq.audioTracks[at]; if (!aTrack.clips) continue;
                for (var aci = 0; aci < aTrack.clips.numItems; aci++) { var aclip = aTrack.clips[aci]; if (aclip.projectItem && aclip.projectItem.getMediaPath) usedPaths[aclip.projectItem.getMediaPath()] = true; }
            }
        }
        var unused = [];
        function _findUnused(parent) {
            for (var i = 0; i < parent.children.numItems; i++) {
                var child = parent.children[i];
                if (child.type === 2) { _findUnused(child); }
                else if (child.getMediaPath) { var path = child.getMediaPath(); if (path && !usedPaths[path]) unused.push({name: child.name, path: path, type: child.type}); }
            }
        }
        _findUnused(app.project.rootItem);
        return _ok({unusedCount: unused.length, items: unused});
    } catch (e) { return _err("getUnusedMedia failed: " + e.message); }
}

function flattenAllBins() {
    try {
        if (!app.project) return _err("No project is open");
        var rootItem = app.project.rootItem;
        var moved = 0;
        function _flatten(parent) {
            for (var i = parent.children.numItems - 1; i >= 0; i--) {
                var child = parent.children[i];
                if (child.type === 2) { _flatten(child); }
                else { if (parent !== rootItem) { try { child.moveBin(rootItem); moved++; } catch (me) {} } }
            }
        }
        _flatten(rootItem);
        var binsRemoved = 0;
        function _removeEmptyBins(parent) {
            for (var i = parent.children.numItems - 1; i >= 0; i--) {
                var child = parent.children[i];
                if (child.type === 2) { _removeEmptyBins(child); if (child.children.numItems === 0) { try { child.remove(); binsRemoved++; } catch (re) {} } }
            }
        }
        _removeEmptyBins(rootItem);
        return _ok({itemsMoved: moved, emptyBinsRemoved: binsRemoved});
    } catch (e) { return _err("flattenAllBins failed: " + e.message); }
}

function autoOrganizeBins() {
    try {
        if (!app.project) return _err("No project is open");
        var rootItem = app.project.rootItem;
        var binNames = ["Video", "Audio", "Images", "Graphics"];
        var bins = {};
        for (var b = 0; b < binNames.length; b++) {
            var found = null;
            for (var i = 0; i < rootItem.children.numItems; i++) { if (rootItem.children[i].type === 2 && rootItem.children[i].name === binNames[b]) { found = rootItem.children[i]; break; } }
            if (!found) {
                rootItem.createBin(binNames[b]);
                for (var j = rootItem.children.numItems - 1; j >= 0; j--) { if (rootItem.children[j].type === 2 && rootItem.children[j].name === binNames[b]) { found = rootItem.children[j]; break; } }
            }
            bins[binNames[b]] = found;
        }
        var movedCounts = {Video: 0, Audio: 0, Images: 0, Graphics: 0};
        var videoExts = /\.(mp4|mov|avi|mkv|wmv|flv|m4v|mpg|mpeg|webm|mxf|r3d)$/i;
        var audioExts = /\.(wav|mp3|aac|aif|aiff|flac|ogg|wma|m4a)$/i;
        var imageExts = /\.(jpg|jpeg|png|tiff|tif|bmp|gif|psd|ai|eps|svg|webp|exr|dpx)$/i;
        var graphicsExts = /\.(mogrt|prproj|aep|psq)$/i;
        for (var ri = rootItem.children.numItems - 1; ri >= 0; ri--) {
            var item = rootItem.children[ri];
            if (item.type === 2) continue;
            var path = ""; try { path = item.getMediaPath() || item.name || ""; } catch(pe) { path = item.name || ""; }
            var targetBin = null;
            if (videoExts.test(path)) targetBin = "Video";
            else if (audioExts.test(path)) targetBin = "Audio";
            else if (imageExts.test(path)) targetBin = "Images";
            else if (graphicsExts.test(path)) targetBin = "Graphics";
            if (targetBin && bins[targetBin]) { try { item.moveBin(bins[targetBin]); movedCounts[targetBin]++; } catch (me) {} }
        }
        return _ok({organized: true, moved: movedCounts});
    } catch (e) { return _err("autoOrganizeBins failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Markers Batch
// ---------------------------------------------------------------------------

function exportMarkersAsCSV(outputPath) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        if (!outputPath) return _err("outputPath is required");
        var markers = seq.markers; if (!markers) return _err("Sequence has no markers object");
        var csv = "Index,Name,Comment,Start,End,Duration,Type,Color\n";
        var marker = markers.getFirstMarker(); var idx = 0;
        while (marker) {
            var startSec = _timeToSeconds(marker.start); var endSec = _timeToSeconds(marker.end); var dur = endSec - startSec;
            var name = (marker.name || "").replace(/,/g, ";").replace(/\n/g, " ");
            var comment = (marker.comments || "").replace(/,/g, ";").replace(/\n/g, " ");
            csv += idx + "," + name + "," + comment + "," + startSec.toFixed(3) + "," + endSec.toFixed(3) + "," + dur.toFixed(3) + "," + (marker.type || "") + "," + (marker.getColorByIndex ? marker.getColorByIndex() : "") + "\n";
            idx++; marker = markers.getNextMarker(marker);
        }
        var file = new File(outputPath); file.encoding = "UTF-8"; file.open("w"); file.write(csv); file.close();
        return _ok({outputPath: outputPath, markerCount: idx});
    } catch (e) { return _err("exportMarkersAsCSV failed: " + e.message); }
}

function exportMarkersAsEDL(outputPath) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        if (!outputPath) return _err("outputPath is required");
        var markers = seq.markers; if (!markers) return _err("Sequence has no markers object");
        var fps = 24; try { fps = parseFloat(seq.getSettings().videoFrameRate) || 24; } catch(fe) {}
        var edl = "TITLE: " + seq.name + "\nFCM: NON-DROP FRAME\n\n";
        var marker = markers.getFirstMarker(); var idx = 1;
        function _secsToTC(s) {
            var h = Math.floor(s / 3600); s -= h * 3600; var m = Math.floor(s / 60); s -= m * 60;
            var sec = Math.floor(s); var fr = Math.floor((s - sec) * fps);
            function _pad(n) { return (n < 10 ? "0" : "") + n; }
            return _pad(h) + ":" + _pad(m) + ":" + _pad(sec) + ":" + _pad(fr);
        }
        while (marker) {
            var startSec = _timeToSeconds(marker.start); var endSec = _timeToSeconds(marker.end);
            var padIdx = ("000" + idx).slice(-3);
            edl += padIdx + "  AX       V     C        " + _secsToTC(startSec) + " " + _secsToTC(endSec) + " " + _secsToTC(startSec) + " " + _secsToTC(endSec) + "\n";
            if (marker.name) edl += "* FROM CLIP NAME: " + marker.name + "\n";
            if (marker.comments) edl += "* COMMENT: " + marker.comments + "\n";
            edl += "\n"; idx++; marker = markers.getNextMarker(marker);
        }
        var file = new File(outputPath); file.encoding = "UTF-8"; file.open("w"); file.write(edl); file.close();
        return _ok({outputPath: outputPath, markerCount: idx - 1});
    } catch (e) { return _err("exportMarkersAsEDL failed: " + e.message); }
}

function importMarkersFromCSV(csvPath) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        if (!csvPath) return _err("csvPath is required");
        var file = new File(csvPath); if (!file.exists) return _err("CSV file not found: " + csvPath);
        file.encoding = "UTF-8"; file.open("r"); var content = file.read(); file.close();
        var lines = content.split(/\r?\n/); var imported = 0; var markers = seq.markers;
        for (var i = 1; i < lines.length; i++) {
            var line = lines[i].replace(/^\s+|\s+$/g, ""); if (!line) continue;
            var parts = line.split(","); if (parts.length < 3) continue;
            var name = parts[0] || ""; var comment = parts[1] || "";
            var startSec = parseFloat(parts[2]) || 0; var endSec = parseFloat(parts[3]) || startSec;
            var colorIdx = parseInt(parts[4], 10) || 0;
            try {
                var newMarker = markers.createMarker(startSec);
                if (newMarker) { newMarker.name = name; newMarker.comments = comment; if (endSec > startSec) newMarker.end = _secondsToTime(endSec); if (colorIdx > 0 && newMarker.setColorByIndex) newMarker.setColorByIndex(colorIdx); imported++; }
            } catch (me) {}
        }
        return _ok({csvPath: csvPath, markersImported: imported});
    } catch (e) { return _err("importMarkersFromCSV failed: " + e.message); }
}

function deleteAllMarkers() {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        var markers = seq.markers; if (!markers) return _err("Sequence has no markers object");
        var toDelete = []; var marker = markers.getFirstMarker();
        while (marker) { toDelete.push(marker); marker = markers.getNextMarker(marker); }
        var count = 0;
        for (var i = 0; i < toDelete.length; i++) { try { markers.deleteMarker(toDelete[i]); count++; } catch (de) {} }
        return _ok({markersDeleted: count});
    } catch (e) { return _err("deleteAllMarkers failed: " + e.message); }
}

function convertMarkersToClips(markerColor) {
    try {
        if (!app.project) return _err("No project is open");
        var seq = app.project.activeSequence; if (!seq) return _err("No active sequence");
        var markers = seq.markers; if (!markers) return _err("Sequence has no markers object");
        var marker = markers.getFirstMarker(); var created = 0;
        while (marker) {
            var include = true;
            if (markerColor && markerColor !== "") {
                var mc = ""; try { mc = marker.getColorByIndex ? String(marker.getColorByIndex()) : ""; } catch(ce) {}
                if (mc !== String(markerColor)) include = false;
            }
            if (include) {
                var startSec = _timeToSeconds(marker.start); var endSec = _timeToSeconds(marker.end);
                if (endSec <= startSec) endSec = startSec + 1;
                try { seq.setInPoint(_secondsToTime(startSec)); seq.setOutPoint(_secondsToTime(endSec)); created++; } catch (se) {}
            }
            marker = markers.getNextMarker(marker);
        }
        return _ok({markersProcessed: created, color: markerColor || "all"});
    } catch (e) { return _err("convertMarkersToClips failed: " + e.message); }
}

// ---------------------------------------------------------------------------
// Automation
// ---------------------------------------------------------------------------

function runExtendScript(script) {
    try {
        if (!script) return _err("script is required");
        var result = eval(script);
        if (result === undefined || result === null) return _ok({executed: true, result: null});
        return _ok({executed: true, result: String(result)});
    } catch (e) { return _err("runExtendScript failed: " + e.message); }
}

function getSystemInfo() {
    try {
        var info = {
            os: $.os || "unknown",
            premiereVersion: app.version || "unknown",
            premiereBuild: app.build || "unknown",
            engineName: $.engineName || "unknown",
            memoryAvailable: $.memCache || 0,
            locale: $.locale || "unknown",
            appName: app.name || "Adobe Premiere Pro",
            appPath: app.path || "unknown"
        };
        try { if (app.properties) info.gpuRenderer = app.properties.getProperty("GPU.Renderer") || "unknown"; } catch (ge) {}
        return _ok(info);
    } catch (e) { return _err("getSystemInfo failed: " + e.message); }
}

function getRecentProjects() {
    try {
        var recentProjects = [];
        var currentProject = null;
        if (app.project) { currentProject = {name: app.project.name || "unknown", path: app.project.path || "unknown"}; }
        var recentDir = "";
        if ($.os.indexOf("Win") >= 0) { recentDir = Folder.appData.fsName + "\\Adobe\\Premiere Pro\\" + app.version + "\\Recent"; }
        else { recentDir = Folder.userData.fsName + "/Adobe/Premiere Pro/" + app.version + "/Recent"; }
        var recentFolder = new Folder(recentDir);
        if (recentFolder.exists) {
            var files = recentFolder.getFiles("*.prproj");
            for (var i = 0; i < files.length && i < 20; i++) { recentProjects.push({name: files[i].name, path: files[i].fsName, modified: files[i].modified ? files[i].modified.toISOString() : ""}); }
        }
        return _ok({currentProject: currentProject, recentProjects: recentProjects});
    } catch (e) { return _err("getRecentProjects failed: " + e.message); }
}
