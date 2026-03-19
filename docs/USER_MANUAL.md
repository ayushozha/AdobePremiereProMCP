# PremierPro MCP Server -- User Manual

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Prerequisites](#2-prerequisites)
3. [Installation](#3-installation)
4. [Usage](#4-usage)
5. [Architecture](#5-architecture)
6. [Troubleshooting](#6-troubleshooting)
7. [Quick Reference -- Common Commands](#7-quick-reference----common-commands)
8. [Build Commands](#8-build-commands)
9. [Environment Variables](#9-environment-variables)
10. [Default Paths](#10-default-paths)

---

## 1. Introduction

### What is PremierPro MCP?

PremierPro MCP is an open-source server that implements the [Model Context Protocol (MCP)](https://modelcontextprotocol.io) for Adobe Premiere Pro. It turns Premiere Pro into an AI-controllable video editing application, allowing you to describe edits in plain English and have them executed automatically.

The server accepts natural-language instructions from any MCP-compatible AI assistant -- such as Claude, GPT, or Codex -- and translates them into real Premiere Pro operations: creating sequences, importing media, placing clips on the timeline, applying effects and color grading, mixing audio, adding graphics, and exporting finished videos.

No plugins. No subscriptions. Fully open source under the MIT license.

### What can it do?

PremierPro MCP provides **907 tools** organized across 36 source files, covering every major aspect of Premiere Pro:

| Category | Tools | What You Can Do |
|---|---|---|
| **Core/Foundation** | 14 | Ping, get project state, create sequences, import media, place clips, export |
| **App Lifecycle** | 3 | Launch, quit, and check Premiere Pro process status |
| **Project Management** | 23 | Create, open, save, close projects; manage bins, scratch disks, metadata |
| **Sequence Management** | 26 | Create, duplicate, delete sequences; playhead, in/out points, markers, nesting |
| **Clip Operations** | 29 | Insert, overwrite, move, trim, split, slip, slide, speed, link/unlink clips |
| **Effects & Transitions** | 36 | Apply/remove effects and transitions, keyframe animation, motion, Lumetri basics |
| **Audio (basic)** | 32 | Levels, gain, mute/solo, effects, Essential Sound, track management |
| **Audio (advanced)** | 30 | Mixer state, EQ, compressor, limiter, de-esser, loudness, sync, waveform analysis |
| **Color Grading** | 30 | Full Lumetri Color: exposure, contrast, curves, HSL, color wheels, LUTs, vignette |
| **Graphics & Titles** | 21 | MOGRTs, titles, lower thirds, captions, color mattes, time remapping |
| **Export (basic)** | 14 | Direct export, AME queue, frame export, AAF/OMF/FCPXML, audio-only export |
| **Advanced Editing** | 31 | Ripple/roll/slip/slide trims, gap management, grouping, snapping, navigation |
| **Batch Operations** | 30 | Batch import/export, apply effects to multiple clips, auto-organize, markers |
| **AI/ML Workflows** | 25 | Smart cut, auto color match, rough cut, B-roll suggestions, social cuts, analysis |
| **Workspace & Multicam** | 25 | Multicam, proxy management, workspaces, undo/redo, source monitor, cache |
| **Playback & Navigation** | 30 | Play/pause/stop, shuttle, step, loop, timecode navigation, render status |
| **Transform & Masking** | 30 | Crop, PIP, fade, stabilizer, noise reduction, blur, sharpen, distortion |
| **Metadata & Labels** | 30 | XMP metadata, labels, footage interpretation, smart bins, media management |
| **Preferences** | 30 | Still/transition durations, auto-save, playback resolution, cache, renderer, codecs |
| **Templates & Presets** | 30 | Sequence/effect/export presets, project templates, batch rename, macros |
| **Motion Graphics** | 30 | Essential Graphics, scrolling titles, shapes, watermarks, split screen, subtitles |
| **Collaboration & Review** | 30 | Review comments, version history, snapshots, EDL/AAF/XML import, delivery checks |
| **VR/Immersive** | 30 | VR projection, HDR, stereoscopic 3D, frame rates, letterboxing, timecode, captions |
| **App Integration** | 28 | Dynamic Link (After Effects), Photoshop, Audition, Media Encoder, Team Projects |
| **Diagnostics** | 30 | Performance metrics, disk space, plugins, render status, health checks, debug logs |
| **Monitoring & Events** | 30 | Event listeners, playhead/render watchers, state snapshots, notifications |
| **UI Control** | 30 | Panel management, window control, track display, label filters, dialogs, console |
| **Compound Operations** | 30 | Montage, slideshow, highlight reel, music bed, social exports, project setup |
| **Encoding & Formats** | 30 | Codec conversion (ProRes, H.264/265, DNxHR, GIF), thumbnails, render queue |
| **Timeline Assembly** | 30 | EDL/CSV assembly, clip sorting/shuffling, compositing, generators, timeline reports |
| **Scripting** | 30 | ExtendScript execution, global variables, conditionals, scheduling, file I/O |
| **Analytics** | 30 | Project/sequence summaries, codec/resolution breakdowns, pacing, comparison reports |
| **Effect Chains** | 30 | Effect chain management, visual presets (sepia, vintage, glow), transition control |

### Supported Premiere Pro Versions

| Version | Year | Support Level |
|---|---|---|
| 14.x | 2020 | Community tested |
| 15.x | 2021 | Community tested |
| 22.x | 2022 | Community tested |
| 23.x | 2023 | Supported |
| 24.x | 2024 | Supported |
| 25.x | 2025 | Primary target |
| 26.x | 2026 | Beta support |

The CEP extension manifest declares compatibility from Premiere Pro version 14.0 (2020) onward.

### Supported Platforms

| Component | macOS | Windows | Linux |
|---|---|---|---|
| MCP server (Go orchestrator) | Yes | Yes | Yes |
| Rust media engine | Yes | Yes | Yes |
| Python intelligence layer | Yes | Yes | Yes |
| TypeScript bridge | Yes | Yes | Yes |
| CEP panel (inside Premiere Pro) | Yes | Yes | N/A |
| Adobe Premiere Pro | Yes | Yes | N/A |

The server itself runs on all three platforms. Adobe Premiere Pro is only available on macOS and Windows, so the CEP panel bridge requires one of those two operating systems. On Linux, you can run the server and connect to a remote Premiere Pro instance, or use it in headless/testing mode.

---

## 2. Prerequisites

Install the following before setting up PremierPro MCP:

| Tool | Minimum Version | Purpose | Install |
|---|---|---|---|
| [Go](https://go.dev/) | 1.22+ | MCP server and orchestrator | [go.dev/dl](https://go.dev/dl/) |
| [Rust](https://rustup.rs/) | 1.77+ | Media processing engine | `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \| sh` |
| [Python](https://python.org/) | 3.12+ | AI intelligence layer | [python.org/downloads](https://python.org/downloads/) |
| [Node.js](https://nodejs.org/) | 20+ | TypeScript bridge and CLI | [nodejs.org](https://nodejs.org/) |
| [just](https://github.com/casey/just) | latest | Unified build system | `cargo install just` or `brew install just` |
| [buf](https://buf.build/) | latest | Protobuf code generation | [buf.build/docs/installation](https://buf.build/docs/installation) |
| [FFmpeg](https://ffmpeg.org/) | latest | Media scanning and processing | `brew install ffmpeg` or [ffmpeg.org/download](https://ffmpeg.org/download.html) |
| Adobe Premiere Pro | 2020 (v14) or later | Target application | [adobe.com](https://www.adobe.com/products/premiere.html) |

You also need an API key from one of the supported AI providers:

- **Anthropic** -- for Claude models (recommended)
- **OpenAI** -- for GPT/Codex models

---

## 3. Installation

### 3.1 Clone and Build

```bash
# Clone the repository
git clone https://github.com/ayushozha/AdobePremiereProMCP.git
cd PremierProMCP

# Copy the environment template
cp .env.example .env

# Install dependencies (Python + Node.js)
just install

# Generate protobuf stubs for all languages
just proto

# Build all components (Go, Rust, TypeScript, CEP panel)
just build
```

The `just build` command runs the following in sequence:
1. `buf generate` -- generates protobuf stubs for Go, Rust, Python, and TypeScript
2. `go build` -- compiles the Go orchestrator to `go-orchestrator/bin/server`
3. `cargo build --release` -- compiles the Rust media engine
4. `npm run build` -- bundles the TypeScript bridge
5. `npm run build` -- bundles the CEP panel

To verify everything compiled correctly:

```bash
just test
```

### 3.2 Install the CEP Panel

```bash
just install-panel
```

This script does two things:

1. **Symlinks** the `cep-panel/` directory into Premiere Pro's CEP extensions folder:
   - macOS: `~/Library/Application Support/Adobe/CEP/extensions/com.premierpro.mcp.bridge`
   - Windows: `%APPDATA%\Adobe\CEP\extensions\com.premierpro.mcp.bridge`

2. **Enables unsigned extensions** by setting `PlayerDebugMode=1` on the CSXS.11 registry/preference, which is required for development extensions to load.

After running this command, restart Premiere Pro if it is already open.

### 3.3 Start Backend Services

```bash
just start
```

This launches three background services:

| Service | Language | Default Port | Log File |
|---|---|---|---|
| Media engine | Rust | 50052 | `scripts/logs/rust-engine.log` |
| Intelligence layer | Python | 50053 | `scripts/logs/python-intelligence.log` |
| Premiere Pro bridge | TypeScript | 50054 | `scripts/logs/ts-bridge.log` |

The Go orchestrator (the MCP server itself) is not started here -- it is spawned on demand by the CLI or by your MCP client when it connects.

To check that all services are running:

```bash
just status
```

To stop all services:

```bash
just stop
```

### 3.4 Open Premiere Pro

1. Open **Adobe Premiere Pro**.
2. Open an existing project or create a new one. (The Extensions menu is grayed out until a project is open.)
3. Go to **Window > Extensions > PremierPro MCP Bridge**.
4. The panel opens inside Premiere Pro and its WebSocket server starts automatically.
5. You should see a "Connected" indicator in the panel.

The panel is lightweight (300x200 pixels by default) and can be docked anywhere in the Premiere Pro workspace.

### 3.5 Open or Create a Project

You can open or create a project in one of two ways:

**Manually in Premiere Pro:**
- Use File > Open or File > New > Project as usual.
- The default projects folder is: `~/Documents/Adobe/Premiere Pro/{version}/`

**Via the AI assistant:**
- Type `Open the Testing project` in the CLI, and the system will search for and open matching projects.
- Type `Create a new project called "My Video"` to create one from scratch.

---

## 4. Usage

### 4.1 Interactive CLI (Recommended)

The simplest way to get started is the one-click launcher for your platform:

- **macOS:** Double-click `PremierPro.command` (or run `./PremierPro.command` in Terminal)
- **Windows:** Double-click `PremierPro.bat`
- **Linux:** Run `./PremierPro.sh`

Alternatively, run directly from the terminal:

```bash
npx --prefix cli tsx cli/src/index.ts
```

The launcher performs the following steps automatically:

1. **Resolves authentication** -- checks for API keys in environment variables, Claude Code auth, Codex CLI auth, or config files.
2. **Installs dependencies** if missing (CLI and bridge `node_modules`).
3. **Builds the MCP server binary** if it does not exist.
4. **Starts backend services** if they are not already running.
5. **Installs the CEP panel** if not already symlinked.
6. **Connects to the MCP server** via stdio.
7. **Auto-launches Premiere Pro** if it is not running.
8. **Enters interactive chat mode**.

Once in the interactive loop, type natural-language commands:

```
you> Open Premiere Pro
you> Create a new sequence called "My Edit" at 1080p 24fps
you> Import all videos from /Users/me/footage/
you> Place the first clip on the timeline at 0 seconds
you> Add a cross dissolve transition between the first two clips
you> Set the Lumetri contrast to 20
you> Export as H.264 1080p to /Users/me/output.mp4
you> exit
```

More examples:

```
you> Parse the script at /Users/me/scripts/video-script.pdf
you> Scan assets in /Users/me/footage/ and show me what's available
you> Auto-edit using script.pdf with footage from /Users/me/media/
you> Apply a warm color grade to all clips on track V1
you> Set audio levels on the interview track to -6 dB
you> Add a lower third saying "John Smith, CEO" at 00:01:30
you> Export three versions: 1080p H.264, 4K ProRes, and a GIF preview
```

To quit the interactive session, type `exit`, `quit`, or `q`, or press Ctrl+C.

### 4.2 As an MCP Server (Claude Code, Cursor, etc.)

To use PremierPro MCP as a tool server for Claude Code, Cursor, or any MCP-compatible client, add it to your MCP configuration file:

```json
{
  "mcpServers": {
    "premiere-pro": {
      "command": "/path/to/AdobePremiereProMCP/go-orchestrator/bin/server",
      "args": ["--transport", "stdio"]
    }
  }
}
```

Replace `/path/to/AdobePremiereProMCP` with the actual path to your clone of the repository.

The binary name is `server` (built by `just go-build`) or `premierpro-mcp` (built by the launcher scripts). Either binary works -- they are built from the same source (`go-orchestrator/cmd/server/main.go`).

**SSE transport** is also available if your client supports it:

```bash
./go-orchestrator/bin/server --transport sse --port 8080
```

This starts an HTTP server at `http://localhost:8080` with Server-Sent Events for the MCP protocol.

### 4.3 Authentication

The system supports multiple authentication methods, checked in this priority order:

| Priority | Method | How to Set Up |
|---|---|---|
| 1 | `ANTHROPIC_API_KEY` environment variable | `export ANTHROPIC_API_KEY="sk-ant-..."` |
| 2 | `OPENAI_API_KEY` environment variable | `export OPENAI_API_KEY="sk-..."` |
| 3 | Claude Code CLI auth | Run `claude auth login --console` |
| 4 | Codex CLI auth | Run `codex login` |
| 5 | Config files | Create `~/.premierpro-mcp/config.json` with your key |

**Config file format** (`~/.premierpro-mcp/config.json`):

```json
{
  "anthropic_api_key": "sk-ant-...",
  "model": "claude-sonnet-4-20250514"
}
```

Or for OpenAI:

```json
{
  "openai_api_key": "sk-...",
  "model": "gpt-4o"
}
```

**Default models:**
- Anthropic: `claude-sonnet-4-20250514`
- OpenAI: `gpt-4o`

Override the model by setting the `MODEL` environment variable:

```bash
export MODEL="claude-opus-4-20250514"
```

**Note on OAuth:** If you are logged into Claude via `claude.ai` OAuth (browser-based login), the system cannot extract an API key. You will need to either set `ANTHROPIC_API_KEY` manually or re-authenticate with `claude auth login --console` which uses API-key-based auth.

---

## 5. Architecture

PremierPro MCP uses a four-language architecture, with each language chosen for its strengths in a specific domain:

```
CLI / MCP Client (Claude, GPT, any AI)
       | stdio / JSON-RPC
       v
+-------------------------------------+
|     Go -- MCP Server & Orchestrator  |
|  Protocol handling . Concurrency     |
|  Service mesh . Health & recovery    |
+------+------------+------------+-----+
       | gRPC       | gRPC       | gRPC
       v            v            v
+------------+ +----------+ +----------------+
|   Rust     | |  Python  | |  TypeScript     |
|   Media    | |  Intel   | |  Premiere Pro   |
|   Engine   | |  Layer   | |  Bridge         |
+------------+ +----------+ +-------+--------+
                                    | CEP / ExtendScript
                                    v
                             Adobe Premiere Pro
```

### Go -- MCP Server and Orchestrator

The Go layer is the entry point for the system. It implements the MCP protocol (JSON-RPC 2.0 over stdio or SSE), registers all 907 tools, and orchestrates requests by fanning out to the downstream services via gRPC. It uses goroutines for concurrency, implements retry logic and circuit breakers, and handles graceful shutdown.

- **Directory:** `go-orchestrator/`
- **Entry point:** `cmd/server/main.go`
- **Tool definitions:** `internal/mcp/*_tools.go` (36 files)
- **Default gRPC port:** 50051 (inbound, when used with SSE)

### Rust -- Media Processing Engine

The Rust layer handles performance-critical media operations: scanning directories for media files, extracting metadata (codec, resolution, duration, frame rate), generating waveforms for silence detection, creating thumbnails, and indexing assets. It uses FFmpeg bindings for media processing and zero-copy I/O for performance.

- **Directory:** `rust-engine/`
- **Modules:** `media/`, `assets/`, `waveform/`, `thumbnails/`
- **Default gRPC port:** 50052

### Python -- Intelligence Layer

The Python layer handles AI and NLP tasks: parsing scripts (screenplay, YouTube, podcast formats), generating Edit Decision Lists (EDLs), matching script segments to media assets using AI embeddings, and analyzing pacing and timing. It uses LLMs, embedding models, and scene detection algorithms.

- **Directory:** `python-intelligence/`
- **Modules:** `parser/`, `edl/`, `matching/`, `analysis/`
- **Default gRPC port:** 50053

### TypeScript -- Premiere Pro Bridge

The TypeScript layer is the bridge between the Go orchestrator and Adobe Premiere Pro. It translates gRPC commands into ExtendScript API calls that Premiere Pro can execute. It supports two modes:

1. **CEP Panel (primary):** Runs inside Premiere Pro as an extension panel. Has direct access to the Premiere Pro DOM. Lowest latency. Communicates with the Go orchestrator over a local WebSocket/HTTP connection.

2. **Standalone Node.js (fallback):** Runs as an external process. Sends commands to Premiere Pro via `osascript` (macOS) or COM automation (Windows). Higher latency but works without the panel installed.

The Go orchestrator auto-detects which bridge mode is available and falls back gracefully.

- **Directory:** `ts-bridge/`
- **Modules:** `extendscript/`, `cep/`, `standalone/`, `timeline/`
- **Default gRPC port:** 50054

### CEP Panel

The CEP panel is an Adobe Common Extensibility Platform extension that runs inside Premiere Pro. It provides a small UI panel and, more importantly, acts as the host for ExtendScript execution. The panel's WebSocket server listens for commands from the TypeScript bridge and executes them against Premiere Pro's scripting DOM.

- **Directory:** `cep-panel/`
- **Panel menu name:** PremierPro MCP Bridge
- **Extension ID:** `com.premierpro.mcp.bridge`
- **CSXS version:** 11.0

### Inter-Service Communication

All services communicate via gRPC with shared protobuf definitions stored in `proto/definitions/`:

| Route | Protocol | Payload |
|---|---|---|
| CLI / Client -> Go | stdio / JSON-RPC 2.0 | MCP tool calls |
| Go -> Rust | gRPC (protobuf) | Media scan requests, asset queries |
| Go -> Python | gRPC (protobuf) | Script text, EDL generation requests |
| Go -> TypeScript | gRPC / HTTP | Premiere Pro commands (EDL execution) |
| TypeScript -> Premiere Pro | CEP / ExtendScript | Native Adobe scripting DOM calls |

### End-to-End Flow Example

When you type *"Edit this video using script.pdf with footage from /media/"*:

1. **CLI** sends the prompt to the AI model (Claude or GPT).
2. The AI model identifies the appropriate MCP tool (`premiere_auto_edit`) and calls it.
3. **Go orchestrator** receives the tool call and fans out:
   - **Rust engine** scans `/media/`, indexes all assets (codec, duration, resolution, waveforms).
   - **Python intelligence** parses `script.pdf`, generates an Edit Decision List, matches shots to assets using AI embeddings.
4. **Go merges results** and sends the assembled EDL to the TypeScript bridge.
5. **TypeScript bridge** executes in Premiere Pro -- creates the sequence, places clips, adds transitions and text.
6. **Premiere Pro renders** the final output.

---

## 6. Troubleshooting

### Panel not showing in the Extensions menu

- **Open a project first.** The Extensions menu is grayed out in Premiere Pro until a project is open.
- **Reinstall the panel:** Run `just install-panel` and restart Premiere Pro.
- **Verify PlayerDebugMode is enabled.** On macOS:
  ```bash
  defaults read com.adobe.CSXS.11 PlayerDebugMode
  ```
  This should return `1`. If not, run:
  ```bash
  defaults write com.adobe.CSXS.11 PlayerDebugMode 1
  ```
  On Windows, check the registry key:
  ```
  HKEY_CURRENT_USER\Software\Adobe\CSXS.11\PlayerDebugMode = 1
  ```
- **Check the symlink.** Verify the extension directory exists:
  ```bash
  ls -la "$HOME/Library/Application Support/Adobe/CEP/extensions/com.premierpro.mcp.bridge"
  ```
  It should be a symlink pointing to your `cep-panel/` directory.

### "EvalScript error"

This typically means the ExtendScript code failed to execute inside Premiere Pro.

- The ExtendScript file may be too large to load in a single evaluation. The CEP panel splits large scripts automatically, but check if the issue persists after restarting Premiere Pro.
- Check CEP logs for detailed error messages:
  ```bash
  # macOS
  ls ~/Library/Logs/CSXS/CEPHtmlEngine12-PPRO-*.log
  cat ~/Library/Logs/CSXS/CEPHtmlEngine12-PPRO-*.log | tail -50
  ```
- Try restarting Premiere Pro completely (quit and reopen).
- Make sure you are running a supported Premiere Pro version (2020 or later).

### Services not starting

If `just start` fails or services die immediately:

- **Check if ports are already in use:**
  ```bash
  lsof -i :50052 :50053 :50054
  ```
  If another process is using a port, either stop that process or configure a different port in `.env`.

- **Check individual service logs:**
  ```bash
  cat scripts/logs/rust-engine.log
  cat scripts/logs/python-intelligence.log
  cat scripts/logs/ts-bridge.log
  ```

- **Check service status:**
  ```bash
  just status
  ```

- **Make sure all prerequisites are installed.** Verify each one:
  ```bash
  go version        # Should be 1.22+
  rustc --version   # Should be 1.77+
  python3 --version # Should be 3.12+
  node --version    # Should be 20+
  ffmpeg -version   # Should be installed
  ```

### WebSocket not connecting

The TypeScript bridge communicates with the CEP panel over a local WebSocket connection.

- **Make sure the CEP panel is open** in Premiere Pro (Window > Extensions > PremierPro MCP Bridge). The panel must be visible for its WebSocket server to be active.
- **Check the bridge log** for connection errors:
  ```bash
  cat scripts/logs/ts-bridge.log
  ```
- If using the standalone (non-CEP) bridge mode, ensure the `FALLBACK_TO_STANDALONE=true` setting is in your `.env` file.

### Python service fails to start

- **Install Python dependencies manually:**
  ```bash
  cd python-intelligence
  pip install -e ".[dev]"
  ```
  Or install specific packages:
  ```bash
  pip install grpcio protobuf pydantic structlog numpy scikit-learn pypdf python-docx
  ```
- **Make sure protobuf is up to date:**
  ```bash
  pip install "protobuf>=5.0.0"
  ```
- **Check the Python version.** Python 3.12 or later is required:
  ```bash
  python3 --version
  ```

### Rust engine fails to build

- **Update Rust toolchain:**
  ```bash
  rustup update
  ```
- **Check for missing system dependencies.** On macOS, you may need Xcode command-line tools:
  ```bash
  xcode-select --install
  ```
- **Verify FFmpeg is installed** and the `FFMPEG_PATH` in `.env` is correct:
  ```bash
  which ffmpeg
  ```

### No API key found

If the launcher shows "No API key found":

- Set an environment variable:
  ```bash
  export ANTHROPIC_API_KEY="sk-ant-..."
  ```
- Or authenticate via CLI:
  ```bash
  claude auth login --console   # For Anthropic
  codex login                   # For OpenAI
  ```
- Or create a config file at `~/.premierpro-mcp/config.json` (see [Authentication](#43-authentication) for format).

### MCP server binary not found

If the CLI reports it cannot find the server binary:

```bash
cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server
```

Or build everything at once:

```bash
just build
```

---

## 7. Quick Reference -- Common Commands

These are natural-language commands you can type in the interactive CLI. The AI assistant will translate them into the appropriate MCP tool calls.

| What You Want to Do | MCP Tool | Example Prompt |
|---|---|---|
| Open Premiere Pro | `premiere_open` | "Open Premiere Pro" |
| Check if PP is running | `premiere_is_running` | "Is Premiere Pro running?" |
| Open a project | `premiere_open_project` | "Open the project called Testing" |
| Get project info | `premiere_get_project` | "What project is currently open?" |
| Create a sequence | `premiere_create_sequence` | "Create a 1080p 24fps sequence called My Edit" |
| Get timeline state | `premiere_get_timeline` | "Show me what's on the timeline" |
| Import media | `premiere_import_media` | "Import all videos from /Users/me/footage/" |
| Place a clip | `premiere_place_clip` | "Put the first clip on track V1 at 0 seconds" |
| Remove a clip | `premiere_remove_clip` | "Remove the clip at the beginning of track V1" |
| Add a transition | `premiere_add_transition` | "Add a cross dissolve between clips 1 and 2" |
| Add text overlay | `premiere_add_text` | "Add title text saying Hello World at 5 seconds" |
| Set audio level | `premiere_set_audio_level` | "Set the audio on this clip to -6 dB" |
| Color grading | `premiere_lumetri_set_*` | "Increase the contrast to 30" |
| Apply an effect | `premiere_apply_effect` | "Apply Gaussian Blur to this clip" |
| Scan media assets | `premiere_scan_assets` | "Scan /Users/me/footage/ for media files" |
| Parse a script | `premiere_parse_script` | "Parse the script at /Users/me/script.pdf" |
| Auto-edit from script | `premiere_auto_edit` | "Edit using script.pdf with footage from /media/" |
| Export video | `premiere_export` | "Export as H.264 1080p to /Users/me/output.mp4" |
| Close Premiere Pro | `premiere_close` | "Close Premiere Pro" |

---

## 8. Build Commands

All build commands use `just` as the unified build system. Run `just` with no arguments to see the full list.

### Top-Level Commands

| Command | Description |
|---|---|
| `just build` | Build all components (proto, Go, Rust, TypeScript, CEP) |
| `just test` | Run all test suites |
| `just lint` | Lint all code (Go, Rust, Python, TypeScript, proto) |
| `just ci` | Full CI pipeline: lint, build, then test |
| `just clean` | Remove all build artifacts |
| `just install` | Install all dependencies (Python + Node.js) |

### Protobuf

| Command | Description |
|---|---|
| `just proto` | Generate protobuf stubs for all languages |
| `just proto-lint` | Lint protobuf definitions |

### Go Orchestrator

| Command | Description |
|---|---|
| `just go-build` | Build the Go orchestrator to `go-orchestrator/bin/server` |
| `just go-run` | Run the Go orchestrator directly |
| `just go-test` | Run Go tests |
| `just go-lint` | Lint Go code with `golangci-lint` |

### Rust Engine

| Command | Description |
|---|---|
| `just rust-build` | Build the Rust engine in release mode |
| `just rust-test` | Run Rust tests |
| `just rust-lint` | Lint Rust code with `clippy` |

### Python Intelligence

| Command | Description |
|---|---|
| `just py-install` | Install Python dependencies (editable mode with dev extras) |
| `just py-test` | Run Python tests with `pytest` |
| `just py-lint` | Lint Python code with `ruff` and type-check with `mypy` |

### TypeScript Bridge

| Command | Description |
|---|---|
| `just ts-install` | Install Node.js dependencies |
| `just ts-build` | Build the TypeScript bridge |
| `just ts-test` | Run TypeScript tests |
| `just ts-lint` | Lint TypeScript code |

### CEP Panel

| Command | Description |
|---|---|
| `just cep-build` | Build the CEP panel |
| `just cep-package` | Package the CEP panel for distribution |

### Services

| Command | Description |
|---|---|
| `just start` | Start all backend services (Rust, Python, TypeScript) |
| `just stop` | Stop all backend services |
| `just status` | Check status of all backend services |
| `just install-panel` | Install (symlink) the CEP panel into Premiere Pro |

### Development

| Command | Description |
|---|---|
| `just dev` | Start all services in development mode |

---

## 9. Environment Variables

Configuration is managed through environment variables. Copy `.env.example` to `.env` and edit as needed. CLI flags override environment variables for the Go orchestrator.

### General

| Variable | Default | Description |
|---|---|---|
| `LOG_LEVEL` | `info` | Global log level (used by `.env.example`) |
| `ENV` | `development` | Environment name |

### Go Orchestrator

| Variable | Default | Description |
|---|---|---|
| `MCP_TRANSPORT` | `stdio` | MCP transport type: `stdio` or `sse` |
| `MCP_SSE_PORT` | `8080` | Port for the SSE HTTP server (only used with `sse` transport) |
| `MCP_LOG_LEVEL` | `info` | Log level for the Go orchestrator: `debug`, `info`, `warn`, `error` |

### Service Addresses

| Variable | Default | Description |
|---|---|---|
| `RUST_ENGINE_ADDR` | `localhost:50052` | gRPC address of the Rust media engine |
| `PYTHON_INTEL_ADDR` | `localhost:50053` | gRPC address of the Python intelligence service |
| `TS_BRIDGE_ADDR` | `localhost:50054` | gRPC address of the TypeScript Premiere Pro bridge |

### Service Timeouts

| Variable | Default | Description |
|---|---|---|
| `RUST_ENGINE_TIMEOUT` | `30` | Timeout in seconds for Rust engine gRPC calls |
| `PYTHON_INTEL_TIMEOUT` | `60` | Timeout in seconds for Python intelligence gRPC calls |
| `TS_BRIDGE_TIMEOUT` | `30` | Timeout in seconds for TypeScript bridge gRPC calls |

### Rust Engine

| Variable | Default | Description |
|---|---|---|
| `RUST_GRPC_PORT` | `50052` | gRPC port for the Rust media engine |
| `FFMPEG_PATH` | `/usr/local/bin/ffmpeg` | Path to the FFmpeg binary |
| `ASSET_CACHE_DIR` | `./tmp/asset-cache` | Directory for cached asset metadata |

### Python Intelligence

| Variable | Default | Description |
|---|---|---|
| `PYTHON_GRPC_PORT` | `50053` | gRPC port for the Python intelligence service |

### TypeScript Bridge

| Variable | Default | Description |
|---|---|---|
| `TS_BRIDGE_PORT` | `50054` | gRPC port for the TypeScript bridge |
| `PREMIERE_PRO_PATH` | `/Applications/Adobe Premiere Pro 2025/Adobe Premiere Pro 2025.app` | Path to the Premiere Pro application |
| `CEP_PANEL_MODE` | `true` | Whether to use the CEP panel as the primary bridge |
| `FALLBACK_TO_STANDALONE` | `true` | Whether to fall back to standalone Node.js bridge if CEP is unavailable |

### Authentication

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | (none) | Anthropic API key for Claude models |
| `OPENAI_API_KEY` | (none) | OpenAI API key for GPT/Codex models |
| `MODEL` | (auto) | Override the default AI model (e.g., `claude-opus-4-20250514`, `gpt-4o`) |

---

## 10. Default Paths

### macOS

| Item | Path |
|---|---|
| Premiere Pro projects | `~/Documents/Adobe/Premiere Pro/{version}/` |
| CEP extensions | `~/Library/Application Support/Adobe/CEP/extensions/` |
| CEP panel (this project) | `~/Library/Application Support/Adobe/CEP/extensions/com.premierpro.mcp.bridge` |
| CEP logs | `~/Library/Logs/CSXS/` |
| CEP engine logs | `~/Library/Logs/CSXS/CEPHtmlEngine12-PPRO-*.log` |
| Auto-save | Near project file in `Adobe Premiere Pro Auto-Save/` |
| Media cache | `~/Library/Application Support/Adobe/Common/Media Cache Files/` |
| Media cache database | `~/Library/Application Support/Adobe/Common/Media Cache/` |
| Premiere Pro preferences | `~/Library/Preferences/com.adobe.PremierePro.plist` |
| PremierPro MCP config | `~/.premierpro-mcp/config.json` |
| Claude credentials | `~/.claude/credentials.json` |
| Service logs | `{project}/scripts/logs/` |
| Service PID file | `{project}/scripts/.pids` |

### Windows

| Item | Path |
|---|---|
| Premiere Pro projects | `%USERPROFILE%\Documents\Adobe\Premiere Pro\{version}\` |
| CEP extensions | `%APPDATA%\Adobe\CEP\extensions\` |
| CEP panel (this project) | `%APPDATA%\Adobe\CEP\extensions\com.premierpro.mcp.bridge` |
| CEP logs | `%USERPROFILE%\AppData\Local\Temp\csxs12-PPRO-*.log` |
| Auto-save | Near project file in `Adobe Premiere Pro Auto-Save\` |
| Media cache | `%APPDATA%\Adobe\Common\Media Cache Files\` |
| Premiere Pro preferences | Registry: `HKCU\Software\Adobe\Premiere Pro\` |
| PlayerDebugMode | Registry: `HKCU\Software\Adobe\CSXS.11\PlayerDebugMode` |
| PremierPro MCP config | `%USERPROFILE%\.premierpro-mcp\config.json` |
| Service logs | `{project}\scripts\logs\` |

### Project Structure

```
PremierProMCP/
+-- go-orchestrator/          # Go -- MCP server & task orchestrator
|   +-- cmd/server/           #   Entry point (main.go)
|   +-- internal/             #   Core packages
|   |   +-- mcp/              #     MCP protocol handler (907 tool definitions, 36 files)
|   |   +-- orchestrator/     #     Task orchestration
|   |   +-- health/           #     Health checks
|   |   +-- grpc/             #     gRPC client/server
|   |   +-- config/           #     Configuration loading
|   +-- configs/              #   Default configuration (defaults.yaml)
|
+-- rust-engine/              # Rust -- Media processing engine
|   +-- src/
|       +-- media/            #   Media probe & metadata
|       +-- assets/           #   Asset indexing & fingerprinting
|       +-- waveform/         #   Waveform & silence detection
|       +-- thumbnails/       #   Thumbnail generation
|
+-- python-intelligence/      # Python -- AI intelligence layer
|   +-- src/
|   |   +-- parser/           #   Script parsing & NLP
|   |   +-- edl/              #   Edit Decision List generation
|   |   +-- matching/         #   Shot-to-asset matching
|   |   +-- analysis/         #   Pacing & timing analysis
|   +-- tests/
|   +-- models/               #   ML model configs
|
+-- ts-bridge/                # TypeScript -- Premiere Pro bridge
|   +-- src/
|       +-- extendscript/     #   ExtendScript API layer
|       +-- cep/              #   CEP Panel bridge (primary)
|       +-- standalone/       #   Node.js fallback bridge
|       +-- timeline/         #   Timeline operations
|
+-- cep-panel/                # CEP Panel -- Premiere Pro extension
|   +-- CSXS/                 #   Adobe extension manifest (manifest.xml)
|   +-- src/
|   |   +-- host/             #   ExtendScript files (core.jsx, premiere.jsx)
|   |   +-- index.html        #   Panel UI
|   |   +-- panel.js          #   Panel logic
|   |   +-- CSInterface.js    #   Adobe CSInterface library
|   +-- assets/
|
+-- cli/                      # Interactive CLI
|   +-- src/
|       +-- index.ts          #   Entry point
|       +-- auth.ts           #   Authentication resolution
|       +-- chat.ts           #   AI chat loop
|       +-- mcp-client.ts     #   MCP client (spawns Go server)
|       +-- ui.ts             #   Terminal UI helpers
|
+-- proto/                    # Shared protobuf definitions
|   +-- definitions/
|
+-- gen/                      # Generated protobuf stubs
+-- shared/                   # Shared utilities
+-- scripts/                  # Build & setup scripts
|   +-- start-all.sh          #   Start all backend services
|   +-- stop-all.sh           #   Stop all backend services
|   +-- status.sh             #   Check service status
|   +-- install-cep-panel.sh  #   Install CEP panel into Premiere Pro
|   +-- logs/                 #   Service log files
|
+-- docs/                     # Documentation
+-- Justfile                  # Unified build system
+-- .env.example              # Environment variable template
+-- PremierPro.command        # macOS launcher
+-- PremierPro.bat            # Windows launcher
+-- PremierPro.sh             # Linux launcher
```

---

## Additional Resources

- [Model Context Protocol specification](https://modelcontextprotocol.io)
- [Adobe Premiere Pro Scripting Guide](https://ppro-scripting.docsforadobe.dev/)
- [Adobe CEP Resources](https://github.com/nicmangroup/CEP-Resources)
- [Contributing Guide](../CONTRIBUTING.md)
- [Architecture Details](architecture.md)
- [Feature Plan](feature-plan.md)

---

*PremierPro MCP Server is open-source software released under the MIT License.*
