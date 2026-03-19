// intelligence_client.go wraps the Python IntelligenceService gRPC RPCs using
// the real generated proto stubs.
package grpc

import (
	"context"
	"fmt"
	"time"

	commonpb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/common/v1"
	intelpb "github.com/anthropics/premierpro-mcp/gen/go/premierpro/intelligence/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// IntelligenceClient provides Go-native access to the Python intelligence service.
type IntelligenceClient struct {
	conn        *grpc.ClientConn
	client      intelpb.IntelligenceServiceClient
	callTimeout time.Duration
	logger      *zap.Logger
}

// newIntelligenceClient dials the intelligence service and returns a ready client.
func newIntelligenceClient(addr string, dialTimeout, callTimeout time.Duration, logger *zap.Logger) (*IntelligenceClient, error) {
	logger = logger.With(zap.String("client", "intelligence"), zap.String("addr", addr))
	logger.Info("connecting to intelligence service")

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("intelligence dial %s: %w", addr, err)
	}

	logger.Info("connected to intelligence service")
	return &IntelligenceClient{
		conn:        conn,
		client:      intelpb.NewIntelligenceServiceClient(conn),
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

	req := &intelpb.ParseScriptRequest{
		FormatHint: params.FormatHint,
	}
	if params.FilePath != "" {
		req.Source = &intelpb.ParseScriptRequest_FilePath{FilePath: params.FilePath}
	} else {
		req.Source = &intelpb.ParseScriptRequest_Text{Text: params.Text}
	}

	resp, err := c.client.ParseScript(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ParseScript rpc: %w", err)
	}

	segments := make([]ScriptSegment, len(resp.GetSegments()))
	for i, s := range resp.GetSegments() {
		segments[i] = protoScriptSegmentToNative(s)
	}

	var metadata ScriptMetadata
	if resp.GetMetadata() != nil {
		metadata = ScriptMetadata{
			Title:                         resp.GetMetadata().GetTitle(),
			Format:                        resp.GetMetadata().GetFormat(),
			EstimatedTotalDurationSeconds: resp.GetMetadata().GetEstimatedTotalDurationSeconds(),
			SegmentCount:                  resp.GetMetadata().GetSegmentCount(),
		}
	}

	return &ParseScriptResult{
		Segments: segments,
		Metadata: metadata,
	}, nil
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

	protoSegments := make([]*intelpb.ScriptSegment, len(params.Segments))
	for i, s := range params.Segments {
		protoSegments[i] = nativeScriptSegmentToProto(s)
	}

	protoAssets := make([]*commonpb.Asset, len(params.AvailableAssets))
	for i, a := range params.AvailableAssets {
		protoAssets[i] = nativeAssetToProto(a)
	}

	protoMatches := make([]*intelpb.AssetMatch, len(params.Matches))
	for i, m := range params.Matches {
		protoMatches[i] = nativeAssetMatchToProto(m)
	}

	req := &intelpb.GenerateEDLRequest{
		Segments:        protoSegments,
		AvailableAssets: protoAssets,
		Matches:         protoMatches,
		Settings: &intelpb.EDLSettings{
			Resolution:                nativeResolutionToProto(params.Settings.Resolution),
			FrameRate:                 params.Settings.FrameRate,
			DefaultTransition:         params.Settings.DefaultTransition,
			DefaultTransitionDuration: params.Settings.DefaultTransitionDuration,
			Pacing:                    intelpb.PacingPreset(params.Settings.Pacing),
		},
	}

	resp, err := c.client.GenerateEDL(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GenerateEDL rpc: %w", err)
	}

	return &GenerateEDLResult{
		EDL:      protoEDLToNative(resp.GetEdl()),
		Warnings: resp.GetWarnings(),
	}, nil
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

	protoSegments := make([]*intelpb.ScriptSegment, len(params.Segments))
	for i, s := range params.Segments {
		protoSegments[i] = nativeScriptSegmentToProto(s)
	}

	protoAssets := make([]*commonpb.Asset, len(params.AvailableAssets))
	for i, a := range params.AvailableAssets {
		protoAssets[i] = nativeAssetToProto(a)
	}

	req := &intelpb.MatchAssetsRequest{
		Segments:        protoSegments,
		AvailableAssets: protoAssets,
		Strategy:        intelpb.MatchStrategy(params.Strategy),
	}

	resp, err := c.client.MatchAssets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MatchAssets rpc: %w", err)
	}

	matches := make([]AssetMatch, len(resp.GetMatches()))
	for i, m := range resp.GetMatches() {
		matches[i] = protoAssetMatchToNative(m)
	}

	unmatched := make([]UnmatchedSegment, len(resp.GetUnmatched()))
	for i, u := range resp.GetUnmatched() {
		unmatched[i] = UnmatchedSegment{
			SegmentIndex: u.GetSegmentIndex(),
			Reason:       u.GetReason(),
			Suggestions:  u.GetSuggestions(),
		}
	}

	return &MatchAssetsResult{
		Matches:   matches,
		Unmatched: unmatched,
	}, nil
}

// AnalyzePacing analyses EDL pacing and suggests timing adjustments for a target mood.
func (c *IntelligenceClient) AnalyzePacing(ctx context.Context, params AnalyzePacingParams) (*AnalyzePacingResult, error) {
	ctx, cancel := c.callCtx(ctx)
	defer cancel()

	c.logger.Debug("AnalyzePacing",
		zap.String("target_mood", params.TargetMood),
		zap.Int("edl_entries", len(params.EDL.Entries)),
	)

	req := &intelpb.AnalyzePacingRequest{
		Edl:        nativeEDLToProto(params.EDL),
		TargetMood: params.TargetMood,
	}

	resp, err := c.client.AnalyzePacing(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AnalyzePacing rpc: %w", err)
	}

	adjustments := make([]PacingAdjustment, len(resp.GetAdjustments()))
	for i, a := range resp.GetAdjustments() {
		adjustments[i] = PacingAdjustment{
			EDLEntryIndex:     a.GetEdlEntryIndex(),
			CurrentDuration:   a.GetCurrentDuration(),
			SuggestedDuration: a.GetSuggestedDuration(),
			Reason:            a.GetReason(),
		}
	}

	return &AnalyzePacingResult{
		Adjustments:              adjustments,
		CurrentAvgClipDuration:   resp.GetCurrentAvgClipDuration(),
		SuggestedAvgClipDuration: resp.GetSuggestedAvgClipDuration(),
	}, nil
}

// ---------------------------------------------------------------------------
// Intelligence-specific proto <-> native converters
// ---------------------------------------------------------------------------

func nativeScriptSegmentToProto(s ScriptSegment) *intelpb.ScriptSegment {
	return &intelpb.ScriptSegment{
		Index:                    s.Index,
		Type:                     intelpb.SegmentType(s.Type),
		Content:                  s.Content,
		Speaker:                  s.Speaker,
		SceneDescription:         s.SceneDescription,
		VisualDirection:          s.VisualDirection,
		AudioDirection:           s.AudioDirection,
		EstimatedDurationSeconds: s.EstimatedDurationSeconds,
		AssetHints:               s.AssetHints,
	}
}

func protoScriptSegmentToNative(s *intelpb.ScriptSegment) ScriptSegment {
	if s == nil {
		return ScriptSegment{}
	}
	return ScriptSegment{
		Index:                    s.GetIndex(),
		Type:                     SegmentType(s.GetType()),
		Content:                  s.GetContent(),
		Speaker:                  s.GetSpeaker(),
		SceneDescription:         s.GetSceneDescription(),
		VisualDirection:          s.GetVisualDirection(),
		AudioDirection:           s.GetAudioDirection(),
		EstimatedDurationSeconds: s.GetEstimatedDurationSeconds(),
		AssetHints:               s.GetAssetHints(),
	}
}

func nativeAssetMatchToProto(m AssetMatch) *intelpb.AssetMatch {
	pm := &intelpb.AssetMatch{
		SegmentIndex: m.SegmentIndex,
		AssetId:      m.AssetID,
		Confidence:   m.Confidence,
		Reasoning:    m.Reasoning,
	}
	if m.SuggestedRange != nil {
		pm.SuggestedRange = nativeTimeRangeToProto(m.SuggestedRange)
	}
	return pm
}

func protoAssetMatchToNative(m *intelpb.AssetMatch) AssetMatch {
	if m == nil {
		return AssetMatch{}
	}
	am := AssetMatch{
		SegmentIndex: m.GetSegmentIndex(),
		AssetID:      m.GetAssetId(),
		Confidence:   m.GetConfidence(),
		Reasoning:    m.GetReasoning(),
	}
	if m.GetSuggestedRange() != nil {
		am.SuggestedRange = protoTimeRangeToNative(m.GetSuggestedRange())
	}
	return am
}
