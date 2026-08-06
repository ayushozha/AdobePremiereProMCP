---
name: premiere-transition-pack
description: Select and apply restrained, installed video or audio transitions in Adobe Premiere Pro. Use for transition cleanup, dissolves, dips, audio crossfades, transition audits, or repeatable transition recipes through the Premiere Pro MCP server.
---

# Premiere Transition Pack

Use transitions to clarify structure or hide a motivated discontinuity. Prefer a clean cut when no transition improves the edit.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=transitions`; the default `standard` profile omits parts of transition discovery and cleanup.

## Workflow

1. Inspect the target cuts and call `premiere_get_installed_transitions`. Use only names confirmed by the host. `premiere_get_available_transitions` is a legacy English-name probe, not a complete inventory.
2. Resolve the active sequence index, call `premiere_duplicate_sequence`, verify exactly one new copy, and activate it with `premiere_set_active_sequence`. Leave the source untouched as the recovery boundary. Record existing transitions with `premiere_get_transitions`; timeline snapshots and saved sequence versions are comparison/audit data only, not whole-sequence rollback points.
3. Choose one small vocabulary: short dissolve for continuity, dip for a time/section break, and an audio crossfade for dialogue/music joins.
4. Apply only a host-confirmed name with `premiere_add_video_transition` or `premiere_add_audio_transition`, supplying the desired duration in the add call.
5. Read transitions back and verify the public DOM count increased and that name, cut location, and duration match the request.

## Guardrails

- Do not run `premiere_apply_transition_to_all_cuts` without an explicit request and an untouched duplicate recovery sequence.
- Keep dialogue audio crossfades short and listen for phase or level changes.
- Wipe direction/color, synthesized zoom/glitch transitions, and alignment mutation are explicitly unsupported by the CEP backend. Use a confirmed installed transition instead.
- Remove a failed or misplaced transition before retrying.
