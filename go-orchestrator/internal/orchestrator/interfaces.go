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

	// EvalImmersiveCommand runs a VR/360/HDR/advanced-format ExtendScript function.
	EvalImmersiveCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error)
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

	// --- Immersive / VR / HDR / Advanced Formats ---
	EvalImmersiveCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error)

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

	// --- Clip/Item Metadata ---
	GetClipMetadata(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	SetClipMetadata(ctx context.Context, projectItemIndex int, field, value string) (*GenericResult, error)
	AddCustomMetadataField(ctx context.Context, fieldName, fieldLabel string, fieldType int) (*GenericResult, error)
	GetMetadataSchema(ctx context.Context) (*GenericResult, error)
	BatchSetMetadata(ctx context.Context, itemIndices []int, field, value string) (*GenericResult, error)

	// --- Labels & Colors ---
	GetAvailableLabelColors(ctx context.Context) (*GenericResult, error)
	SetClipLabelByName(ctx context.Context, projectItemIndex int, colorName string) (*GenericResult, error)
	GetLabelColorForClip(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	BatchSetLabels(ctx context.Context, itemIndices []int, colorIndex int) (*GenericResult, error)
	FilterByLabel(ctx context.Context, colorIndex int) (*GenericResult, error)

	// --- Footage Interpretation ---
	GetFootageInterpretation(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	SetFootageFrameRate(ctx context.Context, projectItemIndex int, fps float64) (*GenericResult, error)
	SetFootageFieldOrder(ctx context.Context, projectItemIndex int, fieldOrder int) (*GenericResult, error)
	SetFootageAlphaChannel(ctx context.Context, projectItemIndex int, alphaType int) (*GenericResult, error)
	SetFootagePixelAspectRatio(ctx context.Context, projectItemIndex int, num, den float64) (*GenericResult, error)
	ResetFootageInterpretation(ctx context.Context, projectItemIndex int) (*GenericResult, error)

	// --- Media Info ---
	GetMediaInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	GetMediaPath(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	RevealInFinder(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	RefreshMedia(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	ReplaceMedia(ctx context.Context, projectItemIndex int, newFilePath string) (*GenericResult, error)
	DuplicateProjectItem(ctx context.Context, projectItemIndex int) (*GenericResult, error)

	// --- Smart Bins ---
	CreateSmartBin(ctx context.Context, name, searchQuery string) (*GenericResult, error)
	GetSmartBinResults(ctx context.Context, binPath string) (*GenericResult, error)

	// --- Clip Usage ---
	GetClipUsageInSequences(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	GetUnusedClips(ctx context.Context) (*GenericResult, error)
	GetUsedClips(ctx context.Context) (*GenericResult, error)
	GetClipUsageCount(ctx context.Context, projectItemIndex int) (*GenericResult, error)

	// --- File Management ---
	GetProjectFileSize(ctx context.Context) (*GenericResult, error)
	GetMediaDiskUsage(ctx context.Context) (*GenericResult, error)

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

	// --- Batch Import ---
	BatchImportWithMetadata(ctx context.Context, itemsJSON string) (*GenericResult, error)
	ImportImageSequence(ctx context.Context, folderPath string, fps float64, targetBin string) (*GenericResult, error)

	// --- Batch Export ---
	BatchExportSequences(ctx context.Context, sequenceIndices []int, outputDir, presetPath string) (*GenericResult, error)
	ExportAllSequences(ctx context.Context, outputDir, presetPath string) (*GenericResult, error)

	// --- Batch Effects ---
	ApplyEffectToMultipleClips(ctx context.Context, trackType string, trackIndex int, clipIndices []int, effectName string) (*GenericResult, error)
	RemoveAllEffects(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	ApplyTransitionToAllCuts(ctx context.Context, trackIndex int, transitionName string, duration float64) (*GenericResult, error)

	// --- Batch Color ---
	ApplyLUTToAllClips(ctx context.Context, trackIndex int, lutPath string) (*GenericResult, error)
	ResetColorOnAllClips(ctx context.Context, trackIndex int) (*GenericResult, error)

	// --- Batch Audio ---
	NormalizeAllAudio(ctx context.Context, targetDB float64) (*GenericResult, error)
	MuteAllAudioTracks(ctx context.Context) (*GenericResult, error)
	UnmuteAllAudioTracks(ctx context.Context) (*GenericResult, error)

	// --- Conforming ---
	ConformSequenceToClip(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	ScaleAllClipsToFrame(ctx context.Context) (*GenericResult, error)

	// --- Timeline Operations (Batch) ---
	SelectAllClipsOnTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	SelectAllClipsBetween(ctx context.Context, startSeconds, endSeconds float64) (*GenericResult, error)
	DeleteAllClipsBetween(ctx context.Context, trackType string, trackIndex int, startSeconds, endSeconds float64) (*GenericResult, error)
	RippleDeleteAllGaps(ctx context.Context) (*GenericResult, error)

	// --- Project Cleanup ---
	RemoveUnusedMedia(ctx context.Context) (*GenericResult, error)
	GetUnusedMedia(ctx context.Context) (*GenericResult, error)
	FlattenAllBins(ctx context.Context) (*GenericResult, error)
	AutoOrganizeBins(ctx context.Context) (*GenericResult, error)

	// --- Markers Batch ---
	ExportMarkersAsCSV(ctx context.Context, outputPath string) (*GenericResult, error)
	ExportMarkersAsEDL(ctx context.Context, outputPath string) (*GenericResult, error)
	ImportMarkersFromCSV(ctx context.Context, csvPath string) (*GenericResult, error)
	DeleteAllMarkers(ctx context.Context) (*GenericResult, error)
	ConvertMarkersToClips(ctx context.Context, markerColor string) (*GenericResult, error)

	// --- Automation ---
	RunExtendScript(ctx context.Context, script string) (*GenericResult, error)
	GetSystemInfo(ctx context.Context) (*GenericResult, error)
	GetRecentProjects(ctx context.Context) (*GenericResult, error)

	// --- Sequence Presets ---
	ListSequencePresets(ctx context.Context) (*GenericResult, error)
	CreateSequenceFromPreset(ctx context.Context, name, presetPath string) (*GenericResult, error)
	ExportSequencePreset(ctx context.Context, sequenceIndex int, outputPath string) (*GenericResult, error)

	// --- Effect Presets ---
	ListEffectPresets(ctx context.Context) (*GenericResult, error)
	ApplyEffectPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, presetPath string) (*GenericResult, error)
	SaveEffectPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, presetName string) (*GenericResult, error)

	// --- Export Presets (disk) ---
	ListExportPresetsFromDisk(ctx context.Context) (*GenericResult, error)
	CreateExportPreset(ctx context.Context, settingsJSON, name string) (*GenericResult, error)
	GetExportPresetDetails(ctx context.Context, presetPath string) (*GenericResult, error)

	// --- Project Templates ---
	SaveAsTemplate(ctx context.Context, templatePath string) (*GenericResult, error)
	CreateFromTemplate(ctx context.Context, templatePath, projectPath string) (*GenericResult, error)

	// --- Keyboard Shortcuts ---
	GetKeyboardShortcuts(ctx context.Context) (*GenericResult, error)
	ExecuteMenuCommand(ctx context.Context, menuPath string) (*GenericResult, error)

	// --- Workflow / Ingest Presets ---
	CreateIngestPreset(ctx context.Context, name, settingsJSON string) (*GenericResult, error)
	GetIngestSettings(ctx context.Context) (*GenericResult, error)
	SetIngestSettings(ctx context.Context, enabled bool, preset string) (*GenericResult, error)

	// --- Clip Presets ---
	SaveClipPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, name string) (*GenericResult, error)
	ApplyClipPreset(ctx context.Context, trackType string, trackIndex, clipIndex int, presetName string) (*GenericResult, error)
	ListClipPresets(ctx context.Context) (*GenericResult, error)

	// --- Batch Operations (extended) ---
	BatchRename(ctx context.Context, trackType string, trackIndex int, pattern string, startNumber int) (*GenericResult, error)
	BatchSetDuration(ctx context.Context, trackType string, trackIndex int, durationSeconds float64) (*GenericResult, error)
	BatchSetSpeed(ctx context.Context, trackType string, trackIndex int, speed float64) (*GenericResult, error)
	BatchApplyTransitions(ctx context.Context, trackIndex int, transitionName string, duration float64) (*GenericResult, error)
	BatchExportFrames(ctx context.Context, trackIndex int, outputDir, format string) (*GenericResult, error)

	// --- Timeline Templates ---
	SaveTimelineTemplate(ctx context.Context, name, description string) (*GenericResult, error)
	ApplyTimelineTemplate(ctx context.Context, templateName string) (*GenericResult, error)
	ListTimelineTemplates(ctx context.Context) (*GenericResult, error)

	// --- Macro Recording ---
	StartMacroRecording(ctx context.Context, name string) (*GenericResult, error)
	StopMacroRecording(ctx context.Context) (*GenericResult, error)
	PlayMacro(ctx context.Context, name string) (*GenericResult, error)

	// --- Preferences & Settings ---
	GetGeneralPreferences(ctx context.Context) (*GenericResult, error)
	SetDefaultStillDuration(ctx context.Context, frames int) (*GenericResult, error)
	SetDefaultTransitionDuration(ctx context.Context, seconds float64) (*GenericResult, error)
	SetDefaultAudioTransitionDuration(ctx context.Context, seconds float64) (*GenericResult, error)
	GetBrightness(ctx context.Context) (*GenericResult, error)
	SetBrightness(ctx context.Context, level int) (*GenericResult, error)
	SetAutoSaveEnabled(ctx context.Context, enabled bool) (*GenericResult, error)
	SetAutoSaveMaxVersions(ctx context.Context, count int) (*GenericResult, error)
	GetAutoSaveLocation(ctx context.Context) (*GenericResult, error)
	GetPlaybackResolution(ctx context.Context) (*GenericResult, error)
	SetPlaybackResolution(ctx context.Context, quality string) (*GenericResult, error)
	GetPrerollFrames(ctx context.Context) (*GenericResult, error)
	SetPrerollFrames(ctx context.Context, frames int) (*GenericResult, error)
	GetPostrollFrames(ctx context.Context) (*GenericResult, error)
	SetPostrollFrames(ctx context.Context, frames int) (*GenericResult, error)
	GetTimelineSettings(ctx context.Context) (*GenericResult, error)
	SetTimeDisplayFormat(ctx context.Context, format string) (*GenericResult, error)
	SetVideoTransitionDefaultDuration(ctx context.Context, frames int) (*GenericResult, error)
	GetMediaCacheSettings(ctx context.Context) (*GenericResult, error)
	SetMediaCacheLocation(ctx context.Context, path string) (*GenericResult, error)
	SetMediaCacheSize(ctx context.Context, maxGB float64) (*GenericResult, error)
	CleanMediaCacheOlderThan(ctx context.Context, days int) (*GenericResult, error)
	GetLabelColorNames(ctx context.Context) (*GenericResult, error)
	SetLabelColorName(ctx context.Context, index int, name string) (*GenericResult, error)
	GetRendererInfo(ctx context.Context) (*GenericResult, error)
	GetGPUInfo(ctx context.Context) (*GenericResult, error)
	SetRenderer(ctx context.Context, rendererName string) (*GenericResult, error)
	GetDefaultSequencePresets(ctx context.Context) (*GenericResult, error)
	SetDefaultSequencePreset(ctx context.Context, presetPath string) (*GenericResult, error)
	GetInstalledCodecs(ctx context.Context) (*GenericResult, error)

	// --- Review & Collaboration ---
	AddReviewComment(ctx context.Context, time float64, text, author string) (*GenericResult, error)
	GetReviewComments(ctx context.Context) (*GenericResult, error)
	ResolveReviewComment(ctx context.Context, markerIndex int) (*GenericResult, error)
	GetUnresolvedComments(ctx context.Context) (*GenericResult, error)
	ExportReviewReport(ctx context.Context, outputPath, format string) (*GenericResult, error)

	// --- Version Control ---
	GetProjectVersionHistory(ctx context.Context) (*GenericResult, error)
	RevertToVersion(ctx context.Context, versionPath string) (*GenericResult, error)
	CreateSnapshot(ctx context.Context, name, description string) (*GenericResult, error)
	CompareSnapshots(ctx context.Context, snapshot1, snapshot2 string) (*GenericResult, error)

	// --- EDL/XML Interchange ---
	ImportEDL(ctx context.Context, edlPath string) (*GenericResult, error)
	ImportAAF(ctx context.Context, aafPath string) (*GenericResult, error)
	ImportFCPXML(ctx context.Context, xmlPath string) (*GenericResult, error)
	ImportXMLTimeline(ctx context.Context, xmlPath string) (*GenericResult, error)
	ExportEDLFile(ctx context.Context, outputPath, format string) (*GenericResult, error)
	ExportProjectSnapshot(ctx context.Context, outputPath string) (*GenericResult, error)

	// --- Collaboration Metadata ---
	SetEditorialNote(ctx context.Context, trackType string, trackIndex, clipIndex int, note string) (*GenericResult, error)
	GetEditorialNotes(ctx context.Context) (*GenericResult, error)
	ClearEditorialNotes(ctx context.Context) (*GenericResult, error)
	TagClipForReview(ctx context.Context, trackType string, trackIndex, clipIndex int, reviewType string) (*GenericResult, error)
	GetClipReviewStatus(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Change Tracking ---
	GetSequenceChangeLog(ctx context.Context) (*GenericResult, error)
	GetProjectActivity(ctx context.Context) (*GenericResult, error)
	GetLastModifiedClips(ctx context.Context, count int) (*GenericResult, error)

	// --- Delivery Checklist ---
	CheckAudioLevels(ctx context.Context, targetLUFS, tolerance float64) (*GenericResult, error)
	CheckFrameRate(ctx context.Context, targetFPS float64) (*GenericResult, error)
	CheckResolution(ctx context.Context, targetWidth, targetHeight int) (*GenericResult, error)
	CheckDuration(ctx context.Context, minSeconds, maxSeconds float64) (*GenericResult, error)
	GenerateDeliveryReport(ctx context.Context, specsJSON string) (*GenericResult, error)
	CheckForBlackFrames(ctx context.Context, thresholdFrames int) (*GenericResult, error)
	CheckForFlashContent(ctx context.Context, threshold float64) (*GenericResult, error)

	// --- Essential Graphics Panel ---
	GetEssentialGraphicsComponents(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	SetEssentialGraphicsProperty(ctx context.Context, trackIndex, clipIndex int, propName, value string) (*GenericResult, error)
	GetEssentialGraphicsText(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	ReplaceAllText(ctx context.Context, trackIndex, clipIndex int, searchText, replaceText string) (*GenericResult, error)

	// --- MOGRT Management (extended) ---
	ListInstalledMOGRTs(ctx context.Context) (*GenericResult, error)
	GetMOGRTInfo(ctx context.Context, mogrtPath string) (*GenericResult, error)
	BatchUpdateMOGRTs(ctx context.Context, trackIndex int, propertyName, value string) (*GenericResult, error)
	CreateMOGRTFromClip(ctx context.Context, trackIndex, clipIndex int, outputPath string) (*GenericResult, error)

	// --- Text Operations ---
	AddScrollingTitle(ctx context.Context, text string, trackIndex int, startTime, duration, speed float64) (*GenericResult, error)
	AddTypewriterText(ctx context.Context, text string, trackIndex int, startTime, duration, typeSpeed float64) (*GenericResult, error)
	AddTextWithBackground(ctx context.Context, text string, trackIndex int, startTime, duration float64, bgColor string, padding int) (*GenericResult, error)
	SetTextAnimation(ctx context.Context, trackIndex, clipIndex int, animationType string, duration float64) (*GenericResult, error)

	// --- Shape Layers ---
	AddRectangle(ctx context.Context, trackIndex int, startTime, duration, x, y float64, width, height int, color string, borderWidth int) (*GenericResult, error)
	AddCircle(ctx context.Context, trackIndex int, startTime, duration, x, y float64, radius int, color string) (*GenericResult, error)
	AddLine(ctx context.Context, trackIndex int, startTime, duration, x1, y1, x2, y2 float64, color string, thickness int) (*GenericResult, error)

	// --- Countdown / Timers ---
	AddCountdown(ctx context.Context, trackIndex int, startTime float64, fromSeconds int, style string) (*GenericResult, error)
	AddTimecode(ctx context.Context, trackIndex int, startTime, duration float64, format string) (*GenericResult, error)

	// --- Watermark ---
	AddWatermark(ctx context.Context, imagePath, position string, opacity, scale float64) (*GenericResult, error)
	AddTextWatermark(ctx context.Context, text, position string, opacity float64, fontSize int, color string) (*GenericResult, error)
	RemoveWatermark(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Picture Layouts ---
	CreateSplitScreen(ctx context.Context, layout, clipRefsJSON string) (*GenericResult, error)
	CreateCollage(ctx context.Context, clipRefsJSON string, rows, cols, gap int) (*GenericResult, error)

	// --- Animated Transitions (custom) ---
	AddWipeTransition(ctx context.Context, trackIndex, clipIndex int, direction, color string, duration float64) (*GenericResult, error)
	AddZoomTransition(ctx context.Context, trackIndex, clipIndex int, zoomIn bool, duration float64) (*GenericResult, error)
	AddGlitchTransition(ctx context.Context, trackIndex, clipIndex int, intensity, duration float64) (*GenericResult, error)

	// --- Subtitling (extended) ---
	AutoGenerateSubtitles(ctx context.Context, language, style string) (*GenericResult, error)
	TranslateSubtitles(ctx context.Context, trackIndex int, targetLanguage string) (*GenericResult, error)
	FormatSubtitles(ctx context.Context, trackIndex, maxCharsPerLine, maxLines int) (*GenericResult, error)
	BurnInSubtitles(ctx context.Context, trackIndex int) (*GenericResult, error)
	AdjustSubtitleTiming(ctx context.Context, trackIndex int, offsetSeconds float64) (*GenericResult, error)

	// --- After Effects Integration ---
	SendToAfterEffects(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	ImportAEComp(ctx context.Context, aepPath, compName, targetBin string) (*GenericResult, error)
	ImportAllAEComps(ctx context.Context, aepPath, targetBin string) (*GenericResult, error)
	RefreshAEComp(ctx context.Context, projectItemIndex int) (*GenericResult, error)

	// --- Photoshop Integration ---
	EditInPhotoshop(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	ImportPSDLayers(ctx context.Context, psdPath, targetBin string, asSequence bool) (*GenericResult, error)

	// --- Audition Integration ---
	EditInAudition(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)
	RefreshAuditionEdit(ctx context.Context, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Media Encoder Integration ---
	QueueInMediaEncoder(ctx context.Context, sequenceIndex int, presetPath string) (*GenericResult, error)
	GetMediaEncoderQueue(ctx context.Context) (*GenericResult, error)
	ClearMediaEncoderQueue(ctx context.Context) (*GenericResult, error)

	// --- Dynamic Link ---
	GetDynamicLinkStatus(ctx context.Context) (*GenericResult, error)
	RefreshAllDynamicLinks(ctx context.Context) (*GenericResult, error)

	// --- File Format Support / Codec ---
	GetCodecInfo(ctx context.Context, projectItemIndex int) (*GenericResult, error)
	TranscodeClip(ctx context.Context, projectItemIndex int, outputPath, presetPath string) (*GenericResult, error)
	ConformMedia(ctx context.Context, projectItemIndex int, targetFps float64, targetCodec string) (*GenericResult, error)

	// --- Project Interchange (OMF/AAF import) ---
	ImportOMFFile(ctx context.Context, omfPath, targetBin string) (*GenericResult, error)
	ImportAAFFile(ctx context.Context, aafPath, targetBin string) (*GenericResult, error)

	// --- Clipboard ---
	CopyToSystemClipboard(ctx context.Context, text string) (*GenericResult, error)
	GetFromSystemClipboard(ctx context.Context) (*GenericResult, error)

	// --- External Tools ---
	OpenInExternalEditor(ctx context.Context, projectItemIndex int, editorPath string) (*GenericResult, error)
	ImportFromExternalSource(ctx context.Context, sourcePath, format string) (*GenericResult, error)

	// --- Team Projects ---
	GetTeamProjectStatus(ctx context.Context) (*GenericResult, error)
	CheckInChanges(ctx context.Context, message string) (*GenericResult, error)
	CheckOutSequence(ctx context.Context, sequenceIndex int) (*GenericResult, error)

	// --- Productions ---
	GetProductionInfo(ctx context.Context) (*GenericResult, error)
	ListProductionProjects(ctx context.Context) (*GenericResult, error)
	OpenProductionProject(ctx context.Context, projectName string) (*GenericResult, error)

	// --- Performance Monitoring ---
	GetPerformanceMetrics(ctx context.Context) (*GenericResult, error)
	GetProjectMemoryUsage(ctx context.Context) (*GenericResult, error)
	GetDiskSpace(ctx context.Context, drivePath string) (*GenericResult, error)
	GetOpenProjectCount(ctx context.Context) (*GenericResult, error)
	GetLoadedPlugins(ctx context.Context) (*GenericResult, error)

	// --- Timeline Performance ---
	GetDroppedFrameCount(ctx context.Context) (*GenericResult, error)
	ResetDroppedFrameCount(ctx context.Context) (*GenericResult, error)
	GetTimelineRenderStatus(ctx context.Context, sequenceIndex int) (*GenericResult, error)
	GetEstimatedRenderTime2(ctx context.Context, sequenceIndex int) (*GenericResult, error)
	GetSequenceComplexity(ctx context.Context, sequenceIndex int) (*GenericResult, error)

	// --- Diagnostics ---
	GetPremiereVersion(ctx context.Context) (*GenericResult, error)
	GetInstalledPlugins2(ctx context.Context) (*GenericResult, error)
	GetInstalledEffects2(ctx context.Context) (*GenericResult, error)
	GetInstalledTransitions2(ctx context.Context) (*GenericResult, error)
	CheckProjectIntegrity(ctx context.Context) (*GenericResult, error)

	// --- Error Handling ---
	GetLastError(ctx context.Context) (*GenericResult, error)
	ClearErrors(ctx context.Context) (*GenericResult, error)
	SetErrorLogging(ctx context.Context, enabled bool, logPath string) (*GenericResult, error)
	GetErrorLog(ctx context.Context) (*GenericResult, error)

	// --- Debug Tools ---
	EnableDebugMode(ctx context.Context, enabled bool) (*GenericResult, error)
	GetDebugLog(ctx context.Context) (*GenericResult, error)
	DumpProjectState(ctx context.Context) (*GenericResult, error)
	DumpSequenceState(ctx context.Context, sequenceIndex int) (*GenericResult, error)
	TestBridgeConnection(ctx context.Context) (*GenericResult, error)

	// --- Health Checks ---
	HealthCheck(ctx context.Context) (*GenericResult, error)
	GetServiceStatus(ctx context.Context) (*GenericResult, error)
	GetBridgeLatency(ctx context.Context) (*GenericResult, error)
	GetExtendScriptVersion(ctx context.Context) (*GenericResult, error)

	// --- Cleanup ---
	CleanTempFiles(ctx context.Context) (*GenericResult, error)
	OptimizeProject(ctx context.Context) (*GenericResult, error)

	// --- UI Panel Control ---
	OpenPanel(ctx context.Context, panelName string) (*GenericResult, error)
	ClosePanel(ctx context.Context, panelName string) (*GenericResult, error)
	GetOpenPanels(ctx context.Context) (*GenericResult, error)
	ResetPanelLayout(ctx context.Context) (*GenericResult, error)
	MaximizePanel(ctx context.Context, panelName string) (*GenericResult, error)

	// --- Window Management ---
	GetWindowInfo(ctx context.Context) (*GenericResult, error)
	SetWindowSize(ctx context.Context, width, height int) (*GenericResult, error)
	MinimizeWindow(ctx context.Context) (*GenericResult, error)
	BringToFront(ctx context.Context) (*GenericResult, error)
	EnterFullscreen(ctx context.Context) (*GenericResult, error)

	// --- Timeline UI ---
	SetTrackHeight(ctx context.Context, trackType string, trackIndex int, height int) (*GenericResult, error)
	CollapseTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	ExpandTrack(ctx context.Context, trackType string, trackIndex int) (*GenericResult, error)
	CollapseAllTracks(ctx context.Context) (*GenericResult, error)
	ExpandAllTracks(ctx context.Context) (*GenericResult, error)

	// --- Label Management ---
	SetLabelPreferences(ctx context.Context, labelsJSON string) (*GenericResult, error)
	GetActiveLabelFilter(ctx context.Context) (*GenericResult, error)
	SetLabelFilter(ctx context.Context, colorIndex int) (*GenericResult, error)
	ClearLabelFilter(ctx context.Context) (*GenericResult, error)

	// --- Timeline Display ---
	SetAudioWaveformDisplay(ctx context.Context, enabled bool) (*GenericResult, error)
	SetVideoThumbnailDisplay(ctx context.Context, enabled bool) (*GenericResult, error)
	SetTrackNameDisplay(ctx context.Context, enabled bool) (*GenericResult, error)

	// --- User Feedback ---
	ShowAlert(ctx context.Context, title, message string) (*GenericResult, error)
	ShowConfirmDialog(ctx context.Context, title, message string) (*GenericResult, error)
	ShowInputDialog(ctx context.Context, title, prompt, defaultValue string) (*GenericResult, error)
	ShowProgressDialog(ctx context.Context, title, message string, progress float64) (*GenericResult, error)
	WriteToConsole(ctx context.Context, message string) (*GenericResult, error)

	// --- Accessibility ---
	GetUIScaling(ctx context.Context) (*GenericResult, error)
	SetHighContrastMode(ctx context.Context, enabled bool) (*GenericResult, error)

	// --- Compound / Multi-Step Editing ---
	CreateMontage(ctx context.Context, clipIndices []int, transitionName string, transitionDuration float64, musicPath string) (*GenericResult, error)
	CreateSlideshow(ctx context.Context, imageIndices []int, slideDuration float64, transitionName string, musicPath string) (*GenericResult, error)
	CreateHighlightReel(ctx context.Context, sequenceIndex int, markerColor string, outputName string) (*GenericResult, error)
	RippleDeleteEmptySpaces(ctx context.Context) (*GenericResult, error)
	AlignAllClipsToTrack(ctx context.Context, sourceTrack, destTrack int) (*GenericResult, error)

	// --- Audio-Visual Sync ---
	SyncAllAudioToVideo(ctx context.Context) (*GenericResult, error)
	ReplaceAudio(ctx context.Context, videoTrackIndex, videoClipIndex int, audioPath string) (*GenericResult, error)
	AddMusicBed(ctx context.Context, audioPath string, trackIndex int, startTime, endTime, fadeIn, fadeOut, volume float64) (*GenericResult, error)
	DuckMusicUnderDialogue(ctx context.Context, musicTrackIndex, dialogueTrackIndex int, duckAmount float64) (*GenericResult, error)
	AddSoundEffect(ctx context.Context, sfxPath string, trackIndex int, time, volume float64) (*GenericResult, error)

	// --- Color Workflow ---
	MatchColorBetweenClips(ctx context.Context, srcTrackIndex, srcClipIndex, destTrackIndex, destClipIndex int) (*GenericResult, error)
	ApplyColorPreset(ctx context.Context, trackIndex, clipIndex int, presetName string) (*GenericResult, error)
	CreateColorGradient(ctx context.Context, trackIndex, startClipIndex, endClipIndex int, startTemp, endTemp float64) (*GenericResult, error)
	AutoCorrectAllClips(ctx context.Context, trackIndex int) (*GenericResult, error)

	// --- Text Workflow ---
	AddSubtitlesFromSRT(ctx context.Context, srtPath string, trackIndex int) (*GenericResult, error)
	AddEndCredits(ctx context.Context, creditsJSON string, trackIndex int, scrollDuration float64, style string) (*GenericResult, error)
	AddChapterMarkers(ctx context.Context, chaptersJSON string) (*GenericResult, error)
	GenerateChaptersFromMarkers(ctx context.Context, outputPath string) (*GenericResult, error)

	// --- Export Workflow ---
	ExportForYouTube(ctx context.Context, outputPath, title, description string) (*GenericResult, error)
	ExportForInstagram(ctx context.Context, outputPath, aspectRatio string) (*GenericResult, error)
	ExportForTikTok(ctx context.Context, outputPath string) (*GenericResult, error)
	ExportForTwitter(ctx context.Context, outputPath string) (*GenericResult, error)
	ExportMultipleFormats(ctx context.Context, outputDir string, formats []string) (*GenericResult, error)

	// --- Project Setup ---
	SetupNewProject(ctx context.Context, name, path, resolution string, fps float64, audioSampleRate int) (*GenericResult, error)
	SetupEditingWorkspace(ctx context.Context, projectPath, mediaFolder, sequenceName string) (*GenericResult, error)
	ImportAndOrganize(ctx context.Context, mediaFolder string, autoCreateBins bool) (*GenericResult, error)
	PrepareForDelivery(ctx context.Context, specsJSON string) (*GenericResult, error)

	// --- Cleanup Workflow ---
	ArchiveProject(ctx context.Context, outputPath string, includeMedia, includeRenders bool) (*GenericResult, error)
	TrimProject(ctx context.Context) (*GenericResult, error)
	ConsolidateAndTranscode(ctx context.Context, outputDir, codec, quality string) (*GenericResult, error)

	// --- Event Monitoring ---
	RegisterEventListener(ctx context.Context, eventName string) (*GenericResult, error)
	UnregisterEventListener(ctx context.Context, eventName string) (*GenericResult, error)
	GetRegisteredEvents(ctx context.Context) (*GenericResult, error)
	GetEventHistory(ctx context.Context, count int) (*GenericResult, error)
	ClearEventHistory(ctx context.Context) (*GenericResult, error)

	// --- State Watching ---
	WatchPlayheadPosition(ctx context.Context, intervalMs int) (*GenericResult, error)
	StopWatchPlayhead(ctx context.Context) (*GenericResult, error)
	WatchRenderProgress(ctx context.Context, intervalMs int) (*GenericResult, error)
	StopWatchRender(ctx context.Context) (*GenericResult, error)
	GetStateSnapshot(ctx context.Context) (*GenericResult, error)

	// --- Project State (Monitoring) ---
	IsProjectModified(ctx context.Context) (*GenericResult, error)
	GetProjectDuration(ctx context.Context) (*GenericResult, error)
	GetProjectStats(ctx context.Context) (*GenericResult, error)
	GetRecentActions(ctx context.Context, count int) (*GenericResult, error)

	// --- Sequence State (Extended Monitoring) ---
	GetActiveTrackTargets(ctx context.Context) (*GenericResult, error)
	SetActiveTrackTargets(ctx context.Context, videoTargets, audioTargets string) (*GenericResult, error)
	GetTrackHeights(ctx context.Context) (*GenericResult, error)
	SetTrackHeights(ctx context.Context, trackType, heights string) (*GenericResult, error)
	IsSequenceModified(ctx context.Context) (*GenericResult, error)
	GetSequenceHash(ctx context.Context) (*GenericResult, error)

	// --- Clip State (Monitoring) ---
	GetClipUnderPlayhead(ctx context.Context) (*GenericResult, error)
	GetClipAtTime(ctx context.Context, trackType string, trackIndex int, seconds float64) (*GenericResult, error)
	GetAdjacentClips(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	IsClipSelected(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)
	GetClipProperties(ctx context.Context, trackType string, trackIndex, clipIndex int) (*GenericResult, error)

	// --- Notifications ---
	ShowNotification(ctx context.Context, title, message string) (*GenericResult, error)
	LogToEventsPanel(ctx context.Context, message, level string) (*GenericResult, error)
	ShowProgressBar(ctx context.Context, title string, current, total int) (*GenericResult, error)
	HideProgressBar(ctx context.Context) (*GenericResult, error)
	ShowDialog(ctx context.Context, title, message, buttons string) (*GenericResult, error)
}
