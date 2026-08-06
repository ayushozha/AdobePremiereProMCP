# PremierPro MCP Server — Unified Build System

default:
    @just --list

# ─── Proto ───

# Generate Go/Python stubs. Rust generates at build time; TypeScript loads the
# versioned .proto definitions directly.
proto:
    @echo "Generating protobuf stubs..."
    bash scripts/generate-proto.sh

# Lint proto definitions
proto-lint:
    buf lint proto/definitions

# ─── Go Orchestrator ───

# Build the Go orchestrator
go-build: proto
    cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server

# Run the Go orchestrator
go-run:
    cd go-orchestrator && go run ./cmd/server

# Test Go code
go-test: proto
    cd go-orchestrator && go test ./...

# Lint Go code
go-lint: proto
    cd go-orchestrator && go vet ./...

# ─── Rust Engine ───

# Build the Rust media engine
rust-build:
    cd rust-engine && cargo build --locked --release

# Run Rust tests
rust-test:
    cd rust-engine && cargo test --locked --all-targets

# Lint Rust code
rust-lint:
    cd rust-engine && cargo clippy --locked --all-targets -- -D warnings

# ─── Python Intelligence ───

# Install Python dependencies
py-install:
    cd python-intelligence && python3 -m pip install --constraint constraints.txt -e ".[dev]"

# Run Python tests
py-test: proto
    cd python-intelligence && python3 -m pytest tests/

# Lint Python code
py-lint: proto
    cd python-intelligence && python3 -m ruff check src/ && python3 -m mypy src/

# ─── TypeScript Bridge ───

# Install TypeScript dependencies
ts-install:
    cd ts-bridge && npm ci

# Build the TypeScript bridge
ts-build:
    cd ts-bridge && npm run build

# Run TypeScript tests
ts-test:
    cd ts-bridge && npm test

# Lint TypeScript code
ts-lint:
    cd ts-bridge && npm run lint

# ─── CEP Panel ───

# Install CEP panel runtime dependencies
cep-install:
    cd cep-panel && npm ci

# Build the CEP panel
cep-build:
    cd cep-panel && npm run build

# Package the CEP panel for installation
cep-package:
    cd cep-panel && npm run package

# Check every Go EvalCommand has a host function
cep-test:
    cd cep-panel && npm run check:commands

# ─── CLI ───

# Install CLI dependencies
cli-install:
    cd cli && npm ci

# Build the CLI
cli-build:
    cd cli && npm run build

# ─── All ───

# Install all dependencies
install: py-install ts-install cep-install cli-install
    @echo "Dependencies installed."

# Build everything
build: proto go-build rust-build ts-build cep-build cli-build
    @echo "All components built."

# Run all tests
test: go-test rust-test py-test ts-test cep-test
    @echo "All tests passed."

# Lint everything
lint: proto-lint go-lint rust-lint py-lint ts-lint
    @echo "All lints passed."

# Full CI pipeline
ci: proto lint build test
    @echo "CI pipeline complete."

# Clean all build artifacts
clean:
    rm -rf go-orchestrator/bin/
    cd rust-engine && cargo clean
    rm -rf python-intelligence/dist/ python-intelligence/build/ python-intelligence/*.egg-info
    rm -rf ts-bridge/dist/ ts-bridge/node_modules/
    rm -rf cep-panel/dist/ cep-panel/build/ cep-panel/node_modules/
    @echo "Cleaned."

# ─── Services ───

# Start all backend services
start:
    ./scripts/start-all.sh

# Stop all backend services
stop:
    ./scripts/stop-all.sh

# Check status of all backend services
status:
    ./scripts/status.sh

# Install CEP panel into Premiere Pro (macOS/Linux)
install-panel:
    ./scripts/install-cep-panel.sh

# Install CEP panel into Premiere Pro (Windows)
install-panel-win:
    scripts\\install-cep-panel-win.bat

# ─── Dev ───

# Start the long-running Rust, Python, and TypeScript development backends.
# The Go MCP server uses stdio and is launched by the configured MCP client.
dev:
    @echo "Starting development backends (the MCP client launches Go)..."
    ./scripts/start-all.sh
