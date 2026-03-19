package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Advanced Trimming
// ---------------------------------------------------------------------------

func (e *Engine) RippleTrim(ctx context.Context, trackType string, trackIndex, clipIndex int, trimEnd bool, deltaSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"trimEnd": trimEnd,
		"deltaSeconds": deltaSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "rippleTrim", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RippleTrim: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RollTrim(ctx context.Context, trackType string, trackIndex, clipIndex int, deltaSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"deltaSeconds": deltaSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "rollTrim", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RollTrim: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SlipClip(ctx context.Context, trackType string, trackIndex, clipIndex int, deltaSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"deltaSeconds": deltaSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "slipClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SlipClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SlideClip(ctx context.Context, trackType string, trackIndex, clipIndex int, deltaSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"deltaSeconds": deltaSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "slideClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SlideClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Paste Operations
// ---------------------------------------------------------------------------

func (e *Engine) PasteInsert(ctx context.Context, trackType string, trackIndex int, time float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"time": time,
	})
	result, err := e.premiere.EvalCommand(ctx, "pasteInsert", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PasteInsert: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) PasteAttributes(ctx context.Context, srcTrackType string, srcTrackIndex, srcClipIndex int, destTrackType string, destTrackIndex, destClipIndex int, attributes string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"srcTrackType": srcTrackType,
		"srcTrackIndex": srcTrackIndex,
		"srcClipIndex": srcClipIndex,
		"destTrackType": destTrackType,
		"destTrackIndex": destTrackIndex,
		"destClipIndex": destClipIndex,
		"attributes": attributes,
	})
	result, err := e.premiere.EvalCommand(ctx, "pasteAttributes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PasteAttributes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Match Frame
// ---------------------------------------------------------------------------

func (e *Engine) MatchFrame(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "matchFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MatchFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ReverseMatchFrame(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "reverseMatchFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReverseMatchFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Lift & Extract
// ---------------------------------------------------------------------------

func (e *Engine) LiftSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "liftSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LiftSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExtractSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "extractSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExtractSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Gap Management
// ---------------------------------------------------------------------------

func (e *Engine) FindGaps(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "findGaps", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FindGaps: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CloseGap(ctx context.Context, trackType string, trackIndex, gapIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"gapIndex": gapIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "closeGap", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CloseGap: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CloseAllGaps(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "closeAllGaps", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CloseAllGaps: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RippleDeleteGap(ctx context.Context, trackType string, trackIndex int, startTime, endTime float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"startTime": startTime,
		"endTime": endTime,
	})
	result, err := e.premiere.EvalCommand(ctx, "rippleDeleteGap", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RippleDeleteGap: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Grouping
// ---------------------------------------------------------------------------

func (e *Engine) GroupClips(ctx context.Context, clipRefsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"clipRefsJSON": clipRefsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "groupClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GroupClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) UngroupClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "ungroupClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("UngroupClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetGroupedClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getGroupedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetGroupedClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Snap & Alignment
// ---------------------------------------------------------------------------

func (e *Engine) SetSnapping(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setSnapping", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetSnapping: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSnapping(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSnapping", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSnapping: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Zoom
// ---------------------------------------------------------------------------

func (e *Engine) ZoomToFitTimeline(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "zoomToFitTimeline", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ZoomToFitTimeline: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ZoomToSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "zoomToSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ZoomToSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTimelineZoom(ctx context.Context, level float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"level": level,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTimelineZoom", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTimelineZoom: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Navigation
// ---------------------------------------------------------------------------

func (e *Engine) GoToNextEditPoint(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "goToNextEditPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToNextEditPoint: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GoToPreviousEditPoint(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "goToPreviousEditPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToPreviousEditPoint: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GoToNextClip(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "goToNextClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToNextClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GoToPreviousClip(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "goToPreviousClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToPreviousClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GoToSequenceStart(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "goToSequenceStart", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToSequenceStart: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GoToSequenceEnd(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "goToSequenceEnd", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToSequenceEnd: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Markers
// ---------------------------------------------------------------------------

func (e *Engine) AddClipMarker(ctx context.Context, trackType string, trackIndex, clipIndex int, time float64, name, comment string, colorIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"time": time,
		"name": name,
		"comment": comment,
		"colorIndex": colorIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "addClipMarker", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddClipMarker: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetClipMarkers(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipMarkers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DeleteClipMarker(ctx context.Context, trackType string, trackIndex, clipIndex, markerIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"markerIndex": markerIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteClipMarker", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DeleteClipMarker: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
