'use strict';

const fs = require('node:fs/promises');
const path = require('node:path');

const assets = Object.freeze({
  'darwin-arm64': 'premierpro-mcp-darwin-arm64',
  'darwin-x64': 'premierpro-mcp-darwin-x64',
  'linux-x64': 'premierpro-mcp-linux-x64',
  'win32-x64': 'premierpro-mcp-win-x64.exe',
});
const maxBinaryBytes = 256 * 1024 * 1024;

function validateVersion(version) {
  // No build metadata: npm normalizes it away, which would break tag pinning.
  const numeric = '(0|[1-9][0-9]*)';
  const identifier = '(?:0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)';
  if (typeof version !== 'string' || version.length > 128 ||
      !new RegExp(`^${numeric}\\.${numeric}\\.${numeric}(?:-${identifier}(?:\\.${identifier})*)?$`).test(version)) {
    throw new Error('Invalid release version; use a SemVer tag such as v1.2.3 or v1.2.3-rc.1 (without build metadata).');
  }
  return version;
}

function versionFromTag(tag) {
  if (typeof tag !== 'string' || !tag.startsWith('v')) throw new Error('Release tag must start with v.');
  return validateVersion(tag.slice(1));
}

async function readRelease(packageDir) {
  const pkg = JSON.parse(await fs.readFile(path.join(packageDir, 'package.json'), 'utf8'));
  const manifest = JSON.parse(await fs.readFile(path.join(packageDir, 'release-manifest.json'), 'utf8'));
  validateVersion(pkg.version);
  if (manifest.version !== pkg.version || !manifest.assets ||
      Object.keys(manifest.assets).length !== Object.keys(assets).length) {
    throw new Error('Invalid release manifest. Install a published package or stage a release with scripts/prepare-release.cjs.');
  }
  for (const name of Object.values(assets)) {
    const asset = manifest.assets[name];
    if (!asset || !/^[a-f0-9]{64}$/.test(asset.sha256) || !Number.isSafeInteger(asset.size) ||
        asset.size < 1 || asset.size > maxBinaryBytes) {
      throw new Error(`Invalid release manifest entry for ${name}.`);
    }
  }
  return manifest;
}

module.exports = { assets, maxBinaryBytes, readRelease, validateVersion, versionFromTag };
