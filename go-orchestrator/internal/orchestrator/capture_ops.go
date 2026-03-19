package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
)

// ---------------------------------------------------------------------------
// Frame Capture Operations
// ---------------------------------------------------------------------------

// CaptureFrameAsBase64 captures the current frame at the playhead position
// and returns it as a base64-encoded PNG image.
func (e *Engine) CaptureFrameAsBase64(ctx context.Context) (*FrameCaptureResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	raw, err := e.premiere.EvalCommand(ctx, "captureFrameAsBase64", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CaptureFrameAsBase64: %w", err)
	}

	// Parse the JSON response from ExtendScript
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ImageBase64 string  `json:"image_base64"`
			Format      string  `json:"format"`
			Width       int     `json:"width"`
			Height      int     `json:"height"`
			Timecode    float64 `json:"timecode"`
		} `json:"data"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("CaptureFrameAsBase64: failed to parse response: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("CaptureFrameAsBase64: %s", resp.Error)
	}

	return &FrameCaptureResult{
		ImageBase64: resp.Data.ImageBase64,
		Format:      resp.Data.Format,
		Width:       resp.Data.Width,
		Height:      resp.Data.Height,
		Timecode:    resp.Data.Timecode,
	}, nil
}

// ---------------------------------------------------------------------------
// Secure ExtendScript Execution Operations
// ---------------------------------------------------------------------------

// ExecuteSecureScript runs an arbitrary ExtendScript string with optional
// security validation that blocks dangerous operations such as system calls,
// file deletion, and infinite loops.
func (e *Engine) ExecuteSecureScript(ctx context.Context, script string, validate bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script":   script,
		"validate": validate,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeSecureScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteSecureScript: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExecuteQEScript runs an arbitrary ExtendScript string with QE DOM access
// enabled. The QE (Quality Engineering) DOM provides access to internal
// Premiere Pro functionality not available through the standard DOM.
func (e *Engine) ExecuteQEScript(ctx context.Context, script string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"script": script,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeQEScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteQEScript: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
