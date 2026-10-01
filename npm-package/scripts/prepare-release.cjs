'use strict';

const fs = require('node:fs/promises');
const path = require('node:path');
const { createHash } = require('node:crypto');
const { assets, maxBinaryBytes, readRelease, versionFromTag } = require('../lib/release.cjs');

async function prepare(tag, artifactPath, outputPath) {
  const version = versionFromTag(tag);
  if (!artifactPath || !outputPath) throw new Error('Usage: node scripts/prepare-release.cjs vVERSION ARTIFACT_DIR NEW_PACKAGE_DIR');
  const source = path.resolve(__dirname, '..');
  const artifacts = path.resolve(artifactPath);
  const output = path.resolve(outputPath);
  const manifest = { version, assets: {} };
  for (const name of Object.values(assets)) {
    const file = path.join(artifacts, name);
    const stat = await fs.lstat(file);
    if (!stat.isFile() || stat.size < 1 || stat.size > maxBinaryBytes) throw new Error(`Invalid artifact: ${name}`);
    const bytes = await fs.readFile(file);
    manifest.assets[name] = { size: bytes.length, sha256: createHash('sha256').update(bytes).digest('hex') };
  }
  // Never overwrite a source directory or an earlier staging attempt.
  await fs.mkdir(output);
  for (const file of ['bin', 'lib', 'scripts/install.cjs', 'scripts/check-release.cjs', 'README.md']) {
    await fs.cp(path.join(source, file), path.join(output, file), { recursive: true });
  }
  await fs.copyFile(path.join(source, '..', 'LICENSE'), path.join(output, 'LICENSE'));
  const pkg = JSON.parse(await fs.readFile(path.join(source, 'package.json'), 'utf8'));
  pkg.version = version;
  delete pkg.scripts.test;
  await fs.writeFile(path.join(output, 'package.json'), `${JSON.stringify(pkg, null, 2)}\n`);
  await fs.writeFile(path.join(output, 'release-manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`);
  await readRelease(output);
  await fs.writeFile(path.join(artifacts, 'SHA256SUMS'), Object.entries(manifest.assets).map(([name, asset]) => `${asset.sha256}  ${name}\n`).join(''));
}

prepare(...process.argv.slice(2)).catch((error) => {
  process.stderr.write(`Release preparation failed: ${error.message}\n`);
  process.exitCode = 1;
});
