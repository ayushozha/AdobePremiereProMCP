package orchestrator

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Compile-time check: *Engine satisfies the Orchestrator interface.
var _ Orchestrator = (*Engine)(nil)

// Engine is the central coordinator of the PremierPro MCP service. It
// delegates simple operations to the appropriate gRPC client and orchestrates
// multi-step workflows (e.g., AutoEdit) across all three back-end services.
type Engine struct {
	media    MediaClient
	intel    IntelClient
	premiere PremiereClient
	logger   *zap.Logger
}

// New creates an Engine with the given service clients and logger.
// All three clients must be non-nil; the constructor panics otherwise so
// misconfiguration is caught at startup rather than at request time.
func New(media MediaClient, intel IntelClient, premiere PremiereClient, logger *zap.Logger) *Engine {
	if media == nil {
		panic("orchestrator: MediaClient must not be nil")
	}
	if intel == nil {
		panic("orchestrator: IntelClient must not be nil")
	}
	if premiere == nil {
		panic("orchestrator: PremiereClient must not be nil")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Engine{
		media:    media,
		intel:    intel,
		premiere: premiere,
		logger:   logger,
	}
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

// Ping delegates to the TypeScript Premiere bridge to check whether Premiere
// Pro is running and responsive.
func (e *Engine) Ping(ctx context.Context) (*PingResult, error) {
	e.logger.Debug("ping: checking Premiere Pro health")
	res, err := e.premiere.Ping(ctx)
	if err != nil {
		e.logger.Error("ping: failed", zap.Error(err))
		return nil, fmt.Errorf("could not reach Premiere Pro — make sure it is running and the CEP panel is loaded: %w", err)
	}
	e.logger.Info("ping: success",
		zap.Bool("premiere_running", res.PremiereRunning),
		zap.String("version", res.PremiereVersion),
		zap.Bool("project_open", res.ProjectOpen),
	)
	return res, nil
}

// ---------------------------------------------------------------------------
// Project
// ---------------------------------------------------------------------------

// GetProject retrieves the current Premiere Pro project state.
func (e *Engine) GetProject(ctx context.Context) (*ProjectState, error) {
	e.logger.Debug("get_project: retrieving project state")
	res, err := e.premiere.GetProjectState(ctx)
	if err != nil {
		e.logger.Error("get_project: failed", zap.Error(err))
		return nil, fmt.Errorf("could not retrieve project state — open a project first with premiere_open_project: %w", err)
	}
	e.logger.Info("get_project: success",
		zap.String("project", res.ProjectName),
		zap.Int("sequences", len(res.Sequences)),
	)
	return res, nil
}

// ---------------------------------------------------------------------------
// Sequence
// ---------------------------------------------------------------------------

// CreateSequence creates a new sequence in the open project.
func (e *Engine) CreateSequence(ctx context.Context, params *CreateSequenceParams) (*SequenceResult, error) {
	if params == nil {
		return nil, fmt.Errorf("create_sequence: params must not be nil — provide name, resolution, and frame rate")
	}
	e.logger.Debug("create_sequence: creating",
		zap.String("name", params.Name),
		zap.Uint32("width", params.Resolution.Width),
		zap.Uint32("height", params.Resolution.Height),
		zap.Float64("fps", params.FrameRate),
	)
	res, err := e.premiere.CreateSequence(ctx, params)
	if err != nil {
		e.logger.Error("create_sequence: failed", zap.Error(err))
		return nil, fmt.Errorf("failed to create sequence %q — make sure a project is open (try premiere_open_project first): %w", params.Name, err)
	}
	e.logger.Info("create_sequence: success",
		zap.String("sequence_id", res.SequenceID),
		zap.String("name", res.Name),
	)
	return res, nil
}

// GetTimeline retrieves the current state of a sequence's timeline.
func (e *Engine) GetTimeline(ctx context.Context, sequenceID string) (*TimelineState, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("get_timeline: sequence_id is required — use premiere_get_sequence_list to find available sequences")
	}
	e.logger.Debug("get_timeline: retrieving", zap.String("sequence_id", sequenceID))
	res, err := e.premiere.GetTimelineState(ctx, sequenceID)
	if err != nil {
		e.logger.Error("get_timeline: failed", zap.String("sequence_id", sequenceID), zap.Error(err))
		return nil, fmt.Errorf("could not get timeline for sequence %q — create a sequence first with premiere_create_sequence: %w", sequenceID, err)
	}
	e.logger.Info("get_timeline: success",
		zap.String("sequence_id", sequenceID),
		zap.Int("video_tracks", len(res.VideoTracks)),
		zap.Int("audio_tracks", len(res.AudioTracks)),
	)
	return res, nil
}

// ---------------------------------------------------------------------------
// Clip Operations
// ---------------------------------------------------------------------------

// ImportMedia imports a media file into the open project.
func (e *Engine) ImportMedia(ctx context.Context, filePath string, targetBin string) (*ImportResult, error) {
	if filePath == "" {
		return nil, fmt.Errorf("import_media: file_path is required — provide the full path to a media file (e.g. /Users/you/Videos/clip.mp4)")
	}
	e.logger.Debug("import_media: importing",
		zap.String("file", filePath),
		zap.String("bin", targetBin),
	)
	res, err := e.premiere.ImportMedia(ctx, filePath, targetBin)
	if err != nil {
		e.logger.Error("import_media: failed", zap.String("file", filePath), zap.Error(err))
		return nil, fmt.Errorf("failed to import %q — verify the file exists and a project is open (try premiere_open_project first): %w", filePath, err)
	}
	e.logger.Info("import_media: success",
		zap.String("project_item_id", res.ProjectItemID),
		zap.String("name", res.Name),
	)
	return res, nil
}

// PlaceClip places a clip on the timeline.
func (e *Engine) PlaceClip(ctx context.Context, params *PlaceClipParams) (*ClipResult, error) {
	if params == nil {
		return nil, fmt.Errorf("place_clip: params must not be nil — provide source_path, track type, and track index")
	}
	e.logger.Debug("place_clip: placing",
		zap.String("source", params.SourcePath),
		zap.Int("track_type", int(params.Track.Type)),
		zap.Uint32("track_index", params.Track.TrackIndex),
	)
	res, err := e.premiere.PlaceClip(ctx, params)
	if err != nil {
		e.logger.Error("place_clip: failed", zap.String("source", params.SourcePath), zap.Error(err))
		return nil, fmt.Errorf("failed to place clip %q — make sure a sequence exists (try premiere_create_sequence) and the media is imported: %w", params.SourcePath, err)
	}
	e.logger.Info("place_clip: success", zap.String("clip_id", res.ClipID))
	return res, nil
}

// RemoveClip removes a clip from a sequence.
func (e *Engine) RemoveClip(ctx context.Context, clipID, sequenceID string) error {
	if clipID == "" || sequenceID == "" {
		return fmt.Errorf("remove_clip: both clip_id and sequence_id are required — use premiere_get_all_clips to find clip IDs")
	}
	e.logger.Debug("remove_clip: removing",
		zap.String("clip_id", clipID),
		zap.String("sequence_id", sequenceID),
	)
	if err := e.premiere.RemoveClip(ctx, clipID, sequenceID); err != nil {
		e.logger.Error("remove_clip: failed",
			zap.String("clip_id", clipID),
			zap.String("sequence_id", sequenceID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove clip %q from sequence %q: %w", clipID, sequenceID, err)
	}
	e.logger.Info("remove_clip: success",
		zap.String("clip_id", clipID),
		zap.String("sequence_id", sequenceID),
	)
	return nil
}

// ---------------------------------------------------------------------------
// Effects & Transitions
// ---------------------------------------------------------------------------

// AddTransition inserts a transition at the specified position.
func (e *Engine) AddTransition(ctx context.Context, params *TransitionParams) error {
	if params == nil {
		return fmt.Errorf("add_transition: params must not be nil — provide sequence_id, transition_type, and duration")
	}
	e.logger.Debug("add_transition: adding",
		zap.String("sequence_id", params.SequenceID),
		zap.String("type", params.TransitionType),
		zap.Float64("duration_s", params.DurationSeconds),
	)
	if err := e.premiere.AddTransition(ctx, params); err != nil {
		e.logger.Error("add_transition: failed", zap.Error(err))
		return fmt.Errorf("failed to add %s transition — make sure clips exist at the edit point in sequence %q: %w", params.TransitionType, params.SequenceID, err)
	}
	e.logger.Info("add_transition: success",
		zap.String("sequence_id", params.SequenceID),
		zap.String("type", params.TransitionType),
	)
	return nil
}

// AddText adds a text overlay to the timeline.
func (e *Engine) AddText(ctx context.Context, params *TextParams) (*ClipResult, error) {
	if params == nil {
		return nil, fmt.Errorf("add_text: params must not be nil — provide sequence_id and text content")
	}
	e.logger.Debug("add_text: adding",
		zap.String("sequence_id", params.SequenceID),
		zap.String("text", params.Text),
	)
	res, err := e.premiere.AddText(ctx, params)
	if err != nil {
		e.logger.Error("add_text: failed", zap.Error(err))
		return nil, fmt.Errorf("failed to add text overlay — create a sequence first with premiere_create_sequence: %w", err)
	}
	e.logger.Info("add_text: success", zap.String("clip_id", res.ClipID))
	return res, nil
}

// ---------------------------------------------------------------------------
// Audio
// ---------------------------------------------------------------------------

// SetAudioLevel adjusts the audio gain for a clip.
func (e *Engine) SetAudioLevel(ctx context.Context, clipID, sequenceID string, levelDB float64) error {
	if clipID == "" || sequenceID == "" {
		return fmt.Errorf("set_audio_level: both clip_id and sequence_id are required — use premiere_get_all_clips to find clip IDs")
	}
	e.logger.Debug("set_audio_level: setting",
		zap.String("clip_id", clipID),
		zap.String("sequence_id", sequenceID),
		zap.Float64("level_db", levelDB),
	)
	if err := e.premiere.SetAudioLevel(ctx, clipID, sequenceID, levelDB); err != nil {
		e.logger.Error("set_audio_level: failed", zap.Error(err))
		return fmt.Errorf("failed to set audio level for clip %q — verify the clip has audio and is on the timeline: %w", clipID, err)
	}
	e.logger.Info("set_audio_level: success",
		zap.String("clip_id", clipID),
		zap.Float64("level_db", levelDB),
	)
	return nil
}

// EvalCommand is the generic dispatcher for calling any ExtendScript function
// by name with JSON-encoded arguments. This is the critical pipeline that
// allows all ~1000 MCP tools to call through to Premiere Pro.
func (e *Engine) EvalCommand(ctx context.Context, functionName, argsJSON string) (string, error) {
	if functionName == "" {
		return "", fmt.Errorf("eval_command: function_name is required")
	}
	e.logger.Debug("eval_command",
		zap.String("function_name", functionName),
	)
	result, err := e.premiere.EvalCommand(ctx, functionName, argsJSON)
	if err != nil {
		e.logger.Error("eval_command: failed",
			zap.String("function_name", functionName),
			zap.Error(err),
		)
		return "", fmt.Errorf("command %q failed — make sure Premiere Pro is running and a project is open: %w", functionName, err)
	}
	e.logger.Debug("eval_command: success",
		zap.String("function_name", functionName),
	)
	return result, nil
}

// EvalAudioCommand is a generic dispatcher for audio and track management
// ExtendScript commands. The MCP audio_tools layer calls this with a command
// name (e.g. "setAudioLevelKeyframe") and an args map, and the engine
// forwards to the Premiere bridge via EvalScript.
func (e *Engine) EvalAudioCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error) {
	if command == "" {
		return nil, fmt.Errorf("eval_audio_command: command name is required")
	}
	e.logger.Debug("eval_audio_command",
		zap.String("command", command),
	)
	result, err := e.premiere.EvalAudioCommand(ctx, command, args)
	if err != nil {
		e.logger.Error("eval_audio_command: failed",
			zap.String("command", command),
			zap.Error(err),
		)
		return nil, fmt.Errorf("audio command %q failed — make sure a sequence with audio tracks exists (try premiere_create_sequence first): %w", command, err)
	}
	e.logger.Debug("eval_audio_command: success",
		zap.String("command", command),
	)
	return result, nil
}

// EvalImmersiveCommand is the passthrough for immersive-video (VR/360, HDR,
// stereoscopic, frame-rate, aspect-ratio, timecode, render, caption)
// ExtendScript commands. The MCP immersive_tools layer calls this with a
// command name and an args map, and the engine forwards to the Premiere
// bridge via EvalImmersiveCommand.
func (e *Engine) EvalImmersiveCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error) {
	if command == "" {
		return nil, fmt.Errorf("eval_immersive_command: command name is required")
	}
	e.logger.Debug("eval_immersive_command",
		zap.String("command", command),
	)
	result, err := e.premiere.EvalImmersiveCommand(ctx, command, args)
	if err != nil {
		e.logger.Error("eval_immersive_command: failed",
			zap.String("command", command),
			zap.Error(err),
		)
		return nil, fmt.Errorf("immersive command %q failed — make sure Premiere Pro is running and a project is open: %w", command, err)
	}
	e.logger.Debug("eval_immersive_command: success",
		zap.String("command", command),
	)
	return result, nil
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

// Export starts an export job for a sequence.
func (e *Engine) Export(ctx context.Context, params *ExportParams) (*ExportResult, error) {
	if params == nil {
		return nil, fmt.Errorf("export: params must not be nil — provide sequence_id, output_path, and preset")
	}
	if params.SequenceID == "" {
		return nil, fmt.Errorf("export: sequence_id is required — use premiere_get_sequence_list to find available sequences, or create one with premiere_create_sequence")
	}
	e.logger.Debug("export: starting",
		zap.String("sequence_id", params.SequenceID),
		zap.String("output", params.OutputPath),
		zap.Int("preset", int(params.Preset)),
	)
	res, err := e.premiere.ExportSequence(ctx, params)
	if err != nil {
		e.logger.Error("export: failed", zap.Error(err))
		return nil, fmt.Errorf("failed to export sequence %q — verify the sequence has content and the output path is writable: %w", params.SequenceID, err)
	}
	e.logger.Info("export: success",
		zap.String("job_id", res.JobID),
		zap.String("status", res.Status),
		zap.String("output", res.OutputPath),
	)
	return res, nil
}

// ---------------------------------------------------------------------------
// Media Analysis (delegated to Rust engine)
// ---------------------------------------------------------------------------

// ScanAssets delegates to the Rust media engine to index a directory.
func (e *Engine) ScanAssets(ctx context.Context, dir string, recursive bool, extensions []string) (*ScanResult, error) {
	if dir == "" {
		return nil, fmt.Errorf("scan_assets: directory path is required — provide the full path to a folder with media files")
	}
	e.logger.Debug("scan_assets: scanning",
		zap.String("dir", dir),
		zap.Bool("recursive", recursive),
		zap.Strings("extensions", extensions),
	)
	res, err := e.media.ScanAssets(ctx, dir, recursive, extensions)
	if err != nil {
		e.logger.Error("scan_assets: failed", zap.String("dir", dir), zap.Error(err))
		return nil, fmt.Errorf("failed to scan %q — verify the directory exists and is readable: %w", dir, err)
	}
	e.logger.Info("scan_assets: success",
		zap.Uint32("total_scanned", res.TotalFilesScanned),
		zap.Uint32("media_found", res.MediaFilesFound),
		zap.Float64("duration_s", res.ScanDurationSeconds),
	)
	return res, nil
}

// ---------------------------------------------------------------------------
// Intelligence (delegated to Python service)
// ---------------------------------------------------------------------------

// ParseScript delegates to the Python intelligence service.
func (e *Engine) ParseScript(ctx context.Context, text string, filePath string, format string) (*ParsedScript, error) {
	if text == "" && filePath == "" {
		return nil, fmt.Errorf("parse_script: provide either script text or a file path to the script")
	}
	e.logger.Debug("parse_script: parsing",
		zap.Bool("has_text", text != ""),
		zap.String("file", filePath),
		zap.String("format", format),
	)
	res, err := e.intel.ParseScript(ctx, text, filePath, format)
	if err != nil {
		e.logger.Error("parse_script: failed", zap.Error(err))
		return nil, fmt.Errorf("failed to parse script — check the script format and content: %w", err)
	}
	e.logger.Info("parse_script: success",
		zap.Uint32("segments", res.Metadata.SegmentCount),
		zap.String("title", res.Metadata.Title),
	)
	return res, nil
}
