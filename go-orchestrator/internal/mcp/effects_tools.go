package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerEffectsTools registers all effects, transitions, motion, keyframing,
// blend mode, adjustment layer, and Lumetri color MCP tools.
func registerEffectsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// Transitions (1-8)
	s.AddTool(gomcp.NewTool("premiere_add_video_transition", gomcp.WithDescription("Add a video transition to a clip using QE DOM. Common transitions: Cross Dissolve, Dip to Black, Film Dissolve, Morph Cut."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")), gomcp.WithString("transition_name", gomcp.Description("Transition name (default: Cross Dissolve)")), gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 1.0)")), gomcp.WithBoolean("apply_to_end", gomcp.Description("Apply to clip end (true) or start (false). Default: true"))), efxH(orch, logger, "add_video_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddVideoTransition(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "transition_name", "Cross Dissolve"), gomcp.ParseFloat64(req, "duration", 1.0), gomcp.ParseBoolean(req, "apply_to_end", true))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_add_audio_transition", gomcp.WithDescription("Add an audio transition to a clip. Common: Constant Power, Constant Gain."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")), gomcp.WithString("transition_name", gomcp.Description("Transition name (default: Constant Power)")), gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 1.0)"))), efxH(orch, logger, "add_audio_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddAudioTransition(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "transition_name", "Constant Power"), gomcp.ParseFloat64(req, "duration", 1.0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_remove_transition", gomcp.WithDescription("Remove a transition from a track by index."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("transition_index", gomcp.Required(), gomcp.Description("Zero-based transition index on the track"))), efxH(orch, logger, "remove_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RemoveTransition(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "transition_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_transitions", gomcp.WithDescription("List all transitions on a track."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index"))), efxH(orch, logger, "get_transitions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTransitions(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_default_video_transition", gomcp.WithDescription("Set the default video transition used by Premiere Pro."), gomcp.WithString("transition_name", gomcp.Required(), gomcp.Description("Transition name, e.g. 'Cross Dissolve'"))), efxH(orch, logger, "set_default_video_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetDefaultVideoTransition(ctx, gomcp.ParseString(req, "transition_name", "Cross Dissolve"))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_default_audio_transition", gomcp.WithDescription("Set the default audio transition used by Premiere Pro."), gomcp.WithString("transition_name", gomcp.Required(), gomcp.Description("Transition name, e.g. 'Constant Power'"))), efxH(orch, logger, "set_default_audio_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetDefaultAudioTransition(ctx, gomcp.ParseString(req, "transition_name", "Constant Power"))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_default_transition", gomcp.WithDescription("Apply the default transition to a clip."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track"))), efxH(orch, logger, "apply_default_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyDefaultTransition(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_available_transitions", gomcp.WithDescription("List all available video transition names in this Premiere Pro installation.")), efxH(orch, logger, "get_available_transitions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAvailableTransitions(ctx)
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	// Video Effects (9-16)
	s.AddTool(gomcp.NewTool("premiere_apply_video_effect", gomcp.WithDescription("Apply a video effect by name to a clip using QE DOM. Examples: Gaussian Blur, Lumetri Color, Warp Stabilizer."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")), gomcp.WithString("effect_name", gomcp.Required(), gomcp.Description("Effect name as it appears in the Effects panel"))), efxH(orch, logger, "apply_video_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		en := gomcp.ParseString(req, "effect_name", "")
		if en == "" { return gomcp.NewToolResultError("parameter 'effect_name' is required"), nil }
		result, err := orch.ApplyVideoEffect(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), en)
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_remove_video_effect", gomcp.WithDescription("Remove an applied effect from a video clip by component index. Indices 0 (Motion) and 1 (Opacity) are intrinsic and cannot be removed."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")), gomcp.WithNumber("effect_index", gomcp.Required(), gomcp.Description("Zero-based component index (2+ for applied effects)"))), efxH(orch, logger, "remove_video_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RemoveVideoEffect(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "effect_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_clip_effects", gomcp.WithDescription("List all effects (components) on a clip with their parameter names and values."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track"))), efxH(orch, logger, "get_clip_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipEffects(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_effect_parameter", gomcp.WithDescription("Set a specific effect parameter value by component and parameter index."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index within the component")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("New parameter value"))), efxH(orch, logger, "set_effect_parameter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetEffectParameter(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_effect_parameter", gomcp.WithDescription("Get the current value of a specific effect parameter."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index within the component"))), efxH(orch, logger, "get_effect_parameter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetEffectParameter(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_enable_effect", gomcp.WithDescription("Enable or disable an effect (component) on a clip."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable, false to disable"))), efxH(orch, logger, "enable_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.EnableEffect(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseBoolean(req, "enabled", true))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_copy_effects", gomcp.WithDescription("Copy all effects and their parameter values from a clip to an internal clipboard."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index"))), efxH(orch, logger, "copy_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CopyEffects(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_paste_effects", gomcp.WithDescription("Paste previously copied effects onto a destination clip."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index"))), efxH(orch, logger, "paste_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PasteEffects(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	// Motion & Transform (17-22)
	s.AddTool(gomcp.NewTool("premiere_set_position", gomcp.WithDescription("Set the position (X, Y) of a video clip's Motion effect."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("Horizontal position in pixels")), gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Vertical position in pixels"))), efxH(orch, logger, "set_position", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetPosition(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "x", 960), gomcp.ParseFloat64(req, "y", 540))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_scale", gomcp.WithDescription("Set the scale of a video clip (100 = original size)."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("scale", gomcp.Required(), gomcp.Description("Scale percentage (100 = normal)"))), efxH(orch, logger, "set_scale", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetScale(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "scale", 100))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_rotation", gomcp.WithDescription("Set the rotation of a video clip in degrees."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("degrees", gomcp.Required(), gomcp.Description("Rotation in degrees"))), efxH(orch, logger, "set_rotation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetRotation(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "degrees", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_anchor_point", gomcp.WithDescription("Set the anchor point of a video clip's Motion effect."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("Anchor X position in pixels")), gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Anchor Y position in pixels"))), efxH(orch, logger, "set_anchor_point", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAnchorPoint(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "x", 960), gomcp.ParseFloat64(req, "y", 540))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_opacity", gomcp.WithDescription("Set the opacity of a video clip (0-100)."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("opacity", gomcp.Required(), gomcp.Description("Opacity percentage (0-100)"))), efxH(orch, logger, "set_opacity", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetOpacity(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "opacity", 100))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_motion_properties", gomcp.WithDescription("Get all motion/transform property values (position, scale, rotation, anchor point, opacity) for a video clip."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index"))), efxH(orch, logger, "get_motion_properties", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMotionProperties(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	// Blend Mode (23)
	s.AddTool(gomcp.NewTool("premiere_set_blend_mode", gomcp.WithDescription("Set the blend mode of a video clip."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithString("mode", gomcp.Required(), gomcp.Description("Blend mode name"), gomcp.Enum("Normal", "Darken", "Multiply", "Color Burn", "Linear Burn", "Lighten", "Screen", "Color Dodge", "Linear Dodge (Add)", "Overlay", "Soft Light", "Hard Light", "Vivid Light", "Linear Light", "Pin Light", "Hard Mix", "Difference", "Exclusion", "Hue", "Saturation", "Color", "Luminosity"))), efxH(orch, logger, "set_blend_mode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetBlendMode(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "mode", "Normal"))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	// Adjustment Layer (24-25)
	s.AddTool(gomcp.NewTool("premiere_create_adjustment_layer", gomcp.WithDescription("Create an adjustment layer in the project bin. Apply effects to it and place on a track above your clips."), gomcp.WithString("name", gomcp.Description("Name for the adjustment layer (default: Adjustment Layer)")), gomcp.WithNumber("width", gomcp.Description("Width in pixels (default: 1920)")), gomcp.WithNumber("height", gomcp.Description("Height in pixels (default: 1080)")), gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: 10)"))), efxH(orch, logger, "create_adjustment_layer", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateAdjustmentLayer(ctx, gomcp.ParseString(req, "name", "Adjustment Layer"), gomcp.ParseInt(req, "width", 1920), gomcp.ParseInt(req, "height", 1080), gomcp.ParseFloat64(req, "duration", 10))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_place_adjustment_layer", gomcp.WithDescription("Place an adjustment layer from the project bin onto the timeline."), gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Index of the adjustment layer in the project root bin")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("start_time", gomcp.Description("Start time in seconds (default: 0)")), gomcp.WithNumber("duration", gomcp.Description("Duration in seconds (default: source duration)"))), efxH(orch, logger, "place_adjustment_layer", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.PlaceAdjustmentLayer(ctx, gomcp.ParseInt(req, "project_item_index", 0), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseFloat64(req, "start_time", 0), gomcp.ParseFloat64(req, "duration", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	// Keyframing (26-30)
	s.AddTool(gomcp.NewTool("premiere_add_keyframe", gomcp.WithDescription("Add a keyframe to an effect parameter at a specific time."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index")), gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Keyframe time in seconds (relative to clip start)")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Parameter value at this keyframe"))), efxH(orch, logger, "add_keyframe", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddKeyframe(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0), gomcp.ParseFloat64(req, "time", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_delete_keyframe", gomcp.WithDescription("Delete a keyframe from an effect parameter at a specific time."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index")), gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Keyframe time in seconds to delete"))), efxH(orch, logger, "delete_keyframe", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.DeleteKeyframe(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0), gomcp.ParseFloat64(req, "time", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_keyframe_interpolation", gomcp.WithDescription("Set the interpolation type for a keyframe (linear, bezier, hold)."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index")), gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Keyframe time in seconds")), gomcp.WithString("interpolation_type", gomcp.Required(), gomcp.Description("Interpolation type"), gomcp.Enum("linear", "bezier", "hold", "ease"))), efxH(orch, logger, "set_keyframe_interpolation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetKeyframeInterpolation(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0), gomcp.ParseFloat64(req, "time", 0), gomcp.ParseString(req, "interpolation_type", "linear"))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_keyframes", gomcp.WithDescription("Get all keyframes for an effect parameter, including their times and values."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index"))), efxH(orch, logger, "get_keyframes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetKeyframes(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_time_varying", gomcp.WithDescription("Enable or disable keyframing (time-varying) for an effect parameter."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index")), gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable keyframing, false to disable"))), efxH(orch, logger, "set_time_varying", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetTimeVarying(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "component_index", 0), gomcp.ParseInt(req, "param_index", 0), gomcp.ParseBoolean(req, "enabled", true))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	// Lumetri Color (31-36)
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_brightness", gomcp.WithDescription("Set Lumetri Color brightness. Auto-applies the effect if not present."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Brightness value"))), efxH(orch, logger, "set_lumetri_brightness", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriBrightness(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_contrast", gomcp.WithDescription("Set Lumetri Color contrast. Auto-applies the effect if not present."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Contrast value"))), efxH(orch, logger, "set_lumetri_contrast", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriContrast(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_saturation", gomcp.WithDescription("Set Lumetri Color saturation. Auto-applies the effect if not present."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Saturation value"))), efxH(orch, logger, "set_lumetri_saturation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriSaturation(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_temperature", gomcp.WithDescription("Set Lumetri Color temperature (white balance). Auto-applies the effect if not present."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Temperature value"))), efxH(orch, logger, "set_lumetri_temperature", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriTemperature(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_tint", gomcp.WithDescription("Set Lumetri Color tint. Auto-applies the effect if not present."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Tint value"))), efxH(orch, logger, "set_lumetri_tint", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriTint(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_exposure", gomcp.WithDescription("Set Lumetri Color exposure. Auto-applies the effect if not present."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Exposure value"))), efxH(orch, logger, "set_lumetri_exposure", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriExposure(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
}

// efxH is a small wrapper that logs the tool name before delegating to the handler.
func efxH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
