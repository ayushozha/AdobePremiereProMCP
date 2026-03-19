"""Entry point for the PremierPro Intelligence gRPC service.

Usage::

    python -m src.main                       # defaults: port 50053, INFO logging
    python -m src.main --port 50054 --log-level DEBUG
"""

from __future__ import annotations

import argparse
import os
import signal
import sys
from pathlib import Path
from types import FrameType

# Ensure the generated proto stubs are importable.
# Resolve the gen/python directory relative to this file's location:
#   python-intelligence/src/main.py  ->  ../../gen/python
_GEN_PYTHON_DIR = str(Path(__file__).resolve().parent.parent.parent / "gen" / "python")
if _GEN_PYTHON_DIR not in sys.path:
    sys.path.insert(0, _GEN_PYTHON_DIR)

import structlog

from src.config import load_settings
from src.grpc_server import create_server


def _configure_logging(level: str) -> None:
    """Set up ``structlog`` with human-readable console output."""
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.StackInfoRenderer(),
            structlog.dev.set_exc_info,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.dev.ConsoleRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(
            {"debug": 10, "info": 20, "warn": 30, "warning": 30, "error": 40}.get(level.lower(), 20),
        ),
        context_class=dict,
        logger_factory=structlog.PrintLoggerFactory(),
        cache_logger_on_first_use=True,
    )


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    """Parse CLI arguments, falling back to env-var settings for defaults."""
    parser = argparse.ArgumentParser(
        description="PremierPro Intelligence gRPC Service",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=None,
        help="gRPC listen port (default: from INTEL_GRPC_PORT or 50053)",
    )
    parser.add_argument(
        "--log-level",
        type=str,
        default=None,
        choices=["DEBUG", "INFO", "WARNING", "ERROR"],
        help="Logging level (default: from INTEL_LOG_LEVEL or INFO)",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    """Initialise configuration, logging, and start the gRPC server."""
    args = _parse_args(argv)

    # Load settings from env, then apply CLI overrides.
    settings = load_settings()
    if args.port is not None:
        settings.grpc_port = args.port
    if args.log_level is not None:
        settings.log_level = args.log_level

    _configure_logging(settings.log_level)
    log = structlog.get_logger("main")

    log.info(
        "settings.loaded",
        grpc_port=settings.grpc_port,
        log_level=settings.log_level,
        embedding_model=settings.embedding_model,
        llm_model=settings.llm_model,
        openai_configured=settings.openai_api_key is not None,
        anthropic_configured=settings.anthropic_api_key is not None,
    )

    server = create_server(settings)
    server.start()
    log.info("server.started", port=settings.grpc_port)

    # ── Graceful shutdown ────────────────────────────────────────────────────
    shutdown_requested = False

    def _shutdown(signum: int, _frame: FrameType | None) -> None:
        nonlocal shutdown_requested
        if shutdown_requested:
            return
        shutdown_requested = True
        sig_name = signal.Signals(signum).name
        log.info("server.shutting_down", signal=sig_name)
        # Give in-flight RPCs 5 seconds to complete.
        server.stop(grace=5.0)

    signal.signal(signal.SIGINT, _shutdown)
    signal.signal(signal.SIGTERM, _shutdown)

    log.info("server.awaiting_termination")
    server.wait_for_termination()
    log.info("server.stopped")


if __name__ == "__main__":
    main()
