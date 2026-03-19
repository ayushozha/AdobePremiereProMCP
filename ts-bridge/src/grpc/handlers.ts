/**
 * gRPC handler implementations for PremiereBridgeService.
 *
 * Every RPC is a thin wrapper that:
 *   1. Logs the incoming call.
 *   2. Delegates to the active {@link PremiereBridge} implementation.
 *   3. Maps the bridge result back to the protobuf response shape.
 *
 * Until ts-proto code-gen is wired up, request/response types are defined
 * inline to keep the project compilable without generated code.
 */

import type { Logger } from "winston";
import type { PremiereBridge } from "../bridge/interface.js";

// ---------------------------------------------------------------------------
// Inline request / response types (mirrors proto messages)
//
// These will be replaced by ts-proto generated types once `buf generate` is
// integrated into the build.  For now they allow the handler signatures to be
// fully typed and the project to compile stand-alone.
// ---------------------------------------------------------------------------

/* eslint-disable @typescript-eslint/no-empty-interface */

export interface GetProjectStateRequest {}
export interface GetProjectStateResponse {
  projectName: string;
  projectPath: string;
  sequences: Array<{
    id: string;
    name: string;
    resolution: { width: number; height: number } | undefined;
    frameRate: number;
    durationSeconds: number;
    videoTrackCount: number;
    audioTrackCount: number;
  }>;
  binCount: number;
  isSaved: boolean;
}

export interface CreateSequenceRequest {
  name: string;
  resolution: { width: number; height: number } | undefined;
  frameRate: number;
  videoTracks: number;
  audioTracks: number;
}
export interface CreateSequenceResponse {
  sequenceId: string;
  name: string;
}

export interface GetTimelineStateRequest {
  sequenceId: string;
}
export interface GetTimelineStateResponse {
  sequenceId: string;
  videoTracks: Array<{
    index: number;
    type: number;
    clips: Array<{
      clipId: string;
      sourcePath: string;
      sourceRange: object | undefined;
      timelineRange: object | undefined;
      speed: number;
    }>;
    isMuted: boolean;
    isLocked: boolean;
  }>;
  audioTracks: Array<{
    index: number;
    type: number;
    clips: Array<{
      clipId: string;
      sourcePath: string;
      sourceRange: object | undefined;
      timelineRange: object | undefined;
      speed: number;
    }>;
    isMuted: boolean;
    isLocked: boolean;
  }>;
  totalDurationSeconds: number;
}

export interface ImportMediaRequest {
  filePath: string;
  targetBin: string;
}
export interface ImportMediaResponse {
  projectItemId: string;
  name: string;
}

export interface PlaceClipRequest {
  sourcePath: string;
  track: { type: number; trackIndex: number } | undefined;
  position: {
    hours: number;
    minutes: number;
    seconds: number;
    frames: number;
    frameRate: number;
  } | undefined;
  sourceRange: object | undefined;
  speed: number;
}
export interface PlaceClipResponse {
  clipId: string;
}

export interface RemoveClipRequest {
  clipId: string;
  sequenceId: string;
}
export interface RemoveClipResponse {}

export interface AddTransitionRequest {
  sequenceId: string;
  track: { type: number; trackIndex: number } | undefined;
  position: {
    hours: number;
    minutes: number;
    seconds: number;
    frames: number;
    frameRate: number;
  } | undefined;
  transitionType: string;
  durationSeconds: number;
}
export interface AddTransitionResponse {
  transitionId: string;
}

export interface AddTextRequest {
  sequenceId: string;
  text: string;
  style: {
    fontFamily: string;
    fontSize: number;
    colorHex: string;
    alignment: string;
    backgroundColorHex: string;
    backgroundOpacity: number;
    position: { x: number; y: number } | undefined;
  } | undefined;
  track: { type: number; trackIndex: number } | undefined;
  position: {
    hours: number;
    minutes: number;
    seconds: number;
    frames: number;
    frameRate: number;
  } | undefined;
  durationSeconds: number;
}
export interface AddTextResponse {
  clipId: string;
}

export interface ApplyEffectRequest {
  clipId: string;
  sequenceId: string;
  effect: { name: string; parameters: Record<string, string> } | undefined;
}
export interface ApplyEffectResponse {}

export interface SetAudioLevelRequest {
  clipId: string;
  sequenceId: string;
  levelDb: number;
}
export interface SetAudioLevelResponse {}

export interface ExportSequenceRequest {
  sequenceId: string;
  outputPath: string;
  preset: number;
}
export interface ExportSequenceResponse {
  jobId: string;
  status: number;
  outputPath: string;
}

export interface ExecuteEDLRequest {
  edl: object | undefined;
  autoImport: boolean;
  autoCreateSequence: boolean;
}
export interface ExecuteEDLResponse {
  sequenceId: string;
  status: number;
  clipsPlaced: number;
  transitionsAdded: number;
  errors: string[];
  warnings: string[];
}

export interface PingRequest {}
export interface PingResponse {
  premiereRunning: boolean;
  premiereVersion: string;
  projectOpen: boolean;
  bridgeMode: string;
}

// ---------------------------------------------------------------------------
// Enum maps (proto enum value <-> bridge string)
// ---------------------------------------------------------------------------

const TRACK_TYPE_MAP: Record<number, "video" | "audio"> = {
  1: "video",
  2: "audio",
};

const EXPORT_PRESET_MAP: Record<
  number,
  "h264_1080p" | "h264_4k" | "prores_422" | "prores_4444" | "dnx_hr" | "custom"
> = {
  1: "h264_1080p",
  2: "h264_4k",
  3: "prores_422",
  4: "prores_4444",
  5: "dnx_hr",
  6: "custom",
};

const OPERATION_STATUS_MAP: Record<string, number> = {
  pending: 1,
  running: 2,
  completed: 3,
  failed: 4,
};

// ---------------------------------------------------------------------------
// Handler implementations
// ---------------------------------------------------------------------------

export interface PremiereBridgeServiceHandlers {
  getProjectState(request: GetProjectStateRequest): Promise<GetProjectStateResponse>;
  createSequence(request: CreateSequenceRequest): Promise<CreateSequenceResponse>;
  getTimelineState(request: GetTimelineStateRequest): Promise<GetTimelineStateResponse>;
  importMedia(request: ImportMediaRequest): Promise<ImportMediaResponse>;
  placeClip(request: PlaceClipRequest): Promise<PlaceClipResponse>;
  removeClip(request: RemoveClipRequest): Promise<RemoveClipResponse>;
  addTransition(request: AddTransitionRequest): Promise<AddTransitionResponse>;
  addText(request: AddTextRequest): Promise<AddTextResponse>;
  applyEffect(request: ApplyEffectRequest): Promise<ApplyEffectResponse>;
  setAudioLevel(request: SetAudioLevelRequest): Promise<SetAudioLevelResponse>;
  exportSequence(request: ExportSequenceRequest): Promise<ExportSequenceResponse>;
  executeEDL(request: ExecuteEDLRequest): Promise<ExecuteEDLResponse>;
  ping(request: PingRequest): Promise<PingResponse>;
}

/**
 * Build the full set of PremiereBridgeService RPC handlers.
 */
export function createHandlers(
  bridge: PremiereBridge,
  logger: Logger,
): PremiereBridgeServiceHandlers {
  return {
    // -- Project -------------------------------------------------------------

    async getProjectState(_request) {
      logger.info("gRPC call: GetProjectState");
      const state = await bridge.getProjectState();
      return {
        projectName: state.projectName,
        projectPath: state.projectPath,
        sequences: state.sequences.map((s) => ({
          id: s.id,
          name: s.name,
          resolution: s.resolution,
          frameRate: s.frameRate,
          durationSeconds: s.durationSeconds,
          videoTrackCount: s.videoTrackCount,
          audioTrackCount: s.audioTrackCount,
        })),
        binCount: state.binCount,
        isSaved: state.isSaved,
      };
    },

    // -- Sequence ------------------------------------------------------------

    async createSequence(request) {
      logger.info("gRPC call: CreateSequence", { name: request.name });
      const result = await bridge.createSequence({
        name: request.name,
        resolution: request.resolution ?? { width: 1920, height: 1080 },
        frameRate: request.frameRate || 24,
        videoTracks: request.videoTracks || 1,
        audioTracks: request.audioTracks || 1,
      });
      return { sequenceId: result.sequenceId, name: result.name };
    },

    async getTimelineState(request) {
      logger.info("gRPC call: GetTimelineState", {
        sequenceId: request.sequenceId,
      });
      const state = await bridge.getTimelineState(request.sequenceId);
      const mapTrack = (t: (typeof state.videoTracks)[number]) => ({
        index: t.index,
        type: t.type === "video" ? 1 : 2,
        clips: t.clips.map((c) => ({
          clipId: c.clipId,
          sourcePath: c.sourcePath,
          sourceRange: c.sourceRange as object | undefined,
          timelineRange: c.timelineRange as object | undefined,
          speed: c.speed,
        })),
        isMuted: t.isMuted,
        isLocked: t.isLocked,
      });
      return {
        sequenceId: state.sequenceId,
        videoTracks: state.videoTracks.map(mapTrack),
        audioTracks: state.audioTracks.map(mapTrack),
        totalDurationSeconds: state.totalDurationSeconds,
      };
    },

    // -- Clip Operations -----------------------------------------------------

    async importMedia(request) {
      logger.info("gRPC call: ImportMedia", { filePath: request.filePath });
      return bridge.importMedia({
        filePath: request.filePath,
        targetBin: request.targetBin,
      });
    },

    async placeClip(request) {
      logger.info("gRPC call: PlaceClip", { sourcePath: request.sourcePath });
      const trackType = TRACK_TYPE_MAP[request.track?.type ?? 0] ?? "video";
      return bridge.placeClip({
        sourcePath: request.sourcePath,
        track: {
          type: trackType,
          trackIndex: request.track?.trackIndex ?? 0,
        },
        position: request.position ?? {
          hours: 0,
          minutes: 0,
          seconds: 0,
          frames: 0,
          frameRate: 24,
        },
        speed: request.speed || 1.0,
      });
    },

    async removeClip(request) {
      logger.info("gRPC call: RemoveClip", { clipId: request.clipId });
      await bridge.removeClip({
        clipId: request.clipId,
        sequenceId: request.sequenceId,
      });
      return {};
    },

    // -- Effects & Transitions -----------------------------------------------

    async addTransition(request) {
      logger.info("gRPC call: AddTransition", {
        sequenceId: request.sequenceId,
        type: request.transitionType,
      });
      const trackType = TRACK_TYPE_MAP[request.track?.type ?? 0] ?? "video";
      return bridge.addTransition({
        sequenceId: request.sequenceId,
        track: {
          type: trackType,
          trackIndex: request.track?.trackIndex ?? 0,
        },
        position: request.position ?? {
          hours: 0,
          minutes: 0,
          seconds: 0,
          frames: 0,
          frameRate: 24,
        },
        transitionType: request.transitionType,
        durationSeconds: request.durationSeconds,
      });
    },

    async addText(request) {
      logger.info("gRPC call: AddText", {
        sequenceId: request.sequenceId,
        text: request.text,
      });
      const trackType = TRACK_TYPE_MAP[request.track?.type ?? 0] ?? "video";
      return bridge.addText({
        sequenceId: request.sequenceId,
        text: request.text,
        style: {
          fontFamily: request.style?.fontFamily ?? "Arial",
          fontSize: request.style?.fontSize ?? 48,
          colorHex: request.style?.colorHex ?? "#FFFFFF",
          alignment: request.style?.alignment ?? "center",
          backgroundColorHex: request.style?.backgroundColorHex ?? "#000000",
          backgroundOpacity: request.style?.backgroundOpacity ?? 0,
          position: request.style?.position ?? { x: 0.5, y: 0.5 },
        },
        track: {
          type: trackType,
          trackIndex: request.track?.trackIndex ?? 0,
        },
        position: request.position ?? {
          hours: 0,
          minutes: 0,
          seconds: 0,
          frames: 0,
          frameRate: 24,
        },
        durationSeconds: request.durationSeconds,
      });
    },

    async applyEffect(request) {
      logger.info("gRPC call: ApplyEffect", {
        clipId: request.clipId,
        effect: request.effect?.name,
      });
      await bridge.applyEffect({
        clipId: request.clipId,
        sequenceId: request.sequenceId,
        effect: {
          name: request.effect?.name ?? "",
          parameters: request.effect?.parameters ?? {},
        },
      });
      return {};
    },

    // -- Audio ---------------------------------------------------------------

    async setAudioLevel(request) {
      logger.info("gRPC call: SetAudioLevel", {
        clipId: request.clipId,
        levelDb: request.levelDb,
      });
      await bridge.setAudioLevel({
        clipId: request.clipId,
        sequenceId: request.sequenceId,
        levelDb: request.levelDb,
      });
      return {};
    },

    // -- Export ---------------------------------------------------------------

    async exportSequence(request) {
      logger.info("gRPC call: ExportSequence", {
        sequenceId: request.sequenceId,
        outputPath: request.outputPath,
      });
      const preset = EXPORT_PRESET_MAP[request.preset] ?? "h264_1080p";
      const result = await bridge.exportSequence({
        sequenceId: request.sequenceId,
        outputPath: request.outputPath,
        preset,
      });
      return {
        jobId: result.jobId,
        status: OPERATION_STATUS_MAP[result.status] ?? 0,
        outputPath: result.outputPath,
      };
    },

    // -- Batch ---------------------------------------------------------------

    async executeEDL(request) {
      logger.info("gRPC call: ExecuteEDL");
      // For now, pass through the raw EDL object.  Once ts-proto types are
      // generated, this will be properly mapped.
      const result = await bridge.executeEDL({
        edl: (request.edl ?? {}) as any,
        autoImport: request.autoImport,
        autoCreateSequence: request.autoCreateSequence,
      });
      return {
        sequenceId: result.sequenceId,
        status: OPERATION_STATUS_MAP[result.status] ?? 0,
        clipsPlaced: result.clipsPlaced,
        transitionsAdded: result.transitionsAdded,
        errors: result.errors,
        warnings: result.warnings,
      };
    },

    // -- Health --------------------------------------------------------------

    async ping(_request) {
      logger.info("gRPC call: Ping");
      return bridge.ping();
    },
  };
}
