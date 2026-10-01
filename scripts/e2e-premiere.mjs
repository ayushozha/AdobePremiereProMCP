#!/usr/bin/env node
/** Opt-in application test. Unit tests inject the external MCP boundary only. */
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { resolve, dirname, join, isAbsolute } from 'node:path';
import { existsSync, lstatSync, realpathSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import os from 'node:os';

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const PURPOSE = 'premiere-mcp-disposable-e2e';
const CHECKS = [
  ['import_media', 'Import video and audio'],
  ['sequence_placement', 'Create sequence and place clips'],
  ['transitions_effects', 'Apply transition and video effect'],
  ['lumetri', 'Set and read back Lumetri exposure'],
  ['audio_levels', 'Set and read back audio level'],
  ['h264_export', 'Export and decode H.264 media'],
  ['frame_capture', 'Capture base64 frame'],
  ['script_edl_timeline', 'Script -> EDL -> timeline'],
];
const fail = message => { throw new Error(message); };
const requireThat = (condition, message) => { if (!condition) fail(message); };
class Blocked extends Error {}
class SafetyStop extends Blocked {}
const sha256 = bytes => createHash('sha256').update(bytes).digest('hex');
const json = file => JSON.parse(readFileSync(file, 'utf8'));
const regular = file => requireThat(lstatSync(file).isFile() && !lstatSync(file).isSymbolicLink(), `Expected regular file, not symlink: ${file}`);

export function parseArgs(args) {
  const options = { mutate: false, timeoutMs: 30000, server: process.env.MCP_SERVER_BIN || join(ROOT, 'go-orchestrator/bin/premierpro-mcp'), sdkRoot: join(ROOT, 'cli'), effect: 'Gaussian Blur', transition: 'Cross Dissolve' };
  const values = { '--prepare': 'prepare', '--fixture-dir': 'fixtureDir', '--project': 'project', '--confirm-disposable': 'confirmation', '--report': 'report', '--server': 'server', '--sdk-root': 'sdkRoot', '--preset': 'preset', '--effect': 'effect', '--transition': 'transition', '--locale': 'locale', '--timeout-ms': 'timeoutMs' };
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--mutate') options.mutate = true;
    else if (args[i] === '--help') options.help = true;
    else if (values[args[i]]) {
      requireThat(args[i + 1] && !args[i + 1].startsWith('--'), `Missing value for ${args[i]}`);
      options[values[args[i]]] = args[++i];
    } else fail(`Unknown option: ${args[i]}`);
  }
  requireThat(!(options.prepare && options.mutate), '--prepare and --mutate cannot be combined');
  options.timeoutMs = Number(options.timeoutMs);
  requireThat(Number.isFinite(options.timeoutMs) && options.timeoutMs >= 100 && options.timeoutMs <= 600000, '--timeout-ms must be 100..600000');
  if (options.mutate) {
    requireThat(options.fixtureDir, '--mutate requires --fixture-dir');
    requireThat(options.project && isAbsolute(options.project), '--mutate requires an absolute --project');
    requireThat(options.confirmation === 'MCP-E2E-DISPOSABLE', '--mutate requires --confirm-disposable MCP-E2E-DISPOSABLE');
  }
  return options;
}

export function decodeResult(result) {
  requireThat(!result.isError, result.content?.filter(c => c.type === 'text').map(c => c.text).join('\n') || 'MCP tool returned isError');
  const text = result.content?.find(c => c.type === 'text')?.text;
  let value = result.structuredContent ?? (text ? JSON.parse(text) : {});
  for (let depth = 0; depth < 6; depth++) {
    requireThat(value && typeof value === 'object', 'MCP tool returned no structured readback');
    requireThat(value.success !== false && value.status !== 'failed' && !value.error, value.error || value.message || 'Native operation failed');
    if (typeof value.message === 'string' && /^[\[{]/.test(value.message.trim())) value = JSON.parse(value.message);
    else if (value.success === true && value.data !== undefined) value = value.data;
    else return value;
  }
  fail('MCP response has too many nested envelopes');
}

function command(binary, args) {
  try { return execFileSync(binary, args, { encoding: 'utf8', timeout: 30000, maxBuffer: 8 * 1024 * 1024, stdio: ['ignore', 'pipe', 'pipe'] }); }
  catch (error) {
    if (error.code === 'ENOENT') throw new Blocked(`Required media executable not found: ${binary}`);
    throw error;
  }
}
function probe(file) {
  return JSON.parse(command(process.env.FFPROBE || 'ffprobe', ['-v', 'error', '-show_streams', '-show_format', '-of', 'json', file]));
}
function decodeMedia(file) {
  command(process.env.FFMPEG || 'ffmpeg', ['-v', 'error', '-xerror', '-i', file, '-f', 'null', '-']);
}
export function prepareFixtures(directory) {
  mkdirSync(resolve(directory));
  const root = realpathSync(directory); // Exclusive: never overwrite a previous run or project.
  mkdirSync(join(root, 'media'));
  const ffmpeg = process.env.FFMPEG || 'ffmpeg';
  const definitions = [
    ['media/e2e_test_pattern.mp4', ['-f', 'lavfi', '-i', 'testsrc2=size=320x180:rate=24:duration=8', '-an', '-c:v', 'libx264', '-threads', '1', '-pix_fmt', 'yuv420p', '-map_metadata', '-1', '-fflags', '+bitexact', '-flags:v', '+bitexact']],
    ['media/e2e_tone.wav', ['-f', 'lavfi', '-i', 'sine=frequency=440:sample_rate=48000:duration=8', '-c:a', 'pcm_s16le', '-map_metadata', '-1', '-fflags', '+bitexact']],
  ];
  const files = {};
  for (const [name, args] of definitions) {
    const file = join(root, name);
    command(ffmpeg, ['-nostdin', '-n', '-v', 'error', ...args, file]);
    decodeMedia(file);
    files[name] = { sha256: sha256(readFileSync(file)), probe: probe(file) };
  }
  writeFileSync(join(root, 'script.txt'), 'B-ROLL: "e2e_test_pattern.mp4"\n', { flag: 'wx' });
  files['script.txt'] = { sha256: sha256(readFileSync(join(root, 'script.txt'))) };
  const manifest = { schema: 1, purpose: PURPOSE, project: join(root, 'MCP-E2E-disposable.prproj'), sequence: 'MCP-E2E-fixture', files, ffmpeg: command(ffmpeg, ['-version']).split('\n')[0] };
  writeFileSync(join(root, 'manifest.json'), JSON.stringify(manifest, null, 2) + '\n', { flag: 'wx' });
  return manifest;
}

export function validateFixture(directory, project) {
  const root = realpathSync(directory);
  requireThat(root === resolve(directory), 'Fixture directory must not contain symlinks');
  regular(join(root, 'manifest.json'));
  const manifest = json(join(root, 'manifest.json'));
  requireThat(manifest.schema === 1 && manifest.purpose === PURPOSE, 'Invalid disposable fixture manifest');
  const expected = join(root, 'MCP-E2E-disposable.prproj');
  requireThat(project === expected && manifest.project === expected, 'Exact disposable project path does not match manifest');
  regular(expected);
  requireThat(realpathSync(expected) === expected, 'Disposable project path must not contain symlinks');
  requireThat(manifest.sequence === 'MCP-E2E-fixture', 'Invalid disposable sequence name');
  for (const name of ['media/e2e_test_pattern.mp4', 'media/e2e_tone.wav', 'script.txt']) {
    const file = join(root, name);
    regular(file);
    requireThat(realpathSync(file) === file, 'Fixture asset path must not contain symlinks');
    requireThat(manifest.files?.[name]?.sha256 === sha256(readFileSync(file)), `Fixture hash mismatch: ${name}`);
  }
  requireThat(!existsSync(join(root, 'artifacts')), 'artifacts already exists; prepare a fresh disposable fixture directory for each run');
  return { ...manifest, root, video: join(root, 'media/e2e_test_pattern.mp4'), audio: join(root, 'media/e2e_tone.wav'), script: join(root, 'script.txt'), artifacts: join(root, 'artifacts') };
}

const seconds = t => {
  requireThat(t && t.frame_rate > 0, 'Missing timecode readback');
  return t.hours * 3600 + t.minutes * 60 + t.seconds + t.frames / t.frame_rate;
};
function near(actual, expected, label, tolerance = 1 / 24 + 0.001) {
  requireThat(Number.isFinite(actual) && Math.abs(actual - expected) <= tolerance, `${label}: expected ${expected}, read back ${actual}`);
}
export function verifyTimeline(timeline, sequence, video, audio) {
  requireThat(timeline.sequence_id === sequence, 'Unexpected active sequence ID');
  const videos = timeline.video_tracks?.find(t => t.index === 0)?.clips || [];
  const audios = timeline.audio_tracks?.find(t => t.index === 0)?.clips || [];
  requireThat(videos.length === 2 && audios.length === 1, 'Expected two video clips and one audio clip');
  for (const [index, clip] of videos.entries()) {
    requireThat(clip.source_path === video, 'Video source does not match fixture');
    near(seconds(clip.timeline_range?.in_point), index * 2, 'Video position');
    near(seconds(clip.timeline_range?.out_point), index * 2 + 2, 'Video end');
    near(seconds(clip.source_range?.in_point), index * 2 + 1, 'Video source in');
    near(seconds(clip.source_range?.out_point), index * 2 + 3, 'Video source out');
  }
  requireThat(audios[0].source_path === audio, 'Audio source does not match fixture');
  near(seconds(audios[0].timeline_range?.in_point), 0, 'Audio position');
  near(seconds(audios[0].timeline_range?.out_point), 4, 'Audio end');
  near(seconds(audios[0].source_range?.in_point), 0, 'Audio source in');
  near(seconds(audios[0].source_range?.out_point), 4, 'Audio source out');
  near(timeline.total_duration_seconds, 4, 'Sequence duration');
}

export function verifyEDLTimeline(timeline, sequence, nativeSequence, edl, assets, fixtureSources) {
  requireThat(timeline.sequence_id === sequence, 'EDL sequence ID differs from timeline readback');
  requireThat(nativeSequence?.id === sequence, 'Native EDL sequence metadata is missing or has a different sequence ID');
  requireThat(Array.isArray(edl?.entries) && edl.entries.length > 0, 'Missing EDL entries');
  requireThat(Number.isFinite(edl.sequence_frame_rate) && edl.sequence_frame_rate > 0, 'Missing EDL frame rate');
  requireThat(Number.isFinite(nativeSequence.frame_rate) && nativeSequence.frame_rate > 0, 'Missing native EDL sequence frame rate');
  near(nativeSequence.frame_rate, edl.sequence_frame_rate, 'EDL sequence frame rate', 0.001);
  const tolerance = 1 / edl.sequence_frame_rate + 0.001;
  const sourcePaths = new Map();
  for (const asset of assets || []) {
    if (!asset.id || !asset.file_path) continue;
    requireThat(!sourcePaths.has(asset.id) || sourcePaths.get(asset.id) === asset.file_path, `Ambiguous scanned asset ID: ${asset.id}`);
    sourcePaths.set(asset.id, asset.file_path);
  }
  const range = (value, label) => {
    const start = seconds(value?.in_point);
    const end = seconds(value?.out_point);
    requireThat(Number.isFinite(start) && start >= 0 && Number.isFinite(end) && end > start, `Invalid ${label}`);
    return [start, end];
  };
  const actual = [];
  for (const [type, tracks] of [[1, timeline.video_tracks], [2, timeline.audio_tracks]]) {
    requireThat(Array.isArray(tracks), 'Missing EDL timeline tracks');
    const indexes = new Set();
    for (const track of tracks) {
      requireThat(Number.isInteger(track.index) && track.index >= 0 && !indexes.has(track.index), 'Invalid or duplicate EDL timeline track index');
      indexes.add(track.index);
      requireThat(track.clips == null || Array.isArray(track.clips), 'Invalid EDL timeline clip list');
      for (const clip of track.clips || []) {
        requireThat(fixtureSources.includes(clip.source_path), 'EDL referenced media outside disposable fixtures');
        actual.push({ type, track: track.index, source: clip.source_path, times: [...range(clip.source_range, 'native source range'), ...range(clip.timeline_range, 'native timeline range')] });
      }
    }
  }
  requireThat(actual.length === edl.entries.length, 'EDL clip count differs from timeline readback');
  const expected = edl.entries.map((entry, index) => {
    const source = fixtureSources.includes(entry.source_asset_id) ? entry.source_asset_id : sourcePaths.get(entry.source_asset_id);
    requireThat(fixtureSources.includes(source), `EDL entry ${index} has no verified fixture source mapping`);
    requireThat([1, 2].includes(entry.track?.type) && Number.isInteger(entry.track?.track_index) && entry.track.track_index >= 0, `EDL entry ${index} has an invalid track target`);
    return { type: entry.track.type, track: entry.track.track_index, source, times: [...range(entry.source_range, 'EDL source range'), ...range(entry.timeline_range, 'EDL timeline range')] };
  });
  const candidates = expected.map(entry => actual.flatMap((clip, index) =>
    entry.type === clip.type && entry.track === clip.track && entry.source === clip.source &&
    entry.times.every((time, i) => Math.abs(time - clip.times[i]) <= tolerance) ? [index] : []));
  // A clip can satisfy only one entry. Reassign earlier matches when frame
  // tolerance creates overlapping candidates, so ordering cannot decide a pass.
  const matched = Array(actual.length).fill(-1);
  function assign(entryIndex, seen) {
    for (const clipIndex of candidates[entryIndex]) {
      if (seen.has(clipIndex)) continue;
      seen.add(clipIndex);
      if (matched[clipIndex] === -1 || assign(matched[clipIndex], seen)) {
        matched[clipIndex] = entryIndex;
        return true;
      }
    }
    return false;
  }
  for (let index = 0; index < expected.length; index++) {
    requireThat(assign(index, new Set()), `EDL entry ${index} has no distinct native clip matching its source, track, source trims, and timeline positions`);
  }
}
export function verifyPNG(base64, width, height) {
  requireThat(typeof base64 === 'string' && /^[A-Za-z0-9+/]+={0,2}$/.test(base64) && base64.length % 4 === 0, 'Invalid PNG base64');
  const bytes = Buffer.from(base64, 'base64');
  requireThat(bytes.length > 45 && bytes.subarray(0, 8).equals(Buffer.from('89504e470d0a1a0a', 'hex')) && bytes.toString('ascii', 12, 16) === 'IHDR' && bytes.toString('ascii', bytes.length - 8, bytes.length - 4) === 'IEND', 'Truncated or invalid PNG');
  requireThat(bytes.readUInt32BE(16) === width && bytes.readUInt32BE(20) === height, 'PNG dimensions do not match sequence');
  return bytes;
}

function reportFor(options) {
  return {
    schema: 1,
    startedAt: new Date().toISOString(),
    mode: options.mutate ? 'mutate' : 'preflight',
    transport: 'mcp-stdio',
    compatibility: {
      hostMetadataSource: 'runner host; remote backend hosts may differ',
      os: os.platform(), osRelease: os.release(), architecture: os.arch(),
      node: process.version,
      locale: options.locale || 'unreported (operator must supply --locale)',
      localeSource: 'operator declaration',
      server: options.server,
      toolProfile: process.env.MCP_TOOL_PROFILE || 'all',
      premiereVersion: 'unverified',
    },
    preflight: { status: 'blocked', reason: 'Not run' },
    checks: CHECKS.map(([id, title]) => ({ id, title, status: 'blocked', reason: 'Native mutation tests not run', evidence: [] })),
    calls: [],
  };
}

function finish(report) {
  report.finishedAt = new Date().toISOString();
  report.summary = { pass: 0, fail: 0, blocked: 0 };
  for (const check of report.checks) report.summary[check.status]++;
  return report;
}
export function exitCode(report) {
  if (report.summary.fail) return 1;
  if (report.preflight.status !== 'pass') return 2;
  return report.mode === 'mutate' && report.summary.blocked ? 2 : 0;
}

export async function runSuite(client, options, fixture) {
  const report = reportFor(options);
  const names = new Set();
  let safetyStopped = false;
  let sequence;
  const allowedSequences = new Set();
  async function call(name, args = {}, timeout = options.timeoutMs) {
    if (!names.has(name)) throw new Blocked(`Tool not exposed: ${name}; use MCP_TOOL_PROFILE=all`);
    const entry = { name, arguments: args, startedAt: new Date().toISOString() };
    report.calls.push(entry);
    try {
      const raw = await client.callTool({ name, arguments: args }, undefined, { timeout });
      const data = decodeResult(raw);
      entry.result = data;
      const images = raw.content?.filter(c => c.type === 'image') || [];
      if (images.length) entry.images = images.map(i => ({ mimeType: i.mimeType, bytes: Buffer.from(i.data, 'base64').length, sha256: sha256(Buffer.from(i.data, 'base64')) }));
      entry.status = 'pass';
      return { data, images };
    } catch (error) {
      entry.status = 'fail'; entry.error = error.message;
      if (options.mutate && /timeout|timed out|connection|closed|transport|EPIPE/i.test(error.message)) safetyStopped = true;
      if (/unsupported|unavailable in this Premiere version/i.test(error.message)) throw new Blocked(error.message);
      throw error;
    } finally { entry.finishedAt = new Date().toISOString(); }
  }
  const data = async (name, args, timeout) => (await call(name, args, timeout)).data;
  try {
    let cursor;
    const cursors = new Set();
    do {
      const page = await client.listTools(cursor ? { cursor } : undefined);
      for (const tool of page.tools) names.add(tool.name);
      cursor = page.nextCursor;
      requireThat(!cursor || !cursors.has(cursor), 'Repeated tool discovery cursor');
      if (cursor) cursors.add(cursor);
    } while (cursor);
    report.compatibility.toolCount = names.size;
    // Standalone ping can invoke AppleScript. Never ping a stopped application.
    const processState = await data('premiere_is_running');
    report.preflight.process = processState;
    if (processState.running !== true) throw new Blocked('Premiere process not detected (or host process detection unsupported); no application ping attempted');
    const ping = await data('premiere_ping');
    report.compatibility.premiereVersion = ping.premiere_version;
    report.compatibility.bridgeMode = ping.bridge_mode;
    if (ping.premiere_running !== true) throw new Blocked('Premiere/CEP is not reachable; native tests cannot run');
    report.preflight = { status: 'pass', process: processState, ping };
    if (ping.project_open) report.preflight.project = await data('premiere_get_project');
    if (!options.mutate) return finish(report);
    if (!ping.project_open) throw new Blocked('Open a saved, empty disposable project before --mutate');
    fixture ??= validateFixture(options.fixtureDir, options.project);
    const project = report.preflight.project;
    requireThat(project.project_path === fixture.project, 'Active project is not the exact disposable project');
    requireThat(Array.isArray(project.sequences) && project.sequences.length === 0, 'Disposable project must start with zero sequences');
    const items = await data('premiere_get_project_items');
    requireThat(Array.isArray(items.items) && items.items.length === 0, 'Disposable project must start with zero project items');
    mkdirSync(fixture.artifacts);
    report.fixture = { project: fixture.project, files: fixture.files, sequence: fixture.sequence };
  } catch (error) {
    report.preflight = { ...report.preflight, status: 'blocked', reason: error.message };
    for (const check of report.checks) check.reason = error.message;
    return finish(report);
  }

  // This is a read-before-write guard, not an atomic Adobe transaction. The
  // operator must leave Premiere untouched for the duration of this run.
  async function guard(active = false) {
    if (safetyStopped) throw new SafetyStop('Safety guard stopped further mutations');
    const project = await data('premiere_get_project');
    if (project.project_path !== fixture.project || !Array.isArray(project.sequences) || project.sequences.some(s => !allowedSequences.has(s.id))) {
      safetyStopped = true;
      throw new SafetyStop('Project or sequence ownership changed; stopped before mutation');
    }
    if (active) {
      const timeline = await data('premiere_get_timeline');
      if (!sequence || timeline.sequence_id !== sequence) {
        safetyStopped = true;
        throw new SafetyStop('Active sequence changed; stopped before mutation');
      }
    }
  }
  async function edit(name, args, active = true, timeout) {
    await guard(active);
    return data(name, args, timeout);
  }
  async function stage(id, prerequisites, body) {
    const check = report.checks.find(c => c.id === id);
    const start = report.calls.length;
    if (safetyStopped || prerequisites.some(p => report.checks.find(c => c.id === p).status !== 'pass')) {
      check.reason = safetyStopped ? 'Safety guard stopped further mutations' : `Prerequisite did not pass: ${prerequisites.join(', ')}`;
      return;
    }
    try { check.evidence = await body(); check.status = 'pass'; delete check.reason; }
    catch (error) { check.status = error instanceof Blocked ? 'blocked' : 'fail'; check.reason = error.message; }
    check.callRange = [start, report.calls.length];
  }
  await stage('import_media', [], async () => {
    await edit('premiere_import_media', { file_path: fixture.video }, false);
    await edit('premiere_import_media', { file_path: fixture.audio }, false);
    const items = await data('premiere_get_project_items');
    for (const source of [fixture.video, fixture.audio]) requireThat(items.items?.some(i => i.media_path === source), `Imported source absent from project readback: ${source}`);
    return [items];
  });
  await stage('sequence_placement', ['import_media'], async () => {
    const created = await edit('premiere_create_sequence', { name: fixture.sequence, width: 320, height: 180, frame_rate: 24, video_tracks: 1, audio_tracks: 1 }, false);
    requireThat(created.sequence_id, 'Sequence creation returned no ID');
    sequence = created.sequence_id;
    allowedSequences.add(sequence);
    const project = await data('premiere_get_project');
    const found = project.sequences?.find(s => s.id === sequence && s.name === fixture.sequence);
    requireThat(found?.resolution?.width === 320 && found?.resolution?.height === 180 && found?.frame_rate === 24, 'Created sequence settings did not read back as 320x180 at 24 fps');
    for (const [source, type, position, start, end] of [[fixture.video, 'video', 0, 1, 3], [fixture.video, 'video', 2, 3, 5], [fixture.audio, 'audio', 0, 0, 4]]) {
      await edit('premiere_place_clip', { source_path: source, track_type: type, track_index: 0, position_seconds: position, in_point_seconds: start, out_point_seconds: end });
    }
    const timeline = await data('premiere_get_timeline');
    verifyTimeline(timeline, sequence, fixture.video, fixture.audio);
    return [found, timeline];
  });
  await stage('transitions_effects', ['sequence_placement'], async () => {
    await edit('premiere_add_video_transition', { track_index: 0, clip_index: 0, transition_name: options.transition, duration: 0.5, apply_to_end: true });
    const transitions = await data('premiere_get_transitions', { track_type: 'video', track_index: 0 });
    const transition = transitions.transitions?.find(t => t.name === options.transition);
    requireThat(transition, 'Transition absent from readback');
    near(transition.duration, 0.5, 'Transition duration');
    await edit('premiere_apply_video_effect', { track_index: 0, clip_index: 0, effect_name: options.effect });
    const effects = await data('premiere_get_clip_effects', { track_type: 'video', track_index: 0, clip_index: 0 });
    requireThat(effects.effects?.some(e => e.displayName === options.effect), 'Effect absent from readback');
    return [transitions, effects];
  });
  await stage('lumetri', ['sequence_placement'], async () => {
    await edit('premiere_lumetri_set_exposure', { track_index: 0, clip_index: 0, value: 0.5 });
    // This getter auto-applies Lumetri on some hosts; guard it as a mutation.
    const readback = await edit('premiere_lumetri_get_all', { track_index: 0, clip_index: 0 });
    near(readback.properties?.Exposure, 0.5, 'Lumetri exposure', 0.001);
    return [readback];
  });
  await stage('audio_levels', ['sequence_placement'], async () => {
    const timeline = await data('premiere_get_timeline');
    const audio = timeline.audio_tracks?.find(t => t.index === 0)?.clips?.[0];
    requireThat(audio?.source_path === fixture.audio && audio.clip_id, 'Fixture audio clip missing');
    await edit('premiere_set_audio_level', { sequence_id: sequence, clip_id: audio.clip_id, level_db: -6 });
    const readback = await data('premiere_get_audio_level', { track_index: 0, clip_index: 0 });
    near(readback.levelDb, -6, 'Audio level dB', 0.1);
    return [readback];
  });
  await stage('h264_export', ['sequence_placement'], async () => {
    if (!options.preset) throw new Blocked('Supply --preset with a real H.264 .epr export preset');
    requireThat(isAbsolute(options.preset) && options.preset.endsWith('.epr'), 'Preset must be an absolute .epr path');
    regular(options.preset);
    command(process.env.FFMPEG || 'ffmpeg', ['-version']);
    command(process.env.FFPROBE || 'ffprobe', ['-version']);
    const output = join(fixture.artifacts, 'export.mp4');
    requireThat(!existsSync(output), 'Export output already exists');
    const result = await edit('premiere_export_direct', { sequence_index: -1, output_path: output, preset_path: options.preset, work_area_type: 0 }, true, Math.max(options.timeoutMs, 180000));
    regular(output);
    const media = probe(output);
    requireThat(media.streams.some(s => s.codec_type === 'video' && s.codec_name === 'h264'), 'Export is not H.264 video');
    requireThat(media.streams.some(s => s.codec_type === 'audio'), 'Export has no audio stream');
    near(Number(media.format.duration), 4, 'Export duration', 0.25);
    decodeMedia(output);
    return [{ result, output, bytes: lstatSync(output).size, sha256: sha256(readFileSync(output)), presetSha256: sha256(readFileSync(options.preset)), probe: media, fullyDecoded: true }];
  });
  await stage('frame_capture', ['sequence_placement'], async () => {
    command(process.env.FFMPEG || 'ffmpeg', ['-version']);
    await edit('premiere_set_playhead_position', { seconds: 1 });
    await guard(true);
    const { data: meta, images } = await call('premiere_capture_frame_base64');
    requireThat(images.length === 1 && images[0].mimeType === 'image/png', 'Expected one PNG MCP image block');
    const bytes = verifyPNG(images[0].data, 320, 180);
    near(meta.timecode, 1, 'Captured frame time');
    const output = join(fixture.artifacts, 'frame.png');
    writeFileSync(output, bytes, { flag: 'wx' });
    decodeMedia(output);
    return [{ output, bytes: bytes.length, sha256: sha256(bytes), metadata: meta, fullyDecoded: true }];
  });
  await stage('script_edl_timeline', ['sequence_placement'], async () => {
    const parsed = await data('premiere_parse_script', { file_path: fixture.script, format: 'youtube' });
    requireThat(parsed.segments?.length > 0, 'Script parser returned no segments');
    // output_name is intentionally omitted: it triggers an unrelated export.
    const result = await edit('premiere_auto_edit', { script_path: fixture.script, assets_directory: join(fixture.root, 'media'), resolution: '1080p', pacing: 'moderate' });
    requireThat(result.edl?.entries?.length > 0, 'Auto-edit returned no EDL entries');
    const execution = result.execution_result;
    requireThat(execution?.sequence_id && execution.sequence_id !== sequence && execution.clips_placed > 0 && !execution.errors?.length, 'EDL execution did not create a verified new sequence');
    requireThat(result.steps?.some(s => s.name === 'execute_edl' && s.status === 'completed') && !result.steps.some(s => s.status === 'failed'), 'Auto-edit pipeline reported incomplete steps');
    const project = await data('premiere_get_project');
    const nativeSequence = project.sequences?.find(s => s.id === execution.sequence_id);
    requireThat(project.project_path === fixture.project && nativeSequence, 'EDL sequence absent from disposable project');
    const timeline = await data('premiere_get_timeline', { sequence_id: execution.sequence_id });
    requireThat(execution.clips_placed === result.edl.entries.length, 'EDL clip count differs from execution result');
    verifyEDLTimeline(timeline, execution.sequence_id, nativeSequence, result.edl, result.scan_result?.assets, [fixture.video, fixture.audio]);
    return [{ parsed, result, nativeSequence, timeline }];
  });
  return finish(report);
}

async function connectClient(options) {
  const require = createRequire(join(resolve(options.sdkRoot), 'package.json'));
  const { Client } = await import(pathToFileURL(require.resolve('@modelcontextprotocol/sdk/client/index.js')));
  const { StdioClientTransport, getDefaultEnvironment } = await import(pathToFileURL(require.resolve('@modelcontextprotocol/sdk/client/stdio.js')));
  const env = getDefaultEnvironment();
  for (const [key, value] of Object.entries(process.env)) if (/^(MCP_|RUST_|PYTHON_|TS_|BRIDGE_|PREMIERE_|INTEL_|MEDIA_)/.test(key) && value !== undefined) env[key] = value;
  env.MCP_TOOL_PROFILE ||= 'all';
  env.MCP_TRANSPORT = 'stdio';
  const client = new Client({ name: 'premiere-live-e2e', version: '1.0.0' }, { capabilities: {} });
  const transport = new StdioClientTransport({ command: resolve(options.server), args: ['--transport=stdio', '--log-level=error'], cwd: ROOT, env, stderr: 'ignore' });
  try { await client.connect(transport, { timeout: options.timeoutMs }); }
  catch (error) { await transport.close(); throw error; }
  return client;
}

async function main() {
  let options;
  let client;
  let report;
  try {
    options = parseArgs(process.argv.slice(2));
    if (options.help) {
      console.log('Read-only: node scripts/e2e-premiere.mjs [--server /path/to/premierpro-mcp] [--sdk-root /path/to/cli] [--report /new/report.json]\nPrepare: node scripts/e2e-premiere.mjs --prepare /new/fixture-directory\nMutate: add --mutate --fixture-dir /fixture-directory --project /fixture-directory/MCP-E2E-disposable.prproj --confirm-disposable MCP-E2E-DISPOSABLE [--preset /path/H264.epr] [--locale en_US]\nSee docs/live-premiere-testing.md for safety requirements and evidence limits.');
      return;
    }
    if (options.prepare) { console.log(JSON.stringify(prepareFixtures(options.prepare), null, 2)); return; }
    // Validate local ownership and missing prerequisites before connecting.
    const fixture = options.mutate ? validateFixture(options.fixtureDir, options.project) : undefined;
    try { client = await connectClient(options); }
    catch (error) { client = { async listTools() { throw error; }, async close() {} }; }
    report = await runSuite(client, options, fixture);
  } catch (error) {
    report = finish(reportFor(options || { mutate: process.argv.includes('--mutate') }));
    report.preflight.reason = error.message;
    for (const check of report.checks) check.reason = error.message;
  } finally {
    if (client) {
      try { await client.close(); }
      catch (error) { if (report) report.transportCloseError = error.message; }
    }
  }
  const output = JSON.stringify(report, null, 2) + '\n';
  if (options?.report) {
    try { writeFileSync(resolve(options.report), output, { flag: 'wx' }); }
    catch (error) { console.error(`Report not saved: ${error.message}`); process.exitCode = 1; }
  }
  process.stdout.write(output);
  process.exitCode ||= exitCode(report);
}
if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) await main();
