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
// Project Management types
// ---------------------------------------------------------------------------

// BinInfo describes a top-level bin in the project.
type BinInfo struct {
	Name       string `json:"name"`
	ChildCount int    `json:"child_count"`
}

// ActiveSequenceInfo summarises the currently active sequence.
type ActiveSequenceInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// ProjectInfoResult is returned by GetProjectInfo.
type ProjectInfoResult struct {
	Name           string              `json:"name"`
	Path           string              `json:"path"`
	DocumentID     string              `json:"document_id"`
	Sequences      []*SequenceListEntry `json:"sequences"`
	Bins           []*BinInfo          `json:"bins"`
	TotalItems     int                 `json:"total_items"`
	ActiveSequence *ActiveSequenceInfo `json:"active_sequence,omitempty"`
}

// ProjectItemInfo describes a single item in the project panel.
type ProjectItemInfo struct {
	Index      int    `json:"index,omitempty"`
	Name       string `json:"name"`
	Path       string `json:"path,omitempty"`
	Type       string `json:"type"`
	MediaPath  string `json:"media_path,omitempty"`
	ChildCount int    `json:"child_count,omitempty"`
}

// ProjectItemsResult is returned by FindProjectItems, GetProjectItems, GetOfflineItems.
type ProjectItemsResult struct {
	Query       string             `json:"query,omitempty"`
	BinPath     string             `json:"bin_path,omitempty"`
	ItemCount   int                `json:"item_count"`
	Items       []*ProjectItemInfo `json:"items"`
}

// ItemMetadataResult is returned by GetItemMetadata.
type ItemMetadataResult struct {
	ItemPath string            `json:"item_path"`
	Metadata map[string]any    `json:"metadata"`
}

// ConsolidateResult is returned by ConsolidateDuplicates.
type ConsolidateResult struct {
	TotalChecked      int `json:"total_checked"`
	DuplicatesFound   int `json:"duplicates_found"`
	DuplicatesRemoved int `json:"duplicates_removed"`
}

// ProjectSettingsResult is returned by GetProjectSettingsInfo.
type ProjectSettingsResult struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	DocumentID     string `json:"document_id"`
	GPURenderer    string `json:"gpu_renderer,omitempty"`
	RootItemCount  int    `json:"root_item_count,omitempty"`
	SequenceCount  int    `json:"sequence_count,omitempty"`
}

// ---------------------------------------------------------------------------
// AutoEdit types
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// AI-powered editing types
// ---------------------------------------------------------------------------

// SmartCutParams configures automatic silence-based cutting.
type SmartCutParams struct {
	SequenceID         string  `json:"sequence_id"`
	TrackIndex         int     `json:"track_index"`
	SilenceThresholdDB float64 `json:"silence_threshold_db"`
	MinSilenceDuration float64 `json:"min_silence_duration"`
	PaddingSeconds     float64 `json:"padding_seconds"`
}

// SmartCutResult is returned by SmartCut.
type SmartCutResult struct {
	RegionsDetected int      `json:"regions_detected"`
	CutsMade        int      `json:"cuts_made"`
	DurationRemoved float64  `json:"duration_removed_seconds"`
	Details         []string `json:"details,omitempty"`
}

// SmartTrimParams configures automatic head/tail silence trimming.
type SmartTrimParams struct {
	SequenceID         string  `json:"sequence_id"`
	TrackIndex         int     `json:"track_index"`
	ClipIndex          int     `json:"clip_index"`
	SilenceThresholdDB float64 `json:"silence_threshold_db"`
	PaddingSeconds     float64 `json:"padding_seconds"`
}

// SmartTrimResult is returned by SmartTrim.
type SmartTrimResult struct {
	HeadTrimmed float64 `json:"head_trimmed_seconds"`
	TailTrimmed float64 `json:"tail_trimmed_seconds"`
	NewDuration float64 `json:"new_duration_seconds"`
}

// AutoColorMatchParams configures automatic color matching between clips.
type AutoColorMatchParams struct {
	SequenceID     string `json:"sequence_id"`
	SrcTrackIndex  int    `json:"src_track_index"`
	SrcClipIndex   int    `json:"src_clip_index"`
	DestTrackIndex int    `json:"dest_track_index"`
	DestClipIndex  int    `json:"dest_clip_index"`
}

// AutoColorMatchResult is returned by AutoColorMatch.
type AutoColorMatchResult struct {
	Status         string  `json:"status"`
	BrightnessAdj  float64 `json:"brightness_adjustment"`
	ContrastAdj    float64 `json:"contrast_adjustment"`
	SaturationAdj  float64 `json:"saturation_adjustment"`
	TemperatureAdj float64 `json:"temperature_adjustment"`
}

// AutoAudioLevelsParams configures automatic audio normalisation.
type AutoAudioLevelsParams struct {
	SequenceID string  `json:"sequence_id"`
	TargetLUFS float64 `json:"target_lufs"`
	MaxPeakDB  float64 `json:"max_peak_db"`
}

// AutoAudioLevelsResult is returned by AutoAudioLevels.
type AutoAudioLevelsResult struct {
	ClipsAdjusted int      `json:"clips_adjusted"`
	AvgAdjustment float64  `json:"avg_adjustment_db"`
	Details       []string `json:"details,omitempty"`
}

// TransitionSuggestion represents a single AI-suggested transition.
type TransitionSuggestion struct {
	Position       float64 `json:"position_seconds"`
	Type           string  `json:"type"`
	Duration       float64 `json:"duration_seconds"`
	Confidence     float64 `json:"confidence"`
	Reason         string  `json:"reason"`
}

// SuggestTransitionsResult is returned by SuggestTransitions.
type SuggestTransitionsResult struct {
	Suggestions []*TransitionSuggestion `json:"suggestions"`
}

// MusicSuggestion represents a single AI-suggested music cut point.
type MusicSuggestion struct {
	TimeSeconds float64 `json:"time_seconds"`
	Type        string  `json:"type"`
	Reason      string  `json:"reason"`
	Intensity   float64 `json:"intensity"`
}

// SuggestMusicResult is returned by SuggestMusic.
type SuggestMusicResult struct {
	Suggestions  []*MusicSuggestion `json:"suggestions"`
	AvgPacing    float64            `json:"avg_pacing_seconds"`
	OverallMood  string             `json:"overall_mood"`
}

// ClipAnalysis contains the analysis of a single clip.
type ClipAnalysis struct {
	FilePath        string          `json:"file_path"`
	Duration        float64         `json:"duration_seconds"`
	PeakAudioDB     float64         `json:"peak_audio_db"`
	RmsAudioDB      float64         `json:"rms_audio_db"`
	SceneChanges    []*SceneChange  `json:"scene_changes,omitempty"`
	SilenceRegions  []*SilenceRegion `json:"silence_regions,omitempty"`
	HasMotion       bool            `json:"has_motion"`
	AvgBrightness   float64         `json:"avg_brightness"`
}

// SequenceAnalysis contains the analysis of a full sequence.
type SequenceAnalysis struct {
	SequenceID       string   `json:"sequence_id"`
	TotalDuration    float64  `json:"total_duration_seconds"`
	ClipCount        int      `json:"clip_count"`
	AvgClipDuration  float64  `json:"avg_clip_duration_seconds"`
	PacingScore      float64  `json:"pacing_score"`
	AudioBalance     float64  `json:"audio_balance_score"`
	GapCount         int      `json:"gap_count"`
	TransitionCount  int      `json:"transition_count"`
	Issues           []string `json:"issues,omitempty"`
	Suggestions      []string `json:"suggestions,omitempty"`
}

// SequenceStatistics contains summary statistics for a sequence.
type SequenceStatistics struct {
	SequenceID        string  `json:"sequence_id"`
	TotalDuration     float64 `json:"total_duration_seconds"`
	VideoClipCount    int     `json:"video_clip_count"`
	AudioClipCount    int     `json:"audio_clip_count"`
	AvgClipDuration   float64 `json:"avg_clip_duration_seconds"`
	VideoTrackUsage   int     `json:"video_tracks_used"`
	AudioTrackUsage   int     `json:"audio_tracks_used"`
	TransitionCount   int     `json:"transition_count"`
	EffectsCount      int     `json:"effects_count"`
	TotalGapDuration  float64 `json:"total_gap_duration_seconds"`
}

// JumpCutInfo describes a detected potential jump cut.
type JumpCutInfo struct {
	PositionSeconds float64 `json:"position_seconds"`
	Confidence      float64 `json:"confidence"`
	ClipBefore      string  `json:"clip_before"`
	ClipAfter       string  `json:"clip_after"`
}

// JumpCutResult is returned by DetectJumpCuts.
type JumpCutResult struct {
	JumpCuts []*JumpCutInfo `json:"jump_cuts"`
	Count    int            `json:"count"`
}

// AudioIssue describes a detected audio problem.
type AudioIssue struct {
	Type            string  `json:"type"`
	PositionSeconds float64 `json:"position_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
	Severity        string  `json:"severity"`
	Description     string  `json:"description"`
}

// AudioIssuesResult is returned by DetectAudioIssues.
type AudioIssuesResult struct {
	Issues []AudioIssue `json:"issues"`
	Count  int          `json:"count"`
}

// RoughCutParams configures rough cut generation from script and assets.
type RoughCutParams struct {
	ScriptPath      string `json:"script_path,omitempty"`
	ScriptText      string `json:"script_text,omitempty"`
	AssetsDirectory string `json:"assets_directory"`
	OutputName      string `json:"output_name,omitempty"`
	Pacing          string `json:"pacing,omitempty"`
}

// RoughCutResult is returned by GenerateRoughCut.
type RoughCutResult struct {
	SequenceID     string        `json:"sequence_id"`
	ClipsPlaced    int           `json:"clips_placed"`
	TotalDuration  float64       `json:"total_duration_seconds"`
	UnmatchedCount int           `json:"unmatched_count"`
	Steps          []*StepStatus `json:"steps"`
}

// RefineEditParams configures AI refinement of an existing edit.
type RefineEditParams struct {
	SequenceID     string `json:"sequence_id"`
	TargetPacing   string `json:"target_pacing,omitempty"`
	AddTransitions bool   `json:"add_transitions"`
	AdjustAudio    bool   `json:"adjust_audio"`
	TargetMood     string `json:"target_mood,omitempty"`
}

// RefineEditResult is returned by RefineEdit.
type RefineEditResult struct {
	SequenceID      string   `json:"sequence_id"`
	PacingChanges   int      `json:"pacing_changes"`
	TransitionsAdded int     `json:"transitions_added"`
	AudioAdjustments int    `json:"audio_adjustments"`
	Summary         string   `json:"summary"`
	Suggestions     []string `json:"suggestions,omitempty"`
}

// BRollSuggestion represents an AI-suggested B-roll placement.
type BRollSuggestion struct {
	PositionSeconds float64  `json:"position_seconds"`
	DurationSeconds float64  `json:"duration_seconds"`
	ContentType     string   `json:"content_type"`
	Reason          string   `json:"reason"`
	Keywords        []string `json:"keywords,omitempty"`
}

// BRollSuggestionsResult is returned by AddBRollSuggestions.
type BRollSuggestionsResult struct {
	Suggestions []*BRollSuggestion `json:"suggestions"`
	Count       int                `json:"count"`
}

// GenerateTrailerParams configures trailer generation.
type GenerateTrailerParams struct {
	SequenceID      string  `json:"sequence_id"`
	MaxDuration     float64 `json:"max_duration_seconds"`
	Style           string  `json:"style,omitempty"`
	IncludeAudio    bool    `json:"include_audio"`
}

// GenerateTrailerResult is returned by GenerateTrailer.
type GenerateTrailerResult struct {
	SequenceID    string  `json:"sequence_id"`
	Duration      float64 `json:"duration_seconds"`
	HighlightCount int    `json:"highlight_count"`
	Summary       string  `json:"summary"`
}

// SocialCutParams configures social media cut creation.
type SocialCutParams struct {
	SequenceID    string  `json:"sequence_id"`
	AspectRatio   string  `json:"aspect_ratio"`
	MaxDuration   float64 `json:"max_duration_seconds"`
	Platform      string  `json:"platform,omitempty"`
}

// SocialCutResult is returned by CreateSocialCuts.
type SocialCutResult struct {
	SequenceID  string  `json:"sequence_id"`
	AspectRatio string  `json:"aspect_ratio"`
	Duration    float64 `json:"duration_seconds"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
}

// AutoOrganizeParams configures AI-powered project organisation.
type AutoOrganizeParams struct {
	Strategy string `json:"strategy,omitempty"`
}

// AutoOrganizeResult is returned by AutoOrganizeProject.
type AutoOrganizeResult struct {
	BinsCreated   int      `json:"bins_created"`
	ItemsMoved    int      `json:"items_moved"`
	Categories    []string `json:"categories,omitempty"`
	Summary       string   `json:"summary"`
}

// ClipTag represents an AI-generated tag for a clip.
type ClipTag struct {
	Tag        string  `json:"tag"`
	Confidence float64 `json:"confidence"`
	Category   string  `json:"category"`
}

// TagClipsResult is returned by TagClips.
type TagClipsResult struct {
	ClipPath string     `json:"clip_path"`
	Tags     []*ClipTag `json:"tags"`
	Count    int        `json:"count"`
}

// SimilarClipInfo describes a clip similar to a reference clip.
type SimilarClipInfo struct {
	FilePath   string  `json:"file_path"`
	Similarity float64 `json:"similarity"`
	MatchType  string  `json:"match_type"`
}

// FindSimilarResult is returned by FindSimilarClips.
type FindSimilarResult struct {
	ReferenceClip string             `json:"reference_clip"`
	SimilarClips  []*SimilarClipInfo `json:"similar_clips"`
	Count         int                `json:"count"`
}

// ReplacementSuggestion describes a suggested clip replacement.
type ReplacementSuggestion struct {
	CurrentClipID    string  `json:"current_clip_id"`
	ReplacementPath  string  `json:"replacement_path"`
	Confidence       float64 `json:"confidence"`
	Reason           string  `json:"reason"`
	PositionSeconds  float64 `json:"position_seconds"`
}

// SuggestReplacementsResult is returned by SuggestReplacements.
type SuggestReplacementsResult struct {
	Suggestions []*ReplacementSuggestion `json:"suggestions"`
	Count       int                      `json:"count"`
}

// ReviewMarkerInfo describes a review marker added by the AI.
type ReviewMarkerInfo struct {
	TimeSeconds float64 `json:"time_seconds"`
	Name        string  `json:"name"`
	Comment     string  `json:"comment"`
	Priority    string  `json:"priority"`
}

// ReviewMarkersResult is returned by CreateReviewMarkers.
type ReviewMarkersResult struct {
	MarkersAdded int                 `json:"markers_added"`
	Markers      []*ReviewMarkerInfo `json:"markers"`
}

// EditSummaryResult is returned by GenerateEditSummary.
type EditSummaryResult struct {
	SequenceID      string   `json:"sequence_id"`
	Summary         string   `json:"summary"`
	Duration        float64  `json:"duration_seconds"`
	ClipCount       int      `json:"clip_count"`
	KeyMoments      []string `json:"key_moments,omitempty"`
	EditingStyle    string   `json:"editing_style"`
}

// RenderTimeEstimate is returned by EstimateRenderTime.
type RenderTimeEstimate struct {
	SequenceID        string  `json:"sequence_id"`
	EstimatedSeconds  float64 `json:"estimated_seconds"`
	Complexity        string  `json:"complexity"`
	EffectsHeavy      bool    `json:"effects_heavy"`
	ResolutionFactor  float64 `json:"resolution_factor"`
	Notes             string  `json:"notes,omitempty"`
}

// DeliverySpec describes a single delivery specification check.
type DeliverySpecCheck struct {
	Spec    string `json:"spec"`
	Status  string `json:"status"`
	Current string `json:"current_value"`
	Target  string `json:"target_value"`
	Pass    bool   `json:"pass"`
}

// DeliverySpecResult is returned by CheckDeliverySpecs.
type DeliverySpecResult struct {
	SequenceID string               `json:"sequence_id"`
	Standard   string               `json:"standard"`
	AllPass    bool                  `json:"all_pass"`
	Checks     []*DeliverySpecCheck `json:"checks"`
}

// ProjectReportResult is returned by CreateProjectReport.
type ProjectReportResult struct {
	ProjectName    string            `json:"project_name"`
	TotalDuration  float64           `json:"total_duration_seconds"`
	SequenceCount  int               `json:"sequence_count"`
	TotalClips     int               `json:"total_clips"`
	UsedClips      int               `json:"used_clips"`
	UnusedClips    int               `json:"unused_clips"`
	EffectsUsed    map[string]int    `json:"effects_used,omitempty"`
	ExportHistory  []string          `json:"export_history,omitempty"`
	Summary        string            `json:"summary"`
}

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
