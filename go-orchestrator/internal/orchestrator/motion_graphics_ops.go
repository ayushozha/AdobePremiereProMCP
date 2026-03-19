package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Essential Graphics Panel
// ---------------------------------------------------------------------------

// GetEssentialGraphicsComponents returns all EGP component properties for a clip.
func (e *Engine) GetEssentialGraphicsComponents(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEssentialGraphicsComponents", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEssentialGraphicsComponents: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetEssentialGraphicsProperty sets a property value on an EGP clip.
func (e *Engine) SetEssentialGraphicsProperty(ctx context.Context, trackIndex, clipIndex int, propName, value string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"propName": propName,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setEssentialGraphicsProperty", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetEssentialGraphicsProperty: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetEssentialGraphicsText returns all text content from a graphics clip.
func (e *Engine) GetEssentialGraphicsText(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEssentialGraphicsText", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEssentialGraphicsText: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ReplaceAllText finds and replaces text in a graphics clip.
func (e *Engine) ReplaceAllText(ctx context.Context, trackIndex, clipIndex int, searchText, replaceText string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"searchText": searchText,
		"replaceText": replaceText,
	})
	result, err := e.premiere.EvalCommand(ctx, "replaceAllText", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReplaceAllText: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// MOGRT Management (extended)
// ---------------------------------------------------------------------------

// ListInstalledMOGRTs lists all installed Motion Graphics Templates.
func (e *Engine) ListInstalledMOGRTs(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listInstalledMOGRTs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListInstalledMOGRTs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetMOGRTInfo returns info about a MOGRT file.
func (e *Engine) GetMOGRTInfo(ctx context.Context, mogrtPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"mogrtPath": mogrtPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getMOGRTInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMOGRTInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchUpdateMOGRTs updates a property on all MOGRTs on a track.
func (e *Engine) BatchUpdateMOGRTs(ctx context.Context, trackIndex int, propertyName, value string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"propertyName": propertyName,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchUpdateMOGRTs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchUpdateMOGRTs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateMOGRTFromClip exports a clip as a MOGRT.
func (e *Engine) CreateMOGRTFromClip(ctx context.Context, trackIndex, clipIndex int, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createMOGRTFromClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateMOGRTFromClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Text Operations
// ---------------------------------------------------------------------------

// AddScrollingTitle adds a scrolling/crawling title overlay.
func (e *Engine) AddScrollingTitle(ctx context.Context, text string, trackIndex int, startTime, duration, speed float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"text": text,
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"speed": speed,
	})
	result, err := e.premiere.EvalCommand(ctx, "addScrollingTitle", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddScrollingTitle: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddTypewriterText adds text with a typewriter animation effect.
func (e *Engine) AddTypewriterText(ctx context.Context, text string, trackIndex int, startTime, duration, typeSpeed float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"text": text,
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"typeSpeed": typeSpeed,
	})
	result, err := e.premiere.EvalCommand(ctx, "addTypewriterText", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddTypewriterText: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddTextWithBackground adds text with a background box.
func (e *Engine) AddTextWithBackground(ctx context.Context, text string, trackIndex int, startTime, duration float64, bgColor string, padding int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"text": text,
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"bgColor": bgColor,
		"padding": padding,
	})
	result, err := e.premiere.EvalCommand(ctx, "addTextWithBackground", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddTextWithBackground: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetTextAnimation sets a text animation type on a clip.
func (e *Engine) SetTextAnimation(ctx context.Context, trackIndex, clipIndex int, animationType string, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"animationType": animationType,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTextAnimation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTextAnimation: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Shape Layers
// ---------------------------------------------------------------------------

// AddRectangle adds a colored rectangle shape to the timeline.
func (e *Engine) AddRectangle(ctx context.Context, trackIndex int, startTime, duration, x, y float64, width, height int, color string, borderWidth int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"x": x,
		"y": y,
		"width": width,
		"height": height,
		"color": color,
		"borderWidth": borderWidth,
	})
	result, err := e.premiere.EvalCommand(ctx, "addRectangle", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddRectangle: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddCircle adds a circle shape to the timeline.
func (e *Engine) AddCircle(ctx context.Context, trackIndex int, startTime, duration, x, y float64, radius int, color string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"x": x,
		"y": y,
		"radius": radius,
		"color": color,
	})
	result, err := e.premiere.EvalCommand(ctx, "addCircle", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddCircle: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddLine adds a line shape to the timeline.
func (e *Engine) AddLine(ctx context.Context, trackIndex int, startTime, duration, x1, y1, x2, y2 float64, color string, thickness int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"x1": x1,
		"y1": y1,
		"x2": x2,
		"y2": y2,
		"color": color,
		"thickness": thickness,
	})
	result, err := e.premiere.EvalCommand(ctx, "addLine", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddLine: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Countdown / Timers
// ---------------------------------------------------------------------------

// AddCountdown adds a countdown timer overlay.
func (e *Engine) AddCountdown(ctx context.Context, trackIndex int, startTime float64, fromSeconds int, style string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"fromSeconds": fromSeconds,
		"style": style,
	})
	result, err := e.premiere.EvalCommand(ctx, "addCountdown", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddCountdown: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddTimecode adds a timecode burn-in overlay.
func (e *Engine) AddTimecode(ctx context.Context, trackIndex int, startTime, duration float64, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startTime": startTime,
		"duration": duration,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "addTimecode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddTimecode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Watermark
// ---------------------------------------------------------------------------

// AddWatermark adds an image watermark overlay.
func (e *Engine) AddWatermark(ctx context.Context, imagePath, position string, opacity, scale float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"imagePath": imagePath,
		"position": position,
		"opacity": opacity,
		"scale": scale,
	})
	result, err := e.premiere.EvalCommand(ctx, "addWatermark", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddWatermark: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddTextWatermark adds a text watermark overlay.
func (e *Engine) AddTextWatermark(ctx context.Context, text, position string, opacity float64, fontSize int, color string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"text": text,
		"position": position,
		"opacity": opacity,
		"fontSize": fontSize,
		"color": color,
	})
	result, err := e.premiere.EvalCommand(ctx, "addTextWatermark", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddTextWatermark: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RemoveWatermark removes a watermark clip from the timeline.
func (e *Engine) RemoveWatermark(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "removeWatermark", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RemoveWatermark: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Picture Layouts
// ---------------------------------------------------------------------------

// CreateSplitScreen creates a split screen layout.
func (e *Engine) CreateSplitScreen(ctx context.Context, layout, clipRefsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"layout": layout,
		"clipRefsJSON": clipRefsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSplitScreen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSplitScreen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateCollage creates a photo/video collage grid.
func (e *Engine) CreateCollage(ctx context.Context, clipRefsJSON string, rows, cols, gap int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"clipRefsJSON": clipRefsJSON,
		"rows": rows,
		"cols": cols,
		"gap": gap,
	})
	result, err := e.premiere.EvalCommand(ctx, "createCollage", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateCollage: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Animated Transitions (custom)
// ---------------------------------------------------------------------------

// AddWipeTransition adds a custom wipe transition with color.
func (e *Engine) AddWipeTransition(ctx context.Context, trackIndex, clipIndex int, direction, color string, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"direction": direction,
		"color": color,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "addWipeTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddWipeTransition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddZoomTransition adds a zoom in/out transition.
func (e *Engine) AddZoomTransition(ctx context.Context, trackIndex, clipIndex int, zoomIn bool, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"zoomIn": zoomIn,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "addZoomTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddZoomTransition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddGlitchTransition adds a glitch effect transition.
func (e *Engine) AddGlitchTransition(ctx context.Context, trackIndex, clipIndex int, intensity, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"intensity": intensity,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "addGlitchTransition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddGlitchTransition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Subtitling (extended)
// ---------------------------------------------------------------------------

// AutoGenerateSubtitles auto-generates subtitles from audio.
func (e *Engine) AutoGenerateSubtitles(ctx context.Context, language, style string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"language": language,
		"style": style,
	})
	result, err := e.premiere.EvalCommand(ctx, "autoGenerateSubtitles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AutoGenerateSubtitles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// TranslateSubtitles translates existing subtitles to a target language.
func (e *Engine) TranslateSubtitles(ctx context.Context, trackIndex int, targetLanguage string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"targetLanguage": targetLanguage,
	})
	result, err := e.premiere.EvalCommand(ctx, "translateSubtitles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TranslateSubtitles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// FormatSubtitles reformats subtitle line breaks.
func (e *Engine) FormatSubtitles(ctx context.Context, trackIndex, maxCharsPerLine, maxLines int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"maxCharsPerLine": maxCharsPerLine,
		"maxLines": maxLines,
	})
	result, err := e.premiere.EvalCommand(ctx, "formatSubtitles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FormatSubtitles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BurnInSubtitles burns subtitles into the video.
func (e *Engine) BurnInSubtitles(ctx context.Context, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "burnInSubtitles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BurnInSubtitles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AdjustSubtitleTiming shifts all subtitle timing by an offset.
func (e *Engine) AdjustSubtitleTiming(ctx context.Context, trackIndex int, offsetSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"offsetSeconds": offsetSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "adjustSubtitleTiming", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AdjustSubtitleTiming: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
