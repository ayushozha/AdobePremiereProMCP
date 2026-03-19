package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// NewMCPServer creates and configures an MCP server that exposes all
// Premiere Pro editing tools to AI clients. The returned server is ready
// to be served over stdio or any other transport supported by mcp-go.
//
// The orchestrator parameter provides the concrete implementation that
// each tool handler delegates to for performing actual editing operations.
func NewMCPServer(orchestrator Orchestrator, logger *zap.Logger) *server.MCPServer {
	s := server.NewMCPServer(
		"premierpro-mcp",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
	)

	registerTools(s, orchestrator, logger)

	logger.Info("MCP server initialized",
		zap.String("name", "premierpro-mcp"),
		zap.String("version", "0.1.0"),
	)

	return s
}
