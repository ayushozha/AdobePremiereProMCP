"""Scoring utilities for asset matching.

Provides the ``ScoredMatch`` container, text normalisation helpers, cosine
similarity, and a combiner that merges keyword and embedding scores.
"""

from __future__ import annotations

import math
import re
from dataclasses import dataclass, field


@dataclass(slots=True)
class ScoredMatch:
    """A single candidate match with its numeric score and explanation."""

    asset_id: str
    score: float  # 0.0 – 1.0
    reasoning: str
    method: str  # "keyword" | "embedding" | "hybrid"


# ── Score combination ────────────────────────────────────────────────────────


def combine_scores(
    keyword_score: float,
    embedding_score: float,
    keyword_weight: float = 0.3,
) -> float:
    """Return the weighted average of *keyword_score* and *embedding_score*.

    Parameters
    ----------
    keyword_score:
        Score from keyword matching (0.0–1.0).
    embedding_score:
        Score from embedding similarity (0.0–1.0).
    keyword_weight:
        Weight given to the keyword score.  The embedding score receives
        ``1.0 - keyword_weight``.
    """
    embedding_weight = 1.0 - keyword_weight
    return keyword_score * keyword_weight + embedding_score * embedding_weight


# ── Text normalisation ───────────────────────────────────────────────────────

# Pre-compiled regex used by ``normalize_text``.
_CAMEL_BOUNDARY = re.compile(r"(?<=[a-z])(?=[A-Z])|(?<=[A-Z])(?=[A-Z][a-z])")
_NON_ALPHA = re.compile(r"[^a-z0-9]+")
_FILE_EXT = re.compile(r"\.[a-zA-Z0-9]{1,5}$")


def normalize_text(text: str) -> list[str]:
    """Normalise *text* into a list of lowercase tokens.

    Handles:
    * File extensions – stripped (``"sunset.mp4"`` -> ``["sunset"]``)
    * camelCase / PascalCase splitting
    * Underscores, hyphens, dots, spaces as delimiters
    * Common video naming patterns (``"B-roll_sunset_001"`` -> ``["broll", "sunset", "001"]``)
    """
    # Strip file extension first.
    text = _FILE_EXT.sub("", text)

    # Split on camelCase boundaries *before* lowercasing so the regex works.
    text = _CAMEL_BOUNDARY.sub(" ", text)
    text = text.lower()

    # Replace non-alphanumeric runs with a single space and split.
    tokens = _NON_ALPHA.sub(" ", text).split()

    # Deduplicate while preserving order.
    seen: set[str] = set()
    unique: list[str] = []
    for tok in tokens:
        if tok and tok not in seen:
            seen.add(tok)
            unique.append(tok)
    return unique


# ── Cosine similarity ────────────────────────────────────────────────────────


def cosine_similarity(a: list[float], b: list[float]) -> float:
    """Compute cosine similarity between vectors *a* and *b*.

    Returns 0.0 when either vector has zero magnitude.  Does **not** depend on
    numpy so the module can be used without heavy dependencies.
    """
    if len(a) != len(b):
        raise ValueError(
            f"Vectors must have equal length, got {len(a)} and {len(b)}"
        )
    dot = sum(x * y for x, y in zip(a, b))
    mag_a = math.sqrt(sum(x * x for x in a))
    mag_b = math.sqrt(sum(y * y for y in b))
    if mag_a == 0.0 or mag_b == 0.0:
        return 0.0
    return dot / (mag_a * mag_b)
