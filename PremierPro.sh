#!/bin/bash
# Run this to launch the PremierPro AI Editor on Linux.
# Usage: ./PremierPro.sh

cd "$(dirname "$0")"

# Load shell profile for env vars
[ -f ~/.bashrc ] && source ~/.bashrc 2>/dev/null
[ -f ~/.bash_profile ] && source ~/.bash_profile 2>/dev/null
[ -f ~/.profile ] && source ~/.profile 2>/dev/null

if [ -f .env ]; then
    set -a
    # shellcheck disable=SC1091
    . ./.env
    set +a
fi

# Abort on failed installs, generation, builds, or launches.
set -e

RUST_BACKEND_PORT="${MEDIA_ENGINE_PORT:-50052}"
PYTHON_BACKEND_PORT="${INTEL_GRPC_PORT:-50053}"
TS_BACKEND_PORT="${BRIDGE_GRPC_PORT:-50054}"

# Authentication is resolved by the CLI from ANTHROPIC_API_KEY,
# OPENAI_API_KEY, or ~/.premierpro-mcp/config.json. Claude/Codex OAuth
# sessions are not API keys and are never scraped by this launcher.

BACKEND_SUPERVISOR_PID=""
STARTED_BACKENDS=false

cleanup() {
    status=$?
    trap - EXIT INT TERM
    if [ "$STARTED_BACKENDS" = true ]; then
        echo "  Stopping backend services started by this launcher..."
        if [ -n "$BACKEND_SUPERVISOR_PID" ] && kill -0 "$BACKEND_SUPERVISOR_PID" 2>/dev/null; then
            kill -TERM "$BACKEND_SUPERVISOR_PID" 2>/dev/null || true
            wait "$BACKEND_SUPERVISOR_PID" 2>/dev/null || true
        else
            ./scripts/stop-all.sh >/dev/null 2>&1 || true
        fi
    fi
    exit "$status"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

backend_port_is_ready() {
    port="$1"
    if command -v nc >/dev/null 2>&1; then
        nc -z 127.0.0.1 "$port" >/dev/null 2>&1
        return
    fi
    if command -v lsof >/dev/null 2>&1; then
        lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
        return
    fi
    (exec 3<>"/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1
}

ready_backend_count() {
    count=0
    backend_port_is_ready "$RUST_BACKEND_PORT" && count=$((count + 1))
    backend_port_is_ready "$PYTHON_BACKEND_PORT" && count=$((count + 1))
    backend_port_is_ready "$TS_BACKEND_PORT" && count=$((count + 1))
    echo "$count"
}

wait_for_backends() {
    timeout_seconds="${PREMIERE_MCP_STARTUP_TIMEOUT:-300}"
    elapsed=0
    while [ "$elapsed" -lt "$timeout_seconds" ]; do
        if backend_port_is_ready "$RUST_BACKEND_PORT" &&
           backend_port_is_ready "$PYTHON_BACKEND_PORT" &&
           backend_port_is_ready "$TS_BACKEND_PORT"; then
            return 0
        fi
        if [ -n "$BACKEND_SUPERVISOR_PID" ] && ! kill -0 "$BACKEND_SUPERVISOR_PID" 2>/dev/null; then
            return 1
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    return 1
}

if [ ! -d "cli/node_modules" ]; then
    echo "  Installing CLI dependencies..."
    cd cli && npm ci --silent && cd ..
fi

if [ ! -d "ts-bridge/node_modules" ]; then
    echo "  Installing bridge dependencies..."
    cd ts-bridge && npm ci --silent && cd ..
fi

# Generated protobuf clients are intentionally not committed. Generate them
# before the first clean-checkout build, then always rebuild through Go's cache
# so pulling source can never leave the launcher on a stale server binary.
if [ ! -f "gen/go/premierpro/premiere/v1/premiere.pb.go" ]; then
    echo "  Generating protobuf clients..."
    ./scripts/generate-proto.sh
fi
echo "  Building MCP server..."
mkdir -p go-orchestrator/bin
cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server/ && cd ..

# Start all backends only from a clean three-port state. A partial listener set
# may belong to a failed or unrelated process and must not be silently reused.
READY_BACKENDS="$(ready_backend_count)"
if [ "$READY_BACKENDS" -eq 0 ]; then
    echo "  Starting backend services..."
    ./scripts/start-all.sh &
    BACKEND_SUPERVISOR_PID=$!
    STARTED_BACKENDS=true
    if ! wait_for_backends; then
        echo "  Backend services did not become ready. Check scripts/logs/."
        exit 1
    fi
elif [ "$READY_BACKENDS" -ne 3 ]; then
    echo "  Partial backend state detected ($READY_BACKENDS of 3 ports are listening)."
    echo "  Stop the existing services or resolve the occupied ports, then try again."
    exit 1
fi

npx --prefix cli tsx cli/src/index.ts
