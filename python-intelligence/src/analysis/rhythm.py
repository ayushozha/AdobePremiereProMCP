"""Rhythm pattern detection and suggestion for clip sequences.

Analyzes the temporal pattern of clip durations in an EDL to classify the
editing rhythm, and suggests duration sequences that match a target mood.
"""

from __future__ import annotations

import math


class RhythmAnalyzer:
    """Detect and suggest rhythm patterns for clip sequences."""

    # ── Public API ──────────────────────────────────────────────────────────

    def detect_pattern(self, durations: list[float]) -> str:
        """Detect the rhythm pattern of clip durations.

        Parameters
        ----------
        durations:
            List of clip durations in seconds.

        Returns
        -------
        str
            One of ``"constant"``, ``"accelerating"``, ``"decelerating"``,
            ``"alternating"``, or ``"random"``.
        """
        if len(durations) < 2:
            return "constant"

        diffs = [durations[i + 1] - durations[i] for i in range(len(durations) - 1)]

        if self._is_constant(durations):
            return "constant"
        if self._is_alternating(diffs):
            return "alternating"
        if self._is_monotone(diffs, direction="decreasing"):
            return "accelerating"
        if self._is_monotone(diffs, direction="increasing"):
            return "decelerating"

        return "random"

    def suggest_rhythm(
        self,
        target_mood: str,
        num_segments: int,
    ) -> list[float]:
        """Suggest a rhythm pattern of durations for the target mood.

        Parameters
        ----------
        target_mood:
            Mood identifier (e.g. ``"dramatic"``, ``"calm"``).
        num_segments:
            Number of segments to generate durations for.

        Returns
        -------
        list[float]
            Suggested durations (one per segment).
        """
        if num_segments <= 0:
            return []

        mood = target_mood.lower()

        if mood == "dramatic":
            return self._dramatic_rhythm(num_segments)
        if mood == "energetic":
            return self._energetic_rhythm(num_segments)
        if mood == "calm":
            return self._calm_rhythm(num_segments)
        if mood == "comedic":
            return self._comedic_rhythm(num_segments)
        if mood == "documentary":
            return self._documentary_rhythm(num_segments)
        # Default: cinematic — balanced with gentle variation
        return self._cinematic_rhythm(num_segments)

    # ── Pattern classification helpers ──────────────────────────────────────

    @staticmethod
    def _is_constant(durations: list[float], tolerance: float = 0.5) -> bool:
        """Return True if all durations are within *tolerance* of the mean."""
        avg = sum(durations) / len(durations)
        return all(abs(d - avg) <= tolerance for d in durations)

    @staticmethod
    def _is_alternating(diffs: list[float], min_amplitude: float = 0.3) -> bool:
        """Return True if the sign of consecutive differences alternates.

        Requires at least 3 diffs and that most sign-changes happen.
        """
        if len(diffs) < 3:
            return False

        sign_changes = 0
        for i in range(len(diffs) - 1):
            if abs(diffs[i]) < min_amplitude or abs(diffs[i + 1]) < min_amplitude:
                continue
            if (diffs[i] > 0 and diffs[i + 1] < 0) or (diffs[i] < 0 and diffs[i + 1] > 0):
                sign_changes += 1

        possible_changes = len(diffs) - 1
        return sign_changes >= possible_changes * 0.6

    @staticmethod
    def _is_monotone(
        diffs: list[float],
        direction: str = "decreasing",
        tolerance: float = 0.3,
    ) -> bool:
        """Return True if the majority of diffs consistently trend in *direction*.

        ``"decreasing"`` means durations get shorter (accelerating edits).
        ``"increasing"`` means durations get longer (decelerating edits).
        """
        if not diffs:
            return False

        if direction == "decreasing":
            count = sum(1 for d in diffs if d < -tolerance)
        else:
            count = sum(1 for d in diffs if d > tolerance)

        return count >= len(diffs) * 0.6

    # ── Rhythm generators ───────────────────────────────────────────────────

    @staticmethod
    def _dramatic_rhythm(n: int) -> list[float]:
        """Gradual acceleration toward a climax then a brief hold.

        Starts with longer shots and builds to rapid cuts at the end.
        """
        if n == 1:
            return [4.0]
        result: list[float] = []
        for i in range(n):
            # Progress 0..1
            t = i / (n - 1) if n > 1 else 0.0
            # Exponential decay from 6 s down to 2 s
            duration = 6.0 - 4.0 * (t ** 1.5)
            result.append(round(max(1.5, duration), 2))
        return result

    @staticmethod
    def _energetic_rhythm(n: int) -> list[float]:
        """Rapid, varied cuts with a base around 2-3 s."""
        if n == 1:
            return [2.5]
        result: list[float] = []
        for i in range(n):
            # Alternate between short and very short
            base = 2.5
            variation = 1.0 * math.sin(i * math.pi)
            duration = base + variation * 0.5
            result.append(round(max(1.0, duration), 2))
        return result

    @staticmethod
    def _calm_rhythm(n: int) -> list[float]:
        """Long, steady shots with gentle undulation."""
        if n == 1:
            return [8.0]
        result: list[float] = []
        for i in range(n):
            base = 8.0
            wave = math.sin(i * math.pi / 3) * 1.5
            result.append(round(max(5.0, base + wave), 2))
        return result

    @staticmethod
    def _comedic_rhythm(n: int) -> list[float]:
        """Snappy base with occasional pauses (longer holds for timing)."""
        if n == 1:
            return [3.0]
        result: list[float] = []
        for i in range(n):
            # Every 3rd clip gets a beat (longer pause)
            if i % 3 == 2:
                result.append(4.5)
            else:
                result.append(round(2.5 + (i % 2) * 0.5, 2))
        return result

    @staticmethod
    def _documentary_rhythm(n: int) -> list[float]:
        """Measured pacing with consistent durations, slight lengthening."""
        if n == 1:
            return [6.0]
        result: list[float] = []
        for i in range(n):
            # Slight linear increase for building information
            duration = 5.0 + (i / n) * 2.0
            result.append(round(duration, 2))
        return result

    @staticmethod
    def _cinematic_rhythm(n: int) -> list[float]:
        """Balanced durations with gentle sinusoidal variation."""
        if n == 1:
            return [5.0]
        result: list[float] = []
        for i in range(n):
            base = 5.0
            wave = math.sin(i * math.pi / 4) * 1.0
            result.append(round(max(3.0, base + wave), 2))
        return result
