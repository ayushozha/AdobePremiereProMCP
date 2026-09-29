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
