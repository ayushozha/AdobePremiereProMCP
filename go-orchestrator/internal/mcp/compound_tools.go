package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerCompoundTools registers compound/multi-step editing MCP tools.
func registerCompoundTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Multi-Step Editing (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_create_montage
	s.AddTool(
		gomcp.NewTool("premiere_create_montage",
			gomcp.WithDescription("Auto-assemble clips into a montage with transitions and optional background music. Combines clip placement, transitions, and music into one operation."),
			gomcp.WithArray("clip_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based project item indices to include in the montage"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithString("transition_name",
				gomcp.Description("Transition to apply between clips (default: Cross Dissolve)"),
			),
			gomcp.WithNumber("transition_duration",
				gomcp.Description("Transition duration in seconds (default: 1.0)"),
			),
			gomcp.WithString("music_path",
				gomcp.Description("Absolute path to background music file (optional)"),
			),
		),
		compoundH(logger, "create_montage", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			indices, err := extractIntSlice(req, "clip_indices")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid clip_indices: %v", err)), nil
			}
			result, err := orch.CreateMontage(ctx, indices,
				gomcp.ParseString(req, "transition_name", "Cross Dissolve"),
				gomcp.ParseFloat64(req, "transition_duration", 1.0),
				gomcp.ParseString(req, "music_path", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 2. premiere_create_slideshow
	s.AddTool(
		gomcp.NewTool("premiere_create_slideshow",
			gomcp.WithDescription("Create a slideshow from images with configurable slide duration, transitions between slides, and optional background music."),
			gomcp.WithArray("image_indices",
				gomcp.Required(),
				gomcp.Description("Array of zero-based project item indices for images"),
				gomcp.WithNumberItems(),
			),
			gomcp.WithNumber("slide_duration",
				gomcp.Description("Duration per slide in seconds (default: 5.0)"),
			),
			gomcp.WithString("transition_name",
				gomcp.Description("Transition between slides (default: Cross Dissolve)"),
			),
			gomcp.WithString("music_path",
				gomcp.Description("Absolute path to background music file (optional)"),
			),
		),
		compoundH(logger, "create_slideshow", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			indices, err := extractIntSlice(req, "image_indices")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid image_indices: %v", err)), nil
			}
			result, err := orch.CreateSlideshow(ctx, indices,
				gomcp.ParseFloat64(req, "slide_duration", 5.0),
				gomcp.ParseString(req, "transition_name", "Cross Dissolve"),
				gomcp.ParseString(req, "music_path", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 3. premiere_create_highlight_reel
	s.AddTool(
		gomcp.NewTool("premiere_create_highlight_reel",
			gomcp.WithDescription("Extract marker-tagged sections from a sequence into a new highlight reel. Filters markers by color to select sections."),
			gomcp.WithNumber("sequence_index",
				gomcp.Description("Zero-based index of the source sequence (default: 0)"),
			),
			gomcp.WithString("marker_color",
				gomcp.Description("Marker color to filter by (empty = all markers)"),
			),
			gomcp.WithString("output_name",
				gomcp.Description("Name for the highlight reel sequence (default: 'Highlight Reel')"),
			),
		),
		compoundH(logger, "create_highlight_reel", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.CreateHighlightReel(ctx,
				gomcp.ParseInt(req, "sequence_index", 0),
				gomcp.ParseString(req, "marker_color", ""),
				gomcp.ParseString(req, "output_name", "Highlight Reel"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 4. premiere_ripple_delete_empty_spaces
	s.AddTool(
		gomcp.NewTool("premiere_ripple_delete_empty_spaces",
			gomcp.WithDescription("Remove all empty spaces (gaps) across all video and audio tracks by shifting clips to close gaps."),
		),
		compoundH(logger, "ripple_delete_empty_spaces", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.RippleDeleteEmptySpaces(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 5. premiere_align_all_clips_to_track
	s.AddTool(
		gomcp.NewTool("premiere_align_all_clips_to_track",
			gomcp.WithDescription("Align clips on a source track to match the timing of clips on a destination track."),
			gomcp.WithNumber("source_track",
				gomcp.Required(),
				gomcp.Description("Zero-based source video track index"),
			),
			gomcp.WithNumber("dest_track",
				gomcp.Required(),
				gomcp.Description("Zero-based destination video track index"),
			),
		),
		compoundH(logger, "align_all_clips_to_track", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.AlignAllClipsToTrack(ctx,
				gomcp.ParseInt(req, "source_track", 0),
				gomcp.ParseInt(req, "dest_track", 1),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Audio-Visual Sync (6-10)
	// -----------------------------------------------------------------------

	// 6. premiere_sync_all_audio_to_video
	s.AddTool(
		gomcp.NewTool("premiere_sync_all_audio_to_video",
			gomcp.WithDescription("Auto-sync all audio clips to their nearest video clip by matching timecodes."),
		),
		compoundH(logger, "sync_all_audio_to_video", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.SyncAllAudioToVideo(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 7. premiere_replace_audio
	s.AddTool(
		gomcp.NewTool("premiere_replace_audio",
			gomcp.WithDescription("Replace the audio track of a video clip with a new audio file. Imports the audio and places it aligned to the video clip."),
			gomcp.WithNumber("video_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
			gomcp.WithNumber("video_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index on the video track"),
			),
			gomcp.WithString("audio_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the replacement audio file"),
			),
		),
		compoundH(logger, "replace_audio", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			audioPath := gomcp.ParseString(req, "audio_path", "")
			if audioPath == "" {
				return gomcp.NewToolResultError("parameter 'audio_path' is required"), nil
			}
			result, err := orch.ReplaceAudio(ctx,
				gomcp.ParseInt(req, "video_track_index", 0),
				gomcp.ParseInt(req, "video_clip_index", 0),
				audioPath,
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 8. premiere_add_music_bed
	s.AddTool(
		gomcp.NewTool("premiere_add_music_bed",
			gomcp.WithDescription("Add background music with fade-in, fade-out, and volume control. Imports the audio file and places it on the timeline."),
			gomcp.WithString("audio_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the music file"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based audio track index (default: 0)"),
			),
			gomcp.WithNumber("start_time",
				gomcp.Description("Start time in seconds (default: 0)"),
			),
			gomcp.WithNumber("end_time",
				gomcp.Description("End time in seconds (default: 0 = full duration)"),
			),
			gomcp.WithNumber("fade_in",
				gomcp.Description("Fade-in duration in seconds (default: 0)"),
			),
			gomcp.WithNumber("fade_out",
				gomcp.Description("Fade-out duration in seconds (default: 0)"),
			),
			gomcp.WithNumber("volume",
				gomcp.Description("Volume in dB, 0 = unity (default: 0)"),
			),
		),
		compoundH(logger, "add_music_bed", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			audioPath := gomcp.ParseString(req, "audio_path", "")
			if audioPath == "" {
				return gomcp.NewToolResultError("parameter 'audio_path' is required"), nil
			}
			result, err := orch.AddMusicBed(ctx, audioPath,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "start_time", 0),
				gomcp.ParseFloat64(req, "end_time", 0),
				gomcp.ParseFloat64(req, "fade_in", 0),
				gomcp.ParseFloat64(req, "fade_out", 0),
				gomcp.ParseFloat64(req, "volume", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 9. premiere_duck_music_under_dialogue
	s.AddTool(
		gomcp.NewTool("premiere_duck_music_under_dialogue",
			gomcp.WithDescription("Auto-duck music volume under dialogue. Analyzes dialogue track timing and reduces music volume in overlapping regions."),
			gomcp.WithNumber("music_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based audio track index for music"),
			),
			gomcp.WithNumber("dialogue_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based audio track index for dialogue"),
			),
			gomcp.WithNumber("duck_amount",
				gomcp.Description("Amount to reduce music in dB (default: 12)"),
			),
		),
		compoundH(logger, "duck_music_under_dialogue", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.DuckMusicUnderDialogue(ctx,
				gomcp.ParseInt(req, "music_track_index", 0),
				gomcp.ParseInt(req, "dialogue_track_index", 1),
				gomcp.ParseFloat64(req, "duck_amount", 12),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 10. premiere_add_sound_effect
	s.AddTool(
		gomcp.NewTool("premiere_add_sound_effect",
			gomcp.WithDescription("Place a sound effect at a specific time on an audio track. Imports the SFX file and places it at the specified position."),
			gomcp.WithString("sfx_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the sound effect file"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based audio track index (default: 0)"),
			),
			gomcp.WithNumber("time",
				gomcp.Description("Time in seconds to place the SFX (default: 0)"),
			),
			gomcp.WithNumber("volume",
				gomcp.Description("Volume in dB (default: 0)"),
			),
		),
		compoundH(logger, "add_sound_effect", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			sfxPath := gomcp.ParseString(req, "sfx_path", "")
			if sfxPath == "" {
				return gomcp.NewToolResultError("parameter 'sfx_path' is required"), nil
			}
			result, err := orch.AddSoundEffect(ctx, sfxPath,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "time", 0),
				gomcp.ParseFloat64(req, "volume", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Color Workflow (11-14)
	// -----------------------------------------------------------------------

	// 11. premiere_match_color_between_clips
	s.AddTool(
		gomcp.NewTool("premiere_match_color_between_clips",
			gomcp.WithDescription("Copy Lumetri Color settings from a source clip to a destination clip, matching color grading between them."),
			gomcp.WithNumber("src_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based source video track index"),
			),
			gomcp.WithNumber("src_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based source clip index"),
			),
			gomcp.WithNumber("dest_track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based destination video track index"),
			),
			gomcp.WithNumber("dest_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based destination clip index"),
			),
		),
		compoundH(logger, "match_color_between_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.MatchColorBetweenClips(ctx,
				gomcp.ParseInt(req, "src_track_index", 0),
				gomcp.ParseInt(req, "src_clip_index", 0),
				gomcp.ParseInt(req, "dest_track_index", 0),
				gomcp.ParseInt(req, "dest_clip_index", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 12. premiere_apply_color_preset
	s.AddTool(
		gomcp.NewTool("premiere_apply_color_preset",
			gomcp.WithDescription("Apply a named color preset to a clip. Available presets: Warm, Cool, Vintage, Cinematic, Desaturated, HighContrast."),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
			gomcp.WithNumber("clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based clip index on the track"),
			),
			gomcp.WithString("preset_name",
				gomcp.Required(),
				gomcp.Description("Color preset name"),
				gomcp.Enum("Warm", "Cool", "Vintage", "Cinematic", "Desaturated", "HighContrast"),
			),
		),
		compoundH(logger, "apply_color_preset", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			presetName := gomcp.ParseString(req, "preset_name", "Warm")
			result, err := orch.ApplyColorPreset(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseInt(req, "clip_index", 0),
				presetName,
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 13. premiere_create_color_gradient
	s.AddTool(
		gomcp.NewTool("premiere_create_color_gradient",
			gomcp.WithDescription("Apply a gradual color temperature change across a range of clips on a track, creating a smooth color gradient."),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
			gomcp.WithNumber("start_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the first clip in the range"),
			),
			gomcp.WithNumber("end_clip_index",
				gomcp.Required(),
				gomcp.Description("Zero-based index of the last clip in the range"),
			),
			gomcp.WithNumber("start_temp",
				gomcp.Required(),
				gomcp.Description("Starting color temperature value"),
			),
			gomcp.WithNumber("end_temp",
				gomcp.Required(),
				gomcp.Description("Ending color temperature value"),
			),
		),
		compoundH(logger, "create_color_gradient", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.CreateColorGradient(ctx,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseInt(req, "start_clip_index", 0),
				gomcp.ParseInt(req, "end_clip_index", 0),
				gomcp.ParseFloat64(req, "start_temp", 0),
				gomcp.ParseFloat64(req, "end_temp", 0),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 14. premiere_auto_correct_all_clips
	s.AddTool(
		gomcp.NewTool("premiere_auto_correct_all_clips",
			gomcp.WithDescription("Apply automatic color correction to all clips on a video track using Lumetri auto-tone."),
			gomcp.WithNumber("track_index",
				gomcp.Required(),
				gomcp.Description("Zero-based video track index"),
			),
		),
		compoundH(logger, "auto_correct_all_clips", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.AutoCorrectAllClips(ctx, gomcp.ParseInt(req, "track_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Text Workflow (15-18)
	// -----------------------------------------------------------------------

	// 15. premiere_add_subtitles_from_srt
	s.AddTool(
		gomcp.NewTool("premiere_add_subtitles_from_srt",
			gomcp.WithDescription("Import an SRT subtitle file and place subtitles on the timeline. Parses SRT format and creates caption entries."),
			gomcp.WithString("srt_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the SRT subtitle file"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based track index for captions (default: 0)"),
			),
		),
		compoundH(logger, "add_subtitles_from_srt", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			srtPath := gomcp.ParseString(req, "srt_path", "")
			if srtPath == "" {
				return gomcp.NewToolResultError("parameter 'srt_path' is required"), nil
			}
			result, err := orch.AddSubtitlesFromSRT(ctx, srtPath, gomcp.ParseInt(req, "track_index", 0))
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 16. premiere_add_end_credits
	s.AddTool(
		gomcp.NewTool("premiere_add_end_credits",
			gomcp.WithDescription("Add scrolling end credits to the timeline. Accepts credit entries as JSON and creates a scrolling title."),
			gomcp.WithString("credits",
				gomcp.Required(),
				gomcp.Description("JSON array of credit entries: [{\"role\":\"Director\",\"name\":\"Jane Smith\"}]"),
			),
			gomcp.WithNumber("track_index",
				gomcp.Description("Zero-based video track index (default: 0)"),
			),
			gomcp.WithNumber("scroll_duration",
				gomcp.Description("Duration of the credits scroll in seconds (default: 10)"),
			),
			gomcp.WithString("style",
				gomcp.Description("Credits style preset"),
				gomcp.Enum("simple", "cinematic", "modern"),
			),
		),
		compoundH(logger, "add_end_credits", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			credits := gomcp.ParseString(req, "credits", "")
			if credits == "" {
				return gomcp.NewToolResultError("parameter 'credits' is required"), nil
			}
			result, err := orch.AddEndCredits(ctx, credits,
				gomcp.ParseInt(req, "track_index", 0),
				gomcp.ParseFloat64(req, "scroll_duration", 10),
				gomcp.ParseString(req, "style", "simple"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 17. premiere_add_chapter_markers
	s.AddTool(
		gomcp.NewTool("premiere_add_chapter_markers",
			gomcp.WithDescription("Add chapter markers to the active sequence from a JSON array of time/title pairs."),
			gomcp.WithString("chapters",
				gomcp.Required(),
				gomcp.Description("JSON array of chapters: [{\"time\":0,\"title\":\"Intro\"},{\"time\":60,\"title\":\"Main\"}]"),
			),
		),
		compoundH(logger, "add_chapter_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			chapters := gomcp.ParseString(req, "chapters", "")
			if chapters == "" {
				return gomcp.NewToolResultError("parameter 'chapters' is required"), nil
			}
			result, err := orch.AddChapterMarkers(ctx, chapters)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 18. premiere_generate_chapters_from_markers
	s.AddTool(
		gomcp.NewTool("premiere_generate_chapters_from_markers",
			gomcp.WithDescription("Export sequence markers as YouTube-compatible chapter format (timestamp + title per line)."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute file path to write the chapters text file"),
			),
		),
		compoundH(logger, "generate_chapters_from_markers", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.GenerateChaptersFromMarkers(ctx, outputPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Export Workflow (19-23)
	// -----------------------------------------------------------------------

	// 19. premiere_export_for_youtube
	s.AddTool(
		gomcp.NewTool("premiere_export_for_youtube",
			gomcp.WithDescription("Export the active sequence with YouTube-optimized settings (H.264, high bitrate, AAC 320kbps)."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output file"),
			),
			gomcp.WithString("title",
				gomcp.Description("Video title metadata"),
			),
			gomcp.WithString("description",
				gomcp.Description("Video description metadata"),
			),
		),
		compoundH(logger, "export_for_youtube", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportForYouTube(ctx, outputPath,
				gomcp.ParseString(req, "title", ""),
				gomcp.ParseString(req, "description", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 20. premiere_export_for_instagram
	s.AddTool(
		gomcp.NewTool("premiere_export_for_instagram",
			gomcp.WithDescription("Export for Instagram with a specific aspect ratio. Supports 1:1 (feed), 4:5 (portrait), 9:16 (stories/reels)."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output file"),
			),
			gomcp.WithString("aspect_ratio",
				gomcp.Description("Target aspect ratio"),
				gomcp.Enum("1:1", "4:5", "9:16"),
			),
		),
		compoundH(logger, "export_for_instagram", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportForInstagram(ctx, outputPath,
				gomcp.ParseString(req, "aspect_ratio", "1:1"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 21. premiere_export_for_tiktok
	s.AddTool(
		gomcp.NewTool("premiere_export_for_tiktok",
			gomcp.WithDescription("Export for TikTok with optimized settings (9:16, 1080x1920, H.264, <3 minutes)."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output file"),
			),
		),
		compoundH(logger, "export_for_tiktok", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportForTikTok(ctx, outputPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 22. premiere_export_for_twitter
	s.AddTool(
		gomcp.NewTool("premiere_export_for_twitter",
			gomcp.WithDescription("Export for Twitter with optimized settings (16:9, 1920x1080, H.264, <2:20 duration, <512MB)."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the output file"),
			),
		),
		compoundH(logger, "export_for_twitter", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ExportForTwitter(ctx, outputPath)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 23. premiere_export_multiple_formats
	s.AddTool(
		gomcp.NewTool("premiere_export_multiple_formats",
			gomcp.WithDescription("Export the active sequence in multiple platform-specific formats at once. Queues exports for YouTube, Instagram, TikTok, and Twitter."),
			gomcp.WithString("output_dir",
				gomcp.Required(),
				gomcp.Description("Absolute path to the output directory"),
			),
			gomcp.WithArray("formats",
				gomcp.Required(),
				gomcp.Description("Array of format names: youtube, instagram_square, instagram_portrait, tiktok, twitter"),
				gomcp.WithStringItems(),
			),
		),
		compoundH(logger, "export_multiple_formats", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputDir := gomcp.ParseString(req, "output_dir", "")
			if outputDir == "" {
				return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
			}
			formats, err := extractStringSlice(req, "formats")
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("invalid formats: %v", err)), nil
			}
			result, err := orch.ExportMultipleFormats(ctx, outputDir, formats)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Project Setup (24-27)
	// -----------------------------------------------------------------------

	// 24. premiere_setup_new_project
	s.AddTool(
		gomcp.NewTool("premiere_setup_new_project",
			gomcp.WithDescription("Full project setup: create project with specified resolution, frame rate, and audio sample rate."),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Project name"),
			),
			gomcp.WithString("path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the project file"),
			),
			gomcp.WithString("resolution",
				gomcp.Description("Project resolution"),
				gomcp.Enum("1080p", "4k"),
			),
			gomcp.WithNumber("fps",
				gomcp.Description("Frame rate (default: 24)"),
			),
			gomcp.WithNumber("audio_sample_rate",
				gomcp.Description("Audio sample rate in Hz (default: 48000)"),
			),
		),
		compoundH(logger, "setup_new_project", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			name := gomcp.ParseString(req, "name", "")
			if name == "" {
				return gomcp.NewToolResultError("parameter 'name' is required"), nil
			}
			path := gomcp.ParseString(req, "path", "")
			if path == "" {
				return gomcp.NewToolResultError("parameter 'path' is required"), nil
			}
			result, err := orch.SetupNewProject(ctx, name, path,
				gomcp.ParseString(req, "resolution", "1080p"),
				gomcp.ParseFloat64(req, "fps", 24),
				gomcp.ParseInt(req, "audio_sample_rate", 48000),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 25. premiere_setup_editing_workspace
	s.AddTool(
		gomcp.NewTool("premiere_setup_editing_workspace",
			gomcp.WithDescription("Complete workspace setup: open project, import media folder, create initial sequence, and set workspace layout."),
			gomcp.WithString("project_path",
				gomcp.Required(),
				gomcp.Description("Absolute path to the project file"),
			),
			gomcp.WithString("media_folder",
				gomcp.Description("Absolute path to the media folder to import"),
			),
			gomcp.WithString("sequence_name",
				gomcp.Description("Name for the initial sequence"),
			),
		),
		compoundH(logger, "setup_editing_workspace", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			projectPath := gomcp.ParseString(req, "project_path", "")
			if projectPath == "" {
				return gomcp.NewToolResultError("parameter 'project_path' is required"), nil
			}
			result, err := orch.SetupEditingWorkspace(ctx, projectPath,
				gomcp.ParseString(req, "media_folder", ""),
				gomcp.ParseString(req, "sequence_name", ""),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 26. premiere_import_and_organize
	s.AddTool(
		gomcp.NewTool("premiere_import_and_organize",
			gomcp.WithDescription("Import all media from a folder and auto-organize into bins by type (Video, Audio, Images)."),
			gomcp.WithString("media_folder",
				gomcp.Required(),
				gomcp.Description("Absolute path to the media folder"),
			),
			gomcp.WithBoolean("auto_create_bins",
				gomcp.Description("Whether to auto-create bins by media type (default: true)"),
			),
		),
		compoundH(logger, "import_and_organize", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			mediaFolder := gomcp.ParseString(req, "media_folder", "")
			if mediaFolder == "" {
				return gomcp.NewToolResultError("parameter 'media_folder' is required"), nil
			}
			result, err := orch.ImportAndOrganize(ctx, mediaFolder,
				gomcp.ParseBoolean(req, "auto_create_bins", true),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 27. premiere_prepare_for_delivery
	s.AddTool(
		gomcp.NewTool("premiere_prepare_for_delivery",
			gomcp.WithDescription("Check project against delivery specifications and report readiness. Validates resolution, duration, and format requirements."),
			gomcp.WithString("specs",
				gomcp.Required(),
				gomcp.Description("JSON object with delivery specs: {\"width\":1920,\"height\":1080,\"fps\":24,\"maxDuration\":3600}"),
			),
		),
		compoundH(logger, "prepare_for_delivery", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			specs := gomcp.ParseString(req, "specs", "")
			if specs == "" {
				return gomcp.NewToolResultError("parameter 'specs' is required"), nil
			}
			result, err := orch.PrepareForDelivery(ctx, specs)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// -----------------------------------------------------------------------
	// Cleanup Workflow (28-30)
	// -----------------------------------------------------------------------

	// 28. premiere_archive_project
	s.AddTool(
		gomcp.NewTool("premiere_archive_project",
			gomcp.WithDescription("Archive the current project with optional media and render files. Prepares a complete project archive."),
			gomcp.WithString("output_path",
				gomcp.Required(),
				gomcp.Description("Absolute path for the archive output"),
			),
			gomcp.WithBoolean("include_media",
				gomcp.Description("Include media files in archive (default: true)"),
			),
			gomcp.WithBoolean("include_renders",
				gomcp.Description("Include rendered files in archive (default: true)"),
			),
		),
		compoundH(logger, "archive_project", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputPath := gomcp.ParseString(req, "output_path", "")
			if outputPath == "" {
				return gomcp.NewToolResultError("parameter 'output_path' is required"), nil
			}
			result, err := orch.ArchiveProject(ctx, outputPath,
				gomcp.ParseBoolean(req, "include_media", true),
				gomcp.ParseBoolean(req, "include_renders", true),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 29. premiere_trim_project
	s.AddTool(
		gomcp.NewTool("premiere_trim_project",
			gomcp.WithDescription("Analyze and remove unused media from the project. Reports used vs unused items to help clean up the project."),
		),
		compoundH(logger, "trim_project", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			result, err := orch.TrimProject(ctx)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)

	// 30. premiere_consolidate_and_transcode
	s.AddTool(
		gomcp.NewTool("premiere_consolidate_and_transcode",
			gomcp.WithDescription("Consolidate all media used in the active sequence and prepare for transcoding to a single codec."),
			gomcp.WithString("output_dir",
				gomcp.Required(),
				gomcp.Description("Absolute path to the output directory for consolidated media"),
			),
			gomcp.WithString("codec",
				gomcp.Description("Target codec (default: h264)"),
				gomcp.Enum("h264", "prores", "dnxhd", "prores_4444"),
			),
			gomcp.WithString("quality",
				gomcp.Description("Quality preset"),
				gomcp.Enum("low", "medium", "high"),
			),
		),
		compoundH(logger, "consolidate_and_transcode", func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
			outputDir := gomcp.ParseString(req, "output_dir", "")
			if outputDir == "" {
				return gomcp.NewToolResultError("parameter 'output_dir' is required"), nil
			}
			result, err := orch.ConsolidateAndTranscode(ctx, outputDir,
				gomcp.ParseString(req, "codec", "h264"),
				gomcp.ParseString(req, "quality", "medium"),
			)
			if err != nil {
				return gomcp.NewToolResultError(fmt.Sprintf("failed: %v", err)), nil
			}
			return toolResultJSON(result)
		}),
	)
}

// compoundH wraps a handler with debug logging for compound tool invocations.
func compoundH(logger *zap.Logger, name string, fn server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling premiere_" + name)
		return fn(ctx, req)
	}
}
