package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/config"
	grpcclients "github.com/anthropics/premierpro-mcp/go-orchestrator/internal/grpc"
	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/mcp"
	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/orchestrator"
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
	clients, err := grpcclients.NewClients(&grpcclients.ClientsConfig{
		MediaAddr:    cfg.RustEngineAddr,
		IntelAddr:    cfg.PythonIntelAddr,
		PremiereAddr: cfg.TypeScriptBridgeAddr,
		DialTimeout:  cfg.RustEngineTimeout,
		CallTimeout:  cfg.TypeScriptBridgeTimeout,
	}, logger)
	if err != nil {
		return fmt.Errorf("connecting gRPC clients: %w", err)
	}
	defer clients.Close()

	logger.Info("all gRPC clients connected",
		zap.String("media", cfg.RustEngineAddr),
		zap.String("intel", cfg.PythonIntelAddr),
		zap.String("premiere", cfg.TypeScriptBridgeAddr),
	)

	// ── Orchestrator Engine ──────────────────────────────────────────
	engine := orchestrator.New(
		&grpcclients.MediaAdapter{C: clients.Media},
		&grpcclients.IntelAdapter{C: clients.Intel},
		&grpcclients.PremiereAdapter{C: clients.Premiere},
		logger,
	)

	// ── MCP Server ────────────────────────────────────────────────────
	mcpSrv := mcp.NewMCPServer(engine, version, logger)

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

// ── Helpers ───────────────────────────────────────────────────────────

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
