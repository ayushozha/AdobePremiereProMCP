package grpc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateSharedTokenUsesExplicitSecret(t *testing.T) {
	t.Setenv("BRIDGE_CEP_TOKEN", "0123456789abcdef0123456789abcdef")
	t.Setenv("MCP_CEP_TOKEN", "")
	token, err := loadOrCreateSharedToken()
	if err != nil {
		t.Fatalf("loadOrCreateSharedToken() error = %v", err)
	}
	if token != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("unexpected token %q", token)
	}
}

func TestLoadOrCreateSharedTokenCreatesPrivateFile(t *testing.T) {
	t.Setenv("BRIDGE_CEP_TOKEN", "")
	t.Setenv("MCP_CEP_TOKEN", "")
	tokenPath := filepath.Join(t.TempDir(), "auth", "cep-token")
	t.Setenv("PREMIERE_MCP_TOKEN_FILE", tokenPath)

	token, err := loadOrCreateSharedToken()
	if err != nil {
		t.Fatalf("loadOrCreateSharedToken() error = %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("generated token length = %d, want 64", len(token))
	}
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("stat token file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("token mode = %o, want 600", got)
	}
}

func TestLoadOrCreateSharedTokenRejectsWeakSecret(t *testing.T) {
	t.Setenv("BRIDGE_CEP_TOKEN", "too-short")
	if _, err := loadOrCreateSharedToken(); err == nil {
		t.Fatal("expected weak token to be rejected")
	}
}

func TestCustomTokenPathDoesNotChangeExistingParentMode(t *testing.T) {
	t.Setenv("BRIDGE_CEP_TOKEN", "")
	t.Setenv("MCP_CEP_TOKEN", "")
	parent := filepath.Join(t.TempDir(), "shared")
	if err := os.Mkdir(parent, 0o755); err != nil {
		t.Fatalf("create custom parent: %v", err)
	}
	if err := os.Chmod(parent, 0o755); err != nil {
		t.Fatalf("set custom parent mode: %v", err)
	}
	t.Setenv("PREMIERE_MCP_TOKEN_FILE", filepath.Join(parent, "token"))

	if _, err := loadOrCreateSharedToken(); err != nil {
		t.Fatalf("loadOrCreateSharedToken() error = %v", err)
	}
	info, err := os.Stat(parent)
	if err != nil {
		t.Fatalf("stat custom parent: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o755 {
		t.Fatalf("custom parent mode = %o, want 755", got)
	}
}
