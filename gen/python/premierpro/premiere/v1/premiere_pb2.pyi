from premierpro.common.v1 import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ExportPreset(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    EXPORT_PRESET_UNSPECIFIED: _ClassVar[ExportPreset]
    EXPORT_PRESET_H264_1080P: _ClassVar[ExportPreset]
    EXPORT_PRESET_H264_4K: _ClassVar[ExportPreset]
    EXPORT_PRESET_PRORES_422: _ClassVar[ExportPreset]
    EXPORT_PRESET_PRORES_4444: _ClassVar[ExportPreset]
    EXPORT_PRESET_DNX_HR: _ClassVar[ExportPreset]
    EXPORT_PRESET_CUSTOM: _ClassVar[ExportPreset]
EXPORT_PRESET_UNSPECIFIED: ExportPreset
EXPORT_PRESET_H264_1080P: ExportPreset
EXPORT_PRESET_H264_4K: ExportPreset
EXPORT_PRESET_PRORES_422: ExportPreset
EXPORT_PRESET_PRORES_4444: ExportPreset
EXPORT_PRESET_DNX_HR: ExportPreset
EXPORT_PRESET_CUSTOM: ExportPreset

class GetProjectStateRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetProjectStateResponse(_message.Message):
    __slots__ = ("project_name", "project_path", "sequences", "bin_count", "is_saved")
    PROJECT_NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_PATH_FIELD_NUMBER: _ClassVar[int]
    SEQUENCES_FIELD_NUMBER: _ClassVar[int]
    BIN_COUNT_FIELD_NUMBER: _ClassVar[int]
    IS_SAVED_FIELD_NUMBER: _ClassVar[int]
    project_name: str
    project_path: str
    sequences: _containers.RepeatedCompositeFieldContainer[SequenceInfo]
    bin_count: int
    is_saved: bool
    def __init__(self, project_name: _Optional[str] = ..., project_path: _Optional[str] = ..., sequences: _Optional[_Iterable[_Union[SequenceInfo, _Mapping]]] = ..., bin_count: _Optional[int] = ..., is_saved: _Optional[bool] = ...) -> None: ...

class SequenceInfo(_message.Message):
    __slots__ = ("id", "name", "resolution", "frame_rate", "duration_seconds", "video_track_count", "audio_track_count")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    FRAME_RATE_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    VIDEO_TRACK_COUNT_FIELD_NUMBER: _ClassVar[int]
    AUDIO_TRACK_COUNT_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    resolution: _common_pb2.Resolution
    frame_rate: float
    duration_seconds: float
    video_track_count: int
    audio_track_count: int
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., resolution: _Optional[_Union[_common_pb2.Resolution, _Mapping]] = ..., frame_rate: _Optional[float] = ..., duration_seconds: _Optional[float] = ..., video_track_count: _Optional[int] = ..., audio_track_count: _Optional[int] = ...) -> None: ...

class CreateSequenceRequest(_message.Message):
    __slots__ = ("name", "resolution", "frame_rate", "video_tracks", "audio_tracks")
    NAME_FIELD_NUMBER: _ClassVar[int]
    RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    FRAME_RATE_FIELD_NUMBER: _ClassVar[int]
    VIDEO_TRACKS_FIELD_NUMBER: _ClassVar[int]
    AUDIO_TRACKS_FIELD_NUMBER: _ClassVar[int]
    name: str
    resolution: _common_pb2.Resolution
    frame_rate: float
    video_tracks: int
    audio_tracks: int
    def __init__(self, name: _Optional[str] = ..., resolution: _Optional[_Union[_common_pb2.Resolution, _Mapping]] = ..., frame_rate: _Optional[float] = ..., video_tracks: _Optional[int] = ..., audio_tracks: _Optional[int] = ...) -> None: ...

class CreateSequenceResponse(_message.Message):
    __slots__ = ("sequence_id", "name")
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    name: str
    def __init__(self, sequence_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class GetTimelineStateRequest(_message.Message):
    __slots__ = ("sequence_id",)
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    def __init__(self, sequence_id: _Optional[str] = ...) -> None: ...

class GetTimelineStateResponse(_message.Message):
    __slots__ = ("sequence_id", "video_tracks", "audio_tracks", "total_duration_seconds")
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    VIDEO_TRACKS_FIELD_NUMBER: _ClassVar[int]
    AUDIO_TRACKS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    video_tracks: _containers.RepeatedCompositeFieldContainer[TimelineTrack]
    audio_tracks: _containers.RepeatedCompositeFieldContainer[TimelineTrack]
    total_duration_seconds: float
    def __init__(self, sequence_id: _Optional[str] = ..., video_tracks: _Optional[_Iterable[_Union[TimelineTrack, _Mapping]]] = ..., audio_tracks: _Optional[_Iterable[_Union[TimelineTrack, _Mapping]]] = ..., total_duration_seconds: _Optional[float] = ...) -> None: ...

class TimelineTrack(_message.Message):
    __slots__ = ("index", "type", "clips", "is_muted", "is_locked")
    INDEX_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    CLIPS_FIELD_NUMBER: _ClassVar[int]
    IS_MUTED_FIELD_NUMBER: _ClassVar[int]
    IS_LOCKED_FIELD_NUMBER: _ClassVar[int]
    index: int
    type: _common_pb2.TrackType
    clips: _containers.RepeatedCompositeFieldContainer[TimelineClip]
    is_muted: bool
    is_locked: bool
    def __init__(self, index: _Optional[int] = ..., type: _Optional[_Union[_common_pb2.TrackType, str]] = ..., clips: _Optional[_Iterable[_Union[TimelineClip, _Mapping]]] = ..., is_muted: _Optional[bool] = ..., is_locked: _Optional[bool] = ...) -> None: ...

class TimelineClip(_message.Message):
    __slots__ = ("clip_id", "source_path", "source_range", "timeline_range", "speed")
    CLIP_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_PATH_FIELD_NUMBER: _ClassVar[int]
    SOURCE_RANGE_FIELD_NUMBER: _ClassVar[int]
    TIMELINE_RANGE_FIELD_NUMBER: _ClassVar[int]
    SPEED_FIELD_NUMBER: _ClassVar[int]
    clip_id: str
    source_path: str
    source_range: _common_pb2.TimeRange
    timeline_range: _common_pb2.TimeRange
    speed: float
    def __init__(self, clip_id: _Optional[str] = ..., source_path: _Optional[str] = ..., source_range: _Optional[_Union[_common_pb2.TimeRange, _Mapping]] = ..., timeline_range: _Optional[_Union[_common_pb2.TimeRange, _Mapping]] = ..., speed: _Optional[float] = ...) -> None: ...

class ImportMediaRequest(_message.Message):
    __slots__ = ("file_path", "target_bin")
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    TARGET_BIN_FIELD_NUMBER: _ClassVar[int]
    file_path: str
    target_bin: str
    def __init__(self, file_path: _Optional[str] = ..., target_bin: _Optional[str] = ...) -> None: ...

class ImportMediaResponse(_message.Message):
    __slots__ = ("project_item_id", "name")
    PROJECT_ITEM_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    project_item_id: str
    name: str
    def __init__(self, project_item_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class PlaceClipRequest(_message.Message):
    __slots__ = ("source_path", "track", "position", "source_range", "speed")
    SOURCE_PATH_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    SOURCE_RANGE_FIELD_NUMBER: _ClassVar[int]
    SPEED_FIELD_NUMBER: _ClassVar[int]
    source_path: str
    track: _common_pb2.TrackTarget
    position: _common_pb2.Timecode
    source_range: _common_pb2.TimeRange
    speed: float
    def __init__(self, source_path: _Optional[str] = ..., track: _Optional[_Union[_common_pb2.TrackTarget, _Mapping]] = ..., position: _Optional[_Union[_common_pb2.Timecode, _Mapping]] = ..., source_range: _Optional[_Union[_common_pb2.TimeRange, _Mapping]] = ..., speed: _Optional[float] = ...) -> None: ...

class PlaceClipResponse(_message.Message):
    __slots__ = ("clip_id",)
    CLIP_ID_FIELD_NUMBER: _ClassVar[int]
    clip_id: str
    def __init__(self, clip_id: _Optional[str] = ...) -> None: ...

class RemoveClipRequest(_message.Message):
    __slots__ = ("clip_id", "sequence_id")
    CLIP_ID_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    clip_id: str
    sequence_id: str
    def __init__(self, clip_id: _Optional[str] = ..., sequence_id: _Optional[str] = ...) -> None: ...

class RemoveClipResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class AddTransitionRequest(_message.Message):
    __slots__ = ("sequence_id", "track", "position", "transition_type", "duration_seconds")
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    TRANSITION_TYPE_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    track: _common_pb2.TrackTarget
    position: _common_pb2.Timecode
    transition_type: str
    duration_seconds: float
    def __init__(self, sequence_id: _Optional[str] = ..., track: _Optional[_Union[_common_pb2.TrackTarget, _Mapping]] = ..., position: _Optional[_Union[_common_pb2.Timecode, _Mapping]] = ..., transition_type: _Optional[str] = ..., duration_seconds: _Optional[float] = ...) -> None: ...

class AddTransitionResponse(_message.Message):
    __slots__ = ("transition_id",)
    TRANSITION_ID_FIELD_NUMBER: _ClassVar[int]
    transition_id: str
    def __init__(self, transition_id: _Optional[str] = ...) -> None: ...

class AddTextRequest(_message.Message):
    __slots__ = ("sequence_id", "text", "style", "track", "position", "duration_seconds")
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    STYLE_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    text: str
    style: _common_pb2.TextStyle
    track: _common_pb2.TrackTarget
    position: _common_pb2.Timecode
    duration_seconds: float
    def __init__(self, sequence_id: _Optional[str] = ..., text: _Optional[str] = ..., style: _Optional[_Union[_common_pb2.TextStyle, _Mapping]] = ..., track: _Optional[_Union[_common_pb2.TrackTarget, _Mapping]] = ..., position: _Optional[_Union[_common_pb2.Timecode, _Mapping]] = ..., duration_seconds: _Optional[float] = ...) -> None: ...

class AddTextResponse(_message.Message):
    __slots__ = ("clip_id",)
    CLIP_ID_FIELD_NUMBER: _ClassVar[int]
    clip_id: str
    def __init__(self, clip_id: _Optional[str] = ...) -> None: ...

class ApplyEffectRequest(_message.Message):
    __slots__ = ("clip_id", "sequence_id", "effect")
    CLIP_ID_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    EFFECT_FIELD_NUMBER: _ClassVar[int]
    clip_id: str
    sequence_id: str
    effect: _common_pb2.EffectInfo
    def __init__(self, clip_id: _Optional[str] = ..., sequence_id: _Optional[str] = ..., effect: _Optional[_Union[_common_pb2.EffectInfo, _Mapping]] = ...) -> None: ...

class ApplyEffectResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SetAudioLevelRequest(_message.Message):
    __slots__ = ("clip_id", "sequence_id", "level_db")
    CLIP_ID_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    LEVEL_DB_FIELD_NUMBER: _ClassVar[int]
    clip_id: str
    sequence_id: str
    level_db: float
    def __init__(self, clip_id: _Optional[str] = ..., sequence_id: _Optional[str] = ..., level_db: _Optional[float] = ...) -> None: ...

class SetAudioLevelResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ExportSequenceRequest(_message.Message):
    __slots__ = ("sequence_id", "output_path", "preset")
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    PRESET_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    output_path: str
    preset: ExportPreset
    def __init__(self, sequence_id: _Optional[str] = ..., output_path: _Optional[str] = ..., preset: _Optional[_Union[ExportPreset, str]] = ...) -> None: ...

class ExportSequenceResponse(_message.Message):
    __slots__ = ("job_id", "status", "output_path")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    status: _common_pb2.OperationStatus
    output_path: str
    def __init__(self, job_id: _Optional[str] = ..., status: _Optional[_Union[_common_pb2.OperationStatus, str]] = ..., output_path: _Optional[str] = ...) -> None: ...

class ExecuteEDLRequest(_message.Message):
    __slots__ = ("edl", "auto_import", "auto_create_sequence")
    EDL_FIELD_NUMBER: _ClassVar[int]
    AUTO_IMPORT_FIELD_NUMBER: _ClassVar[int]
    AUTO_CREATE_SEQUENCE_FIELD_NUMBER: _ClassVar[int]
    edl: _common_pb2.EditDecisionList
    auto_import: bool
    auto_create_sequence: bool
    def __init__(self, edl: _Optional[_Union[_common_pb2.EditDecisionList, _Mapping]] = ..., auto_import: _Optional[bool] = ..., auto_create_sequence: _Optional[bool] = ...) -> None: ...

class ExecuteEDLResponse(_message.Message):
    __slots__ = ("sequence_id", "status", "clips_placed", "transitions_added", "errors", "warnings")
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CLIPS_PLACED_FIELD_NUMBER: _ClassVar[int]
    TRANSITIONS_ADDED_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    WARNINGS_FIELD_NUMBER: _ClassVar[int]
    sequence_id: str
    status: _common_pb2.OperationStatus
    clips_placed: int
    transitions_added: int
    errors: _containers.RepeatedScalarFieldContainer[str]
    warnings: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, sequence_id: _Optional[str] = ..., status: _Optional[_Union[_common_pb2.OperationStatus, str]] = ..., clips_placed: _Optional[int] = ..., transitions_added: _Optional[int] = ..., errors: _Optional[_Iterable[str]] = ..., warnings: _Optional[_Iterable[str]] = ...) -> None: ...

class EvalCommandRequest(_message.Message):
    __slots__ = ("function_name", "args_json")
    FUNCTION_NAME_FIELD_NUMBER: _ClassVar[int]
    ARGS_JSON_FIELD_NUMBER: _ClassVar[int]
    function_name: str
    args_json: str
    def __init__(self, function_name: _Optional[str] = ..., args_json: _Optional[str] = ...) -> None: ...

class EvalCommandResponse(_message.Message):
    __slots__ = ("result_json", "is_error", "error_message")
    RESULT_JSON_FIELD_NUMBER: _ClassVar[int]
    IS_ERROR_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    result_json: str
    is_error: bool
    error_message: str
    def __init__(self, result_json: _Optional[str] = ..., is_error: _Optional[bool] = ..., error_message: _Optional[str] = ...) -> None: ...

class PingRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class PingResponse(_message.Message):
    __slots__ = ("premiere_running", "premiere_version", "project_open", "bridge_mode")
    PREMIERE_RUNNING_FIELD_NUMBER: _ClassVar[int]
    PREMIERE_VERSION_FIELD_NUMBER: _ClassVar[int]
    PROJECT_OPEN_FIELD_NUMBER: _ClassVar[int]
    BRIDGE_MODE_FIELD_NUMBER: _ClassVar[int]
    premiere_running: bool
    premiere_version: str
    project_open: bool
    bridge_mode: str
    def __init__(self, premiere_running: _Optional[bool] = ..., premiere_version: _Optional[str] = ..., project_open: _Optional[bool] = ..., bridge_mode: _Optional[str] = ...) -> None: ...
