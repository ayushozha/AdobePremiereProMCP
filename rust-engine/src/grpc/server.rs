//! gRPC server implementation for MediaEngineService.

use tonic::{Request, Response, Status};
use tracing::{info, instrument, warn};

use crate::assets::scanner::{AssetScanner, ScanOptions};
use crate::media::probe::{self as probe_types, MediaProber};
use crate::media::scenes::{SceneDetectionOptions, SceneDetector};
use crate::thumbnails::generator::{ThumbnailGenerator, ThumbnailOptions};
use crate::waveform::analyzer::{WaveformAnalyzer, WaveformOptions};

use crate::proto::common::{Asset, AssetType, AudioInfo, Resolution, Timecode, VideoInfo};
use crate::proto::media::media_engine_service_server::MediaEngineService;
use crate::proto::media::{
    AnalyzeWaveformRequest, AnalyzeWaveformResponse, DetectScenesRequest, DetectScenesResponse,
    GenerateThumbnailRequest, GenerateThumbnailResponse, ProbeMediaRequest, ProbeMediaResponse,
    ScanAssetsRequest, ScanAssetsResponse, SceneChange, SilenceRegion,
};

// ============================================================================
// Type conversion helpers
// ============================================================================

/// Convert an internal `probe::AssetType` to the proto `AssetType` enum value (i32).
fn asset_type_to_proto(at: &probe_types::AssetType) -> i32 {
    match at {
        probe_types::AssetType::Video => AssetType::Video.into(),
        probe_types::AssetType::Audio => AssetType::Audio.into(),
        probe_types::AssetType::Image => AssetType::Image.into(),
        probe_types::AssetType::Graphics => AssetType::Graphics.into(),
        probe_types::AssetType::Unknown => AssetType::Unspecified.into(),
    }
}

/// Convert an internal `probe::VideoInfo` to the proto `VideoInfo`.
fn video_info_to_proto(v: &probe_types::VideoInfo) -> VideoInfo {
    VideoInfo {
        codec: v.codec.clone(),
        resolution: Some(Resolution {
            width: v.width,
            height: v.height,
        }),
        frame_rate: v.frame_rate,
        bitrate_bps: v.bitrate_bps,
        pixel_format: v.pixel_format.clone(),
        duration_seconds: v.duration_seconds,
    }
}

/// Convert an internal `probe::AudioInfo` to the proto `AudioInfo`.
fn audio_info_to_proto(a: &probe_types::AudioInfo) -> AudioInfo {
    AudioInfo {
        codec: a.codec.clone(),
        sample_rate: a.sample_rate,
        channels: a.channels,
        bitrate_bps: a.bitrate_bps,
        duration_seconds: a.duration_seconds,
    }
}

/// Convert a `MediaInfo` (from the prober) into a proto `Asset`.
fn media_info_to_proto_asset(info: &probe_types::MediaInfo) -> Asset {
    Asset {
        id: String::new(),
        file_path: info.file_path.clone(),
        file_name: info.file_name.clone(),
        file_size_bytes: info.file_size,
        mime_type: info.mime_type.clone(),
        asset_type: asset_type_to_proto(&info.asset_type),
        video: info.video.as_ref().map(video_info_to_proto),
        audio: info.audio.as_ref().map(audio_info_to_proto),
        metadata: info.metadata.clone(),
        fingerprint: String::new(),
    }
}

/// Convert a proto `Timecode` to seconds.
fn timecode_to_seconds(tc: &Timecode) -> f64 {
    let base = (tc.hours as f64) * 3600.0
        + (tc.minutes as f64) * 60.0
        + (tc.seconds as f64);
    let frame_rate = if tc.frame_rate > 0.0 {
        tc.frame_rate
    } else {
        // Default to 24 fps if not specified, so frame offset is still meaningful.
        24.0
    };
    base + (tc.frames as f64) / frame_rate
}

// ============================================================================
// Service implementation
// ============================================================================

/// Implementation of the MediaEngineService gRPC service.
pub struct MediaEngineServiceImpl;

impl MediaEngineServiceImpl {
    pub fn new() -> Self {
        Self
    }
}

impl Default for MediaEngineServiceImpl {
    fn default() -> Self {
        Self::new()
    }
}

#[tonic::async_trait]
impl MediaEngineService for MediaEngineServiceImpl {
    #[instrument(skip(self, request), fields(directory))]
    async fn scan_assets(
        &self,
        request: Request<ScanAssetsRequest>,
    ) -> Result<Response<ScanAssetsResponse>, Status> {
        let req = request.into_inner();
        info!(directory = %req.directory, recursive = req.recursive, "Scanning assets");

        let scan_options = ScanOptions {
            directory: req.directory.clone(),
            recursive: req.recursive,
            extensions: req.extensions,
        };

        let scan_result = tokio::task::spawn_blocking(move || AssetScanner::scan(&scan_options))
            .await
            .map_err(|e| Status::internal(format!("scan task panicked: {e}")))?
            .map_err(|e| Status::internal(format!("asset scan failed: {e}")))?;

        // For each scanned asset, probe it with MediaProber to get full metadata.
        let mut proto_assets = Vec::with_capacity(scan_result.assets.len());
        for scanned in &scan_result.assets {
            match MediaProber::probe(&scanned.file_path) {
                Ok(info) => {
                    let mut asset = media_info_to_proto_asset(&info);
                    asset.id = scanned.id.clone();
                    asset.fingerprint = scanned.fingerprint.clone();
                    proto_assets.push(asset);
                }
                Err(e) => {
                    // Probing may fail for non-media files or when ffprobe is
                    // unavailable.  Fall back to the basic metadata from the scanner.
                    warn!(path = %scanned.file_path, error = %e, "probe failed, using scanner metadata");
                    proto_assets.push(Asset {
                        id: scanned.id.clone(),
                        file_path: scanned.file_path.clone(),
                        file_name: scanned.file_name.clone(),
                        file_size_bytes: scanned.file_size,
                        mime_type: scanned.mime_type.clone(),
                        asset_type: AssetType::Unspecified.into(),
                        video: None,
                        audio: None,
                        metadata: std::collections::HashMap::new(),
                        fingerprint: scanned.fingerprint.clone(),
                    });
                }
            }
        }

        let response = ScanAssetsResponse {
            assets: proto_assets,
            total_files_scanned: scan_result.total_files_scanned,
            media_files_found: scan_result.media_files_found,
            scan_duration_seconds: scan_result.scan_duration.as_secs_f64(),
        };

        Ok(Response::new(response))
    }

    #[instrument(skip(self, request), fields(file_path))]
    async fn probe_media(
        &self,
        request: Request<ProbeMediaRequest>,
    ) -> Result<Response<ProbeMediaResponse>, Status> {
        let req = request.into_inner();
        info!(file_path = %req.file_path, "Probing media file");

        let file_path = req.file_path.clone();
        let info = tokio::task::spawn_blocking(move || MediaProber::probe(&file_path))
            .await
            .map_err(|e| Status::internal(format!("probe task panicked: {e}")))?
            .map_err(|e| Status::internal(format!("media probe failed: {e}")))?;

        let asset = media_info_to_proto_asset(&info);
        let response = ProbeMediaResponse { asset: Some(asset) };
        Ok(Response::new(response))
    }

    #[instrument(skip(self, request), fields(file_path))]
    async fn generate_thumbnail(
        &self,
        request: Request<GenerateThumbnailRequest>,
    ) -> Result<Response<GenerateThumbnailResponse>, Status> {
        let req = request.into_inner();
        info!(
            file_path = %req.file_path,
            format = %req.output_format,
            "Generating thumbnail"
        );

        let timestamp_seconds = req
            .timestamp
            .as_ref()
            .map(timecode_to_seconds)
            .unwrap_or(0.0);

        let (width, height) = req
            .output_size
            .as_ref()
            .map(|r| (r.width, r.height))
            .unwrap_or((320, 180));

        let options = ThumbnailOptions {
            timestamp_seconds,
            width,
            height,
            output_format: if req.output_format.is_empty() {
                "png".to_string()
            } else {
                req.output_format.clone()
            },
        };

        let file_path = req.file_path.clone();
        let result =
            tokio::task::spawn_blocking(move || ThumbnailGenerator::generate(&file_path, &options))
                .await
                .map_err(|e| Status::internal(format!("thumbnail task panicked: {e}")))?
                .map_err(|e| Status::internal(format!("thumbnail generation failed: {e}")))?;

        let response = GenerateThumbnailResponse {
            thumbnail_data: result.data,
            output_path: result.output_path.unwrap_or_default(),
            actual_size: Some(Resolution {
                width: result.actual_width,
                height: result.actual_height,
            }),
        };

        Ok(Response::new(response))
    }

    #[instrument(skip(self, request), fields(file_path))]
    async fn analyze_waveform(
        &self,
        request: Request<AnalyzeWaveformRequest>,
    ) -> Result<Response<AnalyzeWaveformResponse>, Status> {
        let req = request.into_inner();
        info!(
            file_path = %req.file_path,
            audio_track = req.audio_track,
            "Analyzing waveform"
        );

        let options = WaveformOptions {
            audio_track: req.audio_track,
            silence_threshold_db: if req.silence_threshold_db == 0.0 {
                -40.0 // sensible default when the client sends the zero-value
            } else {
                req.silence_threshold_db
            },
            min_silence_duration: if req.min_silence_duration_seconds == 0.0 {
                0.5 // sensible default
            } else {
                req.min_silence_duration_seconds
            },
        };

        let file_path = req.file_path.clone();
        let result =
            tokio::task::spawn_blocking(move || WaveformAnalyzer::analyze(&file_path, &options))
                .await
                .map_err(|e| Status::internal(format!("waveform task panicked: {e}")))?
                .map_err(|e| Status::internal(format!("waveform analysis failed: {e}")))?;

        let silence_regions = result
            .silence_regions
            .iter()
            .map(|sr| SilenceRegion {
                start_seconds: sr.start_seconds,
                end_seconds: sr.end_seconds,
                avg_db: sr.avg_db,
            })
            .collect();

        let response = AnalyzeWaveformResponse {
            silence_regions,
            peak_db: result.peak_db,
            rms_db: result.rms_db,
            duration_seconds: result.duration_seconds,
            waveform_samples: result.waveform_samples,
        };

        Ok(Response::new(response))
    }

    #[instrument(skip(self, request), fields(file_path))]
    async fn detect_scenes(
        &self,
        request: Request<DetectScenesRequest>,
    ) -> Result<Response<DetectScenesResponse>, Status> {
        let req = request.into_inner();
        info!(
            file_path = %req.file_path,
            threshold = req.threshold,
            "Detecting scenes"
        );

        let threshold = if req.threshold == 0.0 {
            0.3 // sensible default when the client sends the zero-value
        } else {
            req.threshold
        };

        let options = SceneDetectionOptions { threshold };

        let file_path = req.file_path.clone();
        let result =
            tokio::task::spawn_blocking(move || SceneDetector::detect(&file_path, &options))
                .await
                .map_err(|e| Status::internal(format!("scene detection task panicked: {e}")))?
                .map_err(|e| {
                    Status::internal(format!(
                        "scene detection failed (requires ffmpeg on PATH): {e}"
                    ))
                })?;

        // Convert internal SceneChangePoint to proto SceneChange.
        let scenes = result
            .scenes
            .iter()
            .map(|sc| {
                let total_secs = sc.timestamp_seconds;
                let hours = (total_secs / 3600.0) as u32;
                let minutes = ((total_secs % 3600.0) / 60.0) as u32;
                let seconds = (total_secs % 60.0) as u32;
                let frac = total_secs - total_secs.floor();
                // Assume 24 fps for frame calculation.
                let frames = (frac * 24.0) as u32;

                SceneChange {
                    timecode: Some(Timecode {
                        hours,
                        minutes,
                        seconds,
                        frames,
                        frame_rate: 24.0,
                    }),
                    confidence: sc.confidence,
                }
            })
            .collect();

        let response = DetectScenesResponse { scenes };
        Ok(Response::new(response))
    }
}
