# Issue status — 2026-09-30

Repository: [ayushozha/AdobePremiereProMCP](https://github.com/ayushozha/AdobePremiereProMCP). This is the follow-up to the [September 29 audit](issue-status-2026-09-29.md), which records the earlier disconnected-host snapshot. The integration branch is `codex/issue-resolution`; its new implementation has **not been merged** at this snapshot. GitHub main remains `914823b` ([PR #18](https://github.com/ayushozha/AdobePremiereProMCP/pull/18)).

The audit began with **11 open issues**. Live GitHub inspection now confirms **9 open / 2 closed**: #16 and #17 were closed as **not planned** on September 30. The proposed closures below distinguish completed requirements from administrative retirement of intake trackers.

## Decisions for all 11 original issues

| Issue | Decision | Evidence and remaining acceptance |
| --- | --- | --- |
| [#1 Community testing](https://github.com/ayushozha/AdobePremiereProMCP/issues/1) | **Keep open** | Ongoing named OS/Premiere testing. Clean-machine installation, 10+ media imports, H.264/ProRes exports, and multiple sequences remain broader requirements than the local test suite or one disposable native run. |
| [#2 Feature requests](https://github.com/ayushozha/AdobePremiereProMCP/issues/2) | **Retire after merge; administrative** | Empty feature-intake umbrella with no comment requests. The README now directs requests to the existing [individual feature template](https://github.com/ayushozha/AdobePremiereProMCP/issues/new?template=feature_request.md). Retirement does not declare every listed feature implemented. |
| [#3 Bug reports](https://github.com/ayushozha/AdobePremiereProMCP/issues/3) | **Retire after merge; administrative** | Its sole concrete [Windows report](https://github.com/ayushozha/AdobePremiereProMCP/issues/3#issuecomment-5368128558) identified reserved-key parsing, swallowed loader errors, stale function-loading checks, and missing `Date.toISOString`; #18 already addressed those defects. The README now directs future bugs to the existing [individual bug template](https://github.com/ayushozha/AdobePremiereProMCP/issues/new?template=bug_report.md). Retirement does not certify all tools or platforms. |
| [#4 Supported versions](https://github.com/ayushozha/AdobePremiereProMCP/issues/4) | **Keep open** | Compatibility needs named version/OS results. Contributor reports and the existing 26.3.2 read-only validation do not fill the version/OS matrix; the new 26.5.2 session passes eight core workflows on macOS 27.0.0 arm64, en_US, but does not fill the broader matrix. |
| [#6 Native end-to-end testing](https://github.com/ayushozha/AdobePremiereProMCP/issues/6) | **Close as completed after verified merge** | The [guarded runner](live-premiere-testing.md) passes **8/8 native workflows**: import, sequence/placement, transitions/effects, Lumetri, audio, H.264 export, frame capture, and script/EDL execution. [Acceptance evidence](verification/native-core-2026-09-30.json) includes exact native readback, output decoding and report/output hashes. Scope: Premiere 26.5.2, macOS 27.0.0 arm64, en_US. |
| [#7 npm distribution](https://github.com/ayushozha/AdobePremiereProMCP/issues/7) | **Keep open** | [Packaging](npm-distribution.md), four-platform binaries, checksum-pinned installer, and an opt-in release workflow are implemented locally. The registry lookup returned **404** and `npm whoami` returned **ENEEDAUTH**. The package has not been published; registry installation and actual execution on the supported targets remain required. |
| [#8 UXP support](https://github.com/ayushozha/AdobePremiereProMCP/issues/8) | **Keep open** | A read-only inspection prototype and [migration guidance](uxp-migration.md) exist. Real native panel validation, authenticated transport, MCP adapter, and CEP/UXP dual-mode workflow support remain incomplete. |
| [#9 Observability](https://github.com/ayushozha/AdobePremiereProMCP/issues/9) | **Close as completed after verified merge** | The branch implements per-tool request/error counts and latency, structured completion logs, a separate Go Prometheus endpoint, six-panel Grafana template, and improved health/readiness endpoints. Tests exercise real stdio/HTTP dispatch, cancellation, recovery, concurrent scrapes, privacy, and listener shutdown. Grafana import/rendering in the target installation remains unverified; this is disclosed rather than a claim that a dashboard has been deployed. |
| [#10 Video tutorials](https://github.com/ayushozha/AdobePremiereProMCP/issues/10) | **Keep open** | [Recording outlines](tutorial-recording-plan.md) cover four topics. **No playable tutorial videos have been produced.** Completion requires four actual recordings with captions and native editing/export evidence; documentation and scripts do not meet the video requirement. |
| [#16 Replacement-download post](https://github.com/ayushozha/AdobePremiereProMCP/issues/16) | **Closed; not planned** | Zero reproducible technical details; external replacement-download promotion. This was moderation, not a software fix. The external link was not followed. |
| [#17 Replacement-download post](https://github.com/ayushozha/AdobePremiereProMCP/issues/17) | **Closed; not planned** | Same unsupported replacement-download pattern. This was moderation, not a software fix. The external link was not followed. |

## Native execution and resulting fixes

The final fresh disposable run completed **8 pass / 0 fail / 0 blocked** on **Premiere Pro 26.5.2, macOS 27.0.0 arm64, en_US**. All backend services ran on this machine. The MCP `all` profile enumerated 1,033 tools without enabling `unsafe`; the runner uses a guarded subset and verifies the exact disposable project and owned sequences before edits.

The actual failures found and fixed during native testing were:

- Nonmodal empty-sequence creation with imported seed cleanup, exact new sequence identity, writable requested name, settings readback and owned rollback.
- Numeric `0` or boolean `true` save status with exact saved path/nonempty-file verification; verified saving before close.
- Audio/video source marks isolated to their respective native stream selectors, including WAV placement and mark restoration.
- Transition durations converted to QE seconds.frames format: 0.5 seconds at 24 fps is `0.12`; actual native duration is read back.
- JSON control-character escaping for Lumetri values and keys, including the already-loaded legacy polyfill.
- Audio write status handling backed by independent finite amplitude readback.
- Frame capture using actual QE timecode and an output stem without a duplicate PNG suffix, complete PNG chunk/checksum validation, and cleanup restricted to its owned temporary file.
- Explicit filenames/paths resolved before fuzzy matching, with missing/ambiguous references rejected instead of substituting unrelated media.
- Matched assets and source/timeline ranges retained across Go/Python protobuf conversions; opaque scan IDs mapped to exact paths only in the execution copy. Failed or partial execution cannot be reported as complete or proceed to export.
- EDL acceptance verifies a distinct native clip for each entry by source, track, source trims and timeline position, using frame rate from the actual project sequence metadata rather than a nonexistent timeline field.
- Project creation/opening verifies native status, unique document identity and exact path. An open but unfocused target reports an error with evidence and focus instructions, preventing edits to the previous project. Previously open projects are not closed. Native retests at `537fcfb` verify active already-open success and inactive already-open failure while preserving all six disposable documents.

The final run imported two media files, created a 320 × 180 / 24 fps sequence with two trimmed video placements and one WAV placement, applied a 0.5-second transition and Gaussian Blur, read back Exposure 0.5 and audio −6 dB, and fully decoded the **467,677-byte H.264/AAC export** (four seconds) and **38,584-byte PNG** at one second. The script pipeline matched the explicit fixture video, generated one EDL entry and created a separate 1920 × 1080 / 24 fps sequence with the exact source/track/trims/positions.

The full report SHA-256 is `7a85fe752637dadec510e3fbb249efdfdd8360c55582be9ab902d7f9eea58292`; the [compact acceptance evidence](verification/native-core-2026-09-30.json) records source revisions and output hashes with private paths removed. Full reports, media and failed-run evidence are preserved outside Git. This verifies one setup and API/output correctness; it does not certify every tool, OS/version/locale, perceptual quality or performance.

Project close is separately limited: the disposable document closed, but CEP disconnected before its RPC reply; the acknowledgement remains unverified.

## Local validation and publication boundary

At `537fcfb`, the final local `just ci` run passed **509 individual non-Go tests** (Rust 36, Python 117, TypeScript 21, CEP 288, CLI 4, npm distribution 16, E2E runner 22, UXP inspection 5), with **0 failures / 0 skips**, plus **7 Go test packages** and lint/build/protobuf/932 command-workflow checks. A separate uncached `go test -race -count=1 ./...` passed all seven Go test packages. CI log SHA-256: `8a0343d33717d05f6e4f5f3875555a3a856321cf8e2fd27e81db801eda103c23`; race log SHA-256: `97ec104da556d8712f707028642e8d3b3867fba79d68f4555a40d1aa31f9b17c`. Fresh production npm audits for CLI, bridge and CEP each report zero known vulnerabilities. These automated results are separate from the eight actual native acceptance checks. GitHub CI and merge are still pending at this snapshot.

No npm credentials or registry/GitHub publisher settings were changed, and no package was published. Package publication and its required account setup remain separate from committing the release workflow. Four binary targets were cross-built and their executable formats verified; only the macOS arm64 packaged wrapper was executed in a real MCP stdio smoke test.

After the validated integration is merged, administratively retiring #2/#3 and completing #6/#9 would leave **5 open issues**: #1, #4, #7, #8 and #10. Until those actions occur, the confirmed issue state remains **9 open / 2 closed**.
