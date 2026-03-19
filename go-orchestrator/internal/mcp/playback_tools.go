package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerPlaybackTools registers all playback control, program monitor,
// sequence navigation, selection, render status, and sequence metadata
// MCP tools.
func registerPlaybackTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -------------------------------------------------------------------
	// Playback Control (1-10)
	// -------------------------------------------------------------------

	// 1. premiere_play
	s.AddTool(gomcp.NewTool("premiere_play",
		gomcp.WithDescription("Play the active sequence. Speed: 1.0 = normal, 0.5 = half, 2.0 = double, -1.0 = reverse."),
		gomcp.WithNumber("speed", gomcp.Description("Playback speed multiplier (default: 1.0)")),
	), pbH(orch, logger, "play", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.Play(ctx, gomcp.ParseFloat64(req, "speed", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_pause
	s.AddTool(gomcp.NewTool("premiere_pause",
		gomcp.WithDescription("Pause playback of the active sequence."),
	), pbH(orch, logger, "pause", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.Pause(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_stop
	s.AddTool(gomcp.NewTool("premiere_stop",
		gomcp.WithDescription("Stop playback and return playhead to the start of the active sequence."),
	), pbH(orch, logger, "stop", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.Stop(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_step_forward
	s.AddTool(gomcp.NewTool("premiere_step_forward",
		gomcp.WithDescription("Step the playhead forward by a given number of frames."),
		gomcp.WithNumber("frames", gomcp.Description("Number of frames to step forward (default: 1)")),
	), pbH(orch, logger, "step_forward", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.StepForward(ctx, gomcp.ParseInt(req, "frames", 1))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_step_backward
	s.AddTool(gomcp.NewTool("premiere_step_backward",
		gomcp.WithDescription("Step the playhead backward by a given number of frames."),
		gomcp.WithNumber("frames", gomcp.Description("Number of frames to step backward (default: 1)")),
	), pbH(orch, logger, "step_backward", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.StepBackward(ctx, gomcp.ParseInt(req, "frames", 1))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_shuttle_forward
	s.AddTool(gomcp.NewTool("premiere_shuttle_forward",
		gomcp.WithDescription("Shuttle the active sequence forward at a given speed multiplier."),
		gomcp.WithNumber("speed", gomcp.Description("Shuttle speed (default: 2.0)")),
	), pbH(orch, logger, "shuttle_forward", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ShuttleForward(ctx, gomcp.ParseFloat64(req, "speed", 2.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_shuttle_backward
	s.AddTool(gomcp.NewTool("premiere_shuttle_backward",
		gomcp.WithDescription("Shuttle the active sequence backward at a given speed multiplier."),
		gomcp.WithNumber("speed", gomcp.Description("Shuttle speed (default: 2.0)")),
	), pbH(orch, logger, "shuttle_backward", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ShuttleBackward(ctx, gomcp.ParseFloat64(req, "speed", 2.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_toggle_play_pause
	s.AddTool(gomcp.NewTool("premiere_toggle_play_pause",
		gomcp.WithDescription("Toggle between play and pause on the active sequence."),
	), pbH(orch, logger, "toggle_play_pause", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.TogglePlayPause(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_play_in_to_out
	s.AddTool(gomcp.NewTool("premiere_play_in_to_out",
		gomcp.WithDescription("Play the active sequence from the in point to the out point."),
	), pbH(orch, logger, "play_in_to_out", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PlayInToOut(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_loop_playback
	s.AddTool(gomcp.NewTool("premiere_loop_playback",
		gomcp.WithDescription("Enable or disable loop playback on the active sequence."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable loop, false to disable")),
	), pbH(orch, logger, "loop_playback", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LoopPlayback(ctx, gomcp.ParseBoolean(req, "enabled", false))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Program Monitor (11-15)
	// -------------------------------------------------------------------

	// 11. premiere_get_program_monitor_zoom
	s.AddTool(gomcp.NewTool("premiere_get_program_monitor_zoom",
		gomcp.WithDescription("Get the current zoom level of the program monitor."),
	), pbH(orch, logger, "get_program_monitor_zoom", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProgramMonitorZoom(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_set_program_monitor_zoom
	s.AddTool(gomcp.NewTool("premiere_set_program_monitor_zoom",
		gomcp.WithDescription("Set the zoom level of the program monitor."),
		gomcp.WithNumber("percent", gomcp.Required(), gomcp.Description("Zoom percentage (e.g. 100 = 100%, 50 = 50%)")),
	), pbH(orch, logger, "set_program_monitor_zoom", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetProgramMonitorZoom(ctx, gomcp.ParseFloat64(req, "percent", 100))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_fit_program_monitor
	s.AddTool(gomcp.NewTool("premiere_fit_program_monitor",
		gomcp.WithDescription("Fit the active sequence to the program monitor viewport."),
	), pbH(orch, logger, "fit_program_monitor", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FitProgramMonitor(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_toggle_safe_margins
	s.AddTool(gomcp.NewTool("premiere_toggle_safe_margins",
		gomcp.WithDescription("Toggle the display of safe margins (title safe / action safe) in the program monitor."),
	), pbH(orch, logger, "toggle_safe_margins", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ToggleSafeMargins(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_get_frame_at_playhead
	s.AddTool(gomcp.NewTool("premiere_get_frame_at_playhead",
		gomcp.WithDescription("Get frame information at the current playhead position including timecode, frame number, and resolution."),
	), pbH(orch, logger, "get_frame_at_playhead", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFrameAtPlayhead(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Sequence Navigation - Extended (16-20)
	// -------------------------------------------------------------------

	// 16. premiere_go_to_timecode
	s.AddTool(gomcp.NewTool("premiere_go_to_timecode",
		gomcp.WithDescription("Navigate the playhead to a specific timecode in HH:MM:SS:FF format."),
		gomcp.WithString("timecode", gomcp.Required(), gomcp.Description("Target timecode in HH:MM:SS:FF format")),
	), pbH(orch, logger, "go_to_timecode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		tc := gomcp.ParseString(req, "timecode", "")
		if tc == "" {
			return gomcp.NewToolResultError("parameter 'timecode' is required"), nil
		}
		result, err := orch.GoToTimecode(ctx, tc)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_go_to_frame
	s.AddTool(gomcp.NewTool("premiere_go_to_frame",
		gomcp.WithDescription("Navigate the playhead to a specific frame number."),
		gomcp.WithNumber("frame_number", gomcp.Required(), gomcp.Description("Target frame number (zero-based)")),
	), pbH(orch, logger, "go_to_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GoToFrame(ctx, gomcp.ParseInt(req, "frame_number", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_get_sequence_duration
	s.AddTool(gomcp.NewTool("premiere_get_sequence_duration",
		gomcp.WithDescription("Get the total duration of the active sequence in seconds, frames, and timecode."),
	), pbH(orch, logger, "get_sequence_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceDuration(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_get_frame_count
	s.AddTool(gomcp.NewTool("premiere_get_frame_count",
		gomcp.WithDescription("Get the total frame count and duration of the active sequence."),
	), pbH(orch, logger, "get_frame_count", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFrameCount(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_get_current_timecode
	s.AddTool(gomcp.NewTool("premiere_get_current_timecode",
		gomcp.WithDescription("Get the current playhead position as a formatted timecode string (HH:MM:SS:FF)."),
	), pbH(orch, logger, "get_current_timecode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetCurrentTimecode(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Selection & Focus (21-24)
	// -------------------------------------------------------------------

	// 21. premiere_select_clips_in_range
	s.AddTool(gomcp.NewTool("premiere_select_clips_in_range",
		gomcp.WithDescription("Select all clips that overlap with a given time range on all tracks."),
		gomcp.WithNumber("start_seconds", gomcp.Required(), gomcp.Description("Start of the time range in seconds")),
		gomcp.WithNumber("end_seconds", gomcp.Required(), gomcp.Description("End of the time range in seconds")),
	), pbH(orch, logger, "select_clips_in_range", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SelectClipsInRange(ctx, gomcp.ParseFloat64(req, "start_seconds", 0), gomcp.ParseFloat64(req, "end_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_select_all_on_track
	s.AddTool(gomcp.NewTool("premiere_select_all_on_track",
		gomcp.WithDescription("Select all clips on a specific track."),
		gomcp.WithString("track_type", gomcp.Description("Track type: video or audio (default: video)"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), pbH(orch, logger, "select_all_on_track", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SelectAllOnTrack(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_invert_selection
	s.AddTool(gomcp.NewTool("premiere_invert_selection",
		gomcp.WithDescription("Invert the current clip selection: selected clips become deselected and vice versa."),
	), pbH(orch, logger, "invert_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.InvertSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_get_selection_range
	s.AddTool(gomcp.NewTool("premiere_get_selection_range",
		gomcp.WithDescription("Get the time range spanned by the currently selected clips."),
	), pbH(orch, logger, "get_selection_range", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSelectionRange(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Render Status (25-26)
	// -------------------------------------------------------------------

	// 25. premiere_get_render_status
	s.AddTool(gomcp.NewTool("premiere_get_render_status",
		gomcp.WithDescription("Get the render bar status of the active sequence including rendered vs unrendered segments."),
	), pbH(orch, logger, "get_render_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetRenderStatus(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_is_rendering
	s.AddTool(gomcp.NewTool("premiere_is_rendering",
		gomcp.WithDescription("Check if Premiere Pro is currently rendering."),
	), pbH(orch, logger, "is_rendering", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.IsRendering(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Sequence Metadata (27-30)
	// -------------------------------------------------------------------

	// 27. premiere_get_sequence_metadata
	s.AddTool(gomcp.NewTool("premiere_get_sequence_metadata",
		gomcp.WithDescription("Get the XMP metadata of the active sequence."),
	), pbH(orch, logger, "get_sequence_metadata", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceMetadata(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_set_sequence_metadata
	s.AddTool(gomcp.NewTool("premiere_set_sequence_metadata",
		gomcp.WithDescription("Set an XMP metadata key-value pair on the active sequence."),
		gomcp.WithString("key", gomcp.Required(), gomcp.Description("Metadata key")),
		gomcp.WithString("value", gomcp.Required(), gomcp.Description("Metadata value")),
	), pbH(orch, logger, "set_sequence_metadata", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		key := gomcp.ParseString(req, "key", "")
		if key == "" {
			return gomcp.NewToolResultError("parameter 'key' is required"), nil
		}
		value := gomcp.ParseString(req, "value", "")
		if value == "" {
			return gomcp.NewToolResultError("parameter 'value' is required"), nil
		}
		result, err := orch.SetSequenceMetadata(ctx, key, value)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_get_sequence_color_space
	s.AddTool(gomcp.NewTool("premiere_get_sequence_color_space",
		gomcp.WithDescription("Get the working color space of the active sequence."),
	), pbH(orch, logger, "get_sequence_color_space", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceColorSpace(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_set_sequence_color_space
	s.AddTool(gomcp.NewTool("premiere_set_sequence_color_space",
		gomcp.WithDescription("Set the working color space of the active sequence."),
		gomcp.WithString("color_space", gomcp.Required(), gomcp.Description("Color space name (e.g. 'Rec. 709', 'Rec. 2020', 'sRGB')")),
	), pbH(orch, logger, "set_sequence_color_space", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		cs := gomcp.ParseString(req, "color_space", "")
		if cs == "" {
			return gomcp.NewToolResultError("parameter 'color_space' is required"), nil
		}
		result, err := orch.SetSequenceColorSpace(ctx, cs)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// pbH is a small wrapper that logs the tool name before delegating to the handler.
func pbH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
