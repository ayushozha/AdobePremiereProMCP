package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	"github.com/anthropics/premierpro-mcp/go-orchestrator/internal/orchestrator"
)

// Re-export AI types used by tool handlers.
type (
	SmartCutParams         = orchestrator.SmartCutParams
	SmartCutResult         = orchestrator.SmartCutResult
	SmartTrimParams        = orchestrator.SmartTrimParams
	SmartTrimResult        = orchestrator.SmartTrimResult
	AutoColorMatchParams   = orchestrator.AutoColorMatchParams
	AutoColorMatchResult   = orchestrator.AutoColorMatchResult
	AutoAudioLevelsParams  = orchestrator.AutoAudioLevelsParams
	AutoAudioLevelsResult  = orchestrator.AutoAudioLevelsResult
	ClipAnalysis           = orchestrator.ClipAnalysis
	SequenceAnalysis       = orchestrator.SequenceAnalysis
	SequenceStatistics     = orchestrator.SequenceStatistics
	JumpCutResult          = orchestrator.JumpCutResult
	AudioIssuesResult      = orchestrator.AudioIssuesResult
	RoughCutParams         = orchestrator.RoughCutParams
	RoughCutResult         = orchestrator.RoughCutResult
	RefineEditParams       = orchestrator.RefineEditParams
	RefineEditResult       = orchestrator.RefineEditResult
	GenerateTrailerParams  = orchestrator.GenerateTrailerParams
	GenerateTrailerResult  = orchestrator.GenerateTrailerResult
	SocialCutParams        = orchestrator.SocialCutParams
	SocialCutResult        = orchestrator.SocialCutResult
	AutoOrganizeParams     = orchestrator.AutoOrganizeParams
	AutoOrganizeResult     = orchestrator.AutoOrganizeResult
	TagClipsResult         = orchestrator.TagClipsResult
	FindSimilarResult      = orchestrator.FindSimilarResult
	DeliverySpecResult     = orchestrator.DeliverySpecResult
	ProjectReportResult    = orchestrator.ProjectReportResult
)

// aiH is a handler wrapper for AI tools (mirrors advH / efxH pattern).
func aiH(_ Orchestrator, logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}

// registerAITools2 registers all 25 AI-powered intelligent editing MCP tools.
// These tools orchestrate across multiple services (Rust + Python + TypeScript)
// for intelligent editing workflows.
func registerAITools2(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	registerAISmartEditTools(s, orch, logger)
	registerAIContentAnalysisTools(s, orch, logger)
	registerAIScriptPipelineTools(s, orch, logger)
	registerAISmartOrgTools(s, orch, logger)
	registerAIWorkflowTools(s, orch, logger)
}

// ---------------------------------------------------------------------------
// Smart Editing (1-6)
// ---------------------------------------------------------------------------

func registerAISmartEditTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// 1. premiere_smart_cut
	s.AddTool(gomcp.NewTool("premiere_smart_cut",
		gomcp.WithDescription("Analyze audio for silence and automatically cut dead air from a track. Uses Rust waveform analysis to detect silence regions, Python to compute optimal cut points, and the TypeScript bridge to execute razor cuts."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the target sequence")),
		gomcp.WithNumber("track_index", gomcp.Description("Zero-based audio track index to analyze (default: 0)")),
		gomcp.WithNumber("silence_threshold_db", gomcp.Description("Silence threshold in dB (default: -40)")),
		gomcp.WithNumber("min_silence_duration", gomcp.Description("Minimum silence duration in seconds to trigger a cut (default: 0.5)")),
		gomcp.WithNumber("padding_seconds", gomcp.Description("Seconds of padding to keep around non-silent audio (default: 0.1)")),
	), aiH(orch, logger, "smart_cut", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		params := &SmartCutParams{
			SequenceID:         seqID,
			TrackIndex:         gomcp.ParseInt(req, "track_index", 0),
			SilenceThresholdDB: gomcp.ParseFloat64(req, "silence_threshold_db", -40),
			MinSilenceDuration: gomcp.ParseFloat64(req, "min_silence_duration", 0.5),
			PaddingSeconds:     gomcp.ParseFloat64(req, "padding_seconds", 0.1),
		}
		result, err := orch.SmartCut(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 2. premiere_smart_trim
	s.AddTool(gomcp.NewTool("premiere_smart_trim",
		gomcp.WithDescription("Trim a clip to remove silence/dead air at its start and end. Uses Rust waveform analysis to detect leading/trailing silence, then trims the clip via the TypeScript bridge."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the target sequence")),
		gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based track index")),
		gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		gomcp.WithNumber("silence_threshold_db", gomcp.Description("Silence threshold in dB (default: -40)")),
		gomcp.WithNumber("padding_seconds", gomcp.Description("Seconds of padding to keep at clip boundaries (default: 0.05)")),
	), aiH(orch, logger, "smart_trim", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		params := &SmartTrimParams{
			SequenceID:         seqID,
			TrackIndex:         gomcp.ParseInt(req, "track_index", 0),
			ClipIndex:          gomcp.ParseInt(req, "clip_index", 0),
			SilenceThresholdDB: gomcp.ParseFloat64(req, "silence_threshold_db", -40),
			PaddingSeconds:     gomcp.ParseFloat64(req, "padding_seconds", 0.05),
		}
		result, err := orch.SmartTrim(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 3. premiere_auto_color_match
	s.AddTool(gomcp.NewTool("premiere_auto_color_match",
		gomcp.WithDescription("Match the color grading of a destination clip to a source reference clip using Lumetri. Analyzes both clips via Rust, computes adjustments with Python, and applies Lumetri parameters via the TypeScript bridge."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the target sequence")),
		gomcp.WithNumber("src_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the source (reference) clip")),
		gomcp.WithNumber("src_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the source (reference) clip")),
		gomcp.WithNumber("dest_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index of the destination clip to adjust")),
		gomcp.WithNumber("dest_clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index of the destination clip to adjust")),
	), aiH(orch, logger, "auto_color_match", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		params := &AutoColorMatchParams{
			SequenceID:     seqID,
			SrcTrackIndex:  gomcp.ParseInt(req, "src_track_index", 0),
			SrcClipIndex:   gomcp.ParseInt(req, "src_clip_index", 0),
			DestTrackIndex: gomcp.ParseInt(req, "dest_track_index", 0),
			DestClipIndex:  gomcp.ParseInt(req, "dest_clip_index", 0),
		}
		result, err := orch.AutoColorMatch(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 4. premiere_auto_audio_levels
	s.AddTool(gomcp.NewTool("premiere_auto_audio_levels",
		gomcp.WithDescription("Analyze and normalize all audio clips in a sequence to broadcast standards. Uses Rust waveform analysis to measure LUFS, Python to compute per-clip gain adjustments, and the TypeScript bridge to apply them. Default target: -24 LUFS (broadcast standard)."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the target sequence")),
		gomcp.WithNumber("target_lufs", gomcp.Description("Target integrated loudness in LUFS (default: -24)")),
		gomcp.WithNumber("max_peak_db", gomcp.Description("Maximum true peak in dBFS (default: -1)")),
	), aiH(orch, logger, "auto_audio_levels", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		params := &AutoAudioLevelsParams{
			SequenceID: seqID,
			TargetLUFS: gomcp.ParseFloat64(req, "target_lufs", -24),
			MaxPeakDB:  gomcp.ParseFloat64(req, "max_peak_db", -1),
		}
		result, err := orch.AutoAudioLevels(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 5. premiere_suggest_transitions
	s.AddTool(gomcp.NewTool("premiere_suggest_transitions",
		gomcp.WithDescription("AI suggests appropriate transition types and durations for each edit point in a sequence, based on content analysis of adjacent clips (scene changes, pacing, mood)."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the target sequence")),
	), aiH(orch, logger, "suggest_transitions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.SuggestTransitions(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 6. premiere_suggest_music
	s.AddTool(gomcp.NewTool("premiere_suggest_music",
		gomcp.WithDescription("AI suggests music timing, cut points, and intensity based on scene pacing analysis. Analyzes the sequence rhythm and suggests where to place music hits, swells, and fades."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the target sequence")),
	), aiH(orch, logger, "suggest_music", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.SuggestMusic(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// ---------------------------------------------------------------------------
// Content Analysis (7-11)
// ---------------------------------------------------------------------------

func registerAIContentAnalysisTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// 7. premiere_analyze_clip
	s.AddTool(gomcp.NewTool("premiere_analyze_clip",
		gomcp.WithDescription("Full analysis of a media clip: duration, audio levels, scene changes, and motion detection. Uses the Rust media engine for frame-level and waveform analysis."),
		gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to the media file to analyze")),
	), aiH(orch, logger, "analyze_clip", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		fp := gomcp.ParseString(req, "file_path", "")
		if fp == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		result, err := orch.AnalyzeClip(ctx, fp)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 8. premiere_analyze_sequence
	s.AddTool(gomcp.NewTool("premiere_analyze_sequence",
		gomcp.WithDescription("Full sequence analysis: evaluates pacing, audio balance, gap detection, transition consistency, and overall edit quality. Combines Rust per-clip analysis with Python AI evaluation."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to analyze")),
	), aiH(orch, logger, "analyze_sequence", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.AnalyzeSequence(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 9. premiere_get_sequence_statistics
	s.AddTool(gomcp.NewTool("premiere_get_sequence_statistics",
		gomcp.WithDescription("Get summary statistics for a sequence: total duration, video/audio clip counts, average clip length, track usage, transition count, effects count, and total gap duration."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence")),
	), aiH(orch, logger, "get_sequence_statistics", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.GetSequenceStatistics(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 10. premiere_detect_jump_cuts
	s.AddTool(gomcp.NewTool("premiere_detect_jump_cuts",
		gomcp.WithDescription("Detect potential jump cuts in a sequence by analyzing visual similarity between adjacent clip boundaries. Uses Rust frame analysis and Python similarity scoring."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to analyze")),
		gomcp.WithNumber("threshold", gomcp.Description("Similarity threshold (0.0-1.0). Higher values detect more subtle jump cuts (default: 0.85)")),
	), aiH(orch, logger, "detect_jump_cuts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		threshold := gomcp.ParseFloat64(req, "threshold", 0.85)
		result, err := orch.DetectJumpCuts(ctx, seqID, threshold)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 11. premiere_detect_audio_issues
	s.AddTool(gomcp.NewTool("premiere_detect_audio_issues",
		gomcp.WithDescription("Detect audio issues in a sequence: clipping, unexpected silence, level mismatches between clips, and phase problems. Uses Rust waveform analysis and Python classification."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to analyze")),
	), aiH(orch, logger, "detect_audio_issues", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.DetectAudioIssues(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// ---------------------------------------------------------------------------
// AI Script-to-Edit Pipeline (12-16)
// ---------------------------------------------------------------------------

func registerAIScriptPipelineTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// 12. premiere_generate_rough_cut
	s.AddTool(gomcp.NewTool("premiere_generate_rough_cut",
		gomcp.WithDescription("Generate a rough cut from a script and media assets. Enhanced version of auto_edit optimised for speed: scans assets (Rust), parses script (Python), matches and assembles timeline (TS). Produces a first-pass edit ready for refinement."),
		gomcp.WithString("script_path", gomcp.Description("Path to the script file")),
		gomcp.WithString("script_text", gomcp.Description("Raw script text (used when script_path is not provided)")),
		gomcp.WithString("assets_directory", gomcp.Required(), gomcp.Description("Absolute path to the directory containing media assets")),
		gomcp.WithString("output_name", gomcp.Description("Name for the generated sequence")),
		gomcp.WithString("pacing", gomcp.Description("Editing pace"), gomcp.Enum("slow", "moderate", "fast", "dynamic")),
	), aiH(orch, logger, "generate_rough_cut", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		assetsDir := gomcp.ParseString(req, "assets_directory", "")
		if assetsDir == "" {
			return gomcp.NewToolResultError("parameter 'assets_directory' is required"), nil
		}
		scriptPath := gomcp.ParseString(req, "script_path", "")
		scriptText := gomcp.ParseString(req, "script_text", "")
		if scriptPath == "" && scriptText == "" {
			return gomcp.NewToolResultError("either 'script_path' or 'script_text' must be provided"), nil
		}
		params := &RoughCutParams{
			ScriptPath:      scriptPath,
			ScriptText:      scriptText,
			AssetsDirectory: assetsDir,
			OutputName:      gomcp.ParseString(req, "output_name", ""),
			Pacing:          gomcp.ParseString(req, "pacing", "moderate"),
		}
		result, err := orch.GenerateRoughCut(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 13. premiere_refine_edit
	s.AddTool(gomcp.NewTool("premiere_refine_edit",
		gomcp.WithDescription("AI refinement pass on an existing edit. Analyzes the current sequence and optionally adjusts pacing, adds transitions, and normalises audio levels. Uses Python for analysis and the TypeScript bridge for adjustments."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to refine")),
		gomcp.WithString("target_pacing", gomcp.Description("Desired pacing"), gomcp.Enum("slow", "moderate", "fast", "dynamic")),
		gomcp.WithBoolean("add_transitions", gomcp.Description("Whether to add transitions at edit points (default: true)")),
		gomcp.WithBoolean("adjust_audio", gomcp.Description("Whether to normalise audio levels (default: true)")),
		gomcp.WithString("target_mood", gomcp.Description("Target mood for pacing decisions (e.g. 'upbeat', 'dramatic', 'calm')")),
	), aiH(orch, logger, "refine_edit", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		params := &RefineEditParams{
			SequenceID:     seqID,
			TargetPacing:   gomcp.ParseString(req, "target_pacing", ""),
			AddTransitions: gomcp.ParseBoolean(req, "add_transitions", true),
			AdjustAudio:    gomcp.ParseBoolean(req, "adjust_audio", true),
			TargetMood:     gomcp.ParseString(req, "target_mood", ""),
		}
		result, err := orch.RefineEdit(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 14. premiere_add_broll_suggestions
	s.AddTool(gomcp.NewTool("premiere_add_broll_suggestions",
		gomcp.WithDescription("Suggest B-roll placement based on dialogue analysis. Examines the A-roll timeline, detects dialogue topics via Rust + Python, and suggests where B-roll would enhance the edit along with content keywords."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to analyze")),
	), aiH(orch, logger, "add_broll_suggestions", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.AddBRollSuggestions(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 15. premiere_generate_trailer
	s.AddTool(gomcp.NewTool("premiere_generate_trailer",
		gomcp.WithDescription("Generate a trailer or highlight reel from an existing sequence. Uses Rust to detect high-energy moments, Python to select highlights, and the TypeScript bridge to assemble a new condensed sequence."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the source sequence")),
		gomcp.WithNumber("max_duration_seconds", gomcp.Description("Maximum trailer duration in seconds (default: 60)")),
		gomcp.WithString("style", gomcp.Description("Trailer style"), gomcp.Enum("dramatic", "upbeat", "cinematic", "fast_cut")),
		gomcp.WithBoolean("include_audio", gomcp.Description("Include audio from the source sequence (default: true)")),
	), aiH(orch, logger, "generate_trailer", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		params := &GenerateTrailerParams{
			SequenceID:   seqID,
			MaxDuration:  gomcp.ParseFloat64(req, "max_duration_seconds", 60),
			Style:        gomcp.ParseString(req, "style", "dramatic"),
			IncludeAudio: gomcp.ParseBoolean(req, "include_audio", true),
		}
		result, err := orch.GenerateTrailer(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 16. premiere_create_social_cuts
	s.AddTool(gomcp.NewTool("premiere_create_social_cuts",
		gomcp.WithDescription("Create social media cuts from a sequence in different aspect ratios. Uses Python to select the best content segments and auto-reframe via the TypeScript bridge. Supports common social media formats."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the source sequence")),
		gomcp.WithString("aspect_ratio", gomcp.Required(), gomcp.Description("Target aspect ratio"), gomcp.Enum("16:9", "9:16", "1:1", "4:5")),
		gomcp.WithNumber("max_duration_seconds", gomcp.Description("Maximum duration in seconds (default: 60 for stories/reels, 0 for no limit)")),
		gomcp.WithString("platform", gomcp.Description("Target platform hint"), gomcp.Enum("youtube", "instagram", "tiktok", "twitter", "linkedin")),
	), aiH(orch, logger, "create_social_cuts", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		ar := gomcp.ParseString(req, "aspect_ratio", "")
		if ar == "" {
			return gomcp.NewToolResultError("parameter 'aspect_ratio' is required"), nil
		}
		params := &SocialCutParams{
			SequenceID:  seqID,
			AspectRatio: ar,
			MaxDuration: gomcp.ParseFloat64(req, "max_duration_seconds", 60),
			Platform:    gomcp.ParseString(req, "platform", ""),
		}
		result, err := orch.CreateSocialCuts(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// ---------------------------------------------------------------------------
// Smart Organisation (17-20)
// ---------------------------------------------------------------------------

func registerAISmartOrgTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// 17. premiere_auto_organize_project
	s.AddTool(gomcp.NewTool("premiere_auto_organize_project",
		gomcp.WithDescription("AI-powered project organisation: automatically creates bins and sorts project items by content type (video, audio, graphics), scene, and date. Uses Rust for metadata extraction, Python for classification, and the TypeScript bridge for bin management."),
		gomcp.WithString("strategy", gomcp.Description("Organisation strategy"), gomcp.Enum("by_type", "by_scene", "by_date", "auto")),
	), aiH(orch, logger, "auto_organize_project", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		params := &AutoOrganizeParams{
			Strategy: gomcp.ParseString(req, "strategy", "auto"),
		}
		result, err := orch.AutoOrganizeProject(ctx, params)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 18. premiere_tag_clips
	s.AddTool(gomcp.NewTool("premiere_tag_clips",
		gomcp.WithDescription("AI-generate tags and metadata for a clip based on content analysis. Uses Rust for visual/audio feature extraction and Python for AI classification and tagging."),
		gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to the media file to tag")),
	), aiH(orch, logger, "tag_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		fp := gomcp.ParseString(req, "file_path", "")
		if fp == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		result, err := orch.AITagClips(ctx, fp)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 19. premiere_find_similar_clips
	s.AddTool(gomcp.NewTool("premiere_find_similar_clips",
		gomcp.WithDescription("Find visually or audibly similar clips in the project compared to a reference clip. Uses Rust for media fingerprinting and Python for similarity scoring."),
		gomcp.WithString("file_path", gomcp.Required(), gomcp.Description("Absolute path to the reference media file")),
		gomcp.WithNumber("max_results", gomcp.Description("Maximum number of similar clips to return (default: 10)")),
	), aiH(orch, logger, "find_similar_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		fp := gomcp.ParseString(req, "file_path", "")
		if fp == "" {
			return gomcp.NewToolResultError("parameter 'file_path' is required"), nil
		}
		maxResults := gomcp.ParseInt(req, "max_results", 10)
		result, err := orch.FindSimilarClips(ctx, fp, maxResults)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 20. premiere_suggest_replacements
	s.AddTool(gomcp.NewTool("premiere_suggest_replacements",
		gomcp.WithDescription("Suggest better clips for timeline positions based on content quality analysis. Analyzes each clip in the sequence and suggests higher-quality alternatives from the project's media pool."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to analyze")),
	), aiH(orch, logger, "suggest_replacements", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.SuggestReplacements(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}

// ---------------------------------------------------------------------------
// Workflow Automation (21-25)
// ---------------------------------------------------------------------------

func registerAIWorkflowTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// 21. premiere_create_review_markers
	s.AddTool(gomcp.NewTool("premiere_create_review_markers",
		gomcp.WithDescription("Add review markers at key decision points in a sequence. AI analyzes the edit for potential issues (pacing problems, audio mismatches, missing transitions) and places markers with descriptions for human review."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to add review markers to")),
	), aiH(orch, logger, "create_review_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.CreateReviewMarkers(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 22. premiere_generate_edit_summary
	s.AddTool(gomcp.NewTool("premiere_generate_edit_summary",
		gomcp.WithDescription("Generate a human-readable text summary of the current edit. Describes the sequence structure, key moments, pacing, and editing style. Useful for review notes and documentation."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to summarize")),
	), aiH(orch, logger, "generate_edit_summary", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.GenerateEditSummary(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 23. premiere_estimate_render_time
	s.AddTool(gomcp.NewTool("premiere_estimate_render_time",
		gomcp.WithDescription("Estimate the render time for a sequence based on its complexity, resolution, effects load, and duration. Useful for scheduling and planning export workflows."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to estimate")),
	), aiH(orch, logger, "estimate_render_time", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		result, err := orch.EstimateRenderTime(ctx, seqID)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 24. premiere_check_delivery_specs
	s.AddTool(gomcp.NewTool("premiere_check_delivery_specs",
		gomcp.WithDescription("Check if a sequence meets delivery specifications for broadcast, web, or cinema distribution. Validates resolution, frame rate, audio levels, duration, and codec compliance."),
		gomcp.WithString("sequence_id", gomcp.Required(), gomcp.Description("ID of the sequence to check")),
		gomcp.WithString("standard", gomcp.Required(), gomcp.Description("Delivery standard to check against"), gomcp.Enum("broadcast_hd", "broadcast_4k", "web_youtube", "web_social", "cinema_dcp", "podcast")),
	), aiH(orch, logger, "check_delivery_specs", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		seqID := gomcp.ParseString(req, "sequence_id", "")
		if seqID == "" {
			return gomcp.NewToolResultError("parameter 'sequence_id' is required"), nil
		}
		standard := gomcp.ParseString(req, "standard", "")
		if standard == "" {
			return gomcp.NewToolResultError("parameter 'standard' is required"), nil
		}
		result, err := orch.CheckDeliverySpecs(ctx, seqID, standard)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))

	// 25. premiere_create_project_report
	s.AddTool(gomcp.NewTool("premiere_create_project_report",
		gomcp.WithDescription("Generate a comprehensive project report: clip usage statistics, total duration, sequence count, effects inventory, used vs unused media, and export history."),
	), aiH(orch, logger, "create_project_report", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		result, err := orch.CreateProjectReport(ctx)
		if err != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
		}
		return toolResultJSON(result)
	}))
}
