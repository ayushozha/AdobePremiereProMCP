package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Batch Import
// ---------------------------------------------------------------------------

func (e *Engine) BatchImportWithMetadata(ctx context.Context, itemsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"itemsJSON": itemsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchImportWithMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchImportWithMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ImportImageSequence(ctx context.Context, folderPath string, fps float64, targetBin string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"folderPath": folderPath,
		"fps": fps,
		"targetBin": targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importImageSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportImageSequence: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Batch Export
// ---------------------------------------------------------------------------

func (e *Engine) BatchExportSequences(ctx context.Context, sequenceIndices []int, outputDir, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndices": sequenceIndices,
		"outputDir": outputDir,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchExportSequences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchExportSequences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExportAllSequences(ctx context.Context, outputDir, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputDir": outputDir,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportAllSequences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportAllSequences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Batch Effects
// ---------------------------------------------------------------------------

func (e *Engine) ApplyEffectToMultipleClips(ctx context.Context, trackType string, trackIndex int, clipIndices []int, effectName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndices": clipIndices,
		"effectName": effectName,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyEffectToMultipleClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyEffectToMultipleClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RemoveAllEffects(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "removeAllEffects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemoveAllEffects: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyTransitionToAllCuts(ctx context.Context, trackIndex int, transitionName string, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"transitionName": transitionName,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyTransitionToAllCuts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyTransitionToAllCuts: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Batch Color
// ---------------------------------------------------------------------------

func (e *Engine) ApplyLUTToAllClips(ctx context.Context, trackIndex int, lutPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"lutPath": lutPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyLUTToAllClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyLUTToAllClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ResetColorOnAllClips(ctx context.Context, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "resetColorOnAllClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetColorOnAllClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Batch Audio
// ---------------------------------------------------------------------------

func (e *Engine) NormalizeAllAudio(ctx context.Context, targetDB float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"targetDB": targetDB,
	})
	result, err := e.premiere.EvalCommand(ctx, "normalizeAllAudio", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("NormalizeAllAudio: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) MuteAllAudioTracks(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "muteAllAudioTracks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MuteAllAudioTracks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) UnmuteAllAudioTracks(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "unmuteAllAudioTracks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("UnmuteAllAudioTracks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Conforming
// ---------------------------------------------------------------------------

func (e *Engine) ConformSequenceToClip(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "conformSequenceToClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConformSequenceToClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ScaleAllClipsToFrame(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "scaleAllClipsToFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ScaleAllClipsToFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Operations
// ---------------------------------------------------------------------------

func (e *Engine) SelectAllClipsOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "selectAllClipsOnTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SelectAllClipsOnTrack: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SelectAllClipsBetween(ctx context.Context, startSeconds, endSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"startSeconds": startSeconds,
		"endSeconds": endSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "selectAllClipsBetween", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SelectAllClipsBetween: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DeleteAllClipsBetween(ctx context.Context, trackType string, trackIndex int, startSeconds, endSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"startSeconds": startSeconds,
		"endSeconds": endSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteAllClipsBetween", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DeleteAllClipsBetween: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RippleDeleteAllGaps(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "rippleDeleteAllGaps", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RippleDeleteAllGaps: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Cleanup
// ---------------------------------------------------------------------------

func (e *Engine) RemoveUnusedMedia(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "removeUnusedMedia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemoveUnusedMedia: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetUnusedMedia(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getUnusedMedia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUnusedMedia: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FlattenAllBins(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "flattenAllBins", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FlattenAllBins: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AutoOrganizeBins(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "autoOrganizeBins", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AutoOrganizeBins: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Markers Batch
// ---------------------------------------------------------------------------

func (e *Engine) ExportMarkersAsCSV(ctx context.Context, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportMarkersAsCSV", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportMarkersAsCSV: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExportMarkersAsEDL(ctx context.Context, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportMarkersAsEDL", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportMarkersAsEDL: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ImportMarkersFromCSV(ctx context.Context, csvPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"csvPath": csvPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "importMarkersFromCSV", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportMarkersFromCSV: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DeleteAllMarkers(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "deleteAllMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DeleteAllMarkers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ConvertMarkersToClips(ctx context.Context, markerColor string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"markerColor": markerColor,
	})
	result, err := e.premiere.EvalCommand(ctx, "convertMarkersToClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConvertMarkersToClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Automation
// ---------------------------------------------------------------------------

func (e *Engine) RunExtendScript(ctx context.Context, script string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script": script,
	})
	result, err := e.premiere.EvalCommand(ctx, "runExtendScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RunExtendScript: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSystemInfo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSystemInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSystemInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetRecentProjects(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getRecentProjects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetRecentProjects: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
