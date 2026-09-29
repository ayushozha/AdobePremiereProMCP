# Live Premiere E2E evidence

`scripts/e2e-premiere.mjs` exercises the real Go MCP server over stdio using the official SDK already declared in `cli/package.json`. It emits one JSON report with **pass / fail / blocked for all eight issue #6 checklist items**. Local unit tests and generated media are not evidence that Adobe executed an edit.

The default is a read-only preflight. Editing requires `--mutate`, a prepared fixture directory, an exact disposable project path, and an explicit disposable-project acknowledgement. The runner does not install panels, start/restart backend services, open/create/save/close projects, or delete project contents.

## Prerequisites

- Node 20+, the CLI SDK dependencies (`npm ci --prefix cli`), and a built Go binary (`just go-build`). `--server` can point to a separately built binary; `--sdk-root` can point to another checkout's `cli` directory with its dependencies installed.
- Existing Rust, Python, and TypeScript backend services, configured to use the intended Premiere/CEP installation. The runner only spawns its own stdio Go process and closes that child afterward. It does not restart shared services.
- Premiere Pro running with the CEP panel connected for native tests. Prefer the CEP bridge. The runner checks `premiere_is_running` before pinging because the standalone bridge's AppleScript ping can otherwise launch a stopped application. The current process-check tool uses `pgrep`; unsupported platforms, including a Windows host without it, conservatively block. This is not Windows compatibility proof.
- FFmpeg and FFprobe on `PATH` for fixture generation, export inspection, and full media decoding. `FFMPEG` and `FFPROBE` may specify absolute executable paths.
- For export: a real Adobe H.264 `.epr` preset with audio enabled. A preset name or queued job is not evidence of an export.

Export the existing service configuration into the runner's environment, including any required `BRIDGE_CEP_TOKEN` and service addresses. The runner deliberately does **not** source `.env` or copy values into reports. It forwards only the MCP/backend environment prefixes used by the CLI. Never paste credentials into evidence or reports.

## Read-only preflight

From the checkout root:

```sh
node scripts/e2e-premiere.mjs
```

To use an existing binary and save an evidence report to a new path:

```sh
node scripts/e2e-premiere.mjs \
  --server /absolute/path/to/go-orchestrator/bin/premierpro-mcp \
  --sdk-root /absolute/path/to/cli \
  --locale en_US \
  --report /absolute/path/to/new-preflight-report.json
```

Preflight enumerates tools, checks the Premiere OS process, calls `premiere_ping` only when the process exists, and calls `premiere_get_project` only when ping reports an open project. It never calls Lumetri getters: the existing getter can add an effect. It never captures frames or writes media. `--report` is the only optional local write in preflight; stdout otherwise contains the report. Existing report files are not overwritten.

All eight native checks remain **blocked** in preflight, even if connection health passes. Exit codes:

- `0`: read-only preflight passed, or all mutation checks passed.
- `1`: at least one attempted native check failed verification, or a requested report could not be saved.
- `2`: preflight/ownership prerequisites failed, or at least one mutation check was blocked and none failed.

Transport failure and missing tools still produce a report containing every checklist item. Missing tools report the required profile. The runner defaults its Go child to `MCP_TOOL_PROFILE=all` (without `unsafe`); an explicitly exported profile is respected.

## Prepare deterministic media

Choose a **new** directory. This command creates local files and never connects to Premiere:

```sh
node scripts/e2e-premiere.mjs --prepare /absolute/path/to/new-e2e-fixture
```

It creates:

- `media/e2e_test_pattern.mp4`: 8 seconds, H.264, 320 × 180, 24 fps, no audio.
- `media/e2e_tone.wav`: 8 seconds, 440 Hz tone, 48 kHz PCM audio.
- `script.txt` and `manifest.json`, including file SHA-256 hashes, media probe results, FFmpeg version, and the exact project path.

Both media files must fully decode before preparation succeeds. Identical inputs repeat byte-for-byte with the same FFmpeg build; compare recorded hashes when changing FFmpeg versions. Existing directories/files are never overwritten. Preparation prints the canonical path (for example `/private/tmp` instead of the macOS `/tmp` symlink); use the printed manifest path in subsequent commands.

## Run against a disposable project

1. In Premiere, manually create and save a **new, empty** project at the manifest's exact `project` path, ending in `MCP-E2E-disposable.prproj`. It must have zero project items and zero sequences. Do not reuse a production project or a previous run.
2. Leave that project open with its CEP panel connected. Do not interact with Premiere or switch projects/sequences while the runner is executing.
3. Run:

```sh
MCP_TOOL_PROFILE=all node scripts/e2e-premiere.mjs \
  --mutate \
  --fixture-dir /absolute/canonical/path/to/new-e2e-fixture \
  --project /absolute/canonical/path/to/new-e2e-fixture/MCP-E2E-disposable.prproj \
  --confirm-disposable MCP-E2E-DISPOSABLE \
  --preset /absolute/path/to/H264-with-audio.epr \
  --locale en_US \
  --report /absolute/path/to/new-native-report.json
```

`--effect` and `--transition` accept the exact installed localized display names; defaults are `Gaussian Blur` and `Cross Dissolve`. The Lumetri check currently expects the English `Exposure` readback property and must fail or block on a host that cannot expose it. `--locale` is an operator declaration, clearly labeled in the report; it is not automatically measured.

The runner checks the manifest hashes and rejects symlinked assets/projects, an unrelated active project, existing project contents, or an existing `artifacts` directory. Before each editing call it rechecks the project path and owned sequence IDs. Operations on an active sequence also check its exact ID. A project/sequence change stops further edits. A transport failure or timeout during an edit also stops further edits because the native call may still be running. These are read-before-write checks, not an atomic Adobe transaction; keep the application untouched during the run.

Successful sequence creation uses `MCP-E2E-fixture`. It places two trimmed video clips at 0–2 and 2–4 seconds and one audio clip at 0–4 seconds, leaving source handles for a 0.5-second transition. The script workflow runs last and may create another sequence in the same disposable project. It omits `auto_edit.output_name`, which currently triggers an export rather than simply naming a sequence.

Failures may leave partial edits in the disposable project. The runner deliberately does not attempt rollback, save, delete, or retry destructive work. Start a fresh fixture directory and empty project for the next run. Inspect/save the disposable project manually if needed for review.

## What earns a pass

| Checklist item | Required evidence |
| --- | --- |
| Video/audio import | Both exact fixture paths read back from project items; returned import IDs alone are insufficient. |
| Sequence/placement | Correct created sequence name/ID, 320 × 180 at 24 fps, two video clips and one audio clip, exact sources, source trims, timeline positions, and 4-second duration. |
| Transitions/effects | Requested transition and 0.5-second duration in native transition readback; requested effect in native component readback. |
| Lumetri | Exposure set to 0.5 and independently read back within 0.001. |
| Audio levels | Fixture audio clip set to −6 dB and independently read back within 0.1 dB. |
| H.264 export | New output file from direct Adobe export, H.264 video and audio streams, duration within 0.25 seconds of 4, full FFmpeg decode, file and preset hashes. |
| Base64 frame | MCP PNG image block, valid PNG header/end marker, 320 × 180 dimensions, playhead time of 1 second, saved image hash, full FFmpeg decode. |
| Script → EDL → timeline | Non-empty parsed segments and edit decision list, completed execution step without errors, a new native sequence, matching EDL/execution/timeline clip counts, fixture-only source paths. |

The JSON contains timestamps, runner-host OS/architecture/Node (remote backend hosts may differ), reported Premiere version and bridge mode, tool count/profile, local fixture metadata, each actual MCP call/result, verification evidence, and summary counts. Image bytes are saved under `artifacts/frame.png`; reports contain their hash and byte count instead of duplicating base64. Export is saved under `artifacts/export.mp4`. Review paths/project names before sharing reports.

These checks establish API readback and decodable output. They do not assess visual taste, audio listening quality, perceptual correctness of a grade/transition, hardware performance, or general support for every Premiere build/locale. Review the resulting frame/video manually for those claims. Record one native report for each supported OS/Premiere/locale combination; do not infer a compatibility matrix from a single machine or tool count.

## Local validation and current evidence

```sh
node --test scripts/e2e-premiere.test.mjs
(cd go-orchestrator && go test ./internal/orchestrator -run TestGetProjectItems -count=1)
```

The runner tests exercise safeguards, error accounting, native-shaped response parsing, timeline rejection cases, and actual generated media decoding/hashes. The MCP boundary in these tests is deliberately a test double, never a live Adobe claim. The real-media fixture test is explicitly skipped if FFmpeg/FFprobe are unavailable. Go regression tests cover CEP's `mediaPath`/`itemCount`/`binPath`/`childCount` fields and preserve the public snake_case response schema.

On 2026-09-29, an actual stdio preflight against the built Go server and installed services discovered 1,033 tools, then reported `premiere_running: false`, `premiere_version: "unknown"`, `project_open: false`, `bridge_mode: "cep"`. The report had **0 pass, 0 fail, 8 blocked** native checks and exited 2. The application was not reachable, so no native mutation/export/capture was attempted. This is connection/blocker evidence only; issue #6 and the Premiere compatibility portion of issue #4 remain unverified until a disposable live session completes.
