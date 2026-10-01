"use strict";

async function inspectProject(premiere, host = {}) {
  if (!premiere || !premiere.Project ||
      typeof premiere.Project.getActiveProject !== "function") {
    throw new Error("This host does not expose UXP Project.getActiveProject");
  }
  const project = await premiere.Project.getActiveProject();
  const report = {
    schemaVersion: 1,
    bridgeMode: "uxp-inspection-prototype",
    status: "ok",
    premiereVersion: host.version || "unknown",
    projectOpen: Boolean(project),
    project: project ? { name: project.name, path: project.path } : null,
    sequences: [],
    warnings: [],
  };
  if (!project) return report;

  // Counts must remain unknown if a host method is unavailable or throws.
  // In particular, an absent caption API never means a sequence has no captions.
  const sequences = await project.getSequences();
  for (const sequence of sequences) {
    const row = { name: sequence.name };
    for (const [field, method] of [
      ["videoTracks", "getVideoTrackCount"],
      ["audioTracks", "getAudioTrackCount"],
      ["captionTracks", "getCaptionTrackCount"],
    ]) {
      row[field] = null;
      try {
        if (typeof sequence[method] !== "function") {
          throw new Error("API unavailable");
        }
        const count = await sequence[method]();
        if (!Number.isSafeInteger(count) || count < 0) {
          throw new Error("Invalid track count");
        }
        row[field] = count;
      } catch (error) {
        report.warnings.push(sequence.name + ": " + method + ": " +
          String(error && error.message || error));
      }
    }
    report.sequences.push(row);
  }
  if (report.warnings.length) report.status = "partial";
  return report;
}

module.exports = { inspectProject };
