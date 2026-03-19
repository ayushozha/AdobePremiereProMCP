package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// After Effects Integration
// ---------------------------------------------------------------------------

// SendToAfterEffects replaces a project item with a Dynamic Link AE composition.
func (e *Engine) SendToAfterEffects(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "sendToAfterEffects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SendToAfterEffects: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportAEComp imports an After Effects composition from a .aep file via Dynamic Link.
func (e *Engine) ImportAEComp(ctx context.Context, aepPath, compName, targetBin string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"aepPath": aepPath,
		"compName": compName,
		"targetBin": targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importAEComp", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportAEComp: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportAllAEComps imports all After Effects compositions from a .aep file.
func (e *Engine) ImportAllAEComps(ctx context.Context, aepPath, targetBin string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"aepPath": aepPath,
		"targetBin": targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importAllAEComps", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportAllAEComps: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RefreshAEComp refreshes a Dynamic Link After Effects composition.
func (e *Engine) RefreshAEComp(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "refreshAEComp", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RefreshAEComp: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Photoshop Integration
// ---------------------------------------------------------------------------

// EditInPhotoshop opens a project item (frame/image) in Adobe Photoshop.
func (e *Engine) EditInPhotoshop(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "editInPhotoshop", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EditInPhotoshop: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportPSDLayers imports a Photoshop PSD file with layer support.
func (e *Engine) ImportPSDLayers(ctx context.Context, psdPath, targetBin string, asSequence bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"psdPath": psdPath,
		"targetBin": targetBin,
		"asSequence": asSequence,
	})
	result, err := e.premiere.EvalCommand(ctx, "importPSDLayers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportPSDLayers: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Audition Integration
// ---------------------------------------------------------------------------

// EditInAudition sends an audio clip to Adobe Audition for editing.
func (e *Engine) EditInAudition(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "editInAudition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EditInAudition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RefreshAuditionEdit refreshes an audio clip after editing in Audition.
func (e *Engine) RefreshAuditionEdit(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "refreshAuditionEdit", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RefreshAuditionEdit: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Media Encoder Integration
// ---------------------------------------------------------------------------

// QueueInMediaEncoder queues a sequence in Adobe Media Encoder for batch encoding.
func (e *Engine) QueueInMediaEncoder(ctx context.Context, sequenceIndex int, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "queueInMediaEncoder", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("QueueInMediaEncoder: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetMediaEncoderQueue returns the current Adobe Media Encoder queue status.
func (e *Engine) GetMediaEncoderQueue(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getMediaEncoderQueue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetMediaEncoderQueue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClearMediaEncoderQueue clears the Adobe Media Encoder queue.
func (e *Engine) ClearMediaEncoderQueue(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearMediaEncoderQueue", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearMediaEncoderQueue: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Dynamic Link
// ---------------------------------------------------------------------------

// GetDynamicLinkStatus returns Dynamic Link connection status for linked compositions.
func (e *Engine) GetDynamicLinkStatus(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getDynamicLinkStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetDynamicLinkStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RefreshAllDynamicLinks refreshes all Dynamic Link clips to pull latest changes.
func (e *Engine) RefreshAllDynamicLinks(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "refreshAllDynamicLinks", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RefreshAllDynamicLinks: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// File Format Support / Codec
// ---------------------------------------------------------------------------

// GetCodecInfo returns detailed codec information for a project item.
func (e *Engine) GetCodecInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getCodecInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetCodecInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// TranscodeClip transcodes a project item clip using a specified preset.
func (e *Engine) TranscodeClip(ctx context.Context, projectItemIndex int, outputPath, presetPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"outputPath": outputPath,
		"presetPath": presetPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "transcodeClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("TranscodeClip: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ConformMedia conforms media to target specifications (frame rate, codec).
func (e *Engine) ConformMedia(ctx context.Context, projectItemIndex int, targetFps float64, targetCodec string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"targetFps": targetFps,
		"targetCodec": targetCodec,
	})
	result, err := e.premiere.EvalCommand(ctx, "conformMedia", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ConformMedia: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Interchange (OMF/AAF import)
// ---------------------------------------------------------------------------

// ImportOMFFile imports an OMF file into the project.
func (e *Engine) ImportOMFFile(ctx context.Context, omfPath, targetBin string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"omfPath": omfPath,
		"targetBin": targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importOMFFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportOMFFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportAAFFile imports an AAF file into the project.
func (e *Engine) ImportAAFFile(ctx context.Context, aafPath, targetBin string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"aafPath": aafPath,
		"targetBin": targetBin,
	})
	result, err := e.premiere.EvalCommand(ctx, "importAAFFile", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportAAFFile: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clipboard
// ---------------------------------------------------------------------------

// CopyToSystemClipboard copies text to the system clipboard.
func (e *Engine) CopyToSystemClipboard(ctx context.Context, text string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"text": text,
	})
	result, err := e.premiere.EvalCommand(ctx, "copyToSystemClipboard", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CopyToSystemClipboard: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetFromSystemClipboard reads text from the system clipboard.
func (e *Engine) GetFromSystemClipboard(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getFromSystemClipboard", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetFromSystemClipboard: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// External Tools
// ---------------------------------------------------------------------------

// OpenInExternalEditor opens a project item in an external editor application.
func (e *Engine) OpenInExternalEditor(ctx context.Context, projectItemIndex int, editorPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectItemIndex": projectItemIndex,
		"editorPath": editorPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "openInExternalEditor", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OpenInExternalEditor: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportFromExternalSource imports media from an external source/format.
func (e *Engine) ImportFromExternalSource(ctx context.Context, sourcePath, format string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sourcePath": sourcePath,
		"format": format,
	})
	result, err := e.premiere.EvalCommand(ctx, "importFromExternalSource", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportFromExternalSource: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Team Projects
// ---------------------------------------------------------------------------

// GetTeamProjectStatus returns the Team Projects connection status.
func (e *Engine) GetTeamProjectStatus(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTeamProjectStatus", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTeamProjectStatus: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckInChanges checks in (shares) changes to Team Projects.
func (e *Engine) CheckInChanges(ctx context.Context, message string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"message": message,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkInChanges", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckInChanges: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckOutSequence checks out a sequence for exclusive editing in Team Projects.
func (e *Engine) CheckOutSequence(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkOutSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckOutSequence: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Productions
// ---------------------------------------------------------------------------

// GetProductionInfo returns information about the current production.
func (e *Engine) GetProductionInfo(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProductionInfo", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProductionInfo: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ListProductionProjects lists all projects within the current production.
func (e *Engine) ListProductionProjects(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "listProductionProjects", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListProductionProjects: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// OpenProductionProject opens a specific project from within the current production.
func (e *Engine) OpenProductionProject(ctx context.Context, projectName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectName": projectName,
	})
	result, err := e.premiere.EvalCommand(ctx, "openProductionProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("OpenProductionProject: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
