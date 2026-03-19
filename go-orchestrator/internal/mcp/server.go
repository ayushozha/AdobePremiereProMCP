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
func NewMCPServer(orchestrator Orchestrator, version string, logger *zap.Logger) *server.MCPServer {
	if version == "" {
		version = "dev"
	}

	s := server.NewMCPServer(
		"premierpro-mcp",
		version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
		server.WithInstructions("PremierPro MCP orchestrator — controls Adobe Premiere Pro through natural language. "+
			"Available tool categories: project inspection, media scanning, timeline editing, "+
			"script-to-edit pipeline, and export."),
	)

	registerTools(s, orchestrator, logger)

	logger.Info("MCP server initialized",
		zap.String("name", "premierpro-mcp"),
		zap.String("version", version),
	)

	return s
}
