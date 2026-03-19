package mcp

import (
	"context"
	"fmt"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// registerAudioTools registers all audio, track control, and track management
// MCP tools. These correspond to ExtendScript functions 1-33 in premiere.jsx.
func registerAudioTools(s *server.MCPServer, orch Orchestrator, logger *zap.Logger) {
	// -----------------------------------------------------------------------
	// Audio Levels (1-5)
	// -----------------------------------------------------------------------

	// 1. premiere_set_audio_level — already registered in tools.go; skip here.

	// 2. premiere_set_audio_level_keyframe
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_level_keyframe",
			gomcp.WithDescription("Add or update an audio volume keyframe at a specific time on an audio clip. Keyframes enable volume automation -- the level smoothly transitions between keyframe values. Add multiple keyframes to create fade-ins, fade-outs, or duck audio under narration. Time is relative to the clip's start on the timeline."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track). Use premiere_get_audio_tracks to list tracks.")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track. Use premiere_get_clips_on_track with track_type='audio' to find indices.")),
			gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Keyframe time in seconds relative to the clip's start position on the timeline. 0 = clip start. Must be within the clip's duration.")),
			gomcp.WithNumber("level_db", gomcp.Required(), gomcp.Description("Audio level in decibels at this keyframe. 0 = unity gain, -6 = half volume, -96 = silence, +6 = double volume, +15 = max boost. Range: -96 to +15.")),
		),
		makeAudioHandler(orch, logger, "setAudioLevelKeyframe", []string{"track_index", "clip_index", "time", "level_db"}),
	)

	// 3. premiere_get_audio_level
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_level",
			gomcp.WithDescription("Get the current audio volume level of a clip, including whether keyframes are set and the flat (non-keyframed) level in dB. Use this to inspect a clip's volume before adjusting it."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track).")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track.")),
		),
		makeAudioHandler(orch, logger, "getAudioLevel", []string{"track_index", "clip_index"}),
	)

	// 4. premiere_normalize_audio
	s.AddTool(
		gomcp.NewTool("premiere_normalize_audio",
			gomcp.WithDescription("Normalize an audio clip to a target level in dB by removing all existing volume keyframes and setting a flat (constant) gain value. Useful for ensuring consistent volume across clips. For broadcast-standard loudness normalization, consider using premiere_set_essential_sound_loudness instead."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track).")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track.")),
			gomcp.WithNumber("target_db", gomcp.Required(), gomcp.Description("Target audio level in dB. Common values: 0 = unity, -3 = broadcast peak, -6 = safe headroom, -12 = typical dialogue level. Range: -96 to +15.")),
		),
		makeAudioHandler(orch, logger, "normalizeAudio", []string{"track_index", "clip_index", "target_db"}),
	)

	// 5. premiere_set_audio_gain
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_gain",
			gomcp.WithDescription("Set the master audio gain on a source project item (the original media in the Project panel, not a timeline clip instance). This affects all instances of the clip on all timelines. For timeline-clip-specific volume, use premiere_set_audio_level or premiere_set_audio_level_keyframe instead."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin. Use premiere_get_project_items to find indices.")),
			gomcp.WithNumber("gain_db", gomcp.Required(), gomcp.Description("Master gain in dB applied to the source media. 0 = no gain change, -6 = reduce by 6 dB, +6 = boost by 6 dB. Range: -96 to +15.")),
		),
		makeAudioHandler(orch, logger, "setAudioGain", []string{"project_item_index", "gain_db"}),
	)

	// -----------------------------------------------------------------------
	// Track Controls (6-9)
	// -----------------------------------------------------------------------

	// 6. premiere_mute_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_mute_audio_track",
			gomcp.WithDescription("Mute or unmute an entire audio track. When muted, no audio from any clip on this track is included in playback or export. The mute state is shown by the 'M' button in the Track Header. This does not affect individual clip volume settings."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track). Use premiere_get_audio_tracks to list tracks.")),
			gomcp.WithBoolean("muted", gomcp.Required(), gomcp.Description("True to mute the track (silence all clips), false to unmute (restore playback).")),
		),
		makeAudioHandler(orch, logger, "muteAudioTrack", []string{"track_index", "muted"}),
	)

	// 7. premiere_solo_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_solo_audio_track",
			gomcp.WithDescription("Solo or unsolo an audio track. When a track is soloed, only soloed tracks are heard during playback -- all non-soloed tracks are temporarily muted. Useful for isolating dialogue, music, or effects during mixing. Multiple tracks can be soloed simultaneously."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track). Use premiere_get_audio_tracks to list tracks.")),
			gomcp.WithBoolean("soloed", gomcp.Required(), gomcp.Description("True to solo the track (mute all non-soloed tracks), false to unsolo.")),
		),
		makeAudioHandler(orch, logger, "soloAudioTrack", []string{"track_index", "soloed"}),
	)

	// 8. premiere_set_audio_track_volume
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_track_volume",
			gomcp.WithDescription("Set the master volume fader level for an entire audio track. This is a track-level gain applied on top of individual clip volumes. Equivalent to the track volume fader in the Audio Mixer panel. For clip-level volume, use premiere_set_audio_level instead."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track). Use premiere_get_audio_tracks to list tracks.")),
			gomcp.WithNumber("volume", gomcp.Required(), gomcp.Description("Volume level as a linear multiplier. 0.0 = silence, 1.0 = unity (no change, default), 2.0 = double (+6 dB), 4.0 = maximum (+12 dB). Range: 0.0 to 4.0.")),
		),
		makeAudioHandler(orch, logger, "setAudioTrackVolume", []string{"track_index", "volume"}),
	)

	// 9. premiere_get_audio_track_info
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_track_info",
			gomcp.WithDescription("Get detailed information about an audio track, including track name, number of clips, mute/solo/lock state, and volume level. Use this to inspect track configuration before making changes."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track). Use premiere_get_audio_tracks to list all tracks.")),
		),
		makeAudioHandler(orch, logger, "getAudioTrackInfo", []string{"track_index"}),
	)

	// -----------------------------------------------------------------------
	// Audio Channels (10-11)
	// -----------------------------------------------------------------------

	// 10. premiere_get_audio_channel_mapping
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_channel_mapping",
			gomcp.WithDescription("Get the audio channel mapping for a project item."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
		),
		makeAudioHandler(orch, logger, "getAudioChannelMapping", []string{"project_item_index"}),
	)

	// 11. premiere_set_audio_channel_mapping
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_channel_mapping",
			gomcp.WithDescription("Set the audio channel mapping for a project item. Channel type: 0=mono, 1=stereo, 2=5.1."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
			gomcp.WithNumber("mapping", gomcp.Required(), gomcp.Description("Channel type value (0=mono, 1=stereo, 2=5.1)")),
		),
		makeAudioHandler(orch, logger, "setAudioChannelMapping", []string{"project_item_index", "mapping"}),
	)

	// -----------------------------------------------------------------------
	// Audio Effects (12-14)
	// -----------------------------------------------------------------------

	// 12. premiere_apply_audio_effect
	s.AddTool(
		gomcp.NewTool("premiere_apply_audio_effect",
			gomcp.WithDescription("Apply a named audio effect to a clip. Requires QE DOM."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithString("effect_name", gomcp.Required(), gomcp.Description("Name of the audio effect (e.g. 'DeEsser', 'Parametric Equalizer')")),
		),
		makeAudioHandler(orch, logger, "applyAudioEffect", []string{"track_index", "clip_index", "effect_name"}),
	)

	// 13. premiere_remove_audio_effect
	s.AddTool(
		gomcp.NewTool("premiere_remove_audio_effect",
			gomcp.WithDescription("Remove an applied audio effect from a clip by effect index (0-based, excluding the built-in Volume component)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("effect_index", gomcp.Required(), gomcp.Description("Zero-based index of the applied effect to remove")),
		),
		makeAudioHandler(orch, logger, "removeAudioEffect", []string{"track_index", "clip_index", "effect_index"}),
	)

	// 14. premiere_get_audio_effects
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_effects",
			gomcp.WithDescription("List all audio effects (components) applied to a clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getAudioEffects", []string{"track_index", "clip_index"}),
	)

	// -----------------------------------------------------------------------
	// Audio Transitions (15)
	// -----------------------------------------------------------------------

	// 15. premiere_add_audio_crossfade
	s.AddTool(
		gomcp.NewTool("premiere_add_audio_crossfade",
			gomcp.WithDescription("Add an audio crossfade transition to a clip. Requires QE DOM."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("duration", gomcp.Description("Crossfade duration in seconds (default: 1.0)")),
			gomcp.WithString("type",
				gomcp.Description("Crossfade type"),
				gomcp.Enum("constant_power", "constant_gain", "exponential"),
			),
		),
		makeAudioHandler(orch, logger, "addAudioCrossfade", []string{"track_index", "clip_index", "duration", "type"}),
	)

	// -----------------------------------------------------------------------
	// Essential Sound (16-18)
	// -----------------------------------------------------------------------

	// 16. premiere_set_essential_sound_type
	s.AddTool(
		gomcp.NewTool("premiere_set_essential_sound_type",
			gomcp.WithDescription("Tag an audio clip with an Essential Sound type, enabling type-specific audio processing. Each type unlocks different auto-processing options: 'dialogue' enables loudness normalization, noise reduction, and speech enhancement; 'music' enables ducking and loudness controls; 'sfx' enables loudness and stereo width; 'ambience' enables ducking sensitivity. Tagging clips is required before using auto-ducking or loudness normalization."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track).")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track.")),
			gomcp.WithString("type", gomcp.Required(),
				gomcp.Description("Essential Sound category. 'dialogue' = speech/narration, 'music' = background music/score, 'sfx' = sound effects/foley, 'ambience' = room tone/environmental audio."),
				gomcp.Enum("dialogue", "music", "sfx", "ambience"),
			),
		),
		makeAudioHandler(orch, logger, "setEssentialSoundType", []string{"track_index", "clip_index", "type"}),
	)

	// 17. premiere_set_essential_sound_loudness
	s.AddTool(
		gomcp.NewTool("premiere_set_essential_sound_loudness",
			gomcp.WithDescription("Set the target loudness (in LUFS) for an Essential Sound-tagged clip. Premiere auto-adjusts the clip's gain to match the target. The clip must first be tagged with premiere_set_essential_sound_type. Common broadcast standards: -23 LUFS (EBU R 128, European broadcast), -24 LUFS (ATSC A/85, US broadcast), -14 LUFS (Spotify/YouTube streaming)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index (0 = first audio track).")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the audio track.")),
			gomcp.WithNumber("target_lufs", gomcp.Required(), gomcp.Description("Target loudness in LUFS (Loudness Units Full Scale). Common values: -23 (EBU R 128), -24 (ATSC A/85 US broadcast), -16 (podcasts), -14 (streaming platforms). Range: roughly -40 to -5.")),
		),
		makeAudioHandler(orch, logger, "setEssentialSoundLoudness", []string{"track_index", "clip_index", "target_lufs"}),
	)

	// 18. premiere_enable_auto_ducking
	s.AddTool(
		gomcp.NewTool("premiere_enable_auto_ducking",
			gomcp.WithDescription("Enable or disable auto-ducking on an audio track. Auto-ducking automatically reduces the volume of music/ambience when dialogue is detected on other tracks. The track should contain music or ambience clips tagged via Essential Sound. Ducking keyframes are generated automatically based on dialogue presence."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index containing the music/ambience to duck (0 = first audio track).")),
			gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("True to enable auto-ducking, false to disable it.")),
			gomcp.WithNumber("duck_amount", gomcp.Description("How much to reduce volume during ducking in dB (default: -15). More negative = more reduction. Typical range: -6 to -24.")),
			gomcp.WithNumber("sensitivity", gomcp.Description("How sensitive the ducking detection is (default: 50). Higher values trigger ducking more easily. Range: 0 (least sensitive) to 100 (most sensitive).")),
		),
		makeAudioHandler(orch, logger, "enableAutoDucking", []string{"track_index", "enabled", "duck_amount", "sensitivity"}),
	)

	// -----------------------------------------------------------------------
	// Audio Analysis (19-20)
	// -----------------------------------------------------------------------

	// 19. premiere_detect_silence
	s.AddTool(
		gomcp.NewTool("premiere_detect_silence",
			gomcp.WithDescription("Detect silence regions in an audio clip. Returns clip info and media path for detailed analysis via the media engine."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("threshold_db", gomcp.Description("Silence threshold in dB (default: -40)")),
			gomcp.WithNumber("min_duration_ms", gomcp.Description("Minimum silence duration in milliseconds (default: 500)")),
		),
		makeAudioHandler(orch, logger, "detectSilence", []string{"track_index", "clip_index", "threshold_db", "min_duration_ms"}),
	)

	// 20. premiere_get_audio_peak_level
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_peak_level",
			gomcp.WithDescription("Get the current audio level and media path for a clip. For true peak analysis, use the media engine."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getAudioPeakLevel", []string{"track_index", "clip_index"}),
	)

	// -----------------------------------------------------------------------
	// Audio Track Management (21-26)
	// -----------------------------------------------------------------------

	// 21. premiere_add_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_add_audio_track",
			gomcp.WithDescription("Add a new audio track to the active sequence. The track is appended after the existing audio tracks. Specify the channel type based on your audio source: mono for single-channel, stereo (default) for two-channel, 5.1 for surround sound, adaptive for multi-channel routing."),
			gomcp.WithString("name", gomcp.Description("Display name for the new track (e.g. 'Dialogue', 'Music', 'SFX'). If omitted, Premiere assigns a default name like 'Audio 4'.")),
			gomcp.WithString("channel_type",
				gomcp.Description("Audio channel type for the track (default: 'stereo'). 'mono' = single channel (good for dialogue), 'stereo' = two channels (most common), '5.1' = surround sound (6 channels), 'adaptive' = flexible multi-channel routing."),
				gomcp.Enum("mono", "stereo", "5.1", "adaptive"),
			),
		),
		makeAudioHandler(orch, logger, "addAudioTrack", []string{"name", "channel_type"}),
	)

	// 22. premiere_delete_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_delete_audio_track",
			gomcp.WithDescription("Delete an audio track from the active sequence. WARNING: Any clips on the deleted track are permanently removed from the timeline. The last audio track cannot be deleted. Use premiere_get_audio_tracks to verify the track is empty or confirm which track to delete."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index to delete (0 = first audio track). Use premiere_get_audio_tracks to list tracks.")),
		),
		makeAudioHandler(orch, logger, "deleteAudioTrack", []string{"track_index"}),
	)

	// 23. premiere_rename_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_rename_audio_track",
			gomcp.WithDescription("Change the display name of an audio track. Track names appear in the Timeline panel header and the Audio Mixer. Useful for organizing tracks by purpose (e.g. 'Dialogue', 'Music', 'SFX', 'Ambience')."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index to rename (0 = first audio track).")),
			gomcp.WithString("name", gomcp.Required(), gomcp.Description("New display name for the track (e.g. 'Dialogue', 'Music', 'Foley').")),
		),
		makeAudioHandler(orch, logger, "renameAudioTrack", []string{"track_index", "name"}),
	)

	// 24. premiere_get_audio_tracks
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_tracks",
			gomcp.WithDescription("List all audio tracks in the active sequence with each track's index, name, clip count, channel type, mute/solo/lock state, and volume level. Use this to find track indices for other audio operations or to understand the audio track layout."),
		),
		makeAudioHandler(orch, logger, "getAudioTracks", nil),
	)

	// 25. premiere_lock_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_lock_audio_track",
			gomcp.WithDescription("Lock or unlock an audio track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithBoolean("locked", gomcp.Required(), gomcp.Description("True to lock, false to unlock")),
		),
		makeAudioHandler(orch, logger, "lockAudioTrack", []string{"track_index", "locked"}),
	)

	// 26. premiere_set_audio_track_target
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_track_target",
			gomcp.WithDescription("Set track targeting for an audio track (determines which track receives new clips)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithBoolean("targeted", gomcp.Required(), gomcp.Description("True to target, false to untarget")),
		),
		makeAudioHandler(orch, logger, "setAudioTrackTarget", []string{"track_index", "targeted"}),
	)

	// -----------------------------------------------------------------------
	// Video Track Management (27-33)
	// -----------------------------------------------------------------------

	// 27. premiere_add_video_track
	s.AddTool(
		gomcp.NewTool("premiere_add_video_track",
			gomcp.WithDescription("Add a new video track to the active sequence. The track is appended above the existing video tracks. Higher-numbered video tracks render on top of lower ones (compositing order)."),
			gomcp.WithString("name", gomcp.Description("Display name for the new track (e.g. 'Titles', 'B-Roll', 'Overlays'). If omitted, Premiere assigns a default name like 'Video 4'.")),
		),
		makeAudioHandler(orch, logger, "addVideoTrack", []string{"name"}),
	)

	// 28. premiere_delete_video_track
	s.AddTool(
		gomcp.NewTool("premiere_delete_video_track",
			gomcp.WithDescription("Delete a video track from the active sequence. WARNING: Any clips on the deleted track are permanently removed from the timeline. The last video track cannot be deleted. Use premiere_get_video_tracks to verify the track is empty or confirm which track to delete."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index to delete (0 = bottom video track). Use premiere_get_video_tracks to list tracks.")),
		),
		makeAudioHandler(orch, logger, "deleteVideoTrack", []string{"track_index"}),
	)

	// 29. premiere_rename_video_track
	s.AddTool(
		gomcp.NewTool("premiere_rename_video_track",
			gomcp.WithDescription("Change the display name of a video track. Track names appear in the Timeline panel header. Useful for organizing tracks by purpose (e.g. 'Main Video', 'B-Roll', 'Titles', 'Adjustment Layers')."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index to rename (0 = bottom video track).")),
			gomcp.WithString("name", gomcp.Required(), gomcp.Description("New display name for the track (e.g. 'Main Video', 'B-Roll', 'Titles').")),
		),
		makeAudioHandler(orch, logger, "renameVideoTrack", []string{"track_index", "name"}),
	)

	// 30. premiere_get_video_tracks
	s.AddTool(
		gomcp.NewTool("premiere_get_video_tracks",
			gomcp.WithDescription("List all video tracks in the active sequence with each track's index, name, clip count, mute (visibility) state, and lock state. Use this to find track indices for clip placement, effects, and other video track operations. Track 0 is the bottom-most (lowest layer) video track."),
		),
		makeAudioHandler(orch, logger, "getVideoTracks", nil),
	)

	// 31. premiere_lock_video_track
	s.AddTool(
		gomcp.NewTool("premiere_lock_video_track",
			gomcp.WithDescription("Lock or unlock a video track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithBoolean("locked", gomcp.Required(), gomcp.Description("True to lock, false to unlock")),
		),
		makeAudioHandler(orch, logger, "lockVideoTrack", []string{"track_index", "locked"}),
	)

	// 32. premiere_mute_video_track
	s.AddTool(
		gomcp.NewTool("premiere_mute_video_track",
			gomcp.WithDescription("Mute (hide) or unmute (show) a video track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithBoolean("muted", gomcp.Required(), gomcp.Description("True to mute/hide, false to unmute/show")),
		),
		makeAudioHandler(orch, logger, "muteVideoTrack", []string{"track_index", "muted"}),
	)

	// 33. premiere_set_video_track_target
	s.AddTool(
		gomcp.NewTool("premiere_set_video_track_target",
			gomcp.WithDescription("Set track targeting for a video track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithBoolean("targeted", gomcp.Required(), gomcp.Description("True to target, false to untarget")),
		),
		makeAudioHandler(orch, logger, "setVideoTrackTarget", []string{"track_index", "targeted"}),
	)
}

// ---------------------------------------------------------------------------
// Generic audio command handler factory
// ---------------------------------------------------------------------------

// makeAudioHandler creates a handler that extracts named parameters from the
// MCP request, packages them into a map, and delegates to EvalAudioCommand.
func makeAudioHandler(orch Orchestrator, logger *zap.Logger, command string, paramNames []string) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		logger.Debug("handling audio tool", zap.String("command", command))

		args := make(map[string]any, len(paramNames))
		for _, name := range paramNames {
			raw := gomcp.ParseArgument(req, name, nil)
			if raw != nil {
				args[name] = raw
			}
		}

		result, err := orch.EvalAudioCommand(ctx, command, args)
		if err != nil {
			logger.Error("audio command failed",
				zap.String("command", command),
				zap.Error(err),
			)
			return gomcp.NewToolResultError(fmt.Sprintf("failed to execute %s: %v", command, err)), nil
		}
		return toolResultJSON(result)
	}
}
