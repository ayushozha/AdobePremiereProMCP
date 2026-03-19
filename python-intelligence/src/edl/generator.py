"""EDL generator -- turns parsed scripts and matched assets into an Edit Decision List.

The generated EDL tells the TypeScript bridge exactly how to assemble the
timeline in Premiere Pro: which asset goes on which track, at what timecode,
with what transitions and text overlays.
"""

from __future__ import annotations

import uuid

from src.models import (
    AssetInfo,
    AssetMatch,
    EDLEntry,
    EDLSettings,
    EditDecisionList,
    Position,
    ScriptSegment,
    SegmentType,
    TextOverlay,
    TextStyle,
    Timecode,
    TimeRange,
    TrackTarget,
    TrackType,
    TransitionInfo,
)

from .timeline_calculator import TimelineCalculator
from .track_assigner import TrackAssigner


class EDLGenerator:
    """Generate an :class:`EditDecisionList` from parsed script segments,
    matched assets, and available asset metadata.

    Usage::

        generator = EDLGenerator()
        edl = generator.generate(segments, matches, assets, settings)

    The generator walks through segments in order, looks up the matched
    asset for each, calculates timeline positions, assigns tracks, adds
    transitions between adjacent clips, and handles text overlays for
    title and lower-third segments.
    """

    def __init__(
        self,
        track_assigner: TrackAssigner | None = None,
    ) -> None:
        self._track_assigner = track_assigner or TrackAssigner()

    def generate(
        self,
        segments: list[ScriptSegment],
        matches: list[AssetMatch],
        assets: list[AssetInfo],
        settings: EDLSettings,
    ) -> EditDecisionList:
        """Generate a complete EDL from script segments and asset matches.

        Parameters
        ----------
        segments:
            Ordered list of parsed script segments.
        matches:
            Asset-to-segment matches produced by the matching engine.
            When multiple matches exist for the same segment the one with
            the highest confidence is used.
        assets:
            Full metadata for every available asset (used to verify
            durations and compute source ranges).
        settings:
            Resolution, frame rate, default transition, and pacing
            controls.

        Returns
        -------
        EditDecisionList
            A fully-populated EDL ready for the TypeScript bridge,
            including entries, text overlays, audio levels, and any
            warnings embedded in entry notes.
        """
        calc = TimelineCalculator(frame_rate=settings.frame_rate)

        # Index helpers -- keep highest-confidence match per segment.
        match_by_segment: dict[int, AssetMatch] = {}
        for m in matches:
            existing = match_by_segment.get(m.segment_index)
            if existing is None or m.confidence > existing.confidence:
                match_by_segment[m.segment_index] = m

        asset_by_id: dict[str, AssetInfo] = {a.id: a for a in assets}

        # Track assignment.
        track_assignments = self._track_assigner.assign_tracks(segments)

        # Accumulators.
        entries: list[EDLEntry] = []
        text_overlays: list[TextOverlay] = []
        audio_levels: dict[int, float] = {}
        warning_notes: list[str] = []
        entry_index = 0

        # Precompute durations list for position calculation.
        durations = [seg.estimated_duration_seconds for seg in segments]

        for seg in segments:
            match = match_by_segment.get(seg.index)

            # ── Text overlays (title / lower third) ──────────────────
            if seg.type in (SegmentType.TITLE, SegmentType.LOWER_THIRD):
                overlay = self._build_text_overlay(seg, calc, durations, settings)
                text_overlays.append(overlay)
                # Title / lower third segments may also have a matched
                # graphic asset.  If no match exists we still emit the
                # text overlay but skip the clip entry.
                if match is None:
                    warning_notes.append(
                        f"Segment {seg.index} ({seg.type.value}): "
                        f"no asset match -- text overlay only."
                    )
                    continue

            # ── Unmatched segment warning ────────────────────────────
            if match is None:
                warning_notes.append(
                    f"Segment {seg.index} ({seg.type.value}): no asset match found."
                )
                continue

            asset = asset_by_id.get(match.asset_id)
            if asset is None:
                warning_notes.append(
                    f"Segment {seg.index}: matched asset '{match.asset_id}' "
                    f"not found in asset list."
                )
                continue

            # ── Duration and source range ────────────────────────────
            segment_duration = seg.estimated_duration_seconds
            if segment_duration <= 0:
                warning_notes.append(
                    f"Segment {seg.index}: zero or negative duration "
                    f"({segment_duration}s), skipping."
                )
                continue

            # Check if the asset is long enough.
            asset_duration = asset.duration_seconds
            if asset_duration > 0 and segment_duration > asset_duration:
                warning_notes.append(
                    f"Segment {seg.index}: required duration ({segment_duration:.2f}s) "
                    f"exceeds asset '{asset.id}' duration ({asset_duration:.2f}s). "
                    f"Clip will be truncated."
                )
                segment_duration = asset_duration

            # Source range.
            source_range = self._compute_source_range(
                match, segment_duration, calc
            )

            # Timeline range -- cumulative position from start.
            timeline_start = calc.calculate_position(seg.index, durations)
            timeline_range = calc.make_time_range(timeline_start, segment_duration)

            # Track.
            track = track_assignments.get(
                seg.index,
                self._track_assigner.get_track_for_type(seg.type),
            )

            # Transition.
            transition = self._maybe_transition(seg, entry_index, entries, settings)

            # Build the video/primary entry.
            entry = EDLEntry(
                index=entry_index,
                source_asset_id=match.asset_id,
                source_range=source_range,
                timeline_range=timeline_range,
                track=track,
                transition=transition,
                effects=[],
                notes=match.reasoning or f"Segment {seg.index}: {seg.type.value}",
            )
            entries.append(entry)

            # ── Companion audio entry ────────────────────────────────
            # Dialogue / action segments on a video track also need
            # their audio placed on A1.
            audio_track = self._track_assigner.needs_audio_track(seg)
            if audio_track is not None and track.type == TrackType.VIDEO:
                entry_index += 1
                audio_entry = EDLEntry(
                    index=entry_index,
                    source_asset_id=match.asset_id,
                    source_range=source_range,
                    timeline_range=timeline_range,
                    track=audio_track,
                    transition=None,
                    effects=[],
                    notes=f"Audio for segment {seg.index}",
                )
                entries.append(audio_entry)

            # ── Audio level defaults ─────────────────────────────────
            if seg.type == SegmentType.MUSIC:
                audio_levels[entry_index] = -6.0
            elif seg.type == SegmentType.SFX:
                audio_levels[entry_index] = -3.0

            entry_index += 1

        # ── Apply transition overlap ─────────────────────────────────
        if (
            settings.default_transition_duration > 0
            and settings.default_transition != "cut"
        ):
            entries = calc.calculate_transitions(
                entries, settings.default_transition_duration
            )

        # ── Check for timeline gaps on V1 ────────────────────────────
        gap_warnings = self._detect_timeline_gaps(entries, calc)
        warning_notes.extend(gap_warnings)

        # ── Build the EDL ────────────────────────────────────────────
        edl = EditDecisionList(
            id=uuid.uuid4().hex[:12],
            name=self._derive_name(segments),
            sequence_resolution=settings.resolution,
            sequence_frame_rate=settings.frame_rate,
            entries=entries,
            text_overlays=text_overlays,
            audio_levels=audio_levels,
        )

        # Stash warnings in the EDL's first entry notes (or create a
        # notes-only entry) so they travel with the EDL object.  The
        # caller can also inspect ``warning_notes`` directly via the
        # ``generate_with_warnings`` helper if needed.
        self._last_warnings = warning_notes

        return edl

    def generate_with_warnings(
        self,
        segments: list[ScriptSegment],
        matches: list[AssetMatch],
        assets: list[AssetInfo],
        settings: EDLSettings,
    ) -> tuple[EditDecisionList, list[str]]:
        """Like :meth:`generate` but also returns the list of warnings.

        Returns
        -------
        tuple[EditDecisionList, list[str]]
            The generated EDL and a list of human-readable warning
            strings.
        """
        edl = self.generate(segments, matches, assets, settings)
        return edl, list(self._last_warnings)

    # ------------------------------------------------------------------
    # Source range computation
    # ------------------------------------------------------------------

    @staticmethod
    def _compute_source_range(
        match: AssetMatch,
        segment_duration: float,
        calc: TimelineCalculator,
    ) -> TimeRange:
        """Determine the source in/out range within the asset.

        If the match includes a suggested range, use that.  Otherwise
        start from 00:00:00:00 and run for *segment_duration*.
        """
        if match.suggested_range is not None:
            return match.suggested_range
        return calc.make_time_range(0.0, segment_duration)

    # ------------------------------------------------------------------
    # Transition logic
    # ------------------------------------------------------------------

    @staticmethod
    def _maybe_transition(
        segment: ScriptSegment,
        entry_index: int,
        existing_entries: list[EDLEntry],
        settings: EDLSettings,
    ) -> TransitionInfo | None:
        """Decide whether to add a transition before this clip.

        Rules:
        - No transition on the very first entry.
        - ``TRANSITION`` segment types always get a transition.
        - Otherwise use the default transition from settings unless it is
          ``"cut"`` (a cut is the absence of a transition).
        """
        if entry_index == 0 or not existing_entries:
            return None

        if settings.default_transition == "cut" and segment.type != SegmentType.TRANSITION:
            return None

        transition_type = settings.default_transition
        duration = settings.default_transition_duration

        # TRANSITION segments may carry their own type hint in
        # visual_direction.
        if segment.type == SegmentType.TRANSITION:
            if segment.visual_direction:
                transition_type = (
                    segment.visual_direction.strip().lower().replace(" ", "_")
                )
            if duration <= 0:
                duration = 1.0  # Sensible fallback.

        if duration <= 0:
            return None

        return TransitionInfo(
            type=transition_type,
            duration_seconds=duration,
            alignment="center",
        )

    # ------------------------------------------------------------------
    # Text overlay builder
    # ------------------------------------------------------------------

    @staticmethod
    def _build_text_overlay(
        segment: ScriptSegment,
        calc: TimelineCalculator,
        durations: list[float],
        settings: EDLSettings,
    ) -> TextOverlay:
        """Build a :class:`TextOverlay` from a title or lower-third segment."""
        timeline_start = calc.calculate_position(segment.index, durations)
        duration = (
            segment.estimated_duration_seconds
            if segment.estimated_duration_seconds > 0
            else 3.0
        )

        if segment.type == SegmentType.LOWER_THIRD:
            style = TextStyle(
                font_family="Arial",
                font_size=32.0,
                color_hex="#FFFFFF",
                alignment="left",
                background_color_hex="#000000",
                background_opacity=0.7,
                position=Position(x=0.1, y=0.85),
            )
        else:
            # Full-screen title.
            style = TextStyle(
                font_family="Arial",
                font_size=72.0,
                color_hex="#FFFFFF",
                alignment="center",
                background_color_hex="#000000",
                background_opacity=0.0,
                position=Position(x=0.5, y=0.5),
            )

        track = TrackTarget(type=TrackType.VIDEO, track_index=2)  # V3

        return TextOverlay(
            text=segment.content,
            style=style,
            track=track,
            position=calc.seconds_to_timecode(timeline_start),
            duration_seconds=duration,
        )

    # ------------------------------------------------------------------
    # Gap detection
    # ------------------------------------------------------------------

    @staticmethod
    def _detect_timeline_gaps(
        entries: list[EDLEntry],
        calc: TimelineCalculator,
    ) -> list[str]:
        """Detect gaps between clips on the primary video track (V1).

        Returns a list of human-readable warning strings.
        """
        v1_spans: list[tuple[int, float, float]] = []

        for entry in entries:
            if entry.track.type == TrackType.VIDEO and entry.track.track_index == 0:
                start = calc.timecode_to_seconds(entry.timeline_range.in_point)
                end = calc.timecode_to_seconds(entry.timeline_range.out_point)
                if end > start:
                    v1_spans.append((entry.index, start, end))

        if not v1_spans:
            return []

        v1_spans.sort(key=lambda s: s[1])
        warnings: list[str] = []

        for i in range(len(v1_spans) - 1):
            current_idx, _cs, current_end = v1_spans[i]
            next_idx, next_start, _ne = v1_spans[i + 1]
            gap = next_start - current_end

            if gap > 1e-6:
                warnings.append(
                    f"Timeline gap of {gap:.3f}s between entry {current_idx} "
                    f"and entry {next_idx} on V1."
                )

        return warnings

    # ------------------------------------------------------------------
    # Name derivation
    # ------------------------------------------------------------------

    @staticmethod
    def _derive_name(segments: list[ScriptSegment]) -> str:
        """Derive a sequence name from the segments.

        Uses the first segment's content (truncated) or falls back to a
        generic name.
        """
        for seg in segments:
            if seg.content.strip():
                clean = seg.content.strip().replace("\n", " ")
                if len(clean) > 60:
                    clean = clean[:57] + "..."
                return f"Sequence - {clean}"
        return "Untitled Sequence"
