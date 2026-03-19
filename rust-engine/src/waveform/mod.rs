//! Waveform analysis module.
//!
//! Analyzes audio tracks to produce waveform visualizations and
//! detect silence regions. The [`analyzer`] submodule uses ffmpeg
//! for PCM extraction when available.

pub mod analyzer;

// Re-export key types for convenience.
pub use analyzer::{
    SilenceRegion, WaveformAnalyzer, WaveformOptions, WaveformResult,
};
