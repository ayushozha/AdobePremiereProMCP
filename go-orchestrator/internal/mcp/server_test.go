package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

func TestNewMCPServer(t *testing.T) {
	s := NewMCPServer(nil, "1.0.0-test", zap.NewNop())
	if s == nil {
		t.Fatal("expected non-nil MCP server")
	}
}

func TestNewMCPServerDefaultVersion(t *testing.T) {
	s := NewMCPServer(nil, "", zap.NewNop())
	if s == nil {
		t.Fatal("expected non-nil MCP server with default version")
	}
}

func TestToolCount(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "all,unsafe")
	s := NewMCPServer(nil, "test", zap.NewNop())

	const expectedToolCount = 1064
	if got := len(s.ListTools()); got != expectedToolCount {
		t.Fatalf("registered tool count = %d, want %d", got, expectedToolCount)
	}
}

func TestDefaultToolProfileFitsModelFunctionLimit(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "")
	tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
	t.Logf("default profile tool count: %d", len(tools))
	if got := len(tools); got != 72 {
		t.Fatalf("default profile contains %d tools, want documented count 72", got)
	}
	if got := len(tools); got > 128 {
		t.Fatalf("default profile contains %d tools, exceeding the 128-function model limit", got)
	}
	for _, name := range []string{
		"premiere_execute_system_command",
		"premiere_split_long_captions",
		"premiere_add_glitch_transition",
		"premiere_apply_film_grain",
		"premiere_undo",
		"premiere_redo",
	} {
		if tools[name] != nil {
			t.Fatalf("default profile exposed unsafe or unverified tool %q", name)
		}
	}
}

func TestCaptionToolProfile(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "captions")
	s := NewMCPServer(nil, "test", zap.NewNop())
	tools := s.ListTools()

	for _, name := range []string{
		"premiere_ping",
		"premiere_get_project",
		"premiere_add_subtitles_from_srt",
		"premiere_get_captions",
		"premiere_validate_closed_captions",
	} {
		if tools[name] == nil {
			t.Fatalf("caption profile omitted %q", name)
		}
	}
	if tools["premiere_apply_video_effect"] != nil {
		t.Fatal("caption profile unexpectedly included an effects-only tool")
	}
	if len(tools) >= 100 {
		t.Fatalf("caption profile contains %d tools, expected a workflow-sized catalog", len(tools))
	}
}

func TestMultipleToolProfiles(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "effects,transitions")
	s := NewMCPServer(nil, "test", zap.NewNop())
	tools := s.ListTools()

	for _, name := range []string{
		"premiere_ping",
		"premiere_apply_video_effect",
		"premiere_add_video_transition",
	} {
		if tools[name] == nil {
			t.Fatalf("combined profile omitted %q", name)
		}
	}
	if tools["premiere_create_proxy"] != nil {
		t.Fatal("combined effects/transitions profile unexpectedly included proxy tools")
	}
}

func TestWorkflowProfilesContainSkillDependencies(t *testing.T) {
	tests := map[string][]string{
		"social": {
			"premiere_duplicate_sequence",
			"premiere_create_vertical_version",
			"premiere_create_square_version",
			"premiere_get_captions",
			"premiere_set_position",
			"premiere_set_scale",
		},
		"effects": {
			"premiere_apply_video_effect",
			"premiere_get_clip_effects",
			"premiere_set_effect_parameter",
		},
		"proxies": {
			"premiere_create_proxy",
			"premiere_attach_proxy",
			"premiere_toggle_proxies",
		},
	}

	for profile, requiredTools := range tests {
		t.Run(profile, func(t *testing.T) {
			t.Setenv("MCP_TOOL_PROFILE", profile)
			tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
			for _, name := range requiredTools {
				if tools[name] == nil {
					t.Fatalf("%s profile omitted skill dependency %q", profile, name)
				}
			}
		})
	}
}

func TestUnknownToolProfileLeavesCatalogIntact(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "typo")
	s := NewMCPServer(nil, "test", zap.NewNop())
	standard := len(s.ListTools())
	t.Setenv("MCP_TOOL_PROFILE", "standard")
	if got := len(NewMCPServer(nil, "test", zap.NewNop()).ListTools()); got != standard {
		t.Fatalf("unknown profile count = %d, want standard count %d", standard, got)
	}
}

func TestUnsafeToolsRequireExplicitProfile(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "all")
	safeTools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
	for _, name := range []string{
		"premiere_execute_system_command",
		"premiere_write_text_file",
		"premiere_if_clip_exists",
		"premiere_if_sequence_open",
		"premiere_if_project_open",
		"premiere_while_condition",
		"premiere_execute_batch",
		"premiere_execute_parallel",
		"premiere_execute_with_retry",
		"premiere_execute_with_timeout",
		"premiere_schedule_script",
		"premiere_schedule_repeating",
	} {
		if safeTools[name] != nil {
			t.Fatalf("all profile exposed unsafe tool %q", name)
		}
	}

	t.Setenv("MCP_TOOL_PROFILE", "all,unsafe")
	unsafeTools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
	if unsafeTools["premiere_execute_system_command"] == nil {
		t.Fatal("explicit unsafe profile omitted system command execution")
	}
}

func TestExactToolProfileEntriesExist(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "all,unsafe")
	tools := NewMCPServer(nil, "test", zap.NewNop()).ListTools()
	for profile, patterns := range toolProfilePatterns {
		for _, pattern := range patterns {
			if strings.ContainsAny(pattern, "*?[") {
				continue
			}
			if tools[pattern] == nil {
				t.Errorf("profile %q references nonexistent tool %q", profile, pattern)
			}
		}
	}
}

func TestToolListPagination(t *testing.T) {
	t.Setenv("MCP_PAGE_SIZE", "7")
	s := NewMCPServer(nil, "test", zap.NewNop())
	c := newInitializedClient(t, s)

	first, err := c.ListToolsByPage(context.Background(), gomcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list first tool page: %v", err)
	}
	if got := len(first.Tools); got != 7 {
		t.Fatalf("first page tool count = %d, want 7", got)
	}
	if first.NextCursor == "" {
		t.Fatal("expected a cursor for the second tool page")
	}

	request := gomcp.ListToolsRequest{}
	request.Params.Cursor = first.NextCursor
	second, err := c.ListToolsByPage(context.Background(), request)
	if err != nil {
		t.Fatalf("list second tool page: %v", err)
	}
	if got := len(second.Tools); got != 7 {
		t.Fatalf("second page tool count = %d, want 7", got)
	}
	if first.Tools[0].Name == second.Tools[0].Name {
		t.Fatalf("pagination repeated first tool %q", first.Tools[0].Name)
	}
}

func TestMCPPageSize(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{name: "default", value: "", want: defaultMCPPageSize},
		{name: "configured", value: "25", want: 25},
		{name: "zero", value: "0", want: defaultMCPPageSize},
		{name: "negative", value: "-3", want: defaultMCPPageSize},
		{name: "not a number", value: "many", want: defaultMCPPageSize},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("MCP_PAGE_SIZE", test.value)
			if got := mcpPageSize(zap.NewNop()); got != test.want {
				t.Fatalf("mcpPageSize() = %d, want %d", got, test.want)
			}
		})
	}
}

func TestRequiredToolArgumentsRejectedBeforeHandler(t *testing.T) {
	handlerCalled := false
	s := mcpserver.NewMCPServer(
		"validation-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	s.AddTool(
		gomcp.NewTool(
			"required-tool",
			gomcp.WithNumber("position", gomcp.Required()),
			gomcp.WithBoolean("enabled", gomcp.Required()),
		),
		func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			handlerCalled = true
			return gomcp.NewToolResultText("called"), nil
		},
	)
	c := newInitializedClient(t, s)

	request := gomcp.CallToolRequest{}
	request.Params.Name = "required-tool"
	request.Params.Arguments = map[string]any{"position": 0}
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call invalid tool request: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected missing required argument to return a tool error: %+v", result)
	}
	if handlerCalled {
		t.Fatal("handler ran despite a missing required argument")
	}
	if text := toolResultText(t, result); !strings.Contains(text, "enabled") {
		t.Fatalf("validation error %q does not name missing property", text)
	}

	request.Params.Arguments = map[string]any{"position": 0, "enabled": false}
	result, err = c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call valid tool request: %v", err)
	}
	if result.IsError {
		t.Fatalf("valid false/zero values were rejected: %s", toolResultText(t, result))
	}
	if !handlerCalled {
		t.Fatal("handler did not run for valid arguments")
	}
}

func TestToolArgumentsRejectInvalidPresentValuesBeforeDestructiveHandler(t *testing.T) {
	handlerCalled := false
	s := mcpserver.NewMCPServer(
		"validation-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	s.AddTool(
		gomcp.NewTool(
			"premiere_remove_clip_from_track",
			gomcp.WithString("track_type", gomcp.Required(), gomcp.Enum("video", "audio")),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Min(0), gomcp.Max(4)),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Min(0)),
			gomcp.WithBoolean("ripple"),
			gomcp.WithString("label", gomcp.MinLength(2), gomcp.MaxLength(4)),
			gomcp.WithArray(
				"clip_ids",
				gomcp.MinItems(1),
				gomcp.MaxItems(2),
				gomcp.WithNumberItems(gomcp.Min(0)),
			),
		),
		func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			handlerCalled = true
			return gomcp.NewToolResultText("called"), nil
		},
	)
	c := newInitializedClient(t, s)

	tests := []struct {
		name      string
		arguments map[string]any
		wantText  string
	}{
		{
			name: "required null",
			arguments: map[string]any{
				"track_type": nil, "track_index": 0, "clip_index": 0,
			},
			wantText: "track_type",
		},
		{
			name: "required wrong type",
			arguments: map[string]any{
				"track_type": "video", "track_index": "0", "clip_index": 0,
			},
			wantText: "track_index",
		},
		{
			name: "optional wrong type",
			arguments: map[string]any{
				"track_type": "video", "track_index": 0, "clip_index": 0, "ripple": "false",
			},
			wantText: "ripple",
		},
		{
			name: "enum",
			arguments: map[string]any{
				"track_type": "titles", "track_index": 0, "clip_index": 0,
			},
			wantText: "one of",
		},
		{
			name: "numeric minimum",
			arguments: map[string]any{
				"track_type": "video", "track_index": -1, "clip_index": 0,
			},
			wantText: ">= 0",
		},
		{
			name: "numeric maximum",
			arguments: map[string]any{
				"track_type": "video", "track_index": 5, "clip_index": 0,
			},
			wantText: "<= 4",
		},
		{
			name: "string minimum",
			arguments: map[string]any{
				"track_type": "video", "track_index": 0, "clip_index": 0, "label": "x",
			},
			wantText: "length must be >= 2",
		},
		{
			name: "string maximum",
			arguments: map[string]any{
				"track_type": "video", "track_index": 0, "clip_index": 0, "label": "abcde",
			},
			wantText: "length must be <= 4",
		},
		{
			name: "array minimum",
			arguments: map[string]any{
				"track_type": "video", "track_index": 0, "clip_index": 0, "clip_ids": []any{},
			},
			wantText: "at least 1 items",
		},
		{
			name: "array maximum",
			arguments: map[string]any{
				"track_type": "video", "track_index": 0, "clip_index": 0, "clip_ids": []any{0, 1, 2},
			},
			wantText: "at most 2 items",
		},
		{
			name: "array item type",
			arguments: map[string]any{
				"track_type": "video", "track_index": 0, "clip_index": 0, "clip_ids": []any{"0"},
			},
			wantText: "[0]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handlerCalled = false
			request := gomcp.CallToolRequest{}
			request.Params.Name = "premiere_remove_clip_from_track"
			request.Params.Arguments = test.arguments
			result, err := c.CallTool(context.Background(), request)
			if err != nil {
				t.Fatalf("call invalid tool request: %v", err)
			}
			if !result.IsError {
				t.Fatalf("expected invalid arguments to return a tool error: %+v", result)
			}
			if handlerCalled {
				t.Fatal("destructive handler ran despite invalid arguments")
			}
			if text := toolResultText(t, result); !strings.Contains(text, test.wantText) {
				t.Fatalf("validation error %q does not contain %q", text, test.wantText)
			}
		})
	}

	handlerCalled = false
	request := gomcp.CallToolRequest{}
	request.Params.Name = "premiere_remove_clip_from_track"
	request.Params.Arguments = map[string]any{
		"track_type":  "video",
		"track_index": 0,
		"clip_index":  0,
		"ripple":      false,
		"label":       "ok",
		"clip_ids":    []any{0},
	}
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call valid tool request: %v", err)
	}
	if result.IsError {
		t.Fatalf("valid false/zero and boundary values were rejected: %s", toolResultText(t, result))
	}
	if !handlerCalled {
		t.Fatal("handler did not run for valid arguments")
	}
}

func TestRequiredToolArgumentsRejectNonObject(t *testing.T) {
	handlerCalled := false
	s := mcpserver.NewMCPServer(
		"validation-test",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithToolHandlerMiddleware(requiredToolArgumentsMiddleware),
	)
	s.AddTool(
		gomcp.NewTool("required-tool", gomcp.WithString("name", gomcp.Required())),
		func(context.Context, gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			handlerCalled = true
			return gomcp.NewToolResultText("called"), nil
		},
	)
	c := newInitializedClient(t, s)

	request := gomcp.CallToolRequest{}
	request.Params.Name = "required-tool"
	request.Params.Arguments = []any{"not", "an", "object"}
	result, err := c.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("call tool with non-object arguments: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected non-object arguments to return a tool error")
	}
	if handlerCalled {
		t.Fatal("handler ran despite non-object arguments")
	}
}

func TestRequiredPromptArgumentsRejectedBeforeHandler(t *testing.T) {
	prompt := gomcp.NewPrompt(
		"test-prompt",
		gomcp.WithArgument("project", gomcp.RequiredArgument()),
		gomcp.WithArgument("style"),
	)
	handlerCalled := false
	handler := validatedPromptHandler(
		prompt,
		func(context.Context, gomcp.GetPromptRequest) (*gomcp.GetPromptResult, error) {
			handlerCalled = true
			return &gomcp.GetPromptResult{}, nil
		},
	)

	request := gomcp.GetPromptRequest{}
	request.Params.Name = prompt.Name
	request.Params.Arguments = map[string]string{"style": "cinematic"}
	_, err := handler(context.Background(), request)
	if err == nil || !strings.Contains(err.Error(), "project") {
		t.Fatalf("expected missing project validation error, got %v", err)
	}
	if handlerCalled {
		t.Fatal("prompt handler ran despite a missing required argument")
	}

	request.Params.Arguments["project"] = "demo"
	if _, err := handler(context.Background(), request); err != nil {
		t.Fatalf("valid prompt request rejected: %v", err)
	}
	if !handlerCalled {
		t.Fatal("prompt handler did not run for valid arguments")
	}
}

func TestRegisteredPromptRejectsMissingRequiredArguments(t *testing.T) {
	s := NewMCPServer(nil, "test", zap.NewNop())
	c := newInitializedClient(t, s)

	request := gomcp.GetPromptRequest{}
	request.Params.Name = "audio-mix"
	_, err := c.GetPrompt(context.Background(), request)
	if err == nil {
		t.Fatal("expected audio-mix prompt without mix_type to be rejected")
	}
	if !strings.Contains(err.Error(), "mix_type") {
		t.Fatalf("prompt validation error %q does not name mix_type", err)
	}
}

func TestAudioMixPromptUsesRegisteredMixerTool(t *testing.T) {
	s := NewMCPServer(nil, "test", zap.NewNop())
	if s.GetTool("premiere_get_audio_mixer_state") == nil {
		t.Fatal("audio mixer state tool is not registered")
	}

	request := gomcp.GetPromptRequest{}
	request.Params.Name = "audio-mix"
	request.Params.Arguments = map[string]string{"mix_type": "dialogue"}
	result, err := handleAudioMixPrompt(context.Background(), request)
	if err != nil {
		t.Fatalf("render audio-mix prompt: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Fatal("audio-mix prompt returned no messages")
	}
	content, ok := result.Messages[0].Content.(gomcp.TextContent)
	if !ok {
		t.Fatalf("audio-mix prompt content type = %T, want mcp.TextContent", result.Messages[0].Content)
	}
	if !strings.Contains(content.Text, "premiere_get_audio_mixer_state") {
		t.Fatal("audio-mix prompt does not reference the registered mixer-state tool")
	}
	if strings.Contains(content.Text, "premiere_get_audio_mix to") {
		t.Fatal("audio-mix prompt still references stale premiere_get_audio_mix tool")
	}
}

func TestResourceCount(t *testing.T) {
	s := NewMCPServer(nil, "test", zap.NewNop())
	c := newInitializedClient(t, s)

	result, err := c.ListResources(context.Background(), gomcp.ListResourcesRequest{})
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	const expectedResourceCount = 5
	if got := len(result.Resources); got != expectedResourceCount {
		t.Fatalf("registered resource count = %d, want %d", got, expectedResourceCount)
	}
}

func TestPromptCount(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "all,unsafe")
	s := NewMCPServer(nil, "test", zap.NewNop())
	c := newInitializedClient(t, s)

	result, err := c.ListPrompts(context.Background(), gomcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("list prompts: %v", err)
	}
	const expectedPromptCount = 5
	if got := len(result.Prompts); got != expectedPromptCount {
		t.Fatalf("registered prompt count = %d, want %d", got, expectedPromptCount)
	}
}

func TestPromptsRespectToolProfile(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "captions")
	s := NewMCPServer(nil, "test", zap.NewNop())
	c := newInitializedClient(t, s)

	result, err := c.ListPrompts(context.Background(), gomcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("list prompts: %v", err)
	}
	if len(result.Prompts) != 0 {
		t.Fatalf("caption-only profile advertised unrelated prompts: %v", result.Prompts)
	}
}

func TestAddTitlesPromptUsesMOGRTWorkflow(t *testing.T) {
	request := gomcp.GetPromptRequest{}
	request.Params.Name = "add-titles"
	request.Params.Arguments = map[string]string{
		"mogrt_path": "/tmp/title.mogrt",
		"title_text": "Verified title",
		"style":      "minimal",
	}
	result, err := handleAddTitlesPrompt(context.Background(), request)
	if err != nil {
		t.Fatalf("render add-titles prompt: %v", err)
	}
	content, ok := result.Messages[0].Content.(gomcp.TextContent)
	if !ok {
		t.Fatalf("add-titles prompt content type = %T, want mcp.TextContent", result.Messages[0].Content)
	}
	if strings.Contains(content.Text, "premiere_add_text") {
		t.Fatal("add-titles prompt references unsupported premiere_add_text")
	}
	for _, name := range []string{"premiere_import_mogrt", "premiere_get_mogrt_properties", "premiere_set_mogrt_text"} {
		if !strings.Contains(content.Text, name) {
			t.Fatalf("add-titles prompt omitted %q", name)
		}
	}
}

func TestSocialExportPromptResolvesAndPassesSequenceID(t *testing.T) {
	request := gomcp.GetPromptRequest{}
	request.Params.Name = "social-export"
	request.Params.Arguments = map[string]string{
		"platform":         "tiktok",
		"output_directory": "/tmp/exports",
	}
	result, err := handleSocialExportPrompt(context.Background(), request)
	if err != nil {
		t.Fatalf("render social-export prompt: %v", err)
	}
	content, ok := result.Messages[0].Content.(gomcp.TextContent)
	if !ok {
		t.Fatalf("social-export prompt content type = %T, want mcp.TextContent", result.Messages[0].Content)
	}
	for _, expected := range []string{
		"premiere_get_timeline with an empty argument object",
		"sequence_id: <sequence_id returned by premiere_get_timeline>",
		"PREMIERE_EXPORT_PRESET_H264_1080P",
	} {
		if !strings.Contains(content.Text, expected) {
			t.Fatalf("social-export prompt omitted %q", expected)
		}
	}
	if strings.Contains(content.Text, "estimated file size") {
		t.Fatal("social-export prompt asks the model to invent an estimated file size")
	}

	request.Params.Arguments["sequence_id"] = "sequence-123"
	result, err = handleSocialExportPrompt(context.Background(), request)
	if err != nil {
		t.Fatalf("render social-export prompt with sequence: %v", err)
	}
	content = result.Messages[0].Content.(gomcp.TextContent)
	if !strings.Contains(content.Text, "sequence_id: sequence-123") {
		t.Fatal("social-export prompt does not pass the requested sequence ID to premiere_export")
	}
}

func newInitializedClient(t *testing.T, s *mcpserver.MCPServer) *client.Client {
	t.Helper()

	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatalf("create in-process MCP client: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("start in-process MCP client: %v", err)
	}
	request := gomcp.InitializeRequest{}
	request.Params.ProtocolVersion = gomcp.LATEST_PROTOCOL_VERSION
	request.Params.ClientInfo = gomcp.Implementation{Name: "mcp-test", Version: "1.0.0"}
	if _, err := c.Initialize(ctx, request); err != nil {
		t.Fatalf("initialize in-process MCP client: %v", err)
	}
	return c
}

func toolResultText(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("tool result contains no content")
	}
	content, ok := result.Content[0].(gomcp.TextContent)
	if !ok {
		t.Fatalf("tool result content type = %T, want mcp.TextContent", result.Content[0])
	}
	return content.Text
}
