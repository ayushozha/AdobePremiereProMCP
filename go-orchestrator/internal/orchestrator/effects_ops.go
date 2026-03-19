package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Effects & Transitions Operations (Extended)
// ---------------------------------------------------------------------------

func (e *Engine) AddVideoTransition(ctx context.Context, trackIndex, clipIndex int, transitionName string, duration float64, applyToEnd bool) (*GenericResult, error) {
	e.logger.Debug("add_video_transition", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.String("transition", transitionName))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"transitionName": transitionName,
		"duration":       duration,
		"applyToEnd":     applyToEnd,
	})
	result, err := e.premiere.EvalCommand(ctx, "addVideoTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddVideoTransition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) AddAudioTransition(ctx context.Context, trackIndex, clipIndex int, transitionName string, duration float64) (*GenericResult, error) {
	e.logger.Debug("add_audio_transition", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.String("transition", transitionName))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"transitionName": transitionName,
		"duration":       duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "addAudioTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddAudioTransition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) RemoveTransition(ctx context.Context, trackType string, trackIndex, transitionIndex int) (*GenericResult, error) {
	e.logger.Debug("remove_transition", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("transition", transitionIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":       trackType,
		"trackIndex":      trackIndex,
		"transitionIndex": transitionIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "removeTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemoveTransition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) GetTransitions(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	e.logger.Debug("get_transitions", zap.String("track_type", trackType), zap.Int("track", trackIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTransitions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTransitions: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetDefaultVideoTransition(ctx context.Context, transitionName string) (*GenericResult, error) {
	e.logger.Debug("set_default_video_transition", zap.String("transition", transitionName))
	argsJSON, _ := json.Marshal(map[string]any{
		"transitionName": transitionName,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDefaultVideoTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDefaultVideoTransition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetDefaultAudioTransition(ctx context.Context, transitionName string) (*GenericResult, error) {
	e.logger.Debug("set_default_audio_transition", zap.String("transition", transitionName))
	argsJSON, _ := json.Marshal(map[string]any{
		"transitionName": transitionName,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDefaultAudioTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDefaultAudioTransition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) ApplyDefaultTransition(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("apply_default_transition", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyDefaultTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyDefaultTransition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) GetAvailableTransitions(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_available_transitions")
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAvailableTransitions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAvailableTransitions: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) ApplyVideoEffect(ctx context.Context, trackIndex, clipIndex int, effectName string) (*GenericResult, error) {
	e.logger.Debug("apply_video_effect", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.String("effect", effectName))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"effectName": effectName,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyVideoEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyVideoEffect: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) RemoveVideoEffect(ctx context.Context, trackIndex, clipIndex, effectIndex int) (*GenericResult, error) {
	e.logger.Debug("remove_video_effect", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Int("effect", effectIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex":  trackIndex,
		"clipIndex":   clipIndex,
		"effectIndex": effectIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "removeVideoEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemoveVideoEffect: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) GetClipEffects(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("get_clip_effects", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  trackType,
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipEffects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipEffects: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetEffectParameter(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_effect_parameter", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Int("component", componentIndex), zap.Int("param", paramIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
		"value":          value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setEffectParameter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetEffectParameter: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) GetEffectParameter(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int) (*GenericResult, error) {
	e.logger.Debug("get_effect_parameter", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Int("component", componentIndex), zap.Int("param", paramIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEffectParameter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEffectParameter: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) EnableEffect(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex int, enabled bool) (*GenericResult, error) {
	e.logger.Debug("enable_effect", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Int("component", componentIndex), zap.Bool("enabled", enabled))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"enabled":        enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "enableEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EnableEffect: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) CopyEffects(ctx context.Context, srcTrackType string, srcTrackIndex, srcClipIndex int) (*GenericResult, error) {
	e.logger.Debug("copy_effects", zap.String("track_type", srcTrackType), zap.Int("track", srcTrackIndex), zap.Int("clip", srcClipIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  srcTrackType,
		"trackIndex": srcTrackIndex,
		"clipIndex":  srcClipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "copyEffects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CopyEffects: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) PasteEffects(ctx context.Context, destTrackType string, destTrackIndex, destClipIndex int) (*GenericResult, error) {
	e.logger.Debug("paste_effects", zap.String("track_type", destTrackType), zap.Int("track", destTrackIndex), zap.Int("clip", destClipIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":  destTrackType,
		"trackIndex": destTrackIndex,
		"clipIndex":  destClipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "pasteEffects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PasteEffects: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetPosition(ctx context.Context, trackIndex, clipIndex int, x, y float64) (*GenericResult, error) {
	e.logger.Debug("set_position", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("x", x), zap.Float64("y", y))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"x":          x,
		"y":          y,
	})
	result, err := e.premiere.EvalCommand(ctx, "setPosition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetPosition: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetScale(ctx context.Context, trackIndex, clipIndex int, scale float64) (*GenericResult, error) {
	e.logger.Debug("set_scale", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("scale", scale))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"scale":      scale,
	})
	result, err := e.premiere.EvalCommand(ctx, "setScale", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetScale: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetRotation(ctx context.Context, trackIndex, clipIndex int, degrees float64) (*GenericResult, error) {
	e.logger.Debug("set_rotation", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("degrees", degrees))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"degrees":    degrees,
	})
	result, err := e.premiere.EvalCommand(ctx, "setRotation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetRotation: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetAnchorPoint(ctx context.Context, trackIndex, clipIndex int, x, y float64) (*GenericResult, error) {
	e.logger.Debug("set_anchor_point", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("x", x), zap.Float64("y", y))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"x":          x,
		"y":          y,
	})
	result, err := e.premiere.EvalCommand(ctx, "setAnchorPoint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAnchorPoint: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetOpacity(ctx context.Context, trackIndex, clipIndex int, opacity float64) (*GenericResult, error) {
	e.logger.Debug("set_opacity", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("opacity", opacity))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"opacity":    opacity,
	})
	result, err := e.premiere.EvalCommand(ctx, "setOpacity", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetOpacity: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) GetMotionProperties(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("get_motion_properties", zap.Int("track", trackIndex), zap.Int("clip", clipIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getMotionProperties", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMotionProperties: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetBlendMode(ctx context.Context, trackIndex, clipIndex int, mode string) (*GenericResult, error) {
	e.logger.Debug("set_blend_mode", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.String("mode", mode))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"mode":       mode,
	})
	result, err := e.premiere.EvalCommand(ctx, "setBlendMode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetBlendMode: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) CreateAdjustmentLayer(ctx context.Context, name string, width, height int, duration float64) (*GenericResult, error) {
	e.logger.Debug("create_adjustment_layer", zap.String("name", name), zap.Int("width", width), zap.Int("height", height), zap.Float64("duration", duration))
	argsJSON, _ := json.Marshal(map[string]any{
		"name":     name,
		"width":    width,
		"height":   height,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "createAdjustmentLayer", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateAdjustmentLayer: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) PlaceAdjustmentLayer(ctx context.Context, projectItemIndex, trackIndex int, startTime, duration float64) (*GenericResult, error) {
	e.logger.Debug("place_adjustment_layer", zap.Int("project_item", projectItemIndex), zap.Int("track", trackIndex), zap.Float64("start", startTime))
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"trackIndex":       trackIndex,
		"startTime":        startTime,
		"duration":         duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "placeAdjustmentLayer", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PlaceAdjustmentLayer: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) AddKeyframe(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, time, value float64) (*GenericResult, error) {
	e.logger.Debug("add_keyframe", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("time", time), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
		"time":           time,
		"value":          value,
	})
	result, err := e.premiere.EvalCommand(ctx, "addKeyframe", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddKeyframe: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) DeleteKeyframe(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, time float64) (*GenericResult, error) {
	e.logger.Debug("delete_keyframe", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("time", time))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
		"time":           time,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteKeyframe", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DeleteKeyframe: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetKeyframeInterpolation(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, time float64, interpType string) (*GenericResult, error) {
	e.logger.Debug("set_keyframe_interpolation", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("time", time), zap.String("interp", interpType))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
		"time":           time,
		"interpType":     interpType,
	})
	result, err := e.premiere.EvalCommand(ctx, "setKeyframeInterpolation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetKeyframeInterpolation: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) GetKeyframes(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int) (*GenericResult, error) {
	e.logger.Debug("get_keyframes", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Int("component", componentIndex), zap.Int("param", paramIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getKeyframes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetKeyframes: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetTimeVarying(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, enabled bool) (*GenericResult, error) {
	e.logger.Debug("set_time_varying", zap.String("track_type", trackType), zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Bool("enabled", enabled))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType":      trackType,
		"trackIndex":     trackIndex,
		"clipIndex":      clipIndex,
		"componentIndex": componentIndex,
		"paramIndex":     paramIndex,
		"enabled":        enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTimeVarying", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTimeVarying: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetLumetriBrightness(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_lumetri_brightness", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"value":      value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLumetriBrightness", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLumetriBrightness: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetLumetriContrast(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_lumetri_contrast", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"value":      value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLumetriContrast", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLumetriContrast: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetLumetriSaturation(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_lumetri_saturation", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"value":      value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLumetriSaturation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLumetriSaturation: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetLumetriTemperature(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_lumetri_temperature", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"value":      value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLumetriTemperature", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLumetriTemperature: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetLumetriTint(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_lumetri_tint", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"value":      value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLumetriTint", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLumetriTint: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}

func (e *Engine) SetLumetriExposure(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error) {
	e.logger.Debug("set_lumetri_exposure", zap.Int("track", trackIndex), zap.Int("clip", clipIndex), zap.Float64("value", value))
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex":  clipIndex,
		"value":      value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLumetriExposure", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLumetriExposure: %w", err)
	}
	return &GenericResult{Status: "ok", Message: result}, nil
}
