//! Media probing and metadata extraction.
//!
//! This module is the core of the Rust media engine.  It inspects media files
//! (video, audio, images) and extracts detailed metadata such as codec,
//! resolution, duration, frame rate, audio channels, MIME type, and embedded
//! tags.
//!
//! # Architecture
//!
//! * [`formats`] — MIME type detection (magic bytes + extension) and lists of
//!   supported media extensions.
//! * [`probe`] — The [`MediaProber`] struct which shells out to `ffprobe` and
//!   parses its JSON output into strongly-typed Rust structs.
//!
//! # Quick start
//!
//! ```rust,no_run
//! use premierpro_media_engine::media::probe::MediaProber;
//!
//! let info = MediaProber::probe("/path/to/clip.mp4").unwrap();
//! println!("{}x{}", info.video.as_ref().unwrap().width,
//!                    info.video.as_ref().unwrap().height);
//! ```

pub mod formats;
pub mod probe;

// Re-export the most commonly used types at the module level for convenience.
pub use formats::{detect_mime_type, is_media_file};
pub use probe::{AssetType, AudioInfo, MediaInfo, MediaProber, VideoInfo};
