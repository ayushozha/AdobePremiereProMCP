"use strict";

var path = require("path");
var identifier = /^[A-Za-z_$][A-Za-z0-9_$]*$/;

function resolveHostPaths(extensionRoot) {
    return {
        core: path.join(extensionRoot, "src", "host", "core.jsx"),
        premiere: path.join(extensionRoot, "src", "host", "premiere.jsx"),
    };
}

function requireIdentifiers(names) {
    names.forEach(function (name) {
        if (!identifier.test(name)) throw new Error("Invalid ExtendScript function name: " + name);
    });
}

function loadAndCheck(hostPath, names, onlyIfMissing) {
    requireIdentifiers(names);
    var missing = names.map(function (name) {
        return 'typeof ' + name + ' !== "function"';
    }).join(" || ");
    var filePath = JSON.stringify(hostPath.replace(/\\/g, "/"));
    var script = "";
    if (onlyIfMissing) script += "if (" + missing + ") { ";
    script += "var __premierHostFile = new File(" + filePath + "); " +
        "if (!__premierHostFile.exists) throw new Error(" +
        JSON.stringify("ExtendScript host file not found: " + hostPath) + "); " +
        "$.evalFile(__premierHostFile); " +
        "if (" + missing + ") throw new Error(" +
        JSON.stringify("ExtendScript host did not define required functions: " + names.join(", ")) + "); ";
    if (onlyIfMissing) script += "} ";
    return script;
}

function withErrors(script) {
    // Keep evaluation at top level so loaded declarations remain in the
    // extension's host context after this call has finished.
    return "try { " + script + " } catch (__premierHostError) { " +
        "JSON.stringify({success:false,error:String(__premierHostError.message || __premierHostError) + " +
        "(__premierHostError.line ? ' (line ' + __premierHostError.line + ')' : '')}); }";
}

function buildLoadScript(hostPath, requiredSymbols) {
    return withErrors(loadAndCheck(hostPath, requiredSymbols, false) +
        "JSON.stringify({success:true});");
}

function buildDispatchScript(hostPath, functionName, argsJson) {
    return withErrors(loadAndCheck(hostPath, ["mcpDispatch", functionName], true) +
        "mcpDispatch(" + JSON.stringify(functionName) + "," + JSON.stringify(argsJson || "{}") + ");");
}

module.exports = {
    resolveHostPaths: resolveHostPaths,
    buildLoadScript: buildLoadScript,
    buildDispatchScript: buildDispatchScript,
};
