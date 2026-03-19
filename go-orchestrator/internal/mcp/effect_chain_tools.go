package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerEffectChainTools registers all effect chain management, effect presets,
// visual effects pipeline, transition effects, effect comparison, and effect
// template MCP tools.
func registerEffectChainTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Effect Chain Management (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_get_effect_chain
	s.AddTool(gomcp.NewTool("premiere_get_effect_chain",
		gomcp.WithDescription("Get the full effect chain on a clip including order, enabled state, and all parameter values."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "get_effect_chain", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEffectChain(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_reorder_effect
	s.AddTool(gomcp.NewTool("premiere_reorder_effect",
		gomcp.WithDescription("Move an effect within the effect chain from one position to another. Cannot move intrinsic effects (Motion/Opacity at indices 0-1)."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("from_index", gomcp.Required(), gomcp.Description("Current effect index (must be >= 2)")),
		gomcp.WithNumber("to_index", gomcp.Required(), gomcp.Description("Target effect index (must be >= 2)")),
	), efxH(orch, logger, "reorder_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ReorderEffect(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "from_index", 0), gomcp.ParseInt(req, "to_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_effect_count
	s.AddTool(gomcp.NewTool("premiere_get_effect_count",
		gomcp.WithDescription("Get the number of applied effects on a clip (excluding intrinsic Motion and Opacity)."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "get_effect_count", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEffectCount(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_clear_all_effects
	s.AddTool(gomcp.NewTool("premiere_clear_all_effects",
		gomcp.WithDescription("Remove all effects from a clip except the intrinsic Motion and Opacity effects."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "clear_all_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearAllEffects(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_duplicate_effect
	s.AddTool(gomcp.NewTool("premiere_duplicate_effect",
		gomcp.WithDescription("Duplicate an existing effect in the effect chain. Cannot duplicate intrinsic effects (indices 0-1)."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("effect_index", gomcp.Required(), gomcp.Description("Zero-based effect index to duplicate (must be >= 2)")),
	), efxH(orch, logger, "duplicate_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.DuplicateEffect(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "effect_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Effect Parameter Animation (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_animate_effect_parameter
	s.AddTool(gomcp.NewTool("premiere_animate_effect_parameter",
		gomcp.WithDescription("Set multiple keyframes on an effect parameter at once. Automatically enables time-varying (keyframing) on the parameter."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
		gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")),
		gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index within the component")),
		gomcp.WithString("keyframes", gomcp.Required(), gomcp.Description("JSON array of keyframes, e.g. [{\"time\":0.5,\"value\":100},{\"time\":2.0,\"value\":0}]")),
	), efxH(orch, logger, "animate_effect_parameter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		kf := gomcp.ParseString(req, "keyframes", "[]")
		result, err := orch.AnimateEffectParameter(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0), kf)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_get_effect_parameter_range
	s.AddTool(gomcp.NewTool("premiere_get_effect_parameter_range",
		gomcp.WithDescription("Get the minimum and maximum allowed values for an effect parameter."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
		gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")),
		gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index within the component")),
	), efxH(orch, logger, "get_effect_parameter_range", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEffectParameterRange(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_reset_effect_parameter
	s.AddTool(gomcp.NewTool("premiere_reset_effect_parameter",
		gomcp.WithDescription("Reset an effect parameter to its default value and clear any keyframes."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
		gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")),
		gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index within the component")),
	), efxH(orch, logger, "reset_effect_parameter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ResetEffectParameter(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_link_effect_parameters
	s.AddTool(gomcp.NewTool("premiere_link_effect_parameters",
		gomcp.WithDescription("Link two effect parameters so they stay in sync. Changes to one parameter will be reflected in the other."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")),
		gomcp.WithNumber("comp1", gomcp.Required(), gomcp.Description("Source component index")),
		gomcp.WithNumber("param1", gomcp.Required(), gomcp.Description("Source parameter index")),
		gomcp.WithNumber("comp2", gomcp.Required(), gomcp.Description("Target component index")),
		gomcp.WithNumber("param2", gomcp.Required(), gomcp.Description("Target parameter index")),
	), efxH(orch, logger, "link_effect_parameters", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.LinkEffectParameters(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "comp1", 0), gomcp.ParseInt(req, "param1", 0), gomcp.ParseInt(req, "comp2", 0), gomcp.ParseInt(req, "param2", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_get_effect_render_order
	s.AddTool(gomcp.NewTool("premiere_get_effect_render_order",
		gomcp.WithDescription("Get the render processing order of all effects on a clip. Effects are rendered top-to-bottom in the chain."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "get_effect_render_order", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEffectRenderOrder(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Common Effect Presets (11-20)
	// -----------------------------------------------------------------------

	// 11. premiere_apply_black_and_white
	s.AddTool(gomcp.NewTool("premiere_apply_black_and_white",
		gomcp.WithDescription("Apply a black and white effect to a video clip by setting saturation to zero."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "apply_black_and_white", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyBlackAndWhite(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_apply_sepia
	s.AddTool(gomcp.NewTool("premiere_apply_sepia",
		gomcp.WithDescription("Apply a sepia tone effect to a video clip with adjustable intensity."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("intensity", gomcp.Description("Sepia intensity from 0-100 (default: 50)")),
	), efxH(orch, logger, "apply_sepia", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplySepia(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "intensity", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_apply_vintage_film
	s.AddTool(gomcp.NewTool("premiere_apply_vintage_film",
		gomcp.WithDescription("Apply a vintage film look to a video clip with reduced saturation, warm tones, faded film, and vignette."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "apply_vintage_film", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyVintageFilm(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_apply_film_grain
	s.AddTool(gomcp.NewTool("premiere_apply_film_grain",
		gomcp.WithDescription("Add a film grain noise effect to a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("Grain amount from 0-100 (default: 30)")),
	), efxH(orch, logger, "apply_film_grain", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyFilmGrain(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", 30))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_apply_vignette
	s.AddTool(gomcp.NewTool("premiere_apply_vignette",
		gomcp.WithDescription("Add a vignette darkening effect to the edges of a video clip via Lumetri Color."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("Vignette amount, negative values darken edges (default: -2.0)")),
		gomcp.WithNumber("feather", gomcp.Description("Vignette feather/softness 0-100 (default: 50)")),
	), efxH(orch, logger, "apply_vignette", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyVignetteEffect(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", -2.0), gomcp.ParseFloat64(req, "feather", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_apply_glow
	s.AddTool(gomcp.NewTool("premiere_apply_glow",
		gomcp.WithDescription("Add a glow/bloom lighting effect to a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("intensity", gomcp.Description("Glow intensity 0-100 (default: 50)")),
		gomcp.WithNumber("radius", gomcp.Description("Glow radius in pixels (default: 20)")),
	), efxH(orch, logger, "apply_glow", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyGlow(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "intensity", 50), gomcp.ParseFloat64(req, "radius", 20))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_apply_drop_shadow
	s.AddTool(gomcp.NewTool("premiere_apply_drop_shadow",
		gomcp.WithDescription("Add a drop shadow effect behind a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("opacity", gomcp.Description("Shadow opacity 0-100 (default: 75)")),
		gomcp.WithNumber("distance", gomcp.Description("Shadow distance in pixels (default: 10)")),
		gomcp.WithNumber("softness", gomcp.Description("Shadow softness/blur (default: 5)")),
		gomcp.WithNumber("direction", gomcp.Description("Shadow direction in degrees 0-360 (default: 315)")),
	), efxH(orch, logger, "apply_drop_shadow", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyDropShadow(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "opacity", 75), gomcp.ParseFloat64(req, "distance", 10), gomcp.ParseFloat64(req, "softness", 5), gomcp.ParseFloat64(req, "direction", 315))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_apply_stroke
	s.AddTool(gomcp.NewTool("premiere_apply_stroke",
		gomcp.WithDescription("Add a stroke/border effect around a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("color", gomcp.Description("Stroke color as hex string (default: '#FFFFFF')")),
		gomcp.WithNumber("width", gomcp.Description("Stroke width in pixels (default: 3)")),
	), efxH(orch, logger, "apply_stroke", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyStroke(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "color", "#FFFFFF"), gomcp.ParseFloat64(req, "width", 3))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_apply_cinematic_bars
	s.AddTool(gomcp.NewTool("premiere_apply_cinematic_bars",
		gomcp.WithDescription("Add cinematic letterbox bars (widescreen crop) to a video clip by applying top and bottom crop."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("bar_height", gomcp.Description("Bar height as percentage of frame (default: 12)")),
	), efxH(orch, logger, "apply_cinematic_bars", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyCinematicBars(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "bar_height", 12))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_apply_flip_horizontal
	s.AddTool(gomcp.NewTool("premiere_apply_flip_horizontal",
		gomcp.WithDescription("Flip a video clip horizontally (mirror effect)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "apply_flip_horizontal", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyFlipHorizontal(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Transition Effects (21-24)
	// -----------------------------------------------------------------------

	// 21. premiere_get_duration_of_transition
	s.AddTool(gomcp.NewTool("premiere_get_duration_of_transition",
		gomcp.WithDescription("Get the duration of a transition on a track."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("transition_index", gomcp.Required(), gomcp.Description("Zero-based transition index on the track")),
	), efxH(orch, logger, "get_duration_of_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDurationOfTransition(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "transition_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_set_transition_duration
	s.AddTool(gomcp.NewTool("premiere_set_transition_duration",
		gomcp.WithDescription("Set the duration of a transition on a track in seconds."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("transition_index", gomcp.Required(), gomcp.Description("Zero-based transition index on the track")),
		gomcp.WithNumber("duration", gomcp.Required(), gomcp.Description("New duration in seconds")),
	), efxH(orch, logger, "set_transition_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTransitionDuration(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "transition_index", 0), gomcp.ParseFloat64(req, "duration", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_set_transition_alignment
	s.AddTool(gomcp.NewTool("premiere_set_transition_alignment",
		gomcp.WithDescription("Set the alignment of a transition relative to the edit point."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("transition_index", gomcp.Required(), gomcp.Description("Zero-based transition index on the track")),
		gomcp.WithString("alignment", gomcp.Required(), gomcp.Description("Alignment type"), gomcp.Enum("center", "start", "end")),
	), efxH(orch, logger, "set_transition_alignment", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTransitionAlignment(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "transition_index", 0), gomcp.ParseString(req, "alignment", "center"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_get_transition_properties
	s.AddTool(gomcp.NewTool("premiere_get_transition_properties",
		gomcp.WithDescription("Get all properties of a transition including name, duration, start/end times, and parameter values."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("transition_index", gomcp.Required(), gomcp.Description("Zero-based transition index on the track")),
	), efxH(orch, logger, "get_transition_properties", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTransitionProperties(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "transition_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Effect Comparison (25-27)
	// -----------------------------------------------------------------------

	// 25. premiere_toggle_effects_preview
	s.AddTool(gomcp.NewTool("premiere_toggle_effects_preview",
		gomcp.WithDescription("Enable or disable all applied effects on a clip for before/after comparison preview."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable effects, false to disable for preview")),
	), efxH(orch, logger, "toggle_effects_preview", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ToggleEffectsPreview(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_get_before_after_snapshot
	s.AddTool(gomcp.NewTool("premiere_get_before_after_snapshot",
		gomcp.WithDescription("Get information about a clip's effects for before/after comparison. Returns clip details and all applied effects."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), efxH(orch, logger, "get_before_after_snapshot", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetBeforeAfterSnapshot(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_compare_effect_settings
	s.AddTool(gomcp.NewTool("premiere_compare_effect_settings",
		gomcp.WithDescription("Compare the effects and parameter values between two clips, showing which effects differ."),
		gomcp.WithString("clip1_ref", gomcp.Required(), gomcp.Description("JSON reference for clip 1: {\"trackType\":\"video\",\"trackIndex\":0,\"clipIndex\":0}")),
		gomcp.WithString("clip2_ref", gomcp.Required(), gomcp.Description("JSON reference for clip 2: {\"trackType\":\"video\",\"trackIndex\":0,\"clipIndex\":1}")),
	), efxH(orch, logger, "compare_effect_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CompareEffectSettings(ctx, gomcp.ParseString(req, "clip1_ref", "{}"), gomcp.ParseString(req, "clip2_ref", "{}"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Effect Templates (28-30)
	// -----------------------------------------------------------------------

	// 28. premiere_save_effect_chain_template
	s.AddTool(gomcp.NewTool("premiere_save_effect_chain_template",
		gomcp.WithDescription("Save the entire effect chain of a clip as a named template for later reuse."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the effect chain template")),
	), efxH(orch, logger, "save_effect_chain_template", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		n := gomcp.ParseString(req, "name", "")
		if n == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SaveEffectChainAsTemplate(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), n)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_load_effect_chain_template
	s.AddTool(gomcp.NewTool("premiere_load_effect_chain_template",
		gomcp.WithDescription("Apply a previously saved effect chain template to a clip."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name of the effect chain template to apply")),
	), efxH(orch, logger, "load_effect_chain_template", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		n := gomcp.ParseString(req, "name", "")
		if n == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.LoadEffectChainTemplate(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), n)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_list_effect_chain_templates
	s.AddTool(gomcp.NewTool("premiere_list_effect_chain_templates",
		gomcp.WithDescription("List all saved effect chain templates with their names and effect counts."),
	), efxH(orch, logger, "list_effect_chain_templates", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListEffectChainTemplates(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
