# PremierPro MCP Server — Architecture

## System Overview

The PremierPro MCP Server is a multi-language system that enables AI-driven end-to-end video editing in Adobe Premiere Pro. It uses four languages — Go, Rust, Python, and TypeScript — each chosen for their strengths in a specific layer of the stack.

## Architecture Diagram

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1000 920" font-family="'Segoe UI', system-ui, -apple-system, sans-serif">
  <defs>
    <filter id="shadow" x="-4%" y="-4%" width="108%" height="108%">
      <feDropShadow dx="0" dy="2" stdDeviation="3" flood-opacity="0.1"/>
    </filter>
    <linearGradient id="goBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#00ADD8"/>
      <stop offset="100%" stop-color="#0097B9"/>
    </linearGradient>
    <linearGradient id="rustBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#DEA584"/>
      <stop offset="100%" stop-color="#C67849"/>
    </linearGradient>
    <linearGradient id="pyBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#3776AB"/>
      <stop offset="100%" stop-color="#2B5E8C"/>
    </linearGradient>
    <linearGradient id="tsBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#3178C6"/>
      <stop offset="100%" stop-color="#265EA0"/>
    </linearGradient>
    <linearGradient id="cliBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#6C5CE7"/>
      <stop offset="100%" stop-color="#5A4BD1"/>
    </linearGradient>
    <linearGradient id="prBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#9999FF"/>
      <stop offset="100%" stop-color="#7A7AE6"/>
    </linearGradient>
    <marker id="arrowGray" viewBox="0 0 10 7" refX="10" refY="3.5" markerWidth="10" markerHeight="7" orient="auto-start-reverse">
      <polygon points="0 0, 10 3.5, 0 7" fill="#8B95A5"/>
    </marker>
    <marker id="arrowBlue" viewBox="0 0 10 7" refX="10" refY="3.5" markerWidth="10" markerHeight="7" orient="auto-start-reverse">
      <polygon points="0 0, 10 3.5, 0 7" fill="#3178C6"/>
    </marker>
  </defs>

  <!-- Background -->
  <rect width="1000" height="920" rx="16" fill="#0F1117"/>

  <!-- Title -->
  <text x="500" y="42" text-anchor="middle" fill="#E2E8F0" font-size="20" font-weight="700">PremierPro MCP Server — System Architecture</text>

  <!-- ─── CLI / MCP Client ─── -->
  <rect x="300" y="65" width="400" height="60" rx="12" fill="url(#cliBg)" filter="url(#shadow)"/>
  <text x="500" y="93" text-anchor="middle" fill="#fff" font-size="15" font-weight="700">CLI / MCP Client</text>
  <text x="500" y="113" text-anchor="middle" fill="#D5CDFF" font-size="11">Claude Code &middot; MCP Plugin &middot; Custom CLI</text>

  <!-- Arrow: CLI → Go -->
  <line x1="500" y1="125" x2="500" y2="155" stroke="#8B95A5" stroke-width="2" marker-end="url(#arrowGray)"/>
  <text x="512" y="146" fill="#8B95A5" font-size="10">stdio / JSON-RPC</text>

  <!-- ─── Go Orchestrator ─── -->
  <rect x="150" y="160" width="700" height="130" rx="14" fill="url(#goBg)" filter="url(#shadow)"/>
  <text x="500" y="190" text-anchor="middle" fill="#fff" font-size="16" font-weight="700">Go — MCP Server &amp; Orchestrator</text>
  <!-- Go feature boxes -->
  <rect x="175" y="205" width="155" height="36" rx="6" fill="rgba(255,255,255,0.15)"/>
  <text x="252" y="228" text-anchor="middle" fill="#fff" font-size="11">MCP Protocol Handler</text>
  <rect x="345" y="205" width="155" height="36" rx="6" fill="rgba(255,255,255,0.15)"/>
  <text x="422" y="228" text-anchor="middle" fill="#fff" font-size="11">Task Orchestration</text>
  <rect x="515" y="205" width="155" height="36" rx="6" fill="rgba(255,255,255,0.15)"/>
  <text x="592" y="228" text-anchor="middle" fill="#fff" font-size="11">Service Mesh</text>
  <rect x="685" y="205" width="140" height="36" rx="6" fill="rgba(255,255,255,0.15)"/>
  <text x="755" y="228" text-anchor="middle" fill="#fff" font-size="11">Health &amp; Recovery</text>

  <rect x="175" y="250" width="155" height="28" rx="6" fill="rgba(255,255,255,0.10)"/>
  <text x="252" y="269" text-anchor="middle" fill="#B8E8F5" font-size="10">Concurrency (goroutines)</text>
  <rect x="345" y="250" width="155" height="28" rx="6" fill="rgba(255,255,255,0.10)"/>
  <text x="422" y="269" text-anchor="middle" fill="#B8E8F5" font-size="10">Retry &amp; Circuit Breaker</text>
  <rect x="515" y="250" width="155" height="28" rx="6" fill="rgba(255,255,255,0.10)"/>
  <text x="592" y="269" text-anchor="middle" fill="#B8E8F5" font-size="10">Graceful Shutdown</text>
  <rect x="685" y="250" width="140" height="28" rx="6" fill="rgba(255,255,255,0.10)"/>
  <text x="755" y="269" text-anchor="middle" fill="#B8E8F5" font-size="10">Logging &amp; Metrics</text>

  <!-- Arrows: Go → services -->
  <line x1="280" y1="290" x2="200" y2="360" stroke="#8B95A5" stroke-width="2" marker-end="url(#arrowGray)"/>
  <text x="210" y="330" fill="#8B95A5" font-size="10">gRPC</text>

  <line x1="500" y1="290" x2="500" y2="360" stroke="#8B95A5" stroke-width="2" marker-end="url(#arrowGray)"/>
  <text x="512" y="330" fill="#8B95A5" font-size="10">gRPC</text>

  <line x1="720" y1="290" x2="800" y2="360" stroke="#8B95A5" stroke-width="2" marker-end="url(#arrowGray)"/>
  <text x="775" y="330" fill="#8B95A5" font-size="10">gRPC / HTTP</text>

  <!-- ─── Rust Engine ─── -->
  <rect x="50" y="365" width="280" height="230" rx="14" fill="url(#rustBg)" filter="url(#shadow)"/>
  <text x="190" y="395" text-anchor="middle" fill="#fff" font-size="15" font-weight="700">Rust — Media Engine</text>
  <rect x="70" y="410" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="190" y="431" text-anchor="middle" fill="#fff" font-size="11">Media Probe &amp; Metadata Extract</text>
  <rect x="70" y="450" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="190" y="471" text-anchor="middle" fill="#fff" font-size="11">Asset Indexing &amp; Fingerprinting</text>
  <rect x="70" y="490" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="190" y="511" text-anchor="middle" fill="#fff" font-size="11">Waveform &amp; Silence Detection</text>
  <rect x="70" y="530" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="190" y="551" text-anchor="middle" fill="#fff" font-size="11">Thumbnail Generation</text>
  <rect x="70" y="570" width="240" height="14" rx="4" fill="rgba(0,0,0,0.15)"/>
  <text x="190" y="581" text-anchor="middle" fill="#F5E0D0" font-size="9">FFmpeg bindings &middot; Zero-copy I/O</text>

  <!-- ─── Python Brain ─── -->
  <rect x="360" y="365" width="280" height="230" rx="14" fill="url(#pyBg)" filter="url(#shadow)"/>
  <text x="500" y="395" text-anchor="middle" fill="#fff" font-size="15" font-weight="700">Python — Intelligence</text>
  <rect x="380" y="410" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="500" y="431" text-anchor="middle" fill="#fff" font-size="11">Script Parsing &amp; NLP</text>
  <rect x="380" y="450" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="500" y="471" text-anchor="middle" fill="#fff" font-size="11">Edit Decision List (EDL) Gen</text>
  <rect x="380" y="490" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="500" y="511" text-anchor="middle" fill="#fff" font-size="11">Shot ↔ Asset Matching (AI)</text>
  <rect x="380" y="530" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="500" y="551" text-anchor="middle" fill="#fff" font-size="11">Pacing &amp; Timing Analysis</text>
  <rect x="380" y="570" width="240" height="14" rx="4" fill="rgba(0,0,0,0.15)"/>
  <text x="500" y="581" text-anchor="middle" fill="#A8C8E8" font-size="9">LLM &middot; Embeddings &middot; Scene Detection</text>

  <!-- ─── TypeScript Bridge ─── -->
  <rect x="670" y="365" width="280" height="230" rx="14" fill="url(#tsBg)" filter="url(#shadow)"/>
  <text x="810" y="395" text-anchor="middle" fill="#fff" font-size="15" font-weight="700">TypeScript — PP Bridge</text>
  <rect x="690" y="410" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="810" y="431" text-anchor="middle" fill="#fff" font-size="11">ExtendScript API Layer</text>
  <rect x="690" y="450" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="810" y="471" text-anchor="middle" fill="#fff" font-size="11">CEP Panel (Primary Bridge)</text>
  <rect x="690" y="490" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="810" y="511" text-anchor="middle" fill="#fff" font-size="11">Timeline &amp; Clip Operations</text>
  <rect x="690" y="530" width="240" height="32" rx="6" fill="rgba(255,255,255,0.18)"/>
  <text x="810" y="551" text-anchor="middle" fill="#fff" font-size="11">Import / Export / Render</text>
  <rect x="690" y="570" width="240" height="14" rx="4" fill="rgba(0,0,0,0.15)"/>
  <text x="810" y="581" text-anchor="middle" fill="#A8C0E8" font-size="9">CEP (primary) &middot; Standalone Node (fallback)</text>

  <!-- Arrow: TS → Premiere Pro -->
  <line x1="810" y1="595" x2="810" y2="640" stroke="#8B95A5" stroke-width="2" marker-end="url(#arrowGray)"/>

  <!-- Fallback path -->
  <rect x="670" y="608" width="120" height="18" rx="4" fill="rgba(255,255,255,0.06)"/>
  <text x="730" y="621" text-anchor="middle" fill="#6B7B8D" font-size="9">CEP Panel</text>
  <rect x="830" y="608" width="120" height="18" rx="4" fill="rgba(255,255,255,0.06)"/>
  <text x="890" y="621" text-anchor="middle" fill="#6B7B8D" font-size="9">Node Standalone</text>

  <!-- ─── Premiere Pro ─── -->
  <rect x="620" y="645" width="360" height="65" rx="14" fill="url(#prBg)" filter="url(#shadow)"/>
  <text x="800" y="675" text-anchor="middle" fill="#fff" font-size="16" font-weight="700">Adobe Premiere Pro</text>
  <text x="800" y="698" text-anchor="middle" fill="#E0E0FF" font-size="11">Scripting DOM &middot; Timeline &middot; Media Browser &middot; Render</text>

  <!-- ─── Shared Data Layer ─── -->
  <rect x="50" y="640" width="540" height="70" rx="14" fill="rgba(255,255,255,0.04)" stroke="#2A2F3A" stroke-width="1"/>
  <text x="320" y="666" text-anchor="middle" fill="#7A8599" font-size="13" font-weight="600">Shared Data Layer</text>
  <rect x="70" y="680" width="120" height="22" rx="4" fill="rgba(255,255,255,0.06)"/>
  <text x="130" y="695" text-anchor="middle" fill="#6B7B8D" font-size="10">Asset Index (DB)</text>
  <rect x="200" y="680" width="120" height="22" rx="4" fill="rgba(255,255,255,0.06)"/>
  <text x="260" y="695" text-anchor="middle" fill="#6B7B8D" font-size="10">EDL Cache</text>
  <rect x="330" y="680" width="120" height="22" rx="4" fill="rgba(255,255,255,0.06)"/>
  <text x="390" y="695" text-anchor="middle" fill="#6B7B8D" font-size="10">Proto Definitions</text>
  <rect x="460" y="680" width="120" height="22" rx="4" fill="rgba(255,255,255,0.06)"/>
  <text x="520" y="695" text-anchor="middle" fill="#6B7B8D" font-size="10">Config &amp; Secrets</text>

  <!-- ─── Flow Example ─── -->
  <rect x="50" y="740" width="900" height="160" rx="14" fill="rgba(108,92,231,0.08)" stroke="#6C5CE7" stroke-width="1" stroke-dasharray="4"/>
  <text x="500" y="770" text-anchor="middle" fill="#B8AAFF" font-size="14" font-weight="700">Example Flow: "Edit video from script"</text>

  <rect x="80" y="790" width="150" height="44" rx="8" fill="rgba(108,92,231,0.25)"/>
  <text x="155" y="810" text-anchor="middle" fill="#D5CDFF" font-size="10" font-weight="600">1. CLI Request</text>
  <text x="155" y="825" text-anchor="middle" fill="#9B8FCC" font-size="9">"Edit with script.pdf"</text>

  <line x1="230" y1="812" x2="260" y2="812" stroke="#6C5CE7" stroke-width="1.5" marker-end="url(#arrowGray)"/>

  <rect x="262" y="790" width="140" height="44" rx="8" fill="rgba(0,173,216,0.25)"/>
  <text x="332" y="810" text-anchor="middle" fill="#7DD8F0" font-size="10" font-weight="600">2. Go Orchestrate</text>
  <text x="332" y="825" text-anchor="middle" fill="#5AAFCC" font-size="9">Fan-out to services</text>

  <line x1="402" y1="812" x2="432" y2="812" stroke="#6C5CE7" stroke-width="1.5" marker-end="url(#arrowGray)"/>

  <rect x="434" y="785" width="140" height="22" rx="6" fill="rgba(222,165,132,0.25)"/>
  <text x="504" y="800" text-anchor="middle" fill="#DEAA84" font-size="10">3a. Rust: Scan assets</text>
  <rect x="434" y="812" width="140" height="22" rx="6" fill="rgba(55,118,171,0.25)"/>
  <text x="504" y="827" text-anchor="middle" fill="#6BAADD" font-size="10">3b. Python: Parse script</text>

  <line x1="574" y1="812" x2="604" y2="812" stroke="#6C5CE7" stroke-width="1.5" marker-end="url(#arrowGray)"/>

  <rect x="606" y="790" width="140" height="44" rx="8" fill="rgba(49,120,198,0.25)"/>
  <text x="676" y="810" text-anchor="middle" fill="#6BA8E0" font-size="10" font-weight="600">4. TS: Execute EDL</text>
  <text x="676" y="825" text-anchor="middle" fill="#4A88B8" font-size="9">Timeline assembly</text>

  <line x1="746" y1="812" x2="776" y2="812" stroke="#6C5CE7" stroke-width="1.5" marker-end="url(#arrowGray)"/>

  <rect x="778" y="790" width="150" height="44" rx="8" fill="rgba(153,153,255,0.25)"/>
  <text x="853" y="810" text-anchor="middle" fill="#B8B8FF" font-size="10" font-weight="600">5. Premiere Pro</text>
  <text x="853" y="825" text-anchor="middle" fill="#8888CC" font-size="9">Render final output</text>

  <!-- ─── Legend ─── -->
  <rect x="80" y="855" width="840" height="40" rx="8" fill="rgba(255,255,255,0.03)"/>
  <rect x="100" y="869" width="14" height="14" rx="3" fill="url(#goBg)"/>
  <text x="120" y="881" fill="#8B95A5" font-size="10">Go</text>
  <rect x="165" y="869" width="14" height="14" rx="3" fill="url(#rustBg)"/>
  <text x="185" y="881" fill="#8B95A5" font-size="10">Rust</text>
  <rect x="235" y="869" width="14" height="14" rx="3" fill="url(#pyBg)"/>
  <text x="255" y="881" fill="#8B95A5" font-size="10">Python</text>
  <rect x="320" y="869" width="14" height="14" rx="3" fill="url(#tsBg)"/>
  <text x="340" y="881" fill="#8B95A5" font-size="10">TypeScript</text>
  <rect x="425" y="869" width="14" height="14" rx="3" fill="url(#prBg)"/>
  <text x="445" y="881" fill="#8B95A5" font-size="10">Premiere Pro</text>
  <text x="565" y="881" fill="#5A6577" font-size="10">──── gRPC</text>
  <text x="650" y="881" fill="#5A6577" font-size="10">- - - Flow</text>
</svg>

## Language Responsibilities

| Language | Layer | Responsibilities |
|---|---|---|
| **Go** | Orchestrator & MCP Server | MCP protocol (JSON-RPC/stdio), task orchestration via goroutines, service mesh, health checks, retry/circuit-breaker, graceful shutdown, logging & metrics |
| **Rust** | Media Processing Engine | Media probing & metadata extraction, asset indexing & fingerprinting, waveform & silence detection, thumbnail generation, high-perf file I/O via FFmpeg bindings |
| **Python** | Intelligence Layer | Script parsing & NLP, Edit Decision List generation, shot-to-asset matching via AI embeddings, pacing & timing analysis, scene detection |
| **TypeScript** | Premiere Pro Bridge | ExtendScript API integration, CEP Panel (primary bridge), standalone Node.js fallback, timeline & clip operations, import/export/render control |

## Inter-Service Communication

All services communicate via **gRPC** with shared protobuf definitions:

| Route | Protocol | Payload |
|---|---|---|
| CLI → Go | stdio / JSON-RPC 2.0 | MCP tool calls |
| Go → Rust | gRPC (protobuf) | Media scan requests, asset queries |
| Go → Python | gRPC (protobuf) | Script text, EDL generation requests |
| Go → TypeScript | gRPC / HTTP | Premiere Pro commands (EDL execution) |
| TypeScript → Premiere Pro | CEP / ExtendScript | Native Adobe scripting DOM calls |

## Bridge Fallback Strategy

The TypeScript bridge to Premiere Pro supports two modes:

1. **CEP Panel (Primary)** — Runs inside Premiere Pro as an extension panel. Direct DOM access, lowest latency. Communicates with the Go orchestrator over a local WebSocket/HTTP server.

2. **Standalone Node.js (Fallback)** — Runs as an external process. Sends commands to Premiere Pro via the ExtendScript Toolkit CLI (`osascript` / COM on Windows). Higher latency but works without a panel installed.

The Go orchestrator auto-detects which bridge is available and falls back gracefully.

## Build System

A unified build using `just` (justfile) or `make` that:

- Compiles the Rust engine (`cargo build`)
- Builds the Go orchestrator (`go build`)
- Installs Python dependencies (`pip install` / `poetry`)
- Bundles the TypeScript bridge (`npm run build`)
- Generates protobuf stubs for all four languages
- Runs all test suites
- Packages the CEP panel for Premiere Pro installation
