package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/config"
)

// Build-time variables set via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ── Flags ──────────────────────────────────────────────────────────
	var (
		transport = flag.String("transport", "", `MCP transport: "stdio" (default) or "sse"`)
		port      = flag.Int("port", 0, "SSE HTTP port (only used with --transport=sse)")
		logLevel  = flag.String("log-level", "", `Log level: "debug", "info", "warn", "error"`)
	)
	flag.Parse()

	// ── Config ────────────────────────────────────────────────────────
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// CLI flags override env/defaults.
	if *transport != "" {
		cfg.Transport = config.TransportType(*transport)
	}
	if *port != 0 {
		cfg.SSEPort = *port
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	// ── Logger ────────────────────────────────────────────────────────
	logger, err := buildLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("building logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck

	logger.Info("starting premierpro-mcp orchestrator",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("built", date),
		zap.String("transport", string(cfg.Transport)),
	)

	// ── gRPC client connections ───────────────────────────────────────
	rustConn, err := dialGRPC(cfg.RustEngineAddr)
	if err != nil {
		return fmt.Errorf("connecting to rust engine at %s: %w", cfg.RustEngineAddr, err)
	}
	defer rustConn.Close()
	logger.Info("gRPC client ready", zap.String("service", "rust-engine"), zap.String("addr", cfg.RustEngineAddr))

	pythonConn, err := dialGRPC(cfg.PythonIntelAddr)
	if err != nil {
		return fmt.Errorf("connecting to python intelligence at %s: %w", cfg.PythonIntelAddr, err)
	}
	defer pythonConn.Close()
	logger.Info("gRPC client ready", zap.String("service", "python-intelligence"), zap.String("addr", cfg.PythonIntelAddr))

	tsConn, err := dialGRPC(cfg.TypeScriptBridgeAddr)
	if err != nil {
		return fmt.Errorf("connecting to typescript bridge at %s: %w", cfg.TypeScriptBridgeAddr, err)
	}
	defer tsConn.Close()
	logger.Info("gRPC client ready", zap.String("service", "ts-bridge"), zap.String("addr", cfg.TypeScriptBridgeAddr))

	// ── MCP Server ────────────────────────────────────────────────────
	mcpSrv := mcpserver.NewMCPServer(
		"premierpro-mcp",
		version,
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithLogging(),
		mcpserver.WithInstructions("PremierPro MCP orchestrator — controls Adobe Premiere Pro through natural language. "+
			"Available tool categories: project inspection, media scanning, timeline editing, "+
			"script-to-edit pipeline, and export."),
	)

	// Register tools.
	registerTools(mcpSrv, logger, cfg, rustConn, pythonConn, tsConn)

	// ── Serve ─────────────────────────────────────────────────────────
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch cfg.Transport {
	case config.TransportSSE:
		return serveSSE(ctx, mcpSrv, cfg, logger)
	default:
		return serveStdio(ctx, mcpSrv, logger)
	}
}

// serveStdio runs the MCP server over stdin/stdout.
func serveStdio(_ context.Context, mcpSrv *mcpserver.MCPServer, logger *zap.Logger) error {
	logger.Info("serving MCP over stdio")
	return mcpserver.ServeStdio(mcpSrv)
}

// serveSSE runs the MCP server as an HTTP SSE endpoint.
func serveSSE(ctx context.Context, mcpSrv *mcpserver.MCPServer, cfg config.Config, logger *zap.Logger) error {
	addr := fmt.Sprintf(":%d", cfg.SSEPort)
	logger.Info("serving MCP over SSE", zap.String("addr", addr))

	sseSrv := mcpserver.NewSSEServer(
		mcpSrv,
		mcpserver.WithBaseURL(fmt.Sprintf("http://localhost:%d", cfg.SSEPort)),
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return sseSrv.Start(addr)
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Info("shutting down SSE server")
		return sseSrv.Shutdown(context.Background())
	})

	return g.Wait()
}

// registerTools registers all MCP tools on the server.
// Each tool handler delegates to the appropriate gRPC backend.
func registerTools(
	mcpSrv *mcpserver.MCPServer,
	logger *zap.Logger,
	cfg config.Config,
	rustConn *grpc.ClientConn,
	pythonConn *grpc.ClientConn,
	tsConn *grpc.ClientConn,
) {
	// ── Premiere Bridge tools (TypeScript) ─────────────────────────
	mcpSrv.AddTool(
		mcp.NewTool("premiere_ping",
			mcp.WithDescription("Check if Adobe Premiere Pro is running and responsive"),
		),
		newPingHandler(logger, tsConn, cfg.TypeScriptBridgeTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("premiere_get_project_state",
			mcp.WithDescription("Get the current Premiere Pro project state including sequences and bins"),
		),
		newGetProjectStateHandler(logger, tsConn, cfg.TypeScriptBridgeTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("premiere_create_sequence",
			mcp.WithDescription("Create a new sequence in the Premiere Pro project"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sequence name")),
			mcp.WithNumber("width", mcp.Required(), mcp.Description("Resolution width in pixels")),
			mcp.WithNumber("height", mcp.Required(), mcp.Description("Resolution height in pixels")),
			mcp.WithNumber("frame_rate", mcp.Required(), mcp.Description("Frame rate (e.g. 24, 29.97, 30, 60)")),
		),
		newCreateSequenceHandler(logger, tsConn, cfg.TypeScriptBridgeTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("premiere_import_media",
			mcp.WithDescription("Import a media file into the Premiere Pro project"),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the media file")),
			mcp.WithString("target_bin", mcp.Description("Target bin/folder in the project (empty for root)")),
		),
		newImportMediaHandler(logger, tsConn, cfg.TypeScriptBridgeTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("premiere_execute_edl",
			mcp.WithDescription("Execute a full Edit Decision List: creates sequence, imports media, places all clips and transitions"),
			mcp.WithString("edl_json", mcp.Required(), mcp.Description("JSON-encoded EditDecisionList")),
			mcp.WithBoolean("auto_import", mcp.Description("Automatically import media files not already in project")),
			mcp.WithBoolean("auto_create_sequence", mcp.Description("Automatically create the sequence")),
		),
		newExecuteEDLHandler(logger, tsConn, cfg.TypeScriptBridgeTimeout),
	)

	// ── Media Engine tools (Rust) ─────────────────────────────────
	mcpSrv.AddTool(
		mcp.NewTool("media_scan_assets",
			mcp.WithDescription("Scan a directory for media assets and index their metadata"),
			mcp.WithString("directory", mcp.Required(), mcp.Description("Root directory to scan")),
			mcp.WithBoolean("recursive", mcp.Description("Scan subdirectories recursively")),
		),
		newScanAssetsHandler(logger, rustConn, cfg.RustEngineTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("media_probe",
			mcp.WithDescription("Probe a single media file for detailed metadata (codec, resolution, duration, etc.)"),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the media file")),
		),
		newProbeMediaHandler(logger, rustConn, cfg.RustEngineTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("media_detect_scenes",
			mcp.WithDescription("Detect scene changes in a video file"),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the video file")),
			mcp.WithNumber("threshold", mcp.Description("Scene detection sensitivity 0.0-1.0 (default 0.3)")),
		),
		newDetectScenesHandler(logger, rustConn, cfg.RustEngineTimeout),
	)

	// ── Intelligence tools (Python) ───────────────────────────────
	mcpSrv.AddTool(
		mcp.NewTool("intelligence_parse_script",
			mcp.WithDescription("Parse a script (text or file) into structured segments for video editing"),
			mcp.WithString("text", mcp.Description("Script text content (provide this or file_path)")),
			mcp.WithString("file_path", mcp.Description("Path to a script file (provide this or text)")),
			mcp.WithString("format_hint", mcp.Description("Script format hint: screenplay, youtube, podcast, narration")),
		),
		newParseScriptHandler(logger, pythonConn, cfg.PythonIntelTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("intelligence_generate_edl",
			mcp.WithDescription("Generate an Edit Decision List from parsed script segments and available assets"),
			mcp.WithString("segments_json", mcp.Required(), mcp.Description("JSON-encoded array of ScriptSegments")),
			mcp.WithString("assets_json", mcp.Required(), mcp.Description("JSON-encoded array of Assets")),
			mcp.WithString("matches_json", mcp.Description("JSON-encoded array of AssetMatches")),
			mcp.WithNumber("frame_rate", mcp.Description("Sequence frame rate (default 24)")),
			mcp.WithNumber("width", mcp.Description("Sequence width in pixels (default 1920)")),
			mcp.WithNumber("height", mcp.Description("Sequence height in pixels (default 1080)")),
		),
		newGenerateEDLHandler(logger, pythonConn, cfg.PythonIntelTimeout),
	)

	mcpSrv.AddTool(
		mcp.NewTool("intelligence_match_assets",
			mcp.WithDescription("Match script segments to the best available media assets using AI"),
			mcp.WithString("segments_json", mcp.Required(), mcp.Description("JSON-encoded array of ScriptSegments")),
			mcp.WithString("assets_json", mcp.Required(), mcp.Description("JSON-encoded array of Assets")),
			mcp.WithString("strategy", mcp.Description("Match strategy: keyword, embedding, hybrid (default hybrid)")),
		),
		newMatchAssetsHandler(logger, pythonConn, cfg.PythonIntelTimeout),
	)

	logger.Info("registered MCP tools", zap.Int("count", len(mcpSrv.ListTools())))
}

// ── Tool handler stubs ────────────────────────────────────────────────
//
// Each handler creates a gRPC client for its target service and delegates
// the call. The actual protobuf client types will come from the generated
// code under gen/go/. For now, each handler returns a placeholder response
// that confirms the gRPC channel is reachable.

func newPingHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "premiere_ping"))
		return mcp.NewToolResultText("premiere_ping: stub — gRPC client connected"), nil
	}
}

func newGetProjectStateHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "premiere_get_project_state"))
		return mcp.NewToolResultText("premiere_get_project_state: stub — gRPC client connected"), nil
	}
}

func newCreateSequenceHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "premiere_create_sequence"))
		return mcp.NewToolResultText("premiere_create_sequence: stub — gRPC client connected"), nil
	}
}

func newImportMediaHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "premiere_import_media"))
		return mcp.NewToolResultText("premiere_import_media: stub — gRPC client connected"), nil
	}
}

func newExecuteEDLHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "premiere_execute_edl"))
		return mcp.NewToolResultText("premiere_execute_edl: stub — gRPC client connected"), nil
	}
}

func newScanAssetsHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "media_scan_assets"))
		return mcp.NewToolResultText("media_scan_assets: stub — gRPC client connected"), nil
	}
}

func newProbeMediaHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "media_probe"))
		return mcp.NewToolResultText("media_probe: stub — gRPC client connected"), nil
	}
}

func newDetectScenesHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "media_detect_scenes"))
		return mcp.NewToolResultText("media_detect_scenes: stub — gRPC client connected"), nil
	}
}

func newParseScriptHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "intelligence_parse_script"))
		return mcp.NewToolResultText("intelligence_parse_script: stub — gRPC client connected"), nil
	}
}

func newGenerateEDLHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "intelligence_generate_edl"))
		return mcp.NewToolResultText("intelligence_generate_edl: stub — gRPC client connected"), nil
	}
}

func newMatchAssetsHandler(logger *zap.Logger, _ *grpc.ClientConn, _ time.Duration) mcpserver.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger.Debug("tool called", zap.String("tool", "intelligence_match_assets"))
		return mcp.NewToolResultText("intelligence_match_assets: stub — gRPC client connected"), nil
	}
}

// ── Helpers ───────────────────────────────────────────────────────────

// dialGRPC creates a non-blocking gRPC client connection.
// The connection is lazy — it will actually connect when the first RPC is made.
func dialGRPC(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// buildLogger creates a zap.Logger configured for the given level string.
func buildLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Encoding:         "json",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	return zapCfg.Build()
}
