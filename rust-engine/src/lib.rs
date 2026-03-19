//! PremierPro Media Engine
//!
//! High-performance media processing engine providing gRPC services for
//! media scanning, probing, waveform analysis, thumbnail generation,
//! and scene detection.

pub mod assets;
pub mod grpc;
pub mod media;
pub mod thumbnails;
pub mod waveform;

/// Generated protobuf/gRPC types.
///
/// The module hierarchy mirrors the proto package structure so that
/// cross-package references (`super::super::common::v1::...`) resolve
/// correctly in the prost-generated code.
pub mod proto {
    pub mod premierpro {
        pub mod common {
            pub mod v1 {
                tonic::include_proto!("premierpro.common.v1");
            }
        }
        pub mod media {
            pub mod v1 {
                tonic::include_proto!("premierpro.media.v1");
            }
        }
    }

    // Re-export at convenient short paths.
    pub use premierpro::common::v1 as common;
    pub use premierpro::media::v1 as media;
}
