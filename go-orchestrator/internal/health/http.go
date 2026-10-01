package health

import (
	"encoding/json"
	"net/http"
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
	Ready     bool   `json:"ready"`
	LatencyMs int64  `json:"latency_ms"`
	LastCheck string `json:"last_check,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewHTTPHandler returns an http.Handler that serves health information from
// the provided Checker.
//
// Routes:
//
//	GET /health          — aggregate status of all services
//	GET /health/{service} — status of a single service
//	GET /livez           — process is serving HTTP, independent of dependencies
//	GET /readyz          — all dependencies have recent successful probes
func NewHTTPHandler(checker *Checker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handleAggregateHealth(checker, w, r)
	})
	mux.HandleFunc("GET /health/{service}", func(w http.ResponseWriter, r *http.Request) {
		handleServiceHealth(checker, r.PathValue("service"), w, r)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		handleAggregateHealth(checker, w, r)
	})
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	return mux
}

func handleAggregateHealth(checker *Checker, w http.ResponseWriter, _ *http.Request) {
	statuses := checker.GetAllStatuses()
	ready := len(statuses) > 0

	overall := StatusHealthy
	for _, sh := range statuses {
		ready = ready && serviceReady(sh)
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
		resp.Services[name] = payload(sh)
	}

	statusCode := http.StatusOK
	if !ready {
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

	sp := payload(sh)

	statusCode := http.StatusOK
	if !sp.Ready {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, sp)
}

func payload(sh *ServiceHealth) *servicePayload {
	sp := &servicePayload{Status: sh.Status.String(), Ready: serviceReady(sh), LatencyMs: sh.Latency.Milliseconds()}
	if !sh.LastCheck.IsZero() {
		sp.LastCheck = sh.LastCheck.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if sh.LastError != nil {
		sp.Error = sh.LastError.Error()
	}
	return sp
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
