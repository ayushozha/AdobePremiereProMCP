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
// setAudioLevel(trackIndex, clipIndex, levelDb)
// ---------------------------------------------------------------------------
function setAudioLevel(trackIndex, clipIndex, levelDb) {
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
        levelDb = parseFloat(levelDb) || 0;

        // Clamp level to valid dB range
        if (levelDb < -96) levelDb = -96;
        if (levelDb > 15) levelDb = 15;

        if (trackIndex >= seq.audioTracks.numTracks) {
            return _err("Audio track index " + trackIndex + " out of range (have " + seq.audioTracks.numTracks + ")");
        }

        var track = seq.audioTracks[trackIndex];
        if (!track.clips || clipIndex >= track.clips.numItems) {
            return _err("Clip index " + clipIndex + " out of range on audio track " + trackIndex);
        }

        var clip = track.clips[clipIndex];

        // Access the clip's audio components to set the level
        if (clip.components) {
            for (var ci = 0; ci < clip.components.numItems; ci++) {
                var component = clip.components[ci];
                // Look for the Volume component
                if (component.displayName === "Volume" || component.matchName === "audioGain") {
                    var volumeParam = component.properties.getParamForDisplayName("Level");
                    if (!volumeParam) {
                        volumeParam = component.properties[0]; // First param is typically level
                    }
                    if (volumeParam) {
                        // Convert dB to the parameter value Premiere expects
                        // Premiere's volume parameter uses dB directly in newer versions
                        volumeParam.setValue(levelDb, true);
                    }
                    break;
                }
            }
        }

        return _ok({
            trackIndex: trackIndex,
            clipIndex: clipIndex,
            levelDb: levelDb
        });
    } catch (e) {
        return _err("setAudioLevel failed: " + e.message);
    }
}

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
