"""Timeline math utilities for frame-accurate EDL generation.

All arithmetic is performed via integer frame counts so that floating-point
drift cannot accumulate over long sequences.  Every public method that
accepts or returns seconds snaps to the nearest frame boundary first.
"""

from __future__ import annotations

import math

from src.models import EDLEntry, Timecode, TimeRange


class TimelineCalculator:
    """Frame-accurate timeline arithmetic at a fixed frame rate.

    Parameters
    ----------
    frame_rate:
        Frames per second for the target timeline.  Common values are
        23.976, 24, 25, 29.97, 30, 50, 59.94, 60.
    """

    def __init__(self, frame_rate: float = 24.0) -> None:
        if frame_rate <= 0 or not math.isfinite(frame_rate):
            raise ValueError(f"Invalid frame rate: {frame_rate}. Must be a positive finite number.")
        self.frame_rate = frame_rate
        self._frames_per_second = math.ceil(frame_rate)

    # ------------------------------------------------------------------
    # Timecode <-> seconds
    # ------------------------------------------------------------------

    def seconds_to_timecode(self, seconds: float) -> Timecode:
        """Convert an absolute number of seconds to a :class:`Timecode`.

        The value is snapped to the nearest frame boundary before conversion
        so the resulting timecode is always frame-aligned.
        """
        if seconds < 0:
            raise ValueError(f"Seconds value {seconds} must be non-negative.")
        total_frames = round(seconds * self.frame_rate)
        return self._frames_to_timecode(total_frames)

    def timecode_to_seconds(self, tc: Timecode) -> float:
        """Convert a :class:`Timecode` to an absolute number of seconds.

        Goes through an integer frame count first so the result is as
        precise as the underlying frame grid allows.
        """
        total_frames = self._timecode_to_frames(tc)
        return total_frames / self.frame_rate

    # ------------------------------------------------------------------
    # Timecode <-> frames
    # ------------------------------------------------------------------

    def timecode_to_frames(self, tc: Timecode) -> int:
        """Convert a :class:`Timecode` to an absolute frame count."""
        return self._timecode_to_frames(tc)

    def frames_to_timecode(self, frames: int) -> Timecode:
        """Convert an absolute frame count to a :class:`Timecode`."""
        if frames < 0:
            raise ValueError(f"Frame count {frames} must be non-negative.")
        return self._frames_to_timecode(frames)

    # ------------------------------------------------------------------
    # Position calculation
    # ------------------------------------------------------------------

    def calculate_position(self, index: int, durations: list[float]) -> float:
        """Calculate the timeline start position for segment *index*.

        The position is the cumulative sum of all preceding durations,
        snapped to a frame boundary.

        Parameters
        ----------
        index:
            Zero-based index of the segment whose start position is needed.
        durations:
            List of segment durations in seconds (one per segment).
        """
        if index < 0 or index > len(durations):
            raise IndexError(
                f"Segment index {index} is out of range for {len(durations)} durations."
            )
        cumulative = sum(durations[:index])
        return self.snap_to_frame(cumulative)

    # ------------------------------------------------------------------
    # Frame snapping
    # ------------------------------------------------------------------

    def snap_to_frame(self, seconds: float) -> float:
        """Snap a seconds value to the nearest frame boundary.

        This eliminates sub-frame offsets that would cause drift in the
        timeline.  The returned value is ``round(seconds * fps) / fps``.
        """
        if seconds < 0:
            raise ValueError(f"Seconds value {seconds} must be non-negative.")
        frame = round(seconds * self.frame_rate)
        return frame / self.frame_rate

    # ------------------------------------------------------------------
    # Time range helpers
    # ------------------------------------------------------------------

    def make_time_range(self, start_seconds: float, duration_seconds: float) -> TimeRange:
        """Build a :class:`TimeRange` from a start point and duration.

        Both the in-point and out-point are frame-aligned.
        """
        snapped_start = self.snap_to_frame(start_seconds)
        snapped_end = self.snap_to_frame(snapped_start + duration_seconds)
        return TimeRange(
            in_point=self.seconds_to_timecode(snapped_start),
            out_point=self.seconds_to_timecode(snapped_end),
        )

    def time_range_duration_seconds(self, time_range: TimeRange) -> float:
        """Return the duration of a :class:`TimeRange` in seconds."""
        in_frames = self._timecode_to_frames(time_range.in_point)
        out_frames = self._timecode_to_frames(time_range.out_point)
        return max(0, out_frames - in_frames) / self.frame_rate

    # ------------------------------------------------------------------
    # Transition overlap
    # ------------------------------------------------------------------

    def calculate_transitions(
        self,
        entries: list[EDLEntry],
        transition_duration: float,
    ) -> list[EDLEntry]:
        """Adjust entries so that transitions eat into adjacent clip durations.

        For each entry that has a transition, the overlap is split evenly
        between the outgoing clip (previous entry on the same track) and the
        incoming clip (current entry).  Both clips are shortened by half the
        transition duration.

        Parameters
        ----------
        entries:
            EDL entries in timeline order.  Entries are **not** modified
            in-place -- a new list of adjusted copies is returned.
        transition_duration:
            Duration of each transition in seconds.

        Returns
        -------
        list[EDLEntry]
            New entries with adjusted timeline ranges to account for
            transition overlap.
        """
        if transition_duration <= 0:
            return list(entries)

        half_transition = self.snap_to_frame(transition_duration / 2.0)
        adjusted: list[EDLEntry] = []

        # Group entries by track key so we only overlap clips on the same track.
        track_groups: dict[str, list[int]] = {}
        for i, entry in enumerate(entries):
            key = f"{entry.track.type}:{entry.track.track_index}"
            track_groups.setdefault(key, []).append(i)

        # Start with copies of all entries.
        entry_copies = [entry.model_copy(deep=True) for entry in entries]

        for _track_key, indices in track_groups.items():
            # Sort by timeline in-point within this track.
            indices.sort(
                key=lambda idx: self._timecode_to_frames(entry_copies[idx].timeline_range.in_point)
            )

            for pos, idx in enumerate(indices):
                entry = entry_copies[idx]
                if entry.transition is None:
                    continue
                if pos == 0:
                    # First clip on track -- no previous clip to overlap with.
                    # Only shorten the incoming clip's start.
                    current_in_sec = self.timecode_to_seconds(entry.timeline_range.in_point)
                    current_out_sec = self.timecode_to_seconds(entry.timeline_range.out_point)
                    # Incoming clip starts earlier by half the transition.
                    new_in = max(0.0, current_in_sec - half_transition)
                    entry.timeline_range = TimeRange(
                        in_point=self.seconds_to_timecode(new_in),
                        out_point=self.seconds_to_timecode(current_out_sec),
                    )
                    continue

                prev_idx = indices[pos - 1]
                prev_entry = entry_copies[prev_idx]

                # Shorten previous clip's out-point by half_transition.
                prev_out_sec = self.timecode_to_seconds(prev_entry.timeline_range.out_point)
                prev_in_sec = self.timecode_to_seconds(prev_entry.timeline_range.in_point)
                new_prev_out = max(prev_in_sec, prev_out_sec - half_transition)
                prev_entry.timeline_range = TimeRange(
                    in_point=prev_entry.timeline_range.in_point,
                    out_point=self.seconds_to_timecode(new_prev_out),
                )

                # Shorten current clip's in-point by half_transition (move it earlier).
                current_in_sec = self.timecode_to_seconds(entry.timeline_range.in_point)
                current_out_sec = self.timecode_to_seconds(entry.timeline_range.out_point)
                new_in = max(0.0, current_in_sec - half_transition)
                entry.timeline_range = TimeRange(
                    in_point=self.seconds_to_timecode(new_in),
                    out_point=self.seconds_to_timecode(current_out_sec),
                )

        return entry_copies

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _timecode_to_frames(self, tc: Timecode) -> int:
        """Convert a timecode to an absolute frame count."""
        fps = self._frames_per_second
        total_seconds = tc.hours * 3600 + tc.minutes * 60 + tc.seconds
        return total_seconds * fps + tc.frames

    def _frames_to_timecode(self, frames: int) -> Timecode:
        """Convert an absolute frame count to a timecode."""
        fps = self._frames_per_second
        remaining = frames

        ff = remaining % fps
        remaining = (remaining - ff) // fps

        ss = remaining % 60
        remaining = (remaining - ss) // 60

        mm = remaining % 60
        hh = (remaining - mm) // 60

        return Timecode(
            hours=hh,
            minutes=mm,
            seconds=ss,
            frames=ff,
            frame_rate=self.frame_rate,
        )
