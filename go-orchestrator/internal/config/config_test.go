package config

import "testing"

func TestDefaultsKeepMutationSurfacesOnLoopback(t *testing.T) {
	cfg := Defaults()
	if cfg.SSEHost != "127.0.0.1" {
		t.Fatalf("SSEHost = %q, want loopback", cfg.SSEHost)
	}
}

func TestLoadFromEnvAllowsExplicitSSEHost(t *testing.T) {
	t.Setenv("MCP_SSE_HOST", "192.0.2.10")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.SSEHost != "192.0.2.10" {
		t.Fatalf("SSEHost = %q, want explicit bind host", cfg.SSEHost)
	}
}

func TestObservabilityOptInAndLoopbackValidation(t *testing.T) {
	t.Setenv("MCP_OBSERVABILITY_ADDR", "")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ObservabilityAddr != "" {
		t.Fatal("observability listener enabled by default")
	}
	for _, address := range []string{"127.0.0.1:9090", "[::1]:9090"} {
		t.Setenv("MCP_OBSERVABILITY_ADDR", address)
		cfg, err := LoadFromEnv()
		if err != nil || cfg.ObservabilityAddr != address {
			t.Fatalf("%s: %+v, %v", address, cfg, err)
		}
	}
	for _, address := range []string{":9090", "0.0.0.0:9090", "192.0.2.1:9090", "localhost:9090", "127.0.0.1:0", "127.0.0.1:65536", "127.0.0.1:http", "127.0.0.1"} {
		t.Setenv("MCP_OBSERVABILITY_ADDR", address)
		if _, err := LoadFromEnv(); err == nil {
			t.Errorf("accepted unsafe/invalid address %q", address)
		}
	}
}
