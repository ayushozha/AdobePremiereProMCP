//! Asset indexing — directory scanning and file fingerprinting.
//!
//! This module provides:
//!
//! - [`scanner::AssetScanner`] — recursive directory walker that discovers
//!   media files, collects metadata, and assigns unique IDs.
//! - [`fingerprint::compute_fingerprint`] — fast partial SHA-256 hashing for
//!   deduplication.

pub mod fingerprint;
pub mod scanner;

// Re-export the primary public types for convenience.
pub use fingerprint::compute_fingerprint;
pub use scanner::{AssetScanner, ScanOptions, ScanResult, ScannedAsset};
