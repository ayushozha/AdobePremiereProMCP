package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

type transitionRecipeRecorder struct {
	Orchestrator
	response string
	err      error
	calls    []struct{ command, arguments string }
}

func (r *transitionRecipeRecorder) EvalCommand(_ context.Context, command, arguments string) (string, error) {
	r.calls = append(r.calls, struct{ command, arguments string }{command, arguments})
	return r.response, r.err
}

func recipeRequest(name string, arguments map[string]any) gomcp.CallToolRequest {
	request := gomcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = arguments
	return request
}

func recipeResultObject(t *testing.T, result *gomcp.CallToolResult) map[string]any {
	t.Helper()
	var object map[string]any
	if err := json.Unmarshal([]byte(toolResultText(t, result)), &object); err != nil {
		t.Fatalf("decode recipe tool result: %v", err)
	}
	return object
}

// Omitting either profile registration would hide the recipes from their intended clients.
func TestTransitionRecipeToolsAvailableInEditingProfiles(t *testing.T) {
	for _, profile := range []string{"standard", "transitions"} {
		t.Run(profile, func(t *testing.T) {
			t.Setenv("MCP_TOOL_PROFILE", profile)
			s := NewMCPServer(nil, "test", zap.NewNop())
			for _, name := range []string{"premiere_list_transition_recipes", "premiere_apply_transition_recipe"} {
				if s.GetTool(name) == nil {
					t.Fatalf("%s profile omitted %s", profile, name)
				}
			}
			tool := s.GetTool("premiere_apply_transition_recipe").Tool
			for _, name := range []string{"recipe_id", "track_index", "clip_index"} {
				if !containsString(tool.InputSchema.Required, name) {
					t.Fatalf("%s is not required", name)
				}
			}
			for _, name := range []string{"track_index", "clip_index", "duration_frames"} {
				property := tool.InputSchema.Properties[name].(map[string]any)
				if property["type"] != "integer" {
					t.Fatalf("%s schema allows fractional values: %v", name, property)
				}
			}
			for _, name := range []string{"dry_run", "duplicate_sequence"} {
				if tool.InputSchema.Properties[name].(map[string]any)["default"] != true {
					t.Fatalf("%s does not default to true", name)
				}
			}
		})
	}
}

// Matching loosely, trusting audio names, or leaking catalog state would report unavailable video recipes as installed.
func TestTransitionRecipesUseExactInstalledVideoNames(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	r := &transitionRecipeRecorder{response: `{"transitions":[{"name":"Cross Dissolve","type":"video","index":0},{"name":"push","type":"video","index":1},{"name":"Flash","type":"audio","index":2},{"name":"Vendor Asset","type":"video","index":3}],"totalCount":4,"videoCount":3,"audioCount":1,"source":"qe"}`}
	c := newInitializedClient(t, NewMCPServer(r, "test", zap.NewNop()))
	result, err := c.CallTool(context.Background(), recipeRequest("premiere_list_transition_recipes", map[string]any{}))
	if err != nil || result.IsError {
		t.Fatalf("list recipes = %v, %v", result, err)
	}
	object := recipeResultObject(t, result)
	recipes, ok := object["recipes"].([]any)
	if !ok || len(recipes) != 12 || object["totalCount"] != float64(12) || object["availableCount"] != float64(1) {
		t.Fatalf("recipe counts: %v", object)
	}
	for _, raw := range recipes {
		recipe := raw.(map[string]any)
		wantAvailable, wantStatus := false, "missing"
		if recipe["recipeId"] == "cross_dissolve" {
			wantAvailable, wantStatus = true, "installed"
		}
		if recipe["available"] != wantAvailable || recipe["availabilityStatus"] != wantStatus {
			t.Fatalf("incorrect availability: %v", recipe)
		}
		if recipe["description"] == "" || recipe["transitionName"] == "Vendor Asset" {
			t.Fatalf("invalid curated recipe: %v", recipe)
		}
	}
	if len(r.calls) != 1 || r.calls[0].command != "getInstalledTransitions" || r.calls[0].arguments != "{}" {
		t.Fatalf("catalog request: %+v", r.calls)
	}
	r.response = `{"transitions":[],"totalCount":0,"source":"qe"}`
	result, err = c.CallTool(context.Background(), recipeRequest("premiere_list_transition_recipes", nil))
	if err != nil || result.IsError || recipeResultObject(t, result)["availableCount"] != float64(0) {
		t.Fatalf("second catalog leaked availability: %v, %v", result, err)
	}
}

// Discovery failure must not be misreported as confirmation that every recipe is missing.
func TestTransitionRecipesKeepAvailabilityUnknownWhenDiscoveryFails(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	for _, test := range []struct {
		name, response string
		err            error
	}{
		{"transport", "", errors.New("Premiere disconnected")},
		{"malformed", `not-json`, nil},
		{"host error", `{"success":false,"error":"QE unavailable"}`, nil},
		{"status error", `{"status":"error","transitions":[]}`, nil},
		{"malformed success flag", `{"success":"false","transitions":[]}`, nil},
		{"null success flag", `{"success":null,"transitions":[]}`, nil},
		{"malformed status", `{"status":false,"transitions":[]}`, nil},
		{"absent catalog", `{}`, nil},
		{"null catalog", `{"transitions":null}`, nil},
		{"wrong catalog type", `{"transitions":"Cross Dissolve"}`, nil},
		{"invalid entry", `{"transitions":[{"name":"Cross Dissolve"}]}`, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &transitionRecipeRecorder{response: test.response, err: test.err}
			c := newInitializedClient(t, NewMCPServer(r, "test", zap.NewNop()))
			result, err := c.CallTool(context.Background(), recipeRequest("premiere_list_transition_recipes", nil))
			if err != nil || result.IsError {
				t.Fatalf("static list unavailable: %v, %v", result, err)
			}
			object := recipeResultObject(t, result)
			if object["availabilityError"] == nil || object["availableCount"] != nil {
				t.Fatalf("unknown discovery counts: %v", object)
			}
			for _, raw := range object["recipes"].([]any) {
				recipe := raw.(map[string]any)
				if recipe["available"] != nil || recipe["availabilityStatus"] != "unknown" {
					t.Fatalf("failure reported availability: %v", recipe)
				}
			}
		})
	}
}

// Resolving an arbitrary transition name or losing dry-run defaults would send unsafe host mutations.
func TestApplyTransitionRecipeResolvesNativeNameAndSafeDefaults(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	for _, test := range []struct {
		id, name string
		frames   int
	}{
		{"cross_dissolve", "Cross Dissolve", 12}, {"blur_dissolve", "Blur Dissolve", 12},
		{"push", "Push", 8}, {"slide", "Slide", 8}, {"whip", "Whip", 6}, {"zoom_blur", "Zoom Blur", 6},
		{"film_dissolve", "Film Dissolve", 12}, {"luma_fade", "Luma Fade", 12}, {"light_leak", "Light Leak", 12},
		{"glitch", "Glitch", 6}, {"flash", "Flash", 6}, {"directional_blur", "Directional Blur", 6},
	} {
		t.Run(test.id, func(t *testing.T) {
			r := &transitionRecipeRecorder{response: `{"status":"ready","requested":{"durationFrames":12},"diagnostics":["dry run"]}`}
			c := newInitializedClient(t, NewMCPServer(r, "test", zap.NewNop()))
			result, err := c.CallTool(context.Background(), recipeRequest("premiere_apply_transition_recipe", map[string]any{"recipe_id": test.id, "track_index": 0, "clip_index": 2}))
			if err != nil || result.IsError {
				t.Fatalf("apply recipe = %v, %v", result, err)
			}
			if len(r.calls) != 1 || r.calls[0].command != "applyTransitionRecipe" {
				t.Fatalf("host dispatch: %+v", r.calls)
			}
			var payload map[string]any
			if err := json.Unmarshal([]byte(r.calls[0].arguments), &payload); err != nil {
				t.Fatal(err)
			}
			want := map[string]any{"recipeId": test.id, "transitionName": test.name, "durationFrames": float64(test.frames), "trackIndex": float64(0), "clipIndex": float64(2), "dryRun": true, "duplicateSequence": true}
			if !reflect.DeepEqual(payload, want) {
				t.Fatalf("host payload = %v, want %v", payload, want)
			}
			if got := recipeResultObject(t, result); got["status"] != "ready" || got["requested"] == nil || got["diagnostics"] == nil {
				t.Fatalf("lost host readback: %v", got)
			}
		})
	}
}

func TestApplyTransitionRecipePreservesExplicitFalseAndDuration(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	r := &transitionRecipeRecorder{response: `{"status":"applied","actual":{"durationFrames":18}}`}
	c := newInitializedClient(t, NewMCPServer(r, "test", zap.NewNop()))
	result, err := c.CallTool(context.Background(), recipeRequest("premiere_apply_transition_recipe", map[string]any{"recipe_id": "push", "track_index": 1, "clip_index": 0, "duration_frames": 18, "dry_run": false, "duplicate_sequence": false}))
	if err != nil || result.IsError {
		t.Fatalf("explicit apply = %v, %v", result, err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(r.calls[0].arguments), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["durationFrames"] != float64(18) || payload["dryRun"] != false || payload["duplicateSequence"] != false {
		t.Fatalf("explicit arguments overwritten: %v", payload)
	}
}

// Legacy transports must unwrap the host envelope without retaining stale envelope fields.
func TestTransitionRecipeLegacyHostEnvelopeRetainsOnlyReadback(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	r := &transitionRecipeRecorder{response: `{"success":true,"data":{"status":"ready","requested":{"durationFrames":8},"diagnostics":[]}}`}
	c := newInitializedClient(t, NewMCPServer(r, "test", zap.NewNop()))
	result, err := c.CallTool(context.Background(), recipeRequest("premiere_apply_transition_recipe", map[string]any{"recipe_id": "push", "track_index": 0, "clip_index": 0}))
	if err != nil || result.IsError {
		t.Fatalf("legacy apply: %v, %v", result, err)
	}
	object := recipeResultObject(t, result)
	if len(object) != 3 || object["status"] != "ready" || object["requested"] == nil || object["diagnostics"] == nil {
		t.Fatalf("legacy envelope leaked into readback: %v", object)
	}
	r.response = `{"success":true,"data":{"transitions":[{"name":"Push","type":"video","index":0}],"totalCount":1,"source":"qe"}}`
	result, err = c.CallTool(context.Background(), recipeRequest("premiere_list_transition_recipes", nil))
	if err != nil || result.IsError || recipeResultObject(t, result)["availableCount"] != float64(1) {
		t.Fatalf("legacy discovery: %v, %v", result, err)
	}
}

// Missing, fractional, overflow, coerced boolean, and custom-name inputs must never reach Premiere.
func TestApplyTransitionRecipeRejectsInvalidArgumentsBeforeDispatch(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	for _, test := range []struct {
		name, key string
		value     any
		remove    bool
	}{
		{"missing recipe", "recipe_id", nil, true}, {"missing track", "track_index", nil, true}, {"missing clip", "clip_index", nil, true},
		{"unknown recipe", "recipe_id", "custom_transition", false}, {"native name as ID", "recipe_id", "Cross Dissolve", false}, {"non-string recipe", "recipe_id", 1, false},
		{"negative track", "track_index", -1, false}, {"fractional clip", "clip_index", 0.5, false}, {"string track", "track_index", "0", false}, {"overflow clip", "clip_index", 1e30, false},
		{"zero duration", "duration_frames", 0, false}, {"negative duration", "duration_frames", -6, false}, {"fractional duration", "duration_frames", 1.5, false}, {"string duration", "duration_frames", "12", false}, {"overflow duration", "duration_frames", 1e30, false},
		{"null duration", "duration_frames", nil, false}, {"null dry run", "dry_run", nil, false}, {"string dry run", "dry_run", "false", false}, {"number duplicate", "duplicate_sequence", 0, false},
		{"custom name override", "transition_name", "Vendor Asset", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &transitionRecipeRecorder{}
			s := NewMCPServer(r, "test", zap.NewNop())
			tool := s.GetTool("premiere_apply_transition_recipe")
			if tool == nil {
				t.Fatal("recipe tool missing")
			}
			arguments := map[string]any{"recipe_id": "push", "track_index": 0, "clip_index": 0}
			if test.remove {
				delete(arguments, test.key)
			} else {
				arguments[test.key] = test.value
			}
			result, err := tool.Handler(context.Background(), recipeRequest("premiere_apply_transition_recipe", arguments))
			if err != nil || !result.IsError || len(r.calls) != 0 {
				t.Fatalf("invalid args reached host: result=%v err=%v calls=%+v", result, err, r.calls)
			}
		})
	}
}

// A returned host failure or malformed readback must not become a successful MCP result.
func TestApplyTransitionRecipeReportsHostFailures(t *testing.T) {
	t.Setenv("MCP_TOOL_PROFILE", "transitions")
	for _, test := range []struct {
		name, response string
		err            error
	}{
		{"transport", "", errors.New("not connected")},
		{"failure", `{"success":false,"error":"native unavailable","diagnostics":["missing"]}`, nil},
		{"status error", `{"status":"error","diagnostics":["duration mismatch"]}`, nil},
		{"empty object", `{}`, nil},
		{"unsupported status", `{"status":"unknown"}`, nil},
		{"malformed", "not-json", nil}, {"null", "null", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &transitionRecipeRecorder{response: test.response, err: test.err}
			c := newInitializedClient(t, NewMCPServer(r, "test", zap.NewNop()))
			result, err := c.CallTool(context.Background(), recipeRequest("premiere_apply_transition_recipe", map[string]any{"recipe_id": "push", "track_index": 0, "clip_index": 0}))
			if err != nil || !result.IsError {
				t.Fatalf("failed host result accepted: %v, %v", result, err)
			}
			if test.name == "failure" && recipeResultObject(t, result)["diagnostics"] == nil {
				t.Fatal("host failure diagnostics lost")
			}
		})
	}
}
