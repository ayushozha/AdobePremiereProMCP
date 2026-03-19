package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Performance Monitoring
// ---------------------------------------------------------------------------

// GetPerformanceMetrics returns CPU, memory, and GPU usage information.
func (e *Engine) GetPerformanceMetrics(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getPerformanceMetrics", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceMetrics: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetProjectMemoryUsage returns memory used by the current project.
func (e *Engine) GetProjectMemoryUsage(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectMemoryUsage", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectMemoryUsage: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetDiskSpace returns available disk space for a given drive path.
func (e *Engine) GetDiskSpace(ctx context.Context, drivePath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"drivePath": drivePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getDiskSpace", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDiskSpace: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetOpenProjectCount returns the number of open projects.
func (e *Engine) GetOpenProjectCount(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getOpenProjectCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetOpenProjectCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetLoadedPlugins lists loaded plugins and extensions.
func (e *Engine) GetLoadedPlugins(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getLoadedPlugins", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLoadedPlugins: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Performance
// ---------------------------------------------------------------------------

// GetDroppedFrameCount returns the number of dropped frames during playback.
func (e *Engine) GetDroppedFrameCount(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getDroppedFrameCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDroppedFrameCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ResetDroppedFrameCount resets the dropped frame counter.
func (e *Engine) ResetDroppedFrameCount(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "resetDroppedFrameCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetDroppedFrameCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTimelineRenderStatus returns the render status per segment for a sequence.
func (e *Engine) GetTimelineRenderStatus(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineRenderStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineRenderStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetEstimatedRenderTime returns an estimated render time for a sequence.
func (e *Engine) GetEstimatedRenderTime2(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEstimatedRenderTime2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEstimatedRenderTime2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetSequenceComplexity rates the complexity of a sequence.
func (e *Engine) GetSequenceComplexity(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceComplexity", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceComplexity: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Diagnostics
// ---------------------------------------------------------------------------

// GetPremiereVersion returns detailed Premiere Pro version information.
func (e *Engine) GetPremiereVersion(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getPremiereVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetPremiereVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetInstalledPlugins2 lists all installed plugins with versions.
func (e *Engine) GetInstalledPlugins2(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getInstalledPlugins2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetInstalledPlugins2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetInstalledEffects2 lists all available effects.
func (e *Engine) GetInstalledEffects2(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getInstalledEffects2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetInstalledEffects2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetInstalledTransitions2 lists all available transitions.
func (e *Engine) GetInstalledTransitions2(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getInstalledTransitions2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetInstalledTransitions2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckProjectIntegrity performs a basic project integrity check.
func (e *Engine) CheckProjectIntegrity(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "checkProjectIntegrity", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckProjectIntegrity: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Error Handling
// ---------------------------------------------------------------------------

// GetLastError returns the last ExtendScript error.
func (e *Engine) GetLastError(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getLastError", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLastError: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClearErrors clears the error state.
func (e *Engine) ClearErrors(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearErrors", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearErrors: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetErrorLogging enables or disables error logging to a file.
func (e *Engine) SetErrorLogging(ctx context.Context, enabled bool, logPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
		"logPath": logPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "setErrorLogging", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetErrorLogging: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetErrorLog returns recent error log entries.
func (e *Engine) GetErrorLog(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getErrorLog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetErrorLog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Debug Tools
// ---------------------------------------------------------------------------

// EnableDebugMode toggles debug mode for verbose logging.
func (e *Engine) EnableDebugMode(ctx context.Context, enabled bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
	})
	result, err := e.premiere.EvalCommand(ctx, "enableDebugMode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EnableDebugMode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetDebugLog returns the debug log contents.
func (e *Engine) GetDebugLog(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getDebugLog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDebugLog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// DumpProjectState returns a full project state dump for debugging.
func (e *Engine) DumpProjectState(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "dumpProjectState", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DumpProjectState: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// DumpSequenceState returns a full sequence state dump for debugging.
func (e *Engine) DumpSequenceState(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "dumpSequenceState", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DumpSequenceState: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// TestBridgeConnection tests the CEP panel connection.
func (e *Engine) TestBridgeConnection(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "testBridgeConnection", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TestBridgeConnection: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Health Checks
// ---------------------------------------------------------------------------

// HealthCheck performs a full system health check.
func (e *Engine) HealthCheck(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "healthCheck", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("HealthCheck: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetServiceStatus returns the status of all MCP services.
func (e *Engine) GetServiceStatus(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getServiceStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetServiceStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetBridgeLatency measures the bridge round-trip time.
func (e *Engine) GetBridgeLatency(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getBridgeLatency", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetBridgeLatency: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetExtendScriptVersion returns the ExtendScript engine version.
func (e *Engine) GetExtendScriptVersion(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getExtendScriptVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExtendScriptVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Cleanup
// ---------------------------------------------------------------------------

// CleanTempFiles cleans temporary files created by the MCP bridge.
func (e *Engine) CleanTempFiles(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "cleanTempFiles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CleanTempFiles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// OptimizeProject runs project optimization tasks (consolidate, clean cache).
func (e *Engine) OptimizeProject(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "optimizeProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OptimizeProject: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
