"""Pacing and timing analysis sub-package.

Analyzes Edit Decision Lists to suggest duration adjustments
that match a target mood or pacing profile.
"""

from src.analysis.pacing import PacingAnalyzer
from src.analysis.rhythm import RhythmAnalyzer

__all__ = ["PacingAnalyzer", "RhythmAnalyzer"]
