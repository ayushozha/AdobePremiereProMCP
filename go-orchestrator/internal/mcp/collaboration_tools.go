package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// colH is a small handler wrapper for collaboration tools.
func colH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerCollaborationTools registers all 30 collaboration, review, project
// sharing, version control, EDL/XML interchange, change tracking, and delivery
// checklist MCP tools.
func registerCollaborationTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Review & Collaboration (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_add_review_comment
	s.AddTool(gomcp.NewTool("premiere_add_review_comment",
		gomcp.WithDescription("Add a review comment as a sequence marker with author metadata at the specified time."),
		gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time position in seconds where the comment should be placed")),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Review comment text")),
		gomcp.WithString("author", gomcp.Description("Author name for the comment (default: 'Reviewer')")),
	), colH(orch, logger, "add_review_comment", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.AddReviewComment(ctx,
			gomcp.ParseFloat64(req, "time", 0),
			text,
			gomcp.ParseString(req, "author", "Reviewer"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_get_review_comments
	s.AddTool(gomcp.NewTool("premiere_get_review_comments",
		gomcp.WithDescription("Get all review comments (markers with comment data) from the active sequence."),
	), colH(orch, logger, "get_review_comments", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetReviewComments(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_resolve_review_comment
	s.AddTool(gomcp.NewTool("premiere_resolve_review_comment",
		gomcp.WithDescription("Mark a review comment as resolved by its marker index."),
		gomcp.WithNumber("marker_index", gomcp.Required(), gomcp.Description("Zero-based index of the marker to resolve")),
	), colH(orch, logger, "resolve_review_comment", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		markerIndex := gomcp.ParseInt(req, "marker_index", -1)
		if markerIndex < 0 {
			return gomcp.NewToolResultError("parameter 'marker_index' is required"), nil
		}
		result, err := orch.ResolveReviewComment(ctx, markerIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_get_unresolved_comments
	s.AddTool(gomcp.NewTool("premiere_get_unresolved_comments",
		gomcp.WithDescription("Get all unresolved review comments from the active sequence."),
	), colH(orch, logger, "get_unresolved_comments", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetUnresolvedComments(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_export_review_report
	s.AddTool(gomcp.NewTool("premiere_export_review_report",
		gomcp.WithDescription("Export a review report containing all comments and their statuses."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output report file")),
		gomcp.WithString("format", gomcp.Description("Report format: json, csv, or html (default: json)"),
			gomcp.Enum("json", "csv", "html")),
	), colH(orch, logger, "export_review_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportReviewReport(ctx, outputPath, gomcp.ParseString(req, "format", "json"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Version Control (6-9)
	// -----------------------------------------------------------------------

	// 6. premiere_get_project_version_history
	s.AddTool(gomcp.NewTool("premiere_get_project_version_history",
		gomcp.WithDescription("Get auto-save versions of the current project with timestamps and file paths."),
	), colH(orch, logger, "get_project_version_history", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectVersionHistory(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_revert_to_version
	s.AddTool(gomcp.NewTool("premiere_revert_to_version",
		gomcp.WithDescription("Open a specific auto-save version of the project, reverting to that point in time."),
		gomcp.WithString("version_path", gomcp.Required(), gomcp.Description("Absolute path to the auto-save version file (.prproj)")),
	), colH(orch, logger, "revert_to_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		versionPath := gomcp.ParseString(req, "version_path", "")
		if versionPath == "" {
			return gomcp.NewToolResultError("parameter 'version_path' is required"), nil
		}
		result, err := orch.RevertToVersion(ctx, versionPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_create_snapshot
	s.AddTool(gomcp.NewTool("premiere_create_snapshot",
		gomcp.WithDescription("Save the current project state as a named snapshot for later comparison or revert."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the snapshot")),
		gomcp.WithString("description", gomcp.Description("Description of what this snapshot captures")),
	), colH(orch, logger, "create_snapshot", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.CreateSnapshot(ctx, name, gomcp.ParseString(req, "description", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_compare_snapshots
	s.AddTool(gomcp.NewTool("premiere_compare_snapshots",
		gomcp.WithDescription("Compare two project snapshots and return a basic diff of changes between them."),
		gomcp.WithString("snapshot1", gomcp.Required(), gomcp.Description("Absolute path to the first snapshot file")),
		gomcp.WithString("snapshot2", gomcp.Required(), gomcp.Description("Absolute path to the second snapshot file")),
	), colH(orch, logger, "compare_snapshots", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		s1 := gomcp.ParseString(req, "snapshot1", "")
		if s1 == "" {
			return gomcp.NewToolResultError("parameter 'snapshot1' is required"), nil
		}
		s2 := gomcp.ParseString(req, "snapshot2", "")
		if s2 == "" {
			return gomcp.NewToolResultError("parameter 'snapshot2' is required"), nil
		}
		result, err := orch.CompareSnapshots(ctx, s1, s2)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// EDL/XML Interchange (10-15)
	// -----------------------------------------------------------------------

	// 10. premiere_import_edl
	s.AddTool(gomcp.NewTool("premiere_import_edl",
		gomcp.WithDescription("Import an EDL (Edit Decision List) file into the project."),
		gomcp.WithString("edl_path", gomcp.Required(), gomcp.Description("Absolute path to the EDL file")),
	), colH(orch, logger, "import_edl", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		edlPath := gomcp.ParseString(req, "edl_path", "")
		if edlPath == "" {
			return gomcp.NewToolResultError("parameter 'edl_path' is required"), nil
		}
		result, err := orch.ImportEDL(ctx, edlPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_import_aaf
	s.AddTool(gomcp.NewTool("premiere_import_aaf",
		gomcp.WithDescription("Import an AAF (Advanced Authoring Format) file into the project."),
		gomcp.WithString("aaf_path", gomcp.Required(), gomcp.Description("Absolute path to the AAF file")),
	), colH(orch, logger, "import_aaf", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		aafPath := gomcp.ParseString(req, "aaf_path", "")
		if aafPath == "" {
			return gomcp.NewToolResultError("parameter 'aaf_path' is required"), nil
		}
		result, err := orch.ImportAAF(ctx, aafPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_import_fcpxml
	s.AddTool(gomcp.NewTool("premiere_import_fcpxml",
		gomcp.WithDescription("Import a Final Cut Pro XML (.fcpxml) file into the project."),
		gomcp.WithString("xml_path", gomcp.Required(), gomcp.Description("Absolute path to the FCP XML file")),
	), colH(orch, logger, "import_fcpxml", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		xmlPath := gomcp.ParseString(req, "xml_path", "")
		if xmlPath == "" {
			return gomcp.NewToolResultError("parameter 'xml_path' is required"), nil
		}
		result, err := orch.ImportFCPXML(ctx, xmlPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_import_xml_timeline
	s.AddTool(gomcp.NewTool("premiere_import_xml_timeline",
		gomcp.WithDescription("Import a Premiere Pro XML timeline file into the project."),
		gomcp.WithString("xml_path", gomcp.Required(), gomcp.Description("Absolute path to the Premiere XML file")),
	), colH(orch, logger, "import_xml_timeline", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		xmlPath := gomcp.ParseString(req, "xml_path", "")
		if xmlPath == "" {
			return gomcp.NewToolResultError("parameter 'xml_path' is required"), nil
		}
		result, err := orch.ImportXMLTimeline(ctx, xmlPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_export_edl_file
	s.AddTool(gomcp.NewTool("premiere_export_edl_file",
		gomcp.WithDescription("Export the active sequence as an EDL file (CMX 3600 format)."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output EDL file")),
		gomcp.WithString("format", gomcp.Description("EDL format: cmx3600 or cmx3600_file32 (default: cmx3600)"),
			gomcp.Enum("cmx3600", "cmx3600_file32")),
	), colH(orch, logger, "export_edl_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportEDLFile(ctx, outputPath, gomcp.ParseString(req, "format", "cmx3600"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_export_project_snapshot
	s.AddTool(gomcp.NewTool("premiere_export_project_snapshot",
		gomcp.WithDescription("Export the current project as a portable snapshot (project file copy with metadata)."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output snapshot file")),
	), colH(orch, logger, "export_project_snapshot", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportProjectSnapshot(ctx, outputPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Collaboration Metadata (16-20)
	// -----------------------------------------------------------------------

	// 16. premiere_set_editorial_note
	s.AddTool(gomcp.NewTool("premiere_set_editorial_note",
		gomcp.WithDescription("Set an editorial note on a specific clip in the active sequence."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("note", gomcp.Required(), gomcp.Description("Editorial note text")),
	), colH(orch, logger, "set_editorial_note", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		note := gomcp.ParseString(req, "note", "")
		if note == "" {
			return gomcp.NewToolResultError("parameter 'note' is required"), nil
		}
		result, err := orch.SetEditorialNote(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			note)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_get_editorial_notes
	s.AddTool(gomcp.NewTool("premiere_get_editorial_notes",
		gomcp.WithDescription("Get all editorial notes from clips in the active sequence."),
	), colH(orch, logger, "get_editorial_notes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEditorialNotes(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_clear_editorial_notes
	s.AddTool(gomcp.NewTool("premiere_clear_editorial_notes",
		gomcp.WithDescription("Clear all editorial notes from clips in the active sequence."),
	), colH(orch, logger, "clear_editorial_notes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearEditorialNotes(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_tag_clip_for_review
	s.AddTool(gomcp.NewTool("premiere_tag_clip_for_review",
		gomcp.WithDescription("Tag a clip with a review status (approved, needs-changes, or rejected)."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("review_type", gomcp.Required(), gomcp.Description("Review status tag"),
			gomcp.Enum("approved", "needs-changes", "rejected")),
	), colH(orch, logger, "tag_clip_for_review", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		reviewType := gomcp.ParseString(req, "review_type", "")
		if reviewType == "" {
			return gomcp.NewToolResultError("parameter 'review_type' is required"), nil
		}
		result, err := orch.TagClipForReview(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			reviewType)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_get_clip_review_status
	s.AddTool(gomcp.NewTool("premiere_get_clip_review_status",
		gomcp.WithDescription("Get the review status tag for a specific clip."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), colH(orch, logger, "get_clip_review_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipReviewStatus(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Change Tracking (21-23)
	// -----------------------------------------------------------------------

	// 21. premiere_get_sequence_change_log
	s.AddTool(gomcp.NewTool("premiere_get_sequence_change_log",
		gomcp.WithDescription("Get recent changes to the active sequence including clips added, removed, and moved."),
	), colH(orch, logger, "get_sequence_change_log", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSequenceChangeLog(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_get_project_activity
	s.AddTool(gomcp.NewTool("premiere_get_project_activity",
		gomcp.WithDescription("Get a recent project activity summary including file changes and edit operations."),
	), colH(orch, logger, "get_project_activity", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProjectActivity(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_get_last_modified_clips
	s.AddTool(gomcp.NewTool("premiere_get_last_modified_clips",
		gomcp.WithDescription("Get the N most recently modified clips in the active sequence."),
		gomcp.WithNumber("count", gomcp.Description("Number of recently modified clips to return (default: 10)")),
	), colH(orch, logger, "get_last_modified_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetLastModifiedClips(ctx, gomcp.ParseInt(req, "count", 10))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Delivery Checklist (24-30)
	// -----------------------------------------------------------------------

	// 24. premiere_check_audio_levels
	s.AddTool(gomcp.NewTool("premiere_check_audio_levels",
		gomcp.WithDescription("Check that all audio levels in the active sequence meet the target LUFS specification."),
		gomcp.WithNumber("target_lufs", gomcp.Description("Target loudness in LUFS (default: -24.0)")),
		gomcp.WithNumber("tolerance", gomcp.Description("Acceptable tolerance in LU (default: 2.0)")),
	), colH(orch, logger, "check_audio_levels", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckAudioLevels(ctx,
			gomcp.ParseFloat64(req, "target_lufs", -24.0),
			gomcp.ParseFloat64(req, "tolerance", 2.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_check_frame_rate
	s.AddTool(gomcp.NewTool("premiere_check_frame_rate",
		gomcp.WithDescription("Verify that the active sequence frame rate matches the target FPS."),
		gomcp.WithNumber("target_fps", gomcp.Required(), gomcp.Description("Expected frame rate in frames per second (e.g. 23.976, 24, 29.97, 30, 60)")),
	), colH(orch, logger, "check_frame_rate", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		targetFPS := gomcp.ParseFloat64(req, "target_fps", 0)
		if targetFPS <= 0 {
			return gomcp.NewToolResultError("parameter 'target_fps' must be a positive number"), nil
		}
		result, err := orch.CheckFrameRate(ctx, targetFPS)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_check_resolution
	s.AddTool(gomcp.NewTool("premiere_check_resolution",
		gomcp.WithDescription("Verify that the active sequence resolution matches the target width and height."),
		gomcp.WithNumber("target_width", gomcp.Required(), gomcp.Description("Expected frame width in pixels (e.g. 1920, 3840)")),
		gomcp.WithNumber("target_height", gomcp.Required(), gomcp.Description("Expected frame height in pixels (e.g. 1080, 2160)")),
	), colH(orch, logger, "check_resolution", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		w := gomcp.ParseInt(req, "target_width", 0)
		h := gomcp.ParseInt(req, "target_height", 0)
		if w <= 0 || h <= 0 {
			return gomcp.NewToolResultError("parameters 'target_width' and 'target_height' must be positive"), nil
		}
		result, err := orch.CheckResolution(ctx, w, h)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_check_duration
	s.AddTool(gomcp.NewTool("premiere_check_duration",
		gomcp.WithDescription("Check that the active sequence duration falls within the specified range."),
		gomcp.WithNumber("min_seconds", gomcp.Required(), gomcp.Description("Minimum acceptable duration in seconds")),
		gomcp.WithNumber("max_seconds", gomcp.Required(), gomcp.Description("Maximum acceptable duration in seconds")),
	), colH(orch, logger, "check_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckDuration(ctx,
			gomcp.ParseFloat64(req, "min_seconds", 0),
			gomcp.ParseFloat64(req, "max_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_generate_delivery_report
	s.AddTool(gomcp.NewTool("premiere_generate_delivery_report",
		gomcp.WithDescription("Generate a full delivery compliance report checking resolution, frame rate, audio levels, duration, and more."),
		gomcp.WithString("specs", gomcp.Required(), gomcp.Description("JSON string of delivery specifications (e.g. target_fps, target_width, target_height, target_lufs, min_duration, max_duration)")),
	), colH(orch, logger, "generate_delivery_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		specs := gomcp.ParseString(req, "specs", "")
		if specs == "" {
			return gomcp.NewToolResultError("parameter 'specs' is required"), nil
		}
		result, err := orch.GenerateDeliveryReport(ctx, specs)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_check_for_black_frames
	s.AddTool(gomcp.NewTool("premiere_check_for_black_frames",
		gomcp.WithDescription("Detect sequences of black frames in the active sequence that exceed a threshold."),
		gomcp.WithNumber("threshold_frames", gomcp.Description("Minimum consecutive black frames to flag (default: 2)")),
	), colH(orch, logger, "check_for_black_frames", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckForBlackFrames(ctx, gomcp.ParseInt(req, "threshold_frames", 2))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_check_for_flash_content
	s.AddTool(gomcp.NewTool("premiere_check_for_flash_content",
		gomcp.WithDescription("Detect rapid luminance changes that could trigger photosensitive epilepsy (PSE compliance check)."),
		gomcp.WithNumber("threshold", gomcp.Description("Luminance change threshold for detection (default: 0.5, range 0.0-1.0)")),
	), colH(orch, logger, "check_for_flash_content", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckForFlashContent(ctx, gomcp.ParseFloat64(req, "threshold", 0.5))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
