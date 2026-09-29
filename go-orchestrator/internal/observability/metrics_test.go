package observability

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMetricsClassifyCompletionWithoutRecordingPayloads(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	metrics := NewMetrics(zap.New(core))
	metrics.SetTools([]string{"premiere_test"})
	tests := []struct {
		outcome string
		handler server.ToolHandlerFunc
	}{
		{"success", func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			return gomcp.NewToolResultText("secret-result"), nil
		}},
		{"error", func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			return gomcp.NewToolResultError("secret-result"), nil
		}},
		{"error", func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			return nil, errors.New("secret-error")
		}},
		{"canceled", func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			return nil, fmt.Errorf("wrapped: %w", context.Canceled)
		}},
		{"timeout", func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			return nil, status.Error(codes.DeadlineExceeded, "secret-error")
		}},
		{"error", func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) { return nil, nil }},
	}
	for _, tc := range tests {
		request := gomcp.CallToolRequest{}
		request.Params.Name = "premiere_test"
		request.Params.Arguments = map[string]any{"prompt": "secret-prompt", "token": "secret-token"}
		_, _ = metrics.Middleware(tc.handler)(context.Background(), request)
		entry := logs.All()[logs.Len()-1]
		if got := entry.ContextMap()["outcome"]; got != tc.outcome {
			t.Fatalf("outcome %v, want %s", got, tc.outcome)
		}
	}
	text := scrape(t, metrics)
	for _, sample := range []string{
		`premiere_mcp_tool_requests_total{outcome="success",tool="premiere_test"} 1`,
		`premiere_mcp_tool_requests_total{outcome="error",tool="premiere_test"} 3`,
		`premiere_mcp_tool_requests_total{outcome="canceled",tool="premiere_test"} 1`,
		`premiere_mcp_tool_requests_total{outcome="timeout",tool="premiere_test"} 1`,
		`premiere_mcp_tool_duration_seconds_count{tool="premiere_test"} 6`,
	} {
		if !strings.Contains(text, sample) {
			t.Errorf("missing metric %s in %s", sample, text)
		}
	}
	if strings.Contains(text+fmt.Sprint(logs.All()), "secret-") {
		t.Fatal("observability exposed request/result/error content")
	}
}

func TestMetricsConcurrentCallsAndScrapesBoundUnknownToolNames(t *testing.T) {
	metrics := NewMetrics(zap.NewNop())
	metrics.SetTools([]string{"premiere_test"})
	handler := metrics.Middleware(func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		return gomcp.NewToolResultText("ok"), nil
	})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request := gomcp.CallToolRequest{}
			request.Params.Name = fmt.Sprintf("untrusted-%d", i)
			_, _ = handler(context.Background(), request)
			response := httptest.NewRecorder()
			metrics.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
		}(i)
	}
	wg.Wait()
	text := scrape(t, metrics)
	if strings.Contains(text, "untrusted-") {
		t.Fatal("request tool names produced unbounded labels")
	}
	if !strings.Contains(text, `premiere_mcp_tool_requests_total{outcome="success",tool="unknown"} 100`) {
		t.Fatalf("lost concurrent requests: %s", text)
	}
	if !strings.Contains(text, `premiere_mcp_tool_duration_seconds_count{tool="unknown"} 100`) {
		t.Fatalf("lost latency observations: %s", text)
	}
}

func TestCanceledContextWithToolErrorCountsAsCanceled(t *testing.T) {
	metrics := NewMetrics(zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	handler := metrics.Middleware(func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		cancel()
		return gomcp.NewToolResultError("context canceled"), nil
	})
	_, _ = handler(ctx, gomcp.CallToolRequest{})
	if got := scrape(t, metrics); !strings.Contains(got, `outcome="canceled",tool="unknown"} 1`) {
		t.Fatalf("cancellation lost: %s", got)
	}
}

func scrape(t *testing.T, metrics *Metrics) string {
	t.Helper()
	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	if response.Code != 200 {
		t.Fatalf("metrics returned %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "text/plain") {
		t.Fatal("missing Prometheus content type")
	}
	return response.Body.String()
}

func TestMetricsRecordElapsedSeconds(t *testing.T) {
	metrics := NewMetrics(nil)
	handler := metrics.Middleware(func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		time.Sleep(10 * time.Millisecond)
		return gomcp.NewToolResultText("ok"), nil
	})
	_, _ = handler(context.Background(), gomcp.CallToolRequest{})
	families, err := metrics.registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != "premiere_mcp_tool_duration_seconds" {
			continue
		}
		histogram := family.Metric[0].Histogram
		if histogram.GetSampleCount() != 1 || histogram.GetSampleSum() < .01 || histogram.GetSampleSum() > 10 {
			t.Fatalf("latency must record elapsed seconds: %v", histogram)
		}
		return
	}
	t.Fatal("missing latency histogram")
}

func TestMetricsExposeZeroOutcomesBeforeFirstCall(t *testing.T) {
	metrics := NewMetrics(nil)
	metrics.SetTools([]string{"premiere_test"})
	text := scrape(t, metrics)
	for _, outcome := range []string{"success", "error", "canceled", "timeout"} {
		sample := `premiere_mcp_tool_requests_total{outcome="` + outcome + `",tool="premiere_test"} 0`
		if !strings.Contains(text, sample) {
			t.Fatalf("missing zero counter for rate queries: %s", sample)
		}
	}
}
