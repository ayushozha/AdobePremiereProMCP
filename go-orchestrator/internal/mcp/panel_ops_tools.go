package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// poH is a small handler wrapper for panel-ops tools.
func poH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerPanelOpsTools registers the 15 Timeline Panel Menu and Panel
// Docking MCP tools.
func registerPanelOpsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -------------------------------------------------------------------
	// Timeline Display — QE DOM (1-7)
	// -------------------------------------------------------------------

	// 1. premiere_set_audio_waveform_label_color
	s.AddTool(gomcp.NewTool("premiere_set_audio_waveform_label_color",
		gomcp.WithDescription("Toggle whether audio waveforms on the timeline use the clip's label color."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to colour waveforms by label, false for default")),
	), poH(orch, logger, "set_audio_waveform_label_color", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAudioWaveformLabelColor(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_set_logarithmic_waveform_scaling
	s.AddTool(gomcp.NewTool("premiere_set_logarithmic_waveform_scaling",
		gomcp.WithDescription("Toggle logarithmic scaling for audio waveforms on the timeline."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true for logarithmic scaling, false for linear")),
	), poH(orch, logger, "set_logarithmic_waveform_scaling", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLogarithmicWaveformScaling(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_set_time_ruler_numbers
	s.AddTool(gomcp.NewTool("premiere_set_time_ruler_numbers",
		gomcp.WithDescription("Show or hide the time ruler numbers above the timeline."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show numbers, false to hide")),
	), poH(orch, logger, "set_time_ruler_numbers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTimeRulerNumbers(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_set_multi_camera_audio_follows_video
	s.AddTool(gomcp.NewTool("premiere_set_multi_camera_audio_follows_video",
		gomcp.WithDescription("Toggle whether the audio track follows the video angle in a multicam sequence."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to have audio follow video, false for independent")),
	), poH(orch, logger, "set_multi_camera_audio_follows_video", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetMultiCameraAudioFollowsVideo(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_set_multi_camera_selection_top_panel
	s.AddTool(gomcp.NewTool("premiere_set_multi_camera_selection_top_panel",
		gomcp.WithDescription("Toggle whether the multicam source selection appears in the top panel."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to show selection in top panel, false to hide")),
	), poH(orch, logger, "set_multi_camera_selection_top_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetMultiCameraSelectionTopPanel(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_set_multi_camera_follows_nest_setting
	s.AddTool(gomcp.NewTool("premiere_set_multi_camera_follows_nest_setting",
		gomcp.WithDescription("Toggle whether multicam editing follows the nested sequence setting."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to follow nest setting, false for independent")),
	), poH(orch, logger, "set_multi_camera_follows_nest_setting", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetMultiCameraFollowsNestSetting(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_set_rectified_audio_waveforms
	s.AddTool(gomcp.NewTool("premiere_set_rectified_audio_waveforms",
		gomcp.WithDescription("Toggle rectified (absolute value) display of audio waveforms on the timeline."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true for rectified waveforms, false for standard")),
	), poH(orch, logger, "set_rectified_audio_waveforms", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetRectifiedAudioWaveforms(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Panel Docking via AppleScript (8-15)
	// -------------------------------------------------------------------

	// 8. premiere_undock_panel
	s.AddTool(gomcp.NewTool("premiere_undock_panel",
		gomcp.WithDescription("Undock the currently focused panel from its group via macOS accessibility (AppleScript)."),
		gomcp.WithString("panel_name", gomcp.Description("Name of the panel to undock (informational)")),
	), poH(orch, logger, "undock_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		result, err := orch.UndockPanel(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_close_other_panels_in_group
	s.AddTool(gomcp.NewTool("premiere_close_other_panels_in_group",
		gomcp.WithDescription("Close all other panels in the same panel group, keeping only the active one."),
		gomcp.WithString("panel_name", gomcp.Description("Name of the panel whose group to modify (informational)")),
	), poH(orch, logger, "close_other_panels_in_group", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		result, err := orch.CloseOtherPanelsInGroup(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_close_panel_group
	s.AddTool(gomcp.NewTool("premiere_close_panel_group",
		gomcp.WithDescription("Close the entire panel group that contains the currently focused panel."),
		gomcp.WithString("panel_name", gomcp.Description("Name of the panel whose group to close (informational)")),
	), poH(orch, logger, "close_panel_group", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		result, err := orch.ClosePanelGroup(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_undock_panel_group
	s.AddTool(gomcp.NewTool("premiere_undock_panel_group",
		gomcp.WithDescription("Undock an entire panel group into a floating window."),
		gomcp.WithString("panel_name", gomcp.Description("Name of the panel whose group to undock (informational)")),
	), poH(orch, logger, "undock_panel_group", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		result, err := orch.UndockPanelGroup(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_maximize_panel_group
	s.AddTool(gomcp.NewTool("premiere_maximize_panel_group",
		gomcp.WithDescription("Toggle maximize on the panel group that contains the currently focused panel."),
		gomcp.WithString("panel_name", gomcp.Description("Name of the panel whose group to maximize (informational)")),
	), poH(orch, logger, "maximize_panel_group", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		panelName := gomcp.ParseString(req, "panel_name", "")
		result, err := orch.MaximizePanelGroup(ctx, panelName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_set_stacked_panels
	s.AddTool(gomcp.NewTool("premiere_set_stacked_panels",
		gomcp.WithDescription("Toggle stacked panel layout for the active panel group."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable stacked panels, false for tabbed")),
	), poH(orch, logger, "set_stacked_panels", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetStackedPanels(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_set_small_tabs
	s.AddTool(gomcp.NewTool("premiere_set_small_tabs",
		gomcp.WithDescription("Toggle small tab display for panel groups."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true for small tabs, false for normal size")),
	), poH(orch, logger, "set_small_tabs", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetSmallTabs(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_simulate_menu_click
	s.AddTool(gomcp.NewTool("premiere_simulate_menu_click",
		gomcp.WithDescription("Simulate clicking a Premiere Pro menu item via macOS accessibility (AppleScript). Use slash-separated paths like 'Window/Extensions/MyPanel'."),
		gomcp.WithString("menu_path", gomcp.Required(), gomcp.Description("Slash-separated menu path, e.g. 'File/Save' or 'Window/Extensions/MyPanel'")),
	), poH(orch, logger, "simulate_menu_click", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		menuPath := gomcp.ParseString(req, "menu_path", "")
		if menuPath == "" {
			return gomcp.NewToolResultError("parameter 'menu_path' is required"), nil
		}
		result, err := orch.SimulateMenuClick(ctx, menuPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
