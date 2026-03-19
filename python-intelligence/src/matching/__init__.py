"""Asset matching module.

Matches script segments to available media assets using keyword overlap,
embedding-based semantic similarity, or a weighted hybrid of both.
"""

from .embedding_matcher import EmbeddingMatcher
from .keyword_matcher import KeywordMatcher
from .matcher import AssetMatcher
from .scoring import ScoredMatch, combine_scores, cosine_similarity, normalize_text
from .suggest import suggest_assets

__all__ = [
    "AssetMatcher",
    "EmbeddingMatcher",
    "KeywordMatcher",
    "ScoredMatch",
    "combine_scores",
    "cosine_similarity",
    "normalize_text",
    "suggest_assets",
]
