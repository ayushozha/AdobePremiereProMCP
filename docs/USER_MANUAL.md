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

The server accepts tool calls from MCP-compatible AI assistants such as Claude, GPT, or Codex and translates supported workflows into Premiere Pro operations. Mutating tools should be treated as successful only when their returned state can be read back from Premiere.

The MCP implementation is open source under the MIT license and uses a local CEP extension inside Premiere Pro.

### What can it do?

The source registry contains **1,064 tool schemas**. The default `standard`
profile exposes **72 curated, readback-first tools** for project/timeline
inspection, asset analysis, editing, SRT captions, installed effects and
transitions, MOGRT titles, audio, and export. Select specialized profiles for
dialogue, captions, social reframing, transitions, effects, proxies, or
delivery.

The complete registry includes experimental and legacy commands. Tool
discovery or a matching host function is not a guarantee of runtime support on
every Premiere version. Audited routes return explicit errors when they cannot
perform and read back an operation; unaudited legacy handlers in the full
catalog are not certified. Unsafe script/shell/file capabilities require an
opt-in profile, and final compatibility must be checked in a real Premiere
session.

### Supported Premiere Pro Versions

| Version | Year | Support Level |
|---|---|---|
| 14.x | 2020 | Manifest target; live matrix pending |
| 15.x | 2021 | Manifest target; live matrix pending |
| 22.x | 2022 | Manifest target; live matrix pending |
| 23.x | 2023 | Manifest target; live matrix pending |
| 24.x | 2024 | Manifest target; live matrix pending |
| 25.x | 2025 | Primary development target; live matrix pending |
| 26.x | 2026 | Beta manifest target; live matrix pending |

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
| [Go](https://go.dev/) | 1.26.1+ | MCP server and orchestrator | [go.dev/dl](https://go.dev/dl/) |
| [Rust](https://rustup.rs/) | 1.85+ | Media processing engine | `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \| sh` |
| [Python](https://python.org/) | 3.12+ | AI intelligence layer | [python.org/downloads](https://python.org/downloads/) |
| [Node.js](https://nodejs.org/) | 20+ | TypeScript bridge and CLI | [nodejs.org](https://nodejs.org/) |
| [just](https://github.com/casey/just) | latest | Unified build system | `cargo install just` or `brew install just` |
| [buf](https://buf.build/) | latest | Protobuf code generation | [buf.build/docs/installation](https://buf.build/docs/installation) |
| [FFmpeg](https://ffmpeg.org/) | latest | Media scanning and processing | `brew install ffmpeg` or [ffmpeg.org/download](https://ffmpeg.org/download.html) |
| Adobe Premiere Pro | 2020 (v14) or later | Target application | [adobe.com](https://www.adobe.com/products/premiere.html) |

You also need an API key from one of the supported AI providers:

- **Anthropic** -- for Claude models (recommended)
- **OpenAI** -- for GPT/Codex models; set an explicit `OPENAI_API_KEY`

---

## 3. Installation

### 3.1 Clone and Build

```bash
# Clone the repository
git clone https://github.com/ayushozha/AdobePremiereProMCP.git
cd AdobePremiereProMCP

# Copy the environment template
cp .env.example .env

# Install dependencies (Python + Node.js)
just install

# Generate Go and Python protobuf stubs
just proto

# Build all components (Go, Rust, TypeScript bridge, CEP panel, CLI)
just build
```

The `just build` command runs the following in sequence:
1. `buf generate` -- generates Go and Python stubs. Rust generates its service
   code in `build.rs`; TypeScript loads the versioned `.proto` definitions
   directly with `@grpc/proto-loader`.
2. `go build` -- compiles the Go orchestrator to `go-orchestrator/bin/premierpro-mcp`
3. `cargo build --release` -- compiles the Rust media engine
4. `npm run build` -- bundles the TypeScript bridge
5. `npm run build` -- assembles the static CEP panel in `cep-panel/dist`
6. `npm run build` -- compiles the interactive CLI

To verify everything compiled correctly:

```bash
just test
```

### 3.2 Install the CEP Panel

```bash
just install-panel
```

This script builds the production panel and installs it for the current user:

1. Builds and symlinks the production `cep-panel/dist/` directory on macOS.
   The Windows installer tries a directory symlink and falls back to a checked
   file copy. The production build omits the development-only `.debug` DevTools
   endpoint:
   - macOS: `~/Library/Application Support/Adobe/CEP/extensions/com.premierpro.mcp.bridge`
   - Windows: `%APPDATA%\Adobe\CEP\extensions\com.premierpro.mcp.bridge`

2. **Enables unsigned extensions** by setting `PlayerDebugMode=1` for CSXS
   11-13, which is required for this unsigned local extension. This is a
   persistent per-user CEP developer setting; install only trusted CEP
   extensions and turn it off when you no longer need unsigned extensions.

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

### Windows Native Setup

The POSIX `just start` supervisor is not a native Windows service manager yet.
From **Command Prompt**, prepare a clean checkout with:

```bat
copy .env.example .env
buf generate
py -3.12 -m venv .venv
.venv\Scripts\python -m pip install --constraint python-intelligence\constraints.txt -e "python-intelligence[dev]"
cargo build --locked --release --manifest-path rust-engine\Cargo.toml
npm ci --prefix ts-bridge
npm run build --prefix ts-bridge
npm ci --prefix cli
scripts\install-cep-panel-win.bat
```

Then keep these three commands open in separate terminals:

```bat
rust-engine\target\release\premierpro-media-engine.exe --port 50052
```

```bat
cd python-intelligence
set PYTHONPATH=..\gen\python;.
..\.venv\Scripts\python -m src.main --port 50053
```

```bat
npm start --prefix ts-bridge
```

After the ports are listening, run `PremierPro.bat` to regenerate missing Go
protobufs, rebuild the Go MCP binary through its cache, and launch the CLI. The
batch file fails fast but does not own, monitor, or stop the three services.

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

The simplest way to get started on macOS/Unix is the platform launcher:

- **macOS:** Double-click `PremierPro.command` (or run `./PremierPro.command` in Terminal)
- **Windows:** `PremierPro.bat` builds/launches the CLI only; first start the
  Rust, Python, and TypeScript services and run
  `scripts\install-cep-panel-win.bat`
- **Linux:** Run `./PremierPro.sh`

Alternatively, run directly from the terminal:

```bash
npx --prefix cli tsx cli/src/index.ts
```

`PremierPro.command` performs the following steps automatically. The Unix
launcher performs the same backend/CLI steps without installing a local
Premiere panel. The Windows batch launcher currently performs only dependency,
Go binary, and CLI startup steps.

1. **Resolves authentication** -- checks explicit provider API keys and the
   PremierPro MCP config file. Claude/Codex subscription or OAuth sessions
   cannot be reused as API keys; set `ANTHROPIC_API_KEY` or `OPENAI_API_KEY`
   explicitly when the config file has no key.
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
      "command": "/path/to/AdobePremiereProMCP/go-orchestrator/bin/premierpro-mcp",
      "args": ["--transport", "stdio"]
    }
  }
}
```

Replace `/path/to/AdobePremiereProMCP` with the actual path to your clone of the repository.

The binary is `go-orchestrator/bin/premierpro-mcp`, built from `go-orchestrator/cmd/server/main.go` by `just go-build`.

**SSE transport** is also available if your client supports it:

```bash
./go-orchestrator/bin/premierpro-mcp --transport sse --port 8080
```

This starts an HTTP server at `http://localhost:8080` with Server-Sent Events for the MCP protocol.
It binds to `127.0.0.1` by default. SSE has no application-level authentication;
never expose it on a LAN or public interface without an authenticated TLS reverse proxy.

### 4.3 Authentication

The system supports multiple authentication methods, checked in this priority order:

| Priority | Method | How to Set Up |
|---|---|---|
| 1 | `ANTHROPIC_API_KEY` process environment | `export ANTHROPIC_API_KEY="sk-ant-..."` |
| 2 | `OPENAI_API_KEY` process environment | `export OPENAI_API_KEY="sk-..."` |
| 3 | Repository `.env` | Set either key in the root `.env`; existing process values win |
| 4 | Config file | Create `~/.premierpro-mcp/config.json` with your key |

Claude and Codex subscription/OAuth logins authenticate those applications;
they do not expose provider API keys to this SDK-based CLI. Set an explicit
provider key in the process environment, root `.env`, or config file below.

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

**Note on OAuth:** Claude/Codex application OAuth is intentionally not scraped
or reused. This CLI requires an explicit provider API key.

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

The Go layer is the entry point for the system. It implements the MCP protocol (JSON-RPC 2.0 over stdio or SSE), registers 1,064 source schemas, applies the selected tool profile, and orchestrates requests across downstream gRPC services. Discovery is cursor-paginated and defaults to the 72-tool `standard` profile.

- **Directory:** `go-orchestrator/`
- **Entry point:** `cmd/server/main.go`
- **Tool definitions:** `internal/mcp/*_tools.go` (36 files)
- **Default SSE HTTP port:** 8080 (the stdio transport does not open a port)

### Rust -- Media Processing Engine

The Rust layer handles performance-critical media operations: scanning directories for media files, extracting metadata (codec, resolution, duration, frame rate), generating waveforms for silence detection, creating thumbnails, and indexing assets. It uses FFmpeg bindings for media processing and zero-copy I/O for performance.

- **Directory:** `rust-engine/`
- **Modules:** `media/`, `assets/`, `waveform/`, `thumbnails/`
- **Default gRPC port:** 50052

### Python -- Intelligence Layer

The Python layer handles deterministic edit-intelligence tasks: parsing screenplay, YouTube, podcast, and narration formats; generating Edit Decision Lists (EDLs); matching script segments to supplied asset metadata; and analyzing pacing and timing. Optional provider SDKs are packaged as an extra, but the current service does not call an LLM as part of its normal pipeline.

- **Directory:** `python-intelligence/`
- **Modules:** `parser/`, `edl/`, `matching/`, `analysis/`
- **Default gRPC port:** 50053

### TypeScript -- Premiere Pro Bridge

The TypeScript layer is the bridge between the Go orchestrator and Adobe Premiere Pro. It translates gRPC commands into ExtendScript API calls that Premiere Pro can execute. It supports two modes:

1. **CEP Panel (primary):** Runs inside Premiere Pro as an extension panel. Has direct access to the Premiere Pro DOM. Lowest latency. Communicates with the Go orchestrator over a local WebSocket/HTTP connection.

2. **Standalone Node.js (macOS only):** Runs as an external process and sends commands to Premiere Pro through `osascript`. Higher latency, but it can work without the panel installed. Windows users must use the CEP panel bridge.

Select the bridge explicitly with `BRIDGE_MODE=cep` (the default) or, on macOS,
`BRIDGE_MODE=standalone`. The service does not automatically switch mutation
backends when the selected bridge is unavailable.

- **Directory:** `ts-bridge/`
- **Modules:** `extendscript/`, `cep/`, `standalone/`, `timeline/`
- **Default gRPC port:** 50054

### CEP Panel

The CEP panel is an Adobe Common Extensibility Platform extension that runs inside Premiere Pro. It provides a small UI panel and, more importantly, acts as the host for ExtendScript execution. The panel's WebSocket server listens for commands from the TypeScript bridge and executes them against Premiere Pro's scripting DOM.

The WebSocket listens on loopback and requires a shared authentication token.
The CEP panel and TypeScript bridge create/read
`~/.premierpro-mcp/cep-token` automatically. Keep that file private. Use
`PREMIERE_MCP_TOKEN_FILE` to select another shared path, or set the same
`BRIDGE_CEP_TOKEN`/`MCP_CEP_TOKEN` value in both processes.

- **Directory:** `cep-panel/`
- **Panel menu name:** PremierPro MCP Bridge
- **Extension ID:** `com.premierpro.mcp.bridge`
- **CSXS version:** 11.0

### Tool Profiles and Workflow Skills

By default the server exposes the 72-tool `standard` catalog. Set
`MCP_TOOL_PROFILE` to select another workflow-sized surface:

| Profile | Focus |
|---|---|
| `core` | Host, project, timeline, duplicate-sequence recovery, audit snapshots, and verified save operations |
| `dialogue` | Rust waveform/silence analysis, reviewable trims, gaps, levels, and crossfades |
| `captions` | Verified SRT import/readback, structural validation, and export |
| `social` | Vertical/square reframing, safe-zone review, and explicit-preset export |
| `transitions` | Video transitions and audio crossfades |
| `effects` | Installed-effect discovery, verified attachment, parameter updates, and readback |
| `proxies` | Proxy creation, attachment, status, and conform |
| `delivery` | Direct/AME export with explicit presets plus external media verification |
| `standard` | Curated, readback-first editing set (default; 72 tools) |
| `all` | Complete catalog except unsafe arbitrary execution/file tools |
| `unsafe` | Explicit opt-in for arbitrary scripts, shell commands, URLs, clipboard, external editors, and file I/O |

Profiles may be combined with commas, such as `captions,effects`; every
specialized selection also includes `core`. Use `all,unsafe` only with trusted
inputs and human review. The repository includes matching
skills under `skills/premiere-dialogue-cut`, `skills/premiere-captions`,
`skills/premiere-social-reframe`, `skills/premiere-transition-pack`,
`skills/premiere-look-effects`, `skills/premiere-proxy-conform`, and
`skills/premiere-batch-delivery`. MCP clients can also read their summary from
`config://workflow-skills`.

Tool discovery is cursor-paginated. The default `MCP_PAGE_SIZE` is 100; clients
must follow `nextCursor` until it is absent rather than assuming the first page
is the complete catalog.

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
  go version        # Should be 1.26.1+
  rustc --version   # Should be 1.85+
  python3 --version # Should be 3.12+
  node --version    # Should be 20+
  ffmpeg -version   # Should be installed
  ```

### WebSocket not connecting

The TypeScript bridge communicates with the CEP panel over a local WebSocket connection.

- **Make sure the CEP panel is open** in Premiere Pro (Window > Extensions > PremierPro MCP Bridge). The panel must be visible for its WebSocket server to be active.
- **Make sure both processes use the same token.** Normally they share `~/.premierpro-mcp/cep-token`; if you set `BRIDGE_CEP_TOKEN`, `MCP_CEP_TOKEN`, or `PREMIERE_MCP_TOKEN_FILE`, apply the same configuration to the panel and bridge.
- **Check the bridge log** for connection errors:
  ```bash
  cat scripts/logs/ts-bridge.log
  ```
- On macOS, if using the standalone (non-CEP) bridge mode, set `BRIDGE_MODE=standalone` in your environment.

If the orchestrator runs in Docker, keep the TypeScript bridge on the host.
The CEP WebSocket is intentionally loopback-only, so it is not directly
reachable from a container. Point `TS_BRIDGE_ADDR` in the container at the
host bridge's gRPC endpoint, such as `host.docker.internal:50054` where that
hostname is available.

### Python service fails to start

- **Install Python dependencies manually:**
  ```bash
  python3 -m pip install \
    --constraint python-intelligence/constraints.txt \
    -e "./python-intelligence[dev]"
  ```
  Do not install an ad-hoc protobuf version over this environment; generated
  clients and runtime packages are pinned together by `constraints.txt`.
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
- Or add the key to the root `.env`:
  ```bash
  ANTHROPIC_API_KEY="sk-ant-..."
  ```
- Claude/Codex logins cannot supply provider API keys. OpenAI users must set
  `OPENAI_API_KEY` explicitly or add `openai_api_key` to the config file.
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
| Add a title | `premiere_import_mogrt` + `premiere_set_mogrt_text` | "Place this title MOGRT at 5 seconds and set its text to Hello World" |
| Set audio level | `premiere_set_audio_level` | "Set the audio on this clip to -6 dB" |
| Color grading | `premiere_lumetri_set_*` | "Increase the contrast to 30" |
| Apply an effect | `premiere_apply_effect` | "Apply Gaussian Blur to this clip" |
| Scan media assets | `premiere_scan_assets` | "Scan /Users/me/footage/ for media files" |
| Parse a script | `premiere_parse_script` | "Parse the script at /Users/me/script.pdf" |
| Auto-edit from script | `premiere_auto_edit` | "Edit using script.pdf with footage from /media/" |
| Export video | `premiere_export` | "Export sequence seq-abc with my configured H.264 1080p `.epr` alias to /Users/me/output.mp4" |
| Close Premiere Pro | `premiere_close` | "Close Premiere Pro" |

---

## 8. Build Commands

All build commands use `just` as the unified build system. Run `just` with no arguments to see the full list.

### Top-Level Commands

| Command | Description |
|---|---|
| `just build` | Build all components (proto, Go, Rust, TypeScript, CEP, CLI) |
| `just test` | Run all test suites |
| `just lint` | Lint all code (Go, Rust, Python, TypeScript, proto) |
| `just ci` | Full CI pipeline: lint, build, then test |
| `just clean` | Remove all build artifacts |
| `just install` | Install Python, bridge, CEP panel, and CLI dependencies |

### Protobuf

| Command | Description |
|---|---|
| `just proto` | Generate Go/Python protobuf stubs |
| `just proto-lint` | Lint protobuf definitions |

### Go Orchestrator

| Command | Description |
|---|---|
| `just go-build` | Build the Go orchestrator to `go-orchestrator/bin/premierpro-mcp` |
| `just go-run` | Run the Go orchestrator directly |
| `just go-test` | Run Go tests |
| `just go-lint` | Run `go vet` over all Go packages |

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

### CLI

| Command | Description |
|---|---|
| `just cli-install` | Install CLI dependencies |
| `just cli-build` | Build the interactive CLI |

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
| `MCP_SSE_HOST` | `127.0.0.1` | SSE bind host; keep loopback unless protected by authenticated TLS |
| `MCP_SSE_PORT` | `8080` | Port for the SSE HTTP server (only used with `sse` transport) |
| `MCP_LOG_LEVEL` | `info` | Log level for the Go orchestrator: `debug`, `info`, `warn`, `error` |
| `MCP_PAGE_SIZE` | `100` | Maximum tools returned per cursor-paginated `tools/list` page |
| `MCP_TOOL_PROFILE` | `standard` | Comma-separated profiles; `all` is safe-only and `unsafe` explicitly enables arbitrary script/shell/file tools |

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
| `MEDIA_ENGINE_HOST` | `127.0.0.1` | Bind interface for the Rust gRPC service |
| `MEDIA_ENGINE_PORT` | `50052` | gRPC port for the Rust media engine |
| `MEDIA_ENGINE_LOG_LEVEL` | `info` | Rust engine log level |
| `MEDIA_ENGINE_LOG_JSON` | `false` | Emit structured JSON logs when `true` |

### Python Intelligence

| Variable | Default | Description |
|---|---|---|
| `PYTHON_BIN` | `python3` | Python 3.12+ executable used by local start scripts |
| `INTEL_GRPC_HOST` | `127.0.0.1` | Bind interface for the Python gRPC service |
| `INTEL_GRPC_PORT` | `50053` | gRPC port for the Python intelligence service |
| `INTEL_LOG_LEVEL` | `INFO` | Python intelligence log level |
| `INTEL_OPENAI_API_KEY` | (none) | Optional OpenAI key used by the intelligence service |
| `INTEL_ANTHROPIC_API_KEY` | (none) | Optional Anthropic key used by the intelligence service |

### TypeScript Bridge

| Variable | Default | Description |
|---|---|---|
| `BRIDGE_GRPC_HOST` | `127.0.0.1` | Interface used by the TypeScript gRPC server |
| `BRIDGE_GRPC_PORT` | `50054` | gRPC port for the TypeScript bridge |
| `PREMIERE_PATH` | `/Applications/Adobe Premiere Pro 2025/Adobe Premiere Pro 2025.app` | Path to the Premiere Pro application |
| `BRIDGE_MODE` | `cep` | Bridge implementation: `cep`, or macOS-only `standalone` |
| `BRIDGE_LOG_LEVEL` | `info` | TypeScript bridge log level |
| `BRIDGE_CEP_WS_PORT` | `9801` | Local CEP panel WebSocket port |
| `BRIDGE_CEP_TOKEN` | (generated) | Explicit shared bridge token; must be at least 32 characters |
| `MCP_CEP_TOKEN` | (generated) | Alternate name for the explicit shared bridge token |
| `PREMIERE_MCP_TOKEN_FILE` | `~/.premierpro-mcp/cep-token` | Private token-file path shared by Go, TypeScript, and the CEP panel |

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
|   |   +-- mcp/              #     MCP protocol handler (1,064 registered schemas)
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
|       +-- standalone/       #   macOS-only Node.js bridge
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
