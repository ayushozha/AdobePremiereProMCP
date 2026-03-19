package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Compound / Multi-Step Editing Operations
// ---------------------------------------------------------------------------

// --- Multi-Step Editing ---

func (e *Engine) CreateMontage(ctx context.Context, clipIndices []int, transitionName string, transitionDuration float64, musicPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"clipIndices": clipIndices,
		"transitionName": transitionName,
		"transitionDuration": transitionDuration,
		"musicPath": musicPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createMontage", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateMontage: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CreateSlideshow(ctx context.Context, imageIndices []int, slideDuration float64, transitionName string, musicPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"imageIndices": imageIndices,
		"slideDuration": slideDuration,
		"transitionName": transitionName,
		"musicPath": musicPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSlideshow", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSlideshow: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CreateHighlightReel(ctx context.Context, sequenceIndex int, markerColor string, outputName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"markerColor": markerColor,
		"outputName": outputName,
	})
	result, err := e.premiere.EvalCommand(ctx, "createHighlightReel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateHighlightReel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) RippleDeleteEmptySpaces(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "rippleDeleteEmptySpaces", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RippleDeleteEmptySpaces: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AlignAllClipsToTrack(ctx context.Context, sourceTrack, destTrack int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sourceTrack": sourceTrack,
		"destTrack": destTrack,
	})
	result, err := e.premiere.EvalCommand(ctx, "alignAllClipsToTrack", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AlignAllClipsToTrack: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Audio-Visual Sync ---

func (e *Engine) SyncAllAudioToVideo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "syncAllAudioToVideo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SyncAllAudioToVideo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ReplaceAudio(ctx context.Context, videoTrackIndex, videoClipIndex int, audioPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"videoTrackIndex": videoTrackIndex,
		"videoClipIndex": videoClipIndex,
		"audioPath": audioPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "replaceAudio", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ReplaceAudio: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AddMusicBed(ctx context.Context, audioPath string, trackIndex int, startTime, endTime, fadeIn, fadeOut, volume float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"audioPath": audioPath,
		"trackIndex": trackIndex,
		"startTime": startTime,
		"endTime": endTime,
		"fadeIn": fadeIn,
		"fadeOut": fadeOut,
		"volume": volume,
	})
	result, err := e.premiere.EvalCommand(ctx, "addMusicBed", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddMusicBed: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) DuckMusicUnderDialogue(ctx context.Context, musicTrackIndex, dialogueTrackIndex int, duckAmount float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"musicTrackIndex": musicTrackIndex,
		"dialogueTrackIndex": dialogueTrackIndex,
		"duckAmount": duckAmount,
	})
	result, err := e.premiere.EvalCommand(ctx, "duckMusicUnderDialogue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DuckMusicUnderDialogue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AddSoundEffect(ctx context.Context, sfxPath string, trackIndex int, time, volume float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sfxPath": sfxPath,
		"trackIndex": trackIndex,
		"time": time,
		"volume": volume,
	})
	result, err := e.premiere.EvalCommand(ctx, "addSoundEffect", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddSoundEffect: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Color Workflow ---

func (e *Engine) MatchColorBetweenClips(ctx context.Context, srcTrackIndex, srcClipIndex, destTrackIndex, destClipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"srcTrackIndex": srcTrackIndex,
		"srcClipIndex": srcClipIndex,
		"destTrackIndex": destTrackIndex,
		"destClipIndex": destClipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "matchColorBetweenClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MatchColorBetweenClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ApplyColorPreset(ctx context.Context, trackIndex, clipIndex int, presetName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"presetName": presetName,
	})
	result, err := e.premiere.EvalCommand(ctx, "applyColorPreset", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ApplyColorPreset: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) CreateColorGradient(ctx context.Context, trackIndex, startClipIndex, endClipIndex int, startTemp, endTemp float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"startClipIndex": startClipIndex,
		"endClipIndex": endClipIndex,
		"startTemp": startTemp,
		"endTemp": endTemp,
	})
	result, err := e.premiere.EvalCommand(ctx, "createColorGradient", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateColorGradient: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AutoCorrectAllClips(ctx context.Context, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "autoCorrectAllClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AutoCorrectAllClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Text Workflow ---

func (e *Engine) AddSubtitlesFromSRT(ctx context.Context, srtPath string, trackIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"srtPath": srtPath,
		"trackIndex": trackIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "addSubtitlesFromSRT", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddSubtitlesFromSRT: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AddEndCredits(ctx context.Context, creditsJSON string, trackIndex int, scrollDuration float64, style string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"creditsJSON": creditsJSON,
		"trackIndex": trackIndex,
		"scrollDuration": scrollDuration,
		"style": style,
	})
	result, err := e.premiere.EvalCommand(ctx, "addEndCredits", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddEndCredits: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) AddChapterMarkers(ctx context.Context, chaptersJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"chaptersJSON": chaptersJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "addChapterMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddChapterMarkers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) GenerateChaptersFromMarkers(ctx context.Context, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateChaptersFromMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateChaptersFromMarkers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Export Workflow ---

func (e *Engine) ExportForYouTube(ctx context.Context, outputPath, title, description string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"title": title,
		"description": description,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForYouTube", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForYouTube: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExportForInstagram(ctx context.Context, outputPath, aspectRatio string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"aspectRatio": aspectRatio,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForInstagram", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForInstagram: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExportForTikTok(ctx context.Context, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForTikTok", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForTikTok: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExportForTwitter(ctx context.Context, outputPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportForTwitter", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportForTwitter: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ExportMultipleFormats(ctx context.Context, outputDir string, formats []string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputDir": outputDir,
		"formats": formats,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportMultipleFormats", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportMultipleFormats: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Project Setup ---

func (e *Engine) SetupNewProject(ctx context.Context, name, path, resolution string, fps float64, audioSampleRate int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"name": name,
		"path": path,
		"resolution": resolution,
		"fps": fps,
		"audioSampleRate": audioSampleRate,
	})
	result, err := e.premiere.EvalCommand(ctx, "setupNewProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetupNewProject: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) SetupEditingWorkspace(ctx context.Context, projectPath, mediaFolder, sequenceName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectPath": projectPath,
		"mediaFolder": mediaFolder,
		"sequenceName": sequenceName,
	})
	result, err := e.premiere.EvalCommand(ctx, "setupEditingWorkspace", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetupEditingWorkspace: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ImportAndOrganize(ctx context.Context, mediaFolder string, autoCreateBins bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"mediaFolder": mediaFolder,
		"autoCreateBins": autoCreateBins,
	})
	result, err := e.premiere.EvalCommand(ctx, "importAndOrganize", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportAndOrganize: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) PrepareForDelivery(ctx context.Context, specsJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"specsJSON": specsJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "prepareForDelivery", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("PrepareForDelivery: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// --- Cleanup Workflow ---

func (e *Engine) ArchiveProject(ctx context.Context, outputPath string, includeMedia, includeRenders bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"includeMedia": includeMedia,
		"includeRenders": includeRenders,
	})
	result, err := e.premiere.EvalCommand(ctx, "archiveProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ArchiveProject: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) TrimProject(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "trimProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TrimProject: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

func (e *Engine) ConsolidateAndTranscode(ctx context.Context, outputDir, codec, quality string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputDir": outputDir,
		"codec": codec,
		"quality": quality,
	})
	result, err := e.premiere.EvalCommand(ctx, "consolidateAndTranscode", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConsolidateAndTranscode: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
