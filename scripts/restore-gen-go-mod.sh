#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

cat > gen/go/go.mod <<'EOF'
module github.com/anthropics/premierpro-mcp/gen/go

go 1.26.1

require (
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)
EOF

cd gen/go && go mod tidy
