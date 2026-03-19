package orchestrator

import "context"

// MediaClient abstracts the Rust media engine gRPC service.
// Implementations live in internal/grpc and translate between Go-native types
// and the generated proto stubs.
type MediaClient interface {
	// ScanAssets indexes all media files under a directory.
	ScanAssets(ctx context.Context, dir string, recursive bool, extensions []string) (*ScanResult, error)

	// ProbeMedia retrieves detailed metadata for a single file.
	ProbeMedia(ctx context.Context, filePath string) (*AssetInfo, error)

	// AnalyzeWaveform inspects audio levels and silence regions.
	AnalyzeWaveform(ctx context.Context, filePath string, opts *WaveformOptions) (*WaveformResult, error)

	// DetectScenes finds scene-change boundaries in a video file.
	DetectScenes(ctx context.Context, filePath string, threshold float64) (*SceneResult, error)
}

// IntelClient abstracts the Python intelligence gRPC service.
type IntelClient interface {
	// ParseScript converts raw script text (or a file) into structured segments.
	ParseScript(ctx context.Context, text string, filePath string, format string) (*ParsedScript, error)

	// GenerateEDL produces an edit decision list from matched segments and assets.
	GenerateEDL(ctx context.Context, segments []*ScriptSegment, assets []*AssetInfo, settings *EDLSettings) (*EDL, error)

	// MatchAssets pairs script segments with the best available media assets.
	MatchAssets(ctx context.Context, segments []*ScriptSegment, assets []*AssetInfo, strategy string) (*MatchResult, error)

	// AnalyzePacing reviews EDL timing and suggests adjustments.
	AnalyzePacing(ctx context.Context, edl *EDL, targetMood string) (*PacingResult, error)
}

// PremiereClient abstracts the TypeScript Premiere bridge gRPC service.
type PremiereClient interface {
	// Ping checks whether Premiere Pro is running and reachable.
	Ping(ctx context.Context) (*PingResult, error)

	// GetProjectState retrieves the current project metadata.
	GetProjectState(ctx context.Context) (*ProjectState, error)

	// CreateSequence creates a new sequence in the open project.
	CreateSequence(ctx context.Context, params *CreateSequenceParams) (*SequenceResult, error)

	// ImportMedia imports a media file into the project (optionally into a bin).
	ImportMedia(ctx context.Context, filePath string, targetBin string) (*ImportResult, error)

	// PlaceClip places a clip on the timeline at the specified track and position.
	PlaceClip(ctx context.Context, params *PlaceClipParams) (*ClipResult, error)

	// RemoveClip removes a clip from a sequence.
	RemoveClip(ctx context.Context, clipID, sequenceID string) error

	// AddTransition inserts a transition at the given position.
	AddTransition(ctx context.Context, params *TransitionParams) error

	// AddText adds a text overlay to the timeline.
	AddText(ctx context.Context, params *TextParams) (*ClipResult, error)

	// SetAudioLevel adjusts the audio gain of a clip.
	SetAudioLevel(ctx context.Context, clipID, sequenceID string, levelDB float64) error

	// GetTimelineState retrieves the current state of a sequence's timeline.
	GetTimelineState(ctx context.Context, sequenceID string) (*TimelineState, error)

	// ExportSequence starts an export job for a sequence.
	ExportSequence(ctx context.Context, params *ExportParams) (*ExportResult, error)

	// ExecuteEDL assembles an entire timeline from an edit decision list.
	ExecuteEDL(ctx context.Context, edl *EDL) (*EDLExecutionResult, error)
}

// Orchestrator defines the complete set of operations exposed by the engine.
// The MCP handler layer calls these methods and translates results to MCP
// tool responses.
type Orchestrator interface {
	// --- Health ---
	Ping(ctx context.Context) (*PingResult, error)

	// --- Project ---
	GetProject(ctx context.Context) (*ProjectState, error)

	// --- Sequence ---
	CreateSequence(ctx context.Context, params *CreateSequenceParams) (*SequenceResult, error)
	GetTimeline(ctx context.Context, sequenceID string) (*TimelineState, error)

	// --- Clip Operations ---
	ImportMedia(ctx context.Context, filePath string, targetBin string) (*ImportResult, error)
	PlaceClip(ctx context.Context, params *PlaceClipParams) (*ClipResult, error)
	RemoveClip(ctx context.Context, clipID, sequenceID string) error

	// --- Effects & Transitions ---
	AddTransition(ctx context.Context, params *TransitionParams) error
	AddText(ctx context.Context, params *TextParams) (*ClipResult, error)

	// --- Audio ---
	SetAudioLevel(ctx context.Context, clipID, sequenceID string, levelDB float64) error

	// --- Export ---
	Export(ctx context.Context, params *ExportParams) (*ExportResult, error)

	// --- Media Analysis (Rust engine) ---
	ScanAssets(ctx context.Context, dir string, recursive bool, extensions []string) (*ScanResult, error)

	// --- Intelligence (Python service) ---
	ParseScript(ctx context.Context, text string, filePath string, format string) (*ParsedScript, error)

	// --- Composite Workflows ---
	AutoEdit(ctx context.Context, params *AutoEditParams) (*AutoEditResult, error)
}
