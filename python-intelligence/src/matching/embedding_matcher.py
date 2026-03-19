"""Embedding-based asset matching.

Uses OpenAI's text-embedding API to compute semantic similarity between
segment descriptions and asset descriptions.  Falls back to keyword matching
when the ``openai`` package is unavailable or no API key is configured.
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from src.models import AssetInfo, ScriptSegment

from .scoring import ScoredMatch, cosine_similarity, normalize_text

if TYPE_CHECKING:
    import openai

log = logging.getLogger(__name__)

# ── Embedding cache ──────────────────────────────────────────────────────────
# Keyed by the raw text that was embedded.  Persists for the lifetime of the
# ``EmbeddingMatcher`` instance so repeated calls within one matching run
# never duplicate API requests.

type _EmbeddingCache = dict[str, list[float]]


class EmbeddingMatcher:
    """Match segments to assets using text-embedding cosine similarity.

    Parameters
    ----------
    model_name:
        The OpenAI embedding model to use.
    api_key:
        Optional explicit API key.  When *None* the ``openai`` client falls
        back to the ``OPENAI_API_KEY`` environment variable.
    """

    def __init__(
        self,
        model_name: str = "text-embedding-3-small",
        api_key: str | None = None,
    ) -> None:
        self.model_name = model_name
        self._api_key = api_key
        self._client: openai.OpenAI | None = None
        self._cache: _EmbeddingCache = {}
        self._available: bool | None = None  # lazily determined

    # ── public API ───────────────────────────────────────────────────────────

    def match(
        self,
        segment: ScriptSegment,
        assets: list[AssetInfo],
    ) -> list[ScoredMatch]:
        """Return scored matches for *segment* against *assets*.

        If the OpenAI client cannot be initialised the method falls back to a
        simple token-overlap heuristic so callers always receive results.
        """
        if not self._is_available():
            log.info("Embedding API unavailable; falling back to keyword heuristic")
            return self._fallback_match(segment, assets)

        segment_text = self._segment_text(segment)
        segment_vec = self._embed(segment_text)
        if segment_vec is None:
            return self._fallback_match(segment, assets)

        scored: list[ScoredMatch] = []
        for asset in assets:
            asset_text = self._asset_text(asset)
            asset_vec = self._embed(asset_text)
            if asset_vec is None:
                continue
            similarity = cosine_similarity(segment_vec, asset_vec)
            if similarity > 0.0:
                scored.append(
                    ScoredMatch(
                        asset_id=asset.id,
                        score=round(max(0.0, min(1.0, similarity)), 4),
                        reasoning=f"Embedding cosine similarity: {similarity:.4f}",
                        method="embedding",
                    )
                )

        scored.sort(key=lambda m: m.score, reverse=True)
        return scored

    # ── text construction ────────────────────────────────────────────────────

    @staticmethod
    def _segment_text(segment: ScriptSegment) -> str:
        """Build a natural-language description from segment fields."""
        parts: list[str] = []
        if segment.visual_direction:
            parts.append(segment.visual_direction)
        if segment.scene_description:
            parts.append(segment.scene_description)
        if segment.content:
            parts.append(segment.content)
        if segment.asset_hints:
            parts.append(" ".join(segment.asset_hints))
        return " ".join(parts) if parts else "unspecified segment"

    @staticmethod
    def _asset_text(asset: AssetInfo) -> str:
        """Build a natural-language description from asset metadata."""
        parts: list[str] = [asset.file_name]
        if asset.file_path:
            parts.append(asset.file_path)
        if asset.metadata:
            parts.extend(f"{k}: {v}" for k, v in asset.metadata.items())
        return " ".join(parts) if parts else "unknown asset"

    # ── embedding helpers ────────────────────────────────────────────────────

    def _is_available(self) -> bool:
        """Return *True* if the OpenAI embedding client can be used."""
        if self._available is not None:
            return self._available

        try:
            import openai as _openai  # noqa: F811
        except ImportError:
            log.warning("openai package is not installed; embedding matching disabled")
            self._available = False
            return False

        try:
            self._client = _openai.OpenAI(api_key=self._api_key)
            # A lightweight validation: the client object is created but we do
            # not make a network call here.  The first real call to ``_embed``
            # will surface auth / network errors.
            self._available = True
        except _openai.OpenAIError:
            log.warning("Failed to initialise OpenAI client; embedding matching disabled")
            self._available = False

        return self._available

    def _embed(self, text: str) -> list[float] | None:
        """Return the embedding vector for *text*, using the cache."""
        if text in self._cache:
            return self._cache[text]

        if self._client is None:
            return None

        try:
            import openai as _openai  # noqa: F811

            response = self._client.embeddings.create(
                model=self.model_name,
                input=text,
            )
            vector = response.data[0].embedding
            self._cache[text] = vector
            return vector
        except Exception:
            log.exception("Embedding API call failed for text (len=%d)", len(text))
            return None

    # ── fallback ─────────────────────────────────────────────────────────────

    @staticmethod
    def _fallback_match(
        segment: ScriptSegment,
        assets: list[AssetInfo],
    ) -> list[ScoredMatch]:
        """Simple token-overlap fallback when embeddings are unavailable."""
        seg_tokens = set(normalize_text(
            " ".join(
                filter(None, [
                    segment.visual_direction,
                    segment.scene_description,
                    segment.content,
                    *segment.asset_hints,
                ])
            )
        ))
        if not seg_tokens:
            return []

        scored: list[ScoredMatch] = []
        for asset in assets:
            asset_tokens = set(normalize_text(
                " ".join(
                    filter(None, [asset.file_name, asset.file_path, *asset.metadata.values()])
                )
            ))
            if not asset_tokens:
                continue
            overlap = seg_tokens & asset_tokens
            if overlap:
                score = len(overlap) / len(seg_tokens | asset_tokens)
                scored.append(
                    ScoredMatch(
                        asset_id=asset.id,
                        score=round(score, 4),
                        reasoning=f"Fallback token overlap: {', '.join(sorted(overlap))}",
                        method="embedding",  # still labelled embedding for the combiner
                    )
                )

        scored.sort(key=lambda m: m.score, reverse=True)
        return scored
