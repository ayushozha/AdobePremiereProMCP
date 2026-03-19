"""Narration / voiceover script parser.

Parses simple narration scripts where the primary content is spoken
voiceover text, interspersed with bracketed visual and audio cues:

- ``[Visual: description]`` — B-roll or visual content
- ``[Title: text]`` — Text overlay / title card
- ``[Music: description]`` — Music cue
- ``[SFX: description]`` — Sound effect cue
- ``[Transition: type]`` — Transition cue
- ``[Lower Third: text]`` — Lower third overlay
- Plain text — Voiceover narration
"""

from __future__ import annotations

import re

from src.models import ScriptSegment, SegmentType

# ── Regex patterns for bracketed cues ────────────────────────────────────────

# Generic bracketed cue: [TYPE: content]
_CUE_RE = re.compile(
    r"\[(\w[\w\s]*?)\s*:\s*(.+?)\]",
    re.IGNORECASE,
)

# Map cue type names to segment types
_CUE_TYPE_MAP: dict[str, SegmentType] = {
    "visual": SegmentType.BROLL,
    "b-roll": SegmentType.BROLL,
    "broll": SegmentType.BROLL,
    "b roll": SegmentType.BROLL,
    "video": SegmentType.BROLL,
    "title": SegmentType.TITLE,
    "text": SegmentType.TITLE,
    "lower third": SegmentType.LOWER_THIRD,
    "lower-third": SegmentType.LOWER_THIRD,
    "music": SegmentType.MUSIC,
    "sfx": SegmentType.SFX,
    "sound": SegmentType.SFX,
    "sound effect": SegmentType.SFX,
    "transition": SegmentType.TRANSITION,
    "fx": SegmentType.SFX,
}


def _resolve_cue_type(cue_name: str) -> SegmentType:
    """Map a bracketed cue name to a SegmentType."""
    return _CUE_TYPE_MAP.get(cue_name.lower().strip(), SegmentType.BROLL)


def _make_cue_segment(
    idx: int,
    seg_type: SegmentType,
    content: str,
) -> ScriptSegment:
    """Build a ScriptSegment for a bracketed cue."""
    visual = ""
    audio = ""
    if seg_type in (SegmentType.BROLL, SegmentType.TITLE, SegmentType.LOWER_THIRD):
        visual = content
    elif seg_type in (SegmentType.MUSIC, SegmentType.SFX):
        audio = content

    return ScriptSegment(
        index=idx,
        type=seg_type,
        content=content,
        visual_direction=visual,
        audio_direction=audio,
    )


def parse_narration(text: str) -> list[ScriptSegment]:
    """Parse narration/voiceover script text into segments.

    The parser scans each line (or multi-line paragraph) for bracketed cues.
    Text that falls outside any bracket is treated as voiceover narration.
    Cues can appear inline (mid-paragraph) or on their own line.

    Args:
        text: Raw narration script text.

    Returns:
        Ordered list of ``ScriptSegment`` objects.
    """
    segments: list[ScriptSegment] = []
    idx = 0

    # Process the text paragraph by paragraph (split on blank lines).
    paragraphs = re.split(r"\n\s*\n", text)

    for paragraph in paragraphs:
        paragraph = paragraph.strip()
        if not paragraph:
            continue

        # Check if the entire paragraph is a single bracketed cue (common pattern)
        full_match = re.match(r"^\s*\[(\w[\w\s]*?)\s*:\s*(.+?)\]\s*$", paragraph, re.DOTALL)
        if full_match:
            cue_name = full_match.group(1)
            cue_content = full_match.group(2).strip()
            seg_type = _resolve_cue_type(cue_name)
            segments.append(_make_cue_segment(idx, seg_type, cue_content))
            idx += 1
            continue

        # Otherwise, the paragraph may contain inline cues mixed with narration.
        # Split the paragraph on cue boundaries.
        parts = _CUE_RE.split(paragraph)

        # re.split with groups: [before, cue_type, cue_content, between, type, content, ...]
        # parts[0] is text before first cue, then groups of 3: (cue_type, cue_content, text_after)
        i = 0
        while i < len(parts):
            if i == 0 or (i > 0 and (i - 1) % 2 == 0 and i % 2 == 0):
                # This is a plain-text portion (narration).
                narration_text = parts[i].strip()
                if narration_text:
                    # Clean up any leftover brackets or whitespace
                    narration_text = re.sub(r"\s+", " ", narration_text)
                    segments.append(ScriptSegment(
                        index=idx,
                        type=SegmentType.VOICEOVER,
                        content=narration_text,
                    ))
                    idx += 1
                i += 1
            elif i + 1 < len(parts):
                # Cue type and content pair
                cue_name = parts[i]
                cue_content = parts[i + 1].strip()
                seg_type = _resolve_cue_type(cue_name)
                segments.append(_make_cue_segment(idx, seg_type, cue_content))
                idx += 1
                i += 2
            else:
                # Orphan part (shouldn't normally happen)
                leftover = parts[i].strip()
                if leftover:
                    segments.append(ScriptSegment(
                        index=idx,
                        type=SegmentType.VOICEOVER,
                        content=leftover,
                    ))
                    idx += 1
                i += 1

    return segments
