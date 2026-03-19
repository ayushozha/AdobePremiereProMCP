/**
 * Abstract bridge interface for communicating with Adobe Premiere Pro.
 *
 * Both the CEP-panel bridge and the standalone (ExtendScript CLI) bridge
 * implement this interface so that the gRPC handler layer is completely
 * decoupled from the transport mechanism.
 */

// ---------------------------------------------------------------------------
// Shared types used across bridge methods.
// These mirror the proto messages but are plain TS types so the bridge layer
// does not depend on generated protobuf code directly.
// ---------------------------------------------------------------------------

export interface Resolution {
  width: number;
  height: number;
}

export interface Timecode {
  hours: number;
  minutes: number;
  seconds: number;
  frames: number;
  frameRate: number;
}

export interface TimeRange {
  inPoint: Timecode;
  outPoint: Timecode;
}

export interface TrackTarget {
  type: "video" | "audio";
  trackIndex: number;
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

export interface EffectParams {
  name: string;
  parameters: Record<string, string>;
}

export interface EDLEntry {
  index: number;
  sourceAssetId: string;
  sourceRange: TimeRange;
  timelineRange: TimeRange;
  track: TrackTarget;
  transition?: { type: string; durationSeconds: number; alignment: string };
  effects: EffectParams[];
  notes: string;
}

export interface EditDecisionList {
  id: string;
  name: string;
  sequenceResolution: Resolution;
  sequenceFrameRate: number;
  entries: EDLEntry[];
}

// ---------------------------------------------------------------------------
// Result types returned by bridge methods
// ---------------------------------------------------------------------------

export interface SequenceInfo {
  id: string;
  name: string;
  resolution: Resolution;
  frameRate: number;
  durationSeconds: number;
  videoTrackCount: number;
  audioTrackCount: number;
}

export interface ProjectState {
  projectName: string;
  projectPath: string;
  sequences: SequenceInfo[];
  binCount: number;
  isSaved: boolean;
}

export interface TimelineClip {
  clipId: string;
  sourcePath: string;
  sourceRange: TimeRange;
  timelineRange: TimeRange;
  speed: number;
}

export interface TimelineTrack {
  index: number;
  type: "video" | "audio";
  clips: TimelineClip[];
  isMuted: boolean;
  isLocked: boolean;
}

export interface TimelineState {
  sequenceId: string;
  videoTracks: TimelineTrack[];
  audioTracks: TimelineTrack[];
  totalDurationSeconds: number;
}

export type OperationStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed";

export type ExportPreset =
  | "h264_1080p"
  | "h264_4k"
  | "prores_422"
  | "prores_4444"
  | "dnx_hr"
  | "custom";

export interface ExportResult {
  jobId: string;
  status: OperationStatus;
  outputPath: string;
}

export interface EDLExecutionResult {
  sequenceId: string;
  status: OperationStatus;
  clipsPlaced: number;
  transitionsAdded: number;
  errors: string[];
  warnings: string[];
}

export interface PingResult {
  premiereRunning: boolean;
  premiereVersion: string;
  projectOpen: boolean;
  bridgeMode: string;
}

// ---------------------------------------------------------------------------
// The bridge interface
// ---------------------------------------------------------------------------

export interface PremiereBridge {
  /** Initialize the bridge connection. */
  connect(): Promise<void>;

  /** Tear down the bridge connection. */
  disconnect(): Promise<void>;

  // -- Project ---------------------------------------------------------------

  getProjectState(): Promise<ProjectState>;

  // -- Sequence --------------------------------------------------------------

  createSequence(params: {
    name: string;
    resolution: Resolution;
    frameRate: number;
    videoTracks: number;
    audioTracks: number;
  }): Promise<{ sequenceId: string; name: string }>;

  getTimelineState(sequenceId: string): Promise<TimelineState>;

  // -- Clip Operations -------------------------------------------------------

  importMedia(params: {
    filePath: string;
    targetBin: string;
  }): Promise<{ projectItemId: string; name: string }>;

  placeClip(params: {
    sourcePath: string;
    track: TrackTarget;
    position: Timecode;
    sourceRange?: TimeRange;
    speed: number;
  }): Promise<{ clipId: string }>;

  removeClip(params: {
    clipId: string;
    sequenceId: string;
  }): Promise<void>;

  // -- Effects & Transitions -------------------------------------------------

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

  applyEffect(params: {
    clipId: string;
    sequenceId: string;
    effect: EffectParams;
  }): Promise<void>;

  // -- Audio -----------------------------------------------------------------

  setAudioLevel(params: {
    clipId: string;
    sequenceId: string;
    levelDb: number;
  }): Promise<void>;

  // -- Export -----------------------------------------------------------------

  exportSequence(params: {
    sequenceId: string;
    outputPath: string;
    preset: ExportPreset;
  }): Promise<ExportResult>;

  // -- Batch -----------------------------------------------------------------

  executeEDL(params: {
    edl: EditDecisionList;
    autoImport: boolean;
    autoCreateSequence: boolean;
  }): Promise<EDLExecutionResult>;

  // -- Health ----------------------------------------------------------------

  ping(): Promise<PingResult>;
}
