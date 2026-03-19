// premiere_client.go wraps the TypeScript PremiereBridgeService gRPC RPCs using
// the real generated proto stubs.
package grpc

import (
	"context"
	"fmt"
	"time"

	commonpb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/common/v1"
	premierepb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/premiere/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// PremiereBridgeClient provides Go-native access to the TypeScript Premiere bridge.
type PremiereBridgeClient struct {
	conn        *grpc.ClientConn
	client      premierepb.PremiereBridgeServiceClient
	callTimeout time.Duration
	logger      *zap.Logger
}

// newPremiereBridgeClient dials the Premiere bridge and returns a ready client.
func newPremiereBridgeClient(addr string, dialTimeout, callTimeout time.Duration, logger *zap.Logger) (*PremiereBridgeClient, error) {
	logger = logger.With(zap.String("client", "premiere_bridge"), zap.String("addr", addr))
	logger.Info("connecting to premiere bridge service")

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("premiere bridge dial %s: %w", addr, err)
	}

	logger.Info("connected to premiere bridge service")
	return &PremiereBridgeClient{
		conn:        conn,
		client:      premierepb.NewPremiereBridgeServiceClient(conn),
		callTimeout: callTimeout,
		logger:      logger,
	}, nil
}

// close shuts down the underlying gRPC connection.
func (c *PremiereBridgeClient) close() error {
	c.logger.Info("closing premiere bridge connection")
	return c.conn.Close()
}

// callCtx derives a context with the configured call timeout.
func (c *PremiereBridgeClient) callCtx(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, c.callTimeout)
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

// Ping checks whether Premiere Pro is running and responsive.
func (c *PremiereBridgeClient) Ping(ctx context.Context) (*PingResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("Ping")

	resp, err := c.client.Ping(ctx, &premierepb.PingRequest{})
	if err != nil {
		return nil, fmt.Errorf("Ping rpc: %w", err)
	}

	return &PingResult{
		PremiereRunning: resp.GetPremiereRunning(),
		PremiereVersion: resp.GetPremiereVersion(),
		ProjectOpen:     resp.GetProjectOpen(),
		BridgeMode:      resp.GetBridgeMode(),
	}, nil
}

// ---------------------------------------------------------------------------
// Project
// ---------------------------------------------------------------------------

// GetProjectState retrieves the current Premiere Pro project state.
func (c *PremiereBridgeClient) GetProjectState(ctx context.Context) (*GetProjectStateResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GetProjectState")

	resp, err := c.client.GetProjectState(ctx, &premierepb.GetProjectStateRequest{})
	if err != nil {
		return nil, fmt.Errorf("GetProjectState rpc: %w", err)
	}

	seqs := make([]SequenceInfo, len(resp.GetSequences()))
	for i, s := range resp.GetSequences() {
		seqs[i] = SequenceInfo{
			ID:              s.GetId(),
			Name:            s.GetName(),
			Resolution:      protoResolutionToNative(s.GetResolution()),
			FrameRate:       s.GetFrameRate(),
			DurationSeconds: s.GetDurationSeconds(),
			VideoTrackCount: s.GetVideoTrackCount(),
			AudioTrackCount: s.GetAudioTrackCount(),
		}
	}

	return &GetProjectStateResult{
		ProjectName: resp.GetProjectName(),
		ProjectPath: resp.GetProjectPath(),
		Sequences:   seqs,
		BinCount:    resp.GetBinCount(),
		IsSaved:     resp.GetIsSaved(),
	}, nil
}

// ---------------------------------------------------------------------------
// Sequence
// ---------------------------------------------------------------------------

// CreateSequence creates a new sequence in the Premiere Pro project.
func (c *PremiereBridgeClient) CreateSequence(ctx context.Context, params CreateSequenceParams) (*CreateSequenceResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("CreateSequence",
		zap.String("name", params.Name),
		zap.Uint32("width", params.Resolution.Width),
		zap.Uint32("height", params.Resolution.Height),
		zap.Float64("frame_rate", params.FrameRate),
	)

	req := &premierepb.CreateSequenceRequest{
		Name:        params.Name,
		Resolution:  nativeResolutionToProto(params.Resolution),
		FrameRate:   params.FrameRate,
		VideoTracks: params.VideoTracks,
		AudioTracks: params.AudioTracks,
	}
	resp, err := c.client.CreateSequence(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateSequence rpc: %w", err)
	}

	return &CreateSequenceResult{
		SequenceID: resp.GetSequenceId(),
		Name:       resp.GetName(),
	}, nil
}

// GetTimelineState retrieves the current timeline state of a sequence.
func (c *PremiereBridgeClient) GetTimelineState(ctx context.Context, params GetTimelineStateParams) (*GetTimelineStateResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GetTimelineState", zap.String("sequence_id", params.SequenceID))

	req := &premierepb.GetTimelineStateRequest{SequenceId: params.SequenceID}
	resp, err := c.client.GetTimelineState(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetTimelineState rpc: %w", err)
	}

	return &GetTimelineStateResult{
		SequenceID:           resp.GetSequenceId(),
		VideoTracks:          protoTimelineTracksToNative(resp.GetVideoTracks()),
		AudioTracks:          protoTimelineTracksToNative(resp.GetAudioTracks()),
		TotalDurationSeconds: resp.GetTotalDurationSeconds(),
	}, nil
}

// ---------------------------------------------------------------------------
// Clip Operations
// ---------------------------------------------------------------------------

// ImportMedia imports a media file into the Premiere Pro project.
func (c *PremiereBridgeClient) ImportMedia(ctx context.Context, params ImportMediaParams) (*ImportMediaResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ImportMedia",
		zap.String("file_path", params.FilePath),
		zap.String("target_bin", params.TargetBin),
	)

	req := &premierepb.ImportMediaRequest{
		FilePath:  params.FilePath,
		TargetBin: params.TargetBin,
	}
	resp, err := c.client.ImportMedia(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ImportMedia rpc: %w", err)
	}

	return &ImportMediaResult{
		ProjectItemID: resp.GetProjectItemId(),
		Name:          resp.GetName(),
	}, nil
}

// PlaceClip places a clip on the Premiere Pro timeline.
func (c *PremiereBridgeClient) PlaceClip(ctx context.Context, params PlaceClipParams) (*PlaceClipResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("PlaceClip",
		zap.String("source_path", params.SourcePath),
		zap.Uint32("track_index", params.Track.TrackIndex),
		zap.Float64("speed", params.Speed),
	)

	req := &premierepb.PlaceClipRequest{
		SourcePath: params.SourcePath,
		Track:      nativeTrackTargetToProto(params.Track),
		Position:   nativeTimecodeToProto(params.Position),
		Speed:      params.Speed,
	}
	if params.SourceRange != nil {
		req.SourceRange = nativeTimeRangeToProto(params.SourceRange)
	}

	resp, err := c.client.PlaceClip(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("PlaceClip rpc: %w", err)
	}

	return &PlaceClipResult{ClipID: resp.GetClipId()}, nil
}

// RemoveClip removes a clip from the timeline.
func (c *PremiereBridgeClient) RemoveClip(ctx context.Context, params RemoveClipParams) error {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("RemoveClip",
		zap.String("clip_id", params.ClipID),
		zap.String("sequence_id", params.SequenceID),
	)

	req := &premierepb.RemoveClipRequest{
		ClipId:     params.ClipID,
		SequenceId: params.SequenceID,
	}
	_, err := c.client.RemoveClip(ctx, req)
	if err != nil {
		return fmt.Errorf("RemoveClip rpc: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Effects & Transitions
// ---------------------------------------------------------------------------

// AddTransition adds a transition between clips on the timeline.
func (c *PremiereBridgeClient) AddTransition(ctx context.Context, params AddTransitionParams) (*AddTransitionResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("AddTransition",
		zap.String("sequence_id", params.SequenceID),
		zap.String("type", params.TransitionType),
		zap.Float64("duration", params.DurationSeconds),
	)

	req := &premierepb.AddTransitionRequest{
		SequenceId:      params.SequenceID,
		Track:           nativeTrackTargetToProto(params.Track),
		Position:        nativeTimecodeToProto(params.Position),
		TransitionType:  params.TransitionType,
		DurationSeconds: params.DurationSeconds,
	}
	resp, err := c.client.AddTransition(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AddTransition rpc: %w", err)
	}

	return &AddTransitionResult{TransitionID: resp.GetTransitionId()}, nil
}

// AddText adds a text overlay to the timeline.
func (c *PremiereBridgeClient) AddText(ctx context.Context, params AddTextParams) (*AddTextResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("AddText",
		zap.String("sequence_id", params.SequenceID),
		zap.String("text", params.Text),
		zap.Float64("duration", params.DurationSeconds),
	)

	req := &premierepb.AddTextRequest{
		SequenceId: params.SequenceID,
		Text:       params.Text,
		Style: &commonpb.TextStyle{
			FontFamily:         params.Style.FontFamily,
			FontSize:           params.Style.FontSize,
			ColorHex:           params.Style.ColorHex,
			Alignment:          params.Style.Alignment,
			BackgroundColorHex: params.Style.BackgroundColorHex,
			BackgroundOpacity:  params.Style.BackgroundOpacity,
			Position: &commonpb.Position{
				X: params.Style.Position.X,
				Y: params.Style.Position.Y,
			},
		},
		Track:           nativeTrackTargetToProto(params.Track),
		Position:        nativeTimecodeToProto(params.Position),
		DurationSeconds: params.DurationSeconds,
	}
	resp, err := c.client.AddText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AddText rpc: %w", err)
	}

	return &AddTextResult{ClipID: resp.GetClipId()}, nil
}

// ApplyEffect applies a visual/audio effect to a clip.
func (c *PremiereBridgeClient) ApplyEffect(ctx context.Context, params ApplyEffectParams) error {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ApplyEffect",
		zap.String("clip_id", params.ClipID),
		zap.String("sequence_id", params.SequenceID),
		zap.String("effect", params.Effect.Name),
	)

	req := &premierepb.ApplyEffectRequest{
		ClipId:     params.ClipID,
		SequenceId: params.SequenceID,
		Effect: &commonpb.EffectInfo{
			Name:       params.Effect.Name,
			Parameters: params.Effect.Parameters,
		},
	}
	_, err := c.client.ApplyEffect(ctx, req)
	if err != nil {
		return fmt.Errorf("ApplyEffect rpc: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Audio
// ---------------------------------------------------------------------------

// SetAudioLevel sets the audio level (in dB) for a clip.
func (c *PremiereBridgeClient) SetAudioLevel(ctx context.Context, params SetAudioLevelParams) error {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("SetAudioLevel",
		zap.String("clip_id", params.ClipID),
		zap.String("sequence_id", params.SequenceID),
		zap.Float64("level_db", params.LevelDB),
	)

	req := &premierepb.SetAudioLevelRequest{
		ClipId:     params.ClipID,
		SequenceId: params.SequenceID,
		LevelDb:    params.LevelDB,
	}
	_, err := c.client.SetAudioLevel(ctx, req)
	if err != nil {
		return fmt.Errorf("SetAudioLevel rpc: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

// ExportSequence starts an export of the given sequence.
func (c *PremiereBridgeClient) ExportSequence(ctx context.Context, params ExportSequenceParams) (*ExportSequenceResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ExportSequence",
		zap.String("sequence_id", params.SequenceID),
		zap.String("output_path", params.OutputPath),
		zap.Int("preset", int(params.Preset)),
	)

	req := &premierepb.ExportSequenceRequest{
		SequenceId: params.SequenceID,
		OutputPath: params.OutputPath,
		Preset:     premierepb.ExportPreset(params.Preset),
	}
	resp, err := c.client.ExportSequence(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ExportSequence rpc: %w", err)
	}

	return &ExportSequenceResult{
		JobID:      resp.GetJobId(),
		Status:     OperationStatus(resp.GetStatus()),
		OutputPath: resp.GetOutputPath(),
	}, nil
}

// ---------------------------------------------------------------------------
// Batch
// ---------------------------------------------------------------------------

// ExecuteEDL executes a full Edit Decision List: creates a sequence, imports media,
// places all clips, and adds transitions.
func (c *PremiereBridgeClient) ExecuteEDL(ctx context.Context, params ExecuteEDLParams) (*ExecuteEDLResult, error) {
	// EDL execution can be long-running; use a generous timeout.
	ctx, cancel := context.WithTimeout(ctx, c.callTimeout*3)
	defer cancel()

	c.logger.Debug("ExecuteEDL",
		zap.String("edl_name", params.EDL.Name),
		zap.Int("entries", len(params.EDL.Entries)),
		zap.Bool("auto_import", params.AutoImport),
		zap.Bool("auto_create_sequence", params.AutoCreateSequence),
	)

	req := &premierepb.ExecuteEDLRequest{
		Edl:                nativeEDLToProto(params.EDL),
		AutoImport:         params.AutoImport,
		AutoCreateSequence: params.AutoCreateSequence,
	}
	resp, err := c.client.ExecuteEDL(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ExecuteEDL rpc: %w", err)
	}

	return &ExecuteEDLResult{
		SequenceID:       resp.GetSequenceId(),
		Status:           OperationStatus(resp.GetStatus()),
		ClipsPlaced:      resp.GetClipsPlaced(),
		TransitionsAdded: resp.GetTransitionsAdded(),
		Errors:           resp.GetErrors(),
		Warnings:         resp.GetWarnings(),
	}, nil
}

// ---------------------------------------------------------------------------
// Audio & Track Management (generic command dispatcher)
// ---------------------------------------------------------------------------

// EvalAudioCommand dispatches a named ExtendScript audio/track command.
// This is a passthrough: the bridge evaluates the function and returns
// the JSON result as a map.
//
// NOTE: The EvalScript RPC has not yet been added to the proto definition.
// Once the premiere.proto is updated with an EvalScript RPC and the stubs
// are regenerated, replace this placeholder with the actual gRPC call.
func (c *PremiereBridgeClient) EvalAudioCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error) {
	c.logger.Debug("EvalAudioCommand", zap.String("command", command))

	// TODO: implement once EvalScript RPC is added to premiere.proto
	return nil, fmt.Errorf("EvalAudioCommand(%s): EvalScript RPC not yet defined in proto", command)
}

// EvalImmersiveCommand dispatches a named ExtendScript immersive-video command
// (VR/360, HDR, stereoscopic, frame-rate, aspect-ratio, timecode, render,
// captions). This is a passthrough: the bridge evaluates the function and
// returns the JSON result as a map.
//
// NOTE: The EvalScript RPC has not yet been added to the proto definition.
// Once the premiere.proto is updated with an EvalScript RPC and the stubs
// are regenerated, replace this placeholder with the actual gRPC call.
func (c *PremiereBridgeClient) EvalImmersiveCommand(ctx context.Context, command string, args map[string]any) (map[string]any, error) {
	c.logger.Debug("EvalImmersiveCommand", zap.String("command", command))

	// TODO: implement once EvalScript RPC is added to premiere.proto
	return nil, fmt.Errorf("EvalImmersiveCommand(%s): EvalScript RPC not yet defined in proto", command)
}

// ---------------------------------------------------------------------------
// Premiere-specific proto -> native converters
// ---------------------------------------------------------------------------

func protoTimelineTracksToNative(tracks []*premierepb.TimelineTrack) []TimelineTrack {
	result := make([]TimelineTrack, len(tracks))
	for i, t := range tracks {
		clips := make([]TimelineClip, len(t.GetClips()))
		for j, c := range t.GetClips() {
			clip := TimelineClip{
				ClipID:     c.GetClipId(),
				SourcePath: c.GetSourcePath(),
				Speed:      c.GetSpeed(),
			}
			if c.GetSourceRange() != nil {
				sr := protoTimeRangeToNative(c.GetSourceRange())
				clip.SourceRange = *sr
			}
			if c.GetTimelineRange() != nil {
				tr := protoTimeRangeToNative(c.GetTimelineRange())
				clip.TimelineRange = *tr
			}
			clips[j] = clip
		}
		result[i] = TimelineTrack{
			Index:    t.GetIndex(),
			Type:     TrackType(t.GetType()),
			Clips:    clips,
			IsMuted:  t.GetIsMuted(),
			IsLocked: t.GetIsLocked(),
		}
	}
	return result
}
