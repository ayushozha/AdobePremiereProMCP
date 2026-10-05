# Windows MCP bridge lifecycle

This optional entrypoint owns one TypeScript bridge for one Go MCP session.
Rust (media, default port 50052) and Python (intelligence, default port 50053)
still run separately. Existing launchers and the default Go command are unchanged.

Build `ts-bridge/dist/index.js`, install Node on PATH, and build the Go server as
`go-orchestrator/bin/server.exe`. Configure Cursor to invoke that executable with
`--embed-ts-bridge` directly so the MCP host terminates the Go owner itself.
`scripts/cursor-mcp-launcher.cmd` is a manual launch convenience. Terminating only
its outer `cmd.exe` does not guarantee termination of Go; do not use that shell
as the MCP host's process ownership boundary.
Keep the binary in this location: the bridge build is resolved relative to it.

Do not also start TypeScript with `start-all.sh` or another launcher. The embedded
mode refuses occupied bridge gRPC port without stopping their owners.
`TS_BRIDGE_ADDR` must be loopback (default `localhost:50054`); the child receives
the same host/port via `BRIDGE_GRPC_HOST` and `BRIDGE_GRPC_PORT`.
The separately running CEP panel owns its WebSocket port (9801 by default);
an occupied CEP port is expected. Other bridge environment settings pass
through. Port preflight cannot reserve a port throughout startup; avoid another
launcher racing for the gRPC port.

The Node child waits on a private startup pipe until Go assigns it to a Windows
Job Object. This OS resource kills contained descendants when its final handle
closes, including when the Go owner is forcibly terminated. The handle is not
inherited. If containment fails, startup fails and the child is killed/reaped.
Go never consumes MCP input on behalf of Node; bridge output is discarded and
errors go to stderr, leaving stdout for JSON-RPC.

Startup waits at most 30 seconds for the bridge TCP listener and observes early
child exit. This confirms a listener, not a Premiere connection. EOF, Go shutdown,
and startup failure kill/reap the owned child and close its job. A bridge failure
after startup is reported by ordinary tool connection errors; there is no restart
loop. Previously running services are never adopted or killed.

Validation: real subprocess tests cover gated ES-module startup, occupied-port
refusal, early exit, listener readiness, idempotent stop/reaping, and timeout
cleanup. The Windows server and tests can be crosscompiled from macOS. These
checks do not execute Windows Job Object APIs, `.cmd`, Cursor, or Premiere.
Native Windows verification remains required: establish MCP, close its input,
then separately force-stop its Go PID and check that its Node PID and any child
processes disappear while unrelated Rust/Python/services remain running.
