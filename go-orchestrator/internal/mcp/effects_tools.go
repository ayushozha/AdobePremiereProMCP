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
	s.AddTool(gomcp.NewTool("premiere_add_video_transition", gomcp.WithDescription("Add a video transition effect to a clip's start or end on the timeline. The transition blends between this clip and its neighbor. Requires overlapping handles (extra source media) at the cut point. Common transitions: 'Cross Dissolve' (smooth blend), 'Dip to Black' (fade through black), 'Film Dissolve' (organic film-style blend), 'Morph Cut' (face-aware morphing for jump cuts). Use premiere_get_available_transitions to list all installed transitions."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track). Use premiere_get_video_tracks to list tracks.")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track. Use premiere_get_clips_on_track to find clip indices.")), gomcp.WithString("transition_name", gomcp.Description("Exact name of the transition as it appears in the Effects panel (default: 'Cross Dissolve'). Use premiere_get_available_transitions to list all available names.")), gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 1.0). Typical range: 0.25 to 2.0. The transition is centered on the cut point, so half the duration overlaps each clip.")), gomcp.WithBoolean("apply_to_end", gomcp.Description("If true (default), apply the transition to the clip's tail (outgoing transition to next clip). If false, apply to the clip's head (incoming transition from previous clip)."))), efxH(orch, logger, "add_video_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddVideoTransition(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "transition_name", "Cross Dissolve"), gomcp.ParseFloat64(req, "duration", 1.0), gomcp.ParseBoolean(req, "apply_to_end", true))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_add_audio_transition", gomcp.WithDescription("Add an audio crossfade transition to a clip at its cut point. 'Constant Power' (default) provides a smooth, natural-sounding crossfade (logarithmic curve). 'Constant Gain' uses a linear crossfade (can sound abrupt). Requires overlapping audio handles at the cut point."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track). Use premiere_get_audio_tracks to list tracks.")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track. Use premiere_get_clips_on_track with track_type='audio' to find indices.")), gomcp.WithString("transition_name", gomcp.Description("Audio transition name (default: 'Constant Power'). 'Constant Power' = smooth natural crossfade, 'Constant Gain' = linear crossfade.")), gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 1.0). Typical range: 0.1 to 3.0."))), efxH(orch, logger, "add_audio_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddAudioTransition(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseString(req, "transition_name", "Constant Power"), gomcp.ParseFloat64(req, "duration", 1.0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_remove_transition", gomcp.WithDescription("Remove a transition from a track by its index. Use premiere_get_transitions first to list all transitions and find the correct index. After removal, the clips will have a hard cut at the edit point."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Type of track the transition is on: 'video' or 'audio'."), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index (0 = first/bottom track).")), gomcp.WithNumber("transition_index", gomcp.Required(), gomcp.Description("Zero-based index of the transition on the track. Use premiere_get_transitions to list transitions and their indices."))), efxH(orch, logger, "remove_transition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RemoveTransition(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "transition_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_transitions", gomcp.WithDescription("List all transitions on a specific track, including each transition's name, type, position, duration, and index. Use this to inspect existing transitions before modifying or removing them."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Type of track to inspect: 'video' or 'audio'."), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index (0 = first/bottom track)."))), efxH(orch, logger, "get_transitions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
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
	s.AddTool(gomcp.NewTool("premiere_apply_video_effect", gomcp.WithDescription("Apply a video effect to a clip by its exact name from the Effects panel. The effect is added as a new component on the clip and can be configured with premiere_set_effect_parameter. Common effects: 'Gaussian Blur', 'Lumetri Color', 'Warp Stabilizer', 'Ultra Key' (green screen), 'Crop', 'Transform'. Multiple effects can be stacked on one clip. Use premiere_get_clip_effects to see currently applied effects."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track). Use premiere_get_video_tracks to list tracks.")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track. Use premiere_get_clips_on_track to find indices.")), gomcp.WithString("effect_name", gomcp.Required(), gomcp.Description("Exact name of the effect as it appears in Premiere Pro's Effects panel (e.g. 'Gaussian Blur', 'Lumetri Color', 'Warp Stabilizer'). Case-sensitive."))), efxH(orch, logger, "apply_video_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		en := gomcp.ParseString(req, "effect_name", "")
		if en == "" { return gomcp.NewToolResultError("parameter 'effect_name' is required"), nil }
		result, err := orch.ApplyVideoEffect(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), en)
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_remove_video_effect", gomcp.WithDescription("Remove an applied effect from a video clip by its component index. Component 0 (Motion) and component 1 (Opacity) are intrinsic to every video clip and cannot be removed -- only user-applied effects at index 2 and above can be removed. Use premiere_get_clip_effects to find component indices."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("effect_index", gomcp.Required(), gomcp.Description("Zero-based component index of the effect to remove. Must be >= 2 (indices 0=Motion and 1=Opacity are built-in). Use premiere_get_clip_effects to list components and their indices."))), efxH(orch, logger, "remove_video_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RemoveVideoEffect(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseInt(req, "effect_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_clip_effects", gomcp.WithDescription("List all effects (components) applied to a clip, including each component's name, index, enabled state, and all parameter names/values/indices. For video clips, component 0 is always Motion and component 1 is Opacity (both intrinsic). User-applied effects start at component index 2. Use the returned indices with premiere_set_effect_parameter."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Type of track the clip is on: 'video' or 'audio'."), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index (0 = first/bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track."))), efxH(orch, logger, "get_clip_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetClipEffects(ctx, gomcp.ParseString(req, "track_type", "video"), gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_effect_parameter", gomcp.WithDescription("Set a specific effect parameter value by component index and parameter index. Use premiere_get_clip_effects to discover the component_index and param_index for the parameter you want to change. For example, to set Gaussian Blur amount: find its component_index from get_clip_effects, then set param_index for 'Blurriness'."), gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Type of track the clip is on: 'video' or 'audio'."), gomcp.Enum("video", "audio")), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index (0 = first/bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("component_index", gomcp.Required(), gomcp.Description("Zero-based component (effect) index. For video: 0=Motion, 1=Opacity, 2+=user effects. Use premiere_get_clip_effects to find indices.")), gomcp.WithNumber("param_index", gomcp.Required(), gomcp.Description("Zero-based parameter index within the component. Use premiere_get_clip_effects to see all parameters and their indices.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("New parameter value. The valid range depends on the specific parameter (e.g., opacity: 0-100, blur: 0-1000, position: pixel coordinates)."))), efxH(orch, logger, "set_effect_parameter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
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
	s.AddTool(gomcp.NewTool("premiere_set_position", gomcp.WithDescription("Set the X/Y position of a video clip using the built-in Motion effect. Position is in pixel coordinates relative to the sequence frame. For a 1920x1080 sequence, center is (960, 540), top-left is (0, 0), bottom-right is (1920, 1080). Moving a clip off-screen hides it."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("Horizontal position in pixels. For 1920x1080: 0=left edge, 960=center, 1920=right edge. Values outside the frame range move the clip off-screen.")), gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Vertical position in pixels. For 1920x1080: 0=top edge, 540=center, 1080=bottom edge. Values outside the frame range move the clip off-screen."))), efxH(orch, logger, "set_position", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetPosition(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "x", 960), gomcp.ParseFloat64(req, "y", 540))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_scale", gomcp.WithDescription("Set the uniform scale of a video clip as a percentage using the built-in Motion effect. Scaling affects the visual size of the clip in the frame. Use this for zoom effects, picture-in-picture, or fitting mismatched footage to the sequence resolution."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("scale", gomcp.Required(), gomcp.Description("Scale as a percentage. 100 = original size (no change), 50 = half size, 200 = double size, 25 = quarter-screen (good for picture-in-picture). Must be > 0."))), efxH(orch, logger, "set_scale", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetScale(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "scale", 100))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_rotation", gomcp.WithDescription("Set the rotation angle of a video clip in degrees using the built-in Motion effect. Rotation pivots around the anchor point. Positive values rotate clockwise, negative counterclockwise. Values beyond 360/-360 produce multiple full rotations (useful for spin animations with keyframes)."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("degrees", gomcp.Required(), gomcp.Description("Rotation angle in degrees. 0 = no rotation, 90 = quarter turn clockwise, 180 = upside down, -90 = quarter turn counterclockwise, 360 = full rotation."))), efxH(orch, logger, "set_rotation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetRotation(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "degrees", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_anchor_point", gomcp.WithDescription("Set the anchor point of a video clip's Motion effect."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index")), gomcp.WithNumber("x", gomcp.Required(), gomcp.Description("Anchor X position in pixels")), gomcp.WithNumber("y", gomcp.Required(), gomcp.Description("Anchor Y position in pixels"))), efxH(orch, logger, "set_anchor_point", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAnchorPoint(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "x", 960), gomcp.ParseFloat64(req, "y", 540))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_opacity", gomcp.WithDescription("Set the opacity (transparency) of a video clip using the built-in Opacity effect. Controls how much of the clip below is visible through this clip. Use with keyframes (via premiere_add_keyframe on component 1) to create fade-in/fade-out effects."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("opacity", gomcp.Required(), gomcp.Description("Opacity percentage. 100 = fully opaque (default), 0 = fully transparent (invisible), 50 = half transparent (blend with clip below). Range: 0 to 100."))), efxH(orch, logger, "set_opacity", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
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
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_brightness", gomcp.WithDescription("Set Lumetri Color brightness on a video clip. Auto-applies the Lumetri Color effect if not already present. Brightness adjusts the overall lightness of the image. For more precise control over tonal ranges, use the dedicated Lumetri tools in color_tools.go (exposure, highlights, shadows, etc.)."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Brightness value. 0 = no change, positive = brighter, negative = darker. Typical range: -150 to 150."))), efxH(orch, logger, "set_lumetri_brightness", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriBrightness(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_contrast", gomcp.WithDescription("Set Lumetri Color contrast on a video clip. Auto-applies the Lumetri Color effect if not already present. Contrast increases the difference between dark and light areas. Positive values increase contrast (punchier image), negative values decrease it (flatter image)."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Contrast value. 0 = no change, positive = more contrast, negative = less contrast. Range: -100 to 100."))), efxH(orch, logger, "set_lumetri_contrast", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriContrast(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_saturation", gomcp.WithDescription("Set Lumetri Color saturation on a video clip. Auto-applies the Lumetri Color effect if not already present. Saturation controls the intensity of colors. 100 = normal, 0 = fully desaturated (grayscale/black-and-white), 200 = maximum color intensity."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Saturation value. 100 = normal, 0 = black and white, 200 = maximum saturation. Range: 0 to 200."))), efxH(orch, logger, "set_lumetri_saturation", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriSaturation(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_temperature", gomcp.WithDescription("Set Lumetri Color white balance temperature on a video clip. Auto-applies the Lumetri Color effect if not already present. Temperature shifts colors along the blue-orange axis. Negative = cooler/bluer (corrects warm/orange cast), positive = warmer/more orange (corrects cool/blue cast)."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Temperature value. 0 = no change, negative = cooler/bluer, positive = warmer/more orange. Typical adjustments: -20 to +20."))), efxH(orch, logger, "set_lumetri_temperature", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriTemperature(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_tint", gomcp.WithDescription("Set Lumetri Color white balance tint on a video clip. Auto-applies the Lumetri Color effect if not already present. Tint shifts colors along the green-magenta axis. Negative = greener (corrects magenta cast), positive = more magenta (corrects green cast). Use with temperature for complete white balance correction."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Tint value. 0 = no change, negative = greener, positive = more magenta. Typical adjustments: -20 to +20."))), efxH(orch, logger, "set_lumetri_tint", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetLumetriTint(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil { return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil }
		return toolResultJSON(result)
	}))
	s.AddTool(gomcp.NewTool("premiere_set_lumetri_exposure", gomcp.WithDescription("Set Lumetri Color exposure on a video clip. Auto-applies the Lumetri Color effect if not already present. Exposure adjusts overall brightness in a way that mimics camera exposure stops. Each full stop (+1.0) doubles the brightness. For the more precise color_tools version, use premiere_lumetri_set_exposure."), gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index (0 = bottom track).")), gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track.")), gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Exposure value in stops. 0 = no change, +1.0 = one stop brighter (2x brightness), -1.0 = one stop darker (half brightness). Range: -4.0 to +4.0."))), efxH(orch, logger, "set_lumetri_exposure", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
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
