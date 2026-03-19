package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerAppTools registers tools for managing the Premiere Pro application
// lifecycle (open, close, check status). These execute OS-level commands
// directly — no gRPC backends needed.
func registerAppTools(s *server.MCPServer, logger *zap.Logger) {
	s.AddTool(
		gomcp.NewTool("premiere_open",
			gomcp.WithDescription("Launch Adobe Premiere Pro. Optionally open a specific project file."),
			gomcp.WithString("project_path", gomcp.Description("Path to a .prproj file to open (optional)")),
			gomcp.WithBoolean("wait", gomcp.Description("Wait for Premiere Pro to finish launching (default: true)")),
		),
		makeOpenHandler(logger),
	)

	s.AddTool(
		gomcp.NewTool("premiere_close",
			gomcp.WithDescription("Quit Adobe Premiere Pro gracefully. Prompts to save unsaved changes."),
			gomcp.WithBoolean("force", gomcp.Description("Force quit without saving (default: false)")),
		),
		makeCloseHandler(logger),
	)

	s.AddTool(
		gomcp.NewTool("premiere_is_running",
			gomcp.WithDescription("Check if Adobe Premiere Pro is currently running as a process."),
		),
		makeIsRunningHandler(logger),
	)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func makeOpenHandler(logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Info("tool:premiere_open")

		projectPath := gomcp.ParseString(req, "project_path", "")
		waitForLaunch := gomcp.ParseBoolean(req, "wait", true)

		// Check if already running.
		if isPremiereRunning() {
			result := map[string]any{
				"status":  "already_running",
				"message": "Adobe Premiere Pro is already running.",
			}
			if projectPath != "" {
				if err := openProjectFile(projectPath); err != nil {
					return gomcp.NewToolResultError(fmt.Sprintf("failed to open project: %v", err)), nil
				}
				result["message"] = fmt.Sprintf("Opened project: %s", projectPath)
				result["project"] = projectPath
			}
			return toolResultJSON(result)
		}

		// Launch Premiere Pro.
		if err := launchPremiere(projectPath); err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed to launch Premiere Pro: %v", err)), nil
		}

		result := map[string]any{
			"status":  "launched",
			"message": "Adobe Premiere Pro is launching.",
		}
		if projectPath != "" {
			result["project"] = projectPath
		}

		if waitForLaunch {
			launched := false
			for i := range 30 {
				time.Sleep(2 * time.Second)
				if isPremiereRunning() {
					launched = true
					result["status"] = "running"
					result["message"] = fmt.Sprintf("Adobe Premiere Pro launched successfully (took ~%ds).", (i+1)*2)
					break
				}
			}
			if !launched {
				result["status"] = "timeout"
				result["message"] = "Premiere Pro was launched but may still be loading (waited 60s)."
			}
		}

		return toolResultJSON(result)
	}
}

func makeCloseHandler(logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Info("tool:premiere_close")

		force := gomcp.ParseBoolean(req, "force", false)

		if !isPremiereRunning() {
			return toolResultJSON(map[string]any{
				"status":  "not_running",
				"message": "Adobe Premiere Pro is not running.",
			})
		}

		if force {
			cmd := exec.Command("pkill", "-9", "-f", "Adobe Premiere Pro")
			_ = cmd.Run()
		} else {
			script := `tell application "Adobe Premiere Pro 2025" to quit`
			cmd := exec.Command("osascript", "-e", script)
			if _, err := cmd.CombinedOutput(); err != nil {
				cmd2 := exec.Command("osascript", "-e", `tell application "Adobe Premiere Pro" to quit`)
				if _, err2 := cmd2.CombinedOutput(); err2 != nil {
					return gomcp.NewToolResultError(fmt.Sprintf("failed to quit Premiere Pro: %v", err2)), nil
				}
			}
		}

		// Wait for it to close.
		for range 15 {
			time.Sleep(1 * time.Second)
			if !isPremiereRunning() {
				return toolResultJSON(map[string]any{
					"status":  "closed",
					"message": "Adobe Premiere Pro has been closed.",
				})
			}
		}

		return toolResultJSON(map[string]any{
			"status":  "closing",
			"message": "Premiere Pro is closing (may be waiting for save prompt).",
		})
	}
}

func makeIsRunningHandler(logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("tool:premiere_is_running")

		running := isPremiereRunning()
		result := map[string]any{"running": running}

		if running {
			if out, err := exec.Command("pgrep", "-f", "Adobe Premiere Pro").Output(); err == nil {
				pids := strings.TrimSpace(string(out))
				result["pids"] = strings.Split(pids, "\n")
			}
		}

		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isPremiereRunning() bool {
	return exec.Command("pgrep", "-f", "Adobe Premiere Pro").Run() == nil
}

func launchPremiere(projectPath string) error {
	args := []string{"-a", "Adobe Premiere Pro 2025"}
	if projectPath != "" {
		args = append(args, projectPath)
	}
	if err := exec.Command("open", args...).Run(); err != nil {
		// Try without year suffix.
		args[1] = "Adobe Premiere Pro"
		return exec.Command("open", args...).Run()
	}
	return nil
}

func openProjectFile(path string) error {
	if err := exec.Command("open", "-a", "Adobe Premiere Pro 2025", path).Run(); err != nil {
		return exec.Command("open", "-a", "Adobe Premiere Pro", path).Run()
	}
	return nil
}
