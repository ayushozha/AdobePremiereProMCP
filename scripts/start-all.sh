#!/bin/bash
# Start all PremierPro MCP backend services

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_FILE="$SCRIPT_DIR/.pids"
LOG_DIR="$SCRIPT_DIR/logs"

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

# Clean up stale PID file
if [ -f "$PID_FILE" ]; then
    echo -e "${YELLOW}Cleaning up stale PID file...${NC}"
    "$SCRIPT_DIR/stop-all.sh" 2>/dev/null || true
fi

# Create log directory
mkdir -p "$LOG_DIR"

# Initialize PID file
> "$PID_FILE"

# --- Start Rust Media Engine (port 50052) ---
echo -e "${CYAN}[1/3] Starting Rust media engine on port 50052...${NC}"
cd "$PROJECT_ROOT/rust-engine"
if [ -f "target/release/rust-engine" ]; then
    ./target/release/rust-engine --port 50052 > "$LOG_DIR/rust-engine.log" 2>&1 &
else
    cargo run --release -- --port 50052 > "$LOG_DIR/rust-engine.log" 2>&1 &
fi
RUST_PID=$!
echo "rust-engine:$RUST_PID:50052" >> "$PID_FILE"
echo "  PID: $RUST_PID"

# --- Start Python Intelligence (port 50053) ---
echo -e "${CYAN}[2/3] Starting Python intelligence on port 50053...${NC}"
cd "$PROJECT_ROOT/python-intelligence"
PYTHONPATH="$PROJECT_ROOT/gen/python:." python src/main.py --port 50053 > "$LOG_DIR/python-intelligence.log" 2>&1 &
PYTHON_PID=$!
echo "python-intelligence:$PYTHON_PID:50053" >> "$PID_FILE"
echo "  PID: $PYTHON_PID"

# --- Start TypeScript Bridge (port 50054) ---
echo -e "${CYAN}[3/3] Starting TypeScript bridge on port 50054...${NC}"
cd "$PROJECT_ROOT/ts-bridge"
npx tsx src/index.ts --port 50054 > "$LOG_DIR/ts-bridge.log" 2>&1 &
TS_PID=$!
echo "ts-bridge:$TS_PID:50054" >> "$PID_FILE"
echo "  PID: $TS_PID"

# --- Wait and verify ---
echo ""
echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 3

echo ""
echo -e "${CYAN}─── Service Status ───${NC}"
printf "%-25s %-10s %-10s\n" "SERVICE" "PID" "STATUS"
printf "%-25s %-10s %-10s\n" "-------" "---" "------"

ALL_OK=true

while IFS=: read -r name pid port; do
    if kill -0 "$pid" 2>/dev/null; then
        printf "%-25s %-10s ${GREEN}%-10s${NC}\n" "$name (port $port)" "$pid" "RUNNING"
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
fi

# Wait for all background processes
wait
