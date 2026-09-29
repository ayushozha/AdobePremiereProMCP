'use strict';

require('../lib/install.cjs').ensureBinary().catch((error) => {
  process.stderr.write(`premierpro-mcp: ${error.message}\n`);
  process.exitCode = 1;
});
