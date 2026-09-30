package grpc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	premierepb "github.com/ayushozha/AdobePremiereProMCP/gen/go/premierpro/premiere/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type evalResultClient struct {
	premierepb.PremiereBridgeServiceClient
	response *premierepb.EvalCommandResponse
}

func (c *evalResultClient) EvalCommand(context.Context, *premierepb.EvalCommandRequest, ...grpc.CallOption) (*premierepb.EvalCommandResponse, error) {
	return c.response, nil
}

func TestEvalCommandNormalizesOnlyCanonicalHostEnvelopes(t *testing.T) {
	tests := []struct {
		name, raw, want, errorText string
	}{
		{"old panel and standalone payload", `{"success":true,"data":{"count":1,"sequences":[{"name":"Master"}]}}`, `{"count":1,"sequences":[{"name":"Master"}]}`, ""},
		{"native failure", `{"success":false,"error":"No project is open"}`, "", "No project is open"},
		{"empty native failure", `{"success":false,"error":""}`, "", "without an error message"},
		{"flat success flag", `{"success":true,"imported":2}`, `{"success":true,"imported":2}`, ""},
		{"flat payload with data field", `{"success":true,"data":"business value","imported":2}`, `{"success":true,"data":"business value","imported":2}`, ""},
		{"missing envelope data", `{"success":true}`, `{"success":true}`, ""},
		{"null success flag", `{"success":null,"data":{"count":1}}`, `{"success":null,"data":{"count":1}}`, ""},
		{"flat payload", `{"count":1}`, `{"count":1}`, ""},
		{"scalar payload", `42`, `42`, ""},
		{"null host data", `{"success":true,"data":null}`, `null`, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &PremiereBridgeClient{client: &evalResultClient{response: &premierepb.EvalCommandResponse{ResultJson: tc.raw}}, callTimeout: time.Second, logger: zap.NewNop()}
			got, err := client.EvalCommand(context.Background(), "getSequenceList", "{}")
			if tc.errorText != "" {
				if err == nil || !strings.Contains(err.Error(), tc.errorText) || !strings.Contains(err.Error(), "getSequenceList") {
					t.Fatalf("error = %v, want command and %q", err, tc.errorText)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("EvalCommand = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestGenericHostReadbackRetainsFieldsThroughLegacyEnvelope(t *testing.T) {
	client := &PremiereBridgeClient{client: &evalResultClient{response: &premierepb.EvalCommandResponse{ResultJson: `{"success":true,"data":{"count":1,"tracks":[{"name":"Dialogue"}]}}`}}, callTimeout: time.Second, logger: zap.NewNop()}
	result, err := client.EvalAudioCommand(context.Background(), "getAudioTracks", nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil || result["count"] != float64(1) || !strings.Contains(string(encoded), "Dialogue") {
		t.Fatalf("lost native readback fields: %s, %v", encoded, err)
	}
}
