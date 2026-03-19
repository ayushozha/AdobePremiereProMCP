package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Project Management — real implementations via EvalCommand
// ---------------------------------------------------------------------------

func (e *Engine) NewProject(ctx context.Context, path string) (*GenericResult, error) {
	if path == "" {
		return nil, fmt.Errorf("new_project: path is required — provide the full path for the new project (e.g. /Users/you/Projects/MyVideo.prproj)")
	}
	e.logger.Info("new_project", zap.String("path", path))
	argsJSON, _ := json.Marshal(map[string]any{
		"path": path,
	})
	result, err := e.premiere.EvalCommand(ctx, "newProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create new project at %q — make sure Premiere Pro is running: %w", path, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) OpenProject(ctx context.Context, path string) (*GenericResult, error) {
	if path == "" {
		return nil, fmt.Errorf("open_project: path is required — provide the full path to a .prproj file (e.g. /Users/you/Projects/MyVideo.prproj)")
	}
	e.logger.Info("open_project", zap.String("path", path))
	argsJSON, _ := json.Marshal(map[string]any{
		"path": path,
	})
	result, err := e.premiere.EvalCommand(ctx, "openProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to open project %q — verify the file exists and is a valid .prproj file: %w", path, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) SaveProject(ctx context.Context) (*GenericResult, error) {
	e.logger.Info("save_project")
	result, err := e.premiere.EvalCommand(ctx, "saveProject", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to save project — open a project first with premiere_open_project: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) SaveProjectAs(ctx context.Context, path string) (*GenericResult, error) {
	if path == "" {
		return nil, fmt.Errorf("save_project_as: path is required — provide the full destination path (e.g. /Users/you/Projects/MyVideo_copy.prproj)")
	}
	e.logger.Info("save_project_as", zap.String("path", path))
	argsJSON, _ := json.Marshal(map[string]any{
		"path": path,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveProjectAs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to save project as %q — open a project first with premiere_open_project: %w", path, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) CloseProject(ctx context.Context, saveFirst bool) (*GenericResult, error) {
	e.logger.Info("close_project", zap.Bool("save_first", saveFirst))
	argsJSON, _ := json.Marshal(map[string]any{
		"saveFirst": saveFirst,
	})
	result, err := e.premiere.EvalCommand(ctx, "closeProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to close project — no project may be open: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) GetProjectInfo(ctx context.Context) (*ProjectInfoResult, error) {
	e.logger.Info("get_project_info")
	result, err := e.premiere.EvalCommand(ctx, "getProjectInfo", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to get project info — open a project first with premiere_open_project: %w", err)
	}
	var out ProjectInfoResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetProjectInfo: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Bin / Item Management
// ---------------------------------------------------------------------------

func (e *Engine) ImportFiles(ctx context.Context, filePaths []string, targetBin string) (*GenericResult, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("import_files: at least one file path is required — provide full paths to media files")
	}
	e.logger.Info("import_files", zap.Int("count", len(filePaths)), zap.String("bin", targetBin))
	argsJSON, _ := json.Marshal(map[string]any{
		"filePaths": filePaths,
		"targetBin": targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importFiles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to import %d files — make sure a project is open (try premiere_open_project) and files exist: %w", len(filePaths), err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) ImportFolder(ctx context.Context, folderPath string, targetBin string) (*GenericResult, error) {
	if folderPath == "" {
		return nil, fmt.Errorf("import_folder: folder path is required — provide the full path to a folder with media files")
	}
	e.logger.Info("import_folder", zap.String("folder", folderPath), zap.String("bin", targetBin))
	argsJSON, _ := json.Marshal(map[string]any{
		"folderPath": folderPath,
		"targetBin":  targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importFolder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to import folder %q — verify the folder exists and a project is open: %w", folderPath, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) CreateBin(ctx context.Context, name string, parentBin string) (*GenericResult, error) {
	if name == "" {
		return nil, fmt.Errorf("create_bin: bin name is required")
	}
	e.logger.Info("create_bin", zap.String("name", name), zap.String("parent", parentBin))
	argsJSON, _ := json.Marshal(map[string]any{
		"name":      name,
		"parentBin": parentBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "createBin", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create bin %q — open a project first with premiere_open_project: %w", name, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) RenameBin(ctx context.Context, binPath string, newName string) (*GenericResult, error) {
	if binPath == "" || newName == "" {
		return nil, fmt.Errorf("rename_bin: both binPath and newName are required — use premiere_get_project_items to find bin paths")
	}
	e.logger.Info("rename_bin", zap.String("bin", binPath), zap.String("new_name", newName))
	argsJSON, _ := json.Marshal(map[string]any{
		"binPath": binPath,
		"newName": newName,
	})
	result, err := e.premiere.EvalCommand(ctx, "renameBin", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to rename bin %q to %q: %w", binPath, newName, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) DeleteBin(ctx context.Context, binPath string) (*GenericResult, error) {
	if binPath == "" {
		return nil, fmt.Errorf("delete_bin: binPath is required — use premiere_get_project_items to find bin paths")
	}
	e.logger.Info("delete_bin", zap.String("bin", binPath))
	argsJSON, _ := json.Marshal(map[string]any{
		"binPath": binPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteBin", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to delete bin %q: %w", binPath, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) MoveBinItem(ctx context.Context, itemPath string, destBin string) (*GenericResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("move_bin_item: itemPath is required — use premiere_get_project_items to find item paths")
	}
	e.logger.Info("move_bin_item", zap.String("item", itemPath), zap.String("dest", destBin))
	argsJSON, _ := json.Marshal(map[string]any{
		"itemPath": itemPath,
		"destBin":  destBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "moveBinItem", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to move item %q to bin %q: %w", itemPath, destBin, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) FindProjectItems(ctx context.Context, searchQuery string) (*ProjectItemsResult, error) {
	if searchQuery == "" {
		return nil, fmt.Errorf("find_project_items: searchQuery is required — enter a name or keyword to search for")
	}
	e.logger.Info("find_project_items", zap.String("query", searchQuery))
	argsJSON, _ := json.Marshal(map[string]any{
		"searchQuery": searchQuery,
	})
	result, err := e.premiere.EvalCommand(ctx, "findProjectItems", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to search for %q — open a project first with premiere_open_project: %w", searchQuery, err)
	}
	var out ProjectItemsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("FindProjectItems: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

func (e *Engine) GetProjectItems(ctx context.Context, binPath string) (*ProjectItemsResult, error) {
	e.logger.Info("get_project_items", zap.String("bin", binPath))
	argsJSON, _ := json.Marshal(map[string]any{
		"binPath": binPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getProjectItems", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to list project items — open a project first with premiere_open_project: %w", err)
	}
	var out ProjectItemsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetProjectItems: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

func (e *Engine) SetItemLabel(ctx context.Context, itemPath string, colorIndex int) (*GenericResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("set_item_label: itemPath is required — use premiere_get_project_items to find item paths")
	}
	if colorIndex < 0 || colorIndex > 15 {
		return nil, fmt.Errorf("set_item_label: colorIndex must be 0-15 (Premiere Pro label colors)")
	}
	e.logger.Info("set_item_label", zap.String("item", itemPath), zap.Int("color", colorIndex))
	argsJSON, _ := json.Marshal(map[string]any{
		"itemPath":   itemPath,
		"colorIndex": colorIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "setItemLabel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set label on %q: %w", itemPath, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) GetItemMetadata(ctx context.Context, itemPath string) (*ItemMetadataResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("get_item_metadata: itemPath is required — use premiere_get_project_items to find item paths")
	}
	e.logger.Info("get_item_metadata", zap.String("item", itemPath))
	argsJSON, _ := json.Marshal(map[string]any{
		"itemPath": itemPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getItemMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata for %q: %w", itemPath, err)
	}
	var out ItemMetadataResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetItemMetadata: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

func (e *Engine) SetItemMetadata(ctx context.Context, itemPath string, key string, value string) (*GenericResult, error) {
	if itemPath == "" || key == "" {
		return nil, fmt.Errorf("set_item_metadata: both itemPath and key are required")
	}
	e.logger.Info("set_item_metadata", zap.String("item", itemPath), zap.String("key", key))
	argsJSON, _ := json.Marshal(map[string]any{
		"itemPath": itemPath,
		"key":      key,
		"value":    value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setItemMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set metadata %q on %q: %w", key, itemPath, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

// ---------------------------------------------------------------------------
// Media Management
// ---------------------------------------------------------------------------

func (e *Engine) RelinkMedia(ctx context.Context, itemPath string, newMediaPath string) (*GenericResult, error) {
	if itemPath == "" || newMediaPath == "" {
		return nil, fmt.Errorf("relink_media: both itemPath and newMediaPath are required — use premiere_get_offline_items to find items that need relinking")
	}
	e.logger.Info("relink_media", zap.String("item", itemPath), zap.String("new_path", newMediaPath))
	argsJSON, _ := json.Marshal(map[string]any{
		"itemPath":     itemPath,
		"newMediaPath": newMediaPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "relinkMedia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to relink %q to %q — verify the new media file exists: %w", itemPath, newMediaPath, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) MakeOffline(ctx context.Context, itemPath string) (*GenericResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("make_offline: itemPath is required — use premiere_get_project_items to find item paths")
	}
	e.logger.Info("make_offline", zap.String("item", itemPath))
	argsJSON, _ := json.Marshal(map[string]any{
		"itemPath": itemPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "makeOffline", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to make %q offline: %w", itemPath, err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) GetOfflineItems(ctx context.Context) (*ProjectItemsResult, error) {
	e.logger.Info("get_offline_items")
	result, err := e.premiere.EvalCommand(ctx, "getOfflineItems", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to get offline items — open a project first with premiere_open_project: %w", err)
	}
	var out ProjectItemsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetOfflineItems: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Project Settings
// ---------------------------------------------------------------------------

func (e *Engine) SetScratchDisk(ctx context.Context, scratchType string, path string) (*GenericResult, error) {
	if scratchType == "" || path == "" {
		return nil, fmt.Errorf("set_scratch_disk: both type and path are required")
	}
	e.logger.Info("set_scratch_disk", zap.String("type", scratchType), zap.String("path", path))
	argsJSON, _ := json.Marshal(map[string]any{
		"scratchType": scratchType,
		"path":        path,
	})
	result, err := e.premiere.EvalCommand(ctx, "setScratchDisk", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to set scratch disk — open a project first with premiere_open_project: %w", err)
	}
	return &GenericResult{
		Status:  "ok",
		Message: result,
	}, nil
}

func (e *Engine) ConsolidateDuplicates(ctx context.Context) (*ConsolidateResult, error) {
	e.logger.Info("consolidate_duplicates")
	result, err := e.premiere.EvalCommand(ctx, "consolidateDuplicates", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to consolidate duplicates — open a project first with premiere_open_project: %w", err)
	}
	var out ConsolidateResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("ConsolidateDuplicates: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}

func (e *Engine) GetProjectSettingsInfo(ctx context.Context) (*ProjectSettingsResult, error) {
	e.logger.Info("get_project_settings_info")
	result, err := e.premiere.EvalCommand(ctx, "getProjectSettings", "{}")
	if err != nil {
		return nil, fmt.Errorf("failed to get project settings — open a project first with premiere_open_project: %w", err)
	}
	var out ProjectSettingsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetProjectSettingsInfo: could not parse response from Premiere Pro: %w", err)
	}
	return &out, nil
}
