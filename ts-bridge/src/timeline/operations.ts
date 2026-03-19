/**
 * High-level timeline operations.
 *
 * These functions orchestrate multi-step Premiere Pro operations through the
 * bridge interface. They are the entry points that gRPC handlers call.
 *
 * Design principles:
 *  - If a single clip fails, log the error and continue with the remaining clips.
 *  - Emit progress updates as a percentage so callers can relay them to the UI.
 *  - Validate the EDL before touching Premiere Pro.
 */

import {
  timecodeToFrames,
  timecodeToSeconds,
  secondsToTimecode,
  formatTimecode,
  zeroTimecode,
  type Timecode,
} from "./timecode.js";

import {
  validateEDL,
  type EditDecisionList,
  type EDLEntry,
  type TrackTarget,
  type TimeRange,
  type TransitionInfo,
  type TextOverlay,
  type TextStyle,
  type Resolution,
} from "./validator.js";

// Re-export types so consumers can import everything from this module.
export type {
  EditDecisionList,
  EDLEntry,
  TrackTarget,
  TimeRange,
  TransitionInfo,
  TextOverlay,
  TextStyle,
  Resolution,
};

// ---------------------------------------------------------------------------
// Bridge interface
// ---------------------------------------------------------------------------

/**
 * Abstraction over the Premiere Pro bridge layer.
 *
 * Matches the gRPC `PremiereBridgeService` RPC surface so that the
 * operations module never depends on transport details.
 */
export interface PremiereBridge {
  createSequence(params: {
    name: string;
    resolution: Resolution;
    frameRate: number;
    videoTracks: number;
    audioTracks: number;
  }): Promise<{ sequenceId: string; name: string }>;

  importMedia(params: {
    filePath: string;
    targetBin: string;
  }): Promise<{ projectItemId: string; name: string }>;

  placeClip(params: {
    sourcePath: string;
    track: TrackTarget;
    position: Timecode;
    sourceRange: TimeRange;
    speed: number;
  }): Promise<{ clipId: string }>;

  addTransition(params: {
    sequenceId: string;
    track: TrackTarget;
    position: Timecode;
    transitionType: string;
    durationSeconds: number;
  }): Promise<{ transitionId: string }>;

  addText(params: {
    sequenceId: string;
    text: string;
    style: TextStyle;
    track: TrackTarget;
    position: Timecode;
    durationSeconds: number;
  }): Promise<{ clipId: string }>;

  setAudioLevel(params: {
    clipId: string;
    sequenceId: string;
    levelDb: number;
  }): Promise<void>;
}

// ---------------------------------------------------------------------------
// Progress tracking
// ---------------------------------------------------------------------------

/** Callback invoked with the current completion percentage (0-100). */
export type ProgressCallback = (percent: number, message: string) => void;

// ---------------------------------------------------------------------------
// Assembly result
// ---------------------------------------------------------------------------

export interface AssembleResult {
  sequenceId: string;
  sequenceName: string;
  clipsPlaced: number;
  clipsFailed: number;
  transitionsAdded: number;
  textOverlaysAdded: number;
  audioLevelsSet: number;
  errors: string[];
  warnings: string[];
}

// ---------------------------------------------------------------------------
// Asset resolver
// ---------------------------------------------------------------------------

/**
 * Maps a source asset ID from the EDL to the local file path that should
 * be imported / placed. Callers must supply this so the operations module
 * does not need to know where assets live on disk.
 */
export type AssetResolver = (assetId: string) => string | undefined;

// ---------------------------------------------------------------------------
// assembleTimeline
// ---------------------------------------------------------------------------

/**
 * Take a full Edit Decision List and execute it step-by-step in Premiere Pro.
 *
 *  1. Validate the EDL.
 *  2. Create a sequence with the EDL settings.
 *  3. Import all referenced media files.
 *  4. Place clips on correct tracks at correct positions.
 *  5. Add transitions between clips.
 *  6. Add text overlays.
 *  7. Set audio levels.
 *  8. Return a summary of what was placed.
 */
export async function assembleTimeline(
  edl: EditDecisionList,
  bridge: PremiereBridge,
  resolveAsset: AssetResolver,
  onProgress?: ProgressCallback,
): Promise<AssembleResult> {
  const progress = onProgress ?? (() => {});
  const errors: string[] = [];
  const warnings: string[] = [];

  // ------------------------------------------------------------------
  // Step 0 — Validate
  // ------------------------------------------------------------------
  progress(0, "Validating EDL");
  const validation = validateEDL(edl);

  for (const w of validation.warnings) {
    warnings.push(`[entry ${w.entryIndex}] ${w.message}`);
  }

  if (!validation.valid) {
    const errorMessages = validation.errors.map(
      (e) => `[entry ${e.entryIndex}] ${e.message}`,
    );
    throw new Error(
      `EDL validation failed with ${validation.errors.length} error(s):\n` +
        errorMessages.join("\n"),
    );
  }

  // ------------------------------------------------------------------
  // Step 1 — Create sequence
  // ------------------------------------------------------------------
  progress(5, "Creating sequence");

  const maxVideoTrack = computeMaxTrackIndex(edl.entries, "video");
  const maxAudioTrack = computeMaxTrackIndex(edl.entries, "audio");

  const { sequenceId, name: sequenceName } = await bridge.createSequence({
    name: edl.name,
    resolution: edl.sequenceResolution,
    frameRate: edl.sequenceFrameRate,
    videoTracks: maxVideoTrack + 1,
    audioTracks: maxAudioTrack + 1,
  });

  // ------------------------------------------------------------------
  // Step 2 — Import media
  // ------------------------------------------------------------------
  progress(10, "Importing media files");

  const uniqueAssetIds = new Set(edl.entries.map((e) => e.sourceAssetId));
  const importedPaths = new Map<string, string>(); // assetId -> filePath
  let importIndex = 0;

  for (const assetId of uniqueAssetIds) {
    const filePath = resolveAsset(assetId);

    if (!filePath) {
      errors.push(`Asset "${assetId}" could not be resolved to a file path.`);
      importIndex++;
      continue;
    }

    try {
      await bridge.importMedia({ filePath, targetBin: "" });
      importedPaths.set(assetId, filePath);
    } catch (err) {
      errors.push(
        `Failed to import asset "${assetId}" (${filePath}): ${errorMessage(err)}`,
      );
    }

    importIndex++;
    const importProgress = 10 + (importIndex / uniqueAssetIds.size) * 20;
    progress(Math.round(importProgress), `Imported ${importIndex}/${uniqueAssetIds.size} media files`);
  }

  // ------------------------------------------------------------------
  // Step 3 — Place clips
  // ------------------------------------------------------------------
  progress(30, "Placing clips on timeline");

  let clipsPlaced = 0;
  let clipsFailed = 0;
  const placedClipIds = new Map<number, string>(); // entry index -> clipId

  for (let i = 0; i < edl.entries.length; i++) {
    const entry = edl.entries[i];

    try {
      const result = await insertClipAtPosition(
        entry,
        importedPaths,
        bridge,
      );
      placedClipIds.set(entry.index, result.clipId);
      clipsPlaced++;
    } catch (err) {
      clipsFailed++;
      errors.push(
        `Failed to place clip for entry ${entry.index} (asset "${entry.sourceAssetId}"): ${errorMessage(err)}`,
      );
    }

    const clipProgress = 30 + ((i + 1) / edl.entries.length) * 30;
    progress(Math.round(clipProgress), `Placed ${clipsPlaced}/${edl.entries.length} clips`);
  }

  // ------------------------------------------------------------------
  // Step 4 — Add transitions
  // ------------------------------------------------------------------
  progress(60, "Adding transitions");

  let transitionsAdded = 0;
  const entriesWithTransitions = edl.entries.filter((e) => e.transition);

  for (let i = 0; i < entriesWithTransitions.length; i++) {
    const entry = entriesWithTransitions[i];

    try {
      await addTransitionBetweenClips(
        sequenceId,
        entry,
        bridge,
      );
      transitionsAdded++;
    } catch (err) {
      errors.push(
        `Failed to add transition at entry ${entry.index}: ${errorMessage(err)}`,
      );
    }

    const transProgress = 60 + ((i + 1) / entriesWithTransitions.length) * 10;
    progress(Math.round(transProgress), `Added ${transitionsAdded} transitions`);
  }

  // ------------------------------------------------------------------
  // Step 5 — Text overlays
  // ------------------------------------------------------------------
  progress(70, "Adding text overlays");

  let textOverlaysAdded = 0;
  const overlays = edl.textOverlays ?? [];

  for (let i = 0; i < overlays.length; i++) {
    const overlay = overlays[i];

    try {
      await buildTextOverlay(sequenceId, overlay, bridge);
      textOverlaysAdded++;
    } catch (err) {
      errors.push(
        `Failed to add text overlay "${overlay.text.slice(0, 30)}...": ${errorMessage(err)}`,
      );
    }

    const textProgress = 70 + ((i + 1) / overlays.length) * 15;
    progress(Math.round(textProgress), `Added ${textOverlaysAdded} text overlays`);
  }

  // ------------------------------------------------------------------
  // Step 6 — Audio levels
  // ------------------------------------------------------------------
  progress(85, "Setting audio levels");

  let audioLevelsSet = 0;

  if (edl.audioLevels) {
    const levelEntries = Array.from(edl.audioLevels.entries());

    for (let i = 0; i < levelEntries.length; i++) {
      const [entryIdx, levelDb] = levelEntries[i];
      const clipId = placedClipIds.get(entryIdx);

      if (!clipId) {
        warnings.push(
          `Skipping audio level for entry ${entryIdx}: clip was not placed.`,
        );
        continue;
      }

      try {
        await bridge.setAudioLevel({
          clipId,
          sequenceId,
          levelDb,
        });
        audioLevelsSet++;
      } catch (err) {
        errors.push(
          `Failed to set audio level for entry ${entryIdx}: ${errorMessage(err)}`,
        );
      }

      const audioProgress = 85 + ((i + 1) / levelEntries.length) * 10;
      progress(Math.round(audioProgress), `Set ${audioLevelsSet} audio levels`);
    }
  }

  // ------------------------------------------------------------------
  // Done
  // ------------------------------------------------------------------
  progress(100, "Assembly complete");

  return {
    sequenceId,
    sequenceName,
    clipsPlaced,
    clipsFailed,
    transitionsAdded,
    textOverlaysAdded,
    audioLevelsSet,
    errors,
    warnings,
  };
}

// ---------------------------------------------------------------------------
// insertClipAtPosition
// ---------------------------------------------------------------------------

/**
 * Place a single clip on the timeline with proper timecode math.
 *
 * Resolves the asset ID to a file path, computes the position in the
 * sequence frame rate, and delegates to the bridge.
 */
export async function insertClipAtPosition(
  entry: EDLEntry,
  importedPaths: Map<string, string>,
  bridge: PremiereBridge,
): Promise<{ clipId: string }> {
  const filePath = importedPaths.get(entry.sourceAssetId);

  if (!filePath) {
    throw new Error(
      `Asset "${entry.sourceAssetId}" was not imported. Cannot place clip.`,
    );
  }

  // The timeline position is the in-point of the timeline range.
  const position = entry.timelineRange.inPoint;

  return bridge.placeClip({
    sourcePath: filePath,
    track: entry.track,
    position,
    sourceRange: entry.sourceRange,
    speed: 1.0,
  });
}

// ---------------------------------------------------------------------------
// addTransitionBetweenClips
// ---------------------------------------------------------------------------

/**
 * Add a transition at the cut point described by an EDL entry.
 *
 * The transition is placed at the in-point of the entry's timeline range
 * (which is the cut point where the previous clip ends and this clip
 * begins). The transition type and duration come from the entry's
 * `transition` field.
 */
export async function addTransitionBetweenClips(
  sequenceId: string,
  entry: EDLEntry,
  bridge: PremiereBridge,
): Promise<{ transitionId: string }> {
  if (!entry.transition) {
    throw new Error(
      `Entry ${entry.index} does not specify a transition.`,
    );
  }

  // Compute the transition position: centred on the cut point. The
  // bridge/Premiere handles alignment details, so we pass the cut
  // point and let the alignment field guide placement.
  const cutPoint = entry.timelineRange.inPoint;

  return bridge.addTransition({
    sequenceId,
    track: entry.track,
    position: cutPoint,
    transitionType: entry.transition.type,
    durationSeconds: entry.transition.durationSeconds,
  });
}

// ---------------------------------------------------------------------------
// buildTextOverlay
// ---------------------------------------------------------------------------

/**
 * Create and place a text/title clip on the timeline.
 */
export async function buildTextOverlay(
  sequenceId: string,
  overlay: TextOverlay,
  bridge: PremiereBridge,
): Promise<{ clipId: string }> {
  return bridge.addText({
    sequenceId,
    text: overlay.text,
    style: overlay.style,
    track: overlay.track,
    position: overlay.position,
    durationSeconds: overlay.durationSeconds,
  });
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Determine the highest track index used by entries of a given type.
 * Returns 0 if no entries target that track type (so at least one track
 * is always created).
 */
function computeMaxTrackIndex(
  entries: EDLEntry[],
  trackType: "video" | "audio",
): number {
  let max = 0;
  for (const entry of entries) {
    if (entry.track?.type === trackType) {
      max = Math.max(max, entry.track.trackIndex);
    }
  }
  return max;
}

/** Safely extract a message string from an unknown error value. */
function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  return String(err);
}
