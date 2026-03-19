"""Main asset-matching orchestrator.

``AssetMatcher`` coordinates keyword matching, embedding matching, and hybrid
scoring to pair script segments with the best-fitting media assets.
"""

from __future__ import annotations

import logging

from src.models import (
    AssetInfo,
    AssetMatch,
    MatchResult,
    MatchStrategy,
    ScriptSegment,
    UnmatchedSegment,
)

from .embedding_matcher import EmbeddingMatcher
from .keyword_matcher import KeywordMatcher
from .scoring import ScoredMatch, combine_scores
from .suggest import suggest_assets

log = logging.getLogger(__name__)

# Default thresholds – can be overridden via ``IntelligenceSettings``.
_DEFAULT_CONFIDENCE_THRESHOLD = 0.6
_DEFAULT_MAX_MATCHES = 3
_DEFAULT_KEYWORD_WEIGHT = 0.3


class AssetMatcher:
    """Orchestrate segment-to-asset matching using the chosen strategy.

    Parameters
    ----------
    strategy:
        Which matching approach to use (KEYWORD, EMBEDDING, or HYBRID).
    confidence_threshold:
        Minimum score to accept a match.
    max_matches_per_segment:
        Maximum number of candidate matches returned per segment.
    keyword_weight:
        Weight given to the keyword score when using the HYBRID strategy.
    embedding_model:
        OpenAI model name forwarded to ``EmbeddingMatcher``.
    openai_api_key:
        Optional explicit API key forwarded to ``EmbeddingMatcher``.
    """

    def __init__(
        self,
        strategy: MatchStrategy = MatchStrategy.HYBRID,
        confidence_threshold: float = _DEFAULT_CONFIDENCE_THRESHOLD,
        max_matches_per_segment: int = _DEFAULT_MAX_MATCHES,
        keyword_weight: float = _DEFAULT_KEYWORD_WEIGHT,
        embedding_model: str = "text-embedding-3-small",
        openai_api_key: str | None = None,
    ) -> None:
        self.strategy = strategy
        self.confidence_threshold = confidence_threshold
        self.max_matches_per_segment = max_matches_per_segment
        self.keyword_weight = keyword_weight

        self.keyword_matcher = KeywordMatcher()
        self.embedding_matcher = EmbeddingMatcher(
            model_name=embedding_model,
            api_key=openai_api_key,
        )

    # ── public API ───────────────────────────────────────────────────────────

    def match(
        self,
        segments: list[ScriptSegment],
        assets: list[AssetInfo],
    ) -> MatchResult:
        """Match every segment in *segments* to the best assets in *assets*.

        Returns a ``MatchResult`` containing accepted matches **and** a list
        of unmatched segments with human-readable suggestions.
        """
        if not assets:
            log.warning("No assets provided; all segments will be unmatched")

        all_matches: list[AssetMatch] = []
        unmatched: list[UnmatchedSegment] = []
        available_types = list({a.asset_type.name for a in assets})

        for segment in segments:
            scored = self._score_segment(segment, assets)

            # Filter by threshold and cap the count.
            accepted = [
                s for s in scored if s.score >= self.confidence_threshold
            ][: self.max_matches_per_segment]

            if accepted:
                for sm in accepted:
                    all_matches.append(
                        AssetMatch(
                            segment_index=segment.index,
                            asset_id=sm.asset_id,
                            confidence=sm.score,
                            reasoning=sm.reasoning,
                        )
                    )
            else:
                # Build suggestions for the editor.
                suggestions = suggest_assets(segment, available_types)
                best_score = scored[0].score if scored else 0.0
                reason = (
                    f"Best candidate score ({best_score:.2f}) "
                    f"below threshold ({self.confidence_threshold:.2f})"
                    if scored
                    else "No candidate assets found"
                )
                unmatched.append(
                    UnmatchedSegment(
                        segment_index=segment.index,
                        reason=reason,
                        suggestions=suggestions,
                    )
                )

        log.info(
            "Matching complete: %d matches, %d unmatched segments",
            len(all_matches),
            len(unmatched),
        )
        return MatchResult(matches=all_matches, unmatched=unmatched)

    # ── private helpers ──────────────────────────────────────────────────────

    def _score_segment(
        self,
        segment: ScriptSegment,
        assets: list[AssetInfo],
    ) -> list[ScoredMatch]:
        """Return all ``ScoredMatch`` entries for a single segment."""
        match self.strategy:
            case MatchStrategy.KEYWORD:
                return self.keyword_matcher.match(segment, assets)
            case MatchStrategy.EMBEDDING:
                return self.embedding_matcher.match(segment, assets)
            case MatchStrategy.HYBRID | _:
                return self._hybrid_match(segment, assets)

    def _hybrid_match(
        self,
        segment: ScriptSegment,
        assets: list[AssetInfo],
    ) -> list[ScoredMatch]:
        """Combine keyword and embedding scores into a single ranked list."""
        kw_results = self.keyword_matcher.match(segment, assets)
        emb_results = self.embedding_matcher.match(segment, assets)

        kw_map: dict[str, ScoredMatch] = {r.asset_id: r for r in kw_results}
        emb_map: dict[str, ScoredMatch] = {r.asset_id: r for r in emb_results}

        all_asset_ids = set(kw_map) | set(emb_map)
        combined: list[ScoredMatch] = []

        for asset_id in all_asset_ids:
            kw_score = kw_map[asset_id].score if asset_id in kw_map else 0.0
            emb_score = emb_map[asset_id].score if asset_id in emb_map else 0.0
            hybrid_score = combine_scores(kw_score, emb_score, self.keyword_weight)

            # Merge reasoning from both methods.
            reasoning_parts: list[str] = []
            if asset_id in kw_map:
                reasoning_parts.append(f"[keyword] {kw_map[asset_id].reasoning}")
            if asset_id in emb_map:
                reasoning_parts.append(f"[embedding] {emb_map[asset_id].reasoning}")
            reasoning_parts.append(
                f"[hybrid] kw={kw_score:.3f} * {self.keyword_weight} "
                f"+ emb={emb_score:.3f} * {1.0 - self.keyword_weight:.1f} "
                f"= {hybrid_score:.4f}"
            )

            combined.append(
                ScoredMatch(
                    asset_id=asset_id,
                    score=round(hybrid_score, 4),
                    reasoning=" | ".join(reasoning_parts),
                    method="hybrid",
                )
            )

        combined.sort(key=lambda m: m.score, reverse=True)
        return combined
