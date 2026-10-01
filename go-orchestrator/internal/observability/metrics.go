// Package observability exposes opt-in local metrics without touching MCP stdout.
package observability

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Metrics owns a private registry; multiple MCP servers can coexist in a process.
// Labels contain only registered tool names and the fixed completion outcomes.
type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	logger   *zap.Logger
	mu       sync.RWMutex
	tools    map[string]struct{}
}

// NewMetrics creates collectors without registering anything globally.
func NewMetrics(logger *zap.Logger) *Metrics {
	if logger == nil {
		logger = zap.NewNop()
	}
	m := &Metrics{
		registry: prometheus.NewRegistry(), logger: logger,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "premiere_mcp_tool_requests_total", Help: "Completed MCP tool handler calls by tool and outcome.",
		}, []string{"tool", "outcome"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "premiere_mcp_tool_duration_seconds", Help: "MCP tool handler duration in seconds, including validation and backend calls.",
			Buckets: []float64{0.005, 0.025, 0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
		}, []string{"tool"}),
	}
	m.registry.MustRegister(m.requests, m.duration)
	return m
}

// SetTools sets the trusted catalog before serving requests. Request-supplied
// names outside this catalog collapse to a single "unknown" label.
func (m *Metrics) SetTools(names []string) {
	tools := make(map[string]struct{}, len(names))
	for _, name := range names {
		tools[name] = struct{}{}
		for _, outcome := range []string{"success", "error", "canceled", "timeout"} {
			m.requests.WithLabelValues(name, outcome)
		}
	}
	m.mu.Lock()
	m.tools = tools
	m.mu.Unlock()
}

// Handler serves Prometheus exposition over HTTP, never the MCP transport.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

// Middleware records completed tool calls. Install it before recovery and
// validation so their failures are included. It preserves handler results.
func (m *Metrics) Middleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		m.mu.RLock()
		_, known := m.tools[request.Params.Name]
		m.mu.RUnlock()
		tool := "unknown"
		if known {
			tool = request.Params.Name
		}
		started := time.Now()
		result, err := next(ctx, request)
		elapsed := time.Since(started)
		outcome := completionOutcome(ctx, result, err)
		m.requests.WithLabelValues(tool, outcome).Inc()
		m.duration.WithLabelValues(tool).Observe(elapsed.Seconds())
		// Never include args, result content, raw errors, request IDs, or credentials.
		m.logger.Info("MCP tool completed", zap.String("tool", tool), zap.String("outcome", outcome), zap.Float64("duration_seconds", elapsed.Seconds()))
		return result, err
	}
}

func completionOutcome(ctx context.Context, result *gomcp.CallToolResult, err error) string {
	if err == nil && result != nil && !result.IsError {
		return "success"
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) || status.Code(err) == codes.Canceled {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded {
		return "timeout"
	}
	return "error"
}
