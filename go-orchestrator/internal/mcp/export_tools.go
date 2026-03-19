package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerExportTools2 registers all extended export and render MCP tools.
// This complements the basic premiere_export tool already in registerExportTools.
func registerExportTools2(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// premiere_export_direct — synchronous direct export
	s.AddTool(
		gomcp.NewTool("premiere_export_direct",
			gomcp.WithDescription("Export a sequence synchronously using exportAsMediaDirect. Blocks until the export is complete."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index (-1 for active sequence, default: -1)"),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported output"),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .epr export preset file"),
			),
			gomcp.WithNumber("work_area_type",
				gomcp.Description("Work area type: 0=entire sequence (default), 1=in-to-out, 2=work area"),
			),
		),
		makeExportDirectHandler(orch, logger),
	)

	// premiere_export_via_ame — async export via Adobe Media Encoder
	s.AddTool(
		gomcp.NewTool("premiere_export_via_ame",
			gomcp.WithDescription("Queue an export job in Adobe Media Encoder (async). AME must be installed."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index (-1 for active sequence, default: -1)"),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported output"),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .epr export preset file"),
			),
			gomcp.WithNumber("work_area_type",
				gomcp.Description("Work area type: 0=entire sequence (default), 1=in-to-out, 2=work area"),
			),
			gomcp.WithBoolean("remove_on_done",
				gomcp.Description("Remove from AME queue after completion (default: false)"),
			),
		),
		makeExportViaAMEHandler(orch, logger),
	)

	// premiere_export_frame — export current frame as image
	s.AddTool(
		gomcp.NewTool("premiere_export_frame",
			gomcp.WithDescription("Export the current frame (at the playhead position) from the active sequence as a PNG or JPEG image."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported image"),
			),
			gomcp.WithString("format",
				gomcp.Description("Image format: PNG (default) or JPEG"),
				gomcp.Enum("PNG", "JPEG"),
			),
		),
		makeExportFrameHandler(orch, logger),
	)

	// premiere_export_aaf — export as AAF
	s.AddTool(
		gomcp.NewTool("premiere_export_aaf",
			gomcp.WithDescription("Export a sequence as an AAF (Advanced Authoring Format) file for interchange with Pro Tools and other DAWs."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index (-1 for active sequence, default: -1)"),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported AAF"),
			),
			gomcp.WithBoolean("mixdown",
				gomcp.Description("Mix down video to a single track (default: false)"),
			),
			gomcp.WithBoolean("explode",
				gomcp.Description("Explode multi-channel audio to mono (default: false)"),
			),
			gomcp.WithNumber("sample_rate",
				gomcp.Description("Audio sample rate in Hz (default: 48000)"),
			),
			gomcp.WithNumber("bits_per_sample",
				gomcp.Description("Audio bit depth (default: 16)"),
			),
		),
		makeExportAAFHandler(orch, logger),
	)

	// premiere_export_omf — export as OMF
	s.AddTool(
		gomcp.NewTool("premiere_export_omf",
			gomcp.WithDescription("Export a sequence as an OMF (Open Media Framework) file for audio post-production interchange."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index (-1 for active sequence, default: -1)"),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported OMF"),
			),
			gomcp.WithNumber("sample_rate",
				gomcp.Description("Audio sample rate in Hz (default: 48000)"),
			),
			gomcp.WithNumber("bits_per_sample",
				gomcp.Description("Audio bit depth (default: 16)"),
			),
			gomcp.WithNumber("handle_frames",
				gomcp.Description("Number of handle frames to include (default: 0)"),
			),
			gomcp.WithBoolean("encapsulate",
				gomcp.Description("Encapsulate audio within the OMF file (default: true)"),
			),
		),
		makeExportOMFHandler(orch, logger),
	)

	// premiere_export_fcpxml — export as Final Cut Pro XML
	s.AddTool(
		gomcp.NewTool("premiere_export_fcpxml",
			gomcp.WithDescription("Export the active sequence as a Final Cut Pro XML file for interchange with Final Cut Pro."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported FCPXML"),
			),
		),
		makeExportFCPXMLHandler(orch, logger),
	)

	// premiere_export_project_xml — export project as XML
	s.AddTool(
		gomcp.NewTool("premiere_export_project_xml",
			gomcp.WithDescription("Export the entire Premiere Pro project as an XML interchange file."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported XML"),
			),
		),
		makeExportProjectXMLHandler(orch, logger),
	)

	// premiere_get_exporters — list available exporters
	s.AddTool(
		gomcp.NewTool("premiere_get_exporters",
			gomcp.WithDescription("List all available exporters (codecs/formats) registered in Premiere Pro's encoder."),
		),
		makeGetExportersHandler(orch, logger),
	)

	// premiere_get_export_presets — get presets for an exporter
	s.AddTool(
		gomcp.NewTool("premiere_get_export_presets",
			gomcp.WithDescription("Get all available export presets for a specific exporter. Use premiere_get_exporters first to find the exporter index."),
			gomcp.WithNumber("exporter_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the exporter (from premiere_get_exporters)"),
			),
		),
		makeGetExportPresetsHandler(orch, logger),
	)

	// premiere_start_ame_batch — start AME render queue
	s.AddTool(
		gomcp.NewTool("premiere_start_ame_batch",
			gomcp.WithDescription("Start rendering all queued jobs in Adobe Media Encoder. Jobs must be queued first using premiere_export_via_ame."),
		),
		makeStartAMEBatchHandler(orch, logger),
	)

	// premiere_launch_ame — launch Adobe Media Encoder
	s.AddTool(
		gomcp.NewTool("premiere_launch_ame",
			gomcp.WithDescription("Launch Adobe Media Encoder application. Required before queuing exports via AME."),
		),
		makeLaunchAMEHandler(orch, logger),
	)

	// premiere_export_audio_only — export audio only
	s.AddTool(
		gomcp.NewTool("premiere_export_audio_only",
			gomcp.WithDescription("Export only the audio from a sequence. Temporarily mutes all video tracks during export."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index (-1 for active sequence, default: -1)"),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported audio"),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to an audio-only export preset (.epr)"),
			),
		),
		makeExportAudioOnlyHandler(orch, logger),
	)

	// premiere_get_export_progress — get export progress
	s.AddTool(
		gomcp.NewTool("premiere_get_export_progress",
			gomcp.WithDescription("Get the current export/render progress from Adobe Media Encoder. Note: progress reporting is limited in ExtendScript."),
		),
		makeGetExportProgressHandler(orch, logger),
	)

	// premiere_render_sequence_preview — render sequence preview via QE work area
	s.AddTool(
		gomcp.NewTool("premiere_render_sequence_preview",
			gomcp.WithDescription("Render preview for a time range by setting in/out points and triggering QE renderWorkArea. Unlike premiere_render_preview (which uses renderPreviewArea), this sets the work area and uses the QE DOM."),
			gomcp.WithNumber("in_seconds",
				gomcp.Required(),
				gomcp.Description("Start time in seconds for the preview render range"),
			),
			gomcp.WithNumber("out_seconds",
				gomcp.Required(),
				gomcp.Description("End time in seconds for the preview render range"),
			),
		),
		makeRenderSequencePreviewHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Handler constructors
// ---------------------------------------------------------------------------

func makeExportDirectHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_direct")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}

		params := &ExportDirectParams{
			SequenceIndex: gomcp.ParseInt(req, "sequence_index", -1),
			OutputPath:    outputPath,
			PresetPath:    presetPath,
			WorkAreaType:  gomcp.ParseInt(req, "work_area_type", 0),
		}

		result, err := orch.ExportDirect(ctx, params)
		if err != nil {
			logger.Error("export direct failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export direct: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportViaAMEHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_via_ame")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}

		params := &ExportViaAMEParams{
			SequenceIndex: gomcp.ParseInt(req, "sequence_index", -1),
			OutputPath:    outputPath,
			PresetPath:    presetPath,
			WorkAreaType:  gomcp.ParseInt(req, "work_area_type", 0),
			RemoveOnDone:  gomcp.ParseBoolean(req, "remove_on_done", false),
		}

		result, err := orch.ExportViaAME(ctx, params)
		if err != nil {
			logger.Error("export via AME failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export via AME: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportFrameHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_frame")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}

		params := &ExportFrameParams{
			OutputPath: outputPath,
			Format:     gomcp.ParseString(req, "format", "PNG"),
		}

		result, err := orch.ExportFrame(ctx, params)
		if err != nil {
			logger.Error("export frame failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export frame: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportAAFHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_aaf")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}

		params := &ExportAAFParams{
			SequenceIndex: gomcp.ParseInt(req, "sequence_index", -1),
			OutputPath:    outputPath,
			Mixdown:       gomcp.ParseBoolean(req, "mixdown", false),
			Explode:       gomcp.ParseBoolean(req, "explode", false),
			SampleRate:    gomcp.ParseInt(req, "sample_rate", 48000),
			BitsPerSample: gomcp.ParseInt(req, "bits_per_sample", 16),
		}

		result, err := orch.ExportAAF(ctx, params)
		if err != nil {
			logger.Error("export AAF failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export AAF: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportOMFHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_omf")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}

		params := &ExportOMFParams{
			SequenceIndex: gomcp.ParseInt(req, "sequence_index", -1),
			OutputPath:    outputPath,
			SampleRate:    gomcp.ParseInt(req, "sample_rate", 48000),
			BitsPerSample: gomcp.ParseInt(req, "bits_per_sample", 16),
			HandleFrames:  gomcp.ParseInt(req, "handle_frames", 0),
			Encapsulate:   gomcp.ParseBoolean(req, "encapsulate", true),
		}

		result, err := orch.ExportOMF(ctx, params)
		if err != nil {
			logger.Error("export OMF failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export OMF: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportFCPXMLHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_fcpxml")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}

		result, err := orch.ExportFCPXML(ctx, outputPath)
		if err != nil {
			logger.Error("export FCPXML failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export FCPXML: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportProjectXMLHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_project_xml")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}

		result, err := orch.ExportProjectAsXML(ctx, outputPath)
		if err != nil {
			logger.Error("export project XML failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export project XML: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetExportersHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_exporters")

		result, err := orch.GetExporters(ctx)
		if err != nil {
			logger.Error("get exporters failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get exporters: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetExportPresetsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_export_presets")

		exporterIndex := gomcp.ParseInt(req, "exporter_index", 0)

		result, err := orch.GetExportPresets(ctx, exporterIndex)
		if err != nil {
			logger.Error("get export presets failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get export presets: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeStartAMEBatchHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_start_ame_batch")

		result, err := orch.StartAMEBatch(ctx)
		if err != nil {
			logger.Error("start AME batch failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to start AME batch: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeLaunchAMEHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_launch_ame")

		result, err := orch.LaunchAME(ctx)
		if err != nil {
			logger.Error("launch AME failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to launch AME: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeExportAudioOnlyHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_export_audio_only")

		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}

		params := &ExportAudioOnlyParams{
			SequenceIndex: gomcp.ParseInt(req, "sequence_index", -1),
			OutputPath:    outputPath,
			PresetPath:    presetPath,
		}

		result, err := orch.ExportAudioOnly(ctx, params)
		if err != nil {
			logger.Error("export audio only failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to export audio only: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetExportProgressHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_export_progress")

		result, err := orch.GetExportProgress(ctx)
		if err != nil {
			logger.Error("get export progress failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get export progress: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRenderSequencePreviewHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_render_sequence_preview")

		inSeconds := gomcp.ParseFloat64(req, "in_seconds", 0)
		outSeconds := gomcp.ParseFloat64(req, "out_seconds", 0)

		if outSeconds <= inSeconds {
			return gomcp.NewToolResultError("'out_seconds' must be greater than 'in_seconds'"), nil
		}

		params := &RenderPreviewParams{
			InSeconds:  inSeconds,
			OutSeconds: outSeconds,
		}

		result, err := orch.RenderSequencePreview(ctx, params)
		if err != nil {
			logger.Error("render preview failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to render preview: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}
