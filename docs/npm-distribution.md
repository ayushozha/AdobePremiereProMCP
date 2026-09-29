# npm distribution

The npm package is `@premierpro/mcp-server`; the executable is `premierpro-mcp`.
Publication requires a maintainer to complete the npm setup below. Adding these
files alone does not make the package available on npm.

## Install and configure

After a version is published:

```sh
npm install -g @premierpro/mcp-server
premierpro-mcp --transport stdio
```

Requires Node.js 20 or newer. Use a currently supported Node release for normal
installations. The wrapper selects the architecture of the Node process; an x64
Node under Rosetta installs the x64 macOS binary.

| System | CPU | GitHub release asset |
| --- | --- | --- |
| macOS | arm64 | `premierpro-mcp-darwin-arm64` |
| macOS | x64 | `premierpro-mcp-darwin-x64` |
| Linux | x64 | `premierpro-mcp-linux-x64` |
| Windows | x64 | `premierpro-mcp-win-x64.exe` |

Linux arm64, Windows arm64 and 32-bit systems are not packaged. Unsupported
combinations fail with an explicit error before downloading anything.

Example MCP client configuration:

```json
{
  "mcpServers": {
    "premiere-pro": {
      "command": "premierpro-mcp",
      "args": ["--transport", "stdio"],
      "env": { "MCP_TOOL_PROFILE": "standard" }
    }
  }
}
```

The client must inherit a PATH containing npm's global executable directory.
Some Windows MCP clients cannot launch npm's `.cmd` shim directly; configure
`node` as the command and pass the absolute path to
`node_modules/@premierpro/mcp-server/bin/premierpro-mcp.cjs` under the directory
printed by `npm prefix -g`, followed by `--transport`, `stdio`.

For reproducible installs, use `npm install -g @premierpro/mcp-server@VERSION`.
The Go binary always comes from `releases/download/vVERSION/ASSET`; the wrapper
never resolves a mutable `latest` release. Ordinary arguments and environment
variables pass directly to the Go process. Logs and failures go to stderr;
stdout and stdin remain the MCP protocol channel. On Unix, termination signals
are forwarded and child signal exits are preserved.

## What is included

Only the Go MCP server is installed. The TypeScript bridge, CEP panel, and Adobe
Premiere Pro must still be installed and running for editing tools. Rust and
Python services are required by the media analysis and intelligence tools.
Follow the repository's [setup instructions](../README.md#quick-start) for those
components. The Go server does not load `.env` or launch the other services.
Pass service addresses and authentication configuration in the MCP client's
environment. Linux hosts can run the MCP server, but Premiere itself requires
macOS or Windows. Installing the package is not evidence of a working Premiere
connection; verify `premiere_ping` after configuring the bridge and panel.

## Download integrity and recovery

Every staged package embeds a `release-manifest.json` containing the version,
byte length, and SHA-256 hash of all four binaries. The release process builds
the binaries before computing those values. The installer checks size and hash
before making a downloaded binary available, writes through a unique temporary
file, and removes partial downloads on failure. Each launch verifies the cached
binary; a modified or missing file is downloaded again. Verified cached binaries
work offline. No Go compiler or npm dependencies are needed on the user's machine.

Downloads require HTTPS access to `github.com`, `release-assets.githubusercontent.com`
and `objects.githubusercontent.com`; redirects to other hosts or HTTP are rejected.
The download has a 60-second total timeout, at most five redirects, and a maximum
binary size of 256 MiB. The checksum is pinned inside the npm tarball rather than
retrieved alongside the binary. It detects a changed GitHub asset, but still
depends on trusting the npm package and its publisher.

If npm lifecycle scripts are disabled, the first launch performs the same verified
installation. Failures exit nonzero and can be retried with:

```sh
npm rebuild -g @premierpro/mcp-server
```

A 404 means that the matching GitHub release asset is missing. A checksum failure
means the release asset differs from the published npm manifest; report it to the
maintainer. Do not replace assets for an already published version: cut a new
version. The wrapper has no alternate-download-URL or checksum-bypass setting.
It uses Node's HTTPS trust and networking configuration; npm registry proxy settings
are not automatically applied to the binary download.

## Build and publish a release

The `Release` workflow runs on `v*` tags. It validates the repository, cross-compiles
four Go binaries with `CGO_ENABLED=0`, stamps version/commit/source timestamp, and stages
a dependency-free npm package. It uploads the binaries, `SHA256SUMS`, and the npm
tarball to the GitHub release. Prerelease versions use the npm `next` tag; stable
versions use `latest`. Tags must be SemVer with a `v` prefix (for example `v1.2.3`
or `v1.2.3-rc.1`) and cannot contain build metadata.

The workflow refuses to overwrite an existing GitHub release. If only npm
publication fails, rerun the failed publish job after fixing its account settings;
do not rebuild or replace an already published version. A failure partway through
GitHub asset upload requires maintainer inspection before choosing a new version.

The `publish-npm` job remains disabled until repository variable
`NPM_PUBLISH_ENABLED` is set to `true`. Before enabling it:

1. Confirm ownership and publishing rights for the `@premierpro` npm scope.
2. Bootstrap the first public package version with an authorized npm account if
   the package does not yet exist. Download the tarball from its GitHub release,
   inspect it, and run `npm publish ./premierpro-mcp-server-VERSION.tgz --access public`
   using the account's required authentication. Use `--tag next` for prereleases.
   Do not run the automated publish job again for that same version.
3. Configure the package's npm trusted publisher for GitHub owner `ayushozha`,
   repository `AdobePremiereProMCP`, workflow `release.yml`, environment `npm-publish`.
   Allow publication in the npm trusted publisher settings.
4. Create the GitHub `npm-publish` environment and configure maintainer approval
   if required by the project's release policy. Enable `NPM_PUBLISH_ENABLED` only
   when publication is authorized.
5. Create the next release tag. The npm job waits for the GitHub release, uses
   GitHub's short-lived OIDC identity, and publishes with provenance. Node 24
   supplies a compatible npm CLI; trusted publishing requires npm 11.5.1+.

See npm's [trusted publishing documentation](https://docs.npmjs.com/trusted-publishers/)
and [public scoped package guidance](https://docs.npmjs.com/creating-and-publishing-scoped-public-packages/).
No npm credentials belong in this repository. A successful GitHub release with
publishing disabled is not an npm publication. After publication, verify the
registry version and install that exact version on the four supported targets.

For local packaging, start with all four correctly named build artifacts:

```sh
node npm-package/scripts/prepare-release.cjs v1.2.3 ./artifacts ./npm-release
cd npm-release
npm pack
```

The staging directory must not already exist. The script copies the repository
MIT license and generates the exact package version and manifest without editing
tracked source files. Packing the source `npm-package/` directly fails because
its manifest deliberately contains no release hashes.

## Verification

```sh
npm test --prefix npm-package
```

Tests use local HTTP streams and executable fixtures: no public release, registry,
credentials, global installation, or Premiere instance is needed. The npm install
test uses a temporary global prefix under `npm-package/.tmp/` and removes it after
the test. It verifies the packed files, postinstall, first-launch fallback, literal
arguments, stdio, exit codes, and Unix signal forwarding. Other tests cover all four
asset mappings, unsupported platforms, manifest validation, corrupted cache repair,
size/hash failures, redirect restrictions, HTTP failures, timeouts, concurrent
installs, executable permissions and cleanup.

Fixture tests cannot establish that a release was published or that the real
Premiere editing workflow works. The release workflow additionally runs the real
Linux Go binary through the repository's MCP stdio smoke test before uploading it.
