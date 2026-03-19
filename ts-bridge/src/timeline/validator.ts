/**
 * EDL validation — run before executing an EDL to catch errors early.
 *
 * Produces a list of hard errors (execution must stop) and soft warnings
 * (execution can continue but the result may be unexpected).
 */

import {
  timecodeToFrames,
  formatTimecode,
  type Timecode,
} from "./timecode.js";

// ---------------------------------------------------------------------------
// Types that mirror the proto definitions
// ---------------------------------------------------------------------------

export type TrackType = "video" | "audio";

export interface Resolution {
  width: number;
  height: number;
}

export interface TimeRange {
  inPoint: Timecode;
  outPoint: Timecode;
}

export interface TrackTarget {
  type: TrackType;
  trackIndex: number;
}

export interface TransitionInfo {
  type: string;
  durationSeconds: number;
  alignment: string;
}

export interface EffectInfo {
  name: string;
  parameters: Record<string, string>;
}

export interface TextOverlay {
  text: string;
  style: TextStyle;
  track: TrackTarget;
  position: Timecode;
  durationSeconds: number;
}

export interface TextStyle {
  fontFamily: string;
  fontSize: number;
  colorHex: string;
  alignment: string;
  backgroundColorHex: string;
  backgroundOpacity: number;
  position: { x: number; y: number };
}

export interface EDLEntry {
  index: number;
  sourceAssetId: string;
  sourceRange: TimeRange;
  timelineRange: TimeRange;
  track: TrackTarget;
  transition?: TransitionInfo;
  effects: EffectInfo[];
  notes: string;
}

export interface EditDecisionList {
  id: string;
  name: string;
  sequenceResolution: Resolution;
  sequenceFrameRate: number;
  entries: EDLEntry[];
  /** Optional text overlays attached to the EDL. */
  textOverlays?: TextOverlay[];
  /** Optional per-clip audio levels (clip index -> dB). */
  audioLevels?: Map<number, number>;
}

// ---------------------------------------------------------------------------
// Validation result
// ---------------------------------------------------------------------------

export type Severity = "error" | "warning";

export interface ValidationIssue {
  severity: Severity;
  /** Index of the EDL entry that caused the issue, or -1 for global issues. */
  entryIndex: number;
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationIssue[];
  warnings: ValidationIssue[];
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Validate an Edit Decision List before execution.
 *
 * Returns a structured result containing all found issues. The `valid`
 * flag is `true` only when there are zero errors (warnings are allowed).
 */
export function validateEDL(edl: EditDecisionList): ValidationResult {
  const issues: ValidationIssue[] = [];

  validateGlobalSettings(edl, issues);
  validateEntries(edl, issues);
  checkOverlappingClips(edl, issues);

  if (edl.textOverlays) {
    validateTextOverlays(edl, issues);
  }

  const errors = issues.filter((i) => i.severity === "error");
  const warnings = issues.filter((i) => i.severity === "warning");

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

// ---------------------------------------------------------------------------
// Internal validators
// ---------------------------------------------------------------------------

function issue(
  severity: Severity,
  entryIndex: number,
  message: string,
): ValidationIssue {
  return { severity, entryIndex, message };
}

/** Validate top-level EDL fields. */
function validateGlobalSettings(
  edl: EditDecisionList,
  issues: ValidationIssue[],
): void {
  if (!edl.name || edl.name.trim().length === 0) {
    issues.push(issue("error", -1, "EDL name is empty."));
  }

  if (!edl.sequenceResolution) {
    issues.push(issue("error", -1, "Sequence resolution is missing."));
  } else {
    if (edl.sequenceResolution.width <= 0 || edl.sequenceResolution.height <= 0) {
      issues.push(
        issue(
          "error",
          -1,
          `Invalid sequence resolution: ${edl.sequenceResolution.width}x${edl.sequenceResolution.height}.`,
        ),
      );
    }
  }

  if (
    edl.sequenceFrameRate <= 0 ||
    !Number.isFinite(edl.sequenceFrameRate)
  ) {
    issues.push(
      issue("error", -1, `Invalid sequence frame rate: ${edl.sequenceFrameRate}.`),
    );
  }

  if (edl.entries.length === 0) {
    issues.push(issue("warning", -1, "EDL contains no entries."));
  }
}

/** Validate each individual entry. */
function validateEntries(
  edl: EditDecisionList,
  issues: ValidationIssue[],
): void {
  const seenIndices = new Set<number>();

  for (const entry of edl.entries) {
    const idx = entry.index;

    // Duplicate indices
    if (seenIndices.has(idx)) {
      issues.push(issue("warning", idx, `Duplicate entry index ${idx}.`));
    }
    seenIndices.add(idx);

    // Source asset
    if (!entry.sourceAssetId || entry.sourceAssetId.trim().length === 0) {
      issues.push(issue("error", idx, "Missing source asset ID."));
    }

    // Track target
    validateTrackTarget(entry, issues);

    // Source range
    if (!entry.sourceRange) {
      issues.push(issue("error", idx, "Missing source range."));
    } else {
      validateTimeRange(entry.sourceRange, "source range", idx, edl.sequenceFrameRate, issues);
    }

    // Timeline range
    if (!entry.timelineRange) {
      issues.push(issue("error", idx, "Missing timeline range."));
    } else {
      validateTimeRange(entry.timelineRange, "timeline range", idx, edl.sequenceFrameRate, issues);
    }

    // Timecode frame rate consistency
    if (entry.sourceRange?.inPoint && entry.sourceRange.inPoint.frameRate !== edl.sequenceFrameRate) {
      issues.push(
        issue(
          "warning",
          idx,
          `Source in-point frame rate (${entry.sourceRange.inPoint.frameRate}) differs from sequence frame rate (${edl.sequenceFrameRate}).`,
        ),
      );
    }

    // Transition
    if (entry.transition) {
      validateTransition(entry.transition, idx, issues);
    }
  }
}

function validateTrackTarget(
  entry: EDLEntry,
  issues: ValidationIssue[],
): void {
  const idx = entry.index;

  if (!entry.track) {
    issues.push(issue("error", idx, "Missing track target."));
    return;
  }

  if (entry.track.type !== "video" && entry.track.type !== "audio") {
    issues.push(
      issue("error", idx, `Invalid track type "${entry.track.type}". Must be "video" or "audio".`),
    );
  }

  if (entry.track.trackIndex < 0 || !Number.isInteger(entry.track.trackIndex)) {
    issues.push(
      issue("error", idx, `Invalid track index ${entry.track.trackIndex}. Must be a non-negative integer.`),
    );
  }
}

function validateTimeRange(
  range: TimeRange,
  label: string,
  entryIndex: number,
  fps: number,
  issues: ValidationIssue[],
): void {
  // Check individual timecodes
  try {
    validateTimecodeValues(range.inPoint, fps);
  } catch (err) {
    issues.push(
      issue("error", entryIndex, `Invalid ${label} in-point: ${(err as Error).message}`),
    );
    return;
  }

  try {
    validateTimecodeValues(range.outPoint, fps);
  } catch (err) {
    issues.push(
      issue("error", entryIndex, `Invalid ${label} out-point: ${(err as Error).message}`),
    );
    return;
  }

  // Out must be >= in
  const inFrames = timecodeToFrames(range.inPoint);
  const outFrames = timecodeToFrames(range.outPoint);

  if (outFrames < inFrames) {
    issues.push(
      issue(
        "error",
        entryIndex,
        `${label} out-point (${formatTimecode(range.outPoint)}) is before in-point (${formatTimecode(range.inPoint)}).`,
      ),
    );
  }

  if (outFrames === inFrames) {
    issues.push(
      issue(
        "warning",
        entryIndex,
        `${label} has zero duration (in == out at ${formatTimecode(range.inPoint)}).`,
      ),
    );
  }
}

function validateTimecodeValues(tc: Timecode, _fps: number): void {
  if (tc.hours < 0 || tc.minutes < 0 || tc.seconds < 0 || tc.frames < 0) {
    throw new Error("Timecode components must be non-negative.");
  }
  if (tc.minutes > 59) {
    throw new Error(`Minutes value ${tc.minutes} exceeds 59.`);
  }
  if (tc.seconds > 59) {
    throw new Error(`Seconds value ${tc.seconds} exceeds 59.`);
  }
  const maxFrames = Math.ceil(tc.frameRate) - 1;
  if (tc.frames > maxFrames) {
    throw new Error(
      `Frames value ${tc.frames} exceeds maximum ${maxFrames} for ${tc.frameRate} fps.`,
    );
  }
}

function validateTransition(
  transition: TransitionInfo,
  entryIndex: number,
  issues: ValidationIssue[],
): void {
  const knownTypes = [
    "cross_dissolve",
    "dip_to_black",
    "dip_to_white",
    "wipe",
    "slide",
    "push",
    "fade",
  ];

  if (!transition.type || transition.type.trim().length === 0) {
    issues.push(issue("error", entryIndex, "Transition type is empty."));
  } else if (!knownTypes.includes(transition.type)) {
    issues.push(
      issue(
        "warning",
        entryIndex,
        `Unknown transition type "${transition.type}". Known types: ${knownTypes.join(", ")}.`,
      ),
    );
  }

  if (transition.durationSeconds <= 0) {
    issues.push(
      issue(
        "error",
        entryIndex,
        `Transition duration must be positive, got ${transition.durationSeconds}s.`,
      ),
    );
  }
}

/** Text overlays validation. */
function validateTextOverlays(
  edl: EditDecisionList,
  issues: ValidationIssue[],
): void {
  if (!edl.textOverlays) return;

  for (let i = 0; i < edl.textOverlays.length; i++) {
    const overlay = edl.textOverlays[i];

    if (!overlay.text || overlay.text.trim().length === 0) {
      issues.push(issue("warning", -1, `Text overlay ${i} has empty text.`));
    }

    if (overlay.durationSeconds <= 0) {
      issues.push(
        issue("error", -1, `Text overlay ${i} has non-positive duration (${overlay.durationSeconds}s).`),
      );
    }

    if (!overlay.track) {
      issues.push(issue("error", -1, `Text overlay ${i} is missing a track target.`));
    }

    if (!overlay.position) {
      issues.push(issue("error", -1, `Text overlay ${i} is missing a position timecode.`));
    }
  }
}

// ---------------------------------------------------------------------------
// Overlap detection
// ---------------------------------------------------------------------------

interface ClipSpan {
  entryIndex: number;
  startFrame: number;
  endFrame: number;
}

/**
 * Check for overlapping clips on the same track.
 *
 * Two clips overlap when they occupy the same track and their timeline
 * ranges intersect. This is usually unintentional and will cause rendering
 * artifacts in Premiere Pro.
 */
function checkOverlappingClips(
  edl: EditDecisionList,
  issues: ValidationIssue[],
): void {
  // Group entries by track key
  const trackMap = new Map<string, ClipSpan[]>();

  for (const entry of edl.entries) {
    if (!entry.track || !entry.timelineRange) continue;

    const key = `${entry.track.type}:${entry.track.trackIndex}`;
    const startFrame = timecodeToFrames(entry.timelineRange.inPoint);
    const endFrame = timecodeToFrames(entry.timelineRange.outPoint);

    if (endFrame <= startFrame) continue; // already reported as an error

    let spans = trackMap.get(key);
    if (!spans) {
      spans = [];
      trackMap.set(key, spans);
    }
    spans.push({ entryIndex: entry.index, startFrame, endFrame });
  }

  // For each track, sort by start frame and check adjacent pairs
  for (const [trackKey, spans] of trackMap) {
    spans.sort((a, b) => a.startFrame - b.startFrame);

    for (let i = 0; i < spans.length - 1; i++) {
      const current = spans[i];
      const next = spans[i + 1];

      if (current.endFrame > next.startFrame) {
        issues.push(
          issue(
            "warning",
            current.entryIndex,
            `Clip at entry ${current.entryIndex} overlaps with clip at entry ${next.entryIndex} on track ${trackKey}. ` +
              `Current ends at frame ${current.endFrame}, next starts at frame ${next.startFrame}.`,
          ),
        );
      }
    }
  }
}
