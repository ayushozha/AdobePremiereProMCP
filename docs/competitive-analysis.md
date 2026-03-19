# Competitive Analysis -- Premiere Pro MCP Servers

Last updated: 2026-03-19

This document surveys every known MCP server for Adobe Premiere Pro on GitHub, compares them feature-by-feature against our project (PremierPro MCP), and identifies gaps or opportunities.

---

## 1. Landscape Overview

There are **7 known Premiere Pro MCP server projects** on GitHub (including ours). None appear on any of the major awesome-mcp-servers lists (punkpeye, appcypher, wong2) or PulseMCP as of this writing.

| # | Repository | Stars | Forks | Tools | Language(s) | Bridge Type | Created | Last Push |
|---|---|---|---|---|---|---|---|---|
| 1 | **hetpatel-11/Adobe_Premiere_Pro_MCP** | 77 | 14 | 97 | TypeScript | CEP (file-based IPC) | 2025-07-07 | 2026-03-13 |
| 2 | **jordanl61/premiere-pro-mcp-server** | 12 | 2 | ~15 | JavaScript | CEP (HTTP) | 2025-07-08 | 2025-07-08 |
| 3 | **leancoderkavy/premiere-pro-mcp** | 9 | 1 | 269 | TypeScript | CEP (file-based IPC) | 2026-02-27 | 2026-03-09 |
| 4 | **antipaster/Adobe-Premiere-Pro-MCP** | 6 | 1 | 170+ | TypeScript (JS CEP) | CEP (WebSocket) | 2026-03-05 | 2026-03-18 |
| 5 | **nepfaff/premiere-pro-mcp** | 2 | 0 | 3 | TypeScript | UXP (file-based IPC) | 2026-02-09 | 2026-02-10 |
| 6 | **bis-code/premiere-pro-mcp** | 2 | 0 | ~30 (est.) | TypeScript | WebSocket | 2026-03-04 | 2026-03-04 |
| 7 | **Interweltler/premiere-mcp** | 1 | 0 | 3 | Python | ExtendScript (file) | 2025-03-22 | 2025-03-22 |
| **Ours** | **ayushozha/AdobePremiereProMCP** | 1 | 0 | **1,057** | Go + Rust + Python + TypeScript | CEP (file-based IPC via gRPC mesh) | 2026-03-19 | 2026-03-19 |

---

## 2. Detailed Competitor Profiles

### 2.1 hetpatel-11/Adobe_Premiere_Pro_MCP (77 stars) -- Primary Competitor

**Overview:** The most established project. Pure TypeScript/Node.js. Uses a CEP panel with file-based IPC through `/tmp/premiere-mcp-bridge`. Claims 97 tools with 43 live-tested and 50 schema-validated.

**Architecture:**
```
MCP Client -> Node.js MCP Server -> File-based IPC -> CEP Panel -> ExtendScript/QE DOM
```

**Key features:**
- CEP panel with live bridge status UI
- MCP resources (premiere://config/get_instructions)
- MCP prompts
- High-level workflow tools: `build_motion_graphics_demo`, `assemble_product_spot`, `build_brand_spot_from_mogrt_and_assets`
- `clipPlan` support for LLM-directed per-clip timing, transitions, motion, trims, effects, and color
- Live tool sweep verifier (`live-tool-sweep.mjs`)
- macOS auto-installer for Claude Desktop
- Also has a UXP plugin directory (for future use)
- Setup doctor script
- Jest test suite

**Tool categories (97 tools):**
- Discovery/inspection (4)
- Project management (~10)
- Timeline editing (~15)
- Effects, transitions, color (~15)
- Keyframes (~5)
- Audio (~10)
- Markers (~5)
- Export (~10)
- High-level assembly (3)
- Misc/workspace/captions (~20)

**Supported platforms:** macOS (primary), manual setup for other OS
**Premiere versions:** 2020+ (tested on 26.0)
**Clients:** Claude Desktop, Codex, Claude Code

**Strengths:**
- Most stars and community traction
- Well-documented with QUICKSTART, KNOWN_ISSUES, CONTRIBUTING guides
- Live validation sweep
- Auto-installer
- clipPlan LLM control for assembly workflows

**Weaknesses:**
- Single-language (TypeScript only)
- No Rust media processing
- No Python AI/NLP layer
- No script-to-edit pipeline
- 97 tools vs our 1,057
- No gRPC service mesh
- No health checks / circuit breakers
- No multi-platform launchers

---

### 2.2 leancoderkavy/premiere-pro-mcp (9 stars)

**Overview:** Claims 269 tools across 28 modules. TypeScript/Node.js. Very similar architecture to hetpatel (file-based CEP bridge). Published on npm. Also supports remote deployment via Fly.io with HTTP/SSE transport.

**Architecture:**
```
MCP Client -> Node.js MCP Server -> File-based IPC -> CEP Panel -> ExtendScript/QE DOM
```

**Unique features:**
- npm installable (`premiere-pro-mcp`)
- Remote deployment via Fly.io with HTTP/SSE transport
- Bearer token auth for remote connections
- MCP resources: `config://premiere-instructions` and `config://extendscript-reference`
- 28 tool modules (most granular of competitors)
- `execute_extendscript`, `evaluate_expression`, `send_raw_script` (scripting tools)
- Frame capture as base64 image
- Source monitor control tools
- DOM object inspector
- Security: script validation, 500KB limit, blocks eval/System.callSystem
- Auto-reframe for social media
- Unnest sequence

**Tool categories (269 tools):**
- Discovery & inspection (20)
- Project management (26)
- Timeline & editing (37)
- Effects & color (8)
- Keyframes (8)
- Export & encoding (14)
- Source monitor & playback (11)
- Selection & clipboard (13)
- Media properties (16)
- Sequence management (11)
- Workspace & captions (3)
- Scripting (6)
- Batch ops, metadata, tracks, navigation (100+)

**Strengths:**
- Largest tool count among TypeScript competitors
- npm distribution
- Remote deployment capability (unique)
- Script security validation
- Good documentation

**Weaknesses:**
- Single-language (TypeScript only)
- No AI/NLP pipeline
- No Rust media engine
- No native binary performance

---

### 2.3 antipaster/Adobe-Premiere-Pro-MCP (6 stars)

**Overview:** 170+ tools. TypeScript MCP server with a WebSocket bridge to CEP (not file-based). Windows-focused. Integrates ElevenLabs TTS for AI voiceover.

**Architecture:**
```
MCP Client -> Node.js MCP Server -> WebSocket -> CEP Panel -> ExtendScript
```

**Unique features:**
- WebSocket bridge (not file-based IPC) -- potentially faster/lower latency
- ElevenLabs TTS integration (3 AI tools: list_voices, generate_speech, generate_captions_voice)
- TikTok caption generator
- Caption sequence tools
- Modular ExtendScript JSX files (12 separate .jsx modules in CEP)
- Windows auto-installer (`install.bat`)
- Config file system for API keys

**Tool categories (170+ tools):**
- Project (28)
- Sequence (23)
- Timeline (25)
- Effects (19)
- Markers (8)
- Audio (10)
- Export (9)
- Graphics & captions (18)
- AI/TTS (3)
- Playback & scripting (22)
- Metadata (10)

**Strengths:**
- WebSocket bridge (lower latency than file polling)
- ElevenLabs TTS integration (unique)
- TikTok caption tool (unique)
- Modular JSX structure in CEP
- Windows-first

**Weaknesses:**
- Windows only
- WebSocket can be less reliable than file-based (firewall, port conflicts)
- Fewer tools than leancoderkavy
- No MCP resources
- No remote deployment

---

### 2.4 jordanl61/premiere-pro-mcp-server (12 stars)

**Overview:** Early prototype. ~15 tools. Plain JavaScript. Uses HTTP API bridge to CEP.

**Key features:**
- Basic project/sequence inspection
- Clip trimming by frames
- Export presets
- Render queue
- Selection info
- Includes raw ExtendScript execution files

**Strengths:**
- Simple, easy to understand
- Includes trim-by-frames UI example for CEP panel

**Weaknesses:**
- Very limited tool set (~15)
- No TypeScript
- HTTP bridge (not stdio MCP)
- Appears abandoned (last push Jul 2025)
- No tests

---

### 2.5 nepfaff/premiere-pro-mcp (2 stars)

**Overview:** Minimal server with only 3 tools. Unique approach: uses UXP instead of CEP. Exposes a single `execute-script` tool that lets the LLM generate arbitrary UXP JavaScript.

**Architecture:**
```
MCP Client -> Node.js MCP Server -> File-based IPC -> UXP Plugin -> Premiere Pro UXP API
```

**Unique features:**
- UXP-based (not CEP) -- forward-looking for future Adobe API direction
- Single `execute-script` tool -- the LLM writes all the code
- `get-results` for async result retrieval
- `get-help` provides API reference to the LLM

**Strengths:**
- UXP is Adobe's future API direction
- Maximum flexibility (LLM writes any code)
- Very simple architecture

**Weaknesses:**
- Requires UXP Developer Tool to be open
- Plugin must be reloaded each Premiere restart
- Only 3 tools
- macOS only (Windows untested)
- Requires Premiere Pro 2025 (v25.3.0+)
- Security concern: eval() of arbitrary code

---

### 2.6 bis-code/premiere-pro-mcp (2 stars)

**Overview:** WebSocket-based bridge. TypeScript. Tool configs are split into separate config files by domain. Approximately 30 tools based on config file structure.

**Tool config modules:**
- bin-configs.ts
- effects-configs.ts
- marker-configs.ts
- media-configs.ts
- placement-configs.ts
- project-configs.ts
- timeline-configs.ts

**Strengths:**
- Clean config-based tool definition
- WebSocket bridge

**Weaknesses:**
- Small tool set
- No documentation (README decode failed)
- Very new, appears minimal

---

### 2.7 Interweltler/premiere-mcp (1 star)

**Overview:** Python-based. Uses FastMCP. Only 3 tools. Oldest project (March 2025) but appears abandoned.

**Tools:**
- `get_project_info`
- `create_sequence`
- `import_media`

**Strengths:**
- Python (FastMCP) -- easy to extend
- Simple architecture

**Weaknesses:**
- Only 3 tools
- Abandoned (no updates since March 2025)
- No CEP panel included
- Minimal documentation

---

## 3. Feature Comparison Matrix

| Feature | **Ours (1,057)** | **hetpatel (97)** | **leancoderkavy (269)** | **antipaster (170+)** | **jordanl61 (~15)** | **nepfaff (3)** | **bis-code (~30)** | **Interweltler (3)** |
|---|---|---|---|---|---|---|---|---|
| **Tool Count** | 1,057 | 97 | 269 | 170+ | ~15 | 3 | ~30 | 3 |
| **Languages** | Go+Rust+Python+TS | TS | TS | TS | JS | TS | TS | Python |
| **CEP Panel** | Yes | Yes | Yes | Yes | Yes | No (UXP) | Unknown | No |
| **UXP Support** | No | Partial | No | No | No | Yes | No | No |
| **File-based IPC** | Yes | Yes | Yes | No | No | Yes | No | Yes |
| **WebSocket Bridge** | No | No | No | Yes | No | No | Yes | No |
| **HTTP Bridge** | No | No | Yes (remote) | No | Yes | No | No | No |
| **gRPC Service Mesh** | Yes | No | No | No | No | No | No | No |
| **Multi-language Architecture** | Yes (4 langs) | No | No | No | No | No | No | No |
| **Rust Media Engine** | Yes | No | No | No | No | No | No | No |
| **Python AI/NLP** | Yes | No | No | No | No | No | No | No |
| **Script-to-Edit Pipeline** | Yes | No | No | No | No | No | No | No |
| **Auto-Edit (Full Pipeline)** | Yes | No | No | No | No | No | No | No |
| **Protobuf Definitions** | Yes | No | No | No | No | No | No | No |
| **Health Checks** | Yes | No | No | No | No | No | No | No |
| **Circuit Breakers** | Yes | No | No | No | No | No | No | No |
| **macOS Support** | Yes | Yes | Yes | No | Maybe | Yes | Maybe | Maybe |
| **Windows Support** | Yes | Manual | Yes | Yes | Yes | No | Maybe | Maybe |
| **Cross-platform Launchers** | Yes (.command/.bat/.sh) | No | No | .bat only | No | No | No | No |
| **npm Installable** | No | No | Yes | No | No | No | No | No |
| **Remote Deployment** | No | No | Yes (Fly.io) | No | No | No | No | No |
| **MCP Resources** | No | Yes | Yes | No | No | No | No | No |
| **MCP Prompts** | No | Yes | No | No | No | No | No | No |
| **ElevenLabs TTS** | No | No | No | Yes | No | No | No | No |
| **TikTok Captions** | No | No | No | Yes | No | No | No | No |
| **clipPlan LLM Control** | No | Yes | No | No | No | No | No | No |
| **Live Tool Sweep** | No | Yes | No | No | No | No | No | No |
| **Auto-installer** | Justfile | npm run setup:mac | npm install-cep | install.bat | No | No | No | No |
| **Script Security** | No | No | Yes | No | No | No | No | No |
| **Frame Capture (base64)** | No | No | Yes | Yes | No | No | No | No |
| **Source Monitor Control** | No | No | Yes | Yes | No | No | No | No |
| **DOM Inspector** | No | No | Yes | No | No | No | No | No |
| **Arbitrary Script Exec** | No | No | Yes | Yes | No | Yes | No | No |
| **QE DOM Access** | Yes (via TS bridge) | Yes | Yes | Yes | No | No | No | No |
| **Color Grading (Lumetri)** | Yes (30 tools) | Partial | Yes | Partial | No | No | No | No |
| **Audio Advanced** | Yes (30 tools) | Partial | Partial | Yes (10) | No | No | No | No |
| **Keyframes** | Yes (via effects) | Yes | Yes (8) | Yes (4) | No | No | No | No |
| **Batch Operations** | Yes (30 tools) | No | Yes | No | No | No | No | No |
| **VR/Immersive** | Yes (30 tools) | No | No | No | No | No | No | No |
| **Multicam** | Yes (via workspace) | No | No | No | No | No | No | No |
| **Collaboration/Review** | Yes (30 tools) | No | No | No | No | No | No | No |
| **Diagnostics** | Yes (30 tools) | Yes (doctor) | No | No | No | No | No | No |
| **Preferences Control** | Yes (30 tools) | No | No | No | No | No | No | No |
| **Templates/Presets** | Yes (30 tools) | No | No | No | No | No | No | No |
| **Motion Graphics** | Yes (30 tools) | Yes (3) | Partial | Yes (18) | No | No | No | No |
| **App Integration (AE/PS)** | Yes (28 tools) | No | No | No | No | No | No | No |
| **Encoding Formats** | Yes (30 tools) | Partial | Yes | Yes | No | No | No | No |
| **Timeline Assembly** | Yes (30 tools) | Yes (3) | Partial | Partial | No | No | No | No |
| **Analytics/Reports** | Yes (30 tools) | No | Yes | No | No | No | No | No |
| **Effect Chains** | Yes (30 tools) | No | No | No | No | No | No | No |
| **Monitoring/Events** | Yes (30 tools) | No | No | No | No | No | No | No |
| **UI Control** | Yes (30 tools) | No | No | No | No | No | No | No |
| **Tests** | Yes | Yes (Jest) | No | No | No | No | Yes | No |

---

## 4. Our Advantages

### 4.1 Scale
- **1,057 tools** vs next closest at 269 (leancoderkavy) -- a 3.9x lead
- 38 tool modules covering every conceivable Premiere Pro operation
- Categories no competitor touches: VR/Immersive, Camera, Versioning, Delivery, Shortcuts, Effect Chains, Monitoring/Events, UI Control, Encoding Formats, App Integration, Collaboration/Review

### 4.2 Architecture
- **Four-language stack** -- each language plays to its strengths:
  - Go: fast MCP server, goroutine concurrency, low memory
  - Rust: raw performance for media scanning, indexing, waveforms
  - Python: AI/NLP for script parsing, EDL generation, shot matching
  - TypeScript: native access to Adobe ExtendScript/CEP DOM
- **gRPC service mesh** for inter-service communication
- **Protobuf definitions** for type-safe service contracts
- **Health checks with circuit breakers** -- no competitor has production-grade resilience
- Every competitor is a single-language Node.js/TypeScript project

### 4.3 AI Intelligence
- **Script-to-edit pipeline** -- parse a script, scan assets, match shots, assemble timeline automatically
- **Pacing and timing analysis** via Python
- **Shot-to-asset matching** using embeddings
- No competitor has any AI/NLP beyond basic TTS (antipaster's ElevenLabs)

### 4.4 Media Processing
- **Rust media engine** for asset scanning, fingerprinting, waveform analysis, thumbnail generation, silence detection
- No competitor has native media processing

### 4.5 Platform Support
- **Cross-platform launchers** (macOS .command, Windows .bat, Linux/universal .sh)
- **Justfile** unified build system
- Most competitors are macOS-only or Windows-only

### 4.6 Tool Categories Unique to Us
These entire categories exist only in our project:
- VR/Immersive (30 tools)
- Camera operations (30 tools)
- Versioning (30 tools)
- Delivery (30 tools)
- Shortcuts (30 tools)
- Effect Chains (30 tools)
- Monitoring & Events (30 tools)
- UI Control (30 tools)
- App Integration / Dynamic Link (28 tools)
- Collaboration & Review (30 tools)
- Preferences (30 tools)
- Templates & Presets (30 tools)
- Diagnostics (30 tools)
- Analytics (30 tools)
- Compound Operations (30 tools)

---

## 5. Features Competitors Have That We Are Missing

These are features found in competitor projects that our project does not currently implement:

### 5.1 High Priority (found in top competitor)

| Feature | Found In | Description | Priority |
|---|---|---|---|
| **MCP Resources** | hetpatel, leancoderkavy | Expose `config://premiere-instructions` and `config://extendscript-reference` as MCP resources that give the LLM operating context | High |
| **MCP Prompts** | hetpatel | MCP prompt templates for common workflows | High |
| **clipPlan LLM Control** | hetpatel | Structured `clipPlan` argument in assembly tools for LLM-directed per-clip timing, transitions, motion, trims, effects, and color | High |
| **Live Tool Sweep / Verifier** | hetpatel | Automated script that exercises every tool against a real Premiere session, writes sweep results to JSON | High |
| **Frame Capture (base64)** | leancoderkavy, antipaster | Export a frame from the timeline and return it as a base64-encoded image to the MCP client | High |

### 5.2 Medium Priority (useful differentiators)

| Feature | Found In | Description | Priority |
|---|---|---|---|
| **npm Distribution** | leancoderkavy | `npm install -g premiere-pro-mcp` for easy install | Medium |
| **Remote Deployment (Fly.io/SSE)** | leancoderkavy | HTTP/SSE transport for cloud-hosted MCP server with bearer token auth | Medium |
| **ElevenLabs TTS Integration** | antipaster | AI voiceover: list voices, generate speech, generate captions with voice | Medium |
| **TikTok Caption Generator** | antipaster | Pre-styled captions for TikTok/Reels/Shorts format | Medium |
| **WebSocket Bridge** | antipaster, bis-code | Lower-latency communication vs file polling | Medium |
| **Arbitrary ExtendScript Execution** | leancoderkavy, antipaster, nepfaff | `execute_extendscript` / `execute_qe_script` / `send_raw_script` tools | Medium |
| **Script Security Validation** | leancoderkavy | Block eval(), new Function(), System.callSystem() in scripts | Medium |
| **Source Monitor Control** | leancoderkavy, antipaster | Open clips in source monitor, insert/overwrite from source, 3-point editing | Medium |
| **DOM Object Inspector** | leancoderkavy | Interactive exploration of any Premiere Pro DOM object | Medium |

### 5.3 Low Priority (niche or future)

| Feature | Found In | Description | Priority |
|---|---|---|---|
| **UXP Plugin Support** | nepfaff | UXP-based bridge using modern Premiere Pro UXP API (vs CEP) | Low (future) |
| **Auto-reframe for Social** | leancoderkavy, antipaster | `auto_reframe_sequence` for vertical/square/social crops | Low |
| **Scene Edit Detection** | antipaster | Built-in scene detection tool | Low |
| **Setup Doctor Script** | hetpatel | `npm run setup:doctor` validates the full install chain | Low |
| **Glama.ai Badge** | hetpatel | Listed on glama.ai MCP registry | Low |

---

## 6. Key Takeaways

1. **We have the most tools by far** (1,057 vs 269 next closest), but tool count alone does not win. hetpatel has 77 stars with only 97 tools because they focus on documentation, ease of install, and a few polished high-level workflows.

2. **MCP Resources and Prompts are table stakes** that two top competitors already implement. We should add these -- they help the LLM make better decisions without counting against tool calls.

3. **Frame capture as base64** is a powerful feature that two competitors have. It lets the LLM "see" what is on the timeline, enabling visual feedback loops.

4. **ElevenLabs TTS and TikTok captions** are niche but signal that competitors are building toward social-media-first workflows. Worth considering.

5. **No competitor touches our AI/NLP pipeline or Rust media engine.** Our script-to-edit and auto-edit capabilities are entirely unique.

6. **UXP is Adobe's future direction.** Only nepfaff has a UXP plugin. Our CEP approach works for 2020-2026, but long-term we should consider UXP support.

7. **Remote deployment (Fly.io)** is unique to leancoderkavy and enables cloud-hosted MCP. This is interesting for team workflows.

8. **hetpatel's main moat is community and polish**, not technical depth. They have 77 stars because of good docs, easy install, and a glama.ai listing.

---

## 7. Recommended Actions

### Immediate (to close competitive gaps)
1. Add MCP Resources (`config://premiere-instructions`, `config://extendscript-reference`)
2. Add MCP Prompts for common workflows
3. Add `premiere_capture_frame` tool that returns a base64 image
4. Add `premiere_execute_script` tool for arbitrary ExtendScript execution
5. Add a live tool sweep / verification script

### Short-term (competitive differentiation)
6. Add `clipPlan` structured arguments to compound/assembly tools
7. Add source monitor control tools
8. Add a setup doctor / validation script
9. Submit to awesome-mcp-servers lists and glama.ai
10. Consider npm distribution for the TypeScript bridge component

### Medium-term (strategic)
11. ElevenLabs or similar TTS integration for AI voiceover workflows
12. Social media caption tools (TikTok, Reels, Shorts)
13. HTTP/SSE remote transport option
14. UXP plugin prototype for future Premiere Pro versions
15. WebSocket bridge option as an alternative to file-based IPC
