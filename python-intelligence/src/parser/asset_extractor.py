"""Extract asset-matching keywords from script text.

Performs lightweight keyword extraction without heavy NLP dependencies.
Pulls nouns, locations, actions, and descriptive terms from segment
descriptions that can later be used to match against available media assets.
"""

from __future__ import annotations

import re

# ── Stop words to filter out ─────────────────────────────────────────────────

_STOP_WORDS: frozenset[str] = frozenset({
    # Articles / determiners
    "a", "an", "the", "this", "that", "these", "those",
    # Prepositions
    "in", "on", "at", "to", "for", "of", "with", "by", "from", "up", "about",
    "into", "through", "during", "before", "after", "above", "below", "between",
    "under", "over", "out", "off", "down", "near", "around",
    # Conjunctions
    "and", "but", "or", "nor", "so", "yet", "both", "either", "neither",
    # Pronouns
    "i", "me", "my", "we", "us", "our", "you", "your", "he", "him", "his",
    "she", "her", "it", "its", "they", "them", "their",
    # Common verbs (too generic to be useful as asset keywords)
    "is", "are", "was", "were", "be", "been", "being", "have", "has", "had",
    "do", "does", "did", "will", "would", "shall", "should", "may", "might",
    "can", "could", "must", "get", "gets", "got", "see", "sees", "saw",
    "go", "goes", "went", "come", "comes", "came", "make", "makes", "made",
    "take", "takes", "took", "give", "gives", "gave", "know", "knows", "knew",
    "say", "says", "said", "tell", "tells", "told", "let", "just",
    # Adverbs / misc
    "not", "no", "very", "really", "also", "too", "then", "now", "here",
    "there", "when", "where", "how", "all", "each", "every", "some", "any",
    "more", "most", "other", "than", "only", "own", "same",
    # Script-specific words that are structural, not descriptive
    "cut", "fade", "shot", "angle", "close", "wide", "medium", "camera",
    "pan", "tilt", "zoom", "rack", "focus", "track", "dolly",
    "cont", "continued", "beat", "pause", "moment",
})

# Words that indicate useful visual/audio categories.
_CATEGORY_KEYWORDS: dict[str, list[str]] = {
    "aerial": ["aerial", "drone", "overhead", "bird's eye", "birds eye"],
    "landscape": ["landscape", "scenery", "vista", "panorama", "horizon"],
    "portrait": ["portrait", "headshot", "face", "closeup", "close-up"],
    "timelapse": ["timelapse", "time-lapse", "time lapse", "hyperlapse"],
    "slowmo": ["slow motion", "slow-motion", "slowmo", "slow mo"],
    "underwater": ["underwater", "submerged", "diving", "ocean floor"],
}


def _tokenize(text: str) -> list[str]:
    """Split text into lowercase alphanumeric tokens."""
    return re.findall(r"[a-z][a-z0-9'-]*[a-z0-9]|[a-z]", text.lower())


def _extract_category_tags(text: str) -> list[str]:
    """Check for category keywords / multi-word phrases in the text."""
    lower = text.lower()
    tags: list[str] = []
    for category, phrases in _CATEGORY_KEYWORDS.items():
        if any(phrase in lower for phrase in phrases):
            tags.append(category)
    return tags


def _extract_quoted_phrases(text: str) -> list[str]:
    """Extract quoted phrases which are often specific asset references."""
    return [m.strip().lower() for m in re.findall(r'"([^"]+)"', text) if m.strip()]


def extract_hints(text: str) -> list[str]:
    """Extract asset-matching keywords from a text description.

    The extraction pipeline:
    1. Pull out quoted phrases (treated as exact references).
    2. Check for category tags (aerial, landscape, etc.).
    3. Tokenize the text and keep meaningful words (>2 chars, not stop words).
    4. Deduplicate while preserving rough order of importance.

    Examples:
        >>> extract_hints("aerial shot of a sunset over the ocean")
        ['aerial', 'sunset', 'ocean', 'landscape']
        >>> extract_hints('[Visual: busy city street at night with neon signs]')
        ['busy', 'city', 'street', 'night', 'neon', 'signs']

    Args:
        text: Free-form text description from a script segment.

    Returns:
        Deduplicated list of lowercase keyword strings.
    """
    if not text or not text.strip():
        return []

    hints: list[str] = []
    seen: set[str] = set()

    def _add(keyword: str) -> None:
        kw = keyword.strip().lower()
        if kw and kw not in seen:
            seen.add(kw)
            hints.append(kw)

    # 1. Quoted phrases
    for phrase in _extract_quoted_phrases(text):
        _add(phrase)

    # 2. Category tags
    for tag in _extract_category_tags(text):
        _add(tag)

    # 3. Individual tokens
    for token in _tokenize(text):
        if len(token) > 2 and token not in _STOP_WORDS:
            _add(token)

    return hints
