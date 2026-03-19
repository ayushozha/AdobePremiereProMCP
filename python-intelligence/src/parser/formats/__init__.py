"""Format-specific script parsers.

Each sub-module exposes a ``parse_<format>(text) -> list[ScriptSegment]``
function that the main ``ScriptParser`` delegates to.
"""

from __future__ import annotations

from src.parser.formats.narration import parse_narration
from src.parser.formats.screenplay import parse_screenplay
from src.parser.formats.youtube import parse_youtube

__all__ = [
    "parse_narration",
    "parse_screenplay",
    "parse_youtube",
]
