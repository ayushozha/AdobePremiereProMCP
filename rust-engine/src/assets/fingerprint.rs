//! File fingerprinting via partial SHA-256.
//!
//! Computes a hex-encoded SHA-256 digest over the first 64 KB of a file.
//! This is intentionally a *partial* hash — it is fast enough for
//! deduplication checks without reading the entire file.

use anyhow::{Context, Result};
use sha2::{Digest, Sha256};
use std::fs::File;
use std::io::Read;

/// Maximum number of bytes read for fingerprinting (64 KB).
const FINGERPRINT_READ_SIZE: usize = 64 * 1024;

/// Compute a SHA-256 fingerprint of the first [`FINGERPRINT_READ_SIZE`] bytes
/// of the file at `file_path`.
///
/// # Errors
///
/// Returns an error if the file cannot be opened or read.
///
/// # Examples
///
/// ```no_run
/// use premierpro_media_engine::assets::fingerprint::compute_fingerprint;
///
/// let hash = compute_fingerprint("/path/to/video.mp4").unwrap();
/// println!("fingerprint: {hash}");
/// ```
pub fn compute_fingerprint(file_path: &str) -> Result<String> {
    let mut file = File::open(file_path)
        .with_context(|| format!("failed to open file for fingerprinting: {file_path}"))?;

    let mut buffer = vec![0u8; FINGERPRINT_READ_SIZE];
    let bytes_read = file
        .read(&mut buffer)
        .with_context(|| format!("failed to read file for fingerprinting: {file_path}"))?;
    buffer.truncate(bytes_read);

    let digest = Sha256::digest(&buffer);
    Ok(hex::encode(digest))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    #[test]
    fn fingerprint_produces_consistent_hex_digest() {
        let dir = std::env::temp_dir().join("fp_test");
        std::fs::create_dir_all(&dir).unwrap();
        let path = dir.join("sample.bin");

        let mut f = File::create(&path).unwrap();
        f.write_all(&[0xDE, 0xAD, 0xBE, 0xEF]).unwrap();

        let fp1 = compute_fingerprint(path.to_str().unwrap()).unwrap();
        let fp2 = compute_fingerprint(path.to_str().unwrap()).unwrap();

        assert_eq!(fp1, fp2, "fingerprint must be deterministic");
        assert_eq!(fp1.len(), 64, "SHA-256 hex digest must be 64 chars");

        std::fs::remove_file(&path).ok();
    }

    #[test]
    fn fingerprint_fails_for_missing_file() {
        let result = compute_fingerprint("/nonexistent/path/file.bin");
        assert!(result.is_err());
    }
}
