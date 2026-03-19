//! Thumbnail generation — extract single frames from video files.
//!
//! The primary entry point is [`generator::ThumbnailGenerator`], which uses
//! ffmpeg to seek to a given timestamp and encode a scaled frame as PNG or
//! JPEG.  Results can be kept in memory or written directly to disk.

pub mod generator;

// Re-export the main public types for convenience.
pub use generator::{ThumbnailGenerator, ThumbnailOptions, ThumbnailResult};
