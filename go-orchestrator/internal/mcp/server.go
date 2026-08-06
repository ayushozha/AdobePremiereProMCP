package mcp

import (
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

const defaultMCPPageSize = 100

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
		server.WithResourceCapabilities(false, true),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
		server.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
		server.WithLogging(),
		server.WithPaginationLimit(mcpPageSize(logger)),
		server.WithInstructions("PremierPro MCP orchestrator — controls Adobe Premiere Pro through natural language. "+
			"Available tool categories: project inspection, media scanning, timeline editing, "+
			"script-to-edit pipeline, and export. "+
			"Read config://premiere-instructions for detailed usage guidance."),
	)

	registerTools(s, orchestrator, logger)
	applyToolProfiles(s, logger)
	registerResources(s)
	registerPrompts(s)

	logger.Info("MCP server initialized",
		zap.String("name", "premierpro-mcp"),
		zap.String("version", version),
	)

	return s
}

func mcpPageSize(logger *zap.Logger) int {
	raw := strings.TrimSpace(os.Getenv("MCP_PAGE_SIZE"))
	if raw == "" {
		return defaultMCPPageSize
	}

	pageSize, err := strconv.Atoi(raw)
	if err != nil || pageSize <= 0 {
		logger.Warn("ignoring invalid MCP_PAGE_SIZE",
			zap.String("value", raw),
			zap.Int("default", defaultMCPPageSize),
		)
		return defaultMCPPageSize
	}
	return pageSize
}
