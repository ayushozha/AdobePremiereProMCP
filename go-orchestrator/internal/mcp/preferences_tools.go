package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerPreferencesTools registers all Premiere Pro preferences and settings
// MCP tools (general prefs, appearance, auto-save, playback, timeline, media
// cache, label colors, GPU/renderer, and project defaults).
func registerPreferencesTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -------------------------------------------------------------------
	// General Preferences (1-4)
	// -------------------------------------------------------------------

	// 1. premiere_get_general_preferences
	s.AddTool(gomcp.NewTool("premiere_get_general_preferences",
		gomcp.WithDescription("Get general Premiere Pro preferences including default still image duration, timeline auto-scrolling, and transition durations."),
	), prefH(orch, logger, "get_general_preferences", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetGeneralPreferences(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_set_default_still_duration
	s.AddTool(gomcp.NewTool("premiere_set_default_still_duration",
		gomcp.WithDescription("Set the default duration for still images when added to the timeline."),
		gomcp.WithNumber("frames", gomcp.Required(), gomcp.Description("Duration in frames (must be a positive integer)")),
	), prefH(orch, logger, "set_default_still_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetDefaultStillDuration(ctx, gomcp.ParseInt(req, "frames", 150))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_set_default_transition_duration
	s.AddTool(gomcp.NewTool("premiere_set_default_transition_duration",
		gomcp.WithDescription("Set the default duration for video transitions."),
		gomcp.WithNumber("seconds", gomcp.Required(), gomcp.Description("Transition duration in seconds (must be positive)")),
	), prefH(orch, logger, "set_default_transition_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetDefaultTransitionDuration(ctx, gomcp.ParseFloat64(req, "seconds", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_set_default_audio_transition_duration
	s.AddTool(gomcp.NewTool("premiere_set_default_audio_transition_duration",
		gomcp.WithDescription("Set the default duration for audio transitions."),
		gomcp.WithNumber("seconds", gomcp.Required(), gomcp.Description("Audio transition duration in seconds (must be positive)")),
	), prefH(orch, logger, "set_default_audio_transition_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetDefaultAudioTransitionDuration(ctx, gomcp.ParseFloat64(req, "seconds", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Appearance (5-6)
	// -------------------------------------------------------------------

	// 5. premiere_get_brightness
	s.AddTool(gomcp.NewTool("premiere_get_brightness",
		gomcp.WithDescription("Get the current UI brightness level of Premiere Pro."),
	), prefH(orch, logger, "get_brightness", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetBrightness(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_set_brightness
	s.AddTool(gomcp.NewTool("premiere_set_brightness",
		gomcp.WithDescription("Set the UI brightness level of Premiere Pro (0 = darkest, 255 = brightest)."),
		gomcp.WithNumber("level", gomcp.Required(), gomcp.Description("Brightness level 0-255")),
	), prefH(orch, logger, "set_brightness", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetBrightness(ctx, gomcp.ParseInt(req, "level", 128))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Auto Save (7-9)
	// -------------------------------------------------------------------

	// 7. premiere_set_auto_save_enabled
	s.AddTool(gomcp.NewTool("premiere_set_auto_save_enabled",
		gomcp.WithDescription("Enable or disable the auto-save feature in Premiere Pro."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable auto-save, false to disable")),
	), prefH(orch, logger, "set_auto_save_enabled", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAutoSaveEnabled(ctx, gomcp.ParseBoolean(req, "enabled", true))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_set_auto_save_max_versions
	s.AddTool(gomcp.NewTool("premiere_set_auto_save_max_versions",
		gomcp.WithDescription("Set the maximum number of auto-save versions to keep."),
		gomcp.WithNumber("count", gomcp.Required(), gomcp.Description("Maximum number of auto-save versions (must be a positive integer)")),
	), prefH(orch, logger, "set_auto_save_max_versions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetAutoSaveMaxVersions(ctx, gomcp.ParseInt(req, "count", 20))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_get_auto_save_location
	s.AddTool(gomcp.NewTool("premiere_get_auto_save_location",
		gomcp.WithDescription("Get the file system path where Premiere Pro stores auto-save files."),
	), prefH(orch, logger, "get_auto_save_location", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetAutoSaveLocation(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Playback Preferences (10-15)
	// -------------------------------------------------------------------

	// 10. premiere_get_playback_resolution
	s.AddTool(gomcp.NewTool("premiere_get_playback_resolution",
		gomcp.WithDescription("Get the current playback resolution setting for the active sequence."),
	), prefH(orch, logger, "get_playback_resolution", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPlaybackResolution(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_set_playback_resolution
	s.AddTool(gomcp.NewTool("premiere_set_playback_resolution",
		gomcp.WithDescription("Set the playback resolution for the active sequence. Lower resolutions improve playback performance."),
		gomcp.WithString("quality", gomcp.Required(), gomcp.Description("Playback resolution quality"), gomcp.Enum("full", "1/2", "1/4", "1/8", "1/16")),
	), prefH(orch, logger, "set_playback_resolution", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		quality := gomcp.ParseString(req, "quality", "full")
		if quality == "" {
			return gomcp.NewToolResultError("parameter 'quality' is required"), nil
		}
		result, err := orch.SetPlaybackResolution(ctx, quality)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_get_preroll_frames
	s.AddTool(gomcp.NewTool("premiere_get_preroll_frames",
		gomcp.WithDescription("Get the number of pre-roll frames used during playback and recording."),
	), prefH(orch, logger, "get_preroll_frames", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPrerollFrames(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_set_preroll_frames
	s.AddTool(gomcp.NewTool("premiere_set_preroll_frames",
		gomcp.WithDescription("Set the number of pre-roll frames for playback and recording."),
		gomcp.WithNumber("frames", gomcp.Required(), gomcp.Description("Number of pre-roll frames (non-negative integer)")),
	), prefH(orch, logger, "set_preroll_frames", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetPrerollFrames(ctx, gomcp.ParseInt(req, "frames", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_get_postroll_frames
	s.AddTool(gomcp.NewTool("premiere_get_postroll_frames",
		gomcp.WithDescription("Get the number of post-roll frames used during playback and recording."),
	), prefH(orch, logger, "get_postroll_frames", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetPostrollFrames(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_set_postroll_frames
	s.AddTool(gomcp.NewTool("premiere_set_postroll_frames",
		gomcp.WithDescription("Set the number of post-roll frames for playback and recording."),
		gomcp.WithNumber("frames", gomcp.Required(), gomcp.Description("Number of post-roll frames (non-negative integer)")),
	), prefH(orch, logger, "set_postroll_frames", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetPostrollFrames(ctx, gomcp.ParseInt(req, "frames", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Timeline Preferences (16-18)
	// -------------------------------------------------------------------

	// 16. premiere_get_timeline_settings
	s.AddTool(gomcp.NewTool("premiere_get_timeline_settings",
		gomcp.WithDescription("Get all timeline preferences including time display format, auto-scroll, transition durations, and track counts."),
	), prefH(orch, logger, "get_timeline_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTimelineSettings(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_set_time_display_format
	s.AddTool(gomcp.NewTool("premiere_set_time_display_format",
		gomcp.WithDescription("Set the time display format for the timeline."),
		gomcp.WithString("format", gomcp.Required(), gomcp.Description("Time display format"), gomcp.Enum("timecode", "frames", "feet+frames")),
	), prefH(orch, logger, "set_time_display_format", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		format := gomcp.ParseString(req, "format", "timecode")
		if format == "" {
			return gomcp.NewToolResultError("parameter 'format' is required"), nil
		}
		result, err := orch.SetTimeDisplayFormat(ctx, format)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_set_video_transition_default_duration
	s.AddTool(gomcp.NewTool("premiere_set_video_transition_default_duration",
		gomcp.WithDescription("Set the default duration for video transitions in frames."),
		gomcp.WithNumber("frames", gomcp.Required(), gomcp.Description("Transition duration in frames (must be a positive integer)")),
	), prefH(orch, logger, "set_video_transition_default_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetVideoTransitionDefaultDuration(ctx, gomcp.ParseInt(req, "frames", 30))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Media Preferences (19-22)
	// -------------------------------------------------------------------

	// 19. premiere_get_media_cache_settings
	s.AddTool(gomcp.NewTool("premiere_get_media_cache_settings",
		gomcp.WithDescription("Get media cache settings including path, database path, and max size."),
	), prefH(orch, logger, "get_media_cache_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMediaCacheSettings(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_set_media_cache_location
	s.AddTool(gomcp.NewTool("premiere_set_media_cache_location",
		gomcp.WithDescription("Set the file system path where Premiere Pro stores media cache files."),
		gomcp.WithString("path", gomcp.Required(), gomcp.Description("Absolute path to the media cache directory")),
	), prefH(orch, logger, "set_media_cache_location", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.SetMediaCacheLocation(ctx, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_set_media_cache_size
	s.AddTool(gomcp.NewTool("premiere_set_media_cache_size",
		gomcp.WithDescription("Set the maximum media cache size in gigabytes."),
		gomcp.WithNumber("max_gb", gomcp.Required(), gomcp.Description("Maximum cache size in GB (must be positive)")),
	), prefH(orch, logger, "set_media_cache_size", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetMediaCacheSize(ctx, gomcp.ParseFloat64(req, "max_gb", 50))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_clean_media_cache
	s.AddTool(gomcp.NewTool("premiere_clean_media_cache",
		gomcp.WithDescription("Clean media cache files older than a specified number of days."),
		gomcp.WithNumber("days", gomcp.Required(), gomcp.Description("Delete cache files older than this many days")),
	), prefH(orch, logger, "clean_media_cache", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CleanMediaCacheOlderThan(ctx, gomcp.ParseInt(req, "days", 30))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Label Colors (23-24)
	// -------------------------------------------------------------------

	// 23. premiere_get_label_color_names
	s.AddTool(gomcp.NewTool("premiere_get_label_color_names",
		gomcp.WithDescription("Get all 16 label color names used in the project panel and timeline."),
	), prefH(orch, logger, "get_label_color_names", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetLabelColorNames(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_set_label_color_name
	s.AddTool(gomcp.NewTool("premiere_set_label_color_name",
		gomcp.WithDescription("Rename a label color by its index (0-15)."),
		gomcp.WithNumber("index", gomcp.Required(), gomcp.Description("Label color index (0-15)")),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("New name for the label color")),
	), prefH(orch, logger, "set_label_color_name", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SetLabelColorName(ctx, gomcp.ParseInt(req, "index", 0), name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// GPU / Renderer (25-27)
	// -------------------------------------------------------------------

	// 25. premiere_get_renderer_info
	s.AddTool(gomcp.NewTool("premiere_get_renderer_info",
		gomcp.WithDescription("Get the current video renderer (Software, CUDA, Metal, OpenCL) and list available renderers."),
	), prefH(orch, logger, "get_renderer_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetRendererInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_get_gpu_info
	s.AddTool(gomcp.NewTool("premiere_get_gpu_info",
		gomcp.WithDescription("Get GPU information including renderer name, GPU device name, and available memory."),
	), prefH(orch, logger, "get_gpu_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetGPUInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_set_renderer
	s.AddTool(gomcp.NewTool("premiere_set_renderer",
		gomcp.WithDescription("Set the active video renderer. May require a Premiere Pro restart to take effect."),
		gomcp.WithString("renderer_name", gomcp.Required(), gomcp.Description("Renderer name (e.g. Mercury Playback Engine GPU Acceleration (Metal))")),
	), prefH(orch, logger, "set_renderer", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		rendererName := gomcp.ParseString(req, "renderer_name", "")
		if rendererName == "" {
			return gomcp.NewToolResultError("parameter 'renderer_name' is required"), nil
		}
		result, err := orch.SetRenderer(ctx, rendererName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -------------------------------------------------------------------
	// Project Defaults (28-30)
	// -------------------------------------------------------------------

	// 28. premiere_get_default_sequence_presets
	s.AddTool(gomcp.NewTool("premiere_get_default_sequence_presets",
		gomcp.WithDescription("List all available sequence presets that can be used when creating new sequences."),
	), prefH(orch, logger, "get_default_sequence_presets", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDefaultSequencePresets(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_set_default_sequence_preset
	s.AddTool(gomcp.NewTool("premiere_set_default_sequence_preset",
		gomcp.WithDescription("Set the default sequence preset used when creating new sequences."),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the .sqpreset file")),
	), prefH(orch, logger, "set_default_sequence_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.SetDefaultSequencePreset(ctx, presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_get_installed_codecs
	s.AddTool(gomcp.NewTool("premiere_get_installed_codecs",
		gomcp.WithDescription("List all installed codecs and export formats available in Premiere Pro."),
	), prefH(orch, logger, "get_installed_codecs", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetInstalledCodecs(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// prefH is a small wrapper that logs the tool name before delegating to the handler.
func prefH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
