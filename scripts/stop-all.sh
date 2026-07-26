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

# Ports start-all.sh binds, used to clean up orphans when .pids is gone.
DEFAULT_PORTS="50052 50053 50054"

# On Git Bash/MSYS the PIDs recorded by start-all.sh are shell PIDs, not the
# Windows PIDs of the detached services, so `kill` reports success while the
# process keeps holding its port and the next start fails with EADDRINUSE.
# Fall back to killing whatever owns the port instead.
case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*) IS_WINDOWS=1 ;;
    *)                    IS_WINDOWS=0 ;;
esac

kill_by_port_windows() {
    local port="$1"
    local pids
    pids=$(powershell.exe -NoProfile -Command \
        "(Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue).OwningProcess" \
        2>/dev/null | tr -d '\r')
    [ -z "$pids" ] && return 1
    for wpid in $pids; do
        [ -z "$wpid" ] && continue
        powershell.exe -NoProfile -Command \
            "Stop-Process -Id $wpid -Force -ErrorAction SilentlyContinue" >/dev/null 2>&1
    done
    return 0
}

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  PremierPro MCP — Stopping All Services${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if [ ! -f "$PID_FILE" ]; then
    echo -e "${YELLOW}No PID file found at $PID_FILE${NC}"
    # A missing PID file does not mean nothing is running: it is deleted on
    # every stop, so a service orphaned by a crash or a previous stop still
    # holds its port and would make the next start fail with EADDRINUSE.
    # Sweep the well-known ports instead of assuming we are clean.
    if [ "$IS_WINDOWS" -eq 1 ]; then
        swept=0
        for port in $DEFAULT_PORTS; do
            if kill_by_port_windows "$port"; then
                echo -e "  ${GREEN}freed orphaned service on port $port${NC}"
                swept=1
            fi
        done
        [ "$swept" -eq 0 ] && echo "No services appear to be running."
    else
        echo "No services appear to be running."
    fi
    exit 0
fi

while IFS=: read -r name pid port; do
    [ -z "$name" ] && continue

    echo -n "Stopping $name (PID $pid, port $port)... "

    if [ "$IS_WINDOWS" -eq 1 ]; then
        kill -TERM "$pid" 2>/dev/null   # best effort for the shell-level job
        if kill_by_port_windows "$port"; then
            echo -e "${GREEN}stopped${NC} (port $port)"
        else
            echo -e "${YELLOW}already stopped${NC}"
        fi
        continue
    fi

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
