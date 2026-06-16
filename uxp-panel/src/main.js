/**
 * PremierPro MCP UXP Bridge
 *
 * Connects to the ts-bridge WebSocket server and serves transcript operations
 * via Premiere Pro's UXP Transcript API (Text panel data, not caption tracks).
 */

(function () {
  "use strict";

  const ppro = require("premierepro");
  const uxp = require("uxp");
  const storage = uxp.storage;
  const fs = storage.localFileSystem;

  const VERSION = "1.0.0";
  const DEFAULT_WS_URL = "ws://127.0.0.1:9802";
  const RECONNECT_MS = 3000;

  let ws = null;
  let reconnectTimer = null;
  let wsUrl = DEFAULT_WS_URL;

  const statusEl = document.getElementById("status");
  const logEl = document.getElementById("log");
  const wsUrlInput = document.getElementById("ws-url");
  const btnReconnect = document.getElementById("btn-reconnect");
  const btnClear = document.getElementById("btn-clear");

  function log(message, level) {
    const line = "[" + new Date().toISOString() + "] " + message;
    if (logEl) {
      logEl.textContent += line + "\n";
      logEl.scrollTop = logEl.scrollHeight;
    }
    console.log("[uxp-bridge][" + (level || "info") + "] " + message);
  }

  function setStatus(text, cls) {
    if (statusEl) {
      statusEl.textContent = text;
      statusEl.className = cls || "warn";
    }
  }

  function loadWsUrl() {
    try {
      const saved = localStorage.getItem("premierpro_mcp_uxp_ws_url");
      if (saved) {
        wsUrl = saved;
        if (wsUrlInput) wsUrlInput.value = saved;
      }
    } catch (e) { /* ignore */ }
  }

  function saveWsUrl(url) {
    wsUrl = url;
    try { localStorage.setItem("premierpro_mcp_uxp_ws_url", url); } catch (e) { /* ignore */ }
  }

  function pad2(n) { return n < 10 ? "0" + n : String(n); }

  function secondsToTimecode(seconds, fps) {
    const rate = fps > 0 ? fps : 30;
    const totalFrames = Math.max(0, Math.round(seconds * rate));
    const h = Math.floor(totalFrames / (rate * 3600));
    const m = Math.floor((totalFrames % (rate * 3600)) / (rate * 60));
    const s = Math.floor((totalFrames % (rate * 60)) / rate);
    const f = totalFrames % rate;
    return pad2(h) + ":" + pad2(m) + ":" + pad2(s) + ":" + pad2(f);
  }

  function toFileUrl(filePath) {
    const norm = String(filePath || "").replace(/\\/g, "/");
    if (/^[A-Za-z]:\//.test(norm)) return "file:/" + norm;
    if (norm.startsWith("/")) return "file://" + norm;
    return "file://" + norm;
  }

  async function writeTextFile(outputPath, content) {
    const url = toFileUrl(outputPath);
    let entry;
    try {
      entry = await fs.getEntryWithUrl(url);
    } catch (e) {
      const lastSlash = Math.max(outputPath.lastIndexOf("/"), outputPath.lastIndexOf("\\"));
      const dirPath = lastSlash > 0 ? outputPath.substring(0, lastSlash) : outputPath;
      const fileName = lastSlash > 0 ? outputPath.substring(lastSlash + 1) : outputPath;
      const dirUrl = toFileUrl(dirPath);
      const dir = await fs.getEntryWithUrl(dirUrl);
      entry = await dir.createFile(fileName, { overwrite: true });
    }
    await entry.write(content, { format: storage.formats.utf8 });
  }

  function segmentText(words) {
    if (!words || !words.length) return "";
    let out = "";
    for (let i = 0; i < words.length; i++) {
      const w = words[i];
      const t = w && w.text ? String(w.text) : "";
      if (!t) continue;
      if (w.type === "punctuation" && out.length > 0) {
        out += t;
      } else if (out.length === 0) {
        out = t;
      } else {
        out += " " + t;
      }
    }
    return out.trim();
  }

  function speakerNameForSegment(segment, speakers, fallback) {
    if (!segment || !segment.speaker || !speakers || !speakers.length) return fallback;
    for (let i = 0; i < speakers.length; i++) {
      if (speakers[i].id === segment.speaker) {
        return speakers[i].name || fallback;
      }
    }
    return fallback;
  }

  async function getSequenceFps(sequence) {
    try {
      const settings = await sequence.getSettings();
      const frameRate = settings.getVideoFrameRate();
      const value = frameRate && frameRate.value ? Number(frameRate.value) : 0;
      return value > 0 ? value : 30;
    } catch (e) {
      return 30;
    }
  }

  async function tickSeconds(tickTime) {
    if (!tickTime) return 0;
    if (typeof tickTime.seconds === "number") return tickTime.seconds;
    if (typeof tickTime.getSeconds === "function") {
      const s = await tickTime.getSeconds();
      return typeof s === "number" ? s : 0;
    }
    return 0;
  }

  async function collectFromTrackItem(trackItem, segments, stats, fps) {
    if (!trackItem) return;

    let projectItem;
    try {
      projectItem = await trackItem.getProjectItem();
    } catch (e) {
      stats.errors.push("getProjectItem failed: " + e.message);
      return;
    }
    if (!projectItem) return;

    const clipProjectItem = ppro.ClipProjectItem.cast(projectItem);
    if (!clipProjectItem) return;

    let hasTranscript = false;
    try {
      hasTranscript = ppro.Transcript.hasTranscript(clipProjectItem);
    } catch (e) {
      stats.errors.push("hasTranscript failed for " + (projectItem.name || "clip") + ": " + e.message);
      return;
    }

    if (!hasTranscript) {
      stats.clipsWithoutTranscript++;
      return;
    }

    let jsonStr;
    try {
      jsonStr = await ppro.Transcript.exportToJSON(clipProjectItem);
    } catch (e) {
      stats.errors.push("exportToJSON failed for " + (projectItem.name || "clip") + ": " + e.message);
      return;
    }

    let transcript;
    try {
      transcript = typeof jsonStr === "string" ? JSON.parse(jsonStr) : jsonStr;
    } catch (e) {
      stats.errors.push("transcript JSON parse failed: " + e.message);
      return;
    }

    const seqStart = await tickSeconds(await trackItem.getStartTime());
    const inPoint = await tickSeconds(await trackItem.getInPoint());

    const speakers = transcript.speakers || [];
    const segs = transcript.segments || [];
    stats.clipsWithTranscript++;

    for (let i = 0; i < segs.length; i++) {
      const seg = segs[i];
      const relStart = typeof seg.start === "number" ? seg.start : 0;
      const relDuration = typeof seg.duration === "number" ? seg.duration : 0;
      const seqSegStart = seqStart + (relStart - inPoint);
      const seqSegEnd = seqSegStart + relDuration;
      if (seqSegEnd <= seqSegStart) continue;

      segments.push({
        startSeconds: seqSegStart,
        endSeconds: seqSegEnd,
        speaker: speakerNameForSegment(seg, speakers, stats.defaultSpeaker),
        text: segmentText(seg.words),
        sourceClip: projectItem.name || "",
      });
    }
  }

  async function walkTrackItems(sequence, trackType, segments, stats) {
    let trackCount = 0;
    try {
      if (trackType === "video") {
        trackCount = await sequence.getVideoTrackCount();
      } else {
        trackCount = await sequence.getAudioTrackCount();
      }
    } catch (e) {
      stats.errors.push("track count failed: " + e.message);
      return;
    }

    const clipType = ppro.Constants.TrackItemType.CLIP;

    for (let t = 0; t < trackCount; t++) {
      let track;
      try {
        track = trackType === "video"
          ? await sequence.getVideoTrack(t)
          : await sequence.getAudioTrack(t);
      } catch (e) {
        stats.errors.push("getTrack(" + t + ") failed: " + e.message);
        continue;
      }
      if (!track || typeof track.getTrackItems !== "function") continue;

      let items = [];
      try {
        items = track.getTrackItems(clipType, false) || [];
      } catch (e) {
        stats.errors.push("getTrackItems failed on track " + t + ": " + e.message);
        continue;
      }

      for (let i = 0; i < items.length; i++) {
        await collectFromTrackItem(items[i], segments, stats, stats.fps);
      }
    }
  }

  function formatPrtranscript(segments, fps) {
    const lines = [];
    for (let i = 0; i < segments.length; i++) {
      const seg = segments[i];
      if (!seg.text) continue;
      lines.push(
        secondsToTimecode(seg.startSeconds, fps) + " - " + secondsToTimecode(seg.endSeconds, fps)
      );
      lines.push(seg.speaker || "Unknown");
      lines.push(seg.text);
      lines.push("");
    }
    return lines.join("\n").trim() + (lines.length ? "\n" : "");
  }

  function formatPlainText(segments) {
    const lines = [];
    for (let i = 0; i < segments.length; i++) {
      const seg = segments[i];
      if (!seg.text) continue;
      lines.push("[" + seg.startSeconds.toFixed(3) + "] " + seg.speaker + ": " + seg.text);
    }
    return lines.join("\n").trim() + (lines.length ? "\n" : "");
  }

  async function exportSequenceTranscript(params) {
    const outputPath = params && params.outputPath ? String(params.outputPath) : "";
    const format = (params && params.format ? String(params.format) : "prtranscript").toLowerCase();
    const defaultSpeaker = params && params.speakerLabel ? String(params.speakerLabel) : "Unknown";
    const includeAudioTracks = !!(params && (params.includeAudioTracks || params.include_audio_tracks));

    if (!outputPath) {
      return { success: false, error: "outputPath is required" };
    }

    const project = await ppro.Project.getActiveProject();
    if (!project) {
      return { success: false, error: "No project is open" };
    }

    const sequence = await project.getActiveSequence();
    if (!sequence) {
      return { success: false, error: "No active sequence" };
    }

    const fps = await getSequenceFps(sequence);
    const segments = [];
    const stats = {
      fps: fps,
      defaultSpeaker: defaultSpeaker,
      clipsWithTranscript: 0,
      clipsWithoutTranscript: 0,
      errors: [],
      sequenceName: sequence.name || "",
    };

    await walkTrackItems(sequence, "video", segments, stats);
    if (includeAudioTracks) {
      await walkTrackItems(sequence, "audio", segments, stats);
    }

    segments.sort(function (a, b) { return a.startSeconds - b.startSeconds; });

    if (segments.length === 0) {
      return {
        success: false,
        error: "No transcript segments found on the active sequence. Transcribe source clips in Text > Transcript, then retry.",
        sequenceName: stats.sequenceName,
        clipsWithTranscript: stats.clipsWithTranscript,
        clipsWithoutTranscript: stats.clipsWithoutTranscript,
        errors: stats.errors,
        uxpTranscriptApi: true,
      };
    }

    let content;
    if (format === "json") {
      content = JSON.stringify({
        sequenceName: stats.sequenceName,
        fps: fps,
        segmentCount: segments.length,
        segments: segments,
        stats: {
          clipsWithTranscript: stats.clipsWithTranscript,
          clipsWithoutTranscript: stats.clipsWithoutTranscript,
          errors: stats.errors,
        },
      }, null, 2);
    } else if (format === "text" || format === "txt") {
      content = formatPlainText(segments);
    } else {
      content = formatPrtranscript(segments, fps);
    }

    await writeTextFile(outputPath, content);

    return {
      success: true,
      outputPath: outputPath,
      format: format,
      segmentCount: segments.length,
      sequenceName: stats.sequenceName,
      fps: fps,
      clipsWithTranscript: stats.clipsWithTranscript,
      clipsWithoutTranscript: stats.clipsWithoutTranscript,
      errors: stats.errors,
      uxpTranscriptApi: true,
    };
  }

  async function handleCommand(message) {
    const action = message.action;
    const params = message.params || {};
    const requestId = message.requestId;

    if (action === "ping") {
      return {
        requestId: requestId,
        success: true,
        result: {
          role: "uxp-panel",
          version: VERSION,
          transcriptApi: true,
        },
      };
    }

    if (action === "exportSequenceTranscript") {
      try {
        const result = await exportSequenceTranscript(params);
        return {
          requestId: requestId,
          success: !!result.success,
          result: result,
          error: result.success ? undefined : result.error,
        };
      } catch (e) {
        return {
          requestId: requestId,
          success: false,
          error: e && e.message ? e.message : String(e),
        };
      }
    }

    return {
      requestId: requestId,
      success: false,
      error: "Unknown UXP action: " + action,
    };
  }

  function scheduleReconnect() {
    if (reconnectTimer) return;
    reconnectTimer = setTimeout(function () {
      reconnectTimer = null;
      connect();
    }, RECONNECT_MS);
  }

  function connect() {
    if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
      return;
    }

    if (wsUrlInput && wsUrlInput.value) {
      saveWsUrl(wsUrlInput.value.trim());
    }

    setStatus("Connecting to " + wsUrl + "...", "warn");
    log("Connecting to " + wsUrl);

    try {
      ws = new WebSocket(wsUrl);
    } catch (e) {
      setStatus("WebSocket error: " + e.message, "err");
      scheduleReconnect();
      return;
    }

    ws.onopen = function () {
      setStatus("Connected (v" + VERSION + ")", "ok");
      log("Connected");
      try {
        ws.send(JSON.stringify({ type: "hello", role: "uxp-panel", version: VERSION }));
      } catch (e) { /* ignore */ }
    };

    ws.onclose = function () {
      setStatus("Disconnected — retrying...", "warn");
      log("Disconnected");
      scheduleReconnect();
    };

    ws.onerror = function () {
      setStatus("WebSocket error — retrying...", "err");
    };

    ws.onmessage = async function (event) {
      let message;
      try {
        message = JSON.parse(event.data);
      } catch (e) {
        log("Invalid JSON from MCP: " + e.message, "error");
        return;
      }

      const response = await handleCommand(message);
      try {
        ws.send(JSON.stringify(response));
      } catch (e) {
        log("Failed to send response: " + e.message, "error");
      }
    };
  }

  if (btnReconnect) {
    btnReconnect.addEventListener("click", function () {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      if (ws) {
        try { ws.close(); } catch (e) { /* ignore */ }
        ws = null;
      }
      connect();
    });
  }

  if (btnClear && logEl) {
    btnClear.addEventListener("click", function () {
      logEl.textContent = "";
    });
  }

  loadWsUrl();
  connect();
})();
