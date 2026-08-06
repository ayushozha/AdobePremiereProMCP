# Repository and Runtime Audit — 2026-08-05

## Outcome

The Go MCP server, Rust media backend, Python intelligence backend,
TypeScript bridge, CLI, and CEP host source all build or test successfully in
the macOS development environment used for this audit. The Go server was run
over stdio, and the real Rust/Python gRPC services were exercised with their
defined RPCs.

This is **not** a claim that every one of the 1,064 schemas works in every
Premiere version. Static coverage confirms 932 referenced host command symbols
exist, while live host mutation still requires a disposable Premiere project,
the installed CEP panel, and an OS/version-specific test matrix.

## Checked Surfaces

| Surface | Result |
|---|---|
| Go | Full tests and `go vet` pass |
| MCP stdio | `standard`: 72 tools, 5 resources, 4 prompts; `all,unsafe`: 1,064 tools, 5 resources, 5 prompts; pagination and schema argument rejection pass |
| Rust | 31 unit, 5 integration, and 2 doc tests pass; locked clippy passes |
| Python | 89 tests pass; Ruff and strict mypy (27 source files) pass |
| TypeScript bridge | 9 tests, lint, and build pass |
| CLI | 4 tests, typecheck, and build pass |
| CEP host | 932/932 Go command names have host symbols; dispatcher, EDL, sequence rollback, social clone, effect/transition readback, effect rollback, caption import, and explicit-unsupported fixtures pass |
| Agent Skills | All seven `SKILL.md` packages pass the official skill validator; every referenced MCP tool exists |
| Dependency audit | npm audit reports zero known vulnerabilities in the three Node packages at audit time |

## Important Boundaries

- The default `standard` profile is curated and readback-first. It is not yet
  live-certified against a checked-in Premiere/OS compatibility matrix.
- The full catalog contains legacy and experimental handlers. Some are useful
  inspection/planning commands; others still require individual mutation and
  readback audits. Use it only with explicit review.
- A set of legacy caption, graphics, transition, and effect handlers that
  previously returned planning metadata as success now returns explicit
  `unsupported` errors. Timed SRT import, generic installed-effect application,
  and generic installed-transition application have focused readback fixtures.
- Premiere Speech to Text, automatic translation/diarization, FCC compliance
  certification, arbitrary Essential Graphics construction, and GPU shader
  injection are not implemented.
- Live CEP/Premiere execution was not performed because installing the unsigned
  development panel changes the user's CEP extension path and debug setting.
  That final sweep needs explicit permission and a disposable project.
- Docker was not runtime-tested because Docker was unavailable in the audit
  environment. The compose layout intentionally runs only Rust/Python; the Go
  and TypeScript processes stay on the host for loopback CEP access.
- Windows service supervision was not runtime-tested. `PremierPro.bat` now
  fails fast, generates clean-clone protobuf prerequisites, builds the CLI/Go
  pieces, and launches the CLI, but the three services and CEP panel must be
  started/installed separately.

## High-Impact Fixes Included

- Cursor-paginated MCP discovery, schema-required-argument enforcement, compact
  tool profiles, profile-aware prompts, workflow resources, and protocol smoke
  tests.
- Authenticated loopback Go → TypeScript → CEP transport, origin/payload checks,
  and a production CEP build without the development `.debug` file.
- Marked-source EDL overwrite with preflight, source-mark restoration, partial
  failure reporting, state readback, and rollback for invalid sequence/effect
  mutations.
- Deterministic startup readiness, PID identity checks, safer stop/status
  scripts, pinned protobuf generators, locked Rust/Python dependencies, and
  reproducible CI commands.
- Clean-clone generation now emits only the Go/Python clients actually used;
  the TypeScript bridge loads the versioned proto definitions directly. CI
  covers Linux, Windows builds, service-specific Docker images, and release
  tags are gated on the reusable validation workflow.
- Removed 4,511 ignored `cli/node_modules` files and compiled CLI output from
  Git tracking; lockfiles and reproducible build steps are the source of truth.
- Seven workflow skills for dialogue cuts, captions, social reframing,
  transitions, effects/looks, proxy conform, and batch delivery.

See [Workflow Research and Adoption Notes](workflow-research.md) for the
external projects evaluated and the integrations deliberately left separate.
