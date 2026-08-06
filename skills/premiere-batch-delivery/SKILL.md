---
name: premiere-batch-delivery
description: Validate, export, and inspect multiple Adobe Premiere Pro deliverables with explicit pending and unverified states. Use for YouTube, social, web, mobile, archive, audio-only, versioned export, delivery QA, caption validation, direct export, AME submission, or repeatable batch output through the Premiere Pro MCP server.
---

# Premiere Batch Delivery

Define a delivery matrix first. Treat a queued job as pending, not as a successful export.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=delivery` to load the focused export, caption-check, and media-probe surface used by this skill.

## Workflow

1. List every target with its unique sequence index and ID, aspect ratio, resolution, frame rate, codec/preset, audio spec, captions, destination, and filename. Resolve the index/ID pair from a fresh `premiere_get_sequence_list` response; do not select by name alone.
2. Check project integrity, delivery specs, captions, and frame accuracy. Before every active-sequence-only caption check or sidecar export, call `premiere_set_active_sequence` for that target index, require `verified: true` with the exact target ID, and re-read `premiere_get_sequence_list`; stop unless `activeSequenceID` still matches. Pass the same ID as `sequence_id` to `premiere_export_captions`. Treat host-only loudness or color summaries as advisory unless a decoded-media analyzer or visual review verifies them. Stop on unresolved errors.
3. Call `premiere_list_export_presets_disk` to find real `.epr` files in known Premiere/AME directories, or use an explicitly supplied preset path. Confirm destination directories and collision policy. Create versioned filenames; never overwrite unless explicitly authorized.
4. Prefer one-at-a-time `premiere_export_direct` calls with a confirmed preset; the call blocks, which bounds concurrency and makes failures attributable. Use `premiere_export_via_ame` only when asynchronous queueing is required, and record its returned job identifier.
5. The CEP bridge cannot poll individual AME jobs to a trustworthy terminal state. Keep AME jobs pending until an external filesystem check observes a stable, non-empty output; if that check is unavailable, leave the job unverified rather than guessing.
6. Call `premiere_probe_media` for every completed output and compare duration, dimensions, frame rate, and audio/video streams with the delivery matrix. Produce a concise delivery report.

## Guardrails

- Do not equate `queued` or `started` with `completed`.
- Do not use the legacy platform compound exporters or `premiere_batch_export_sequences` as proof of completion; their CEP responses do not provide per-output completion readback.
- Keep archive masters separate from distribution encodes.
- Validate sidecar caption paths and names alongside the video deliverable.
- Never reuse a prior active-sequence assumption across delivery-matrix rows; activate and verify each target immediately before its caption sidecar.
- Report successful, failed, skipped, and unverified outputs separately.
