package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// monH is a small handler wrapper for monitoring tools.
func monH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerMonitoringTools registers all 30 event-driven monitoring,
// notification, and real-time state tracking MCP tools.
func registerMonitoringTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -------------------------------------------------------------------
	// Event Monitoring (1-5)
	// -------------------------------------------------------------------

	// 1. premiere_register_event_listener
	s.AddTool(gomcp.NewTool("premiere_register_event_listener",
		gomcp.WithDescription("Register for a Premiere Pro event (e.g. onActiveSequenceChanged, onItemsAddedToProject, onProjectChanged)."),
		gomcp.WithString("event_name", gomcp.Required(), gomcp.Description("Name of the Premiere Pro event to listen for")),
	), monH(orch, logger, "register_event_listener", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		eventName := gomcp.ParseString(req, "event_name", "")
		if eventName == "" {
			return gomcp.NewToolResultError("parameter 'event_name' is required"), nil
		}
		result, err := orch.RegisterEventListener(ctx, eventName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_unregister_event_listener
	s.AddTool(gomcp.NewTool("premiere_unregister_event_listener",
		gomcp.WithDescription("Unregister a previously registered Premiere Pro event listener."),
		gomcp.WithString("event_name", gomcp.Required(), gomcp.Description("Name of the event to unregister")),
	), monH(orch, logger, "unregister_event_listener", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		eventName := gomcp.ParseString(req, "event_name", "")
		if eventName == "" {
			return gomcp.NewToolResultError("parameter 'event_name' is required"), nil
		}
		result, err := orch.UnregisterEventListener(ctx, eventName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_registered_events
	s.AddTool(gomcp.NewTool("premiere_get_registered_events",
		gomcp.WithDescription("List all currently active event registrations."),
	), monH(orch, logger, "get_registered_events", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetRegisteredEvents(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_get_event_history
	s.AddTool(gomcp.NewTool("premiere_get_event_history",
		gomcp.WithDescription("Get the last N events that have fired in Premiere Pro."),
		gomcp.WithNumber("count", gomcp.Description("Number of recent events to return (default: 50)")),
	), monH(orch, logger, "get_event_history", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		count := gomcp.ParseInt(req, "count", 50)
		result, err := orch.GetEventHistory(ctx, count)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_clear_event_history
	s.AddTool(gomcp.NewTool("premiere_clear_event_history",
		gomcp.WithDescription("Clear the event history buffer."),
	), monH(orch, logger, "clear_event_history", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearEventHistory(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// State Watching (6-10)
	// -------------------------------------------------------------------

	// 6. premiere_watch_playhead_position
	s.AddTool(gomcp.NewTool("premiere_watch_playhead_position",
		gomcp.WithDescription("Start polling the playhead position at a given interval in milliseconds. Results appear in event history."),
		gomcp.WithNumber("interval_ms", gomcp.Description("Polling interval in milliseconds (default: 500)")),
	), monH(orch, logger, "watch_playhead_position", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		intervalMs := gomcp.ParseInt(req, "interval_ms", 500)
		result, err := orch.WatchPlayheadPosition(ctx, intervalMs)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_stop_watch_playhead
	s.AddTool(gomcp.NewTool("premiere_stop_watch_playhead",
		gomcp.WithDescription("Stop the playhead position watcher."),
	), monH(orch, logger, "stop_watch_playhead", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.StopWatchPlayhead(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_watch_render_progress
	s.AddTool(gomcp.NewTool("premiere_watch_render_progress",
		gomcp.WithDescription("Start watching render progress at a given interval in milliseconds."),
		gomcp.WithNumber("interval_ms", gomcp.Description("Polling interval in milliseconds (default: 1000)")),
	), monH(orch, logger, "watch_render_progress", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		intervalMs := gomcp.ParseInt(req, "interval_ms", 1000)
		result, err := orch.WatchRenderProgress(ctx, intervalMs)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_stop_watch_render
	s.AddTool(gomcp.NewTool("premiere_stop_watch_render",
		gomcp.WithDescription("Stop the render progress watcher."),
	), monH(orch, logger, "stop_watch_render", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.StopWatchRender(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_get_state_snapshot
	s.AddTool(gomcp.NewTool("premiere_get_state_snapshot",
		gomcp.WithDescription("Get a complete state snapshot including project, active sequence, playhead position, and current selection."),
	), monH(orch, logger, "get_state_snapshot", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetStateSnapshot(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Project State (11-14)
	// -------------------------------------------------------------------

	// 11. premiere_is_project_modified
	s.AddTool(gomcp.NewTool("premiere_is_project_modified",
		gomcp.WithDescription("Check if the current project has unsaved changes."),
	), monH(orch, logger, "is_project_modified", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.IsProjectModified(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_get_project_duration
	s.AddTool(gomcp.NewTool("premiere_get_project_duration",
		gomcp.WithDescription("Get the total project duration across all sequences."),
	), monH(orch, logger, "get_project_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectDuration(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_get_project_stats
	s.AddTool(gomcp.NewTool("premiere_get_project_stats",
		gomcp.WithDescription("Get project statistics including total clips, sequences, bins, and effects counts."),
	), monH(orch, logger, "get_project_stats", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectStats(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_get_recent_actions
	s.AddTool(gomcp.NewTool("premiere_get_recent_actions",
		gomcp.WithDescription("Get recent user actions from the event history."),
		gomcp.WithNumber("count", gomcp.Description("Number of recent actions to return (default: 20)")),
	), monH(orch, logger, "get_recent_actions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		count := gomcp.ParseInt(req, "count", 20)
		result, err := orch.GetRecentActions(ctx, count)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Sequence State - Extended (15-20)
	// -------------------------------------------------------------------

	// 15. premiere_get_active_track_targets
	s.AddTool(gomcp.NewTool("premiere_get_active_track_targets",
		gomcp.WithDescription("Get which tracks are targeted for insert/overwrite operations."),
	), monH(orch, logger, "get_active_track_targets", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetActiveTrackTargets(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_set_active_track_targets
	s.AddTool(gomcp.NewTool("premiere_set_active_track_targets",
		gomcp.WithDescription("Set which tracks are targeted for insert/overwrite operations."),
		gomcp.WithString("video_targets", gomcp.Description("JSON array of video track indices to target (e.g. '[0,1]')")),
		gomcp.WithString("audio_targets", gomcp.Description("JSON array of audio track indices to target (e.g. '[0,1]')")),
	), monH(orch, logger, "set_active_track_targets", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		videoTargets := gomcp.ParseString(req, "video_targets", "[]")
		audioTargets := gomcp.ParseString(req, "audio_targets", "[]")
		result, err := orch.SetActiveTrackTargets(ctx, videoTargets, audioTargets)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_get_track_heights
	s.AddTool(gomcp.NewTool("premiere_get_track_heights",
		gomcp.WithDescription("Get the height/mute state of all tracks in the active sequence."),
	), monH(orch, logger, "get_track_heights", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTrackHeights(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_set_track_heights
	s.AddTool(gomcp.NewTool("premiere_set_track_heights",
		gomcp.WithDescription("Set track heights/mute state for tracks in the active sequence."),
		gomcp.WithString("track_type", gomcp.Description("Track type: video or audio (default: video)"), gomcp.Enum("video", "audio")),
		gomcp.WithString("heights", gomcp.Required(), gomcp.Description("JSON array of objects with trackIndex and muted fields")),
	), monH(orch, logger, "set_track_heights", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "video")
		heights := gomcp.ParseString(req, "heights", "[]")
		if heights == "[]" {
			return gomcp.NewToolResultError("parameter 'heights' is required"), nil
		}
		result, err := orch.SetTrackHeights(ctx, trackType, heights)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_is_sequence_modified
	s.AddTool(gomcp.NewTool("premiere_is_sequence_modified",
		gomcp.WithDescription("Check if the active sequence has unsaved changes."),
	), monH(orch, logger, "is_sequence_modified", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.IsSequenceModified(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_get_sequence_hash
	s.AddTool(gomcp.NewTool("premiere_get_sequence_hash",
		gomcp.WithDescription("Get a hash fingerprint of the active sequence state for change detection."),
	), monH(orch, logger, "get_sequence_hash", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceHash(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Clip State (21-25)
	// -------------------------------------------------------------------

	// 21. premiere_get_clip_under_playhead
	s.AddTool(gomcp.NewTool("premiere_get_clip_under_playhead",
		gomcp.WithDescription("Get information about all clips at the current playhead position across all tracks."),
	), monH(orch, logger, "get_clip_under_playhead", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipUnderPlayhead(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_get_clip_at_time
	s.AddTool(gomcp.NewTool("premiere_get_clip_at_time",
		gomcp.WithDescription("Get clip information at a specific time on a specific track."),
		gomcp.WithString("track_type", gomcp.Description("Track type: video or audio (default: video)"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("seconds", gomcp.Required(), gomcp.Description("Time position in seconds")),
	), monH(orch, logger, "get_clip_at_time", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		seconds := gomcp.ParseFloat64(req, "seconds", 0)
		result, err := orch.GetClipAtTime(ctx, trackType, trackIndex, seconds)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_get_adjacent_clips
	s.AddTool(gomcp.NewTool("premiere_get_adjacent_clips",
		gomcp.WithDescription("Get the previous and next clips adjacent to a specified clip."),
		gomcp.WithString("track_type", gomcp.Description("Track type: video or audio (default: video)"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
	), monH(orch, logger, "get_adjacent_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.GetAdjacentClips(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_is_clip_selected
	s.AddTool(gomcp.NewTool("premiere_is_clip_selected",
		gomcp.WithDescription("Check if a specific clip is currently selected in the timeline."),
		gomcp.WithString("track_type", gomcp.Description("Track type: video or audio (default: video)"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
	), monH(orch, logger, "is_clip_selected", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.IsClipSelected(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_get_clip_properties
	s.AddTool(gomcp.NewTool("premiere_get_clip_properties",
		gomcp.WithDescription("Get all properties of a clip including effects, timing, and metadata as JSON."),
		gomcp.WithString("track_type", gomcp.Description("Track type: video or audio (default: video)"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
	), monH(orch, logger, "get_clip_properties", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.GetClipProperties(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Notifications (26-30)
	// -------------------------------------------------------------------

	// 26. premiere_show_notification
	s.AddTool(gomcp.NewTool("premiere_show_notification",
		gomcp.WithDescription("Show a notification in the Premiere Pro Events panel."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Notification title")),
		gomcp.WithString("message", gomcp.Description("Notification message body")),
	), monH(orch, logger, "show_notification", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		message := gomcp.ParseString(req, "message", "")
		result, err := orch.ShowNotification(ctx, title, message)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_log_to_events_panel
	s.AddTool(gomcp.NewTool("premiere_log_to_events_panel",
		gomcp.WithDescription("Log a message to the Premiere Pro Events panel at a specified severity level."),
		gomcp.WithString("message", gomcp.Required(), gomcp.Description("Message to log")),
		gomcp.WithString("level", gomcp.Description("Log level: info, warning, or error (default: info)"), gomcp.Enum("info", "warning", "error")),
	), monH(orch, logger, "log_to_events_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		message := gomcp.ParseString(req, "message", "")
		if message == "" {
			return gomcp.NewToolResultError("parameter 'message' is required"), nil
		}
		level := gomcp.ParseString(req, "level", "info")
		result, err := orch.LogToEventsPanel(ctx, message, level)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_show_progress_bar
	s.AddTool(gomcp.NewTool("premiere_show_progress_bar",
		gomcp.WithDescription("Show a progress bar notification in the Premiere Pro Events panel."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Progress bar title")),
		gomcp.WithNumber("current", gomcp.Required(), gomcp.Description("Current progress value")),
		gomcp.WithNumber("total", gomcp.Required(), gomcp.Description("Total progress value")),
	), monH(orch, logger, "show_progress_bar", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		current := gomcp.ParseInt(req, "current", 0)
		total := gomcp.ParseInt(req, "total", 100)
		result, err := orch.ShowProgressBar(ctx, title, current, total)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_hide_progress_bar
	s.AddTool(gomcp.NewTool("premiere_hide_progress_bar",
		gomcp.WithDescription("Hide the progress bar and log completion to the Events panel."),
	), monH(orch, logger, "hide_progress_bar", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.HideProgressBar(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_show_dialog
	s.AddTool(gomcp.NewTool("premiere_show_dialog",
		gomcp.WithDescription("Show a dialog in Premiere Pro with a title, message, and custom buttons."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Dialog title")),
		gomcp.WithString("message", gomcp.Description("Dialog message body")),
		gomcp.WithString("buttons", gomcp.Description("Comma-separated button labels (e.g. 'OK,Cancel'). Default: 'OK'")),
	), monH(orch, logger, "show_dialog", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		message := gomcp.ParseString(req, "message", "")
		buttons := gomcp.ParseString(req, "buttons", "OK")
		result, err := orch.ShowDialog(ctx, title, message, buttons)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
