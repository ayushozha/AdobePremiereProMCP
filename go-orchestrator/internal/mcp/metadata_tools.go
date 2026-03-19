package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerMetadataTools registers clip metadata, labels, footage interpretation,
// media management, smart bins, clip usage, and file management MCP tools.
func registerMetadataTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Clip/Item Metadata (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_get_clip_metadata
	s.AddTool(
		gomcp.NewTool("premiere_get_clip_metadata",
			gomcp.WithDescription("Get all metadata (XMP + project metadata) for a project item by index."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetClipMetadataHandler(orch, logger),
	)

	// 2. premiere_set_clip_metadata
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_metadata",
			gomcp.WithDescription("Set a metadata field on a project item."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithString("field",
				gomcp.Required(),
				gomcp.Description("Metadata field name to set (e.g. 'dc:description')"),
			),
			gomcp.WithString("value",
				gomcp.Required(),
				gomcp.Description("Value to set for the metadata field"),
			),
		),
		makeSetClipMetadataHandler(orch, logger),
	)

	// 3. premiere_add_custom_metadata_field
	s.AddTool(
		gomcp.NewTool("premiere_add_custom_metadata_field",
			gomcp.WithDescription("Add a custom metadata schema field to the project."),
			gomcp.WithString("field_name",
				gomcp.Required(),
				gomcp.Description("Internal name for the metadata field"),
			),
			gomcp.WithString("field_label",
				gomcp.Description("Display label for the field (defaults to field_name)"),
			),
			gomcp.WithNumber("field_type",
				gomcp.Description("Field type: 0=string, 1=integer, 2=real (default: 0)"),
			),
		),
		makeAddCustomMetadataFieldHandler(orch, logger),
	)

	// 4. premiere_get_metadata_schema
	s.AddTool(
		gomcp.NewTool("premiere_get_metadata_schema",
			gomcp.WithDescription("Get available metadata fields from the project metadata schema."),
		),
		makeGetMetadataSchemaHandler(orch, logger),
	)

	// 5. premiere_batch_set_metadata
	s.AddTool(
		gomcp.NewTool("premiere_batch_set_metadata",
			gomcp.WithDescription("Set a metadata field on multiple project items at once."),
			gomcp.WithString("item_indices",
				gomcp.Required(),
				gomcp.Description("Comma-separated list of zero-based project item indices (e.g. '0,1,3,5')"),
			),
			gomcp.WithString("field",
				gomcp.Required(),
				gomcp.Description("Metadata field name to set"),
			),
			gomcp.WithString("value",
				gomcp.Required(),
				gomcp.Description("Value to set for the metadata field"),
			),
		),
		makeBatchSetMetadataHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Labels & Colors (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_get_available_label_colors
	s.AddTool(
		gomcp.NewTool("premiere_get_available_label_colors",
			gomcp.WithDescription("Get all available label colors with their indices (0-15) and names."),
		),
		makeGetAvailableLabelColorsHandler(orch, logger),
	)

	// 7. premiere_set_clip_label_by_name
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_label_by_name",
			gomcp.WithDescription("Set a label color on a project item by color name (e.g. 'Violet', 'Iris', 'Caribbean')."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithString("color_name",
				gomcp.Required(),
				gomcp.Description("Label color name: Violet, Iris, Caribbean, Lavender, Cerulean, Forest, Rose, Mango, Purple, Blue, Teal, Magenta, Tan, Green, Brown, Yellow"),
			),
		),
		makeSetClipLabelByNameHandler(orch, logger),
	)

	// 8. premiere_get_label_color_for_clip
	s.AddTool(
		gomcp.NewTool("premiere_get_label_color_for_clip",
			gomcp.WithDescription("Get the label color index and name for a project item."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetLabelColorForClipHandler(orch, logger),
	)

	// 9. premiere_batch_set_labels
	s.AddTool(
		gomcp.NewTool("premiere_batch_set_labels",
			gomcp.WithDescription("Set a label color on multiple project items at once."),
			gomcp.WithString("item_indices",
				gomcp.Required(),
				gomcp.Description("Comma-separated list of zero-based project item indices (e.g. '0,1,3,5')"),
			),
			gomcp.WithNumber("color_index",
				gomcp.Required(),
				gomcp.Description("Label color index (0-15)"),
			),
		),
		makeBatchSetLabelsHandler(orch, logger),
	)

	// 10. premiere_filter_by_label
	s.AddTool(
		gomcp.NewTool("premiere_filter_by_label",
			gomcp.WithDescription("Get all project items that have a specific label color."),
			gomcp.WithNumber("color_index",
				gomcp.Required(),
				gomcp.Description("Label color index (0-15) to filter by"),
			),
		),
		makeFilterByLabelHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Footage Interpretation (11-16)
	// -----------------------------------------------------------------------

	// 11. premiere_get_footage_interpretation
	s.AddTool(
		gomcp.NewTool("premiere_get_footage_interpretation",
			gomcp.WithDescription("Get footage interpretation settings (fps, fields, alpha, pixel aspect ratio) for a project item."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetFootageInterpretationHandler(orch, logger),
	)

	// 12. premiere_set_footage_frame_rate
	s.AddTool(
		gomcp.NewTool("premiere_set_footage_frame_rate",
			gomcp.WithDescription("Override the frame rate on a project item's footage."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithNumber("fps",
				gomcp.Required(),
				gomcp.Description("Frame rate in frames per second (e.g. 23.976, 24, 29.97, 30, 60)"),
			),
		),
		makeSetFootageFrameRateHandler(orch, logger),
	)

	// 13. premiere_set_footage_field_order
	s.AddTool(
		gomcp.NewTool("premiere_set_footage_field_order",
			gomcp.WithDescription("Set the field order on a project item's footage."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithNumber("field_order",
				gomcp.Required(),
				gomcp.Description("Field order: 0=progressive, 1=upper field first, 2=lower field first"),
			),
		),
		makeSetFootageFieldOrderHandler(orch, logger),
	)

	// 14. premiere_set_footage_alpha_channel
	s.AddTool(
		gomcp.NewTool("premiere_set_footage_alpha_channel",
			gomcp.WithDescription("Set the alpha channel interpretation on a project item's footage."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithNumber("alpha_type",
				gomcp.Required(),
				gomcp.Description("Alpha type: 0=none/ignore, 1=straight/unmatted, 2=premultiplied/matted"),
			),
		),
		makeSetFootageAlphaChannelHandler(orch, logger),
	)

	// 15. premiere_set_footage_pixel_aspect_ratio
	s.AddTool(
		gomcp.NewTool("premiere_set_footage_pixel_aspect_ratio",
			gomcp.WithDescription("Set the pixel aspect ratio on a project item's footage."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithNumber("numerator",
				gomcp.Required(),
				gomcp.Description("Pixel aspect ratio numerator (e.g. 1 for square pixels)"),
			),
			gomcp.WithNumber("denominator",
				gomcp.Required(),
				gomcp.Description("Pixel aspect ratio denominator (e.g. 1 for square pixels)"),
			),
		),
		makeSetFootagePixelAspectRatioHandler(orch, logger),
	)

	// 16. premiere_reset_footage_interpretation
	s.AddTool(
		gomcp.NewTool("premiere_reset_footage_interpretation",
			gomcp.WithDescription("Reset footage interpretation to auto-detected settings."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeResetFootageInterpretationHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Media Info (17-22)
	// -----------------------------------------------------------------------

	// 17. premiere_get_media_info
	s.AddTool(
		gomcp.NewTool("premiere_get_media_info",
			gomcp.WithDescription("Get full media info for a project item including codec, resolution, fps, duration, audio channels, and file size."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetMediaInfoHandler(orch, logger),
	)

	// 18. premiere_get_media_path
	s.AddTool(
		gomcp.NewTool("premiere_get_media_path",
			gomcp.WithDescription("Get the file path for a project item's media on disk."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetMediaPathHandler(orch, logger),
	)

	// 19. premiere_reveal_in_finder
	s.AddTool(
		gomcp.NewTool("premiere_reveal_in_finder",
			gomcp.WithDescription("Reveal a project item's media file in Finder (macOS) or Explorer (Windows)."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeRevealInFinderHandler(orch, logger),
	)

	// 20. premiere_refresh_media
	s.AddTool(
		gomcp.NewTool("premiere_refresh_media",
			gomcp.WithDescription("Force refresh media for a project item, re-reading the source file from disk."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeRefreshMediaHandler(orch, logger),
	)

	// 21. premiere_replace_media
	s.AddTool(
		gomcp.NewTool("premiere_replace_media",
			gomcp.WithDescription("Replace a project item's media with a different file on disk."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
			gomcp.WithString("new_file_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the replacement media file"),
			),
		),
		makeReplaceMediaHandler(orch, logger),
	)

	// 22. premiere_duplicate_project_item
	s.AddTool(
		gomcp.NewTool("premiere_duplicate_project_item",
			gomcp.WithDescription("Duplicate a project item in the project panel."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item to duplicate"),
			),
		),
		makeDuplicateProjectItemHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Smart Bins (23-24)
	// -----------------------------------------------------------------------

	// 23. premiere_create_smart_bin
	s.AddTool(
		gomcp.NewTool("premiere_create_smart_bin",
			gomcp.WithDescription("Create a smart bin with search criteria. Smart bins automatically show items matching the query."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Name for the smart bin"),
			),
			gomcp.WithString("search_query",
				gomcp.Required(),
				gomcp.Description("Search query string for the smart bin criteria"),
			),
		),
		makeCreateSmartBinHandler(orch, logger),
	)

	// 24. premiere_get_smart_bin_results
	s.AddTool(
		gomcp.NewTool("premiere_get_smart_bin_results",
			gomcp.WithDescription("Get items matching a smart bin's criteria by navigating to the bin path."),
			gomcp.WithString("bin_path",
				gomcp.Required(),
				gomcp.Description("Slash-separated path to the smart bin (e.g. 'SmartBinName' or 'Parent/SmartBinName')"),
			),
		),
		makeGetSmartBinResultsHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// Clip Usage (25-28)
	// -----------------------------------------------------------------------

	// 25. premiere_get_clip_usage_in_sequences
	s.AddTool(
		gomcp.NewTool("premiere_get_clip_usage_in_sequences",
			gomcp.WithDescription("Find all sequences where a project item (clip) is used, with usage count per sequence."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetClipUsageInSequencesHandler(orch, logger),
	)

	// 26. premiere_get_unused_clips
	s.AddTool(
		gomcp.NewTool("premiere_get_unused_clips",
			gomcp.WithDescription("List all clips in the project that are not used in any sequence."),
		),
		makeGetUnusedClipsHandler(orch, logger),
	)

	// 27. premiere_get_used_clips
	s.AddTool(
		gomcp.NewTool("premiere_get_used_clips",
			gomcp.WithDescription("List all clips in the project that are used in at least one sequence."),
		),
		makeGetUsedClipsHandler(orch, logger),
	)

	// 28. premiere_get_clip_usage_count
	s.AddTool(
		gomcp.NewTool("premiere_get_clip_usage_count",
			gomcp.WithDescription("Count how many times a specific clip is used across all sequences (including multiple uses in the same sequence)."),
			gomcp.WithNumber("project_item_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the project item in the root bin"),
			),
		),
		makeGetClipUsageCountHandler(orch, logger),
	)

	// -----------------------------------------------------------------------
	// File Management (29-30)
	// -----------------------------------------------------------------------

	// 29. premiere_get_project_file_size
	s.AddTool(
		gomcp.NewTool("premiere_get_project_file_size",
			gomcp.WithDescription("Get the .prproj project file size on disk."),
		),
		makeGetProjectFileSizeHandler(orch, logger),
	)

	// 30. premiere_get_media_disk_usage
	s.AddTool(
		gomcp.NewTool("premiere_get_media_disk_usage",
			gomcp.WithDescription("Calculate total disk usage of all media files referenced by the project."),
		),
		makeGetMediaDiskUsageHandler(orch, logger),
	)
}

// ---------------------------------------------------------------------------
// Handler constructors — Clip/Item Metadata
// ---------------------------------------------------------------------------

func makeGetClipMetadataHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_clip_metadata")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetClipMetadata(ctx, piIndex)
		if err != nil {
			logger.Error("get clip metadata failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get clip metadata: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetClipMetadataHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_metadata")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		field := gomcp.ParseString(req, "field", "")
		if field == "" {
			return gomcp.NewToolResultError("parameter 'field' is required"), nil
		}
		value := gomcp.ParseString(req, "value", "")
		if value == "" {
			return gomcp.NewToolResultError("parameter 'value' is required"), nil
		}
		result, err := orch.SetClipMetadata(ctx, piIndex, field, value)
		if err != nil {
			logger.Error("set clip metadata failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip metadata: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeAddCustomMetadataFieldHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_add_custom_metadata_field")
		fieldName := gomcp.ParseString(req, "field_name", "")
		if fieldName == "" {
			return gomcp.NewToolResultError("parameter 'field_name' is required"), nil
		}
		fieldLabel := gomcp.ParseString(req, "field_label", "")
		fieldType := gomcp.ParseInt(req, "field_type", 0)
		result, err := orch.AddCustomMetadataField(ctx, fieldName, fieldLabel, fieldType)
		if err != nil {
			logger.Error("add custom metadata field failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to add custom metadata field: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetMetadataSchemaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_metadata_schema")
		result, err := orch.GetMetadataSchema(ctx)
		if err != nil {
			logger.Error("get metadata schema failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get metadata schema: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeBatchSetMetadataHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_batch_set_metadata")
		indicesStr := gomcp.ParseString(req, "item_indices", "")
		if indicesStr == "" {
			return gomcp.NewToolResultError("parameter 'item_indices' is required"), nil
		}
		field := gomcp.ParseString(req, "field", "")
		if field == "" {
			return gomcp.NewToolResultError("parameter 'field' is required"), nil
		}
		value := gomcp.ParseString(req, "value", "")
		if value == "" {
			return gomcp.NewToolResultError("parameter 'value' is required"), nil
		}
		indices, err := parseIntList(indicesStr)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("invalid 'item_indices': %v", err)), nil
		}
		result, orchErr := orch.BatchSetMetadata(ctx, indices, field, value)
		if orchErr != nil {
			logger.Error("batch set metadata failed", zap.Error(orchErr))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to batch set metadata: %v", orchErr)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Labels & Colors
// ---------------------------------------------------------------------------

func makeGetAvailableLabelColorsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_available_label_colors")
		result, err := orch.GetAvailableLabelColors(ctx)
		if err != nil {
			logger.Error("get available label colors failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get available label colors: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetClipLabelByNameHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_clip_label_by_name")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		colorName := gomcp.ParseString(req, "color_name", "")
		if colorName == "" {
			return gomcp.NewToolResultError("parameter 'color_name' is required"), nil
		}
		result, err := orch.SetClipLabelByName(ctx, piIndex, colorName)
		if err != nil {
			logger.Error("set clip label by name failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set clip label: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetLabelColorForClipHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_label_color_for_clip")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetLabelColorForClip(ctx, piIndex)
		if err != nil {
			logger.Error("get label color for clip failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get label color: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeBatchSetLabelsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_batch_set_labels")
		indicesStr := gomcp.ParseString(req, "item_indices", "")
		if indicesStr == "" {
			return gomcp.NewToolResultError("parameter 'item_indices' is required"), nil
		}
		colorIndex := gomcp.ParseInt(req, "color_index", -1)
		if colorIndex < 0 || colorIndex > 15 {
			return gomcp.NewToolResultError("parameter 'color_index' must be between 0 and 15"), nil
		}
		indices, err := parseIntList(indicesStr)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("invalid 'item_indices': %v", err)), nil
		}
		result, orchErr := orch.BatchSetLabels(ctx, indices, colorIndex)
		if orchErr != nil {
			logger.Error("batch set labels failed", zap.Error(orchErr))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to batch set labels: %v", orchErr)), nil
		}
		return toolResultJSON(result)
	}
}

func makeFilterByLabelHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_filter_by_label")
		colorIndex := gomcp.ParseInt(req, "color_index", -1)
		if colorIndex < 0 || colorIndex > 15 {
			return gomcp.NewToolResultError("parameter 'color_index' must be between 0 and 15"), nil
		}
		result, err := orch.FilterByLabel(ctx, colorIndex)
		if err != nil {
			logger.Error("filter by label failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to filter by label: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Footage Interpretation
// ---------------------------------------------------------------------------

func makeGetFootageInterpretationHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_footage_interpretation")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetFootageInterpretation(ctx, piIndex)
		if err != nil {
			logger.Error("get footage interpretation failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get footage interpretation: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetFootageFrameRateHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_footage_frame_rate")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		fps := gomcp.ParseFloat64(req, "fps", 0)
		if fps <= 0 {
			return gomcp.NewToolResultError("parameter 'fps' must be a positive number"), nil
		}
		result, err := orch.SetFootageFrameRate(ctx, piIndex, fps)
		if err != nil {
			logger.Error("set footage frame rate failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set footage frame rate: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetFootageFieldOrderHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_footage_field_order")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		fieldOrder := gomcp.ParseInt(req, "field_order", -1)
		if fieldOrder < 0 || fieldOrder > 2 {
			return gomcp.NewToolResultError("parameter 'field_order' must be 0 (progressive), 1 (upper), or 2 (lower)"), nil
		}
		result, err := orch.SetFootageFieldOrder(ctx, piIndex, fieldOrder)
		if err != nil {
			logger.Error("set footage field order failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set footage field order: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetFootageAlphaChannelHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_footage_alpha_channel")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		alphaType := gomcp.ParseInt(req, "alpha_type", -1)
		if alphaType < 0 || alphaType > 2 {
			return gomcp.NewToolResultError("parameter 'alpha_type' must be 0 (none), 1 (straight), or 2 (premultiplied)"), nil
		}
		result, err := orch.SetFootageAlphaChannel(ctx, piIndex, alphaType)
		if err != nil {
			logger.Error("set footage alpha channel failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set footage alpha channel: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeSetFootagePixelAspectRatioHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_set_footage_pixel_aspect_ratio")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		num := gomcp.ParseFloat64(req, "numerator", 0)
		den := gomcp.ParseFloat64(req, "denominator", 0)
		if num <= 0 || den <= 0 {
			return gomcp.NewToolResultError("parameters 'numerator' and 'denominator' must be positive numbers"), nil
		}
		result, err := orch.SetFootagePixelAspectRatio(ctx, piIndex, num, den)
		if err != nil {
			logger.Error("set footage pixel aspect ratio failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to set footage pixel aspect ratio: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeResetFootageInterpretationHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_reset_footage_interpretation")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.ResetFootageInterpretation(ctx, piIndex)
		if err != nil {
			logger.Error("reset footage interpretation failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to reset footage interpretation: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Media Info
// ---------------------------------------------------------------------------

func makeGetMediaInfoHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_media_info")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetMediaInfo(ctx, piIndex)
		if err != nil {
			logger.Error("get media info failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get media info: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetMediaPathHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_media_path")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetMediaPath(ctx, piIndex)
		if err != nil {
			logger.Error("get media path failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get media path: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRevealInFinderHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_reveal_in_finder")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.RevealInFinder(ctx, piIndex)
		if err != nil {
			logger.Error("reveal in finder failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to reveal in finder: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeRefreshMediaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_refresh_media")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.RefreshMedia(ctx, piIndex)
		if err != nil {
			logger.Error("refresh media failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to refresh media: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeReplaceMediaHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_replace_media")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		newFilePath := gomcp.ParseString(req, "new_file_path", "")
		if newFilePath == "" {
			return gomcp.NewToolResultError("parameter 'new_file_path' is required"), nil
		}
		result, err := orch.ReplaceMedia(ctx, piIndex, newFilePath)
		if err != nil {
			logger.Error("replace media failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to replace media: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeDuplicateProjectItemHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_duplicate_project_item")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.DuplicateProjectItem(ctx, piIndex)
		if err != nil {
			logger.Error("duplicate project item failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to duplicate project item: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Smart Bins
// ---------------------------------------------------------------------------

func makeCreateSmartBinHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_create_smart_bin")
		name := gomcp.ParseString(req, "name", "")
		if name == "" {
			return gomcp.NewToolResultError("parameter 'name' is required"), nil
		}
		searchQuery := gomcp.ParseString(req, "search_query", "")
		if searchQuery == "" {
			return gomcp.NewToolResultError("parameter 'search_query' is required"), nil
		}
		result, err := orch.CreateSmartBin(ctx, name, searchQuery)
		if err != nil {
			logger.Error("create smart bin failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to create smart bin: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetSmartBinResultsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_smart_bin_results")
		binPath := gomcp.ParseString(req, "bin_path", "")
		if binPath == "" {
			return gomcp.NewToolResultError("parameter 'bin_path' is required"), nil
		}
		result, err := orch.GetSmartBinResults(ctx, binPath)
		if err != nil {
			logger.Error("get smart bin results failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get smart bin results: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — Clip Usage
// ---------------------------------------------------------------------------

func makeGetClipUsageInSequencesHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_clip_usage_in_sequences")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetClipUsageInSequences(ctx, piIndex)
		if err != nil {
			logger.Error("get clip usage in sequences failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get clip usage: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetUnusedClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_unused_clips")
		result, err := orch.GetUnusedClips(ctx)
		if err != nil {
			logger.Error("get unused clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get unused clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetUsedClipsHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_used_clips")
		result, err := orch.GetUsedClips(ctx)
		if err != nil {
			logger.Error("get used clips failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get used clips: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetClipUsageCountHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_clip_usage_count")
		piIndex := gomcp.ParseInt(req, "project_item_index", -1)
		if piIndex < 0 {
			return gomcp.NewToolResultError("parameter 'project_item_index' is required"), nil
		}
		result, err := orch.GetClipUsageCount(ctx, piIndex)
		if err != nil {
			logger.Error("get clip usage count failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get clip usage count: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Handler constructors — File Management
// ---------------------------------------------------------------------------

func makeGetProjectFileSizeHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_project_file_size")
		result, err := orch.GetProjectFileSize(ctx)
		if err != nil {
			logger.Error("get project file size failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get project file size: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

func makeGetMediaDiskUsageHandler(orch Orchestrator, logger *zap.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_get_media_disk_usage")
		result, err := orch.GetMediaDiskUsage(ctx)
		if err != nil {
			logger.Error("get media disk usage failed", zap.Error(err))
			return gomcp.NewToolResultError(fmt.Sprintf("failed to get media disk usage: %v", err)), nil
		}
		return toolResultJSON(result)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseIntList splits a comma-separated string of integers into a slice.
func parseIntList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q: %w", p, err)
		}
		result = append(result, v)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no valid indices provided")
	}
	return result, nil
}
