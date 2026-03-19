// Package orchestrator implements the core coordination engine for the
// PremierPro MCP service. It fans work out to the Rust media engine,
// Python intelligence service, and TypeScript Premiere bridge, then
// assembles results into coherent responses.
package orchestrator

import "time"

// ---------------------------------------------------------------------------
// Media / Asset types  (mirrors common.proto + media.proto)
// ---------------------------------------------------------------------------

// AssetType classifies a media file.
type AssetType int

const (
	AssetTypeUnspecified AssetType = iota
	AssetTypeVideo
	AssetTypeAudio
	AssetTypeImage
	AssetTypeGraphics
)

// Resolution describes a pixel dimension pair.
type Resolution struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

// VideoInfo holds video-stream metadata.
type VideoInfo struct {
	Codec           string  `json:"codec"`
	Resolution      Resolution `json:"resolution"`
	FrameRate       float64 `json:"frame_rate"`
	BitrateBPS      uint64  `json:"bitrate_bps"`
	PixelFormat     string  `json:"pixel_format"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// AudioInfo holds audio-stream metadata.
type AudioInfo struct {
	Codec           string  `json:"codec"`
	SampleRate      uint32  `json:"sample_rate"`
	Channels        uint32  `json:"channels"`
	BitrateBPS      uint64  `json:"bitrate_bps"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// AssetInfo represents a single media asset with full metadata.
type AssetInfo struct {
	ID            string            `json:"id"`
	FilePath      string            `json:"file_path"`
	FileName      string            `json:"file_name"`
	FileSizeBytes uint64            `json:"file_size_bytes"`
	MIMEType      string            `json:"mime_type"`
	Type          AssetType         `json:"asset_type"`
	Video         *VideoInfo        `json:"video,omitempty"`
	Audio         *AudioInfo        `json:"audio,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Fingerprint   string            `json:"fingerprint"`
}

// ScanResult is the output of a directory scan for media assets.
type ScanResult struct {
	Assets              []*AssetInfo `json:"assets"`
	TotalFilesScanned   uint32       `json:"total_files_scanned"`
	MediaFilesFound     uint32       `json:"media_files_found"`
	ScanDurationSeconds float64      `json:"scan_duration_seconds"`
}

// SilenceRegion describes a silence interval in an audio track.
type SilenceRegion struct {
	StartSeconds float64 `json:"start_seconds"`
	EndSeconds   float64 `json:"end_seconds"`
	AvgDB        float64 `json:"avg_db"`
}

// WaveformOptions configures audio waveform analysis.
type WaveformOptions struct {
	AudioTrack              uint32  `json:"audio_track"`
	SilenceThresholdDB      float64 `json:"silence_threshold_db"`
	MinSilenceDurationSecs  float64 `json:"min_silence_duration_seconds"`
}

// WaveformResult contains the results of waveform analysis.
type WaveformResult struct {
	SilenceRegions  []*SilenceRegion `json:"silence_regions"`
	PeakDB          float64          `json:"peak_db"`
	RMSDB           float64          `json:"rms_db"`
	DurationSeconds float64          `json:"duration_seconds"`
	WaveformSamples []float32        `json:"waveform_samples"`
}

// SceneChange marks a detected scene boundary.
type SceneChange struct {
	TimecodeSeconds float64 `json:"timecode_seconds"`
	Confidence      float64 `json:"confidence"`
}

// SceneResult contains detected scene changes.
type SceneResult struct {
	Scenes []*SceneChange `json:"scenes"`
}

// ---------------------------------------------------------------------------
// Script / Intelligence types  (mirrors intelligence.proto)
// ---------------------------------------------------------------------------

// SegmentType classifies a script segment.
type SegmentType int

const (
	SegmentTypeUnspecified SegmentType = iota
	SegmentTypeDialogue
	SegmentTypeAction
	SegmentTypeBRoll
	SegmentTypeTransition
	SegmentTypeTitle
	SegmentTypeLowerThird
	SegmentTypeVoiceover
	SegmentTypeMusic
	SegmentTypeSFX
)

// ScriptSegment is one logical block of a parsed script.
type ScriptSegment struct {
	Index                    uint32      `json:"index"`
	Type                     SegmentType `json:"type"`
	Content                  string      `json:"content"`
	Speaker                  string      `json:"speaker,omitempty"`
	SceneDescription         string      `json:"scene_description,omitempty"`
	VisualDirection          string      `json:"visual_direction,omitempty"`
	AudioDirection           string      `json:"audio_direction,omitempty"`
	EstimatedDurationSeconds float64     `json:"estimated_duration_seconds"`
	AssetHints               []string    `json:"asset_hints,omitempty"`
}

// ScriptMetadata contains summary info about a parsed script.
type ScriptMetadata struct {
	Title                        string  `json:"title"`
	Format                       string  `json:"format"`
	EstimatedTotalDurationSeconds float64 `json:"estimated_total_duration_seconds"`
	SegmentCount                 uint32  `json:"segment_count"`
}

// ParsedScript is the full result of script parsing.
type ParsedScript struct {
	Segments []*ScriptSegment `json:"segments"`
	Metadata *ScriptMetadata  `json:"metadata"`
}

// AssetMatch pairs a script segment with a matched asset.
type AssetMatch struct {
	SegmentIndex   uint32     `json:"segment_index"`
	AssetID        string     `json:"asset_id"`
	Confidence     float64    `json:"confidence"`
	Reasoning      string     `json:"reasoning"`
	SuggestedRange *TimeRange `json:"suggested_range,omitempty"`
}

// UnmatchedSegment records a segment that could not be matched.
type UnmatchedSegment struct {
	SegmentIndex uint32   `json:"segment_index"`
	Reason       string   `json:"reason"`
	Suggestions  []string `json:"suggestions,omitempty"`
}

// MatchResult holds the output of asset-to-segment matching.
type MatchResult struct {
	Matches   []*AssetMatch     `json:"matches"`
	Unmatched []*UnmatchedSegment `json:"unmatched,omitempty"`
}

// PacingPreset selects a pacing profile.
type PacingPreset int

const (
	PacingPresetUnspecified PacingPreset = iota
	PacingPresetSlow
	PacingPresetModerate
	PacingPresetFast
	PacingPresetDynamic
)

// EDLSettings controls edit decision list generation.
type EDLSettings struct {
	Resolution                Resolution   `json:"resolution"`
	FrameRate                 float64      `json:"frame_rate"`
	DefaultTransition         string       `json:"default_transition"`
	DefaultTransitionDuration float64      `json:"default_transition_duration"`
	Pacing                    PacingPreset `json:"pacing"`
}

// PacingAdjustment is a suggested change to a single EDL entry's timing.
type PacingAdjustment struct {
	EDLEntryIndex     uint32  `json:"edl_entry_index"`
	CurrentDuration   float64 `json:"current_duration"`
	SuggestedDuration float64 `json:"suggested_duration"`
	Reason            string  `json:"reason"`
}

// PacingResult holds the output of pacing analysis.
type PacingResult struct {
	Adjustments              []*PacingAdjustment `json:"adjustments"`
	CurrentAvgClipDuration   float64             `json:"current_avg_clip_duration"`
	SuggestedAvgClipDuration float64             `json:"suggested_avg_clip_duration"`
}

// ---------------------------------------------------------------------------
// Timecode / time-range types  (mirrors common.proto)
// ---------------------------------------------------------------------------

// Timecode represents a position in HH:MM:SS:FF form.
type Timecode struct {
	Hours     uint32  `json:"hours"`
	Minutes   uint32  `json:"minutes"`
	Seconds   uint32  `json:"seconds"`
	Frames    uint32  `json:"frames"`
	FrameRate float64 `json:"frame_rate"`
}

// TimeRange is a span defined by in/out timecodes.
type TimeRange struct {
	InPoint  Timecode `json:"in_point"`
	OutPoint Timecode `json:"out_point"`
}

// TrackType distinguishes video from audio tracks.
type TrackType int

const (
	TrackTypeUnspecified TrackType = iota
	TrackTypeVideo
	TrackTypeAudio
)

// TrackTarget identifies a specific track.
type TrackTarget struct {
	Type       TrackType `json:"type"`
	TrackIndex uint32    `json:"track_index"`
}

// TransitionInfo describes a transition between clips.
type TransitionInfo struct {
	Type            string  `json:"type"`
	DurationSeconds float64 `json:"duration_seconds"`
	Alignment       string  `json:"alignment"`
}

// EffectInfo describes an effect applied to a clip.
type EffectInfo struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// TextStyle describes visual properties of a text overlay.
type TextStyle struct {
	FontFamily         string   `json:"font_family"`
	FontSize           float64  `json:"font_size"`
	ColorHex           string   `json:"color_hex"`
	Alignment          string   `json:"alignment"`
	BackgroundColorHex string   `json:"background_color_hex,omitempty"`
	BackgroundOpacity  float64  `json:"background_opacity,omitempty"`
	Position           Position `json:"position"`
}

// Position is a normalized (0.0-1.0) screen coordinate.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ---------------------------------------------------------------------------
// EDL types  (mirrors common.proto EditDecisionList)
// ---------------------------------------------------------------------------

// EDLEntry is a single instruction in an edit decision list.
type EDLEntry struct {
	Index         uint32          `json:"index"`
	SourceAssetID string          `json:"source_asset_id"`
	SourceRange   *TimeRange      `json:"source_range,omitempty"`
	TimelineRange *TimeRange      `json:"timeline_range,omitempty"`
	Track         *TrackTarget    `json:"track,omitempty"`
	Transition    *TransitionInfo `json:"transition,omitempty"`
	Effects       []*EffectInfo   `json:"effects,omitempty"`
	Notes         string          `json:"notes,omitempty"`
}

// EDL is a complete edit decision list that can be executed by Premiere.
type EDL struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	SequenceResolution Resolution `json:"sequence_resolution"`
	SequenceFrameRate float64    `json:"sequence_frame_rate"`
	Entries           []*EDLEntry `json:"entries"`
}

// ---------------------------------------------------------------------------
// Premiere / Bridge operation types  (mirrors premiere.proto)
// ---------------------------------------------------------------------------

// PingResult reports health of the Premiere Pro connection.
type PingResult struct {
	PremiereRunning bool   `json:"premiere_running"`
	PremiereVersion string `json:"premiere_version"`
	ProjectOpen     bool   `json:"project_open"`
	BridgeMode      string `json:"bridge_mode"`
}

// SequenceInfo summarises an existing sequence.
type SequenceInfo struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Resolution      Resolution `json:"resolution"`
	FrameRate       float64    `json:"frame_rate"`
	DurationSeconds float64    `json:"duration_seconds"`
	VideoTrackCount uint32     `json:"video_track_count"`
	AudioTrackCount uint32     `json:"audio_track_count"`
}

// ProjectState describes the current Premiere Pro project.
type ProjectState struct {
	ProjectName string          `json:"project_name"`
	ProjectPath string          `json:"project_path"`
	Sequences   []*SequenceInfo `json:"sequences"`
	BinCount    uint32          `json:"bin_count"`
	IsSaved     bool            `json:"is_saved"`
}

// TimelineClip describes one clip on the timeline.
type TimelineClip struct {
	ClipID        string     `json:"clip_id"`
	SourcePath    string     `json:"source_path"`
	SourceRange   *TimeRange `json:"source_range,omitempty"`
	TimelineRange *TimeRange `json:"timeline_range,omitempty"`
	Speed         float64    `json:"speed"`
}

// TimelineTrack describes one track in a sequence.
type TimelineTrack struct {
	Index    uint32          `json:"index"`
	Type     TrackType       `json:"type"`
	Clips    []*TimelineClip `json:"clips"`
	IsMuted  bool            `json:"is_muted"`
	IsLocked bool            `json:"is_locked"`
}

// TimelineState is a snapshot of a sequence's timeline.
type TimelineState struct {
	SequenceID           string           `json:"sequence_id"`
	VideoTracks          []*TimelineTrack `json:"video_tracks"`
	AudioTracks          []*TimelineTrack `json:"audio_tracks"`
	TotalDurationSeconds float64          `json:"total_duration_seconds"`
}

// ---------------------------------------------------------------------------
// Operation parameter types
// ---------------------------------------------------------------------------

// CreateSequenceParams defines how to create a new sequence.
type CreateSequenceParams struct {
	Name        string     `json:"name"`
	Resolution  Resolution `json:"resolution"`
	FrameRate   float64    `json:"frame_rate"`
	VideoTracks uint32     `json:"video_tracks"`
	AudioTracks uint32     `json:"audio_tracks"`
}

// PlaceClipParams defines how to place a clip on the timeline.
type PlaceClipParams struct {
	SourcePath  string      `json:"source_path"`
	Track       TrackTarget `json:"track"`
	Position    Timecode    `json:"position"`
	SourceRange *TimeRange  `json:"source_range,omitempty"`
	Speed       float64     `json:"speed"`
}

// TransitionParams defines how to add a transition.
type TransitionParams struct {
	SequenceID      string      `json:"sequence_id"`
	Track           TrackTarget `json:"track"`
	Position        Timecode    `json:"position"`
	TransitionType  string      `json:"transition_type"`
	DurationSeconds float64     `json:"duration_seconds"`
}

// TextParams defines a text overlay to add.
type TextParams struct {
	SequenceID      string      `json:"sequence_id"`
	Text            string      `json:"text"`
	Style           TextStyle   `json:"style"`
	Track           TrackTarget `json:"track"`
	Position        Timecode    `json:"position"`
	DurationSeconds float64     `json:"duration_seconds"`
}

// ExportPreset selects a predefined export profile.
type ExportPreset int

const (
	ExportPresetUnspecified ExportPreset = iota
	ExportPresetH264_1080P
	ExportPresetH264_4K
	ExportPresetProRes422
	ExportPresetProRes4444
	ExportPresetDNxHR
	ExportPresetCustom
)

// ExportParams defines how to export a sequence.
type ExportParams struct {
	SequenceID string       `json:"sequence_id"`
	OutputPath string       `json:"output_path"`
	Preset     ExportPreset `json:"preset"`
}

// ---------------------------------------------------------------------------
// Operation result types
// ---------------------------------------------------------------------------

// SequenceResult is returned after creating a sequence.
type SequenceResult struct {
	SequenceID string `json:"sequence_id"`
	Name       string `json:"name"`
}

// ImportResult is returned after importing media.
type ImportResult struct {
	ProjectItemID string `json:"project_item_id"`
	Name          string `json:"name"`
}

// ClipResult is returned after placing a clip or adding text.
type ClipResult struct {
	ClipID string `json:"clip_id"`
}

// ExportResult is returned after starting an export.
type ExportResult struct {
	JobID      string `json:"job_id"`
	Status     string `json:"status"`
	OutputPath string `json:"output_path"`
}

// ---------------------------------------------------------------------------
// Extended export parameter / result types
// ---------------------------------------------------------------------------

// ExportDirectParams defines how to perform a synchronous direct export.
type ExportDirectParams struct {
	SequenceIndex int    `json:"sequence_index"`
	OutputPath    string `json:"output_path"`
	PresetPath    string `json:"preset_path"`
	WorkAreaType  int    `json:"work_area_type"` // 0=entire, 1=in-to-out, 2=work area
}

// ExportViaAMEParams defines how to queue an export through Adobe Media Encoder.
type ExportViaAMEParams struct {
	SequenceIndex int    `json:"sequence_index"`
	OutputPath    string `json:"output_path"`
	PresetPath    string `json:"preset_path"`
	WorkAreaType  int    `json:"work_area_type"`
	RemoveOnDone  bool   `json:"remove_on_done"`
}

// ExportFrameParams defines how to export a single frame.
type ExportFrameParams struct {
	OutputPath string `json:"output_path"`
	Format     string `json:"format"` // "PNG" or "JPEG"
}

// ExportAAFParams defines how to export as AAF.
type ExportAAFParams struct {
	SequenceIndex int    `json:"sequence_index"`
	OutputPath    string `json:"output_path"`
	Mixdown       bool   `json:"mixdown"`
	Explode       bool   `json:"explode"`
	SampleRate    int    `json:"sample_rate"`
	BitsPerSample int    `json:"bits_per_sample"`
}

// ExportOMFParams defines how to export as OMF.
type ExportOMFParams struct {
	SequenceIndex int    `json:"sequence_index"`
	OutputPath    string `json:"output_path"`
	SampleRate    int    `json:"sample_rate"`
	BitsPerSample int    `json:"bits_per_sample"`
	HandleFrames  int    `json:"handle_frames"`
	Encapsulate   bool   `json:"encapsulate"`
}

// ExportAudioOnlyParams defines how to export audio only.
type ExportAudioOnlyParams struct {
	SequenceIndex int    `json:"sequence_index"`
	OutputPath    string `json:"output_path"`
	PresetPath    string `json:"preset_path"`
}

// RenderPreviewParams defines the range for preview rendering.
type RenderPreviewParams struct {
	InSeconds  float64 `json:"in_seconds"`
	OutSeconds float64 `json:"out_seconds"`
}

// ExporterInfo describes a single available exporter.
type ExporterInfo struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	ClassID  string `json:"class_id"`
	FileType string `json:"file_type"`
}

// ExporterListResult is returned by getExporters.
type ExporterListResult struct {
	Exporters []ExporterInfo `json:"exporters"`
	Count     int            `json:"count"`
}

// ExportPresetDetailInfo describes a single export preset.
type ExportPresetDetailInfo struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	MatchName string `json:"match_name"`
	Path      string `json:"path"`
}

// ExportPresetListResult is returned by getExportPresets.
type ExportPresetListResult struct {
	ExporterIndex int                      `json:"exporter_index"`
	ExporterName  string                   `json:"exporter_name"`
	Presets       []ExportPresetDetailInfo  `json:"presets"`
	Count         int                      `json:"count"`
}

// ExportProgressResult is returned by getExportProgress.
type ExportProgressResult struct {
	EncoderAvailable   bool   `json:"encoder_available"`
	ExportersAvailable bool   `json:"exporters_available"`
	Status             string `json:"status"`
	Note               string `json:"note"`
}

// GenericExportResult is a flexible result returned by various export operations.
type GenericExportResult struct {
	Status       string `json:"status"`
	OutputPath   string `json:"output_path,omitempty"`
	SequenceName string `json:"sequence_name,omitempty"`
	ProjectName  string `json:"project_name,omitempty"`
	JobID        string `json:"job_id,omitempty"`
}

// EDLExecutionResult is returned after executing a full EDL.
type EDLExecutionResult struct {
	SequenceID       string   `json:"sequence_id"`
	Status           string   `json:"status"`
	ClipsPlaced      uint32   `json:"clips_placed"`
	TransitionsAdded uint32   `json:"transitions_added"`
	Errors           []string `json:"errors,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
}

// ---------------------------------------------------------------------------
// Sequence management types
// ---------------------------------------------------------------------------

// SequenceSettings holds comprehensive settings for a sequence.
type SequenceSettings struct {
	Name                  string  `json:"name"`
	SequenceID            string  `json:"sequence_id"`
	FrameSizeHorizontal   int     `json:"frame_size_horizontal"`
	FrameSizeVertical     int     `json:"frame_size_vertical"`
	Timebase              string  `json:"timebase"`
	VideoTrackCount       int     `json:"video_track_count"`
	AudioTrackCount       int     `json:"audio_track_count"`
	InPoint               float64 `json:"in_point"`
	OutPoint              float64 `json:"out_point"`
	EndSeconds            float64 `json:"end_seconds"`
	AudioSampleRate       float64 `json:"audio_sample_rate,omitempty"`
	AudioChannelCount     int     `json:"audio_channel_count,omitempty"`
	VideoFieldType        int     `json:"video_field_type,omitempty"`
	VideoPixelAspectRatio string  `json:"video_pixel_aspect_ratio,omitempty"`
	CompositeLinearColor  bool    `json:"composite_linear_color,omitempty"`
	MaximumBitDepth       bool    `json:"maximum_bit_depth,omitempty"`
	MaximumRenderQuality  bool    `json:"maximum_render_quality,omitempty"`
}

// SetSequenceSettingsParams defines which settings to update.
type SetSequenceSettingsParams struct {
	SequenceIndex        int      `json:"sequence_index"`
	Width                *int     `json:"width,omitempty"`
	Height               *int     `json:"height,omitempty"`
	AudioSampleRate      *float64 `json:"audio_sample_rate,omitempty"`
	VideoFieldType       *int     `json:"video_field_type,omitempty"`
	CompositeLinearColor *bool    `json:"composite_linear_color,omitempty"`
	MaximumBitDepth      *bool    `json:"maximum_bit_depth,omitempty"`
	MaximumRenderQuality *bool    `json:"maximum_render_quality,omitempty"`
}

// SequenceListResult contains a list of all sequences in the project.
type SequenceListResult struct {
	Count             int                  `json:"count"`
	Sequences         []*SequenceListEntry `json:"sequences"`
	ActiveSequenceID  string               `json:"active_sequence_id"`
}

// SequenceListEntry is a summary of a single sequence in the project list.
type SequenceListEntry struct {
	Index               int    `json:"index"`
	Name                string `json:"name"`
	SequenceID          string `json:"sequence_id"`
	FrameSizeHorizontal int    `json:"frame_size_horizontal"`
	FrameSizeVertical   int    `json:"frame_size_vertical"`
	Timebase            string `json:"timebase"`
	VideoTrackCount     int    `json:"video_track_count"`
	AudioTrackCount     int    `json:"audio_track_count"`
	IsActive            bool   `json:"is_active"`
}

// PlayheadResult describes the current playhead position.
type PlayheadResult struct {
	Seconds      float64 `json:"seconds"`
	Ticks        string  `json:"ticks"`
	SequenceName string  `json:"sequence_name"`
	SequenceID   string  `json:"sequence_id"`
}

// InOutPointsResult describes the in/out points of a sequence.
type InOutPointsResult struct {
	InPoint      float64 `json:"in_point"`
	OutPoint     float64 `json:"out_point"`
	SequenceName string  `json:"sequence_name"`
	SequenceID   string  `json:"sequence_id"`
}

// MarkerInfo describes a single sequence marker.
type MarkerInfo struct {
	Index      int     `json:"index"`
	Name       string  `json:"name"`
	Comment    string  `json:"comment"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Type       string  `json:"type"`
	ColorIndex int     `json:"color_index"`
}

// MarkersResult contains all markers on a sequence.
type MarkersResult struct {
	Count        int           `json:"count"`
	Markers      []*MarkerInfo `json:"markers"`
	SequenceName string        `json:"sequence_name"`
	SequenceID   string        `json:"sequence_id"`
}

// AddMarkerParams defines how to add a marker to a sequence.
type AddMarkerParams struct {
	Time     float64 `json:"time"`
	Name     string  `json:"name"`
	Comment  string  `json:"comment"`
	Color    int     `json:"color"`
	Duration float64 `json:"duration"`
}

// GenericResult is a simple result for operations that return basic status info.
type GenericResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ---------------------------------------------------------------------------
// AutoEdit types
// ---------------------------------------------------------------------------

// AutoEditParams is the input to the fully-automated edit workflow.
type AutoEditParams struct {
	// ScriptText is the raw script content (mutually exclusive with ScriptPath).
	ScriptText string `json:"script_text,omitempty"`

	// ScriptPath is a file path to a script (mutually exclusive with ScriptText).
	ScriptPath string `json:"script_path,omitempty"`

	// ScriptFormat hints at the script style: "screenplay", "youtube", etc.
	ScriptFormat string `json:"script_format,omitempty"`

	// AssetsDirectory is the root folder to scan for media assets.
	AssetsDirectory string `json:"assets_directory"`

	// Recursive controls whether sub-directories are scanned.
	Recursive bool `json:"recursive"`

	// Extensions optionally restricts which file types to include.
	Extensions []string `json:"extensions,omitempty"`

	// MatchStrategy selects the asset-matching algorithm.
	MatchStrategy string `json:"match_strategy,omitempty"`

	// EDLSettings overrides for resolution, frame rate, transitions, pacing.
	EDLSettings *EDLSettings `json:"edl_settings,omitempty"`

	// OutputName, when non-empty, triggers export after assembly.
	OutputName string `json:"output_name,omitempty"`

	// ExportPreset selects the export profile (only used when OutputName is set).
	ExportPreset ExportPreset `json:"export_preset,omitempty"`
}

// StepStatus records the outcome of one stage in the auto-edit pipeline.
type StepStatus struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"` // "pending", "running", "completed", "failed", "skipped"
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Detail    string        `json:"detail,omitempty"`
}

// AutoEditResult captures the complete outcome of an auto-edit run.
type AutoEditResult struct {
	// Steps records the status of each pipeline stage.
	Steps []*StepStatus `json:"steps"`

	// ScanResult from the Rust engine asset scan.
	ScanResult *ScanResult `json:"scan_result,omitempty"`

	// ParsedScript from the Python intelligence service.
	ParsedScript *ParsedScript `json:"parsed_script,omitempty"`

	// MatchResult from the Python asset matcher.
	MatchResult *MatchResult `json:"match_result,omitempty"`

	// EDL generated by the Python service.
	EDL *EDL `json:"edl,omitempty"`

	// ExecutionResult from the TypeScript bridge.
	ExecutionResult *EDLExecutionResult `json:"execution_result,omitempty"`

	// ExportResult, present only when OutputName was provided.
	ExportResult *ExportResult `json:"export_result,omitempty"`

	// TotalDuration is the wall-clock time for the entire pipeline.
	TotalDuration time.Duration `json:"total_duration"`
}
