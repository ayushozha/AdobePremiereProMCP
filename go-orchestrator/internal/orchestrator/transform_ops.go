package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Transform, Crop, Masking, Stabilization, Blur & Distortion Operations
// ---------------------------------------------------------------------------

// --- Crop Effect ---

func (e *Engine) SetCrop(ctx context.Context, trackIndex, clipIndex int, left, right, top, bottom float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"left": left,
		"right": right,
		"top": top,
		"bottom": bottom,
	})
	result, err := e.premiere.EvalCommand(ctx, "setCrop", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetCrop: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetCrop(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getCrop", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetCrop: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ResetCrop(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "resetCrop", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetCrop: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Transform (extended) ---

func (e *Engine) SetUniformScale(ctx context.Context, trackIndex, clipIndex int, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setUniformScale", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetUniformScale: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTransformProperties(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTransformProperties", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTransformProperties: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetAntiFlicker(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setAntiFlicker", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAntiFlicker: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ResetTransform(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "resetTransform", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetTransform: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CenterClip(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "centerClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CenterClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FitClipToFrame(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "fitClipToFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FitClipToFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FillFrame(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "fillFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FillFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Picture-in-Picture ---

func (e *Engine) CreatePIP(ctx context.Context, mainTrackIndex, mainClipIndex, pipTrackIndex, pipClipIndex int, position string, scale float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"mainTrackIndex": mainTrackIndex,
		"mainClipIndex": mainClipIndex,
		"pipTrackIndex": pipTrackIndex,
		"pipClipIndex": pipClipIndex,
		"position": position,
		"scale": scale,
	})
	result, err := e.premiere.EvalCommand(ctx, "createPIP", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreatePIP: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RemovePIP(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "removePIP", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemovePIP: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Opacity & Masking ---

func (e *Engine) SetOpacityKeyframes(ctx context.Context, trackIndex, clipIndex int, keyframesJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"keyframesJSON": keyframesJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "setOpacityKeyframes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetOpacityKeyframes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FadeIn(ctx context.Context, trackIndex, clipIndex int, durationSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"durationSeconds": durationSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "fadeIn", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FadeIn: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FadeOut(ctx context.Context, trackIndex, clipIndex int, durationSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"durationSeconds": durationSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "fadeOut", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FadeOut: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CrossFadeClips(ctx context.Context, trackIndex, clipIndexA, clipIndexB int, durationSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndexA": clipIndexA,
		"clipIndexB": clipIndexB,
		"durationSeconds": durationSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "crossFadeClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CrossFadeClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Stabilization ---

func (e *Engine) ApplyWarpStabilizer(ctx context.Context, trackIndex, clipIndex int, smoothness float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"smoothness": smoothness,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyWarpStabilizer", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyWarpStabilizer: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetStabilizationStatus(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getStabilizationStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetStabilizationStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Lens Distortion ---

func (e *Engine) ApplyLensDistortionRemoval(ctx context.Context, trackIndex, clipIndex int, curvature float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"curvature": curvature,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyLensDistortionRemoval", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyLensDistortionRemoval: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Noise Reduction ---

func (e *Engine) ApplyVideoNoiseReduction(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyVideoNoiseReduction", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyVideoNoiseReduction: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyAudioNoiseReduction(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyAudioNoiseReduction", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyAudioNoiseReduction: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyDeReverb(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyDeReverb", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyDeReverb: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyDeHum(ctx context.Context, trackIndex, clipIndex int, frequency float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"frequency": frequency,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyDeHum", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyDeHum: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Blur & Sharpen ---

func (e *Engine) ApplyGaussianBlur(ctx context.Context, trackIndex, clipIndex int, blurriness float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"blurriness": blurriness,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyGaussianBlur", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyGaussianBlur: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyDirectionalBlur(ctx context.Context, trackIndex, clipIndex int, direction, length float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"direction": direction,
		"length": length,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyDirectionalBlur", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyDirectionalBlur: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplySharpen(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
	})
	result, err := e.premiere.EvalCommand(ctx, "applySharpen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplySharpen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyUnsharpMask(ctx context.Context, trackIndex, clipIndex int, amount, radius, threshold float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"amount": amount,
		"radius": radius,
		"threshold": threshold,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyUnsharpMask", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyUnsharpMask: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Distortion ---

func (e *Engine) ApplyMirror(ctx context.Context, trackIndex, clipIndex int, angle float64, centerX, centerY float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"angle": angle,
		"centerX": centerX,
		"centerY": centerY,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyMirror", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyMirror: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyCornerPin(ctx context.Context, trackIndex, clipIndex int, cornersJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"cornersJSON": cornersJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyCornerPin", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyCornerPin: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplySpherize(ctx context.Context, trackIndex, clipIndex int, radius, centerX, centerY float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"radius": radius,
		"centerX": centerX,
		"centerY": centerY,
	})
	result, err := e.premiere.EvalCommand(ctx, "applySpherize", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplySpherize: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
