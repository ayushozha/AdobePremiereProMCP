package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestHTTPHealthRoutesAndReadiness(t *testing.T) {
	checker := NewChecker(zap.NewNop())
	handler := NewHTTPHandler(checker)
	check := func(path string, code int, body string) {
		t.Helper()
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != code || !strings.Contains(response.Body.String(), body) {
			t.Fatalf("%s: got %d %s; want %d containing %q", path, response.Code, response.Body.String(), code, body)
		}
	}
	check("/livez", 200, `"status": "alive"`)
	check("/readyz", 503, `"ready": false`)
	check("/health/premiere-bridge", 503, `"status": "unhealthy"`)
	check("/health/missing", 404, "unknown service")
	for _, name := range defaultServices {
		checker.RegisterProbe(name, func(context.Context) error { return nil })
	}
	checker.CheckAll(context.Background())
	check("/readyz", 200, `"ready": true`)
	check("/health", 200, `"status": "healthy"`)
	checker.RegisterProbe("intelligence", func(context.Context) error { return errors.New("disconnected") })
	checker.Check(context.Background(), "intelligence")
	check("/livez", 200, `"status": "alive"`)
	check("/readyz", 503, `"ready": false`)
	check("/health", 503, "disconnected")
	check("/health/intelligence", 503, "disconnected")
	request := httptest.NewRecorder()
	handler.ServeHTTP(request, httptest.NewRequest(http.MethodPost, "/livez", nil))
	if request.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /livez = %d", request.Code)
	}
}

func TestCheckReturnsIndependentSnapshot(t *testing.T) {
	checker := NewChecker(zap.NewNop())
	checker.RegisterProbe("premiere-bridge", func(context.Context) error { return nil })
	snapshot := checker.Check(context.Background(), "premiere-bridge")
	snapshot.Status = StatusUnhealthy
	if got := checker.GetStatus("premiere-bridge").Status; got != StatusHealthy {
		t.Fatal("mutating a returned check result changed shared health state")
	}
}
