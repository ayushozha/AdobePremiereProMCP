package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Clip Operations (Extended)
// ---------------------------------------------------------------------------

// InsertClip ripple-inserts a project item at the given time.
func (e *Engine) InsertClip(ctx context.Context, projectItemIndex int, time float64, vTrackIndex, aTrackIndex int) (*GenericResult, error) {
	e.logger.Debug("insert_clip",
		zap.Int("project_item_index", projectItemIndex),
		zap.Float64("time", time),
		zap.Int("v_track", vTrackIndex),
		zap.Int("a_track", aTrackIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"time":             time,
		"vTrackIndex":      vTrackIndex,
		"aTrackIndex":      aTrackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "insertClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("InsertClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// OverwriteClip overwrites at the given time with a project item.
func (e *Engine) OverwriteClip(ctx context.Context, projectItemIndex int, time float64, vTrackIndex, aTrackIndex int) (*GenericResult, error) {
	e.logger.Debug("overwrite_clip",
		zap.Int("project_item_index", projectItemIndex),
		zap.Float64("time", time),
		zap.Int("v_track", vTrackIndex),
		zap.Int("a_track", aTrackIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"time":             time,
		"vTrackIndex":      vTrackIndex,
		"aTrackIndex":      aTrackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "overwriteClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OverwriteClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// RemoveClipFromTrack removes a clip from a specific track.
func (e *Engine) RemoveClipFromTrack(ctx context.Context, trackType string, trackIndex, clipIndex int, ripple bool) (*GenericResult, error) {
	e.logger.Debug("remove_clip_from_track",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Bool("ripple", ripple),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"ripple":     ripple,
	})
	result, err := e.premiere.EvalCommand(ctx, "removeClipFromTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemoveClipFromTrack: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// MoveClip moves a clip to a new start time.
func (e *Engine) MoveClip(ctx context.Context, trackType string, trackIndex, clipIndex int, newStartTime float64) (*GenericResult, error) {
	e.logger.Debug("move_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Float64("new_start_time", newStartTime),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":    trackType,
		"trackIndex":   trackIndex,
		"clipIndex":    clipIndex,
		"newStartTime": newStartTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "moveClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MoveClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// CopyClip copies a clip to the internal clipboard.
func (e *Engine) CopyClip(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("copy_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "copyClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CopyClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// PasteClip pastes the previously copied clip at a new position.
func (e *Engine) PasteClip(ctx context.Context, trackType string, trackIndex int, time float64) (*GenericResult, error) {
	e.logger.Debug("paste_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Float64("time", time),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"time":       time,
	})
	result, err := e.premiere.EvalCommand(ctx, "pasteClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PasteClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// DuplicateClip duplicates a clip to a new position.
func (e *Engine) DuplicateClip(ctx context.Context, trackType string, trackIndex, clipIndex, destTrackIndex int, destTime float64) (*GenericResult, error) {
	e.logger.Debug("duplicate_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Int("dest_track", destTrackIndex),
		zap.Float64("dest_time", destTime),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"destTrackIndex": destTrackIndex,
		"destTime":       destTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "duplicateClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DuplicateClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// RazorClip splits a clip at the given time.
func (e *Engine) RazorClip(ctx context.Context, trackType string, trackIndex int, time float64) (*GenericResult, error) {
	e.logger.Debug("razor_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Float64("time", time),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"time":       time,
	})
	result, err := e.premiere.EvalCommand(ctx, "razorClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RazorClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// RazorAllTracks splits all tracks at the given time.
func (e *Engine) RazorAllTracks(ctx context.Context, time float64) (*GenericResult, error) {
	e.logger.Debug("razor_all_tracks", zap.Float64("time", time))
	argsJSON, _ := json.Marshal(map[string]any{
		"time": time,
	})
	result, err := e.premiere.EvalCommand(ctx, "razorAllTracks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RazorAllTracks: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// GetClipInfo returns detailed information about a clip.
func (e *Engine) GetClipInfo(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("get_clip_info",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipInfo: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// GetClipsOnTrack returns all clips on a specific track.
func (e *Engine) GetClipsOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	e.logger.Debug("get_clips_on_track",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipsOnTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipsOnTrack: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// GetAllClips returns all clips across all tracks.
func (e *Engine) GetAllClips(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_all_clips")
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAllClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAllClips: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// SetClipName renames a clip.
func (e *Engine) SetClipName(ctx context.Context, trackType string, trackIndex, clipIndex int, name string) (*GenericResult, error) {
	e.logger.Debug("set_clip_name",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.String("name", name),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"name":       name,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipName", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipName: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// SetClipEnabled enables or disables a clip.
func (e *Engine) SetClipEnabled(ctx context.Context, trackType string, trackIndex, clipIndex int, enabled bool) (*GenericResult, error) {
	e.logger.Debug("set_clip_enabled",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Bool("enabled", enabled),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"enabled":    enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipEnabled", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipEnabled: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// SetClipSpeed changes clip playback speed.
func (e *Engine) SetClipSpeed(ctx context.Context, trackType string, trackIndex, clipIndex int, speed float64, ripple bool) (*GenericResult, error) {
	e.logger.Debug("set_clip_speed",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Float64("speed", speed),
		zap.Bool("ripple", ripple),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"speed":      speed,
		"ripple":     ripple,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipSpeed", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipSpeed: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// ReverseClip reverses a clip's playback direction.
func (e *Engine) ReverseClip(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("reverse_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "reverseClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReverseClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// SetClipInPoint sets a clip's source in point.
func (e *Engine) SetClipInPoint(ctx context.Context, trackType string, trackIndex, clipIndex int, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_clip_in_point",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Float64("seconds", seconds),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"seconds":    seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipInPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipInPoint: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// SetClipOutPoint sets a clip's source out point.
func (e *Engine) SetClipOutPoint(ctx context.Context, trackType string, trackIndex, clipIndex int, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_clip_out_point",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Float64("seconds", seconds),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"seconds":    seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipOutPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipOutPoint: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// GetClipSpeed returns the current speed and direction of a clip.
func (e *Engine) GetClipSpeed(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("get_clip_speed",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipSpeed", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipSpeed: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// TrimClipStart trims the start of a clip.
func (e *Engine) TrimClipStart(ctx context.Context, trackType string, trackIndex, clipIndex int, newStartTime float64) (*GenericResult, error) {
	e.logger.Debug("trim_clip_start",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Float64("new_start_time", newStartTime),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":    trackType,
		"trackIndex":   trackIndex,
		"clipIndex":    clipIndex,
		"newStartTime": newStartTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "trimClipStart", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TrimClipStart: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// TrimClipEnd trims the end of a clip.
func (e *Engine) TrimClipEnd(ctx context.Context, trackType string, trackIndex, clipIndex int, newEndTime float64) (*GenericResult, error) {
	e.logger.Debug("trim_clip_end",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Float64("new_end_time", newEndTime),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"newEndTime": newEndTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "trimClipEnd", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TrimClipEnd: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// ExtendClipToPlayhead extends or trims a clip to the playhead position.
func (e *Engine) ExtendClipToPlayhead(ctx context.Context, trackType string, trackIndex, clipIndex int, trimEnd bool) (*GenericResult, error) {
	e.logger.Debug("extend_clip_to_playhead",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
		zap.Bool("trim_end", trimEnd),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"trimEnd":    trimEnd,
	})
	result, err := e.premiere.EvalCommand(ctx, "extendClipToPlayhead", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExtendClipToPlayhead: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// CreateSubclip creates a subclip from a project item.
func (e *Engine) CreateSubclip(ctx context.Context, projectItemIndex int, name string, inPoint, outPoint float64) (*GenericResult, error) {
	e.logger.Debug("create_subclip",
		zap.Int("project_item_index", projectItemIndex),
		zap.String("name", name),
		zap.Float64("in_point", inPoint),
		zap.Float64("out_point", outPoint),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"name":             name,
		"inPoint":          inPoint,
		"outPoint":         outPoint,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSubclip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSubclip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// SelectClip selects a clip on the timeline.
func (e *Engine) SelectClip(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("select_clip",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "selectClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SelectClip: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// DeselectAll deselects all clips on the timeline.
func (e *Engine) DeselectAll(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("deselect_all")
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "deselectAll", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DeselectAll: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// GetSelectedClips returns all currently selected clips.
func (e *Engine) GetSelectedClips(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_selected_clips")
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSelectedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSelectedClips: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// LinkClips links video and audio clips.
func (e *Engine) LinkClips(ctx context.Context, clipPairsJSON string) (*GenericResult, error) {
	e.logger.Debug("link_clips", zap.String("clip_pairs", clipPairsJSON))
	argsJSON, _ := json.Marshal(map[string]any{
		"clipPairs": clipPairsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "linkClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LinkClips: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// UnlinkClips unlinks a clip from its linked counterpart.
func (e *Engine) UnlinkClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("unlink_clips",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "unlinkClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("UnlinkClips: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

// GetLinkedClips returns all clips linked to a given clip.
func (e *Engine) GetLinkedClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("get_linked_clips",
		zap.String("track_type", trackType),
		zap.Int("track_index", trackIndex),
		zap.Int("clip_index", clipIndex),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getLinkedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLinkedClips: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}
