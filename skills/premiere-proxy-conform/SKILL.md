---
name: premiere-proxy-conform
description: Speed Adobe Premiere Pro editing with proxy creation, attachment, media matching, and original-media readiness checks. Use for high-resolution or remote-media projects, batch proxy preparation, relinking, proxy status audits, or pre-delivery original-media checks through the Premiere Pro MCP server.
---

# Premiere Proxy Conform

Treat identity and timebase matching as mandatory. A fast proxy with the wrong media or frame rate is worse than no proxy.

## Required server profile

Start or restart the MCP server with `MCP_TOOL_PROFILE=proxies`; the default `standard` profile does not expose the complete proxy mutation and readback workflow.

## Workflow

1. Inventory project media with `premiere_get_project_items`, `premiere_get_media_info`, and `premiere_get_offline_items`. Capture source path, duration, frame rate, audio presence, and offline status. Note that proxy mutation tools address root-bin item indices; stop or reorganize deliberately when an item cannot be addressed safely.
2. Choose an edit-friendly intraframe proxy preset at an existing absolute `.epr` path and a deterministic absolute output path whose parent directory exists. Preserve source-relative names and refuse to overwrite an existing file.
3. For a new proxy, call `premiere_create_proxy` with `output_path` and require a non-empty `jobId`, `requestAccepted: true`, and `state: pending_encode`. This queues Adobe Media Encoder; it does not attach media. Poll `premiere_probe_media` for that exact output path at a bounded interval. Require two consecutive successful probes at least 10 seconds apart with the same positive `file_size_bytes` before treating the encode as stable. Default to a 30-minute timeout when the user gives no limit. On timeout or probe failure, report the encode as pending or failed and do not attach the partial file. Once stable, call `premiere_attach_proxy` for that output path and require `verified: true` plus the same canonical path. Never attach a second time after attachment becomes readable.
4. Verify every item using both proxy read tools; stop if their proxy state or canonical paths disagree. Call `premiere_probe_media` for both source and proxy, then compare duration, frame rate, and audio streams.
5. Enable proxy playback with `premiere_toggle_proxies` and require `verified: true` plus `proxiesEnabled: true`. Before final delivery, verify originals are online, disable proxy playback, and require verified `proxiesEnabled: false`. Do not infer that an export used originals solely from this playback-state readback; validate the rendered output independently.
6. Run a final relink/status audit. If direct-export tools are also loaded, render and probe a short representative segment before the full export; otherwise report that delivery validation remains pending.

## Guardrails

- Never delete originals or existing proxies.
- Detach only through `premiere_detach_proxy`; require cleared `hasProxy` and path readback before reporting success.
- Stop on duplicate basenames, variable-frame-rate mismatches, offline originals, or duration drift.
- Use one encode per unique source; reuse verified proxy files across sequences.
- Report created, reused, attached, skipped, and failed counts separately.
