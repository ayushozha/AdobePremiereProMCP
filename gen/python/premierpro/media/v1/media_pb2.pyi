from premierpro.common.v1 import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ScanAssetsRequest(_message.Message):
    __slots__ = ("directory", "recursive", "extensions")
    DIRECTORY_FIELD_NUMBER: _ClassVar[int]
    RECURSIVE_FIELD_NUMBER: _ClassVar[int]
    EXTENSIONS_FIELD_NUMBER: _ClassVar[int]
    directory: str
    recursive: bool
    extensions: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, directory: _Optional[str] = ..., recursive: _Optional[bool] = ..., extensions: _Optional[_Iterable[str]] = ...) -> None: ...

class ScanAssetsResponse(_message.Message):
    __slots__ = ("assets", "total_files_scanned", "media_files_found", "scan_duration_seconds")
    ASSETS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FILES_SCANNED_FIELD_NUMBER: _ClassVar[int]
    MEDIA_FILES_FOUND_FIELD_NUMBER: _ClassVar[int]
    SCAN_DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    assets: _containers.RepeatedCompositeFieldContainer[_common_pb2.Asset]
    total_files_scanned: int
    media_files_found: int
    scan_duration_seconds: float
    def __init__(self, assets: _Optional[_Iterable[_Union[_common_pb2.Asset, _Mapping]]] = ..., total_files_scanned: _Optional[int] = ..., media_files_found: _Optional[int] = ..., scan_duration_seconds: _Optional[float] = ...) -> None: ...

class ProbeMediaRequest(_message.Message):
    __slots__ = ("file_path",)
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    file_path: str
    def __init__(self, file_path: _Optional[str] = ...) -> None: ...

class ProbeMediaResponse(_message.Message):
    __slots__ = ("asset",)
    ASSET_FIELD_NUMBER: _ClassVar[int]
    asset: _common_pb2.Asset
    def __init__(self, asset: _Optional[_Union[_common_pb2.Asset, _Mapping]] = ...) -> None: ...

class GenerateThumbnailRequest(_message.Message):
    __slots__ = ("file_path", "timestamp", "output_size", "output_format")
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_SIZE_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FORMAT_FIELD_NUMBER: _ClassVar[int]
    file_path: str
    timestamp: _common_pb2.Timecode
    output_size: _common_pb2.Resolution
    output_format: str
    def __init__(self, file_path: _Optional[str] = ..., timestamp: _Optional[_Union[_common_pb2.Timecode, _Mapping]] = ..., output_size: _Optional[_Union[_common_pb2.Resolution, _Mapping]] = ..., output_format: _Optional[str] = ...) -> None: ...

class GenerateThumbnailResponse(_message.Message):
    __slots__ = ("thumbnail_data", "output_path", "actual_size")
    THUMBNAIL_DATA_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_PATH_FIELD_NUMBER: _ClassVar[int]
    ACTUAL_SIZE_FIELD_NUMBER: _ClassVar[int]
    thumbnail_data: bytes
    output_path: str
    actual_size: _common_pb2.Resolution
    def __init__(self, thumbnail_data: _Optional[bytes] = ..., output_path: _Optional[str] = ..., actual_size: _Optional[_Union[_common_pb2.Resolution, _Mapping]] = ...) -> None: ...

class AnalyzeWaveformRequest(_message.Message):
    __slots__ = ("file_path", "audio_track", "silence_threshold_db", "min_silence_duration_seconds")
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    AUDIO_TRACK_FIELD_NUMBER: _ClassVar[int]
    SILENCE_THRESHOLD_DB_FIELD_NUMBER: _ClassVar[int]
    MIN_SILENCE_DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    file_path: str
    audio_track: int
    silence_threshold_db: float
    min_silence_duration_seconds: float
    def __init__(self, file_path: _Optional[str] = ..., audio_track: _Optional[int] = ..., silence_threshold_db: _Optional[float] = ..., min_silence_duration_seconds: _Optional[float] = ...) -> None: ...

class AnalyzeWaveformResponse(_message.Message):
    __slots__ = ("silence_regions", "peak_db", "rms_db", "duration_seconds", "waveform_samples")
    SILENCE_REGIONS_FIELD_NUMBER: _ClassVar[int]
    PEAK_DB_FIELD_NUMBER: _ClassVar[int]
    RMS_DB_FIELD_NUMBER: _ClassVar[int]
    DURATION_SECONDS_FIELD_NUMBER: _ClassVar[int]
    WAVEFORM_SAMPLES_FIELD_NUMBER: _ClassVar[int]
    silence_regions: _containers.RepeatedCompositeFieldContainer[SilenceRegion]
    peak_db: float
    rms_db: float
    duration_seconds: float
    waveform_samples: _containers.RepeatedScalarFieldContainer[float]
    def __init__(self, silence_regions: _Optional[_Iterable[_Union[SilenceRegion, _Mapping]]] = ..., peak_db: _Optional[float] = ..., rms_db: _Optional[float] = ..., duration_seconds: _Optional[float] = ..., waveform_samples: _Optional[_Iterable[float]] = ...) -> None: ...

class SilenceRegion(_message.Message):
    __slots__ = ("start_seconds", "end_seconds", "avg_db")
    START_SECONDS_FIELD_NUMBER: _ClassVar[int]
    END_SECONDS_FIELD_NUMBER: _ClassVar[int]
    AVG_DB_FIELD_NUMBER: _ClassVar[int]
    start_seconds: float
    end_seconds: float
    avg_db: float
    def __init__(self, start_seconds: _Optional[float] = ..., end_seconds: _Optional[float] = ..., avg_db: _Optional[float] = ...) -> None: ...

class DetectScenesRequest(_message.Message):
    __slots__ = ("file_path", "threshold")
    FILE_PATH_FIELD_NUMBER: _ClassVar[int]
    THRESHOLD_FIELD_NUMBER: _ClassVar[int]
    file_path: str
    threshold: float
    def __init__(self, file_path: _Optional[str] = ..., threshold: _Optional[float] = ...) -> None: ...

class DetectScenesResponse(_message.Message):
    __slots__ = ("scenes",)
    SCENES_FIELD_NUMBER: _ClassVar[int]
    scenes: _containers.RepeatedCompositeFieldContainer[SceneChange]
    def __init__(self, scenes: _Optional[_Iterable[_Union[SceneChange, _Mapping]]] = ...) -> None: ...

class SceneChange(_message.Message):
    __slots__ = ("timecode", "confidence")
    TIMECODE_FIELD_NUMBER: _ClassVar[int]
    CONFIDENCE_FIELD_NUMBER: _ClassVar[int]
    timecode: _common_pb2.Timecode
    confidence: float
    def __init__(self, timecode: _Optional[_Union[_common_pb2.Timecode, _Mapping]] = ..., confidence: _Optional[float] = ...) -> None: ...
