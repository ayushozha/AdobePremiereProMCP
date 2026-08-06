---
name: premiere-social-reframe
description: Produce vertical, square, and platform-specific versions of an Adobe Premiere Pro sequence while preserving the master edit. Use for 9:16, 1:1, and social cutdowns, Auto Reframe workflows, safe-zone checks, title repositioning, or multi-platform exports through the Premiere Pro MCP server.
---

# Premiere Social Reframe

Create derivatives from the master; never resize the only copy of a sequence.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=social`; the default `standard` profile does not expose the complete reframe workflow.

## Workflow

1. Resolve the intended source's unique index and ID with `premiere_get_sequence_list`. Activate that index with `premiere_set_active_sequence`, require `verified: true` and the exact source ID, then re-read the sequence list and stop unless `activeSequenceID` is still that ID.
2. With the verified source active, record its hash, resolution, duration, markers from `premiere_get_sequence_markers`, and captions. Establish the recovery boundary; treat any timeline snapshot or saved sequence version as audit data only.
3. For a standard 9:16 or 1:1 derivative, call `premiere_create_vertical_version` or `premiere_create_square_version` with the verified source index. Each compound tool creates its own clone, so do not duplicate twice. Require a non-empty returned derivative ID, `activeSequenceID` equal to that derivative ID, and exactly one new sequence in a fresh sequence-list readback.
4. For a custom aspect ratio, call `premiere_auto_reframe` with the verified `source_sequence_id` and a unique `new_name`. Choose `motion_preset` from Premiere's exact values (`slower`, `default`, or `faster`) and set `use_nested_sequences` deliberately. Require the returned derivative `sequenceID` and `activeSequenceID` to match, then re-read the sequence list and stop unless that ID is active and is the only new sequence. Do not duplicate first unless a separately preserved intermediate is explicitly required.
5. Keep the returned derivative active. Re-read its captions with `premiere_get_captions`, then visually review every shot boundary, graphic, and platform safe zone in Premiere. Before each active-sequence-only mutation, re-read the sequence list and stop if the active ID differs. Use `premiere_set_position` or `premiere_set_scale` only for visual clips whose values can be read back. Position uses normalized coordinates—`(0.5, 0.5)` is center and `(0, 0)` is top-left—not pixels. Reposition captions manually when needed.
6. Re-read the variant timeline. Verify its duration and edit points still match the master unless a cutdown was explicitly requested.
7. Find a real `.epr` path with `premiere_list_export_presets_disk` or use an explicitly supplied preset. Export the returned derivative index with `premiere_export_direct` when a blocking export is acceptable. For `premiere_export_via_ame`, report the job as queued until an external file check sees a stable, non-empty output. Only then call `premiere_probe_media` and validate dimensions, streams, and duration before reporting success.

## Guardrails

- Default to 1080x1920 for 9:16 and 1080x1080 for 1:1 unless delivery requirements say otherwise.
- Keep faces and primary action inside safe zones; do not assume a centered crop works across a whole clip.
- Treat safe-zone compliance as a visual review step; the verified social profile does not create or verify Premiere guide overlays.
- Never overwrite the master sequence or export.
- Stop immediately if an activation response or sequence-list readback does not match the intended source or derivative ID.
- Report shots that need human reframing rather than silently accepting bad crops.
- Do not use `premiere_export_for_youtube`, `premiere_export_for_instagram`, `premiere_export_for_tiktok`, or `premiere_export_for_twitter` as exporters; the legacy CEP handlers only return planning specifications.
