"""Duration estimation for script segments.

Estimates how long each segment will take based on its type, word count,
and content characteristics. These are rough estimates intended to provide
a starting point for timeline layout; editors will refine them.
"""

from __future__ import annotations

import re

from src.models import ScriptSegment, SegmentType

# ── Constants ────────────────────────────────────────────────────────────────

# Average speaking rates (words per minute).
_DIALOGUE_WPM: float = 150.0
_VOICEOVER_WPM: float = 150.0

# Minimum durations (seconds) for non-speech segments.
_MIN_BROLL_DURATION: float = 3.0
_MIN_TITLE_DURATION: float = 3.0
_MAX_TITLE_DURATION: float = 5.0
_TRANSITION_DURATION_SHORT: float = 1.0
_TRANSITION_DURATION_LONG: float = 2.0
_SFX_DEFAULT_DURATION: float = 2.0
_MUSIC_DEFAULT_DURATION: float = 5.0

# For B-roll, each descriptive clause adds roughly this many seconds.
_BROLL_SECONDS_PER_CLAUSE: float = 2.0

# Characters per second for reading on-screen text.
_TITLE_CHARS_PER_SECOND: float = 15.0


def _word_count(text: str) -> int:
    """Return the number of whitespace-delimited words in *text*."""
    return len(text.split())


def _extract_explicit_duration(text: str) -> float | None:
    """Try to extract an explicit duration from text like '(5 seconds)' or '10s'.

    Returns the duration in seconds if found, otherwise ``None``.
    """
    # Match patterns like "5 seconds", "10s", "3 sec", "2.5 seconds"
    match = re.search(
        r"(\d+(?:\.\d+)?)\s*(?:seconds?|secs?|s)\b",
        text,
        re.IGNORECASE,
    )
    if match:
        return float(match.group(1))
    return None


def _estimate_speech_duration(text: str, wpm: float) -> float:
    """Estimate duration for spoken content based on word count."""
    words = _word_count(text)
    if words == 0:
        return 1.0
    return (words / wpm) * 60.0


def _estimate_broll_duration(text: str) -> float:
    """Estimate B-roll duration from description complexity.

    More complex descriptions (multiple clauses, adjectives) suggest longer
    shots. We use comma/semicolon-separated clause count as a proxy.
    """
    explicit = _extract_explicit_duration(text)
    if explicit is not None:
        return max(explicit, _MIN_BROLL_DURATION)

    # Count clauses (separated by commas, semicolons, or 'and'/'then')
    clauses = re.split(r"[,;]|\band\b|\bthen\b", text)
    clause_count = max(len(clauses), 1)

    return max(clause_count * _BROLL_SECONDS_PER_CLAUSE, _MIN_BROLL_DURATION)


def _estimate_title_duration(text: str) -> float:
    """Estimate how long a title/text overlay should be displayed."""
    explicit = _extract_explicit_duration(text)
    if explicit is not None:
        return max(explicit, _MIN_TITLE_DURATION)

    char_count = len(text.strip())
    if char_count == 0:
        return _MIN_TITLE_DURATION

    reading_time = char_count / _TITLE_CHARS_PER_SECOND
    return max(_MIN_TITLE_DURATION, min(reading_time, _MAX_TITLE_DURATION))


def _estimate_transition_duration(text: str) -> float:
    """Estimate transition duration. Longer transitions for dissolves, shorter for cuts."""
    lower = text.lower()
    if any(word in lower for word in ("dissolve", "fade", "slow")):
        return _TRANSITION_DURATION_LONG
    return _TRANSITION_DURATION_SHORT


def _estimate_music_duration(text: str) -> float:
    """Estimate music cue duration."""
    explicit = _extract_explicit_duration(text)
    if explicit is not None:
        return explicit
    return _MUSIC_DEFAULT_DURATION


# ── Public API ───────────────────────────────────────────────────────────────


def estimate_duration(segment: ScriptSegment) -> float:
    """Estimate the duration of a script segment in seconds.

    The estimation strategy depends on the segment type:

    - **Dialogue / Voiceover**: Word count divided by speaking rate (~150 WPM).
    - **B-roll**: Minimum 3 s, scaled by description complexity.
    - **Title / Lower third**: 3-5 s based on text length.
    - **Transition**: 1-2 s depending on type (cut vs. dissolve).
    - **Music**: Extracted from text if specified, otherwise 5 s default.
    - **SFX**: 2 s default unless explicit duration is given.
    - **Action**: Treated like dialogue for word-count estimation.

    Args:
        segment: The script segment to estimate.

    Returns:
        Estimated duration in seconds (always > 0).
    """
    text = segment.content or segment.visual_direction or segment.audio_direction or ""

    match segment.type:
        case SegmentType.DIALOGUE:
            return _estimate_speech_duration(text, _DIALOGUE_WPM)
        case SegmentType.VOICEOVER:
            return _estimate_speech_duration(text, _VOICEOVER_WPM)
        case SegmentType.ACTION:
            return _estimate_speech_duration(text, _DIALOGUE_WPM)
        case SegmentType.BROLL:
            return _estimate_broll_duration(text)
        case SegmentType.TITLE | SegmentType.LOWER_THIRD:
            return _estimate_title_duration(text)
        case SegmentType.TRANSITION:
            return _estimate_transition_duration(text)
        case SegmentType.MUSIC:
            return _estimate_music_duration(text)
        case SegmentType.SFX:
            explicit = _extract_explicit_duration(text)
            return explicit if explicit is not None else _SFX_DEFAULT_DURATION
        case _:
            # UNSPECIFIED or unknown — fall back to speech estimate.
            return _estimate_speech_duration(text, _DIALOGUE_WPM) if text.strip() else 2.0
