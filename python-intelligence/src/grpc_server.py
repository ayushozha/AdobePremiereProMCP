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

from concurrent import futures

import grpc
import structlog

from src.config import IntelligenceSettings
from src.models import (
    EditDecisionList,
    MatchResult,
    PacingResult,
    ParsedScript,
)

logger: structlog.stdlib.BoundLogger = structlog.get_logger(__name__)

# ── Service name / method descriptors ────────────────────────────────────────────

_SERVICE_NAME = "premierpro.intelligence.v1.IntelligenceService"

_METHOD_PARSE_SCRIPT = f"/{_SERVICE_NAME}/ParseScript"
_METHOD_GENERATE_EDL = f"/{_SERVICE_NAME}/GenerateEDL"
_METHOD_MATCH_ASSETS = f"/{_SERVICE_NAME}/MatchAssets"
_METHOD_ANALYZE_PACING = f"/{_SERVICE_NAME}/AnalyzePacing"


# ── Placeholder handlers ────────────────────────────────────────────────────────
#
# These will be replaced with real logic once the parser, edl, matching, and
# analysis sub-packages are implemented.  For now they return minimal valid
# responses so the gRPC wiring can be tested end-to-end.


def _handle_parse_script(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Parse a script into structured segments."""
    logger.info("parse_script.called", request_size=len(request_data))

    # Placeholder: return an empty ParsedScript as JSON-encoded bytes.
    result = ParsedScript()
    return result.model_dump_json().encode("utf-8")


def _handle_generate_edl(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Generate an Edit Decision List from segments and matched assets."""
    logger.info("generate_edl.called", request_size=len(request_data))

    result = EditDecisionList()
    return result.model_dump_json().encode("utf-8")


def _handle_match_assets(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Match script segments to available assets."""
    logger.info("match_assets.called", request_size=len(request_data))

    result = MatchResult()
    return result.model_dump_json().encode("utf-8")


def _handle_analyze_pacing(
    request_data: bytes,
    context: grpc.ServicerContext,
) -> bytes:
    """Analyze pacing of an EDL and suggest adjustments."""
    logger.info("analyze_pacing.called", request_size=len(request_data))

    result = PacingResult()
    return result.model_dump_json().encode("utf-8")


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

    Args:
        settings: Validated application settings.

    Returns:
        A ``grpc.Server`` ready to be started (not yet calling ``start()``).
    """
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        handlers=[IntelligenceServiceHandler()],
    )
    listen_addr = f"[::]:{settings.grpc_port}"
    server.add_insecure_port(listen_addr)
    logger.info("grpc.server_created", listen_addr=listen_addr)
    return server
