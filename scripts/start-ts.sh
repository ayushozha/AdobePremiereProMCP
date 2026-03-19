#!/bin/bash
# Start just the TypeScript bridge

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PORT="${1:-50054}"

echo "Starting TypeScript bridge on port $PORT..."
cd "$PROJECT_ROOT/ts-bridge" && npx tsx src/index.ts --port "$PORT"
