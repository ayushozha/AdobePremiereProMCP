//! Thumbnail generation via ffmpeg CLI.
//!
//! Extracts a single video frame at a given timestamp and encodes it as PNG
//! or JPEG.  Supports both in-memory output (raw bytes returned to the
//! caller) and writing directly to a file on disk.

use anyhow::{Context, Result};
use std::fs;
use std::process::{Command, Stdio};
use tracing::{debug, info};

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

/// Options that control thumbnail extraction.
#[derive(Debug, Clone)]
pub struct ThumbnailOptions {
    /// Timestamp (in seconds) at which to capture the frame.
    pub timestamp_seconds: f64,
    /// Desired thumbnail width in pixels.
    pub width: u32,
    /// Desired thumbnail height in pixels.
    pub height: u32,
    /// Output image format — `"png"` or `"jpg"` (aliases: `"jpeg"`).
    pub output_format: String,
}

impl Default for ThumbnailOptions {
    fn default() -> Self {
        Self {
            timestamp_seconds: 0.0,
            width: 320,
            height: 180,
            output_format: "png".into(),
        }
    }
}

// ---------------------------------------------------------------------------
// Result
// ---------------------------------------------------------------------------

/// Result of a thumbnail generation pass.
#[derive(Debug, Clone)]
pub struct ThumbnailResult {
    /// Raw image bytes (PNG or JPEG encoded).
    pub data: Vec<u8>,
    /// If the caller requested file output, this is the path that was
    /// written.  `None` for in-memory–only results.
    pub output_path: Option<String>,
    /// Actual pixel width of the generated thumbnail.
    pub actual_width: u32,
    /// Actual pixel height of the generated thumbnail.
    pub actual_height: u32,
}

// ---------------------------------------------------------------------------
// Generator
// ---------------------------------------------------------------------------

/// Stateless thumbnail generator backed by the ffmpeg CLI.
pub struct ThumbnailGenerator;

impl ThumbnailGenerator {
    /// Generate a thumbnail from a video file and return the image bytes.
    ///
    /// The frame is captured by seeking to `options.timestamp_seconds` and
    /// then encoding a single frame at the requested resolution.
    ///
    /// # Errors
    ///
    /// Returns an error if ffmpeg is not found, the input file cannot be
    /// read, or no video stream is present.
    pub fn generate(file_path: &str, options: &ThumbnailOptions) -> Result<ThumbnailResult> {
        info!(
            path = %file_path,
            timestamp = options.timestamp_seconds,
            size = format!("{}x{}", options.width, options.height),
            format = %options.output_format,
            "generating thumbnail"
        );

        let data = Self::extract_frame(file_path, options)
            .context("failed to extract video frame via ffmpeg")?;

        if data.is_empty() {
            anyhow::bail!(
                "ffmpeg produced zero bytes — does the file contain a video stream?"
            );
        }

        debug!(bytes = data.len(), "thumbnail captured in memory");

        Ok(ThumbnailResult {
            data,
            output_path: None,
            actual_width: options.width,
            actual_height: options.height,
        })
    }

    /// Generate a thumbnail and write it to `output_path`.
    ///
    /// This is a convenience wrapper around [`Self::generate`] that
    /// additionally persists the image bytes to disk.
    ///
    /// # Errors
    ///
    /// Propagates errors from [`Self::generate`] and from filesystem I/O.
    pub fn generate_to_file(
        file_path: &str,
        options: &ThumbnailOptions,
        output_path: &str,
    ) -> Result<ThumbnailResult> {
        let mut result = Self::generate(file_path, options)?;

        // Ensure parent directory exists.
        if let Some(parent) = std::path::Path::new(output_path).parent() {
            fs::create_dir_all(parent)
                .with_context(|| format!("failed to create directory: {}", parent.display()))?;
        }

        fs::write(output_path, &result.data)
            .with_context(|| format!("failed to write thumbnail to {output_path}"))?;

        info!(path = %output_path, bytes = result.data.len(), "thumbnail written to disk");
        result.output_path = Some(output_path.to_string());
        Ok(result)
    }

    // -----------------------------------------------------------------------
    // Internal
    // -----------------------------------------------------------------------

    /// Invoke ffmpeg to extract a single frame as PNG or JPEG via stdout
    /// pipe.
    fn extract_frame(file_path: &str, options: &ThumbnailOptions) -> Result<Vec<u8>> {
        let timestamp = format!("{:.3}", options.timestamp_seconds);
        let size = format!("{}x{}", options.width, options.height);
        let (ffmpeg_format, codec) = Self::format_args(&options.output_format);

        let child = Command::new("ffmpeg")
            .args([
                "-hide_banner",
                "-loglevel",
                "error",
                // Seek *before* input for fast seeking.
                "-ss",
                &timestamp,
                "-i",
                file_path,
                "-vframes",
                "1",
                "-s",
                &size,
                "-f",
                ffmpeg_format,
                "-c:v",
                codec,
                // Disable audio.
                "-an",
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

        Ok(output.stdout)
    }

    /// Map the user-facing format string to ffmpeg `-f` and `-c:v` values.
    fn format_args(format: &str) -> (&'static str, &'static str) {
        match format.to_lowercase().as_str() {
            "jpg" | "jpeg" => ("image2pipe", "mjpeg"),
            // Default to PNG for everything else.
            _ => ("image2pipe", "png"),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn format_args_png() {
        let (fmt, codec) = ThumbnailGenerator::format_args("png");
        assert_eq!(fmt, "image2pipe");
        assert_eq!(codec, "png");
    }

    #[test]
    fn format_args_jpeg_aliases() {
        for alias in &["jpg", "jpeg", "JPG", "JPEG"] {
            let (fmt, codec) = ThumbnailGenerator::format_args(alias);
            assert_eq!(fmt, "image2pipe");
            assert_eq!(codec, "mjpeg");
        }
    }

    #[test]
    fn default_options() {
        let opts = ThumbnailOptions::default();
        assert_eq!(opts.timestamp_seconds, 0.0);
        assert_eq!(opts.width, 320);
        assert_eq!(opts.height, 180);
        assert_eq!(opts.output_format, "png");
    }
}
