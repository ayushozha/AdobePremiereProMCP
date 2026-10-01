# Tutorial recording plan

Preparation for [issue #10](https://github.com/ayushozha/AdobePremiereProMCP/issues/10). **No tutorial videos have been recorded or published.** These recording outlines are ready to use once the live validation checklist passes in a disposable project. Do not present a scripted or mocked response as a successful Premiere edit.

Use generated fixture media from the live-test runner so the demonstrations can be repeated. Record the Premiere version, OS, repository commit, tool profile, original capture, final output, and any failed tool call for every tutorial. Use actual discovered tool schemas for arguments rather than relying on a narrated command.

## 1. Getting started — target 5 minutes

- 0:00–1:30: Show prerequisites, checkout, dependencies, backend services, and CEP panel installation. Show the exact platform being tested.
- 1:30–2:30: Connect the MCP client, run `premiere_ping`, and inspect the active disposable project with `premiere_get_project`.
- 2:30–4:15: Import fixture media, create a sequence, place a clip, then read back the timeline. Show the same clip in Premiere's Program Monitor.
- 4:15–5:00: Explain disconnected-panel diagnostics and show a successful read-only recheck.

Evidence: non-empty project name, fixture media paths, created sequence ID, clip timing readback, and a visible matching timeline. A running MCP process alone is not the completion point.

## 2. Auto-edit pipeline — target 10 minutes

- 0:00–2:00: Show the fixture assets and a short original script with named shots and durations.
- 2:00–4:00: Discover `premiere_parse_script` and `premiere_auto_edit` in the active profile; parse the script and inspect segments. If either tool is absent, configure the required profile before recording.
- 4:00–7:00: Run the pipeline only in the disposable project. Inspect the resulting EDL and timeline; check source paths and every clip's start/end.
- 7:00–9:00: Play the edit in Premiere and identify any manual corrections required.
- 9:00–10:00: Show the saved evidence report and resulting media output if an export was actually completed.

Evidence: script, parsed segments, EDL, matching native timeline, and playable output. Stop and disclose unsupported pipeline steps instead of editing together a simulated success.

## 3. Color grading — target 5 minutes

- 0:00–1:00: Use a fixture with visible dark/light/color patches; show its starting frame.
- 1:00–2:30: Discover the Lumetri tool and parameters. Apply one measurable correction, such as a specific exposure value, in a dedicated demo sequence.
- 2:30–4:00: Read back the effect and parameter; compare before/after frames and scopes inside Premiere.
- 4:00–5:00: Demonstrate recovery or undo and state the host version on which it was tested.

Evidence: effect name, parameter value readback, before/after frames, and a working recovery. A success envelope without changed pixels is insufficient.

## 4. Social media export — target 3 minutes

- 0:00–1:00: Show two prepared sequences with verified landscape and portrait dimensions, then inspect their content framing.
- 1:00–2:00: Select explicit installed `.epr` presets and export to unique paths. Keep the full export recording; speed it up visibly if necessary.
- 2:00–3:00: Probe both outputs for codec, dimensions, duration, and audio, then play each exported file.

Evidence: two non-empty, decodable video files with expected dimensions. An AME queue ID alone does not prove an export completed.

## Completion record

For each tutorial, attach the capture/video path or published URL, runtime, caption file, tested commit and host version, and the associated live evidence report. Until all four playable videos exist, keep #10 open.
