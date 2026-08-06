package mcp

import (
	"context"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type socialReframeRecorder struct {
	Orchestrator
	autoReframeCalls []struct {
		numerator, denominator int
		sourceSequenceID       string
		motionPreset, newName  string
		useNestedSequences     bool
	}
	positionCalls []struct {
		trackIndex, clipIndex int
		x, y                  float64
	}
	scaleCalls []struct {
		trackIndex, clipIndex int
		scale                 float64
	}
}

func (r *socialReframeRecorder) AutoReframeSequence(
	_ context.Context,
	sourceSequenceID string,
	numerator, denominator int,
	motionPreset, newName string,
	useNestedSequences bool,
) (*GenericResult, error) {
	r.autoReframeCalls = append(r.autoReframeCalls, struct {
		numerator, denominator int
		sourceSequenceID       string
		motionPreset, newName  string
		useNestedSequences     bool
	}{numerator, denominator, sourceSequenceID, motionPreset, newName, useNestedSequences})
	return &GenericResult{Status: "ok"}, nil
}

func (r *socialReframeRecorder) SetPosition(
	_ context.Context,
	trackIndex, clipIndex int,
	x, y float64,
) (*GenericResult, error) {
	r.positionCalls = append(r.positionCalls, struct {
		trackIndex, clipIndex int
		x, y                  float64
	}{trackIndex, clipIndex, x, y})
	return &GenericResult{Status: "ok"}, nil
}

func (r *socialReframeRecorder) SetScale(
	_ context.Context,
	trackIndex, clipIndex int,
	scale float64,
) (*GenericResult, error) {
	r.scaleCalls = append(r.scaleCalls, struct {
		trackIndex, clipIndex int
		scale                 float64
	}{trackIndex, clipIndex, scale})
	return &GenericResult{Status: "ok"}, nil
}

func newSocialToolTestServer(t *testing.T, recorder *socialReframeRecorder) *mcpserver.MCPServer {
	t.Helper()
	s := mcpserver.NewMCPServer(
		"social-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	registerSequenceTools(s, recorder, zap.NewNop())
	registerEffectsTools(s, recorder, zap.NewNop())
	return s
}

func TestAutoReframeUsesDocumentedContract(t *testing.T) {
	recorder := &socialReframeRecorder{}
	c := newInitializedClient(t, newSocialToolTestServer(t, recorder))

	request := gomcp.CallToolRequest{}
	request.Params.Name = "premiere_auto_reframe"
	request.Params.Arguments = map[string]any{
		"source_sequence_id":   "source-1",
		"numerator":            9,
		"denominator":          16,
		"motion_preset":        "faster",
		"new_name":             "Vertical Cut",
		"use_nested_sequences": true,
	}
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call auto reframe: %v", err)
	}
	if result.IsError {
		t.Fatalf("auto reframe returned tool error: %s", toolResultText(t, result))
	}
	if len(recorder.autoReframeCalls) != 1 {
		t.Fatalf("AutoReframeSequence call count = %d, want 1", len(recorder.autoReframeCalls))
	}
	call := recorder.autoReframeCalls[0]
	if call.sourceSequenceID != "source-1" || call.numerator != 9 || call.denominator != 16 || call.motionPreset != "faster" || call.newName != "Vertical Cut" || !call.useNestedSequences {
		t.Fatalf("AutoReframeSequence arguments = %+v", call)
	}

	request.Params.Arguments = map[string]any{
		"numerator":   9,
		"denominator": 16,
	}
	result, err = c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call auto reframe without source identity: %v", err)
	}
	if !result.IsError {
		t.Fatal("auto reframe without source_sequence_id was accepted")
	}

	request.Params.Arguments = map[string]any{
		"source_sequence_id": "source-1",
		"numerator":          9,
		"denominator":        16,
		"motion_preset":      "fast",
	}
	result, err = c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call invalid auto reframe preset: %v", err)
	}
	if !result.IsError {
		t.Fatal("legacy motion preset 'fast' was accepted")
	}

	request.Params.Arguments = map[string]any{
		"source_sequence_id": "source-1",
		"numerator":          9.5,
		"denominator":        16,
	}
	result, err = c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call fractional auto reframe ratio: %v", err)
	}
	if !result.IsError {
		t.Fatal("fractional aspect-ratio numerator was accepted")
	}
	if len(recorder.autoReframeCalls) != 1 {
		t.Fatalf("invalid requests reached orchestrator; call count = %d", len(recorder.autoReframeCalls))
	}
}

func TestVerifiedMotionToolsValidateAndForwardNormalizedValues(t *testing.T) {
	recorder := &socialReframeRecorder{}
	c := newInitializedClient(t, newSocialToolTestServer(t, recorder))

	position := gomcp.CallToolRequest{}
	position.Params.Name = "premiere_set_position"
	position.Params.Arguments = map[string]any{
		"track_index": 1,
		"clip_index":  2,
		"x":           0.25,
		"y":           0.75,
	}
	result, err := c.CallTool(context.Background(), position)
	if err != nil {
		t.Fatalf("call set position: %v", err)
	}
	if result.IsError {
		t.Fatalf("set position returned tool error: %s", toolResultText(t, result))
	}
	if len(recorder.positionCalls) != 1 || recorder.positionCalls[0].x != 0.25 || recorder.positionCalls[0].y != 0.75 {
		t.Fatalf("SetPosition calls = %+v", recorder.positionCalls)
	}

	position.Params.Arguments = map[string]any{
		"track_index": 0.5,
		"clip_index":  2,
		"x":           0.5,
		"y":           0.5,
	}
	result, err = c.CallTool(context.Background(), position)
	if err != nil {
		t.Fatalf("call fractional set position index: %v", err)
	}
	if !result.IsError || len(recorder.positionCalls) != 1 {
		t.Fatal("fractional track index reached SetPosition")
	}

	scale := gomcp.CallToolRequest{}
	scale.Params.Name = "premiere_set_scale"
	scale.Params.Arguments = map[string]any{
		"track_index": 0,
		"clip_index":  3,
		"scale":       125,
	}
	result, err = c.CallTool(context.Background(), scale)
	if err != nil {
		t.Fatalf("call set scale: %v", err)
	}
	if result.IsError {
		t.Fatalf("set scale returned tool error: %s", toolResultText(t, result))
	}
	if len(recorder.scaleCalls) != 1 || recorder.scaleCalls[0].scale != 125 {
		t.Fatalf("SetScale calls = %+v", recorder.scaleCalls)
	}

	scale.Params.Arguments = map[string]any{
		"track_index": 0,
		"clip_index":  3,
		"scale":       0,
	}
	result, err = c.CallTool(context.Background(), scale)
	if err != nil {
		t.Fatalf("call invalid set scale: %v", err)
	}
	if !result.IsError || len(recorder.scaleCalls) != 1 {
		t.Fatal("zero scale reached SetScale")
	}
}
