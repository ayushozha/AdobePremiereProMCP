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
			gomcp.WithDescription("Create a new sequence by dragging project items onto a new timeline, automatically matching sequence settings (resolution, frame rate, codec) to the first clip. This is the recommended way to create a sequence when you have footage -- it avoids mismatched settings. The new sequence becomes the active sequence."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Display name for the new sequence (e.g. 'Interview Edit', 'Rough Cut')."),
			),
			gomcp.WithArray("clip_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based project item indices to include, in the order they should appear on the timeline. Use premiere_get_project_items to find indices. Example: [0, 2, 5] places items 0, 2, and 5 sequentially."),
				gomcp.WithNumberItems(),
			),
		),
		makeCreateSequenceFromClipsHandler(orch, logger),
	)

	// premiere_duplicate_sequence
	s.AddTool(
		gomcp.NewTool("premiere_duplicate_sequence",
			gomcp.WithDescription("Create an exact copy of an existing sequence, including all clips, effects, transitions, and settings. The duplicate appears in the project panel with ' Copy' appended to the name. Useful for creating alternate versions of an edit without modifying the original."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to duplicate. Use premiere_get_sequence_list to find sequence indices."),
			),
		),
		makeDuplicateSequenceHandler(orch, logger),
	)

	// premiere_delete_sequence
	s.AddTool(
		gomcp.NewTool("premiere_delete_sequence",
			gomcp.WithDescription("Permanently delete a sequence from the project. WARNING: This is destructive and cannot be undone via the MCP. The sequence and all its timeline edits are removed. Consider using premiere_duplicate_sequence to create a backup first."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to delete. Use premiere_get_sequence_list to find sequence indices. Cannot delete the last remaining sequence."),
			),
		),
		makeDeleteSequenceHandler(orch, logger),
	)

	// premiere_rename_sequence
	s.AddTool(
		gomcp.NewTool("premiere_rename_sequence",
			gomcp.WithDescription("Change the display name of an existing sequence. The sequence keeps all its contents, settings, and timeline position."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to rename. Use premiere_get_sequence_list to find sequence indices."),
			),
			gomcp.WithString("new_name",
				gomcp.Required(),
				gomcp.Description("New display name for the sequence (e.g. 'Final Cut v3', 'Social Media 9x16')."),
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
			gomcp.WithDescription("Retrieve the full technical settings of a sequence, including resolution (width/height), frame rate, pixel aspect ratio, field order, audio sample rate, and render quality flags. Use this to verify settings before export or when troubleshooting playback issues."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to inspect. Use premiere_get_sequence_list to find sequence indices."),
			),
		),
		makeGetSequenceSettingsHandler(orch, logger),
	)

	// premiere_set_sequence_settings
	s.AddTool(
		gomcp.NewTool("premiere_set_sequence_settings",
			gomcp.WithDescription("Update one or more technical settings on a sequence. Only provide the parameters you want to change -- omitted parameters remain unchanged. WARNING: Changing resolution mid-edit may cause mismatched clip framing. Use premiere_get_sequence_settings first to see current values."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to update. Use premiere_get_sequence_list to find sequence indices."),
			),
			gomcp.WithNumber("width",
				gomcp.Description("New frame width in pixels. Common values: 1920 (1080p), 3840 (4K UHD), 1080 (vertical 9:16). Must be a positive even number."),
			),
			gomcp.WithNumber("height",
				gomcp.Description("New frame height in pixels. Common values: 1080 (1080p/vertical), 2160 (4K UHD), 1920 (vertical 9:16). Must be a positive even number."),
			),
			gomcp.WithNumber("audio_sample_rate",
				gomcp.Description("Audio sample rate in Hz. Standard values: 44100 (CD quality), 48000 (video standard, most common), 96000 (high-resolution audio)."),
			),
			gomcp.WithNumber("video_field_type",
				gomcp.Description("Video field order: 0 = progressive (modern default, no interlacing), 1 = upper field first, 2 = lower field first. Use 0 unless working with interlaced broadcast footage."),
			),
			gomcp.WithBoolean("composite_linear_color",
				gomcp.Description("Enable linear color compositing for more accurate color blending between layers. Recommended for VFX work. May slightly increase render time."),
			),
			gomcp.WithBoolean("maximum_bit_depth",
				gomcp.Description("Enable maximum bit depth (32-bit float) for effects processing. Reduces banding in gradients and color grading. Recommended for high-end color work."),
			),
			gomcp.WithBoolean("maximum_render_quality",
				gomcp.Description("Enable maximum render quality for scaling operations. Produces sharper results when resizing clips but significantly increases render time. Recommended only for final output."),
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
			gomcp.WithDescription("Get detailed information about the currently active (open in Timeline panel) sequence, including name, resolution, frame rate, video/audio track counts, in/out points, and duration. Most editing tools operate on the active sequence by default, so call this to understand what you are editing."),
		),
		makeGetActiveSequenceHandler(orch, logger),
	)

	// premiere_set_active_sequence
	s.AddTool(
		gomcp.NewTool("premiere_set_active_sequence",
			gomcp.WithDescription("Switch the active sequence in the Premiere Pro Timeline panel. All subsequent editing operations (clip placement, effects, markers) will apply to this sequence. The sequence opens in the Timeline panel."),
			gomcp.WithNumber("sequence_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the sequence to activate. Use premiere_get_sequence_list to see all available sequences and their indices."),
			),
		),
		makeSetActiveSequenceHandler(orch, logger),
	)

	// premiere_get_sequence_list
	s.AddTool(
		gomcp.NewTool("premiere_get_sequence_list",
			gomcp.WithDescription("List all sequences in the project with summary information for each: name, index, resolution, frame rate, video/audio track counts, and whether it is the currently active sequence. Use this to find sequence indices for other sequence operations."),
		),
		makeGetSequenceListHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Playhead
	// -----------------------------------------------------------------------

	// premiere_get_playhead_position
	s.AddTool(
		gomcp.NewTool("premiere_get_playhead_position",
			gomcp.WithDescription("Get the current playhead (CTI - Current Time Indicator) position on the active sequence, returned as both timecode (HH:MM:SS:FF) and seconds. Useful for determining where to insert clips or markers relative to the current viewing position."),
		),
		makeGetPlayheadPositionHandler(orch, logger),
	)

	// premiere_set_playhead_position
	s.AddTool(
		gomcp.NewTool("premiere_set_playhead_position",
			gomcp.WithDescription("Move the playhead (CTI) to a specific position on the active sequence. This updates the Program Monitor to show the frame at the specified time. Also useful as a prerequisite for premiere_export_frame (which captures at the playhead position)."),
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("Position in seconds to move the playhead to. Must be >= 0. For example, 30.5 moves to the 30.5-second mark. Use premiere_get_active_sequence to check the sequence duration."),
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
			gomcp.WithDescription("Set the sequence in point (mark in) at a specific time. The in point defines the start of the region for export, rendering, or clip insertion. Must be less than the current out point."),
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("Position in seconds for the in point. Must be >= 0 and less than the out point. For example, 5.0 marks 5 seconds into the sequence."),
			),
		),
		makeSetInPointHandler(orch, logger),
	)

	// premiere_set_out_point
	s.AddTool(
		gomcp.NewTool("premiere_set_out_point",
			gomcp.WithDescription("Set the sequence out point (mark out) at a specific time. The out point defines the end of the region for export, rendering, or clip insertion. Must be greater than the current in point."),
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("Position in seconds for the out point. Must be >= 0 and greater than the in point. For example, 60.0 marks 1 minute into the sequence."),
			),
		),
		makeSetOutPointHandler(orch, logger),
	)

	// premiere_get_in_out_points
	s.AddTool(
		gomcp.NewTool("premiere_get_in_out_points",
			gomcp.WithDescription("Get the current in and out points of the active sequence, returned as both timecode and seconds. Also indicates whether they have been explicitly set or are at the sequence boundaries (cleared state)."),
		),
		makeGetInOutPointsHandler(orch, logger),
	)

	// premiere_clear_in_out_points
	s.AddTool(
		gomcp.NewTool("premiere_clear_in_out_points",
			gomcp.WithDescription("Clear (reset) both the in and out points on the active sequence, restoring them to the full sequence boundaries. Equivalent to pressing Alt+X / Option+X in Premiere Pro."),
		),
		makeClearInOutPointsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Work Area & Preview
	// -----------------------------------------------------------------------

	// premiere_set_work_area
	s.AddTool(
		gomcp.NewTool("premiere_set_work_area",
			gomcp.WithDescription("Set the work area bar (gray bar above the timeline ruler) for rendering and export scope. The work area defines which portion of the sequence is rendered when using 'Render Work Area' or exported with work_area_type=2. The out must be greater than in."),
			gomcp.WithNumber("in_seconds",
				gomcp.Required(),
				gomcp.Description("Start of the work area in seconds. Must be >= 0 and less than out_seconds."),
			),
			gomcp.WithNumber("out_seconds",
				gomcp.Required(),
				gomcp.Description("End of the work area in seconds. Must be greater than in_seconds."),
			),
		),
		makeSetWorkAreaHandler(orch, logger),
	)

	// premiere_render_preview
	s.AddTool(
		gomcp.NewTool("premiere_render_preview",
			gomcp.WithDescription("Render preview files for a specific time range on the active sequence. Preview files enable real-time playback of effects-heavy sections. The render bar above the timeline turns green for rendered sections. For clearing rendered previews, use premiere_delete_preview_files."),
			gomcp.WithNumber("in_seconds",
				gomcp.Required(),
				gomcp.Description("Start of the render range in seconds. Must be >= 0."),
			),
			gomcp.WithNumber("out_seconds",
				gomcp.Required(),
				gomcp.Description("End of the render range in seconds. Must be greater than in_seconds."),
			),
		),
		makeRenderPreviewHandler(orch, logger),
	)

	// premiere_delete_preview_files
	s.AddTool(
		gomcp.NewTool("premiere_delete_preview_files",
			gomcp.WithDescription("Delete all rendered preview files for the active sequence to free disk space. After deletion, the render bar reverts to yellow/red (unrendered) and effects-heavy sections may not play in real time until re-rendered. Useful when disk space is low."),
		),
		makeDeletePreviewFilesHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Nesting & Reframing
	// -----------------------------------------------------------------------

	// premiere_create_nested_sequence
	s.AddTool(
		gomcp.NewTool("premiere_create_nested_sequence",
			gomcp.WithDescription("Nest (collapse) multiple clips on a single video track into a new subsequence. The selected clips are replaced on the timeline by a single nested sequence clip. The original clips are preserved inside the nest. Nesting is useful for applying a single effect across multiple clips, simplifying complex timelines, or creating reusable segments."),
			gomcp.WithArray("clip_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based clip indices on the specified track to nest together (e.g. [0, 1, 2] nests the first three clips). Use premiere_get_clips_on_track to find clip indices. Clips must be on the same track."),
				gomcp.WithNumberItems(),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index containing the clips to nest (default: 0). Track 0 is the bottom-most video track."),
			),
		),
		makeCreateNestedSequenceHandler(orch, logger),
	)

	// premiere_auto_reframe
	s.AddTool(
		gomcp.NewTool("premiere_auto_reframe",
			gomcp.WithDescription("Auto-reframe the active sequence to a new aspect ratio using Premiere Pro's AI-powered motion tracking (Adobe Sensei). Creates a new sequence with the target aspect ratio and intelligently repositions/crops clips to keep the main subject in frame. Common uses: converting 16:9 horizontal to 9:16 vertical (TikTok/Reels/Shorts), 1:1 square (Instagram), or 4:5 (Facebook)."),
			gomcp.WithNumber("numerator",
				gomcp.Required(),
				gomcp.Description("Aspect ratio numerator. Examples: 9 (for 9:16 vertical), 1 (for 1:1 square), 4 (for 4:5), 16 (for 16:9 landscape)."),
			),
			gomcp.WithNumber("denominator",
				gomcp.Required(),
				gomcp.Description("Aspect ratio denominator. Examples: 16 (for 9:16 vertical), 1 (for 1:1 square), 5 (for 4:5), 9 (for 16:9 landscape)."),
			),
			gomcp.WithString("motion_preset",
				gomcp.Description("Motion tracking speed/sensitivity preset (default: 'default'). 'slow' = minimal camera movement reframing, best for interviews. 'default' = balanced. 'fast' = aggressive reframing, best for action footage."),
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
			gomcp.WithDescription("Insert a synthetic black video clip (solid black frame) onto a track at a specified position. Useful for creating pauses, fade-to-black segments, or placeholder gaps in the timeline. The black clip is a generated media item, not a file on disk."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index to place the black clip on (default: 0). Track 0 is the bottom-most video track."),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start position on the timeline in seconds (default: 0)."),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration of the black clip in seconds (default: 5.0). Typical range: 1.0 to 30.0."),
			),
		),
		makeInsertBlackVideoHandler(orch, logger),
	)

	// premiere_insert_bars_and_tone
	s.AddTool(
		gomcp.NewTool("premiere_insert_bars_and_tone",
			gomcp.WithDescription("Insert SMPTE color bars and 1kHz reference tone at the beginning of the active sequence. Standard practice for broadcast delivery -- provides a visual and audio calibration reference. The bars and tone clip is placed at time 0 on video track 0 and audio track 0."),
			gomcp.WithNumber("width",
				gomcp.Description("Frame width in pixels (default: 1920). Should match the sequence resolution."),
			),
			gomcp.WithNumber("height",
				gomcp.Description("Frame height in pixels (default: 1080). Should match the sequence resolution."),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration of the bars and tone in seconds (default: 10.0). Broadcast standard is typically 10 to 30 seconds."),
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
			gomcp.WithDescription("Get all markers on the active sequence, including each marker's name, comment, time (timecode and seconds), color, duration, and index. Markers can represent chapter points, notes, edit decisions, or synchronization references."),
		),
		makeGetSequenceMarkersHandler(orch, logger),
	)

	// premiere_add_sequence_marker
	s.AddTool(
		gomcp.NewTool("premiere_add_sequence_marker",
			gomcp.WithDescription("Add a marker to the active sequence timeline at a specific time. Markers are visual flags on the timeline ruler used for notes, chapter points, or marking review feedback. A marker with duration > 0 spans a range (useful for highlighting sections that need attention)."),
			gomcp.WithNumber("time",
				gomcp.Required(),
				gomcp.Description("Position in seconds for the marker on the sequence timeline. Must be >= 0."),
			),
			gomcp.WithString("name",
				gomcp.Description("Short name/label for the marker (e.g. 'Chapter 1', 'Fix Audio', 'Approved'). Appears on the timeline ruler."),
			),
			gomcp.WithString("comment",
				gomcp.Description("Longer comment or note for the marker (e.g. 'Replace this shot with B-roll of city'). Visible when hovering or in the Markers panel."),
			),
			gomcp.WithNumber("color",
				gomcp.Description("Marker color index for visual categorization: 0=Green (default), 1=Red, 2=Purple, 3=Orange, 4=Yellow, 5=White, 6=Blue, 7=Cyan. Use consistent colors (e.g., red for issues, green for approved)."),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration of the marker in seconds. 0 (default) = instant point marker. Values > 0 create a range marker that spans from 'time' to 'time + duration'."),
			),
		),
		makeAddSequenceMarkerHandler(orch, logger),
	)

	// premiere_delete_sequence_marker
	s.AddTool(
		gomcp.NewTool("premiere_delete_sequence_marker",
			gomcp.WithDescription("Delete a marker from the active sequence by its index. Use premiere_get_sequence_markers first to find the marker index. Note: deleting a marker shifts the indices of all subsequent markers."),
			gomcp.WithNumber("marker_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the marker to delete. Use premiere_get_sequence_markers to list all markers and their indices."),
			),
		),
		makeDeleteSequenceMarkerHandler(orch, logger),
	)

	// premiere_navigate_to_marker
	s.AddTool(
		gomcp.NewTool("premiere_navigate_to_marker",
			gomcp.WithDescription("Move the playhead to the position of a specific marker on the active sequence. Useful for quickly jumping to marked points of interest. Use premiere_get_sequence_markers to find marker indices."),
			gomcp.WithNumber("marker_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the marker to navigate to. Use premiere_get_sequence_markers to list all markers and their indices."),
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
