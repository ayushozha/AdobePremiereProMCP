package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerClipTools registers all clip operations MCP tools.
func registerClipTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Basic clip placement
	// -----------------------------------------------------------------------

	// premiere_insert_clip — ripple insert a project item
	s.AddTool(
		gomcp.NewTool("premiere_insert_clip",
			gomcp.WithDescription("Ripple-insert a project item onto the active sequence timeline at a given time. Unlike overwrite, insert pushes all existing clips after the insertion point forward to make room, preserving the rest of the edit. Both video and audio portions of the clip are placed on the specified tracks."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin. Use premiere_get_project_items to list items and their indices."),
			),
			gomcp.WithNumber("time",
				gomcp.Description("Insert time in seconds on the timeline (default: 0). Existing clips after this point are pushed forward."),
			),
			gomcp.WithNumber("video_track_index",
				gomcp.Description("Zero-based video track index for the video portion of the clip (default: 0). Track 0 is the bottom-most video track. Use premiere_get_video_tracks to list tracks."),
			),
			gomcp.WithNumber("audio_track_index",
				gomcp.Description("Zero-based audio track index for the audio portion of the clip (default: 0). Track 0 is the first audio track. Use premiere_get_audio_tracks to list tracks."),
			),
		),
		makeInsertClipHandler(orch, logger),
	)

	// premiere_overwrite_clip — overwrite at position
	s.AddTool(
		gomcp.NewTool("premiere_overwrite_clip",
			gomcp.WithDescription("Overwrite (replace) clips on the active sequence timeline with a project item at a given time. Unlike insert, overwrite does not push existing clips forward -- it replaces whatever is at the target position for the duration of the new clip. This is the most common way to place clips on a timeline."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin. Use premiere_get_project_items to list items and their indices."),
			),
			gomcp.WithNumber("time",
				gomcp.Description("Overwrite time in seconds on the timeline (default: 0). Any existing content at this position is replaced."),
			),
			gomcp.WithNumber("video_track_index",
				gomcp.Description("Zero-based video track index for the video portion of the clip (default: 0). Track 0 is the bottom-most video track."),
			),
			gomcp.WithNumber("audio_track_index",
				gomcp.Description("Zero-based audio track index for the audio portion of the clip (default: 0). Track 0 is the first audio track."),
			),
		),
		makeOverwriteClipHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Clip manipulation
	// -----------------------------------------------------------------------

	// premiere_remove_clip_from_track
	s.AddTool(
		gomcp.NewTool("premiere_remove_clip_from_track",
			gomcp.WithDescription("Remove a clip from a specific track by its index. With ripple=false (default), a gap is left where the clip was (lift edit). With ripple=true, subsequent clips slide back to fill the gap (ripple delete). Use premiere_get_clips_on_track to find clip indices."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track). Use premiere_get_video_tracks or premiere_get_audio_tracks to list tracks."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip on the specified track. Use premiere_get_clips_on_track to list clips and their indices."),
			),
			gomcp.WithBoolean("ripple",
				gomcp.Description("If true, close the gap by sliding subsequent clips backward (ripple delete). If false (default), leave the gap (lift edit)."),
			),
		),
		makeRemoveClipFromTrackHandler(orch, logger),
	)

	// premiere_move_clip
	s.AddTool(
		gomcp.NewTool("premiere_move_clip",
			gomcp.WithDescription("Move a clip to a new start time on the same track. The clip's duration remains unchanged. If the new position overlaps another clip, behavior depends on Premiere's overwrite rules. Use premiere_get_clip_info to check the current position first."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip on the specified track."),
			),
			gomcp.WithNumber("new_start_time",
				gomcp.Required(),
				gomcp.Description("New start time in seconds on the timeline. Must be >= 0."),
			),
		),
		makeMoveClipHandler(orch, logger),
	)

	// premiere_copy_clip
	s.AddTool(
		gomcp.NewTool("premiere_copy_clip",
			gomcp.WithDescription("Copy a clip to the internal clipboard for later pasting."),
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
		makeCopyClipHandler(orch, logger),
	)

	// premiere_paste_clip
	s.AddTool(
		gomcp.NewTool("premiere_paste_clip",
			gomcp.WithDescription("Paste the previously copied clip at a new position on a track."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index"),
			),
			gomcp.WithNumber("time",
				gomcp.Required(),
				gomcp.Description("Paste time in seconds"),
			),
		),
		makePasteClipHandler(orch, logger),
	)

	// premiere_duplicate_clip
	s.AddTool(
		gomcp.NewTool("premiere_duplicate_clip",
			gomcp.WithDescription("Duplicate a clip to a new position, optionally on a different track of the same type."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based source track index"),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index on the source track"),
			),
			gomcp.WithNumber("dest_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based destination track index"),
			),
			gomcp.WithNumber("dest_time",
				gomcp.Required(),
				gomcp.Description("Destination time in seconds"),
			),
		),
		makeDuplicateClipHandler(orch, logger),
	)

	// premiere_razor_clip
	s.AddTool(
		gomcp.NewTool("premiere_razor_clip",
			gomcp.WithDescription("Split (razor/cut) the clip at a specific time position on a single track, creating two separate clips from the original. The split point must fall within a clip on the specified track. After splitting, the original clip index refers to the left portion, and the right portion gets the next index. Use premiere_razor_all_tracks to split all tracks at once."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track to split on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("time",
				gomcp.Required(),
				gomcp.Description("Time in seconds at which to split the clip. Must fall within a clip on the specified track."),
			),
		),
		makeRazorClipHandler(orch, logger),
	)

	// premiere_razor_all_tracks
	s.AddTool(
		gomcp.NewTool("premiere_razor_all_tracks",
			gomcp.WithDescription("Split (razor/cut) all clips on every video and audio track at a specific time position, equivalent to using the Razor tool with 'All Tracks' mode. Creates a clean cut point across the entire timeline. Useful for creating a scene break or segmenting a timeline."),
			gomcp.WithNumber("time",
				gomcp.Required(),
				gomcp.Description("Time in seconds at which to split all tracks. Any clip spanning this time point on any track will be split into two clips."),
			),
		),
		makeRazorAllTracksHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Clip information
	// -----------------------------------------------------------------------

	// premiere_get_clip_info
	s.AddTool(
		gomcp.NewTool("premiere_get_clip_info",
			gomcp.WithDescription("Get detailed information about a specific clip on the timeline, including: clip name, source media path, timeline in/out points, source in/out points, duration, playback speed, direction (forward/reverse), enabled state, and a list of all applied effects with their parameters. This is the most detailed clip inspection tool."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track). Use premiere_get_video_tracks or premiere_get_audio_tracks to list tracks."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip on the specified track. Use premiere_get_clips_on_track to list clips."),
			),
		),
		makeGetClipInfoHandler(orch, logger),
	)

	// premiere_get_clips_on_track
	s.AddTool(
		gomcp.NewTool("premiere_get_clips_on_track",
			gomcp.WithDescription("List all clips on a specific track with their positions, durations, names, and indices. Returns clips ordered by timeline position (left to right). Use this to find clip indices for other clip operations (move, trim, split, add effects)."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track to list clips from: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track). Use premiere_get_video_tracks or premiere_get_audio_tracks to list tracks."),
			),
		),
		makeGetClipsOnTrackHandler(orch, logger),
	)

	// premiere_get_all_clips
	s.AddTool(
		gomcp.NewTool("premiere_get_all_clips",
			gomcp.WithDescription("Get a summary of all clips across every video and audio track in the active sequence. Returns each clip's track type, track index, clip index, name, start/end times, and duration. Use this for a full inventory of the current edit. For detailed info on a single clip, use premiere_get_clip_info."),
		),
		makeGetAllClipsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Clip properties
	// -----------------------------------------------------------------------

	// premiere_set_clip_name
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_name",
			gomcp.WithDescription("Rename a clip on the timeline."),
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
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("New name for the clip"),
			),
		),
		makeSetClipNameHandler(orch, logger),
	)

	// premiere_set_clip_enabled
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_enabled",
			gomcp.WithDescription("Enable or disable a clip on the timeline. Disabled clips are skipped during playback."),
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
			gomcp.WithBoolean("enabled",
				gomcp.Required(),
				gomcp.Description("true to enable, false to disable"),
			),
		),
		makeSetClipEnabledHandler(orch, logger),
	)

	// premiere_set_clip_speed
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_speed",
			gomcp.WithDescription("Change the playback speed of a clip on the timeline. Speed affects duration: doubling speed halves duration, halving speed doubles duration. Audio pitch is affected unless time-remapping is used separately. To reverse a clip, use premiere_reverse_clip instead."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip on the specified track."),
			),
			gomcp.WithNumber("speed",
				gomcp.Required(),
				gomcp.Description("Playback speed multiplier. Must be positive. 1.0 = normal speed, 2.0 = double speed (clip plays in half the time), 0.5 = half speed (slow motion, clip takes twice as long), 0.25 = quarter speed (extreme slow-mo)."),
			),
			gomcp.WithBoolean("ripple",
				gomcp.Description("If true, shift subsequent clips on the same track to accommodate the duration change (ripple edit). If false (default), the clip duration changes but other clips stay in place (may create gaps or overlaps)."),
			),
		),
		makeSetClipSpeedHandler(orch, logger),
	)

	// premiere_reverse_clip
	s.AddTool(
		gomcp.NewTool("premiere_reverse_clip",
			gomcp.WithDescription("Reverse the playback direction of a clip."),
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
		makeReverseClipHandler(orch, logger),
	)

	// premiere_set_clip_in_point
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_in_point",
			gomcp.WithDescription("Set the source in point of a clip (controls which part of the source media is shown)."),
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
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("New source in point in seconds"),
			),
		),
		makeSetClipInPointHandler(orch, logger),
	)

	// premiere_set_clip_out_point
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_out_point",
			gomcp.WithDescription("Set the source out point of a clip (controls where the source media stops)."),
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
			gomcp.WithNumber("seconds",
				gomcp.Required(),
				gomcp.Description("New source out point in seconds"),
			),
		),
		makeSetClipOutPointHandler(orch, logger),
	)

	// premiere_get_clip_speed
	s.AddTool(
		gomcp.NewTool("premiere_get_clip_speed",
			gomcp.WithDescription("Get the current playback speed and direction (forward/reverse) of a clip."),
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
		makeGetClipSpeedHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Trimming
	// -----------------------------------------------------------------------

	// premiere_trim_clip_start
	s.AddTool(
		gomcp.NewTool("premiere_trim_clip_start",
			gomcp.WithDescription("Trim the start (head) of a clip by moving its in point on the timeline. Trimming the start later (increasing new_start_time) reveals less of the beginning of the source media. Trimming earlier (decreasing) reveals more, provided there is available source media (handle). This does not move subsequent clips -- use ripple editing for that."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip on the specified track."),
			),
			gomcp.WithNumber("new_start_time",
				gomcp.Required(),
				gomcp.Description("New start time for the clip in seconds on the timeline. Must be >= 0 and less than the clip's current end time. Use premiere_get_clip_info to see current start/end."),
			),
		),
		makeTrimClipStartHandler(orch, logger),
	)

	// premiere_trim_clip_end
	s.AddTool(
		gomcp.NewTool("premiere_trim_clip_end",
			gomcp.WithDescription("Trim the end (tail) of a clip by moving its out point on the timeline. Trimming the end earlier (decreasing new_end_time) shortens the clip. Trimming later (increasing) extends the clip, provided there is available source media (handle). This does not move subsequent clips."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip on the specified track."),
			),
			gomcp.WithNumber("new_end_time",
				gomcp.Required(),
				gomcp.Description("New end time for the clip in seconds on the timeline. Must be greater than the clip's current start time. Use premiere_get_clip_info to see current start/end."),
			),
		),
		makeTrimClipEndHandler(orch, logger),
	)

	// premiere_extend_clip_to_playhead
	s.AddTool(
		gomcp.NewTool("premiere_extend_clip_to_playhead",
			gomcp.WithDescription("Extend or trim a clip's start or end to the current playhead position."),
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
			gomcp.WithBoolean("trim_end",
				gomcp.Description("If true, adjust clip end to playhead; if false, adjust clip start (default: true)"),
			),
		),
		makeExtendClipToPlayheadHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Subclips
	// -----------------------------------------------------------------------

	// premiere_create_subclip
	s.AddTool(
		gomcp.NewTool("premiere_create_subclip",
			gomcp.WithDescription("Create a subclip from a project item with specified in and out points."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithString("name",
				gomcp.Description("Name for the subclip (defaults to source name + '_subclip')"),
			),
			gomcp.WithNumber("in_point",
				gomcp.Required(),
				gomcp.Description("Source in point in seconds"),
			),
			gomcp.WithNumber("out_point",
				gomcp.Required(),
				gomcp.Description("Source out point in seconds"),
			),
		),
		makeCreateSubclipHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Selection
	// -----------------------------------------------------------------------

	// premiere_select_clip
	s.AddTool(
		gomcp.NewTool("premiere_select_clip",
			gomcp.WithDescription("Select a specific clip on the timeline."),
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
		makeSelectClipHandler(orch, logger),
	)

	// premiere_deselect_all
	s.AddTool(
		gomcp.NewTool("premiere_deselect_all",
			gomcp.WithDescription("Deselect all clips across all tracks in the active sequence."),
		),
		makeDeselectAllHandler(orch, logger),
	)

	// premiere_get_selected_clips
	s.AddTool(
		gomcp.NewTool("premiere_get_selected_clips",
			gomcp.WithDescription("Get all currently selected clips across all tracks."),
		),
		makeGetSelectedClipsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Linking
	// -----------------------------------------------------------------------

	// premiere_link_clips
	s.AddTool(
		gomcp.NewTool("premiere_link_clips",
			gomcp.WithDescription("Link video and audio clips together so they move, trim, and delete as a unit. When clips are linked, selecting or moving one automatically selects/moves the other. This is the default state for imported A/V clips. Provide clip pairs identifying which video clip links to which audio clip."),
			gomcp.WithString("clip_pairs",
				gomcp.Required(),
				gomcp.Description(`JSON array of clip pair objects. Each pair specifies a video clip and audio clip to link together. Format: [{"vTrack":0,"vClip":0,"aTrack":0,"aClip":0}] where vTrack/vClip identify the video clip and aTrack/aClip identify the audio clip (all zero-based indices).`),
			),
		),
		makeLinkClipsHandler(orch, logger),
	)

	// premiere_unlink_clips
	s.AddTool(
		gomcp.NewTool("premiere_unlink_clips",
			gomcp.WithDescription("Unlink a clip from its linked video/audio counterpart, allowing them to be moved, trimmed, and deleted independently. Useful when you need to replace audio without affecting video, or move audio to a different track. Use premiere_get_linked_clips to see what is linked first."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the clip to unlink is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the clip to unlink from its counterpart."),
			),
		),
		makeUnlinkClipsHandler(orch, logger),
	)

	// premiere_get_linked_clips
	s.AddTool(
		gomcp.NewTool("premiere_get_linked_clips",
			gomcp.WithDescription("Find all clips that are linked to a specific clip. Returns the linked counterparts with their track types, track indices, and clip indices. Linked clips move together when editing. If a clip has no linked counterparts, an empty list is returned."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Type of track the reference clip is on: 'video' or 'audio'."),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index (0 = first/bottom track)."),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the reference clip whose linked counterparts you want to find."),
			),
		),
		makeGetLinkedClipsHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Handler constructors — Basic clip placement
// ---------------------------------------------------------------------------

func makeInsertClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_insert_clip")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		time := gomcp.ParseFloat64(req, "time", 0)
		vTrack := gomcp.ParseInt(req, "video_track_index", 0)
		aTrack := gomcp.ParseInt(req, "audio_track_index", 0)
		result, err := orch.InsertClip(ctx, piIndex, time, vTrack, aTrack)
		if err != nil {
			logger.Error("insert clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to insert clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeOverwriteClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_overwrite_clip")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		time := gomcp.ParseFloat64(req, "time", 0)
		vTrack := gomcp.ParseInt(req, "video_track_index", 0)
		aTrack := gomcp.ParseInt(req, "audio_track_index", 0)
		result, err := orch.OverwriteClip(ctx, piIndex, time, vTrack, aTrack)
		if err != nil {
			logger.Error("overwrite clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to overwrite clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Clip manipulation
// ---------------------------------------------------------------------------

func makeRemoveClipFromTrackHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_remove_clip_from_track")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		ripple := gomcp.ParseBoolean(req, "ripple", false)
		result, err := orch.RemoveClipFromTrack(ctx, trackType, trackIndex, clipIndex, ripple)
		if err != nil {
			logger.Error("remove clip from track failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to remove clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeMoveClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_move_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		newStartTime := gomcp.ParseFloat64(req, "new_start_time", 0)
		result, err := orch.MoveClip(ctx, trackType, trackIndex, clipIndex, newStartTime)
		if err != nil {
			logger.Error("move clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to move clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeCopyClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_copy_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.CopyClip(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("copy clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to copy clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makePasteClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_paste_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		time := gomcp.ParseFloat64(req, "time", 0)
		result, err := orch.PasteClip(ctx, trackType, trackIndex, time)
		if err != nil {
			logger.Error("paste clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to paste clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDuplicateClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_duplicate_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		destTrackIndex := gomcp.ParseInt(req, "dest_track_index", 0)
		destTime := gomcp.ParseFloat64(req, "dest_time", 0)
		result, err := orch.DuplicateClip(ctx, trackType, trackIndex, clipIndex, destTrackIndex, destTime)
		if err != nil {
			logger.Error("duplicate clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to duplicate clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRazorClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_razor_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		time := gomcp.ParseFloat64(req, "time", 0)
		result, err := orch.RazorClip(ctx, trackType, trackIndex, time)
		if err != nil {
			logger.Error("razor clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to razor clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRazorAllTracksHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_razor_all_tracks")
		time := gomcp.ParseFloat64(req, "time", 0)
		result, err := orch.RazorAllTracks(ctx, time)
		if err != nil {
			logger.Error("razor all tracks failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to razor all tracks: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Clip information
// ---------------------------------------------------------------------------

func makeGetClipInfoHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_clip_info")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.GetClipInfo(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("get clip info failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get clip info: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetClipsOnTrackHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_clips_on_track")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		result, err := orch.GetClipsOnTrack(ctx, trackType, trackIndex)
		if err != nil {
			logger.Error("get clips on track failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get clips on track: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetAllClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_all_clips")
		result, err := orch.GetAllClips(ctx)
		if err != nil {
			logger.Error("get all clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get all clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Clip properties
// ---------------------------------------------------------------------------

func makeSetClipNameHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_name")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SetClipName(ctx, trackType, trackIndex, clipIndex, name)
		if err != nil {
			logger.Error("set clip name failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip name: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetClipEnabledHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_enabled")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		enabled := gomcp.ParseBoolean(req, "enabled", true)
		result, err := orch.SetClipEnabled(ctx, trackType, trackIndex, clipIndex, enabled)
		if err != nil {
			logger.Error("set clip enabled failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip enabled: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetClipSpeedHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_speed")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		speed := gomcp.ParseFloat64(req, "speed", 1.0)
		ripple := gomcp.ParseBoolean(req, "ripple", false)
		result, err := orch.SetClipSpeed(ctx, trackType, trackIndex, clipIndex, speed, ripple)
		if err != nil {
			logger.Error("set clip speed failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip speed: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeReverseClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_reverse_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.ReverseClip(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("reverse clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to reverse clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetClipInPointHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_in_point")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		seconds := gomcp.ParseFloat64(req, "seconds", 0)
		result, err := orch.SetClipInPoint(ctx, trackType, trackIndex, clipIndex, seconds)
		if err != nil {
			logger.Error("set clip in point failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip in point: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetClipOutPointHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_out_point")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		seconds := gomcp.ParseFloat64(req, "seconds", 0)
		result, err := orch.SetClipOutPoint(ctx, trackType, trackIndex, clipIndex, seconds)
		if err != nil {
			logger.Error("set clip out point failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip out point: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetClipSpeedHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_clip_speed")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.GetClipSpeed(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("get clip speed failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get clip speed: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Trimming
// ---------------------------------------------------------------------------

func makeTrimClipStartHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_trim_clip_start")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		newStartTime := gomcp.ParseFloat64(req, "new_start_time", 0)
		result, err := orch.TrimClipStart(ctx, trackType, trackIndex, clipIndex, newStartTime)
		if err != nil {
			logger.Error("trim clip start failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to trim clip start: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeTrimClipEndHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_trim_clip_end")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		newEndTime := gomcp.ParseFloat64(req, "new_end_time", 0)
		result, err := orch.TrimClipEnd(ctx, trackType, trackIndex, clipIndex, newEndTime)
		if err != nil {
			logger.Error("trim clip end failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to trim clip end: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExtendClipToPlayheadHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_extend_clip_to_playhead")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		trimEnd := gomcp.ParseBoolean(req, "trim_end", true)
		result, err := orch.ExtendClipToPlayhead(ctx, trackType, trackIndex, clipIndex, trimEnd)
		if err != nil {
			logger.Error("extend clip to playhead failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to extend clip to playhead: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Subclips
// ---------------------------------------------------------------------------

func makeCreateSubclipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_create_subclip")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		name := gomcp.ParseString(req, "name", "")
		inPoint := gomcp.ParseFloat64(req, "in_point", 0)
		outPoint := gomcp.ParseFloat64(req, "out_point", 0)
		result, err := orch.CreateSubclip(ctx, piIndex, name, inPoint, outPoint)
		if err != nil {
			logger.Error("create subclip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create subclip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Selection
// ---------------------------------------------------------------------------

func makeSelectClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_select_clip")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.SelectClip(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("select clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to select clip: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDeselectAllHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_deselect_all")
		result, err := orch.DeselectAll(ctx)
		if err != nil {
			logger.Error("deselect all failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to deselect all: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetSelectedClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_selected_clips")
		result, err := orch.GetSelectedClips(ctx)
		if err != nil {
			logger.Error("get selected clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get selected clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Linking
// ---------------------------------------------------------------------------

func makeLinkClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_link_clips")
		clipPairs := gomcp.ParseString(req, "clip_pairs", "")
		if clipPairs == "" {
			return gomcp.NewToolResultError("parameter 'clip_pairs' is required"), nil
		}
		result, err := orch.LinkClips(ctx, clipPairs)
		if err != nil {
			logger.Error("link clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to link clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeUnlinkClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_unlink_clips")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.UnlinkClips(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("unlink clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to unlink clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetLinkedClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_linked_clips")
		trackType := gomcp.ParseString(req, "track_type", "video")
		trackIndex := gomcp.ParseInt(req, "track_index", 0)
		clipIndex := gomcp.ParseInt(req, "clip_index", 0)
		result, err := orch.GetLinkedClips(ctx, trackType, trackIndex, clipIndex)
		if err != nil {
			logger.Error("get linked clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get linked clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}
