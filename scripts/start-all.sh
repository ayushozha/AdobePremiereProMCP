#!/bin/bash
# Start all PremierPro MCP backend services

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_FILE="$SCRIPT_DIR/.pids"
LOG_DIR="$SCRIPT_DIR/logs"

if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$PROJECT_ROOT/.env"
    set +a
fi

RUST_PORT="${MEDIA_ENGINE_PORT:-50052}"
PYTHON_PORT="${INTEL_GRPC_PORT:-50053}"
TS_PORT="${BRIDGE_GRPC_PORT:-50054}"
PYTHON_BIN="${PYTHON_BIN:-python3}"
SERVICE_START_TIMEOUT="${SERVICE_START_TIMEOUT:-30}"

process_start_stamp() {
    ps -p "$1" -o lstart= 2>/dev/null | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
}

record_process() {
    local name="$1"
    local pid="$2"
    local port="$3"
    local start_stamp=""
    local attempt=0

    while [ "$attempt" -lt 20 ]; do
        start_stamp="$(process_start_stamp "$pid")"
        if [ -n "$start_stamp" ]; then
            echo "$name:$pid:$port:$start_stamp" >> "$PID_FILE"
            return 0
        fi
        if ! kill -0 "$pid" 2>/dev/null; then
            break
        fi
        sleep 0.05
        attempt=$((attempt + 1))
    done

    echo -e "${RED}Could not record a safe identity for $name (PID $pid).${NC}" >&2
    # This PID came directly from the spawn above, so it is safe to stop even
    # though ps did not provide the identity stamp needed for later reuse.
    kill -TERM "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
    "$SCRIPT_DIR/stop-all.sh" >/dev/null 2>&1 || true
    exit 1
}

process_matches() {
    local pid="$1"
    local expected_start="$2"

    kill -0 "$pid" 2>/dev/null &&
        [ -n "$expected_start" ] &&
        [ "$(process_start_stamp "$pid")" = "$expected_start" ]
}

tcp_port_open() {
    local port="$1"

    if command -v nc >/dev/null 2>&1; then
        nc -z 127.0.0.1 "$port" >/dev/null 2>&1
        return
    fi

    if command -v lsof >/dev/null 2>&1; then
        lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
        return
    fi

    # Bash provides /dev/tcp even on systems without nc or lsof.
    (exec 3<>"/dev/tcp/127.0.0.1/$port") >/dev/null 2>&1
}

wait_for_service() {
    local pid="$1"
    local port="$2"
    local expected_start="$3"
    local deadline=$((SECONDS + SERVICE_START_TIMEOUT))

    while [ "$SECONDS" -lt "$deadline" ]; do
        if ! process_matches "$pid" "$expected_start"; then
            return 1
        fi
        if tcp_port_open "$port"; then
            return 0
        fi
        sleep 0.25
    done

    return 1
}

validate_port() {
    local name="$1"
    local port="$2"

    if ! [[ "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        echo "Invalid $name port: $port" >&2
        exit 2
    fi
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

cleanup() {
    echo ""
    echo -e "${YELLOW}Caught signal, shutting down all services...${NC}"
    "$SCRIPT_DIR/stop-all.sh"
    exit 0
}

trap cleanup SIGINT SIGTERM

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  PremierPro MCP — Starting All Services${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""
echo "Project root: $PROJECT_ROOT"
echo ""

validate_port "Rust media engine" "$RUST_PORT"
validate_port "Python intelligence" "$PYTHON_PORT"
validate_port "TypeScript bridge" "$TS_PORT"
if ! [[ "$SERVICE_START_TIMEOUT" =~ ^[1-9][0-9]*$ ]]; then
    echo "SERVICE_START_TIMEOUT must be a positive number of seconds." >&2
    exit 2
fi
if [ "$RUST_PORT" = "$PYTHON_PORT" ] ||
   [ "$RUST_PORT" = "$TS_PORT" ] ||
   [ "$PYTHON_PORT" = "$TS_PORT" ]; then
    echo -e "${RED}Backend services must use three distinct ports.${NC}" >&2
    exit 2
fi

# Never signal PIDs merely because a stale file exists. A recycled PID may
# belong to an unrelated application.
if [ -f "$PID_FILE" ]; then
    while IFS=: read -r existing_name existing_pid existing_port existing_start; do
        [ -z "$existing_name" ] && continue
        if kill -0 "$existing_pid" 2>/dev/null &&
           [ -n "$existing_start" ] &&
           [ "$(process_start_stamp "$existing_pid")" = "$existing_start" ]; then
            echo -e "${RED}$existing_name is already managed as PID $existing_pid.${NC}"
            echo "Run: just status, then just stop"
            exit 1
        fi
    done < "$PID_FILE"
    echo -e "${YELLOW}Ignoring stale or unverifiable PID records; no processes were signaled.${NC}"
fi

# Refuse to launch into ports already owned by unmanaged processes. Without
# this preflight, a newly spawned process can fail while a foreign listener
# makes a simple port readiness check look successful.
for service_port in "$RUST_PORT" "$PYTHON_PORT" "$TS_PORT"; do
    if tcp_port_open "$service_port"; then
        echo -e "${RED}Port $service_port is already in use; no services were started.${NC}" >&2
        exit 1
    fi
done

# Create log directory
mkdir -p "$LOG_DIR"

# Initialize PID file
> "$PID_FILE"

# --- Start Rust Media Engine ---
echo -e "${CYAN}[1/3] Starting Rust media engine on port $RUST_PORT...${NC}"
cd "$PROJECT_ROOT/rust-engine"
if [ -x "target/release/premierpro-media-engine" ]; then
    ./target/release/premierpro-media-engine --port "$RUST_PORT" > "$LOG_DIR/rust-engine.log" 2>&1 &
else
    cargo build --release
    ./target/release/premierpro-media-engine --port "$RUST_PORT" > "$LOG_DIR/rust-engine.log" 2>&1 &
fi
RUST_PID=$!
record_process "rust-engine" "$RUST_PID" "$RUST_PORT"
echo "  PID: $RUST_PID"

# --- Start Python Intelligence ---
echo -e "${CYAN}[2/3] Starting Python intelligence on port $PYTHON_PORT...${NC}"
cd "$PROJECT_ROOT/python-intelligence"
PYTHONPATH="$PROJECT_ROOT/gen/python:." "$PYTHON_BIN" src/main.py --port "$PYTHON_PORT" > "$LOG_DIR/python-intelligence.log" 2>&1 &
PYTHON_PID=$!
record_process "python-intelligence" "$PYTHON_PID" "$PYTHON_PORT"
echo "  PID: $PYTHON_PID"

# --- Start TypeScript Bridge ---
echo -e "${CYAN}[3/3] Starting TypeScript bridge on port $TS_PORT...${NC}"
cd "$PROJECT_ROOT/ts-bridge"
BRIDGE_GRPC_PORT="$TS_PORT" node --import tsx src/index.ts > "$LOG_DIR/ts-bridge.log" 2>&1 &
TS_PID=$!
record_process "ts-bridge" "$TS_PID" "$TS_PORT"
echo "  PID: $TS_PID"

# --- Wait and verify ---
echo ""
echo -e "${YELLOW}Waiting up to ${SERVICE_START_TIMEOUT}s for each gRPC listener...${NC}"

echo ""
echo -e "${CYAN}─── Service Status ───${NC}"
printf "%-25s %-10s %-10s\n" "SERVICE" "PID" "STATUS"
printf "%-25s %-10s %-10s\n" "-------" "---" "------"

ALL_OK=true

while IFS=: read -r name pid port expected_start; do
    if wait_for_service "$pid" "$port" "$expected_start"; then
        printf "%-25s %-10s ${GREEN}%-10s${NC}\n" "$name (port $port)" "$pid" "READY"
    else
        printf "%-25s %-10s ${RED}%-10s${NC}\n" "$name (port $port)" "$pid" "FAILED"
        echo -e "  ${RED}Check logs: $LOG_DIR/${name}.log${NC}"
        ALL_OK=false
    fi
done < "$PID_FILE"

echo ""
if [ "$ALL_OK" = true ]; then
    echo -e "${GREEN}All services started successfully.${NC}"
    echo "Logs are in: $LOG_DIR/"
    echo "PIDs saved to: $PID_FILE"
    echo ""
    echo "Press Ctrl+C to stop all services, or run: just stop"
else
    echo -e "${RED}Some services failed to start. Check the logs above.${NC}"
    "$SCRIPT_DIR/stop-all.sh" || true
    exit 1
fi

# Wait for all background processes
wait
