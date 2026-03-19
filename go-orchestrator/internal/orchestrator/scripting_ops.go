package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Scripting, Expression & Programmatic Control Operations
// ---------------------------------------------------------------------------

// --- Script Execution ---

func (e *Engine) EvaluateExpression(ctx context.Context, expression string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"expression": expression,
	})
	result, err := e.premiere.EvalCommand(ctx, "evaluateExpression", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EvaluateExpression: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteScript(ctx context.Context, scriptPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"scriptPath": scriptPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteScript: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteScriptWithArgs(ctx context.Context, scriptPath string, argsJSON string) (*GenericResult, error) {
	cmdArgs, _ := json.Marshal(map[string]any{
		"scriptPath": scriptPath,
		"argsJSON":   argsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeScriptWithArgs", string(cmdArgs))
	if err != nil {
		return nil, fmt.Errorf("ExecuteScriptWithArgs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetScriptResult(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getScriptResult", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetScriptResult: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ListAvailableScripts(ctx context.Context, directory string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"directory": directory,
	})
	result, err := e.premiere.EvalCommand(ctx, "listAvailableScripts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListAvailableScripts: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Variable/State Management ---

func (e *Engine) SetGlobalVariable(ctx context.Context, name string, value string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setGlobalVariable", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetGlobalVariable: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetGlobalVariable(ctx context.Context, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "getGlobalVariable", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetGlobalVariable: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ListGlobalVariables(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listGlobalVariables", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListGlobalVariables: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ClearGlobalVariables(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearGlobalVariables", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearGlobalVariables: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Conditional Operations ---

func (e *Engine) IfClipExists(ctx context.Context, trackType string, trackIndex, clipIndex int, thenScript, elseScript string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"thenScript": thenScript,
		"elseScript": elseScript,
	})
	result, err := e.premiere.EvalCommand(ctx, "ifClipExists", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IfClipExists: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) IfSequenceOpen(ctx context.Context, thenScript, elseScript string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"thenScript": thenScript,
		"elseScript": elseScript,
	})
	result, err := e.premiere.EvalCommand(ctx, "ifSequenceOpen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IfSequenceOpen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) IfProjectOpen(ctx context.Context, thenScript, elseScript string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"thenScript": thenScript,
		"elseScript": elseScript,
	})
	result, err := e.premiere.EvalCommand(ctx, "ifProjectOpen", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IfProjectOpen: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) WhileCondition(ctx context.Context, conditionScript, bodyScript string, maxIterations int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"conditionScript": conditionScript,
		"bodyScript": bodyScript,
		"maxIterations": maxIterations,
	})
	result, err := e.premiere.EvalCommand(ctx, "whileCondition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WhileCondition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Batch Scripting ---

func (e *Engine) ExecuteBatch(ctx context.Context, scriptsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"scriptsJSON": scriptsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeBatch", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteBatch: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteParallel(ctx context.Context, scriptsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"scriptsJSON": scriptsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeParallel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteParallel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteWithRetry(ctx context.Context, script string, maxRetries int, delayMs int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script": script,
		"maxRetries": maxRetries,
		"delayMs": delayMs,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeWithRetry", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteWithRetry: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteWithTimeout(ctx context.Context, script string, timeoutMs int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script": script,
		"timeoutMs": timeoutMs,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeWithTimeout", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteWithTimeout: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Timer/Scheduling ---

func (e *Engine) ScheduleScript(ctx context.Context, script string, delayMs int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script": script,
		"delayMs": delayMs,
	})
	result, err := e.premiere.EvalCommand(ctx, "scheduleScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ScheduleScript: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ScheduleRepeating(ctx context.Context, script string, intervalMs, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script": script,
		"intervalMs": intervalMs,
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "scheduleRepeating", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ScheduleRepeating: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CancelScheduledScript(ctx context.Context, scheduleID string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"scheduleID": scheduleID,
	})
	result, err := e.premiere.EvalCommand(ctx, "cancelScheduledScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CancelScheduledScript: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GetScheduledScripts(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getScheduledScripts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetScheduledScripts: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Data Operations ---

func (e *Engine) ReadJSONFile(ctx context.Context, filePath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "readJSONFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReadJSONFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) WriteJSONFile(ctx context.Context, filePath, data string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
		"data": data,
	})
	result, err := e.premiere.EvalCommand(ctx, "writeJSONFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WriteJSONFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ReadCSVFile(ctx context.Context, filePath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "readCSVFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReadCSVFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) WriteCSVFile(ctx context.Context, filePath, headers, rows string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
		"headers": headers,
		"rows": rows,
	})
	result, err := e.premiere.EvalCommand(ctx, "writeCSVFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WriteCSVFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ReadTextFile(ctx context.Context, filePath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "readTextFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReadTextFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) WriteTextFile(ctx context.Context, filePath, content string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
		"content": content,
	})
	result, err := e.premiere.EvalCommand(ctx, "writeTextFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WriteTextFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AppendTextFile(ctx context.Context, filePath, content string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
		"content": content,
	})
	result, err := e.premiere.EvalCommand(ctx, "appendTextFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AppendTextFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- System Integration ---

func (e *Engine) OpenURL(ctx context.Context, url string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"url": url,
	})
	result, err := e.premiere.EvalCommand(ctx, "openURL", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OpenURL: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExecuteSystemCommand(ctx context.Context, command string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"command": command,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeSystemCommand", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteSystemCommand: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
