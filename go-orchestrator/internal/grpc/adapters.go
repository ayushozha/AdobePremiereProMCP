// adapters.go provides thin adapter types that make the concrete gRPC client
// types satisfy the orchestrator-package interfaces. Each adapter translates
// between orchestrator-native types and the grpc-package types used by the
// underlying client methods.
package grpc

import (
	"context"

	orch "github.com/anthropics/premierpro-mcp/go-orchestrator/internal/orchestrator"
)

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

var _ orch.PremiereClient = (*PremiereAdapter)(nil)
var _ orch.MediaClient = (*MediaAdapter)(nil)
var _ orch.IntelClient = (*IntelAdapter)(nil)

// ---------------------------------------------------------------------------
// PremiereAdapter wraps *PremiereBridgeClient to satisfy orch.PremiereClient.
// ---------------------------------------------------------------------------

// PremiereAdapter adapts PremiereBridgeClient to the orchestrator.PremiereClient interface.
type PremiereAdapter struct {
	C *PremiereBridgeClient
}

func (a *PremiereAdapter) Ping(ctx context.Context) (*orch.PingResult, error) {
	res, err := a.C.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &orch.PingResult{
		PremiereRunning: res.PremiereRunning,
		PremiereVersion: res.PremiereVersion,
		ProjectOpen:     res.ProjectOpen,
		BridgeMode:      res.BridgeMode,
	}, nil
}

func (a *PremiereAdapter) GetProjectState(ctx context.Context) (*orch.ProjectState, error) {
	res, err := a.C.GetProjectState(ctx)
	if err != nil {
		return nil, err
	}
	seqs := make([]*orch.SequenceInfo, len(res.Sequences))
	for i, s := range res.Sequences {
		seqs[i] = &orch.SequenceInfo{
			ID:              s.ID,
			Name:            s.Name,
			Resolution:      orch.Resolution{Width: s.Resolution.Width, Height: s.Resolution.Height},
			FrameRate:       s.FrameRate,
			DurationSeconds: s.DurationSeconds,
			VideoTrackCount: s.VideoTrackCount,
			AudioTrackCount: s.AudioTrackCount,
		}
	}
	return &orch.ProjectState{
		ProjectName: res.ProjectName,
		ProjectPath: res.ProjectPath,
		Sequences:   seqs,
		BinCount:    res.BinCount,
		IsSaved:     res.IsSaved,
	}, nil
}

func (a *PremiereAdapter) CreateSequence(ctx context.Context, params *orch.CreateSequenceParams) (*orch.SequenceResult, error) {
	grpcParams := CreateSequenceParams{
		Name:        params.Name,
		Resolution:  Resolution{Width: params.Resolution.Width, Height: params.Resolution.Height},
		FrameRate:   params.FrameRate,
		VideoTracks: params.VideoTracks,
		AudioTracks: params.AudioTracks,
	}
	res, err := a.C.CreateSequence(ctx, grpcParams)
	if err != nil {
		return nil, err
	}
	return &orch.SequenceResult{
		SequenceID: res.SequenceID,
		Name:       res.Name,
	}, nil
}

func (a *PremiereAdapter) ImportMedia(ctx context.Context, filePath string, targetBin string) (*orch.ImportResult, error) {
	grpcParams := ImportMediaParams{
		FilePath:  filePath,
		TargetBin: targetBin,
	}
	res, err := a.C.ImportMedia(ctx, grpcParams)
	if err != nil {
		return nil, err
	}
	return &orch.ImportResult{
		ProjectItemID: res.ProjectItemID,
		Name:          res.Name,
	}, nil
}

func (a *PremiereAdapter) PlaceClip(ctx context.Context, params *orch.PlaceClipParams) (*orch.ClipResult, error) {
	grpcParams := PlaceClipParams{
		SourcePath: params.SourcePath,
		Track: TrackTarget{
			Type:       TrackType(params.Track.Type),
			TrackIndex: params.Track.TrackIndex,
		},
		Position: Timecode{
			Hours:     params.Position.Hours,
			Minutes:   params.Position.Minutes,
			Seconds:   params.Position.Seconds,
			Frames:    params.Position.Frames,
			FrameRate: params.Position.FrameRate,
		},
		Speed: params.Speed,
	}
	if params.SourceRange != nil {
		grpcParams.SourceRange = &TimeRange{
			InPoint: Timecode{
				Hours:     params.SourceRange.InPoint.Hours,
				Minutes:   params.SourceRange.InPoint.Minutes,
				Seconds:   params.SourceRange.InPoint.Seconds,
				Frames:    params.SourceRange.InPoint.Frames,
				FrameRate: params.SourceRange.InPoint.FrameRate,
			},
			OutPoint: Timecode{
				Hours:     params.SourceRange.OutPoint.Hours,
				Minutes:   params.SourceRange.OutPoint.Minutes,
				Seconds:   params.SourceRange.OutPoint.Seconds,
				Frames:    params.SourceRange.OutPoint.Frames,
				FrameRate: params.SourceRange.OutPoint.FrameRate,
			},
		}
	}
	res, err := a.C.PlaceClip(ctx, grpcParams)
	if err != nil {
		return nil, err
	}
	return &orch.ClipResult{
		ClipID: res.ClipID,
	}, nil
}

func (a *PremiereAdapter) RemoveClip(ctx context.Context, clipID, sequenceID string) error {
	return a.C.RemoveClip(ctx, RemoveClipParams{
		ClipID:     clipID,
		SequenceID: sequenceID,
	})
}

func (a *PremiereAdapter) AddTransition(ctx context.Context, params *orch.TransitionParams) error {
	grpcParams := AddTransitionParams{
		SequenceID: params.SequenceID,
		Track: TrackTarget{
			Type:       TrackType(params.Track.Type),
			TrackIndex: params.Track.TrackIndex,
		},
		Position: Timecode{
			Hours:     params.Position.Hours,
			Minutes:   params.Position.Minutes,
			Seconds:   params.Position.Seconds,
			Frames:    params.Position.Frames,
			FrameRate: params.Position.FrameRate,
		},
		TransitionType:  params.TransitionType,
		DurationSeconds: params.DurationSeconds,
	}
	_, err := a.C.AddTransition(ctx, grpcParams)
	return err
}

func (a *PremiereAdapter) AddText(ctx context.Context, params *orch.TextParams) (*orch.ClipResult, error) {
	grpcParams := AddTextParams{
		SequenceID: params.SequenceID,
		Text:       params.Text,
		Style: TextStyle{
			FontFamily:         params.Style.FontFamily,
			FontSize:           params.Style.FontSize,
			ColorHex:           params.Style.ColorHex,
			Alignment:          params.Style.Alignment,
			BackgroundColorHex: params.Style.BackgroundColorHex,
			BackgroundOpacity:  params.Style.BackgroundOpacity,
			Position:           Position{X: params.Style.Position.X, Y: params.Style.Position.Y},
		},
		Track: TrackTarget{
			Type:       TrackType(params.Track.Type),
			TrackIndex: params.Track.TrackIndex,
		},
		Position: Timecode{
			Hours:     params.Position.Hours,
			Minutes:   params.Position.Minutes,
			Seconds:   params.Position.Seconds,
			Frames:    params.Position.Frames,
			FrameRate: params.Position.FrameRate,
		},
		DurationSeconds: params.DurationSeconds,
	}
	res, err := a.C.AddText(ctx, grpcParams)
	if err != nil {
		return nil, err
	}
	return &orch.ClipResult{
		ClipID: res.ClipID,
	}, nil
}

func (a *PremiereAdapter) SetAudioLevel(ctx context.Context, clipID, sequenceID string, levelDB float64) error {
	return a.C.SetAudioLevel(ctx, SetAudioLevelParams{
		ClipID:     clipID,
		SequenceID: sequenceID,
		LevelDB:    levelDB,
	})
}

func (a *PremiereAdapter) GetTimelineState(ctx context.Context, sequenceID string) (*orch.TimelineState, error) {
	res, err := a.C.GetTimelineState(ctx, GetTimelineStateParams{SequenceID: sequenceID})
	if err != nil {
		return nil, err
	}
	return &orch.TimelineState{
		SequenceID:           res.SequenceID,
		VideoTracks:          convertTimelineTracks(res.VideoTracks),
		AudioTracks:          convertTimelineTracks(res.AudioTracks),
		TotalDurationSeconds: res.TotalDurationSeconds,
	}, nil
}

func (a *PremiereAdapter) ExportSequence(ctx context.Context, params *orch.ExportParams) (*orch.ExportResult, error) {
	grpcParams := ExportSequenceParams{
		SequenceID: params.SequenceID,
		OutputPath: params.OutputPath,
		Preset:     ExportPreset(params.Preset),
	}
	res, err := a.C.ExportSequence(ctx, grpcParams)
	if err != nil {
		return nil, err
	}
	return &orch.ExportResult{
		JobID:      res.JobID,
		Status:     operationStatusToString(res.Status),
		OutputPath: res.OutputPath,
	}, nil
}

func (a *PremiereAdapter) ExecuteEDL(ctx context.Context, edl *orch.EDL) (*orch.EDLExecutionResult, error) {
	grpcEDL := convertOrchestratorEDLToGRPC(edl)
	grpcParams := ExecuteEDLParams{
		EDL:                grpcEDL,
		AutoImport:         true,
		AutoCreateSequence: true,
	}
	res, err := a.C.ExecuteEDL(ctx, grpcParams)
	if err != nil {
		return nil, err
	}
	return &orch.EDLExecutionResult{
		SequenceID:       res.SequenceID,
		Status:           operationStatusToString(res.Status),
		ClipsPlaced:      res.ClipsPlaced,
		TransitionsAdded: res.TransitionsAdded,
		Errors:           res.Errors,
		Warnings:         res.Warnings,
	}, nil
}

func (a *PremiereAdapter) EvalAudioCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error) {
	return a.C.EvalAudioCommand(ctx, command, args)
}

// ---------------------------------------------------------------------------
// MediaAdapter wraps *MediaEngineClient to satisfy orch.MediaClient.
// ---------------------------------------------------------------------------

// MediaAdapter adapts MediaEngineClient to the orchestrator.MediaClient interface.
type MediaAdapter struct {
	C *MediaEngineClient
}

func (a *MediaAdapter) ScanAssets(ctx context.Context, dir string, recursive bool, extensions []string) (*orch.ScanResult, error) {
	res, err := a.C.ScanAssets(ctx, ScanAssetsParams{
		Directory:  dir,
		Recursive:  recursive,
		Extensions: extensions,
	})
	if err != nil {
		return nil, err
	}
	assets := make([]*orch.AssetInfo, len(res.Assets))
	for i, a := range res.Assets {
		assets[i] = convertGRPCAssetToOrchestrator(&a)
	}
	return &orch.ScanResult{
		Assets:              assets,
		TotalFilesScanned:   res.TotalFilesScanned,
		MediaFilesFound:     res.MediaFilesFound,
		ScanDurationSeconds: res.ScanDurationSeconds,
	}, nil
}

func (a *MediaAdapter) ProbeMedia(ctx context.Context, filePath string) (*orch.AssetInfo, error) {
	res, err := a.C.ProbeMedia(ctx, ProbeMediaParams{FilePath: filePath})
	if err != nil {
		return nil, err
	}
	return convertGRPCAssetToOrchestrator(&res.Asset), nil
}

func (a *MediaAdapter) AnalyzeWaveform(ctx context.Context, filePath string, opts *orch.WaveformOptions) (*orch.WaveformResult, error) {
	var params AnalyzeWaveformParams
	params.FilePath = filePath
	if opts != nil {
		params.AudioTrack = opts.AudioTrack
		params.SilenceThresholdDB = opts.SilenceThresholdDB
		params.MinSilenceDurationSeconds = opts.MinSilenceDurationSecs
	}
	res, err := a.C.AnalyzeWaveform(ctx, params)
	if err != nil {
		return nil, err
	}
	regions := make([]*orch.SilenceRegion, len(res.SilenceRegions))
	for i, r := range res.SilenceRegions {
		regions[i] = &orch.SilenceRegion{
			StartSeconds: r.StartSeconds,
			EndSeconds:   r.EndSeconds,
			AvgDB:        r.AvgDB,
		}
	}
	return &orch.WaveformResult{
		SilenceRegions:  regions,
		PeakDB:          res.PeakDB,
		RMSDB:           res.RmsDB,
		DurationSeconds: res.DurationSeconds,
		WaveformSamples: res.WaveformSamples,
	}, nil
}

func (a *MediaAdapter) DetectScenes(ctx context.Context, filePath string, threshold float64) (*orch.SceneResult, error) {
	res, err := a.C.DetectScenes(ctx, DetectScenesParams{
		FilePath:  filePath,
		Threshold: threshold,
	})
	if err != nil {
		return nil, err
	}
	scenes := make([]*orch.SceneChange, len(res.Scenes))
	for i, s := range res.Scenes {
		scenes[i] = &orch.SceneChange{
			TimecodeSeconds: s.Timecode.FrameRate, // approximate seconds from timecode
			Confidence:      s.Confidence,
		}
	}
	return &orch.SceneResult{
		Scenes: scenes,
	}, nil
}

// ---------------------------------------------------------------------------
// IntelAdapter wraps *IntelligenceClient to satisfy orch.IntelClient.
// ---------------------------------------------------------------------------

// IntelAdapter adapts IntelligenceClient to the orchestrator.IntelClient interface.
type IntelAdapter struct {
	C *IntelligenceClient
}

func (a *IntelAdapter) ParseScript(ctx context.Context, text, filePath, format string) (*orch.ParsedScript, error) {
	res, err := a.C.ParseScript(ctx, ParseScriptParams{
		Text:       text,
		FilePath:   filePath,
		FormatHint: format,
	})
	if err != nil {
		return nil, err
	}
	segments := make([]*orch.ScriptSegment, len(res.Segments))
	for i, s := range res.Segments {
		segments[i] = &orch.ScriptSegment{
			Index:                    s.Index,
			Type:                     orch.SegmentType(s.Type),
			Content:                  s.Content,
			Speaker:                  s.Speaker,
			SceneDescription:         s.SceneDescription,
			VisualDirection:          s.VisualDirection,
			AudioDirection:           s.AudioDirection,
			EstimatedDurationSeconds: s.EstimatedDurationSeconds,
			AssetHints:               s.AssetHints,
		}
	}
	return &orch.ParsedScript{
		Segments: segments,
		Metadata: &orch.ScriptMetadata{
			Title:                         res.Metadata.Title,
			Format:                        res.Metadata.Format,
			EstimatedTotalDurationSeconds: res.Metadata.EstimatedTotalDurationSeconds,
			SegmentCount:                  res.Metadata.SegmentCount,
		},
	}, nil
}

func (a *IntelAdapter) GenerateEDL(ctx context.Context, segments []*orch.ScriptSegment, assets []*orch.AssetInfo, settings *orch.EDLSettings) (*orch.EDL, error) {
	grpcSegments := make([]ScriptSegment, len(segments))
	for i, s := range segments {
		grpcSegments[i] = ScriptSegment{
			Index:                    s.Index,
			Type:                     SegmentType(s.Type),
			Content:                  s.Content,
			Speaker:                  s.Speaker,
			SceneDescription:         s.SceneDescription,
			VisualDirection:          s.VisualDirection,
			AudioDirection:           s.AudioDirection,
			EstimatedDurationSeconds: s.EstimatedDurationSeconds,
			AssetHints:               s.AssetHints,
		}
	}
	grpcAssets := make([]Asset, len(assets))
	for i, a := range assets {
		grpcAssets[i] = convertOrchestratorAssetToGRPC(a)
	}

	var grpcSettings EDLSettings
	if settings != nil {
		grpcSettings = EDLSettings{
			Resolution:                Resolution{Width: settings.Resolution.Width, Height: settings.Resolution.Height},
			FrameRate:                 settings.FrameRate,
			DefaultTransition:         settings.DefaultTransition,
			DefaultTransitionDuration: settings.DefaultTransitionDuration,
			Pacing:                    PacingPreset(settings.Pacing),
		}
	}

	res, err := a.C.GenerateEDL(ctx, GenerateEDLParams{
		Segments:        grpcSegments,
		AvailableAssets: grpcAssets,
		Settings:        grpcSettings,
	})
	if err != nil {
		return nil, err
	}
	return convertGRPCEDLToOrchestrator(&res.EDL), nil
}

func (a *IntelAdapter) MatchAssets(ctx context.Context, segments []*orch.ScriptSegment, assets []*orch.AssetInfo, strategy string) (*orch.MatchResult, error) {
	grpcSegments := make([]ScriptSegment, len(segments))
	for i, s := range segments {
		grpcSegments[i] = ScriptSegment{
			Index:                    s.Index,
			Type:                     SegmentType(s.Type),
			Content:                  s.Content,
			Speaker:                  s.Speaker,
			SceneDescription:         s.SceneDescription,
			VisualDirection:          s.VisualDirection,
			AudioDirection:           s.AudioDirection,
			EstimatedDurationSeconds: s.EstimatedDurationSeconds,
			AssetHints:               s.AssetHints,
		}
	}
	grpcAssets := make([]Asset, len(assets))
	for i, a := range assets {
		grpcAssets[i] = convertOrchestratorAssetToGRPC(a)
	}

	grpcStrategy := matchStrategyFromString(strategy)

	res, err := a.C.MatchAssets(ctx, MatchAssetsParams{
		Segments:        grpcSegments,
		AvailableAssets: grpcAssets,
		Strategy:        grpcStrategy,
	})
	if err != nil {
		return nil, err
	}

	matches := make([]*orch.AssetMatch, len(res.Matches))
	for i, m := range res.Matches {
		matches[i] = &orch.AssetMatch{
			SegmentIndex: m.SegmentIndex,
			AssetID:      m.AssetID,
			Confidence:   m.Confidence,
			Reasoning:    m.Reasoning,
		}
		if m.SuggestedRange != nil {
			matches[i].SuggestedRange = convertGRPCTimeRangeToOrchestrator(m.SuggestedRange)
		}
	}
	unmatched := make([]*orch.UnmatchedSegment, len(res.Unmatched))
	for i, u := range res.Unmatched {
		unmatched[i] = &orch.UnmatchedSegment{
			SegmentIndex: u.SegmentIndex,
			Reason:       u.Reason,
			Suggestions:  u.Suggestions,
		}
	}

	return &orch.MatchResult{
		Matches:   matches,
		Unmatched: unmatched,
	}, nil
}

func (a *IntelAdapter) AnalyzePacing(ctx context.Context, edl *orch.EDL, targetMood string) (*orch.PacingResult, error) {
	grpcEDL := convertOrchestratorEDLToGRPC(edl)
	res, err := a.C.AnalyzePacing(ctx, AnalyzePacingParams{
		EDL:        grpcEDL,
		TargetMood: targetMood,
	})
	if err != nil {
		return nil, err
	}
	adjustments := make([]*orch.PacingAdjustment, len(res.Adjustments))
	for i, adj := range res.Adjustments {
		adjustments[i] = &orch.PacingAdjustment{
			EDLEntryIndex:     adj.EDLEntryIndex,
			CurrentDuration:   adj.CurrentDuration,
			SuggestedDuration: adj.SuggestedDuration,
			Reason:            adj.Reason,
		}
	}
	return &orch.PacingResult{
		Adjustments:              adjustments,
		CurrentAvgClipDuration:   res.CurrentAvgClipDuration,
		SuggestedAvgClipDuration: res.SuggestedAvgClipDuration,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal conversion helpers
// ---------------------------------------------------------------------------

func convertGRPCAssetToOrchestrator(a *Asset) *orch.AssetInfo {
	info := &orch.AssetInfo{
		ID:            a.ID,
		FilePath:      a.FilePath,
		FileName:      a.FileName,
		FileSizeBytes: a.FileSizeBytes,
		MIMEType:      a.MimeType,
		Type:          orch.AssetType(a.AssetType),
		Metadata:      a.Metadata,
		Fingerprint:   a.Fingerprint,
	}
	if a.Video != nil {
		info.Video = &orch.VideoInfo{
			Codec:           a.Video.Codec,
			Resolution:      orch.Resolution{Width: a.Video.Resolution.Width, Height: a.Video.Resolution.Height},
			FrameRate:       a.Video.FrameRate,
			BitrateBPS:      a.Video.BitrateBps,
			PixelFormat:     a.Video.PixelFormat,
			DurationSeconds: a.Video.DurationSeconds,
		}
	}
	if a.Audio != nil {
		info.Audio = &orch.AudioInfo{
			Codec:           a.Audio.Codec,
			SampleRate:      a.Audio.SampleRate,
			Channels:        a.Audio.Channels,
			BitrateBPS:      a.Audio.BitrateBps,
			DurationSeconds: a.Audio.DurationSeconds,
		}
	}
	return info
}

func convertOrchestratorAssetToGRPC(a *orch.AssetInfo) Asset {
	asset := Asset{
		ID:            a.ID,
		FilePath:      a.FilePath,
		FileName:      a.FileName,
		FileSizeBytes: a.FileSizeBytes,
		MimeType:      a.MIMEType,
		AssetType:     AssetType(a.Type),
		Metadata:      a.Metadata,
		Fingerprint:   a.Fingerprint,
	}
	if a.Video != nil {
		asset.Video = &VideoInfo{
			Codec:           a.Video.Codec,
			Resolution:      Resolution{Width: a.Video.Resolution.Width, Height: a.Video.Resolution.Height},
			FrameRate:       a.Video.FrameRate,
			BitrateBps:      a.Video.BitrateBPS,
			PixelFormat:     a.Video.PixelFormat,
			DurationSeconds: a.Video.DurationSeconds,
		}
	}
	if a.Audio != nil {
		asset.Audio = &AudioInfo{
			Codec:           a.Audio.Codec,
			SampleRate:      a.Audio.SampleRate,
			Channels:        a.Audio.Channels,
			BitrateBps:      a.Audio.BitrateBPS,
			DurationSeconds: a.Audio.DurationSeconds,
		}
	}
	return asset
}

func convertTimelineTracks(tracks []TimelineTrack) []*orch.TimelineTrack {
	out := make([]*orch.TimelineTrack, len(tracks))
	for i, t := range tracks {
		clips := make([]*orch.TimelineClip, len(t.Clips))
		for j, c := range t.Clips {
			clip := &orch.TimelineClip{
				ClipID:     c.ClipID,
				SourcePath: c.SourcePath,
				Speed:      c.Speed,
			}
			sr := convertGRPCTimeRangeToOrchestrator(&c.SourceRange)
			clip.SourceRange = sr
			tr := convertGRPCTimeRangeToOrchestrator(&c.TimelineRange)
			clip.TimelineRange = tr
			clips[j] = clip
		}
		out[i] = &orch.TimelineTrack{
			Index:    t.Index,
			Type:     orch.TrackType(t.Type),
			Clips:    clips,
			IsMuted:  t.IsMuted,
			IsLocked: t.IsLocked,
		}
	}
	return out
}

func convertOrchestratorEDLToGRPC(edl *orch.EDL) EditDecisionList {
	entries := make([]EDLEntry, len(edl.Entries))
	for i, e := range edl.Entries {
		entry := EDLEntry{
			Index:         e.Index,
			SourceAssetID: e.SourceAssetID,
			Notes:         e.Notes,
		}
		if e.Track != nil {
			entry.Track = TrackTarget{
				Type:       TrackType(e.Track.Type),
				TrackIndex: e.Track.TrackIndex,
			}
		}
		if e.Transition != nil {
			entry.Transition = &TransitionInfo{
				Type:            e.Transition.Type,
				DurationSeconds: e.Transition.DurationSeconds,
				Alignment:       e.Transition.Alignment,
			}
		}
		if len(e.Effects) > 0 {
			effects := make([]EffectInfo, len(e.Effects))
			for j, ef := range e.Effects {
				effects[j] = EffectInfo{
					Name:       ef.Name,
					Parameters: ef.Parameters,
				}
			}
			entry.Effects = effects
		}
		entries[i] = entry
	}
	return EditDecisionList{
		ID:                 edl.ID,
		Name:               edl.Name,
		SequenceResolution: Resolution{Width: edl.SequenceResolution.Width, Height: edl.SequenceResolution.Height},
		SequenceFrameRate:  edl.SequenceFrameRate,
		Entries:            entries,
	}
}

func convertGRPCEDLToOrchestrator(edl *EditDecisionList) *orch.EDL {
	entries := make([]*orch.EDLEntry, len(edl.Entries))
	for i, e := range edl.Entries {
		entry := &orch.EDLEntry{
			Index:         e.Index,
			SourceAssetID: e.SourceAssetID,
			Notes:         e.Notes,
		}
		entry.Track = &orch.TrackTarget{
			Type:       orch.TrackType(e.Track.Type),
			TrackIndex: e.Track.TrackIndex,
		}
		if e.Transition != nil {
			entry.Transition = &orch.TransitionInfo{
				Type:            e.Transition.Type,
				DurationSeconds: e.Transition.DurationSeconds,
				Alignment:       e.Transition.Alignment,
			}
		}
		if len(e.Effects) > 0 {
			effects := make([]*orch.EffectInfo, len(e.Effects))
			for j, ef := range e.Effects {
				effects[j] = &orch.EffectInfo{
					Name:       ef.Name,
					Parameters: ef.Parameters,
				}
			}
			entry.Effects = effects
		}
		entries[i] = entry
	}
	return &orch.EDL{
		ID:                 edl.ID,
		Name:               edl.Name,
		SequenceResolution: orch.Resolution{Width: edl.SequenceResolution.Width, Height: edl.SequenceResolution.Height},
		SequenceFrameRate:  edl.SequenceFrameRate,
		Entries:            entries,
	}
}

func convertGRPCTimeRangeToOrchestrator(tr *TimeRange) *orch.TimeRange {
	return &orch.TimeRange{
		InPoint: orch.Timecode{
			Hours:     tr.InPoint.Hours,
			Minutes:   tr.InPoint.Minutes,
			Seconds:   tr.InPoint.Seconds,
			Frames:    tr.InPoint.Frames,
			FrameRate: tr.InPoint.FrameRate,
		},
		OutPoint: orch.Timecode{
			Hours:     tr.OutPoint.Hours,
			Minutes:   tr.OutPoint.Minutes,
			Seconds:   tr.OutPoint.Seconds,
			Frames:    tr.OutPoint.Frames,
			FrameRate: tr.OutPoint.FrameRate,
		},
	}
}

func operationStatusToString(s OperationStatus) string {
	switch s {
	case OperationStatusPending:
		return "pending"
	case OperationStatusRunning:
		return "running"
	case OperationStatusCompleted:
		return "completed"
	case OperationStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

func matchStrategyFromString(s string) MatchStrategy {
	switch s {
	case "keyword":
		return MatchStrategyKeyword
	case "embedding":
		return MatchStrategyEmbedding
	case "hybrid":
		return MatchStrategyHybrid
	default:
		return MatchStrategyHybrid
	}
}
