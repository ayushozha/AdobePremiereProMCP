package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/orchestrator"
)

// registerTools registers every MCP tool with the server. Each tool is defined
// with its JSON-Schema parameters and bound to a handler that delegates to the
// Orchestrator interface.
func registerTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	registerAppTools(s, logger)
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
}

// ---------------------------------------------------------------------------
// Project tools
// ---------------------------------------------------------------------------

func registerProjectTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// premiere_ping — check if Premiere Pro is running.
	s.AddTool(
		gomcp.NewTool("premiere_ping",
			gomcp.WithDescription("Check if Adobe Premiere Pro is running and reachable. Returns connection status and version information."),
		),
		makePingHandler(orch, logger),
	)

	// premiere_get_project — get current project state.
	s.AddTool(
		gomcp.NewTool("premiere_get_project",
			gomcp.WithDescription("Get the current Premiere Pro project state including project name, sequences, and bins."),
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
			gomcp.WithDescription("Create a new sequence in the active Premiere Pro project."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Name for the new sequence"),
			),
			gomcp.WithNumber("width",
				gomcp.Description("Frame width in pixels (default: 1920)"),
			),
			gomcp.WithNumber("height",
				gomcp.Description("Frame height in pixels (default: 1080)"),
			),
			gomcp.WithNumber("frame_rate",
				gomcp.Description("Frame rate in fps (default: 24)"),
			),
			gomcp.WithNumber("video_tracks",
				gomcp.Description("Number of video tracks (default: 3)"),
			),
			gomcp.WithNumber("audio_tracks",
				gomcp.Description("Number of audio tracks (default: 2)"),
			),
		),
		makeCreateSequenceHandler(orch, logger),
	)

	// premiere_import_media
	s.AddTool(
		gomcp.NewTool("premiere_import_media",
			gomcp.WithDescription("Import media files into the Premiere Pro project."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path to import"),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Name of the bin to import into (default: root)"),
			),
		),
		makeImportMediaHandler(orch, logger),
	)

	// premiere_place_clip
	s.AddTool(
		gomcp.NewTool("premiere_place_clip",
			gomcp.WithDescription("Place a clip on the timeline at a specified position."),
			gomcp.WithString("source_path",
				gomcp.Required(),
				gomcp.Description("Path to the source media file or clip ID"),
			),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithNumber("position_seconds",
				gomcp.Description("Position on the timeline in seconds (default: 0)"),
			),
			gomcp.WithNumber("in_point_seconds",
				gomcp.Description("Source in-point in seconds"),
			),
			gomcp.WithNumber("out_point_seconds",
				gomcp.Description("Source out-point in seconds"),
			),
			gomcp.WithNumber("speed",
				gomcp.Description("Playback speed multiplier (default: 1.0)"),
			),
		),
		makePlaceClipHandler(orch, logger),
	)

	// premiere_remove_clip
	s.AddTool(
		gomcp.NewTool("premiere_remove_clip",
			gomcp.WithDescription("Remove a clip from the timeline."),
			gomcp.WithString("clip_id",
				gomcp.Required(),
				gomcp.Description("ID of the clip to remove"),
			),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("ID of the sequence containing the clip"),
			),
		),
		makeRemoveClipHandler(orch, logger),
	)

	// premiere_add_transition
	s.AddTool(
		gomcp.NewTool("premiere_add_transition",
			gomcp.WithDescription("Add a transition effect between clips on the timeline."),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("ID of the target sequence"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithNumber("position_seconds",
				gomcp.Description("Position of the transition in seconds"),
			),
			gomcp.WithString("type",
				gomcp.Description("Transition type"),
				gomcp.Enum("cross_dissolve", "dip_to_black", "dip_to_white", "film_dissolve", "morph_cut"),
			),
			gomcp.WithNumber("duration_seconds",
				gomcp.Description("Duration of the transition in seconds (default: 1.0)"),
			),
		),
		makeAddTransitionHandler(orch, logger),
	)

	// premiere_add_text
	s.AddTool(
		gomcp.NewTool("premiere_add_text",
			gomcp.WithDescription("Add a text overlay to the timeline."),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("ID of the target sequence"),
			),
			gomcp.WithString("text",
				gomcp.Required(),
				gomcp.Description("Text content to display"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: topmost)"),
			),
			gomcp.WithNumber("position_seconds",
				gomcp.Description("Start position on the timeline in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration_seconds",
				gomcp.Description("Duration the text is visible in seconds (default: 5.0)"),
			),
			gomcp.WithNumber("font_size",
				gomcp.Description("Font size in points (default: 48)"),
			),
			gomcp.WithString("color",
				gomcp.Description("Text color as hex string, e.g. '#FFFFFF'"),
			),
			gomcp.WithNumber("x",
				gomcp.Description("Horizontal position 0.0-1.0 (default: 0.5, centered)"),
			),
			gomcp.WithNumber("y",
				gomcp.Description("Vertical position 0.0-1.0 (default: 0.5, centered)"),
			),
		),
		makeAddTextHandler(orch, logger),
	)

	// premiere_set_audio_level
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_level",
			gomcp.WithDescription("Set the audio level of a clip in decibels."),
			gomcp.WithString("clip_id",
				gomcp.Required(),
				gomcp.Description("ID of the audio clip"),
			),
			gomcp.WithString("sequence_id",
				gomcp.Required(),
				gomcp.Description("ID of the sequence containing the clip"),
			),
			gomcp.WithNumber("level_db",
				gomcp.Required(),
				gomcp.Description("Audio level in decibels (0 = unity, negative = quieter)"),
			),
		),
		makeSetAudioLevelHandler(orch, logger),
	)

	// premiere_get_timeline
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline",
			gomcp.WithDescription("Get the current state of a sequence's timeline, including all tracks and clips."),
			gomcp.WithString("sequence_id",
				gomcp.Description("ID of the sequence (default: active sequence)"),
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
			gomcp.WithDescription("Export a sequence to a media file using a preset."),
			gomcp.WithString("sequence_id",
				gomcp.Description("ID of the sequence to export (default: active sequence)"),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output file"),
			),
			gomcp.WithString("preset",
				gomcp.Description("Export preset name"),
				gomcp.Enum("h264_1080p", "h264_4k", "prores_422", "prores_4444", "dnxhd", "gif"),
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
			gomcp.WithDescription("Scan a directory for usable media assets (video, audio, images). Returns metadata for each discovered file."),
			gomcp.WithString("directory",
				gomcp.Required(),
				gomcp.Description("Absolute path to the directory to scan"),
			),
			gomcp.WithBoolean("recursive",
				gomcp.Description("Whether to scan subdirectories (default: true)"),
			),
			gomcp.WithArray("extensions",
				gomcp.Description("File extensions to include, e.g. ['.mp4', '.wav']. Empty means all supported."),
				gomcp.WithStringItems(),
			),
		),
		makeScanAssetsHandler(orch, logger),
	)

	// premiere_parse_script
	s.AddTool(
		gomcp.NewTool("premiere_parse_script",
			gomcp.WithDescription("Parse a script file or raw text into structured segments for editing. Supports screenplay, YouTube, and podcast formats."),
			gomcp.WithString("file_path",
				gomcp.Description("Path to the script file to parse"),
			),
			gomcp.WithString("text",
				gomcp.Description("Raw script text to parse (used when file_path is not provided)"),
			),
			gomcp.WithString("format",
				gomcp.Description("Script format hint"),
				gomcp.Enum("screenplay", "youtube", "podcast"),
			),
		),
		makeParseScriptHandler(orch, logger),
	)

	// premiere_auto_edit
	s.AddTool(
		gomcp.NewTool("premiere_auto_edit",
			gomcp.WithDescription("Perform a fully automated edit by matching a script against available assets. Creates a complete sequence with clips, transitions, and text."),
			gomcp.WithString("script_path",
				gomcp.Description("Path to the script file"),
			),
			gomcp.WithString("script_text",
				gomcp.Description("Raw script text (used when script_path is not provided)"),
			),
			gomcp.WithString("assets_directory",
				gomcp.Required(),
				gomcp.Description("Absolute path to the directory containing media assets"),
			),
			gomcp.WithString("output_name",
				gomcp.Description("Name for the generated sequence"),
			),
			gomcp.WithString("resolution",
				gomcp.Description("Output resolution"),
				gomcp.Enum("1080p", "4k"),
			),
			gomcp.WithString("pacing",
				gomcp.Description("Editing pace / rhythm"),
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

		params := &PlaceClipParams{
			SourcePath: sourcePath,
			Track: TrackTarget{
				Type:       trackType,
				TrackIndex: uint32(gomcp.ParseInt(req, "track_index", 0)),
			},
			Position: secondsToTimecode(positionSecs, 24),
			Speed:    gomcp.ParseFloat64(req, "speed", 1.0),
		}

		if inSecs > 0 || outSecs > 0 {
			params.SourceRange = &TimeRange{
				InPoint:  secondsToTimecode(inSecs, 24),
				OutPoint: secondsToTimecode(outSecs, 24),
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
