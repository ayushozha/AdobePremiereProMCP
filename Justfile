# PremierPro MCP Server — Unified Build System

default:
    @just --list

# ─── Proto ───

# Generate protobuf stubs for all languages
proto:
    @echo "Generating protobuf stubs..."
    buf generate proto/definitions

# Lint proto definitions
proto-lint:
    buf lint proto/definitions

# ─── Go Orchestrator ───

# Build the Go orchestrator
go-build:
    cd go-orchestrator && go build -o bin/server ./cmd/server

# Run the Go orchestrator
go-run:
    cd go-orchestrator && go run ./cmd/server

# Test Go code
go-test:
    cd go-orchestrator && go test ./...

# Lint Go code
go-lint:
    cd go-orchestrator && golangci-lint run

# ─── Rust Engine ───

# Build the Rust media engine
rust-build:
    cd rust-engine && cargo build --release

# Run Rust tests
rust-test:
    cd rust-engine && cargo test

# Lint Rust code
rust-lint:
    cd rust-engine && cargo clippy -- -D warnings

# ─── Python Intelligence ───

# Install Python dependencies
py-install:
    cd python-intelligence && pip install -e ".[dev]"

# Run Python tests
py-test:
    cd python-intelligence && pytest tests/

# Lint Python code
py-lint:
    cd python-intelligence && ruff check src/ && mypy src/

# ─── TypeScript Bridge ───

# Install TypeScript dependencies
ts-install:
    cd ts-bridge && npm install

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

# Build the CEP panel
cep-build:
    cd cep-panel && npm run build

# Package the CEP panel for installation
cep-package:
    cd cep-panel && npm run package

# ─── All ───

# Install all dependencies
install: py-install ts-install
    @echo "Dependencies installed."

# Build everything
build: proto go-build rust-build ts-build cep-build
    @echo "All components built."

# Run all tests
test: go-test rust-test py-test ts-test
    @echo "All tests passed."

# Lint everything
lint: proto-lint go-lint rust-lint py-lint ts-lint
    @echo "All lints passed."

# Full CI pipeline
ci: lint build test
    @echo "CI pipeline complete."

# Clean all build artifacts
clean:
    rm -rf go-orchestrator/bin/
    cd rust-engine && cargo clean
    rm -rf python-intelligence/dist/ python-intelligence/build/ python-intelligence/*.egg-info
    rm -rf ts-bridge/dist/ ts-bridge/node_modules/
    rm -rf cep-panel/dist/ cep-panel/build/ cep-panel/node_modules/
    @echo "Cleaned."

# ─── Dev ───

# Start all services in dev mode
dev:
    @echo "Starting all services..."
    just go-run &
    @echo "Go orchestrator started."
