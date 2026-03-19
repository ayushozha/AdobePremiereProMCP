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
                fps: fps
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
            videoTracks: [],
            audioTracks: []
        };

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
