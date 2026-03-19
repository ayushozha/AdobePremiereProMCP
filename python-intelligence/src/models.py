"""Pydantic models mirroring the protobuf definitions.

These are the internal Python representations used throughout the intelligence
service. They map 1:1 to the proto messages in ``premierpro/intelligence/v1``
and ``premierpro/common/v1`` so that conversion is straightforward.
"""

from __future__ import annotations

from enum import Enum

from pydantic import BaseModel, Field


# ── Enums ────────────────────────────────────────────────────────────────────────


class SegmentType(str, Enum):
    """Type of script segment, mirroring the proto ``SegmentType`` enum."""

    UNSPECIFIED = "unspecified"
    DIALOGUE = "dialogue"
    ACTION = "action"
    BROLL = "broll"
    TRANSITION = "transition"
    TITLE = "title"
    LOWER_THIRD = "lower_third"
    VOICEOVER = "voiceover"
    MUSIC = "music"
    SFX = "sfx"


class ScriptFormat(str, Enum):
    """Detected or requested script format."""

    AUTO = "auto"
    SCREENPLAY = "screenplay"
    YOUTUBE = "youtube"
    PODCAST = "podcast"
    NARRATION = "narration"


class MatchStrategy(str, Enum):
    """Strategy for matching assets to script segments."""

    UNSPECIFIED = "unspecified"
    KEYWORD = "keyword"
    EMBEDDING = "embedding"
    HYBRID = "hybrid"


class PacingPreset(str, Enum):
    """Preset pacing profiles for EDL generation."""

    UNSPECIFIED = "unspecified"
    SLOW = "slow"
    MODERATE = "moderate"
    FAST = "fast"
    DYNAMIC = "dynamic"


class AssetType(str, Enum):
    """Type of media asset."""

    UNSPECIFIED = "unspecified"
    VIDEO = "video"
    AUDIO = "audio"
    IMAGE = "image"
    GRAPHICS = "graphics"


class TrackType(str, Enum):
    """Target track type for clip placement."""

    UNSPECIFIED = "unspecified"
    VIDEO = "video"
    AUDIO = "audio"


# ── Common Models (from common.proto) ────────────────────────────────────────────


class Timecode(BaseModel):
    """Timecode in HH:MM:SS:FF format."""

    hours: int = 0
    minutes: int = 0
    seconds: int = 0
    frames: int = 0
    frame_rate: float = 24.0

    def to_total_frames(self) -> int:
        """Convert to an absolute frame count."""
        fr = int(self.frame_rate)
        return ((self.hours * 3600 + self.minutes * 60 + self.seconds) * fr) + self.frames

    def to_seconds(self) -> float:
        """Convert to fractional seconds."""
        return self.hours * 3600.0 + self.minutes * 60.0 + self.seconds + (
            self.frames / self.frame_rate if self.frame_rate > 0 else 0.0
        )


class TimeRange(BaseModel):
    """A time range defined by in and out points."""

    in_point: Timecode = Field(default_factory=Timecode)
    out_point: Timecode = Field(default_factory=Timecode)

    def duration_seconds(self) -> float:
        """Return the duration of this range in seconds."""
        return self.out_point.to_seconds() - self.in_point.to_seconds()


class Resolution(BaseModel):
    """Video resolution."""

    width: int = 1920
    height: int = 1080


class VideoInfo(BaseModel):
    """Video stream metadata."""

    codec: str = ""
    resolution: Resolution = Field(default_factory=Resolution)
    frame_rate: float = 0.0
    bitrate_bps: int = 0
    pixel_format: str = ""
    duration_seconds: float = 0.0


class AudioInfo(BaseModel):
    """Audio stream metadata."""

    codec: str = ""
    sample_rate: int = 0
    channels: int = 0
    bitrate_bps: int = 0
    duration_seconds: float = 0.0


class TransitionInfo(BaseModel):
    """Transition between clips."""

    type: str = ""
    duration_seconds: float = 0.0
    alignment: str = ""


class EffectInfo(BaseModel):
    """Effect applied to a clip."""

    name: str = ""
    parameters: dict[str, str] = Field(default_factory=dict)


class TrackTarget(BaseModel):
    """Target track for clip placement."""

    type: TrackType = TrackType.UNSPECIFIED
    track_index: int = 0


class Position(BaseModel):
    """Normalised screen position (0.0--1.0)."""

    x: float = 0.5
    y: float = 0.5


class TextStyle(BaseModel):
    """Visual style for text overlays."""

    font_family: str = "Arial"
    font_size: float = 48.0
    color_hex: str = "#FFFFFF"
    alignment: str = "center"
    background_color_hex: str = "#000000"
    background_opacity: float = 0.0
    position: Position = Field(default_factory=Position)


class TextOverlay(BaseModel):
    """A text element placed on the timeline."""

    text: str = ""
    style: TextStyle = Field(default_factory=TextStyle)
    track: TrackTarget = Field(default_factory=TrackTarget)
    position: Timecode = Field(default_factory=Timecode)
    duration_seconds: float = 0.0


# ── Script Models ────────────────────────────────────────────────────────────────


class ScriptSegment(BaseModel):
    """A single parsed segment of a script.

    Mirrors the proto ``ScriptSegment`` message with Pythonic field names.
    """

    index: int = Field(description="Zero-based index of this segment in the script")
    type: SegmentType = Field(default=SegmentType.UNSPECIFIED, description="Segment type")
    content: str = Field(default="", description="Raw text content of the segment")
    speaker: str = Field(default="", description="Speaker / character name (if dialogue or VO)")
    scene_description: str = Field(
        default="",
        description="Scene heading or location description",
    )
    visual_direction: str = Field(
        default="",
        description="Visual notes (camera direction, B-roll description)",
    )
    audio_direction: str = Field(
        default="",
        description="Audio notes (music cue, SFX description)",
    )
    estimated_duration_seconds: float = Field(
        default=0.0,
        ge=0.0,
        description="Estimated duration in seconds",
    )
    asset_hints: list[str] = Field(
        default_factory=list,
        description="Keywords extracted for asset matching",
    )


class ScriptMetadata(BaseModel):
    """High-level metadata about a parsed script."""

    title: str = Field(default="", description="Detected or inferred title")
    format: ScriptFormat = Field(
        default=ScriptFormat.AUTO,
        description="Script format that was detected or specified",
    )
    estimated_total_duration_seconds: float = Field(
        default=0.0,
        ge=0.0,
        description="Total estimated duration across all segments",
    )
    segment_count: int = Field(
        default=0,
        ge=0,
        description="Number of segments in the parsed script",
    )


class ParsedScript(BaseModel):
    """Complete result of parsing a script.

    Returned by ``ScriptParser.parse`` and ``ScriptParser.parse_file``.
    """

    segments: list[ScriptSegment] = Field(
        default_factory=list,
        description="Ordered list of parsed segments",
    )
    metadata: ScriptMetadata = Field(
        default_factory=ScriptMetadata,
        description="Metadata about the parsed script",
    )


# ── Asset Models ─────────────────────────────────────────────────────────────────


class AssetInfo(BaseModel):
    """A media asset with full metadata (maps to ``common.Asset``)."""

    id: str = ""
    file_path: str = ""
    file_name: str = ""
    file_size_bytes: int = 0
    mime_type: str = ""
    asset_type: AssetType = AssetType.UNSPECIFIED
    video: VideoInfo | None = None
    audio: AudioInfo | None = None
    metadata: dict[str, str] = Field(default_factory=dict)
    fingerprint: str = ""

    @property
    def duration_seconds(self) -> float:
        """Best-effort duration from whichever stream is available."""
        if self.video and self.video.duration_seconds > 0:
            return self.video.duration_seconds
        if self.audio and self.audio.duration_seconds > 0:
            return self.audio.duration_seconds
        return 0.0


class AssetMatch(BaseModel):
    """A match between a script segment and a media asset."""

    segment_index: int = 0
    asset_id: str = ""
    confidence: float = Field(default=0.0, ge=0.0, le=1.0)
    reasoning: str = ""
    suggested_range: TimeRange | None = None


class UnmatchedSegment(BaseModel):
    """A segment that could not be matched to any asset."""

    segment_index: int = 0
    reason: str = ""
    suggestions: list[str] = Field(default_factory=list)


class MatchResult(BaseModel):
    """Complete result of matching assets to segments."""

    matches: list[AssetMatch] = Field(default_factory=list)
    unmatched: list[UnmatchedSegment] = Field(default_factory=list)


# ── EDL Models ───────────────────────────────────────────────────────────────────


class EDLEntry(BaseModel):
    """A single entry in an Edit Decision List."""

    index: int = 0
    source_asset_id: str = ""
    source_range: TimeRange = Field(default_factory=TimeRange)
    timeline_range: TimeRange = Field(default_factory=TimeRange)
    track: TrackTarget = Field(default_factory=TrackTarget)
    transition: TransitionInfo | None = None
    effects: list[EffectInfo] = Field(default_factory=list)
    notes: str = ""


class EDLSettings(BaseModel):
    """Settings controlling EDL generation."""

    resolution: Resolution = Field(default_factory=Resolution)
    frame_rate: float = 24.0
    default_transition: str = "cut"
    default_transition_duration: float = 0.0
    pacing: PacingPreset = PacingPreset.MODERATE


class EditDecisionList(BaseModel):
    """A full Edit Decision List."""

    id: str = ""
    name: str = ""
    sequence_resolution: Resolution = Field(default_factory=Resolution)
    sequence_frame_rate: float = 24.0
    entries: list[EDLEntry] = Field(default_factory=list)
    text_overlays: list[TextOverlay] = Field(default_factory=list)
    audio_levels: dict[int, float] = Field(default_factory=dict)


# ── Pacing Models ────────────────────────────────────────────────────────────────


class PacingAdjustment(BaseModel):
    """A single pacing adjustment for an EDL entry."""

    edl_entry_index: int = 0
    current_duration: float = 0.0
    suggested_duration: float = 0.0
    reason: str = ""


class PacingResult(BaseModel):
    """Complete result of pacing analysis."""

    adjustments: list[PacingAdjustment] = Field(default_factory=list)
    current_avg_clip_duration: float = 0.0
    suggested_avg_clip_duration: float = 0.0


# ── Validation ───────────────────────────────────────────────────────────────────


class ValidationIssue(BaseModel):
    """A single validation error or warning."""

    severity: str = "error"  # "error" | "warning"
    entry_index: int = -1  # -1 for global issues
    message: str = ""


class ValidationResult(BaseModel):
    """Result of validating an EDL."""

    valid: bool = True
    errors: list[ValidationIssue] = Field(default_factory=list)
    warnings: list[ValidationIssue] = Field(default_factory=list)
