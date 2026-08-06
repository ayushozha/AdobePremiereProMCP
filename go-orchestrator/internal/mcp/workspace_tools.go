package mcp

import (
	"context"
	"fmt"
	"math"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerWorkspaceTools registers all multicam, proxy workflow, workspace,
// undo/redo, project panel, source monitor, preferences, and media cache
// MCP tools.
func registerWorkspaceTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -------------------------------------------------------------------
	// Multicam (1-4)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_create_multicam_sequence",
		gomcp.WithDescription("Create a multicam source sequence from multiple clips. Requires at least 2 clips."),
		gomcp.WithString("name", gomcp.Description("Name for the multicam sequence (default: Multicam Sequence)")),
		gomcp.WithArray("clip_indices", gomcp.Required(), gomcp.Description("Array of zero-based project item indices to include"), gomcp.WithNumberItems()),
		gomcp.WithString("sync_point", gomcp.Description("Sync point method"), gomcp.Enum("inPoint", "outPoint", "timecode", "marker")),
	), wsH(orch, logger, "create_multicam_sequence", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "Multicam Sequence")
		syncPoint := gomcp.ParseString(req, "sync_point", "inPoint")
		clipIndices, err := extractIntSlice(req, "clip_indices")
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("invalid clip_indices: %v", err)), nil
		}
		result, err := orch.CreateMulticamSequence(ctx, name, clipIndices, syncPoint)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_switch_multicam_angle",
		gomcp.WithDescription("Switch the active camera angle on a multicam clip at a given time."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time in seconds where the angle switch occurs")),
		gomcp.WithNumber("angle_index", gomcp.Required(), gomcp.Description("Zero-based camera angle index to switch to")),
	), wsH(orch, logger, "switch_multicam_angle", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SwitchMulticamAngle(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "time", 0), gomcp.ParseInt(req, "angle_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_flatten_multicam",
		gomcp.WithDescription("Flatten a multicam sequence into a regular (non-multicam) sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index to flatten")),
	), wsH(orch, logger, "flatten_multicam", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FlattenMulticam(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_multicam_angles",
		gomcp.WithDescription("List all available camera angles for a multicam clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), wsH(orch, logger, "get_multicam_angles", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMulticamAngles(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Proxy Workflow (5-10)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_create_proxy",
		gomcp.WithDescription("Queue asynchronous proxy generation for a root-bin project item through Adobe Media Encoder. Poll the explicit output path until the file is stable, then call premiere_attach_proxy and verify its canonical path."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Min(0), gomcp.Description("Zero-based index of a proxy-capable project item in the root bin")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.MinLength(1), gomcp.Description("Absolute path for a new proxy output file; its parent directory must exist and an existing file will not be overwritten")),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.MinLength(1), gomcp.Description("Absolute path to an existing .epr encoder preset")),
	), wsH(orch, logger, "create_proxy", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		projectItemIndex, err := proxyProjectItemIndex(req)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		presetPath := strings.TrimSpace(gomcp.ParseString(req, "preset_path", ""))
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		outputPath := strings.TrimSpace(gomcp.ParseString(req, "output_path", ""))
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.CreateProxy(ctx, projectItemIndex, outputPath, presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_attach_proxy",
		gomcp.WithDescription("Attach an existing local proxy file to a root-bin project item, then verify hasProxy and the canonical attached path."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Min(0), gomcp.Description("Zero-based index of a proxy-capable project item in the root bin")),
		gomcp.WithString("proxy_path", gomcp.Required(), gomcp.MinLength(1), gomcp.Description("Absolute path to an existing proxy media file")),
	), wsH(orch, logger, "attach_proxy", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		projectItemIndex, err := proxyProjectItemIndex(req)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		proxyPath := strings.TrimSpace(gomcp.ParseString(req, "proxy_path", ""))
		if proxyPath == "" {
			return gomcp.NewToolResultError("parameter 'proxy_path' is required"), nil
		}
		result, err := orch.AttachProxy(ctx, projectItemIndex, proxyPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_has_proxy",
		gomcp.WithDescription("Read and cross-check whether a root-bin project item has proxy media attached, including its canonical proxy path when present."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Min(0), gomcp.Description("Zero-based index of a proxy-capable project item in the root bin")),
	), wsH(orch, logger, "has_proxy", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		projectItemIndex, err := proxyProjectItemIndex(req)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		result, err := orch.HasProxy(ctx, projectItemIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_proxy_path",
		gomcp.WithDescription("Get the canonical proxy path for a root-bin project item and cross-check it against hasProxy state."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Min(0), gomcp.Description("Zero-based index of a proxy-capable project item in the root bin")),
	), wsH(orch, logger, "get_proxy_path", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		projectItemIndex, err := proxyProjectItemIndex(req)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		result, err := orch.GetProxyPath(ctx, projectItemIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_toggle_proxies",
		gomcp.WithDescription("Set proxy playback mode on or off using Premiere's query/set APIs and return verified state readback."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable proxy playback, false to disable")),
	), wsH(orch, logger, "toggle_proxies", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ToggleProxies(ctx, gomcp.ParseBoolean(req, "enabled", false))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_detach_proxy",
		gomcp.WithDescription("Detach proxy media from a root-bin project item with Premiere's detach API and verify that proxy state and path are cleared."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Min(0), gomcp.Description("Zero-based index of a proxy-capable project item in the root bin")),
	), wsH(orch, logger, "detach_proxy", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		projectItemIndex, err := proxyProjectItemIndex(req)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		result, err := orch.DetachProxy(ctx, projectItemIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Workspace (11-13)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_get_workspaces",
		gomcp.WithDescription("List all available workspace layouts in Premiere Pro."),
	), wsH(orch, logger, "get_workspaces", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetWorkspaces(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_workspace",
		gomcp.WithDescription("Switch to a named workspace layout (e.g. Editing, Color, Audio, Effects)."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name of the workspace to activate")),
	), wsH(orch, logger, "set_workspace", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SetWorkspace(ctx, name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_save_workspace",
		gomcp.WithDescription("Save the current panel layout as a named workspace."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name to save the workspace as")),
	), wsH(orch, logger, "save_workspace", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SaveWorkspace(ctx, name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Undo / Redo (14-15)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_undo",
		gomcp.WithDescription("Undo the last action performed in Premiere Pro."),
	), wsH(orch, logger, "undo", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.Undo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_redo",
		gomcp.WithDescription("Redo the last undone action in Premiere Pro."),
	), wsH(orch, logger, "redo", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.Redo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Project Panel (16-17)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_sort_project_panel",
		gomcp.WithDescription("Sort items in the project panel by a specified field."),
		gomcp.WithString("field", gomcp.Required(), gomcp.Description("Field to sort by"), gomcp.Enum("name", "label", "type", "frameRate", "duration", "videoInfo", "audioInfo", "dateCreated", "dateModified", "filePath")),
		gomcp.WithBoolean("ascending", gomcp.Description("Sort ascending (true) or descending (false). Default: true")),
	), wsH(orch, logger, "sort_project_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SortProjectPanel(ctx, gomcp.ParseString(req, "field", "name"), gomcp.ParseBoolean(req, "ascending", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_search_project_panel",
		gomcp.WithDescription("Search for items in the project panel by name."),
		gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query string")),
	), wsH(orch, logger, "search_project_panel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := gomcp.ParseString(req, "query", "")
		if query == "" {
			return gomcp.NewToolResultError("parameter 'query' is required"), nil
		}
		result, err := orch.SearchProjectPanel(ctx, query)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Source Monitor (18-20)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_open_in_source_monitor",
		gomcp.WithDescription("Open a project item in the source monitor for preview."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in root bin")),
	), wsH(orch, logger, "open_in_source_monitor", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.OpenInSourceMonitor(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_source_monitor_position",
		gomcp.WithDescription("Get the current playhead position in the source monitor."),
	), wsH(orch, logger, "get_source_monitor_position", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSourceMonitorPosition(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_source_monitor_position",
		gomcp.WithDescription("Set the playhead position in the source monitor."),
		gomcp.WithNumber("seconds", gomcp.Required(), gomcp.Description("Playhead position in seconds")),
	), wsH(orch, logger, "set_source_monitor_position", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetSourceMonitorPosition(ctx, gomcp.ParseFloat64(req, "seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Preferences (21-23)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_get_auto_save_settings",
		gomcp.WithDescription("Get the current auto-save configuration including interval and enabled state."),
	), wsH(orch, logger, "get_auto_save_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAutoSaveSettings(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_auto_save_interval",
		gomcp.WithDescription("Set the auto-save interval in minutes (1-99)."),
		gomcp.WithNumber("minutes", gomcp.Required(), gomcp.Description("Auto-save interval in minutes (1-99)")),
	), wsH(orch, logger, "set_auto_save_interval", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAutoSaveInterval(ctx, gomcp.ParseInt(req, "minutes", 15))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_memory_settings",
		gomcp.WithDescription("Get memory and performance settings including GPU renderer info."),
	), wsH(orch, logger, "get_memory_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMemorySettings(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Media Cache (24-25)
	// -------------------------------------------------------------------

	s.AddTool(gomcp.NewTool("premiere_clear_media_cache",
		gomcp.WithDescription("Clear all media cache files to free disk space. Premiere will regenerate caches as needed."),
	), wsH(orch, logger, "clear_media_cache", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearMediaCache(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_media_cache_path",
		gomcp.WithDescription("Get the file system path where Premiere Pro stores media cache files."),
	), wsH(orch, logger, "get_media_cache_path", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMediaCachePath(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

func proxyProjectItemIndex(req gomcp.CallToolRequest) (int, error) {
	arguments, err := objectArguments(req.Params.Arguments)
	if err != nil {
		return 0, fmt.Errorf("invalid project_item_index: %w", err)
	}
	raw, ok := arguments["project_item_index"]
	if !ok {
		return 0, fmt.Errorf("parameter 'project_item_index' is required")
	}
	number, ok := raw.(float64)
	maxInt := int(^uint(0) >> 1)
	if !ok || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || math.Trunc(number) != number || number >= float64(maxInt) {
		return 0, fmt.Errorf("parameter 'project_item_index' must be a non-negative integer")
	}
	return int(number), nil
}

// wsH is a small wrapper that logs the tool name before delegating to the handler.
func wsH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// extractIntSlice pulls a []int from the request arguments map.
func extractIntSlice(req gomcp.CallToolRequest, key string) ([]int, error) {
	raw := gomcp.ParseArgument(req, key, nil)
	if raw == nil {
		return nil, fmt.Errorf("key %q not found", key)
	}

	switch v := raw.(type) {
	case []int:
		return v, nil
	case []float64:
		out := make([]int, 0, len(v))
		for _, f := range v {
			out = append(out, int(f))
		}
		return out, nil
	case []any:
		out := make([]int, 0, len(v))
		for i, item := range v {
			switch n := item.(type) {
			case float64:
				out = append(out, int(n))
			case int:
				out = append(out, n)
			default:
				return nil, fmt.Errorf("element %d of %q is not a number", i, key)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("key %q is not an array (got %T)", key, raw)
	}
}
