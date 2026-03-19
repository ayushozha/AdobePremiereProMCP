"""Configuration for the Intelligence Service.

Uses pydantic-settings to load from environment variables with sensible defaults.
All settings can be overridden via env vars prefixed with ``INTEL_`` or set directly.
"""

from __future__ import annotations

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class IntelligenceSettings(BaseSettings):
    """Top-level configuration loaded from environment variables."""

    model_config = SettingsConfigDict(
        env_prefix="INTEL_",
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # ── Server ──────────────────────────────────────────────────────────────────
    grpc_port: int = Field(
        default=50053,
        description="Port the gRPC server listens on",
    )
    log_level: str = Field(
        default="INFO",
        description="Logging level (DEBUG, INFO, WARNING, ERROR)",
    )

    # ── AI / LLM Providers (optional) ───────────────────────────────────────────
    openai_api_key: str | None = Field(
        default=None,
        description="OpenAI API key for embedding generation",
    )
    anthropic_api_key: str | None = Field(
        default=None,
        description="Anthropic API key for script analysis",
    )

    # ── Model Selection ─────────────────────────────────────────────────────────
    embedding_model: str = Field(
        default="text-embedding-3-small",
        description="OpenAI embedding model to use for asset matching",
    )
    llm_model: str = Field(
        default="claude-sonnet-4-20250514",
        description="Anthropic LLM model for script analysis and reasoning",
    )

    # ── Matching Tuning ─────────────────────────────────────────────────────────
    match_confidence_threshold: float = Field(
        default=0.6,
        ge=0.0,
        le=1.0,
        description="Minimum confidence score to accept an asset match",
    )
    max_matches_per_segment: int = Field(
        default=3,
        ge=1,
        description="Maximum number of candidate matches returned per segment",
    )


def load_settings() -> IntelligenceSettings:
    """Create and return a validated ``IntelligenceSettings`` instance."""
    return IntelligenceSettings()
