package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerBatchTools registers all batch operations and automation MCP tools.
func registerBatchTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Batch Import (1-2)
	// -----------------------------------------------------------------------

	// 1. premiere_batch_import_with_metadata
	s.AddTool(
		gomcp.NewTool("premiere_batch_import_with_metadata",
			gomcp.WithDescription("Import multiple files with optional metadata assignment. Each item can specify a target bin, label color, and metadata key-value pairs."),
			gomcp.WithString("items_json",
				gomcp.Required(),
				gomcp.Description("JSON array of objects: [{\"path\":\"/file.mp4\", \"bin\":\"Footage\", \"label\":3, \"metadata\":{\"key\":\"val\"}}]"),
			),
		),
		batchH(orch, logger, "batch_import_with_metadata", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			itemsJSON := gomcp.ParseString(req, "items_json", "")
			if itemsJSON == "" {
				return gomcp.NewToolResultError("parameter 'items_json' is required"), nil
			}
			result, err := orch.BatchImportWithMetadata(ctx, itemsJSON)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 2. premiere_import_image_sequence
	s.AddTool(
		gomcp.NewTool("premiere_import_image_sequence",
			gomcp.WithDescription("Import an image sequence from a folder as a video clip. Premiere Pro detects numbered image files automatically."),
			gomcp.WithString("folder_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the folder containing numbered image files"),
			),
			gomcp.WithNumber("fps",
				gomcp.Description("Frame rate for the image sequence (default: 24)"),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Name of the bin to import into (default: root)"),
			),
		),
		batchH(orch, logger, "import_image_sequence", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			folderPath := gomcp.ParseString(req, "folder_path", "")
			if folderPath == "" {
				return gomcp.NewToolResultError("parameter 'folder_path' is required"), nil
			}
			fps := gomcp.ParseFloat64(req, "fps", 24)
			targetBin := gomcp.ParseString(req, "target_bin", "")
			result, err := orch.ImportImageSequence(ctx, folderPath, fps, targetBin)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Batch Export (3-4)
	// -----------------------------------------------------------------------

	// 3. premiere_batch_export_sequences
	s.AddTool(
		gomcp.NewTool("premiere_batch_export_sequences",
			gomcp.WithDescription("Export multiple sequences by their indices using a specified preset."),
			gomcp.WithArray("sequence_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based sequence indices to export"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithString("output_dir",
				gomcp.Required(),
				gomcp.Description("Absolute path to the output directory"),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the AME export preset (.epr)"),
			),
		),
		batchH(orch, logger, "batch_export_sequences", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputDir := gomcp.ParseString(req, "output_dir", "")
			if outputDir == "" {
				return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
			}
			presetPath := gomcp.ParseString(req, "preset_path", "")
			if presetPath == "" {
				return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
			}
			indices, err := extractIntSlice(req, "sequence_indices")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid sequence_indices: %v", err)), nil
			}
			result, err := orch.BatchExportSequences(ctx, indices, outputDir, presetPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 4. premiere_export_all_sequences
	s.AddTool(
		gomcp.NewTool("premiere_export_all_sequences",
			gomcp.WithDescription("Export every sequence in the project using a specified preset."),
			gomcp.WithString("output_dir",
				gomcp.Required(),
				gomcp.Description("Absolute path to the output directory"),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the AME export preset (.epr)"),
			),
		),
		batchH(orch, logger, "export_all_sequences", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputDir := gomcp.ParseString(req, "output_dir", "")
			if outputDir == "" {
				return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
			}
			presetPath := gomcp.ParseString(req, "preset_path", "")
			if presetPath == "" {
				return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
			}
			result, err := orch.ExportAllSequences(ctx, outputDir, presetPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Batch Effects (5-7)
	// -----------------------------------------------------------------------

	// 5. premiere_apply_effect_to_multiple_clips
	s.AddTool(
		gomcp.NewTool("premiere_apply_effect_to_multiple_clips",
			gomcp.WithDescription("Apply the same effect to multiple clips on a track. Uses the QE DOM to apply effects by name."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index"),
			),
			gomcp.WithArray("clip_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based clip indices to apply the effect to"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithString("effect_name",
				gomcp.Required(),
				gomcp.Description("Effect name as it appears in the Effects panel"),
			),
		),
		batchH(orch, logger, "apply_effect_to_multiple_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			trackType := gomcp.ParseString(req, "track_type", "video")
			trackIndex := gomcp.ParseInt(req, "track_index", 0)
			effectName := gomcp.ParseString(req, "effect_name", "")
			if effectName == "" {
				return gomcp.NewToolResultError("parameter 'effect_name' is required"), nil
			}
			clipIndices, err := extractIntSlice(req, "clip_indices")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid clip_indices: %v", err)), nil
			}
			result, err := orch.ApplyEffectToMultipleClips(ctx, trackType, trackIndex, clipIndices, effectName)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 6. premiere_remove_all_effects
	s.AddTool(
		gomcp.NewTool("premiere_remove_all_effects",
			gomcp.WithDescription("Remove all applied effects from a clip, keeping intrinsic properties (Motion, Opacity for video)."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index"),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index on the track"),
			),
		),
		batchH(orch, logger, "remove_all_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.RemoveAllEffects(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 7. premiere_apply_transition_to_all_cuts
	s.AddTool(
		gomcp.NewTool("premiere_apply_transition_to_all_cuts",
			gomcp.WithDescription("Apply a transition to every cut point on a video track. Uses QE DOM to add transitions between consecutive clips."),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
			gomcp.WithString("transition_name",
				gomcp.Description("Transition name (default: Cross Dissolve)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Transition duration in seconds (default: 1.0)"),
			),
		),
		batchH(orch, logger, "apply_transition_to_all_cuts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ApplyTransitionToAllCuts(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseString(req, "transition_name", "Cross Dissolve"), gomcp.ParseFloat64(req, "duration", 1.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Batch Color (8-9)
	// -----------------------------------------------------------------------

	// 8. premiere_apply_lut_to_all_clips
	s.AddTool(
		gomcp.NewTool("premiere_apply_lut_to_all_clips",
			gomcp.WithDescription("Apply a LUT file to all clips on a video track via Lumetri Color."),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
			gomcp.WithString("lut_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the LUT file (.cube, .3dl, .look)"),
			),
		),
		batchH(orch, logger, "apply_lut_to_all_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			lutPath := gomcp.ParseString(req, "lut_path", "")
			if lutPath == "" {
				return gomcp.NewToolResultError("parameter 'lut_path' is required"), nil
			}
			result, err := orch.ApplyLUTToAllClips(ctx, gomcp.ParseInt(req, "track_index", 0), lutPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 9. premiere_reset_color_on_all_clips
	s.AddTool(
		gomcp.NewTool("premiere_reset_color_on_all_clips",
			gomcp.WithDescription("Reset Lumetri Color on all clips of a video track by removing the Lumetri Color effect."),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
		),
		batchH(orch, logger, "reset_color_on_all_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ResetColorOnAllClips(ctx, gomcp.ParseInt(req, "track_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Batch Audio (10-12)
	// -----------------------------------------------------------------------

	// 10. premiere_normalize_all_audio
	s.AddTool(
		gomcp.NewTool("premiere_normalize_all_audio",
			gomcp.WithDescription("Normalize audio to a target dB level on all audio clips across all tracks in the active sequence."),
			gomcp.WithNumber("target_db",
				gomcp.Required(),
				gomcp.Description("Target audio level in dB (-96 to +15)"),
			),
		),
		batchH(orch, logger, "normalize_all_audio", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.NormalizeAllAudio(ctx, gomcp.ParseFloat64(req, "target_db", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 11. premiere_mute_all_audio_tracks
	s.AddTool(
		gomcp.NewTool("premiere_mute_all_audio_tracks",
			gomcp.WithDescription("Mute all audio tracks in the active sequence."),
		),
		batchH(orch, logger, "mute_all_audio_tracks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.MuteAllAudioTracks(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 12. premiere_unmute_all_audio_tracks
	s.AddTool(
		gomcp.NewTool("premiere_unmute_all_audio_tracks",
			gomcp.WithDescription("Unmute all audio tracks in the active sequence."),
		),
		batchH(orch, logger, "unmute_all_audio_tracks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.UnmuteAllAudioTracks(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Conforming (13-14)
	// -----------------------------------------------------------------------

	// 13. premiere_conform_sequence_to_clip
	s.AddTool(
		gomcp.NewTool("premiere_conform_sequence_to_clip",
			gomcp.WithDescription("Match the active sequence settings (resolution, frame rate) to a clip's native properties."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin to conform to"),
			),
		),
		batchH(orch, logger, "conform_sequence_to_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ConformSequenceToClip(ctx, gomcp.ParseInt(req, "project_item_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 14. premiere_scale_all_clips_to_frame
	s.AddTool(
		gomcp.NewTool("premiere_scale_all_clips_to_frame",
			gomcp.WithDescription("Scale all video clips on all tracks to fit the sequence frame size."),
		),
		batchH(orch, logger, "scale_all_clips_to_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ScaleAllClipsToFrame(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Timeline Operations (15-18)
	// -----------------------------------------------------------------------

	// 15. premiere_select_all_clips_on_track
	s.AddTool(
		gomcp.NewTool("premiere_select_all_clips_on_track",
			gomcp.WithDescription("Select all clips on a specific track."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index"),
			),
		),
		batchH(orch, logger, "select_all_clips_on_track", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SelectAllClipsOnTrack(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 16. premiere_select_all_clips_between
	s.AddTool(
		gomcp.NewTool("premiere_select_all_clips_between",
			gomcp.WithDescription("Select all clips (on all tracks) that overlap with a time range."),
			gomcp.WithNumber("start_seconds",
				gomcp.Required(),
				gomcp.Description("Start time in seconds"),
			),
			gomcp.WithNumber("end_seconds",
				gomcp.Required(),
				gomcp.Description("End time in seconds"),
			),
		),
		batchH(orch, logger, "select_all_clips_between", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SelectAllClipsBetween(ctx, gomcp.ParseFloat64(req, "start_seconds", 0), gomcp.ParseFloat64(req, "end_seconds", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 17. premiere_delete_all_clips_between
	s.AddTool(
		gomcp.NewTool("premiere_delete_all_clips_between",
			gomcp.WithDescription("Delete all clips within a time range on a specific track."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index"),
			),
			gomcp.WithNumber("start_seconds",
				gomcp.Required(),
				gomcp.Description("Start time in seconds"),
			),
			gomcp.WithNumber("end_seconds",
				gomcp.Required(),
				gomcp.Description("End time in seconds"),
			),
		),
		batchH(orch, logger, "delete_all_clips_between", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.DeleteAllClipsBetween(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_seconds", 0), gomcp.ParseFloat64(req, "end_seconds", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 18. premiere_ripple_delete_all_gaps
	s.AddTool(
		gomcp.NewTool("premiere_ripple_delete_all_gaps",
			gomcp.WithDescription("Close all gaps on all tracks by moving clips to eliminate empty space between them."),
		),
		batchH(orch, logger, "ripple_delete_all_gaps", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.RippleDeleteAllGaps(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Project Cleanup (19-22)
	// -----------------------------------------------------------------------

	// 19. premiere_remove_unused_media
	s.AddTool(
		gomcp.NewTool("premiere_remove_unused_media",
			gomcp.WithDescription("Find and remove media files that are not used in any sequence. WARNING: This permanently removes project items."),
		),
		batchH(orch, logger, "remove_unused_media", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.RemoveUnusedMedia(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 20. premiere_get_unused_media
	s.AddTool(
		gomcp.NewTool("premiere_get_unused_media",
			gomcp.WithDescription("List media files that are not used in any sequence, without removing them."),
		),
		batchH(orch, logger, "get_unused_media", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetUnusedMedia(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 21. premiere_flatten_all_bins
	s.AddTool(
		gomcp.NewTool("premiere_flatten_all_bins",
			gomcp.WithDescription("Move all project items to the root bin, flattening the bin structure. Removes empty bins afterwards."),
		),
		batchH(orch, logger, "flatten_all_bins", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.FlattenAllBins(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 22. premiere_auto_organize_bins
	s.AddTool(
		gomcp.NewTool("premiere_auto_organize_bins",
			gomcp.WithDescription("Auto-organize project items into bins by media type. Creates Video, Audio, Images, and Graphics bins and moves items based on file extension."),
		),
		batchH(orch, logger, "auto_organize_bins", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.AutoOrganizeBins(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Markers Batch (23-27)
	// -----------------------------------------------------------------------

	// 23. premiere_export_markers_as_csv
	s.AddTool(
		gomcp.NewTool("premiere_export_markers_as_csv",
			gomcp.WithDescription("Export all markers from the active sequence as a CSV file."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the output CSV"),
			),
		),
		batchH(orch, logger, "export_markers_as_csv", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportMarkersAsCSV(ctx, outputPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 24. premiere_export_markers_as_edl
	s.AddTool(
		gomcp.NewTool("premiere_export_markers_as_edl",
			gomcp.WithDescription("Export all markers from the active sequence as an EDL (Edit Decision List) file."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the output EDL"),
			),
		),
		batchH(orch, logger, "export_markers_as_edl", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportMarkersAsEDL(ctx, outputPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 25. premiere_import_markers_from_csv
	s.AddTool(
		gomcp.NewTool("premiere_import_markers_from_csv",
			gomcp.WithDescription("Import markers from a CSV file into the active sequence. CSV format: Name,Comment,Start(seconds),End(seconds),Color"),
			gomcp.WithString("csv_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the CSV file"),
			),
		),
		batchH(orch, logger, "import_markers_from_csv", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			csvPath := gomcp.ParseString(req, "csv_path", "")
			if csvPath == "" {
				return gomcp.NewToolResultError("parameter 'csv_path' is required"), nil
			}
			result, err := orch.ImportMarkersFromCSV(ctx, csvPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 26. premiere_delete_all_markers
	s.AddTool(
		gomcp.NewTool("premiere_delete_all_markers",
			gomcp.WithDescription("Delete all markers from the active sequence. WARNING: This cannot be undone."),
		),
		batchH(orch, logger, "delete_all_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.DeleteAllMarkers(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 27. premiere_convert_markers_to_clips
	s.AddTool(
		gomcp.NewTool("premiere_convert_markers_to_clips",
			gomcp.WithDescription("Set in/out points at marker positions. Optionally filter by marker color index."),
			gomcp.WithString("marker_color",
				gomcp.Description("Marker color index to filter by (empty = all markers)"),
			),
		),
		batchH(orch, logger, "convert_markers_to_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			markerColor := gomcp.ParseString(req, "marker_color", "")
			result, err := orch.ConvertMarkersToClips(ctx, markerColor)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Automation (28-30)
	// -----------------------------------------------------------------------

	// 28. premiere_run_extend_script
	s.AddTool(
		gomcp.NewTool("premiere_run_extend_script",
			gomcp.WithDescription("Execute arbitrary ExtendScript code in Premiere Pro. This is an escape hatch for operations not covered by other tools."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("ExtendScript code to evaluate"),
			),
		),
		batchH(orch, logger, "run_extend_script", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			result, err := orch.RunExtendScript(ctx, script)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 29. premiere_get_system_info
	s.AddTool(
		gomcp.NewTool("premiere_get_system_info",
			gomcp.WithDescription("Get system information including OS, Premiere Pro version, GPU, and memory details."),
		),
		batchH(orch, logger, "get_system_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetSystemInfo(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 30. premiere_get_recent_projects
	s.AddTool(
		gomcp.NewTool("premiere_get_recent_projects",
			gomcp.WithDescription("List recent Premiere Pro projects including the currently open project."),
		),
		batchH(orch, logger, "get_recent_projects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetRecentProjects(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)
}

// batchH is a small wrapper that logs the tool name before delegating to the handler.
func batchH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// extractIntSlice is defined in workspace_tools.go and reused here.
