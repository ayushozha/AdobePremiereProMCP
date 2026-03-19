package health

import (
	"encoding/json"
	"net/http"
	"strings"
)

// httpResponse is the JSON body returned by the /health endpoint.
type httpResponse struct {
	Status   string                     `json:"status"`
	Ready    bool                       `json:"ready"`
	Services map[string]*servicePayload `json:"services"`
}

// servicePayload is the per-service slice of the health response.
type servicePayload struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// NewHTTPHandler returns an http.Handler that serves health information from
// the provided Checker.
//
// Routes:
//
//	GET /health          — aggregate status of all services
//	GET /health/{service} — status of a single service
func NewHTTPHandler(checker *Checker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// The standard library mux matches "/health" as a prefix, so we
		// distinguish between the aggregate endpoint and the per-service
		// endpoint by inspecting the remainder of the path.
		trimmed := strings.TrimPrefix(r.URL.Path, "/health")
		trimmed = strings.TrimPrefix(trimmed, "/")

		if trimmed != "" {
			handleServiceHealth(checker, trimmed, w, r)
			return
		}
		handleAggregateHealth(checker, w, r)
	})
	return mux
}

func handleAggregateHealth(checker *Checker, w http.ResponseWriter, _ *http.Request) {
	statuses := checker.GetAllStatuses()
	ready := checker.IsReady()

	overall := StatusHealthy
	for _, sh := range statuses {
		if sh.Status > overall {
			overall = sh.Status
		}
	}

	resp := httpResponse{
		Status:   overall.String(),
		Ready:    ready,
		Services: make(map[string]*servicePayload, len(statuses)),
	}
	for name, sh := range statuses {
		sp := &servicePayload{
			Status:    sh.Status.String(),
			LatencyMs: sh.Latency.Milliseconds(),
		}
		if sh.LastError != nil {
			sp.Error = sh.LastError.Error()
		}
		resp.Services[name] = sp
	}

	statusCode := http.StatusOK
	if overall == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, resp)
}

func handleServiceHealth(checker *Checker, serviceName string, w http.ResponseWriter, _ *http.Request) {
	sh := checker.GetStatus(serviceName)
	if sh == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "unknown service: " + serviceName,
		})
		return
	}

	sp := &servicePayload{
		Status:    sh.Status.String(),
		LatencyMs: sh.Latency.Milliseconds(),
	}
	if sh.LastError != nil {
		sp.Error = sh.LastError.Error()
	}

	statusCode := http.StatusOK
	if sh.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, sp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
