"""Tests for the script parser module.

Covers YouTube format, narration format, duration estimation,
and auto-detection of format.
"""

from __future__ import annotations

import pytest

from src.models import ScriptFormat, SegmentType
from src.parser import ScriptParser


# ── Fixtures ────────────────────────────────────────────────────────────────


@pytest.fixture
def parser() -> ScriptParser:
    return ScriptParser()


YOUTUBE_SCRIPT = """\
[INTRO]
ON CAMERA: Welcome to today's video about wildlife photography.
B-ROLL: sweeping aerial shots of African savanna

[MAIN]
ON CAMERA: Let me show you the top five tips for getting great shots.
B-ROLL: close-up of a lion in golden hour light
We always want to make sure the lighting is right.

[OUTRO]
ON CAMERA: Thanks for watching, and don't forget to subscribe.
"""

NARRATION_SCRIPT = """\
[Title: The Art of Cooking]

[Music: soft piano background]

Welcome to our cooking show where we explore amazing recipes.

[Visual: overhead shot of ingredients on a wooden table]

Today we will learn how to make a perfect sourdough bread.

[Visual: close-up of dough being kneaded]

The key is patience and proper fermentation.
"""


# ── YouTube format tests ────────────────────────────────────────────────────


class TestYouTubeFormat:
    def test_parses_section_markers(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        assert result.metadata.format == ScriptFormat.YOUTUBE
        assert len(result.segments) > 0

    def test_detects_broll_segments(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        broll_segments = [s for s in result.segments if s.type == SegmentType.BROLL]
        assert len(broll_segments) >= 2
        # B-roll should have visual direction populated
        for seg in broll_segments:
            assert seg.visual_direction != ""

    def test_detects_on_camera_segments(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        dialogue_segments = [s for s in result.segments if s.type == SegmentType.DIALOGUE]
        assert len(dialogue_segments) >= 3

    def test_broll_has_asset_hints(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        broll = [s for s in result.segments if s.type == SegmentType.BROLL]
        assert len(broll) > 0
        # The B-roll description should be parsed into keywords
        first_broll = broll[0]
        assert len(first_broll.asset_hints) > 0

    def test_segment_indices_are_sequential(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        indices = [s.index for s in result.segments]
        assert indices == list(range(len(result.segments)))

    def test_metadata_segment_count(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        assert result.metadata.segment_count == len(result.segments)


# ── Narration format tests ──────────────────────────────────────────────────


class TestNarrationFormat:
    def test_parses_visual_directives(self, parser: ScriptParser) -> None:
        result = parser.parse(NARRATION_SCRIPT, format_hint="narration")
        broll = [s for s in result.segments if s.type == SegmentType.BROLL]
        assert len(broll) == 2

    def test_parses_title_directive(self, parser: ScriptParser) -> None:
        result = parser.parse(NARRATION_SCRIPT, format_hint="narration")
        titles = [s for s in result.segments if s.type == SegmentType.TITLE]
        assert len(titles) == 1
        assert "cooking" in titles[0].content.lower()

    def test_parses_music_directive(self, parser: ScriptParser) -> None:
        result = parser.parse(NARRATION_SCRIPT, format_hint="narration")
        music = [s for s in result.segments if s.type == SegmentType.MUSIC]
        assert len(music) == 1

    def test_narration_text_becomes_voiceover(self, parser: ScriptParser) -> None:
        result = parser.parse(NARRATION_SCRIPT, format_hint="narration")
        vo_segments = [s for s in result.segments if s.type == SegmentType.VOICEOVER]
        # The narration parser should produce VO segments for non-cue text
        assert len(vo_segments) >= 1

    def test_narration_metadata_format(self, parser: ScriptParser) -> None:
        result = parser.parse(NARRATION_SCRIPT, format_hint="narration")
        assert result.metadata.format == ScriptFormat.NARRATION


# ── Duration estimation tests ───────────────────────────────────────────────


class TestDurationEstimation:
    def test_speech_duration_scales_with_word_count(self, parser: ScriptParser) -> None:
        short_text = "ON CAMERA: Hello world"
        long_text = (
            "ON CAMERA: This is a much longer sentence with many more words "
            "to speak aloud during the recording session"
        )
        short_result = parser.parse(short_text, format_hint="youtube")
        long_result = parser.parse(long_text, format_hint="youtube")
        short_dur = short_result.segments[0].estimated_duration_seconds
        long_dur = long_result.segments[0].estimated_duration_seconds
        assert long_dur > short_dur

    def test_speech_duration_is_positive(self, parser: ScriptParser) -> None:
        result = parser.parse("ON CAMERA: Hi there", format_hint="youtube")
        assert result.segments[0].estimated_duration_seconds > 0

    def test_total_duration_is_sum_of_segments(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="youtube")
        expected_total = sum(s.estimated_duration_seconds for s in result.segments)
        assert abs(result.metadata.estimated_total_duration_seconds - expected_total) < 0.01


# ── Auto-detection tests ───────────────────────────────────────────────────


class TestAutoDetection:
    def test_detects_youtube_format(self, parser: ScriptParser) -> None:
        result = parser.parse(YOUTUBE_SCRIPT, format_hint="auto")
        assert result.metadata.format == ScriptFormat.YOUTUBE

    def test_detects_narration_format(self, parser: ScriptParser) -> None:
        result = parser.parse(NARRATION_SCRIPT, format_hint="auto")
        assert result.metadata.format == ScriptFormat.NARRATION

    def test_plain_text_produces_segments(self, parser: ScriptParser) -> None:
        plain = "Just some plain text that someone might speak on camera."
        result = parser.parse(plain, format_hint="auto")
        # Should produce at least one segment regardless of detected format
        assert len(result.segments) >= 1
