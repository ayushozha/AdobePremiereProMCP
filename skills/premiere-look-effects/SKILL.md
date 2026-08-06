---
name: premiere-look-effects
description: Build consistent, reversible visual effect recipes in Adobe Premiere Pro from effects confirmed as installed. Use for clip effects, parameterized looks, motion treatments, vignettes, controlled tests, or applying one approved recipe across multiple clips through the Premiere Pro MCP server.
---

# Premiere Look Effects

Discover capabilities from the host and test a look on one representative clip before batching it.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=effects` for the smallest dedicated surface. The default `standard` profile also contains this skill's verified discovery, attachment, parameter-update, and readback tools.

## Workflow

1. Inspect the project, timeline, installed effects, and current clip effect chain. Resolve the active sequence index, call `premiere_duplicate_sequence`, verify exactly one new copy, and activate it with `premiere_set_active_sequence`. Leave the source sequence untouched as the recovery boundary.
2. Record the current chain with `premiere_get_clip_effects` and the sequence hash. Use `premiere_snapshot_timeline` or `premiere_save_sequence_version` only as optional audit records; neither restores a whole sequence.
3. Build a minimal ordered recipe using only display names returned by `premiere_get_installed_effects`.
4. Apply one effect at a time with `premiere_apply_video_effect`. Require a public-DOM component-count increase and the attached component name.
5. Set only properties actually exposed by that component with `premiere_set_effect_parameter`; read the chain and values back before continuing.
6. Test the recipe on one representative clip. If application or parameter readback fails, stop, reactivate the untouched source sequence, and report the failed duplicate for inspection; do not attempt an unverified component removal or undo. After approval, repeat the same verified steps on additional clips and stop on the first mismatch.

## Guardrails

- Never assume an effect display name exists; confirm it with `premiere_get_installed_effects`.
- Avoid destructive baking. Keep changes parameterized and reversible.
- Do not use legacy effect-enable or named grain, glow, vignette, drop-shadow, stroke, and cinematic-bar helpers in this verified workflow; they either lack portable component-state readback or can return instructions/requested values instead of a verified mutation. Use the generic verified effect path.
- Do not claim visual quality from parameter success alone; request a frame review when appearance matters.
