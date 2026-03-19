# Multi-stage build for Go orchestrator
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go-orchestrator/ ./go-orchestrator/
COPY gen/go/ ./gen/go/
RUN cd go-orchestrator && go build -o /bin/premierpro-mcp ./cmd/server/

# Rust builder
FROM rust:1.77-alpine AS rust-builder
RUN apk add --no-cache musl-dev protobuf-dev
WORKDIR /app
COPY rust-engine/ ./rust-engine/
COPY proto/ ./proto/
RUN cd rust-engine && cargo build --release && cp target/release/premierpro-media-engine /bin/

# Python + Node runtime
FROM python:3.12-slim
RUN apt-get update && apt-get install -y nodejs npm ffmpeg && rm -rf /var/lib/apt/lists/*

# Copy binaries
COPY --from=go-builder /bin/premierpro-mcp /usr/local/bin/
COPY --from=rust-builder /bin/premierpro-media-engine /usr/local/bin/

# Copy Python service
COPY python-intelligence/ /app/python-intelligence/
COPY gen/python/ /app/gen/python/
RUN pip install grpcio protobuf pydantic structlog numpy scikit-learn pypdf python-docx "protobuf>=7.34.0"

# Copy TypeScript bridge
COPY ts-bridge/ /app/ts-bridge/
RUN cd /app/ts-bridge && npm install

WORKDIR /app
EXPOSE 50051 50052 50053 50054

CMD ["premierpro-mcp", "--transport", "stdio"]
