/**
 * Timecode utility functions for frame-accurate timeline operations.
 *
 * All arithmetic is performed in integer frame counts to avoid
 * floating-point drift that would accumulate over long sequences.
 */

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** Frame-accurate timecode representation (HH:MM:SS:FF). */
export interface Timecode {
  hours: number;
  minutes: number;
  seconds: number;
  frames: number;
  frameRate: number;
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

function assertValidFrameRate(fps: number): void {
  if (fps <= 0 || !Number.isFinite(fps)) {
    throw new RangeError(`Invalid frame rate: ${fps}. Must be a positive finite number.`);
  }
}

function assertValidTimecode(tc: Timecode): void {
  assertValidFrameRate(tc.frameRate);

  const maxFrames = Math.ceil(tc.frameRate) - 1;
  if (tc.frames < 0 || tc.frames > maxFrames) {
    throw new RangeError(
      `Frames value ${tc.frames} out of range for ${tc.frameRate} fps (0-${maxFrames}).`,
    );
  }
  if (tc.seconds < 0 || tc.seconds > 59) {
    throw new RangeError(`Seconds value ${tc.seconds} out of range (0-59).`);
  }
  if (tc.minutes < 0 || tc.minutes > 59) {
    throw new RangeError(`Minutes value ${tc.minutes} out of range (0-59).`);
  }
  if (tc.hours < 0) {
    throw new RangeError(`Hours value ${tc.hours} must be non-negative.`);
  }
}

// ---------------------------------------------------------------------------
// Conversion: Timecode <-> seconds
// ---------------------------------------------------------------------------

/**
 * Convert a timecode to an absolute number of seconds.
 *
 * The conversion goes through integer frames first so that the result
 * is as precise as the underlying frame grid allows.
 */
export function timecodeToSeconds(tc: Timecode): number {
  assertValidTimecode(tc);
  const totalFrames = timecodeToFrames(tc);
  return totalFrames / tc.frameRate;
}

/**
 * Convert an absolute number of seconds to a timecode at the given frame rate.
 *
 * The value is rounded to the nearest frame boundary.
 */
export function secondsToTimecode(seconds: number, fps: number): Timecode {
  if (seconds < 0) {
    throw new RangeError(`Seconds value ${seconds} must be non-negative.`);
  }
  assertValidFrameRate(fps);

  const totalFrames = Math.round(seconds * fps);
  return framesToTimecode(totalFrames, fps);
}

// ---------------------------------------------------------------------------
// Conversion: Timecode <-> frames
// ---------------------------------------------------------------------------

/** Convert a timecode to an absolute frame count. */
export function timecodeToFrames(tc: Timecode): number {
  assertValidTimecode(tc);

  const framesPerSecond = Math.ceil(tc.frameRate);
  const totalSeconds = tc.hours * 3600 + tc.minutes * 60 + tc.seconds;
  return totalSeconds * framesPerSecond + tc.frames;
}

/** Convert an absolute frame count to a timecode at the given frame rate. */
export function framesToTimecode(frames: number, fps: number): Timecode {
  if (frames < 0) {
    throw new RangeError(`Frame count ${frames} must be non-negative.`);
  }
  assertValidFrameRate(fps);

  const framesPerSecond = Math.ceil(fps);

  let remaining = Math.round(frames);

  const ff = remaining % framesPerSecond;
  remaining = (remaining - ff) / framesPerSecond;

  const ss = remaining % 60;
  remaining = (remaining - ss) / 60;

  const mm = remaining % 60;
  const hh = (remaining - mm) / 60;

  return {
    hours: hh,
    minutes: mm,
    seconds: ss,
    frames: ff,
    frameRate: fps,
  };
}

// ---------------------------------------------------------------------------
// Arithmetic
// ---------------------------------------------------------------------------

/**
 * Add two timecodes.
 *
 * Both timecodes must share the same frame rate. The result uses that
 * frame rate.
 */
export function addTimecodes(a: Timecode, b: Timecode): Timecode {
  assertMatchingFrameRates(a, b);
  const totalFrames = timecodeToFrames(a) + timecodeToFrames(b);
  return framesToTimecode(totalFrames, a.frameRate);
}

/**
 * Subtract timecode `b` from timecode `a` (`a - b`).
 *
 * Throws if the result would be negative.
 */
export function subtractTimecodes(a: Timecode, b: Timecode): Timecode {
  assertMatchingFrameRates(a, b);
  const framesA = timecodeToFrames(a);
  const framesB = timecodeToFrames(b);

  if (framesB > framesA) {
    throw new RangeError(
      `Cannot subtract ${formatTimecode(b)} from ${formatTimecode(a)}: result would be negative.`,
    );
  }

  return framesToTimecode(framesA - framesB, a.frameRate);
}

function assertMatchingFrameRates(a: Timecode, b: Timecode): void {
  if (a.frameRate !== b.frameRate) {
    throw new Error(
      `Frame rate mismatch: ${a.frameRate} fps vs ${b.frameRate} fps. ` +
        `Convert to a common frame rate before performing arithmetic.`,
    );
  }
}

// ---------------------------------------------------------------------------
// Formatting & parsing
// ---------------------------------------------------------------------------

/**
 * Format a timecode as the standard "HH:MM:SS:FF" string.
 *
 * Each component is zero-padded to two digits (frames may use more digits
 * for frame rates above 99 fps, though that is extremely rare).
 */
export function formatTimecode(tc: Timecode): string {
  const pad = (n: number, width = 2): string => String(n).padStart(width, "0");
  return `${pad(tc.hours)}:${pad(tc.minutes)}:${pad(tc.seconds)}:${pad(tc.frames)}`;
}

/**
 * Parse a "HH:MM:SS:FF" string into a Timecode.
 *
 * @param str  The timecode string. Accepts colon or semicolon separators.
 * @param fps  The frame rate to attach to the parsed timecode.
 */
export function parseTimecode(str: string, fps: number): Timecode {
  assertValidFrameRate(fps);

  // Accept both ':' and ';' as separators (semicolons are common in drop-frame notation).
  const normalized = str.trim().replace(/;/g, ":");
  const parts = normalized.split(":");

  if (parts.length !== 4) {
    throw new Error(
      `Invalid timecode format "${str}". Expected "HH:MM:SS:FF".`,
    );
  }

  const [hh, mm, ss, ff] = parts.map((p) => {
    const n = Number(p);
    if (!Number.isFinite(n) || n < 0 || !Number.isInteger(n)) {
      throw new Error(
        `Invalid timecode component "${p}" in "${str}". ` +
          `All components must be non-negative integers.`,
      );
    }
    return n;
  });

  const tc: Timecode = {
    hours: hh,
    minutes: mm,
    seconds: ss,
    frames: ff,
    frameRate: fps,
  };

  // Validate the constructed timecode so callers get a clear error for
  // out-of-range values (e.g. 70 seconds).
  assertValidTimecode(tc);

  return tc;
}

// ---------------------------------------------------------------------------
// Comparison helpers
// ---------------------------------------------------------------------------

/** Return true if timecodes `a` and `b` represent the same point in time. */
export function timecodesEqual(a: Timecode, b: Timecode): boolean {
  assertMatchingFrameRates(a, b);
  return timecodeToFrames(a) === timecodeToFrames(b);
}

/**
 * Compare two timecodes. Returns a negative number if `a < b`, zero if
 * they are equal, and a positive number if `a > b`.
 */
export function compareTimecodes(a: Timecode, b: Timecode): number {
  assertMatchingFrameRates(a, b);
  return timecodeToFrames(a) - timecodeToFrames(b);
}

/** Create a zero-valued timecode at the given frame rate. */
export function zeroTimecode(fps: number): Timecode {
  assertValidFrameRate(fps);
  return { hours: 0, minutes: 0, seconds: 0, frames: 0, frameRate: fps };
}
