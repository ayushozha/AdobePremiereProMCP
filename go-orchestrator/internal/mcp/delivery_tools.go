package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// dlvH wraps a handler with debug logging for delivery tool invocations.
func dlvH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerDeliveryTools registers all 30 social media optimization, delivery,
// distribution, QA, and versioning MCP tools.
func registerDeliveryTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {

	// -----------------------------------------------------------------------
	// Social Media Optimization (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_create_vertical_version
	s.AddTool(gomcp.NewTool("premiere_create_vertical_version",
		gomcp.WithDescription("Create a 9:16 vertical version of a 16:9 sequence for mobile/social platforms like TikTok, Reels, and Shorts."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_name", gomcp.Description("Name for the new vertical sequence (default: auto-generated)")),
	), dlvH(logger, "create_vertical_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateVerticalVersion(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "output_name", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_create_square_version
	s.AddTool(gomcp.NewTool("premiere_create_square_version",
		gomcp.WithDescription("Create a 1:1 square version of a sequence for Instagram feed posts and other square-format platforms."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_name", gomcp.Description("Name for the new square sequence (default: auto-generated)")),
	), dlvH(logger, "create_square_version", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateSquareVersion(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "output_name", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_add_safe_zone_guides
	s.AddTool(gomcp.NewTool("premiere_add_safe_zone_guides",
		gomcp.WithDescription("Add safe zone guides for a specific social media platform (YouTube, Instagram, TikTok) to avoid UI overlay areas."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("platform", gomcp.Description("Target platform: YouTube, Instagram, or TikTok (default: YouTube)"),
			gomcp.Enum("YouTube", "Instagram", "TikTok")),
	), dlvH(logger, "add_safe_zone_guides", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.AddSafeZoneGuides(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "platform", "YouTube"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_optimize_for_platform
	s.AddTool(gomcp.NewTool("premiere_optimize_for_platform",
		gomcp.WithDescription("Auto-optimize sequence settings (resolution, frame rate, bitrate) for a target social media or distribution platform."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("platform", gomcp.Required(), gomcp.Description("Target platform"),
			gomcp.Enum("YouTube", "Instagram", "TikTok", "Twitter", "Facebook")),
	), dlvH(logger, "optimize_for_platform", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		platform := gomcp.ParseString(req, "platform", "")
		if platform == "" {
			return gomcp.NewToolResultError("parameter 'platform' is required"), nil
		}
		result, err := orch.OptimizeForPlatform(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			platform,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_create_thumbnail_from_frame
	s.AddTool(gomcp.NewTool("premiere_create_thumbnail_from_frame",
		gomcp.WithDescription("Export a frame from a sequence as a thumbnail image (e.g., for YouTube). Optionally specify text to overlay."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithNumber("time_seconds", gomcp.Description("Time position in seconds for the thumbnail frame (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output thumbnail image")),
		gomcp.WithString("add_text", gomcp.Description("Optional text to overlay on the thumbnail")),
	), dlvH(logger, "create_thumbnail_from_frame", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.CreateThumbnailFromFrame(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseFloat64(req, "time_seconds", 0),
			outputPath,
			gomcp.ParseString(req, "add_text", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Content Segmentation (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_split_into_segments
	s.AddTool(gomcp.NewTool("premiere_split_into_segments",
		gomcp.WithDescription("Split a long video into segments under a maximum duration. Useful for platform upload limits or multi-part content."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithNumber("max_duration_seconds", gomcp.Description("Maximum duration per segment in seconds (default: 600)")),
	), dlvH(logger, "split_into_segments", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SplitIntoSegments(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseFloat64(req, "max_duration_seconds", 600),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 7. premiere_create_chapters_file
	s.AddTool(gomcp.NewTool("premiere_create_chapters_file",
		gomcp.WithDescription("Create a chapters file from sequence markers in YouTube timestamp format or podcast chapter XML format."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output chapters file")),
		gomcp.WithString("format", gomcp.Description("Output format: youtube or podcast (default: youtube)"),
			gomcp.Enum("youtube", "podcast")),
	), dlvH(logger, "create_chapters_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.CreateChaptersFile(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseString(req, "format", "youtube"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_extract_segment_by_markers
	s.AddTool(gomcp.NewTool("premiere_extract_segment_by_markers",
		gomcp.WithDescription("Extract a segment of a sequence between two markers by setting in/out points."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithNumber("start_marker_index", gomcp.Required(), gomcp.Description("Zero-based index of the start marker")),
		gomcp.WithNumber("end_marker_index", gomcp.Required(), gomcp.Description("Zero-based index of the end marker")),
		gomcp.WithString("output_name", gomcp.Description("Name for the extracted segment (default: auto-generated)")),
	), dlvH(logger, "extract_segment_by_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ExtractSegmentByMarkers(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseInt(req, "start_marker_index", 0),
			gomcp.ParseInt(req, "end_marker_index", 0),
			gomcp.ParseString(req, "output_name", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_create_teaser
	s.AddTool(gomcp.NewTool("premiere_create_teaser",
		gomcp.WithDescription("Auto-create a short teaser/trailer from a sequence by selecting evenly spaced highlight moments."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithNumber("duration_seconds", gomcp.Description("Target teaser duration in seconds (default: 30)")),
		gomcp.WithString("output_name", gomcp.Description("Name for the teaser sequence (default: auto-generated)")),
	), dlvH(logger, "create_teaser", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateTeaser(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseFloat64(req, "duration_seconds", 30),
			gomcp.ParseString(req, "output_name", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_create_bumper
	s.AddTool(gomcp.NewTool("premiere_create_bumper",
		gomcp.WithDescription("Create an intro/outro bumper sequence with text, duration, and style settings."),
		gomcp.WithString("text", gomcp.Required(), gomcp.Description("Text content for the bumper")),
		gomcp.WithNumber("duration", gomcp.Description("Duration of the bumper in seconds (default: 5)")),
		gomcp.WithString("style", gomcp.Description("Visual style: simple, bold, minimal, cinematic (default: simple)"),
			gomcp.Enum("simple", "bold", "minimal", "cinematic")),
		gomcp.WithString("output_name", gomcp.Description("Name for the bumper sequence (default: auto-generated)")),
	), dlvH(logger, "create_bumper", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		text := gomcp.ParseString(req, "text", "")
		if text == "" {
			return gomcp.NewToolResultError("parameter 'text' is required"), nil
		}
		result, err := orch.CreateBumper(ctx,
			text,
			gomcp.ParseFloat64(req, "duration", 5),
			gomcp.ParseString(req, "style", "simple"),
			gomcp.ParseString(req, "output_name", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Delivery Formats (11-15)
	// -----------------------------------------------------------------------

	// 11. premiere_export_for_broadcast
	s.AddTool(gomcp.NewTool("premiere_export_for_broadcast",
		gomcp.WithDescription("Prepare export settings for broadcast delivery standards (ATSC for North America, DVB for Europe, ISDB for Japan/Brazil)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("standard", gomcp.Description("Broadcast standard: ATSC, DVB, or ISDB (default: ATSC)"),
			gomcp.Enum("ATSC", "DVB", "ISDB")),
	), dlvH(logger, "export_for_broadcast", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportForBroadcast(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseString(req, "standard", "ATSC"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 12. premiere_export_for_streaming
	s.AddTool(gomcp.NewTool("premiere_export_for_streaming",
		gomcp.WithDescription("Prepare export settings for streaming platform delivery (Netflix, Amazon Prime Video, Disney+)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("platform", gomcp.Description("Streaming platform: Netflix, Amazon, or Disney+ (default: Netflix)"),
			gomcp.Enum("Netflix", "Amazon", "Disney+")),
	), dlvH(logger, "export_for_streaming", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportForStreaming(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseString(req, "platform", "Netflix"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_export_for_archive
	s.AddTool(gomcp.NewTool("premiere_export_for_archive",
		gomcp.WithDescription("Prepare export settings for archival with lossless or near-lossless codecs (ProRes 4444, DNxHR 444, FFV1)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("codec", gomcp.Description("Archival codec: prores4444, dnxhr444, or lossless (default: prores4444)"),
			gomcp.Enum("prores4444", "dnxhr444", "lossless")),
	), dlvH(logger, "export_for_archive", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportForArchive(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseString(req, "codec", "prores4444"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_export_for_web
	s.AddTool(gomcp.NewTool("premiere_export_for_web",
		gomcp.WithDescription("Prepare export settings for web delivery with adaptive bitrate-ready H.264 encoding at various quality levels."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("quality", gomcp.Description("Quality level: high (1080p), medium (720p), or low (480p) (default: medium)"),
			gomcp.Enum("high", "medium", "low")),
	), dlvH(logger, "export_for_web", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportForWeb(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseString(req, "quality", "medium"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_export_for_mobile
	s.AddTool(gomcp.NewTool("premiere_export_for_mobile",
		gomcp.WithDescription("Prepare export settings optimized for mobile device playback (iPhone/iOS, Android, iPad/Tablet)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_path", gomcp.Required(), gomcp.Description("Absolute path for the output file")),
		gomcp.WithString("device", gomcp.Description("Target device: iPhone, Android, iPad, or generic (default: generic)"),
			gomcp.Enum("iPhone", "Android", "iPad", "generic")),
	), dlvH(logger, "export_for_mobile", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputPath := gomcp.ParseString(req, "output_path", "")
		if outputPath == "" {
			return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
		}
		result, err := orch.ExportForMobile(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputPath,
			gomcp.ParseString(req, "device", "generic"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Metadata for Distribution (16-20)
	// -----------------------------------------------------------------------

	// 16. premiere_set_distribution_metadata
	s.AddTool(gomcp.NewTool("premiere_set_distribution_metadata",
		gomcp.WithDescription("Set distribution metadata (title, description, tags, category) on a sequence for platform publishing."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("title", gomcp.Description("Title for distribution")),
		gomcp.WithString("description", gomcp.Description("Description for distribution")),
		gomcp.WithString("tags", gomcp.Description("Comma-separated tags for distribution")),
		gomcp.WithString("category", gomcp.Description("Content category for distribution")),
	), dlvH(logger, "set_distribution_metadata", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.SetDistributionMetadata(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "title", ""),
			gomcp.ParseString(req, "description", ""),
			gomcp.ParseString(req, "tags", ""),
			gomcp.ParseString(req, "category", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 17. premiere_get_distribution_metadata
	s.AddTool(gomcp.NewTool("premiere_get_distribution_metadata",
		gomcp.WithDescription("Get distribution metadata (title, description, tags, category) from a sequence."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
	), dlvH(logger, "get_distribution_metadata", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetDistributionMetadata(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_embed_thumbnail_in_file
	s.AddTool(gomcp.NewTool("premiere_embed_thumbnail_in_file",
		gomcp.WithDescription("Embed a thumbnail image in an exported video file via XMP sidecar. Provides ffmpeg command for direct embedding."),
		gomcp.WithString("video_path", gomcp.Required(), gomcp.Description("Absolute path to the video file")),
		gomcp.WithString("thumbnail_path", gomcp.Required(), gomcp.Description("Absolute path to the thumbnail image")),
	), dlvH(logger, "embed_thumbnail_in_file", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		videoPath := gomcp.ParseString(req, "video_path", "")
		if videoPath == "" {
			return gomcp.NewToolResultError("parameter 'video_path' is required"), nil
		}
		thumbnailPath := gomcp.ParseString(req, "thumbnail_path", "")
		if thumbnailPath == "" {
			return gomcp.NewToolResultError("parameter 'thumbnail_path' is required"), nil
		}
		result, err := orch.EmbedThumbnailInFile(ctx, videoPath, thumbnailPath)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_add_chapter_metadata
	s.AddTool(gomcp.NewTool("premiere_add_chapter_metadata",
		gomcp.WithDescription("Add chapter metadata to an exported video file. Creates ffmetadata file and provides ffmpeg command for embedding."),
		gomcp.WithString("video_path", gomcp.Required(), gomcp.Description("Absolute path to the video file")),
		gomcp.WithString("chapters_json", gomcp.Required(), gomcp.Description("JSON array of chapters: [{\"startSeconds\":0,\"endSeconds\":60,\"title\":\"Intro\"},...]")),
	), dlvH(logger, "add_chapter_metadata", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		videoPath := gomcp.ParseString(req, "video_path", "")
		if videoPath == "" {
			return gomcp.NewToolResultError("parameter 'video_path' is required"), nil
		}
		chaptersJSON := gomcp.ParseString(req, "chapters_json", "")
		if chaptersJSON == "" {
			return gomcp.NewToolResultError("parameter 'chapters_json' is required"), nil
		}
		result, err := orch.AddChapterMetadata(ctx, videoPath, chaptersJSON)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_set_content_rating
	s.AddTool(gomcp.NewTool("premiere_set_content_rating",
		gomcp.WithDescription("Set content rating metadata on a sequence (e.g., TV-PG, PG-13, R, TV-MA)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("rating", gomcp.Required(), gomcp.Description("Content rating (e.g., G, PG, PG-13, R, TV-PG, TV-14, TV-MA)")),
	), dlvH(logger, "set_content_rating", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		rating := gomcp.ParseString(req, "rating", "")
		if rating == "" {
			return gomcp.NewToolResultError("parameter 'rating' is required"), nil
		}
		result, err := orch.SetContentRating(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			rating,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Quality Assurance (21-25)
	// -----------------------------------------------------------------------

	// 21. premiere_run_qa_checklist
	s.AddTool(gomcp.NewTool("premiere_run_qa_checklist",
		gomcp.WithDescription("Run a QA checklist against delivery specs. Checks duration, tracks, gaps, and other quality metrics."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("specs", gomcp.Description("JSON object with target specs: {\"maxDuration\":3600,\"targetWidth\":1920,\"targetHeight\":1080}")),
	), dlvH(logger, "run_qa_checklist", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.RunQAChecklist(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "specs", "{}"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_check_loudness_compliance
	s.AddTool(gomcp.NewTool("premiere_check_loudness_compliance",
		gomcp.WithDescription("Check audio loudness compliance against broadcast standards (EBU R128 for Europe, ATSC A/85 for North America)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("standard", gomcp.Description("Loudness standard: 'EBU R128' or 'ATSC A/85' (default: EBU R128)"),
			gomcp.Enum("EBU R128", "ATSC A/85")),
	), dlvH(logger, "check_loudness_compliance", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckLoudnessCompliance(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "standard", "EBU R128"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_check_color_compliance
	s.AddTool(gomcp.NewTool("premiere_check_color_compliance",
		gomcp.WithDescription("Check color space compliance against delivery standards (Rec.709 for HD, Rec.2020 for UHD/HDR)."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("standard", gomcp.Description("Color standard: 'Rec.709' or 'Rec.2020' (default: Rec.709)"),
			gomcp.Enum("Rec.709", "Rec.2020")),
	), dlvH(logger, "check_color_compliance", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckColorCompliance(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			gomcp.ParseString(req, "standard", "Rec.709"),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_check_frame_accuracy
	s.AddTool(gomcp.NewTool("premiere_check_frame_accuracy",
		gomcp.WithDescription("Check for frame-accurate edits by detecting flash frames (< 1 frame duration) and sub-frame gaps between clips."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
	), dlvH(logger, "check_frame_accuracy", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CheckFrameAccuracy(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_validate_closed_captions
	s.AddTool(gomcp.NewTool("premiere_validate_closed_captions",
		gomcp.WithDescription("Validate closed captions for FCC compliance. Checks for caption track presence, coverage, and format."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
	), dlvH(logger, "validate_closed_captions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.ValidateClosedCaptions(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// -----------------------------------------------------------------------
	// Versioning (26-30)
	// -----------------------------------------------------------------------

	// 26. premiere_create_versioned_export
	s.AddTool(gomcp.NewTool("premiere_create_versioned_export",
		gomcp.WithDescription("Prepare an export with version tracking. Creates a versioned filename and updates the export history manifest."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_dir", gomcp.Required(), gomcp.Description("Absolute path to the output directory")),
		gomcp.WithString("version_name", gomcp.Description("Version identifier (default: auto-generated timestamp)")),
		gomcp.WithString("notes", gomcp.Description("Version notes or changelog")),
	), dlvH(logger, "create_versioned_export", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputDir := gomcp.ParseString(req, "output_dir", "")
		if outputDir == "" {
			return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
		}
		result, err := orch.CreateVersionedExport(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputDir,
			gomcp.ParseString(req, "version_name", ""),
			gomcp.ParseString(req, "notes", ""),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 27. premiere_get_export_history
	s.AddTool(gomcp.NewTool("premiere_get_export_history",
		gomcp.WithDescription("Get the export version history for a sequence, including all previously versioned exports."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
	), dlvH(logger, "get_export_history", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.GetExportHistory2(ctx, gomcp.ParseInt(req, "sequence_index", 0))
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 28. premiere_compare_export_versions
	s.AddTool(gomcp.NewTool("premiere_compare_export_versions",
		gomcp.WithDescription("Compare two export version files by file size, modification date, and existence."),
		gomcp.WithString("version1_path", gomcp.Required(), gomcp.Description("Absolute path to the first version file")),
		gomcp.WithString("version2_path", gomcp.Required(), gomcp.Description("Absolute path to the second version file")),
	), dlvH(logger, "compare_export_versions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		v1 := gomcp.ParseString(req, "version1_path", "")
		if v1 == "" {
			return gomcp.NewToolResultError("parameter 'version1_path' is required"), nil
		}
		v2 := gomcp.ParseString(req, "version2_path", "")
		if v2 == "" {
			return gomcp.NewToolResultError("parameter 'version2_path' is required"), nil
		}
		result, err := orch.CompareExportVersions(ctx, v1, v2)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 29. premiere_create_approval_package
	s.AddTool(gomcp.NewTool("premiere_create_approval_package",
		gomcp.WithDescription("Create a client approval package containing a review report, placeholder for video, thumbnail, and stills."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("output_dir", gomcp.Required(), gomcp.Description("Absolute path to the output directory")),
	), dlvH(logger, "create_approval_package", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		outputDir := gomcp.ParseString(req, "output_dir", "")
		if outputDir == "" {
			return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
		}
		result, err := orch.CreateApprovalPackage(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			outputDir,
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 30. premiere_archive_and_cleanup
	s.AddTool(gomcp.NewTool("premiere_archive_and_cleanup",
		gomcp.WithDescription("Archive the project to a directory (copies project file, writes manifest) and optionally delete preview/render files."),
		gomcp.WithNumber("sequence_index", gomcp.Description("Zero-based sequence index (default: 0)")),
		gomcp.WithString("archive_dir", gomcp.Required(), gomcp.Description("Absolute path to the archive directory")),
		gomcp.WithBoolean("delete_renders", gomcp.Description("Delete preview/render files after archiving (default: false)")),
	), dlvH(logger, "archive_and_cleanup", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		archiveDir := gomcp.ParseString(req, "archive_dir", "")
		if archiveDir == "" {
			return gomcp.NewToolResultError("parameter 'archive_dir' is required"), nil
		}
		result, err := orch.ArchiveAndCleanup(ctx,
			gomcp.ParseInt(req, "sequence_index", 0),
			archiveDir,
			gomcp.ParseBoolean(req, "delete_renders", false),
		)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
