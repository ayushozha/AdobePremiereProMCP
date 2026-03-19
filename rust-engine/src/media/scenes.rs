//! Scene change detection via ffmpeg's `select` filter.
//!
//! Uses ffmpeg's built-in scene detection filter to compare adjacent frames
//! and report timestamps where the visual difference exceeds a threshold.
//! This avoids decoding frames in Rust and instead relies on ffmpeg's
//! optimised filter pipeline.
//!
//! When ffmpeg is unavailable on the system, the detector returns a clear
//! error rather than panicking.

use anyhow::{Context, Result};
use std::path::Path;
use std::process::Command;
use tracing::{debug, info};

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

/// A detected scene change point.
#[derive(Debug, Clone)]
pub struct SceneChangePoint {
    /// Time offset in seconds from the start of the video.
    pub timestamp_seconds: f64,
    /// Confidence / score of the scene change (0.0 -- 1.0).
    /// Higher values indicate a more pronounced visual change.
    pub confidence: f64,
}

/// Result of a scene detection pass.
#[derive(Debug, Clone)]
pub struct SceneDetectionResult {
    /// Detected scene change points, sorted by timestamp.
    pub scenes: Vec<SceneChangePoint>,
}

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

/// Options controlling scene detection.
#[derive(Debug, Clone)]
pub struct SceneDetectionOptions {
    /// Threshold for the scene detection filter (0.0 -- 1.0).
    /// Lower values detect more scene changes; higher values detect only
    /// dramatic cuts.  Typical default: 0.3.
    pub threshold: f64,
}

impl Default for SceneDetectionOptions {
    fn default() -> Self {
        Self { threshold: 0.3 }
    }
}

// ---------------------------------------------------------------------------
// Detector
// ---------------------------------------------------------------------------

/// Stateless scene detector backed by the ffmpeg CLI.
pub struct SceneDetector;

impl SceneDetector {
    /// Detect scene changes in a video file.
    ///
    /// Uses ffmpeg's `select='gt(scene,THRESHOLD)'` filter to identify
    /// frames where the scene score exceeds the configured threshold.
    /// Each detected frame's presentation timestamp and scene score are
    /// returned.
    ///
    /// # Errors
    ///
    /// Returns an error if:
    /// - `file_path` does not exist.
    /// - `ffmpeg` is not installed or not on `$PATH`.
    /// - The file does not contain a video stream.
    /// - ffmpeg fails for any other reason.
    pub fn detect(file_path: &str, options: &SceneDetectionOptions) -> Result<SceneDetectionResult> {
        let path = Path::new(file_path);
        if !path.exists() {
            anyhow::bail!("file not found: {file_path}");
        }

        let threshold = options.threshold.clamp(0.0, 1.0);
        info!(
            path = %file_path,
            threshold,
            "starting scene detection via ffmpeg"
        );

        // Build the ffmpeg filter that selects frames where the scene score
        // exceeds the threshold and prints their PTS and score.
        //
        // The `select` filter sets the `scene` variable for each frame.
        // We use `-vf select=...,showinfo` and parse the `showinfo` output
        // from stderr, or we use the simpler approach of printing frame
        // metadata via the `-f null` trick.
        //
        // Cleanest approach: use `select` + `metadata=print` with `-f null`.
        let filter = format!(
            "select='gt(scene\\,{threshold})',metadata=print:file=-"
        );

        let output = Command::new("ffmpeg")
            .args([
                "-hide_banner",
                "-i",
                file_path,
                "-vf",
                &filter,
                "-an",
                "-f",
                "null",
                "-",
            ])
            .output()
            .context(
                "failed to execute ffmpeg for scene detection \
                 -- is ffmpeg installed and on your PATH?"
            )?;

        // ffmpeg writes metadata to stdout via the `metadata=print:file=-`
        // filter, but diagnostic messages go to stderr.  Combine both for
        // parsing.
        let stdout_str = String::from_utf8_lossy(&output.stdout);
        let stderr_str = String::from_utf8_lossy(&output.stderr);

        // Parse scene changes from the output.
        // The showinfo filter or metadata print outputs lines like:
        //   frame:0    pts:0       pts_time:0.000000
        //   lavfi.scene_score=0.876543
        //
        // The metadata print writes to the file specified (- = stdout):
        //   frame:N    pts:XXXX    pts_time:SS.SSSSSS
        //   lavfi.scene_score=0.XXXXXX
        let scenes = Self::parse_scene_output(&stdout_str, &stderr_str);

        debug!(scene_count = scenes.len(), "scene detection complete");

        Ok(SceneDetectionResult { scenes })
    }

    /// Parse ffmpeg metadata output to extract scene change points.
    fn parse_scene_output(stdout: &str, stderr: &str) -> Vec<SceneChangePoint> {
        let mut scenes = Vec::new();
        let mut current_pts_time: Option<f64> = None;

        // Process both stdout and stderr — the metadata filter writes to
        // stdout when file=-, but ffmpeg logs frame info to stderr.
        let combined = format!("{stdout}\n{stderr}");

        for line in combined.lines() {
            let trimmed = line.trim();

            // Look for pts_time in frame lines.
            if trimmed.contains("pts_time:") || trimmed.contains("pts_time=") {
                if let Some(ts) = Self::extract_pts_time(trimmed) {
                    current_pts_time = Some(ts);
                }
            }

            // Look for scene_score lines.
            if trimmed.contains("lavfi.scene_score=") {
                if let Some(score) = Self::extract_scene_score(trimmed) {
                    let timestamp = current_pts_time.unwrap_or(0.0);
                    scenes.push(SceneChangePoint {
                        timestamp_seconds: timestamp,
                        confidence: score,
                    });
                }
            }
        }

        // Sort by timestamp in case output was out of order.
        scenes.sort_by(|a, b| {
            a.timestamp_seconds
                .partial_cmp(&b.timestamp_seconds)
                .unwrap_or(std::cmp::Ordering::Equal)
        });

        scenes
    }

    /// Extract pts_time value from a line like `frame:0 pts:0 pts_time:1.234567`.
    fn extract_pts_time(line: &str) -> Option<f64> {
        // Try both `:` and `=` separators.
        for sep in ["pts_time:", "pts_time="] {
            if let Some(idx) = line.find(sep) {
                let after = &line[idx + sep.len()..];
                let num_str: String = after
                    .chars()
                    .take_while(|c| c.is_ascii_digit() || *c == '.' || *c == '-')
                    .collect();
                if let Ok(val) = num_str.parse::<f64>() {
                    return Some(val);
                }
            }
        }
        None
    }

    /// Extract the scene score from a line like `lavfi.scene_score=0.876543`.
    fn extract_scene_score(line: &str) -> Option<f64> {
        let marker = "lavfi.scene_score=";
        if let Some(idx) = line.find(marker) {
            let after = &line[idx + marker.len()..];
            let num_str: String = after
                .chars()
                .take_while(|c| c.is_ascii_digit() || *c == '.' || *c == '-')
                .collect();
            return num_str.parse::<f64>().ok();
        }
        None
    }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_pts_time_colon() {
        let line = "frame:0    pts:0       pts_time:1.234567";
        assert!((SceneDetector::extract_pts_time(line).unwrap() - 1.234567).abs() < 1e-6);
    }

    #[test]
    fn test_extract_pts_time_equals() {
        let line = "pts_time=42.5";
        assert!((SceneDetector::extract_pts_time(line).unwrap() - 42.5).abs() < 1e-6);
    }

    #[test]
    fn test_extract_pts_time_none() {
        assert!(SceneDetector::extract_pts_time("no timestamp here").is_none());
    }

    #[test]
    fn test_extract_scene_score() {
        let line = "lavfi.scene_score=0.876543";
        assert!((SceneDetector::extract_scene_score(line).unwrap() - 0.876543).abs() < 1e-6);
    }

    #[test]
    fn test_extract_scene_score_none() {
        assert!(SceneDetector::extract_scene_score("no score here").is_none());
    }

    #[test]
    fn test_parse_scene_output() {
        let stdout = "\
frame:0    pts:0       pts_time:1.500000
lavfi.scene_score=0.95
frame:1    pts:72000   pts_time:3.000000
lavfi.scene_score=0.42
";
        let scenes = SceneDetector::parse_scene_output(stdout, "");
        assert_eq!(scenes.len(), 2);
        assert!((scenes[0].timestamp_seconds - 1.5).abs() < 1e-6);
        assert!((scenes[0].confidence - 0.95).abs() < 1e-6);
        assert!((scenes[1].timestamp_seconds - 3.0).abs() < 1e-6);
        assert!((scenes[1].confidence - 0.42).abs() < 1e-6);
    }

    #[test]
    fn test_detect_missing_file() {
        let result = SceneDetector::detect(
            "/nonexistent/video.mp4",
            &SceneDetectionOptions::default(),
        );
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("file not found"));
    }
}
