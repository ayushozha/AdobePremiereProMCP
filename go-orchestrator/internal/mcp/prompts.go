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
	if hasRegisteredTools(s,
		"premiere_scan_assets", "premiere_is_running", "premiere_open",
		"premiere_create_sequence", "premiere_import_media", "premiere_parse_script",
		"premiere_place_clip", "premiere_set_audio_level", "premiere_get_timeline",
	) {
		addValidatedPrompt(s,
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
	}

	if hasRegisteredTools(s,
		"premiere_get_timeline", "premiere_lumetri_get_all",
		"premiere_lumetri_set_contrast", "premiere_lumetri_set_shadows",
		"premiere_lumetri_set_highlights", "premiere_lumetri_set_temperature",
		"premiere_lumetri_set_tint", "premiere_lumetri_set_saturation",
		"premiere_lumetri_set_vibrance", "premiere_lumetri_set_blacks",
		"premiere_lumetri_set_whites", "premiere_lumetri_apply_lut",
	) {
		addValidatedPrompt(s,
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
	}

	if hasRegisteredTools(s, "premiere_get_timeline", "premiere_export") {
		addValidatedPrompt(s,
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
					gomcp.ArgumentDescription("Sequence ID to export. If omitted, the workflow first resolves the active sequence ID."),
				),
			),
			handleSocialExportPrompt,
		)
	}

	if hasRegisteredTools(s,
		"premiere_get_timeline", "premiere_normalize_audio",
		"premiere_set_audio_level", "premiere_apply_audio_effect",
		"premiere_get_audio_mixer_state",
	) {
		addValidatedPrompt(s,
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
	}

	if hasRegisteredTools(s,
		"premiere_get_timeline", "premiere_import_mogrt",
		"premiere_get_mogrt_properties", "premiere_set_mogrt_text",
	) {
		addValidatedPrompt(s,
			gomcp.NewPrompt("add-titles",
				gomcp.WithPromptDescription("Add titles and lower thirds from a verified Motion Graphics Template"),
				gomcp.WithArgument("mogrt_path",
					gomcp.ArgumentDescription("Absolute path to the .mogrt template to place"),
					gomcp.RequiredArgument(),
				),
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
}

func hasRegisteredTools(s *server.MCPServer, names ...string) bool {
	tools := s.ListTools()
	for _, name := range names {
		if tools[name] == nil {
			return false
		}
	}
	return true
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

	var instructions strings.Builder
	fmt.Fprintf(&instructions, `Export a sequence for %s.

Step-by-step workflow:

`, platform)

	resolvedSequenceID := sequenceID
	if resolvedSequenceID == "" {
		resolvedSequenceID = "<sequence_id returned by premiere_get_timeline>"
		instructions.WriteString(`1. RESOLVE AND INSPECT THE ACTIVE SEQUENCE
   - Call premiere_get_timeline with an empty argument object.
   - Copy the returned sequenceId/sequence_id exactly. Stop if there is no active sequence.
   - Review its resolution, frame rate, duration, captions, and safe-area needs.
`)
	} else {
		fmt.Fprintf(&instructions, `1. INSPECT THE REQUESTED SEQUENCE
   - Call premiere_get_timeline with sequence_id: %q.
   - Confirm the response identifies the same sequence before exporting.
   - Review its resolution, frame rate, duration, captions, and safe-area needs.
`, resolvedSequenceID)
	}

	aspectGuidance := "confirm the destination's current aspect-ratio and delivery requirements before export"
	switch strings.ToLower(platform) {
	case "youtube":
		aspectGuidance = "normally preserve the sequence's 16:9 presentation unless the requested YouTube format is vertical"
	case "instagram":
		aspectGuidance = "confirm whether the deliverable is Reel/Story (9:16), portrait feed (4:5), or square feed (1:1)"
	case "tiktok":
		aspectGuidance = "confirm a 9:16 vertical sequence and safe placement of captions and graphics"
	case "twitter", "x":
		aspectGuidance = "confirm the requested X placement and aspect ratio instead of assuming a fixed duration or file-size limit"
	case "linkedin":
		aspectGuidance = "confirm the requested LinkedIn placement and aspect ratio instead of assuming a fixed duration or file-size limit"
	}

	fmt.Fprintf(&instructions, `
2. CONFIRM THE DELIVERABLE
   - %s.
   - Do not silently resize or reframe the sequence. If its geometry is wrong, report that and use the dedicated reframing workflow first.
   - Platform upload limits change; verify current platform documentation when limits matter.

3. EXPORT
   - Choose h264_4k only for a verified 4K deliverable; otherwise choose h264_1080p.
   - These names are aliases for administrator-configured .epr files. The matching
     PREMIERE_EXPORT_PRESET_H264_4K or PREMIERE_EXPORT_PRESET_H264_1080P path must exist.
   - Call premiere_export with every required argument:
     sequence_id: %s
     output_path: %s/<project_name>_%s.mp4
     preset: <the verified configured alias>
   - If the alias is not configured, stop and report the missing preset mapping; do not claim an export occurred.

4. VERIFY
   - Require a successful tool result and report its real job/status and output path.
   - Check that the output file exists when filesystem access is available.
   - Do not invent or estimate a file size.

FINAL STEP: Report the sequence ID, configured preset alias, actual status, and output path.
`, aspectGuidance, resolvedSequenceID, outputDir, strings.ToLower(platform))

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
   - Use premiere_get_audio_mixer_state to review the final mix state
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
	mogrtPath := req.Params.Arguments["mogrt_path"]
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
	fmt.Fprintf(&instructions, `Add titles and lower thirds to %s using the supplied Motion Graphics Template.

Step-by-step workflow:

1. INSPECT TIMELINE
   - Use premiere_get_timeline to see the current sequence state
   - Record the current clips so the newly imported MOGRT can be identified

2. PLACE AND INSPECT THE TEMPLATE
   - Use premiere_import_mogrt with mogrt_path: %q
   - Re-read the timeline and identify the newly inserted MOGRT clip
   - Use premiere_get_mogrt_properties on that exact clip
   - Stop if the import is not visible in the timeline or no editable text property is exposed

3. SET AND VERIFY THE MAIN TITLE
   - Find the text property index returned by premiere_get_mogrt_properties
   - Use premiere_set_mogrt_text to set it to %q
   - Read the MOGRT properties again and require the returned value to match exactly
`, seqRef, mogrtPath, titleText)

	if style != "" {
		fmt.Fprintf(&instructions, `
   Requested visual style: %q
   - Change style properties only when the template exposes a clearly named matching property
   - Use premiere_set_mogrt_property and read the properties back after each change
   - Do not invent font, color, or layout controls that the template does not expose
`, style)
	}

	if subtitleText != "" {
		fmt.Fprintf(&instructions, `
4. OPTIONAL SUBTITLE/TAGLINE
   - If the same MOGRT exposes a second text property, set it to %q and read it back
   - Otherwise stop and request a subtitle-capable template; do not create an unverified text clip
`, subtitleText)
	}

	if lowerThirds != "" {
		fmt.Fprintf(&instructions, `
5. OPTIONAL LOWER THIRDS
   Requested entries: %q
   - Require an explicit timeline position for every entry (or a named marker that supplies it)
   - Import a fresh copy of the supplied MOGRT for each entry, set only exposed text properties, and read each copy back
   - If the template lacks separate name/title fields, report that limitation before placing anything
`, lowerThirds)
	}

	instructions.WriteString(`
FINAL STEP: Use premiere_get_timeline and premiere_get_mogrt_properties to
verify every placed title. Report the track, clip index, timing, template path,
and exact text values. A successful import alone is not proof that text changed.
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
