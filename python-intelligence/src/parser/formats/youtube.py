"""YouTube script format parser.

Parses YouTube-style video scripts which typically include:
- Section markers: [INTRO], [HOOK], [MAIN], [OUTRO], [SECTION: title]
- Visual cues: B-ROLL: description, ON CAMERA: dialogue
- Text overlays: TEXT ON SCREEN: content
- Music cues: MUSIC: description
- Optional timestamps: 0:00 - 0:30 or (0:00)
- Casual tone with visual/audio notes interspersed
"""

from __future__ import annotations

import re

from src.models import ScriptSegment, SegmentType

# ── Regex patterns ───────────────────────────────────────────────────────────

# Section markers: [INTRO], [HOOK], [MAIN CONTENT], [OUTRO], [SECTION: Title]
_SECTION_RE = re.compile(
    r"^\s*\[(\w[\w\s]*?)(?::\s*(.+?))?\]\s*$",
    re.IGNORECASE,
)

# Timestamp ranges: 0:00 - 0:30, 1:30-2:45, (0:00 - 0:30)
_TIMESTAMP_RANGE_RE = re.compile(
    r"^\s*\(?\s*(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})\s*\)?\s*$"
)

# Inline timestamp prefix: (0:00) or 0:00 -
_TIMESTAMP_PREFIX_RE = re.compile(
    r"^\s*\(?(\d{1,2}:\d{2})\)?\s*[-:]?\s*"
)

# B-ROLL: description
_BROLL_RE = re.compile(r"^\s*B-?ROLL\s*:\s*(.+)", re.IGNORECASE)

# ON CAMERA: dialogue / CTA (call to action)
_ON_CAMERA_RE = re.compile(r"^\s*ON\s+CAMERA\s*:\s*(.+)", re.IGNORECASE)

# TEXT ON SCREEN: overlay text
_TEXT_SCREEN_RE = re.compile(r"^\s*TEXT\s+ON\s+SCREEN\s*:\s*(.+)", re.IGNORECASE)

# LOWER THIRD: name/title overlay
_LOWER_THIRD_RE = re.compile(r"^\s*LOWER\s+THIRD\s*:\s*(.+)", re.IGNORECASE)

# MUSIC: cue
_MUSIC_RE = re.compile(r"^\s*MUSIC\s*:\s*(.+)", re.IGNORECASE)

# SFX: sound effect
_SFX_RE = re.compile(r"^\s*(?:SFX|SOUND\s+EFFECT)\s*:\s*(.+)", re.IGNORECASE)

# TRANSITION: type
_TRANSITION_RE = re.compile(
    r"^\s*(?:TRANSITION|CUT|FADE)\s*:\s*(.+)",
    re.IGNORECASE,
)

# VO: or VOICEOVER: narration line
_VO_RE = re.compile(r"^\s*(?:VO|VOICEOVER|V\.O\.)\s*:\s*(.+)", re.IGNORECASE)

# CTA: Call to action (common YouTube element)
_CTA_RE = re.compile(r"^\s*CTA\s*:\s*(.+)", re.IGNORECASE)

# Known section names mapped to SegmentType
_SECTION_TYPES: dict[str, SegmentType] = {
    "intro": SegmentType.DIALOGUE,
    "hook": SegmentType.DIALOGUE,
    "main": SegmentType.DIALOGUE,
    "main content": SegmentType.DIALOGUE,
    "outro": SegmentType.DIALOGUE,
    "cta": SegmentType.TITLE,
    "call to action": SegmentType.TITLE,
    "end screen": SegmentType.TITLE,
    "end card": SegmentType.TITLE,
}


def _parse_timestamp(ts: str) -> float:
    """Convert a M:SS or MM:SS string to seconds."""
    parts = ts.split(":")
    if len(parts) == 2:
        return int(parts[0]) * 60 + int(parts[1])
    return 0.0


def parse_youtube(text: str) -> list[ScriptSegment]:
    """Parse YouTube-style script text into segments.

    The parser processes line-by-line, looking for labeled cues (B-ROLL:,
    ON CAMERA:, etc.) and section markers ([INTRO], [OUTRO]).  Unlabeled
    text between cues is treated as on-camera dialogue.

    Args:
        text: Raw YouTube script text.

    Returns:
        Ordered list of ``ScriptSegment`` objects.
    """
    lines = text.splitlines()
    segments: list[ScriptSegment] = []
    idx = 0

    current_section: str = ""
    dialogue_buffer: list[str] = []

    def _flush_dialogue() -> None:
        nonlocal idx, dialogue_buffer
        joined = " ".join(dialogue_buffer).strip()
        if joined:
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.DIALOGUE,
                content=joined,
                scene_description=current_section,
            ))
            idx += 1
        dialogue_buffer = []

    for line in lines:
        stripped = line.strip()
        if not stripped:
            continue

        # Strip leading timestamp if present
        stripped = _TIMESTAMP_PREFIX_RE.sub("", stripped).strip()
        if not stripped:
            continue

        # Skip pure timestamp ranges (they're structural, not content)
        if _TIMESTAMP_RANGE_RE.match(stripped):
            continue

        # --- Section marker ---
        m_section = _SECTION_RE.match(stripped)
        if m_section:
            _flush_dialogue()
            section_name = m_section.group(1).strip()
            section_detail = (m_section.group(2) or "").strip()
            current_section = section_detail if section_detail else section_name

            # Some sections like [CTA] produce a title segment
            seg_type = _SECTION_TYPES.get(section_name.lower(), SegmentType.DIALOGUE)
            if seg_type == SegmentType.TITLE and section_detail:
                segments.append(ScriptSegment(
                    index=idx,
                    type=SegmentType.TITLE,
                    content=section_detail,
                    scene_description=section_name,
                ))
                idx += 1
            continue

        # --- B-ROLL ---
        m = _BROLL_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.BROLL,
                content=m.group(1).strip(),
                visual_direction=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- ON CAMERA ---
        m = _ON_CAMERA_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.DIALOGUE,
                content=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- TEXT ON SCREEN ---
        m = _TEXT_SCREEN_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.TITLE,
                content=m.group(1).strip(),
                visual_direction=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- LOWER THIRD ---
        m = _LOWER_THIRD_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.LOWER_THIRD,
                content=m.group(1).strip(),
                visual_direction=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- MUSIC ---
        m = _MUSIC_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.MUSIC,
                content=m.group(1).strip(),
                audio_direction=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- SFX ---
        m = _SFX_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.SFX,
                content=m.group(1).strip(),
                audio_direction=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- TRANSITION ---
        m = _TRANSITION_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.TRANSITION,
                content=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- VO ---
        m = _VO_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.VOICEOVER,
                content=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- CTA ---
        m = _CTA_RE.match(stripped)
        if m:
            _flush_dialogue()
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.TITLE,
                content=m.group(1).strip(),
                scene_description=current_section,
            ))
            idx += 1
            continue

        # --- Default: treat as on-camera dialogue ---
        dialogue_buffer.append(stripped)

    _flush_dialogue()
    return segments
