"""gRPC server and IntelligenceService implementation.

Uses the generated proto stubs to subclass ``IntelligenceServiceServicer``
directly. Each RPC method receives a typed proto request, converts it to
the corresponding Pydantic model, delegates to the domain implementation,
and converts the result back to a proto response.
"""

from __future__ import annotations

from concurrent import futures
from typing import Any

import grpc
import structlog

from premierpro.common.v1 import common_pb2
from premierpro.intelligence.v1 import intelligence_pb2, intelligence_pb2_grpc

from src.analysis import PacingAnalyzer
from src.config import IntelligenceSettings
from src.edl import EDLGenerator
from src.matching import AssetMatcher
from src.models import (
    AssetInfo,
    AssetMatch as PydanticAssetMatch,
    AssetType,
    AudioInfo,
    EDLEntry,
    EDLSettings,
    EditDecisionList,
    EffectInfo,
    MatchStrategy,
    PacingPreset,
    Resolution,
    ScriptSegment,
    SegmentType,
    Timecode,
    TimeRange,
    TrackTarget,
    TrackType,
    TransitionInfo,
    VideoInfo,
)
from src.parser import ScriptParser

logger: structlog.stdlib.BoundLogger = structlog.get_logger(__name__)

# ── Proto enum <-> Pydantic enum mapping tables ──────────────────────────────

_PROTO_SEGMENT_TYPE_TO_PYDANTIC: dict[int, SegmentType] = {
    intelligence_pb2.SEGMENT_TYPE_UNSPECIFIED: SegmentType.UNSPECIFIED,
    intelligence_pb2.SEGMENT_TYPE_DIALOGUE: SegmentType.DIALOGUE,
    intelligence_pb2.SEGMENT_TYPE_ACTION: SegmentType.ACTION,
    intelligence_pb2.SEGMENT_TYPE_BROLL: SegmentType.BROLL,
    intelligence_pb2.SEGMENT_TYPE_TRANSITION: SegmentType.TRANSITION,
    intelligence_pb2.SEGMENT_TYPE_TITLE: SegmentType.TITLE,
    intelligence_pb2.SEGMENT_TYPE_LOWER_THIRD: SegmentType.LOWER_THIRD,
    intelligence_pb2.SEGMENT_TYPE_VOICEOVER: SegmentType.VOICEOVER,
    intelligence_pb2.SEGMENT_TYPE_MUSIC: SegmentType.MUSIC,
    intelligence_pb2.SEGMENT_TYPE_SFX: SegmentType.SFX,
}

_PYDANTIC_SEGMENT_TYPE_TO_PROTO: dict[SegmentType, int] = {
    v: k for k, v in _PROTO_SEGMENT_TYPE_TO_PYDANTIC.items()
}

_PROTO_PACING_TO_PYDANTIC: dict[int, PacingPreset] = {
    intelligence_pb2.PACING_PRESET_UNSPECIFIED: PacingPreset.UNSPECIFIED,
    intelligence_pb2.PACING_PRESET_SLOW: PacingPreset.SLOW,
    intelligence_pb2.PACING_PRESET_MODERATE: PacingPreset.MODERATE,
    intelligence_pb2.PACING_PRESET_FAST: PacingPreset.FAST,
    intelligence_pb2.PACING_PRESET_DYNAMIC: PacingPreset.DYNAMIC,
}

_PROTO_MATCH_STRATEGY_TO_PYDANTIC: dict[int, MatchStrategy] = {
    intelligence_pb2.MATCH_STRATEGY_UNSPECIFIED: MatchStrategy.UNSPECIFIED,
    intelligence_pb2.MATCH_STRATEGY_KEYWORD: MatchStrategy.KEYWORD,
    intelligence_pb2.MATCH_STRATEGY_EMBEDDING: MatchStrategy.EMBEDDING,
    intelligence_pb2.MATCH_STRATEGY_HYBRID: MatchStrategy.HYBRID,
}

_PROTO_ASSET_TYPE_TO_PYDANTIC: dict[int, AssetType] = {
    common_pb2.ASSET_TYPE_UNSPECIFIED: AssetType.UNSPECIFIED,
    common_pb2.ASSET_TYPE_VIDEO: AssetType.VIDEO,
    common_pb2.ASSET_TYPE_AUDIO: AssetType.AUDIO,
    common_pb2.ASSET_TYPE_IMAGE: AssetType.IMAGE,
    common_pb2.ASSET_TYPE_GRAPHICS: AssetType.GRAPHICS,
}

_PYDANTIC_ASSET_TYPE_TO_PROTO: dict[AssetType, int] = {
    v: k for k, v in _PROTO_ASSET_TYPE_TO_PYDANTIC.items()
}

_PROTO_TRACK_TYPE_TO_PYDANTIC: dict[int, TrackType] = {
    common_pb2.TRACK_TYPE_UNSPECIFIED: TrackType.UNSPECIFIED,
    common_pb2.TRACK_TYPE_VIDEO: TrackType.VIDEO,
    common_pb2.TRACK_TYPE_AUDIO: TrackType.AUDIO,
}

_PYDANTIC_TRACK_TYPE_TO_PROTO: dict[TrackType, int] = {
    v: k for k, v in _PROTO_TRACK_TYPE_TO_PYDANTIC.items()
}


# ── Proto -> Pydantic converters ─────────────────────────────────────────────


def _proto_timecode_to_pydantic(tc: common_pb2.Timecode) -> Timecode:
    return Timecode(
        hours=tc.hours,
        minutes=tc.minutes,
        seconds=tc.seconds,
        frames=tc.frames,
        frame_rate=tc.frame_rate if tc.frame_rate > 0 else 24.0,
    )


def _proto_time_range_to_pydantic(tr: common_pb2.TimeRange) -> TimeRange:
    return TimeRange(
        in_point=_proto_timecode_to_pydantic(tr.in_point),
        out_point=_proto_timecode_to_pydantic(tr.out_point),
    )


def _proto_resolution_to_pydantic(res: common_pb2.Resolution) -> Resolution:
    return Resolution(
        width=res.width if res.width > 0 else 1920,
        height=res.height if res.height > 0 else 1080,
    )


def _proto_segment_to_pydantic(seg: intelligence_pb2.ScriptSegment) -> ScriptSegment:
    return ScriptSegment(
        index=seg.index,
        type=_PROTO_SEGMENT_TYPE_TO_PYDANTIC.get(seg.type, SegmentType.UNSPECIFIED),
        content=seg.content,
        speaker=seg.speaker,
        scene_description=seg.scene_description,
        visual_direction=seg.visual_direction,
        audio_direction=seg.audio_direction,
        estimated_duration_seconds=seg.estimated_duration_seconds,
        asset_hints=list(seg.asset_hints),
    )


def _proto_asset_to_pydantic(asset: common_pb2.Asset) -> AssetInfo:
    video = None
    if asset.HasField("video"):
        v = asset.video
        video = VideoInfo(
            codec=v.codec,
            resolution=_proto_resolution_to_pydantic(v.resolution) if v.HasField("resolution") else Resolution(),
            frame_rate=v.frame_rate,
            bitrate_bps=v.bitrate_bps,
            pixel_format=v.pixel_format,
            duration_seconds=v.duration_seconds,
        )

    audio = None
    if asset.HasField("audio"):
        a = asset.audio
        audio = AudioInfo(
            codec=a.codec,
            sample_rate=a.sample_rate,
            channels=a.channels,
            bitrate_bps=a.bitrate_bps,
            duration_seconds=a.duration_seconds,
        )

    return AssetInfo(
        id=asset.id,
        file_path=asset.file_path,
        file_name=asset.file_name,
        file_size_bytes=asset.file_size_bytes,
        mime_type=asset.mime_type,
        asset_type=_PROTO_ASSET_TYPE_TO_PYDANTIC.get(asset.asset_type, AssetType.UNSPECIFIED),
        video=video,
        audio=audio,
        metadata=dict(asset.metadata),
        fingerprint=asset.fingerprint,
    )


def _proto_asset_match_to_pydantic(m: intelligence_pb2.AssetMatch) -> PydanticAssetMatch:
    suggested_range = None
    if m.HasField("suggested_range"):
        suggested_range = _proto_time_range_to_pydantic(m.suggested_range)

    return PydanticAssetMatch(
        segment_index=m.segment_index,
        asset_id=m.asset_id,
        confidence=m.confidence,
        reasoning=m.reasoning,
        suggested_range=suggested_range,
    )


def _proto_edl_settings_to_pydantic(s: intelligence_pb2.EDLSettings) -> EDLSettings:
    resolution = Resolution()
    if s.HasField("resolution"):
        resolution = _proto_resolution_to_pydantic(s.resolution)

    return EDLSettings(
        resolution=resolution,
        frame_rate=s.frame_rate if s.frame_rate > 0 else 24.0,
        default_transition=s.default_transition or "cut",
        default_transition_duration=s.default_transition_duration,
        pacing=_PROTO_PACING_TO_PYDANTIC.get(s.pacing, PacingPreset.MODERATE),
    )


def _proto_track_target_to_pydantic(t: common_pb2.TrackTarget) -> TrackTarget:
    return TrackTarget(
        type=_PROTO_TRACK_TYPE_TO_PYDANTIC.get(t.type, TrackType.UNSPECIFIED),
        track_index=t.track_index,
    )


def _proto_transition_to_pydantic(t: common_pb2.TransitionInfo) -> TransitionInfo:
    return TransitionInfo(
        type=t.type,
        duration_seconds=t.duration_seconds,
        alignment=t.alignment,
    )


def _proto_effect_to_pydantic(e: common_pb2.EffectInfo) -> EffectInfo:
    return EffectInfo(
        name=e.name,
        parameters=dict(e.parameters),
    )


def _proto_edl_entry_to_pydantic(entry: common_pb2.EDLEntry) -> EDLEntry:
    transition = None
    if entry.HasField("transition"):
        transition = _proto_transition_to_pydantic(entry.transition)

    return EDLEntry(
        index=entry.index,
        source_asset_id=entry.source_asset_id,
        source_range=_proto_time_range_to_pydantic(entry.source_range),
        timeline_range=_proto_time_range_to_pydantic(entry.timeline_range),
        track=_proto_track_target_to_pydantic(entry.track),
        transition=transition,
        effects=[_proto_effect_to_pydantic(e) for e in entry.effects],
        notes=entry.notes,
    )


def _proto_edl_to_pydantic(edl: common_pb2.EditDecisionList) -> EditDecisionList:
    return EditDecisionList(
        id=edl.id,
        name=edl.name,
        sequence_resolution=_proto_resolution_to_pydantic(edl.sequence_resolution)
        if edl.HasField("sequence_resolution") else Resolution(),
        sequence_frame_rate=edl.sequence_frame_rate if edl.sequence_frame_rate > 0 else 24.0,
        entries=[_proto_edl_entry_to_pydantic(e) for e in edl.entries],
    )


# ── Pydantic -> Proto converters ─────────────────────────────────────────────


def _pydantic_timecode_to_proto(tc: Timecode) -> common_pb2.Timecode:
    return common_pb2.Timecode(
        hours=tc.hours,
        minutes=tc.minutes,
        seconds=tc.seconds,
        frames=tc.frames,
        frame_rate=tc.frame_rate,
    )


def _pydantic_time_range_to_proto(tr: TimeRange) -> common_pb2.TimeRange:
    return common_pb2.TimeRange(
        in_point=_pydantic_timecode_to_proto(tr.in_point),
        out_point=_pydantic_timecode_to_proto(tr.out_point),
    )


def _pydantic_segment_to_proto(seg: ScriptSegment) -> intelligence_pb2.ScriptSegment:
    return intelligence_pb2.ScriptSegment(
        index=seg.index,
        type=_PYDANTIC_SEGMENT_TYPE_TO_PROTO.get(
            seg.type, intelligence_pb2.SEGMENT_TYPE_UNSPECIFIED
        ),
        content=seg.content,
        speaker=seg.speaker,
        scene_description=seg.scene_description,
        visual_direction=seg.visual_direction,
        audio_direction=seg.audio_direction,
        estimated_duration_seconds=seg.estimated_duration_seconds,
        asset_hints=seg.asset_hints,
    )


def _pydantic_asset_match_to_proto(m: PydanticAssetMatch) -> intelligence_pb2.AssetMatch:
    suggested_range = None
    if m.suggested_range is not None:
        suggested_range = _pydantic_time_range_to_proto(m.suggested_range)

    return intelligence_pb2.AssetMatch(
        segment_index=m.segment_index,
        asset_id=m.asset_id,
        confidence=m.confidence,
        reasoning=m.reasoning,
        suggested_range=suggested_range,
    )


def _pydantic_track_target_to_proto(t: TrackTarget) -> common_pb2.TrackTarget:
    return common_pb2.TrackTarget(
        type=_PYDANTIC_TRACK_TYPE_TO_PROTO.get(t.type, common_pb2.TRACK_TYPE_UNSPECIFIED),
        track_index=t.track_index,
    )


def _pydantic_transition_to_proto(t: TransitionInfo) -> common_pb2.TransitionInfo:
    return common_pb2.TransitionInfo(
        type=t.type,
        duration_seconds=t.duration_seconds,
        alignment=t.alignment,
    )


def _pydantic_effect_to_proto(e: EffectInfo) -> common_pb2.EffectInfo:
    return common_pb2.EffectInfo(
        name=e.name,
        parameters=e.parameters,
    )


def _pydantic_edl_entry_to_proto(entry: EDLEntry) -> common_pb2.EDLEntry:
    transition = None
    if entry.transition is not None:
        transition = _pydantic_transition_to_proto(entry.transition)

    return common_pb2.EDLEntry(
        index=entry.index,
        source_asset_id=entry.source_asset_id,
        source_range=_pydantic_time_range_to_proto(entry.source_range),
        timeline_range=_pydantic_time_range_to_proto(entry.timeline_range),
        track=_pydantic_track_target_to_proto(entry.track),
        transition=transition,
        effects=[_pydantic_effect_to_proto(e) for e in entry.effects],
        notes=entry.notes,
    )


def _pydantic_edl_to_proto(edl: EditDecisionList) -> common_pb2.EditDecisionList:
    return common_pb2.EditDecisionList(
        id=edl.id,
        name=edl.name,
        sequence_resolution=common_pb2.Resolution(
            width=edl.sequence_resolution.width,
            height=edl.sequence_resolution.height,
        ),
        sequence_frame_rate=edl.sequence_frame_rate,
        entries=[_pydantic_edl_entry_to_proto(e) for e in edl.entries],
    )


# ── IntelligenceServicer implementation ──────────────────────────────────────


class IntelligenceServicer(intelligence_pb2_grpc.IntelligenceServiceServicer):
    """Implements the IntelligenceService using the generated proto servicer base class.

    Each method converts proto request -> Pydantic models, calls the domain
    implementation, and converts the result back to a proto response.
    """

    def __init__(
        self,
        script_parser: ScriptParser,
        edl_generator: EDLGenerator,
        asset_matcher: AssetMatcher,
        pacing_analyzer: PacingAnalyzer,
    ) -> None:
        self._script_parser = script_parser
        self._edl_generator = edl_generator
        self._asset_matcher = asset_matcher
        self._pacing_analyzer = pacing_analyzer

    # ── ParseScript ──────────────────────────────────────────────────────

    def ParseScript(
        self,
        request: intelligence_pb2.ParseScriptRequest,
        context: grpc.ServicerContext,
    ) -> intelligence_pb2.ParseScriptResponse:
        """Parse a script (text or file) into structured segments."""
        logger.info("parse_script.called")

        text = request.text or None
        file_path = request.file_path or None
        format_hint = request.format_hint or "auto"

        if not text and not file_path:
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                "Request must include either 'text' or 'file_path'",
            )

        try:
            if file_path:
                result = self._script_parser.parse_file(file_path, format_hint=format_hint)
            else:
                result = self._script_parser.parse(text, format_hint=format_hint)  # type: ignore[arg-type]
        except FileNotFoundError as exc:
            context.abort(grpc.StatusCode.NOT_FOUND, str(exc))
        except ValueError as exc:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
        except Exception as exc:
            logger.exception("parse_script.error")
            context.abort(grpc.StatusCode.INTERNAL, f"Script parsing failed: {exc}")

        # Convert Pydantic result -> proto response.
        proto_segments = [_pydantic_segment_to_proto(s) for s in result.segments]  # type: ignore[possibly-undefined]
        proto_metadata = intelligence_pb2.ScriptMetadata(
            title=result.metadata.title,
            format=result.metadata.format.value if hasattr(result.metadata.format, 'value') else str(result.metadata.format),
            estimated_total_duration_seconds=result.metadata.estimated_total_duration_seconds,
            segment_count=result.metadata.segment_count,
        )

        return intelligence_pb2.ParseScriptResponse(
            segments=proto_segments,
            metadata=proto_metadata,
        )

    # ── GenerateEDL ──────────────────────────────────────────────────────

    def GenerateEDL(
        self,
        request: intelligence_pb2.GenerateEDLRequest,
        context: grpc.ServicerContext,
    ) -> intelligence_pb2.GenerateEDLResponse:
        """Generate an Edit Decision List from a parsed script and available assets."""
        logger.info("generate_edl.called")

        try:
            segments = [_proto_segment_to_pydantic(s) for s in request.segments]
            assets = [_proto_asset_to_pydantic(a) for a in request.available_assets]
            matches = [_proto_asset_match_to_pydantic(m) for m in request.matches]
            settings = _proto_edl_settings_to_pydantic(request.settings) if request.HasField("settings") else EDLSettings()
        except Exception as exc:
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"Failed to deserialise request fields: {exc}",
            )

        try:
            edl, warnings = self._edl_generator.generate_with_warnings(  # type: ignore[possibly-undefined]
                segments, matches, assets, settings,
            )
        except Exception as exc:
            logger.exception("generate_edl.error")
            context.abort(grpc.StatusCode.INTERNAL, f"EDL generation failed: {exc}")

        return intelligence_pb2.GenerateEDLResponse(
            edl=_pydantic_edl_to_proto(edl),  # type: ignore[possibly-undefined]
            warnings=warnings,  # type: ignore[possibly-undefined]
        )

    # ── MatchAssets ──────────────────────────────────────────────────────

    def MatchAssets(
        self,
        request: intelligence_pb2.MatchAssetsRequest,
        context: grpc.ServicerContext,
    ) -> intelligence_pb2.MatchAssetsResponse:
        """Match script segments to the best available assets."""
        logger.info("match_assets.called")

        try:
            segments = [_proto_segment_to_pydantic(s) for s in request.segments]
            assets = [_proto_asset_to_pydantic(a) for a in request.available_assets]
        except Exception as exc:
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"Failed to deserialise request fields: {exc}",
            )

        # Map proto strategy enum to Pydantic MatchStrategy.
        strategy = _PROTO_MATCH_STRATEGY_TO_PYDANTIC.get(
            request.strategy, MatchStrategy.HYBRID,
        )
        # Use HYBRID as default if UNSPECIFIED.
        if strategy == MatchStrategy.UNSPECIFIED:
            strategy = MatchStrategy.HYBRID

        # Apply per-request strategy override.
        original_strategy = self._asset_matcher.strategy
        self._asset_matcher.strategy = strategy
        try:
            result = self._asset_matcher.match(segments, assets)  # type: ignore[possibly-undefined]
        except Exception as exc:
            logger.exception("match_assets.error")
            context.abort(grpc.StatusCode.INTERNAL, f"Asset matching failed: {exc}")
        finally:
            self._asset_matcher.strategy = original_strategy

        # Convert Pydantic result -> proto response.
        proto_matches = [_pydantic_asset_match_to_proto(m) for m in result.matches]  # type: ignore[possibly-undefined]
        proto_unmatched = [
            intelligence_pb2.UnmatchedSegment(
                segment_index=u.segment_index,
                reason=u.reason,
                suggestions=u.suggestions,
            )
            for u in result.unmatched
        ]

        return intelligence_pb2.MatchAssetsResponse(
            matches=proto_matches,
            unmatched=proto_unmatched,
        )

    # ── AnalyzePacing ────────────────────────────────────────────────────

    def AnalyzePacing(
        self,
        request: intelligence_pb2.AnalyzePacingRequest,
        context: grpc.ServicerContext,
    ) -> intelligence_pb2.AnalyzePacingResponse:
        """Analyze pacing and suggest timing adjustments."""
        logger.info("analyze_pacing.called")

        try:
            edl = _proto_edl_to_pydantic(request.edl)
        except Exception as exc:
            context.abort(
                grpc.StatusCode.INVALID_ARGUMENT,
                f"Failed to deserialise EDL: {exc}",
            )

        target_mood = request.target_mood or "cinematic"

        try:
            result = self._pacing_analyzer.analyze(edl, target_mood=target_mood)  # type: ignore[possibly-undefined]
        except Exception as exc:
            logger.exception("analyze_pacing.error")
            context.abort(grpc.StatusCode.INTERNAL, f"Pacing analysis failed: {exc}")

        # Convert Pydantic result -> proto response.
        proto_adjustments = [
            intelligence_pb2.PacingAdjustment(
                edl_entry_index=a.edl_entry_index,
                current_duration=a.current_duration,
                suggested_duration=a.suggested_duration,
                reason=a.reason,
            )
            for a in result.adjustments  # type: ignore[possibly-undefined]
        ]

        return intelligence_pb2.AnalyzePacingResponse(
            adjustments=proto_adjustments,
            current_avg_clip_duration=result.current_avg_clip_duration,
            suggested_avg_clip_duration=result.suggested_avg_clip_duration,
        )


# ── Server lifecycle ─────────────────────────────────────────────────────────


def create_server(settings: IntelligenceSettings) -> grpc.Server:
    """Build a gRPC server with the IntelligenceService registered.

    Initialises the real domain service instances (parser, EDL generator,
    asset matcher, pacing analyzer) using the provided settings.

    Args:
        settings: Validated application settings.

    Returns:
        A ``grpc.Server`` ready to be started (not yet calling ``start()``).
    """
    # Initialise domain services.
    script_parser = ScriptParser()
    edl_generator = EDLGenerator()
    asset_matcher = AssetMatcher(
        strategy=MatchStrategy.HYBRID,
        confidence_threshold=settings.match_confidence_threshold,
        max_matches_per_segment=settings.max_matches_per_segment,
        embedding_model=settings.embedding_model,
        openai_api_key=settings.openai_api_key,
    )
    pacing_analyzer = PacingAnalyzer()

    logger.info(
        "services.initialised",
        parser=type(script_parser).__name__,
        edl_generator=type(edl_generator).__name__,
        asset_matcher=type(asset_matcher).__name__,
        pacing_analyzer=type(pacing_analyzer).__name__,
        match_strategy=asset_matcher.strategy.value,
        match_threshold=settings.match_confidence_threshold,
    )

    servicer = IntelligenceServicer(
        script_parser=script_parser,
        edl_generator=edl_generator,
        asset_matcher=asset_matcher,
        pacing_analyzer=pacing_analyzer,
    )

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
    )
    intelligence_pb2_grpc.add_IntelligenceServiceServicer_to_server(servicer, server)

    listen_addr = f"[::]:{settings.grpc_port}"
    server.add_insecure_port(listen_addr)
    logger.info("grpc.server_created", listen_addr=listen_addr)
    return server
