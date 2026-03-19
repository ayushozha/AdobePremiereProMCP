package orchestrator

import (
	"context"
	"fmt"

	"go.uber.org/zap"
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
	e.logger.Debug("export_direct: starting",
		zap.Int("sequence_index", params.SequenceIndex),
		zap.String("output", params.OutputPath),
		zap.String("preset", params.PresetPath),
		zap.Int("work_area_type", params.WorkAreaType),
	)
	// TODO: call through Premiere bridge once the bridge supports this method.
	return nil, fmt.Errorf("export direct: not yet implemented in bridge")
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
	e.logger.Debug("export_via_ame: queuing",
		zap.Int("sequence_index", params.SequenceIndex),
		zap.String("output", params.OutputPath),
		zap.String("preset", params.PresetPath),
		zap.Int("work_area_type", params.WorkAreaType),
		zap.Bool("remove_on_done", params.RemoveOnDone),
	)
	return nil, fmt.Errorf("export via AME: not yet implemented in bridge")
}

// ExportFrame exports the current frame from the active sequence as an image.
func (e *Engine) ExportFrame(ctx context.Context, params *ExportFrameParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export frame: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export frame: output_path must not be empty")
	}
	e.logger.Debug("export_frame: exporting",
		zap.String("output", params.OutputPath),
		zap.String("format", params.Format),
	)
	return nil, fmt.Errorf("export frame: not yet implemented in bridge")
}

// ExportAAF exports a sequence as an AAF file.
func (e *Engine) ExportAAF(ctx context.Context, params *ExportAAFParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export AAF: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export AAF: output_path must not be empty")
	}
	e.logger.Debug("export_aaf: exporting",
		zap.Int("sequence_index", params.SequenceIndex),
		zap.String("output", params.OutputPath),
		zap.Bool("mixdown", params.Mixdown),
		zap.Bool("explode", params.Explode),
		zap.Int("sample_rate", params.SampleRate),
		zap.Int("bits_per_sample", params.BitsPerSample),
	)
	return nil, fmt.Errorf("export AAF: not yet implemented in bridge")
}

// ExportOMF exports a sequence as an OMF file.
func (e *Engine) ExportOMF(ctx context.Context, params *ExportOMFParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export OMF: params must not be nil")
	}
	if params.OutputPath == "" {
		return nil, fmt.Errorf("export OMF: output_path must not be empty")
	}
	e.logger.Debug("export_omf: exporting",
		zap.Int("sequence_index", params.SequenceIndex),
		zap.String("output", params.OutputPath),
		zap.Int("sample_rate", params.SampleRate),
		zap.Int("bits_per_sample", params.BitsPerSample),
		zap.Int("handle_frames", params.HandleFrames),
		zap.Bool("encapsulate", params.Encapsulate),
	)
	return nil, fmt.Errorf("export OMF: not yet implemented in bridge")
}

// ExportFCPXML exports the active sequence as Final Cut Pro XML.
func (e *Engine) ExportFCPXML(ctx context.Context, outputPath string) (*GenericExportResult, error) {
	if outputPath == "" {
		return nil, fmt.Errorf("export FCPXML: output_path must not be empty")
	}
	e.logger.Debug("export_fcpxml: exporting", zap.String("output", outputPath))
	return nil, fmt.Errorf("export FCPXML: not yet implemented in bridge")
}

// ExportProjectAsXML exports the entire project as XML.
func (e *Engine) ExportProjectAsXML(ctx context.Context, outputPath string) (*GenericExportResult, error) {
	if outputPath == "" {
		return nil, fmt.Errorf("export project XML: output_path must not be empty")
	}
	e.logger.Debug("export_project_xml: exporting", zap.String("output", outputPath))
	return nil, fmt.Errorf("export project XML: not yet implemented in bridge")
}

// GetExporters lists all available exporters.
func (e *Engine) GetExporters(ctx context.Context) (*ExporterListResult, error) {
	e.logger.Debug("get_exporters: listing")
	return nil, fmt.Errorf("get exporters: not yet implemented in bridge")
}

// GetExportPresets returns presets for a specific exporter.
func (e *Engine) GetExportPresets(ctx context.Context, exporterIndex int) (*ExportPresetListResult, error) {
	e.logger.Debug("get_export_presets: listing", zap.Int("exporter_index", exporterIndex))
	return nil, fmt.Errorf("get export presets: not yet implemented in bridge")
}

// StartAMEBatch starts the Adobe Media Encoder render queue.
func (e *Engine) StartAMEBatch(ctx context.Context) (*GenericExportResult, error) {
	e.logger.Debug("start_ame_batch: starting")
	return nil, fmt.Errorf("start AME batch: not yet implemented in bridge")
}

// LaunchAME launches Adobe Media Encoder.
func (e *Engine) LaunchAME(ctx context.Context) (*GenericExportResult, error) {
	e.logger.Debug("launch_ame: launching")
	return nil, fmt.Errorf("launch AME: not yet implemented in bridge")
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
	e.logger.Debug("export_audio_only: exporting",
		zap.Int("sequence_index", params.SequenceIndex),
		zap.String("output", params.OutputPath),
		zap.String("preset", params.PresetPath),
	)
	return nil, fmt.Errorf("export audio only: not yet implemented in bridge")
}

// GetExportProgress returns current export/render progress.
func (e *Engine) GetExportProgress(ctx context.Context) (*ExportProgressResult, error) {
	e.logger.Debug("get_export_progress: checking")
	return nil, fmt.Errorf("get export progress: not yet implemented in bridge")
}

// RenderSequencePreview renders preview frames for a time range.
func (e *Engine) RenderSequencePreview(ctx context.Context, params *RenderPreviewParams) (*GenericExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("render sequence preview: params must not be nil")
	}
	if params.OutSeconds <= params.InSeconds {
		return nil, fmt.Errorf("render sequence preview: out_seconds must be greater than in_seconds")
	}
	e.logger.Debug("render_sequence_preview: rendering",
		zap.Float64("in_seconds", params.InSeconds),
		zap.Float64("out_seconds", params.OutSeconds),
	)
	return nil, fmt.Errorf("render sequence preview: not yet implemented in bridge")
}
