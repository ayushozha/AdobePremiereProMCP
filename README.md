# PremierPro MCP Server -- AI-Powered Video Editing for Adobe Premiere Pro

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/ayushozha/AdobePremiereProMCP/pulls)
[![Premiere Pro 2020-2026](https://img.shields.io/badge/Premiere%20Pro-2020--2026-9999FF.svg)](https://www.adobe.com/products/premiere.html)
[![MCP Protocol](https://img.shields.io/badge/MCP-Model%20Context%20Protocol-blue.svg)](https://modelcontextprotocol.io)
[![GitHub stars](https://img.shields.io/github/stars/ayushozha/AdobePremiereProMCP?style=social)](https://github.com/ayushozha/AdobePremiereProMCP/stargazers)

**An open-source MCP server for Adobe Premiere Pro.** Automate supported timeline, media, audio, effect, caption, graphics-template, and export workflows through Claude, GPT, or another [Model Context Protocol](https://modelcontextprotocol.io) client.

> Give it a script and your footage, review the proposed edit, then let it assemble a recoverable rough cut on a duplicated sequence.

```
"Edit this 5-minute video using script.pdf with the footage in /media/"
```

The server can parse a script, scan media, generate an edit decision list, and assemble a Premiere Pro timeline. Mutations are checked with timeline/DOM readback where the host API exposes it.

---

## Why This Exists

Video editors spend hours on repetitive tasks: syncing clips, rough cuts, color matching, audio leveling, exporting variants. This MCP server turns Adobe Premiere Pro into an AI-controllable tool, so you can describe edits in plain English and let your AI assistant execute them.

**No separate MCP subscription. Local CEP bridge. Fully open source.**

## Capabilities and Verification Boundary

The source registry contains **1,064 MCP tool schemas** backed by **932 host command names**. The default `standard` profile exposes **72 curated, readback-first tools**, which fits common model function limits and covers normal editing work. The full registry remains available for compatibility and specialist use; schema or symbol presence is not proof that a command works in every Premiere version.

The curated surface includes:

- Project/timeline inspection, sequence versioning, import, placement, trim, ripple, markers, and export.
- Rust-backed media probe, thumbnail, waveform, silence, scene, and asset-scan operations.
- Script parsing, asset matching, EDL generation, and pacing analysis.
- Caption-track creation from validated SRT with count readback.
- Installed-effect and transition discovery, application, and public-DOM readback.
- MOGRT-based titles, social derivatives, proxy workflows, and batch-delivery recipes.

The audited caption, transition, effect, and graphics-template routes return
explicit `unsupported` errors when Premiere cannot perform and read back the
requested mutation. The broader compatibility catalog still contains legacy
and experimental handlers that have not completed the same audit, so it must
not be treated as a certified surface. Arbitrary script, shell, URL, clipboard,
external-editor, and file-I/O tools are hidden unless `unsafe` is selected. A
real Premiere session is still required to certify host mutations on a
particular OS/Premiere build.

## Supported Premiere Pro Versions

| Version | Year | Support |
|---|---|---|
| 14.x | 2020 | Manifest target; live matrix pending |
| 15.x | 2021 | Manifest target; live matrix pending |
| 22.x | 2022 | Manifest target; live matrix pending |
| 23.x | 2023 | Manifest target; live matrix pending |
| 24.x | 2024 | Manifest target; live matrix pending |
| 25.x | 2025 | Primary development target; live matrix pending |
| 26.x | 2026 | Beta manifest target; live matrix pending |

The CEP bridge targets **macOS** and **Windows** and declares Premiere Pro 14.0+ in its extension manifest.

The manifest range is not a live compatibility certification. Run the host smoke workflow against your exact Premiere build before trusting project mutations.

Help us expand compatibility -- [report your setup](https://github.com/ayushozha/AdobePremiereProMCP/issues/4).

## Architecture

Four languages, each playing to their strengths:

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

| Language | Role | Why |
|---|---|---|
| **Go** | MCP server, orchestration | Goroutines for concurrency, fast startup, low memory |
| **Rust** | Media processing | Raw performance for scanning, indexing, waveform analysis |
| **Python** | Edit intelligence | Deterministic script parsing, asset matching, EDL generation, pacing analysis |
| **TypeScript** | Premiere Pro bridge | Native access to Adobe's ExtendScript/CEP DOM |

Full architecture diagram: [`docs/architecture.md`](docs/architecture.md)

## Project Structure

```
PremierProMCP/
+-- go-orchestrator/          # Go -- MCP server & task orchestrator
|   +-- cmd/server/           #   Entry point
|   +-- internal/             #   Core packages
|   |   +-- mcp/              #     MCP protocol handler (1,064 registered schemas)
|   |   +-- orchestrator/     #     Task orchestration
|   |   +-- health/           #     Health checks
|   |   +-- grpc/             #     gRPC client/server
|   +-- configs/              #   Configuration files
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
|   +-- src/
|   +-- assets/
|   +-- CSXS/                 #   Adobe extension manifest
|
+-- proto/                    # Shared protobuf definitions
+-- docs/                     # Documentation
+-- scripts/                  # Build & setup scripts
+-- Justfile                  # Unified build system
+-- .env.example              # Environment template
```

## Prerequisites

- [Go](https://go.dev/) 1.26.1+
- [Rust](https://rustup.rs/) 1.85+
- [Python](https://python.org/) 3.12+
- [Node.js](https://nodejs.org/) 20+
- [just](https://github.com/casey/just) (command runner)
- [buf](https://buf.build/) (protobuf toolchain)
- `rsync` (generated-client synchronization on macOS/Linux)
- [FFmpeg](https://ffmpeg.org/) (media processing)
- Adobe Premiere Pro (2020 or later)

## Quick Start

```bash
# Clone the repository
git clone https://github.com/ayushozha/AdobePremiereProMCP.git
cd AdobePremiereProMCP

# Copy env template
cp .env.example .env

# Install dependencies
just install

# Generate protobuf stubs
just proto

# Build all components
just build

# Run tests
just test

# Install the CEP panel into Premiere Pro
just install-panel
```

Those commands use a POSIX shell. Native Windows users should follow the
platform-specific build/start commands in
[`docs/USER_MANUAL.md`](docs/USER_MANUAL.md#windows-native-setup); the current
`PremierPro.bat` is a fail-fast CLI launcher, not a service supervisor.

## Usage

### As an MCP Server (Claude Code, Claude Desktop, Cursor, etc.)

Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "premiere-pro": {
      "command": "./go-orchestrator/bin/premierpro-mcp",
      "args": ["--transport", "stdio"]
    }
  }
}
```

The complete tool catalog is returned in cursor-paginated pages of 100 by
default. MCP clients should continue requesting `tools/list` with the returned
cursor until `nextCursor` is absent. Set `MCP_PAGE_SIZE` to tune the page size.

The default `standard` profile keeps the normal editing surface small enough
for model tool limits. Set `MCP_TOOL_PROFILE` to `core`, `dialogue`, `captions`,
`social`, `transitions`, `effects`, `proxies`, `delivery`, `standard`, or a
comma-separated combination. Specialized profiles always include core
inspection and duplicate-sequence recovery tools. `all` exposes the safe full catalog; arbitrary
script, shell, URL, and file-I/O tools additionally require `unsafe` (for
example `MCP_TOOL_PROFILE=all,unsafe`) and should only be used with trusted
prompts and explicit human review.

### Workflow Skills

Seven reusable Agent Skills live in [`skills/`](skills/): dialogue cutting,
captions, social reframing, transition recipes, look/effect chains, proxy
conform, and batch delivery. Mutation skills preserve an untouched duplicate
sequence (or a tool-created derivative) as the recovery boundary and require
readback verification. Timeline snapshots and saved sequence versions are
audit/comparison records, not whole-sequence rollback points. The same catalog is available to MCP clients through
the `config://workflow-skills` resource.

The sources evaluated, patterns adopted, and integrations intentionally left
out are recorded in [`docs/workflow-research.md`](docs/workflow-research.md).
The exact test matrix and remaining host boundaries are recorded in
[`docs/repository-audit-2026-08-05.md`](docs/repository-audit-2026-08-05.md).

### Via CLI

```bash
# Start the server
just go-run

# Or run directly
./go-orchestrator/bin/premierpro-mcp --transport stdio
```

### Platform Launchers

Platform-specific CLI launchers are included:

- **macOS:** `./PremierPro.command` starts the three backends, installs the CEP
  panel if needed, waits for readiness, and launches the CLI.
- **Linux/Unix:** `./PremierPro.sh` starts the three backends, waits for
  readiness, and launches the CLI. Premiere/CEP itself must run on macOS or
  Windows.
- **Windows:** `PremierPro.bat` prepares the Node/Go CLI and launches it. Start
  the Rust, Python, and TypeScript services separately and install the panel
  with `scripts\install-cep-panel-win.bat`; full Windows supervision is not yet
  implemented by the batch launcher.

### CEP Bridge Security and Docker

The CEP panel accepts authenticated WebSocket connections only on loopback.
On first start, the Go orchestrator, TypeScript bridge, or panel creates a
shared token at `~/.premierpro-mcp/cep-token`. It authenticates both the
Go-to-TypeScript gRPC hop and the TypeScript-to-CEP WebSocket hop. To override
it, set `BRIDGE_CEP_TOKEN` or `MCP_CEP_TOKEN` consistently, or point all three
processes at the same file with `PREMIERE_MCP_TOKEN_FILE`.

Because Premiere Pro and its CEP panel run on the host, do not run the
TypeScript bridge solely inside Docker: a container cannot reach the panel's
loopback-only WebSocket. The supported live-editing layout keeps the Go
orchestrator and TypeScript bridge on the host; Docker Compose runs only the
Rust and Python analysis backends.

Before `docker compose up`, set `PREMIERE_MEDIA_ROOT` to the narrowest absolute
directory containing the input media and `PREMIERE_OUTPUT_ROOT` to an absolute
writable output directory. Compose mounts each directory at the identical
container path, so file paths sent over gRPC remain meaningful on both sides.
The input mount is read-only. Service-specific Docker targets avoid building a
duplicate all-language runtime for each analysis backend.

The SSE transport also binds to `127.0.0.1` by default. It has no application-
level authentication, so do not set `MCP_SSE_HOST` (or `--host`) to a public or
LAN interface unless an authenticated TLS reverse proxy protects it.

## How It Works

1. **You send a prompt** -- "Edit this video using the script with footage from /media/"
2. **Go orchestrator** receives the MCP tool call and fans out:
   - **Rust engine** scans `/media/`, indexes all assets (codec, duration, resolution, waveforms)
   - **Python intelligence** parses the script, generates an Edit Decision List, matches shots to assets
3. **Go merges results** and sends the EDL to the TypeScript bridge
4. **TypeScript bridge** executes in Premiere Pro -- creates the sequence,
   places clips, and applies supported transitions and graphics templates
5. **Premiere Pro renders** the final output

## Build Commands

| Command | Description |
|---|---|
| `just build` | Build all components |
| `just test` | Run all test suites |
| `just lint` | Lint all code |
| `just ci` | Full CI pipeline (lint + build + test) |
| `just proto` | Generate protobuf stubs |
| `just clean` | Remove all build artifacts |
| `just go-build` | Build Go orchestrator only |
| `just rust-build` | Build Rust engine only |
| `just py-test` | Run Python tests only |
| `just ts-build` | Build TypeScript bridge only |
| `just cep-build` | Build CEP panel only |
| `just install-panel` | Install CEP panel into Premiere Pro |
| `just start` | Start all backend services |
| `just stop` | Stop all backend services |
| `just status` | Check service status |

## Use Cases

- **Automated rough cuts** -- Parse a script and assemble a timeline from raw footage
- **Batch color grading** -- Apply Lumetri Color adjustments across clips via natural language
- **Audio post-production** -- Set levels, apply effects, and mix tracks through AI prompts
- **Template-based editing** -- Generate videos from MOGRTs and data using AI
- **Multi-format export** -- Queue multiple export presets from a single command
- **Review workflows** -- Add markers, comments, and metadata programmatically
- **AI-assisted editing** -- Let Claude or GPT analyze your footage and suggest edits

## Community

We are actively looking for testers and contributors!

- **Test the server** with your Premiere Pro setup and [report results](https://github.com/ayushozha/AdobePremiereProMCP/issues/1)
- **Request features** you need for your workflow in [the feature tracker](https://github.com/ayushozha/AdobePremiereProMCP/issues/2)
- **Report bugs** with reproduction steps in [the bug tracker](https://github.com/ayushozha/AdobePremiereProMCP/issues/3)
- **Confirm your Premiere Pro version** works in [the compatibility tracker](https://github.com/ayushozha/AdobePremiereProMCP/issues/4)
- **Start or join a discussion** in [GitHub Discussions](https://github.com/ayushozha/AdobePremiereProMCP/discussions)
- **Read the [Contributing Guide](CONTRIBUTING.md)** to get started with development

If this project is useful to you, please **star the repository** to help others find it.

## Related

- [Model Context Protocol](https://modelcontextprotocol.io) -- The open protocol for AI tool use
- [Adobe Premiere Pro Scripting Guide](https://ppro-scripting.docsforadobe.dev/) -- ExtendScript API reference
- [Adobe CEP Resources](https://github.com/nicmangroup/CEP-Resources) -- CEP panel development

## License

MIT
