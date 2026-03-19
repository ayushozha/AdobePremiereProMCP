package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
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
			gomcp.WithDescription("Launch Adobe Premiere Pro on macOS. If Premiere is already running and a project_path is provided, the project is opened in the existing instance. When wait is true (default), the call blocks until Premiere finishes launching (up to 60 seconds). Returns status ('launched', 'already_running', 'timeout') and timing information."),
			gomcp.WithString("project_path", gomcp.Description("Absolute path to a .prproj file to open on launch (e.g. '/Users/me/Projects/MyEdit.prproj'). If omitted, Premiere opens with no project.")),
			gomcp.WithBoolean("wait", gomcp.Description("If true (default), block until Premiere Pro is confirmed running or 60 seconds elapse. Set to false for fire-and-forget launch.")),
		),
		makeOpenHandler(logger),
	)

	s.AddTool(
		gomcp.NewTool("premiere_close",
			gomcp.WithDescription("Quit Adobe Premiere Pro. By default, sends a graceful quit via AppleScript, which may trigger a 'Save changes?' dialog. Use force=true to kill the process immediately without saving. Returns status ('closed', 'closing', 'not_running')."),
			gomcp.WithBoolean("force", gomcp.Description("If true, force-kill the process (SIGKILL) without saving. If false (default), send a graceful quit that allows Premiere to prompt for unsaved changes.")),
		),
		makeCloseHandler(logger),
	)

	s.AddTool(
		gomcp.NewTool("premiere_is_running",
			gomcp.WithDescription("Check whether Adobe Premiere Pro is currently running as a macOS process. Returns {running: true/false} and, when running, the process IDs. Use this before calling tools that require Premiere to be open."),
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
			name := premiereAppName()
			script := fmt.Sprintf(`tell application "%s" to quit`, name)
			cmd := exec.Command("osascript", "-e", script)
			if _, err := cmd.CombinedOutput(); err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed to quit Premiere Pro: %v", err)), nil
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

// premiereAppName discovers the installed Premiere Pro application name.
// Tries year-suffixed names from newest to oldest, then the bare name.
func premiereAppName() string {
	for _, year := range []string{"2026", "2025", "2024", "2023"} {
		name := "Adobe Premiere Pro " + year
		if err := exec.Command("open", "-Ra", name).Run(); err == nil {
			return name
		}
	}
	return "Adobe Premiere Pro"
}

func launchPremiere(projectPath string) error {
	name := premiereAppName()
	args := []string{"-a", name}
	if projectPath != "" {
		args = append(args, projectPath)
	}
	return exec.Command("open", args...).Run()
}

func openProjectFile(path string) error {
	name := premiereAppName()
	return exec.Command("open", "-a", name, path).Run()
}

// ---------------------------------------------------------------------------
// Health monitor
// ---------------------------------------------------------------------------

// registerHealthTools registers tools that need both OS-level checks and
// orchestrator access (e.g. ping via the bridge).
func registerHealthTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	s.AddTool(
		gomcp.NewTool("premiere_health_monitor",
			gomcp.WithDescription(
				"Run a comprehensive health check on the Premiere Pro pipeline. "+
					"Verifies that the Premiere Pro process is running, that the WebSocket bridge "+
					"is reachable (via premiere_ping), and reports an overall health status. "+
					"If Premiere Pro has crashed, the response includes a 'restart_available' flag "+
					"indicating you can call premiere_open to relaunch it. "+
					"Returns: process_running, bridge_connected, premiere_version, project_open, "+
					"overall_status ('healthy', 'degraded', 'down'), and restart_available."),
			gomcp.WithBoolean("attempt_restart",
				gomcp.Description("If true and Premiere Pro is not running, automatically attempt to relaunch it. Default false."),
			),
		),
		makeHealthMonitorHandler(orch, logger),
	)
}

func makeHealthMonitorHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Info("tool:premiere_health_monitor")

		attemptRestart := gomcp.ParseBoolean(req, "attempt_restart", false)

		result := map[string]any{
			"process_running":   false,
			"bridge_connected":  false,
			"premiere_version":  "unknown",
			"project_open":      false,
			"overall_status":    "down",
			"restart_available": false,
		}

		// Step 1: Check if the Premiere Pro process is running (OS-level).
		processRunning := isPremiereRunning()
		result["process_running"] = processRunning

		if !processRunning {
			result["overall_status"] = "down"
			result["message"] = "Adobe Premiere Pro is not running."

			// On macOS/Linux we can offer to restart.
			if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
				result["restart_available"] = true
			}

			if attemptRestart {
				logger.Info("Premiere Pro not running -- attempting restart")
				if err := launchPremiere(""); err != nil {
					result["message"] = fmt.Sprintf("Premiere Pro is not running and restart failed: %v", err)
					return toolResultJSON(result)
				}

				// Wait up to 30 seconds for Premiere to start.
				launched := false
				for i := range 15 {
					time.Sleep(2 * time.Second)
					if isPremiereRunning() {
						launched = true
						result["process_running"] = true
						result["message"] = fmt.Sprintf(
							"Premiere Pro was relaunched successfully (took ~%ds). "+
								"Bridge may need a moment to reconnect.", (i+1)*2,
						)
						break
					}
				}
				if !launched {
					result["message"] = "Restart was initiated but Premiere Pro has not appeared after 30s."
					return toolResultJSON(result)
				}
			} else {
				return toolResultJSON(result)
			}
		}

		// Step 2: Check bridge connectivity via ping.
		pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		pingResult, err := orch.Ping(pingCtx)
		if err != nil {
			logger.Warn("health monitor: ping failed", zap.Error(err))
			result["bridge_connected"] = false
			result["overall_status"] = "degraded"
			result["message"] = fmt.Sprintf(
				"Premiere Pro process is running but the bridge is unreachable: %v. "+
					"The CEP panel may not be loaded -- open Window > Extensions > PremierPro MCP Bridge.",
				err,
			)
			return toolResultJSON(result)
		}

		// Ping succeeded -- populate details.
		result["bridge_connected"] = pingResult.PremiereRunning
		result["premiere_version"] = pingResult.PremiereVersion
		result["project_open"] = pingResult.ProjectOpen

		if pingResult.PremiereRunning {
			result["overall_status"] = "healthy"
			result["message"] = "Premiere Pro is running and the bridge is connected."
		} else {
			result["overall_status"] = "degraded"
			result["message"] = "Bridge responded but reports Premiere Pro is not fully ready."
		}

		return toolResultJSON(result)
	}
}
