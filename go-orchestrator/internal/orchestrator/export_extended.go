package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
)

// ---------------------------------------------------------------------------
// Extended Export & Render operations
// ---------------------------------------------------------------------------

// ExportDirect performs a synchronous export using exportAsMediaDirect.
func (e *Engine) ExportDirect(ctx context.Context, params *ExportDirectParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export direct: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export direct: output_path must not be empty")
	}
	if params.PresetPath == "" {
		return nil, fmt.Errorf("export direct: preset_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportDirect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportDirect: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportViaAME queues an export job through Adobe Media Encoder.
func (e *Engine) ExportViaAME(ctx context.Context, params *ExportViaAMEParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export via AME: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export via AME: output_path must not be empty")
	}
	if params.PresetPath == "" {
		return nil, fmt.Errorf("export via AME: preset_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportViaAME", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportViaAME: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportFrame exports the current frame from the active sequence as an image.
func (e *Engine) ExportFrame(ctx context.Context, params *ExportFrameParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export frame: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export frame: output_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportFrame: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportAAF exports a sequence as an AAF file.
func (e *Engine) ExportAAF(ctx context.Context, params *ExportAAFParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export AAF: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export AAF: output_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportAAF", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportAAF: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportOMF exports a sequence as an OMF file.
func (e *Engine) ExportOMF(ctx context.Context, params *ExportOMFParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export OMF: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export OMF: output_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportOMF", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportOMF: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportFCPXML exports the active sequence as Final Cut Pro XML.
func (e *Engine) ExportFCPXML(ctx context.Context, outputPath string) (*GenericExportResult, error) {
	if outputPath == "" {
		return nil, fmt.Errorf("export FCPXML: output_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportFCPXML", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportFCPXML: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportProjectAsXML exports the entire project as XML.
func (e *Engine) ExportProjectAsXML(ctx context.Context, outputPath string) (*GenericExportResult, error) {
	if outputPath == "" {
		return nil, fmt.Errorf("export project XML: output_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportProjectAsXML", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportProjectAsXML: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// GetExporters lists all available exporters.
func (e *Engine) GetExporters(ctx context.Context) (*ExporterListResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getExporters", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExporters: %w", err)
	}
	_ = result
	return &ExporterListResult{}, nil
}

// GetExportPresets returns presets for a specific exporter.
func (e *Engine) GetExportPresets(ctx context.Context, exporterIndex int) (*ExportPresetListResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"exporterIndex": exporterIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getExportPresets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExportPresets: %w", err)
	}
	_ = result
	return &ExportPresetListResult{}, nil
}

// StartAMEBatch starts the Adobe Media Encoder render queue.
func (e *Engine) StartAMEBatch(ctx context.Context) (*GenericExportResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "startAMEBatch", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StartAMEBatch: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// LaunchAME launches Adobe Media Encoder.
func (e *Engine) LaunchAME(ctx context.Context) (*GenericExportResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "launchAME", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LaunchAME: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// ExportAudioOnly exports only the audio from a sequence.
func (e *Engine) ExportAudioOnly(ctx context.Context, params *ExportAudioOnlyParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export audio only: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export audio only: output_path must not be empty")
	}
	if params.PresetPath == "" {
		return nil, fmt.Errorf("export audio only: preset_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportAudioOnly", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportAudioOnly: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}

// GetExportProgress returns current export/render progress.
func (e *Engine) GetExportProgress(ctx context.Context) (*ExportProgressResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	_, err := e.premiere.EvalCommand(ctx, "getExportProgress", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExportProgress: %w", err)
	}
	return &ExportProgressResult{Status: "success"}, nil
}

// RenderSequencePreview renders preview frames for a time range.
func (e *Engine) RenderSequencePreview(ctx context.Context, params *RenderPreviewParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("render sequence preview: params must not be nil")
	}
	if params.OutSeconds <= params.InSeconds {
		return nil, fmt.Errorf("render sequence preview: out_seconds must be greater than in_seconds")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "renderSequencePreview", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RenderSequencePreview: %w", err)
	}
	return &GenericExportResult{Status: "success", OutputPath: result}, nil
}
