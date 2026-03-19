package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Sequence Presets
// ---------------------------------------------------------------------------

// ListSequencePresets lists all available .sqpreset files.
func (e *Engine) ListSequencePresets(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listSequencePresets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListSequencePresets: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateSequenceFromPreset creates a new sequence from a preset file using QE DOM.
func (e *Engine) CreateSequenceFromPreset(ctx context.Context, name, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSequenceFromPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSequenceFromPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportSequencePreset exports current sequence settings as a preset file.
func (e *Engine) ExportSequencePreset(ctx context.Context, sequenceIndex int, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportSequencePreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportSequencePreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Effect Presets
// ---------------------------------------------------------------------------

// ListEffectPresets lists all available .ffx effect preset files.
func (e *Engine) ListEffectPresets(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listEffectPresets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListEffectPresets: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ApplyEffectPreset applies an effect preset to a clip.
func (e *Engine) ApplyEffectPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyEffectPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyEffectPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SaveEffectPreset saves a clip's effects as a preset.
func (e *Engine) SaveEffectPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, presetName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"presetName": presetName,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveEffectPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveEffectPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Export Presets
// ---------------------------------------------------------------------------

// ListExportPresetsFromDisk lists all available .epr export preset files.
func (e *Engine) ListExportPresetsFromDisk(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listExportPresetsFromDisk", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListExportPresetsFromDisk: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateExportPreset creates a custom export preset.
func (e *Engine) CreateExportPreset(ctx context.Context, settingsJSON, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"settingsJSON": settingsJSON,
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "createExportPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateExportPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetExportPresetDetails returns preset details (codec, bitrate, resolution).
func (e *Engine) GetExportPresetDetails(ctx context.Context, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getExportPresetDetails", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExportPresetDetails: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Templates
// ---------------------------------------------------------------------------

// SaveAsTemplate saves the current project as a template.
func (e *Engine) SaveAsTemplate(ctx context.Context, templatePath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"templatePath": templatePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveAsTemplate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveAsTemplate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateFromTemplate creates a new project from a template.
func (e *Engine) CreateFromTemplate(ctx context.Context, templatePath, projectPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"templatePath": templatePath,
		"projectPath": projectPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createFromTemplate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateFromTemplate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Keyboard Shortcuts
// ---------------------------------------------------------------------------

// GetKeyboardShortcuts lists assigned keyboard shortcuts.
func (e *Engine) GetKeyboardShortcuts(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getKeyboardShortcuts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetKeyboardShortcuts: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExecuteMenuCommand executes a menu command by path (e.g., "File/Save").
func (e *Engine) ExecuteMenuCommand(ctx context.Context, menuPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"menuPath": menuPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "executeMenuCommand", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExecuteMenuCommand: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Workflow / Ingest Presets
// ---------------------------------------------------------------------------

// CreateIngestPreset creates an ingest preset for transcode on import.
func (e *Engine) CreateIngestPreset(ctx context.Context, name, settingsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"settingsJSON": settingsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "createIngestPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateIngestPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetIngestSettings returns current ingest settings.
func (e *Engine) GetIngestSettings(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getIngestSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetIngestSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetIngestSettings enables or disables ingest with a specified preset.
func (e *Engine) SetIngestSettings(ctx context.Context, enabled bool, preset string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"enabled": enabled,
		"preset": preset,
	})
	result, err := e.premiere.EvalCommand(ctx, "setIngestSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetIngestSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Presets
// ---------------------------------------------------------------------------

// SaveClipPreset saves clip settings (speed, effects, motion) as a preset.
func (e *Engine) SaveClipPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveClipPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveClipPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ApplyClipPreset applies a saved clip preset.
func (e *Engine) ApplyClipPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, presetName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"presetName": presetName,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyClipPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyClipPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ListClipPresets lists saved clip presets.
func (e *Engine) ListClipPresets(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listClipPresets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListClipPresets: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Batch Operations (extended)
// ---------------------------------------------------------------------------

// BatchRename renames clips on a track using a pattern (e.g., "Shot_001", "Shot_002").
func (e *Engine) BatchRename(ctx context.Context, trackType string, trackIndex int, pattern string, startNumber int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"pattern": pattern,
		"startNumber": startNumber,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchRename", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchRename: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchSetDuration sets all clips on a track to the same duration.
func (e *Engine) BatchSetDuration(ctx context.Context, trackType string, trackIndex int, durationSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"durationSeconds": durationSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchSetDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchSetDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchSetSpeed sets speed on all clips on a track.
func (e *Engine) BatchSetSpeed(ctx context.Context, trackType string, trackIndex int, speed float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"speed": speed,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchSetSpeed", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchSetSpeed: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchApplyTransitions applies a transition to all cuts on a track.
func (e *Engine) BatchApplyTransitions(ctx context.Context, trackIndex int, transitionName string, duration float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"transitionName": transitionName,
		"duration": duration,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchApplyTransitions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchApplyTransitions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchExportFrames exports the first frame of each clip on a track.
func (e *Engine) BatchExportFrames(ctx context.Context, trackIndex int, outputDir, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"outputDir": outputDir,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchExportFrames", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchExportFrames: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Timeline Templates
// ---------------------------------------------------------------------------

// SaveTimelineTemplate saves the current timeline as a template.
func (e *Engine) SaveTimelineTemplate(ctx context.Context, name, description string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"description": description,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveTimelineTemplate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveTimelineTemplate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ApplyTimelineTemplate applies a saved timeline template.
func (e *Engine) ApplyTimelineTemplate(ctx context.Context, templateName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"templateName": templateName,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyTimelineTemplate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyTimelineTemplate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ListTimelineTemplates lists available timeline templates.
func (e *Engine) ListTimelineTemplates(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listTimelineTemplates", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListTimelineTemplates: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Macro Recording
// ---------------------------------------------------------------------------

// StartMacroRecording starts recording actions as a macro.
func (e *Engine) StartMacroRecording(ctx context.Context, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "startMacroRecording", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StartMacroRecording: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// StopMacroRecording stops recording and saves the macro.
func (e *Engine) StopMacroRecording(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "stopMacroRecording", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StopMacroRecording: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// PlayMacro plays back a recorded macro.
func (e *Engine) PlayMacro(ctx context.Context, name string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
	})
	result, err := e.premiere.EvalCommand(ctx, "playMacro", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PlayMacro: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
