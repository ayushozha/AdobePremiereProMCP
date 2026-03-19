package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerProjectMgmtTools registers the 23 project management MCP tools.
func registerProjectMgmtTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// 1. premiere_new_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_new_project",
			gomcp.WithDescription("Create a new Premiere Pro project at the specified path. The path should end with .prproj."),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the new project (must end with .prproj)"),
			),
		),
		makeNewProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 2. premiere_open_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_open_project",
			gomcp.WithDescription("Open an existing Premiere Pro project file (.prproj)."),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .prproj file to open"),
			),
		),
		makeOpenProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 3. premiere_save_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_save_project",
			gomcp.WithDescription("Save the currently open Premiere Pro project."),
		),
		makeSaveProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 4. premiere_save_project_as
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_save_project_as",
			gomcp.WithDescription("Save the current Premiere Pro project to a new file path."),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the saved copy (must end with .prproj)"),
			),
		),
		makeSaveProjectAsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 5. premiere_close_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_close_project",
			gomcp.WithDescription("Close the currently open Premiere Pro project."),
			gomcp.WithBoolean("save_first",
				gomcp.Description("Whether to save the project before closing (default: false)"),
			),
		),
		makeCloseProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 6. premiere_get_project_info
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_project_info",
			gomcp.WithDescription("Get detailed information about the current project including name, path, sequences, bins, and item counts."),
		),
		makeGetProjectInfoHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 7. premiere_import_files
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_import_files",
			gomcp.WithDescription("Import multiple media files into the Premiere Pro project. Optionally specify a target bin."),
			gomcp.WithArray("file_paths",
				gomcp.Required(),
				gomcp.Description("Array of absolute file paths to import"),
				gomcp.WithStringItems(),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Slash-separated bin path to import into (e.g. 'Footage/Raw'). Created if it does not exist."),
			),
		),
		makeImportFilesHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 8. premiere_import_folder
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_import_folder",
			gomcp.WithDescription("Import all media files from a folder recursively into the project."),
			gomcp.WithString("folder_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the folder to import"),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Slash-separated bin path to import into (e.g. 'Footage/Raw'). Created if it does not exist."),
			),
		),
		makeImportFolderHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 9. premiere_create_bin
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_create_bin",
			gomcp.WithDescription("Create a new bin (folder) in the Premiere Pro project panel."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Name for the new bin"),
			),
			gomcp.WithString("parent_bin",
				gomcp.Description("Slash-separated path to the parent bin (default: project root)"),
			),
		),
		makeCreateBinHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 10. premiere_rename_bin
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_rename_bin",
			gomcp.WithDescription("Rename a bin in the project panel."),
			gomcp.WithString("bin_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the bin to rename (e.g. 'Footage/Raw')"),
			),
			gomcp.WithString("new_name",
				gomcp.Required(),
				gomcp.Description("New name for the bin"),
			),
		),
		makeRenameBinHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 11. premiere_delete_bin
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_delete_bin",
			gomcp.WithDescription("Delete a bin and all its contents from the project."),
			gomcp.WithString("bin_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the bin to delete"),
			),
		),
		makeDeleteBinHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 12. premiere_move_bin_item
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_move_bin_item",
			gomcp.WithDescription("Move a project item (clip or bin) to a different bin."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the item to move"),
			),
			gomcp.WithString("dest_bin",
				gomcp.Description("Slash-separated path to the destination bin (default: project root)"),
			),
		),
		makeMoveBinItemHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 13. premiere_find_project_items
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_find_project_items",
			gomcp.WithDescription("Search for project items by name. Searches recursively through all bins."),
			gomcp.WithString("query",
				gomcp.Required(),
				gomcp.Description("Search query (case-insensitive substring match)"),
			),
		),
		makeFindProjectItemsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 14. premiere_get_project_items
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_project_items",
			gomcp.WithDescription("List all items in a specific bin of the project panel."),
			gomcp.WithString("bin_path",
				gomcp.Description("Slash-separated path to the bin (default: project root)"),
			),
		),
		makeGetProjectItemsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 15. premiere_set_item_label
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_set_item_label",
			gomcp.WithDescription("Set the label color on a project item. Colors are indexed 0-15."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item"),
			),
			gomcp.WithNumber("color_index",
				gomcp.Required(),
				gomcp.Description("Label color index (0-15)"),
			),
		),
		makeSetItemLabelHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 16. premiere_get_item_metadata
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_item_metadata",
			gomcp.WithDescription("Get XMP metadata and properties for a project item."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item"),
			),
		),
		makeGetItemMetadataHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 17. premiere_set_item_metadata
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_set_item_metadata",
			gomcp.WithDescription("Set XMP metadata on a project item. Use namespace:property format for the key (e.g. 'dc:description')."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item"),
			),
			gomcp.WithString("key",
				gomcp.Required(),
				gomcp.Description("Metadata key (e.g. 'dc:description', 'xmp:Rating')"),
			),
			gomcp.WithString("value",
				gomcp.Required(),
				gomcp.Description("Metadata value to set"),
			),
		),
		makeSetItemMetadataHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 18. premiere_relink_media
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_relink_media",
			gomcp.WithDescription("Relink an offline or missing media file to a new path on disk."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item to relink"),
			),
			gomcp.WithString("new_media_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the new media file on disk"),
			),
		),
		makeRelinkMediaHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 19. premiere_make_offline
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_make_offline",
			gomcp.WithDescription("Make a project item offline, disconnecting it from its media file."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item"),
			),
		),
		makeMakeOfflineHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 20. premiere_get_offline_items
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_offline_items",
			gomcp.WithDescription("List all offline (missing/disconnected) items in the project."),
		),
		makeGetOfflineItemsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 21. premiere_set_scratch_disk
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_set_scratch_disk",
			gomcp.WithDescription("Set a scratch disk path for the project (video capture, audio capture, previews, etc.)."),
			gomcp.WithString("type",
				gomcp.Required(),
				gomcp.Description("Scratch disk type"),
				gomcp.Enum("capturedVideo", "capturedAudio", "videoPreview", "audioPreview", "autoSave", "cclibrary"),
			),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the scratch disk directory"),
			),
		),
		makeSetScratchDiskHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 22. premiere_consolidate_duplicates
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_consolidate_duplicates",
			gomcp.WithDescription("Find and remove duplicate project items that reference the same media file."),
		),
		makeConsolidateDuplicatesHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 23. premiere_get_project_settings
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_project_settings",
			gomcp.WithDescription("Get project-level settings including renderer, item counts, and active sequence information."),
		),
		makeGetProjectSettingsHandler(orch, logger),
	)
}

// ===========================================================================
// Handler constructors
// ===========================================================================

func makeNewProjectHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_new_project")
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.NewProject(ctx, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create project: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeOpenProjectHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_open_project")
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.OpenProject(ctx, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to open project: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSaveProjectHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_save_project")
		result, err := orch.SaveProject(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to save project: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSaveProjectAsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_save_project_as")
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.SaveProjectAs(ctx, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to save project as: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeCloseProjectHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_close_project")
		saveFirst := gomcp.ParseBoolean(req, "save_first", false)
		result, err := orch.CloseProject(ctx, saveFirst)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to close project: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetProjectInfoHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_project_info")
		result, err := orch.GetProjectInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get project info: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeImportFilesHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_import_files")
		filePaths, err := extractStringSlice(req, "file_paths")
		if err != nil || len(filePaths) == 0 {
			return gomcp.NewToolResultError("parameter 'file_paths' is required and must be a non-empty array of strings"), nil
		}
		targetBin := gomcp.ParseString(req, "target_bin", "")
		result, orchErr := orch.ImportFiles(ctx, filePaths, targetBin)
		if orchErr != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to import files: %v", orchErr)), nil
		}
		return toolResultJSON(result)
	}
}

func makeImportFolderHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_import_folder")
		folderPath := gomcp.ParseString(req, "folder_path", "")
		if folderPath == "" {
			return gomcp.NewToolResultError("parameter 'folder_path' is required"), nil
		}
		targetBin := gomcp.ParseString(req, "target_bin", "")
		result, err := orch.ImportFolder(ctx, folderPath, targetBin)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to import folder: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeCreateBinHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_create_bin")
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		parentBin := gomcp.ParseString(req, "parent_bin", "")
		result, err := orch.CreateBin(ctx, name, parentBin)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create bin: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRenameBinHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_rename_bin")
		binPath := gomcp.ParseString(req, "bin_path", "")
		if binPath == "" {
			return gomcp.NewToolResultError("parameter 'bin_path' is required"), nil
		}
		newName := gomcp.ParseString(req, "new_name", "")
		if newName == "" {
			return gomcp.NewToolResultError("parameter 'new_name' is required"), nil
		}
		result, err := orch.RenameBin(ctx, binPath, newName)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to rename bin: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDeleteBinHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_delete_bin")
		binPath := gomcp.ParseString(req, "bin_path", "")
		if binPath == "" {
			return gomcp.NewToolResultError("parameter 'bin_path' is required"), nil
		}
		result, err := orch.DeleteBin(ctx, binPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to delete bin: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeMoveBinItemHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_move_bin_item")
		itemPath := gomcp.ParseString(req, "item_path", "")
		if itemPath == "" {
			return gomcp.NewToolResultError("parameter 'item_path' is required"), nil
		}
		destBin := gomcp.ParseString(req, "dest_bin", "")
		result, err := orch.MoveBinItem(ctx, itemPath, destBin)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to move item: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeFindProjectItemsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_find_project_items")
		query := gomcp.ParseString(req, "query", "")
		if query == "" {
			return gomcp.NewToolResultError("parameter 'query' is required"), nil
		}
		result, err := orch.FindProjectItems(ctx, query)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to find items: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetProjectItemsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_project_items")
		binPath := gomcp.ParseString(req, "bin_path", "")
		result, err := orch.GetProjectItems(ctx, binPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get items: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetItemLabelHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_item_label")
		itemPath := gomcp.ParseString(req, "item_path", "")
		if itemPath == "" {
			return gomcp.NewToolResultError("parameter 'item_path' is required"), nil
		}
		colorIndex := gomcp.ParseInt(req, "color_index", -1)
		if colorIndex < 0 || colorIndex > 15 {
			return gomcp.NewToolResultError("parameter 'color_index' must be between 0 and 15"), nil
		}
		result, err := orch.SetItemLabel(ctx, itemPath, colorIndex)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set label: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetItemMetadataHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_item_metadata")
		itemPath := gomcp.ParseString(req, "item_path", "")
		if itemPath == "" {
			return gomcp.NewToolResultError("parameter 'item_path' is required"), nil
		}
		result, err := orch.GetItemMetadata(ctx, itemPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get metadata: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetItemMetadataHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_item_metadata")
		itemPath := gomcp.ParseString(req, "item_path", "")
		if itemPath == "" {
			return gomcp.NewToolResultError("parameter 'item_path' is required"), nil
		}
		key := gomcp.ParseString(req, "key", "")
		if key == "" {
			return gomcp.NewToolResultError("parameter 'key' is required"), nil
		}
		value := gomcp.ParseString(req, "value", "")
		if value == "" {
			return gomcp.NewToolResultError("parameter 'value' is required"), nil
		}
		result, err := orch.SetItemMetadata(ctx, itemPath, key, value)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set metadata: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRelinkMediaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_relink_media")
		itemPath := gomcp.ParseString(req, "item_path", "")
		if itemPath == "" {
			return gomcp.NewToolResultError("parameter 'item_path' is required"), nil
		}
		newMediaPath := gomcp.ParseString(req, "new_media_path", "")
		if newMediaPath == "" {
			return gomcp.NewToolResultError("parameter 'new_media_path' is required"), nil
		}
		result, err := orch.RelinkMedia(ctx, itemPath, newMediaPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to relink media: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeMakeOfflineHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_make_offline")
		itemPath := gomcp.ParseString(req, "item_path", "")
		if itemPath == "" {
			return gomcp.NewToolResultError("parameter 'item_path' is required"), nil
		}
		result, err := orch.MakeOffline(ctx, itemPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to make offline: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetOfflineItemsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_offline_items")
		result, err := orch.GetOfflineItems(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get offline items: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetScratchDiskHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_scratch_disk")
		scratchType := gomcp.ParseString(req, "type", "")
		if scratchType == "" {
			return gomcp.NewToolResultError("parameter 'type' is required"), nil
		}
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.SetScratchDisk(ctx, scratchType, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set scratch disk: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeConsolidateDuplicatesHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_consolidate_duplicates")
		result, err := orch.ConsolidateDuplicates(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to consolidate duplicates: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetProjectSettingsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_project_settings")
		result, err := orch.GetProjectSettingsInfo(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get project settings: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}
