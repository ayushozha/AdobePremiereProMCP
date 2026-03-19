package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// camH wraps a handler with debug logging for camera tool invocations.
func camH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerCameraTools registers all 30 camera, shot type detection, and
// cinematography MCP tools.
func registerCameraTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Shot/Camera Metadata (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_get_clip_camera_info
	s.AddTool(gomcp.NewTool("premiere_get_clip_camera_info",
		gomcp.WithDescription("Get camera metadata (make, model, lens, ISO, shutter speed, aperture) from a clip's embedded XMP metadata."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
	), camH(logger, "get_clip_camera_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipCameraInfo(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_get_clip_gps_info
	s.AddTool(gomcp.NewTool("premiere_get_clip_gps_info",
		gomcp.WithDescription("Get GPS coordinates (latitude, longitude, altitude) from a clip's embedded XMP metadata."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
	), camH(logger, "get_clip_gps_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipGPSInfo(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_clip_record_date
	s.AddTool(gomcp.NewTool("premiere_get_clip_record_date",
		gomcp.WithDescription("Get the recording date and time from a clip's XMP metadata (DateTimeOriginal or CreateDate)."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
	), camH(logger, "get_clip_record_date", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipRecordDate(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_sort_clips_by_record_date
	s.AddTool(gomcp.NewTool("premiere_sort_clips_by_record_date",
		gomcp.WithDescription("Sort clips in a bin by their recording date from XMP metadata. Returns sorted order."),
		gomcp.WithString("bin_path",
			gomcp.Description("Path to the bin to sort (default: root bin). Use '/' for root.")),
	), camH(logger, "sort_clips_by_record_date", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SortClipsByRecordDate(ctx, gomcp.ParseString(req, "bin_path", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_group_clips_by_camera
	s.AddTool(gomcp.NewTool("premiere_group_clips_by_camera",
		gomcp.WithDescription("Group clips by camera make/model into separate bins. Creates sub-bins named after each camera."),
		gomcp.WithString("bin_path",
			gomcp.Description("Path to the bin to organize (default: root bin). Use '/' for root.")),
	), camH(logger, "group_clips_by_camera", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GroupClipsByCamera(ctx, gomcp.ParseString(req, "bin_path", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Shot Management (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_mark_shot_type
	s.AddTool(gomcp.NewTool("premiere_mark_shot_type",
		gomcp.WithDescription("Mark a clip on the timeline with a shot type classification (wide, medium, closeup, insert, cutaway). Uses colored markers."),
		gomcp.WithString("track_type", gomcp.Required(),
			gomcp.Description("Track type"),
			gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(),
			gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(),
			gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("shot_type", gomcp.Required(),
			gomcp.Description("Shot type classification"),
			gomcp.Enum("wide", "medium", "closeup", "insert", "cutaway")),
	), camH(logger, "mark_shot_type", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "")
		if trackType == "" {
			return gomcp.NewToolResultError("parameter 'track_type' is required"), nil
		}
		shotType := gomcp.ParseString(req, "shot_type", "")
		if shotType == "" {
			return gomcp.NewToolResultError("parameter 'shot_type' is required"), nil
		}
		result, err := orch.MarkShotType(ctx, trackType,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			shotType,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_get_shot_type
	s.AddTool(gomcp.NewTool("premiere_get_shot_type",
		gomcp.WithDescription("Get the shot type marker (wide, medium, closeup, insert, cutaway) from a clip on the timeline."),
		gomcp.WithString("track_type", gomcp.Required(),
			gomcp.Description("Track type"),
			gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(),
			gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(),
			gomcp.Description("Zero-based clip index on the track")),
	), camH(logger, "get_shot_type", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "")
		if trackType == "" {
			return gomcp.NewToolResultError("parameter 'track_type' is required"), nil
		}
		result, err := orch.GetShotType(ctx, trackType,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_filter_by_shot_type
	s.AddTool(gomcp.NewTool("premiere_filter_by_shot_type",
		gomcp.WithDescription("Get all clips in a sequence that match a specific shot type (wide, medium, closeup, insert, cutaway)."),
		gomcp.WithNumber("sequence_index",
			gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("shot_type", gomcp.Required(),
			gomcp.Description("Shot type to filter by"),
			gomcp.Enum("wide", "medium", "closeup", "insert", "cutaway")),
	), camH(logger, "filter_by_shot_type", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		shotType := gomcp.ParseString(req, "shot_type", "")
		if shotType == "" {
			return gomcp.NewToolResultError("parameter 'shot_type' is required"), nil
		}
		result, err := orch.FilterByShotType(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			shotType,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_create_shot_list
	s.AddTool(gomcp.NewTool("premiere_create_shot_list",
		gomcp.WithDescription("Export a shot list from a sequence as a CSV file, including clip names, timecodes, shot types, scenes, and takes."),
		gomcp.WithNumber("sequence_index",
			gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(),
			gomcp.Description("Absolute file path for the output CSV file")),
	), camH(logger, "create_shot_list", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.CreateShotList(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_import_shot_list
	s.AddTool(gomcp.NewTool("premiere_import_shot_list",
		gomcp.WithDescription("Import a shot list from a CSV file and apply shot types, scenes, and takes to timeline clips."),
		gomcp.WithString("csv_path", gomcp.Required(),
			gomcp.Description("Absolute path to the CSV shot list file")),
		gomcp.WithNumber("sequence_index",
			gomcp.Description("Zero-based sequence index (default: 0)")),
	), camH(logger, "import_shot_list", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		csvPath := gomcp.ParseString(req, "csv_path", "")
		if csvPath == "" {
			return gomcp.NewToolResultError("parameter 'csv_path' is required"), nil
		}
		result, err := orch.ImportShotList(ctx, csvPath,
			gomcp.ParseInt(req, "sequence_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Scene/Take Management (11-15)
	// -----------------------------------------------------------------------

	// 11. premiere_mark_scene
	s.AddTool(gomcp.NewTool("premiere_mark_scene",
		gomcp.WithDescription("Mark a clip on the timeline with a scene number. Replaces any existing scene marker on the clip."),
		gomcp.WithString("track_type", gomcp.Required(),
			gomcp.Description("Track type"),
			gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(),
			gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(),
			gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("scene_number", gomcp.Required(),
			gomcp.Description("Scene number or identifier (e.g. '1', '2A')")),
	), camH(logger, "mark_scene", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "")
		if trackType == "" {
			return gomcp.NewToolResultError("parameter 'track_type' is required"), nil
		}
		sceneNumber := gomcp.ParseString(req, "scene_number", "")
		if sceneNumber == "" {
			return gomcp.NewToolResultError("parameter 'scene_number' is required"), nil
		}
		result, err := orch.MarkScene(ctx, trackType,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			sceneNumber,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_mark_take
	s.AddTool(gomcp.NewTool("premiere_mark_take",
		gomcp.WithDescription("Mark a clip on the timeline with a take number. Replaces any existing take marker on the clip."),
		gomcp.WithString("track_type", gomcp.Required(),
			gomcp.Description("Track type"),
			gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(),
			gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(),
			gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("take_number", gomcp.Required(),
			gomcp.Description("Take number or identifier (e.g. '1', '2')")),
	), camH(logger, "mark_take", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "")
		if trackType == "" {
			return gomcp.NewToolResultError("parameter 'track_type' is required"), nil
		}
		takeNumber := gomcp.ParseString(req, "take_number", "")
		if takeNumber == "" {
			return gomcp.NewToolResultError("parameter 'take_number' is required"), nil
		}
		result, err := orch.MarkTake(ctx, trackType,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			takeNumber,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_get_best_take
	s.AddTool(gomcp.NewTool("premiere_get_best_take",
		gomcp.WithDescription("Get the longest (best) take for a given scene number from the active sequence. Returns all takes sorted by duration."),
		gomcp.WithString("scene_number", gomcp.Required(),
			gomcp.Description("Scene number to find the best take for")),
	), camH(logger, "get_best_take", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		sceneNumber := gomcp.ParseString(req, "scene_number", "")
		if sceneNumber == "" {
			return gomcp.NewToolResultError("parameter 'scene_number' is required"), nil
		}
		result, err := orch.GetBestTake(ctx, sceneNumber)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_organize_by_scenes_and_takes
	s.AddTool(gomcp.NewTool("premiere_organize_by_scenes_and_takes",
		gomcp.WithDescription("Auto-organize project items into scene bins by parsing scene/take info from filenames (e.g. S01T01_clip.mp4, Scene1_Take2_clip.mov)."),
	), camH(logger, "organize_by_scenes_and_takes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.OrganizeByScenesAndTakes(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_get_scene_list
	s.AddTool(gomcp.NewTool("premiere_get_scene_list",
		gomcp.WithDescription("Get all scenes with their takes from the active sequence. Returns scene numbers, take counts, clip names, and durations."),
	), camH(logger, "get_scene_list", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSceneList(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Camera Matching (16-18)
	// -----------------------------------------------------------------------

	// 16. premiere_match_camera_settings
	s.AddTool(gomcp.NewTool("premiere_match_camera_settings",
		gomcp.WithDescription("Compare camera settings (make, model, ISO, shutter, aperture, lens) between two clips. Shows matches and differences."),
		gomcp.WithNumber("clip1_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the first project item")),
		gomcp.WithNumber("clip2_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the second project item")),
	), camH(logger, "match_camera_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.MatchCameraSettings(ctx,
			gomcp.ParseInt(req, "clip1_index", 0),
			gomcp.ParseInt(req, "clip2_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_find_clips_from_same_camera
	s.AddTool(gomcp.NewTool("premiere_find_clips_from_same_camera",
		gomcp.WithDescription("Find all project items that were shot with the same camera (matching make and model) as the specified clip."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the reference project item")),
	), camH(logger, "find_clips_from_same_camera", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FindClipsFromSameCamera(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_create_multicam_by_camera
	s.AddTool(gomcp.NewTool("premiere_create_multicam_by_camera",
		gomcp.WithDescription("Create a multicam sequence that groups clips by camera make/model, placing each camera's footage on a separate track."),
		gomcp.WithString("output_name",
			gomcp.Description("Name for the new multicam sequence (default: 'Multicam_ByCamera')")),
	), camH(logger, "create_multicam_by_camera", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateMulticamByCamera(ctx,
			gomcp.ParseString(req, "output_name", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Timecode Management (19-22)
	// -----------------------------------------------------------------------

	// 19. premiere_get_source_timecode
	s.AddTool(gomcp.NewTool("premiere_get_source_timecode",
		gomcp.WithDescription("Get the original source timecode embedded in a clip's XMP metadata (startTimecode or altTimecode)."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
	), camH(logger, "get_source_timecode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetSourceTimecode(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_set_source_timecode_offset
	s.AddTool(gomcp.NewTool("premiere_set_source_timecode_offset",
		gomcp.WithDescription("Set a source timecode offset on a project item's XMP metadata."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
		gomcp.WithString("offset", gomcp.Required(),
			gomcp.Description("Timecode offset value (e.g. '01:00:00:00')")),
	), camH(logger, "set_source_timecode_offset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		offset := gomcp.ParseString(req, "offset", "")
		if offset == "" {
			return gomcp.NewToolResultError("parameter 'offset' is required"), nil
		}
		result, err := orch.SetSourceTimecodeOffset(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			offset,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_sync_by_timecode
	s.AddTool(gomcp.NewTool("premiere_sync_by_timecode",
		gomcp.WithDescription("Sync clips across multiple video tracks by aligning their source timecodes. Requires at least 2 track indices."),
		gomcp.WithString("track_indices", gomcp.Required(),
			gomcp.Description("Comma-separated list of zero-based video track indices to sync (e.g. '0,1,2')")),
	), camH(logger, "sync_by_timecode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		indicesStr := gomcp.ParseString(req, "track_indices", "")
		if indicesStr == "" {
			return gomcp.NewToolResultError("parameter 'track_indices' is required"), nil
		}
		parts := strings.Split(indicesStr, ",")
		indices := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			v, err := strconv.Atoi(p)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid track index %q: %v", p, err)), nil
			}
			indices = append(indices, v)
		}
		if len(indices) < 2 {
			return gomcp.NewToolResultError("need at least 2 track indices"), nil
		}
		result, err := orch.SyncByTimecode(ctx, indices)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_find_timecode_breaks
	s.AddTool(gomcp.NewTool("premiere_find_timecode_breaks",
		gomcp.WithDescription("Find gaps in timecode continuity on a track. Reports gap positions and durations."),
		gomcp.WithString("track_type", gomcp.Required(),
			gomcp.Description("Track type"),
			gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(),
			gomcp.Description("Zero-based track index")),
	), camH(logger, "find_timecode_breaks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		trackType := gomcp.ParseString(req, "track_type", "")
		if trackType == "" {
			return gomcp.NewToolResultError("parameter 'track_type' is required"), nil
		}
		result, err := orch.FindTimecodeBreaks(ctx, trackType,
			gomcp.ParseInt(req, "track_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clip Rating (23-26)
	// -----------------------------------------------------------------------

	// 23. premiere_rate_clip
	s.AddTool(gomcp.NewTool("premiere_rate_clip",
		gomcp.WithDescription("Rate a clip from 1 to 5 stars by setting the XMP Rating metadata field."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
		gomcp.WithNumber("rating", gomcp.Required(),
			gomcp.Description("Rating value from 1 to 5")),
	), camH(logger, "rate_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RateClip(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			gomcp.ParseInt(req, "rating", 3),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_get_clip_rating
	s.AddTool(gomcp.NewTool("premiere_get_clip_rating",
		gomcp.WithDescription("Get the rating (1-5 stars) of a clip from its XMP metadata."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
	), camH(logger, "get_clip_rating", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipRating(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_filter_by_rating
	s.AddTool(gomcp.NewTool("premiere_filter_by_rating",
		gomcp.WithDescription("Get all project items with a rating greater than or equal to the specified minimum (1-5)."),
		gomcp.WithNumber("min_rating", gomcp.Required(),
			gomcp.Description("Minimum rating threshold (1-5)")),
	), camH(logger, "filter_by_rating", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FilterByRating(ctx,
			gomcp.ParseInt(req, "min_rating", 1),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_get_top_rated_clips
	s.AddTool(gomcp.NewTool("premiere_get_top_rated_clips",
		gomcp.WithDescription("Get the top N highest-rated clips from the project, sorted by rating descending."),
		gomcp.WithNumber("count",
			gomcp.Description("Number of top-rated clips to return (default: 10)")),
	), camH(logger, "get_top_rated_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTopRatedClips(ctx,
			gomcp.ParseInt(req, "count", 10),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clip Notes (27-30)
	// -----------------------------------------------------------------------

	// 27. premiere_set_clip_note
	s.AddTool(gomcp.NewTool("premiere_set_clip_note",
		gomcp.WithDescription("Set a text note on a clip using XMP dc:description metadata. Overwrites any existing note."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
		gomcp.WithString("note", gomcp.Required(),
			gomcp.Description("Text note to set on the clip")),
	), camH(logger, "set_clip_note", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		note := gomcp.ParseString(req, "note", "")
		if note == "" {
			return gomcp.NewToolResultError("parameter 'note' is required"), nil
		}
		result, err := orch.SetClipNote(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			note,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_get_clip_note
	s.AddTool(gomcp.NewTool("premiere_get_clip_note",
		gomcp.WithDescription("Get the text note (XMP dc:description) from a clip."),
		gomcp.WithNumber("project_item_index", gomcp.Required(),
			gomcp.Description("Zero-based index of the project item in the root bin")),
	), camH(logger, "get_clip_note", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipNote(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_search_clip_notes
	s.AddTool(gomcp.NewTool("premiere_search_clip_notes",
		gomcp.WithDescription("Search all clip notes (XMP dc:description) for a text string. Case-insensitive search."),
		gomcp.WithString("search_text", gomcp.Required(),
			gomcp.Description("Text to search for in clip notes")),
	), camH(logger, "search_clip_notes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		searchText := gomcp.ParseString(req, "search_text", "")
		if searchText == "" {
			return gomcp.NewToolResultError("parameter 'search_text' is required"), nil
		}
		result, err := orch.SearchClipNotes(ctx, searchText)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_export_clip_notes
	s.AddTool(gomcp.NewTool("premiere_export_clip_notes",
		gomcp.WithDescription("Export all clip notes to a CSV or JSON file. Includes clip name, note text, and rating."),
		gomcp.WithString("output_path", gomcp.Required(),
			gomcp.Description("Absolute file path for the output file")),
		gomcp.WithString("format",
			gomcp.Description("Output format: csv or json (default: csv)"),
			gomcp.Enum("csv", "json")),
	), camH(logger, "export_clip_notes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportClipNotes(ctx,
			outputPath,
			gomcp.ParseString(req, "format", "csv"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
