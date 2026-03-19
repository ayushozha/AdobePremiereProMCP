#!/bin/bash
# Check status of all PremierPro MCP backend services

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="$SCRIPT_DIR/.pids"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  PremierPro MCP — Service Status${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if [ ! -f "$PID_FILE" ]; then
    echo -e "${YELLOW}No PID file found at $PID_FILE${NC}"
    echo "No services appear to be managed. Run: just start"
    exit 0
fi

printf "%-25s %-8s %-12s %-12s\n" "SERVICE" "PID" "PROCESS" "gRPC PORT"
printf "%-25s %-8s %-12s %-12s\n" "-------" "---" "-------" "---------"

while IFS=: read -r name pid port; do
    [ -z "$name" ] && continue

    # Check if the process is alive
    if kill -0 "$pid" 2>/dev/null; then
        proc_status="${GREEN}RUNNING${NC}"
    else
        proc_status="${RED}DEAD${NC}"
    fi

    # Check if the gRPC port is reachable
    if command -v nc &>/dev/null; then
        if nc -z 127.0.0.1 "$port" 2>/dev/null; then
            port_status="${GREEN}OPEN${NC}"
        else
            port_status="${RED}CLOSED${NC}"
        fi
    elif command -v lsof &>/dev/null; then
        if lsof -i :"$port" -sTCP:LISTEN &>/dev/null; then
            port_status="${GREEN}LISTENING${NC}"
        else
            port_status="${RED}CLOSED${NC}"
        fi
    else
        port_status="${YELLOW}UNKNOWN${NC}"
    fi

    printf "%-25s %-8s " "$name" "$pid"
    printf "%-22b " "$proc_status"
    printf "%-22b\n" "$port_status"
done < "$PID_FILE"

echo ""

# Summary
total=0
running=0
while IFS=: read -r name pid port; do
    [ -z "$name" ] && continue
    total=$((total + 1))
    if kill -0 "$pid" 2>/dev/null; then
        running=$((running + 1))
    fi
done < "$PID_FILE"

if [ "$running" -eq "$total" ] && [ "$total" -gt 0 ]; then
    echo -e "${GREEN}All $total services are running.${NC}"
elif [ "$running" -eq 0 ]; then
    echo -e "${RED}No services are running.${NC}"
else
    echo -e "${YELLOW}$running of $total services are running.${NC}"
fi
