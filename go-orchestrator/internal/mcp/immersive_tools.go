package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// makeImmersiveHandler creates a handler that extracts named parameters from
// the MCP request, packages them into a map, and delegates to
// EvalImmersiveCommand on the orchestrator.
func makeImmersiveHandler(orch Orchestrator, logger *zap.Logger, command string, paramNames []string) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling immersive tool", zap.String("command", command))

		args := make(map[string]any, len(paramNames))
		for _, name := range paramNames {
			raw := gomcp.ParseArgument(req, name, nil)
			if raw != nil {
				args[name] = raw
			}
		}

		result, err := orch.EvalImmersiveCommand(ctx, command, args)
		if err != nil {
			logger.Error("immersive command failed",
				zap.String("command", command),
				zap.Error(err),
			)
			return gomcp.NewToolResultError(fmt.Sprintf("failed to execute %s: %v", command, err)), nil
		}
		return toolResultJSON(result)
	}
}

// registerImmersiveTools registers all 30 VR/360, HDR, stereoscopic, frame
// rate, aspect ratio, timecode, render settings, and extended captions MCP
// tools. These correspond to ExtendScript functions in the immersive section
// of premiere.jsx.
func registerImmersiveTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// VR / 360 Video (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_set_vr_projection
	s.AddTool(
		gomcp.NewTool("premiere_set_vr_projection",
			gomcp.WithDescription("Set the VR projection type for a sequence (equirectangular, cubemap)."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithString("projection", gomcp.Required(), gomcp.Description("VR projection type"),
				gomcp.Enum("equirectangular", "cubemap")),
		),
		makeImmersiveHandler(orch, logger, "setVRProjection", []string{"sequence_index", "projection"}),
	)

	// 2. premiere_get_vr_projection
	s.AddTool(
		gomcp.NewTool("premiere_get_vr_projection",
			gomcp.WithDescription("Get the current VR projection type for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
		),
		makeImmersiveHandler(orch, logger, "getVRProjection", []string{"sequence_index"}),
	)

	// 3. premiere_set_vr_field_of_view
	s.AddTool(
		gomcp.NewTool("premiere_set_vr_field_of_view",
			gomcp.WithDescription("Set the field of view for a VR clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("fov", gomcp.Required(), gomcp.Description("Field of view in degrees (1-360)")),
		),
		makeImmersiveHandler(orch, logger, "setVRFieldOfView", []string{"track_index", "clip_index", "fov"}),
	)

	// 4. premiere_rotate_vr_view
	s.AddTool(
		gomcp.NewTool("premiere_rotate_vr_view",
			gomcp.WithDescription("Rotate the VR viewpoint for a clip (pan, tilt, roll)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("pan", gomcp.Required(), gomcp.Description("Pan angle in degrees (-180 to 180)")),
			gomcp.WithNumber("tilt", gomcp.Required(), gomcp.Description("Tilt angle in degrees (-90 to 90)")),
			gomcp.WithNumber("roll", gomcp.Required(), gomcp.Description("Roll angle in degrees (-180 to 180)")),
		),
		makeImmersiveHandler(orch, logger, "rotateVRView", []string{"track_index", "clip_index", "pan", "tilt", "roll"}),
	)

	// 5. premiere_create_vr_sequence
	s.AddTool(
		gomcp.NewTool("premiere_create_vr_sequence",
			gomcp.WithDescription("Create a VR-specific sequence with projection settings."),
			gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the new VR sequence")),
			gomcp.WithNumber("width", gomcp.Description("Frame width in pixels (default: 3840)")),
			gomcp.WithNumber("height", gomcp.Description("Frame height in pixels (default: 1920)")),
			gomcp.WithNumber("fps", gomcp.Description("Frame rate (default: 30)")),
			gomcp.WithString("projection", gomcp.Description("VR projection type (default: equirectangular)"),
				gomcp.Enum("equirectangular", "cubemap")),
		),
		makeImmersiveHandler(orch, logger, "createVRSequence", []string{"name", "width", "height", "fps", "projection"}),
	)

	// -----------------------------------------------------------------------
	// HDR (6-8)
	// -----------------------------------------------------------------------

	// 6. premiere_set_hdr_settings
	s.AddTool(
		gomcp.NewTool("premiere_set_hdr_settings",
			gomcp.WithDescription("Configure HDR settings for a sequence (color space, max luminance)."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithString("color_space", gomcp.Required(), gomcp.Description("HDR color space"),
				gomcp.Enum("rec709", "rec2020", "rec2100hlg", "rec2100pq")),
			gomcp.WithNumber("max_luminance", gomcp.Description("Maximum luminance in nits (default: 1000)")),
		),
		makeImmersiveHandler(orch, logger, "setHDRSettings", []string{"sequence_index", "color_space", "max_luminance"}),
	)

	// 7. premiere_get_hdr_settings
	s.AddTool(
		gomcp.NewTool("premiere_get_hdr_settings",
			gomcp.WithDescription("Get the HDR configuration for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
		),
		makeImmersiveHandler(orch, logger, "getHDRSettings", []string{"sequence_index"}),
	)

	// 8. premiere_is_hdr_sequence
	s.AddTool(
		gomcp.NewTool("premiere_is_hdr_sequence",
			gomcp.WithDescription("Check if a sequence is configured for HDR."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
		),
		makeImmersiveHandler(orch, logger, "isHDRSequence", []string{"sequence_index"}),
	)

	// -----------------------------------------------------------------------
	// Stereoscopic 3D (9-10)
	// -----------------------------------------------------------------------

	// 9. premiere_set_stereoscopic_mode
	s.AddTool(
		gomcp.NewTool("premiere_set_stereoscopic_mode",
			gomcp.WithDescription("Set the stereoscopic 3D mode for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithString("mode", gomcp.Required(), gomcp.Description("Stereoscopic 3D mode"),
				gomcp.Enum("side-by-side", "over-under", "anaglyph")),
		),
		makeImmersiveHandler(orch, logger, "setStereoscopicMode", []string{"sequence_index", "mode"}),
	)

	// 10. premiere_get_stereoscopic_mode
	s.AddTool(
		gomcp.NewTool("premiere_get_stereoscopic_mode",
			gomcp.WithDescription("Get the stereoscopic 3D mode for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
		),
		makeImmersiveHandler(orch, logger, "getStereoscopicMode", []string{"sequence_index"}),
	)

	// -----------------------------------------------------------------------
	// Frame Rate (11-13)
	// -----------------------------------------------------------------------

	// 11. premiere_set_sequence_frame_rate
	s.AddTool(
		gomcp.NewTool("premiere_set_sequence_frame_rate",
			gomcp.WithDescription("Change the frame rate of a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithNumber("fps", gomcp.Required(), gomcp.Description("Target frame rate (e.g. 23.976, 24, 25, 29.97, 30, 50, 59.94, 60)")),
		),
		makeImmersiveHandler(orch, logger, "setSequenceFrameRate", []string{"sequence_index", "fps"}),
	)

	// 12. premiere_interpret_footage_frame_rate
	s.AddTool(
		gomcp.NewTool("premiere_interpret_footage_frame_rate",
			gomcp.WithDescription("Interpret footage at a different frame rate without re-encoding."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based project item index")),
			gomcp.WithNumber("fps", gomcp.Required(), gomcp.Description("Frame rate to interpret the footage at")),
		),
		makeImmersiveHandler(orch, logger, "interpretFootageFrameRate", []string{"project_item_index", "fps"}),
	)

	// 13. premiere_get_available_frame_rates
	s.AddTool(
		gomcp.NewTool("premiere_get_available_frame_rates",
			gomcp.WithDescription("List all standard frame rates supported by Premiere Pro."),
		),
		makeImmersiveHandler(orch, logger, "getAvailableFrameRates", nil),
	)

	// -----------------------------------------------------------------------
	// Aspect Ratio (14-17)
	// -----------------------------------------------------------------------

	// 14. premiere_set_pixel_aspect_ratio
	s.AddTool(
		gomcp.NewTool("premiere_set_pixel_aspect_ratio",
			gomcp.WithDescription("Set the pixel aspect ratio for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithNumber("num", gomcp.Required(), gomcp.Description("Numerator of the pixel aspect ratio")),
			gomcp.WithNumber("den", gomcp.Required(), gomcp.Description("Denominator of the pixel aspect ratio")),
		),
		makeImmersiveHandler(orch, logger, "setPixelAspectRatio", []string{"sequence_index", "num", "den"}),
	)

	// 15. premiere_get_pixel_aspect_ratio
	s.AddTool(
		gomcp.NewTool("premiere_get_pixel_aspect_ratio",
			gomcp.WithDescription("Get the pixel aspect ratio for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
		),
		makeImmersiveHandler(orch, logger, "getPixelAspectRatio", []string{"sequence_index"}),
	)

	// 16. premiere_add_letterboxing
	s.AddTool(
		gomcp.NewTool("premiere_add_letterboxing",
			gomcp.WithDescription("Add letterbox bars (horizontal bars) to a clip for a target aspect ratio."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("target_aspect", gomcp.Required(), gomcp.Description("Target aspect ratio as a decimal (e.g. 2.35 for 2.35:1 cinemascope)")),
		),
		makeImmersiveHandler(orch, logger, "addLetterboxing", []string{"track_index", "clip_index", "target_aspect"}),
	)

	// 17. premiere_add_pillarboxing
	s.AddTool(
		gomcp.NewTool("premiere_add_pillarboxing",
			gomcp.WithDescription("Add pillarbox bars (vertical bars) to a clip for a target aspect ratio."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("target_aspect", gomcp.Required(), gomcp.Description("Target aspect ratio as a decimal (e.g. 1.33 for 4:3)")),
		),
		makeImmersiveHandler(orch, logger, "addPillarboxing", []string{"track_index", "clip_index", "target_aspect"}),
	)

	// -----------------------------------------------------------------------
	// Timecode (18-21)
	// -----------------------------------------------------------------------

	// 18. premiere_set_timecode_offset
	s.AddTool(
		gomcp.NewTool("premiere_set_timecode_offset",
			gomcp.WithDescription("Set the starting timecode offset for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithString("offset", gomcp.Required(), gomcp.Description("Timecode offset string (e.g. '01:00:00:00')")),
		),
		makeImmersiveHandler(orch, logger, "setTimecodeOffset", []string{"sequence_index", "offset"}),
	)

	// 19. premiere_get_timecode_offset
	s.AddTool(
		gomcp.NewTool("premiere_get_timecode_offset",
			gomcp.WithDescription("Get the timecode offset for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
		),
		makeImmersiveHandler(orch, logger, "getTimecodeOffset", []string{"sequence_index"}),
	)

	// 20. premiere_set_drop_frame
	s.AddTool(
		gomcp.NewTool("premiere_set_drop_frame",
			gomcp.WithDescription("Toggle drop frame timecode for a sequence."),
			gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based sequence index")),
			gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true for drop frame, false for non-drop frame")),
		),
		makeImmersiveHandler(orch, logger, "setDropFrame", []string{"sequence_index", "enabled"}),
	)

	// 21. premiere_convert_timecode
	s.AddTool(
		gomcp.NewTool("premiere_convert_timecode",
			gomcp.WithDescription("Convert a timecode between frame rates."),
			gomcp.WithString("timecode", gomcp.Required(), gomcp.Description("Source timecode string (e.g. '01:00:00:00')")),
			gomcp.WithNumber("from_fps", gomcp.Required(), gomcp.Description("Source frame rate")),
			gomcp.WithNumber("to_fps", gomcp.Required(), gomcp.Description("Target frame rate")),
		),
		makeImmersiveHandler(orch, logger, "convertTimecode", []string{"timecode", "from_fps", "to_fps"}),
	)

	// -----------------------------------------------------------------------
	// Render Settings (22-25)
	// -----------------------------------------------------------------------

	// 22. premiere_get_max_render_quality
	s.AddTool(
		gomcp.NewTool("premiere_get_max_render_quality",
			gomcp.WithDescription("Check the current maximum render quality setting."),
		),
		makeImmersiveHandler(orch, logger, "getMaxRenderQuality", nil),
	)

	// 23. premiere_set_max_render_quality
	s.AddTool(
		gomcp.NewTool("premiere_set_max_render_quality",
			gomcp.WithDescription("Toggle maximum render quality on or off."),
			gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable max render quality, false to disable")),
		),
		makeImmersiveHandler(orch, logger, "setMaxRenderQuality", []string{"enabled"}),
	)

	// 24. premiere_set_max_bit_depth
	s.AddTool(
		gomcp.NewTool("premiere_set_max_bit_depth",
			gomcp.WithDescription("Toggle maximum bit depth rendering on or off."),
			gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable max bit depth, false to disable")),
		),
		makeImmersiveHandler(orch, logger, "setMaxBitDepth", []string{"enabled"}),
	)

	// 25. premiere_get_gpu_render_status
	s.AddTool(
		gomcp.NewTool("premiere_get_gpu_render_status",
			gomcp.WithDescription("Get the current GPU rendering status and configuration."),
		),
		makeImmersiveHandler(orch, logger, "getGPURenderStatus", nil),
	)

	// -----------------------------------------------------------------------
	// Closed Captions Extended (26-30)
	// -----------------------------------------------------------------------

	// 26. premiere_get_caption_formats
	s.AddTool(
		gomcp.NewTool("premiere_get_caption_formats",
			gomcp.WithDescription("List all supported caption formats in this Premiere Pro installation."),
		),
		makeImmersiveHandler(orch, logger, "getCaptionFormats", nil),
	)

	// 27. premiere_set_caption_position
	s.AddTool(
		gomcp.NewTool("premiere_set_caption_position",
			gomcp.WithDescription("Position an individual caption on screen."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based caption track index")),
			gomcp.WithNumber("caption_index", gomcp.Required(), gomcp.Description("Zero-based caption index on the track")),
			gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("Horizontal position 0.0-1.0 (0=left, 1=right)")),
			gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Vertical position 0.0-1.0 (0=top, 1=bottom)")),
		),
		makeImmersiveHandler(orch, logger, "setCaptionPosition", []string{"track_index", "caption_index", "x", "y"}),
	)

	// 28. premiere_set_caption_background
	s.AddTool(
		gomcp.NewTool("premiere_set_caption_background",
			gomcp.WithDescription("Set the background color and opacity of a caption."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based caption track index")),
			gomcp.WithNumber("caption_index", gomcp.Required(), gomcp.Description("Zero-based caption index on the track")),
			gomcp.WithString("color", gomcp.Required(), gomcp.Description("Background color as hex string (e.g. '#000000')")),
			gomcp.WithNumber("opacity", gomcp.Description("Background opacity 0-100 (default: 80)")),
		),
		makeImmersiveHandler(orch, logger, "setCaptionBackground", []string{"track_index", "caption_index", "color", "opacity"}),
	)

	// 29. premiere_align_caption_to_speech
	s.AddTool(
		gomcp.NewTool("premiere_align_caption_to_speech",
			gomcp.WithDescription("Auto-align captions to detected speech on a track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based caption track index")),
		),
		makeImmersiveHandler(orch, logger, "alignCaptionToSpeech", []string{"track_index"}),
	)

	// 30. premiere_split_long_captions
	s.AddTool(
		gomcp.NewTool("premiere_split_long_captions",
			gomcp.WithDescription("Split captions longer than a maximum character count into shorter segments."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based caption track index")),
			gomcp.WithNumber("max_chars", gomcp.Description("Maximum characters per caption (default: 42)")),
		),
		makeImmersiveHandler(orch, logger, "splitLongCaptions", []string{"track_index", "max_chars"}),
	)
}
