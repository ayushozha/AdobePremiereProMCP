"""Tests for asset matching.

Covers keyword matching, text normalization, scoring combination,
and unmatched segment suggestions.
"""

from __future__ import annotations

import pytest

from src.matching.keyword_matcher import KeywordMatcher
from src.matching.scoring import ScoredMatch, combine_scores, cosine_similarity, normalize_text
from src.matching.suggest import suggest_assets
from src.models import (
    AssetInfo,
    AssetType,
    ScriptSegment,
    SegmentType,
)


# ── Fixtures ────────────────────────────────────────────────────────────────


@pytest.fixture
def keyword_matcher() -> KeywordMatcher:
    return KeywordMatcher()


def _make_segment(
    index: int = 0,
    seg_type: SegmentType = SegmentType.BROLL,
    content: str = "sunset over the ocean",
    visual_direction: str = "",
    asset_hints: list[str] | None = None,
) -> ScriptSegment:
    return ScriptSegment(
        index=index,
        type=seg_type,
        content=content,
        visual_direction=visual_direction,
        asset_hints=asset_hints or [],
    )


def _make_asset(
    asset_id: str = "a1",
    file_name: str = "sunset_beach.mp4",
    metadata: dict[str, str] | None = None,
) -> AssetInfo:
    return AssetInfo(
        id=asset_id,
        file_name=file_name,
        asset_type=AssetType.VIDEO,
        metadata=metadata or {},
    )


# ── Keyword matching tests ─────────────────────────────────────────────────


class TestKeywordMatching:
    def test_matches_by_filename_overlap(
        self, keyword_matcher: KeywordMatcher
    ) -> None:
        segment = _make_segment(content="sunset on the beach")
        assets = [
            _make_asset("a1", "sunset_beach_001.mp4"),
            _make_asset("a2", "city_traffic.mp4"),
        ]
        results = keyword_matcher.match(segment, assets)
        assert len(results) > 0
        # The sunset_beach asset should score higher than city_traffic
        a1_scores = [r for r in results if r.asset_id == "a1"]
        a2_scores = [r for r in results if r.asset_id == "a2"]
        if a1_scores and a2_scores:
            assert a1_scores[0].score > a2_scores[0].score

    def test_returns_empty_for_no_overlap(
        self, keyword_matcher: KeywordMatcher
    ) -> None:
        segment = _make_segment(content="quantum physics lecture")
        assets = [_make_asset("a1", "sunset_beach.mp4")]
        results = keyword_matcher.match(segment, assets)
        # Either empty or very low scores
        for r in results:
            assert r.score < 0.5

    def test_exact_hint_match_boosts_score(
        self, keyword_matcher: KeywordMatcher
    ) -> None:
        segment_no_hints = _make_segment(
            content="sunset over the ocean",
            asset_hints=[],
        )
        segment_with_hints = _make_segment(
            content="sunset over the ocean",
            asset_hints=["sunset", "ocean"],
        )
        assets = [_make_asset("a1", "sunset_ocean_clip.mp4")]

        results_no = keyword_matcher.match(segment_no_hints, assets)
        results_with = keyword_matcher.match(segment_with_hints, assets)

        score_no = results_no[0].score if results_no else 0.0
        score_with = results_with[0].score if results_with else 0.0
        # Having explicit hints should boost or at least not lower the score
        assert score_with >= score_no

    def test_reasoning_contains_overlapping_keywords(
        self, keyword_matcher: KeywordMatcher
    ) -> None:
        segment = _make_segment(content="sunset on the beach")
        assets = [_make_asset("a1", "sunset_beach_clip.mp4")]
        results = keyword_matcher.match(segment, assets)
        assert len(results) > 0
        assert "sunset" in results[0].reasoning.lower() or "beach" in results[0].reasoning.lower()

    def test_sorted_by_descending_score(
        self, keyword_matcher: KeywordMatcher
    ) -> None:
        segment = _make_segment(content="mountain sunset landscape")
        assets = [
            _make_asset("a1", "sunset_mountain.mp4"),
            _make_asset("a2", "mountain_landscape_sunset_view.mp4"),
            _make_asset("a3", "random_clip.mp4"),
        ]
        results = keyword_matcher.match(segment, assets)
        scores = [r.score for r in results]
        assert scores == sorted(scores, reverse=True)


# ── Text normalization tests ───────────────────────────────────────────────


class TestNormalizeText:
    def test_lowercases(self) -> None:
        tokens = normalize_text("HELLO World")
        assert all(t == t.lower() for t in tokens)

    def test_strips_file_extensions(self) -> None:
        tokens = normalize_text("sunset_beach.mp4")
        assert "mp4" not in tokens
        assert "sunset" in tokens

    def test_splits_camelcase(self) -> None:
        tokens = normalize_text("sunsetBeach")
        assert "sunset" in tokens
        assert "beach" in tokens

    def test_splits_underscores_and_hyphens(self) -> None:
        tokens = normalize_text("sunset_beach-clip")
        assert "sunset" in tokens
        assert "beach" in tokens
        assert "clip" in tokens

    def test_deduplicates_tokens(self) -> None:
        tokens = normalize_text("sunset sunset sunset")
        assert tokens.count("sunset") == 1

    def test_handles_broll_naming(self) -> None:
        tokens = normalize_text("B-roll_sunset_001")
        # Should produce tokens without losing meaningful words
        assert any("sunset" in t for t in tokens)


# ── Score combination tests ─────────────────────────────────────────────────


class TestScoreCombination:
    def test_combine_scores_weighted_average(self) -> None:
        result = combine_scores(1.0, 0.0, keyword_weight=0.3)
        assert result == pytest.approx(0.3)

    def test_combine_scores_equal_weights(self) -> None:
        result = combine_scores(0.8, 0.8, keyword_weight=0.5)
        assert result == pytest.approx(0.8)

    def test_combine_scores_embedding_dominant(self) -> None:
        result = combine_scores(0.0, 1.0, keyword_weight=0.3)
        assert result == pytest.approx(0.7)

    def test_cosine_similarity_identical_vectors(self) -> None:
        vec = [1.0, 2.0, 3.0]
        assert cosine_similarity(vec, vec) == pytest.approx(1.0)

    def test_cosine_similarity_orthogonal_vectors(self) -> None:
        a = [1.0, 0.0]
        b = [0.0, 1.0]
        assert cosine_similarity(a, b) == pytest.approx(0.0)

    def test_cosine_similarity_zero_vector(self) -> None:
        a = [0.0, 0.0]
        b = [1.0, 2.0]
        assert cosine_similarity(a, b) == 0.0

    def test_cosine_similarity_mismatched_lengths_raises(self) -> None:
        with pytest.raises(ValueError):
            cosine_similarity([1.0], [1.0, 2.0])


# ── Unmatched segment suggestion tests ─────────────────────────────────────


class TestUnmatchedSuggestions:
    def test_broll_suggests_footage(self) -> None:
        seg = _make_segment(
            seg_type=SegmentType.BROLL,
            content="aerial view of city at night",
            visual_direction="aerial view of city at night",
        )
        suggestions = suggest_assets(seg)
        assert len(suggestions) > 0
        combined = " ".join(suggestions).lower()
        assert "b-roll" in combined or "footage" in combined or "stock" in combined

    def test_music_suggests_audio(self) -> None:
        seg = ScriptSegment(
            index=0,
            type=SegmentType.MUSIC,
            content="upbeat electronic track",
            audio_direction="upbeat electronic",
        )
        suggestions = suggest_assets(seg)
        combined = " ".join(suggestions).lower()
        assert "music" in combined

    def test_voiceover_suggests_recording(self) -> None:
        seg = ScriptSegment(
            index=0,
            type=SegmentType.VOICEOVER,
            content="narrator explaining the process",
        )
        suggestions = suggest_assets(seg)
        combined = " ".join(suggestions).lower()
        assert "voiceover" in combined or "recording" in combined

    def test_suggestions_include_duration_hint(self) -> None:
        seg = ScriptSegment(
            index=0,
            type=SegmentType.BROLL,
            content="time-lapse of clouds",
            estimated_duration_seconds=10.0,
        )
        suggestions = suggest_assets(seg)
        combined = " ".join(suggestions).lower()
        assert "duration" in combined or "10" in combined

    def test_warns_when_asset_type_missing(self) -> None:
        seg = ScriptSegment(
            index=0,
            type=SegmentType.MUSIC,
            content="piano background",
        )
        # available_types does NOT include "audio"
        suggestions = suggest_assets(seg, available_types=["VIDEO"])
        combined = " ".join(suggestions).lower()
        assert "no " in combined or "import" in combined
