package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// diagH is a small handler wrapper for diagnostics tools.
func diagH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerDiagnosticsTools registers all 30 performance monitoring, diagnostics,
// and system health MCP tools.
func registerDiagnosticsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Performance Monitoring (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_get_performance_metrics
	s.AddTool(gomcp.NewTool("premiere_get_performance_metrics",
		gomcp.WithDescription("Get system performance metrics including CPU, memory, and GPU usage from the Premiere Pro host."),
	), diagH(orch, logger, "get_performance_metrics", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPerformanceMetrics(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_get_project_memory_usage
	s.AddTool(gomcp.NewTool("premiere_get_project_memory_usage",
		gomcp.WithDescription("Get memory usage information for the currently open Premiere Pro project."),
	), diagH(orch, logger, "get_project_memory_usage", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectMemoryUsage(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_disk_space
	s.AddTool(gomcp.NewTool("premiere_get_disk_space",
		gomcp.WithDescription("Get available disk space for a given drive or directory path."),
		gomcp.WithString("drive_path", gomcp.Description("Path to the drive or directory to check (default: root '/')")),
	), diagH(orch, logger, "get_disk_space", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDiskSpace(ctx, gomcp.ParseString(req, "drive_path", "/"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_get_open_project_count
	s.AddTool(gomcp.NewTool("premiere_get_open_project_count",
		gomcp.WithDescription("Get the number of currently open projects in Premiere Pro."),
	), diagH(orch, logger, "get_open_project_count", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetOpenProjectCount(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_get_loaded_plugins
	s.AddTool(gomcp.NewTool("premiere_get_loaded_plugins",
		gomcp.WithDescription("List all loaded plugins and extensions in the current Premiere Pro session."),
	), diagH(orch, logger, "get_loaded_plugins", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetLoadedPlugins(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Timeline Performance (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_get_dropped_frame_count
	s.AddTool(gomcp.NewTool("premiere_get_dropped_frame_count",
		gomcp.WithDescription("Get the number of dropped frames during playback of the active sequence."),
	), diagH(orch, logger, "get_dropped_frame_count", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDroppedFrameCount(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_reset_dropped_frame_count
	s.AddTool(gomcp.NewTool("premiere_reset_dropped_frame_count",
		gomcp.WithDescription("Reset the dropped frame counter to zero for the active sequence."),
	), diagH(orch, logger, "reset_dropped_frame_count", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ResetDroppedFrameCount(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_get_timeline_render_status
	s.AddTool(gomcp.NewTool("premiere_get_timeline_render_status",
		gomcp.WithDescription("Get the render status (red/yellow/green bar) for each segment of a sequence timeline."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), diagH(orch, logger, "get_timeline_render_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTimelineRenderStatus(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_get_estimated_render_time
	s.AddTool(gomcp.NewTool("premiere_get_estimated_render_time",
		gomcp.WithDescription("Get an estimated render time for a sequence based on complexity heuristics."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), diagH(orch, logger, "get_estimated_render_time", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEstimatedRenderTime2(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_get_sequence_complexity
	s.AddTool(gomcp.NewTool("premiere_get_sequence_complexity",
		gomcp.WithDescription("Rate the complexity of a sequence (effects count, track count, clip count) on a 0-100 scale."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), diagH(orch, logger, "get_sequence_complexity", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceComplexity(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Diagnostics (11-15)
	// -----------------------------------------------------------------------

	// 11. premiere_get_premiere_version
	s.AddTool(gomcp.NewTool("premiere_get_premiere_version",
		gomcp.WithDescription("Get detailed Premiere Pro version information including version number, build, architecture, and locale."),
	), diagH(orch, logger, "get_premiere_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPremiereVersion(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_get_installed_plugins
	s.AddTool(gomcp.NewTool("premiere_get_installed_plugins",
		gomcp.WithDescription("List all installed plugins in Premiere Pro with names, types, and file paths."),
	), diagH(orch, logger, "get_installed_plugins", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetInstalledPlugins2(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_get_installed_effects
	s.AddTool(gomcp.NewTool("premiere_get_installed_effects",
		gomcp.WithDescription("List all available video and audio effects installed in Premiere Pro."),
	), diagH(orch, logger, "get_installed_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetInstalledEffects2(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_get_installed_transitions
	s.AddTool(gomcp.NewTool("premiere_get_installed_transitions",
		gomcp.WithDescription("List all available video and audio transitions installed in Premiere Pro."),
	), diagH(orch, logger, "get_installed_transitions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetInstalledTransitions2(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_check_project_integrity
	s.AddTool(gomcp.NewTool("premiere_check_project_integrity",
		gomcp.WithDescription("Run a basic project integrity check: detect offline items, missing media, and sequence issues."),
	), diagH(orch, logger, "check_project_integrity", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckProjectIntegrity(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Error Handling (16-19)
	// -----------------------------------------------------------------------

	// 16. premiere_get_last_error
	s.AddTool(gomcp.NewTool("premiere_get_last_error",
		gomcp.WithDescription("Get the last ExtendScript error that occurred in the Premiere Pro bridge."),
	), diagH(orch, logger, "get_last_error", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetLastError(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_clear_errors
	s.AddTool(gomcp.NewTool("premiere_clear_errors",
		gomcp.WithDescription("Clear the ExtendScript error state and reset the error log."),
	), diagH(orch, logger, "clear_errors", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearErrors(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_set_error_logging
	s.AddTool(gomcp.NewTool("premiere_set_error_logging",
		gomcp.WithDescription("Enable or disable error logging to a file for debugging Premiere Pro bridge issues."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("Whether to enable (true) or disable (false) error logging")),
		gomcp.WithString("log_path", gomcp.Description("Absolute path to the log file (default: in-memory logging)")),
	), diagH(orch, logger, "set_error_logging", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetErrorLogging(ctx,
			gomcp.ParseBoolean(req, "enabled", false),
			gomcp.ParseString(req, "log_path", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_get_error_log
	s.AddTool(gomcp.NewTool("premiere_get_error_log",
		gomcp.WithDescription("Get recent error log entries from the Premiere Pro bridge."),
	), diagH(orch, logger, "get_error_log", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetErrorLog(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Debug Tools (20-24)
	// -----------------------------------------------------------------------

	// 20. premiere_enable_debug_mode
	s.AddTool(gomcp.NewTool("premiere_enable_debug_mode",
		gomcp.WithDescription("Toggle debug mode in the Premiere Pro ExtendScript bridge for verbose logging."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("Whether to enable (true) or disable (false) debug mode")),
	), diagH(orch, logger, "enable_debug_mode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.EnableDebugMode(ctx, gomcp.ParseBoolean(req, "enabled", false))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_get_debug_log
	s.AddTool(gomcp.NewTool("premiere_get_debug_log",
		gomcp.WithDescription("Get the debug log contents from the Premiere Pro ExtendScript bridge."),
	), diagH(orch, logger, "get_debug_log", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDebugLog(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_dump_project_state
	s.AddTool(gomcp.NewTool("premiere_dump_project_state",
		gomcp.WithDescription("Dump the full project state for debugging: all items, sequences, paths, and metadata."),
	), diagH(orch, logger, "dump_project_state", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.DumpProjectState(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_dump_sequence_state
	s.AddTool(gomcp.NewTool("premiere_dump_sequence_state",
		gomcp.WithDescription("Dump the full state of a sequence for debugging: all tracks, clips, effects, and timing."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), diagH(orch, logger, "dump_sequence_state", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.DumpSequenceState(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_test_bridge_connection
	s.AddTool(gomcp.NewTool("premiere_test_bridge_connection",
		gomcp.WithDescription("Test the CEP panel bridge connection by running API access, JSON, and file system checks."),
	), diagH(orch, logger, "test_bridge_connection", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.TestBridgeConnection(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Health Checks (25-28)
	// -----------------------------------------------------------------------

	// 25. premiere_health_check
	s.AddTool(gomcp.NewTool("premiere_health_check",
		gomcp.WithDescription("Run a full system health check covering Premiere Pro, the bridge, ExtendScript engine, and project status."),
	), diagH(orch, logger, "health_check", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.HealthCheck(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_get_service_status
	s.AddTool(gomcp.NewTool("premiere_get_service_status",
		gomcp.WithDescription("Get the status of all MCP services: ExtendScript engine, Premiere API, QE DOM, and file system."),
	), diagH(orch, logger, "get_service_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetServiceStatus(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_get_bridge_latency
	s.AddTool(gomcp.NewTool("premiere_get_bridge_latency",
		gomcp.WithDescription("Measure the round-trip latency of the Premiere Pro ExtendScript bridge in milliseconds."),
	), diagH(orch, logger, "get_bridge_latency", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetBridgeLatency(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_get_extendscript_version
	s.AddTool(gomcp.NewTool("premiere_get_extendscript_version",
		gomcp.WithDescription("Get the ExtendScript engine version, build date, locale, and runtime information."),
	), diagH(orch, logger, "get_extendscript_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetExtendScriptVersion(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Cleanup (29-30)
	// -----------------------------------------------------------------------

	// 29. premiere_clean_temp_files
	s.AddTool(gomcp.NewTool("premiere_clean_temp_files",
		gomcp.WithDescription("Clean temporary files created by the MCP bridge and trim oversized in-memory logs."),
	), diagH(orch, logger, "clean_temp_files", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CleanTempFiles(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_optimize_project
	s.AddTool(gomcp.NewTool("premiere_optimize_project",
		gomcp.WithDescription("Run project optimization: consolidate duplicates, audit unused items, clean cache, and save."),
	), diagH(orch, logger, "optimize_project", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.OptimizeProject(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
