"""Main script parser — entry point for all script-to-segments parsing.

``ScriptParser`` auto-detects (or accepts) the script format, delegates to
the appropriate format-specific parser, then enriches every segment with
estimated durations and asset-matching keywords.
"""

from __future__ import annotations

import logging
from pathlib import Path
from typing import Callable

from src.models import (
    ParsedScript,
    ScriptFormat,
    ScriptMetadata,
    ScriptSegment,
    SegmentType,
)
from src.parser.asset_extractor import extract_hints
from src.parser.duration_estimator import estimate_duration
from src.parser.formats.narration import parse_narration
from src.parser.formats.screenplay import parse_screenplay
from src.parser.formats.youtube import parse_youtube

logger = logging.getLogger(__name__)

# ── Format detection heuristics ──────────────────────────────────────────────

# Patterns that strongly suggest a screenplay.
_SCREENPLAY_INDICATORS: list[str] = [
    "FADE IN",
    "FADE OUT",
    "INT.",
    "EXT.",
    "INT/EXT.",
    "CUT TO:",
    "DISSOLVE TO:",
    "SMASH CUT",
]

# Patterns that strongly suggest a YouTube script.
_YOUTUBE_INDICATORS: list[str] = [
    "[INTRO]",
    "[HOOK]",
    "[MAIN]",
    "[OUTRO]",
    "B-ROLL:",
    "ON CAMERA:",
    "TEXT ON SCREEN:",
    "BROLL:",
    "B ROLL:",
]

# Patterns that suggest narration / voiceover format.
_NARRATION_INDICATORS: list[str] = [
    "[Visual:",
    "[Title:",
    "[Music:",
    "[SFX:",
    "[Sound:",
    "[Transition:",
]

# Map from ScriptFormat to parser function.
_FORMAT_PARSERS: dict[ScriptFormat, Callable[[str], list[ScriptSegment]]] = {
    ScriptFormat.SCREENPLAY: parse_screenplay,
    ScriptFormat.YOUTUBE: parse_youtube,
    ScriptFormat.NARRATION: parse_narration,
}


def _detect_format(text: str) -> ScriptFormat:
    """Auto-detect the script format from its content.

    Scans the first ~2000 characters (and a few deeper lines) for format-
    specific indicators. Returns the best guess, defaulting to narration
    if nothing matches strongly.
    """
    # Use a generous sample from the beginning.
    sample = text[:3000].upper()

    scores: dict[ScriptFormat, int] = {
        ScriptFormat.SCREENPLAY: 0,
        ScriptFormat.YOUTUBE: 0,
        ScriptFormat.NARRATION: 0,
    }

    for indicator in _SCREENPLAY_INDICATORS:
        if indicator in sample:
            scores[ScriptFormat.SCREENPLAY] += 1

    for indicator in _YOUTUBE_INDICATORS:
        if indicator in sample:
            scores[ScriptFormat.YOUTUBE] += 1

    # Narration indicators are case-sensitive (brackets matter).
    sample_original = text[:3000]
    for indicator in _NARRATION_INDICATORS:
        if indicator in sample_original:
            scores[ScriptFormat.NARRATION] += 1

    # Pick the format with the highest score.
    best_format = max(scores, key=lambda f: scores[f])

    # If no indicators matched at all, default to narration (simplest format).
    if scores[best_format] == 0:
        logger.debug("No format indicators found; defaulting to narration")
        return ScriptFormat.NARRATION

    logger.debug("Detected format: %s (scores: %s)", best_format.value, scores)
    return best_format


def _extract_title(text: str, fmt: ScriptFormat) -> str:
    """Try to extract a title from the script text.

    Looks for common title patterns at the start of the document.
    """
    lines = text.strip().splitlines()
    if not lines:
        return ""

    # Many scripts start with a title on the first non-blank line.
    for line in lines[:5]:
        stripped = line.strip()
        if not stripped:
            continue
        # Skip known structural lines.
        upper = stripped.upper()
        if any(upper.startswith(ind) for ind in ("FADE", "INT.", "EXT.", "[")):
            break
        # A short, non-structural first line is likely a title.
        if len(stripped) < 80 and not stripped.startswith("("):
            return stripped
        break

    return ""


def _read_file_text(file_path: str) -> str:
    """Read text content from a file, supporting .txt, .pdf, and .docx.

    Args:
        file_path: Path to the script file.

    Returns:
        Extracted text content.

    Raises:
        FileNotFoundError: If the file does not exist.
        ValueError: If the file type is not supported.
    """
    path = Path(file_path)

    if not path.exists():
        raise FileNotFoundError(f"Script file not found: {file_path}")

    suffix = path.suffix.lower()

    if suffix in (".txt", ".fountain", ".fdx"):
        return path.read_text(encoding="utf-8")

    if suffix == ".pdf":
        try:
            from pypdf import PdfReader  # type: ignore[import-untyped]
        except ImportError as exc:
            raise ImportError(
                "pypdf is required for PDF parsing. Install it with: pip install pypdf"
            ) from exc

        reader = PdfReader(str(path))
        pages: list[str] = []
        for page in reader.pages:
            extracted = page.extract_text()
            if extracted:
                pages.append(extracted)
        return "\n\n".join(pages)

    if suffix in (".docx", ".doc"):
        try:
            import docx  # type: ignore[import-untyped]
        except ImportError as exc:
            raise ImportError(
                "python-docx is required for DOCX parsing. "
                "Install it with: pip install python-docx"
            ) from exc

        doc = docx.Document(str(path))
        return "\n".join(para.text for para in doc.paragraphs)

    raise ValueError(
        f"Unsupported file type '{suffix}'. Supported: .txt, .pdf, .docx"
    )


# ── Public API ───────────────────────────────────────────────────────────────


class ScriptParser:
    """Parse video scripts into structured segments for automated editing.

    Supports multiple script formats (screenplay, YouTube, narration) with
    automatic format detection.  Each parsed segment is enriched with an
    estimated duration and asset-matching keywords.

    Usage::

        parser = ScriptParser()
        result = parser.parse(script_text, format_hint="auto")
        for seg in result.segments:
            print(seg.type, seg.estimated_duration_seconds, seg.asset_hints)
    """

    def parse(self, text: str, format_hint: str = "auto") -> ParsedScript:
        """Parse script text into structured segments.

        Args:
            text: Raw script text to parse.
            format_hint: One of ``"auto"``, ``"screenplay"``, ``"youtube"``,
                ``"narration"``, ``"podcast"``.  When ``"auto"`` the format
                is detected from content patterns.

        Returns:
            A ``ParsedScript`` containing the ordered segment list and metadata.

        Raises:
            ValueError: If *text* is empty or *format_hint* is not recognized.
        """
        if not text or not text.strip():
            raise ValueError("Cannot parse empty script text")

        # Resolve format.
        fmt = self._resolve_format(text, format_hint)
        logger.info("Parsing script with format: %s", fmt.value)

        # Delegate to format-specific parser.
        parser_fn = _FORMAT_PARSERS.get(fmt)
        if parser_fn is None:
            # Fallback: podcast scripts are narration-like.
            parser_fn = parse_narration

        raw_segments = parser_fn(text)

        # Enrich each segment with duration and asset hints.
        for segment in raw_segments:
            segment.estimated_duration_seconds = round(
                estimate_duration(segment), 2
            )
            hint_source = (
                segment.visual_direction
                or segment.audio_direction
                or segment.content
            )
            if hint_source and segment.type in (
                SegmentType.BROLL,
                SegmentType.ACTION,
                SegmentType.MUSIC,
                SegmentType.SFX,
                SegmentType.TITLE,
                SegmentType.LOWER_THIRD,
            ):
                segment.asset_hints = extract_hints(hint_source)

        # Build metadata.
        total_duration = sum(s.estimated_duration_seconds for s in raw_segments)
        title = _extract_title(text, fmt)

        metadata = ScriptMetadata(
            title=title,
            format=fmt,
            estimated_total_duration_seconds=round(total_duration, 2),
            segment_count=len(raw_segments),
        )

        return ParsedScript(segments=raw_segments, metadata=metadata)

    def parse_file(self, file_path: str, format_hint: str = "auto") -> ParsedScript:
        """Parse a script from a file.

        Supports ``.txt``, ``.pdf``, and ``.docx`` files.

        Args:
            file_path: Path to the script file.
            format_hint: Format hint (see :meth:`parse`).

        Returns:
            A ``ParsedScript`` with segments and metadata.

        Raises:
            FileNotFoundError: If the file does not exist.
            ValueError: If the file type is unsupported.
        """
        text = _read_file_text(file_path)
        result = self.parse(text, format_hint=format_hint)

        # If no title was extracted from content, use the filename.
        if not result.metadata.title:
            result.metadata.title = Path(file_path).stem.replace("_", " ").replace("-", " ").title()

        return result

    # ── Private helpers ──────────────────────────────────────────────────────

    @staticmethod
    def _resolve_format(text: str, format_hint: str) -> ScriptFormat:
        """Convert *format_hint* string to ``ScriptFormat``, auto-detecting if needed."""
        hint_lower = format_hint.strip().lower()

        if hint_lower == "auto":
            return _detect_format(text)

        # Map known strings to enum values.
        try:
            return ScriptFormat(hint_lower)
        except ValueError:
            logger.warning(
                "Unknown format_hint '%s', falling back to auto-detection",
                format_hint,
            )
            return _detect_format(text)
