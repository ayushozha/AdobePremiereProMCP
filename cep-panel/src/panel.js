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

    // ---------------------------------------------------------------------------
    // State
    // ---------------------------------------------------------------------------
    var csInterface = new CSInterface();
    var wss = null;
    var activeConnections = new Set();
    var heartbeatTimer = null;
    var serverPort = DEFAULT_PORT;

    // ---------------------------------------------------------------------------
    // UI helpers
    // ---------------------------------------------------------------------------
    var logArea = document.getElementById("log-area");
    var statusIndicator = document.getElementById("status-indicator");
    var statusText = document.getElementById("status-text");
    var portDisplay = document.getElementById("port-display");

    function log(message, level) {
        level = level || "info";
        var timestamp = new Date().toLocaleTimeString();
        var entry = document.createElement("div");
        entry.className = "log-entry " + level;
        entry.textContent = "[" + timestamp + "] " + message;
        logArea.appendChild(entry);
        // Keep log area scrolled to bottom
        logArea.scrollTop = logArea.scrollHeight;
        // Limit log entries to prevent memory bloat
        while (logArea.children.length > 500) {
            logArea.removeChild(logArea.firstChild);
        }
    }

    function updateStatus(connected, clientCount) {
        if (connected) {
            statusIndicator.className = "connected";
            statusText.textContent = "Connected (" + clientCount + " client" + (clientCount !== 1 ? "s" : "") + ")";
        } else {
            statusIndicator.className = "";
            statusText.textContent = "Waiting for connection...";
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

        portDisplay.textContent = "Port: " + serverPort;
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

    // ---------------------------------------------------------------------------
    // Command execution
    // ---------------------------------------------------------------------------
    function executeCommand(action, params, requestId, ws) {
        log("Exec: " + action + " [" + requestId + "]");

        var builder = ACTION_MAP[action];
        if (!builder) {
            sendResponse(ws, requestId, false, null, "Unknown action: " + action);
            return;
        }

        var script;
        try {
            script = builder(params || {});
        } catch (buildErr) {
            sendResponse(ws, requestId, false, null, "Failed to build script: " + buildErr.message);
            return;
        }

        csInterface.evalScript(script, function (rawResult) {
            // ExtendScript returns "EvalScript error." on failure
            if (rawResult === "EvalScript error.") {
                log("ExtendScript error for " + action, "error");
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

            log("Done: " + action + " [" + requestId + "]", "success");
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

        try {
            wss = new WebSocketServer({ port: port });
        } catch (err) {
            log("Failed to start WebSocket server on port " + port + ": " + err.message, "error");
            // Retry after delay
            setTimeout(startServer, RECONNECT_DELAY_MS);
            return;
        }

        log("WebSocket server listening on port " + port, "success");

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
                    sendResponse(ws, null, false, null, "Invalid JSON message");
                    return;
                }

                var action = message.action;
                var params = message.params || {};
                var requestId = message.requestId || "no-id";

                if (!action) {
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
                log("Port " + port + " in use. Retrying in " + (RECONNECT_DELAY_MS / 1000) + "s...", "error");
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
        log("PremierPro MCP Bridge v1.0.0 initializing...", "info");
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
