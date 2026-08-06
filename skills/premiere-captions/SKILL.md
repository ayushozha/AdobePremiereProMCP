---
name: premiere-captions
description: Import, verify, structurally validate, and export timed captions or subtitles in Adobe Premiere Pro. Use for SRT workflows, supplied timed transcripts, caption QA, accessibility checks, or sidecar subtitle delivery through the Premiere Pro MCP server.
---

# Premiere Captions

Use supplied timed text as the source of truth. The current server does not provide trustworthy speech transcription by itself.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=captions`. The default `standard` profile also contains this workflow, but `captions` is the smaller dedicated surface.

## Workflow

1. Verify the host and target sequence with `premiere_ping`, `premiere_get_project`, and `premiere_get_timeline`.
2. Obtain a valid SRT file with segment timestamps. If only audio exists, request or generate a timestamped transcript with a separate transcription system; this server does not claim to transcribe speech.
3. Fix line wrapping in the SRT before import: keep at most two lines, avoid single-word orphan lines, and reject overlaps or non-positive durations.
4. Resolve the active sequence index with `premiere_get_sequence_list`, duplicate it with `premiere_duplicate_sequence`, identify exactly one new copy and its ID, and activate that copy with `premiere_set_active_sequence`. Require `verified: true` with that ID, then re-read the sequence list and stop unless `activeSequenceID` matches. Leave the source untouched as the recovery boundary. A timeline snapshot or saved sequence version is audit data only, not a whole-sequence restore point.
5. Call `premiere_add_subtitles_from_srt` on the duplicate. Require `verified: true` and an exact caption-count match.
6. Call `premiere_get_captions` and compare every readable segment's text, start, and end against the source. Stop on a mismatch.
7. Run `premiere_validate_closed_captions` for structural QA. Reconfirm the duplicate's active ID, export sidecar captions with `premiere_export_captions` using that ID as `sequence_id`, then verify through the filesystem that the sidecar exists, is non-empty, and parses to the expected segment count. If filesystem verification is unavailable, report the export as unverified.

## Guardrails

- Preserve source language, names, numbers, and punctuation; flag uncertain words instead of inventing them.
- Request a visual title-safe review; the CEP caption API cannot verify styling or on-screen placement.
- CEP caption styling, speech alignment, and in-Premiere segment splitting are explicitly unsupported because they cannot be mutated and read back safely. Style manually or use a verified MOGRT-based burned-in workflow.
- Structural validation is not FCC certification. Report caption count, overlaps, warnings, and errors; do not infer transcription accuracy, timeline coverage, reading speed, or placement.
