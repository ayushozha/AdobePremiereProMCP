package mcp

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ayushozha/AdobePremiereProMCP/go-orchestrator/internal/observability"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

func TestServerMetricsIncludeArgumentValidationAndRecoveredPanic(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "standard")
	metrics := observability.NewMetrics(zap.NewNop())
	s := NewMCPServer(nil, "test", zap.NewNop(), server.WithToolHandlerMiddleware(metrics.Middleware))
	metrics.SetTools([]string{"premiere_scan_assets", "premiere_ping"})
	c := newInitializedClient(t, s)
	// Missing the required directory must count even though validation stops the handler.
	request := gomcp.CallToolRequest{}
	request.Params.Name = "premiere_scan_assets"
	result, err := c.CallTool(context.Background(), request)
	if err != nil || !result.IsError {
		t.Fatalf("validation result %+v, err %v", result, err)
	}
	// The nil orchestrator panics in this real tool handler and is recovered by mcp-go.
	request.Params.Name = "premiere_ping"
	_, err = c.CallTool(context.Background(), request)
	if err == nil {
		t.Fatal("expected recovered handler error")
	}
	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	for _, name := range []string{"premiere_scan_assets", "premiere_ping"} {
		expected := `premiere_mcp_tool_requests_total{outcome="error",tool="` + name + `"} 1`
		if !strings.Contains(response.Body.String(), expected) {
			t.Fatalf("missing %s in %s", expected, response.Body.String())
		}
	}
}
