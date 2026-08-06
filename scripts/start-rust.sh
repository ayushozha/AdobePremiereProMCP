#!/bin/bash
# Start just the Rust media engine

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    . "$PROJECT_ROOT/.env"
    set +a
fi

PORT="${1:-${MEDIA_ENGINE_PORT:-50052}}"
BINARY="$PROJECT_ROOT/rust-engine/target/release/premierpro-media-engine"

echo "Starting Rust media engine on port $PORT..."
if [ -x "$BINARY" ]; then
    exec "$BINARY" --port "$PORT"
else
    cd "$PROJECT_ROOT/rust-engine"
    cargo build --release
    exec "$BINARY" --port "$PORT"
fi
