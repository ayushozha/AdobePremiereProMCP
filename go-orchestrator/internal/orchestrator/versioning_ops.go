package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Timeline Diff
// ---------------------------------------------------------------------------

// SnapshotTimeline creates a JSON snapshot of a timeline state.
func (e *Engine) SnapshotTimeline(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "snapshotTimeline", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SnapshotTimeline: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CompareTimelineSnapshots diffs two timeline snapshots.
func (e *Engine) CompareTimelineSnapshots(ctx context.Context, snapshot1JSON, snapshot2JSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"snapshot1JSON": snapshot1JSON,
		"snapshot2JSON": snapshot2JSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "compareTimelineSnapshots", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CompareTimelineSnapshots: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTimelineChanges returns changes to a timeline since a given timestamp.
func (e *Engine) GetTimelineChanges(ctx context.Context, sequenceIndex int, sinceTimestamp string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"sinceTimestamp": sinceTimestamp,
	})
	result, err := e.premiere.EvalCommand(ctx, "getTimelineChanges", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTimelineChanges: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// HighlightChangedClips selects/highlights clips that changed.
func (e *Engine) HighlightChangedClips(ctx context.Context, sequenceIndex int, changedClipIDs string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"changedClipIDs": changedClipIDs,
	})
	result, err := e.premiere.EvalCommand(ctx, "highlightChangedClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("HighlightChangedClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RevertClipToSnapshot reverts a clip to a previous snapshot state.
func (e *Engine) RevertClipToSnapshot(ctx context.Context, trackType string, trackIndex, clipIndex int, snapshotJSON string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
		"snapshotJSON": snapshotJSON,
	})
	result, err := e.premiere.EvalCommand(ctx, "revertClipToSnapshot", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RevertClipToSnapshot: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Sequence Versioning
// ---------------------------------------------------------------------------

// SaveSequenceVersion saves a named version of a sequence.
func (e *Engine) SaveSequenceVersion(ctx context.Context, sequenceIndex int, versionName, notes string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"versionName": versionName,
		"notes": notes,
	})
	result, err := e.premiere.EvalCommand(ctx, "saveSequenceVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SaveSequenceVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ListSequenceVersions lists all saved versions for a sequence.
func (e *Engine) ListSequenceVersions(ctx context.Context, sequenceIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "listSequenceVersions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ListSequenceVersions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LoadSequenceVersion loads a saved version of a sequence.
func (e *Engine) LoadSequenceVersion(ctx context.Context, sequenceIndex int, versionName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"versionName": versionName,
	})
	result, err := e.premiere.EvalCommand(ctx, "loadSequenceVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LoadSequenceVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// DeleteSequenceVersion deletes a saved version.
func (e *Engine) DeleteSequenceVersion(ctx context.Context, sequenceIndex int, versionName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceIndex": sequenceIndex,
		"versionName": versionName,
	})
	result, err := e.premiere.EvalCommand(ctx, "deleteSequenceVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DeleteSequenceVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// MergeSequenceVersions merges two sequence versions using a strategy.
func (e *Engine) MergeSequenceVersions(ctx context.Context, baseVersion, overlayVersion, strategy string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"baseVersion": baseVersion,
		"overlayVersion": overlayVersion,
		"strategy": strategy,
	})
	result, err := e.premiere.EvalCommand(ctx, "mergeSequenceVersions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("MergeSequenceVersions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// A/B Comparison
// ---------------------------------------------------------------------------

// CreateABComparison sets up an A/B comparison between two sequences.
func (e *Engine) CreateABComparison(ctx context.Context, seqIndexA, seqIndexB int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seqIndexA": seqIndexA,
		"seqIndexB": seqIndexB,
	})
	result, err := e.premiere.EvalCommand(ctx, "createABComparison", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateABComparison: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SwitchABView switches between A, B, or split view.
func (e *Engine) SwitchABView(ctx context.Context, view string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"view": view,
	})
	result, err := e.premiere.EvalCommand(ctx, "switchABView", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SwitchABView: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetABDifferences lists differences between two sequences.
func (e *Engine) GetABDifferences(ctx context.Context, seqIndexA, seqIndexB int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"seqIndexA": seqIndexA,
		"seqIndexB": seqIndexB,
	})
	result, err := e.premiere.EvalCommand(ctx, "getABDifferences", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetABDifferences: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clipboard Extended
// ---------------------------------------------------------------------------

// GetClipboardContents returns information about current clipboard contents.
func (e *Engine) GetClipboardContents(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getClipboardContents", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipboardContents: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClearClipboard clears the editing clipboard.
func (e *Engine) ClearClipboard(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearClipboard", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearClipboard: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClipboardHasContent checks whether the clipboard has content.
func (e *Engine) ClipboardHasContent(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clipboardHasContent", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClipboardHasContent: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// History
// ---------------------------------------------------------------------------

// GetUndoHistory returns recent undo history entries.
func (e *Engine) GetUndoHistory(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "getUndoHistory", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUndoHistory: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetUndoCount returns the number of available undo steps.
func (e *Engine) GetUndoCount(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getUndoCount", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetUndoCount: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// UndoMultiple undoes multiple steps.
func (e *Engine) UndoMultiple(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "undoMultiple", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("UndoMultiple: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RedoMultiple redoes multiple steps.
func (e *Engine) RedoMultiple(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "redoMultiple", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RedoMultiple: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Backup
// ---------------------------------------------------------------------------

// CreateProjectBackup creates a full project backup.
func (e *Engine) CreateProjectBackup(ctx context.Context, outputPath string, includeMedia bool) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"includeMedia": includeMedia,
	})
	result, err := e.premiere.EvalCommand(ctx, "createProjectBackup", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateProjectBackup: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAutoSaveVersions lists auto-save versions of the project.
func (e *Engine) GetAutoSaveVersions(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getAutoSaveVersions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAutoSaveVersions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// RestoreAutoSave restores from an auto-save version.
func (e *Engine) RestoreAutoSave(ctx context.Context, versionPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"versionPath": versionPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "restoreAutoSave", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RestoreAutoSave: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetAutoSaveNow triggers an immediate auto-save.
func (e *Engine) SetAutoSaveNow(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "setAutoSaveNow", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetAutoSaveNow: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetBackupSchedule returns backup schedule information.
func (e *Engine) GetBackupSchedule(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getBackupSchedule", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetBackupSchedule: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project Migration
// ---------------------------------------------------------------------------

// UpgradeProjectVersion upgrades an older project file.
func (e *Engine) UpgradeProjectVersion(ctx context.Context, projectPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectPath": projectPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "upgradeProjectVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("UpgradeProjectVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetProjectVersion returns the project file version.
func (e *Engine) GetProjectVersion(ctx context.Context, projectPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectPath": projectPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "getProjectVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ExportProjectForOlderVersion saves the project for an older Premiere Pro version.
func (e *Engine) ExportProjectForOlderVersion(ctx context.Context, outputPath, targetVersion string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"outputPath": outputPath,
		"targetVersion": targetVersion,
	})
	result, err := e.premiere.EvalCommand(ctx, "exportProjectForOlderVersion", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ExportProjectForOlderVersion: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// CheckProjectCompatibility checks a project's compatibility with the current version.
func (e *Engine) CheckProjectCompatibility(ctx context.Context, projectPath string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"projectPath": projectPath,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkProjectCompatibility", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckProjectCompatibility: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportProjectFromOtherNLE imports a project from another NLE (DaVinci, FCPX, Avid).
func (e *Engine) ImportProjectFromOtherNLE(ctx context.Context, sourcePath, sourceFormat string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"sourcePath": sourcePath,
		"sourceFormat": sourceFormat,
	})
	result, err := e.premiere.EvalCommand(ctx, "importProjectFromOtherNLE", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportProjectFromOtherNLE: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
