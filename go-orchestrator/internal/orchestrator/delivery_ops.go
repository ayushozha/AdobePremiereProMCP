package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Social Media Optimization (1-5)
// ---------------------------------------------------------------------------

// CreateVerticalVersion creates a 9:16 vertical version of a 16:9 sequence.
func (e *Engine) CreateVerticalVersion(ctx context.Context, sequenceIndex int, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "createVerticalVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateVerticalVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateSquareVersion creates a 1:1 square version of a sequence.
func (e *Engine) CreateSquareVersion(ctx context.Context, sequenceIndex int, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSquareVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSquareVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddSafeZoneGuides adds safe zone guides for a specific platform.
func (e *Engine) AddSafeZoneGuides(ctx context.Context, sequenceIndex int, platform string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"platform": platform,
	})
	result, err := e.premiere.EvalCommand(ctx, "addSafeZoneGuides", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddSafeZoneGuides: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// OptimizeForPlatform auto-optimizes sequence settings for a platform.
func (e *Engine) OptimizeForPlatform(ctx context.Context, sequenceIndex int, platform string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"platform": platform,
	})
	result, err := e.premiere.EvalCommand(ctx, "optimizeForPlatform", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OptimizeForPlatform: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateThumbnailFromFrame creates a thumbnail from a specific frame.
func (e *Engine) CreateThumbnailFromFrame(ctx context.Context, sequenceIndex int, timeSeconds float64, outputPath, addText string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"timeSeconds": timeSeconds,
		"outputPath": outputPath,
		"addText": addText,
	})
	result, err := e.premiere.EvalCommand(ctx, "createThumbnailFromFrame", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateThumbnailFromFrame: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Content Segmentation (6-10)
// ---------------------------------------------------------------------------

// SplitIntoSegments splits a long video into segments under a max duration.
func (e *Engine) SplitIntoSegments(ctx context.Context, sequenceIndex int, maxDurationSeconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"maxDurationSeconds": maxDurationSeconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "splitIntoSegments", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SplitIntoSegments: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateChaptersFile creates a chapters file from sequence markers.
func (e *Engine) CreateChaptersFile(ctx context.Context, sequenceIndex int, outputPath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "createChaptersFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateChaptersFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExtractSegmentByMarkers extracts a segment between two markers.
func (e *Engine) ExtractSegmentByMarkers(ctx context.Context, sequenceIndex, startMarkerIndex, endMarkerIndex int, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"startMarkerIndex": startMarkerIndex,
		"endMarkerIndex": endMarkerIndex,
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "extractSegmentByMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExtractSegmentByMarkers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateTeaser auto-creates a short teaser from a sequence.
func (e *Engine) CreateTeaser(ctx context.Context, sequenceIndex int, durationSeconds float64, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"durationSeconds": durationSeconds,
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "createTeaser", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateTeaser: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateBumper creates an intro/outro bumper sequence.
func (e *Engine) CreateBumper(ctx context.Context, text string, duration float64, style, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"text": text,
		"duration": duration,
		"style": style,
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "createBumper", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateBumper: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Delivery Formats (11-15)
// ---------------------------------------------------------------------------

// ExportForBroadcast exports for broadcast standards (ATSC, DVB, ISDB).
func (e *Engine) ExportForBroadcast(ctx context.Context, sequenceIndex int, outputPath, standard string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"standard": standard,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForBroadcast", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForBroadcast: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportForStreaming exports for streaming platforms (Netflix, Amazon, Disney+).
func (e *Engine) ExportForStreaming(ctx context.Context, sequenceIndex int, outputPath, platform string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"platform": platform,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForStreaming", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForStreaming: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportForArchive exports for archival (lossless, ProRes 4444, DNxHR 444).
func (e *Engine) ExportForArchive(ctx context.Context, sequenceIndex int, outputPath, codec string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"codec": codec,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForArchive", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForArchive: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportForWeb exports for web with adaptive bitrate settings.
func (e *Engine) ExportForWeb(ctx context.Context, sequenceIndex int, outputPath, quality string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"quality": quality,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForWeb", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForWeb: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportForMobile exports optimized for mobile devices.
func (e *Engine) ExportForMobile(ctx context.Context, sequenceIndex int, outputPath, device string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputPath": outputPath,
		"device": device,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForMobile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForMobile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Metadata for Distribution (16-20)
// ---------------------------------------------------------------------------

// SetDistributionMetadata sets distribution metadata on a sequence.
func (e *Engine) SetDistributionMetadata(ctx context.Context, sequenceIndex int, title, description, tags, category string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"title": title,
		"description": description,
		"tags": tags,
		"category": category,
	})
	result, err := e.premiere.EvalCommand(ctx, "setDistributionMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetDistributionMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetDistributionMetadata gets distribution metadata from a sequence.
func (e *Engine) GetDistributionMetadata(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getDistributionMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDistributionMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// EmbedThumbnailInFile embeds a thumbnail image in a video file.
func (e *Engine) EmbedThumbnailInFile(ctx context.Context, videoPath, thumbnailPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"videoPath": videoPath,
		"thumbnailPath": thumbnailPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "embedThumbnailInFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EmbedThumbnailInFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// AddChapterMetadata adds chapter metadata to an exported video file.
func (e *Engine) AddChapterMetadata(ctx context.Context, videoPath, chaptersJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"videoPath": videoPath,
		"chaptersJSON": chaptersJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "addChapterMetadata", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddChapterMetadata: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetContentRating sets content rating metadata on a sequence.
func (e *Engine) SetContentRating(ctx context.Context, sequenceIndex int, rating string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"rating": rating,
	})
	result, err := e.premiere.EvalCommand(ctx, "setContentRating", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetContentRating: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Quality Assurance (21-25)
// ---------------------------------------------------------------------------

// RunQAChecklist runs a QA checklist against specs.
func (e *Engine) RunQAChecklist(ctx context.Context, sequenceIndex int, specsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"specsJSON": specsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "runQAChecklist", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RunQAChecklist: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckLoudnessCompliance checks loudness compliance against a standard.
func (e *Engine) CheckLoudnessCompliance(ctx context.Context, sequenceIndex int, standard string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"standard": standard,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkLoudnessCompliance", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckLoudnessCompliance: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckColorCompliance checks color compliance against a standard.
func (e *Engine) CheckColorCompliance(ctx context.Context, sequenceIndex int, standard string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"standard": standard,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkColorCompliance", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckColorCompliance: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckFrameAccuracy checks for frame-accurate edits.
func (e *Engine) CheckFrameAccuracy(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkFrameAccuracy", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckFrameAccuracy: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ValidateClosedCaptions validates closed captions for FCC compliance.
func (e *Engine) ValidateClosedCaptions(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "validateClosedCaptions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ValidateClosedCaptions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Versioning (26-30)
// ---------------------------------------------------------------------------

// CreateVersionedExport exports with version tracking.
func (e *Engine) CreateVersionedExport(ctx context.Context, sequenceIndex int, outputDir, versionName, notes string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputDir": outputDir,
		"versionName": versionName,
		"notes": notes,
	})
	result, err := e.premiere.EvalCommand(ctx, "createVersionedExport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateVersionedExport: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetExportHistory2 gets export history for a sequence.
func (e *Engine) GetExportHistory2(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getExportHistory2", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetExportHistory2: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CompareExportVersions compares two export versions.
func (e *Engine) CompareExportVersions(ctx context.Context, version1Path, version2Path string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"version1Path": version1Path,
		"version2Path": version2Path,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareExportVersions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareExportVersions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CreateApprovalPackage creates a package for client approval.
func (e *Engine) CreateApprovalPackage(ctx context.Context, sequenceIndex int, outputDir string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"outputDir": outputDir,
	})
	result, err := e.premiere.EvalCommand(ctx, "createApprovalPackage", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateApprovalPackage: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ArchiveAndCleanup archives the project and cleans up.
func (e *Engine) ArchiveAndCleanup(ctx context.Context, sequenceIndex int, archiveDir string, deleteRenders bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"archiveDir": archiveDir,
		"deleteRenders": deleteRenders,
	})
	result, err := e.premiere.EvalCommand(ctx, "archiveAndCleanup", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ArchiveAndCleanup: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
