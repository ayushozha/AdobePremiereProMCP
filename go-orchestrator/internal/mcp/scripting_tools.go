package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerScriptingTools registers scripting, expression, and programmatic
// control MCP tools.
func registerScriptingTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Script Execution (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_evaluate_expression
	s.AddTool(
		gomcp.NewTool("premiere_evaluate_expression",
			gomcp.WithDescription("Evaluate an ExtendScript expression and return the result. Useful for querying state or running one-liners."),
			gomcp.WithString("expression",
				gomcp.Required(),
				gomcp.Description("The ExtendScript expression to evaluate"),
			),
		),
		scriptH(logger, "evaluate_expression", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			expr := gomcp.ParseString(req, "expression", "")
			if expr == "" {
				return gomcp.NewToolResultError("parameter 'expression' is required"), nil
			}
			result, err := orch.EvaluateExpression(ctx, expr)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 2. premiere_execute_script
	s.AddTool(
		gomcp.NewTool("premiere_execute_script",
			gomcp.WithDescription("Execute a .jsx script file in Premiere Pro's ExtendScript engine."),
			gomcp.WithString("script_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .jsx script file to execute"),
			),
		),
		scriptH(logger, "execute_script", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			scriptPath := gomcp.ParseString(req, "script_path", "")
			if scriptPath == "" {
				return gomcp.NewToolResultError("parameter 'script_path' is required"), nil
			}
			result, err := orch.ExecuteScript(ctx, scriptPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 3. premiere_execute_script_with_args
	s.AddTool(
		gomcp.NewTool("premiere_execute_script_with_args",
			gomcp.WithDescription("Execute a .jsx script file with arguments passed as a JSON object. Arguments are available to the script via $.global._scriptArgs."),
			gomcp.WithString("script_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the .jsx script file"),
			),
			gomcp.WithString("args_json",
				gomcp.Required(),
				gomcp.Description("JSON string of arguments to pass to the script"),
			),
		),
		scriptH(logger, "execute_script_with_args", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			scriptPath := gomcp.ParseString(req, "script_path", "")
			if scriptPath == "" {
				return gomcp.NewToolResultError("parameter 'script_path' is required"), nil
			}
			argsJSON := gomcp.ParseString(req, "args_json", "")
			if argsJSON == "" {
				return gomcp.NewToolResultError("parameter 'args_json' is required"), nil
			}
			result, err := orch.ExecuteScriptWithArgs(ctx, scriptPath, argsJSON)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 4. premiere_get_script_result
	s.AddTool(
		gomcp.NewTool("premiere_get_script_result",
			gomcp.WithDescription("Get the result of the last script execution. Returns the value stored by the most recent evaluateExpression, executeScript, or executeScriptWithArgs call."),
		),
		scriptH(logger, "get_script_result", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetScriptResult(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 5. premiere_list_available_scripts
	s.AddTool(
		gomcp.NewTool("premiere_list_available_scripts",
			gomcp.WithDescription("List .jsx script files in a directory. Returns file names and paths for scripts that can be executed."),
			gomcp.WithString("directory",
				gomcp.Required(),
				gomcp.Description("Absolute path to the directory to scan for .jsx files"),
			),
		),
		scriptH(logger, "list_available_scripts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			dir := gomcp.ParseString(req, "directory", "")
			if dir == "" {
				return gomcp.NewToolResultError("parameter 'directory' is required"), nil
			}
			result, err := orch.ListAvailableScripts(ctx, dir)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Variable/State Management (6-9)
	// -----------------------------------------------------------------------

	// 6. premiere_set_global_variable
	s.AddTool(
		gomcp.NewTool("premiere_set_global_variable",
			gomcp.WithDescription("Set a global variable accessible to all scripts. Variables persist until cleared or Premiere Pro restarts."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Variable name"),
			),
			gomcp.WithString("value",
				gomcp.Required(),
				gomcp.Description("Variable value (stored as string; use JSON for complex data)"),
			),
		),
		scriptH(logger, "set_global_variable", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			name := gomcp.ParseString(req, "name", "")
			if name == "" {
				return gomcp.NewToolResultError("parameter 'name' is required"), nil
			}
			value := gomcp.ParseString(req, "value", "")
			if value == "" {
				return gomcp.NewToolResultError("parameter 'value' is required"), nil
			}
			result, err := orch.SetGlobalVariable(ctx, name, value)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 7. premiere_get_global_variable
	s.AddTool(
		gomcp.NewTool("premiere_get_global_variable",
			gomcp.WithDescription("Get the value of a global variable previously set via premiere_set_global_variable."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Variable name to retrieve"),
			),
		),
		scriptH(logger, "get_global_variable", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			name := gomcp.ParseString(req, "name", "")
			if name == "" {
				return gomcp.NewToolResultError("parameter 'name' is required"), nil
			}
			result, err := orch.GetGlobalVariable(ctx, name)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 8. premiere_list_global_variables
	s.AddTool(
		gomcp.NewTool("premiere_list_global_variables",
			gomcp.WithDescription("List all global variables that have been set. Returns variable names and their current values."),
		),
		scriptH(logger, "list_global_variables", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ListGlobalVariables(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 9. premiere_clear_global_variables
	s.AddTool(
		gomcp.NewTool("premiere_clear_global_variables",
			gomcp.WithDescription("Clear all global variables, removing all previously set name-value pairs."),
		),
		scriptH(logger, "clear_global_variables", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.ClearGlobalVariables(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Conditional Operations (10-13)
	// -----------------------------------------------------------------------

	// 10. premiere_if_clip_exists
	s.AddTool(
		gomcp.NewTool("premiere_if_clip_exists",
			gomcp.WithDescription("Conditional execution: run thenScript if a clip exists at the specified track/index, otherwise run elseScript."),
			gomcp.WithString("track_type",
				gomcp.Required(),
				gomcp.Description("Track type: video or audio"),
				gomcp.Enum("video", "audio"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based track index"),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index"),
			),
			gomcp.WithString("then_script",
				gomcp.Required(),
				gomcp.Description("ExtendScript to execute if the clip exists"),
			),
			gomcp.WithString("else_script",
				gomcp.Description("ExtendScript to execute if the clip does not exist (optional)"),
			),
		),
		scriptH(logger, "if_clip_exists", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			trackType := gomcp.ParseString(req, "track_type", "video")
			thenScript := gomcp.ParseString(req, "then_script", "")
			if thenScript == "" {
				return gomcp.NewToolResultError("parameter 'then_script' is required"), nil
			}
			result, err := orch.IfClipExists(ctx, trackType,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseInt(req, "clip_index", 0),
				thenScript,
				gomcp.ParseString(req, "else_script", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 11. premiere_if_sequence_open
	s.AddTool(
		gomcp.NewTool("premiere_if_sequence_open",
			gomcp.WithDescription("Conditional execution: run thenScript if an active sequence is open, otherwise run elseScript."),
			gomcp.WithString("then_script",
				gomcp.Required(),
				gomcp.Description("ExtendScript to execute if a sequence is open"),
			),
			gomcp.WithString("else_script",
				gomcp.Description("ExtendScript to execute if no sequence is open (optional)"),
			),
		),
		scriptH(logger, "if_sequence_open", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			thenScript := gomcp.ParseString(req, "then_script", "")
			if thenScript == "" {
				return gomcp.NewToolResultError("parameter 'then_script' is required"), nil
			}
			result, err := orch.IfSequenceOpen(ctx,
				thenScript,
				gomcp.ParseString(req, "else_script", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 12. premiere_if_project_open
	s.AddTool(
		gomcp.NewTool("premiere_if_project_open",
			gomcp.WithDescription("Conditional execution: run thenScript if a project is open, otherwise run elseScript."),
			gomcp.WithString("then_script",
				gomcp.Required(),
				gomcp.Description("ExtendScript to execute if a project is open"),
			),
			gomcp.WithString("else_script",
				gomcp.Description("ExtendScript to execute if no project is open (optional)"),
			),
		),
		scriptH(logger, "if_project_open", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			thenScript := gomcp.ParseString(req, "then_script", "")
			if thenScript == "" {
				return gomcp.NewToolResultError("parameter 'then_script' is required"), nil
			}
			result, err := orch.IfProjectOpen(ctx,
				thenScript,
				gomcp.ParseString(req, "else_script", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 13. premiere_while_condition
	s.AddTool(
		gomcp.NewTool("premiere_while_condition",
			gomcp.WithDescription("Loop execution: repeatedly run bodyScript while conditionScript evaluates to true, up to maxIterations."),
			gomcp.WithString("condition_script",
				gomcp.Required(),
				gomcp.Description("ExtendScript that returns true/false to control the loop"),
			),
			gomcp.WithString("body_script",
				gomcp.Required(),
				gomcp.Description("ExtendScript to execute on each iteration"),
			),
			gomcp.WithNumber("max_iterations",
				gomcp.Description("Maximum number of iterations to prevent infinite loops (default: 100)"),
			),
		),
		scriptH(logger, "while_condition", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			condScript := gomcp.ParseString(req, "condition_script", "")
			if condScript == "" {
				return gomcp.NewToolResultError("parameter 'condition_script' is required"), nil
			}
			bodyScript := gomcp.ParseString(req, "body_script", "")
			if bodyScript == "" {
				return gomcp.NewToolResultError("parameter 'body_script' is required"), nil
			}
			result, err := orch.WhileCondition(ctx,
				condScript, bodyScript,
				gomcp.ParseInt(req, "max_iterations", 100),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Batch Scripting (14-17)
	// -----------------------------------------------------------------------

	// 14. premiere_execute_batch
	s.AddTool(
		gomcp.NewTool("premiere_execute_batch",
			gomcp.WithDescription("Execute multiple scripts in sequence. Each script runs after the previous one completes. Input is a JSON array of {name, script} objects."),
			gomcp.WithString("scripts",
				gomcp.Required(),
				gomcp.Description("JSON array of script objects: [{\"name\":\"step1\",\"script\":\"...\"},{\"name\":\"step2\",\"script\":\"...\"}]"),
			),
		),
		scriptH(logger, "execute_batch", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			scripts := gomcp.ParseString(req, "scripts", "")
			if scripts == "" {
				return gomcp.NewToolResultError("parameter 'scripts' is required"), nil
			}
			result, err := orch.ExecuteBatch(ctx, scripts)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 15. premiere_execute_parallel
	s.AddTool(
		gomcp.NewTool("premiere_execute_parallel",
			gomcp.WithDescription("Execute scripts that do not depend on each other. Input is a JSON array of {name, script} objects. Note: ExtendScript is single-threaded, so scripts run sequentially but failures are isolated."),
			gomcp.WithString("scripts",
				gomcp.Required(),
				gomcp.Description("JSON array of script objects: [{\"name\":\"task1\",\"script\":\"...\"},{\"name\":\"task2\",\"script\":\"...\"}]"),
			),
		),
		scriptH(logger, "execute_parallel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			scripts := gomcp.ParseString(req, "scripts", "")
			if scripts == "" {
				return gomcp.NewToolResultError("parameter 'scripts' is required"), nil
			}
			result, err := orch.ExecuteParallel(ctx, scripts)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 16. premiere_execute_with_retry
	s.AddTool(
		gomcp.NewTool("premiere_execute_with_retry",
			gomcp.WithDescription("Execute a script with automatic retry on failure. Retries up to maxRetries times with a delay between attempts."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("ExtendScript code to execute"),
			),
			gomcp.WithNumber("max_retries",
				gomcp.Description("Maximum number of retry attempts (default: 3)"),
			),
			gomcp.WithNumber("delay_ms",
				gomcp.Description("Delay in milliseconds between retries (default: 1000)"),
			),
		),
		scriptH(logger, "execute_with_retry", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			result, err := orch.ExecuteWithRetry(ctx, script,
				gomcp.ParseInt(req, "max_retries", 3),
				gomcp.ParseInt(req, "delay_ms", 1000),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 17. premiere_execute_with_timeout
	s.AddTool(
		gomcp.NewTool("premiere_execute_with_timeout",
			gomcp.WithDescription("Execute a script with a timeout. If the script does not complete within the specified time, it is aborted."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("ExtendScript code to execute"),
			),
			gomcp.WithNumber("timeout_ms",
				gomcp.Description("Timeout in milliseconds (default: 30000)"),
			),
		),
		scriptH(logger, "execute_with_timeout", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			result, err := orch.ExecuteWithTimeout(ctx, script,
				gomcp.ParseInt(req, "timeout_ms", 30000),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Timer/Scheduling (18-21)
	// -----------------------------------------------------------------------

	// 18. premiere_schedule_script
	s.AddTool(
		gomcp.NewTool("premiere_schedule_script",
			gomcp.WithDescription("Schedule a script to execute after a specified delay."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("ExtendScript code to execute after the delay"),
			),
			gomcp.WithNumber("delay_ms",
				gomcp.Required(),
				gomcp.Description("Delay in milliseconds before executing the script"),
			),
		),
		scriptH(logger, "schedule_script", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			result, err := orch.ScheduleScript(ctx, script,
				gomcp.ParseInt(req, "delay_ms", 1000),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 19. premiere_schedule_repeating
	s.AddTool(
		gomcp.NewTool("premiere_schedule_repeating",
			gomcp.WithDescription("Schedule a script to execute repeatedly at a fixed interval."),
			gomcp.WithString("script",
				gomcp.Required(),
				gomcp.Description("ExtendScript code to execute on each interval"),
			),
			gomcp.WithNumber("interval_ms",
				gomcp.Required(),
				gomcp.Description("Interval in milliseconds between executions"),
			),
			gomcp.WithNumber("count",
				gomcp.Description("Number of times to repeat (default: 10, 0 = until cancelled)"),
			),
		),
		scriptH(logger, "schedule_repeating", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			script := gomcp.ParseString(req, "script", "")
			if script == "" {
				return gomcp.NewToolResultError("parameter 'script' is required"), nil
			}
			result, err := orch.ScheduleRepeating(ctx, script,
				gomcp.ParseInt(req, "interval_ms", 1000),
				gomcp.ParseInt(req, "count", 10),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 20. premiere_cancel_scheduled_script
	s.AddTool(
		gomcp.NewTool("premiere_cancel_scheduled_script",
			gomcp.WithDescription("Cancel a previously scheduled script by its schedule ID."),
			gomcp.WithString("schedule_id",
				gomcp.Required(),
				gomcp.Description("The schedule ID returned by premiere_schedule_script or premiere_schedule_repeating"),
			),
		),
		scriptH(logger, "cancel_scheduled_script", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			scheduleID := gomcp.ParseString(req, "schedule_id", "")
			if scheduleID == "" {
				return gomcp.NewToolResultError("parameter 'schedule_id' is required"), nil
			}
			result, err := orch.CancelScheduledScript(ctx, scheduleID)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 21. premiere_get_scheduled_scripts
	s.AddTool(
		gomcp.NewTool("premiere_get_scheduled_scripts",
			gomcp.WithDescription("List all active scheduled scripts, including their IDs, intervals, and remaining executions."),
		),
		scriptH(logger, "get_scheduled_scripts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.GetScheduledScripts(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Data Operations (22-28)
	// -----------------------------------------------------------------------

	// 22. premiere_read_json_file
	s.AddTool(
		gomcp.NewTool("premiere_read_json_file",
			gomcp.WithDescription("Read and parse a JSON file from disk. Returns the parsed data."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the JSON file to read"),
			),
		),
		scriptH(logger, "read_json_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			result, err := orch.ReadJSONFile(ctx, filePath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 23. premiere_write_json_file
	s.AddTool(
		gomcp.NewTool("premiere_write_json_file",
			gomcp.WithDescription("Write data as a JSON file to disk."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output JSON file"),
			),
			gomcp.WithString("data",
				gomcp.Required(),
				gomcp.Description("JSON string data to write to the file"),
			),
		),
		scriptH(logger, "write_json_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			data := gomcp.ParseString(req, "data", "")
			if data == "" {
				return gomcp.NewToolResultError("parameter 'data' is required"), nil
			}
			result, err := orch.WriteJSONFile(ctx, filePath, data)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 24. premiere_read_csv_file
	s.AddTool(
		gomcp.NewTool("premiere_read_csv_file",
			gomcp.WithDescription("Read and parse a CSV file. Returns headers and rows as structured data."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the CSV file to read"),
			),
		),
		scriptH(logger, "read_csv_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			result, err := orch.ReadCSVFile(ctx, filePath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 25. premiere_write_csv_file
	s.AddTool(
		gomcp.NewTool("premiere_write_csv_file",
			gomcp.WithDescription("Write data as a CSV file with headers and rows."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output CSV file"),
			),
			gomcp.WithString("headers",
				gomcp.Required(),
				gomcp.Description("JSON array of column header strings, e.g. [\"Name\",\"Time\",\"Duration\"]"),
			),
			gomcp.WithString("rows",
				gomcp.Required(),
				gomcp.Description("JSON array of row arrays, e.g. [[\"clip1\",\"0.0\",\"5.0\"],[\"clip2\",\"5.0\",\"3.0\"]]"),
			),
		),
		scriptH(logger, "write_csv_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			headers := gomcp.ParseString(req, "headers", "")
			if headers == "" {
				return gomcp.NewToolResultError("parameter 'headers' is required"), nil
			}
			rows := gomcp.ParseString(req, "rows", "")
			if rows == "" {
				return gomcp.NewToolResultError("parameter 'rows' is required"), nil
			}
			result, err := orch.WriteCSVFile(ctx, filePath, headers, rows)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 26. premiere_read_text_file
	s.AddTool(
		gomcp.NewTool("premiere_read_text_file",
			gomcp.WithDescription("Read a text file from disk and return its contents as a string."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the text file to read"),
			),
		),
		scriptH(logger, "read_text_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			result, err := orch.ReadTextFile(ctx, filePath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 27. premiere_write_text_file
	s.AddTool(
		gomcp.NewTool("premiere_write_text_file",
			gomcp.WithDescription("Write text content to a file on disk. Overwrites the file if it exists."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output text file"),
			),
			gomcp.WithString("content",
				gomcp.Required(),
				gomcp.Description("Text content to write"),
			),
		),
		scriptH(logger, "write_text_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			content := gomcp.ParseString(req, "content", "")
			result, err := orch.WriteTextFile(ctx, filePath, content)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 28. premiere_append_text_file
	s.AddTool(
		gomcp.NewTool("premiere_append_text_file",
			gomcp.WithDescription("Append text content to an existing file. Creates the file if it does not exist."),
			gomcp.WithString("file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the text file to append to"),
			),
			gomcp.WithString("content",
				gomcp.Required(),
				gomcp.Description("Text content to append"),
			),
		),
		scriptH(logger, "append_text_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			filePath := gomcp.ParseString(req, "file_path", "")
			if filePath == "" {
				return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
			}
			content := gomcp.ParseString(req, "content", "")
			result, err := orch.AppendTextFile(ctx, filePath, content)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// System Integration (29-30)
	// -----------------------------------------------------------------------

	// 29. premiere_open_url
	s.AddTool(
		gomcp.NewTool("premiere_open_url",
			gomcp.WithDescription("Open a URL in the system's default web browser."),
			gomcp.WithString("url",
				gomcp.Required(),
				gomcp.Description("The URL to open"),
			),
		),
		scriptH(logger, "open_url", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			url := gomcp.ParseString(req, "url", "")
			if url == "" {
				return gomcp.NewToolResultError("parameter 'url' is required"), nil
			}
			result, err := orch.OpenURL(ctx, url)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 30. premiere_execute_system_command
	s.AddTool(
		gomcp.NewTool("premiere_execute_system_command",
			gomcp.WithDescription("Execute a system command and return its output. Use with caution as this runs commands on the host system."),
			gomcp.WithString("command",
				gomcp.Required(),
				gomcp.Description("The system command to execute"),
			),
		),
		scriptH(logger, "execute_system_command", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			command := gomcp.ParseString(req, "command", "")
			if command == "" {
				return gomcp.NewToolResultError("parameter 'command' is required"), nil
			}
			result, err := orch.ExecuteSystemCommand(ctx, command)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)
}

// scriptH wraps a handler with debug logging for scripting tool invocations.
func scriptH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
