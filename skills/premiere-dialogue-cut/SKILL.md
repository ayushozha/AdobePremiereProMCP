---
name: premiere-dialogue-cut
description: Tighten interviews, podcasts, tutorials, and talking-head edits in Adobe Premiere Pro by detecting silence, preserving natural speech margins, keeping an untouched recovery sequence, and verifying sync. Use for pause removal, dialogue cleanup, jump-cut planning, or rough-cut acceleration through the Premiere Pro MCP server.
---

# Premiere Dialogue Cut

Build an evidence-backed cut plan before touching the timeline. Favor natural pacing over maximum compression.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=dialogue`. The default `standard` profile also contains this workflow, but `dialogue` is the smaller dedicated surface.

## Workflow

1. Call `premiere_ping`, `premiere_get_project`, and `premiere_get_timeline`. Stop if the host, project, or target sequence is unavailable.
2. Resolve the active sequence index with `premiere_get_sequence_list`. Before duplication, capture the source settings and a canonical timeline inventory: track counts, clip order, source-media paths, source in/out ranges, timeline start/end ranges, speed, and duration. Call `premiere_duplicate_sequence`, re-read the list to identify exactly one new copy, and activate it with `premiere_set_active_sequence`. Re-read the same canonical fields and require them to match while ignoring the clone name and sequence/clip identities. Do not compare sequence hashes across distinct sequences: the hash intentionally includes sequence and clip identity, so a valid clone has a different hash. Leave the source sequence untouched as the recovery boundary. `premiere_snapshot_timeline` and `premiere_save_sequence_version` may record before/after audit data, but they cannot restore a whole sequence.
3. List the dialogue clips with `premiere_get_clips_on_track`, then call `premiere_get_clip_info` to resolve each source-media path, source in/out range, speed, and timeline range.
4. Call `premiere_analyze_waveform` once per unique source path. Use its decoded-audio `silence_regions`; the legacy `premiere_detect_silence` host command only returns clip metadata and is not silence evidence. Stop if waveform analysis is unavailable or empty for a clip that should contain audio.
5. Intersect source silence with each used source range and map it to timeline time. Stop on reversed, retimed, or variable-frame-rate media unless that mapping can be verified. Keep 80–150 ms before speech, 120–250 ms after speech, and preserve longer rhetorical pauses unless the user asks for an aggressive cut.
6. Show a proposed plan with edit count, removed duration, shortest retained pause, affected tracks, and protected markers before applying changes.
7. Apply approved edits in reverse timeline order with `premiere_razor_clip`, `premiere_remove_clip_from_track`, and `premiere_ripple_delete_gap`, or use `premiere_ripple_trim` when it expresses the same edit safely.
8. Re-read the timeline and sequence hash. Verify duration change, A/V sync, clip ordering, and that no edit overlaps protected markers.

## Guardrails

- Preserve linked audio/video and room tone unless explicitly told otherwise.
- Do not treat `premiere_detect_silence` as analysis, and do not call `premiere_smart_cut`, `premiere_smart_trim`, or `premiere_generate_rough_cut` when the server reports them as unsupported.
- Stop on the first failed edit; report the last verified time range and the untouched recovery-sequence name.
- Add short audio crossfades only after cuts are verified, using the readback-verified `premiere_add_audio_transition` route.
