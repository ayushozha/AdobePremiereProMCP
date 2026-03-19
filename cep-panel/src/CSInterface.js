/**
 * CSInterface.js - Adobe Common Extensibility Platform Interface
 *
 * PLACEHOLDER: This file must be replaced with the actual CSInterface.js
 * from Adobe's CEP SDK before the panel will function.
 *
 * Download or copy from:
 *   https://github.com/Adobe-CEP/CEP-Resources/tree/master/CEP_11.x/CSInterface.js
 *
 * The CSInterface library provides the bridge between the CEP panel's
 * JavaScript/Node.js environment and the host application's ExtendScript
 * engine. Key methods used by this panel:
 *
 *   - new CSInterface()                   - Create the interface instance
 *   - csInterface.evalScript(script, cb)  - Execute ExtendScript and get result
 *   - csInterface.addEventListener(type, handler) - Listen for app events
 *   - csInterface.getHostEnvironment()    - Get host app info
 *   - csInterface.getCurrentApiVersion()  - Get CEP API version
 *
 * Installation:
 *   1. Clone or download: https://github.com/Adobe-CEP/CEP-Resources
 *   2. Copy CEP_11.x/CSInterface.js to this directory (src/CSInterface.js)
 *   3. Overwrite this placeholder file
 *
 * Alternatively, run:
 *   npm run setup-csinterface
 */

if (typeof CSInterface === "undefined") {
    console.warn(
        "[MCP Bridge] CSInterface.js is a placeholder. " +
        "Replace with the real library from Adobe CEP SDK. " +
        "See: https://github.com/Adobe-CEP/CEP-Resources"
    );

    // Minimal stub so the panel can load without immediately crashing.
    // It will NOT function correctly until the real library is in place.
    function CSInterface() {}
    CSInterface.prototype.evalScript = function (script, callback) {
        console.error("[CSInterface STUB] evalScript called but real library not loaded.");
        if (callback) callback("EvalScript error.");
    };
    CSInterface.prototype.addEventListener = function () {};
    CSInterface.prototype.getHostEnvironment = function () {
        return { appName: "STUB", appVersion: "0.0" };
    };
    CSInterface.prototype.getCurrentApiVersion = function () {
        return { major: 0, minor: 0, micro: 0 };
    };
}
