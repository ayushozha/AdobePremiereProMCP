package mcp

import (
	"strings"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

func TestProxyProjectItemIndexRejectsLossyValues(t *testing.T) {
	tests := []struct {
		name      string
		arguments any
		want      int
		wantError bool
	}{
		{name: "zero", arguments: map[string]any{"project_item_index": 0}, want: 0},
		{name: "positive", arguments: map[string]any{"project_item_index": 12}, want: 12},
		{name: "missing", arguments: map[string]any{}, wantError: true},
		{name: "negative", arguments: map[string]any{"project_item_index": -1}, wantError: true},
		{name: "fractional", arguments: map[string]any{"project_item_index": 1.5}, wantError: true},
		{name: "overflow", arguments: map[string]any{"project_item_index": 1e30}, wantError: true},
		{name: "string", arguments: map[string]any{"project_item_index": "1"}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := gomcp.CallToolRequest{}
			request.Params.Arguments = test.arguments
			got, err := proxyProjectItemIndex(request)
			if test.wantError {
				if err == nil {
					t.Fatalf("proxyProjectItemIndex() = %d, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("proxyProjectItemIndex() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("proxyProjectItemIndex() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestProxyToolSchemasDescribeVerifiedWorkflow(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "proxies")
	tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()

	createTool := tools["premiere_create_proxy"]
	if createTool == nil {
		t.Fatal("proxies profile omitted premiere_create_proxy")
	}
	if !strings.Contains(createTool.Tool.Description, "asynchronous") || !strings.Contains(createTool.Tool.Description, "Poll") {
		t.Fatalf("create proxy description does not explain asynchronous verification: %q", createTool.Tool.Description)
	}
	for _, required := range []string{"project_item_index", "output_path", "preset_path"} {
		if !containsString(createTool.Tool.InputSchema.Required, required) {
			t.Fatalf("create proxy required parameters = %v, want %s", createTool.Tool.InputSchema.Required, required)
		}
	}

	for _, toolName := range []string{
		"premiere_create_proxy",
		"premiere_attach_proxy",
		"premiere_has_proxy",
		"premiere_get_proxy_path",
		"premiere_detach_proxy",
	} {
		registered := tools[toolName]
		if registered == nil {
			t.Fatalf("proxies profile omitted %s", toolName)
		}
		property, ok := registered.Tool.InputSchema.Properties["project_item_index"].(map[string]any)
		if !ok {
			t.Fatalf("%s project_item_index schema = %T, want map[string]any", toolName, registered.Tool.InputSchema.Properties["project_item_index"])
		}
		if minimum, ok := property["minimum"].(float64); !ok || minimum != 0 {
			t.Fatalf("%s project_item_index minimum = %v, want 0", toolName, property["minimum"])
		}
	}

	toggleTool := tools["premiere_toggle_proxies"]
	if toggleTool == nil || !strings.Contains(toggleTool.Tool.Description, "verified") {
		t.Fatalf("toggle proxy description does not promise only verified readback: %+v", toggleTool)
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
