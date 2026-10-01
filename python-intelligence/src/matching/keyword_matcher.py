"""Keyword-based asset matching.

Compares a segment's ``asset_hints`` (and descriptive text) against asset file
names, paths, and metadata using token overlap (Jaccard similarity) with
an exact-match boost.
"""

from __future__ import annotations

import posixpath
import re
from typing import TYPE_CHECKING

from .scoring import ScoredMatch, normalize_text

if TYPE_CHECKING:
    from src.models import AssetInfo, ScriptSegment

# Bonus applied when a segment hint token exactly matches an asset token.
_EXACT_MATCH_BOOST = 0.15

# Explicit media references are identifiers, not a bag of directory words.
# Keep this in step with the scanner's supported video, audio and image formats.
_MEDIA_EXTENSIONS = (
    r"mp4|mov|avi|mkv|wmv|flv|webm|m4v|mpg|mpeg|3gp|ts|mts|m2ts|vob|ogv|mxf|prores|"
    r"mp3|wav|aac|flac|ogg|wma|m4a|aiff|aif|opus|ac3|dts|pcm|"
    r"jpg|jpeg|png|bmp|tiff|tif|gif|webp|exr|dpx|tga|psd|svg|ico|heic|heif|avif"
)
_MEDIA_REFERENCE_SUFFIX = re.compile(r"\.(?:" + _MEDIA_EXTENSIONS + r")$", re.IGNORECASE)
_QUOTED_TEXT = re.compile(r"(?<!\w)([\"'])(.*?)\1(?!\w)")
_PATH_REFERENCE = re.compile(
    r"(?<!\S)((?:/|[a-z]:[/\\]|\\\\|[^\s\"']+[/\\])[^\r\n\"']*?\."
    r"(?:" + _MEDIA_EXTENSIONS + r"))(?=$|[\s.,;!?])", re.IGNORECASE,
)
_TOKEN_REFERENCE = re.compile(
    r"(?<!\S)([^\s\"'<>]+?\.(?:" + _MEDIA_EXTENSIONS + r"))(?=$|[\s.,;!?])",
    re.IGNORECASE,
)


def _media_references(segment: ScriptSegment) -> set[str] | None:
    """Return literal references, or None for an ambiguous unquoted relative path."""
    references: set[str] = set()

    def add(value: str) -> None:
        if _MEDIA_REFERENCE_SUFFIX.search(value):
            references.add(posixpath.normpath(value.replace("\\", "/")))

    for text in (segment.content, segment.visual_direction, segment.audio_direction,
                 segment.scene_description):
        for quoted in _QUOTED_TEXT.finditer(text):
            add(quoted.group(2))
        unquoted = _QUOTED_TEXT.sub(" ", text)
        for match in _PATH_REFERENCE.finditer(unquoted):
            candidate = match.group(1).replace("\\", "/")
            absolute = candidate.startswith("/") or bool(re.match(
                r"^[a-z]:/", candidate, re.IGNORECASE,
            ))
            if not absolute and unquoted[:match.start()].strip():
                # "wanted folder/clip.mp4" could name a folder with spaces;
                # "show wanted folder/clip.mp4" adds an ambiguous prose prefix.
                # Require quotes instead of binding only "folder/clip.mp4".
                return None
            add(match.group(1))
        # Preserve paths with spaces as a whole; their trailing filename must
        # not become a second, less specific reference to another directory.
        unquoted = _PATH_REFERENCE.sub(" ", unquoted)
        for token in _TOKEN_REFERENCE.finditer(unquoted):
            add(token.group(1))
    # The parser lowercases derived hints. Raw references above are authoritative
    # for case-sensitive POSIX identities; hints only fill otherwise empty input.
    if not references:
        for hint in segment.asset_hints:
            add(hint)
    return references


class KeywordMatcher:
    """Match script segments to assets by keyword / token overlap."""

    # ── public API ───────────────────────────────────────────────────────────

    def match(
        self,
        segment: ScriptSegment,
        assets: list[AssetInfo],
    ) -> list[ScoredMatch]:
        """Return a list of ``ScoredMatch`` for *segment* against *assets*.

        Results are sorted by descending score.
        """
        explicit = self.match_explicit_references(segment, assets)
        if explicit is not None:
            return explicit
        segment_tokens = self._segment_tokens(segment)
        if not segment_tokens:
            return []

        scored: list[ScoredMatch] = []
        for asset in assets:
            asset_tokens = self._asset_tokens(asset)
            if not asset_tokens:
                continue

            score = self._jaccard(segment_tokens, asset_tokens)

            # Boost for exact hint-to-filename token matches.
            hint_tokens = set(tok for hint in segment.asset_hints for tok in normalize_text(hint))
            exact_hits = hint_tokens & asset_tokens
            if exact_hits:
                score = min(1.0, score + _EXACT_MATCH_BOOST * len(exact_hits))

            if score > 0.0:
                reasoning = self._build_reasoning(segment_tokens, asset_tokens, exact_hits)
                scored.append(
                    ScoredMatch(
                        asset_id=asset.id,
                        score=round(score, 4),
                        reasoning=reasoning,
                        method="keyword",
                    )
                )

        scored.sort(key=lambda m: m.score, reverse=True)
        return scored

    # ── private helpers ──────────────────────────────────────────────────────

    @staticmethod
    def match_explicit_references(
        segment: ScriptSegment,
        assets: list[AssetInfo],
    ) -> list[ScoredMatch] | None:
        """Resolve literal filenames/paths before applying description heuristics.

        ``None`` means no explicit reference, so normal matching still applies.
        An empty list means a referenced file is absent; similar filenames or
        shared parent directories must not substitute unrelated media.
        """
        references = _media_references(segment)
        if references is None:
            return []
        if not references:
            return None
        matched_assets: dict[str, ScoredMatch] = {}
        for reference in sorted(references):
            candidates: list[AssetInfo] = []
            for asset in assets:
                asset_path = posixpath.normpath(asset.file_path.replace("\\", "/"))
                filename = asset.file_name or posixpath.basename(asset_path)
                windows_path = bool(re.match(r"^[a-z]:/", asset_path, re.IGNORECASE)) or (
                    asset_path.startswith("//")
                )
                compared_ref = reference.casefold() if windows_path else reference
                compared_path = asset_path.casefold() if windows_path else asset_path
                compared_name = filename.casefold() if windows_path else filename
                if "/" not in compared_ref:
                    matched = compared_ref == compared_name
                elif compared_ref.startswith("/") or re.match(
                    r"^[a-z]:/", compared_ref, re.IGNORECASE,
                ):
                    matched = compared_ref == compared_path
                else:
                    matched = compared_path == compared_ref or compared_path.endswith(
                        "/" + compared_ref,
                    )
                if matched:
                    candidates.append(asset)
            if len(candidates) > 1:
                return []  # A basename or suffix is ambiguous; never bind by scan order.
            if not candidates:
                return []  # Do not silently omit one of the explicitly requested files.
            if candidates:
                asset = candidates[0]
                matched_assets[asset.id] = ScoredMatch(
                    asset_id=asset.id, score=1.0,
                    reasoning=f"Explicit media reference: {reference}", method="keyword",
                )
        # EDL generation selects one asset per segment. Multiple explicit assets
        # provide no deterministic choice, so require separate script segments.
        return list(matched_assets.values()) if len(matched_assets) <= 1 else []

    @staticmethod
    def _segment_tokens(segment: ScriptSegment) -> set[str]:
        """Collect normalised tokens from all descriptive fields."""
        parts: list[str] = list(segment.asset_hints)
        if segment.visual_direction:
            parts.append(segment.visual_direction)
        if segment.scene_description:
            parts.append(segment.scene_description)
        if segment.content:
            parts.append(segment.content)
        return set(tok for part in parts for tok in normalize_text(part))

    @staticmethod
    def _asset_tokens(asset: AssetInfo) -> set[str]:
        """Collect normalised tokens from the asset's identifying fields."""
        parts: list[str] = [asset.file_name, asset.file_path]
        parts.extend(asset.metadata.values())
        return set(tok for part in parts for tok in normalize_text(part))

    @staticmethod
    def _jaccard(a: set[str], b: set[str]) -> float:
        """Jaccard similarity coefficient."""
        if not a or not b:
            return 0.0
        intersection = a & b
        union = a | b
        return len(intersection) / len(union)

    @staticmethod
    def _build_reasoning(
        segment_tokens: set[str],
        asset_tokens: set[str],
        exact_hits: set[str],
    ) -> str:
        overlap = segment_tokens & asset_tokens
        parts: list[str] = []
        if overlap:
            parts.append(f"Overlapping keywords: {', '.join(sorted(overlap))}")
        if exact_hits:
            parts.append(f"Exact hint matches: {', '.join(sorted(exact_hits))}")
        return "; ".join(parts) if parts else "Low keyword overlap"
