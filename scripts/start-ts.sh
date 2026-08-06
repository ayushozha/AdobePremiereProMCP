#!/bin/bash
# Start just the TypeScript bridge

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$PROJECT_ROOT/.env"
    set +a
fi

PORT="${1:-${BRIDGE_GRPC_PORT:-50054}}"

echo "Starting TypeScript bridge on port $PORT..."
cd "$PROJECT_ROOT/ts-bridge"
export BRIDGE_GRPC_PORT="$PORT"
exec node --import tsx src/index.ts
