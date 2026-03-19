/**
 * Timeline operations module.
 *
 * Barrel export — re-exports the public API from every sub-module so
 * consumers can import everything from `"./timeline/index.js"`.
 */

// Timecode utilities
export {
  timecodeToSeconds,
  secondsToTimecode,
  timecodeToFrames,
  framesToTimecode,
  addTimecodes,
  subtractTimecodes,
  formatTimecode,
  parseTimecode,
  timecodesEqual,
  compareTimecodes,
  zeroTimecode,
  type Timecode,
} from "./timecode.js";

// EDL validation
export {
  validateEDL,
  type ValidationResult,
  type ValidationIssue,
  type Severity,
} from "./validator.js";

// High-level operations
export {
  assembleTimeline,
  insertClipAtPosition,
  addTransitionBetweenClips,
  buildTextOverlay,
  type PremiereBridge,
  type ProgressCallback,
  type AssembleResult,
  type AssetResolver,
  type EditDecisionList,
  type EDLEntry,
  type TrackTarget,
  type TimeRange,
  type TransitionInfo,
  type TextOverlay,
  type TextStyle,
  type Resolution,
} from "./operations.js";
