// Package grpc provides client wrappers for the three PremierPro MCP backend services.
//
// PROTO STUB NOTE: This file defines Go-native types that mirror the protobuf messages
// from the proto definitions at proto/definitions/premierpro/. Once proto stubs are
// generated into gen/go/, these types should be replaced with converters that translate
// between proto types and these Go-native types.
//
// Proto packages to import once generated:
//   commonpb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/common/v1"
//   mediapb  "github.com/anthropics/premierpro-mcp/gen/go/premierpro/media/v1"
//   intelpb  "github.com/anthropics/premierpro-mcp/gen/go/premierpro/intelligence/v1"
//   prempb   "github.com/anthropics/premierpro-mcp/gen/go/premierpro/premiere/v1"
package grpc

// ---------------------------------------------------------------------------
// Common types (mirrors premierpro.common.v1)
// ---------------------------------------------------------------------------

// Timecode represents a timecode in HH:MM:SS:FF format.
type Timecode struct {
	Hours     uint32
	Minutes   uint32
	Seconds   uint32
	Frames    uint32
	FrameRate float64
}

// TimeRange is a span defined by in and out points.
type TimeRange struct {
	InPoint  Timecode
	OutPoint Timecode
}

// Resolution represents video width x height.
type Resolution struct {
	Width  uint32
	Height uint32
}

// VideoInfo holds video stream metadata.
type VideoInfo struct {
	Codec           string
	Resolution      Resolution
	FrameRate       float64
	BitrateBps      uint64
	PixelFormat     string
	DurationSeconds float64
}

// AudioInfo holds audio stream metadata.
type AudioInfo struct {
	Codec           string
	SampleRate      uint32
	Channels        uint32
	BitrateBps      uint64
	DurationSeconds float64
}

// AssetType categorises a media asset.
type AssetType int

const (
	AssetTypeUnspecified AssetType = 0
	AssetTypeVideo      AssetType = 1
	AssetTypeAudio      AssetType = 2
	AssetTypeImage      AssetType = 3
	AssetTypeGraphics   AssetType = 4
)

// Asset is a media asset with full metadata.
type Asset struct {
	ID            string
	FilePath      string
	FileName      string
	FileSizeBytes uint64
	MimeType      string
	AssetType     AssetType
	Video         *VideoInfo
	Audio         *AudioInfo
	Metadata      map[string]string
	Fingerprint   string
}

// TrackType distinguishes video and audio tracks.
type TrackType int

const (
	TrackTypeUnspecified TrackType = 0
	TrackTypeVideo      TrackType = 1
	TrackTypeAudio      TrackType = 2
)

// TrackTarget identifies a specific track on the timeline.
type TrackTarget struct {
	Type       TrackType
	TrackIndex uint32
}

// TransitionInfo describes a transition between clips.
type TransitionInfo struct {
	Type            string
	DurationSeconds float64
	Alignment       string
}

// EffectInfo describes an effect applied to a clip.
type EffectInfo struct {
	Name       string
	Parameters map[string]string
}

// EDLEntry is a single entry in an Edit Decision List.
type EDLEntry struct {
	Index         uint32
	SourceAssetID string
	SourceRange   TimeRange
	TimelineRange TimeRange
	Track         TrackTarget
	Transition    *TransitionInfo
	Effects       []EffectInfo
	Notes         string
}

// EditDecisionList is a full EDL.
type EditDecisionList struct {
	ID                string
	Name              string
	SequenceResolution Resolution
	SequenceFrameRate float64
	Entries           []EDLEntry
}

// TextStyle defines a text overlay appearance.
type TextStyle struct {
	FontFamily         string
	FontSize           float64
	ColorHex           string
	Alignment          string
	BackgroundColorHex string
	BackgroundOpacity  float64
	Position           Position
}

// Position is a normalised screen coordinate (0.0-1.0).
type Position struct {
	X float64
	Y float64
}

// OperationStatus represents the state of an async operation.
type OperationStatus int

const (
	OperationStatusUnspecified OperationStatus = 0
	OperationStatusPending    OperationStatus = 1
	OperationStatusRunning    OperationStatus = 2
	OperationStatusCompleted  OperationStatus = 3
	OperationStatusFailed     OperationStatus = 4
)

// ---------------------------------------------------------------------------
// Media engine types (mirrors premierpro.media.v1)
// ---------------------------------------------------------------------------

// ScanAssetsParams are the inputs for the ScanAssets RPC.
type ScanAssetsParams struct {
	Directory  string
	Recursive  bool
	Extensions []string
}

// ScanAssetsResult is the output of the ScanAssets RPC.
type ScanAssetsResult struct {
	Assets              []Asset
	TotalFilesScanned   uint32
	MediaFilesFound     uint32
	ScanDurationSeconds float64
}

// ProbeMediaParams are the inputs for the ProbeMedia RPC.
type ProbeMediaParams struct {
	FilePath string
}

// ProbeMediaResult is the output of the ProbeMedia RPC.
type ProbeMediaResult struct {
	Asset Asset
}

// GenerateThumbnailParams are the inputs for the GenerateThumbnail RPC.
type GenerateThumbnailParams struct {
	FilePath     string
	Timestamp    Timecode
	OutputSize   Resolution
	OutputFormat string // "png" or "jpg"
}

// GenerateThumbnailResult is the output of the GenerateThumbnail RPC.
type GenerateThumbnailResult struct {
	ThumbnailData []byte
	OutputPath    string
	ActualSize    Resolution
}

// AnalyzeWaveformParams are the inputs for the AnalyzeWaveform RPC.
type AnalyzeWaveformParams struct {
	FilePath                  string
	AudioTrack                uint32
	SilenceThresholdDB        float64
	MinSilenceDurationSeconds float64
}

// SilenceRegion is a detected silence span.
type SilenceRegion struct {
	StartSeconds float64
	EndSeconds   float64
	AvgDB        float64
}

// AnalyzeWaveformResult is the output of the AnalyzeWaveform RPC.
type AnalyzeWaveformResult struct {
	SilenceRegions  []SilenceRegion
	PeakDB          float64
	RmsDB           float64
	DurationSeconds float64
	WaveformSamples []float32
}

// DetectScenesParams are the inputs for the DetectScenes RPC.
type DetectScenesParams struct {
	FilePath  string
	Threshold float64
}

// SceneChange is a detected scene boundary.
type SceneChange struct {
	Timecode   Timecode
	Confidence float64
}

// DetectScenesResult is the output of the DetectScenes RPC.
type DetectScenesResult struct {
	Scenes []SceneChange
}

// ---------------------------------------------------------------------------
// Intelligence types (mirrors premierpro.intelligence.v1)
// ---------------------------------------------------------------------------

// SegmentType categorises a script segment.
type SegmentType int

const (
	SegmentTypeUnspecified SegmentType = 0
	SegmentTypeDialogue   SegmentType = 1
	SegmentTypeAction     SegmentType = 2
	SegmentTypeBRoll      SegmentType = 3
	SegmentTypeTransition SegmentType = 4
	SegmentTypeTitle      SegmentType = 5
	SegmentTypeLowerThird SegmentType = 6
	SegmentTypeVoiceover  SegmentType = 7
	SegmentTypeMusic      SegmentType = 8
	SegmentTypeSFX        SegmentType = 9
)

// ScriptSegment is a parsed section of a script.
type ScriptSegment struct {
	Index                    uint32
	Type                     SegmentType
	Content                  string
	Speaker                  string
	SceneDescription         string
	VisualDirection          string
	AudioDirection           string
	EstimatedDurationSeconds float64
	AssetHints               []string
}

// ScriptMetadata is high-level information about a parsed script.
type ScriptMetadata struct {
	Title                         string
	Format                        string
	EstimatedTotalDurationSeconds float64
	SegmentCount                  uint32
}

// ParseScriptParams are the inputs for the ParseScript RPC.
// Exactly one of Text or FilePath should be set.
type ParseScriptParams struct {
	Text       string
	FilePath   string
	FormatHint string // "screenplay", "youtube", "podcast", "narration"
}

// ParseScriptResult is the output of the ParseScript RPC.
type ParseScriptResult struct {
	Segments []ScriptSegment
	Metadata ScriptMetadata
}

// MatchStrategy controls how asset matching is performed.
type MatchStrategy int

const (
	MatchStrategyUnspecified MatchStrategy = 0
	MatchStrategyKeyword    MatchStrategy = 1
	MatchStrategyEmbedding  MatchStrategy = 2
	MatchStrategyHybrid     MatchStrategy = 3
)

// AssetMatch pairs a script segment with a matched asset.
type AssetMatch struct {
	SegmentIndex   uint32
	AssetID        string
	Confidence     float64
	Reasoning      string
	SuggestedRange *TimeRange
}

// UnmatchedSegment records a segment that could not be matched.
type UnmatchedSegment struct {
	SegmentIndex uint32
	Reason       string
	Suggestions  []string
}

// MatchAssetsParams are the inputs for the MatchAssets RPC.
type MatchAssetsParams struct {
	Segments        []ScriptSegment
	AvailableAssets []Asset
	Strategy        MatchStrategy
}

// MatchAssetsResult is the output of the MatchAssets RPC.
type MatchAssetsResult struct {
	Matches   []AssetMatch
	Unmatched []UnmatchedSegment
}

// PacingPreset selects a target pacing feel.
type PacingPreset int

const (
	PacingPresetUnspecified PacingPreset = 0
	PacingPresetSlow       PacingPreset = 1
	PacingPresetModerate   PacingPreset = 2
	PacingPresetFast       PacingPreset = 3
	PacingPresetDynamic    PacingPreset = 4
)

// EDLSettings control EDL generation parameters.
type EDLSettings struct {
	Resolution                 Resolution
	FrameRate                  float64
	DefaultTransition          string
	DefaultTransitionDuration  float64
	Pacing                     PacingPreset
}

// GenerateEDLParams are the inputs for the GenerateEDL RPC.
type GenerateEDLParams struct {
	Segments        []ScriptSegment
	AvailableAssets []Asset
	Matches         []AssetMatch
	Settings        EDLSettings
}

// GenerateEDLResult is the output of the GenerateEDL RPC.
type GenerateEDLResult struct {
	EDL      EditDecisionList
	Warnings []string
}

// AnalyzePacingParams are the inputs for the AnalyzePacing RPC.
type AnalyzePacingParams struct {
	EDL        EditDecisionList
	TargetMood string // "energetic", "calm", "dramatic"
}

// PacingAdjustment is a single timing suggestion.
type PacingAdjustment struct {
	EDLEntryIndex     uint32
	CurrentDuration   float64
	SuggestedDuration float64
	Reason            string
}

// AnalyzePacingResult is the output of the AnalyzePacing RPC.
type AnalyzePacingResult struct {
	Adjustments              []PacingAdjustment
	CurrentAvgClipDuration   float64
	SuggestedAvgClipDuration float64
}

// ---------------------------------------------------------------------------
// Premiere bridge types (mirrors premierpro.premiere.v1)
// ---------------------------------------------------------------------------

// PingResult is the output of the Ping RPC.
type PingResult struct {
	PremiereRunning bool
	PremiereVersion string
	ProjectOpen     bool
	BridgeMode      string // "cep" or "standalone"
}

// SequenceInfo describes a sequence in the project.
type SequenceInfo struct {
	ID              string
	Name            string
	Resolution      Resolution
	FrameRate       float64
	DurationSeconds float64
	VideoTrackCount uint32
	AudioTrackCount uint32
}

// GetProjectStateResult is the output of the GetProjectState RPC.
type GetProjectStateResult struct {
	ProjectName string
	ProjectPath string
	Sequences   []SequenceInfo
	BinCount    uint32
	IsSaved     bool
}

// CreateSequenceParams are the inputs for the CreateSequence RPC.
type CreateSequenceParams struct {
	Name        string
	Resolution  Resolution
	FrameRate   float64
	VideoTracks uint32
	AudioTracks uint32
}

// CreateSequenceResult is the output of the CreateSequence RPC.
type CreateSequenceResult struct {
	SequenceID string
	Name       string
}

// TimelineClip is a clip on the timeline.
type TimelineClip struct {
	ClipID        string
	SourcePath    string
	SourceRange   TimeRange
	TimelineRange TimeRange
	Speed         float64
}

// TimelineTrack is a single track with its clips.
type TimelineTrack struct {
	Index   uint32
	Type    TrackType
	Clips   []TimelineClip
	IsMuted bool
	IsLocked bool
}

// GetTimelineStateParams are the inputs for the GetTimelineState RPC.
type GetTimelineStateParams struct {
	SequenceID string
}

// GetTimelineStateResult is the output of the GetTimelineState RPC.
type GetTimelineStateResult struct {
	SequenceID           string
	VideoTracks          []TimelineTrack
	AudioTracks          []TimelineTrack
	TotalDurationSeconds float64
}

// ImportMediaParams are the inputs for the ImportMedia RPC.
type ImportMediaParams struct {
	FilePath  string
	TargetBin string
}

// ImportMediaResult is the output of the ImportMedia RPC.
type ImportMediaResult struct {
	ProjectItemID string
	Name          string
}

// PlaceClipParams are the inputs for the PlaceClip RPC.
type PlaceClipParams struct {
	SourcePath  string
	Track       TrackTarget
	Position    Timecode
	SourceRange *TimeRange
	Speed       float64
}

// PlaceClipResult is the output of the PlaceClip RPC.
type PlaceClipResult struct {
	ClipID string
}

// RemoveClipParams are the inputs for the RemoveClip RPC.
type RemoveClipParams struct {
	ClipID     string
	SequenceID string
}

// AddTransitionParams are the inputs for the AddTransition RPC.
type AddTransitionParams struct {
	SequenceID     string
	Track          TrackTarget
	Position       Timecode
	TransitionType string
	DurationSeconds float64
}

// AddTransitionResult is the output of the AddTransition RPC.
type AddTransitionResult struct {
	TransitionID string
}

// AddTextParams are the inputs for the AddText RPC.
type AddTextParams struct {
	SequenceID      string
	Text            string
	Style           TextStyle
	Track           TrackTarget
	Position        Timecode
	DurationSeconds float64
}

// AddTextResult is the output of the AddText RPC.
type AddTextResult struct {
	ClipID string
}

// ApplyEffectParams are the inputs for the ApplyEffect RPC.
type ApplyEffectParams struct {
	ClipID     string
	SequenceID string
	Effect     EffectInfo
}

// SetAudioLevelParams are the inputs for the SetAudioLevel RPC.
type SetAudioLevelParams struct {
	ClipID     string
	SequenceID string
	LevelDB    float64
}

// ExportPreset selects the export format/quality.
type ExportPreset int

const (
	ExportPresetUnspecified ExportPreset = 0
	ExportPresetH264_1080P ExportPreset = 1
	ExportPresetH264_4K    ExportPreset = 2
	ExportPresetProRes422  ExportPreset = 3
	ExportPresetProRes4444 ExportPreset = 4
	ExportPresetDNxHR      ExportPreset = 5
	ExportPresetCustom     ExportPreset = 6
)

// ExportSequenceParams are the inputs for the ExportSequence RPC.
type ExportSequenceParams struct {
	SequenceID string
	OutputPath string
	Preset     ExportPreset
}

// ExportSequenceResult is the output of the ExportSequence RPC.
type ExportSequenceResult struct {
	JobID      string
	Status     OperationStatus
	OutputPath string
}

// ExecuteEDLParams are the inputs for the ExecuteEDL RPC.
type ExecuteEDLParams struct {
	EDL                EditDecisionList
	AutoImport         bool
	AutoCreateSequence bool
}

// ExecuteEDLResult is the output of the ExecuteEDL RPC.
type ExecuteEDLResult struct {
	SequenceID       string
	Status           OperationStatus
	ClipsPlaced      uint32
	TransitionsAdded uint32
	Errors           []string
	Warnings         []string
}
