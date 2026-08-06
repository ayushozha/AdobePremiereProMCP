import { existsSync, readFileSync } from "node:fs";
import * as path from "node:path";

/** Parse the small KEY=VALUE subset used by this repository's .env file. */
export function parseDotEnv(contents: string): Array<[string, string]> {
  const values: Array<[string, string]> = [];
  for (const line of contents.split(/\r?\n/)) {
    const match = line.match(/^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*?)\s*$/);
    if (!match) continue;
    let value = match[2] ?? "";
    if (
      value.length >= 2 &&
      ((value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'")))
    ) {
      value = value.slice(1, -1);
    }
    values.push([match[1] as string, value]);
  }
  return values;
}

/**
 * Load the repository .env before authentication. Existing process values win,
 * matching normal dotenv behavior and avoiding launcher-specific auth rules.
 */
export function loadRepositoryEnvironment(
  repositoryRoot: string,
  environment: NodeJS.ProcessEnv = process.env,
): void {
  const dotenvPath = path.join(repositoryRoot, ".env");
  if (!existsSync(dotenvPath)) return;
  for (const [key, value] of parseDotEnv(readFileSync(dotenvPath, "utf-8"))) {
    if (environment[key] === undefined) environment[key] = value;
  }
}
