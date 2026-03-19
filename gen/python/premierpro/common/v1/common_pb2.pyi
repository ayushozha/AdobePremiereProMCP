from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AssetType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ASSET_TYPE_UNSPECIFIED: _ClassVar[AssetType]
    ASSET_TYPE_VIDEO: _ClassVar[AssetType]
    ASSET_TYPE_AUDIO: _ClassVar[AssetType]
    ASSET_TYPE_IMAGE: _ClassVar[AssetType]
    ASSET_TYPE_GRAPHICS: _ClassVar[AssetType]

class TrackType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRACK_TYPE_UNSPECIFIED: _ClassVar[TrackType]
    TRACK_TYPE_VIDEO: _ClassVar[TrackType]
    TRACK_TYPE_AUDIO: _ClassVar[TrackType]

class OperationStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    OPERATION_STATUS_UNSPECIFIED: _ClassVar[OperationStatus]
    OPERATION_STATUS_PENDING: _ClassVar[OperationStatus]
    OPERATION_STATUS_RUNNING: _ClassVar[OperationStatus]
    OPERATION_STATUS_COMPLETED: _ClassVar[OperationStatus]
    OPERATION_STATUS_FAILED: _ClassVar[OperationStatus]
ASSET_TYPE_UNSPECIFIED: AssetType
ASSET_TYPE_VIDEO: AssetType
ASSET_TYPE_AUDIO: AssetType
ASSET_TYPE_IMAGE: AssetType
ASSET_TYPE_GRAPHICS: AssetType
TRACK_TYPE_UNSPECIFIED: TrackType
TRACK_TYPE_VIDEO: TrackType
TRACK_TYPE_AUDIO: TrackType
OPERATION_STATUS_UNSPECIFIED: OperationStatus
OPERATION_STATUS_PENDING: OperationStatus
OPERATION_STATUS_RUNNING: OperationStatus
OPERATION_STATUS_COMPLETED: OperationStatus
OPERATION_STATUS_FAILED: OperationStatus

class Timecode(_message.Message):
    __slots__ = ("hours", "minutes", "seconds", "frames", "frame_rate")
    HOURS_FIELD_NUMBER: _ClassVar[int]
    MINUTES_FIELD_NUMBER: _ClassVar[int]
    SECONDS_FIELD_NUMBER: _ClassVar[int]
    FRAMES_FIELD_NUMBER: _ClassVar[int]
    FRAME_RATE_FIELD_NUMBER: _ClassVar[int]
    hours: int
    minutes: int
    seconds: int
    frames: int
    frame_rate: float
    def __init__(self, hours: _Optional[int] = ..., minutes: _Optional[int] = ..., seconds: _Optional[int] = ..., frames: _Optional[int] = ..., frame_rate: _Optional[float] = ...) -> None: ...

class TimeRange(_message.Message):
    __slots__ = ("in_point", "out_point")
    IN_POINT_FIELD_NUMBER: _ClassVar[int]
    OUT_POINT_FIELD_NUMBER: _ClassVar[int]
    in_point: Timecode
    out_point: Timecode
    def __init__(self, in_point: _Optional[_Union[Timecode, _Mapping]] = ..., out_point: _Optional[_Union[Timecode, _Mapping]] = ...) -> None: ...

class Resolution(_message.Message):
    __slots__ = ("width", "height")
    WIDTH_FIELD_NUMBER: _ClassVar[int]
    HEIGHT_FIELD_NUMBER: _ClassVar[int]
    width: int
    height: int
    def __init__(self, width: _Optional[int] = ..., height: _Optional[int] = ...) -> None: ...

class VideoInfo(_message.Message):
    __slots__ = ("codec", "resolution", "frame_rate", "bitrate_bps", "pixel_format", "duration_seconds")
    CODEC_FIELD_NUMBER: _ClassVar[int]
    RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    FRAME_RATE_FIELD_NUMBER: _ClassVar[int]
    BITRATE_BPS_FIELD_NUMBER: _ClassVar[int]
    PIXEL_FORMAT_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    codec: str
    resolution: Resolution
    frame_rate: float
    bitrate_bps: int
    pixel_format: str
    duration_seconds: float
    def __init__(self, codec: _Optional[str] = ..., resolution: _Optional[_Union[Resolution, _Mapping]] = ..., frame_rate: _Optional[float] = ..., bitrate_bps: _Optional[int] = ..., pixel_format: _Optional[str] = ..., duration_seconds: _Optional[float] = ...) -> None: ...

class AudioInfo(_message.Message):
    __slots__ = ("codec", "sample_rate", "channels", "bitrate_bps", "duration_seconds")
    CODEC_FIELD_NUMBER: _ClassVar[int]
    SAMPLE_RATE_FIELD_NUMBER: _ClassVar[int]
    CHANNELS_FIELD_NUMBER: _ClassVar[int]
    BITRATE_BPS_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    codec: str
    sample_rate: int
    channels: int
    bitrate_bps: int
    duration_seconds: float
    def __init__(self, codec: _Optional[str] = ..., sample_rate: _Optional[int] = ..., channels: _Optional[int] = ..., bitrate_bps: _Optional[int] = ..., duration_seconds: _Optional[float] = ...) -> None: ...

class Asset(_message.Message):
    __slots__ = ("id", "file_path", "file_name", "file_size_bytes", "mime_type", "asset_type", "video", "audio", "metadata", "fingerprint")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    FILE_NAME_FIELD_NUMBER: _ClassVar[int]
    FILE_SIZE_BYTES_FIELD_NUMBER: _ClassVar[int]
    MIME_TYPE_FIELD_NUMBER: _ClassVar[int]
    ASSET_TYPE_FIELD_NUMBER: _ClassVar[int]
    VIDEO_FIELD_NUMBER: _ClassVar[int]
    AUDIO_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    FINGERPRINT_FIELD_NUMBER: _ClassVar[int]
    id: str
    file_path: str
    file_name: str
    file_size_bytes: int
    mime_type: str
    asset_type: AssetType
    video: VideoInfo
    audio: AudioInfo
    metadata: _containers.ScalarMap[str, str]
    fingerprint: str
    def __init__(self, id: _Optional[str] = ..., file_path: _Optional[str] = ..., file_name: _Optional[str] = ..., file_size_bytes: _Optional[int] = ..., mime_type: _Optional[str] = ..., asset_type: _Optional[_Union[AssetType, str]] = ..., video: _Optional[_Union[VideoInfo, _Mapping]] = ..., audio: _Optional[_Union[AudioInfo, _Mapping]] = ..., metadata: _Optional[_Mapping[str, str]] = ..., fingerprint: _Optional[str] = ...) -> None: ...

class EDLEntry(_message.Message):
    __slots__ = ("index", "source_asset_id", "source_range", "timeline_range", "track", "transition", "effects", "notes")
    INDEX_FIELD_NUMBER: _ClassVar[int]
    SOURCE_ASSET_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_RANGE_FIELD_NUMBER: _ClassVar[int]
    TIMELINE_RANGE_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    TRANSITION_FIELD_NUMBER: _ClassVar[int]
    EFFECTS_FIELD_NUMBER: _ClassVar[int]
    NOTES_FIELD_NUMBER: _ClassVar[int]
    index: int
    source_asset_id: str
    source_range: TimeRange
    timeline_range: TimeRange
    track: TrackTarget
    transition: TransitionInfo
    effects: _containers.RepeatedCompositeFieldContainer[EffectInfo]
    notes: str
    def __init__(self, index: _Optional[int] = ..., source_asset_id: _Optional[str] = ..., source_range: _Optional[_Union[TimeRange, _Mapping]] = ..., timeline_range: _Optional[_Union[TimeRange, _Mapping]] = ..., track: _Optional[_Union[TrackTarget, _Mapping]] = ..., transition: _Optional[_Union[TransitionInfo, _Mapping]] = ..., effects: _Optional[_Iterable[_Union[EffectInfo, _Mapping]]] = ..., notes: _Optional[str] = ...) -> None: ...

class TrackTarget(_message.Message):
    __slots__ = ("type", "track_index")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    TRACK_INDEX_FIELD_NUMBER: _ClassVar[int]
    type: TrackType
    track_index: int
    def __init__(self, type: _Optional[_Union[TrackType, str]] = ..., track_index: _Optional[int] = ...) -> None: ...

class TransitionInfo(_message.Message):
    __slots__ = ("type", "duration_seconds", "alignment")
    TYPE_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    ALIGNMENT_FIELD_NUMBER: _ClassVar[int]
    type: str
    duration_seconds: float
    alignment: str
    def __init__(self, type: _Optional[str] = ..., duration_seconds: _Optional[float] = ..., alignment: _Optional[str] = ...) -> None: ...

class EffectInfo(_message.Message):
    __slots__ = ("name", "parameters")
    class ParametersEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    NAME_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    name: str
    parameters: _containers.ScalarMap[str, str]
    def __init__(self, name: _Optional[str] = ..., parameters: _Optional[_Mapping[str, str]] = ...) -> None: ...

class EditDecisionList(_message.Message):
    __slots__ = ("id", "name", "sequence_resolution", "sequence_frame_rate", "entries")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_RESOLUTION_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_FRAME_RATE_FIELD_NUMBER: _ClassVar[int]
    ENTRIES_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    sequence_resolution: Resolution
    sequence_frame_rate: float
    entries: _containers.RepeatedCompositeFieldContainer[EDLEntry]
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., sequence_resolution: _Optional[_Union[Resolution, _Mapping]] = ..., sequence_frame_rate: _Optional[float] = ..., entries: _Optional[_Iterable[_Union[EDLEntry, _Mapping]]] = ...) -> None: ...

class TextStyle(_message.Message):
    __slots__ = ("font_family", "font_size", "color_hex", "alignment", "background_color_hex", "background_opacity", "position")
    FONT_FAMILY_FIELD_NUMBER: _ClassVar[int]
    FONT_SIZE_FIELD_NUMBER: _ClassVar[int]
    COLOR_HEX_FIELD_NUMBER: _ClassVar[int]
    ALIGNMENT_FIELD_NUMBER: _ClassVar[int]
    BACKGROUND_COLOR_HEX_FIELD_NUMBER: _ClassVar[int]
    BACKGROUND_OPACITY_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    font_family: str
    font_size: float
    color_hex: str
    alignment: str
    background_color_hex: str
    background_opacity: float
    position: Position
    def __init__(self, font_family: _Optional[str] = ..., font_size: _Optional[float] = ..., color_hex: _Optional[str] = ..., alignment: _Optional[str] = ..., background_color_hex: _Optional[str] = ..., background_opacity: _Optional[float] = ..., position: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Position(_message.Message):
    __slots__ = ("x", "y")
    X_FIELD_NUMBER: _ClassVar[int]
    Y_FIELD_NUMBER: _ClassVar[int]
    x: float
    y: float
    def __init__(self, x: _Optional[float] = ..., y: _Optional[float] = ...) -> None: ...
