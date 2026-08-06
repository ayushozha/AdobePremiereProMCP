package mcp

import (
	"context"
	"strings"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type snapshotTimelineRecorder struct {
	Orchestrator
	sequenceIndexes []int
	savedVersions   []struct {
		sequenceIndex int
		versionName   string
		notes         string
	}
}

func (r *snapshotTimelineRecorder) SaveSequenceVersion(
	_ context.Context,
	sequenceIndex int,
	versionName, notes string,
) (*GenericResult, error) {
	r.savedVersions = append(r.savedVersions, struct {
		sequenceIndex int
		versionName   string
		notes         string
	}{sequenceIndex, versionName, notes})
	return &GenericResult{Status: "success"}, nil
}

func (r *snapshotTimelineRecorder) SnapshotTimeline(
	_ context.Context,
	sequenceIndex int,
) (*GenericResult, error) {
	r.sequenceIndexes = append(r.sequenceIndexes, sequenceIndex)
	return &GenericResult{Status: "success"}, nil
}

func TestSaveSequenceVersionDefaultsToActiveSequenceSentinel(t *testing.T) {
	recorder := &snapshotTimelineRecorder{}
	s := mcpserver.NewMCPServer(
		"version-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	registerVersioningTools(s, recorder, zap.NewNop())

	registered := s.GetTool("premiere_save_sequence_version")
	if registered == nil {
		t.Fatal("premiere_save_sequence_version is not registered")
	}
	property, ok := registered.Tool.InputSchema.Properties["sequence_index"].(map[string]any)
	if !ok {
		t.Fatalf("sequence_index schema = %T, want map[string]any", registered.Tool.InputSchema.Properties["sequence_index"])
	}
	if got := property["default"]; got != float64(-1) {
		t.Fatalf("sequence_index schema default = %v, want -1", got)
	}
	description, _ := property["description"].(string)
	if !strings.Contains(description, "active sequence") {
		t.Fatalf("sequence_index description %q does not explain the active-sequence default", description)
	}

	c := newInitializedClient(t, s)
	request := gomcp.CallToolRequest{}
	request.Params.Name = "premiere_save_sequence_version"
	request.Params.Arguments = map[string]any{
		"version_name": "Rough Cut",
		"notes":        "before dialogue cleanup",
	}
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call save version without sequence_index: %v", err)
	}
	if result.IsError {
		t.Fatalf("save version returned a tool error: %s", toolResultText(t, result))
	}
	if len(recorder.savedVersions) != 1 {
		t.Fatalf("SaveSequenceVersion call count = %d, want 1", len(recorder.savedVersions))
	}
	call := recorder.savedVersions[0]
	if call.sequenceIndex != -1 || call.versionName != "Rough Cut" || call.notes != "before dialogue cleanup" {
		t.Fatalf("SaveSequenceVersion arguments = %+v", call)
	}
}

func TestSnapshotTimelineDefaultsToActiveSequenceSentinel(t *testing.T) {
	recorder := &snapshotTimelineRecorder{}
	s := mcpserver.NewMCPServer(
		"snapshot-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	registerVersioningTools(s, recorder, zap.NewNop())

	registered := s.GetTool("premiere_snapshot_timeline")
	if registered == nil {
		t.Fatal("premiere_snapshot_timeline is not registered")
	}
	property, ok := registered.Tool.InputSchema.Properties["sequence_index"].(map[string]any)
	if !ok {
		t.Fatalf("sequence_index schema = %T, want map[string]any", registered.Tool.InputSchema.Properties["sequence_index"])
	}
	if got := property["default"]; got != float64(-1) {
		t.Fatalf("sequence_index schema default = %v, want -1", got)
	}
	description, _ := property["description"].(string)
	if !strings.Contains(description, "active sequence") {
		t.Fatalf("sequence_index description %q does not explain the active-sequence default", description)
	}

	c := newInitializedClient(t, s)
	request := gomcp.CallToolRequest{}
	request.Params.Name = "premiere_snapshot_timeline"
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call snapshot without sequence_index: %v", err)
	}
	if result.IsError {
		t.Fatalf("snapshot without sequence_index returned a tool error: %s", toolResultText(t, result))
	}

	request.Params.Arguments = map[string]any{"sequence_index": 3}
	result, err = c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call snapshot with explicit sequence_index: %v", err)
	}
	if result.IsError {
		t.Fatalf("snapshot with explicit sequence_index returned a tool error: %s", toolResultText(t, result))
	}

	if len(recorder.sequenceIndexes) != 2 {
		t.Fatalf("SnapshotTimeline call count = %d, want 2", len(recorder.sequenceIndexes))
	}
	if got := recorder.sequenceIndexes[0]; got != -1 {
		t.Fatalf("omitted sequence_index forwarded %d, want active sentinel -1", got)
	}
	if got := recorder.sequenceIndexes[1]; got != 3 {
		t.Fatalf("explicit sequence_index forwarded %d, want 3", got)
	}
}
