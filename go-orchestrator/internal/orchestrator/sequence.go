package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Sequence Management
// ---------------------------------------------------------------------------

// CreateSequenceFromClips creates a new sequence from specified project items.
func (e *Engine) CreateSequenceFromClips(ctx context.Context, name string, clipIndices []int) (*SequenceResult, error) {
	e.logger.Debug("create_sequence_from_clips",
		zap.String("name", name),
		zap.Ints("clip_indices", clipIndices),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"name":        name,
		"clipIndices": clipIndices,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSequenceFromClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create sequence %q from clips — make sure a project is open (try premiere_open_project) and clips are imported: %w", name, err)
	}
	var out SequenceResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("CreateSequenceFromClips: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// DuplicateSequence duplicates an existing sequence.
func (e *Engine) DuplicateSequence(ctx context.Context, sequenceIndex int) (*SequenceResult, error) {
	e.logger.Debug("duplicate_sequence", zap.Int("sequence_index", sequenceIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "duplicateSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate sequence at index %d — use premiere_get_sequence_list to verify the index: %w", sequenceIndex, err)
	}
	var out SequenceResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("DuplicateSequence: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// DeleteSequence removes a sequence from the project.
func (e *Engine) DeleteSequence(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	e.logger.Debug("delete_sequence", zap.Int("sequence_index", sequenceIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to delete sequence at index %d — use premiere_get_sequence_list to verify the index: %w", sequenceIndex, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// RenameSequence renames a sequence.
func (e *Engine) RenameSequence(ctx context.Context, sequenceIndex int, newName string) (*GenericResult, error) {
	e.logger.Debug("rename_sequence",
		zap.Int("sequence_index", sequenceIndex),
		zap.String("new_name", newName),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"newName":       newName,
	})
	result, err := e.premiere.EvalCommand(ctx, "renameSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to rename sequence at index %d — use premiere_get_sequence_list to verify the index: %w", sequenceIndex, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// GetSequenceSettings returns the full settings of a sequence.
func (e *Engine) GetSequenceSettings(ctx context.Context, sequenceIndex int) (*SequenceSettings, error) {
	e.logger.Debug("get_sequence_settings", zap.Int("sequence_index", sequenceIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for sequence at index %d — create a sequence first with premiere_create_sequence: %w", sequenceIndex, err)
	}
	var out SequenceSettings
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetSequenceSettings: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// SetSequenceSettings updates sequence settings.
func (e *Engine) SetSequenceSettings(ctx context.Context, params *SetSequenceSettingsParams) (*GenericResult, error) {
	if params == nil {
		return nil, fmt.Errorf("set_sequence_settings: params must not be nil — provide sequenceIndex and the settings to update")
	}
	e.logger.Debug("set_sequence_settings", zap.Int("sequence_index", params.SequenceIndex))
	argsJSON, _ := json.Marshal(params)
	result, err := e.premiere.EvalCommand(ctx, "setSequenceSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to update settings for sequence at index %d: %w", params.SequenceIndex, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// GetActiveSequence returns details of the currently active sequence.
func (e *Engine) GetActiveSequence(ctx context.Context) (*SequenceSettings, error) {
	e.logger.Debug("get_active_sequence")
	result, err := e.premiere.EvalCommand(ctx, "getActiveSequence", "{}")
	if err != nil {
		return nil, fmt.Errorf("no active sequence found — create a sequence first with premiere_create_sequence, or set one active with premiere_set_active_sequence: %w", err)
	}
	var out SequenceSettings
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetActiveSequence: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// SetActiveSequence makes a sequence the active one.
func (e *Engine) SetActiveSequence(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	e.logger.Debug("set_active_sequence", zap.Int("sequence_index", sequenceIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "setActiveSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set active sequence at index %d — use premiere_get_sequence_list to verify the index: %w", sequenceIndex, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// GetSequenceList returns all sequences in the project.
func (e *Engine) GetSequenceList(ctx context.Context) (*SequenceListResult, error) {
	e.logger.Debug("get_sequence_list")
	result, err := e.premiere.EvalCommand(ctx, "getSequenceList", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to list sequences — open a project first with premiere_open_project: %w", err)
	}
	var out SequenceListResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetSequenceList: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Playhead & In/Out Points
// ---------------------------------------------------------------------------

// GetPlayheadPosition returns the current playhead position.
func (e *Engine) GetPlayheadPosition(ctx context.Context) (*PlayheadResult, error) {
	e.logger.Debug("get_playhead_position")
	result, err := e.premiere.EvalCommand(ctx, "getPlayheadPosition", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to get playhead position — make sure a sequence is active (try premiere_create_sequence or premiere_set_active_sequence): %w", err)
	}
	var out PlayheadResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetPlayheadPosition: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// SetPlayheadPosition moves the playhead to a specific position.
func (e *Engine) SetPlayheadPosition(ctx context.Context, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_playhead_position", zap.Float64("seconds", seconds))
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setPlayheadPosition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set playhead position — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// SetInPoint sets the sequence in point.
func (e *Engine) SetInPoint(ctx context.Context, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_in_point", zap.Float64("seconds", seconds))
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setInPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set in point — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// SetOutPoint sets the sequence out point.
func (e *Engine) SetOutPoint(ctx context.Context, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_out_point", zap.Float64("seconds", seconds))
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setOutPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set out point — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// GetInOutPoints returns the current in/out points.
func (e *Engine) GetInOutPoints(ctx context.Context) (*InOutPointsResult, error) {
	e.logger.Debug("get_in_out_points")
	result, err := e.premiere.EvalCommand(ctx, "getInOutPoints", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to get in/out points — make sure a sequence is active: %w", err)
	}
	var out InOutPointsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetInOutPoints: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// ClearInOutPoints resets the in/out points to the sequence boundaries.
func (e *Engine) ClearInOutPoints(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("clear_in_out_points")
	result, err := e.premiere.EvalCommand(ctx, "clearInOutPoints", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to clear in/out points — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// ---------------------------------------------------------------------------
// Work Area & Preview
// ---------------------------------------------------------------------------

// SetWorkArea sets the work area boundaries.
func (e *Engine) SetWorkArea(ctx context.Context, inSeconds, outSeconds float64) (*GenericResult, error) {
	e.logger.Debug("set_work_area",
		zap.Float64("in_seconds", inSeconds),
		zap.Float64("out_seconds", outSeconds),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"inSeconds":  inSeconds,
		"outSeconds": outSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setWorkArea", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set work area — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// RenderPreviewFiles renders preview files for a time range.
func (e *Engine) RenderPreviewFiles(ctx context.Context, inSeconds, outSeconds float64) (*GenericResult, error) {
	e.logger.Debug("render_preview_files",
		zap.Float64("in_seconds", inSeconds),
		zap.Float64("out_seconds", outSeconds),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"inSeconds":  inSeconds,
		"outSeconds": outSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "renderPreviewFiles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to render preview files — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// DeletePreviewFiles deletes all preview/render files for the active sequence.
func (e *Engine) DeletePreviewFiles(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("delete_preview_files")
	result, err := e.premiere.EvalCommand(ctx, "deletePreviewFiles", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to delete preview files — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// ---------------------------------------------------------------------------
// Nesting & Reframing
// ---------------------------------------------------------------------------

// CreateNestedSequence nests the specified clips into a subsequence.
func (e *Engine) CreateNestedSequence(ctx context.Context, trackIndex int, clipIndices []int) (*GenericResult, error) {
	e.logger.Debug("create_nested_sequence",
		zap.Int("track_index", trackIndex),
		zap.Ints("clip_indices", clipIndices),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex":  trackIndex,
		"clipIndices": clipIndices,
	})
	result, err := e.premiere.EvalCommand(ctx, "createNestedSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create nested sequence — make sure the clips exist on the specified track: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// AutoReframeSequence auto-reframes the active sequence to a new aspect ratio.
func (e *Engine) AutoReframeSequence(ctx context.Context, numerator, denominator int, motionPreset string) (*GenericResult, error) {
	e.logger.Debug("auto_reframe_sequence",
		zap.Int("numerator", numerator),
		zap.Int("denominator", denominator),
		zap.String("motion_preset", motionPreset),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"numerator":    numerator,
		"denominator":  denominator,
		"motionPreset": motionPreset,
	})
	result, err := e.premiere.EvalCommand(ctx, "autoReframeSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to auto-reframe sequence — make sure a sequence is active with video content: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// ---------------------------------------------------------------------------
// Generated Media
// ---------------------------------------------------------------------------

// InsertBlackVideo inserts black video onto a track.
func (e *Engine) InsertBlackVideo(ctx context.Context, trackIndex int, startTime, duration float64) (*GenericResult, error) {
	e.logger.Debug("insert_black_video",
		zap.Int("track_index", trackIndex),
		zap.Float64("start_time", startTime),
		zap.Float64("duration", duration),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime":  startTime,
		"duration":   duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "insertBlackVideo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to insert black video — create a sequence first with premiere_create_sequence: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// InsertBarsAndTone inserts bars and tone.
func (e *Engine) InsertBarsAndTone(ctx context.Context, width, height int, duration float64) (*GenericResult, error) {
	e.logger.Debug("insert_bars_and_tone",
		zap.Int("width", width),
		zap.Int("height", height),
		zap.Float64("duration", duration),
	)
	argsJSON, _ := json.Marshal(map[string]any{
		"width":    width,
		"height":   height,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "insertBarsAndTone", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to insert bars and tone — create a sequence first with premiere_create_sequence: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// ---------------------------------------------------------------------------
// Markers
// ---------------------------------------------------------------------------

// GetSequenceMarkers returns all markers on the active sequence.
func (e *Engine) GetSequenceMarkers(ctx context.Context) (*MarkersResult, error) {
	e.logger.Debug("get_sequence_markers")
	result, err := e.premiere.EvalCommand(ctx, "getSequenceMarkers", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to get sequence markers — make sure a sequence is active (try premiere_set_active_sequence): %w", err)
	}
	// ExtendScript returns _ok() => {"success":true,"data":{...}} ; unwrap data for Go structs.
	type extendScriptEnvelope struct {
		Data *MarkersResult `json:"data"`
	}
	var env extendScriptEnvelope
	if err := json.Unmarshal([]byte(result), &env); err != nil {
		return nil, fmt.Errorf("GetSequenceMarkers: could not parse response from Premiere Pro: %w", err)
	}
	if env.Data != nil {
		return env.Data, nil
	}
	var out MarkersResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetSequenceMarkers: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// AddSequenceMarker adds a marker to the active sequence.
func (e *Engine) AddSequenceMarker(ctx context.Context, params *AddMarkerParams) (*GenericResult, error) {
	if params == nil {
		return nil, fmt.Errorf("add_sequence_marker: params must not be nil — provide time and name at minimum")
	}
	e.logger.Debug("add_sequence_marker",
		zap.Float64("time", params.Time),
		zap.String("name", params.Name),
	)
	argsJSON, _ := json.Marshal(params)
	result, err := e.premiere.EvalCommand(ctx, "addSequenceMarker", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to add marker — make sure a sequence is active: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// DeleteSequenceMarker deletes a marker from the active sequence.
func (e *Engine) DeleteSequenceMarker(ctx context.Context, markerIndex int) (*GenericResult, error) {
	e.logger.Debug("delete_sequence_marker", zap.Int("marker_index", markerIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"markerIndex": markerIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteSequenceMarker", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to delete marker at index %d — use premiere_get_sequence_markers to verify the index: %w", markerIndex, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// NavigateToMarker moves the playhead to a specific marker.
func (e *Engine) NavigateToMarker(ctx context.Context, markerIndex int) (*GenericResult, error) {
	e.logger.Debug("navigate_to_marker", zap.Int("marker_index", markerIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"markerIndex": markerIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "navigateToMarker", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to marker at index %d — use premiere_get_sequence_markers to verify the index: %w", markerIndex, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}
