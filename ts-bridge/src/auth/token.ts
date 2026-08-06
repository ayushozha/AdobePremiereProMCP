import { randomBytes } from "node:crypto";
import {
  chmodSync,
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import { homedir } from "node:os";
import { dirname, join } from "node:path";

const MIN_TOKEN_LENGTH = 32;

/** Resolve the shared Go, TypeScript, and CEP authentication token. */
export function resolveSharedToken(configuredToken?: string): string {
  const direct = configuredToken?.trim();
  if (direct) return validateToken(direct, "configured token");

  const configuredPath = process.env["PREMIERE_MCP_TOKEN_FILE"];
  const tokenPath = configuredPath ?? join(homedir(), ".premierpro-mcp", "cep-token");
  if (existsSync(tokenPath)) {
    const existing = readFileSync(tokenPath, "utf-8").trim();
    if (existing) {
      hardenPermissions(tokenPath);
      return validateToken(existing, tokenPath);
    }
  }

  const tokenDir = dirname(tokenPath);
  const directoryExisted = existsSync(tokenDir);
  mkdirSync(tokenDir, { recursive: true, mode: 0o700 });
  if (configuredPath === undefined || !directoryExisted) {
    hardenPermissions(tokenDir, true);
  }
  const generated = randomBytes(32).toString("hex");
  try {
    writeFileSync(tokenPath, `${generated}\n`, {
      encoding: "utf-8",
      flag: "wx",
      mode: 0o600,
    });
    return generated;
  } catch (error) {
    // Another local component may have won the first-start race.
    if (existsSync(tokenPath)) {
      const existing = readFileSync(tokenPath, "utf-8").trim();
      if (existing) {
        hardenPermissions(tokenPath);
        return validateToken(existing, tokenPath);
      }
    }
    throw error;
  }
}

function validateToken(token: string, source: string): string {
  if (token.length < MIN_TOKEN_LENGTH) {
    throw new Error(`${source} must contain at least ${MIN_TOKEN_LENGTH} characters`);
  }
  return token;
}

function hardenPermissions(targetPath: string, directory = false): void {
  if (process.platform !== "win32") {
    chmodSync(targetPath, directory ? 0o700 : 0o600);
  }
}
