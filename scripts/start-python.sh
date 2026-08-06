#!/bin/bash
# Start just the Python intelligence service

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$PROJECT_ROOT/.env"
    set +a
fi

PORT="${1:-${INTEL_GRPC_PORT:-50053}}"
PYTHON_BIN="${PYTHON_BIN:-python3}"

echo "Starting Python intelligence on port $PORT..."
cd "$PROJECT_ROOT/python-intelligence"
export PYTHONPATH="$PROJECT_ROOT/gen/python:."
exec "$PYTHON_BIN" src/main.py --port "$PORT"
