package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Playback Control
// ---------------------------------------------------------------------------

func (e *Engine) Play(ctx context.Context, speed float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"speed": speed,
	})
	result, err := e.premiere.EvalCommand(ctx, "play", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("Play: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) Pause(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "pause", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("Pause: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) Stop(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "stop", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("Stop: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) StepForward(ctx context.Context, frames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frames": frames,
	})
	result, err := e.premiere.EvalCommand(ctx, "stepForward", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StepForward: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) StepBackward(ctx context.Context, frames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frames": frames,
	})
	result, err := e.premiere.EvalCommand(ctx, "stepBackward", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StepBackward: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShuttleForward(ctx context.Context, speed float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"speed": speed,
	})
	result, err := e.premiere.EvalCommand(ctx, "shuttleForward", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShuttleForward: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShuttleBackward(ctx context.Context, speed float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"speed": speed,
	})
	result, err := e.premiere.EvalCommand(ctx, "shuttleBackward", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShuttleBackward: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) TogglePlayPause(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "togglePlayPause", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TogglePlayPause: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) PlayInToOut(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "playInToOut", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PlayInToOut: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) LoopPlayback(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "loopPlayback", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LoopPlayback: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Program Monitor
// ---------------------------------------------------------------------------

func (e *Engine) GetProgramMonitorZoom(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProgramMonitorZoom", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProgramMonitorZoom: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetProgramMonitorZoom(ctx context.Context, percent float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"percent": percent,
	})
	result, err := e.premiere.EvalCommand(ctx, "setProgramMonitorZoom", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetProgramMonitorZoom: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FitProgramMonitor(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "fitProgramMonitor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FitProgramMonitor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ToggleSafeMargins(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "toggleSafeMargins", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ToggleSafeMargins: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetFrameAtPlayhead(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getFrameAtPlayhead", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFrameAtPlayhead: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Sequence Navigation (extended)
// ---------------------------------------------------------------------------

func (e *Engine) GoToTimecode(ctx context.Context, timecode string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"timecode": timecode,
	})
	result, err := e.premiere.EvalCommand(ctx, "goToTimecode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToTimecode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GoToFrame(ctx context.Context, frameNumber int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frameNumber": frameNumber,
	})
	result, err := e.premiere.EvalCommand(ctx, "goToFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GoToFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSequenceDuration(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetFrameCount(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getFrameCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFrameCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetCurrentTimecode(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getCurrentTimecode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetCurrentTimecode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Selection & Focus
// ---------------------------------------------------------------------------

func (e *Engine) SelectClipsInRange(ctx context.Context, startSeconds, endSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"startSeconds": startSeconds,
		"endSeconds": endSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "selectClipsInRange", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SelectClipsInRange: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SelectAllOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "selectAllOnTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SelectAllOnTrack: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) InvertSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "invertSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("InvertSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSelectionRange(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSelectionRange", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSelectionRange: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Render Status
// ---------------------------------------------------------------------------

func (e *Engine) GetRenderStatus(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getRenderStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetRenderStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) IsRendering(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "isRendering", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsRendering: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Sequence Metadata
// ---------------------------------------------------------------------------

func (e *Engine) GetSequenceMetadata(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetSequenceMetadata(ctx context.Context, key, value string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"key": key,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setSequenceMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetSequenceMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSequenceColorSpace(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceColorSpace", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceColorSpace: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetSequenceColorSpace(ctx context.Context, colorSpace string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"colorSpace": colorSpace,
	})
	result, err := e.premiere.EvalCommand(ctx, "setSequenceColorSpace", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetSequenceColorSpace: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
