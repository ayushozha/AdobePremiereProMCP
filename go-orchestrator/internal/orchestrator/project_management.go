package orchestrator

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Project Management — stub implementations
// These delegate to the Premiere bridge once the gRPC wiring is complete.
// ---------------------------------------------------------------------------

func (e *Engine) NewProject(ctx context.Context, path string) (*GenericResult, error) {
	if path == "" {
		return nil, fmt.Errorf("new_project: path must not be empty")
	}
	e.logger.Info("new_project", zap.String("path", path))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("NewProject(%s) — gRPC bridge not yet wired", path),
	}, nil
}

func (e *Engine) OpenProject(ctx context.Context, path string) (*GenericResult, error) {
	if path == "" {
		return nil, fmt.Errorf("open_project: path must not be empty")
	}
	e.logger.Info("open_project", zap.String("path", path))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("OpenProject(%s) — gRPC bridge not yet wired", path),
	}, nil
}

func (e *Engine) SaveProject(ctx context.Context) (*GenericResult, error) {
	e.logger.Info("save_project")
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: "SaveProject — gRPC bridge not yet wired",
	}, nil
}

func (e *Engine) SaveProjectAs(ctx context.Context, path string) (*GenericResult, error) {
	if path == "" {
		return nil, fmt.Errorf("save_project_as: path must not be empty")
	}
	e.logger.Info("save_project_as", zap.String("path", path))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("SaveProjectAs(%s) — gRPC bridge not yet wired", path),
	}, nil
}

func (e *Engine) CloseProject(ctx context.Context, saveFirst bool) (*GenericResult, error) {
	e.logger.Info("close_project", zap.Bool("save_first", saveFirst))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: "CloseProject — gRPC bridge not yet wired",
	}, nil
}

func (e *Engine) GetProjectInfo(ctx context.Context) (*ProjectInfoResult, error) {
	e.logger.Info("get_project_info")
	return &ProjectInfoResult{
		Name: "not_yet_implemented",
	}, nil
}

// ---------------------------------------------------------------------------
// Bin / Item Management
// ---------------------------------------------------------------------------

func (e *Engine) ImportFiles(ctx context.Context, filePaths []string, targetBin string) (*GenericResult, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("import_files: filePaths must not be empty")
	}
	e.logger.Info("import_files", zap.Int("count", len(filePaths)), zap.String("bin", targetBin))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("ImportFiles(%d files) — gRPC bridge not yet wired", len(filePaths)),
	}, nil
}

func (e *Engine) ImportFolder(ctx context.Context, folderPath string, targetBin string) (*GenericResult, error) {
	if folderPath == "" {
		return nil, fmt.Errorf("import_folder: folderPath must not be empty")
	}
	e.logger.Info("import_folder", zap.String("folder", folderPath), zap.String("bin", targetBin))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("ImportFolder(%s) — gRPC bridge not yet wired", folderPath),
	}, nil
}

func (e *Engine) CreateBin(ctx context.Context, name string, parentBin string) (*GenericResult, error) {
	if name == "" {
		return nil, fmt.Errorf("create_bin: name must not be empty")
	}
	e.logger.Info("create_bin", zap.String("name", name), zap.String("parent", parentBin))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("CreateBin(%s) — gRPC bridge not yet wired", name),
	}, nil
}

func (e *Engine) RenameBin(ctx context.Context, binPath string, newName string) (*GenericResult, error) {
	if binPath == "" || newName == "" {
		return nil, fmt.Errorf("rename_bin: binPath and newName must not be empty")
	}
	e.logger.Info("rename_bin", zap.String("bin", binPath), zap.String("new_name", newName))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("RenameBin(%s -> %s) — gRPC bridge not yet wired", binPath, newName),
	}, nil
}

func (e *Engine) DeleteBin(ctx context.Context, binPath string) (*GenericResult, error) {
	if binPath == "" {
		return nil, fmt.Errorf("delete_bin: binPath must not be empty")
	}
	e.logger.Info("delete_bin", zap.String("bin", binPath))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("DeleteBin(%s) — gRPC bridge not yet wired", binPath),
	}, nil
}

func (e *Engine) MoveBinItem(ctx context.Context, itemPath string, destBin string) (*GenericResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("move_bin_item: itemPath must not be empty")
	}
	e.logger.Info("move_bin_item", zap.String("item", itemPath), zap.String("dest", destBin))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("MoveBinItem(%s -> %s) — gRPC bridge not yet wired", itemPath, destBin),
	}, nil
}

func (e *Engine) FindProjectItems(ctx context.Context, searchQuery string) (*ProjectItemsResult, error) {
	if searchQuery == "" {
		return nil, fmt.Errorf("find_project_items: searchQuery must not be empty")
	}
	e.logger.Info("find_project_items", zap.String("query", searchQuery))
	return &ProjectItemsResult{
		Query:     searchQuery,
		ItemCount: 0,
		Items:     []*ProjectItemInfo{},
	}, nil
}

func (e *Engine) GetProjectItems(ctx context.Context, binPath string) (*ProjectItemsResult, error) {
	e.logger.Info("get_project_items", zap.String("bin", binPath))
	return &ProjectItemsResult{
		BinPath:   binPath,
		ItemCount: 0,
		Items:     []*ProjectItemInfo{},
	}, nil
}

func (e *Engine) SetItemLabel(ctx context.Context, itemPath string, colorIndex int) (*GenericResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("set_item_label: itemPath must not be empty")
	}
	if colorIndex < 0 || colorIndex > 15 {
		return nil, fmt.Errorf("set_item_label: colorIndex must be 0-15")
	}
	e.logger.Info("set_item_label", zap.String("item", itemPath), zap.Int("color", colorIndex))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("SetItemLabel(%s, %d) — gRPC bridge not yet wired", itemPath, colorIndex),
	}, nil
}

func (e *Engine) GetItemMetadata(ctx context.Context, itemPath string) (*ItemMetadataResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("get_item_metadata: itemPath must not be empty")
	}
	e.logger.Info("get_item_metadata", zap.String("item", itemPath))
	return &ItemMetadataResult{
		ItemPath: itemPath,
		Metadata: map[string]any{"status": "not_yet_implemented"},
	}, nil
}

func (e *Engine) SetItemMetadata(ctx context.Context, itemPath string, key string, value string) (*GenericResult, error) {
	if itemPath == "" || key == "" {
		return nil, fmt.Errorf("set_item_metadata: itemPath and key must not be empty")
	}
	e.logger.Info("set_item_metadata", zap.String("item", itemPath), zap.String("key", key))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("SetItemMetadata(%s, %s) — gRPC bridge not yet wired", itemPath, key),
	}, nil
}

// ---------------------------------------------------------------------------
// Media Management
// ---------------------------------------------------------------------------

func (e *Engine) RelinkMedia(ctx context.Context, itemPath string, newMediaPath string) (*GenericResult, error) {
	if itemPath == "" || newMediaPath == "" {
		return nil, fmt.Errorf("relink_media: itemPath and newMediaPath must not be empty")
	}
	e.logger.Info("relink_media", zap.String("item", itemPath), zap.String("new_path", newMediaPath))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("RelinkMedia(%s -> %s) — gRPC bridge not yet wired", itemPath, newMediaPath),
	}, nil
}

func (e *Engine) MakeOffline(ctx context.Context, itemPath string) (*GenericResult, error) {
	if itemPath == "" {
		return nil, fmt.Errorf("make_offline: itemPath must not be empty")
	}
	e.logger.Info("make_offline", zap.String("item", itemPath))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("MakeOffline(%s) — gRPC bridge not yet wired", itemPath),
	}, nil
}

func (e *Engine) GetOfflineItems(ctx context.Context) (*ProjectItemsResult, error) {
	e.logger.Info("get_offline_items")
	return &ProjectItemsResult{
		ItemCount: 0,
		Items:     []*ProjectItemInfo{},
	}, nil
}

// ---------------------------------------------------------------------------
// Project Settings
// ---------------------------------------------------------------------------

func (e *Engine) SetScratchDisk(ctx context.Context, scratchType string, path string) (*GenericResult, error) {
	if scratchType == "" || path == "" {
		return nil, fmt.Errorf("set_scratch_disk: type and path must not be empty")
	}
	e.logger.Info("set_scratch_disk", zap.String("type", scratchType), zap.String("path", path))
	return &GenericResult{
		Status:  "not_yet_implemented",
		Message: fmt.Sprintf("SetScratchDisk(%s, %s) — gRPC bridge not yet wired", scratchType, path),
	}, nil
}

func (e *Engine) ConsolidateDuplicates(ctx context.Context) (*ConsolidateResult, error) {
	e.logger.Info("consolidate_duplicates")
	return &ConsolidateResult{
		TotalChecked:      0,
		DuplicatesFound:   0,
		DuplicatesRemoved: 0,
	}, nil
}

func (e *Engine) GetProjectSettingsInfo(ctx context.Context) (*ProjectSettingsResult, error) {
	e.logger.Info("get_project_settings_info")
	return &ProjectSettingsResult{
		Name: "not_yet_implemented",
	}, nil
}
