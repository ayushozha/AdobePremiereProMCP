"""Script parsing module.

Parses video scripts (screenplay, YouTube, podcast, narration formats)
into structured ``ScriptSegment`` objects that drive automated video editing.

Quick start::

    from src.parser import ScriptParser

    parser = ScriptParser()
    result = parser.parse(script_text)
    # result.segments — list of ScriptSegment
    # result.metadata — ScriptMetadata with title, format, duration
"""

from __future__ import annotations

from src.parser.asset_extractor import extract_hints
from src.parser.duration_estimator import estimate_duration
from src.parser.script_parser import ScriptParser

__all__ = [
    "ScriptParser",
    "estimate_duration",
    "extract_hints",
]
