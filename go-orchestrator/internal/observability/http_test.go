package observability

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ayushozha/AdobePremiereProMCP/go-orchestrator/internal/health"
	"go.uber.org/zap"
)

func TestListenerServesMetricsHealthAndStopsOnCancellation(t *testing.T) {
	server, err := NewServer("127.0.0.1:0", NewMetrics(zap.NewNop()), health.NewChecker(zap.NewNop()))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if _, err := NewServer(server.listener.Addr().String(), NewMetrics(nil), health.NewChecker(nil)); err == nil {
		t.Fatal("occupied port accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- server.Serve(ctx) }()
	client := &http.Client{Timeout: time.Second}
	for _, tc := range []struct {
		path string
		code int
	}{
		{"/metrics", 200}, {"/livez", 200}, {"/readyz", 503}, {"/health/premiere-bridge", 503}, {"/sse", 404},
	} {
		response, err := client.Get("http://" + server.listener.Addr().String() + tc.path)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		if response.StatusCode != tc.code {
			t.Errorf("%s: %d, want %d", tc.path, response.StatusCode, tc.code)
		}
	}
	cancel()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("observability listener did not shut down")
	}
}
