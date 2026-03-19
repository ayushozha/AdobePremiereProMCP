// intelligence_client.go wraps the Python IntelligenceService gRPC RPCs.
//
// PROTO STUB NOTE: Once proto stubs are generated, replace the placeholder
// grpcStub field with the real generated client:
//
//   import intelpb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/intelligence/v1"
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

// IntelligenceClient provides Go-native access to the Python intelligence service.
type IntelligenceClient struct {
	conn        *grpc.ClientConn
	callTimeout time.Duration
	logger      *zap.Logger
}

// newIntelligenceClient dials the intelligence service and returns a ready client.
func newIntelligenceClient(addr string, dialTimeout, callTimeout time.Duration, logger *zap.Logger) (*IntelligenceClient, error) {
	logger = logger.With(zap.String("client", "intelligence"), zap.String("addr", addr))
	logger.Info("connecting to intelligence service")

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
		return nil, fmt.Errorf("intelligence dial %s: %w", addr, err)
	}

	logger.Info("connected to intelligence service")
	return &IntelligenceClient{
		conn:        conn,
		callTimeout: callTimeout,
		logger:      logger,
	}, nil
}

// close shuts down the underlying gRPC connection.
func (c *IntelligenceClient) close() error {
	c.logger.Info("closing intelligence connection")
	return c.conn.Close()
}

// callCtx derives a context with the configured call timeout.
func (c *IntelligenceClient) callCtx(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, c.callTimeout)
}

// ---------------------------------------------------------------------------
// RPC wrappers
// ---------------------------------------------------------------------------

// ParseScript parses a script (text or file) into structured segments.
func (c *IntelligenceClient) ParseScript(ctx context.Context, params ParseScriptParams) (*ParseScriptResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("ParseScript",
		zap.String("format_hint", params.FormatHint),
		zap.Bool("from_file", params.FilePath != ""),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &intelpb.ParseScriptRequest{
	//     FormatHint: params.FormatHint,
	// }
	// if params.FilePath != "" {
	//     req.Source = &intelpb.ParseScriptRequest_FilePath{FilePath: params.FilePath}
	// } else {
	//     req.Source = &intelpb.ParseScriptRequest_Text{Text: params.Text}
	// }
	// resp, err := c.stub.ParseScript(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("ParseScript rpc: %w", err)
	// }
	// return convertParseScriptResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("ParseScript: proto stubs not yet generated")
}

// GenerateEDL creates an Edit Decision List from parsed script segments,
// available assets, and asset matches.
func (c *IntelligenceClient) GenerateEDL(ctx context.Context, params GenerateEDLParams) (*GenerateEDLResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("GenerateEDL",
		zap.Int("segments", len(params.Segments)),
		zap.Int("assets", len(params.AvailableAssets)),
		zap.Int("matches", len(params.Matches)),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &intelpb.GenerateEDLRequest{
	//     Segments:        toProtoSegments(params.Segments),
	//     AvailableAssets: toProtoAssets(params.AvailableAssets),
	//     Matches:         toProtoMatches(params.Matches),
	//     Settings:        toProtoEDLSettings(params.Settings),
	// }
	// resp, err := c.stub.GenerateEDL(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("GenerateEDL rpc: %w", err)
	// }
	// return convertGenerateEDLResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("GenerateEDL: proto stubs not yet generated")
}

// MatchAssets matches script segments to the best available media assets.
func (c *IntelligenceClient) MatchAssets(ctx context.Context, params MatchAssetsParams) (*MatchAssetsResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("MatchAssets",
		zap.Int("segments", len(params.Segments)),
		zap.Int("assets", len(params.AvailableAssets)),
		zap.Int("strategy", int(params.Strategy)),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &intelpb.MatchAssetsRequest{
	//     Segments:        toProtoSegments(params.Segments),
	//     AvailableAssets: toProtoAssets(params.AvailableAssets),
	//     Strategy:        intelpb.MatchStrategy(params.Strategy),
	// }
	// resp, err := c.stub.MatchAssets(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("MatchAssets rpc: %w", err)
	// }
	// return convertMatchAssetsResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("MatchAssets: proto stubs not yet generated")
}

// AnalyzePacing analyses EDL pacing and suggests timing adjustments for a target mood.
func (c *IntelligenceClient) AnalyzePacing(ctx context.Context, params AnalyzePacingParams) (*AnalyzePacingResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("AnalyzePacing",
		zap.String("target_mood", params.TargetMood),
		zap.Int("edl_entries", len(params.EDL.Entries)),
	)

	// TODO: Replace with real proto call once stubs are generated.
	//
	// req := &intelpb.AnalyzePacingRequest{
	//     Edl:        toProtoEDL(params.EDL),
	//     TargetMood: params.TargetMood,
	// }
	// resp, err := c.stub.AnalyzePacing(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("AnalyzePacing rpc: %w", err)
	// }
	// return convertAnalyzePacingResponse(resp), nil

	_ = ctx
	return nil, fmt.Errorf("AnalyzePacing: proto stubs not yet generated")
}
