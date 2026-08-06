# Generate the ignored Go/Python protobuf clients so a clean checkout builds.
FROM bufbuild/buf:1.72.0 AS proto-builder
WORKDIR /workspace
COPY buf.yaml buf.gen.yaml ./
COPY proto/ ./proto/
# buf.gen.yaml intentionally preserves the generated Go module metadata.
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
RUN buf generate

# Go orchestrator
FROM golang:1.26.1-alpine AS go-builder
WORKDIR /app
COPY go-orchestrator/ ./go-orchestrator/
COPY --from=proto-builder /workspace/gen/go/ ./gen/go/
RUN cd go-orchestrator && go build -trimpath -o /bin/premierpro-mcp ./cmd/server/

# Rust media engine
FROM rust:1.97-bookworm AS rust-builder
WORKDIR /app
COPY rust-engine/ ./rust-engine/
COPY proto/ ./proto/
RUN cd rust-engine \
    && cargo build --locked --release \
    && cp target/release/premierpro-media-engine /bin/

# Compile the TypeScript bridge under the Node version declared by package.json.
FROM node:20-bookworm-slim AS ts-builder
WORKDIR /app/ts-bridge
COPY ts-bridge/package.json ts-bridge/package-lock.json ./
RUN npm ci
COPY ts-bridge/src/ ./src/
COPY ts-bridge/tsconfig.json ./
RUN npm run build && npm prune --omit=dev

# Shared runtime dependencies. libatomic1 is required by the Node 20 binary in
# the optional full-runtime target.
FROM python:3.12-slim-bookworm AS runtime-base
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg libatomic1 \
    && rm -rf /var/lib/apt/lists/*

# Compose builds these service-specific targets instead of duplicating one
# all-language image for both analysis services.
FROM runtime-base AS rust-runtime
COPY --from=rust-builder /bin/premierpro-media-engine /usr/local/bin/
EXPOSE 50052
CMD ["premierpro-media-engine", "--port", "50052"]

FROM runtime-base AS python-runtime
COPY python-intelligence/ /app/python-intelligence/
COPY --from=proto-builder /workspace/gen/python/ /app/gen/python/
RUN python -m pip install \
    --no-cache-dir \
    --constraint /app/python-intelligence/constraints.txt \
    /app/python-intelligence
WORKDIR /app/python-intelligence
ENV PYTHONPATH=/app/gen/python:/app/python-intelligence
EXPOSE 50053
CMD ["python", "-m", "src.main", "--port", "50053"]

# Optional combined diagnostics image. Live Premiere editing still keeps Go and
# TypeScript on the host because the authenticated CEP socket is loopback-only.
FROM python-runtime AS full-runtime
COPY --from=go-builder /bin/premierpro-mcp /usr/local/bin/
COPY --from=rust-builder /bin/premierpro-media-engine /usr/local/bin/
COPY --from=ts-builder /usr/local/bin/node /usr/local/bin/node

COPY --from=ts-builder /app/ts-bridge/package.json /app/ts-bridge/package.json
COPY --from=ts-builder /app/ts-bridge/node_modules/ /app/ts-bridge/node_modules/
COPY --from=ts-builder /app/ts-bridge/dist/ /app/ts-bridge/dist/

# The bridge dynamically loads the source protobuf definitions at runtime.
COPY proto/definitions/ /app/proto/definitions/

WORKDIR /app
EXPOSE 50052 50053 50054 8080

CMD ["premierpro-mcp", "--transport", "stdio"]
