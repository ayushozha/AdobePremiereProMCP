package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// requiredToolArgumentsMiddleware validates calls against the subset of JSON
// Schema used by the registered tools. mcp-go exposes schemas to clients, but
// it intentionally leaves validation to handlers; centralizing validation here
// keeps malformed values from being interpreted as destructive zero/default
// values by Premiere Pro or any of the backend services.
func requiredToolArgumentsMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		mcpServer := server.ServerFromContext(ctx)
		if mcpServer == nil {
			return next(ctx, request)
		}

		registeredTool := mcpServer.GetTool(request.Params.Name)
		if registeredTool == nil {
			return next(ctx, request)
		}

		schema, err := inputSchema(registeredTool.Tool)
		if err != nil {
			return gomcp.NewToolResultErrorf(
				"invalid input schema for tool %q: %v",
				request.Params.Name,
				err,
			), nil
		}
		arguments, err := objectArguments(request.Params.Arguments)
		if err != nil {
			return gomcp.NewToolResultErrorf(
				"invalid arguments for tool %q: expected a JSON object: %v",
				request.Params.Name,
				err,
			), nil
		}

		if err := validateObject(arguments, schema, ""); err != nil {
			var schemaErr *inputSchemaValidationError
			if errors.As(err, &schemaErr) {
				return gomcp.NewToolResultErrorf(
					"invalid input schema for tool %q: %v",
					request.Params.Name,
					schemaErr,
				), nil
			}
			return gomcp.NewToolResultErrorf(
				"invalid arguments for tool %q: %v",
				request.Params.Name,
				err,
			), nil
		}

		return next(ctx, request)
	}
}

type inputSchemaValidationError struct {
	message string
}

func (e *inputSchemaValidationError) Error() string {
	return e.message
}

func inputSchemaErrorf(format string, args ...any) error {
	return &inputSchemaValidationError{message: fmt.Sprintf(format, args...)}
}

func inputSchema(tool gomcp.Tool) (map[string]any, error) {
	encoded := tool.RawInputSchema
	if len(encoded) == 0 {
		var err error
		encoded, err = json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("encode input schema: %w", err)
		}
	}

	var schema map[string]any
	if err := json.Unmarshal(encoded, &schema); err != nil {
		return nil, fmt.Errorf("decode input schema: %w", err)
	}
	if schema == nil {
		return nil, fmt.Errorf("input schema must be a JSON object")
	}
	return schema, nil
}

func objectArguments(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode arguments: %w", err)
	}

	var arguments map[string]any
	if err := json.Unmarshal(encoded, &arguments); err != nil {
		return nil, fmt.Errorf("decode arguments: %w", err)
	}
	return arguments, nil
}

func validateObject(arguments map[string]any, schema map[string]any, path string) error {
	required, err := requiredProperties(schema)
	if err != nil {
		return err
	}

	missing := make([]string, 0, len(required))
	for _, property := range required {
		if _, ok := arguments[property]; !ok {
			missing = append(missing, property)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		if path == "" {
			return fmt.Errorf("missing required properties: %s", strings.Join(missing, ", "))
		}
		return fmt.Errorf("%s is missing required properties: %s", path, strings.Join(missing, ", "))
	}

	properties, err := schemaProperties(schema)
	if err != nil {
		return err
	}
	propertyNames := make([]string, 0, len(properties))
	for name := range properties {
		propertyNames = append(propertyNames, name)
	}
	sort.Strings(propertyNames)

	for _, name := range propertyNames {
		value, present := arguments[name]
		if !present {
			continue
		}
		propertySchema, allowed, err := schemaMap(properties[name], fmt.Sprintf("property %q", name))
		if err != nil {
			return err
		}
		propertyPath := joinPropertyPath(path, name)
		if !allowed {
			return fmt.Errorf("%s is not allowed", propertyPath)
		}
		if err := validateValue(value, propertySchema, propertyPath); err != nil {
			return err
		}
	}

	return nil
}

func requiredProperties(schema map[string]any) ([]string, error) {
	value, ok := schema["required"]
	if !ok || value == nil {
		return nil, nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil, inputSchemaErrorf("required must be an array of strings")
	}
	required := make([]string, 0, len(items))
	for _, item := range items {
		name, ok := item.(string)
		if !ok {
			return nil, inputSchemaErrorf("required must be an array of strings")
		}
		required = append(required, name)
	}
	return required, nil
}

func schemaProperties(schema map[string]any) (map[string]any, error) {
	value, ok := schema["properties"]
	if !ok || value == nil {
		return nil, nil
	}
	properties, ok := value.(map[string]any)
	if !ok {
		return nil, inputSchemaErrorf("properties must be an object")
	}
	return properties, nil
}

// schemaMap handles both object schemas and the boolean-schema form. A true
// schema accepts any non-null JSON value; a false schema accepts no value.
func schemaMap(value any, location string) (schema map[string]any, allowed bool, err error) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true, nil
	case bool:
		return map[string]any{}, typed, nil
	default:
		return nil, false, inputSchemaErrorf("schema for %s must be an object or boolean", location)
	}
}

func validateValue(value any, schema map[string]any, path string) error {
	types, hasType, err := schemaTypes(schema)
	if err != nil {
		return err
	}

	if value == nil {
		if !explicitlyAllowsNull(schema, types, hasType) {
			return fmt.Errorf("%s must not be null", path)
		}
		return validateEnum(value, schema, path)
	}

	if hasType && !matchesAnyJSONType(value, types) {
		return fmt.Errorf("%s must be %s, got %s", path, strings.Join(types, " or "), jsonTypeName(value))
	}
	if err := validateEnum(value, schema, path); err != nil {
		return err
	}

	switch typed := value.(type) {
	case float64:
		if err := validateNumberRange(typed, schema, path); err != nil {
			return err
		}
	case string:
		if err := validateStringRange(typed, schema, path); err != nil {
			return err
		}
	case []any:
		if err := validateArray(typed, schema, path); err != nil {
			return err
		}
	case map[string]any:
		if err := validateObject(typed, schema, path); err != nil {
			return err
		}
	}

	return nil
}

func schemaTypes(schema map[string]any) ([]string, bool, error) {
	value, ok := schema["type"]
	if !ok {
		return nil, false, nil
	}
	switch typed := value.(type) {
	case string:
		return []string{typed}, true, nil
	case []any:
		types := make([]string, 0, len(typed))
		for _, item := range typed {
			typeName, ok := item.(string)
			if !ok {
				return nil, false, inputSchemaErrorf("type array must contain only strings")
			}
			types = append(types, typeName)
		}
		if len(types) == 0 {
			return nil, false, inputSchemaErrorf("type array must not be empty")
		}
		return types, true, nil
	default:
		return nil, false, inputSchemaErrorf("type must be a string or an array of strings")
	}
}

func explicitlyAllowsNull(schema map[string]any, types []string, hasType bool) bool {
	if hasType {
		for _, typeName := range types {
			if typeName == "null" {
				return true
			}
		}
		return false
	}

	values, ok := schema["enum"].([]any)
	if !ok {
		return false
	}
	for _, value := range values {
		if value == nil {
			return true
		}
	}
	return false
}

func matchesAnyJSONType(value any, types []string) bool {
	for _, typeName := range types {
		if matchesJSONType(value, typeName) {
			return true
		}
	}
	return false
}

func matchesJSONType(value any, typeName string) bool {
	switch typeName {
	case "null":
		return value == nil
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && math.Trunc(number) == number
	case "string":
		_, ok := value.(string)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	default:
		return false
	}
}

func jsonTypeName(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("unsupported JSON value %T", value)
	}
}

func validateEnum(value any, schema map[string]any, path string) error {
	rawValues, ok := schema["enum"]
	if !ok {
		return nil
	}
	values, ok := rawValues.([]any)
	if !ok || len(values) == 0 {
		return inputSchemaErrorf("enum for %s must be a non-empty array", path)
	}
	for _, allowed := range values {
		if reflect.DeepEqual(value, allowed) {
			return nil
		}
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return inputSchemaErrorf("encode enum for %s: %v", path, err)
	}
	return fmt.Errorf("%s must be one of %s", path, encoded)
}

func validateNumberRange(value float64, schema map[string]any, path string) error {
	minimum, present, err := numericConstraint(schema, "minimum", path)
	if err != nil {
		return err
	}
	if present && value < minimum {
		return fmt.Errorf("%s must be >= %v", path, minimum)
	}
	maximum, present, err := numericConstraint(schema, "maximum", path)
	if err != nil {
		return err
	}
	if present && value > maximum {
		return fmt.Errorf("%s must be <= %v", path, maximum)
	}
	exclusiveMinimum, present, err := numericConstraint(schema, "exclusiveMinimum", path)
	if err != nil {
		return err
	}
	if present && value <= exclusiveMinimum {
		return fmt.Errorf("%s must be > %v", path, exclusiveMinimum)
	}
	exclusiveMaximum, present, err := numericConstraint(schema, "exclusiveMaximum", path)
	if err != nil {
		return err
	}
	if present && value >= exclusiveMaximum {
		return fmt.Errorf("%s must be < %v", path, exclusiveMaximum)
	}
	return nil
}

func numericConstraint(schema map[string]any, keyword, path string) (float64, bool, error) {
	value, ok := schema[keyword]
	if !ok {
		return 0, false, nil
	}
	number, ok := value.(float64)
	if !ok || math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false, inputSchemaErrorf("%s for %s must be a finite number", keyword, path)
	}
	return number, true, nil
}

func validateStringRange(value string, schema map[string]any, path string) error {
	length := utf8.RuneCountInString(value)
	minimum, present, err := nonNegativeIntegerConstraint(schema, "minLength", path)
	if err != nil {
		return err
	}
	if present && length < minimum {
		return fmt.Errorf("%s length must be >= %d", path, minimum)
	}
	maximum, present, err := nonNegativeIntegerConstraint(schema, "maxLength", path)
	if err != nil {
		return err
	}
	if present && length > maximum {
		return fmt.Errorf("%s length must be <= %d", path, maximum)
	}
	return nil
}

func validateArray(value []any, schema map[string]any, path string) error {
	minimum, present, err := nonNegativeIntegerConstraint(schema, "minItems", path)
	if err != nil {
		return err
	}
	if present && len(value) < minimum {
		return fmt.Errorf("%s must contain at least %d items", path, minimum)
	}
	maximum, present, err := nonNegativeIntegerConstraint(schema, "maxItems", path)
	if err != nil {
		return err
	}
	if present && len(value) > maximum {
		return fmt.Errorf("%s must contain at most %d items", path, maximum)
	}

	rawItems, ok := schema["items"]
	if !ok {
		return nil
	}
	itemSchema, allowed, err := schemaMap(rawItems, fmt.Sprintf("items for %s", path))
	if err != nil {
		return err
	}
	for index, item := range value {
		itemPath := fmt.Sprintf("%s[%d]", path, index)
		if !allowed {
			return fmt.Errorf("%s is not allowed", itemPath)
		}
		if err := validateValue(item, itemSchema, itemPath); err != nil {
			return err
		}
	}
	return nil
}

func nonNegativeIntegerConstraint(schema map[string]any, keyword, path string) (int, bool, error) {
	number, present, err := numericConstraint(schema, keyword, path)
	if err != nil || !present {
		return 0, present, err
	}
	if number < 0 || math.Trunc(number) != number || number > float64(math.MaxInt) {
		return 0, false, inputSchemaErrorf("%s for %s must be a non-negative integer", keyword, path)
	}
	return int(number), true, nil
}

func joinPropertyPath(parent, name string) string {
	if parent == "" {
		return fmt.Sprintf("property %q", name)
	}
	return fmt.Sprintf("%s.%s", parent, name)
}

func validatedPromptHandler(
	prompt gomcp.Prompt,
	next server.PromptHandlerFunc,
) server.PromptHandlerFunc {
	return func(ctx context.Context, request gomcp.GetPromptRequest) (*gomcp.GetPromptResult, error) {
		missing := make([]string, 0, len(prompt.Arguments))
		for _, argument := range prompt.Arguments {
			if !argument.Required {
				continue
			}
			if _, ok := request.Params.Arguments[argument.Name]; !ok {
				missing = append(missing, argument.Name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			return nil, fmt.Errorf(
				"invalid arguments for prompt %q: missing required arguments: %s",
				prompt.Name,
				strings.Join(missing, ", "),
			)
		}
		return next(ctx, request)
	}
}

func addValidatedPrompt(
	s *server.MCPServer,
	prompt gomcp.Prompt,
	handler server.PromptHandlerFunc,
) {
	s.AddPrompt(prompt, validatedPromptHandler(prompt, handler))
}
