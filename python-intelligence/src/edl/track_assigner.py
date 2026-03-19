"""Track assignment logic for EDL generation.

Maps script segments to the appropriate timeline tracks based on segment type,
and resolves conflicts when two segments want the same track at the same time.

Track layout convention (matching the TypeScript bridge expectations):

    V1 -- Primary video (dialogue, action, on-camera)
    V2 -- B-roll (supplementary footage that can overlap V1)
    V3 -- Titles / lower thirds (text overlays rendered as graphics)
    A1 -- Main audio (dialogue audio, on-camera sound)
    A2 -- Voiceover
    A3 -- Music
    A4 -- Sound effects
"""

from __future__ import annotations

from src.models import (
    ScriptSegment,
    SegmentType,
    TrackTarget,
    TrackType,
)


class TrackAssigner:
    """Assign script segments to timeline tracks.

    The assigner applies a deterministic mapping from :class:`SegmentType`
    to :class:`TrackTarget` and resolves conflicts when two segments would
    occupy the same track at the same time.
    """

    # Default mapping from segment type to (TrackType, track_index).
    # Track indices are zero-based to match Premiere Pro's internal model.
    _DEFAULT_TRACK_MAP: dict[SegmentType, TrackTarget] = {
        # Video tracks
        SegmentType.DIALOGUE:    TrackTarget(type=TrackType.VIDEO, track_index=0),   # V1
        SegmentType.ACTION:      TrackTarget(type=TrackType.VIDEO, track_index=0),   # V1
        SegmentType.BROLL:       TrackTarget(type=TrackType.VIDEO, track_index=1),   # V2
        SegmentType.TITLE:       TrackTarget(type=TrackType.VIDEO, track_index=2),   # V3
        SegmentType.LOWER_THIRD: TrackTarget(type=TrackType.VIDEO, track_index=2),   # V3
        SegmentType.TRANSITION:  TrackTarget(type=TrackType.VIDEO, track_index=0),   # V1

        # Audio tracks
        SegmentType.VOICEOVER:   TrackTarget(type=TrackType.AUDIO, track_index=1),   # A2
        SegmentType.MUSIC:       TrackTarget(type=TrackType.AUDIO, track_index=2),   # A3
        SegmentType.SFX:         TrackTarget(type=TrackType.AUDIO, track_index=3),   # A4
    }

    def __init__(
        self,
        custom_map: dict[SegmentType, TrackTarget] | None = None,
    ) -> None:
        """Initialise the track assigner.

        Parameters
        ----------
        custom_map:
            Optional overrides for the default segment-type-to-track mapping.
            Any segment types not present in *custom_map* fall back to the
            built-in defaults.
        """
        self._track_map = dict(self._DEFAULT_TRACK_MAP)
        if custom_map:
            self._track_map.update(custom_map)

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def assign_tracks(
        self,
        segments: list[ScriptSegment],
    ) -> dict[int, TrackTarget]:
        """Map segment indices to track targets.

        Parameters
        ----------
        segments:
            Ordered list of parsed script segments.

        Returns
        -------
        dict[int, TrackTarget]
            Mapping from segment index to the track where it should be
            placed.  Every segment in the input list will have an entry
            in the result.
        """
        raw_assignments = self._initial_assignment(segments)
        resolved = self._resolve_conflicts(segments, raw_assignments)
        return resolved

    def get_track_for_type(self, segment_type: SegmentType) -> TrackTarget:
        """Return the default track target for a given segment type.

        Falls back to V1 for unknown / unspecified types.
        """
        return self._track_map.get(
            segment_type,
            TrackTarget(type=TrackType.VIDEO, track_index=0),
        )

    def needs_audio_track(self, segment: ScriptSegment) -> TrackTarget | None:
        """Return the companion audio track for a segment, if applicable.

        Video segments that carry dialogue or action audio also need a
        corresponding audio placement on A1.  Pure audio segment types
        (VO, music, SFX) are already routed to audio tracks by the
        primary mapping and return ``None`` here.
        """
        if segment.type in (SegmentType.DIALOGUE, SegmentType.ACTION):
            return TrackTarget(type=TrackType.AUDIO, track_index=0)  # A1
        return None

    # ------------------------------------------------------------------
    # Internals
    # ------------------------------------------------------------------

    def _initial_assignment(
        self,
        segments: list[ScriptSegment],
    ) -> dict[int, TrackTarget]:
        """Produce a first-pass assignment using the static type map."""
        assignments: dict[int, TrackTarget] = {}
        for seg in segments:
            assignments[seg.index] = self.get_track_for_type(seg.type)
        return assignments

    def _resolve_conflicts(
        self,
        segments: list[ScriptSegment],
        assignments: dict[int, TrackTarget],
    ) -> dict[int, TrackTarget]:
        """Resolve conflicts where two segments overlap on the same track.

        The strategy is to bump the later segment to the next available
        track of the same type.  For example, if two B-roll clips overlap
        on V2 the second one moves to V3 (or V4, etc.).

        Overlap detection uses the ``estimated_duration_seconds`` on each
        segment and assumes segments are laid out sequentially according
        to their index order.
        """
        # Build a timeline of occupied intervals per track key.
        # Track key = "video:0", "audio:2", etc.
        seg_by_index: dict[int, ScriptSegment] = {s.index: s for s in segments}

        # Compute cumulative start times.
        ordered_indices = sorted(assignments.keys())
        start_times: dict[int, float] = {}
        cumulative = 0.0
        for idx in ordered_indices:
            start_times[idx] = cumulative
            seg = seg_by_index.get(idx)
            if seg:
                cumulative += seg.estimated_duration_seconds

        # occupied: track_key -> list of (start, end) intervals
        occupied: dict[str, list[tuple[float, float]]] = {}
        resolved: dict[int, TrackTarget] = {}

        for idx in ordered_indices:
            seg = seg_by_index.get(idx)
            if seg is None:
                resolved[idx] = assignments[idx]
                continue

            target = assignments[idx]
            start = start_times[idx]
            end = start + seg.estimated_duration_seconds

            # B-roll and titles/lower-thirds may overlap with the primary
            # video track, so we only check for conflicts within the
            # *same* track, not across tracks.
            placed = False
            candidate = TrackTarget(type=target.type, track_index=target.track_index)

            # Try up to 8 tracks of the same type before giving up.
            for offset in range(8):
                candidate = TrackTarget(
                    type=target.type,
                    track_index=target.track_index + offset,
                )
                key = f"{candidate.type}:{candidate.track_index}"
                intervals = occupied.setdefault(key, [])

                if not self._overlaps(intervals, start, end):
                    intervals.append((start, end))
                    resolved[idx] = candidate
                    placed = True
                    break

            if not placed:
                # All 8 candidate tracks are full -- fall back to original.
                key = f"{target.type}:{target.track_index}"
                occupied.setdefault(key, []).append((start, end))
                resolved[idx] = target

        return resolved

    @staticmethod
    def _overlaps(
        intervals: list[tuple[float, float]],
        start: float,
        end: float,
    ) -> bool:
        """Return True if (start, end) overlaps any existing interval."""
        for s, e in intervals:
            if start < e and end > s:
                return True
        return False
