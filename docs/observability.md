# Local observability

The Go orchestrator can expose Prometheus metrics and health on a separate HTTP listener. It is disabled by default and works with both stdio and SSE MCP transports. It does not launch Prometheus, Grafana, or the backend services.

## Enable

Set this environment variable on the MCP server process (or in the MCP client's server environment):

```sh
MCP_OBSERVABILITY_ADDR=127.0.0.1:9091 ./go-orchestrator/bin/premierpro-mcp --transport stdio
```

The address must use a literal loopback IP and a port from 1–65535; `[::1]:9091` also works. Wildcard, public, and hostname bindings are rejected. Use a distinct port for each MCP process. A port conflict fails startup with an error on stderr. The server reads environment variables directly; it does not load `.env` files.

```sh
curl http://127.0.0.1:9091/metrics
curl http://127.0.0.1:9091/livez
curl -i http://127.0.0.1:9091/readyz
curl -i http://127.0.0.1:9091/health/premiere-bridge
```

MCP responses remain on stdout. Logs remain JSON on stderr. Closing stdio or sending SIGTERM/SIGINT stops the metrics listener and background probes. Disabling the variable removes the listener, probes, and completion instrumentation.

## Metrics and logs

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `premiere_mcp_tool_requests_total` | Counter | `tool`, `outcome` | Completed registered tool calls, including failed argument validation and recovered handler panics |
| `premiere_mcp_tool_duration_seconds` | Histogram | `tool` | Elapsed seconds through validation and handler execution, including backend calls, for all outcomes |

Outcomes are `success`, `error`, `timeout`, and `canceled`. A returned MCP result with `isError: true` counts as an error even when the Go handler returns no error. Go errors, nil results, and recovered panics also count as errors. Context cancellation/deadline errors (including gRPC status errors) have separate outcomes. When a handler converts cancellation into an MCP error result, the canceled request context preserves its classification. A successful result remains a success even if cancellation races with its return.

The count is recorded once when a call finishes, not when it starts. Work still running has no completed sample. Malformed JSON-RPC and unknown/unregistered tool names are rejected before handler dispatch and are excluded. Cancellation sent only by a transport that does not cancel the handler context cannot be identified here.

The tool label is restricted to the active registered catalog; other names seen by the middleware collapse to `unknown`. Outcome labels are fixed. Counters start at zero for each registered tool/outcome so error-rate queries work before the first failure. Durations use buckets at 5 ms, 25 ms, 100 ms, 500 ms, 1 s, 2.5 s, 5 s, 10 s, 30 s, 60 s, 120 s, and 300 s, plus the infinite bucket. Prometheus also exposes `_bucket`, `_sum`, and `_count`. Latency percentiles above 300 seconds are limited by the last finite bucket. All measurements reset when the process restarts.

Each completion adds a JSON log with `msg: "MCP tool completed"`, `tool`, `outcome`, and `duration_seconds`. The new metrics and completion logs contain no prompts, arguments, results, raw errors, file paths, request IDs, tokens, or user identifiers. Existing backend/tool logs have their own logging behavior; this feature does not rewrite them.

For example, unsuccessful calls per tool as a percentage of completions:

```promql
100 * sum by (tool) (rate(premiere_mcp_tool_requests_total{outcome=~"error|timeout|canceled"}[5m]))
  / (sum by (tool) (rate(premiere_mcp_tool_requests_total[5m])) > 0)
```

Canceled calls are included in that ratio; filter the outcomes to `error|timeout` if operator cancellations should be excluded. Rates need at least two scrapes. An idle tool has no error percentage or latency percentile.

## Health semantics

| Route | HTTP behavior |
| --- | --- |
| `/livez` | 200 when the Go HTTP server is responding, independent of dependencies |
| `/readyz` | 200 only when every tracked dependency has a successful probe no older than 60 seconds; otherwise 503 |
| `/health` | Same readiness status, with each dependency's last observed status, readiness, latency, last check time, and sanitized probe error |
| `/health/{service}` | 200 for a ready dependency, 503 otherwise, 404 for an unknown service |

Dependencies are `media-engine`, `intelligence`, and `premiere-bridge`. Initial state is not ready. Probes start immediately, then run every 10 seconds, with a 5-second limit per dependency. They run concurrently; scrapes read cached state rather than launching new probes. Dependency loss can take one probe interval plus its timeout to appear. One failed probe makes readiness false immediately; the descriptive status becomes `unhealthy` after three failures. A slow but successful response is `degraded` and remains ready. Stale last-success status is accompanied by `ready: false`.

Media and intelligence currently have no application health RPC, so their probes verify a gRPC HTTP/2 connection. They do **not** validate model loading, FFmpeg, or a specific RPC operation. Premiere's probe calls the existing read-only Ping RPC and also requires `premiere_running=true`; a reachable TypeScript service without a connected Premiere application is not ready. An open project is not required. Readiness does not disable individual tools: partial workflows may remain usable while a dependency is unavailable.

Health responses use generic errors for the production probes, without returning raw authentication or RPC error contents. The listener has no remote authentication or TLS and deliberately accepts only local binds. Do not publicly proxy it without access controls.

## Prometheus and Grafana

Use [`monitoring/prometheus.yml`](../monitoring/prometheus.yml) with Prometheus on the same host/network namespace as the MCP server. The example scrapes `127.0.0.1:9091` every 15 seconds. A container's loopback is its own network namespace; the sample does not automatically reach a host MCP process.

Import [`monitoring/grafana/premiere-mcp.json`](../monitoring/grafana/premiere-mcp.json) in Grafana and select the Prometheus data source when prompted. Select job, instance, and tool using the dashboard filters. The template provides request rate, unsuccessful-call percentage, p95 latency, completions by outcome, cumulative calls, and scrape reachability. `up` measures scrape reachability, not dependency readiness.

## Verification

```sh
bash scripts/generate-proto.sh
cd go-orchestrator
go test -race ./...
go vet ./...
```

The Go tests cover real stdio dispatch and stdout isolation, validation and recovered panic counts, MCP `isError`, cancellation and timeout classification, 100 concurrent calls with scrapes, elapsed seconds, label/payload privacy, loopback configuration, HTTP routing, dependency loss, Premiere-disconnected readiness, stale probes, and listener shutdown. No test edits a live Premiere project. The Grafana JSON is a template; verify import/rendering against your Grafana installation.
