'use strict';

require('../lib/release.cjs').readRelease(require('node:path').resolve(__dirname, '..')).catch((error) => {
  process.stderr.write(`Cannot pack an unprepared release: ${error.message}\n`);
  process.exitCode = 1;
});
