package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

)

// ---------------------------------------------------------------------------
// Event Monitoring (1-5)
// ---------------------------------------------------------------------------

// RegisterEventListener registers for a Premiere Pro event by name.
func (e *Engine) RegisterEventListener(ctx context.Context, eventName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"eventName": eventName,
	})
	result, err := e.premiere.EvalCommand(ctx, "registerEventListener", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RegisterEventListener: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// UnregisterEventListener unregisters a previously registered event listener.
func (e *Engine) UnregisterEventListener(ctx context.Context, eventName string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"eventName": eventName,
	})
	result, err := e.premiere.EvalCommand(ctx, "unregisterEventListener", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("UnregisterEventListener: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetRegisteredEvents lists all active event registrations.
func (e *Engine) GetRegisteredEvents(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getRegisteredEvents", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetRegisteredEvents: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetEventHistory returns the last N events that fired.
func (e *Engine) GetEventHistory(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "getEventHistory", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetEventHistory: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ClearEventHistory clears the event history buffer.
func (e *Engine) ClearEventHistory(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "clearEventHistory", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ClearEventHistory: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// State Watching (6-10)
// ---------------------------------------------------------------------------

// WatchPlayheadPosition starts polling the playhead position at the given interval.
func (e *Engine) WatchPlayheadPosition(ctx context.Context, intervalMs int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"intervalMs": intervalMs,
	})
	result, err := e.premiere.EvalCommand(ctx, "watchPlayheadPosition", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WatchPlayheadPosition: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// StopWatchPlayhead stops the playhead position watcher.
func (e *Engine) StopWatchPlayhead(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "stopWatchPlayhead", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StopWatchPlayhead: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// WatchRenderProgress starts watching render progress at the given interval.
func (e *Engine) WatchRenderProgress(ctx context.Context, intervalMs int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"intervalMs": intervalMs,
	})
	result, err := e.premiere.EvalCommand(ctx, "watchRenderProgress", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("WatchRenderProgress: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// StopWatchRender stops the render progress watcher.
func (e *Engine) StopWatchRender(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "stopWatchRender", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("StopWatchRender: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetStateSnapshot returns a complete state snapshot of the project, sequence,
// playhead, and current selection.
func (e *Engine) GetStateSnapshot(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getStateSnapshot", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetStateSnapshot: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Project State (11-14)
// ---------------------------------------------------------------------------

// IsProjectModified checks if the current project has unsaved changes.
func (e *Engine) IsProjectModified(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "isProjectModified", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsProjectModified: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetProjectDuration returns the total duration across all sequences in the project.
func (e *Engine) GetProjectDuration(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectDuration", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectDuration: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetProjectStats returns project statistics (clips, sequences, bins, effects).
func (e *Engine) GetProjectStats(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getProjectStats", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetProjectStats: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetRecentActions returns recent user actions from the event history.
func (e *Engine) GetRecentActions(ctx context.Context, count int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"count": count,
	})
	result, err := e.premiere.EvalCommand(ctx, "getRecentActions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetRecentActions: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Sequence State - Extended (15-20)
// ---------------------------------------------------------------------------

// GetActiveTrackTargets returns which tracks are targeted for insert/overwrite.
func (e *Engine) GetActiveTrackTargets(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getActiveTrackTargets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetActiveTrackTargets: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetActiveTrackTargets sets which tracks are targeted for insert/overwrite.
func (e *Engine) SetActiveTrackTargets(ctx context.Context, videoTargets, audioTargets string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"videoTargets": videoTargets,
		"audioTargets": audioTargets,
	})
	result, err := e.premiere.EvalCommand(ctx, "setActiveTrackTargets", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetActiveTrackTargets: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetTrackHeights returns the height/mute state of all tracks.
func (e *Engine) GetTrackHeights(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getTrackHeights", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetTrackHeights: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// SetTrackHeights sets track heights/mute state for the given track type.
func (e *Engine) SetTrackHeights(ctx context.Context, trackType, heights string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"heights": heights,
	})
	result, err := e.premiere.EvalCommand(ctx, "setTrackHeights", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SetTrackHeights: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// IsSequenceModified checks if the active sequence has unsaved changes.
func (e *Engine) IsSequenceModified(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "isSequenceModified", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsSequenceModified: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetSequenceHash returns a hash fingerprint of the sequence state for change detection.
func (e *Engine) GetSequenceHash(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceHash", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceHash: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Clip State (21-25)
// ---------------------------------------------------------------------------

// GetClipUnderPlayhead returns clip information at the current playhead position.
func (e *Engine) GetClipUnderPlayhead(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "getClipUnderPlayhead", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipUnderPlayhead: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipAtTime returns the clip at a specific time on a specific track.
func (e *Engine) GetClipAtTime(ctx context.Context, trackType string, trackIndex int, seconds float64) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"seconds": seconds,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipAtTime", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipAtTime: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetAdjacentClips returns the previous and next clips relative to a specified clip.
func (e *Engine) GetAdjacentClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getAdjacentClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetAdjacentClips: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// IsClipSelected checks if a specific clip is currently selected.
func (e *Engine) IsClipSelected(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "isClipSelected", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("IsClipSelected: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetClipProperties returns all properties of a clip as a JSON-compatible result.
func (e *Engine) GetClipProperties(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"trackType": trackType,
		"trackIndex": trackIndex,
		"clipIndex": clipIndex,
	})
	result, err := e.premiere.EvalCommand(ctx, "getClipProperties", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetClipProperties: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ---------------------------------------------------------------------------
// Notifications (26-30)
// ---------------------------------------------------------------------------

// ShowNotification displays a notification in the Premiere Pro Events panel.
func (e *Engine) ShowNotification(ctx context.Context, title, message string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"message": message,
	})
	result, err := e.premiere.EvalCommand(ctx, "showNotification", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowNotification: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// LogToEventsPanel logs a message to the Events panel at the given level.
func (e *Engine) LogToEventsPanel(ctx context.Context, message, level string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"message": message,
		"level": level,
	})
	result, err := e.premiere.EvalCommand(ctx, "logToEventsPanel", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("LogToEventsPanel: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ShowProgressBar displays a progress bar notification in the Events panel.
func (e *Engine) ShowProgressBar(ctx context.Context, title string, current, total int) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"current": current,
		"total": total,
	})
	result, err := e.premiere.EvalCommand(ctx, "showProgressBar", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowProgressBar: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// HideProgressBar hides the progress bar and logs completion.
func (e *Engine) HideProgressBar(ctx context.Context) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "hideProgressBar", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("HideProgressBar: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ShowDialog displays a dialog with custom buttons in Premiere Pro.
func (e *Engine) ShowDialog(ctx context.Context, title, message, buttons string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"title": title,
		"message": message,
		"buttons": buttons,
	})
	result, err := e.premiere.EvalCommand(ctx, "showDialog", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ShowDialog: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}
