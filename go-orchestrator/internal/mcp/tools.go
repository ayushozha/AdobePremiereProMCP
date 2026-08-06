package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/ayushozha/AdobePremiereProMCP/go-orchestrator/internal/orchestrator"
)

// registerTools registers every MCP tool with the server. Each tool is defined
// with its JSON-Schema parameters and bound to a handler that delegates to the
// Orchestrator interface.
func registerTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	registerAppTools(s, logger)
	registerHealthTools(s, orch, logger)
	registerProjectTools(s, orch, logger)
	registerProjectMgmtTools(s, orch, logger)
	registerEditingTools(s, orch, logger)
	registerSequenceTools(s, orch, logger)
	registerClipTools(s, orch, logger)
	registerExportTools(s, orch, logger)
	registerExportTools2(s, orch, logger)
	registerAudioTools(s, orch, logger)
	registerEffectsTools(s, orch, logger)
	registerColorTools(s, orch, logger)
	registerGraphicsTools(s, orch, logger)
	registerWorkspaceTools(s, orch, logger)
	registerAdvancedEditTools(s, orch, logger)
	registerPlaybackTools(s, orch, logger)
	registerTransformTools(s, orch, logger)
	registerAITools(s, orch, logger)
	registerAITools2(s, orch, logger)
	registerMetadataTools(s, orch, logger)
	registerBatchTools(s, orch, logger)
	registerAudioAdvancedTools(s, orch, logger)
	registerTemplateTools(s, orch, logger)
	registerPreferencesTools(s, orch, logger)
	registerCollaborationTools(s, orch, logger)
	registerImmersiveTools(s, orch, logger)
	registerMotionGraphicsTools(s, orch, logger)
	registerIntegrationTools(s, orch, logger)
	registerDiagnosticsTools(s, orch, logger)
	registerUITools(s, orch, logger)
	registerMonitoringTools(s, orch, logger)
	registerCompoundTools(s, orch, logger)
	registerEncodingTools(s, orch, logger)
	registerAssemblyTools(s, orch, logger)
	registerScriptingTools(s, orch, logger)
	registerAnalyticsTools(s, orch, logger)
	registerEffectChainTools(s, orch, logger)
	registerShortcutTools(s, orch, logger)
	registerDeliveryTools(s, orch, logger)
	registerVersioningTools(s, orch, logger)
	registerCameraTools(s, orch, logger)
	registerMediaBrowserTools(s, orch, logger)
	registerPanelOpsTools(s, orch, logger)
	registerCaptureTools(s, orch, logger)
}

// ---------------------------------------------------------------------------
// Project tools
// ---------------------------------------------------------------------------

func registerProjectTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// premiere_ping — check if Premiere Pro is running.
	s.AddTool(
		gomcp.NewTool("premiere_ping",
			gomcp.WithDescription("Verify that Adobe Premiere Pro is running and the ExtendScript bridge is reachable. Returns connection status and Premiere Pro version. Call this first to confirm the MCP pipeline is working before issuing editing commands."),
		),
		makePingHandler(orch, logger),
	)

	// premiere_get_project — get current project state.
	s.AddTool(
		gomcp.NewTool("premiere_get_project",
			gomcp.WithDescription("Retrieve a snapshot of the currently open Premiere Pro project, including project name, file path, all sequences (with names, resolutions, and track counts), and top-level bins. Use this to orient yourself before performing edits."),
		),
		makeGetProjectHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Editing tools
// ---------------------------------------------------------------------------

func registerEditingTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// premiere_create_sequence
	s.AddTool(
		gomcp.NewTool("premiere_create_sequence",
			gomcp.WithDescription("Create a new empty sequence in the active Premiere Pro project with the specified resolution, frame rate, and track layout. The sequence becomes the active sequence after creation. For creating a sequence from existing clips (auto-detecting settings), use premiere_create_sequence_from_clips instead."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Display name for the new sequence (e.g. 'Main Edit', 'Social Cut 9x16')"),
			),
			gomcp.WithNumber("width",
				gomcp.Description("Frame width in pixels (default: 1920). Common values: 1920 (1080p), 3840 (4K UHD), 1080 (vertical 9:16)."),
			),
			gomcp.WithNumber("height",
				gomcp.Description("Frame height in pixels (default: 1080). Common values: 1080 (1080p/vertical), 2160 (4K UHD), 1920 (vertical 9:16)."),
			),
			gomcp.WithNumber("frame_rate",
				gomcp.Description("Frame rate in frames per second (default: 24). Common values: 23.976, 24, 25 (PAL), 29.97, 30, 50, 59.94, 60."),
			),
			gomcp.WithNumber("video_tracks",
				gomcp.Description("Number of video tracks to create (default: 3). Minimum 1. More tracks can be added later."),
			),
			gomcp.WithNumber("audio_tracks",
				gomcp.Description("Number of stereo audio tracks to create (default: 2). Minimum 1. More tracks can be added later."),
			),
		),
		makeCreateSequenceHandler(orch, logger),
	)

	// premiere_import_media
	s.AddTool(
		gomcp.NewTool("premiere_import_media",
			gomcp.WithDescription("Import a single media file (video, audio, or image) into the Premiere Pro project. The file appears in the Project panel and can then be placed on a timeline. For importing multiple files at once, use premiere_import_files. For importing an entire folder, use premiere_import_folder."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the media file to import (e.g. '/Users/me/footage/clip01.mp4'). Supports video (.mp4, .mov, .mxf), audio (.wav, .mp3, .aac), and image (.png, .jpg, .tiff) formats."),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Name of the project bin to import into (e.g. 'Footage'). If omitted, the file is imported into the project root."),
			),
		),
		makeImportMediaHandler(orch, logger),
	)

	// premiere_place_clip
	s.AddTool(
		gomcp.NewTool("premiere_place_clip",
			gomcp.WithDescription("Place (overwrite) a media clip onto the active sequence timeline at a specified position and track. Optionally set source in/out points to use only a portion of the clip, and adjust playback speed. The clip must already be imported into the project."),
			gomcp.WithString("source_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the source media file on disk, or the clip's project item ID. The file must already be imported into the project."),
			),
			gomcp.WithString("track_type",
				gomcp.Description("Type of track to place the clip on (default: 'video')."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0). Track 0 is the bottom-most track. Use premiere_get_video_tracks or premiere_get_audio_tracks to list available tracks."),
			),
			gomcp.WithNumber("position_seconds",
				gomcp.Description("Timeline position in seconds where the clip's start should be placed (default: 0). For example, 10.5 places the clip at the 10.5-second mark."),
			),
			gomcp.WithNumber("in_point_seconds",
				gomcp.Description("Source in-point in seconds. Omit to preserve the project item's current source in mark."),
			),
			gomcp.WithNumber("out_point_seconds",
				gomcp.Description("Source out-point in seconds. Omit to preserve the project item's current source out mark."),
			),
			gomcp.WithNumber("speed",
				gomcp.Description("Playback speed multiplier (default: 1.0). Use 2.0 for double speed, 0.5 for half speed. Must be positive."),
			),
		),
		makePlaceClipHandler(orch, logger),
	)

	// premiere_remove_clip
	s.AddTool(
		gomcp.NewTool("premiere_remove_clip",
			gomcp.WithDescription("Remove a clip from the timeline by its clip ID. This performs a lift edit (leaves a gap). For ripple delete (closing the gap), use premiere_remove_clip_from_track with ripple=true instead. Obtain clip IDs from premiere_get_timeline or premiere_get_all_clips."),
			gomcp.WithString("clip_id",
				gomcp.Required(),
				gomcp.Description("Unique identifier of the clip to remove. Obtain from premiere_get_timeline or premiere_get_all_clips."),
			),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("Unique identifier of the sequence containing the clip. Obtain from premiere_get_project or premiere_get_sequence_list."),
			),
		),
		makeRemoveClipHandler(orch, logger),
	)

	// premiere_add_transition
	s.AddTool(
		gomcp.NewTool("premiere_add_transition",
			gomcp.WithDescription("Add a video transition effect at a cut point on the timeline. The transition is placed at the specified position on the given track. For more control over transitions (apply to clip start/end, set duration), use premiere_add_video_transition instead. For audio transitions, use premiere_add_audio_transition."),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("Unique identifier of the target sequence. Obtain from premiere_get_project or premiere_get_sequence_list."),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index where the transition is applied (default: 0). Track 0 is the bottom-most video track."),
			),
			gomcp.WithNumber("position_seconds",
				gomcp.Description("Timeline position in seconds where the transition is placed. Should align with a cut point between two adjacent clips."),
			),
			gomcp.WithString("type",
				gomcp.Description("Transition type to apply (default: 'cross_dissolve'). cross_dissolve blends between clips, dip_to_black fades through black, morph_cut uses face-aware morphing for jump cuts."),
				gomcp.Enum("cross_dissolve", "dip_to_black", "dip_to_white", "film_dissolve", "morph_cut"),
			),
			gomcp.WithNumber("duration_seconds",
				gomcp.Description("Duration of the transition in seconds (default: 1.0). Typical range: 0.25 to 2.0 seconds."),
			),
		),
		makeAddTransitionHandler(orch, logger),
	)

	// premiere_set_audio_level
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_level",
			gomcp.WithDescription("Set the audio volume level (in decibels) for a specific clip on the timeline. Use 0 for unity gain (no change), negative values to reduce volume (e.g. -6 dB for roughly half perceived volume, -12 dB for quarter), and positive values to boost (e.g. +6 dB for double). Range: -96 dB (silence) to +15 dB (max boost). For track-level volume, use premiere_set_audio_track_volume instead."),
			gomcp.WithString("clip_id",
				gomcp.Required(),
				gomcp.Description("Unique identifier of the audio clip. Obtain from premiere_get_timeline or premiere_get_all_clips."),
			),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("Unique identifier of the sequence containing the clip. Obtain from premiere_get_project or premiere_get_sequence_list."),
			),
			gomcp.WithNumber("level_db",
				gomcp.Required(),
				gomcp.Description("Audio level in decibels. 0 = unity gain (no change), -6 = half perceived volume, -96 = silence, +6 = double perceived volume, +15 = maximum boost. Typical dialogue: -12 to -6 dB."),
			),
		),
		makeSetAudioLevelHandler(orch, logger),
	)

	// premiere_get_timeline
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline",
			gomcp.WithDescription("Retrieve the typed state of a sequence timeline: sequence ID and duration; video/audio track indices, types, mute/lock state; and each clip's canonical ID, source path, source/timeline ranges, and speed. For the active sequence, omit sequence_id. Use premiere_get_clip_info when names or effect parameters are needed."),
			gomcp.WithString("sequence_id",
				gomcp.Description("Unique identifier of the sequence to inspect. If omitted, uses the currently active sequence. Obtain IDs from premiere_get_project."),
			),
		),
		makeGetTimelineHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Export tools
// ---------------------------------------------------------------------------

func registerExportTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// premiere_export
	s.AddTool(
		gomcp.NewTool("premiere_export",
			gomcp.WithDescription("Export a sequence through Adobe Media Encoder using a named .epr preset configured on the TypeScript bridge. For an explicit preset path or direct blocking export, use premiere_export_via_ame or premiere_export_direct."),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("Unique identifier of the sequence to export. Obtain IDs from premiere_get_project."),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported output (e.g. '/Users/me/exports/final.mp4'). The file extension should match the chosen preset."),
			),
			gomcp.WithString("preset",
				gomcp.Description("Configured export preset name (default: 'h264_1080p'). Each name must map to an absolute .epr path through PREMIERE_EXPORT_PRESET_* on the bridge."),
				gomcp.Enum("h264_1080p", "h264_4k", "prores_422", "prores_4444", "dnxhd"),
			),
		),
		makeExportHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// AI-powered tools
// ---------------------------------------------------------------------------

func registerAITools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// premiere_scan_assets
	s.AddTool(
		gomcp.NewTool("premiere_scan_assets",
			gomcp.WithDescription("Scan a directory on disk for usable media assets (video, audio, images) and return metadata for each discovered file including file size, duration, codec, resolution, and type. Useful for inventorying footage before import. Does not import files -- use premiere_import_files or premiere_import_folder afterward."),
			gomcp.WithString("directory",
				gomcp.Required(),
				gomcp.Description("Absolute path to the directory to scan (e.g. '/Users/me/footage')."),
			),
			gomcp.WithBoolean("recursive",
				gomcp.Description("If true (default), scan all subdirectories recursively. Set to false to scan only the top-level directory."),
			),
			gomcp.WithArray("extensions",
				gomcp.Description("File extensions to include, e.g. ['.mp4', '.wav', '.png']. If empty or omitted, all supported media formats are included."),
				gomcp.WithStringItems(),
			),
		),
		makeScanAssetsHandler(orch, logger),
	)

	// premiere_probe_media
	s.AddTool(
		gomcp.NewTool("premiere_probe_media",
			gomcp.WithDescription("Probe one local media file through the Rust media engine and return codec, duration, resolution, audio, file-size, and fingerprint metadata without importing it into Premiere."),
			gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to the media file to inspect.")),
		),
		makeProbeMediaHandler(orch, logger),
	)

	// premiere_generate_thumbnail
	s.AddTool(
		gomcp.NewTool("premiere_generate_thumbnail",
			gomcp.WithDescription("Extract one PNG or JPEG video frame through the Rust media engine. Returns encoded image bytes as base64 JSON plus the actual dimensions."),
			gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to a video file.")),
			gomcp.WithNumber("timestamp_seconds", gomcp.Description("Frame timestamp in seconds (default: 0).")),
			gomcp.WithNumber("width", gomcp.Description("Output width in pixels (default: 320; maximum: 8192).")),
			gomcp.WithNumber("height", gomcp.Description("Output height in pixels (default: 180; maximum: 8192).")),
			gomcp.WithString("output_format", gomcp.Description("Encoded image format (default: png)."), gomcp.Enum("png", "jpg")),
		),
		makeGenerateThumbnailHandler(orch, logger),
	)

	// premiere_analyze_waveform
	s.AddTool(
		gomcp.NewTool("premiere_analyze_waveform",
			gomcp.WithDescription("Analyze a media file's audio through the Rust engine. Returns peak/RMS levels, duration, waveform samples, and measured silence regions for reviewable dialogue-cut planning."),
			gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to a media file with audio.")),
			gomcp.WithNumber("audio_track", gomcp.Description("Zero-based audio stream index (default: 0).")),
			gomcp.WithNumber("silence_threshold_db", gomcp.Description("Samples below this dBFS level are considered silent (default: -40).")),
			gomcp.WithNumber("min_silence_duration_seconds", gomcp.Description("Minimum silence-region duration (default: 0.5 seconds).")),
		),
		makeAnalyzeWaveformHandler(orch, logger),
	)

	// premiere_detect_scenes
	s.AddTool(
		gomcp.NewTool("premiere_detect_scenes",
			gomcp.WithDescription("Detect video scene-change boundaries through the Rust media engine and return their timestamps and confidence scores."),
			gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to a video file.")),
			gomcp.WithNumber("threshold", gomcp.Description("Scene sensitivity threshold from 0 to 1 (default: 0.4).")),
		),
		makeDetectScenesHandler(orch, logger),
	)

	// premiere_parse_script
	s.AddTool(
		gomcp.NewTool("premiere_parse_script",
			gomcp.WithDescription("Parse a script file or raw text into structured segments (scenes, shots, dialogue blocks) for use with premiere_auto_edit. Supports screenplay (.fountain), YouTube, and podcast transcript formats. Provide either file_path or text, not both."),
			gomcp.WithString("file_path",
				gomcp.Description("Absolute path to the script file to parse (e.g. '/Users/me/scripts/episode1.fountain'). Provide this or 'text', not both."),
			),
			gomcp.WithString("text",
				gomcp.Description("Raw script text to parse directly. Provide this or 'file_path', not both."),
			),
			gomcp.WithString("format",
				gomcp.Description("Script format hint to improve parsing accuracy. 'screenplay' = Fountain/industry format, 'youtube' = YouTube video script with sections, 'podcast' = timestamped transcript."),
				gomcp.Enum("screenplay", "youtube", "podcast"),
			),
		),
		makeParseScriptHandler(orch, logger),
	)

	// premiere_auto_edit
	s.AddTool(
		gomcp.NewTool("premiere_auto_edit",
			gomcp.WithDescription("Perform a fully automated rough cut: scan an assets directory, parse a script, match script segments to media files by name/content, and assemble a sequence with deterministic clip placement and supported transitions. Titles are a separate verified MOGRT workflow. For step-by-step control, use premiere_scan_assets, premiere_parse_script, premiere_import_files, and premiere_place_clip individually."),
			gomcp.WithString("script_path",
				gomcp.Description("Absolute path to the script file. Provide this or 'script_text', not both."),
			),
			gomcp.WithString("script_text",
				gomcp.Description("Raw script text. Provide this or 'script_path', not both."),
			),
			gomcp.WithString("assets_directory",
				gomcp.Required(),
				gomcp.Description("Absolute path to the directory containing media assets to match against the script (e.g. '/Users/me/footage'). Scanned recursively."),
			),
			gomcp.WithString("output_name",
				gomcp.Description("Name for the generated sequence (default: auto-generated from script content)."),
			),
			gomcp.WithString("resolution",
				gomcp.Description("Output resolution for the generated sequence (default: '1080p'). '4k' = 3840x2160."),
				gomcp.Enum("1080p", "4k"),
			),
			gomcp.WithString("pacing",
				gomcp.Description("Editing pace that controls cut rhythm and transition timing (default: 'moderate'). 'slow' = longer holds, 'fast' = quick cuts, 'dynamic' = varies with script energy."),
				gomcp.Enum("slow", "moderate", "fast", "dynamic"),
			),
		),
		makeAutoEditHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Handler constructors
// ---------------------------------------------------------------------------

func makePingHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_ping")

		result, err := orch.Ping(ctx)
		if err != nil {
			logger.Error("ping failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to ping Premiere Pro: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetProjectHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_project")

		result, err := orch.GetProject(ctx)
		if err != nil {
			logger.Error("get project failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get project: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeCreateSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_create_sequence")

		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}

		width := gomcp.ParseInt(req, "width", 1920)
		height := gomcp.ParseInt(req, "height", 1080)

		params := &CreateSequenceParams{
			Name:        name,
			Resolution:  Resolution{Width: uint32(width), Height: uint32(height)},
			FrameRate:   gomcp.ParseFloat64(req, "frame_rate", 24),
			VideoTracks: uint32(gomcp.ParseInt(req, "video_tracks", 3)),
			AudioTracks: uint32(gomcp.ParseInt(req, "audio_tracks", 2)),
		}

		result, err := orch.CreateSequence(ctx, params)
		if err != nil {
			logger.Error("create sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeImportMediaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_import_media")

		filePath := gomcp.ParseString(req, "file_path", "")
		if filePath == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		targetBin := gomcp.ParseString(req, "target_bin", "")

		result, err := orch.ImportMedia(ctx, filePath, targetBin)
		if err != nil {
			logger.Error("import media failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to import media: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makePlaceClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_place_clip")

		sourcePath := gomcp.ParseString(req, "source_path", "")
		if sourcePath == "" {
			return gomcp.NewToolResultError("parameter 'source_path' is required"), nil
		}

		trackType := orchestrator.TrackTypeVideo
		if gomcp.ParseString(req, "track_type", "video") == "audio" {
			trackType = orchestrator.TrackTypeAudio
		}

		positionSecs := gomcp.ParseFloat64(req, "position_seconds", 0)
		inSecs := gomcp.ParseFloat64(req, "in_point_seconds", 0)
		outSecs := gomcp.ParseFloat64(req, "out_point_seconds", 0)
		arguments, _ := objectArguments(req.Params.Arguments)
		_, hasInPoint := arguments["in_point_seconds"]
		_, hasOutPoint := arguments["out_point_seconds"]
		if positionSecs < 0 {
			return gomcp.NewToolResultError("parameter 'position_seconds' must be non-negative"), nil
		}
		if (hasInPoint && inSecs < 0) || (hasOutPoint && outSecs <= 0) {
			return gomcp.NewToolResultError("source in point must be non-negative and source out point must be positive"), nil
		}
		if hasInPoint && hasOutPoint && outSecs <= inSecs {
			return gomcp.NewToolResultError("parameter 'out_point_seconds' must be after 'in_point_seconds'"), nil
		}

		params := &PlaceClipParams{
			SourcePath: sourcePath,
			Track: TrackTarget{
				Type:       trackType,
				TrackIndex: uint32(gomcp.ParseInt(req, "track_index", 0)),
			},
			Position: secondsToTimecode(positionSecs, 24),
			Speed:    gomcp.ParseFloat64(req, "speed", 1.0),
		}

		if hasInPoint || hasOutPoint {
			params.SourceRange = &TimeRange{}
			if hasInPoint {
				params.SourceRange.InPoint = secondsToTimecode(inSecs, 24)
			}
			if hasOutPoint {
				params.SourceRange.OutPoint = secondsToTimecode(outSecs, 24)
			}
		}

		result, err := orch.PlaceClip(ctx, params)
		if err != nil {
			logger.Error("place clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to place clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRemoveClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_remove_clip")

		clipID := gomcp.ParseString(req, "clip_id", "")
		if clipID == "" {
			return gomcp.NewToolResultError("parameter 'clip_id' is required"), nil
		}
		sequenceID := gomcp.ParseString(req, "sequence_id", "")
		if sequenceID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}

		if err := orch.RemoveClip(ctx, clipID, sequenceID); err != nil {
			logger.Error("remove clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to remove clip: %v", err)), nil
		}
		return gomcp.NewToolResultText(`{"status":"ok"}`), nil
	}
}

func makeAddTransitionHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_add_transition")

		sequenceID := gomcp.ParseString(req, "sequence_id", "")
		if sequenceID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}

		positionSecs := gomcp.ParseFloat64(req, "position_seconds", 0)

		params := &TransitionParams{
			SequenceID: sequenceID,
			Track: TrackTarget{
				Type:       orchestrator.TrackTypeVideo,
				TrackIndex: uint32(gomcp.ParseInt(req, "track_index", 0)),
			},
			Position:        secondsToTimecode(positionSecs, 24),
			TransitionType:  gomcp.ParseString(req, "type", "cross_dissolve"),
			DurationSeconds: gomcp.ParseFloat64(req, "duration_seconds", 1.0),
		}

		if err := orch.AddTransition(ctx, params); err != nil {
			logger.Error("add transition failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to add transition: %v", err)), nil
		}
		return gomcp.NewToolResultText(`{"status":"ok"}`), nil
	}
}

func makeAddTextHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_add_text")

		sequenceID := gomcp.ParseString(req, "sequence_id", "")
		if sequenceID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}

		positionSecs := gomcp.ParseFloat64(req, "position_seconds", 0)

		params := &TextParams{
			SequenceID: sequenceID,
			Text:       text,
			Style: TextStyle{
				FontSize: gomcp.ParseFloat64(req, "font_size", 48),
				ColorHex: gomcp.ParseString(req, "color", "#FFFFFF"),
				Position: Position{
					X: gomcp.ParseFloat64(req, "x", 0.5),
					Y: gomcp.ParseFloat64(req, "y", 0.5),
				},
			},
			Track: TrackTarget{
				Type:       orchestrator.TrackTypeVideo,
				TrackIndex: uint32(gomcp.ParseInt(req, "track_index", 0)),
			},
			Position:        secondsToTimecode(positionSecs, 24),
			DurationSeconds: gomcp.ParseFloat64(req, "duration_seconds", 5.0),
		}

		result, err := orch.AddText(ctx, params)
		if err != nil {
			logger.Error("add text failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to add text: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetAudioLevelHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_audio_level")

		clipID := gomcp.ParseString(req, "clip_id", "")
		if clipID == "" {
			return gomcp.NewToolResultError("parameter 'clip_id' is required"), nil
		}
		sequenceID := gomcp.ParseString(req, "sequence_id", "")
		if sequenceID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		levelDB := gomcp.ParseFloat64(req, "level_db", 0)

		if err := orch.SetAudioLevel(ctx, clipID, sequenceID, levelDB); err != nil {
			logger.Error("set audio level failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set audio level: %v", err)), nil
		}
		return gomcp.NewToolResultText(`{"status":"ok"}`), nil
	}
}

func makeGetTimelineHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_timeline")

		sequenceID := gomcp.ParseString(req, "sequence_id", "")

		result, err := orch.GetTimeline(ctx, sequenceID)
		if err != nil {
			logger.Error("get timeline failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get timeline: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}

		params := &ExportParams{
			SequenceID: gomcp.ParseString(req, "sequence_id", ""),
			OutputPath: outputPath,
			Preset:     exportPresetFromString(gomcp.ParseString(req, "preset", "h264_1080p")),
		}

		result, err := orch.Export(ctx, params)
		if err != nil {
			logger.Error("export failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeScanAssetsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_scan_assets")

		directory := gomcp.ParseString(req, "directory", "")
		if directory == "" {
			return gomcp.NewToolResultError("parameter 'directory' is required"), nil
		}

		recursive := gomcp.ParseBoolean(req, "recursive", true)
		extensions, _ := extractStringSlice(req, "extensions")

		result, err := orch.ScanAssets(ctx, directory, recursive, extensions)
		if err != nil {
			logger.Error("scan assets failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to scan assets: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeProbeMediaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		filePath := gomcp.ParseString(req, "file_path", "")
		if filePath == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		result, err := orch.ProbeMedia(ctx, filePath)
		if err != nil {
			logger.Error("probe media failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to probe media: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGenerateThumbnailHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		filePath := gomcp.ParseString(req, "file_path", "")
		if filePath == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		timestamp := gomcp.ParseFloat64(req, "timestamp_seconds", 0)
		width := gomcp.ParseInt(req, "width", 320)
		height := gomcp.ParseInt(req, "height", 180)
		if timestamp < 0 {
			return gomcp.NewToolResultError("parameter 'timestamp_seconds' must be non-negative"), nil
		}
		if width < 1 || width > 8192 || height < 1 || height > 8192 {
			return gomcp.NewToolResultError("parameters 'width' and 'height' must be between 1 and 8192"), nil
		}
		result, err := orch.GenerateThumbnail(ctx, filePath, &orchestrator.ThumbnailOptions{
			TimestampSeconds: timestamp,
			Width:            uint32(width),
			Height:           uint32(height),
			OutputFormat:     gomcp.ParseString(req, "output_format", "png"),
		})
		if err != nil {
			logger.Error("generate thumbnail failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to generate thumbnail: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeAnalyzeWaveformHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		filePath := gomcp.ParseString(req, "file_path", "")
		if filePath == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		audioTrack := gomcp.ParseInt(req, "audio_track", 0)
		minDuration := gomcp.ParseFloat64(req, "min_silence_duration_seconds", 0.5)
		if audioTrack < 0 {
			return gomcp.NewToolResultError("parameter 'audio_track' must be non-negative"), nil
		}
		if minDuration <= 0 {
			return gomcp.NewToolResultError("parameter 'min_silence_duration_seconds' must be positive"), nil
		}
		result, err := orch.AnalyzeWaveform(ctx, filePath, &orchestrator.WaveformOptions{
			AudioTrack:             uint32(audioTrack),
			SilenceThresholdDB:     gomcp.ParseFloat64(req, "silence_threshold_db", -40),
			MinSilenceDurationSecs: minDuration,
		})
		if err != nil {
			logger.Error("analyze waveform failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to analyze waveform: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDetectScenesHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		filePath := gomcp.ParseString(req, "file_path", "")
		if filePath == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		threshold := gomcp.ParseFloat64(req, "threshold", 0.4)
		if threshold < 0 || threshold > 1 {
			return gomcp.NewToolResultError("parameter 'threshold' must be between 0 and 1"), nil
		}
		result, err := orch.DetectScenes(ctx, filePath, threshold)
		if err != nil {
			logger.Error("detect scenes failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to detect scenes: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeParseScriptHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_parse_script")

		filePath := gomcp.ParseString(req, "file_path", "")
		text := gomcp.ParseString(req, "text", "")
		if filePath == "" && text == "" {
			return gomcp.NewToolResultError("either 'file_path' or 'text' must be provided"), nil
		}

		format := gomcp.ParseString(req, "format", "")

		result, err := orch.ParseScript(ctx, text, filePath, format)
		if err != nil {
			logger.Error("parse script failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to parse script: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeAutoEditHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_auto_edit")

		assetsDir := gomcp.ParseString(req, "assets_directory", "")
		if assetsDir == "" {
			return gomcp.NewToolResultError("parameter 'assets_directory' is required"), nil
		}

		scriptPath := gomcp.ParseString(req, "script_path", "")
		scriptText := gomcp.ParseString(req, "script_text", "")
		if scriptPath == "" && scriptText == "" {
			return gomcp.NewToolResultError("either 'script_path' or 'script_text' must be provided"), nil
		}

		resolution := gomcp.ParseString(req, "resolution", "1080p")
		var resWidth, resHeight uint32 = 1920, 1080
		if resolution == "4k" {
			resWidth, resHeight = 3840, 2160
		}

		pacing := gomcp.ParseString(req, "pacing", "moderate")

		params := &AutoEditParams{
			ScriptPath:      scriptPath,
			ScriptText:      scriptText,
			AssetsDirectory: assetsDir,
			OutputName:      gomcp.ParseString(req, "output_name", ""),
			Recursive:       true,
			MatchStrategy:   "hybrid",
			EDLSettings: &EDLSettings{
				Resolution: Resolution{Width: resWidth, Height: resHeight},
				FrameRate:  24,
				Pacing:     pacingFromString(pacing),
			},
		}

		result, err := orch.AutoEdit(ctx, params)
		if err != nil {
			logger.Error("auto edit failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to auto-edit: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// toolResultJSON serializes v to indented JSON and wraps it in a text tool result.
func toolResultJSON(v any) (*gomcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// extractStringSlice pulls a []string from the request arguments map.
// It handles both []string and []any (the common JSON-unmarshaled form).
func extractStringSlice(req gomcp.CallToolRequest, key string) ([]string, error) {
	raw := gomcp.ParseArgument(req, key, nil)
	if raw == nil {
		return nil, fmt.Errorf("key %q not found", key)
	}

	switch v := raw.(type) {
	case []string:
		return v, nil
	case []any:
		out := make([]string, 0, len(v))
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("element %d of %q is not a string", i, key)
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("key %q is not an array (got %T)", key, raw)
	}
}

// secondsToTimecode converts a floating-point seconds value to a Timecode struct.
func secondsToTimecode(secs float64, fps float64) Timecode {
	totalSecs := uint32(secs)
	fracFrames := uint32((secs - float64(totalSecs)) * fps)
	return Timecode{
		Hours:     totalSecs / 3600,
		Minutes:   (totalSecs % 3600) / 60,
		Seconds:   totalSecs % 60,
		Frames:    fracFrames,
		FrameRate: fps,
	}
}

// exportPresetFromString converts a string preset name to the ExportPreset enum.
func exportPresetFromString(s string) ExportPreset {
	switch s {
	case "h264_1080p":
		return orchestrator.ExportPresetH264_1080P
	case "h264_4k":
		return orchestrator.ExportPresetH264_4K
	case "prores_422":
		return orchestrator.ExportPresetProRes422
	case "prores_4444":
		return orchestrator.ExportPresetProRes4444
	case "dnxhd":
		return orchestrator.ExportPresetDNxHR
	default:
		return orchestrator.ExportPresetH264_1080P
	}
}

// pacingFromString converts a string pacing name to the PacingPreset enum.
func pacingFromString(s string) orchestrator.PacingPreset {
	switch s {
	case "slow":
		return orchestrator.PacingPresetSlow
	case "moderate":
		return orchestrator.PacingPresetModerate
	case "fast":
		return orchestrator.PacingPresetFast
	case "dynamic":
		return orchestrator.PacingPresetDynamic
	default:
		return orchestrator.PacingPresetModerate
	}
}
