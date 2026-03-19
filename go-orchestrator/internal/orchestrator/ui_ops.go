package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// UI Panel Control
// ---------------------------------------------------------------------------

func (e *Engine) OpenPanel(ctx context.Context, panelName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"panelName": panelName,
	})
	result, err := e.premiere.EvalCommand(ctx, "openPanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OpenPanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ClosePanel(ctx context.Context, panelName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"panelName": panelName,
	})
	result, err := e.premiere.EvalCommand(ctx, "closePanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClosePanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetOpenPanels(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getOpenPanels", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetOpenPanels: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ResetPanelLayout(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "resetPanelLayout", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetPanelLayout: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) MaximizePanel(ctx context.Context, panelName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"panelName": panelName,
	})
	result, err := e.premiere.EvalCommand(ctx, "maximizePanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MaximizePanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Window Management
// ---------------------------------------------------------------------------

func (e *Engine) GetWindowInfo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getWindowInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetWindowInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetWindowSize(ctx context.Context, width, height int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"width": width,
		"height": height,
	})
	result, err := e.premiere.EvalCommand(ctx, "setWindowSize", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetWindowSize: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) MinimizeWindow(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "minimizeWindow", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MinimizeWindow: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) BringToFront(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "bringToFront", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BringToFront: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) EnterFullscreen(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "enterFullscreen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EnterFullscreen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline UI
// ---------------------------------------------------------------------------

func (e *Engine) SetTrackHeight(ctx context.Context, trackType string, trackIndex int, height int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"height": height,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTrackHeight", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTrackHeight: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CollapseTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "collapseTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CollapseTrack: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExpandTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "expandTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExpandTrack: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CollapseAllTracks(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "collapseAllTracks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CollapseAllTracks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExpandAllTracks(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "expandAllTracks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExpandAllTracks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Label Management
// ---------------------------------------------------------------------------

func (e *Engine) SetLabelPreferences(ctx context.Context, labelsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"labelsJSON": labelsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLabelPreferences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLabelPreferences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetActiveLabelFilter(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getActiveLabelFilter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetActiveLabelFilter: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetLabelFilter(ctx context.Context, colorIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"colorIndex": colorIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "setLabelFilter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetLabelFilter: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ClearLabelFilter(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearLabelFilter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearLabelFilter: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Display
// ---------------------------------------------------------------------------

func (e *Engine) SetAudioWaveformDisplay(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setAudioWaveformDisplay", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAudioWaveformDisplay: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetVideoThumbnailDisplay(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setVideoThumbnailDisplay", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetVideoThumbnailDisplay: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTrackNameDisplay(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTrackNameDisplay", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTrackNameDisplay: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// User Feedback
// ---------------------------------------------------------------------------

func (e *Engine) ShowAlert(ctx context.Context, title, message string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"message": message,
	})
	result, err := e.premiere.EvalCommand(ctx, "showAlert", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowAlert: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShowConfirmDialog(ctx context.Context, title, message string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"message": message,
	})
	result, err := e.premiere.EvalCommand(ctx, "showConfirmDialog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowConfirmDialog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShowInputDialog(ctx context.Context, title, prompt, defaultValue string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"prompt": prompt,
		"defaultValue": defaultValue,
	})
	result, err := e.premiere.EvalCommand(ctx, "showInputDialog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowInputDialog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShowProgressDialog(ctx context.Context, title, message string, progress float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"message": message,
		"progress": progress,
	})
	result, err := e.premiere.EvalCommand(ctx, "showProgressDialog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowProgressDialog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) WriteToConsole(ctx context.Context, message string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"message": message,
	})
	result, err := e.premiere.EvalCommand(ctx, "writeToConsole", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WriteToConsole: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Accessibility
// ---------------------------------------------------------------------------

func (e *Engine) GetUIScaling(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getUIScaling", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUIScaling: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetHighContrastMode(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "setHighContrastMode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetHighContrastMode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
