package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Review & Collaboration
// ---------------------------------------------------------------------------

// AddReviewComment adds a review comment as a marker with metadata at the given time.
func (e *Engine) AddReviewComment(ctx context.Context, time float64, text, author string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"time": time,
		"text": text,
		"author": author,
	})
	result, err := e.premiere.EvalCommand(ctx, "addReviewComment", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddReviewComment: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetReviewComments returns all review comments (markers with comment data).
func (e *Engine) GetReviewComments(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getReviewComments", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetReviewComments: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ResolveReviewComment marks a review comment (by marker index) as resolved.
func (e *Engine) ResolveReviewComment(ctx context.Context, markerIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"markerIndex": markerIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "resolveReviewComment", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResolveReviewComment: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetUnresolvedComments returns all unresolved review comments.
func (e *Engine) GetUnresolvedComments(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getUnresolvedComments", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUnresolvedComments: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportReviewReport exports a review report in the given format (json, csv, html).
func (e *Engine) ExportReviewReport(ctx context.Context, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportReviewReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportReviewReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Version Control
// ---------------------------------------------------------------------------

// GetProjectVersionHistory returns auto-save versions with timestamps.
func (e *Engine) GetProjectVersionHistory(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectVersionHistory", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectVersionHistory: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RevertToVersion opens a specific auto-save version.
func (e *Engine) RevertToVersion(ctx context.Context, versionPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"versionPath": versionPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "revertToVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RevertToVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateSnapshot saves the project as a named snapshot with description.
func (e *Engine) CreateSnapshot(ctx context.Context, name, description string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"description": description,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSnapshot", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSnapshot: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CompareSnapshots compares two project snapshots and returns a basic diff.
func (e *Engine) CompareSnapshots(ctx context.Context, snapshot1, snapshot2 string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"snapshot1": snapshot1,
		"snapshot2": snapshot2,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareSnapshots", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareSnapshots: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// EDL/XML Interchange
// ---------------------------------------------------------------------------

// ImportEDL imports an EDL file into the project.
func (e *Engine) ImportEDL(ctx context.Context, edlPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"edlPath": edlPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "importEDL", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportEDL: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportAAF imports an AAF file into the project.
func (e *Engine) ImportAAF(ctx context.Context, aafPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"aafPath": aafPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "importAAF", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportAAF: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportFCPXML imports a Final Cut Pro XML file into the project.
func (e *Engine) ImportFCPXML(ctx context.Context, xmlPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"xmlPath": xmlPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "importFCPXML", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportFCPXML: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportXMLTimeline imports a Premiere XML timeline into the project.
func (e *Engine) ImportXMLTimeline(ctx context.Context, xmlPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"xmlPath": xmlPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "importXMLTimeline", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportXMLTimeline: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportEDLFile exports the active sequence as an EDL file (CMX 3600).
func (e *Engine) ExportEDLFile(ctx context.Context, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportEDLFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportEDLFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportProjectSnapshot exports the project as a portable snapshot.
func (e *Engine) ExportProjectSnapshot(ctx context.Context, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportProjectSnapshot", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportProjectSnapshot: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Collaboration Metadata
// ---------------------------------------------------------------------------

// SetEditorialNote sets an editorial note on a clip.
func (e *Engine) SetEditorialNote(ctx context.Context, trackType string, trackIndex, clipIndex int, note string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"note": note,
	})
	result, err := e.premiere.EvalCommand(ctx, "setEditorialNote", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetEditorialNote: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetEditorialNotes returns all editorial notes in the active sequence.
func (e *Engine) GetEditorialNotes(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getEditorialNotes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEditorialNotes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClearEditorialNotes clears all editorial notes from the active sequence.
func (e *Engine) ClearEditorialNotes(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearEditorialNotes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearEditorialNotes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// TagClipForReview tags a clip with a review status (approved, needs-changes, rejected).
func (e *Engine) TagClipForReview(ctx context.Context, trackType string, trackIndex, clipIndex int, reviewType string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"reviewType": reviewType,
	})
	result, err := e.premiere.EvalCommand(ctx, "tagClipForReview", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TagClipForReview: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipReviewStatus returns the review status for a clip.
func (e *Engine) GetClipReviewStatus(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipReviewStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipReviewStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Change Tracking
// ---------------------------------------------------------------------------

// GetSequenceChangeLog returns recent changes to the active sequence.
func (e *Engine) GetSequenceChangeLog(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceChangeLog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceChangeLog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetProjectActivity returns a recent project activity summary.
func (e *Engine) GetProjectActivity(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectActivity", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectActivity: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetLastModifiedClips returns the N most recently modified clips.
func (e *Engine) GetLastModifiedClips(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "getLastModifiedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLastModifiedClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Delivery Checklist
// ---------------------------------------------------------------------------

// CheckAudioLevels checks all audio levels meet the target LUFS.
func (e *Engine) CheckAudioLevels(ctx context.Context, targetLUFS, tolerance float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"targetLUFS": targetLUFS,
		"tolerance": tolerance,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkAudioLevels", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckAudioLevels: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckFrameRate verifies the sequence frame rate matches the target.
func (e *Engine) CheckFrameRate(ctx context.Context, targetFPS float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"targetFPS": targetFPS,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkFrameRate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckFrameRate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckResolution verifies the sequence resolution matches the target.
func (e *Engine) CheckResolution(ctx context.Context, targetWidth, targetHeight int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"targetWidth": targetWidth,
		"targetHeight": targetHeight,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkResolution", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckResolution: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckDuration checks the sequence duration is within the given range.
func (e *Engine) CheckDuration(ctx context.Context, minSeconds, maxSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"minSeconds": minSeconds,
		"maxSeconds": maxSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GenerateDeliveryReport generates a full delivery compliance report.
func (e *Engine) GenerateDeliveryReport(ctx context.Context, specsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"specsJSON": specsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateDeliveryReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateDeliveryReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckForBlackFrames detects black frames in the active sequence.
func (e *Engine) CheckForBlackFrames(ctx context.Context, thresholdFrames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"thresholdFrames": thresholdFrames,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkForBlackFrames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckForBlackFrames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckForFlashContent detects rapid luminance changes for epilepsy safety.
func (e *Engine) CheckForFlashContent(ctx context.Context, threshold float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"threshold": threshold,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkForFlashContent", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckForFlashContent: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
