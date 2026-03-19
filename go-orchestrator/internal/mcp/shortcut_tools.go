package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// scH is a small handler wrapper for shortcut / menu / view-control tools.
func scH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerShortcutTools registers all 30 keyboard-shortcut, command-execution,
// menu-control, quick-action, view-control, and sequence-display MCP tools.
func registerShortcutTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -------------------------------------------------------------------
	// Menu Commands (1-6)
	// -------------------------------------------------------------------

	// 1. premiere_get_menu_items
	s.AddTool(gomcp.NewTool("premiere_get_menu_items",
		gomcp.WithDescription("List all top-level menu items in Premiere Pro (File, Edit, Sequence, etc.). Uses the QE DOM."),
	), scH(orch, logger, "get_menu_items", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMenuItems(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_get_submenu_items
	s.AddTool(gomcp.NewTool("premiere_get_submenu_items",
		gomcp.WithDescription("List items in a submenu. Provide the menu path, e.g. 'File' or 'File/Export'."),
		gomcp.WithString("menu_path", gomcp.Required(), gomcp.Description("Menu path, e.g. 'File' or 'Edit/Preferences'")),
	), scH(orch, logger, "get_submenu_items", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		menuPath := gomcp.ParseString(req, "menu_path", "")
		if menuPath == "" {
			return gomcp.NewToolResultError("parameter 'menu_path' is required"), nil
		}
		result, err := orch.GetSubmenuItems(ctx, menuPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_execute_menu_item
	s.AddTool(gomcp.NewTool("premiere_execute_menu_item",
		gomcp.WithDescription("Execute a menu item by its ID (from the QE DOM). Use premiere_find_menu_item to discover IDs."),
		gomcp.WithString("menu_item_id", gomcp.Required(), gomcp.Description("ID of the menu item to execute")),
	), scH(orch, logger, "execute_menu_item", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		menuItemID := gomcp.ParseString(req, "menu_item_id", "")
		if menuItemID == "" {
			return gomcp.NewToolResultError("parameter 'menu_item_id' is required"), nil
		}
		result, err := orch.ExecuteMenuItemByID(ctx, menuItemID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_find_menu_item
	s.AddTool(gomcp.NewTool("premiere_find_menu_item",
		gomcp.WithDescription("Search for a menu item by name. Returns matching menu items with their IDs and paths."),
		gomcp.WithString("search_text", gomcp.Required(), gomcp.Description("Text to search for in menu item names")),
	), scH(orch, logger, "find_menu_item", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		searchText := gomcp.ParseString(req, "search_text", "")
		if searchText == "" {
			return gomcp.NewToolResultError("parameter 'search_text' is required"), nil
		}
		result, err := orch.FindMenuItem(ctx, searchText)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_is_menu_item_enabled
	s.AddTool(gomcp.NewTool("premiere_is_menu_item_enabled",
		gomcp.WithDescription("Check if a menu item is currently enabled (not grayed out)."),
		gomcp.WithString("menu_item_id", gomcp.Required(), gomcp.Description("ID of the menu item to check")),
	), scH(orch, logger, "is_menu_item_enabled", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		menuItemID := gomcp.ParseString(req, "menu_item_id", "")
		if menuItemID == "" {
			return gomcp.NewToolResultError("parameter 'menu_item_id' is required"), nil
		}
		result, err := orch.IsMenuItemEnabled(ctx, menuItemID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_is_menu_item_checked
	s.AddTool(gomcp.NewTool("premiere_is_menu_item_checked",
		gomcp.WithDescription("Check if a menu item is currently checked/toggled (e.g. snapping, linked selection)."),
		gomcp.WithString("menu_item_id", gomcp.Required(), gomcp.Description("ID of the menu item to check")),
	), scH(orch, logger, "is_menu_item_checked", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		menuItemID := gomcp.ParseString(req, "menu_item_id", "")
		if menuItemID == "" {
			return gomcp.NewToolResultError("parameter 'menu_item_id' is required"), nil
		}
		result, err := orch.IsMenuItemChecked(ctx, menuItemID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Keyboard Shortcuts (7-10)
	// -------------------------------------------------------------------

	// 7. premiere_get_shortcut_for_command
	s.AddTool(gomcp.NewTool("premiere_get_shortcut_for_command",
		gomcp.WithDescription("Get the keyboard shortcut assigned to a specific command ID."),
		gomcp.WithString("command_id", gomcp.Required(), gomcp.Description("Command ID to look up the shortcut for")),
	), scH(orch, logger, "get_shortcut_for_command", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		commandID := gomcp.ParseString(req, "command_id", "")
		if commandID == "" {
			return gomcp.NewToolResultError("parameter 'command_id' is required"), nil
		}
		result, err := orch.GetShortcutForCommand(ctx, commandID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_get_all_shortcuts
	s.AddTool(gomcp.NewTool("premiere_get_all_shortcuts",
		gomcp.WithDescription("List all keyboard shortcuts currently configured in Premiere Pro."),
	), scH(orch, logger, "get_all_shortcuts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAllShortcuts(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_simulate_key_press
	s.AddTool(gomcp.NewTool("premiere_simulate_key_press",
		gomcp.WithDescription("Simulate a keyboard shortcut press. Provide the key and optional modifiers."),
		gomcp.WithString("key", gomcp.Required(), gomcp.Description("Key to press, e.g. 'C', 'V', 'Space', 'Delete', 'Left', 'Right'")),
		gomcp.WithString("modifiers", gomcp.Description("Comma-separated modifiers: ctrl, shift, alt, cmd (e.g. 'ctrl,shift')")),
	), scH(orch, logger, "simulate_key_press", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		key := gomcp.ParseString(req, "key", "")
		if key == "" {
			return gomcp.NewToolResultError("parameter 'key' is required"), nil
		}
		modifiers := gomcp.ParseString(req, "modifiers", "")
		result, err := orch.SimulateKeyPress(ctx, key, modifiers)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_get_shortcut_conflicts
	s.AddTool(gomcp.NewTool("premiere_get_shortcut_conflicts",
		gomcp.WithDescription("Find conflicting keyboard shortcuts where multiple commands share the same key combination."),
	), scH(orch, logger, "get_shortcut_conflicts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetShortcutConflicts(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Quick Actions (11-18)
	// -------------------------------------------------------------------

	// 11. premiere_toggle_full_screen
	s.AddTool(gomcp.NewTool("premiere_toggle_full_screen",
		gomcp.WithDescription("Toggle fullscreen mode on or off in Premiere Pro."),
	), scH(orch, logger, "toggle_full_screen", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ToggleFullScreen(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_toggle_maximize_frame
	s.AddTool(gomcp.NewTool("premiere_toggle_maximize_frame",
		gomcp.WithDescription("Toggle maximize for the currently focused panel/frame in Premiere Pro."),
	), scH(orch, logger, "toggle_maximize_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ToggleMaximizeFrame(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_clear_selection
	s.AddTool(gomcp.NewTool("premiere_clear_selection",
		gomcp.WithDescription("Clear all selections in the active sequence (deselect all clips)."),
	), scH(orch, logger, "clear_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_select_all
	s.AddTool(gomcp.NewTool("premiere_select_all",
		gomcp.WithDescription("Select all clips in the active sequence."),
	), scH(orch, logger, "select_all", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SelectAll(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_cut_selection
	s.AddTool(gomcp.NewTool("premiere_cut_selection",
		gomcp.WithDescription("Cut the currently selected clips in the active sequence."),
	), scH(orch, logger, "cut_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CutSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_copy_selection
	s.AddTool(gomcp.NewTool("premiere_copy_selection",
		gomcp.WithDescription("Copy the currently selected clips in the active sequence."),
	), scH(orch, logger, "copy_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CopySelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_paste_selection
	s.AddTool(gomcp.NewTool("premiere_paste_selection",
		gomcp.WithDescription("Paste clipboard contents at the current playhead position in the active sequence."),
	), scH(orch, logger, "paste_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PasteSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_duplicate_selection
	s.AddTool(gomcp.NewTool("premiere_duplicate_selection",
		gomcp.WithDescription("Duplicate the currently selected clips in place on the timeline."),
	), scH(orch, logger, "duplicate_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.DuplicateSelection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// View Controls (19-24)
	// -------------------------------------------------------------------

	// 19. premiere_set_zoom_level
	s.AddTool(gomcp.NewTool("premiere_set_zoom_level",
		gomcp.WithDescription("Set the timeline zoom level as a percentage (0-100, where 100 is fully zoomed in)."),
		gomcp.WithNumber("level", gomcp.Required(), gomcp.Description("Zoom level percentage (0-100)")),
	), scH(orch, logger, "set_zoom_level", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		level := gomcp.ParseFloat64(req, "level", 50)
		result, err := orch.SetZoomLevel(ctx, level)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_get_zoom_level
	s.AddTool(gomcp.NewTool("premiere_get_zoom_level",
		gomcp.WithDescription("Get the current timeline zoom level as a percentage."),
	), scH(orch, logger, "get_zoom_level", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetZoomLevel(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_scroll_timeline_to
	s.AddTool(gomcp.NewTool("premiere_scroll_timeline_to",
		gomcp.WithDescription("Scroll the timeline view to a specific time position in seconds."),
		gomcp.WithNumber("seconds", gomcp.Required(), gomcp.Description("Time position in seconds to scroll to")),
	), scH(orch, logger, "scroll_timeline_to", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seconds := gomcp.ParseFloat64(req, "seconds", 0)
		result, err := orch.ScrollTimelineTo(ctx, seconds)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_enable_linked_selection
	s.AddTool(gomcp.NewTool("premiere_enable_linked_selection",
		gomcp.WithDescription("Toggle linked selection mode. When enabled, selecting a video clip also selects its linked audio."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable linked selection, false to disable")),
	), scH(orch, logger, "enable_linked_selection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		enabled := gomcp.ParseBoolean(req, "enabled", true)
		result, err := orch.EnableLinkedSelection(ctx, enabled)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_get_linked_selection_state
	s.AddTool(gomcp.NewTool("premiere_get_linked_selection_state",
		gomcp.WithDescription("Get the current linked selection state (whether selecting a clip also selects its linked counterpart)."),
	), scH(orch, logger, "get_linked_selection_state", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetLinkedSelectionState(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_enable_insert_and_overwrite
	s.AddTool(gomcp.NewTool("premiere_enable_insert_and_overwrite",
		gomcp.WithDescription("Toggle insert/overwrite targeting for a specific track."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type: 'video' or 'audio'"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable targeting, false to disable")),
	), scH(orch, logger, "enable_insert_and_overwrite", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		enabled := gomcp.ParseBoolean(req, "enabled", true)
		result, err := orch.EnableInsertAndOverwrite(ctx, trackType, trackIndex, enabled)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Sequence Display (25-30)
	// -------------------------------------------------------------------

	// 25. premiere_show_audio_time_units
	s.AddTool(gomcp.NewTool("premiere_show_audio_time_units",
		gomcp.WithDescription("Toggle display of audio time units (samples) in the timeline."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show audio time units, false to hide")),
	), scH(orch, logger, "show_audio_time_units", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		enabled := gomcp.ParseBoolean(req, "enabled", false)
		result, err := orch.ShowAudioTimeUnits(ctx, enabled)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_show_duplicate_frame_markers
	s.AddTool(gomcp.NewTool("premiere_show_duplicate_frame_markers",
		gomcp.WithDescription("Toggle display of duplicate frame markers (colored bars on clips that share the same source frames)."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show duplicate frame markers, false to hide")),
	), scH(orch, logger, "show_duplicate_frame_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		enabled := gomcp.ParseBoolean(req, "enabled", false)
		result, err := orch.ShowDuplicateFrameMarkers(ctx, enabled)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_show_clip_mismatch_warning
	s.AddTool(gomcp.NewTool("premiere_show_clip_mismatch_warning",
		gomcp.WithDescription("Toggle clip mismatch warning indicators on timeline clips that differ from the sequence settings."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show mismatch warnings, false to hide")),
	), scH(orch, logger, "show_clip_mismatch_warning", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		enabled := gomcp.ParseBoolean(req, "enabled", false)
		result, err := orch.ShowClipMismatchWarning(ctx, enabled)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_set_timeline_snap
	s.AddTool(gomcp.NewTool("premiere_set_timeline_snap",
		gomcp.WithDescription("Set the snap type for the timeline (controls what items snap to when dragging)."),
		gomcp.WithString("snap_type", gomcp.Required(), gomcp.Description("Snap type"), gomcp.Enum("none", "frame", "marker", "clip")),
	), scH(orch, logger, "set_timeline_snap", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		snapType := gomcp.ParseString(req, "snap_type", "frame")
		result, err := orch.SetTimelineSnap(ctx, snapType)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_get_timeline_view_extents
	s.AddTool(gomcp.NewTool("premiere_get_timeline_view_extents",
		gomcp.WithDescription("Get the currently visible timeline range (start and end times in seconds)."),
	), scH(orch, logger, "get_timeline_view_extents", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTimelineViewExtents(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_set_timeline_view_extents
	s.AddTool(gomcp.NewTool("premiere_set_timeline_view_extents",
		gomcp.WithDescription("Set the visible timeline range by specifying start and end times in seconds."),
		gomcp.WithNumber("start_seconds", gomcp.Required(), gomcp.Description("Start time in seconds for the visible range")),
		gomcp.WithNumber("end_seconds", gomcp.Required(), gomcp.Description("End time in seconds for the visible range")),
	), scH(orch, logger, "set_timeline_view_extents", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		startSeconds := gomcp.ParseFloat64(req, "start_seconds", 0)
		endSeconds := gomcp.ParseFloat64(req, "end_seconds", 10)
		result, err := orch.SetTimelineViewExtents(ctx, startSeconds, endSeconds)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
