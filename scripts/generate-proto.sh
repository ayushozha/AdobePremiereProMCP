#!/bin/bash
# Generate protobuf clients into a staging directory, then synchronize only
# the generated package trees. This removes stale stubs without deleting
# gen/go/go.mod or gen/go/go.sum.

set -euo pipefail

for required_command in buf rsync; do
    if ! command -v "$required_command" >/dev/null 2>&1; then
        echo "Missing required command: $required_command" >&2
        exit 1
    fi
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
STAGING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/premierpro-proto.XXXXXX")"

cleanup() {
    case "$STAGING_DIR" in
        */premierpro-proto.*)
            rm -rf -- "$STAGING_DIR"
            ;;
        *)
            echo "Refusing to remove unexpected staging path: $STAGING_DIR" >&2
            ;;
    esac
}
trap cleanup EXIT

cd "$PROJECT_ROOT"
buf generate -o "$STAGING_DIR"

# Verify every expected output exists before changing the working tree.
for language in go python; do
    if [ ! -d "$STAGING_DIR/gen/$language/premierpro" ]; then
        echo "Missing generated $language package tree; existing stubs were preserved." >&2
        exit 1
    fi
done

for language in go python; do
    source_dir="$STAGING_DIR/gen/$language/premierpro/"
    target_dir="$PROJECT_ROOT/gen/$language/premierpro/"
    mkdir -p "$target_dir"
    rsync --archive --delete "$source_dir" "$target_dir"
done

echo "Generated Go and Python protobuf clients. The TypeScript bridge loads the"
echo "versioned .proto definitions directly through @grpc/proto-loader."
