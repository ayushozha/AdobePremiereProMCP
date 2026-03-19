package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Clip/Item Metadata Operations
// ---------------------------------------------------------------------------

// GetClipMetadata retrieves all metadata (XMP + project metadata) for a project item.
func (e *Engine) GetClipMetadata(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetClipMetadata sets a metadata field on a project item.
func (e *Engine) SetClipMetadata(ctx context.Context, projectItemIndex int, field, value string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"field": field,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddCustomMetadataField adds a custom metadata schema field to the project.
func (e *Engine) AddCustomMetadataField(ctx context.Context, fieldName, fieldLabel string, fieldType int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"fieldName": fieldName,
		"fieldLabel": fieldLabel,
		"fieldType": fieldType,
	})
	result, err := e.premiere.EvalCommand(ctx, "addCustomMetadataField", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddCustomMetadataField: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetMetadataSchema returns available metadata fields from the project metadata schema.
func (e *Engine) GetMetadataSchema(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMetadataSchema", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMetadataSchema: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchSetMetadata sets a metadata field on multiple items at once.
func (e *Engine) BatchSetMetadata(ctx context.Context, itemIndices []int, field, value string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"itemIndices": itemIndices,
		"field": field,
		"value": value,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchSetMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchSetMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Labels & Colors Operations
// ---------------------------------------------------------------------------

// GetAvailableLabelColors returns all label colors (indices 0-15 with names).
func (e *Engine) GetAvailableLabelColors(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAvailableLabelColors", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAvailableLabelColors: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetClipLabelByName sets a label by color name on a project item.
func (e *Engine) SetClipLabelByName(ctx context.Context, projectItemIndex int, colorName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"colorName": colorName,
	})
	result, err := e.premiere.EvalCommand(ctx, "setClipLabelByName", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetClipLabelByName: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetLabelColorForClip returns the label color index and name for a project item.
func (e *Engine) GetLabelColorForClip(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getLabelColorForClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetLabelColorForClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BatchSetLabels sets a label on multiple items at once.
func (e *Engine) BatchSetLabels(ctx context.Context, itemIndices []int, colorIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"itemIndices": itemIndices,
		"colorIndex": colorIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "batchSetLabels", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BatchSetLabels: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// FilterByLabel returns all items with a specific label color.
func (e *Engine) FilterByLabel(ctx context.Context, colorIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"colorIndex": colorIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "filterByLabel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FilterByLabel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Footage Interpretation Operations
// ---------------------------------------------------------------------------

// GetFootageInterpretation returns interpretation settings for a project item.
func (e *Engine) GetFootageInterpretation(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getFootageInterpretation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFootageInterpretation: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetFootageFrameRate overrides frame rate on a project item.
func (e *Engine) SetFootageFrameRate(ctx context.Context, projectItemIndex int, fps float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"fps": fps,
	})
	result, err := e.premiere.EvalCommand(ctx, "setFootageFrameRate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetFootageFrameRate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetFootageFieldOrder sets field order on a project item.
func (e *Engine) SetFootageFieldOrder(ctx context.Context, projectItemIndex int, fieldOrder int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"fieldOrder": fieldOrder,
	})
	result, err := e.premiere.EvalCommand(ctx, "setFootageFieldOrder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetFootageFieldOrder: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetFootageAlphaChannel sets alpha interpretation on a project item.
func (e *Engine) SetFootageAlphaChannel(ctx context.Context, projectItemIndex int, alphaType int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"alphaType": alphaType,
	})
	result, err := e.premiere.EvalCommand(ctx, "setFootageAlphaChannel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetFootageAlphaChannel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetFootagePixelAspectRatio sets pixel aspect ratio on a project item.
func (e *Engine) SetFootagePixelAspectRatio(ctx context.Context, projectItemIndex int, num, den float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"num": num,
		"den": den,
	})
	result, err := e.premiere.EvalCommand(ctx, "setFootagePixelAspectRatio", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetFootagePixelAspectRatio: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ResetFootageInterpretation resets footage interpretation to auto-detected defaults.
func (e *Engine) ResetFootageInterpretation(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "resetFootageInterpretation", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResetFootageInterpretation: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Media Info Operations
// ---------------------------------------------------------------------------

// GetMediaInfo returns full media info for a project item.
func (e *Engine) GetMediaInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getMediaInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetMediaPath returns the file path for a project item's media.
func (e *Engine) GetMediaPath(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getMediaPath", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaPath: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RevealInFinder reveals a project item's media file in Finder/Explorer.
func (e *Engine) RevealInFinder(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "revealInFinder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RevealInFinder: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RefreshMedia forces a refresh of a project item's media.
func (e *Engine) RefreshMedia(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "refreshMedia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RefreshMedia: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ReplaceMedia replaces a project item's media with a different file.
func (e *Engine) ReplaceMedia(ctx context.Context, projectItemIndex int, newFilePath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"newFilePath": newFilePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "replaceMedia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReplaceMedia: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// DuplicateProjectItem duplicates a project item in the project panel.
func (e *Engine) DuplicateProjectItem(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "duplicateProjectItem", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DuplicateProjectItem: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Smart Bins Operations
// ---------------------------------------------------------------------------

// CreateSmartBin creates a smart bin with search criteria.
func (e *Engine) CreateSmartBin(ctx context.Context, name, searchQuery string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"searchQuery": searchQuery,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSmartBin", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSmartBin: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetSmartBinResults returns items matching smart bin criteria.
func (e *Engine) GetSmartBinResults(ctx context.Context, binPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"binPath": binPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSmartBinResults", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSmartBinResults: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip Usage Operations
// ---------------------------------------------------------------------------

// GetClipUsageInSequences finds all sequences where a clip is used.
func (e *Engine) GetClipUsageInSequences(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipUsageInSequences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipUsageInSequences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetUnusedClips lists clips not used in any sequence.
func (e *Engine) GetUnusedClips(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getUnusedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUnusedClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetUsedClips lists clips used in at least one sequence.
func (e *Engine) GetUsedClips(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getUsedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUsedClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipUsageCount counts how many times a clip is used across all sequences.
func (e *Engine) GetClipUsageCount(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipUsageCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipUsageCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// File Management Operations
// ---------------------------------------------------------------------------

// GetProjectFileSize returns the .prproj file size.
func (e *Engine) GetProjectFileSize(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectFileSize", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectFileSize: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetMediaDiskUsage calculates total disk usage of all media in the project.
func (e *Engine) GetMediaDiskUsage(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMediaDiskUsage", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaDiskUsage: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
