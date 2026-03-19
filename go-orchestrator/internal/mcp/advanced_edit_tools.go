package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// advH is a small handler wrapper for advanced edit tools (mirrors efxH pattern).
func advH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerAdvancedEditTools registers all 31 advanced editing MCP tools.
func registerAdvancedEditTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Advanced Trimming (1-4)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_ripple_trim",
		gomcp.WithDescription("Ripple trim a clip edge by a time delta. Adjusts the clip boundary and shifts all subsequent clips on the track."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithBoolean("trim_end", gomcp.Description("If true, trim the outgoing edge; if false, trim the incoming edge (default: true)")),
		gomcp.WithNumber("delta_seconds", gomcp.Required(), gomcp.Description("Seconds to trim (positive = extend, negative = shorten)")),
	), advH(orch, logger, "ripple_trim", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RippleTrim(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseBoolean(req, "trim_end", true), gomcp.ParseFloat64(req, "delta_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_roll_trim",
		gomcp.WithDescription("Roll trim: move the edit point between two adjacent clips. Shortens one clip while extending the other so total duration stays the same."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the outgoing clip")),
		gomcp.WithNumber("delta_seconds", gomcp.Required(), gomcp.Description("Seconds to move the edit point (positive = later, negative = earlier)")),
	), advH(orch, logger, "roll_trim", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RollTrim(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "delta_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_slip_clip",
		gomcp.WithDescription("Slip a clip: change which portion of the source media is shown without changing the clip's position or duration on the timeline."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("delta_seconds", gomcp.Required(), gomcp.Description("Seconds to shift the source window (positive = later in source, negative = earlier)")),
	), advH(orch, logger, "slip_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SlipClip(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "delta_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_slide_clip",
		gomcp.WithDescription("Slide a clip: move a clip between its neighbours without changing its source content. Adjusts adjacent clips to fill the gap."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index (must not be first or last on track)")),
		gomcp.WithNumber("delta_seconds", gomcp.Required(), gomcp.Description("Seconds to slide the clip (positive = later, negative = earlier)")),
	), advH(orch, logger, "slide_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SlideClip(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "delta_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Paste Operations (5-6)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_paste_insert",
		gomcp.WithDescription("Paste from clipboard using insert mode, pushing subsequent clips forward on the timeline."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Insert time in seconds")),
	), advH(orch, logger, "paste_insert", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PasteInsert(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "time", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_paste_attributes",
		gomcp.WithDescription("Copy specific attributes (motion, opacity, speed, effects, audio) from one clip and apply them to another."),
		gomcp.WithString("src_track_type", gomcp.Required(), gomcp.Description("Source track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("src_track_index", gomcp.Required(), gomcp.Description("Source zero-based track index")),
		gomcp.WithNumber("src_clip_index", gomcp.Required(), gomcp.Description("Source zero-based clip index")),
		gomcp.WithString("dest_track_type", gomcp.Required(), gomcp.Description("Destination track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("dest_track_index", gomcp.Required(), gomcp.Description("Destination zero-based track index")),
		gomcp.WithNumber("dest_clip_index", gomcp.Required(), gomcp.Description("Destination zero-based clip index")),
		gomcp.WithString("attributes", gomcp.Description("Comma-separated list of attributes to copy: motion,opacity,speed,effects,audio (default: motion,opacity,effects)")),
	), advH(orch, logger, "paste_attributes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PasteAttributes(ctx,
			gomcp.ParseString(req, "src_track_type", "video"), gomcp.ParseInt(req, "src_track_index", 0), gomcp.ParseInt(req, "src_clip_index", 0),
			gomcp.ParseString(req, "dest_track_type", "video"), gomcp.ParseInt(req, "dest_track_index", 0), gomcp.ParseInt(req, "dest_clip_index", 0),
			gomcp.ParseString(req, "attributes", "motion,opacity,effects"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Match Frame (7-8)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_match_frame",
		gomcp.WithDescription("Match frame: find the source frame in the project that corresponds to the current playhead position. Opens the source clip in the Source Monitor."),
	), advH(orch, logger, "match_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.MatchFrame(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_reverse_match_frame",
		gomcp.WithDescription("Reverse match frame: find the timeline position that corresponds to the current source monitor position."),
	), advH(orch, logger, "reverse_match_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ReverseMatchFrame(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Lift & Extract (9-10)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_lift_selection",
		gomcp.WithDescription("Lift the area between the sequence in/out points, removing clips but leaving a gap. Set in/out points first."),
	), advH(orch, logger, "lift_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LiftSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_extract_selection",
		gomcp.WithDescription("Extract the area between the sequence in/out points, removing clips and closing the gap (ripple). Set in/out points first."),
	), advH(orch, logger, "extract_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExtractSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Gap Management (11-14)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_find_gaps",
		gomcp.WithDescription("Find all gaps (empty spaces) on a specific track. Returns gap positions and durations."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), advH(orch, logger, "find_gaps", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FindGaps(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_close_gap",
		gomcp.WithDescription("Close a specific gap on a track by shifting subsequent clips backward. Use premiere_find_gaps first to get the gap index."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("gap_index", gomcp.Required(), gomcp.Description("Zero-based gap index (from premiere_find_gaps)")),
	), advH(orch, logger, "close_gap", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CloseGap(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "gap_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_close_all_gaps",
		gomcp.WithDescription("Close all gaps on a track by packing clips tightly from left to right."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), advH(orch, logger, "close_all_gaps", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CloseAllGaps(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_ripple_delete_gap",
		gomcp.WithDescription("Ripple delete a specific time range on a track: removes all clips within the range and shifts later clips backward."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("start_time", gomcp.Required(), gomcp.Description("Start of the range to delete in seconds")),
		gomcp.WithNumber("end_time", gomcp.Required(), gomcp.Description("End of the range to delete in seconds")),
	), advH(orch, logger, "ripple_delete_gap", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RippleDeleteGap(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_time", 0), gomcp.ParseFloat64(req, "end_time", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clip Grouping (15-17)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_group_clips",
		gomcp.WithDescription("Group clips together. Provide a JSON array of clip references."),
		gomcp.WithString("clip_refs", gomcp.Required(), gomcp.Description(`JSON array of clip references, e.g. [{"trackType":"video","trackIndex":0,"clipIndex":0},{"trackType":"video","trackIndex":0,"clipIndex":1}]`)),
	), advH(orch, logger, "group_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		clipRefs := gomcp.ParseString(req, "clip_refs", "")
		if clipRefs == "" {
			return gomcp.NewToolResultError("parameter 'clip_refs' is required"), nil
		}
		result, err := orch.GroupClips(ctx, clipRefs)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_ungroup_clips",
		gomcp.WithDescription("Ungroup clips that are in a group containing the specified clip."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), advH(orch, logger, "ungroup_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.UngroupClips(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_grouped_clips",
		gomcp.WithDescription("Get all clips that belong to the same group as the specified clip."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), advH(orch, logger, "get_grouped_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetGroupedClips(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Snap & Alignment (18-19)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_set_snapping",
		gomcp.WithDescription("Enable or disable timeline snapping."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable snapping, false to disable")),
	), advH(orch, logger, "set_snapping", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetSnapping(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_snapping",
		gomcp.WithDescription("Get the current timeline snapping state."),
	), advH(orch, logger, "get_snapping", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSnapping(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Timeline Zoom (20-22)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_zoom_to_fit",
		gomcp.WithDescription("Zoom the timeline view to fit the entire sequence."),
	), advH(orch, logger, "zoom_to_fit", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ZoomToFitTimeline(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_zoom_to_selection",
		gomcp.WithDescription("Zoom the timeline view to fit the currently selected clips."),
	), advH(orch, logger, "zoom_to_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ZoomToSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_timeline_zoom",
		gomcp.WithDescription("Set the timeline zoom level (0.0 = fully zoomed out, 1.0 = fully zoomed in)."),
		gomcp.WithNumber("level", gomcp.Required(), gomcp.Description("Zoom level from 0.0 (zoomed out) to 1.0 (zoomed in)")),
	), advH(orch, logger, "set_timeline_zoom", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTimelineZoom(ctx, gomcp.ParseFloat64(req, "level", 0.5))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Timeline Navigation (23-28)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_go_to_next_edit_point",
		gomcp.WithDescription("Move the playhead to the next edit point (clip boundary) across all tracks."),
	), advH(orch, logger, "go_to_next_edit_point", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToNextEditPoint(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_go_to_previous_edit_point",
		gomcp.WithDescription("Move the playhead to the previous edit point (clip boundary) across all tracks."),
	), advH(orch, logger, "go_to_previous_edit_point", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToPreviousEditPoint(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_go_to_next_clip",
		gomcp.WithDescription("Move the playhead to the start of the next clip on a specific track."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), advH(orch, logger, "go_to_next_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToNextClip(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_go_to_previous_clip",
		gomcp.WithDescription("Move the playhead to the start of the previous clip on a specific track."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), advH(orch, logger, "go_to_previous_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToPreviousClip(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_go_to_sequence_start",
		gomcp.WithDescription("Move the playhead to the very beginning of the active sequence."),
	), advH(orch, logger, "go_to_sequence_start", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToSequenceStart(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_go_to_sequence_end",
		gomcp.WithDescription("Move the playhead to the end of the active sequence."),
	), advH(orch, logger, "go_to_sequence_end", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToSequenceEnd(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clip Markers (29-31)
	// -----------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_add_clip_marker",
		gomcp.WithDescription("Add a marker to a specific clip on the timeline."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Marker time relative to clip start in seconds")),
		gomcp.WithString("name", gomcp.Description("Marker name (default: 'Marker')")),
		gomcp.WithString("comment", gomcp.Description("Marker comment text")),
		gomcp.WithNumber("color", gomcp.Description("Marker color index (0-7)")),
	), advH(orch, logger, "add_clip_marker", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddClipMarker(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			gomcp.ParseFloat64(req, "time", 0),
			gomcp.ParseString(req, "name", "Marker"),
			gomcp.ParseString(req, "comment", ""),
			gomcp.ParseInt(req, "color", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_clip_markers",
		gomcp.WithDescription("Get all markers on a specific clip."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), advH(orch, logger, "get_clip_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipMarkers(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_delete_clip_marker",
		gomcp.WithDescription("Delete a specific marker from a clip by marker index."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("marker_index", gomcp.Required(), gomcp.Description("Zero-based marker index on the clip")),
	), advH(orch, logger, "delete_clip_marker", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.DeleteClipMarker(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "marker_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
