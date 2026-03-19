// Package mcp provides the MCP protocol handler for the Premiere Pro orchestrator.
// It registers all MCP tools that an AI client (Claude) can call to control
// video editing operations in Adobe Premiere Pro.
//
// This package re-exports the orchestrator.Orchestrator interface so that
// callers only need to pass an *orchestrator.Engine.
package mcp

import (
	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/orchestrator"
)

// Orchestrator is an alias for the canonical interface in the orchestrator package.
// The MCP tool handlers call these methods and translate results to MCP tool responses.
type Orchestrator = orchestrator.Orchestrator

// Re-export the orchestrator types used by tool handlers so that the tools.go
// file can refer to them without a package prefix.
type (
	PingResult           = orchestrator.PingResult
	ProjectState         = orchestrator.ProjectState
	CreateSequenceParams = orchestrator.CreateSequenceParams
	Resolution           = orchestrator.Resolution
	SequenceResult       = orchestrator.SequenceResult
	ImportResult         = orchestrator.ImportResult
	PlaceClipParams      = orchestrator.PlaceClipParams
	TrackTarget          = orchestrator.TrackTarget
	TrackType            = orchestrator.TrackType
	Timecode             = orchestrator.Timecode
	TimeRange            = orchestrator.TimeRange
	ClipResult           = orchestrator.ClipResult
	TransitionParams     = orchestrator.TransitionParams
	TextParams           = orchestrator.TextParams
	TextStyle            = orchestrator.TextStyle
	Position             = orchestrator.Position
	TimelineState        = orchestrator.TimelineState
	ExportParams         = orchestrator.ExportParams
	ExportPreset         = orchestrator.ExportPreset
	ExportResult         = orchestrator.ExportResult
	ScanResult           = orchestrator.ScanResult
	AssetInfo            = orchestrator.AssetInfo
	ParsedScript         = orchestrator.ParsedScript
	ScriptSegment        = orchestrator.ScriptSegment
	AutoEditParams       = orchestrator.AutoEditParams
	AutoEditResult       = orchestrator.AutoEditResult
	EDLSettings          = orchestrator.EDLSettings

	// Sequence management types
	SequenceSettings          = orchestrator.SequenceSettings
	SetSequenceSettingsParams = orchestrator.SetSequenceSettingsParams
	SequenceListResult        = orchestrator.SequenceListResult
	SequenceListEntry         = orchestrator.SequenceListEntry
	PlayheadResult            = orchestrator.PlayheadResult
	InOutPointsResult         = orchestrator.InOutPointsResult
	MarkerInfo                = orchestrator.MarkerInfo
	MarkersResult             = orchestrator.MarkersResult
	AddMarkerParams           = orchestrator.AddMarkerParams
	GenericResult             = orchestrator.GenericResult

	// Export & render extended types
	ExportDirectParams     = orchestrator.ExportDirectParams
	ExportViaAMEParams     = orchestrator.ExportViaAMEParams
	ExportFrameParams      = orchestrator.ExportFrameParams
	ExportAAFParams        = orchestrator.ExportAAFParams
	ExportOMFParams        = orchestrator.ExportOMFParams
	ExportAudioOnlyParams  = orchestrator.ExportAudioOnlyParams
	RenderPreviewParams    = orchestrator.RenderPreviewParams
	ExporterListResult     = orchestrator.ExporterListResult
	ExportPresetListResult = orchestrator.ExportPresetListResult
	ExportProgressResult   = orchestrator.ExportProgressResult
	GenericExportResult    = orchestrator.GenericExportResult
)
