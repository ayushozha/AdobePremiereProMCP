'use strict';

const fs = require('node:fs/promises');
const { createReadStream, createWriteStream } = require('node:fs');
const { createHash, randomUUID } = require('node:crypto');
const path = require('node:path');
const { Transform, Readable } = require('node:stream');
const { pipeline } = require('node:stream/promises');
const { assets, readRelease } = require('./release.cjs');

const allowedHosts = new Set(['github.com', 'release-assets.githubusercontent.com', 'objects.githubusercontent.com']);

function validateURL(value) {
  const url = new URL(value);
  if (url.protocol !== 'https:' || url.username || url.password || url.port || !allowedHosts.has(url.hostname)) {
    throw new Error('Refusing download URL: only HTTPS GitHub release hosts without credentials or custom ports are allowed.');
  }
  return url;
}

async function matches(file, asset) {
  try {
    const stat = await fs.lstat(file);
    if (!stat.isFile() || stat.size !== asset.size) return false;
    const hash = createHash('sha256');
    for await (const chunk of createReadStream(file)) hash.update(chunk);
    return hash.digest('hex') === asset.sha256;
  } catch (error) {
    if (error.code === 'ENOENT') return false;
    throw error;
  }
}

async function download(url, file, asset, fetchImpl, timeoutMs) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(new Error('Release download timed out.')), timeoutMs);
  try {
    let response;
    for (let redirects = 0; ; redirects++) {
      response = await fetchImpl(validateURL(url), {
        redirect: 'manual', signal: controller.signal,
        headers: { 'User-Agent': '@premierpro/mcp-server', 'Accept-Encoding': 'identity' },
      });
      if (![301, 302, 303, 307, 308].includes(response.status)) break;
      await response.body?.cancel();
      const location = response.headers.get('location');
      if (!location || redirects >= 5) throw new Error('Invalid or excessive release download redirects.');
      url = validateURL(new URL(location, url));
    }
    if (response.status !== 200) {
      await response.body?.cancel();
      throw new Error(`Release download returned HTTP ${response.status}. Check that this version has published GitHub assets.`);
    }
    const length = response.headers.get('content-length');
    if (length !== null && Number(length) !== asset.size) {
      await response.body?.cancel();
      throw new Error(`Release binary size mismatch: expected ${asset.size} bytes, received Content-Length ${length}.`);
    }
    if (!response.body) throw new Error('Empty release download response.');
    const hash = createHash('sha256');
    let received = 0;
    const verify = new Transform({
      transform(chunk, _encoding, callback) {
        received += chunk.length;
        if (received > asset.size) return callback(new Error('Release binary exceeds expected size.'));
        hash.update(chunk);
        callback(null, chunk);
      },
    });
    await pipeline(Readable.fromWeb(response.body), verify, createWriteStream(file, { flags: 'wx', mode: 0o700 }), { signal: controller.signal });
    if (received !== asset.size) throw new Error(`Release binary size mismatch: expected ${asset.size}, received ${received}.`);
    if (hash.digest('hex') !== asset.sha256) throw new Error('Release binary SHA-256 checksum mismatch. Refusing to execute it.');
  } finally {
    clearTimeout(timer);
  }
}

async function ensureBinary({
  packageDir = path.resolve(__dirname, '..'), platform = process.platform, arch = process.arch,
  fetchImpl = globalThis.fetch, timeoutMs = 60000,
} = {}) {
  const name = assets[`${platform}-${arch}`];
  if (!name) throw new Error(`Unsupported platform ${platform}-${arch}. Supported: macOS arm64/x64, Linux x64, Windows x64.`);
  const manifest = await readRelease(packageDir);
  const asset = manifest.assets[name];
  const directory = path.join(packageDir, 'native');
  await fs.mkdir(directory, { recursive: true });
  if (!(await fs.lstat(directory)).isDirectory()) throw new Error('Native binary directory must be a real directory, not a symbolic link.');
  const binary = path.join(directory, name);
  if (await matches(binary, asset)) {
    if (process.platform !== 'win32' && ((await fs.stat(binary)).mode & 0o111) !== 0o111) {
      await fs.chmod(binary, 0o755);
    }
    return binary;
  }
  const temporary = path.join(directory, `.${name}-${randomUUID()}.tmp`);
  try {
    const url = `https://github.com/ayushozha/AdobePremiereProMCP/releases/download/v${manifest.version}/${name}`;
    await download(url, temporary, asset, fetchImpl, timeoutMs);
    await fs.chmod(temporary, 0o755);
    try {
      await fs.rename(temporary, binary);
    } catch (error) {
      // Windows may refuse to replace a binary another concurrent launcher has
      // already installed/opened. Accept it only after independently hashing it.
      if (!(await matches(binary, asset))) throw error;
    }
    return binary;
  } catch (error) {
    throw new Error(`Unable to install ${name} for v${manifest.version}: ${error.message} Retry with npm rebuild @premierpro/mcp-server.`, { cause: error });
  } finally {
    await fs.rm(temporary, { force: true });
  }
}

module.exports = { ensureBinary };
