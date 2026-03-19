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

	// --- Effects & Transitions (Extended) ---
	AddVideoTransition(ctx context.Context, trackIndex, clipIndex int, transitionName string, duration float64, applyToEnd bool) (*GenericResult, error)
	AddAudioTransition(ctx context.Context, trackIndex, clipIndex int, transitionName string, duration float64) (*GenericResult, error)
	RemoveTransition(ctx context.Context, trackType string, trackIndex, transitionIndex int) (*GenericResult, error)
	GetTransitions(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	SetDefaultVideoTransition(ctx context.Context, transitionName string) (*GenericResult, error)
	SetDefaultAudioTransition(ctx context.Context, transitionName string) (*GenericResult, error)
	ApplyDefaultTransition(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	GetAvailableTransitions(ctx context.Context) (*GenericResult, error)
	ApplyVideoEffect(ctx context.Context, trackIndex, clipIndex int, effectName string) (*GenericResult, error)
	RemoveVideoEffect(ctx context.Context, trackIndex, clipIndex, effectIndex int) (*GenericResult, error)
	GetClipEffects(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	SetEffectParameter(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, value float64) (*GenericResult, error)
	GetEffectParameter(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int) (*GenericResult, error)
	EnableEffect(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex int, enabled bool) (*GenericResult, error)
	CopyEffects(ctx context.Context, srcTrackType string, srcTrackIndex, srcClipIndex int) (*GenericResult, error)
	PasteEffects(ctx context.Context, destTrackType string, destTrackIndex, destClipIndex int) (*GenericResult, error)
	SetPosition(ctx context.Context, trackIndex, clipIndex int, x, y float64) (*GenericResult, error)
	SetScale(ctx context.Context, trackIndex, clipIndex int, scale float64) (*GenericResult, error)
	SetRotation(ctx context.Context, trackIndex, clipIndex int, degrees float64) (*GenericResult, error)
	SetAnchorPoint(ctx context.Context, trackIndex, clipIndex int, x, y float64) (*GenericResult, error)
	SetOpacity(ctx context.Context, trackIndex, clipIndex int, opacity float64) (*GenericResult, error)
	GetMotionProperties(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	SetBlendMode(ctx context.Context, trackIndex, clipIndex int, mode string) (*GenericResult, error)
	CreateAdjustmentLayer(ctx context.Context, name string, width, height int, duration float64) (*GenericResult, error)
	PlaceAdjustmentLayer(ctx context.Context, projectItemIndex, trackIndex int, startTime, duration float64) (*GenericResult, error)
	AddKeyframe(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, time, value float64) (*GenericResult, error)
	DeleteKeyframe(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, time float64) (*GenericResult, error)
	SetKeyframeInterpolation(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, time float64, interpType string) (*GenericResult, error)
	GetKeyframes(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int) (*GenericResult, error)
	SetTimeVarying(ctx context.Context, trackType string, trackIndex, clipIndex, componentIndex, paramIndex int, enabled bool) (*GenericResult, error)
	SetLumetriBrightness(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	SetLumetriContrast(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	SetLumetriSaturation(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	SetLumetriTemperature(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	SetLumetriTint(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	SetLumetriExposure(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)

	// --- Color Correction & Lumetri (Extended) ---
	LumetriGetAll(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	LumetriSetExposure2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetContrast2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetHighlights(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetShadows(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetWhites(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetBlacks(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetTemperature2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetTint2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetSaturation2(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetVibrance(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetFadedFilm(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetSharpen(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetCurvePoint(ctx context.Context, trackIndex, clipIndex int, channel string, inputValue, outputValue float64) (*GenericResult, error)
	LumetriSetShadowColor(ctx context.Context, trackIndex, clipIndex int, hue, saturation, brightness float64) (*GenericResult, error)
	LumetriSetMidtoneColor(ctx context.Context, trackIndex, clipIndex int, hue, saturation, brightness float64) (*GenericResult, error)
	LumetriSetHighlightColor(ctx context.Context, trackIndex, clipIndex int, hue, saturation, brightness float64) (*GenericResult, error)
	LumetriSetVignetteAmount(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetVignetteMidpoint(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetVignetteRoundness(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriSetVignetteFeather(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	LumetriApplyLUT(ctx context.Context, trackIndex, clipIndex int, lutPath string) (*GenericResult, error)
	LumetriRemoveLUT(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	LumetriAutoColor(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	LumetriReset(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	GetColorInfo(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	CopyColorGrade(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	PasteColorGrade(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	ApplyColorGradeToAll(ctx context.Context, srcTrackIndex, srcClipIndex, destTrackIndex int) (*GenericResult, error)
	LumetriAutoWhiteBalance(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Graphics, Titles, Captions ---
	ImportMOGRT(ctx context.Context, mogrtPath, timeTicks string, videoTrackOffset, audioTrackOffset int) (*GenericResult, error)
	GetMOGRTProperties(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	SetMOGRTText(ctx context.Context, trackIndex, clipIndex, propertyIndex int, text string) (*GenericResult, error)
	SetMOGRTProperty(ctx context.Context, trackIndex, clipIndex int, propertyName string, value string) (*GenericResult, error)
	AddTitle(ctx context.Context, text string, trackIndex int, startTime, duration float64, styleJSON string) (*GenericResult, error)
	AddLowerThird(ctx context.Context, name, title string, trackIndex int, startTime, duration float64) (*GenericResult, error)
	CreateCaptionTrack(ctx context.Context, format string) (*GenericResult, error)
	ImportCaptions(ctx context.Context, filePath, format string) (*GenericResult, error)
	GetCaptions(ctx context.Context, trackIndex int) (*GenericResult, error)
	AddCaption(ctx context.Context, trackIndex int, startTime, endTime float64, text string) (*GenericResult, error)
	EditCaption(ctx context.Context, trackIndex, captionIndex int, text string) (*GenericResult, error)
	DeleteCaption(ctx context.Context, trackIndex, captionIndex int) (*GenericResult, error)
	ExportCaptions(ctx context.Context, outputPath, format string) (*GenericResult, error)
	StyleCaptions(ctx context.Context, trackIndex int, font string, size float64, color, bgColor, position string) (*GenericResult, error)
	CreateColorMatte(ctx context.Context, name string, red, green, blue, width, height int) (*GenericResult, error)
	PlaceColorMatte(ctx context.Context, projectItemIndex, trackIndex int, startTime, duration float64) (*GenericResult, error)
	CreateTransparentVideo(ctx context.Context, name string, width, height int, duration float64) (*GenericResult, error)
	SetTimeRemapping(ctx context.Context, trackIndex, clipIndex int, enabled bool) (*GenericResult, error)
	AddTimeRemapKeyframe(ctx context.Context, trackIndex, clipIndex int, time, speed float64) (*GenericResult, error)
	FreezeFrame(ctx context.Context, trackIndex, clipIndex int, time, duration float64) (*GenericResult, error)
	DetectSceneEdits(ctx context.Context, trackIndex, clipIndex int, sensitivity float64) (*GenericResult, error)

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

	// --- Multicam ---
	CreateMulticamSequence(ctx context.Context, name string, clipIndices []int, syncPoint string) (*GenericResult, error)
	SwitchMulticamAngle(ctx context.Context, trackIndex int, time float64, angleIndex int) (*GenericResult, error)
	FlattenMulticam(ctx context.Context, sequenceIndex int) (*GenericResult, error)
	GetMulticamAngles(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Proxy Workflow ---
	CreateProxy(ctx context.Context, projectItemIndex int, presetPath string) (*GenericResult, error)
	AttachProxy(ctx context.Context, projectItemIndex int, proxyPath string) (*GenericResult, error)
	HasProxy(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	GetProxyPath(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	ToggleProxies(ctx context.Context, enabled bool) (*GenericResult, error)
	DetachProxy(ctx context.Context, projectItemIndex int) (*GenericResult, error)

	// --- Workspace ---
	GetWorkspaces(ctx context.Context) (*GenericResult, error)
	SetWorkspace(ctx context.Context, name string) (*GenericResult, error)
	SaveWorkspace(ctx context.Context, name string) (*GenericResult, error)

	// --- Undo/Redo ---
	Undo(ctx context.Context) (*GenericResult, error)
	Redo(ctx context.Context) (*GenericResult, error)

	// --- Project Panel ---
	SortProjectPanel(ctx context.Context, field string, ascending bool) (*GenericResult, error)
	SearchProjectPanel(ctx context.Context, query string) (*GenericResult, error)

	// --- Source Monitor ---
	OpenInSourceMonitor(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	GetSourceMonitorPosition(ctx context.Context) (*GenericResult, error)
	SetSourceMonitorPosition(ctx context.Context, seconds float64) (*GenericResult, error)

	// --- Preferences ---
	GetAutoSaveSettings(ctx context.Context) (*GenericResult, error)
	SetAutoSaveInterval(ctx context.Context, minutes int) (*GenericResult, error)
	GetMemorySettings(ctx context.Context) (*GenericResult, error)

	// --- Media Cache ---
	ClearMediaCache(ctx context.Context) (*GenericResult, error)
	GetMediaCachePath(ctx context.Context) (*GenericResult, error)

	// --- Advanced Editing ---
	RippleTrim(ctx context.Context, trackType string, trackIndex, clipIndex int, trimEnd bool, deltaSeconds float64) (*GenericResult, error)
	RollTrim(ctx context.Context, trackType string, trackIndex, clipIndex int, deltaSeconds float64) (*GenericResult, error)
	SlipClip(ctx context.Context, trackType string, trackIndex, clipIndex int, deltaSeconds float64) (*GenericResult, error)
	SlideClip(ctx context.Context, trackType string, trackIndex, clipIndex int, deltaSeconds float64) (*GenericResult, error)
	PasteInsert(ctx context.Context, trackType string, trackIndex int, time float64) (*GenericResult, error)
	PasteAttributes(ctx context.Context, srcTrackType string, srcTrackIndex, srcClipIndex int, destTrackType string, destTrackIndex, destClipIndex int, attributes string) (*GenericResult, error)
	MatchFrame(ctx context.Context) (*GenericResult, error)
	ReverseMatchFrame(ctx context.Context) (*GenericResult, error)
	LiftSelection(ctx context.Context) (*GenericResult, error)
	ExtractSelection(ctx context.Context) (*GenericResult, error)
	FindGaps(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	CloseGap(ctx context.Context, trackType string, trackIndex, gapIndex int) (*GenericResult, error)
	CloseAllGaps(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	RippleDeleteGap(ctx context.Context, trackType string, trackIndex int, startTime, endTime float64) (*GenericResult, error)
	GroupClips(ctx context.Context, clipRefsJSON string) (*GenericResult, error)
	UngroupClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	GetGroupedClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	SetSnapping(ctx context.Context, enabled bool) (*GenericResult, error)
	GetSnapping(ctx context.Context) (*GenericResult, error)
	ZoomToFitTimeline(ctx context.Context) (*GenericResult, error)
	ZoomToSelection(ctx context.Context) (*GenericResult, error)
	SetTimelineZoom(ctx context.Context, level float64) (*GenericResult, error)
	GoToNextEditPoint(ctx context.Context) (*GenericResult, error)
	GoToPreviousEditPoint(ctx context.Context) (*GenericResult, error)
	GoToNextClip(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	GoToPreviousClip(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	GoToSequenceStart(ctx context.Context) (*GenericResult, error)
	GoToSequenceEnd(ctx context.Context) (*GenericResult, error)
	AddClipMarker(ctx context.Context, trackType string, trackIndex, clipIndex int, time float64, name, comment string, colorIndex int) (*GenericResult, error)
	GetClipMarkers(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	DeleteClipMarker(ctx context.Context, trackType string, trackIndex, clipIndex, markerIndex int) (*GenericResult, error)

	// --- Playback Control ---
	Play(ctx context.Context, speed float64) (*GenericResult, error)
	Pause(ctx context.Context) (*GenericResult, error)
	Stop(ctx context.Context) (*GenericResult, error)
	StepForward(ctx context.Context, frames int) (*GenericResult, error)
	StepBackward(ctx context.Context, frames int) (*GenericResult, error)
	ShuttleForward(ctx context.Context, speed float64) (*GenericResult, error)
	ShuttleBackward(ctx context.Context, speed float64) (*GenericResult, error)
	TogglePlayPause(ctx context.Context) (*GenericResult, error)
	PlayInToOut(ctx context.Context) (*GenericResult, error)
	LoopPlayback(ctx context.Context, enabled bool) (*GenericResult, error)

	// --- Program Monitor ---
	GetProgramMonitorZoom(ctx context.Context) (*GenericResult, error)
	SetProgramMonitorZoom(ctx context.Context, percent float64) (*GenericResult, error)
	FitProgramMonitor(ctx context.Context) (*GenericResult, error)
	ToggleSafeMargins(ctx context.Context) (*GenericResult, error)
	GetFrameAtPlayhead(ctx context.Context) (*GenericResult, error)

	// --- Sequence Navigation (extended) ---
	GoToTimecode(ctx context.Context, timecode string) (*GenericResult, error)
	GoToFrame(ctx context.Context, frameNumber int) (*GenericResult, error)
	GetSequenceDuration(ctx context.Context) (*GenericResult, error)
	GetFrameCount(ctx context.Context) (*GenericResult, error)
	GetCurrentTimecode(ctx context.Context) (*GenericResult, error)

	// --- Selection & Focus ---
	SelectClipsInRange(ctx context.Context, startSeconds, endSeconds float64) (*GenericResult, error)
	SelectAllOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	InvertSelection(ctx context.Context) (*GenericResult, error)
	GetSelectionRange(ctx context.Context) (*GenericResult, error)

	// --- Render Status ---
	GetRenderStatus(ctx context.Context) (*GenericResult, error)
	IsRendering(ctx context.Context) (*GenericResult, error)

	// --- Sequence Metadata ---
	GetSequenceMetadata(ctx context.Context) (*GenericResult, error)
	SetSequenceMetadata(ctx context.Context, key, value string) (*GenericResult, error)
	GetSequenceColorSpace(ctx context.Context) (*GenericResult, error)
	SetSequenceColorSpace(ctx context.Context, colorSpace string) (*GenericResult, error)

	// --- Composite Workflows ---
	AutoEdit(ctx context.Context, params *AutoEditParams) (*AutoEditResult, error)

	// --- AI-Powered Smart Editing ---
	SmartCut(ctx context.Context, params *SmartCutParams) (*SmartCutResult, error)
	SmartTrim(ctx context.Context, params *SmartTrimParams) (*SmartTrimResult, error)
	AutoColorMatch(ctx context.Context, params *AutoColorMatchParams) (*AutoColorMatchResult, error)
	AutoAudioLevels(ctx context.Context, params *AutoAudioLevelsParams) (*AutoAudioLevelsResult, error)
	SuggestTransitions(ctx context.Context, sequenceID string) (*SuggestTransitionsResult, error)
	SuggestMusic(ctx context.Context, sequenceID string) (*SuggestMusicResult, error)

	// --- AI-Powered Content Analysis ---
	AnalyzeClip(ctx context.Context, filePath string) (*ClipAnalysis, error)
	AnalyzeSequence(ctx context.Context, sequenceID string) (*SequenceAnalysis, error)
	GetSequenceStatistics(ctx context.Context, sequenceID string) (*SequenceStatistics, error)
	DetectJumpCuts(ctx context.Context, sequenceID string, threshold float64) (*JumpCutResult, error)
	DetectAudioIssues(ctx context.Context, sequenceID string) (*AudioIssuesResult, error)

	// --- AI Script-to-Edit Pipeline (enhanced) ---
	GenerateRoughCut(ctx context.Context, params *RoughCutParams) (*RoughCutResult, error)
	RefineEdit(ctx context.Context, params *RefineEditParams) (*RefineEditResult, error)
	AddBRollSuggestions(ctx context.Context, sequenceID string) (*BRollSuggestionsResult, error)
	GenerateTrailer(ctx context.Context, params *GenerateTrailerParams) (*GenerateTrailerResult, error)
	CreateSocialCuts(ctx context.Context, params *SocialCutParams) (*SocialCutResult, error)

	// --- AI-Powered Smart Organisation ---
	AutoOrganizeProject(ctx context.Context, params *AutoOrganizeParams) (*AutoOrganizeResult, error)
	AITagClips(ctx context.Context, filePath string) (*TagClipsResult, error)
	FindSimilarClips(ctx context.Context, filePath string, maxResults int) (*FindSimilarResult, error)
	SuggestReplacements(ctx context.Context, sequenceID string) (*SuggestReplacementsResult, error)

	// --- AI-Powered Workflow Automation ---
	CreateReviewMarkers(ctx context.Context, sequenceID string) (*ReviewMarkersResult, error)
	GenerateEditSummary(ctx context.Context, sequenceID string) (*EditSummaryResult, error)
	EstimateRenderTime(ctx context.Context, sequenceID string) (*RenderTimeEstimate, error)
	CheckDeliverySpecs(ctx context.Context, sequenceID string, standard string) (*DeliverySpecResult, error)
	CreateProjectReport(ctx context.Context) (*ProjectReportResult, error)

	// --- Transform, Crop, Masking, Stabilization, Blur & Distortion ---
	SetCrop(ctx context.Context, trackIndex, clipIndex int, left, right, top, bottom float64) (*GenericResult, error)
	GetCrop(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	ResetCrop(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	SetUniformScale(ctx context.Context, trackIndex, clipIndex int, enabled bool) (*GenericResult, error)
	GetTransformProperties(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	SetAntiFlicker(ctx context.Context, trackIndex, clipIndex int, value float64) (*GenericResult, error)
	ResetTransform(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	CenterClip(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	FitClipToFrame(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	FillFrame(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	CreatePIP(ctx context.Context, mainTrackIndex, mainClipIndex, pipTrackIndex, pipClipIndex int, position string, scale float64) (*GenericResult, error)
	RemovePIP(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	SetOpacityKeyframes(ctx context.Context, trackIndex, clipIndex int, keyframesJSON string) (*GenericResult, error)
	FadeIn(ctx context.Context, trackIndex, clipIndex int, durationSeconds float64) (*GenericResult, error)
	FadeOut(ctx context.Context, trackIndex, clipIndex int, durationSeconds float64) (*GenericResult, error)
	CrossFadeClips(ctx context.Context, trackIndex, clipIndexA, clipIndexB int, durationSeconds float64) (*GenericResult, error)
	ApplyWarpStabilizer(ctx context.Context, trackIndex, clipIndex int, smoothness float64) (*GenericResult, error)
	GetStabilizationStatus(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	ApplyLensDistortionRemoval(ctx context.Context, trackIndex, clipIndex int, curvature float64) (*GenericResult, error)
	ApplyVideoNoiseReduction(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error)
	ApplyAudioNoiseReduction(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error)
	ApplyDeReverb(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error)
	ApplyDeHum(ctx context.Context, trackIndex, clipIndex int, frequency float64) (*GenericResult, error)
	ApplyGaussianBlur(ctx context.Context, trackIndex, clipIndex int, blurriness float64) (*GenericResult, error)
	ApplyDirectionalBlur(ctx context.Context, trackIndex, clipIndex int, direction, length float64) (*GenericResult, error)
	ApplySharpen(ctx context.Context, trackIndex, clipIndex int, amount float64) (*GenericResult, error)
	ApplyUnsharpMask(ctx context.Context, trackIndex, clipIndex int, amount, radius, threshold float64) (*GenericResult, error)
	ApplyMirror(ctx context.Context, trackIndex, clipIndex int, angle float64, centerX, centerY float64) (*GenericResult, error)
	ApplyCornerPin(ctx context.Context, trackIndex, clipIndex int, cornersJSON string) (*GenericResult, error)
	ApplySpherize(ctx context.Context, trackIndex, clipIndex int, radius, centerX, centerY float64) (*GenericResult, error)
}
