"""gRPC server and IntelligenceService implementation.

Because the proto-generated Python stubs are not yet available (they will be
produced by ``grpcio-tools`` / ``buf generate``), this module defines the service
using grpcio's *generic* service handler pattern.  Each RPC method receives the
raw serialised request bytes, delegates to the appropriate domain module, and
returns serialised response bytes.

Once the generated ``_pb2`` / ``_pb2_grpc`` modules exist, this can be
refactored to subclass the generated servicer directly.
"""

from __future__ import annotations

import json
from concurrent import futures
from typing import Any

import grpc
import structlog

from src.analysis import PacingAnalyzer
from src.config import IntelligenceSettings
from src.edl import EDLGenerator
from src.matching import AssetMatcher
from src.models import (
    AssetInfo,
    AssetMatch,
    EDLSettings,
    EditDecisionList,
    MatchStrategy,
    ScriptSegment,
)
from src.parser import ScriptParser

logger: structlog.stdlib.BoundLogger = structlog.get_logger(__name__)

# ── Service name / method descriptors ────────────────────────────────────────────

_SERVICE_NAME = "premierpro.intelligence.v1.IntelligenceService"

_METHOD_PARSE_SCRIPT = f"/{_SERVICE_NAME}/ParseScript"
_METHOD_GENERATE_EDL = f"/{_SERVICE_NAME}/GenerateEDL"
_METHOD_MATCH_ASSETS = f"/{_SERVICE_NAME}/MatchAssets"
_METHOD_ANALYZE_PACING = f"/{_SERVICE_NAME}/AnalyzePacing"

# ── Shared service instances (initialised in create_server) ──────────────────────

_script_parser: ScriptParser | None = None
_edl_generator: EDLGenerator | None = None
_asset_matcher: AssetMatcher | None = None
_pacing_analyzer: PacingAnalyzer | None = None


def _decode_request(request_data: bytes) -> dict[str, Any]:
    """Decode incoming request bytes as a JSON object.

    Raises ``json.JSONDecodeError`` on malformed input.
    """
    return json.loads(request_data.decode("utf-8"))


# ── Real handlers ────────────────────────────────────────────────────────────────


def _handle_parse_script(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Parse a script into structured segments.

    Expected JSON request fields:
        - ``text`` (str, optional): Raw script text to parse.
        - ``file_path`` (str, optional): Path to a script file on disk.
        - ``format_hint`` (str, optional): Format hint (default ``"auto"``).

    At least one of ``text`` or ``file_path`` must be provided.
    """
    logger.info("parse_script.called", request_size=len(request_data))
    assert _script_parser is not None

    try:
        req = _decode_request(request_data)
    except (json.JSONDecodeError, UnicodeDecodeError) as exc:
        context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Malformed JSON request: {exc}")

    text: str | None = req.get("text")
    file_path: str | None = req.get("file_path")
    format_hint: str = req.get("format_hint", "auto")

    if not text and not file_path:
        context.abort(
            grpc.StatusCode.INVALID_ARGUMENT,
            "Request must include either 'text' or 'file_path'",
        )

    try:
        if file_path:
            result = _script_parser.parse_file(file_path, format_hint=format_hint)
        else:
            result = _script_parser.parse(text, format_hint=format_hint)  # type: ignore[arg-type]
    except FileNotFoundError as exc:
        context.abort(grpc.StatusCode.NOT_FOUND, str(exc))
    except ValueError as exc:
        context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
    except Exception as exc:
        logger.exception("parse_script.error")
        context.abort(grpc.StatusCode.INTERNAL, f"Script parsing failed: {exc}")

    return result.model_dump_json().encode("utf-8")  # type: ignore[possibly-undefined]


def _handle_generate_edl(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Generate an Edit Decision List from segments and matched assets.

    Expected JSON request fields:
        - ``segments`` (list[dict]): Serialised ``ScriptSegment`` objects.
        - ``matches`` (list[dict]): Serialised ``AssetMatch`` objects.
        - ``assets`` (list[dict]): Serialised ``AssetInfo`` objects.
        - ``settings`` (dict, optional): Serialised ``EDLSettings``.
    """
    logger.info("generate_edl.called", request_size=len(request_data))
    assert _edl_generator is not None

    try:
        req = _decode_request(request_data)
    except (json.JSONDecodeError, UnicodeDecodeError) as exc:
        context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Malformed JSON request: {exc}")

    try:
        segments = [ScriptSegment(**s) for s in req.get("segments", [])]  # type: ignore[possibly-undefined]
        matches = [AssetMatch(**m) for m in req.get("matches", [])]
        assets = [AssetInfo(**a) for a in req.get("assets", [])]
        settings = EDLSettings(**req.get("settings", {}))
    except Exception as exc:
        context.abort(
            grpc.StatusCode.INVALID_ARGUMENT,
            f"Failed to deserialise request fields: {exc}",
        )

    try:
        result = _edl_generator.generate(segments, matches, assets, settings)  # type: ignore[possibly-undefined]
    except Exception as exc:
        logger.exception("generate_edl.error")
        context.abort(grpc.StatusCode.INTERNAL, f"EDL generation failed: {exc}")

    return result.model_dump_json().encode("utf-8")  # type: ignore[possibly-undefined]


def _handle_match_assets(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Match script segments to available assets.

    Expected JSON request fields:
        - ``segments`` (list[dict]): Serialised ``ScriptSegment`` objects.
        - ``assets`` (list[dict]): Serialised ``AssetInfo`` objects.
        - ``strategy`` (str, optional): One of ``"keyword"``, ``"embedding"``,
          ``"hybrid"`` (default ``"hybrid"``).
    """
    logger.info("match_assets.called", request_size=len(request_data))
    assert _asset_matcher is not None

    try:
        req = _decode_request(request_data)
    except (json.JSONDecodeError, UnicodeDecodeError) as exc:
        context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Malformed JSON request: {exc}")

    try:
        segments = [ScriptSegment(**s) for s in req.get("segments", [])]  # type: ignore[possibly-undefined]
        assets = [AssetInfo(**a) for a in req.get("assets", [])]
    except Exception as exc:
        context.abort(
            grpc.StatusCode.INVALID_ARGUMENT,
            f"Failed to deserialise request fields: {exc}",
        )

    # Allow overriding strategy per-request.
    strategy_str: str = req.get("strategy", _asset_matcher.strategy.value)  # type: ignore[possibly-undefined]
    try:
        strategy = MatchStrategy(strategy_str.lower())
    except ValueError:
        context.abort(
            grpc.StatusCode.INVALID_ARGUMENT,
            f"Unknown match strategy '{strategy_str}'. "
            f"Valid values: {[s.value for s in MatchStrategy]}",
        )

    # Apply per-request strategy override.
    original_strategy = _asset_matcher.strategy
    _asset_matcher.strategy = strategy  # type: ignore[possibly-undefined]
    try:
        result = _asset_matcher.match(segments, assets)  # type: ignore[possibly-undefined]
    except Exception as exc:
        logger.exception("match_assets.error")
        context.abort(grpc.StatusCode.INTERNAL, f"Asset matching failed: {exc}")
    finally:
        _asset_matcher.strategy = original_strategy

    return result.model_dump_json().encode("utf-8")  # type: ignore[possibly-undefined]


def _handle_analyze_pacing(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Analyze pacing of an EDL and suggest adjustments.

    Expected JSON request fields:
        - ``edl`` (dict): Serialised ``EditDecisionList``.
        - ``target_mood`` (str, optional): Mood name (default ``"cinematic"``).
    """
    logger.info("analyze_pacing.called", request_size=len(request_data))
    assert _pacing_analyzer is not None

    try:
        req = _decode_request(request_data)
    except (json.JSONDecodeError, UnicodeDecodeError) as exc:
        context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Malformed JSON request: {exc}")

    try:
        edl = EditDecisionList(**req.get("edl", {}))  # type: ignore[possibly-undefined]
    except Exception as exc:
        context.abort(
            grpc.StatusCode.INVALID_ARGUMENT,
            f"Failed to deserialise EDL: {exc}",
        )

    target_mood: str = req.get("target_mood", "cinematic")  # type: ignore[possibly-undefined]

    try:
        result = _pacing_analyzer.analyze(edl, target_mood=target_mood)  # type: ignore[possibly-undefined]
    except Exception as exc:
        logger.exception("analyze_pacing.error")
        context.abort(grpc.StatusCode.INTERNAL, f"Pacing analysis failed: {exc}")

    return result.model_dump_json().encode("utf-8")  # type: ignore[possibly-undefined]


# ── Method routing table ────────────────────────────────────────────────────────

type RpcHandler = (
    grpc.unary_unary_rpc_method_handler
    | grpc.unary_stream_rpc_method_handler
    | None
)

_HANDLERS: dict[str, grpc.RpcMethodHandler] = {
    _METHOD_PARSE_SCRIPT: grpc.unary_unary_rpc_method_handler(_handle_parse_script),
    _METHOD_GENERATE_EDL: grpc.unary_unary_rpc_method_handler(_handle_generate_edl),
    _METHOD_MATCH_ASSETS: grpc.unary_unary_rpc_method_handler(_handle_match_assets),
    _METHOD_ANALYZE_PACING: grpc.unary_unary_rpc_method_handler(_handle_analyze_pacing),
}


class IntelligenceServiceHandler(grpc.GenericRpcHandler):
    """Generic gRPC handler that routes to the IntelligenceService methods.

    This approach avoids depending on generated proto stubs while still
    registering a properly-named service on the gRPC server.
    """

    def service_name(self) -> str | None:
        return _SERVICE_NAME

    def service(
        self,
        handler_call_details: grpc.HandlerCallDetails,
    ) -> grpc.RpcMethodHandler | None:
        method = handler_call_details.method
        handler = _HANDLERS.get(method)
        if handler is None:
            logger.warning("grpc.unknown_method", method=method)
        return handler


# ── Server lifecycle ─────────────────────────────────────────────────────────────


def create_server(settings: IntelligenceSettings) -> grpc.Server:
    """Build a gRPC server with the IntelligenceService registered.

    Initialises the real domain service instances (parser, EDL generator,
    asset matcher, pacing analyzer) using the provided settings.

    Args:
        settings: Validated application settings.

    Returns:
        A ``grpc.Server`` ready to be started (not yet calling ``start()``).
    """
    global _script_parser, _edl_generator, _asset_matcher, _pacing_analyzer

    # Initialise domain services.
    _script_parser = ScriptParser()
    _edl_generator = EDLGenerator()
    _asset_matcher = AssetMatcher(
        strategy=MatchStrategy.HYBRID,
        confidence_threshold=settings.match_confidence_threshold,
        max_matches_per_segment=settings.max_matches_per_segment,
        embedding_model=settings.embedding_model,
        openai_api_key=settings.openai_api_key,
    )
    _pacing_analyzer = PacingAnalyzer()

    logger.info(
        "services.initialised",
        parser=type(_script_parser).__name__,
        edl_generator=type(_edl_generator).__name__,
        asset_matcher=type(_asset_matcher).__name__,
        pacing_analyzer=type(_pacing_analyzer).__name__,
        match_strategy=_asset_matcher.strategy.value,
        match_threshold=settings.match_confidence_threshold,
    )

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        handlers=[IntelligenceServiceHandler()],
    )
    listen_addr = f"[::]:{settings.grpc_port}"
    server.add_insecure_port(listen_addr)
    logger.info("grpc.server_created", listen_addr=listen_addr)
    return server
