"""Screenplay format parser.

Parses standard screenplay / film script formatting:
- FADE IN / FADE OUT transitions
- Scene headings: INT. or EXT. (with location and time of day)
- Character names in ALL CAPS (centered or indented)
- Dialogue (indented text under a character name)
- Parentheticals: (beat), (whispering), etc.
- Action / description lines
- Transitions: CUT TO:, DISSOLVE TO:, SMASH CUT TO:, etc.
"""

from __future__ import annotations

import re

from src.models import ScriptSegment, SegmentType

# ── Regex patterns ───────────────────────────────────────────────────────────

# Scene headings: INT. COFFEE SHOP - DAY  or  EXT. PARK - NIGHT
_SCENE_HEADING_RE = re.compile(
    r"^\s*(INT\.|EXT\.|INT/EXT\.|I/E\.)\s+(.+)",
    re.IGNORECASE,
)

# Transitions: CUT TO:, FADE TO BLACK., DISSOLVE TO:, SMASH CUT TO:, etc.
_TRANSITION_RE = re.compile(
    r"^\s*(FADE\s+IN[:.]*|FADE\s+OUT[:.]*|FADE\s+TO\s+\w+[:.]*"
    r"|CUT\s+TO[:.]*|SMASH\s+CUT\s+TO[:.]*|MATCH\s+CUT\s+TO[:.]*"
    r"|DISSOLVE\s+TO[:.]*|WIPE\s+TO[:.]*|IRIS\s+(?:IN|OUT)[:.]*"
    r"|INTERCUT[:.]*)\s*$",
    re.IGNORECASE,
)

# Character name: all caps, possibly with (V.O.) or (O.S.) or (CONT'D)
_CHARACTER_RE = re.compile(
    r"^\s{10,}([A-Z][A-Z\s.''-]{1,40}?)(?:\s*\((?:V\.?O\.?|O\.?S\.?|CONT'?D?|O\.?C\.?)\))?\s*$"
)

# Simplified character name: all caps on its own line (fallback when indentation varies)
_CHARACTER_SIMPLE_RE = re.compile(
    r"^([A-Z][A-Z\s.''-]{1,40}?)(?:\s*\((?:V\.?O\.?|O\.?S\.?|CONT'?D?|O\.?C\.?)\))?\s*$"
)

# Parenthetical: (beat), (sotto voce), etc.
_PARENTHETICAL_RE = re.compile(r"^\s*\(([^)]+)\)\s*$")

# V.O. or O.S. annotation on a character name
_VO_ANNOTATION_RE = re.compile(r"\(V\.?O\.?\)", re.IGNORECASE)
_OS_ANNOTATION_RE = re.compile(r"\(O\.?S\.?\)", re.IGNORECASE)


def _is_likely_character_name(line: str) -> bool:
    """Heuristic: line is a character name if it's mostly uppercase and short."""
    stripped = line.strip()
    # Remove parenthetical annotation for the check
    name_part = re.sub(r"\s*\([^)]*\)\s*", "", stripped)
    if not name_part:
        return False
    # Must be mostly uppercase letters
    alpha_chars = [c for c in name_part if c.isalpha()]
    if len(alpha_chars) < 2:
        return False
    upper_ratio = sum(1 for c in alpha_chars if c.isupper()) / len(alpha_chars)
    return upper_ratio > 0.8 and len(name_part) < 40


def _extract_scene_location(heading_text: str) -> str:
    """Extract the location from a scene heading, stripping the time of day."""
    # Remove time of day: ' - DAY', ' - NIGHT', etc.
    cleaned = re.sub(r"\s*-\s*(DAY|NIGHT|MORNING|EVENING|DAWN|DUSK|LATER|CONTINUOUS)\s*$",
                     "", heading_text, flags=re.IGNORECASE)
    return cleaned.strip()


def parse_screenplay(text: str) -> list[ScriptSegment]:
    """Parse screenplay-formatted text into a list of ``ScriptSegment`` objects.

    This parser processes the script line-by-line, using a simple state machine
    to track whether we are inside a dialogue block, action, etc.

    Args:
        text: Raw screenplay text.

    Returns:
        List of parsed segments in script order.
    """
    lines = text.splitlines()
    segments: list[ScriptSegment] = []
    idx = 0

    current_speaker: str = ""
    current_scene: str = ""
    action_buffer: list[str] = []
    dialogue_buffer: list[str] = []
    is_voiceover: bool = False

    def _flush_action() -> None:
        nonlocal idx, action_buffer
        joined = " ".join(action_buffer).strip()
        if joined:
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.ACTION,
                content=joined,
                scene_description=current_scene,
            ))
            idx += 1
        action_buffer = []

    def _flush_dialogue() -> None:
        nonlocal idx, dialogue_buffer, current_speaker, is_voiceover
        joined = " ".join(dialogue_buffer).strip()
        if joined:
            seg_type = SegmentType.VOICEOVER if is_voiceover else SegmentType.DIALOGUE
            segments.append(ScriptSegment(
                index=idx,
                type=seg_type,
                content=joined,
                speaker=current_speaker,
                scene_description=current_scene,
            ))
            idx += 1
        dialogue_buffer = []
        is_voiceover = False

    in_dialogue = False

    for line in lines:
        stripped = line.strip()

        # Skip blank lines (they separate blocks)
        if not stripped:
            if in_dialogue and dialogue_buffer:
                _flush_dialogue()
                in_dialogue = False
            continue

        # --- Transition ---
        if _TRANSITION_RE.match(line):
            _flush_action()
            if in_dialogue:
                _flush_dialogue()
                in_dialogue = False
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.TRANSITION,
                content=stripped,
            ))
            idx += 1
            continue

        # --- Scene heading ---
        m_scene = _SCENE_HEADING_RE.match(line)
        if m_scene:
            _flush_action()
            if in_dialogue:
                _flush_dialogue()
                in_dialogue = False
            current_scene = _extract_scene_location(m_scene.group(2))
            segments.append(ScriptSegment(
                index=idx,
                type=SegmentType.ACTION,
                content=stripped,
                scene_description=current_scene,
            ))
            idx += 1
            continue

        # --- Character name ---
        if not in_dialogue and (_CHARACTER_RE.match(line) or
                                (_CHARACTER_SIMPLE_RE.match(stripped)
                                 and _is_likely_character_name(stripped))):
            _flush_action()
            current_speaker = re.sub(r"\s*\([^)]*\)\s*", "", stripped).strip()
            is_voiceover = bool(_VO_ANNOTATION_RE.search(stripped))
            in_dialogue = True
            dialogue_buffer = []
            continue

        # --- Parenthetical inside dialogue ---
        m_paren = _PARENTHETICAL_RE.match(stripped)
        if in_dialogue and m_paren:
            # Parentheticals are stage directions within dialogue; append to dialogue.
            dialogue_buffer.append(f"({m_paren.group(1)})")
            continue

        # --- Dialogue text ---
        if in_dialogue:
            dialogue_buffer.append(stripped)
            continue

        # --- Action / description ---
        action_buffer.append(stripped)

    # Flush remaining buffers
    if in_dialogue:
        _flush_dialogue()
    _flush_action()

    return segments
