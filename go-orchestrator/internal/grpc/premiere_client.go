// premiere_client.go wraps the TypeScript PremiereBridgeService gRPC RPCs.
//
// PROTO STUB NOTE: Once proto stubs are generated, replace the placeholder
// grpcStub field with the real generated client:
//
//   import prempb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/premiere/v1"
//
// and swap the method bodies to convert between Go-native types and proto types.
package grpc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// PremiereBridgeClient provides Go-native access to the TypeScript Premiere bridge.
type PremiereBridgeClient struct {
	conn        *grpc.ClientConn
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// resp, err := c.stub.Ping(ctx, &prempb.PingRequest{})
	// if err != nil {
	//     return nil, fmt.Errorf("Ping rpc: %w", err)
	// }
	// return &PingResult{
	//     PremiereRunning: resp.PremiereRunning,
	//     PremiereVersion: resp.PremiereVersion,
	//     ProjectOpen:     resp.ProjectOpen,
	//     BridgeMode:      resp.BridgeMode,
	// }, nil

	_ = ctx
	return nil, fmt.Errorf("Ping: proto stubs not yet generated")
}

// ---------------------------------------------------------------------------
// Project
// ---------------------------------------------------------------------------

// GetProjectState retrieves the current Premiere Pro project state.
func (c *PremiereBridgeClient) GetProjectState(ctx context.Context) (*GetProjectStateResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GetProjectState")

	// TODO: Replace with real proto call once stubs are generated.
	//
	// resp, err := c.stub.GetProjectState(ctx, &prempb.GetProjectStateRequest{})
	// if err != nil {
	//     return nil, fmt.Errorf("GetProjectState rpc: %w", err)
	// }
	// return convertProjectStateResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("GetProjectState: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.CreateSequenceRequest{
	//     Name:        params.Name,
	//     Resolution:  toProtoResolution(params.Resolution),
	//     FrameRate:   params.FrameRate,
	//     VideoTracks: params.VideoTracks,
	//     AudioTracks: params.AudioTracks,
	// }
	// resp, err := c.stub.CreateSequence(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("CreateSequence rpc: %w", err)
	// }
	// return &CreateSequenceResult{SequenceID: resp.SequenceId, Name: resp.Name}, nil

	_ = ctx
	return nil, fmt.Errorf("CreateSequence: proto stubs not yet generated")
}

// GetTimelineState retrieves the current timeline state of a sequence.
func (c *PremiereBridgeClient) GetTimelineState(ctx context.Context, params GetTimelineStateParams) (*GetTimelineStateResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GetTimelineState", zap.String("sequence_id", params.SequenceID))

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.GetTimelineStateRequest{SequenceId: params.SequenceID}
	// resp, err := c.stub.GetTimelineState(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("GetTimelineState rpc: %w", err)
	// }
	// return convertTimelineStateResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("GetTimelineState: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.ImportMediaRequest{
	//     FilePath:  params.FilePath,
	//     TargetBin: params.TargetBin,
	// }
	// resp, err := c.stub.ImportMedia(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("ImportMedia rpc: %w", err)
	// }
	// return &ImportMediaResult{ProjectItemID: resp.ProjectItemId, Name: resp.Name}, nil

	_ = ctx
	return nil, fmt.Errorf("ImportMedia: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.PlaceClipRequest{
	//     SourcePath:  params.SourcePath,
	//     Track:       toProtoTrackTarget(params.Track),
	//     Position:    toProtoTimecode(params.Position),
	//     SourceRange: toProtoTimeRange(params.SourceRange),
	//     Speed:       params.Speed,
	// }
	// resp, err := c.stub.PlaceClip(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("PlaceClip rpc: %w", err)
	// }
	// return &PlaceClipResult{ClipID: resp.ClipId}, nil

	_ = ctx
	return nil, fmt.Errorf("PlaceClip: proto stubs not yet generated")
}

// RemoveClip removes a clip from the timeline.
func (c *PremiereBridgeClient) RemoveClip(ctx context.Context, params RemoveClipParams) error {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("RemoveClip",
		zap.String("clip_id", params.ClipID),
		zap.String("sequence_id", params.SequenceID),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.RemoveClipRequest{
	//     ClipId:     params.ClipID,
	//     SequenceId: params.SequenceID,
	// }
	// _, err := c.stub.RemoveClip(ctx, req)
	// if err != nil {
	//     return fmt.Errorf("RemoveClip rpc: %w", err)
	// }
	// return nil

	_ = ctx
	return fmt.Errorf("RemoveClip: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.AddTransitionRequest{
	//     SequenceId:     params.SequenceID,
	//     Track:          toProtoTrackTarget(params.Track),
	//     Position:       toProtoTimecode(params.Position),
	//     TransitionType: params.TransitionType,
	//     DurationSeconds: params.DurationSeconds,
	// }
	// resp, err := c.stub.AddTransition(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("AddTransition rpc: %w", err)
	// }
	// return &AddTransitionResult{TransitionID: resp.TransitionId}, nil

	_ = ctx
	return nil, fmt.Errorf("AddTransition: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.AddTextRequest{
	//     SequenceId:      params.SequenceID,
	//     Text:            params.Text,
	//     Style:           toProtoTextStyle(params.Style),
	//     Track:           toProtoTrackTarget(params.Track),
	//     Position:        toProtoTimecode(params.Position),
	//     DurationSeconds: params.DurationSeconds,
	// }
	// resp, err := c.stub.AddText(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("AddText rpc: %w", err)
	// }
	// return &AddTextResult{ClipID: resp.ClipId}, nil

	_ = ctx
	return nil, fmt.Errorf("AddText: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.ApplyEffectRequest{
	//     ClipId:     params.ClipID,
	//     SequenceId: params.SequenceID,
	//     Effect:     toProtoEffectInfo(params.Effect),
	// }
	// _, err := c.stub.ApplyEffect(ctx, req)
	// if err != nil {
	//     return fmt.Errorf("ApplyEffect rpc: %w", err)
	// }
	// return nil

	_ = ctx
	return fmt.Errorf("ApplyEffect: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.SetAudioLevelRequest{
	//     ClipId:     params.ClipID,
	//     SequenceId: params.SequenceID,
	//     LevelDb:    params.LevelDB,
	// }
	// _, err := c.stub.SetAudioLevel(ctx, req)
	// if err != nil {
	//     return fmt.Errorf("SetAudioLevel rpc: %w", err)
	// }
	// return nil

	_ = ctx
	return fmt.Errorf("SetAudioLevel: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.ExportSequenceRequest{
	//     SequenceId: params.SequenceID,
	//     OutputPath: params.OutputPath,
	//     Preset:     prempb.ExportPreset(params.Preset),
	// }
	// resp, err := c.stub.ExportSequence(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("ExportSequence rpc: %w", err)
	// }
	// return &ExportSequenceResult{
	//     JobID:      resp.JobId,
	//     Status:     OperationStatus(resp.Status),
	//     OutputPath: resp.OutputPath,
	// }, nil

	_ = ctx
	return nil, fmt.Errorf("ExportSequence: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &prempb.ExecuteEDLRequest{
	//     Edl:                toProtoEDL(params.EDL),
	//     AutoImport:         params.AutoImport,
	//     AutoCreateSequence: params.AutoCreateSequence,
	// }
	// resp, err := c.stub.ExecuteEDL(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("ExecuteEDL rpc: %w", err)
	// }
	// return &ExecuteEDLResult{
	//     SequenceID:       resp.SequenceId,
	//     Status:           OperationStatus(resp.Status),
	//     ClipsPlaced:      resp.ClipsPlaced,
	//     TransitionsAdded: resp.TransitionsAdded,
	//     Errors:           resp.Errors,
	//     Warnings:         resp.Warnings,
	// }, nil

	_ = ctx
	return nil, fmt.Errorf("ExecuteEDL: proto stubs not yet generated")
}
