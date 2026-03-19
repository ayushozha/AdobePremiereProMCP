#!/bin/bash
# Start just the Rust media engine

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PORT="${1:-50052}"

echo "Starting Rust media engine on port $PORT..."
cd "$PROJECT_ROOT/rust-engine" && cargo run --release -- --port "$PORT"
