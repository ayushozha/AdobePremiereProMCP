package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerGraphicsTools registers all titles, graphics, captions, speed, and
// scene-detection MCP tools.
func registerGraphicsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// =======================================================================
	// MOGRTs
	// =======================================================================

	s.AddTool(
		gomcp.NewTool("premiere_import_mogrt",
			gomcp.WithDescription("Import a Motion Graphics Template (.mogrt) into the active sequence timeline."),
			gomcp.WithString("mogrt_path", gomcp.Required(), gomcp.Description("Absolute path to the .mogrt file")),
			gomcp.WithString("time_ticks", gomcp.Description("Insert position in ticks (254016000000 ticks = 1 second). Default: \"0\"")),
			gomcp.WithNumber("video_track_offset", gomcp.Description("Zero-based video track offset (default: 0)")),
			gomcp.WithNumber("audio_track_offset", gomcp.Description("Zero-based audio track offset (default: 0)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_import_mogrt")
			mogrtPath := gomcp.ParseString(req, "mogrt_path", "")
			if mogrtPath == "" {
				return gomcp.NewToolResultError("parameter 'mogrt_path' is required"), nil
			}
			timeTicks := gomcp.ParseString(req, "time_ticks", "0")
			vOff := gomcp.ParseInt(req, "video_track_offset", 0)
			aOff := gomcp.ParseInt(req, "audio_track_offset", 0)
			result, err := orch.ImportMOGRT(ctx, mogrtPath, timeTicks, vOff, aOff)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to import MOGRT: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_get_mogrt_properties",
			gomcp.WithDescription("Get all editable properties from a MOGRT clip on the timeline."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_get_mogrt_properties")
			result, err := orch.GetMOGRTProperties(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to get MOGRT properties: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_set_mogrt_text",
			gomcp.WithDescription("Set text in a MOGRT clip by property index."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("property_index", gomcp.Required(), gomcp.Description("Zero-based property index from getMOGRTProperties")),
			gomcp.WithString("text", gomcp.Required(), gomcp.Description("The text content to set")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_set_mogrt_text")
			text := gomcp.ParseString(req, "text", "")
			if text == "" {
				return gomcp.NewToolResultError("parameter 'text' is required"), nil
			}
			result, err := orch.SetMOGRTText(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "property_index", 0), text)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to set MOGRT text: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_set_mogrt_property",
			gomcp.WithDescription("Set any property on a MOGRT clip by display name."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithString("property_name", gomcp.Required(), gomcp.Description("Display name of the property to set")),
			gomcp.WithString("value", gomcp.Required(), gomcp.Description("Value to set (as string)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_set_mogrt_property")
			propName := gomcp.ParseString(req, "property_name", "")
			value := gomcp.ParseString(req, "value", "")
			if propName == "" {
				return gomcp.NewToolResultError("parameter 'property_name' is required"), nil
			}
			result, err := orch.SetMOGRTProperty(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), propName, value)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to set MOGRT property: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	// =======================================================================
	// Titles
	// =======================================================================

	s.AddTool(
		gomcp.NewTool("premiere_add_title",
			gomcp.WithDescription("Add a styled title to the timeline using a Motion Graphics Template."),
			gomcp.WithString("text", gomcp.Required(), gomcp.Description("Title text content")),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
			gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
			gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5.0)")),
			gomcp.WithString("style_json", gomcp.Description("JSON string with style options: mogrtPath, font, size, color, position")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_add_title")
			text := gomcp.ParseString(req, "text", "")
			if text == "" {
				return gomcp.NewToolResultError("parameter 'text' is required"), nil
			}
			result, err := orch.AddTitle(ctx, text, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_time", 0), gomcp.ParseFloat64(req, "duration", 5.0), gomcp.ParseString(req, "style_json", ""))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to add title: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_add_lower_third",
			gomcp.WithDescription("Add a lower third overlay with name and title text."),
			gomcp.WithString("name", gomcp.Required(), gomcp.Description("Person's name or primary text")),
			gomcp.WithString("title", gomcp.Required(), gomcp.Description("Job title or secondary text")),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
			gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
			gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5.0)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_add_lower_third")
			name := gomcp.ParseString(req, "name", "")
			title := gomcp.ParseString(req, "title", "")
			if name == "" {
				return gomcp.NewToolResultError("parameter 'name' is required"), nil
			}
			result, err := orch.AddLowerThird(ctx, name, title, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_time", 0), gomcp.ParseFloat64(req, "duration", 5.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to add lower third: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	// =======================================================================
	// Captions & Subtitles
	// =======================================================================

	s.AddTool(
		gomcp.NewTool("premiere_create_caption_track",
			gomcp.WithDescription("Create a new caption track on the active sequence."),
			gomcp.WithString("format", gomcp.Description("Caption format: 'Subtitle' (default), 'Closed'/'608', or '708'"), gomcp.Enum("Subtitle", "Closed", "608", "708")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_create_caption_track")
			result, err := orch.CreateCaptionTrack(ctx, gomcp.ParseString(req, "format", "Subtitle"))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to create caption track: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_import_captions",
			gomcp.WithDescription("Import captions from an SRT or VTT file into the project."),
			gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to the caption file (.srt or .vtt)")),
			gomcp.WithString("format", gomcp.Description("Caption format: 'SRT' (default) or 'VTT'"), gomcp.Enum("SRT", "VTT")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_import_captions")
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			result, err := orch.ImportCaptions(ctx, filePath, gomcp.ParseString(req, "format", "SRT"))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to import captions: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_get_captions",
			gomcp.WithDescription("Get all captions with text and timecodes from a caption track."),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_get_captions")
			result, err := orch.GetCaptions(ctx, gomcp.ParseInt(req, "track_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to get captions: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_add_caption",
			gomcp.WithDescription("Add a single caption to a caption track."),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
			gomcp.WithNumber("start_time", gomcp.Required(), gomcp.Description("Start time in seconds")),
			gomcp.WithNumber("end_time", gomcp.Required(), gomcp.Description("End time in seconds")),
			gomcp.WithString("text", gomcp.Required(), gomcp.Description("Caption text content")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_add_caption")
			text := gomcp.ParseString(req, "text", "")
			if text == "" {
				return gomcp.NewToolResultError("parameter 'text' is required"), nil
			}
			result, err := orch.AddCaption(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_time", 0), gomcp.ParseFloat64(req, "end_time", 3), text)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to add caption: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_edit_caption",
			gomcp.WithDescription("Edit the text of an existing caption."),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
			gomcp.WithNumber("caption_index", gomcp.Required(), gomcp.Description("Zero-based caption index on the track")),
			gomcp.WithString("text", gomcp.Required(), gomcp.Description("New caption text")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_edit_caption")
			text := gomcp.ParseString(req, "text", "")
			if text == "" {
				return gomcp.NewToolResultError("parameter 'text' is required"), nil
			}
			result, err := orch.EditCaption(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "caption_index", 0), text)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to edit caption: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_delete_caption",
			gomcp.WithDescription("Delete a caption from a caption track."),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
			gomcp.WithNumber("caption_index", gomcp.Required(), gomcp.Description("Zero-based caption index to delete")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_delete_caption")
			result, err := orch.DeleteCaption(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "caption_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to delete caption: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_export_captions",
			gomcp.WithDescription("Export captions from the active sequence as SRT or VTT file."),
			gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output caption file")),
			gomcp.WithString("format", gomcp.Description("Export format: 'SRT' (default) or 'VTT'"), gomcp.Enum("SRT", "VTT")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_export_captions")
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportCaptions(ctx, outputPath, gomcp.ParseString(req, "format", "SRT"))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to export captions: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_style_captions",
			gomcp.WithDescription("Style all captions on a track (font, size, color, background, position)."),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
			gomcp.WithString("font", gomcp.Description("Font family name")),
			gomcp.WithNumber("size", gomcp.Description("Font size in points (default: 24)")),
			gomcp.WithString("color", gomcp.Description("Font color as hex string, e.g. '#FFFFFF'")),
			gomcp.WithString("background_color", gomcp.Description("Background color as hex string")),
			gomcp.WithString("position", gomcp.Description("Position: 'bottom', 'top', or 'middle'")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_style_captions")
			result, err := orch.StyleCaptions(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseString(req, "font", ""), gomcp.ParseFloat64(req, "size", 24), gomcp.ParseString(req, "color", ""), gomcp.ParseString(req, "background_color", ""), gomcp.ParseString(req, "position", ""))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to style captions: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	// =======================================================================
	// Graphics
	// =======================================================================

	s.AddTool(
		gomcp.NewTool("premiere_create_color_matte",
			gomcp.WithDescription("Create a color matte project item with a specified RGB color."),
			gomcp.WithString("name", gomcp.Description("Name for the color matte (default: 'Color Matte')")),
			gomcp.WithNumber("red", gomcp.Description("Red component 0-255 (default: 0)")),
			gomcp.WithNumber("green", gomcp.Description("Green component 0-255 (default: 0)")),
			gomcp.WithNumber("blue", gomcp.Description("Blue component 0-255 (default: 0)")),
			gomcp.WithNumber("width", gomcp.Description("Width in pixels (default: 1920)")),
			gomcp.WithNumber("height", gomcp.Description("Height in pixels (default: 1080)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_create_color_matte")
			result, err := orch.CreateColorMatte(ctx, gomcp.ParseString(req, "name", "Color Matte"), gomcp.ParseInt(req, "red", 0), gomcp.ParseInt(req, "green", 0), gomcp.ParseInt(req, "blue", 0), gomcp.ParseInt(req, "width", 1920), gomcp.ParseInt(req, "height", 1080))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to create color matte: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_place_color_matte",
			gomcp.WithDescription("Place a color matte (or any project item) on the timeline."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
			gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
			gomcp.WithNumber("start_time", gomcp.Description("Start time in seconds (default: 0)")),
			gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5.0)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_place_color_matte")
			result, err := orch.PlaceColorMatte(ctx, gomcp.ParseInt(req, "project_item_index", 0), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_time", 0), gomcp.ParseFloat64(req, "duration", 5.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to place color matte: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_create_transparent_video",
			gomcp.WithDescription("Create a transparent video project item."),
			gomcp.WithString("name", gomcp.Description("Name for the transparent video (default: 'Transparent Video')")),
			gomcp.WithNumber("width", gomcp.Description("Width in pixels (default: 1920)")),
			gomcp.WithNumber("height", gomcp.Description("Height in pixels (default: 1080)")),
			gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 10.0)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_create_transparent_video")
			result, err := orch.CreateTransparentVideo(ctx, gomcp.ParseString(req, "name", "Transparent Video"), gomcp.ParseInt(req, "width", 1920), gomcp.ParseInt(req, "height", 1080), gomcp.ParseFloat64(req, "duration", 10.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to create transparent video: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	// =======================================================================
	// Speed & Time
	// =======================================================================

	s.AddTool(
		gomcp.NewTool("premiere_set_time_remapping",
			gomcp.WithDescription("Enable or disable time remapping on a video clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable, false to disable time remapping")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_set_time_remapping")
			result, err := orch.SetTimeRemapping(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseBoolean(req, "enabled", true))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to set time remapping: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_add_time_remap_keyframe",
			gomcp.WithDescription("Add a speed keyframe to a clip's time remapping."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time in seconds for the keyframe")),
			gomcp.WithNumber("speed", gomcp.Required(), gomcp.Description("Speed multiplier at this keyframe (1.0 = normal)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_add_time_remap_keyframe")
			result, err := orch.AddTimeRemapKeyframe(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "time", 0), gomcp.ParseFloat64(req, "speed", 1.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to add time remap keyframe: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	s.AddTool(
		gomcp.NewTool("premiere_freeze_frame",
			gomcp.WithDescription("Create a freeze frame at a specified time for a given duration."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time in seconds where the frame should freeze")),
			gomcp.WithNumber("duration", gomcp.Description("Duration of the freeze in seconds (default: 2.0)")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_freeze_frame")
			result, err := orch.FreezeFrame(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "time", 0), gomcp.ParseFloat64(req, "duration", 2.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to create freeze frame: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)

	// =======================================================================
	// Scene Edit Detection
	// =======================================================================

	s.AddTool(
		gomcp.NewTool("premiere_detect_scene_edits",
			gomcp.WithDescription("Run scene edit detection on a clip to find cut points. Creates markers at detected scene changes."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("sensitivity", gomcp.Description("Detection sensitivity 0-100 (default: 50). Higher = more cuts detected.")),
		),
		func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			logger.Debug("handling premiere_detect_scene_edits")
			result, err := orch.DetectSceneEdits(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "sensitivity", 50.0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to detect scene edits: %v", err)), nil
			}
			return toolResultJSON(result)
		},
	)
}
