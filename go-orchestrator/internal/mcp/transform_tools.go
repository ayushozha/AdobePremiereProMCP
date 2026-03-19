package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerTransformTools registers all transform, crop, masking, stabilization,
// blur, sharpen, distortion, noise reduction, and PIP MCP tools.
func registerTransformTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// --- Crop Effect (1-3) ---

	s.AddTool(gomcp.NewTool("premiere_set_crop",
		gomcp.WithDescription("Set crop values on a video clip. Each value is a percentage (0-100) of the frame to crop from that edge."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("left", gomcp.Description("Left crop percentage 0-100 (default: 0)")),
		gomcp.WithNumber("right", gomcp.Description("Right crop percentage 0-100 (default: 0)")),
		gomcp.WithNumber("top", gomcp.Description("Top crop percentage 0-100 (default: 0)")),
		gomcp.WithNumber("bottom", gomcp.Description("Bottom crop percentage 0-100 (default: 0)")),
	), txH(orch, logger, "set_crop", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetCrop(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "left", 0), gomcp.ParseFloat64(req, "right", 0), gomcp.ParseFloat64(req, "top", 0), gomcp.ParseFloat64(req, "bottom", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_crop",
		gomcp.WithDescription("Get the current crop values (left, right, top, bottom) for a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "get_crop", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetCrop(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_reset_crop",
		gomcp.WithDescription("Reset all crop values to 0 (no crop) for a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "reset_crop", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ResetCrop(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Transform Extended (4-10) ---

	s.AddTool(gomcp.NewTool("premiere_set_uniform_scale",
		gomcp.WithDescription("Toggle uniform scale on a video clip's Motion effect. When enabled, Scale Width and Scale Height are locked together."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable uniform scale, false to allow independent width/height scaling")),
	), txH(orch, logger, "set_uniform_scale", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetUniformScale(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_transform_properties",
		gomcp.WithDescription("Get all transform property values for a video clip including position, scale, rotation, anchor point, anti-flicker filter, and uniform scale state."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "get_transform_properties", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTransformProperties(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_set_anti_flicker",
		gomcp.WithDescription("Set the anti-flicker filter value on a video clip's Motion effect. Helps reduce flicker on interlaced or thin-line content."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("value", gomcp.Required(), gomcp.Description("Anti-flicker filter value (0.0 to 1.0)")),
	), txH(orch, logger, "set_anti_flicker", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAntiFlicker(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "value", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_reset_transform",
		gomcp.WithDescription("Reset all transform properties (position, scale, rotation, anchor point) to their default values for a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "reset_transform", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ResetTransform(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_center_clip",
		gomcp.WithDescription("Center a video clip in the frame by setting position to the frame center (sequence width/2, height/2)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "center_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CenterClip(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_fit_clip_to_frame",
		gomcp.WithDescription("Scale a video clip to fit exactly within the sequence frame, preserving aspect ratio. The entire clip will be visible, but letterboxing/pillarboxing may occur."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "fit_clip_to_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FitClipToFrame(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_fill_frame",
		gomcp.WithDescription("Scale a video clip to fill the entire sequence frame, preserving aspect ratio. Parts of the clip may be cropped if aspect ratios differ."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "fill_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FillFrame(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Picture-in-Picture (11-12) ---

	s.AddTool(gomcp.NewTool("premiere_create_pip",
		gomcp.WithDescription("Set up a picture-in-picture layout. Scales and positions the PIP clip at a corner or center of the frame over the main clip."),
		gomcp.WithNumber("main_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the main (background) clip")),
		gomcp.WithNumber("main_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the main clip")),
		gomcp.WithNumber("pip_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the PIP (overlay) clip")),
		gomcp.WithNumber("pip_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the PIP clip")),
		gomcp.WithString("position", gomcp.Description("PIP position: top-left, top-right, bottom-left, bottom-right, center (default: bottom-right)"), gomcp.Enum("top-left", "top-right", "bottom-left", "bottom-right", "center")),
		gomcp.WithNumber("scale", gomcp.Description("PIP scale percentage (default: 30)")),
	), txH(orch, logger, "create_pip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreatePIP(ctx, gomcp.ParseInt(req, "main_track_index", 0), gomcp.ParseInt(req, "main_clip_index", 0), gomcp.ParseInt(req, "pip_track_index", 0), gomcp.ParseInt(req, "pip_clip_index", 0), gomcp.ParseString(req, "position", "bottom-right"), gomcp.ParseFloat64(req, "scale", 30))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_remove_pip",
		gomcp.WithDescription("Remove picture-in-picture by resetting the clip to full frame (scale 100%, centered position)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the PIP clip")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the PIP clip")),
	), txH(orch, logger, "remove_pip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RemovePIP(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Opacity & Masking (13-16) ---

	s.AddTool(gomcp.NewTool("premiere_set_opacity_keyframes",
		gomcp.WithDescription("Set multiple opacity keyframes on a video clip. Provide an array of {time, value} objects where time is in seconds relative to clip start and value is 0-100."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("keyframes", gomcp.Required(), gomcp.Description("JSON array of keyframes, e.g. '[{\"time\":0,\"value\":0},{\"time\":1,\"value\":100}]'")),
	), txH(orch, logger, "set_opacity_keyframes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		kf := gomcp.ParseString(req, "keyframes", "[]")
		result, err := orch.SetOpacityKeyframes(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), kf)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_fade_in",
		gomcp.WithDescription("Add a fade-in effect to a video clip by animating opacity from 0 to 100 over the specified duration at the clip start."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("duration_seconds", gomcp.Description("Fade duration in seconds (default: 1.0)")),
	), txH(orch, logger, "fade_in", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FadeIn(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "duration_seconds", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_fade_out",
		gomcp.WithDescription("Add a fade-out effect to a video clip by animating opacity from 100 to 0 over the specified duration at the clip end."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("duration_seconds", gomcp.Description("Fade duration in seconds (default: 1.0)")),
	), txH(orch, logger, "fade_out", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.FadeOut(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "duration_seconds", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_cross_fade_clips",
		gomcp.WithDescription("Cross-fade between two adjacent clips on the same track using opacity keyframes. Clip A fades out while clip B fades in over the overlap duration."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index_a", gomcp.Required(), gomcp.Description("Zero-based clip index of the first (outgoing) clip")),
		gomcp.WithNumber("clip_index_b", gomcp.Required(), gomcp.Description("Zero-based clip index of the second (incoming) clip")),
		gomcp.WithNumber("duration_seconds", gomcp.Description("Cross-fade duration in seconds (default: 1.0)")),
	), txH(orch, logger, "cross_fade_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CrossFadeClips(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index_a", 0), gomcp.ParseInt(req, "clip_index_b", 0), gomcp.ParseFloat64(req, "duration_seconds", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Stabilization (17-18) ---

	s.AddTool(gomcp.NewTool("premiere_apply_warp_stabilizer",
		gomcp.WithDescription("Apply the Warp Stabilizer effect to a video clip. Smoothness controls stabilization intensity (higher = smoother but more cropping). The effect requires background analysis after application."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("smoothness", gomcp.Description("Smoothness percentage (default: 50). Range: 0-100")),
	), txH(orch, logger, "apply_warp_stabilizer", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyWarpStabilizer(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "smoothness", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_get_stabilization_status",
		gomcp.WithDescription("Check the analysis status of the Warp Stabilizer effect on a clip. Returns whether stabilization analysis is pending, in progress, or complete."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
	), txH(orch, logger, "get_stabilization_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetStabilizationStatus(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Lens Distortion (19) ---

	s.AddTool(gomcp.NewTool("premiere_apply_lens_distortion_removal",
		gomcp.WithDescription("Apply lens distortion removal effect to correct barrel or pincushion distortion from wide-angle or action camera lenses."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("curvature", gomcp.Description("Curvature correction value. Negative removes barrel distortion, positive removes pincushion. Default: -50")),
	), txH(orch, logger, "apply_lens_distortion_removal", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyLensDistortionRemoval(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "curvature", -50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Noise Reduction (20-23) ---

	s.AddTool(gomcp.NewTool("premiere_apply_video_noise_reduction",
		gomcp.WithDescription("Apply video noise reduction (Median) effect to reduce visual noise/grain in a clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("Noise reduction radius/amount (default: 2). Higher values = more noise reduction but softer image")),
	), txH(orch, logger, "apply_video_noise_reduction", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyVideoNoiseReduction(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", 2))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_audio_noise_reduction",
		gomcp.WithDescription("Apply audio DeNoise effect to reduce background noise on an audio clip. Uses Premiere Pro's built-in noise reduction."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("Noise reduction amount percentage (default: 50). Range: 0-100")),
	), txH(orch, logger, "apply_audio_noise_reduction", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyAudioNoiseReduction(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_de_reverb",
		gomcp.WithDescription("Apply DeReverb effect to reduce room reverb/echo on an audio clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("DeReverb amount percentage (default: 50). Range: 0-100")),
	), txH(orch, logger, "apply_de_reverb", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyDeReverb(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_de_hum",
		gomcp.WithDescription("Apply DeHum effect to remove electrical hum (50Hz or 60Hz) from an audio clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("frequency", gomcp.Description("Hum frequency to remove: 50 or 60 Hz (default: 60)")),
	), txH(orch, logger, "apply_de_hum", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyDeHum(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "frequency", 60))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Blur & Sharpen (24-27) ---

	s.AddTool(gomcp.NewTool("premiere_apply_gaussian_blur",
		gomcp.WithDescription("Apply Gaussian Blur effect to a video clip."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("blurriness", gomcp.Description("Blurriness amount (default: 10). Higher values = more blur")),
	), txH(orch, logger, "apply_gaussian_blur", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyGaussianBlur(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "blurriness", 10))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_directional_blur",
		gomcp.WithDescription("Apply Directional Blur effect to a video clip, creating motion blur along a specified direction."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("direction", gomcp.Description("Blur direction in degrees (default: 0). 0 = horizontal, 90 = vertical")),
		gomcp.WithNumber("length", gomcp.Description("Blur length/amount (default: 10)")),
	), txH(orch, logger, "apply_directional_blur", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyDirectionalBlur(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "direction", 0), gomcp.ParseFloat64(req, "length", 10))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_sharpen",
		gomcp.WithDescription("Apply Sharpen effect to a video clip to enhance edge detail."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("Sharpen amount (default: 50). Range: 0-300")),
	), txH(orch, logger, "apply_sharpen", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplySharpen(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_unsharp_mask",
		gomcp.WithDescription("Apply Unsharp Mask effect to a video clip for precise sharpening control with amount, radius, and threshold parameters."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("amount", gomcp.Description("Sharpening amount percentage (default: 100). Range: 1-500")),
		gomcp.WithNumber("radius", gomcp.Description("Radius in pixels (default: 2.0). Range: 0.1-250")),
		gomcp.WithNumber("threshold", gomcp.Description("Threshold level (default: 3). Range: 0-255")),
	), txH(orch, logger, "apply_unsharp_mask", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyUnsharpMask(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "amount", 100), gomcp.ParseFloat64(req, "radius", 2.0), gomcp.ParseFloat64(req, "threshold", 3))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// --- Distortion (28-30) ---

	s.AddTool(gomcp.NewTool("premiere_apply_mirror",
		gomcp.WithDescription("Apply Mirror effect to a video clip, reflecting the image at a specified angle around a center point."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("angle", gomcp.Description("Reflection angle in degrees (default: 0)")),
		gomcp.WithNumber("center_x", gomcp.Description("Mirror center X position as fraction of frame width 0.0-1.0 (default: 0.5)")),
		gomcp.WithNumber("center_y", gomcp.Description("Mirror center Y position as fraction of frame height 0.0-1.0 (default: 0.5)")),
	), txH(orch, logger, "apply_mirror", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplyMirror(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "angle", 0), gomcp.ParseFloat64(req, "center_x", 0.5), gomcp.ParseFloat64(req, "center_y", 0.5))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_corner_pin",
		gomcp.WithDescription("Apply Corner Pin effect to a video clip for perspective distortion. Provide four corner positions as a JSON string of [{x,y}] coordinates (normalized 0.0-1.0)."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("corners", gomcp.Required(), gomcp.Description("JSON array of 4 corner positions: '[{\"x\":0,\"y\":0},{\"x\":1,\"y\":0},{\"x\":1,\"y\":1},{\"x\":0,\"y\":1}]' (top-left, top-right, bottom-right, bottom-left)")),
	), txH(orch, logger, "apply_corner_pin", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		corners := gomcp.ParseString(req, "corners", "")
		if corners == "" {
			return gomcp.NewToolResultError("parameter 'corners' is required"), nil
		}
		result, err := orch.ApplyCornerPin(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), corners)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	s.AddTool(gomcp.NewTool("premiere_apply_spherize",
		gomcp.WithDescription("Apply Spherize effect to a video clip, creating a spherical distortion around a center point."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("radius", gomcp.Description("Spherize radius/amount (default: 50)")),
		gomcp.WithNumber("center_x", gomcp.Description("Center X position as fraction of frame width 0.0-1.0 (default: 0.5)")),
		gomcp.WithNumber("center_y", gomcp.Description("Center Y position as fraction of frame height 0.0-1.0 (default: 0.5)")),
	), txH(orch, logger, "apply_spherize", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ApplySpherize(ctx, gomcp.ParseInt(req, "track_index", 0), gomcp.ParseInt(req, "clip_index", 0), gomcp.ParseFloat64(req, "radius", 50), gomcp.ParseFloat64(req, "center_x", 0.5), gomcp.ParseFloat64(req, "center_y", 0.5))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// txH is a small wrapper that logs the tool name before delegating to the handler.
func txH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
