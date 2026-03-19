package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Color Correction & Lumetri Color Operations
// ---------------------------------------------------------------------------

// LumetriGetAll returns all Lumetri Color parameter values for a clip.
func (e *Engine) LumetriGetAll(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriGetAll", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriGetAll: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetExposure sets the Lumetri Color exposure (-4.0 to 4.0).
func (e *Engine) LumetriSetExposure2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetExposure2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetExposure2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetContrast2 sets the Lumetri Color contrast (-100 to 100).
func (e *Engine) LumetriSetContrast2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetContrast2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetContrast2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetHighlights sets the Lumetri Color highlights (-100 to 100).
func (e *Engine) LumetriSetHighlights(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetHighlights", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetHighlights: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetShadows sets the Lumetri Color shadows (-100 to 100).
func (e *Engine) LumetriSetShadows(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetShadows", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetShadows: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetWhites sets the Lumetri Color whites (-100 to 100).
func (e *Engine) LumetriSetWhites(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetWhites", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetWhites: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetBlacks sets the Lumetri Color blacks (-100 to 100).
func (e *Engine) LumetriSetBlacks(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetBlacks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetBlacks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetTemperature2 sets the Lumetri Color white balance temperature.
func (e *Engine) LumetriSetTemperature2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetTemperature2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetTemperature2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetTint2 sets the Lumetri Color white balance tint.
func (e *Engine) LumetriSetTint2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetTint2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetTint2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetSaturation2 sets the Lumetri Color saturation (0 to 200).
func (e *Engine) LumetriSetSaturation2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetSaturation2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetSaturation2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetVibrance sets the Lumetri Color vibrance (-100 to 100).
func (e *Engine) LumetriSetVibrance(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetVibrance", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetVibrance: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetFadedFilm sets the Lumetri Color faded film amount.
func (e *Engine) LumetriSetFadedFilm(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetFadedFilm", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetFadedFilm: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetSharpen sets the Lumetri Color sharpening amount.
func (e *Engine) LumetriSetSharpen(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetSharpen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetSharpen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetCurvePoint sets a control point on a Lumetri Color curve.
func (e *Engine) LumetriSetCurvePoint(ctx context.Context, trackIndex, clipIndex int, channel string, inputValue, outputValue float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"channel": channel,
		"inputValue": inputValue,
		"outputValue": outputValue,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetCurvePoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetCurvePoint: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetShadowColor sets the shadow color wheel values.
func (e *Engine) LumetriSetShadowColor(ctx context.Context, trackIndex, clipIndex int, hue, saturation, brightness float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"hue": hue,
		"saturation": saturation,
		"brightness": brightness,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetShadowColor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetShadowColor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetMidtoneColor sets the midtone color wheel values.
func (e *Engine) LumetriSetMidtoneColor(ctx context.Context, trackIndex, clipIndex int, hue, saturation, brightness float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"hue": hue,
		"saturation": saturation,
		"brightness": brightness,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetMidtoneColor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetMidtoneColor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetHighlightColor sets the highlight color wheel values.
func (e *Engine) LumetriSetHighlightColor(ctx context.Context, trackIndex, clipIndex int, hue, saturation, brightness float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"hue": hue,
		"saturation": saturation,
		"brightness": brightness,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetHighlightColor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetHighlightColor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetVignetteAmount sets the Lumetri Color vignette amount.
func (e *Engine) LumetriSetVignetteAmount(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetVignetteAmount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetVignetteAmount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetVignetteMidpoint sets the Lumetri Color vignette midpoint.
func (e *Engine) LumetriSetVignetteMidpoint(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetVignetteMidpoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetVignetteMidpoint: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetVignetteRoundness sets the Lumetri Color vignette roundness.
func (e *Engine) LumetriSetVignetteRoundness(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetVignetteRoundness", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetVignetteRoundness: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriSetVignetteFeather sets the Lumetri Color vignette feather.
func (e *Engine) LumetriSetVignetteFeather(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriSetVignetteFeather", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriSetVignetteFeather: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriApplyLUT applies a LUT file to the Lumetri Color effect.
func (e *Engine) LumetriApplyLUT(ctx context.Context, trackIndex, clipIndex int, lutPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"lutPath": lutPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriApplyLUT", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriApplyLUT: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriRemoveLUT removes an applied LUT from the Lumetri Color effect.
func (e *Engine) LumetriRemoveLUT(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriRemoveLUT", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriRemoveLUT: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriAutoColor applies automatic color correction with reasonable defaults.
func (e *Engine) LumetriAutoColor(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriAutoColor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriAutoColor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriReset resets all Lumetri Color settings to their defaults.
func (e *Engine) LumetriReset(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriReset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriReset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetColorInfo retrieves basic color statistics and current Lumetri settings.
func (e *Engine) GetColorInfo(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getColorInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetColorInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CopyColorGrade copies the Lumetri Color settings from a source clip.
func (e *Engine) CopyColorGrade(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "copyColorGrade", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CopyColorGrade: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// PasteColorGrade pastes previously copied Lumetri Color settings to a clip.
func (e *Engine) PasteColorGrade(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "pasteColorGrade", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PasteColorGrade: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ApplyColorGradeToAll applies a color grade from a source clip to all clips on a track.
func (e *Engine) ApplyColorGradeToAll(ctx context.Context, srcTrackIndex, srcClipIndex, destTrackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"srcTrackIndex": srcTrackIndex,
		"srcClipIndex": srcClipIndex,
		"destTrackIndex": destTrackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyColorGradeToAll", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyColorGradeToAll: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LumetriAutoWhiteBalance applies auto white balance by resetting temperature and tint.
func (e *Engine) LumetriAutoWhiteBalance(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "lumetriAutoWhiteBalance", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LumetriAutoWhiteBalance: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
