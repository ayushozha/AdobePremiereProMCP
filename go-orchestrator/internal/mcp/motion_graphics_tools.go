package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// mgH is a small handler wrapper for motion graphics tools.
func mgH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerMotionGraphicsTools registers all 30 motion graphics, essential graphics,
// MOGRT management, text operations, shape layers, countdown/timer, watermark,
// picture layout, animated transition, and subtitling MCP tools.
func registerMotionGraphicsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// =======================================================================
	// Essential Graphics Panel (1-4)
	// =======================================================================

	// 1. premiere_get_essential_graphics_components
	s.AddTool(gomcp.NewTool("premiere_get_essential_graphics_components",
		gomcp.WithDescription("Get all Essential Graphics Panel component properties from a graphics clip on the timeline."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), mgH(orch, logger, "get_essential_graphics_components", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEssentialGraphicsComponents(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_set_essential_graphics_property
	s.AddTool(gomcp.NewTool("premiere_set_essential_graphics_property",
		gomcp.WithDescription("Set a property value on an Essential Graphics Panel clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("property_name", gomcp.Required(), gomcp.Description("Display name of the EGP property to set")),
		gomcp.WithString("value", gomcp.Required(), gomcp.Description("Value to set (as string)")),
	), mgH(orch, logger, "set_essential_graphics_property", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		propName := gomcp.ParseString(req, "property_name", "")
		if propName == "" {
			return gomcp.NewToolResultError("parameter 'property_name' is required"), nil
		}
		value := gomcp.ParseString(req, "value", "")
		if value == "" {
			return gomcp.NewToolResultError("parameter 'value' is required"), nil
		}
		result, err := orch.SetEssentialGraphicsProperty(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			propName, value)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_essential_graphics_text
	s.AddTool(gomcp.NewTool("premiere_get_essential_graphics_text",
		gomcp.WithDescription("Get all text content from a graphics clip on the timeline."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), mgH(orch, logger, "get_essential_graphics_text", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEssentialGraphicsText(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_replace_all_text
	s.AddTool(gomcp.NewTool("premiere_replace_all_text",
		gomcp.WithDescription("Find and replace text in all text properties of a graphics clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("search_text", gomcp.Required(), gomcp.Description("Text to search for")),
		gomcp.WithString("replace_text", gomcp.Required(), gomcp.Description("Replacement text")),
	), mgH(orch, logger, "replace_all_text", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		searchText := gomcp.ParseString(req, "search_text", "")
		if searchText == "" {
			return gomcp.NewToolResultError("parameter 'search_text' is required"), nil
		}
		result, err := orch.ReplaceAllText(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			searchText,
			gomcp.ParseString(req, "replace_text", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// MOGRT Management - extended (5-8)
	// =======================================================================

	// 5. premiere_list_installed_mogrts
	s.AddTool(gomcp.NewTool("premiere_list_installed_mogrts",
		gomcp.WithDescription("List all installed Motion Graphics Templates (.mogrt files) from standard directories."),
	), mgH(orch, logger, "list_installed_mogrts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListInstalledMOGRTs(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_get_mogrt_info
	s.AddTool(gomcp.NewTool("premiere_get_mogrt_info",
		gomcp.WithDescription("Get info about a MOGRT file (name, path, size, dates)."),
		gomcp.WithString("mogrt_path", gomcp.Required(), gomcp.Description("Absolute path to the .mogrt file")),
	), mgH(orch, logger, "get_mogrt_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		mogrtPath := gomcp.ParseString(req, "mogrt_path", "")
		if mogrtPath == "" {
			return gomcp.NewToolResultError("parameter 'mogrt_path' is required"), nil
		}
		result, err := orch.GetMOGRTInfo(ctx, mogrtPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_batch_update_mogrts
	s.AddTool(gomcp.NewTool("premiere_batch_update_mogrts",
		gomcp.WithDescription("Update a property on all MOGRT clips on a specific video track."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithString("property_name", gomcp.Required(), gomcp.Description("Display name of the property to update")),
		gomcp.WithString("value", gomcp.Required(), gomcp.Description("Value to set on all matching clips")),
	), mgH(orch, logger, "batch_update_mogrts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		propName := gomcp.ParseString(req, "property_name", "")
		if propName == "" {
			return gomcp.NewToolResultError("parameter 'property_name' is required"), nil
		}
		result, err := orch.BatchUpdateMOGRTs(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			propName,
			gomcp.ParseString(req, "value", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_create_mogrt_from_clip
	s.AddTool(gomcp.NewTool("premiere_create_mogrt_from_clip",
		gomcp.WithDescription("Export a clip from the timeline as a Motion Graphics Template (.mogrt)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output .mogrt file")),
	), mgH(orch, logger, "create_mogrt_from_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.CreateMOGRTFromClip(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			outputPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Text Operations (9-12)
	// =======================================================================

	// 9. premiere_add_scrolling_title
	s.AddTool(gomcp.NewTool("premiere_add_scrolling_title",
		gomcp.WithDescription("Add a scrolling/crawling title overlay to the timeline."),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Title text content")),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 10)")),
		gomcp.WithNumber("speed", gomcp.Description("Scroll speed in pixels/sec (default: 100)")),
	), mgH(orch, logger, "add_scrolling_title", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.AddScrollingTitle(ctx, text,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 10),
			gomcp.ParseFloat64(req, "speed", 100))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_add_typewriter_text
	s.AddTool(gomcp.NewTool("premiere_add_typewriter_text",
		gomcp.WithDescription("Add text with a typewriter animation effect to the timeline."),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Text content")),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5)")),
		gomcp.WithNumber("type_speed", gomcp.Description("Typing speed in milliseconds per character (default: 50)")),
	), mgH(orch, logger, "add_typewriter_text", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.AddTypewriterText(ctx, text,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 5),
			gomcp.ParseFloat64(req, "type_speed", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_add_text_with_background
	s.AddTool(gomcp.NewTool("premiere_add_text_with_background",
		gomcp.WithDescription("Add text with a colored background box to the timeline."),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Text content")),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5)")),
		gomcp.WithString("bg_color", gomcp.Description("Background color as hex string (default: '#000000')")),
		gomcp.WithNumber("padding", gomcp.Description("Padding around text in pixels (default: 10)")),
	), mgH(orch, logger, "add_text_with_background", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.AddTextWithBackground(ctx, text,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 5),
			gomcp.ParseString(req, "bg_color", "#000000"),
			gomcp.ParseInt(req, "padding", 10))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_set_text_animation
	s.AddTool(gomcp.NewTool("premiere_set_text_animation",
		gomcp.WithDescription("Set a text animation on a clip (fade, slide, scale, typewriter)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("animation_type", gomcp.Description("Animation type: fade, slide, scale, typewriter (default: fade)"),
			gomcp.Enum("fade", "slide", "scale", "typewriter")),
		gomcp.WithNumber("duration", gomcp.Description("Animation duration in seconds (default: 1.0)")),
	), mgH(orch, logger, "set_text_animation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTextAnimation(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			gomcp.ParseString(req, "animation_type", "fade"),
			gomcp.ParseFloat64(req, "duration", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Shape Layers (13-15)
	// =======================================================================

	// 13. premiere_add_rectangle
	s.AddTool(gomcp.NewTool("premiere_add_rectangle",
		gomcp.WithDescription("Add a colored rectangle shape to the timeline."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5)")),
		gomcp.WithNumber("x", gomcp.Description("X position in pixels (default: 0)")),
		gomcp.WithNumber("y", gomcp.Description("Y position in pixels (default: 0)")),
		gomcp.WithNumber("width", gomcp.Description("Width in pixels (default: 200)")),
		gomcp.WithNumber("height", gomcp.Description("Height in pixels (default: 100)")),
		gomcp.WithString("color", gomcp.Description("Fill color as hex string (default: '#FFFFFF')")),
		gomcp.WithNumber("border_width", gomcp.Description("Border width in pixels (default: 0)")),
	), mgH(orch, logger, "add_rectangle", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddRectangle(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 5),
			gomcp.ParseFloat64(req, "x", 0),
			gomcp.ParseFloat64(req, "y", 0),
			gomcp.ParseInt(req, "width", 200),
			gomcp.ParseInt(req, "height", 100),
			gomcp.ParseString(req, "color", "#FFFFFF"),
			gomcp.ParseInt(req, "border_width", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_add_circle
	s.AddTool(gomcp.NewTool("premiere_add_circle",
		gomcp.WithDescription("Add a circle shape to the timeline."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5)")),
		gomcp.WithNumber("x", gomcp.Description("Center X position 0.0-1.0 (default: 0.5)")),
		gomcp.WithNumber("y", gomcp.Description("Center Y position 0.0-1.0 (default: 0.5)")),
		gomcp.WithNumber("radius", gomcp.Description("Radius in pixels (default: 100)")),
		gomcp.WithString("color", gomcp.Description("Fill color as hex string (default: '#FFFFFF')")),
	), mgH(orch, logger, "add_circle", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddCircle(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 5),
			gomcp.ParseFloat64(req, "x", 0.5),
			gomcp.ParseFloat64(req, "y", 0.5),
			gomcp.ParseInt(req, "radius", 100),
			gomcp.ParseString(req, "color", "#FFFFFF"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_add_line
	s.AddTool(gomcp.NewTool("premiere_add_line",
		gomcp.WithDescription("Add a line shape to the timeline."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 5)")),
		gomcp.WithNumber("x1", gomcp.Description("Start X position 0.0-1.0 (default: 0)")),
		gomcp.WithNumber("y1", gomcp.Description("Start Y position 0.0-1.0 (default: 0.5)")),
		gomcp.WithNumber("x2", gomcp.Description("End X position 0.0-1.0 (default: 1.0)")),
		gomcp.WithNumber("y2", gomcp.Description("End Y position 0.0-1.0 (default: 0.5)")),
		gomcp.WithString("color", gomcp.Description("Line color as hex string (default: '#FFFFFF')")),
		gomcp.WithNumber("thickness", gomcp.Description("Line thickness in pixels (default: 2)")),
	), mgH(orch, logger, "add_line", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddLine(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 5),
			gomcp.ParseFloat64(req, "x1", 0),
			gomcp.ParseFloat64(req, "y1", 0.5),
			gomcp.ParseFloat64(req, "x2", 1.0),
			gomcp.ParseFloat64(req, "y2", 0.5),
			gomcp.ParseString(req, "color", "#FFFFFF"),
			gomcp.ParseInt(req, "thickness", 2))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Countdown / Timers (16-17)
	// =======================================================================

	// 16. premiere_add_countdown
	s.AddTool(gomcp.NewTool("premiere_add_countdown",
		gomcp.WithDescription("Add a countdown timer overlay to the timeline."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("from_seconds", gomcp.Description("Countdown start value in seconds (default: 10)")),
		gomcp.WithString("style", gomcp.Description("Visual style: simple, digital, cinematic (default: simple)"),
			gomcp.Enum("simple", "digital", "cinematic")),
	), mgH(orch, logger, "add_countdown", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddCountdown(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseInt(req, "from_seconds", 10),
			gomcp.ParseString(req, "style", "simple"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_add_timecode
	s.AddTool(gomcp.NewTool("premiere_add_timecode",
		gomcp.WithDescription("Add a timecode burn-in overlay to the timeline."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based video track index (default: 0)")),
		gomcp.WithNumber("start_time", gomcp.Description("Start position in seconds (default: 0)")),
		gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 10)")),
		gomcp.WithString("format", gomcp.Description("Timecode format (default: 'HH:MM:SS:FF')"),
			gomcp.Enum("HH:MM:SS:FF", "HH:MM:SS", "MM:SS:FF", "frames")),
	), mgH(orch, logger, "add_timecode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddTimecode(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "start_time", 0),
			gomcp.ParseFloat64(req, "duration", 10),
			gomcp.ParseString(req, "format", "HH:MM:SS:FF"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Watermark (18-20)
	// =======================================================================

	// 18. premiere_add_watermark
	s.AddTool(gomcp.NewTool("premiere_add_watermark",
		gomcp.WithDescription("Add an image watermark overlay to the timeline."),
		gomcp.WithString("image_path", gomcp.Required(), gomcp.Description("Absolute path to the watermark image file")),
		gomcp.WithString("position", gomcp.Description("Position: top-left, top-right, bottom-left, bottom-right, center (default: bottom-right)"),
			gomcp.Enum("top-left", "top-right", "bottom-left", "bottom-right", "center")),
		gomcp.WithNumber("opacity", gomcp.Description("Opacity 0-100 (default: 50)")),
		gomcp.WithNumber("scale", gomcp.Description("Scale percentage (default: 25)")),
	), mgH(orch, logger, "add_watermark", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		imagePath := gomcp.ParseString(req, "image_path", "")
		if imagePath == "" {
			return gomcp.NewToolResultError("parameter 'image_path' is required"), nil
		}
		result, err := orch.AddWatermark(ctx, imagePath,
			gomcp.ParseString(req, "position", "bottom-right"),
			gomcp.ParseFloat64(req, "opacity", 50),
			gomcp.ParseFloat64(req, "scale", 25))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_add_text_watermark
	s.AddTool(gomcp.NewTool("premiere_add_text_watermark",
		gomcp.WithDescription("Add a text watermark overlay to the timeline."),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Watermark text content")),
		gomcp.WithString("position", gomcp.Description("Position: top-left, top-right, bottom-left, bottom-right, center (default: bottom-right)"),
			gomcp.Enum("top-left", "top-right", "bottom-left", "bottom-right", "center")),
		gomcp.WithNumber("opacity", gomcp.Description("Opacity 0-100 (default: 30)")),
		gomcp.WithNumber("font_size", gomcp.Description("Font size in points (default: 24)")),
		gomcp.WithString("color", gomcp.Description("Text color as hex string (default: '#FFFFFF')")),
	), mgH(orch, logger, "add_text_watermark", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.AddTextWatermark(ctx, text,
			gomcp.ParseString(req, "position", "bottom-right"),
			gomcp.ParseFloat64(req, "opacity", 30),
			gomcp.ParseInt(req, "font_size", 24),
			gomcp.ParseString(req, "color", "#FFFFFF"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_remove_watermark
	s.AddTool(gomcp.NewTool("premiere_remove_watermark",
		gomcp.WithDescription("Remove a watermark clip from the timeline."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), mgH(orch, logger, "remove_watermark", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RemoveWatermark(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Picture Layouts (21-22)
	// =======================================================================

	// 21. premiere_create_split_screen
	s.AddTool(gomcp.NewTool("premiere_create_split_screen",
		gomcp.WithDescription("Create a split screen layout (2-up, 3-up, 4-up, or custom)."),
		gomcp.WithString("layout", gomcp.Description("Layout type: 2-up, 3-up, 4-up (default: 2-up)"),
			gomcp.Enum("2-up", "3-up", "4-up")),
		gomcp.WithString("clip_refs", gomcp.Description("JSON array of clip references [{trackIndex, clipIndex}, ...]")),
	), mgH(orch, logger, "create_split_screen", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateSplitScreen(ctx,
			gomcp.ParseString(req, "layout", "2-up"),
			gomcp.ParseString(req, "clip_refs", "[]"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_create_collage
	s.AddTool(gomcp.NewTool("premiere_create_collage",
		gomcp.WithDescription("Create a photo/video collage grid layout."),
		gomcp.WithString("clip_refs", gomcp.Description("JSON array of clip references [{trackIndex, clipIndex}, ...]")),
		gomcp.WithNumber("rows", gomcp.Description("Number of rows (default: 2)")),
		gomcp.WithNumber("cols", gomcp.Description("Number of columns (default: 2)")),
		gomcp.WithNumber("gap", gomcp.Description("Gap between cells in pixels (default: 0)")),
	), mgH(orch, logger, "create_collage", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateCollage(ctx,
			gomcp.ParseString(req, "clip_refs", "[]"),
			gomcp.ParseInt(req, "rows", 2),
			gomcp.ParseInt(req, "cols", 2),
			gomcp.ParseInt(req, "gap", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Animated Transitions - custom (23-25)
	// =======================================================================

	// 23. premiere_add_wipe_transition
	s.AddTool(gomcp.NewTool("premiere_add_wipe_transition",
		gomcp.WithDescription("Add a custom wipe transition with direction and color to a clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("direction", gomcp.Description("Wipe direction: left, right, up, down (default: left)"),
			gomcp.Enum("left", "right", "up", "down")),
		gomcp.WithString("color", gomcp.Description("Wipe color as hex string (default: '#000000')")),
		gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 1.0)")),
	), mgH(orch, logger, "add_wipe_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddWipeTransition(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			gomcp.ParseString(req, "direction", "left"),
			gomcp.ParseString(req, "color", "#000000"),
			gomcp.ParseFloat64(req, "duration", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_add_zoom_transition
	s.AddTool(gomcp.NewTool("premiere_add_zoom_transition",
		gomcp.WithDescription("Add a zoom in/out transition to a clip using scale keyframes."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithBoolean("zoom_in", gomcp.Description("true for zoom in, false for zoom out (default: true)")),
		gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 0.5)")),
	), mgH(orch, logger, "add_zoom_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddZoomTransition(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			gomcp.ParseBoolean(req, "zoom_in", true),
			gomcp.ParseFloat64(req, "duration", 0.5))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_add_glitch_transition
	s.AddTool(gomcp.NewTool("premiere_add_glitch_transition",
		gomcp.WithDescription("Add a glitch effect transition to a clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("intensity", gomcp.Description("Glitch intensity 0-100 (default: 50)")),
		gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 0.5)")),
	), mgH(orch, logger, "add_glitch_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddGlitchTransition(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			gomcp.ParseFloat64(req, "intensity", 50),
			gomcp.ParseFloat64(req, "duration", 0.5))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// =======================================================================
	// Subtitling - extended (26-30)
	// =======================================================================

	// 26. premiere_auto_generate_subtitles
	s.AddTool(gomcp.NewTool("premiere_auto_generate_subtitles",
		gomcp.WithDescription("Auto-generate subtitles from audio using Premiere Pro's Speech to Text feature."),
		gomcp.WithString("language", gomcp.Description("Language code, e.g. 'en', 'es', 'fr' (default: 'en')")),
		gomcp.WithString("style", gomcp.Description("Subtitle style: default, bold, outline (default: 'default')"),
			gomcp.Enum("default", "bold", "outline")),
	), mgH(orch, logger, "auto_generate_subtitles", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AutoGenerateSubtitles(ctx,
			gomcp.ParseString(req, "language", "en"),
			gomcp.ParseString(req, "style", "default"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_translate_subtitles
	s.AddTool(gomcp.NewTool("premiere_translate_subtitles",
		gomcp.WithDescription("Extract existing subtitles for translation to a target language."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
		gomcp.WithString("target_language", gomcp.Required(), gomcp.Description("Target language code, e.g. 'es', 'fr', 'de'")),
	), mgH(orch, logger, "translate_subtitles", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		targetLang := gomcp.ParseString(req, "target_language", "")
		if targetLang == "" {
			return gomcp.NewToolResultError("parameter 'target_language' is required"), nil
		}
		result, err := orch.TranslateSubtitles(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			targetLang)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_format_subtitles
	s.AddTool(gomcp.NewTool("premiere_format_subtitles",
		gomcp.WithDescription("Reformat subtitle line breaks based on character and line limits."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
		gomcp.WithNumber("max_chars_per_line", gomcp.Description("Maximum characters per line (default: 42)")),
		gomcp.WithNumber("max_lines", gomcp.Description("Maximum number of lines per subtitle (default: 2)")),
	), mgH(orch, logger, "format_subtitles", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FormatSubtitles(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "max_chars_per_line", 42),
			gomcp.ParseInt(req, "max_lines", 2))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_burn_in_subtitles
	s.AddTool(gomcp.NewTool("premiere_burn_in_subtitles",
		gomcp.WithDescription("Burn subtitles into the video (returns instructions for export settings)."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
	), mgH(orch, logger, "burn_in_subtitles", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.BurnInSubtitles(ctx, gomcp.ParseInt(req, "track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_adjust_subtitle_timing
	s.AddTool(gomcp.NewTool("premiere_adjust_subtitle_timing",
		gomcp.WithDescription("Shift all subtitle timing by an offset in seconds (positive = later, negative = earlier)."),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based caption track index (default: 0)")),
		gomcp.WithNumber("offset_seconds", gomcp.Required(), gomcp.Description("Time offset in seconds (positive shifts later, negative shifts earlier)")),
	), mgH(orch, logger, "adjust_subtitle_timing", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AdjustSubtitleTiming(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "offset_seconds", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
