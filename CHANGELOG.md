# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-03-19

### Added
- Initial release with 1,060 MCP tools
- Go orchestrator with MCP server (stdio + SSE transport)
- Rust media engine (scanning, probing, waveform, thumbnails, scene detection)
- Python intelligence (script parsing, EDL generation, asset matching, pacing)
- TypeScript bridge to Premiere Pro (CEP panel + standalone)
- Interactive CLI with Claude and OpenAI support
- Cross-platform launchers (macOS, Windows, Linux)
- MCP Resources (4) and MCP Prompts (5)
- Frame capture as base64 for visual AI feedback
- Secure ExtendScript execution with blocklist
- Professional CEP panel with dark/light mode
- Full Lumetri Color control (30 tools)
- Audio mixing and effects (62 tools)
- Effects, transitions, and keyframing (66 tools)
- Batch operations (30 tools)
- AI-powered editing pipeline (script → EDL → auto-edit)
- Adobe Stock search integration
- Media Browser filesystem access
- Project analytics and reporting
- Competitive analysis (beats all 7 competitors)

### Architecture
- Go: MCP server, health checks, circuit breaker, gRPC orchestration
- Rust: Media processing, FFmpeg integration, SHA-256 fingerprinting
- Python: NLP, embeddings, AI decision making
- TypeScript: CEP/ExtendScript bridge, WebSocket, gRPC server
