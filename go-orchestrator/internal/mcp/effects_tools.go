package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerEffectsTools registers effects and transitions MCP tools.
// Placeholder — tool registrations will be added when the effects gRPC
// bridge is wired.
func registerEffectsTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// TODO: register effects-specific tools here.
}
