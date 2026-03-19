package orchestrator

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Playback Control
// ---------------------------------------------------------------------------

func (e *Engine) Play(ctx context.Context, speed float64) (*GenericResult, error) {
	e.logger.Debug("play", zap.Float64("speed", speed))
	return nil, fmt.Errorf("play: not yet implemented in bridge")
}

func (e *Engine) Pause(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("pause")
	return nil, fmt.Errorf("pause: not yet implemented in bridge")
}

func (e *Engine) Stop(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("stop")
	return nil, fmt.Errorf("stop: not yet implemented in bridge")
}

func (e *Engine) StepForward(ctx context.Context, frames int) (*GenericResult, error) {
	e.logger.Debug("step_forward", zap.Int("frames", frames))
	return nil, fmt.Errorf("step forward: not yet implemented in bridge")
}

func (e *Engine) StepBackward(ctx context.Context, frames int) (*GenericResult, error) {
	e.logger.Debug("step_backward", zap.Int("frames", frames))
	return nil, fmt.Errorf("step backward: not yet implemented in bridge")
}

func (e *Engine) ShuttleForward(ctx context.Context, speed float64) (*GenericResult, error) {
	e.logger.Debug("shuttle_forward", zap.Float64("speed", speed))
	return nil, fmt.Errorf("shuttle forward: not yet implemented in bridge")
}

func (e *Engine) ShuttleBackward(ctx context.Context, speed float64) (*GenericResult, error) {
	e.logger.Debug("shuttle_backward", zap.Float64("speed", speed))
	return nil, fmt.Errorf("shuttle backward: not yet implemented in bridge")
}

func (e *Engine) TogglePlayPause(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("toggle_play_pause")
	return nil, fmt.Errorf("toggle play/pause: not yet implemented in bridge")
}

func (e *Engine) PlayInToOut(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("play_in_to_out")
	return nil, fmt.Errorf("play in to out: not yet implemented in bridge")
}

func (e *Engine) LoopPlayback(ctx context.Context, enabled bool) (*GenericResult, error) {
	e.logger.Debug("loop_playback", zap.Bool("enabled", enabled))
	return nil, fmt.Errorf("loop playback: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Program Monitor
// ---------------------------------------------------------------------------

func (e *Engine) GetProgramMonitorZoom(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_program_monitor_zoom")
	return nil, fmt.Errorf("get program monitor zoom: not yet implemented in bridge")
}

func (e *Engine) SetProgramMonitorZoom(ctx context.Context, percent float64) (*GenericResult, error) {
	e.logger.Debug("set_program_monitor_zoom", zap.Float64("percent", percent))
	return nil, fmt.Errorf("set program monitor zoom: not yet implemented in bridge")
}

func (e *Engine) FitProgramMonitor(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("fit_program_monitor")
	return nil, fmt.Errorf("fit program monitor: not yet implemented in bridge")
}

func (e *Engine) ToggleSafeMargins(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("toggle_safe_margins")
	return nil, fmt.Errorf("toggle safe margins: not yet implemented in bridge")
}

func (e *Engine) GetFrameAtPlayhead(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_frame_at_playhead")
	return nil, fmt.Errorf("get frame at playhead: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Sequence Navigation (extended)
// ---------------------------------------------------------------------------

func (e *Engine) GoToTimecode(ctx context.Context, timecode string) (*GenericResult, error) {
	e.logger.Debug("go_to_timecode", zap.String("timecode", timecode))
	return nil, fmt.Errorf("go to timecode: not yet implemented in bridge")
}

func (e *Engine) GoToFrame(ctx context.Context, frameNumber int) (*GenericResult, error) {
	e.logger.Debug("go_to_frame", zap.Int("frame_number", frameNumber))
	return nil, fmt.Errorf("go to frame: not yet implemented in bridge")
}

func (e *Engine) GetSequenceDuration(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_sequence_duration")
	return nil, fmt.Errorf("get sequence duration: not yet implemented in bridge")
}

func (e *Engine) GetFrameCount(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_frame_count")
	return nil, fmt.Errorf("get frame count: not yet implemented in bridge")
}

func (e *Engine) GetCurrentTimecode(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_current_timecode")
	return nil, fmt.Errorf("get current timecode: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Selection & Focus
// ---------------------------------------------------------------------------

func (e *Engine) SelectClipsInRange(ctx context.Context, startSeconds, endSeconds float64) (*GenericResult, error) {
	e.logger.Debug("select_clips_in_range", zap.Float64("start", startSeconds), zap.Float64("end", endSeconds))
	return nil, fmt.Errorf("select clips in range: not yet implemented in bridge")
}

func (e *Engine) SelectAllOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error) {
	e.logger.Debug("select_all_on_track", zap.String("track_type", trackType), zap.Int("track_index", trackIndex))
	return nil, fmt.Errorf("select all on track: not yet implemented in bridge")
}

func (e *Engine) InvertSelection(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("invert_selection")
	return nil, fmt.Errorf("invert selection: not yet implemented in bridge")
}

func (e *Engine) GetSelectionRange(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_selection_range")
	return nil, fmt.Errorf("get selection range: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Render Status
// ---------------------------------------------------------------------------

func (e *Engine) GetRenderStatus(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_render_status")
	return nil, fmt.Errorf("get render status: not yet implemented in bridge")
}

func (e *Engine) IsRendering(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("is_rendering")
	return nil, fmt.Errorf("is rendering: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Sequence Metadata
// ---------------------------------------------------------------------------

func (e *Engine) GetSequenceMetadata(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_sequence_metadata")
	return nil, fmt.Errorf("get sequence metadata: not yet implemented in bridge")
}

func (e *Engine) SetSequenceMetadata(ctx context.Context, key, value string) (*GenericResult, error) {
	e.logger.Debug("set_sequence_metadata", zap.String("key", key), zap.String("value", value))
	return nil, fmt.Errorf("set sequence metadata: not yet implemented in bridge")
}

func (e *Engine) GetSequenceColorSpace(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_sequence_color_space")
	return nil, fmt.Errorf("get sequence color space: not yet implemented in bridge")
}

func (e *Engine) SetSequenceColorSpace(ctx context.Context, colorSpace string) (*GenericResult, error) {
	e.logger.Debug("set_sequence_color_space", zap.String("color_space", colorSpace))
	return nil, fmt.Errorf("set sequence color space: not yet implemented in bridge")
}
