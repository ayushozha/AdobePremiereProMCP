// media_client.go wraps the Rust MediaEngineService gRPC RPCs using
// the real generated proto stubs.
package grpc

import (
	"context"
	"fmt"
	"time"

	commonpb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/common/v1"
	mediapb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/media/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// MediaEngineClient provides Go-native access to the Rust media engine service.
type MediaEngineClient struct {
	conn        *grpc.ClientConn
	client      mediapb.MediaEngineServiceClient
	callTimeout time.Duration
	logger      *zap.Logger
}

// newMediaEngineClient dials the media engine service and returns a ready client.
func newMediaEngineClient(addr string, dialTimeout, callTimeout time.Duration, logger *zap.Logger) (*MediaEngineClient, error) {
	logger = logger.With(zap.String("client", "media_engine"), zap.String("addr", addr))
	logger.Info("connecting to media engine service")

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("media engine dial %s: %w", addr, err)
	}

	logger.Info("connected to media engine service")
	return &MediaEngineClient{
		conn:        conn,
		client:      mediapb.NewMediaEngineServiceClient(conn),
		callTimeout: callTimeout,
		logger:      logger,
	}, nil
}

// close shuts down the underlying gRPC connection.
func (c *MediaEngineClient) close() error {
	c.logger.Info("closing media engine connection")
	return c.conn.Close()
}

// callCtx derives a context with the configured call timeout.
func (c *MediaEngineClient) callCtx(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		// Caller already set a deadline; respect it.
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, c.callTimeout)
}

// ---------------------------------------------------------------------------
// RPC wrappers
// ---------------------------------------------------------------------------

// ScanAssets scans a directory for media files and returns indexed assets.
func (c *MediaEngineClient) ScanAssets(ctx context.Context, params ScanAssetsParams) (*ScanAssetsResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ScanAssets",
		zap.String("directory", params.Directory),
		zap.Bool("recursive", params.Recursive),
		zap.Strings("extensions", params.Extensions),
	)

	req := &mediapb.ScanAssetsRequest{
		Directory:  params.Directory,
		Recursive:  params.Recursive,
		Extensions: params.Extensions,
	}
	resp, err := c.client.ScanAssets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ScanAssets rpc: %w", err)
	}

	assets := make([]Asset, len(resp.GetAssets()))
	for i, a := range resp.GetAssets() {
		assets[i] = protoAssetToNative(a)
	}
	return &ScanAssetsResult{
		Assets:              assets,
		TotalFilesScanned:   resp.GetTotalFilesScanned(),
		MediaFilesFound:     resp.GetMediaFilesFound(),
		ScanDurationSeconds: resp.GetScanDurationSeconds(),
	}, nil
}

// ProbeMedia inspects a single media file and returns detailed metadata.
func (c *MediaEngineClient) ProbeMedia(ctx context.Context, params ProbeMediaParams) (*ProbeMediaResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ProbeMedia", zap.String("file_path", params.FilePath))

	req := &mediapb.ProbeMediaRequest{FilePath: params.FilePath}
	resp, err := c.client.ProbeMedia(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ProbeMedia rpc: %w", err)
	}

	return &ProbeMediaResult{
		Asset: protoAssetToNative(resp.GetAsset()),
	}, nil
}

// GenerateThumbnail extracts a thumbnail image from a video at the given timestamp.
func (c *MediaEngineClient) GenerateThumbnail(ctx context.Context, params GenerateThumbnailParams) (*GenerateThumbnailResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GenerateThumbnail",
		zap.String("file_path", params.FilePath),
		zap.String("format", params.OutputFormat),
	)

	req := &mediapb.GenerateThumbnailRequest{
		FilePath:     params.FilePath,
		Timestamp:    nativeTimecodeToProto(params.Timestamp),
		OutputSize:   nativeResolutionToProto(params.OutputSize),
		OutputFormat: params.OutputFormat,
	}
	resp, err := c.client.GenerateThumbnail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GenerateThumbnail rpc: %w", err)
	}

	result := &GenerateThumbnailResult{
		ThumbnailData: resp.GetThumbnailData(),
		OutputPath:    resp.GetOutputPath(),
	}
	if resp.GetActualSize() != nil {
		result.ActualSize = protoResolutionToNative(resp.GetActualSize())
	}
	return result, nil
}

// AnalyzeWaveform analyses the audio waveform of a file and detects silence regions.
func (c *MediaEngineClient) AnalyzeWaveform(ctx context.Context, params AnalyzeWaveformParams) (*AnalyzeWaveformResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("AnalyzeWaveform",
		zap.String("file_path", params.FilePath),
		zap.Uint32("audio_track", params.AudioTrack),
		zap.Float64("silence_threshold_db", params.SilenceThresholdDB),
	)

	req := &mediapb.AnalyzeWaveformRequest{
		FilePath:                  params.FilePath,
		AudioTrack:                params.AudioTrack,
		SilenceThresholdDb:        params.SilenceThresholdDB,
		MinSilenceDurationSeconds: params.MinSilenceDurationSeconds,
	}
	resp, err := c.client.AnalyzeWaveform(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AnalyzeWaveform rpc: %w", err)
	}

	regions := make([]SilenceRegion, len(resp.GetSilenceRegions()))
	for i, r := range resp.GetSilenceRegions() {
		regions[i] = SilenceRegion{
			StartSeconds: r.GetStartSeconds(),
			EndSeconds:   r.GetEndSeconds(),
			AvgDB:        r.GetAvgDb(),
		}
	}
	return &AnalyzeWaveformResult{
		SilenceRegions:  regions,
		PeakDB:          resp.GetPeakDb(),
		RmsDB:           resp.GetRmsDb(),
		DurationSeconds: resp.GetDurationSeconds(),
		WaveformSamples: resp.GetWaveformSamples(),
	}, nil
}

// DetectScenes detects scene change boundaries in a video file.
func (c *MediaEngineClient) DetectScenes(ctx context.Context, params DetectScenesParams) (*DetectScenesResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("DetectScenes",
		zap.String("file_path", params.FilePath),
		zap.Float64("threshold", params.Threshold),
	)

	req := &mediapb.DetectScenesRequest{
		FilePath:  params.FilePath,
		Threshold: params.Threshold,
	}
	resp, err := c.client.DetectScenes(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("DetectScenes rpc: %w", err)
	}

	scenes := make([]SceneChange, len(resp.GetScenes()))
	for i, s := range resp.GetScenes() {
		scenes[i] = SceneChange{
			Timecode:   protoTimecodeToNative(s.GetTimecode()),
			Confidence: s.GetConfidence(),
		}
	}
	return &DetectScenesResult{
		Scenes: scenes,
	}, nil
}

// ---------------------------------------------------------------------------
// Proto <-> Native conversion helpers
// ---------------------------------------------------------------------------

func nativeTimecodeToProto(tc Timecode) *commonpb.Timecode {
	return &commonpb.Timecode{
		Hours:     tc.Hours,
		Minutes:   tc.Minutes,
		Seconds:   tc.Seconds,
		Frames:    tc.Frames,
		FrameRate: tc.FrameRate,
	}
}

func protoTimecodeToNative(tc *commonpb.Timecode) Timecode {
	if tc == nil {
		return Timecode{}
	}
	return Timecode{
		Hours:     tc.GetHours(),
		Minutes:   tc.GetMinutes(),
		Seconds:   tc.GetSeconds(),
		Frames:    tc.GetFrames(),
		FrameRate: tc.GetFrameRate(),
	}
}

func nativeResolutionToProto(r Resolution) *commonpb.Resolution {
	return &commonpb.Resolution{
		Width:  r.Width,
		Height: r.Height,
	}
}

func protoResolutionToNative(r *commonpb.Resolution) Resolution {
	if r == nil {
		return Resolution{}
	}
	return Resolution{
		Width:  r.GetWidth(),
		Height: r.GetHeight(),
	}
}

func nativeTimeRangeToProto(tr *TimeRange) *commonpb.TimeRange {
	if tr == nil {
		return nil
	}
	return &commonpb.TimeRange{
		InPoint:  nativeTimecodeToProto(tr.InPoint),
		OutPoint: nativeTimecodeToProto(tr.OutPoint),
	}
}

func protoTimeRangeToNative(tr *commonpb.TimeRange) *TimeRange {
	if tr == nil {
		return nil
	}
	return &TimeRange{
		InPoint:  protoTimecodeToNative(tr.GetInPoint()),
		OutPoint: protoTimecodeToNative(tr.GetOutPoint()),
	}
}

func nativeTrackTargetToProto(t TrackTarget) *commonpb.TrackTarget {
	return &commonpb.TrackTarget{
		Type:       commonpb.TrackType(t.Type),
		TrackIndex: t.TrackIndex,
	}
}

func protoAssetToNative(a *commonpb.Asset) Asset {
	if a == nil {
		return Asset{}
	}
	asset := Asset{
		ID:            a.GetId(),
		FilePath:      a.GetFilePath(),
		FileName:      a.GetFileName(),
		FileSizeBytes: a.GetFileSizeBytes(),
		MimeType:      a.GetMimeType(),
		AssetType:     AssetType(a.GetAssetType()),
		Metadata:      a.GetMetadata(),
		Fingerprint:   a.GetFingerprint(),
	}
	if a.GetVideo() != nil {
		asset.Video = &VideoInfo{
			Codec:           a.GetVideo().GetCodec(),
			Resolution:      protoResolutionToNative(a.GetVideo().GetResolution()),
			FrameRate:       a.GetVideo().GetFrameRate(),
			BitrateBps:      a.GetVideo().GetBitrateBps(),
			PixelFormat:     a.GetVideo().GetPixelFormat(),
			DurationSeconds: a.GetVideo().GetDurationSeconds(),
		}
	}
	if a.GetAudio() != nil {
		asset.Audio = &AudioInfo{
			Codec:           a.GetAudio().GetCodec(),
			SampleRate:      a.GetAudio().GetSampleRate(),
			Channels:        a.GetAudio().GetChannels(),
			BitrateBps:      a.GetAudio().GetBitrateBps(),
			DurationSeconds: a.GetAudio().GetDurationSeconds(),
		}
	}
	return asset
}

func nativeAssetToProto(a Asset) *commonpb.Asset {
	asset := &commonpb.Asset{
		Id:            a.ID,
		FilePath:      a.FilePath,
		FileName:      a.FileName,
		FileSizeBytes: a.FileSizeBytes,
		MimeType:      a.MimeType,
		AssetType:     commonpb.AssetType(a.AssetType),
		Metadata:      a.Metadata,
		Fingerprint:   a.Fingerprint,
	}
	if a.Video != nil {
		asset.Video = &commonpb.VideoInfo{
			Codec:           a.Video.Codec,
			Resolution:      nativeResolutionToProto(a.Video.Resolution),
			FrameRate:       a.Video.FrameRate,
			BitrateBps:      a.Video.BitrateBps,
			PixelFormat:     a.Video.PixelFormat,
			DurationSeconds: a.Video.DurationSeconds,
		}
	}
	if a.Audio != nil {
		asset.Audio = &commonpb.AudioInfo{
			Codec:           a.Audio.Codec,
			SampleRate:      a.Audio.SampleRate,
			Channels:        a.Audio.Channels,
			BitrateBps:      a.Audio.BitrateBps,
			DurationSeconds: a.Audio.DurationSeconds,
		}
	}
	return asset
}

func nativeEDLToProto(edl EditDecisionList) *commonpb.EditDecisionList {
	entries := make([]*commonpb.EDLEntry, len(edl.Entries))
	for i, e := range edl.Entries {
		entry := &commonpb.EDLEntry{
			Index:         e.Index,
			SourceAssetId: e.SourceAssetID,
			SourceRange:   nativeTimeRangeToProto(&e.SourceRange),
			TimelineRange: nativeTimeRangeToProto(&e.TimelineRange),
			Track:         nativeTrackTargetToProto(e.Track),
			Notes:         e.Notes,
		}
		if e.Transition != nil {
			entry.Transition = &commonpb.TransitionInfo{
				Type:            e.Transition.Type,
				DurationSeconds: e.Transition.DurationSeconds,
				Alignment:       e.Transition.Alignment,
			}
		}
		if len(e.Effects) > 0 {
			effects := make([]*commonpb.EffectInfo, len(e.Effects))
			for j, ef := range e.Effects {
				effects[j] = &commonpb.EffectInfo{
					Name:       ef.Name,
					Parameters: ef.Parameters,
				}
			}
			entry.Effects = effects
		}
		entries[i] = entry
	}
	return &commonpb.EditDecisionList{
		Id:                 edl.ID,
		Name:               edl.Name,
		SequenceResolution: nativeResolutionToProto(edl.SequenceResolution),
		SequenceFrameRate:  edl.SequenceFrameRate,
		Entries:            entries,
	}
}

func protoEDLToNative(edl *commonpb.EditDecisionList) EditDecisionList {
	if edl == nil {
		return EditDecisionList{}
	}
	entries := make([]EDLEntry, len(edl.GetEntries()))
	for i, e := range edl.GetEntries() {
		entry := EDLEntry{
			Index:         e.GetIndex(),
			SourceAssetID: e.GetSourceAssetId(),
			Track: TrackTarget{
				Type:       TrackType(e.GetTrack().GetType()),
				TrackIndex: e.GetTrack().GetTrackIndex(),
			},
			Notes: e.GetNotes(),
		}
		if e.GetSourceRange() != nil {
			sr := protoTimeRangeToNative(e.GetSourceRange())
			entry.SourceRange = *sr
		}
		if e.GetTimelineRange() != nil {
			tr := protoTimeRangeToNative(e.GetTimelineRange())
			entry.TimelineRange = *tr
		}
		if e.GetTransition() != nil {
			entry.Transition = &TransitionInfo{
				Type:            e.GetTransition().GetType(),
				DurationSeconds: e.GetTransition().GetDurationSeconds(),
				Alignment:       e.GetTransition().GetAlignment(),
			}
		}
		if len(e.GetEffects()) > 0 {
			effects := make([]EffectInfo, len(e.GetEffects()))
			for j, ef := range e.GetEffects() {
				effects[j] = EffectInfo{
					Name:       ef.GetName(),
					Parameters: ef.GetParameters(),
				}
			}
			entry.Effects = effects
		}
		entries[i] = entry
	}
	return EditDecisionList{
		ID:                 edl.GetId(),
		Name:               edl.GetName(),
		SequenceResolution: protoResolutionToNative(edl.GetSequenceResolution()),
		SequenceFrameRate:  edl.GetSequenceFrameRate(),
		Entries:            entries,
	}
}
