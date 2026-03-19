#!/bin/bash
# Stop all PremierPro MCP backend services

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="$SCRIPT_DIR/.pids"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

TIMEOUT=5  # Seconds to wait before SIGKILL

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  PremierPro MCP — Stopping All Services${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if [ ! -f "$PID_FILE" ]; then
    echo -e "${YELLOW}No PID file found at $PID_FILE${NC}"
    echo "No services appear to be running."
    exit 0
fi

while IFS=: read -r name pid port; do
    [ -z "$name" ] && continue

    echo -n "Stopping $name (PID $pid, port $port)... "

    if ! kill -0 "$pid" 2>/dev/null; then
        echo -e "${YELLOW}already stopped${NC}"
        continue
    fi

    # Send SIGTERM for graceful shutdown
    kill -TERM "$pid" 2>/dev/null

    # Wait up to TIMEOUT seconds for process to exit
    elapsed=0
    while kill -0 "$pid" 2>/dev/null && [ "$elapsed" -lt "$TIMEOUT" ]; do
        sleep 1
        elapsed=$((elapsed + 1))
    done

    if kill -0 "$pid" 2>/dev/null; then
        # Process did not exit gracefully, force kill
        echo -n "force killing... "
        kill -KILL "$pid" 2>/dev/null
        sleep 1
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "${RED}FAILED to kill${NC}"
        else
            echo -e "${YELLOW}killed (SIGKILL)${NC}"
        fi
    else
        echo -e "${GREEN}stopped${NC}"
    fi
done < "$PID_FILE"

# Clean up the PID file
rm -f "$PID_FILE"
echo ""
echo -e "${GREEN}All services stopped. PID file cleaned up.${NC}"
