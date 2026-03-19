"""Suggestion engine for unmatched script segments.

When a segment cannot be matched to any available asset this module generates
human-readable suggestions describing what kind of media the editor should
provide.
"""

from __future__ import annotations

from src.models import AssetType, ScriptSegment, SegmentType

# ── Segment-type to asset-type mapping ───────────────────────────────────────

_SEGMENT_ASSET_MAP: dict[SegmentType, AssetType] = {
    SegmentType.DIALOGUE: AssetType.VIDEO,
    SegmentType.ACTION: AssetType.VIDEO,
    SegmentType.BROLL: AssetType.VIDEO,
    SegmentType.TRANSITION: AssetType.VIDEO,
    SegmentType.TITLE: AssetType.GRAPHICS,
    SegmentType.LOWER_THIRD: AssetType.GRAPHICS,
    SegmentType.VOICEOVER: AssetType.AUDIO,
    SegmentType.MUSIC: AssetType.AUDIO,
    SegmentType.SFX: AssetType.AUDIO,
}

_ASSET_TYPE_LABELS: dict[AssetType, str] = {
    AssetType.VIDEO: "video clip",
    AssetType.AUDIO: "audio track",
    AssetType.IMAGE: "image",
    AssetType.GRAPHICS: "graphic / title card",
}


def suggest_assets(
    segment: ScriptSegment,
    available_types: list[str] | None = None,
) -> list[str]:
    """Suggest what kind of assets the user should provide for *segment*.

    Parameters
    ----------
    segment:
        The unmatched ``ScriptSegment``.
    available_types:
        Optional list of asset-type labels already present in the project
        (e.g. ``["video", "audio"]``).  Used to tailor the phrasing.

    Returns
    -------
    list[str]
        One or more human-readable suggestions.
    """
    suggestions: list[str] = []
    available_lower = {t.lower() for t in available_types} if available_types else set()

    # Determine the expected asset type from the segment type.
    expected_type = _SEGMENT_ASSET_MAP.get(segment.type, AssetType.VIDEO)
    type_label = _ASSET_TYPE_LABELS.get(expected_type, "media file")

    # Build a description from segment content.
    description = _describe_content(segment)

    # Primary suggestion based on segment type.
    suggestions.append(_primary_suggestion(segment.type, type_label, description))

    # If the expected asset type is not among the available ones, note that.
    if available_types and expected_type.name.lower() not in available_lower:
        suggestions.append(
            f"No {type_label}s found in the project. "
            f"Import {type_label} assets to match this segment."
        )

    # Specific hints per segment type.
    suggestions.extend(_type_specific_hints(segment))

    return suggestions


# ── Private helpers ──────────────────────────────────────────────────────────


def _describe_content(segment: ScriptSegment) -> str:
    """Build a short human-readable description of the segment's intent."""
    parts: list[str] = []
    if segment.visual_direction:
        parts.append(segment.visual_direction)
    elif segment.scene_description:
        parts.append(segment.scene_description)
    elif segment.content:
        # Truncate long dialogue/content to keep suggestions readable.
        content = segment.content
        if len(content) > 80:
            content = content[:77] + "..."
        parts.append(content)
    if segment.asset_hints:
        parts.append(f"(hints: {', '.join(segment.asset_hints)})")
    return " ".join(parts) if parts else "unspecified content"


def _primary_suggestion(
    seg_type: SegmentType,
    type_label: str,
    description: str,
) -> str:
    """Return the main suggestion sentence."""
    match seg_type:
        case SegmentType.BROLL:
            return f"Need B-roll footage of {description}"
        case SegmentType.MUSIC:
            return f"Need background music track for {description}"
        case SegmentType.SFX:
            return f"Need sound effect for {description}"
        case SegmentType.VOICEOVER:
            return f"Need voiceover recording for {description}"
        case SegmentType.TITLE:
            return f"Need title card graphic for {description}"
        case SegmentType.LOWER_THIRD:
            return f"Need lower-third graphic for {description}"
        case SegmentType.DIALOGUE:
            return f"Need {type_label} with dialogue: {description}"
        case SegmentType.ACTION:
            return f"Need {type_label} showing: {description}"
        case SegmentType.TRANSITION:
            return f"Need transition {type_label} for {description}"
        case _:
            return f"Need {type_label} for {description}"


def _type_specific_hints(segment: ScriptSegment) -> list[str]:
    """Return extra tips depending on the segment type and fields."""
    hints: list[str] = []

    if segment.type == SegmentType.BROLL and segment.visual_direction:
        hints.append(
            f"Consider stock footage matching: {segment.visual_direction}"
        )

    if segment.type in (SegmentType.MUSIC, SegmentType.SFX) and segment.audio_direction:
        hints.append(
            f"Audio direction specifies: {segment.audio_direction}"
        )

    if segment.estimated_duration_seconds > 0:
        hints.append(
            f"Required duration: ~{segment.estimated_duration_seconds:.1f}s"
        )

    return hints
