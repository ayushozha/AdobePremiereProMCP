package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Exercise the real run path in a child so stdin/stdout and flags are isolated.
// This uses only an invalid tool request and disconnected test dependencies.
func TestStdioObservabilityProcess(t *testing.T) {
	if os.Getenv("MCP_OBSERVABILITY_TEST_CHILD") == "1" {
		flag.CommandLine = flag.NewFlagSet("observability-test", flag.ExitOnError)
		os.Args = os.Args[:1]
		if err := run(); err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStdioObservabilityProcess$")
	command.Env = append(os.Environ(), "MCP_OBSERVABILITY_TEST_CHILD=1", "MCP_TRANSPORT=stdio", "MCP_TOOL_PROFILE=standard", "MCP_OBSERVABILITY_ADDR="+address,
		"BRIDGE_CEP_TOKEN=observability-test-token-not-a-real-secret", "RUST_ENGINE_ADDR=127.0.0.1:1", "PYTHON_INTEL_ADDR=127.0.0.1:1", "TS_BRIDGE_ADDR=127.0.0.1:1")
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = command.Process.Kill() }()
	lines := make(chan []byte, 10)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- append([]byte(nil), scanner.Bytes()...)
		}
	}()
	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(3 * time.Second)
	for {
		response, err := client.Get("http://" + address + "/livez")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == 200 {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("observability listener did not become live")
		}
		time.Sleep(10 * time.Millisecond)
	}
	send := func(message string) {
		t.Helper()
		if _, err := io.WriteString(stdin, message+"\n"); err != nil {
			t.Fatal(err)
		}
	}
	receive := func() map[string]any {
		t.Helper()
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatal("MCP stdout closed unexpectedly")
			}
			var response map[string]any
			if err := json.Unmarshal(line, &response); err != nil {
				t.Fatalf("non-JSON MCP stdout: %s", line)
			}
			return response
		case <-ctx.Done():
			t.Fatal("MCP response timed out")
		}
		return nil
	}
	send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`)
	if receive()["result"] == nil {
		t.Fatal("initialize failed")
	}
	send(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	send(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"premiere_scan_assets","arguments":{}}}`)
	response := receive()
	result, ok := response["result"].(map[string]any)
	if !ok || result["isError"] != true {
		t.Fatalf("expected validation tool error: %v", response)
	}
	metricsResponse, err := client.Get("http://" + address + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(metricsResponse.Body)
	metricsResponse.Body.Close()
	if !strings.Contains(string(body), `premiere_mcp_tool_requests_total{outcome="error",tool="premiere_scan_assets"} 1`) {
		t.Fatalf("missing real stdio call metric: %s", body)
	}
	readiness, err := client.Get("http://" + address + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	readiness.Body.Close()
	if readiness.StatusCode != 503 {
		t.Fatalf("disconnected dependencies reported ready: %d", readiness.StatusCode)
	}
	stdin.Close()
	if err := command.Wait(); err != nil {
		t.Fatalf("process exit: %v; stderr: %s", err, stderr.String())
	}
	for line := range lines {
		if !json.Valid(line) {
			t.Fatalf("stdout contaminated: %s", line)
		}
	}
	if !strings.Contains(stderr.String(), `"outcome":"error"`) {
		t.Fatal("missing structured completion log on stderr")
	}
}
