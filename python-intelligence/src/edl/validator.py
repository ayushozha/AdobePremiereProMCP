"""EDL validation -- catch errors before sending to the TypeScript bridge.

Produces hard errors (execution must stop) and soft warnings (execution can
continue but the result may be unexpected).  This mirrors the validation in
``ts-bridge/src/timeline/validator.ts`` so problems are caught on the Python
side before any gRPC round-trip.
"""

from __future__ import annotations

import math

from src.models import (
    EditDecisionList,
    Timecode,
    TimeRange,
    TrackTarget,
    ValidationIssue,
    ValidationResult,
)


class EDLValidator:
    """Validate an :class:`EditDecisionList` for timeline conflicts and issues.

    Usage::

        validator = EDLValidator()
        result = validator.validate(edl)
        if not result.valid:
            for err in result.errors:
                print(f"ERROR [{err.entry_index}]: {err.message}")
    """

    def validate(self, edl: EditDecisionList) -> ValidationResult:
        """Validate *edl* and return all discovered issues.

        The returned :class:`ValidationResult` has ``valid=True`` only when
        there are zero errors.  Warnings are always included regardless.
        """
        issues: list[ValidationIssue] = []

        self._validate_global_settings(edl, issues)
        self._validate_entries(edl, issues)
        self._check_overlapping_clips(edl, issues)
        self._check_timeline_gaps(edl, issues)
        self._validate_text_overlays(edl, issues)

        errors = [i for i in issues if i.severity == "error"]
        warnings = [i for i in issues if i.severity == "warning"]

        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
        )

    # ------------------------------------------------------------------
    # Global settings
    # ------------------------------------------------------------------

    def _validate_global_settings(
        self,
        edl: EditDecisionList,
        issues: list[ValidationIssue],
    ) -> None:
        if not edl.name or not edl.name.strip():
            issues.append(_issue("error", -1, "EDL name is empty."))

        if edl.sequence_resolution.width <= 0 or edl.sequence_resolution.height <= 0:
            issues.append(
                _issue(
                    "error",
                    -1,
                    f"Invalid sequence resolution: "
                    f"{edl.sequence_resolution.width}x{edl.sequence_resolution.height}.",
                )
            )

        if edl.sequence_frame_rate <= 0 or not math.isfinite(edl.sequence_frame_rate):
            issues.append(
                _issue("error", -1, f"Invalid sequence frame rate: {edl.sequence_frame_rate}.")
            )

        if not edl.entries:
            issues.append(_issue("warning", -1, "EDL contains no entries."))

    # ------------------------------------------------------------------
    # Per-entry validation
    # ------------------------------------------------------------------

    def _validate_entries(
        self,
        edl: EditDecisionList,
        issues: list[ValidationIssue],
    ) -> None:
        seen_indices: set[int] = set()

        for entry in edl.entries:
            idx = entry.index

            # Duplicate indices.
            if idx in seen_indices:
                issues.append(_issue("warning", idx, f"Duplicate entry index {idx}."))
            seen_indices.add(idx)

            # Source asset.
            if not entry.source_asset_id or not entry.source_asset_id.strip():
                issues.append(_issue("error", idx, "Missing source asset ID."))

            # Track target.
            self._validate_track_target(entry.track, idx, issues)

            # Source range.
            self._validate_time_range(
                entry.source_range, "source range", idx, edl.sequence_frame_rate, issues
            )

            # Timeline range.
            self._validate_time_range(
                entry.timeline_range, "timeline range", idx, edl.sequence_frame_rate, issues
            )

            # Frame-rate consistency between source in-point and sequence.
            if (
                entry.source_range.in_point.frame_rate != edl.sequence_frame_rate
                and entry.source_range.in_point.frame_rate > 0
            ):
                issues.append(
                    _issue(
                        "warning",
                        idx,
                        f"Source in-point frame rate ({entry.source_range.in_point.frame_rate}) "
                        f"differs from sequence frame rate ({edl.sequence_frame_rate}).",
                    )
                )

            # Timecode frame alignment.
            self._validate_frame_alignment(
                entry.timeline_range, "timeline range", idx, edl.sequence_frame_rate, issues
            )
            self._validate_frame_alignment(
                entry.source_range, "source range", idx, edl.sequence_frame_rate, issues
            )

            # Transition.
            if entry.transition is not None:
                self._validate_transition(entry.transition, idx, issues)

    def _validate_track_target(
        self,
        track: TrackTarget,
        entry_index: int,
        issues: list[ValidationIssue],
    ) -> None:
        valid_types = {"video", "audio"}
        if track.type not in valid_types:
            issues.append(
                _issue(
                    "error",
                    entry_index,
                    f'Invalid track type "{track.type}". Must be "video" or "audio".',
                )
            )
        if track.track_index < 0:
            issues.append(
                _issue(
                    "error",
                    entry_index,
                    f"Invalid track index {track.track_index}. Must be non-negative.",
                )
            )

    def _validate_time_range(
        self,
        time_range: TimeRange,
        label: str,
        entry_index: int,
        fps: float,
        issues: list[ValidationIssue],
    ) -> None:
        # Validate individual timecodes.
        for point_name, tc in [("in-point", time_range.in_point), ("out-point", time_range.out_point)]:
            err = self._validate_timecode(tc, fps)
            if err:
                issues.append(
                    _issue("error", entry_index, f"Invalid {label} {point_name}: {err}")
                )
                return

        # Out must be >= in.
        in_frames = _timecode_to_frames(time_range.in_point)
        out_frames = _timecode_to_frames(time_range.out_point)

        if out_frames < in_frames:
            issues.append(
                _issue(
                    "error",
                    entry_index,
                    f"{label} out-point ({time_range.out_point}) is before "
                    f"in-point ({time_range.in_point}).",
                )
            )

        if out_frames == in_frames:
            issues.append(
                _issue(
                    "warning",
                    entry_index,
                    f"{label} has zero duration (in == out at {time_range.in_point}).",
                )
            )

    def _validate_timecode(self, tc: Timecode, fps: float) -> str | None:
        """Return an error message if the timecode is invalid, else None."""
        if tc.hours < 0 or tc.minutes < 0 or tc.seconds < 0 or tc.frames < 0:
            return "Timecode components must be non-negative."
        if tc.minutes > 59:
            return f"Minutes value {tc.minutes} exceeds 59."
        if tc.seconds > 59:
            return f"Seconds value {tc.seconds} exceeds 59."
        max_frames = math.ceil(tc.frame_rate) - 1 if tc.frame_rate > 0 else 0
        if tc.frames > max_frames:
            return (
                f"Frames value {tc.frames} exceeds maximum {max_frames} "
                f"for {tc.frame_rate} fps."
            )
        return None

    def _validate_frame_alignment(
        self,
        time_range: TimeRange,
        label: str,
        entry_index: int,
        fps: float,
        issues: list[ValidationIssue],
    ) -> None:
        """Warn if timecodes are not aligned to frame boundaries."""
        for point_name, tc in [("in-point", time_range.in_point), ("out-point", time_range.out_point)]:
            total_seconds = tc.to_seconds()
            frame_number = total_seconds * fps
            rounded_frame = round(frame_number)
            if abs(frame_number - rounded_frame) > 1e-6:
                issues.append(
                    _issue(
                        "warning",
                        entry_index,
                        f"{label} {point_name} ({tc}) is not frame-aligned at {fps} fps.",
                    )
                )

    def _validate_transition(
        self,
        transition: object,
        entry_index: int,
        issues: list[ValidationIssue],
    ) -> None:
        """Validate a transition on an EDL entry."""
        from src.models import TransitionInfo

        if not isinstance(transition, TransitionInfo):
            return

        known_types = [
            "cross_dissolve",
            "dip_to_black",
            "dip_to_white",
            "wipe",
            "slide",
            "push",
            "fade",
            "cut",
        ]

        if not transition.type or not transition.type.strip():
            issues.append(_issue("error", entry_index, "Transition type is empty."))
        elif transition.type not in known_types:
            issues.append(
                _issue(
                    "warning",
                    entry_index,
                    f'Unknown transition type "{transition.type}". '
                    f"Known types: {', '.join(known_types)}.",
                )
            )

        if transition.duration_seconds <= 0:
            issues.append(
                _issue(
                    "error",
                    entry_index,
                    f"Transition duration must be positive, got {transition.duration_seconds}s.",
                )
            )

    # ------------------------------------------------------------------
    # Overlap detection
    # ------------------------------------------------------------------

    def _check_overlapping_clips(
        self,
        edl: EditDecisionList,
        issues: list[ValidationIssue],
    ) -> None:
        """Check for overlapping clips on the same track."""
        track_map: dict[str, list[tuple[int, int, int]]] = {}  # key -> [(idx, start, end)]

        for entry in edl.entries:
            key = f"{entry.track.type}:{entry.track.track_index}"
            start_frame = _timecode_to_frames(entry.timeline_range.in_point)
            end_frame = _timecode_to_frames(entry.timeline_range.out_point)

            if end_frame <= start_frame:
                continue  # Already reported.

            track_map.setdefault(key, []).append((entry.index, start_frame, end_frame))

        for track_key, spans in track_map.items():
            spans.sort(key=lambda s: s[1])

            for i in range(len(spans) - 1):
                current_idx, _current_start, current_end = spans[i]
                next_idx, next_start, _next_end = spans[i + 1]

                if current_end > next_start:
                    issues.append(
                        _issue(
                            "warning",
                            current_idx,
                            f"Clip at entry {current_idx} overlaps with clip at entry "
                            f"{next_idx} on track {track_key}. "
                            f"Current ends at frame {current_end}, "
                            f"next starts at frame {next_start}.",
                        )
                    )

    # ------------------------------------------------------------------
    # Gap detection
    # ------------------------------------------------------------------

    def _check_timeline_gaps(
        self,
        edl: EditDecisionList,
        issues: list[ValidationIssue],
    ) -> None:
        """Check for gaps between clips on the primary video track (V1)."""
        v1_spans: list[tuple[int, int, int]] = []

        for entry in edl.entries:
            if entry.track.type == "video" and entry.track.track_index == 0:
                start_frame = _timecode_to_frames(entry.timeline_range.in_point)
                end_frame = _timecode_to_frames(entry.timeline_range.out_point)
                if end_frame > start_frame:
                    v1_spans.append((entry.index, start_frame, end_frame))

        if not v1_spans:
            return

        v1_spans.sort(key=lambda s: s[1])

        for i in range(len(v1_spans) - 1):
            current_idx, _cs, current_end = v1_spans[i]
            next_idx, next_start, _ne = v1_spans[i + 1]

            gap_frames = next_start - current_end
            if gap_frames > 0:
                gap_seconds = gap_frames / edl.sequence_frame_rate if edl.sequence_frame_rate > 0 else 0
                issues.append(
                    _issue(
                        "warning",
                        current_idx,
                        f"Gap of {gap_frames} frames ({gap_seconds:.3f}s) between "
                        f"entry {current_idx} and entry {next_idx} on V1.",
                    )
                )

    # ------------------------------------------------------------------
    # Text overlay validation
    # ------------------------------------------------------------------

    def _validate_text_overlays(
        self,
        edl: EditDecisionList,
        issues: list[ValidationIssue],
    ) -> None:
        for i, overlay in enumerate(edl.text_overlays):
            if not overlay.text or not overlay.text.strip():
                issues.append(_issue("warning", -1, f"Text overlay {i} has empty text."))

            if overlay.duration_seconds <= 0:
                issues.append(
                    _issue(
                        "error",
                        -1,
                        f"Text overlay {i} has non-positive duration "
                        f"({overlay.duration_seconds}s).",
                    )
                )


# ------------------------------------------------------------------
# Module-level helpers
# ------------------------------------------------------------------


def _issue(severity: str, entry_index: int, message: str) -> ValidationIssue:
    return ValidationIssue(severity=severity, entry_index=entry_index, message=message)


def _timecode_to_frames(tc: Timecode) -> int:
    """Convert a timecode to an absolute frame count."""
    fps = math.ceil(tc.frame_rate) if tc.frame_rate > 0 else 1
    total_seconds = tc.hours * 3600 + tc.minutes * 60 + tc.seconds
    return total_seconds * fps + tc.frames
