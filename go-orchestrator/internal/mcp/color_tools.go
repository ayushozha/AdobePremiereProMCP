package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerColorTools registers all Lumetri Color / color correction MCP tools.
func registerColorTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// 1. premiere_lumetri_get_all — Get all Lumetri Color values
	s.AddTool(gomcp.NewTool("premiere_lumetri_get_all",
		gomcp.WithDescription("Get all Lumetri Color parameter values for a clip in a single call. Returns every adjustable parameter across all Lumetri sections: Basic Correction (exposure, contrast, highlights, shadows, whites, blacks, temperature, tint, saturation, vibrance), Creative (faded film, sharpen, LUT), Color Wheels, Curves, and Vignette settings. Returns null/default values if Lumetri is not yet applied. Use this for a complete color snapshot before or after grading."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track). Use premiere_get_video_tracks to list tracks.")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the video track. Use premiere_get_clips_on_track to find indices.")),
	), colorH(orch, logger, "lumetri_get_all", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriGetAll(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_lumetri_set_exposure — Set exposure (-4.0 to 4.0)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_exposure",
		gomcp.WithDescription("Set Lumetri Color exposure. Auto-applies the Lumetri effect if not present. Range: -4.0 to 4.0."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Exposure value (-4.0 to 4.0)")),
	), colorH(orch, logger, "lumetri_set_exposure", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetExposure2(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_lumetri_set_contrast — Set contrast (-100 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_contrast",
		gomcp.WithDescription("Set Lumetri Color contrast. Auto-applies the Lumetri effect if not present. Range: -100 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Contrast value (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_contrast", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetContrast2(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_lumetri_set_highlights — Set highlights (-100 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_highlights",
		gomcp.WithDescription("Set Lumetri Color highlights. Auto-applies the Lumetri effect if not present. Range: -100 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Highlights value (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_highlights", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetHighlights(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_lumetri_set_shadows — Set shadows (-100 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_shadows",
		gomcp.WithDescription("Set Lumetri Color shadows. Auto-applies the Lumetri effect if not present. Range: -100 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Shadows value (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_shadows", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetShadows(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_lumetri_set_whites — Set whites (-100 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_whites",
		gomcp.WithDescription("Set Lumetri Color whites. Auto-applies the Lumetri effect if not present. Range: -100 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Whites value (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_whites", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetWhites(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_lumetri_set_blacks — Set blacks (-100 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_blacks",
		gomcp.WithDescription("Set Lumetri Color blacks. Auto-applies the Lumetri effect if not present. Range: -100 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Blacks value (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_blacks", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetBlacks(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_lumetri_set_temperature — Set white balance temperature
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_temperature",
		gomcp.WithDescription("Set Lumetri Color white balance temperature. Auto-applies the Lumetri effect if not present. Shifts the image along the blue-orange axis. Use to correct color casts from lighting (e.g., fluorescent = too green/blue, tungsten = too orange). Negative = cooler/bluer, positive = warmer/more orange."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the video track.")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Temperature value. 0 = no shift (neutral), negative = cooler/bluer, positive = warmer/more orange. Typical correction range: -30 to +30.")),
	), colorH(orch, logger, "lumetri_set_temperature", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetTemperature2(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_lumetri_set_tint — Set white balance tint
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_tint",
		gomcp.WithDescription("Set Lumetri Color white balance tint. Auto-applies the Lumetri effect if not present. Shifts the image along the green-magenta axis. Use with temperature for complete white balance correction. Negative = greener, positive = more magenta."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the video track.")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Tint value. 0 = no shift (neutral), negative = greener, positive = more magenta. Typical correction range: -30 to +30.")),
	), colorH(orch, logger, "lumetri_set_tint", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetTint2(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_lumetri_set_saturation — Set saturation (0 to 200)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_saturation",
		gomcp.WithDescription("Set Lumetri Color saturation. Auto-applies the Lumetri effect if not present. Range: 0 to 200 (100 = normal)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Saturation value (0-200, 100 = normal)")),
	), colorH(orch, logger, "lumetri_set_saturation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetSaturation2(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 100))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_lumetri_set_vibrance — Set vibrance (-100 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_vibrance",
		gomcp.WithDescription("Set Lumetri Color vibrance. Auto-applies the Lumetri effect if not present. Range: -100 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Vibrance value (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_vibrance", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetVibrance(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_lumetri_set_faded_film — Set faded film amount (0 to 100)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_faded_film",
		gomcp.WithDescription("Set Lumetri Color faded film amount. Auto-applies the Lumetri effect if not present. Range: 0 to 100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Faded film amount (0-100)")),
	), colorH(orch, logger, "lumetri_set_faded_film", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetFadedFilm(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_lumetri_set_sharpen — Set sharpening (0 to 200)
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_sharpen",
		gomcp.WithDescription("Set Lumetri Color sharpening amount. Auto-applies the Lumetri effect if not present. Range: 0 to 200."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Sharpen amount (0-200)")),
	), colorH(orch, logger, "lumetri_set_sharpen", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetSharpen(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_lumetri_set_curve_point — Set point on RGB/Luma curve
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_curve_point",
		gomcp.WithDescription("Set a control point on a Lumetri Color curve (luma, red, green, or blue channel). Input/output values range 0-255."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("channel", gomcp.Required(), gomcp.Description("Curve channel"), gomcp.Enum("luma", "red", "green", "blue")),
		gomcp.WithNumber("input_value", gomcp.Required(), gomcp.Description("Input value on the curve (0-255)")),
		gomcp.WithNumber("output_value", gomcp.Required(), gomcp.Description("Output value on the curve (0-255)")),
	), colorH(orch, logger, "lumetri_set_curve_point", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetCurvePoint(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "channel", "luma"), gomcp.ParseFloat64(req, "input_value", 0), gomcp.ParseFloat64(req, "output_value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_lumetri_set_shadow_color — Set shadow color wheel
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_shadow_color",
		gomcp.WithDescription("Set the shadow color wheel in Lumetri Color (three-way color corrector). Auto-applies the Lumetri effect if not present."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("hue", gomcp.Required(), gomcp.Description("Shadow hue value")),
		gomcp.WithNumber("saturation", gomcp.Required(), gomcp.Description("Shadow saturation value")),
		gomcp.WithNumber("brightness", gomcp.Required(), gomcp.Description("Shadow brightness value")),
	), colorH(orch, logger, "lumetri_set_shadow_color", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetShadowColor(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "hue", 0), gomcp.ParseFloat64(req, "saturation", 0), gomcp.ParseFloat64(req, "brightness", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_lumetri_set_midtone_color — Set midtone color wheel
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_midtone_color",
		gomcp.WithDescription("Set the midtone color wheel in Lumetri Color (three-way color corrector). Auto-applies the Lumetri effect if not present."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("hue", gomcp.Required(), gomcp.Description("Midtone hue value")),
		gomcp.WithNumber("saturation", gomcp.Required(), gomcp.Description("Midtone saturation value")),
		gomcp.WithNumber("brightness", gomcp.Required(), gomcp.Description("Midtone brightness value")),
	), colorH(orch, logger, "lumetri_set_midtone_color", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetMidtoneColor(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "hue", 0), gomcp.ParseFloat64(req, "saturation", 0), gomcp.ParseFloat64(req, "brightness", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_lumetri_set_highlight_color — Set highlight color wheel
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_highlight_color",
		gomcp.WithDescription("Set the highlight color wheel in Lumetri Color (three-way color corrector). Auto-applies the Lumetri effect if not present."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("hue", gomcp.Required(), gomcp.Description("Highlight hue value")),
		gomcp.WithNumber("saturation", gomcp.Required(), gomcp.Description("Highlight saturation value")),
		gomcp.WithNumber("brightness", gomcp.Required(), gomcp.Description("Highlight brightness value")),
	), colorH(orch, logger, "lumetri_set_highlight_color", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetHighlightColor(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "hue", 0), gomcp.ParseFloat64(req, "saturation", 0), gomcp.ParseFloat64(req, "brightness", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_lumetri_set_vignette_amount — Set vignette amount
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_vignette_amount",
		gomcp.WithDescription("Set the Lumetri Color vignette amount. Negative values darken edges, positive values brighten them."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Vignette amount (-5 to 5)")),
	), colorH(orch, logger, "lumetri_set_vignette_amount", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetVignetteAmount(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_lumetri_set_vignette_midpoint — Set vignette midpoint
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_vignette_midpoint",
		gomcp.WithDescription("Set the Lumetri Color vignette midpoint. Controls where the vignette effect starts from the center."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Vignette midpoint (0-100)")),
	), colorH(orch, logger, "lumetri_set_vignette_midpoint", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetVignetteMidpoint(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_lumetri_set_vignette_roundness — Set vignette roundness
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_vignette_roundness",
		gomcp.WithDescription("Set the Lumetri Color vignette roundness. Controls the shape of the vignette."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Vignette roundness (-100 to 100)")),
	), colorH(orch, logger, "lumetri_set_vignette_roundness", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetVignetteRoundness(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_lumetri_set_vignette_feather — Set vignette feather
	s.AddTool(gomcp.NewTool("premiere_lumetri_set_vignette_feather",
		gomcp.WithDescription("Set the Lumetri Color vignette feather. Controls the softness of the vignette edge."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Vignette feather (0-100)")),
	), colorH(orch, logger, "lumetri_set_vignette_feather", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriSetVignetteFeather(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_lumetri_apply_lut — Apply a LUT file (.cube, .3dl)
	s.AddTool(gomcp.NewTool("premiere_lumetri_apply_lut",
		gomcp.WithDescription("Apply a LUT (Look Up Table) file to a clip via Lumetri Color. LUTs are pre-built color transformations used for creative looks (e.g. film emulation) or technical conversions (e.g. LOG to Rec.709). The LUT is applied in Lumetri's Creative section. Auto-applies the Lumetri effect if not present. To remove an applied LUT, use premiere_lumetri_remove_lut."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the video track.")),
		gomcp.WithString("lut_path", gomcp.Required(), gomcp.Description("Absolute path to the LUT file on disk. Supported formats: .cube (most common), .3dl (legacy). Example: '/Users/me/LUTs/FilmLook.cube'.")),
	), colorH(orch, logger, "lumetri_apply_lut", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		lutPath := gomcp.ParseString(req, "lut_path", "")
		if lutPath == "" {
			return gomcp.NewToolResultError("parameter 'lut_path' is required"), nil
		}
		result, err := orch.LumetriApplyLUT(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), lutPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_lumetri_remove_lut — Remove applied LUT
	s.AddTool(gomcp.NewTool("premiere_lumetri_remove_lut",
		gomcp.WithDescription("Remove the currently applied LUT from a clip's Lumetri Color effect."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), colorH(orch, logger, "lumetri_remove_lut", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriRemoveLUT(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_lumetri_auto_color — Auto color correction
	s.AddTool(gomcp.NewTool("premiere_lumetri_auto_color",
		gomcp.WithDescription("Apply automatic color correction to a clip using Lumetri Color. Sets reasonable defaults for exposure, contrast, highlights, shadows, saturation, and vibrance."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), colorH(orch, logger, "lumetri_auto_color", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriAutoColor(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_lumetri_reset — Reset all Lumetri settings to default
	s.AddTool(gomcp.NewTool("premiere_lumetri_reset",
		gomcp.WithDescription("Reset all Lumetri Color settings to their default values on a clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), colorH(orch, logger, "lumetri_reset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriReset(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_get_color_info — Get basic color statistics
	s.AddTool(gomcp.NewTool("premiere_get_color_info",
		gomcp.WithDescription("Get basic color information and current Lumetri Color settings for a clip. Includes clip name, duration, whether Lumetri is applied, and current parameter values."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), colorH(orch, logger, "get_color_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetColorInfo(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_copy_color_grade — Copy Lumetri settings from a clip
	s.AddTool(gomcp.NewTool("premiere_copy_color_grade",
		gomcp.WithDescription("Copy all Lumetri Color settings from a source clip to an internal clipboard. Use premiere_paste_color_grade to apply to other clips."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the source clip")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the source clip")),
	), colorH(orch, logger, "copy_color_grade", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CopyColorGrade(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_paste_color_grade — Paste Lumetri settings to a clip
	s.AddTool(gomcp.NewTool("premiere_paste_color_grade",
		gomcp.WithDescription("Paste previously copied Lumetri Color settings to a destination clip. Must call premiere_copy_color_grade first."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the destination clip")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the destination clip")),
	), colorH(orch, logger, "paste_color_grade", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PasteColorGrade(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_apply_color_grade_to_all — Apply grade to all clips on a track
	s.AddTool(gomcp.NewTool("premiere_apply_color_grade_to_all",
		gomcp.WithDescription("Copy the Lumetri Color grade from a source clip and apply it to all clips on a destination track. Skips the source clip if it is on the same track."),
		gomcp.WithNumber("src_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the source clip")),
		gomcp.WithNumber("src_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the source clip")),
		gomcp.WithNumber("dest_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index to apply the grade to")),
	), colorH(orch, logger, "apply_color_grade_to_all", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyColorGradeToAll(ctx, gomcp.ParseInt(req, "src_track_index", 0), gomcp.ParseInt(req, "src_clip_index", 0), gomcp.ParseInt(req, "dest_track_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_lumetri_auto_white_balance — Auto white balance
	s.AddTool(gomcp.NewTool("premiere_lumetri_auto_white_balance",
		gomcp.WithDescription("Apply automatic white balance correction by resetting temperature and tint to neutral values. Auto-applies the Lumetri effect if not present."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), colorH(orch, logger, "lumetri_auto_white_balance", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LumetriAutoWhiteBalance(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// colorH is a small wrapper that logs the tool name before delegating to the handler.
func colorH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
