"""Keyword-based asset matching.

Compares a segment's ``asset_hints`` (and descriptive text) against asset file
names, paths, and metadata using token overlap (Jaccard similarity) with
an exact-match boost.
"""

from __future__ import annotations

from src.models import AssetInfo, ScriptSegment

from .scoring import ScoredMatch, normalize_text

# Bonus applied when a segment hint token exactly matches an asset token.
_EXACT_MATCH_BOOST = 0.15


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
            hint_tokens = set(
                tok
                for hint in segment.asset_hints
                for tok in normalize_text(hint)
            )
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
