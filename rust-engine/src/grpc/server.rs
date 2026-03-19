//! gRPC server implementation for MediaEngineService.

use tonic::{Request, Response, Status};
use tracing::{info, instrument};

use crate::proto::common::{Asset, AssetType, AudioInfo, Resolution, Timecode, VideoInfo};
use crate::proto::media::media_engine_service_server::MediaEngineService;
use crate::proto::media::{
    AnalyzeWaveformRequest, AnalyzeWaveformResponse, DetectScenesRequest, DetectScenesResponse,
    GenerateThumbnailRequest, GenerateThumbnailResponse, ProbeMediaRequest, ProbeMediaResponse,
    ScanAssetsRequest, ScanAssetsResponse, SceneChange, SilenceRegion,
};

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

        // TODO: Delegate to assets::scanner module.
        // Stub: return an empty result indicating a successful scan.
        let response = ScanAssetsResponse {
            assets: vec![],
            total_files_scanned: 0,
            media_files_found: 0,
            scan_duration_seconds: 0.0,
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

        // TODO: Delegate to media::probe module.
        // Stub: return a placeholder asset with basic info.
        let asset = Asset {
            id: String::new(),
            file_path: req.file_path.clone(),
            file_name: std::path::Path::new(&req.file_path)
                .file_name()
                .map(|n| n.to_string_lossy().to_string())
                .unwrap_or_default(),
            file_size_bytes: 0,
            mime_type: String::from("application/octet-stream"),
            asset_type: AssetType::Unspecified.into(),
            video: Some(VideoInfo {
                codec: String::new(),
                resolution: Some(Resolution {
                    width: 0,
                    height: 0,
                }),
                frame_rate: 0.0,
                bitrate_bps: 0,
                pixel_format: String::new(),
                duration_seconds: 0.0,
            }),
            audio: Some(AudioInfo {
                codec: String::new(),
                sample_rate: 0,
                channels: 0,
                bitrate_bps: 0,
                duration_seconds: 0.0,
            }),
            metadata: std::collections::HashMap::new(),
            fingerprint: String::new(),
        };

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

        // TODO: Delegate to thumbnails module.
        // Stub: return an empty thumbnail.
        let response = GenerateThumbnailResponse {
            thumbnail_data: vec![],
            output_path: String::new(),
            actual_size: Some(Resolution {
                width: 0,
                height: 0,
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

        // TODO: Delegate to waveform module.
        // Stub: return placeholder analysis.
        let response = AnalyzeWaveformResponse {
            silence_regions: vec![SilenceRegion {
                start_seconds: 0.0,
                end_seconds: 0.0,
                avg_db: -96.0,
            }],
            peak_db: 0.0,
            rms_db: -20.0,
            duration_seconds: 0.0,
            waveform_samples: vec![],
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

        // TODO: Delegate to media::scenes module.
        // Stub: return an empty scene list.
        let response = DetectScenesResponse {
            scenes: vec![SceneChange {
                timecode: Some(Timecode {
                    hours: 0,
                    minutes: 0,
                    seconds: 0,
                    frames: 0,
                    frame_rate: 24.0,
                }),
                confidence: 0.0,
            }],
        };

        Ok(Response::new(response))
    }
}
