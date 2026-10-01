# @premierpro/mcp-server

The prebuilt Go MCP server from [AdobePremiereProMCP](https://github.com/ayushozha/AdobePremiereProMCP). Requires Node.js 20 or newer. Supports macOS arm64/x64, Linux x64, and Windows x64.

```sh
npm install -g @premierpro/mcp-server
premierpro-mcp --transport stdio
```

Configure your MCP client to run `premierpro-mcp` with arguments `["--transport", "stdio"]`. Set `MCP_TOOL_PROFILE=standard` in the client's environment for the curated tool set. To pin a version, install `@premierpro/mcp-server@VERSION`.

Installation downloads one native binary from the matching GitHub release and verifies its SHA-256 checksum against the manifest shipped in this package. First launch also installs it if npm lifecycle scripts were disabled. Later launches verify the cached file and need no network. Downloads require HTTPS access to `github.com` and its release asset hosts. Installation errors go to stderr.

**This package installs only the Go MCP server.** Editing Premiere still requires the TypeScript bridge, the CEP panel, and Adobe Premiere Pro on macOS or Windows. Media analysis and intelligence tools also need the Rust and Python services. Linux can host the server but cannot run Premiere Pro. The server does not start those services or load `.env` files; pass configuration through your MCP client's environment.

See [setup and release instructions](https://github.com/ayushozha/AdobePremiereProMCP/blob/main/docs/npm-distribution.md). The source checkout intentionally has an empty release manifest: maintainers must stage a release before packing it.
