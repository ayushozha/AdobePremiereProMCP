package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// tplH is a small handler wrapper for template/preset/workflow tools.
func tplH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerTemplateTools registers all 30 template, preset, and workflow automation MCP tools.
func registerTemplateTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Sequence Presets (1-3)
	// -----------------------------------------------------------------------

	// 1. premiere_list_sequence_presets
	s.AddTool(gomcp.NewTool("premiere_list_sequence_presets",
		gomcp.WithDescription("List all available sequence preset (.sqpreset) files from the Premiere Pro presets directory."),
	), tplH(orch, logger, "list_sequence_presets", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListSequencePresets(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_create_sequence_from_preset
	s.AddTool(gomcp.NewTool("premiere_create_sequence_from_preset",
		gomcp.WithDescription("Create a new sequence from a .sqpreset file using the QE DOM."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the new sequence")),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the .sqpreset file")),
	), tplH(orch, logger, "create_sequence_from_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.CreateSequenceFromPreset(ctx, name, presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_export_sequence_preset
	s.AddTool(gomcp.NewTool("premiere_export_sequence_preset",
		gomcp.WithDescription("Export the settings of a sequence as a .sqpreset file."),
		gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based index of the sequence")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output .sqpreset file")),
	), tplH(orch, logger, "export_sequence_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqIndex := gomcp.ParseInt(req, "sequence_index", -1)
		if seqIndex < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportSequencePreset(ctx, seqIndex, outputPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Effect Presets (4-6)
	// -----------------------------------------------------------------------

	// 4. premiere_list_effect_presets
	s.AddTool(gomcp.NewTool("premiere_list_effect_presets",
		gomcp.WithDescription("List all available effect preset (.ffx) files."),
	), tplH(orch, logger, "list_effect_presets", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListEffectPresets(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_apply_effect_preset
	s.AddTool(gomcp.NewTool("premiere_apply_effect_preset",
		gomcp.WithDescription("Apply an effect preset (.ffx file) to a clip on the timeline."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the .ffx preset file")),
	), tplH(orch, logger, "apply_effect_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.ApplyEffectPreset(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_save_effect_preset
	s.AddTool(gomcp.NewTool("premiere_save_effect_preset",
		gomcp.WithDescription("Save the effects on a clip as a named preset."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("preset_name", gomcp.Required(), gomcp.Description("Name for the saved preset")),
	), tplH(orch, logger, "save_effect_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetName := gomcp.ParseString(req, "preset_name", "")
		if presetName == "" {
			return gomcp.NewToolResultError("parameter 'preset_name' is required"), nil
		}
		result, err := orch.SaveEffectPreset(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			presetName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Export Presets (7-9)
	// -----------------------------------------------------------------------

	// 7. premiere_list_export_presets_disk
	s.AddTool(gomcp.NewTool("premiere_list_export_presets_disk",
		gomcp.WithDescription("List all available export preset (.epr) files on disk."),
	), tplH(orch, logger, "list_export_presets_disk", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListExportPresetsFromDisk(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_create_export_preset
	s.AddTool(gomcp.NewTool("premiere_create_export_preset",
		gomcp.WithDescription("Create a custom export preset with the given settings."),
		gomcp.WithString("settings", gomcp.Required(), gomcp.Description("JSON string of export settings (codec, bitrate, resolution, etc.)")),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the new export preset")),
	), tplH(orch, logger, "create_export_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		settings := gomcp.ParseString(req, "settings", "")
		if settings == "" {
			return gomcp.NewToolResultError("parameter 'settings' is required"), nil
		}
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.CreateExportPreset(ctx, settings, name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_get_export_preset_details
	s.AddTool(gomcp.NewTool("premiere_get_export_preset_details",
		gomcp.WithDescription("Get details (codec, bitrate, resolution) of an export preset file."),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the .epr export preset file")),
	), tplH(orch, logger, "get_export_preset_details", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.GetExportPresetDetails(ctx, presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Project Templates (10-11)
	// -----------------------------------------------------------------------

	// 10. premiere_save_as_template
	s.AddTool(gomcp.NewTool("premiere_save_as_template",
		gomcp.WithDescription("Save the current project as a reusable template file."),
		gomcp.WithString("template_path", gomcp.Required(), gomcp.Description("Absolute path for the template file")),
	), tplH(orch, logger, "save_as_template", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		templatePath := gomcp.ParseString(req, "template_path", "")
		if templatePath == "" {
			return gomcp.NewToolResultError("parameter 'template_path' is required"), nil
		}
		result, err := orch.SaveAsTemplate(ctx, templatePath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_create_from_template
	s.AddTool(gomcp.NewTool("premiere_create_from_template",
		gomcp.WithDescription("Create a new project from a saved template."),
		gomcp.WithString("template_path", gomcp.Required(), gomcp.Description("Absolute path to the template file")),
		gomcp.WithString("project_path", gomcp.Required(), gomcp.Description("Absolute path for the new project file")),
	), tplH(orch, logger, "create_from_template", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		templatePath := gomcp.ParseString(req, "template_path", "")
		if templatePath == "" {
			return gomcp.NewToolResultError("parameter 'template_path' is required"), nil
		}
		projectPath := gomcp.ParseString(req, "project_path", "")
		if projectPath == "" {
			return gomcp.NewToolResultError("parameter 'project_path' is required"), nil
		}
		result, err := orch.CreateFromTemplate(ctx, templatePath, projectPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Keyboard Shortcuts (12-13)
	// -----------------------------------------------------------------------

	// 12. premiere_get_keyboard_shortcuts
	s.AddTool(gomcp.NewTool("premiere_get_keyboard_shortcuts",
		gomcp.WithDescription("List assigned keyboard shortcuts from Premiere Pro."),
	), tplH(orch, logger, "get_keyboard_shortcuts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetKeyboardShortcuts(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_execute_menu_command
	s.AddTool(gomcp.NewTool("premiere_execute_menu_command",
		gomcp.WithDescription("Execute a Premiere Pro menu command by path (e.g., 'File/Save', 'Edit/Undo')."),
		gomcp.WithString("menu_path", gomcp.Required(), gomcp.Description("Menu command path using forward slashes (e.g. 'File/Save')")),
	), tplH(orch, logger, "execute_menu_command", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		menuPath := gomcp.ParseString(req, "menu_path", "")
		if menuPath == "" {
			return gomcp.NewToolResultError("parameter 'menu_path' is required"), nil
		}
		result, err := orch.ExecuteMenuCommand(ctx, menuPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Workflow / Ingest Presets (14-16)
	// -----------------------------------------------------------------------

	// 14. premiere_create_ingest_preset
	s.AddTool(gomcp.NewTool("premiere_create_ingest_preset",
		gomcp.WithDescription("Create an ingest preset for transcode-on-import workflows."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the ingest preset")),
		gomcp.WithString("settings", gomcp.Required(), gomcp.Description("JSON string of ingest settings (codec, destination, etc.)")),
	), tplH(orch, logger, "create_ingest_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		settings := gomcp.ParseString(req, "settings", "")
		if settings == "" {
			return gomcp.NewToolResultError("parameter 'settings' is required"), nil
		}
		result, err := orch.CreateIngestPreset(ctx, name, settings)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_get_ingest_settings
	s.AddTool(gomcp.NewTool("premiere_get_ingest_settings",
		gomcp.WithDescription("Get the current ingest settings (transcode-on-import configuration)."),
	), tplH(orch, logger, "get_ingest_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetIngestSettings(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_set_ingest_settings
	s.AddTool(gomcp.NewTool("premiere_set_ingest_settings",
		gomcp.WithDescription("Enable or disable ingest (transcode-on-import) with a specified preset."),
		gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("true to enable ingest, false to disable")),
		gomcp.WithString("preset", gomcp.Description("Name or path of the ingest preset to use (ignored when disabling)")),
	), tplH(orch, logger, "set_ingest_settings", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetIngestSettings(ctx,
			gomcp.ParseBoolean(req, "enabled", false),
			gomcp.ParseString(req, "preset", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clip Presets (17-19)
	// -----------------------------------------------------------------------

	// 17. premiere_save_clip_preset
	s.AddTool(gomcp.NewTool("premiere_save_clip_preset",
		gomcp.WithDescription("Save a clip's settings (speed, effects, motion) as a reusable preset."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the clip preset")),
	), tplH(orch, logger, "save_clip_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SaveClipPreset(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_apply_clip_preset
	s.AddTool(gomcp.NewTool("premiere_apply_clip_preset",
		gomcp.WithDescription("Apply a previously saved clip preset to a clip on the timeline."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithString("preset_name", gomcp.Required(), gomcp.Description("Name of the clip preset to apply")),
	), tplH(orch, logger, "apply_clip_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetName := gomcp.ParseString(req, "preset_name", "")
		if presetName == "" {
			return gomcp.NewToolResultError("parameter 'preset_name' is required"), nil
		}
		result, err := orch.ApplyClipPreset(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0),
			presetName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_list_clip_presets
	s.AddTool(gomcp.NewTool("premiere_list_clip_presets",
		gomcp.WithDescription("List all saved clip presets."),
	), tplH(orch, logger, "list_clip_presets", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListClipPresets(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Batch Operations - extended (20-24)
	// -----------------------------------------------------------------------

	// 20. premiere_batch_rename
	s.AddTool(gomcp.NewTool("premiere_batch_rename",
		gomcp.WithDescription("Batch rename all clips on a track using a pattern (e.g., 'Shot_001', 'Shot_002')."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithString("pattern", gomcp.Required(), gomcp.Description("Naming pattern with # for sequential numbers (e.g. 'Shot_###')")),
		gomcp.WithNumber("start_number", gomcp.Description("Starting number for the sequence (default: 1)")),
	), tplH(orch, logger, "batch_rename", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		pattern := gomcp.ParseString(req, "pattern", "")
		if pattern == "" {
			return gomcp.NewToolResultError("parameter 'pattern' is required"), nil
		}
		result, err := orch.BatchRename(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			pattern,
			gomcp.ParseInt(req, "start_number", 1))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 21. premiere_batch_set_duration
	s.AddTool(gomcp.NewTool("premiere_batch_set_duration",
		gomcp.WithDescription("Set all clips on a track to the same duration."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("duration_seconds", gomcp.Required(), gomcp.Description("Target duration in seconds for each clip")),
	), tplH(orch, logger, "batch_set_duration", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.BatchSetDuration(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "duration_seconds", 5.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_batch_set_speed
	s.AddTool(gomcp.NewTool("premiere_batch_set_speed",
		gomcp.WithDescription("Set the playback speed on all clips on a track."),
		gomcp.WithString("track_type", gomcp.Required(), gomcp.Description("Track type"), gomcp.Enum("video", "audio")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("speed", gomcp.Required(), gomcp.Description("Playback speed multiplier (1.0 = normal, 2.0 = double speed, 0.5 = half speed)")),
	), tplH(orch, logger, "batch_set_speed", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.BatchSetSpeed(ctx,
			gomcp.ParseString(req, "track_type", "video"),
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseFloat64(req, "speed", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_batch_apply_transitions
	s.AddTool(gomcp.NewTool("premiere_batch_apply_transitions",
		gomcp.WithDescription("Apply a transition to all cut points on a video track."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithString("transition_name", gomcp.Required(), gomcp.Description("Name of the transition to apply (e.g. 'Cross Dissolve')")),
		gomcp.WithNumber("duration", gomcp.Description("Transition duration in seconds (default: 1.0)")),
	), tplH(orch, logger, "batch_apply_transitions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		transitionName := gomcp.ParseString(req, "transition_name", "")
		if transitionName == "" {
			return gomcp.NewToolResultError("parameter 'transition_name' is required"), nil
		}
		result, err := orch.BatchApplyTransitions(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			transitionName,
			gomcp.ParseFloat64(req, "duration", 1.0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_batch_export_frames
	s.AddTool(gomcp.NewTool("premiere_batch_export_frames",
		gomcp.WithDescription("Export the first frame of each clip on a track as an image file."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
		gomcp.WithString("output_dir", gomcp.Required(), gomcp.Description("Absolute path to the output directory")),
		gomcp.WithString("format", gomcp.Description("Image format: png, jpg, bmp (default: png)")),
	), tplH(orch, logger, "batch_export_frames", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputDir := gomcp.ParseString(req, "output_dir", "")
		if outputDir == "" {
			return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
		}
		result, err := orch.BatchExportFrames(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			outputDir,
			gomcp.ParseString(req, "format", "png"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Timeline Templates (25-27)
	// -----------------------------------------------------------------------

	// 25. premiere_save_timeline_template
	s.AddTool(gomcp.NewTool("premiere_save_timeline_template",
		gomcp.WithDescription("Save the current timeline layout as a reusable template."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the timeline template")),
		gomcp.WithString("description", gomcp.Description("Description of the template")),
	), tplH(orch, logger, "save_timeline_template", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.SaveTimelineTemplate(ctx, name, gomcp.ParseString(req, "description", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_apply_timeline_template
	s.AddTool(gomcp.NewTool("premiere_apply_timeline_template",
		gomcp.WithDescription("Apply a saved timeline template to the active sequence."),
		gomcp.WithString("template_name", gomcp.Required(), gomcp.Description("Name of the timeline template to apply")),
	), tplH(orch, logger, "apply_timeline_template", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		templateName := gomcp.ParseString(req, "template_name", "")
		if templateName == "" {
			return gomcp.NewToolResultError("parameter 'template_name' is required"), nil
		}
		result, err := orch.ApplyTimelineTemplate(ctx, templateName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_list_timeline_templates
	s.AddTool(gomcp.NewTool("premiere_list_timeline_templates",
		gomcp.WithDescription("List all available timeline templates."),
	), tplH(orch, logger, "list_timeline_templates", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListTimelineTemplates(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Macro Recording (28-30)
	// -----------------------------------------------------------------------

	// 28. premiere_start_macro_recording
	s.AddTool(gomcp.NewTool("premiere_start_macro_recording",
		gomcp.WithDescription("Start recording user actions as a macro for later playback."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name for the macro")),
	), tplH(orch, logger, "start_macro_recording", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.StartMacroRecording(ctx, name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_stop_macro_recording
	s.AddTool(gomcp.NewTool("premiere_stop_macro_recording",
		gomcp.WithDescription("Stop recording the current macro and save it."),
	), tplH(orch, logger, "stop_macro_recording", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.StopMacroRecording(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_play_macro
	s.AddTool(gomcp.NewTool("premiere_play_macro",
		gomcp.WithDescription("Play back a previously recorded macro by name."),
		gomcp.WithString("name", gomcp.Required(), gomcp.Description("Name of the macro to play")),
	), tplH(orch, logger, "play_macro", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		result, err := orch.PlayMacro(ctx, name)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
