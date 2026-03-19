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
			gomcp.WithDescription("Export a sequence synchronously (blocking) using Premiere Pro's direct export pipeline. The call blocks until the export file is written to disk. Use this for automated workflows where you need to wait for the file before proceeding. For background/non-blocking export, use premiere_export_via_ame instead. Requires an .epr export preset file -- use premiere_get_exporters and premiere_get_export_presets to find available presets."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index to export. Use -1 (default) to export the currently active sequence. Use premiere_get_sequence_list to find sequence indices."),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported output file (e.g. '/Users/me/exports/final.mp4'). The file extension should match the format specified in the preset."),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .epr (Export Preset) file that defines codec, resolution, bitrate, and other encoding settings. Find presets in Premiere's presets folder or use premiere_get_export_presets."),
			),
			gomcp.WithNumber("work_area_type",
				gomcp.Description("Which portion of the sequence to export: 0 = entire sequence (default), 1 = in-point to out-point only, 2 = work area bar range only. Set in/out points with premiere_set_in_point/premiere_set_out_point first."),
			),
		),
		makeExportDirectHandler(orch, logger),
	)

	// premiere_export_via_ame — async export via Adobe Media Encoder
	s.AddTool(
		gomcp.NewTool("premiere_export_via_ame",
			gomcp.WithDescription("Queue an export job in Adobe Media Encoder (AME) for asynchronous rendering. AME must be installed and running (use premiere_launch_ame first). The job is added to AME's queue but does not start automatically -- call premiere_start_ame_batch to begin rendering. This frees Premiere Pro for continued editing while AME renders in the background."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index to export. Use -1 (default) for the currently active sequence."),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported output file (e.g. '/Users/me/exports/final.mp4')."),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .epr (Export Preset) file defining codec, resolution, bitrate, etc."),
			),
			gomcp.WithNumber("work_area_type",
				gomcp.Description("Which portion to export: 0 = entire sequence (default), 1 = in-to-out points, 2 = work area range."),
			),
			gomcp.WithBoolean("remove_on_done",
				gomcp.Description("If true, automatically remove the job from AME's queue after it finishes rendering. If false (default), the completed job remains in the queue for review."),
			),
		),
		makeExportViaAMEHandler(orch, logger),
	)

	// premiere_export_frame — export current frame as image
	s.AddTool(
		gomcp.NewTool("premiere_export_frame",
			gomcp.WithDescription("Export the current frame at the playhead position from the active sequence as a still image file. Move the playhead to the desired frame first using premiere_set_playhead_position. Useful for creating thumbnails, reference images, or frame grabs for review."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported image (e.g. '/Users/me/frames/thumbnail.png'). The file extension should match the chosen format."),
			),
			gomcp.WithString("format",
				gomcp.Description("Image format for the exported frame (default: 'PNG'). PNG is lossless with transparency support. JPEG is smaller but lossy with no transparency."),
				gomcp.Enum("PNG", "JPEG"),
			),
		),
		makeExportFrameHandler(orch, logger),
	)

	// premiere_export_aaf — export as AAF
	s.AddTool(
		gomcp.NewTool("premiere_export_aaf",
			gomcp.WithDescription("Export a sequence as an AAF (Advanced Authoring Format) file for interchange with audio post-production tools like Avid Pro Tools, Nuendo, or Logic Pro. AAF preserves the timeline structure (clip positions, fades, levels) so an audio engineer can continue working from the same edit. The exported file includes embedded or referenced audio media."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index to export. Use -1 (default) for the currently active sequence."),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported AAF file (e.g. '/Users/me/exports/audio_mix.aaf'). Should end with .aaf."),
			),
			gomcp.WithBoolean("mixdown",
				gomcp.Description("If true, mix down video to a single track in the AAF. If false (default), preserve the original track structure."),
			),
			gomcp.WithBoolean("explode",
				gomcp.Description("If true, split multi-channel audio clips into separate mono clips. If false (default), keep the original channel layout. Useful when the DAW expects mono tracks."),
			),
			gomcp.WithNumber("sample_rate",
				gomcp.Description("Audio sample rate in Hz for the exported media (default: 48000). Standard values: 44100, 48000, 96000. Should match the project's audio sample rate."),
			),
			gomcp.WithNumber("bits_per_sample",
				gomcp.Description("Audio bit depth for the exported media (default: 16). Values: 16 (CD quality), 24 (professional standard), 32 (float, highest quality)."),
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
			gomcp.WithDescription("List all available export format modules (exporters) registered in Premiere Pro, such as H.264, HEVC, ProRes, DNxHR, etc. Each exporter has an index and name. Use the index with premiere_get_export_presets to list the available presets (resolution/bitrate combinations) for that format."),
		),
		makeGetExportersHandler(orch, logger),
	)

	// premiere_get_export_presets — get presets for an exporter
	s.AddTool(
		gomcp.NewTool("premiere_get_export_presets",
			gomcp.WithDescription("Get all available export presets for a specific exporter format. Returns preset names and .epr file paths that can be used with premiere_export_direct or premiere_export_via_ame. Call premiere_get_exporters first to find the exporter_index for your desired format (e.g. H.264)."),
			gomcp.WithNumber("exporter_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the exporter format. Obtain from premiere_get_exporters (e.g. 0 might be H.264, 1 might be HEVC, etc.)."),
			),
		),
		makeGetExportPresetsHandler(orch, logger),
	)

	// premiere_start_ame_batch — start AME render queue
	s.AddTool(
		gomcp.NewTool("premiere_start_ame_batch",
			gomcp.WithDescription("Start rendering all queued jobs in Adobe Media Encoder's render queue. Jobs must be queued first using premiere_export_via_ame. Rendering happens in the background, freeing Premiere Pro for editing. Use premiere_get_export_progress to check render status."),
		),
		makeStartAMEBatchHandler(orch, logger),
	)

	// premiere_launch_ame — launch Adobe Media Encoder
	s.AddTool(
		gomcp.NewTool("premiere_launch_ame",
			gomcp.WithDescription("Launch Adobe Media Encoder (AME) application. AME must be running before you can queue exports with premiere_export_via_ame. If AME is already running, this is a no-op. AME may take 10-30 seconds to fully launch and be ready to receive jobs."),
		),
		makeLaunchAMEHandler(orch, logger),
	)

	// premiere_export_audio_only — export audio only
	s.AddTool(
		gomcp.NewTool("premiere_export_audio_only",
			gomcp.WithDescription("Export only the audio tracks from a sequence, ignoring all video content. Internally mutes video tracks during export and restores them afterward. Useful for creating audio-only deliverables (podcast episodes, radio spots, music tracks) or sending audio stems to a mixing engineer. Use an audio-specific export preset (WAV, MP3, AAC)."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based sequence index to export. Use -1 (default) for the currently active sequence."),
			),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the exported audio file (e.g. '/Users/me/exports/podcast_ep1.wav'). The file extension should match the preset format."),
			),
			gomcp.WithString("preset_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to an audio-format export preset (.epr), e.g. a WAV or MP3 preset. Use premiere_get_export_presets to find available audio presets."),
			),
		),
		makeExportAudioOnlyHandler(orch, logger),
	)

	// premiere_get_export_progress — get export progress
	s.AddTool(
		gomcp.NewTool("premiere_get_export_progress",
			gomcp.WithDescription("Check the current export/render progress from Adobe Media Encoder. Returns the AME queue status and any available progress information. Note: detailed percentage progress may be limited depending on the ExtendScript bridge capabilities. Useful for polling export completion after premiere_start_ame_batch."),
		),
		makeGetExportProgressHandler(orch, logger),
	)

	// premiere_render_sequence_preview — render sequence preview via QE work area
	s.AddTool(
		gomcp.NewTool("premiere_render_sequence_preview",
			gomcp.WithDescription("Render preview files for a time range on the active sequence using the QE DOM work area method. Sets the work area to the specified range and triggers a full quality render. Unlike premiere_render_preview (which uses renderPreviewArea), this method sets in/out on the QE sequence and renders via renderWorkArea, which can produce different results depending on render settings."),
			gomcp.WithNumber("in_seconds",
				gomcp.Required(),
				gomcp.Description("Start time in seconds for the preview render range. Must be >= 0 and less than out_seconds."),
			),
			gomcp.WithNumber("out_seconds",
				gomcp.Required(),
				gomcp.Description("End time in seconds for the preview render range. Must be greater than in_seconds."),
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
