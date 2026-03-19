"""Tests for pacing analysis and rhythm detection.

Covers mood-based duration targets, dialogue segment protection,
and rhythm pattern detection/suggestion.
"""

from __future__ import annotations

import pytest

from src.analysis.pacing import PacingAnalyzer
from src.analysis.rhythm import RhythmAnalyzer
from src.models import (
    EDLEntry,
    EditDecisionList,
    Timecode,
    TimeRange,
    TrackTarget,
    TrackType,
)


# ── Helpers ─────────────────────────────────────────────────────────────────


def _make_entry(
    index: int,
    start_sec: float,
    end_sec: float,
    seg_type_name: str = "BROLL",
    frame_rate: float = 24.0,
) -> EDLEntry:
    """Create an EDL entry with the given timeline range and segment type."""
    return EDLEntry(
        index=index,
        source_asset_id=f"asset_{index}",
        timeline_range=TimeRange(
            in_point=_sec_to_tc(start_sec, frame_rate),
            out_point=_sec_to_tc(end_sec, frame_rate),
        ),
        track=TrackTarget(type=TrackType.VIDEO, track_index=1),
        notes=f"Segment {index}: {seg_type_name}",
    )


def _sec_to_tc(seconds: float, frame_rate: float = 24.0) -> Timecode:
    total = int(seconds)
    frac = seconds - total
    return Timecode(
        hours=total // 3600,
        minutes=(total % 3600) // 60,
        seconds=total % 60,
        frames=int(frac * frame_rate),
        frame_rate=frame_rate,
    )


def _make_edl(entries: list[EDLEntry]) -> EditDecisionList:
    return EditDecisionList(
        id="test-edl",
        name="Test EDL",
        entries=entries,
    )


# ── Fixtures ────────────────────────────────────────────────────────────────


@pytest.fixture
def pacing() -> PacingAnalyzer:
    return PacingAnalyzer()


@pytest.fixture
def rhythm() -> RhythmAnalyzer:
    return RhythmAnalyzer()


# ── Mood-based duration target tests ────────────────────────────────────────


class TestMoodTargets:
    def test_all_mood_targets_are_positive(self, pacing: PacingAnalyzer) -> None:
        for mood, target in pacing.MOOD_TARGETS.items():
            assert target > 0, f"Mood '{mood}' has non-positive target {target}"

    def test_energetic_has_shortest_target(self, pacing: PacingAnalyzer) -> None:
        assert pacing.MOOD_TARGETS["energetic"] < pacing.MOOD_TARGETS["calm"]
        assert pacing.MOOD_TARGETS["energetic"] < pacing.MOOD_TARGETS["cinematic"]

    def test_calm_has_longest_target(self, pacing: PacingAnalyzer) -> None:
        assert pacing.MOOD_TARGETS["calm"] == max(pacing.MOOD_TARGETS.values())

    def test_analyze_returns_suggested_avg_closer_to_target(
        self, pacing: PacingAnalyzer
    ) -> None:
        # All entries are 10 s (long), target "energetic" is 2.5 s
        entries = [
            _make_entry(0, 0, 10, "BROLL"),
            _make_entry(1, 10, 20, "BROLL"),
            _make_entry(2, 20, 30, "BROLL"),
        ]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="energetic")

        # Suggested average should be lower than the current 10 s
        assert result.suggested_avg_clip_duration < result.current_avg_clip_duration
        # But we apply damped adjustments, so it won't jump all the way to 2.5
        assert result.suggested_avg_clip_duration < 10.0

    def test_analyze_extends_short_clips_for_calm_mood(
        self, pacing: PacingAnalyzer
    ) -> None:
        # All entries are 2 s, target "calm" is 8.0 s
        entries = [
            _make_entry(0, 0, 2, "BROLL"),
            _make_entry(1, 2, 4, "BROLL"),
        ]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="calm")

        assert result.suggested_avg_clip_duration > result.current_avg_clip_duration

    def test_unknown_mood_falls_back_to_cinematic(
        self, pacing: PacingAnalyzer
    ) -> None:
        entries = [_make_entry(0, 0, 10, "BROLL")]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="nonexistent_mood")
        # Should use cinematic target (5.0) and adjust the 10 s clip down
        assert result.suggested_avg_clip_duration < 10.0

    def test_empty_edl_returns_empty_result(self, pacing: PacingAnalyzer) -> None:
        edl = _make_edl([])
        result = pacing.analyze(edl, target_mood="dramatic")
        assert len(result.adjustments) == 0
        assert result.current_avg_clip_duration == 0.0

    def test_each_entry_gets_an_adjustment(self, pacing: PacingAnalyzer) -> None:
        entries = [
            _make_entry(0, 0, 5, "BROLL"),
            _make_entry(1, 5, 12, "BROLL"),
            _make_entry(2, 12, 15, "BROLL"),
        ]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="cinematic")
        assert len(result.adjustments) == 3


# ── Dialogue protection tests ───────────────────────────────────────────────


class TestDialogueProtection:
    def test_dialogue_not_shortened(self, pacing: PacingAnalyzer) -> None:
        """Dialogue segments should never be trimmed shorter than their current duration."""
        entries = [
            _make_entry(0, 0, 12, "DIALOGUE"),  # 12 s of dialogue
            _make_entry(1, 12, 20, "BROLL"),     # 8 s of B-roll
        ]
        edl = _make_edl(entries)
        # energetic target is 2.5 s — the dialogue must NOT be cut to 2.5
        result = pacing.analyze(edl, target_mood="energetic")

        dialogue_adj = result.adjustments[0]
        assert dialogue_adj.suggested_duration >= dialogue_adj.current_duration

    def test_voiceover_not_shortened(self, pacing: PacingAnalyzer) -> None:
        entries = [
            _make_entry(0, 0, 8, "VOICEOVER"),
        ]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="energetic")

        vo_adj = result.adjustments[0]
        assert vo_adj.suggested_duration >= vo_adj.current_duration

    def test_broll_can_be_shortened(self, pacing: PacingAnalyzer) -> None:
        entries = [
            _make_entry(0, 0, 15, "BROLL"),
        ]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="energetic")

        broll_adj = result.adjustments[0]
        assert broll_adj.suggested_duration < broll_adj.current_duration

    def test_adjustment_has_reason(self, pacing: PacingAnalyzer) -> None:
        entries = [_make_entry(0, 0, 10, "BROLL")]
        edl = _make_edl(entries)
        result = pacing.analyze(edl, target_mood="energetic")
        assert result.adjustments[0].reason != ""


# ── Rhythm detection tests ──────────────────────────────────────────────────


class TestRhythmDetection:
    def test_constant_pattern(self, rhythm: RhythmAnalyzer) -> None:
        durations = [5.0, 5.0, 5.0, 5.0, 5.0]
        assert rhythm.detect_pattern(durations) == "constant"

    def test_accelerating_pattern(self, rhythm: RhythmAnalyzer) -> None:
        # Durations getting shorter = accelerating edits
        durations = [8.0, 6.0, 4.5, 3.0, 1.5]
        assert rhythm.detect_pattern(durations) == "accelerating"

    def test_decelerating_pattern(self, rhythm: RhythmAnalyzer) -> None:
        # Durations getting longer = decelerating edits
        durations = [1.5, 3.0, 4.5, 6.0, 8.0]
        assert rhythm.detect_pattern(durations) == "decelerating"

    def test_alternating_pattern(self, rhythm: RhythmAnalyzer) -> None:
        durations = [2.0, 6.0, 2.0, 6.0, 2.0, 6.0]
        assert rhythm.detect_pattern(durations) == "alternating"

    def test_single_clip_is_constant(self, rhythm: RhythmAnalyzer) -> None:
        assert rhythm.detect_pattern([5.0]) == "constant"

    def test_two_clips(self, rhythm: RhythmAnalyzer) -> None:
        # With only 2 clips, pattern detection is limited
        result = rhythm.detect_pattern([5.0, 5.0])
        assert result in ("constant", "random")


# ── Rhythm suggestion tests ────────────────────────────────────────────────


class TestRhythmSuggestion:
    def test_dramatic_creates_accelerating_sequence(
        self, rhythm: RhythmAnalyzer
    ) -> None:
        durations = rhythm.suggest_rhythm("dramatic", 5)
        assert len(durations) == 5
        # Should generally trend shorter (accelerate toward climax)
        assert durations[0] > durations[-1]

    def test_energetic_produces_short_durations(
        self, rhythm: RhythmAnalyzer
    ) -> None:
        durations = rhythm.suggest_rhythm("energetic", 4)
        assert len(durations) == 4
        assert all(d <= 4.0 for d in durations)

    def test_calm_produces_long_durations(self, rhythm: RhythmAnalyzer) -> None:
        durations = rhythm.suggest_rhythm("calm", 4)
        assert len(durations) == 4
        assert all(d >= 5.0 for d in durations)

    def test_zero_segments_returns_empty(self, rhythm: RhythmAnalyzer) -> None:
        assert rhythm.suggest_rhythm("dramatic", 0) == []

    def test_single_segment(self, rhythm: RhythmAnalyzer) -> None:
        durations = rhythm.suggest_rhythm("cinematic", 1)
        assert len(durations) == 1
        assert durations[0] > 0

    def test_all_suggested_durations_are_positive(
        self, rhythm: RhythmAnalyzer
    ) -> None:
        for mood in ("dramatic", "energetic", "calm", "comedic", "documentary", "cinematic"):
            durations = rhythm.suggest_rhythm(mood, 8)
            assert all(d > 0 for d in durations), f"Mood '{mood}' produced non-positive duration"
