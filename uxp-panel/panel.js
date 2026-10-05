"use strict";

const { inspectProject } = require("./inspect.js");
const button = document.getElementById("inspect");
const output = document.getElementById("output");

button.addEventListener("click", async () => {
  button.disabled = true;
  output.textContent = "Reading project…";
  try {
    const report = await inspectProject(require("premierepro"), require("uxp").host);
    output.textContent = JSON.stringify(report, null, 2);
  } catch (error) {
    output.textContent = "Inspection failed: " + String(error && error.message || error);
  } finally {
    button.disabled = false;
  }
});

const { exportSourceTranscript } = require("./transcript-export.js");
const exportButton = document.getElementById("export-transcript");
exportButton.addEventListener("click", async () => {
  exportButton.disabled = true;
  button.disabled = true;
  output.textContent = "Reading the selected source clip transcript…";
  try {
    const result = await exportSourceTranscript(require("premierepro"), require("uxp").storage);
    output.textContent = result.status === "cancelled"
      ? "Export cancelled."
      : "Source clip transcript exported and file content verified.\n" + JSON.stringify(result, null, 2);
  } catch (error) {
    output.textContent = "Source clip transcript export failed: " + String(error && error.message || error);
  } finally {
    exportButton.disabled = false;
    button.disabled = false;
  }
});
