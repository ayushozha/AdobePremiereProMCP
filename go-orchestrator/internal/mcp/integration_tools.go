package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// intH is a small handler wrapper for integration tools.
func intH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerIntegrationTools registers all 30 interoperability, integration,
// and cross-app MCP tools for After Effects, Photoshop, Audition, Media
// Encoder, Dynamic Link, codec, project interchange, clipboard, external
// tools, Team Projects, and Productions.
func registerIntegrationTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// After Effects Integration (1-4)
	// -----------------------------------------------------------------------

	// 1. premiere_send_to_after_effects
	s.AddTool(gomcp.NewTool("premiere_send_to_after_effects",
		gomcp.WithDescription("Replace a project item with a Dynamic Link After Effects composition. Launches AE if needed."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item to send to After Effects")),
	), intH(orch, logger, "send_to_after_effects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.SendToAfterEffects(ctx, idx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_import_ae_comp
	s.AddTool(gomcp.NewTool("premiere_import_ae_comp",
		gomcp.WithDescription("Import a specific After Effects composition from a .aep project file via Dynamic Link."),
		gomcp.WithString("aep_path", gomcp.Required(), gomcp.Description("Absolute path to the After Effects .aep project file")),
		gomcp.WithString("comp_name", gomcp.Required(), gomcp.Description("Name of the composition to import")),
		gomcp.WithString("target_bin", gomcp.Description("Name of the bin to import into (default: root)")),
	), intH(orch, logger, "import_ae_comp", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		aepPath := gomcp.ParseString(req, "aep_path", "")
		if aepPath == "" {
			return gomcp.NewToolResultError("parameter 'aep_path' is required"), nil
		}
		compName := gomcp.ParseString(req, "comp_name", "")
		if compName == "" {
			return gomcp.NewToolResultError("parameter 'comp_name' is required"), nil
		}
		result, err := orch.ImportAEComp(ctx, aepPath, compName, gomcp.ParseString(req, "target_bin", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_import_all_ae_comps
	s.AddTool(gomcp.NewTool("premiere_import_all_ae_comps",
		gomcp.WithDescription("Import all After Effects compositions from a .aep project file."),
		gomcp.WithString("aep_path", gomcp.Required(), gomcp.Description("Absolute path to the After Effects .aep project file")),
		gomcp.WithString("target_bin", gomcp.Description("Name of the bin to import into (default: root)")),
	), intH(orch, logger, "import_all_ae_comps", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		aepPath := gomcp.ParseString(req, "aep_path", "")
		if aepPath == "" {
			return gomcp.NewToolResultError("parameter 'aep_path' is required"), nil
		}
		result, err := orch.ImportAllAEComps(ctx, aepPath, gomcp.ParseString(req, "target_bin", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_refresh_ae_comp
	s.AddTool(gomcp.NewTool("premiere_refresh_ae_comp",
		gomcp.WithDescription("Refresh a linked Dynamic Link After Effects composition to pull latest changes."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the linked AE composition project item")),
	), intH(orch, logger, "refresh_ae_comp", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.RefreshAEComp(ctx, idx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Photoshop Integration (5-6)
	// -----------------------------------------------------------------------

	// 5. premiere_edit_in_photoshop
	s.AddTool(gomcp.NewTool("premiere_edit_in_photoshop",
		gomcp.WithDescription("Open a project item (frame or image) in Adobe Photoshop for editing via BridgeTalk."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item to open in Photoshop")),
	), intH(orch, logger, "edit_in_photoshop", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.EditInPhotoshop(ctx, idx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_import_psd_layers
	s.AddTool(gomcp.NewTool("premiere_import_psd_layers",
		gomcp.WithDescription("Import a Photoshop PSD file with layer support, optionally as a sequence."),
		gomcp.WithString("psd_path", gomcp.Required(), gomcp.Description("Absolute path to the Photoshop .psd file")),
		gomcp.WithString("target_bin", gomcp.Description("Name of the bin to import into (default: root)")),
		gomcp.WithBoolean("as_sequence", gomcp.Description("If true, import layers as a sequence; otherwise as individual items (default: false)")),
	), intH(orch, logger, "import_psd_layers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		psdPath := gomcp.ParseString(req, "psd_path", "")
		if psdPath == "" {
			return gomcp.NewToolResultError("parameter 'psd_path' is required"), nil
		}
		result, err := orch.ImportPSDLayers(ctx, psdPath, gomcp.ParseString(req, "target_bin", ""), gomcp.ParseBoolean(req, "as_sequence", false))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Audition Integration (7-8)
	// -----------------------------------------------------------------------

	// 7. premiere_edit_in_audition
	s.AddTool(gomcp.NewTool("premiere_edit_in_audition",
		gomcp.WithDescription("Send an audio clip from the active sequence to Adobe Audition for editing."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track")),
	), intH(orch, logger, "edit_in_audition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.EditInAudition(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_refresh_audition_edit
	s.AddTool(gomcp.NewTool("premiere_refresh_audition_edit",
		gomcp.WithDescription("Refresh an audio clip after editing in Adobe Audition to pull back changes."),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track")),
	), intH(orch, logger, "refresh_audition_edit", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RefreshAuditionEdit(ctx,
			gomcp.ParseInt(req, "track_index", 0),
			gomcp.ParseInt(req, "clip_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Media Encoder Integration (9-11)
	// -----------------------------------------------------------------------

	// 9. premiere_queue_in_media_encoder
	s.AddTool(gomcp.NewTool("premiere_queue_in_media_encoder",
		gomcp.WithDescription("Queue a sequence in Adobe Media Encoder for batch encoding."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (-1 or omitted for active sequence)")),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the AME export preset file (.epr)")),
	), intH(orch, logger, "queue_in_media_encoder", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.QueueInMediaEncoder(ctx,
			gomcp.ParseInt(req, "sequence_index", -1),
			presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_get_media_encoder_queue
	s.AddTool(gomcp.NewTool("premiere_get_media_encoder_queue",
		gomcp.WithDescription("Get the current Adobe Media Encoder queue status."),
	), intH(orch, logger, "get_media_encoder_queue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetMediaEncoderQueue(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_clear_media_encoder_queue
	s.AddTool(gomcp.NewTool("premiere_clear_media_encoder_queue",
		gomcp.WithDescription("Clear all items from the Adobe Media Encoder queue."),
	), intH(orch, logger, "clear_media_encoder_queue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ClearMediaEncoderQueue(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Dynamic Link (12-13)
	// -----------------------------------------------------------------------

	// 12. premiere_get_dynamic_link_status
	s.AddTool(gomcp.NewTool("premiere_get_dynamic_link_status",
		gomcp.WithDescription("Get Dynamic Link connection status showing all linked After Effects compositions."),
	), intH(orch, logger, "get_dynamic_link_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDynamicLinkStatus(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_refresh_all_dynamic_links
	s.AddTool(gomcp.NewTool("premiere_refresh_all_dynamic_links",
		gomcp.WithDescription("Refresh all Dynamic Link clips to pull latest changes from linked applications."),
	), intH(orch, logger, "refresh_all_dynamic_links", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RefreshAllDynamicLinks(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// File Format Support / Codec (14-16)
	// -----------------------------------------------------------------------

	// 14. premiere_get_codec_info
	s.AddTool(gomcp.NewTool("premiere_get_codec_info",
		gomcp.WithDescription("Get detailed codec information for a project item including video codec, audio codec, frame size, and duration."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item")),
	), intH(orch, logger, "get_codec_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetCodecInfo(ctx, idx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_transcode_clip
	s.AddTool(gomcp.NewTool("premiere_transcode_clip",
		gomcp.WithDescription("Transcode a project item clip to a new format using an export preset via Adobe Media Encoder."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item to transcode")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the transcoded output file")),
		gomcp.WithString("preset_path", gomcp.Required(), gomcp.Description("Absolute path to the export preset file (.epr)")),
	), intH(orch, logger, "transcode_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		presetPath := gomcp.ParseString(req, "preset_path", "")
		if presetPath == "" {
			return gomcp.NewToolResultError("parameter 'preset_path' is required"), nil
		}
		result, err := orch.TranscodeClip(ctx, idx, outputPath, presetPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_conform_media
	s.AddTool(gomcp.NewTool("premiere_conform_media",
		gomcp.WithDescription("Conform media to target specifications by adjusting footage interpretation (frame rate, codec)."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item to conform")),
		gomcp.WithNumber("target_fps", gomcp.Required(), gomcp.Description("Target frame rate to conform to (e.g. 23.976, 24, 29.97, 30)")),
		gomcp.WithString("target_codec", gomcp.Description("Target codec hint (informational, actual transcoding uses presets)")),
	), intH(orch, logger, "conform_media", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		targetFps := gomcp.ParseFloat64(req, "target_fps", 0)
		if targetFps <= 0 {
			return gomcp.NewToolResultError("parameter 'target_fps' must be a positive number"), nil
		}
		result, err := orch.ConformMedia(ctx, idx, targetFps, gomcp.ParseString(req, "target_codec", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Project Interchange (17-20)
	// -----------------------------------------------------------------------

	// 17. premiere_export_as_omf (extended — existing exportOMF covers basic; this adds target bin import)
	// Note: exportOMF already exists as premiere_export_omf. This adds the import side.

	// 18. premiere_import_omf
	s.AddTool(gomcp.NewTool("premiere_import_omf",
		gomcp.WithDescription("Import an OMF (Open Media Framework) file into the Premiere Pro project."),
		gomcp.WithString("omf_path", gomcp.Required(), gomcp.Description("Absolute path to the OMF file to import")),
		gomcp.WithString("target_bin", gomcp.Description("Name of the bin to import into (default: root)")),
	), intH(orch, logger, "import_omf", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		omfPath := gomcp.ParseString(req, "omf_path", "")
		if omfPath == "" {
			return gomcp.NewToolResultError("parameter 'omf_path' is required"), nil
		}
		result, err := orch.ImportOMFFile(ctx, omfPath, gomcp.ParseString(req, "target_bin", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_export_as_aaf (extended — existing exportAAF covers basic; this adds target bin import)
	// Note: exportAAF already exists as premiere_export_aaf. This adds the import side.

	// 20. premiere_import_aaf_file
	s.AddTool(gomcp.NewTool("premiere_import_aaf_file",
		gomcp.WithDescription("Import an AAF (Advanced Authoring Format) file into the Premiere Pro project with target bin support."),
		gomcp.WithString("aaf_path", gomcp.Required(), gomcp.Description("Absolute path to the AAF file to import")),
		gomcp.WithString("target_bin", gomcp.Description("Name of the bin to import into (default: root)")),
	), intH(orch, logger, "import_aaf_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		aafPath := gomcp.ParseString(req, "aaf_path", "")
		if aafPath == "" {
			return gomcp.NewToolResultError("parameter 'aaf_path' is required"), nil
		}
		result, err := orch.ImportAAFFile(ctx, aafPath, gomcp.ParseString(req, "target_bin", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Clipboard (21-22)
	// -----------------------------------------------------------------------

	// 21. premiere_copy_to_clipboard
	s.AddTool(gomcp.NewTool("premiere_copy_to_clipboard",
		gomcp.WithDescription("Copy text to the system clipboard from within Premiere Pro."),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Text content to copy to the clipboard")),
	), intH(orch, logger, "copy_to_clipboard", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.CopyToSystemClipboard(ctx, text)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_get_from_clipboard
	s.AddTool(gomcp.NewTool("premiere_get_from_clipboard",
		gomcp.WithDescription("Read text content from the system clipboard."),
	), intH(orch, logger, "get_from_clipboard", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFromSystemClipboard(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// External Tools (23-24)
	// -----------------------------------------------------------------------

	// 23. premiere_open_in_external_editor
	s.AddTool(gomcp.NewTool("premiere_open_in_external_editor",
		gomcp.WithDescription("Open a project item's media in an external editor application."),
		gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item")),
		gomcp.WithString("editor_path", gomcp.Required(), gomcp.Description("Absolute path to the external editor application")),
	), intH(orch, logger, "open_in_external_editor", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "project_item_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		editorPath := gomcp.ParseString(req, "editor_path", "")
		if editorPath == "" {
			return gomcp.NewToolResultError("parameter 'editor_path' is required"), nil
		}
		result, err := orch.OpenInExternalEditor(ctx, idx, editorPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_import_from_external_source
	s.AddTool(gomcp.NewTool("premiere_import_from_external_source",
		gomcp.WithDescription("Import media from an external source path, with optional format hint for special handling."),
		gomcp.WithString("source_path", gomcp.Required(), gomcp.Description("Absolute path to the source media file")),
		gomcp.WithString("format", gomcp.Description("Format hint for special handling (e.g. 'fcpxml', 'edl', 'aaf', 'auto'). Default: auto")),
	), intH(orch, logger, "import_from_external_source", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		sourcePath := gomcp.ParseString(req, "source_path", "")
		if sourcePath == "" {
			return gomcp.NewToolResultError("parameter 'source_path' is required"), nil
		}
		result, err := orch.ImportFromExternalSource(ctx, sourcePath, gomcp.ParseString(req, "format", "auto"))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Team Projects (25-27)
	// -----------------------------------------------------------------------

	// 25. premiere_get_team_project_status
	s.AddTool(gomcp.NewTool("premiere_get_team_project_status",
		gomcp.WithDescription("Get Team Projects connection status, including whether the current project is a team project."),
	), intH(orch, logger, "get_team_project_status", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetTeamProjectStatus(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 26. premiere_check_in_changes
	s.AddTool(gomcp.NewTool("premiere_check_in_changes",
		gomcp.WithDescription("Check in (share) changes to Team Projects with a commit message."),
		gomcp.WithString("message", gomcp.Description("Commit message describing the changes (default: empty)")),
	), intH(orch, logger, "check_in_changes", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckInChanges(ctx, gomcp.ParseString(req, "message", ""))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_check_out_sequence
	s.AddTool(gomcp.NewTool("premiere_check_out_sequence",
		gomcp.WithDescription("Check out a sequence for exclusive editing in Team Projects."),
		gomcp.WithNumber("sequence_index", gomcp.Required(), gomcp.Description("Zero-based index of the sequence to check out")),
	), intH(orch, logger, "check_out_sequence", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		idx := gomcp.ParseInt(req, "sequence_index", -1)
		if idx < 0 {
			return gomcp.NewToolResultError("parameter 'sequence_index' is required"), nil
		}
		result, err := orch.CheckOutSequence(ctx, idx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Productions (28-30)
	// -----------------------------------------------------------------------

	// 28. premiere_get_production_info
	s.AddTool(gomcp.NewTool("premiere_get_production_info",
		gomcp.WithDescription("Get information about the current production (Productions feature, Premiere Pro 2020+)."),
	), intH(orch, logger, "get_production_info", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetProductionInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_list_production_projects
	s.AddTool(gomcp.NewTool("premiere_list_production_projects",
		gomcp.WithDescription("List all projects within the current production."),
	), intH(orch, logger, "list_production_projects", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ListProductionProjects(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_open_production_project
	s.AddTool(gomcp.NewTool("premiere_open_production_project",
		gomcp.WithDescription("Open a specific project by name from within the current production."),
		gomcp.WithString("project_name", gomcp.Required(), gomcp.Description("Name of the project to open from the production")),
	), intH(orch, logger, "open_production_project", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		projectName := gomcp.ParseString(req, "project_name", "")
		if projectName == "" {
			return gomcp.NewToolResultError("parameter 'project_name' is required"), nil
		}
		result, err := orch.OpenProductionProject(ctx, projectName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
