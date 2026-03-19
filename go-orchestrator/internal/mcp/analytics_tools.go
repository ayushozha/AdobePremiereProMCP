package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// analyticsH is a small handler wrapper for analytics tools.
func analyticsH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerAnalyticsTools registers all 30 project analytics, reporting, and
// data extraction MCP tools.
func registerAnalyticsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Project Analytics (1-7)
	// -----------------------------------------------------------------------

	// 1. premiere_get_project_summary
	s.AddTool(gomcp.NewTool("premiere_get_project_summary",
		gomcp.WithDescription("Get a comprehensive project summary including sequence count, bin count, clip count, total duration, and disk usage."),
	), analyticsH(logger, "get_project_summary", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectSummary(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_get_media_type_breakdown
	s.AddTool(gomcp.NewTool("premiere_get_media_type_breakdown",
		gomcp.WithDescription("Get a breakdown of project media by type (video, audio, image, graphics) with counts and total duration per type."),
	), analyticsH(logger, "get_media_type_breakdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMediaTypeBreakdown(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_codec_breakdown
	s.AddTool(gomcp.NewTool("premiere_get_codec_breakdown",
		gomcp.WithDescription("Get a breakdown of all codecs used across project media with counts per codec."),
	), analyticsH(logger, "get_codec_breakdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetCodecBreakdown(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_get_resolution_breakdown
	s.AddTool(gomcp.NewTool("premiere_get_resolution_breakdown",
		gomcp.WithDescription("Get a breakdown of all resolutions used across project media with counts per resolution."),
	), analyticsH(logger, "get_resolution_breakdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetResolutionBreakdown(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_get_frame_rate_breakdown
	s.AddTool(gomcp.NewTool("premiere_get_frame_rate_breakdown",
		gomcp.WithDescription("Get a breakdown of all frame rates used across project media with counts per frame rate."),
	), analyticsH(logger, "get_frame_rate_breakdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFrameRateBreakdown(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_get_duration_distribution
	s.AddTool(gomcp.NewTool("premiere_get_duration_distribution",
		gomcp.WithDescription("Get the distribution of clip durations across the project as histogram data (buckets with counts)."),
	), analyticsH(logger, "get_duration_distribution", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDurationDistribution(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_get_color_space_breakdown
	s.AddTool(gomcp.NewTool("premiere_get_color_space_breakdown",
		gomcp.WithDescription("Get a breakdown of all color spaces used across project media with counts per color space."),
	), analyticsH(logger, "get_color_space_breakdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetColorSpaceBreakdown(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Sequence Analytics (8-14)
	// -----------------------------------------------------------------------

	// 8. premiere_get_sequence_summary
	s.AddTool(gomcp.NewTool("premiere_get_sequence_summary",
		gomcp.WithDescription("Get a comprehensive summary of a specific sequence including track count, clip count, duration, and effects used."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_sequence_summary", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceSummary(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_get_effects_usage_report
	s.AddTool(gomcp.NewTool("premiere_get_effects_usage_report",
		gomcp.WithDescription("List all effects used in a sequence with usage counts per effect."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_effects_usage_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEffectsUsageReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_get_transitions_usage_report
	s.AddTool(gomcp.NewTool("premiere_get_transitions_usage_report",
		gomcp.WithDescription("List all transitions used in a sequence with usage counts per transition type."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_transitions_usage_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTransitionsUsageReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_get_track_utilization_report
	s.AddTool(gomcp.NewTool("premiere_get_track_utilization_report",
		gomcp.WithDescription("Get track utilization percentages for each track in a sequence (used time vs total duration)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_track_utilization_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTrackUtilizationReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_get_edit_point_density
	s.AddTool(gomcp.NewTool("premiere_get_edit_point_density",
		gomcp.WithDescription("Calculate the edit point density (edit points per minute) for a sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_edit_point_density", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEditPointDensity(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_get_pacing_report
	s.AddTool(gomcp.NewTool("premiere_get_pacing_report",
		gomcp.WithDescription("Get a pacing report for a sequence including average clip duration and cuts per minute."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_pacing_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPacingReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_get_audio_levels_report
	s.AddTool(gomcp.NewTool("premiere_get_audio_levels_report",
		gomcp.WithDescription("Get audio level statistics across all audio tracks in a sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_audio_levels_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAudioLevelsReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Timeline Reports (15-19)
	// -----------------------------------------------------------------------

	// 15. premiere_get_clip_source_report
	s.AddTool(gomcp.NewTool("premiere_get_clip_source_report",
		gomcp.WithDescription("Report which source files are used in a sequence and where each source appears on the timeline."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_clip_source_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipSourceReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_get_timeline_structure_report
	s.AddTool(gomcp.NewTool("premiere_get_timeline_structure_report",
		gomcp.WithDescription("Get the structural layout of a timeline including sections, acts, and major divisions."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_timeline_structure_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTimelineStructureReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_get_gap_analysis_report
	s.AddTool(gomcp.NewTool("premiere_get_gap_analysis_report",
		gomcp.WithDescription("Get a detailed gap analysis report for a sequence, listing all gaps with their positions and durations."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_gap_analysis_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetGapAnalysisReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_get_duplicate_clips_report
	s.AddTool(gomcp.NewTool("premiere_get_duplicate_clips_report",
		gomcp.WithDescription("Find duplicate clips in a sequence timeline (same source file used multiple times)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_duplicate_clips_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDuplicateClipsReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_get_unused_tracks_report
	s.AddTool(gomcp.NewTool("premiere_get_unused_tracks_report",
		gomcp.WithDescription("Identify empty or unused tracks in a sequence that contain no clips."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
	), analyticsH(logger, "get_unused_tracks_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetUnusedTracksReport(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Export Reports (20-24)
	// -----------------------------------------------------------------------

	// 20. premiere_export_project_report
	s.AddTool(gomcp.NewTool("premiere_export_project_report",
		gomcp.WithDescription("Export a comprehensive project report to a file. Supports JSON and HTML formats."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output report file")),
		gomcp.WithString("format", gomcp.Description("Report format: json or html (default: json)"), gomcp.Enum("json", "html")),
	), analyticsH(logger, "export_project_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExportProjectReport(ctx,
			gomcp.ParseString(req, "output_path", ""),
			gomcp.ParseString(req, "format", "json"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_export_timeline_as_text
	s.AddTool(gomcp.NewTool("premiere_export_timeline_as_text",
		gomcp.WithDescription("Export a sequence timeline as human-readable text to a file."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output text file")),
	), analyticsH(logger, "export_timeline_as_text", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExportTimelineAsText(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "output_path", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_export_clip_list
	s.AddTool(gomcp.NewTool("premiere_export_clip_list",
		gomcp.WithDescription("Export the clip list from a sequence to a file. Supports CSV and JSON formats."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("format", gomcp.Description("Output format: csv or json (default: csv)"), gomcp.Enum("csv", "json")),
	), analyticsH(logger, "export_clip_list", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExportClipList(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "output_path", ""),
			gomcp.ParseString(req, "format", "csv"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_export_effects_list
	s.AddTool(gomcp.NewTool("premiere_export_effects_list",
		gomcp.WithDescription("Export the list of all effects used in a sequence to a JSON file."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0 for active sequence)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output JSON file")),
	), analyticsH(logger, "export_effects_list", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExportEffectsList(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "output_path", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_export_media_list
	s.AddTool(gomcp.NewTool("premiere_export_media_list",
		gomcp.WithDescription("Export a list of all media files in the project to a file. Supports CSV and JSON formats."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("format", gomcp.Description("Output format: csv or json (default: csv)"), gomcp.Enum("csv", "json")),
	), analyticsH(logger, "export_media_list", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExportMediaList(ctx,
			gomcp.ParseString(req, "output_path", ""),
			gomcp.ParseString(req, "format", "csv"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Comparison (25-26)
	// -----------------------------------------------------------------------

	// 25. premiere_compare_sequences
	s.AddTool(gomcp.NewTool("premiere_compare_sequences",
		gomcp.WithDescription("Compare two sequences side by side: clip counts, duration, effects, track usage, and structural differences."),
		gomcp.WithNumber("sequence_index_1", gomcp.Required(), gomcp.Description("Zero-based index of the first sequence")),
		gomcp.WithNumber("sequence_index_2", gomcp.Required(), gomcp.Description("Zero-based index of the second sequence")),
	), analyticsH(logger, "compare_sequences", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CompareSequences(ctx,
			gomcp.ParseInt(req, "sequence_index_1", 0),
			gomcp.ParseInt(req, "sequence_index_2", 1))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_compare_clips
	s.AddTool(gomcp.NewTool("premiere_compare_clips",
		gomcp.WithDescription("Compare two clips including their properties, effects, and timing. Clips are identified by track type, track index, and clip index."),
		gomcp.WithString("clip1_track_type", gomcp.Required(), gomcp.Description("Track type for clip 1: video or audio"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("clip1_track_index", gomcp.Required(), gomcp.Description("Zero-based track index for clip 1")),
		gomcp.WithNumber("clip1_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index for clip 1")),
		gomcp.WithString("clip2_track_type", gomcp.Required(), gomcp.Description("Track type for clip 2: video or audio"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("clip2_track_index", gomcp.Required(), gomcp.Description("Zero-based track index for clip 2")),
		gomcp.WithNumber("clip2_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index for clip 2")),
	), analyticsH(logger, "compare_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CompareClips(ctx,
			gomcp.ParseString(req, "clip1_track_type", "video"),
			gomcp.ParseInt(req, "clip1_track_index", 0),
			gomcp.ParseInt(req, "clip1_clip_index", 0),
			gomcp.ParseString(req, "clip2_track_type", "video"),
			gomcp.ParseInt(req, "clip2_track_index", 0),
			gomcp.ParseInt(req, "clip2_clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Usage Statistics (27-30)
	// -----------------------------------------------------------------------

	// 27. premiere_get_editing_session_stats
	s.AddTool(gomcp.NewTool("premiere_get_editing_session_stats",
		gomcp.WithDescription("Get statistics for the current editing session including uptime, actions performed, and resource usage."),
	), analyticsH(logger, "get_editing_session_stats", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEditingSessionStats(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_get_project_age_info
	s.AddTool(gomcp.NewTool("premiere_get_project_age_info",
		gomcp.WithDescription("Get project age information including creation date, last modified date, and estimated edit sessions."),
	), analyticsH(logger, "get_project_age_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectAgeInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_get_storage_report
	s.AddTool(gomcp.NewTool("premiere_get_storage_report",
		gomcp.WithDescription("Get a storage usage report covering source media, renders, previews, and cache sizes."),
	), analyticsH(logger, "get_storage_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetStorageReport(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_get_performance_report
	s.AddTool(gomcp.NewTool("premiere_get_performance_report",
		gomcp.WithDescription("Get a performance report including estimated render times, playback quality, dropped frames, and system resource status."),
	), analyticsH(logger, "get_performance_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPerformanceReport2(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
