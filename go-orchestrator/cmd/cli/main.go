// CLI client for testing the PremierPro MCP server.
// Spawns the server as a subprocess and communicates via JSON-RPC over stdio.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// jsonRPCRequest is a JSON-RPC 2.0 request or notification.
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"` // nil for notifications
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// jsonRPCError is the error object in a JSON-RPC 2.0 response.
type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// toolInfo holds the name and description of an MCP tool.
type toolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	serverBin := flag.String("server", "bin/premierpro-mcp", "path to the MCP server binary")
	flag.Parse()

	if err := run(*serverBin); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run(serverBin string) error {
	// ── Spawn the MCP server as a subprocess ──────────────────────
	cmd := exec.Command(serverBin, "--transport", "stdio")
	cmd.Stderr = os.Stderr // let server logs flow to our stderr

	serverStdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("creating stdin pipe: %w", err)
	}
	serverStdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting server %q: %w", serverBin, err)
	}
	defer func() {
		serverStdin.Close()
		_ = cmd.Wait()
	}()

	fmt.Println("PremierPro MCP CLI Test Client")
	fmt.Printf("Server: %s (pid %d)\n\n", serverBin, cmd.Process.Pid)

	// ── Set up response reader goroutine ──────────────────────────
	// The server writes one JSON-RPC message per line to stdout.
	// We read them in a goroutine and dispatch to waiting callers
	// keyed by request ID.
	type pendingCall struct {
		ch chan jsonRPCResponse
	}

	var (
		mu      sync.Mutex
		pending = make(map[int]*pendingCall)
	)

	go func() {
		scanner := bufio.NewScanner(serverStdout)
		// Allow up to 1 MB per line for large responses.
		scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}

			var resp jsonRPCResponse
			if err := json.Unmarshal([]byte(line), &resp); err != nil {
				fmt.Fprintf(os.Stderr, "[cli] failed to parse server output: %v\nraw: %s\n", err, line)
				continue
			}

			// Notifications from the server (no ID) are printed inline.
			if resp.ID == nil {
				fmt.Fprintf(os.Stderr, "[server notification] %s\n", line)
				continue
			}

			mu.Lock()
			pc, ok := pending[*resp.ID]
			if ok {
				delete(pending, *resp.ID)
			}
			mu.Unlock()

			if ok {
				pc.ch <- resp
			} else {
				fmt.Fprintf(os.Stderr, "[cli] unexpected response id=%d\n", *resp.ID)
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "[cli] reader error: %v\n", err)
		}
	}()

	// nextID is the next JSON-RPC request ID.
	nextID := 1

	// sendRequest sends a JSON-RPC request and waits for the response.
	sendRequest := func(method string, params interface{}) (*jsonRPCResponse, error) {
		id := nextID
		nextID++

		req := jsonRPCRequest{
			JSONRPC: "2.0",
			ID:      &id,
			Method:  method,
			Params:  params,
		}

		data, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}

		ch := make(chan jsonRPCResponse, 1)
		mu.Lock()
		pending[id] = &pendingCall{ch: ch}
		mu.Unlock()

		// Write the request as a single line followed by newline.
		if _, err := fmt.Fprintf(serverStdin, "%s\n", data); err != nil {
			mu.Lock()
			delete(pending, id)
			mu.Unlock()
			return nil, fmt.Errorf("writing to server: %w", err)
		}

		resp := <-ch
		return &resp, nil
	}

	// sendNotification sends a JSON-RPC notification (no ID, no response expected).
	sendNotification := func(method string) error {
		req := jsonRPCRequest{
			JSONRPC: "2.0",
			Method:  method,
		}
		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshaling notification: %w", err)
		}
		if _, err := fmt.Fprintf(serverStdin, "%s\n", data); err != nil {
			return fmt.Errorf("writing notification: %w", err)
		}
		return nil
	}

	// prettyPrint formats a JSON-RPC response for display.
	prettyPrint := func(resp *jsonRPCResponse) {
		if resp.Error != nil {
			fmt.Printf("  ERROR [%d]: %s\n", resp.Error.Code, resp.Error.Message)
			if resp.Error.Data != nil {
				var pretty bytes.Buffer
				if err := json.Indent(&pretty, resp.Error.Data, "    ", "  "); err == nil {
					fmt.Printf("    data: %s\n", pretty.String())
				}
			}
			return
		}
		if resp.Result != nil {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, resp.Result, "  ", "  "); err == nil {
				fmt.Printf("  %s\n", pretty.String())
			} else {
				fmt.Printf("  %s\n", string(resp.Result))
			}
		}
	}

	// ── MCP Initialize handshake ─────────────────────────────────
	fmt.Println("Sending initialize...")
	initResp, err := sendRequest("initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "premierpro-cli",
			"version": "0.1.0",
		},
	})
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	fmt.Println("Initialize response:")
	prettyPrint(initResp)
	fmt.Println()

	// Send initialized notification.
	if err := sendNotification("notifications/initialized"); err != nil {
		return fmt.Errorf("sending initialized notification: %w", err)
	}
	fmt.Println("Sent notifications/initialized")
	fmt.Println()

	// ── List tools ───────────────────────────────────────────────
	listAndPrintTools := func() error {
		resp, err := sendRequest("tools/list", map[string]interface{}{})
		if err != nil {
			return fmt.Errorf("tools/list: %w", err)
		}
		if resp.Error != nil {
			prettyPrint(resp)
			return nil
		}

		// Parse the tools list from the result.
		var result struct {
			Tools []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			fmt.Println("Available tools (raw):")
			prettyPrint(resp)
			return nil
		}

		fmt.Printf("Available tools (%d):\n", len(result.Tools))
		for i, t := range result.Tools {
			fmt.Printf("  %d. %-35s %s\n", i+1, t.Name, t.Description)
		}
		fmt.Println()
		return nil
	}

	if err := listAndPrintTools(); err != nil {
		return err
	}

	// ── Interactive REPL ─────────────────────────────────────────
	fmt.Println("Commands:")
	fmt.Println("  <tool_name>   Call a tool (you will be prompted for JSON params)")
	fmt.Println("  list          Re-list available tools")
	fmt.Println("  quit / exit   Exit the CLI")
	fmt.Println()

	userScanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("mcp> ")
		if !userScanner.Scan() {
			break
		}
		input := strings.TrimSpace(userScanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "quit", "exit":
			fmt.Println("Goodbye.")
			return nil

		case "list":
			if err := listAndPrintTools(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}

		case "help":
			fmt.Println("Commands:")
			fmt.Println("  <tool_name>   Call a tool (you will be prompted for JSON params)")
			fmt.Println("  list          Re-list available tools")
			fmt.Println("  quit / exit   Exit the CLI")
			fmt.Println()

		default:
			// Treat input as a tool name.
			toolName := input

			fmt.Printf("Enter JSON arguments for %q (or empty for {}):\n", toolName)
			fmt.Print("params> ")
			if !userScanner.Scan() {
				break
			}
			paramsStr := strings.TrimSpace(userScanner.Text())

			var args map[string]interface{}
			if paramsStr == "" {
				args = map[string]interface{}{}
			} else {
				if err := json.Unmarshal([]byte(paramsStr), &args); err != nil {
					fmt.Fprintf(os.Stderr, "invalid JSON: %v\n", err)
					continue
				}
			}

			fmt.Printf("Calling %s...\n", toolName)
			resp, err := sendRequest("tools/call", map[string]interface{}{
				"name":      toolName,
				"arguments": args,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			fmt.Println("Response:")
			prettyPrint(resp)
			fmt.Println()
		}
	}

	if err := userScanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	return nil
}
