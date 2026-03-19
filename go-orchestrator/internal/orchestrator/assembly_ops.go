package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Timeline Assembly Operations
// ---------------------------------------------------------------------------

func (e *Engine) AssembleFromEDL(ctx context.Context, edlJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"edlJSON": edlJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "assembleFromEDL", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AssembleFromEDL: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AssembleFromCSV(ctx context.Context, csvPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"csvPath": csvPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "assembleFromCSV", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AssembleFromCSV: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AssembleFromFolderOrder(ctx context.Context, folderPath, transitionName string, transitionDuration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"folderPath": folderPath,
		"transitionName": transitionName,
		"transitionDuration": transitionDuration,
	})
	result, err := e.premiere.EvalCommand(ctx, "assembleFromFolderOrder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AssembleFromFolderOrder: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) InterleaveClips(ctx context.Context, trackIndexA, trackIndexB int, transitionDuration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndexA": trackIndexA,
		"trackIndexB": trackIndexB,
		"transitionDuration": transitionDuration,
	})
	result, err := e.premiere.EvalCommand(ctx, "interleaveClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("InterleaveClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShuffleClips(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "shuffleClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShuffleClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Arrangement Operations
// ---------------------------------------------------------------------------

func (e *Engine) SortClipsByDuration(ctx context.Context, trackType string, trackIndex int, ascending bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"ascending": ascending,
	})
	result, err := e.premiere.EvalCommand(ctx, "sortClipsByDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SortClipsByDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SortClipsByName(ctx context.Context, trackType string, trackIndex int, ascending bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"ascending": ascending,
	})
	result, err := e.premiere.EvalCommand(ctx, "sortClipsByName", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SortClipsByName: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SortClipsByFileName(ctx context.Context, trackType string, trackIndex int, ascending bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"ascending": ascending,
	})
	result, err := e.premiere.EvalCommand(ctx, "sortClipsByFileName", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SortClipsByFileName: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ReverseClipOrder(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "reverseClipOrder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReverseClipOrder: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DistributeClipsEvenly(ctx context.Context, trackType string, trackIndex int, totalDuration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"totalDuration": totalDuration,
	})
	result, err := e.premiere.EvalCommand(ctx, "distributeClipsEvenly", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DistributeClipsEvenly: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) StackClips(ctx context.Context, trackType string, trackIndex int, startTime float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"startTime": startTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "stackClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StackClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Multi-Track Composition Operations
// ---------------------------------------------------------------------------

func (e *Engine) CreateOverlayTrack(ctx context.Context, sourceTrack, destTrack int, opacity float64, blendMode string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sourceTrack": sourceTrack,
		"destTrack": destTrack,
		"opacity": opacity,
		"blendMode": blendMode,
	})
	result, err := e.premiere.EvalCommand(ctx, "createOverlayTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateOverlayTrack: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CreateGreenScreenComposite(ctx context.Context, fgTrackIndex, fgClipIndex, bgTrackIndex, bgClipIndex int, keyColor string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"fgTrackIndex": fgTrackIndex,
		"fgClipIndex": fgClipIndex,
		"bgTrackIndex": bgTrackIndex,
		"bgClipIndex": bgClipIndex,
		"keyColor": keyColor,
	})
	result, err := e.premiere.EvalCommand(ctx, "createGreenScreenComposite", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateGreenScreenComposite: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CreatePictureInPictureGrid(ctx context.Context, trackIndices []int, layout string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndices": trackIndices,
		"layout": layout,
	})
	result, err := e.premiere.EvalCommand(ctx, "createPictureInPictureGrid", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreatePictureInPictureGrid: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) LayerTracks(ctx context.Context, baseTrack int, overlayTracks []int, opacities []float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"baseTrack": baseTrack,
		"overlayTracks": overlayTracks,
		"opacities": opacities,
	})
	result, err := e.premiere.EvalCommand(ctx, "layerTracks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LayerTracks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Generation Operations
// ---------------------------------------------------------------------------

func (e *Engine) GenerateBlackClip(ctx context.Context, trackIndex int, startTime, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateBlackClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateBlackClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GenerateColorClip(ctx context.Context, trackIndex int, startTime, duration float64, color string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"color": color,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateColorClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateColorClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GenerateGradientClip(ctx context.Context, trackIndex int, startTime, duration float64, colorStart, colorEnd, direction string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"colorStart": colorStart,
		"colorEnd": colorEnd,
		"direction": direction,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateGradientClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateGradientClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GenerateTestPattern(ctx context.Context, trackIndex int, startTime, duration float64, pattern string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"pattern": pattern,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateTestPattern", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateTestPattern: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GenerateSilence(ctx context.Context, trackIndex int, startTime, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateSilence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateSilence: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GenerateTone(ctx context.Context, trackIndex int, startTime, duration, frequency, amplitude float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"frequency": frequency,
		"amplitude": amplitude,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateTone", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateTone: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Duplication Operations
// ---------------------------------------------------------------------------

func (e *Engine) DuplicateTimelineSection(ctx context.Context, startTime, endTime, destTime float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"startTime": startTime,
		"endTime": endTime,
		"destTime": destTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "duplicateTimelineSection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DuplicateTimelineSection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RepeatTimelineSection(ctx context.Context, startTime, endTime float64, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"startTime": startTime,
		"endTime": endTime,
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "repeatTimelineSection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RepeatTimelineSection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) MirrorTimeline(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "mirrorTimeline", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MirrorTimeline: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SplitTimelineAtPlayhead(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "splitTimelineAtPlayhead", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SplitTimelineAtPlayhead: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Analysis Operations
// ---------------------------------------------------------------------------

func (e *Engine) GetTimelineGapReport(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineGapReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineGapReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTimelineConflictReport(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineConflictReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineConflictReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTimelineEffectsReport(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineEffectsReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineEffectsReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTimelineDurationBreakdown(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineDurationBreakdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineDurationBreakdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTimelineTrackUsageReport(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineTrackUsageReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineTrackUsageReport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
