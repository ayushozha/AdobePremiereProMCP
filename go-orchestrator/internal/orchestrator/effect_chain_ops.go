package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Effect Chain Management Operations
// ---------------------------------------------------------------------------

func (e *Engine) GetEffectChain(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEffectChain", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEffectChain: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ReorderEffect(ctx context.Context, trackType string, trackIndex, clipIndex, fromIndex, toIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"fromIndex": fromIndex,
		"toIndex": toIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "reorderEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReorderEffect: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetEffectCount(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEffectCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEffectCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ClearAllEffects(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "clearAllEffects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearAllEffects: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DuplicateEffect(ctx context.Context, trackType string, trackIndex, clipIndex, effectIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"effectIndex": effectIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "duplicateEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DuplicateEffect: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Effect Parameter Animation Operations
// ---------------------------------------------------------------------------

func (e *Engine) AnimateEffectParameter(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, keyframesJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"componentIndex": componentIndex,
		"paramIndex": paramIndex,
		"keyframesJSON": keyframesJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "animateEffectParameter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AnimateEffectParameter: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetEffectParameterRange(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"componentIndex": componentIndex,
		"paramIndex": paramIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEffectParameterRange", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEffectParameterRange: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ResetEffectParameter(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"componentIndex": componentIndex,
		"paramIndex": paramIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "resetEffectParameter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetEffectParameter: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) LinkEffectParameters(ctx context.Context, trackType string, trackIndex, clipIndex, comp1, param1, comp2, param2 int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"comp1": comp1,
		"param1": param1,
		"comp2": comp2,
		"param2": param2,
	})
	result, err := e.premiere.EvalCommand(ctx, "linkEffectParameters", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LinkEffectParameters: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetEffectRenderOrder(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEffectRenderOrder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEffectRenderOrder: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Common Effect Presets Operations
// ---------------------------------------------------------------------------

func (e *Engine) ApplyBlackAndWhite(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyBlackAndWhite", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyBlackAndWhite: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplySepia(ctx context.Context, trackIndex, clipIndex int, intensity float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"intensity": intensity,
	})
	result, err := e.premiere.EvalCommand(ctx, "applySepia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplySepia: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyVintageFilm(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyVintageFilm", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyVintageFilm: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyFilmGrain(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyFilmGrain", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyFilmGrain: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyVignetteEffect(ctx context.Context, trackIndex, clipIndex int, amount, feather float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
		"feather": feather,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyVignetteEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyVignetteEffect: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyGlow(ctx context.Context, trackIndex, clipIndex int, intensity, radius float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"intensity": intensity,
		"radius": radius,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyGlow", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyGlow: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyDropShadow(ctx context.Context, trackIndex, clipIndex int, opacity, distance, softness, direction float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"opacity": opacity,
		"distance": distance,
		"softness": softness,
		"direction": direction,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyDropShadow", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyDropShadow: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyStroke(ctx context.Context, trackIndex, clipIndex int, color string, width float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"color": color,
		"width": width,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyStroke", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyStroke: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyCinematicBars(ctx context.Context, trackIndex, clipIndex int, barHeight float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"barHeight": barHeight,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyCinematicBars", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyCinematicBars: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyFlipHorizontal(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyFlipHorizontal", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyFlipHorizontal: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Transition Effects Operations
// ---------------------------------------------------------------------------

func (e *Engine) GetDurationOfTransition(ctx context.Context, trackType string, trackIndex, transitionIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"transitionIndex": transitionIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getDurationOfTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDurationOfTransition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTransitionDuration(ctx context.Context, trackType string, trackIndex, transitionIndex int, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"transitionIndex": transitionIndex,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTransitionDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTransitionDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTransitionAlignment(ctx context.Context, trackType string, trackIndex, transitionIndex int, alignment string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"transitionIndex": transitionIndex,
		"alignment": alignment,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTransitionAlignment", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTransitionAlignment: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTransitionProperties(ctx context.Context, trackType string, trackIndex, transitionIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"transitionIndex": transitionIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTransitionProperties", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTransitionProperties: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Effect Comparison Operations
// ---------------------------------------------------------------------------

func (e *Engine) ToggleEffectsPreview(ctx context.Context, trackType string, trackIndex, clipIndex int, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "toggleEffectsPreview", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ToggleEffectsPreview: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetBeforeAfterSnapshot(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getBeforeAfterSnapshot", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetBeforeAfterSnapshot: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CompareEffectSettings(ctx context.Context, clip1RefJSON, clip2RefJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"clip1RefJSON": clip1RefJSON,
		"clip2RefJSON": clip2RefJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareEffectSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareEffectSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Effect Templates Operations
// ---------------------------------------------------------------------------

func (e *Engine) SaveEffectChainAsTemplate(ctx context.Context, trackType string, trackIndex, clipIndex int, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveEffectChainAsTemplate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveEffectChainAsTemplate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) LoadEffectChainTemplate(ctx context.Context, trackType string, trackIndex, clipIndex int, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "loadEffectChainTemplate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LoadEffectChainTemplate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ListEffectChainTemplates(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listEffectChainTemplates", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListEffectChainTemplates: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
