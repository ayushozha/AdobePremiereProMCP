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
			gomcp.WithDescription("Create a new, empty Premiere Pro project at the specified file path. The path must end with '.prproj'. Any intermediate directories must already exist. The new project becomes the active project. If another project is already open, it will be closed (you may want to save first with premiere_save_project)."),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the new project, ending with .prproj (e.g. '/Users/me/Projects/NewEdit.prproj'). Parent directories must exist."),
			),
		),
		makeNewProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 2. premiere_open_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_open_project",
			gomcp.WithDescription("Open an existing Premiere Pro project file (.prproj). Closes any currently open project. If there are unsaved changes, Premiere may prompt to save. After opening, use premiere_get_project_info to inspect the project structure."),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .prproj file to open (e.g. '/Users/me/Projects/MyEdit.prproj'). The file must exist."),
			),
		),
		makeOpenProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 3. premiere_save_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_save_project",
			gomcp.WithDescription("Save the currently open Premiere Pro project to its existing file path. Equivalent to Cmd+S / Ctrl+S. No parameters required. To save to a different location, use premiere_save_project_as instead."),
		),
		makeSaveProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 4. premiere_save_project_as
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_save_project_as",
			gomcp.WithDescription("Save the current Premiere Pro project to a new file path (Save As). The project continues working from the new path. Useful for creating backups or versioned copies."),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute file path for the saved copy, ending with .prproj (e.g. '/Users/me/Projects/MyEdit_v2.prproj'). Parent directories must exist."),
			),
		),
		makeSaveProjectAsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 5. premiere_close_project
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_close_project",
			gomcp.WithDescription("Close the currently open Premiere Pro project. Optionally save before closing. Premiere Pro remains running with no project open. Use premiere_open_project or premiere_new_project to work with a project afterward."),
			gomcp.WithBoolean("save_first",
				gomcp.Description("If true, save the project before closing. If false (default), close without saving and discard unsaved changes."),
			),
		),
		makeCloseProjectHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 6. premiere_get_project_info
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_project_info",
			gomcp.WithDescription("Retrieve detailed information about the currently open project, including project name, file path, all sequences (with indices, names, and resolutions), bin structure, total item counts, and the active sequence. This is the best starting point for understanding a project's contents."),
		),
		makeGetProjectInfoHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 7. premiere_import_files
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_import_files",
			gomcp.WithDescription("Import multiple media files into the Premiere Pro project in a single operation. Faster than calling premiere_import_media repeatedly. The files appear in the Project panel and can then be placed on timelines. Optionally import into a specific bin (created automatically if it does not exist)."),
			gomcp.WithArray("file_paths",
				gomcp.Required(),
				gomcp.Description("Array of absolute file paths to import (e.g. ['/Users/me/footage/clip01.mp4', '/Users/me/audio/narration.wav']). All paths must be absolute and the files must exist."),
				gomcp.WithStringItems(),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Slash-separated bin path to import into (e.g. 'Footage/Raw' or 'Audio/SFX'). The bin hierarchy is created automatically if it does not exist. If omitted, files are imported into the project root."),
			),
		),
		makeImportFilesHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 8. premiere_import_folder
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_import_folder",
			gomcp.WithDescription("Import all supported media files from a folder (and its subfolders) into the project. Premiere auto-detects supported formats. Useful for bulk-importing an entire shoot or asset directory. For selective imports, use premiere_import_files instead."),
			gomcp.WithString("folder_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the folder to import recursively (e.g. '/Users/me/footage/Day1'). All supported media files in the folder and subfolders will be imported."),
			),
			gomcp.WithString("target_bin",
				gomcp.Description("Slash-separated bin path to import into (e.g. 'Footage/Day1'). The bin hierarchy is created automatically if it does not exist. If omitted, files are imported into the project root."),
			),
		),
		makeImportFolderHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 9. premiere_create_bin
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_create_bin",
			gomcp.WithDescription("Create a new bin (folder) in the Premiere Pro project panel for organizing media. Bins can be nested. If the parent bin does not exist, the operation fails -- create parent bins first or use premiere_import_files with a target_bin (which auto-creates hierarchy)."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Name for the new bin (e.g. 'B-Roll', 'Interview Footage', 'Music'). Must be unique within the parent bin."),
			),
			gomcp.WithString("parent_bin",
				gomcp.Description("Slash-separated path to the parent bin (e.g. 'Footage/Day1'). If omitted, the bin is created at the project root."),
			),
		),
		makeCreateBinHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 10. premiere_rename_bin
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_rename_bin",
			gomcp.WithDescription("Rename an existing bin in the project panel. The bin keeps its contents and location, only the display name changes."),
			gomcp.WithString("bin_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the bin to rename (e.g. 'Footage/Raw'). Use premiere_get_project_items to find bin paths."),
			),
			gomcp.WithString("new_name",
				gomcp.Required(),
				gomcp.Description("New display name for the bin (e.g. 'Selects'). Must be unique within the parent bin."),
			),
		),
		makeRenameBinHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 11. premiere_delete_bin
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_delete_bin",
			gomcp.WithDescription("Delete a bin and all its contents (clips, nested bins) from the project panel. WARNING: This is destructive and cannot be undone via the MCP. Clips in the deleted bin that are used on timelines will go offline."),
			gomcp.WithString("bin_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the bin to delete (e.g. 'Footage/Unused'). Use premiere_get_project_items to find bin paths."),
			),
		),
		makeDeleteBinHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 12. premiere_move_bin_item
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_move_bin_item",
			gomcp.WithDescription("Move a project item (clip, nested sequence, or bin) from one bin to another. The item retains all its properties and timeline references. Useful for reorganizing the project panel."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the item to move (e.g. 'Footage/Raw/clip01.mp4'). Use premiere_get_project_items to find item paths."),
			),
			gomcp.WithString("dest_bin",
				gomcp.Description("Slash-separated path to the destination bin (e.g. 'Footage/Selects'). If omitted, moves to the project root."),
			),
		),
		makeMoveBinItemHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 13. premiere_find_project_items
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_find_project_items",
			gomcp.WithDescription("Search for project items by name across all bins in the project (recursive). Returns matching items with their bin paths, types, and metadata. Useful for finding specific clips or media before placing them on the timeline."),
			gomcp.WithString("query",
				gomcp.Required(),
				gomcp.Description("Search query string (case-insensitive substring match). Examples: 'interview', '.mp4', 'B-roll'. Matches against item names."),
			),
		),
		makeFindProjectItemsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 14. premiere_get_project_items
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_project_items",
			gomcp.WithDescription("List all items (clips, sequences, bins) in a specific bin of the project panel. Returns item names, types, indices, and metadata. Use this to browse the project structure and find items for editing operations."),
			gomcp.WithString("bin_path",
				gomcp.Description("Slash-separated path to the bin to list (e.g. 'Footage/Interviews'). If omitted, lists items at the project root."),
			),
		),
		makeGetProjectItemsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 15. premiere_set_item_label
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_set_item_label",
			gomcp.WithDescription("Set the label color on a project item for visual organization. Colors appear as colored bars in the Project panel and on timeline clips. Common colors: 0=Violet, 1=Iris, 2=Caribbean, 3=Lavender, 4=Cerulean, 5=Forest, 6=Rose, 7=Mango, 8=Purple, 9=Blue, 10=Teal, 11=Magenta, 12=Tan, 13=Green, 14=Brown, 15=Yellow."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item (e.g. 'Footage/clip01.mp4'). Use premiere_get_project_items to find item paths."),
			),
			gomcp.WithNumber("color_index",
				gomcp.Required(),
				gomcp.Description("Label color index (0-15). Each index corresponds to a color in Premiere Pro's label color preferences."),
			),
		),
		makeSetItemLabelHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 16. premiere_get_item_metadata
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_item_metadata",
			gomcp.WithDescription("Retrieve XMP metadata and properties for a project item, including format details, duration, frame rate, codec, and custom metadata fields. Useful for inspecting media properties before editing."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item (e.g. 'Footage/clip01.mp4'). Use premiere_get_project_items to find item paths."),
			),
		),
		makeGetItemMetadataHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 17. premiere_set_item_metadata
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_set_item_metadata",
			gomcp.WithDescription("Set an XMP metadata field on a project item. Uses the standard namespace:property format for keys. Useful for adding descriptions, ratings, scene/shot numbers, or custom tags."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item (e.g. 'Footage/clip01.mp4'). Use premiere_get_project_items to find item paths."),
			),
			gomcp.WithString("key",
				gomcp.Required(),
				gomcp.Description("Metadata key in namespace:property format. Common keys: 'dc:description' (description), 'dc:title' (title), 'xmp:Rating' (1-5 star rating), 'premiere:Scene' (scene number), 'premiere:Shot' (shot name)."),
			),
			gomcp.WithString("value",
				gomcp.Required(),
				gomcp.Description("Metadata value to set as a string. For numeric fields like Rating, pass the number as a string (e.g. '4')."),
			),
		),
		makeSetItemMetadataHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 18. premiere_relink_media
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_relink_media",
			gomcp.WithDescription("Relink an offline or missing media file to a new file path on disk. Use this when media files have been moved, renamed, or when restoring a project on a different machine. The project item keeps its timeline references. Use premiere_get_offline_items first to find items that need relinking."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item to relink (e.g. 'Footage/clip01.mp4'). Use premiere_get_offline_items to find offline items."),
			),
			gomcp.WithString("new_media_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the new media file on disk (e.g. '/Volumes/External/footage/clip01.mp4'). The file must exist and be a compatible format."),
			),
		),
		makeRelinkMediaHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 19. premiere_make_offline
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_make_offline",
			gomcp.WithDescription("Make a project item offline, disconnecting it from its media file on disk. The item remains in the project panel and on timelines but shows as offline (red 'Media Offline' screen). Use premiere_relink_media to reconnect it later."),
			gomcp.WithString("item_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the project item to take offline (e.g. 'Footage/clip01.mp4'). Use premiere_get_project_items to find item paths."),
			),
		),
		makeMakeOfflineHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 20. premiere_get_offline_items
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_offline_items",
			gomcp.WithDescription("List all offline (missing or disconnected) items in the project. Returns item paths, names, and expected media locations. Use this to identify items that need relinking with premiere_relink_media."),
		),
		makeGetOfflineItemsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 21. premiere_set_scratch_disk
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_set_scratch_disk",
			gomcp.WithDescription("Set a scratch disk path for a specific type of temporary/cache data. Scratch disks control where Premiere stores captured media, preview renders, and auto-save files. Setting this to a fast SSD can improve performance."),
			gomcp.WithString("type",
				gomcp.Required(),
				gomcp.Description("Scratch disk type to configure. 'capturedVideo'/'capturedAudio' = tape capture destination, 'videoPreview'/'audioPreview' = rendered preview files, 'autoSave' = auto-save backup location, 'cclibrary' = Creative Cloud Libraries cache."),
				gomcp.Enum("capturedVideo", "capturedAudio", "videoPreview", "audioPreview", "autoSave", "cclibrary"),
			),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the scratch disk directory (e.g. '/Volumes/FastSSD/Premiere_Cache'). The directory must exist and be writable."),
			),
		),
		makeSetScratchDiskHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 22. premiere_consolidate_duplicates
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_consolidate_duplicates",
			gomcp.WithDescription("Find and remove duplicate project items that reference the same underlying media file. Keeps the first instance and removes subsequent duplicates. Useful for cleaning up projects after multiple imports of the same footage. Returns a count of removed duplicates."),
		),
		makeConsolidateDuplicatesHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// 23. premiere_get_project_settings
	// -----------------------------------------------------------------------
	s.AddTool(
		gomcp.NewTool("premiere_get_project_settings",
			gomcp.WithDescription("Get project-level settings including the video rendering engine (Mercury GPU/Software), total item and sequence counts, scratch disk paths, and active sequence information. Useful for diagnosing performance issues or verifying project configuration."),
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
