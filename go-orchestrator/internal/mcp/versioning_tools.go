package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// verH is a small handler wrapper for versioning tools.
func verH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerVersioningTools registers all 30 timeline diff, sequence versioning,
// A/B comparison, clipboard, history, project backup, and project migration
// MCP tools.
func registerVersioningTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Timeline Diff (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_snapshot_timeline
	s.AddTool(gomcp.NewTool("premiere_snapshot_timeline",
		gomcp.WithDescription("Create a JSON snapshot of a timeline's current state for later comparison."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: active sequence, 0)")),
	), verH(orch, logger, "snapshot_timeline", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SnapshotTimeline(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_compare_timeline_snapshots
	s.AddTool(gomcp.NewTool("premiere_compare_timeline_snapshots",
		gomcp.WithDescription("Diff two timeline snapshots and return added, removed, and modified clips."),
		gomcp.WithString("snapshot1_json", gomcp.Required(), gomcp.Description("JSON string of the first timeline snapshot")),
		gomcp.WithString("snapshot2_json", gomcp.Required(), gomcp.Description("JSON string of the second timeline snapshot")),
	), verH(orch, logger, "compare_timeline_snapshots", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		s1 := gomcp.ParseString(req, "snapshot1_json", "")
		if s1 == "" {
			return gomcp.NewToolResultError("parameter 'snapshot1_json' is required"), nil
		}
		s2 := gomcp.ParseString(req, "snapshot2_json", "")
		if s2 == "" {
			return gomcp.NewToolResultError("parameter 'snapshot2_json' is required"), nil
		}
		result, err := orch.CompareTimelineSnapshots(ctx, s1, s2)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_timeline_changes
	s.AddTool(gomcp.NewTool("premiere_get_timeline_changes",
		gomcp.WithDescription("Get changes to a timeline since a given ISO-8601 timestamp."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("since_timestamp", gomcp.Required(), gomcp.Description("ISO-8601 timestamp to compare against")),
	), verH(orch, logger, "get_timeline_changes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		ts := gomcp.ParseString(req, "since_timestamp", "")
		if ts == "" {
			return gomcp.NewToolResultError("parameter 'since_timestamp' is required"), nil
		}
		result, err := orch.GetTimelineChanges(ctx, gomcp.ParseInt(req, "sequence_index", 0), ts)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_highlight_changed_clips
	s.AddTool(gomcp.NewTool("premiere_highlight_changed_clips",
		gomcp.WithDescription("Select and highlight changed clips in a sequence by their clip IDs."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("changed_clip_ids", gomcp.Required(), gomcp.Description("JSON array of clip IDs to highlight")),
	), verH(orch, logger, "highlight_changed_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		ids := gomcp.ParseString(req, "changed_clip_ids", "")
		if ids == "" {
			return gomcp.NewToolResultError("parameter 'changed_clip_ids' is required"), nil
		}
		result, err := orch.HighlightChangedClips(ctx, gomcp.ParseInt(req, "sequence_index", 0), ids)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_revert_clip_to_snapshot
	s.AddTool(gomcp.NewTool("premiere_revert_clip_to_snapshot",
		gomcp.WithDescription("Revert a single clip to the state captured in a previous snapshot."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("snapshot_json", gomcp.Required(), gomcp.Description("JSON string of the snapshot containing the original clip state")),
	), verH(orch, logger, "revert_clip_to_snapshot", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		snap := gomcp.ParseString(req, "snapshot_json", "")
		if snap == "" {
			return gomcp.NewToolResultError("parameter 'snapshot_json' is required"), nil
		}
		result, err := orch.RevertClipToSnapshot(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			snap)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Sequence Versioning (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_save_sequence_version
	s.AddTool(gomcp.NewTool("premiere_save_sequence_version",
		gomcp.WithDescription("Save a named version of a sequence with optional notes."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("version_name", gomcp.Required(), gomcp.Description("Name for this version")),
		gomcp.WithString("notes", gomcp.Description("Optional notes describing this version")),
	), verH(orch, logger, "save_sequence_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "version_name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'version_name' is required"), nil
		}
		result, err := orch.SaveSequenceVersion(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			name,
			gomcp.ParseString(req, "notes", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_list_sequence_versions
	s.AddTool(gomcp.NewTool("premiere_list_sequence_versions",
		gomcp.WithDescription("List all saved versions for a sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
	), verH(orch, logger, "list_sequence_versions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListSequenceVersions(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_load_sequence_version
	s.AddTool(gomcp.NewTool("premiere_load_sequence_version",
		gomcp.WithDescription("Load a previously saved version of a sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("version_name", gomcp.Required(), gomcp.Description("Name of the version to load")),
	), verH(orch, logger, "load_sequence_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "version_name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'version_name' is required"), nil
		}
		result, err := orch.LoadSequenceVersion(ctx, gomcp.ParseInt(req, "sequence_index", 0), name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_delete_sequence_version
	s.AddTool(gomcp.NewTool("premiere_delete_sequence_version",
		gomcp.WithDescription("Delete a saved version of a sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("version_name", gomcp.Required(), gomcp.Description("Name of the version to delete")),
	), verH(orch, logger, "delete_sequence_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "version_name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'version_name' is required"), nil
		}
		result, err := orch.DeleteSequenceVersion(ctx, gomcp.ParseInt(req, "sequence_index", 0), name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_merge_sequence_versions
	s.AddTool(gomcp.NewTool("premiere_merge_sequence_versions",
		gomcp.WithDescription("Merge two sequence versions using a merge strategy."),
		gomcp.WithString("base_version", gomcp.Required(), gomcp.Description("Name of the base version")),
		gomcp.WithString("overlay_version", gomcp.Required(), gomcp.Description("Name of the overlay version to merge in")),
		gomcp.WithString("strategy", gomcp.Description("Merge strategy: overlay_wins, base_wins, or interleave (default: overlay_wins)"),
			gomcp.Enum("overlay_wins", "base_wins", "interleave")),
	), verH(orch, logger, "merge_sequence_versions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		base := gomcp.ParseString(req, "base_version", "")
		if base == "" {
			return gomcp.NewToolResultError("parameter 'base_version' is required"), nil
		}
		overlay := gomcp.ParseString(req, "overlay_version", "")
		if overlay == "" {
			return gomcp.NewToolResultError("parameter 'overlay_version' is required"), nil
		}
		result, err := orch.MergeSequenceVersions(ctx, base, overlay, gomcp.ParseString(req, "strategy", "overlay_wins"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// A/B Comparison (11-13)
	// -----------------------------------------------------------------------

	// 11. premiere_create_ab_comparison
	s.AddTool(gomcp.NewTool("premiere_create_ab_comparison",
		gomcp.WithDescription("Set up a side-by-side A/B comparison between two sequences."),
		gomcp.WithNumber("seq_index_a", gomcp.Required(), gomcp.Description("Zero-based index of sequence A")),
		gomcp.WithNumber("seq_index_b", gomcp.Required(), gomcp.Description("Zero-based index of sequence B")),
	), verH(orch, logger, "create_ab_comparison", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateABComparison(ctx,
			gomcp.ParseInt(req, "seq_index_a", 0),
			gomcp.ParseInt(req, "seq_index_b", 1))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_switch_ab_view
	s.AddTool(gomcp.NewTool("premiere_switch_ab_view",
		gomcp.WithDescription("Switch the monitor view between A, B, or split view during A/B comparison."),
		gomcp.WithString("view", gomcp.Required(), gomcp.Description("View to switch to"), gomcp.Enum("a", "b", "split")),
	), verH(orch, logger, "switch_ab_view", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		view := gomcp.ParseString(req, "view", "")
		if view == "" {
			return gomcp.NewToolResultError("parameter 'view' is required"), nil
		}
		result, err := orch.SwitchABView(ctx, view)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_get_ab_differences
	s.AddTool(gomcp.NewTool("premiere_get_ab_differences",
		gomcp.WithDescription("List all differences between two sequences in an A/B comparison."),
		gomcp.WithNumber("seq_index_a", gomcp.Required(), gomcp.Description("Zero-based index of sequence A")),
		gomcp.WithNumber("seq_index_b", gomcp.Required(), gomcp.Description("Zero-based index of sequence B")),
	), verH(orch, logger, "get_ab_differences", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetABDifferences(ctx,
			gomcp.ParseInt(req, "seq_index_a", 0),
			gomcp.ParseInt(req, "seq_index_b", 1))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clipboard Extended (14-16)
	// -----------------------------------------------------------------------

	// 14. premiere_get_clipboard_contents
	s.AddTool(gomcp.NewTool("premiere_get_clipboard_contents",
		gomcp.WithDescription("Get information about the current editing clipboard contents."),
	), verH(orch, logger, "get_clipboard_contents", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipboardContents(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_clear_clipboard
	s.AddTool(gomcp.NewTool("premiere_clear_clipboard",
		gomcp.WithDescription("Clear the editing clipboard."),
	), verH(orch, logger, "clear_clipboard", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearClipboard(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_clipboard_has_content
	s.AddTool(gomcp.NewTool("premiere_clipboard_has_content",
		gomcp.WithDescription("Check whether the editing clipboard currently has content."),
	), verH(orch, logger, "clipboard_has_content", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClipboardHasContent(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// History (17-20)
	// -----------------------------------------------------------------------

	// 17. premiere_get_undo_history
	s.AddTool(gomcp.NewTool("premiere_get_undo_history",
		gomcp.WithDescription("Get the most recent undo history entries."),
		gomcp.WithNumber("count", gomcp.Description("Number of history entries to return (default: 20)")),
	), verH(orch, logger, "get_undo_history", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetUndoHistory(ctx, gomcp.ParseInt(req, "count", 20))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_get_undo_count
	s.AddTool(gomcp.NewTool("premiere_get_undo_count",
		gomcp.WithDescription("Get the number of available undo steps."),
	), verH(orch, logger, "get_undo_count", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetUndoCount(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_undo_multiple
	s.AddTool(gomcp.NewTool("premiere_undo_multiple",
		gomcp.WithDescription("Undo multiple editing steps at once."),
		gomcp.WithNumber("count", gomcp.Required(), gomcp.Description("Number of undo steps to perform")),
	), verH(orch, logger, "undo_multiple", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		count := gomcp.ParseInt(req, "count", 0)
		if count <= 0 {
			return gomcp.NewToolResultError("parameter 'count' must be a positive integer"), nil
		}
		result, err := orch.UndoMultiple(ctx, count)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_redo_multiple
	s.AddTool(gomcp.NewTool("premiere_redo_multiple",
		gomcp.WithDescription("Redo multiple editing steps at once."),
		gomcp.WithNumber("count", gomcp.Required(), gomcp.Description("Number of redo steps to perform")),
	), verH(orch, logger, "redo_multiple", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		count := gomcp.ParseInt(req, "count", 0)
		if count <= 0 {
			return gomcp.NewToolResultError("parameter 'count' must be a positive integer"), nil
		}
		result, err := orch.RedoMultiple(ctx, count)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Project Backup (21-25)
	// -----------------------------------------------------------------------

	// 21. premiere_create_project_backup
	s.AddTool(gomcp.NewTool("premiere_create_project_backup",
		gomcp.WithDescription("Create a full project backup to the specified directory."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the backup output directory")),
		gomcp.WithBoolean("include_media", gomcp.Description("Whether to include media files in the backup (default: false)")),
	), verH(orch, logger, "create_project_backup", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.CreateProjectBackup(ctx, outputPath, gomcp.ParseBoolean(req, "include_media", false))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_get_auto_save_versions
	s.AddTool(gomcp.NewTool("premiere_get_auto_save_versions",
		gomcp.WithDescription("List all auto-save versions of the current project."),
	), verH(orch, logger, "get_auto_save_versions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAutoSaveVersions(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_restore_auto_save
	s.AddTool(gomcp.NewTool("premiere_restore_auto_save",
		gomcp.WithDescription("Restore the project from a specific auto-save version."),
		gomcp.WithString("version_path", gomcp.Required(), gomcp.Description("Absolute path to the auto-save version file (.prproj)")),
	), verH(orch, logger, "restore_auto_save", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		vp := gomcp.ParseString(req, "version_path", "")
		if vp == "" {
			return gomcp.NewToolResultError("parameter 'version_path' is required"), nil
		}
		result, err := orch.RestoreAutoSave(ctx, vp)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_auto_save_now
	s.AddTool(gomcp.NewTool("premiere_auto_save_now",
		gomcp.WithDescription("Trigger an immediate auto-save of the current project."),
	), verH(orch, logger, "auto_save_now", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAutoSaveNow(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_get_backup_schedule
	s.AddTool(gomcp.NewTool("premiere_get_backup_schedule",
		gomcp.WithDescription("Get the current backup schedule information including auto-save interval and max versions."),
	), verH(orch, logger, "get_backup_schedule", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetBackupSchedule(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Project Migration (26-30)
	// -----------------------------------------------------------------------

	// 26. premiere_upgrade_project_version
	s.AddTool(gomcp.NewTool("premiere_upgrade_project_version",
		gomcp.WithDescription("Upgrade an older Premiere Pro project file to the current version format."),
		gomcp.WithString("project_path", gomcp.Required(), gomcp.Description("Absolute path to the project file to upgrade")),
	), verH(orch, logger, "upgrade_project_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		pp := gomcp.ParseString(req, "project_path", "")
		if pp == "" {
			return gomcp.NewToolResultError("parameter 'project_path' is required"), nil
		}
		result, err := orch.UpgradeProjectVersion(ctx, pp)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_get_project_version
	s.AddTool(gomcp.NewTool("premiere_get_project_version",
		gomcp.WithDescription("Get the version information of a Premiere Pro project file."),
		gomcp.WithString("project_path", gomcp.Required(), gomcp.Description("Absolute path to the project file")),
	), verH(orch, logger, "get_project_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		pp := gomcp.ParseString(req, "project_path", "")
		if pp == "" {
			return gomcp.NewToolResultError("parameter 'project_path' is required"), nil
		}
		result, err := orch.GetProjectVersion(ctx, pp)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_export_project_for_older_version
	s.AddTool(gomcp.NewTool("premiere_export_project_for_older_version",
		gomcp.WithDescription("Export the current project in a format compatible with an older version of Premiere Pro."),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output project file")),
		gomcp.WithString("target_version", gomcp.Required(), gomcp.Description("Target Premiere Pro version (e.g. '2023', '2022', '2021')")),
	), verH(orch, logger, "export_project_for_older_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		op := gomcp.ParseString(req, "output_path", "")
		if op == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		tv := gomcp.ParseString(req, "target_version", "")
		if tv == "" {
			return gomcp.NewToolResultError("parameter 'target_version' is required"), nil
		}
		result, err := orch.ExportProjectForOlderVersion(ctx, op, tv)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_check_project_compatibility
	s.AddTool(gomcp.NewTool("premiere_check_project_compatibility",
		gomcp.WithDescription("Check whether a project file is compatible with the current version of Premiere Pro."),
		gomcp.WithString("project_path", gomcp.Required(), gomcp.Description("Absolute path to the project file to check")),
	), verH(orch, logger, "check_project_compatibility", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		pp := gomcp.ParseString(req, "project_path", "")
		if pp == "" {
			return gomcp.NewToolResultError("parameter 'project_path' is required"), nil
		}
		result, err := orch.CheckProjectCompatibility(ctx, pp)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_import_project_from_other_nle
	s.AddTool(gomcp.NewTool("premiere_import_project_from_other_nle",
		gomcp.WithDescription("Import a project from another NLE (DaVinci Resolve, Final Cut Pro X, or Avid Media Composer)."),
		gomcp.WithString("source_path", gomcp.Required(), gomcp.Description("Absolute path to the source project file")),
		gomcp.WithString("source_format", gomcp.Required(), gomcp.Description("Source NLE format"),
			gomcp.Enum("davinci", "fcpx", "avid")),
	), verH(orch, logger, "import_project_from_other_nle", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		sp := gomcp.ParseString(req, "source_path", "")
		if sp == "" {
			return gomcp.NewToolResultError("parameter 'source_path' is required"), nil
		}
		sf := gomcp.ParseString(req, "source_format", "")
		if sf == "" {
			return gomcp.NewToolResultError("parameter 'source_format' is required"), nil
		}
		result, err := orch.ImportProjectFromOtherNLE(ctx, sp, sf)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
