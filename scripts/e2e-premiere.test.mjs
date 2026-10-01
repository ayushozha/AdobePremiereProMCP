import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync, readFileSync, mkdirSync, symlinkSync, rmSync, realpathSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseArgs, runSuite, exitCode, decodeResult, verifyTimeline, verifyEDLTimeline, verifyPNG, validateFixture, prepareFixtures } from './e2e-premiere.mjs';

const textResult = value => ({ content: [{ type: 'text', text: JSON.stringify(value) }] });
const tc = seconds => ({ hours: 0, minutes: 0, seconds, frames: 0, frame_rate: 24 });
const clip = (source, start, end) => ({ clip_id: source, source_path: source, timeline_range: { in_point: tc(start), out_point: tc(end) }, source_range: source.endsWith('.wav') ? { in_point: tc(0), out_point: tc(4) } : { in_point: tc(start + 1), out_point: tc(end + 1) } });
const fakeClient = (ping, extra = {}) => ({
  async listTools() { return { tools: [{ name: 'premiere_is_running' }, { name: 'premiere_ping' }, { name: 'premiere_get_project' }] }; },
  async callTool({ name }) {
    if (name === 'premiere_is_running') return textResult({ running: true });
    if (name === 'premiere_ping') return textResult(ping);
    if (name === 'premiere_get_project') return textResult({ project_path: '/unrelated.prproj', sequences: [] });
    throw new Error(`Unexpected mutation: ${name}`);
  }, ...extra,
});

test('default preflight never edits or claims the eight native checks passed', async () => {
  const report = await runSuite(fakeClient({ premiere_running: true, project_open: true, premiere_version: '26.0' }), parseArgs([]));
  assert.equal(report.mode, 'preflight');
  assert.equal(report.preflight.status, 'pass');
  assert.equal(report.checks.length, 8);
  assert.deepEqual(report.summary, { pass: 0, fail: 0, blocked: 8 });
  assert.equal(exitCode(report), 0);
});

test('unreachable Premiere blocks every native check and exits 2', async () => {
  const report = await runSuite(fakeClient({ premiere_running: false, premiere_version: 'unknown', project_open: false }), parseArgs([]));
  assert.equal(report.preflight.status, 'blocked');
  assert.equal(report.checks.filter(c => c.status === 'blocked').length, 8);
  assert.equal(exitCode(report), 2);
});

test('transport loss produces complete failure accounting', async () => {
  const report = await runSuite(fakeClient({}, { async listTools() { throw new Error('connection lost'); } }), parseArgs([]));
  assert.equal(report.checks.length, 8);
  assert.equal(report.preflight.status, 'blocked');
  assert.match(report.preflight.reason, /connection lost/);
});

test('mutation requires all disposable acknowledgements before connecting', () => {
  assert.throws(() => parseArgs(['--mutate']), /fixture-dir/);
  assert.throws(() => parseArgs(['--mutate', '--fixture-dir', '/tmp/a', '--project', '/tmp/a/x.prproj']), /confirm-disposable/);
  assert.throws(() => parseArgs(['--prepare', '/tmp/a', '--mutate']), /combined/);
  assert.throws(() => parseArgs(['--unknown']), /Unknown/);
});

test('nested native errors cannot be hidden by a success wrapper', () => {
  assert.throws(() => decodeResult(textResult({ status: 'ok', message: JSON.stringify({ success: false, error: 'readback mismatch' }) })), /readback mismatch/);
  assert.throws(() => decodeResult({ isError: true, content: [{ type: 'text', text: 'bridge rejected call' }] }), /bridge rejected/);
  assert.deepEqual(decodeResult(textResult({ status: 'success', message: '{"properties":{"Exposure":0.5}}' })), { properties: { Exposure: 0.5 } });
});

test('timeline evidence rejects missing clips, wrong source and wrong time', () => {
  const timeline = { sequence_id: 'seq', total_duration_seconds: 4, video_tracks: [{ index: 0, clips: [clip('/video.mp4', 0, 2), clip('/video.mp4', 2, 4)] }], audio_tracks: [{ index: 0, clips: [clip('/audio.wav', 0, 4)] }] };
  verifyTimeline(timeline, 'seq', '/video.mp4', '/audio.wav');
  assert.throws(() => verifyTimeline({ ...timeline, sequence_id: 'other' }, 'seq', '/video.mp4', '/audio.wav'), /sequence/);
  assert.throws(() => verifyTimeline(timeline, 'seq', '/wrong.mp4', '/audio.wav'), /source/);
  timeline.video_tracks[0].clips[1].timeline_range.in_point = tc(3);
  assert.throws(() => verifyTimeline(timeline, 'seq', '/video.mp4', '/audio.wav'), /position/);
});

test('base64 image evidence rejects empty, truncated or wrong-size PNGs', () => {
  assert.throws(() => verifyPNG('not-an-image', 320, 180), /PNG/);
  const header = Buffer.alloc(33);
  Buffer.from('89504e470d0a1a0a', 'hex').copy(header);
  header.write('IHDR', 12); header.writeUInt32BE(320, 16); header.writeUInt32BE(180, 20);
  assert.throws(() => verifyPNG(header.toString('base64'), 320, 180), /PNG/);
});

test('failed native checks win over blocked checks in the process exit code', () => {
  assert.equal(exitCode({ mode: 'mutate', preflight: { status: 'pass' }, summary: { pass: 5, fail: 1, blocked: 1 } }), 1);
  assert.equal(exitCode({ mode: 'mutate', preflight: { status: 'pass' }, summary: { pass: 6, fail: 0, blocked: 1 } }), 2);
});

test('fixture guard rejects unrelated project and symlinked fixture assets', () => {
  const directory = realpathSync(mkdtempSync(join(tmpdir(), 'premiere-e2e-test-')));
  try {
    mkdirSync(join(directory, 'media'));
    writeFileSync(join(directory, 'manifest.json'), JSON.stringify({ schema: 1, purpose: 'premiere-mcp-disposable-e2e', project: join(directory, 'MCP-E2E-disposable.prproj'), sequence: 'MCP-E2E-fixture', files: {} }));
    assert.throws(() => validateFixture(directory, '/unrelated.prproj'), /project/);
    symlinkSync('/tmp', join(directory, 'MCP-E2E-disposable.prproj'));
    assert.throws(() => validateFixture(directory, join(directory, 'MCP-E2E-disposable.prproj')), /symlink|regular file/);
  } finally { rmSync(directory, { recursive: true, force: true }); }
});

test('source trims must match the requested source range, not only timeline timing', () => {
  const first = clip('/video.mp4', 0, 2);
  const second = clip('/video.mp4', 2, 4);
  second.source_range = { in_point: tc(3), out_point: tc(5) };
  const audio = clip('/audio.wav', 0, 4);
  audio.source_range = { in_point: tc(0), out_point: tc(4) };
  const timeline = { sequence_id: 'seq', total_duration_seconds: 4, video_tracks: [{ index: 0, clips: [first, second] }], audio_tracks: [{ index: 0, clips: [audio] }] };
  verifyTimeline(timeline, 'seq', '/video.mp4', '/audio.wav');
  first.source_range.in_point = tc(0);
  assert.throws(() => verifyTimeline(timeline, 'seq', '/video.mp4', '/audio.wav'), /source in/);
});

const edlFixture = () => {
  const first = clip('/video.mp4', 0, 2);
  const second = clip('/video.mp4', 2, 4);
  const audio = clip('/audio.wav', 0, 4);
  const entry = (native, type, index, asset) => ({ source_asset_id: asset, track: { type, track_index: index }, source_range: structuredClone(native.source_range), timeline_range: structuredClone(native.timeline_range) });
  return {
    edl: { sequence_frame_rate: 24, entries: [entry(second, 1, 0, 'video-id'), entry(audio, 2, 1, 'audio-id'), entry(first, 1, 0, 'video-id')] },
    assets: [{ id: 'video-id', file_path: '/video.mp4' }, { id: 'audio-id', file_path: '/audio.wav' }],
    sources: ['/video.mp4', '/audio.wav'],
    nativeSequence: { id: 'auto', name: 'Auto edit', resolution: { width: 320, height: 180 }, frame_rate: 24 },
    timeline: { sequence_id: 'auto', total_duration_seconds: 4, video_tracks: [{ index: 2, clips: [] }, { index: 0, clips: [first, second] }], audio_tracks: [{ index: 1, clips: [audio] }, { index: 0, clips: [] }] },
  };
};

test('EDL readback matches scanned asset IDs, repeated sources, and unordered tracks and entries', () => {
  const { timeline, nativeSequence, edl, assets, sources } = edlFixture();
  assert.equal(Object.hasOwn(timeline, 'frame_rate'), false, 'Go TimelineState has no frame_rate field');
  verifyEDLTimeline(timeline, 'auto', nativeSequence, edl, assets, sources);
  // Within-one-frame candidates may overlap. A complete distinct assignment
  // exists here, but a greedy match of the first entry to the first clip fails.
  const ranged = frame => ({ in_point: { ...tc(0), frames: frame }, out_point: { ...tc(2), frames: frame } });
  edl.entries = [1, 0].map(frame => ({ source_asset_id: 'video-id', track: { type: 1, track_index: 0 }, source_range: ranged(frame), timeline_range: ranged(frame) }));
  timeline.video_tracks = [{ index: 0, clips: [0, 2].map(frame => ({ source_path: '/video.mp4', source_range: ranged(frame), timeline_range: ranged(frame) })) }];
  timeline.audio_tracks = [];
  verifyEDLTimeline(timeline, 'auto', nativeSequence, edl, assets, sources);
});

test('EDL readback rejects equal-count edits with wrong positions, source trims, targets, or duplicate clips', () => {
  for (const alter of [
    f => { f.timeline.video_tracks[1].clips[1].timeline_range = { in_point: tc(99), out_point: tc(120) }; },
    f => { f.timeline.video_tracks[1].clips[1].source_range = { in_point: tc(6), out_point: tc(7) }; },
    f => { f.timeline.video_tracks[1].index = 99; },
    f => { f.timeline.video_tracks[1].clips[1].source_path = '/audio.wav'; },
    f => { f.timeline.video_tracks[1].clips[1] = structuredClone(f.timeline.video_tracks[1].clips[0]); },
    f => { f.edl.entries[0].source_asset_id = 'unmapped-id'; },
    f => { f.nativeSequence.frame_rate = 30; },
  ]) {
    const f = edlFixture();
    alter(f);
    assert.throws(() => verifyEDLTimeline(f.timeline, 'auto', f.nativeSequence, f.edl, f.assets, f.sources), /matching|mapping|frame rate/);
  }
});

test('EDL readback requires matching native sequence identity and finite frame rate metadata', () => {
  const f = edlFixture();
  for (const nativeSequence of [undefined, null, { id: 'another-sequence', frame_rate: 24 },
    ...[undefined, null, 0, -1, NaN, Infinity, '24'].map(frame_rate => ({ id: 'auto', frame_rate }))]) {
    assert.throws(() => verifyEDLTimeline(f.timeline, 'auto', nativeSequence, f.edl, f.assets, f.sources), /sequence metadata|frame rate/);
  }
  assert.throws(() => verifyEDLTimeline({ ...f.timeline, sequence_id: 'another-sequence' }, 'auto', f.nativeSequence, f.edl, f.assets, f.sources), /sequence ID/);
});

for (const outcome of ['correct Go readback', 'wrong timing and track', 'missing frame rate', 'wrong frame rate']) {
test('script workflow validates native sequence metadata: ' + outcome, async () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'premiere-e2e-edl-')));
  try {
    const fixture = { root, project: join(root, 'MCP-E2E-disposable.prproj'), artifacts: join(root, 'artifacts'), sequence: 'MCP-E2E-fixture', video: '/video.mp4', audio: '/audio.wav', script: '/script.txt' };
    const tools = ['premiere_is_running', 'premiere_ping', 'premiere_get_project', 'premiere_get_project_items', 'premiere_import_media', 'premiere_create_sequence', 'premiere_place_clip', 'premiere_get_timeline', 'premiere_parse_script', 'premiere_auto_edit'];
    let imported = 0;
    let created = false;
    let auto = false;
    const client = {
      async listTools() { return { tools: tools.map(name => ({ name })) }; },
      async callTool({ name, arguments: args }) {
        if (name === 'premiere_is_running') return textResult({ running: true });
        if (name === 'premiere_ping') return textResult({ premiere_running: true, project_open: true });
        if (name === 'premiere_get_project') return textResult({ project_path: fixture.project, sequences: created ? [{ id: 'seq', name: fixture.sequence, resolution: { width: 320, height: 180 }, frame_rate: 24 }, ...(auto ? [{ id: 'auto', name: 'Auto edit', resolution: { width: 320, height: 180 }, frame_rate: outcome === 'missing frame rate' ? undefined : outcome === 'wrong frame rate' ? 30 : 24 }] : [])] : [] });
        if (name === 'premiere_get_project_items') return textResult({ items: imported === 2 ? [{ media_path: fixture.video }, { media_path: fixture.audio }] : [] });
        if (name === 'premiere_import_media') { imported++; return textResult({ imported: true }); }
        if (name === 'premiere_create_sequence') { created = true; return textResult({ sequence_id: 'seq' }); }
        if (name === 'premiere_place_clip') return textResult({ placed: true });
        if (name === 'premiere_get_timeline') {
          const autoClip = clip(fixture.video, 0, 4);
          autoClip.source_range = { in_point: tc(0), out_point: tc(4) };
          return textResult(args.sequence_id === 'auto'
            ? { sequence_id: 'auto', total_duration_seconds: 4, video_tracks: [{ index: outcome === 'wrong timing and track' ? 99 : 0, clips: [outcome === 'wrong timing and track' ? clip(fixture.video, 99, 120) : autoClip] }], audio_tracks: [] }
            : { sequence_id: 'seq', total_duration_seconds: 4, video_tracks: [{ index: 0, clips: [clip(fixture.video, 0, 2), clip(fixture.video, 2, 4)] }], audio_tracks: [{ index: 0, clips: [clip(fixture.audio, 0, 4)] }] });
        }
        if (name === 'premiere_parse_script') return textResult({ segments: [{ text: 'show clip' }] });
        if (name === 'premiere_auto_edit') {
          auto = true;
          return textResult({ edl: { sequence_frame_rate: 24, entries: [{ source_asset_id: fixture.video, track: { type: 1, track_index: 0 }, source_range: { in_point: tc(0), out_point: tc(4) }, timeline_range: { in_point: tc(0), out_point: tc(4) } }] }, execution_result: { sequence_id: 'auto', clips_placed: 1, errors: [] }, steps: [{ name: 'execute_edl', status: 'completed' }] });
        }
        throw new Error(`Unexpected tool: ${name}`);
      },
    };
    const report = await runSuite(client, { ...parseArgs([]), mutate: true }, fixture);
    const check = report.checks.find(c => c.id === 'script_edl_timeline');
    if (outcome === 'correct Go readback') {
      assert.equal(check.status, 'pass');
      assert.deepEqual(report.summary, { pass: 3, fail: 0, blocked: 5 });
    } else {
      assert.equal(check.status, 'fail');
      assert.match(check.reason, outcome === 'wrong timing and track' ? /no distinct native clip/ : /frame rate/);
      assert.deepEqual(report.summary, { pass: 2, fail: 1, blocked: 5 });
      assert.equal(exitCode(report), 1);
    }
  } finally { rmSync(root, { recursive: true, force: true }); }
});
}

test('project switch between imports blocks the next write and every dependent test', async () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'premiere-e2e-guard-')));
  try {
    const fixture = { project: join(root, 'MCP-E2E-disposable.prproj'), artifacts: join(root, 'artifacts'), sequence: 'MCP-E2E-fixture', video: '/fixture/video.mp4', audio: '/fixture/audio.wav' };
    const options = { ...parseArgs([]), mutate: true };
    let reads = 0;
    const writes = [];
    const client = {
      async listTools() { return { tools: ['premiere_is_running', 'premiere_ping', 'premiere_get_project', 'premiere_get_project_items', 'premiere_import_media'].map(name => ({ name })) }; },
      async callTool({ name, arguments: args }) {
        if (name === 'premiere_is_running') return textResult({ running: true });
    if (name === 'premiere_ping') return textResult({ premiere_running: true, project_open: true });
        if (name === 'premiere_get_project') return textResult({ project_path: ++reads <= 2 ? fixture.project : '/another-project.prproj', sequences: [] });
        if (name === 'premiere_get_project_items') return textResult({ items: [] });
        writes.push(args.file_path);
        return textResult({ project_item_id: 'fixture-video' });
      },
    };
    const report = await runSuite(client, options, fixture);
    assert.deepEqual(writes, ['/fixture/video.mp4']);
    assert.equal(report.checks[0].status, 'blocked');
    assert.match(report.checks[0].reason, /ownership changed/);
    assert.deepEqual(report.summary, { pass: 0, fail: 0, blocked: 8 });
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test('fake successful import without native readback fails import and blocks dependent checks', async () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'premiere-e2e-readback-')));
  try {
    const fixture = { project: join(root, 'MCP-E2E-disposable.prproj'), artifacts: join(root, 'artifacts'), sequence: 'MCP-E2E-fixture', video: '/fixture/video.mp4', audio: '/fixture/audio.wav' };
    const client = {
      async listTools() { return { tools: ['premiere_is_running', 'premiere_ping', 'premiere_get_project', 'premiere_get_project_items', 'premiere_import_media'].map(name => ({ name })) }; },
      async callTool({ name }) {
        if (name === 'premiere_is_running') return textResult({ running: true });
    if (name === 'premiere_ping') return textResult({ premiere_running: true, project_open: true });
        if (name === 'premiere_get_project') return textResult({ project_path: fixture.project, sequences: [] });
        if (name === 'premiere_get_project_items') return textResult({ items: [] });
        return textResult({ project_item_id: 'alleged-success' });
      },
    };
    const report = await runSuite(client, { ...parseArgs([]), mutate: true }, fixture);
    assert.deepEqual(report.summary, { pass: 0, fail: 1, blocked: 7 });
    assert.equal(exitCode(report), 1);
  } finally { rmSync(root, { recursive: true, force: true }); }
});


test('real generated fixtures decode, repeat exactly, resist overwrite and detect tampering', { skip: spawnSync(process.env.FFMPEG || 'ffmpeg', ['-version']).status !== 0 || spawnSync(process.env.FFPROBE || 'ffprobe', ['-version']).status !== 0 }, () => {
  const base = realpathSync(mkdtempSync(join(tmpdir(), 'premiere-e2e-media-')));
  try {
    const first = prepareFixtures(join(base, 'first'));
    const second = prepareFixtures(join(base, 'second'));
    assert.equal(first.files['media/e2e_test_pattern.mp4'].probe.streams[0].codec_name, 'h264');
    assert.equal(first.files['media/e2e_tone.wav'].probe.streams[0].codec_name, 'pcm_s16le');
    assert.equal(Number(first.files['media/e2e_test_pattern.mp4'].probe.format.duration), 8);
    assert.equal(readFileSync(join(base, 'first/script.txt'), 'utf8'), 'B-ROLL: "e2e_test_pattern.mp4"\n');
    for (const name of ['media/e2e_test_pattern.mp4', 'media/e2e_tone.wav', 'script.txt']) assert.equal(first.files[name].sha256, second.files[name].sha256);
    assert.throws(() => prepareFixtures(join(base, 'first')), /EEXIST/);
    writeFileSync(first.project, 'test-only placeholder; never opened in Premiere');
    assert.equal(validateFixture(join(base, 'first'), first.project).sequence, 'MCP-E2E-fixture');
    writeFileSync(join(base, 'first', 'media/e2e_tone.wav'), 'tampered');
    assert.throws(() => validateFixture(join(base, 'first'), first.project), /hash mismatch/);
  } finally { rmSync(base, { recursive: true, force: true }); }
});

const intelligenceRoot = fileURLToPath(new URL('../python-intelligence/', import.meta.url));
const fixturePython = process.env.E2E_PYTHON || 'python3';
const fixturePythonReady = spawnSync(fixturePython, ['-c', 'import sys; assert sys.version_info >= (3, 12); import pydantic'], { cwd: intelligenceRoot }).status === 0;
const fixtureMediaReady = spawnSync(process.env.FFMPEG || 'ffmpeg', ['-version']).status === 0 && spawnSync(process.env.FFPROBE || 'ffprobe', ['-version']).status === 0;
const requireParserTest = process.env.E2E_REQUIRE_PARSER_TEST === '1';

test('generated script uses the real parser and matches only its exact fixture video', {
  skip: !requireParserTest && (!fixturePythonReady || !fixtureMediaReady),
}, () => {
  assert.ok(fixturePythonReady, 'Requires intelligence Python3.12+pydantic (set E2E_PYTHON)');
  assert.ok(fixtureMediaReady, 'Requires ffmpeg and ffprobe for the real parser fixture regression');
  const base = realpathSync(mkdtempSync(join(tmpdir(), 'premiere-e2e-script-')));
  try {
    prepareFixtures(join(base, 'fixture'));
    const script = readFileSync(join(base, 'fixture/script.txt'), 'utf8');
    const pythonCheck = String.raw`
import json, sys
from pathlib import Path
from src.parser.script_parser import ScriptParser
from src.matching.matcher import AssetMatcher
from src.models import AssetInfo, AssetType, SegmentType
request = json.load(sys.stdin)
parser = ScriptParser()
matcher = AssetMatcher()
matcher.embedding_matcher._available = False  # Never call a paid embedding API.
video = AssetInfo(id='fixture-video', file_name=Path(request['video']).name, file_path=request['video'], asset_type=AssetType.VIDEO)
audio = AssetInfo(id='fixture-audio', file_name=Path(request['audio']).name, file_path=request['audio'], asset_type=AssetType.AUDIO)
similar = AssetInfo(id='similar-video', file_name='e2e_test_pattern_backup.mp4', file_path=str(Path(request['video']).with_name('e2e_test_pattern_backup.mp4')), asset_type=AssetType.VIDEO)
for format_hint in ['youtube', 'auto']:
    parsed = parser.parse(request['script'], format_hint=format_hint)
    assert [s.type for s in parsed.segments] == [SegmentType.BROLL], parsed
    result = matcher.match(parsed.segments, [video, audio, similar])
    assert [(m.asset_id, m.confidence) for m in result.matches] == [('fixture-video', 1.0)], result
    assert not result.unmatched, result
    missing = matcher.match(parsed.segments, [audio, similar])
    assert not missing.matches and len(missing.unmatched) == 1, missing
print(json.dumps({'segmentType':'broll', 'matchedAsset':'fixture-video', 'confidence':1.0, 'unrelatedMatches':0}))
`;
    const checked = spawnSync(fixturePython, ['-c', pythonCheck], {
      cwd: intelligenceRoot, encoding: 'utf8', timeout: 10000,
      input: JSON.stringify({ script, video: join(base, 'fixture/media/e2e_test_pattern.mp4'), audio: join(base, 'fixture/media/e2e_tone.wav') }),
    });
    assert.equal(checked.status, 0, checked.stderr || checked.error?.message);
    assert.deepEqual(JSON.parse(checked.stdout), { segmentType: 'broll', matchedAsset: 'fixture-video', confidence: 1, unrelatedMatches: 0 });
  } finally { rmSync(base, { recursive: true, force: true }); }
});

test('preflight checks the OS process before ping so a stopped standalone host cannot be launched', async () => {
  const client = {
    async listTools() { return { tools: [{ name: 'premiere_is_running' }, { name: 'premiere_ping' }] }; },
    async callTool({ name }) {
      assert.equal(name, 'premiere_is_running', 'must not ping a stopped application');
      return textResult({ running: false });
    },
  };
  const report = await runSuite(client, parseArgs([]));
  assert.equal(report.preflight.status, 'blocked');
  assert.match(report.preflight.reason, /process/);
  assert.deepEqual(report.summary, { pass: 0, fail: 0, blocked: 8 });
});
