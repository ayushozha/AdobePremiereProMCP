#!/usr/bin/env node

/**
 * Validates the CEP panel structure and reports any issues.
 */

var fs = require("fs");
var path = require("path");

var ROOT = path.join(__dirname, "..");

var checks = [
    { path: "CSXS/manifest.xml", label: "CEP manifest" },
    { path: ".debug", label: "Debug configuration" },
    { path: "src/index.html", label: "Panel HTML" },
    { path: "src/panel.js", label: "Panel JavaScript" },
    { path: "src/CSInterface.js", label: "CSInterface library" },
    { path: "src/host/premiere.jsx", label: "ExtendScript host" },
    { path: "package.json", label: "Package config" },
    { path: "node_modules/ws", label: "ws dependency (run npm install)" },
];

var allOk = true;

console.log("PremierPro MCP Bridge - CEP Panel Validation");
console.log("=============================================\n");

checks.forEach(function (check) {
    var fullPath = path.join(ROOT, check.path);
    var exists = fs.existsSync(fullPath);
    var status = exists ? "OK" : "MISSING";
    var icon = exists ? "[+]" : "[-]";

    if (!exists) allOk = false;

    console.log("  " + icon + " " + check.label + " (" + check.path + ") ... " + status);
});

// Check if CSInterface.js is the placeholder or the real library
var csPath = path.join(ROOT, "src", "CSInterface.js");
if (fs.existsSync(csPath)) {
    var content = fs.readFileSync(csPath, "utf8");
    if (content.indexOf("PLACEHOLDER") !== -1 || content.indexOf("CSInterface STUB") !== -1) {
        console.log("\n  [!] CSInterface.js is still the placeholder. Run: npm run setup-csinterface");
        allOk = false;
    }
}

// Check symlink
var os = require("os");
var symlinkPath = path.join(
    os.homedir(),
    "Library",
    "Application Support",
    "Adobe",
    "CEP",
    "extensions",
    "com.premierpro.mcp.bridge"
);

if (process.platform === "darwin") {
    if (fs.existsSync(symlinkPath)) {
        console.log("\n  [+] Panel symlink exists at:\n      " + symlinkPath);
    } else {
        console.log("\n  [-] Panel symlink not found. Run: npm run link-panel:mac");
        allOk = false;
    }
}

// Check debug mode
if (process.platform === "darwin") {
    var exec = require("child_process").execSync;
    try {
        var debugMode = exec("defaults read com.adobe.CSXS.11 PlayerDebugMode 2>/dev/null").toString().trim();
        if (debugMode === "1") {
            console.log("  [+] PlayerDebugMode is enabled (CSXS.11)");
        } else {
            console.log("  [-] PlayerDebugMode is not set to 1. Run: npm run enable-debug:mac");
            allOk = false;
        }
    } catch (e) {
        console.log("  [-] PlayerDebugMode not configured. Run: npm run enable-debug:mac");
        allOk = false;
    }
}

console.log("\n" + (allOk ? "All checks passed." : "Some checks failed. See above.") + "\n");
process.exit(allOk ? 0 : 1);
