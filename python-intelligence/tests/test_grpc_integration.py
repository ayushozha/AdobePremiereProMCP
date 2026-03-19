"""End-to-end gRPC integration tests for the Intelligence service.

Starts the gRPC server in a background thread, creates a real gRPC client,
and exercises the ParseScript, MatchAssets, and GenerateEDL RPCs over the wire.
"""

from __future__ import annotations

import threading
import time

import grpc
import pytest

from premierpro.common.v1 import common_pb2
from premierpro.intelligence.v1 import intelligence_pb2, intelligence_pb2_grpc

from src.config import IntelligenceSettings
from src.grpc_server import create_server

# ── Fixtures ──────────────────────────────────────────────────────────────────

_TEST_PORT = 50099  # Use a non-default port so we never clash with a real server.


@pytest.fixture(scope="module")
def grpc_server():
    """Start the gRPC server in a daemon thread, yield, then shut down."""
    settings = IntelligenceSettings(grpc_port=_TEST_PORT)
    server = create_server(settings)
    server.start()

    yield server

    server.stop(grace=2.0)
    server.wait_for_termination(timeout=5.0)


@pytest.fixture(scope="module")
def stub(grpc_server):
    """Create a gRPC client stub connected to the test server."""
    channel = grpc.insecure_channel(f"localhost:{_TEST_PORT}")
    # Wait until the channel is actually ready.
    grpc.channel_ready_future(channel).result(timeout=5.0)
    yield intelligence_pb2_grpc.IntelligenceServiceStub(channel)
    channel.close()


# ── Sample data ───────────────────────────────────────────────────────────────

_YOUTUBE_SCRIPT = """\
[INTRO]
Hey everyone, welcome back to the channel!

B-ROLL: Aerial drone shot of a city skyline at sunset

ON CAMERA: Today we are diving into the top 5 productivity tips that changed my life

[MAIN]
B-ROLL: Person typing on laptop in a coffee shop

VO: Tip number one -- start your day with the hardest task first

MUSIC: Upbeat lo-fi background music

TEXT ON SCREEN: Tip #1: Eat the Frog

B-ROLL: Close-up of a planner with tasks being crossed off

ON CAMERA: The second tip is all about time blocking

[OUTRO]
ON CAMERA: If you enjoyed this video, smash the like button and subscribe!

LOWER THIRD: @ProductivityGuru

SFX: Subscribe bell sound
"""


def _make_assets() -> list[common_pb2.Asset]:
    """Build a small asset library that can partially match the script."""
    return [
        common_pb2.Asset(
            id="asset-drone-city",
            file_name="drone_city_skyline_sunset.mp4",
            file_path="/media/drone_city_skyline_sunset.mp4",
            asset_type=common_pb2.ASSET_TYPE_VIDEO,
            video=common_pb2.VideoInfo(
                duration_seconds=45.0,
                codec="h264",
                resolution=common_pb2.Resolution(width=3840, height=2160),
                frame_rate=24.0,
            ),
        ),
        common_pb2.Asset(
            id="asset-laptop-typing",
            file_name="person_typing_laptop_coffeeshop.mp4",
            file_path="/media/person_typing_laptop_coffeeshop.mp4",
            asset_type=common_pb2.ASSET_TYPE_VIDEO,
            video=common_pb2.VideoInfo(
                duration_seconds=30.0,
                codec="h264",
                resolution=common_pb2.Resolution(width=1920, height=1080),
                frame_rate=24.0,
            ),
        ),
        common_pb2.Asset(
            id="asset-planner",
            file_name="planner_tasks_crossoff.mp4",
            file_path="/media/planner_tasks_crossoff.mp4",
            asset_type=common_pb2.ASSET_TYPE_VIDEO,
            video=common_pb2.VideoInfo(
                duration_seconds=20.0,
                codec="h264",
                resolution=common_pb2.Resolution(width=1920, height=1080),
                frame_rate=24.0,
            ),
        ),
        common_pb2.Asset(
            id="asset-lofi-music",
            file_name="lofi_background_beats.mp3",
            file_path="/media/lofi_background_beats.mp3",
            asset_type=common_pb2.ASSET_TYPE_AUDIO,
            audio=common_pb2.AudioInfo(
                duration_seconds=180.0,
                codec="mp3",
                sample_rate=44100,
                channels=2,
            ),
        ),
        common_pb2.Asset(
            id="asset-bell-sfx",
            file_name="subscribe_bell_sound.wav",
            file_path="/media/subscribe_bell_sound.wav",
            asset_type=common_pb2.ASSET_TYPE_AUDIO,
            audio=common_pb2.AudioInfo(
                duration_seconds=2.0,
                codec="pcm",
                sample_rate=48000,
                channels=1,
            ),
        ),
        common_pb2.Asset(
            id="asset-talking-head",
            file_name="talking_head_camera.mp4",
            file_path="/media/talking_head_camera.mp4",
            asset_type=common_pb2.ASSET_TYPE_VIDEO,
            video=common_pb2.VideoInfo(
                duration_seconds=600.0,
                codec="h264",
                resolution=common_pb2.Resolution(width=1920, height=1080),
                frame_rate=24.0,
            ),
        ),
    ]


# ── Tests ─────────────────────────────────────────────────────────────────────


class TestParseScript:
    """ParseScript RPC tests."""

    def test_parse_youtube_script(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Send a YouTube script and verify segments are returned."""
        request = intelligence_pb2.ParseScriptRequest(
            text=_YOUTUBE_SCRIPT,
            format_hint="youtube",
        )
        response = stub.ParseScript(request)

        # Should have multiple segments (dialogue, broll, vo, music, etc.)
        assert len(response.segments) > 0, "Expected at least one segment"

        # Verify segment types are populated.
        types_seen = set()
        for seg in response.segments:
            assert seg.content, f"Segment {seg.index} has empty content"
            assert seg.type != intelligence_pb2.SEGMENT_TYPE_UNSPECIFIED, (
                f"Segment {seg.index} has UNSPECIFIED type"
            )
            types_seen.add(seg.type)

        # The YouTube script contains dialogue, broll, VO, music, title, lower_third, sfx.
        assert intelligence_pb2.SEGMENT_TYPE_DIALOGUE in types_seen, "Expected DIALOGUE segments"
        assert intelligence_pb2.SEGMENT_TYPE_BROLL in types_seen, "Expected BROLL segments"

        # Metadata should be populated.
        assert response.metadata.format == "youtube"
        assert response.metadata.segment_count == len(response.segments)
        assert response.metadata.estimated_total_duration_seconds > 0

    def test_parse_script_autodetect(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """ParseScript with format_hint='auto' should auto-detect the format."""
        request = intelligence_pb2.ParseScriptRequest(
            text=_YOUTUBE_SCRIPT,
            format_hint="auto",
        )
        response = stub.ParseScript(request)
        assert len(response.segments) > 0
        # Auto-detection should detect this as youtube format.
        assert response.metadata.format == "youtube"

    def test_parse_empty_text_returns_error(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Sending empty text should return an INVALID_ARGUMENT error."""
        request = intelligence_pb2.ParseScriptRequest(text="", format_hint="auto")
        with pytest.raises(grpc.RpcError) as exc_info:
            stub.ParseScript(request)
        assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT


class TestMatchAssets:
    """MatchAssets RPC tests."""

    def test_keyword_matching(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Match segments to assets using the KEYWORD strategy."""
        # First parse the script to get real segments.
        parse_resp = stub.ParseScript(
            intelligence_pb2.ParseScriptRequest(
                text=_YOUTUBE_SCRIPT,
                format_hint="youtube",
            )
        )
        assert len(parse_resp.segments) > 0

        # Now match those segments against our asset library.
        match_req = intelligence_pb2.MatchAssetsRequest(
            segments=parse_resp.segments,
            available_assets=_make_assets(),
            strategy=intelligence_pb2.MATCH_STRATEGY_KEYWORD,
        )
        match_resp = stub.MatchAssets(match_req)

        # Should have at least one match (drone/city skyline segment should match
        # the drone_city_skyline_sunset asset).
        assert len(match_resp.matches) > 0, (
            f"Expected matches but got 0. Unmatched: {len(match_resp.unmatched)}"
        )

        # Every match should have valid fields.
        for m in match_resp.matches:
            assert m.asset_id, f"Match for segment {m.segment_index} missing asset_id"
            assert 0.0 <= m.confidence <= 1.0, (
                f"Confidence {m.confidence} out of range for segment {m.segment_index}"
            )
            assert m.reasoning, f"Match for segment {m.segment_index} missing reasoning"

        # Unmatched segments should have suggestions.
        for u in match_resp.unmatched:
            assert u.reason, f"Unmatched segment {u.segment_index} missing reason"

    def test_matching_with_no_assets(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Matching with an empty asset library produces all unmatched."""
        parse_resp = stub.ParseScript(
            intelligence_pb2.ParseScriptRequest(
                text=_YOUTUBE_SCRIPT,
                format_hint="youtube",
            )
        )

        match_req = intelligence_pb2.MatchAssetsRequest(
            segments=parse_resp.segments,
            available_assets=[],
            strategy=intelligence_pb2.MATCH_STRATEGY_KEYWORD,
        )
        match_resp = stub.MatchAssets(match_req)

        assert len(match_resp.matches) == 0
        assert len(match_resp.unmatched) == len(parse_resp.segments)


class TestGenerateEDL:
    """GenerateEDL RPC tests."""

    def test_full_pipeline(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Run the full pipeline: parse -> match -> generate EDL."""
        # Step 1: Parse.
        parse_resp = stub.ParseScript(
            intelligence_pb2.ParseScriptRequest(
                text=_YOUTUBE_SCRIPT,
                format_hint="youtube",
            )
        )
        assert len(parse_resp.segments) > 0

        # Step 2: Match.
        assets = _make_assets()
        match_resp = stub.MatchAssets(
            intelligence_pb2.MatchAssetsRequest(
                segments=parse_resp.segments,
                available_assets=assets,
                strategy=intelligence_pb2.MATCH_STRATEGY_KEYWORD,
            )
        )

        # Step 3: Generate EDL.
        edl_req = intelligence_pb2.GenerateEDLRequest(
            segments=parse_resp.segments,
            available_assets=assets,
            matches=match_resp.matches,
            settings=intelligence_pb2.EDLSettings(
                resolution=common_pb2.Resolution(width=1920, height=1080),
                frame_rate=24.0,
                default_transition="cut",
                default_transition_duration=0.0,
                pacing=intelligence_pb2.PACING_PRESET_MODERATE,
            ),
        )
        edl_resp = stub.GenerateEDL(edl_req)

        # The EDL should have at least as many entries as we have matches.
        # (Some segments may produce both a video entry and a companion audio entry.)
        assert edl_resp.edl is not None
        assert edl_resp.edl.id, "EDL should have a non-empty id"
        assert edl_resp.edl.name, "EDL should have a non-empty name"
        assert edl_resp.edl.sequence_resolution.width == 1920
        assert edl_resp.edl.sequence_resolution.height == 1080
        assert edl_resp.edl.sequence_frame_rate == 24.0

        if match_resp.matches:
            assert len(edl_resp.edl.entries) > 0, "Expected EDL entries for matched segments"

            for entry in edl_resp.edl.entries:
                assert entry.source_asset_id, f"Entry {entry.index} missing source_asset_id"
                # Track should be valid.
                assert entry.track.type in (
                    common_pb2.TRACK_TYPE_VIDEO,
                    common_pb2.TRACK_TYPE_AUDIO,
                ), f"Entry {entry.index} has unexpected track type {entry.track.type}"

    def test_edl_with_explicit_matches(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Generate an EDL with hand-crafted segments and matches."""
        segments = [
            intelligence_pb2.ScriptSegment(
                index=0,
                type=intelligence_pb2.SEGMENT_TYPE_DIALOGUE,
                content="Welcome to the show",
                estimated_duration_seconds=3.0,
            ),
            intelligence_pb2.ScriptSegment(
                index=1,
                type=intelligence_pb2.SEGMENT_TYPE_BROLL,
                content="City skyline at sunset",
                visual_direction="City skyline at sunset",
                estimated_duration_seconds=5.0,
            ),
        ]

        assets = [
            common_pb2.Asset(
                id="a1",
                file_name="host.mp4",
                asset_type=common_pb2.ASSET_TYPE_VIDEO,
                video=common_pb2.VideoInfo(duration_seconds=60.0),
            ),
            common_pb2.Asset(
                id="a2",
                file_name="skyline.mp4",
                asset_type=common_pb2.ASSET_TYPE_VIDEO,
                video=common_pb2.VideoInfo(duration_seconds=30.0),
            ),
        ]

        matches = [
            intelligence_pb2.AssetMatch(
                segment_index=0,
                asset_id="a1",
                confidence=0.95,
                reasoning="Direct match for dialogue",
            ),
            intelligence_pb2.AssetMatch(
                segment_index=1,
                asset_id="a2",
                confidence=0.90,
                reasoning="Skyline B-roll match",
            ),
        ]

        req = intelligence_pb2.GenerateEDLRequest(
            segments=segments,
            available_assets=assets,
            matches=matches,
            settings=intelligence_pb2.EDLSettings(
                resolution=common_pb2.Resolution(width=1920, height=1080),
                frame_rate=24.0,
                default_transition="cut",
            ),
        )
        resp = stub.GenerateEDL(req)

        assert resp.edl is not None
        # Dialogue on V1 + companion audio on A1 + B-roll on V2 = 3 entries.
        assert len(resp.edl.entries) == 3, (
            f"Expected 3 entries (video+audio for dialogue, video for broll), "
            f"got {len(resp.edl.entries)}"
        )

        # Check the first entry is the dialogue video track.
        e0 = resp.edl.entries[0]
        assert e0.source_asset_id == "a1"
        assert e0.track.type == common_pb2.TRACK_TYPE_VIDEO
        assert e0.track.track_index == 0  # V1

        # Second entry should be the companion audio.
        e1 = resp.edl.entries[1]
        assert e1.source_asset_id == "a1"
        assert e1.track.type == common_pb2.TRACK_TYPE_AUDIO
        assert e1.track.track_index == 0  # A1

        # Third entry is the B-roll.
        e2 = resp.edl.entries[2]
        assert e2.source_asset_id == "a2"
        assert e2.track.type == common_pb2.TRACK_TYPE_VIDEO
        assert e2.track.track_index == 1  # V2


class TestAnalyzePacing:
    """AnalyzePacing RPC tests."""

    def test_pacing_analysis(self, stub: intelligence_pb2_grpc.IntelligenceServiceStub):
        """Analyze pacing of a hand-built EDL."""
        edl = common_pb2.EditDecisionList(
            id="test-edl-1",
            name="Test Sequence",
            sequence_resolution=common_pb2.Resolution(width=1920, height=1080),
            sequence_frame_rate=24.0,
            entries=[
                common_pb2.EDLEntry(
                    index=0,
                    source_asset_id="a1",
                    source_range=common_pb2.TimeRange(
                        in_point=common_pb2.Timecode(hours=0, minutes=0, seconds=0, frames=0, frame_rate=24.0),
                        out_point=common_pb2.Timecode(hours=0, minutes=0, seconds=3, frames=0, frame_rate=24.0),
                    ),
                    timeline_range=common_pb2.TimeRange(
                        in_point=common_pb2.Timecode(hours=0, minutes=0, seconds=0, frames=0, frame_rate=24.0),
                        out_point=common_pb2.Timecode(hours=0, minutes=0, seconds=3, frames=0, frame_rate=24.0),
                    ),
                    track=common_pb2.TrackTarget(type=common_pb2.TRACK_TYPE_VIDEO, track_index=0),
                    notes="Segment 0: dialogue",
                ),
                common_pb2.EDLEntry(
                    index=1,
                    source_asset_id="a2",
                    source_range=common_pb2.TimeRange(
                        in_point=common_pb2.Timecode(hours=0, minutes=0, seconds=0, frames=0, frame_rate=24.0),
                        out_point=common_pb2.Timecode(hours=0, minutes=0, seconds=5, frames=0, frame_rate=24.0),
                    ),
                    timeline_range=common_pb2.TimeRange(
                        in_point=common_pb2.Timecode(hours=0, minutes=0, seconds=3, frames=0, frame_rate=24.0),
                        out_point=common_pb2.Timecode(hours=0, minutes=0, seconds=8, frames=0, frame_rate=24.0),
                    ),
                    track=common_pb2.TrackTarget(type=common_pb2.TRACK_TYPE_VIDEO, track_index=1),
                    notes="Segment 1: broll",
                ),
            ],
        )

        req = intelligence_pb2.AnalyzePacingRequest(
            edl=edl,
            target_mood="energetic",
        )
        resp = stub.AnalyzePacing(req)

        assert len(resp.adjustments) == 2, f"Expected 2 adjustments, got {len(resp.adjustments)}"
        assert resp.current_avg_clip_duration > 0
        assert resp.suggested_avg_clip_duration > 0

        for adj in resp.adjustments:
            assert adj.current_duration > 0
            assert adj.suggested_duration > 0
            assert adj.reason, f"Adjustment for entry {adj.edl_entry_index} missing reason"
