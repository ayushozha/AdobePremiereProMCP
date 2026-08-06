# Workflow Research and Adoption Notes

Research date: 2026-08-06

This review focused on workflow patterns that make editing faster, safer, and
more repeatable. No third-party source code was copied or vendored. The project
adopts ideas at the workflow/architecture level and keeps each integration
behind its own dependency and license boundary.

## Adopted in This Repository

| Source | Useful pattern | What was added here |
|---|---|---|
| [Auto-Editor](https://github.com/WyattBlue/auto-editor) | Compose audio/motion detectors, keep configurable margins around speech, and generate an edit plan before mutation. | The `premiere-dialogue-cut` skill resolves used source ranges, calls decoded waveform analysis, maps silence to timeline time, keeps speech margins, requires plan approval, and verifies the resulting edit. It explicitly rejects the legacy CEP-only silence placeholder as evidence. |
| [OpenTimelineIO](https://github.com/AcademySoftwareFoundation/OpenTimelineIO) | Treat editorial decisions as a rational-time timeline model with media references and adapter boundaries. | EDL validation now checks source ranges, sequence bounds, overlaps, frame rate/timecode assumptions, and assembly readback before accepting a result. |
| [faster-whisper](https://github.com/SYSTRAN/faster-whisper) and [WhisperX](https://github.com/m-bain/whisperX) | Batch transcription, voice-activity filtering, word timestamps/alignment, and optional diarization. | The `premiere-captions` skill accepts a supplied timed SRT, validates it, imports it through Premiere's caption API, and checks the resulting caption count. It does not pretend transcription or FCC compliance was performed. |
| [PySceneDetect](https://github.com/Breakthrough/PySceneDetect) | Detect scene boundaries first, then let an editor review or transform those boundaries. | The Rust media service exposes scene detection, while dialogue/social skills analyze before mutation and preserve an untouched sequence or derivative as the recovery boundary. |
| [LosslessCut](https://github.com/mifi/lossless-cut) | Edit against lightweight media while preserving original media for final output. | The `premiere-proxy-conform` skill creates/attaches proxies, verifies proxy state, and requires an original-media conform check before delivery. |
| [gl-transitions](https://github.com/gl-transitions/gl-transitions) | Give transitions normalized duration/parameter defaults and treat the transition as a reusable recipe. | The `premiere-transition-pack` skill applies installed transitions through a small verified surface and requires name/duration readback. |
| [Av1an](https://github.com/rust-av/Av1an) | Split long work into bounded jobs, use worker limits, and make progress resumable/observable. | The `premiere-batch-delivery` skill bounds direct-export concurrency to one job, records AME submissions as pending, and probes stable output files. CEP does not provide a trustworthy per-job AME poller, so the workflow says so instead of inventing progress. |
| [Remotion](https://github.com/remotion-dev/remotion) | Reuse design systems/templates and batch-render variants from one content plan. | MOGRT-based title prompting, social variants, and delivery matrices reuse named templates and explicit output specifications. |
| [hetpatel-11/Adobe_Premiere_Pro_MCP](https://github.com/hetpatel-11/Adobe_Premiere_Pro_MCP) | Setup diagnostics, live tool sweeps, and polished high-level workflows are more useful than tool-count claims. | Added an MCP protocol smoke runner, smaller curated/readback-first tool profiles, workflow resources/prompts, and seven reusable Agent Skills. |
| [leancoderkavy/premiere-pro-mcp](https://github.com/leancoderkavy/premiere-pro-mcp) | Organize a broad command surface around discoverable editorial jobs. | Added named profiles for dialogue, captions, social, transitions, effects, proxies, and delivery instead of exposing the full registry by default. |

## Deliberately Not Bundled Yet

- **Automatic speech recognition:** faster-whisper/WhisperX are strong optional
  backends, but model downloads, GPU/CPU choices, diarization credentials, and
  language-specific quality need an explicit integration contract. The current
  caption workflow starts from timed SRT and reports that boundary honestly.
- **OTIO import/export:** the current EDL model has stricter validation, but an
  OpenTimelineIO adapter still needs round-trip tests for Premiere XML/FCP XML,
  transitions, nested sequences, speed changes, and drop-frame timecode.
- **Adaptive scene detection:** the existing detector is useful for analysis;
  PySceneDetect-style selectable content/adaptive/fade detectors remain a
  candidate once fixtures establish accuracy and performance thresholds.
- **Lossless smart cutting:** direct stream-copy cuts and hybrid boundary
  re-encoding do not belong in Premiere's host bridge. They would be a separate
  media-engine feature with codec/GOP validation.
- **GPU shader transition packs:** gl-transitions defines a portable model, but
  Premiere still needs an installed effect/plugin that the public DOM can
  discover and verify. The server does not inject unlicensed shaders.
- **Per-scene distributed encoding:** Av1an's encoder orchestration is not copied.
  Delivery can run blocking direct exports or submit Premiere/AME jobs, but CEP
  cannot observe an individual AME job reaching a terminal state. Verification
  begins only after an external check observes a stable output file.

## Acceptance Rule for New Workflow Integrations

A workflow is a candidate for the default `standard` profile only when it has:

1. deterministic input validation and an explicit preflight;
2. a verified duplicate sequence or tool-created derivative before timeline mutation; JSON timeline snapshots and saved sequence versions count only as audit records, not recovery copies;
3. public-DOM or backend readback proving the requested state changed;
4. fixture/unit coverage plus MCP schema and required-argument checks; and
5. a disposable-project live sweep on every Premiere/OS combination claimed.

Meeting the first four items is not live certification. Until the fifth item is
run, the repository calls a workflow "readback-first" but does not call it
live-certified for that Premiere/OS combination.
