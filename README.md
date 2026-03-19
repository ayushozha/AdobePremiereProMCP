# PremierPro MCP Server

An AI-powered MCP (Model Context Protocol) server that enables end-to-end video editing in Adobe Premiere Pro. Give it a script and your assets — it handles the rest.

```
"Edit this 5-minute video using script.pdf with the footage in /media/"
```

The server parses your script, scans your media library, generates an edit decision list, and assembles the timeline in Premiere Pro — all from a single prompt.

## Architecture

Four languages, each playing to their strengths:

```
CLI / MCP Client
       │ stdio / JSON-RPC
       ▼
┌─────────────────────────────────────┐
│     Go — MCP Server & Orchestrator  │
│  Protocol handling · Concurrency    │
│  Service mesh · Health & recovery   │
└──────┬────────────┬────────────┬────┘
       │ gRPC       │ gRPC       │ gRPC
       ▼            ▼            ▼
┌────────────┐ ┌──────────┐ ┌────────────────┐
│   Rust     │ │  Python  │ │  TypeScript     │
│   Media    │ │  Intel   │ │  Premiere Pro   │
│   Engine   │ │  Layer   │ │  Bridge         │
└────────────┘ └──────────┘ └───────┬────────┘
                                    │ CEP / ExtendScript
                                    ▼
                             Adobe Premiere Pro
```

| Language | Role | Why |
|---|---|---|
| **Go** | MCP server, orchestration | Goroutines for concurrency, fast startup, low memory |
| **Rust** | Media processing | Raw performance for scanning, indexing, waveform analysis |
| **Python** | AI & NLP | Script parsing, edit decisions, shot matching via embeddings |
| **TypeScript** | Premiere Pro bridge | Native access to Adobe's ExtendScript/CEP DOM |

Full architecture diagram: [`docs/architecture.md`](docs/architecture.md)

## Project Structure

```
PremierProMCP/
├── go-orchestrator/          # Go — MCP server & task orchestrator
│   ├── cmd/server/           #   Entry point
│   ├── internal/             #   Core packages
│   │   ├── mcp/              #     MCP protocol handler
│   │   ├── orchestrator/     #     Task orchestration
│   │   ├── health/           #     Health checks
│   │   └── grpc/             #     gRPC client/server
│   └── configs/              #   Configuration files
│
├── rust-engine/              # Rust — Media processing engine
│   └── src/
│       ├── media/            #   Media probe & metadata
│       ├── assets/           #   Asset indexing & fingerprinting
│       ├── waveform/         #   Waveform & silence detection
│       └── thumbnails/       #   Thumbnail generation
│
├── python-intelligence/      # Python — AI intelligence layer
│   ├── src/
│   │   ├── parser/           #   Script parsing & NLP
│   │   ├── edl/              #   Edit Decision List generation
│   │   ├── matching/         #   Shot-to-asset matching
│   │   └── analysis/         #   Pacing & timing analysis
│   ├── tests/
│   └── models/               #   ML model configs
│
├── ts-bridge/                # TypeScript — Premiere Pro bridge
│   └── src/
│       ├── extendscript/     #   ExtendScript API layer
│       ├── cep/              #   CEP Panel bridge (primary)
│       ├── standalone/       #   Node.js fallback bridge
│       └── timeline/         #   Timeline operations
│
├── cep-panel/                # CEP Panel — Premiere Pro extension
│   ├── src/
│   ├── assets/
│   └── CSXS/                 #   Adobe extension manifest
│
├── proto/                    # Shared protobuf definitions
│   └── definitions/
│
├── docs/                     # Documentation
│   └── architecture.md       #   Architecture diagram (SVG)
│
├── scripts/                  # Build & setup scripts
├── shared/                   # Shared schemas & configs
├── Justfile                  # Unified build system
└── .env.example              # Environment template
```

## Prerequisites

- [Go](https://go.dev/) 1.22+
- [Rust](https://rustup.rs/) 1.77+
- [Python](https://python.org/) 3.12+
- [Node.js](https://nodejs.org/) 20+
- [just](https://github.com/casey/just) (command runner)
- [buf](https://buf.build/) (protobuf toolchain)
- [FFmpeg](https://ffmpeg.org/) (media processing)
- Adobe Premiere Pro (2020 or later)

## Quick Start

```bash
# Clone
git clone https://github.com/ayushozha/AdobePremiereProMCP.git
cd PremierProMCP

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
```

## Usage

### As an MCP Server (stdio)

Add to your Claude Code MCP config:

```json
{
  "mcpServers": {
    "premiere-pro": {
      "command": "./go-orchestrator/bin/server",
      "args": ["--transport", "stdio"]
    }
  }
}
```

### Via CLI

```bash
# Start the server
just go-run

# Or run directly
./go-orchestrator/bin/server --transport stdio
```

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

## How It Works

1. **You send a prompt** — "Edit this video using the script with footage from /media/"
2. **Go orchestrator** receives the MCP tool call and fans out:
   - **Rust engine** scans `/media/`, indexes all assets (codec, duration, resolution, waveforms)
   - **Python intelligence** parses the script, generates an Edit Decision List, matches shots to assets
3. **Go merges results** and sends the EDL to the TypeScript bridge
4. **TypeScript bridge** executes in Premiere Pro — creates sequence, places clips, adds transitions, text
5. **Premiere Pro renders** the final output

## License

MIT
