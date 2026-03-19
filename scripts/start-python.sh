#!/bin/bash
# Start just the Python intelligence service

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PORT="${1:-50053}"

echo "Starting Python intelligence on port $PORT..."
cd "$PROJECT_ROOT/python-intelligence" && PYTHONPATH="$PROJECT_ROOT/gen/python:." python src/main.py --port "$PORT"
