# Contributing to PremierPro MCP Server

Thank you for your interest in contributing to the Adobe Premiere Pro MCP server! This project enables AI-powered video editing through the Model Context Protocol, and we welcome contributions of all kinds.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment Setup](#development-environment-setup)
- [Architecture Overview](#architecture-overview)
- [Adding New MCP Tools](#adding-new-mcp-tools)
- [Testing](#testing)
- [Code Style Guidelines](#code-style-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/AdobePremiereProMCP.git
   cd PremierProMCP
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Environment Setup

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| [Go](https://go.dev/) | 1.22+ | MCP server & orchestrator |
| [Rust](https://rustup.rs/) | 1.77+ | Media processing engine |
| [Python](https://python.org/) | 3.12+ | AI intelligence layer |
| [Node.js](https://nodejs.org/) | 20+ | Premiere Pro bridge |
| [just](https://github.com/casey/just) | latest | Build system |
| [buf](https://buf.build/) | latest | Protobuf toolchain |
| [FFmpeg](https://ffmpeg.org/) | latest | Media processing |
| Adobe Premiere Pro | 2020+ | Target application |

### Install & Build

```bash
# Copy environment config
cp .env.example .env

# Install dependencies (Python + Node.js)
just install

# Generate protobuf stubs
just proto

# Build all components
just build

# Run all tests
just test
```

### Running the Server

```bash
# Start the MCP server
just go-run

# Or run directly
./go-orchestrator/bin/server --transport stdio
```

### Install CEP Panel

```bash
# Install the extension into Premiere Pro
just install-panel
```

## Architecture Overview

The project uses four languages, each responsible for a specific domain:

| Component | Language | Directory | Responsibility |
|---|---|---|---|
| MCP Server | Go | `go-orchestrator/` | Protocol handling, orchestration, concurrency |
| Media Engine | Rust | `rust-engine/` | File scanning, indexing, waveform analysis |
| Intelligence | Python | `python-intelligence/` | Script parsing, AI edit decisions, shot matching |
| Premiere Bridge | TypeScript | `ts-bridge/` | ExtendScript/CEP communication with Premiere Pro |
| CEP Panel | HTML/JS | `cep-panel/` | Adobe extension running inside Premiere Pro |

Communication between services uses **gRPC** with protobuf definitions in `proto/definitions/`.

## Adding New MCP Tools

MCP tools are defined in the Go orchestrator. Each tool category has its own file:

```
go-orchestrator/internal/mcp/
  audio_tools.go        # Audio mixing tools
  clip_tools.go         # Clip operations
  color_tools.go        # Color grading (Lumetri)
  effects_tools.go      # Video effects
  export_tools.go       # Export & encoding
  graphics_tools.go     # Graphics & titles
  project_tools.go      # Project management
  sequence_tools.go     # Sequence & timeline
  workspace_tools.go    # Workspace & panels
  tools.go              # Tool registration
```

### Steps to Add a New Tool

1. **Choose the right file** based on the tool's category
2. **Define the tool** with name, description, and parameters:
   ```go
   {
       Name:        "premiere_your_tool",
       Description: "Brief description of what the tool does",
       InputSchema: map[string]interface{}{
           "type": "object",
           "properties": map[string]interface{}{
               "param_name": map[string]interface{}{
                   "type":        "string",
                   "description": "What this parameter does",
               },
           },
           "required": []string{"param_name"},
       },
   }
   ```
3. **Add the handler** in the tool's handler function
4. **Add protobuf messages** in `proto/definitions/` if the tool requires bridge communication
5. **Implement the bridge side** in `ts-bridge/src/` if the tool needs to execute ExtendScript
6. **Add tests** for both the Go handler and the TypeScript bridge
7. **Register the tool** in `tools.go`

### Naming Conventions

- Tool names: `premiere_` prefix + snake_case (e.g., `premiere_set_audio_level`)
- Go functions: PascalCase (e.g., `HandleSetAudioLevel`)
- TypeScript functions: camelCase (e.g., `setAudioLevel`)
- Proto messages: PascalCase (e.g., `SetAudioLevelRequest`)

## Testing

### Run All Tests

```bash
just test
```

### Run Tests by Component

```bash
just go-test       # Go orchestrator tests
just rust-test     # Rust engine tests
just py-test       # Python intelligence tests
just ts-test       # TypeScript bridge tests
```

### Linting

```bash
just lint          # Lint all code
just go-lint       # Go only
just rust-lint     # Rust only
just py-lint       # Python only
just ts-lint       # TypeScript only
```

### Integration Testing

Integration tests require a running instance of Premiere Pro with the CEP panel installed:

```bash
just install-panel
# Open Premiere Pro manually
just test
```

## Code Style Guidelines

### Go
- Follow standard Go conventions (`gofmt`, `golangci-lint`)
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly (no silent swallowing)

### Rust
- Follow `rustfmt` and `clippy` conventions
- Use `Result<T, E>` for fallible operations
- Document public APIs with `///` doc comments

### Python
- Follow PEP 8, enforced via `ruff`
- Type hints on all function signatures (checked by `mypy`)
- Docstrings on all public functions

### TypeScript
- Follow the project ESLint config
- Use strict TypeScript (no `any` where avoidable)
- Async/await over raw promises

### General
- Keep functions small and focused
- Write descriptive commit messages
- Add tests for new functionality

## Pull Request Process

1. **Ensure tests pass:** Run `just ci` before submitting
2. **Update documentation:** If you added a new tool, update the feature plan in `docs/feature-plan.md`
3. **Keep PRs focused:** One feature or fix per PR
4. **Write a clear description:** Explain what changed and why
5. **Link issues:** Reference related issues (e.g., "Closes #12")
6. **Request review:** Tag a maintainer for review

### PR Checklist

- [ ] Code compiles and tests pass (`just ci`)
- [ ] New tools follow naming conventions
- [ ] Protobuf definitions are updated (if applicable)
- [ ] Documentation is updated
- [ ] No secrets or credentials committed

## Reporting Issues

- **Bugs:** Use the [bug report template](https://github.com/ayushozha/AdobePremiereProMCP/issues/new?template=bug_report.md)
- **Features:** Use the [feature request template](https://github.com/ayushozha/AdobePremiereProMCP/issues/new?template=feature_request.md)
- **Questions:** Start a [Discussion](https://github.com/ayushozha/AdobePremiereProMCP/discussions)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
