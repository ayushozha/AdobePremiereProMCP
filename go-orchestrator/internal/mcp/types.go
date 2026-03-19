// Package mcp provides the MCP protocol handler for the Premiere Pro orchestrator.
// It registers all MCP tools that an AI client (Claude) can call to control
// video editing operations in Adobe Premiere Pro.
package mcp

import "context"

// Orchestrator defines the interface for all operations that MCP tools can invoke.
// Implementations coordinate between the AI client and the Premiere Pro plugin.
type Orchestrator interface {
	// Ping checks whether Premiere Pro is running and reachable.
	Ping(ctx context.Context) (*PingResult, error)

	// GetProject returns the current project state including sequences and bins.
	GetProject(ctx context.Context) (*ProjectState, error)

	// CreateSequence creates a new sequence in the active project.
	CreateSequence(ctx context.Context, params *CreateSequenceParams) (*SequenceResult, error)

	// ImportMedia imports one or more media files into the project.
	ImportMedia(ctx context.Context, params *ImportMediaParams) (*ImportResult, error)

	// PlaceClip places a clip onto the timeline at a specified position.
	PlaceClip(ctx context.Context, params *PlaceClipParams) (*ClipResult, error)

	// RemoveClip removes a clip from a sequence.
	RemoveClip(ctx context.Context, clipID, sequenceID string) error

	// AddTransition adds a transition effect between clips on the timeline.
	AddTransition(ctx context.Context, params *TransitionParams) error

	// AddText adds a text overlay to the timeline.
	AddText(ctx context.Context, params *TextParams) (*ClipResult, error)

	// SetAudioLevel adjusts the audio level of a clip in decibels.
	SetAudioLevel(ctx context.Context, clipID, sequenceID string, levelDB float64) error

	// GetTimeline returns the current state of a sequence's timeline.
	GetTimeline(ctx context.Context, sequenceID string) (*TimelineState, error)

	// Export exports a sequence to a file using the specified preset.
	Export(ctx context.Context, params *ExportParams) (*ExportResult, error)

	// ScanAssets scans a directory for usable media assets.
	ScanAssets(ctx context.Context, params *ScanAssetsParams) (*ScanResult, error)

	// ParseScript parses a script file or text into structured editing instructions.
	ParseScript(ctx context.Context, params *ParseScriptParams) (*ParseResult, error)

	// AutoEdit performs a fully automated edit from a script and asset directory.
	AutoEdit(ctx context.Context, params *AutoEditParams) (*AutoEditResult, error)
}

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// PingResult holds the response from a Premiere Pro connectivity check.
type PingResult struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version"`
	Status    string `json:"status"`
}

// ProjectState describes the currently open project.
type ProjectState struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Sequences []Sequence `json:"sequences"`
	Bins      []Bin      `json:"bins"`
}

// Sequence is a lightweight representation of a Premiere Pro sequence.
type Sequence struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Duration  float64 `json:"duration"`
	FrameRate float64 `json:"frame_rate"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
}

// Bin represents a project bin (folder) that contains media items.
type Bin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	NumItems int    `json:"num_items"`
}

// SequenceResult is returned after creating a new sequence.
type SequenceResult struct {
	SequenceID string `json:"sequence_id"`
	Name       string `json:"name"`
}

// ImportResult is returned after importing media files.
type ImportResult struct {
	ImportedFiles []ImportedFile `json:"imported_files"`
	FailedFiles   []FailedFile  `json:"failed_files,omitempty"`
}

// ImportedFile describes a single successfully imported media file.
type ImportedFile struct {
	Path    string  `json:"path"`
	ClipID  string  `json:"clip_id"`
	Type    string  `json:"type"`
	Duration float64 `json:"duration"`
}

// FailedFile describes a media file that could not be imported.
type FailedFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// ClipResult is returned after placing or creating a clip on the timeline.
type ClipResult struct {
	ClipID     string  `json:"clip_id"`
	TrackIndex int     `json:"track_index"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
}

// TimelineState describes the current state of a sequence's timeline.
type TimelineState struct {
	SequenceID   string       `json:"sequence_id"`
	SequenceName string       `json:"sequence_name"`
	Duration     float64      `json:"duration"`
	VideoTracks  []TrackState `json:"video_tracks"`
	AudioTracks  []TrackState `json:"audio_tracks"`
}

// TrackState describes a single track on the timeline.
type TrackState struct {
	Index int             `json:"index"`
	Clips []ClipOnTimeline `json:"clips"`
}

// ClipOnTimeline represents a clip currently placed on a timeline track.
type ClipOnTimeline struct {
	ClipID    string  `json:"clip_id"`
	Name      string  `json:"name"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	InPoint   float64 `json:"in_point"`
	OutPoint  float64 `json:"out_point"`
}

// ExportResult is returned after an export operation completes or starts.
type ExportResult struct {
	OutputPath string `json:"output_path"`
	Status     string `json:"status"`
	JobID      string `json:"job_id,omitempty"`
}

// ScanResult holds the output of an asset directory scan.
type ScanResult struct {
	Directory string      `json:"directory"`
	Assets    []AssetInfo `json:"assets"`
	Total     int         `json:"total"`
}

// AssetInfo describes a single media asset discovered during a scan.
type AssetInfo struct {
	Path      string  `json:"path"`
	FileName  string  `json:"file_name"`
	Type      string  `json:"type"`
	Extension string  `json:"extension"`
	SizeBytes int64   `json:"size_bytes"`
	Duration  float64 `json:"duration,omitempty"`
}

// ParseResult holds the output of script parsing.
type ParseResult struct {
	Title    string         `json:"title,omitempty"`
	Format   string         `json:"format"`
	Segments []ScriptSegment `json:"segments"`
}

// ScriptSegment represents one segment extracted from a parsed script.
type ScriptSegment struct {
	Index       int      `json:"index"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Duration    float64  `json:"duration,omitempty"`
	Notes       []string `json:"notes,omitempty"`
}

// AutoEditResult is returned after an automated editing session.
type AutoEditResult struct {
	SequenceID string `json:"sequence_id"`
	Name       string `json:"name"`
	Duration   float64 `json:"duration"`
	ClipsUsed  int    `json:"clips_used"`
	Status     string `json:"status"`
}

// ---------------------------------------------------------------------------
// Parameter types
// ---------------------------------------------------------------------------

// CreateSequenceParams contains the parameters for creating a new sequence.
type CreateSequenceParams struct {
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FrameRate   float64 `json:"frame_rate"`
	VideoTracks int     `json:"video_tracks"`
	AudioTracks int     `json:"audio_tracks"`
}

// ImportMediaParams contains the parameters for importing media files.
type ImportMediaParams struct {
	FilePaths []string `json:"file_paths"`
	TargetBin string   `json:"target_bin"`
}

// PlaceClipParams contains the parameters for placing a clip on the timeline.
type PlaceClipParams struct {
	SourcePath      string  `json:"source_path"`
	TrackType       string  `json:"track_type"`
	TrackIndex      int     `json:"track_index"`
	PositionSeconds float64 `json:"position_seconds"`
	InPointSeconds  float64 `json:"in_point_seconds"`
	OutPointSeconds float64 `json:"out_point_seconds"`
	Speed           float64 `json:"speed"`
}

// TransitionParams contains the parameters for adding a transition.
type TransitionParams struct {
	SequenceID      string  `json:"sequence_id"`
	TrackIndex      int     `json:"track_index"`
	PositionSeconds float64 `json:"position_seconds"`
	Type            string  `json:"type"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// TextParams contains the parameters for adding a text overlay.
type TextParams struct {
	SequenceID      string  `json:"sequence_id"`
	Text            string  `json:"text"`
	TrackIndex      int     `json:"track_index"`
	PositionSeconds float64 `json:"position_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
	FontSize        float64 `json:"font_size"`
	Color           string  `json:"color"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
}

// ExportParams contains the parameters for exporting a sequence.
type ExportParams struct {
	SequenceID string `json:"sequence_id"`
	OutputPath string `json:"output_path"`
	Preset     string `json:"preset"`
}

// ScanAssetsParams contains the parameters for scanning a directory for assets.
type ScanAssetsParams struct {
	Directory  string   `json:"directory"`
	Recursive  bool     `json:"recursive"`
	Extensions []string `json:"extensions"`
}

// ParseScriptParams contains the parameters for parsing a script.
type ParseScriptParams struct {
	FilePath string `json:"file_path"`
	Text     string `json:"text"`
	Format   string `json:"format"`
}

// AutoEditParams contains the parameters for an automated edit.
type AutoEditParams struct {
	ScriptPath      string `json:"script_path"`
	ScriptText      string `json:"script_text"`
	AssetsDirectory string `json:"assets_directory"`
	OutputName      string `json:"output_name"`
	Resolution      string `json:"resolution"`
	Pacing          string `json:"pacing"`
}
