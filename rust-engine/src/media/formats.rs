//! MIME type detection and media format recognition.
//!
//! Provides utilities for identifying media files by extension and magic bytes,
//! and defines the set of formats supported by the media engine.

use std::fs::File;
use std::io::Read;

use anyhow::Result;

// ---------------------------------------------------------------------------
// Supported extension tables
// ---------------------------------------------------------------------------

/// Video container extensions recognised by the engine.
pub const SUPPORTED_VIDEO_EXTENSIONS: &[&str] = &[
    "mp4", "mov", "avi", "mkv", "wmv", "flv", "webm", "m4v", "mpg", "mpeg",
    "3gp", "ts", "mts", "m2ts", "vob", "ogv", "mxf", "prores",
];

/// Audio extensions recognised by the engine.
pub const SUPPORTED_AUDIO_EXTENSIONS: &[&str] = &[
    "mp3", "wav", "aac", "flac", "ogg", "wma", "m4a", "aiff", "aif", "opus",
    "ac3", "dts", "pcm",
];

/// Image extensions recognised by the engine.
pub const SUPPORTED_IMAGE_EXTENSIONS: &[&str] = &[
    "jpg", "jpeg", "png", "bmp", "tiff", "tif", "gif", "webp", "exr", "dpx",
    "tga", "psd", "svg", "ico", "heic", "heif", "avif",
];

// ---------------------------------------------------------------------------
// Public helpers
// ---------------------------------------------------------------------------

/// Check whether `file_path` has a recognised media extension.
pub fn is_media_file(file_path: &str) -> bool {
    let ext = match extension_lower(file_path) {
        Some(e) => e,
        None => return false,
    };
    SUPPORTED_VIDEO_EXTENSIONS.contains(&ext.as_str())
        || SUPPORTED_AUDIO_EXTENSIONS.contains(&ext.as_str())
        || SUPPORTED_IMAGE_EXTENSIONS.contains(&ext.as_str())
}

/// Detect the MIME type for `file_path`.
///
/// The function first tries magic-byte identification (reading the first few
/// bytes of the file). If that is inconclusive it falls back to the file
/// extension.  Returns `"application/octet-stream"` when nothing matches.
pub fn detect_mime_type(file_path: &str) -> String {
    // Try magic bytes first — they are more reliable than extensions.
    if let Ok(mime) = detect_from_magic_bytes(file_path) {
        return mime;
    }

    // Fall back to extension-based detection.
    detect_from_extension(file_path)
}

// ---------------------------------------------------------------------------
// Extension-based detection
// ---------------------------------------------------------------------------

/// Derive a MIME type purely from the file extension.
pub fn detect_from_extension(file_path: &str) -> String {
    let ext = match extension_lower(file_path) {
        Some(e) => e,
        None => return "application/octet-stream".to_string(),
    };

    match ext.as_str() {
        // Video
        "mp4" | "m4v" => "video/mp4",
        "mov" => "video/quicktime",
        "avi" => "video/x-msvideo",
        "mkv" => "video/x-matroska",
        "wmv" => "video/x-ms-wmv",
        "flv" => "video/x-flv",
        "webm" => "video/webm",
        "mpg" | "mpeg" => "video/mpeg",
        "3gp" => "video/3gpp",
        "ts" | "mts" | "m2ts" => "video/mp2t",
        "vob" => "video/dvd",
        "ogv" => "video/ogg",
        "mxf" => "application/mxf",

        // Audio
        "mp3" => "audio/mpeg",
        "wav" => "audio/wav",
        "aac" => "audio/aac",
        "flac" => "audio/flac",
        "ogg" | "opus" => "audio/ogg",
        "wma" => "audio/x-ms-wma",
        "m4a" => "audio/mp4",
        "aiff" | "aif" => "audio/aiff",
        "ac3" => "audio/ac3",

        // Image
        "jpg" | "jpeg" => "image/jpeg",
        "png" => "image/png",
        "bmp" => "image/bmp",
        "tiff" | "tif" => "image/tiff",
        "gif" => "image/gif",
        "webp" => "image/webp",
        "svg" => "image/svg+xml",
        "ico" => "image/x-icon",
        "heic" | "heif" => "image/heif",
        "avif" => "image/avif",
        "psd" => "image/vnd.adobe.photoshop",
        "exr" => "image/x-exr",
        "dpx" => "image/x-dpx",
        "tga" => "image/x-tga",

        _ => "application/octet-stream",
    }
    .to_string()
}

// ---------------------------------------------------------------------------
// Magic-byte detection
// ---------------------------------------------------------------------------

/// Attempt to identify the MIME type from the first bytes of the file.
fn detect_from_magic_bytes(file_path: &str) -> Result<String> {
    let mut file = File::open(file_path)?;
    let mut buf = [0u8; 32];
    let n = file.read(&mut buf)?;
    let bytes = &buf[..n];

    // PNG: 89 50 4E 47
    if bytes.starts_with(&[0x89, 0x50, 0x4E, 0x47]) {
        return Ok("image/png".into());
    }
    // JPEG: FF D8 FF
    if bytes.starts_with(&[0xFF, 0xD8, 0xFF]) {
        return Ok("image/jpeg".into());
    }
    // GIF: GIF87a / GIF89a
    if bytes.starts_with(b"GIF87a") || bytes.starts_with(b"GIF89a") {
        return Ok("image/gif".into());
    }
    // BMP: BM
    if bytes.starts_with(b"BM") {
        return Ok("image/bmp".into());
    }
    // WEBP: RIFF....WEBP
    if n >= 12 && bytes.starts_with(b"RIFF") && &bytes[8..12] == b"WEBP" {
        return Ok("image/webp".into());
    }
    // TIFF: II (little-endian) or MM (big-endian)
    if bytes.starts_with(&[0x49, 0x49, 0x2A, 0x00])
        || bytes.starts_with(&[0x4D, 0x4D, 0x00, 0x2A])
    {
        return Ok("image/tiff".into());
    }
    // WAV: RIFF....WAVE
    if n >= 12 && bytes.starts_with(b"RIFF") && &bytes[8..12] == b"WAVE" {
        return Ok("audio/wav".into());
    }
    // AVI: RIFF....AVI
    if n >= 12 && bytes.starts_with(b"RIFF") && &bytes[8..12] == b"AVI " {
        return Ok("video/x-msvideo".into());
    }
    // MP4 / MOV / M4A / M4V — ftyp box
    if n >= 8 && &bytes[4..8] == b"ftyp" {
        // Further discrimination based on the brand.
        if n >= 12 {
            let brand = &bytes[8..12];
            if brand == b"M4A " || brand == b"M4B " {
                return Ok("audio/mp4".into());
            }
            if brand == b"qt  " {
                return Ok("video/quicktime".into());
            }
        }
        return Ok("video/mp4".into());
    }
    // Matroska / WebM: EBML header 1A 45 DF A3
    if bytes.starts_with(&[0x1A, 0x45, 0xDF, 0xA3]) {
        // Could be MKV or WebM — fall back to extension for distinction.
        let ext = extension_lower(file_path).unwrap_or_default();
        return Ok(match ext.as_str() {
            "webm" => "video/webm",
            _ => "video/x-matroska",
        }
        .into());
    }
    // FLAC: fLaC
    if bytes.starts_with(b"fLaC") {
        return Ok("audio/flac".into());
    }
    // OGG: OggS
    if bytes.starts_with(b"OggS") {
        return Ok("audio/ogg".into());
    }
    // MP3: ID3 tag or sync word FF FB / FF F3 / FF F2
    if bytes.starts_with(b"ID3")
        || (n >= 2 && bytes[0] == 0xFF && (bytes[1] & 0xE0) == 0xE0)
    {
        return Ok("audio/mpeg".into());
    }
    // AIFF: FORM....AIFF
    if n >= 12 && bytes.starts_with(b"FORM") && &bytes[8..12] == b"AIFF" {
        return Ok("audio/aiff".into());
    }
    // FLV: FLV
    if bytes.starts_with(b"FLV") {
        return Ok("video/x-flv".into());
    }

    anyhow::bail!("unrecognised magic bytes")
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

/// Extract the lowercase extension from a file path.
fn extension_lower(path: &str) -> Option<String> {
    std::path::Path::new(path)
        .extension()
        .and_then(|e| e.to_str())
        .map(|e| e.to_ascii_lowercase())
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_media_file() {
        assert!(is_media_file("clip.mp4"));
        assert!(is_media_file("/path/to/song.MP3"));
        assert!(is_media_file("photo.JPEG"));
        assert!(!is_media_file("document.pdf"));
        assert!(!is_media_file("noext"));
    }

    #[test]
    fn test_detect_from_extension() {
        assert_eq!(detect_from_extension("test.mp4"), "video/mp4");
        assert_eq!(detect_from_extension("test.MP3"), "audio/mpeg");
        assert_eq!(detect_from_extension("test.png"), "image/png");
        assert_eq!(detect_from_extension("test.xyz"), "application/octet-stream");
    }
}
