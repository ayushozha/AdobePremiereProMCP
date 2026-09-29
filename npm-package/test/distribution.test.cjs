'use strict';

const assert = require('node:assert/strict');
const { createHash } = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');
const http = require('node:http');
const { spawn, spawnSync } = require('node:child_process');
const { test } = require('node:test');

const source = path.resolve(__dirname, '..');
const scratch = path.join(source, '.tmp');
const names = [
  'premierpro-mcp-darwin-arm64',
  'premierpro-mcp-darwin-x64',
  'premierpro-mcp-linux-x64',
  'premierpro-mcp-win-x64.exe',
];
const digest = (bytes) => createHash('sha256').update(bytes).digest('hex');

async function executableFixture() {
  // Homebrew's Node depends on dylibs relative to its installation. A tiny
  // executable fixture forwards to that unchanged runtime on Unix; Windows
  // needs a real PE executable and official Node is self-contained there.
  return process.platform === 'win32' ? fs.readFile(process.execPath) :
    Buffer.from(`#!/bin/sh\nexec '${process.execPath.replaceAll("'", "'\\''")}' "$@"\n`);
}

async function launcherFixture(t, bytes) {
  const { root } = await fixture(t, bytes);
  await fs.cp(path.join(source, 'bin'), path.join(root, 'bin'), { recursive: true });
  await fs.cp(path.join(source, 'lib'), path.join(root, 'lib'), { recursive: true });
  await fs.mkdir(path.join(root, 'native'));
  const name = process.platform === 'win32' ? names[3] : process.platform === 'linux' ? names[2] : process.arch === 'arm64' ? names[0] : names[1];
  await fs.writeFile(path.join(root, 'native', name), bytes, { mode: 0o755 });
  return { root, launcher: path.join(root, 'bin', 'premierpro-mcp.cjs') };
}

async function fixture(t, bytes = Buffer.from('release binary fixture')) {
  await fs.mkdir(scratch, { recursive: true });
  const root = await fs.mkdtemp(path.join(scratch, 'package with spaces-'));
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  await fs.writeFile(path.join(root, 'package.json'), JSON.stringify({ version: '1.2.3' }));
  const assets = Object.fromEntries(names.map((name) => [name, { sha256: digest(bytes), size: bytes.length }]));
  await fs.writeFile(path.join(root, 'release-manifest.json'), JSON.stringify({ version: '1.2.3', assets }));
  return { root, bytes };
}

// Serve actual HTTP streams locally; production URL checks still see the exact
// GitHub URL, and only the transport is substituted at the network boundary.
async function transport(t, handle) {
  const requests = [];
  const server = http.createServer(handle);
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  t.after(() => new Promise((resolve) => { server.close(resolve); server.closeAllConnections(); }));
  return {
    requests,
    fetchImpl: (url, options) => {
      requests.push(String(url));
      return fetch(`http://127.0.0.1:${server.address().port}${new URL(url).pathname}`, options);
    },
  };
}

test('a release can be staged from four binary artifacts', async (t) => {
  const { root, bytes } = await fixture(t);
  const input = path.join(root, 'artifacts');
  const output = path.join(root, 'staged');
  await fs.mkdir(input);
  await Promise.all(names.map((name) => fs.writeFile(path.join(input, name), bytes)));
  const result = spawnSync(process.execPath, [path.join(source, 'scripts/prepare-release.cjs'), 'v1.2.3', input, output], { encoding: 'utf8' });
  assert.equal(result.status, 0, result.stderr);
  const pkg = JSON.parse(await fs.readFile(path.join(output, 'package.json')));
  const manifest = JSON.parse(await fs.readFile(path.join(output, 'release-manifest.json')));
  assert.equal(pkg.version, '1.2.3');
  assert.equal(pkg.name, '@premierpro/mcp-server');
  assert.equal(manifest.version, '1.2.3');
  for (const name of names) assert.deepEqual(manifest.assets[name], { size: bytes.length, sha256: digest(bytes) });
  assert.equal((await fs.readFile(path.join(input, 'SHA256SUMS'), 'utf8')).split('\n').filter(Boolean).length, 4);
});

test('platforms select exact, pinned assets, including Windows .exe', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  for (const [platform, arch, expected] of [
    ['darwin', 'arm64', names[0]], ['darwin', 'x64', names[1]],
    ['linux', 'x64', names[2]], ['win32', 'x64', names[3]],
  ]) {
    const { root, bytes } = await fixture(t);
    const network = await transport(t, (_req, res) => res.end(bytes));
    const binary = await ensureBinary({ packageDir: root, platform, arch, ...network });
    assert.equal(path.basename(binary), expected);
    assert.equal(network.requests[0], `https://github.com/ayushozha/AdobePremiereProMCP/releases/download/v1.2.3/${expected}`);
    assert.deepEqual(await fs.readFile(binary), bytes);
    if (process.platform !== 'win32') assert.equal((await fs.stat(binary)).mode & 0o111, 0o111);
    assert.equal(await ensureBinary({ packageDir: root, platform, arch, ...network }), binary);
    assert.equal(network.requests.length, 1, 'verified installations must work offline without another download');
  }
});

test('unsupported OS and CPU fail before any download', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root } = await fixture(t);
  for (const [platform, arch] of [['linux', 'arm64'], ['win32', 'arm64'], ['freebsd', 'x64'], ['linux', 'ia32']]) {
    await assert.rejects(ensureBinary({ packageDir: root, platform, arch, fetchImpl: () => assert.fail('unexpected network request') }), /Unsupported platform/);
  }
});

test('version mismatch and malformed manifests fail before download', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  for (const change of [
    (m) => { m.version = '1.2.4'; },
    (m) => { m.assets[names[0]].sha256 = 'invalid'; },
    (m) => { m.assets[names[0]].size = 0; },
    (m) => { delete m.assets[names[3]]; },
  ]) {
    const { root } = await fixture(t);
    const file = path.join(root, 'release-manifest.json');
    const manifest = JSON.parse(await fs.readFile(file));
    change(manifest);
    await fs.writeFile(file, JSON.stringify(manifest));
    await assert.rejects(ensureBinary({ packageDir: root, fetchImpl: () => assert.fail('unexpected download') }), /manifest/i);
  }
});

test('checksum mismatch, truncated and oversized downloads never install executable bytes', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  for (const body of [Buffer.from('wrong!!'), Buffer.from('short'), Buffer.from('too long for expected file')]) {
    const { root } = await fixture(t, Buffer.from('correct'));
    const network = await transport(t, (_req, res) => { res.write(body); res.end(); });
    await assert.rejects(ensureBinary({ packageDir: root, platform: 'linux', arch: 'x64', ...network }), /checksum|size/i);
    assert.deepEqual(await fs.readdir(path.join(root, 'native')), []);
  }
});

test('GitHub asset redirects work; HTTP, foreign hosts and redirect loops fail closed', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root, bytes } = await fixture(t);
  const network = await transport(t, (req, res) => {
    if (req.url === '/asset') return res.end(bytes);
    res.writeHead(302, { location: 'https://release-assets.githubusercontent.com/asset' });
    res.end();
  });
  await ensureBinary({ packageDir: root, platform: 'linux', arch: 'x64', ...network });
  assert.equal(network.requests.length, 2);
  for (const location of ['http://github.com/asset', 'https://example.com/asset', 'https://github.com.evil.test/asset', 'https://user:password@github.com/asset', 'https://github.com:444/asset', 'https://github.com/loop']) {
    const { root: other } = await fixture(t);
    const bad = await transport(t, (_req, res) => { res.writeHead(302, { location }); res.end(); });
    await assert.rejects(ensureBinary({ packageDir: other, platform: 'linux', arch: 'x64', ...bad }), /redirect|HTTPS|host|URL/i);
    assert.ok(bad.requests.length <= 6);
  }
});

test('404, connection failures and stalled downloads produce actionable errors and clean up', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root } = await fixture(t);
  const missing = await transport(t, (_req, res) => { res.writeHead(404); res.end('missing'); });
  await assert.rejects(ensureBinary({ packageDir: root, ...missing }), /404/);
  await assert.rejects(ensureBinary({ packageDir: root, fetchImpl: async () => { throw new Error('connection refused'); } }), /connection refused/);
  const stalled = await transport(t, (_req, res) => { res.writeHead(200); res.flushHeaders(); });
  await assert.rejects(ensureBinary({ packageDir: root, timeoutMs: 40, ...stalled }), /timeout|timed out|abort/i);
  assert.deepEqual(await fs.readdir(path.join(root, 'native')), []);
});

test('corrupt cached binaries are replaced and permissions are repaired', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root, bytes } = await fixture(t);
  const network = await transport(t, (_req, res) => res.end(bytes));
  const binary = await ensureBinary({ packageDir: root, platform: 'linux', arch: 'x64', ...network });
  await fs.writeFile(binary, 'tampered');
  await ensureBinary({ packageDir: root, platform: 'linux', arch: 'x64', ...network });
  assert.deepEqual(await fs.readFile(binary), bytes);
  assert.equal(network.requests.length, 2);
  if (process.platform !== 'win32') {
    await fs.chmod(binary, 0o600);
    await ensureBinary({ packageDir: root, platform: 'linux', arch: 'x64', ...network });
    assert.equal((await fs.stat(binary)).mode & 0o111, 0o111);
  }
});

test('a verified executable works in an installation owned by another account', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root, bytes } = await fixture(t);
  const network = await transport(t, (_req, res) => res.end(bytes));
  const binary = await ensureBinary({ packageDir: root, ...network });
  // Emulate an OS denying metadata writes to a root-owned global package, while
  // still allowing the real stat/hash/read path used by ordinary MCP clients.
  const chmod = fs.chmod;
  fs.chmod = async () => { throw Object.assign(new Error('Operation not permitted'), { code: 'EPERM' }); };
  t.after(() => { fs.chmod = chmod; });
  assert.equal(await ensureBinary({ packageDir: root, ...network }), binary);
  assert.equal(network.requests.length, 1);
});

test('a symlinked installation directory cannot redirect downloads', { skip: process.platform === 'win32' }, async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root } = await fixture(t);
  await fs.mkdir(path.join(root, 'other'));
  await fs.symlink(path.join(root, 'other'), path.join(root, 'native'));
  await assert.rejects(ensureBinary({ packageDir: root }), /symbolic|symlink|directory/i);
  assert.deepEqual(await fs.readdir(path.join(root, 'other')), []);
});

test('concurrent first installs leave one verified binary and no partial files', async (t) => {
  const { ensureBinary } = require('../lib/install.cjs');
  const { root, bytes } = await fixture(t);
  const network = await transport(t, (_req, res) => res.end(bytes));
  const results = await Promise.all(Array.from({ length: 4 }, () => ensureBinary({ packageDir: root, ...network })));
  assert.equal(new Set(results).size, 1);
  assert.deepEqual(await fs.readFile(results[0]), bytes);
  assert.deepEqual(await fs.readdir(path.join(root, 'native')), [path.basename(results[0])]);
});

test('launcher reports spawn errors on stderr with a nonzero exit', { skip: process.platform === 'win32' }, async (t) => {
  const { launcher } = await launcherFixture(t, Buffer.from('#!/no/such/interpreter\n'));
  const run = spawnSync(process.execPath, [launcher], { encoding: 'utf8', timeout: 5000 });
  assert.equal(run.status, 1);
  assert.equal(run.stdout, '');
  assert.match(run.stderr, /cannot start.*ENOENT/);
});

test('launcher forwards SIGTERM to the child and preserves the child signal exit', { skip: process.platform === 'win32', timeout: 10000 }, async (t) => {
  const { launcher } = await launcherFixture(t, await executableFixture());
  const child = spawn(process.execPath, [launcher, '-e', 'process.on("SIGTERM", () => {process.stderr.write("forwarded"); process.exit(42)}); process.stdout.write("ready"); setInterval(() => {}, 1000)'], { stdio: ['pipe', 'pipe', 'pipe'] });
  t.after(() => { child.kill('SIGKILL'); });
  const done = new Promise((resolve) => child.once('close', (code, signal) => resolve({ code, signal })));
  let errors = '';
  child.stderr.on('data', (data) => { errors += data; });
  await new Promise((resolve, reject) => {
    child.once('error', reject);
    child.once('exit', () => reject(new Error('Child exited before ready')));
    child.stdout.once('data', (data) => { assert.equal(data.toString(), 'ready'); resolve(); });
  });
  child.kill('SIGTERM');
  assert.deepEqual(await done, { code: 42, signal: null });
  assert.equal(errors, 'forwarded');
  const signaled = spawnSync(process.execPath, [launcher, '-e', 'process.kill(process.pid, "SIGTERM")'], { encoding: 'utf8', timeout: 5000 });
  assert.equal(signaled.signal, 'SIGTERM');
  assert.equal(signaled.stdout, '');
});

test('source checkout cannot be packed with missing release integrity data', async () => {
  const result = spawnSync(process.execPath, [path.join(source, 'scripts/check-release.cjs')], { encoding: 'utf8' });
  assert.equal(result.status, 1);
  assert.match(result.stderr, /manifest/);
});

test('release staging rejects unsafe versions and missing artifacts', async (t) => {
  const { root } = await fixture(t);
  for (const tag of ['latest', 'v01.2.3', 'v1.2.3/../other', 'v1.2.3;touch bad', 'v1.2.3+build', 'v1.2.3-01']) {
    const result = spawnSync(process.execPath, [path.join(source, 'scripts/prepare-release.cjs'), tag, root, path.join(root, 'out')], { encoding: 'utf8' });
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /version|tag/i);
  }
  const missing = spawnSync(process.execPath, [path.join(source, 'scripts/prepare-release.cjs'), 'v1.2.3-rc.1', root, path.join(root, 'out')], { encoding: 'utf8' });
  assert.notEqual(missing.status, 0);
  assert.match(missing.stderr, /ENOENT|missing|artifact/i);
});

test('packed global install, postinstall and launcher preserve arguments, stdio and exit code offline', { timeout: 60000 }, async (t) => {
  const { root } = await fixture(t);
  const input = path.join(root, 'assets');
  const staged = path.join(root, 'staged');
  await fs.mkdir(input);
  const bytes = await executableFixture();
  await Promise.all(names.map((name) => fs.writeFile(path.join(input, name), bytes)));
  const prepared = spawnSync(process.execPath, [path.join(source, 'scripts/prepare-release.cjs'), 'v1.2.3', input, staged], { encoding: 'utf8' });
  assert.equal(prepared.status, 0, prepared.stderr);
  const npmCli = process.env.npm_execpath;
  assert.ok(npmCli, 'run these integration tests using npm test');
  const runNpm = (args, extra = {}) => spawnSync(process.execPath, [npmCli, ...args], {
    cwd: root, encoding: 'utf8', timeout: 30000,
    env: { ...process.env, npm_config_cache: path.join(root, 'cache'), npm_config_update_notifier: 'false', ...extra },
  });
  const packed = runNpm(['pack', staged, '--json', '--offline']);
  assert.equal(packed.status, 0, packed.stderr);
  const pack = JSON.parse(packed.stdout)[0];
  const packedPaths = pack.files.map((file) => file.path);
  assert.ok(packedPaths.includes('release-manifest.json'));
  assert.ok(packedPaths.includes('LICENSE'));
  assert.ok(!packedPaths.some((file) => /^(test|native|\.tmp)\//.test(file)));
  assert.ok(!packedPaths.includes('scripts/prepare-release.cjs'));
  const preload = path.join(root, 'offline-network.cjs');
  await fs.writeFile(preload, `const fs = require('node:fs');\nglobal.fetch = async (url) => {\n  if (!/^https:\\/\\/github\\.com\\/ayushozha\\/AdobePremiereProMCP\\/releases\\/download\\/v1\\.2\\.3\\/premierpro-mcp-(darwin-arm64|darwin-x64|linux-x64|win-x64\\.exe)$/.test(String(url))) throw new Error('Unexpected download URL: ' + url);\n  return new Response(fs.readFileSync(${JSON.stringify(path.join(input, names[0]))}));\n};\n`);
  const prefix = path.join(root, 'global prefix');
  const env = { ...process.env, NODE_OPTIONS: `--require=${JSON.stringify(preload)}` };
  const installed = runNpm(['install', '--global', '--prefix', prefix, path.join(root, pack.filename), '--offline', '--no-audit', '--no-fund', '--foreground-scripts'], env);
  assert.equal(installed.status, 0, installed.stderr);
  const packageDir = process.platform === 'win32' ? path.join(prefix, 'node_modules', '@premierpro', 'mcp-server') : path.join(prefix, 'lib', 'node_modules', '@premierpro', 'mcp-server');
  assert.equal((await fs.readdir(path.join(packageDir, 'native'))).length, 1, 'postinstall must fetch the binary');
  const bin = process.platform === 'win32' ? path.join(packageDir, 'bin', 'premierpro-mcp.cjs') : path.join(prefix, 'bin', 'premierpro-mcp');
  const dangerous = ['spaces in arguments', '$(touch injected)', '; touch injected', '"quotes"'];
  const script = 'process.stdin.pipe(process.stdout); process.stderr.write(JSON.stringify(process.argv.slice(1))); process.stdin.on("end", () => { process.exitCode=23 });';
  const run = spawnSync(process.execPath, [bin, '-e', script, '--', ...dangerous], { cwd: root, input: '{"jsonrpc":"2.0","id":1}\n', encoding: 'utf8', timeout: 10000 });
  assert.equal(run.status, 23, run.stderr);
  assert.equal(run.stdout, '{"jsonrpc":"2.0","id":1}\n');
  assert.deepEqual(JSON.parse(run.stderr), dangerous);
  await assert.rejects(fs.stat(path.join(root, 'injected')), { code: 'ENOENT' });
  // npm --ignore-scripts users get the same verified install on first launch.
  await fs.rm(path.join(packageDir, 'native'), { recursive: true });
  const fallback = spawnSync(process.execPath, [bin, '-e', 'process.stdout.write("ready")'], { env, encoding: 'utf8', timeout: 10000 });
  assert.equal(fallback.status, 0, fallback.stderr);
  assert.equal(fallback.stdout, 'ready');
});
