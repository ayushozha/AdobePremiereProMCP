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
