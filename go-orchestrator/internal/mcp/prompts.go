package mcp

import (
	"context"
	"fmt"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerPrompts registers all MCP prompts with the server.
// Prompts are reusable workflow templates that guide the AI assistant
// through common video editing tasks step by step.
func registerPrompts(s *server.MCPServer) {
	s.AddPrompt(
		gomcp.NewPrompt("rough-cut",
			gomcp.WithPromptDescription("Create a rough cut from raw footage"),
			gomcp.WithArgument("footage_path",
				gomcp.ArgumentDescription("Absolute path to the directory containing raw footage"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("project_name",
				gomcp.ArgumentDescription("Name for the new project/sequence"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("script",
				gomcp.ArgumentDescription("Script or shot list to guide the edit (optional)"),
			),
			gomcp.WithArgument("duration_target",
				gomcp.ArgumentDescription("Target duration in minutes, e.g. '5' (optional)"),
			),
		),
		handleRoughCutPrompt,
	)

	s.AddPrompt(
		gomcp.NewPrompt("color-grade",
			gomcp.WithPromptDescription("Apply color grading to a sequence"),
			gomcp.WithArgument("style",
				gomcp.ArgumentDescription("Color grading style: cinematic, warm, cool, desaturated, vintage, high-contrast"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("sequence_id",
				gomcp.ArgumentDescription("Sequence ID to grade (defaults to active sequence)"),
			),
			gomcp.WithArgument("lut_path",
				gomcp.ArgumentDescription("Path to a .cube LUT file to apply (optional)"),
			),
		),
		handleColorGradePrompt,
	)

	s.AddPrompt(
		gomcp.NewPrompt("social-export",
			gomcp.WithPromptDescription("Export for social media platforms"),
			gomcp.WithArgument("platform",
				gomcp.ArgumentDescription("Target platform: youtube, instagram, tiktok, twitter, linkedin"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("output_directory",
				gomcp.ArgumentDescription("Directory to save exported files"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("sequence_id",
				gomcp.ArgumentDescription("Sequence ID to export (defaults to active sequence)"),
			),
		),
		handleSocialExportPrompt,
	)

	s.AddPrompt(
		gomcp.NewPrompt("audio-mix",
			gomcp.WithPromptDescription("Mix and master audio for a sequence"),
			gomcp.WithArgument("mix_type",
				gomcp.ArgumentDescription("Type of mix: dialogue, music-video, podcast, documentary, commercial"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("sequence_id",
				gomcp.ArgumentDescription("Sequence ID to mix (defaults to active sequence)"),
			),
			gomcp.WithArgument("loudness_standard",
				gomcp.ArgumentDescription("Loudness standard: broadcast (-24 LUFS), streaming (-14 LUFS), podcast (-16 LUFS)"),
			),
		),
		handleAudioMixPrompt,
	)

	s.AddPrompt(
		gomcp.NewPrompt("add-titles",
			gomcp.WithPromptDescription("Add titles and lower thirds to a sequence"),
			gomcp.WithArgument("title_text",
				gomcp.ArgumentDescription("Main title text to display"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("style",
				gomcp.ArgumentDescription("Title style: minimal, bold, cinematic, news, corporate"),
				gomcp.RequiredArgument(),
			),
			gomcp.WithArgument("sequence_id",
				gomcp.ArgumentDescription("Sequence ID (defaults to active sequence)"),
			),
			gomcp.WithArgument("subtitle_text",
				gomcp.ArgumentDescription("Subtitle or tagline text (optional)"),
			),
			gomcp.WithArgument("lower_thirds",
				gomcp.ArgumentDescription("Comma-separated list of lower third entries as 'name|title' pairs, e.g. 'John Doe|CEO,Jane Smith|CTO'"),
			),
		),
		handleAddTitlesPrompt,
	)
}

// ---------------------------------------------------------------------------
// Prompt handlers
// ---------------------------------------------------------------------------

func handleRoughCutPrompt(
	_ context.Context,
	req gomcp.GetPromptRequest,
) (*gomcp.GetPromptResult, error) {
	footagePath := req.Params.Arguments["footage_path"]
	projectName := req.Params.Arguments["project_name"]
	script := req.Params.Arguments["script"]
	durationTarget := req.Params.Arguments["duration_target"]

	var instructions strings.Builder
	fmt.Fprintf(&instructions, `Create a rough cut from raw footage for project "%s".

Step-by-step workflow:

1. SCAN FOOTAGE
   Use premiere_scan_assets to scan the footage directory:
   - Directory: %s
   - Look at the returned metadata to understand what footage is available
   - Note file types, durations, and names

2. SET UP PROJECT
   - Use premiere_is_running to check if Premiere Pro is running; launch with premiere_open if not
   - Use premiere_create_sequence to create a new sequence named "%s"
   - Use recommended settings: 1920x1080, 24fps unless footage suggests otherwise

3. IMPORT MEDIA
   - Use premiere_import_media to import all relevant footage files
   - Organize into bins if there are many files
`, projectName, footagePath, projectName)

	if script != "" {
		fmt.Fprintf(&instructions, `
4. PARSE SCRIPT
   - Use premiere_parse_script with the following script/shot list to guide edit order:
   ---
   %s
   ---
   - Match script segments to scanned footage

`, script)
		instructions.WriteString("5. ASSEMBLE ROUGH CUT\n")
	} else {
		instructions.WriteString("\n4. ASSEMBLE ROUGH CUT\n")
	}

	instructions.WriteString(`   - Use premiere_place_clip to lay clips on the timeline in order
   - Place primary footage on video track 0
   - Place B-roll on video track 1
   - Add cross_dissolve transitions between major segments
   - Set audio levels appropriately with premiere_set_audio_level
`)

	if durationTarget != "" {
		fmt.Fprintf(&instructions, `
   TARGET DURATION: %s minutes
   - Trim clips to fit within the target duration
   - Prioritize the strongest footage
`, durationTarget)
	}

	fmt.Fprintf(&instructions, `
FINAL STEP: Use premiere_get_timeline to review the assembled sequence
and report what was created.
`)

	return &gomcp.GetPromptResult{
		Description: fmt.Sprintf("Rough cut workflow for %q", projectName),
		Messages: []gomcp.PromptMessage{
			{
				Role:    gomcp.RoleUser,
				Content: gomcp.NewTextContent(instructions.String()),
			},
		},
	}, nil
}

func handleColorGradePrompt(
	_ context.Context,
	req gomcp.GetPromptRequest,
) (*gomcp.GetPromptResult, error) {
	style := req.Params.Arguments["style"]
	sequenceID := req.Params.Arguments["sequence_id"]
	lutPath := req.Params.Arguments["lut_path"]

	seqRef := "the active sequence"
	if sequenceID != "" {
		seqRef = fmt.Sprintf("sequence %s", sequenceID)
	}

	var instructions strings.Builder
	fmt.Fprintf(&instructions, `Apply %q color grading to %s.

Step-by-step workflow:

1. INSPECT TIMELINE
   - Use premiere_get_timeline to see all clips on the sequence
   - Note which video tracks have clips and how many clips there are

2. ANALYZE CURRENT GRADE
   - For each clip, use premiere_lumetri_get_all to see current color values
   - Identify if any clips already have grading applied

3. APPLY COLOR GRADE
   Apply the "%s" look to all video clips:
`, style, seqRef, style)

	switch strings.ToLower(style) {
	case "cinematic":
		instructions.WriteString(`
   Cinematic look settings per clip:
   - premiere_lumetri_set_contrast: 15 to 25
   - premiere_lumetri_set_shadows: -10 to -20 (crush blacks slightly)
   - premiere_lumetri_set_highlights: -5 to -15 (roll off highlights)
   - premiere_lumetri_set_temperature: slight warm shift (5 to 10)
   - premiere_lumetri_set_saturation: 85 to 95 (slightly desaturated)
   - premiere_lumetri_set_vibrance: 10 to 20
`)
	case "warm":
		instructions.WriteString(`
   Warm look settings per clip:
   - premiere_lumetri_set_temperature: 15 to 25 (shift warm)
   - premiere_lumetri_set_tint: 5 to 10 (slight magenta)
   - premiere_lumetri_set_highlights: 5 to 10 (lift highlights)
   - premiere_lumetri_set_saturation: 105 to 115
   - premiere_lumetri_set_vibrance: 15 to 25
`)
	case "cool":
		instructions.WriteString(`
   Cool look settings per clip:
   - premiere_lumetri_set_temperature: -15 to -25 (shift cool)
   - premiere_lumetri_set_tint: -5 to -10 (slight green)
   - premiere_lumetri_set_contrast: 10 to 15
   - premiere_lumetri_set_saturation: 90 to 100
`)
	case "desaturated":
		instructions.WriteString(`
   Desaturated look settings per clip:
   - premiere_lumetri_set_saturation: 40 to 60
   - premiere_lumetri_set_contrast: 15 to 25
   - premiere_lumetri_set_shadows: -10 to -15
   - premiere_lumetri_set_highlights: -5 to -10
`)
	case "vintage":
		instructions.WriteString(`
   Vintage look settings per clip:
   - premiere_lumetri_set_temperature: 10 to 15
   - premiere_lumetri_set_tint: 5 to 10
   - premiere_lumetri_set_saturation: 75 to 85
   - premiere_lumetri_set_blacks: 5 to 15 (lift blacks / faded look)
   - premiere_lumetri_set_contrast: -5 to 5 (reduce contrast)
   - premiere_lumetri_set_highlights: -10 to -20
`)
	case "high-contrast":
		instructions.WriteString(`
   High contrast look settings per clip:
   - premiere_lumetri_set_contrast: 30 to 50
   - premiere_lumetri_set_shadows: -15 to -25
   - premiere_lumetri_set_highlights: 10 to 20
   - premiere_lumetri_set_blacks: -10 to -15
   - premiere_lumetri_set_whites: 10 to 15
   - premiere_lumetri_set_saturation: 105 to 115
`)
	default:
		fmt.Fprintf(&instructions, `
   For the "%s" style, use your judgment to set appropriate Lumetri values.
   Use premiere_lumetri_set_* tools for exposure, contrast, highlights,
   shadows, temperature, tint, saturation, and vibrance.
`, style)
	}

	if lutPath != "" {
		fmt.Fprintf(&instructions, `
4. APPLY LUT
   - Use premiere_lumetri_apply_lut with path: %s
   - Apply to all clips after the base grade is set
`, lutPath)
	}

	instructions.WriteString(`
FINAL STEP: Verify the grade by checking premiere_lumetri_get_all on
a few clips and report the applied settings.
`)

	return &gomcp.GetPromptResult{
		Description: fmt.Sprintf("Color grading workflow: %s style", style),
		Messages: []gomcp.PromptMessage{
			{
				Role:    gomcp.RoleUser,
				Content: gomcp.NewTextContent(instructions.String()),
			},
		},
	}, nil
}

func handleSocialExportPrompt(
	_ context.Context,
	req gomcp.GetPromptRequest,
) (*gomcp.GetPromptResult, error) {
	platform := req.Params.Arguments["platform"]
	outputDir := req.Params.Arguments["output_directory"]
	sequenceID := req.Params.Arguments["sequence_id"]

	seqRef := "the active sequence"
	if sequenceID != "" {
		seqRef = fmt.Sprintf("sequence %s", sequenceID)
	}

	var instructions strings.Builder
	fmt.Fprintf(&instructions, `Export %s for %s.

Step-by-step workflow:

1. INSPECT SEQUENCE
   - Use premiere_get_timeline to review the current sequence
   - Note the sequence duration and resolution
`, seqRef, platform)

	fmt.Fprintf(&instructions, `
2. EXPORT WITH PLATFORM SETTINGS
   Export to: %s
`, outputDir)

	switch strings.ToLower(platform) {
	case "youtube":
		fmt.Fprintf(&instructions, `
   YouTube recommended settings:
   - Format: H.264 (MP4)
   - Resolution: 1920x1080 or 3840x2160
   - Frame rate: match source (typically 24, 30, or 60 fps)
   - Bitrate: 15-20 Mbps for 1080p, 45-50 Mbps for 4K

   Use premiere_export with:
   - output_path: %s/<project_name>_youtube.mp4
   - preset: h264_1080p (or h264_4k for 4K content)
`, outputDir)

	case "instagram":
		fmt.Fprintf(&instructions, `
   Instagram recommended settings:
   - Feed video: 1080x1080 (1:1) or 1080x1350 (4:5), max 60s
   - Reels: 1080x1920 (9:16), max 90s
   - Stories: 1080x1920 (9:16), max 60s
   - Format: H.264 MP4
   - Bitrate: 10-15 Mbps

   Consider the content type and export accordingly:
   - output_path: %s/<project_name>_instagram.mp4
   - preset: h264_1080p

   NOTE: If the sequence is not in the correct aspect ratio,
   recommend creating a new sequence with the proper dimensions
   and re-editing.
`, outputDir)

	case "tiktok":
		fmt.Fprintf(&instructions, `
   TikTok recommended settings:
   - Resolution: 1080x1920 (9:16 vertical)
   - Frame rate: 30 fps
   - Duration: 15s to 3 min (optimal: 15-60s)
   - Format: H.264 MP4
   - Bitrate: 10-15 Mbps

   Use premiere_export with:
   - output_path: %s/<project_name>_tiktok.mp4
   - preset: h264_1080p

   NOTE: If sequence is 16:9, recommend creating a 9:16 sequence
   (1080x1920) and repositioning footage.
`, outputDir)

	case "twitter":
		fmt.Fprintf(&instructions, `
   Twitter/X recommended settings:
   - Resolution: 1280x720 or 1920x1080
   - Frame rate: 30 or 60 fps
   - Duration: max 2 min 20 seconds (140s)
   - Format: H.264 MP4
   - Max file size: 512 MB

   Use premiere_export with:
   - output_path: %s/<project_name>_twitter.mp4
   - preset: h264_1080p
`, outputDir)

	case "linkedin":
		fmt.Fprintf(&instructions, `
   LinkedIn recommended settings:
   - Resolution: 1920x1080 or 1280x720
   - Frame rate: 30 fps
   - Duration: 3 seconds to 10 minutes (optimal: 1-2 min)
   - Format: H.264 MP4
   - Max file size: 5 GB

   Use premiere_export with:
   - output_path: %s/<project_name>_linkedin.mp4
   - preset: h264_1080p
`, outputDir)

	default:
		fmt.Fprintf(&instructions, `
   For %s, use general web-optimized settings:
   - Format: H.264 MP4
   - Resolution: 1920x1080
   - Bitrate: 15-20 Mbps

   Use premiere_export with:
   - output_path: %s/<project_name>_%s.mp4
   - preset: h264_1080p
`, platform, outputDir, platform)
	}

	instructions.WriteString(`
3. VERIFY EXPORT
   - Check that the export completed successfully
   - Report the output file path and estimated file size

FINAL STEP: Report the exported file details.
`)

	return &gomcp.GetPromptResult{
		Description: fmt.Sprintf("Social media export workflow for %s", platform),
		Messages: []gomcp.PromptMessage{
			{
				Role:    gomcp.RoleUser,
				Content: gomcp.NewTextContent(instructions.String()),
			},
		},
	}, nil
}

func handleAudioMixPrompt(
	_ context.Context,
	req gomcp.GetPromptRequest,
) (*gomcp.GetPromptResult, error) {
	mixType := req.Params.Arguments["mix_type"]
	sequenceID := req.Params.Arguments["sequence_id"]
	loudnessStd := req.Params.Arguments["loudness_standard"]

	seqRef := "the active sequence"
	if sequenceID != "" {
		seqRef = fmt.Sprintf("sequence %s", sequenceID)
	}

	if loudnessStd == "" {
		switch strings.ToLower(mixType) {
		case "podcast":
			loudnessStd = "-16 LUFS"
		case "commercial":
			loudnessStd = "-24 LUFS"
		case "documentary", "dialogue":
			loudnessStd = "-24 LUFS"
		default:
			loudnessStd = "-14 LUFS (streaming)"
		}
	}

	var instructions strings.Builder
	fmt.Fprintf(&instructions, `Mix and master audio for %s.
Mix type: %s | Target loudness: %s

Step-by-step workflow:

1. INSPECT TIMELINE
   - Use premiere_get_timeline to identify all audio tracks and clips
   - Categorize tracks: dialogue, music, SFX, ambient

2. SET BASE LEVELS
`, seqRef, mixType, loudnessStd)

	switch strings.ToLower(mixType) {
	case "dialogue":
		instructions.WriteString(`
   Dialogue-focused mix levels:
   - Dialogue tracks: -6 dB to -3 dB (primary)
   - Music tracks: -18 dB to -24 dB (well under dialogue)
   - SFX tracks: -12 dB to -18 dB
   - Ambient/room tone: -24 dB to -30 dB

   For each dialogue clip:
   - Use premiere_normalize_audio to normalize
   - Use premiere_set_audio_level to fine-tune
   - Apply noise reduction if needed via premiere_apply_audio_effect
`)
	case "music-video":
		instructions.WriteString(`
   Music video mix levels:
   - Music tracks: -3 dB to 0 dB (primary)
   - Vocal tracks: -6 dB to -9 dB
   - SFX tracks: -12 dB to -18 dB

   For the music track:
   - Use premiere_set_audio_level to set as primary
   - Ensure it drives the overall loudness
`)
	case "podcast":
		instructions.WriteString(`
   Podcast mix levels:
   - Host voice: -6 dB to -3 dB
   - Guest voice(s): -6 dB to -3 dB (match host level)
   - Music (intro/outro/beds): -20 dB to -30 dB
   - SFX/stingers: -12 dB to -15 dB

   For each voice track:
   - Use premiere_normalize_audio
   - Apply compression via premiere_apply_audio_effect
   - Use premiere_set_audio_level to balance voices
`)
	case "documentary":
		instructions.WriteString(`
   Documentary mix levels:
   - Narration/interview: -6 dB to -3 dB
   - Natural sound/ambient: -18 dB to -24 dB
   - Music underscore: -20 dB to -27 dB
   - SFX: -12 dB to -18 dB

   For narration/interview tracks:
   - Use premiere_normalize_audio
   - Apply EQ to improve clarity
   - Use premiere_set_audio_level to balance
`)
	case "commercial":
		instructions.WriteString(`
   Commercial mix levels:
   - Voiceover: -6 dB to -3 dB
   - Music: -15 dB to -20 dB
   - SFX: -9 dB to -15 dB

   Important: commercials must meet broadcast loudness
   standards (-24 LUFS typically).
   - Use premiere_normalize_audio on all tracks
   - Use premiere_set_audio_level for final balance
`)
	default:
		fmt.Fprintf(&instructions, `
   For %s mix type, use balanced levels:
   - Primary audio: -6 dB to -3 dB
   - Secondary audio: -12 dB to -18 dB
   - Background: -20 dB to -30 dB
`, mixType)
	}

	fmt.Fprintf(&instructions, `
3. APPLY AUDIO EFFECTS
   - Use premiere_apply_audio_effect for EQ, compression, and limiting
   - Consider noise reduction for dialogue tracks
   - Add a limiter on the master to prevent clipping

4. VERIFY MIX
   - Use premiere_get_audio_mix to review the final mix state
   - Target loudness: %s
   - Ensure no clipping (peaks should not exceed -1 dB)

FINAL STEP: Report the audio mix settings applied to each track.
`, loudnessStd)

	return &gomcp.GetPromptResult{
		Description: fmt.Sprintf("Audio mix workflow: %s mix", mixType),
		Messages: []gomcp.PromptMessage{
			{
				Role:    gomcp.RoleUser,
				Content: gomcp.NewTextContent(instructions.String()),
			},
		},
	}, nil
}

func handleAddTitlesPrompt(
	_ context.Context,
	req gomcp.GetPromptRequest,
) (*gomcp.GetPromptResult, error) {
	titleText := req.Params.Arguments["title_text"]
	style := req.Params.Arguments["style"]
	sequenceID := req.Params.Arguments["sequence_id"]
	subtitleText := req.Params.Arguments["subtitle_text"]
	lowerThirds := req.Params.Arguments["lower_thirds"]

	seqRef := "the active sequence"
	if sequenceID != "" {
		seqRef = fmt.Sprintf("sequence %s", sequenceID)
	}

	var instructions strings.Builder
	fmt.Fprintf(&instructions, `Add titles and lower thirds to %s.

Step-by-step workflow:

1. INSPECT TIMELINE
   - Use premiere_get_timeline to see the current sequence state
   - Identify available video tracks for title placement
   - Note the sequence duration for timing

2. ADD MAIN TITLE
   Use premiere_add_text with:
   - text: "%s"
   - Position at the beginning of the sequence (position_seconds: 0)
   - Duration: 4-5 seconds
`, seqRef, titleText)

	switch strings.ToLower(style) {
	case "minimal":
		instructions.WriteString(`
   Minimal style:
   - font_size: 60
   - color: #FFFFFF (white)
   - x: 0.5 (centered)
   - y: 0.45 (slightly above center)
   - Clean, simple look — no background or effects needed
`)
	case "bold":
		instructions.WriteString(`
   Bold style:
   - font_size: 96
   - color: #FFFFFF (white)
   - x: 0.5 (centered)
   - y: 0.5 (centered)
   - Consider adding a drop shadow or outline effect
`)
	case "cinematic":
		instructions.WriteString(`
   Cinematic style:
   - font_size: 72
   - color: #F5F5DC (warm off-white)
   - x: 0.5 (centered)
   - y: 0.55 (slightly below center)
   - Add a subtle fade-in transition on the text clip
   - Consider letter spacing if available
`)
	case "news":
		instructions.WriteString(`
   News style:
   - font_size: 48
   - color: #FFFFFF (white)
   - x: 0.5 (centered)
   - y: 0.15 (upper area)
   - Typically uses a colored background bar
`)
	case "corporate":
		instructions.WriteString(`
   Corporate style:
   - font_size: 54
   - color: #333333 (dark gray) or #FFFFFF on dark backgrounds
   - x: 0.5 (centered)
   - y: 0.45
   - Clean, professional look
`)
	default:
		fmt.Fprintf(&instructions, `
   For "%s" style, use appropriate font size (48-72),
   color, and positioning.
`, style)
	}

	if subtitleText != "" {
		fmt.Fprintf(&instructions, `
3. ADD SUBTITLE/TAGLINE
   Use premiere_add_text with:
   - text: "%s"
   - Place directly after or overlapping with the main title
   - Use a smaller font_size (e.g., 36-42)
   - Position below the main title (y: 0.6)
`, subtitleText)
	}

	if lowerThirds != "" {
		instructions.WriteString("\n")
		if subtitleText != "" {
			instructions.WriteString("4. ADD LOWER THIRDS\n")
		} else {
			instructions.WriteString("3. ADD LOWER THIRDS\n")
		}
		instructions.WriteString("   For each person/speaker, add a lower third:\n\n")

		entries := strings.Split(lowerThirds, ",")
		for _, entry := range entries {
			parts := strings.SplitN(strings.TrimSpace(entry), "|", 2)
			if len(parts) == 2 {
				fmt.Fprintf(&instructions, `   - Name: %s, Title: %s
     Use premiere_add_text:
     - text: "%s\n%s"
     - font_size: 36
     - x: 0.2 (left-aligned area)
     - y: 0.85 (lower third position)
     - duration_seconds: 4

`, parts[0], parts[1], parts[0], parts[1])
			} else if len(parts) == 1 {
				fmt.Fprintf(&instructions, `   - Name: %s
     Use premiere_add_text:
     - text: "%s"
     - font_size: 36
     - x: 0.2 (left-aligned area)
     - y: 0.85 (lower third position)
     - duration_seconds: 4

`, parts[0], parts[0])
			}
		}
	}

	instructions.WriteString(`
FINAL STEP: Use premiere_get_timeline to verify all titles were placed
correctly and report the title positions and durations.
`)

	return &gomcp.GetPromptResult{
		Description: fmt.Sprintf("Add titles workflow: %s style", style),
		Messages: []gomcp.PromptMessage{
			{
				Role:    gomcp.RoleUser,
				Content: gomcp.NewTextContent(instructions.String()),
			},
		},
	}, nil
}
