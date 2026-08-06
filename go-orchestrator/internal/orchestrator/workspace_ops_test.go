package orchestrator

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

type recordingPremiereClient struct {
	*mockPremiereClient
	command   string
	arguments string
}

func (r *recordingPremiereClient) EvalCommand(_ context.Context, command, arguments string) (string, error) {
	r.command = command
	r.arguments = arguments
	return `{"success":true}`, nil
}

func TestCreateProxyForwardsDeterministicOutputPath(t *testing.T) {
	premiere := &recordingPremiereClient{mockPremiereClient: &mockPremiereClient{}}
	engine := New(&mockMediaClient{}, &mockIntelClient{}, premiere, zap.NewNop())

	if _, err := engine.CreateProxy(context.Background(), 4, "/tmp/camera-a-proxy.mov", "/tmp/proxy-preset.epr"); err != nil {
		t.Fatalf("CreateProxy() error = %v", err)
	}
	if premiere.command != "createProxy" {
		t.Fatalf("EvalCommand command = %q, want createProxy", premiere.command)
	}

	var arguments map[string]any
	if err := json.Unmarshal([]byte(premiere.arguments), &arguments); err != nil {
		t.Fatalf("decode CreateProxy arguments: %v", err)
	}
	if arguments["projectItemIndex"] != float64(4) {
		t.Fatalf("projectItemIndex = %v, want 4", arguments["projectItemIndex"])
	}
	if arguments["outputPath"] != "/tmp/camera-a-proxy.mov" {
		t.Fatalf("outputPath = %v, want deterministic proxy path", arguments["outputPath"])
	}
	if arguments["presetPath"] != "/tmp/proxy-preset.epr" {
		t.Fatalf("presetPath = %v, want preset path", arguments["presetPath"])
	}
}
