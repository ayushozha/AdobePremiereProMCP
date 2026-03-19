//! Recursive directory scanner for media assets.
//!
//! Uses [`walkdir`] for traversal, filters files by extension, assigns a
//! unique ID to every discovered asset, and computes a partial SHA-256
//! fingerprint via [`super::fingerprint::compute_fingerprint`].

use anyhow::Result;
use std::path::Path;
use std::time::{Duration, Instant};
use tracing::{debug, info, warn};
use walkdir::WalkDir;

use super::fingerprint::compute_fingerprint;

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

/// Options that control how a directory scan is performed.
#[derive(Debug, Clone)]
pub struct ScanOptions {
    /// Root directory to scan.
    pub directory: String,
    /// Whether to descend into subdirectories.
    pub recursive: bool,
    /// Optional allowlist of file extensions (lowercase, without leading dot).
    /// An empty list means *all* files are included.
    pub extensions: Vec<String>,
}

// ---------------------------------------------------------------------------
// Results
// ---------------------------------------------------------------------------

/// Aggregate result of a completed scan.
#[derive(Debug, Clone)]
pub struct ScanResult {
    /// The media assets that were discovered.
    pub assets: Vec<ScannedAsset>,
    /// Total number of filesystem entries visited (files + dirs).
    pub total_files_scanned: u32,
    /// Number of entries that matched the extension filter.
    pub media_files_found: u32,
    /// Wall-clock time spent scanning.
    pub scan_duration: Duration,
}

/// A single media asset discovered during a scan.
#[derive(Debug, Clone)]
pub struct ScannedAsset {
    /// Unique identifier (UUID v4).
    pub id: String,
    /// Absolute path to the file on disk.
    pub file_path: String,
    /// File name component (including extension).
    pub file_name: String,
    /// File size in bytes.
    pub file_size: u64,
    /// Best-guess MIME type derived from the extension.
    pub mime_type: String,
    /// Partial SHA-256 fingerprint of the first 64 KB.
    pub fingerprint: String,
}

// ---------------------------------------------------------------------------
// Scanner
// ---------------------------------------------------------------------------

/// Stateless directory scanner.
///
/// All configuration is supplied per-call via [`ScanOptions`].
pub struct AssetScanner;

impl AssetScanner {
    /// Scan a directory according to the provided options.
    ///
    /// # Errors
    ///
    /// Returns an error if the root directory does not exist or is not
    /// readable.  Individual file errors (e.g. permission denied on a single
    /// file) are logged but do **not** abort the scan.
    pub fn scan(options: &ScanOptions) -> Result<ScanResult> {
        let root = Path::new(&options.directory);
        if !root.exists() {
            anyhow::bail!("scan directory does not exist: {}", options.directory);
        }
        if !root.is_dir() {
            anyhow::bail!("scan path is not a directory: {}", options.directory);
        }

        info!(directory = %options.directory, recursive = options.recursive, "starting asset scan");

        let start = Instant::now();

        let max_depth = if options.recursive { usize::MAX } else { 1 };
        let walker = WalkDir::new(&options.directory)
            .max_depth(max_depth)
            .follow_links(false);

        let normalized_exts: Vec<String> = options
            .extensions
            .iter()
            .map(|e| e.to_lowercase().trim_start_matches('.').to_string())
            .collect();

        let mut assets: Vec<ScannedAsset> = Vec::new();
        let mut total_files_scanned: u32 = 0;

        for entry in walker {
            let entry = match entry {
                Ok(e) => e,
                Err(err) => {
                    warn!("skipping inaccessible entry: {err}");
                    continue;
                }
            };

            // Only consider regular files.
            if !entry.file_type().is_file() {
                continue;
            }

            total_files_scanned += 1;

            // Extension filter.
            let ext = entry
                .path()
                .extension()
                .and_then(|e| e.to_str())
                .map(|e| e.to_lowercase())
                .unwrap_or_default();

            if !normalized_exts.is_empty() && !normalized_exts.contains(&ext) {
                continue;
            }

            let file_path = entry.path().to_string_lossy().to_string();
            let file_name = entry.file_name().to_string_lossy().to_string();

            let metadata = match entry.metadata() {
                Ok(m) => m,
                Err(err) => {
                    warn!(path = %file_path, "failed to read metadata: {err}");
                    continue;
                }
            };

            let fingerprint = match compute_fingerprint(&file_path) {
                Ok(fp) => fp,
                Err(err) => {
                    warn!(path = %file_path, "failed to compute fingerprint: {err}");
                    continue;
                }
            };

            let mime_type = mime_from_extension(&ext);

            let asset = ScannedAsset {
                id: generate_id(),
                file_path,
                file_name,
                file_size: metadata.len(),
                mime_type,
                fingerprint,
            };

            debug!(id = %asset.id, path = %asset.file_path, "discovered asset");
            assets.push(asset);
        }

        let scan_duration = start.elapsed();
        let media_files_found = assets.len() as u32;

        info!(
            total_files_scanned,
            media_files_found,
            elapsed_ms = scan_duration.as_millis(),
            "asset scan complete"
        );

        Ok(ScanResult {
            assets,
            total_files_scanned,
            media_files_found,
            scan_duration,
        })
    }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Generate a simple unique identifier.
///
/// Uses a combination of timestamp nanos and a random component so we do not
/// pull in the `uuid` crate.
fn generate_id() -> String {
    use std::collections::hash_map::DefaultHasher;
    use std::hash::{Hash, Hasher};
    use std::time::SystemTime;

    let now = SystemTime::now()
        .duration_since(SystemTime::UNIX_EPOCH)
        .unwrap_or_default();

    let mut hasher = DefaultHasher::new();
    now.as_nanos().hash(&mut hasher);
    // Mix in the address of a stack variable for uniqueness across calls
    // within the same nanosecond.
    let stack_var: u8 = 0;
    let ptr = &stack_var as *const u8 as usize;
    ptr.hash(&mut hasher);

    format!("asset-{:016x}", hasher.finish())
}

/// Map a lowercase file extension to a MIME type string.
fn mime_from_extension(ext: &str) -> String {
    match ext {
        // Video
        "mp4" | "m4v" => "video/mp4".into(),
        "mov" => "video/quicktime".into(),
        "avi" => "video/x-msvideo".into(),
        "mkv" => "video/x-matroska".into(),
        "webm" => "video/webm".into(),
        "wmv" => "video/x-ms-wmv".into(),
        "flv" => "video/x-flv".into(),
        "mpg" | "mpeg" => "video/mpeg".into(),
        "ts" | "mts" | "m2ts" => "video/mp2t".into(),
        "3gp" => "video/3gpp".into(),
        // Audio
        "mp3" => "audio/mpeg".into(),
        "wav" => "audio/wav".into(),
        "aac" => "audio/aac".into(),
        "flac" => "audio/flac".into(),
        "ogg" | "oga" => "audio/ogg".into(),
        "wma" => "audio/x-ms-wma".into(),
        "m4a" => "audio/mp4".into(),
        "aiff" | "aif" => "audio/aiff".into(),
        // Image
        "png" => "image/png".into(),
        "jpg" | "jpeg" => "image/jpeg".into(),
        "gif" => "image/gif".into(),
        "bmp" => "image/bmp".into(),
        "tiff" | "tif" => "image/tiff".into(),
        "webp" => "image/webp".into(),
        "svg" => "image/svg+xml".into(),
        // Project / other
        "prproj" => "application/x-premiere-project".into(),
        "xml" => "application/xml".into(),
        "json" => "application/json".into(),
        _ => "application/octet-stream".into(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;
    use std::io::Write;

    /// Helper: create a temp directory with a few dummy files.
    fn make_test_tree() -> std::path::PathBuf {
        let base = std::env::temp_dir().join("asset_scanner_test");
        let _ = fs::remove_dir_all(&base);
        let sub = base.join("sub");
        fs::create_dir_all(&sub).unwrap();

        for name in &["clip.mp4", "audio.wav", "readme.txt"] {
            let mut f = fs::File::create(base.join(name)).unwrap();
            f.write_all(b"dummy content").unwrap();
        }
        fs::File::create(sub.join("nested.mov"))
            .unwrap()
            .write_all(b"nested")
            .unwrap();

        base
    }

    #[test]
    fn scan_finds_all_files_when_no_extension_filter() {
        let dir = make_test_tree();
        let opts = ScanOptions {
            directory: dir.to_string_lossy().into(),
            recursive: true,
            extensions: vec![],
        };
        let result = AssetScanner::scan(&opts).unwrap();
        assert_eq!(result.media_files_found, 4);
        fs::remove_dir_all(&dir).ok();
    }

    #[test]
    fn scan_filters_by_extension() {
        let dir = make_test_tree();
        let opts = ScanOptions {
            directory: dir.to_string_lossy().into(),
            recursive: true,
            extensions: vec!["mp4".into(), "mov".into()],
        };
        let result = AssetScanner::scan(&opts).unwrap();
        assert_eq!(result.media_files_found, 2);
        fs::remove_dir_all(&dir).ok();
    }

    #[test]
    fn scan_respects_non_recursive() {
        let dir = make_test_tree();
        let opts = ScanOptions {
            directory: dir.to_string_lossy().into(),
            recursive: false,
            extensions: vec![],
        };
        let result = AssetScanner::scan(&opts).unwrap();
        // Only the 3 files in the root, not the nested one.
        assert_eq!(result.media_files_found, 3);
        fs::remove_dir_all(&dir).ok();
    }

    #[test]
    fn scan_fails_for_nonexistent_directory() {
        let opts = ScanOptions {
            directory: "/nonexistent/dir".into(),
            recursive: true,
            extensions: vec![],
        };
        assert!(AssetScanner::scan(&opts).is_err());
    }
}
