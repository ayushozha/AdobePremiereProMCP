// media_client.go wraps the Rust MediaEngineService gRPC RPCs.
//
// PROTO STUB NOTE: Once proto stubs are generated, replace the placeholder
// grpcStub field with the real generated client:
//
//   import mediapb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/media/v1"
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

// MediaEngineClient provides Go-native access to the Rust media engine service.
type MediaEngineClient struct {
	conn        *grpc.ClientConn
	callTimeout time.Duration
	logger      *zap.Logger
}

// newMediaEngineClient dials the media engine service and returns a ready client.
func newMediaEngineClient(addr string, dialTimeout, callTimeout time.Duration, logger *zap.Logger) (*MediaEngineClient, error) {
	logger = logger.With(zap.String("client", "media_engine"), zap.String("addr", addr))
	logger.Info("connecting to media engine service")

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("media engine dial %s: %w", addr, err)
	}

	logger.Info("connected to media engine service")
	return &MediaEngineClient{
		conn:        conn,
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &mediapb.ScanAssetsRequest{
	//     Directory:  params.Directory,
	//     Recursive:  params.Recursive,
	//     Extensions: params.Extensions,
	// }
	// resp, err := c.stub.ScanAssets(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("ScanAssets rpc: %w", err)
	// }
	// return convertScanAssetsResponse(resp), nil

	_ = ctx // suppress unused warning until proto stubs arrive
	return nil, fmt.Errorf("ScanAssets: proto stubs not yet generated")
}

// ProbeMedia inspects a single media file and returns detailed metadata.
func (c *MediaEngineClient) ProbeMedia(ctx context.Context, params ProbeMediaParams) (*ProbeMediaResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ProbeMedia", zap.String("file_path", params.FilePath))

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &mediapb.ProbeMediaRequest{FilePath: params.FilePath}
	// resp, err := c.stub.ProbeMedia(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("ProbeMedia rpc: %w", err)
	// }
	// return convertProbeMediaResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("ProbeMedia: proto stubs not yet generated")
}

// GenerateThumbnail extracts a thumbnail image from a video at the given timestamp.
func (c *MediaEngineClient) GenerateThumbnail(ctx context.Context, params GenerateThumbnailParams) (*GenerateThumbnailResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GenerateThumbnail",
		zap.String("file_path", params.FilePath),
		zap.String("format", params.OutputFormat),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &mediapb.GenerateThumbnailRequest{
	//     FilePath:     params.FilePath,
	//     Timestamp:    toProtoTimecode(params.Timestamp),
	//     OutputSize:   toProtoResolution(params.OutputSize),
	//     OutputFormat: params.OutputFormat,
	// }
	// resp, err := c.stub.GenerateThumbnail(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("GenerateThumbnail rpc: %w", err)
	// }
	// return convertGenerateThumbnailResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("GenerateThumbnail: proto stubs not yet generated")
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

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &mediapb.AnalyzeWaveformRequest{
	//     FilePath:                  params.FilePath,
	//     AudioTrack:               params.AudioTrack,
	//     SilenceThresholdDb:       params.SilenceThresholdDB,
	//     MinSilenceDurationSeconds: params.MinSilenceDurationSeconds,
	// }
	// resp, err := c.stub.AnalyzeWaveform(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("AnalyzeWaveform rpc: %w", err)
	// }
	// return convertAnalyzeWaveformResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("AnalyzeWaveform: proto stubs not yet generated")
}

// DetectScenes detects scene change boundaries in a video file.
func (c *MediaEngineClient) DetectScenes(ctx context.Context, params DetectScenesParams) (*DetectScenesResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("DetectScenes",
		zap.String("file_path", params.FilePath),
		zap.Float64("threshold", params.Threshold),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &mediapb.DetectScenesRequest{
	//     FilePath:  params.FilePath,
	//     Threshold: params.Threshold,
	// }
	// resp, err := c.stub.DetectScenes(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("DetectScenes rpc: %w", err)
	// }
	// return convertDetectScenesResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("DetectScenes: proto stubs not yet generated")
}
