#!/usr/bin/env node

/**
 * Assemble the static CEP extension into dist/ without requiring a bundler.
 * CEP loads the manifest, HTML, JavaScript, and ExtendScript files directly.
 */

var childProcess = require("child_process");
var fs = require("fs");
var path = require("path");

var root = path.join(__dirname, "..");
var output = path.join(root, "dist");
var packageMode = process.argv.indexOf("--package") !== -1;
var developmentMode = process.argv.indexOf("--dev") !== -1;

var requiredPaths = [
    "CSXS/manifest.xml",
    "src/index.html",
    "src/panel.js",
    "src/CSInterface.js",
    "src/host/core.jsx",
    "src/host/premiere.jsx",
    "package.json",
    "package-lock.json",
    "node_modules/ws",
];

if (developmentMode) {
    requiredPaths.push(".debug");
}

requiredPaths.forEach(function (relativePath) {
    var source = path.join(root, relativePath);
    if (!fs.existsSync(source)) {
        throw new Error("Missing required CEP asset: " + relativePath);
    }
});

// Catch ordinary JavaScript syntax errors without attempting to parse
// ExtendScript-only syntax in the host .jsx files.
childProcess.execFileSync(process.execPath, ["--check", path.join(root, "src", "panel.js")], {
    stdio: "inherit",
});

fs.rmSync(output, { recursive: true, force: true });
fs.mkdirSync(output, { recursive: true });

["CSXS", "src", "node_modules/ws"].forEach(function (relativePath) {
    fs.cpSync(path.join(root, relativePath), path.join(output, relativePath), {
        recursive: true,
    });
});

["package.json", "package-lock.json"].forEach(function (relativePath) {
    fs.copyFileSync(path.join(root, relativePath), path.join(output, relativePath));
});

// CEP's .debug file opens a local Chrome DevTools endpoint. Keep it out of
// production builds and packages; developers must opt in with --dev.
if (developmentMode) {
    fs.copyFileSync(path.join(root, ".debug"), path.join(output, ".debug"));
}

console.log(
    (developmentMode ? "Built development" : packageMode ? "Packaged" : "Built") +
        " static CEP extension at " +
        path.relative(process.cwd(), output),
);
