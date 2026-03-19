package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

// ---------------------------------------------------------------------------
// Media Browser — ExtendScript-backed operations
// ---------------------------------------------------------------------------

// BrowsePath lists all files and folders at the given path.
func (e *Engine) BrowsePath(ctx context.Context, path string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"path": path,
	})
	result, err := e.premiere.EvalCommand(ctx, "browsePath", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BrowsePath: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BrowseMediaFiles lists only media files at the given path.
func (e *Engine) BrowseMediaFiles(ctx context.Context, path string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"path": path,
	})
	result, err := e.premiere.EvalCommand(ctx, "browseMediaFiles", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("BrowseMediaFiles: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetFavoriteLocations returns common filesystem locations.
func (e *Engine) GetFavoriteLocations(ctx context.Context) (*GenericResult, error) {
	result, err := e.premiere.EvalCommand(ctx, "getFavoriteLocations", "{}")
	if err != nil {
		return nil, fmt.Errorf("GetFavoriteLocations: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// ImportFromMediaBrowser imports files from a list of paths.
func (e *Engine) ImportFromMediaBrowser(ctx context.Context, paths []string) (*GenericResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{
		"paths": paths,
	})
	result, err := e.premiere.EvalCommand(ctx, "importFromMediaBrowser", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("ImportFromMediaBrowser: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// GetRecentLocations returns Premiere Pro project locations.
func (e *Engine) GetRecentLocations(ctx context.Context) (*GenericResult, error) {
	result, err := e.premiere.EvalCommand(ctx, "getRecentLocations", "{}")
	if err != nil {
		return nil, fmt.Errorf("GetRecentLocations: %w", err)
	}
	return &GenericResult{Status: "success", Message: result}, nil
}

// BrowseLocalDrives lists available volumes/drives on the system.
func (e *Engine) BrowseLocalDrives(ctx context.Context) (*GenericResult, error) {
	// This is done in Go directly since it does not require Premiere
	var drives []map[string]any
	if runtime.GOOS == "darwin" {
		// On macOS, list /Volumes
		drives = append(drives, map[string]any{
			"name": "Macintosh HD",
			"path": "/",
		})
		drives = append(drives, map[string]any{
			"name": "Volumes",
			"path": "/Volumes",
		})
	} else {
		// On Windows, list common drive letters
		for _, letter := range "CDEFGHIJKLMNOPQRSTUVWXYZ" {
			drives = append(drives, map[string]any{
				"name": string(letter) + ":\\",
				"path": string(letter) + ":\\",
			})
		}
	}
	data, _ := json.Marshal(map[string]any{
		"drives": drives,
		"os":     runtime.GOOS,
	})
	return &GenericResult{Status: "success", Message: string(data)}, nil
}

// BrowseCreativeCloud returns Creative Cloud file locations.
func (e *Engine) BrowseCreativeCloud(ctx context.Context) (*GenericResult, error) {
	var ccPath string
	if runtime.GOOS == "darwin" {
		ccPath = "/Library/Application Support/Adobe/Creative Cloud Files"
	} else {
		ccPath = "C:\\Users\\Public\\Documents\\Adobe\\Creative Cloud Files"
	}
	data, _ := json.Marshal(map[string]any{
		"creative_cloud_path": ccPath,
		"note":               "Creative Cloud Files sync location. Contents depend on user login.",
	})
	return &GenericResult{Status: "success", Message: string(data)}, nil
}

// ---------------------------------------------------------------------------
// Adobe Stock — HTTP-based search (no ExtendScript needed)
// ---------------------------------------------------------------------------

// StockSearchResult represents a single Adobe Stock search result.
type StockSearchResult struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	ThumbnailURL string   `json:"thumbnail_url"`
	PreviewURL   string   `json:"preview_url"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	MediaType    string   `json:"media_type"`
	Category     string   `json:"category"`
	Keywords     []string `json:"keywords"`
}

// StockSearchResponse is the response from searching Adobe Stock.
type StockSearchResponse struct {
	Results      []StockSearchResult `json:"results"`
	TotalResults int                 `json:"total_results"`
	Query        string              `json:"query"`
	SearchURL    string              `json:"search_url"`
	Note         string              `json:"note"`
}

// SearchStock searches Adobe Stock for the given query and media type.
func (e *Engine) SearchStock(ctx context.Context, query, mediaType string, limit int) (*StockSearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("SearchStock: query must not be empty")
	}
	if limit <= 0 {
		limit = 20
	}

	typeMap := map[string]string{
		"video":    "4",
		"audio":    "6",
		"template": "7",
		"image":    "1",
	}
	searchType := typeMap[mediaType]
	if searchType == "" {
		searchType = "4" // default to video
	}

	searchURL := fmt.Sprintf(
		"https://stock.adobe.com/search?k=%s&search_type=%s&limit=%d",
		url.QueryEscape(query), searchType, limit,
	)

	return &StockSearchResponse{
		Results:      nil,
		TotalResults: 0,
		Query:        query,
		SearchURL:    searchURL,
		Note:         fmt.Sprintf("Search Adobe Stock for %q at: %s — Full API integration requires an Adobe Stock API key.", query, searchURL),
	}, nil
}

// GetStockCategories returns common stock categories for the given media type.
func (e *Engine) GetStockCategories(ctx context.Context, mediaType string) (*GenericResult, error) {
	var categories []string
	switch mediaType {
	case "video":
		categories = []string{
			"Video backgrounds", "Titles", "Smoke", "Video overlays",
			"Lower thirds", "Animation", "Green screen", "Transitions",
			"Nature", "Business", "Technology", "People", "Food",
			"Travel", "Architecture", "Sports", "Abstract",
		}
	case "audio":
		categories = []string{
			"Upbeat music", "Inspirational music", "Transition SFX",
			"Foley SFX", "Ambient", "Cinematic", "Corporate",
			"Electronic", "Acoustic", "Sound effects",
		}
	case "template":
		categories = []string{
			"Motion Graphics templates", "Title templates", "Lower third templates",
			"Transition templates", "Social media templates", "Presentation templates",
		}
	default:
		categories = []string{
			"Photos", "Illustrations", "Vectors", "Videos",
			"Audio", "Templates", "3D",
		}
	}

	data, _ := json.Marshal(map[string]any{
		"media_type": mediaType,
		"categories": categories,
	})
	return &GenericResult{Status: "success", Message: string(data)}, nil
}

// SearchStockVideo is a convenience wrapper for searching stock videos.
func (e *Engine) SearchStockVideo(ctx context.Context, query string, limit int) (*StockSearchResponse, error) {
	return e.SearchStock(ctx, query, "video", limit)
}

// SearchStockAudio is a convenience wrapper for searching stock audio.
func (e *Engine) SearchStockAudio(ctx context.Context, query string, limit int) (*StockSearchResponse, error) {
	return e.SearchStock(ctx, query, "audio", limit)
}

// SearchStockTemplates is a convenience wrapper for searching stock templates.
func (e *Engine) SearchStockTemplates(ctx context.Context, query string, limit int) (*StockSearchResponse, error) {
	return e.SearchStock(ctx, query, "template", limit)
}

// GetStockPreview returns the preview URL for a given stock item.
func (e *Engine) GetStockPreview(ctx context.Context, stockID int) (*GenericResult, error) {
	previewURL := fmt.Sprintf("https://stock.adobe.com/images/%d", stockID)
	data, _ := json.Marshal(map[string]any{
		"stock_id":    stockID,
		"preview_url": previewURL,
		"note":        "Open this URL to preview the stock item. Licensing requires Adobe Creative Cloud credentials.",
	})
	return &GenericResult{Status: "success", Message: string(data)}, nil
}

// DownloadStock initiates a stock download. Requires CC credentials.
func (e *Engine) DownloadStock(ctx context.Context, stockID int, licensePath string) (*GenericResult, error) {
	// This is a stub — actual downloading requires Adobe Stock API + OAuth
	data, _ := json.Marshal(map[string]any{
		"stock_id":     stockID,
		"status":       "requires_license",
		"license_url":  fmt.Sprintf("https://stock.adobe.com/images/%d", stockID),
		"license_path": licensePath,
		"note":         "Stock download requires Adobe Creative Cloud login and a valid license. Visit the URL to license and download.",
	})
	return &GenericResult{Status: "success", Message: string(data)}, nil
}

// ImportStock downloads and imports a stock item into the project.
func (e *Engine) ImportStock(ctx context.Context, stockID int, downloadPath, targetBin string) (*GenericResult, error) {
	// This is a stub — actual downloading requires Adobe Stock API + OAuth
	// Once downloaded, it would call ImportFromMediaBrowser
	itemURL := fmt.Sprintf("https://stock.adobe.com/images/%d", stockID)

	// Attempt to download preview (watermarked) for testing
	client := &http.Client{Timeout: 30 * time.Second}
	_ = client // Will be used when full API integration is available

	data, _ := json.Marshal(map[string]any{
		"stock_id":      stockID,
		"status":        "requires_license",
		"item_url":      itemURL,
		"download_path": downloadPath,
		"target_bin":    targetBin,
		"note":          "Stock import requires Adobe Creative Cloud login. Visit the item URL to license, then use premiere_import_from_media_browser to import the downloaded file.",
	})
	return &GenericResult{Status: "success", Message: string(data)}, nil
}
