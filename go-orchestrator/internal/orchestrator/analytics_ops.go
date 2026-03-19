package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Project Analytics (1-7)
// ---------------------------------------------------------------------------

// GetProjectSummary returns a comprehensive project summary.
func (e *Engine) GetProjectSummary(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectSummary", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectSummary: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetMediaTypeBreakdown returns a breakdown of media by type.
func (e *Engine) GetMediaTypeBreakdown(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMediaTypeBreakdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaTypeBreakdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetCodecBreakdown returns a breakdown by codec across all media.
func (e *Engine) GetCodecBreakdown(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getCodecBreakdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetCodecBreakdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetResolutionBreakdown returns a breakdown by resolution across all media.
func (e *Engine) GetResolutionBreakdown(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getResolutionBreakdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetResolutionBreakdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetFrameRateBreakdown returns a breakdown by frame rate across all media.
func (e *Engine) GetFrameRateBreakdown(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getFrameRateBreakdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFrameRateBreakdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetDurationDistribution returns the distribution of clip durations.
func (e *Engine) GetDurationDistribution(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getDurationDistribution", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDurationDistribution: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetColorSpaceBreakdown returns a breakdown by color space.
func (e *Engine) GetColorSpaceBreakdown(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getColorSpaceBreakdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetColorSpaceBreakdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Sequence Analytics (8-14)
// ---------------------------------------------------------------------------

// GetSequenceSummary returns a summary of a specific sequence.
func (e *Engine) GetSequenceSummary(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceSummary", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceSummary: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetEffectsUsageReport returns all effects used in a sequence with counts.
func (e *Engine) GetEffectsUsageReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEffectsUsageReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEffectsUsageReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTransitionsUsageReport returns all transitions used in a sequence.
func (e *Engine) GetTransitionsUsageReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTransitionsUsageReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTransitionsUsageReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTrackUtilizationReport returns track utilization percentages.
func (e *Engine) GetTrackUtilizationReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTrackUtilizationReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTrackUtilizationReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetEditPointDensity returns edit points per minute for a sequence.
func (e *Engine) GetEditPointDensity(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEditPointDensity", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEditPointDensity: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetPacingReport returns average clip duration and cuts per minute.
func (e *Engine) GetPacingReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getPacingReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPacingReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAudioLevelsReport returns audio level statistics across a sequence.
func (e *Engine) GetAudioLevelsReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getAudioLevelsReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAudioLevelsReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Reports (15-19)
// ---------------------------------------------------------------------------

// GetClipSourceReport reports which source files are used where.
func (e *Engine) GetClipSourceReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipSourceReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipSourceReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTimelineStructureReport returns the structural layout of a timeline.
func (e *Engine) GetTimelineStructureReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineStructureReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineStructureReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetGapAnalysisReport returns a detailed gap analysis for a sequence.
func (e *Engine) GetGapAnalysisReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getGapAnalysisReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetGapAnalysisReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetDuplicateClipsReport finds duplicate clips in a timeline.
func (e *Engine) GetDuplicateClipsReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getDuplicateClipsReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDuplicateClipsReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetUnusedTracksReport identifies empty or unused tracks.
func (e *Engine) GetUnusedTracksReport(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getUnusedTracksReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUnusedTracksReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Export Reports (20-24)
// ---------------------------------------------------------------------------

// ExportProjectReport exports a comprehensive project report.
func (e *Engine) ExportProjectReport(ctx context.Context, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportProjectReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportProjectReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportTimelineAsText exports a timeline as human-readable text.
func (e *Engine) ExportTimelineAsText(ctx context.Context, sequenceIndex int, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportTimelineAsText", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportTimelineAsText: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportClipList exports a clip list from a sequence.
func (e *Engine) ExportClipList(ctx context.Context, sequenceIndex int, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportClipList", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportClipList: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportEffectsList exports the effects list from a sequence.
func (e *Engine) ExportEffectsList(ctx context.Context, sequenceIndex int, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportEffectsList", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportEffectsList: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportMediaList exports all media files in the project.
func (e *Engine) ExportMediaList(ctx context.Context, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportMediaList", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportMediaList: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Comparison (25-26)
// ---------------------------------------------------------------------------

// CompareSequences compares two sequences.
func (e *Engine) CompareSequences(ctx context.Context, seqIndex1, seqIndex2 int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seqIndex1": seqIndex1,
		"seqIndex2": seqIndex2,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareSequences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareSequences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CompareClips compares two clips by their track type, track index, and clip index.
func (e *Engine) CompareClips(ctx context.Context, clip1TrackType string, clip1TrackIndex, clip1ClipIndex int, clip2TrackType string, clip2TrackIndex, clip2ClipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"clip1TrackType": clip1TrackType,
		"clip1TrackIndex": clip1TrackIndex,
		"clip1ClipIndex": clip1ClipIndex,
		"clip2TrackType": clip2TrackType,
		"clip2TrackIndex": clip2TrackIndex,
		"clip2ClipIndex": clip2ClipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Usage Statistics (27-30)
// ---------------------------------------------------------------------------

// GetEditingSessionStats returns current editing session statistics.
func (e *Engine) GetEditingSessionStats(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getEditingSessionStats", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEditingSessionStats: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetProjectAgeInfo returns project creation date, last modified, and edit sessions.
func (e *Engine) GetProjectAgeInfo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectAgeInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectAgeInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetStorageReport returns a storage usage report.
func (e *Engine) GetStorageReport(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getStorageReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetStorageReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetPerformanceReport2 returns a performance report with render times and playback stats.
func (e *Engine) GetPerformanceReport2(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getPerformanceReport2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceReport2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
