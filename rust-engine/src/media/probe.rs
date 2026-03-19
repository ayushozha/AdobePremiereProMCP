//! Media file probing via `ffprobe`.
//!
//! [`MediaProber`] shells out to the `ffprobe` CLI (part of FFmpeg) and parses
//! the JSON output to build a [`MediaInfo`] descriptor.  This avoids any native
//! FFmpeg bindings while still providing rich, reliable metadata extraction.

use std::collections::HashMap;
use std::path::Path;
use std::process::Command;

use anyhow::{Context, Result, bail};
use serde::Deserialize;

use crate::media::formats;

// ============================================================================
// Public types
// ============================================================================

/// The kind of media asset.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum AssetType {
    Video,
    Audio,
    Image,
    Graphics,
    Unknown,
}

impl std::fmt::Display for AssetType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Video => write!(f, "Video"),
            Self::Audio => write!(f, "Audio"),
            Self::Image => write!(f, "Image"),
            Self::Graphics => write!(f, "Graphics"),
            Self::Unknown => write!(f, "Unknown"),
        }
    }
}

/// Metadata about a video stream.
#[derive(Debug, Clone)]
pub struct VideoInfo {
    /// Codec name (e.g. `"h264"`, `"hevc"`, `"prores"`).
    pub codec: String,
    /// Frame width in pixels.
    pub width: u32,
    /// Frame height in pixels.
    pub height: u32,
    /// Frames per second (e.g. `23.976`, `29.97`, `60.0`).
    pub frame_rate: f64,
    /// Bitrate in bits per second.
    pub bitrate_bps: u64,
    /// Pixel format string (e.g. `"yuv420p"`).
    pub pixel_format: String,
    /// Stream duration in seconds.
    pub duration_seconds: f64,
}

/// Metadata about an audio stream.
#[derive(Debug, Clone)]
pub struct AudioInfo {
    /// Codec name (e.g. `"aac"`, `"pcm_s16le"`).
    pub codec: String,
    /// Sample rate in Hz (e.g. `44100`, `48000`).
    pub sample_rate: u32,
    /// Number of audio channels.
    pub channels: u32,
    /// Bitrate in bits per second.
    pub bitrate_bps: u64,
    /// Stream duration in seconds.
    pub duration_seconds: f64,
}

/// Complete metadata for a probed media file.
#[derive(Debug, Clone)]
pub struct MediaInfo {
    /// Absolute or as-supplied path to the file.
    pub file_path: String,
    /// Base file name (including extension).
    pub file_name: String,
    /// File size in bytes.
    pub file_size: u64,
    /// Detected MIME type.
    pub mime_type: String,
    /// High-level asset classification.
    pub asset_type: AssetType,
    /// First video stream, if present.
    pub video: Option<VideoInfo>,
    /// First audio stream, if present.
    pub audio: Option<AudioInfo>,
    /// Overall duration in seconds (from the container format).
    pub duration_seconds: f64,
    /// Arbitrary key/value metadata tags embedded in the file.
    pub metadata: HashMap<String, String>,
}

// ============================================================================
// MediaProber
// ============================================================================

/// Zero-sized prober that delegates to `ffprobe` for metadata extraction.
pub struct MediaProber;

impl MediaProber {
    /// Probe a media file and return its metadata.
    ///
    /// # Errors
    ///
    /// Returns an error when:
    /// - `file_path` does not exist or is not readable.
    /// - `ffprobe` is not installed / not on `$PATH`.
    /// - `ffprobe` exits with a non-zero status (e.g. corrupt file).
    /// - The JSON output cannot be parsed.
    pub fn probe(file_path: &str) -> Result<MediaInfo> {
        // -----------------------------------------------------------------
        // 1. Validate the path
        // -----------------------------------------------------------------
        let path = Path::new(file_path);
        if !path.exists() {
            bail!("file not found: {file_path}");
        }

        let file_name = path
            .file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("unknown")
            .to_string();

        let file_size = std::fs::metadata(path)
            .with_context(|| format!("cannot stat file: {file_path}"))?
            .len();

        // -----------------------------------------------------------------
        // 2. Run ffprobe
        // -----------------------------------------------------------------
        let output = Command::new("ffprobe")
            .args([
                "-v", "quiet",
                "-print_format", "json",
                "-show_format",
                "-show_streams",
                file_path,
            ])
            .output()
            .context(
                "failed to execute ffprobe — is FFmpeg installed and on your PATH?",
            )?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            bail!(
                "ffprobe exited with status {}: {}",
                output.status,
                stderr.trim()
            );
        }

        let json_str = String::from_utf8(output.stdout)
            .context("ffprobe produced non-UTF-8 output")?;

        // -----------------------------------------------------------------
        // 3. Parse JSON
        // -----------------------------------------------------------------
        let raw: FfprobeOutput = serde_json::from_str(&json_str)
            .context("failed to parse ffprobe JSON output")?;

        // -----------------------------------------------------------------
        // 4. Extract video / audio info from streams
        // -----------------------------------------------------------------
        let video = Self::extract_video(&raw);
        let audio = Self::extract_audio(&raw);

        // -----------------------------------------------------------------
        // 5. Overall duration
        // -----------------------------------------------------------------
        let duration_seconds = raw
            .format
            .as_ref()
            .and_then(|f| f.duration.as_deref())
            .and_then(|d| d.parse::<f64>().ok())
            .or_else(|| video.as_ref().map(|v| v.duration_seconds))
            .or_else(|| audio.as_ref().map(|a| a.duration_seconds))
            .unwrap_or(0.0);

        // -----------------------------------------------------------------
        // 6. MIME type
        // -----------------------------------------------------------------
        let mime_type = formats::detect_mime_type(file_path);

        // -----------------------------------------------------------------
        // 7. Asset type
        // -----------------------------------------------------------------
        let asset_type = Self::classify_asset(&raw, &mime_type);

        // -----------------------------------------------------------------
        // 8. Metadata tags
        // -----------------------------------------------------------------
        let metadata = Self::collect_tags(&raw);

        Ok(MediaInfo {
            file_path: file_path.to_string(),
            file_name,
            file_size,
            mime_type,
            asset_type,
            video,
            audio,
            duration_seconds,
            metadata,
        })
    }

    // ------------------------------------------------------------------
    // Internal helpers
    // ------------------------------------------------------------------

    /// Find the first video stream and convert it to [`VideoInfo`].
    fn extract_video(raw: &FfprobeOutput) -> Option<VideoInfo> {
        let streams = raw.streams.as_deref()?;
        let s = streams
            .iter()
            .find(|s| s.codec_type.as_deref() == Some("video"))?;

        let frame_rate = parse_rational(s.r_frame_rate.as_deref().unwrap_or("0/1"));

        Some(VideoInfo {
            codec: s.codec_name.clone().unwrap_or_default(),
            width: s.width.unwrap_or(0),
            height: s.height.unwrap_or(0),
            frame_rate,
            bitrate_bps: s
                .bit_rate
                .as_deref()
                .and_then(|b| b.parse().ok())
                .unwrap_or(0),
            pixel_format: s.pix_fmt.clone().unwrap_or_default(),
            duration_seconds: s
                .duration
                .as_deref()
                .and_then(|d| d.parse().ok())
                .unwrap_or(0.0),
        })
    }

    /// Find the first audio stream and convert it to [`AudioInfo`].
    fn extract_audio(raw: &FfprobeOutput) -> Option<AudioInfo> {
        let streams = raw.streams.as_deref()?;
        let s = streams
            .iter()
            .find(|s| s.codec_type.as_deref() == Some("audio"))?;

        Some(AudioInfo {
            codec: s.codec_name.clone().unwrap_or_default(),
            sample_rate: s
                .sample_rate
                .as_deref()
                .and_then(|r| r.parse().ok())
                .unwrap_or(0),
            channels: s.channels.unwrap_or(0),
            bitrate_bps: s
                .bit_rate
                .as_deref()
                .and_then(|b| b.parse().ok())
                .unwrap_or(0),
            duration_seconds: s
                .duration
                .as_deref()
                .and_then(|d| d.parse().ok())
                .unwrap_or(0.0),
        })
    }

    /// Determine the high-level [`AssetType`] from the stream layout.
    fn classify_asset(raw: &FfprobeOutput, mime_type: &str) -> AssetType {
        if mime_type.starts_with("image/") {
            return AssetType::Image;
        }

        let streams = match raw.streams.as_deref() {
            Some(s) => s,
            None => return AssetType::Unknown,
        };

        let has_video = streams
            .iter()
            .any(|s| s.codec_type.as_deref() == Some("video"));
        let has_audio = streams
            .iter()
            .any(|s| s.codec_type.as_deref() == Some("audio"));

        // Some image codecs (e.g. mjpeg, png) can appear as video streams in
        // containers — check for a single-frame "video" stream.
        if has_video {
            let is_image_codec = streams.iter().any(|s| {
                matches!(
                    s.codec_name.as_deref(),
                    Some("mjpeg" | "png" | "bmp" | "gif" | "tiff" | "webp")
                )
            });
            if is_image_codec && !has_audio {
                return AssetType::Image;
            }
            return AssetType::Video;
        }

        if has_audio {
            return AssetType::Audio;
        }

        AssetType::Unknown
    }

    /// Merge all tag dictionaries (format-level and per-stream) into one map.
    fn collect_tags(raw: &FfprobeOutput) -> HashMap<String, String> {
        let mut map = HashMap::new();

        // Format-level tags.
        if let Some(fmt) = &raw.format {
            if let Some(tags) = &fmt.tags {
                for (k, v) in tags {
                    map.insert(k.clone(), v.clone());
                }
            }
        }

        // Per-stream tags (prefixed with the stream index to avoid collisions).
        if let Some(streams) = &raw.streams {
            for (i, s) in streams.iter().enumerate() {
                if let Some(tags) = &s.tags {
                    for (k, v) in tags {
                        map.insert(format!("stream_{i}:{k}"), v.clone());
                    }
                }
            }
        }

        map
    }
}

// ============================================================================
// ffprobe JSON schema (subset we care about)
// ============================================================================

#[derive(Debug, Deserialize)]
struct FfprobeOutput {
    streams: Option<Vec<FfprobeStream>>,
    format: Option<FfprobeFormat>,
}

#[derive(Debug, Deserialize)]
struct FfprobeStream {
    codec_name: Option<String>,
    codec_type: Option<String>,
    width: Option<u32>,
    height: Option<u32>,
    pix_fmt: Option<String>,
    r_frame_rate: Option<String>,
    sample_rate: Option<String>,
    channels: Option<u32>,
    bit_rate: Option<String>,
    duration: Option<String>,
    tags: Option<HashMap<String, String>>,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct FfprobeFormat {
    duration: Option<String>,
    bit_rate: Option<String>,
    format_name: Option<String>,
    tags: Option<HashMap<String, String>>,
}

// ============================================================================
// Utility
// ============================================================================

/// Parse a rational number string like `"30000/1001"` into a floating-point
/// value.  Returns `0.0` on any parse failure.
fn parse_rational(s: &str) -> f64 {
    let parts: Vec<&str> = s.split('/').collect();
    if parts.len() == 2 {
        let num: f64 = parts[0].parse().unwrap_or(0.0);
        let den: f64 = parts[1].parse().unwrap_or(1.0);
        if den != 0.0 {
            return num / den;
        }
    }
    s.parse().unwrap_or(0.0)
}

// ============================================================================
// Tests
// ============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_rational_fraction() {
        let fps = parse_rational("30000/1001");
        assert!((fps - 29.97).abs() < 0.01);
    }

    #[test]
    fn test_parse_rational_integer() {
        assert!((parse_rational("25/1") - 25.0).abs() < f64::EPSILON);
    }

    #[test]
    fn test_parse_rational_plain_number() {
        assert!((parse_rational("24") - 24.0).abs() < f64::EPSILON);
    }

    #[test]
    fn test_parse_rational_zero_denominator() {
        assert!((parse_rational("30/0") - 0.0).abs() < f64::EPSILON);
    }

    #[test]
    fn test_probe_missing_file() {
        let result = MediaProber::probe("/nonexistent/path/video.mp4");
        assert!(result.is_err());
        let msg = result.unwrap_err().to_string();
        assert!(
            msg.contains("file not found"),
            "expected 'file not found' in error message, got: {msg}"
        );
    }

    #[test]
    fn test_classify_asset_with_image_mime() {
        let raw = FfprobeOutput {
            streams: None,
            format: None,
        };
        assert_eq!(
            MediaProber::classify_asset(&raw, "image/png"),
            AssetType::Image
        );
    }

    #[test]
    fn test_classify_asset_video_and_audio() {
        let raw = FfprobeOutput {
            streams: Some(vec![
                FfprobeStream {
                    codec_name: Some("h264".into()),
                    codec_type: Some("video".into()),
                    width: Some(1920),
                    height: Some(1080),
                    pix_fmt: None,
                    r_frame_rate: None,
                    sample_rate: None,
                    channels: None,
                    bit_rate: None,
                    duration: None,
                    tags: None,
                },
                FfprobeStream {
                    codec_name: Some("aac".into()),
                    codec_type: Some("audio".into()),
                    width: None,
                    height: None,
                    pix_fmt: None,
                    r_frame_rate: None,
                    sample_rate: Some("48000".into()),
                    channels: Some(2),
                    bit_rate: None,
                    duration: None,
                    tags: None,
                },
            ]),
            format: None,
        };
        assert_eq!(
            MediaProber::classify_asset(&raw, "video/mp4"),
            AssetType::Video
        );
    }

    #[test]
    fn test_classify_asset_audio_only() {
        let raw = FfprobeOutput {
            streams: Some(vec![FfprobeStream {
                codec_name: Some("aac".into()),
                codec_type: Some("audio".into()),
                width: None,
                height: None,
                pix_fmt: None,
                r_frame_rate: None,
                sample_rate: Some("44100".into()),
                channels: Some(2),
                bit_rate: None,
                duration: None,
                tags: None,
            }]),
            format: None,
        };
        assert_eq!(
            MediaProber::classify_asset(&raw, "audio/mpeg"),
            AssetType::Audio
        );
    }
}
