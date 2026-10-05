"use strict";

// Adobe and UXP storage are the external boundaries; no project mutations.
async function exportSourceTranscript(premiere, storage) {
  for (const [object, method] of [
    [premiere && premiere.Project, "getActiveProject"],
    [premiere && premiere.ProjectUtils, "getSelection"],
    [premiere && premiere.ClipProjectItem, "cast"],
    [premiere && premiere.Transcript, "exportToJSON"],
    [storage && storage.localFileSystem, "getFileForSaving"],
  ]) {
    if (!object || typeof object[method] !== "function") {
      throw new Error("Required API unavailable: " + method + " (Premiere 25.6+ required)");
    }
  }
  const project = await premiere.Project.getActiveProject();
  if (!project) throw new Error("No project is open");
  const selection = await premiere.ProjectUtils.getSelection(project);
  if (!selection || typeof selection.getItems !== "function") {
    throw new Error("Project panel selection API unavailable");
  }
  const items = await selection.getItems();
  if (!Array.isArray(items) || items.length !== 1) {
    throw new Error("Select exactly one source media clip in the Project panel");
  }
  const clip = premiere.ClipProjectItem.cast(items[0]);
  if (!clip || typeof clip.isSequence !== "function" || await clip.isSequence()) {
    throw new Error("Select a source media clip, rather than a bin or sequence");
  }
  if (typeof premiere.Transcript.hasTranscript === "function" &&
      !premiere.Transcript.hasTranscript(clip)) {
    throw new Error("The selected source clip has no existing transcript");
  }
  const raw = await premiere.Transcript.exportToJSON(clip);
  if (typeof raw !== "string" || !raw.trim()) {
    throw new Error("Premiere returned an empty existing transcript");
  }
  let transcript;
  try { transcript = JSON.parse(raw); } catch (_) {
    throw new Error("Premiere returned malformed transcript JSON");
  }
  if (!transcript || !Array.isArray(transcript.segments) ||
      transcript.segments.length === 0 || !Array.isArray(transcript.speakers)) {
    throw new Error("Transcript JSON must contain non-empty segments and speakers");
  }
  const file = await storage.localFileSystem.getFileForSaving(
    "source-transcript.json", { types: ["json"] });
  if (!file) return { status: "cancelled" };
  // Preserve Adobe's exact JSON string, including unknown schema fields.
  await file.write(raw, { format: storage.formats.utf8 });
  if (await file.read({ format: storage.formats.utf8 }) !== raw) {
    throw new Error("Transcript file readback did not match; the file may have been written incompletely");
  }
  return {
    status: "exported",
    source: "source_clip_transcript",
    clipName: clip.name,
    outputPath: file.nativePath,
    segmentCount: transcript.segments.length,
    verified: true,
  };
}

module.exports = { exportSourceTranscript };
