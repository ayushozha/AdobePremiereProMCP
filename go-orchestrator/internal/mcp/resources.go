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
		handleCurrentToolCategories,
	)

	s.AddResource(
		gomcp.NewResource(
			"config://workflow-skills",
			"Workflow Skills",
			gomcp.WithResourceDescription("Reusable Premiere workflow packs and matching tool profiles"),
			gomcp.WithMIMEType("text/markdown"),
		),
		handleWorkflowSkills,
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

func handleCurrentToolCategories(
	_ context.Context,
	_ gomcp.ReadResourceRequest,
) ([]gomcp.ResourceContents, error) {
	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "config://tool-categories",
			MIMEType: "text/plain",
			Text: `PremierPro MCP Tool Categories
==============================

The source registry contains 1,064 tools and is cursor-paginated. Treat the
live tools/list response as authoritative. The default standard profile
exposes a compact everyday-editing set. MCP_TOOL_PROFILE can expose a smaller
union of the following workflow groups; every group also includes core
inspection/versioning tools.

Core
  Host, project, timeline, sequence hash, duplicate-sequence recovery, audit
  snapshots, and verified save operations.
  Examples: premiere_ping, premiere_get_project, premiere_get_timeline,
            premiere_get_sequence_hash, premiere_duplicate_sequence,
            premiere_snapshot_timeline.

dialogue
  Source-path inspection, decoded waveform/silence analysis, reviewable cut
  plans, trims, razors, gaps, audio levels, and crossfades.

captions
  Supplied timed-SRT import, structural validation, readback, and
  identity-checked active-sequence sidecar export. Speech transcription,
  portable styling, and FCC certification are not provided by the CEP backend.
  Examples: premiere_add_subtitles_from_srt, premiere_get_captions,
            premiere_validate_closed_captions, premiere_export_captions.

social
  Auto Reframe, verified vertical/square derivatives, and direct/AME export
  with post-export media probing. Safe-zone and visual crop review remain human
  steps.

transitions
  Installed transition discovery, video transitions, and audio crossfades.

effects
  Installed-effect discovery and generic, readable effect-chain mutation.
  Visual quality still requires frame review.

proxies
  Source/proxy media inspection, creation, attachment, detachment, path, and
  status readback.

delivery
  Confirmed presets, blocking direct exports, AME queue submission, and
  post-export media probing. Caption sidecars require exact active-sequence
  identity. CEP cannot poll individual AME jobs to completion.

unsafe
  Arbitrary ExtendScript, host shell, URL, and filesystem access. This is
  intentionally opt-in and must only be used with trusted input and review.

The all profile exposes the legacy catalog except for the explicitly classified
arbitrary execution and host-filesystem tools. It still contains destructive
Premiere operations, so inspect each tool and preserve an untouched duplicate
sequence before use. Timeline snapshots and saved sequence versions are audit
records, not whole-sequence rollback points.
Use all,unsafe only when arbitrary execution and filesystem access are intended.

Read config://workflow-skills for scoped, fail-closed workflow recipes.`,
		},
	}, nil
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
- Color Grading: Discover effects and edit supported Lumetri parameters
- Titles & Graphics: Import MOGRTs, edit exposed properties, import SRT captions
- Export: Use supported direct/AME presets; probe completed output files
- Workspace: Manage panels, workspaces, and UI layout
- Playback: Control playback, scrub timeline, set playhead
- Planning Tools: Scan assets, parse scripts, match shots, and validate EDLs
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
- Prefer a reviewable EDL and an untouched duplicate sequence before assembly;
  use timeline snapshots only as audit/comparison records
- Treat a queued export as pending until a stable output file is independently
  observed and validated with premiere_probe_media`,
		},
	}, nil
}

func handleWorkflowSkills(
	_ context.Context,
	_ gomcp.ReadResourceRequest,
) ([]gomcp.ResourceContents, error) {
	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "config://workflow-skills",
			MIMEType: "text/markdown",
			Text: `# Premiere workflow skills

The repository ships reusable Agent Skills under skills/. Each skill states its
required preflight, recovery-copy, and readback boundaries. A workflow must not
treat a queued or attempted command as success.

| Skill | Tool profile | Purpose |
|---|---|---|
| premiere-dialogue-cut | dialogue | Decoded-waveform, review-first spoken-word tightening |
| premiere-captions | captions | Import, structurally validate, and export supplied timed captions |
| premiere-social-reframe | social | Verified vertical and square derivatives |
| premiere-transition-pack | transitions | Restrained video/audio transition recipes |
| premiere-look-effects | effects | Parameterized effect chains and visual looks on a recovery copy |
| premiere-proxy-conform | proxies | Proxy creation, attachment, and final conform |
| premiere-batch-delivery | delivery | Sequential direct export, AME submission, and post-file media probing |

Set MCP_TOOL_PROFILE to a comma-separated list of profile names before server
startup, for example captions,effects. Specialized profiles always include the
small core inspection/versioning set. The default is standard. The all profile
excludes classified arbitrary execution/filesystem tools but still contains
destructive Premiere operations. Use all,unsafe only with trusted input when
arbitrary host execution and filesystem access are explicitly intended.
Unknown-only profile values fall back to standard to avoid hiding tools because
of a typo.

Do not claim that an operation succeeded solely because it was attempted. Stop
on explicit unsupported errors, verify mutations with state readback, and treat
queued exports as pending until stable files are observed outside CEP and then
validated with premiere_probe_media. For timeline mutation, preserve an
untouched duplicate sequence or tool-created derivative. Treat timeline
snapshots and saved sequence versions as audit/comparison records only.`,
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

Named Export Presets:
  h264_1080p, h264_4k, prores_422, prores_4444, and dnxhd are aliases.
  Each alias must be mapped to a real Adobe Media Encoder .epr file through
  its PREMIERE_EXPORT_PRESET_* bridge setting. The selected .epr file—not the
  alias—defines the actual codec, dimensions, bitrate, and audio settings.
  Use premiere_export_direct or premiere_export_via_ame when supplying an
  explicit preset_path per export.

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
  Typed placement accepts positive speed values only; use a separately
  verified reverse-clip workflow for reverse playback.`,
		},
	}, nil
}
