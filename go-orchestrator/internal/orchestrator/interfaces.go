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

	// EvalAudioCommand runs an audio/track management ExtendScript function.
	EvalAudioCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error)
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

	// --- Sequence Management ---
	CreateSequenceFromClips(ctx context.Context, name string, clipIndices []int) (*SequenceResult, error)
	DuplicateSequence(ctx context.Context, sequenceIndex int) (*SequenceResult, error)
	DeleteSequence(ctx context.Context, sequenceIndex int) (*GenericResult, error)
	RenameSequence(ctx context.Context, sequenceIndex int, newName string) (*GenericResult, error)
	GetSequenceSettings(ctx context.Context, sequenceIndex int) (*SequenceSettings, error)
	SetSequenceSettings(ctx context.Context, params *SetSequenceSettingsParams) (*GenericResult, error)
	GetActiveSequence(ctx context.Context) (*SequenceSettings, error)
	SetActiveSequence(ctx context.Context, sequenceIndex int) (*GenericResult, error)
	GetSequenceList(ctx context.Context) (*SequenceListResult, error)

	// --- Playhead & In/Out Points ---
	GetPlayheadPosition(ctx context.Context) (*PlayheadResult, error)
	SetPlayheadPosition(ctx context.Context, seconds float64) (*GenericResult, error)
	SetInPoint(ctx context.Context, seconds float64) (*GenericResult, error)
	SetOutPoint(ctx context.Context, seconds float64) (*GenericResult, error)
	GetInOutPoints(ctx context.Context) (*InOutPointsResult, error)
	ClearInOutPoints(ctx context.Context) (*GenericResult, error)

	// --- Work Area & Preview ---
	SetWorkArea(ctx context.Context, inSeconds, outSeconds float64) (*GenericResult, error)
	RenderPreviewFiles(ctx context.Context, inSeconds, outSeconds float64) (*GenericResult, error)
	DeletePreviewFiles(ctx context.Context) (*GenericResult, error)

	// --- Nesting & Reframing ---
	CreateNestedSequence(ctx context.Context, trackIndex int, clipIndices []int) (*GenericResult, error)
	AutoReframeSequence(ctx context.Context, numerator, denominator int, motionPreset string) (*GenericResult, error)

	// --- Generated Media ---
	InsertBlackVideo(ctx context.Context, trackIndex int, startTime, duration float64) (*GenericResult, error)
	InsertBarsAndTone(ctx context.Context, width, height int, duration float64) (*GenericResult, error)

	// --- Markers ---
	GetSequenceMarkers(ctx context.Context) (*MarkersResult, error)
	AddSequenceMarker(ctx context.Context, params *AddMarkerParams) (*GenericResult, error)
	DeleteSequenceMarker(ctx context.Context, markerIndex int) (*GenericResult, error)
	NavigateToMarker(ctx context.Context, markerIndex int) (*GenericResult, error)

	// --- Project Management ---
	NewProject(ctx context.Context, path string) (*GenericResult, error)
	OpenProject(ctx context.Context, path string) (*GenericResult, error)
	SaveProject(ctx context.Context) (*GenericResult, error)
	SaveProjectAs(ctx context.Context, path string) (*GenericResult, error)
	CloseProject(ctx context.Context, saveFirst bool) (*GenericResult, error)
	GetProjectInfo(ctx context.Context) (*ProjectInfoResult, error)

	// --- Bin / Item Management ---
	ImportFiles(ctx context.Context, filePaths []string, targetBin string) (*GenericResult, error)
	ImportFolder(ctx context.Context, folderPath string, targetBin string) (*GenericResult, error)
	CreateBin(ctx context.Context, name string, parentBin string) (*GenericResult, error)
	RenameBin(ctx context.Context, binPath string, newName string) (*GenericResult, error)
	DeleteBin(ctx context.Context, binPath string) (*GenericResult, error)
	MoveBinItem(ctx context.Context, itemPath string, destBin string) (*GenericResult, error)
	FindProjectItems(ctx context.Context, searchQuery string) (*ProjectItemsResult, error)
	GetProjectItems(ctx context.Context, binPath string) (*ProjectItemsResult, error)
	SetItemLabel(ctx context.Context, itemPath string, colorIndex int) (*GenericResult, error)
	GetItemMetadata(ctx context.Context, itemPath string) (*ItemMetadataResult, error)
	SetItemMetadata(ctx context.Context, itemPath string, key string, value string) (*GenericResult, error)

	// --- Media Management ---
	RelinkMedia(ctx context.Context, itemPath string, newMediaPath string) (*GenericResult, error)
	MakeOffline(ctx context.Context, itemPath string) (*GenericResult, error)
	GetOfflineItems(ctx context.Context) (*ProjectItemsResult, error)

	// --- Project Settings ---
	SetScratchDisk(ctx context.Context, scratchType string, path string) (*GenericResult, error)
	ConsolidateDuplicates(ctx context.Context) (*ConsolidateResult, error)
	GetProjectSettingsInfo(ctx context.Context) (*ProjectSettingsResult, error)

	// --- Clip Operations ---
	ImportMedia(ctx context.Context, filePath string, targetBin string) (*ImportResult, error)
	PlaceClip(ctx context.Context, params *PlaceClipParams) (*ClipResult, error)
	RemoveClip(ctx context.Context, clipID, sequenceID string) error

	// --- Clip Operations (Extended) ---
	InsertClip(ctx context.Context, projectItemIndex int, time float64, vTrackIndex, aTrackIndex int) (*GenericResult, error)
	OverwriteClip(ctx context.Context, projectItemIndex int, time float64, vTrackIndex, aTrackIndex int) (*GenericResult, error)
	RemoveClipFromTrack(ctx context.Context, trackType string, trackIndex, clipIndex int, ripple bool) (*GenericResult, error)
	MoveClip(ctx context.Context, trackType string, trackIndex, clipIndex int, newStartTime float64) (*GenericResult, error)
	CopyClip(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	PasteClip(ctx context.Context, trackType string, trackIndex int, time float64) (*GenericResult, error)
	DuplicateClip(ctx context.Context, trackType string, trackIndex, clipIndex, destTrackIndex int, destTime float64) (*GenericResult, error)
	RazorClip(ctx context.Context, trackType string, trackIndex int, time float64) (*GenericResult, error)
	RazorAllTracks(ctx context.Context, time float64) (*GenericResult, error)
	GetClipInfo(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	GetClipsOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	GetAllClips(ctx context.Context) (*GenericResult, error)
	SetClipName(ctx context.Context, trackType string, trackIndex, clipIndex int, name string) (*GenericResult, error)
	SetClipEnabled(ctx context.Context, trackType string, trackIndex, clipIndex int, enabled bool) (*GenericResult, error)
	SetClipSpeed(ctx context.Context, trackType string, trackIndex, clipIndex int, speed float64, ripple bool) (*GenericResult, error)
	ReverseClip(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	SetClipInPoint(ctx context.Context, trackType string, trackIndex, clipIndex int, seconds float64) (*GenericResult, error)
	SetClipOutPoint(ctx context.Context, trackType string, trackIndex, clipIndex int, seconds float64) (*GenericResult, error)
	GetClipSpeed(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	TrimClipStart(ctx context.Context, trackType string, trackIndex, clipIndex int, newStartTime float64) (*GenericResult, error)
	TrimClipEnd(ctx context.Context, trackType string, trackIndex, clipIndex int, newEndTime float64) (*GenericResult, error)
	ExtendClipToPlayhead(ctx context.Context, trackType string, trackIndex, clipIndex int, trimEnd bool) (*GenericResult, error)
	CreateSubclip(ctx context.Context, projectItemIndex int, name string, inPoint, outPoint float64) (*GenericResult, error)
	SelectClip(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	DeselectAll(ctx context.Context) (*GenericResult, error)
	GetSelectedClips(ctx context.Context) (*GenericResult, error)
	LinkClips(ctx context.Context, clipPairsJSON string) (*GenericResult, error)
	UnlinkClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	GetLinkedClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Effects & Transitions ---
	AddTransition(ctx context.Context, params *TransitionParams) error
	AddText(ctx context.Context, params *TextParams) (*ClipResult, error)

	// --- Audio ---
	SetAudioLevel(ctx context.Context, clipID, sequenceID string, levelDB float64) error

	// --- Audio & Track Management (extended) ---
	EvalAudioCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error)

	// --- Export ---
	Export(ctx context.Context, params *ExportParams) (*ExportResult, error)

	// --- Export & Render (Extended) ---
	ExportDirect(ctx context.Context, params *ExportDirectParams) (*GenericExportResult, error)
	ExportViaAME(ctx context.Context, params *ExportViaAMEParams) (*GenericExportResult, error)
	ExportFrame(ctx context.Context, params *ExportFrameParams) (*GenericExportResult, error)
	ExportAAF(ctx context.Context, params *ExportAAFParams) (*GenericExportResult, error)
	ExportOMF(ctx context.Context, params *ExportOMFParams) (*GenericExportResult, error)
	ExportFCPXML(ctx context.Context, outputPath string) (*GenericExportResult, error)
	ExportProjectAsXML(ctx context.Context, outputPath string) (*GenericExportResult, error)
	GetExporters(ctx context.Context) (*ExporterListResult, error)
	GetExportPresets(ctx context.Context, exporterIndex int) (*ExportPresetListResult, error)
	StartAMEBatch(ctx context.Context) (*GenericExportResult, error)
	LaunchAME(ctx context.Context) (*GenericExportResult, error)
	ExportAudioOnly(ctx context.Context, params *ExportAudioOnlyParams) (*GenericExportResult, error)
	GetExportProgress(ctx context.Context) (*ExportProgressResult, error)
	RenderSequencePreview(ctx context.Context, params *RenderPreviewParams) (*GenericExportResult, error)

	// --- Media Analysis (Rust engine) ---
	ScanAssets(ctx context.Context, dir string, recursive bool, extensions []string) (*ScanResult, error)

	// --- Intelligence (Python service) ---
	ParseScript(ctx context.Context, text string, filePath string, format string) (*ParsedScript, error)

	// --- Composite Workflows ---
	AutoEdit(ctx context.Context, params *AutoEditParams) (*AutoEditResult, error)
}
