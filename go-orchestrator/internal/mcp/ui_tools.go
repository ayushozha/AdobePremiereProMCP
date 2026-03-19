package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// uiH is a small handler wrapper for UI tools.
func uiH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerUITools registers all 30 accessibility, UX enhancement, and user
// interface control MCP tools.
func registerUITools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -------------------------------------------------------------------
	// UI Panel Control (1-5)
	// -------------------------------------------------------------------

	// 1. premiere_open_panel
	s.AddTool(gomcp.NewTool("premiere_open_panel",
		gomcp.WithDescription("Open a Premiere Pro panel by name (e.g. Effect Controls, Lumetri Color, Audio Track Mixer, Essential Graphics)."),
		gomcp.WithString("panel_name", gomcp.Required(), gomcp.Description("Name of the panel to open")),
	), uiH(orch, logger, "open_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		if panelName == "" {
			return gomcp.NewToolResultError("parameter 'panel_name' is required"), nil
		}
		result, err := orch.OpenPanel(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_close_panel
	s.AddTool(gomcp.NewTool("premiere_close_panel",
		gomcp.WithDescription("Close a Premiere Pro panel by name."),
		gomcp.WithString("panel_name", gomcp.Required(), gomcp.Description("Name of the panel to close")),
	), uiH(orch, logger, "close_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		if panelName == "" {
			return gomcp.NewToolResultError("parameter 'panel_name' is required"), nil
		}
		result, err := orch.ClosePanel(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_open_panels
	s.AddTool(gomcp.NewTool("premiere_get_open_panels",
		gomcp.WithDescription("List all currently open panels in the Premiere Pro workspace."),
	), uiH(orch, logger, "get_open_panels", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetOpenPanels(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_reset_panel_layout
	s.AddTool(gomcp.NewTool("premiere_reset_panel_layout",
		gomcp.WithDescription("Reset the Premiere Pro workspace to the default panel layout."),
	), uiH(orch, logger, "reset_panel_layout", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ResetPanelLayout(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_maximize_panel
	s.AddTool(gomcp.NewTool("premiere_maximize_panel",
		gomcp.WithDescription("Maximize a Premiere Pro panel to fill the entire application window."),
		gomcp.WithString("panel_name", gomcp.Required(), gomcp.Description("Name of the panel to maximize")),
	), uiH(orch, logger, "maximize_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		if panelName == "" {
			return gomcp.NewToolResultError("parameter 'panel_name' is required"), nil
		}
		result, err := orch.MaximizePanel(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Window Management (6-10)
	// -------------------------------------------------------------------

	// 6. premiere_get_window_info
	s.AddTool(gomcp.NewTool("premiere_get_window_info",
		gomcp.WithDescription("Get the main Premiere Pro window size, position, and display information."),
	), uiH(orch, logger, "get_window_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetWindowInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_set_window_size
	s.AddTool(gomcp.NewTool("premiere_set_window_size",
		gomcp.WithDescription("Set the main Premiere Pro window size in pixels."),
		gomcp.WithNumber("width", gomcp.Required(), gomcp.Description("Window width in pixels")),
		gomcp.WithNumber("height", gomcp.Required(), gomcp.Description("Window height in pixels")),
	), uiH(orch, logger, "set_window_size", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetWindowSize(ctx, gomcp.ParseInt(req, "width", 1920), gomcp.ParseInt(req, "height", 1080))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_minimize_window
	s.AddTool(gomcp.NewTool("premiere_minimize_window",
		gomcp.WithDescription("Minimize the Premiere Pro application window."),
	), uiH(orch, logger, "minimize_window", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.MinimizeWindow(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_bring_to_front
	s.AddTool(gomcp.NewTool("premiere_bring_to_front",
		gomcp.WithDescription("Bring the Premiere Pro application window to the front of all other windows."),
	), uiH(orch, logger, "bring_to_front", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.BringToFront(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_enter_fullscreen
	s.AddTool(gomcp.NewTool("premiere_enter_fullscreen",
		gomcp.WithDescription("Enter fullscreen mode in Premiere Pro."),
	), uiH(orch, logger, "enter_fullscreen", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.EnterFullscreen(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Timeline UI (11-15)
	// -------------------------------------------------------------------

	// 11. premiere_set_track_height
	s.AddTool(gomcp.NewTool("premiere_set_track_height",
		gomcp.WithDescription("Set the display height of a specific track in the timeline."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type: video or audio"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("height", gomcp.Required(), gomcp.Description("Track height in pixels")),
	), uiH(orch, logger, "set_track_height", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTrackHeight(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "height", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_collapse_track
	s.AddTool(gomcp.NewTool("premiere_collapse_track",
		gomcp.WithDescription("Collapse a specific track in the timeline to its minimum height."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type: video or audio"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), uiH(orch, logger, "collapse_track", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CollapseTrack(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_expand_track
	s.AddTool(gomcp.NewTool("premiere_expand_track",
		gomcp.WithDescription("Expand a specific track in the timeline to show full detail."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type: video or audio"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
	), uiH(orch, logger, "expand_track", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExpandTrack(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_collapse_all_tracks
	s.AddTool(gomcp.NewTool("premiere_collapse_all_tracks",
		gomcp.WithDescription("Collapse all video and audio tracks in the active sequence timeline."),
	), uiH(orch, logger, "collapse_all_tracks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CollapseAllTracks(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_expand_all_tracks
	s.AddTool(gomcp.NewTool("premiere_expand_all_tracks",
		gomcp.WithDescription("Expand all video and audio tracks in the active sequence timeline."),
	), uiH(orch, logger, "expand_all_tracks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExpandAllTracks(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Label Management (16-19)
	// -------------------------------------------------------------------

	// 16. premiere_set_label_preferences
	s.AddTool(gomcp.NewTool("premiere_set_label_preferences",
		gomcp.WithDescription("Set all label color names for the project. Provide a JSON string mapping color indices to custom names."),
		gomcp.WithString("labels_json", gomcp.Required(), gomcp.Description("JSON string mapping label indices to names, e.g. '{\"0\":\"Interviews\",\"1\":\"B-Roll\"}'")),
	), uiH(orch, logger, "set_label_preferences", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		labelsJSON := gomcp.ParseString(req, "labels_json", "")
		if labelsJSON == "" {
			return gomcp.NewToolResultError("parameter 'labels_json' is required"), nil
		}
		result, err := orch.SetLabelPreferences(ctx, labelsJSON)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_get_active_label_filter
	s.AddTool(gomcp.NewTool("premiere_get_active_label_filter",
		gomcp.WithDescription("Get the currently active label color filter applied to the project panel."),
	), uiH(orch, logger, "get_active_label_filter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetActiveLabelFilter(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_set_label_filter
	s.AddTool(gomcp.NewTool("premiere_set_label_filter",
		gomcp.WithDescription("Filter the project panel to show only items with a specific label color index."),
		gomcp.WithNumber("color_index", gomcp.Required(), gomcp.Description("Zero-based label color index to filter by")),
	), uiH(orch, logger, "set_label_filter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLabelFilter(ctx, gomcp.ParseInt(req, "color_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_clear_label_filter
	s.AddTool(gomcp.NewTool("premiere_clear_label_filter",
		gomcp.WithDescription("Clear the label color filter from the project panel, showing all items."),
	), uiH(orch, logger, "clear_label_filter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearLabelFilter(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Timeline Display (20-23)
	// -------------------------------------------------------------------

	// 20. premiere_set_time_display_format_ui
	s.AddTool(gomcp.NewTool("premiere_set_time_display_format_ui",
		gomcp.WithDescription("Set the timecode display format on the timeline UI (timecode, frames, feet+frames)."),
		gomcp.WithString("format", gomcp.Required(), gomcp.Description("Display format"), gomcp.Enum("timecode", "frames", "feet+frames")),
	), uiH(orch, logger, "set_time_display_format_ui", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		format := gomcp.ParseString(req, "format", "")
		if format == "" {
			return gomcp.NewToolResultError("parameter 'format' is required"), nil
		}
		result, err := orch.SetTimeDisplayFormat(ctx, format)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_set_audio_waveform_display
	s.AddTool(gomcp.NewTool("premiere_set_audio_waveform_display",
		gomcp.WithDescription("Show or hide audio waveforms on the timeline tracks."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show waveforms, false to hide")),
	), uiH(orch, logger, "set_audio_waveform_display", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAudioWaveformDisplay(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_set_video_thumbnail_display
	s.AddTool(gomcp.NewTool("premiere_set_video_thumbnail_display",
		gomcp.WithDescription("Show or hide video thumbnail previews on the timeline tracks."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show thumbnails, false to hide")),
	), uiH(orch, logger, "set_video_thumbnail_display", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetVideoThumbnailDisplay(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_set_track_name_display
	s.AddTool(gomcp.NewTool("premiere_set_track_name_display",
		gomcp.WithDescription("Show or hide track name labels on the timeline."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show track names, false to hide")),
	), uiH(orch, logger, "set_track_name_display", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTrackNameDisplay(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// User Feedback (24-28)
	// -------------------------------------------------------------------

	// 24. premiere_show_alert
	s.AddTool(gomcp.NewTool("premiere_show_alert",
		gomcp.WithDescription("Show an alert dialog in Premiere Pro with a title and message."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Alert dialog title")),
		gomcp.WithString("message", gomcp.Required(), gomcp.Description("Alert dialog message body")),
	), uiH(orch, logger, "show_alert", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		message := gomcp.ParseString(req, "message", "")
		if message == "" {
			return gomcp.NewToolResultError("parameter 'message' is required"), nil
		}
		result, err := orch.ShowAlert(ctx, title, message)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_show_confirm_dialog
	s.AddTool(gomcp.NewTool("premiere_show_confirm_dialog",
		gomcp.WithDescription("Show a confirmation dialog (yes/no) in Premiere Pro and return the user's choice."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Dialog title")),
		gomcp.WithString("message", gomcp.Required(), gomcp.Description("Dialog message body")),
	), uiH(orch, logger, "show_confirm_dialog", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		message := gomcp.ParseString(req, "message", "")
		if message == "" {
			return gomcp.NewToolResultError("parameter 'message' is required"), nil
		}
		result, err := orch.ShowConfirmDialog(ctx, title, message)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_show_input_dialog
	s.AddTool(gomcp.NewTool("premiere_show_input_dialog",
		gomcp.WithDescription("Show an input dialog in Premiere Pro that prompts the user for text input."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Dialog title")),
		gomcp.WithString("prompt", gomcp.Required(), gomcp.Description("Prompt text shown to the user")),
		gomcp.WithString("default_value", gomcp.Description("Default value pre-filled in the input field")),
	), uiH(orch, logger, "show_input_dialog", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		prompt := gomcp.ParseString(req, "prompt", "")
		if prompt == "" {
			return gomcp.NewToolResultError("parameter 'prompt' is required"), nil
		}
		defaultValue := gomcp.ParseString(req, "default_value", "")
		result, err := orch.ShowInputDialog(ctx, title, prompt, defaultValue)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_show_progress_dialog
	s.AddTool(gomcp.NewTool("premiere_show_progress_dialog",
		gomcp.WithDescription("Show a progress dialog in Premiere Pro with a title, message, and progress percentage."),
		gomcp.WithString("title", gomcp.Required(), gomcp.Description("Progress dialog title")),
		gomcp.WithString("message", gomcp.Required(), gomcp.Description("Progress dialog message")),
		gomcp.WithNumber("progress", gomcp.Required(), gomcp.Description("Progress value from 0.0 to 1.0")),
	), uiH(orch, logger, "show_progress_dialog", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		title := gomcp.ParseString(req, "title", "")
		if title == "" {
			return gomcp.NewToolResultError("parameter 'title' is required"), nil
		}
		message := gomcp.ParseString(req, "message", "")
		if message == "" {
			return gomcp.NewToolResultError("parameter 'message' is required"), nil
		}
		progress := gomcp.ParseFloat64(req, "progress", 0)
		result, err := orch.ShowProgressDialog(ctx, title, message, progress)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_write_to_console
	s.AddTool(gomcp.NewTool("premiere_write_to_console",
		gomcp.WithDescription("Write a message to the ExtendScript console in Premiere Pro for debugging."),
		gomcp.WithString("message", gomcp.Required(), gomcp.Description("Message to write to the console")),
	), uiH(orch, logger, "write_to_console", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		message := gomcp.ParseString(req, "message", "")
		if message == "" {
			return gomcp.NewToolResultError("parameter 'message' is required"), nil
		}
		result, err := orch.WriteToConsole(ctx, message)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Accessibility (29-30)
	// -------------------------------------------------------------------

	// 29. premiere_get_ui_scaling
	s.AddTool(gomcp.NewTool("premiere_get_ui_scaling",
		gomcp.WithDescription("Get the current UI scaling factor for Premiere Pro's interface elements."),
	), uiH(orch, logger, "get_ui_scaling", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetUIScaling(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_set_high_contrast_mode
	s.AddTool(gomcp.NewTool("premiere_set_high_contrast_mode",
		gomcp.WithDescription("Enable or disable high contrast mode in Premiere Pro for improved accessibility."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable high contrast mode, false to disable")),
	), uiH(orch, logger, "set_high_contrast_mode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetHighContrastMode(ctx, gomcp.ParseBoolean(req, "enabled", false))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
