"""Pacing analyzer for Edit Decision Lists.

Analyzes the pacing of an EDL and suggests timing adjustments to match a
target mood. Dialogue and voiceover segments are treated as inflexible
(speech should not be cut), while B-roll and other visual segments are
adjusted freely to hit the target average duration.
"""

from __future__ import annotations

from src.models import (
    EDLEntry,
    EditDecisionList,
    PacingAdjustment,
    PacingResult,
    SegmentType,
)

# Segment types whose duration should not be shortened because they contain
# speech that would sound unnatural if truncated.
_SPEECH_TYPES: frozenset[str] = frozenset({
    SegmentType.DIALOGUE.name,
    SegmentType.VOICEOVER.name,
})

# Segment types that are most flexible for pacing adjustments.
_FLEXIBLE_TYPES: frozenset[str] = frozenset({
    SegmentType.BROLL.name,
    SegmentType.ACTION.name,
    SegmentType.TITLE.name,
    SegmentType.LOWER_THIRD.name,
    SegmentType.TRANSITION.name,
})


class PacingAnalyzer:
    """Analyze EDL pacing and suggest adjustments for a target mood.

    Class Attributes
    ----------------
    MOOD_TARGETS:
        Target average clip durations (seconds) by mood name.
    """

    MOOD_TARGETS: dict[str, float] = {
        "energetic": 2.5,     # fast cuts
        "calm": 8.0,          # long, breathing shots
        "dramatic": 4.0,      # medium, building tension
        "comedic": 3.0,       # snappy timing
        "documentary": 6.0,   # informational pacing
        "cinematic": 5.0,     # balanced
    }

    # ── Public API ──────────────────────────────────────────────────────────

    def analyze(
        self,
        edl: EditDecisionList,
        target_mood: str = "cinematic",
    ) -> PacingResult:
        """Analyze pacing and suggest adjustments.

        Parameters
        ----------
        edl:
            The Edit Decision List to analyze.
        target_mood:
            One of the keys in :attr:`MOOD_TARGETS`, or a custom mood name.
            When the mood is not recognized the ``"cinematic"`` target is used.

        Returns
        -------
        PacingResult
            Contains per-entry adjustments, the current average clip duration,
            and the suggested average clip duration.
        """
        if not edl.entries:
            return PacingResult()

        target_duration = self.MOOD_TARGETS.get(
            target_mood.lower(),
            self.MOOD_TARGETS["cinematic"],
        )

        current_durations = [
            self._entry_duration(entry) for entry in edl.entries
        ]
        current_avg = sum(current_durations) / len(current_durations)

        adjustments: list[PacingAdjustment] = []
        suggested_durations: list[float] = []

        for entry, current_dur in zip(edl.entries, current_durations):
            seg_type = self._infer_segment_type(entry)
            suggested_dur, reason = self._suggest_duration(
                current_dur, target_duration, seg_type,
            )
            suggested_durations.append(suggested_dur)
            adjustments.append(PacingAdjustment(
                edl_entry_index=entry.index,
                current_duration=round(current_dur, 3),
                suggested_duration=round(suggested_dur, 3),
                reason=reason,
            ))

        suggested_avg = (
            sum(suggested_durations) / len(suggested_durations)
            if suggested_durations
            else current_avg
        )

        return PacingResult(
            adjustments=adjustments,
            current_avg_clip_duration=round(current_avg, 3),
            suggested_avg_clip_duration=round(suggested_avg, 3),
        )

    # ── Private helpers ─────────────────────────────────────────────────────

    @staticmethod
    def _entry_duration(entry: EDLEntry) -> float:
        """Compute the duration of an EDL entry from its timeline range."""
        dur = entry.timeline_range.duration_seconds()
        # Guard against zero/negative durations from malformed data.
        return max(dur, 0.1)

    @staticmethod
    def _infer_segment_type(entry: EDLEntry) -> str:
        """Infer the segment type from the entry's notes field.

        The EDL generator writes notes like ``"Segment 3: BROLL"`` or
        ``"Segment 3: broll"``, so we extract the type name and normalize
        to uppercase.  Falls back to ``""`` when the pattern is not found.
        """
        notes = entry.notes or ""
        # Look for "Segment N: TYPE" pattern.
        if ": " in notes:
            raw = notes.split(": ", 1)[1].strip()
            # Normalize: "broll" -> "BROLL", "DIALOGUE" -> "DIALOGUE"
            return raw.upper()
        return ""

    @staticmethod
    def _suggest_duration(
        current: float,
        target: float,
        seg_type: str,
    ) -> tuple[float, str]:
        """Suggest a new duration for a single entry.

        Returns ``(suggested_duration, reason)``.

        Rules:
        - Speech segments (DIALOGUE, VOICEOVER) are never shortened; they
          may be padded slightly if below target but the original is
          preferred.
        - B-roll and other flexible segments are moved toward the target
          duration.  We apply a damped adjustment (50 % of the gap) to
          avoid extreme changes.
        - No segment is reduced below 0.5 s or extended beyond 30 s.
        """
        is_speech = seg_type in _SPEECH_TYPES
        is_flexible = seg_type in _FLEXIBLE_TYPES

        if is_speech:
            # Do not shorten speech segments.
            if current >= target:
                return current, "Dialogue/VO kept at natural duration (not shortened)"
            # Slightly extend with breathing room, but do not force.
            padded = current + min(0.5, (target - current) * 0.25)
            return (
                round(padded, 3),
                "Dialogue/VO lightly padded for pacing but not cut",
            )

        # Flexible / unknown segments: adjust toward target.
        if abs(current - target) < 0.3:
            return current, "Duration already close to target"

        if current > target:
            # Trim: apply 50 % of the gap so changes are gradual.
            adjustment = (current - target) * 0.5
            suggested = max(0.5, current - adjustment)
            return (
                round(suggested, 3),
                f"Trimmed toward target ({target:.1f}s) for faster pacing"
                if is_flexible
                else f"Adjusted toward target ({target:.1f}s)",
            )

        # current < target — extend.
        adjustment = (target - current) * 0.5
        suggested = min(30.0, current + adjustment)
        return (
            round(suggested, 3),
            f"Extended toward target ({target:.1f}s) for slower pacing"
            if is_flexible
            else f"Adjusted toward target ({target:.1f}s)",
        )
