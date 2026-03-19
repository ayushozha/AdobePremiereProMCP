package mcp

import (
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerAudioAdvancedTools registers all advanced audio processing and mixing
// MCP tools. These correspond to ExtendScript functions 1-30 in the advanced
// audio section of premiere.jsx.
func registerAudioAdvancedTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Audio Mixing (1-4)
	// -----------------------------------------------------------------------

	// 1. premiere_get_audio_mixer_state
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_mixer_state",
			gomcp.WithDescription("Get all audio track volumes, panning, mute/solo states from the audio mixer."),
		),
		makeAudioHandler(orch, logger, "getAudioMixerState", nil),
	)

	// 2. premiere_set_track_panning
	s.AddTool(
		gomcp.NewTool("premiere_set_track_panning",
			gomcp.WithDescription("Set track panning (-100 to 100, 0=center)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("pan_value", gomcp.Required(), gomcp.Description("Pan value from -100 (full left) to 100 (full right), 0 is center")),
		),
		makeAudioHandler(orch, logger, "setTrackPanning", []string{"track_index", "pan_value"}),
	)

	// 3. premiere_set_clip_panning
	s.AddTool(
		gomcp.NewTool("premiere_set_clip_panning",
			gomcp.WithDescription("Set clip panning (-100 to 100, 0=center)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("pan_value", gomcp.Required(), gomcp.Description("Pan value from -100 (full left) to 100 (full right), 0 is center")),
		),
		makeAudioHandler(orch, logger, "setClipPanning", []string{"track_index", "clip_index", "pan_value"}),
	)

	// 4. premiere_get_clip_panning
	s.AddTool(
		gomcp.NewTool("premiere_get_clip_panning",
			gomcp.WithDescription("Get the current panning value of an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getClipPanning", []string{"track_index", "clip_index"}),
	)

	// -----------------------------------------------------------------------
	// Audio Keyframes Extended (5-9)
	// -----------------------------------------------------------------------

	// 5. premiere_add_volume_keyframe
	s.AddTool(
		gomcp.NewTool("premiere_add_volume_keyframe",
			gomcp.WithDescription("Add a volume keyframe at a specific time on an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time in seconds (relative to clip start) for the keyframe")),
			gomcp.WithNumber("level_db", gomcp.Required(), gomcp.Description("Audio level in dB (-96 to +15)")),
		),
		makeAudioHandler(orch, logger, "addVolumeKeyframe", []string{"track_index", "clip_index", "time", "level_db"}),
	)

	// 6. premiere_add_panning_keyframe
	s.AddTool(
		gomcp.NewTool("premiere_add_panning_keyframe",
			gomcp.WithDescription("Add a panning keyframe at a specific time on an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time in seconds (relative to clip start) for the keyframe")),
			gomcp.WithNumber("pan_value", gomcp.Required(), gomcp.Description("Pan value from -100 (full left) to 100 (full right)")),
		),
		makeAudioHandler(orch, logger, "addPanningKeyframe", []string{"track_index", "clip_index", "time", "pan_value"}),
	)

	// 7. premiere_get_volume_keyframes
	s.AddTool(
		gomcp.NewTool("premiere_get_volume_keyframes",
			gomcp.WithDescription("Get all volume keyframes for an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getVolumeKeyframes", []string{"track_index", "clip_index"}),
	)

	// 8. premiere_get_panning_keyframes
	s.AddTool(
		gomcp.NewTool("premiere_get_panning_keyframes",
			gomcp.WithDescription("Get all panning keyframes for an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getPanningKeyframes", []string{"track_index", "clip_index"}),
	)

	// 9. premiere_remove_all_audio_keyframes
	s.AddTool(
		gomcp.NewTool("premiere_remove_all_audio_keyframes",
			gomcp.WithDescription("Remove all audio keyframes (volume and panning) from a clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "removeAllAudioKeyframes", []string{"track_index", "clip_index"}),
	)

	// -----------------------------------------------------------------------
	// Audio Routing (10-12)
	// -----------------------------------------------------------------------

	// 10. premiere_set_track_output
	s.AddTool(
		gomcp.NewTool("premiere_set_track_output",
			gomcp.WithDescription("Set the output channel assignment for an audio track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("output_channels", gomcp.Required(), gomcp.Description("Output channel assignment index")),
		),
		makeAudioHandler(orch, logger, "setTrackOutput", []string{"track_index", "output_channels"}),
	)

	// 11. premiere_get_track_output
	s.AddTool(
		gomcp.NewTool("premiere_get_track_output",
			gomcp.WithDescription("Get the output channel assignment for an audio track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
		),
		makeAudioHandler(orch, logger, "getTrackOutput", []string{"track_index"}),
	)

	// 12. premiere_create_submix
	s.AddTool(
		gomcp.NewTool("premiere_create_submix",
			gomcp.WithDescription("Create a submix audio track in the active sequence."),
			gomcp.WithString("name", gomcp.Description("Name for the submix track (default: Submix)")),
			gomcp.WithString("channel_type",
				gomcp.Description("Channel type for the submix track"),
				gomcp.Enum("mono", "stereo", "5.1", "adaptive"),
			),
		),
		makeAudioHandler(orch, logger, "createSubmix", []string{"name", "channel_type"}),
	)

	// -----------------------------------------------------------------------
	// Audio Effects Extended (13-18)
	// -----------------------------------------------------------------------

	// 13. premiere_apply_eq
	s.AddTool(
		gomcp.NewTool("premiere_apply_eq",
			gomcp.WithDescription("Apply a parametric EQ to an audio clip with frequency bands. Each band has freq (Hz), gain (dB), and q (quality factor)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithString("bands", gomcp.Description("JSON array of EQ bands, e.g. '[{\"freq\":1000,\"gain\":3,\"q\":1.0}]'")),
		),
		makeAudioHandler(orch, logger, "applyEQ", []string{"track_index", "clip_index", "bands"}),
	)

	// 14. premiere_apply_compressor
	s.AddTool(
		gomcp.NewTool("premiere_apply_compressor",
			gomcp.WithDescription("Apply a compressor effect to an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("threshold", gomcp.Description("Threshold in dB (default: -20)")),
			gomcp.WithNumber("ratio", gomcp.Description("Compression ratio (default: 4)")),
			gomcp.WithNumber("attack", gomcp.Description("Attack time in ms (default: 10)")),
			gomcp.WithNumber("release", gomcp.Description("Release time in ms (default: 100)")),
		),
		makeAudioHandler(orch, logger, "applyCompressor", []string{"track_index", "clip_index", "threshold", "ratio", "attack", "release"}),
	)

	// 15. premiere_apply_limiter
	s.AddTool(
		gomcp.NewTool("premiere_apply_limiter",
			gomcp.WithDescription("Apply a limiter effect to an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("ceiling", gomcp.Description("Ceiling level in dB (default: -0.1)")),
		),
		makeAudioHandler(orch, logger, "applyLimiter", []string{"track_index", "clip_index", "ceiling"}),
	)

	// 16. premiere_apply_deesser
	s.AddTool(
		gomcp.NewTool("premiere_apply_deesser",
			gomcp.WithDescription("Apply a de-esser effect to an audio clip to reduce sibilance."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("frequency", gomcp.Description("Center frequency in Hz (default: 6000)")),
			gomcp.WithNumber("reduction", gomcp.Description("Reduction amount in dB (default: -10)")),
		),
		makeAudioHandler(orch, logger, "applyDeEsser", []string{"track_index", "clip_index", "frequency", "reduction"}),
	)

	// 17. premiere_get_audio_effect_presets
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_effect_presets",
			gomcp.WithDescription("List all available audio effect presets in this Premiere Pro installation."),
		),
		makeAudioHandler(orch, logger, "getAudioEffectPresets", nil),
	)

	// 18. premiere_apply_audio_preset
	s.AddTool(
		gomcp.NewTool("premiere_apply_audio_preset",
			gomcp.WithDescription("Apply a named audio effect preset to a clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithString("preset_name", gomcp.Required(), gomcp.Description("Name of the audio preset to apply")),
		),
		makeAudioHandler(orch, logger, "applyAudioPreset", []string{"track_index", "clip_index", "preset_name"}),
	)

	// -----------------------------------------------------------------------
	// Audio Analysis Extended (19-23)
	// -----------------------------------------------------------------------

	// 19. premiere_get_audio_waveform_data
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_waveform_data",
			gomcp.WithDescription("Get waveform data points for an audio clip. Returns clip metadata and media path for detailed analysis."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("samples", gomcp.Description("Number of waveform samples to retrieve (default: 100, max: 10000)")),
		),
		makeAudioHandler(orch, logger, "getAudioWaveformData", []string{"track_index", "clip_index", "samples"}),
	)

	// 20. premiere_get_loudness_info
	s.AddTool(
		gomcp.NewTool("premiere_get_loudness_info",
			gomcp.WithDescription("Get loudness information (LUFS) for an audio clip including current level and target loudness."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getLoudnessInfo", []string{"track_index", "clip_index"}),
	)

	// 21. premiere_get_sequence_loudness
	s.AddTool(
		gomcp.NewTool("premiere_get_sequence_loudness",
			gomcp.WithDescription("Get overall sequence loudness information including all audio tracks."),
		),
		makeAudioHandler(orch, logger, "getSequenceLoudness", nil),
	)

	// 22. premiere_find_audio_peaks
	s.AddTool(
		gomcp.NewTool("premiere_find_audio_peaks",
			gomcp.WithDescription("Find audio peaks above a threshold in a clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("threshold_db", gomcp.Description("Peak detection threshold in dB (default: -6)")),
		),
		makeAudioHandler(orch, logger, "findAudioPeaks", []string{"track_index", "clip_index", "threshold_db"}),
	)

	// 23. premiere_detect_clipping
	s.AddTool(
		gomcp.NewTool("premiere_detect_clipping",
			gomcp.WithDescription("Detect audio clipping in a clip by analyzing current gain settings and media."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "detectClipping", []string{"track_index", "clip_index"}),
	)

	// -----------------------------------------------------------------------
	// Voiceover (24-25)
	// -----------------------------------------------------------------------

	// 24. premiere_prepare_voiceover_track
	s.AddTool(
		gomcp.NewTool("premiere_prepare_voiceover_track",
			gomcp.WithDescription("Mute all other audio tracks and prepare a specific track for voiceover recording."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index to prepare for voiceover")),
		),
		makeAudioHandler(orch, logger, "prepareVoiceoverTrack", []string{"track_index"}),
	)

	// 25. premiere_set_voiceover_ducking
	s.AddTool(
		gomcp.NewTool("premiere_set_voiceover_ducking",
			gomcp.WithDescription("Set up audio ducking between a voiceover track and a music track so music automatically lowers when voice is present."),
			gomcp.WithNumber("vo_track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index for voiceover")),
			gomcp.WithNumber("music_track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index for music")),
			gomcp.WithNumber("duck_amount", gomcp.Description("Ducking amount in dB (default: -15)")),
			gomcp.WithNumber("sensitivity", gomcp.Description("Ducking sensitivity 0-100 (default: 50)")),
		),
		makeAudioHandler(orch, logger, "setVoiceoverDucking", []string{"vo_track_index", "music_track_index", "duck_amount", "sensitivity"}),
	)

	// -----------------------------------------------------------------------
	// Audio Sync (26-27)
	// -----------------------------------------------------------------------

	// 26. premiere_sync_audio_to_video
	s.AddTool(
		gomcp.NewTool("premiere_sync_audio_to_video",
			gomcp.WithDescription("Sync an audio clip to a video clip by aligning their start positions. For waveform-based sync, use the media engine."),
			gomcp.WithNumber("audio_track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("audio_clip_index", gomcp.Required(), gomcp.Description("Zero-based audio clip index on the track")),
			gomcp.WithNumber("video_track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithNumber("video_clip_index", gomcp.Required(), gomcp.Description("Zero-based video clip index on the track")),
		),
		makeAudioHandler(orch, logger, "syncAudioToVideo", []string{"audio_track_index", "audio_clip_index", "video_track_index", "video_clip_index"}),
	)

	// 27. premiere_detect_audio_drift
	s.AddTool(
		gomcp.NewTool("premiere_detect_audio_drift",
			gomcp.WithDescription("Detect audio sync drift by comparing an audio clip's position to its linked video clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "detectAudioDrift", []string{"track_index", "clip_index"}),
	)

	// -----------------------------------------------------------------------
	// Channel Operations (28-30)
	// -----------------------------------------------------------------------

	// 28. premiere_convert_stereo_to_mono
	s.AddTool(
		gomcp.NewTool("premiere_convert_stereo_to_mono",
			gomcp.WithDescription("Convert a stereo audio project item to mono by changing its audio channel mapping."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
		),
		makeAudioHandler(orch, logger, "convertStereoToMono", []string{"project_item_index"}),
	)

	// 29. premiere_swap_audio_channels
	s.AddTool(
		gomcp.NewTool("premiere_swap_audio_channels",
			gomcp.WithDescription("Swap left and right audio channels of a project item."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
		),
		makeAudioHandler(orch, logger, "swapAudioChannels", []string{"project_item_index"}),
	)

	// 30. premiere_extract_audio_from_video
	s.AddTool(
		gomcp.NewTool("premiere_extract_audio_from_video",
			gomcp.WithDescription("Extract audio information from a video project item. Returns audio metadata and media path for further processing."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
		),
		makeAudioHandler(orch, logger, "extractAudioFromVideo", []string{"project_item_index"}),
	)
}
