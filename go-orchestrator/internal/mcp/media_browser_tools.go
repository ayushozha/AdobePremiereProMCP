package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// mbH is a small handler wrapper for media browser tools.
func mbH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerMediaBrowserTools registers all 15 media browser and Adobe Stock
// MCP tools for browsing the filesystem, finding media files, and searching
// Adobe Stock for video, audio, and template assets.
func registerMediaBrowserTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Media Browser (1-7)
	// -----------------------------------------------------------------------

	// 1. premiere_browse_path
	s.AddTool(gomcp.NewTool("premiere_browse_path",
		gomcp.WithDescription("Browse any filesystem path and list all files and folders. Returns name, path, size, and modification date for each item."),
		gomcp.WithString("path", gomcp.Required(), gomcp.Description("Absolute filesystem path to browse")),
	), mbH(orch, logger, "browse_path", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.BrowsePath(ctx, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_browse_media_files
	s.AddTool(gomcp.NewTool("premiere_browse_media_files",
		gomcp.WithDescription("Browse a path showing only media files (video, audio, images, project files). Filters out non-media files automatically."),
		gomcp.WithString("path", gomcp.Required(), gomcp.Description("Absolute filesystem path to browse for media files")),
	), mbH(orch, logger, "browse_media_files", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		path := gomcp.ParseString(req, "path", "")
		if path == "" {
			return gomcp.NewToolResultError("parameter 'path' is required"), nil
		}
		result, err := orch.BrowseMediaFiles(ctx, path)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_get_favorite_locations
	s.AddTool(gomcp.NewTool("premiere_get_favorite_locations",
		gomcp.WithDescription("Get common filesystem locations for quick navigation: Home, Documents, Desktop, Movies, Downloads, and Premiere Pro projects folder."),
	), mbH(orch, logger, "get_favorite_locations", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetFavoriteLocations(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_import_from_media_browser
	s.AddTool(gomcp.NewTool("premiere_import_from_media_browser",
		gomcp.WithDescription("Import one or more files into the Premiere Pro project from the media browser. Supports batch import of multiple files."),
		gomcp.WithArray("paths", gomcp.Required(), gomcp.Description("Array of absolute file paths to import into the project"), gomcp.WithStringItems()),
	), mbH(orch, logger, "import_from_media_browser", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		paths, err := extractStringSlice(req, "paths")
		if err != nil || len(paths) == 0 {
			return gomcp.NewToolResultError("parameter 'paths' is required and must contain at least one file path"), nil
		}
		result, err := orch.ImportFromMediaBrowser(ctx, paths)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_get_recent_locations
	s.AddTool(gomcp.NewTool("premiere_get_recent_locations",
		gomcp.WithDescription("Get Premiere Pro project locations and version-specific folders. Useful for finding recent projects."),
	), mbH(orch, logger, "get_recent_locations", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetRecentLocations(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_browse_local_drives
	s.AddTool(gomcp.NewTool("premiere_browse_local_drives",
		gomcp.WithDescription("List available local drives and volumes. On macOS lists /Volumes, on Windows lists drive letters."),
	), mbH(orch, logger, "browse_local_drives", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.BrowseLocalDrives(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_browse_creative_cloud
	s.AddTool(gomcp.NewTool("premiere_browse_creative_cloud",
		gomcp.WithDescription("Browse Creative Cloud Files sync location. Shows the local path where Creative Cloud syncs files."),
	), mbH(orch, logger, "browse_creative_cloud", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.BrowseCreativeCloud(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Adobe Stock Search (8-15)
	// -----------------------------------------------------------------------

	// 8. premiere_search_stock
	s.AddTool(gomcp.NewTool("premiere_search_stock",
		gomcp.WithDescription("Search Adobe Stock for video, audio, templates, or images. Returns search URL and result metadata."),
		gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query string (e.g. 'sunset timelapse', 'corporate music')")),
		gomcp.WithString("media_type", gomcp.Description("Type of media to search for"),
			gomcp.Enum("video", "audio", "template", "image")),
		gomcp.WithNumber("limit", gomcp.Description("Maximum number of results to return (default: 20)")),
	), mbH(orch, logger, "search_stock", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := gomcp.ParseString(req, "query", "")
		if query == "" {
			return gomcp.NewToolResultError("parameter 'query' is required"), nil
		}
		mediaType := gomcp.ParseString(req, "media_type", "video")
		limit := gomcp.ParseInt(req, "limit", 20)
		result, err := orch.SearchStock(ctx, query, mediaType, limit)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_get_stock_categories
	s.AddTool(gomcp.NewTool("premiere_get_stock_categories",
		gomcp.WithDescription("Get available Adobe Stock categories for a given media type. Useful for discovering what types of stock assets are available."),
		gomcp.WithString("media_type", gomcp.Description("Media type to get categories for"),
			gomcp.Enum("video", "audio", "template", "image")),
	), mbH(orch, logger, "get_stock_categories", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		mediaType := gomcp.ParseString(req, "media_type", "video")
		result, err := orch.GetStockCategories(ctx, mediaType)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_search_stock_video
	s.AddTool(gomcp.NewTool("premiere_search_stock_video",
		gomcp.WithDescription("Search Adobe Stock specifically for video clips. Shortcut for premiere_search_stock with media_type=video."),
		gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query for stock videos (e.g. 'drone aerial cityscape', 'slow motion water')")),
		gomcp.WithNumber("limit", gomcp.Description("Maximum number of results (default: 20)")),
	), mbH(orch, logger, "search_stock_video", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := gomcp.ParseString(req, "query", "")
		if query == "" {
			return gomcp.NewToolResultError("parameter 'query' is required"), nil
		}
		limit := gomcp.ParseInt(req, "limit", 20)
		result, err := orch.SearchStockVideo(ctx, query, limit)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_search_stock_audio
	s.AddTool(gomcp.NewTool("premiere_search_stock_audio",
		gomcp.WithDescription("Search Adobe Stock specifically for audio and music. Shortcut for premiere_search_stock with media_type=audio."),
		gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query for stock audio (e.g. 'upbeat corporate', 'ambient nature sounds')")),
		gomcp.WithNumber("limit", gomcp.Description("Maximum number of results (default: 20)")),
	), mbH(orch, logger, "search_stock_audio", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := gomcp.ParseString(req, "query", "")
		if query == "" {
			return gomcp.NewToolResultError("parameter 'query' is required"), nil
		}
		limit := gomcp.ParseInt(req, "limit", 20)
		result, err := orch.SearchStockAudio(ctx, query, limit)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_search_stock_templates
	s.AddTool(gomcp.NewTool("premiere_search_stock_templates",
		gomcp.WithDescription("Search Adobe Stock for Motion Graphics templates (MOGRTs), title templates, and transition templates."),
		gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query for stock templates (e.g. 'lower third', 'title animation')")),
		gomcp.WithNumber("limit", gomcp.Description("Maximum number of results (default: 20)")),
	), mbH(orch, logger, "search_stock_templates", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := gomcp.ParseString(req, "query", "")
		if query == "" {
			return gomcp.NewToolResultError("parameter 'query' is required"), nil
		}
		limit := gomcp.ParseInt(req, "limit", 20)
		result, err := orch.SearchStockTemplates(ctx, query, limit)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_get_stock_preview
	s.AddTool(gomcp.NewTool("premiere_get_stock_preview",
		gomcp.WithDescription("Get the preview URL for an Adobe Stock item by its ID. Returns a link to preview the asset on stock.adobe.com."),
		gomcp.WithNumber("stock_id", gomcp.Required(), gomcp.Description("Adobe Stock item ID")),
	), mbH(orch, logger, "get_stock_preview", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		stockID := gomcp.ParseInt(req, "stock_id", 0)
		if stockID <= 0 {
			return gomcp.NewToolResultError("parameter 'stock_id' must be a positive integer"), nil
		}
		result, err := orch.GetStockPreview(ctx, stockID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_download_stock
	s.AddTool(gomcp.NewTool("premiere_download_stock",
		gomcp.WithDescription("Download and license an Adobe Stock item. Requires Adobe Creative Cloud login and valid license."),
		gomcp.WithNumber("stock_id", gomcp.Required(), gomcp.Description("Adobe Stock item ID to download")),
		gomcp.WithString("download_path", gomcp.Required(), gomcp.Description("Absolute path where the downloaded file should be saved")),
	), mbH(orch, logger, "download_stock", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		stockID := gomcp.ParseInt(req, "stock_id", 0)
		if stockID <= 0 {
			return gomcp.NewToolResultError("parameter 'stock_id' must be a positive integer"), nil
		}
		downloadPath := gomcp.ParseString(req, "download_path", "")
		if downloadPath == "" {
			return gomcp.NewToolResultError("parameter 'download_path' is required"), nil
		}
		result, err := orch.DownloadStock(ctx, stockID, downloadPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_import_stock
	s.AddTool(gomcp.NewTool("premiere_import_stock",
		gomcp.WithDescription("Download an Adobe Stock item and import it directly into the Premiere Pro project. Requires CC login for licensed assets."),
		gomcp.WithNumber("stock_id", gomcp.Required(), gomcp.Description("Adobe Stock item ID to download and import")),
		gomcp.WithString("download_path", gomcp.Required(), gomcp.Description("Absolute path where the downloaded file should be saved")),
		gomcp.WithString("target_bin", gomcp.Description("Name of the bin to import into (default: root)")),
	), mbH(orch, logger, "import_stock", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		stockID := gomcp.ParseInt(req, "stock_id", 0)
		if stockID <= 0 {
			return gomcp.NewToolResultError("parameter 'stock_id' must be a positive integer"), nil
		}
		downloadPath := gomcp.ParseString(req, "download_path", "")
		if downloadPath == "" {
			return gomcp.NewToolResultError("parameter 'download_path' is required"), nil
		}
		targetBin := gomcp.ParseString(req, "target_bin", "")
		result, err := orch.ImportStock(ctx, stockID, downloadPath, targetBin)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
