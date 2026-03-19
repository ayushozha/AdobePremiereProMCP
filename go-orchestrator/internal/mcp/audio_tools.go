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
			gomcp.WithDescription("Set an audio level keyframe at a specific time on an audio clip."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("time", gomcp.Required(), gomcp.Description("Time in seconds (relative to clip start) for the keyframe")),
			gomcp.WithNumber("level_db", gomcp.Required(), gomcp.Description("Audio level in dB (-96 to +15)")),
		),
		makeAudioHandler(orch, logger, "setAudioLevelKeyframe", []string{"track_index", "clip_index", "time", "level_db"}),
	)

	// 3. premiere_get_audio_level
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_level",
			gomcp.WithDescription("Get the current audio level of a clip, including keyframe info."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
		),
		makeAudioHandler(orch, logger, "getAudioLevel", []string{"track_index", "clip_index"}),
	)

	// 4. premiere_normalize_audio
	s.AddTool(
		gomcp.NewTool("premiere_normalize_audio",
			gomcp.WithDescription("Normalize audio to a target level, removing existing keyframes and setting a flat gain."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("target_db", gomcp.Required(), gomcp.Description("Target audio level in dB (-96 to +15)")),
		),
		makeAudioHandler(orch, logger, "normalizeAudio", []string{"track_index", "clip_index", "target_db"}),
	)

	// 5. premiere_set_audio_gain
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_gain",
			gomcp.WithDescription("Set the master audio gain on a source project item (not a timeline clip)."),
			gomcp.WithNumber("project_item_index", gomcp.Required(), gomcp.Description("Zero-based index of the project item in the root bin")),
			gomcp.WithNumber("gain_db", gomcp.Required(), gomcp.Description("Gain in dB (-96 to +15)")),
		),
		makeAudioHandler(orch, logger, "setAudioGain", []string{"project_item_index", "gain_db"}),
	)

	// -----------------------------------------------------------------------
	// Track Controls (6-9)
	// -----------------------------------------------------------------------

	// 6. premiere_mute_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_mute_audio_track",
			gomcp.WithDescription("Mute or unmute an audio track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithBoolean("muted", gomcp.Required(), gomcp.Description("True to mute, false to unmute")),
		),
		makeAudioHandler(orch, logger, "muteAudioTrack", []string{"track_index", "muted"}),
	)

	// 7. premiere_solo_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_solo_audio_track",
			gomcp.WithDescription("Solo or unsolo an audio track. Requires QE DOM (app.enableQE())."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithBoolean("soloed", gomcp.Required(), gomcp.Description("True to solo, false to unsolo")),
		),
		makeAudioHandler(orch, logger, "soloAudioTrack", []string{"track_index", "soloed"}),
	)

	// 8. premiere_set_audio_track_volume
	s.AddTool(
		gomcp.NewTool("premiere_set_audio_track_volume",
			gomcp.WithDescription("Set the volume of an audio track (0.0 = silence, 1.0 = unity, up to 4.0 = +12dB)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("volume", gomcp.Required(), gomcp.Description("Volume level (0.0 to 4.0, where 1.0 is unity)")),
		),
		makeAudioHandler(orch, logger, "setAudioTrackVolume", []string{"track_index", "volume"}),
	)

	// 9. premiere_get_audio_track_info
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_track_info",
			gomcp.WithDescription("Get information about an audio track including name, clip count, mute, and lock state."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
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
			gomcp.WithDescription("Tag an audio clip with an Essential Sound type (dialogue, music, sfx, or ambience)."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithString("type", gomcp.Required(),
				gomcp.Description("Essential Sound type"),
				gomcp.Enum("dialogue", "music", "sfx", "ambience"),
			),
		),
		makeAudioHandler(orch, logger, "setEssentialSoundType", []string{"track_index", "clip_index", "type"}),
	)

	// 17. premiere_set_essential_sound_loudness
	s.AddTool(
		gomcp.NewTool("premiere_set_essential_sound_loudness",
			gomcp.WithDescription("Set the target loudness for an Essential Sound-tagged clip in LUFS."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithNumber("clip_index", gomcp.Required(), gomcp.Description("Zero-based clip index on the track")),
			gomcp.WithNumber("target_lufs", gomcp.Required(), gomcp.Description("Target loudness in LUFS (e.g. -23 for EBU R 128)")),
		),
		makeAudioHandler(orch, logger, "setEssentialSoundLoudness", []string{"track_index", "clip_index", "target_lufs"}),
	)

	// 18. premiere_enable_auto_ducking
	s.AddTool(
		gomcp.NewTool("premiere_enable_auto_ducking",
			gomcp.WithDescription("Enable or disable auto-ducking on an audio track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithBoolean("enabled", gomcp.Required(), gomcp.Description("True to enable, false to disable")),
			gomcp.WithNumber("duck_amount", gomcp.Description("Duck amount in dB (default: -15)")),
			gomcp.WithNumber("sensitivity", gomcp.Description("Sensitivity 0-100 (default: 50)")),
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
			gomcp.WithDescription("Add a new audio track to the active sequence."),
			gomcp.WithString("name", gomcp.Description("Name for the new track (optional)")),
			gomcp.WithString("channel_type",
				gomcp.Description("Channel type for the track"),
				gomcp.Enum("mono", "stereo", "5.1", "adaptive"),
			),
		),
		makeAudioHandler(orch, logger, "addAudioTrack", []string{"name", "channel_type"}),
	)

	// 22. premiere_delete_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_delete_audio_track",
			gomcp.WithDescription("Delete an audio track from the active sequence. Requires QE DOM."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index to delete")),
		),
		makeAudioHandler(orch, logger, "deleteAudioTrack", []string{"track_index"}),
	)

	// 23. premiere_rename_audio_track
	s.AddTool(
		gomcp.NewTool("premiere_rename_audio_track",
			gomcp.WithDescription("Rename an audio track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based audio track index")),
			gomcp.WithString("name", gomcp.Required(), gomcp.Description("New name for the track")),
		),
		makeAudioHandler(orch, logger, "renameAudioTrack", []string{"track_index", "name"}),
	)

	// 24. premiere_get_audio_tracks
	s.AddTool(
		gomcp.NewTool("premiere_get_audio_tracks",
			gomcp.WithDescription("List all audio tracks in the active sequence with clip counts, mute, and lock state."),
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
			gomcp.WithDescription("Add a new video track to the active sequence."),
			gomcp.WithString("name", gomcp.Description("Name for the new track (optional)")),
		),
		makeAudioHandler(orch, logger, "addVideoTrack", []string{"name"}),
	)

	// 28. premiere_delete_video_track
	s.AddTool(
		gomcp.NewTool("premiere_delete_video_track",
			gomcp.WithDescription("Delete a video track from the active sequence. Requires QE DOM."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index to delete")),
		),
		makeAudioHandler(orch, logger, "deleteVideoTrack", []string{"track_index"}),
	)

	// 29. premiere_rename_video_track
	s.AddTool(
		gomcp.NewTool("premiere_rename_video_track",
			gomcp.WithDescription("Rename a video track."),
			gomcp.WithNumber("track_index", gomcp.Required(), gomcp.Description("Zero-based video track index")),
			gomcp.WithString("name", gomcp.Required(), gomcp.Description("New name for the track")),
		),
		makeAudioHandler(orch, logger, "renameVideoTrack", []string{"track_index", "name"}),
	)

	// 30. premiere_get_video_tracks
	s.AddTool(
		gomcp.NewTool("premiere_get_video_tracks",
			gomcp.WithDescription("List all video tracks in the active sequence with clip counts, mute, and lock state."),
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
