package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Encoding Settings
// ---------------------------------------------------------------------------

// GetExportSettingsForPreset returns detailed encoding settings from a preset file.
func (e *Engine) GetExportSettingsForPreset(ctx context.Context, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getExportSettingsForPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExportSettingsForPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateCustomExportSettings creates custom export settings from a JSON spec.
func (e *Engine) CreateCustomExportSettings(ctx context.Context, settingsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"settingsJSON": settingsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "createCustomExportSettings", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateCustomExportSettings: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAvailableCodecs lists all available video codecs.
func (e *Engine) GetAvailableCodecs(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAvailableCodecs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAvailableCodecs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAvailableAudioCodecs lists all available audio codecs.
func (e *Engine) GetAvailableAudioCodecs(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAvailableAudioCodecs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAvailableAudioCodecs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAvailableContainers lists available container formats.
func (e *Engine) GetAvailableContainers(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAvailableContainers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAvailableContainers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Format Conversion
// ---------------------------------------------------------------------------

// ConvertToProRes converts a project item to Apple ProRes.
func (e *Engine) ConvertToProRes(ctx context.Context, projectItemIndex int, variant, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"variant": variant,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "convertToProRes", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConvertToProRes: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ConvertToH264 converts a project item to H.264.
func (e *Engine) ConvertToH264(ctx context.Context, projectItemIndex int, outputPath string, bitrate int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"outputPath": outputPath,
		"bitrate": bitrate,
	})
	result, err := e.premiere.EvalCommand(ctx, "convertToH264", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConvertToH264: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ConvertToH265 converts a project item to H.265/HEVC.
func (e *Engine) ConvertToH265(ctx context.Context, projectItemIndex int, outputPath string, bitrate int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"outputPath": outputPath,
		"bitrate": bitrate,
	})
	result, err := e.premiere.EvalCommand(ctx, "convertToH265", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConvertToH265: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ConvertToDNxHR converts a project item to DNxHR.
func (e *Engine) ConvertToDNxHR(ctx context.Context, projectItemIndex int, outputPath, profile string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"outputPath": outputPath,
		"profile": profile,
	})
	result, err := e.premiere.EvalCommand(ctx, "convertToDNxHR", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConvertToDNxHR: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ConvertToGIF exports a sequence as an animated GIF.
func (e *Engine) ConvertToGIF(ctx context.Context, sequenceIndex int, outputPath string, width, fps int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"width": width,
		"fps": fps,
	})
	result, err := e.premiere.EvalCommand(ctx, "convertToGIF", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConvertToGIF: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Thumbnail/Preview Generation
// ---------------------------------------------------------------------------

// GenerateClipThumbnail generates a thumbnail image from a clip.
func (e *Engine) GenerateClipThumbnail(ctx context.Context, projectItemIndex int, timeOffset float64, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"timeOffset": timeOffset,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateClipThumbnail", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateClipThumbnail: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GenerateSequenceThumbnail generates a thumbnail image from a sequence.
func (e *Engine) GenerateSequenceThumbnail(ctx context.Context, sequenceIndex int, timeOffset float64, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"timeOffset": timeOffset,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateSequenceThumbnail", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateSequenceThumbnail: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GenerateContactSheet generates a contact sheet from a clip.
func (e *Engine) GenerateContactSheet(ctx context.Context, projectItemIndex int, outputPath string, cols, rows int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"outputPath": outputPath,
		"cols": cols,
		"rows": rows,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateContactSheet", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateContactSheet: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GenerateStoryboard generates storyboard frames from a sequence.
func (e *Engine) GenerateStoryboard(ctx context.Context, sequenceIndex int, outputPath string, interval float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"interval": interval,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateStoryboard", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateStoryboard: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Media Analysis
// ---------------------------------------------------------------------------

// AnalyzeMediaCodec returns detailed codec analysis for a project item.
func (e *Engine) AnalyzeMediaCodec(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "analyzeMediaCodec", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AnalyzeMediaCodec: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CompareMediaSpecs compares specs of two project items.
func (e *Engine) CompareMediaSpecs(ctx context.Context, itemIndex1, itemIndex2 int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"itemIndex1": itemIndex1,
		"itemIndex2": itemIndex2,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareMediaSpecs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareMediaSpecs: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetBitRateInfo returns bitrate information for a project item.
func (e *Engine) GetBitRateInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getBitRateInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetBitRateInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetColorDepthInfo returns color depth and subsampling info.
func (e *Engine) GetColorDepthInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getColorDepthInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetColorDepthInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAudioSpecsDetailed returns detailed audio specs for a project item.
func (e *Engine) GetAudioSpecsDetailed(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getAudioSpecsDetailed", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAudioSpecsDetailed: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// IsVariableFrameRate checks if a clip has variable frame rate.
func (e *Engine) IsVariableFrameRate(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "isVariableFrameRate", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsVariableFrameRate: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// File Operations
// ---------------------------------------------------------------------------

// GetFileHash returns a file hash/fingerprint for a project item's media file.
func (e *Engine) GetFileHash(ctx context.Context, projectItemIndex int, algorithm string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"algorithm": algorithm,
	})
	result, err := e.premiere.EvalCommand(ctx, "getFileHash", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFileHash: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetFileDates returns creation/modification dates for a media file.
func (e *Engine) GetFileDates(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getFileDates", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFileDates: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// MoveMediaFile moves a media file and relinks in the project.
func (e *Engine) MoveMediaFile(ctx context.Context, projectItemIndex int, newDirectory string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"newDirectory": newDirectory,
	})
	result, err := e.premiere.EvalCommand(ctx, "moveMediaFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MoveMediaFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CopyMediaFile copies a media file to a destination directory.
func (e *Engine) CopyMediaFile(ctx context.Context, projectItemIndex int, destDirectory string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"destDirectory": destDirectory,
	})
	result, err := e.premiere.EvalCommand(ctx, "copyMediaFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CopyMediaFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RenameMediaFile renames a media file on disk and relinks in the project.
func (e *Engine) RenameMediaFile(ctx context.Context, projectItemIndex int, newName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"newName": newName,
	})
	result, err := e.premiere.EvalCommand(ctx, "renameMediaFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RenameMediaFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Render Queue
// ---------------------------------------------------------------------------

// AddToRenderQueue adds a sequence to the internal render queue.
func (e *Engine) AddToRenderQueue(ctx context.Context, sequenceIndex int, presetPath, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"presetPath": presetPath,
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "addToRenderQueue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddToRenderQueue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetRenderQueueStatus returns the render queue status.
func (e *Engine) GetRenderQueueStatus(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getRenderQueueStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetRenderQueueStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClearRenderQueue clears the render queue.
func (e *Engine) ClearRenderQueue(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearRenderQueue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearRenderQueue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// PauseRenderQueue pauses rendering.
func (e *Engine) PauseRenderQueue(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "pauseRenderQueue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PauseRenderQueue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ResumeRenderQueue resumes rendering.
func (e *Engine) ResumeRenderQueue(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "resumeRenderQueue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ResumeRenderQueue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
