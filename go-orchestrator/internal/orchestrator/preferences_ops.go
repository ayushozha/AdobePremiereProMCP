package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// General Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetGeneralPreferences(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getGeneralPreferences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetGeneralPreferences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetDefaultStillDuration(ctx context.Context, frames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frames": frames,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDefaultStillDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDefaultStillDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetDefaultTransitionDuration(ctx context.Context, seconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDefaultTransitionDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDefaultTransitionDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetDefaultAudioTransitionDuration(ctx context.Context, seconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDefaultAudioTransitionDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDefaultAudioTransitionDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Appearance
// ---------------------------------------------------------------------------

func (e *Engine) GetBrightness(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getBrightness", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetBrightness: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetBrightness(ctx context.Context, level int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"level": level,
	})
	result, err := e.premiere.EvalCommand(ctx, "setBrightness", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetBrightness: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Auto Save
// ---------------------------------------------------------------------------

func (e *Engine) SetAutoSaveEnabled(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setAutoSaveEnabled", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAutoSaveEnabled: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetAutoSaveMaxVersions(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "setAutoSaveMaxVersions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAutoSaveMaxVersions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetAutoSaveLocation(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAutoSaveLocation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAutoSaveLocation: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Playback Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetPlaybackResolution(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getPlaybackResolution", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPlaybackResolution: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetPlaybackResolution(ctx context.Context, quality string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"quality": quality,
	})
	result, err := e.premiere.EvalCommand(ctx, "setPlaybackResolution", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetPlaybackResolution: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetPrerollFrames(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getPrerollFrames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPrerollFrames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetPrerollFrames(ctx context.Context, frames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frames": frames,
	})
	result, err := e.premiere.EvalCommand(ctx, "setPrerollFrames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetPrerollFrames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetPostrollFrames(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getPostrollFrames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPostrollFrames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetPostrollFrames(ctx context.Context, frames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frames": frames,
	})
	result, err := e.premiere.EvalCommand(ctx, "setPostrollFrames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetPostrollFrames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetTimelineSettings(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTimeDisplayFormat(ctx context.Context, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTimeDisplayFormat", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTimeDisplayFormat: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetVideoTransitionDefaultDuration(ctx context.Context, frames int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"frames": frames,
	})
	result, err := e.premiere.EvalCommand(ctx, "setVideoTransitionDefaultDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetVideoTransitionDefaultDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Media Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetMediaCacheSettings(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMediaCacheSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaCacheSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetMediaCacheLocation(ctx context.Context, path string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"path": path,
	})
	result, err := e.premiere.EvalCommand(ctx, "setMediaCacheLocation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetMediaCacheLocation: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetMediaCacheSize(ctx context.Context, maxGB float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"maxGB": maxGB,
	})
	result, err := e.premiere.EvalCommand(ctx, "setMediaCacheSize", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetMediaCacheSize: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CleanMediaCacheOlderThan(ctx context.Context, days int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"days": days,
	})
	result, err := e.premiere.EvalCommand(ctx, "cleanMediaCacheOlderThan", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CleanMediaCacheOlderThan: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Label Colors
// ---------------------------------------------------------------------------

func (e *Engine) GetLabelColorNames(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getLabelColorNames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLabelColorNames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetLabelColorName(ctx context.Context, index int, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"index": index,
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLabelColorName", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLabelColorName: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// GPU / Renderer
// ---------------------------------------------------------------------------

func (e *Engine) GetRendererInfo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getRendererInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetRendererInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetGPUInfo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getGPUInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetGPUInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetRenderer(ctx context.Context, rendererName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"rendererName": rendererName,
	})
	result, err := e.premiere.EvalCommand(ctx, "setRenderer", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetRenderer: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Defaults
// ---------------------------------------------------------------------------

func (e *Engine) GetDefaultSequencePresets(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getDefaultSequencePresets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDefaultSequencePresets: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetDefaultSequencePreset(ctx context.Context, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDefaultSequencePreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDefaultSequencePreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetInstalledCodecs(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getInstalledCodecs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetInstalledCodecs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
