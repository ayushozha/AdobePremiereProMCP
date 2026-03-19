package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerSequenceTools registers all sequence/timeline management MCP tools.
func registerSequenceTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Sequence CRUD
	// -----------------------------------------------------------------------

	// premiere_create_sequence_from_clips
	s.AddTool(
		gomcp.NewTool("premiere_create_sequence_from_clips",
			gomcp.WithDescription("Create a new sequence from project items, auto-detecting settings from the first clip."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Name for the new sequence"),
			),
			gomcp.WithArray("clip_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based project item indices to include"),
				gomcp.WithNumberItems(),
			),
		),
		makeCreateSequenceFromClipsHandler(orch, logger),
	)

	// premiere_duplicate_sequence
	s.AddTool(
		gomcp.NewTool("premiere_duplicate_sequence",
			gomcp.WithDescription("Duplicate an existing sequence in the project."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to duplicate"),
			),
		),
		makeDuplicateSequenceHandler(orch, logger),
	)

	// premiere_delete_sequence
	s.AddTool(
		gomcp.NewTool("premiere_delete_sequence",
			gomcp.WithDescription("Delete a sequence from the project."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to delete"),
			),
		),
		makeDeleteSequenceHandler(orch, logger),
	)

	// premiere_rename_sequence
	s.AddTool(
		gomcp.NewTool("premiere_rename_sequence",
			gomcp.WithDescription("Rename an existing sequence."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to rename"),
			),
			gomcp.WithString("new_name",
				gomcp.Required(),
				gomcp.Description("New name for the sequence"),
			),
		),
		makeRenameSequenceHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Sequence Settings
	// -----------------------------------------------------------------------

	// premiere_get_sequence_settings
	s.AddTool(
		gomcp.NewTool("premiere_get_sequence_settings",
			gomcp.WithDescription("Get the full settings of a sequence including resolution, frame rate, pixel aspect ratio, fields, and audio sample rate."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence"),
			),
		),
		makeGetSequenceSettingsHandler(orch, logger),
	)

	// premiere_set_sequence_settings
	s.AddTool(
		gomcp.NewTool("premiere_set_sequence_settings",
			gomcp.WithDescription("Update sequence settings such as resolution, audio sample rate, and render quality."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to update"),
			),
			gomcp.WithNumber("width",
				gomcp.Description("New frame width in pixels"),
			),
			gomcp.WithNumber("height",
				gomcp.Description("New frame height in pixels"),
			),
			gomcp.WithNumber("audio_sample_rate",
				gomcp.Description("Audio sample rate in Hz (e.g. 48000)"),
			),
			gomcp.WithNumber("video_field_type",
				gomcp.Description("Video field type (0=progressive, 1=upper first, 2=lower first)"),
			),
			gomcp.WithBoolean("composite_linear_color",
				gomcp.Description("Enable linear color compositing"),
			),
			gomcp.WithBoolean("maximum_bit_depth",
				gomcp.Description("Enable maximum bit depth rendering"),
			),
			gomcp.WithBoolean("maximum_render_quality",
				gomcp.Description("Enable maximum render quality"),
			),
		),
		makeSetSequenceSettingsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Active Sequence
	// -----------------------------------------------------------------------

	// premiere_get_active_sequence
	s.AddTool(
		gomcp.NewTool("premiere_get_active_sequence",
			gomcp.WithDescription("Get details of the currently active sequence including resolution, track counts, and in/out points."),
		),
		makeGetActiveSequenceHandler(orch, logger),
	)

	// premiere_set_active_sequence
	s.AddTool(
		gomcp.NewTool("premiere_set_active_sequence",
			gomcp.WithDescription("Make a specific sequence the active one in the Premiere Pro timeline."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to activate"),
			),
		),
		makeSetActiveSequenceHandler(orch, logger),
	)

	// premiere_get_sequence_list
	s.AddTool(
		gomcp.NewTool("premiere_get_sequence_list",
			gomcp.WithDescription("List all sequences in the project with basic info (name, resolution, track counts, active status)."),
		),
		makeGetSequenceListHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Playhead
	// -----------------------------------------------------------------------

	// premiere_get_playhead_position
	s.AddTool(
		gomcp.NewTool("premiere_get_playhead_position",
			gomcp.WithDescription("Get the current playhead position as timecode and seconds."),
		),
		makeGetPlayheadPositionHandler(orch, logger),
	)

	// premiere_set_playhead_position
	s.AddTool(
		gomcp.NewTool("premiere_set_playhead_position",
			gomcp.WithDescription("Move the playhead to a specific position in seconds."),
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("Position in seconds to move the playhead to"),
			),
		),
		makeSetPlayheadPositionHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// In/Out Points
	// -----------------------------------------------------------------------

	// premiere_set_in_point
	s.AddTool(
		gomcp.NewTool("premiere_set_in_point",
			gomcp.WithDescription("Set the sequence in point at a specific time in seconds."),
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("Position in seconds for the in point"),
			),
		),
		makeSetInPointHandler(orch, logger),
	)

	// premiere_set_out_point
	s.AddTool(
		gomcp.NewTool("premiere_set_out_point",
			gomcp.WithDescription("Set the sequence out point at a specific time in seconds."),
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("Position in seconds for the out point"),
			),
		),
		makeSetOutPointHandler(orch, logger),
	)

	// premiere_get_in_out_points
	s.AddTool(
		gomcp.NewTool("premiere_get_in_out_points",
			gomcp.WithDescription("Get the current in and out points of the active sequence."),
		),
		makeGetInOutPointsHandler(orch, logger),
	)

	// premiere_clear_in_out_points
	s.AddTool(
		gomcp.NewTool("premiere_clear_in_out_points",
			gomcp.WithDescription("Clear the in and out points, resetting them to the sequence boundaries."),
		),
		makeClearInOutPointsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Work Area & Preview
	// -----------------------------------------------------------------------

	// premiere_set_work_area
	s.AddTool(
		gomcp.NewTool("premiere_set_work_area",
			gomcp.WithDescription("Set the work area (in/out region) for rendering and export."),
			gomcp.WithNumber("in_seconds",
				gomcp.Required(),
				gomcp.Description("Start of the work area in seconds"),
			),
			gomcp.WithNumber("out_seconds",
				gomcp.Required(),
				gomcp.Description("End of the work area in seconds"),
			),
		),
		makeSetWorkAreaHandler(orch, logger),
	)

	// premiere_render_preview
	s.AddTool(
		gomcp.NewTool("premiere_render_preview",
			gomcp.WithDescription("Render preview files for a specific time range on the active sequence."),
			gomcp.WithNumber("in_seconds",
				gomcp.Required(),
				gomcp.Description("Start of the render range in seconds"),
			),
			gomcp.WithNumber("out_seconds",
				gomcp.Required(),
				gomcp.Description("End of the render range in seconds"),
			),
		),
		makeRenderPreviewHandler(orch, logger),
	)

	// premiere_delete_preview_files
	s.AddTool(
		gomcp.NewTool("premiere_delete_preview_files",
			gomcp.WithDescription("Delete all preview/render files for the active sequence to free disk space."),
		),
		makeDeletePreviewFilesHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Nesting & Reframing
	// -----------------------------------------------------------------------

	// premiere_create_nested_sequence
	s.AddTool(
		gomcp.NewTool("premiere_create_nested_sequence",
			gomcp.WithDescription("Nest selected clips into a subsequence on the active timeline."),
			gomcp.WithArray("clip_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based clip indices on the specified track to nest"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index containing the clips (default: 0)"),
			),
		),
		makeCreateNestedSequenceHandler(orch, logger),
	)

	// premiere_auto_reframe
	s.AddTool(
		gomcp.NewTool("premiere_auto_reframe",
			gomcp.WithDescription("Auto-reframe the active sequence to a new aspect ratio using AI motion tracking."),
			gomcp.WithNumber("numerator",
				gomcp.Required(),
				gomcp.Description("Aspect ratio numerator (e.g. 9 for 9:16 vertical)"),
			),
			gomcp.WithNumber("denominator",
				gomcp.Required(),
				gomcp.Description("Aspect ratio denominator (e.g. 16 for 9:16 vertical)"),
			),
			gomcp.WithString("motion_preset",
				gomcp.Description("Motion tracking preset"),
				gomcp.Enum("default", "slow", "fast"),
			),
		),
		makeAutoReframeHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Generated Media
	// -----------------------------------------------------------------------

	// premiere_insert_black_video
	s.AddTool(
		gomcp.NewTool("premiere_insert_black_video",
			gomcp.WithDescription("Insert a black video clip onto a track at a specified position."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start position in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5.0)"),
			),
		),
		makeInsertBlackVideoHandler(orch, logger),
	)

	// premiere_insert_bars_and_tone
	s.AddTool(
		gomcp.NewTool("premiere_insert_bars_and_tone",
			gomcp.WithDescription("Insert SMPTE color bars and tone at the beginning of the sequence."),
			gomcp.WithNumber("width",
				gomcp.Description("Frame width in pixels (default: 1920)"),
			),
			gomcp.WithNumber("height",
				gomcp.Description("Frame height in pixels (default: 1080)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 10.0)"),
			),
		),
		makeInsertBarsAndToneHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Markers
	// -----------------------------------------------------------------------

	// premiere_get_sequence_markers
	s.AddTool(
		gomcp.NewTool("premiere_get_sequence_markers",
			gomcp.WithDescription("Get all markers on the active sequence including name, comment, time, type, and color."),
		),
		makeGetSequenceMarkersHandler(orch, logger),
	)

	// premiere_add_sequence_marker
	s.AddTool(
		gomcp.NewTool("premiere_add_sequence_marker",
			gomcp.WithDescription("Add a marker to the active sequence at a specific time."),
			gomcp.WithNumber("time",
				gomcp.Required(),
				gomcp.Description("Position in seconds for the marker"),
			),
			gomcp.WithString("name",
				gomcp.Description("Name/label for the marker"),
			),
			gomcp.WithString("comment",
				gomcp.Description("Comment text for the marker"),
			),
			gomcp.WithNumber("color",
				gomcp.Description("Color index (0=green, 1=red, 2=purple, 3=orange, 4=yellow, 5=white, 6=blue, 7=cyan)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration of the marker in seconds (0 = instant marker)"),
			),
		),
		makeAddSequenceMarkerHandler(orch, logger),
	)

	// premiere_delete_sequence_marker
	s.AddTool(
		gomcp.NewTool("premiere_delete_sequence_marker",
			gomcp.WithDescription("Delete a marker from the active sequence by index."),
			gomcp.WithNumber("marker_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the marker to delete"),
			),
		),
		makeDeleteSequenceMarkerHandler(orch, logger),
	)

	// premiere_navigate_to_marker
	s.AddTool(
		gomcp.NewTool("premiere_navigate_to_marker",
			gomcp.WithDescription("Move the playhead to a specific marker on the active sequence."),
			gomcp.WithNumber("marker_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the marker to navigate to"),
			),
		),
		makeNavigateToMarkerHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Handler constructors — Sequence CRUD
// ---------------------------------------------------------------------------

func makeCreateSequenceFromClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_create_sequence_from_clips")

		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}

		rawIndices, err := extractNumberSlice(req, "clip_indices")
		if err != nil || len(rawIndices) == 0 {
			return gomcp.NewToolResultError("parameter 'clip_indices' is required and must be a non-empty array of numbers"), nil
		}
		clipIndices := make([]int, len(rawIndices))
		for i, v := range rawIndices {
			clipIndices[i] = int(v)
		}

		result, err := orch.CreateSequenceFromClips(ctx, name, clipIndices)
		if err != nil {
			logger.Error("create sequence from clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create sequence from clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDuplicateSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_duplicate_sequence")

		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}

		result, err := orch.DuplicateSequence(ctx, seqIndex)
		if err != nil {
			logger.Error("duplicate sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to duplicate sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDeleteSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_delete_sequence")

		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}

		result, err := orch.DeleteSequence(ctx, seqIndex)
		if err != nil {
			logger.Error("delete sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to delete sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRenameSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_rename_sequence")

		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}
		newName := gomcp.ParseString(req, "new_name", "")
		if newName == "" {
			return gomcp.NewToolResultError("parameter 'new_name' is required"), nil
		}

		result, err := orch.RenameSequence(ctx, seqIndex, newName)
		if err != nil {
			logger.Error("rename sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to rename sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Sequence Settings
// ---------------------------------------------------------------------------

func makeGetSequenceSettingsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_sequence_settings")

		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}

		result, err := orch.GetSequenceSettings(ctx, seqIndex)
		if err != nil {
			logger.Error("get sequence settings failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get sequence settings: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetSequenceSettingsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_sequence_settings")

		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}

		params := &SetSequenceSettingsParams{
			SequenceIndex: seqIndex,
		}

		// Only set pointers for provided optional fields.
		if w := gomcp.ParseInt(req, "width", -1); w >= 0 {
			wi := int(w)
			params.Width = &wi
		}
		if h := gomcp.ParseInt(req, "height", -1); h >= 0 {
			hi := int(h)
			params.Height = &hi
		}
		if asr := gomcp.ParseFloat64(req, "audio_sample_rate", -1); asr >= 0 {
			params.AudioSampleRate = &asr
		}
		if vft := gomcp.ParseInt(req, "video_field_type", -1); vft >= 0 {
			vi := int(vft)
			params.VideoFieldType = &vi
		}
		// Boolean fields: use a sentinel-based approach.
		if raw := gomcp.ParseArgument(req, "composite_linear_color", nil); raw != nil {
			b := gomcp.ParseBoolean(req, "composite_linear_color", false)
			params.CompositeLinearColor = &b
		}
		if raw := gomcp.ParseArgument(req, "maximum_bit_depth", nil); raw != nil {
			b := gomcp.ParseBoolean(req, "maximum_bit_depth", false)
			params.MaximumBitDepth = &b
		}
		if raw := gomcp.ParseArgument(req, "maximum_render_quality", nil); raw != nil {
			b := gomcp.ParseBoolean(req, "maximum_render_quality", false)
			params.MaximumRenderQuality = &b
		}

		result, err := orch.SetSequenceSettings(ctx, params)
		if err != nil {
			logger.Error("set sequence settings failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set sequence settings: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Active Sequence
// ---------------------------------------------------------------------------

func makeGetActiveSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_active_sequence")

		result, err := orch.GetActiveSequence(ctx)
		if err != nil {
			logger.Error("get active sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get active sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetActiveSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_active_sequence")

		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}

		result, err := orch.SetActiveSequence(ctx, seqIndex)
		if err != nil {
			logger.Error("set active sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set active sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetSequenceListHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_sequence_list")

		result, err := orch.GetSequenceList(ctx)
		if err != nil {
			logger.Error("get sequence list failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get sequence list: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Playhead
// ---------------------------------------------------------------------------

func makeGetPlayheadPositionHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_playhead_position")

		result, err := orch.GetPlayheadPosition(ctx)
		if err != nil {
			logger.Error("get playhead position failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get playhead position: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetPlayheadPositionHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_playhead_position")

		seconds := gomcp.ParseFloat64(req, "seconds", -1)
		if seconds < 0 {
			return gomcp.NewToolResultError("parameter 'seconds' is required and must be >= 0"), nil
		}

		result, err := orch.SetPlayheadPosition(ctx, seconds)
		if err != nil {
			logger.Error("set playhead position failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set playhead position: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — In/Out Points
// ---------------------------------------------------------------------------

func makeSetInPointHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_in_point")

		seconds := gomcp.ParseFloat64(req, "seconds", -1)
		if seconds < 0 {
			return gomcp.NewToolResultError("parameter 'seconds' is required and must be >= 0"), nil
		}

		result, err := orch.SetInPoint(ctx, seconds)
		if err != nil {
			logger.Error("set in point failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set in point: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetOutPointHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_out_point")

		seconds := gomcp.ParseFloat64(req, "seconds", -1)
		if seconds < 0 {
			return gomcp.NewToolResultError("parameter 'seconds' is required and must be >= 0"), nil
		}

		result, err := orch.SetOutPoint(ctx, seconds)
		if err != nil {
			logger.Error("set out point failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set out point: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetInOutPointsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_in_out_points")

		result, err := orch.GetInOutPoints(ctx)
		if err != nil {
			logger.Error("get in/out points failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get in/out points: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeClearInOutPointsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_clear_in_out_points")

		result, err := orch.ClearInOutPoints(ctx)
		if err != nil {
			logger.Error("clear in/out points failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to clear in/out points: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Work Area & Preview
// ---------------------------------------------------------------------------

func makeSetWorkAreaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_work_area")

		inSecs := gomcp.ParseFloat64(req, "in_seconds", -1)
		outSecs := gomcp.ParseFloat64(req, "out_seconds", -1)
		if inSecs < 0 || outSecs < 0 {
			return gomcp.NewToolResultError("parameters 'in_seconds' and 'out_seconds' are required"), nil
		}

		result, err := orch.SetWorkArea(ctx, inSecs, outSecs)
		if err != nil {
			logger.Error("set work area failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set work area: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRenderPreviewHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_render_preview")

		inSecs := gomcp.ParseFloat64(req, "in_seconds", -1)
		outSecs := gomcp.ParseFloat64(req, "out_seconds", -1)
		if inSecs < 0 || outSecs < 0 {
			return gomcp.NewToolResultError("parameters 'in_seconds' and 'out_seconds' are required"), nil
		}

		result, err := orch.RenderPreviewFiles(ctx, inSecs, outSecs)
		if err != nil {
			logger.Error("render preview failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to render preview: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDeletePreviewFilesHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_delete_preview_files")

		result, err := orch.DeletePreviewFiles(ctx)
		if err != nil {
			logger.Error("delete preview files failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to delete preview files: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Nesting & Reframing
// ---------------------------------------------------------------------------

func makeCreateNestedSequenceHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_create_nested_sequence")

		rawIndices, err := extractNumberSlice(req, "clip_indices")
		if err != nil || len(rawIndices) == 0 {
			return gomcp.NewToolResultError("parameter 'clip_indices' is required and must be a non-empty array of numbers"), nil
		}
		clipIndices := make([]int, len(rawIndices))
		for i, v := range rawIndices {
			clipIndices[i] = int(v)
		}

		trackIndex := gomcp.ParseInt(req, "track_index", 0)

		result, err := orch.CreateNestedSequence(ctx, trackIndex, clipIndices)
		if err != nil {
			logger.Error("create nested sequence failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create nested sequence: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeAutoReframeHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_auto_reframe")

		numerator := gomcp.ParseInt(req, "numerator", -1)
		denominator := gomcp.ParseInt(req, "denominator", -1)
		if numerator <= 0 || denominator <= 0 {
			return gomcp.NewToolResultError("parameters 'numerator' and 'denominator' are required and must be > 0"), nil
		}

		motionPreset := gomcp.ParseString(req, "motion_preset", "default")

		result, err := orch.AutoReframeSequence(ctx, numerator, denominator, motionPreset)
		if err != nil {
			logger.Error("auto reframe failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to auto reframe: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Generated Media
// ---------------------------------------------------------------------------

func makeInsertBlackVideoHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_insert_black_video")

		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		startTime := gomcp.ParseFloat64(req, "start_time", 0)
		duration := gomcp.ParseFloat64(req, "duration", 5.0)

		result, err := orch.InsertBlackVideo(ctx, trackIndex, startTime, duration)
		if err != nil {
			logger.Error("insert black video failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to insert black video: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeInsertBarsAndToneHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_insert_bars_and_tone")

		width := gomcp.ParseInt(req, "width", 1920)
		height := gomcp.ParseInt(req, "height", 1080)
		duration := gomcp.ParseFloat64(req, "duration", 10.0)

		result, err := orch.InsertBarsAndTone(ctx, width, height, duration)
		if err != nil {
			logger.Error("insert bars and tone failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to insert bars and tone: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Markers
// ---------------------------------------------------------------------------

func makeGetSequenceMarkersHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_sequence_markers")

		result, err := orch.GetSequenceMarkers(ctx)
		if err != nil {
			logger.Error("get sequence markers failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get sequence markers: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeAddSequenceMarkerHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_add_sequence_marker")

		t := gomcp.ParseFloat64(req, "time", -1)
		if t < 0 {
			return gomcp.NewToolResultError("parameter 'time' is required and must be >= 0"), nil
		}

		params := &AddMarkerParams{
			Time:     t,
			Name:     gomcp.ParseString(req, "name", ""),
			Comment:  gomcp.ParseString(req, "comment", ""),
			Color:    gomcp.ParseInt(req, "color", 0),
			Duration: gomcp.ParseFloat64(req, "duration", 0),
		}

		result, err := orch.AddSequenceMarker(ctx, params)
		if err != nil {
			logger.Error("add sequence marker failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to add sequence marker: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDeleteSequenceMarkerHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_delete_sequence_marker")

		markerIndex := gomcp.ParseInt(req, "marker_index", -1)
		if markerIndex < 0 {
			return gomcp.NewToolResultError("parameter 'marker_index' is required"), nil
		}

		result, err := orch.DeleteSequenceMarker(ctx, markerIndex)
		if err != nil {
			logger.Error("delete sequence marker failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to delete sequence marker: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeNavigateToMarkerHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_navigate_to_marker")

		markerIndex := gomcp.ParseInt(req, "marker_index", -1)
		if markerIndex < 0 {
			return gomcp.NewToolResultError("parameter 'marker_index' is required"), nil
		}

		result, err := orch.NavigateToMarker(ctx, markerIndex)
		if err != nil {
			logger.Error("navigate to marker failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to navigate to marker: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// extractNumberSlice pulls a []float64 from the request arguments map.
// It handles both []float64 and []any (the common JSON-unmarshaled form).
func extractNumberSlice(req gomcp.CallToolRequest, key string) ([]float64, error) {
	raw := gomcp.ParseArgument(req, key, nil)
	if raw == nil {
		return nil, fmt.Errorf("key %q not found", key)
	}

	switch v := raw.(type) {
	case []float64:
		return v, nil
	case []any:
		out := make([]float64, 0, len(v))
		for i, item := range v {
			switch n := item.(type) {
			case float64:
				out = append(out, n)
			case int:
				out = append(out, float64(n))
			case int64:
				out = append(out, float64(n))
			default:
				return nil, fmt.Errorf("element %d of %q is not a number (got %T)", i, key, item)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("key %q is not an array (got %T)", key, raw)
	}
}
