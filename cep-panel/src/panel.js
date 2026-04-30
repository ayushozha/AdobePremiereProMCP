/**
 * PremierPro MCP Bridge - CEP Panel
 *
 * Runs inside Adobe Premiere Pro as a CEP panel. Opens a WebSocket server
 * that the ts-bridge gRPC server connects to. Incoming JSON commands are
 * routed to ExtendScript functions via CSInterface.evalScript(), and the
 * results are sent back over the WebSocket.
 *
 * Command format (inbound):
 *   { "action": "getProjectState", "params": {}, "requestId": "uuid" }
 *
 * Response format (outbound):
 *   { "requestId": "uuid", "success": true, "result": {...} }
 *   { "requestId": "uuid", "success": false, "error": "message" }
 */

(function () {
    "use strict";

    // ---------------------------------------------------------------------------
    // Node.js modules (available because --enable-nodejs is set in the manifest)
    // ---------------------------------------------------------------------------
    var WebSocketServer = require("ws").Server;
    var path = require("path");
    var fs = require("fs");

    // ---------------------------------------------------------------------------
    // Configuration
    // ---------------------------------------------------------------------------
    var DEFAULT_PORT = 9801;
    var RECONNECT_DELAY_MS = 3000;
    var HEARTBEAT_INTERVAL_MS = 15000;
    var VERSION = "1.0.0";

    // ---------------------------------------------------------------------------
    // State
    // ---------------------------------------------------------------------------
    var csInterface = new CSInterface();
    var wss = null;
    var activeConnections = new Set();
    var heartbeatTimer = null;
    var serverPort = DEFAULT_PORT;
    var autoScroll = true;

    // Lazy-loading state for the full premiere.jsx ExtendScript library
    var premiereJsxLoaded = false;
    var premiereJsxPath = resolveHostScriptPath("premiere.jsx");

    // Stats tracking
    var stats = {
        commandsExecuted: 0,
        errorsCount: 0,
        lastCommand: "",
        totalResponseTime: 0,
        responseCount: 0,
        startTime: Date.now(),
    };

    // In-flight command timers: requestId -> start timestamp
    var commandTimers = {};

    function resolveHostScriptPath(fileName) {
        var candidates = [
            path.join(__dirname, "host", fileName),
            path.join(__dirname, "src", "host", fileName),
        ];
        for (var i = 0; i < candidates.length; i++) {
            if (fs.existsSync(candidates[i])) {
                return candidates[i].replace(/\\/g, "/");
            }
        }
        return candidates[0].replace(/\\/g, "/");
    }

    // ---------------------------------------------------------------------------
    // UI element references
    // ---------------------------------------------------------------------------
    var logArea = document.getElementById("log-area");
    var statusDot = document.getElementById("status-dot");
    var statusLabel = document.getElementById("status-label");
    var portDisplay = document.getElementById("port-display");
    var clientCount = document.getElementById("client-count");
    var uptimeDisplay = document.getElementById("uptime-display");
    var statCommands = document.getElementById("stat-commands");
    var statAvgTime = document.getElementById("stat-avg-time");
    var statLastCmd = document.getElementById("stat-last-cmd");
    var statErrors = document.getElementById("stat-errors");
    var statUptimeShort = document.getElementById("stat-uptime-short");
    var themeSwitch = document.getElementById("theme-switch");
    var themeIconLabel = document.getElementById("theme-icon-label");
    var btnAutoscroll = document.getElementById("btn-autoscroll");
    var btnClearLog = document.getElementById("btn-clear-log");
    var headerVersion = document.getElementById("header-version");
    var footerVersion = document.getElementById("footer-version");
    var githubLink = document.getElementById("github-link");

    // ---------------------------------------------------------------------------
    // Version display
    // ---------------------------------------------------------------------------
    if (headerVersion) headerVersion.textContent = "v" + VERSION;
    if (footerVersion) footerVersion.textContent = "v" + VERSION;

    // ---------------------------------------------------------------------------
    // Theme management
    // ---------------------------------------------------------------------------
    function setTheme(theme) {
        document.documentElement.setAttribute("data-theme", theme);
        if (themeSwitch) themeSwitch.checked = (theme === "light");
        if (themeIconLabel) {
            // Moon for dark, sun for light
            themeIconLabel.innerHTML = theme === "dark" ? "&#9790;" : "&#9788;";
        }
        try {
            localStorage.setItem("mcp-bridge-theme", theme);
        } catch (e) { /* storage may not be available */ }
    }

    function loadTheme() {
        var saved = null;
        try {
            saved = localStorage.getItem("mcp-bridge-theme");
        } catch (e) { /* ignore */ }
        setTheme(saved || "dark");
    }

    if (themeSwitch) {
        themeSwitch.addEventListener("change", function () {
            setTheme(this.checked ? "light" : "dark");
        });
    }

    loadTheme();

    // ---------------------------------------------------------------------------
    // Auto-scroll toggle
    // ---------------------------------------------------------------------------
    if (btnAutoscroll) {
        btnAutoscroll.addEventListener("click", function () {
            autoScroll = !autoScroll;
            this.classList.toggle("active", autoScroll);
        });
    }

    // ---------------------------------------------------------------------------
    // Clear log
    // ---------------------------------------------------------------------------
    if (btnClearLog) {
        btnClearLog.addEventListener("click", function () {
            while (logArea.firstChild) {
                logArea.removeChild(logArea.firstChild);
            }
            log("Log cleared", "info");
        });
    }

    // ---------------------------------------------------------------------------
    // GitHub link
    // ---------------------------------------------------------------------------
    if (githubLink) {
        githubLink.addEventListener("click", function (e) {
            e.preventDefault();
            // Open in default browser via CSInterface if possible
            if (csInterface && csInterface.openURLInDefaultBrowser) {
                csInterface.openURLInDefaultBrowser("https://github.com/ayushozha/AdobePremiereProMCP");
            }
        });
    }

    // ---------------------------------------------------------------------------
    // Uptime tracking
    // ---------------------------------------------------------------------------
    var uptimeTimer = setInterval(function () {
        var elapsed = Math.floor((Date.now() - stats.startTime) / 1000);
        var h = Math.floor(elapsed / 3600);
        var m = Math.floor((elapsed % 3600) / 60);
        var s = elapsed % 60;

        var hh = h < 10 ? "0" + h : "" + h;
        var mm = m < 10 ? "0" + m : "" + m;
        var ss = s < 10 ? "0" + s : "" + s;

        if (uptimeDisplay) uptimeDisplay.textContent = hh + ":" + mm + ":" + ss;

        // Short uptime for stats card
        if (statUptimeShort) {
            if (h > 0) {
                statUptimeShort.textContent = h + "h" + m + "m";
            } else if (m > 0) {
                statUptimeShort.textContent = m + "m";
            } else {
                statUptimeShort.textContent = s + "s";
            }
        }
    }, 1000);

    // ---------------------------------------------------------------------------
    // UI helpers
    // ---------------------------------------------------------------------------
    function log(message, level) {
        level = level || "info";
        var timestamp = new Date().toLocaleTimeString();
        var entry = document.createElement("div");
        entry.className = "log-entry " + level;

        var tsSpan = document.createElement("span");
        tsSpan.className = "log-ts";
        tsSpan.textContent = "[" + timestamp + "]";
        entry.appendChild(tsSpan);

        entry.appendChild(document.createTextNode(" " + message));
        logArea.appendChild(entry);

        // Scroll to bottom if auto-scroll is on
        if (autoScroll) {
            logArea.scrollTop = logArea.scrollHeight;
        }

        // Limit log entries to prevent memory bloat
        while (logArea.children.length > 500) {
            logArea.removeChild(logArea.firstChild);
        }
    }

    function updateStatus(connected, count) {
        count = count || 0;
        if (connected) {
            statusDot.className = "status-dot connected";
            statusLabel.textContent = "Connected";
        } else {
            statusDot.className = "status-dot";
            statusLabel.textContent = "Disconnected";
        }
        if (clientCount) clientCount.textContent = "" + count;
    }

    function updateStatsUI() {
        if (statCommands) statCommands.textContent = "" + stats.commandsExecuted;
        if (statErrors) statErrors.textContent = "" + stats.errorsCount;
        if (statLastCmd) {
            statLastCmd.textContent = stats.lastCommand || "No commands yet";
            statLastCmd.title = stats.lastCommand || "";
        }
        if (statAvgTime) {
            if (stats.responseCount > 0) {
                var avg = Math.round(stats.totalResponseTime / stats.responseCount);
                statAvgTime.textContent = avg + "ms";
            } else {
                statAvgTime.textContent = "--";
            }
        }
    }

    // ---------------------------------------------------------------------------
    // Port configuration
    // ---------------------------------------------------------------------------
    function loadPort() {
        // Check for port override via environment or config file
        var configPath = path.join(__dirname, "..", "config.json");
        try {
            if (fs.existsSync(configPath)) {
                var config = JSON.parse(fs.readFileSync(configPath, "utf8"));
                if (config.port) {
                    serverPort = parseInt(config.port, 10);
                }
            }
        } catch (e) {
            // Ignore config errors, use default
        }

        // Environment variable takes highest priority
        if (typeof process !== "undefined" && process.env.MCP_CEP_PORT) {
            serverPort = parseInt(process.env.MCP_CEP_PORT, 10);
        }

        if (portDisplay) portDisplay.textContent = "" + serverPort;
        return serverPort;
    }

    // ---------------------------------------------------------------------------
    // Action-to-ExtendScript mapping
    // ---------------------------------------------------------------------------
    // Maps each incoming action name to the ExtendScript function call string.
    // Parameters are serialized as JSON and passed as a single string argument.
    var ACTION_MAP = {
        ping:               function ()        { return "ping()"; },
        getProjectState:    function ()        { return "getProjectState()"; },
        createSequence:     function (p)       { return "createSequence(" + escapeForEval(JSON.stringify(p)) + ")"; },
        getTimelineState:   function (p)       { return "getTimelineState(" + (p.sequenceIndex || 0) + ")"; },
        importMedia:        function (p)       { return "importMedia(" + escapeForEval(p.filePath) + "," + escapeForEval(p.binPath || "") + ")"; },
        placeClip:          function (p)       { return "placeClip(" + (p.projectItemIndex || 0) + "," + (p.trackIndex || 0) + "," + (p.startTime || 0) + ")"; },
        addTransition:      function (p)       { return "addTransition(" + (p.trackIndex || 0) + "," + (p.clipIndex || 0) + "," + escapeForEval(p.transitionName || "") + "," + (p.duration || 1) + ")"; },
        addText:            function (p)       { return "addText(" + escapeForEval(p.text || "") + "," + (p.trackIndex || 0) + "," + (p.startTime || 0) + "," + (p.duration || 5) + ")"; },
        setAudioLevel:      function (p)       { return "setAudioLevel(" + (p.trackIndex || 0) + "," + (p.clipIndex || 0) + "," + (p.levelDb || 0) + ")"; },
        exportSequence:     function (p)       { return "exportSequence(" + escapeForEval(p.outputPath || "") + "," + escapeForEval(p.presetPath || "") + ")"; },
        evalCommand:        function (p)       {
            var fn = p.function_name || "";
            var argsJson = p.args_json || "";

            // Before every evalCommand: if fn is missing, load premiere.jsx.
            // Always emit this guard — never skip it based on premiereJsxLoaded.
            // A prior bug set that flag after the first successful evalCommand even when
            // only core.jsx ran (e.g. getSequenceList), leaving premiere-only helpers
            // unloaded and causing "EvalScript error" on clips/MOGRT APIs.
            var loadScript =
                'if (typeof ' + fn + ' !== "function") { ' +
                '  try { $.evalFile("' + premiereJsxPath + '"); } catch(loadErr) {} ' +
                '} ';

            var callScript;
            var expanded = expandEvalCommandArgs(fn, argsJson);
            if (expanded !== null && expanded !== undefined) {
                callScript = expanded;
            } else if (argsJson && argsJson !== "{}" && argsJson !== "[]") {
                callScript = fn + "(" + escapeForEval(argsJson) + ")";
            } else {
                callScript = fn + "()";
            }

            return loadScript + callScript;
        },
    };

    /**
     * Escape a string value for safe inclusion inside an evalScript() call.
     * Wraps in single quotes and escapes internal quotes/backslashes.
     */
    function escapeForEval(str) {
        if (str === undefined || str === null) { str = ""; }
        str = String(str);
        return "'" + str.replace(/\\/g, "\\\\").replace(/'/g, "\\'") + "'";
    }

    /** Go sends evalCommand(fn, "{\"a\":...}") — expand into positional ExtendScript calls. */
    function expandEvalCommandArgs(fn, argsJson) {
        if (!argsJson || typeof argsJson !== "string") return null;
        if (argsJson.charAt(0) !== "{") return null;
        var po;
        try { po = JSON.parse(argsJson); } catch (e0) { return null; }

        var ti;
        var ci;
        var pi;
        var txt;
        var pn;
        var val;

        if (fn === "getClipsOnTrack") {
            var tt = po.trackType || po.track_type || "video";
            ti = po.trackIndex !== undefined ? po.trackIndex : po.track_index;
            if (ti === undefined || ti === null) ti = 0;
            return fn + "(" + escapeForEval(String(tt)) + "," + Number(ti) + ")";
        }
        if (fn === "getMOGRTProperties") {
            ti = po.trackIndex !== undefined ? po.trackIndex : po.track_index;
            ci = po.clipIndex !== undefined ? po.clipIndex : po.clip_index;
            if (ti === undefined || ti === null) ti = 0;
            if (ci === undefined || ci === null) ci = 0;
            return fn + "(" + Number(ti) + "," + Number(ci) + ")";
        }
        if (fn === "setMOGRTText") {
            ti = po.trackIndex !== undefined ? po.trackIndex : po.track_index;
            ci = po.clipIndex !== undefined ? po.clipIndex : po.clip_index;
            pi = po.propertyIndex !== undefined ? po.propertyIndex : po.property_index;
            txt = po.text !== undefined && po.text !== null ? String(po.text) : "";
            if (ti === undefined || ti === null) ti = 0;
            if (ci === undefined || ci === null) ci = 0;
            if (pi === undefined || pi === null) pi = 0;
            return fn + "(" + Number(ti) + "," + Number(ci) + "," + Number(pi) + "," + escapeForEval(txt) + ")";
        }
        if (fn === "setMOGRTProperty") {
            ti = po.trackIndex !== undefined ? po.trackIndex : po.track_index;
            ci = po.clipIndex !== undefined ? po.clipIndex : po.clip_index;
            pn = po.propertyName || po.property_name || "";
            val = po.value !== undefined && po.value !== null ? String(po.value) : "";
            if (ti === undefined || ti === null) ti = 0;
            if (ci === undefined || ci === null) ci = 0;
            return fn + "(" + Number(ti) + "," + Number(ci) + "," + escapeForEval(pn) + "," + escapeForEval(val) + ")";
        }
        return null;
    }

    // ---------------------------------------------------------------------------
    // Command execution
    // ---------------------------------------------------------------------------
    function executeCommand(action, params, requestId, ws) {
        log("Exec: " + action + " [" + requestId.substring(0, 8) + "]");

        // Track stats
        stats.commandsExecuted++;
        stats.lastCommand = action;
        commandTimers[requestId] = Date.now();
        updateStatsUI();

        var builder = ACTION_MAP[action];
        if (!builder) {
            stats.errorsCount++;
            updateStatsUI();
            sendResponse(ws, requestId, false, null, "Unknown action: " + action);
            return;
        }

        var script;
        try {
            script = builder(params || {});
        } catch (buildErr) {
            stats.errorsCount++;
            updateStatsUI();
            sendResponse(ws, requestId, false, null, "Failed to build script: " + buildErr.message);
            return;
        }

        csInterface.evalScript(script, function (rawResult) {
            // Measure response time
            if (commandTimers[requestId]) {
                var elapsed = Date.now() - commandTimers[requestId];
                stats.totalResponseTime += elapsed;
                stats.responseCount++;
                delete commandTimers[requestId];
            }

            // ExtendScript returns "EvalScript error." on failure
            if (rawResult === "EvalScript error.") {
                log("ExtendScript error for " + action, "error");
                stats.errorsCount++;
                updateStatsUI();
                sendResponse(ws, requestId, false, null, "ExtendScript evaluation error for action: " + action);
                return;
            }

            // Try to parse the JSON result from ExtendScript
            var result;
            try {
                result = JSON.parse(rawResult);
            } catch (e) {
                // If it's not JSON, return it as a plain string value
                result = rawResult;
            }

            log("Done: " + action + " [" + requestId.substring(0, 8) + "]", "success");
            updateStatsUI();
            sendResponse(ws, requestId, true, result, null);
        });
    }

    function sendResponse(ws, requestId, success, result, error) {
        if (ws.readyState !== 1) { // WebSocket.OPEN
            log("Cannot send response - WebSocket not open", "error");
            return;
        }

        var response = {
            requestId: requestId,
            success: success,
        };

        if (success) {
            response.result = result;
        } else {
            response.error = error || "Unknown error";
        }

        try {
            ws.send(JSON.stringify(response));
        } catch (sendErr) {
            log("Send error: " + sendErr.message, "error");
        }
    }

    // ---------------------------------------------------------------------------
    // WebSocket server
    // ---------------------------------------------------------------------------
    function startServer() {
        var port = loadPort();

        // Show connecting state
        statusDot.className = "status-dot connecting";
        statusLabel.textContent = "Starting...";

        try {
            wss = new WebSocketServer({ port: port });
        } catch (err) {
            log("Failed to start WebSocket server on port " + port + ": " + err.message, "error");
            statusDot.className = "status-dot";
            statusLabel.textContent = "Failed";
            // Retry after delay
            setTimeout(startServer, RECONNECT_DELAY_MS);
            return;
        }

        log("WebSocket server listening on port " + port, "success");
        statusDot.className = "status-dot";
        statusLabel.textContent = "Waiting for clients...";

        wss.on("connection", function (ws, req) {
            var clientAddr = req.socket.remoteAddress || "unknown";
            activeConnections.add(ws);
            updateStatus(true, activeConnections.size);
            log("Client connected: " + clientAddr, "success");

            ws.isAlive = true;

            ws.on("pong", function () {
                ws.isAlive = true;
            });

            ws.on("message", function (data) {
                var message;
                try {
                    message = JSON.parse(data.toString());
                } catch (parseErr) {
                    log("Invalid JSON received: " + parseErr.message, "error");
                    stats.errorsCount++;
                    updateStatsUI();
                    sendResponse(ws, null, false, null, "Invalid JSON message");
                    return;
                }

                var action = message.action;
                var params = message.params || {};
                var requestId = message.requestId || "no-id";

                if (!action) {
                    stats.errorsCount++;
                    updateStatsUI();
                    sendResponse(ws, requestId, false, null, "Missing 'action' field");
                    return;
                }

                executeCommand(action, params, requestId, ws);
            });

            ws.on("close", function (code, reason) {
                activeConnections.delete(ws);
                updateStatus(activeConnections.size > 0, activeConnections.size);
                log("Client disconnected: " + clientAddr + " (code=" + code + ")", "info");
            });

            ws.on("error", function (err) {
                log("WebSocket client error: " + err.message, "error");
                activeConnections.delete(ws);
                updateStatus(activeConnections.size > 0, activeConnections.size);
            });
        });

        wss.on("error", function (err) {
            log("WebSocket server error: " + err.message, "error");
            if (err.code === "EADDRINUSE") {
                log("Port " + port + " in use. Retrying in " + (RECONNECT_DELAY_MS / 1000) + "s...", "warning");
                wss.close();
                setTimeout(startServer, RECONNECT_DELAY_MS);
            }
        });

        // Start heartbeat to detect dead connections
        startHeartbeat();
    }

    // ---------------------------------------------------------------------------
    // Heartbeat - detect and clean up dead connections
    // ---------------------------------------------------------------------------
    function startHeartbeat() {
        if (heartbeatTimer) {
            clearInterval(heartbeatTimer);
        }
        heartbeatTimer = setInterval(function () {
            if (!wss) return;
            wss.clients.forEach(function (ws) {
                if (ws.isAlive === false) {
                    log("Terminating unresponsive client", "error");
                    activeConnections.delete(ws);
                    ws.terminate();
                    updateStatus(activeConnections.size > 0, activeConnections.size);
                    return;
                }
                ws.isAlive = false;
                ws.ping();
            });
        }, HEARTBEAT_INTERVAL_MS);
    }

    // ---------------------------------------------------------------------------
    // Shutdown
    // ---------------------------------------------------------------------------
    function shutdown() {
        log("Shutting down...", "info");
        if (uptimeTimer) {
            clearInterval(uptimeTimer);
            uptimeTimer = null;
        }
        if (heartbeatTimer) {
            clearInterval(heartbeatTimer);
            heartbeatTimer = null;
        }
        if (wss) {
            activeConnections.forEach(function (ws) {
                try { ws.close(1001, "Panel closing"); } catch (e) { /* ignore */ }
            });
            activeConnections.clear();
            wss.close();
            wss = null;
        }
    }

    // ---------------------------------------------------------------------------
    // Load ExtendScript host functions
    // ---------------------------------------------------------------------------
    function loadHostScript() {
        var corePath = resolveHostScriptPath("core.jsx");
        var coreEval = [
            "(function(){",
            "try {",
            "  $.evalFile(\"" + corePath + "\");",
            "  return \"OK\";",
            "} catch (e) {",
            "  return \"ERROR: \" + e.name + \": \" + e.message + \" @ line \" + e.line;",
            "}",
            "}())",
        ].join("");
        var premiereEval = [
            "(function(){",
            "try {",
            "  $.evalFile(\"" + premiereJsxPath + "\");",
            "  return \"OK\";",
            "} catch (e) {",
            "  return \"ERROR: \" + e.name + \": \" + e.message + \" @ line \" + e.line;",
            "}",
            "}())",
        ].join("");
        log("Loading core ExtendScript: " + corePath);
        csInterface.evalScript(coreEval, function (result) {
            if (result !== "OK") {
                log("Failed to load core.jsx -- " + result, "error");
            } else {
                log("Core ExtendScript loaded", "success");
                // Eagerly attempt to load the full premiere.jsx in the background.
                // If it fails (too large for $.evalFile), functions will be
                // lazy-loaded on the first evalCommand call instead.
                log("Loading extended functions from premiere.jsx...");
                csInterface.evalScript(premiereEval, function (result2) {
                    if (result2 !== "OK") {
                        log("Extended functions will be loaded on first use (lazy): " + result2, "info");
                    } else {
                        premiereJsxLoaded = true;
                        log("All ExtendScript functions loaded", "success");
                    }
                });
            }
        });
    }

    // Load the host script at startup
    loadHostScript();

    // Listen for panel close event
    csInterface.addEventListener("com.adobe.csxs.events.WindowVisibilityChanged", function (event) {
        if (event.data === "false") {
            shutdown();
        }
    });

    // Also handle process exit in Node.js context
    if (typeof process !== "undefined") {
        process.on("exit", shutdown);
    }

    // ---------------------------------------------------------------------------
    // Broadcast utility (for future use - push events to all connected clients)
    // ---------------------------------------------------------------------------
    function broadcast(eventType, data) {
        var message = JSON.stringify({
            type: "event",
            event: eventType,
            data: data,
            timestamp: Date.now(),
        });
        activeConnections.forEach(function (ws) {
            if (ws.readyState === 1) {
                try { ws.send(message); } catch (e) { /* ignore */ }
            }
        });
    }

    // Expose broadcast for potential use from ExtendScript callbacks
    window.mcpBroadcast = broadcast;

    // ---------------------------------------------------------------------------
    // Initialization
    // ---------------------------------------------------------------------------
    function init() {
        log("PremierPro MCP Bridge v" + VERSION + " initializing...", "info");
        log("CEP Engine: " + JSON.stringify(csInterface.getCurrentApiVersion()), "info");

        // Verify we're running in Premiere Pro
        var hostEnv = csInterface.getHostEnvironment();
        if (hostEnv && hostEnv.appName) {
            log("Host application: " + hostEnv.appName, "info");
        }

        startServer();
    }

    // Start when DOM is ready (it should already be, but be safe)
    if (document.readyState === "complete" || document.readyState === "interactive") {
        init();
    } else {
        document.addEventListener("DOMContentLoaded", init);
    }

})();
