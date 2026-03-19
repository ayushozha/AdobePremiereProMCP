package orchestrator

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Motion Graphics Templates (MOGRTs)
// ---------------------------------------------------------------------------

func (e *Engine) ImportMOGRT(ctx context.Context, mogrtPath, timeTicks string, videoTrackOffset, audioTrackOffset int) (*GenericResult, error) {
	e.logger.Debug("import_mogrt", zap.String("mogrt_path", mogrtPath), zap.String("time_ticks", timeTicks))
	return nil, fmt.Errorf("import mogrt: not yet implemented in bridge")
}

func (e *Engine) GetMOGRTProperties(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error) {
	e.logger.Debug("get_mogrt_properties", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex))
	return nil, fmt.Errorf("get mogrt properties: not yet implemented in bridge")
}

func (e *Engine) SetMOGRTText(ctx context.Context, trackIndex, clipIndex, propertyIndex int, text string) (*GenericResult, error) {
	e.logger.Debug("set_mogrt_text", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex), zap.Int("property_index", propertyIndex), zap.String("text", text))
	return nil, fmt.Errorf("set mogrt text: not yet implemented in bridge")
}

func (e *Engine) SetMOGRTProperty(ctx context.Context, trackIndex, clipIndex int, propertyName string, value string) (*GenericResult, error) {
	e.logger.Debug("set_mogrt_property", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex), zap.String("property_name", propertyName))
	return nil, fmt.Errorf("set mogrt property: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Titles & Lower Thirds
// ---------------------------------------------------------------------------

func (e *Engine) AddTitle(ctx context.Context, text string, trackIndex int, startTime, duration float64, styleJSON string) (*GenericResult, error) {
	e.logger.Debug("add_title", zap.String("text", text), zap.Int("track_index", trackIndex), zap.Float64("start_time", startTime), zap.Float64("duration", duration))
	return nil, fmt.Errorf("add title: not yet implemented in bridge")
}

func (e *Engine) AddLowerThird(ctx context.Context, name, title string, trackIndex int, startTime, duration float64) (*GenericResult, error) {
	e.logger.Debug("add_lower_third", zap.String("name", name), zap.String("title", title), zap.Int("track_index", trackIndex))
	return nil, fmt.Errorf("add lower third: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Captions & Subtitles
// ---------------------------------------------------------------------------

func (e *Engine) CreateCaptionTrack(ctx context.Context, format string) (*GenericResult, error) {
	e.logger.Debug("create_caption_track", zap.String("format", format))
	return nil, fmt.Errorf("create caption track: not yet implemented in bridge")
}

func (e *Engine) ImportCaptions(ctx context.Context, filePath, format string) (*GenericResult, error) {
	e.logger.Debug("import_captions", zap.String("file_path", filePath), zap.String("format", format))
	return nil, fmt.Errorf("import captions: not yet implemented in bridge")
}

func (e *Engine) GetCaptions(ctx context.Context, trackIndex int) (*GenericResult, error) {
	e.logger.Debug("get_captions", zap.Int("track_index", trackIndex))
	return nil, fmt.Errorf("get captions: not yet implemented in bridge")
}

func (e *Engine) AddCaption(ctx context.Context, trackIndex int, startTime, endTime float64, text string) (*GenericResult, error) {
	e.logger.Debug("add_caption", zap.Int("track_index", trackIndex), zap.Float64("start_time", startTime), zap.Float64("end_time", endTime), zap.String("text", text))
	return nil, fmt.Errorf("add caption: not yet implemented in bridge")
}

func (e *Engine) EditCaption(ctx context.Context, trackIndex, captionIndex int, text string) (*GenericResult, error) {
	e.logger.Debug("edit_caption", zap.Int("track_index", trackIndex), zap.Int("caption_index", captionIndex), zap.String("text", text))
	return nil, fmt.Errorf("edit caption: not yet implemented in bridge")
}

func (e *Engine) DeleteCaption(ctx context.Context, trackIndex, captionIndex int) (*GenericResult, error) {
	e.logger.Debug("delete_caption", zap.Int("track_index", trackIndex), zap.Int("caption_index", captionIndex))
	return nil, fmt.Errorf("delete caption: not yet implemented in bridge")
}

func (e *Engine) ExportCaptions(ctx context.Context, outputPath, format string) (*GenericResult, error) {
	e.logger.Debug("export_captions", zap.String("output_path", outputPath), zap.String("format", format))
	return nil, fmt.Errorf("export captions: not yet implemented in bridge")
}

func (e *Engine) StyleCaptions(ctx context.Context, trackIndex int, font string, size float64, color, bgColor, position string) (*GenericResult, error) {
	e.logger.Debug("style_captions", zap.Int("track_index", trackIndex), zap.String("font", font), zap.Float64("size", size))
	return nil, fmt.Errorf("style captions: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Graphics
// ---------------------------------------------------------------------------

func (e *Engine) CreateColorMatte(ctx context.Context, name string, red, green, blue, width, height int) (*GenericResult, error) {
	e.logger.Debug("create_color_matte", zap.String("name", name), zap.Int("red", red), zap.Int("green", green), zap.Int("blue", blue))
	return nil, fmt.Errorf("create color matte: not yet implemented in bridge")
}

func (e *Engine) PlaceColorMatte(ctx context.Context, projectItemIndex, trackIndex int, startTime, duration float64) (*GenericResult, error) {
	e.logger.Debug("place_color_matte", zap.Int("project_item_index", projectItemIndex), zap.Int("track_index", trackIndex))
	return nil, fmt.Errorf("place color matte: not yet implemented in bridge")
}

func (e *Engine) CreateTransparentVideo(ctx context.Context, name string, width, height int, duration float64) (*GenericResult, error) {
	e.logger.Debug("create_transparent_video", zap.String("name", name), zap.Int("width", width), zap.Int("height", height))
	return nil, fmt.Errorf("create transparent video: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Speed & Time (Time Remapping, Freeze Frame)
// ---------------------------------------------------------------------------

func (e *Engine) SetTimeRemapping(ctx context.Context, trackIndex, clipIndex int, enabled bool) (*GenericResult, error) {
	e.logger.Debug("set_time_remapping", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex), zap.Bool("enabled", enabled))
	return nil, fmt.Errorf("set time remapping: not yet implemented in bridge")
}

func (e *Engine) AddTimeRemapKeyframe(ctx context.Context, trackIndex, clipIndex int, time, speed float64) (*GenericResult, error) {
	e.logger.Debug("add_time_remap_keyframe", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex), zap.Float64("time", time), zap.Float64("speed", speed))
	return nil, fmt.Errorf("add time remap keyframe: not yet implemented in bridge")
}

func (e *Engine) FreezeFrame(ctx context.Context, trackIndex, clipIndex int, time, duration float64) (*GenericResult, error) {
	e.logger.Debug("freeze_frame", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex), zap.Float64("time", time), zap.Float64("duration", duration))
	return nil, fmt.Errorf("freeze frame: not yet implemented in bridge")
}

// ---------------------------------------------------------------------------
// Scene Edit Detection
// ---------------------------------------------------------------------------

func (e *Engine) DetectSceneEdits(ctx context.Context, trackIndex, clipIndex int, sensitivity float64) (*GenericResult, error) {
	e.logger.Debug("detect_scene_edits", zap.Int("track_index", trackIndex), zap.Int("clip_index", clipIndex), zap.Float64("sensitivity", sensitivity))
	return nil, fmt.Errorf("detect scene edits: not yet implemented in bridge")
}
