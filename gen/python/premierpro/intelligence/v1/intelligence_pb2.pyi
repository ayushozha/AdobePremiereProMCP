from premierpro.common.v1 import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SegmentType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SEGMENT_TYPE_UNSPECIFIED: _ClassVar[SegmentType]
    SEGMENT_TYPE_DIALOGUE: _ClassVar[SegmentType]
    SEGMENT_TYPE_ACTION: _ClassVar[SegmentType]
    SEGMENT_TYPE_BROLL: _ClassVar[SegmentType]
    SEGMENT_TYPE_TRANSITION: _ClassVar[SegmentType]
    SEGMENT_TYPE_TITLE: _ClassVar[SegmentType]
    SEGMENT_TYPE_LOWER_THIRD: _ClassVar[SegmentType]
    SEGMENT_TYPE_VOICEOVER: _ClassVar[SegmentType]
    SEGMENT_TYPE_MUSIC: _ClassVar[SegmentType]
    SEGMENT_TYPE_SFX: _ClassVar[SegmentType]

class PacingPreset(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PACING_PRESET_UNSPECIFIED: _ClassVar[PacingPreset]
    PACING_PRESET_SLOW: _ClassVar[PacingPreset]
    PACING_PRESET_MODERATE: _ClassVar[PacingPreset]
    PACING_PRESET_FAST: _ClassVar[PacingPreset]
    PACING_PRESET_DYNAMIC: _ClassVar[PacingPreset]

class MatchStrategy(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    MATCH_STRATEGY_UNSPECIFIED: _ClassVar[MatchStrategy]
    MATCH_STRATEGY_KEYWORD: _ClassVar[MatchStrategy]
    MATCH_STRATEGY_EMBEDDING: _ClassVar[MatchStrategy]
    MATCH_STRATEGY_HYBRID: _ClassVar[MatchStrategy]
SEGMENT_TYPE_UNSPECIFIED: SegmentType
SEGMENT_TYPE_DIALOGUE: SegmentType
SEGMENT_TYPE_ACTION: SegmentType
SEGMENT_TYPE_BROLL: SegmentType
SEGMENT_TYPE_TRANSITION: SegmentType
SEGMENT_TYPE_TITLE: SegmentType
SEGMENT_TYPE_LOWER_THIRD: SegmentType
SEGMENT_TYPE_VOICEOVER: SegmentType
SEGMENT_TYPE_MUSIC: SegmentType
SEGMENT_TYPE_SFX: SegmentType
PACING_PRESET_UNSPECIFIED: PacingPreset
PACING_PRESET_SLOW: PacingPreset
PACING_PRESET_MODERATE: PacingPreset
PACING_PRESET_FAST: PacingPreset
PACING_PRESET_DYNAMIC: PacingPreset
MATCH_STRATEGY_UNSPECIFIED: MatchStrategy
MATCH_STRATEGY_KEYWORD: MatchStrategy
MATCH_STRATEGY_EMBEDDING: MatchStrategy
MATCH_STRATEGY_HYBRID: MatchStrategy

class ParseScriptRequest(_message.Message):
    __slots__ = ("text", "file_path", "format_hint")
    TEXT_FIELD_NUMBER: _ClassVar[int]
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    FORMAT_HINT_FIELD_NUMBER: _ClassVar[int]
    text: str
    file_path: str
    format_hint: str
    def __init__(self, text: _Optional[str] = ..., file_path: _Optional[str] = ..., format_hint: _Optional[str] = ...) -> None: ...

class ParseScriptResponse(_message.Message):
    __slots__ = ("segments", "metadata")
    SEGMENTS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    segments: _containers.RepeatedCompositeFieldContainer[ScriptSegment]
    metadata: ScriptMetadata
    def __init__(self, segments: _Optional[_Iterable[_Union[ScriptSegment, _Mapping]]] = ..., metadata: _Optional[_Union[ScriptMetadata, _Mapping]] = ...) -> None: ...

class ScriptSegment(_message.Message):
    __slots__ = ("index", "type", "content", "speaker", "scene_description", "visual_direction", "audio_direction", "estimated_duration_seconds", "asset_hints")
    INDEX_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    SPEAKER_FIELD_NUMBER: _ClassVar[int]
    SCENE_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    VISUAL_DIRECTION_FIELD_NUMBER: _ClassVar[int]
    AUDIO_DIRECTION_FIELD_NUMBER: _ClassVar[int]
    ESTIMATED_DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    ASSET_HINTS_FIELD_NUMBER: _ClassVar[int]
    index: int
    type: SegmentType
    content: str
    speaker: str
    scene_description: str
    visual_direction: str
    audio_direction: str
    estimated_duration_seconds: float
    asset_hints: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, index: _Optional[int] = ..., type: _Optional[_Union[SegmentType, str]] = ..., content: _Optional[str] = ..., speaker: _Optional[str] = ..., scene_description: _Optional[str] = ..., visual_direction: _Optional[str] = ..., audio_direction: _Optional[str] = ..., estimated_duration_seconds: _Optional[float] = ..., asset_hints: _Optional[_Iterable[str]] = ...) -> None: ...

class ScriptMetadata(_message.Message):
    __slots__ = ("title", "format", "estimated_total_duration_seconds", "segment_count")
    TITLE_FIELD_NUMBER: _ClassVar[int]
    FORMAT_FIELD_NUMBER: _ClassVar[int]
    ESTIMATED_TOTAL_DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    SEGMENT_COUNT_FIELD_NUMBER: _ClassVar[int]
    title: str
    format: str
    estimated_total_duration_seconds: float
    segment_count: int
    def __init__(self, title: _Optional[str] = ..., format: _Optional[str] = ..., estimated_total_duration_seconds: _Optional[float] = ..., segment_count: _Optional[int] = ...) -> None: ...

class GenerateEDLRequest(_message.Message):
    __slots__ = ("segments", "available_assets", "matches", "settings")
    SEGMENTS_FIELD_NUMBER: _ClassVar[int]
    AVAILABLE_ASSETS_FIELD_NUMBER: _ClassVar[int]
    MATCHES_FIELD_NUMBER: _ClassVar[int]
    SETTINGS_FIELD_NUMBER: _ClassVar[int]
    segments: _containers.RepeatedCompositeFieldContainer[ScriptSegment]
    available_assets: _containers.RepeatedCompositeFieldContainer[_common_pb2.Asset]
    matches: _containers.RepeatedCompositeFieldContainer[AssetMatch]
    settings: EDLSettings
    def __init__(self, segments: _Optional[_Iterable[_Union[ScriptSegment, _Mapping]]] = ..., available_assets: _Optional[_Iterable[_Union[_common_pb2.Asset, _Mapping]]] = ..., matches: _Optional[_Iterable[_Union[AssetMatch, _Mapping]]] = ..., settings: _Optional[_Union[EDLSettings, _Mapping]] = ...) -> None: ...

class EDLSettings(_message.Message):
    __slots__ = ("resolution", "frame_rate", "default_transition", "default_transition_duration", "pacing")
    RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    FRAME_RATE_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_TRANSITION_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_TRANSITION_DURATION_FIELD_NUMBER: _ClassVar[int]
    PACING_FIELD_NUMBER: _ClassVar[int]
    resolution: _common_pb2.Resolution
    frame_rate: float
    default_transition: str
    default_transition_duration: float
    pacing: PacingPreset
    def __init__(self, resolution: _Optional[_Union[_common_pb2.Resolution, _Mapping]] = ..., frame_rate: _Optional[float] = ..., default_transition: _Optional[str] = ..., default_transition_duration: _Optional[float] = ..., pacing: _Optional[_Union[PacingPreset, str]] = ...) -> None: ...

class GenerateEDLResponse(_message.Message):
    __slots__ = ("edl", "warnings")
    EDL_FIELD_NUMBER: _ClassVar[int]
    WARNINGS_FIELD_NUMBER: _ClassVar[int]
    edl: _common_pb2.EditDecisionList
    warnings: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, edl: _Optional[_Union[_common_pb2.EditDecisionList, _Mapping]] = ..., warnings: _Optional[_Iterable[str]] = ...) -> None: ...

class MatchAssetsRequest(_message.Message):
    __slots__ = ("segments", "available_assets", "strategy")
    SEGMENTS_FIELD_NUMBER: _ClassVar[int]
    AVAILABLE_ASSETS_FIELD_NUMBER: _ClassVar[int]
    STRATEGY_FIELD_NUMBER: _ClassVar[int]
    segments: _containers.RepeatedCompositeFieldContainer[ScriptSegment]
    available_assets: _containers.RepeatedCompositeFieldContainer[_common_pb2.Asset]
    strategy: MatchStrategy
    def __init__(self, segments: _Optional[_Iterable[_Union[ScriptSegment, _Mapping]]] = ..., available_assets: _Optional[_Iterable[_Union[_common_pb2.Asset, _Mapping]]] = ..., strategy: _Optional[_Union[MatchStrategy, str]] = ...) -> None: ...

class MatchAssetsResponse(_message.Message):
    __slots__ = ("matches", "unmatched")
    MATCHES_FIELD_NUMBER: _ClassVar[int]
    UNMATCHED_FIELD_NUMBER: _ClassVar[int]
    matches: _containers.RepeatedCompositeFieldContainer[AssetMatch]
    unmatched: _containers.RepeatedCompositeFieldContainer[UnmatchedSegment]
    def __init__(self, matches: _Optional[_Iterable[_Union[AssetMatch, _Mapping]]] = ..., unmatched: _Optional[_Iterable[_Union[UnmatchedSegment, _Mapping]]] = ...) -> None: ...

class AssetMatch(_message.Message):
    __slots__ = ("segment_index", "asset_id", "confidence", "reasoning", "suggested_range")
    SEGMENT_INDEX_FIELD_NUMBER: _ClassVar[int]
    ASSET_ID_FIELD_NUMBER: _ClassVar[int]
    CONFIDENCE_FIELD_NUMBER: _ClassVar[int]
    REASONING_FIELD_NUMBER: _ClassVar[int]
    SUGGESTED_RANGE_FIELD_NUMBER: _ClassVar[int]
    segment_index: int
    asset_id: str
    confidence: float
    reasoning: str
    suggested_range: _common_pb2.TimeRange
    def __init__(self, segment_index: _Optional[int] = ..., asset_id: _Optional[str] = ..., confidence: _Optional[float] = ..., reasoning: _Optional[str] = ..., suggested_range: _Optional[_Union[_common_pb2.TimeRange, _Mapping]] = ...) -> None: ...

class UnmatchedSegment(_message.Message):
    __slots__ = ("segment_index", "reason", "suggestions")
    SEGMENT_INDEX_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    SUGGESTIONS_FIELD_NUMBER: _ClassVar[int]
    segment_index: int
    reason: str
    suggestions: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, segment_index: _Optional[int] = ..., reason: _Optional[str] = ..., suggestions: _Optional[_Iterable[str]] = ...) -> None: ...

class AnalyzePacingRequest(_message.Message):
    __slots__ = ("edl", "target_mood")
    EDL_FIELD_NUMBER: _ClassVar[int]
    TARGET_MOOD_FIELD_NUMBER: _ClassVar[int]
    edl: _common_pb2.EditDecisionList
    target_mood: str
    def __init__(self, edl: _Optional[_Union[_common_pb2.EditDecisionList, _Mapping]] = ..., target_mood: _Optional[str] = ...) -> None: ...

class AnalyzePacingResponse(_message.Message):
    __slots__ = ("adjustments", "current_avg_clip_duration", "suggested_avg_clip_duration")
    ADJUSTMENTS_FIELD_NUMBER: _ClassVar[int]
    CURRENT_AVG_CLIP_DURATION_FIELD_NUMBER: _ClassVar[int]
    SUGGESTED_AVG_CLIP_DURATION_FIELD_NUMBER: _ClassVar[int]
    adjustments: _containers.RepeatedCompositeFieldContainer[PacingAdjustment]
    current_avg_clip_duration: float
    suggested_avg_clip_duration: float
    def __init__(self, adjustments: _Optional[_Iterable[_Union[PacingAdjustment, _Mapping]]] = ..., current_avg_clip_duration: _Optional[float] = ..., suggested_avg_clip_duration: _Optional[float] = ...) -> None: ...

class PacingAdjustment(_message.Message):
    __slots__ = ("edl_entry_index", "current_duration", "suggested_duration", "reason")
    EDL_ENTRY_INDEX_FIELD_NUMBER: _ClassVar[int]
    CURRENT_DURATION_FIELD_NUMBER: _ClassVar[int]
    SUGGESTED_DURATION_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    edl_entry_index: int
    current_duration: float
    suggested_duration: float
    reason: str
    def __init__(self, edl_entry_index: _Optional[int] = ..., current_duration: _Optional[float] = ..., suggested_duration: _Optional[float] = ..., reason: _Optional[str] = ...) -> None: ...
