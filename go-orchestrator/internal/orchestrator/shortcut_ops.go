package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Menu Commands
// ---------------------------------------------------------------------------

func (e *Engine) GetMenuItems(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMenuItems", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMenuItems: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSubmenuItems(ctx context.Context, menuPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"menuPath": menuPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSubmenuItems", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSubmenuItems: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteMenuItemByID(ctx context.Context, menuItemID string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"menuItemID": menuItemID,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeMenuItemByID", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteMenuItemByID: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FindMenuItem(ctx context.Context, searchText string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"searchText": searchText,
	})
	result, err := e.premiere.EvalCommand(ctx, "findMenuItem", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FindMenuItem: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) IsMenuItemEnabled(ctx context.Context, menuItemID string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"menuItemID": menuItemID,
	})
	result, err := e.premiere.EvalCommand(ctx, "isMenuItemEnabled", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsMenuItemEnabled: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) IsMenuItemChecked(ctx context.Context, menuItemID string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"menuItemID": menuItemID,
	})
	result, err := e.premiere.EvalCommand(ctx, "isMenuItemChecked", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsMenuItemChecked: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Keyboard Shortcuts
// ---------------------------------------------------------------------------

func (e *Engine) GetShortcutForCommand(ctx context.Context, commandID string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"commandID": commandID,
	})
	result, err := e.premiere.EvalCommand(ctx, "getShortcutForCommand", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetShortcutForCommand: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetAllShortcuts(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAllShortcuts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAllShortcuts: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SimulateKeyPress(ctx context.Context, key string, modifiers string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"key": key,
		"modifiers": modifiers,
	})
	result, err := e.premiere.EvalCommand(ctx, "simulateKeyPress", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SimulateKeyPress: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetShortcutConflicts(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getShortcutConflicts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetShortcutConflicts: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Quick Actions
// ---------------------------------------------------------------------------

func (e *Engine) ToggleFullScreen(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "toggleFullScreen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ToggleFullScreen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ToggleMaximizeFrame(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "toggleMaximizeFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ToggleMaximizeFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ClearSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SelectAll(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "selectAll", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SelectAll: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CutSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "cutSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CutSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CopySelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "copySelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CopySelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) PasteSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "pasteSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PasteSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DuplicateSelection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "duplicateSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DuplicateSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// View Controls
// ---------------------------------------------------------------------------

func (e *Engine) SetZoomLevel(ctx context.Context, level float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"level": level,
	})
	result, err := e.premiere.EvalCommand(ctx, "setZoomLevel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetZoomLevel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetZoomLevel(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getZoomLevel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetZoomLevel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ScrollTimelineTo(ctx context.Context, seconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "scrollTimelineTo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ScrollTimelineTo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) EnableLinkedSelection(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "enableLinkedSelection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EnableLinkedSelection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetLinkedSelectionState(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getLinkedSelectionState", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLinkedSelectionState: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) EnableInsertAndOverwrite(ctx context.Context, trackType string, trackIndex int, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "enableInsertAndOverwrite", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EnableInsertAndOverwrite: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Sequence Display
// ---------------------------------------------------------------------------

func (e *Engine) ShowAudioTimeUnits(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "showAudioTimeUnits", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowAudioTimeUnits: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShowDuplicateFrameMarkers(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "showDuplicateFrameMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowDuplicateFrameMarkers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ShowClipMismatchWarning(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "showClipMismatchWarning", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowClipMismatchWarning: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTimelineSnap(ctx context.Context, snapType string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"snapType": snapType,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTimelineSnap", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTimelineSnap: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetTimelineViewExtents(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineViewExtents", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineViewExtents: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetTimelineViewExtents(ctx context.Context, startSeconds, endSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"startSeconds": startSeconds,
		"endSeconds": endSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTimelineViewExtents", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTimelineViewExtents: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
