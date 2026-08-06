package mcp

import (
	"context"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type captionExportRecorder struct {
	Orchestrator
	calls []struct {
		sequenceID string
		outputPath string
		format     string
	}
}

func (r *captionExportRecorder) ExportCaptions(
	_ context.Context,
	sequenceID, outputPath, format string,
) (*GenericResult, error) {
	r.calls = append(r.calls, struct {
		sequenceID string
		outputPath string
		format     string
	}{sequenceID, outputPath, format})
	return &GenericResult{Status: "ok"}, nil
}

func TestExportCaptionsRequiresAndForwardsSequenceIdentity(t *testing.T) {
	recorder := &captionExportRecorder{}
	s := mcpserver.NewMCPServer(
		"caption-export-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	registerGraphicsTools(s, recorder, zap.NewNop())

	tool := s.GetTool("premiere_export_captions")
	if tool == nil {
		t.Fatal("premiere_export_captions is not registered")
	}
	for _, required := range []string{"sequence_id", "output_path"} {
		if !containsString(tool.Tool.InputSchema.Required, required) {
			t.Fatalf("export captions required parameters = %v, want %s", tool.Tool.InputSchema.Required, required)
		}
	}

	c := newInitializedClient(t, s)
	request := gomcp.CallToolRequest{}
	request.Params.Name = "premiere_export_captions"
	request.Params.Arguments = map[string]any{
		"output_path": "/tmp/captions.srt",
		"format":      "SRT",
	}
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call without sequence identity: %v", err)
	}
	if !result.IsError || len(recorder.calls) != 0 {
		t.Fatal("caption export without sequence identity reached the orchestrator")
	}

	request.Params.Arguments = map[string]any{
		"sequence_id": "sequence-42",
		"output_path": "/tmp/captions.srt",
		"format":      "VTT",
	}
	result, err = c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call valid caption export: %v", err)
	}
	if result.IsError {
		t.Fatalf("valid caption export returned tool error: %s", toolResultText(t, result))
	}
	if len(recorder.calls) != 1 || recorder.calls[0].sequenceID != "sequence-42" ||
		recorder.calls[0].outputPath != "/tmp/captions.srt" || recorder.calls[0].format != "VTT" {
		t.Fatalf("ExportCaptions calls = %+v", recorder.calls)
	}
}
