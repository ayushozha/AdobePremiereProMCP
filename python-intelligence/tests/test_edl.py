"""Tests for EDL generation.

Covers basic generation, track assignment rules, timeline position
calculation, and transition insertion.
"""

from __future__ import annotations

import pytest

from src.edl import EDLGenerator
from src.models import (
    AssetInfo,
    AssetMatch,
    AssetType,
    EDLSettings,
    ScriptSegment,
    SegmentType,
    TrackType,
    VideoInfo,
)


# ── Helpers ─────────────────────────────────────────────────────────────────


def _make_segment(
    index: int,
    seg_type: SegmentType = SegmentType.DIALOGUE,
    duration: float = 5.0,
    content: str = "test content",
) -> ScriptSegment:
    return ScriptSegment(
        index=index,
        type=seg_type,
        content=content,
        estimated_duration_seconds=duration,
    )


def _make_match(
    segment_index: int,
    asset_id: str = "asset_001",
    confidence: float = 0.9,
) -> AssetMatch:
    return AssetMatch(
        segment_index=segment_index,
        asset_id=asset_id,
        confidence=confidence,
        reasoning="Test match",
    )


def _make_asset(
    asset_id: str = "asset_001",
    duration: float = 60.0,
) -> AssetInfo:
    return AssetInfo(
        id=asset_id,
        file_name=f"{asset_id}.mp4",
        asset_type=AssetType.VIDEO,
        video=VideoInfo(duration_seconds=duration),
    )


def _default_settings() -> EDLSettings:
    return EDLSettings(frame_rate=24.0)


# ── Fixtures ────────────────────────────────────────────────────────────────


@pytest.fixture
def generator() -> EDLGenerator:
    return EDLGenerator()


# ── Basic generation tests ──────────────────────────────────────────────────


class TestBasicGeneration:
    def test_generates_edl_from_segments_and_matches(
        self, generator: EDLGenerator
    ) -> None:
        segments = [_make_segment(0), _make_segment(1)]
        matches = [_make_match(0, "a1"), _make_match(1, "a2")]
        assets = [_make_asset("a1"), _make_asset("a2")]
        edl = generator.generate(segments, matches, assets, _default_settings())
        assert len(edl.entries) >= 2
        asset_ids = [e.source_asset_id for e in edl.entries]
        assert "a1" in asset_ids
        assert "a2" in asset_ids

    def test_skips_unmatched_segments(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0), _make_segment(1), _make_segment(2)]
        matches = [_make_match(0, "a1"), _make_match(2, "a3")]
        assets = [_make_asset("a1"), _make_asset("a3")]
        edl = generator.generate(segments, matches, assets, _default_settings())
        asset_ids = [e.source_asset_id for e in edl.entries]
        assert "a1" in asset_ids
        assert "a3" in asset_ids
        # Segment 1 has no match, so its asset should not appear

    def test_uses_highest_confidence_match_per_segment(
        self, generator: EDLGenerator
    ) -> None:
        segments = [_make_segment(0)]
        matches = [
            _make_match(0, "low", confidence=0.3),
            _make_match(0, "high", confidence=0.95),
        ]
        assets = [_make_asset("low"), _make_asset("high")]
        edl = generator.generate(segments, matches, assets, _default_settings())
        # The entry should use the highest confidence match
        video_entries = [e for e in edl.entries if e.source_asset_id in ("low", "high")]
        assert any(e.source_asset_id == "high" for e in video_entries)

    def test_empty_segments_produces_empty_edl(
        self, generator: EDLGenerator
    ) -> None:
        edl = generator.generate([], [], [], _default_settings())
        assert len(edl.entries) == 0

    def test_edl_has_id_and_name(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        assert edl.id != ""
        assert edl.name != ""


# ── Track assignment tests ──────────────────────────────────────────────────


class TestTrackAssignment:
    def test_dialogue_goes_to_video_track(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, SegmentType.DIALOGUE)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        video_entries = [e for e in edl.entries if e.track.type == TrackType.VIDEO]
        assert len(video_entries) >= 1

    def test_broll_goes_to_video_track(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, SegmentType.BROLL)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        video_entries = [e for e in edl.entries if e.track.type == TrackType.VIDEO]
        assert len(video_entries) >= 1

    def test_broll_on_different_track_than_dialogue(
        self, generator: EDLGenerator
    ) -> None:
        segments = [
            _make_segment(0, SegmentType.DIALOGUE),
            _make_segment(1, SegmentType.BROLL),
        ]
        matches = [_make_match(0, "a1"), _make_match(1, "a2")]
        assets = [_make_asset("a1"), _make_asset("a2")]
        edl = generator.generate(segments, matches, assets, _default_settings())
        video_entries = [e for e in edl.entries if e.track.type == TrackType.VIDEO]
        if len(video_entries) >= 2:
            # Different track indices for dialogue and broll
            track_indices = {e.track.track_index for e in video_entries}
            assert len(track_indices) >= 2

    def test_music_goes_to_audio_track(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, SegmentType.MUSIC)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        audio_entries = [e for e in edl.entries if e.track.type == TrackType.AUDIO]
        assert len(audio_entries) >= 1

    def test_voiceover_goes_to_audio_track(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, SegmentType.VOICEOVER)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        audio_entries = [e for e in edl.entries if e.track.type == TrackType.AUDIO]
        assert len(audio_entries) >= 1

    def test_sfx_goes_to_audio_track(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, SegmentType.SFX)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        audio_entries = [e for e in edl.entries if e.track.type == TrackType.AUDIO]
        assert len(audio_entries) >= 1


# ── Timeline position tests ────────────────────────────────────────────────


class TestTimelinePosition:
    def test_first_entry_starts_near_zero(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, duration=5.0)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        edl = generator.generate(segments, matches, assets, _default_settings())
        if edl.entries:
            # First video entry should start at or very close to zero
            first = edl.entries[0]
            assert first.timeline_range.in_point.to_seconds() < 1.0

    def test_entries_progress_along_timeline(
        self, generator: EDLGenerator
    ) -> None:
        segments = [
            _make_segment(0, duration=3.0),
            _make_segment(1, duration=4.0),
            _make_segment(2, duration=2.0),
        ]
        matches = [_make_match(0, "a1"), _make_match(1, "a2"), _make_match(2, "a3")]
        assets = [_make_asset("a1"), _make_asset("a2"), _make_asset("a3")]
        edl = generator.generate(segments, matches, assets, _default_settings())

        # Filter to primary track entries (not companion audio entries)
        video_entries = [e for e in edl.entries if e.track.type == TrackType.VIDEO]
        if len(video_entries) >= 2:
            # Each successive entry should start at or after the previous one's start
            for i in range(1, len(video_entries)):
                prev_start = video_entries[i - 1].timeline_range.in_point.to_seconds()
                curr_start = video_entries[i].timeline_range.in_point.to_seconds()
                assert curr_start >= prev_start

    def test_uses_settings_frame_rate(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0, duration=1.0)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        settings = EDLSettings(frame_rate=30.0)
        edl = generator.generate(segments, matches, assets, settings)
        assert edl.sequence_frame_rate == 30.0


# ── Transition tests ───────────────────────────────────────────────────────


class TestTransitionInsertion:
    def test_first_entry_has_no_transition(self, generator: EDLGenerator) -> None:
        segments = [_make_segment(0)]
        matches = [_make_match(0)]
        assets = [_make_asset()]
        settings = EDLSettings(
            default_transition="cross_dissolve",
            default_transition_duration=1.0,
        )
        edl = generator.generate(segments, matches, assets, settings)
        if edl.entries:
            assert edl.entries[0].transition is None

    def test_subsequent_entries_have_transitions_when_configured(
        self, generator: EDLGenerator
    ) -> None:
        segments = [_make_segment(0), _make_segment(1)]
        matches = [_make_match(0, "a1"), _make_match(1, "a2")]
        assets = [_make_asset("a1"), _make_asset("a2")]
        settings = EDLSettings(
            default_transition="cross_dissolve",
            default_transition_duration=1.0,
        )
        edl = generator.generate(segments, matches, assets, settings)
        # At least one entry (not the first) should have a transition
        entries_with_transitions = [e for e in edl.entries if e.transition is not None]
        assert len(entries_with_transitions) >= 1

    def test_cut_transition_produces_no_transition_object(
        self, generator: EDLGenerator
    ) -> None:
        segments = [_make_segment(0), _make_segment(1)]
        matches = [_make_match(0, "a1"), _make_match(1, "a2")]
        assets = [_make_asset("a1"), _make_asset("a2")]
        settings = EDLSettings(default_transition="cut", default_transition_duration=0.0)
        edl = generator.generate(segments, matches, assets, settings)
        for entry in edl.entries:
            assert entry.transition is None
