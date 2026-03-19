package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// encH wraps a handler with debug logging for encoding tool invocations.
func encH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerEncodingTools registers all 30 encoding, format conversion, and media
// management MCP tools.
func registerEncodingTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Encoding Settings (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_get_export_settings_for_preset
	s.AddTool(gomcp.NewTool("premiere_get_export_settings_for_preset",
		gomcp.WithDescription("Get detailed encoding settings from an export preset file. Returns codec, bitrate, resolution, and audio settings."),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the export preset file (.epr)")),
	), encH(logger, "get_export_settings_for_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.GetExportSettingsForPreset(ctx, presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_create_custom_export_settings
	s.AddTool(gomcp.NewTool("premiere_create_custom_export_settings",
		gomcp.WithDescription("Create custom export settings with specific codec, bitrate, resolution, fps, and audio parameters."),
		gomcp.WithString("settings", gomcp.Required(), gomcp.Description("JSON object with export settings: {\"codec\":\"H.264\",\"bitrate\":10000,\"width\":1920,\"height\":1080,\"fps\":29.97,\"audioCodec\":\"AAC\",\"audioSampleRate\":48000,\"audioBitrate\":320,\"audioChannels\":2}")),
	), encH(logger, "create_custom_export_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		settings := gomcp.ParseString(req, "settings", "")
		if settings == "" {
			return gomcp.NewToolResultError("parameter 'settings' is required"), nil
		}
		result, err := orch.CreateCustomExportSettings(ctx, settings)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_available_codecs
	s.AddTool(gomcp.NewTool("premiere_get_available_codecs",
		gomcp.WithDescription("List all available video codecs supported by the installed Premiere Pro exporters."),
	), encH(logger, "get_available_codecs", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAvailableCodecs(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_get_available_audio_codecs
	s.AddTool(gomcp.NewTool("premiere_get_available_audio_codecs",
		gomcp.WithDescription("List all available audio codecs (AAC, MP3, PCM/WAV, AIFF, FLAC, etc.) with descriptions."),
	), encH(logger, "get_available_audio_codecs", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAvailableAudioCodecs(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_get_available_containers
	s.AddTool(gomcp.NewTool("premiere_get_available_containers",
		gomcp.WithDescription("List available container formats (MP4, MKV, MOV, AVI, MXF, WebM, etc.) with extensions and MIME types."),
	), encH(logger, "get_available_containers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAvailableContainers(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Format Conversion (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_convert_to_prores
	s.AddTool(gomcp.NewTool("premiere_convert_to_prores",
		gomcp.WithDescription("Convert a project item to Apple ProRes format. Supports 422, 4444, LT, and Proxy variants."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("variant", gomcp.Description("ProRes variant: 422, 4444, LT, or Proxy (default: 422)"),
			gomcp.Enum("422", "4444", "LT", "Proxy")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
	), encH(logger, "convert_to_prores", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ConvertToProRes(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			gomcp.ParseString(req, "variant", "422"),
			outputPath,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_convert_to_h264
	s.AddTool(gomcp.NewTool("premiere_convert_to_h264",
		gomcp.WithDescription("Convert a project item to H.264 format with configurable bitrate."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithNumber("bitrate", gomcp.Description("Video bitrate in kbps (default: 10000)")),
	), encH(logger, "convert_to_h264", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ConvertToH264(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			outputPath,
			gomcp.ParseInt(req, "bitrate", 10000),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_convert_to_h265
	s.AddTool(gomcp.NewTool("premiere_convert_to_h265",
		gomcp.WithDescription("Convert a project item to H.265/HEVC format with configurable bitrate."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithNumber("bitrate", gomcp.Description("Video bitrate in kbps (default: 8000)")),
	), encH(logger, "convert_to_h265", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ConvertToH265(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			outputPath,
			gomcp.ParseInt(req, "bitrate", 8000),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_convert_to_dnxhr
	s.AddTool(gomcp.NewTool("premiere_convert_to_dnxhr",
		gomcp.WithDescription("Convert a project item to DNxHR format. Supports LB, SQ, HQ, HQX, and 444 profiles."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("profile", gomcp.Description("DNxHR profile (default: HQ)"),
			gomcp.Enum("LB", "SQ", "HQ", "HQX", "444")),
	), encH(logger, "convert_to_dnxhr", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ConvertToDNxHR(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			outputPath,
			gomcp.ParseString(req, "profile", "HQ"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_convert_to_gif
	s.AddTool(gomcp.NewTool("premiere_convert_to_gif",
		gomcp.WithDescription("Export a sequence as an animated GIF with configurable width and frame rate."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output GIF file")),
		gomcp.WithNumber("width", gomcp.Description("Output width in pixels (default: 480)")),
		gomcp.WithNumber("fps", gomcp.Description("Frame rate for the GIF (default: 15)")),
	), encH(logger, "convert_to_gif", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ConvertToGIF(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseInt(req, "width", 480),
			gomcp.ParseInt(req, "fps", 15),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Thumbnail/Preview Generation (11-14)
	// -----------------------------------------------------------------------

	// 11. premiere_generate_clip_thumbnail
	s.AddTool(gomcp.NewTool("premiere_generate_clip_thumbnail",
		gomcp.WithDescription("Generate a thumbnail image from a project item at a specified time offset."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithNumber("time_offset", gomcp.Description("Time offset in seconds for the thumbnail (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output image file")),
	), encH(logger, "generate_clip_thumbnail", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.GenerateClipThumbnail(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			gomcp.ParseFloat64(req, "time_offset", 0),
			outputPath,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_generate_sequence_thumbnail
	s.AddTool(gomcp.NewTool("premiere_generate_sequence_thumbnail",
		gomcp.WithDescription("Generate a thumbnail image from a sequence at a specified time offset."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithNumber("time_offset", gomcp.Description("Time offset in seconds (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output image file")),
	), encH(logger, "generate_sequence_thumbnail", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.GenerateSequenceThumbnail(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseFloat64(req, "time_offset", 0),
			outputPath,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_generate_contact_sheet
	s.AddTool(gomcp.NewTool("premiere_generate_contact_sheet",
		gomcp.WithDescription("Generate a contact sheet (grid of thumbnails) from a project item at evenly spaced intervals."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output image file")),
		gomcp.WithNumber("cols", gomcp.Description("Number of columns in the grid (default: 4)")),
		gomcp.WithNumber("rows", gomcp.Description("Number of rows in the grid (default: 4)")),
	), encH(logger, "generate_contact_sheet", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.GenerateContactSheet(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			outputPath,
			gomcp.ParseInt(req, "cols", 4),
			gomcp.ParseInt(req, "rows", 4),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_generate_storyboard
	s.AddTool(gomcp.NewTool("premiere_generate_storyboard",
		gomcp.WithDescription("Generate storyboard frames from a sequence at regular time intervals."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute directory path for the output frame images")),
		gomcp.WithNumber("interval", gomcp.Description("Interval between frames in seconds (default: 5)")),
	), encH(logger, "generate_storyboard", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.GenerateStoryboard(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseFloat64(req, "interval", 5),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Media Analysis (15-20)
	// -----------------------------------------------------------------------

	// 15. premiere_analyze_media_codec
	s.AddTool(gomcp.NewTool("premiere_analyze_media_codec",
		gomcp.WithDescription("Get detailed codec analysis for a project item including video/audio codec, frame rate, and metadata."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
	), encH(logger, "analyze_media_codec", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AnalyzeMediaCodec(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_compare_media_specs
	s.AddTool(gomcp.NewTool("premiere_compare_media_specs",
		gomcp.WithDescription("Compare the technical specifications of two project items and report differences."),
		gomcp.WithNumber("item_index_1", gomcp.Required(), gomcp.Description("Zero-based index of the first project item")),
		gomcp.WithNumber("item_index_2", gomcp.Required(), gomcp.Description("Zero-based index of the second project item")),
	), encH(logger, "compare_media_specs", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CompareMediaSpecs(ctx,
			gomcp.ParseInt(req, "item_index_1", 0),
			gomcp.ParseInt(req, "item_index_2", 0),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_get_bit_rate_info
	s.AddTool(gomcp.NewTool("premiere_get_bit_rate_info",
		gomcp.WithDescription("Get bitrate information for a project item including VBR/CBR detection, average and estimated bitrate."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
	), encH(logger, "get_bit_rate_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetBitRateInfo(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_get_color_depth_info
	s.AddTool(gomcp.NewTool("premiere_get_color_depth_info",
		gomcp.WithDescription("Get color depth, bit depth, color space, and chroma subsampling information for a project item."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
	), encH(logger, "get_color_depth_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetColorDepthInfo(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_get_audio_specs_detailed
	s.AddTool(gomcp.NewTool("premiere_get_audio_specs_detailed",
		gomcp.WithDescription("Get detailed audio specifications for a project item: sample rate, bit depth, channels, codec, and channel layout."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
	), encH(logger, "get_audio_specs_detailed", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAudioSpecsDetailed(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_is_variable_frame_rate
	s.AddTool(gomcp.NewTool("premiere_is_variable_frame_rate",
		gomcp.WithDescription("Check if a project item has variable frame rate (VFR) which can cause sync issues in editing."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
	), encH(logger, "is_variable_frame_rate", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.IsVariableFrameRate(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// File Operations (21-25)
	// -----------------------------------------------------------------------

	// 21. premiere_get_file_hash
	s.AddTool(gomcp.NewTool("premiere_get_file_hash",
		gomcp.WithDescription("Get a file hash/fingerprint for a project item's media file. Uses file size and modification time as fingerprint."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("algorithm", gomcp.Description("Hash algorithm hint (default: simple). Note: full crypto hashing requires external tools.")),
	), encH(logger, "get_file_hash", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFileHash(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			gomcp.ParseString(req, "algorithm", "simple"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_get_file_dates
	s.AddTool(gomcp.NewTool("premiere_get_file_dates",
		gomcp.WithDescription("Get creation and modification dates for a project item's media file on disk."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
	), encH(logger, "get_file_dates", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFileDates(ctx, gomcp.ParseInt(req, "project_item_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_move_media_file
	s.AddTool(gomcp.NewTool("premiere_move_media_file",
		gomcp.WithDescription("Move a project item's media file to a new directory on disk and automatically relink it in the project."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("new_directory", gomcp.Required(), gomcp.Description("Absolute path to the destination directory")),
	), encH(logger, "move_media_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		newDir := gomcp.ParseString(req, "new_directory", "")
		if newDir == "" {
			return gomcp.NewToolResultError("parameter 'new_directory' is required"), nil
		}
		result, err := orch.MoveMediaFile(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			newDir,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_copy_media_file
	s.AddTool(gomcp.NewTool("premiere_copy_media_file",
		gomcp.WithDescription("Copy a project item's media file to a destination directory. The original file and project link remain unchanged."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("dest_directory", gomcp.Required(), gomcp.Description("Absolute path to the destination directory")),
	), encH(logger, "copy_media_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		destDir := gomcp.ParseString(req, "dest_directory", "")
		if destDir == "" {
			return gomcp.NewToolResultError("parameter 'dest_directory' is required"), nil
		}
		result, err := orch.CopyMediaFile(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			destDir,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_rename_media_file
	s.AddTool(gomcp.NewTool("premiere_rename_media_file",
		gomcp.WithDescription("Rename a project item's media file on disk and automatically relink it in the project."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
		gomcp.WithString("new_name", gomcp.Required(), gomcp.Description("New file name (extension is preserved if not provided)")),
	), encH(logger, "rename_media_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		newName := gomcp.ParseString(req, "new_name", "")
		if newName == "" {
			return gomcp.NewToolResultError("parameter 'new_name' is required"), nil
		}
		result, err := orch.RenameMediaFile(ctx,
			gomcp.ParseInt(req, "project_item_index", 0),
			newName,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Render Queue (26-30)
	// -----------------------------------------------------------------------

	// 26. premiere_add_to_render_queue
	s.AddTool(gomcp.NewTool("premiere_add_to_render_queue",
		gomcp.WithDescription("Add a sequence to Premiere Pro's internal render queue via Adobe Media Encoder."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("preset_path", gomcp.Description("Absolute path to an export preset file (optional)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
	), encH(logger, "add_to_render_queue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.AddToRenderQueue(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "preset_path", ""),
			outputPath,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_get_render_queue_status
	s.AddTool(gomcp.NewTool("premiere_get_render_queue_status",
		gomcp.WithDescription("Get the current status of the render queue including encoder availability and queued jobs."),
	), encH(logger, "get_render_queue_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetRenderQueueStatus(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_clear_render_queue
	s.AddTool(gomcp.NewTool("premiere_clear_render_queue",
		gomcp.WithDescription("Clear all pending jobs from the render queue."),
	), encH(logger, "clear_render_queue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearRenderQueue(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_pause_render_queue
	s.AddTool(gomcp.NewTool("premiere_pause_render_queue",
		gomcp.WithDescription("Pause the currently active render queue processing."),
	), encH(logger, "pause_render_queue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PauseRenderQueue(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_resume_render_queue
	s.AddTool(gomcp.NewTool("premiere_resume_render_queue",
		gomcp.WithDescription("Resume the paused render queue and start processing pending jobs."),
	), encH(logger, "resume_render_queue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ResumeRenderQueue(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
