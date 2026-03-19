package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Multicam Operations
// ---------------------------------------------------------------------------

func (e *Engine) CreateMulticamSequence(ctx context.Context, name string, clipIndices []int, syncPoint string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"clipIndices": clipIndices,
		"syncPoint": syncPoint,
	})
	result, err := e.premiere.EvalCommand(ctx, "createMulticamSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateMulticamSequence: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SwitchMulticamAngle(ctx context.Context, trackIndex int, time float64, angleIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"time": time,
		"angleIndex": angleIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "switchMulticamAngle", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SwitchMulticamAngle: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) FlattenMulticam(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "flattenMulticam", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FlattenMulticam: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetMulticamAngles(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getMulticamAngles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMulticamAngles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Proxy Workflow Operations
// ---------------------------------------------------------------------------

func (e *Engine) CreateProxy(ctx context.Context, projectItemIndex int, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createProxy", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateProxy: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AttachProxy(ctx context.Context, projectItemIndex int, proxyPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"proxyPath": proxyPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "attachProxy", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AttachProxy: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) HasProxy(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "hasProxy", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("HasProxy: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetProxyPath(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getProxyPath", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProxyPath: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ToggleProxies(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "toggleProxies", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ToggleProxies: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DetachProxy(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "detachProxy", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DetachProxy: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Workspace Operations
// ---------------------------------------------------------------------------

func (e *Engine) GetWorkspaces(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getWorkspaces", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetWorkspaces: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetWorkspace(ctx context.Context, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "setWorkspace", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetWorkspace: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SaveWorkspace(ctx context.Context, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveWorkspace", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveWorkspace: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Undo / Redo Operations
// ---------------------------------------------------------------------------

func (e *Engine) Undo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "undo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("Undo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) Redo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "redo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("Redo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Panel Operations
// ---------------------------------------------------------------------------

func (e *Engine) SortProjectPanel(ctx context.Context, field string, ascending bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"field": field,
		"ascending": ascending,
	})
	result, err := e.premiere.EvalCommand(ctx, "sortProjectPanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SortProjectPanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SearchProjectPanel(ctx context.Context, query string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"query": query,
	})
	result, err := e.premiere.EvalCommand(ctx, "searchProjectPanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SearchProjectPanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Source Monitor Operations
// ---------------------------------------------------------------------------

func (e *Engine) OpenInSourceMonitor(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "openInSourceMonitor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OpenInSourceMonitor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetSourceMonitorPosition(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSourceMonitorPosition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSourceMonitorPosition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetSourceMonitorPosition(ctx context.Context, seconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "setSourceMonitorPosition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetSourceMonitorPosition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Preferences Operations
// ---------------------------------------------------------------------------

func (e *Engine) GetAutoSaveSettings(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAutoSaveSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAutoSaveSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetAutoSaveInterval(ctx context.Context, minutes int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"minutes": minutes,
	})
	result, err := e.premiere.EvalCommand(ctx, "setAutoSaveInterval", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAutoSaveInterval: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetMemorySettings(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMemorySettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMemorySettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Media Cache Operations
// ---------------------------------------------------------------------------

func (e *Engine) ClearMediaCache(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearMediaCache", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearMediaCache: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetMediaCachePath(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMediaCachePath", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaCachePath: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
