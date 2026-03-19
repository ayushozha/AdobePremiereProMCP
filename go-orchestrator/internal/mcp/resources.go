package mcp

import (
	"context"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerResources registers all MCP resources with the server.
// Resources provide static context that AI assistants can read to understand
// how to work with this MCP server and Premiere Pro.
func registerResources(s *server.MCPServer) {
	s.AddResource(
		gomcp.NewResource(
			"config://premiere-instructions",
			"Premiere Pro MCP Instructions",
			gomcp.WithResourceDescription("Instructions for controlling Adobe Premiere Pro via this MCP server"),
			gomcp.WithMIMEType("text/plain"),
		),
		handlePremiereInstructions,
	)

	s.AddResource(
		gomcp.NewResource(
			"config://tool-categories",
			"Tool Categories",
			gomcp.WithResourceDescription("List of all tool categories with descriptions"),
			gomcp.WithMIMEType("text/plain"),
		),
		handleToolCategories,
	)

	s.AddResource(
		gomcp.NewResource(
			"config://extendscript-reference",
			"ExtendScript Quick Reference",
			gomcp.WithResourceDescription("Quick ExtendScript API reference for Premiere Pro"),
			gomcp.WithMIMEType("text/plain"),
		),
		handleExtendScriptReference,
	)

	s.AddResource(
		gomcp.NewResource(
			"config://project-defaults",
			"Project Defaults",
			gomcp.WithResourceDescription("Default project settings and paths used by the MCP server"),
			gomcp.WithMIMEType("text/plain"),
		),
		handleProjectDefaults,
	)
}

// ---------------------------------------------------------------------------
// Resource handlers
// ---------------------------------------------------------------------------

func handlePremiereInstructions(
	_ context.Context,
	_ gomcp.ReadResourceRequest,
) ([]gomcp.ResourceContents, error) {
	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "config://premiere-instructions",
			MIMEType: "text/plain",
			Text: `You are controlling Adobe Premiere Pro via the PremierPro MCP server.

Available tool categories:
- Project Management: Open/save/close projects, import media, manage bins
- Sequence/Timeline: Create sequences, navigate timeline, set in/out points
- Clip Operations: Insert, overwrite, trim, split, move clips
- Effects & Transitions: Apply effects, transitions, keyframes
- Audio: Set levels, apply audio effects, mix tracks
- Color Grading: Full Lumetri Color control (exposure, contrast, temperature, etc.)
- Titles & Graphics: Add text, MOGRTs, captions, lower thirds
- Export: Export in any format via AME or direct export
- Workspace: Manage panels, workspaces, and UI layout
- Playback: Control playback, scrub timeline, set playhead
- AI Tools: Scan assets, parse scripts, auto-edit
- Batch Operations: Bulk operations across multiple clips
- Advanced Editing: Multi-camera, nesting, compound clips
- Diagnostics: Check system state and troubleshoot issues

Tips:
- Always check if Premiere Pro is running first (premiere_is_running or premiere_ping)
- Open a project before editing (premiere_open with project_path)
- Create or select a sequence before clip operations
- Use premiere_get_project to understand the current state
- Use premiere_get_timeline to inspect what is on the active sequence
- When placing clips, import media first with premiere_import_media
- Export presets: h264_1080p, h264_4k, prores_422, prores_4444, dnxhd
- Track indices are zero-based (first video track = 0)
- Time positions are specified in seconds (floating point)
- Use premiere_scan_assets to discover media files in a directory
- Use premiere_auto_edit for fully automated script-to-edit pipeline`,
		},
	}, nil
}

func handleToolCategories(
	_ context.Context,
	_ gomcp.ReadResourceRequest,
) ([]gomcp.ResourceContents, error) {
	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "config://tool-categories",
			MIMEType: "text/plain",
			Text: `PremierPro MCP Tool Categories
==============================

1. Application (app_tools)
   Launch, close, and check if Premiere Pro is running.
   Tools: premiere_open, premiere_close, premiere_is_running

2. Project Management (project_tools)
   Inspect and manage the current project state.
   Tools: premiere_ping, premiere_get_project, premiere_get_project_info,
          premiere_list_project_items, premiere_get_item_metadata,
          premiere_create_bin, premiere_move_items, premiere_consolidate,
          premiere_get_project_settings

3. Sequence / Timeline (sequence_tools)
   Create, configure, and navigate sequences.
   Tools: premiere_create_sequence, premiere_list_sequences,
          premiere_set_active_sequence, premiere_get_sequence_settings,
          premiere_set_sequence_settings, premiere_get_playhead,
          premiere_set_playhead, premiere_get_in_out_points,
          premiere_set_in_out_points, premiere_clear_in_out_points,
          premiere_add_marker, premiere_list_markers

4. Clip Operations (clip_tools)
   Place, move, trim, split, and remove clips on the timeline.
   Tools: premiere_place_clip, premiere_remove_clip, premiere_import_media,
          premiere_move_clip, premiere_trim_clip, premiere_split_clip,
          premiere_get_clip_properties, premiere_set_clip_enabled

5. Effects & Transitions (effects_tools)
   Apply video/audio effects, transitions, and keyframes.
   Tools: premiere_add_transition, premiere_apply_effect,
          premiere_list_effects, premiere_get_effect_properties,
          premiere_set_effect_property, premiere_remove_effect,
          premiere_add_keyframe, premiere_list_keyframes

6. Audio (audio_tools, audio_advanced_tools)
   Control audio levels, effects, and mixing.
   Tools: premiere_set_audio_level, premiere_apply_audio_effect,
          premiere_get_audio_mix, premiere_set_audio_pan,
          premiere_mute_track, premiere_solo_track,
          premiere_normalize_audio, premiere_set_audio_gain

7. Color Grading (color_tools)
   Full Lumetri Color control panel.
   Tools: premiere_lumetri_get_all, premiere_lumetri_set_exposure,
          premiere_lumetri_set_contrast, premiere_lumetri_set_highlights,
          premiere_lumetri_set_shadows, premiere_lumetri_set_whites,
          premiere_lumetri_set_blacks, premiere_lumetri_set_temperature,
          premiere_lumetri_set_tint, premiere_lumetri_set_saturation,
          premiere_lumetri_set_vibrance, premiere_lumetri_apply_lut

8. Titles & Graphics (graphics_tools, motion_graphics_tools)
   Add text overlays, titles, MOGRTs, and captions.
   Tools: premiere_add_text, premiere_add_mogrt,
          premiere_set_mogrt_property, premiere_add_caption,
          premiere_add_lower_third

9. Export (export_tools, encoding_tools, delivery_tools)
   Export sequences in various formats.
   Tools: premiere_export, premiere_export_direct,
          premiere_export_via_ame, premiere_export_frame,
          premiere_export_aaf, premiere_export_omf,
          premiere_export_audio_only, premiere_render_preview,
          premiere_list_exporters, premiere_list_presets

10. Workspace & UI (workspace_tools, ui_tools, panel_ops_tools)
    Manage workspace layout, panels, and UI state.
    Tools: premiere_set_workspace, premiere_open_panel,
           premiere_get_workspace, premiere_resize_panel

11. Playback (playback_tools)
    Control playback and transport.
    Tools: premiere_play, premiere_pause, premiere_stop,
           premiere_step_forward, premiere_step_backward,
           premiere_shuttle

12. AI-Powered (ai_tools)
    Automated editing using AI intelligence.
    Tools: premiere_scan_assets, premiere_parse_script,
           premiere_auto_edit

13. Transform (transform_tools)
    Position, scale, rotate, and opacity of clips.
    Tools: premiere_set_position, premiere_set_scale,
           premiere_set_rotation, premiere_set_opacity,
           premiere_set_anchor_point

14. Metadata (metadata_tools)
    Read and write clip and project metadata.
    Tools: premiere_get_metadata, premiere_set_metadata

15. Batch Operations (batch_tools)
    Bulk operations across multiple clips or tracks.
    Tools: premiere_batch_apply_effect, premiere_batch_set_property,
           premiere_batch_export

16. Advanced Editing (advanced_edit_tools)
    Multi-cam, nesting, and compound clips.
    Tools: premiere_nest_clips, premiere_create_multicam,
           premiere_flatten_multicam, premiere_create_subclip

17. Templates (template_tools)
    Project and sequence templates.
    Tools: premiere_apply_template, premiere_save_template,
           premiere_list_templates

18. Preferences (preferences_tools)
    Application preferences and settings.
    Tools: premiere_get_preferences, premiere_set_preference

19. Collaboration (collaboration_tools)
    Team workflows and shared projects.
    Tools: premiere_lock_project, premiere_unlock_project

20. Diagnostics & Monitoring (diagnostics_tools, monitoring_tools)
    System health checks and performance monitoring.
    Tools: premiere_diagnostics, premiere_get_system_info,
           premiere_get_performance_stats

21. Scripting (scripting_tools)
    Execute custom ExtendScript in Premiere Pro.
    Tools: premiere_run_extendscript

22. Analytics (analytics_tools)
    Project analytics and statistics.
    Tools: premiere_get_project_stats, premiere_get_timeline_stats

23. Integration (integration_tools)
    Integrations with After Effects, Audition, etc.
    Tools: premiere_dynamic_link_ae, premiere_send_to_audition

24. Camera & Immersive (camera_tools, immersive_tools)
    VR/360 video and camera metadata.
    Tools: premiere_set_vr_projection, premiere_get_camera_metadata

25. Versioning (versioning_tools)
    Project versioning and snapshots.
    Tools: premiere_create_snapshot, premiere_list_snapshots,
           premiere_restore_snapshot

26. Media Browser (media_browser_tools)
    Browse and search for media.
    Tools: premiere_browse_media, premiere_search_stock`,
		},
	}, nil
}

func handleExtendScriptReference(
	_ context.Context,
	_ gomcp.ReadResourceRequest,
) ([]gomcp.ResourceContents, error) {
	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "config://extendscript-reference",
			MIMEType: "text/plain",
			Text: `ExtendScript Quick Reference for Adobe Premiere Pro
====================================================

The MCP server wraps ExtendScript calls internally. This reference
is for understanding what operations are possible and how the
underlying API works.

Core Objects:
  app                       The Application object
  app.project               Current project (Project)
  app.project.activeSequence Active sequence (Sequence)
  app.project.rootItem      Root bin (ProjectItem)

Project:
  app.project.name          Project name
  app.project.path          Project file path
  app.project.sequences     Array of all Sequence objects
  app.project.importFiles(paths)         Import media files
  app.project.createNewSequence(name)    Create a new sequence
  app.project.openSequence(id)           Set active sequence

ProjectItem:
  item.name                 Item name
  item.type                 1=clip, 2=bin, 3=root, 4=file
  item.treePath             Full path in project panel
  item.getMediaPath()       File path on disk
  item.setInPoint(secs)     Set source in point
  item.setOutPoint(secs)    Set source out point
  item.children             Array of child items (for bins)
  item.createBin(name)      Create sub-bin
  item.moveBin(destBin)     Move to another bin

Sequence:
  seq.name                  Sequence name
  seq.sequenceID            Unique ID
  seq.videoTracks           Array of Track objects
  seq.audioTracks           Array of Track objects
  seq.getPlayerPosition()   Current playhead position (Time)
  seq.setPlayerPosition(t)  Set playhead position
  seq.setInPoint(secs)      Set sequence in point
  seq.setOutPoint(secs)     Set sequence out point
  seq.getInPoint()          Get sequence in point
  seq.getOutPoint()         Get sequence out point
  seq.insertClip(item, t)   Insert at position
  seq.overwriteClip(item,t) Overwrite at position
  seq.createSubSequence()   Create subsequence

Track:
  track.clips               Array of TrackItem objects
  track.name                Track name
  track.id                  Track index
  track.isMuted()           Check if muted
  track.setMute(bool)       Mute/unmute

TrackItem (Clip):
  clip.name                 Clip name
  clip.start                Start time on timeline (Time)
  clip.end                  End time on timeline (Time)
  clip.duration             Clip duration (Time)
  clip.inPoint              Source in point (Time)
  clip.outPoint             Source out point (Time)
  clip.type                 1=clip, 2=transition
  clip.components           Array of Component (effects)
  clip.remove(false, false) Remove from timeline
  clip.disabled             Is clip disabled

Component (Effect):
  comp.displayName          Effect name
  comp.properties           Array of ComponentParam
  comp.matchName            Internal match name

ComponentParam:
  param.displayName         Parameter name
  param.getValue()          Current value
  param.setValue(v, true)    Set value (with undo)
  param.addKey(time)        Add keyframe
  param.removeKey(time)     Remove keyframe
  param.getKeys()           Array of keyframe times

Time:
  time.seconds              Time in seconds (float)
  time.ticks                Time in ticks (string)

Common Patterns:
  // Get active sequence clips on video track 0
  var track = app.project.activeSequence.videoTracks[0];
  for (var i = 0; i < track.clips.numItems; i++) {
      var clip = track.clips[i];
      // work with clip
  }

  // Apply effect by matchName
  var fx = qe.project.getVideoEffectByName("matchName");

  // Lumetri Color match name: "Lumetri Color"
  // Cross Dissolve match name: "Cross Dissolve"`,
		},
	}, nil
}

func handleProjectDefaults(
	_ context.Context,
	_ gomcp.ReadResourceRequest,
) ([]gomcp.ResourceContents, error) {
	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "config://project-defaults",
			MIMEType: "text/plain",
			Text: `PremierPro MCP Default Project Settings
========================================

Sequence Defaults:
  Resolution:        1920x1080 (Full HD)
  Frame Rate:        24 fps
  Pixel Aspect:      Square Pixels (1.0)
  Video Tracks:      3
  Audio Tracks:      2
  Audio Sample Rate: 48000 Hz
  Audio Bit Depth:   16-bit

Export Presets:
  h264_1080p    H.264, 1920x1080, ~20 Mbps VBR
  h264_4k       H.264, 3840x2160, ~50 Mbps VBR
  prores_422    Apple ProRes 422
  prores_4444   Apple ProRes 4444 (with alpha)
  dnxhd         Avid DNxHR HQX
  gif           Animated GIF (low res)

Supported Media Formats (Import):
  Video:  .mp4, .mov, .avi, .mkv, .mxf, .r3d, .braw, .ari
  Audio:  .wav, .mp3, .aac, .aif, .flac, .ogg
  Image:  .png, .jpg, .jpeg, .tiff, .psd, .exr, .dpx
  Other:  .mogrt, .prproj, .xml, .edl, .aaf, .omf

Project File Paths:
  Premiere Pro projects use the .prproj extension.
  Auto-save:       ~/Documents/Adobe/Premiere Pro Auto-Save/
  Media Cache:     ~/Library/Application Support/Adobe/Common/Media Cache Files/
  Presets:         ~/Documents/Adobe/Premiere Pro/<version>/Profile-<user>/Settings/Export/
  Effect Presets:  ~/Documents/Adobe/Premiere Pro/<version>/Profile-<user>/Effect Presets/

Track Index Convention:
  Track indices are zero-based.
  Video track 0 is the bottom-most video track (V1 in Premiere UI).
  Audio track 0 is the top-most audio track (A1 in Premiere UI).

Timecode:
  Positions are in seconds (float64). For example, 61.5 = 1 minute, 1.5 seconds.
  The server converts seconds to internal timecode representation automatically.

Speed:
  Default playback speed is 1.0 (100%).
  Values < 1.0 create slow motion, > 1.0 create fast motion.
  Negative values reverse playback.`,
		},
	}, nil
}
