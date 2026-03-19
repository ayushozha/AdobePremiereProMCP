package orchestrator

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// General Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetGeneralPreferences(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_general_preferences")
	return nil, fmt.Errorf("get general preferences: not yet implemented in bridge")
}

func (e *Engine) SetDefaultStillDuration(ctx context.Context, frames int) (*GenericResult, error) {
	e.logger.Debug("set_default_still_duration", zap.Int("frames", frames))
	return nil, fmt.Errorf("set default still duration: not yet implemented in bridge")
}

func (e *Engine) SetDefaultTransitionDuration(ctx context.Context, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_default_transition_duration", zap.Float64("seconds", seconds))
	return nil, fmt.Errorf("set default transition duration: not yet implemented in bridge")
}

func (e *Engine) SetDefaultAudioTransitionDuration(ctx context.Context, seconds float64) (*GenericResult, error) {
	e.logger.Debug("set_default_audio_transition_duration", zap.Float64("seconds", seconds))
	return nil, fmt.Errorf("set default audio transition duration: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Appearance
// ---------------------------------------------------------------------------

func (e *Engine) GetBrightness(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_brightness")
	return nil, fmt.Errorf("get brightness: not yet implemented in bridge")
}

func (e *Engine) SetBrightness(ctx context.Context, level int) (*GenericResult, error) {
	e.logger.Debug("set_brightness", zap.Int("level", level))
	return nil, fmt.Errorf("set brightness: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Auto Save
// ---------------------------------------------------------------------------

func (e *Engine) SetAutoSaveEnabled(ctx context.Context, enabled bool) (*GenericResult, error) {
	e.logger.Debug("set_auto_save_enabled", zap.Bool("enabled", enabled))
	return nil, fmt.Errorf("set auto save enabled: not yet implemented in bridge")
}

func (e *Engine) SetAutoSaveMaxVersions(ctx context.Context, count int) (*GenericResult, error) {
	e.logger.Debug("set_auto_save_max_versions", zap.Int("count", count))
	return nil, fmt.Errorf("set auto save max versions: not yet implemented in bridge")
}

func (e *Engine) GetAutoSaveLocation(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_auto_save_location")
	return nil, fmt.Errorf("get auto save location: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Playback Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetPlaybackResolution(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_playback_resolution")
	return nil, fmt.Errorf("get playback resolution: not yet implemented in bridge")
}

func (e *Engine) SetPlaybackResolution(ctx context.Context, quality string) (*GenericResult, error) {
	e.logger.Debug("set_playback_resolution", zap.String("quality", quality))
	return nil, fmt.Errorf("set playback resolution: not yet implemented in bridge")
}

func (e *Engine) GetPrerollFrames(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_preroll_frames")
	return nil, fmt.Errorf("get preroll frames: not yet implemented in bridge")
}

func (e *Engine) SetPrerollFrames(ctx context.Context, frames int) (*GenericResult, error) {
	e.logger.Debug("set_preroll_frames", zap.Int("frames", frames))
	return nil, fmt.Errorf("set preroll frames: not yet implemented in bridge")
}

func (e *Engine) GetPostrollFrames(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_postroll_frames")
	return nil, fmt.Errorf("get postroll frames: not yet implemented in bridge")
}

func (e *Engine) SetPostrollFrames(ctx context.Context, frames int) (*GenericResult, error) {
	e.logger.Debug("set_postroll_frames", zap.Int("frames", frames))
	return nil, fmt.Errorf("set postroll frames: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Timeline Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetTimelineSettings(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_timeline_settings")
	return nil, fmt.Errorf("get timeline settings: not yet implemented in bridge")
}

func (e *Engine) SetTimeDisplayFormat(ctx context.Context, format string) (*GenericResult, error) {
	e.logger.Debug("set_time_display_format", zap.String("format", format))
	return nil, fmt.Errorf("set time display format: not yet implemented in bridge")
}

func (e *Engine) SetVideoTransitionDefaultDuration(ctx context.Context, frames int) (*GenericResult, error) {
	e.logger.Debug("set_video_transition_default_duration", zap.Int("frames", frames))
	return nil, fmt.Errorf("set video transition default duration: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Media Preferences
// ---------------------------------------------------------------------------

func (e *Engine) GetMediaCacheSettings(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_media_cache_settings")
	return nil, fmt.Errorf("get media cache settings: not yet implemented in bridge")
}

func (e *Engine) SetMediaCacheLocation(ctx context.Context, path string) (*GenericResult, error) {
	e.logger.Debug("set_media_cache_location", zap.String("path", path))
	return nil, fmt.Errorf("set media cache location: not yet implemented in bridge")
}

func (e *Engine) SetMediaCacheSize(ctx context.Context, maxGB float64) (*GenericResult, error) {
	e.logger.Debug("set_media_cache_size", zap.Float64("max_gb", maxGB))
	return nil, fmt.Errorf("set media cache size: not yet implemented in bridge")
}

func (e *Engine) CleanMediaCacheOlderThan(ctx context.Context, days int) (*GenericResult, error) {
	e.logger.Debug("clean_media_cache_older_than", zap.Int("days", days))
	return nil, fmt.Errorf("clean media cache older than: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Label Colors
// ---------------------------------------------------------------------------

func (e *Engine) GetLabelColorNames(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_label_color_names")
	return nil, fmt.Errorf("get label color names: not yet implemented in bridge")
}

func (e *Engine) SetLabelColorName(ctx context.Context, index int, name string) (*GenericResult, error) {
	e.logger.Debug("set_label_color_name", zap.Int("index", index), zap.String("name", name))
	return nil, fmt.Errorf("set label color name: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// GPU / Renderer
// ---------------------------------------------------------------------------

func (e *Engine) GetRendererInfo(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_renderer_info")
	return nil, fmt.Errorf("get renderer info: not yet implemented in bridge")
}

func (e *Engine) GetGPUInfo(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_gpu_info")
	return nil, fmt.Errorf("get GPU info: not yet implemented in bridge")
}

func (e *Engine) SetRenderer(ctx context.Context, rendererName string) (*GenericResult, error) {
	e.logger.Debug("set_renderer", zap.String("renderer", rendererName))
	return nil, fmt.Errorf("set renderer: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Project Defaults
// ---------------------------------------------------------------------------

func (e *Engine) GetDefaultSequencePresets(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_default_sequence_presets")
	return nil, fmt.Errorf("get default sequence presets: not yet implemented in bridge")
}

func (e *Engine) SetDefaultSequencePreset(ctx context.Context, presetPath string) (*GenericResult, error) {
	e.logger.Debug("set_default_sequence_preset", zap.String("preset_path", presetPath))
	return nil, fmt.Errorf("set default sequence preset: not yet implemented in bridge")
}

func (e *Engine) GetInstalledCodecs(ctx context.Context) (*GenericResult, error) {
	e.logger.Debug("get_installed_codecs")
	return nil, fmt.Errorf("get installed codecs: not yet implemented in bridge")
}
