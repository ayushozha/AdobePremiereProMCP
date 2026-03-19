//! Audio waveform analysis via ffmpeg CLI.
//!
//! Extracts raw PCM samples by piping through ffmpeg, then computes:
//!
//! - Peak and RMS amplitude (dB)
//! - Silence regions (configurable threshold + minimum duration)
//! - A downsampled waveform suitable for visualisation (~1 000 samples)

use anyhow::{Context, Result};
use std::process::{Command, Stdio};
use tracing::{debug, info, warn};

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

/// Options that control waveform analysis.
#[derive(Debug, Clone)]
pub struct WaveformOptions {
    /// Zero-based audio track index (passed to ffmpeg via stream mapping).
    pub audio_track: u32,
    /// Silence threshold in dBFS.  Samples whose amplitude is below this
    /// value are considered silent.  Typical default: -40.0.
    pub silence_threshold_db: f64,
    /// Minimum contiguous silence duration (seconds) to report.
    pub min_silence_duration: f64,
}

impl Default for WaveformOptions {
    fn default() -> Self {
        Self {
            audio_track: 0,
            silence_threshold_db: -40.0,
            min_silence_duration: 0.5,
        }
    }
}

// ---------------------------------------------------------------------------
// Results
// ---------------------------------------------------------------------------

/// Result of an audio waveform analysis pass.
#[derive(Debug, Clone)]
pub struct WaveformResult {
    /// Detected silence regions.
    pub silence_regions: Vec<SilenceRegion>,
    /// Peak amplitude in dBFS.
    pub peak_db: f64,
    /// Root-mean-square amplitude in dBFS.
    pub rms_db: f64,
    /// Total audio duration in seconds (as reported by the sample count and
    /// sample rate).
    pub duration_seconds: f64,
    /// Downsampled waveform suitable for visualization.
    /// Each value is a linear amplitude in the range \[0.0, 1.0\].
    pub waveform_samples: Vec<f32>,
}

/// A contiguous region of silence.
#[derive(Debug, Clone)]
pub struct SilenceRegion {
    /// Start time in seconds.
    pub start_seconds: f64,
    /// End time in seconds.
    pub end_seconds: f64,
    /// Average amplitude of the region in dBFS.
    pub avg_db: f64,
}

// ---------------------------------------------------------------------------
// Analyzer
// ---------------------------------------------------------------------------

/// Sample rate used when decoding audio via ffmpeg.
const DECODE_SAMPLE_RATE: u32 = 8_000;

/// Target number of samples in the downsampled waveform for visualization.
const WAVEFORM_VIS_SAMPLES: usize = 1_000;

/// Minimum amplitude treated as non-zero to avoid log(0).
const AMPLITUDE_FLOOR: f64 = 1e-10;

/// Stateless audio waveform analyzer.
pub struct WaveformAnalyzer;

impl WaveformAnalyzer {
    /// Analyse the audio content of a media file.
    ///
    /// Spawns an ffmpeg subprocess that decodes the requested audio track
    /// into mono 32-bit float PCM at [`DECODE_SAMPLE_RATE`] Hz, then streams
    /// the raw samples through the analysis pipeline.
    ///
    /// # Errors
    ///
    /// Returns an error if ffmpeg is not found, the input file cannot be
    /// read, or no audio samples are produced.
    pub fn analyze(file_path: &str, options: &WaveformOptions) -> Result<WaveformResult> {
        info!(path = %file_path, track = options.audio_track, "starting waveform analysis");

        let samples = Self::extract_pcm_samples(file_path, options)
            .context("failed to extract PCM samples via ffmpeg")?;

        if samples.is_empty() {
            anyhow::bail!(
                "ffmpeg produced zero audio samples — does the file contain an audio track?"
            );
        }

        let duration_seconds = samples.len() as f64 / DECODE_SAMPLE_RATE as f64;

        let (peak_db, rms_db) = Self::compute_levels(&samples);
        let silence_regions =
            Self::detect_silence(&samples, DECODE_SAMPLE_RATE, options);
        let waveform_samples = Self::downsample_waveform(&samples, WAVEFORM_VIS_SAMPLES);

        info!(
            duration_s = duration_seconds,
            peak_db,
            rms_db,
            silence_count = silence_regions.len(),
            "waveform analysis complete"
        );

        Ok(WaveformResult {
            silence_regions,
            peak_db,
            rms_db,
            duration_seconds,
            waveform_samples,
        })
    }

    // -----------------------------------------------------------------------
    // Internal helpers
    // -----------------------------------------------------------------------

    /// Spawn ffmpeg and capture raw f32le PCM from stdout.
    fn extract_pcm_samples(file_path: &str, options: &WaveformOptions) -> Result<Vec<f32>> {
        let audio_map = format!("0:a:{}", options.audio_track);

        let child = Command::new("ffmpeg")
            .args([
                "-hide_banner",
                "-loglevel",
                "error",
                "-i",
                file_path,
                "-map",
                &audio_map,
                "-ac",
                "1",
                "-ar",
                &DECODE_SAMPLE_RATE.to_string(),
                "-f",
                "f32le",
                "-vn",
                "pipe:1",
            ])
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .context("failed to spawn ffmpeg — is it installed and on PATH?")?;

        let output = child
            .wait_with_output()
            .context("failed to read ffmpeg output")?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            anyhow::bail!("ffmpeg exited with {}: {}", output.status, stderr.trim());
        }

        let raw = &output.stdout;
        if raw.len() % 4 != 0 {
            warn!(
                byte_count = raw.len(),
                "ffmpeg output is not a multiple of 4 bytes; truncating trailing bytes"
            );
        }

        let sample_count = raw.len() / 4;
        let mut samples = Vec::with_capacity(sample_count);
        for chunk in raw.chunks_exact(4) {
            let bytes: [u8; 4] = chunk.try_into().unwrap();
            samples.push(f32::from_le_bytes(bytes));
        }

        debug!(sample_count, "decoded PCM samples");
        Ok(samples)
    }

    /// Compute peak and RMS levels (dBFS) over the full sample buffer.
    fn compute_levels(samples: &[f32]) -> (f64, f64) {
        let mut peak: f64 = 0.0;
        let mut sum_sq: f64 = 0.0;

        for &s in samples {
            let abs = (s as f64).abs();
            if abs > peak {
                peak = abs;
            }
            sum_sq += (s as f64) * (s as f64);
        }

        let rms = (sum_sq / samples.len() as f64).sqrt();
        let peak_db = amplitude_to_db(peak);
        let rms_db = amplitude_to_db(rms);

        (peak_db, rms_db)
    }

    /// Walk through samples and identify contiguous regions whose amplitude
    /// stays below the silence threshold for at least
    /// [`WaveformOptions::min_silence_duration`] seconds.
    fn detect_silence(
        samples: &[f32],
        sample_rate: u32,
        options: &WaveformOptions,
    ) -> Vec<SilenceRegion> {
        let threshold_linear = db_to_amplitude(options.silence_threshold_db);
        let min_samples = (options.min_silence_duration * sample_rate as f64) as usize;

        let mut regions: Vec<SilenceRegion> = Vec::new();
        let mut silence_start: Option<usize> = None;
        let mut silence_sum: f64 = 0.0;
        let mut silence_count: usize = 0;

        for (i, &sample) in samples.iter().enumerate() {
            let abs = (sample as f64).abs();

            if abs < threshold_linear {
                if silence_start.is_none() {
                    silence_start = Some(i);
                    silence_sum = 0.0;
                    silence_count = 0;
                }
                silence_sum += abs;
                silence_count += 1;
            } else {
                if let Some(start) = silence_start {
                    let length = i - start;
                    if length >= min_samples {
                        let avg_amplitude = silence_sum / silence_count as f64;
                        regions.push(SilenceRegion {
                            start_seconds: start as f64 / sample_rate as f64,
                            end_seconds: i as f64 / sample_rate as f64,
                            avg_db: amplitude_to_db(avg_amplitude),
                        });
                    }
                }
                silence_start = None;
            }
        }

        // Handle trailing silence.
        if let Some(start) = silence_start {
            let length = samples.len() - start;
            if length >= min_samples {
                let avg_amplitude = if silence_count > 0 {
                    silence_sum / silence_count as f64
                } else {
                    0.0
                };
                regions.push(SilenceRegion {
                    start_seconds: start as f64 / sample_rate as f64,
                    end_seconds: samples.len() as f64 / sample_rate as f64,
                    avg_db: amplitude_to_db(avg_amplitude),
                });
            }
        }

        regions
    }

    /// Downsample the raw sample buffer into `target_count` bins.
    ///
    /// Each bin contains the *peak absolute amplitude* of the samples that
    /// fall within it, normalised to \[0.0, 1.0\].
    fn downsample_waveform(samples: &[f32], target_count: usize) -> Vec<f32> {
        if samples.is_empty() {
            return Vec::new();
        }

        let target = target_count.min(samples.len());
        let bin_size = samples.len() as f64 / target as f64;

        let mut waveform = Vec::with_capacity(target);
        let mut global_peak: f32 = 0.0;

        for i in 0..target {
            let start = (i as f64 * bin_size) as usize;
            let end = (((i + 1) as f64) * bin_size) as usize;
            let end = end.min(samples.len());

            let peak = samples[start..end]
                .iter()
                .map(|s| s.abs())
                .fold(0.0f32, f32::max);

            if peak > global_peak {
                global_peak = peak;
            }
            waveform.push(peak);
        }

        // Normalise to [0, 1].
        if global_peak > 0.0 {
            for v in &mut waveform {
                *v /= global_peak;
            }
        }

        waveform
    }
}

// ---------------------------------------------------------------------------
// dB <-> amplitude conversions
// ---------------------------------------------------------------------------

/// Convert a linear amplitude to dBFS.
fn amplitude_to_db(amplitude: f64) -> f64 {
    20.0 * amplitude.max(AMPLITUDE_FLOOR).log10()
}

/// Convert a dBFS value to linear amplitude.
fn db_to_amplitude(db: f64) -> f64 {
    10.0_f64.powf(db / 20.0)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn amplitude_db_roundtrip() {
        let db = -20.0;
        let amp = db_to_amplitude(db);
        let back = amplitude_to_db(amp);
        assert!((back - db).abs() < 1e-9);
    }

    #[test]
    fn full_scale_is_zero_db() {
        let db = amplitude_to_db(1.0);
        assert!(db.abs() < 1e-9, "1.0 linear should be 0 dBFS, got {db}");
    }

    #[test]
    fn downsample_empty() {
        let result = WaveformAnalyzer::downsample_waveform(&[], 100);
        assert!(result.is_empty());
    }

    #[test]
    fn downsample_normalises() {
        let samples: Vec<f32> = (0..1000).map(|i| (i as f32 / 999.0) * 0.5).collect();
        let ds = WaveformAnalyzer::downsample_waveform(&samples, 10);
        assert_eq!(ds.len(), 10);
        // Last bin should be normalised to 1.0.
        assert!((ds.last().unwrap() - 1.0).abs() < 0.02);
    }

    #[test]
    fn silence_detection_basic() {
        // 8000 Hz, 2 seconds of silence at amplitude 0.0001, then 1 second loud.
        let sr = 8_000u32;
        let mut samples: Vec<f32> = vec![0.0001; sr as usize * 2];
        samples.extend(vec![0.5f32; sr as usize]);

        let opts = WaveformOptions {
            audio_track: 0,
            silence_threshold_db: -40.0,
            min_silence_duration: 0.5,
        };

        let regions = WaveformAnalyzer::detect_silence(&samples, sr, &opts);
        assert_eq!(regions.len(), 1);
        assert!((regions[0].start_seconds - 0.0).abs() < 0.01);
        assert!((regions[0].end_seconds - 2.0).abs() < 0.01);
    }
}
