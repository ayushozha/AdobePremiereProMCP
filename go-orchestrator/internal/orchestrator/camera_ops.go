package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Shot/Camera Metadata (1-5)
// ---------------------------------------------------------------------------

// GetClipCameraInfo retrieves camera metadata (make, model, lens, ISO, shutter, aperture) from XMP.
func (e *Engine) GetClipCameraInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipCameraInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipCameraInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipGPSInfo retrieves GPS coordinates from clip XMP metadata.
func (e *Engine) GetClipGPSInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipGPSInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipGPSInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipRecordDate retrieves the recording date/time from clip XMP metadata.
func (e *Engine) GetClipRecordDate(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipRecordDate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipRecordDate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SortClipsByRecordDate sorts clips in a bin by their recording date.
func (e *Engine) SortClipsByRecordDate(ctx context.Context, binPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"binPath": binPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "sortClipsByRecordDate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SortClipsByRecordDate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GroupClipsByCamera groups clips by camera make/model into separate bins.
func (e *Engine) GroupClipsByCamera(ctx context.Context, binPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"binPath": binPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "groupClipsByCamera", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GroupClipsByCamera: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Shot Management (6-10)
// ---------------------------------------------------------------------------

// MarkShotType marks a clip with a shot type (wide, medium, closeup, insert, cutaway).
func (e *Engine) MarkShotType(ctx context.Context, trackType string, trackIndex, clipIndex int, shotType string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"shotType": shotType,
	})
	result, err := e.premiere.EvalCommand(ctx, "markShotType", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MarkShotType: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetShotType retrieves the shot type marker from a clip.
func (e *Engine) GetShotType(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getShotType", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetShotType: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// FilterByShotType returns all clips in a sequence matching a specific shot type.
func (e *Engine) FilterByShotType(ctx context.Context, sequenceIndex int, shotType string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"shotType": shotType,
	})
	result, err := e.premiere.EvalCommand(ctx, "filterByShotType", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FilterByShotType: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateShotList exports a shot list from a sequence to a file.
func (e *Engine) CreateShotList(ctx context.Context, sequenceIndex int, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createShotList", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateShotList: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportShotList imports a shot list from CSV and applies to timeline.
func (e *Engine) ImportShotList(ctx context.Context, csvPath string, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"csvPath": csvPath,
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "importShotList", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportShotList: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Scene/Take Management (11-15)
// ---------------------------------------------------------------------------

// MarkScene marks a clip with a scene number.
func (e *Engine) MarkScene(ctx context.Context, trackType string, trackIndex, clipIndex int, sceneNumber string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"sceneNumber": sceneNumber,
	})
	result, err := e.premiere.EvalCommand(ctx, "markScene", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MarkScene: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// MarkTake marks a clip with a take number.
func (e *Engine) MarkTake(ctx context.Context, trackType string, trackIndex, clipIndex int, takeNumber string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"takeNumber": takeNumber,
	})
	result, err := e.premiere.EvalCommand(ctx, "markTake", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MarkTake: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetBestTake returns the longest/marked-best take for a scene number.
func (e *Engine) GetBestTake(ctx context.Context, sceneNumber string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sceneNumber": sceneNumber,
	})
	result, err := e.premiere.EvalCommand(ctx, "getBestTake", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetBestTake: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// OrganizeByScenesAndTakes auto-organizes project items by scene/take from filename parsing.
func (e *Engine) OrganizeByScenesAndTakes(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "organizeByScenesAndTakes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OrganizeByScenesAndTakes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetSceneList returns all scenes with their takes from the active sequence.
func (e *Engine) GetSceneList(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSceneList", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSceneList: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Camera Matching (16-18)
// ---------------------------------------------------------------------------

// MatchCameraSettings compares camera settings between two project items.
func (e *Engine) MatchCameraSettings(ctx context.Context, clip1Index, clip2Index int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"clip1Index": clip1Index,
		"clip2Index": clip2Index,
	})
	result, err := e.premiere.EvalCommand(ctx, "matchCameraSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MatchCameraSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// FindClipsFromSameCamera finds all clips shot with the same camera as the given clip.
func (e *Engine) FindClipsFromSameCamera(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "findClipsFromSameCamera", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FindClipsFromSameCamera: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateMulticamByCamera creates a multicam sequence grouping clips by camera make/model.
func (e *Engine) CreateMulticamByCamera(ctx context.Context, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "createMulticamByCamera", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateMulticamByCamera: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timecode Management (19-22)
// ---------------------------------------------------------------------------

// GetSourceTimecode retrieves the original source timecode from a project item.
func (e *Engine) GetSourceTimecode(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSourceTimecode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSourceTimecode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetSourceTimecodeOffset sets a source timecode offset on a project item.
func (e *Engine) SetSourceTimecodeOffset(ctx context.Context, projectItemIndex int, offset string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"offset": offset,
	})
	result, err := e.premiere.EvalCommand(ctx, "setSourceTimecodeOffset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetSourceTimecodeOffset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SyncByTimecode syncs clips across specified tracks by aligning their source timecodes.
func (e *Engine) SyncByTimecode(ctx context.Context, trackIndices []int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndices": trackIndices,
	})
	result, err := e.premiere.EvalCommand(ctx, "syncByTimecode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SyncByTimecode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// FindTimecodeBreaks finds gaps in timecode continuity on a track.
func (e *Engine) FindTimecodeBreaks(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "findTimecodeBreaks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FindTimecodeBreaks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Rating (23-26)
// ---------------------------------------------------------------------------

// RateClip sets a rating (1-5 stars) on a clip via XMP metadata.
func (e *Engine) RateClip(ctx context.Context, projectItemIndex, rating int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"rating": rating,
	})
	result, err := e.premiere.EvalCommand(ctx, "rateClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RateClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipRating retrieves the rating from a clip's XMP metadata.
func (e *Engine) GetClipRating(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipRating", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipRating: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// FilterByRating returns all project items with rating >= minRating.
func (e *Engine) FilterByRating(ctx context.Context, minRating int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"minRating": minRating,
	})
	result, err := e.premiere.EvalCommand(ctx, "filterByRating", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FilterByRating: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTopRatedClips returns the top N rated clips sorted by rating descending.
func (e *Engine) GetTopRatedClips(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTopRatedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTopRatedClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Notes (27-30)
// ---------------------------------------------------------------------------

// SetClipNote sets a text note on a clip via XMP dc:description.
func (e *Engine) SetClipNote(ctx context.Context, projectItemIndex int, note string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"note": note,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipNote", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipNote: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipNote retrieves the clip note from XMP dc:description.
func (e *Engine) GetClipNote(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipNote", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipNote: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SearchClipNotes searches all clip notes for a text string.
func (e *Engine) SearchClipNotes(ctx context.Context, searchText string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"searchText": searchText,
	})
	result, err := e.premiere.EvalCommand(ctx, "searchClipNotes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SearchClipNotes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportClipNotes exports all clip notes as CSV or JSON to a file.
func (e *Engine) ExportClipNotes(ctx context.Context, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportClipNotes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportClipNotes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
