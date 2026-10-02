package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

type transitionRecipe struct {
	RecipeID              string `json:"recipeId"`
	Name                  string `json:"name"`
	TransitionName        string `json:"transitionName"`
	Category              string `json:"category"`
	Description           string `json:"description"`
	DefaultDurationFrames int    `json:"defaultDurationFrames"`
	Available             *bool  `json:"available"`
	AvailabilityStatus    string `json:"availabilityStatus"`
}

// These original editing suggestions refer to installed native transitions;
// they contain no preset files, media, or third-party transition assets.
var transitionRecipes = []transitionRecipe{
	{RecipeID: "cross_dissolve", Name: "Cross Dissolve", TransitionName: "Cross Dissolve", Category: "clean", DefaultDurationFrames: 12, Description: "Blend two related shots for a quiet change of scene."},
	{RecipeID: "blur_dissolve", Name: "Blur Dissolve", TransitionName: "Blur Dissolve", Category: "clean", DefaultDurationFrames: 12, Description: "Soften detail during a dissolve between busy shots."},
	{RecipeID: "push", Name: "Push", TransitionName: "Push", Category: "navigation", DefaultDurationFrames: 8, Description: "Move from one view to the next to suggest progress."},
	{RecipeID: "slide", Name: "Slide", TransitionName: "Slide", Category: "navigation", DefaultDurationFrames: 8, Description: "Reveal the next shot with a short sliding movement."},
	{RecipeID: "whip", Name: "Whip", TransitionName: "Whip", Category: "fast", DefaultDurationFrames: 6, Description: "Connect energetic shots with a quick sweep."},
	{RecipeID: "zoom_blur", Name: "Zoom Blur", TransitionName: "Zoom Blur", Category: "fast", DefaultDurationFrames: 6, Description: "Use a brief zoom burst to emphasize a change in focus."},
	{RecipeID: "film_dissolve", Name: "Film Dissolve", TransitionName: "Film Dissolve", Category: "cinematic", DefaultDurationFrames: 12, Description: "Blend scenes with the native film dissolve for a gentle passage of time."},
	{RecipeID: "luma_fade", Name: "Luma Fade", TransitionName: "Luma Fade", Category: "cinematic", DefaultDurationFrames: 12, Description: "Let brightness guide the handoff between contrasting shots."},
	{RecipeID: "light_leak", Name: "Light Leak", TransitionName: "Light Leak", Category: "cinematic", DefaultDurationFrames: 12, Description: "Add a brief light wash at a scene change."},
	{RecipeID: "glitch", Name: "Glitch", TransitionName: "Glitch", Category: "accent", DefaultDurationFrames: 6, Description: "Mark a sharp change in rhythm with a short digital disturbance."},
	{RecipeID: "flash", Name: "Flash", TransitionName: "Flash", Category: "accent", DefaultDurationFrames: 6, Description: "Punctuate a cut with a quick flash."},
	{RecipeID: "directional_blur", Name: "Directional Blur", TransitionName: "Directional Blur", Category: "fast", DefaultDurationFrames: 6, Description: "Carry movement through a cut with a short directional blur."},
}

func registerTransitionRecipeTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	s.AddTool(gomcp.NewTool("premiere_list_transition_recipes",
		gomcp.WithDescription("List 12 original editing recipes using native Premiere video transitions, with exact installed-name availability from this Premiere instance. Discovery failure leaves availability unknown. Installed availability does not establish that a transition was applied or visually tested."),
		gomcp.WithSchemaAdditionalProperties(false),
	), func(ctx context.Context, _ gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_list_transition_recipes")
		recipes := append([]transitionRecipe(nil), transitionRecipes...)
		var availableCount *int
		availabilitySource, availabilityError := "", ""
		raw, err := orch.EvalCommand(ctx, "getInstalledTransitions", "{}")
		if err == nil {
			var installed map[string]bool
			installed, err = installedRecipeTransitions(raw)
			if err == nil {
				count := 0
				availableCount, availabilitySource = &count, "qe"
				for i := range recipes {
					available := installed[recipes[i].TransitionName]
					recipes[i].Available = &available
					recipes[i].AvailabilityStatus = "missing"
					if available {
						recipes[i].AvailabilityStatus = "installed"
						count++
					}
				}
			}
		}
		if err != nil {
			availabilityError = err.Error()
			for i := range recipes {
				recipes[i].AvailabilityStatus = "unknown"
			}
		}
		return toolResultJSON(struct {
			Recipes            []transitionRecipe `json:"recipes"`
			TotalCount         int                `json:"totalCount"`
			AvailableCount     *int               `json:"availableCount"`
			AvailabilitySource string             `json:"availabilitySource,omitempty"`
			AvailabilityError  string             `json:"availabilityError,omitempty"`
		}{recipes, len(recipes), availableCount, availabilitySource, availabilityError})
	})

	ids := make([]string, 0, len(transitionRecipes))
	for _, recipe := range transitionRecipes {
		ids = append(ids, recipe.RecipeID)
	}
	integer := func(property map[string]any) { property["type"] = "integer" }
	// ExtendScript uses IEEE-754 numbers. Reject values it cannot represent exactly.
	const maxSafeInteger = 1<<53 - 1
	applyTool := gomcp.NewTool("premiere_apply_transition_recipe",
		gomcp.WithDescription("Plan or apply one curated native video transition recipe at the outgoing cut of a clip. Defaults to a dry run; an actual application duplicates the active sequence by default. Host results include sequence identity and requested/actual readback. Use premiere_list_transition_recipes to choose an exact recipe ID."),
		gomcp.WithSchemaAdditionalProperties(false),
		gomcp.WithString("recipe_id", gomcp.Required(), gomcp.Enum(ids...), gomcp.Description("Recipe ID returned by premiere_list_transition_recipes. Custom transition names are not accepted.")),
		gomcp.WithNumber("track_index", gomcp.Required(), integer, gomcp.Min(0), gomcp.Max(maxSafeInteger), gomcp.Description("Zero-based video track index.")),
		gomcp.WithNumber("clip_index", gomcp.Required(), integer, gomcp.Min(0), gomcp.Max(maxSafeInteger), gomcp.Description("Zero-based outgoing clip index on that track.")),
		gomcp.WithNumber("duration_frames", integer, gomcp.Min(1), gomcp.Max(maxSafeInteger), gomcp.Description("Positive whole frame count. Omit to use the selected recipe's defaultDurationFrames.")),
		gomcp.WithBoolean("dry_run", gomcp.DefaultBool(true), gomcp.Description("Plan without timeline changes (default: true). Pass false to apply.")),
		gomcp.WithBoolean("duplicate_sequence", gomcp.DefaultBool(true), gomcp.Description("Duplicate the active sequence before application (default: true).")),
	)
	s.AddTool(applyTool, func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_apply_transition_recipe")
		arguments, err := objectArguments(req.Params.Arguments)
		if err != nil {
			return gomcp.NewToolResultErrorf("invalid recipe arguments: %v", err), nil
		}
		for name := range arguments {
			if _, known := applyTool.InputSchema.Properties[name]; !known {
				return gomcp.NewToolResultErrorf("unsupported recipe argument: %s", name), nil
			}
		}
		schema, err := inputSchema(applyTool)
		if err == nil {
			err = validateObject(arguments, schema, "")
		}
		if err != nil {
			return gomcp.NewToolResultErrorf("invalid recipe arguments: %v", err), nil
		}
		var recipe transitionRecipe
		for _, candidate := range transitionRecipes {
			if candidate.RecipeID == arguments["recipe_id"] {
				recipe = candidate
				break
			}
		}
		duration := recipe.DefaultDurationFrames
		if value, present := arguments["duration_frames"]; present {
			duration = int(value.(float64))
		}
		dryRun, duplicate := true, true
		if value, present := arguments["dry_run"]; present {
			dryRun = value.(bool)
		}
		if value, present := arguments["duplicate_sequence"]; present {
			duplicate = value.(bool)
		}
		payload, err := json.Marshal(map[string]any{
			"recipeId": recipe.RecipeID, "transitionName": recipe.TransitionName,
			"durationFrames": duration, "trackIndex": int(arguments["track_index"].(float64)),
			"clipIndex": int(arguments["clip_index"].(float64)), "dryRun": dryRun, "duplicateSequence": duplicate,
		})
		if err != nil {
			return gomcp.NewToolResultErrorf("encode recipe arguments: %v", err), nil
		}
		raw, err := orch.EvalCommand(ctx, "applyTransitionRecipe", string(payload))
		if err != nil {
			return gomcp.NewToolResultErrorf("apply transition recipe failed: %v", err), nil
		}
		result, err := recipeHostObject(raw)
		if err != nil {
			return gomcp.NewToolResultErrorf("invalid recipe host result: %v", err), nil
		}
		var success bool
		var status string
		_ = json.Unmarshal(result["success"], &success)
		_ = json.Unmarshal(result["status"], &status)
		if recipeHostFailure(result) == nil && !success && status != "ready" && status != "applied" {
			return gomcp.NewToolResultError("invalid recipe host result: expected status ready/applied or success=true"), nil
		}
		response, err := toolResultJSON(result)
		if err == nil {
			response.IsError = recipeHostFailure(result) != nil
		}
		return response, err
	})
}

func recipeHostObject(raw string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, fmt.Errorf("expected a JSON object")
	}
	var success *bool
	if value, present := object["success"]; present {
		if json.Unmarshal(value, &success) != nil || success == nil {
			return nil, fmt.Errorf("host success must be a boolean")
		}
	}
	if value, present := object["status"]; present {
		var status *string
		if json.Unmarshal(value, &status) != nil || status == nil {
			return nil, fmt.Errorf("host status must be a string")
		}
	}
	// The gRPC adapter normally unwraps the legacy _ok envelope. Also handle it
	// here for callers providing an Orchestrator with an older transport.
	if len(object) == 2 && success != nil && *success && object["data"] != nil {
		return recipeHostObject(string(object["data"]))
	}
	return object, nil
}

func recipeHostFailure(object map[string]json.RawMessage) error {
	var success *bool
	var status string
	_ = json.Unmarshal(object["success"], &success)
	_ = json.Unmarshal(object["status"], &status)
	if (success != nil && !*success) || status == "error" {
		var message string
		_ = json.Unmarshal(object["error"], &message)
		if message == "" {
			message = "Premiere reported a transition recipe error"
		}
		return fmt.Errorf("%s", message)
	}
	return nil
}

func installedRecipeTransitions(raw string) (map[string]bool, error) {
	object, err := recipeHostObject(raw)
	if err != nil {
		return nil, err
	}
	if err := recipeHostFailure(object); err != nil {
		return nil, err
	}
	var entries []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	value, present := object["transitions"]
	if !present || string(value) == "null" {
		return nil, fmt.Errorf("Premiere did not return a transition catalog")
	}
	if err := json.Unmarshal(value, &entries); err != nil {
		return nil, fmt.Errorf("invalid Premiere transition catalog: %w", err)
	}
	installed := make(map[string]bool)
	for _, entry := range entries {
		if entry.Name == "" || (entry.Type != "video" && entry.Type != "audio") {
			return nil, fmt.Errorf("invalid Premiere transition catalog entry")
		}
		if entry.Type == "video" {
			installed[entry.Name] = true
		}
	}
	return installed, nil
}
