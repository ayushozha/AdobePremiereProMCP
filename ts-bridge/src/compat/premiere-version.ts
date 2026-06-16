/**
 * Parse `app.version` from Premiere (via ping / health) for PP24 vs PP25 code paths.
 */

export interface ParsedPremiereVersion {
  major: number;
  minor: number;
}

/** Parses Adobe-style version strings: "25.0", "24.6.2", "25". */
export function parsePremiereVersionString(version: string): ParsedPremiereVersion | null {
  const v = version.trim();
  if (!v) return null;
  const m = /^(\d+)(?:\.(\d+))?/.exec(v);
  if (!m) return null;
  return {
    major: parseInt(m[1], 10),
    minor: m[2] !== undefined ? parseInt(m[2], 10) : 0,
  };
}

/** True if version >= minMajor.minMinor (major beats minor). */
export function premiereVersionAtLeast(
  version: string,
  minMajor: number,
  minMinor = 0,
): boolean {
  const p = parsePremiereVersionString(version);
  if (!p) return false;
  if (p.major > minMajor) return true;
  if (p.major < minMajor) return false;
  return p.minor >= minMinor;
}
