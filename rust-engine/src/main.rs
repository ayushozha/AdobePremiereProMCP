use anyhow::Result;
use clap::Parser;
use tokio::signal;
use tonic::transport::Server;
use tracing::{info, warn};
use tracing_subscriber::{fmt, EnvFilter};

use premierpro_media_engine::grpc::server::MediaEngineServiceImpl;
use premierpro_media_engine::proto::media::media_engine_service_server::MediaEngineServiceServer;

/// PremierPro Media Engine - High-performance media processing via gRPC.
#[derive(Parser, Debug)]
#[command(name = "premierpro-media-engine")]
#[command(about = "High-performance media processing engine for PremierPro")]
struct Args {
    /// gRPC server port.
    #[arg(long, default_value_t = 50052, env = "MEDIA_ENGINE_PORT")]
    port: u16,

    /// Log level (trace, debug, info, warn, error).
    #[arg(long, default_value = "info", env = "MEDIA_ENGINE_LOG_LEVEL")]
    log_level: String,

    /// Enable JSON log output.
    #[arg(long, default_value_t = false, env = "MEDIA_ENGINE_LOG_JSON")]
    log_json: bool,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Initialize tracing subscriber.
    let env_filter = EnvFilter::try_new(&args.log_level)
        .unwrap_or_else(|_| EnvFilter::new("info"));

    if args.log_json {
        fmt()
            .json()
            .with_env_filter(env_filter)
            .init();
    } else {
        fmt()
            .with_env_filter(env_filter)
            .init();
    }

    info!(port = args.port, "Starting PremierPro Media Engine");

    #[cfg(feature = "ffmpeg")]
    info!("FFmpeg support: enabled");

    #[cfg(not(feature = "ffmpeg"))]
    info!("FFmpeg support: disabled (using Symphonia for audio)");

    let addr = format!("0.0.0.0:{}", args.port).parse()?;
    let service = MediaEngineServiceImpl::new();

    info!(%addr, "gRPC server listening");

    Server::builder()
        .add_service(MediaEngineServiceServer::new(service))
        .serve_with_shutdown(addr, async {
            shutdown_signal().await;
            warn!("Shutdown signal received, draining connections...");
        })
        .await?;

    info!("Media engine shut down gracefully");
    Ok(())
}

/// Wait for SIGINT or SIGTERM to trigger graceful shutdown.
async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install SIGTERM handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }
}
