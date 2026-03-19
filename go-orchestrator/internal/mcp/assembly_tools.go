package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerAssemblyTools registers timeline assembly, arrangement, composition,
// clip generation, timeline duplication, and timeline analysis MCP tools.
func registerAssemblyTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Timeline Assembly (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_assemble_from_edl
	s.AddTool(
		gomcp.NewTool("premiere_assemble_from_edl",
			gomcp.WithDescription("Assemble a timeline from an EDL JSON specification containing clips, positions, and transitions."),
			gomcp.WithString("edl_json",
				gomcp.Required(),
				gomcp.Description("JSON object with clips array: {\"clips\":[{\"file\":\"/path.mp4\",\"inPoint\":0,\"outPoint\":10,\"trackIndex\":0,\"position\":0}]}"),
			),
		),
		assemblyH(logger, "assemble_from_edl", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			edlJSON := gomcp.ParseString(req, "edl_json", "")
			if edlJSON == "" {
				return gomcp.NewToolResultError("parameter 'edl_json' is required"), nil
			}
			result, err := orch.AssembleFromEDL(ctx, edlJSON)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 2. premiere_assemble_from_csv
	s.AddTool(
		gomcp.NewTool("premiere_assemble_from_csv",
			gomcp.WithDescription("Assemble a timeline from a CSV file with columns: file, in, out, track, position."),
			gomcp.WithString("csv_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the CSV file"),
			),
		),
		assemblyH(logger, "assemble_from_csv", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			csvPath := gomcp.ParseString(req, "csv_path", "")
			if csvPath == "" {
				return gomcp.NewToolResultError("parameter 'csv_path' is required"), nil
			}
			result, err := orch.AssembleFromCSV(ctx, csvPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 3. premiere_assemble_from_folder_order
	s.AddTool(
		gomcp.NewTool("premiere_assemble_from_folder_order",
			gomcp.WithDescription("Import and assemble clips from a folder in filename order, with optional transitions between clips."),
			gomcp.WithString("folder_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the folder containing media files"),
			),
			gomcp.WithString("transition_name",
				gomcp.Description("Transition to apply between clips (default: none)"),
			),
			gomcp.WithNumber("transition_duration",
				gomcp.Description("Transition duration in seconds (default: 1.0)"),
			),
		),
		assemblyH(logger, "assemble_from_folder_order", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			folderPath := gomcp.ParseString(req, "folder_path", "")
			if folderPath == "" {
				return gomcp.NewToolResultError("parameter 'folder_path' is required"), nil
			}
			result, err := orch.AssembleFromFolderOrder(ctx, folderPath,
				gomcp.ParseString(req, "transition_name", ""),
				gomcp.ParseFloat64(req, "transition_duration", 1.0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 4. premiere_interleave_clips
	s.AddTool(
		gomcp.NewTool("premiere_interleave_clips",
			gomcp.WithDescription("Interleave clips from two video tracks, alternating clips A-B-A-B with optional transition between each."),
			gomcp.WithNumber("track_index_a",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the first video track"),
			),
			gomcp.WithNumber("track_index_b",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the second video track"),
			),
			gomcp.WithNumber("transition_duration",
				gomcp.Description("Transition duration in seconds between interleaved clips (default: 0)"),
			),
		),
		assemblyH(logger, "interleave_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.InterleaveClips(ctx,
				gomcp.ParseInt(req, "track_index_a", 0),
				gomcp.ParseInt(req, "track_index_b", 1),
				gomcp.ParseFloat64(req, "transition_duration", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 5. premiere_shuffle_clips
	s.AddTool(
		gomcp.NewTool("premiere_shuffle_clips",
			gomcp.WithDescription("Randomly shuffle the order of clips on a track."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
		),
		assemblyH(logger, "shuffle_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ShuffleClips(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Arrangement (6-11)
	// -----------------------------------------------------------------------

	// 6. premiere_sort_clips_by_duration
	s.AddTool(
		gomcp.NewTool("premiere_sort_clips_by_duration",
			gomcp.WithDescription("Sort clips on a track by their duration, shortest to longest or vice versa."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithBoolean("ascending",
				gomcp.Description("Sort ascending (shortest first) if true, descending if false (default: true)"),
			),
		),
		assemblyH(logger, "sort_clips_by_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SortClipsByDuration(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseBoolean(req, "ascending", true),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 7. premiere_sort_clips_by_name
	s.AddTool(
		gomcp.NewTool("premiere_sort_clips_by_name",
			gomcp.WithDescription("Sort clips on a track alphabetically by clip name."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithBoolean("ascending",
				gomcp.Description("Sort A-Z if true, Z-A if false (default: true)"),
			),
		),
		assemblyH(logger, "sort_clips_by_name", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SortClipsByName(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseBoolean(req, "ascending", true),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 8. premiere_sort_clips_by_file_name
	s.AddTool(
		gomcp.NewTool("premiere_sort_clips_by_file_name",
			gomcp.WithDescription("Sort clips on a track alphabetically by their source file name."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithBoolean("ascending",
				gomcp.Description("Sort A-Z if true, Z-A if false (default: true)"),
			),
		),
		assemblyH(logger, "sort_clips_by_file_name", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SortClipsByFileName(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseBoolean(req, "ascending", true),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 9. premiere_reverse_clip_order
	s.AddTool(
		gomcp.NewTool("premiere_reverse_clip_order",
			gomcp.WithDescription("Reverse the order of all clips on a track."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
		),
		assemblyH(logger, "reverse_clip_order", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ReverseClipOrder(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 10. premiere_distribute_clips_evenly
	s.AddTool(
		gomcp.NewTool("premiere_distribute_clips_evenly",
			gomcp.WithDescription("Distribute clips evenly across a specified total duration by adding equal gaps between them."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithNumber("total_duration",
				gomcp.Required(),
				gomcp.Description("Total duration in seconds to distribute clips over"),
			),
		),
		assemblyH(logger, "distribute_clips_evenly", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.DistributeClipsEvenly(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "total_duration", 60),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 11. premiere_stack_clips
	s.AddTool(
		gomcp.NewTool("premiere_stack_clips",
			gomcp.WithDescription("Remove all gaps on a track and stack clips contiguously from a specified start time."),
			gomcp.WithString("track_type",
				gomcp.Description("Track type: video or audio (default: video)"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds to begin stacking (default: 0)"),
			),
		),
		assemblyH(logger, "stack_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.StackClips(ctx,
				gomcp.ParseString(req, "track_type", "video"),
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Multi-Track Composition (12-15)
	// -----------------------------------------------------------------------

	// 12. premiere_create_overlay_track
	s.AddTool(
		gomcp.NewTool("premiere_create_overlay_track",
			gomcp.WithDescription("Create an overlay composition from a source track over a destination track with specified opacity and blend mode."),
			gomcp.WithNumber("source_track",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the source video track"),
			),
			gomcp.WithNumber("dest_track",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the destination video track"),
			),
			gomcp.WithNumber("opacity",
				gomcp.Description("Overlay opacity percentage 0-100 (default: 50)"),
			),
			gomcp.WithString("blend_mode",
				gomcp.Description("Blend mode (default: Normal)"),
				gomcp.Enum("Normal", "Multiply", "Screen", "Overlay", "SoftLight", "HardLight", "Difference"),
			),
		),
		assemblyH(logger, "create_overlay_track", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.CreateOverlayTrack(ctx,
				gomcp.ParseInt(req, "source_track", 0),
				gomcp.ParseInt(req, "dest_track", 1),
				gomcp.ParseFloat64(req, "opacity", 50),
				gomcp.ParseString(req, "blend_mode", "Normal"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 13. premiere_create_green_screen_composite
	s.AddTool(
		gomcp.NewTool("premiere_create_green_screen_composite",
			gomcp.WithDescription("Create a chroma key composite by keying out a color from a foreground clip and compositing over a background clip."),
			gomcp.WithNumber("fg_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index of the foreground clip"),
			),
			gomcp.WithNumber("fg_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index of the foreground clip"),
			),
			gomcp.WithNumber("bg_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index of the background clip"),
			),
			gomcp.WithNumber("bg_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index of the background clip"),
			),
			gomcp.WithString("key_color",
				gomcp.Description("Chroma key color as hex (default: #00FF00 for green)"),
			),
		),
		assemblyH(logger, "create_green_screen_composite", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.CreateGreenScreenComposite(ctx,
				gomcp.ParseInt(req, "fg_track_index", 0),
				gomcp.ParseInt(req, "fg_clip_index", 0),
				gomcp.ParseInt(req, "bg_track_index", 1),
				gomcp.ParseInt(req, "bg_clip_index", 0),
				gomcp.ParseString(req, "key_color", "#00FF00"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 14. premiere_create_pip_grid
	s.AddTool(
		gomcp.NewTool("premiere_create_pip_grid",
			gomcp.WithDescription("Create a picture-in-picture grid layout from multiple video tracks arranged in a grid pattern."),
			gomcp.WithArray("track_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based video track indices to include in the grid"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithString("layout",
				gomcp.Description("Grid layout (e.g. '2x2', '3x3', '2x3') (default: 2x2)"),
			),
		),
		assemblyH(logger, "create_pip_grid", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			indices, err := extractIntSlice(req, "track_indices")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid track_indices: %v", err)), nil
			}
			result, err := orch.CreatePictureInPictureGrid(ctx, indices,
				gomcp.ParseString(req, "layout", "2x2"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 15. premiere_layer_tracks
	s.AddTool(
		gomcp.NewTool("premiere_layer_tracks",
			gomcp.WithDescription("Layer multiple overlay tracks over a base track with individual opacity settings for each layer."),
			gomcp.WithNumber("base_track",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the base video track"),
			),
			gomcp.WithArray("overlay_tracks",
				gomcp.Required(),
				gomcp.Description("Array of zero-based track indices to overlay"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithArray("opacities",
				gomcp.Description("Array of opacity values (0-100) corresponding to each overlay track"),
				gomcp.WithNumberItems(),
			),
		),
		assemblyH(logger, "layer_tracks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			overlays, err := extractIntSlice(req, "overlay_tracks")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid overlay_tracks: %v", err)), nil
			}
			opacities, _ := extractFloat64Slice(req, "opacities")
			result, err := orch.LayerTracks(ctx,
				gomcp.ParseInt(req, "base_track", 0),
				overlays,
				opacities,
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Clip Generation (16-21)
	// -----------------------------------------------------------------------

	// 16. premiere_generate_black_clip
	s.AddTool(
		gomcp.NewTool("premiere_generate_black_clip",
			gomcp.WithDescription("Generate a black video clip and place it on the timeline."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5)"),
			),
		),
		assemblyH(logger, "generate_black_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GenerateBlackClip(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "duration", 5),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 17. premiere_generate_color_clip
	s.AddTool(
		gomcp.NewTool("premiere_generate_color_clip",
			gomcp.WithDescription("Generate a solid color clip and place it on the timeline."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5)"),
			),
			gomcp.WithString("color",
				gomcp.Description("Solid color as hex string e.g. '#FF0000' (default: #FF0000)"),
			),
		),
		assemblyH(logger, "generate_color_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GenerateColorClip(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "duration", 5),
				gomcp.ParseString(req, "color", "#FF0000"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 18. premiere_generate_gradient_clip
	s.AddTool(
		gomcp.NewTool("premiere_generate_gradient_clip",
			gomcp.WithDescription("Generate a gradient clip transitioning between two colors."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5)"),
			),
			gomcp.WithString("color_start",
				gomcp.Description("Start color as hex (default: #000000)"),
			),
			gomcp.WithString("color_end",
				gomcp.Description("End color as hex (default: #FFFFFF)"),
			),
			gomcp.WithString("direction",
				gomcp.Description("Gradient direction"),
				gomcp.Enum("horizontal", "vertical", "diagonal"),
			),
		),
		assemblyH(logger, "generate_gradient_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GenerateGradientClip(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "duration", 5),
				gomcp.ParseString(req, "color_start", "#000000"),
				gomcp.ParseString(req, "color_end", "#FFFFFF"),
				gomcp.ParseString(req, "direction", "horizontal"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 19. premiere_generate_test_pattern
	s.AddTool(
		gomcp.NewTool("premiere_generate_test_pattern",
			gomcp.WithDescription("Generate a test pattern clip such as color bars, grid, or checkerboard."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5)"),
			),
			gomcp.WithString("pattern",
				gomcp.Description("Test pattern type"),
				gomcp.Enum("bars", "grid", "checkerboard"),
			),
		),
		assemblyH(logger, "generate_test_pattern", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GenerateTestPattern(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "duration", 5),
				gomcp.ParseString(req, "pattern", "bars"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 20. premiere_generate_silence
	s.AddTool(
		gomcp.NewTool("premiere_generate_silence",
			gomcp.WithDescription("Generate a silent audio clip on the timeline."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based audio track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5)"),
			),
		),
		assemblyH(logger, "generate_silence", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GenerateSilence(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "duration", 5),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 21. premiere_generate_tone
	s.AddTool(
		gomcp.NewTool("premiere_generate_tone",
			gomcp.WithDescription("Generate an audio tone clip (sine, square, sawtooth waveform) on the timeline."),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based audio track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("duration",
				gomcp.Description("Duration in seconds (default: 5)"),
			),
			gomcp.WithNumber("frequency",
				gomcp.Description("Tone frequency in Hz (default: 440)"),
			),
			gomcp.WithNumber("amplitude",
				gomcp.Description("Amplitude 0.0-1.0 (default: 0.5)"),
			),
		),
		assemblyH(logger, "generate_tone", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GenerateTone(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "duration", 5),
				gomcp.ParseFloat64(req, "frequency", 440),
				gomcp.ParseFloat64(req, "amplitude", 0.5),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Timeline Duplication (22-25)
	// -----------------------------------------------------------------------

	// 22. premiere_duplicate_timeline_section
	s.AddTool(
		gomcp.NewTool("premiere_duplicate_timeline_section",
			gomcp.WithDescription("Copy a section of the timeline (all tracks) from a time range to another position."),
			gomcp.WithNumber("start_time",
				gomcp.Required(),
				gomcp.Description("Start time of the section in seconds"),
			),
			gomcp.WithNumber("end_time",
				gomcp.Required(),
				gomcp.Description("End time of the section in seconds"),
			),
			gomcp.WithNumber("dest_time",
				gomcp.Required(),
				gomcp.Description("Destination time in seconds to paste the duplicated section"),
			),
		),
		assemblyH(logger, "duplicate_timeline_section", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.DuplicateTimelineSection(ctx,
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "end_time", 0),
				gomcp.ParseFloat64(req, "dest_time", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 23. premiere_repeat_timeline_section
	s.AddTool(
		gomcp.NewTool("premiere_repeat_timeline_section",
			gomcp.WithDescription("Repeat a section of the timeline N times, appending copies after the section end."),
			gomcp.WithNumber("start_time",
				gomcp.Required(),
				gomcp.Description("Start time of the section in seconds"),
			),
			gomcp.WithNumber("end_time",
				gomcp.Required(),
				gomcp.Description("End time of the section in seconds"),
			),
			gomcp.WithNumber("count",
				gomcp.Required(),
				gomcp.Description("Number of times to repeat the section"),
			),
		),
		assemblyH(logger, "repeat_timeline_section", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.RepeatTimelineSection(ctx,
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "end_time", 0),
				gomcp.ParseInt(req, "count", 1),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 24. premiere_mirror_timeline
	s.AddTool(
		gomcp.NewTool("premiere_mirror_timeline",
			gomcp.WithDescription("Mirror (reverse) the entire timeline so clips play in reverse chronological order."),
		),
		assemblyH(logger, "mirror_timeline", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.MirrorTimeline(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 25. premiere_split_timeline_at_playhead
	s.AddTool(
		gomcp.NewTool("premiere_split_timeline_at_playhead",
			gomcp.WithDescription("Split the timeline into two logical sections at the current playhead position, reporting clips before and after."),
		),
		assemblyH(logger, "split_timeline_at_playhead", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SplitTimelineAtPlayhead(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Timeline Analysis (26-30)
	// -----------------------------------------------------------------------

	// 26. premiere_get_timeline_gap_report
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline_gap_report",
			gomcp.WithDescription("Generate a detailed report of all gaps (empty spaces) across all tracks in the timeline."),
		),
		assemblyH(logger, "get_timeline_gap_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetTimelineGapReport(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 27. premiere_get_timeline_conflict_report
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline_conflict_report",
			gomcp.WithDescription("Generate a report of all overlapping clips (conflicts) on the timeline."),
		),
		assemblyH(logger, "get_timeline_conflict_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetTimelineConflictReport(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 28. premiere_get_timeline_effects_report
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline_effects_report",
			gomcp.WithDescription("Generate a report of all effects applied to clips across the entire timeline."),
		),
		assemblyH(logger, "get_timeline_effects_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetTimelineEffectsReport(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 29. premiere_get_timeline_duration_breakdown
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline_duration_breakdown",
			gomcp.WithDescription("Generate a duration breakdown report showing total time by clip type and source file format."),
		),
		assemblyH(logger, "get_timeline_duration_breakdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetTimelineDurationBreakdown(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 30. premiere_get_timeline_track_usage_report
	s.AddTool(
		gomcp.NewTool("premiere_get_timeline_track_usage_report",
			gomcp.WithDescription("Generate a track usage report showing used vs empty time and clip count per track."),
		),
		assemblyH(logger, "get_timeline_track_usage_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetTimelineTrackUsageReport(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)
}

// assemblyH wraps a handler with debug logging for assembly tool invocations.
func assemblyH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// extractFloat64Slice pulls a []float64 from the request arguments map.
func extractFloat64Slice(req gomcp.CallToolRequest, key string) ([]float64, error) {
	raw := gomcp.ParseArgument(req, key, nil)
	if raw == nil {
		return nil, fmt.Errorf("key %q not found", key)
	}
	switch v := raw.(type) {
	case []float64:
		return v, nil
	case []any:
		out := make([]float64, 0, len(v))
		for i, item := range v {
			switch n := item.(type) {
			case float64:
				out = append(out, n)
			case int:
				out = append(out, float64(n))
			default:
				return nil, fmt.Errorf("element %d of %q is not a number", i, key)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("key %q is not an array (got %T)", key, raw)
	}
}
