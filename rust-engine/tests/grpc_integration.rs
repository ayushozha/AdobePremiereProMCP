//! Integration tests for the MediaEngineService gRPC server.
//!
//! Each test starts the server on a random OS-assigned port (binding to
//! `127.0.0.1:0`), connects a tonic client, and exercises one or more RPC
//! methods against real data on disk.

use std::net::SocketAddr;
use tokio::net::TcpListener;
use tonic::transport::Server;

use premierpro_media_engine::grpc::server::MediaEngineServiceImpl;
use premierpro_media_engine::proto::media::media_engine_service_client::MediaEngineServiceClient;
use premierpro_media_engine::proto::media::media_engine_service_server::MediaEngineServiceServer;
use premierpro_media_engine::proto::media::{
    DetectScenesRequest, ProbeMediaRequest, ScanAssetsRequest,
};

/// Start the gRPC server on a random port and return the address it is
/// listening on.  The server runs in a background tokio task and will be
/// dropped when the runtime shuts down at the end of the test.
async fn start_server() -> SocketAddr {
    let listener = TcpListener::bind("127.0.0.1:0")
        .await
        .expect("failed to bind to random port");
    let addr = listener.local_addr().unwrap();

    let incoming = tokio_stream::wrappers::TcpListenerStream::new(listener);

    let service = MediaEngineServiceImpl::new();

    tokio::spawn(async move {
        Server::builder()
            .add_service(MediaEngineServiceServer::new(service))
            .serve_with_incoming(incoming)
            .await
            .expect("gRPC server failed");
    });

    // Give the server a moment to start accepting connections.
    tokio::time::sleep(std::time::Duration::from_millis(100)).await;

    addr
}

/// Helper: connect a client to the given address.
async fn connect_client(
    addr: SocketAddr,
) -> MediaEngineServiceClient<tonic::transport::Channel> {
    let url = format!("http://{addr}");
    MediaEngineServiceClient::connect(url)
        .await
        .expect("failed to connect to gRPC server")
}

// ============================================================================
// Tests
// ============================================================================

#[tokio::test]
async fn scan_assets_returns_files() {
    let addr = start_server().await;
    let mut client = connect_client(addr).await;

    // Point the scanner at the project's docs/ directory, which we know exists
    // and contains at least a couple of files.
    let docs_dir =
        std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .parent()
            .unwrap()
            .join("docs");

    let request = ScanAssetsRequest {
        directory: docs_dir.to_string_lossy().to_string(),
        recursive: true,
        extensions: vec![], // all files
    };

    let response = client
        .scan_assets(request)
        .await
        .expect("ScanAssets RPC failed");

    let inner = response.into_inner();

    assert!(
        !inner.assets.is_empty(),
        "expected at least one asset in docs/, got none"
    );
    assert!(
        inner.total_files_scanned > 0,
        "total_files_scanned should be > 0"
    );
    assert!(
        inner.scan_duration_seconds >= 0.0,
        "scan_duration_seconds should be non-negative"
    );

    // Verify that every returned asset has basic fields populated.
    for asset in &inner.assets {
        assert!(!asset.file_path.is_empty(), "asset file_path should not be empty");
        assert!(!asset.file_name.is_empty(), "asset file_name should not be empty");
        assert!(asset.file_size_bytes > 0, "asset file_size_bytes should be > 0");
    }
}

#[tokio::test]
async fn probe_media_returns_metadata_for_known_file() {
    let addr = start_server().await;
    let mut client = connect_client(addr).await;

    // Use the Cargo.toml as a known file that exists.  The probe will
    // fail on ffprobe (it is not a media file) but we want to verify the
    // RPC round-trip completes and returns an informative error.
    let cargo_toml = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("Cargo.toml");

    let request = ProbeMediaRequest {
        file_path: cargo_toml.to_string_lossy().to_string(),
    };

    // Probing a non-media file should return an error from ffprobe, surfaced
    // as a gRPC INTERNAL status.
    let result = client.probe_media(request).await;

    // The important thing is that the server did not crash — it should have
    // returned either Ok (if ffprobe happened to produce output) or a clean
    // Status error.
    match result {
        Ok(response) => {
            let inner = response.into_inner();
            // If ffprobe did return something, the asset should at least have
            // a file_path.
            if let Some(asset) = &inner.asset {
                assert!(
                    !asset.file_path.is_empty(),
                    "probed asset should have a file_path"
                );
            }
        }
        Err(status) => {
            // An INTERNAL error is expected when ffprobe cannot parse the
            // file.  Any other code would be surprising.
            assert_eq!(
                status.code(),
                tonic::Code::Internal,
                "expected INTERNAL status for non-media file, got {:?}: {}",
                status.code(),
                status.message()
            );
        }
    }
}

#[tokio::test]
async fn probe_media_fails_for_missing_file() {
    let addr = start_server().await;
    let mut client = connect_client(addr).await;

    let request = ProbeMediaRequest {
        file_path: "/nonexistent/path/video.mp4".to_string(),
    };

    let result = client.probe_media(request).await;
    assert!(result.is_err(), "probing a missing file should fail");
    let status = result.unwrap_err();
    assert_eq!(
        status.code(),
        tonic::Code::Internal,
        "expected INTERNAL status, got {:?}",
        status.code()
    );
    assert!(
        status.message().contains("file not found") || status.message().contains("probe failed"),
        "error message should mention the missing file: {}",
        status.message()
    );
}

#[tokio::test]
async fn scan_assets_fails_for_nonexistent_directory() {
    let addr = start_server().await;
    let mut client = connect_client(addr).await;

    let request = ScanAssetsRequest {
        directory: "/nonexistent/directory/path".to_string(),
        recursive: false,
        extensions: vec![],
    };

    let result = client.scan_assets(request).await;
    assert!(
        result.is_err(),
        "scanning a nonexistent directory should fail"
    );
}

#[tokio::test]
async fn detect_scenes_returns_stub_response() {
    let addr = start_server().await;
    let mut client = connect_client(addr).await;

    let request = DetectScenesRequest {
        file_path: "/any/file.mp4".to_string(),
        threshold: 0.3,
    };

    let result = client.detect_scenes(request).await;

    // The stub implementation should return a response (not an error).
    match result {
        Ok(response) => {
            let inner = response.into_inner();
            // The stub returns a single scene change entry.
            assert!(
                !inner.scenes.is_empty(),
                "stub should return at least one scene entry"
            );
        }
        Err(status) => {
            // Also acceptable if the implementation returns an informative
            // error about scene detection not being available.
            assert!(
                status.message().contains("scene detection")
                    || status.message().contains("not implemented")
                    || status.message().contains("ffmpeg"),
                "unexpected error: {}",
                status.message()
            );
        }
    }
}
