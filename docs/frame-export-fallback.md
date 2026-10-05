# Raw source frame recovery

`premiere_export_frame` normally exports the rendered Premiere frame. Older Premiere versions can fail native export even when the QE method exists. `allow_source_fallback: true` permits a **raw source-file PNG** when native export fails. It defaults to false. `premiere_capture_frame` remains a native timeline capture and never substitutes raw footage.

Example:

```json
{"output_path":"/absolute/existing-directory/frame.png","format":"PNG","allow_source_fallback":true}
```

On source recovery the result has `status: "source_file_frame"`, a warning, `source_path`, `source_seconds` and `timeline_seconds`. Zero times may be omitted by JSON serialization. The image excludes Premiere effects, color corrections, titles, overlays, transitions, reframing and composition. It is not proof of timeline appearance.

The host resolves the source without moving the playhead. It must identify exactly one active normal-speed forward video clip with a real online source file. Source time includes the clip's source in-point and the elapsed timeline time. Gaps, overlapping video, nested sequences, multicam, transitions and uncertain speed/mapping are refused. Go requires explicit mapped source and timeline times; missing metadata never defaults to an invented mapping.

Recovery requires FFmpeg on PATH, a local regular source file, PNG format and an existing output directory. It uses a private temporary directory, deletes temporary output, checks that FFmpeg produced PNG data, and refuses existing destination files or destination symbolic links. FFmpeg runs with request cancellation and bounded diagnostic output. Recovery does not overwrite an existing native output; select a new destination if native export left a partial file.

Tests exercise actual FFmpeg extraction from a two-color video, including the second-color source offset, end-of-file refusal and decoding failure. Additional tests cover cancellation, missing FFmpeg, output collision and symbolic-link protection, native failure without opt-in and missing host mapping. Native recovery on older Premiere remains unverified; automated tests cannot prove compatibility with that host.
