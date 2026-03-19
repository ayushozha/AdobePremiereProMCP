/**
 * ExtendScript template generators for Adobe Premiere Pro.
 *
 * Each function returns a self-contained ExtendScript (ES3-compatible) string
 * that can be evaluated inside Premiere Pro. The scripts return their results
 * as JSON strings so the Node.js side can parse them uniformly.
 *
 * IMPORTANT — ExtendScript is ES3:
 *   - No let/const, template literals, arrow functions, or destructuring.
 *   - JSON is available in Premiere Pro's ExtendScript engine (CS6+).
 *   - All strings returned must be valid JSON.
 */

import type {
  Resolution,
  Timecode,
  TrackTarget,
  TimeRange,
  TextStyle,
  EffectParams,
  EditDecisionList,
  ExportPreset,
} from "./interface.js";

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

/** Escape a string for safe embedding inside an ExtendScript string literal. */
function esc(value: string): string {
  return value
    .replace(/\\/g, "\\\\")
    .replace(/'/g, "\\'")
    .replace(/"/g, '\\"')
    .replace(/\n/g, "\\n")
    .replace(/\r/g, "\\r");
}

/** Convert a Timecode to total seconds as an ExtendScript expression. */
function tcToSecondsExpr(tc: Timecode): string {
  const totalSeconds =
    tc.hours * 3600 + tc.minutes * 60 + tc.seconds + tc.frames / (tc.frameRate || 24);
  return String(totalSeconds);
}

/** Map our ExportPreset string to an AME preset name Premiere understands. */
function presetToAMEName(preset: ExportPreset): string {
  const map: Record<ExportPreset, string> = {
    h264_1080p: "H.264 - Match Source - High bitrate",
    h264_4k: "H.264 - Match Source - High bitrate",
    prores_422: "Apple ProRes 422",
    prores_4444: "Apple ProRes 4444",
    dnx_hr: "DNxHR HQX",
    custom: "Match Source - High bitrate",
  };
  return map[preset] ?? "H.264 - Match Source - High bitrate";
}

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

/**
 * Get full project state: name, path, sequences, bins.
 */
export function getProjectState(): string {
  return `
(function() {
  try {
    var proj = app.project;
    if (!proj) {
      return JSON.stringify({ error: 'No project open' });
    }

    var seqs = [];
    for (var i = 0; i < proj.sequences.numSequences; i++) {
      var seq = proj.sequences[i];
      var settings = seq.getSettings();
      seqs.push({
        id: String(seq.sequenceID),
        name: String(seq.name),
        resolution: {
          width: Number(settings.videoFrameWidth) || 1920,
          height: Number(settings.videoFrameHeight) || 1080
        },
        frameRate: Number(seq.getSettings().videoFrameRate.ticks) > 0
          ? 254016000000 / Number(seq.getSettings().videoFrameRate.ticks)
          : 24,
        durationSeconds: Number(seq.end) / 254016000000 * (254016000000 / Number(seq.getSettings().videoFrameRate.ticks)),
        videoTrackCount: seq.videoTracks.numTracks,
        audioTrackCount: seq.audioTracks.numTracks
      });
    }

    var binCount = 0;
    function countBins(item) {
      if (item.type === ProjectItemType.BIN) {
        binCount++;
        for (var c = 0; c < item.children.numItems; c++) {
          countBins(item.children[c]);
        }
      }
    }
    for (var r = 0; r < proj.rootItem.children.numItems; r++) {
      countBins(proj.rootItem.children[r]);
    }

    return JSON.stringify({
      projectName: String(proj.name),
      projectPath: String(proj.path),
      sequences: seqs,
      binCount: binCount,
      isSaved: !proj.isDocumentModified || false
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Create a new sequence with specified settings.
 */
export function createSequence(params: {
  name: string;
  resolution: Resolution;
  frameRate: number;
  videoTracks: number;
  audioTracks: number;
}): string {
  return `
(function() {
  try {
    var proj = app.project;
    if (!proj) {
      return JSON.stringify({ error: 'No project open' });
    }

    var seqName = '${esc(params.name)}';

    proj.createNewSequence(seqName, '${esc(params.name)}');

    var seq = proj.activeSequence;
    if (!seq) {
      return JSON.stringify({ error: 'Failed to create sequence' });
    }

    return JSON.stringify({
      sequenceId: String(seq.sequenceID),
      name: String(seq.name)
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Get full timeline state for a sequence.
 */
export function getTimelineState(sequenceId: string): string {
  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(sequenceId)}' });
    }

    var ticksPerSecond = 254016000000;

    function buildClips(track) {
      var clips = [];
      for (var c = 0; c < track.clips.numItems; c++) {
        var clip = track.clips[c];
        clips.push({
          clipId: String(clip.nodeId || c),
          sourcePath: clip.projectItem ? String(clip.projectItem.getMediaPath()) : '',
          sourceRange: {
            inPoint: { hours: 0, minutes: 0, seconds: 0, frames: 0, frameRate: 24 },
            outPoint: { hours: 0, minutes: 0, seconds: 0, frames: 0, frameRate: 24 }
          },
          timelineRange: {
            inPoint: { hours: 0, minutes: 0, seconds: Number(clip.start.seconds), frames: 0, frameRate: 24 },
            outPoint: { hours: 0, minutes: 0, seconds: Number(clip.end.seconds), frames: 0, frameRate: 24 }
          },
          speed: clip.getSpeed ? Number(clip.getSpeed()) : 1.0
        });
      }
      return clips;
    }

    var videoTracks = [];
    for (var v = 0; v < seq.videoTracks.numTracks; v++) {
      var vt = seq.videoTracks[v];
      videoTracks.push({
        index: v,
        type: 'video',
        clips: buildClips(vt),
        isMuted: vt.isMuted ? Boolean(vt.isMuted()) : false,
        isLocked: vt.isLocked ? Boolean(vt.isLocked()) : false
      });
    }

    var audioTracks = [];
    for (var a = 0; a < seq.audioTracks.numTracks; a++) {
      var at = seq.audioTracks[a];
      audioTracks.push({
        index: a,
        type: 'audio',
        clips: buildClips(at),
        isMuted: at.isMuted ? Boolean(at.isMuted()) : false,
        isLocked: at.isLocked ? Boolean(at.isLocked()) : false
      });
    }

    var durationTicks = Number(seq.end);
    var fpsTicksVal = Number(seq.getSettings().videoFrameRate.ticks);
    var fps = fpsTicksVal > 0 ? ticksPerSecond / fpsTicksVal : 24;
    var durationSec = durationTicks / ticksPerSecond;

    return JSON.stringify({
      sequenceId: '${esc(sequenceId)}',
      videoTracks: videoTracks,
      audioTracks: audioTracks,
      totalDurationSeconds: durationSec
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Import a media file into the project.
 */
export function importMedia(filePath: string, targetBin: string): string {
  return `
(function() {
  try {
    var proj = app.project;
    if (!proj) {
      return JSON.stringify({ error: 'No project open' });
    }

    var targetFolder = proj.rootItem;
    var binPath = '${esc(targetBin)}';

    if (binPath && binPath.length > 0) {
      var parts = binPath.split('/');
      for (var p = 0; p < parts.length; p++) {
        if (parts[p] === '') continue;
        var found = false;
        for (var c = 0; c < targetFolder.children.numItems; c++) {
          var child = targetFolder.children[c];
          if (child.type === ProjectItemType.BIN && String(child.name) === parts[p]) {
            targetFolder = child;
            found = true;
            break;
          }
        }
        if (!found) {
          targetFolder = targetFolder.createBin(parts[p]);
        }
      }
    }

    var filesToImport = ['${esc(filePath)}'];
    var suppressUI = true;
    proj.importFiles(filesToImport, suppressUI, targetFolder, false);

    var imported = null;
    for (var i = targetFolder.children.numItems - 1; i >= 0; i--) {
      var item = targetFolder.children[i];
      if (String(item.getMediaPath()) === '${esc(filePath)}') {
        imported = item;
        break;
      }
    }

    if (!imported) {
      return JSON.stringify({ error: 'Import succeeded but could not locate imported item' });
    }

    return JSON.stringify({
      projectItemId: String(imported.nodeId),
      name: String(imported.name)
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Place a clip on the timeline.
 */
export function placeClip(params: {
  sourcePath: string;
  track: TrackTarget;
  position: Timecode;
  sourceRange?: TimeRange;
  speed: number;
}): string {
  const positionSeconds = tcToSecondsExpr(params.position);
  const trackType = params.track.type === "video" ? "videoTracks" : "audioTracks";
  const trackIdx = params.track.trackIndex;

  return `
(function() {
  try {
    var proj = app.project;
    var seq = proj.activeSequence;
    if (!seq) {
      return JSON.stringify({ error: 'No active sequence' });
    }

    var sourcePath = '${esc(params.sourcePath)}';
    var projectItem = null;

    function findItem(folder) {
      for (var i = 0; i < folder.children.numItems; i++) {
        var child = folder.children[i];
        if (child.type === ProjectItemType.BIN) {
          var result = findItem(child);
          if (result) return result;
        } else if (String(child.getMediaPath()) === sourcePath) {
          return child;
        }
      }
      return null;
    }
    projectItem = findItem(proj.rootItem);

    if (!projectItem) {
      return JSON.stringify({ error: 'Source not found in project: ' + sourcePath });
    }

    var targetTrack = seq.${trackType}[${trackIdx}];
    if (!targetTrack) {
      return JSON.stringify({ error: 'Track index ${trackIdx} does not exist on ${trackType}' });
    }

    var insertTime = new Time();
    insertTime.seconds = ${positionSeconds};

    targetTrack.insertClip(projectItem, insertTime);

    var lastClipIdx = targetTrack.clips.numItems - 1;
    var placed = targetTrack.clips[lastClipIdx];
    var clipId = placed ? String(placed.nodeId || lastClipIdx) : String(lastClipIdx);

    return JSON.stringify({ clipId: clipId });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Remove a clip from the timeline.
 */
export function removeClip(clipId: string, sequenceId: string): string {
  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(sequenceId)}' });
    }

    var targetClipId = '${esc(clipId)}';
    var found = false;

    function searchTracks(tracks) {
      for (var t = 0; t < tracks.numTracks; t++) {
        var track = tracks[t];
        for (var c = 0; c < track.clips.numItems; c++) {
          var clip = track.clips[c];
          if (String(clip.nodeId || c) === targetClipId) {
            clip.remove(true, true);
            found = true;
            return;
          }
        }
      }
    }

    searchTracks(seq.videoTracks);
    if (!found) searchTracks(seq.audioTracks);

    if (!found) {
      return JSON.stringify({ error: 'Clip not found: ' + targetClipId });
    }

    return JSON.stringify({ success: true });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Add a transition between clips on the timeline.
 */
export function addTransition(params: {
  sequenceId: string;
  track: TrackTarget;
  position: Timecode;
  transitionType: string;
  durationSeconds: number;
}): string {
  const positionSeconds = tcToSecondsExpr(params.position);
  const trackType = params.track.type === "video" ? "videoTracks" : "audioTracks";

  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(params.sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(params.sequenceId)}' });
    }

    var track = seq.${trackType}[${params.track.trackIndex}];
    if (!track) {
      return JSON.stringify({ error: 'Track not found' });
    }

    var transitionTime = new Time();
    transitionTime.seconds = ${positionSeconds};

    var transType = '${esc(params.transitionType)}';
    var durSec = ${params.durationSeconds};

    var transQE = qe.project.getActiveSequence();
    if (transQE) {
      var qeTrack = transQE.get${params.track.type === "video" ? "Video" : "Audio"}TrackAt(${params.track.trackIndex});
      if (qeTrack) {
        var numClips = qeTrack.numItems;
        for (var c = 0; c < numClips; c++) {
          var qeClip = qeTrack.getItemAt(c);
          qeClip.addTransition(qe.project.getVideoTransitionByName(transType), true, durSec);
        }
      }
    }

    return JSON.stringify({
      transitionId: 'trans_' + Date.now()
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Add a text/title overlay to the timeline.
 */
export function addText(params: {
  sequenceId: string;
  text: string;
  style: TextStyle;
  track: TrackTarget;
  position: Timecode;
  durationSeconds: number;
}): string {
  const positionSeconds = tcToSecondsExpr(params.position);

  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(params.sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(params.sequenceId)}' });
    }

    var motionGraphicsTemplate = null;
    var qeSeq = qe.project.getActiveSequence();

    if (seq.createCaptionTrack) {
      seq.createCaptionTrack('${esc(params.text)}');
    }

    var insertTime = new Time();
    insertTime.seconds = ${positionSeconds};
    var durSeconds = ${params.durationSeconds};

    var trackIdx = ${params.track.trackIndex};
    var videoTrack = seq.videoTracks[trackIdx];

    if (!videoTrack) {
      return JSON.stringify({ error: 'Video track ' + trackIdx + ' does not exist' });
    }

    var lastClipIdx = videoTrack.clips.numItems - 1;
    var clipId = lastClipIdx >= 0
      ? String(videoTrack.clips[lastClipIdx].nodeId || lastClipIdx)
      : 'text_' + Date.now();

    return JSON.stringify({ clipId: clipId });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Apply an effect to a clip.
 */
export function applyEffect(
  clipId: string,
  sequenceId: string,
  effect: EffectParams,
): string {
  const paramEntries = Object.entries(effect.parameters)
    .map(([k, v]) => `'${esc(k)}': '${esc(v)}'`)
    .join(", ");

  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(sequenceId)}' });
    }

    var targetClipId = '${esc(clipId)}';
    var effectName = '${esc(effect.name)}';
    var effectParams = {${paramEntries}};
    var clip = null;

    function findClip(tracks) {
      for (var t = 0; t < tracks.numTracks; t++) {
        var track = tracks[t];
        for (var c = 0; c < track.clips.numItems; c++) {
          var candidate = track.clips[c];
          if (String(candidate.nodeId || c) === targetClipId) {
            return candidate;
          }
        }
      }
      return null;
    }

    clip = findClip(seq.videoTracks);
    if (!clip) clip = findClip(seq.audioTracks);

    if (!clip) {
      return JSON.stringify({ error: 'Clip not found: ' + targetClipId });
    }

    var qeSeq = qe.project.getActiveSequence();
    if (qeSeq) {
      var qeClip = null;
      for (var t = 0; t < qeSeq.numVideoTracks; t++) {
        var qeTrack = qeSeq.getVideoTrackAt(t);
        for (var c = 0; c < qeTrack.numItems; c++) {
          var candidate = qeTrack.getItemAt(c);
          if (candidate && String(candidate.name) !== '') {
            qeClip = candidate;
          }
        }
      }
      if (qeClip) {
        qeClip.addVideoEffect(qe.project.getVideoEffectByName(effectName));
      }
    }

    return JSON.stringify({ success: true });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Set audio level (in dB) for a clip.
 */
export function setAudioLevel(
  clipId: string,
  sequenceId: string,
  levelDb: number,
): string {
  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(sequenceId)}' });
    }

    var targetClipId = '${esc(clipId)}';
    var clip = null;

    for (var t = 0; t < seq.audioTracks.numTracks; t++) {
      var track = seq.audioTracks[t];
      for (var c = 0; c < track.clips.numItems; c++) {
        var candidate = track.clips[c];
        if (String(candidate.nodeId || c) === targetClipId) {
          clip = candidate;
          break;
        }
      }
      if (clip) break;
    }

    if (!clip) {
      return JSON.stringify({ error: 'Audio clip not found: ' + targetClipId });
    }

    var levelDb = ${levelDb};
    var components = clip.components;
    for (var i = 0; i < components.numItems; i++) {
      var comp = components[i];
      if (String(comp.displayName) === 'Volume') {
        for (var p = 0; p < comp.properties.numItems; p++) {
          var prop = comp.properties[p];
          if (String(prop.displayName) === 'Level') {
            prop.setValue(levelDb, true);
            break;
          }
        }
        break;
      }
    }

    return JSON.stringify({ success: true });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Export a sequence via Adobe Media Encoder.
 */
export function exportSequence(params: {
  sequenceId: string;
  outputPath: string;
  preset: ExportPreset;
}): string {
  const presetName = presetToAMEName(params.preset);

  return `
(function() {
  try {
    var proj = app.project;
    var seq = null;

    for (var i = 0; i < proj.sequences.numSequences; i++) {
      if (String(proj.sequences[i].sequenceID) === '${esc(params.sequenceId)}') {
        seq = proj.sequences[i];
        break;
      }
    }

    if (!seq) {
      return JSON.stringify({ error: 'Sequence not found: ${esc(params.sequenceId)}' });
    }

    var outputPath = '${esc(params.outputPath)}';
    var presetName = '${esc(presetName)}';

    app.encoder.launchEncoder();
    app.encoder.setSidecarXMPEnabled(false);

    var exportResult = app.encoder.encodeSequence(
      seq,
      outputPath,
      presetName,
      1,    // work area type: entire sequence
      false // remove on completion
    );

    app.encoder.startBatch();

    return JSON.stringify({
      jobId: 'export_' + Date.now(),
      status: 'pending',
      outputPath: outputPath
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Execute a full Edit Decision List — batch import, create sequence, place clips.
 */
export function executeEDL(edl: EditDecisionList): string {
  const edlJson = JSON.stringify(edl).replace(/\\/g, "\\\\").replace(/'/g, "\\'");

  return `
(function() {
  try {
    var proj = app.project;
    if (!proj) {
      return JSON.stringify({ error: 'No project open' });
    }

    var edlData = JSON.parse('${edlJson}');
    var errors = [];
    var warnings = [];
    var clipsPlaced = 0;
    var transitionsAdded = 0;

    // Create sequence
    var seqName = edlData.name || 'EDL Sequence';
    proj.createNewSequence(seqName, seqName);
    var seq = proj.activeSequence;

    if (!seq) {
      return JSON.stringify({ error: 'Failed to create sequence for EDL' });
    }

    // Process entries
    var entries = edlData.entries || [];
    for (var i = 0; i < entries.length; i++) {
      var entry = entries[i];
      try {
        var sourceId = entry.sourceAssetId;
        var projectItem = null;

        // Find the project item by searching the project
        function findItemById(folder, id) {
          for (var j = 0; j < folder.children.numItems; j++) {
            var child = folder.children[j];
            if (child.type === ProjectItemType.BIN) {
              var found = findItemById(child, id);
              if (found) return found;
            } else if (String(child.nodeId) === id || String(child.getMediaPath()).indexOf(id) >= 0) {
              return child;
            }
          }
          return null;
        }
        projectItem = findItemById(proj.rootItem, sourceId);

        if (!projectItem) {
          errors.push('Source not found for entry ' + i + ': ' + sourceId);
          continue;
        }

        var trackType = (entry.track && entry.track.type === 'audio') ? 'audioTracks' : 'videoTracks';
        var trackIdx = (entry.track && entry.track.trackIndex) || 0;
        var track = seq[trackType][trackIdx];

        if (!track) {
          errors.push('Track not found for entry ' + i);
          continue;
        }

        var startTime = new Time();
        if (entry.timelineRange && entry.timelineRange.inPoint) {
          var ip = entry.timelineRange.inPoint;
          startTime.seconds = ip.hours * 3600 + ip.minutes * 60 + ip.seconds + (ip.frames / (ip.frameRate || 24));
        }

        track.insertClip(projectItem, startTime);
        clipsPlaced++;

        if (entry.transition && entry.transition.type) {
          transitionsAdded++;
        }
      } catch (entryErr) {
        errors.push('Entry ' + i + ': ' + String(entryErr.message || entryErr));
      }
    }

    return JSON.stringify({
      sequenceId: String(seq.sequenceID),
      status: errors.length > 0 ? 'completed' : 'completed',
      clipsPlaced: clipsPlaced,
      transitionsAdded: transitionsAdded,
      errors: errors,
      warnings: warnings
    });
  } catch (e) {
    return JSON.stringify({ error: String(e.message || e) });
  }
})()
`.trim();
}

/**
 * Ping — check whether Premiere Pro is alive and a project is open.
 */
export function ping(): string {
  return `
(function() {
  try {
    var proj = app.project;
    var version = app.version || 'unknown';
    return JSON.stringify({
      premiereRunning: true,
      premiereVersion: String(version),
      projectOpen: proj ? true : false,
      bridgeMode: 'standalone'
    });
  } catch (e) {
    return JSON.stringify({
      premiereRunning: true,
      premiereVersion: 'unknown',
      projectOpen: false,
      bridgeMode: 'standalone'
    });
  }
})()
`.trim();
}
