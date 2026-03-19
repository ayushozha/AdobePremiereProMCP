"""EDL generation module.

Builds Edit Decision Lists from parsed scripts and matched assets.
The EDL tells the TypeScript bridge exactly how to assemble the timeline
in Premiere Pro.

Public API::

    from src.edl import EDLGenerator, TimelineCalculator, TrackAssigner, EDLValidator

    generator = EDLGenerator()
    edl = generator.generate(segments, matches, assets, settings)

    validator = EDLValidator()
    result = validator.validate(edl)
"""

from src.edl.generator import EDLGenerator
from src.edl.timeline_calculator import TimelineCalculator
from src.edl.track_assigner import TrackAssigner
from src.edl.validator import EDLValidator

__all__ = [
    "EDLGenerator",
    "EDLValidator",
    "TimelineCalculator",
    "TrackAssigner",
]
