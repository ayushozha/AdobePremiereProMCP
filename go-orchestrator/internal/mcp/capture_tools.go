package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerCaptureTools registers frame capture and secure ExtendScript
// execution MCP tools.
func registerCaptureTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Frame Capture
	// -----------------------------------------------------------------------

	// premiere_capture_frame_base64
	s.AddTool(
		gomcp.NewTool("premiere_capture_frame_base64",
			gomcp.WithDescription("Capture the current frame at the playhead position and return it as a base64-encoded PNG image. "+
				"The image is returned as an MCP image content block that the AI can visually inspect."),
		),
		captH(logger, "capture_frame_base64", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.CaptureFrameAsBase64(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}

			// Build metadata JSON for the text portion
			meta := map[string]any{
				"format":   result.Format,
				"width":    result.Width,
				"height":   result.Height,
				"timecode": result.Timecode,
			}
			metaJSON, _ := json.MarshalIndent(meta, "", "  ")

			// Return as an MCP image content block so the AI can "see" the frame
			return gomcp.NewToolResultImage(string(metaJSON), result.ImageBase64, "image/png"), nil
		}),
	)

	// -----------------------------------------------------------------------
	// Secure ExtendScript Execution
	// -----------------------------------------------------------------------

	// premiere_execute_extendscript
	s.AddTool(
		gomcp.NewTool("premiere_execute_extendscript",
			gomcp.WithDescription("Execute an arbitrary ExtendScript snippet with security validation. "+
				"Dangerous operations (system calls, file deletion, infinite loops, app.quit) are blocked by default. "+
				"Set validate=false to disable security checks (use with caution)."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("The ExtendScript code to execute"),
			),
			gomcp.WithBoolean("validate",
				gomcp.Description("When true (default), runs security checks to block dangerous operations"),
			),
		),
		captH(logger, "execute_extendscript", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			validate := gomcp.ParseBoolean(req, "validate", true)
			result, err := orch.ExecuteSecureScript(ctx, script, validate)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// premiere_execute_qe_script
	s.AddTool(
		gomcp.NewTool("premiere_execute_qe_script",
			gomcp.WithDescription("Execute an ExtendScript snippet with QE DOM access enabled. "+
				"The QE (Quality Engineering) DOM provides access to internal Premiere Pro functionality "+
				"not available through the standard scripting DOM, such as exporting frames, "+
				"managing transitions by name, and accessing internal sequence properties."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("The ExtendScript code to execute with QE DOM access"),
			),
		),
		captH(logger, "execute_qe_script", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			result, err := orch.ExecuteQEScript(ctx, script)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)
}

// captH wraps a handler with debug logging for capture/scripting tool invocations.
func captH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
