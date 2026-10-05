# Existing source clip transcript export

The development UXP panel can save the existing transcript of **one source media clip selected in the Project panel** as raw Adobe JSON. It requires Premiere Pro 25.6 or newer and a clip that has already been transcribed. It preserves the JSON returned by Adobe, including source timing and speaker data; it does not generate a transcript or reconstruct an edited sequence transcript.

## Use

1. Load `uxp-panel/manifest.json` using the UXP Developer Tool and open the inspection panel.
2. Open a project and select exactly one transcribed source media clip in the Project panel. Timeline selection is not used. Bins and sequences are rejected.
3. Click **Export source clip transcript JSON** and choose a JSON output file in the save dialog. The save dialog controls the destination and overwrite confirmation.
4. The panel reports success only after writing and reading back the same JSON. Cancellation and errors are displayed separately. A write or readback error may leave an incomplete file at the chosen destination.

This is a local panel action, **not an MCP tool**. The panel has no WebSocket/network connection or arbitrary-path filesystem access. Its manifest requests file-picker access (`localFileSystem: "request"`).

## Documented APIs

- [`ProjectUtils.getSelection(project)`](https://developer.adobe.com/premiere-pro/uxp/ppro-reference/classes/projectutils/) and the selection's `getItems()` read Project panel selection. Adobe's [project-panel sample](https://github.com/AdobeDocs/uxp-premiere-pro-samples/blob/main/sample-panels/premiere-api/src/projectPanel.ts) demonstrates this flow.
- [`ClipProjectItem.cast` and `isSequence`](https://developer.adobe.com/premiere-pro/uxp/ppro-reference/classes/clipprojectitem/) identify the source clip and exclude sequences; both are documented since 25.6.
- [`Transcript.exportToJSON`](https://developer.adobe.com/premiere-pro/uxp/ppro-reference/classes/transcript/) returns existing clip transcript JSON since 25.6. `hasTranscript` is optional and used only when exposed: Adobe introduced it in 26.3. No transcription-generation API is called.
- [`getFileForSaving`](https://developer.adobe.com/premiere-pro/uxp/uxp-api/reference-js/modules/uxp/persistent-file-storage/file-system-provider) chooses a writable file. [Filesystem permission guidance](https://developer.adobe.com/uxp/guides/how-to/recipes/filesystem-operations/) explains picker access.
- Adobe publishes the [transcript JSON schema](https://github.com/AdobeDocs/uxp-premiere-pro-samples/blob/main/sample-panels/premiere-api/assets/transcript_format_spec.json). The exporter checks for parseable JSON with non-empty segments and a speakers array; it preserves unknown fields rather than converting or claiming full schema validation.

## Validation

Run `node --test uxp-panel/test/*.test.js`. Tests use mocked Adobe/picker boundaries and real temporary-file writes/readback. They cover exact JSON preservation on a 25.6-style API without `hasTranscript`, selection errors, missing APIs, absent/empty/malformed transcripts, host/picker/write failures, cancellation, and readback mismatch.

Live Premiere loading, native selection behavior, native transcript contents, and operating-system save/overwrite dialogs still require manual host validation. Automated tests do not prove runtime compatibility inside Premiere. This scope does not claim `.prtranscript`, caption extraction, source-to-sequence timing conversion, or MCP-controlled export.
